# tasksdecisions Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Item | Current planning posture |
| --- | --- |
| Target path | `internal/modules/tasksdecisions` |
| Target label | `tasksdecisions`, derived from the target path and already valid lowercase kebab case |
| Output path | `docs/handoffs/tasksdecisions-module-refactor-tracker.md` |
| Status | Complete: CG-01 passed; SL-00 through SL-07 are done; TD-012 and the overall handoff are closed |
| Allowed change in this session | SL-00, SL-04, SL-05, SL-06, CG-01, SL-01, SL-02, SL-03, and SL-07 in that order, including their named owner, implementation, test, harness, generated, boundary, validation, and tracker surfaces |
| Non-goals | No unrelated behavior change; no route, OpenAPI, view-schema identity, database-schema, frontend, selector, or valid portability-format change; no hand-edited generated artifact or lockfile |
| Current authority | The 2026-08-01 10:14 EDT user work order explicitly authorizes both the conformance-correction and structural-refactor work orders through final validation and handoff |
| Decision closure | RB-001 through RB-004 remain decided implementation matters; RB-005 is resolved by the complete authorized work order |

The planning framework is doctrine and a tracker template, not evidence of the
current repository. Live source, contracts, tests, and harness projections were
inspected before the findings below were recorded. Observable behavior is frozen
by default. Test rows and harness phase maps are evidence accounting only and do
not define the runtime architecture.

Source hierarchy used for this tracker:

1. Adopted subsystem NLSpecs within their named scopes.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only if timed or fixture-sensitive claim publication becomes
   applicable; it is not applicable to this implementation work order.
4. Domain vocabulary and implementation-support guidance.
5. Current repository code and tests as implementation-state evidence.
6. This framework and prior handoffs as evidence only.

Owner documents inspected:

- `docs/spec/00_document_set_status_and_precedence.md`;
- `docs/spec/01_architecture_storage_and_view_contracts.md`, especially the
  modular-monolith, workbook mutation, view-schema, projection-provider,
  import, incident-portability, and source-owner sections;
