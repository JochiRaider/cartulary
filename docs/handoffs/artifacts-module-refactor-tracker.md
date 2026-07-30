# artifacts Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Item | Current posture |
| --- | --- |
| Target path | `internal/modules/artifacts` |
| Target label | `artifacts`, derived from the target path and normalized to lowercase kebab case |
| Output path | `docs/handoffs/artifacts-module-refactor-tracker.md` |
| Status | Remediation complete; WS-00 through WS-08 are `DONE` |
| Allowed change | This tracker, the modular-refactor planning framework, owner-aligned tests and harness inputs, generated projections through Make-owned generation, and production/composition code admitted by S-01 through S-06 |
| Non-goals | No database migration, public route or wire redesign, frontend redesign, reporting-family expansion, import-target expansion, public test endpoint, runtime fault switch, or manual generated-file edit |
| Implementation authority | The 2026-07-30 user request explicitly authorizes execution of WS-00 through WS-08 under the gates in this tracker |

Post-closure Markdown validation passed at
`.cartulary/test-results/20260730T153725Z-p229477`, and the corresponding
`git diff --check` reported no whitespace errors. The handoff is complete; no
successor workstream is active.

The key words **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY**
are normative for execution of this refactor and its evidence closure. They do
not promote this tracker into a product owner. If a requirement in this tracker
conflicts with an adopted owner, the adopted owner prevails and the affected
work MUST be marked `BLOCKED: owner contradiction`.

The requirements in this tracker determine the permitted refactor and
verification outcome. Internal method names, private file decomposition, and
equivalent private data structures remain intentional implementation choices
unless a requirement below fixes them because they affect ownership,
transactionality, or interoperability.

The source hierarchy used for this tracker is:

1. Adopted subsystem NLSpecs within their named scopes.
2. Core 00 through Core 04 for current implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. Domain vocabulary and implementation-support guides.
5. Current repository code and tests.
6. Prior trackers, handoffs, and the planning framework as evidence only.

Core 05 is not applicable to this non-claim-bearing refactor. No timed or
fixture-sensitive publication claim is proposed.

Owner and support documents inspected:

- `AGENTS.md`;
- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`;
- `docs/domain.md`;
- `docs/spec/00_document_set_status_and_precedence.md`;
- relevant artifact, projection, transport, and module-owner sections of
  `docs/spec/01_architecture_storage_and_view_contracts.md`;
- relevant record, artifact, history, rollback, and source-owner sections of
  `docs/spec/02_domain_model_schema_and_history.md`;
- relevant saved-view, conflict, collaboration, and workbook sections of
  `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`;
- relevant authorization sections of
  `docs/spec/04_security_deployment_and_conformance.md`;
- the scope boundary in
  `docs/spec/05_claim_publication_and_benchmark_reproducibility.md`;
- relevant owner, command, evidence-accounting, and product-subordination
  sections of `docs/testing-harness-nlspec.md`; and
- relevant source-provider sections of `docs/reporting-subsystem-nlspec.md`;
- `docs/research/nlspec-spec.md`, as writing and completeness guidance rather
  than product authority; and
- `temp/analysis-notes.md`, as supporting closure analysis rather than
  authority.

Repository evidence inspected:

- every file under `internal/modules/artifacts`, inventoried individually in
  Section 2;
- composition callers in `internal/app/importassembly`,
  `internal/app/incidentportabilityassembly`, `internal/app/projectionassembly`,
  `internal/app/recoveryassembly`, `internal/app/revisionassembly`, and
  `internal/app/workbookassembly`;
- direct consumers in `internal/modules/imports`, `internal/modules/projections`,
  `internal/modules/reporting`, `internal/modules/revisions`, and
  `internal/modules/workbook`;
- the authored artifact view-schema, import-target, incident-bundle,
  projection-provider, recovery, OpenAPI, and verification inputs under
  `contracts`;
- the artifact test-family manifest and Make-owned target surface under
  `tools`;
- representative generated Go and TypeScript projections under the declared
  generated roots;
- artifact source and projection migrations under `db/migrations`; and
- generic frontend view-contract, workbook query/collaboration, selector, and
  grid-adapter consumers under `apps/web` and `packages`.

No owner contradiction was found. If later inspection finds two adopted owners
that prescribe incompatible behavior, the affected work must be marked
`BLOCKED: owner contradiction`; an implementation task must not choose a side.

No Core 00 through Core 04 amendment and no Testing Harness NLSpec amendment is
required to close RB-001 or RB-002. If S-00 reveals an actual unowned or
contradictory behavior, the affected test and refactor slice MUST stop until the
primary owner adopts or repairs the semantic. A test MUST NOT turn current code,
generated output, visible UI text, or this tracker into product authority.

The original planning framework did not establish `artifacts` as a permanent
module in its default module catalog. Live Core 02 ownership text, current
contracts, and repository composition do establish artifact source behavior.
WS-00 closed that documentation gap by naming the artifact source owner without
treating the package as a bounded context or moving workbook, relationship,
revision, projection-lifecycle, or reporting orchestration ownership into it.
The existence of `internal/modules/artifacts` alone would not have been
sufficient proof.

### 1.1 Execution checkpoint protocol

Only one remediation workstream may be `IN_PROGRESS`. Before beginning its
successor, the active workstream MUST:

1. record its base commit and the repository status at entry;
2. record changed files, substantive decisions, compatibility removals,
   validation commands and result roots, failure attribution, rollback
   posture, and residual risk in Section 13;
3. run `make lint-markdown` after updating this tracker;
4. become `DONE` or `BLOCKED`; and
5. mark its successor `IN_PROGRESS` only after the preceding four steps.

An owner-backed S-00 failure is not frozen merely because the implementation
already behaves that way. It MUST be placed in a separate `S-00R-n`
conformance-correction workstream before structural movement. An unowned
behavior or owner contradiction remains `BLOCKED: owner contradiction`.
Feature expansion remains outside this refactor.

### 1.2 Remediation gap register

| Gap | Required remediation | Areas | Compatibility posture | Completion evidence |
| --- | --- | --- | --- | --- |
| Planning and ownership drift | Add the artifact source owner to the non-normative modular-refactor framework with explicit exclusions. | Documentation | No runtime impact. | Framework review and Markdown lint |
| Inadequate executable characterization | Implement the exact S-00 symbols, matrices, fault boundaries, and transitional evidence adoption. | Tests, harness | No product change; transitional catalog contains 10 rows. | RB-001 binary exit |
| Fragmented source policy | Consolidate exact surface, scalar-storage, value, nullability, writeability, collection, and subtype policy; reject unknown fields before SQL. | Implementation, tests | No schema or wire change. | Exact-policy and negative tests |
| Duplicate transaction coordination | Reuse one transaction-supplied artifact mutation kernel from workbook and contextual-note entrypoints through typed ports. | Implementation, tests | Preserve route, storage, revision, replay, and event semantics. | Atomicity and replay evidence |
| Duplicated surface/filter registry | Use one cycle-free artifact-owned eight-surface registry and generated canonical filters. | Implementation, tests | Retain useful root ID exports; remove local fallbacks. | Root/provider parity evidence |
| Adapter responsibility leakage | Keep import, rollback, portability, recovery, reporting, revision, and projection contributions thin over the source boundary. | Implementation, tests | Preserve caller transactions and frozen family/file sets. | Affected owner slices |
| Obsolete internal compatibility | Migrate composition, then remove unused constructors, duplicate helpers, and the independent linked-note mutation path. | Implementation, composition, tests | No permanent runtime alias or feature flag. | Caller search and boundary checks |
| Incomplete evidence accounting | Complete the immutable 10-to-9 row cutover and add durable boundary guardrails. | Harness, generated projections, tests | Retain terminal disposition, never aliases. | RB-002 binary exit |
| Validation and handoff | Run the ordered S-07 validation and close this tracker with attributable evidence. | Validation, documentation | Record skipped checks and rationale. | ART-012 and Section 12.4 |

## 2. Baseline Repository Inventory

This inventory records the pre-remediation state used to authorize and sequence
the work. Deleted compatibility files and their former responsibilities remain
listed as historical evidence; the final implementation disposition and changed
paths are recorded in Section 13.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/artifacts/collection_descriptors.go` | Declares artifact collection families, allowed operations, and link-target policy for tags, record refs, party refs, and handoff risk refs. | `CollectionFamily`, `CollectionPolicy`, policy methods, `LookupCollectionPolicy` | Artifact workbook mutation code | `module.links` collection vocabulary by semantic contract | Workbook coordination and collection tests exercise the policies through public mutation paths. | Authored view schemas declare collection fields; generated view contracts expose them. | `module.artifacts` | High | Source-owner policy; keep with artifact field semantics. |
| `internal/modules/artifacts/create_validation.go` | Validates artifact creates and direct scalar patches, including subtype minimum signal, enums, ranges, and closure/finding state. | `ValidationError`, `ValidateCreateParams`, `ValidateDirectPatchChange`, enum/range validators | `WorkbookFacade`, import creation, source store | Artifact `CreateParams` and `FieldValue` | Workbook create/patch tests provide indirect coverage. | View-schema writeback contracts and generated validators constrain the same public fields. | `module.artifacts` | High | Direct per-surface characterization is incomplete. |
| `internal/modules/artifacts/handoff_risk_refs.go` | Validates and applies handoff open-risk child-row mutations, owner checks, upsert, and tombstone behavior. | `HandoffOpenRiskRefsFieldKey`, risk action DTOs, validation, `Store` mutation methods, item-ref wrappers | Artifact workbook collection mutation path | PostgreSQL, UUIDs, artifact `Store`, `riskrefs` | Risk-ref grammar unit tests and workbook collection tests | Handoff view schema and incident-bundle/recovery contracts include risk references. | `module.artifacts` | High | Source mutation with relationship-like collection semantics; relationship ownership must remain explicit. |
| `internal/modules/artifacts/import_create.go` | Binds the generic import owner façade and creates artifact records, source rows, revisions, and refreshed projection rows in the caller transaction. | `ImportCreateCommand`, `NewImportCreateFacade`, `Store.CreateImportRowTx` | `internal/app/importassembly/owner_registry.go`, `module.imports` apply coordinator | Imports owner façade, records, revisions, projection refresh, auth/user lookup, PostgreSQL | Import registry/boundary tests; no direct artifact apply characterization found | `contracts/imports/view-targets.v1.json`, generated import registries and validators | `module.artifacts` source adapter; `module.imports` orchestration | High | Current targets enable notes, comm log, handoff, status review, and lesson; optional query/finding surfaces retain current availability. |
| `internal/modules/artifacts/import_projection.go` | Refreshes and reloads one imported artifact projection row using artifact query surfaces. | `Store.RefreshImportRowTx` | `CreateImportRowTx` | `module.projections`, `projectionprovider.QuerySurfaces` | Indirect import and projection tests | Projection-provider contract and generated view schemas | `module.artifacts` adapter with `module.projections` storage owner | Medium | Preserve post-create refresh and returned row shape. |
| `internal/modules/artifacts/incident_bundle_portability.go` | Exports and imports artifact-owned NDJSON files for artifact and subtype rows. | `ExportIncidentBundleFiles`, `ImportIncidentBundleFilesTx` | Artifact incident-bundle source port | Incident portability codecs/query interfaces, PostgreSQL | Incident-bundle route/integration tests compare artifact tables and rebuilt projections | `contracts/incident-bundles/source_catalog.json` | `module.artifacts` source adapter | High | Five current files are part of portability behavior; do not infer new files from names. |
| `internal/modules/artifacts/incident_bundle_source_port.go` | Supplies the artifact source descriptor, export/import functions, and invariant adapter to incident portability composition. | `NewIncidentBundleSourcePort` | `internal/app/incidentportabilityassembly/catalog.go` | Incident-bundle source-port contracts | Incident-bundle catalog and integration tests | Incident-bundle source catalog and generated descriptors | `module.artifacts` source adapter | Medium | Legitimate thin source-owner contribution. |
| `internal/modules/artifacts/incident_bundle_subtype_presence.go` | Declares artifact subtype-presence ownership for record-level incident-bundle validation. | `IncidentBundleSubtypeContribution` | `internal/app/incidentportabilityassembly/catalog.go` | Record subtype-presence contract, PostgreSQL | Incident-bundle catalog/validation tests | Incident-bundle subtype contracts | `module.artifacts` | Medium | Preserve supported record type and invariant behavior. |
| `internal/modules/artifacts/mutations.go` | Owns authoritative artifact/subtype SQL inserts and scalar updates, default persistence, finding lifecycle normalization, and row touching. | `Store`, `FieldValue`, `CreateParams`, `NewStore`, store methods, `IsArtifactBackedField` | Workbook, import, rollback-adjacent artifact paths | PostgreSQL, UUID/time, revisions appender | Direct workbook persistence test plus broader workbook and revision integration tests | Artifact schemas and generated view contracts; authored migration `00014_artifacts_and_optional_surfaces.sql` | `module.artifacts` | High | Direct SQL is legitimate source persistence; it should be hidden behind a smaller façade, not moved to a consumer. |
| `internal/modules/artifacts/recovery_state.go` | Declares authoritative artifact tables and derived artifact projection state for recovery catalog composition. | `RecoveryStateContribution` | `internal/app/recoveryassembly/state_catalog.go` | Recovery-state contract | Recovery assembly catalog and coverage tests | Recovery state fixtures/manifests | `module.artifacts` contribution; `module.recovery` orchestration | Medium | Legitimate owner declaration; projection rebuild remains derived-state behavior. |
| `internal/modules/artifacts/revision_provider_contribution.go` | Registers artifact delete/restore and rollback providers for all artifact-backed view schemas. | `RevisionProviderContribution` | `internal/app/revisionassembly/revisions.go` | `module.revisions`, `deleterestore`, `rollbackprovider` | Revision provider matrix and delete/restore/rollback integration tests | Eight authored/generated view schemas | `module.artifacts` source adapter; `module.revisions` generic coordination | High | Preserve the complete eight-view registration. |
| `internal/modules/artifacts/surfaces.go` | Defines eight stable artifact view-schema IDs and resolves each schema to its canonical artifact type through the view-schema registry. | Eight `*ViewSchemaID` constants, `ArtifactTypeForView`, `ArtifactTypeForSurface`, `IsArtifactBackedView` | Workbook/projection composition, revisions, projections, tests | `internal/platform/viewschema` generated registry | Projection, workbook, view-schema, and contract tests | All eight authored view schemas and generated Go/TypeScript registries | `module.artifacts` | High | Source-owner surface registry; overlaps filter knowledge in `projectionprovider/query_surfaces.go`. |
| `internal/modules/artifacts/workbook_conflict.go` | Resolves artifact same-field conflicts and delegates accepted resolutions to ordinary patch behavior. | `WorkbookConflictCommand`, `WorkbookFacade.ResolveConflict` | Workbook conflict provider in composition | Revisions conflict-resolution mechanics, artifact patch façade | Workbook conflict route and integration tests | OpenAPI conflict envelope and generated protocol types | `module.artifacts` source revalidation; `module.revisions` generic mechanics | High | Must preserve token opacity, stale-window behavior, authorization precedence, and no-side-effect rejection. |
| `internal/modules/artifacts/workbook_facade.go` | Coordinates artifact create/patch transactions, incident admission, idempotency, validation, collections, records, revisions, projections, and mutation results. | `WorkbookFacade`, request/command/result DTOs, conflict errors, `NewWorkbookFacade`, `Create`, `Patch` | `internal/app/workbookassembly`, `internal/modules/workbook` contribution adapters | PostgreSQL/authn plus incidents, links, projections, records, revisions, conflict tokens, idempotency | Direct note persistence test and broad workbook coordination/integration tests | OpenAPI workbook routes, eight view contracts, generated protocol and view-contract outputs | `module.artifacts` application façade | High | Legitimate façade with excessive orchestration and platform coupling; primary refactor seam. |
| `internal/modules/artifacts/workbook_facade_test.go` | Service-backed characterization that artifact workbook creation persists an artifact-backed note. | `TestArtifactWorkbookFacadePersistsArtifactBackedNotes_Unit` | Harness row for `module.artifacts` | PostgreSQL test support and workbook façade | The file is the test. | `tools/test_families/module.artifacts.json` selects this test. | `module.artifacts` tests | Medium | Despite its name, the harness classifies the row as integration. It is the only active artifact owner row. |
| `internal/modules/artifacts/deleterestore/provider.go` | Snapshots artifact source state, resolves its view schema, and supplies delete/restore source hooks. | `Source`, `NewSource`, revision source-interface methods | `RevisionProviderContribution`, `module.revisions` | PostgreSQL, records/revision contracts, artifact surface mapping | Revision delete/restore matrix and integration tests | Revision route contracts and eight view schemas | `module.artifacts` source adapter | High | Current source delete-state hook is intentionally a no-op because the record envelope owns deletion state. |
| `internal/modules/artifacts/linkednotes/facade.go` | Creates contextual note artifacts linked to timeline events, hosts, identities, or evidence while coordinating revisions, collections, projections, and idempotency. | `Facade`, create/action/result DTOs, `MutationValidationError`, `NewFacade`, `SourceIncident`, `Create` | `internal/app/workbookassembly/store.go`, linked-note workbook route | PostgreSQL/authn plus incidents, links, projections, records, revisions | Linked-note route/integration and workbook coordination tests | Linked-note OpenAPI/protocol contracts and note view schema | `module.artifacts` contextual application façade | High | Duplicates much of the root mutation sequence; source creation remains artifact-owned while contextual links remain link-owned. |
| `internal/modules/artifacts/linkednotes/revision_append_port.go` | Adapts the revisions appender to the narrow linked-note revision interface. | No exported surface; private adapter methods | `linkednotes.Facade` | `module.revisions` | Linked-note integration tests | Revision/change-set contract indirectly | `module.artifacts` adapter | Low | Already demonstrates the intended narrow-port direction. |
| `internal/modules/artifacts/projectionprovider/provider.go` | Refreshes or rebuilds `artifact_grid_projection` from records, artifacts, links, parties, and subtype tables. | `RefreshArtifactTx`, `RebuildIncidentArtifactsTx`; provider descriptor through package internals | Projection assembly/catalog and artifact mutation/import paths | PostgreSQL and projection-provider contracts | Projections query/rebuild and incident-bundle rebuild tests | `contracts/projection-providers/index.json`; generated view schemas | `module.artifacts` provider intent; `module.projections` storage lifecycle | High | Source-owner query intent is legitimate; derived-table lifecycle remains projection-owned. |
| `internal/modules/artifacts/projectionprovider/query_surfaces.go` | Declares eight artifact query surfaces, fields, collection SQL expressions, and canonical artifact-type filters. | `QuerySurfaces` | Projection assembly, artifact import refresh, projection query tests | Projection-provider contracts and view-schema registry | Projection query/sort/filter/group and workbook coordination tests | Eight view schemas and projection-provider contract | `module.artifacts` provider descriptor | High | Duplicates surface IDs/filter mapping from the root package; consolidate without changing SQL or wire rows. |
| `internal/modules/artifacts/reportingprovider/provider.go` | Materializes typed reporting facts from the artifact projection for six current artifact families. | `CollectFieldsTx`, `CollectFactsTx` | `internal/modules/reporting/export_materializer.go` | PostgreSQL and reporting export-provider contracts | Reporting boundary guard and reporting integration tests | Reporting subsystem contracts; snapshot route behavior indirectly | `module.artifacts` source reporting adapter | Medium | Current output covers notes, findings, comm log, handoffs, status reviews, and lessons. Preserve omission of investigative queries and forensic keywords unless separately authorized. |
| `internal/modules/artifacts/riskrefs/item_refs.go` | Encodes and strictly parses canonical handoff risk item references. | `RiskRefItemRef`, `ParseRiskRefItemRef` | Root handoff risk helpers | UUID | `riskrefs/item_refs_test.go` | Handoff collection wire values indirectly | `module.artifacts` | Low | Small semantic helper with clear artifact ownership. |
| `internal/modules/artifacts/riskrefs/item_refs_test.go` | Characterizes canonical and strict risk item-reference grammar. | Two unit tests | Direct Go test discovery | `riskrefs` helper | The file is the test. | No direct generated output | `module.artifacts` tests | Low | Retain as unit evidence even though it is not currently an active owner row. |
| `internal/modules/artifacts/rollbackprovider/provider.go` | Validates rollback source snapshots and restores authoritative scalar artifact/subtype fields. | `Provider`, `NewProvider`, revision rollback-provider methods | `RevisionProviderContribution`, `module.revisions` | PostgreSQL, rollback contracts, artifact field validation | Direct provider unit tests and revisions integration tests | Revision/rollback contracts and view schemas | `module.artifacts` source adapter | High | Collection cells are deliberately outside this scalar row restore mapping; their owners remain separate mutation targets. |
| `internal/modules/artifacts/rollbackprovider/provider_test.go` | Characterizes rollback mapping for every artifact variant and subtype invariants. | Two unit tests | Direct Go test discovery | `rollbackprovider` | The file is the test. | No direct generated output | `module.artifacts` tests | Medium | Partial provider characterization; transaction restore effects remain integration-covered. |

