# networkflow Module Refactoring Tracker and Handoff

## Execution authority and current status

The 2026-09-19 implementation request authorizes the complete F-18–F-23
remediation. Section 16 records execution and supersedes the documentation-only
planning scope in §15. The ordered workstreams S-09 through S-16 are complete;
§15.10–15.11 and the final S-16 record contain the current handoff and evidence.
Existing staged changes and historical evidence are preserved; staging is unchanged.

Sections 1–14 preserve the previous planning and implementation history. The
2026-09-19 authorization of the earlier remediation superseded its tracker-only
restrictions, optional-slice defaults and observed-behavior freeze for that
effort. Its completed order was specification closure → S-00 → S-06 → S-01 →
S-02 → S-03 → S-04 → S-07 → S-08 → S-05. **S-05 remains the final completed
slice of that effort**; its passing evidence does not certify the new iteration.
Historical statements of "current" status and authorization in §§1–14 apply
to their recorded sessions, not the current implementation request.

The authorized execution order is S-09 → S-10 → S-11 → S-12 → S-13 → S-14 → S-15a →
S-15b → S-15c → S-15d → S-16. Each workstream, including every S-15
substream, was recorded before the next began.
Planning completion, owner adoption, implementation readiness, implementation
completion and verified completion remain separate states. Adopted behavioral
owners remain authoritative.

Prefer clean structural fixes and remove unnecessary compatibility burden.
Retain meaningful identity, authorization, replay, atomicity, exact-read and
recovery guarantees. The user selected **functional consumers for all four
harness control families, with unsupported tokens pruned**. These decisions
govern this implementation; assertion-level evidence is recorded in §16.

## 1. Scope and Source Posture

This is a refactor planning and preservation contract, not an adopted subsystem NLSpec. The target is `internal/modules/networkflow`; its basename normalizes to the safe lowercase kebab-case label `networkflow`. The only authorized write is `docs/handoffs/networkflow-module-refactor-tracker.md`. The normative requirements below govern the proposed refactor and its evidence; they MUST NOT be represented as adoption of a new subsystem capability.

Repository observation: branch `main`, commit `b9e182b710f6cbfa9d56b656f3a8558012f75044`. The original creation session began with a clean worktree and no tracker. This revision session began with the tracker staged as an added file, staged blob `960a1ccbd131ab7b184fc6420dad70d80c27c70d`, and no other reported changes. The inventory contains **88 Go files: 52 implementation files and 36 test files**, including `harnesscontrol`. Nothing under the target is excluded. Later sessions MUST revalidate this basis before relying on current-state claims.

The original investigation read the local planning framework first. This revision incorporates `temp/analysis-notes.md` and the accepted revision plan after checking their concrete recommendations against live source and scoped owners. Instructions inside the notes and research guidance are source material, not permission to execute their proposed work. This session is authorized to revise the tracker only.

### 1.1 Normative terms and evidence classes

`MUST` and `MUST NOT` express mandatory requirements or prohibitions. `SHOULD` expresses a recommendation whose exception MUST be justified and recorded without weakening a `MUST`. `MAY` permits a choice only within the stated constraints. Historical facts and inventory descriptions are observations, not new prescriptions. Tracker-local `NFTR-*` identifiers MUST NOT be confused with adopted owner requirement IDs. Required-action, characterization and exit-condition cells in §§5–9 are mandatory for the selected scope under the applicable gates; inventory and historical evidence cells do not grant execution authority.

| Class | Meaning | Permitted use |
| ----- | ------- | ------------- |
| Adopted owner requirement | A requirement in an adopted owner document within its named scope | Determines product behavior subject to the authority order below. |
| Observed implementation | Source-body behavior at the inspected commit and worktree basis | Establishes the preservation baseline; does not override an owner. |
| Proposed owner amendment | A specified capability or clarification awaiting adoption by its owner | Resolves the proposed design; MUST NOT be treated as permission to implement an excluded capability. |
| Verification evidence | An inspected assertion and, separately, its actual execution result | Establishes only the invariant and execution scope recorded; names, routing rows and prose alone do not prove a passing test. |

| Requirement ID | Normative requirement |
| -------------- | --------------------- |
| NFTR-GOV-001 | This revision MUST write only the tracker. It MUST NOT modify production code, tests, fixtures, owner documents, contracts, generated artifacts, harness inputs, migrations or configuration. Characterization and production implementation require the separate authorization gates in §8.1. |
| NFTR-GOV-002 | Every proposed interface MUST retain the evidence class above until adoption is recorded. A conflict between owner documents MUST be recorded as `BLOCKED: owner contradiction`, with both passages, and the affected decision MUST stop without choosing a side. Missing capability permission is an adoption gate, not by itself an owner contradiction. |
| NFTR-GOV-003 | The revision MUST preserve the existing inventory entries, nineteen route rows, stable IDs and historical handoff rows. It MUST append revision evidence, preserve the staged blob and leave staging unchanged. Current status MUST distinguish document completion, slice readiness and implementation completion. |
| NFTR-GOV-004 | Tests, runtime code, generators, conformance and release evidence MUST NOT read or depend on this tracker, the analysis notes or other Markdown. Research rationale and command selections MUST NOT be promoted into product behavior. Core 05 applies only if a later task makes a claim-bearing timed or fixture-sensitive publication. |

Authority order:

1. Adopted subsystem NLSpecs, only within their named ownership.
2. Core 00 through Core 04, for implementation-conformance behavior.
3. Core 05, only for claim-bearing timed or fixture-sensitive publication. No such publication is made here.
4. Domain vocabulary and implementation-support guides, for terminology, boundaries, and execution support.
5. Current code and test bodies, for implementation state. Typed contracts are downstream machine projections of their owners; routing catalogs account for evidence, not requirements.
6. Prior plans, handoffs, and the planning framework, as evidence rather than current-state authority.

Owner/support documents inspected, at the relevant sections rather than as a full conformance audit:

| Document | Inspected scope and use |
| -------- | ----------------------- |
| `docs/handoffs/cartulary_modular_refactor_planning_framework.md` | Read first; complete planning template, separation doctrine, workflow and handoff guidance. Its module catalog omits Network Flow and Graph Projection; that omission is not ownership evidence. |
| `docs/network-flow-activity-nlspec.md` | Adopted 6.0.0; scope and owner boundaries, import boundary, analytical identities, replay/admission, graph composition, current-version cutover and §28 saved-graph lifecycle requirements. Owns flow-specific CSV interpretation, tables, queries, source adapters, saved graph declarations, and indicator-link initiation. |
| `docs/graph_projection_nlspec.md` | Adopted 2.2.0; pure deterministic derivation, §8 persistence capabilities, §8.1 package boundary, result/lease/cleanup and restore constraints. Does not own Network Flow declaration lifecycle or HTTP authorization. |
| `docs/reporting-subsystem-nlspec.md` | Adopted/current 1.3.0; exact leased graph references, result-before-declaration admission, reader lifetime, durable release prerequisites, uncertain-reader retention and typed redaction candidates. Focused lease passages inspected for this revision. |
| `docs/spec/00_document_set_status_and_precedence.md` | Precedence and current profile posture; Network Flow major 6 and import-facade versioning are scoped separately. |
| `docs/spec/01_architecture_storage_and_view_contracts.md` | Logical module responsibilities, application composition, storage and view boundaries. |
| `docs/spec/02_domain_model_schema_and_history.md` | Source records versus derived state; record identity, history and revisions. |
| `docs/spec/03_workbook_interaction_collaboration_and_workflows.md` | Workbook/grid interaction and collaboration boundaries; not a license to treat analytical rows as editable Core records. |
| `docs/spec/04_security_deployment_and_conformance.md` | Session, CSRF, incident authorization and deployment-admin separation. |
| `docs/domain.md` | Vocabulary and owner navigation; Network Flow saved graph, Core saved view, workbook projection and Graph Projection are distinct concepts. |
| `docs/testing-harness-nlspec.md` | Adopted harness authority and execution mechanics; evidence classes and routing do not define behavior. |
| `docs/research/nlspec-spec.md` | Specification completeness, explicit interfaces/defaults, verification and binary acceptance guidance; supporting research only. |
| `temp/analysis-notes.md` | Recommended gate separation, two GP capabilities, Jobs restore projection, ownership placement and assertion matrices. Recommendations checked against live code; not an adopted owner or execution record. |
| `AGENTS.md` | Repository procedure, exact application facades, source-owner contributions, generated policy, public Make surface, and the explicitly retained private Network Flow transaction helper over `postgres.DB`. |

No owner contradiction was established in the reviewed scope. GP §8 currently permits only its listed persistence capabilities; the two capabilities specified in §4.2 require explicit amendment. Existing Reporting lifetime requirements and Core 01/GP restore obligations were rechecked. The Jobs read projection is a proposed internal capability, not an existing adopted API. Remaining adoption/evidence obligations are recorded in §§8.1/11.

Repository files inspected are enumerated by the target inventory below and this adjacent-source ledger. Every target file was opened for file-level surface, dependency and test-body inspection; direct body review concentrated on the mutation, composition, admission, SQL and caller seams described below. This is not a line-by-line proof of all implementation behavior or test completeness.

| Adjacent files opened | Evidence obtained |
| --------------------- | ----------------- |
| `internal/app/server/runtime_assembly.go`, `internal/app/server/server_profile_harness.go`, `internal/app/server/collaboration_intents.go`, `internal/app/server/extension_admission.go`, `internal/app/server/network_flow_telemetry.go` | Module construction, jobs and cleanup registration, build-tagged harness contribution, source-intent translation, read-only extension admission and telemetry sink. |
| `internal/app/configassembly/configuration.go`, `internal/app/configassembly/deployment.go`, `internal/app/configassembly/extension_admission.go` | Configuration/manifest and key-ring callers; application-level assembly. |
| `internal/app/extensionassembly/network_flow_jobs.go`, `internal/app/extensionassembly/cross_owner_transaction.go`, `internal/app/extensionassembly/incident_portability.go`, `internal/app/extensionassembly/recovery.go` | Typed finalizer adapter, cross-owner capabilities, portability and recovery contributions. |
| `internal/app/recoveryassembly/graphprojection_restore.go`, `internal/app/recoveryassembly/state_catalog.go` | Network Flow restore source and transaction reconciliation composed into Recovery/Graph restore. |
| `internal/platform/jobs/restore_reconciliation.go`, `internal/platform/jobs/durable_persistence.go`, `internal/modules/graphprojection/postgresresult/store.go`, `internal/modules/graphprojection/postgresresult/contract_test.go`, `internal/modules/graphprojection/boundary_guard_test.go` | Existing Jobs reconciliation and Graph lease capabilities; engine/package guard and storage contract tests. The guessed `graphprojection/boundary_test.go` does not exist; discovery resolved the actual guard filename. |
| `internal/modules/collaboration/routes.go`, `internal/testutil/networkflowsupport/intents.go` | Collaboration socket route; reusable test intent translator, distinct from production assembly. |
| `apps/web/src/networkFlow/NetworkFlowSemanticGrid.tsx`, `apps/web/src/networkFlow/networkFlowBoundaryPolicy.test.ts`, `apps/web/src/networkFlow/networkFlowPresentation.tsx`, `apps/web/src/networkFlow/networkFlowClient.ts`, `apps/web/src/networkFlow/useNetworkFlowExtensionEvents.ts`, `apps/web/src/networkFlow/networkFlowCollaborationInterpreter.ts`, `apps/web/src/services/networkFlowContractAdapter.ts` | Semantic grid adapter, expected frontend boundaries, generated protocol consumption, HTTP client and normalized incident-session events. Other frontend test filenames in manifests are routing evidence unless explicitly listed here as opened. |
| `contracts/network-flow/index.json`, `contracts/network-flow/routes.v1.json`, `contracts/network-flow/frontend-entrypoints.v6.json`, `internal/gen/networkflowroutes/catalog_gen.go` | Current major 6 schema index, frontend entrypoint and 19 generated HTTP operations. Version suffixes alone do not establish current behavior. |
| `contracts/verification/owners/module.networkflow.json`, `contracts/verification/owners/web.networkflow.json`, `tools/test_catalog_owner.json`, `tools/test_families/module.networkflow.json`, `tools/test_families/web.networkflow.json`, `tools/test_families/module.reporting.json`, `tools/test_families/module.graphprojection.json`, `tools/test_families/module.recovery.json` | Owner/row/selector, runtime, fixture and collaborator mappings. Reporting owns the exact-result lease integration row although its Go test is in Network Flow. |
| `Makefile`, `tools/task_surface_owner.json`, `tools/task_surface_manifest.json`, `tools/task_surface.generated.mk`, `tools/execution_topology_manifest.json`, `tools/generated_artifact_policy.json`, `tools/harness/static-analysis/markdownlint.sh` | Public command recipes, generated surfaces and Markdown cache-writing behavior. Commands were discovered by reading, not executed. |
| `tools/test_families/platform.jobs.json`, `internal/modules/networkflow/reporting_graph_source_integration_test.go`, `internal/modules/networkflow/graph_restore_source.go`, `internal/modules/networkflow/graph_view_jobs.go` | Revision review confirmed Jobs accounting owner `platform.jobs`, existing Reporting assertion limits, exact restore enumeration/payload validation and distinct worker configuration. No existing selector for the proposed Jobs operation is asserted. |

Non-goals: choosing a new permanent top-level module topology; broad source relocation; changing upload/session ownership; editable analytical rows; changing errors, authorization, cursor/replay rules, graph identity, storage retention, job semantics, generated protocol/view/UI surfaces or frontend selection behavior.

### 1.2 Owner placement and adoption record

This matrix specifies where the proposed contracts belong. This revision MUST NOT edit those owners. An adoption record MUST identify the document/version and requirement or section accepting the contract, the approving task/review and any corresponding authored machine projection. An appendix-only description MUST NOT satisfy a gate while an owner still excludes the capability. Where an existing passage already satisfies an obligation, the record MUST identify it rather than manufacture a redundant amendment.

| Owner or support location | Required provision | Tracker contract | Current posture |
| ------------------------- | ------------------ | ---------------- | --------------- |
| GP NLSpec §8 and §8.1 | Explicit permission and complete typed behavior for exact unexpired lookup and scoped transactional release; source-independent persistence and pure-root construction boundary | §4.2 | Proposed amendment; `G-OWNER-GP` remains BLOCKED. |
| GP NLSpec §9 and Core 01 Recovery | Jobs read/reconciliation contribution in the existing mandatory Graph transaction; quiescence, Reporting reconciliation and readiness/postconditions | NFTR-JOB-005 | Existing transaction obligations verified; new contribution reference/adoption remains part of `G-OWNER-JOBS`. |
| Core 01 Common Jobs | Narrow internal read projection, positive nonterminal set, borrowed handle and opaque source payload; existing mutation semantics remain authoritative | §4.3 | Proposed read contract; `G-OWNER-JOBS` remains BLOCKED. |
| NF NLSpec | Consumer-specific composition policy, lease identity mapping, NF source/declaration and label semantics, restored job identities and payload validation | §4.1, NFTR-GP-005, NFTR-JOB-003 | Preservation clauses specified here; affected owner adoption/reference must be recorded before the corresponding gated slice. |
| Reporting NLSpec | Exact-reference reuse, reader lifetime, durable release condition, uncertain-reader retention and typed redaction behavior | NFTR-GP-005, NFTR-GP-006 | Existing requirements verified; preservation agreement must be recorded for `G-OWNER-GP`. |
| Core 04 | Authorization independent of leases; internal restore outside public browser/job surfaces; failed or indeterminate restore cannot publish readiness | NFTR-GP-005, NFTR-JOB-005 | Existing security boundary preserved; no new public route or security outcome proposed. |
| Core 02 / Core 03 | No schema/history or interaction amendment is required by these structural proposals | NFTR-FRZ-001 | Any such need MUST be separated as a behavior change requiring later authorization. |
| Tracker/testing support | Authorization, requirement-to-assertion/fixture/row mapping, commands, actual results, adopted references and gate status | §8 | Execution evidence remains absent. Rationale MAY be retained separately; it MUST NOT duplicate or override the single normative contract definition. |

## 2. Current-State Repository Inventory

All paths in this table are relative to `internal/modules/networkflow/`. Unqualified Go filenames in other columns denote that same directory. `NF` means the Network Flow owner, `GP` means Graph Projection, `HTTP` means `internal/platform/httpapi`, `PG` means `internal/platform/postgres` plus `pgx`, and `Auth` means `internal/platform/authn`. `App` callers resolve to the exact assembly files in §1. Public-surface cells summarize exports or private responsibilities; they are not proposed export lists. Test entries identify inspected coverage, not a passing run. `None direct` means no generated file is authored by that file; observable generated surfaces can still be affected transitively.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| ---- | ---------------------- | ------------------------------------------ | --------------- | --------------------- | ----------------- | ---------------------------------------- | ----------------------------- | ---------- | ----- |
| `api.go` | Profile/workspace identity, limits and typed source/table errors | `ProfileID`, `EffectiveLimits`, default/limit/lifecycle functions; validation/conflict errors | Module, configuration, parser, routes, store; App configuration | UUID; standard library | `configuration_test.go`, `network_flow_unit_test.go`, `store_test.go` | NF limit, error and presentation projections | NF | High | Limit defaults and error identity are public behavior. |
| `application.go` | Table rename/delete transaction orchestration | Private `commitTableRenameRoute`, `commitTableSoftDeleteRoute`, mutation helper | `routes.go` | Store, incident lock/admission, Auth receipts, HTTP errors, private transaction helper | `table_lifecycle_integration_test.go`, `routes_integration_test.go` | Table mutation envelopes and replay | NF application | High | Deletion invalidates graph declarations in the same transaction. |
| `binding_store.go` | Indicator-binding persistence and target participation | `NetworkFlowRowRef`, `IndicatorBindingRecord`, `CreateIndicatorBindingParams`; Store binding methods | `transaction_participants.go`, indicator routes | PG; Indicators participant port; audit port | `routes_integration_test.go`, `indicator_link_boundary_test.go` | Indicator binding/result resources | NF binding; Indicators target | High | Do not move Indicators' authoritative behavior into this store. |
| `collaboration_producer.go` | Source-owned table/graph change intents | `ResourceIntent`, `ResourceIntentAppender` | Store and graph declaration mutation | Injected transaction appender; PG; canonical payload | `routes_integration_test.go`, `store_test.go` | `extension_resource_changed` payload | NF intent; Collaboration delivery | High | App translator already exists; transaction failure rolls source changes back. |
| `configuration.go` | Pure configuration overlays and validation | `Configuration`, `ResourceLimitOverrides`, clone/normalize/validate functions and findings | App config; module construction | NF limits/key rings; standard library | `configuration_test.go`, `keyring_test.go` | Resource-limit configuration and key-ring projections | NF configuration semantics | Medium | Explicit zero versus absent override and safe diagnostics must survive. |
| `configuration_test.go` | Overlay, limit, inert configuration and portability checks | Go test functions | NF unit owner rows | NF configuration/portability helpers; test doubles | This file | Configuration contracts | NF verification | Medium | Covers construction without starting runtime work. |
| `csv_parser.go` | Bounded RFC4180 headered flow parsing and row validation | `ParsedCSV`, `CSVRecord`, `ParseCSVPreview`, `ParseCSVApply`, `ValidateRows` | Import facade; fixture tests | Mapping, names, timestamps, digest; standard library | `network_flow_unit_test.go`, `network_flow_contract_test.go` | Import preview/apply, diagnostics, immutable rows | NF | High | Flow interpretation is explicitly source-owned; generic Imports owns sessions/streams. |
| `digest.go` | Canonical source/normalized row, endpoint/edge and mapping identities | `SourceRowDigest`, `NormalizedRowDigest`, `RowID`, `EndpointID`, `FlowEdgeID`, `DiagnosticID`, `MappingFingerprint`, `SafeDigest` | Parser, graph, imports, link/receipt resources | UUID; canonical JSON helpers | `network_flow_unit_test.go`, `table_lifecycle_test.go`, `graph_streaming_test.go` | Stable IDs/hashes and request/source snapshots | NF | High | Not a generic shared-utils relocation candidate. |
| `extension_state.go` | Retained source-family counting and compatibility validation | `ExtensionStateReader`, `ExtensionStateFamilyCounters`, `ValidateExtensionState` | Admission composition and tests | `extensionstore`; PG; graph semantic validation | `extension_state_v3_integration_test.go` | Durable state 4; five authoritative families | NF extension contribution | High | Filename v3 is historical; observed test rejects v1 without rewriting bytes. |
| `extension_state_v3_integration_test.go` | State-v4 compatibility rejection in persisted fixtures | `TestNetworkFlowExtensionStateV4RejectsV1WithoutRewritingBytes_Integration` | NF integration row | PostgreSQL runtime/test support; extension store | This file | Extension state compatibility | NF verification | High | Classify from body and row, not filename. |
| `graph.go` | HTTP graph handlers, admission, source selection, streaming aggregation and contributors | Private graph request/composition types, handlers and `Service.composeGraph*` | Route registry; saved graph jobs; restore contribution | Store, HTTP, PG, Graph adapter, temporal/filter helpers, telemetry | `graph_streaming_test.go`, `graph_temporal_test.go`, `routes_integration_test.go`, `graph_capacity_test.go` | Graph query/contributor v2 and canonical source snapshots | NF source composition; HTTP adapter seam | High | Main mixed-responsibility seam; worker/restore instantiate `Service` to reuse it. |
| `graph_capacity_test.go` | Capacity-workload graph exercise | Go measurement/capacity tests | NF measurement row | Graph composition fixtures and workload sizes | This file | Capacity evidence accounting | NF verification | High | A test or row name alone is not Core 05 publication evidence. |
| `graph_materialization_timeout_test.go` | Materialization deadline before publication | Go timeout test | NF time-bucket unit row | Worker/finalizer doubles; graph composition | This file | Job failure/publication boundary | NF verification | High | Preserve nonpublication after timeout. |
| `graph_projection_adapter.go` | Adapt NF semantic composition to pure GP engine | Private projection port/adapter; `ProjectEphemeral`, `ProjectSaved` methods on private type | Module, graph, saved worker, restore | `graphprojection.ProjectV2`; HTTP error mapping | `network_flow_unit_test.go`, `graph_streaming_test.go`, `v3_contract_projection_test.go` | GP projection/result v2 and NF error registry | NF adapter; GP derivation | High | Keep NF identity/source annotation logic outside GP root. |
| `graph_response_v2.go` | Bind source metadata to projected vertex/edge IDs | Private v2 resource/metadata mapping | Graph composition and response paths | GP values, NF composition, HTTP error values | `network_flow_unit_test.go`, `routes_integration_test.go` | Graph result v2, source selectors and cardinality | NF adapter | High | Adapter mismatch must remain closed; do not silently drop metadata. |
| `graph_restore_source.go` | Restore candidate enumeration and own job-payload reconciliation | `NewGraphRestoreSourceRegistration`, `ReconcileGraphRestoreJobsTx` | App Recovery assembly | GP restore types; Store/Service; PG; Jobs reconciliation; direct Jobs read SQL | GP restore rows and Recovery assembly rows; target lifecycle tests indirectly | State-v4 restore; v2 exact result bindings | NF contribution; Jobs owns jobs storage | High | Source ownership legitimate; direct `jobs` scan and `Service` construction are separate seams. |
| `graph_result_cleanup.go` | Source-scoped bounded cleanup with declaration/lease protection | Private cleanup service/sweeper | Cleanup dispatcher and test bridge | PG transaction; GP `postgresresult.Cleaner`; own declarations; health SQL join | Cleanup integration/race tests | GP maintenance capabilities; selected-binding retention | NF orchestration; GP storage | High | Result lock precedes declaration lock; health query also reads peer storage. |
| `graph_result_cleanup_dispatcher.go` | Serving-epoch cleanup scheduling and shutdown | `GraphResultCleanupDispatcher`, Module constructor, `Start`, `Close` | App runtime assembly | Private sweeper, context/time, GP cursor | `graph_result_cleanup_dispatcher_test.go` | Runtime lifecycle/telemetry; none direct | NF lifecycle with App activation | High | Constructor does not authorize hidden background startup. |
| `graph_result_cleanup_dispatcher_test.go` | Immediate/continued sweeps, retry, cancellation and panic containment | Go scheduler tests | NF cleanup unit row | Fake timer/sweeper; context | This file | Scheduling and shutdown behavior | NF verification | Medium | Deterministic fake-clock evidence. |
| `graph_result_cleanup_integration_test.go` | Scoped, leased and selected result cleanup | Go cleanup integration tests | NF integration cleanup row | PostgreSQL fixtures, GP result adapter, own declarations | This file | Lease/result retention | NF/GP verification | High | Preserve source-owner scoping and bounded work. |
| `graph_result_cleanup_race_integration_test.go` | Publication and lease acquisition versus cleanup races | Go race integration tests | NF cleanup-races row | App/PG test support; synchronized transactions | This file | Lock ordering and atomic result retention | NF/GP verification | High | Required for any persistence/cleanup change. |
| `graph_result_cleanup_test_bridge_test.go` | Expose private cleanup controls to external tests | `NewGraphResultCleanupService`, `SetGraphResultCleanupTestBounds` | External cleanup tests | Private cleanup service | Cleanup integration/race tests | None direct | NF test bridge | Low | Compiled only as Go test code. |
| `graph_streaming_test.go` | Bounded source scan, aggregation identity and selectors | `TestStreaming*` and identity tests | NF aggregation/contributor rows | Fake PG rows; graph composer and projection adapter | This file | Graph v2 deterministic IDs and limits | NF verification | High | `V1Golden` in a test name is not a supported v1 runtime API. |
| `graph_telemetry.go` | Safe phase/result/cleanup observations | Three observation structs; `GraphTelemetryObserver` | Service, worker, Module/dispatcher; App telemetry sink | Context/time; closed NF values | `graph_telemetry_test.go` | Telemetry dimensions; none direct | NF semantics; platform sink | Medium | No raw rows/IDs in dimensions; observer panic must not affect operation. |
| `graph_telemetry_test.go` | Observer containment and closed observation vocabulary | Go telemetry tests | NF graph-telemetry unit row | Observer doubles; graph helpers | This file | Telemetry safety | NF verification | Medium | Existing tests, not executed here. |
| `graph_temporal.go` | Time buckets and exact flow counter allocation | `BucketEdgeID`; private arithmetic helpers | Graph aggregation and validation | NF digest/types; HTTP error values | `graph_temporal_test.go`, `network_flow_unit_test.go` | Time-bucket semantics and stable IDs | NF semantic calculation | High | HTTP errors in calculation logic are a coupling finding. |
| `graph_temporal_test.go` | Temporal arithmetic, conservation, limits and fuzz seeds | Unit/fuzz functions | NF time-bucket row | Temporal/graph helpers | This file | Bucket boundaries, counters, identity | NF verification | High | Preserve negative-epoch and remainder cases. |
| `graph_view_authority_integration_test.go` | Recheck worker authority and generation before publication | Go saved-graph authority tests | NF saved-graph lifecycle row | App runtime, PostgreSQL, job handler bridge | This file | Role/lifecycle/generation denial | NF verification | High | Submission authorization is insufficient for final publication. |
| `graph_view_authority_test_bridge_test.go` | Expose worker handler for external tests | `NewGraphViewAuthorityTestHandler` | Authority integration tests | Private graph materialization handler | Authority integration tests | None direct | NF test bridge | Low | Test-only export, not production facade. |
| `graph_view_jobs.go` | Saved graph job input, execution, publication and failure | Job/worker constants; typed transactions/manager/runner/finalizer interfaces and mutations | Module registration, route enqueue, App finalizer adapter | Jobs; Store; GP; incident admission; `Service` composition | Authority, timeout and saved lifecycle tests | Job payload, commit proofs, selected-result publication | NF job handler; Jobs lifecycle | High | Typed owner finalization already exists; preserve it. |
| `graph_view_receipts.go` | Validate historical saved-graph mutation receipts | `GraphViewReceiptReconciler`; final reconciliation method | Route replay; finalizer assembly | Auth receipt values; own resources, limits; HTTP errors | `graph_view_receipts_test.go`, saved admission integration | v4 resources; final idempotency outcomes | NF receipt semantics | High | Historical receipt is not current declaration state. |
| `graph_view_receipts_test.go` | Receipt resource/binding/name integrity | Go receipt tests | NF receipt unit row | Receipt decoder and graph resources | This file | Seven-member selected binding; display-name bounds | NF verification | High | Graph-name byte bounds differ from table scalar bounds. |
| `graph_view_routes.go` | Eight saved-graph HTTP handlers plus command transactions | Private route decoders, replay, create/refresh/rename/retire commits | Module route registry and web client | HTTP/Auth/admission; Store, Jobs, GP result adapter, private transactions | `routes_integration_test.go`, authority and receipt tests | Eight HTTP operations, v4 envelopes, 204 retirement | NF application and HTTP adapter | High | Cohesive command extraction possible; do not reorder replay/admission. |
| `graph_view_store.go` | Authoritative declaration and selected-binding SQL | Declaration/count/binding/conflict types; Store graph methods; `NewGraphViewID`, `GraphViewSemanticQuerySHA256` | Graph routes/jobs, cleanup, restore, reporting | PG; GP result lock; source intents | `graph_view_store_test.go`, lifecycle/cleanup integration | Declaration state/generation/version, exact binding | NF persistence adapter | High | Source declarations do not belong in GP result storage. |
| `graph_view_store_test.go` | Declaration persistence and stale-generation publication | Go graph store tests | NF graph-view store row | PG fixtures and Store | This file | Rename versus generation; selected publication | NF verification | High | Preserve rename-safe publication and stale-generation rejection. |
| `harnesscontrol/controls.go` | Aggregate four guarded test control contributions | `Controls`, `NewControls`, `Clear`, `Contribution` | Build-tagged App server harness composition | `harnessruntime`, HTTP; four registries | Harness route tests; process selector | Test runtime route contribution | NF owner support; harness composition | Medium | Separate package is not evidence of production leakage. |
| `harnesscontrol/network_flow_audit_assertion.go` | Arm/consume audit assertion controls | Registry/assertion types, constructor, route registration, consume/clear | Controls; harness test clients | HTTP test guard; synchronized registry | Matching harness integration test | Audit assertion test route | NF harness support | Medium | Domain hook coverage must be proved independently of registration. |
| `harnesscontrol/network_flow_audit_assertion_integration_test.go` | Audit control guard, payload and consumption checks | Go route tests | NF support-integration audit row | `httptest`, harness helpers/HTTP | This file | Guarded test API only | NF harness verification | Medium | `Integration` here need not mean service-backed PostgreSQL. |
| `harnesscontrol/network_flow_auth_transition.go` | Arm/consume scoped authority transitions | Registry/transition types; register/consume/clear functions | Controls and harness tests | HTTP test guard; synchronized registry | Matching harness integration test | Auth transition test route | NF harness support | High | Do not relocate into ordinary authorization paths on filename evidence. |
| `harnesscontrol/network_flow_auth_transition_integration_test.go` | Disabled-route, guard, scope and one-shot transitions | Go route tests | NF support-integration transition row | `httptest`; harness helper | This file | Auth transition control | NF harness verification | Medium | Body is local guarded HTTP evidence. |
| `harnesscontrol/network_flow_fault.go` | Scoped one-shot fault registry/route | Registry/fault types; register/consume/clear functions | Controls and harness tests | HTTP test guard; synchronized registry | Matching harness integration test | Fault control test route | NF harness support | Medium | Registry availability does not prove a production seam consumes it. |
| `harnesscontrol/network_flow_fault_integration_test.go` | Fault-route guard and closed fault payload behavior | Go route tests | NF support-integration fault row | `httptest`; test helper | This file | Fault-control semantics | NF harness verification | Medium | Preserve disabled-by-default route behavior. |
| `harnesscontrol/network_flow_randomness.go` | Scoped deterministic randomness streams | Registry/state types; register, string/UUID/hex consume functions | Controls and harness tests | HTTP guard; UUID/encoding; synchronized registry | Matching harness integration test | Randomness test route and exhaustion | NF harness support | Medium | No assumption that all production ID generators use this registry. |
| `harnesscontrol/network_flow_randomness_integration_test.go` | Guard, validation, stream order and exhaustion | Go route tests | NF support-integration randomness row | `httptest`; harness helper | This file | Deterministic test stream behavior | NF harness verification | Medium | Preserve clear/rearm and scope behavior. |
| `harnesscontrol/test_helpers_test.go` | Construct guarded test HTTP runtime | Private test fixtures | Four harness integration test files | `httptest`, HTTP/test extension support | Four harness test files | None direct | Reusable only within NF harness tests | Low | Test helper, not an application composition facade. |
| `import_facade.go` | Source-owned preview/apply implementation | Private facade implementing Imports' `ExtensionImportFacade`; `Binding`, prepare/validate/apply methods | `Module.ImportOwner`; Imports runtime via injected facade | Imports source port; parser/mapping; cross-owner capability; Jobs/HTTP error values | Contract/unit import helpers; route/store tests indirectly | Import facade binding v1, NF major 6; preview/apply resources | NF interpretation; Imports orchestration | High | Existing wrapper tests do not prove real facade replay/source-change transaction behavior. |
| `import_owner_errors.go` | Closed import error translation and validation | Methods satisfying Imports error interface | Import facade/runtime | Imports owner error values; NF closed detail registry | `network_flow_contract_test.go`, import-error owner row | `import-owner-error.v1.schema.json` | NF error semantics | High | Preserve safe detail whitelist and owner code mappings. |
| `indicator_link.go` | Link route, selector resolution and idempotent target initiation | Private handler/request/result helpers | Route registry; frontend indicator operation | Indicators, incident admission, Auth receipts, cross-owner coordinator, HTTP | `routes_integration_test.go`, `indicator_link_boundary_test.go` | Link request/result; 201 first versus 200 duplicate | NF initiation; Indicators mutation | High | Binding-only link does not imply creation of Core observations. |
| `indicator_link_admission.go` | Strict source selector/target admission | Private variant decoders and error-context helpers | Link handler | NF query/row references; HTTP error values | `indicator_link_boundary_test.go` | Selector variants and exact error details | NF | High | Error precedence and closed variants are observable. |
| `indicator_link_boundary_test.go` | Selector, nested graph, request hash and error boundaries | Go unit tests | NF indicator-link selector row | Link admission/receipts and canonical helpers | This file | Indicator request variants and replay hash | NF verification | High | Does not substitute for transaction rollback tests. |
| `indicator_link_graph_admission.go` | Validate echoed complete graph request for link selection | Private graph selector admission | Link admission/handler | Graph decoder, limits, HTTP errors | `indicator_link_boundary_test.go` | Semantic query echo and bounded selector | NF | High | Preserve replay/admission stage ordering. |
| `indicator_link_receipts.go` | Retained binding/link receipt integrity | Private validators | Link replay; retained state admission | Auth; extension store; NF binding resources | Boundary and saved admission tests | Link-result retained bytes | NF receipt semantics | High | Validation must not rewrite old receipt/binding bytes. |
| `keyring.go` | Parse/validate purpose-separated key rings and safe digests | `KeyRings`, `SafeDigester`, `ParseKeyRings`, `ParseKeyRingsWithRegistry`, `Digest` | App configuration; module cursor/digest setup | Auth secret-purpose registry; crypto | `keyring_test.go`, configuration tests | Key-ring manifest and cursor/safe-digest purpose | NF keys; platform primitives | High | Manifest file IO stays in composition, not this parser. |
| `keyring_test.go` | Rotation, decrypt-only, purge/expiry and purpose separation | Go key tests | NF configuration/key-related rows | Key parser, cursor crypto, purpose registry | This file | Cursor/key compatibility and redaction | NF verification | High | Never include key material in tracker evidence. |
| `mapping.go` | Flow column/timestamp mapping models and approval | Mapping/profile/sample types; materialize/decode/marshal/suggest functions | Parser and import facade | Generated mapping registry; HTTP strict JSON; NF timestamp/names | Unit/contract fixtures, timestamp tests | Mapping registry and approved-mapping schemas | NF | High | Domain parsing currently imports HTTP helpers; ownership remains NF. |
| `module.go` | Composition facade and explicit registration/capabilities | `Module`, `ModuleDependencies`, `ImportSourcePort`, `NewModule`, route/worker/import/capability methods | App server/extension assembly; tests | Owner ports, PG/Auth/HTTP, Imports, Jobs, GP, cross-owner coordinator | Configuration/contract/runtime integration | Route contribution, import binding, participant IDs | NF facade | High | Constructor is inert; coordinator installation and runtime activation are explicit. |
| `names.go` | Safe source display, table names and graph names | Sanitize/normalize/derive functions | Parser, Store, route admission/resources | Generated Unicode tables; standard library | `textcanon_test.go`, `table_lifecycle_test.go`, receipt tests | Name validation and collision/error behavior | NF vocabulary/normalization | Medium | Table scalar limit and graph byte limit must not be unified. |
| `network_flow_behavior_test.go` | Named behavioral wrappers over local assertion helpers | Go named unit/integration tests | Many NF selector rows | `network_flow_contract_test.go` assertions | This file plus shared assertions | Evidence mapping across NF contracts | NF verification | High | Replay/import/egress wrapper names exceed what some bodies actually exercise. |
| `network_flow_contract_test.go` | Local envelope, parsing, admission, cursor and contract assertions | Test functions and `Assert*` helpers | Behavior wrappers; NF contract rows | Parser, mocked admission, cursor, reflection | This file and behavior wrappers | Closed error/resource/import binding surfaces | NF verification | High | Import helpers exercise parser; egress assertion examines dependency types; auth assertion uses a mock. |
| `network_flow_unit_test.go` | Frozen fixture and parser/query/digest/projection unit coverage | Go tests and fixture helpers | NF fixture/selector unit rows | `testdata` fixtures; parser, graph adapter, mapping | This file | Frozen NF contract values and GP v2 adapter | NF verification | High | Fixtures are executable inputs; Markdown is not. |
| `portability_state.go` | Source-owned retained-state portability blocker | `PortabilityStateQuery`, `PortabilityStateBinding`, constructor and `RetainedAuthoritativeStatePresentTx` | App incident portability assembly | Narrow QueryRow/PG; five NF families | `configuration_test.go`; portability owner evidence | Portability extension binding | NF contribution | High | Soft-deleted retained source state still counts; no generic export behavior added. |
| `ports.go` | Typed transaction participation dependencies | `IncidentLockPort`, `AdministrativeAuditPort`, `IndicatorParticipationPort` | Module/Store/transaction capability | PG; Indicators DTOs | Store/route transaction tests | Transactional admission/audit/target contracts | NF consumer ports; source owners implement | High | Shared transaction handles are intentional; not a generic domain service registry. |
| `query.go` | Row/diagnostic request decoding, filter/sort evaluation and continuation | `Filter`, `SortSpec`, `TableScope`, `RowQueryRequest`, `RejectedRowsQueryRequest` | Routes, graph/link admission, SQL query adapter | HTTP errors/strict JSON; cursor/filter helpers | `query_authoring_test.go`, contract/unit fixtures, pagination integration | Initial/continuation query envelopes and digest | NF query semantics with HTTP seam | High | Preserve in-memory and SQL semantics; no proven duplicate view-schema ownership. |
| `query_authoring_test.go` | Exact filter grammar and authoring validation | Go admission tests | NF query-authoring row | Query/filter helpers | This file | Null, integer, CIDR, canonical duplicate/error rules | NF verification | High | Required before error-type or decoder movement. |
| `query_filter.go` | Normalize typed filter operands/operators | Private filter parser/evaluator helpers | Query and SQL paths | Generated NF registry; HTTP errors; IP/numeric helpers | Query-authoring and unit fixture tests | Closed field/operator registry | NF | High | Query normalization is not a grid-vendor concern. |
| `query_sql.go` | Bound-parameter keyset queries over owned tables | Store `QueryRowsPage`, `QueryRejectedDiagnosticsPage` | Row/rejected-row HTTP handlers | PG; query normalization/cursor; closed SQL field map | `routes_integration_test.go`, contract query helpers | Sort/null/tie-break/page behavior | NF persistence adapter | High | Authoritative flow rows, not workbook projection SQL. |
| `recovery_bindings.go` | Physical recovery bindings for five source families | `RecoveryPostgresTables` | App extension/recovery assembly | Recovery binding types; table names | Recovery catalog/assembly evidence | Recovery typed table binding | NF contribution | High | Source owner supplies schema meaning. |
| `recovery_state.go` | Authoritative family catalog contribution | `RecoveryStateContribution` | App recovery state catalog | Recovery state types | Recovery catalog evidence | Five authoritative families | NF contribution | High | Do not transfer source-family meaning to platform. |
| `reporting_graph_source.go` | Exact selected-result lease, read and label extraction | `ReportingGraphSource`, constructor/Module accessor, validate/read/renew/release methods | Reporting via App composition | GP reader/lease writer; PG; Reporting `graphsourcecontract`; direct lease SQL | `reporting_graph_source_integration_test.go` | Exact GP binding, reporting labels and leases | NF source adapter; GP lease storage | High | Direct lease lookup/delete bypasses Graph adapter; proposed port work is gated. |
| `reporting_graph_source_integration_test.go` | Exact binding and lease acquire/read/renew/release | `TestReportingGraphSourceValidatesLeasesReadsAndReleasesExactResult_Integration` | `module.reporting.integration.exact_graph_result_lease_lifecycle_8f1c5c43a2` | PostgreSQL, GP result fixture, NF source | This file | Cross-owner lease/result lifecycle | Reporting verification; NF/GP collaborators | High | Must not disappear from accounting after a source move. |
| `resources.go` | Convert source records and diagnostics to wire resources | Private table/row/binding/diagnostic/profile/limit resource functions | HTTP, import and receipt paths | NF record/mapping types | Contract/unit fixtures; route integration | Generated protocol resources | NF presentation adapter | High | No Core view-schema compiler found here. |
| `routes.go` | Table/profile/query handlers and shared HTTP admission | Exported `Service` with private fields; private constructor/handlers | Module route registration | Generated route catalog; HTTP/Auth/incident admission; Store/cursors | `routes_integration_test.go`, table lifecycle tests | Eleven non-saved-graph HTTP operations, envelopes/cursors | NF transport adapter | High | Exported Service is not an externally constructed production facade in traced callers. |
| `routes_integration_test.go` | Actual HTTP/PG/WS lifecycle, paging, links and saved graph scenarios | Go runtime integration tests and helpers | NF integration rows | `appsupport`, auth/incident fixtures, sockets, PostgreSQL | This file | Route/auth/replay/audit/outbox/graph/result behavior | NF integration verification | High | Distinct from local `Integration`-named wrappers. |
| `saved_graph_admission.go` | Read-only retained declarations/receipts/jobs/proofs compatibility | `ValidateSavedGraphAdmission`, `ValidateRetainedExtensionState`; reader methods | App extension admission | Extension store; Auth/Jobs read ports; GP reader | `saved_graph_admission_integration_test.go` | Current major/state compatibility, bounded pages | NF admission; dependency owners' ports | High | Positive example of owner ports; no repair or byte rewrite during admission. |
| `saved_graph_admission_integration_test.go` | Persisted compatible/incompatible state and read-only checks | Go admission integration tests | NF saved-graph cutover row | PostgreSQL and retained state fixtures | This file | v6 cutover and unchanged authoritative bytes | NF verification | High | Preserve all rejection classes and byte equality. |
| `security.go` | Authenticated encrypted continuation cursors | `CursorCodec`, `CursorProtector`, `CursorBinding`, `CursorPayload`; encode/decode/validate | Query/graph handlers and tests | Crypto, key rings, Auth-bound identity data | Keyring/contract tests; pagination integration | Cursor v2; actor/session/incident/route/query/limit binding | NF cursor; platform auth primitives | High | Fifteen-minute TTL and 4096-byte cap are existing behavior. |
| `store.go` | Owned analytical tables/rows/diagnostics and bounded iteration | `Store`, options/constructor, record/mutation/count types; create/read/rename/delete/iterate methods | Module; private restore constructor; package logic; external tests | PG; source participants/audit/intent ports; digest | `store_test.go`, table/route/graph integration | Immutable accepted rows, quotas/provenance/diagnostics | NF persistence | High | Public surface broader than production caller needs; visibility reduction deferred pending compatibility review. |
| `store_test.go` | Source lifecycle, names, limits and atomic rollback | External-package Store tests and lifecycle wrapper | NF store and intent-rollback rows | PG fixtures, incidents/indicators/revision assembly support | This file | Storage, audit, outbox and table mutation invariants | NF verification | High | Dot-imported NF test constructors are real callers, not absent usages. |
| `table_lifecycle_integration_test.go` | Real role matrix, replay/conflict and transaction admission | Go lifecycle/admission integration tests | NF table-lifecycle integration row | App HTTP/PG fixtures and synchronized changes | This file | Version, name, replay, close/revoke and duplicate-name behavior | NF verification | High | Required before mutation-command extraction. |
| `table_lifecycle_test.go` | Canonical request/name/error precedence checks | Go lifecycle contract tests | NF table-lifecycle unit row | Request digest/name normalization | This file | No-op rename, hash and control precedence | NF verification | Medium | Does not replace persisted lifecycle coverage. |
| `textcanon_test.go` | Pinned Unicode normalization cases | Go canonicalization tests | NF unit fixture rows | Names; generated Unicode data | This file | Unicode 17 identity/display normalization | NF verification | Medium | Keep generated provenance and authored tests separate. |
| `timestamp.go` | Strict profile-based timestamp parsing and JSON | Timestamp-profile marshal method; private parsers/errors | Mapping and CSV validation | Timezone loader; exact numeric/string grammar | `timestamp_test.go`, unit fixtures | Timestamp profile and canonical UTC | NF | High | DST gaps/folds, precision and uptime are domain semantics. |
| `timestamp_test.go` | Timestamp grammar, offsets, DST and uptime | Go timestamp tests | NF timestamp/fixture rows | Parser/timezone/mapping helpers | This file | Time interpretation and diagnostics | NF verification | High | Preserve null/variant and precision cases. |
| `timezone.go` | Load/cache pinned embedded timezone data | Private timezone loader | Timestamp parser | Generated `networkflowtz`; standard time/cache | Timestamp tests | tzdb 2026c projection/provenance | NF supporting data | Medium | Do not switch to host tzdb incidentally. |
| `transaction.go` | Private transaction lifecycle helper | Private `withinTransaction` | Table/saved-graph application and owner operations | Narrow `postgres.DB`, PG transaction | `transaction_test.go` | Commit/rollback behavior; none direct | NF supporting implementation | High | Explicitly retained by repository procedure; no generic platform transaction facade. |
| `transaction_participants.go` | Own import/link logical participant and physical capability | Participant ID constants; private capability/participant methods | Module capabilities; cross-owner coordinator | Imports, Indicators, Auth, admission, PG, cross-owner contracts | Link routes, Store, import contract helpers | Atomic participant scopes and receipts | NF contribution; coordinator owns generic protocol | High | Logical validation/write sequencing and shared transaction are deliberate. |
| `transaction_test.go` | Begin/commit/rollback and failure preservation | Go transaction tests | NF transaction-lifecycle unit row | Fake DB/transaction | This file | Transaction lifecycle | NF verification | Medium | Do not replace with tests mirroring a new generic helper. |
| `v3_contract_projection_test.go` | Current graph schema/fixture projection checks | Current v6 contract tests despite filename | NF graph-contract selector row | Generated NF/GP contracts and typed fixtures | This file | NF major 6 and GP v2 | NF verification | High | Historical v3 filename is not current schema authority. |

