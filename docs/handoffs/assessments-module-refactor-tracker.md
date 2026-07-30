# assessments Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Item | Posture |
| --- | --- |
| Target path | `internal/modules/assessments` |
| Target label | `assessments` |
| Output path | `docs/handoffs/assessments-module-refactor-tracker.md` |
| Planning status | Planning and documentation only |
| Allowed changes | This tracker file only |
| Non-goals | No production refactor, test change, contract change, generated output, dependency change, migration, package configuration, or harness/accounting change |
| Implementation authority | A later explicitly authorized task is required before any implementation slice may start |

The normalized target label is the lowercase kebab-safe value `assessments`.
The target path exists and contained 18 files at repository revision
`dd3e3b195b2484835fc45edb0b4ab39e5600a284`.

This tracker uses three explicit postures:

| Posture | Meaning | Normative effect |
| --- | --- | --- |
| **Current fact** | Directly observed owner text, repository code, test behavior, or versioned release evidence | Describes the state that a later task MUST characterize and preserve unless an adopted owner explicitly changes it. |
| **Proposed disposition** | A decision-complete resolution recommended by `temp/analysis-notes.md` and reconciled with live repository evidence | Defines planned owner amendments and acceptance gates; it is not adopted runtime authority. |
| **Adoption gate** | Owner-document, compatibility, projection, or executable-evidence work required before implementation | A gated slice MUST NOT implement the proposed disposition until its named gate passes. |

Normative `MUST`, `MUST NOT`, and `MAY` statements in this tracker govern
planning, sequencing, and proposed acceptance obligations. They do not
supersede an adopted Core or subsystem owner.

### Source hierarchy

1. Adopted subsystem NLSpecs govern only their named subsystem. No
   assessment-specific adopted subsystem NLSpec was found.
2. Core 00 through Core 04 govern implementation-conformance behavior.
3. Core 05 is not applicable to this tracker because it publishes no timed,
   fixture-sensitive, or other claim-bearing result.
4. `docs/domain.md`, design and implementation-support guidance, and the Testing
   Harness owner govern vocabulary, navigation, boundaries, and verification
   mechanics within their scopes.
5. Current repository code and tests establish current implementation state.
6. Prior plans, handoffs, research notes, and the planning framework are
   evidence only.

The live Core 01 corpus remains internally contradictory until its owner is
amended. The current state is therefore `BLOCKED: owner contradiction`.
RB-001 records a decision-complete proposed disposition with status
`RESOLVED_PENDING_ADOPTION`; the tracker itself does not adopt that
disposition.

### Owner and guidance documents inspected

- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/domain.md`
- `docs/research/nlspec-spec.md`
- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `temp/analysis-notes.md`

Core 01 explicitly recognizes `assessments` as a refinement for compromise
assessment source-state behavior. The planning framework's candidate module
catalog does not name assessments. That is a framework/repository mismatch, not
evidence that the live source-owner boundary is invalid. This tracker follows
the adopted owner and live repository.

Relevant prior OpenAPI, Workbook, Records, Recovery, Collaboration, incident,
and web trackers were inspected only as historical evidence. Their prior
completion states and retained run results are not current validation evidence
for this tracker.

### NLSpec drafting posture

`docs/research/nlspec-spec.md` requires behavioral and boundary completeness,
explicit defaults, stable mappings, and binary acceptance. This tracker applies
that discipline by:

- distinguishing observed behavior from a proposed owner repair;
- naming every default, rejection effect, transaction owner, and publication
  consequence needed to execute a later slice safely;
- routing normative changes to Core 01, Core 03, Core 04, and their
  traceability projections rather than treating this handoff as an owner; and
- retaining `TODO:` only where live evidence cannot safely establish a fact.

### Released-baseline evidence and compatibility posture

`contracts/openapi-releases/index.json` contains one indexed release entry:
version `1.0.0`, source commit
`53f4553150117d09535794e59a8f500485bdb94c`, with publication state
`historical_baseline`; `latest_release_id` is `null`. Static inspection at that
source commit found:

- the assessment discovery projection emitted
  `assessment.support_refs.manage` and `evidence.refs.manage`;
- the generic patch decoder rejected assessment fields whose view-schema entry
  was not writable; and
- the assessment frontend rendered generic inspector feature buttons without
  an assessment feature-action handler, so no functioning existing-row
  support-reference mutation was wired.

No indexed evidence of a successfully implemented
`assessment.support_refs.manage` mutation was found. The proposed disposition
therefore retains `cartulary.view.assessments.v1` as a contradiction repair.
That compatibility judgment remains an adoption gate: SL-00 MUST add executable
characterization, and the owner amendment MUST explicitly confirm the version
decision before any discovery or error-detail projection changes.

### Repository files inspected

Every target file was opened:

- `internal/modules/assessments/api.go`
- `internal/modules/assessments/assessment_contract_test.go`
- `internal/modules/assessments/assessments_integration_test.go`
- `internal/modules/assessments/deleterestore/provider.go`
- `internal/modules/assessments/import_create.go`
- `internal/modules/assessments/incident_bundle_portability.go`
- `internal/modules/assessments/incident_bundle_source_port.go`
- `internal/modules/assessments/incident_bundle_subtype_presence.go`
- `internal/modules/assessments/merge_effects.go`
- `internal/modules/assessments/projectionprovider/provider.go`
- `internal/modules/assessments/projectionprovider/query_surfaces.go`
- `internal/modules/assessments/recovery_state.go`
- `internal/modules/assessments/revision_append_port.go`
- `internal/modules/assessments/revision_provider_contribution.go`
- `internal/modules/assessments/rollbackprovider/provider.go`
- `internal/modules/assessments/rollbackprovider/provider_test.go`
- `internal/modules/assessments/store.go`
- `internal/modules/assessments/testsupport/assessments.go`

Current callers and composition were inspected in
`internal/app/workbookassembly/catalog.go`,
`internal/modules/workbook/mutation_contributions.go`,
`internal/modules/workbook/routes.go`, `internal/app/server/runtime.go`,
`internal/modules/entities/merge/ports.go`,
`internal/modules/entities/merge/merge_store.go`,
`internal/app/projectionassembly/catalog.go`,
`internal/modules/projections/assessments.go`,
`internal/modules/projections/services.go`,
`internal/app/importassembly/owner_registry.go`,
`internal/app/incidentportabilityassembly/catalog.go`,
`internal/app/recoveryassembly/state_catalog.go`, and
`internal/app/revisionassembly/revisions.go`.

Contract and verification inputs inspected include
`contracts/view-schemas/cartulary.view.assessments.v1.json`,
`contracts/openapi-source/owners/module.assessments/openapi.json`,
`contracts/imports/view-targets.v1.json`,
`contracts/projection-providers/index.json`,
`contracts/verification/owners/module.assessments.json`,
`tools/test_families/module.assessments.json`,
`tools/test_catalog_owner.json`,
`tools/backend_module_boundaries.json`, and
`tools/schema_object_ownership_manifest.json`. Related generated Go and
TypeScript projections were inspected only to identify affected public
surfaces; they remain read-only.

The frontend inspection covered
`apps/web/src/workbook/components/AssessmentWorkbookSurface.tsx`,
`apps/web/src/workbook/models/assessmentWorkbookModel.ts`,
`apps/web/src/workbook/timeline/models/workbookTimelineModel.ts`,
`apps/web/src/workbook/hooks/useAssessmentSupportRows.ts`,
`apps/web/src/workbook/policies/assessmentSurfacePolicies.ts`,
`apps/web/src/workbook/WorkbookActiveSurface.tsx`,
`apps/web/src/workbook/WorkbookShell.assessments.test.tsx`,
`apps/web/e2e/workbook.assessments.spec.ts`, and assessment browser fixtures.

No implementation is authorized by this document. A later task must select a
slice, establish its characterization gate, and authorize its writes.

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/assessments/api.go` | Assessment create DTOs, request decoding and normalization, idempotency hash, row and mutation payload presentation | `AssessmentsViewSchemaID`; `CreateRequest`; `ProjectionRecord`; `SupportRef`; `MutationResult`; `CreateValidationError`; decoder/hash/payload builders | Workbook create provider; assessment store/import path; application catalogs use the schema ID | Platform HTTP errors, field normalization, view-schema collection contracts, UUID/JSON | Assessment contract and integration tests; Workbook provider tests | Assessment view schema, assessment OpenAPI create request, generated Go/TS view and protocol contracts | Assessment application contracts, with HTTP adaptation owned by Workbook/platform | high | Transport-shaped errors and status-bearing results are mixed with semantic contracts. |
| `internal/modules/assessments/assessment_contract_test.go` | Store-backed assessment append-only, state/band, query, atomicity, patch-rejection, and link-confidence characterization | Test functions only | Go test harness and `module.assessments` test-family rows | Application test support, assessment store, Workbook/query, Records/Links test support | This file | Assessment view/query/link behavior | Assessments verification evidence | medium | Direct current evidence; it does not directly characterize assessment create WebSocket publication. |
| `internal/modules/assessments/assessments_integration_test.go` | HTTP create/query, normalization, support references, replay/divergence, validation, filtering, and ordering | Test functions only | Go test harness and `module.assessments` test-family rows | Server/application test composition and HTTP client support | This file | Generic Workbook route and assessment view contract | Assessments verification evidence | medium | Freezes `201`, replay `200`, divergent `409`, and failure envelopes. |
| `internal/modules/assessments/deleterestore/provider.go` | Assessment source snapshot and soft-delete/restore source-state adapter | `Source`; `NewSource`; source snapshot/delete/restore methods | Assessment revision provider contribution | Caller transaction and assessment source SQL | Revisions delete/restore and integration tests | Revision provider contract | Assessments source owner | medium | Correctly omits authorization, HTTP, publication, and projection coordination. |
| `internal/modules/assessments/import_create.go` | Assessment-owned import create and refresh facade inside the import caller's transaction | `ImportCreateCommand`; `NewImportCreateFacade`; `CreateImportRowTx`; `RefreshImportRowTx` | Import assembly generated target registry | Imports owner facade, Records, Links, Revisions, Projections, assessment store internals | Imports characterization/integration and boundary tests | `module.assessments@1`, `assessments.import_create`, generated import registry | Assessments source owner | high | Legitimate source-owner facade, but it reuses the broad store and duplicate row materialization. |
| `internal/modules/assessments/incident_bundle_portability.go` | Export/import of assessment source rows as bundle NDJSON | Export/import functions | Assessment incident-bundle source port | Incident portability query/transaction APIs and assessment source SQL | Incident-bundle API/integration tests | `data/compromise_assessments.ndjson` portability contract | Assessments source owner | high | Source data only; projection/runtime state remains excluded. |
| `internal/modules/assessments/incident_bundle_source_port.go` | Declares the assessment bundle source port, path, stable identity, owner, relation, and dependency metadata | `NewIncidentBundleSourcePort` | Incident portability application catalog | Incident-bundle source-port contract and local export/import functions | Incident-bundle catalog and integration tests | Incident portability source catalog | Assessments source owner | medium | Legitimate narrow contribution. |
| `internal/modules/assessments/incident_bundle_subtype_presence.go` | Validates that assessment record envelopes have corresponding assessment subtype rows | `IncidentBundleSubtypeContribution` and source methods | Incident portability application subtype catalog | Caller transaction, Records subtype-presence contract, assessment SQL | Incident-bundle subtype and import tests | Record-subtype presence contract | Assessments source owner | medium | Protects envelope/source invariants during bundle processing. |
| `internal/modules/assessments/merge_effects.go` | Loads the protected assessment set and repoints entity subjects in the entity merge transaction, refreshes projections, and returns mutations | Merge error/result/command DTOs and `Store` merge methods | Entities merge adapter around `assessments.Store` | Assessment SQL, projection facade, UUID/transaction contracts | Entity merge protected-set, resolution, support, and unit tests | Entity merge/change-set and assessment projection behavior | Assessments source owner behind a narrow merge-effects port | high | The effect belongs to assessments; the broad Store and exported DTO coupling do not. |
| `internal/modules/assessments/projectionprovider/provider.go` | Refreshes and rebuilds `assessment_grid_projection` with assessment, record, and link-derived values | `RefreshAssessmentTx`; `RebuildIncidentAssessmentsTx` | Projections assessment service and projection assembly | Direct assessment, Records, Links, and projection-table SQL | Assessment contract tests; Projections query/rebuild and source-ownership tests | Assessment projection-provider descriptor and table contract | Assessment derivation intent; physical storage lifecycle belongs to Projections | high | Source semantics and projection storage ownership are currently combined. |
| `internal/modules/assessments/projectionprovider/query_surfaces.go` | Declares assessment projection query surface, fields, joins, sort/filter behavior, and row presentation | `QuerySurfaces` | Projection assembly, assessment store, Projections tests | Projection provider contract and assessment schema/query vocabulary | Projections query-surface and query-store tests; assessment tests | Projection-provider manifest and assessment view contract | Assessments source/query semantics | high | Retain semantic descriptor while separating physical storage mechanics. |
| `internal/modules/assessments/recovery_state.go` | Declares assessment source table participation in recovery-state accounting | `RecoveryStateContribution` | Recovery application state catalog | Recovery-state contribution contract | Recovery catalog tests | Recovery state catalog | Assessments source owner | low | Evidence contribution only; not recovery orchestration. |
| `internal/modules/assessments/revision_append_port.go` | Local adapter from assessment mutation coordination to the Revisions appender | Private port/adapter methods | Assessment store construction and mutation paths | Revisions appender and caller transaction | Assessment and Revisions integration tests indirectly | Change-set, mutation, revision, and live-change contracts | Assessments application port with application-owned construction | medium | The abstraction is narrow, but concrete construction remains inside `NewStore`. |
| `internal/modules/assessments/revision_provider_contribution.go` | Contributes assessment delete/restore and rollback providers, view schema, and required live-record-change policy | `RevisionProviderContribution` | Revision assembly and focused entity/revision tests | Assessment delete/restore and rollback subpackages; Revisions provider contract | Revisions integration/delete/restore/rollback tests; entity merge tests | Revision provider catalog and `record_changed` intent policy | Assessments source owner | medium | Legitimate owner contribution. |
| `internal/modules/assessments/rollbackprovider/provider.go` | Validates historical assessment values and restores assessment source fields without derived confidence band | `Provider`; `NewProvider`; rollback validation/restore methods | Assessment revision provider contribution | Caller transaction, rollback contract, assessment/record SQL | Local provider test and Revisions rollback/integration tests | Rollback/source-history contract | Assessments source owner | high | Preserve source invariants and append-only history semantics. |
| `internal/modules/assessments/rollbackprovider/provider_test.go` | Characterizes rollback source shape and owner invariant rejection | Test functions only | Go test harness | Assessment rollback provider | This file | Revision rollback contract | Assessments verification evidence | low | Current narrow unit evidence. |
| `internal/modules/assessments/store.go` | Opens assessment create transaction; coordinates idempotency, validation, source insert, Records envelope, support links, projection refresh, revision/change set, response, and commit | `Store`; `NewStore`; `CreateAssessmentRow` | Workbook assembly/provider, Entities merge adapter, Projections/Workbook tests | Postgres, platform Auth/idempotency, Records, Links, Revisions, Projections, assessment projection provider; direct Hosts/Identities/Users/Records SQL | Assessment contract/integration; Projections query; Workbook create; Entities merge tests | HTTP behavior, assessment view row, revision, link, and projection contracts | Assessment mutation facade plus private source repository and injected peer ports | high | Central mixed-responsibility unit and primary refactor seam. |
| `internal/modules/assessments/testsupport/assessments.go` | Seeds assessments and reads assessment subjects for cross-module tests | `SeedAssessment`; `LookupAssessmentSubject` | Entities and Revisions tests only | Test database, Records and projection test setup | Entities resolution/support/unit and Revisions integration/delete/restore tests | Test fixtures only | Assessments owner-local test support | medium | No production import was found; retain but re-audit direct setup after facade changes. |