No target file is out of scope. The target path exists, and all 25 files are
represented above.

## 3. Module Boundary Diagnosis

The target is a legitimate artifact source owner and a mixed-responsibility
package. It currently acts as a thin owner contribution in several composition
paths, but `workbook_facade.go` and `linkednotes/facade.go` also make it a broad
application/mutation coordinator. It is projection- and persistence-adjacent,
but it is not a frontend controller, grid-vendor layer, or unrelated catch-all.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Artifact field validation, subtype defaults, and minimum-signal rules | Root artifact validation and mutation files | `module.artifacts` | keep | Core 02 artifact source-owner requirements and current Store/facade calls | Hide behind a smaller artifact façade; do not move to workbook transport. |
| Authoritative artifact and subtype persistence | `mutations.go`, handoff risk child-row code | `module.artifacts` | keep | Core 02 source/projection split and authored artifact migration | Direct SQL is acceptable inside the source owner. |
| Workbook create/patch transaction coordination | `workbook_facade.go` | `module.artifacts` source service with narrow peer ports | split | Composition calls one artifact owner; façade directly coordinates multiple owners | Split private components while preserving the public contribution until callers migrate. |
| Contextual linked-note creation | `linkednotes/facade.go` | `module.artifacts` for the note; `module.links` for contextual relationship | split | Core artifact-backed note semantics and current link creation | Reuse one artifact mutation coordinator; do not create a second note source owner. |
| Artifact collection policy | `collection_descriptors.go`, `handoff_risk_refs.go` | `module.artifacts`; link persistence remains `module.links` | keep | Field-key policies and link/risk child-row behavior | Preserve distinct item kinds and operation semantics. |
| Artifact query-provider intent | `projectionprovider` | `module.artifacts` | keep | Core 01 source-owner provider intent and projection-provider contract | `module.projections` continues to own derived storage and generic query coordination. |
| Artifact projection materialization lifecycle | Artifact provider called from projection assembly | `module.projections` with artifact provider contribution | defer | Current assembly and contract split ownership | No runtime ownership move is required for the first refactor slices. |
| Delete, restore, rollback, and conflict source revalidation | Revision provider subpackages and root conflict façade | `module.artifacts` source adapters plus `module.revisions` mechanics | keep | Core 02/03 owner text and revision composition | Preserve token, revision-window, and scalar/collection boundaries. |
| Import target creation | Root import adapter | `module.artifacts` source adapter plus `module.imports` orchestration | keep | Import owner registry and target contract | Keep the adapter thin and transaction-compatible. |
| Incident portability and recovery contributions | Root portability/recovery adapters | `module.artifacts` source contributions | keep | Authored source and recovery catalogs | Orchestration stays with incident bundles and recovery. |
| Reporting extraction | `reportingprovider` | `module.artifacts` typed source provider | keep | Adopted reporting source boundary and reporting materializer | Preserve current six-family coverage; expansion is deferred. |
| HTTP routing and authorization order | Workbook/revisions/import/incident-bundle/reporting transports outside target | Owning transport/platform modules | keep | OpenAPI, Core 03 conflict contract, Core 04 authorization model | Artifact source code performs required source revalidation but must not become the route owner. |
| Collaboration publication and replay | Downstream of artifact mutation results | `module.collaboration` | keep | Core 03 durable intent and `record_changed` requirements | Artifact mutations supply deterministic change facts; publication remains post-commit. |
| Saved-view persistence | Outside target; artifact views consume view-schema IDs | `module.savedviews` | keep | Core 03 saved-view distinction and current generic workbook surfaces | Refactor must preserve saved-view additivity and stable schema identity. |
| Frontend shell/controller and grid vendor behavior | `apps/web`, `packages/grid-adapter` | Existing frontend owners | defer | No direct target implementation dependency on `react-data-grid` | No frontend implementation slice is proposed. |

### 3.1 Transaction ownership contract

| Operation family | Transaction owner | Artifact boundary | Normative rule |
| --- | --- | --- | --- |
| Workbook create and patch | Artifact workbook façade | Artifact source mutation service | The façade MUST execute one logical mutation in one database transaction. No collaborator MAY commit an independently durable partial effect. |
| Artifact conflict resolution | Artifact workbook façade after route-owned admission | Artifact source revalidation and ordinary patch path | Accepted resolution MUST repeat current source-state validation and use the ordinary artifact write target. Rejection MUST commit no mutation effect. |
| Contextual linked-note create | Artifact linked-note façade | Note source mutation plus typed contextual-link capability | Note, optional tags, contextual link, history, projection, idempotency, and committed mutation facts MUST be atomic. |
| Import owner create | `module.imports` | `CreateImportRowTx(context.Context, pgx.Tx, ImportOwnerCreateCommand)` | Imports supplies `pgx.Tx`. The artifact façade MUST NOT begin, commit, or roll back that transaction. A later caller failure MUST roll back all artifact-owned effects. |
| Delete, restore, and rollback | `module.revisions` | Artifact delete/restore and rollback source adapters | The adapter MUST use the caller transaction and MUST NOT acquire independent commit authority. |
| Incident-bundle import | Incident Portability owner | Artifact source-port import callback | The adapter MUST use the admitted import transaction and preserve atomic target publication. |
| Projection refresh during mutation | Owner of the enclosing mutation transaction | Artifact projection-provider capability | Refresh MUST observe the same transaction's source state. It MUST NOT make projection state independently durable. |
| Reporting collection | `module.reporting` read/materialization transaction | Artifact typed reporting provider | The provider MUST return the current six-family fact set and MUST NOT expand reporting scope. |

### 3.2 Required capability boundaries

The implementation MAY choose private interface and method names. It MUST
preserve the following capability separation:

| Capability | Owning module | Input/output obligation | Forbidden coupling |
| --- | --- | --- | --- |
| Record envelope | `module.records` | Create or load the canonical incident-bound record and return stable `record_id` and row version facts. | Artifact code MUST NOT redefine record lifecycle or authorization. |
| Typed relationship mutation | `module.links` | Apply validated tag, record-reference, party-reference, or contextual-link mutations in the caller transaction and return deterministic mutation facts. | Artifact code MUST NOT persist generic relationships through an untyped callback or open map. |
| Revision append and conflict mechanics | `module.revisions` | Append change-set mutations/revisions and verify generic conflict-window mechanics in the caller transaction. | Workbook transport and Artifacts MUST NOT mint or parse client-editable token authority. |
| Projection refresh | `module.projections` with artifact provider intent | Refresh or load the exact affected artifact row and affected-view facts in the caller transaction. | Artifacts MUST NOT become the owner of generic query, cursor, or projection lifecycle behavior. |
| Idempotency | Existing owner-specific mutation capability | Distinguish first commit, exact replay, and same-key/different-content conflict before duplicate effects can survive. | No façade MAY implement a second incompatible idempotency namespace for the same operation. |
| Committed mutation facts | Source mutation transaction and `module.collaboration` publication boundary | Commit deterministic changed-field and affected-view intent; publish or replay only after commit. | Delivery failure MUST NOT rerun the source mutation. |

## 4. Public Contract and Behavior Freeze Map

### 4.1 Exact artifact surface matrix