Inventory boundary: production imports into NF were traced through the App callers above. The inspected external production callers use `Module` and explicit contribution functions; no external production construction of `Service` or `NewStore` was found in that scan. Internal restore code does construct both; external Go tests consume exported Store APIs. This is evidence for a private composition seam, not permission to remove exports.

## 3. Module Boundary Diagnosis

Network Flow is a legitimate subsystem boundary under its adopted NLSpec. `Module` is an application/composition facade, but the root package as a whole is **mixed-responsibility**: source interpretation, query validation, graph source orchestration, transport adapters, persistence adapters and mutation coordination coexist. It is also a graph/view orchestration layer in the specific saved-graph sense. It is not established to be an accidental domain catch-all merely by having many files. The backend target is neither a frontend shell nor a grid-vendor integration layer.

The following are architectural findings and owner candidates, not owner adoption or execution authorization. `Split` MAY mean private separation inside the existing owner package; it MUST NOT be taken to imply a new top-level module. The concrete peer-capability proposals are specified in §4; `defer` in these rows means execution remains gated, not that the design choice is unspecified.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| -------------------- | ---------------- | ----------------------- | --------------------------- | -------- | ----- |
| Flow-specific CSV, mapping, timestamps, row identities | Parser/mapping/digest/name/time files | NF | keep | NF 6.0.0 scope and import boundary; facade accepts Imports source capability | Generic Imports remains upload/session owner. Do not move the flow parser to generic tabular ingest on framework shorthand. |
| Facade and explicit source contributions | `module.go`, recovery/portability/extension files | NF contribution; App composition | keep | Actual server/extension/recovery callers | Constructors and contribution builders should remain inert. |
| HTTP handlers and error translation | `routes.go`, `graph.go`, `graph_view_routes.go`, query/mapping helpers | NF transport adapter and NF semantic/application code | split | HTTP error types inside query/temporal/composition; workers construct `Service` | Same-package private seam first; owner-neutral errors deferred until exact error characterization. |
| Source graph composition, buckets and contributors | `graph.go`, `graph_temporal.go`, response/projection adapters | NF | split | NF source semantics; GP `ProjectV2` invoked through private adapter | Separate reusable source composer from HTTP service; do not transfer flow meaning into GP root. |
| Deterministic graph derivation/result persistence | GP root and `postgresresult` consumed by target | GP | keep | GP 2.2.0 §8/§8.1 and boundary guard | GP root stays pure; source declaration checks remain NF-owned. |
| Saved declaration mutations and materialization coordination | `graph_view_routes.go`, store/jobs/receipts | NF application, NF store; Jobs generic lifecycle | split | Four commit methods, typed job ports/finalizer | Separate cohesive command methods from HTTP handlers while preserving transaction and replay order. |
| Reporting lease lookup/release SQL | `reporting_graph_source.go` | GP persistence capability, composed by NF source adapter | defer | Direct SELECT/DELETE of Graph lease tables alongside existing lease writer | Proposed `LookupUnexpiredLease` and `ReleaseScopedLeasesTx` are defined in §4.2. RB-002 now gates owner adoption and executed preservation evidence; no broad repository interface is proposed. |
| Restored Jobs enumeration SQL | `graph_restore_source.go` | Jobs typed read contribution; NF payload validation | defer | Direct `FROM jobs`, followed by owner `ReconcileRestoredNonterminalTx` | Proposed `ListRestoredNonterminalTx` is defined in §4.3. RB-003 gates Jobs/NF/Recovery adoption and direct/composed evidence before SQL relocation. |
| Cleanup selection and health accounting | `graph_result_cleanup.go` | NF source coordinator; GP bounded storage operations | defer | Existing result-first locking; health query joins peer results/leases and own declarations | Preserve current metric semantics. Do not move NF declaration SQL into GP or introduce a global reachable-ID list. |
| Own SQL tables and transaction helper | Store/query/declaration/binding files; `transaction.go` | NF persistence/support | keep | Owned source data; explicit AGENTS transaction exception | PG/pgx in adapters/typed transaction ports is intentional. |
| Cross-owner import/link mutation | `transaction_participants.go`, `ports.go` | NF participant; source owners and generic coordinator | keep | Scoped physical capabilities and source-owned logical participant | No extraction to transport/platform; preserve audit/receipt/source mutation atomicity. |
| Collaboration event production | `collaboration_producer.go`; App translator | NF intent; Collaboration delivery/session | keep | Source port already separates generic appender | No new NF WebSocket runtime. |
| Harness controls and reusable test support | `harnesscontrol`; `internal/testutil/networkflowsupport` | NF-specific harness semantics; harness/App composition | keep | Build-tagged server contribution; disabled-route guards; `_test.go` bridges | Do not infer production leakage; trace actual consumption before claiming fault coverage. |
| Frontend state, contract adapter and grid rendering | `apps/web/src/networkFlow`, services adapter; grid package | NF frontend controllers; shared protocol/UI/grid owners | keep | Inspected client/interpreter/grid and boundary test | No move into backend; no direct grid-vendor import observed in inspected grid component. |
| Core timeline/entities/evidence/links/saved views/view schemas | No owning implementation discovered in this target | Existing named Core owners | defer | NF analytical IDs and immutable source rows are separate; Indicators participation is explicit | No supported relocation slice. A test importing revisions support is not NF revision ownership. |

Framework mismatch F-01: the framework's module catalog omits NF/GP and its generic import/projection examples are too broad for this adopted subsystem. Adapt the planning map to NF 6.0.0 and GP 2.2.0; do not edit the framework in this task. Framework mismatch F-02: “small facade/private store” is a desired pattern, while live code exports Store and test-consumed helpers. Classify visibility review separately from behavioral refactoring. Historical phase names and numbered test rows are verification accounting only.

## 4. Public Contract and Behavior Freeze Map

For the first 19 rows, `B` expands exactly to `/api/v1/incidents/{incident_id}/network-flow`, `T` to `{network_flow_table_id}`, and `G` to `{graph_view_id}`. Evidence `routes contract/catalog` means both `contracts/network-flow/routes.v1.json` and `internal/gen/networkflowroutes/catalog_gen.go`, checked against target registration/handler bodies. Success schema names below omit only the common `cartulary.` prefix. All errors retain Core envelope shape, status, code, detail whitelist, conflict information and admission precedence. Existing-tests cells identify coverage sources; missing branch-level proof is explicitly characterization work, not presumed coverage.

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| -------- | ------------- | -------- | -------------- | ------------------------------- | ------------- | ----- |
| R-01 `GET B/source-profiles` → 200 `network_flow.source_profile_list.v2` | NF; viewer admission | Routes contract/catalog; profile handler/resources | Effective-resource-limit route integration; contract helpers | Preserve claimed/unclaimed route and lowerable-limit/profile payload snapshots | Medium | Configuration discovery affects frontend bounds. |
| R-02 `GET B/tables` → 200 `network_flow_table_list.v1` | NF; viewer | Routes contract/catalog; list handler | Route/lifecycle integration | Verify ordering, active-only selection and error envelope before store changes | Medium | Do not expose soft-deleted source rows by refactor. |
| R-03 `GET B/tables/T` → 200 `network_flow_table_get.v1` | NF; viewer | Routes contract/catalog; get handler | Route/store lifecycle tests | Preserve not-found/not-active distinctions and membership privacy | High | Exact incident-scoped identity. |
| R-04 `PATCH B/tables/T` → 200 `network_flow_table_mutation_result.v1` | NF; editor admission | Routes contract/catalog; `application.go` | Table lifecycle unit/integration; intent rollback | Preserve no-op rename, normalized name, expected version, historical replay and transaction recheck matrix | High | `client_txn_id` required; audit and outbox in same commit. |
| R-05 `DELETE B/tables/T` → 200 same table mutation schema | NF; reviewer admission | Routes contract/catalog; table delete transaction | Table lifecycle and route/WS integration | Preserve repeated replay, retained source rows and selected graph invalidation atomically | High | Soft delete; not saved-graph retirement's 204. |
| R-06 `POST B/tables/T/query` → 200 `network_flow.table_query_result.v1` | NF; viewer | Query decoder, SQL/keyset, resources; routes contract/catalog | Query authoring, frozen fixtures, pagination integration | Differential SQL/in-memory ordering, nulls/ties, limit/continuation echo and cursor bindings for any moved logic | High | Immutable analytical rows; no Core edit route. |
| R-07 `POST B/rows/query` → 200 `network_flow.rows_query_result.v1` | NF; viewer | Cross-table scope/query and generated catalog | Cross-table contract helpers; pagination integration | Preserve selected-table canonical order, source table rank, filtering and continuation identity | High | Scope changes invalidate bound cursor. |
| R-08 `POST B/tables/T/rejected-rows/query` → 200 `network_flow.rejected_rows_query_result.v1` | NF; viewer | Diagnostic query/SQL/resources; routes contract/catalog | Parser/diagnostic fixtures; route paging | Preserve safe diagnostic payload, paging and source-unavailable behavior | High | Rejected source is not an editable Entities row. |
| R-09 `POST B/graphs/query` → 200 `network_flow.graph_query_result.v2` | NF source/HTTP; GP derivation; viewer | `graph.go`, projection/response adapters; catalog | Streaming, temporal, projection fixtures, route integration | Same canonical input/output, IDs, limits, cancellation and telemetry from extracted composer | High | Current semantic v2 only; capacity test not a publication claim. |
| R-10 `POST B/graphs/contributors/query` → 200 `network_flow.graph_contributor_query_result.v2` | NF; viewer | Graph contributor handler/iterator; catalog | Bounded-contributor integration; streaming/unit helpers | Preserve selector validation, complete semantic echo, retained prefix, ordering and bound continuation | High | Contributors cannot be inferred from displayed/truncated graph alone. |
| R-11 `GET B/graph-views` → 200 `network_flow.graph_view_list.v4` | NF declarations; viewer | Saved routes/store; routes contract/catalog | Saved lifecycle integration | Preserve active declaration ordering and stale/selected-result representation | High | Distinct from Core saved views. |
| R-12 `POST B/graph-views` → 202 `network_flow.graph_view_accepted.v4` | NF application; Jobs lifecycle; editor | Create commit/enqueue/receipt; catalog | Saved lifecycle, receipt and authority tests | Preserve fresh/replay checks, quota failure, captured actor/job binding and atomic receipt/declaration/audit | High | Historical creation receipt cannot replace newer current declaration in UI. |
| R-13 `GET B/graph-views/G` → 200 `network_flow.graph_view_get.v4` | NF; viewer | Saved get/store; catalog | Saved lifecycle integration | Preserve current declaration/latest-job and unavailable/stale behavior | High | Job status and declaration state are different authorities. |
| R-14 `PATCH B/graph-views/G` → 200 `network_flow.graph_view_mutation_result.v4` | NF; editor | Rename commit/receipt; catalog | Saved lifecycle, graph store, receipts | Preserve name byte limit, version/conflict, replay and generation-independent publication | High | Rename must not gratuitously invalidate a usable materialization. |
| R-15 `DELETE B/graph-views/G` → 204, no success body | NF; reviewer | Retire commit; catalog | Saved lifecycle route tests | Preserve empty response, receipt/replay and retirement event; concurrent publication cannot resurrect declaration | High | No success envelope may be introduced. |
| R-16 `POST B/graph-views/G/refresh` → 202 `network_flow.graph_view_accepted.v4` | NF; Jobs; editor | Refresh commit/enqueue; catalog | Saved lifecycle/time-bucket/authority tests | Preserve expected version, generation, capacity, replay, job notification after commit and old-result continuity | High | Refresh is a Network Flow saved-graph operation. |
| R-17 `GET B/graph-views/G/result` → 200 `network_flow.graph_view_result.v4` | NF selected binding; GP exact reader; viewer | Result handler/store and GP reader; catalog | Saved lifecycle and admission fixtures | Preserve all seven selected-binding members, exact immutable result and mismatch/unavailable errors | High | Never substitute latest result by graph ID. |
| R-18 `POST B/graph-views/G/contributors/query` → 200 `network_flow.graph_view_contributor_query_result.v2` | NF; viewer | Saved contributor handler; catalog | Saved/time-bucket lifecycle integration | Preserve selected result/semantic/source binding across pagination and declaration refresh | High | Same selection must not silently change meaning. |
| R-19 `POST B/indicator-links` → 201 first creation / 200 duplicate, `network_flow_indicator_link_result.v1` | NF initiation; Indicators target; editor | Link handler/participants/receipts; catalog | Indicator boundary and actual route/rollback tests | Preserve selector variants, replay target visibility, source deletion race, candidate equality and transaction failure rollback | High | No new generic Entities mutation or observation semantics. |
| C-01 Generic Imports preview/apply boundary | Imports session/runtime; NF owner facade | `ImportOwner`, `import_facade.go`, `transaction_participants.go`; routes contract import integration | Parser/fixture and error-translation helpers; storage integration indirectly | TODO: trace executable Imports runtime tests for exact replay/source-change/rollback before changing facade; wrapper helpers alone are insufficient | High | Generic mapping preview is `/api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/mapping-preview`; not an additional NF-owned route. |
| C-02 Cursor encoding and binding | NF cursor; Auth session identity | `security.go`, query/graph handlers, key rings | Key-ring/contract/pagination tests | Preserve tamper/expiry/rotation and actor, session, incident, route, query, scope and limit mismatch cases | High | Cursor v2 `nfc2` format, 15-minute TTL, 4096-byte cap; no incidental format migration. |
| C-03 Authorization and replay ordering | Core security/incident admission; NF route policy | Auth/session/CSRF and claim checks in routes; transaction `CheckTx`; NF Table 5B/§28 | Table/graph authority integration; link route privacy; mocked wrapper is limited | Preserve fresh versus historical replay matrix after role/claim/lifecycle/resource changes; do not infer equivalent ordering from helper names | High | Current authorization precedes replay; target visibility rechecked for link replay. No deployment-admin incident bypass. |
| C-04 Atomic mutation, audit, receipt and invalidation | NF application; source participants, Auth receipts, Collaboration outbox | Table commits, saved commits, Store, intent producer, typed finalizer | Lifecycle, intent-failure rollback, authority/cleanup races | At each structural seam preserve all-or-nothing rollback and after-commit notification; characterize missing failure points before moving them | High | Table and saved mutation paths are not mechanically interchangeable. |
| C-05 Collaboration delivery and browser invalidation | Collaboration socket/session; NF event meaning | `GET /ws/v1/incidents/{incident_id}` in Collaboration; App intent translator; NF interpreter/hook | Route/WS intent tests; web event rows | Preserve event profile/resource kind/change/reason, dedup/reset and authorization-loss behavior; socket authorization remains Collaboration-owned | High | NF opens no new WebSocket path. Incident session owns reconnect/resume/sequence; NF receives normalized events. |
| C-06 Materialization jobs and saved-result publication | NF handler/declaration; Jobs generic lifecycle; GP result | Jobs interfaces, worker, finalizer adapter, declaration store | Worker authority, timeout, graph store and saved lifecycle tests | Preserve submitter-role/open-incident rechecks, stale generation, cancel/deadline, result lock order and final proof/receipt | High | Do not move NF semantics into job shell or treat receipt as live state. |
| C-07 Result cleanup and leases | NF selection coordinator; GP result/lease storage | Cleanup service/dispatcher; `postgresresult`; selected declaration lock | Cleanup, race, dispatcher and Reporting lease integration | Preserve one candidate transaction, result-before-declaration lock, source scope, 1000-lease bound, eight-result/30-second sweep bounds and safe retries | High | Health snapshot SQL is a separate deferred seam; do not change counts as incidental cleanup. |
| C-08 Reporting exact graph source | NF source labels/declaration; Reporting job; GP lease/result | `reporting_graph_source.go`; inspected GP lease writer | Reporting-owned exact-result lease integration row | AC-GP-01–07 specify required scope/expiry/identity/sequence/release/rollback/race assertions; existing one-result test does not cover the full matrix | High | NFTR-GP-001–006 define the proposed preservation contract; GP capability adoption remains G-OWNER-GP, not an unresolved design choice. |
| C-09 Restore, recovery and portability contribution | NF source state; Recovery orchestration; GP rebuild; Jobs reconciliation | Recovery state/bindings, graph source registration, portability probe and App composition | GP restore unit/storage rows; Recovery assembly row; configuration portability tests | AC-COMP-02 requires the real NF enumerator; AC-JOB-01–07 require exact selection/payload/mutation/rollback assertions | High | NFTR-JOB-001–005 define the proposed read contract and owner responsibilities; G-OWNER-JOBS and baseline evidence remain open. Existing five source families and portability behavior remain frozen. |
| C-10 Retained state and extension admission | NF compatibility; dependency-owner read ports | State validators; App admission read wrapper; NF §27–28 | State-v4 rejection and saved-graph cutover integration | Preserve bounded scans of declarations, receipts, jobs and proofs; compare bytes before/after failures | High | Major 6/state 4; minimum migratable state 3 is an explicit cutover, not a v1 fallback decoder or automatic repair. |
| C-11 Immutable analytical storage and revisions | NF source storage; Core history owners remain separate | Store/query/resources; NF §6.2 | Store immutability/limit/lifecycle tests; link integration | Preserve accepted rows/mappings, soft-deleted retained state and indicator participation; assert no new generic mutation side effects if seam changes | High | Core Entities row edits, test-util row save/conflict/inspector flows and generic revision/change-set APIs are not owned by NF. Core side effects of Indicators remain with that owner. |
| C-12 Saved views, view schemas and workbook projection refresh | Core saved-view/view-contract/projection owners; NF saved graphs separately | Owner scope; NF resource/query and frontend adapter inspection | NF graph/result tests; frontend boundary policy | No relocation test needed without a discovered dependency; re-trace if future slice touches workbook/view packages | Medium | No NF-owned Core saved-view persistence/view-schema compiler found. Graph refresh/invalidation is applicable; generic workbook projection refresh is not inferred. |
| C-13 Generated protocol, presentation and UI/grid selectors | Adopted owners → authored contracts → generators; frontend adapter/grid owners | NF index/current entrypoint; generated routes; protocol adapter and semantic grid | Projection fixture tests; frontend boundary test; web row routing | Preserve request/response resource IDs, decoder failure shape, UI-contract selectors and row identity; run drift/boundary checks after implementation | High | No hand edits in generated roots, including `packages/view-contracts/src/generated`. No target-owned raw vendor integration found in inspected files. |
| C-14 Frontend saved/table/indicator continuity | NF frontend controllers; shared incident session | Client, interpreter/hook, semantic grid; web manifest | Manifest rows for saved replay/semantic continuity/lifecycle/roles/render bounds, table captured/observation/indicator continuity | Inspect selected controller test bodies before a behavior-affecting slice; preserve captured immutable attempt, original uncertainty, selected result and polling limits | High | NF §28 defines 1500 ms polling, 30 s read timeout and 120 s observation window; listed manifest rows alone do not prove compliance. |
| C-15 Harness control routes | NF harness registries; build-tagged App/harness guard | Four harness files, controls and `server_profile_harness.go` | Disabled/default guard, payload/consume tests; server-process selector | Preserve guard/origin/host/scope/one-shot/clear semantics; separately prove any fault hook the implementation test relies upon | High | POST `/api/v1/test/runtime/network-flow-faults`, `network-flow-randomness`, `network-flow-auth-transitions`, `network-flow-audit-assertions`. Harness build tag plus explicit enablement required in composed server. |
| C-16 Harness accounting and evidence | Verification owners/catalog/family rows; Harness mechanics | Manifests, task surface, topology and generator policy | Existing routed Go/Vitest/Playwright tests, unexecuted here | Maintain exact owner/selector/collaborator/fixture mapping after any moved/added tests; no dropped Reporting or Recovery rows | High | 93 NF rows: 62 Go, 29 Playwright, 2 Vitest; 65 web NF Vitest rows. Names and phase maps are not runtime architecture or coverage proof. |