No target file is out of scope.

### Current transaction and publication map

| Entry path | Current transaction owner | Required target boundary | Commit/publication rule |
| --- | --- | --- | --- |
| Standalone Workbook create | `assessments.Store.CreateAssessmentRow` opens and commits the transaction after Workbook route authorization | Assessment `Facade.Create` MUST own one transaction and coordinate injected Records, Links, Revisions, Projections, and idempotency participants | Participants MUST use the caller transaction and MUST NOT commit or publish independently; publication follows the committed Revisions intent. |
| Import target create/refresh | Imports supplies the transaction to `CreateImportRowTx` or `RefreshImportRowTx` | Assessment import facade MUST retain caller-transaction participation | The assessment import facade MUST NOT commit, publish, or create a second transaction. |
| Entity merge effects | Entities supplies its merge transaction and change-set context | A narrow assessment merge-effects port MUST replace broad `Store` exposure | Assessment effects and projection refresh MUST remain in the Entities-owned transaction and change set; the port MUST NOT commit or publish. |
| Delete, restore, and rollback | Revisions owns generic coordination and supplies the transaction | Existing assessment source-owner providers remain narrow contributions | Providers MUST mutate only assessment-owned source state and MUST NOT authorize, commit, refresh, or publish independently. |
| Incident-bundle reconstruction | Incident Bundles owns import coordination and supplies the transaction | Assessment portability/source contributions remain owner-local | Contributions MUST preserve source rows and subtype invariants without creating runtime projections or ordinary patch semantics. |

## 3. Module Boundary Diagnosis