The eight rows below are exhaustive for the current artifact workbook surface
registry. Tests and implementation MUST use `view_schema_id` and the authored
canonical source filter. They MUST NOT derive this registry from visible
labels, package names, SQL switches, or Core 02's six-row tagged-variant
registry.

| `view_schema_id` | Durable discriminator | Minimum valid inline-create signal | Required omitted defaults |
| --- | --- | --- | --- |
| `cartulary.view.notes.v1` | `artifact_type='note'` | Non-empty normalized `note.title` or `note.body` | Artifact type, actor, and timestamps are server-filled; tags and contextual links do not satisfy the minimum. |
| `cartulary.view.comm_log.v1` | `artifact_type='comm_log'` | Valid `comm_type` and non-empty `audience`, `channel_or_meeting`, and `summary` | Generate `comm_id`; timestamp defaults to commit time; reference collections default empty; `next_report_at` and `privilege_tag` default `null`. |
| `cartulary.view.handoff.v1` | `artifact_type='handoff'` | `incoming_owner_user_id` and non-empty `current_state_summary` | Generate `handoff_id`; timestamp defaults to commit time; outgoing owner defaults to actor; record/risk collections default empty; `next_checks` and `acknowledged_at` default `null`. |
| `cartulary.view.status_review.v1` | `artifact_type='status_review'` | Non-empty `current_state_summary` | Generate `status_review_id`; timestamp and review owner default to commit time and actor; reference collections default empty; risk summary and next-report time default `null`. |
| `cartulary.view.lesson.v1` | `artifact_type='lesson'` | Non-empty `lesson.summary` | Generate `lesson_id`; timestamp and owner default to commit time and actor; collections default empty; `closure_state='open'`. |
| `cartulary.view.findings.v1` | `artifact_type='finding'` | Non-empty `finding.statement` | `kind='finding'`; `state='open'`; owner defaults to actor; confidence and `closed_at` default `null`. |
| `cartulary.view.investigative_queries.v1` | `artifact_type='investigative_query'` | Non-empty `platform`, `purpose`, and `query_text` | Generate `query_id`; creator defaults to actor; creation time defaults to commit time. |
| `cartulary.view.forensic_keywords.v1` | `artifact_type='forensic_keyword'` | Non-empty `pattern` and `reason` | Generate `keyword_id`; `match_mode='literal'`; `case_sensitive=false`; creation time defaults to commit time. |

Core 01 remains the owner of exhaustive field, write-target, default,
collection, and conflict declarations. A mismatch between that owner and this
table is `BLOCKED: owner contradiction`; the implementation MUST NOT choose the
tracker value.

### 4.2 Import availability matrix

| Artifact surface | Current availability | Façade binding | Required behavior |
| --- | --- | --- | --- |
| Notes | enabled | `artifacts.note.import_create` | Imports MUST dispatch through the registered artifact façade. |
| Communications Log | enabled | `artifacts.import_create` | Imports MUST dispatch through the registered artifact façade. |
| Handoffs | enabled | `artifacts.import_create` | Imports MUST dispatch through the registered artifact façade. |
| Status Reviews | enabled | `artifacts.import_create` | Imports MUST dispatch through the registered artifact façade. |
| Lessons | enabled | `artifacts.import_create` | Imports MUST dispatch through the registered artifact façade. |
| Findings | reserved and unavailable | none | The generated registry MUST prevent dispatch. |
| Investigative Queries | reserved and unavailable | none | The generated registry MUST prevent dispatch. |
| Forensic Keywords | reserved and unavailable | none | The generated registry MUST prevent dispatch. |

Import availability MUST be read from
`contracts/imports/view-targets.v1.json` and its generated projection. The
existence of a surface, source implementation, or package branch MUST NOT imply
importability.

### 4.3 Reporting freeze

| Included artifact family | Required disposition |
| --- | --- |
| note | Preserve current typed fact output. |
| finding | Preserve current typed fact output. |
| comm log | Preserve current typed fact output. |
| handoff | Preserve current typed fact output. |
| status review | Preserve current typed fact output. |
| lesson | Preserve current typed fact output. |
| investigative query | MUST remain absent. |
| forensic keyword | MUST remain absent. |

No refactor slice authorizes a reporting expansion.

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Eight artifact-backed `view_schema_id` values and canonical artifact-type filters | Core 01/Core 02, authored view schemas, `module.artifacts` projection intent | `surfaces.go`, `projectionprovider/query_surfaces.go`, `contracts/view-schemas` | View-schema registry, projection, and workbook coordination tests | Direct table-driven artifact owner test covering all eight IDs, filters, defaults, and required fields | High | IDs and field keys must not drift. |
| `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` | `module.workbook` transport; source/provider owners supply rows | OpenAPI and workbook/projection composition | Workbook query, pagination, sorting, filtering, grouping, and row-wire tests | Preserve current tests; add direct provider characterization for every artifact surface | High | Row envelope, cells, collection values, cursor behavior, and row version are frozen. |
| `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` | `module.workbook` transport; `module.artifacts` source create | OpenAPI, workbook contribution catalog, artifact façade | Direct note persistence plus coordination-surface integration tests | Direct artifact owner create matrix for all eight surfaces, defaults, minimum signal, and failures | High | Rejected creates must not create source, revision, projection, idempotency, or events. |
| `PATCH /api/v1/records/{record_id}` | `module.workbook` transport; `module.artifacts` field/collection semantics | OpenAPI, Core 03, root façade | Workbook mutation and coordination tests | Direct scalar/collection patch matrix, different-field rebase, same-field rejection, and rollback-on-failure tests | High | Preserve changed-fields-only behavior and row-version high-water semantics. |
| `POST /api/v1/records/{record_id}/conflicts/{conflict_token}/resolve` | Core 03 transport contract; revisions mechanics; artifact source revalidation | Core 03 §3.3.4, OpenAPI, `workbook_conflict.go` | Workbook conflict route/integration tests | Artifact-specific stale-token, keep-saved, use-unsaved, merged-value, and rejection side-effect characterization | High | Preserve token opacity, error precedence, exact replay, and fresh-conflict behavior. |
| `POST /api/v1/records/{record_id}/linked-notes` | Workbook transport; artifact note source; links relationship owner | OpenAPI and `linkednotes.Facade` | Linked-note route/integration tests | One-transaction characterization for note, link, tags, revision, projection, idempotency, and rollback | High | Contextual source types are timeline event, host, identity, and evidence. |
| Generic record delete, restore, and rollback routes | `module.revisions`; artifact source provider | OpenAPI, revision provider contribution | Revision matrix and integration tests | Artifact-specific source snapshot, view resolution, scalar restore, and collection-boundary cases | High | Record envelope owns deletion state; artifact source hook remains no-op unless an owner changes it. |
| Import apply through `POST /api/v1/import-sessions/{import_session_id}/apply` | `module.imports` orchestration; artifact source adapter | OpenAPI, import target registry, import owner registry | Import registry/boundary tests | End-to-end artifact target apply for enabled surfaces, actor fields, projection row, revision, and idempotency | High | No direct artifact apply characterization was found. |
| Incident-bundle export/import | Incident portability subsystem; artifact source adapter | OpenAPI and source catalog | Incident-bundle route/integration tests compare source tables and rebuilt projections | Preserve existing tests; add source-port unit coverage only if the adapter shape changes | High | Freeze the five current artifact-owned NDJSON files and invariant behavior. |
| Reporting snapshot materialization | Adopted Reporting NLSpec; `module.reporting`; artifact typed provider | Reporting materializer and provider | Reporting boundary guard and integration tests | Preserve six-family output and support-ref behavior | Medium | Investigative-query and forensic-keyword omission is frozen, not approved for expansion. |
| Recovery state and projection rebuild | Recovery owner with artifact state contribution | Recovery catalog/fixtures and provider rebuild | Recovery catalog, coverage, and restore verification tests | Preserve current owner/table classification; add only if contribution shape changes | Medium | Authoritative source tables and disposable projection table must remain distinct. |
| Authorization and incident lifecycle admission | Core 04; route owners; artifact source revalidation | Core 04 §2 and artifact façades | Workbook route guards, closed-incident, and integration tests | Explicit create/patch/conflict/linked-note viewer, non-member, inactive-user, and closed-incident cases where absent | High | Coordination artifacts inherit incident ACLs; `deployment_admin` is not an incident-data bypass. |
| Revision/change-set and idempotency semantics | `module.revisions` plus source-owner mutation transaction | Artifact/linked-note façades and Core 02/03 | Workbook, linked-note, revisions, and collaboration tests | One-change-set and exact-replay/no-duplicate matrix across create, patch, linked note, and conflict resolution | High | One successful logical mutation must not fan out into duplicate durable effects. |
| Projection refresh and `artifact_grid_projection` storage | `module.projections` lifecycle; artifact provider intent | Provider contract, projection code, authored migration | Projection query/rebuild and incident-bundle tests | Transaction rollback and refresh parity across all artifact types | High | Projection remains derived and rebuildable. |
| `record_changed` WebSocket semantics | `module.collaboration` | Core 03 §4.3.1 and downstream mutation contribution contracts | Collaboration/workbook tests, including rejected-mutation no-event behavior | Preserve existing cross-owner tests; add artifact affected-view/field coverage if missing | High | Target code must not take over sequencing or publication. |
| Saved-view and view-schema interoperability | `module.savedviews` and view-schema owners | Core 03 and generic workbook composition | Saved-view and coordination-surface tests | Preserve additive saved-view behavior for all eight artifact schemas | Medium | Visible labels or tab positions are not identities. |
| Generated Go and TypeScript surfaces | Authored contracts and generators | Generated artifact policy and generated registries/types | Generation and JSON-shape checks | No new characterization; run drift checks after authored changes only | High | Never hand-edit generated roots. |
| Grid adapter and UI selector contracts | `packages/grid-adapter`, `packages/ui-contracts`, generic web shell | Frontend imports and generated view contracts | Frontend unit/import-boundary and browser suites | None for backend-only slices; browser coverage becomes required if public wire/UI behavior changes | Low | No direct grid-vendor coupling was found in the artifact target. |
| Harness/test accounting | Testing Harness NLSpec and `module.artifacts` verification inputs | Owner verification contract and test-family manifest | One active artifact integration row | Audit ownership and add behavior-aligned rows before relying on the owner slice as complete evidence | High | Harness maps evidence; they do not define product behavior or runtime architecture. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Direct artifact characterization is materially narrower than the eight-surface, multi-adapter responsibility. | Only one active `module.artifacts` harness row; direct tests cover note persistence, risk-ref grammar, and rollback mapping. | High | `must_fix` | `module.artifacts` tests, with collaborators where the postcondition crosses owners | Complete RB-001 before moving the façade or mutation sequence. |
| `workbook_facade.go` mixes source validation/persistence with incident admission, authn, idempotency, records, links, revisions, projections, and HTTP-like status outcomes. | Direct imports and the 1,100-line façade implementation | High | `should_fix` | `module.artifacts` façade over narrow owner ports | Decompose in behavior-preserving slices while retaining current public DTOs during migration. |
| Contextual linked-note creation duplicates artifact create, collection, revision, projection, and idempotency sequencing. | `linkednotes/facade.go` and root workbook façade implement parallel transaction flows. | High | `should_fix` | `module.artifacts` transactional mutation coordinator | Characterize atomicity, then make both façades call one private coordinator. |
| Surface IDs and artifact-type filter knowledge are duplicated between root and projection-provider code. | `surfaces.go` and `projectionprovider/query_surfaces.go` | Medium | `should_fix` | `module.artifacts` surface registry | Consolidate behind one owner API without changing query SQL or generated schema IDs. |
| Peer dependencies are concrete and broad where narrow ports would better expose ownership. | Root and linked-note façades import records, links, revisions, projections, incidents, and platform packages. | Medium | `should_fix` | Application composition supplies typed owner capabilities | Define transaction-compatible ports; do not import peer internals. |
| Artifact source persistence uses direct SQL. | `mutations.go`, handoff risk refs, rollback, portability, and provider code | Medium | `intentional/no_action` | `module.artifacts` | Keep source SQL owner-local; hide it from callers rather than moving it to platform or workbook. |
| Projection-provider SQL joins records, artifacts, links, and parties. | `projectionprovider/provider.go` and query surfaces | High | `intentional/no_action` | Artifact provider intent plus `module.projections` lifecycle | Preserve the adopted provider boundary and validate source-owner allowlists. |
| Import, portability, recovery, revision, and reporting contributions live beneath artifacts. | Composition registries call explicit artifact contribution constructors. | Medium | `intentional/no_action` | `module.artifacts` thin adapters | Keep these contributions thin; do not move source semantics into orchestrators. |
| Reporting currently emits six artifact families but not investigative-query or forensic-keyword facts. | `reportingprovider/provider.go` | Medium | `defer` | Reporting owner plus `module.artifacts` provider | Preserve current behavior. Any expansion requires separate owner evidence and authorization. |
| Artifact collection and link semantics meet in the mutation transaction. | Collection policy and linked-note/handoff code call link-owned capabilities. | High | `should_fix` | Artifact field semantics; `module.links` relationship persistence | Define narrow link mutation ports and preserve item-kind distinctions and transaction atomicity. |
| Authorization checks span transport and source revalidation. | Core 03/04 ordering plus workbook/artifact façade behavior | High | `intentional/no_action` | Transport/platform authorization and artifact source owner | Preserve request-time membership/role checks and source revalidation; do not centralize policy in a test helper. |
| Generated Go/TypeScript projections describe artifact surfaces and imports. | Generated artifact policy and generated registries/types | High | `intentional/no_action` | Authored contracts and generators | Change authored owners first when authorized, then run generators; never hand-edit generated files. |
| Test-only assumptions do not appear to control production artifact behavior. | Production code does not import test utilities; harness selection is external. | Low | `intentional/no_action` | Existing module/harness boundaries | Retain the separation and add a boundary guard only if a future move creates a new risk. |
| No direct grid-vendor import exists outside the grid adapter in the inspected artifact flow. | `react-data-grid` imports remain under `packages/grid-adapter`; web uses its adapter. | Low | `intentional/no_action` | `packages/grid-adapter` | No artifact refactor action. |
| The framework's default module catalog omits artifacts, while current Core 02 and repository contracts assign artifact source ownership. | Framework catalog compared with Core 02 and live composition/contracts | Medium | `intentional/no_action` | Adopted owners and current repository | Record the mismatch; do not use the framework omission to dissolve the source owner. |