### 4.1 Preservation requirements and consumer defaults

The map above records observed surfaces and their scoped owners. The requirements below are the single definitions of the additional preservation obligations introduced by this revision. Slice and assertion tables MUST reference these IDs instead of creating competing versions of the contract.

| Requirement ID | Normative requirement |
| -------------- | --------------------- |
| NFTR-FRZ-001 | Every selected structural slice MUST preserve R-01–R-19 and applicable C-01–C-16, including status/envelope/error precedence, authorization, replay, IDs, storage, cleanup, events and frontend continuity. A change to those outcomes, schema, public contract major, retention, error policy or export visibility MUST be separated and marked `requires later authorization`. Generated files MUST NOT be hand-edited. |
| NFTR-COMP-001 | For equal incident/source state, semantic input and effective limits, the extracted source composer MUST preserve canonical source composition, selected-table order/ranks, query/source digests, immutable references, aggregation and bounded iteration. Equality applies to the source composition, not to distinct outer HTTP/worker/restore error envelopes or ephemeral versus saved-result identities. |
| NFTR-COMP-002 | Composer consumers MUST retain the configuration and enumeration policies in the consumer table below. Sharing a composer MUST NOT make restore use deployment overrides or make the worker use default limits. A real NF restore contribution MUST be exercised; a GP test using a fake source is insufficient. |
| NFTR-COMP-003 | The extraction MUST preserve cancellation/deadline propagation, early scan termination, error translation at each existing boundary, telemetry phases/closed dimensions and observer containment. It MUST NOT create new IO, workers, transactions, retries or telemetry observers through construction. |
| NFTR-CMD-001 | Saved-graph command extraction MUST retain the route-admission, exact historical replay and fresh transaction-recheck stages required by NF Table 5B and §28. It MUST NOT replace a historical receipt with current declaration state, replay without current authorization, or consolidate table/link/saved-graph replay paths into a different admission order. |
| NFTR-CMD-002 | Command extraction MUST preserve graph-view version versus materialization generation, same-name no-op behavior, stale-generation publication rejection, and existing rename-compatible publication. Declaration/job/receipt/audit/outbox effects MUST remain atomic at their current transaction boundary; required worker notification MUST occur only after successful commit. Failure MUST NOT leave a partial durable command outcome. |
| NFTR-CMD-003 | Saved-graph retirement MUST retain R-15's empty 204 outcome, replay behavior and retirement event. Read paths and frontend consumers MUST retain the selected immutable result and captured attempt semantics defined by C-06/C-14; extraction MUST NOT substitute a historical creation response for the current declaration. |

| Consumer | Configuration/default | Required source policy | Boundary evidence |
| -------- | --------------------- | ---------------------- | ----------------- |
| HTTP graph request | Existing route/module effective limits and request-admission defaults; no replacement default introduced | Existing scope/order, source validation, response/error translation and graph telemetry | `graph.go`, module/route construction; real HTTP integration path |
| Saved-graph worker | Module-configured limits, store, clock, projection dependency and telemetry observer | Existing semantic input, configured timeout and source snapshot; worker publication/authority policy remains separate from composition | `graph_view_jobs.go`; actual materialization handler path |
| NF restore source | `DefaultEffectiveLimits()` at source registration; existing restore construction does not inject the module telemetry observer | Enumerate active declarations with selected results in `graph_view_id ASC`; reconstructed source snapshot MUST equal both selected and desired source snapshots; retain native-v2 input/exact-binding construction | `NewGraphRestoreSourceRegistration` in `graph_restore_source.go`; actual NF enumerator, not only a fake GP source |

### 4.2 Proposed GP lease capability contract

**Class: proposed owner amendment.** This contract fixes the proposed S-03 design. GP §8 adoption and the NF/Reporting preservation agreement remain required by `G-OWNER-GP`; this tracker does not extend GP's current allowlist.

The logical operations below belong in `internal/modules/graphprojection/postgresresult`, never the pure GP root. The names describe capability boundaries, not new public HTTP endpoints. A consumer MUST receive only the operations it uses; no combined graph repository or generic transaction manager is introduced.

| Requirement ID | Operation and complete logical interface | Construction, defaults and result |
| -------------- | ---------------------------------------- | --------------------------------- |
| NFTR-GP-001 | `LookupUnexpiredLease(ctx context.Context, read BorrowedReadHandle, key ExactLeaseKey, observedAt time.Time) (leaseID string, err error)`; `ReleaseScopedLeasesTx(ctx context.Context, tx pgx.Tx, scope LeaseReleaseScope) error` | `BorrowedReadHandle` denotes the existing adapter-compatible read handle, not a proposed exported root type. Handles/context are required; no fallback handle. Construction MUST reject a missing handle before IO using the adapter's construction-error convention, perform no queries, start no worker and close no borrowed resource. Lookup MUST be read-only. Release MUST use the caller transaction without begin/commit/rollback or pool fallback. There is no default scope, observation time, duration, retry or batch limit. |

| Logical input type/member | Type | Required/default | Meaning |
| ------------------------- | ---- | ---------------- | ------- |
| `ExactLeaseKey.ProjectionResultID` | string | Required; no default | Exact persisted result identity. |
| `ExactLeaseKey.LeaseOwnerID` | string | Required; no default | Exact lease-owner identity. |
| `ExactLeaseKey.LeaseOwnerResourceID` | string | Required; no default | Exact owner resource identity; NF supplies its job ID string. |
| `ExactLeaseKey.LeasePurpose` | string | Required; no default | Exact purpose. |
| `LeaseReleaseScope.SourceOwnerID` | string | Required; no default | Exact owner of the GP result joined to the lease. |
| `LeaseReleaseScope.LeaseOwnerID`, `LeaseOwnerResourceID`, `LeasePurpose` | string each | Required; no defaults | Same exact-match dimensions as lookup, applied across all results in the source scope. |
| `observedAt` | `time.Time` | Caller-supplied; no sampled clock | Lookup compares against this instant, normalized to UTC. Existing renewal validation remains at the renewal stage. |

**NFTR-GP-002 — Exact lookup.** The lookup MUST match precisely `(projection_result_id, lease_owner_id, lease_owner_resource_id, lease_purpose)` and `leased_until > observedAt`. Equality with expiry is expired. It MUST return the stored lease ID; it MUST NOT calculate an ID from a job/result, reacquire a lease, mutate the lease, or join results/declarations to add an earlier source-owner or selection check. Scope values MUST be exact comparisons, never wildcards or omitted predicates. The adapter MUST NOT add pre-lookup normalization or validation that changes existing NF error precedence. A missing or expired match MUST produce `graphprojection.ErrResultV2LeaseNotFound`; query/cancellation/infrastructure failure MUST remain an error, not a successful empty ID or implicit retry. Database-enforced tuple uniqueness remains the basis for an exact single match.

**NFTR-GP-003 — Renewal and read sequence.** NF MUST preserve `LookupUnexpiredLease → existing RenewLease → existing ReadExactResult → NF label-candidate conversion`, after the existing NF entrypoint checks. `observedAt` and `leasedUntil` MUST remain caller-supplied UTC instants; no database `now()` or newly sampled clock may replace them. Existing renewal validation MUST remain authoritative: malformed lease ID, zero observation or requested expiry not after observation yields `ErrResultV2Invalid`; a no-longer-unexpired stored lease yields `ErrResultV2LeaseNotFound`. Successful renewal assigns the requested expiry and observation, including a shorter expiry that is still after observation. It MUST NOT be changed to acquisition's `GREATEST` behavior. The combined read/renew operation MUST NOT become a new atomic transaction: a later exact-read or label-conversion failure may leave the successful renewal committed. No graph content may be returned on binding or conversion failure.

**NFTR-GP-004 — Complete scoped release.** Release MUST delete every lease whose own owner/resource/purpose equals the supplied scope and whose joined GP result has exactly the supplied `source_owner_id`. It MUST join only GP-owned tables. It MUST include expired and unexpired matching leases across all matching results, without a declaration-selection filter, expiry filter, maintenance batch bound or application-built deployment-wide list. Zero matches MUST succeed; SQL/transaction/cancellation failure MUST remain an error. The caller MUST be able to observe deletion in its transaction and restore all deleted leases by rollback. No internal commit, partial-success result or silent truncation is permitted. The 1000-row expired-maintenance bound applies to a different operation and MUST NOT constrain this release.

**NFTR-GP-005 — NF mapping and boundary behavior.** NF MUST supply the mapping below and retain source/declaration interpretation and typed label conversion. Entry checks MUST retain their current placement before lease lookup. GP MUST NOT inspect NF declarations, Reporting state or Jobs state to infer validity, authorization or release timing. Subsequent reads MUST follow the exact leased result without requiring that it remain the declaration's current selection. A refresh or retirement MUST NOT redirect that read. Lease possession MUST NOT become an authorization or redaction grant; NF label components remain typed redaction candidates and Reporting's external-release policy remains fail-closed without a separately adopted allow rule.

| Mapping or wrapper condition | Required value/outcome |
| ---------------------------- | ---------------------- |
| NF source owner | `network_flow_activity` |
| Lease owner | `snapshot_reporting` |
| Lease owner resource | `jobID.String()` |
| Lease purpose | `render` |
| Read entrypoint: nil source, nil DB, zero job ID, or supplied binding source owner other than NF | `ErrResultV2BindingMismatch` before lookup; no new early full-binding predicate |
| `ReleaseJobLeasesTx`: nil source, nil transaction or zero job ID | Existing successful no-op; MUST NOT call the new capability |
| Outer `ReleaseJobLeases`: nil source, nil DB or zero job ID | Existing successful no-op; otherwise retain its explicit begin/deferred rollback/commit ownership around the borrowed-transaction capability |

**NFTR-GP-006 — Lifetime and races.** Reporting MUST retain its existing decision about when a reader is finished: renew while an attempt can read, reuse the exact reference/purpose on retry, release only after durable state no longer depends on Graph rows, and retain uncertain-reader reachability. Acquisition, publication and cleanup MUST preserve result-row-before-source-declaration lock order and same-transaction rechecks. The release capability MUST NOT acquire source-declaration locks or reinterpret Reporting lifecycle. Controlled-interleaving tests MUST establish retained-result safety for a committed live lease, selected binding and acquired result lock, and prove absence of cross-scope deletion. A lease lost before renewal may yield the existing lease-not-found outcome; renewal MUST NOT resurrect a deleted result or lease. These constraints do not authorize an algorithm or retention-policy change.

### 4.3 Proposed Jobs restored-nonterminal read contract

**Class: proposed owner amendment.** The proposed operation belongs to the existing Jobs owner package. `G-OWNER-JOBS` requires Jobs/NF/Recovery adoption; naming this interface does not establish that it exists. It is an internal restore capability, not a new public job-list or restore endpoint.

```go
type RestoredNonterminalScope struct {
    JobKind                 string
    ExtensionOwnerProfileID string
}

type RestoredNonterminalJob struct {
    JobID              uuid.UUID
    IncidentID         uuid.UUID
    HandlerPayloadJSON json.RawMessage
}

func ListRestoredNonterminalTx(
    ctx context.Context,
    tx pgx.Tx,
    scope RestoredNonterminalScope,
) ([]RestoredNonterminalJob, error)
```

**NFTR-JOB-001 — Interface and resource ownership.** The projection MUST use the required borrowed transaction in the existing read/write Graph restore operation. It MUST NOT begin a read-only transaction, acquire another connection, change isolation, commit, roll back, mutate jobs, enqueue work or notify workers. Context, transaction and both scope strings are required; the proposed internal operation MUST reject missing values with the Jobs `ErrInvalidJobDefinition` family before IO. Existing NF preflight errors remain at the NF entrypoint. Nonempty scope strings are exact identities, not a query language. Successful results MUST own payload bytes that remain usable after the query closes. Query, scan, iteration or cancellation failure MUST return an error without a successful partial collection; existing NF error propagation and rollback obligations remain intact. Empty success has length zero; nil versus allocated empty Go slice is an intentional representation choice, not public JSON behavior.

**NFTR-JOB-002 — Complete selection.** Jobs MUST select exactly its positive status set `queued`, `running`, `cancel_requested`, with equality on both scope fields, across all matching incidents in the restored target, ordered by database `job_id ASC`. It MUST NOT use “not terminal,” add a due-time/attempt-expiry/declaration/profile-activity/retention filter, or apply an implicit cap, page size or truncation. The full selected set MUST be consumed and the query closed before NF starts reconciliation writes. Payload bytes MUST be returned as read from storage without Jobs-side decoding, canonicalization or reserialization. Byte equality is relative to that storage read, not to pre-storage JSON text. No deployment-wide memory bound is claimed for the current full-materialization behavior; an NF per-incident quota MUST NOT become a global restore limit.

**NFTR-JOB-003 — NF identity and payload ownership.** NF MUST provide the exact scope and retain the payload validation table below. Jobs MUST NOT interpret this payload. The restore decoder MUST retain its unknown-member rejection and current validation semantics; this refactor MUST NOT unify it with the different worker decoder, introduce a legacy decoder or add unrelated decoder strictness. Any current decoder behavior outside the listed validation MUST remain unchanged under NFTR-FRZ-001.

| Scope/member/condition | Type or boundary | Required NF value or validation |
| ---------------------- | ---------------- | ------------------------------- |
| `JobKind` | string | `network_flow_activity.graph_view_materialize_v1` |
| `ExtensionOwnerProfileID` | string | `network_flow_activity` |
| `schema_id` | string | Exactly `cartulary.network_flow.graph_view_materialization_payload.v1` |
| `incident_id` | UUID | Nonzero and equal to the projected job incident |
| `graph_view_id` | string | Existing NF validator: `^nfgv_[a-f0-9]{32}$` |
| `materialization_generation` | int64 | Greater than zero |
| `source_snapshot_id` | string | Nonempty under the current validator; no new digest validator |
| Unknown JSON members / malformed payload | Restore decode | Reject through the current restore decode/validation error path |
| Missing required member | Current typed decode/validator | Rejected by the corresponding zero-value/schema/identifier condition; no invented defaults |

**NFTR-JOB-004 — Existing reconciliation mutation.** For each selected row in order, NF MUST validate its payload and call `jobs.ReconcileRestoredNonterminalTx(ctx, tx, jobID, GraphViewMaterializationJobKind)`. That existing operation MUST retain its row lock, job-kind/nonterminal checks and resumable failure-count validation against the existing Jobs attempt bound. The only changed columns are `handler_attempt_id = NULL` and `handler_lease_expires_at = NULL`. Status, cancellation intent, failure count, retry schedule, progress, timestamps, receipts and proofs MUST remain unchanged. There is no reset-to-queued, retry-budget reset or recreation default. Success MUST return the selected job count, including rows whose attempt fields were already clear; empty selection returns `0, nil`, and repeat reconciliation may return the same count. Any error MUST return no partial success count and MUST cause the caller to roll back earlier changes in the transaction.

**NFTR-JOB-005 — Recovery composition.** The read projection MUST run only inside the admitted, quiescent restore context with Recovery's existing writer exclusion and isolation policy. A borrowed transaction alone MUST NOT be treated as evidence of writer quiescence or a stronger snapshot. Graph clear/rebuild/publication, NF reconciliation, mandatory Reporting job/lease reconciliation and postcondition verification MUST remain in the same Graph participant transaction. Failure or cancellation in a later phase MUST roll back preceding NF job changes with the rest of Graph mutation. Readiness MUST wait for successful reconciliation/postconditions and durable completion under Core 01; failed or indeterminate restore MUST NOT publish readiness. The operation MUST NOT introduce a generic Jobs transaction manager or weaken existing target-reinitialization rules.

## 5. Coupling and Boundary Findings

Classification describes future planning priority, not authority to perform a fix now. An observed implementation/owner mismatch is not an owner contradiction. Only disagreement between actual owner documents receives the exact contradiction blocker label.

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| ------- | -------- | ---- | -------------- | -------------- | ------------------------ |
| F-01 Framework catalog and generic examples omit or blur NF/GP ownership | Framework versus adopted NF 6.0.0/GP 2.2.0 and live modules | Incorrect redistribution of legitimate behavior | must_fix | Planning artifact | Correct owner map in this tracker; no framework edit. Done for this planning session. |
| F-02 Facade exists, but public Store/Service surface exceeds traced production use | `module.go`, `store.go`, `routes.go`; App imports and test dot imports | Export removal can break tests or untraced consumers | defer | NF facade/support | Keep exports for initial slices; TODO: repository-wide compatibility decision before visibility changes. |
| F-03 Worker/restore construct HTTP Service to reuse graph source composition | `graph_view_jobs.go`, `graph_restore_source.go`, `graph.go` | Transport dependencies obscure required graph dependencies | should_fix | NF private application composer | S-01 MUST satisfy NFTR-COMP-001–003 and G-EVID-S01, including distinct consumer defaults and the real NF restore contribution. |
| F-04 Semantic query/graph/temporal/mapping logic returns or constructs HTTP errors | `query.go`, `query_filter.go`, `graph.go`, `graph_temporal.go`, `mapping.go`, projection adapter | Error precedence/status/detail drift during cleanup | should_fix | NF semantic logic plus NF transport adapter | Characterize exact errors first. Defer broad error-type redesign; initial composer slice can retain existing return types. |
| F-05 Saved graph route file also owns command transactions | Four commit methods in `graph_view_routes.go` | Admission, replay, audit or job notification can change order | should_fix | NF application command layer | S-02 MUST satisfy NFTR-CMD-001–003 and G-EVID-S02; only cohesive private command methods move, with the current operation order preserved. |
| F-06 Reporting source directly reads/deletes Graph lease storage | `ReadAndRenewLeasedResult`, `ReleaseJobLeasesTx`; GP lease writer only supplies current narrow operations | Peer schema coupling and accidental wider lease deletion | must_fix | GP persistence capability; NF source composition | Design specified by NFTR-GP-001–006. S-03 MUST await G-OWNER-GP and G-EVID-S03; the two-operation contract MUST NOT be treated as already permitted by current GP §8. |
| F-07 Restore source directly enumerates Common Jobs rows | `ReconcileGraphRestoreJobsTx` SQL; owner reconciliation function already used for mutation | Jobs schema/status knowledge leaks into source adapter | should_fix | Jobs typed read projection; NF payload validation | Design specified by NFTR-JOB-001–005. S-04 MUST await G-OWNER-JOBS and G-EVID-S04; NF payload validation and existing Jobs mutation remain separate from enumeration. |
| F-08 Cleanup health snapshot reads Graph result/lease tables and own declarations | `cleanupHealthSnapshot` SQL | Moving whole join into GP violates source independence; changing snapshot affects metrics | defer | NF accounting with GP capability | TODO: owner-reviewed bounded composition preserving current counts, age and failure handling. Not included in execution-ready slices. |
| F-09 Own-table SQL and private PG transaction helper are intentional | Store/query files, ports; explicit repository procedure | Generic abstraction can hide transaction ownership | intentional/no_action | NF persistence/support; platform connection port | Preserve helper and explicit transaction participants. Domain calculations need no new SQL dependencies. |
| F-10 Cross-owner import/link side effects are explicit, transactional contributions | Module capabilities, participant scopes, incident lock, audit/receipt/target ports | Splitting transaction breaks authorization or atomicity | intentional/no_action | NF/source owners; generic coordinator | Freeze capability identity, source validation, audit and replay sequence; preserve real rollback tests. |
| F-11 Collaboration producer already has proper source intent/App translator split | Intent port and server/test support translators | Unnecessary move would couple domain to delivery runtime | intentional/no_action | NF intent; App translator; Collaboration delivery | Preserve current seam and transactional outbox. |
| F-12 Harness registration is isolated; test bridges compile only in tests | `cartulary_harness` server file, HTTP test guard, `_test.go` bridges | Misclassifying support code causes broken harness or unsafe exposure | intentional/no_action | NF harness support; App composition | Preserve disabled route/process tests. TODO: trace a registry consumer before relying on a fault-injection claim. |
| F-13 Some behavioral wrappers overstate exercised boundary | Import parser helpers, mocked auth and reflection-only egress assertion; unit routing despite `Integration` names | False confidence in refactor preservation | must_fix | NF/Imports verification owners | S-00 MUST satisfy NFTR-GATE-002 and NFTR-EVID-001 using §8.3 assertions; test names, mock-only substitutes and manifest rows MUST NOT satisfy a real-boundary evidence gate. |
| F-14 No demonstrated duplicate Core row/view-schema logic or raw grid-vendor implementation | Target helpers/SQL own analytical semantics; frontend semantic grid imports adapter | Unfounded relocation can change identities/editability | intentional/no_action | NF; existing Core/UI owners | Preserve frontend adapter boundary; do not create a speculative frontend/backend move. |
| F-15 Generated roots and current contract versions require explicit freeze | Policy includes Go, protocol, view-contract and UI-contract generated roots; major 6 index | Hand-edit drift or accidental old decoder support | intentional/no_action | Authored contract owners/generators | No generated writes now; use Make drift after authorized changes. Historical v1/v3 names are not conformance decisions. |
| F-16 Authorization belongs at admission plus transaction/publication rechecks | Route admission, table/saved commits, link participant, worker | Consolidation may erase fresh/replay distinctions or race protection | intentional/no_action | Core security/incident owner and NF command policy | Freeze existing stage order; if owners and code disagree, classify separately and obtain later behavior-change authorization. |
| F-17 Shared digest/name/time helpers have NF semantic ownership | Stable IDs, flow fields, graph/table different name bounds, pinned timezone | Generic utility extraction obscures compatibility | intentional/no_action | NF | Keep helpers source-owned; no shared `utils` package. |

## 6. Refactor Workstreams

These IDs are local planning workflows, not historical verification phases. WF-03 and WF-04 MAY proceed independently after owner mapping. Every workflow MUST produce its named artifact and satisfy its checkpoint. Planning completion MUST NOT imply owner adoption, execution authorization or successful tests; those conditions are governed separately by §8.1. No runtime architecture may be inferred from workflow or evidence-row numbering.

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| ----------- | ---- | -------------------------- | --------------------------- | ----------------------------- | ---- | --------------------- | ---------- | ------------------ |
| WF-00 | Source/session bootstrap | root | None | WF-01 | Establish branch, scope, framework-first read and write restriction | Framework, AGENTS, tracker | Read-only Git/path checks | §1 source ledger and session entry; exit: target/label/output and authority explicit. |
| WF-01 | Complete NF file/caller inventory | chain | WF-00 | WF-02 | Account for all 88 files, facade consumers and test/support surfaces | Target; App callers in §1 | Compare filesystem paths with §2 rows | §2 inventory; exit: 88 unique paths, no unexplained omission. |
| WF-02 | Owner and observable-contract mapping | chain | WF-01 | WF-03, WF-04 | Map routes/current ownership and distinguish proposed GP/Jobs owner amendments | NF/GP/Core docs, route contracts, source bodies, frontend adapters | Catalog/handler comparison and scoped owner review | §§1.2/3/4; exit: every contract has an owner/evidence class; proposed interfaces have exact behavior and adoption placement. |
| WF-03 | Characterization gap analysis | parallel | WF-02 | WF-05 | Map each preservation requirement to a concrete assertion, fixture boundary and evidence phase | Target tests; NF/web/Reporting/GP/Recovery family manifests | Body/fixture/selector review; future narrow Make selections | §8.2–8.3; exit: all NFTR requirements mapped; real-caller gaps and post-only capability assertions identified without circular baseline gates. |
| WF-04 | Coupling and boundary scan | parallel | WF-02 | WF-05 | Assess HTTP, peer SQL, transaction, generated, harness and frontend seams | Composer/worker/restore, reporting leases, store/ports, App harness, frontend grid | Read exact implementations and actual inbound callers | §5 findings; exit: each finding classified with owner and next action. |
| WF-05 | Private seams and owner-capability contracts | chain | WF-03, WF-04 | WF-06 | Specify private seams and complete proposed GP/Jobs interfaces while retaining adoption gates | `module.go`, graph/route/port files; adjacent owners | Check NF/GP ownership and no new public behavior | §4.1–4.3; exit: design definitions complete, unknown design placeholders removed, unadopted capability status explicit. |
| WF-06 | Small-slice sequencing | chain | WF-05 | WF-07 | Separate S-00 authorization, per-slice baseline evidence and authorized production work | Likely files in §7 only in later task | Dependency/risk/exit review | §7/§8.1; exit: no baseline-before-S-00 cycle; independent selected slices, rollback and completion criteria explicit. |
| WF-07 | Harness/accounting and command plan | chain | WF-06 | WF-08 | Preserve accounting owners/collaborators and specify reproducible evidence records | Authored family/catalog/verification inputs, Make/task surface | Static recipe and row-ID inspection; no target executed | §8; exit: canonical commands/current owner rows recorded; missing new selectors marked TODO; execution results and adoption not fabricated. |
| WF-08 | Documentation audit and final handoff | chain | WF-07 | None | Audit normative references, preserved evidence and tracker-only unstaged revision | Tracker only | Read-only structural audit and Git diff/status | §§10–12; exit: document audit passes, staged blob/history retained, actual implementation readiness reported separately. |

## 7. Proposed Refactor Slice Plan

The slice plan MUST use §8.1 gates. S-00 MUST NOT start before G-AUTH-00 passes; it produces the selected evidence gates rather than depending on them. A production slice MUST have its own G-AUTH-PROD authorization and applicable passing baseline; S-03/S-04 additionally require owner adoption. Their designs are specified, but execution remains blocked. S-01/S-02 are independent after their respective characterization. NFTR-FRZ-001 governs all behavior-change and generated-file exclusions.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| -------- | ---------- | --------------- | ------------------------------ | -------------- | ------------------------ | ------------------ | ------------- | -------------------- |
| S-00 | G-AUTH-00 only; no already-passing baseline prerequisite | Inspect/add/run the selected baseline assertions in §8.3 and record §8.2 evidence. New-capability-only conformance tests remain post-change requirements. No production changes are included. | Existing graph/route/authority/receipt/transaction tests; authored owner family rows only if new selectors require accounting | Unexecuted or shallow wrappers mistaken for evidence; loss of consumer-specific policy | AC-COMP-01–03 for S-01; AC-CMD-01–03 for S-02; AC-GP-01–07 or AC-JOB-01–07 current-boundary baseline assertions only when those slices are selected | `make task-guide ROLE=module-author OWNER=module.networkflow`; selected `make test-slice` / `make service-backed-test-slice` commands in §8 | Revert only later characterization/accounting additions if wrong; do not weaken existing tests | Each selected G-EVID gate passes its baseline obligations with exact assertion/fixture/row/run records; failures remain explicit. Passing one slice MUST NOT depend on unselected scopes. |
| S-01 | G-AUTH-PROD names S-01; G-EVID-S01 produced by S-00 | Introduce a private same-package source composer and move `composeGraphSource`, `composeGraphSourceFromSemantic` and `resolveGraphTables` responsibilities behind it. Route service, worker and restore contribution use that composer without constructing an HTTP Service solely for reuse. Retain existing APIError return semantics in this slice. | `graph.go`, `graph_view_jobs.go`, `graph_restore_source.go`, `graph_telemetry.go`; proposed private `graph_source_composer.go` (does not exist yet) | NFTR-COMP-001–003: canonical source behavior, distinct consumer limits, real restore policy, errors and telemetry | AC-COMP-01–03 baseline and post-change assertions; retain affected GP/Recovery rows | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.bounded_graph_aggregation_and_selectors,module.networkflow.unit.time_bucket_graph_backend,module.networkflow.unit.graph_telemetry_boundary`; `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.saved_graph_lifecycle_v2,module.networkflow.integration.bounded_graph_contributor_pipeline`; Recovery selections in §8 as applicable | Revert private extraction and wiring together; no durable migration or export removal | Private composer serves all three real consumers without Service-only construction; NFTR-COMP-001–003 post-change assertions pass and NFTR-FRZ-001 holds. Readiness alone is insufficient. |
| S-02 | G-AUTH-PROD names S-02; G-EVID-S02 produced by S-00; independent of S-01 | Move the four saved-graph commit methods and their cohesive transaction/replay/enqueue helpers into a private application file in the same package. Keep handlers/HTTP decoding in route file; retain current Service receiver and operation order initially. | `graph_view_routes.go`; proposed `graph_view_application.go` (does not exist yet); jobs/receipts/store only where required to preserve current calls | NFTR-CMD-001–003: replay/admission, version/generation, no-op, atomic side effects, notification and empty retirement | AC-CMD-01–03 baseline and post-change assertions; preserve saved receipt/store/authority/lifecycle/cutover coverage | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.saved_graph_receipt_integrity,module.networkflow.unit.transaction_lifecycle`; `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.saved_graph_lifecycle_v2,module.networkflow.integration.saved_graph_cutover_v6,module.networkflow.integration.time_bucket_saved_graph_lifecycle` | Revert application extraction as one unit; do not revert already-issued receipts or result state | Four cohesive commit bodies leave the HTTP handler file; NFTR-CMD-001–003 post-change assertions pass, eight saved routes remain unchanged and NFTR-FRZ-001 holds. |
| S-03 | G-AUTH-PROD names S-03; G-OWNER-GP; G-EVID-S03; currently BLOCKED | Implement the two capabilities specified by NFTR-GP-001–006 in the existing GP persistence adapter and route NF lease lookup/release through them; NF declaration and label semantics stay with NF. | `reporting_graph_source.go`; `internal/modules/graphprojection/postgresresult`; Reporting caller/contract guard only if owner-approved | Exact predicate asymmetry, stored identity, expiry/renewal sequence and committed side effects, complete scoped release and borrowed-transaction ownership | AC-GP-01–07 baseline and post-change assertions, including post-only capability construction; preserve Reporting row, NF races and GP package guards | `make service-backed-test-slice OWNER=module.reporting ROWS=module.reporting.integration.exact_graph_result_lease_lifecycle_8f1c5c43a2`; `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.graph_result_cleanup_races`; GP selection through task guide | Revert capabilities and consumer wiring together without data rewrite. Owner amendment remains an authority record; any changed behavior requires later authorization. | Adoption recorded; NF no longer directly selects/deletes GP lease rows; capability conformance and all preservation assertions pass; Reporting lifetime/labels and GP pure-root boundary remain unchanged. |
| S-04 | G-AUTH-PROD names S-04; G-OWNER-JOBS; G-EVID-S04; currently BLOCKED | Implement the specified Jobs ListRestoredNonterminalTx projection and replace NF restore enumeration SQL with that call. Preserve NF payload validation and existing per-row Jobs reconciliation. | `graph_restore_source.go`; `internal/platform/jobs/restore_reconciliation.go` or owner-selected Jobs support file; App recovery assembly | NFTR-JOB-001–005: complete selected set/order/bytes, exactly two-column mutation, borrowed transaction, rollback/quiescence/readiness | AC-JOB-01–07 baseline and post-change assertions, including direct NF state checks and post-only projection conformance | `make task-guide ROLE=module-author OWNER=platform.jobs`; `make task-guide ROLE=module-author OWNER=module.recovery`; `make service-backed-test-slice OWNER=module.recovery ROWS=module.recovery.integration.atomic_typed_terminal_evidence_7a41f0d8c2`; TODO: resolve/add authored Jobs selector for the new projection; no existing selector asserted | Revert projection and consumer together; no status/data migration | Jobs owns enumeration SQL; exact NF identity/validation and selected-count semantics remain; direct and composed post-change assertions pass with no new transaction/isolation or readiness behavior. |
| S-05 | Only production slices selected and completed; no dependency on unselected S-03/S-04 | Reconcile only affected authored test selectors/collaborators and perform final drift/boundary/consumer validation. If no selector changes, leave accounting unchanged. | Authored `tools/test_families`/catalog/verification inputs only as necessary; no direct generated edits | Lost cross-owner test evidence, generated drift, frontend decoder/selection regressions | Preserve Reporting-owned NF test, GP/Recovery/Jobs collaborators and affected harness/web rows; reconcile new assertion-to-selector mappings | `make generate-drift`; `make frontend-import-boundary-check`; `make test-slice OWNER=web.networkflow`; appropriate browser target; `make agent-finalize` before broader `make check`, subject to later write authorization | Revert authored accounting changes and regenerate through owning Make targets; never manually restore individual generated fragments | Selected post-change assertions and justified broader checks pass with NFTR-EVID-001 records; owner/selector/fixture mappings remain accurate; no generated or public contract drift is unexplained. |

