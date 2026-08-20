# assessments Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Item | Current posture |
| --- | --- |
| Target path | internal/modules/assessments |
| Target label | assessments |
| Output path | docs/handoffs/assessments-module-refactor-tracker.md |
| Repository baseline | Historical remediation baseline: main at 8c8c52a43069; APR planning baseline: main at 6cb809190bf107e805720a9552a894783db10cbc on 2026-08-19 |
| Status | AS-S00A through AS-S06, the Sections 14-15 follow-up, and APR-S00 through APR-S05 are COMPLETE; the APR iteration is COMPLETE |
| Allowed change | APR-S01 through APR-S05 are authorized by the 2026-08-19 implementation request and must follow the serial checkpoint protocol in Section 18 |
| Non-goals | APR-S00 makes no product, contract, migration, generated-artifact, test-catalog, or other documentation change. The planned iteration does not change public routes, OpenAPI, view schemas, bundle formats, frontend behavior, or public protocols |
| Execution authority | The 2026-08-19 implementation request authorized APR-S01 through APR-S05; all serial dependency checkpoints are complete |

The target began with 21 files. Authorized additions and relocations are added
to the inventory as they land. The normalized label is
assessments: lowercase kebab case with no spaces, path separators, shell
metacharacters, or unsafe filename characters. This tracker preserves current
observable behavior by default and separates structural refactoring from the
one discovered normative correction.

### Source hierarchy

1. Adopted subsystem NLSpecs govern only their named subsystem. No
   assessment-specific adopted subsystem NLSpec was found.
2. Core 00 through Core 04 govern current implementation-conformance behavior.
3. Core 05 applies only to claim-bearing timed or fixture-sensitive
   publication. No such publication is in this target, so Core 05 is not
   applicable to this refactor.
4. Domain vocabulary and implementation-support guides govern terminology,
   boundaries, harness mechanics, and execution support.
5. Current code and tests describe current repository state.
6. Prior plans, handoffs, and the planning framework are evidence and planning
   doctrine, not proof of current state.

No owner-document contradiction was discovered. The assessor finding below is
a repository-versus-owner conformance mismatch, not an owner contradiction.

### Owner and supporting documents inspected

- docs/handoffs/cartulary_modular_refactor_planning_framework.md
- docs/domain.md
- docs/research/nlspec-spec.md
- docs/spec/00_document_set_status_and_precedence.md
- docs/spec/01_architecture_storage_and_view_contracts.md
- docs/spec/02_domain_model_schema_and_history.md
- docs/spec/03_workbook_interaction_collaboration_and_workflows.md
- docs/spec/04_security_deployment_and_conformance.md
- docs/spec/05_claim_publication_and_benchmark_reproducibility.md
- docs/decisions/projections-module-boundary.md
- docs/decisions/revisions-module-boundary.md
- docs/spec/I_projection_authority_boundary_and_characterization.md
  (non-normative supporting evidence only)

### Repository files inspected

- Every file under internal/modules/assessments, listed individually in
  Section 2.
- Composition and callers under internal/app/assessmentassembly,
  internal/app/importassembly, internal/app/incidentportabilityassembly,
  internal/app/projectionassembly, internal/app/recoveryassembly,
  internal/app/revisionassembly, internal/app/timelineassembly, and
  internal/app/workbookassembly.
- Relevant owner implementations in internal/modules/workbook,
  internal/modules/projections, internal/modules/entities/merge,
  internal/modules/revisions, internal/modules/incidentbundles,
  internal/modules/collaboration, internal/modules/links, and
  internal/modules/tasksdecisions.
- Assessment frontend model, controller, query, shell, support picker, mutation
  ports, and surface registry files under apps/web/src/workbook.
- Authored assessment OpenAPI, view-schema, import, projection-provider,
  revision, incident-bundle, recovery, and verification contracts.
- tools/backend_module_boundaries.json,
  tools/test_catalog_owner.json, and
  tools/test_families/module.assessments.json.
- Generated assessment view, inspector, OpenAPI, import-target, and TypeScript
  protocol registries were inspected only as downstream surfaces. They must
  never be hand-edited.

### Framework/repository finding

The planning framework's baseline module catalog omitted assessments. Live
Core requirements, the assessment view and OpenAPI contracts, the
module.assessments verification owner, the assessment test-family manifest,
and application composition all establish assessments as a current source
owner. Plan alignment repairs that documentation mismatch and does not infer
that the module should be removed.

The framework catalog omission is documentation drift. It is corrected during
plan alignment without changing product ownership or behavior.

### Normative language and tracker authority

The keywords MUST, MUST NOT, SHOULD, SHOULD NOT, and MAY are normative for the
authorized remediation slices in this tracker. MUST and MUST NOT
state an unconditional requirement; SHOULD and SHOULD NOT require a recorded
reason to deviate; MAY identifies a permitted option. These tracker
requirements are subordinate to adopted Core owners and adopted subsystem
boundary decisions. They do not amend those owners or decisions. A later owner
contradiction MUST be recorded as BLOCKED: owner contradiction; an implementing
agent MUST NOT choose a side.

The stable AMR-REQ identifiers in Section 4 are the sole normative requirement
catalog for this plan. Other sections cross-reference that catalog and MUST NOT
be read as creating additional product requirements.

### Defined terms

| Term | Exact meaning in this tracker |
| --- | --- |
| interactive create | Workbook row creation through POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows for cartulary.view.assessments.v1 |
| import owner create | One assessment owner-facade invocation inside an Imports-owned unit transaction |
| unit transaction | The caller-owned PostgreSQL transaction supplied by Imports to an owner facade |
| fresh create | A request for which no exact committed idempotency result is returned |
| exact committed replay | A request whose identity and canonical payload match an already committed result |
| materialized assessor | The explicit assessor, or the actor substituted when assessment.assessor is omitted |
| owner effect set | The assessment envelope, source row, support links when applicable, projection, revision/change set, and owner-side result |
| private provider | An implementation below internal/modules/assessments/internal that cannot be imported outside the assessments parent tree |

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| internal/modules/assessments/api.go | Stable assessment view ID, support-action limit, and validation error | AssessmentsViewSchemaID; CreateValidationError | Workbook mutation decoding; frontend/generated callers through the view ID; target tests | fmt only | Assessment contract and HTTP integration tests | Assessment OpenAPI and view-schema identities are downstream | assessments public facade | medium | Small public surface whose identifiers must not drift |
| internal/modules/assessments/assessment_contract_test.go | Broad owner contract, portability, revision, support-reference, and creation evidence | Test package surface only | module.assessments harness rows | App test harnesses, contracts, modules under test | File is itself test evidence | Verification rows reference selected tests | assessments tests | low | Preserve selector names represented in the harness |
| internal/modules/assessments/assessments_integration_test.go | HTTP create, projection, filter, append-only, and validation characterization | Test package surface only | module.assessments store rows | Server harness, Workbook route, PostgreSQL | File is itself integration evidence | OpenAPI/view behavior exercised indirectly | assessments tests | low | Current invalid-assessor case covers a missing user, not an active non-member |
| internal/modules/assessments/internal/providers/deleterestore/provider.go | Source-owned assessment delete/restore provider with fixed SQL and snapshots | Source; NewSource | RevisionProviderContribution in the root package | pgx; records; revisions delete/restore contracts | Revisions delete/restore tests and assessment owner tests | Revision target and snapshot registries are downstream | assessments private revision provider | medium | Go internal visibility and boundary policy confine construction to Assessments |
| internal/modules/assessments/facade.go | Transactional assessment row-create coordinator, validation, idempotency, projection, revisions, and support links | Facade; NewFacade; CreateInput/Result; port interfaces and DTOs | assessmentassembly; workbookassembly; tests | platform/postgres, pgx, uuid, injected record/entity/user/link/projection/revision/idempotency ports | facade_contract_test.go; assessment contract and integration tests | Assessment create schema and generated protocol types are downstream | assessments application facade | high | Legitimate source-owner coordination over a borrowed database handle |
| internal/modules/assessments/facade_contract_test.go | Transaction rollback, replay, participant, and facade contract characterization | Test package surface only | module.assessments harness row | In-memory port doubles and facade | File is itself unit evidence | None directly | assessments tests | low | Strong atomicity/replay coverage; no same-incident assessor case |
| internal/modules/assessments/import_create.go | Assessment-owned import create mapping and transactional source/projection/revision creation | ImportCreateFacade; NewImportCreateFacade | importassembly owner registry | imports/ownerfacade, pgx, postgres, assessment ports | Import registry composition tests; indirect owner tests | Import target registry and generated TypeScript binding are downstream | assessments import facade | high | Direct behavior, rollback, replay, and defaulting lack focused tests |
| internal/modules/assessments/incident_bundle_source_port.go | Stable root constructor for the private assessment incident-bundle source port | NewIncidentBundleSourcePort | incidentportabilityassembly catalog | private incident-bundle provider; incidentbundles/sourceport return contract | Incident-bundle catalog and route tests | contracts/incident-bundles/source_catalog.json | assessments source contribution | medium | Thin root wrapper is the only repository-visible source-port construction surface |
| internal/modules/assessments/incident_bundle_subtype_presence.go | Stable root constructor for assessment subtype-presence bindings | IncidentBundleSubtypeContribution | Incident portability assembly | private incident-bundle provider; records/subtypepresence return contract | Portability tests indirectly | Incident-bundle contract | assessments source contribution | medium | Thin root wrapper preserves the assembly API without exposing implementation |
| internal/modules/assessments/internal/providers/incidentbundle/portability.go | Implements private assessment bundle export and import apply behavior | Private export/import helpers | Private assessment incident-bundle source port | incidentportability; pgx | Incident Bundles archive/import evidence | Incident-bundle source catalog | assessments private incident-bundle provider | high | Fixed v2 path, ordering, stable identity, insert shape, and attribution semantics remain exact |
| internal/modules/assessments/internal/providers/incidentbundle/source_port.go | Owns the assessment source descriptor, prepare/apply lifecycle, and tuple validation | NewSourcePort | Root NewIncidentBundleSourcePort only | incidentbundles/sourceport; private portability implementation | Source-catalog, admission, and import route tests | contracts/incident-bundles/source_catalog.json | assessments private incident-bundle provider | high | Descriptor, declared failures, and source validation are private behind the root contribution |
| internal/modules/assessments/internal/providers/incidentbundle/subtype_presence.go | Lists assessment subtype bindings for incident-bundle admission | SubtypeContribution | Root IncidentBundleSubtypeContribution only | records/subtypepresence and fixed assessment SQL | Incident-bundle admission and route tests | Incident-bundle contract | assessments private incident-bundle provider | medium | Go internal visibility prevents direct repository consumers |
| internal/modules/assessments/internal/policy/assessment.go | Owns the canonical Go assessment subject, state, and confidence-band policy | ValidSubjectType; ValidState; ConfidenceBand | Assessment create, projection, and rollback paths | None | Policy contract plus existing create/projection/rollback rows | Core 02 vocabulary and frontend contract remain the normative/downstream boundaries | assessments private policy | high | One Go implementation covers valid and invalid boundary values |
| internal/modules/assessments/merge_effects.go | Assessment source participant for entity merge protected-set checks, subject repointing, projection refresh, and snapshots | MergeEffects; NewMergeEffects and typed result/port surface | internal/modules/entities/merge; timelineassembly composition | pgx, postgres, revisions, projection and snapshot ports | Entity merge protected-set and composition tests; revision integration tests | Entity merge contract behavior is downstream | assessments source merge participant | high | Entities owns generic merge; assessments owns assessment-row effects |
| internal/modules/assessments/policy_contract_test.go | Exercises the canonical policy vocabulary and exact confidence boundaries | Test package surface only | Assessment owner row | Private assessment policy | File is selected with the assessment state/band row | None directly | assessments tests | low | Covers nil, -1, 0, 39, 40, 69, 70, 100, and 101 |
| internal/modules/assessments/internal/providers/projection/provider.go | Reads authoritative assessment facts and derives typed projection inputs, paging, and mutations | Source; NewSource; provider methods implementing assessment contribution contracts | Root NewProjectionContribution only | pgx, postgres, workbookprojection, injected record/link fact readers | Projection assembly ownership/manifest tests; private projection runtime tests | Projection-provider manifest and assessment view contract | assessments private projection source provider | high | Physical projection SQL remains Projections-owned; Go internal visibility prevents application imports |
| internal/modules/assessments/projection_provider_contribution.go | Stable root constructor for the private assessment projection provider | ProjectionContributionDependencies; NewProjectionContribution | assessmentassembly projection composition | private provider and workbookprojection contract | Root construction test and Projections composition rows | Projection-provider manifest is unchanged | assessments source contribution | high | Root validates typed dependencies and is the sole provider-construction surface |
| internal/modules/assessments/projection_provider_contribution_test.go | Exercises root projection construction guards and descriptor identity | Test package surface only | Assessment owner row | Root contribution and workbookprojection contract | File is selected with the assessment state/band row | None directly | assessments tests | low | Protects the stable root boundary after provider internalization |
| internal/modules/assessments/recovery_state.go | Declares assessments as authoritative recovery state | RecoveryStateContribution | recoveryassembly state catalog | platform/recoverystate | Recovery state catalog tests | Recovery fixture/catalog contracts | assessments source contribution | medium | Global catalog test covers registration; focused assessment test is absent |
| internal/modules/assessments/revision_provider_contribution.go | Builds the assessment source-owned Revisions contribution | RevisionProviderContribution | revisionassembly | private delete/restore and rollback providers; revisions | Revisions catalog, delete/restore, rollback, and integration tests | Revision snapshot and target registries | assessments source contribution | medium | Current source-owner construction matches the adopted Revisions boundary |
| internal/modules/assessments/internal/providers/rollback/provider.go | Loads, validates, and reapplies canonical assessment rollback snapshots | Provider; NewProvider; provider methods | Root RevisionProviderContribution | pgx, postgres, records, revisions | provider_test.go and Revisions integration tests | Revision snapshot schema registry | assessments private revision provider | high | Uses the private canonical assessment policy; current membership is intentionally not consulted |
| internal/modules/assessments/internal/providers/rollback/provider_test.go | Focused invalid rollback snapshot validation | Test package surface only | Go test runner | rollback provider | File is itself test evidence | None directly | assessments tests | low | Relocated with the provider and retained as private-package evidence |
| internal/modules/assessments/source_repository.go | Fixed SQL insertion into the authoritative assessments table | Unexported repository and source-row types | Facade and import-create code | pgx and uuid | Facade and HTTP/integration tests | Assessment storage semantics reflected in snapshot and bundle contracts | assessments source persistence | high | Direct source-table SQL is intentional source-owner persistence |
| internal/modules/assessments/testsupport/assessments.go | Seeds assessment record, source, and projection state for cross-owner tests | SeedAssessment; LookupAssessmentSubject and fixture input | Revisions, incident-bundle, and other integration tests | database/sql, postgres, records, typed projections/testsupport mutations | Cross-owner integration suites | No physical projection-table access remains | assessments semantic test support | medium | Both pgx and database/sql projection effects delegate to Projections-owned capabilities |
| internal/modules/assessments/workbookprojection/contribution.go | Typed assessment source contribution, descriptors, read/write/rebuild ports, and semantic surface intent | Contribution; NewContribution; Descriptor and port interfaces | assessmentassembly; projectionassembly; projections adapters | projections/providercontract and querypage | contribution_test.go; projection assembly tests | contracts/projection-providers/index.json and assessment view schema | assessments public projection contract | high | Retain as the typed owner/storage boundary even if implementation moves |
| internal/modules/assessments/workbookprojection/contribution_test.go | Descriptor and semantic contribution characterization | Test package surface only | Go test runner | workbookprojection package | File is itself unit evidence | Projection provider manifest parity | assessments tests | low | Protect stable descriptor and view/source ownership |
| internal/modules/assessments/workbookprojection/model.go | Typed source and projection DTOs plus assessment input validation | ProjectionInput; SourceRow; Row and related typed models | assessment projection provider and Projections adapters/storage | uuid and validation utilities | Contribution and projection runtime tests | Generated view/protocol row shape is downstream | assessments public projection contract | high | Duplicates state vocabulary and confidence constraints |

No file in the target is out of scope for inventory. Test files are out of
scope for modification in this planning-only task, but remain in scope as
evidence.

## 3. Module Boundary Diagnosis

The target is a legitimate assessment source-owner and mutation coordinator
with a mixed-responsibility package tree. It is also projection-source
orchestration and source-persistence adjacent. It is not the HTTP transport
owner, frontend shell owner, physical projection-storage owner, generic
revision coordinator, Collaboration owner, or grid-vendor integration layer.

### Proposed topology

Authorized implementation MUST use this topology:

    internal/modules/assessments/
      projection_provider_contribution.go
      workbookprojection/
      internal/
        policy/assessment.go
        providers/
          projection/provider.go
          deleterestore/provider.go
          rollback/provider.go
          incidentbundle/

The assessments root facade and workbookprojection contract package remain
public within the repository. Root incident-bundle constructors remain stable,
and the root adds:

    type ProjectionContributionDependencies struct {
        Envelopes workbookprojection.EnvelopeReader
        Support   workbookprojection.SupportFactReader
    }

    func NewProjectionContribution(
        ProjectionContributionDependencies,
    ) (workbookprojection.Contribution, error)

Application assemblies depend on the assessments root. Retired provider paths
and portability helpers receive no compatibility aliases. This topology
implements the existing Projections decision requirement for typed owner
contributions and the Revisions decision requirement that source owners
construct their providers. It does not amend either adopted decision.