### 5.1 Postcondition ownership map

Harness ownership MUST follow the normative postcondition or primary durable
mutation. Package location, test file, runner, and maintainer identity MUST NOT
decide ownership.

| Postcondition | Catalog owner | Required collaborators or separate owner evidence |
| --- | --- | --- |
| Eight artifact surface IDs, discriminators, source policies, defaults, and direct field mapping | `module.artifacts` | `module.projections`, `module.revisions`, `module.workbook` |
| Artifact source create/patch and owner transaction result | `module.artifacts` | `module.collaboration`, `module.projections`, `module.records`, `module.revisions`, `module.workbook` |
| Artifact collection field policy and source validation | `module.artifacts` | `module.links` owns generic relationship persistence; publication/projection/history remain collaborator evidence. |
| Contextual linked-note atomic creation | `module.artifacts` | `module.collaboration`, `module.links`, `module.projections`, `module.records`, `module.revisions`, `module.workbook` |
| Artifact import create façade | `module.artifacts` | `module.collaboration`, `module.imports`, `module.projections`, `module.records`, `module.revisions` |
| Source-field and conflict-window revalidation during accepted artifact resolution | `module.artifacts` | `module.workbook` owns route guards; `module.revisions` owns token/history mechanics; projection/publication remain collaborators. |
| Risk-reference grammar and parent-scoped identity | `module.artifacts` | none |
| Artifact rollback source mapping | `module.artifacts` | `module.revisions` owns rollback coordination. |
| Generic query, sort, filter, group, and cursor behavior | `module.workbook` and `module.projections` | Artifacts supplies provider intent. |
| `record_changed` publication, sequencing, delivery, and replay | `module.collaboration` | Artifacts supplies committed mutation facts. |
| Delete, restore, and rollback HTTP coordination | `module.revisions` | Artifacts supplies source adapters. |
| Incident-bundle orchestration and atomic publication | `module.incidentbundles` | Artifacts supplies a source contribution. |
| Recovery orchestration and readiness | `module.recovery` | Artifacts declares authoritative and derived state. |
| Report materialization and release output | `module.reporting` | Artifacts supplies the frozen six-family fact set. |

### 5.2 Owner-document and artifact disposition

| Location | Required action |
| --- | --- |
| Core 00 | No edit. Current precedence and primary-owner rules are sufficient. |
| Core 01 | No edit to close current blockers. If S-00 finds a missing public, view, import, or provider rule, the affected work MUST stop for owner adoption. |
| Core 02 | No edit. Current artifact source, relationship, child-risk, history, and import-provenance semantics govern. |
| Core 03 | No edit. Current rebase, conflict, collection payload, token, idempotency, and rejection rules govern. |
| Core 04 | No edit. Current authentication, CSRF, hidden-resource, role, and fail-closed rules govern. |
| Core 05 | Not applicable because no timed or fixture-sensitive publication claim is made. |
| Testing Harness NLSpec | No edit. Current owner precedence, immutable IDs, selectors, evidence gates, and temporary crosswalk rules govern. |
| `contracts/verification/owners/module.artifacts.json` | A later authorized RB-002 task MUST replace the two coarse active entries with the seven entries in Section 11 after transitional evidence succeeds. |
| `tools/test_families/module.artifacts.json` | A later authorized RB-002 task MUST perform the ten-row transitional adoption and nine-row final cutover in Section 11. |
| Generated catalog/topology | Regenerate through `make generate`; MUST NOT be hand-edited. |
| This tracker | Record scenario authority, mutation effects, immutable-ID cutover, result roots, and status transitions. |
| `temp/analysis-notes.md` and research material | Supporting evidence only; no edit and no product authority. |

## 6. Refactor Workstreams

### 6.1 Required gate ordering

| Gate | Required result | Work unblocked |
| --- | --- | --- |
| RB-001 | Complete S-00 characterization passes with no production-code diff. | S-01 source decomposition and S-02 coordinator unification MAY begin. |
| RB-002 | Final verification contract, nine-row owner manifest, generated accounting, exact selectors, and owner evidence audit pass. | S-06 accounting cleanup and S-07 final refactor completion MAY close. |
| Reporting freeze | The exact six included families and two excluded families remain unchanged. | S-04 reporting-adapter thinning MAY proceed without behavior authorization. |

A structural refactor MUST NOT begin before RB-001 closes. The overall refactor
MUST NOT be declared complete before RB-002 closes.

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `WF-00` | Session/source bootstrap and tracker initialization | root | none | `WF-01` | Fix target, authority, constraints, commit, and tracker state. | Tracker only during planning | `git status --short`; framework and `AGENTS.md` inspection | Scope, source hierarchy, and no-code constraint recorded. |
| `WF-01` | Target inventory | chain | `WF-00` | `WF-02` | Inventory every target file, caller, dependency, test, and contract touchpoint. | All 25 target files and named composition callers | Read-only `find`, `rg`, `sed`, `wc` | Section 2 has one row per file. |
| `WF-02` | Contract-owner mapping | chain | `WF-01` | `WF-03`, `WF-04` | Freeze artifact source, transport, revision, projection, collaboration, reporting, and authorization ownership. | Owner docs, OpenAPI, authored contracts, composition | Human owner review; `make explain-test-owner OWNER=module.artifacts` for evidence routing only | Sections 3 and 4 identify owner and test posture for every risk. |
| `WF-03` | Characterization test gap analysis | parallel | `WF-02` | `WF-05`, `WF-07` | Identify missing pre-move evidence without changing behavior. | Artifact tests plus workbook, projection, revision, import, portability, reporting, recovery, collaboration tests | `make task-guide ROLE=module-author OWNER=module.artifacts`; planned owner slices | RB-001 is either closed with tests or remains an implementation blocker. |
| `WF-04` | Boundary/coupling scan | parallel | `WF-02` | `WF-05` | Classify mixed responsibilities, concrete peer coupling, generated risk, and intentional adapters. | Root façade, linked notes, provider subpackages, callers | `make backend-module-boundary-check` during implementation | Every finding has classification, owner, and action. |
| `WF-05` | Façade and ownership redesign plan | chain | `WF-03`, `WF-04` | `WF-06` | Define an artifact-owned source service and narrow transaction-compatible peer ports without changing contracts. | Root artifact package, linked notes, application composition | Characterization gate, then artifact owner slices | Public compatibility seam and private decomposition are reviewable. |
| `WF-06` | Slice sequencing plan | chain | `WF-05` | `WF-07`, `WF-08` | Sequence characterization, private extraction, coordinator reuse, provider cleanup, composition migration, and cleanup. | Target package plus explicitly named composition files | Per-slice commands in Section 7 | Each slice has rollback and binary completion criteria. |
| `WF-07` | Harness/test/accounting update plan | chain | `WF-03`, `WF-06` | `WF-08` | Align owner rows and collaborator evidence with actual postconditions without making harness maps architectural. | Authored verification/test-family inputs only if later authorized | Owner task guide, owner slices, JSON/generation checks | RB-002 closed or explicitly retained as blocker. |
| `WF-08` | Validation and final handoff | chain | `WF-06`, `WF-07` | none | Run narrow-to-broad proof, generated/boundary checks, finalizer, and update handoff. | Changed implementation/test/contract inputs from authorized slices; tracker | Section 8 commands | Passing evidence and residual risks are recorded without overclaiming. |

## 7. Proposed Refactor Slice Plan

All slices below require a later implementation authorization. They preserve
observable behavior. Any correction or expansion discovered during a slice is a
separate change marked `requires later authorization`.

### 7.1 S-00 test interface

S-00 MUST add these exact top-level test symbols:

| Package | Required symbol | Owner postcondition |
| --- | --- | --- |
| `./internal/modules/artifacts` | `TestArtifactSurfaceContractMatrix` | Root surface registry, canonical discriminators, and revision contribution |
| `./internal/modules/artifacts/projectionprovider` | `TestArtifactProjectionProviderSurfaceContractMatrix` | Projection-provider descriptors and canonical filters |
| `./internal/modules/artifacts` | `TestArtifactWorkbookMutationContractMatrix` | Eight-surface source create/patch behavior and durable effects |
| `./internal/modules/artifacts` | `TestArtifactCollectionMutationContractMatrix` | Artifact collection policy and source validation |
| `./internal/modules/artifacts/linkednotes` | `TestArtifactLinkedNoteAtomicity` | Four-context linked-note primary mutation |
| `./internal/modules/artifacts` | `TestArtifactImportCreateFacadeContract` | Artifact import owner façade |
| `./internal/modules/artifacts` | `TestArtifactConflictSourceRevalidation` | Accepted-resolution source and conflict-window revalidation |

S-00 MUST also route the existing
`TestRiskRefItemRefIsCanonical`,
`TestRiskRefItemRefParsingIsStrict`,
`TestSourceForRollbackValueMapsAllArtifactVariantsWithoutCollections`, and
`TestValidSourceRejectsSubtypeInvariantViolations` symbols into artifact owner
accounting. A catalog selector MUST name exact top-level symbols. It MUST NOT
use a glob, regular expression, package-wide selector, or symbol already
selected by another active row.

### 7.2 Surface, create, and patch characterization

`TestArtifactSurfaceContractMatrix` and
`TestArtifactProjectionProviderSurfaceContractMatrix` MUST prove:

1. Section 4.1 contains exactly the recognized artifact surfaces and
   discriminators.
2. Unknown and non-artifact surfaces are rejected.
3. Root surface mapping and `projectionprovider.QuerySurfaces` agree.
4. Query-provider filters use the exact canonical discriminator.
5. Revision provider contribution registers all eight views exactly once.
6. Provider intent remains artifact-owned while projection storage and generic
   query orchestration remain projection/workbook-owned.
7. The six tagged artifact variants are not treated as the exhaustive
   eight-surface registry.

`TestArtifactWorkbookMutationContractMatrix` MUST run table-driven cases over
all eight rows in Section 4.1. For each surface it MUST cover:

- minimum valid create, every omitted default, normalized-empty input, and
  below-minimum rejection;
- unknown, read-only, and create-only field rejection;
- persisted discriminator, subtype state, actor, timestamp, and mutation-source
  attribution;
- the full canonical `view_row_v1`, one logical change set, corresponding
  revisions, and projection parity;
- first commit, exact replay, and same-key/different-normalized-content
  `client_txn_conflict`;
- a material scalar patch, permitted clear, non-clearable `null` rejection,
  different-field stale-version rebase, same-field conflict, and
  changed-fields-only behavior.

The patch matrix MUST cover every artifact-owned declared writable field. It
MAY use table-driven subtests rather than one catalog row per field.

### 7.3 Collection characterization

Artifact field policy and generic relationship persistence are separate
postconditions. The collection matrix MUST characterize every field below and
MUST exercise one service-backed case for each distinct family.

| Family | Fields or representative routing | Required semantics |
| --- | --- | --- |
| `record_tag` | `note.tags` | Add; normalized duplicate coalescing; remove by stable `item_ref`; invalid operation rejection; no raw-array replacement |
| `record_ref` | Comm-log decisions/tasks; handoff decisions/tasks; status-review decisions/tasks/evidence; lesson tasks/evidence; finding support/contradiction refs | Enforce declared target type when one exists; enforce incident and active-record boundaries; preserve declared link type |
| `party_ref` | `comm_log.audience_party_ids`, `comm_log.attendee_party_ids` | Target MUST be an active party in the incident; removing a party ref MUST NOT clear source-preserving `comm_log.audience` text |
| `risk_ref` | `handoff.open_risk_refs` | Duplicate normalization is parent-scoped; child ID is server-generated and stable; foreign/malformed `item_ref` is rejected |

For each family, the matrix MUST cover valid add, duplicate add, valid remove,
wrong-incident target, wrong record type, soft-deleted target, malformed or
foreign `item_ref`, raw array, raw `null`, exact replay,
same-key/different-content rejection, and rollback when a later transaction
participant fails.

For a `collection_review` conflict, the payload MUST retain distinct
`collection_value_v1` item kinds and base/server/client values. A merged
resolution MUST apply `collection_actions_v1` against the current server
collection; it MUST NOT replace the collection with an untyped array.

### 7.4 Linked-note atomicity

`TestArtifactLinkedNoteAtomicity` MUST cover exactly these contextual source
record families: timeline event, host, identity, and evidence.

For each family it MUST prove:

1. The new note has the same artifact/subtype shape as Notes-surface creation.
2. The contextual link is a generic directed `record_links` association from
   the source record to the new artifact.
3. A preseeded contextual link does not satisfy the note minimum signal.
4. Optional tags, contextual link, source record, history, projection,
   idempotency, and committed mutation facts are one logical operation.
5. First success leaves exactly one note and one contextual link.
6. Exact replay returns the original identifiers and creates no additional
   source or downstream effect.
7. Same-key/different-note-content reuse fails before mutation.
8. Failure at link, tag, revision, or projection participation rolls back all
   earlier writes.
9. Failure after source commit but before delivery leaves committed source
   state and recovers delivery from committed facts without rerunning the
   mutation.

### 7.5 Import façade characterization

`TestArtifactImportCreateFacadeContract` MUST use Section 4.2 rather than
assuming every artifact surface is importable. For every enabled target it
MUST prove:

- Imports resolves exactly one target-bound artifact façade.
- The façade uses the caller's transaction and does not begin or finish one.
- Interactive source validation, defaults, writeability, and subtype
  persistence apply.
- Required import provenance is persisted.
- The expected record, artifact, subtype, mutation, revision, projection,
  collaboration intent, and owner result are produced.
- The response contains canonical `record_id`, row version,
  change-set-mutation reference, created/reused result, owner result code, and
  full row refresh.
- Owner rejection leaves no artifact-owned durable effect.
- A later caller-transaction failure rolls back every artifact-owned effect.
- Replay or post-commit recovery does not create another source record.