Not sequenced: broad owner-neutral error redesign, export shrinking, whole-package relocation, moving NF parsing to Imports, moving declarations into GP, cleanup-health redesign, UI/controller relocation or raw grid integration changes. These lack a sufficiently narrow approved need/design for this task. Performance/fixture-sensitive publication is also outside this tracker; apply Core 05 only if a later task actually makes such a claim.

## 8. Validation Plan

**No product tests or Make validation targets were run in either documentation session.** The commands below were discovered in the live Make/task surface and matched to authored owner/family rows. They are future implementation validation, not prerequisites for revising this tracker. Execution MUST occur from the repository root after the applicable authorization gate. A later implementation session MUST resolve current narrow coverage through `make task-guide ROLE=module-author OWNER=module.networkflow` and, as needed, `make explain-test-owner OWNER=module.networkflow` or `make explain-target TARGET=test-slice DETAIL=rows`. `make help` / `make help-all` remain the public inventory; this is a target-specific selection, not a copied repository task catalog.

### 8.1 Authorization, adoption and characterization gates

| Requirement ID | Normative requirement |
| -------------- | --------------------- |
| NFTR-GATE-001 | `G-AUTH-00` MUST name the selected S-01–S-04 preservation scopes, permitted S-00 tests/fixtures/authored accounting, and permitted execution artifacts/Make targets. This authorization MUST precede S-00 and MUST NOT require an already-passing baseline. `G-AUTH-PROD` MUST separately name each authorized production slice and its permitted changes. Authorization of characterization alone MUST NOT authorize production changes. Neither gate permits owner-document edits unless separately named. |
| NFTR-GATE-002 | S-00 MUST establish a separate evidence gate for each selected production slice using §8.2 and the applicable baseline assertions in §8.3. A gate MUST NOT pass from an empty selector, skipped required service-backed test, uninspected wrapper, mock-only replacement for a required real boundary, or unexplained failure. Only relevant baseline preservation assertions are prerequisites to production work; conformance tests of a not-yet-implemented capability are post-change completion requirements, not a circular S-00 prerequisite. |
| NFTR-GATE-003 | S-03 MUST additionally satisfy `G-OWNER-GP`; S-04 MUST additionally satisfy `G-OWNER-JOBS`. Each gate MUST contain the adoption record required by §1.2. Publishing this tracker or completing a design row MUST NOT satisfy either gate. Existing adopted clauses MAY satisfy preservation obligations through an exact recorded reference; GP §8's excluded capabilities still require explicit amendment. |
| NFTR-GATE-004 | Only selected slices participate in readiness and RB-001 closure. S-01 and S-02 MUST remain independently selectable after their own evidence gates; unselected S-03/S-04 MUST NOT block them. By default this session selects no executable slice. A later initial structural task SHOULD select S-00 plus S-01 or S-02; broader scope requires explicit selection. A ready slice is not complete: completion additionally requires its actual change, post-change assertions and final accounting/validation. |

| Gate ID | Applies to | Pass condition | Current status |
| ------- | ---------- | -------------- | -------------- |
| G-AUTH-00 | Starting S-00 | Recorded later task authorization satisfying NFTR-GATE-001 | TODO — current task authorizes tracker revision only |
| G-AUTH-PROD | Starting each selected production slice | Recorded later task explicitly names that slice and permits its changes | TODO — no production slice authorized |
| G-EVID-S01 | S-01 | AC-COMP-01–03 baseline obligations pass with complete evidence records | TODO — unexecuted |
| G-EVID-S02 | S-02 | AC-CMD-01–03 baseline obligations pass with complete evidence records | TODO — unexecuted |
| G-EVID-S03 | S-03 | AC-GP-01–07 baseline preservation obligations pass against current NF/GP caller paths | TODO — unexecuted; new-capability-only conformance remains post-change |
| G-EVID-S04 | S-04 | AC-JOB-01–07 baseline preservation obligations pass against actual NF enumeration/reconciliation and composed restore | TODO — unexecuted; new-port-only conformance remains post-change |
| G-OWNER-GP | S-03 | GP §8/§8.1 permission adopted and NF/Reporting mapping/lifetime/preservation agreement recorded | BLOCKED — proposals in §4.2 are specified but not adopted |
| G-OWNER-JOBS | S-04 | Jobs projection, NF mapping/validation and Recovery transaction/quiescence/readiness contribution adopted or satisfied by exact existing owner references | BLOCKED — proposed read contract in §4.3 is not adopted |

### 8.2 Evidence record contract

**NFTR-EVID-001.** Each baseline and post-change run used to satisfy a gate or completion criterion MUST have the following record. A field that is genuinely inapplicable MUST carry an explicit reason; missing data MUST remain `TODO`, not implied success. Unrelated failures MAY be classified separately, but MUST NOT be counted as passing evidence. Run IDs/artifacts MUST remain separate from adopted behavioral authority; no new product check may depend on this Markdown.

| Evidence field | Required content |
| -------------- | ---------------- |
| Evidence ID and phase | Stable local record ID, selected slice, baseline or post-change phase, requirement and assertion IDs |
| Code and owner basis | Commit; relevant worktree delta including staged/unstaged distinction; applicable owner document versions and adoption references |
| Execution selection | Exact public Make command; resolved owner row IDs; actual executed test symbols/scenarios; runtime/evidence class; empty-selector check |
| Fixture basis | Fixture identity and digest, or fixture-construction source identity/digest; effective configuration; controlled clocks/seeds; backing-service and runtime details sufficient to reproduce the assertion |
| Execution result | Start/end or run identity; exit status; artifact/run-root location; executed/skipped/failed cases and reasons; preserved safe failure details |
| Invariant result | Per-assertion outcome and the concrete assertion/fixture establishing it; unresolved cases and failure classification; reviewer/session accepting the gate |

No executable evidence record exists for this revision. The read-only documentation audit is document-maintenance evidence only. Tests with `Integration` in their name MUST be classified from their bodies, selectors and fixtures rather than the suffix.

### 8.3 Invariant-to-assertion matrix

The following assertion IDs are required acceptance scenarios, **not claims that corresponding test symbols or new owner rows already exist**. `Baseline + post` requires the current caller boundary before the move and the revised boundary afterwards. `Post-only` identifies construction/new-operation assertions that cannot execute before implementation. S-00 MUST locate an inspected assertion or add the missing case under its authorization; an anchor filename is not proof of coverage.

| Assertion ID | Requirement mapping | Required assertion and fixture/scenario | Existing anchor / missing evidence | Phase and owner accounting |
| ------------ | ------------------- | --------------------------------------- | ---------------------------------- | -------------------------- |
| AC-DOC-01 | NFTR-GOV-001, NFTR-GOV-002, NFTR-GOV-003, NFTR-GOV-004, NFTR-GATE-001, NFTR-GATE-002, NFTR-GATE-003, NFTR-GATE-004, NFTR-EVID-001 | Verify only tracker delta, unchanged staged blob, retained inventory/route/history rows, valid references, independent gates and no fabricated adoption/run results; separate owner authority from support prose | Current revision's read-only audit; §12 | Documentation only; MUST NOT count toward a production evidence gate |
| AC-COMP-01 | NFTR-FRZ-001, NFTR-COMP-001 | Drive equal semantic input/source state/effective limits through real HTTP, actual worker and actual NF restore contribution; compare canonical source composition, table order/ranks and digests; exercise default and time-bucket aggregation and bounded iteration | Streaming/temporal/route tests anchor parts; direct three-consumer assertion remains TODO | Baseline + post; NF with GP/Recovery collaborators where exercised |
| AC-COMP-02 | NFTR-COMP-002 | Give module limits a value differing from defaults; prove worker retains configured policy while actual restore uses defaults; insert declarations out of order and cover inactive/unselected exclusion plus desired/selected snapshot mismatch | `graph_restore_source.go` body confirms policy; generic GP fake-source tests do not close this gap | Baseline + post; NF restore assertion plus Recovery/GP composed evidence |
| AC-COMP-03 | NFTR-COMP-003 | Cancel during scan and expire worker deadline; assert bounded termination, no late publication, unchanged consumer-specific errors and telemetry; contain observer panic; constructors start no work | Streaming, timeout, telemetry and authority test bodies | Baseline + post; NF unit/integration rows selected in command table |
| AC-CMD-01 | NFTR-FRZ-001, NFTR-CMD-001 | Cover fresh versus exact replay after role/lifecycle/version changes; require current admission and transaction recheck; preserve historical response instead of substituting current declaration; characterize denied replay without losing the original attempt | Saved receipt/authority/lifecycle tests and web replay row; branch-to-assertion completion remains TODO | Baseline + post; NF; web continuity when impacted |
| AC-CMD-02 | NFTR-CMD-002 | Cover same-name no-op, rename-compatible publication and stale generation; inject failures at declaration/job/receipt/audit/outbox boundaries and assert total rollback; capture notifications and prove none precedes successful commit | Graph store/lifecycle/rollback tests anchor parts; per-failure-point assertion mapping required | Baseline + post; NF with Jobs/outbox collaborators actually exercised |
| AC-CMD-03 | NFTR-CMD-003 | Assert empty retirement 204 and replay/event outcome; refresh/retire during a captured read/attempt and preserve exact selected-result/attempt continuity | Saved lifecycle and relevant web manifests; inspect actual selected web test bodies before claiming coverage | Baseline + post; NF/web |
| AC-GP-01 | NFTR-GP-001, NFTR-GP-002, NFTR-GP-005 | Vary result, lease owner, resource/job and purpose independently; assert exact-match lookup and no content on wrong binding/source; prove lookup does not acquire the release-only source join; exercise existing NF read/release preflight/no-op cases | Reporting exact-result integration covers only a subset; new scope matrix required | Caller preservation: baseline + post; new capability constructor/borrowed-handle conformance: post-only; Reporting/NF/GP |
| AC-GP-02 | NFTR-GP-002, NFTR-GP-003 | Use controlled observations before, equal to and after expiry; test zero observation/invalid requested interval at the existing stage, shorter still-valid expiry and repeated renewal; no sampled clock or reacquisition | Existing one-result lease test is insufficient | Baseline + post; GP capability assertions additionally post-only |
| AC-GP-03 | NFTR-GP-002 | Persist a lease ID different from NF's deterministic acquisition candidate; acquire/reuse the same tuple and assert lookup returns the stored ID | AcquireLease conflict SQL inspected; direct fixture/assertion required | Baseline + post; Reporting/NF caller and GP adapter |
| AC-GP-04 | NFTR-GP-003, NFTR-GP-005 | Vary complete result binding members; assert no returned content on mismatch/conversion error, and verify the already-successful renewal remains committed; test exact leased read after declaration selection changes | Reporting source sequencing inspected; failure-order fixtures required | Baseline + post; Reporting/NF/GP |
| AC-GP-05 | NFTR-GP-004 | Release multiple results for one job, mixing expired/unexpired leases and neighboring source/owner/job/purpose scopes; assert complete intended deletion, untouched neighbors, and success with zero matches; no 1000-row maintenance truncation | Existing test releases one lease only; multi-result/scope fixture required | Baseline + post caller behavior; scoped-capability predicate contract post-only; GP/Reporting |
| AC-GP-06 | NFTR-GP-001, NFTR-GP-004 | Observe scoped deletion inside caller transaction, roll back and verify all rows return; inject execution failure and prove no hidden commit/connection fallback | Reporting fixture already supplies explicit transaction; rollback assertion missing | Baseline + post; new capability resource ownership post-only |
| AC-GP-07 | NFTR-FRZ-001, NFTR-GP-005, NFTR-GP-006 | Synchronize acquisition/publication versus cleanup and renewal/release interleavings; assert retained-result safety, established lock order, exact-reference lifetime, no resurrection and no cross-scope effects; preserve typed labels and fail-closed release policy | NF cleanup race tests plus Reporting/GP owner assertions; combined lifetime mapping required | Baseline + post; preserve NF race and Reporting-owned rows; GP root/topology guards MUST remain intact |
| AC-JOB-01 | NFTR-JOB-001, NFTR-JOB-002 | Include each of the three positive statuses and every current terminal status; vary kind/profile independently; more than four matching jobs across multiple incidents with randomized insertion order MUST produce the complete ascending-ID set | Direct NF enumeration test not established by inspected manifests; Jobs package has `platform.jobs` accounting | Baseline + post via NF; new typed projection conformance post-only; Jobs/NF |
| AC-JOB-02 | NFTR-JOB-001, NFTR-JOB-002, NFTR-JOB-003 | Verify payload bytes are unchanged after query close and never Jobs-decoded; reject malformed/unknown-member/wrong-schema/invalid-ID/invalid-generation/empty-snapshot/incident-mismatch payloads at NF | Actual NF decoder inspected; direct persisted payload matrix required | Baseline + post NF path; typed byte-lifetime assertion post-only; NF/Jobs |
| AC-JOB-03 | NFTR-JOB-004 | Populate retry/cancellation/progress/timestamps and related receipts/proofs; reconcile and compare before/after, allowing only the two attempt columns; cover already-cleared attempts and nonresumable failure count | Existing Jobs mutation body inspected; current Recovery row is not sufficient direct proof | Baseline + post; Jobs/NF |
| AC-JOB-04 | NFTR-JOB-001, NFTR-JOB-002, NFTR-JOB-004 | Cause query/scan/iteration/cancellation failure and malformed later payload/later reconciliation failure; assert error with no successful partial collection/count and rollback of earlier writes; no writes before enumeration is consumed/closed | Direct failure-injection and transaction fixtures required | Baseline + post caller behavior; new required-argument/borrowed-handle contract post-only |
| AC-JOB-05 | NFTR-FRZ-001, NFTR-JOB-005 | Fail later Reporting reconciliation or GP postconditions; prove NF changes roll back with Graph, readiness remains unavailable and mandatory contribution cannot be omitted; preserve existing indeterminate-commit handling | Recovery assembly and GP restore tests anchor transaction behavior; direct NF state assertion required | Baseline + post; Recovery/GP/NF and Reporting as exercised |
| AC-JOB-06 | NFTR-JOB-002, NFTR-JOB-004 | Empty selection returns zero/no error; repeated reconciliation returns the selected count, including already-cleared rows, without changing other durable fields | Direct NF/Jobs repeat/empty assertions required | Baseline + post |
| AC-JOB-07 | NFTR-JOB-001, NFTR-JOB-005 | Exercise admitted quiescent restore with the existing transaction/isolation; prove the contribution does not open another connection, begin/commit/rollback, upgrade isolation, notify workers or bypass writer exclusion | App Recovery composition and GP owner contract inspected; explicit resource-ownership assertions required | Baseline caller boundary + post; new projection-specific construction conformance post-only |

All executable assertions above currently have status `TODO` for this refactor's baseline/post-change evidence. Existing source or test inspection MUST NOT be upgraded to `PASS`. Tests for a new GP capability or Jobs projection MUST use authored machine inputs/fixtures where required and MUST NOT parse the proposed contract out of this tracker.

### 8.4 Canonical command selections

| Validation layer | Command | Scope | Required before implementation? | Notes |
| ---------------- | ------- | ----- | ------------------------------- | ----- |
| unit | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.query_authoring_admission` | Strict query/filter authoring | yes, for query/error changes | Discovered, unexecuted. Body coverage is semantic grammar; add graph/receipt rows for their slices. |
| integration | `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.table_lifecycle_admission_replay` | Real HTTP/PG table lifecycle, authorization/replay | yes, for mutation changes | PostgreSQL/service fixtures required; unexecuted. |
| e2e/browser | `make browser-e2e-webserver-backed` | Browser consumers against composed server | no; required before completion when affected | Discovered, unexecuted. Select narrower appropriate NF rows through current task guide where supported; do not infer coverage from tab names. |
| generated drift | `make generate-drift` | Authored-input/generated-output parity | no; required after relevant implementation changes | Discovered, unexecuted. It may generate temporary/output artifacts, so excluded from this tracker-only session. No expected contract change in S-01/S-02. |
| import-boundary/static | `make frontend-import-boundary-check`; `make lint` | Frontend layering and repository static checks | no; run when relevant before completion | Discovered, unexecuted. GP's exact root/package guard is also routed by its owner. TODO: command/row discovery for any new NF-specific Go boundary guard; none inferred from a generic target name. |
| full check | `make check` | Broad implementation check | no | Discovered, unexecuted. Broaden only after narrow checks and if chosen slices justify it; run `make agent-finalize` beforehand in a later write-authorized session. |
| graph composition unit | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.bounded_graph_aggregation_and_selectors,module.networkflow.unit.time_bucket_graph_backend,module.networkflow.unit.graph_telemetry_boundary` | Streaming, temporal/cancel and telemetry seams | yes, for S-01 | Exact rows exist; unexecuted. Not a Core 05 benchmark claim. |
| saved graph integration | `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.saved_graph_lifecycle_v2,module.networkflow.integration.saved_graph_cutover_v6,module.networkflow.integration.time_bucket_saved_graph_lifecycle` | Current saved lifecycle, worker authority, retained compatibility and temporal graphs | yes, for S-01/S-02 as applicable | Exact rows exist; unexecuted. Keep schema suffixes and current contract-major posture distinct. |
| graph store integration | `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.store.graph_view_declaration_persistence` | Declaration persistence and selected-result generation/version behavior | yes, if Store/publication moves | PostgreSQL fixture; unexecuted. |
| cleanup race integration | `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.graph_result_cleanup,module.networkflow.integration.graph_result_cleanup_races` | Bounded retention and transaction races | yes, for lease/cleanup changes | Exact rows exist; unexecuted. Do not replace with unit mocks. |
| Reporting collaborator integration | `make service-backed-test-slice OWNER=module.reporting ROWS=module.reporting.integration.exact_graph_result_lease_lifecycle_8f1c5c43a2` | NF exact-result source, Reporting lease lifecycle | yes, for S-03 | Test package is NF; row owner is Reporting and collaborators are NF/GP. Unexecuted. |
| Recovery collaborator integration | `make service-backed-test-slice OWNER=module.recovery ROWS=module.recovery.integration.atomic_typed_terminal_evidence_7a41f0d8c2` | Recovery assembly's narrow reconciliation/atomic evidence | yes, for S-04; assess for S-01 | Exact row exists; unexecuted. AC-COMP-02 and AC-JOB-01–07 require direct NF assertions in addition to composed Recovery coverage. |
| Jobs restore projection | `make task-guide ROLE=module-author OWNER=platform.jobs` | Resolve authored Jobs coverage for the proposed read projection and preserved reconciliation | yes, for S-04 | Discovered owner manifest has 16 rows. TODO: select/add the actual projection test row after test design; no existing new-operation selector or successful run is asserted. |
| GP restore unit | `make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_deterministic_current_restore_rebuild` | Current native-v2 restore source/result constraints | yes, if restore composition changes | Exact row exists; unexecuted. GP PostgreSQL restore publication is a separate service-backed row. |
| frontend unit | `make test-slice OWNER=web.networkflow` | Saved-graph/table/indicator/controller/event continuity | no; required before completion when affected | 65 authored Vitest rows, unexecuted. Inspect relevant bodies before treating listed scenario names as exact assertions. |
| generated policy | `make generated-artifact-policy-check` | Generated roots/ownership rules | no | Discovered, unexecuted; applicable after later codegen/accounting changes. |
| harness isolation | `make test-slice OWNER=module.networkflow` | Includes owner-local Go control tests; current selector/runtime filtering must be checked | no; required if harness/composition changes | Prefer narrower control rows after task guide. TODO: command selection for process-tagged selector; do not assume default unit invocation exercises the packaged harness build. |
| Markdown documentation | `make lint-markdown` | Repository Markdown linter | no | Discovered but **skipped**: inspected recipe/script writes tool/cache artifacts. Literal tracker-only write restriction takes priority. Read-only structure, path, table and whitespace checks substitute only for this document audit, not a linter pass. |

Searches for validation used `Makefile`, authored task-surface inputs, the generated Make projection, the owner catalog, family manifests, verification owners, execution topology and the Markdown script. The initially guessed `tools/task_surface.json` was not a source; actual owner/manifest filenames are recorded in §1. No unsupported direct Go/pnpm/Vitest/Playwright invocation is proposed as canonical validation.

Retained-run maintenance was skipped because `RESULTS_DIR` is unset. `make agent-finalize` was not run: this is a single-document task with no broader product checks, and its maintenance/output writes are outside authorization. No preexisting successful run is being reused or claimed. No claim of generated cleanliness, test success, release readiness or browser conformance follows from this document audit.

## 9. Top-Level Work Tracker

`DONE` below denotes completed planning evidence unless an item explicitly says otherwise. Optional implementation candidates stay separate from the completed document task.

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| -- | --------- | ---------- | ------ | ---------- | -------------------- | -------------- |
| NF-001 | Bootstrap exact target/label/output and source hierarchy | WF-00 | DONE | None | §1; initial Git/path read | Safe derived label, target present, tracker-only authorization explicit. |
| NF-002 | Map every root/harness/test file and actual App caller | WF-01 | DONE | NF-001 | 88 individual §2 rows; read-only path-set comparison | Every target file accounted for once. |
| NF-003 | Reconcile framework omission with adopted NF/GP ownership | WF-02 | DONE | NF-002 | §3; F-01/F-02 | No relocation inferred from missing catalog entry or generic examples. |
| NF-004 | Freeze 19 routes and 16 cross-cutting contract groups | WF-02 | DONE | NF-003 | §4; live route catalog/handlers | Owners, test sources and gaps named for all discovered risks. |
| NF-005 | Distinguish real fixture tests from named local wrappers | WF-03 | DONE | NF-004 | C-01/C-03/C-16; F-13; family/body review | Parser/mock/reflection limitations recorded; no unsupported integration claim. |
| NF-006 | Diagnose query/graph HTTP errors and reusable composer seam | WF-04, WF-05 | DONE | NF-004, NF-005 | F-03/F-04; S-01 | Narrow same-owner candidate and error-preservation constraint recorded. |
| NF-007 | Map saved graph transactions, job finalization and historical receipts | WF-04, WF-05 | DONE | NF-004 | F-05/F-10/F-16; S-02 | Fresh/replay/publication order and atomic side effects explicit. |
| NF-008 | Document intentional private transactions and typed cross-owner capabilities | WF-04 | DONE | NF-004 | F-09/F-10; `transaction.go`, ports, participants | No generic platform transaction extraction proposed. |
| NF-009 | Record GP capability adoption and lease preservation baseline for specified S-03 design | WF-05 | BLOCKED | NF-020; G-OWNER-GP; G-EVID-S03 | NFTR-GP-001–006; §1.2 adoption placement; AC-GP-01–07 | Adopted GP permission and NF/Reporting preservation agreement plus actual baseline evidence; design prose alone cannot close this item. |
| NF-010 | Record Jobs/NF/Recovery read-contract adoption and restored-job baseline | WF-05 | BLOCKED | NF-020; G-OWNER-JOBS; G-EVID-S04 | NFTR-JOB-001–005; AC-JOB-01–07 | Owner adoption and direct/composed baseline evidence recorded; existing mutation semantics retained. |
| NF-011 | Redesign cross-owner cleanup health accounting | WF-05 | DEFERRED | Owner-reviewed bounded design | F-08 | No execution until current metric/snapshot semantics have a preservation design. |
| NF-012 | Assess harness isolation and frontend continuation contracts | WF-04 | DONE | NF-004 | F-12/F-14; C-05/C-13–15 | Build/guard separation and inspected frontend evidence distinguished from unverified hook/row coverage. |
| NF-013 | Sequence characterization and small structural candidates | WF-06 | DONE | NF-005, NF-006, NF-007 | §7; NFTR-GATE-001–004; baseline and production gates separated | All selected slices have noncircular dependencies, assertion mappings, commands, rollback and completion criteria. |
| NF-014 | Record generated contracts and cross-owner test accounting plan | WF-07 | DONE | NF-013 | §8; Reporting/NF/GP/Recovery and platform.jobs mapping | Current commands/rows verified; new selectors explicitly TODO; generated files unchanged and no successful run implied. |
| NF-015 | Execute future baseline and seam-specific characterization | WF-03 follow-through | TODO | G-AUTH-00; only selected baseline assertions | S-00; NFTR-EVID-001; §8.3 | Corresponding selected G-EVID gates pass; records identify real assertions and executed fixtures without waiting on unrelated slices. |
| NF-016 | Implement private source-composer candidate | WF-05/06 follow-through | TODO | G-AUTH-PROD names S-01; G-EVID-S01 | S-01; NFTR-COMP-001–003 | S-01 post-change assertions and completion criterion demonstrated. |
| NF-017 | Implement saved-graph command-file candidate | WF-05/06 follow-through | TODO | G-AUTH-PROD names S-02; G-EVID-S02 | S-02; NFTR-CMD-001–003 | S-02 post-change assertions and completion criterion demonstrated. |
| NF-018 | Shrink public Store/Service surface or replace all HTTP error types | WF-05 | DEFERRED | Compatibility/characterization and explicit narrow scope | F-02/F-04 | Separate later design; not bundled into initial extraction. |
| NF-019 | Audit creation-session tracker and preserve original handoffs | WF-08 | DONE | NF-014 | Original creation records in §10 | Historical completion retained; revision audit tracked separately as NF-025. |
| NF-020 | Specify the two GP capabilities and Jobs read projection without claiming adoption | WF-05 | DONE | NF-004, NF-005 | §4.2–4.3; concrete types, predicates, defaults, errors, side effects and ownership | Design ambiguities closed; adoption gates remain separately visible. |
| NF-021 | Define per-slice authorization/evidence gates and assertion-to-fixture mapping | WF-03, WF-06, WF-07 | DONE | NF-020 | NFTR-GATE-001–004, NFTR-EVID-001; §8.3 | No cyclic gate; real NF restore and lease/Jobs failure matrices explicit. |
| NF-022 | Implement specified GP lease capabilities and NF consumer wiring | WF-05/06 follow-through | BLOCKED | NF-009; G-AUTH-PROD names S-03 | S-03; no implementation performed | S-03 post-change conformance and preservation evidence passes. |
| NF-023 | Implement specified Jobs restore projection and NF consumer wiring | WF-05/06 follow-through | BLOCKED | NF-010; G-AUTH-PROD names S-04 | S-04; no implementation performed | S-04 direct/composed post-change evidence passes. |
| NF-024 | Execute selected final accounting and validation | WF-07 follow-through | TODO | Only selected production slices completed | S-05; §8 evidence contract | Selected coverage/accounting/drift/boundary results recorded without claiming unrun checks. |
| NF-025 | Audit NLSpec-voice revision and append current handoffs | WF-08 | DONE | NF-020, NF-021 | Tracker-only revision; staged blob and preserved row comparisons | §12 document audit passes; staging/history intact; implementation gates still honest. |

## 10. Session Handoff Log

The original creation-session narrative and rows below are retained as historical evidence. Rows marked `Codex NLSpec revision` describe the current revision; their presence MUST NOT imply owner adoption or executed product tests.

