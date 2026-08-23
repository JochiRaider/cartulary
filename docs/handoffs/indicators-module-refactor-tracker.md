# indicators Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Item | Decision or evidence |
| --- | --- |
| Target path | `internal/modules/indicators` |
| Target label | `indicators`; derived from the final path segment and already valid lowercase kebab case |
| Output path | `docs/handoffs/indicators-module-refactor-tracker.md` |
| Repository baseline | Branch `main`, commit `07b2f2b703fc80dc10fcdba28cb855db660c44d2`; the tracker is the sole staged addition at W-00 start on 2026-08-22. |
| Revision baseline | Earlier handoff rows accurately describe their session-local state. The current implementation session rechecked live state and found the tracker staged, 63 Indicators files, and 55 guarded root exports. |
| Planning status | End-to-end remediation complete; W-00 through S-06 are closed with retained evidence and no unresolved blocker. |
| Allowed change | Adopted owner documents, the Indicators module and its tests, application composition, authored contracts and verification inputs, the next migration, Make-generated projections, and this tracker, only as required by W-01 and S-00 through S-06. |
| Non-goals | No frontend, grid-adapter, public-route, Inspector-tuple, view-schema, Network Flow `binding_only`, or canonical-dedupe-query change; no compatibility shim, silent data repair, generated-file hand edit, or cross-owner relocation. |
| Source posture | The local planning framework is a template and doctrine, not proof of repository state. Every repository claim below comes from live inspection. |
| Implementation authorization | The user task titled “Indicators Specification and Implementation Remediation” is `IND-AUTHORIZATION-001`. It authorizes the sequential workstreams in Section 7, including owner-conformance corrections, lifecycle contract and migration work, source-catalog construction, and the reviewed facade contraction. |

### Normative posture

This tracker is a prescriptive execution plan and handoff, not an adopted product-behavior owner. The keywords `MUST`, `MUST NOT`, `REQUIRED`, `SHOULD`, `SHOULD NOT`, and `MAY` governed the authorized Indicators remediation sequence and govern any explicit deferred gate recorded here. They do not amend an adopted owner, independently authorize a repository write, or make this tracker executable evidence. A conflict with an adopted owner is always `BLOCKED: owner contradiction`.

| Material class | Authority and effect | Default when silent |
| --- | --- | --- |
| Owner-defined product behavior | Adopted subsystem NLSpecs and Core 00 through Core 04 govern their named behavior. The implementation MUST conform to them. | Preserve the live observable behavior only where it is consistent with the owner. |
| Tracker-local execution requirement | Stable `IND-*`, workflow, slice, blocker, and acceptance IDs govern sequencing, scope, evidence, rollback, and handoff for this refactor. | The later implementer MUST stop rather than broaden scope. |
| Adopted governance decision | `IND-BOUNDARY-001` is adopted by its decision and Core owner artifacts; this tracker records execution evidence only. | Treat the owner artifacts as authority and stop on contradiction. |
| Implementation latitude | A mechanism is free only when this tracker and its owners explicitly leave it free and independent implementations remain behaviorally interchangeable. | Prefer the smallest owner-internal mechanism with no new export or compatibility layer. |
| Conditional work | A named activation event is REQUIRED before the work or evidence row applies. | The work remains inactive. |
| Deferred work | The item is deliberately excluded from S-00 through S-06. | It MUST NOT be implemented opportunistically. |

`temp/analysis-notes.md` supplies recommendations, and `docs/research/nlspec-spec.md` supplies specification-writing guidance. Both are informative inputs. Neither is an adopted owner or proof of repository state. Each normative fact introduced from them below is independently bounded by an adopted owner or live source evidence.

The authority order used for this tracker is:

1. Adopted subsystem NLSpecs for their named scope.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. Domain vocabulary and implementation-support guides for terminology, boundaries, and harness mechanics.
5. Current repository code and tests for implementation state.
6. Prior trackers, handoffs, and the planning framework as evidence only.

Core 05 is not applicable to this remediation: no timed or fixture-sensitive claim publication is activated, and the live Indicators test-family rows have implementation evidence posture. No owner contradiction was found. If a later workstream finds one, it must record `BLOCKED: owner contradiction` without selecting a side.

Owner and support documents inspected:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`, read first as required;
- `docs/spec/00_document_set_status_and_precedence.md`;
- `docs/spec/01_architecture_storage_and_view_contracts.md`;
- `docs/spec/02_domain_model_schema_and_history.md`;
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`;
- `docs/spec/04_security_deployment_and_conformance.md`;
- `docs/network-flow-activity-nlspec.md`, adopted/current for the Network Flow-owned binding boundary only;
- `docs/testing-harness-nlspec.md`, adopted/current for harness mechanics only;
- `docs/domain.md`, for vocabulary and bounded-context navigation;
- `docs/research/nlspec-spec.md`, as informative NLSpec authoring guidance only;
- `temp/analysis-notes.md`, as informative closure recommendations only.

Repository files inspected include all 63 files inventoried in Section 2 and these direct callers or contract/evidence surfaces:

- application assembly: `internal/app/server/runtime_assembly.go`, `internal/app/workbookassembly/indicator_adapter.go`, `internal/app/indicatorassembly/source_text.go`, `internal/app/importassembly/owner_registry.go`, `internal/app/incidentportabilityassembly/catalog.go`, `internal/app/revisionassembly/revisions.go`, `internal/app/recoveryassembly/state_catalog.go`, and `internal/app/projectionassembly/{build.go,catalog.go}`;
- cross-owner consumers: `internal/modules/networkflow/{ports.go,indicator_link.go,binding_store.go,transaction_participants.go}`;
- Records boundary evidence: `internal/modules/records/{route_target.go,route_target_integration_test.go,store.go}` and the two live Indicators validation call sites/order in `source_repository.go`, `observation_repository.go`, and `observation_service.go`;
- frontend consumers: `apps/web/src/workbook/features/indicators/{IndicatorInspectorWorkflow.tsx,IndicatorInspectorWorkflow.test.tsx,indicatorInspectorHandlers.ts}` and `apps/web/src/workbook/mutations/{createIndicatorWorkflowPort.ts,createIndicatorWorkflowPort.test.ts}`;
- authored contracts: `contracts/view-schemas/cartulary.view.indicators.v1.json`, `contracts/view-inspector/index.json`, `contracts/imports/view-targets.v1.json`, `contracts/incident-bundles/{source_catalog.json,indicators.row.v1.schema.json,indicator_observations.row.v1.schema.json,indicator_state_intervals.row.v1.schema.json}`, `contracts/openapi-source/manifest.json`, and `contracts/openapi-source/owners/module.indicators/openapi.json`;
- verification and harness routing: `contracts/verification/owners/module.indicators.json`, `tools/test_families/{module.indicators.json,module.workbook.json}`, `tools/test_catalog_owner.json`, and `tools/execution_topology_manifest.json`;
- adopted-decision registration examples: `docs/decisions/entities-module-boundary.md` and the corresponding Core 00/Core 04 registrations, used only to plan the future `IND-BOUNDARY-001` adoption shape.

Planning finding PF-001 records the principal framework/repository mismatch: the framework catalog summarizes Indicators as canonical indicators and observations with defanging and lifecycle complexity, while the live target also contains HTTP adaptation, import participation, projection source contribution, revision/rollback/delete-restore providers, Incident Bundle portability, recovery state, Network Flow participation, and test support. This is not evidence that all those responsibilities form one permanent package boundary. The remediation therefore diagnoses each seam independently and preserves owner-conformant behavior, while correcting behavior that conflicts with an adopted owner.

Planning finding PF-002 records an analysis-note/repository mismatch: the notes describe the two standalone Records resolutions generically as admission reads before a final mutation lock, while the live observation-create path first locks and revalidates all affected Records envelopes and then invokes both repository validations during insertion. S-01 therefore replaces the reads in their live post-lock positions and forbids statement-order movement. Repository state governs this implementation fact; Core 01 and Core 04 continue to govern envelope ownership and observable behavior.

## 2. Current-State Repository Inventory