The target is a legitimate assessment source-owner boundary, not merely a
directory-shaped module. Its root package is nevertheless mixed: it is a
mutation coordinator, view/projection orchestration layer, transport-adjacent
adapter, and persistence-adjacent adapter. It is not a frontend shell or direct
grid-vendor integration layer.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Assessment source semantics and append-only admission | `api.go`, `store.go`, rollback/delete providers | Assessments | keep/split | Core 01 assessment refinement; Core 02 assessment model; live tests | Keep semantic validation and source writes; split transport and peer mechanics. |
| HTTP request/error/status adaptation | `api.go`, `store.go`, Workbook provider | Workbook/platform HTTP with typed assessment errors | split | Generic Workbook route owns method/path/auth; Store returns status codes | Assessment facade should return semantic replay/result state. |
| Assessment transaction sequencing | `store.go`, `import_create.go` | Assessments application facade | keep/split | Standalone create opens its transaction; import and merge paths receive caller transactions | Preserve atomic order behind injected ports and the entry-path-specific ownership in §2. |
| Assessment source persistence | `store.go`, rollback/delete/import/portability files | Assessments private repository | keep | Direct writes target assessment-owned source state | Do not move source SQL out of its owner. |
| Peer validation and state access | Direct Hosts/Identities/Users/Records SQL in `store.go` | Entities, Auth, and Records typed ports | move | Cross-owner table reads are visible in live source | Ports must accept caller transactions where atomicity requires it. |
| Projection derivation and query intent | `projectionprovider/**` | Assessments | keep/split | Provider descriptor names assessments as source owner | Retain canonical semantic input/derivation and query descriptors. |
| Projection table DML/lifecycle | `projectionprovider/provider.go` | Projections | move | Descriptor names Projections as storage owner; Core 01 separates source and storage ownership | Do not move assessment source semantics with the physical writer. |
| Entity merge assessment effects | `merge_effects.go`, Entities adapter | Assessments narrow effect provider; Entities coordinates merge | keep/split | Core 02 requires repoint/tombstone in the same change set | Remove broad Store exposure, not the source-owned effect. |
| Revision/delete/restore/rollback rules | Assessment provider contributions | Assessments contributions; Revisions coordinates | keep | Core 01 requires source owners to construct record-type adapters | Current subpackages are appropriately narrow. |
| Import target creation | `import_create.go` | Assessments owner facade; Imports coordinates | keep/split | Generated target binds `module.assessments@1` to `assessments.import_create` | Migrate away from broad Store internals without changing target identity. |
| Incident portability and recovery presence | Assessment contribution files | Assessments contributions; outer subsystems coordinate | keep | Live application catalogs consume narrow contributions | No runtime architecture should be inferred from harness rows. |
| Assessment frontend controller/grid | `AssessmentWorkbookSurface.tsx` and Workbook shell | Web Workbook assessment surface | keep | Dedicated contract-driven renderer and shell registration | It correctly imports `@cartulary/grid-adapter`, not a vendor grid. |
| Assessment draft types and band/payload builder | Frontend Timeline model | Assessment Workbook model | move | Assessment component/tests import assessment types from a Timeline-owned file | Pure structural move; wire payload must remain identical. |
| Supporting-record picker limited to Timeline rows | `useAssessmentSupportRows.ts` | Assessment Workbook current-profile policy | keep | Current hook queries `cartulary.view.timeline.v2`; backend accepts broader first-class record references | Timeline-only is the resolved current-profile interaction. Broader selection is separate product work. |
| Saved-view/query state | Generic Workbook shell and contracts | Workbook/Saved Views | keep | No assessment-specific saved-view implementation exists in the target | Assessment refactor must preserve stable field keys and query behavior. |

### Required target interfaces

The names below are planning names, not pre-adopted Go or TypeScript symbols.
A later design slice MAY choose idiomatic local names, but it MUST preserve
these ownership, input, output, and side-effect boundaries.