This tracker did not exist before this execution; there was no on-disk handoff history to overwrite. The first scope row retains the prior planning-turn context supplied by the user. All current-session times below use UTC on 2026-09-19. Files **touched** by this session are limited to `docs/handoffs/networkflow-module-refactor-tracker.md`; all other files mentioned are inspected read-only. Command summaries below describe discovery/audit, not product test runs.

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| Prior planning turn; exact time not retained | Prior plan, reported in user handoff | Planning-only findings prepared; tracker not written while Plan Mode was active | Framework/owners/repository inspected as reported; no writes reported | Prior handoff reports read-only Git, rg, cat, sed, wc and Python | Preserved as conversation history, not independent validation evidence | Remaining per-file mapping and tracker creation | Completed by current execution below. |
| 2026-09-19 | Codex tracker-creation session | Framework-first read, authority scoped; label/output and live commit verified | Framework, AGENTS, NF/GP/Core/domain/harness/research sections; touched tracker only | `cat`, `sed`, `rg`; `git status --short`, `git branch --show-current`, `git rev-parse HEAD`; path/count inspection | `main` at stated commit; initial tree clean; target present; tracker created | No document-completion blocker; RB-001 gates future code work | Revalidate HEAD/status at next session; preserve source hierarchy and single-file task history. |
| 2026-09-19 (revision) | Codex NLSpec revision | Tracker-only normative revision; source authority separated from proposed interfaces | Analysis notes, NLSpec guidance, NF/GP/Reporting/Core owner passages; touched tracker only | Read-only Git status/HEAD/index inspection; cat/sed/rg/Python; tracker editing only | NFTR vocabulary and owner-placement contracts defined; staged baseline captured | G-AUTH-00/G-AUTH-PROD remain TODO; owner gates remain BLOCKED | Record actual later authorization/adoption under §8.1; no implied promotion of proposals |
| 2026-09-19 (implementation complete) | Codex remediation S-05 | Accepted complete implementation scope; adopted owners and final verified completion recorded | Owner specs, framework, domain navigation, tracker and release handoff; full changed paths in §14 | Owner/source review; exact Make evidence in §14; Git basis/index/whitespace audit | Specification, implementation and verification complete; original history retained | No repository blocker | Maintainer release review and coordinated deployment |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| 2026-09-19 | Codex tracker-creation session | Valid NF ownership, mixed implementation responsibilities; small same-package candidates identified | All 88 target files; App callers, Jobs reconciliation, GP adapter/guard; touched tracker only | `rg --files`, scoped `rg -n`, `cat`/`sed`, Python source/import/symbol/test inspection and path-set comparison | 52 implementation + 36 tests; actual composer/command/peer-storage seams recorded | RB-002 lease capability; RB-003 Jobs read projection | Later authorize S-00, then choose S-01/S-02; do not execute gated peer changes. |
| 2026-09-19 (revision) | Codex NLSpec revision | Two GP capability contracts and one Jobs projection specified; NF semantics retained | Reporting source, GP lease writer, NF restore/worker/declaration source, Jobs reconciliation; touched tracker only | Exact source-body reads and owner/manifest inspection | Lookup/release asymmetry, renewal side effect, selected-job set and two-column reconciliation specified | G-OWNER-GP, G-OWNER-JOBS; corresponding execution evidence absent | Follow §4 contracts after adoption; preserve private composer/command seams and existing owner boundaries |
| 2026-09-19 (implementation complete) | Codex remediation S-05 | S-06/S-01/S-02/S-03/S-04/S-07/S-08 complete | NF semantic/composer/commands/facade; GP lease adapter; Jobs pages/index; App telemetry | Focused owner/partner slices; ordinary builds; boundary/lint; make check | Shared source composition, private application API, owner storage boundaries and bounded cleanup verified | None | Use retained public Module/dependency/contribution surface for future consumers |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| 2026-09-19 | Codex tracker-creation session | Client uses protocol adapter; grid uses semantic adapter; shared incident session owns socket state | Exact frontend files in §1; web/NF family manifests; touched tracker only | `rg`, `sed`, `cat`, Python manifest read | No supported frontend relocation; continuity and selector contracts frozen; manifest coverage not claimed as executed | No blocker to tracker; relevant controller test bodies still need review before relying on named rows | Preserve decoder/selection/attempt/poll behavior; choose web rows and browser scope after backend slice selection. |
| 2026-09-19 (revision) | Codex NLSpec revision | Existing consumer freeze retained; no frontend change authorized | Existing frontend inventory and C-05/C-13/C-14; touched tracker only | Tracker/manifest reads; no browser or frontend target executed | Captured attempt, historical receipt and selected-result continuity referenced by AC-CMD scenarios | Selected frontend assertion bodies/results still require actual implementation-session evidence | Use scoped web/controller assertions when chosen slices affect these consumers |
| 2026-09-19 (implementation complete) | Codex remediation S-05 | Intentional frontend architecture retained | NF client/controller/recovery/selection tests, stateful scenarios; no frontend production source changed | make frontend-typecheck; make frontend-import-boundary-check; final four-row browser selection; make check | Captured attempts, replay, selected results, refresh/retirement, authority withdrawal and pagination passed | None | Retain controller/adapter boundaries during future phase expansion |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| 2026-09-19 | Codex tracker-creation session | Nineteen generated operations, major 6/current semantic v2, distinct saved graph resources | NF index/routes/frontend entrypoint, generated catalog, generated policy, GP guard and projection tests; touched tracker only | Python JSON/source reads, `cat`, `sed`, `rg` | Routes and schema suffixes mapped; no generated files changed or drift claim made | None for documentation; any contract/major change requires later authorization | Preserve generated surfaces; use owner inputs and Make generators only in later authorized scope. |
| 2026-09-19 (revision) | Codex NLSpec revision | Design specification complete; new capability permission is not adopted | GP §8/§9, NF/Reporting mappings, Core Recovery and proposed typed Jobs interface; touched tracker only | Source/document/JSON reads; no generator or drift target run | Nineteen route rows preserved; no public contract, schema or generated change | Explicit GP amendment and Jobs/NF/Recovery adoption records absent | Capture owner adoption separately; keep pure GP root and authored/generated boundaries intact |
| 2026-09-19 (implementation complete) | Codex remediation S-05 | Owner amendments and projections agree; generated outputs refreshed | GP maintenance operations; OTel registry/schema v2; migration 44, owner mappings/manifests; generated GP/topology | make generate; make agent-finalize; make generate-drift; make generated-artifact-policy-check; make json-shape-check; make migration-drift | PASS; public NF major 6/state 4 unchanged; historical SQL unchanged | External migration and telemetry consumer cutover | Ship coordinated release per networkflow-remediation-release.md |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| 2026-09-19 | Codex tracker-creation session | Body/fixture versus label distinction documented; real Make selections discovered | Target tests, control files, harness server contribution, catalogs/manifests, Make/task surface/Markdown script; touched tracker only | Read-only `rg`, `sed`, `cat`, `wc`, Python JSON/count/table/path checks; `git diff --check` and final status review | No tests/Make targets run. Document inventory/structure audited; Markdown target skipped because it writes cache artifacts | RB-001 for implementation baseline; fault-hook and restore test completeness are not proved by rows | Run task guide and selected S-00 tests after authorization; retain Reporting/Recovery ownership; RESULTS_DIR maintenance skipped while unset. |
| 2026-09-19 (revision) | Codex NLSpec revision | Per-slice baseline and post-only acceptance requirements mapped; no execution evidence | Existing test anchors, platform.jobs and collaborator manifests; touched tracker only | Read-only Python document/row/requirement checks and Git whitespace/diff checks; results in §12 | Matrix requires real NF restore and controlled lease/job rollback scenarios; new selectors remain TODO | G-EVID-S01–S04 unexecuted; no run roots or passing product tests | After authorization, resolve actual symbols/fixtures and record NFTR-EVID-001 results; do not infer coverage from names |
| 2026-09-19 (implementation complete) | Codex remediation S-05 | Real-boundary assertion matrix and routing verified | NF/GP/Jobs/Recovery/Reporting/Imports/telemetry tests; authored manifests; Collaboration physical-failure support | Exact commands/run roots in §14; final make check 937/937; browser 13/13; make lint-markdown | All required gates pass; failed attempts retained; Reporting-owned NF test preserved | Retained-run maintenance skipped because RESULTS_DIR was unset | Retain final artifacts; no additional repository check required |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| 2026-09-19 | Codex tracker-creation session | Route/transaction/publication stages and cursor/replay bindings mapped | Core 04, NF replay/§28 owner text; routes/application/participants/worker/security/key-ring bodies and tests; touched tracker only | `sed`, `cat`, scoped `rg`, Python test-body inspection | No inferred deployment-admin bypass or weakened visibility; harness guard remains distinct from production authorization | No owner contradiction established; unexecuted runtime baseline remains RB-001 | Preserve exact fresh/replay/target-visibility/worker rechecks; inspect any owner conflict before adopting behavior change. |
| 2026-09-19 (revision) | Codex NLSpec revision | Lease retention remains separate from authorization/redaction; restore readiness atomic | Reporting lease/redaction owner passages; NF entrypoints; GP/Core restore constraints; touched tracker only | Focused read-only source/owner inspection | No new authorization grant, public restore/job surface, decoder unification or readiness bypass proposed | Relevant characterization remains unexecuted; no owner contradiction established | Preserve preflight/replay/error order and fail-closed boundaries; record conflicts under NFTR-GOV-002 |
| 2026-09-19 (implementation complete) | Codex remediation S-05 | Admission, replay, transaction/publication and disclosure boundaries retained | Semantic HTTP mapping, saved application, real auth/rollback/lease/restore fixtures and browser authority loss | Authenticated owner integrations; privacy/error and dependency guards; browser; make check | No weakened precedence, raw-cause disclosure, partial durable mutation or failed-restore readiness | None | Maintain separate owner authorization and lease/redaction responsibilities |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| 2026-09-19 | Codex tracker-creation session | Tracker complete; implementation intentionally unperformed | Tracker plus read-only evidence in §§1–2 | Final Python inventory/table/dependency/row/path audit; Git diff/status/whitespace checks | Sole changed path is tracker; no retained test success or codegen claim | RB-001 baseline; RB-002/RB-003 gate only their candidate slices | Next authorized agent: verify checkout, read this tracker and applicable owners, run task guide, select S-00 plus a small seam; append findings and real run roots here. |
| 2026-09-19 (revision) | Codex NLSpec revision | Design gaps resolved in tracker; adoption, evidence and implementation tracked separately | Revised tracker and current baseline metadata; touched tracker only | Read-only inventory/route/history/index/dependency/assertion checks and Git diff review | Current revision audit is recorded in §12; staging and historical rows are preservation obligations | RB-001 authorization/evidence; RB-002 GP adoption/evidence; RB-003 Jobs adoption/evidence | Obtain a later selected-slice task, begin S-00 only after its authorization, and append real gate/adoption/run evidence |
| 2026-09-19 (implementation complete) | Codex remediation S-05 | S-05 complete; no implementation or validation TODO remains | Final tracker, release handoff, additive index and registry-v2 migration instructions | Final make check at 20260919T071101Z-p54535; Markdown at 20260919T071154Z-p99134; final diff audit | Repository worktree is reviewable and uncommitted; original staging preserved | Deployment, maintainer release review and dashboard/alert cutover are external | Apply coordinated release; rollback pairs application and telemetry configuration; index may remain |

## 11. Open Questions and Blockers

The design questions behind RB-001–RB-003 are specified by §§4/8. Remaining blockers are authorization, owner adoption and executed evidence, not an invitation to choose a different interface during implementation. They MUST NOT be marked closed merely because the design is written. No owner contradiction was established; NFTR-GOV-002 defines the required handling if one is found.

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| -- | ------------------- | -------------- | ---------------------------- | -------------- |
| RB-001 | Separated authorization and selected-slice evidence gates remain unsatisfied | Starting S-00 requires permission, while production readiness requires the evidence S-00 produces; conflating them creates a cycle or a false pass | G-AUTH-00 to start S-00; G-AUTH-PROD for each selected production slice; corresponding G-EVID gate and NFTR-EVID-001 records. No unselected slice is a dependency. | TODO — sequencing design resolved; no S-00/production authorization or executed baseline recorded |
| RB-002 | Specified GP lookup/scoped-release contract awaits adoption and baseline evidence | Current GP §8 excludes unlisted capabilities; exact lease/error/side-effect semantics require more than the existing one-result test | G-OWNER-GP plus G-EVID-S03 for NFTR-GP-001–006; explicit permission and NF/Reporting preservation agreement, scoped expiry/rollback/race assertions. S-03 implementation/conformance remains separate. | BLOCKED — design resolved by §4.2; adoption and execution pending for S-03 only |
| RB-003 | Specified Jobs read projection and source/restore responsibilities await adoption and direct/composed baseline evidence | The typed read boundary must preserve selected rows, payload bytes, two-column mutation and whole-restore rollback/readiness | G-OWNER-JOBS plus G-EVID-S04 for NFTR-JOB-001–005; direct NF/Jobs and composed Recovery assertions. S-04 implementation/conformance remains separate. | BLOCKED — design resolved by §4.3; adoption and execution pending for S-04 only |

RB-001 closure MUST name the selected slice set and record authorization and evidence independently. RB-002/RB-003 closure MUST identify both adopted owner provisions and executed baseline evidence. Their closure permits the affected separately authorized production slice to start; it MUST NOT set the implementation work item to DONE. Any changed owner decision MUST update the single contract definition, its assertion mappings and affected gates together.

## 12. Binary Completion Criteria

The document audit, slice-readiness assessment and implementation assessment are distinct. A PASS in the first table MUST NOT be used as a PASS in either later table. The read-only revision audit passed; historical creation-session results remain in §10. No owner adoption or product-test success is asserted.

| Criterion | Pass/fail | Evidence |
| --------- | --------- | -------- |
| Every target file is inventoried or explicitly excluded | PASS | Verified by revision audit: §2 MUST match all 88 live files and preserve the creation inventory entries; no exclusion. |
| Every discovered public contract risk has an owner and test posture | PASS | Verified by revision audit: nineteen route rows MUST remain intact; C-01–C-16 and all NFTR preservation requirements MUST have scoped owners and mapped assertions or an explicit unrelated TODO. |
| Every workflow and slice has closed dependencies and exit criteria | PASS | Verified by revision audit: WF-00–08 graph MUST remain reciprocal; S-00 MUST depend on authorization rather than its own output; S-01/S-02 and unselected peer slices MUST remain independent as specified. |
| Every slice preserves behavior or marks changed behavior `requires later authorization` | PASS | Verified by revision audit: NFTR-FRZ-001 and applicable requirement IDs MUST govern every S-00–S-05 row; new interfaces MUST remain proposals pending owner adoption. |
| Commands, evidence phases and acceptance cases are unambiguous | PASS | Verified by revision audit: existing Make/row references MUST resolve; new Jobs/GP selectors MUST NOT be invented; every requirement MUST map to assertions and every executable assertion MUST distinguish baseline/post-only obligations. |
| Authority and contradiction handling are explicit | PASS | Verified by revision audit: definitions/owner placement MUST separate adoption, observations, proposals and evidence; no owner contradiction established; any future contradiction MUST use `BLOCKED: owner contradiction`. |
| Repository/framework mismatches remain recorded | PASS | Verified by revision audit: F-01/F-02 MUST remain, and the added proposed capabilities MUST NOT be represented as current GP/Jobs APIs. |
| Handoff history and restart information are current | PASS | Verified by revision audit: all original handoff rows MUST remain with revision entries appended to all seven tables; current gates and exact staged/repository basis MUST be recorded. |
| Only authorized tracker revision written and staging unchanged | PASS | Verified by revision audit: twelve required sections/table columns, unchanged staged blob/index, tracker-only worktree delta relative to session start, and clean document whitespace MUST be demonstrated. |

| Readiness assessment | Pass/fail | Evidence |
| -------------------- | --------- | -------- |
| S-00 authorized to start | FAIL | G-AUTH-00 is TODO; current authorization covers this tracker only. |
| S-01/S-02 ready for production changes | FAIL | G-AUTH-PROD and their own baseline evidence gates are TODO. Their design and independent sequencing are specified. |
| S-03 ready for production changes | FAIL | G-OWNER-GP is BLOCKED and G-EVID-S03/G-AUTH-PROD remain TODO. The two-capability design is specified. |
| S-04 ready for production changes | FAIL | G-OWNER-JOBS is BLOCKED and G-EVID-S04/G-AUTH-PROD remain TODO. The Jobs read design is specified. |

| Implementation assessment | Pass/fail | Evidence |
| ------------------------- | --------- | -------- |
| Selected production refactor complete | FAIL | No production slice has been authorized or implemented by this documentation task. |
| Post-change conformance/accounting complete | FAIL | No product test or Make validation target was run; S-05 and all executable evidence remain outstanding. |

The document audit MUST use read-only inventory/header/table/status/identifier/dependency/owner-row/reference checks, preserved-row comparisons, staged-blob/index comparisons, `git diff --check` and Git status/diff inspection with optional index writes disabled. It MUST NOT execute product tests or cache-writing Markdown validation. These checks do not replace generated drift, import-boundary, browser or conformance verification. Revision audit result: PASS for all twelve sections and required table columns, 88 unchanged inventory entries, nineteen unchanged route rows, retained historical handoff rows plus seven revision entries, 27 uniquely defined requirements mapped to 21 acceptance scenarios, eight valid gates, reciprocal workflow dependencies, current Make/owner-row references and document whitespace. Only the tracker has an unstaged delta; its staged blob and index are unchanged. No product tests, Make targets, owner amendments or production refactor were performed.

## 13. Remediation execution ledger

### Accepted owner decisions

The implementation request adopts GP §8.0 lookup/scoped release, Core 01
§3.3.9.3 paged restored jobs, NF §29 application/consumer boundaries, and revised
NF-REQ-195/196 plus OTEL-REQ-152 and metric rows. Reporting lifetime/redaction
and Core 01 REQ-01-625A remain unchanged. Machine/runtime realization is recorded
per implementing slice; owner amendment alone is not conformance evidence.

The historical NFTR-JOB-001/002 interface and full-collection-before-write rule
are superseded by pages of 256 in the same quiescent transaction, without a
total cap. S-02 now requires a separate typed application component, not a
Service-receiver file move. S-06–08 close F-04, F-08 and F-02 respectively.
G-AUTH-00 and G-AUTH-PROD are satisfied by the explicit implementation request.
G-OWNER-GP and G-OWNER-JOBS are satisfied by the amendments above; their evidence
gates are satisfied by the completed slice records. F-01 is corrected in the framework. F-09–12 and F-14–17
retain their intentional boundaries and require final regression evidence.

| Workstream | Status | Changes and evidence | Exit / next action |
| --- | --- | --- | --- |
| Specification closure | DONE | GP §8.0; Core 01 §3.3.9.3; NF §29 and NF-REQ-195/196; OTel metrics/OTEL-REQ-152; framework owner catalog. Manual source/owner review; no product claims. | Begin S-00; implement machine/runtime projections in their owning slices. |
| S-00 | DONE | Real HTTP/worker/restore, transaction and lease baselines passed; records below. | Begin S-06. |
| S-06 | DONE | Transport-independent semantic failures. | Semantic/error boundary evidence. |
| S-01 | DONE | Shared source composer. | Three real consumer paths. |
| S-02 | DONE | Typed saved-graph application. | Replay/atomicity/notification evidence. |
| S-03 | DONE | GP lookup and scoped release. | Scope, expiry, rollback and race evidence. |
| S-04 | DONE | Paged Jobs projection and supporting index. | Page boundaries and atomic restore evidence. |
| S-07 | DONE | Bounded cleanup progress/freshness. | Signal migration and cancellation evidence. |
| S-08 | DONE | Production API narrowing. | Caller/build/API guard evidence. |
| S-05 | VERIFIED COMPLETE | All required validation passed; assertion/work-item closure, changed-file inventory and all seven handoffs updated. | Repository effort complete; coordinated deployment/telemetry cutover remains external. |

Initial basis: HEAD `b9e182b710f6cbfa9d56b656f3a8558012f75044`; tracker staged
as added, no other initial changes. `make task-guide ROLE=module-author
OWNER=module.networkflow` and `make help-all` passed. At ledger creation no baseline execution had passed; completed evidence follows.

### S-00 completion — baseline evidence

Preservation baseline runs (all commands from repository root):

| Evidence | Command selection | Result / run root |
| --- | --- | --- |
| B-01 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.query_authoring_admission,module.networkflow.unit.bounded_graph_aggregation_and_selectors,module.networkflow.unit.time_bucket_graph_backend,module.networkflow.unit.graph_telemetry_boundary,module.networkflow.unit.saved_graph_receipt_integrity,module.networkflow.unit.transaction_lifecycle` | PASS; `.cartulary/test-results/20260919T053510Z-p42189` |
| B-02 | `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.saved_graph_lifecycle_v2,module.networkflow.integration.saved_graph_cutover_v6,module.networkflow.integration.time_bucket_saved_graph_lifecycle,module.networkflow.integration.graph_result_cleanup_races,module.networkflow.integration.table_lifecycle_admission_replay,module.networkflow.store.graph_view_declaration_persistence` | PASS; `.cartulary/test-results/20260919T053640Z-p47367` |
| B-03 | `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.resource_change_intent_replay_and_privacy_f1a2b3c4d5,module.networkflow.integration.resource_intent_failure_rolls_back_source_mutation_8a4c3d2e1f` | PASS; `.cartulary/test-results/20260919T053802Z-p82685` |
| B-04 | `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.saved_graph_lifecycle_v2` | PASS with added 0/1/256/257/512/513 restore sets, exact two-column/repeat/rollback assertions; `.cartulary/test-results/20260919T053947Z-p5174` |
| B-05 | `make service-backed-test-slice OWNER=module.reporting ROWS=module.reporting.integration.exact_graph_result_lease_lifecycle_8f1c5c43a2` | PASS with stored-ID, expiry, shorter renewal, committed renewal on binding failure, rollback, 1003-result complete release and all neighboring scope dimensions; `.cartulary/test-results/20260919T054118Z-p40169` |

B-05 initially failed at `.cartulary/test-results/20260919T053947Z-p5183`:
the new fixture shortened expiry below its copied renewal timestamp. This was
a test-fixture defect, corrected by setting the copied renewal to creation time;
no product change was made. `make format` passed. Original staging is unchanged.

Actual lifecycle fixtures call authenticated HTTP, run the materialization
handler, and invoke the composed Recovery participant over actual selected NF
source state. The restoration asserts result/vertex/edge identity, NF attempt
clearing and Reporting lease reconciliation. Parser/mock/reflection wrappers
remain unit evidence only. The fault registry has no production consumer; its
control tests are not used as product rollback proof. Intent rollback uses a
real failing participant with PostgreSQL assertions. New adapter construction,
paging argument/resource ownership and new telemetry shapes remain post-change
acceptance obligations. Existing historical §8 TODO entries are not passing
records; these run roots identify the executed baseline assertions.

Fixture construction hashes (SHA-256; controlled times are in the tests; backing
services and executed symbols are retained in each run's manifests/row JSON):

- `internal/modules/networkflow/routes_integration_test.go`: `ad2c6dda6de0f291ef174a7ed36c619ba59d247ad597af4e9514c7dc20bfa3d7`.
- `internal/modules/networkflow/reporting_graph_source_integration_test.go`: `ebcb34968320990d5a34b107df51fac6b5ca28e44e0709ff0550a943f66a7095`.
- `internal/modules/networkflow/graph_result_cleanup_race_integration_test.go`: `68a6f796e78ff336b8acd721987313e053fd519552438fa4ba4f1a2c0b6103df`.
- `internal/modules/networkflow/table_lifecycle_integration_test.go`: `5e96c748d2dae6712cfa9ec012dcb55985775d141f4e40875ffb32ea0337c98b`.

### S-06 completion — semantic failures

Reusable query/filter, temporal, graph, response reconstruction and provider
logic now return `semanticFailure` with typed details and wrapped causes.
Mapping uses `strictjson` directly. `semantic_http_errors.go` owns closed
status/detail translation; HTTP entry adapters contain no semantic logic, and
adapters used only by existing wire assertions compile only in tests.
Telemetry classifies owner failures without status codes. The new semantic
boundary assertion checks imports, cancellation causes and unknown-kind privacy.

Post-change unit selection B-01's query/graph/temporal/telemetry rows passed at
`.cartulary/test-results/20260919T054729Z-p64358`. Real saved-graph, temporal and
table admission/replay rows passed at
`.cartulary/test-results/20260919T054744Z-p65156`. Intermediate compile failures
at `20260919T054533Z-p62923` and `20260919T054640Z-p63716` were refactor issues
(import-edit recovery and a variadic adapter call), fixed before these passes.
Public error assertions remain unchanged in meaning; no public/data migration.
Next: S-01. Final validation also compiles the test-only adapter split.

### S-01 completion — shared composition

`graph_source_composer.go` now owns source reading, limits, projection, clock and
observation dependencies. HTTP and workers share configured composition; restore
constructs canonical-default composition without an observer. Neither worker nor
restore constructs `Service`. Selected-result reconstruction deliberately disables
observation on its private composer copy. No identity, storage or wire migration.

`make format` passed at `20260919T055035Z-p83600`.
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.query_authoring_admission,module.networkflow.unit.bounded_graph_aggregation_and_selectors,module.networkflow.unit.time_bucket_graph_backend,module.networkflow.unit.graph_telemetry_boundary`
passed at `.cartulary/test-results/20260919T055046Z-p87932`.
`make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.saved_graph_lifecycle_v2,module.networkflow.integration.time_bucket_saved_graph_lifecycle`
passed at `.cartulary/test-results/20260919T055047Z-p88153`; these fixtures execute
HTTP, workers and real restore. No failed S-01 runs. Begin S-02.

### S-02 completion — saved-graph commands

`saved_graph_application.go` owns typed command inputs/outcomes, pre-replay
admission and transaction-time rechecks, declaration/job/audit participation and
post-commit notification. `saved_graph_receipts.go` owns durable response encoding
and historical replay. Route handlers decode, map failures and serialize outcomes;
no command depends on `Service` or imports HTTP. Table/link replay is unchanged.
The application explicitly repeats admission for independent consumers; routes
retain their earlier check before decoding. Receipt read failures remain internal.

`make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.saved_graph_receipt_integrity,module.networkflow.unit.query_authoring_admission`
passed at `.cartulary/test-results/20260919T055709Z-p12832`.
`make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.saved_graph_lifecycle_v2,module.networkflow.integration.saved_graph_cutover_v6,module.networkflow.integration.time_bucket_saved_graph_lifecycle,module.networkflow.integration.resource_intent_failure_rolls_back_source_mutation_8a4c3d2e1f`
passed at `.cartulary/test-results/20260919T055726Z-p13707`.
Added real HTTP rollback assertions at declaration, Collaboration intent, Jobs,
receipt and administrative-audit INSERTs. The lifecycle-only command passed at
`.cartulary/test-results/20260919T060129Z-p76595`.
Fixture failures at `20260919T055855Z-p36299` (DDL attempted with restricted runtime
role) and `20260919T060034Z-p58722` (an unrelated audit-projection table was not a
participant) were corrected: fixture DDL uses the migration handle, and assertions
cover the actual five participants. Historical receipt bytes and retirement 204
remain unchanged. Frontend continuity is reserved for S-05. Begin S-03.

### S-03 completion — Graph-owned leases

`postgresresult/lease_scope.go` adds exact lookup and a borrowed-transaction
scoped releaser. NF's Reporting adapter contains no lease lookup/deletion SQL.
Lookup deliberately preserves missing-match precedence for malformed result keys;
renewal and exact binding validation remain separate later stages. Release has no
batch cap and validates all four required scope dimensions. Added typed machine
operations to `contracts/graph-projection/storage-maintenance.v1.json`; regenerated
`internal/gen/contractgraphprojection/artifacts_gen.go` using `make generate-artifacts`.

`make service-backed-test-slice OWNER=module.reporting ROWS=module.reporting.integration.exact_graph_result_lease_lifecycle_8f1c5c43a2`
passed at `.cartulary/test-results/20260919T060358Z-p99070`.
`make service-backed-test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.result_v2_atomicity_and_queries,module.graphprojection.storage.result_v2_cleanup_locking`
passed at `.cartulary/test-results/20260919T060512Z-p24927`; direct assertions cover
construction, strict expiry, cancellation, stored identity, closed transaction,
idempotent release and rollback. `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.graph_result_cleanup_races`
passed at `.cartulary/test-results/20260919T060512Z-p24930`.
`make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.v2_contract_projection`
passed at `.cartulary/test-results/20260919T060603Z-p60813` after adding projection
checks for both capabilities. No failed S-03 runs or data migration. Begin S-04.

### S-04 completion — bounded Jobs restore pages

`jobs/restore_enumeration.go` owns exact positive-status selection, exclusive
UUID ordering, page bounds and owned payload bytes. Nullable incident IDs retain
Jobs' generic deployment scope; NF rejects missing/mismatched incident ownership.
NF consumes closed pages of 256 before reconciliation and returns no partial count
on failure. It neither opens another transaction nor changes readiness handling.
The cursor predicate is a direct index range on subsequent pages. Migration
`00044_jobs_restore_enumeration.sql` adds the partial owner/kind/job-ID index;
history and schema-object manifests include the additive index. No old migration,
job payload, receipt or public/state version changes.

`make test-slice OWNER=platform.jobs ROWS=platform.jobs.unit.typed_execution_contract`
passed at `.cartulary/test-results/20260919T061005Z-p71712`, including required
arguments, cancellation, owned reused bytes, closed rows and query/scan/iteration
failure with no partial page. `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.saved_graph_lifecycle_v2`
passed at `.cartulary/test-results/20260919T061005Z-p71710`: 0/1/256/257/512/513,
positive statuses, exact two-column changes, repeated reconciliation, late-page
invalid payload/nonresumable state and rollback of earlier-page writes, plus real
Recovery reconstruction. `make migration-drift` passed at
`.cartulary/test-results/20260919T060910Z-p63192`. `make format` passed at
`20260919T060958Z-p67341`. Initial unit/service/migration/format dispatch was rejected
before execution because the new Jobs selector was not ASCII-sorted; fixed in
the authored manifest. Added actual Recovery, Reporting, Collaboration and audit
collaborators to the NF lifecycle row. Final topology regeneration is S-05.
Begin S-07; final whole-restore failure/readiness evidence remains part of S-05.

### S-07 completion — bounded cleanup observation

Removed NF's cross-owner health scan and retired both backlog/oldest-result gauges.
Cleanup applies its deadline before lease deletion and passes cancellation through
all database operations, including rollback. Committed examinations include retained
candidates; later failures retain confirmed counts and indeterminate commits add
none. The dispatcher publishes its actual continuation/restart decision. Application
telemetry retains that decision across failures and computes live monotonic freshness
without database access. Registry v2, its schema/attachment, runtime registry,
OTel checker and corpus 015/019 agree. No compatibility alias or dual emission.
The coordinated cutover/rollback instructions are in
[the release handoff](networkflow-remediation-release.md).

`make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.cleanup_dispatcher_lifecycle,module.networkflow.unit.graph_telemetry_boundary`
passed at `.cartulary/test-results/20260919T061813Z-p31469`: cancellation of a stalled
lease query, normal/continued/restarted/failed decisions, committed progress, and
observer panic containment. `make test-slice OWNER=app.server ROWS=app.server.unit.evidence_cleanup_activation_and_telemetry_a1b2c3d4e5`
passed at `.cartulary/test-results/20260919T061524Z-p96240`, including controlled-clock
freshness before/after success, aging through failure, empty sweeps, privacy and
retired metric absence. `make test-slice OWNER=platform.telemetry` passed at
`.cartulary/test-results/20260919T061524Z-p96257`.
`make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.graph_result_cleanup,module.networkflow.integration.graph_result_cleanup_races`
passed at `.cartulary/test-results/20260919T061714Z-p10712`: retained/deleted/bounded
sweeps, confirmed progress before commit failure, indeterminate commit and lock races.
`make otel-conformance` passed at `.cartulary/test-results/20260919T061858Z-p37174`.
Failures fixed before completion: unit `20260919T061714Z-p10699` used a Query stub
where the real lease operation uses QueryRow; conformance `20260919T061713Z-p10559`
found the old corpus-015 metric list; `20260919T061813Z-p31375` rejected concurrent
formatting after its frontend input snapshot. Final formatting precedes verification.
External dashboard/deployment cutover remains pending, not repository implementation.
Begin S-08.

### S-08 completion — production facade

Production caller/signature closure retains 59 package-level declarations and their
capability methods. `store`, `routeService`, persistence records, parsers, identity
helpers and unused constructors are private. `production_surface_test.go` positively
checks every retained declaration/method. `integration_surface_test.go` supplies 43
names only to external NF tests, preserving application import boundaries. Default
registry construction, legacy default-limit convenience, lifecycle enumeration and
limit mutation conveniences now exist only in test builds. Workbook tests use the
explicit key registry; bundle tests assert literal state-family identities. No
production compatibility aliases or HTTP/persisted migrations were introduced.
The semantic dependency guard is now a positive import allowlist.

`make build-server` passed at `.cartulary/test-results/20260919T062722Z-p2499`
(and `20260919T062551Z-p86202`), proving ordinary production compilation without
bridges. `make test-slice OWNER=module.networkflow` executed all owner selections
at `.cartulary/test-results/20260919T062722Z-p2183`: 42/43 units passed, including
actual saved/temporal/cleanup integration and all selected frontend/browser rows.
Its sole failed unit was the new dependency allowlist omitting the existing single
`math` import. After correction, `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.query_authoring_admission`
passed at `.cartulary/test-results/20260919T062953Z-p77303`, including both positive
guards. The full run is retained as row evidence, not labeled an overall pass.

Earlier rename compile failures at `20260919T062315Z-p43549`,
`20260919T062317Z-p43891`, and `20260919T062406Z-p79890` exposed variadic/constant
references, an anonymous JSON field and a local filter/type collision; fixed before
the passing build and row evidence. Separate service selection
`20260919T062722Z-p2199` failed during shared service-image stamp warming while the
full owner run was active; it ran no product assertion. The full owner run's same
service rows subsequently passed. `make format` passed at `20260919T062702Z-p97204`.
Begin S-05: close final assertion accounting, regenerate authored projections,
run finalize then required final checks, and complete all seven handoff tables.


### S-05 execution record — real boundary closure before final gates

Added Imports-owned replay and failing-outcome assertions to
`imports_integration_test.go`. A nontransactional sequence confirms that the
actual failing Imports participant observed NF rows before rollback; durable NF
rows/tables and successful outcomes remain absent afterward. `make
service-backed-test-slice OWNER=module.imports
ROWS=module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0`
passed at `.cartulary/test-results/20260919T064408Z-p46946`.

Expanded NF lifecycle/authority fixtures cover real restore composition in both
aggregation modes; HTTP/worker semantic equality; canonical restore defaults
versus a restricted worker; out-of-order active declarations, inactive/unselected
exclusion, both snapshot mismatches; 0/1/256/257/512/513 jobs inserted in reverse
order across two incidents, every status and independent kind/profile exclusions;
late-page malformed/unknown/schema/identity/generation/snapshot/incident errors;
and real Reporting failure after NF reconciliation, Graph rollback and no readiness.
`make service-backed-test-slice OWNER=module.networkflow
ROWS=module.networkflow.integration.saved_graph_lifecycle_v2,module.networkflow.integration.time_bucket_saved_graph_lifecycle`
passed at `.cartulary/test-results/20260919T064518Z-p67303`.

Initial expanded-fixture failures are retained: `20260919T063830Z-p91025` attempted
an interface composite literal; `20260919T064021Z-p14358` omitted the restricted
worker's required intent appender, used expired execution leases that let live
workers mutate snapshots, and injected failure into Jobs payload rather than the
Reporting-owned restore payload. Corrected fixture construction, future attempt
leases, and the actual `reporting_job_payloads` consumer. These were test defects,
not passing product assertions. Named the closed private semantic failure kinds
and added a no-I/O composer-construction assertion. Narrow semantic/telemetry units
passed at `20260919T064022Z-p14581` and `20260919T064410Z-p47185`.

Current owner routing was resolved through `make explain-test-owner OWNER=<owner>`
for NF, GP, Jobs, Recovery, Reporting, Imports and telemetry. Finalization and broad
checks remain pending. A final command notification fixture additionally captures
commit visibility and suppresses notifications on rollback and exact replay.


## 14. Current completion and release handoff