- `docs/spec/02_domain_model_schema_and_history.md`, especially task-request,
  decision, link, lifecycle, history, and rollback ownership;
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`, especially
  system views, conflicts, save state, collaboration, pending work, and
  coordination surfaces;
- `docs/spec/04_security_deployment_and_conformance.md`, especially incident
  roles, record authorization, route-time authorization, and task/decision
  conformance criteria;
- `docs/reporting-subsystem-nlspec.md` for the adopted Reporting Subsystem scope;
- `docs/testing-harness-nlspec.md` for Make-owned command and evidence mechanics;
- `docs/domain.md` for vocabulary and boundary interpretation; and
- `docs/handoffs/cartulary_modular_refactor_planning_framework.md` as the
  planning template and doctrine.

Writing doctrine and decision evidence inspected:

- `docs/research/nlspec-spec.md` defines the precision, completeness, interface,
  default, mapping, and acceptance-criteria standard used by this revision. It
  does not own Cartulary product behavior.
- `temp/analysis-notes.md` supplies the reviewed closure recommendations adopted
  as tracker decisions below. It remains evidence and does not supersede an
  adopted owner.

Repository evidence inspected or searched includes every target file listed in
Section 2 and the following caller, contract, generated-surface, frontend, and
harness families:

- `internal/app/{importassembly,incidentportabilityassembly,projectionassembly,recoveryassembly,revisionassembly,workbookassembly}`;
- `internal/modules/{collaboration,imports,incidentbundles,projections,reporting,revisions,workbook}`;
- `internal/modules/workbook/{routes.go,store.go,mutation_store.go,mutation_contributions.go}`
  and relevant integration, conflict, row-wire, and boundary tests;
- `contracts/view-schemas/cartulary.view.{task_requests,decisions}.v1.json`;
- `contracts/openapi-source/owners/module.workbook/openapi.json`;
- `contracts/{imports,incident-bundles,projection-providers}/index.json` where
  present, including `contracts/incident-bundles/source_catalog.json`;
- `contracts/verification/owners/module.tasksdecisions.json`;
- `tools/test_families/module.tasksdecisions.json`,
  `tools/backend_module_boundaries.json`, and
  `tools/schema_object_ownership_manifest.json`;
- `apps/web/src/workbook/features/coordination/useCoordinationWorkflowController.ts`
  and its test;
- `apps/web/src/workbook/mutations/createWorkbookMutationCommandPorts.ts` and
  its test;
- `apps/web/e2e/{sentinel,coordination-public-route,workbook.generic}.spec.ts`;
- `packages/view-contracts/src/index.ts` and the stable selector helpers under
  `packages/ui-contracts/src`; and
- references in generated roots under `internal/gen/**` and
  `packages/protocol-ts/src/generated/**`, inspected only to identify downstream
  risk; these files are not proposed hand-edit targets.

No owner-document contradiction was found. The live repository differences
identified below are implementation or evidence findings; they do not grant
this tracker authority to invent replacement behavior.

### Normative language and decision posture

Within this tracker, **MUST** and **MUST NOT** bind a later implementation only
after the matching work order is explicitly authorized. **SHOULD** identifies a
required planning default that may be changed only by recording contrary owner
evidence. **MAY** identifies intentional implementation freedom whose alternatives
do not alter observable behavior or interoperability. Present-tense statements
about the repository describe the inspected revision; they are evidence, not
product requirements.

This tracker owns refactor sequencing, the target-internal package boundary, and
handoff completeness. Adopted owner documents continue to own product behavior.
The exclusive portability attribution described below is therefore a required
Core 01 clarification before its implementation can claim conformance. Closing a
planning decision does not claim that its code, tests, owner text, or evidence
projection already exists.

## 2. Pre-remediation Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/tasksdecisions/create_validation.go` | Minimum-create and direct-patch field validation for tasks and decisions | `ValidateTaskCreateParams`, `ValidateDecisionCreateParams`, `ValidateTaskDirectPatchChange`, `ValidateDecisionDirectPatchChange` | Workbook facade and import create path | Target-local field values and enum helpers | Target store tests; workbook mutation tests indirectly | Task and decision view-schema field contracts | `tasksdecisions` source policy | high | Direct superseded decision create is rejected here; lifecycle-dependent validation continues in the store. |
| `internal/modules/tasksdecisions/deleterestore/provider.go` | Fixed task/decision source snapshots and no-op source tombstone hooks for generic delete/restore coordination | `TaskRequestSource`, `DecisionSource`, constructors and `revisions.DeleteRestoreSource` methods | `revision_provider_contribution.go`; Revisions catalog indirectly | `pgx`; revisions provider contract | `internal/modules/revisions/delete_restore_test.go` | Revision provider catalog and view-schema IDs | `tasksdecisions` source-owned Revisions adapter | medium | Machine boundary policy permits these providers only through the root owner contribution. |
| `internal/modules/tasksdecisions/import_create.go` | Import owner-create facade, source insert, direct-reference validation, link synchronization, and revision finalization | `ImportCreateCommand`, `NewImportCreateFacade`; exported `Store.CreateImportRowTx` | `internal/app/importassembly/owner_registry.go` through the generated target registry | Imports owner facade, records, links, revisions, target store | Import registry/characterization tests; no target-specific apply test found | `contracts/imports/index.json`; generated import-target registry and protocol bindings | `tasksdecisions` import adapter with generic orchestration owned by `imports` | high | Active imported user IDs are checked, but current same-incident membership is not. |
| `internal/modules/tasksdecisions/import_projection.go` | Refresh and load the owner row after import creation | Exported `Store.RefreshImportRowTx` | `import_create.go` | Projections coordinator | Import apply indirectly | Projection-provider descriptors and generated row contracts | `tasksdecisions` adapter over `projections` | medium | Projection storage remains owned by `projections`. |
| `internal/modules/tasksdecisions/incident_bundle_portability.go` | Fixed NDJSON export and import for task and decision source tables | `ExportIncidentBundleFiles`, `ImportIncidentBundleFilesTx` | Source-port adapter | Generic incident-portability codecs and `pgx` | Generic incident-bundle path/catalog tests only | `contracts/incident-bundles/source_catalog.json` | `tasksdecisions` source portability adapter | high | Direct source SQL is legitimate here, but semantic validation occurs after apply through the source port. |
| `internal/modules/tasksdecisions/incident_bundle_source_port.go` | Declares the `tasks_decisions` source family, paths, dependencies, and four invariants; adapts prepare/apply/validate | `NewIncidentBundleSourcePort` | `internal/app/incidentportabilityassembly/catalog.go` | Incident-bundle source-port contract and portability functions | Catalog-shape tests; no task/decision invariant behavior test found | Incident-bundle source catalog | `tasksdecisions` source-owned port | high | Runtime validation checks only envelope type/scope despite declaring four required invariants. |
| `internal/modules/tasksdecisions/incident_bundle_subtype_presence.go` | Contributes task-request and decision subtype presence for Records validation | `IncidentBundleSubtypeContribution`; unexported source methods implement the port | Incident-portability assembly and Records subtype catalog | Records subtype-presence contract | Generic incident-bundle catalog/import tests indirectly | Records subtype mapping in the portability contract | `tasksdecisions` source-owned contribution | medium | Correctly contributes owner facts to a generic Records coordinator. |
| `internal/modules/tasksdecisions/mutations.go` | Authoritative task/decision source SQL, defaults, lifecycle validation, decision-machine consistency, and supersedes-link source effects | `Store`, create parameter/value/error/state types, constructors, enum validators, lifecycle validators, exported transactional methods | Workbook/import/supersede facades and target tests | `pgx`, links, revisions, platform `authn` unique-violation helper | Five target store tests; workbook route tests indirectly | Task/decision schemas, lifecycle contracts, source-table ownership manifest | `tasksdecisions` authoritative source state | high | Direct SQL to `task_requests` and `decisions` is intentional; the platform error-classifier dependency should be narrowed. |
| `internal/modules/tasksdecisions/projectionprovider/provider.go` | Deterministic refresh and incident rebuild SQL for task and decision projections | `RefreshTaskRequestTx`, `RefreshDecisionTx`, `RebuildIncidentTaskRequestsTx`, `RebuildIncidentDecisionsTx` | Projections adapter and projection assembly | `pgx`, records and links via SQL, projection storage tables | Target store tests; projection query/rebuild tests | `contracts/projection-providers/index.json`; generated view-schema artifacts | Source intent: `tasksdecisions`; storage lifecycle: `projections` | high | Provider intent belongs with the source owner; projection table lifecycle belongs with `projections`. |
| `internal/modules/tasksdecisions/projectionprovider/query_surfaces.go` | Code-backed query-field, collection, sort, filter, and grouping descriptors | `TaskRequestQuerySurfaces`, `DecisionQuerySurfaces` | Projection assembly and query-surface tests | Projection provider contract | `internal/modules/projections/query_surfaces_test.go` | Task and decision view-schema query contracts | `tasksdecisions` source-owned projection descriptor | high | Field keys and SQL expressions are public query-compatibility risks. |
| `internal/modules/tasksdecisions/recovery_state.go` | Declares task and decision source tables as authoritative recovery state | `RecoveryStateContribution` | `internal/app/recoveryassembly/state_catalog.go` | Platform recovery-state catalog | Recovery catalog and coverage tests indirectly | Schema object ownership and recovery state catalogs | `tasksdecisions` recovery contribution | low | Generic recovery owns catalog validation and execution. |
| `internal/modules/tasksdecisions/reportingprovider/provider.go` | Collects task/decision working-material facts during snapshot materialization | `CollectFieldsTx`, `CollectFactsTx` | `internal/modules/reporting/export_materializer.go` | Reporting export-provider contract; projection and records tables | Reporting boundary guard and reporting integration tests | Reporting export-model inputs | `tasksdecisions` reporting contribution | medium | Projection reads occur during materialization before immutable reporting export-model closure; no Reporting-owner contradiction was found. |
| `internal/modules/tasksdecisions/revision_append_port.go` | Narrow local interface and adapter for change sets, mutations, record revisions, and durable collaboration intents | Unexported `revisionAppendPort` and adapter | Supersede facade | Revisions appender contract | Supersede target tests indirectly | Revision and collaboration intent contracts | `tasksdecisions` facade seam backed by `revisions` | medium | The workbook facade still uses the concrete appender directly and should converge on a narrow seam. |
| `internal/modules/tasksdecisions/revision_provider_contribution.go` | Aggregates task and decision delete/restore and rollback providers for Revisions | `RevisionProviderContribution` | `internal/app/revisionassembly/revisions.go` | Root-owned delete/restore and rollback subpackages; Revisions contracts | Revision assembly and delete/restore tests | Revisions provider catalog and record/view routing | `tasksdecisions` source-owner contribution | medium | Correct source-owner construction pattern; Revisions validates the complete catalog. |
| `internal/modules/tasksdecisions/rollbackprovider/provider.go` | Validates and restores authoritative task/decision scalar source state during rollback | `TaskRequestProvider`, `DecisionProvider`, constructors, validation and restore methods | Root revision contribution | Revisions rollback contract, `pgx` | Local provider tests; generic Revisions integration tests | Rollback source vocabulary and task/decision lifecycle | `tasksdecisions` source-owned rollback adapter | high | Repeats enum/lifecycle knowledge and intentionally excludes link-backed collections. |
| `internal/modules/tasksdecisions/rollbackprovider/provider_test.go` | Unit evidence for null preservation, collection exclusion, and invalid owner values | Two test functions only | Go test runner when the package is selected | Standard test and local provider | Self | None | `tasksdecisions` test evidence | low | No owner-family row currently names this package or its two tests. |
| `internal/modules/tasksdecisions/supersede_facade.go` | Atomic explicit decision supersession, replay, lifecycle guards, links, two record versions, projections, revisions, and response | `DecisionsViewSchemaID`, `SupersedeFacade`, request/command/result/error types, constructor and `SupersedeDecision` | Workbook store and mutation adapter; app workbook assembly | Platform `authn` and Postgres, records, links, projections, revisions | Target store tests; frontend controller and browser scenario indirectly | Workbook OpenAPI supersede operation, decision view schema, collaboration contract | Decision semantics: `tasksdecisions`; generic route orchestration: `workbook`/`revisions` | high | Executed targets remain executed while becoming projection-superseded; this is a high-risk aggregate command. |
| `internal/modules/tasksdecisions/task_decisions_store_test.go` | Service-backed characterization of create defaults, query, lifecycle, links, supersession, replay, and inconsistent-machine rejection | Five owner-mapped test functions plus local helpers | `module.tasksdecisions.store` test-family row | App test support, incidents, workbook facade, Postgres | Self | Verification owner/test-family projections | `tasksdecisions` test evidence | medium | Despite `_Unit` names, the harness classifies this as Postgres-backed integration evidence. It does not exercise incident-bundle import/export. |
| `internal/modules/tasksdecisions/workbook_conflict.go` | Source-owner same-field conflict resolution and current-window revalidation | `WorkbookConflictCommand`; exported `WorkbookFacade.ResolveConflict` | Workbook conflict provider adapter | Revisions conflict-resolution mechanics plus the workbook facade | Workbook conflict-route and source-owner contribution tests | Public conflict route and view-field conflict classes | Source revalidation: `tasksdecisions`; generic mechanics: `revisions`/`workbook` | high | `keep_saved` correctly uses generic mechanics; mutating choices return through ordinary owner patch semantics. |
| `internal/modules/tasksdecisions/workbook_facade.go` | Task/decision create and patch transaction coordinator, replay, membership and reference checks, source and link mutation, row-version conflict detection, projections, revisions, and transport-oriented result shaping | View IDs; `WorkbookFacade`; request, change, command, result, and error types; constructor; create/patch methods; collection policy helpers | App workbook assembly and workbook mutation contribution adapters | Platform `authn` and Postgres, incidents, records, links, projections, revisions and conflict packages | Target store tests; workbook route/conflict/row-wire tests; browser tests | Workbook OpenAPI operations, view schemas, generated protocol/view contracts | Mixed: `tasksdecisions` source facade plus generic `workbook`/`revisions` coordination | high | Largest mixed-responsibility file; contains HTTP status values and constructs concrete cross-owner stores. |

The 20 files present before implementation are accounted for above. No initial
file was marked out of scope. Implementation added the following nine files;
they are part of the completed target boundary and handoff:

| Added path | Final responsibility | Root or evidence posture |
| --- | --- | --- |
| `internal/modules/tasksdecisions/incident_bundle_contribution.go` | Aggregates source-port and subtype-presence behavior | `NewIncidentBundleContribution` is the only application entry point. |
| `internal/modules/tasksdecisions/incident_bundle_source_port_test.go` | Eight exact portability and atomicity tests | Truthfully routed to `module.tasksdecisions`. |
| `internal/modules/tasksdecisions/internal/policy/policy.go` | Closed enums, defaults, lifecycle, decision-machine, reference, and portability policy | Go-private and boundary-enforced. |
| `internal/modules/tasksdecisions/internal/source/direct_updates.go` | Fixed per-field source updates and PostgreSQL error classification | Go-private and boundary-enforced. |
| `internal/modules/tasksdecisions/internal/source/facts.go` | Transaction-bound lifecycle, machine, envelope, and portability fact loads | Go-private with exact source-table access. |
| `internal/modules/tasksdecisions/internal/source/references.go` | Transaction-bound member and record-reference validation | Go-private with exact Records access. |
| `internal/modules/tasksdecisions/mutation_capabilities.go` | Seven consumer-owned capability groups and transport-neutral facts | Root facade contract used by application composition. |
| `internal/modules/tasksdecisions/projection_contribution.go` | Task/decision projection source plus query contracts | `NewProjectionContribution` hides the projection leaf. |
| `internal/modules/tasksdecisions/reporting_contribution.go` | Task/decision Reporting fact provider | `NewReportingContribution` hides the reporting leaf. |

## 3. Module Boundary Diagnosis

The target is a legitimate tasks/decisions source-owner module because Core 02
defines task requests and decisions as first-class incident records, the schema
ownership manifest assigns their source tables to `tasksdecisions`, and the
projection/import/portability catalogs name the same source owner. That evidence
does not validate the current root-package shape.

The current target is simultaneously:

- a legitimate source-state and lifecycle owner;
- a mutation coordinator;
- a view/projection orchestration participant;
- a persistence-adjacent adapter;
- transport-adjacent through HTTP status shaping and route idempotency;
- a source-owner provider factory for Revisions, Imports, Incident Bundles,
  Recovery, Projections, and Reporting; and
- a mixed-responsibility package with a broad root surface.

It is not a frontend shell/controller, direct grid-vendor adapter, saved-view
owner, WebSocket hub, generalized approval engine, Timeline owner, or accidental
home for entity, indicator, evidence, or artifact semantics.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Task/decision source rows, defaults, enums, lifecycle, and direct field writes | `mutations.go`, `create_validation.go` | `tasksdecisions` | keep | Core 01/02; schema ownership manifest; source SQL | Consolidate policy so every entry path uses one owner implementation. |
| Decision supersession semantics and guards | `supersede_facade.go`, `mutations.go` | `tasksdecisions` | keep | Core 01 decision supersede contract; Core 02 decision machine | Keep semantic atomicity, including executed-target behavior. |
| Generic route replay/idempotency and transport status | Workbook and supersede facades | `workbook` coordinator with platform adapter supplied by app assembly | split | Public routes are owned by module.workbook; facades construct `authn.Store` and return HTTP status | Preserve all replay scopes, precedence, and envelopes. |
| Record-envelope persistence | Workbook/import/supersede facades through `records.Store` | `records` behind a narrow port | split | Core 00/02 owner matrix; existing Records store | Do not move record-envelope SQL into tasksdecisions. |
| Link-backed collections and supersedes edge persistence | Facades and mutation store through `links.Store` or SQL helper | `links` for generic relation persistence; `tasksdecisions` for allowed semantics | split | Core 02 link ownership and task/decision field contracts | Preserve atomicity with source writes and row versions. |
| Projection derivation intent | `projectionprovider/*` | `tasksdecisions` | keep | Core 01 source-owner provider model; projection-provider catalog | Keep descriptors and source joins owner-authored. |
| Projection storage lifecycle and generic query/rebuild coordination | Facades and `internal/modules/projections` | `projections` | split | Core 01 and provider catalog name `projections` as storage owner | Retain narrow refresh/load/rebuild ports. |
| Same-field source revalidation | `workbook_conflict.go` | `tasksdecisions` | keep | Core 03 explicitly retains source-field and collection semantics with the source owner | Generic token and merge mechanics remain outside. |
| Conflict tokens, revision windows, and keep-saved mechanics | Workbook facade/conflict code using Revisions packages | `revisions` behind narrow capabilities | split | Core 03 conflict ownership clarification | Preserve token opacity, route binding, and stale-window behavior. |
| Change-set, mutation, record-revision, and collaboration-intent append | Workbook/import/supersede facades | `revisions` with Collaboration intent appender | split | Revisions appender emits durable `record_changed` intent in the transaction | No source owner may publish directly after commit. |
| Import owner create semantics | `import_create.go`, `import_projection.go` | `tasksdecisions` facade under generic `imports` orchestration | keep | Core 01 import target and owner-create contract | Correct the membership validation gap only in an authorized behavior slice. |
| Incident-bundle source family | `incident_bundle_*` | `tasksdecisions` port under generic Incident Bundles coordination | keep | Core 01 closed source catalog | Implement all declared invariants in a separately authorized slice. |
| Delete/restore/rollback source adapters | `deleterestore/*`, `rollbackprovider/*`, root contribution | `tasksdecisions` providers under generic `revisions` coordination | keep | Core 01/02 history ownership; Revisions catalog | Share owner policy without making Revisions infer source fields. |
| Reporting working-material contribution | `reportingprovider/provider.go` | `tasksdecisions` provider under Reporting materialization | keep | Adopted Reporting NLSpec and reporting boundary tests | Mutable projection reads are bounded to snapshot materialization. |
| Recovery state classification | `recovery_state.go` | `tasksdecisions` contribution under Recovery catalog | keep | Recovery state catalog and schema ownership | No recovery operation belongs in the target. |
| Frontend coordination workflow | `apps/web/src/workbook/features/coordination` | Web workbook feature/controller | defer | Core 03 workbook-surface rule; frontend owner test row | No frontend move is supported by target evidence; freeze contracts only. |
| Grid selectors and vendor integration | `packages/ui-contracts`, `packages/grid-adapter`, generic web grid | UI contracts and grid adapter | defer | No frontend or vendor imports exist in the target | Do not create backend-to-grid coupling. |
| Exact extracted subpackage topology | Not present | Root facade, private policy/source packages, and source-owned leaf providers | split | TD-DEC-004 adopts the graph below; current external leaf imports are migration targets | Implementation MUST prove the adopted graph with compilation and the backend boundary manifest. |

### Closure decision register

The following decisions close the technical planning questions. They do not
authorize or claim implementation.

| Decision ID | Normative decision | Product owner or planning authority | Implementation slice | Required acceptance evidence | Required validation |
| --- | --- | --- | --- | --- | --- |
| TD-DEC-001 | Tasks/Decisions portability rule attribution MUST be exclusive and failure selection MUST follow the four declared invariant IDs in descriptor order. | Core 01 REQ-01-640 and REQ-01-642; exact attribution requires the planned Core clarification | SL-05 | Eight exact portability tests, including coordinator no-partial-state and safe-diagnostics proof | Owner service-backed slice; harness and drift gates |
| TD-DEC-002 | Imported task and decision owner references MUST obey `incident_member_user_ref_v1` inside the unit transaction. | Core 01 REQ-01-628 and REQ-01-619; Core 04 AC-465 through AC-467A | SL-04 | Input/default/race/replay matrix in Section 4 | Owner service-backed slice |
| TD-DEC-003 | Portability and rollback evidence MUST use truthful, exact, non-overlapping owner rows. | Adopted Testing Harness NLSpec and authored verification/test-family inputs | SL-06 | Owner explanation selects each exact test once and store evidence no longer claims portability | Harness contract, generation, and drift gates |
| TD-DEC-004 | The permanent Go boundary MUST use one root facade, private policy/source packages, seven narrow injected capabilities, and source-owned leaf providers with no prohibited reverse edge. | This tracker under the modular-refactor framework; product behavior remains Core-owned | SL-01 through SL-03 | Compilation, caller migration, and boundary-manifest proof | Backend boundary check and owner slices |
| TD-DEC-005 | Conformance correction MUST reach a green evidence baseline before structural movement begins. The complete authorized work order retains separate conformance and structural gates. | Refactor doctrine and the resolved RB-005 authorization boundary | SL-00, SL-04 through SL-06, then SL-01 through SL-03 | Both work-order gates in Section 7 are binary and independently reviewable | Per-gate validation ladders |

### Adopted target package topology

The later structural work MUST converge on this topology. Exact filenames within
each declared package MAY vary when the package responsibility and all dependency
rules remain unchanged.

| Layer | Permanent location | Required responsibility | Explicit exclusion | External import posture |
| --- | --- | --- | --- | --- |
| Root facade | `internal/modules/tasksdecisions` | Public commands, results, semantic errors, transaction orchestration, supersession semantics, conflict revalidation, compatibility wrappers during migration, and source-owner contribution constructors | Route decoding, HTTP status selection, concrete peer-store construction, generic coordinator behavior, and generated DTOs as domain types | The only target package that code outside the target subtree may import |
| Private policy | `internal/modules/tasksdecisions/internal/policy` | Task lifecycle and dependent-field rules, decision lifecycle and supersession rules, reference rules, and portability invariant attribution | SQL, `pgx`, transport, application assembly, generic coordinators, frontend, and grid code | Importable only from the target subtree |
| Private source | `internal/modules/tasksdecisions/internal/source` | Fixed source SQL, transaction-bound lookup/mutation, deterministic invariant-fact queries, and source persistence error classification | HTTP, collaboration publication, dynamic descriptor-derived SQL, and transaction ownership | Importable only from the target subtree; MAY import private policy |
| Source-owned leaves | `portability`, `deleterestore`, `rollbackprovider`, `projectionprovider`, and `reportingprovider` | Construct exact source-owner providers against the generic coordinator contracts | Root-facade import, application composition, and generic coordinator ownership | Constructed behind root contribution functions; current external leaf imports are migration targets |
| Application composition | Matching `internal/app/*assembly` packages | Supply concrete adapters and register root-provided contributions with generic coordinators | Product semantics and direct target-leaf imports | MUST import the root facade only |

The permitted dependency direction is:

```text
internal/app/*assembly -> tasksdecisions root
tasksdecisions root -> tasksdecisions/internal/policy
tasksdecisions root -> tasksdecisions/internal/source
tasksdecisions root -> source-owned leaf providers
tasksdecisions/internal/source -> tasksdecisions/internal/policy
source-owned leaf providers -> private policy/source when required
source-owned leaf providers -> narrow generic coordinator contracts
```

Every reverse edge is forbidden. In particular:

- no package outside the target subtree may import `tasksdecisions/internal/*`
  or a target leaf provider;
- no leaf provider may import the root package;
- private policy MUST NOT import Postgres, `pgx`, app assembly, Workbook,
  Imports, Revisions, Projections, Reporting, Incident Bundles, HTTP, frontend,
  or grid packages;
- generic coordinator modules MUST receive target contributions through
  application composition and MUST NOT discover or construct target internals;
- Workbook and Imports MUST NOT import a target projection, rollback, or
  portability implementation; and
- target packages MUST NOT import frontend selectors, browser components,
  React Data Grid, Handsontable, or grid-vendor DTOs.

Existing imports of `tasksdecisions/projectionprovider` from Projection assembly,
the Projections coordinator, and Projections tests and the import of
`tasksdecisions/reportingprovider` from Reporting are migration targets under
SL-03. Existing imports of `deleterestore` and `rollbackprovider` from the target
root are already in the permitted direction.

### Root facade and injected capability contract

The root facade MUST expose contribution constructors equivalent in behavior to
`NewWorkbookContribution`, `NewImportContribution`, `NewSupersedeFacade`,
`NewIncidentBundleContribution`, `NewProjectionContribution`,
`NewRevisionContribution`, `NewReportingContribution`, and
`NewRecoveryContribution`. Names MAY follow existing repository conventions;
each constructor MUST return the generic interface owned by its coordinator and
MUST hide the leaf implementation from the caller.

Interfaces MUST be declared by the consuming `tasksdecisions` package and remain
operation-specific. The incident-state capability omitted by the planning-only
revision is included explicitly below, bringing the permanent total to seven:

| Capability | Required input | Required output | Failure/default behavior | Explicit exclusion |
| --- | --- | --- | --- | --- |
| Incident state | Existing transaction and exact `incident_id` | Success or typed closed-incident/not-found failure | Revalidate the authoritative lifecycle in the mutation transaction; no cached or route-only fallback | Membership authorization and source mutation |
| Member reference | Existing transaction, target `incident_id`, exact `user_id`, and stable field key | Success or typed owner-validation failure | No trimming, label/email resolution, fuzzy match, auto-create, or distinction among missing/inactive/nonmember identities beyond the safe owner schema | Actor authorization and membership administration |
| Idempotency | Existing transaction, operation scope, key, normalized request identity, and typed committed-result codec | Exact committed replay, changed-content conflict, or newly stored successful result | No implicit retry or scope fallback; rejection and replay MUST create no owner effects | Source mutation and response-shape invention |
| Record envelope | Existing transaction plus exact incident, record type, actor, and expected/current version inputs | Inserted or locked envelope identity and authoritative row version | Affected-row mismatch and stale version are typed failures; no silent create-or-update fallback | Revision append and projection refresh |
| Links | Existing transaction plus exact typed link or field-derived collection command | Deterministic mutation result or typed validation failure | Empty collection means the declared exact empty action, never omission inference; invalid endpoint/type/scope fails atomically | Task/decision lifecycle inference |
| Projection | Existing transaction, record identity, and exact task/decision view identity | Refreshed authoritative `view_row_v1` or typed failure | No stale-row fallback and no ownership of source state | Source mutation and transaction control |
| Revision | Existing transaction plus change-set metadata, mutation entries, record revisions, and durable intent inputs | Exact committed mutation/revision references or typed failure | Append counts and sequence are caller-declared; replay and rejection append nothing | Current-envelope ownership, commit, and post-commit publication |

Every transaction-bound capability MUST reuse the supplied transaction. It MUST
NOT begin, commit, roll back, or nest a transaction; publish an event; perform a
network or object-store call; accept callbacks or untyped maps; use reflection
dispatch; or participate in a generic service locator. No capability has an
implicit retry default.

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` for task and decision views | Core 01; module.workbook transport; source projection descriptors | OpenAPI owner file, view schemas, projection catalog and query surfaces | Target store test; workbook projection query tests; browser scenarios | Preserve exact sort/filter/group/null and collection behavior for every declared task/decision query field | high | No query route shape or cursor behavior may change. |
| `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` | Core 01/02/03/04; module.workbook transport; `tasksdecisions` source create | OpenAPI, view schemas, workbook route and owner adapter | Target store test; workbook mutation integration; public-route and generic browser tests | Active nonmember user refs, replay/no-partial-write, and authorization-precedence cases | high | First success is `201`; exact replay is `200` with original result. |
| `PATCH /api/v1/records/{record_id}` | Core 01/02/03/04; module.workbook transport; source owner applies fields | OpenAPI, workbook route, facade and source store | Target lifecycle tests; workbook mutation, conflict, and browser tests | Explicit task/decision role, incident closure, no-op, auto-rebase, and same-field coverage at the route boundary | high | Preserve changed-fields-only writes and row-version high-water semantics. |
| `POST /api/v1/records/{record_id}/conflicts/{conflict_token}/resolve` | Core 03; module.workbook route; Revisions mechanics; source owner revalidation | OpenAPI, conflict route, target conflict facade | Workbook conflict-route and source-owner contribution tests | Task/decision token replay, authorization precedence, stale token refresh, and collection resolution | high | `keep_saved` creates no revision; mutating resolutions create a new attributed change set. |
| `POST /api/v1/records/{record_id}/supersede` decision branch | Core 01/02/03/04; module.workbook route; `tasksdecisions` aggregate command | OpenAPI decision envelope and target facade | Target supersession tests; frontend command-port test; sentinel browser scenario | HTTP role/visibility precedence, one change set, two revisions/intents, replay emits nothing, and all rollback paths are atomic | high | Timeline and Decision share the route but keep distinct owner responses. |
| `GET /ws/v1/incidents/{incident_id}` and `record_changed` | Core 01/03/04; Collaboration stream; Revisions intent appender | Collaboration route/protocol and Revisions appender | Collaboration suites; workbook row-wire sparse task patch test | Decision supersede emits exactly two canonical events after commit and none on rejection/replay | high | The target has no socket implementation but source revisions produce durable intents. |
| Task-request row semantics | Core 01 view registry; Core 02 lifecycle; Core 03 interaction | Task view schema and source/projection code | Target store tests; workbook mutation/query tests; browser tests | Full field/default/minimum-create, member refs, collection shapes, lifecycle matrix, and no-partial-write cases | high | Freeze `open`, current actor, and `normal` defaults plus done/completion coupling. |
| Decision row semantics and machine consistency | Core 01 view registry and supersede route; Core 02 lifecycle | Decision view schema and source/projection code | Target lifecycle, terminal matrix, inconsistency and supersession tests | Ordinary scalar and collection writes fail closed for inconsistent machines; exact source/target guard matrix | high | Direct creation or ordinary patch to `superseded` remains forbidden. |
| View-schema identity and generated field contracts | Core 01; authored view-schema contracts | Two authored JSON view schemas and generated Go/TypeScript projections | Contract/view-schema checks; frontend contract tests | Drift check plus explicit parity for field keys, enum order, reference contracts, and conflict classes | high | Generated files must be refreshed from owners, never hand-edited. |
| Saved views over task/decision schemas | Core 03 and `savedviews` | Saved-view contract and task-view browser coverage | Saved-view/workbook startup and accessibility tests | Query/layout replay over both immutable view-schema IDs after facade changes | medium | The target owns no saved-view persistence. |
| Projection refresh and incident/restore rebuild | Core 01; source provider intent plus `projections` storage owner | Provider catalog, provider SQL, projection adapters | Target store tests; projection query/rebuild tests | Real task/decision restore and incident rebuild parity, including links and supersession | high | Source tables and links remain authoritative. |
| Authorization and reference visibility | Core 04; workbook route guards; source reference validation | Role matrix, route handlers, facade membership/reference checks | Public coordination editor test; generic workbook authorization tests | Viewer/editor/reviewer/admin matrix, removed membership, inactive/nonmember owner refs, hidden cross-incident refs | high | UI visibility is not authorization. |
| Change sets, record revisions, and rollback | Core 01/02; Revisions coordinator; source rollback provider | Revisions appender and target provider contribution | Target tests; generic Revisions delete/restore and integration tests | Task and decision rollback for every scalar/lifecycle tuple plus collection non-interference | high | Link-backed collections are excluded from row-source rollback and handled by link history. |
| Import owner-create facades `tasksdecisions.*.import_create` | Core 01 Imports; source owner semantics | Import contract index, generated registry, import facade | Registry/characterization tests only | End-to-end task/decision import defaults, references, member refs, projections, revisions, rollback on failure | high | Existing evidence does not exercise the target facade's source behavior. |
| Incident-bundle family `tasks_decisions` | Core 01 Incident Portability; source owner port | Source catalog and target port | Generic catalog/path tests only | Export/import round trip and one fail-closed case for each of the four invariant IDs | high | Current validator implements only envelope type/scope. |
| Recovery authoritative-table contribution | Core 01/04 recovery; Recovery catalog | `recovery_state.go` and recovery assembly | Recovery catalog and database coverage tests | Preserve exact authoritative classification after any package move | low | No product mutation belongs in this contribution. |
| Reporting task/decision working material | Reporting NLSpec within snapshot-reporting; source provider | Reporting materializer/provider and boundary guard | Reporting boundary and integration tests | Preserve paths, content class, support refs, deleted-row exclusion, and snapshot-bound materialization | medium | No live query is permitted after export-model materialization. |
| Stable UI selectors and coordination controller outcomes | Core 03 behavior; UI Contracts and web workbook owners | Selector helpers, command ports, controller | Frontend unit row; generic/sentinel/public-route browser tests | Preserve semantic outcome normalization, late-result invalidation, save-state transitions, focus, and selector identities | medium | No grid-vendor contract belongs in the backend module. |
| Harness/test accounting | Adopted Testing Harness NLSpec; authored verification and test-family projections | Owner verification contract and three-row family manifest | Harness contract checks | Map portability to real portability evidence and account for rollback-provider tests | high | Evidence routing does not prove product behavior. |

### Portability behavior and invariant attribution

The future conformance correction MUST retain
`source_family_id='tasks_decisions'`, contract major `1`, dependency
`artifacts`, and the existing logical paths `data/task_requests.ndjson` and
`data/decisions.ndjson`. It MUST NOT introduce a bundle version, source-family
identity, invariant identity, route, error code, migration, view-schema change,
or frontend change.

Core 01 MUST first receive an adopted clarification assigning every semantic
condition to exactly one invariant:

| Precedence | Invariant ID | Exclusive rule ownership |
| --- | --- | --- |
| 1 | `tasks_decisions.envelope_type_scope` | Every task source row binds exactly one imported-incident envelope of type `task_request`, and every decision source row binds exactly one imported-incident envelope of type `decision`. Missing, different-incident, wrong-type, or multiply bound envelopes violate this invariant. |
| 2 | `tasks_decisions.lifecycle_legal` | Task and decision lifecycle tokens belong to their closed Core 02 vocabularies. Persisted decision status plus the active decision-to-decision `supersedes` relation resolves to exactly one legal decision-machine condition. Direct/source `superseded` state without its required relation, illegal superseder or target state, wrong relation direction, or target state inconsistent with the relation violates this invariant. Scalar tuple guards assigned below are excluded. |
| 3 | `tasks_decisions.dependent_fields_legal` | Task `status`, `owner_user_id`, `blocked_reason`, `completed_at`, and `created_at` form a legal Core 02 tuple. A task convenience decision reference agrees with its authoritative relation when both are materialized. Required Decision source members retain admitted presence/nullability, and projection-only `supersedes_record_id` never becomes authoritative source state. |
| 4 | `tasks_decisions.references_same_incident` | Task requester-party, linked-record, and decision references plus Decision support and affected-record references resolve to the required type, lifecycle, and incident. This owner-field invariant is primary for task/decision-derived links; `links_tags.endpoints_same_incident` remains the fallback for generic links without an owner-field contract. |

One semantic condition MUST map to one invariant ID. A candidate with multiple
defects MUST select the lowest precedence number above and then use stable owner
row identity only as a private tie-break. Selection MUST NOT depend on archive,
NDJSON, filesystem, map, or unsorted SQL order.

The port operations have this behavioral contract:

| Operation | Input | Output | Required behavior | Forbidden behavior |
| --- | --- | --- | --- | --- |
| `PrepareImport` | Bounded bundle capability and immutable import context | Port-bound opaque prepared value or typed failure | Strictly decode both admitted files; reject malformed, lossy, unknown, duplicate, noncanonical, or wrong-type input; validate row-local vocabulary and canonical form | Database, object-store, projection, revision, publication, job, or other visible mutation |
| `ApplyImportTx` | Coordinator-supplied transaction, matching prepared value, and import context | Exact affected-row success or typed failure | Use fixed owner-controlled parameterized SQL or SQLC and insert only admitted task/decision source rows | Decode source bytes; begin, nest, commit, or publish; conflict-ignore or silent row skipping |
| `ValidateImportTx` | Same transaction, matching prepared value, and import context | Success or typed `{source_family_id, invariant_id}` violation | Read deterministic applied source, envelope, and link facts; sort by stable identity; evaluate all four groups through source-owned policy | Generic first-false attribution, unsorted result dependence, repair, mutation, or publication |

The public failure remains exactly `incident_bundle_import_rejected` with
`retryable=false`, `reason_code='source_family_invalid'`,
`source_family_id='tasks_decisions'`, and one exact invariant ID. Public or
retained diagnostics MUST NOT expose a raw row value, supplied user identifier,
record title, SQL text, relation name, package name, storage path, staging
identifier, or hostile input. Any failure MUST roll back source rows, envelopes,
links, revisions, projections, administration, audit, idempotency success, job
success, and incident publication in the coordinator's single final transaction.

| Required portability test | Deliberate condition | Binary result |
| --- | --- | --- |
| `TestIncidentBundleTasksDecisionsRoundTrip` | Valid source rows, envelopes, actors, references, and supersession | Import succeeds and deterministic re-export matches the admitted state. |
| `TestIncidentBundleTasksDecisionsRejectsEnvelopeTypeScope` | Decision row binds a same-incident task envelope | Exact `tasks_decisions.envelope_type_scope`. |
| `TestIncidentBundleTasksDecisionsRejectsLifecycleIllegal` | Persisted superseded decision lacks the active supersession relation | Exact `tasks_decisions.lifecycle_legal`. |
| `TestIncidentBundleTasksDecisionsRejectsDependentFieldsIllegal` | Blocked task lacks nonempty reason or completion tuple is inconsistent | Exact `tasks_decisions.dependent_fields_legal`. |
| `TestIncidentBundleTasksDecisionsRejectsReferencesOutsideIncident` | Task decision, linked record, support, or affected-record target is foreign | Exact `tasks_decisions.references_same_incident`. |
| `TestIncidentBundleTasksDecisionsFailureSelectionIsOrderIndependent` | Candidate has envelope and dependent-field defects in permuted input order | Envelope invariant wins for every permutation. |
| `TestIncidentBundleTasksDecisionsCoordinatorRejectsWithoutPartialPublication` | Each semantic failure runs through the full coordinator | No incident or partial authoritative/derived state is visible. |
| `TestIncidentBundleTasksDecisionsDiagnosticsAreSafe` | Invalid input contains conspicuous secret-like and hostile values | Only the closed safe identifiers appear in public and retained diagnostics. |

Fixtures MUST be valid JSON and database-convertible before the deliberate
semantic defect is evaluated. Parser or database-constraint rejection alone does
not satisfy an invariant test.

### Imported owner-reference contract

Operation authorization and imported owner-reference validity are distinct and
MUST be evaluated in this order inside the serialized import-unit transaction:

| Stage | Owner | Required checks | Failure effect |
| --- | --- | --- | --- |
| Operation authorization | Imports and route/application boundary | Current actor, target-incident membership and role, incident lifecycle, target importability, and current source/mapping fingerprints | Dispatcher failure before the owner facade mutates |
| Owner-reference validity | `tasksdecisions` | Exact stable `user_id`, active account, and current membership of that referenced user in the target incident | Typed owner-validation failure with no owner effect |

The second stage MUST NOT authorize the actor, and the first stage MUST NOT make
an arbitrary owner reference valid. The owner facade failure MUST translate
through the existing Imports contract as `error.code='import_apply_blocked'` and
`error.details.reason_code='owner_create_validation_failed'`. Nested details MAY
identify the stable field key and reference-contract ID only when the registered
safe owner-error schema permits them; they MUST NOT disclose the supplied user ID
or distinguish nonexistent, inactive, and nonmember identities.

| Input or timing case | Required result |
| --- | --- |
| Explicit active current same-incident member | Accept the exact internal `user_id`; do not normalize or resolve aliases. |
| Explicit active account without current target-incident membership | Reject with `owner_create_validation_failed`; create no owner effect. |
| Explicit account belonging only to another incident | Same rejection and atomicity. |
| Inactive account with a retained membership row | Same rejection and atomicity. |
| Explicit JSON `null` | Reject because task/decision owner fields are not clearable. |
| Owner omitted | Default to the authenticated import actor and succeed only if that actor remains authorized at commit. |
| Membership removal commits before import validation | Import unit fails with no owner effect. |
| Import commits while membership remains valid | Record remains valid at its commit point; later membership change does not rewrite history. |
| Actor loses membership or required role before apply | Dispatcher authorization fails before owner mutation. |
| Exact successful replay | Return the committed unit result and create no duplicate owner effect. |
| Same key with changed normalized content | Preserve existing import idempotency conflict behavior and create no second owner record. |

A rejected unit MUST leave no source row, envelope, link, revision, change set,
projection row, collaboration intent, apply-journal success, owner result,
idempotency success, or durable successful unit outcome.

### Harness selector mapping

SL-06 MUST apply this authored mapping. The row identities below are exact and
MUST be used unchanged.

| Row disposition | Family and row identity | Exact selector | Verification and profiles |
| --- | --- | --- | --- |
| Retain store row, narrow claims | `module.tasksdecisions.store`; existing row `module.tasksdecisions.store.task_requests_and_decisions_persist_as_workbook_a08a0c4c47` | Preserve its five current exact test symbols | Retain behavior contract only; remove incident-portability verification; integration, `default`, `go_transaction_heavy`, `postgres_transaction` |
| Add portability row | Family `module.tasksdecisions.incident_portability`; row `module.tasksdecisions.incident_portability.source_port_invariants_and_atomicity` | Package `./internal/modules/tasksdecisions`; the exact ASCII-sorted selector below | Incident-portability verification only; collaborator `module.incidentbundles`; integration, `default`, `go_transaction_heavy`, `postgres_transaction` |
| Add rollback row | Family `module.tasksdecisions.rollback_provider`; row `module.tasksdecisions.rollback_provider.source_policy_null_and_owner_validation` | Package `./internal/modules/tasksdecisions/rollbackprovider`; `TestProvidersRejectInvalidOwnerValues`, `TestTaskSourcePreservesNullAndExcludesCollections` | Behavior contract only; unit, `none`, `none`, `none` |
| Add import row | Family `module.tasksdecisions.import_create`; row `module.tasksdecisions.import_create.owner_reference_contract_and_atomicity` | Package `./internal/app/importassembly`; `TestTasksDecisionsImportMembershipAndAuthorizationRaces`, `TestTasksDecisionsImportOwnerReferenceMatrix`, `TestTasksDecisionsImportReplayAndAtomicity` | Behavior contract only; collaborator `module.imports`; integration, `default`, `go_transaction_heavy`, `postgres_transaction` |
| Preserve frontend and browser rows | Existing `module.tasksdecisions.frontend_unit` and `module.tasksdecisions.browser` rows | Existing exact selectors | Unchanged unless their selected behavior changes |

The portability selector MUST contain exactly these symbols in this ASCII order:

```text
TestIncidentBundleTasksDecisionsCoordinatorRejectsWithoutPartialPublication
TestIncidentBundleTasksDecisionsDiagnosticsAreSafe
TestIncidentBundleTasksDecisionsFailureSelectionIsOrderIndependent
TestIncidentBundleTasksDecisionsRejectsDependentFieldsIllegal
TestIncidentBundleTasksDecisionsRejectsEnvelopeTypeScope
TestIncidentBundleTasksDecisionsRejectsLifecycleIllegal
TestIncidentBundleTasksDecisionsRejectsReferencesOutsideIncident
TestIncidentBundleTasksDecisionsRoundTrip
```

Every exact test symbol MUST be selected by exactly one active owner row. The
owner explanation MUST show the store, portability, rollback, frontend, and
browser rows once each with accurate runner, fixture, resource, runtime, evidence,
claim, and verification identities. Generated schedules and typed projections
MUST be produced through Make and MUST NOT be hand-edited.

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Workbook and supersede facades construct/use platform `authn` storage for generic replay and membership concerns | `workbook_facade.go`, `supersede_facade.go`; root imports include platform `authn` | Crosses source semantics, platform storage, and route policy in one facade | `must_fix` | App/workbook assembly supplies narrow membership and idempotency ports | Characterize precedence and replay first, then inject capabilities without changing transaction ownership. |
| Workbook facade mixes source mutation with records, links, projections, revisions, conflict, and transport status shaping | The 1,050-line facade performs the full transaction and returns HTTP status | Large blast radius and difficult owner testing | `should_fix` | `tasksdecisions` source facade plus generic Workbook/Revisions capabilities | WF-05 must define the narrow facade and compatibility adapter before movement. |
| Supersede facade combines legitimate aggregate semantics with generic idempotency/revision/projection plumbing | `supersede_facade.go` | Two-record, one-link, one-change-set atomicity is easy to break | `should_fix` | Task/decision aggregate remains source-owned; generic capabilities injected | Add exact revision/intent characterization before introducing seams. |
| Root package exposes low-level source methods, lifecycle types, DTOs, and helpers beyond its actual external caller needs | Exported-symbol inventory and workbook conversion adapters | Encourages cross-owner coupling and duplicate translation | `should_fix` | Thin `tasksdecisions` application facade | Preserve compatibility until all inbound callers migrate; shrink only after an import audit. |
| Workbook adapters duplicate task/decision DTO translation | `internal/modules/workbook/mutation_{store,contributions}.go` | Conversion drift can alter nulls, collection actions, or errors | `should_fix` | Explicit owner contribution contract at the workbook boundary | Characterize request/result/error parity, then centralize the adapter shape. |
| Lifecycle and enum knowledge is repeated in mutation and rollback paths | `mutations.go`, `rollbackprovider/provider.go` | Interactive writes and rollback can diverge | `should_fix` | Shared source-owned policy below the root facade | Choose an acyclic internal package in WF-05 and use it from every owner entry path. |
| Direct SQL writes to `task_requests` and `decisions` live in the source owner | `mutations.go`, portability and rollback providers; schema ownership manifest | Moving this SQL would weaken source ownership | `intentional/no_action` | `tasksdecisions` | Keep exact fixed SQL source-owned; do not replace it with generic table metadata. |
| Source-owned projection descriptor and derivation SQL remain under the target | `projectionprovider/*`; projection provider catalog | Moving intent to Projections would make storage infer domain semantics | `intentional/no_action` | `tasksdecisions` provider, `projections` storage coordinator | Preserve the current owner/provider split and narrow only the calling port. |
| Source-owner contributions exist for imports, portability, revisions, recovery, and reporting | Assembly catalogs call target constructors; coordinators validate complete catalogs | Treating adapters as catch-all logic could cause incorrect moves | `intentional/no_action` | Respective generic coordinator plus `tasksdecisions` provider | Keep providers source-constructed; move no generic coordination into them. |
| Incident-bundle descriptor declares four invariants but validation checks only envelope type/scope | Target source port and Core 01 closed invariant table | Invalid lifecycle/dependent/reference state can pass owner validation or map to the wrong invariant | `must_fix` | `tasksdecisions` incident-bundle source port | Add one characterization per invariant, then implement the three missing checks in an authorized behavior slice. |
| Import user validation checks active account but not same-incident membership | `validateImportActiveUserTx` versus Core 01 `incident_member_user_ref_v1` | Import can create an owner reference that interactive create rejects | `must_fix` | `tasksdecisions` import owner facade | Add active-nonmember characterization, then enforce membership in an authorized behavior correction. |
| Portability verification is assigned to the store row that does not call the portability port | Verification contract, test-family manifest, and target test source | Evidence accounting overstates coverage | `must_fix` | Harness owner projection for `module.tasksdecisions` | Add real portability tests and remap the authored family manifest; regenerate downstream schedules. |
| Rollback-provider unit tests are not named by a tasksdecisions owner-family row | Test-family search found no row for the provider package | Owner-focused runs can omit relevant source policy evidence | `should_fix` | Harness owner projection for `module.tasksdecisions` | Add or extend an authored row after the intended test package is settled. |
| Durable collaboration intents are appended through Revisions, but decision supersede has no exact event-count assertion | Revisions appender; target test asserts result changes but not persisted intents | Refactor could duplicate or omit events | `should_fix` | Revisions/Collaboration mechanics with tasksdecisions characterization | Assert two committed events and zero replay/rejection events before seam extraction. |
| Target contains no saved-view, frontend, grid-adapter, or vendor integration | File/import inventory and frontend reference search | Inventing a move would create a false boundary | `intentional/no_action` | Existing Saved Views, web workbook, UI Contracts, and grid adapter owners | Freeze downstream behavior; do not add these packages to the backend target. |
| Generated Go/TypeScript surfaces mirror task/decision view, OpenAPI, and import contracts | Generated-reference search under declared generated roots | Hand edits would drift from owners | `must_fix` | Authored contracts and Make generators | Any later contract change must update owner inputs and run Make generation/drift targets. |
| Current callers import both the root and target leaf providers | Projection assembly, Projections coordinator/tests, and target-root import scan | External leaf imports prevent the adopted root-only boundary and increase cycle risk during movement | `must_fix` | Root contribution facade plus application composition | Implement TD-DEC-004 in SL-01 through SL-03 and encode every permitted/prohibited edge in the backend boundary manifest. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Fix scope, authority, repository revision, and write boundary | This tracker only in the planning session | `make lint-markdown` | Tracker identifies a clean starting state and later-authorization boundary. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Account for every target file, inbound caller, dependency, and test | `internal/modules/tasksdecisions/**` | Read-only inventory; later `make backend-module-boundary-check` | All 20 files have a responsibility and risk row. |
| WF-02 | Contract-owner mapping | parallel | WF-01 | WF-03, WF-05 | Bind every observable surface to its authoritative owner and current evidence | Core owner docs, view schemas, OpenAPI, import/portability/projection contracts | `make generate-drift`; `make json-shape-check` later | Freeze map has owner, evidence, test posture, and risk for every contract. |
| WF-03 | Characterization test gap analysis | parallel | WF-01, WF-02 | WF-06, WF-07 | Define the exact conformance and behavior-preservation baseline before any production edit | Target tests; workbook, revisions, import, portability, frontend tests | Owner-focused test slices | Every required test has an exact behavior, selector posture, and binary result. |
| WF-04 | Boundary and coupling scan | parallel | WF-01 | WF-05 | Separate legitimate source ownership from generic coordination and platform coupling | Target facades, provider subpackages, app/workbook adapters, boundary manifests | Backend/frontend import-boundary targets later | Every finding is classified and has a proposed owner/action. |
| WF-05 | Facade and ownership boundary contract | chain | WF-02, WF-04 | WF-06 | Apply TD-DEC-004's root/private/leaf graph and capability boundaries without reopening topology design | Target root, app/workbook/projection assembly, generic coordinator ports, boundary manifest | Static boundary checks plus owner tests | Exact permitted and forbidden edges, contribution construction, and migration targets are fixed in Section 3. |
| WF-06 | Conformance-first slice sequencing | chain | WF-03, WF-05 | WF-07, WF-08 | Correct and evidence adopted behavior before beginning structural movement | Target and named owner/caller packages | Per-slice commands and gates in Section 7 | SL-00, SL-04, SL-05, and SL-06 precede SL-01 through SL-03. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-03, WF-06 | conformance gate, WF-08 | Align owner evidence with actual task/decision behavior and close the correction work order | Authored verification/test-family inputs and new tests | Owner explanation; `make harness-contract`; generation and drift checks | Portability and rollback evidence are selected exactly once and the correction gate is green. |
| WF-08 | Validation and final handoff | chain | conformance gate, WF-06, WF-07 | none | Run structural validation only after corrected behavior is green and record related/unrelated failures | No new design work; verification artifacts and tracker update | `make agent-finalize`; `make check` after focused targets | Both work orders record commands, results, run roots, blockers, and rollback state independently. |

## 7. Proposed Refactor Slice Plan

The 2026-08-01 10:14 EDT work order authorizes every slice below while retaining
separate conformance and structural gates. The mandatory order is:

```text
SL-00 -> SL-04 -> SL-05 -> SL-06 -> CG-01
CG-01 -> SL-01 -> SL-02 -> SL-03 -> SL-07
```

SL-04 and SL-05 correct repository/owner mismatches. SL-01 through SL-03 MUST
preserve the corrected behavior established at CG-01.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| SL-00 | WF-03 and conformance-correction authorization | Add exact characterization and semantic fixtures without moving production packages | Target import/portability tests, rollback tests, workbook/revision evidence | A parser or constraint failure could masquerade as semantic proof; tests could bind implementation details | Every Section 4 import and portability case; rollback tuples; authorization precedence; exact supersede revision/intent counts | `make test`; repeat through owner slices after SL-06 | Revert only new test/fixture source; do not change owners to make an incorrect assertion pass | Every required test reaches its intended semantic boundary and fails only for the documented current gap or passes the frozen behavior. |
| SL-04 | SL-00 and conformance-correction authorization | **Behavior correction:** revalidate imported task/decision owner values through `incident_member_user_ref_v1` immediately before mutation | Target import facade and target-specific import tests | Accepted input narrows; error redaction, defaults, replay, and unit atomicity can drift | Entire imported-owner matrix in Section 4 | `make test`; after SL-06, `make service-backed-test-slice OWNER=module.tasksdecisions` | Revert this isolated correction if owner review rejects it; retain the characterization evidence | Interactive and imported owner references have identical exact active same-incident membership semantics at commit. |
| SL-05 | SL-00 and conformance-correction authorization | **Owner clarification and behavior correction:** adopt the exclusive Core 01 attribution table, then validate all four invariant groups deterministically | Core 01 owner text; target source port, portability policy/helpers, and tests; Appendix F only if navigation changes | Invalid-input attribution changes; wrong precedence or unsafe details violate a versioned portability contract | Eight exact portability tests in Section 4 | `make test`; after SL-06, focused service-backed owner row | Revert owner clarification and implementation together before adoption; after adoption, repair forward under owner authority | Core text is adopted, all descriptor invariants are reachable, multiple defects select deterministically, and no invalid candidate publishes state. |
| SL-06 | SL-04 and SL-05 | Repair authored verification/test-family ownership and regenerate all downstream harness projections through Make | Exact authored files in Section 4; generated schedules only through Make | Evidence can be duplicated, omitted, or assigned to unrelated behavior | Preserve five store tests and frontend/browser rows; select eight portability and two rollback tests exactly once | `make explain-test-owner OWNER=module.tasksdecisions`; owner slices; `make harness-contract`; generation/drift gates | Revert authored inputs and generated projections as one slice; never hand-edit generated files | Owner explanation matches the Section 4 harness mapping and all selected obligations have one compatible terminal result. |
| SL-01 | CG-01, WF-05, and structural-refactor authorization | Introduce the seven narrow injected capabilities while retaining compatible constructors and behavior | Target workbook/supersede/import facades and application composition | Transaction ownership, failure precedence, replay identity, and event counts | Preserve the complete corrected owner baseline | Owner unit/service-backed slices; `make backend-module-boundary-check` | Keep compatibility adapters until all new seams pass; revert the seam slice as one unit | Root facades no longer construct concrete cross-owner stores and every capability satisfies Section 3. |
| SL-02 | SL-01 | Move shared task/decision policy and fixed source persistence into the adopted private packages | Validation, mutation, import, rollback, portability, and source SQL paths | Lifecycle/default/reference behavior or SQL atomicity can drift | Cross-entry parity for interactive, import, rollback, and portability; preserve all corrected tests | Owner slices and backend boundary check | Keep old wrappers until all target-local callers use the private packages; revert without schema change | One private policy implementation owns each rule, fixed SQL resides in private source, and no forbidden dependency exists. |
| SL-03 | SL-02 | Migrate every external caller to the root facade/contribution constructors, thin the root surface, and encode boundary rules | Target root/leaves, workbook and projection callers, app assemblies, backend boundary manifest | Package cycles, public error translation, nulls, collections, query descriptors, and provider registration | Workbook route/conflict/row-wire, projection query/rebuild, frontend unit, and browser rows | Owner slices; both import-boundary targets; broader browser target | Migrate behind compatibility wrappers, remove them only after the final caller moves, and revert caller groups independently | No external leaf import remains, root exports only the adopted surface, compilation is acyclic, and boundary checks pass. |
| SL-07 | SL-03 | Run final narrow-to-broad verification and produce the implementation handoff without new design changes | All authorized changes and retained result metadata | A broad failure can expose cross-owner drift missed by focused rows | Preserve every focused, integration, browser, generated, and boundary result | `make agent-finalize`, then `make check`; retain only successful qualifying `RESULTS_DIR` evidence | Revert only the causal slice and retain/classify failure evidence | Every required target passes or the handoff records target, run root, relatedness, and rollback state. |

### CG-01 conformance-correction gate

CG-01 passes only when SL-00, SL-04, SL-05, and SL-06 are complete; the Core 01
clarification is adopted; exact owner unit and service-backed slices pass; all
four portability invariants and imported-owner cases pass; invalid input leaves
no partial state; owner evidence is exact and non-overlapping; and generation is
clean. Structural work MUST NOT begin while any CG-01 condition is false.

### Work-order authorization boundaries

| Work order | Permitted after explicit authorization | Prohibited | Exit |
| --- | --- | --- | --- |
| Conformance correction | SL-00, SL-04, SL-05, and SL-06; narrow Core 01 attribution clarification; Appendix F navigation only if needed; target tests and implementation; authored tasksdecisions verification/test-family inputs; generated outputs through Make; tracker evidence | Unrelated package moves; route/OpenAPI/view-schema identity changes; migrations/DDL; frontend/grid changes; hand-edited generated files; generalized workflow/callback infrastructure | CG-01 passes in full. |
| Structural refactor | After CG-01, SL-01 through SL-03 and SL-07; private policy/source packages; typed capability adapters; root contribution wrappers; caller migration; temporary compatibility wrappers; boundary-manifest edits | Behavior, route/status/envelope, lifecycle, import default/provenance, valid portability output, transaction, revision/intent/replay count, frontend, selector, grid, or schema changes | Corrected baseline and all focused/broad gates pass; temporary wrappers are removed only after caller migration. |

Authorization of one work order MUST NOT ordinarily be inferred as authorization
of the other. RB-005 is closed for this effort by the explicit instruction to
implement the complete ordered plan and its stated write scope.

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| documentation | `make lint-markdown` | This tracker and authored Markdown | yes | Required for this tracker-only session; it is not product validation. |
| owner explanation | `make explain-test-owner OWNER=module.tasksdecisions` | Exact owner rows, selectors, fixtures, resources, evidence, and verification routing | yes | Run before and after SL-06; the latter output MUST match the Section 4 mapping. |
| unit | `make test-slice OWNER=module.tasksdecisions` | Owner-selected non-service rows, including frontend and the planned rollback-provider row | yes | Every intended unit symbol MUST appear exactly once. |
| integration | `make service-backed-test-slice OWNER=module.tasksdecisions` | Owner-selected Postgres-backed store and portability evidence | yes | After SL-06 this MUST select both service-backed families without borrowing unrelated tests. |
| e2e/browser | `make service-backed-test-slice OWNER=module.tasksdecisions ROWS=module.tasksdecisions.browser.the_browser_workbook_opens_task_requests_and_dec_8048a75ceb` | Current owner browser scenario | yes | Broaden to `make browser-e2e-webserver-backed` after cross-layer changes. |
| generated drift | `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check` | Generated Go/TypeScript, harness, policy, and JSON contract projections | yes | Generate only after authorized owner/harness inputs change; never hand-edit generated roots. |
| import-boundary/static | `make backend-module-boundary-check`; `make frontend-import-boundary-check` | Backend owner imports and downstream frontend boundary preservation | yes | Backend is mandatory for package movement; frontend guards against accidental shell/grid coupling. |
| harness accounting | `make harness-contract` | Verification contracts, test-family manifests, topology, and scheduler accounting | yes | Required in SL-06 and before CG-01. |
| full check | `make agent-finalize`; `make check` | Final repository developer gate | no | Run after focused validation. Supply `RESULTS_DIR` only when retaining a successful qualifying run. |

The conformance-correction work order MUST run, in order:

1. `make explain-test-owner OWNER=module.tasksdecisions`;
2. `make test-slice OWNER=module.tasksdecisions`;
3. `make service-backed-test-slice OWNER=module.tasksdecisions`;
4. `make harness-contract`;
5. `make generate`;
6. `make generate-drift`;
7. `make generated-artifact-policy-check`; and
8. `make json-shape-check`.

The structural-refactor work order MUST then add
`make backend-module-boundary-check`, `make frontend-import-boundary-check`, the
focused browser row, `make browser-e2e-webserver-backed`, `make agent-finalize`,
and `make check`, in that narrow-to-broad order. A failed target MUST be reported
with its run root, relevant summary, and relatedness; it MUST NOT be described as
product success.

Command discovery completed through Make-owned public targets. No product test,
integration suite, browser suite, drift gate, generation target, or full check
was run during either documentation-only tracker session.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| TD-001 | Establish scope, authority order, and write boundary | WF-00 | DONE | none | Section 1 and clean pre-edit worktree | Tracker is the only touched file and later authorization is explicit. |
| TD-002 | Inventory all target files and callers | WF-01 | DONE | TD-001 | Section 2; 20-file live inventory | Every target file has responsibility, callers, dependencies, tests, owner, and risk. |
| TD-003 | Map observable contracts to owners and evidence | WF-02 | DONE | TD-002 | Section 4 | Every discovered contract has an owner and test posture. |
| TD-004 | Diagnose legitimate and misplaced responsibilities | WF-04 | DONE | TD-002 | Sections 3 and 5 | Findings use only the allowed classifications and name proposed owners. |
| TD-005 | Define characterization additions | WF-03 | DONE | TD-003 | Section 4 required-test column; SL-00 | Missing coverage is behavior-specific and owner-routable. |
| TD-006 | Adopt an acyclic facade/subpackage redesign | WF-05 | DONE | TD-004 | TD-DEC-004 and Section 3 | Root/private/leaf topology, capability boundaries, and forbidden edges are decision-complete. |
| TD-007 | Add missing characterization tests | WF-03 | DONE | TD-005, resolved RB-005 authorization | `internal/modules/tasksdecisions/incident_bundle_source_port_test.go`; `internal/app/importassembly/tasksdecisions_integration_test.go`; strengthened supersession evidence in `task_decisions_store_test.go` | Exact portability/import symbols, scalar/null/collection behavior, and exact supersession durable-effect counts are characterized without package movement. |
| TD-008 | Introduce dependency seams and thin the root facade | WF-05, WF-06 | DONE | CG-01, structural authorization | SL-01 through SL-03 | Corrected behavior remains stable and the adopted root-only graph is mechanically proven. |
| TD-009 | Correct import incident-member validation | WF-06 | DONE | TD-007, conformance authorization | `internal/modules/tasksdecisions/import_create.go`; TD-DEC-002; SL-04 | Present owner UUIDs and actor-default owners use the same active same-incident member validation immediately before mutation; explicit null fails closed. |
| TD-010 | Complete incident-bundle invariant validation | WF-06 | DONE | TD-007, adopted Core clarification, conformance authorization | Core 01 Tasks/Decisions attribution table; typed prepared source port and fixed inserts; SL-05 | Exact shape/canonical admission, deterministic four-group validation, fixed apply SQL, and closed safe failures are implemented. |
| TD-011 | Repair verification and harness accounting | WF-07 | DONE | TD-009, TD-010, conformance authorization | `tools/test_families/module.tasksdecisions.json`; generated topology; measured duration baselines; SL-06 | Store, portability, rollback, frontend, browser, and application import evidence select exactly once. |
| TD-012 | Run implementation verification and final handoff | WF-08 | DONE | CG-01, TD-008 through TD-011 | SL-07 evidence rooted at `.cartulary/test-results/20260801T163132Z-p502410`; non-qualifying retention attempt `.cartulary/test-results/20260801T163441Z-p622068` | Both work-order ladders pass and the timing-only retention rejection is fully classified. |
| TD-013 | Preserve frontend, saved-view, WebSocket, and grid contracts without moving their implementation into the target | WF-02, WF-04 | DONE | TD-003 | Sections 3 through 5 | These surfaces are frozen as downstream risks and retain their existing owners. |
| TD-014 | Close RB-001 through RB-004 as technical planning decisions | WF-02 through WF-06 | DONE | TD-003 through TD-006 | TD-DEC-001 through TD-DEC-005 | No technical design choice remains; execution is authorized and remains gated by CG-01. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-01 08:46 EDT | Codex planning-only tracker session | Authority and write scope fixed; tracker did not previously exist | Inspected framework, domain, Core 00-04, Reporting and Testing Harness NLSpecs; touched only this tracker | `sed`, `rg`, `find`, `git status --short` | All 20 target files exist; pre-edit tracked worktree was clean | Implementation is outside this task | Run Markdown lint and hand off the tracker. |
| 2026-08-01 09:50 EDT | Codex NLSpec-style tracker revision | Prior scope retained; analysis recommendations converted into tracker-owned decisions without changing product authority | Inspected NLSpec writing doctrine, analysis notes, tracker, framework, and exact owner clauses; touched only this tracker | `sed`, `rg`, `git status --short`, `sha256sum` | RB-001 through RB-004 are decision-complete; RB-005 remains the sole authorization gate | No implementation work order is authorized | Use the exact work-order text in Section 7 when implementation is requested. |
| 2026-08-01 10:14 EDT | Codex complete remediation work order | Full conformance and structural execution authorized; SL-00 started | Baseline revision `6fc76a1731fbd50d9aaac4245448f4ab51323e3f`; tracker activation only | `git status --short`; `git rev-parse HEAD`; `make lint-markdown` | Clean tracked worktree; baseline Markdown lint passed at `.cartulary/test-results/20260801T141439Z-p3426800` | CG-01 blocks structural edits until SL-00, SL-04, SL-05, and SL-06 are green | Add characterization evidence only, then close SL-00 in the tracker before SL-04. |
| 2026-08-01 10:37 EDT | Codex SL-00 | Characterization complete; SL-04 started | Added eight exact portability tests, three application-composed import tests, and supersession revision/intent/change-set assertions; production behavior unchanged | `make format`; `make test`; `make explain-run RESULTS_DIR=.cartulary/test-results/20260801T142614Z-p3436990`; `make test` | First run exposed test-only import cycles and was related to SL-00; external-package correction followed. Final `make test` passed 1,002 tests at `.cartulary/test-results/20260801T143302Z-p3540240`. Exact new symbols remain intentionally outside active owner selection until SL-06; source inspection confines their expected assertions to SL-04 and SL-05. | SL-04 and SL-05 semantic corrections remain; CG-01 remains closed | Enforce the incident-member owner-reference contract in SL-04 only. |
| 2026-08-01 10:46 EDT | Codex SL-04 | Imported-owner conformance correction complete; SL-05 started | Changed only `internal/modules/tasksdecisions/import_create.go` plus tracker evidence | `make format`; `make test` | Active suite passed 1,002 tests at `.cartulary/test-results/20260801T144040Z-p3635879`; exact import selectors remain compilation-checked but inactive until SL-06 makes their ownership truthful | SL-05 portability remains; CG-01 remains closed | Adopt the Core 01 attribution clarification before changing the source port. |
| 2026-08-01 11:06 EDT | Codex SL-05 | Portability conformance correction complete; SL-06 started | Adopted the Core 01 exclusive attribution table; added typed strict preparation, prepared-value apply, fixed SQL, affected-row checks, and deterministic envelope/lifecycle/dependent/reference validation; corrected stale portability fixture tokens | `make format`; focused `make service-backed-test-slice OWNER=module.incidentbundles ROWS=...`; `make test`; `make lint-markdown` | Two affected Incident Bundle integration rows passed at `.cartulary/test-results/20260801T145714Z-p3827192`. First full rerun had no product-test failure but ended on unrelated harness resource conflict at `.cartulary/test-results/20260801T145753Z-p3829063`; unchanged rerun passed 1,002 tests at `.cartulary/test-results/20260801T150207Z-p3917652`; Markdown lint passed at `.cartulary/test-results/20260801T150633Z-p4006082`. | Exact new target symbols still await SL-06 owner routing; CG-01 remains closed | Add authored store/import/portability/rollback rows and generate projections through Make. |
| 2026-08-01 11:29 EDT | Codex SL-06 and CG-01 | Harness ownership repair complete; conformance gate passed; SL-01 started | Narrowed the store claim; added exact import, portability, and rollback rows; regenerated `tools/scheduler_manifest.json` and `tools/execution_topology_render_index.json` through Make; updated measured Go duration baselines and harness cardinalities | Prescribed eight-command SL-06 ladder; corrective `make test-service-backed`; `make go-test-duration-baselines RESULTS_DIR=... PRUNE_OBSERVED_PACKAGES=1` | Final ladder passed: owner unit `.cartulary/test-results/20260801T152636Z-p108794`, service-backed `.cartulary/test-results/20260801T152716Z-p127526`, harness `.cartulary/test-results/20260801T152753Z-p146099`, generation `.cartulary/test-results/20260801T152824Z-p147401`, drift `.cartulary/test-results/20260801T152836Z-p149645`, artifact policy `.cartulary/test-results/20260801T152850Z-p153600`, and JSON shape `.cartulary/test-results/20260801T152901Z-p153998`. A first full service run hit an unrelated Collaboration revocation-order race; a pre-generation retry exposed the related stale-topology selector mismatch. After Make generation, full service-backed evidence passed at `.cartulary/test-results/20260801T152230Z-p24759` and supplied measured baselines. | None for conformance; rollback remains the whole authored harness slice plus generated projections, with no runtime rollback required | Begin only SL-01 dependency seams; preserve the green CG-01 behavior baseline. |
| 2026-08-01 11:40 EDT | Codex SL-01 | Seven-capability injection complete; SL-02 started | Added consumer-owned incident-state, member-reference, idempotency, record-envelope, link, projection, and revision contracts; application composition now constructs every peer/platform adapter; Workbook and supersede commands carry actor UUIDs; owner results carry rows and supersession facts while Workbook rebuilds the unchanged status/envelope | `make format`; owner unit and service-backed slices; `make backend-module-boundary-check` | Unit slice passed at `.cartulary/test-results/20260801T153843Z-p163569`; service-backed slice passed at `.cartulary/test-results/20260801T153924Z-p183595`; final boundary check passed at `.cartulary/test-results/20260801T154039Z-p203221`. Its first run at `.cartulary/test-results/20260801T154002Z-p202630` exposed three related static-policy omissions for the already-adopted source-port imports and `record_links` fact read; the exact owner edges were added without widening runtime access. Existing application-composed tests retain failure precedence, replay identity, one transaction, and revision/intent counts. | Temporary internal constructors and direct cross-owner types remain until SL-02/SL-03; this slice rolls back as capability contracts, app adapters, facade signatures, Workbook translation, and the three exact boundary entries | Consolidate policy and transaction-bound source persistence in SL-02 only. |
| 2026-08-01 11:57 EDT | Codex SL-02 | Private policy/source consolidation complete; SL-03 started | Added `internal/policy` as the only closed vocabulary, defaults, create/patch, lifecycle, decision-machine, reference registry, and portability-invariant kernel; added `internal/source` for fixed field updates, member/record reference checks, lifecycle/machine and portability fact loads, and PostgreSQL classification; migrated interactive mutation, import validation, rollback, supersession, and portability; admitted valid canceled/done tasks with null owners during portability | `make format`; owner unit and service-backed slices; `make backend-module-boundary-check`; static searches for interpolated source SQL and `authn.IsUniqueViolation` | Final unit slice passed at `.cartulary/test-results/20260801T155540Z-p275267`; final service-backed slice, including the extended null-owner round trip, passed at `.cartulary/test-results/20260801T155621Z-p294280`; boundary check passed at `.cartulary/test-results/20260801T155658Z-p313092`. The first unit attempt at `.cartulary/test-results/20260801T154741Z-p211420` failed on a related missing `strings` import; the first boundary attempt at `.cartulary/test-results/20260801T155500Z-p271485` truthfully identified the new private source files until their exact Records/Links read edges were recorded. No dynamic source-column SQL or target dependency on `authn.IsUniqueViolation` remains. | No DDL or data rollback exists; revert policy/source, root forwarding aliases, migrated entry paths, exact boundary entries, and parity assertions as one internal slice | Expose root contribution constructors, migrate every external leaf caller, and encode the permanent graph in SL-03 only. |
| 2026-08-01 12:35 EDT | Codex SL-07 | Final validation complete; TD-012 and the handoff closed | Validation and tracker only after SL-03; no new design or product change | Tasks/Decisions owner unit/service-backed slices; backend/frontend boundaries; focused owner and broader webserver-backed browser checks; harness, generation, drift, artifact-policy, and JSON checks; `make agent-finalize`; `make check`; conditional retained-run `make agent-finalize RESULTS_DIR=...`; `git diff --check` | Owner unit `.cartulary/test-results/20260801T162238Z-p418562` and service-backed/focused browser `.cartulary/test-results/20260801T162315Z-p437292`; boundaries `.cartulary/test-results/20260801T162356Z-p456380` and `.cartulary/test-results/20260801T162403Z-p456757`; broader browser `.cartulary/test-results/20260801T162416Z-p457349`; harness `.cartulary/test-results/20260801T162910Z-p487094`; generate `.cartulary/test-results/20260801T163015Z-p488620`; drift `.cartulary/test-results/20260801T163028Z-p490855`; artifact policy `.cartulary/test-results/20260801T163043Z-p494798`; JSON shape `.cartulary/test-results/20260801T163054Z-p495190`; no-retained-run finalizer `.cartulary/test-results/20260801T163105Z-p495783`; full `make check` passed 192/192 work units and 866 tests at `.cartulary/test-results/20260801T163132Z-p502410`. The conditional retention attempt failed at `.cartulary/test-results/20260801T163441Z-p622068`: the successful check root was not qualifying because warm build readiness, the service-backed warm budget, and one Workbook peer-timing ratio exceeded harness thresholds. This is timing-only, unrelated to every remediation slice, caused no repository mutation, and is not reported as retained evidence. | No rollback is indicated by product, contract, boundary, browser, generation, or full-check evidence. If timing retention is desired later, rerun a fresh warm `make check`; do not alter product code or thresholds for this handoff. | None; implementation and handoff are complete. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-01 08:46 EDT | Codex planning-only tracker session | Legitimate source owner diagnosed as a mixed-responsibility facade/provider package | Inspected all target Go files and app/workbook/projection/revision/import/portability/reporting/recovery callers; touched only this tracker | `rg` import/symbol searches; `sed` exact-source reads; `jq` boundary projections | Source semantics stay in `tasksdecisions`; generic coordination should move behind injected ports | Exact acyclic package topology is deferred to WF-05 | Characterize, then design seams before moving code. |
| 2026-08-01 09:50 EDT | Codex NLSpec-style tracker revision | Root-only facade, private policy/source, and source-owned leaf-provider topology adopted | Reinspected target leaf imports, Projection assembly/coordinator callers, and backend boundary manifest; touched only this tracker | `rg`, `sed`, `jq` | Six capability contracts and every permitted/prohibited edge are fixed by TD-DEC-004 | Mechanical proof and caller migration await structural authorization | After CG-01, implement SL-01 through SL-03 without reopening topology design. |
| 2026-08-01 12:35 EDT | Codex SL-07 | Permanent topology is implemented and mechanically enforced | Root contributions, private policy/source, provider leaves, application callers, Projections/Reporting consumers, and `tools/backend_module_boundaries.json` | Static import search; owner slices; `make backend-module-boundary-check`; `make check` | No external leaf import or reverse leaf-to-root edge remains; exact coordinator-contract and source-table paths pass | None | Maintain new callers through a root contribution and update the boundary manifest in the same change. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-01 08:46 EDT | Codex planning-only tracker session | Frontend remains a generic workbook coordination surface; no target-local frontend or vendor code exists | Inspected coordination controller/command ports/tests, browser scenarios, view-contract constants, UI selector helpers; touched only this tracker | `rg` and `sed` | Task lifecycle and decision supersede use stable view IDs, public operations, row versions, save-state hooks, and semantic selectors | Cross-layer tests must remain stable during backend seams | Preserve controller outcomes, selectors, focus/save state, and grid-adapter isolation. |
| 2026-08-01 09:50 EDT | Codex NLSpec-style tracker revision | Frontend exclusion converted into an explicit forbidden-edge and compatibility requirement | Reused inspected frontend/controller/selector evidence; touched only this tracker | Tracker reconciliation only | Structural work MUST preserve routes, selectors, focus, save state, and grid isolation and MUST add no frontend-to-target internals | Browser verification awaits later structural authorization | Run the existing frontend and browser owner rows after caller migration. |
| 2026-08-01 12:35 EDT | Codex SL-07 | Frontend remained unchanged and isolated from target internals | Frontend boundary projection and existing task/decision browser scenario | `make frontend-import-boundary-check`; focused owner service-backed slice; `make browser-e2e-webserver-backed`; `make check` | Boundary, focused workflow, and broader browser evidence pass; selectors and behavior are unchanged | None | No frontend migration or compatibility action is required. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-01 08:46 EDT | Codex planning-only tracker session | Public routes, view schemas, imports, projections, portability, reporting, recovery, and generated surfaces mapped | Inspected authored contracts and searched declared generated roots; touched only this tracker | `jq`, `rg`, `sed` | No generated file is an edit source; no owner contradiction found | Portability and import implementations differ from adopted owners | Correct owners in isolated authorized slices, then regenerate only through Make. |
| 2026-08-01 09:50 EDT | Codex NLSpec-style tracker revision | Exact portability attribution, operation contracts, safe failure shape, and stable identity freeze are specified | Reinspected Core 01 REQ-01-628, REQ-01-619, REQ-01-639 through REQ-01-643 and target port code; touched only this tracker | `rg`, `sed` | TD-DEC-001 requires a narrow Core clarification before implementation; no Core or generated file changed | Core adoption and SL-05 await conformance authorization | Apply owner text first, then implementation, authored projections, and Make generation. |
| 2026-08-01 12:35 EDT | Codex SL-07 | Owner clarification, projection-provider manifest, harness inputs, and generated projections are synchronized | Core 01, `contracts/projection-providers/index.json`, authored test-family/boundary inputs, and Make-owned generated outputs | `make harness-contract`; `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make check` | All contract, generation, drift, artifact, and shape gates pass | None | Regenerate through Make after future authored owner or topology changes. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-01 08:46 EDT | Codex planning-only tracker session | Three owner rows discovered: one frontend unit, one Postgres-backed store row, one browser row | Inspected verification contract, test-family manifest, target/provider/workbook/revision/browser tests; touched only this tracker | `make help`; `make task-guide ROLE=module-author OWNER=module.tasksdecisions`; `make explain-test-owner OWNER=module.tasksdecisions`; relevant `make help-all` and `make explain-target` queries | Discovery succeeded; no validation suite was run | Portability verification points to tests that do not exercise portability; rollback-provider tests are not owner-mapped | Add characterization and repair authored accounting in WF-07. |
| 2026-08-01 08:46 EDT | Codex planning-only tracker session | One investigation invocation was invalid | Touched only this tracker | `make target-plan-json OWNER=module.tasksdecisions` | Failed with usage error because `target-plan-json` does not declare `OWNER`; this is not a product failure | None beyond using the correct command shape | Do not repeat with `OWNER`; use task-guide/explain-test-owner for owner routing. |
| 2026-08-01 09:50 EDT | Codex NLSpec-style tracker revision | Exact future store, portability, rollback, frontend, and browser row dispositions are specified | Reinspected tasksdecisions verification/test-family JSON and exact rollback test symbols; touched only this tracker | `jq`, `rg`, `sed` | Store loses the false portability claim; eight portability and two rollback symbols receive exact non-overlapping rows | Authored harness inputs were not authorized or changed | Implement SL-06 only after SL-04 and SL-05; verify with owner explanation and harness/drift gates. |
| 2026-08-01 10:37 EDT | Codex SL-00 | Exact test symbols now exist; harness ownership remains deliberately unchanged until SL-06 | Added external-boundary portability and application-composed import fixtures plus supersession durable-effect checks | `make format`; `make test` | Full active harness passed at `.cartulary/test-results/20260801T143302Z-p3540240`; the current selector catalog does not yet execute the new symbols, which is the already-scoped SL-06 routing gap | SL-04/SL-05 must correct semantics before SL-06 activates these rows | Apply only the imported-owner correction, then rerun broad active evidence. |
| 2026-08-01 12:35 EDT | Codex SL-07 | Exact owner accounting and full repository evidence are green | Six active owner rows, generated schedules, browser evidence, and full check run | Owner slices; browser target; harness/generation ladder; no-retained and conditional-retained finalizers; `make check` | All product/evidence gates pass. The full check root is successful but not retained because the conditional finalizer rejected timing health at `.cartulary/test-results/20260801T163441Z-p622068`; no test failed and no mutation occurred. | No product blocker; retained-run metadata is intentionally absent | A later maintenance run may produce qualifying warm evidence; this handoff must continue to cite the successful non-retained check root. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-01 08:46 EDT | Codex planning-only tracker session | Route-time incident authorization is outside the source module, while source reference validity is partly checked inside it | Inspected Core 04, workbook route guards, facade membership/reference queries, public-route tests; touched only this tracker | `rg`, `sed` | Interactive owner refs require active membership; import refs currently require only an active account | RB-002 blocks a conformance claim for import owner refs | Add nonmember characterization before the authorized correction. |
| 2026-08-01 09:50 EDT | Codex NLSpec-style tracker revision | Import operation authorization and owner-reference validity are separate ordered transaction checks | Reinspected Core 01 direct-reference/import clauses and `validateImportActiveUserTx`; touched only this tracker | `rg`, `sed` | Exact, null, omission, race, replay, error-redaction, and no-partial-effect behavior is fixed by TD-DEC-002 | SL-04 remains unimplemented pending conformance authorization | Add the full matrix before enforcing membership at unit commit. |
| 2026-08-01 10:37 EDT | Codex SL-00 | Imported-owner security boundary characterized through the application-composed owner registry | Added current-member, nonmember, foreign-only, inactive, explicit-null, omission/default, membership-removal, actor-authorization-loss, replay/atomicity, and no-durable-effect cases for both views | `make test` | Active baseline remained green; exact new selectors are ready for truthful SL-06 ownership after conformance correction | Current active-user-only implementation still admits nonmembers and treats explicit null as omission | Implement exact same-incident membership immediately before source mutation in SL-04. |
| 2026-08-01 10:46 EDT | Codex SL-04 | Import source now enforces the shared owner-reference contract at the unit transaction boundary | Effective owner is the exact supplied UUID or actor default; active same-incident membership is checked immediately before any record insert; present null/non-UUID owner values fail as `invalid_value`; direct errors contain no supplied identifier | `make format`; `make test` | Active broad evidence passed at `.cartulary/test-results/20260801T144040Z-p3635879`; no DDL or historical rewrite | Exact selector execution awaits the already-sequenced SL-06 routing slice | Add no compatibility mode; proceed to portability integrity. |
| 2026-08-01 12:35 EDT | Codex SL-07 | Imported-owner authorization/reference separation remains closed and fail-safe | Application-composed import tests and shared member-reference capability | Owner service-backed slice; `make check` | Member/nonmember/inactive/foreign/null/omitted/race/replay/atomicity matrix passes without diagnostic leakage or partial state | None | Preserve the shared member-reference policy for any future ingest path. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-01 08:46 EDT | Codex planning-only tracker session | Tracker content is complete; no production refactor was performed | Touched only `docs/handoffs/tasksdecisions-module-refactor-tracker.md` | `make lint-markdown` | Passed; structural plan, slices, validation ladder, and blockers are recorded | RB-001 through RB-005 | Begin only with a later authorized characterization task; do not start package movement first. |
| 2026-08-01 09:50 EDT | Codex NLSpec-style tracker revision | Decision-complete revision finished; no owner, production, test, contract, generated, migration, frontend, or harness edit performed | Touched only `docs/handoffs/tasksdecisions-module-refactor-tracker.md` | `make lint-markdown` | Passed; prior history preserved and normative decisions, mappings, work orders, and completion gates added | RB-005 only | Obtain explicit authorization for one Section 7 work order; do not infer authorization for the other. |
| 2026-08-01 12:35 EDT | Codex SL-07 | Remediation and handoff are complete | Entire authorized change set and validation artifacts | Full Section 8 ladder; final tracker lint | Product, boundary, browser, contract, generation, and full-check evidence pass; retained evidence was not published because the successful check root failed timing qualification | No open implementation blocker. Timing qualification is maintenance-only and unrelated to the change. | Use this tracker and the successful check root for handoff; rerun a warm check only if retained timing evidence is later required. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Four portability invariants require exclusive runtime validation and deterministic attribution. | Without the Section 4 partition, invalid state may pass or receive implementation-defined attribution. | Adopt the narrow Core 01 clarification, implement SL-05, and retain exact semantic/no-partial-state evidence. | RESOLVED: owner clarification, typed source port, and eight-test evidence pass |
| RB-002 | Imported owner references require current same-incident membership at unit commit. | Current import validation is weaker than `incident_member_user_ref_v1` and interactive mutation. | Implement SL-04 against the complete input/race/replay matrix. | RESOLVED: shared member-reference validation and three application-composed tests pass |
| RB-003 | Portability and rollback verification require exact truthful owner rows. | Current accounting overstates portability coverage and omits rollback policy evidence. | Implement the Section 4 mapping through authored inputs, Make generation, and SL-06 validation. | RESOLVED: six exact owner rows, generated topology, and harness contract pass |
| RB-004 | The permanent root/private/leaf package graph requires mechanical enforcement. | Direct external leaf imports violate the adopted root-only posture and can create cycles during movement. | Implement SL-01 through SL-03, update the backend boundary manifest, compile, and pass boundary checks. | RESOLVED: all external callers use the root; leaf/private/root direction, exact provider edges, and source-table access are enforced and green |
| RB-005 | Complete implementation authority was required. | Every action changed owner text, code, tests, contracts, generated outputs, harness inputs, or boundaries outside the prior tracker-only task. | The 2026-08-01 10:14 EDT user instruction explicitly authorized the complete ordered conformance and structural effort. | RESOLVED: full work order authorized and completed; CG-01 was observed before structural work |

There is no remaining technical planning question or unresolved blocker.
RB-001 through RB-005 retain stable IDs for later audit and are all resolved.

There is no `BLOCKED: owner contradiction` entry because no conflict between
applicable owner documents was found.

## 12. Binary Completion Criteria

### Tracker revision definition of done

- [x] All 20 target files remain individually inventoried.
- [x] Every discovered observable contract has an owner, evidence posture,
  characterization requirement, and freeze rule.
- [x] TD-DEC-001 through TD-DEC-005 each trace to owner/planning authority,
  slice, acceptance evidence, and Make validation.
- [x] Portability, imported-owner, harness, package-edge, capability, default,
  failure, atomicity, and work-order mappings are explicit.
- [x] Slice dependencies place conformance correction and CG-01 before
  structural movement.
- [x] RB-001 through RB-004 are decision-complete and RB-005 is resolved by the
  complete authorized work order.
- [x] No owner contradiction is present and no supporting source is promoted
  above its authority.
- [x] Prior handoff history is preserved and a `2026-08-01` revision session is
  appended to every workstream table.
- [x] `make lint-markdown` passes after the final tracker edit.
- [x] No product code, test, owner contract, generated artifact, migration,
  frontend, package configuration, harness input, or analysis source changed.

### Conformance-correction definition of done

- [x] Core 01 contains the adopted exclusive Tasks/Decisions invariant mapping
  and deterministic failure precedence.
- [x] Each invariant has a valid, database-convertible target-specific semantic
  fixture, and multiple-defect input proves order-independent selection.
- [x] Every invalid portability candidate returns only the exact safe family and
  invariant identity and leaves no visible or partial authoritative state.
- [x] Imported Task and Decision owners require exact active same-incident
  membership at commit; explicit `null` fails and omission retains the actor
  default.
- [x] Membership-removal, authorization-loss, replay, and changed-content races
  satisfy the Section 4 outcomes without partial effects.
- [x] The store row no longer claims portability; eight portability tests, two
  rollback tests, and three application import tests select exactly once through
  `module.tasksdecisions`.
- [x] Owner unit and service-backed slices, harness contract, generation, drift,
  artifact-policy, and JSON-shape checks pass.
- [x] CG-01 passes and the handoff records retained evidence or exact failure
  classifications.

### Structural-refactor definition of done

- [x] CG-01 passed before the first structural edit.
- [x] All seven injected capabilities satisfy their input, output, failure,
  transaction, and exclusion contracts.
- [x] Private policy and source packages contain their adopted responsibilities
  without a prohibited import.
- [x] Every external caller imports only the root facade; application composition
  obtains leaf behavior through root contribution constructors.
- [x] No leaf imports the root, no external package imports target internals or
  leaves, compilation is acyclic, and the backend manifest enforces the graph.
- [x] Routes, envelopes, errors, view schemas, valid bundle output, defaults,
  transactions, revisions, events, replay, frontend behavior, and database
  schema match the corrected CG-01 baseline.
- [x] Owner slices, backend/frontend boundaries, focused and broader browser
  checks, generated drift, `make agent-finalize`, and `make check` pass.
- [x] Temporary compatibility wrappers are removed only after every caller moves,
  and the final handoff records commands, run roots, relatedness, and rollback
  state.

Tracker planning verdict: complete and executed.
Conformance-correction verdict: SL-00, SL-04, SL-05, and SL-06 done; CG-01
passed. Structural-refactor verdict: SL-01 through SL-03 and SL-07 done; the
overall handoff is closed. No public API, schema, or valid persisted-behavior
compatibility surface changed.