For every reserved target, the test MUST prove that registry admission prevents
dispatch. Generic import-session lifecycle, unit outcomes, cancellation,
dispatcher finalization, and crash recovery remain `module.imports`
postconditions.

### 7.6 Conflict source revalidation

`TestArtifactConflictSourceRevalidation` MUST exercise:

| Conflict class | Required representative field |
| --- | --- |
| `text_compare_merge` | `note.body` |
| `atomic_replace` | `finding.state` or `comm_log.comm_type` |
| `collection_review` | `note.tags` or `handoff.open_risk_refs` |

For each applicable class the test MUST cover `keep_saved`, `use_unsaved`,
valid `merged_value`, fresh token, stale token returning a fresh conflict,
current incident/record/view/field/row-version revalidation, exact replay,
changed replay returning `client_txn_conflict`, and same-token/new-key
revalidation.

The complete HTTP guard-precedence matrix remains Workbook/Revisions-owned.
Artifact tests MUST NOT duplicate that route ownership merely to increase
artifact row count.

### 7.7 Required mutation-effect ledger

Every S-00 service-backed test MUST reconcile the applicable effects below.

| Outcome | Required effects | Forbidden effects |
| --- | --- | --- |
| First successful create or material patch | Exact source mutation; one logical change set; correct revisions; one resulting projection state; one committed idempotency result; deterministic changed-field and affected-view facts | Duplicate record, subtype, link, child row, revision, projection, idempotency success, or event intent |
| Exact committed replay | Original result and identifiers | Any new source, history, projection, idempotency, or collaboration effect |
| Same key with different normalized request | `client_txn_conflict` | Any source or downstream mutation |
| Validation, authorization, hidden-resource, or malformed-token rejection | Applicable owner-defined stable error | Source row, row-version change, change set, revision, projection update, successful idempotency result, or publication |
| Stale conflict token | Fresh same-field conflict payload | Source, history, projection, idempotency, or publication mutation |
| Failure before commit | No surviving participant effect | Partial record, artifact, subtype, link, tag, risk child row, revision, projection, or idempotency success |
| Source commit followed by publication/delivery failure | Committed source remains; delivery is retryable from committed facts | Source mutation re-execution |
| Import owner validation failure | Imports MAY record its owner-defined unit failure or diagnostic. | Artifact source, history, projection, or publication effect |

### 7.8 Fault-injection and retained-output boundary

S-00 MUST select the first feasible fault mechanism in this order:

1. Existing injected port or fake.
2. Test transaction wrapper returning a controlled error.
3. Test-local PostgreSQL trigger or constraint removed by test cleanup.
4. Existing harness-owned guarded control that already satisfies Core 04.

S-00 MUST NOT add a public `/api/v1/test/*` route, an
environment-controlled production bypass, or a package-global failure switch.
Retained output MUST use synthetic fixture content and MUST NOT disclose note
bodies, investigative queries, conflict tokens, session values, idempotency
keys, or raw incident data.

### 7.9 Binary S-00 exit

RB-001 becomes `DONE` only when:

- all eight surfaces pass create/default/rejection coverage;
- every writable field is present in the patch-policy matrix;
- all four collection families and all four linked-note contexts pass;
- import coverage matches the exact current registry;
- all three conflict classes, exact replay, and changed replay pass;
- rollback is proven across source, relationship, history, projection, and
  idempotency boundaries;
- applicable cross-owner rejection and publication tests pass;
- S-00 contains no production-code diff;
- every expectation traces to an adopted owner or explicitly frozen current
  behavior; and
- no expectation expands reporting, importability, route behavior, or surface
  availability.

Only then MAY ART-005 become `DONE` and S-01/S-02 begin.

### 7.10 Behavior-preserving slice sequence

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `S-00` | none | Implement Sections 7.1 through 7.9 without a production-code change. | Artifact `_test.go` files plus existing cross-owner test packages; authored verification/family inputs during RB-002 transition | A test expectation could invent behavior or duplicate another owner's postcondition. | Exact symbols, matrices, and ledger above | `make test-slice OWNER=module.artifacts`; `make service-backed-test-slice OWNER=module.artifacts`; prescribed collaborator slices | Revert characterization and catalog-adoption commits if an expectation lacks owner authority; do not alter production to make characterization pass. | Every Section 7.9 condition passes and retained evidence is recorded. |
| `S-01` | `S-00` | Decompose source policy, validation, subtype defaults, and persistence into focused unexported components inside `module.artifacts`; retain current exported façades and SQL. | Root artifact `.go` files; no migrations | Defaults, field keys, enum/range validation, SQL column mapping, and timestamps | Preserve S-00 and direct Store behavior tests | Artifact owner slices; `make backend-module-boundary-check` | Revert the slice as one unit; no caller migration depends on it yet. | Public types/results and persisted rows are unchanged; source components have one semantic responsibility. |
| `S-02` | `S-01` | Introduce one private artifact-owned transactional mutation coordinator reused by workbook create/patch and contextual linked-note creation through narrow peer ports. | `workbook_facade.go`, `workbook_conflict.go`, `linkednotes`, private artifact coordinator, application-injected ports | Transaction atomicity, idempotency, one change set, revision count, projection refresh, contextual link, and event intent | S-00 plus exact replay, rollback, and no-duplicate effect cases | Artifact service-backed slice and prescribed workbook/collaboration slices | Keep the old paths until both callers pass characterization; revert caller-by-caller adapters if needed. | Both façades use one source mutation sequence and all frozen effects remain byte/semantically equivalent. |
| `S-03` | `S-01` | Consolidate eight surface IDs and canonical artifact-type filter lookup behind one artifact-owned registry. | `surfaces.go`, `projectionprovider/query_surfaces.go`, projection composition/tests | View IDs, filter values, query SQL, fields, collection expressions, and affected views | Eight-surface registry/provider parity tests and existing projection/workbook tests | Artifact and projections owner slices; `make backend-module-boundary-check` | Restore the former duplicated lookup while retaining new tests. | One owner source supplies IDs/filter intent; projection rows and contracts do not change. |
| `S-04` | `S-02`, `S-03` | Reduce import, portability, recovery, reporting, rollback, and delete/restore integrations to thin adapters over the stable artifact source boundary. | Root adapter files and provider subpackages; affected composition registries only as needed | Import transaction, five portability files, recovery classification, six reporting families, delete/restore/rollback behavior | Preserve/import direct tests plus incident-bundle, recovery, reporting, and revisions integration tests | Artifact owner slice plus prescribed affected-owner slices | Revert one adapter at a time; retain compatibility constructors until composition migrates. | Adapters contain translation/contribution logic only and preserve current inputs and outputs. |
| `S-05` | `S-02`, `S-04` | Migrate application composition and workbook contribution adapters to the narrower artifact ports. | Import, portability, projection, recovery, revision, reporting, and workbook assembly seams | Constructor wiring, route dispatch, error mapping, auth ordering, transaction sharing | Existing assembly, route, integration, and boundary tests | Affected owner slices; `make test-fast`; `make backend-module-boundary-check` | Keep compatibility constructors for one slice and revert composition wiring if any route diverges. | All callers use the intended façade/ports; HTTP DTOs and generated contracts are unchanged. |
| `S-06` | `S-05` | Remove obsolete duplicate helpers, add owner/import guardrails, and update authored verification/catalog inputs only when movement changes actual paths or accounting. | Artifact package; boundary tests; authored contracts/tools inputs only if justified | Accidental generated edits, stale selectors, missing owner rows, or phase-shaped accounting | Boundary guard and exact owner-row selector tests | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make backend-module-boundary-check` | Revert cleanup separately from required authored input changes; regenerate rather than restoring generated files manually. | No obsolete path remains, generated output is clean, and RB-002 is closed or explicitly blocks completion. |
| `S-07` | `S-06` | Run narrow-to-broad verification, review the complete diff, and update this handoff. | Verification only plus tracker handoff | False confidence from incomplete owner rows or unrelated broad-suite failures | All retained characterization and collaborator suites | `make agent-finalize`, then `make test-fast`, then risk-based `make check` | Revert the most recent failing slice, not the baseline characterization. | Required targets pass or failures are attributed with result roots; handoff lists remaining risk. |
| `S-08` | Separate later authority | Make any behavior correction or expand reporting/import/surface availability found during the refactor. | Owner docs first, then authored contracts, generated outputs, implementation, tests | Intentional public behavior change | Owner-defined characterization and migration cases | Owner-specific commands discovered at that time | Roll back the behavior slice independently of the structural refactor. | `requires later authorization`; adopted owner and migration impact are explicit before edits. |

No slice authorizes a database migration, route change, frontend redesign, or
manual generated-file edit.

## 8. Validation Plan

The commands below were discovered from the current Make-owned task surface.
They are a future implementation validation plan, not evidence that the
refactor has already passed.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make test-slice OWNER=module.artifacts` | Full active owner partition | yes | The current sole row is evidence-class `integration`; RB-002 requires an ownership audit before treating this as complete artifact coverage. |
| integration | `make service-backed-test-slice OWNER=module.artifacts` | PostgreSQL-backed active artifact owner rows | yes | Add characterization first, then run the full owner partition. |
| collaborator integration | `make task-guide ROLE=module-author OWNER=<affected-owner>`, followed by the prescribed slice | Workbook, projections, revisions, imports, incident bundles, reporting, recovery, and collaboration as affected | yes | Discover the exact owner target/rows before each slice; do not guess selector names. |
| e2e/browser | `make browser-e2e` | Public workbook/browser behavior | no | Not required for the proposed backend-only structural slices. It becomes required if later authorization changes public wire or frontend behavior. |
| generated drift | `make generate-drift` | Authored inputs versus generated outputs | no | Required before completion if authored contract or generator inputs change; useful final anti-drift evidence. |
| generated policy | `make generated-artifact-policy-check` | Generated roots and hand-edit policy | no | Required for any slice near generated projections. |
| JSON contract shape | `make json-shape-check` | Authored and generated JSON shapes | no | Required if authored contract or harness JSON changes. |
| test catalog | `make test-catalog-check` | Verification contracts, owner manifests, exact selectors, and semantic identities | no | Required during RB-002 transition and after final retirement. |
| import-boundary/static | `make backend-module-boundary-check` | Backend dependency directions | yes | Run after each boundary-moving slice. |
| documentation | `make lint-markdown` | This tracker and repository Markdown | no | Required for this planning-only tracker change. |
| fast aggregate | `make test-fast` | Broad fast repository verification | no | Run after narrow owner/collaborator slices pass. |
| finalizer | `make agent-finalize` | End-of-run evidence and retained-run maintenance | no | Run before broader end-of-run verification; record when `RESULTS_DIR` is unset. |
| full check | `make check` | Default broad local correctness gate | no | Risk-based final implementation gate; report result root and attribution for failures. |

Command discovery performed during this planning session:

- `make task-guide ROLE=module-author OWNER=module.artifacts` identified the
  focused and service-backed owner slices and `make test-fast`;
- `make explain-test-owner OWNER=module.artifacts` reported one active,
  service-backed row; and
- filtered `make help-all` output confirmed the validation targets above.

These were discovery commands only. No product test, generated-drift, browser,
or broad validation target was run during planning.

### 8.1 Required later-execution order

An authorized RB-001/RB-002 execution MUST run these gates in order:

1. `make task-guide ROLE=module-author OWNER=module.artifacts`
2. `make explain-test-owner OWNER=module.artifacts`
3. `make json-shape-check`
4. `make test-catalog-check`
5. `make generated-artifact-policy-check`
6. `make generate`
7. `make generate-drift`
8. `make test-slice OWNER=module.artifacts`
9. `make service-backed-test-slice OWNER=module.artifacts`
10. Each exact collaborator slice prescribed by that owner's current task guide
11. `make backend-module-boundary-check`
12. `make lint-markdown`
13. `make agent-finalize`
14. `make test-fast`
15. `make check`