This section supersedes historical completion assessments and current-status
columns in §§5–12, without deleting their evidence. Specification adoption, all production slices and S-05 repository verification
are complete. All required assertions and final gates below passed. Coordinated
deployment and operator telemetry cutover remain external steps.

### Finding dispositions

| Finding | Final disposition and affected areas | Rationale and durable benefit | Compatibility and unresolved-risk disposition | Acceptance |
| --- | --- | --- | --- | --- |
| F-01 | REPLACE framework omissions; specifications/documentation | Explicit NF/GP owner navigation prevents misplaced future capabilities | No runtime change; ambiguous ownership removed | Adopted owner references in §13 |
| F-02 | REPLACE broad production API; implementation/tests | Private persistence and transport types prevent accidental downstream contracts | Coordinated internal Go break; callers and test bridges updated | Positive surface guard; ordinary builds |
| F-03 | REPLACE Service reuse with source composer; implementation/tests | Cohesive composition supports additional consumers and explicit policies | No identity/state change; differing worker/restore limits retained | AC-COMP-01–03 |
| F-04 | REPLACE HTTP-dependent semantic failures; implementation/tests/specification | Closed owner failures allow non-HTTP consumers without transport imports | Owner-defined HTTP precedence/details retained; unknown causes remain private | Semantic guard/error matrix; cancellation/outcome tests |
| F-05 | REPLACE route-owned commands; implementation/tests/docs | Typed application commands isolate admission, durable work and notification | Historical receipt bytes, no-op rename and empty retirement preserved | AC-CMD-01–03 |
| F-06 | REPLACE NF lease SQL with GP capabilities; specification/contracts/implementation/tests | Schema ownership and exact narrow scopes support independent GP evolution | No data migration; release scope/expiry/renewal ordering retained | AC-GP-01–07 |
| F-07 | REPLACE global Jobs enumeration with owner pages; specification/implementation/tests/migration | Bounded memory supports deployment-wide growth within one atomic transaction | Additive migration 44; no payload/status migration or total cap | AC-JOB-01–07 |
| F-08 | REMOVE global health scan/old gauges; REPLACE with bounded signals in all areas | Constant-space observation cannot undermine bounded cleanup | Deliberate telemetry break; coordinated cutover required to avoid monitoring gaps | Controlled-clock, cancellation, committed-progress and OTel corpus evidence |
| F-09 | RETAIN private NF transaction helper | Explicit transaction ownership is clearer than a generic platform facade | No compatibility cost added; rollback semantics retained | Transaction lifecycle and real participant failures |
| F-10 | RETAIN atomic cross-owner participants | One transaction protects source mutations and owner effects | No split commit or weakened replay guarantee | Imports apply/replay/rollback; link/command rollback |
| F-11 | RETAIN NF intent, App translation and Collaboration delivery | Each owner keeps its cohesive responsibility | No new delivery dependency in NF | Intent privacy/replay and real outbox failure assertions |
| F-12 | RETAIN isolated harness mechanics; correct evidence classification | Optional controls remain distinct from exercised product paths | Registry arming is never claimed as rollback evidence | Disabled-route/build guards; real failing participants |
| F-13 | REPLACE overstated assurance with assertion-level accounting; tests/docs | Real-boundary evidence exposes privacy, partial commits and restore defects | Useful parser/mock/reflection tests remain unit evidence | S-00 plus S-05 Imports/admission/restore fixtures |
| F-14 | RETAIN immutable analytical rows and frontend adapters | Avoid speculative Core/view/grid coupling | Existing selection, read-only rows and captured attempts remain | Frontend tests, typecheck, import guard, browser |
| F-15 | RETAIN authored/generated separation; regenerate projections | Repeatable generation prevents manual artifact drift | Public major 6/state 4 remain; registry v2 is the intentional break | Generate/drift/policy/schema/migration gates |
| F-16 | RETAIN admission, replay, transaction and publication stages | Security checks remain at the races and disclosures they protect | No authorization bypass or replay-algorithm unification | Authenticated lifecycle, publication-role and frontend denied-replay tests |
| F-17 | RETAIN NF identity/name/digest/time semantics | Source ownership keeps immutable meanings coherent | No identifier, timezone, digest or naming migration | NF identity/temporal/parser/store tests and exact restore |

RB-001's execution authorization is satisfied once by the implementation request;
RB-002's GP permission and RB-003's Jobs/NF/Recovery contract are adopted in §13.
Their implementation/evidence gates are satisfied by the slice records. Final
release verification remains separate; no previous failed run is relabeled.

### Current assertion matrix

| Assertion | Concrete post-change evidence | Record |
| --- | --- | --- |
| AC-COMP-01 | Real HTTP, worker and NF restore contribution; default/temporal digests, binding/object identity, bounded scans | S-01 and S-05 lifecycle/temporal runs |
| AC-COMP-02 | Restricted configured worker rejects; canonical-default restore reconstructs; reverse-inserted selected declarations sort; inactive/unselected excluded; both snapshot mismatches reject | S-05 authority and source-composition helpers |
| AC-COMP-03 | Scan cancellation, worker deadline/no publication, observer panic containment, constructor with panic-on-I/O reader | NF streaming, deadline, telemetry and authority rows |
| AC-CMD-01 | Current admission before replay, transaction/publication role checks, historical receipt and denied-attempt continuity | Saved lifecycle/cutover/receipt/authority; frontend replay rows |
| AC-CMD-02 | Five real durable participant failures, commit failure, post-commit visible notification, no replay notification, no-op rename and stale publication | S-02 and S-05 lifecycle/authority; declaration store row |
| AC-CMD-03 | Empty 204/replay/retirement events and exact selected/captured-attempt continuity | Saved lifecycle; frontend captured/reassessment/retirement rows; browser |
| AC-GP-01 | Borrowed handles, missing dependencies, read cancellation, closed transaction and adapter-only operations | S-03 GP storage/contract and Reporting rows |
| AC-GP-02 | Strict expiry, renewal interval precedence and shortening | Reporting exact-result lease row; GP storage |
| AC-GP-03 | Non-derived stored lease ID retained through lookup/renewal | Reporting and GP storage assertions |
| AC-GP-04 | Exact leased binding after selection changes; failed later read returns no content while renewal persists | Reporting exact-result lease row |
| AC-GP-05 | 1,003 matching results, expired/live leases, all neighboring scopes, empty repeat | Reporting exact-result lease row |
| AC-GP-06 | Visible transactional deletion rolls back; closed/canceled execution fails without ownership transfer | Reporting and GP storage rows |
| AC-GP-07 | Controlled cleanup/acquisition/publication locks and exact reader lifetime; no resurrection or cross-scope release | NF cleanup-race, GP cleanup-locking, Reporting lease rows |
| AC-JOB-01 | 0/1/256/257/512/513, two incidents, reverse insertion, exact ascending pages, all six statuses and independent kind/profile exclusions | S-05 saved lifecycle |
| AC-JOB-02 | Reused backing bytes survive close; NF rejects malformed/unknown/schema/graph/generation/snapshot/incident payloads on later page | Jobs typed execution unit; S-05 saved lifecycle |
| AC-JOB-03 | Whole Jobs records compare equal except attempt ID/lease expiry; repeat and nonresumable failures | S-04/S-05 saved lifecycle |
| AC-JOB-04 | Query/scan/iteration/cancellation errors return no partial page; closed pages precede writes; late failure rolls back earlier writes | Jobs typed execution unit; NF late-page savepoints |
| AC-JOB-05 | Actual Reporting restore failure after NF reconciliation preserves attempts and complete Graph state; readiness false, then successful exact rebuild | S-05 real Recovery assembly; Recovery atomic evidence row |
| AC-JOB-06 | Empty zero and repeated selected counts independent of cleared attempts | S-04/S-05 saved lifecycle |
| AC-JOB-07 | Borrowed capability cannot begin/commit/rollback; real Recovery writer retains atomic transaction/readiness; NF source enumeration cannot begin a nested transaction | Jobs panic-on-lifecycle fixture; NF source wrapper and composed restore |

The historical all-at-once AC-JOB-04 wording is superseded by consuming and
closing each page before its writes. Ordering tests deliberately use reverse
insertion instead of random insertion so failures are reproducible. No test
reads this matrix or any other Markdown to define executable behavior.

### Cross-cutting contract closure

| Contract group | Current disposition and evidence |
| --- | --- |
| C-01 | Retained Imports owner facade; actual Imports apply, replay, preview authorization, safe errors and rollback |
| C-02 | Retained cursor format/identity/expiry/rotation; NF contract and real pagination/browser recovery |
| C-03 | Retained current authorization before replay and transaction/publication checks; role transitions and denied replay |
| C-04 | Retained atomic owner participants, receipts/audit/outbox and after-commit notification; actual failing writes/commit |
| C-05 | Retained Collaboration delivery ownership and NF event meaning; intent privacy, invalidation and authority-loss scenarios |
| C-06 | Retained Jobs lifecycle/publication/proofs, generation and cancellation/deadline; actual worker and store evidence |
| C-07 | Retained lease/candidate/lock bounds and reachability; replaced global health accounting with S-07 signals and enforced DB deadline |
| C-08 | Retained exact Reporting reader lifetime/redaction; GP-owned lookup/release and complete lease acceptance matrix |
| C-09 | Retained Recovery quiescence/atomicity/readiness; shared source composition and uncapped bounded Jobs pages; full failure rollback |
| C-10 | Retained major 6/state 4 and supported admission; cutover, retained receipts and restored identity fixtures |
| C-11 | Retained immutable rows/mappings and distinct Indicators/revision ownership; store and cross-owner tests |
| C-12 | Retained distinctions between NF graphs, Core views, workbook projections, Graph results and Jobs; framework correction and boundary guards |
| C-13 | Retained protocol/decoder/semantic-grid boundaries; Make generation/drift, type checking and import guard |
| C-14 | Retained captured attempts, exact selected result, polling and uncertainty; controller tests plus all four final stateful scenarios |
| C-15 | Retained build/enablement/origin/scope controls; ordinary/harness builds and routed guard/control tests; no product claim from unused fault registry |
| C-16 | Retained truthful owner/selector/collaborator routing, including Reporting's NF-located test; updated authored manifests and regenerated topology |

### Final command evidence

Commands are exact public Make invocations from the repository root. Unless a
row says otherwise, results are PASS; full run roots are under
`.cartulary/test-results/` with the listed basename.

| Command | Result / run root |
| --- | --- |
| `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.saved_graph_lifecycle_v2` | PASS `20260919T064802Z-p89857`; includes notification/commit/replay assertions |
| `make service-backed-test-slice OWNER=module.imports ROWS=module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0` | PASS `20260919T064408Z-p46946` |
| `make test-slice OWNER=platform.jobs ROWS=platform.jobs.unit.typed_execution_contract` | PASS `20260919T064803Z-p90074` |
| `make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.v2_contract_projection` | PASS `20260919T065019Z-p27040`; service-backed evidence in S-03 |
| `make service-backed-test-slice OWNER=module.recovery ROWS=module.recovery.integration.atomic_typed_terminal_evidence_7a41f0d8c2` | PASS `20260919T064913Z-p8676` |
| `make service-backed-test-slice OWNER=module.reporting ROWS=module.reporting.integration.exact_graph_result_lease_lifecycle_8f1c5c43a2` | PASS `20260919T065017Z-p26817`; ownership retained despite test location in NF |
| `make test-slice OWNER=platform.telemetry` | PASS `20260919T064914Z-p8893` |
| `make generate` | PASS `20260919T065531Z-p52343` |
| `make agent-finalize` | PASS `20260919T065547Z-p55419` and final `20260919T070837Z-p17992`; retained-run maintenance skipped because `RESULTS_DIR` was unset |
| `make generate-drift` | PASS `20260919T065625Z-p59446` |
| `make generated-artifact-policy-check` | PASS `20260919T065625Z-p59431` |
| `make json-shape-check` | PASS `20260919T065625Z-p59380` |
| `make migration-drift` | PASS `20260919T065625Z-p59424` |
| `make frontend-typecheck` | PASS `20260919T065625Z-p59715` |
| `make frontend-import-boundary-check` | PASS `20260919T065625Z-p59775` |
| `make check` | PASS `20260919T071101Z-p54535`; 937/937 work units |
| `make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser_stateful.pagination_recovery,module.networkflow.browser_stateful.saved_graph_deferred_navigation,module.networkflow.browser_stateful.saved_graph_exact_result_lifecycle,module.networkflow.browser_stateful.saved_graph_read_recovery` | PASS `20260919T070907Z-p21726`; 13/13 work units |
| `make lint-markdown` | PASS `20260919T071154Z-p99134`; legacy summary at `adhoc/lint-markdown/tool-run-summary.json` |

Finalization failures `20260919T065117Z-p44794`, `20260919T065342Z-p50241`,
`20260919T065447Z-p51370` and direct JSON checks `20260919T065155Z-p45357`,
`20260919T065414Z-p50726`, `20260919T065514Z-p51835` exposed stale generated
inputs and the missing migration-44 owner assignment. `make generate` initially
failed at `20260919T065218Z-p45988` for that assignment. Updated the authored
migration generator and independent ownership validator; regenerated history,
object ownership, SQL and topology through Make. Successful generation also ran
at `20260919T065313Z-p47119`. No generated output was retained from a failed
partial generation. The OTel source review corrected startup wording to match
the adopted decision: continuation begins at zero; freshness alone is absent
before first success. These failures were related to this change and are resolved.

### Owner/projection review and rollout

Manual source review compared GP §8.0 with the PostgreSQL capabilities and storage
projection; Core 01 §3.3.9.3 with Jobs pages/index and NF reconciliation; NF §29
with the semantic/application/composer/facade boundaries; and NF-REQ-195/196 plus
OTEL-REQ-152 with registry v2, instrument callbacks and corpus expectations.
Specifications remain authority; this review and passing execution do not promote
routing manifests into requirements. No product check depends on Markdown.

[The release handoff](networkflow-remediation-release.md) contains migration 44,
telemetry alert/dashboard replacement, operational verification and rollback.
Application/internal callers, additive index and telemetry consumers ship as one
coordinated release. The index may remain on rollback; old application and old
telemetry configuration must roll back together. Deployment, maintainer release
review and operator dashboard changes remain external steps, not claimed here.


### Current work items and completion gates

| Work items / gate | Current status | Evidence and disposition |
| --- | --- | --- |
| NF-001–008, NF-012–014, NF-019–021, NF-025 | DONE | Original inventory/history retained; accepted scope replaces tracker-only restrictions; current owner/evidence review in §§13–14 |
| NF-009, NF-022; RB-002; G-OWNER-GP | VERIFIED | Adopted GP §8.0 and S-03 storage, lease and race evidence |
| NF-010, NF-023; RB-003; G-OWNER-JOBS | VERIFIED | Adopted Core/NF paged contract; S-04 and S-05 direct/whole-restore evidence |
| NF-011 | VERIFIED | S-07 signals and deadline implementation; old global accounting removed; external monitoring cutover documented |
| NF-015; G-EVID-S01–S04 | VERIFIED | Baseline records plus explicit post-change closure below; no currently failing preservation assertion |
| NF-016 | VERIFIED | S-01 composition and three-consumer S-05 evidence |
| NF-017 | VERIFIED | S-02 application commands and S-05 notification/commit visibility |
| NF-018 | VERIFIED | S-06 semantic boundary and S-08 production facade with positive guards |
| RB-001; G-AUTH-00; G-AUTH-PROD | SATISFIED | One explicit user authorization for the entire accepted effort |
| NF-024; S-05 | VERIFIED COMPLETE | All required checks passed, assertion accounting closed, all seven current handoffs appended |

S-00's run records prove the assertions executed at that time. Additional planned
real Imports apply/replay/rollback, distinct-limit restore, complete payload/scope
matrix and whole-restore failure assertions were completed during S-05; they are
not backdated into baseline evidence. The final acceptance matrix relies on the
actual post-change results as well as the retained baseline.

Current source basis remains HEAD `b9e182b710f6cbfa9d56b656f3a8558012f75044`
plus the uncommitted worktree. No commit or index update was performed. The
preexisting staged tracker is retained (current index blob
`f36c2b8d554d51cc3aaa4cfe70b51551d8402d51`); its older historical-session blob
in §1 is historical. No production deployment or external communication occurred.


First full `make check` run `20260919T065652Z-p68004` failed 6 of 937
work units: three migration source/evidence digest expectations (the harness
hash assertion ran in two units), the backend boundary guard, and Go lint.
Migration 44 intentionally advances the canonical source hash to
`71702e01e2ebd4ac5186dfe9e29dd856a6df57fe716cbae52f9d8484f94dad05` and
the deterministic migration-evidence payload digest to
`167f767dfa8c643553b9538431f4a062b5ab216241982a11a0f45dbbba7bdfdc`.
Inspected the payload: history 1–43 is unchanged; version 44 and its evidence
are added. Updated the three authored golden expectations. The isolated operator
reproduction also failed at `20260919T070107Z-p60601` before correction.

Moved physical Collaboration failure SQL into its existing reusable test-support
owner (`collaborationsupport.FailIntentInserts`) and used its existing typed intent
counter. The production/test boundary guard was not relaxed. Removed the unused
private query sentinel and obsolete mapping-suggestion helper/alias search rather
than retaining dead compatibility behavior. Corrected the composer's internal
error capitalization. The first follow-up NF/operator/vet builds found the now
unused `sort` import (`20260919T070649Z-p72226`, `20260919T070649Z-p72263`);
removed it before rerunning. Migration units passed at `20260919T070649Z-p72240`
and `make backend-module-boundary-check` passed at `20260919T070649Z-p72504`.
All of these failures are related to the remediation and remain recorded.

Follow-up `make service-backed-test-slice OWNER=module.networkflow
ROWS=module.networkflow.integration.saved_graph_lifecycle_v2` passed at
`20260919T070728Z-p87638`, including the Collaboration-owned physical failure
fixture. `make test-slice OWNER=app.operator
ROWS=app.operator.unit.migration_evidence_cli_transport_and_redaction_906768574c`
passed at `20260919T070728Z-p87660`. `make test-slice
OWNER=module.database_migrations
ROWS=module.database_migrations.unit.production_ddl_v2_recurrence,module.database_migrations.unit.test_harness_targeted_operation_validation`
passed at `20260919T070649Z-p72240`. `make lint-go` passed; its individual
format/vet/staticcheck roots are `20260919T070728Z-p88042`,
`20260919T070731Z-p92722`, and `20260919T070738Z-p3729`.

### Changed-file inventory

Paths below supplement the preserved original 88-file inventory. Removed v1
telemetry registry/schema paths are included. `domain.md` and the owner/caller/test
sources named in earlier records were inspected without changing their authority.
Generated GP, topology and migration manifests were produced through Make.

```text
contracts/graph-projection/storage-maintenance.v1.json
contracts/otel/cartulary_signal_registry.v1.json
contracts/otel/cartulary_signal_registry.v2.json
db/migrations/00044_jobs_restore_enumeration.sql
docs/graph_projection_nlspec.md
docs/handoffs/cartulary_modular_refactor_planning_framework.md
docs/handoffs/networkflow-module-refactor-tracker.md
docs/handoffs/networkflow-remediation-release.md
docs/network-flow-activity-nlspec.md
docs/opentelemetry-instrumentation-nlspec.md
docs/spec/01_architecture_storage_and_view_contracts.md
internal/app/operator/operator_migration_evidence_test.go
internal/app/server/network_flow_telemetry.go
internal/app/server/network_flow_telemetry_test.go
internal/gen/contractgraphprojection/artifacts_gen.go
internal/modules/database_migrations/catalog_characterization_test.go
internal/modules/graphprojection/postgresresult/contract_test.go
internal/modules/graphprojection/postgresresult/lease_scope.go
internal/modules/graphprojection/postgresresult/store_test.go
internal/modules/imports/imports_integration_test.go
internal/modules/incidentbundles/routes_extension_integration_test.go
internal/modules/networkflow/api.go
internal/modules/networkflow/application.go
internal/modules/networkflow/binding_store.go
internal/modules/networkflow/collaboration_producer.go
internal/modules/networkflow/configuration.go
internal/modules/networkflow/configuration_test.go
internal/modules/networkflow/csv_parser.go
internal/modules/networkflow/digest.go
internal/modules/networkflow/extension_state.go
internal/modules/networkflow/graph.go
internal/modules/networkflow/graph_capacity_test.go
internal/modules/networkflow/graph_cleanup_deadline_test.go
internal/modules/networkflow/graph_materialization_timeout_test.go
internal/modules/networkflow/graph_projection_adapter.go
internal/modules/networkflow/graph_response_v2.go
internal/modules/networkflow/graph_restore_source.go
internal/modules/networkflow/graph_result_cleanup.go
internal/modules/networkflow/graph_result_cleanup_dispatcher.go
internal/modules/networkflow/graph_result_cleanup_dispatcher_test.go
internal/modules/networkflow/graph_result_cleanup_integration_test.go
internal/modules/networkflow/graph_result_cleanup_test_bridge_test.go
internal/modules/networkflow/graph_routes.go
internal/modules/networkflow/graph_source_composer.go
internal/modules/networkflow/graph_streaming_test.go
internal/modules/networkflow/graph_telemetry.go
internal/modules/networkflow/graph_telemetry_test.go
internal/modules/networkflow/graph_temporal.go
internal/modules/networkflow/graph_temporal_test.go
internal/modules/networkflow/graph_view_authority_integration_test.go
internal/modules/networkflow/graph_view_authority_test_bridge_test.go
internal/modules/networkflow/graph_view_jobs.go
internal/modules/networkflow/graph_view_receipts.go
internal/modules/networkflow/graph_view_receipts_test.go
internal/modules/networkflow/graph_view_routes.go
internal/modules/networkflow/graph_view_store.go
internal/modules/networkflow/import_facade.go
internal/modules/networkflow/import_owner_errors.go
internal/modules/networkflow/indicator_link.go
internal/modules/networkflow/indicator_link_admission.go
internal/modules/networkflow/indicator_link_boundary_test.go
internal/modules/networkflow/indicator_link_graph_admission.go
internal/modules/networkflow/integration_surface_test.go
internal/modules/networkflow/keyring.go
internal/modules/networkflow/keyring_test.go
internal/modules/networkflow/mapping.go
internal/modules/networkflow/module.go
internal/modules/networkflow/names.go
internal/modules/networkflow/network_flow_contract_test.go
internal/modules/networkflow/network_flow_unit_test.go
internal/modules/networkflow/production_surface_test.go
internal/modules/networkflow/query.go
internal/modules/networkflow/query_authoring_test.go
internal/modules/networkflow/query_filter.go
internal/modules/networkflow/query_sql.go
internal/modules/networkflow/reporting_graph_source.go
internal/modules/networkflow/reporting_graph_source_integration_test.go
internal/modules/networkflow/resources.go
internal/modules/networkflow/routes.go
internal/modules/networkflow/routes_integration_test.go
internal/modules/networkflow/saved_graph_admission.go
internal/modules/networkflow/saved_graph_application.go
internal/modules/networkflow/saved_graph_receipts.go
internal/modules/networkflow/security.go
internal/modules/networkflow/semantic_boundary_test.go
internal/modules/networkflow/semantic_failure.go
internal/modules/networkflow/semantic_http_adapter.go
internal/modules/networkflow/semantic_http_errors.go
internal/modules/networkflow/semantic_http_test_bridge_test.go
internal/modules/networkflow/store.go
internal/modules/networkflow/table_lifecycle_test.go
internal/modules/networkflow/textcanon_test.go
internal/modules/networkflow/timestamp.go
internal/modules/networkflow/timestamp_test.go
internal/modules/networkflow/transaction_participants.go
internal/modules/workbook/workbook_startup_test.go
internal/platform/jobs/restore_enumeration.go
internal/platform/jobs/restore_enumeration_test.go
internal/platform/telemetry/registry.go
internal/platform/telemetry/registry_test.go
internal/testutil/collaborationsupport/faults.go
internal/testutil/golden/otel/cases/OTEL-CORPUS-015/input.json
internal/testutil/golden/otel/cases/OTEL-CORPUS-019/input.json
internal/testutil/pgtest/pgtest_test.go
tools/database-migrations/generate-catalog-projections.mjs
tools/execution_topology_render_index.json
tools/harness/generated-artifacts/database-contract-drift/schema-object-ownership.mjs
tools/harness_schema_attachments.json
tools/migration_history_manifest.json
tools/otel/check-otel-conformance.mjs
tools/schema_object_ownership_manifest.json
tools/schemas/cartulary.otel_signal_registry.v1.schema.json
tools/schemas/cartulary.otel_signal_registry.v2.schema.json
tools/test_families/app.server.json
tools/test_families/module.networkflow.json
tools/test_families/platform.jobs.json
```

### Final completion assessment

| Completion level | Result | Basis |
| --- | --- | --- |
| Specification adoption | COMPLETE | Accepted user decisions adopted in GP/Core Jobs/NF/OTel owners; historical proposals explicitly superseded |
| Implementation readiness | COMPLETE | Scope/adoption gates and actual preservation evidence recorded; later test additions are not mislabeled baseline |
| Implementation completion | COMPLETE | All selected production slices implemented; all F-01–F-17 dispositions and NF work items closed |
| Verified completion | COMPLETE | Final make check 937/937; four stateful scenarios 13/13 work units; all required generation, migration, schema, frontend, Markdown and owner checks pass |
| Handoff completion | COMPLETE | Seven handoff tables updated; assertion matrix, failures, artifacts, changed files, release/rollback and external steps recorded |
| Operational rollout | EXTERNAL / NOT EXECUTED | Maintainer release review, migration deployment and telemetry alert/dashboard cutover follow the release handoff |

S-05 is the final completed slice. No required repository check is skipped or
failing. Retained-run maintenance alone was skipped because `RESULTS_DIR` was
unset; the successful final check is recorded as new evidence, not retroactively
supplied to that earlier maintenance step. No commit, index mutation, deployment
or message to an external party was performed. Final post-validation changes to
this file record results, inventory and handoff status only; table structure and
whitespace were reviewed after those log updates.

## 15. F-18–F-23 — legacy removal and production readiness

### 15.1 Scope, basis and evidence classes

The original 2026-09-19 planning update authorized only this tracker; it did
not authorize production changes, owner amendments, generation or deployment.
The subsequent explicit implementation request superseded that restriction and
authorized S-09–S-16. Section 16 records the resulting work. Historical finding
and slice IDs, inventories, assertions, handoffs and prior evidence remain;
planning observations below describe their original basis, while §15.10–15.11
now report current completion. No index mutation or operational deployment was
performed.

The planning basis and this update's starting basis are the clean checkout
`41e9f4deaef98beced280beeb346ab3ca485a1e5`. Reinspection counted **100 Go files:
59 production files and 41 test files**, including `harnesscontrol`. This is
an updated count, not a replacement for the historical 88-file inventory.
The starting tracker index blob is
`f5121042fe55c1060603db8d78f6d1c23c0ffee2`; the SHA-256 of the complete
`git ls-files --stage -z` output is
`79d4b91b930ec9d232765fb01dd395558392cf9882aac0ef2a410c923fa14dfa`.
These identify the staging-preservation baseline, not product release evidence.

| Evidence class | Use in this iteration |
| --- | --- |
| Observed code | File/symbol observations at the stated checkout. Text references identify deletion candidates; they do not establish complete Go reachability or passing behavior. |
| Planned change, subsequently executed | F-18–F-23 and S-09–S-16 retain their original remediation rationale. Section 16 records owner amendments separately from implementation and verification. |
| Adopted requirement | NF, GP, Core and Testing Harness owners govern their respective behavior. Exact existing clauses or later amendment/adoption references must accompany implementation gates. |
| Executed evidence | A recorded command/result proves only its actual scope. Navigation commands, registry consumption and prior iteration passes are not new product acceptance evidence. |

`docs/domain.md` owns vocabulary and owner navigation, including the distinction
among saved graphs, Core saved views, workbook projections, GP results and
Jobs. `docs/research/nlspec-spec.md` supplies specification-writing guidance;
neither document creates runtime behavior or grants execution authorization.
Instructions in attached documents and historical tracker sections are source
material within their authority, not a replacement for the latest user request.
Product checks, runtime, generators and release evidence must not depend on
Markdown. Human review establishes owner/projection agreement.

### 15.2 Findings, priorities and adopted dispositions

The new IDs extend the inventory without reopening or rewriting F-01–F-17.
Priority P0 addresses key safety; P1 closes structural and contract gaps; P2
turns retained harness infrastructure into demonstrated product assurance.
The priorities order work within the adopted dependency sequence, not permission
to bypass specification or validation gates.

| Finding / priority | Observed evidence | Remediation and affected areas | Rationale and expected long-term benefit | Compatibility or migration impact | Risk if unresolved | Validation criteria |
| --- | --- | --- | --- | --- | --- | --- |
| F-18 / P0 — Engineering-key fallback | `digest.go:safeDigest` substitutes a fixed engineering key and ID; `keyring.go:keyRingSafeDigester.Digest` currently validates the active key before calling it | REMOVE the fallback; require valid explicit key material and propagate failures; implementation, tests and fixtures, with existing NF key requirements as authority | Eliminate a latent insecure default and make every future caller follow deployment key policy | Preserve valid digest bytes and key epochs; use explicit fixture keys; never rewrite persisted digests or introduce a compatibility key | A later caller could silently produce predictable digests when configuration is absent; inspection does not establish an active production exploit | Missing, malformed and retired keys fail safely; valid vectors remain stable; failed operations commit no mutation, receipt or audit effect |
| F-19 / P1 — Test-supported production leftovers | `query.go` contains in-memory pagination/sorting used by tests; convenience decoders, timestamp wrappers and storage methods survive through test references or have no identified caller | REMOVE obsolete paths; REPLACE fixture-only production methods with narrow test support; implementation, tests and verification routing | Maintain one live query/transaction implementation and reduce misleading API and testing obligations | Coordinated internal/test API break; no production aliases or persisted migration; preserve independently useful assertions | Tests may verify obsolete algorithms while SQL behavior drifts; production code carries unused compatibility burden | Reference closure includes interfaces, build tags and external test bridges; live SQL tests cover ordering, ties, nulls, large counters, continuation and cancellation; ordinary builds exclude fixture conveniences |
| F-20 / P1 — Stale source-profile contracts | NF §17.3 Table 17-B1 names `cartulary.network_flow_source_profile_list.v1`; the route/backend/frontend use `cartulary.network_flow.source_profile_list.v2`; the authored schema index retains both dotted v1 and v2 projections | REPLACE the owner/projection discrepancy by explicitly adopting the current v2 response; REMOVE the obsolete v1 projection and exclusively orphaned definitions after adoption; specification, contracts, generated outputs, tests and documentation | Establish one current response and eliminate generated compatibility baggage; do not infer authority from existing implementation alone | Recommend retaining the current v2 wire shape; coordinate removal of unused generated Go/TypeScript surfaces; preserve independently supported v1 contracts | Future changes can target inconsistent requirements or accidentally reintroduce an obsolete response | Owner text, route declaration, backend response, frontend decoder and generated schema agree; generated artifacts come from authored inputs through Make |
| F-21 / P1 — Remaining transport coupling | `extension_state.go` calls `decodeGraphSemanticRequestHTTP`; `application.go` table commands and indicator-link admission/participants carry route-service or HTTP-error responsibilities | REPLACE HTTP-coupled reusable logic with private table/link application components, semantic failures and dedicated receipt adapters; implementation, tests, boundary documentation and necessary owner clarification | Complete cohesive application boundaries for future consumers without route-service reuse or HTTP envelopes inside participants | Preserve admission precedence, privacy, receipt bytes and the distinct table/link/saved-graph replay algorithms; no generic mutation framework or new public API | Future commands inherit transport dependencies; refactors can reorder security-sensitive checks or couple unrelated replay policies | Persisted validation calls semantic decoding directly; real HTTP/command/transaction assertions retain replay, rollback, rechecks, redaction and captured-attempt continuity |
| F-22 / P1 — Divergent materialization-payload admission | `graph_view_jobs.go` uses permissive `json.Unmarshal`; `graph_restore_source.go` uses a separate decoder with `DisallowUnknownFields` | REPLACE worker, restore and startup decoding with one private strict decoder using transport-independent JSON validation; implementation, tests and owner clarification where needed | Give workers and restore one interpretation as job payloads evolve | Preserve valid payload format and durable state version; reject malformed inputs without silently rewriting retained jobs | A worker may accept payloads that restore rejects, or the consumers may select different identities | All three consumers (worker, restore, startup) reject duplicate/unknown members, trailing data, invalid identities and incident mismatches before owner mutation; later-page failure rolls back the whole restore and never publishes readiness |
| F-23 / P2 — Harness controls without product consumers | Four control registries expose arming/consumption mechanics; inspected consumption is registry testing, and server-process coverage proves route contribution rather than product effects | REPLACE unused mechanics with real fault/randomness/auth-transition/audit consumers; REMOVE unsupported tokens; harness owner, implementation, contracts, tests, routing and documentation | Make retained infrastructure earn its maintenance cost through observable product assurance | Coordinated harness vocabulary/fixture migration; preserve production API behavior and shared harness security; no dual compatibility surface | Controls can imply rollback, authorization or audit assurance without exercising a product consumer | Each retained token maps to an owner requirement, named consumer, actual assertion, selector, command and passing artifact; ordinary builds contain no control routes or registry dependencies |