The target contains 78 Go files: 41 production or support files and 37 tests. Every file is represented by one row below; none is out of scope for inventory.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/indicators/active_identity_claims_integration_test.go` | Characterizes active canonical identity claims across Records delete/restore and deterministic rebuild. | Test-only integration surface. | `module.indicators` harness row. | Indicators, Records, Recovery, PostgreSQL test harness. | Self. | Recovery contribution and active-identity storage contract. | Indicators with Records/Recovery collaboration. | High | Service-backed; freezes rebuildable coordination state rather than authoritative state. |
| `internal/modules/indicators/admission/create.go` | Strict Indicators system-view create decoding, view-schema admission, validation mapping, and request hashing. | `DecodeCreateRequest`, `CreateRequestHash`. | Workbook row-create adapter. | Indicators root DTOs, platform HTTP errors, generated view-schema registry. | `admission/create_test.go`; route/integration tests indirectly. | `cartulary.view.indicators.v1` and generated view-schema projection. | Indicators admission adapter. | High | Transport-neutral owner admission; unknown and nullable members are contract-sensitive. |
| `internal/modules/indicators/admission/create_test.go` | Characterizes create JSON admission and stable request hashing. | Test-only unit surface. | `module.indicators` harness row. | Admission package and view-schema fixtures. | Self. | View-schema create contract. | Indicators. | Medium | Protects replay digest and wire admission. |
| `internal/modules/indicators/boundary_guard_test.go` | Enforces sibling-module imports, transport-neutral admission, legacy-prefix removal, and Records envelope-mirror absence. | Test-only architecture guard. | `module.indicators` structure rows. | Go AST/parser and repository source. | Self. | Backend boundary posture. | Indicators architecture evidence. | Medium | Evidence guard, not runtime architecture. |
| `internal/modules/indicators/child_coordination.go` | Shared transaction helpers for child idempotency, UTF-8 spans, record locking/versioning, snapshots, revisions, projection refresh, and affected-record ordering. | Private root-package helpers. | Observation and lifecycle services. | authn idempotency, Records, Revisions, SourceText and projection ports. | Child routes, transaction, resolution, and unit tests. | Child mutation envelopes and `record_changed` intents indirectly. | Indicators mutation coordination. | High | Central atomicity seam; not a generic cross-owner coordinator. |
| `internal/modules/indicators/child_routes_integration_test.go` | Full child HTTP workflow characterization including paging, replay, transitions, roles, and Collaboration intents. | Test-only integration surface. | `module.indicators` route/production-workflow rows. | HTTP test application, Postgres, auth/session helpers. | Self. | Six OpenAPI paths, eight operations, frontend workflow expectations. | Indicators with Workbook/Collaboration collaboration. | High | Primary service-backed route freeze. |
| `internal/modules/indicators/contracts.go` | Cohesive root public contracts for create, participant, observation, lifecycle, source-text, results, and classified errors plus owner validation. | The reviewed DTO/error/validation portion of the 50-export root facade. | Server, HTTP, Workbook, Imports, Network Flow, contributions, and tests. | authn actor value, pgx transaction type, owner identity/origin vocabulary. | Export, adapter, route, replay, transaction, and unit tests. | View/OpenAPI contracts and export allowlist. | Indicators facade. | High | `CreateResult` exposes only `Created` and `Replayed`; no outcome enum or compatibility alias remains. |
| `internal/modules/indicators/create_service.go` | Canonical create/reuse/update orchestration, idempotency, Records envelope mutation, projection refresh, revision append, and caller-transaction participant. | `Store.CreateIndicatorRow`, `Store.FindOrCreateIndicatorParticipantTx` methods through root facade. | Workbook create adapter, Imports facade, Network Flow participant. | authn, Postgres/pgx, Records, Revisions, projections, identity/repositories. | Indicators integration, identity concurrency, transaction, unit tests. | Workbook create envelope, revision and Collaboration contracts. | Indicators mutation coordinator. | High | Preserves canonical dedupe and participant atomicity. |
| `internal/modules/indicators/envelope_contract_migration_test.go` | Verifies Indicators migration contraction to Records-owned envelopes across down/up behavior. | Test-only integration surface. | `module.indicators` storage/migration row. | PostgreSQL migration harness. | Self. | Authored migrations and envelope schema behavior; migrations are not edited here. | Indicators/Records evidence. | High | Service-backed and migration-sensitive. |
| `internal/modules/indicators/exported_surface_test.go` | AST allowlist for the exact 50-declaration root exported surface and its production-role inventory. | Test-only architecture guard; exact reviewed surface. | `module.indicators.surface.iteration2_exported_surface_reachability`. | Go AST/parser and target source. | Self. | Public internal-Go boundary only. | Indicators architecture evidence. | Medium | Fails on any stale/unapproved symbol, non-50 count, or retained symbol without a production-role reason. |
| `internal/modules/indicators/httpapi/decoding.go` | Closed child-route JSON decoders, UUID/timestamp/exact-token validation, and request hashing. | Private `httpapi` package surface. | `httpapi/routes.go`. | Indicators DTOs, owner vocabulary, platform HTTP errors, standard JSON/crypto. | Child route integration, production contract, and exact HTTP-vocabulary tests. | Indicator OpenAPI request schemas. | Indicators transport-adjacent adapter. | High | Exact token membership delegates to the S-02A owner registry; HTTP does not canonicalize tokens. |
| `internal/modules/indicators/httpapi/error_mapping_test.go` | Verifies target-role public errors and secret-safe internal storage mapping. | Test-only unit surface. | `module.indicators.target_resolution.safe_http_error_mapping`. | Private HTTP mutation error adapter and root owner errors. | Self. | Core 04 target-specific error projections. | Indicators security evidence. | High | Added in S-01; prevents storage diagnostics from becoming false semantic responses or leaking internals. |
| `internal/modules/indicators/httpapi/vocabulary_admission_test.go` | Rejects case, spacing, alias, and unknown Indicator/lifecycle tokens at HTTP admission. | Test-only unit surface. | `module.indicators.vocabulary.http_admission`. | Private child decoders. | Self. | OpenAPI Indicator observation/lifecycle token projections. | Indicators. | Medium | S-02A exact-boundary evidence. |
| `internal/modules/indicators/httpapi/routes.go` | Registers and handles eight child operations with session, CSRF, membership/role, visibility, cursor, envelope, and error mapping. | `Service`, `RegisterRoutes`; private `ownerApplication` port. | Server runtime route assembly. | Indicators facade, incidents admission, Records, authn/httpauth, platform HTTP/pagination. | Child routes and production contract tests; frontend typed-operation tests indirectly. | Owner OpenAPI source and generated protocol operation IDs. | Indicators transport-adjacent adapter. | High | Six paths; reads require membership and mutations require editor-or-higher. |
| `internal/modules/indicators/identity_concurrency_test.go` | Proves concurrent find-or-create convergence on one active identity. | Test-only integration surface. | `module.indicators` identity row. | Postgres dedicated fixture, Indicators participant. | Self. | Active-identity uniqueness behavior. | Indicators with Network Flow collaboration. | High | Locking and dedupe ordering are refactor-sensitive. |
| `internal/modules/indicators/identity_portability_test.go` | Checks canonical identity parity between live create and portability preparation. | Test-only unit/integration support. | Indicators identity/portability harness selectors. | Identity and Incident Bundle provider logic. | Self. | Portable Indicator normalization contract. | Indicators. | High | Prevents import/export identity drift. |
| `internal/modules/indicators/vocabulary_contract_test.go` | Verifies registry parity with applicable OpenAPI and all three portable row schemas and rejects non-exact portable tokens/duplicate support refs. | Test-only unit surface. | `module.indicators.vocabulary.contract_parity_and_portable_admission`. | Owner vocabulary, authored contracts, Incident Bundle contribution. | Self. | OpenAPI and three Incident Bundle row schemas. | Indicators with Incident Bundles collaboration. | High | Reads contracts only in test code; production remains independent of documentation and evidence metadata. |
| `internal/modules/indicators/import_create.go` | Implements the Indicators-owned Imports create facade and caller-transaction revision/projection finalization. | `NewImportCreateFacade`; private `Store.createImportRowTx`. | Imports owner registry/application assembly. | Imports ownerfacade, authn actor value, Indicators create/revision/projection logic. | Indicators integration and cross-owner Imports rows. | Import target `module.indicators@1`, facade `indicators.import_create`. | Indicators owner facade; Imports orchestrates. | High | Does not own parsing, sessions, mappings, or job orchestration. |
| `internal/modules/indicators/incident_bundle_contribution.go` | Fallibly loads the source-state catalog and projects its ordered portability descriptors into the source port alongside subtype presence. | `IncidentBundleContribution`, `NewIncidentBundleContribution`. | Incident portability assembly. | Owner source-state catalog, Incident Bundle sourceport/subtype presence, and internal provider. | Source-state parity, provider encapsulation, and portability tests. | Incident Bundle family `indicators`. | Indicators contribution facade. | High | Generic coordinators consume only this public contribution; construction errors propagate through assembly. |
| `internal/modules/indicators/indicators_test.go` | Broad canonical create/query/projection/observation/lifecycle behavior characterization. | Test-only integration surface. | Multiple `module.indicators` rows. | Server/test application, Postgres, view query. | Self. | System-view create/query and projection contracts. | Indicators. | High | Covers canonical rows rather than source observations as grid identity. |
| `internal/modules/indicators/internal/identity/identity.go` | Context-specific type/value canonicalization, defanging, hash-pair rules, IP/domain/URL/email normalization, observation candidate normalization, and dedupe keys. | Internal exports `Input`, `Canonical`, `ValidationError`, normalization functions, `IsIPType`, `DedupeKey`. | Root create/store, observation repository, Incident Bundle preparation, testsupport. | Owner vocabulary, standard net/url/crypto, platform field normalization. | `identity_test.go`, identity portability/concurrency and unit tests. | Indicator type/value contracts and portability schemas indirectly. | Indicators domain logic. | High | Creation deliberately tolerates case/outer spacing but aliases remain invalid; exact membership lives in vocabulary. |
| `internal/modules/indicators/internal/identity/identity_test.go` | Exhaustive registry, alias rejection, canonicalization, IP, and dedupe characterization. | Test-only unit surface. | `module.indicators` identity row. | Internal identity. | Self. | Core 02 normalization vocabulary. | Indicators. | High | Primary pure-domain characterization. |
| `internal/modules/indicators/internal/origin/origin.go` | Closed observation-origin registry and parser. | Internal `ObservationOrigin`, seven constants, `ValidationError`, `Parse`. | Observation repository and Incident Bundle provider. | Standard strings. | Origin unit/integration and rollback tests. | OpenAPI and portability origin enums indirectly. | Indicators domain logic. | Medium | Manual public route admits only `manual_entry`; portability retains all owner tokens. |
| `internal/modules/indicators/internal/sourcestate/catalog.go` | Fallibly constructs and caches one immutable ordered catalog for authoritative/rebuildable relations and portable path descriptors. | Owner-internal `Load`, catalog accessors, and fact types. | Recovery and Incident Bundle contribution constructors. | Standard validation, sorting, and defensive-copy primitives only. | `catalog_test.go` and root source-state parity tests. | Recovery fixture and authored Incident Bundle source catalog indirectly. | Indicators source-state owner. | High | Validates relation/path/schema/version/order/identity uniqueness and never reads contracts at runtime. |
| `internal/modules/indicators/internal/sourcestate/catalog_test.go` | Verifies exact 3/1/3 facts, invalid-construction rejection, deterministic ordering, and defensive copies. | Test-only unit surface. | `module.indicators.source_state.catalog_validation`. | Owner source-state catalog. | Self. | Recovery and Incident Bundle projections indirectly. | Indicators. | High | Includes duplicate, empty, unsafe pairing, order, version, identity, and mutation cases. |
| `internal/modules/indicators/internal/vocabulary/vocabulary.go` | Immutable closed registries for Indicator type, value kind, observation status, and lifecycle state, with exact membership and creation-only canonicalization. | Owner-internal registry accessors/checks and type/value canonicalizers. | Identity, HTTP, lifecycle service, rollback, and Incident Bundle preparation. | Standard strings only. | Vocabulary, identity, route, rollback, portability, and lifecycle tests. | Core 02, OpenAPI, and Incident Bundle token projections. | Indicators domain vocabulary. | High | Returns defensive copies; no runtime consumer reads owner Markdown or schemas. |
| `internal/modules/indicators/internal/vocabulary/vocabulary_test.go` | Exhaustively verifies closed sets, exact rejection, aliases, creation canonicalization, and defensive copies. | Test-only unit surface. | `module.indicators.vocabulary.closed_registry_and_canonical_creation`. | Owner vocabulary. | Self. | Applicable token projections indirectly. | Indicators. | Medium | Pure deterministic S-02A evidence. |
| `internal/modules/indicators/internal/providers/deleterestore/provider.go` | Source-owned delete/restore snapshot provider joining Records envelope and Indicator subtype. | Internal `Source`, `NewSource`. | `NewRevisionContribution`. | Revisions delete/restore contract, pgx. | Provider encapsulation, revision and portability tests indirectly. | Revision snapshot `cartulary.revisions.snapshot.indicator.v1`. | Indicators source provider. | High | Records remains envelope/deletion authority. |
| `internal/modules/indicators/internal/providers/incidentbundle/portable_apply.go` | Applies prepared Indicator, observation, and interval rows and validates referenced state during bundle import. | Private provider functions. | Incident Bundle source port. | pgx, incidentportability contexts, prepared models. | Portability characterization/final-state/invariant tests. | Three bundle row schemas and paths. | Indicators portability provider. | High | Historical import path; no Collaboration live publication. |
| `internal/modules/indicators/internal/providers/incidentbundle/portable_export.go` | Deterministically exports Records-joined Indicators and child rows with portable attribution. | Private provider functions. | Incident Bundle source port. | pgx, sourceport export context, identity/origin helpers. | Portability characterization and invariant tests. | `data/indicators.ndjson`, observation and interval paths. | Indicators portability provider. | High | Portable bytes and ordering are frozen. |
| `internal/modules/indicators/internal/providers/incidentbundle/portable_model.go` | Internal portable row models, canonical JSON/digest helpers, and shared conversions. | Private provider types/functions. | Prepare, apply, export, validate. | uuid, crypto/json, origin registry. | Portability tests indirectly. | Bundle v2 row shapes. | Indicators portability provider. | High | Runtime-only model; schemas remain authored contracts. |
| `internal/modules/indicators/internal/providers/incidentbundle/portable_prepare.go` | Strictly admits three NDJSON row families with exact registry membership, unique lifecycle support refs, identity uniqueness, attribution, and state coherence. | Private provider functions. | Incident Bundle source port. | incidentportability contexts, identity/origin/vocabulary, JSON. | Portability invariant/final-state/vocabulary tests. | Three Incident Bundle schemas and invariant IDs. | Indicators portability provider. | High | Repeated observations remain distinct; portable tokens and support references are never canonicalized or silently repaired. |
| `internal/modules/indicators/internal/providers/incidentbundle/portable_validate.go` | Validates applied final state and envelope/source consistency. | Private provider functions. | Incident Bundle source port. | pgx and prepared models. | Portability final-state and invariant tests. | Incident Bundle family invariants. | Indicators portability provider. | High | Fail-closed verification after apply. |
| `internal/modules/indicators/internal/providers/incidentbundle/source_port.go` | Binds catalog-projected portability paths to sourceport export/prepare/apply/validate dispatch and owner invariants. | Internal constructor `NewSourcePort` through contribution wrapper. | Root Incident Bundle contribution. | Incident Bundle sourceport contracts and provider helpers. | Provider encapsulation, source-state parity, and portability tests. | Family `indicators`; three logical paths, version 2. | Indicators contribution implementation. | High | It no longer owns or duplicates the path/schema/version/stable-identity inventory. |
| `internal/modules/indicators/internal/providers/incidentbundle/subtype_presence.go` | Detects Indicators subtype presence for Incident Bundle claim/admission. | Internal `SubtypeContribution`. | Root Incident Bundle contribution. | Records subtype-presence contract, pgx. | Provider/portability tests indirectly. | Incident Bundle subtype contribution. | Indicators contribution implementation. | Medium | Presence is source state, not generated evidence. |
| `internal/modules/indicators/internal/providers/projection/source.go` | Reads source-owned projection input, including observation aggregates, latest lifecycle state, and supporting-link count. | Internal `LoadProjectionInputTx`, `ListProjectionInputsTx`. | Public `projectionprovider` adapter. | pgx, typed workbook projection input; SQL reads Indicators, Records, `active_record_links_v1`. | Projection contribution/delegation and integration tests. | Projection descriptor for `indicator_grid_projection`. | Indicators source semantics; Projections owns storage. | High | Descriptor explicitly declares source authorities `indicators`, `links`, `records`; this cross-owner read is intentional. |
| `internal/modules/indicators/internal/providers/rollback/children.go` | Parses, exactly validates, applies, tombstones, and restores observation/interval non-row history targets. | Private child rollback implementation. | Internal rollback provider. | Revisions rollback contracts, pgx, Records validation SQL, origin/vocabulary registries. | `children_test.go`, vocabulary and integration rollback/portability tests indirectly. | Revision target kinds `indicator_observation`, `indicator_state_interval`. | Indicators rollback provider. | High | Rejects non-exact statuses/lifecycle states and malformed or duplicate support refs with the existing safe rollback error. |
| `internal/modules/indicators/internal/providers/rollback/children_test.go` | Characterizes child snapshot parsing, origins, transitions, and rollback invariants. | Test-only unit surface. | `module.indicators` origin/revision rows. | Internal rollback provider. | Self. | Revision non-row snapshot contract. | Indicators. | High | Protects exact state coherence. |
| `internal/modules/indicators/internal/providers/rollback/provider.go` | Implements Indicator row rollback and exposes the combined child provider. | Internal `NewProvider`, `NewChildProvider`. | `NewRevisionContribution`. | Revisions rollback contracts, pgx, identity. | `provider_test.go`, revision contribution tests. | Indicator revision snapshot/rollback contract. | Indicators rollback provider. | High | Source-owner implementation remains internal. |
| `internal/modules/indicators/internal/providers/rollback/provider_test.go` | Characterizes Indicator snapshot parsing, identity validation, and row rollback preparation/application. | Test-only unit surface. | `module.indicators` revision/provider rows. | Internal rollback provider. | Self. | Revision snapshot schema. | Indicators. | High | Prevents invalid historical state restoration. |
| `internal/modules/indicators/internal/providers/rollback/vocabulary_test.go` | Rejects case, spacing, alias, unknown, and duplicate-reference historical vocabulary at rollback admission. | Test-only unit surface. | `module.indicators.vocabulary.rollback_admission`. | Internal rollback providers. | Self. | Revision Indicator/non-row snapshot contracts. | Indicators with Revisions collaboration. | High | Confirms existing `ErrTargetNotReversible` classification remains stable. |
| `internal/modules/indicators/lifecycle_repository.go` | Persists and keyset-lists append-only lifecycle intervals. | Private root repository. | Lifecycle service and Store list method. | pgx, JSON support refs. | Lifecycle/unit/route/integration tests. | Lifecycle OpenAPI resources and portability interval rows. | Indicators persistence. | High | Ordering is `(valid_from DESC, interval_id DESC)`. |
| `internal/modules/indicators/lifecycle_integrity_migration_test.go` | Proves clean install, valid upgrade, malformed legacy rollback, direct-write rejection, privilege preservation, and Down/Up lifecycle integrity. | Test-only migration integration surface. | `module.indicators.storage.lifecycle_support_reference_integrity`. | Migration scratch capability, migration 38, PostgreSQL error contracts. | Self and head-schema storage row. | Migration history and schema-object ownership projections. | Indicators with database-migrations/PostgreSQL collaboration. | High | Invalid legacy rows remain untouched at version 37; no test or migration performs silent deduplication. |
| `internal/modules/indicators/lifecycle_service.go` | Validates exact lifecycle membership, time/confidence/support refs, and atomically appends revisions/projections/idempotency. | `Store.AppendIndicatorLifecycleInterval`, list method through root facade. | HTTP adapter and tests. | Owner vocabulary, shared child coordination, Records, Revisions, projection port. | Child route, lifecycle, transaction and unit tests. | Lifecycle request/result and `record_changed` contracts. | Indicators mutation coordinator. | High | Service and HTTP admission now share one registry while retaining context-specific errors. |
| `internal/modules/indicators/network_flow_participant.go` | Narrow active Indicator lookup with ordinary and caller-transaction variants. | `Store.GetActiveIndicatorParticipant`, `Store.GetActiveIndicatorParticipantTx`. | Network Flow binding/store transaction participants. | pgx, Indicators and Records tables. | Network Flow and Indicators identity tests indirectly. | Network Flow `binding_only` Indicator reference. | Indicators participation port; Network Flow owns binding. | High | Must never create an observation or alter lifecycle implicitly. |
| `internal/modules/indicators/observation_origin_integration_test.go` | Verifies database origin constraints and producer-context behavior. | Test-only integration surface. | `module.indicators` origin row. | Postgres and observation paths. | Self. | Origin enum storage constraint. | Indicators. | High | Service-backed. |
| `internal/modules/indicators/observation_origin_test.go` | Unit characterization of the origin registry and public-route origin restrictions. | Test-only unit surface. | `module.indicators` origin row. | Internal origin and observation logic. | Self. | Origin vocabulary. | Indicators. | Medium | Complements database constraint evidence. |
| `internal/modules/indicators/observation_repository.go` | Persists, loads, transitions, and keyset-lists observations after service-level Records-owned target validation. | Private root repository. | Observation service. | pgx, identity/origin and Revisions error mapping. | Observation, child route, resolution, transaction tests. | Observation OpenAPI/portable row schemas. | Indicators persistence. | High | S-01 removed the standalone Records existence query; the repository consumes already-validated target facts. |
| `internal/modules/indicators/observation_service.go` | Captures verified UTF-8 source spans and coordinates resolve/dismiss/restore with revisions, projections, idempotency, and affected rows. | Store observation create/get/list/transition methods through root facade. | HTTP adapter, frontend typed operations, tests. | SourceText, Records, Revisions, authn idempotency, observation/source repositories. | Child route, resolution, transaction, span and unit tests. | Child operation envelopes and Collaboration intents. | Indicators mutation coordinator. | High | Observation creation remains distinct from canonical Indicator creation. |
| `internal/modules/indicators/persistence_scanners.go` | Owns the private Indicator persistence record and pgtype scanners/conversions for Indicator and observation rows. | Private root persistence surface. | Source and observation repositories. | uuid, pgtype, observation-origin registry. | Store, route, portability, and integration tests indirectly. | Database row shape only. | Indicators persistence. | High | Required-UUID and origin validation remain at the scanner boundary. |
| `internal/modules/indicators/portability_characterization_test.go` | Characterizes deterministic complete portable row bytes and round-trip posture. | Test-only integration surface. | `module.indicators` portability rows. | Incident Bundle provider and Postgres. | Self. | Bundle v2 paths/schemas. | Indicators. | High | Portable byte compatibility freeze. |
| `internal/modules/indicators/portability_final_state_test.go` | Verifies imported final state, active claims, attribution, and no partial results. | Test-only integration surface. | `module.indicators` portability rows. | Incident Bundle provider, Postgres. | Self. | Bundle final-state invariants. | Indicators. | High | Service-backed. |
| `internal/modules/indicators/portability_invariants_test.go` | Exercises strict shape, ordering, identity, lifecycle, observation, and partition invariants. | Test-only unit/integration surface. | `module.indicators` portability rows. | Incident Bundle provider/contracts. | Self. | Incident Bundle invariant IDs. | Indicators. | High | Rejects malformed or ambiguous portable state. |
| `internal/modules/indicators/production_contract_test.go` | Static/runtime checks for production route and contract completeness. | Test-only architecture/product surface. | `module.indicators` production workflow rows. | Target code, OpenAPI/view contracts. | Self. | Public child operations and system-view contract. | Indicators. | High | Ensures planned movement does not create a shadow route. |
| `internal/modules/indicators/projectionprovider/contribution.go` | Public adapter that constructs the Indicators source-owned projection contribution while hiding internal SQL provider. | `NewContribution`. | Projection application assembly. | Internal projection source and workbookprojection contract. | `contribution_test.go`, provider encapsulation tests. | Provider descriptor and surface intent. | Indicators contribution facade. | High | Approved facade between Indicators and Projections. |
| `internal/modules/indicators/projectionprovider/contribution_test.go` | Validates construction and delegation of the projection provider contribution. | Test-only unit surface. | `module.indicators` projection row. | Projection adapter/contracts. | Self. | Provider descriptor. | Indicators. | Medium | No projection-table implementation ownership transfer. |
| `internal/modules/indicators/provider_encapsulation_test.go` | Prevents generic coordinators from importing internal Indicators provider implementations. | Test-only architecture guard. | `module.indicators` provider-boundary row. | Go AST/parser and repository packages. | Self. | Provider facade packages. | Indicators architecture evidence. | Medium | Confirms owner constructs providers and coordinators consume contributions. |
| `internal/modules/indicators/recovery_state.go` | Fallibly projects catalog-owned authoritative and rebuildable relations into Recovery state facts. | `RecoveryStateContribution`. | Recovery application assembly. | Owner source-state catalog and platform recoverystate contract. | Source-state parity, active-identity, and recovery catalog tests. | Recovery source-state contribution. | Indicators source-state owner. | High | The relation inventory is no longer declared independently from portability facts. |
| `internal/modules/indicators/replay_codec.go` | Encodes and strictly decodes the stored create idempotency payload and its identity/version fields. | Private root replay codec. | Create service. | Standard JSON and UUID parsing. | `replay_codec_test.go`, create replay, route, and integration tests. | Stored idempotency JSON shape. | Indicators mutation coordination. | High | Split from the facade without changing stored bytes or status codes. |
| `internal/modules/indicators/replay_codec_test.go` | Freezes exact stored create JSON bytes and replay UUID/version decoding. | Test-only unit surface. | `module.indicators.surface.iteration2_exported_surface_reachability`. | Private replay codec. | Self. | Stored idempotency compatibility. | Indicators. | High | Exact byte fixture guards the cohesion-only move. |
| `internal/modules/indicators/repositories.go` | Declares private repository marker types. | Private root types only. | Root Store. | None. | Store/unit tests indirectly. | None. | Indicators persistence. | Low | Cohesion-only file. |
| `internal/modules/indicators/resolution_integration_test.go` | Verifies observation resolution/dismiss/restore state transitions and projection effects. | Test-only integration surface. | `module.indicators` observation/transition rows. | Postgres and Indicators Store. | Self. | Observation mutation semantics. | Indicators. | High | Service-backed transition characterization. |
| `internal/modules/indicators/revision_append_port.go` | Defines the narrow Revisions append/snapshot port and concrete adapter. | Private root port/adapter. | Store and all mutation services. | Revisions Appender/contracts. | Transaction, revision contribution, composition tests. | Change-set, mutation, revision, Collaboration intent behavior. | Indicators application boundary. | High | Keeps generic Revisions coordination outside source logic. |
| `internal/modules/indicators/revision_provider_contribution.go` | Declares Indicator row and child non-row revision/history/rollback/delete-restore contributions. | `NewRevisionContribution`. | Revision application assembly. | Internal rollback/delete-restore providers and Revisions contracts. | `revision_provider_contribution_test.go`, provider encapsulation tests. | Snapshot schema, target kinds, view route. | Indicators contribution facade. | High | Source owner constructs provider implementations. |
| `internal/modules/indicators/revision_provider_contribution_test.go` | Characterizes contribution IDs, snapshot schema, target kinds, history facets, and providers. | Test-only unit surface. | `module.indicators` revision row. | Revision contribution. | Self. | Revisions provider catalog. | Indicators. | Medium | Protects generic coordination mapping. |
| `internal/modules/indicators/source_repository.go` | Locks canonical dedupe identity and loads/upserts Indicator subtype state. | Private root repository. | Create, observation resolution, and Network Flow participation. | pgx; Indicators, active-identity, and Records relations. | Identity concurrency, create, transaction, active-claim tests. | Canonical identity and Records envelope contracts. | Indicators persistence with Records dependency. | High | S-01 removed the standalone validation query; the protected multi-table `loadByDedupeTx` lock join remains deliberately unchanged. |
| `internal/modules/indicators/source_state_contract_test.go` | Proves authored Incident Bundle descriptor parity, exact Recovery classification, deterministic projection bytes, and defensive sourceport copies. | Test-only unit surface. | `module.indicators.source_state.projection_parity`. | Indicators contribution constructors, Incident Bundle source catalog, Recovery contracts. | Self. | `contracts/incident-bundles/source_catalog.json` and Recovery contribution shape. | Indicators with Incident Bundles/Recovery collaboration. | High | Contract reading remains test-only; production depends solely on owner-internal facts. |
| `internal/modules/indicators/store_composition.go` | Composes the Store, its exact required dependencies, owner repositories/services, Records envelope port, incident admission, revisions, and projection helpers. | `Store`, `StoreDependencies`, `NewStore`; private ports/helpers. | Server/test assembly and all Store method callers. | Postgres, authn, Incidents admission, Records, Revisions, workbook projection port. | `store_composition_test.go`, export guard, provider boundary, server build. | Root construction and import-topology boundary. | Indicators application facade. | High | Constructor path and dependency requirements remain unchanged after the split. |
| `internal/modules/indicators/store_composition_test.go` | Verifies required Store dependencies, clock behavior, and projection delegation. | Test-only unit surface. | `module.indicators` structure/projection rows. | Store constructor and fakes. | Self. | Store composition contract. | Indicators. | Medium | Required before constructor/facade changes. |
| `internal/modules/indicators/store_test_helpers_test.go` | Reusable same-package fixtures and fakes for Store tests. | Test-only private helpers. | Target test files. | Postgres/testcontainers, auth and projection fakes. | Many target tests. | None. | Indicators test support. | Medium | Test-only assumptions must not leak into production. |
| `internal/modules/indicators/target_resolution_integration_test.go` | Exercises source/requested/addressed/prior/support target-role failures and proves durable footprint atomicity. | Test-only integration surface. | `module.indicators.target_resolution.role_failures_are_atomic`. | Indicators Store, Records envelopes, PostgreSQL transaction fixture. | Self. | Core 01/Core 04 resolution and concealment behavior. | Indicators with Records collaboration. | High | Added in S-01; each semantic/storage mismatch has deterministic no-side-effect evidence. |
| `internal/modules/indicators/testsupport/fixtures.go` | Cross-owner deterministic Indicator examples and direct test seeding helpers. | `Example`, `BaseTime`, `PastTime`, `Examples`, payload/dedupe/seed helpers. | Other modules' integration tests and route fixtures. | Internal identity, Records testsupport, SQL. | Cross-owner tests. | No production contract; test fixtures only. | Indicators test support. | Medium | Legitimate semantic fixture owner; not a production API. |
| `internal/modules/indicators/testsupport/routetest/inventory.go` | Supplies the generic Indicators view-query control route to shared route inventory tests. | `ControlQuery`. | Shared route inventory/conformance tests. | Indicators `ViewSchemaID`, platform test route inventory. | Cross-owner route tests. | Generic Workbook query route. | Indicators test support with Workbook route owner. | Medium | Evidence accounting only. |
| `internal/modules/indicators/transaction_atomicity_test.go` | Proves rollback of source writes, revisions, projections, idempotency, and transitions on failure. | Test-only integration surface. | `module.indicators` transaction rows. | Postgres and fault fakes. | Self. | Atomic mutation and Collaboration intent behavior. | Indicators/Revisions/Records collaboration. | High | Required before transaction seam changes. |
| `internal/modules/indicators/unit_test.go` | Pure unit characterization for create admission, observation spans/transitions, lifecycle vocabulary, and helper behavior. | Test-only unit surface. | Several `module.indicators` unit rows. | Root Store/services and fakes. | Self. | Core 02/01 behavior projection. | Indicators. | High | Cheapest characterization layer. |
| `internal/modules/indicators/value_serialization.go` | Owns snapshot/version identifiers, projection value maps, pointer cloning, UTC formatting, UUID lists, and semantic map equality. | Private root value-serialization surface. | Create, observation, lifecycle, repository, and participant code. | Standard time/UUID/reflection only. | Replay, route, transaction, portability, and unit tests indirectly. | Revision snapshots, projection values, public success rows. | Indicators domain/application support. | High | Same functions and value shapes moved without a compatibility layer or byte change. |
| `internal/modules/indicators/workbookprojection/contribution.go` | Defines source-owned typed projection input, Projections row/rebuild ports, descriptor, and 13-field surface intent. | `SourceReader`, `ProjectionInput`, `ProjectionInputPage`, `Rows`, `Rebuilder`, `Ports`, `Contribution`, `NewContribution`, `Descriptor`, `SurfaceIntent`. | Projection assembly and Indicators Store dependency. | Projections provider contract and pgx transaction type. | `contribution_test.go`, projection integration tests. | `indicator_grid_projection`, `cartulary.view.indicators.v1`. | Indicators semantic contract; Projections storage owner. | High | Approved facade; contains no projection-table SQL. |
| `internal/modules/indicators/workbookprojection/contribution_test.go` | Characterizes descriptor ownership, capabilities, field order, and defensive contribution behavior. | Test-only unit surface. | `module.indicators` projection row. | Workbook projection contribution. | Self. | Projection descriptor/surface intent. | Indicators. | Medium | Guards the explicit Links/Records authority declaration. |

## 3. Module Boundary Diagnosis

The live target is a mixed-responsibility package tree, but current evidence does not support calling it an accidental catch-all. Its root is a legitimate Indicators application/service facade and mutation coordinator; its subpackages are transport-adjacent, persistence-adjacent, projection-source, revision, portability, recovery, and test-support seams constructed by the Indicators source owner. The existence of the directory is not treated as proof that these are permanent package locations.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Canonical Indicator identity, normalization, dedupe, and source subtype | Root plus `internal/identity` and `source_repository.go` | Indicators | keep | Core 02 §§10.2, 17, 18; identity tests; active-claim tests | Central domain responsibility. |
| Source-bound Indicator observations | Root observation service/repository | Indicators | keep | Core 02 observation/canonical distinction; Core 01 child routes; Core 03 source-bound workflow | Must remain distinct from canonical create even in one analyst action. |
| Append-only Indicator lifecycle intervals | Root lifecycle service/repository | Indicators | keep | Core 02 closed lifecycle vocabulary and history; Core 01 child routes | Revisions owns generic history coordination, not lifecycle meaning. |
| Mutation transaction coordination | Root Store/services/child coordination | Indicators | split | Atomic source write, Records version, Revisions, projection, idempotency, and intent tests | Split by cohesive behavior inside the module; do not move generic ownership inward. |
| Child HTTP transport and authorization adapter | `httpapi` | Indicators owner-local transport adapter under adopted `IND-BOUNDARY-001` | keep | Core 01 route family; Core 04 REQ-04-150; server assembly | Current subpackage is explicit and tested. Relocation is forbidden in S-00 through S-06 and requires a later superseding decision. |
| Workbook row-create transport adapter | `internal/app/workbookassembly/indicator_adapter.go` outside target | Workbook application assembly | keep | Generic Workbook create route and live assembly | Correctly outside the source module. |
| Projection source semantics and typed inputs | `workbookprojection`, `projectionprovider`, internal projection source | Indicators | keep | Core 01 REQ-01-622/658; descriptor names Indicators source owner | Source owner retains derivation meaning. |
| Projection table storage/query/rebuild mechanics | `internal/modules/projections` outside target | Projections | keep | Descriptor names Projections storage owner; no projection table SQL in public Indicators contribution | No move into Indicators. |
| Supporting-link count in Indicator projection | Internal projection source reads `active_record_links_v1` | Indicators semantic input with Links source authority | keep | Descriptor declares `links`; Core 01 source-authority split | Intentional/no action unless owner contract changes. |
| Import parsing, sessions, mapping, and dispatch | Imports outside target | Imports | keep | Core 00 REQ-00-066; live owner registry | Indicators owns only its create facade. |
| Indicators import create semantics | `import_create.go` | Indicators | keep | Import target `module.indicators@1`; owner facade composition | Caller transaction preserves generic Import orchestration. |
| Revision/change-set coordination and Collaboration publication | Revisions/Collaboration outside target | Revisions and Collaboration | keep | Core 00 REQ-00-067; Core 03 collaboration application; narrow revision port | Indicators contributes source mutations and semantic intents only. |
| Indicator row and child rollback semantics | Internal providers plus public contribution | Indicators | keep | Core 02 target-specific reversible history; provider encapsulation test | Generic Revisions coordinator must not import implementation internals. |
| Incident Bundle orchestration | Incident Bundles outside target | Incident Bundles | keep | Core 01 portability owner; sourceport boundary | Indicators owns only its three source families and validation. |
| Indicator portability rows and paths | Internal Incident Bundle provider | Indicators | split | Three source paths and family invariants | Retain ownership; consolidate duplicated source inventory internally. |
| Network Flow table/binding behavior | Network Flow outside target | Network Flow | keep | Adopted Network Flow §15 | Binding-only semantics remain outside Indicators. |
| Canonical create/dedupe participant used by Network Flow | `network_flow_participant.go`, create service | Indicators | keep | NF-REQ-142 through NF-REQ-144b | Must not create observations or lifecycle entries. |
| Recovery authoritative/derived state declaration | `recovery_state.go` | Indicators contribution to Recovery | split | Three authoritative tables plus one rebuildable relation | Consolidate source inventory; Recovery retains orchestration. |
| Saved views and generic view-query behavior | Workbook/Saved Views/Projections outside target | Workbook, Saved Views, Projections | keep | Core 01/03; target only supplies view schema/projection input | Indicators must not create a private saved-view implementation. |
| Frontend inspector/controller state | `apps/web` outside target | Web Workbook | keep | Core 03 REQ-03-306; typed operation port files | No frontend production code exists under the target. |
| Grid-vendor integration | `packages/grid-adapter` outside target | Grid Adapter | keep | Direct `react-data-grid` search finds vendor imports only in adapter | No direct vendor coupling in Indicators backend or indicator workflow files. |
| Test fixtures and route inventory | `testsupport` | Indicators test support | keep | Cross-owner test imports and harness rows | Test evidence does not define runtime architecture. Production packages MUST NOT import this package. |

### Adopted decision IND-BOUNDARY-001

`IND-BOUNDARY-001 — Indicators owner-local implementation topology v1` is adopted by `docs/decisions/indicators-module-boundary.md`, Core 00 `REQ-00-074`, Core 04 `AC-560`, the Base claim manifest, and executable root-export/import-topology evidence. This tracker records execution only and MUST NOT be cited as adoption authority.

W-01 re-scanned the live Core files immediately before editing and confirmed that `REQ-00-074` and `AC-560` were unallocated. They are now assigned to Indicators boundary adoption; later work MUST NOT reuse or renumber them.

The adopted `IND-BOUNDARY-001` requires all of the following:

1. `internal/modules/indicators` remains the source-owner implementation boundary for canonical Indicator identity and normalization, source-bound observations, lifecycle intervals, owner-side create/reuse/resolve/dismiss/restore/append behavior, caller-transaction participation, and construction of Indicators-owned coordinator contributions.
2. The root `indicators` package remains the narrow internal Go facade and owner transaction-coordination boundary. Its import path remains unchanged. Its 55-symbol guarded export baseline contracts to the reviewed 50-symbol surface only in S-04; other export contraction or expansion requires separate authorization.
3. `admission` remains transport-neutral. It MAY depend on Indicators request vocabulary and generated view-schema contracts; it MUST NOT depend on HTTP sessions, CSRF, routers, cookies, or transport response types.
4. `httpapi` remains an owner-local edge adapter. It MAY depend on HTTP, authentication, authorization, pagination, session, and common-envelope facilities; Indicators domain, repository, identity, and admission code MUST NOT import it. Public authorization and error mapping terminate at this adapter.
5. Indicators-owned Incident Portability, Reporting, Recovery, Revisions, Projections, and related provider implementations remain owner-local. Generic coordinators MUST consume narrow contributions or ports and MUST NOT import owner-internal providers. Application assembly remains the composition point.
6. The Imports-facing create facade remains Indicators-owned and caller-transaction based. Imports retains ingest session, unit, mapping, job, and finalization orchestration; Indicators retains canonical Indicator create/reuse semantics.
7. Network Flow continues to consume only the narrow Indicators transaction-participation contract. Network Flow retains binding state; Indicators retains canonical identity; a binding MUST NOT become an Indicator observation.
8. `testsupport` remains test-only. Production code MUST NOT import it, and no compatibility facade, forwarding package, or production alias may expose it.
9. Indicator browser workflow remains under `apps/web`; grid-vendor integration remains under `packages/grid-adapter`. Backend Indicators code MUST NOT absorb browser controller state, and application feature code MUST NOT bypass the grid adapter for vendor imports.
10. A later cross-owner relocation MUST be authorized by a new adopted decision naming source and destination owners, import direction, compatibility consequences, assembly changes, verification ownership, and rollback or migration handling.

| Concern | Required placement after adoption | Required dependency rule |
| --- | --- | --- |
| Public internal-Go facade and DTOs | Indicators root | MUST expose only stable owner operations and contribution constructors; the live 55-symbol surface contracts to exactly 50 reviewed declarations in S-04. |
| Canonical identity and normalization | Indicators owner-internal identity and repository seams | MUST remain transport- and coordinator-neutral. |
| Observation and lifecycle behavior | Indicators owner packages | MAY borrow Records, Revisions, and Projections ports; MUST NOT own those subsystems. |
| JSON admission and request hashing | `indicators/admission` | MUST remain transport-neutral. |
| HTTP routes and public error adaptation | `indicators/httpapi` | MUST remain an owner-local edge adapter with no reverse import. |
| Workbook projection contribution | Indicators contribution package | MUST supply semantic source input; Projections retains physical storage, query, and rebuild ownership. |
| Portability, Reporting, Recovery, Revisions providers | Owner-local implementations behind contribution facades | Generic coordinators MUST consume only public contribution contracts. |
| Import create facade | Indicators root | MUST borrow the Imports-owned transaction and MUST NOT acquire ingest orchestration. |
| Test helpers | `indicators/testsupport` | MUST have test importers only. |
| Browser workflow | `apps/web` | MUST consume generated protocol and application-owned handlers. |
| Grid-vendor binding | `packages/grid-adapter` | Application feature code MUST NOT import the vendor directly. |

| Required content | Normative authority after adoption | This tracker may do only |
| --- | --- | --- |
| Durable Indicators ownership and cross-owner direction | Core 01 and applicable Core 02 semantics | Map and preserve the owner requirements. |
| Exact package topology and compatibility posture | `docs/decisions/indicators-module-boundary.md`, registered by `REQ-00-074` | Record implementation and validation evidence without redefining it. |
| Boundary conformance | `AC-560` plus authored machine policy and owner-routed tests | Record required acceptance evidence. |
| Public HTTP, Inspector, security, history, projection, and portability behavior | Core 01 through Core 04 and adopted subsystem owners | Freeze and trace behavior; never redefine it. |
| Exact browser row, selector, runtime, and fixture routing | Authored test-family and verification inputs under the Testing Harness NLSpec | Define the conditional row contract; never treat it as product authority. |
| Allowed paths, start/stop gates, and result roots | Later implementation authorization and this handoff | Define the authorization charter and record evidence. |

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Generic create `POST /api/v1/incidents/{incident_id}/views/cartulary.view.indicators.v1/rows` | Workbook route; Indicators create semantics | Core 01 route/view schema; Workbook adapter; admission/create service | `indicators_test.go`, admission and route tests | Preserve current create/reuse/update/replay cases before facade movement | High | First create `201`; replay/reuse/update status and payload remain exact. |
| Generic query `POST /api/v1/incidents/{incident_id}/views/cartulary.view.indicators.v1/query` | Workbook/View Query/Projections | Core 01 query and cursor owner; test route inventory | `indicators_test.go`; shared route inventory | Preserve query, row identity, sort/filter/paging if projection input changes | High | Indicators does not own the generic route. |
| `GET/POST /api/v1/records/{source_record_id}/indicator-observations` | Indicators route family | Core 01 REQ-01-652/654; Core 04 REQ-04-150; owner OpenAPI | `child_routes_integration_test.go`; frontend Vitest rows | Add only if a newly moved authorization/error branch lacks direct evidence | High | Create requires verified UTF-8 source-field byte span. |
| `GET /api/v1/indicators/{indicator_id}/observations` | Indicators | Same owners and OpenAPI | Child route and frontend paging tests | No additional characterization is required for same-package structural movement; any query or cursor change requires new evidence. | High | Active observations newest first; opaque actor/route/record/limit-bound cursor. |
| `POST /api/v1/indicator-observations/{observation_id}/{resolve,dismiss,restore}` | Indicators | Core 01/02/04; owner OpenAPI | Child route, resolution, transaction, frontend operation tests | Preserve exact replay-before-transition and stale-version cases | High | Three operations, closed request bodies, `200` success/replay. |
| `GET/POST /api/v1/indicators/{indicator_id}/state-intervals` | Indicators | Core 01/02/04; owner OpenAPI | Child route, lifecycle, unit and frontend tests | Preserve interval ordering and append replay/validation | High | Append first success `201`, replay `200`. |
| Child request/result envelopes | Indicators route family; common envelope owner in Core 01 | OpenAPI schemas and `contracts.go` DTOs | Production contract, child route, frontend port tests | Characterize any DTO relocation by exact JSON shape | High | Mutation result contains child resource, `change_set_id`, `replayed`, sorted unique `affected_records`. |
| Authorization and concealment | Core 04; Indicators adapter applies it | REQ-04-002/003/021-023/150; `httpapi/routes.go` | Child route role/session/CSRF tests | Add negative precedence tests only if handler admission is reorganized | High | Session, CSRF, path/visibility, role precede body/idempotency/child lookup. |
| Canonical Indicator normalization/dedupe | Core 02; Indicators implements | `internal/identity`; active claim relation | Identity unit/concurrency/portability tests | Preserve every token and type-specific canonicalization vector | High | Active claims are rebuildable coordination state, not source truth. |
| Source observation semantics | Core 02; Indicators implements | Observation service/repository and origin registry | Origin, span, resolution, child route tests | Preserve raw source field and exact observed substring | High | Observation occurrence is never collapsed into canonical Indicator identity. |
| Lifecycle vocabulary | Core 02; Indicators implements | Four tokens in Core/OpenAPI/service/decoder | Unit, OpenAPI alignment through production tests, child routes | Add one shared-registry parity test before deduplication | High | Exact current tokens: `active`, `benign`, `false_positive`, `retired`. |
| Observation state machine | Core 02; Indicators implements | Service/repository, OpenAPI and portability schema | Unit, resolution integration, rollback, portability tests | Add shared-registry parity test if validation is centralized | High | Exact states `unresolved`, `resolved`, `dismissed`. |
| `cartulary.view.indicators.v1` | Core 01; view-schema projection | Authored view schema | Contract tests, Indicators integration, frontend contract tests | Preserve schema ID and all 13 field keys/order/capabilities | High | Canonical Indicator is row identity; all existing-row fields are immutable. |
| Create-only Indicator fields | Core 01/03 | View schema and create admission | Admission, create and frontend surface tests | Existing evidence plus exact field-registry comparison if refactored | High | Eight create-writable fields; zero-field create forbidden. |
| Inspector tuples | Core 01/03 | `contracts/view-inspector/index.json`, view schema, handlers | Two module-owned Vitest rows | Conditional `IND-BROWSER-001` before any frontend shell change | Medium | Four exact tuples MUST resolve before generic record patch. |
| Saved-view behavior over Indicators | Saved Views/Workbook | Core 03 fallback/persistence owner | Shared Workbook/saved-view tests | No Indicators-specific addition unless source refactor changes query results | Medium | System view remains directly addressable without a saved-view object. |
| Projection refresh/load/rebuild | Indicators source semantics; Projections storage | Descriptor, source input, Projections ports | Projection contribution/delegation, integration, recovery tests | Preserve 13 field values and transaction ordering | High | Derived state; never source truth. |
| Supporting-link count | Links source semantics, Indicators projection input | Descriptor declares `links`; source SQL | Projection tests/integration | Characterize exact count if source SQL is replaced by a port | Medium | Current direct declared source read is intentional. |
| Revision/change-set/history | Revisions generic owner; Indicators target semantics | Revision port/contribution and Core 02 | Revision contribution, child routes, rollback, transaction tests | Preserve target kinds, snapshot schema, ordering, rollback selectors | High | Each affected first-class record advances once in sorted lock order. |
| Collaboration `record_changed` | Collaboration transport; Revisions intent sequencing; Indicators contributes mutations | Core 03 REQ-03-091/096-098 and implementation note | Child route intent assertions, transaction atomicity; shared Collaboration tests | Preserve one intent per committed affected record and none on replay/rollback | High | No child-specific WebSocket message exists. |
| Import owner facade | Imports orchestration; Indicators create semantics | Import target registry and `import_create.go` | Indicators and Imports owner rows | Preserve field mapping, caller transaction, result codes and row payload | High | No new import parser or session behavior in Indicators. |
| Network Flow participant | Network Flow binding; Indicators canonical identity | Adopted NF §15, narrow public port | Identity concurrency and Network Flow tests | Preserve same-transaction create/reuse and `binding_only` no-observation rule | High | The public internal-Go interface MUST remain unchanged in S-00 through S-06; a later change requires separate compatibility authorization and complete consumer migration. |
| Incident Bundle portability | Core 01 orchestration; Indicators source rows | Source catalog and three schemas/paths | Characterization/final-state/invariant tests | Preserve deterministic version-2 bytes and all invariants | High | Paths: Indicators, observations, state intervals NDJSON. |
| Recovery contribution | Recovery orchestration; Indicators state classification | `recovery_state.go` | Active-identity and Recovery catalog tests | Preserve authoritative/derived classification and rebuild behavior | High | `indicator_active_identities` remains excluded from portability. |
| Generated OpenAPI/protocol/view contracts | Authored owner contracts upstream; generated outputs downstream | OpenAPI owner source, view-schema contracts, generated policy | Generation and JSON-shape gates | Run drift/policy/shape checks only if authored inputs change | High | Never hand-edit generated roots. |
| Harness/test accounting | Testing Harness mechanics; `module.indicators` semantic owner | Verification contract and family manifest | 19 verification entries, 26 active rows | Update authored rows only for actual moves/new evidence | Medium | 24 Go, 2 Vitest; 15 unit and 11 integration/service-backed. |
| Browser evidence | Browser harness and owning product rows | Live test-family and execution-topology manifests | No `module.indicators` Playwright row | Conditional `IND-BROWSER-001` | Medium | Backend-only S-00 through S-06 do not require a browser row. The activation event is any Indicator frontend, Inspector, generated frontend-contract consumption, handler-routing, or grid-adapter diff. |

### Frozen child HTTP operation map

| Method and path | Operation ID | Admitted response statuses | Mutation default |
| --- | --- | --- | --- |
| `GET /api/v1/records/{source_record_id}/indicator-observations` | `listSourceRecordIndicatorObservations` | `200`, `400`, `401`, `403`, `404` | Read-only; no idempotency effect. |
| `POST /api/v1/records/{source_record_id}/indicator-observations` | `createManualIndicatorObservation` | `200`, `201`, `400`, `401`, `403`, `404`, `409` | First commit `201`; exact replay `200`. |
| `GET /api/v1/indicators/{indicator_id}/observations` | `listIndicatorObservations` | `200`, `400`, `401`, `403`, `404` | Read-only; no idempotency effect. |
| `POST /api/v1/indicator-observations/{observation_id}/resolve` | `resolveIndicatorObservation` | `200`, `400`, `401`, `403`, `404`, `409` | First commit and exact replay both `200`; replay has no second effect. |
| `POST /api/v1/indicator-observations/{observation_id}/dismiss` | `dismissIndicatorObservation` | `200`, `400`, `401`, `403`, `404`, `409` | First commit and exact replay both `200`; replay has no second effect. |
| `POST /api/v1/indicator-observations/{observation_id}/restore` | `restoreIndicatorObservation` | `200`, `400`, `401`, `403`, `404`, `409` | First commit and exact replay both `200`; replay has no second effect. |
| `GET /api/v1/indicators/{indicator_id}/state-intervals` | `listIndicatorStateIntervals` | `200`, `400`, `401`, `403`, `404` | Read-only; no idempotency effect. |
| `POST /api/v1/indicators/{indicator_id}/state-intervals` | `appendIndicatorStateInterval` | `200`, `201`, `400`, `401`, `403`, `404`, `409` | First commit `201`; exact replay `200`. |

The six path shapes, eight operation IDs, request/response schemas, admitted statuses, cursor binding, authorization precedence, and replay outcomes MUST remain unchanged. This table records the authored owner surface and does not add a route.

### Frozen `cartulary.view.indicators.v1` field surface

Every field is `default_hidden=false`, `grid_editable=false`, `writable=false`, and `clearable=false`. Existing rows are therefore immutable through the generic patch surface. `create_writable=true` applies only to create admission and does not imply existing-row writability. Field order is the table order and MUST remain unchanged.

| Field key | Read model / kind | Create-writable | Sort / filter / group |
| --- | --- | --- | --- |
| `indicator.indicator_type` | `indicator_type` / `text` | yes | sort; `eq`,`prefix`; group |
| `indicator.value_kind` | `value_kind` / `text` | yes | sort; `eq`,`prefix`; group |
| `indicator.display_value` | `display_value` / `text` | yes | sort; no filter; no group |
| `indicator.normalized_value` | `normalized_value` / `text` | yes | sort; no filter; no group |
| `indicator.defanged_value` | `defanged_value` / `text` | yes | sort; no filter; no group |
| `indicator.hash_algorithm` | `hash_algorithm` / `text` | yes | sort; `eq`,`prefix`; no group |
| `indicator.hash_value` | `hash_value` / `text` | yes | sort; no filter; no group |
| `indicator.stix_pattern` | `stix_pattern` / `text` | yes | sort; no filter; no group |
| `indicator.first_observed_at` | `first_observed_at` / `timestamp` | no | sort; `eq`,`range`; no group |
| `indicator.last_observed_at` | `last_observed_at` / `timestamp` | no | sort; `eq`,`range`; no group |
| `indicator.observation_count` | `observation_count` / `number` | no | sort; no filter; no group |
| `indicator.lifecycle_summary` | `lifecycle_summary` / `text` | no | sort; `eq`,`prefix`; group |
| `indicator.supporting_link_count` | `supporting_link_count` / `number` | no | sort; no filter; no group |

### Frozen Indicator Inspector operation tuples

| View / feature | Panel | Route binding / owner | Role / mutation | Disabled conditions | Success / failure behavior |
| --- | --- | --- | --- | --- | --- |
| `cartulary.view.timeline.v2` / `indicator.observations.manage` | `relationships` | `indicator_observations` / `indicator_observations_route` | `editor`; mutates; no confirmation | no row, incident closed, authorization lost, row version changed, record deleted | preserve selection / invalidate pending action in same-shell error |
| `cartulary.view.indicators.v1` / `indicator.observations.pivot` | `relationships` | `indicator_observations` / `indicator_observations_route` | no minimum role beyond route read authority; read-only | no row, authorization lost, record deleted | preserve selection / preserve selection in same-shell error |
| `cartulary.view.indicators.v1` / `indicator.lifecycle.read` | `history` | `indicator_lifecycle` / `indicator_lifecycle_route` | no minimum role beyond route read authority; read-only | no row, authorization lost, record deleted | preserve selection / preserve selection in same-shell error |
| `cartulary.view.indicators.v1` / `indicator.lifecycle.manage` | `history` | `indicator_lifecycle` / `indicator_lifecycle_route` | `editor`; mutates; no confirmation | no row, incident closed, authorization lost, row version changed, record deleted | preserve selection / invalidate pending action in same-shell error |

All four tuples have empty seed bindings. They MUST resolve before any generic record-patch fallback. A missing tuple, changed route owner, widened role, changed disabled condition, or different selection behavior is an observable contract change requiring later authorization.

### Frozen source-state and Incident Bundle map

| State kind | Relation or logical path | Schema / rebuild identity | Stable identity or rule |
| --- | --- | --- | --- |
| Authoritative relation | `indicators` | source-owned database relation | Canonical Indicator source row. |
| Authoritative relation | `indicator_observations` | source-owned database relation | Observation occurrence; repeated observations remain distinct. |
| Authoritative relation | `indicator_state_intervals` | source-owned database relation | Append-only lifecycle interval source row. |
| Rebuildable relation | `indicator_active_identities` | `indicators.restore_active_identities.v1` | Derived active canonical-claim coordination state; excluded from portability. |
| Portability path, version `2` | `data/indicators.ndjson` | `cartulary.incident_bundle.indicators.row.v1` | `record_id` |
| Portability path, version `2` | `data/indicator_observations.ndjson` | `cartulary.incident_bundle.indicator_observations.row.v1` | `indicator_observation_id` |
| Portability path, version `2` | `data/indicator_state_intervals.ndjson` | `cartulary.incident_bundle.indicator_state_intervals.row.v1` | `indicator_state_interval_id` |

The Incident Bundle family is `indicators`, the owner is `module.indicators`, the owner relation is `indicator-source`, and the dependency is `entities`. S-03 MUST preserve exact path ordering, version sets, schema IDs, stable identities, deterministic bytes, and the authoritative/rebuildable distinction.

### IND-RECORDS-001 — caller-transaction target resolution

The live Records capability is:

```text
RouteTargetResolver.ResolveTx(
    context.Context,
    pgx.Tx,
    uuid.UUID,
) (records.RouteTarget, error)
```

| Interface element | Exact contract |
| --- | --- |
| Context | Caller-supplied cancellation and deadline context. The resolver MUST NOT detach from it. |
| Transaction | Caller-owned `pgx.Tx`. The resolver MUST NOT begin, commit, roll back, nest, replace, or retain it. |
| Input identity | One `uuid.UUID` record ID. No view-schema, relation, column, or SQL-fragment input is permitted. |
| Success output | `records.RouteTarget{IncidentID, RecordType, Deleted, RowVersion}` and no additional source or authorization facts. |
| Deleted row | Returns the tuple with `Deleted=true` to the trusted caller. |
| Missing row | Returns the Records missing-envelope outcome, currently `records.ErrEnvelopeNotFound`. |
| Storage failure | Returns a classified or wrapped storage failure; no SQL, relation, constraint, or value may reach a public response. |
| Locking | The current implementation performs a non-locking envelope read. S-01 MUST NOT treat it as the mutation linearization point. |
| Forbidden effects | No authorization, incident-membership decision, Indicator type policy, HTTP mapping, history, projection, Collaboration publication, idempotency write, network call, or transaction-lifecycle action. |

Both direct SQL reads targeted by S-01 occur inside `observationRepository.insertTx`, after `indicatorObservationService.createManualObservation` has called `lockAffectedRecordsTx` for the sorted affected record IDs and revalidated active same-incident envelopes. This live order supersedes the analysis note's generic description of the standalone read as an earlier admission read. S-01 MUST replace each read at its existing position and MUST NOT move either read before the existing deterministic lock/recheck sequence.

### IND-ERROR-001 — target-resolution mapping

| Indicators target | Records or owner-local fact | Required Indicators result | Public result | Forbidden distinction |
| --- | --- | --- | --- | --- |
| Observation source record | Missing, deleted, foreign incident, or unsupported source contract | Existing source-unavailable classification | `404 indicator_source_record_not_found` | Existence, deletion, incident, record type, source-field support |
| Resolved Indicator record | Missing, deleted, foreign incident, or `RecordType != "indicator"` | `ErrResolvedIndicatorNotFound` on the observation route | `404 resolved_indicator_not_found` | Existence, deletion, incident, actual type |
| Indicator addressed by observation/lifecycle route | Missing, deleted, foreign incident, or wrong type | `ErrIndicatorNotFound` | `404 indicator_not_found` | Existence, deletion, incident, actual type |
| Any target | Storage failure | Existing safe internal failure path | Existing safe internal error envelope | SQL text, relation, constraint, driver value, hidden identifier |

These mappings MUST preserve Core 04 admission precedence. A failed target resolution MUST commit no source row, history, projection, idempotency success, change-set effect, or Collaboration intent.

### IND-BROWSER-001 — conditional Indicator Inspector evidence

`IND-BROWSER-001` is inactive for backend-only S-00 through S-06. Any diff under `apps/web`, `packages/grid-adapter`, generated frontend protocol consumption, Indicator Inspector routing, or visible Indicator pending/empty/error/retry/paging state activates it before merge.

| Row field | Required value when activated |
| --- | --- |
| Owner | `module.indicators` |
| Collaborators | `harness.browser`, `web.workbook` |
| Semantic family | `module.indicators.inspector_workflows` |
| Runner and project | `playwright`; `chromium` |
| Stage | `webserver_backed` |
| Runtime and resource profiles | `default`; `browser_functional` |
| Fixture and services | `browser_stack`; `postgres` and `object_store` |
| Evidence and claim posture | `browser`; `implementation` |
| Status and skip default | `active`; skip policy `forbid` |
| Selector | One exact repository-relative Playwright file, five nonempty stable scenario IDs, and five matching exact titles. Globs and regular expressions are forbidden. |
| Verification ownership | Indicator route and semantic verification IDs, with applicable Workbook interaction verification represented through collaboration rather than ownership transfer. |

The single row MUST select five independently reportable scenarios:

1. A Timeline row with a supported source field opens observation creation, invokes the source-record Indicator-observation route rather than generic record patch, preserves source text, and preserves selection.
2. An Indicator row resolves `indicator.observations.pivot` to the Indicator observation collection route and renders accessible loading, populated, empty, error, retry, and next-page states when those states apply.
3. An Indicator row resolves `indicator.lifecycle.read` to the state-interval collection route and represents server ordering and paging without client redefinition.
4. An editor resolves `indicator.lifecycle.manage` to interval append, observes committed state without leaving the Workbook or losing selection, and a viewer is not offered the mutation action.
5. A hidden, foreign-incident, deleted, or otherwise unavailable target discloses no type, incident, deletion state, or existence through UI copy, action availability, or selected route.

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Two owner-local observation-insertion validation paths read Records envelopes through standalone SQL after the affected envelopes are already locked. | `source_repository.go::validateIncidentTx`; `observation_repository.go::validateSourceIncidentTx`; `observation_service.go` calls `lockAffectedRecordsTx` first; Core 00 REQ-00-067; `IND-RECORDS-001`. | Envelope authority, concealed-error mapping, and statement-order drift. | `must_fix` | Records capability consumed by Indicators | S-01 MUST replace only these reads in place and MUST preserve `IND-ERROR-001` and the existing lock/recheck order. |
| `IND-DEDUPE-001`: canonical dedupe lookup locks active identity, Indicator subtype, and Records envelope in one compound query. | `source_repository.go::loadByDedupeTx`; identity concurrency and active-claim tests. | Splitting it prematurely could introduce inconsistent snapshots, lock inversion, delete/restore races, or duplicate canonical claims. | `defer` | Indicators with the Records transaction boundary | S-01 MUST leave the query semantically unchanged. A future replacement MUST satisfy the decision gates below. |
| Projection input directly reads Records and Links authoritative state. | `workbookprojection.Descriptor` declares `source_authority_modules=[indicators,links,records]`; Core 01 REQ-01-622/658. | A generic facade could obscure source derivation or break aggregation. | `intentional/no_action` | Indicators semantic provider with Links/Records authority | Preserve the declared authorities; reconsider only with an adopted provider-boundary change. |
| Indicator type validation appears in identity logic and HTTP decoder; lifecycle and observation-state tokens appear in several runtime/provider paths. | `internal/identity/identity.go`, `httpapi/decoding.go`, lifecycle/observation services, rollback and portability providers, authored schemas. | Token drift across transport, live state, rollback, and import. | `should_fix` | Indicators | Introduce one owner-internal vocabulary registry; keep authored schemas/OpenAPI aligned through generators and tests. |
| Recovery relations and portability paths/state families are declared in separate code locations. | `recovery_state.go`, Incident Bundle `source_port.go`, provider SQL, source catalog. | Added/removed relation can drift across backup, portability, and rebuild. | `should_fix` | Indicators source-state catalog | Create one immutable validated internal catalog and project contribution inventories from it. |
| Root exported facade mixed public DTOs/errors with private JSON replay/value serialization and Store mechanics. | S-04 cohesive root files; exact 50-symbol AST allowlist and role inventory. | Broad edits risk public internal-Go consumers and wire serialization. | `fixed` | Indicators | S-04 split the files by concern and removed the unused five-symbol outcome enum without a shim; exact bytes and consumers are covered. |
| Store constructs platform authn and incident admission helpers from Postgres. | `NewStore` in `store_composition.go`. | Concrete dependency coupling may impede narrow testing/ownership, but current composition is characterized. | `defer` | Indicators/application assembly | Decide only in a later facade-design task; do not expand public ports speculatively. |
| HTTP adapter owns transport/auth concerns but is nested under the module. | `httpapi` imports platform HTTP/auth/pagination; root admission guard remains transport-neutral. | Moving it could churn route composition without behavioral value. | `defer` | Indicators adapter or application edge | Require an Indicators-specific implementation-boundary decision before relocation. |
| Internal revision and portability providers are encapsulated behind source-owner contributions. | Provider encapsulation test and application assemblies. | Moving implementations to generic coordinators would invert ownership. | `intentional/no_action` | Indicators | Retain provider construction within the owner. |
| Imports orchestration is outside the target; only owner create semantics are inside. | Import target registry and import assembly. | Treating the facade as Imports logic would transfer source ownership. | `intentional/no_action` | Imports orchestration / Indicators creation | Preserve caller-transaction owner facade. |
| Network Flow imports only the narrow Indicator participation contract and owns binding storage. | Network Flow ports/transaction participant; adopted NF §15. | Moving binding logic into Indicators would conflate observations and bindings. | `intentional/no_action` | Network Flow and Indicators | Preserve `binding_only` semantics and same-transaction participation. |
| No production frontend or grid-vendor code is under the target; indicator workflows use typed contracts/operations. | Frontend source search; `react-data-grid` imports remain in `packages/grid-adapter`. | Cross-layer refactor could accidentally widen scope. | `intentional/no_action` | Web Workbook / Grid Adapter | Keep frontend and vendor work out of backend slices. |
| Test support exports direct SQL seed helpers. | `testsupport/fixtures.go`; production import scan excludes testsupport. | Test assumptions could become production dependencies. | `intentional/no_action` | Indicators test support | Retain test-only status and boundary coverage. |
| Generated contracts mirror owner inputs. | Generated artifact policy and Testing Harness NLSpec. | Hand edits would drift or be overwritten. | `intentional/no_action` | Generators from authored inputs | Update owner/authored inputs first and run Make generation only when required. |

### IND-DEDUPE-001 — deliberately retained compound query

`sourceRepository.loadByDedupeTx` is an approved fixed owner-specific source/coordination query for S-01. It coordinates the active canonical claim, Indicators subtype state, and Records envelope in one locking statement. S-01 MUST NOT decompose, reorder, generalize, or dynamically parameterize it, and MUST prove through a focused diff or source digest that it is unchanged.

A future replacement is `DEFERRED` until one separately authorized decision closes every row below:

| Required decision | Required closure |
| --- | --- |
| Lock inventory | Enumerate every claim, subtype, envelope row, and identity key that participates. |
| Global lock order | Define one order across advisory identity, claim state, subtype state, and sorted Records envelopes. |
| Lock modes | Define blocking versus `NOWAIT` behavior and the exact public contention mapping. |
| Linearization point | Define when create, reuse, delete, and restore become authoritative. |
| Uniqueness arbitration | Preserve the active-claim constraint/index mechanism and its conflict mapping. |
| Concurrent create | Preserve created/reused response behavior, one stable canonical `record_id`, and at most one active canonical claim. |
| Create/delete and create/restore | Enumerate the legal serial outcomes and preserve concealment and the active-claim bound. |
| Deadlock handling | Name retry ownership, retry bound, and safe terminal result. The default is no new automatic retry. |
| Failure atomicity | Commit no source, history, projection, idempotency success, or Collaboration effect before complete admission. |
| Security | Expose no hidden-resource distinction or raw database failure. |
| Compatibility | Require separate authorization for any migration or public-contract change. |

## 6. Remediation Workstreams

Workstreams execute strictly in the order below. The tracker MUST be updated and validated after each workstream and before the next begins.

| Workflow ID | Name | Required previous workflow | Goal | Principal areas | Exit and tracker checkpoint |
| --- | --- | --- | --- | --- | --- |
| W-00 | Rebaseline and repair the tracker | none | Replace stale planning-only assumptions with current staged state, the 55-export baseline, and the authorized remediation charter. | Tracker only. | Markdown and whitespace checks pass; append W-00 evidence before W-01. |
| W-01 | Specification and boundary adoption | W-00 | Adopt `IND-BOUNDARY-001`, exact lifecycle vocabularies, portable uniqueness, and executable boundary evidence. | Core 00/01/02/04, decision, Domain pointers, base-claim manifest, boundary evidence. | Owner artifacts agree; ID collision, Markdown, boundary, and JSON-shape checks pass; update tracker before S-00. |
| S-00 | Baseline and characterization | W-01 | Retain the 26-row baseline, record the dedupe-query digest, and separate preservation evidence from known conformance repairs. | Tests, harness inputs when new evidence is added, tracker. | Existing owner-conformant behavior is green and RB-004 closes for S-01; update tracker. |
| S-01 | Records ownership and target resolution | S-00 | Use one Records-owned locked envelope snapshot and correct role-specific and storage-failure mappings. | Indicators services/repositories and deterministic tests. | Standalone SQL is absent, role/error matrix is green, dedupe query unchanged; update tracker. |
| S-02A | Vocabulary and portable-contract convergence | S-01 | Establish one immutable runtime vocabulary and exact boundary validation. | Core projection, vocabulary package, HTTP/services/rollback/portability, interval schema, tests/harness. | Exhaustive parity and negative tests plus generation/policy/shape gates pass; update tracker. |
| S-02B | Database lifecycle integrity | S-02A | Enforce unique lifecycle support references and reject invalid legacy state atomically. | Migration `00038`, migration ownership projections, storage/migration tests. | Clean/upgrade/malformed/direct-write/down-up evidence and migration gates pass; update tracker. |
| S-03 | Source-state catalog | S-02B | Derive Recovery and Incident Bundle contributions from one validated immutable 3/1/3 catalog. | Owner-internal catalog, contribution constructors, application assembly, tests. | Exact parity, unchanged bytes/behavior, builds, and owner slices pass; update tracker. |
| S-04 | Facade cohesion and compatibility contraction | S-03 | Split same-package concerns and replace the unused outcome enum with `Created` and `Replayed` booleans. | Indicators root, application consumer, tests/export guard. | Exactly 50 reviewed exports and unchanged public/stored behavior; update tracker. |
| S-05 | Harness and boundary reconciliation | S-04 | Reconcile every authored selector, verification owner, policy, and generated topology affected by prior slices. | Authored harness/boundary inputs and Make-generated outputs. | Every test resolves once and drift/policy checks pass; update tracker. |
| S-06 | Validation and handoff completion | S-05 | Remove obsolete helpers, run narrow-to-broad validation, and complete the handoff. | Entire authorized diff and tracker. | No required failure, unexplained skip, obsolete path, drift, contradiction, or unrecorded blocker remains. |

## 7. Authorized Remediation Slice Plan

All slices preserve owner-conformant behavior. Intentional compatibility corrections are limited to target-specific errors, safe dependency-failure handling, rejection of invalid lifecycle tokens and duplicate support references, fallible internal contribution construction, and removal of the unused internal outcome enum. Public paths, operation IDs, success shapes, authorization order, cursor binding, idempotency, history, projections, Collaboration effects, view fields, valid Incident Bundle bytes, and `loadByDedupeTx` remain unchanged.

### IND-AUTHORIZATION-001 — active implementation charter

The current user task authorizes W-01 and S-00 through S-06, including adopted-owner corrections, authored contract and migration work, direct application-consumer migration, verification routing, Make-generated projections, and tracker maintenance. RB-003 is closed. Work remains bounded as follows:

| Path family | Authorization |
| --- | --- |
| `internal/modules/indicators/**` and Indicators-owned tests | Allowed for every named remediation and deterministic evidence case. |
| Adopted Core owners, the Indicators decision, and `docs/domain.md` | Allowed only for W-01 owner adoption and pointer-only Domain navigation. |
| Authored Incident Bundle and verification contracts | Allowed for the exact lifecycle and evidence projections in S-02A/S-05. |
| `db/migrations/**` and migration ownership inputs/tests | Allowed for the next migration and its exact integrity evidence in S-02B. |
| Application assembly and direct internal consumers | Allowed only for fallible contribution construction and `CreateResult` migration. |
| Generated roots and topology | Refresh only through public Make generators; hand edits are forbidden. |
| `apps/web/**`, `packages/grid-adapter/**`, public routes, Inspector tuples, and view schemas | Excluded; an unexpected diff stops the effort and requires separate `IND-BROWSER-001` authorization. |

Before implementation, this session recorded branch `main`, commit `07b2f2b703fc80dc10fcdba28cb855db660c44d2`, the staged tracker state, 63 target files, and the exact 55-export allowlist. S-00 records the focused digest for `loadByDedupeTx`. The effort stops on an owner contradiction, unexplained baseline failure, required compound-query change, unauthorized frontend need, silent legacy repair, or generated output that cannot be reproduced from authored inputs.

| Slice ID | Depends on | Intended change | Compatibility posture | Required validation | Completion criterion |
| --- | --- | --- | --- | --- | --- |
| S-00 | W-01 | Run and retain focused/service-backed Indicators evidence; add deterministic missing coverage without freezing known nonconformance. | Test-only. | Indicators focused/service-backed selections; deterministic matrix below; dedupe source digest. | The 26-row owner baseline is accounted for, conformant behavior is green, and every known mismatch is assigned to its fixing slice. |
| S-01 | S-00; RB-004 `CLOSED FOR S-01` | Remove `validateIncidentTx` and `validateSourceIncidentTx`; validate semantic roles from the one sorted Records `LoadEnvelopesTx(..., true)` snapshot. | Correct currently misclassified client errors without adding codes or statuses. | Indicators and Records focused/service-backed slices; backend boundary; role/error, outage, concurrency, concealment, and no-side-effect tests. | One lock read; no standalone Records SQL; exact role mapping; safe storage failures; dedupe query byte-for-byte unchanged. |
| S-02A | S-01 | Add one immutable Indicator type/value-kind/observation-status/lifecycle-state vocabulary with context-specific canonicalization; route every runtime boundary through it; tighten portable schema uniqueness. | Valid tokens/bytes unchanged; invalid case, spacing, aliases, unknown lifecycle tokens, and duplicate refs rejected. | Indicators focused/service-backed; exhaustive set equality and defensive copy; generation, artifact policy, and JSON shape. | One runtime owner per family and safe owner-classified failures at every boundary. |
| S-02B | S-02A | Add migration `00038_indicators_lifecycle_integrity.sql` with uniqueness enforcement and in-transaction legacy preflight; Down restores prior validation. | Valid data untouched; invalid legacy state aborts upgrade for explicit operator repair. | Indicators storage and database-migrations owner slices; clean/upgrade/malformed/direct-write/down-up tests; migration drift; migrate build. | Durable uniqueness holds for all writers and invalid upgrade failure is atomic. |
| S-03 | S-02B | Add fallibly constructed immutable `internal/sourcestate` catalog and derive Recovery/Incident Bundle contributions; migrate composition roots atomically. | Internal constructors become fallible; public paths, IDs, order, versions, bytes, and Recovery behavior unchanged. | Indicators, Incident Bundles, Recovery owner slices; exact 3/1/3 parity and negative construction tests; server build and drift gates. | One catalog validates and produces all inventories with defensive copies and deterministic ordering. |
| S-04 | S-03 | Split root declarations by public contracts, store composition, replay codecs, value serialization, and scanners; replace `CreateOutcome` and four constants with `CreateResult.Created`/`Replayed`. | Internal Go breaking change migrated atomically; no alias/deprecation shim; stored/public bytes unchanged. | Exact export guard with production-role reasons; focused/service-backed Indicators; constructor/replay/route/transaction tests; server build and boundary. | Guarded exports contract from 55 to exactly 50 and all live consumers use the smaller result. |
| S-05 | S-04 | Reconcile all authored selectors, verification ownership, source/import policies, and generated topology introduced earlier. | Harness-only; no runtime reads documentation or evidence metadata. | Test-owner resolution, boundary, generation drift, artifact policy, and JSON shape. | Every active test resolves exactly once with no stale or duplicate row; frontend activation remains absent. |
| S-06 | S-05 | Remove obsolete private paths, run affected owner slices and required broad gates in order, append complete evidence, and mark completion. | No additional compatibility change. | Affected focused/service-backed owners; boundary, migration, generation, artifact-policy, shape; builds; `make agent-finalize`; `make test-fast`; `make check`. | No failed required gate, unexplained skip, out-of-scope diff, drift, contradiction, or unrecorded blocker remains. |

### IND-CHAR-001 — S-00 deterministic characterization matrix

Every concurrency case MUST use an explicit barrier, transaction hook, or deterministic harness primitive. Time-based sleeps, probabilistic racing, and retry-until-pass loops are forbidden.

| Case | Owner-conformant outcome | S-00 evidence posture |
| --- | --- | --- |
| Record ID is absent | Role-specific concealed not-found result; no durable side effect. | Known mapping mismatch; deterministic S-01 repair matrix, not frozen. |
| Record has the wrong type for the target role | Same role-specific concealed result as unavailable; no actual type disclosure; no durable side effect. | Known mapping mismatch; deterministic S-01 repair matrix, not frozen. |
| Record is soft-deleted | Same role-specific concealed result; no deletion-state disclosure; no durable side effect. | Known mapping mismatch; deterministic S-01 repair matrix, not frozen. |
| Record belongs to another incident | Same role-specific concealed result; no foreign-incident disclosure; no durable side effect. | Known mapping mismatch; deterministic S-01 repair matrix, not frozen. |
| Record is active, correct type, and same incident | Validation succeeds and the existing operation continues. | Green in production child-route and store integration rows. |
| Delete commits before target resolution | Concealed role-specific failure. | Deterministic committed-delete ordering exists for active identity; target-role assertion belongs to S-01. |
| Delete commits after an initial observation but before final lock/recheck | Mutation does not commit against stale active state; it returns the owner-defined concealed or contention result. | Deterministic Records lock/transaction case assigned to S-01. |
| Restore commits before final lock/recheck | Operation observes restored current state and follows the normal current-state contract. | Deterministic restore ordering exists for active identity; target-role assertion belongs to S-01. |
| Restore is uncommitted or commits after the operation's linearization point | Operation does not speculate on the future restore. | Deterministic Records lock/transaction case assigned to S-01. |
| Two canonical creates race for one dedupe key | Outcomes converge to the created/reused contract, one stable canonical record, and at most one active claim. | Green in `z_concurrent_find_or_create_convergence`; PostgreSQL lock observation provides the barrier. |
| Canonical create follows committed delete | The new create and claim follow the legal delete-then-create order with no partial effect. | Green in `active_identity_claims_follow_records_and_rebuild`. |
| Restore follows replacement create | Conflicting restore fails atomically and the replacement retains the sole active claim. | Green in `active_identity_claims_follow_records_and_rebuild`. |
| Storage fails during target resolution | Safe internal failure; no SQL, relation, constraint, driver value, or hidden identifier leaks. | Known false-semantic mapping risk; deterministic injected S-01 repair case, not frozen. |
| Failure occurs after validation but before commit | No idempotency success, history, projection, Collaboration intent, claim publication, or partial source mutation. | Green in the transaction-atomicity row for create, observation, lifecycle, revision, and projection failures. |

S-00 evidence MUST include an Indicators service-backed characterization row, the existing identity-concurrency row, active-identity delete/restore and rebuild evidence, and transaction/child-mutation atomicity evidence. Exact route error codes MUST be asserted per target role rather than normalized to a new generic error.

Deferred ownership changes are not slices. Moving `httpapi`, providers, test support, frontend behavior, grid integration, or public contracts across owners is forbidden in S-00 through S-06. After `IND-BOUNDARY-001` is adopted, any such move requires a new adopted superseding decision and separate implementation authorization.

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| governance adoption | `make lint-markdown`; `make backend-module-boundary-check`; `make json-shape-check` | `IND-BOUNDARY-001`, `REQ-00-074`, `AC-560`, Base claim manifest, and executable boundary evidence | complete | W-01 used current public targets after a fresh ID collision scan. |
| unit | `make test-slice OWNER=module.indicators` | All active focused Indicators owner rows selected by the harness | yes | Canonical owner selection discovered through `make task-guide`; record result root. |
| integration | `make service-backed-test-slice OWNER=module.indicators` | Eleven current service-backed Indicators rows using PostgreSQL fixtures | yes | Required before transaction, persistence, route, portability, or recovery work. |
| e2e/browser | `make browser-e2e-webserver-backed` | Exact future `IND-BROWSER-001` row within the webserver-backed browser graph | no | Required only when its named activation event occurs. A broad target pass without the exact active owner row is insufficient. |
| generated drift | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check` | Generated code/contracts/topology policy | no | Required after any authored contract, verification, boundary-dependent generator input, or codegen change; generated files must not be hand-edited. |
| import-boundary/static | `make backend-module-boundary-check` | Backend module import policy | yes | Run after every boundary-affecting slice. Add `make frontend-import-boundary-check` and `make frontend-typecheck` only for frontend changes. |
| assembly/build | `make build-server` | Server composition and binary build | no | Required when constructor, contribution, route, or application assembly changes. |
| full check | `make agent-finalize`; `make test-fast`; `make check` | Final maintenance, fast graph, and repository correctness gate | no | Required at final implementation handoff; run `agent-finalize` before broad verification and record `RESULTS_DIR` posture. |
| documentation | `make lint-markdown`; `git diff --check`; `git diff --cached --check`; `git status --short` | Changed documentation and workstream checkpoint evidence | after every tracker update | The tracker began staged and may contain both staged and unstaged changes; both whitespace views are checked. |

Command discovery used `make help`, `make help-all`, `make task-guide ROLE=module-author OWNER=module.indicators`, `make explain-test-owner OWNER=module.indicators`, and `make explain-target` for relevant targets. Those successful discovery commands do not constitute product validation.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| IND-001 | Lock target, label, authority, and sole-write boundary | WF-00 | DONE | none | Section 1; branch/commit/status inspection | Target and exclusions are explicit. |
| IND-002 | Inventory the target files and direct callers | WF-01 | DONE | IND-001 | Section 2; maintained from the 63-file baseline through the current 78-file inventory | Every target file has one row. |
| IND-003 | Diagnose legitimate and misplaced responsibility candidates | WF-04 | DONE | IND-002 | Sections 3 and 5 | Every discovered responsibility has a disposition and owner candidate. |
| IND-004 | Freeze public, storage, revision, projection, portability, security, frontend, and harness contracts | WF-02 | DONE | IND-002 | Section 4 | Every discovered contract risk has owner and test posture. |
| IND-005 | Record focused evidence posture and conditional browser contract | WF-03 | DONE | IND-004 | 19 verifications, 26 rows; `IND-BROWSER-001`; RB-002 | Existing evidence, activation event, owner mapping, row defaults, and exact scenario set are explicit. |
| IND-006 | Establish implementation baseline | S-00 | DONE | IND-005, IND-016, IND-017 | Focused `.cartulary/test-results/20260823T030329Z-p1835662`; service-backed `.cartulary/test-results/20260823T030412Z-p1851702`; `IND-CHAR-001`; dedupe digest | All 26 active rows are accounted for; conformant preservation cases are green and repair cases are assigned before S-01. |
| IND-007 | Replace mutation-path Records validation SQL | S-01 | DONE | IND-006; RB-004 `CLOSED FOR S-01` | One locked snapshot, target-role/security tests, owner slices, boundary and harness evidence | Two standalone queries are absent; owner-specific errors and safe storage handling are correct; `IND-DEDUPE-001` is unchanged. |
| IND-008 | Consolidate Indicators runtime vocabularies and lifecycle integrity | S-02A/S-02B | DONE | IND-007 | Registry/parity/portable evidence plus migration 38, storage, migration-owner, and drift evidence | Runtime validators use one owner registry and the database rejects duplicate support refs. |
| IND-009 | Consolidate source-state and portability inventory | S-03 | DONE | IND-008 | Catalog/parity rows plus Indicators, Incident Bundles, Recovery, build, generation, and harness evidence | One validated 3/1/3 catalog drives inventories. |
| IND-010 | Reorganize root facade by cohesive concern | S-04 | DONE | IND-007, IND-008, IND-009 | Cohesive root files, exact export/role guard, consumer-signal and replay-byte tests, owner slices, build, boundary, and drift evidence | Exact 50-symbol reviewed surface and payload behavior are verified. |
| IND-011 | Reconcile authored harness/boundary inputs for actual moves only | WF-07 | DONE | Activated by new/moved S-01 through S-04 tests and root files | 37 unique active rows; catalog, backend/frontend import boundaries, generated topology, policy, shape, and harness evidence | All activated selectors/imports resolve without hand edits; frontend/browser activation remains false. |
| IND-012 | Run final validation and complete implementation handoff | WF-08 | DONE | IND-007 through IND-010 `DONE`; IND-011 `DONE` or explicitly inactive | S-06 result roots, final authorized-diff ledger, protected-invariant audit, and session log | Required checks, skips, failures, files, and next action are recorded. |
| IND-013 | Move backend HTTP/providers/testsupport across module boundaries | WF-05 | DROPPED | adopted `IND-BOUNDARY-001` | Adopted owner-local topology | No move is part of this remediation; any later move requires a superseding adopted decision. |
| IND-014 | Change frontend Indicator workflow or grid integration | frontend boundary | DEFERRED | `IND-BROWSER-001` activation and later authorization | Conditional row contract in Section 4 | An authorized frontend task supplies the exact row/file IDs and all five scenarios before merge. |
| IND-015 | Perform the authorized lifecycle contract and migration corrections | W-01/S-02A/S-02B | DONE | IND-017 | Owner documents, runtime, portable schema, migration 38, generated ownership, and compatibility evidence | Exact vocabularies and uniqueness agree from owner through database. |
| IND-016 | Formally adopt `IND-BOUNDARY-001` | W-01 | DONE | W-00 | Decision, `REQ-00-074`, `AC-560`, Base claim, export roles, import-topology evidence | RB-001 is `CLOSED`; IDs were collision-checked; owner and evidence artifacts agree. |
| IND-017 | Issue `IND-AUTHORIZATION-001` implementation charter | W-00 | DONE | user authorization | Current implementation task and Section 7 path envelope | RB-003 is `CLOSED`; W-01 is the next permitted workstream. |
| IND-018 | Rebaseline controlling tracker | W-00 | DONE | current repository evidence | Branch/commit/staged state, 63-file inventory, 55-export baseline, repaired charter | Tracker reflects current facts and owner-conformance corrections before W-01. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-22T20:51:17-04:00 | Codex planning session | Tracker created from a clean `main`/`07b2f2b7` baseline; no prior tracker history existed. | Inspected framework, Core/adopted owners, Domain, all 63 target files, direct callers and contracts; touched only this tracker. | `sed`, `rg`, `find`, `jq`, Git status/branch/revision, Make help/explain commands. | Scope, authority, inventory, and later-authorization boundary recorded. | RB-001 and RB-003 apply only to later implementation. | Obtain explicit authorization and start S-00, not production edits in this session. |
| 2026-08-22T22:08:46-04:00 | Codex NLSpec revision session | Existing tracker revised in place; original baseline and handoff row preserved; only the tracker remains untracked. | Inspected the tracker, analysis notes, NLSpec guidance, Core/decision registration examples, Records capability/order, view/Inspector/OpenAPI contracts, and harness topology; touched only this tracker. | `wc`, `sed`, `rg`, `find`, `jq`, Git status/branch/revision, `make lint-markdown`, untracked-file `git diff --no-index --check`. | Normative posture, stable decisions, defaults, mappings, gates, traceability, and binary acceptance were added. Markdown lint passed at `.cartulary/test-results/20260823T021043Z-p1794231`; the no-index check returned the expected difference status `1` with no whitespace diagnostic. | RB-001 and RB-003 are `BLOCKED`; no owner adoption or implementation occurred. | Formally adopt `IND-BOUNDARY-001`, then issue `IND-AUTHORIZATION-001`; S-00 remains prohibited before both. |
| 2026-08-22 implementation session, W-00 | Rebaseline and tracker repair complete. | Rechecked and changed only this tracker. | Git branch/revision/status, `rg` file/export/ID scans, tracker review and patch; `make lint-markdown`; `git diff --check`; `git diff --cached --check`. | Confirmed `main`/`07b2f2b703fc80dc10fcdba28cb855db660c44d2`, staged tracker, 63 target files, 55 exports, and free `REQ-00-074`/`AC-560`; replaced the stale planning charter. Markdown and whitespace checks passed; result root `.cartulary/test-results/20260823T025049Z-p1809750`. | RB-003 is closed by the user authorization; RB-001 is gated on W-01 adoption. | Begin W-01 specification and boundary adoption. |
| 2026-08-22 implementation session, W-01 | Specification and boundary adoption complete. | Added `docs/decisions/indicators-module-boundary.md`; changed Core 00/01/02/04, `docs/domain.md`, Indicators boundary/export guards, and this tracker. | ID/owner review, `make format`, focused export row, `make lint-markdown`, `make backend-module-boundary-check`, `make json-shape-check`, staged/unstaged/untracked whitespace checks. | Adopted `REQ-00-074`/`AC-560`; made portable lifecycle/uniqueness and the four Core 02 token families exact; added pointer-only Domain rows and executable export-role/import-topology evidence. Final format passed at `.cartulary/test-results/20260823T030030Z-p1827431`; focused row at `.cartulary/test-results/20260823T030033Z-p1831154`; final tracker Markdown at `.cartulary/test-results/20260823T030229Z-p1834440`; boundary at `.cartulary/test-results/20260823T030056Z-p1832429`; JSON shape at `.cartulary/test-results/20260823T030056Z-p1832255`. | RB-001 closed. Preliminary focused runs `.cartulary/test-results/20260823T025800Z-p1818954`, `.cartulary/test-results/20260823T025854Z-p1824934`, and `.cartulary/test-results/20260823T025926Z-p1826034` exposed incomplete legacy import inventory; corrected without product changes. No skip or owner contradiction. | Begin S-00 baseline and characterization; do not begin S-01 until RB-004 closes. |
| 2026-08-22 implementation session, S-00 | Baseline and characterization complete. | Inspected current Indicators tests, harness owner/family data, run artifacts, and `source_repository.go`; changed only this tracker in S-00. | Focused digest; `make test-slice OWNER=module.indicators`; `make service-backed-test-slice OWNER=module.indicators`; `make explain-test-owner OWNER=module.indicators`; run-summary review. | `loadByDedupeTx` lines 20–61 digest `sha256:3343ded88d9e54fe69ba516dbfe1df22c35be956a626d1682fa52ec823a85af4`. Focused passed with all 26 row evidence records at `.cartulary/test-results/20260823T030329Z-p1835662`; service-backed passed all 11 catalog rows at `.cartulary/test-results/20260823T030412Z-p1851702` (seven grouped execution units). Catalog remains 24 Go/two Vitest and 15 unit/11 integration. | RB-004 closed for S-01. No failure or skip. Current target-role and safe-storage misclassifications are explicitly S-01 repair cases and were not frozen. Existing deterministic evidence covers valid routes, failure atomicity, dedupe convergence, committed delete/create, conflicting restore, and rebuild/recovery order. | Begin S-01 and make every assigned repair case green before any S-02A edit. |
| 2026-08-22 implementation session, S-01 | Records ownership and target resolution complete. | Changed Indicators store/coordination, observation/lifecycle services and repository, removed the source helper, added target-role/storage/HTTP/atomicity tests, updated Indicators verification/family and PostgreSQL fixture policy, refreshed `tools/execution_topology_render_index.json` through Make, updated harness count evidence, and changed this tracker. No Records implementation changed. | `make format`; three focused target-resolution rows; full Indicators/Records focused and service-backed slices; backend boundary; source/SQL scans; `make test-catalog-check`; `make generate`; JSON shape; generation drift; harness contract; whitespace checks. | One sorted `LoadEnvelopesTx(..., true)` snapshot now precedes role validation. Source, requested target, addressed Indicator, prior dependency, support-ref, row-version, and storage failures have exact mappings and no effects. Target rows passed at `.cartulary/test-results/20260823T031710Z-p1877507`; Indicators/Records focused at `.cartulary/test-results/20260823T031801Z-p1894172` and `.cartulary/test-results/20260823T031801Z-p1894209`; boundary at `.cartulary/test-results/20260823T031801Z-p1894270`; service-backed at `.cartulary/test-results/20260823T031847Z-p1926561` and `.cartulary/test-results/20260823T031847Z-p1926555`; generation at `.cartulary/test-results/20260823T032003Z-p1959282`; JSON shape/drift at `.cartulary/test-results/20260823T032021Z-p1962413` and `.cartulary/test-results/20260823T032021Z-p1962415`; harness at `.cartulary/test-results/20260823T032048Z-p1966374`. Catalog is now 29 rows: 27 Go/two Vitest, 17 unit/12 integration, 12 service-backed. | No blocker or skip. Initial format rejected unsorted authored rows and missing transaction approval; corrected. Pre-generation JSON/harness runs `.cartulary/test-results/20260823T031945Z-p1957685` and `.cartulary/test-results/20260823T031945Z-p1957921` correctly detected stale topology; post-generation harness `.cartulary/test-results/20260823T032021Z-p1962646` exposed the expected count mirror, then passed after its authored update. | Begin S-02A. Keep the dedupe digest `sha256:3343ded88d9e54fe69ba516dbfe1df22c35be956a626d1682fa52ec823a85af4` exact. |
| 2026-08-22 implementation session, S-02A | Vocabulary and portable-contract convergence complete. | Added owner-internal `vocabulary.go`/test; changed identity, lifecycle service, HTTP decoding, rollback row/child providers, and Incident Bundle preparation; added root contract/portable, HTTP, and rollback tests; tightened the interval row schema; updated Indicators verification/family inputs and Make-generated topology; refreshed the complete 70-file inventory and this tracker. | `make format`; `make generate`; five-row vocabulary selection; complete Indicators and Incident Bundles focused/service-backed slices; token-ownership and dedupe digest scans; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make test-catalog-check`; `make harness-contract`; tracker Markdown/whitespace checks. | One defensive-copy registry now owns nine Indicator types, three value kinds, three observation statuses, and four lifecycle states. Creation retains tolerant type/kind canonicalization; HTTP, services, rollback, and portable data require exact membership. Portable and rollback support refs reject duplicates. Interval schema has the exact lifecycle enum and `uniqueItems`. Final vocabulary rows passed at `.cartulary/test-results/20260823T033132Z-p1984208`; Indicators focused/service-backed at `.cartulary/test-results/20260823T033200Z-p1984932` and `.cartulary/test-results/20260823T033348Z-p2034204`; Incident Bundles focused/service-backed at `.cartulary/test-results/20260823T033247Z-p2018312` and `.cartulary/test-results/20260823T033428Z-p2049935`; generation at `.cartulary/test-results/20260823T032958Z-p1975076`; drift/policy/shape at `.cartulary/test-results/20260823T033529Z-p2065844`, `.cartulary/test-results/20260823T033529Z-p2065870`, and `.cartulary/test-results/20260823T033529Z-p2065893`; harness at `.cartulary/test-results/20260823T033529Z-p2066275`. Catalog is now 33 rows: 31 Go/two Vitest, 21 unit/12 integration, 12 service-backed; verification count is 21. `loadByDedupeTx` remains `sha256:3343ded88d9e54fe69ba516dbfe1df22c35be956a626d1682fa52ec823a85af4`. | No blocker or skip. The first vocabulary selection `.cartulary/test-results/20260823T033018Z-p1978083` and isolated row `.cartulary/test-results/20260823T033056Z-p1979435` exposed rollback trimming before exact membership; corrected, with isolated pass `.cartulary/test-results/20260823T033125Z-p1983777`. Parallel Incident Bundles run `.cartulary/test-results/20260823T033200Z-p1984934` failed harness coordination; the isolated focused rerun passed. | Begin S-02B. Add the next collision-checked migration; do not deduplicate invalid legacy rows. |
| 2026-08-22 implementation session, S-02B | Database lifecycle integrity complete. | Added `00038_indicators_lifecycle_integrity.sql` and its Indicators migration test; assigned migration 38 to Indicators in generator and shape-check ownership inputs; regenerated migration history, schema-object ownership, and execution topology through Make; updated migration schema-hash mirrors, Indicators test routing, harness migration-fixture count, the 71-file inventory, and this tracker. | `make format`; `make generate`; focused lifecycle/storage rows; full database-migrations focused/service-backed slices; `make migration-drift`; `make build-migrate`; generation/artifact/shape/catalog/harness gates; dedupe digest and whitespace scans. | Migration 38 preflights every historical interval in its transaction, replaces the helper without permissive replacement DDL, recreates the validated same-name CHECK, preserves runtime/recovery EXECUTE grants, and restores the prior helper in Down without rewriting data. Clean/valid/malformed/direct insert/direct update/Down-Up evidence passed at `.cartulary/test-results/20260823T034550Z-p2090643` and final storage pair `.cartulary/test-results/20260823T034842Z-p2125483`. Database-migrations focused/service-backed passed at `.cartulary/test-results/20260823T035146Z-p2162629` and `.cartulary/test-results/20260823T035252Z-p2178602`; migration drift at `.cartulary/test-results/20260823T035403Z-p2194358`; migrate build at `.cartulary/test-results/20260823T035415Z-p2197472`; final JSON shape/harness at `.cartulary/test-results/20260823T035603Z-p2208719` and `.cartulary/test-results/20260823T035639Z-p2209879`; final generation drift/artifact policy at `.cartulary/test-results/20260823T035710Z-p2210468` and `.cartulary/test-results/20260823T035721Z-p2213404`. Catalog is now 34 rows: 32 Go/two Vitest, 21 unit/13 integration, 13 service-backed. `loadByDedupeTx` remains exact. | No blocker or skip. Generation failures `.cartulary/test-results/20260823T034306Z-p2084054`, `.cartulary/test-results/20260823T034457Z-p2086078`, and `.cartulary/test-results/20260823T034509Z-p2087103` exposed missing owner allocation, forbidden replacement DDL, then sqlc rename incompatibility; corrected structurally. Storage run `.cartulary/test-results/20260823T034656Z-p2106747` exposed lost helper grants. Database-owner run `.cartulary/test-results/20260823T034932Z-p2141421` exposed two expected schema-hash mirrors. JSON/harness roots `.cartulary/test-results/20260823T035424Z-p2199692`, `.cartulary/test-results/20260823T035424Z-p2200008`, `.cartulary/test-results/20260823T035445Z-p2204514`, `.cartulary/test-results/20260823T035538Z-p2205310`, and `.cartulary/test-results/20260823T035622Z-p2209305` exposed owner-shape, stale topology, and fixture-count projections before final passes. | Begin S-03. Keep the exact 3/1/3 inventory and portable bytes unchanged. |

| 2026-08-23 implementation session, S-03 | Source-state catalog complete. | Added owner-internal `sourcestate/catalog.go` and its exhaustive validation/immutability test plus root projection-parity evidence; made the Incident Bundle and Recovery constructors fallible; projected both inventories from the catalog; migrated both application composition roots and all direct tests; removed path facts from the provider; added one verification and two authored rows; regenerated topology through Make; refreshed the 74-file inventory and this tracker. | `make format`; `make generate`; new source-state rows; complete Indicators, Incident Bundles, and Recovery focused/service-backed slices; `make build-server`; generation drift, artifact policy, JSON shape, catalog, Markdown, and harness gates; exact order and dedupe digest scans. | The catalog owns exactly three authoritative relations, one rebuildable relation, and three ordered portability descriptors; it rejects empty, duplicate, unsafe, mismatched, wrongly versioned/ordered, or identity-less facts and returns defensive copies. Authored Incident Bundle parity, Recovery class parity, unchanged path order, and deterministic descriptor bytes are executable. Final format passed at `.cartulary/test-results/20260823T041415Z-p2416440`; Indicators/Incident Bundles/Recovery focused passed at `.cartulary/test-results/20260823T041424Z-p2420348`, `.cartulary/test-results/20260823T041424Z-p2420353`, and `.cartulary/test-results/20260823T041537Z-p2502671`; their service-backed selections passed at `.cartulary/test-results/20260823T041658Z-p2554186`, `.cartulary/test-results/20260823T041658Z-p2554196`, and `.cartulary/test-results/20260823T041757Z-p2585492`. Generation passed at `.cartulary/test-results/20260823T041121Z-p2391661`; final server build/generation drift passed at `.cartulary/test-results/20260823T042036Z-p2637727` and `.cartulary/test-results/20260823T042036Z-p2637222`; artifact policy/shape/harness passed at `.cartulary/test-results/20260823T041152Z-p2406811`, `.cartulary/test-results/20260823T041152Z-p2406826`, and `.cartulary/test-results/20260823T041153Z-p2407121`; catalog check exited zero and tracker Markdown passed at `.cartulary/test-results/20260823T042036Z-p2637523`. Catalog is now 36 rows: 34 Go/two Vitest, 23 unit/13 integration, 13 service-backed; verification count is 22. `loadByDedupeTx` remains exact. | No blocker or skip. Parallel Recovery focused root `.cartulary/test-results/20260823T041424Z-p2420362` failed because its browser support summary could not see a concurrently owned group result; the isolated rerun passed and no product test failed. | Begin S-04. Preserve constructor paths and portable/stored bytes while contracting only the unused outcome surface. |
| 2026-08-23 implementation session, S-04 | Facade cohesion and compatibility contraction complete. | Removed mixed `api.go`/`store.go`; added cohesive `contracts.go`, `store_composition.go`, `replay_codec.go`, `value_serialization.go`, and `persistence_scanners.go`; removed `CreateOutcome` and four constants without aliases; added `CreateResult.Created`/`Replayed`; migrated Workbook and all tests; added exact stored-byte and consumer-signal tests; updated boundary evidence, the authored family row, generated topology, the 78-file inventory, and this tracker. | `make format`; facade/surface, route, replay, transaction, complete focused/service-backed Indicators selections; `make build-server`; `make backend-module-boundary-check`; `make generate`; catalog, generation drift, artifact policy, JSON shape, and harness checks; stale-symbol/file and dedupe scans. | The guarded root surface is exactly 50 declarations and every retained symbol has a production-role reason. Creation maps to 201 only from `Created`; reuse/update map to 200 without a public distinction; replay maps to 200 plus `Replayed`; record/change-set/version identity and exact stored JSON bytes remain stable. Final format/focused/service-backed roots are `.cartulary/test-results/20260823T043207Z-p2721356`, `.cartulary/test-results/20260823T043407Z-p2745384`, and `.cartulary/test-results/20260823T043051Z-p2692593`; the focused facade/route/transaction selection passed at `.cartulary/test-results/20260823T043324Z-p2729639`; generation passed at `.cartulary/test-results/20260823T043244Z-p2726486`; final server/boundary passed at `.cartulary/test-results/20260823T043516Z-p2766401` and `.cartulary/test-results/20260823T043516Z-p2766239`; drift/policy/shape/harness passed at `.cartulary/test-results/20260823T043457Z-p2761686`, `.cartulary/test-results/20260823T043457Z-p2761707`, `.cartulary/test-results/20260823T043457Z-p2761706`, and `.cartulary/test-results/20260823T043457Z-p2762002`; catalog check exited zero. Catalog is now 37 rows: 35 Go/two Vitest, 24 unit/13 integration, 13 service-backed; verification count remains 22. `loadByDedupeTx` remains exact. | No blocker or skip. Initial surface root `.cartulary/test-results/20260823T042807Z-p2665046` exposed the stale allowed filename after the structural split; corrected. Initial boundary root `.cartulary/test-results/20260823T043051Z-p2692727` exposed a test-only owner-port import in the S-03 parity test; replaced with a local contract projection and the final boundary passed. | Begin S-05 reconciliation; frontend/browser activation remains false. |
| 2026-08-23 implementation session, S-05 | Harness and boundary reconciliation complete. | Reviewed the authored Indicators verification/family inputs, Make target inventory, generated topology, current diff boundaries, and active row identities; changed only this tracker in the reconciliation slice because S-01 through S-04 already updated the required authored/generated inputs. | `make explain-test-owner OWNER=module.indicators`; `make test-catalog-check`; `make backend-module-boundary-check`; `make frontend-import-boundary-check`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make harness-contract`; duplicate/inactive-row and frontend/generated-protocol diff scans. | All 37 rows are active and uniquely keyed: 35 Go/two Vitest, 24 unit/13 integration, 13 service-backed. Catalog resolution exited zero; backend/frontend boundaries passed at `.cartulary/test-results/20260823T043853Z-p2782329` and `.cartulary/test-results/20260823T043853Z-p2782296`; generation/shape passed at `.cartulary/test-results/20260823T043853Z-p2782053` and `.cartulary/test-results/20260823T043853Z-p2782093`; artifact policy/harness passed at `.cartulary/test-results/20260823T043907Z-p2786200` and `.cartulary/test-results/20260823T043907Z-p2786408`. No stale selector, duplicate row, unapproved import, generated drift, frontend diff, browser activation, or generated protocol diff is present. | No blocker, skip, or inactive required row. `IND-BROWSER-001` remains resolved-conditional and no browser evidence claim is made. | Begin S-06 in the prescribed validation order and complete the final handoff. |
| 2026-08-23 implementation session, S-06 | Validation and handoff completion complete. | Audited and removed obsolete paths in their fixing slices; ran every affected focused/service-backed owner selection; reconciled the migration-evidence v2 golden after migration 38 changed its exact manifest bytes; audited the complete 69-path authorized diff, the 78-file Indicators inventory, excluded frontend/generated protocol paths, the 50-export guard, and the protected dedupe query; completed this tracker. | Indicators, Records, Incident Bundles, Database Migrations, Recovery, and Workbook focused/service-backed slices; boundary, migration drift, generation drift, artifact policy, JSON shape, server/migrate builds; `make agent-finalize`; `make test-fast`; `make check`; Operator migration-evidence row; source/scope/stale-symbol scans; tracker Markdown and whitespace checks. | Focused roots: Indicators `.cartulary/test-results/20260823T044054Z-p2789020`, Records `.cartulary/test-results/20260823T044054Z-p2789021`, Incident Bundles `.cartulary/test-results/20260823T044145Z-p2820926`, Database Migrations `.cartulary/test-results/20260823T044145Z-p2820925`, Recovery `.cartulary/test-results/20260823T044258Z-p2852949`, Workbook `.cartulary/test-results/20260823T044413Z-p2905369`. Service-backed roots: Indicators `.cartulary/test-results/20260823T044636Z-p2965550`, Records `.cartulary/test-results/20260823T044636Z-p2965552`, Incident Bundles `.cartulary/test-results/20260823T044726Z-p2996933`, Database Migrations `.cartulary/test-results/20260823T044726Z-p2996935`, Recovery `.cartulary/test-results/20260823T044840Z-p3028218`, Workbook `.cartulary/test-results/20260823T044958Z-p3079670`. Boundary/migration/generation/policy/shape passed at `.cartulary/test-results/20260823T045217Z-p3136696`, `.cartulary/test-results/20260823T045223Z-p3137087`, `.cartulary/test-results/20260823T045235Z-p3139920`, `.cartulary/test-results/20260823T045246Z-p3142874`, and `.cartulary/test-results/20260823T045252Z-p3143335`; server/migrate builds passed at `.cartulary/test-results/20260823T045259Z-p3143904` and `.cartulary/test-results/20260823T045312Z-p3155708`. Final `agent-finalize`, `test-fast`, and `check` passed at `.cartulary/test-results/20260823T045726Z-p3202053`, `.cartulary/test-results/20260823T045742Z-p3204981` (425/425), and `.cartulary/test-results/20260823T045758Z-p3205991` (646/646). Final tracker Markdown passed at `.cartulary/test-results/20260823T050715Z-p3332114`. The guarded facade is exactly 50 exports, obsolete helper/outcome scans are empty for Indicators, no excluded path changed, and `loadByDedupeTx` remains `sha256:3343ded88d9e54fe69ba516dbfe1df22c35be956a626d1682fa52ec823a85af4`. | No unresolved blocker or unexplained skip. `RESULTS_DIR` was unset because no qualifying retained successful full warm-check root existed, so retained-run maintenance was explicitly skipped; `agent-finalize` itself passed. Initial `test-fast` root `.cartulary/test-results/20260823T045338Z-p3160413` failed one Operator golden because migration 38 correctly changed migration-manifest evidence bytes; the digest was advanced to `8b4acb6c9b6ebe6fe9f18c3c200ebf3c5a37698cb697087164ee4b0ae044afdf`, its authored row passed at `.cartulary/test-results/20260823T045717Z-p3201606`, and finalization plus both broad gates then passed. Browser evidence remained conditionally inactive. | Handoff complete; no remaining in-scope action. `IND-DEDUPE-001`, frontend/browser work, public route changes, and cross-owner relocation remain separately authorized future work only. |

### S-06 final authorized-diff ledger

The completed worktree contains 69 changed paths. This ledger records every path and its disposition; generated topology and migration projections were produced through public Make targets.

Modified owner, contract, documentation, and migration-projection paths:

- `contracts/incident-bundles/indicator_state_intervals.row.v1.schema.json`
- `contracts/verification/owners/module.indicators.json`
- `docs/domain.md`
- `docs/handoffs/indicators-module-refactor-tracker.md`
- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `internal/app/incidentportabilityassembly/catalog.go`
- `internal/app/operator/operator_migration_evidence_test.go`
- `internal/app/recoveryassembly/state_catalog.go`
- `internal/app/workbookassembly/indicator_adapter.go`
- `internal/modules/database_migrations/catalog_characterization_test.go`
- `internal/testutil/pgtest/pgtest_test.go`
- `tools/database-migrations/generate-catalog-projections.mjs`
- `tools/execution_topology_render_index.json`
- `tools/harness/generated-artifacts/database-contract-drift/schema-object-ownership.mjs`
- `tools/harness/tests/contract-suite-support.mjs`
- `tools/migration_history_manifest.json`
- `tools/postgres_fixture_policy_registry.json`
- `tools/schema_object_ownership_manifest.json`
- `tools/test_families/module.indicators.json`

Modified Indicators implementation and test paths:

- `internal/modules/indicators/boundary_guard_test.go`
- `internal/modules/indicators/child_coordination.go`
- `internal/modules/indicators/create_service.go`
- `internal/modules/indicators/exported_surface_test.go`
- `internal/modules/indicators/httpapi/decoding.go`
- `internal/modules/indicators/identity_portability_test.go`
- `internal/modules/indicators/incident_bundle_contribution.go`
- `internal/modules/indicators/indicators_test.go`
- `internal/modules/indicators/internal/identity/identity.go`
- `internal/modules/indicators/internal/providers/incidentbundle/portable_prepare.go`
- `internal/modules/indicators/internal/providers/incidentbundle/source_port.go`
- `internal/modules/indicators/internal/providers/rollback/children.go`
- `internal/modules/indicators/internal/providers/rollback/provider.go`
- `internal/modules/indicators/lifecycle_service.go`
- `internal/modules/indicators/observation_origin_test.go`
- `internal/modules/indicators/observation_repository.go`
- `internal/modules/indicators/observation_service.go`
- `internal/modules/indicators/portability_characterization_test.go`
- `internal/modules/indicators/portability_final_state_test.go`
- `internal/modules/indicators/portability_invariants_test.go`
- `internal/modules/indicators/production_contract_test.go`
- `internal/modules/indicators/provider_encapsulation_test.go`
- `internal/modules/indicators/recovery_state.go`
- `internal/modules/indicators/source_repository.go`
- `internal/modules/indicators/unit_test.go`

Added paths:

- `db/migrations/00038_indicators_lifecycle_integrity.sql`
- `docs/decisions/indicators-module-boundary.md`
- `internal/app/workbookassembly/indicator_adapter_test.go`
- `internal/modules/indicators/contracts.go`
- `internal/modules/indicators/httpapi/error_mapping_test.go`
- `internal/modules/indicators/httpapi/vocabulary_admission_test.go`
- `internal/modules/indicators/internal/providers/rollback/vocabulary_test.go`
- `internal/modules/indicators/internal/sourcestate/catalog.go`
- `internal/modules/indicators/internal/sourcestate/catalog_test.go`
- `internal/modules/indicators/internal/vocabulary/vocabulary.go`
- `internal/modules/indicators/internal/vocabulary/vocabulary_test.go`
- `internal/modules/indicators/lifecycle_integrity_migration_test.go`
- `internal/modules/indicators/persistence_scanners.go`
- `internal/modules/indicators/replay_codec.go`
- `internal/modules/indicators/replay_codec_test.go`
- `internal/modules/indicators/source_state_contract_test.go`
- `internal/modules/indicators/store_composition.go`
- `internal/modules/indicators/target_resolution_integration_test.go`
- `internal/modules/indicators/value_serialization.go`
- `internal/modules/indicators/vocabulary_contract_test.go`

Removed after all callers migrated:

- `internal/modules/indicators/api.go`
- `internal/modules/indicators/store.go`

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-22T20:51:17-04:00 | Codex planning session | Live target diagnosed as a valid Indicators source-owner module with mixed internal seams, not an evidenced catch-all. | All production files in Section 2; application and Network Flow callers; touched only this tracker. | Target file enumeration; source/import/symbol searches; exact file reads. | Records validation SQL marked `must_fix`; vocabulary, source catalog, and facade cohesion marked `should_fix`; cross-module moves deferred. | RB-001; RB-004 before S-01. | Characterize Records outcomes and transaction ordering in S-00/S-01. |
| 2026-08-22T22:08:46-04:00 | Codex NLSpec revision session | `IND-BOUNDARY-001` now specifies the complete proposed owner-local topology without claiming adoption. | Tracker; `records/route_target.go`, Records store/test evidence; Indicators source/observation repositories and observation service; touched only this tracker. | Exact symbol/order searches and source reads. | `IND-RECORDS-001` records the live four-field caller-transaction interface; PF-002 corrects the notes to the live post-lock call order; `IND-DEDUPE-001` keeps the compound query unchanged. | RB-001 is `BLOCKED`; RB-004 is `GATED`. | Adopt the boundary decision, authorize implementation, then produce deterministic S-00 evidence. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-22T20:51:17-04:00 | Codex planning session | Frontend Indicator workflow remains in `apps/web`; grid vendor remains isolated in `packages/grid-adapter`; frontend changes are out of scope. | Indicator workflow/handler/port source and tests; grid-vendor import search; touched only this tracker. | `rg` and exact source reads. | Four Inspector tuples and two Vitest rows mapped; no direct target grid-vendor coupling found. | RB-002 blocks a direct Indicator browser-evidence claim for future frontend work. | Discover exact browser owner row only if a later task includes frontend changes. |
| 2026-08-22T22:08:46-04:00 | Codex NLSpec revision session | Frontend and grid work remain excluded; their future evidence contract is now exact and conditional. | Tracker, Inspector registry, Workbook browser-family examples, execution topology; touched only this tracker. | `jq`, `rg`, and exact manifest reads. | `IND-BROWSER-001` defines owner/collaborators, runner/project/stage, runtime/resource/fixture defaults, skip posture, and five exact semantic scenarios. | RB-002 is `RESOLVED — CONDITIONAL`; activation requires a named frontend/contract/routing/grid diff plus authorization. | Do nothing for backend-only slices; create the exact row before merge only if the activation event occurs. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-22T20:51:17-04:00 | Codex planning session | Authored view, inspector, import, portability, and OpenAPI inputs mapped; generated files remain downstream and untouched. | Contracts listed in Section 1; projection contribution; touched only this tracker. | `jq`, `rg`, `make explain-target` for drift/policy targets. | Thirteen fields, eight operations/six paths, three bundle paths, and generated-risk posture recorded. | None for planning. | Use authored inputs first and Make generation only if a later slice changes contracts. |
| 2026-08-22T22:08:46-04:00 | Codex NLSpec revision session | Existing contract ownership remains frozen; proposed implementation authority is explicitly separate. | Tracker, view schema, Inspector registry, Indicators OpenAPI owner source, Core registration examples; touched only this tracker. | `jq`, `rg`, exact contract reads. | Added exact eight-operation status/default map, ordered 13-field create/immutability map, four Inspector tuples, and provisional `REQ-00-074`/`AC-560` collision rule. No authored or generated contract changed. | RB-001 blocks adoption; provisional IDs are not authority. | Future governance task collision-checks IDs and adopts through owner artifacts; later code uses authored inputs and Make generation only. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-22T20:51:17-04:00 | Codex planning session | Existing routing contains 19 verification contracts and 26 active rows: 24 Go, two Vitest, 15 unit, 11 integration/service-backed. | All target tests, verification owner, family manifest, catalog owner; touched only this tracker. | `make task-guide`, `make explain-test-owner`, `make help`, `make help-all`, relevant `make explain-target`; `make lint-markdown`; scoped `git diff --check`; no product test target. | Canonical implementation commands discovered. Markdown lint passed at `.cartulary/test-results/20260823T010140Z-p1774773`; the scoped diff check passed. No product validation pass is claimed. | RB-002 for frontend browser ownership. | Run S-00 only in a later authorized implementation task. |
| 2026-08-22T22:08:46-04:00 | Codex NLSpec revision session | The 19-verification/26-row evidence baseline is preserved; no product row ran or changed. | Tracker, owner verification/family manifests, browser topology/examples; touched only this tracker. | Repository discovery; `make lint-markdown`; untracked-file `git diff --no-index --check`; no product test. | `IND-CHAR-001` adds 14 deterministic S-00 outcomes and forbids sleeps; `IND-BROWSER-001` defines conditional browser accounting. Markdown lint passed at `.cartulary/test-results/20260823T021043Z-p1794231`; no-index check emitted no whitespace diagnostic and exited `1` because the file differs from `/dev/null`. | RB-001 and RB-003 prohibit S-00; RB-004 remains `GATED`. | After adoption and authorization, run the exact owner slices and record every result root. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-22T20:51:17-04:00 | Codex planning session | Session, CSRF, membership, role, concealment, replay, row-version, and atomic side-effect precedence is frozen. | Core 04 REQ-04-150, HTTP routes/decoders, child route and transaction tests; touched only this tracker. | `rg` and exact owner/source reads. | Security behavior assigned high refactor risk with existing service-backed evidence. | RB-004 requires outcome/locking characterization before Records-port replacement. | Preserve route admission order and failure-side-effect absence in S-01/S-04. |
| 2026-08-22T22:08:46-04:00 | Codex NLSpec revision session | Concealment and failure atomicity now have exact target-resolution mappings and pre/post evidence gates. | Tracker, Core 04 REQ-04-150/AC-511/AC-532, Records resolver, Indicators route/error/lock paths; touched only this tracker. | `rg` and exact source/owner reads. | `IND-ERROR-001` freezes route-specific source/Indicator outcomes, storage-failure redaction, admission precedence, and absence of durable effects; S-01 cannot move the post-lock reads. | RB-004 is `GATED` until all deterministic baseline cases pass. | Produce `IND-CHAR-001`, close RB-004 for S-01, and preserve the compound lock query. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-22T20:51:17-04:00 | Codex planning session | Planning is complete; all implementation rows remain inactive. | This tracker and blocker evidence only; touched only this tracker. | Scoped repository discovery only. | No owner contradiction and no repository/framework claim substituted for live state. | RB-001 through RB-004 as scoped in Section 11. | Later authorized session runs S-00, records result roots here, then decides whether S-01 can start. |
| 2026-08-22T22:08:46-04:00 | Codex NLSpec revision session | Tracker content is decision-complete; implementation readiness remains false. | Tracker and exact blocker evidence; touched only this tracker. | Scoped discovery plus documentation-only validation. | RB-001/RB-003 are `BLOCKED`, RB-002 is `RESOLVED — CONDITIONAL`, RB-004 is `GATED`, and RB-005 is `DEFERRED`; no owner contradiction was found. | Formal adoption, authorization, and S-00 evidence remain external prerequisites. | Execute `IND-BOUNDARY-001` adoption → `IND-AUTHORIZATION-001` → S-00 → close RB-004 for S-01; do not activate RB-002 or RB-005 opportunistically. |

## 11. Open Questions and Blockers

The blocker states are closed and have these meanings:

| State | Meaning | Default transition rule |
| --- | --- | --- |
| `BLOCKED` | Required external authority is absent. Named downstream work MUST NOT begin. | Transition only when the authority artifact exists and is recorded. |
| `GATED` | The resolution is specified, but required executable evidence is absent. | Transition only when every named evidence condition passes. |
| `RESOLVED — CONDITIONAL` | The issue does not apply to the active scope; a named event reactivates its requirement. | Remain resolved unless that event occurs. |
| `CLOSED` | Every stated closure criterion has current evidence. | Reopen only if the evidence becomes stale or contradictory. |
| `DEFERRED` | The decision is deliberately excluded from the active sequence. | It MUST NOT block the active sequence and MUST NOT be implemented without separate authorization. |

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | `IND-BOUNDARY-001` adoption gate. | A tracker cannot own permanent package placement or permit cross-module redistribution. | Decision, Core 00 registration, Core 04 acceptance and Base manifest, executable boundary evidence, ID collision check, and owner review. | `CLOSED`: W-01 adopted and validated every named artifact. |
| RB-002 | No active `module.indicators` Playwright row exists. | Backend-only work needs no direct browser row, but a frontend or Inspector diff requires exact owner-routed evidence. | `IND-BROWSER-001` only when its activation event occurs. | `RESOLVED — CONDITIONAL`: inactive for S-00 through S-06; reactivates on the named frontend/contract/routing/grid diff. |
| RB-003 | A maintainer implementation charter is required. | Repository writes must remain within an explicit path and compatibility envelope. | The current user task and the exact active charter in Section 7. | `CLOSED`: the current task authorizes W-01 and S-00 through S-06. |
| RB-004 | S-01 Records-resolution characterization gate. | Without a baseline, a boundary correction could obscure unrelated regression or freeze an owner violation. | S-00 owner selections, repair/preservation classification, live capability review, and focused `loadByDedupeTx` digest. | `CLOSED FOR S-01`: every conformant baseline is green and every mismatch is assigned to deterministic S-01 evidence. |
| RB-005 | Replacing compound `loadByDedupeTx` is outside S-01 and the active refactor. | Decomposition without an explicit cross-owner lock protocol could break canonical convergence or create deadlocks and inconsistent snapshots. | A separately authorized decision satisfying every `IND-DEDUPE-001` gate. | `DEFERRED`: does not block S-01 because S-01 MUST leave the query unchanged. |

No `BLOCKED: owner contradiction` entry is present because no contradiction was found during this session.

## 12. Binary Completion Criteria

Tracker completion and implementation readiness are separate states. W-00 through S-06 are complete; the authorized remediation has passed final validation and handoff.

| Requirement or decision | Workflow | Slice or activation | Validation/evidence | Binary acceptance |
| --- | --- | --- | --- | --- |
| `IND-BOUNDARY-001` | W-01 | Before S-00 | Decision/Core/Base-manifest/export-role/import-topology evidence | RB-001 is `CLOSED`, IDs were collision-checked, and the adopted topology is executable. |
| `IND-AUTHORIZATION-001` | W-00 | Before W-01 | Current maintainer task charter | RB-003 is `CLOSED`, allowed/excluded paths and stop conditions are explicit. |
| `IND-CHAR-001` | WF-03/WF-06 | S-00 | Focused and service-backed retained results | Every matrix row passes deterministically with no sleep-based race. |
| `IND-RECORDS-001` and `IND-ERROR-001` | S-01 | S-01 | Pre/post characterization, Records and Indicators rows, boundary and harness checks | Two standalone SQL reads are absent; one sorted snapshot is used; owner-specific public errors, durable-effect absence, and safe failures conform to Core 04. |
| `IND-DEDUPE-001` | WF-04/WF-06 | S-01 and deferred backlog | Focused source digest/diff plus concurrency rows | Compound query is unchanged in S-01; future replacement remains inactive. |
| Owner vocabulary convergence and integrity | S-02A/S-02B | S-02A/S-02B | Set-equality, owner, portability, rollback, route, migration, direct-write, and drift evidence | Exactly one runtime registry owns each token family and the database enforces unique support references. |
| Source-state catalog | WF-05/WF-06 | S-03 | Catalog, portability, Recovery, active-claim, and drift evidence | Exact 3/1/3 inventory derives from one immutable validated catalog. |
| Facade cohesion | S-04 | S-04 | Export, constructor, replay-byte, route, build, and boundary evidence | Import path and bytes are unchanged; the reviewed export surface is exactly 50 declarations. |
| Final validation and handoff | WF-08 | S-06 | Affected owner slices, boundary/drift/policy/shape gates, builds, finalization, fast/full graphs, protected-invariant audit, and complete diff ledger | Every required gate passes, every skip/failure is explained and resolved, and no obsolete, excluded, contradictory, or unrecorded path remains. |
| `IND-BROWSER-001` | WF-07 | Conditional frontend activation | One exact active Playwright row and `make browser-e2e-webserver-backed` | All five scenarios resolve exactly once; a broad target without the row cannot satisfy it. |

| Criterion | Current status | Evidence |
| --- | --- | --- |
| Every file in `internal/modules/indicators` is inventoried or explicitly out of scope. | PASS | Section 2 contains one row for each of 78 live files; none is omitted. |
| Every discovered public contract risk has an owner and test posture. | PASS | Section 4 covers HTTP, generic Workbook routes, WebSocket consequences, source behavior, projections, revisions, portability, recovery, generated contracts, frontend, authorization, and harness accounting. |
| Every remediation workflow has dependencies and exit criteria. | PASS | Section 6 records the strict checkpoint sequence. |
| Every remediation slice preserves owner-conformant behavior and bounds intentional corrections. | PASS | Section 7 names the exact compatibility corrections, rollback/stop rules, and exclusions. |
| Validation commands are discovered and current. | PASS | Section 8 defines the workstream and final validation ladder using public Make targets. |
| Contradictions are marked `BLOCKED: owner contradiction`. | PASS | No contradiction found; Section 11 records the required future handling. |
| Repository/framework mismatches are recorded as planning findings. | PASS | PF-001 in Section 1. |
| Adopted governance is sourced from owners rather than this tracker. | PASS | Decision, `REQ-00-074`, `AC-560`, Base manifest, and executable evidence agree; Section 3 remains a navigation record. |
| Every blocker has one defined state and transition. | PASS | Section 11 defines the state vocabulary and RB-001 through RB-005 without undifferentiated `OPEN` rows. |
| Every tracker-local requirement maps to workflow, execution, evidence, and acceptance. | PASS | The traceability table above. |
| Prior handoff history is preserved and the implementation session is current. | PASS | Section 10 retains every earlier row and appends a checkpoint row for each completed workstream. |
| W-00 changed only the tracker and did not begin W-01 early. | PASS | Current Git status, W-00 log, Markdown result root, and staged/unstaged whitespace checks. |
| S-06 completed the required narrow-to-broad validation order and final handoff. | PASS | Section 10 records every owner and gate root, the repaired Operator golden, the explicit `RESULTS_DIR` posture, the 69-path ledger, and the final protected-invariant audit. |

The overall tracker is complete. Every active work row is `DONE`; the only remaining items are explicitly dropped, conditionally inactive, or deferred outside this authorization.

## 13. Iteration 2 Charter and Planning Posture

Iteration 1 remains complete history. Sections 1 through 12 are not reopened,
reinterpreted, or replaced by this iteration. Iteration 2 begins with a
document-only activation and plans a separately authorized production-readiness
pass over the same bounded Indicators owner module.

| Item | Iteration 2 decision or live evidence |
| --- | --- |
| Planning baseline | Clean `main` at `0f7a33b0cf1c9405484b2c6ae17432d916da9a97` on 2026-08-23. |
| Target | `internal/modules/indicators` and only the direct owner, application-composition, verification, policy, and test consumers named by an active workstream. |
| Live package inventory | 78 Go files. The exact 50-declaration root export guard remains green at this baseline. |
| Live verification inventory | 37 owner-routed rows: 35 Go and two Vitest; 24 unit and 13 integration; 13 service-backed. |
| Current compound-query evidence | `loadByDedupeTx` lines 20 through 61 retain `sha256:3343ded88d9e54fe69ba516dbfe1df22c35be956a626d1682fa52ec823a85af4`. |
| Current status | `I2-DOC-00`, `I2-DOC-01`, and `I2-S00` through `I2-S06` are `DONE`; all Iteration 2 acceptance rows are closed with no unresolved blocker. |
| Tracker authority | This tracker controls sequencing, evidence, rollback, and handoff only. Adopted owners continue to govern product behavior. |
| Compatibility posture | Broader retirement was permitted for planning, but live evidence supports no feature-level retirement. Internal Go and test seams may break atomically without shims when their production replacement is named here. |
| Public behavior posture | Preserve all adopted public HTTP, authorization, cursor, JSON, idempotency, history, projection, Collaboration, portability, Recovery, and Network Flow behavior. |
| Migration posture | No database migration is planned. Existing production idempotency rows must remain replayable. |

### 13.1 Active and retained capabilities

Absence of a repository-internal caller is not sufficient evidence that a
public capability is dead. The live client, adopted owners, portability
catalog, and source-state catalog were reviewed together. The following
capabilities remain materially useful to the production shape and are retained.

| Capability | Live evidence | Iteration 2 disposition |
| --- | --- | --- |
| Eight Indicator child HTTP operations | All eight are adopted by Core 01/Core 04, registered by `httpapi.RegisterRoutes`, and referenced by the Workbook web client. | Retain paths, operation IDs, admission, statuses, request/success shapes, and authorization order. |
| Indicator, observation, and lifecycle portable families | All three are authoritative source-owner contributions with current bundle schemas, tests, Recovery classification, and history behavior. | Retain the exact three portability descriptors and valid bytes. |
| Indicator type, value-kind, observation-status, lifecycle-state, and origin vocabularies | The closed sets are adopted and consumed at creation, HTTP, rollback, and portability boundaries. | Retain exact membership, including owner-required `ipv4_addr`. |
| Network Flow Indicator participant | Network Flow has live transaction-participant callers and Core-owned `binding_only` behavior. | Retain result shape, schema ID value, status values, and explicit `OperationOccurred`. |
| Compound canonical-dedupe query | Iteration 1 deliberately deferred decomposition because no replacement cross-owner lock protocol exists. | Keep byte-for-byte unchanged under `IND-DEDUPE-001`. |

### 13.2 Scope and exclusions

Iteration 2 may amend the adopted Indicators boundary decision and `AC-560`
only for the exact internal topology and export changes in `I2-S01` and
`I2-S02`. It does not authorize a new product feature, public route, public
error code, database schema, migration, portable schema/version, browser
behavior, or owner relocation.

The following remain excluded:

- frontend, Inspector, grid-adapter, OpenAPI, and generated protocol changes;
- canonical-dedupe SQL or lock-order changes;
- removal or reinterpretation of `ipv4_addr` or another adopted token;
- idempotency-row deletion, rehash migration, silent repair, or replay
  invalidation;
- compatibility aliases, forwarding wrappers, dual constructors, or
  deprecation periods;
- hand edits to generated artifacts or existing migrations; and
- opportunistic cleanup in another owner.

An unexpected public contract, frontend, Inspector, database, portable-byte,
or generated protocol diff stops the active workstream. The tracker records
`BLOCKED: scope expansion` and implementation does not continue without new
authorization. An adopted-owner contradiction is recorded as
`BLOCKED: owner contradiction` without selecting a side.

## 14. Iteration 2 Delta Inventory and Remediation Matrix

Staticcheck and the Go compiler already reject ordinary unreachable local
code. The cleanup candidates below are production-reachability, construction,
ownership, malformed-state, and compatibility-surface findings established by
live source and caller inspection.

| ID | Gap and live evidence | Remediation and areas | Rationale and long-term benefit | Compatibility or migration impact | Risk if unresolved | Completion evidence |
| --- | --- | --- | --- | --- | --- | --- |
| `I2-GAP-001` | Construction contradicts the adopted boundary: Indicators constructs a concrete Records store, accepts nil Clock fallback, and checks interface dependencies only with `== nil`. HTTP also falls back to wall-clock time. | **Specification, implementation, tests:** amend the boundary decision and `AC-560`; inject a typed Records envelope port; require Clock; reject nil and typed-nil dependencies; require the HTTP clock. | Makes application composition the single dependency boundary, converts latent startup panics into deterministic errors, and keeps time testable. | Internal Go constructor migration only. Production assembly already supplies a clock. No wire or data migration. | A typed-nil dependency can survive startup and panic later; hidden clock fallback makes behavior nondeterministic and violates fail-fast construction. | Nil/typed-nil matrices pass; no partial Store or route service is returned; source scans find no Records construction or clock fallback in Indicators. |
| `I2-GAP-002` | Store owns three forwarding service objects that point back to the Store, three empty repository namespace values, and a revision adapter whose only purpose is method forwarding. | **Implementation, tests:** make Store methods the direct application entry points; replace stateless repository methods with explicitly named package functions in their concern files; align the private Revisions port with the real appender methods; delete obsolete types, fields, and `repositories.go`. | Removes cyclic self-reference and layers with no independent state or policy. Each concern remains cohesive without pretending to be a separate service or repository object. | Internal-only compile-time migration. No method behavior, SQL, transaction, or result change. | New behavior can be placed in the wrong layer, dependencies stay hidden, and high-fanout changes remain harder to reason about. | Removed-symbol and Store-field scans pass; all mutation/list/transaction tests pass; the dedupe query digest is unchanged. |
| `I2-GAP-003` | Mutation application methods trust hashes supplied by Workbook or HTTP adapters. Tests and internal callers can substitute arbitrary replay identity. | **Specification, implementation, tests:** make Indicators privately derive all five established logical hash forms; remove hash arguments and fields and the exported admission hash; remove only the Indicator use of Workbook preferred-hash override. | The owner that validates the semantic command also owns its replay identity, preventing false replay/conflict behavior and eliminating adapter coupling. | Internal Go break only. Preserve the exact deployed preimages and SHA-256 bytes so existing idempotency rows replay. No rehash or data migration. | A caller can cause an unrelated payload to replay or force a conflict, and future adapters can drift from owner normalization. | Golden preimages/digests, seeded retained-row replay, divergent payload, route-scope, and no-caller-override tests pass; stored JSON/status behavior is unchanged. |
| `I2-GAP-004` | Rollback validation re-reads untyped maps, uses unchecked assertions after partial validation, silently converts malformed nullable values to `nil`, ignores unknown-only patches, and omits complete presentation-field equality. | **Implementation, tests:** introduce one closed, fallible typed row-patch parser used by validation and restore; distinguish absent from explicit `null`; reject unknown, empty, blank, or wrongly typed patches; validate the complete canonical state before SQL. | Converts malformed retained history into a safe classified failure instead of a destructive clear and provides one extension point when a versioned field is added later. | Invalid historical values that were never owner-conformant are rejected with the existing safe error. Valid partial patches and explicit clears remain supported. No migration. | Malformed retained history can silently erase values, panic, or mutate a row using a partially validated identity. | Complete malformed-field matrix proves `ErrTargetNotReversible` and zero SQL effects; valid partial, clear, omission, dedupe, hash, defanged, and STIX cases pass. |
| `I2-GAP-005` | Several symbols are exported only for same-package implementation or test convenience: the HTTP service type, participant schema constant, five identity helpers, two projection fact helpers, and the admission hash. | **Specification, implementation, tests:** exchange the participant constant for the production Records port in the exact 50-root-export inventory; privatize owner-local helpers; make cross-owner tests consume the production projection contribution. | Prevents tests and future code from coupling to implementation facts while retaining the smallest production contract surface. | Internal Go break with atomic repository caller migration. Returned schema ID and projection facts remain exact. No aliases. | Test-only usage becomes permanent compatibility burden and blocks later structural changes. | Exact 50-role guard and repository-wide definition/caller scans pass; production contribution parity remains exact. |
| `I2-GAP-006` | Test support publishes a mutable four-example slice although only entry zero is used, includes four unused fields, and exposes mutable time variables. | **Tests/test support:** replace it with one immutable value-returning fixture containing only used fields and value-returning time helpers. | Makes fixtures intention-revealing, mutation-safe, and cheaper to extend without implying unsupported coverage. | Test-only compile-time migration. | Dead examples misrepresent coverage and shared mutable fixtures permit order-dependent tests. | No `Examples` slice or mutable time global remains; all migrated fixtures and owner rows pass. |
| `I2-GAP-007` | New or moved tests can drift from authored verification rows, generated topology, and export/import policy. | **Tests, tooling, tracker:** reconcile authored routing after each implementation slice and regenerate only through Make when an authored generator input changes. | Keeps executable evidence aligned with semantic ownership and makes the final handoff resumable. | Harness-only. Runtime must not read evidence or Markdown. | Coverage can become unselected, duplicated, or attributed to the wrong owner. | Every active test resolves once; harness, generation, artifact, shape, boundary, and catalog checks pass. |

## 15. Iteration 2 Execution Policy and Tracker Gate

Workstreams are serial. An implementation session performs this checkpoint
protocol for every active workstream:

1. Reconfirm live callers, adopted owners, worktree state, and dependency
   workstream status; record the workstream `IN_PROGRESS` in this tracker.
2. Make only the independently reversible change named by that workstream.
3. Run the narrow affected-owner unit/static evidence, then service-backed
   evidence where applicable, then the workstream's policy checks.
4. On a related failure, retain its result root, record diagnosis and rollback
   decision, and do not begin the next workstream.
5. On success, record changed files, substantive behavior, commands and roots,
   failures and their resolution, skips, rollback posture, remaining risks,
   and next action; mark the workstream `DONE`.
6. Run and record `make lint-markdown` after the checkpoint update and before
   beginning the next workstream.

Only the latest independently failing workstream may be rolled back. Tests,
owners, policies, and generated checks are not weakened to obtain a pass. User
changes outside the active slice are preserved. The next workstream may begin
only when every declared dependency is `DONE` and the preceding checkpoint is
complete.

## 16. Iteration 2 Sequential Workstreams

| ID | Workstream | Depends on | Planned change | Required evidence and exit criteria | Primary risk and rollback | Status |
| --- | --- | --- | --- | --- | --- | --- |
| `I2-DOC-00` | Document-only activation | Iteration 1 complete | Append Sections 13 through 19; preserve Sections 1 through 12; change no product, owner, test, contract, migration, or generated artifact. | Only this tracker differs; post-edit Markdown lint and scoped whitespace/status checks pass. | Accidental implementation or history rewrite. Revert only the appended planning body if the path audit is not tracker-only. | DONE |
| `I2-S00` | Rebaseline and characterize | `I2-DOC-00`; `I2-DOC-01`; separate implementation authorization | Refresh branch, commit, worktree, package/export inventory, production callers, test routing, dedupe digest, five idempotency preimages, retained-row replay, and valid/malformed rollback behavior. Classify every proposed deletion by production, test, tool, generated, and owner use. | Indicators focused and service-backed slices, backend boundary, and `test-fast` pass from the live baseline. Every repair case is assigned without committing a skipped or failing test. | Live drift can invalidate a deletion or hash assumption. Amend this plan before mutation; do not preserve a stale candidate. | DONE |
| `I2-S01` | Construction and facade cohesion | `I2-S00` | Amend the boundary decision and `AC-560` atomically with the implementation. Inject `RecordEnvelopePort`; require Clock and typed dependency validity; use Store time for autonomous create/child mutations; preserve participant time. Privatize HTTP service; remove service self-wrappers, repository namespace values, revision forwarding, and the obsolete exported schema constant. | Exact 50 reviewed root exports; nil/typed-nil and clock tests; app.server, Indicators, Records, Workbook, Network Flow, and Revisions affected slices; build and backend boundary pass. No concrete Records construction or time fallback remains in Indicators. Dedupe source/digest is exact. | Broad constructor migration can miss a composition or test caller. Roll back S01 atomically if a live caller cannot use the typed port; do not add an overload or default. | DONE |
| `I2-S02` | Owner-controlled idempotency | `I2-S01` | Amend the boundary decision/`AC-560` as needed for owner hash responsibility. Move exact create, observation-create, resolve, action, and lifecycle hashes into private owner code; remove hash arguments/fields, admission export, and Indicator preferred-hash use. | Golden current digests, existing persisted-hash replay, exact/divergent replay, route scope, no-side-effect, stored-byte, Workbook, HTTP, and Indicators transaction tests pass. No migration or compatibility branch exists. | A changed preimage would convert a valid retry into conflict. Seed the old digest before exercising new code; roll back S02 on any mismatch rather than versioning or deleting rows. | DONE |
| `I2-S03` | Rollback hardening | `I2-S02` | Replace permissive map helpers with one closed typed patch parser shared by validate/restore; make all recognized fields, nullability, partial-patch, canonicalization, and unknown-member behavior explicit. | Indicators and Revisions rollback rows pass; every malformed recognized or unknown value fails safely before update; valid partial and clear cases, history effects, and portability round trips remain exact. | An overly strict parser could reject valid current snapshots. Exercise actual source snapshots and partial selections; roll back S03 if a valid owner snapshot cannot be represented without weakening the closed parser. | DONE |
| `I2-S04` | Dead surface and test-support removal | `I2-S03` | Privatize remaining caller-free identity/projection helpers; migrate Projections tests to the production contribution; replace dead/mutable Indicators test fixtures; delete obsolete tests/comments after callers move. | Repository-wide callers are classified; Indicators and Projections focused/service-backed slices pass; export, private-package, contribution-parity, immutable-fixture, and removed-symbol scans pass with no alias. | A tool or generator can be a non-obvious consumer. Refresh repository-wide scans before deletion and restore only a proven production contract, not test convenience. | DONE |
| `I2-S05` | Harness and policy reconciliation | `I2-S04` | Reconcile authored verification/family rows, root export reasons, backend boundaries, and generated topology activated by S01 through S04. Run `make generate` only if an authored input changed. | Every test resolves exactly once; no stale selector, duplicate row, unapproved import/export, generated drift, artifact-policy failure, or JSON-shape mismatch remains. | Generated output may expose a stale authored owner fact. Repair the authored fact and regenerate; never hand-edit the generated output. | DONE |
| `I2-S06` | Validation and handoff | `I2-S05` | Remove any final proven residue, execute Section 18 in order, reconcile final inventory and compatibility assertions, append complete handoff evidence, and close every acceptance row. | Every applicable target passes with retained roots; failures/skips are classified; no obsolete path, contradiction, excluded diff, migration edit, browser claim, or generated hand edit remains. | Broad validation can expose baseline failures. Related failures return to their owning I2 slice; unrelated failures are recorded without rewriting user work. | DONE |

### 16.1 Checkpoint ledger

| Workstream | Started | Completed | Status | Files, commands, roots, outcome, rollback, and next action |
| --- | --- | --- | --- | --- |
| `I2-DOC-00` | 2026-08-23 | 2026-08-23 | DONE | Changed only `docs/handoffs/indicators-module-refactor-tracker.md`; preserved Sections 1 through 12 and appended the Iteration 2 charter, inventory, remediation, serial plan, interface migration, validation, acceptance, deferrals, and handoff templates. Planning-body Markdown lint passed at `.cartulary/test-results/20260823T144036Z-p3457725`; the final evidence-replacement lint passed at `.cartulary/test-results/20260823T144134Z-p3458907`. No implementation, owner, test, contract, migration, generated artifact, rollback, or skip. `I2-S00` remains inactive pending separate authorization. |
| `I2-S00` | 2026-08-23 | 2026-08-23 | DONE | Reconfirmed `main` at `0f7a33b0cf1c9405484b2c6ae17432d916da9a97`, the staged-plus-unstaged tracker-only worktree, 78 Indicators files, exactly 50 guarded root exports, and 37 routed rows with 13 service-backed. Classified every named constructor, Store-method, hash, rollback, identity, projection, fixture, tool, and generated caller; no stale candidate invalidated the plan. The Iteration 1 wrapper digest remains `3343ded88d9e54fe69ba516dbfe1df22c35be956a626d1682fa52ec823a85af4`; the protected SQL digest is `d665f06c2526b0118e33eaa887da279ad54025967618662c2a0b47b7bfde857b`. Indicators focused passed 19/19 execution units at `.cartulary/test-results/20260823T150223Z-p3470731`; service-backed passed 8/8 at `.cartulary/test-results/20260823T150314Z-p3486834`; backend boundary passed 3/3 at `.cartulary/test-results/20260823T150402Z-p3502714`; `test-fast` passed 425/425 at `.cartulary/test-results/20260823T150407Z-p3503164`; checkpoint Markdown passed at `.cartulary/test-results/20260823T150449Z-p3504047`. No characterization test, production edit, failure, skip, rollback, drift, or blocker. `I2-S01` is the sole next action. |
| `I2-S01` | 2026-08-23 | 2026-08-23 | DONE | Added the required root Records port and private HTTP reader, composed one Records Store in `internal/app/server/runtime_assembly.go`, required injected clocks and nil/typed-nil validity, moved orchestration directly to Store, converted repository receivers to named functions, assigned Revisions directly, privatized HTTP Service and the participant schema constant, deleted `repositories.go`, protected the exact SQL literal, migrated all callers, and amended the boundary decision and `AC-560`. The first focused run failed at `.cartulary/test-results/20260823T151500Z-p3514021` because the Store field retained the removed private interface name; it was corrected and all reruns passed. Final format passed at `.cartulary/test-results/20260823T152556Z-p3793110`; Indicators focused 19/19 at `.cartulary/test-results/20260823T152600Z-p3796871`; HTTP construction row at `.cartulary/test-results/20260823T151643Z-p3547616`; Indicators service-backed 8/8 at `.cartulary/test-results/20260823T151729Z-p3548588`; Records 8/8 at `.cartulary/test-results/20260823T151816Z-p3564626`; Workbook 65/65 at `.cartulary/test-results/20260823T151851Z-p3580400`; Network Flow 35/35 at `.cartulary/test-results/20260823T152104Z-p3638160`; Revisions 27/27 at `.cartulary/test-results/20260823T152314Z-p3695080`; app.server 24/24 at `.cartulary/test-results/20260823T152417Z-p3740194`; build-server at `.cartulary/test-results/20260823T152515Z-p3780492`; backend boundary 3/3 at `.cartulary/test-results/20260823T152529Z-p3792514`; checkpoint Markdown lint at `.cartulary/test-results/20260823T152740Z-p3813453`. No skip, compatibility shim, migration, generated edit, or rollback was required. The raw SQL digest remains `d665f06c2526b0118e33eaa887da279ad54025967618662c2a0b47b7bfde857b`; exact 50 exports and public behavior remain green. `I2-S02` is the sole next action. |
| `I2-S02` | 2026-08-23 | 2026-08-23 | DONE | Added one private owner hash boundary for the exact create, observation-create, observation-resolve, dismiss/restore action, and lifecycle JSON forms; derives before owner normalization, preserves nullable and optional membership and caller support-reference order, and uses current JSON plus SHA-256 semantics. Removed all Indicator `RequestHash` fields and parameters, the admission hash export, HTTP hash derivation, and only Indicator use of Workbook's preferred override. Added exact preimage/digest goldens, hard-coded deployed-row replay with zero durable record effects, route/scope isolation, exact/divergent route replays, and defensive lifecycle sorting. The first `make format` preflight failed without a retained root because the admission test entrypoint had been renamed; restoring its routed name corrected the authored-selector mismatch without changing evidence semantics. Final format passed at `.cartulary/test-results/20260823T154602Z-p4053706`; full Indicators focused passed 19/19 at `.cartulary/test-results/20260823T154222Z-p3975144`; the final hash-golden row passed at `.cartulary/test-results/20260823T154606Z-p4057481`; Indicators service-backed passed 8/8 at `.cartulary/test-results/20260823T154316Z-p3992309`; Workbook focused passed 65/65 at `.cartulary/test-results/20260823T153646Z-p3855798` and service-backed 37/37 at `.cartulary/test-results/20260823T153901Z-p3913600`; Revisions focused passed 27/27 at `.cartulary/test-results/20260823T154403Z-p4008143`; backend boundary passed 3/3 at `.cartulary/test-results/20260823T154627Z-p4058429`; build-server passed at `.cartulary/test-results/20260823T154629Z-p4058877`; checkpoint Markdown lint passed at `.cartulary/test-results/20260823T154728Z-p4071103`. No persisted row, route, status/JSON contract, migration, alternate digest, compatibility branch, generated artifact, skip, or rollback changed. `I2-S03` is the sole next action. |
| `I2-S03` | 2026-08-23 | 2026-08-23 | DONE | Replaced the permissive rollback helpers with one closed typed parser shared by validation and restore, locked and overlaid the current Indicator row, validated canonical identity and presentation before UPDATE, and made identity/dedupe mismatch safe. Added parser matrices plus public rollback evidence for malformed-history atomicity and valid full, partial, omitted-field, explicit-clear, rekey, hash, defanged, and STIX restoration. Final format passed at `.cartulary/test-results/20260823T160057Z-p4102141`; the focused parser row passed at `.cartulary/test-results/20260823T155858Z-p4085121`; Indicators focused passed 19/19 at `.cartulary/test-results/20260823T160109Z-p4106022` and service-backed passed 8/8, including portability rows, at `.cartulary/test-results/20260823T160225Z-p4167640`; Revisions focused passed 27/27 at `.cartulary/test-results/20260823T160109Z-p4106019`; the Indicator rollback plus transaction-failure service rows passed 3/3 at `.cartulary/test-results/20260823T160225Z-p4167646`. The first public rollback-row run also passed 3/3 at `.cartulary/test-results/20260823T155905Z-p4085550`; checkpoint Markdown lint passed at `.cartulary/test-results/20260823T160402Z-p5424`. No failure, rerun after failure, skip, migration, portable-schema change, generated edit, compatibility path, or rollback was required. The SQL literal digest remains exact. `I2-S04` is the sole next action. |
| `I2-S04` | 2026-08-23 | 2026-08-23 | DONE | Privatized five identity helpers and both projection fact helpers with no aliases; moved the cross-owner compiled-plan test through fallible `NewContribution`; replaced the mutable four-example/time globals with `PrimaryExample()`, `BaseTime()`, and `PastTime()` value factories; removed four unused fixture fields; migrated every caller; renamed the obsolete Indicator rollback-helper test; and extended the exact export guard to lock private helper and minimal immutable-fixture structure. Format passed at `.cartulary/test-results/20260823T160916Z-p9406`; Indicators focused passed 19/19 at `.cartulary/test-results/20260823T160934Z-p13386` and service-backed passed 8/8 at `.cartulary/test-results/20260823T161027Z-p47647`; Projections focused passed 15/15 at `.cartulary/test-results/20260823T160934Z-p13392` and service-backed passed 11/11 at `.cartulary/test-results/20260823T161027Z-p47646`; the exact 50-export/private-helper/fixture guard passed at `.cartulary/test-results/20260823T161135Z-p79613`; the production-contribution catalog parity row passed 3/3 at `.cartulary/test-results/20260823T161135Z-p79622`; checkpoint Markdown lint passed at `.cartulary/test-results/20260823T161301Z-p95800`. Repository-wide production, test, tool, and generated caller scans found no stale Indicator export, mutable fixture, direct projection-fact caller, obsolete helper name, or alias. No failure, skip, generated edit, compatibility path, or rollback occurred. `I2-S05` is the sole next action. |
| `I2-S05` | 2026-08-23 | 2026-08-23 | DONE | Added the renamed closed rollback-parser entrypoint to the existing Indicators rollback-admission selector and removed its temporary duplicate invocation from the vocabulary entrypoint. Ran `make generate` immediately after the authored family change; only the tool-managed execution-topology render-index digests changed. Generation passed at `.cartulary/test-results/20260823T161449Z-p98676`; format passed at `.cartulary/test-results/20260823T161512Z-p101836`; `test-catalog-check` passed with no retained run root emitted; Indicators focused passed 19/19 at `.cartulary/test-results/20260823T161534Z-p106008`; harness contract passed 2/2 at `.cartulary/test-results/20260823T161624Z-p123318`; backend boundary passed 3/3 at `.cartulary/test-results/20260823T161624Z-p123274`; generation drift passed 4/4 at `.cartulary/test-results/20260823T161624Z-p122965`; generated-artifact policy passed 3/3 at `.cartulary/test-results/20260823T161624Z-p122993`; JSON shape passed 3/3 at `.cartulary/test-results/20260823T161624Z-p123021`; checkpoint Markdown lint passed at `.cartulary/test-results/20260823T161715Z-p127843`. Every active selected test resolves exactly once; no stale selector, duplicate row, unapproved boundary, generated hand edit, failure, rerun, skip, or rollback remains. `I2-S06` is the sole next action. |
| `I2-S06` | 2026-08-23 | 2026-08-23 | DONE | Re-ran every affected focused and service-backed owner slice, then completed the Section 18 ladder in the required order. Focused roots: Indicators 19/19 `.cartulary/test-results/20260823T161829Z-p130197`, Records 8/8 `.cartulary/test-results/20260823T161829Z-p130188`, Workbook 65/65 `.cartulary/test-results/20260823T161829Z-p130170`, Projections 15/15 `.cartulary/test-results/20260823T161829Z-p130194`, Network Flow 35/35 `.cartulary/test-results/20260823T162051Z-p234870`, Revisions 27/27 `.cartulary/test-results/20260823T162051Z-p234863`, and app.server 24/24 `.cartulary/test-results/20260823T162051Z-p234879`. Service-backed roots: Indicators 8/8 `.cartulary/test-results/20260823T162326Z-p376017`, Records 5/5 `.cartulary/test-results/20260823T162326Z-p376028`, Workbook 37/37 `.cartulary/test-results/20260823T162326Z-p376038`, Projections 11/11 `.cartulary/test-results/20260823T162326Z-p376052`, Network Flow 30/30 `.cartulary/test-results/20260823T162549Z-p479816`, Revisions 20/20 `.cartulary/test-results/20260823T162549Z-p479810`, and app.server 17/17 `.cartulary/test-results/20260823T162549Z-p479813`. Ordered ladder roots: backend boundary 3/3 `.cartulary/test-results/20260823T162817Z-p618836`; `test-catalog-check` passed without a retained root; harness 2/2 `.cartulary/test-results/20260823T162833Z-p619708`; generation drift 4/4 `.cartulary/test-results/20260823T162849Z-p620284`; artifact policy 3/3 `.cartulary/test-results/20260823T162901Z-p623254`; JSON shape 3/3 `.cartulary/test-results/20260823T162907Z-p623716`; migration drift 5/5 `.cartulary/test-results/20260823T162913Z-p624190`; build-server `.cartulary/test-results/20260823T162924Z-p627273`; build-migrate `.cartulary/test-results/20260823T162937Z-p639374`; agent-finalize 1/1 `.cartulary/test-results/20260823T162944Z-p641198`; test-fast 425/425 `.cartulary/test-results/20260823T163000Z-p644097`; check 646/646 `.cartulary/test-results/20260823T163046Z-p650930`; ladder Markdown lint `.cartulary/test-results/20260823T163515Z-p769696`; completed-handoff Markdown lint `.cartulary/test-results/20260823T164024Z-p772549`. `RESULTS_DIR` was unset because no qualifying successful full warm-check root existed before agent-finalize, so retained-run maintenance was intentionally skipped. Final branch/HEAD, 52-path worktree, staged-user-work, whitespace, generated-root, protected-SQL, excluded-path, removed-symbol, eight-operation, and exact-50-export audits passed. No S06 failure, rerun, unexplained skip, rollback, owner contradiction, browser claim, or remaining workstream exists. |

## 17. Iteration 2 Internal Interface Migration Map

These are repository-internal Go changes. All callers migrate in the owning
workstream without aliases, overloads, forwarding packages, or a deprecation
period.

| Current interface or surface | Planned interface or disposition | Compatibility invariant |
| --- | --- | --- |
| `StoreDependencies` constructs Records internally. | Add required `RecordEnvelopePort` with `InsertTx`, `LoadEnvelopesTx`, and `AdvanceVersionTx`; application/test composition supplies the Records implementation. | Records owns envelopes and caller transactions remain borrowed. |
| Optional `StoreDependencies.Clock` and HTTP wall-clock fallback. | Clock is required and nil is a deterministic construction error. Store time owns autonomous create and child mutations. | Network Flow participant continues to use required `OperationOccurred`; public timestamp meanings remain unchanged. |
| `CreateIndicatorRow(ctx, actor, incidentID, command, requestHash, requestID, now)`. | `CreateIndicatorRow(ctx, actor, incidentID, command, requestID)`. | Owner-derived hash bytes and production timestamps remain compatible. |
| Observation/lifecycle parameter structs include `RequestHash`. | Remove every `RequestHash` field; owner derives the exact current preimage before mutation normalization changes its representation. | Existing production idempotency rows replay; same key plus divergent semantic payload still conflicts. |
| Exported `admission.CreateRequestHash`. | Delete; retain only private owner hash construction and golden evidence. | Workbook request admission and public payloads are unchanged. |
| Exported `IndicatorFindOrCreateParticipantV1`. | Private exact-value constant; add exported production `RecordEnvelopePort` so the reviewed root total stays 50. | Returned `SchemaID`, `Status`, and result shape remain exact. |
| Exported `httpapi.Service`. | Private `service`; `RegisterRoutes` remains the production route entry point. | All eight registered operations remain exact. |
| Exported same-package identity helpers. | Privatize `NormalizeIndicatorType`, `NormalizeValueKind`, `NormalizeValue`, `IsIPType`, and `DedupeKey`; keep cross-package semantic entry points only. | Canonical values and dedupe bytes remain exact. |
| Exported projection `Descriptor` and `SurfaceIntent`. | Private facts reached through the fallible production `NewContribution`. | Descriptor, intent, field, source, and ordering parity remain exact. |
| Four-element mutable `testsupport.Examples` plus mutable `BaseTime`/`PastTime`. | One value-returning primary example with only used fields and value-returning time helpers. | Test semantics remain deterministic; no production API is involved. |

The current idempotency hash preimages are deliberately preserved rather than
redesigned. They retain the current client transaction membership and current
support-reference ordering even where those facts are redundant with the
route key or later semantic sorting. This is not a general compatibility shim:
it is the single owner implementation required to compare new retries with
already committed production rows. A future incompatible hash algorithm would
require an explicit versioned owner contract and migration strategy.

## 18. Iteration 2 Validation and Binary Acceptance

### 18.1 Workstream routing

`I2-S00` runs:

1. `make test-slice OWNER=module.indicators`
2. `make service-backed-test-slice OWNER=module.indicators`
3. `make backend-module-boundary-check`
4. `make test-fast`
5. `make lint-markdown`

Every implementation workstream begins with the narrowest changed-owner rows,
then the full affected owner slice, then its service-backed slice when present.
The affected owner set is expanded only by live changed paths:

- `I2-S01`: Indicators, Records, Workbook, Network Flow, Revisions, and
  `app.server`;
- `I2-S02`: Indicators and Workbook plus route/service-backed replay evidence;
- `I2-S03`: Indicators and Revisions;
- `I2-S04`: Indicators and Projections; and
- `I2-S05`: every owner whose authored verification or boundary input changed.

`make generate` runs immediately after an authored generator input changes and
before drift checks. It does not run merely to rewrite stable outputs.

### 18.2 Final validation order

`I2-S06` runs changed-owner focused and service-backed slices, then:

1. `make backend-module-boundary-check`
2. `make harness-contract`
3. `make generate-drift`
4. `make generated-artifact-policy-check`
5. `make json-shape-check`
6. `make migration-drift`
7. `make build-server`
8. `make build-migrate`
9. `make agent-finalize`, with `RESULTS_DIR` only when a qualifying retained
   successful full warm-check root exists; otherwise record that retained-run
   maintenance was skipped because it was unset.
10. `make test-fast`
11. `make check`
12. `make lint-markdown`

No direct browser row is required because frontend behavior is outside the
active plan. An unexpected frontend, Inspector, OpenAPI, generated protocol,
or browser-family diff activates `IND-BROWSER-001`, stops the current
workstream, and requires separate authorization before merge.

### 18.3 Binary acceptance criteria

| ID | Requirement | Pass condition | Current state |
| --- | --- | --- | --- |
| `I2-AC-001` | Historical integrity and activation | Sections 1 through 12 remain intact; only this tracker changed in `I2-DOC-00`; every later slice remains separately authorized and serial. | DONE |
| `I2-AC-002` | Fail-fast construction | Records, time, projections, source text, Revisions, and Postgres dependencies have one explicit construction path that rejects nil and typed nil without partial capability or fallback. | DONE — `I2-S01` |
| `I2-AC-003` | Cohesive Store | No forwarding service/self-cycle, stateless repository namespace value, revision forwarding adapter, concrete Records construction, or obsolete repository file remains. | DONE — `I2-S01` |
| `I2-AC-004` | Minimal production surface | The root has exactly 50 production-role exports after the constant/port exchange; named child/helper exports are private and no alias remains. | DONE — `I2-S04` |
| `I2-AC-005` | Owner idempotency | The owner derives every exact established hash; existing stored rows replay; divergence conflicts; no caller override or rehash migration exists. | DONE — `I2-S02` |
| `I2-AC-006` | Safe rollback | One closed typed parser handles validation and restore; malformed history fails safely before SQL; valid partial and explicit-clear behavior remains exact. | DONE — `I2-S03` |
| `I2-AC-007` | Dead test support removed | Only used immutable fixtures remain; no dead example, unused field, mutable time global, obsolete test, or stale comment remains. | DONE — `I2-S04` |
| `I2-AC-008` | Protected behavior | Eight HTTP operations, public bytes, authorization order, canonical identity, history, projection, Collaboration, 3/1/3 source inventory, portability, Recovery, participant behavior, and current vocabularies remain owner-conformant. | DONE — `I2-S06` |
| `I2-AC-009` | Dedupe protection | `loadByDedupeTx`, its lock order, and its focused digest remain unchanged. | DONE — `I2-S01` |
| `I2-AC-010` | Harness and generated integrity | Every test resolves once; all boundary, generation, artifact, shape, migration, build, and harness checks pass without a generated hand edit. | DONE — `I2-S06` |
| `I2-AC-011` | Final handoff | Section 18.2 passes in order; every failure, skip, result root, changed file, rollback, deferred item, and final status is recorded. | DONE — `I2-S06` |

## 19. Iteration 2 Deferrals and Handoff

### 19.1 Deferred items

| ID | Item | Reason and activation rule | Status |
| --- | --- | --- | --- |
| `I2-DEF-001` | Decompose `loadByDedupeTx`. | Still requires the separately authorized `IND-DEDUPE-001` cross-owner lock protocol and concurrency proof. | DEFERRED |
| `I2-DEF-002` | Retire a public Indicator operation or source family. | Every current operation/family is adopted and live. Activate only with product-owner evidence, replacement/retention policy, and explicit authorization. | DEFERRED |
| `I2-DEF-003` | Change `ipv4_addr` or another adopted token. | Core 02 requires the token and existing state depends on its exact meaning. Requires a versioned vocabulary and data/API migration plan. | DEFERRED |
| `I2-DEF-004` | Redesign or version idempotency hashes. | Existing committed rows must replay. Activate only with an owner-approved version discriminator and safe transition strategy. | DEFERRED |
| `I2-DEF-005` | Frontend, Inspector, grid, OpenAPI, protocol, or browser work. | Inactive for this backend iteration. Any named diff activates `IND-BROWSER-001` and requires separate authorization. | RESOLVED — CONDITIONAL |

### 19.2 Document-only handoff

| Date | Scope | Files changed | Verification | Outcome |
| --- | --- | --- | --- | --- |
| 2026-08-23 | Iteration 2 document-only planning and activation | `docs/handoffs/indicators-module-refactor-tracker.md` only | Markdown lint passed at `.cartulary/test-results/20260823T144036Z-p3457725` and `.cartulary/test-results/20260823T144134Z-p3458907`; scoped Git status and whitespace audit | Sections 1 through 12 remain Iteration 1 history; `I2-DOC-00` is complete; `I2-S00` through `I2-S06` and `I2-AC-002` through `I2-AC-011` remain `PLANNED`; no implementation or owner artifact changed. |

### 19.3 Future implementation handoff template

Every future checkpoint and the final handoff record:

- branch, commit, worktree state, and active authorization;
- files added, modified, deleted, or generated through Make;
- substantive owner, implementation, test, policy, and documentation edits;
- exact commands, result roots, row counts, failures, reruns, and skips;
- compatibility evidence for public bytes, persisted hashes, timestamps,
  canonical identity, history, portability, and participant results;
- rollback taken or explicitly not required;
- blocker transitions and deferred-item activation state; and
- the only eligible next workstream.

Iteration 2 implementation readiness is intentionally false after
`I2-DOC-00`. The next permitted action is a separately authorized `I2-S00`
rebaseline and characterization pass, not a production edit.

## 20. Iteration 2 Decision-Complete Amendment

### 20.1 Amendment checkpoint

| Workstream | Started | Completed | Status | Files, commands, roots, outcome, rollback, and next action |
| --- | --- | --- | --- | --- |
| `I2-DOC-01` | 2026-08-23 | 2026-08-23 | DONE | Appended only Section 20 to this tracker; preserved Sections 1 through 19 and changed no owner, implementation, test, contract, migration, or generated artifact. Recorded the split Records capabilities, required clocks and typed-nil posture, protected SQL-literal digest, five exact replay forms, closed rollback fields, amended sequence, and binary evidence. Planning-body and post-checkpoint Markdown lint passed at `.cartulary/test-results/20260823T150108Z-p3468205` and `.cartulary/test-results/20260823T150138Z-p3469410`; no failure, skip, rollback, or blocker. `I2-S00` is the sole next action. |

### 20.2 Controlling clarifications

This amendment is additive. Sections 1 through 12 remain closed Iteration 1
history, and Sections 13 through 19 retain the `I2-DOC-00` planning record.
Where an Iteration 2 detail below is more specific, it controls `I2-S00`
through `I2-S06` without changing any adopted product behavior.

1. Application composition MUST construct one Records adapter and supply it to
   both the root Store and the HTTP adapter. The root exports
   `RecordEnvelopePort` with `InsertTx`, `LoadEnvelopesTx`, and
   `AdvanceVersionTx`; HTTP accepts a separate private read-only capability
   with `LoadEnvelope`. No production Indicators package constructs Records.
2. Store construction MUST reject nil or typed-nil Postgres, Revisions,
   Records, Projections, SourceText, or Clock dependencies and return no
   partial Store. HTTP construction MUST likewise reject nil or typed-nil
   owner and Records capabilities and a nil Clock. No Indicators wall-clock
   fallback remains.
3. The Iteration 1 line-range digest
   `sha256:3343ded88d9e54fe69ba516dbfe1df22c35be956a626d1682fa52ec823a85af4`
   remains historical evidence. Iteration 2 protects the raw compound-query
   SQL bytes and lock order with
   `sha256:d665f06c2526b0118e33eaa887da279ad54025967618662c2a0b47b7bfde857b`.
   Receiver removal may change wrapper bytes but MUST NOT change this SQL
   digest, parameter order, joins, predicates, limit, or `FOR UPDATE` set.
4. Indicators privately derives the five deployed SHA-256 replay forms before
   later service normalization can change their representation: create;
   observation create; observation resolve; dismiss/restore action; and
   lifecycle append. Create hashes the view schema, client transaction,
   required identity fields, and only present optional representation fields.
   Child hashes preserve the current admitted JSON members. Lifecycle hashing
   retains nullable members as JSON `null` and preserves caller support-ref
   order before service sorting. Route and scope keys remain outside these
   preimages exactly as deployed.
5. Indicator row rollback accepts only `record_id`, `incident_id`,
   `indicator_type`, `value_kind`, `display_value`, `normalized_value`,
   `dedupe_key`, `defanged_value`, `hash_algorithm`, `hash_value`, and
   `stix_pattern` in the retained `source` object. Absent members retain the
   current value. Explicit `null` clears only nullable members. Unknown,
   empty, blank, NUL-containing, null-required, wrongly typed, invalid-token,
   invalid-UUID, malformed hash-pair, identity-mismatched, or dedupe-mismatched
   input returns the existing safe target-not-reversible classification before
   UPDATE. Validation and restore use the same typed parser and validate the
   complete overlaid canonical state, including presentation-field equality.
6. `docs/domain.md` remains unchanged because this iteration changes internal
   topology and conformance rather than domain vocabulary. Database,
   frontend, Inspector, grid, OpenAPI, generated protocol, portable schema,
   public route, and feature-retirement work remain excluded.

### 20.3 Amended workstream sequence

The strict dependency chain is now `I2-DOC-00` → `I2-DOC-01` → `I2-S00`
→ `I2-S01` → `I2-S02` → `I2-S03` → `I2-S04` → `I2-S05` →
`I2-S06`. Every workstream MUST be marked `IN_PROGRESS` before its first
non-tracker edit and `DONE` with retained evidence before its successor begins.

`I2-S01` owns construction, direct Store orchestration, stateless repository
receiver removal, direct Revisions-port alignment, the participant-constant to
Records-port export exchange, HTTP service privatization, owner decision and
`AC-560` amendments, and complete caller migration. `I2-S02` owns replay-hash
derivation and removal of all Indicators caller hash overrides. `I2-S03` owns
the typed rollback parser. `I2-S04` owns remaining owner-local export and test
fixture cleanup. `I2-S05` owns authored verification and generated topology
reconciliation. `I2-S06` owns final validation and handoff.

### 20.4 Amended binary evidence

- The root export surface remains exactly 50 declarations by exchanging
  `IndicatorFindOrCreateParticipantV1` for `RecordEnvelopePort`.
- `CreateIndicatorRow` accepts only context, actor, incident, command, and
  request ID. A non-replay captures Store time once. Import time and Network
  Flow `OperationOccurred` retain their distinct owner semantics.
- Observation and lifecycle parameter structs contain no `RequestHash`.
  Existing committed idempotency rows replay against byte-identical owner
  digests without migration, rehash, alternate algorithm, or compatibility
  branch.
- The malformed rollback matrix proves safe classification and zero UPDATE,
  revision, projection, idempotency-success, or Collaboration effect. Full
  snapshots, valid partial patches, omissions, explicit nullable clears,
  rekeys, hash pairs, defanged values, and STIX values remain reversible.
- Final validation runs affected owner slices followed by backend boundary,
  test catalog, harness, generation drift, generated-artifact policy, JSON
  shape, migration drift, server and migrate builds, agent finalization,
  `test-fast`, `check`, Markdown lint, and final protected-invariant and scope
  audits in that order.

### 20.5 `I2-S01` changed-path record

The completed construction and cohesion slice changed no generated, migration,
frontend, OpenAPI, protocol, portable-schema, or domain-vocabulary path.

- Owner decisions: `docs/decisions/indicators-module-boundary.md` and
  `docs/spec/04_security_deployment_and_conformance.md`.
- Application and test composition:
  `internal/app/server/runtime_assembly.go`,
  `internal/app/importassembly/owner_registry_test.go`,
  `internal/app/importassembly/tasksdecisions_integration_test.go`,
  `internal/app/workbookassembly/indicator_adapter.go`, and
  `internal/testutil/appsupport/workbook.go`.
- Indicators production:
  `internal/modules/indicators/contracts.go`, `create_service.go`,
  `store_composition.go`, `revision_append_port.go`,
  `source_repository.go`, `observation_service.go`,
  `observation_repository.go`, `lifecycle_service.go`,
  `lifecycle_repository.go`, and `httpapi/routes.go`; deleted
  `internal/modules/indicators/repositories.go`.
- Indicators evidence:
  `internal/modules/indicators/exported_surface_test.go`,
  `store_composition_test.go`, `store_test_helpers_test.go`,
  `transaction_atomicity_test.go`, `indicators_test.go`, `unit_test.go`,
  `active_identity_claims_integration_test.go`,
  `portability_characterization_test.go`,
  `target_resolution_integration_test.go`,
  `httpapi/construction_test.go`, and
  `httpapi/vocabulary_admission_test.go`.
- Adjacent-owner evidence: `internal/modules/networkflow/store_test.go`,
  `internal/modules/revisions/indicator_children_test.go`, and
  `internal/modules/workbook/notes_indicators_test.go`.
- Controlling execution record:
  `docs/handoffs/indicators-module-refactor-tracker.md`.

The only related failure was the recorded stale Store field type on the first
focused compile. The correction was local, no rollback or compatibility path
was introduced, no required evidence was skipped, and the sole remaining
construction risk is future application composition adding a dependency
without extending the closed constructor matrices. `I2-S02` is active and is
the only eligible successor.

### 20.6 `I2-S02` changed-path record

The completed owner-idempotency slice added
`internal/modules/indicators/idempotency_hash.go` and
`internal/modules/indicators/idempotency_hash_test.go`; changed
`internal/modules/indicators/contracts.go`, `create_service.go`,
`observation_service.go`, `lifecycle_service.go`, `child_coordination.go`,
`httpapi/decoding.go`, `httpapi/routes.go`, `admission/create.go`,
`admission/create_test.go`, `store_composition_test.go`,
`production_contract_test.go`, `child_routes_integration_test.go`,
`unit_test.go`, `transaction_atomicity_test.go`,
`store_test_helpers_test.go`, `indicators_test.go`,
`active_identity_claims_integration_test.go`,
`portability_characterization_test.go`, and
`target_resolution_integration_test.go`. Direct consumers changed in
`internal/app/workbookassembly/indicator_adapter.go`,
`internal/modules/workbook/notes_indicators_test.go`, and
`internal/modules/revisions/indicator_children_test.go`. This controlling
tracker records the slice; the owner decision and `AC-560` already contained
the decision-complete hash amendment from `I2-S01`.

No database, generated, frontend, Inspector, grid, OpenAPI, protocol,
portable-schema, domain-vocabulary, or feature path changed. The selector-name
preflight failure and correction are preserved in the ledger. No evidence was
skipped, no rollback was required, and the remaining risk is confined to
malformed retained rollback data owned by `I2-S03`, now the only eligible
successor.

### 20.7 `I2-S03` changed-path record

The completed rollback-hardening slice changed
`internal/modules/indicators/internal/providers/rollback/provider.go`,
`provider_test.go`, and `vocabulary_test.go`, plus the cross-owner public
transaction evidence in
`internal/modules/revisions/indicator_children_test.go`. The parser recognizes
only the eleven owner fields listed in Section 20.2, represents omission and
nullable clear distinctly, validates exact tokens and UUIDs without unchecked
assertions, and is the sole admission path for validation and restore. Restore
locks the current Indicator source row, overlays the patch, canonicalizes the
complete result, verifies identity, representation, and supplied dedupe facts,
and performs no UPDATE on an invalid result.

Malformed parser inputs were exercised without a transaction capability to
prove rejection before any query. Public retained-history cases additionally
proved that unknown fields, mismatched record or incident identity, malformed
hash pairs, noncanonical presentation, and incorrect dedupe return
`target_not_reversible` with byte-identical durable Records, Indicators,
active-identity, projection, change-set, mutation, revision, and idempotency
state. Valid full and partial snapshots proved omission, explicit nullable
clears, rekeying, hash pairs, defanged values, STIX values, projection rebuild,
and active-identity synchronization. The complete Indicators service-backed
slice retained its portability evidence.

No owner specification, database migration, portable schema, frontend,
Inspector, grid contract, OpenAPI, protocol, generated artifact, public route,
or domain vocabulary changed. No check failed or was skipped, no rollback was
required, and the remaining mutable production/test surface is confined to
`I2-S04`, the only eligible successor after the checkpoint lint.

### 20.8 `I2-S04` changed-path record

The completed dead-surface slice changed
`internal/modules/indicators/internal/identity/identity.go` and
`identity_test.go`;
`internal/modules/indicators/workbookprojection/contribution.go` and
`contribution_test.go`;
`internal/modules/indicators/testsupport/fixtures.go`;
`internal/modules/indicators/exported_surface_test.go`, `unit_test.go`,
`resolution_integration_test.go`, `child_routes_integration_test.go`,
`portability_characterization_test.go`, and
`internal/providers/rollback/provider_test.go`; plus the cross-owner consumer
`internal/modules/projections/internal/runtime/query_plans_test.go`. This
tracker is the only documentation path in the slice.

`NormalizeIndicatorType`, `NormalizeValueKind`, `NormalizeValue`, `IsIPType`,
and `DedupeKey` are now owner-local implementation functions. Indicator
projection descriptor and semantic-intent facts are likewise private and are
observed across owners only through `NewContribution`. The root remains
exactly 50 reviewed production-role exports. Test support now returns one
four-field primary example by value and returns both fixed times by value;
there is no mutable example slice, mutable time global, unused fixture field,
compatibility alias, or forwarding helper.

No specification, database migration, portable schema, frontend, Inspector,
grid contract, OpenAPI, protocol, generated artifact, public route, replay
hash, persisted row, or domain vocabulary changed. All caller classes were
rescanned after the migration. No check failed or was skipped, no rollback was
required, and authored selector reconciliation for the renamed rollback test
is intentionally confined to `I2-S05`, the only eligible successor after the
checkpoint lint.

### 20.9 `I2-S05` changed-path record

The completed harness slice changed the authored selector
`tools/test_families/module.indicators.json` and the corresponding test
organization in
`internal/modules/indicators/internal/providers/rollback/vocabulary_test.go`.
`make generate` changed only the tool-managed input hashes in
`tools/execution_topology_render_index.json`; it changed no task surface,
execution schedule, product contract, generated Go or TypeScript, migration,
or runtime path. This tracker is the only documentation path in the slice.

The closed rollback-parser matrix and rollback vocabulary checks now each have
one explicit selected entrypoint in their existing owner row. Catalog closure,
harness contract, backend boundaries, generation drift, generated-artifact
policy, and JSON shape all pass. The 37-row Indicators baseline remains 37
rows; the existing row now names both active tests without duplication or
selector drift.

No specification, database migration, portable schema, frontend, Inspector,
grid contract, OpenAPI, protocol, public route, replay hash, persisted row, or
domain vocabulary changed. No generated file was hand-edited, no check failed
or was skipped, no rollback was required, and `I2-S06` is the only eligible
successor after the checkpoint lint.

### 20.10 `I2-S06` final handoff and 52-path ledger

Iteration 2 is complete on branch `main` at unchanged implementation-base HEAD
`0f7a33b0cf1c9405484b2c6ae17432d916da9a97`. The pre-existing staged tracker
work remains staged and preserved; all implementation-session changes remain
visible in the shared worktree. The final status contains exactly 52 paths:
three added files, one deleted file, 47 authored modified files, and one
Make-generated modified index.

- Owner and handoff documentation:
  `docs/decisions/indicators-module-boundary.md`,
  `docs/handoffs/indicators-module-refactor-tracker.md`, and
  `docs/spec/04_security_deployment_and_conformance.md`.
- Application composition and adjacent-owner evidence:
  `internal/app/importassembly/owner_registry_test.go`,
  `internal/app/importassembly/tasksdecisions_integration_test.go`,
  `internal/app/server/runtime_assembly.go`,
  `internal/app/workbookassembly/indicator_adapter.go`,
  `internal/modules/networkflow/store_test.go`,
  `internal/modules/projections/internal/runtime/query_plans_test.go`,
  `internal/modules/revisions/indicator_children_test.go`,
  `internal/modules/workbook/notes_indicators_test.go`, and
  `internal/testutil/appsupport/workbook.go`.
- Indicators production implementation:
  `internal/modules/indicators/admission/create.go`,
  `internal/modules/indicators/child_coordination.go`,
  `internal/modules/indicators/contracts.go`,
  `internal/modules/indicators/create_service.go`,
  `internal/modules/indicators/httpapi/decoding.go`,
  `internal/modules/indicators/httpapi/routes.go`,
  added `internal/modules/indicators/idempotency_hash.go`,
  `internal/modules/indicators/internal/identity/identity.go`,
  `internal/modules/indicators/internal/providers/rollback/provider.go`,
  `internal/modules/indicators/lifecycle_repository.go`,
  `internal/modules/indicators/lifecycle_service.go`,
  `internal/modules/indicators/observation_repository.go`,
  `internal/modules/indicators/observation_service.go`,
  `internal/modules/indicators/revision_append_port.go`,
  `internal/modules/indicators/source_repository.go`,
  `internal/modules/indicators/store_composition.go`,
  `internal/modules/indicators/workbookprojection/contribution.go`, and deleted
  `internal/modules/indicators/repositories.go`.
- Indicators tests and test support:
  `internal/modules/indicators/active_identity_claims_integration_test.go`,
  `internal/modules/indicators/admission/create_test.go`,
  `internal/modules/indicators/child_routes_integration_test.go`,
  `internal/modules/indicators/exported_surface_test.go`, added
  `internal/modules/indicators/httpapi/construction_test.go`,
  `internal/modules/indicators/httpapi/vocabulary_admission_test.go`, added
  `internal/modules/indicators/idempotency_hash_test.go`,
  `internal/modules/indicators/indicators_test.go`,
  `internal/modules/indicators/internal/identity/identity_test.go`,
  `internal/modules/indicators/internal/providers/rollback/provider_test.go`,
  `internal/modules/indicators/portability_characterization_test.go`,
  `internal/modules/indicators/production_contract_test.go`,
  `internal/modules/indicators/resolution_integration_test.go`,
  `internal/modules/indicators/store_composition_test.go`,
  `internal/modules/indicators/store_test_helpers_test.go`,
  `internal/modules/indicators/target_resolution_integration_test.go`,
  `internal/modules/indicators/testsupport/fixtures.go`,
  `internal/modules/indicators/transaction_atomicity_test.go`,
  `internal/modules/indicators/unit_test.go`, and
  `internal/modules/indicators/workbookprojection/contribution_test.go`.
- Harness inputs and generated evidence:
  `tools/test_families/module.indicators.json` and Make-generated
  `tools/execution_topology_render_index.json`.

The final diff has no database migration, `docs/domain.md`, frontend,
Inspector, grid-contract, OpenAPI, generated protocol, portable-schema,
feature-retirement, or browser path. Generated Go and TypeScript roots are
untouched; the one generated index was produced by `make generate` and passes
drift and artifact-policy checks. The protected SQL literal independently
hashes to
`d665f06c2526b0118e33eaa887da279ad54025967618662c2a0b47b7bfde857b`,
the root export guard remains exactly 50, and all eight HTTP operations remain
registered.

The complete failure history is preserved: the first S01 focused compile at
`.cartulary/test-results/20260823T151500Z-p3514021` exposed the stale removed
Store field type and passed after the local correction; the first S02 format
preflight emitted no retained root and exposed a temporarily renamed authored
selector, which was restored until its deliberate S05 reconciliation. No
other workstream or final-ladder check failed. No rollback was required.
Direct browser evidence was not run because no frontend or browser claim is in
scope. `RESULTS_DIR` was unset for `agent-finalize` because no qualifying full
warm-check root existed at that ordered point; retained-run maintenance was
therefore the only intentional skip.

`I2-DEF-001` through `I2-DEF-004` remain deferred behind their stated future
authorization gates, and the conditional browser gate remains inactive. No
database or public-contract migration occurred: repository-internal Go callers
migrated atomically while routes, operation IDs, status/JSON bytes, persisted
idempotency rows, canonical identity, history, projections, portability,
Recovery, Collaboration, and Network Flow participation remained compatible.
Every Iteration 2 workstream and acceptance row is `DONE`; there is no next
eligible Iteration 2 slice or unresolved blocker.