`make generate` is authorized only in the later task that changes authored
verification/catalog inputs. Generated outputs MUST NOT be edited directly.
The final handoff MUST record every result root and distinguish product
failures from catalog, harness, configuration, or infrastructure failures.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| `ART-001` | Establish planning-only scope and source hierarchy | `WF-00` | DONE | none | Section 1 | Target, constraints, authority, and later-authorization rule are explicit. |
| `ART-002` | Inventory all 25 target files and inbound seams | `WF-01` | DONE | `ART-001` | Section 2 | Every target file has a responsibility, caller, dependency, test, owner, and risk row. |
| `ART-003` | Diagnose the artifact module boundary | `WF-02` | DONE | `ART-002` | Section 3 | Legitimate source ownership is separated from mixed façade/adaptor responsibilities. |
| `ART-004` | Freeze public and storage behavior | `WF-02` | DONE | `ART-002` | Section 4 | Every discovered contract has an owner and test posture. |
| `ART-005` | Close direct characterization gaps | `WF-03` | DONE | `ART-004` | RB-001; Sections 7.1 through 7.9 | Every binary S-00 exit condition passes with no production-code diff. |
| `ART-006` | Define coupling classifications and guardrails | `WF-04` | DONE | `ART-003`, `ART-004` | Section 5 | Every inspected coupling is classified with an owner and action. |
| `ART-007` | Decompose source internals behind the current façade | `WF-05` | DONE | `ART-005`, `ART-006` | S-01 | Focused private components preserve public façade and SQL behavior. |
| `ART-008` | Unify artifact and contextual-note mutation coordination | `WF-05` | DONE | `ART-007` | S-02 | Both public flows share one characterized transactional coordinator. |
| `ART-009` | Consolidate artifact surface/provider intent | `WF-06` | DONE | `ART-007` | S-03 | One owner registry supplies all eight IDs/filter mappings without projection drift. |
| `ART-010` | Thin source-owner adapters and migrate composition | `WF-06` | DONE | `ART-008`, `ART-009` | S-04, S-05 | Callers use narrow ports and all routes/envelopes remain unchanged. |
| `ART-011` | Audit harness ownership and evidence accounting | `WF-07` | DONE | `ART-005`, `ART-010` | RB-002; Section 11.2; S-06 | Final accounting is exactly 9 rows, 4 unit, 5 integration, and 7 verifications with complete terminal evidence. |
| `ART-012` | Run implementation validation and final handoff | `WF-08` | DONE | `ART-010`, `ART-011` | S-07; Section 8 | Narrow and broad evidence is recorded with attributable results. |
| `ART-013` | Create and make the planning tracker decision-complete | `WF-08` | DONE | `ART-001` through `ART-006` | This file | The tracker defines exact behavior mappings, interfaces, blocker closure, catalog cutover, and acceptance criteria; no implementation is claimed. |
| `ART-014` | Separate any behavior correction from structural work | later authorization | DEFERRED | owner evidence | S-08 | No correction proceeds without adopted authority and an explicit task. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30 08:33 EDT | Codex planning session on `main` at `ae4854b3079ba8d9166e76412b1f7c777f1b3ab5` | Planning tracker created; implementation is not authorized. | Inspected `AGENTS.md`, framework, domain, relevant Core/Testing Harness/Reporting owner sections; touched this tracker only. | `sed`, `rg`, `git branch --show-current`, `git rev-parse`, `git status --short`, `date` | Authority order applied; no owner contradiction found; Core 05 is not applicable. | None for tracker completion | Begin S-00 only under a later implementation task. |
| 2026-07-30 09:15 EDT | Codex NLSpec-style tracker revision on `main` at `ae4854b3079ba8d9166e76412b1f7c777f1b3ab5` | Planning gaps are decision-complete; no blocker execution is claimed. | Inspected NLSpec guidance, closure analysis, current tracker, and targeted live harness/contracts; touched this tracker only. | `wc`, `sed`, `rg`, `jq`, `git`, `make author-test-row-id`, `make lint-markdown` | Normative scope, owner subordination, exact interfaces, defaults, mappings, and binary criteria are recorded; Markdown validation passed. | RB-001 and RB-002 remain execution blockers. | An authorized later task implements S-00 and the transitional catalog cutover. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30 08:33 EDT | Codex planning session | All 25 artifact files and named backend callers inventoried. | Inspected `internal/modules/artifacts/**`, application assemblies, and direct module consumers; touched this tracker only. | `find`, `rg`, `sed`, `wc` | Artifacts is a legitimate source owner and a mixed-responsibility façade/adapter package, not an unrelated catch-all. | RB-001 | Add characterization before façade decomposition. |
| 2026-07-30 09:15 EDT | Codex NLSpec-style tracker revision | Transaction ownership and six narrow capability boundaries are fixed without prescribing private method names. | Reinspected collection policy, validation, import façade/finalizer, catalog schemas, and caller boundaries; touched this tracker only. | `sed`, `rg`, `jq` | Workbook/linked-note transaction ownership, caller-owned import/revision transactions, and post-commit publication boundaries are unambiguous. | RB-001 | Implement only the test-only S-00 package before structural movement. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30 08:33 EDT | Codex planning session | No frontend implementation change is planned. | Inspected representative `apps/web`, view-contract, UI-contract, and grid-adapter consumers; touched this tracker only. | `rg`, `sed` | Frontend consumes generic generated view contracts; direct grid-vendor imports remain isolated in `packages/grid-adapter`. | None for backend-only plan | Require browser validation only if later scope changes public wire/UI behavior. |
| 2026-07-30 09:15 EDT | Codex NLSpec-style tracker revision | Frontend remains out of implementation scope. | Touched this tracker only. | Tracker comparison | Stable schema IDs, field keys, selectors, and grid-adapter isolation are explicit frozen contracts. | None for documentation revision | Do not add browser work unless later authorization changes public wire/UI behavior. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30 08:33 EDT | Codex planning session | Eight view schemas and import/projection/portability/recovery/OpenAPI/generated relationships mapped. | Inspected authored contract JSON/YAML, generated policy, and representative generated outputs; touched this tracker only. | `rg`, `sed`, `jq` | Generated roots are downstream projections and are not edit targets. Two exploratory `jq` queries used incorrect guessed shapes, returned errors, and were replaced by exact JSON reads. | None unless a future slice changes authored inputs | Run Make-owned generation and drift targets after any authorized authored-input change. |
| 2026-07-30 09:15 EDT | Codex NLSpec-style tracker revision | Final seven-verification and nine-row target states plus transitional retirement are specified; no catalog file changed. | Inspected current artifact verification/family inputs, schemas, owner registry, row-ID authoring tool, import registry, and Incident Bundles collaborator row; touched this tracker only. | `jq`, `rg`, `sed`, nine `make author-test-row-id` calls | Exact hashed row IDs, selectors, profiles, defaults, collaborators, and final 9/4/5/7 totals are fixed. | RB-002 | Later task adds authored inputs, generates outputs, retains reconciliation, and removes the temporary crosswalk. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30 08:33 EDT | Codex planning session | Direct and cross-owner tests mapped; owner accounting remains under-complete. | Inspected artifact tests, collaborator tests, `contracts/verification/owners/module.artifacts.json`, and `tools/test_families/module.artifacts.json`; touched this tracker only. | `make task-guide ROLE=module-author OWNER=module.artifacts`; `make explain-test-owner OWNER=module.artifacts`; filtered `make help-all`; `make lint-markdown` | Discovery commands passed. Markdown lint passed with run root `.cartulary/test-results/20260730T123834Z-p3373823`. One active service-backed artifact row was reported. No product validation suite ran. | RB-001, RB-002 | Add owner-backed characterization, then audit row ownership under Testing Harness rules. |
| 2026-07-30 09:15 EDT | Codex NLSpec-style tracker revision | S-00 now has exact test symbols, matrices, effect ledger, fault policy, and binary exit criteria. | Inspected current artifact tests and harness schemas; touched this tracker only. | `rg`, `sed`, `jq`, `make author-test-row-id`, `make lint-markdown` | Documentation validation passed; no product test, catalog check, generation, owner slice, or broad suite ran. | RB-001, RB-002 | Execute the exact Section 8.1 gate sequence only after later authorization. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30 08:33 EDT | Codex planning session | Incident role, hidden-resource, conflict-token, source revalidation, and post-commit publication boundaries are frozen. | Inspected relevant Core 03/Core 04 sections, OpenAPI, artifact/workbook conflict code, and route tests; touched this tracker only. | `rg`, `sed` | Authorization stays request-time at transport/platform boundaries; artifacts retains source-state revalidation. | RB-001 for missing direct negative cases | Include authorization precedence and no-side-effect failures in S-00. |
| 2026-07-30 09:15 EDT | Codex NLSpec-style tracker revision | Security and fault-injection boundaries are normative and testable. | Touched this tracker only; relied on previously inspected Core 03/Core 04 and current harness boundaries. | Tracker comparison | Guard ownership remains Workbook/Revisions; artifact source rejection is side-effect-free; no public or production-enabled test failpoint is permitted. | RB-001 | Prove source revalidation and effect absence without duplicating the route-owned guard matrix. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30 08:33 EDT | Codex planning session | Refactor slices and rollback boundaries are decision-complete; no implementation performed. | Touched this tracker only. | Read-only discovery commands; `make lint-markdown` | Documentation validation passed; reporting omission is preserved, not treated as an implied defect. | RB-001, RB-002 | Later authorized session starts with S-00 and uses `make task-guide ROLE=module-author OWNER=module.artifacts`. |
| 2026-07-30 09:15 EDT | Codex NLSpec-style tracker revision | Planning completeness is closed; execution remains gated. | Touched this tracker only. | Read-only discovery and documentation validation | Reporting remains six-family; RB-001 gates structural start; RB-002 gates final completion. | RB-001, RB-002 | Start with S-00 tests and the ten-row transitional catalog state; do not mark either blocker done from prose alone. |

## 11. Open Questions and Blockers

The product design is not open. The execution blockers concerned executable
characterization and evidence accounting; both are closed by retained
terminal evidence. Prose completion alone did not change either blocker to
`DONE`.

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| `RB-001` | The exact S-00 package in Sections 7.1 through 7.9 requires retained passing evidence before structural movement. | Moving the broad façade or shared transaction sequence without that evidence could change public, storage, revision, projection, idempotency, or publication behavior. | Owner-backed S-00 tests and applicable cross-owner evidence | DONE: exact symbols, transitional accounting, artifact owner rows, collaborator slices, and no-product-diff review passed |
| `RB-002` | The artifact owner manifest required the immutable final catalog and coarse-ID retirement. | A passing transitional owner slice could not demonstrate complete artifact behavior or the final evidence-class gates. | The exact verification/row cutover and owner evidence audit below | DONE: final accounting is 9 rows, 4 unit, 5 integration, and 7 verifications with exact selector resolution and no retired aliases |

The current reporting provider's omission of investigative-query and
forensic-keyword facts is not an open blocker for a behavior-preserving
refactor. It must be preserved. Any expansion requires separate adopted owner
evidence and later authorization.

### 11.1 RB-001 state transition

RB-001 MUST remain `BLOCKED` until every Section 7.9 condition is supported by
named result roots. When those conditions pass, the handoff MUST record the
exact test symbols, cross-owner suites, result roots, and a repository diff
proving S-00 changed no production code. Only then MAY RB-001 and ART-005 become
`DONE`.

### 11.2 Final artifact verification contract

A later authorized RB-002 task MUST make the final active artifact verification
contract exactly:

| Verification ID | Behavior class | Profile | Evidence kinds | Skip policy |
| --- | --- | --- | --- | --- |
| `module.artifacts.verification.surface_contracts` | `product` | `base` | `["go_test"]` | default `forbid` |
| `module.artifacts.verification.source_mutations` | `product` | `base` | `["go_test"]` | default `forbid` |
| `module.artifacts.verification.collection_semantics` | `product` | `base` | `["go_test"]` | default `forbid` |
| `module.artifacts.verification.linked_note_atomicity` | `product` | `base` | `["go_test"]` | default `forbid` |
| `module.artifacts.verification.import_create_facade` | `product` | `extension.import` | `["go_test"]` | default `forbid` |
| `module.artifacts.verification.conflict_source_revalidation` | `product` | `base` | `["go_test"]` | default `forbid` |
| `module.artifacts.verification.rollback_source_mapping` | `product` | `base` | `["go_test"]` | default `forbid` |

Verification entries MUST contain routing semantics only. They MUST NOT copy
Core requirement text, acceptance prose, document paths, or test scenarios.

### 11.3 Final nine-row artifact catalog

The IDs below were derived with the current Make-owned
`author-test-row-id` command from the fixed family, semantic claim, and exact
selector. They are immutable. An implementation MUST NOT replace them with
unhashed aliases, recycle the old row ID, or widen a selector.

| Row ID | Verification | Exact Go selector | Collaborators | Evidence/default |
| --- | --- | --- | --- | --- |
| `module.artifacts.surface_contracts.artifact_surface_registry_recognizes_all_eight_c_778e06f594` | `surface_contracts` | `./internal/modules/artifacts`: `TestArtifactSurfaceContractMatrix` | `module.projections`, `module.revisions`, `module.workbook` | unit / `true` |
| `module.artifacts.surface_contracts.artifact_projection_provider_matches_canonical_a_6e485ff11b` | `surface_contracts` | `./internal/modules/artifacts/projectionprovider`: `TestArtifactProjectionProviderSurfaceContractMatrix` | `module.projections` | unit / `true` |
| `module.artifacts.source_mutations.artifact_source_create_and_patch_preserve_all_ei_c8d63cf5cd` | `source_mutations` | `./internal/modules/artifacts`: `TestArtifactWorkbookMutationContractMatrix` | `module.collaboration`, `module.projections`, `module.records`, `module.revisions`, `module.workbook` | integration / `false` |
| `module.artifacts.collection_semantics.artifact_collection_mutations_preserve_typed_col_ab3546b03e` | `collection_semantics` | `./internal/modules/artifacts`: `TestArtifactCollectionMutationContractMatrix` | `module.collaboration`, `module.links`, `module.projections`, `module.revisions`, `module.workbook` | integration / `false` |
| `module.artifacts.collection_semantics.artifact_risk_references_use_canonical_parent_sc_b69b618649` | `collection_semantics` | `./internal/modules/artifacts/riskrefs`: `TestRiskRefItemRefIsCanonical`, `TestRiskRefItemRefParsingIsStrict` | none | unit / `true` |
| `module.artifacts.linked_notes.artifact_linked_note_creation_is_atomic_across_f_bce1ac123b` | `linked_note_atomicity` | `./internal/modules/artifacts/linkednotes`: `TestArtifactLinkedNoteAtomicity` | `module.collaboration`, `module.links`, `module.projections`, `module.records`, `module.revisions`, `module.workbook` | integration / `false` |
| `module.artifacts.import_facade.artifact_import_facade_preserves_registered_targ_a54c97551b` | `import_create_facade` | `./internal/modules/artifacts`: `TestArtifactImportCreateFacadeContract` | `module.collaboration`, `module.imports`, `module.projections`, `module.records`, `module.revisions` | integration / `false` |
| `module.artifacts.conflict_revalidation.artifact_conflict_resolution_revalidates_source_07b695d48b` | `conflict_source_revalidation` | `./internal/modules/artifacts`: `TestArtifactConflictSourceRevalidation` | `module.collaboration`, `module.projections`, `module.revisions`, `module.workbook` | integration / `false` |
| `module.artifacts.rollback_provider.artifact_rollback_maps_all_scalar_artifact_varia_4126aafbba` | `rollback_source_mapping` | `./internal/modules/artifacts/rollbackprovider`: `TestSourceForRollbackValueMapsAllArtifactVariantsWithoutCollections`, `TestValidSourceRejectsSubtypeInvariantViolations` | `module.revisions` | unit / `true` |

Every row MUST use `runner="go"`, `claim_posture="implementation"`, and
`status="active"`. Unit rows MUST use runtime `none`, resource `go_balanced`,
and fixture `none`. Integration rows MUST use runtime `default`, resource
`go_clone_heavy`, and fixture `postgres_template_clone`. Collaborator arrays
MUST be sorted active owner IDs and MUST contain no aliases.

### 11.4 Immutable cutover

The cutover MUST execute in this order:

1. Keep
   `module.artifacts.store.workbook_facade_persists_artifact_backed_notes_780a6ca4f2`
   active with its current `default_check` value.