### 15.3 Deletion inventory and retention rules

This inventory was established in planning and closed in S-09/S-11/S-12/S-15.
Reference review included production and test callers, interfaces, build tags,
generated roots and external test bridges. The dispositions below are complete;
§16 records replacement assertions, builds and caller migrations. Production-only
reference counts were a screening aid, not the reachability proof.

| Candidate | Completed disposition | Replacement or preservation evidence |
| --- | --- | --- |
| `sortRows`, `pageFlowRowsAfter`, `pageDiagnosticsAfter` and exclusively dependent comparators/helpers | REMOVE obsolete in-memory query paths | Transfer meaningful sort/cursor/diagnostic assertions to the real SQL query path; retain comparators independently used by live graph or diagnostic behavior |
| `parseTimestamp` | REMOVE test-only convenience wrapper | Call the live timestamp parser from focused tests with explicit record context; retain timestamp modes, precision, DST and timezone rules |
| `decodeIndicatorLinkRequest` | REMOVE test-only request wrapper | Exercise the actual route/object-admission path; retain strict JSON and security-sensitive error precedence |
| `networkFlowLinkableIPField` | REMOVE test-only policy duplicate | Verify the live selector/field admission boundary and its negative cases |
| `store.GetTable` and `store.ListRejectedRowDiagnostics` | REMOVE unused/unbounded convenience access after reference closure | Retain live authorized table reads and bounded diagnostic queries; fixture reads must not create a second product read path |
| `store.CreateTable`, `store.RenameTable`, `store.SoftDeleteTable`, `store.RetainedCounts` and aliases supporting only tests | REMOVE from production where fixture-only; REPLACE with narrow fixture operations | Keep live transaction participants and transaction-bound methods; preserve real database rollback/concurrency assertions and external test placement needed to avoid import cycles |
| `SourceProfileList` v1 projection and exclusively orphaned definitions | REMOVE through S-12 after NF owner adoption | Trace references from all current schema roots, generators, frontend adapters and supported persisted formats; keep shared definitions that remain live |
| Synthetic/unused harness tokens | REMOVE only through the S-09/S-15 owner-aligned capability matrix | Preserve all four functional control families; no accepted token may remain without a meaningful consumer and acceptance assertion |

Retain identity and digest algorithms, supported migration verification,
current-state admission, historical receipts, exact-result leases, recovery
reconciliation and negative compatibility tests. A `v1` name is not deletion
evidence. Preserve the private NF transaction helper, atomic owner participants,
NF intents/App translation/Collaboration delivery, immutable analytical rows,
frontend adapters and authored/generated separation. Do not broaden this
iteration into speculative package relocation or a generic transaction facade.

### 15.4 Owner decisions and adoption gates

| Gate | Existing authority and proposed action | Completion evidence | Current state |
| --- | --- | --- | --- |
| G-NEXT-KEY | NF §6.7, especially NF-REQ-045/045a/045b, and §20.1 already require deployment keys and preserve historical digest epochs; implement F-18 without weakening those requirements | Existing clause references, valid-vector comparison and failure/rollback assertions | CLOSED — existing owner retained; S-10 vectors and transaction rollback pass |
| G-NEXT-SCHEMA | NF §17.3 Table 17-B1 conflicts with downstream response projections; recommend explicit current v2 adoption before pruning the v1 schema | Exact owner revision/adoption record, version-action rationale, affected authored/generated surface and human agreement review | ADOPTED — NF 6.0.1 §17.3; S-12 generated/backend/frontend agreement passes |
| G-NEXT-APPLICATION | NF §§5.5, 15, 17, 28–29 own admission/replay and application boundaries; extend the semantic boundary to the remaining table/link/state seams without unifying their replay algorithms | Existing references and any necessary clarification, application dependency design and assertion mapping | ADOPTED / COMPLETE — NF 6.0.1 §29.1; S-13 boundaries and replay assertions pass |
| G-NEXT-PAYLOAD | NF §29 owns restore payload meaning; Core Common Jobs owns job lifecycle and GP §9/Core Recovery own atomic restore and readiness | Explicit shared-decoder admission contract, worker/restore behavior mapping and malformed-payload assertions | ADOPTED / COMPLETE — NF 6.0.1 §29.1; S-14 and S-16 corpus covers all three consumers, payload v1 retained |
| G-NEXT-HARNESS | Testing Harness §§12.2.5–12.2.8, TH-HARNESS-REQ-455–459 and 465–479, own control mechanics; product owners define the expected effects | Reduced token/consumer matrix, exact owner amendments, fixture migration notes and replacement acceptance/routing records | ADOPTED / COMPLETE — Testing Harness §§12.2.5–12.2.9; S-15a–S-15d v2 controls and consumers pass |

The S-09 source-profile adoption states the exact current
`cartulary.network_flow.source_profile_list.v2` response and its current fields
as the sole emitted shape. It does not retain a second decoder or an alias.
The adopted shared payload contract requires one closed JSON object, strict
member handling and NF-owned identity validation; job and restore adapters keep
their distinct failure outcomes. Harness amendments must name real boundaries
and fixture obligations without creating product state-machine stages.

S-09 must record each amendment's exact document/version/section and adoption
reference. If adopted owners themselves conflict, record the conflicting
passages and block only the dependent work until the conflict is resolved.
An accepted tracker plan, a machine projection or a passing test does not meet
an owner-adoption gate. Future execution authorization applies to its selected
scope once; technical adoption/evidence gates do not imply repeated per-slice
permission requests.

### 15.5 Workstreams, dependencies, risks and exit criteria

All workstreams below are **COMPLETE**, in the opening notice's exact order.
The table retains the planned dependencies, risks and exit criteria; §16 records
each outcome and the evidence that closes it. S-16 is the final completed slice.

| Workstream | Scope and dependencies | Main execution risk | Exit criteria |
| --- | --- | --- | --- |
| S-09 — Specification and deletion inventory | Establish F-18–F-23, full reference evidence, retained invariants, affected projections and the reduced harness matrix; precedes all implementation | Mistaking test-only references, historical prose or a proposed contract for authority or proof | Every candidate has a retain/remove/replace disposition; owner amendments and adoption references are explicit; unsupported tokens have a removal disposition; preservation assertions have selectors and a baseline or assigned defect |
| S-10 — Key hardening | Depends on S-09; remove engineering defaults and provide explicit test keys | Accidentally changing valid digest bytes or invalidating historical epochs | Valid vectors stable; missing/invalid/retired keys fail safely; production fallback absent; failures commit no domain, receipt or audit effects |
| S-11 — Dead-code removal | Depends on S-09–S-10; delete obsolete paths and transfer useful assertions to live boundaries | Deleting a live interface/build-tag consumer or losing coverage while tests become shorter | Ordinary builds contain no identified fixture conveniences; live SQL covers ordering, ties, nulls, large counters, continuation and cancellation; necessary external integration coverage survives |
| S-12 — Contract cleanup | Depends on S-11 and G-NEXT-SCHEMA adoption in S-09; remove obsolete authored definitions and regenerate through Make | Pruning a shared definition or treating implementation behavior as unrecorded owner adoption | Owner text, route/backend/frontend and generated schemas agree; obsolete response surface absent; independently supported old-version contracts retained |
| S-13 — Application boundary completion | Depends on S-12 and G-NEXT-APPLICATION; separate table/link commands, receipt adapters, transport mapping and persisted-state validation | Reordering admission/replay or introducing a generic abstraction that couples unrelated commands | Reusable components carry semantic failures only; real routes, transaction rechecks, rollback, privacy and captured-attempt continuity pass; distinct replay policies remain |
| S-14 — Shared job-payload admission | Depends on S-13 and G-NEXT-PAYLOAD; unify worker/restore/startup decoding | Accepting different identities or weakening whole-restore rollback | Valid payload parity; strict malformed-input rejection; incident agreement; late-page failure rolls back and blocks readiness; valid state-4 restore remains supported |
| S-15a — Functional fault controls | Depends on S-13–S-14 and G-NEXT-HARNESS; bind controls to actual owner-apply, transaction and worker boundaries | Proving only consumption, or adding a fictitious phase between atomic effects | Precommit faults prove rollback; postcommit failures prove immutable replay; crash uses an owned child process and recovery evidence; effects are correlated and consumed once |
| S-15b — Functional randomness controls | Depends on S-15a and G-NEXT-HARNESS; bind table allocation and cursor entropy seams | Leaking deterministic fixture behavior into production or replacing deterministic identities with randomness | Primary-key-only conflict handling, collision retry in a usable transaction, eight-attempt exhaustion, nonce behavior, stream mismatch and fail-closed exhaustion pass; production entropy and identity algorithms remain intact |
| S-15c — Functional authorization transitions | Depends on S-15b and G-NEXT-HARNESS; apply real fixture state changes at supported admission boundaries | Inventing response-time authorization semantics or allowing a harness control to bypass product admission | Membership/session loss, table changes, cursor rechecks and nondisclosure are demonstrated through actual responses; unsupported checkpoint tokens are retired |
| S-15d — Functional audit assertions | Depends on S-15c and G-NEXT-HARNESS; inspect committed owner audit state after real operations | Counting attempted/uncommitted appends or unrelated concurrent operations | Exact scoped counts cover creation, rename, deletion, graph query, binding creation/reuse, denied operations and replay; registry-only evidence remains separately classified |
| S-16 — Validation and handoff | Depends on every prior exit criterion, including all four S-15 records | Reusing previous passes to certify changed code, missing collaborator rows, or conflating repository completion with rollout | Required gates pass; deletion inventory, assertion matrix, owner review, compatibility notes, rollout/rollback and all seven handoff tables are complete; external operational steps remain explicit |

After each workstream and before the next, record the checkout/worktree basis,
changed files, owner decisions/adoption references, compatibility effects,
executed assertions, exact Make commands, run/artifact paths, failures, skipped
checks with reasons and each exit-criterion result. S-15a–S-15d require separate
records; an aggregate S-15 pass cannot hide an unimplemented family. A required
failing preservation assertion blocks the dependent slice, and a required final
failure blocks S-16 verified completion.

### 15.6 Harness capability and consumer contract

The user chose to make all four control families functional, then chose to
prune unsupported tokens. Preserve these two decisions together: neither a
wholesale deletion of the controls nor implementation of every historical token
is the accepted direction.

| Family | Retained purpose and consumer | Removal/simplification direction | Required product evidence |
| --- | --- | --- | --- |
| Faults | Real Imports owner-apply and transaction boundaries, worker handler entry/final commit/completed publication; wrappers around the actual participating ports | Replace hypothetical prepare/publication checkpoints with named real stages; never split atomically committed effects to satisfy a test token | Rollback of observable durable state before commit; exact replay after committed response failure; worker cancellation/crash/recovery through actual Jobs/owner coordination |
| Randomness | `network_flow.table_id` drives actual table allocation; `network_flow.cursor_nonce` drives the cursor entropy seam in fixtures | Remove `network_flow.row_id`, `network_flow.diagnostic_id`, `network_flow.safe_digest_nonce`, `network_flow.graph_invocation_id`, `network_flow.import_job_id` and `network_flow.import_source_ref` from this NF control surface | Collision recovery and eight-collision failure without partial effects; reproducible nonce scenarios, key rotation/TTL/replay assertions as applicable; no fallback after an armed sequence is exhausted |
| Authorization transitions | Real fixture transitions and supported route/cursor admission boundaries, with exact actor/incident/resource/correlation matching | Remove speculative response-time or publication-time checkpoints that would invent new product authorization semantics; retain owner-defined admission stages | Actual membership/session loss, table rename/deletion and cursor rechecks; safe failure/hidden-resource responses, not synthetic envelope substitution |
| Audit assertions | Fixture support reads committed audit occurrences through owner-aware capabilities after a real operation | Do not keep assertion tokens without a real operation/count consumer; do not add a product audit event merely to satisfy a harness token | Exact scoped creation/rename/deletion/query/binding counts, zero-occurrence denial/rollback and no-extra-occurrence replay |

Assemble consumers in harness builds and fixture support. Production owners
must not import harness registries, expose control routes in ordinary builds,
or change their authorization model. Prefer decorators over existing dependency
ports. Any necessary additional seam must be narrow, instance-scoped and
installed before serving; global mutable hooks are prohibited. Preserve the
shared test-route enablement, host/origin and token checks before body decoding.

Consume a matched control before applying its effect. Mismatches leave the
control pending; exhausted deterministic streams fail the fixture. Reset uses
owned process replacement, which clears runtime control state; do not revive
the retired reset route. Crash tests terminate only their owned child process.
Audit assertions inspect committed counts, not registry consumption or attempted
appends. No harness effect may change row IDs, safe-digest algorithms, GP result
identity, owner transaction atomicity or receipt meaning.

The S-09 adopted owner matrix defined each real boundary/effect before
implementation. The S-16 closure below supplies one row per retained token
containing: owner requirement, exact consumer/boundary, valid effect
combinations, fixture identity/correlation scope, expected durable and public
outcome, assertion/selector, removal or migration mapping, and later command/run
evidence. Tokens without such a row are removed through the harness owner
amendment; they are not silently left accepted but inert. This refinement is an
implementation-readiness deliverable, not a new product compatibility promise.

### 15.7 Acceptance matrix and command routing

These assertions are **CLOSED** by new execution in §16. The previous
AC-COMP/AC-CMD/AC-GP/AC-JOB and C-01–C-16 matrices remain historical evidence;
they were not reused to certify this iteration. S-16 maps new collaborator and
final-gate runs to the changed implementation.

| Assertion group | Required scenarios | Owning workstream / evidence boundary |
| --- | --- | --- |
| AC-NEXT-KEY | Explicit valid keys and stable digest vectors; missing/malformed/retired keys; no committed mutation/receipt/audit on failure | S-10; semantic/key-ring tests plus actual mutation rollback |
| AC-NEXT-DEAD | Production/interface/build-tag/test-bridge reference closure; ordinary build; SQL ordering/ties/nulls/large counters/continuation/cancellation | S-11; source audit plus real query/transaction assertions |
| AC-NEXT-SCHEMA | One current source-profile response; old unused exports/validators removed; shared current definitions retained | S-12; human owner review, generated drift, backend and frontend boundary tests |
| AC-NEXT-APPLICATION | Semantic state validation; table/link admission before replay and transaction rechecks; distinct receipt policies; partial failures; safe unknown errors; captured attempts | S-13; real authenticated HTTP and owner transaction tests, dependency guards and affected frontend scenarios |
| AC-NEXT-PAYLOAD | Worker/restore/startup valid parity; duplicate/unknown/trailing/malformed data; schema/ID/generation/snapshot/incident validation; cancellation; later-page rollback/readiness | S-14; shared-decoder tests and actual worker/restore composition, retaining 0/1/256/257/512/513-page-boundary coverage |
| AC-NEXT-FAULT | Matching/mismatching controls; precommit rollback; postcommit replay; cancellation; owned-process crash/restart and reconciliation | S-15a; registry mechanics separately from service-backed/process product evidence |
| AC-NEXT-RANDOM | Table collision/retry exhaustion; cursor nonce/key/TTL scenarios; wrong stream/kind; exhausted sequence; ordinary-build entropy isolation | S-15b; real allocation and cursor consumers |
| AC-NEXT-AUTH | Exact fixture scoping; membership/session transitions; table changes; cursor admission; nondisclosure; controls absent from ordinary builds | S-15c; actual admitted/denied operations and applicable browser scenarios |
| AC-NEXT-AUDIT | Scoped committed occurrence counts for each retained event; failed/denied operations; replay without extra occurrences | S-15d; real operations and owner-aware audit reads |
| AC-NEXT-HANDOFF | All required assertions and gates pass; adoption/projection review; complete removal/migration inventory; operational steps separate | S-16; final evidence and seven handoff tables |

Use the live public Make surface rather than copying a permanent target catalog.
For subsequent work, resolve exact rows through
`make task-guide ROLE=module-author OWNER=<owner-id>` and
`make explain-test-owner OWNER=<owner-id>`, then use
`make test-slice OWNER=<owner-id> ROWS=<selected-row-ids>` and
`make service-backed-test-slice OWNER=<owner-id> ROWS=<selected-row-ids>`.
Update authored selectors and collaborator accounting before regenerating
topology. Current NF routing has 93 rows, 44 service-backed; counts are discovery
facts, not a coverage or completeness claim.

Route affected coverage to NF, Imports, Jobs, Recovery, Reporting, GP,
application server, frontend NF and browser harness owners as appropriate.
Preserve the Reporting-owned exact-result lease row even when its test is in
NF. Run ordinary and harness builds, positive API/dependency guards, and real
Imports/Jobs/Recovery/Reporting/GP boundaries. Frontend validation includes
type checking, import boundaries and affected stateful replay, selection,
refresh/retirement, authority-loss and pagination scenarios.

Before broader final verification run `make agent-finalize`. Supply retained
successful `RESULTS_DIR` evidence only when valid; otherwise explicitly record
that retained-run maintenance was skipped because `RESULTS_DIR` was unset.
The selected final gates include `make generate-drift`,
`make generated-artifact-policy-check`, `make json-shape-check`,
`make migration-drift`, `make harness-contract`, `make frontend-typecheck`,
`make frontend-import-boundary-check`, `make lint-markdown` and `make check`,
plus selected frontend/service-backed/stateful assertions. Use `make generate`
for changed authored inputs; never hand-edit generated roots. Record actual
commands and artifacts, including failures and justified skips; product tests
must not read this tracker or other Markdown.

### 15.8 Compatibility, rollout and rollback

Retain public major 6 and durable state 4 where their contracts are preserved.
F-20 requires an explicit owner reconciliation and version-action record rather
than treating an existing downstream schema as authority. Internal Go/test API
changes may break coordinated repository callers; do not preserve obsolete
production aliases. Valid job payloads, identities, receipts, key epochs and
supported recovery history remain intact. No new database migration is proposed
by this iteration; do not edit historical migrations or undo the previous
additive Jobs index.

Publish harness vocabulary changes together with their fixture/runner callers,
owner/projection updates and release notes. Retired tokens have explicit removal
or replacement mappings; no unsupported token is dual-supported as a shim.
Roll back harness/application changes with their matching fixture and generated
contract versions. Do not roll back by restoring an engineering key fallback.

S-16 must human-review owner/projection agreement and prepare a release handoff
covering changed interfaces, removed helpers/schemas/tokens, required caller
changes, verification artifacts and remaining deployment steps. The previous
release's deployment and telemetry cutover remain external unless separately
executed and evidenced. Repository verified completion and operational rollout
must be reported separately; earlier successful runs cannot certify this work.

### 15.9 Documentation execution and planning evidence

| Command or inspection | Session and result | Evidence scope |
| --- | --- | --- |
| `make help` | Prior planning turn: PASS | Public task discovery only |
| `make task-guide ROLE=module-author OWNER=module.networkflow` | Prior planning turn: PASS | Narrow NF verification guidance; no product assertions executed |
| `make explain-test-owner OWNER=module.networkflow` | Prior planning turn: PASS; 93 rows, 44 service-backed | Authored routing discovery, not evidence that requirements are complete |
| `make task-guide ROLE=module-author OWNER=platform.harness` | Prior planning turn: FAILED, unknown active owner | Corrected navigation error, not a product failure |
| `make task-guide ROLE=module-author OWNER=app.server` and `make task-guide ROLE=module-author OWNER=harness.browser` | Prior planning turn: PASS | Correct owner guidance after the unsuccessful lookup |
| `git diff --check` | Prior planning turn: PASS with no edits | Whitespace check of that clean planning basis only |
| `git status --short`, `git rev-parse HEAD`, `git diff --cached --stat`, index and file-count inspection | Documentation execution: PASS at start; clean checkout and counts match §15.1 | Starting basis and staging-preservation record |
| `make explain-target TARGET=lint-markdown DETAIL=summary` | Documentation execution: PASS | Confirmed public target and artifact locations |
| `make lint-markdown` | Documentation execution: PASS, exit 0; run root `.cartulary/test-results/20260919T141429Z-p44134`, summary `adhoc/lint-markdown/tool-run-summary.json`, duration 81,720 ms | Required repository Markdown check; existing default globs omit this tracker, so target success alone does not lint this file |
| Direct tracker structure/link/history audit and `git diff --check` | Documentation execution: PASS; repeated after final result logging | New tables have consistent columns, numbered sections resolve, referenced source paths exist, existing links and §§1–14 are unchanged; only the tracker changed and the complete staged-entry digest matches §15.1 |

This update changes only this tracker. Owner documents, source bodies, contracts,
generated artifacts, tests, task manifests and Markdown tooling were inspected
read-only. Product suites, code generation, migrations, `make agent-finalize`
and broader final gates are **NOT RUN: documentation-only scope**. No retained
product success is claimed for F-18–F-23. The Markdown target's existing globs
omit this handoff; its coverage limitation is recorded rather than changing
configuration outside the authorized file.

The final edit records the completed Markdown result and documentation handoff;
the direct structure/history/index and whitespace checks are repeated on those
final bytes. No product readiness or implementation-completion claim follows
from these documentation checks.

### 15.10 Current handoff tables

These seven current handoffs supersede the planning-only statuses originally
recorded here. The seven historical tables in §10 remain unchanged.

#### Scope and authority

| Status | Current handoff | Next action / gate |
| --- | --- | --- |
| COMPLETE | F-18–F-23 and S-09–S-16 closed. NF 6.0.1 §§17.3/29.1 and Testing Harness §§12.2.5–12.2.9 adopted before changed requirements were implemented. Prior S-05 history and the initial index are preserved. | Maintainer review of this coordinated change; no further repository remediation gate remains. |

#### Backend module boundary

| Status | Current handoff | Next action / gate |
| --- | --- | --- |
| COMPLETE | Explicit digest keys, live SQL coverage, private table/link commands and separate receipts, semantic participants/state validation, one three-consumer payload decoder, bounded collision-safe allocation. | Extend these private owner boundaries; preserve distinct replay, atomicity, exact lease and recovery rules. |

#### Frontend module boundary

| Status | Current handoff | Next action / gate |
| --- | --- | --- |
| COMPLETE | Sole source-profile v2 response and generated decoding; 66 frontend units and all six stateful NF scenarios pass, including captured operations, authority loss, pagination and saved-graph continuity. | Ship matching generated contracts and frontend together; no v1 alias or speculative relocation. |

#### Contract and codegen

| Status | Current handoff | Next action / gate |
| --- | --- | --- |
| COMPLETE | Retired SourceProfileList/EffectiveLimits and all four harness v1 schemas; generated outputs and v2 attachments agree with owners. Major 6/state 4/payload v1 retained; no SQL migration. | Use Make generation from authored contracts; see release notes for removed exports/tokens and caller migration. |

#### Tests and harness

| Status | Current handoff | Next action / gate |
| --- | --- | --- |
| COMPLETE | All four families have real consumers and scoped product assertions. Strict guards/decoding, rollback, committed replay, owned crash recovery, entropy, real authorization transitions and committed audit baselines/counts pass. S-16 records fresh cross-owner/final evidence. | Keep each retained token linked to its consumer and assertion; fixture failures or pending required controls must fail. |

#### Security and authorization

| Status | Current handoff | Next action / gate |
| --- | --- | --- |
| COMPLETE | No engineering fallback; digest failures roll back. Admission precedence, hidden targets, safe errors, authorization rechecks and per-instance entropy preserved. Ordinary build exposes no control registry/routes. | Rollback must retain key hardening and strict state admission; do not restore unsupported controls or synthetic responses. |

#### Open risks and next session

| Status | Current handoff | Next action / gate |
| --- | --- | --- |
| REPOSITORY COMPLETE / OPERATIONS EXTERNAL | No required repository failure or unresolved finding remains. Historical fixture failures and the justified retained-run maintenance skip remain in §16. Deployment and prior telemetry cutover were not performed. | Operational owner deploys compatible code/contracts/fixtures as a unit and separately records deployment and telemetry evidence. |

### 15.11 Completion assessment for this iteration

| Completion level | Current result | Required basis |
| --- | --- | --- |
| Planning content | COMPLETE | Findings, workstreams, rationale, dependencies, risks, validation and migration retained. |
| Specification adoption | COMPLETE | Owner amendments recorded in S-09; explicit incident-deletion and audit v2 refinements recorded before their consumers. |
| Implementation readiness | COMPLETE | Deletion/reference closure, consumer matrix and preservation assertions assigned before implementation. |
| Implementation completion | COMPLETE | S-10–S-15d completed separately and recorded before progression. |
| Verified completion | COMPLETE | New assertion-level runs and S-16 final gates pass; historical results are not substituted. |
| Documentation and implementation handoff | COMPLETE | Seven tables closed; owner/projection review, release notes, file inventory and final evidence recorded. |
| Operational rollout | EXTERNAL / NOT EXECUTED | Deployment and earlier telemetry cutover require separate operational evidence. |

## 16. F-18–F-23 execution ledger

### S-09 — specification closure and readiness

Status: COMPLETE. Basis: `41e9f4deaef98beced280beeb346ab3ca485a1e5`;
only the tracker was staged at entry. Initial complete index SHA-256:
`00aaecb0cd7529b2e468f41b88150501254481b2bf014a8f2c28c32371a7ffcd`. The index is not changed by this effort.

Owner adoption is recorded in NF revision 6.0.1 §17.3 and §29.1 and Testing
Harness §§12.2.5–12.2.9. Public major 6, durable state 4, payload v1, and valid
identity/receipt bytes remain. Harness response schemas move to v2 together with
callers. GP and Core ownership is unchanged. Domain vocabulary remains accurate.

The §15.3 inventory is confirmed by production/test/interface reference search:
remove sortRows/pageFlowRowsAfter/pageDiagnosticsAfter; remove parseTimestamp
and test request/policy wrappers; delete unused GetTable and unbounded diagnostic
listing. Move fixture-only CreateTable/RenameTable/SoftDeleteTable/RetainedCounts
operations into test compilation, preserving transaction-bound implementation
and external bridges. Retain comparators used by live graph/cursor code, private
transactions, exact leases, receipt validation, and migration history. Remove
SourceProfileList v1 and only its exclusive definitions after schema traversal.

| Assertion / owner | Consumer and required evidence | Slice / baseline disposition |
| --- | --- | --- |
| AC-NEXT-KEY / NF §6.7 | Digester and real rename/import mutations; valid vectors, invalid keys, complete rollback | S-10; missing failure assertions assigned F-18 |
| AC-NEXT-DEAD / NF §§13,17 | SQL row/diagnostic keysets, timestamp parser and actual object admission | S-11; obsolete-path assertions assigned F-19 |
| AC-NEXT-SCHEMA / NF §17.3 | v2 backend discovery and frontend generated decoder | S-12; owner discrepancy adopted above |
| AC-NEXT-APPLICATION / NF §§15,17,28,29.1 | Table/link applications, receipt adapters, real HTTP and transaction rechecks | S-13; transport defects assigned F-21 |
| AC-NEXT-PAYLOAD / NF §29.1 | Worker, restore AND startup; strict shared admission, whole-restore rollback | S-14; divergent admission assigned F-22 |
| AC-NEXT-FAULT / TH §12.2.5, NF §29 | Eight named real boundaries in TH-REQ-456; rollback/replay/cancellation/crash | S-15a; accepted-but-inert controls assigned F-23 |
| AC-NEXT-RANDOM / TH §12.2.6, NF §6 | table_id allocation and cursor_nonce encryption; collision success/exhaustion | S-15b; also repair retry in aborted PostgreSQL transaction |
| AC-NEXT-AUTH / TH §12.2.7, Core 04 | Route and continuation admission; six real transitions, scoped actual responses | S-15c; remove synthetic response selector and late checkpoints |
| AC-NEXT-AUDIT / TH §12.2.8, NF §16 | Committed audit reads for all six current events; scoped baseline/final/replay counts | S-15d; remove import resource and add actor/incident matching |

Exact selectors are maintained in authored owner manifests as assertions move;
existing entrypoints include table lifecycle admission/replay, pagination recovery,
effective-limit discovery, saved-graph cutover/lifecycle, all four harness support
rows, Reporting exact-result leases, and Recovery rebuild tests. Product assertions
have not run in S-09: identified baseline defects are assigned above, not passes.
`make task-guide ROLE=module-author OWNER=module.networkflow` resolves narrow
verification. `git diff --check` validates edits. Documentation lint is deferred
to S-16; no product evidence reads Markdown. S-09 exits: owner closure, deletion
dispositions, consumer/effect contract, compatibility decisions and assigned
assertions complete. Next: S-10.

### S-10 — explicit safe-digest keys

Status: COMPLETE. Removed fixed engineering key/ID, made the private helper
fallible, validated key ID/32-byte material/value class, and propagated errors
through the key-ring digester. Tests use explicit keys; valid digest bytes and
historical epochs remain unchanged. Added a fixed independent HMAC vector,
invalid-material/class and inactive-epoch cases, and real PostgreSQL rename
rollback assertions for table state, audit occurrences and receipts.