### Responsibility disposition

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Assessment vocabulary, source semantics, and source SQL | Root and source_repository.go | Assessments | keep | Core 02; source table and snapshots | AMR-REQ-017 |
| Interactive create orchestration | assessments.Facade | Assessments | keep | Workbook delegates to the owner facade | AMR-REQ-006 and AMR-REQ-007 |
| Scalar import owner create | ImportCreateFacade | Assessments within Imports transaction | keep | Live owner registry and facade ports | AMR-REQ-010 and AMR-REQ-013 |
| Entity merge effects | merge_effects.go | Assessments participant; Entities coordinator | split | Typed merge port and caller-owned transaction | Generic merge remains Entities-owned |
| Projection semantic input and typed intent | projectionprovider; workbookprojection | Assessments | move implementation; keep contract | Adopted Projections decision | Physical storage remains Projections-owned |
| Delete/restore and rollback semantics | deleterestore; rollbackprovider | Assessments private providers | move | Root Revisions contribution | Generic coordination remains Revisions-owned |
| Subject-type, state, and confidence policy | Facade, rollback, projection paths | Assessments private policy | move and consolidate | Live duplication | AMR-REQ-008 |
| Bundle and recovery contributions | Portability and recovery files | Assessments contribution; subsystem orchestration | split and internalize bundle implementation | Live catalogs and descriptors | Public identities remain frozen; AMR-REQ-024 |
| Authorization and direct-reference validation | Workbook admission; assessmentassembly validator | Workbook plus Incidents/Auth capability plus Assessments field contract | split | Live route and adapter interfaces | AMR-REQ-004 through AMR-REQ-006 |
| Frontend assessment controller and shell | apps/web Workbook feature | Workbook frontend | keep | Live imports and owner tests | External consumer |
| Grid-vendor integration | packages/grid-adapter | Grid adapter | keep | No direct vendor import in assessment feature | Intentional boundary |
| Collaboration publication | Collaboration and app/revision capabilities | Collaboration | keep | No direct assessment implementation import | External observable consequence |

### Provider path-impact map

| Slice | Old authored path | New authored path or surface | Every known path-bound consumer to reconcile |
| --- | --- | --- | --- |
| AS-S03 | internal/modules/assessments/deleterestore/provider.go | internal/modules/assessments/internal/providers/deleterestore/provider.go | Root revision_provider_contribution.go; Revisions delete_restore_test.go reflected type expectation; backend boundary delete/restore contract path, records-read path, and generic SQL metadata path |
| AS-S03 | internal/modules/assessments/rollbackprovider/provider.go and provider_test.go | internal/modules/assessments/internal/providers/rollback/provider.go and provider_test.go | Root revision_provider_contribution.go; backend boundary assessment provider-import and records-read paths |
| AS-S04 | internal/modules/assessments/projectionprovider/provider.go | internal/modules/assessments/internal/providers/projection/provider.go, reached through root NewProjectionContribution | internal/app/assessmentassembly/projection_source.go; internal/app/projectionassembly/source_ownership_test.go; internal/modules/projections/adapters/boundary_guard_test.go; backend boundary provider-import and records-read paths |
| AS-S05-B | Root incident-bundle portability and subtype implementation files | internal/modules/assessments/internal/providers/incidentbundle, reached through stable root contribution constructors | internal/app/incidentportabilityassembly/catalog.go; assessment and Incident Bundles tests; records-read and source-port boundary paths |

The exact tools/backend_module_boundaries.json entries for AS-S03 are
assessment-delete-restore-provider-import,
assessment-rollback-provider-import, the assessment path in
revisions-delete-restore-source-contract-import, the delete/restore and
rollback paths in records-current-envelope-access, and the assessment scan
path in delete-restore-sources-no-generic-sql-metadata. The AS-S04 boundary
entry is the projection-provider path in
records-current-envelope-access. The projection-specific authored guard is
internal/modules/projections/adapters/boundary_guard_test.go. A later move MUST
resolve entries by stable policy identity and before/after path, not by
historical line numbers.

## 4. Public Contract and Behavior Freeze Map

### Normative requirement catalog

| Requirement | Normative statement |
| --- | --- |
| AMR-REQ-001 | Every remediation change MUST remain inside the active Section 7 slice and MUST update this tracker after the slice passes, before the next slice begins. Generated artifacts MUST change only through Make-owned generation when an authored input requires it. |
| AMR-REQ-002 | Subject to decision-owner adoption, the Section 3 topology MUST retain the assessments root and workbookprojection, and MUST move policy and provider implementations to the stated private paths. |
| AMR-REQ-003 | Application assemblies MUST depend on the assessments root; the root MUST expose NewProjectionContribution with ProjectionContributionDependencies; retired provider paths MUST NOT receive compatibility aliases. |
| AMR-REQ-004 | The assessor port MUST have the semantic interface ValidateAssessmentAssessorTx(ctx, borrowedTx, incidentID, userID) returning valid and error. The borrowed transaction MUST be caller-owned and the validator MUST NOT begin, commit, roll back, or nest a transaction. |
| AMR-REQ-005 | The application adapter MUST call incidents/admission.Checker.CheckTx with RolesMember and LifecycleAny, then lock/read the Auth user with GetUserByIDForUpdateTx and require is_active=true. The checker MUST preserve its incident FOR SHARE then membership FOR SHARE order; membership removal MUST retain its incident FOR UPDATE first order. Rejection MUST NOT reveal whether a cross-incident user exists. |
| AMR-REQ-006 | For every fresh interactive or import owner create, the facade MUST materialize assessment.assessor from the actor when omitted and MUST transactionally validate the materialized assessor. |
| AMR-REQ-007 | Interactive create MUST follow the sequencing table. Existing subject-before-support-before-assessor validation precedence MUST remain unchanged. No owner effect may occur before assessor validation succeeds. |
| AMR-REQ-008 | Go assessment subject-type, state, and confidence policy MUST implement the exact subject, state, and band tables from one assessment-owned private seam. |
| AMR-REQ-009 | Interactive create MUST implement the interactive column of the field/default table without changing writable fields or observable defaults. |
| AMR-REQ-010 | ImportCreateFacade MUST remain scalar. Import omission MUST produce an empty support set. assessment.support_refs mapping MUST continue to fail normalization with collection_owner_support_required. The facade MUST NOT gain support-collection behavior. |
| AMR-REQ-011 | Assessor failures MUST follow the error-translation table. No new public token or cross-incident detail may be introduced. Internal owner errors MAY retain the field for tests and diagnostics. |
| AMR-REQ-012 | Exact replay and fresh-request behavior after membership change MUST follow the replay table. |
| AMR-REQ-013 | ImportCreateFacade MUST own source creation, owner defaults and validation, projection, and revision effects inside the borrowed unit transaction. Imports MUST own exact replay, conflict, apply journal, durable unit outcome, idempotency, and finalization. |
| AMR-REQ-014 | Every negative fresh-create assessor case MUST prove absence of all applicable effects in the negative-effect table. |
| AMR-REQ-015 | Assessment routes, envelopes, view ID, field keys, inspector features, saved-view compatibility, import target, bundle path/schema, recovery identity, frontend behavior, authorization results, storage semantics, and generated surfaces MUST remain unchanged unless a later task explicitly authorizes a behavior change. |
| AMR-REQ-016 | GET /ws/v1/incidents/{incident_id} authorization and record_changed semantics MUST remain Collaboration-owned and observably unchanged; assessment code MUST NOT acquire a direct Collaboration implementation dependency. |
| AMR-REQ-017 | Owner responsibility MUST follow the owner map. Verification and phase maps MUST be evidence accounting, not runtime architecture. |
| AMR-REQ-018 | AS-S03, AS-S04, and AS-S05-B MUST apply the exact Section 3 path-impact procedure. A move MUST atomically update all and only authored consumers whose before/after comparison proves a path dependency. |
| AMR-REQ-019 | contracts/projection-providers/index.json, verification contracts, and test-catalog ownership MUST remain unchanged unless a before/after comparison proves a serialized or selector path changed. Generated files MUST NOT be hand-edited. |
| AMR-REQ-020 | The 15 existing module.assessments rows MUST remain. AS-S00A MUST add independently selectable service-backed TestAssessmentImportCreateFacadeContract_Integration, yielding 16. AS-S01 MUST add TestAssessmentAssessorMembershipContract_Integration, yielding 17. Existing semantic rows SHOULD be extended instead of duplicated. |
| AMR-REQ-021 | Authorized execution MUST follow AS-S00A, AS-S01, AS-S02, AS-S03, AS-S04, AS-S05-T, AS-S05-B, then AS-S06. Metadata reconciliation is atomic with the matching move. |
| AMR-REQ-022 | Each selected slice MUST satisfy its acceptance evidence and Make-owned commands in Sections 7, 8, and 12, then update this tracker, before the next slice begins. |
| AMR-REQ-023 | Assessment test support MUST delegate physical assessment-projection fixture writes to typed Projections test support for both pgx and database/sql transactions; the assessment projection-table test-write exception MUST be removed. |
| AMR-REQ-024 | Assessment Incident Bundle export, import, validation, and subtype implementation MUST be private behind the stable root source-port and subtype-contribution constructors. ExportIncidentBundleFiles and ImportIncidentBundleFilesTx MUST be removed without aliases while bundle v2 behavior remains exact. |
| AMR-REQ-025 | The modular-refactor planning framework MUST list Assessments as a source owner and MUST keep Workbook transport, physical projection storage, generic Imports/Revisions/Incident Bundles coordination, and Collaboration outside that ownership. |

### State and confidence mappings

| Input | Canonical output | Validity |
| --- | --- | --- |
| assessment_state = unknown | unknown | valid |
| assessment_state = suspected | suspected | valid |
| assessment_state = confirmed | confirmed | valid |
| assessment_state = disproven | disproven | valid |
| assessment_state = cleared | cleared | valid |
| Any other assessment_state | none | invalid |
| confidence_score = nil | unset | valid |
| confidence_score = 0 through 39 | low | valid |
| confidence_score = 40 through 69 | medium | valid |
| confidence_score = 70 through 100 | high | valid |
| confidence_score below 0 or above 100 | none | invalid |

### Interactive and import field/default mapping

| Semantic field | Interactive create | Import owner create |
| --- | --- | --- |
| incident | Route incident; not writable in row fields | Unit-context incident; not a scalar mapping field |
| assessment.subject_ref | Required writable record reference | Required writable scalar reference |
| assessment.subject_type | Required writable host or identity token | Required writable scalar host or identity token |
| assessment.assessment_state | Required writable enum | Required writable scalar enum |
| assessment.confidence_score | Optional writable integer; omission means nil | Optional writable scalar integer; omission means nil |
| assessment.rationale | Required writable nonempty text after create-time normalization | Required writable nonempty scalar text after create-time normalization |
| assessment.assessor | Optional writable user reference; omission materializes actor | Optional writable scalar user reference; omission materializes actor |
| assessment.assessed_at | Optional writable timestamp; omission uses owner commit time | Optional writable scalar timestamp; omission uses owner transaction commit time |
| assessment.support_refs | Optional create-only unordered collection_actions_v1 containing 1 through 64 add_record_ref actions when present; omission means empty | Not writable by the scalar facade; omission means empty; mapping fails with collection_owner_support_required |
| Source row | Created once after validation | Created once after validation in the borrowed unit transaction |

### Create sequencing

| Order | Required step | Participant/effect constraint |
| --- | --- | --- |
| 1 | Validate request shape | No participant or persistent effect |
| 2 | Resolve idempotency and return an exact committed replay | Exact replay returns before participant validation; hash conflict remains an error |
| 3 | Begin or receive the owner transaction | Interactive owner begins; import facade receives the Imports-owned transaction |
| 4 | Validate subject, then support targets when interactive | Preserve subject-before-support-before-assessor precedence |
| 5 | Materialize and validate the assessor | Use AMR-REQ-004 through AMR-REQ-006 |
| 6 | Perform allowed effects | Source, envelope, support, projection, revision, idempotency, journal, and commit effects occur only in their owning boundary after validation |

### Error translation

| Boundary | Invalid assessor representation | Publicly safe result |
| --- | --- | --- |
| Interactive owner facade | CreateValidationError with field assessment.assessor and reason invalid_value | Internal field/reason retained |
| Workbook HTTP | Existing mutation error mapping | 400 invalid_mutation_payload with field assessment.assessor and reason invalid_value |
| Import owner facade | Existing owner-create validation failure; diagnostics MAY retain assessment.assessor | No new owner reason token |
| Public Imports apply | Existing closed-registry safe translation | import_apply_blocked with owner_create_validation_failed and no cross-incident existence detail |

### Replay and membership-change mapping

| Request state after assessor membership changes | Required behavior |
| --- | --- |
| Exact committed replay and actor remains authorized for the operation | Return the original committed result before participant validation |
| Fresh client transaction identity with omitted departed assessor | Materialize actor; reject if actor no longer satisfies membership or active-user requirements |
| Fresh client transaction identity with explicit departed assessor | Reject through AMR-REQ-011 mapping |
| Payload conflict using an existing identity | Preserve current conflict behavior; do not invoke create participants |

### Negative fresh-create effect proof

| Effect | Interactive negative case | Import negative case |
| --- | --- | --- |
| Assessment envelope and source row | MUST be absent | MUST be absent |
| Support links | MUST be absent | Not applicable; import support mapping is prohibited |
| Assessment projection | MUST be absent | MUST be absent |
| Revision and change set | MUST be absent | MUST be absent |
| Committed idempotency fact | MUST be absent | MUST be absent at the Imports boundary |
| Apply journal and durable unit outcome | Not applicable | MUST have no successful or completed fact |
| Collaboration record_changed publication | MUST be absent | MUST be absent |

### Owner responsibility map

| Owner | Owns | MUST NOT acquire through this refactor |
| --- | --- | --- |
| Workbook | HTTP decoding/envelopes, route authorization, view/saved-view/UI orchestration | Assessment source semantics or private providers |
| Assessments | Source semantics/SQL, owner facades, merge participation, typed projection intent, revisions/recovery/portability contributions | HTTP transport, projection-table SQL, generic import finalization, WebSocket transport |
| Incidents and Auth | Incident membership/lifecycle capability and active-user truth | Assessment source creation |
| Imports | Unit transaction, normalization, replay/conflict, journal, outcome, idempotency, finalization | Assessment support-collection semantics or source policy |
| Projections | Physical storage, query compilation, refresh/rebuild runtime | Assessment source policy |
| Revisions | Generic history, delete/restore, rollback, and catalog coordination | Assessment snapshot meaning |
| Links | Support-link persistence capability | Assessment transport or source ownership |
| Collaboration | WebSocket authorization, event protocol, and publication | Assessment source writes |

### Assessor acceptance matrix

These scenario IDs apply to interactive and import creation unless stated
otherwise.

| Scenario | Assessor condition | Expected result | Required proof |
| --- | --- | --- | --- |
| ASR-01 | Omitted actor; active same-incident member | Accept materialized actor | Stored source, projection, and revision agree |
| ASR-02 | Explicit actor; active same-incident member | Accept | Explicit value is stored |
| ASR-03 | Another active same-incident member | Accept | Actor and assessor remain distinct |
| ASR-04 | Unknown user | Reject | AMR-REQ-011 and AMR-REQ-014 |
| ASR-05 | Inactive same-incident member | Reject | AMR-REQ-011 and AMR-REQ-014 |
| ASR-06 | Active user with no incident membership | Reject | AMR-REQ-011 and AMR-REQ-014 |
| ASR-07 | Active user belonging only to another incident | Reject without existence detail | AMR-REQ-011 and AMR-REQ-014 |
| ASR-08 | Membership removal commits before validation | Reject | Deterministic ordering and AMR-REQ-014 |
| ASR-09 | Create obtains required locks before membership removal | Create commits, then removal completes | Serial order; no partial result |
| ASR-10 | Assessor leaves after committed create; exact replay requested | Return original result | Participants are not reinvoked |
| ASR-11 | Assessor leaves after committed create; fresh client transaction identity used | Reject | AMR-REQ-012 and AMR-REQ-014 |

### Assessment import acceptance matrix

AS-S00A adds only cases that pass before correction. AS-S01 adds intended
assessor rejection cases atomically with the correction.

| Scenario | Boundary and case | Required assertion | Slice |
| --- | --- | --- | --- |
| IMP-01 | Host subject and identity subject | Both create matching source/projection/revision facts | AS-S00A |
| IMP-02 | Omitted and explicit assessor/time | Values follow the field/default table | AS-S00A; rejection variants in AS-S01 |
| IMP-03 | Minimum semantic fields | subject_ref, host or identity subject_type, assessment_state, and nonempty normalized rationale are all required; defaults and support do not satisfy the minimum; a valid minimum creates one source, projection, and revision result | AS-S00A |
| IMP-04 | Missing, deleted, wrong-type, and cross-incident subject | Owner validation fails atomically with existing safe mapping | AS-S00A |
| IMP-05 | Scalar normalization | Admitted scalar values normalize without semantic drift | AS-S00A |
| IMP-06 | assessment.support_refs mapping | Fails with collection_owner_support_required; adds no support behavior | AS-S00A |
| IMP-07 | Source/projection/revision agreement | Stored facts identify the same assessment and values | AS-S00A |
| IMP-08 | Failure at envelope, source, projection, or revision boundary | Borrowed transaction rolls back all owner effects | AS-S00A |
| IMP-09 | Replay, conflict, journal, outcome, and finalization | Asserted at Imports unit-commit boundary, not direct facade | AS-S00A |
| IMP-10 | Unknown, inactive, nonmember, other-incident-only assessor | AMR-REQ-011 and AMR-REQ-014 | AS-S01 |

### Structural characterization matrix

| Scenario | Surface | Required assertion | Slice |
| --- | --- | --- | --- |
| CHR-01 | Projection input and output | Nil confidence, every band boundary, support count, subject cell, assessor cell, assessment state, envelope, and row version remain exact | AS-S00A and existing selected projection rows |
| CHR-02 | Projection rebuild | Incremental create and incident rebuild produce equivalent canonical assessment rows | AS-S00A by extending existing selected projection evidence |
| CHR-03 | Rollback provider | Canonical snapshot remains record plus source; malformed snapshot fails without effects; valid reapply refreshes the same canonical row | AS-S00A by extending the selected rollback evidence |
| CHR-04 | Incident-bundle contribution | Source descriptor, data/compromise_assessments.ndjson identity, tuple ordering, and required subtype-presence rejection remain exact | AS-S00A by extending selected portability evidence |
| CHR-05 | Recovery contribution | Contribution identity remains assessments, authoritative assessment state remains included, and derived projection state remains excluded | AS-S00A by extending selected recovery evidence |

### Contract freeze inventory