2. Add the seven Section 11.2 verifications and all nine Section 11.3 rows.
3. Generate accounting and run old and new evidence. The transitional state
   MUST contain 10 rows: 4 unit and 6 integration.
4. Prove that the new source-mutation matrix fully subsumes the old note-only
   postcondition.
5. Retire the old row and the coarse
   `module.artifacts.verification.behavior_contract` and
   `module.artifacts.verification.incident_portability` entries. Their IDs MUST
   NOT be reused.
6. Record old-to-new disposition in temporary migration evidence and retain
   the final reconciliation report. The temporary crosswalk MUST NOT become a
   runtime alias, selector input, or permanent catalog dependency.
7. Remove the temporary crosswalk after successful final reconciliation.
8. Regenerate and audit the final state: exactly 9 rows, 4 unit rows,
   5 integration rows, and 7 verification IDs.

The existing `module.incidentbundles` integration row
`module.incidentbundles.integration.incident_bundle_import_reuses_the_shared_upload_698c072755`
remains the owner of incident-bundle orchestration and retains
`module.artifacts` as a collaborator. Artifact source portability coverage MUST
NOT be lost when the coarse artifact portability verification is retired.

The four unit rows SHOULD remain in the normal default topology because their
declared cost is ordinary. The five broad integration rows MUST initially set
`default_check=false`. Promotion to default `make check` requires a later
measured topology decision; `default_check=false` does not make a row optional
for an explicit owner audit.

### 11.5 Binary RB-002 exit

RB-002 becomes `DONE` only when:

- final owner explanation reports 9 rows, 4 unit, and 5 integration;
- the full owner slice selects all nine rows;
- the service-backed owner slice selects exactly five rows;
- every selector resolves exactly once and no active selectors overlap;
- every row and derived applicable gate has one compatible successful terminal
  record;
- all prescribed collaborator slices pass;
- authored and generated accounting are drift-free;
- old identifiers have a retained terminal disposition and no alias; and
- the handoff records exact result roots and the final reconciliation.

Only then MAY ART-011 become `DONE` and S-07 declare the overall refactor
complete.

## 12. Binary Completion Criteria

Completion has four distinct states. A later session MUST NOT use completion
of an earlier state as evidence that a later state is complete.

### 12.1 Tracker specification completeness

The documentation-only revision is complete only when all of the following
criteria are true:

| Acceptance ID | Binary criterion | Current result |
| --- | --- | --- |
| ART-AC-T01 | Every file under `internal/modules/artifacts` has an individual Section 2 inventory row. | PASS: 25 of 25 files are inventoried. |
| ART-AC-T02 | Every discovered observable-contract risk has an explicit owner and test posture. | PASS: Sections 4, 5.1, and 7 define both. |
| ART-AC-T03 | The eight-surface, import-availability, collection, mutation-effect, transaction-owner, postcondition-owner, reporting, and document-disposition mappings are defined once and referenced by later requirements. | PASS: Sections 3.1, 4.1 through 4.3, 5.1 through 5.2, and 7.3 through 7.7 define the mappings. |
| ART-AC-T04 | Every workflow has explicit predecessors, successors, validation, and a handoff checkpoint. | PASS: Section 6 defines the complete dependency graph and checkpoints. |
| ART-AC-T05 | Every slice has a dependency, intended change, contract risks, tests, validation, rollback, and completion criterion. | PASS: Section 7.10 defines all fields. |
| ART-AC-T06 | Every planned structural slice preserves observable behavior; any behavior correction is marked `requires later authorization`. | PASS: S-01 through S-07 preserve behavior and S-08 carries the authorization gate. |
| ART-AC-T07 | RB-001 names exact test interfaces, required cases, effect assertions, fault-injection precedence, and a binary exit gate. | PASS: Sections 7.1 through 7.9 are decision-complete. |
| ART-AC-T08 | RB-002 names exact immutable verification IDs, hashed row IDs, selectors, metadata, cutover order, accounting totals, and a binary exit gate. | PASS: Sections 11.2 through 11.5 are decision-complete. |
| ART-AC-T09 | Canonical validation commands are discovered or explicitly marked `TODO:` with the missing evidence. | PASS: Sections 8 and 8.1 define the future execution order. |
| ART-AC-T10 | No owner contradiction is unresolved and the framework/repository mismatch is recorded. | PASS: Section 1 records both postures. |
| ART-AC-T11 | Generated files remain generator-owned, and phase maps and harness rows are used only for evidence accounting. | PASS: Sections 5.2 and 11.4 prohibit hand edits and runtime inference. |
| ART-AC-T12 | Every required handoff subsection preserves prior history and contains a current documentation-revision row. | PASS: Section 10 contains both sessions in all seven subsections. |
| ART-AC-T13 | Only this tracker changed, Markdown lint passed, and whitespace validation passed. | PASS only after the commands recorded in the current Section 10 handoff rows succeed. |

When ART-AC-T01 through ART-AC-T13 pass, this tracker is complete as a
planning specification. That result MUST NOT change RB-001, RB-002, ART-005,
or ART-011 from `BLOCKED`.

### 12.2 RB-001 execution closure

RB-001 is complete only when all of the following are true:

- every Section 7.1 S-00 test exists and passes;
- every required case in Sections 7.2 through 7.7 is exercised;
- every expected effect traces to an adopted owner or an explicitly frozen
  current behavior;
- fault injection complies with Section 7.8;
- the S-00 change contains no production-code diff; and
- the result root and exact terminal evidence are recorded in Section 10.

Until all six criteria pass, RB-001 and ART-005 MUST remain `BLOCKED`.

### 12.3 RB-002 execution closure

RB-002 is complete only when the Section 11.4 cutover has completed and the
Section 11.5 gate proves exactly:

- 9 active `module.artifacts` rows;
- 4 unit rows and 5 integration rows;
- 7 active artifact verification IDs;
- all nine rows selected by the owner slice;
- exactly five rows selected by the service-backed owner slice; and
- compatible retained evidence for every applicable row and gate.

Until every criterion passes, RB-002 and ART-011 MUST remain `BLOCKED`.

### 12.4 Refactor completion

The artifact refactor is complete only when:

- RB-001 and RB-002 are `DONE`;
- S-01 through S-07 satisfy their individual completion criteria;
- all Section 8.1 validation required by the affected risk surface passes;
- the eight surface mappings, public routes, exported façade DTOs, field keys,
  error envelopes, storage shapes, row-version semantics, authorization order,
  revisions, projections, idempotency, and collaboration semantics remain
  observably unchanged;
- reporting still emits exactly the six Section 4.3 families and emits no
  investigative-query or forensic-keyword facts;
- no schema migration or public test interface was introduced;
- generated outputs are reproduced only through Make-owned generation; and
- the final handoff records changed files, result roots, reconciliation
  evidence, rollback posture, and any separately authorized behavior change.

Current result: **PASS**. RB-001 and RB-002 are `DONE`; S-01 through S-07 and
the final S-07 validation slice pass; the final owner contract is 9 rows,
4 unit rows, 5 integration rows, and 7 verification IDs; no owner-defined
observable behavior, schema, public test interface, import target set, or
reporting family set changed. Section 13 records the exact evidence and
rollback posture.

No acceptance claim MAY be inferred from filenames, visible labels, SQL text,
generated files, package location, historical phase names, or research
material. Such sources MAY identify evidence to inspect, but only adopted
owners and explicitly frozen live behavior establish the acceptance
expectation.

## 13. Remediation Execution Ledger

This table is the controlling between-workstream checkpoint. Historical rows
MUST remain append-only. A successor MUST NOT begin until its predecessor row
records a terminal status and a passing `make lint-markdown` result root.