| Interface or component | Owner | Required inputs | Required result | Forbidden behavior |
| --- | --- | --- | --- | --- |
| Standalone assessment create facade | Assessments | Authorized actor context, incident ID, canonical create command, injected owner ports | Semantic result distinguishing new commit, exact replay, divergent replay, validation failure, and internal failure | MUST NOT return HTTP status as domain state, bypass caller authorization, accept client-selected link types, or expose source persistence. |
| Assessment import facade | Assessments | Caller transaction, import owner context, stable target ID, canonical assessment values | Owner-local mutation and canonical row/revision contribution within the caller transaction | MUST NOT open or commit a transaction, publish independently, or change `module.assessments@1`. |
| Support-target validation port | Records or the adopted record-visibility owner | Caller transaction, authorized actor context, incident ID, deduplicated stable record IDs | Validated active, visible, same-incident first-class record references or one semantic validation failure | MUST NOT disclose a target before incident authorization, commit, publish, or infer identity from a row index. |
| Support-link mutation port | Links | Caller transaction, assessment record ID, validated support record IDs, owner field identity | Field-derived active `supported_by` links and deterministic replay behavior | MUST NOT accept client-provided link type, direction, table, or storage metadata; MUST NOT commit or publish. |
| Revision append participant | Revisions | Caller transaction, change-set and mutation intents, live-change policy | Durable revision/mutation contribution and post-commit `record_changed` intent | MUST NOT publish before commit or create a second change set for the same assessment create. |
| Projection participant | Projections with assessment derivation supplied by Assessments | Caller transaction, incident/record identity, canonical assessment source snapshot | Refreshed or rebuilt projection state matching the canonical row derivation | MUST NOT treat projection state as source authority or commit independently. |
| Workbook HTTP adapter | Workbook | Generic route path/envelope, authenticated request, assessment semantic result | Existing `201`, replay `200`, validation `400`, and divergent replay `409` envelopes | MUST NOT reconstruct assessment persistence or owner validation rules. |
| Workbook view-query client | Web Workbook | Existing generic query route, Timeline view-schema ID, current request defaults | Authorized Timeline view rows | MUST NOT query storage tables, enumerate every view, or create an assessment-specific search route. |
| Assessment candidate-source policy | Assessment Workbook model/hook | `cartulary.view.timeline.v2`, query result row | `{recordId, displayText}` keyed by stable `record_id` | MUST NOT own persistence validation, use row position as identity, or import Timeline mutation/draft logic. |
| Generic candidate picker | Web Workbook shared presentation | Controlled candidates, selected record IDs, callbacks | Rendered and keyboard-operable controlled selection | MUST NOT define assessment target eligibility or mutate persisted state on cancel. |
| Grid adapter | `@cartulary/grid-adapter` | Contract-shaped rows, columns, focus, and selection state | Rendering and interaction primitives | MUST NOT own support-link semantics, authorization, or target validation; assessment code MUST NOT import a vendor grid directly. |

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` for assessments | Workbook route; Assessments create provider/source owner | Workbook routes/provider, assessment OpenAPI fragment, Core 01 | Assessment integration and Workbook route tests | Preserve exact decoder/hash/error vectors while splitting facade | high | New create `201`, exact replay `200`, divergent replay `409`. |
| Generic assessment query route and row envelope | Workbook/Projections with assessment query descriptor | View schema, query surface, projection service | Assessment contract/integration, Projections query tests, browser tests | Row-byte/JSON parity before canonical row-path consolidation | high | Preserve cursor, sort, filter, hidden fields, and saved-view queries. |
| Create authorization and CSRF | Workbook route and platform auth | Route code and scenario inventory | Workbook/auth route scenarios | Re-run role/CSRF scenarios after facade migration | high | Viewer denial and missing-membership behavior remain unchanged. |
| Incident WebSocket `/ws/v1/incidents/{incident_id}` and `record_changed` | Collaboration stream; Revisions live-change catalog; Assessments contribution | Collaboration route, Revisions appender/catalog, assessment revision contribution | Generic Collaboration/Revisions tests | Direct assessment create success, failure, replay, import/history policy assertions | high | No assessment-specific socket path exists. |
| Append-only assessment source semantics | Assessments | Core 02 assessment requirements and source tests | Assessment contract/integration and rollback tests | Preserve accepted states, prohibited response states, defaults, and patch rejection | high | Existing semantic fields remain immutable after first commit. |
| Confidence score/band derivation | Assessments source plus projection | Core 01/02 and query descriptor | Go, frontend unit, and browser assessment tests | Canonical derivation equivalence during projection split | high | `unset`, `low`, `medium`, `high` map to `NULL`, `25`, `55`, `85` in band-first UI. |
| `assessment.support_refs` create/read/link behavior | Core 01; Assessments; Links | REQ-01-332 through REQ-01-335, live decoder/store, view schema | Link-confidence, integration, browser tests | Omission, heterogeneous targets, duplicates, every invalid-target class, exact add/remove patch rejection, and side-effect absence | high | Current owner contradiction is resolved only as a proposed disposition; adoption is required. |
| Assessment projection refresh/rebuild/query | Assessments source intent; Projections storage/lifecycle | Provider descriptor, source provider, projection service | Assessment and Projections query/rebuild tests | Refresh/load/rebuild equivalence and deterministic rebuild | high | Projection is derived/disposable and never source authority. |
| Change set, mutation, revision, and history | Revisions coordination; Assessments source rules | Store/import paths and revision contribution | Assessment, Revisions, merge tests | Exact live and historical intent assertions across create/import/rollback | high | Preserve row version, change-set identity, and history payloads. |
| Entity merge repoint/protected set | Entities coordinates; Assessments owns assessment effect | Merge adapter/effect and Core 02 | Entity merge protected-set/resolution/support tests | Narrow-port parity, changed protected-set failure, projection refresh | high | Must remain in the same merge transaction and change set. |
| Delete/restore/rollback | Revisions coordinates; Assessments providers own source rules | Provider contribution and subpackages | Revisions delete/restore/rollback and local provider tests | Provider-catalog parity after package/facade movement | high | No provider may acquire its own transaction or publish independently. |
| Import target `module.assessments@1` | Imports lifecycle; Assessments owner facade | Authored and generated import target registries | Imports characterization/integration tests | Direct assessment import parity, revision, projection, and failure atomicity | high | Target and facade IDs are frozen. |
| Incident bundle assessment source | Incident Bundles coordinates; Assessments contributes source | Source port, subtype contribution, assembly | Incident-bundle API/integration tests | Round-trip and missing-subtype failure coverage tied to contribution | high | Path remains `data/compromise_assessments.ndjson`. |
| Recovery-state participation | Recovery coordinates; Assessments contributes source table | Recovery contribution/catalog | Recovery catalog tests | Catalog presence/parity after any file move | medium | Test/accounting presence is not runtime ownership. |
| Assessment view schema and inspector config | Core 01 and view-contract owner | Authored assessment view schema and generated artifacts | View-contract and Workbook tests | Exact removed/retained/added feature set, seed bindings, role, result behavior, and compatibility characterization | high | Proposed repair retains v1, removes two contradictory mutations, and adds follow-on create only after adoption. |
| Assessment frontend draft, submit, stale-query, presence, and grid behavior | Core 03 interaction owner; Web Workbook implementation | Assessment component/model/tests and browser spec | Vitest and Playwright assessment rows | Assessment-owned DTO move, stable record-key selection, non-Timeline read preservation, follow-on defaults, cancellation, stale-target handling | high | Timeline-only selection is the resolved current-profile policy; backend admissibility remains broader. |
| Generated OpenAPI/protocol/view/import projections | Contract/codegen owners | Authored inputs and generated Go/TS outputs | Drift, JSON-shape, protocol/view-contract tests | No generated semantic diff for behavior-preserving slices | high | Generated roots must never be hand-edited. |
| Harness owner and row accounting | Verification/harness owners | Verification owner, test family, catalog, topology | `make explain-test-owner` evidence | Add or move authored rows only when executable identities change | medium | Rows prove selection/accounting, not architecture or specification completeness. |

### Proposed assessment-support contract

The following rules are the proposed RB-001 disposition. They are
decision-complete planning inputs but remain subject to the Core 01/Core 03/Core
04 adoption gate.

| Concern | Required proposed behavior | Owner location | Compatibility/evidence gate |
| --- | --- | --- | --- |
| Create route | `assessment.support_refs` MAY appear only in `POST /api/v1/incidents/{incident_id}/views/cartulary.view.assessments.v1/rows`. | Core 01 field and route contract | Existing create/replay tests plus SL-00 action-envelope characterization |
| Omission | Omission MUST mean an empty initial support set. It MUST NOT imply copying from another assessment or an owner-selected default. | Core 01 field defaults | Create-without-support characterization |
| Action envelope | The value MUST be `collection_actions_v1` containing between 1 and 64 actions. Every action MUST be exactly `add_record_ref` plus `linked_record_id`; other operations or metadata MUST fail. | Core 01 wire contract | Decoder boundary and unknown-member tests |
| Target universe | Each target MUST resolve at submission time to an active, visible, same-incident first-class record. Foreign, hidden, deleted, malformed, non-record, or applicable self-reference targets MUST fail atomically. | Core 01 backend target contract; Records visibility port | Owner-port unit tests and service-backed target isolation |
| Relationship routing | The server MUST derive `link_type='supported_by'`, `src_record_id=<new assessment>`, and `dst_record_id=<support target>` from the assessment field. The client MUST NOT provide a link type, direction, table, or storage selector. | Core 01 link mapping; Links participant | Link-direction and forbidden-metadata tests |
| Duplicates | Repeated target IDs in one command MUST coalesce to one active logical support reference before link effects are applied. | Core 01 collection semantics; Assessments command normalization | Duplicate-action characterization |
| Read shape | Every valid support target type MUST remain addressable by `linked_record_id` and `item_ref` in `collection_value_v1`; additive display metadata MUST NOT change identity. | Core 01 view read contract | Heterogeneous read and compatibility tests |
| Existing-row writeability | After first commit, every grid, inspector, row-context, generated-client, conflict-resolution, bulk, or direct API patch targeting `assessment.support_refs` MUST be rejected. | Core 01 patch contract | Exact add/remove/origin rejection tests |
| Excluded owner operations | Delete, restore, rollback, owner-coordinated merge repointing, and incident-bundle reconstruction MUST remain governed by their existing owner contracts and MUST NOT be reclassified as ordinary support-reference patching. | Core 01/Core 02 plus Revisions, Entities, and Incident Bundles owners | Existing contribution and integration evidence |
| Assessment succession | Creating a follow-on assessment MUST create no automatic relation to the prior assessment. In particular, it MUST NOT create a `supersedes` link. | Core 01 link vocabulary; Core 03 workflow | Link and history assertions |

### Create-field defaults and minima

These are current facts that a structural refactor MUST preserve unless an
adopted owner changes them.

| Field or effect | Omitted/default behavior | Required behavior |
| --- | --- | --- |
| `client_txn_id` | No default | MUST be non-empty; exact replay returns the original result and divergent replay returns `client_txn_conflict`. |
| `assessment.subject_ref` | No default | MUST identify the subject record. |
| `assessment.subject_type` | No default | MUST be `host` or `identity` and MUST match the subject. |
| `assessment.assessment_state` | No default | MUST be one of `unknown`, `suspected`, `confirmed`, `disproven`, or `cleared`. |
| `assessment.confidence_score` | Omission or `null` means no score | When present, MUST be an integer from 0 through 100; derived band behavior MUST remain unchanged. |
| `assessment.rationale` | No default | MUST be non-empty after normalization. |
| `assessment.assessor` | Defaults to the authorized actor | A supplied assessor MUST pass the Auth-owner validation port. |
| `assessment.assessed_at` | Defaults to the assessment command clock in UTC | A supplied value MUST be a valid normalized timestamp instant. |
| `assessment.support_refs` | Omission means an empty support set | When present, MUST satisfy the proposed create-only action contract above. |

### Existing-row rejection mapping

| Aspect | Current fact | Proposed owner projection | Required evidence |
| --- | --- | --- | --- |
| Status and error family | Generic patch decoding returns `400 invalid_mutation_payload`. | MUST remain `400 invalid_mutation_payload`. | Direct add and remove requests |
| Error field | Current generic non-writable-field rejection reports `details.field="field_key"`. | MUST report `details.field="assessment.support_refs"`. | Exact envelope assertion |
| Reason code | Current generic rejection reports `unsupported_field_key`. | MUST report `reason_code="unsupported_field_key"`. | Exact envelope assertion |
| Source and link effects | Current non-writable decoding rejects before assessment mutation. | MUST create, remove, restore, or repoint no assessment support link and MUST mutate no assessment source field. | Durable before/after state |
| Revision and projection effects | No successful change should exist for a decoder rejection; direct assessment-specific evidence is incomplete. | MUST create no successful change set or mutation, MUST refresh no projection as a successful change, and MUST persist no successful idempotency result. | Revision, projection, and idempotency counts |
| Collaboration effect | Generic rejection should not publish. | MUST publish no `record_changed` event. | Direct WebSocket negative assertion |

Changing the error detail from the current generic `field_key` value is an
observable contract repair. It is `requires later authorization` and MUST NOT
be bundled into a behavior-preserving facade or ownership slice.

### Proposed inspector projection

| Feature group | Proposed action | Exact required properties |
| --- | --- | --- |
| `assessment.support_refs.manage` | remove | MUST NOT appear in assessment discovery or generated view/protocol projections. |
| `evidence.refs.manage` | remove | MUST NOT remain as a second generic `record_patch` path; the assessment field registry has no independently mutable `evidence.refs` field. |
| `details.read`, `relationships.read`, `history.read` | retain | MUST preserve current panel, non-mutating, role, and result behavior. |
| `record.delete`, `record.restore`, `history.rollback` | retain | MUST preserve current route bindings, roles, confirmation, and result behavior. |
| `assessment.subject_pivot`, `assessment.prior_history` | retain | MUST preserve stable same-shell read/pivot behavior. |
| `create_related.task_request`, `create_related.decision` | retain | MUST preserve current route bindings and seed behavior. |
| `create_related.assessment` | add after owner adoption | MUST use `view_row_create`, owner `view_row_create_route`, role `editor`, `mutates=true`, `requires_confirmation=false`, success `preserve_selected_row`, and failure `show_same_shell_error_invalidate_pending_action`. |

### Follow-on assessment defaults

| Property | Required proposed behavior |
| --- | --- |
| Target schema | `cartulary.view.assessments.v1` |
| Route | Existing view-row create route; no new HTTP route |
| Seed bindings | Selected row `assessment.subject_ref` to target `assessment.subject_ref`, and selected row `assessment.subject_type` to target `assessment.subject_type` |
| Automatically copied values | None beyond the two subject fields |
| Support-reference copying | None; prior references remain readable and MAY be deliberately reselected |
| Remaining minimum signal | A valid assessment state and non-empty rationale are still required; seeded subject context MUST NOT satisfy those minima |
| Relation to prior assessment | None; no `supersedes` or assessment-private successor relation |
| Authorization | Minimum incident role `editor`; route authorization and membership MUST be re-derived at submission |
| Confirmation | None |
| Success | `preserve_selected_row`; the original assessment remains selected and unchanged |
| Failure | `show_same_shell_error_invalidate_pending_action`; the shell and original selection remain available |
| Cancel | MUST commit nothing and MUST leave the original assessment selected and unchanged |

### Current-profile support-picker policy

| Concern | Required current-profile behavior | Owner |
| --- | --- | --- |
| Candidate source | Query only `cartulary.view.timeline.v2` through the existing generic view-query route. The current empty request body MUST continue to use that route's existing pagination, ordering, filtering, and authorization defaults. | Core 03 interaction; Web Workbook |
| Identity | Candidate and selection identity MUST be the Timeline row's stable `record_id`; row index, visible position, timestamp, and label MUST NOT be identity. | Core 03 interaction |
| Controlled selection | Sorting, filtering, refreshing, rerendering, or virtualization MUST NOT retarget an already selected record ID. | Assessment picker policy and shared picker |
| Submission | Stale query results are not authority. The backend MUST revalidate every target at create submission. | Core 01 backend validation |
| Cancellation | Closing or cancelling the picker MUST NOT mutate the draft or persisted assessment. | Core 03 interaction |
| Non-Timeline references | Valid existing non-Timeline references MUST remain readable, displayable, serializable, and preserved across row open, refresh, and follow-on draft creation. The picker MUST NOT convert or discard them. | Core 01 read contract; Core 03 interaction |
| Backend breadth | Timeline-only candidate discovery MUST NOT narrow the backend's record-type-neutral target universe. | Core 01 |
| Forbidden expansion | The current profile MUST NOT add an assessment search route, query every Workbook view, read projection/storage tables directly, or place target policy in grid-vendor APIs. | Core 01/Core 03 boundaries |
| Future expansion | Additional candidate families require separate product adoption and MUST NOT be treated as a prerequisite or blocker for this refactor. | Future Core 03/roadmap owner |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| `NewStore` constructs concrete Auth/idempotency, Records, Links, Revisions, and Projections services | `store.go` constructor | Hidden lifecycle and hard-to-test cross-owner coupling | `must_fix` | Application assessment assembly injects owner ports | Introduce a facade dependency bundle and private repository. |
| Assessment create validates Hosts, Identities, Users, and Records with direct SQL | `store.go` validation queries | Cross-owner model knowledge and duplicated visibility rules | `must_fix` | Entities, Auth, Records | Add narrow caller-transaction validation/read ports before removing SQL. |
| HTTP status, API errors, request route key, and mutation payload live in or beside source coordination | `api.go`, `store.go`, Workbook provider | Transport changes can destabilize source logic and replay behavior | `must_fix` | Workbook/platform adapter over Assessments semantic results | Characterize error/hash/status vectors, then split without wire change. |
| Assessment source package writes `assessment_grid_projection` directly | `projectionprovider/provider.go`; descriptor storage owner is Projections | Storage-owner inversion and future duplicate lifecycle paths | `must_fix` | Projections physical writer; Assessments derivation intent | Add equivalence tests, then move physical DML behind a Projections port. |
| Create/import row presentation duplicates the Projections row facade | `BuildAssessmentRow`, local projection load, `AssessmentRows.LoadTx` | Row-shape drift across create/query/import | `must_fix` | One canonical assessment row derivation/load path | Prove row JSON parity before consolidation. |
| Broad `*assessments.Store` is passed to Workbook and Entities | Application catalogs and Entities merge adapter | Callers can depend on unrelated capabilities | `should_fix` | Assessment facade and narrow merge-effects port | Migrate one caller family at a time; remove broad exposure after parity. |
| Assessment frontend types and payload builder reside in the Timeline model | Frontend imports and tests | Semantic ownership leak and misleading dependency | `should_fix` | Assessment Workbook model | Move types/helpers only; retain payload bytes and selectors. |
| Assessment browser fixture uses a Timeline-named mutation helper | Assessment Playwright fixture | Test support obscures semantic ownership | `should_fix` | Generic Workbook or assessment test support | Rename/move only if selected with exact test accounting. |
| Assessment source SQL and record-type providers remain assessment-owned | Source table writes and narrow provider contributions | Low when kept narrow | `intentional/no_action` | Assessments | Preserve these capabilities while changing construction boundaries. |
| Authorization is enforced by the generic Workbook/import/revision route owners | Route/provider inspection; no assessment-specific handler | Moving it inward would duplicate route policy | `intentional/no_action` | Workbook/import/revision owners | Characterize and preserve; do not add assessment-local route authorization. |
| Assessment UI imports `@cartulary/grid-adapter` and no vendor grid directly | `AssessmentWorkbookSurface.tsx` import scan | Low | `intentional/no_action` | Grid adapter and Web Workbook | Retain the current adapter boundary. |
| Assessment test support uses direct fixture setup but has no production import | Inbound import search | Could diverge after facade movement, but does not leak into production | `intentional/no_action` | Assessments test support | Re-audit after facade changes; do not force runtime APIs into fixtures. |
| Frontend support selection currently queries Timeline only | `useAssessmentSupportRows.ts`; resolved current-profile disposition in §4 | Low when kept explicit; expansion would change visible behavior and query traffic | `intentional/no_action` | Core 03 interaction and Assessment Workbook policy | Preserve Timeline-only discovery, broad backend admission, and non-Timeline read compatibility. Route expansion to separate product adoption. |
| Core 01 declares both create-only support refs and a required mutating support-ref inspector action | REQ-01-332/335 and §7.4.1A | The owner corpus is self-contradictory and generated discovery advertises an unusable mutation | `must_fix` | Core 01 owner, with Core 03 interaction and Core 04 acceptance projections | Keep current state `BLOCKED: owner contradiction`; adopt the RB-001 disposition before changing discovery or error details. |
| Current generic patch rejection identifies the request member as `field_key`, while the proposed repaired contract identifies `assessment.support_refs` | `decodePatchChange`; proposed disposition in `temp/analysis-notes.md` | Observable error-detail drift if smuggled into structural work | `defer` | Core 01 error contract and Workbook projection | Treat as `requires later authorization`; characterize both current and target envelopes and project only after adoption. |
| Generated assessment surfaces could drift if moved code is mistaken for a contract change | Generated policy and contract manifests | Hand-edit or unnecessary regeneration risk | `intentional/no_action` | Authored contract and codegen owners | Structural slices expect no public generated diff; use Make-owned generation only if inputs change. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | None | WF-01 | Freeze scope, authority, revision, and planning-only posture | This tracker and owner documents | `make lint-markdown` for tracker | Tracker records source hierarchy and write boundary. |
| WF-01 | Target inventory | chain | WF-00 | WF-02 | Account for every target file, caller, dependency, and test | `internal/modules/assessments/**`; caller catalogs | Inventory/search audit | All 18 files have explicit rows. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-04 | Map observable behavior to adopted owners and live evidence | Core owners, view/OpenAPI/import/projection contracts | Contract-to-code/test review | Every discovered contract has owner and test posture. |
| WF-03 | Current and released-baseline characterization | parallel | WF-02 | WF-05 | Freeze current runtime behavior, side-effect absence, and the v1 compatibility basis before any owner or structural movement | Assessment, Workbook, Revisions, Imports, Entities, Collaboration tests; indexed `1.0.0` baseline | `make task-guide ROLE=module-author OWNER=module.assessments`; owner slices; static release inspection | Every selected implementation slice has a prerequisite test gate and the version decision cites evidence. |
| WF-04 | Boundary and coupling scan | parallel | WF-02 | WF-05 | Classify peer SQL, transport leakage, projection ownership, frontend seams, and intentional adapters | Store, providers, application assembly, frontend model | Boundary checks and exact import scans | Every finding has classification and proposed owner. |
| WF-05 | Owner adoption package | chain | WF-03, WF-04 | WF-06 | Repair Core 01, specify Core 03 interaction, add Core 04 acceptance, close traceability, and project the compatibility decision | Core 01/Core 03/Core 04 and authored traceability/contracts in a later authorized task | `make lint-markdown`; owner/projected contract checks; Make-owned generation/drift | One internally consistent adopted owner contract exists before any repaired behavior is implemented. |
| WF-06 | Implementation sequencing | chain | WF-05 | WF-07 | Implement facade/ports, canonical projection ownership, narrow callers, frontend ownership, picker policy, and follow-on behavior in dependency order | Assessments, application assembly, Projections, Records, Links, Entities, Revisions, Workbook web | Per-slice commands in §7 | Each slice satisfies its protecting `AMR-AC-*` rows and rollback checkpoint. |
| WF-07 | Harness/test/accounting update plan | chain | WF-06 | WF-08 | Update authored verification/boundary inputs only when identities or imports move | Authored boundary/test-family inputs; generated topology via Make | `make generate-drift`; policy/shape/boundary checks | No stale selector and no hand-edited generated file. |
| WF-08 | Validation and final handoff | chain | WF-07 | None | Run focused-to-broad gates and leave current continuation evidence | Changed slice plus tracker | `make agent-finalize`; `make check` after focused gates | Results, failures, run roots, blockers, and next action are recorded. |

The required dependency graph is:

`WF-00 → WF-01 → WF-02 → {WF-03, WF-04} → WF-05 → WF-06 → WF-07 → WF-08`

WF-00 through WF-04 are complete as planning work. WF-05 through WF-08
remain later authorized adoption and implementation work. WF-06 MUST NOT
implement the repaired support-reference or inspector behavior before WF-05
passes.

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| SL-00 | None | Add current and indexed-release characterization before changing an owner or implementation | Assessment, Workbook, Revisions, Imports, Entities, Collaboration, inspector, and frontend tests; indexed `1.0.0` source | Freezing an accidental implementation detail or misclassifying v1 compatibility | All §12 current-behavior cases, including exact current patch error details, absence of patch side effects, create publication, import parity, and historical non-functioning inspector action | `make test-slice OWNER=module.assessments`; `make service-backed-test-slice OWNER=module.assessments` | Revert only tests that exceed the current owner; retain factual release notes in the tracker | Current behavior is executable, the indexed baseline is documented, and each later slice names protecting AC rows. |
| SL-01 | SL-00 | `requires later authorization`: amend Core 01 create/patch/inspector/target rules, Core 03 follow-on/picker interaction, Core 04 binary acceptance, and traceability | Core 01, Core 03, Core 04, Appendix F or adopted traceability inputs | Owner contradiction, error-detail compatibility, accidental tracker-as-owner duplication | Owner self-consistency, REQ-to-AC completeness, exact defaults and mappings in §§4 and 12 | `make lint-markdown`; applicable owner conformance targets discovered at execution time | Revert the owner amendment as one adoption package; do not leave partial cross-Core rules | Adopted owners contain one non-contradictory contract, identify the v1 compatibility decision, and map every proposed rule to binary acceptance. |
| SL-02 | SL-01 | `requires later authorization`: update authored assessment view/inspector projections, regenerate through Make, and remove contradictory generated patch capability | Authored view-schema/contract inputs and Make-owned generated Go/TypeScript outputs | Discovery membership, error details, compatibility, generated drift | Exact feature-set uniqueness, route/role/seed/result defaults, protocol/view-schema parity | `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; conditional `make openapi-compatibility-check` | Revert authored inputs and their reproducible outputs together; never hand-edit generated files | Generated discovery omits both forbidden manage features, includes exact follow-on configuration, retains the adopted schema version, and all drift/policy checks pass. |
| SL-03 | SL-00 | Introduce an assessment facade with private source persistence and injected Records, Auth, Links, Revisions, Projections, and idempotency ports; replace cross-owner SQL | Assessments root, app assembly, Records/Auth/Links/Revisions/Projections adapters, Workbook provider | Create status/error/hash/replay, target visibility, assessor eligibility, transaction ownership | Existing create/replay/authorization tests; target matrix; link failure rollback; standalone versus caller-transaction assertions | Assessment and affected owner slices; `make backend-module-boundary-check` | Migrate one dependency family per checkpoint; keep the old constructor only until the complete facade caller cutover passes | Workbook maps semantic results to unchanged HTTP envelopes; standalone create owns one transaction; import participants do not commit; assessment code reads peer state only through typed ports. |
| SL-04 | SL-03 | Centralize canonical assessment source-to-row derivation and move physical projection-table DML/lifecycle to Projections | Assessment projection provider/query surfaces, Projections services/storage, facade/import refresh | Row JSON, band/count derivation, refresh timing, rebuild determinism, saved-view compatibility | Create/query/import row equivalence; refresh, rebuild, sort/filter, and saved-view tests | Assessment and Projections owner slices; `make backend-module-boundary-check`; `make generate-drift` | Keep the old writer until row and rebuild equivalence pass; revert the projection seam as one checkpoint | No assessment production path writes physical projection storage directly; Assessments retains derivation/query intent and canonical rows remain equivalent. |
| SL-05 | SL-03, SL-04 | Expose narrow assessment merge effects; migrate Workbook, Imports, Revisions, portability, recovery, Entities, and remaining callers; retire broad `Store` exposure without forwarding shims | Assessment contributions and application catalogs; Entities merge adapter | Same-change-set merge effects, import target identity, provider catalogs, bundle path, recovery presence | Merge protected-set/repoint/projection tests; import, revision, portability, recovery, and caller parity | Affected owner slices; service-backed assessment slice; backend boundary check | Roll back by caller family before removing the broad surface | Entities consumes only the merge-effects port; no production caller receives `assessments.Store`; transaction and contribution identities remain unchanged. |
| SL-06 | SL-01, SL-03 | Move assessment DTOs/helpers from Timeline ownership, codify the Timeline candidate source, preserve non-Timeline reads, and implement the adopted follow-on create action | Assessment frontend model/hook/component/tests, Workbook shell/shared picker, Timeline model | Payload bytes, draft retention, stable selection, stale target, inspector routing, focus, cancellation, non-Timeline loss | Frontend and browser cases in §12, including exact seeds/defaults, record-key stability, authorization loss, same-shell failure, and original-row preservation | `make frontend-unit`; `make frontend-import-boundary-check`; `make browser-e2e-webserver-backed` | Separate structural ownership movement from adopted interaction activation; revert each checkpoint without changing wire payloads | Assessment owns its DTOs/policy, Timeline-only selection is explicit, no vendor-grid policy leaks, and the follow-on action satisfies every adopted default. |
| SL-07 | SL-02, SL-05, SL-06 | Conditionally update authored boundary/test-accounting inputs for changed imports or executable identities; regenerate topology only through Make | Authored verification, test-family, boundary, and topology inputs only when selected moves require them | Stale selectors, duplicate execution, generated hand edits, architecture inferred from harness rows | Exact selector occurrence, owner/collaborator parity, generated policy and JSON shape | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; both boundary checks | Revert authored accounting with its owning code/contract slice; generated outputs MUST remain reproducible | Every changed executable identity selects exactly once, no harness row defines runtime architecture, and generated drift/policy checks pass. |
| SL-08 | SL-07 | Run focused-to-broad validation and finalize the handoff | Complete later-authorized diff and this tracker | Undiscovered cross-owner, security, compatibility, or browser regression | All affected owner, service-backed, frontend, browser, boundary, drift, and acceptance evidence | `make test-fast`; `make agent-finalize`; `make check` after focused gates | Roll back to the last passing slice checkpoint rather than weakening an owner or verifier | Required gates pass, or the tracker records the exact failed target/artifact and remains blocked. |

SL-01, SL-02, and the follow-on/error-detail portion of SL-06 are
`requires later authorization`. Behavior-preserving facade, projection, port,
caller, and frontend-ownership work MUST NOT implicitly implement those
contract repairs. Broader-than-Timeline picker expansion is outside these
slices.

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make test-slice OWNER=module.assessments` | Authored non-service assessment owner rows | yes | Establish and rerun the narrow owner baseline for each backend slice; select exact rows after `make explain-test-owner`. |
| integration | `make service-backed-test-slice OWNER=module.assessments` | Store, integration, browser-backed, and frontend rows selected for the owner | yes | Required before changing transaction, projection, import, revision, inspector, or publication seams. |
| e2e/browser | `make browser-e2e-webserver-backed` | Full webserver-backed browser set | yes for SL-06; no for backend-only slice start | Run focused owner rows first, then the public browser target when frontend behavior, inspector discovery, or route composition moves. |
| generated drift | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check` | Generated outputs, policy, and authored JSON shape | no | Structural slices expect no semantic contract diff; SL-02 expects only the adopted, reproducible inspector projection diff. |
| import-boundary/static | `make backend-module-boundary-check`; `make frontend-import-boundary-check` | Backend owner imports and frontend architecture imports | yes | Baseline and post-slice checks for Store, projection, and Timeline-model decoupling. |
| full check | `make check` | Repository developer gate | no | Run after focused evidence and `make agent-finalize`, not as the first diagnostic. |
| documentation | `make lint-markdown` | This tracker and repository Markdown | yes for tracker completion | The only validation target required by this planning-only write. |

For this tracker-only revision, the sole required command is
`make lint-markdown`. Product, integration, browser, drift, boundary, and
full-check targets MUST NOT be reported as run unless a later authorized task
actually runs them.

`make frontend-unit`, `make test-fast`, and
`make browser-e2e-webserver-backed` are broader checkpoints after the narrow
owner slices. `make generate` is required only after a later authorized
authored-contract change. `make openapi-compatibility-check` is additionally
required only when that change affects an authored OpenAPI input.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| AS-001 | Freeze target, authority, revision, and write boundary | WF-00 | DONE | None | §1 | Scope and source hierarchy are explicit. |
| AS-002 | Inventory every assessment target file and caller | WF-01 | DONE | AS-001 | §2, 18 inventory rows | Every target file is accounted for. |
| AS-003 | Map public behavior and owner contracts | WF-02 | DONE | AS-002 | §4 | Every discovered contract has an owner and test posture. |
| AS-004 | Classify package boundary and coupling findings | WF-04 | DONE | AS-003 | §§3 and 5 | Each finding has classification, owner, and planning action. |
| AS-005 | Add current and indexed-release assessment characterization | WF-03 | TODO | AS-003 | Planned SL-00; static `1.0.0` evidence in §1 | Direct patch/add/remove, side-effect absence, WebSocket, import, merge, and contribution evidence passes. |
| AS-006 | Introduce facade/private persistence/injected dependencies | WF-06 | TODO | AS-005, AS-012 | Planned SL-03 | Workbook consumes semantic results; standalone and caller-transaction paths preserve their distinct ownership. |
| AS-007 | Correct projection source/storage boundary and canonical row path | WF-06 | TODO | AS-006 | Planned SL-04 | Projections owns physical storage and canonical row equivalence passes. |
| AS-008 | Replace cross-owner validation SQL with typed ports | WF-06 | TODO | AS-005, AS-012 | Planned SL-03 | Peer state is accessed through approved caller-transaction ports without disclosure or independent commit. |
| AS-009 | Narrow the Entities merge effect dependency | WF-06 | TODO | AS-006, AS-007 | Planned SL-05 | Entities consumes only the assessment merge-effects port and same-change-set behavior passes. |
| AS-010 | Migrate all contribution/caller families and retire broad Store | WF-06 | TODO | AS-007, AS-008, AS-009 | Planned SL-05 | No production broad Store caller remains and contribution identities are unchanged. |
| AS-011 | Move assessment frontend semantics out of Timeline and implement adopted interaction | WF-06 | TODO | AS-005, AS-012 | Planned SL-06 | Assessment model owns DTOs/policy; Timeline-only and follow-on acceptance pass. |
| AS-012 | Adopt the decision-complete support-reference and inspector owner repair | WF-05 | TODO | AS-005, RB-001 | Planned SL-01; §§4 and 11 | Core 01, Core 03, Core 04, compatibility, and traceability form one adopted, non-contradictory contract. |
| AS-013 | Update authored contract, boundary, and test-accounting inputs conditionally | WF-07 | TODO | AS-010, AS-011, AS-018 | Planned SL-07 | Exact selectors/import policies match the final shape and generated outputs are reproducible. |
| AS-014 | Complete focused-to-broad validation and final handoff | WF-08 | TODO | AS-013 | Planned SL-08 | Required gates and current handoff evidence are recorded. |
| AS-015 | Expand supporting-record picker beyond Timeline | WF-06 | DROPPED | RB-002 | §4 current-profile policy | Current refactor retains Timeline-only discovery; any expansion is separate future product adoption. |
| AS-016 | Validate this revised planning tracker | WF-00 | DONE | AS-001 | `make lint-markdown` passed in this revision session | Markdown lint passes and the new session rows record the result. |
| AS-017 | Characterize indexed release metadata and source statically | WF-03 | DONE | AS-003 | §1 historical-baseline evidence | The sole indexed baseline, emitted features, patch rejection, and missing frontend handler are recorded without claiming an executable release run. |
| AS-018 | Project the adopted view/inspector contract through Make-owned generation | WF-05 | TODO | AS-012 | Planned SL-02 | Authored and generated discovery contain the exact adopted feature set and all drift/policy gates pass. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30T12:00:26-04:00 | Codex planning-only tracker session | Target exists; authority and planning-only boundary are frozen | Inspected Core 00-04, domain, framework, target and current callers; touched only this tracker | `sed`, `rg`, `find`, `jq`, Git revision/status commands; `make lint-markdown` | New tracker created; Markdown lint passed at `.cartulary/test-results/20260730T160507Z-p247418` | RB-001 affects only later support-reference behavior work | Seek later authorization for behavior-preserving SL-00. |
| 2026-07-30T12:47:57-04:00 | Codex NLSpec-rigor tracker revision | Current facts, proposed dispositions, and adoption gates are separated; the sole-file boundary remains in force | Inspected NLSpec guidance, analysis notes, Core 01/Core 03, release index/source, patch decoder, target/callers/contracts/frontend; touched only this tracker | `sed`, `rg`, `find`, `jq`, Git status/revision/log/show/merge-base commands; filtered `make help-all`; `make lint-markdown` | RB-001 and RB-002 dispositions are decision-complete; Markdown lint passed; no owner or implementation file changed | Live Core remains `BLOCKED: owner contradiction` until the proposed repair is adopted | Begin SL-00 only in a later authorized task. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30T12:00:26-04:00 | Codex planning-only tracker session | Assessments is a legitimate source owner with mixed facade, projection, transport, and persistence responsibilities | Inspected all 18 target files, app assemblies, Workbook, Entities, Projections, Revisions; touched only this tracker | Exact import/caller/symbol searches; source reads | Store, peer-SQL, projection-DML, row-duplication, and narrow-port findings classified | No blocker for characterization/facade planning; RB-001 blocks support-ref behavior change | Begin SL-00 only under later authorization. |
| 2026-07-30T12:47:57-04:00 | Codex NLSpec-rigor tracker revision | The target interface map now distinguishes standalone facade transaction ownership from import, merge, revision, and portability caller-transaction participation | Re-inspected assessment store/import APIs, Workbook adapter, peer ports, projection and revision contributions; touched only this tracker | Exact symbol and transaction-flow searches; source reads | Inputs, semantic results, commit/publication prohibitions, and owner boundaries are explicit | Owner adoption gates repaired support behavior, not current characterization | Characterize the interfaces in SL-00 before executing SL-03 through SL-05. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30T12:00:26-04:00 | Codex planning-only tracker session | Dedicated assessment surface is contract/grid-adapter based; assessment DTO/helper ownership leaks into Timeline | Inspected assessment component/model/hooks/tests/browser fixture; touched only this tracker | Frontend import and symbol searches; exact file reads | Structural SL-06 defined; current Timeline-only support picker frozen | RB-002 blocks picker expansion, not helper movement | Move DTOs/helpers only in a later authorized SL-06. |
| 2026-07-30T12:47:57-04:00 | Codex NLSpec-rigor tracker revision | Timeline-only discovery is resolved for the current profile while backend admissibility and non-Timeline reads remain broader | Re-inspected support-row hook, assessment model/surface, Timeline model, inspector rendering, frontend tests and browser spec; touched only this tracker | Frontend symbol/import searches and exact source reads | Stable `record_id` identity, cancellation, stale validation, non-Timeline preservation, and follow-on defaults are explicit | Broader picker expansion is dropped from this refactor; adopted Core 03 behavior is still required for follow-on activation | Execute structural ownership and adopted interaction checkpoints separately within SL-06. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30T12:00:26-04:00 | Codex planning-only tracker session | HTTP, view, projection, import, revision, portability, and generated surfaces are frozen | Inspected authored assessment contracts, registries, generated projections, and owner sections; touched only this tracker | `jq`, `rg`, `sed`; `make explain-target` discovery | No contract or generated file changed; conditional Make-owned generation plan recorded | RB-001 is `BLOCKED: owner contradiction` | Do not change assessment support-ref contracts until the owner is resolved. |
| 2026-07-30T12:47:57-04:00 | Codex NLSpec-rigor tracker revision | Proposed create-only support contract, exact patch repair, inspector matrix, v1 compatibility posture, and owner routing are mapped | Inspected Core 01 registry/defaults, Core 03 append-only rule, assessment authored/generated view contracts, release index and historical source; touched only this tracker | `rg`, `sed`, `jq`, Git history/source inspection | Proposed error detail is explicitly classified as a later-authorized observable repair; no contract or generated file changed | `RESOLVED_PENDING_ADOPTION` does not clear the live owner contradiction | Adopt SL-01, then project SL-02 through Make; never hand-edit generated output. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30T12:00:26-04:00 | Codex planning-only tracker session | Six active `module.assessments` rows and existing cross-owner tests are mapped; gaps are explicit | Inspected verification owner, test family/catalog, scenario inventory, target tests; touched only this tracker | `make task-guide ROLE=module-author OWNER=module.assessments`; `make explain-test-owner OWNER=module.assessments`; `make explain-target`; filtered `make help-all`; `make lint-markdown` | Discovery completed and Markdown lint passed; no product, integration, browser, drift, or full-check validation was run | Missing characterization is SL-00 work, not a planning blocker | Run owner tests only in a later implementation task. |
| 2026-07-30T12:47:57-04:00 | Codex NLSpec-rigor tracker revision | Stable `AMR-AC-*` rows map create, patch, inspector, frontend, security, contribution, and generation obligations | Inspected assessment/Workbook tests, verification routing, current Make target inventory, and release source tests; touched only this tracker | Test/symbol searches; filtered `make help-all`; `make lint-markdown` | Current-versus-proposed evidence is separated and Markdown lint passed; no product, integration, browser, drift, boundary, or full-check target was run | SL-00 executable characterization remains required before adoption or implementation | Later run focused owner evidence before broad gates. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30T12:00:26-04:00 | Codex planning-only tracker session | Workbook/import/revision route owners retain authorization; assessment source code does not invent route policy | Inspected Workbook routes/scenario inventory and Core 04 assessment criteria; touched only this tracker | Route/auth searches and exact source reads | Editor/reviewer/admin create, member query, CSRF, and visibility outcomes are frozen | None for behavior-preserving structure | Preserve outer-owner authorization and rerun route scenarios after facade migration. |
| 2026-07-30T12:47:57-04:00 | Codex NLSpec-rigor tracker revision | Route owners retain authorization; target validation and picker discovery now have explicit non-disclosure and revalidation obligations | Re-inspected Workbook create route, assessment validation, Core 04 authorization rules, picker query path, and relevant tests; touched only this tracker | Route, CSRF, membership, target-validation, and egress searches | Editor minimum, submission-time membership/capability revalidation, stale-target rejection, and no-event/no-egress outcomes are testable | Hidden-target semantics require the adopted Records visibility port rather than assessment-owned SQL | Cover the security rows in SL-00, SL-03, and SL-06 before final validation. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30T12:00:26-04:00 | Codex planning-only tracker session | Planning inventory, diagnosis, workflows, slices, validation, and handoff are decision-complete | Inspected current target/callers/contracts/tests/frontend and historical handoffs; touched only this tracker | Discovery commands listed above; `make lint-markdown` | Tracker validation passed; implementation remains unauthorized and no production refactor occurred | RB-001; RB-002 only for deferred behavior expansion | Resolve RB-001 separately or authorize behavior-preserving SL-00. |
| 2026-07-30T12:47:57-04:00 | Codex NLSpec-rigor tracker revision | Both planning questions have decision-complete dispositions; adoption and executable evidence remain the only gates | Inspected all decision sources and live mismatches named above; touched only this tracker | Discovery commands listed above; `make lint-markdown` | AS-012 is owner-adoption work; AS-015 is dropped for the current profile; Markdown lint passed; implementation remains unauthorized | Live Core 01 contradiction, SL-00 evidence, and later owner adoption | Next authorized session starts with SL-00, then SL-01. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | `BLOCKED: owner contradiction` remains the current corpus fact: Core 01 REQ-01-332/335 makes `assessment.support_refs` create-only, while §7.4.1A emits mutating `assessment.support_refs.manage` and `evidence.refs.manage`. Proposed disposition: keep support references create-only, remove both patch features, and add exact `create_related.assessment` follow-on behavior from §4. | Runtime code, tests, or this tracker cannot select between contradictory owner statements. The proposed error detail and inspector membership are observable contract repairs. | SL-00 executable characterization; adopted Core 01/Core 03/Core 04 and traceability package; faithful authored/generated projection; explicit v1 compatibility confirmation | `RESOLVED_PENDING_ADOPTION` |
| RB-002 | The current-profile interactive picker remains Timeline-only; backend target admission remains record-type-neutral, and existing non-Timeline references remain readable and preserved. Broader selection is separate future product work. | This prevents an architecture-expanding search route or all-view fan-out while retaining wire compatibility and backend semantics. | Core 01 backend-target clarification, Core 03 current-profile interaction, Core 04 evidence, and frontend characterization when the owner package is adopted | `RESOLVED_CURRENT_PROFILE` |

No planning-choice question remains open. RB-001 remains an adoption and
implementation gate because the owner corpus has not yet changed. It MUST NOT
block SL-00 characterization, but under the workflow selected in §6, WF-06
implementation begins only after WF-05 adopts the complete repair. RB-002 is
not a refactor blocker; AS-015 is dropped for the current profile.

## 12. Binary Completion Criteria

The `AMR-AC-*` identifiers are tracker-local acceptance identifiers. They are
not adopted Core 04 identifiers and MUST be mapped to adopted owner criteria
before a contract-repair slice is implemented.

| ID | Posture | Scenario | Binary pass condition | Required evidence owner/slice |
| --- | --- | --- | --- | --- |
| AMR-AC-001 | Current fact | Create without support references | Create succeeds with an empty `collection_value_v1`; no support link exists. | Assessments, Links / SL-00 |
| AMR-AC-002 | Current fact plus proposed completeness | Create with one, multiple heterogeneous, and duplicate support targets | Every valid target is returned by stable record identity; duplicates produce exactly one active logical link. | Assessments, Records, Links / SL-00, SL-03 |
| AMR-AC-003 | Proposed disposition | Invalid support target matrix | Foreign-incident, hidden, deleted, malformed, non-record, and applicable self-reference cases return `400 invalid_mutation_payload` before any assessment or link commits. | Core 01, Assessments, Records / SL-00, SL-01, SL-03 |
| AMR-AC-004 | Current fact plus proposed precision | Support action envelope | Empty actions, more than 64 actions, non-`add_record_ref` operations, malformed IDs, unknown action members, or client routing metadata are rejected; 1 through 64 valid actions are admitted and normalized. | Core 01, Workbook decoder / SL-00, SL-01 |
| AMR-AC-005 | Current fact | Link derivation | Every committed support reference is `supported_by` from the new assessment record to the validated support record; no client value selects direction or link type. | Assessments, Links / SL-00 |
| AMR-AC-006 | Current fact | Create replay | Exact replay returns the original assessment with `200` and creates no duplicate record or link; reuse with a different support set returns `409 client_txn_conflict`. | Workbook, Assessments, idempotency / SL-00, SL-03 |
| AMR-AC-007 | Preservation | Failure after a participant starts | An injected validation, link, revision, or projection failure rolls back assessment source, record envelope, links, revision state, projection state, and idempotency success as one transaction. | Assessments and participants / SL-00, SL-03 |
| AMR-AC-008 | Current fact | Existing-row support add/remove characterization | Both requests fail with `400 invalid_mutation_payload`, current `details.field="field_key"`, and `reason_code="unsupported_field_key"`. | Workbook / SL-00 |
| AMR-AC-009 | Proposed disposition | Adopted existing-row rejection envelope | After owner adoption and projection, both requests fail with `400 invalid_mutation_payload`, `details.field="assessment.support_refs"`, and `reason_code="unsupported_field_key"`. | Core 01, Workbook / SL-01, SL-02 |
| AMR-AC-010 | Proposed disposition | Rejected patch side effects | Add/remove rejection changes no assessment source, link, change set, mutation, projection, successful idempotency result, row version, history item, or `record_changed` stream. | Workbook, Assessments, Links, Revisions, Projections, Collaboration / SL-00, SL-01 |
| AMR-AC-011 | Proposed disposition | Inspector discovery | Assessment discovery contains each retained feature exactly once, contains neither forbidden manage feature, and contains one `create_related.assessment` with the exact route, owner, role, mutation, confirmation, seed, success, and failure values from §4. | Core 01, view schema / SL-01, SL-02 |
| AMR-AC-012 | Proposed disposition | Follow-on draft initialization | The action opens a new assessment draft seeded only with subject reference and type; state, confidence, rationale, assessor, time, and support references are not copied, and state plus non-empty rationale remain required. | Core 03, Web Workbook / SL-01, SL-06 |
| AMR-AC-013 | Proposed disposition | Follow-on cancel, reject, and success | Cancel commits nothing; rejection preserves the same shell and original selection; success creates a distinct assessment, preserves the original selection and history, and creates no prior-assessment relation. | Core 03, Workbook, Assessments / SL-01, SL-06 |
| AMR-AC-014 | Preservation plus proposed authorization | Capability and membership changes | Submission re-derives incident membership and minimum `editor` role; viewer, non-member, deployment-admin-only, or capability-lost callers cannot create or infer hidden target existence. | Core 04, Workbook, Records / SL-00, SL-03, SL-06 |
| AMR-AC-015 | Current fact | Collaboration publication | A newly committed assessment publishes the ordinary new-record `record_changed` intent exactly as owned; exact replay and every rejected create/patch publish nothing. | Revisions, Collaboration / SL-00 |
| AMR-AC-016 | Preservation | History and source-owner operations | Delete, restore, rollback, merge repointing, and follow-on creation preserve append-only assessment history and do not create ordinary support-reference patch capability. | Revisions, Entities, Assessments / SL-00, SL-05 |
| AMR-AC-017 | Preservation | Import, portability, recovery, and revision contributions | `module.assessments@1`, `assessments.import_create`, `data/compromise_assessments.ndjson`, recovery presence, provider identities, caller-transaction behavior, and revision/live-change policy remain unchanged. | Imports, Incident Bundles, Recovery, Revisions / SL-00, SL-05 |
| AMR-AC-018 | Preservation | Canonical row and projection lifecycle | Create, query, import, refresh, and rebuild return equivalent assessment rows; saved-view filter/sort behavior is unchanged; no assessment production path performs physical projection DML after cutover. | Assessments, Projections, Workbook / SL-04 |
| AMR-AC-019 | Proposed current-profile policy | Picker candidate universe | Only visible same-incident Timeline rows are returned as interactive candidates through the existing query route; no all-view query, new route, storage read, or external request occurs. | Core 03, Web Workbook / SL-01, SL-06 |
| AMR-AC-020 | Proposed current-profile policy | Candidate identity and stale results | Selection remains bound to `record_id` after sort, filter, refresh, rerender, and virtualization; a target that becomes invalid before submit is rejected without silently retargeting or discarding the draft. | Core 03, Records, Web Workbook / SL-03, SL-06 |
| AMR-AC-021 | Proposed current-profile policy | Existing non-Timeline references | Every valid non-Timeline reference returned in `collection_value_v1` remains visible and is not discarded, converted, or automatically copied when the row is opened, refreshed, or used for follow-on creation. | Core 01/Core 03, Web Workbook / SL-01, SL-06 |
| AMR-AC-022 | Boundary acceptance | Transaction and interface ownership | Standalone facade owns exactly one transaction; import, merge, revision, portability, Records, Links, and Projections participants use the supplied transaction and never commit or publish independently. | Assessments and caller owners / SL-03, SL-05 |
| AMR-AC-023 | Boundary acceptance | Frontend ownership | Assessment DTOs and candidate policy are assessment-owned; the shared picker is controlled presentation; assessment code imports only the grid adapter and no vendor-grid policy API. | Web Workbook, grid adapter / SL-06 |
| AMR-AC-024 | Projection and accounting acceptance | Generated and verification integrity | Authored owner inputs project through Make; generated files are not hand-edited; each changed executable identity selects exactly once; boundary, drift, policy, and JSON-shape checks pass. | Contract/codegen and verification owners / SL-02, SL-07, SL-08 |

The tracker revision is complete only when:

- [x] Every file in `internal/modules/assessments` is inventoried; no target file
  is silently out of scope.
- [x] Every discovered public contract risk has an owner and current/proposed
  evidence posture.
- [x] Every proposed workflow has dependencies and a handoff exit criterion.
- [x] Every implementation slice names its authorization posture, dependencies,
  rollback boundary, validation, and binary completion condition.
- [x] Explicit defaults, transaction ownership, interface boundaries, error
  mappings, publication behavior, and exclusions are recorded.
- [x] Validation commands are discovered through current Make-owned task
  guidance.
- [x] The live support-reference conflict is marked
  `BLOCKED: owner contradiction`; its disposition is
  `RESOLVED_PENDING_ADOPTION` rather than falsely presented as adopted.
- [x] The Timeline-only picker decision is
  `RESOLVED_CURRENT_PROFILE`; broader selection is not a refactor blocker.
- [x] The planning-framework/live-repository mismatch is recorded and the live
  adopted owner is followed.
- [x] The seven handoff areas preserve prior history and contain a new revision
  session entry sufficient for another agent to continue without rediscovery.
- [x] `make lint-markdown` passes for this revised tracker.

Planning completion does not authorize or claim completion of the refactor.