| Contract | Current owner | Evidence | Existing tests | Required acceptance | Risk |
| --- | --- | --- | --- | --- | --- |
| Row-create route and envelopes | Workbook; Assessments | OpenAPI, view ID, facade | HTTP, facade, browser | ASR matrix; sequencing/error tables | high |
| Query route | Workbook; Projections; Assessments intent | OpenAPI; SurfaceIntent | Projection, filter, browser | Existing paging/filter/null/support-count rows remain | high |
| WebSocket and record_changed | Collaboration | Protocol and app wiring | Collaboration and Workbook | AMR-REQ-016 | high |
| Append-only source and defaults | Assessments | Core 02; repository | Contract and integration | Field/default and state tables | high |
| Projection refresh and confidence band | Assessments intent; Projections storage | Descriptor/model | Projection/runtime | AMR-REQ-008 | high |
| Revision, delete/restore, rollback | Revisions plus Assessments providers | Contribution/registries | Revisions suites | AS-S03 behavior preservation | high |
| Merge repointing | Entities plus Assessments | Merge port/effects | Merge and revision | Existing selected evidence | high |
| Import target | Imports plus Assessments | Registry/facade | Registry/composition | IMP matrix | high |
| Bundle source and recovery identity | Incident Bundles/Recovery plus Assessments | Catalogs/ports | Contract/catalog | AS-S00A focused cases | high |
| View, inspector, saved views, frontend | View contract; Workbook; Saved Views; apps/web | cartulary.view.assessments.v1 | Contract/frontend/browser | Existing rows remain | high |
| Generated registries | Contract owners/generators | Generated policy | Drift/contract | AMR-REQ-019 | high |
| Harness accounting | Verification/test-family owners | Authored catalogs | 15 current rows | AMR-REQ-020 | medium |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Assessor validation omits same-incident membership | View binding requires incident_member_user_ref_v1; adapter checks only active user | high | must_fix | Assessments using Incidents/Auth | Execute AMR-REQ-004 through AMR-REQ-007 only in AS-S01 |
| State validation is duplicated | Facade, rollback provider, workbookprojection model | medium | should_fix | Assessments private policy | AS-S02 implements AMR-REQ-008 |
| Confidence-band policy is duplicated | Projection provider and workbookprojection model | medium | should_fix | Assessments private policy | AS-S02 implements AMR-REQ-008 |
| Providers are public sibling packages | Live imports and boundary manifest | medium | must_fix | Assessments private providers | AS-S03/AS-S04 atomically move code and metadata |
| Testsupport writes projection table | Explicit boundary exception | medium | should_fix | Assessments plus Projections testsupport | AS-S05-T adds replacement capability, then removes exception |
| Source repository uses fixed SQL | Writes only authoritative assessment source | low | intentional/no_action | Assessments | Preserve source ownership |
| Facades use postgres/pgx transactions | Transactions are owned or borrowed at defined boundaries | medium | intentional/no_action | Assessments/Imports | Preserve AMR-REQ-004 and AMR-REQ-013 |
| Merge effects update assessment rows | Entities calls typed assessment participant | medium | intentional/no_action | Assessments; Entities coordinator | Preserve dependency direction |
| Projection SQL is outside assessments | Projections owns assessment_grid_projection SQL | low | intentional/no_action | Projections | Do not move physical SQL |
| Workbook owns transport/frontend | Generic routes and Workbook feature consume assessment surfaces | low | intentional/no_action | Workbook | Preserve AMR-REQ-015 |
| Grid access is adapter-only | Frontend imports packages/grid-adapter | low | intentional/no_action | Grid adapter | No direct vendor import |
| Collaboration is indirect | No assessment implementation import | low | intentional/no_action | Collaboration | Preserve AMR-REQ-016 |
| Generated registries are downstream | Authored contracts feed generated roots | high | intentional/no_action | Contract owners/generators | Preserve AMR-REQ-019 |
| Framework omits assessments | Framework versus live owners/composition | medium | should_fix | Planning framework documentation | Plan alignment implements AMR-REQ-025 |
| Exported portability helpers remain beside the source port | Root helpers are consumed only by the root port | medium | should_fix | Assessments/Incident Bundles | Characterize in AS-S00A, then internalize in AS-S05-B |