| Time | Workstream | Status | Base and entry state | Changed files | Decisions and compatibility | Validation and result roots | Rollback and residual risk | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30 09:38 EDT | `WS-00` specification and tracker alignment | DONE | `main` at `b0e7d7dd2635f71374752ecc0029b52107bf4bcb`; clean entry tree | This tracker and `docs/handoffs/cartulary_modular_refactor_planning_framework.md` | Execution is authorized; Core/domain owners remain unchanged; the framework names the artifact source owner without creating a domain bounded context. No runtime compatibility change. | Baseline artifact slice passed at `.cartulary/test-results/20260730T133824Z-p3397835`; baseline service-backed slice passed at `.cartulary/test-results/20260730T133834Z-p3399151`; `make lint-markdown` passed at `.cartulary/test-results/20260730T133922Z-p3400695`; `git diff --check` passed. | Documentation can be reverted independently; no product risk introduced. | Begin WS-01/S-00. |
| 2026-07-30 09:40 EDT | `WS-01 / S-00` characterization and transitional evidence adoption | DONE | `main` at `b0e7d7dd2635f71374752ecc0029b52107bf4bcb`; WS-00 documentation changes present | Added the seven exact test symbols, adopted four existing risk-reference/rollback symbols, added seven verification IDs and nine immutable rows while retaining the old row, regenerated topology, updated harness catalog assertions, and refreshed duration baselines through the public target. | Transitional accounting is exactly 10 rows, 4 unit, and 6 integration. S-00 changed tests, authored/generated harness inputs, and documentation only; `git diff --name-only` found no product-code path. S-00R-1 remains the separately recorded harness correction. No wire, schema, route, authorization, or data compatibility change. | `make test-catalog-check` passed at `.cartulary/test-results/20260730T135727Z-p3425562`; artifact service-backed slice passed 6 tests at `.cartulary/test-results/20260730T135841Z-p3432041`; artifact full owner slice passed all 10 rows at `.cartulary/test-results/20260730T135902Z-p3433675`; full service-backed evidence passed at `.cartulary/test-results/20260730T135959Z-p3436294`; duration baseline refresh passed at `.cartulary/test-results/20260730T140420Z-p3516695`; harness contract passed at `.cartulary/test-results/20260730T140444Z-p3517072`. Collaborator full-owner slices passed for workbook `.cartulary/test-results/20260730T140539Z-p3520339`, projections `.cartulary/test-results/20260730T140843Z-p3557279`, revisions `.cartulary/test-results/20260730T140903Z-p3558974`, imports `.cartulary/test-results/20260730T140957Z-p3579258`, incident bundles `.cartulary/test-results/20260730T141043Z-p3599829`, reporting `.cartulary/test-results/20260730T141127Z-p3601823`, recovery `.cartulary/test-results/20260730T141137Z-p3603235`, collaboration `.cartulary/test-results/20260730T141254Z-p3633457`, records `.cartulary/test-results/20260730T141518Z-p3662031`, and links `.cartulary/test-results/20260730T141534Z-p3663949`. `git diff --check` passed. | Characterization, catalog inputs, generated projections, and baseline refresh can be reverted as one slice; S-00R-1 can be reverted separately. No residual structural behavior change exists. | Run the Markdown checkpoint, then mark WS-02/S-01 active. |
| 2026-07-30 09:53 EDT | `S-00R-1` explicit Go evidence-class routing | DONE | `main` at `b0e7d7dd2635f71374752ecc0029b52107bf4bcb`; S-00 tests/catalog and WS-00 docs present | `tools/harness/test-catalog/target-routing.mjs` and `tools/harness/tests/test-harness-contracts.mjs` | A Go row with explicit unit evidence and runtime `none` routes to `backend-unit` independent of semantic family naming. Service-backed unit evidence and specialized store, support-integration, and process routing remain unchanged. No product, wire, or data compatibility impact. | Defect reproduced by `make generate` at `.cartulary/test-results/20260730T135215Z-p3413726`; focused router assertion passed within the 96 successful `make harness-contract` cases at `.cartulary/test-results/20260730T135508Z-p3418488`; `make generate` passed at `.cartulary/test-results/20260730T135543Z-p3419745`. The remaining two harness-contract failures are attributable to expected transitional catalog counts and not-yet-measured new test baselines. | The two-file harness correction can be reverted independently, but doing so makes semantic unit families unroutable. | Resume S-00 and close its catalog accounting and evidence. |
| 2026-07-30 10:18 EDT | `WS-02 / S-01` source-policy and persistence decomposition | DONE | `main` at `b0e7d7dd2635f71374752ecc0029b52107bf4bcb`; WS-00/WS-01 changes present | Added `internal/modules/artifacts/source_policy.go`; changed `create_validation.go`, `mutations.go`, and `workbook_facade.go`; extended `surface_contract_test.go`. | One artifact-owned catalog now cross-checks all eight canonical filters and every authored writable direct/collection field. SQL identifiers come only from an exact static mapping; prefix-derived identifiers are gone. Unknown views/fields, read-only fields, cross-surface fields, collection-as-scalar values, wrong value kinds, non-clearable nulls, and cross-surface collections fail before mutation SQL. Existing DTOs, SQL tables, defaults, routes, and the temporary `Store` façade remain compatible. | Entry Markdown passed at `.cartulary/test-results/20260730T141908Z-p3684271`; `make format` passed at `.cartulary/test-results/20260730T142158Z-p3686937`; all 10 artifact rows passed at `.cartulary/test-results/20260730T142202Z-p3690056`; boundary check passed at `.cartulary/test-results/20260730T142230Z-p3692922`. An initial drift check at `.cartulary/test-results/20260730T142232Z-p3693194` correctly reported that refreshed duration baselines had not yet been projected; `make generate` passed at `.cartulary/test-results/20260730T142307Z-p3697147` and the rerun drift check passed at `.cartulary/test-results/20260730T142314Z-p3699349`. | Revert the four implementation files and focused test additions as one slice; no caller migration depends on the catalog yet. Residual transaction-coordination duplication is assigned to S-02. | Run the Markdown checkpoint, then mark WS-03/S-02 active. |
| 2026-07-30 10:27 EDT | `WS-03 / S-02` unified transactional mutation coordinator | DONE | `main` at `b0e7d7dd2635f71374752ecc0029b52107bf4bcb`; 20 paths reflected completed WS-00 through WS-02 work plus the initial unformatted S-02 extraction | Added root `contextual_notes.go`, `mutation_kernel.go`, and `mutation_ports.go`; changed `workbook_facade.go` and the `linkednotes` façade/test; removed `linkednotes/revision_append_port.go`. | Workbook and contextual-note creation now enter one transaction-owning coordinator and one transaction-supplied source kernel. Narrow typed ports cover incident admission, record envelopes, relationships, projections, revisions/history, idempotency, and source mutation. The old `linkednotes` package is a DTO/error translation bridge only; its mutation implementation and revision adapter are gone. The existing records envelope port supplies contextual source admission, preserving dependency direction. Routes, public request/result DTOs, idempotency keys, two-mutation contextual revisions, projection refresh, and error adaptation remain unchanged. Fault triggers prove all effects roll back at relationship, projection, revision, and idempotency stages. | Entry Markdown passed at `.cartulary/test-results/20260730T142752Z-p3706492`; formatting passed at `.cartulary/test-results/20260730T143209Z-p3721237`; artifact service-backed rows passed at `.cartulary/test-results/20260730T143213Z-p3724320`; workbook passed 87 tests at `.cartulary/test-results/20260730T143235Z-p3725962`; collaboration passed 32 tests at `.cartulary/test-results/20260730T143544Z-p3763695`. The first boundary run at `.cartulary/test-results/20260730T143810Z-p3792948` rejected a direct `records` table read in the extracted file; replacing it with the records envelope port made the boundary check pass at `.cartulary/test-results/20260730T143901Z-p3796835`. The final artifact slice passed all 10 rows at `.cartulary/test-results/20260730T143904Z-p3797146`, and records passed 9 tests at `.cartulary/test-results/20260730T143923Z-p3799903`. | The kernel, ports, contextual façade, and thin linked-note bridge can be reverted together without disturbing S-01. The translation bridge remains intentionally temporary until composition migration in S-05. | Run the Markdown checkpoint, then mark WS-04/S-03 active. |
| 2026-07-30 10:40 EDT | `WS-04 / S-03` canonical artifact surface registry | DONE | `main` at `b0e7d7dd2635f71374752ecc0029b52107bf4bcb`; WS-00 through WS-03 changes present | Added `internal/modules/artifacts/surfacecatalog/catalog.go`; changed root `surfaces.go`, source policy, projection provider and tests, revision contribution, and delete/restore provider. | One cycle-free exact-eight registry owns surface admission. It derives every discriminator from generated view-schema canonical filters without fallback, supports exact reverse lookup for adapters, and is consumed by root mapping, source policy, projection queries, revision registration, import admission through the root façade, and delete/restore routing. Root constants remain aliases. Unknown generated or artifact-like surfaces are not admitted automatically; adding one remains a deliberate registry and provider-field change. Provider-local IDs and canonical-filter fallback logic are gone. | Entry Markdown passed at `.cartulary/test-results/20260730T144045Z-p3803753`; formatting passed at `.cartulary/test-results/20260730T144231Z-p3805991`; artifact passed all 10 rows at `.cartulary/test-results/20260730T144235Z-p3809107`; projections passed 8 tests at `.cartulary/test-results/20260730T144254Z-p3812368`; revisions passed 24 tests at `.cartulary/test-results/20260730T144316Z-p3814974`; the first boundary check passed at `.cartulary/test-results/20260730T144413Z-p3836964`. After moving source policy and delete/restore to the same registry, formatting passed at `.cartulary/test-results/20260730T144508Z-p3837776`, the final artifact slice passed at `.cartulary/test-results/20260730T144512Z-p3840859`, the boundary check passed at `.cartulary/test-results/20260730T144531Z-p3843777`, and `git diff --check` passed. | Revert the catalog and its consumers as one S-03 slice; the mutation kernel remains independent. No residual registry duplication remains in the root or projection provider. | Run the Markdown checkpoint, then mark WS-05/S-04 active. |
| 2026-07-30 10:46 EDT | `WS-05 / S-04` thin source-owner adapters | DONE | `main` at `b0e7d7dd2635f71374752ecc0029b52107bf4bcb`; WS-00 through WS-04 changes present | Added `source_kernel.go` and `sourcecontract/rollback.go`; changed workbook/contextual mutation internals, import adapter/projection seam, source policy, and rollback provider/tests. Portability, recovery, reporting, delete/restore, and provider-contribution files already meeting the boundary were left unchanged except for the S-03 delete/restore registry lookup. | Interactive and import creation now share one transaction-supplied source kernel for record envelope, source rows, and projection refresh. The import adapter retains the caller's transaction and delegates revision finalization to the import owner contract. Exact scalar storage identifiers, rollback projection mapping, and reusable rollback validation now live behind the lower artifact source contract. The rollback provider translates retained values and executes subtype restore SQL only. No reviewed adapter begins or commits a transaction. Five portability files, authoritative recovery tables, eight revision routes, five enabled import targets, and six reporting families remain unchanged. | Entry Markdown passed at `.cartulary/test-results/20260730T144634Z-p3845978`. A syntax error during the initial source-kernel extraction was caught by `make format` at `.cartulary/test-results/20260730T144825Z-p3848363`; the corrected formatting run passed at `.cartulary/test-results/20260730T144857Z-p3851641`. Artifact passed at `.cartulary/test-results/20260730T145213Z-p3861708`; imports passed 27 tests at `.cartulary/test-results/20260730T145232Z-p3864569`; revisions passed 24 tests at `.cartulary/test-results/20260730T145321Z-p3885663`; the boundary check passed at `.cartulary/test-results/20260730T145416Z-p3906817`; incident bundles passed 24 tests at `.cartulary/test-results/20260730T145423Z-p3907154`; recovery passed 23 tests at `.cartulary/test-results/20260730T145502Z-p3910495`; reporting passed 11 tests at `.cartulary/test-results/20260730T145621Z-p3940291`. A projections guard at `.cartulary/test-results/20260730T145648Z-p3942105` correctly rejected moving projection-provider construction out of its approved integration seam; restoring a thin `import_projection.go` constructor made projections pass 8 tests at `.cartulary/test-results/20260730T145742Z-p3947988` and the boundary check pass at `.cartulary/test-results/20260730T145804Z-p3950254`. `git diff --check` passed. | The source kernel/import adapter and rollback source contract can be reverted independently. The approved projection construction seam is intentionally retained; it owns no mutation semantics. Composition still carries the temporary linked-note DTO bridge for S-05. | Run the Markdown checkpoint, then mark WS-06/S-05 active. |
| 2026-07-30 10:59 EDT | `WS-06 / S-05` composition migration | DONE | `main` at `b0e7d7dd2635f71374752ecc0029b52107bf4bcb`; WS-00 through WS-05 changes present | Changed `internal/app/workbookassembly/store.go`, `internal/modules/workbook/store.go`, and `internal/modules/workbook/mutation_store.go`. | Workbook composition now injects the root `ContextualNoteFacade` through a narrow contextual-note port. Workbook translates its DTOs directly to root artifact DTOs and adapts artifact and collection validation errors; no production caller imports the legacy `linkednotes` façade. Route keys, request hashes, source-record scoping, authentication/authorization ownership, public DTOs, and response envelopes are unchanged. The bridge remains only for its owned characterization test and will be deleted in S-06 after that test is migrated to the root façade. | Entry Markdown passed at `.cartulary/test-results/20260730T145931Z-p3952556`; formatting passed at `.cartulary/test-results/20260730T150034Z-p3954750`; artifact passed all 10 rows at `.cartulary/test-results/20260730T150038Z-p3957838`; workbook passed all 87 tests at `.cartulary/test-results/20260730T150057Z-p3960541`; the boundary check passed at `.cartulary/test-results/20260730T150410Z-p3998624`; collaboration passed all 32 tests at `.cartulary/test-results/20260730T150422Z-p3998974`; production-caller search found only stale projection metadata references to the legacy path, assigned to S-06; `git diff --check` passed. | Revert the three composition/workbook files to restore the DTO bridge without changing the unified root mutation path. The remaining legacy production façade and stale provider metadata are explicitly assigned to S-06 removal. | Run the Markdown checkpoint, then mark WS-07/S-06 active. |
| 2026-07-30 11:07 EDT | `WS-07 / S-06` legacy removal, guardrails, and final catalog cutover | DONE | `main` at `b0e7d7dd2635f71374752ecc0029b52107bf4bcb`; WS-00 through WS-06 changes present | Deleted the production `linkednotes` façade/revision adapter and old note-only test; migrated the immutable linked-note test selector to the root façade; made the source store private; updated projection provider metadata/policy; changed artifact verification/family inputs, harness final counts, boundary policy, and Make-generated topology. | No production linked-note compatibility package, exported artifact `Store` constructor, old test row, coarse verification ID, runtime alias, or feature flag remains. The test-only linkednotes directory remains solely because the immutable final row selects that package. New boundary rules prevent the artifact source/mutation kernels from importing consumer orchestration, publication, generic projection, or HTTP concerns and restrict direct surface-catalog imports to the artifact owner. Final accounting is exactly 9 rows, 4 unit, 5 integration, and 7 verification IDs. The old note postcondition is included in the eight-surface source mutation matrix, so retirement loses no product evidence. | Entry Markdown passed at `.cartulary/test-results/20260730T150815Z-p4030511`; formatting passed at `.cartulary/test-results/20260730T151000Z-p4032858`. The expected pre-generation JSON check at `.cartulary/test-results/20260730T151004Z-p4035940` reported stale topology; `make generate` passed at `.cartulary/test-results/20260730T151014Z-p4036504`, JSON shape passed at `.cartulary/test-results/20260730T151021Z-p4038705`, catalog check passed at `.cartulary/test-results/20260730T151024Z-p4039362`, harness contract passed at `.cartulary/test-results/20260730T151026Z-p4039612`, and boundary policy passed at `.cartulary/test-results/20260730T151053Z-p4040869`. Drift passed at `.cartulary/test-results/20260730T151112Z-p4041280` and generated-artifact policy at `.cartulary/test-results/20260730T151123Z-p4045101`. The first final artifact slice at `.cartulary/test-results/20260730T151126Z-p4045463` exposed a test-helper dependency on the retired file; moving the helper into the surviving matrix made all 9 rows pass at `.cartulary/test-results/20260730T151220Z-p4051815` and all 5 service-backed rows pass at `.cartulary/test-results/20260730T151232Z-p4053606`. After removing stale projection metadata, an initial projection run at `.cartulary/test-results/20260730T151351Z-p4057564` identified the remaining facade-package mirror; the complete metadata cleanup made projections pass at `.cartulary/test-results/20260730T151434Z-p4060418`, boundary policy pass at `.cartulary/test-results/20260730T151454Z-p4062143`, and drift pass at `.cartulary/test-results/20260730T151457Z-p4062422`. Exact owner explanation reports 9 rows and 5 service-backed rows; repository search finds no retired IDs, selector, production bridge, or stale path outside tool-owned duration history. `git diff --check` passed. | Cleanup can be reverted separately from the required authored/generated catalog cutover, but restoring either retired path would reopen RB-002. Residual risk is limited to broad-suite integration and duration-baseline reconciliation in S-07. | Run the Markdown checkpoint, then mark WS-08/S-07 active. |
| 2026-07-30 11:36 EDT | `WS-08 / S-07` validation and handoff completion | DONE | `main` at `b0e7d7dd2635f71374752ecc0029b52107bf4bcb`; WS-00 through WS-07 changes present | Validation and this tracker only; final diff contains the authored specification, implementation, test, harness, boundary-policy, composition, generated-topology, and tracker changes recorded by the preceding workstreams. | Final validation followed the required narrow-to-broad order. Exact owner explanation reports 9 rows and 5 service-backed rows. Repository searches found no retired verification IDs, retired row selector, production linked-note façade caller, or exported artifact source-store constructor; the surviving `linkednotes` package is test-only and preserves the immutable atomicity selector. No owner-defined route, wire, field, error, storage, row-version, authorization, import-target, reporting-family, or collaboration behavior changed. No migration, feature flag, runtime alias, or public test interface was introduced. | Entry Markdown passed at `.cartulary/test-results/20260730T151648Z-p4068332`; JSON shape at `.cartulary/test-results/20260730T151657Z-p4070226`; catalog at `.cartulary/test-results/20260730T151700Z-p4070876`; generated-artifact policy at `.cartulary/test-results/20260730T151701Z-p4071051`; generation at `.cartulary/test-results/20260730T151703Z-p4071389`; drift at `.cartulary/test-results/20260730T151710Z-p4073579`; all 9 artifact rows at `.cartulary/test-results/20260730T151726Z-p4077491`; all 5 service-backed artifact rows at `.cartulary/test-results/20260730T151739Z-p4079230`. Collaborator slices passed for workbook `.cartulary/test-results/20260730T151807Z-p4080733`, projections `.cartulary/test-results/20260730T152121Z-p4118744`, revisions `.cartulary/test-results/20260730T152143Z-p4120422`, imports `.cartulary/test-results/20260730T152247Z-p4141603`, incident bundles `.cartulary/test-results/20260730T152334Z-p4161978`, reporting `.cartulary/test-results/20260730T152413Z-p4165139`, recovery `.cartulary/test-results/20260730T152429Z-p4166839`, records `.cartulary/test-results/20260730T152549Z-p2426`, links `.cartulary/test-results/20260730T152605Z-p4324`, and collaboration `.cartulary/test-results/20260730T152637Z-p22855`. Boundary policy passed at `.cartulary/test-results/20260730T152947Z-p52013`; pre-closure Markdown at `.cartulary/test-results/20260730T152950Z-p52364`; finalization at `.cartulary/test-results/20260730T152952Z-p53794`; `test-fast` passed 987 tests at `.cartulary/test-results/20260730T153023Z-p60399`; and `check` passed 191/191 work units and 838 reported tests at `.cartulary/test-results/20260730T153301Z-p124244`. The finalizer made no generated changes and skipped retained-run duration maintenance because `RESULTS_DIR` was unset. `git diff --check` passed. Browser-only follow-up was not run separately because no public wire or UI behavior changed; the broad check nevertheless passed its scheduled browser-backed and stateful work. | Revert by completed workstream in reverse order. Restoring the deleted compatibility façade or retired harness aliases is not part of rollback because doing so would reopen RB-002; if rollback must cross S-06, restore the corresponding authored catalog and generated topology together. No residual product or evidence blocker remains. | Run the post-closure Markdown lint and whitespace check, then hand off the complete diff without creating an unrequested commit. |