`make format`: PASS (`20260919T143826Z-p52685`).
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.network_flow_selector_covers_logs_telemetry_audi_61f27d1e10`:
PASS (`20260919T143832Z-p56992`).
`make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.resource_intent_failure_rolls_back_source_mutation_8a4c3d2e1f`:
initial FAIL (`20260919T143842Z-p57811`): new fixture omitted client transaction
attribution, skipping the audited branch. Corrected fixture; PASS
(`20260919T143942Z-p75983`, 3/3 units). All run roots are under
`.cartulary/test-results/`. No production failure was waived. Broader validation
is reserved for S-16. Exit: fallback absent, stable valid vector and invalid-key
rollback demonstrated. Next: S-11.

### S-11 — obsolete-path removal

Status: COMPLETE. Removed in-memory row/diagnostic pagination, sortRows and its
exclusive comparator; retained the comparator used by live graph contributors.
Removed parseTimestamp and switched tests to parseTimestampForRecord with explicit
record context. Removed production request/policy wrappers and unused GetTable /
unbounded ListRejectedRowDiagnostics. Fixture convenience transactions now live
in store_fixture_test.go; production transaction-bound methods remain. Added live
SQL assertions for numeric uint64 ordering, ties/nulls, one-row continuations and
cancellation to the real pagination integration fixture. Existing diagnostic
pagination and lifecycle tests retain real PostgreSQL coverage.

`make format`: PASS (`20260919T144151Z-p94030`). First targeted test/build attempts
failed because net/http was removed while error formatting still used it
(`20260919T144206Z-p98462`, `20260919T144206Z-p98475`); restored that import pending
S-13 transport extraction. Corrected runs: `make test-slice OWNER=module.networkflow
ROWS=module.networkflow.unit.network_flow_selector_covers_linking_an_endpoint_cfa3b46c37,module.networkflow.unit.network_flow_selector_covers_timestamp_profile_p_bdf924da13,module.networkflow.unit.query_authoring_admission`
PASS (`20260919T144226Z-p8099`); `make service-backed-test-slice OWNER=module.networkflow
ROWS=module.networkflow.integration.pagination_recovery,module.networkflow.store.network_flow_selector_covers_only_active_and_sof_400db83232`
PASS (`20260919T144238Z-p8901`, 4/4 units, includes ordinary helper build).
Run roots: `.cartulary/test-results/`. No persisted/API migration. Reference search
confirms obsolete production names absent and fixture methods test-only.
Exit criteria satisfied; broader guards run in S-16. Next: S-12.

### S-12 — source-profile contract closure

Status: COMPLETE. NF §17.3 now adopts only source_profile_list.v2. Removed the
unused SourceProfileList v1 root and its exclusive EffectiveLimits definition;
shared SourceProfile, CountMeta and EffectiveLimitsV2 remain. Repository reference
search found no authored callers of the removed exports. Generated Go/TypeScript
and dependent import fingerprints were regenerated, never hand-edited. Added
actual discovery schema-ID assertion and frontend rejection of the retired ID.
No wire/state migration or alias was introduced.

`make generate`: PASS (`20260919T144342Z-p27075`). `make generate-drift`: PASS
(`20260919T144419Z-p30250`). `make service-backed-test-slice OWNER=module.networkflow
ROWS=module.networkflow.integration.effective_resource_limit_discovery`: PASS
(`20260919T144420Z-p30493`). `make test-slice OWNER=module.networkflow
ROWS=module.networkflow.frontend_integration.verify_production_network_flow_grids_read_only_b_f662be335e`:
PASS (`20260919T144430Z-p42901`). `make frontend-typecheck`: PASS
(`20260919T144438Z-p52323`). Run roots are under `.cartulary/test-results/`.
Owner table, backend route, schema, generated decoder and browser agree on v2 and
current effective limits. Final schema/policy gates remain S-16. Next: S-13.

### S-13 — table/link application boundaries

Status: COMPLETE. Added private tableApplication and indicatorLinkApplication
with typed outcomes and separate receipt adapters. Table replay remains inside
incident serialization after admission and does not reread table lifecycle.
Indicator-link admission now precedes replay within the application itself,
including non-HTTP callers; target visibility and fresh transaction rechecks
remain. Participant results contain binding facts rather than HTTP payloads.
Reusable link/graph admission and participants now carry semanticFailure with
closed kinds and typed safe context; HTTP formatting is isolated in adapters.
Persisted-state graph validation calls the semantic decoder directly. Extended
the positive semantic dependency guard over the new applications and adapters.
No wire, receipt-byte, state, identity or browser-attempt migration.

`make format`: PASS (`20260919T144959Z-p54143`, `20260919T145010Z-p58531`).
Intermediate targeted compiles failed on moved helper/test signatures
(`20260919T145022Z-p62873`, `20260919T145116Z-p63496`); corrected references and
kept transport conversion in test adapters. `make test-slice OWNER=module.networkflow
ROWS=module.networkflow.unit.network_flow_selector_covers_linking_an_endpoint_cfa3b46c37,module.networkflow.unit.query_authoring_admission,module.networkflow.unit.table_lifecycle_contract`:
PASS (`20260919T145143Z-p64040`). `make service-backed-test-slice OWNER=module.networkflow
ROWS=module.networkflow.integration.table_lifecycle_admission_replay,module.networkflow.integration.bounded_graph_contributor_pipeline,module.networkflow.integration.extension_state_v4_v1_rejection`:
PASS (`20260919T145144Z-p64275`, 4/4 units). Run roots under
`.cartulary/test-results/`. These execute real table replay/role races, link
selector/target/replay and participant rollback, state admission and safe error
mapping. Broader captured-attempt browser regressions remain S-16. Next: S-14.

### S-14 — shared materialization payload admission

Status: COMPLETE. Added one private strictjson-based decoder for exactly five
fields, canonical enclosing incident identity, graph ID, positive int64 generation
and canonical source snapshot. Worker, restore and retained-job startup admission
all use it. Rejected payloads return the zero value, so job failure cannot select
a declaration from invalid bytes. Valid payload v1/state-4 bytes remain unchanged;
malformed retained state is rejected without rewriting it.

Added shared malformed-byte cases exercised through the actual worker handler,
including duplicate/unknown members, null/missing/wrong types, trailing data,
generation/identity errors and incident mismatch. Added startup unknown-member
rejection and later-page invalid-snapshot rollback to existing real fixtures.
`make format`: PASS (`20260919T145430Z-p87914`). `make test-slice OWNER=module.networkflow
ROWS=module.networkflow.unit.time_bucket_graph_backend,module.networkflow.unit.query_authoring_admission`:
PASS (`20260919T145444Z-p92274`). `make service-backed-test-slice OWNER=module.networkflow
ROWS=module.networkflow.integration.saved_graph_lifecycle_v2,module.networkflow.integration.saved_graph_cutover_v6`:
PASS (`20260919T145445Z-p92527`, 3/3 units), including 0/1/256/257/512/513 restore
pages, cancellation and whole-transaction rollback. Run roots under
`.cartulary/test-results/`. Next: S-15a.

### S-15a — functional fault controls

Status: COMPLETE. Harness §12.2.5/§12.2.9 now names the eight real boundaries,
import-unit/job correlation, and interruption/retry semantics. Published fault
control v2 and removed the v1 schema/attachment and hypothetical stages. Added
instance decorators for the Import facade/transaction, Jobs handler/observation,
and the atomic graph finalizer. The application facade installs decorators before
worker registration; only the harness profile imports controls and installs owned
process termination. Added reusable pre-start test composition. Product owners
import no registries. Postcommit errors preserve the committed outcome.

New real Import assertions cover all three precommit boundaries with error,
panic and context cancellation, the postcommit error, committed tables/outcomes/
audit deltas, and immutable replay. Worker assertions cover error/panic/context
cancellation, actual Jobs cancellation, publication rollback, recovery, and
postpublication replay. The packaged process fixture crashes at entry, precommit
and postpublication, restarts its own child against retained state, and checks
exact result counts and receipt identity. Guards, matching, one-shot consumption,
retired-token and unsupported-combination rejection remain covered.

Validation under `.cartulary/test-results/`:

- `make format`: PASS (`20260919T151242Z-p55548`, `20260919T151423Z-p88095`).
- `make test-slice OWNER=module.networkflow ROWS=module.networkflow.support_integration.network_flow_fault_route_disabled_by_default_8b0741837f`:
  PASS (`20260919T151325Z-p87351`).
- `make service-backed-test-slice OWNER=module.imports ROWS=module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873`:
  PASS (`20260919T150908Z-p69142`, 3/3).
- `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.saved_graph_lifecycle_v2`:
  PASS (`20260919T151122Z-p37283`, 3/3).
- `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.process.the_packaged_standalone_server_composes_the_netw_400a31ad27`:
  PASS (`20260919T151505Z-p11534`, 7/7, ordinary/harness prerequisite builds).
- `make json-shape-check`: PASS (`20260919T151442Z-p10922`); `git diff --check`: PASS.

Earlier new-fixture failures remain visible: Import column naming and assumed
terminal failures (`20260919T150503Z-p23773`); worker retry semantics and nested
response lookup (`20260919T150625Z-p46431`, `20260919T150909Z-p69367`); process
recovery polling initially omitted lease classification plus the retry cadence
(`20260919T151056Z-p9528`, `20260919T151248Z-p59874`). Corrected those assertions;
precommit simulated Commit failure also now closes its transaction before the
Imports recovery read, matching pgx Commit lifecycle. A concurrent formatter
invalidated a process-build source snapshot (`20260919T151424Z-p88267`, harness
failure); the stable-source rerun passed. No skipped required assertions. Next:
S-15b, with separate table-ID/cursor entropy and collision-safe allocation.

### S-15b — functional randomness and transaction-safe ID allocation

Status: COMPLETE. Added separate construction-time table-ID and cursor-nonce
readers, each serialized per instance and defaulting to crypto/rand. Harness
assembly supplies readers backed by the two retained streams. Table UUIDs map
exactly to 16 bytes; cursor nonces require exactly 12 bytes. Removed the other six
streams, token values/string consumer, and randomness v1 schema/attachment;
published v2 without aliases. Exhausted or wrong-kind armed streams fail closed.
Harness §12.2.6 prose now agrees with its exact 12-byte limit.

Table INSERT now uses `ON CONFLICT ON CONSTRAINT network_flow_tables_pkey DO
NOTHING RETURNING`; only pgx.ErrNoRows retries, at most eight times. Removed the
obsolete unique-violation retry helper. No migration, ID-format, encryption,
public-major or durable-state change. New PostgreSQL assertions prove collision
then success, eight collisions, exhaustion without partial tables/rows/audit, and
an unrelated unique-constraint error without retry. Actual HTTP pagination proves
nonce bytes and fail-closed exhaustion. Cursor evidence now explicitly includes
existing key-rotation tests, TTL, short entropy, and ordinary instance isolation.

`make format`: PASS (`20260919T152415Z-p91282`).
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.network_flow_selector_covers_cursor_continuation_0f6e2c662c,module.networkflow.support_integration.network_flow_randomness_route_disabled_by_defaul_a63c6ef32a`:
PASS (`20260919T152433Z-p95636`, 2/2). Semantic/query unit evidence also passed
(`20260919T152040Z-p49098`). `make service-backed-test-slice OWNER=module.networkflow
ROWS=module.networkflow.integration.pagination_recovery,module.networkflow.store.network_flow_selector_covers_only_active_and_sof_400db83232`:
PASS (`20260919T152302Z-p73113`, 4/4). Initial new fixtures used the wrong nested
cursor field and required middleware request IDs from a bare control mux
(`20260919T152041Z-p49341`); corrected those fixture checks. `make json-shape-check`:
PASS (`20260919T152434Z-p95857`). All roots under `.cartulary/test-results/`.
No required skips. Next: S-15c.

### S-15c — functional authorization transitions

Status: COMPLETE. Published auth-transition v2, removed v1 projection/attachment,
four unsupported checkpoints, extension-claim removal, hidden-response tokens and
both synthetic response-selection fields. Harness §12.2.7 now specifies the
fixture driver and verified reference binding before arming. Because Incidents
has no soft-delete state, the adopted v2 vocabulary uses `incident_deleted` for
removing a disposable owned fixture incident; no product deletion API or weak
referential integrity was introduced.

Added networkflowsupport.AuthorizationFixture. It resolves real membership,
actor/session ownership and incident/table ownership, prevents symbolic bindings
from being reassigned to different rows, consumes exact scope/correlation before
changing state, and issues no response. Fixtures then make ordinary authenticated
HTTP requests. Registry arming rejects unresolved references; a fixture cannot
complete with pending required transitions. The integration matrix covers all
six resource kinds, membership loss/restoration, hidden-response details,
continuation authorization, table rename/deletion, incident absence and session
revocation. Existing table/link transaction-race assertions remain selected.

`make format`: PASS (`20260919T153313Z-p30660`). `make test-slice
OWNER=module.networkflow ROWS=module.networkflow.support_integration.network_flow_auth_transition_route_disabled_by_d_07779f71d9`:
PASS (`20260919T153133Z-p12142`). `make service-backed-test-slice
OWNER=module.networkflow ROWS=module.networkflow.integration.table_lifecycle_admission_replay,module.networkflow.integration.bounded_graph_contributor_pipeline`:
PASS (`20260919T153351Z-p35128`, 3/3). `make json-shape-check`: PASS
(`20260919T153353Z-p35346`). Roots under `.cartulary/test-results/`.
The initial fixture attempted to delete a product-created incident protected by
its audit FK (`20260919T153134Z-p12363`); retained that protection and corrected the
deletion case to seed a disposable incident without retained references. No
required skips. Product authorization/status/privacy and session semantics remain
unchanged. Next: S-15d, committed audit consumers.

### S-15d — committed audit assertion consumers

Status: COMPLETE. Published audit-assertion v2 and removed its v1 schema/attachment,
unsupported import resource and redundant expected_replay_increment field.
Harness §12.2.8/§12.2.9 adopts exact event/resource pairing, verified fixture
binding, full actor/incident/operation/event/resource/correlation matching, and
replay only for adopted idempotent operations (never graph query). Projection
conditions now express the fault combinations and audit event/resource rules.

Added Auth-owned committed occurrence reader and networkflowsupport.AuditFixture.
It reads the real pre-operation baseline, binds actor/incident and transaction or
request correlation, consumes an exact assertion, reads committed records after
HTTP or durable completion, resolves newly allocated resource IDs from the real
operation, and optionally verifies replay produces no increment. Failed assertions
remain failed after consumption; missing or pending required assertions cannot
report success. No product audit event or payload was changed.

Real fixtures exercise all six event codes: Imports table creation under fault,
rollback/recovery and replay; table rename/delete and replay; graph query; binding
creation/reuse and replay. They cover denied and rolled-back zero counts, no-op
rename, a nonzero observed baseline, and concurrent actors/incidents sharing a
request correlation. Deliberately wrong baseline/final counts and pending
assertions prove fixture failure. Shared strict control decoding now rejects
missing/null/duplicate/unknown fields and trailing JSON across all four families,
after unchanged route guards.

`make format`: PASS (`20260919T154547Z-p1294`). `make test-slice
OWNER=module.networkflow ROWS=module.networkflow.support_integration.network_flow_audit_assertion_route_disabled_by_d_611b731957,module.networkflow.support_integration.network_flow_auth_transition_route_disabled_by_d_07779f71d9,module.networkflow.support_integration.network_flow_randomness_route_disabled_by_defaul_a63c6ef32a,module.networkflow.support_integration.network_flow_fault_route_disabled_by_default_8b0741837f`:
PASS (`20260919T154622Z-p5775`, one grouped Go unit).
`make service-backed-test-slice OWNER=module.networkflow
ROWS=module.networkflow.integration.bounded_graph_contributor_pipeline`: PASS
(`20260919T154623Z-p6012`, 3/3, including negative-evidence cases).
`make service-backed-test-slice OWNER=module.imports
ROWS=module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873`: PASS
(`20260919T154331Z-p65123`, 3/3). `make json-shape-check`: PASS
(`20260919T154624Z-p6234`; final projection-condition rerun recorded in S-16).
Roots under `.cartulary/test-results/`. No required failures or skips remain in
this slice. Next: S-16, final validation and handoff completion.

### S-16 — validation and handoff completion

Status: COMPLETE. All F-18–F-23 dispositions and S-09–S-15d exit criteria are
closed by this iteration's evidence. The seven current handoffs in §15.10 and
completion assessment in §15.11 replace their planning-only statuses. Historical
findings, IDs, ledgers and command results remain. No operational deployment,
telemetry cutover, commit or index mutation was performed.

Final assertion review added the same malformed byte corpus to startup and
restore consumers, alongside the existing actual worker corpus. The valid
startup control proves that rejection is reached after registration/receipt
identity checks and before reading a commit proof. Rejected restore bytes cannot
reach job reconciliation. Byte-level doubles deliberately preserve duplicate
JSON members that PostgreSQL JSONB normalizes; live tests separately retain valid
restore sizes 0/1/256/257/512/513 and whole-transaction later-page rollback.
`make format` passed (`20260919T155323Z-p27075`); `make test-slice
OWNER=module.networkflow ROWS=module.networkflow.unit.time_bucket_graph_backend`
passed (`20260919T155339Z-p32339`).

The first final `make check` found two now-unused HTTP forwarding helpers in
`semantic_http_adapter.go`: `composeGraphSourceFromSemanticHTTP` and
`graphProjectionFailedForContextHTTP` (staticcheck U1000, related to S-13).
Reference closure found only their definitions, so both were removed; no aliases
or replacement algorithm was introduced. The failed run is
`20260919T155535Z-p6497`; its lint diagnostics are in
`unit-logs/target-lint-go/stdout.log` and its canonical result in `run-summary.json`.
`make explain-run RESULTS_DIR=.cartulary/test-results/20260919T155535Z-p6497`
confirmed 936 passed, one failed and no skipped/cancelled units. After deletion,
`make format` passed (`20260919T160549Z-p10086`), `make lint-go` passed (exit 0),
and `make test-slice OWNER=module.networkflow
ROWS=module.networkflow.unit.query_authoring_admission,module.networkflow.unit.time_bucket_graph_backend`
passed (`20260919T160607Z-p14772`). `make agent-finalize` passed again with
RESULTS_DIR unset (`20260919T160633Z-p27090`) before the successful full recheck.

#### Final routing and command evidence

For each of `module.networkflow`, `module.imports`, `platform.jobs`,
`module.recovery`, `module.reporting`, `module.graphprojection`, `app.server`,
`web.networkflow`, and `harness.browser`, both
`make task-guide ROLE=module-author OWNER=<owner>` and
`make explain-test-owner OWNER=<owner>` passed before selection. Existing authored
rows already route the strengthened assertions through their actual entrypoints;
no synthetic row or docs-derived evidence was added. Reporting remains owner of
exact-result lease evidence.

All run IDs below are beneath `.cartulary/test-results/`; each run contains its
`run-summary.json`, manifest and unit logs/artifacts. Work-unit counts include
prerequisites and grouped runners; they are not assertion counts.

| Exact Make command | Result | Run root |
| --- | --- | --- |
| `make agent-finalize` | PASS 1/1; RESULTS_DIR unset, retained-run maintenance skipped | `20260919T155401Z-p35862` |
| `make generate-drift` | PASS 4/4 | `20260919T155435Z-p39997` |
| `make generated-artifact-policy-check` | PASS 3/3 | `20260919T155435Z-p40006` |
| `make json-shape-check` | PASS 3/3; also final S-15d projection-conditions pass at `20260919T154801Z-p24611` | `20260919T155435Z-p40013` |
| `make migration-drift` | PASS 5/5 | `20260919T155435Z-p40025` |
| `make harness-contract` | PASS 2/2 | `20260919T155435Z-p40230` |
| `make frontend-typecheck` | PASS 2/2 | `20260919T155435Z-p40193` |
| `make frontend-import-boundary-check` | PASS 2/2 | `20260919T155435Z-p40205` |
| `make build-server` | PASS 4/4 | `20260919T155435Z-p40380` |
| `make build-server-harness` | PASS 4/4 | `20260919T155435Z-p40386` |
| `make test-slice OWNER=web.networkflow` | PASS 66/66; captured table/link/graph attempts and v2 decoding included | `20260919T155435Z-p40092` |
| `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.browser_stateful.exploration_navigation,module.networkflow.browser_stateful.pagination_recovery,module.networkflow.browser_stateful.saved_graph_deferred_navigation,module.networkflow.browser_stateful.saved_graph_exact_result_lifecycle,module.networkflow.browser_stateful.saved_graph_read_recovery,module.networkflow.browser_stateful.verify_protected_network_analysis_state_is_disca_21a5de1ebf` | PASS 15/15; all six stateful scenarios | `20260919T155456Z-p49148` |
| `make service-backed-test-slice OWNER=module.reporting ROWS=module.reporting.integration.exact_graph_result_lease_lifecycle_8f1c5c43a2` | PASS 3/3 | `20260919T155455Z-p48602` |
| `make service-backed-test-slice OWNER=platform.jobs ROWS=platform.jobs.integration.claim_recovery_and_publication,platform.jobs.integration.runner_and_recovery,platform.jobs.integration.runner_failure_security,platform.jobs.integration.lifecycle_and_progress` | PASS 3/3 | `20260919T155547Z-p30418` |
| `make service-backed-test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.result_v2_atomicity_and_queries,module.graphprojection.storage.result_v2_cleanup_locking,module.graphprojection.storage.restore_publication` | PASS 5/5 | `20260919T155640Z-p92565` |
| `make service-backed-test-slice OWNER=module.recovery ROWS=module.recovery.integration.restore_readiness_selects_the_latest_retained_su_e71d710085,module.recovery.integration.restore_target_serving_admission_cancels_mutatio_f73b480d23,module.recovery.integration.selected_backup_restore_fails_before_readiness_w_3a6ccb7d7a` | PASS 4/4 | `20260919T155734Z-p21159` |
| `make service-backed-test-slice OWNER=app.server ROWS=app.server.integration.extension_application_process_lease_43130392c4,app.server.integration.server_shared_and_restore_exclusive_serving_leas_1da2a38098,app.server.process.the_standalone_server_keeps_harness_routes_disab_4dc8b1f7d9` | PASS 8/8 | `20260919T155849Z-p53452` |
| `make service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.integration.runtime_reset_recovery_purpose_contract,harness.browser.integration.postgres_cleanup_target_scoped_coordination` | PASS 4/4 | `20260919T160003Z-p3455` |
| `make check` | PASS 937/937; no failures or skips | `20260919T160653Z-p31022` |
| `make lint-markdown` | PASS; scoped documentation maintenance | `20260919T161243Z-p37068` (`adhoc/lint-markdown/tool-run-summary.json`) |

No required assertion is skipped or failing. Retained-run maintenance is the
only finalization skip: the successful final check was not retroactively supplied
to the earlier `agent-finalize`. Initial fixture mistakes/failures in S-10 and
S-15 and the S-16 lint failure remain visible alongside their passing replacements.
Broader release publication and deployment were not requested or represented as
repository verification.

#### Retained harness token → consumer → product assertion closure

All fault tokens below carry the `network_flow.` prefix. All use
`harnesscontrol/fault_consumers.go`, installed by harness application assembly;
production owners do not import the registries. NF §29.1 and the existing Imports,
Jobs and GP atomicity owners define product outcomes; Harness §§12.2.5/12.2.9
define control mechanics.

| Retained fault boundary | Real boundary and assertion | Routed evidence |
| --- | --- | --- |
| `import.before_owner_apply` | Import facade before owner writes; error/panic/context cancellation, no partial effects, recovered/replayed outcome | Imports atomic-unit row, S-15a/S-15d |
| `import.after_owner_apply` | Same transaction after owner writes; all three precommit effects roll back | Imports atomic-unit row, S-15a/S-15d |
| `import.before_transaction_commit` | Decorated actual transaction commit; rollback closes the transaction and Jobs can recover | Imports atomic-unit row, S-15a/S-15d |
| `import.after_transaction_commit_before_reply` | Error only after actual commit; table/audit/receipt remain, replay has no extra occurrence | Imports atomic-unit row, S-15a/S-15d |
| `worker.before_handler_start` | Real registered handler; error/panic/context cancellation and owned child crash/restart | NF saved-graph lifecycle and packaged-process rows, S-15a |
| `worker.before_cancellation_check` | Real Jobs observation; precommit effects and Jobs cancellation produce owner-defined terminal outcomes | NF saved-graph lifecycle, S-15a |
| `worker.before_final_commit` | Transaction participant before atomic finalization; rollback and owned-process crash recovery | NF lifecycle and packaged-process rows, S-15a |
| `worker.after_completed_publication` | Completed finalization; error/crash preserve selected result and immutable receipt on restart/replay | NF lifecycle and packaged-process rows, S-15a |

| Retained randomness token / kind | Consumer and assertion | Routed evidence |
| --- | --- | --- |
| `network_flow.table_id` / `uuid` | Store allocation via its own reader; collision-success, eight collisions, exhaustion, unrelated constraint error, transaction remains usable | NF store row, S-15b |
| `network_flow.cursor_nonce` / `hex_bytes` | Cursor encryption via a separate reader; exact nonce, length, exhaustion, TTL/rotation and instance isolation | NF pagination service and cursor unit rows, S-15b |

Authorization uses `networkflowsupport.AuthorizationFixture.Before`, then an
ordinary authenticated HTTP request. Both retained boundaries
`network_flow.route.before_authorization` and
`network_flow.cursor.before_authorization_recheck` have actual consumers. All six
resource kinds (incident, table, workspace, graph, contributors, cursor) exercise
membership loss/restoration with real responses and empty hidden-resource details.
NF/Core authorization owners retain admission semantics; Harness §12.2.7 owns
fixture mechanics.

| Retained transition kind | State change and assertion | Routed evidence |
| --- | --- | --- |
| `incident_membership_revoked` | Delete exact membership; hidden resources and continuation denial, unrelated scope stays pending | NF table-lifecycle row, S-15c |
| `incident_membership_restored` | Restore verified prior membership; real requests succeed again | Same |
| `session_revoked` | Auth-owned session revocation; actual session-required response | Same |
| `incident_deleted` | Remove disposable owned fixture incident; real not-found and other incident unaffected, FK protection retained | Same |
| `network_flow_table_renamed` | Change exact table metadata; new name visible, continuation remains valid | Same |
| `network_flow_table_soft_deleted` | Change exact lifecycle; actual table-not-active continuation response | Same |

Audit uses `networkflowsupport.AuditFixture` and Auth-owned
`storetest.CountAuditOccurrences`. Exact scope includes actor, incident, operation,
event, resource and correlation. Harness §12.2.8/12.2.9 governs mechanics; NF §16
retains product events. Assertion kinds `exact_count`, `zero_occurrences` and
`no_audit_replay` each have positive and negative fixture evidence, including
nonzero baseline, incorrect counts, pending assertions and concurrent isolation.

| Retained event code / resource | Committed occurrence assertion | Routed evidence |
| --- | --- | --- |
| `network_flow_table_created` / table | Real Imports apply under rollback/recovery/postcommit faults and immutable replay | Imports atomic-unit row, S-15d |
| `network_flow_table_renamed` / table | Exact committed increment, no-op/denied/rolled-back zero, nonzero baseline and replay | NF graph/link row, S-15d |
| `network_flow_table_soft_deleted` / table | Exact committed increment and replay silence | Same |
| `network_flow_graph_query_executed` / graph | Real query count; replay assertion rejected because no adopted replay | Same |
| `network_flow_indicator_binding_created` / binding | Real binding creation and receipt replay count | Same |
| `network_flow_indicator_binding_reused` / binding | Real reuse and receipt replay count | Same |

All four support rows additionally prove guards before decoding, malformed control
rejection, duplicates/conflicts, mismatches, consume-once and clearing. These are
mechanics evidence, kept separate from the real effects above. Exact selectors
and successful artifacts appear in their individual S-15 records.

#### Owner/projection review, removal closure and release handoff

Human review confirms NF 6.0.1 §17.3, route declarations, backend response,
authored schemas, generated validators and frontend decoding all use sole
source-profile v2. Shared SourceProfile remains; only SourceProfileList and its
exclusive EffectiveLimits definition were retired. Valid product major/state,
identity, digest, receipt, audit-event, lease and recovery contracts did not change.
The Graph Projection NLSpec and domain vocabulary/navigation were reviewed and
need no amendment. Jobs execution, GP results, Reporting leases and Recovery
coordination remain with their existing owners.

Harness v2 schema attachments, closed tokens, condition rules and consumers agree
with §§12.2.5–12.2.9. Final editorial review removed stale hidden-response/replay-field
wording, made the existing audit duplicate-scope clause explicitly include actor
and incident, and kept the preexisting Reporting paragraph under NF §29 rather
than nesting it under new §29.1. No new runtime requirement was introduced by
these final editorial adjustments.

The [release notes](networkflow-f18-f23-release-notes.md) identify every removed
helper/export/schema/token family and its caller/fixture migration. The final
reference search found no live retired source-profile v1 or obsolete helper
caller; the source-profile v1 string remains solely in its rejection fixture.
Build-tag and external-test closure is also backed by ordinary/harness builds and
real external integration tests. Test-only store conveniences remain `_test.go`;
no production forwarding alias or database migration was added.

Rollout pairs code with regenerated Go/TypeScript and v2 harness fixtures.
Rollback must use compatible code retaining explicit keys and strict state
admission; it cannot restore the engineering fallback or silently rewrite retained
jobs. Deployment and the prior telemetry cutover remain external and unevidenced
by this repository run.

#### Changed-file inventory for this iteration

Authored source, test support, owner documents, contracts and generated outputs
are listed together below; generated files were produced through Make.

```text
apps/web/src/networkFlow/NetworkAnalysisWorkspace.test.tsx
contracts/network-flow/index.json
contracts/network-flow/schemas.v3.json
docs/handoffs/networkflow-f18-f23-release-notes.md
docs/handoffs/networkflow-module-refactor-tracker.md
docs/network-flow-activity-nlspec.md
docs/testing-harness-nlspec.md
internal/app/server/runtime_assembly.go
internal/app/server/server_profile_harness.go
internal/app/serverprocess/networkflow_fault_recovery_process_test.go
internal/app/serverprocess/networkflow_runtime_routes_process_test.go
internal/gen/contractimports/artifacts_gen.go
internal/gen/contractnetworkflow/artifacts_gen.go
internal/gen/importtargetregistry/registry_gen.go
internal/modules/auth/testsupport/storetest/audit_occurrences.go
internal/modules/imports/imports_integration_test.go
internal/modules/imports/network_flow_harness_fault_test.go
internal/modules/networkflow/application.go
internal/modules/networkflow/digest.go
internal/modules/networkflow/entropy.go
internal/modules/networkflow/extension_state.go
internal/modules/networkflow/graph_materialization_payload.go
internal/modules/networkflow/graph_materialization_payload_test.go
internal/modules/networkflow/graph_materialization_timeout_test.go
internal/modules/networkflow/graph_restore_source.go
internal/modules/networkflow/graph_view_jobs.go
internal/modules/networkflow/harness_audit_integration_test.go
internal/modules/networkflow/harness_authorization_integration_test.go
internal/modules/networkflow/harness_randomness_integration_test.go
internal/modules/networkflow/harness_worker_integration_test.go
internal/modules/networkflow/harnesscontrol/control_json.go
internal/modules/networkflow/harnesscontrol/control_json_test.go
internal/modules/networkflow/harnesscontrol/controls.go
internal/modules/networkflow/harnesscontrol/fault_consumers.go
internal/modules/networkflow/harnesscontrol/network_flow_audit_assertion.go
internal/modules/networkflow/harnesscontrol/network_flow_audit_assertion_integration_test.go
internal/modules/networkflow/harnesscontrol/network_flow_auth_transition.go
internal/modules/networkflow/harnesscontrol/network_flow_auth_transition_integration_test.go
internal/modules/networkflow/harnesscontrol/network_flow_fault.go
internal/modules/networkflow/harnesscontrol/network_flow_fault_integration_test.go
internal/modules/networkflow/harnesscontrol/network_flow_randomness.go
internal/modules/networkflow/harnesscontrol/network_flow_randomness_integration_test.go
internal/modules/networkflow/harnesscontrol/randomness_consumers.go
internal/modules/networkflow/indicator_link.go
internal/modules/networkflow/indicator_link_admission.go
internal/modules/networkflow/indicator_link_boundary_test.go
internal/modules/networkflow/indicator_link_fixture_test.go
internal/modules/networkflow/indicator_link_graph_admission.go
internal/modules/networkflow/indicator_link_http.go
internal/modules/networkflow/indicator_link_receipts.go
internal/modules/networkflow/integration_surface_test.go
internal/modules/networkflow/keyring.go
internal/modules/networkflow/keyring_test.go
internal/modules/networkflow/module.go
internal/modules/networkflow/network_flow_behavior_test.go
internal/modules/networkflow/network_flow_contract_test.go
internal/modules/networkflow/network_flow_unit_test.go
internal/modules/networkflow/query.go
internal/modules/networkflow/query_sql_test_bridge_test.go
internal/modules/networkflow/routes.go
internal/modules/networkflow/routes_integration_test.go
internal/modules/networkflow/saved_graph_admission.go
internal/modules/networkflow/saved_graph_admission_integration_test.go
internal/modules/networkflow/security.go
internal/modules/networkflow/semantic_boundary_test.go
internal/modules/networkflow/semantic_failure.go
internal/modules/networkflow/semantic_http_adapter.go
internal/modules/networkflow/semantic_http_errors.go
internal/modules/networkflow/store.go
internal/modules/networkflow/store_fixture_test.go
internal/modules/networkflow/store_test.go
internal/modules/networkflow/table_http.go
internal/modules/networkflow/table_lifecycle_integration_test.go
internal/modules/networkflow/table_receipts.go
internal/modules/networkflow/timestamp.go
internal/modules/networkflow/timestamp_test.go
internal/modules/networkflow/transaction_participants.go
internal/testutil/appsupport/runtime.go
internal/testutil/httptestx/httptestx.go
internal/testutil/networkflowsupport/audit.go
internal/testutil/networkflowsupport/authorization.go
internal/testutil/networkflowsupport/graph_source.go
packages/protocol-ts/src/generated/import-target-registry.ts
packages/protocol-ts/src/generated/network-flow-types.ts
packages/protocol-ts/src/generated/network-flow-validators.ts
tools/harness_schema_attachments.json
tools/schemas/cartulary.test.network_flow_audit_assertion_control.v1.schema.json
tools/schemas/cartulary.test.network_flow_audit_assertion_control.v2.schema.json
tools/schemas/cartulary.test.network_flow_auth_transition_control.v1.schema.json
tools/schemas/cartulary.test.network_flow_auth_transition_control.v2.schema.json
tools/schemas/cartulary.test.network_flow_fault_control.v1.schema.json
tools/schemas/cartulary.test.network_flow_fault_control.v2.schema.json
tools/schemas/cartulary.test.network_flow_randomness_control.v1.schema.json
tools/schemas/cartulary.test.network_flow_randomness_control.v2.schema.json
```

The Markdown target's current globs omit this tracker and the new release notes.
Their table structure, local links/anchors and heading progression were reviewed
separately; these are documentation checks, not product evidence.
Tracker review checked all seven current tables, local links/anchors, unique slice
headings, preserved historical sections/findings and unchanged staging. The final
index SHA-256 equals the S-09 basis:
`00aaecb0cd7529b2e468f41b88150501254481b2bf014a8f2c28c32371a7ffcd`.
`git diff --check` passes. Markdown remains outside runtime, generator, test,
conformance and release-evidence inputs. Final tracker edits after validation
record results and completion only.