Non-authoritative rationale is consistent with
[OWASP authorization guidance](https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html),
[PostgreSQL transaction semantics](https://www.postgresql.org/docs/current/tutorial-transactions.html),
and [Go internal package layout](https://go.dev/doc/modules/layout). Product
behavior MUST NOT depend on those external pages.

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Source bootstrap and tracker initialization | root | none | WF-01 | Lock authority, baseline, vocabulary, scope | Tracker/owners | Read-only review | Requirement catalog established |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-04 | Account for every live target file and authorized addition | assessments/** | Inventory comparison | Section 2 updated at each structural checkpoint |
| WF-02 | Contract-owner mapping | parallel | WF-01 | WF-03, WF-05 | Freeze behavior and assign owners | Core, contracts, composition | Evidence review | AMR-REQ-015 through AMR-REQ-017 mapped |
| WF-03 | Characterization analysis | chain | WF-01, WF-02 | WF-05 | Separate pre-change from correction evidence | Assessment/adjacent tests | Section 4 matrices | AS-S00A/AS-S01 evidence separated |
| WF-04 | Boundary/coupling scan | parallel | WF-01 | WF-05 | Classify imports, SQL, providers, policy | Target/adjacent/boundary manifest | Path-impact review | Section 5 complete |
| WF-05 | Facade/ownership design | chain | WF-02, WF-03, WF-04 | WF-06 | Apply the owner-contribution topology already required by adopted decisions | Root/providers/adapters/ADRs | Owner review | Existing decisions mapped; no contradiction |
| WF-06 | Slice sequencing | chain | WF-05 | WF-07, WF-08 | Execute AMR-REQ-021 reversibly | Selected packages | Per-slice acceptance | Each slice has rollback point |
| WF-07 | Harness/accounting reconciliation | parallel | WF-06 | WF-08 | Preserve/extend authored routing | Verification/catalog/boundaries | Before/after comparisons | AMR-REQ-018 through 020 satisfied |
| WF-08 | Validation and handoff | chain | WF-06, WF-07 | none | Run scoped-to-broad checks | Selected slice/tracker | Section 8 | AMR-REQ-022 satisfied |

WF-00 through WF-05 and planning for WF-06 are DONE. The current remediation
request authorizes WF-06 through WF-08 in the Section 7 order.

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| AS-S00A | WF-03 | Add passing characterization for import, projection, rollback, portability, recovery. MUST NOT add ignored, expected-failure, or current-acceptance assessor-defect tests. | Assessment tests; importassembly integration; existing adjacent tests | Misassigned Imports behavior or blessed defect | IMP-01 through IMP-09 and CHR-01 through CHR-05; extend selected rows; add TestAssessmentImportCreateFacadeContract_Integration | make test-slice OWNER=module.assessments; make service-backed-test-slice OWNER=module.assessments | Remove new tests/authored row only | New tests pass before correction; count 16 |
| AS-S01 | AS-S00A | Implement AMR-REQ-004 through 007 and intended rejection tests atomically | Facades/ports; assessmentassembly; Incidents/Auth; integration tests | Intentional active-nonmember correction; error/transaction drift | ASR-01 through 11; IMP-10; add TestAssessmentAssessorMembershipContract_Integration | make service-backed-test-slice OWNER=module.assessments; make test-slice OWNER=module.assessments | Revert implementation/tests/row atomically | All cases pass; no effects; count 17 |
| AS-S02 | AS-S01 | Consolidate Go subject-type/state/confidence policy privately | Facade; model; projection/rollback providers | Token/boundary drift | Extend selected subject/state/band, projection, rollback, browser tests; no redundant row | make test-slice OWNER=module.assessments | Restore call sites/local rules together | AMR-REQ-008 has one implementation |
| AS-S03 | AS-S02 | Move delete/restore, rollback, and rollback test privately; atomically reconcile exact authored metadata | Section 3 AS-S03 paths/consumers | Snapshot/dispatch/refresh/reflection drift | Existing assessment/Revisions tests | assessment and Revisions owner slices; backend boundary | Restore paths/imports/metadata; no alias | Runtime and boundary checks pass together |
| AS-S04 | AS-S03 | Move projection provider privately; add root constructor; atomically reconcile exact authored metadata | Section 3 AS-S04 paths/consumers | Descriptor/paging/query/refresh drift | Existing provider/descriptor/runtime/integration tests | assessment and Projections owner slices; backend boundary and drift | Restore path/app import/metadata; no alias | App imports root; contract/storage unchanged |
| AS-S05-T | AS-S04 | Move database/sql assessment projection fixture mutation into typed Projections test support and remove the boundary exception | Assessment and Projections testsupport; boundary policy | Test-driver divergence or test-only permission widening | Existing cross-owner assessment fixture consumers | assessment, Entities, Revisions, and Projections owner slices; backend boundary | Restore test capability, caller, and exception together | AMR-REQ-023 passes with no raw projection SQL in assessment testsupport |
| AS-S05-B | AS-S05-T | Move bundle export/import/validation/subtype implementation privately; retain thin root constructors | Section 3 AS-S05-B paths/consumers | Bundle path/schema/order/attribution drift | CHR-04 and existing assessment/Incident Bundles evidence | assessment and Incident Bundles owner slices; backend boundary | Restore implementation and paths; no helper aliases | AMR-REQ-024 passes byte-for-byte |
| AS-S06 | AS-S05-B | Scoped-to-broad verification and handoff | Changed packages/tracker | Cross-owner/browser/generated regression | All 17 rows plus applicable adjacent/frontend/browser evidence | make agent-finalize; Section 8 | Revert smallest failing slice | Checks pass and artifacts/ownership are recorded |

AS-S05-R is DROPPED into AS-S03 and AS-S05-P is DROPPED into AS-S04. The
current remediation request authorizes all remaining slices.

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| task selection | make task-guide ROLE=module-author OWNER=module.assessments; make explain-test-owner OWNER=module.assessments | Owner routing/accounting | yes | Repeat after row/path changes |
| owner slice | make test-slice OWNER=module.assessments | Assessment owner rows | yes | Counts 15, 16, then 17 |
| service-backed | make service-backed-test-slice OWNER=module.assessments | Store-backed rows | yes | New rows independently selectable |
| adjacent owners | make test-slice for module.imports, module.incidents, module.revisions, module.projections, module.incidentbundles, and module.recovery | Slice-specific owners | no | Required after matching behavior or move |
| browser | make browser-e2e-webserver-backed | HTTP/projection/frontend | no | Required in AS-S06 |
| frontend | make frontend-typecheck; make frontend-unit; make frontend-import-boundary-check | Workbook consumer/boundary | no | Required only when affected |
| generated drift | make generate-drift; make generated-artifact-policy-check; make json-shape-check | Generated policy/shapes | no | Required after authored metadata changes |
| backend static | make backend-module-boundary-check | Imports, SQL, path allowances | yes | Required AS-S03 through AS-S05-B |
| broad | make test-fast; make check | Fast/full repository | no | Narrow checks first |
| handoff | make agent-finalize | Harness/run maintenance | no | Required before final broad verification |
| tracker | make lint-markdown | Markdown | yes | Run after plan alignment and final handoff |

Every slice runs its narrow checks before its tracker checkpoint. AS-S06 runs
the complete ladder and records retained run roots or the reason none exist.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| AS-T001 | Lock scope, authority, vocabulary, requirements | WF-00 | DONE | none | Sections 1/4 | AMR-REQ-001/precedence explicit |
| AS-T002 | Inventory original target and authorized additions | WF-01 | DONE | AS-T001 | Section 2 | One row per current target file |
| AS-T003 | Map contracts/owners | WF-02 | DONE | AS-T002 | Sections 3/4 | Maps complete |
| AS-T004 | Classify coupling | WF-04 | DONE | AS-T002 | Section 5 | Allowed classifications only |
| AS-T005 | Specify topology/path impacts | WF-05 | DONE | AS-T003, AS-T004 | Section 3 | Existing adopted decisions support the topology |
| AS-T006 | Add pre-correction characterization | WF-03 | DONE | AS-T005 | AS-S00A | Tests pass; count 16 |
| AS-T007 | Correct assessor membership | WF-05 | DONE | AS-T006 | AS-S01 | ASR/IMP-10 pass; count 17 |
| AS-T008 | Consolidate private policy | WF-06 | DONE | AS-T007 | AS-S02 | One policy implementation |
| AS-T009 | Move revision providers and metadata | WF-06/WF-07 | DONE | AS-T008 | AS-S03 | Runtime and boundary checks pass atomically |
| AS-T010 | Retired standalone revision metadata slice | WF-07 | DROPPED | AS-T009 | Absorbed into AS-S03 | No intermediate red boundary state |
| AS-T011 | Move projection provider and metadata | WF-06/WF-07 | DONE | AS-T009 | AS-S04 | Root constructor and boundary checks pass atomically |
| AS-T012 | Retired standalone projection metadata slice | WF-07 | DROPPED | AS-T011 | Absorbed into AS-S04 | No intermediate red boundary state |
| AS-T013 | Move fixture projection writes to Projections | WF-06/WF-07 | DONE | AS-T011 | AS-S05-T | Boundary exception removed |
| AS-T014 | Internalize Incident Bundle implementation | WF-06/WF-07 | DONE | AS-T013 | AS-S05-B | Root constructors only; bundle behavior exact |
| AS-T015 | Final validation/handoff | WF-08 | DONE | AS-T014 | AS-S06 | Checks pass and failures are attributed |
| AS-T016 | Align authorized tracker/framework | WF-00/WF-08 | DONE | AS-T001-AS-T005 | Tracker and framework | Markdown lint, whitespace, authority, and sequence pass |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-19T08:29:33-04:00 | Codex tracker session | Target exists; tracker created under planning-only authority | Touched: docs/handoffs/assessments-module-refactor-tracker.md only. Inspected: framework, domain, NLSpec research guide, Core 00-05, adopted Projections/Revisions decisions | sed, rg, find, wc, git status, git rev-parse | Authority ordered; Core 05 not applicable; no owner contradiction | None for tracker completion | Obtain separate authorization before any code/test change |
| 2026-08-19T09:56:08-04:00 | Codex NLSpec revision | Normative catalog and subordinate authority specified | Touched: tracker only. Inspected: notes, tracker, NLSpec guide, Core, live interfaces | sed, rg, git status, sha256sum, make lint-markdown, git diff HEAD --check | Recommendations accepted only where verified; documentation and scope checks passed | RB-001/RB-003 affect later work | Obtain separate authorization for AS-S00A |
| 2026-08-19T10:24:09-04:00 | Codex remediation execution | Plan alignment complete under explicit implementation authority | Touched: tracker and modular-refactor framework. Inspected: Core 01/02, adopted Projections/Revisions decisions, live symbols and paths | rg, sed, make task-guide, make explain-test-owner, make explain-target, make lint-markdown, git diff --check | PASS; lint run root .cartulary/test-results/20260819T142732Z-p1963948; Core already owns assessor membership and assessment policy; metadata slices made atomic; framework owner catalog repaired | None | Execute AS-S00A and checkpoint before AS-S01 |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-19T08:29:33-04:00 | Codex tracker session | Legitimate source owner with mixed provider/adapter surface | Touched: tracker only. Inspected: all 21 assessment files, app composition, Projections, Entities merge, Workbook, Revisions, bundle/recovery callers | rg, sed, find | Root facade/source semantics retained; provider internalization deferred | RB-003 | Execute AS-S00, then approve WF-05 topology |
| 2026-08-19T09:56:08-04:00 | Codex NLSpec revision | Private topology/root constructor are decision-complete | Touched: tracker only. Inspected: provider imports, contributions, app assembly, guards, boundary manifest | rg, sed, jq | Exact AS-S03/AS-S04 consumers recorded; no aliases | RB-003 | AS-S00A, owner approval, AMR-REQ-021 |
| 2026-08-19T11:03:39-04:00 | Codex remediation execution | Assessment Revisions implementation is private | Touched: private providers, root contribution, Revisions expectation, exact boundary policy paths | Revisions/Assessments owner slices; backend module boundary check | Only `assessments.RevisionProviderContribution` constructs the providers; no compatibility package or alias remains; reflection expectation now names the private delete/restore type; boundary run `.cartulary/test-results/20260819T145912Z-p2564640` | None | Preserve root construction while moving projection implementation in AS-S04 |
| 2026-08-19T11:13:31-04:00 | Codex remediation execution | Assessment projection implementation is private behind the root contribution | Touched: private projection provider, root constructor/test, assessmentassembly, Projections ownership/guard tests, records-read boundary path, generated execution-topology index, tracker | rg before/after import inventory; make format-go; make backend-module-boundary-check; Assessments/Projections owner slices; make generate; make generate-drift; git diff --check | PASS; application composition imports the Assessments root and typed workbookprojection contract only; no old provider path or alias remains; exact same-owner private-provider allowlist entries cover existing Artifacts/Indicators/Tasks-Decisions patterns without wildcard expansion. Boundary `.cartulary/test-results/20260819T150635Z-p2774080`; generated run `.cartulary/test-results/20260819T151217Z-p2924692`; drift `.cartulary/test-results/20260819T151234Z-p2927748` | None | Execute AS-S05-T only after this checkpoint |
| 2026-08-19T11:32:02-04:00 | Codex remediation execution | Assessment fixture projection writes are Projections-owned | Touched: Projections typed fixture capability and selected SQL-transaction evidence, assessment test support delegation, Projections ownership expectation, backend boundary policy/self-test, Projections row fixture classification and transaction approval, generated execution-topology index, tracker | make format; selected Projections service-backed row; make backend-module-boundary-check; Assessments/Projections/Revisions/Entities owner slices; make generate; make generate-drift; rg; git diff --check | PASS; assessment testsupport contains no `assessment_grid_projection` token; pgx callers and a real database/sql upsert/delete transaction passed. Boundary `.cartulary/test-results/20260819T152239Z-p2972430`; Assessments focused/service-backed `.cartulary/test-results/20260819T152246Z-p2972829` and `.cartulary/test-results/20260819T152342Z-p3014130`; Projections `.cartulary/test-results/20260819T152434Z-p3054387` and `.cartulary/test-results/20260819T152513Z-p3069757`; Revisions `.cartulary/test-results/20260819T152553Z-p3084843` and `.cartulary/test-results/20260819T152657Z-p3128678`; Entities `.cartulary/test-results/20260819T152801Z-p3171471` and `.cartulary/test-results/20260819T152943Z-p3223810`; generation/drift `.cartulary/test-results/20260819T153126Z-p3275789` and `.cartulary/test-results/20260819T153138Z-p3278721`. The initial selected SQL row run failed because a transaction-scoped row cannot open the required isolated database; the existing Projections storage row was correctly reclassified as dedicated and its transaction approval removed before the passing retry `.cartulary/test-results/20260819T152116Z-p2956736` | None; no test-only permission was broadened beyond the Projections-owned capability path | Execute AS-S05-B only after this checkpoint |
| 2026-08-19T11:43:28-04:00 | Codex remediation execution | Assessment incident-bundle implementation is private behind root contributions | Touched: private assessment incident-bundle provider, thin root source/subtype constructors, retired root portability helper file, exact incident-bundle/Records/import boundary paths and retired-helper guard, tracker | make format; make backend-module-boundary-check; Assessments/Incident Bundles/Recovery owner slices; rg; git diff --check | PASS; no assessment root helper definitions or external helper references remain; v2 descriptor/path/stable identity, export ordering, import tuple validation, attribution, and Recovery inclusion evidence pass. Boundary `.cartulary/test-results/20260819T153614Z-p3288312`; Assessments focused/service-backed `.cartulary/test-results/20260819T153620Z-p3288698` and `.cartulary/test-results/20260819T153716Z-p3330019`; Incident Bundles `.cartulary/test-results/20260819T153811Z-p3370306` and `.cartulary/test-results/20260819T153920Z-p3385521`; Recovery `.cartulary/test-results/20260819T154021Z-p3400479` and `.cartulary/test-results/20260819T154145Z-p3450918`. The first boundary run exposed that thin root signatures still require their public contract imports; the exact two root paths were retained while SQL and implementation permissions moved privately | None; no compatibility alias retained | Execute AS-S06 only after this checkpoint |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-19T08:29:33-04:00 | Codex tracker session | Assessment UI is a Workbook-owned external consumer | Touched: tracker only. Inspected: assessment model, controller, query, surface, shell, mutation ports, support picker, surface registry, view-contract adapters | rg, sed | No frontend move proposed; grid vendor remains behind grid-adapter | None | Preserve six frontend and three browser owner rows |
| 2026-08-19T09:56:08-04:00 | Codex NLSpec revision | Frontend/saved-view/grid boundaries are frozen consumers | Touched: tracker only. Inspected: existing frontend evidence and notes | sed, rg | No frontend move or behavior change | None | Extend existing selected rows only if affected |
| 2026-08-19T12:20:51-04:00 | Codex remediation execution | Frontend and browser compatibility validation complete | Inspected unchanged Workbook assessment consumers; no frontend source changed | make frontend-typecheck; make frontend-unit; make frontend-import-boundary-check; make browser-e2e-webserver-backed | PASS; typecheck `.cartulary/test-results/20260819T160725Z-p4101824`; 390-unit frontend suite `.cartulary/test-results/20260819T160737Z-p4102341`; import boundary `.cartulary/test-results/20260819T160924Z-p4143986`; 58-unit browser suite `.cartulary/test-results/20260819T160929Z-p4144425` | None | None; compatibility handoff complete |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-19T08:29:33-04:00 | Codex tracker session | Authored view/OpenAPI/import/projection/revision/bundle/recovery contracts mapped to downstream generated surfaces | Touched: tracker only. Inspected: relevant contracts and generated Go/TypeScript registries | jq, rg, sed | No contract change proposed; generated files remain read-only | RB-004 if paths later move | Update authored sources first, then use Make generation |
| 2026-08-19T09:56:08-04:00 | Codex NLSpec revision | Serialized contracts frozen; path edits require proof | Touched: tracker only. Inspected: provider index, verification/catalog ownership, generated policy, boundary paths | jq, rg, sed | AMR-REQ-018/019 prevent speculation | RB-004 | Compare during AS-S05-R/P |
| 2026-08-19T12:20:51-04:00 | Codex remediation execution | Serialized contracts remain unchanged; authored harness metadata and Make-generated digest are synchronized | Touched: assessment/Projections test-family manifests, PostgreSQL fixture policy, harness closed-count assertion, generated execution-topology render index; inspected generated/lockfile status | make harness-contract; make generate-drift; make generated-artifact-policy-check; make json-shape-check; git status/diff | PASS; harness `.cartulary/test-results/20260819T160635Z-p4096515`; drift `.cartulary/test-results/20260819T160650Z-p4097001`; artifact policy `.cartulary/test-results/20260819T160701Z-p4099924`; JSON shape `.cartulary/test-results/20260819T160705Z-p4100359`. The first harness-contract run `.cartulary/test-results/20260819T160615Z-p4095698` correctly identified the stale closed dedicated-fixture count (254 versus 256), which was updated before the clean rerun | None; no lockfile, OpenAPI, view-schema, protocol, migration, or database-schema delta | None; generated handoff complete |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-19T08:29:33-04:00 | Codex tracker session | module.assessments has 15 rows: 3 Playwright, 6 Vitest, 6 Go; 8 are service-backed | Touched: tracker only. Inspected: assessment tests, adjacent merge/projection/revision/import/recovery/bundle tests, verification and test-family manifests | make task-guide ROLE=module-author OWNER=module.assessments; make explain-test-owner OWNER=module.assessments; make help-all; rg; sed | Existing coverage mapped; focused gaps recorded | RB-002 and parts of RB-001 | Add AS-S00 characterization in a later task |
| 2026-08-19T09:56:08-04:00 | Codex NLSpec revision | Passing pre-change and correction evidence are separated | Touched: tracker only. Inspected: owner rows, import facade, Imports ownership, tests | jq, rg, sed | Two service-backed rows specified; 15 to 16 to 17 | RB-002 | Implement AS-S00A without expected failures |
| 2026-08-19T10:41:11-04:00 | Codex remediation execution | AS-S00A pre-correction characterization complete | Touched: importassembly assessment integration test, assessment state/band contract test, Incident Bundles source-catalog test, Recovery state-catalog test, assessment owner manifest, PostgreSQL fixture policy, tracker. Inspected: existing Projections rebuild parity and Revisions rollback evidence | make author-test-row-id; make format-go; focused and full make test-slice/service-backed-test-slice; make explain-test-owner; make lint-markdown; git diff --check | PASS; 16 owner rows and 9 service-backed rows. Full focused run `.cartulary/test-results/20260819T144000Z-p2062685`; full service-backed run `.cartulary/test-results/20260819T144000Z-p2062692`; import row run `.cartulary/test-results/20260819T143408Z-p1990079`; state/band row run `.cartulary/test-results/20260819T143919Z-p2047777`; Incident Bundles row `.cartulary/test-results/20260819T143701Z-p2009815`; Recovery row `.cartulary/test-results/20260819T143701Z-p2009823`; Markdown lint `.cartulary/test-results/20260819T144131Z-p2144270`. Two earlier targeted runs failed on newly authored assertions and were corrected before the passing checkpoint | None; RB-002 resolved | Execute AS-S01 only after this checkpoint |
| 2026-08-19T10:52:19-04:00 | Codex remediation execution | AS-S01 assessor-membership correction complete | Touched: assessment facade/import port use, assessmentassembly validator, facade participant test, membership integration test, assessment owner manifest, PostgreSQL fixture policy, tracker | make author-test-row-id; make format-go; make test-slice/service-backed-test-slice for Assessments, Imports, and Incidents; make explain-test-owner; git diff --check | PASS; 17 owner rows and 10 service-backed rows. Dedicated row `.cartulary/test-results/20260819T144740Z-p2169959`; Assessments focused `.cartulary/test-results/20260819T144825Z-p2184874`; Assessments service-backed `.cartulary/test-results/20260819T144825Z-p2184883`; Imports focused/service-backed `.cartulary/test-results/20260819T144922Z-p2266334` and `.cartulary/test-results/20260819T145049Z-p2349487`; Incidents focused/service-backed `.cartulary/test-results/20260819T144922Z-p2266332` and `.cartulary/test-results/20260819T145049Z-p2349492`. The first dedicated-row run used a transaction-scoped fixture that cannot prove cross-connection blocking; the row was corrected to dedicated PostgreSQL and passed | None | Execute AS-S02 only after this checkpoint |
| 2026-08-19T10:57:37-04:00 | Codex remediation execution | AS-S02 canonical private policy complete | Touched: private assessment policy, create validation, workbook projection validation, projection derivation, rollback representation validation, policy contract test, assessment row selector, tracker | rg duplicate-policy search; make format-go; focused and service-backed assessment slices; make backend-unit; git diff --check | PASS; the only Go subject/state/band implementation is `assessments/internal/policy`; exact nil, -1, 0, 39, 40, 69, 70, 100, and 101 cases pass. Assessment focused/service-backed runs `.cartulary/test-results/20260819T145527Z-p2455903` and `.cartulary/test-results/20260819T145527Z-p2455906`; backend unit run `.cartulary/test-results/20260819T145623Z-p2537019` | None | Execute AS-S03 only after this checkpoint |
| 2026-08-19T11:03:39-04:00 | Codex remediation execution | AS-S03 private revision providers and path metadata complete | Touched: delete/restore and rollback providers/tests, root revision contribution, Revisions reflected-type expectation, backend boundary import/read/contract/metadata paths, tracker | rg before/after path inventory; make format-go; make backend-module-boundary-check; focused and service-backed Assessments/Revisions slices; git diff --check | PASS; old provider directories and imports are absent, root construction and snapshot/route behavior are unchanged. Assessments focused/service-backed `.cartulary/test-results/20260819T150037Z-p2645897` and `.cartulary/test-results/20260819T150131Z-p2686234`; Revisions focused/service-backed `.cartulary/test-results/20260819T145919Z-p2565092` and `.cartulary/test-results/20260819T150225Z-p2726486`. One parallel focused run hit a harness-only shared image-stamp race; the sequential retry passed | None | Execute AS-S04 only after this checkpoint |
| 2026-08-19T11:13:31-04:00 | Codex remediation execution | AS-S04 private projection provider and stable root constructor complete | Touched: projection provider/root contribution/tests, assessment and projection application composition, Projections authored guards/ownership expectations, boundary path, generated topology index, tracker | make format-go; backend boundary; focused and service-backed Assessments/Projections slices; make generate/generate-drift; git diff --check | PASS; Assessments focused/service-backed `.cartulary/test-results/20260819T150643Z-p2774512` and `.cartulary/test-results/20260819T150741Z-p2816721`; Projections focused/service-backed `.cartulary/test-results/20260819T151029Z-p2891178` and `.cartulary/test-results/20260819T151110Z-p2906407`. The first Projections run exposed an authored guard that did not recognize same-owner private projection paths; exact existing owner entries were added and the focused retry `.cartulary/test-results/20260819T150951Z-p2876300` passed. Initial drift identified the expected generated topology digest update and passed after Make-owned generation | None | Execute AS-S05-T only after this checkpoint |
| 2026-08-19T11:32:02-04:00 | Codex remediation execution | AS-S05-T typed fixture boundary cleanup complete | Touched: assessment and Projections testsupport, selected Projections SQL transaction evidence, Projections ownership test, boundary manifest/self-test, Projections row isolation metadata, generated topology index, tracker | make format; selected and full owner slices; backend boundary; make generate/generate-drift; make lint-markdown; rg; git diff --check | PASS; `SeedAssessment` remains stable, both transaction variants are exercised, and only Projections testsupport contains fixture SQL. Exact owner run roots are recorded in the backend checkpoint above; assessment owner accounting remains 17 rows and 10 service-backed rows; Markdown lint `.cartulary/test-results/20260819T153239Z-p3282004` | None | Execute AS-S05-B only after this checkpoint |
| 2026-08-19T11:43:28-04:00 | Codex remediation execution | AS-S05-B portability surface consolidation complete | Touched: private incident-bundle provider, root contribution wrappers, old portability deletion, exact boundary metadata/guard, tracker | make format; backend boundary; focused/service-backed Assessments, Incident Bundles, and Recovery slices; make lint-markdown; rg; git diff --check | PASS; only `NewIncidentBundleSourcePort` and `IncidentBundleSubtypeContribution` remain as root bundle construction surfaces, and existing bundle v2 behavior is unchanged. Exact run roots are recorded in the backend checkpoint above; Markdown lint `.cartulary/test-results/20260819T154411Z-p3501022` | None | Execute AS-S06 validation and handoff completion |
| 2026-08-19T12:20:51-04:00 | Codex remediation execution | AS-S06 validation and handoff complete | Touched: final harness accounting assertion and tracker; inspected every changed path, generated output, lockfiles, owner routing, and final status | make task-guide/explain-test-owner; focused/service-backed slices for Assessments, Imports, Incidents, Revisions, Projections, Incident Bundles, Recovery; backend boundary; harness/drift/artifact/shape/Markdown; frontend/browser; make agent-finalize; make test-fast; make check; make format; git diff/status | PASS; owner accounting is 17 rows/10 service-backed. Final focused/service roots: Assessments `.cartulary/test-results/20260819T154453Z-p3502483` / `.cartulary/test-results/20260819T154546Z-p3542646`; Imports `.cartulary/test-results/20260819T154959Z-p3679232` / `.cartulary/test-results/20260819T154846Z-p3638974`; Incidents `.cartulary/test-results/20260819T155111Z-p3719341` / `.cartulary/test-results/20260819T155235Z-p3762825`; Revisions `.cartulary/test-results/20260819T155848Z-p3892803` / `.cartulary/test-results/20260819T155737Z-p3850065`; Projections `.cartulary/test-results/20260819T160004Z-p3935568` / `.cartulary/test-results/20260819T160045Z-p3951003`; Incident Bundles `.cartulary/test-results/20260819T160123Z-p3966107` / `.cartulary/test-results/20260819T160223Z-p3981049`; Recovery `.cartulary/test-results/20260819T160319Z-p3995978` / `.cartulary/test-results/20260819T160437Z-p4045584`. Boundary `.cartulary/test-results/20260819T160610Z-p4095266`; finalize `.cartulary/test-results/20260819T161333Z-p1637`; final format `.cartulary/test-results/20260819T162207Z-p128720`; final Markdown lint `.cartulary/test-results/20260819T162214Z-p132371`; final 408-unit fast suite `.cartulary/test-results/20260819T162230Z-p133371`; 629-unit check `.cartulary/test-results/20260819T161430Z-p12586` | None. Imports first focused root `.cartulary/test-results/20260819T154638Z-p3582814` hit a transient pre-existing partial-cancellation deadlock; its isolated row `.cartulary/test-results/20260819T154806Z-p3624233` and both full reruns passed. Revisions first focused root `.cartulary/test-results/20260819T155401Z-p3805495` had an object-store readiness timeout before one row started; its service and full focused reruns passed | None; remediation is ready for review |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-19T08:29:33-04:00 | Codex tracker session | Workbook checks actor membership/role; assessment assessor validator checks only active user | Touched: tracker only. Inspected: Workbook routes, assessmentassembly validator, direct-reference contract, comparable member validators | rg, sed | Repository does not enforce same-incident membership for an explicit assessor | RB-001 | Characterize and separately authorize AS-S01 |
| 2026-08-19T09:56:08-04:00 | Codex NLSpec revision | Interface, admission, Auth check, ordering, and safe errors specified | Touched: tracker only. Inspected: validator, adapter, admission checker, removal, Auth store, errors | rg, sed | RolesMember/LifecycleAny plus locked active-user read selected | RB-001 | Authorize and implement AS-S01 atomically |
| 2026-08-19T10:52:19-04:00 | Codex remediation execution | Same-incident active-member assessor policy is enforced | Touched: assessor validator contract/adapter, interactive and import create paths, membership contract evidence | assessment/Imports/Incidents owner slices; dedicated concurrency row | Omitted and explicit assessors are materialized and checked through Incidents admission before Auth; unknown, inactive, nonmember, and foreign-only users return the existing safe assessor validation with no effects; exact replay and history remain intact; membership and user-state mutations wait for validation locks | None | Preserve this ordering during AS-S02 and provider moves |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-19T08:29:33-04:00 | Codex tracker session | Planning complete; implementation deliberately not started | Touched: tracker only. Inspected: all evidence summarized above | git status; make help-all and owner-explanation targets | Safe first implementation slice is AS-S00 | RB-001 through RB-004 | Start with characterization, then make one independently reversible slice |
| 2026-08-19T09:56:08-04:00 | Codex NLSpec revision | Plan is decision-complete; approvals explicit | Touched: tracker only. Inspected: all named evidence | sed, rg, git status, make lint-markdown, git diff HEAD --check | Tracker-only verification passed; AS-S00A is first; product work unauthorized | RB-001 through RB-004 | Request separate authorization for AS-S00A |
| 2026-08-19T12:20:51-04:00 | Codex remediation execution | Remediation complete; no known residual implementation risk | Final diff, status, compatibility, skipped-check, and requirement audit | git diff --check; git status --short --untracked-files=all; owner/static/frontend/browser/broad ladder above | Invalid active cross-incident assessor requests now intentionally fail with the existing safe validation response. Existing assessment history is not revalidated, rewritten, or backfilled. Repository-internal provider imports and portability helpers were removed without aliases. No public wire/storage/bundle/frontend contract changed. `make agent-finalize` ran with `RESULTS_DIR` unset because no successful full warm-check root predated finalization; retained-run maintenance was therefore skipped, while finalization itself passed | None; historical rows remain durable by design | Review and commit the completed change set |
| 2026-08-19T18:13:00-04:00 | Codex Imports/readiness follow-up | Product deadlock correction and harness diagnostic/propagation changes are complete; one independently reproduced object-store infrastructure residual remains open | Jobs/Imports concurrency, object-store lifecycle, service evidence, scheduler propagation, schemas, NLSpec, and this tracker | Focused owner rows; ten Imports repetitions; Revisions startup stress; harness contract/smokes; generation/static/broad gates | Imports no longer deadlocks and its selected row passed ten consecutive fresh runs. The eighth Revisions stress run reproduced a fully classified same-lane readiness timeout after seven passes; see Section 14 | None for the implemented changes; environmental/provider recurrence remains an explicit residual | Use the retained normalized diagnostic to investigate SeaweedFS/container/network stability; do not hide recurrence with timeout increases or automatic reruns |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Assessor correction is an existing-owner conformance fix | AS-S01 rejects requests live code accepts | Core 01 REQ-01-332/628 plus explicit remediation authority | RESOLVED |
| RB-002 | Import characterization is implemented and independently selectable | Moves require focused owner/Imports evidence | IMP-01 through IMP-09 and CHR-01 through CHR-05 passing, with adjacent selected evidence retained | RESOLVED |
| RB-003 | Provider topology follows adopted owner-contribution decisions | Provider visibility crosses Projections/Revisions boundaries | Projections typed-owner contribution and Revisions source-provider construction decisions | RESOLVED |
| RB-004 | Path reconciliation must follow real moves | Authored paths must follow before/after evidence | Reconcile atomically inside AS-S03, AS-S04, and AS-S05-B | RESOLVED BY SEQUENCING |

No authority, owner, implementation, validation, or handoff blocker remains.

## 12. Binary Completion Criteria

### Tracker acceptance

- PASS only if every current target file has exactly one inventory row.
- PASS only if every AMR-REQ maps to an exact scenario, assertion, or command.
- PASS only if every public contract has an owner and test posture.
- PASS only if workflows and slices state dependencies, rollback, validation,
  and binary completion.
- PASS only if AS-S00A adds passing evidence and no ignored, expected-failure,
  or current-acceptance defect test.
- PASS only if AS-S01 alone is identified as behavior-changing and its negative
  cases prove the complete applicable no-effect set.
- PASS only if 15 current rows and the 16/17 transitions are explicit.
- PASS only if the repaired framework mismatch, intentionally absent
  assessment NLSpec, inapplicable Core 05, and absence of owner contradiction
  are stated.
- PASS only if handoff tables preserve history and record every remediation
  checkpoint.
- PASS only if every authorized slice changes no surface outside its declared
  implementation, test, documentation, or authored-policy scope.

### Requirement-to-evidence traceability

| Requirement | Exact test scenario or static assertion | Required command or review |
| --- | --- | --- |
| AMR-REQ-001 | Plan-alignment diff names only tracker and planning framework; later diffs match their active slice | git diff --check; git status |
| AMR-REQ-002 | Tree assertion after moves; root/workbookprojection remain | Backend boundary check; owner review |
| AMR-REQ-003 | App import scan; constructor compile/use; no aliases | Assessment owner slice; backend boundary |
| AMR-REQ-004 | Port test passes incidentID; validator never completes transaction | ASR-01 through ASR-11 |
| AMR-REQ-005 | Store-backed admission and removal ordering | ASR-03 through ASR-09; membership integration row |
| AMR-REQ-006 | Defaulted/explicit assessor at both boundaries | ASR-01 through ASR-07; IMP-02/10 |
| AMR-REQ-007 | Participant spies/store assertions prove sequence | ASR-04 through ASR-09 |
| AMR-REQ-008 | Cases cover host/identity, every state, nil, 0, 39, 40, 69, 70, 100, -1, 101 | Existing state/projection/rollback/browser rows |
| AMR-REQ-009 | HTTP/facade cases cover field/default table | Existing integration/facade; ASR-01/02 |
| AMR-REQ-010 | IMP-06 proves rejection/no support; omission proves empty | Import facade integration row |
| AMR-REQ-011 | Exact HTTP status/code/field/reason and import safe code/reason | ASR-04 through 07; IMP-10 |
| AMR-REQ-012 | Replay spy and fresh identity after departure | ASR-10/11 |
| AMR-REQ-013 | Direct owner effects; Imports replay/conflict/journal/outcome/finalization | IMP-07 through 09 |
| AMR-REQ-014 | Store/event assertions cover negative-effect table | ASR-04 through 08/11; IMP-04/10 |
| AMR-REQ-015 | Existing route/query/bundle/recovery/frontend/browser/contract rows remain/pass | Owner/browser/frontend/drift commands as applicable |
| AMR-REQ-016 | Existing socket/event tests; no direct Collaboration import | test-fast; backend boundary |
| AMR-REQ-017 | Static import/SQL assertions match owner map | Backend boundary; owner review |
| AMR-REQ-018 | Before/after inventories account for every Section 3 consumer | Atomic AS-S03/AS-S04/AS-S05-B searches and boundary check |
| AMR-REQ-019 | JSON/selector comparison and generated roots clean | Generate drift; artifact policy; json shape |
| AMR-REQ-020 | Baseline 15; AS-S00A checkpoint reports 16 rows/9 service-backed; AS-S01 must report 17/10 | explain-test-owner and both slice commands |
| AMR-REQ-021 | Dependency audit matches the authorized eight-slice order | Handoff review before each slice |
| AMR-REQ-022 | Scoped commands pass and tracker is updated before dependents; broad evidence recorded | Section 8 ladder; final handoff |
| AMR-REQ-023 | Assessment testsupport contains no assessment projection-table SQL or boundary exception | Cross-owner fixture tests; backend boundary |
| AMR-REQ-024 | Root constructors remain; legacy helper exports and external imports are absent; bundle bytes remain exact | CHR-04; assessment/Incident Bundles slices; backend boundary |
| AMR-REQ-025 | Framework contains the Assessments owner row with exclusions matching the owner map | Markdown lint; owner review |

All AMR-REQ-001 through AMR-REQ-025 mappings above have passing implementation,
test, static, generated, documentation, or handoff evidence. AS-S00A through
AS-S06 are complete at 17 owner rows and 10 service-backed rows.

## 13. Final Changed-File Inventory

This inventory is the historical AS-S06 assessment-remediation snapshot. The
follow-up inventory and validation checkpoint in Section 14 extend it without
rewriting the assessment handoff history. No lockfile is changed.

```text
AM docs/handoffs/assessments-module-refactor-tracker.md
 M docs/handoffs/cartulary_modular_refactor_planning_framework.md
 M internal/app/assessmentassembly/adapters.go
 M internal/app/assessmentassembly/projection_source.go
 M internal/app/projectionassembly/source_ownership_test.go
 M internal/app/recoveryassembly/state_catalog_test.go
 M internal/modules/assessments/assessment_contract_test.go
 D internal/modules/assessments/deleterestore/provider.go
 M internal/modules/assessments/facade.go
 M internal/modules/assessments/facade_contract_test.go
 M internal/modules/assessments/import_create.go
 D internal/modules/assessments/incident_bundle_portability.go
 M internal/modules/assessments/incident_bundle_source_port.go
 M internal/modules/assessments/incident_bundle_subtype_presence.go
 D internal/modules/assessments/projectionprovider/provider.go
 M internal/modules/assessments/revision_provider_contribution.go
 D internal/modules/assessments/rollbackprovider/provider.go
 D internal/modules/assessments/rollbackprovider/provider_test.go
 M internal/modules/assessments/testsupport/assessments.go
 M internal/modules/assessments/workbookprojection/model.go
 M internal/modules/incidentbundles/source_catalog_test.go
 M internal/modules/projections/adapters/boundary_guard_test.go
 M internal/modules/projections/internal/runtime/query_store_test.go
 M internal/modules/projections/testsupport/capability.go
 M internal/modules/revisions/delete_restore_test.go
 M tools/backend_module_boundaries.json
 M tools/execution_topology_render_index.json
 M tools/harness/static-analysis/backend-module-boundary-check-cli.mjs
 M tools/harness/tests/contract-suite-support.mjs
 M tools/postgres_fixture_policy_registry.json
 M tools/test_families/module.assessments.json
 M tools/test_families/module.projections.json
?? internal/app/importassembly/assessment_integration_test.go
?? internal/app/importassembly/assessment_membership_integration_test.go
?? internal/modules/assessments/internal/policy/assessment.go
?? internal/modules/assessments/internal/providers/deleterestore/provider.go
?? internal/modules/assessments/internal/providers/incidentbundle/portability.go
?? internal/modules/assessments/internal/providers/incidentbundle/source_port.go
?? internal/modules/assessments/internal/providers/incidentbundle/subtype_presence.go
?? internal/modules/assessments/internal/providers/projection/provider.go
?? internal/modules/assessments/internal/providers/rollback/provider.go
?? internal/modules/assessments/internal/providers/rollback/provider_test.go
?? internal/modules/assessments/policy_contract_test.go
?? internal/modules/assessments/projection_provider_contribution.go
?? internal/modules/assessments/projection_provider_contribution_test.go
```

## 14. Imports Deadlock and Object-Store Readiness Follow-up

### Authority and compatibility

Core REQ-01-249, REQ-01-453, and REQ-01-620c remain the unchanged product
authority for cancellation and partial application. The Testing Harness NLSpec
now makes container startup and readiness separate bounded phases and adds
TH-HARNESS-REQ-414 for safe readiness diagnostics and exact cross-process
failure propagation. No Core, HTTP, database, Imports state, public Go,
provider, timeout, or retry-policy contract changed.

The private Jobs lock hierarchy is now transition advisory lock before job-row
and route-idempotency locks. All live cancellation callers already hold that
lock and use the private lock-held helper; an initially retained no-lock
wrapper was removed because the completed caller audit proved it unreachable
and static analysis correctly rejected the dead code.

The internal `cartulary-test-services start-suite` command and its sole broker
caller migrated atomically to the required mode-0600
`cartulary.test_services.start_result.v1` handshake. Existing retained service
scope v2 artifacts require no migration because the bounded
`cartulary.object_store_readiness_diagnostic.v1` attachment uses the existing
optional extension surface.

### Implemented workstreams

| Workstream | Remediation and validation | Status |
| --- | --- | --- |
| Authority alignment | Clarified the post-start 120-second readiness window, terminal polling behavior, safe diagnostic fields, and failure propagation in the Testing Harness NLSpec; registered both new schemas | COMPLETE |
| Jobs lock order | `Manager.Cancel` acquires the transition advisory lock before the visible job row and idempotency state; inactive completion uses the explicit lock-held helper; both winner orderings are covered | COMPLETE |
| Imports synchronization | Replaced the global advisory key and SQL-substring waiter inference with test-unique two-part keys, exact `pg_locks` identity, `pg_blocking_pids`, current-database filtering, and bounded lock graphs | COMPLETE |
| Readiness diagnostics | Added typed phase/stage/outcome/cause/container/cleanup diagnostics with strict bounds, deterministic encoding, malformed/forbidden-field rejection, and no endpoint, credential, resource identity, raw error, or log retention | COMPLETE |
| Object-store lifecycle | Startup retry now covers only creation and endpoint resolution; readiness receives one fresh 120-second context on the same lane; cleanup has an independent bound and cannot replace the primary failure | COMPLETE |
| Failure propagation | The broker validates owner-only files, run-root containment, identities, current scope agreement, and schema before removing the private result; unit, target, and run evidence preserve the exact class, reason, exit 3, and service-scope reference | COMPLETE |

### Focused and stress evidence

- Jobs lifecycle/lock-order row passed at
  `.cartulary/test-results/20260819T175411Z-p808692`.
- The Imports partial-cancellation row passed ten consecutive fresh canonical
  runs at `.cartulary/test-results/20260819T172858Z-p275422`,
  `20260819T172943Z-p290210`, `20260819T173019Z-p304928`,
  `20260819T173055Z-p319632`, `20260819T173129Z-p334355`,
  `20260819T173204Z-p349056`, `20260819T173239Z-p363764`,
  `20260819T173313Z-p378484`, `20260819T173349Z-p393370`, and
  `20260819T173424Z-p408087`. Its final post-format check also passed at
  `.cartulary/test-results/20260819T175450Z-p823921`.
- Object-store fixture-admission and test-services lifecycle rows passed at
  `.cartulary/test-results/20260819T175526Z-p838658`; the harness contract
  passed at `.cartulary/test-results/20260819T175342Z-p807251`.
- Revisions startup stress produced seven consecutive 27/27 owner passes at
  `.cartulary/test-results/20260819T173504Z-p422833`,
  `20260819T173607Z-p466887`, `20260819T173710Z-p509714`,
  `20260819T173816Z-p552443`, `20260819T173921Z-p595125`,
  `20260819T174025Z-p637951`, and `20260819T174125Z-p680638`.
  The eighth run, `.cartulary/test-results/20260819T174226Z-p723312`,
  reproduced the infrastructure failure, so the ten-consecutive-pass exit
  criterion was not claimed.

The reproduced failure is classified `infra/service_readiness_timeout` with
public exit 3 at the unit, target, and run layers, all referencing
`_shared/test-services/ac4a2bfd0e74aadb6a7e734b/service-scope.json`. Its safe
diagnostic records `initial_lane/list/deadline_expired`, 33 attempts over the
full 120000 ms window, three operation timeouts, 30 transport-unreachable
causes, and a still-running, non-OOM container. Startup itself succeeded on the
first attempt in 311 ms. This localizes the open issue to the SeaweedFS lane,
container runtime, or host/network environment rather than Revisions product
behavior or overlapping harness deadlines.

### Final validation and attributed failures

| Command | Result and retained evidence |
| --- | --- |
| `make task-guide ROLE=module-author OWNER=platform.jobs` and matching Imports/harness.browser guides | PASS; routing confirmed |
| `make run-harness-smoke-execution` | FAIL at `.cartulary/test-results/20260819T175102Z-p795339`; two pre-existing assertion drifts remain: fixture-strategy aggregate text and canonical unit-result output inclusion. Neither implicated source/model file was changed by this follow-up |
| `make run-harness-smoke-lifecycle` | FAIL at `.cartulary/test-results/20260819T175128Z-p801806`; the pre-existing web-E2E lifecycle assertion expects an early-return readiness block to be absent although the live script contains it. Neither assertion nor script was changed here |
| `make generate` | PASS at `.cartulary/test-results/20260819T175639Z-p854731` |
| `make generate-drift` | PASS at `.cartulary/test-results/20260819T181340Z-p1177422` |
| `make generated-artifact-policy-check` | PASS at `.cartulary/test-results/20260819T181348Z-p1180305` |
| `make json-shape-check` | PASS at `.cartulary/test-results/20260819T181349Z-p1180715` |
| `make lint-markdown` | PASS at `.cartulary/test-results/20260819T181352Z-p1181185` |
| `make agent-finalize` with `RESULTS_DIR` unset | PASS at `.cartulary/test-results/20260819T180725Z-p1060193`; retained-run maintenance was intentionally skipped |
| `make backend-unit` | PASS, 111/111, at `.cartulary/test-results/20260819T175952Z-p903597` |
| `make test-fast` | PASS, 408/408, at `.cartulary/test-results/20260819T180117Z-p927476` |
| `make check` | PASS, 629/629, at `.cartulary/test-results/20260819T180744Z-p1063067` |

Two intermediate quality-gate findings were fixed rather than attributed: the
first `make test-fast` found a local diagnostics-test redeclaration, and the
first `make check` found the audited no-lock cancellation wrapper was unused.
The passing broad runs above include both corrections. An earlier finalizer
also correctly detected stale topology metadata after a scheduler edit;
Make-owned regeneration repaired it before the passing finalizer.

### Follow-up changed-file scope

- Product concurrency: `internal/platform/jobs/{jobs_test.go,lifecycle_persistence.go,stored_job.go,transaction_service.go,transition_persistence.go}`
  and `internal/modules/imports/imports_integration_test.go`.
- Object-store lifecycle and evidence:
  `internal/testutil/s3test/{s3test.go,s3test_test.go}`,
  `internal/testutil/suiteservices/{diagnostics.go,diagnostics_test.go}`, and
  `tools/testservices/{main.go,main_test.go}`.
- Cross-process propagation:
  `tools/harness/contract/index.mjs`, scheduler fixture-broker/work-graph
  sources, local-session sources/tests, and harness contract support.
- Contract projections: Testing Harness NLSpec, schema attachment inventory,
  service-scope schema, two new strict schemas, affected authored test-family
  rows, and the Make-generated topology render index.

No generated protocol root, lockfile, public product contract, or unrelated
product implementation was changed. The only residual is the explicitly
retained object-store infrastructure recurrence above; it must be investigated
from the normalized evidence rather than masked by more retries, longer
timeouts, or a provider substitution.

## 15. Testing Harness Gap Remediation Closure

### Authority, causal finding, and compatibility

The adopted Testing Harness NLSpec remains authoritative. TH-HARNESS-REQ-411
now defines the exact fixture strategy aggregate as service, target, operation,
preparation strategy, fixture policy, fixture class, and reuse scope. Caller
package and test identities remain separate hotspot and slowest-event
dimensions. `docs/domain.md` is unchanged because no domain vocabulary or owner
navigation changed. Canonical unit-result publication and fail-closed browser
readiness were already authoritative and required no compatibility path.

Private recurrence capture identified whole-host pressure as the object-store
readiness cause. The failed broad replay at
`.cartulary/test-results/20260819T192915Z-p2567279` reached 62.60% CPU and
34.61% I/O pressure while SeaweedFS remained running, non-OOM, and
non-restarting. Focused Revisions capture peaked at 0.36% CPU pressure, and the
final ten-run series peaked at 0.18%. Reducing object-store lanes from four to
three alone did not prevent recurrence, so the final policy also reserves a
portable 25% CPU admission margin: detected capacity is transformed with
`floor(detected * 100 / 125)`, bounded to at least one token. This host now
admits 19 of 24 detected CPU tokens. The policy is authored in the scheduler
resource registry and covered by deterministic 24-to-19, 4-to-3, and 1-to-1
contract vectors.

No Revisions product, HTTP, database, object-storage API, provider image,
120-second readiness deadline, retry count, fallback provider, replacement
lane, or automatic failed-run rerun changed. The expected compatibility effect
is lower admission concurrency and potentially longer broad-suite wall time.
The service-scope v2 producer and schema no longer retain service host, port,
endpoint, Docker endpoint, container identity, raw error, or log-tail fields;
these test-only ephemeral diagnostics require no product or data migration.

### Implemented workstreams

| Workstream | Remediation | Status |
| --- | --- | --- |
| Object-store infrastructure | Added normalized typed readiness text, preserved the readiness failure across cleanup failure, propagated `cleanup_outcome=failed`, removed raw endpoint/error fields at the event source, reduced object-store lanes to three, and added the 25% CPU admission reserve | COMPLETE |
| Execution smoke | Aligned Go and JavaScript aggregation with the exact seven-field identity; replaced the malformed tie fixture with two distinct equal-duration aggregates; asserted structured ordering and exact canonical unit-result outputs for dedicated and migration batches | COMPLETE |
| Browser lifecycle | Replaced the multiline source `grep` with behavioral missing-readiness, missing-diagnostic, and publication-verification cases; retained the fail-closed guard; made publication and final verification status propagation explicit | COMPLETE |
| Integrated harness stability | Isolated the process-wide umask proof in a child process so concurrent cache-mode assertions cannot observe transient global mode changes | COMPLETE |

### Focused, stress, and retained-evidence proof

- Latest generation and contracts passed at
  `.cartulary/test-results/20260819T195458Z-p2956341` (`make generate`),
  `20260819T195649Z-p2963636` (JSON shape),
  `20260819T195652Z-p2964042` (generated-artifact policy),
  `20260819T195654Z-p2964442` (generation drift), and
  `20260819T195702Z-p2967381` (harness contract). The focused generated-drift
  row passed at `20260819T195714Z-p2967907`.
- Harness Browser passed 28/28 at
  `.cartulary/test-results/20260819T195729Z-p2970600` and service-backed 6/6 at
  `20260819T195830Z-p2988521`. Execution and lifecycle smoke passed at
  `20260819T201330Z-p3524598` and `20260819T201350Z-p3531976`.
- Strict SeaweedFS compatibility passed 3/3 at
  `.cartulary/test-results/20260819T200006Z-p3003898`; the Revisions
  service-backed slice passed 20/20 at `20260819T200040Z-p3020141`.
- Ten consecutive fresh Revisions owner slices passed 27/27 at
  `.cartulary/test-results/20260819T200149Z-p3066306`,
  `20260819T200255Z-p3112563`, `20260819T200357Z-p3158211`,
  `20260819T200459Z-p3203782`, `20260819T200601Z-p3249443`,
  `20260819T200708Z-p3295285`, `20260819T200814Z-p3341195`,
  `20260819T200922Z-p3387096`, `20260819T201024Z-p3432674`, and
  `20260819T201126Z-p3478204`. No failure reset was required.
- The final recurrence capture observed 12 balanced SeaweedFS and PostgreSQL
  create/start/die/destroy lifecycles, zero capture errors, zero OOM or restart
  evidence, and 0.18%/0.00%/14.55% maximum CPU/memory/I/O ten-second pressure.
  All raw private captures were owner-only, reduced to these aggregate facts,
  and deleted.
- Scans of every final Revisions service scope and service journal found zero
  endpoint, socket, container, raw-error, or log-tail fields. Canonical browser
  stack publications retain their NLSpec-required allocated ports and are not
  service-readiness diagnostics. Docker reported zero managed test-container
  residue after the browser and Revisions service-backed runs.

### Integrated validation and failure attribution

`make agent-finalize` with `RESULTS_DIR` unset passed at
`.cartulary/test-results/20260819T201358Z-p3533720`; retained-run maintenance
was intentionally skipped. `make test-fast` passed 408/408 at
`20260819T201411Z-p3536549`. The final `make check` passed 629/629 at
`20260819T201632Z-p3577758`, retained-run finalization against that root passed
at `20260819T202136Z-p3691112`, and `make release-check` passed 786/786 at
`20260819T202204Z-p3694098`.

Intermediate failures were causal evidence or related harness findings, not
waived gates. The 4-lane check at `20260819T191722Z-p2387068` reproduced the
readiness timeout; the 3-lane replay at `20260819T192915Z-p2567279` proved that
service-lane reduction alone could not protect the daemon under whole-host
pressure. The policy-corrected replay at `20260819T194604Z-p2794172` passed the
object-store admission row in 23.04 seconds but exposed the concurrent umask
test race and a transient generated-drift snapshot failure. The umask mutation
was isolated structurally; generation drift and its exact row then passed
narrowly and both passed in the final broad graph. An earlier ShellCheck
finding was fixed with scoped annotations rather than suppressing the scripts.

The three testing-harness gaps are closed. No raw investigation capture,
unexpected generated root, lockfile, provider substitution, or Revisions
product change remains in the worktree.

## 16. APR Iteration Scope and Planning Posture

APR is the Assessments Production-Readiness cleanup iteration. Sections 1-15
remain the immutable execution and handoff history for the completed module
remediation and its testing-harness follow-up. APR does not reopen either
effort. In particular, the transient failures recorded in Sections 14 and 15
were corrected and closed; they are not APR requirements, risks, or planned
work.

This tracker is an execution-support artifact, not product authority. Core
owners continue to define the behavior preserved by this iteration. No
Assessment NLSpec, Core owner, `docs/domain.md`, OpenAPI document, view schema,
bundle contract, frontend contract, or public protocol change is planned.

### 16.1 Planning baseline

| Baseline fact | APR-S00 record |
| --- | --- |
| Commit | `6cb809190bf107e805720a9552a894783db10cbc` |
| Date | 2026-08-19 |
| Assessments file inventory | 27 files |
| Assessments owner rows | 17 total |
| Assessments service-backed rows | 10 |
| Go lint | `make lint-go` passed |
| Assessments slice | `make test-slice OWNER=module.assessments` passed 23/23 |
| Assessments slice result root | `.cartulary/test-results/20260819T204641Z-p3914346` |
| Prior transient failures | Closed in Sections 14-15; excluded from APR scope |

The baseline is frozen for planning purposes. Migration 35 is the next legal
migration on this baseline. If another change occupies that number before
APR-S01 is authorized, execution must stop and rebaseline this tracker rather
than silently renumbering the planned migration.

### 16.2 Iteration objective and retained surfaces

APR removes compatibility behavior, duplicate orchestration, dead merge state,
panic-based construction, and permissive adapter normalization that do not
serve the production architecture. It also installs a narrow export guard so
the cleaned package surface remains intentional as later phases expand.

The following surfaces have clear continuing value and remain:

- the Assessments `Facade` and typed cross-owner ports;
- the `workbookprojection` contract surface;
- `NewProjectionContribution`;
- `RevisionProviderContribution`;
- `RecoveryStateContribution`;
- `NewIncidentBundleSourcePort`;
- `IncidentBundleSubtypeContribution`;
- dual-driver assessment test support required by current cross-owner tests.

APR is repository-internal cleanup except for the persisted idempotency data
migration. It does not authorize a database domain-schema change, public API
change, route change, bundle-version change, generated protocol change, or
frontend behavior change. Internal removals receive no aliases, `Must`
constructors, or deprecation window.

### 16.3 Authorization boundary

APR-S00 was completed as a tracker-only planning slice. The 2026-08-19
implementation request authorizes APR-S01 through APR-S05. Authorization does
not bypass the dependency gates: each successor remains ineligible until the
preceding implementation, focused validation, and tracker checkpoint are
complete.

## 17. APR Requirements and Gap Register

### 17.1 Requirements

| Requirement | Planned outcome | Primary evidence |
| --- | --- | --- |
| APR-REQ-001 | Preserve Sections 1-15 as completed historical evidence and treat their transient failures as closed | Tracker review |
| APR-REQ-002 | Execute APR serially and checkpoint this tracker after every authorized workstream before starting its successor | Section 18 ledger |
| APR-REQ-003 | Preserve valid persisted assessment-create replays while converging storage and runtime decoding on one strict v1 payload | Codec and migration tests |
| APR-REQ-004 | Reject malformed, unknown, or internally inconsistent persisted payloads before migration mutation, with aggregate-only diagnostics | Migration preflight tests |
| APR-REQ-005 | Replace Assessments-specific import finalization with the Imports-owned generic revision-and-intent finalizer and field mapper | IMP-01 through IMP-10 |
| APR-REQ-006 | Remove unused merge API, version state, and query state without changing merge results, history, tombstones, timestamps, refresh, or rollback | Entities, Revisions, Timeline, and Assessments tests |
| APR-REQ-007 | Make merge dependency failures ordinary constructor errors and propagate them through application composition without a `Must` helper | Construction and composition tests |
| APR-REQ-008 | Accept only positive `int64` row versions from the live projection boundary | Facade unit tests |
| APR-REQ-009 | Keep serialization-specific number handling inside persisted-payload and Imports-owned codecs | Codec and import tests |
| APR-REQ-010 | Lock the final Assessments root and `workbookprojection` exports with an exact AST allowlist and negative fixture | Export-guard tests |
| APR-REQ-011 | Retain the five root contribution surfaces, `Facade`, typed ports, `workbookprojection`, and dual-driver test support | Export allowlist and boundary checks |
| APR-REQ-012 | Add exactly one authored unit row and one service-backed migration row in APR-S01, reaching 19 total owner rows and 11 service-backed rows | Owner routing and generated topology checks |
| APR-REQ-013 | Preserve all public HTTP, WebSocket, OpenAPI, view-schema, bundle, frontend, field-key, and generated-protocol behavior | Affected owner and broad validation |
| APR-REQ-014 | Remove retired internal APIs without aliases or a compatibility layer | Exact symbol and import searches |
| APR-REQ-015 | Complete APR only after every requirement and gap is traced to passing evidence or an explicitly attributed failure | APR-S05 handoff |

### 17.2 Gap decisions

| Gap | Remediation and areas | Rationale and long-term benefit | Compatibility or migration impact | Risk if unresolved | Completion validation |
| --- | --- | --- | --- | --- | --- |
| Legacy assessment idempotency decoder remains indefinitely | **Migration, implementation, tests:** add `00035_assessment_create_idempotency_v1.sql`; migrate valid legacy `assessments.rows.create` payloads to the canonical v1 shape; remove the runtime fallback and accept only strict v1 payloads | `route_idempotency` rows do not expire. A one-time migration bridge preserves valuable replay history while leaving one durable runtime representation | Valid persisted replays remain valid. Malformed or unknown formats block migration with aggregate-only diagnostics. Down migration mechanically reconstructs the old response shape | Blind fallback removal can allow duplicate assessments; permanent dual-format decoding weakens corruption handling and increases maintenance | Up/Down/Up passes; canonical rows remain unchanged; valid legacy rows replay identically; malformed rows fail without mutation or sensitive diagnostics; no legacy decoder remains |
| Assessment-specific import finalization duplicates Imports' generic finalizer | **Implementation, tests:** use `ownerfacade.RecordRevisionAndIntentAppender`, `FinalizeRecordRevisionAndIntentTx`, and `ownerfacade.ValuesByField`; remove the local revision DTO, appender, adapter, alias, and map helper | One Imports-owned finalization path keeps mutation, revision, intent, and result semantics coherent as Imports evolves | Repository-internal Go break only; no alias. Import results and transaction behavior remain exact | Assessments can drift from every other import owner and require duplicate maintenance | IMP-01 through IMP-10 pass; retired types and helpers have zero references; Imports and Assessments slices pass |
| Merge participant carries unused API and query state | **Implementation, tests:** remove `ErrMergeProtectedSetChanged`, its `Unwrap`, unused version IDs, and unused selected/scanned timestamps; retain the typed error and `RecordID`, actual `updated_at` write, and exact tombstone fields | Contracts the participant to facts that influence merge behavior and removes false extension points | Internal DTO break with atomic Entities adapter updates; public merge behavior and history are unchanged | Dead fields imply unsupported semantics and create needless merge/revision coupling | Protected-set, repointing, rollback, snapshot, and revision tests remain exact; retired symbols and selected columns are absent |
| Merge construction panics on invalid dependencies | **Implementation, tests:** change `NewMergeEffects` and its application wrapper to return `(*MergeEffects, error)`; propagate errors through `timelineassembly.NewBundle`; add no `Must` helper | Predictable startup errors remain composable as the dependency graph grows | Internal constructor migration only; valid startup behavior is unchanged | A future composition path can convert a configuration defect into a process panic | Nil dependency cases return stable errors; valid composition and Timeline/Entities tests pass; no constructor panic remains |
| Live projection row-version parsing accepts types production never emits | **Implementation, tests:** accept only positive `int64` from the interactive facade; reject `int`, `float64`, zero, and negative values; retain JSON-number handling only in persisted-payload and Imports-owned codecs | Separates a typed live boundary from serialization concerns and exposes adapter drift immediately | Stricter internal failure for malformed adapters only | Broken adapters can be silently normalized and remain hidden | Positive `int64` succeeds; every other type or value fails before revision or idempotency effects |
| Final exported surface is not permanently locked | **Tests and harness:** add an AST-based exact allowlist for the root and `workbookprojection`, including a negative fixture | Prevents accidental compatibility-surface growth after cleanup while private providers and test support remain evolvable | Test-only enforcement; this iteration adds no production export | Retired aliases and convenience APIs can return unnoticed | Exact allowlist passes and an injected export fails |

### 17.3 Interface and compatibility decisions

The authorized implementation iteration will make these repository-internal
changes:

- `NewMergeEffects(...)` becomes `(*MergeEffects, error)`.
- `ImportCreateDependencies.Revisions` uses the Imports-owned generic
  revision-and-intent appender.
- `ImportCreateCommand`, `ImportRevision`, `ImportRevisionAppender`,
  `ErrMergeProtectedSetChanged`, its `Unwrap`, and the two unused merge version
  fields are removed.
- Live row-version acceptance narrows to positive `int64` values.

Valid persisted replay history has continuing value and is migrated. Malformed
or unknown stored payloads are neither deleted nor silently ignored. Existing
public behavior and the retained contribution surfaces in Section 16.2 remain
unchanged.

## 18. APR Serial Workstreams

Sequence:

`APR-S00 -> APR-S01 -> APR-S02 -> APR-S03 -> APR-S04 -> APR-S05`

| Slice | Purpose | Dependency | Status | Authorization |
| --- | --- | --- | --- | --- |
| APR-S00 | Tracker rebaseline and decision-complete plan | Completed Sections 1-15 | COMPLETE | Authorized documentation-only slice |
| APR-S01 | Persisted idempotency cutover | APR-S00 checkpoint | COMPLETE | Checkpoint recorded in Section 20.3 |
| APR-S02 | Import finalization consolidation | APR-S01 checkpoint | COMPLETE | Checkpoint recorded in Section 20.4 |
| APR-S03 | Merge participant contraction | APR-S02 checkpoint | COMPLETE | Checkpoint recorded in Section 20.5 |
| APR-S04 | Fail-safe construction, strict live row version, and export lock | APR-S03 checkpoint | COMPLETE | Checkpoint recorded in Section 20.6 |
| APR-S05 | Validation, traceability, and final handoff | APR-S04 checkpoint | COMPLETE | Checkpoint and handoff recorded in Section 20.7 |

### 18.1 Checkpoint protocol

Every authorized implementation slice is serial. After its code and focused
checks pass, and before its successor begins, update this tracker with:

- status and requirement-to-evidence changes;
- every authored, moved, generated, or deleted file;
- exact commands, results, and result or artifact roots;
- owner-row and service-backed-row counts when changed;
- compatibility and migration effects;
- generated outputs and lockfile status;
- failures, causal attribution, skipped checks, and residual risks;
- the next eligible slice and its authorization state.

Do not advance with a failing focused test, stale path, occupied migration
number, hand-edited generated artifact, incomplete checkpoint, or unexplained
failure. A checkpoint never grants authority for its successor.

### 18.2 APR-S00 - Tracker rebaseline

Update the header and append the APR scope, baseline, requirements, gap
decisions, workstreams, validation matrix, checkpoint ledger, and handoff
template. Preserve Sections 1-15 byte-for-byte except for the header posture
needed to distinguish the completed work from this new planned iteration.

Exit criteria:

- only this tracker changed;
- `make lint-markdown` passes;
- `git diff --check` passes;
- APR-S00 is `COMPLETE`; and
- APR-S01 remains `PLANNED` and explicitly authorization-gated.

### 18.3 APR-S01 - Persisted idempotency cutover

Add `db/migrations/00035_assessment_create_idempotency_v1.sql` and a strict
canonical assessment-create payload codec. The migration must first classify
all `assessments.rows.create` payloads without mutation. It may update only
valid legacy rows; existing canonical v1 rows remain byte-for-byte or
semantically unchanged as required by the codec contract. Unknown schemas,
malformed values, or redundant-field disagreements fail the migration with
aggregate-only diagnostics that reveal no incident, actor, assessment,
version, or payload value.

The canonical decoder validates the schema identifier, exact top-level member
set, UUID fields, positive row version, row presence, and redundant
record/version agreement. After migration evidence passes, remove the legacy
runtime fallback. Down migration mechanically reconstructs the former response
shape so Up/Down/Up is deterministic.

Use `make author-test-row-id` rather than editing generated topology. Add one
authored Assessments unit row for codec behavior and one service-backed
Database Migrations row for stored-data conversion. Expected Assessments
routing after generation is 19 total owner rows and 11 service-backed rows.

Exit criteria:

- strict round trip and malformed, unknown, mismatched, and boundary matrices
  pass;
- valid legacy, existing-v1, malformed-preflight, safe-diagnostic,
  Up/Down/Up, and replay-equivalence migration cases pass;
- no legacy decoder or permissive fallback remains;
- Assessments, Workbook, and Database Migrations checks pass; and
- migration drift, generation drift, and the APR-S01 tracker checkpoint pass.

### 18.4 APR-S02 - Import finalization consolidation

Replace the Assessments-local import finalization path with
`ownerfacade.RecordRevisionAndIntentAppender`,
`FinalizeRecordRevisionAndIntentTx`, and `ownerfacade.ValuesByField`. Delete
the local import revision DTO, appender interface,
`assessmentImportRevisionAdapter`, `ImportCreateCommand` alias, and field-map
helper without replacements.

Preserve borrowed-transaction ownership, validation order, projection refresh,
mutation identity, revision values, intent contents, result codes, and rollback
behavior. Do not introduce a second generic abstraction in Assessments.

Exit criteria:

- IMP-01 through IMP-10 pass, including rollback and negative effects;
- retired types, aliases, adapters, and helpers have zero references;
- focused and service-backed Assessments and Imports slices pass; and
- the APR-S02 tracker checkpoint is complete.

### 18.5 APR-S03 - Merge participant contraction

Remove `ErrMergeProtectedSetChanged` and its `Unwrap`, retaining the typed
protected-set error and `RecordID`. Remove unused `BeforeVersionID` and
`AfterVersionID` state and update the Entities adapter atomically. Stop
selecting and scanning `created_at` and `updated_at` when those values do not
affect the mutation result.

The cleanup must retain the actual `updated_at` write, exact tombstone fields,
`protected_set_changed` behavior, record detail, snapshots, canonical before
and after values, projection refresh, repointing, revision history, and
rollback.

Exit criteria:

- protected-set race, repointing, snapshot, rollback, and history parity pass;
- removed API, fields, and selected timestamp columns are absent;
- Assessments, Entities, Revisions, and Timeline slices pass; and
- the APR-S03 tracker checkpoint is complete.

### 18.6 APR-S04 - Fail-safe construction and surface lock

Change `NewMergeEffects` and its application wrapper to return
`(*MergeEffects, error)`. Propagate errors through `timelineassembly.NewBundle`
and every composition caller. Nil dependencies return stable errors; do not add
a `Must` constructor or preserve a panic path.

Constrain the live projection row-version helper to positive `int64` values.
Reject `int`, `float64`, zero, negative, and other values before any revision or
idempotency effect. Serialization-specific numeric handling remains inside the
strict persisted-payload codec and Imports-owned generic finalization.

Add an AST-based exact export allowlist for the Assessments root and
`workbookprojection`, with a negative fixture proving that an injected export
fails. Extend the existing Assessments unit owner row with constructor,
row-version, and export-lock evidence; do not add another catalog row.

Exit criteria:

- nil dependency cases return errors and valid Timeline composition passes;
- the row-version acceptance and rejection matrix passes;
- the exact export allowlist and negative fixture pass;
- `make lint-go`, affected owner slices, module-boundary checks, and exact
  retired-symbol searches pass; and
- the APR-S04 tracker checkpoint is complete.

### 18.7 APR-S05 - Validation and handoff

Reconcile every APR requirement, gap, removal, retained surface, and
compatibility decision against evidence. Run the final ladder in Section 19,
record every result root and failure attribution, inspect the final diff, and
complete the handoff. No unexplained failing gate may be waived.

Exit criteria:

- APR routing is exactly 19 total Assessments rows and 11 service-backed rows;
- every APR requirement and gap maps to passing evidence;
- all failures are fixed or explicitly attributed to a documented external or
  unrelated cause without waiving an applicable gate;
- generated and lockfile inspection is clean and no unrelated change remains;
  and
- all APR slices and the iteration status are `COMPLETE`.

## 19. APR Validation and Acceptance Matrix

### 19.1 Slice evidence

| Slice | Required evidence | Required narrow owners or gates |
| --- | --- | --- |
| APR-S00 | Tracker-only diff and Markdown validity | `make lint-markdown`; `git diff --check`; final status inspection |
| APR-S01 | Strict codec matrix; migration classification, safe preflight, Up/Down/Up, and replay equivalence; exact 19/11 routing | Assessments, Workbook, Database Migrations; `make migration-drift`; generation and topology drift |
| APR-S02 | IMP-01 through IMP-10; borrowed transaction, revision, intent, result, refresh, and rollback parity | Assessments and Imports focused and service-backed slices |
| APR-S03 | Protected-set race, repointing, snapshot, canonical values, history, refresh, and rollback parity | Assessments, Entities, Revisions, and Timeline slices |
| APR-S04 | Constructor errors, valid composition, strict live row version, export allowlist positive and negative cases, retired-symbol absence | `make lint-go`; affected owner slices; `make backend-module-boundary-check` |
| APR-S05 | Complete requirement traceability and broad regression evidence | Final ladder and handoff inspection |

### 19.2 Final validation ladder

Run from narrow routing evidence to broad release evidence:

1. Run `make task-guide ROLE=module-author OWNER=module.assessments` and
   `make explain-test-owner OWNER=module.assessments`; confirm 19 owner rows and
   11 service-backed rows.
2. Run focused and service-backed slices for Assessments, Imports, Entities,
   Revisions, Timeline, Workbook, Projections, and Database Migrations.
3. Run `make backend-module-boundary-check`, `make harness-contract`,
   `make migration-drift`, `make generate-drift`,
   `make generated-artifact-policy-check`, `make json-shape-check`, and
   `make lint`.
4. Run `make frontend-typecheck`, `make frontend-unit`,
   `make frontend-import-boundary-check`, and
   `make browser-e2e-webserver-backed`.
5. Run `make agent-finalize`, then `make test-fast`, `make check`, and
   `make release-check`. If no retained successful full warm-check root exists,
   leave `RESULTS_DIR` unset and record that retained-run maintenance was
   skipped.
6. Finish with `make lint-markdown`, `git diff --check`, and inspection of
   final status, generated roots, dependency files, and lockfiles.

Use public Make targets for repository checks. Generated topology may change
only through its authored owner inputs and the relevant generator; never edit
generated topology, generated protocol roots, `go.sum`, or `pnpm-lock.yaml`
by hand.

## 20. APR Checkpoint Ledger and Handoff Template

### 20.1 Checkpoint ledger

| Checkpoint | Status | Files changed | Commands and evidence | Compatibility or risk | Next eligible slice |
| --- | --- | --- | --- | --- | --- |
| APR-S00 | COMPLETE | `docs/handoffs/assessments-module-refactor-tracker.md` only | Baseline: `make lint-go` passed; Assessments slice passed 23/23 at `.cartulary/test-results/20260819T204641Z-p3914346`; revalidation: `make lint-markdown` passed at `.cartulary/test-results/20260819T211602Z-p3967606`; `git diff --cached --check`, commit, migration inventory, and tracker-only status checks passed | Planning only; no product, migration, test, catalog, generated, lockfile, or public compatibility effect | APR-S01 is authorized and eligible |
| APR-S01 | COMPLETE | Codec, migration, tests, owner catalog, migration catalog projection, generated manifests/index, schema-hash expectations, harness count, and tracker | Assessments 25/25 and service-backed 17/17; codec and migration rows pass; Workbook and Database Migrations affected rows pass; migration/generation/artifact/JSON drift passes; details in Section 20.3 | Valid replay history is retained in strict v1 storage; public response shape is unchanged; Up precedes the strict binary and Down requires the old binary | APR-S02 is authorized and eligible |
| APR-S02 | COMPLETE | Assessment import facade, application composition, import characterization, and tracker | Assessments passed 25/25 and service-backed 17/17; Imports passed 22/22 and service-backed 14/14; details in Section 20.4 | Internal duplicate orchestration was removed without changing transaction ownership, mutation/revision/intent order, response fields, or rollback | APR-S03 is authorized and eligible |
| APR-S03 | COMPLETE | Assessment merge participant, Entities adapter, and tracker | Assessments, Entities, Revisions, and Timeline focused/service-backed suites pass; details in Section 20.5 | Public merge results and history are unchanged; only unused internal state was removed | APR-S04 is authorized and eligible |
| APR-S04 | COMPLETE | Merge constructors/composition, live row-version boundary, constructor/boundary/export tests, owner-row selectors, generated topology index, and tracker | Assessments, Entities, Timeline, real server bootstrap composition, lint, and module boundaries pass; exact evidence is in Section 20.6 | Internal constructors now return errors; malformed internal row-version values now fail closed; no public or protocol behavior changed | APR-S05 is authorized and eligible |
| APR-S05 | COMPLETE | Final validation evidence, operator migration-evidence digest expectation, tracker traceability, and handoff | All owner, drift, frontend, browser, broad check, and release gates pass; details in Section 20.7 | Public behavior remains unchanged; migration deployment and rollback ordering are explicit | None; APR is complete |

APR-S01 added exactly one service-free codec row and one
`postgres_migration` row. The current count is 19 total and 11 service-backed;
the repository-wide `postgres_migration` count is 9.

### 20.2 Implementation checkpoint template

Copy and complete this block after each authorized implementation slice and
before starting its successor:

| Field | Required entry |
| --- | --- |
| Slice and status | Slice ID, `COMPLETE` or `BLOCKED`, and date |
| Baseline and final commit/worktree | Exact commit and relevant status summary |
| Files | Authored, moved, deleted, and generated files |
| Requirements and gaps | IDs closed, evidence added, and anything still open |
| Commands | Exact commands, results, counts, and result/artifact roots |
| Compatibility and migration | Preserved behavior, intentional breaks, data effects, and rollback posture |
| Generated and dependency state | Generator results, generated roots, dependency files, and lockfiles |
| Failures and skips | Exact failure, causal attribution, resolution, or explicit reason for a skipped non-applicable check |
| Residual risk | Remaining risk carried to the next slice |
| Next slice | Next eligible slice and confirmation of authorization |

### 20.3 APR-S01 checkpoint

| Field | APR-S01 record |
| --- | --- |
| Slice and status | APR-S01 `COMPLETE`, 2026-08-19 |
| Baseline and final worktree | Baseline remains `6cb809190bf107e805720a9552a894783db10cbc`; the pre-existing staged tracker is preserved and APR implementation changes remain unstaged |
| Authored files | Added `db/migrations/00035_assessment_create_idempotency_v1.sql`, `internal/app/workbookassembly/assessment_idempotency.go`, `internal/app/workbookassembly/assessment_idempotency_test.go`, and `internal/app/workbookassembly/assessment_idempotency_migration_test.go`; changed the workbook adapter, two schema-hash expectations, migration owner generator, Assessments owner manifest, PostgreSQL fixture count, and this tracker |
| Generated files | `make generate` updated `tools/migration_history_manifest.json` and `tools/execution_topology_render_index.json`; no generated protocol root changed |
| Requirements and gaps | Closed APR-REQ-003, APR-REQ-004, APR-REQ-009 for persisted decoding, and APR-REQ-012; removed the runtime legacy decoder; retained exact public response behavior |
| Focused commands | Codec row passed at `.cartulary/test-results/20260819T213108Z-p4090346`; migration row passed at `.cartulary/test-results/20260819T213215Z-p4106308`; Assessments passed 25/25 at `.cartulary/test-results/20260819T213342Z-p4125288`; service-backed Assessments passed 17/17 at `.cartulary/test-results/20260819T213452Z-p4166362`; affected Database Migrations rows passed 2/2 at `.cartulary/test-results/20260819T213547Z-p12926`; affected Workbook passed 3/3 at `.cartulary/test-results/20260819T213622Z-p19914`; the Database Migrations head/rollback row passed 3/3 at `.cartulary/test-results/20260819T213700Z-p34790` |
| Drift and generation | `make generate` passed at `.cartulary/test-results/20260819T213309Z-p4121641`; `make migration-drift` passed 5/5 at `.cartulary/test-results/20260819T213551Z-p13384`; `make generate-drift` passed 4/4 at `.cartulary/test-results/20260819T213559Z-p16103`; generated-artifact policy passed 3/3 at `.cartulary/test-results/20260819T213607Z-p18981`; JSON shape passed 3/3 at `.cartulary/test-results/20260819T213608Z-p19387` |
| Compatibility and migration | Storage v1 contains the scope-bound incident identity while replay strips that internal member so the HTTP row remains unchanged. Up preflights all target rows, enriches valid pre-cutover canonical and legacy rows, and leaves unrelated rows unchanged. Down removes the storage-only incident identity while rebuilding the historical three-member envelope. Deploy Up before the strict binary; roll back the binary before Down |
| Generated and dependency state | Intentional migration SHA and schema hash are updated through authored inputs and `make generate`; Assessments routing is 19/11 and repository `postgres_migration` fixtures are 9; no dependency or lockfile changed |
| Failures and resolution | The first codec fixture used ineffective string mutations and was corrected. The first migration fixture omitted a required unrelated hash and read PostgreSQL DETAIL from the wrong wrapper level; both tests were corrected. Full Assessments then exposed that public rows omit `incident_id`; storage-only enrichment and replay stripping resolved it without a public contract change. A Down SQL precedence ambiguity was parenthesized. All affected gates were rerun and passed |
| Residual risk | Operational preflight can still reject unsupported production data by design; such a failure requires explicit data disposition while retaining the pre-cutover binary. No known code or test defect remains in APR-S01 |
| Next slice | APR-S02 is authorized and eligible; its APR-S01 dependency checkpoint is complete |

### 20.4 APR-S02 checkpoint

| Field | APR-S02 record |
| --- | --- |
| Slice and status | APR-S02 `COMPLETE`, 2026-08-19 |
| Baseline and final worktree | Baseline remains `6cb809190bf107e805720a9552a894783db10cbc`; APR-S01 changes are retained and the pre-existing staged tracker remains preserved |
| Authored files | Changed `internal/modules/assessments/import_create.go`, `internal/app/importassembly/assessment_facade.go`, `internal/app/importassembly/assessment_integration_test.go`, and this tracker |
| Requirements and gaps | Closed APR-REQ-005 and the assessment-specific import-finalization gap. IMP-01 through IMP-10 remain passing. The characterization now explicitly requires the record-change intent alongside source, projection, mutation, and revision facts |
| Commands | `make format` passed at `.cartulary/test-results/20260819T214159Z-p82956`; `make lint-go` passed; Assessments passed 25/25 at `.cartulary/test-results/20260819T214217Z-p97968`; service-backed Assessments passed 17/17 at `.cartulary/test-results/20260819T214312Z-p139001`; Imports passed 22/22 at `.cartulary/test-results/20260819T214419Z-p179547`; service-backed Imports passed 14/14 at `.cartulary/test-results/20260819T214531Z-p221059` |
| Compatibility | `ImportCreateDependencies.Revisions` now consumes `ownerfacade.RecordRevisionAndIntentAppender` directly. The owner still borrows the caller transaction, validates before effects, refreshes before finalization, publishes the same create mutation/revision/intent, and returns the same record, row version, mutation reference, result codes, and row refresh |
| Generated and dependency state | No generated artifact, dependency, or lockfile changed in APR-S02; routing remains 19/11 |
| Failures and resolution | The first focused service-backed attempt failed during build because the removed DTO left one unused `uuid` import. The import was removed and all affected gates passed on rerun |
| Retired symbols | Exact searches are empty for `ImportCreateCommand`, `ImportRevision`, `ImportRevisionAppender`, `AppendAssessmentImportRevisionTx`, `assessmentImportRevisionAdapter`, and `assessmentImportValuesByField` within the Assessments and import-assembly scope |
| Residual risk | No known semantic drift or compatibility shim remains. Future import-finalization changes are now owned once by Imports |
| Next slice | APR-S03 is authorized and eligible; its APR-S02 dependency checkpoint is complete |

### 20.5 APR-S03 checkpoint

| Field | APR-S03 record |
| --- | --- |
| Slice and status | APR-S03 `COMPLETE`, 2026-08-19 |
| Baseline and final worktree | Baseline remains `6cb809190bf107e805720a9552a894783db10cbc`; APR-S01 and APR-S02 changes are retained and the staged tracker baseline remains preserved |
| Authored files | Changed `internal/modules/assessments/merge_effects.go`, `internal/modules/entities/merge/ports.go`, and this tracker |
| Requirements and gaps | Closed APR-REQ-006 and the dead merge API/query-state gap. The typed `MergeProtectedSetChangedError`, `RecordID`, Entities reason translation, actual source `updated_at` write, tombstone members, before/after values, snapshots, projection refresh, and history remain |
| Commands | `make format` passed at `.cartulary/test-results/20260819T214810Z-p263078`; protected-set integration passed 3/3 at `.cartulary/test-results/20260819T214814Z-p266652`; Assessments passed 25/25 at `.cartulary/test-results/20260819T214856Z-p282148` and service-backed 17/17 at `.cartulary/test-results/20260819T214956Z-p323149`; Entities passed 32/32 at `.cartulary/test-results/20260819T215056Z-p363692` and service-backed 29/29 at `.cartulary/test-results/20260819T215238Z-p416633`; Revisions passed 27/27 at `.cartulary/test-results/20260819T215422Z-p468616` and service-backed 20/20 at `.cartulary/test-results/20260819T215530Z-p512579`; Timeline passed 48/48 at `.cartulary/test-results/20260819T215641Z-p555384` and service-backed 29/29 at `.cartulary/test-results/20260819T220117Z-p612949` |
| Compatibility | Internal `MergeMutation` version members and the sentinel/unwrap contract were removed atomically with the Entities adapter. Typed error classification and all observable merge effects remain unchanged |
| Query contraction | `created_at` and the read-side `updated_at` were removed from the assessment merge select, scan, record DTO, and normalization. The `UPDATE assessments ... updated_at = $3` write remains the sole timestamp use |
| Generated and dependency state | No generated artifact, owner row, dependency, or lockfile changed in APR-S03; routing remains 19/11 |
| Failures and skips | No APR-S03 command failed and no applicable focused check was skipped |
| Retired symbols | Exact searches are empty in the assessment participant for `ErrMergeProtectedSetChanged`, `Unwrap`, `BeforeVersionID`, `AfterVersionID`, and `created_at`; the only `updated_at` occurrence is the retained source write. The assessment-to-Entities mapping contains no version mapping |
| Residual risk | No known behavior or history risk remains. The constructor is intentionally unchanged here and is owned by APR-S04 |
| Next slice | APR-S04 is authorized and eligible; its APR-S03 dependency checkpoint is complete |

### 20.6 APR-S04 checkpoint

| Field | APR-S04 record |
| --- | --- |
| Slice and status | APR-S04 `COMPLETE`, 2026-08-19 |
| Baseline and final worktree | Baseline remains `6cb809190bf107e805720a9552a894783db10cbc`; APR-S01 through APR-S03 changes are retained, the pre-existing staged tracker baseline is preserved, and APR-S04 changes remain unstaged |
| Authored files | Changed `internal/modules/assessments/merge_effects.go`, `internal/modules/assessments/facade.go`, `internal/modules/assessments/facade_contract_test.go`, `internal/app/assessmentassembly/adapters.go`, `internal/app/timelineassembly/assembly.go`, `internal/modules/entities/merge/merge_protected_set_test.go`, `internal/modules/entities/merge/merge_protected_set_composition_test.go`, `tools/test_families/module.assessments.json`, and this tracker; added `internal/modules/assessments/merge_effects_constructor_test.go` and `internal/modules/assessments/export_surface_test.go`; no file was moved or deleted |
| Generated files | `make generate` updated `tools/execution_topology_render_index.json` from the authored selector expansion. The APR-S01 migration manifest change remains present but did not change again in APR-S04; no generated protocol root changed |
| Requirements and gaps | Closed APR-REQ-007, APR-REQ-008, the live portion of APR-REQ-009, APR-REQ-010, APR-REQ-011, and the remaining APR-REQ-014 construction/surface enforcement. Both merge constructors return errors, Timeline and server composition propagate them, the live facade accepts only positive `int64`, and the exact root/workbookprojection AST allowlists cover declarations, exported values, functions, types, methods, fields, and interface methods |
| Primary commands | `make format` passed at `.cartulary/test-results/20260819T224148Z-p1152513`; `make lint-go` passed; `make generate` passed at `.cartulary/test-results/20260819T221604Z-p740111`; Assessments passed 25/25 at `.cartulary/test-results/20260819T221633Z-p747506` and service-backed 17/17 at `.cartulary/test-results/20260819T221737Z-p789004`; the final constructor/version/export row passed 3/3 at `.cartulary/test-results/20260819T224207Z-p1160940` |
| Downstream commands | Entities rerun passed 32/32 at `.cartulary/test-results/20260819T222435Z-p882868` and service-backed passed 29/29 at `.cartulary/test-results/20260819T222618Z-p935008`; the affected Timeline failure set rerun passed 12/12 at `.cartulary/test-results/20260819T223515Z-p1043832` and Timeline service-backed passed 29/29 at `.cartulary/test-results/20260819T223607Z-p1081682`; real app-server bootstrap composition passed 3/3 at `.cartulary/test-results/20260819T224059Z-p1137548`; `make backend-module-boundary-check` passed 3/3 at `.cartulary/test-results/20260819T224201Z-p1160551` |
| Surface and routing evidence | The positive allowlist passes; the synthetic injected function is reported as unexpected and the inverse fixture proves a removed export is reported as missing. The allowlist retains `Facade`, its typed ports, `workbookprojection`, and the five contribution constructors. `make explain-test-owner OWNER=module.assessments` reports exactly 19 rows and 11 service-backed rows. The existing facade-transaction row owns all new tests, so no catalog row was added |
| Compatibility | `NewMergeEffects` and the assessment-assembly wrapper have an intentional internal signature break and no alias or `Must` helper. Valid Timeline/server startup is unchanged. Invalid dependencies now produce stable contextual startup errors. Malformed live projection adapters now fail before revision or idempotency writes and the transaction rolls back; valid production `int64` behavior is unchanged |
| Generated and dependency state | Routing remains 19/11 and repository `postgres_migration` fixtures remain 9. Generated change is generator-produced from the authored test selector. No dependency, `go.sum`, `pnpm-lock.yaml`, or other lockfile changed |
| Failures and resolution | The first live-version test attempt at `.cartulary/test-results/20260819T220936Z-p685978` miscounted envelope creation as a revision; the counter was moved to the revision appender. The first export run at `.cartulary/test-results/20260819T221306Z-p706001` intentionally used empty allowlists to capture the exact surface, which was then frozen. Entities first reached 31/32 at `.cartulary/test-results/20260819T221840Z-p829577` because an isolated object-store lane timed out; the same full route passed 32/32 on rerun. Timeline first reached 45/48 at `.cartulary/test-results/20260819T222800Z-p987019` because one browser row timed out and one isolated object-store lane failed readiness; every affected row passed in the exact rerun and the full service-backed slice passed. No product assertion failure remains and no applicable check was skipped |
| Retired symbols | Exact owning-scope searches are empty for the retired import DTOs/adapter/helper, merge sentinel/unwrap/version state, constructor `Must` or panic path, and live `int`, `int32`, or `float64` coercion cases. The assessment merge read has no `created_at` or read-side `updated_at`; the actual `updated_at = $3` write remains |
| Residual risk | No known APR-S04 defect remains. The exact export list intentionally makes future public surface additions review-visible; they require an explicit allowlist decision rather than an implicit compatibility commitment |
| Next slice | APR-S05 is authorized and eligible; its APR-S04 dependency checkpoint is complete |

### 20.7 APR-S05 validation and final handoff

| Field | APR-S05 record |
| --- | --- |
| Slice and status | APR-S05 `COMPLETE`, 2026-08-19; the complete APR iteration is `COMPLETE` |
| Baseline and final worktree | Baseline remains `6cb809190bf107e805720a9552a894783db10cbc`. Final status contains exactly the 27 planned paths listed below. The pre-existing staged tracker baseline remains preserved; its APR execution additions and every implementation path remain unstaged. No unrelated path was found |
| Requirements and gaps | APR-REQ-001 through APR-REQ-015 and all six Section 17.2 gaps are closed against the traceability tables below. No product owner, `docs/domain.md`, Core specification, OpenAPI, view schema, bundle, frontend, field-key, or generated protocol changed |
| Routing | `make task-guide ROLE=module-author OWNER=module.assessments` returned the focused and service-backed routes. `make explain-test-owner OWNER=module.assessments` reports exactly 19 total rows and 11 service-backed rows. Repository-wide `postgres_migration` routing is 9 |
| Migration completion | Migration 35 is owned by Assessments and generated in `tools/migration_history_manifest.json` with SHA-256 `f5800e4d9b733a279d93743506d6e12ca86f9b2d8a3637161c4b6072678dfacb`. Codec and migration matrices prove valid legacy conversion, strict canonical decoding, unchanged already-strict rows, unrelated-route isolation, aggregate-only fail-closed preflight, replay equivalence, and deterministic Up/Down/Up |
| Compatibility and deployment | Public create/replay responses, merge results/history, import responses, Timeline/server startup, and every external protocol remain unchanged. Deploy migration Up before the strict binary; the old binary already reads v1. Roll back the binary before Down, and use Down only on disposable rollback evidence. Invalid persisted state blocks Up without mutation; malformed internal projection adapters now fail closed and roll back |
| Finalization | `make agent-finalize` passed 1/1 at `.cartulary/test-results/20260819T231753Z-p1861826`. No successful full warm-check root was available at invocation, so `RESULTS_DIR` was intentionally unset; retained-run canonical-evidence and scheduler maintenance were skipped with `results-dir-not-provided`, while JSON shape, catalog/tier coverage, generated-structure refresh, and retained secret scan passed |
| Generated and dependency state | Only `tools/migration_history_manifest.json` and `tools/execution_topology_render_index.json` are generated changes, both produced by `make generate` from authored inputs. Finalizer reported generated state `unchanged`; generation drift and generated-artifact policy pass. No file under `internal/gen/**`, `packages/protocol-ts/src/generated/**`, or `packages/ui-contracts/src/generated/**` differs. No dependency manifest, `go.sum`, `pnpm-lock.yaml`, or other lockfile differs |
| Failures and skips | The APR-S04 authoring and infrastructure retries are recorded in Section 20.6. In APR-S05, Timeline first reached 47/48 because one isolated object-store lane failed readiness at `.cartulary/test-results/20260819T225447Z-p1530920`; its exact affected store set passed at `.cartulary/test-results/20260819T230144Z-p1587507`, service-backed Timeline passed 29/29, and the exact full focused command then passed 48/48. `test-fast` first reached 408/409 because the operator migration-evidence golden still described the pre-migration-35 manifest at `.cartulary/test-results/20260819T232824Z-p1954234`; the intentional digest was updated, its owner row passed 1/1, and `test-fast` passed 409/409. The only skip is retained-run maintenance described above; no applicable product or release gate was skipped or waived |
| Residual risk | Production Up can intentionally reject unsupported durable replay state. Operators must keep the pre-cutover binary available and disposition such data explicitly before retrying. This fail-closed operational possibility and the documented deployment order are the only residual concerns; no known implementation, compatibility, security, or validation defect remains |
| Final assertion | Every applicable red gate was corrected or rerun to green. No unexplained applicable failure was waived. There is no next implementation slice |

#### 20.7.1 Final owner and broad validation evidence

| Gate | Result and run root |
| --- | --- |
| Assessments | Focused 25/25 at `.cartulary/test-results/20260819T224441Z-p1178064`; service-backed 17/17 at `.cartulary/test-results/20260819T224542Z-p1218732` |
| Imports | Focused 22/22 at `.cartulary/test-results/20260819T224641Z-p1259138`; service-backed 14/14 at `.cartulary/test-results/20260819T224751Z-p1299648` |
| Entities | Focused 32/32 at `.cartulary/test-results/20260819T224906Z-p1340045`; service-backed 29/29 at `.cartulary/test-results/20260819T225052Z-p1392330` |
| Revisions | Focused 27/27 at `.cartulary/test-results/20260819T225237Z-p1444404`; service-backed 20/20 at `.cartulary/test-results/20260819T225342Z-p1488096` |
| Timeline | Final exact focused rerun 48/48 at `.cartulary/test-results/20260819T235049Z-p2281418`; service-backed 29/29 at `.cartulary/test-results/20260819T230224Z-p1602379` |
| Workbook | Focused 67/67 at `.cartulary/test-results/20260819T230659Z-p1658177`; service-backed 39/39 at `.cartulary/test-results/20260819T230930Z-p1717005` |
| Projections | Focused 15/15 at `.cartulary/test-results/20260819T231148Z-p1772772`; service-backed 11/11 at `.cartulary/test-results/20260819T231232Z-p1788576` |
| Database Migrations | Focused 9/9 at `.cartulary/test-results/20260819T231314Z-p1803779`; service-backed 4/4 at `.cartulary/test-results/20260819T231433Z-p1819103` |
| App server composition | Real successful-bootstrap composition passed 3/3 at `.cartulary/test-results/20260819T231554Z-p1834139` |
| Structure and drift | Backend boundaries 3/3 at `.cartulary/test-results/20260819T231639Z-p1849027`; harness contract 2/2 at `.cartulary/test-results/20260819T231646Z-p1849499`; migration drift 5/5 at `.cartulary/test-results/20260819T231701Z-p1850084`; generation drift 4/4 at `.cartulary/test-results/20260819T231714Z-p1852871`; generated-artifact policy 3/3 at `.cartulary/test-results/20260819T231724Z-p1855819`; JSON shape 3/3 at `.cartulary/test-results/20260819T231730Z-p1856277`; lint 11/11 at `.cartulary/test-results/20260819T231737Z-p1856842` |
| Frontend and browser | Typecheck 2/2 at `.cartulary/test-results/20260819T231808Z-p1864675`; unit 390/390 at `.cartulary/test-results/20260819T231821Z-p1865195`; import boundaries 2/2 at `.cartulary/test-results/20260819T232402Z-p1902191`; webserver-backed browser 58/58 at `.cartulary/test-results/20260819T232411Z-p1902679` |
| Broad completion | Operator migration-evidence row 1/1 at `.cartulary/test-results/20260819T232952Z-p1965827`; `test-fast` 409/409 at `.cartulary/test-results/20260819T233000Z-p1966359`; `check` 631/631 at `.cartulary/test-results/20260819T233014Z-p1967365`; `release-check` 788/788 at `.cartulary/test-results/20260819T233504Z-p2074191` |
| Final inspection | `make lint-markdown` passed at `.cartulary/test-results/20260819T235818Z-p2338109`; staged and unstaged `git diff --check`, exact 27-path status, staged-baseline preservation, no-deletion, public/generated-root, dependency/lockfile, 19/11 routing, 9-migration-fixture, migration-35 manifest, and retired-symbol assertions all passed |

#### 20.7.2 Requirement traceability

| Requirement | Final evidence and disposition |
| --- | --- |
| APR-REQ-001 | Sections 1-15 remain historical evidence; only the top posture table distinguishes the completed APR iteration |
| APR-REQ-002 | APR-S00 through APR-S05 executed serially, with Sections 20.3 through 20.7 completed before successor work began |
| APR-REQ-003 | Strict codec and migration/replay rows pass; valid durable history converges on the v1 payload and the runtime legacy decoder is absent |
| APR-REQ-004 | Migration preflight tests prove malformed/unknown state blocks mutation with aggregate-only diagnostics and unrelated rows remain isolated |
| APR-REQ-005 | Assessment imports use the Imports-owned appender, finalizer, and field mapper; IMP-01 through IMP-10 and both owner matrices pass |
| APR-REQ-006 | Dead assessment merge sentinel/version/timestamp read state is absent; Assessments, Entities, Revisions, and Timeline evidence proves behavior/history parity |
| APR-REQ-007 | Both merge constructors return stable errors, Timeline/server composition propagates them, valid composition passes, and there is no `Must` or constructor panic path |
| APR-REQ-008 | The live projection boundary accepts only positive `int64`; all other matrix values fail before revision/idempotency writes and roll back |
| APR-REQ-009 | Serialized-number handling remains in the persisted codec and Imports-owned finalization; no live coercion branch remains |
| APR-REQ-010 | Exact standard-library AST guards pass for the root and `workbookprojection`; injected extra and simulated missing exports both fail comparison |
| APR-REQ-011 | The allowlist retains `Facade`, typed ports, `workbookprojection`, dual-driver test support, and all five contribution constructors |
| APR-REQ-012 | The authored catalog has exactly 19 Assessments rows and 11 service-backed rows; repository `postgres_migration` count is 9 |
| APR-REQ-013 | No public contract file changed; frontend/browser, `check`, and `release-check` all pass |
| APR-REQ-014 | Owning-scope searches are empty for every retired import, merge, constructor, and coercion symbol; no alias or compatibility layer was added |
| APR-REQ-015 | This checkpoint maps all requirements, gaps, files, effects, failures, skips, and validation evidence; no applicable failure is unexplained or waived |

#### 20.7.3 Gap closure

| Gap | Final remediation and validation |
| --- | --- |
| Indefinite legacy idempotency decoding | Migration 35 performs fail-closed one-time convergence and deterministic Down; one strict runtime codec remains. Codec, migration, replay, drift, operator evidence, and release gates pass |
| Assessment-specific import finalization | Local DTO/adapter/helper orchestration was deleted and the Imports-owned finalizer is used directly. Assessment and Imports matrices plus explicit intent characterization pass |
| Dead merge API and query state | The sentinel/unwrap, unused assessment mutation version fields, and unused timestamp reads were removed atomically with the Entities adapter. Merge, history, snapshot, refresh, rollback, and downstream matrices pass |
| Panic-based merge construction | Constructors return errors through Timeline/server composition; stable nil errors and valid composition pass with no `Must` helper |
| Permissive live row-version coercion | Only positive `int64` succeeds; every other tested type/value fails before finalization and transaction effects roll back |
| Unlocked exported package surface | Exact AST allowlists protect root and `workbookprojection` declarations and exports; positive, extra-export, and missing-export evidence passes without adding a catalog row |

#### 20.7.4 Final 27-path inventory

Added authored files:

- `db/migrations/00035_assessment_create_idempotency_v1.sql`
- `internal/app/workbookassembly/assessment_idempotency.go`
- `internal/app/workbookassembly/assessment_idempotency_test.go`
- `internal/app/workbookassembly/assessment_idempotency_migration_test.go`
- `internal/modules/assessments/export_surface_test.go`
- `internal/modules/assessments/merge_effects_constructor_test.go`

Modified authored files:

- `docs/handoffs/assessments-module-refactor-tracker.md`
- `internal/app/assessmentassembly/adapters.go`
- `internal/app/importassembly/assessment_facade.go`
- `internal/app/importassembly/assessment_integration_test.go`
- `internal/app/operator/operator_migration_evidence_test.go`
- `internal/app/timelineassembly/assembly.go`
- `internal/app/workbookassembly/assessment_facade.go`
- `internal/modules/assessments/facade.go`
- `internal/modules/assessments/facade_contract_test.go`
- `internal/modules/assessments/import_create.go`
- `internal/modules/assessments/merge_effects.go`
- `internal/modules/database_migrations/catalog_characterization_test.go`
- `internal/modules/entities/merge/merge_protected_set_composition_test.go`
- `internal/modules/entities/merge/merge_protected_set_test.go`
- `internal/modules/entities/merge/ports.go`
- `internal/testutil/pgtest/pgtest_test.go`
- `tools/database-migrations/generate-catalog-projections.mjs`
- `tools/harness/tests/contract-suite-support.mjs`
- `tools/test_families/module.assessments.json`

Generated files:

- `tools/execution_topology_render_index.json`
- `tools/migration_history_manifest.json`

No file was moved or deleted.

## 21. APR Binary Completion and Deferred Scope

APR-S00 is complete only when this tracker is the sole changed file and both
Markdown lint and diff whitespace validation pass. That completion establishes
a decision-ready plan; it does not start implementation.

APR-S01 through APR-S05 were separately authorized and executed in order,
every checkpoint is complete, the 19/11 routing expectation is proven, every
APR requirement and gap is closed, and the final validation ladder passes
without unexplained failure. The APR iteration is complete.

The following remain deliberately outside APR:

- creation of an Assessment NLSpec or changes to Core owners or
  `docs/domain.md`;
- public route, HTTP, WebSocket, OpenAPI, view-schema, bundle-version,
  frontend, public protocol, field-key, or generated-protocol changes;
- historical assessment rewrites beyond the narrowly planned idempotency
  payload migration;
- compatibility aliases, deprecated copies, or a parallel legacy decoder;
- removal of retained contribution facades, typed ports,
  `workbookprojection`, or dual-driver test support; and
- implementation based solely on the documentation-only APR-S00 checkpoint;
  APR-S01 through APR-S05 instead used the separate 2026-08-19 implementation
  authorization recorded above.
