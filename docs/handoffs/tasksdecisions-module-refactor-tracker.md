# tasksdecisions Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

This tracker governs execution of the `tasksdecisions` refactor. It MUST NOT
supersede an adopted product owner, Core requirement, subsystem NLSpec, or
machine projection of an adopted owner. When this tracker and an adopted owner
conflict, the adopted owner controls and the session MUST record
`BLOCKED: owner contradiction` without choosing a side.

| Item | Normative posture |
| --- | --- |
| Target path | `internal/modules/tasksdecisions` |
| Target label | `tasksdecisions`, derived from the target directory and normalized to lowercase kebab case |
| Output path | `docs/handoffs/tasksdecisions-module-refactor-tracker.md` |
| Current status | `RS-08A` remains complete historical work. Iteration 2 remediation is complete against bound commit `dd06b39a1ec901b0c610c87606546324552693da`; every `RM-00..RM-11` row and acceptance criterion is `DONE`/`PASS`. |
| Administrative authorization | The 2026-08-25 implementation request closes `RB-001` against repository commit `bbb0fb7298e746210a5b8123d4d251aece8e585e` and pre-revision tracker SHA-256 `aa2f8e4f725afeb4775b915ae7b9caf54f517211470d84265ea5929b0a1093c2`. |
| Presently authorized production change | The serial `RM-00..RM-11` remediation ledger in Section 13 is complete; no additional production change is authorized by this tracker. |
| Pending planned work | None within the authorized remediation iteration. |
| Non-goals | No public route, wire schema, lifecycle, authorization, event, transaction, database schema, frontend behavior, or persisted-data change; no generalized review/approval engine and no compatibility wrapper for intentionally retired internal Go APIs. |
| Generated files | Generated outputs MAY change only through `make generate`; generated roots MUST NOT be hand-edited. |

The following states are distinct and MUST NOT be conflated:

| State | Meaning | Current disposition |
| --- | --- | --- |
| Administrative authorization | An explicit request authorizes a bounded implementation against identified repository and tracker revisions. | Closed for `RS-08A` and Iteration 2; unplanned observable behavior change remains unauthorized. |
| Baseline release | The owner, service-backed, and backend-boundary baselines pass before production edits. | Released for Iteration 2 by `RM-00` evidence in Section 13. |
| Slice progress | One independently revertible slice has its specified implementation and evidence. | Every `RM-00..RM-11` slice is checkpointed `DONE`. |
| Refactor completion | Every required slice is `DONE` or a justified no-op and final handoff evidence is complete. | Reached by the `RM-11` validation and handoff checkpoint. |

The source hierarchy is:

1. adopted subsystem NLSpecs within their named scopes;
2. Core 00 through Core 04 for implementation-conformance behavior;
3. Core 05 only for claim-bearing timed or fixture-sensitive publication;
4. domain vocabulary and implementation-support guides;
5. current repository code and tests;
6. prior plans, handoffs, research, and the planning framework as evidence only.

Core 05 is not applicable because this work publishes no timed or
fixture-sensitive claim. No owner contradiction was found.

Owner and doctrine documents inspected are
`docs/handoffs/cartulary_modular_refactor_planning_framework.md`,
`docs/domain.md`, Core 00 through Core 04 under `docs/spec/`, the adopted
Workbook, Projections, and Revisions module-boundary decisions,
`docs/reporting-subsystem-nlspec.md`, `docs/testing-harness-nlspec.md`, and
`docs/guides/cartulary-dev-guide.md`. `temp/analysis-notes.md` is review
evidence, and `docs/research/nlspec-spec.md` is writing guidance; neither is a
runtime or product authority.

Repository evidence inspected includes every file in the target; exact
Workbook, Imports, Projections, Revisions, Recovery, Incident Portability,
Reporting, Collaboration, and server assembly callers; the Workbook routes and
failure vocabulary; canonical view and generated-contract inputs; the backend
boundary policy; and Tasks/Decisions verification routing.

Planning finding: the generic framework anticipates possible cross-module
extraction, while the live repository already enforces source-owner/provider
separation through application assemblies and boundary manifests.
`tasksdecisions` is therefore a legitimate Coordination source owner and
mutation coordinator, not an accidental catch-all. Its provider contributions
MUST remain owner-local unless later adopted owner evidence says otherwise.

The governing execution requirements are:

| Requirement ID | Normative requirement |
| --- | --- |
| TD-REQ-001 | Implementation MUST preserve adopted-owner behavior and MUST stop on an owner contradiction. |
| TD-REQ-002 | A session MUST distinguish authorization, baseline release, slice progress, and full refactor completion. |
| TD-REQ-003 | Public HTTP shapes, authorization precedence, normalized hashes, mutation effects, storage, revisions, projections, and Collaboration events MUST remain unchanged by `RS-08A`. |
| TD-REQ-004 | Route authentication, authorization, incident visibility, and incident-open checks MUST remain outside Tasks/Decisions semantic admission. Rejected admission MUST produce zero mutation effects. |
| TD-REQ-005 | One caller-owned transaction MUST continue to contain source writes, links, revisions, projection refresh, idempotency, and Collaboration intent publication. |
| TD-REQ-006 | Generated files MUST NOT be hand-edited. An authored harness change MUST be refreshed only by `make generate` and proved by drift and policy checks. |
| TD-REQ-007 | Tasks/Decisions production code MUST NOT import `net/http`, `internal/platform/httpapi`, or `internal/platform/httpauth`. |
| TD-REQ-008 | Harness rows MUST be treated only as verification accounting. An authored row MAY change only when its selected test topology actually changes. |
| TD-REQ-009 | A failing affected product assertion, boundary rule, generated drift check, or owner contradiction MUST stop the active slice. Infrastructure failures MUST be repaired and rerun; they MUST NOT be reported as product passes. |
| TD-REQ-010 | Each slice MUST have one rollback unit and MUST record exact changed files, commands, retained result roots, failures, skipped checks, and completion evidence. |

## 2. Current-State Repository Inventory

Every one of the 38 current target files is inventoried below. No target file
is out of scope.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/tasksdecisions/admission_failure.go` | Closed, transport-neutral semantic admission failure. | `AdmissionFailure`; `Error`, `Field`, `ReasonCode`, `CollectionFieldKey`, `CountLimit`. | Owner admission functions and Workbook assembly adapters. | Standard library only. | Admission failure and adapter translation tests. | Public invalid-payload detail vocabulary indirectly. | tasksdecisions | high | Added by `RS-08A`; state and constructors are private, optional accessors fail closed. |
| `internal/modules/tasksdecisions/conflict_snapshot.go` | Selects revision snapshot projection for Task Request and Decision conflicts. | Package-private projector. | Mutation contribution and Workbook adapters. | Conflict tokens and Revisions projection contracts. | Mutation composition and Workbook conflict tests. | Conflict payload and revision snapshot behavior. | tasksdecisions | medium | Keep with source-owner conflict semantics. |
| `internal/modules/tasksdecisions/decision_reference_admission_test.go` | Characterizes exact stable-ID admission for direct Decision references. | Test-only surface. | Owner test family. | Admission and UUID handling. | Its own unit test. | Decision writeback contract indirectly. | tasksdecisions tests | low | Updated to the new admission API. |
| `internal/modules/tasksdecisions/import_create.go` | Implements Task Request and Decision create semantics in an Imports-owned transaction. | `ImportCreateCommand`, capability ports, `ImportDependencies`, `NewImportContribution`. | Import assembly owner registry. | Imports, PostgreSQL, Records, Links, Revisions, Projections, Collaboration, policy. | Composition, import-owner, service-backed import tests. | Import view targets and both view contracts. | tasksdecisions | high | Imports orchestrates; source semantics remain here. |
| `internal/modules/tasksdecisions/incident_bundle_contribution.go` | Exposes owner Incident Bundle source and subtype contributions. | `IncidentBundleContribution`, `NewIncidentBundleContribution`. | Incident portability assembly/catalog. | Owner-local providers. | Source-port and assembly tests. | Incident Bundle source catalog. | tasksdecisions | medium | Thin owner contribution. |
| `internal/modules/tasksdecisions/incident_bundle_source_port_test.go` | Proves portability round-trip, invariants, ordering, diagnostics, and atomicity. | Test-only helpers. | Owner test family. | Portability coordinator, source port, PostgreSQL support. | Its own service-backed tests. | Bundle family and invariant IDs. | tasksdecisions tests | low | Verification evidence only. |
| `internal/modules/tasksdecisions/internal/policy/policy.go` | Owns vocabulary, create/direct-patch validation, lifecycle machines, references, and portability invariant IDs. | Private typed policy surface. | Root mutation/import and rollback/portability providers. | Standard library. | Admission, store, rollback, portability tests. | Core 02 and view enums. | tasksdecisions | high | Semantic center; MUST NOT move to transport or storage. |
| `internal/modules/tasksdecisions/internal/providers/deleterestore/provider.go` | Supplies source snapshot and lifecycle hooks to generic delete/restore. | `TaskRequestSource`, `DecisionSource`, constructors and provider implementations. | Revision contribution/assembly. | Revisions contract, source readers, PostgreSQL. | Revision/delete-restore tests. | Revision provider catalog. | tasksdecisions | medium | Source-owned adapter. |
| `internal/modules/tasksdecisions/internal/providers/incidentbundle/portability.go` | Encodes, decodes, validates, and applies portable owner rows. | Private prepared/portable row types and helpers. | Owner source port. | Incident portability, policy, SQL transaction. | Source-port tests. | `data/task_requests.ndjson`, `data/decisions.ndjson`, invariant IDs. | tasksdecisions | high | Family semantics remain owner-local. |
| `internal/modules/tasksdecisions/internal/providers/incidentbundle/source_port.go` | Builds source descriptor and validates prepared rows/references. | Private `NewSourcePort` and helpers. | Root Incident Bundle contribution. | Incident portability contract and source queries. | Source-port and assembly tests. | Incident Bundle source catalog. | tasksdecisions | high | Coordinator-independent source port. |
| `internal/modules/tasksdecisions/internal/providers/incidentbundle/subtype_presence.go` | Reports Task Request and Decision subtype bindings. | Private `SubtypeContribution`. | Root Incident Bundle contribution. | Subtype-presence contract and PostgreSQL. | Portability and assembly tests. | Supported bundle record types. | tasksdecisions | medium | Correctly owner-local. |
| `internal/modules/tasksdecisions/internal/providers/projection/decision_source.go` | Reads authoritative Decision state and relations into typed projection input. | `DecisionSource`, constructor, typed reads. | Projection contribution. | PostgreSQL, projection input types, source/link/record tables. | Projection and service-backed mutation tests. | Decision provider descriptor/view. | tasksdecisions | high | Projections owns physical storage. |
| `internal/modules/tasksdecisions/internal/providers/projection/task_source.go` | Reads authoritative Task Request state and relations into typed projection input. | `TaskRequestSource`, constructor, typed reads. | Projection contribution. | PostgreSQL, projection input types, source/link/record tables. | Projection and service-backed mutation tests. | Task provider descriptor/view. | tasksdecisions | high | Projections owns physical storage. |
| `internal/modules/tasksdecisions/internal/providers/reporting/provider.go` | Selects Reporting facts from typed Task projections. | Private `CollectFactsTx`. | Root Reporting contribution. | Reporting provider contract and Task projection reader. | Reporting/server composition tests. | Reporting provider key. | tasksdecisions | medium | Reporting owns export orchestration. |
| `internal/modules/tasksdecisions/internal/providers/rollback/provider.go` | Validates and restores Task Request and Decision snapshots. | Provider types, constructors, rollback methods. | Revision contribution/assembly. | Revisions rollback, policy, PostgreSQL. | Provider and revision tests. | Revision snapshot schemas. | tasksdecisions | high | Preserves source lifecycle/reference invariants. |
| `internal/modules/tasksdecisions/internal/providers/rollback/provider_test.go` | Proves nullable Task state and invalid-owner rejection. | Test-only surface. | Owner test family. | Rollback providers. | Its own unit tests. | Revision snapshot semantics indirectly. | tasksdecisions tests | low | Verification evidence only. |
| `internal/modules/tasksdecisions/internal/source/direct_updates.go` | Applies allowlisted direct changes and classifies uniqueness failures. | Private update/error helpers. | Root mutation helpers. | PostgreSQL and policy. | Admission and service-backed mutation tests. | Writable field set/error behavior. | tasksdecisions | high | Owner-private SQL. |
| `internal/modules/tasksdecisions/internal/source/facts.go` | Loads lifecycle facts and validates envelopes, supersession, references, and links. | Private fact/validation functions. | Mutation/supersession and portability/rollback. | Source, record, and link tables. | Store, supersession, portability, rollback tests. | Lifecycle/link/supersession contracts. | tasksdecisions | high | Authoritative invariants. |
| `internal/modules/tasksdecisions/internal/source/inserts.go` | Inserts typed Task Request and Decision source rows. | Private insert functions. | Create and import-create paths. | PostgreSQL and policy. | Mutation/import tests. | Source schema and create contracts. | tasksdecisions | high | Owner-private persistence adapter. |
| `internal/modules/tasksdecisions/internal/source/lifecycle.go` | Touches rows, normalizes Task lifecycle, and applies Decision supersession. | Private lifecycle functions. | Patch and supersede paths. | PostgreSQL and policy. | Lifecycle, supersession, rollback, import tests. | Core 02 lifecycle/supersession. | tasksdecisions | high | Effect ordering is observable. |
| `internal/modules/tasksdecisions/internal/source/references.go` | Validates incident-member users and typed target records. | Private reference validators. | Mutation/import and capability adapter. | Records and incident membership. | Reference, mutation, import, portability tests. | Reference fields and validation precedence. | tasksdecisions | high | Source validity, not route authorization. |
| `internal/modules/tasksdecisions/mutation_admission.go` | Strictly admits create, patch, conflict-resolution, and supersede JSON and computes canonical hashes. | `AdmitCreateJSON`, `AdmitPatchJSON`, `AdmitConflictResolveJSON`, `AdmitSupersedeJSON`; four hash functions; request/claims types. | Workbook assembly adapters and tests. | `strictjson`, view schema, field normalization, Links, conflict tokens, canonical JSON/hash support. | Admission, direct-reference, Workbook contract, timestamp, and assessment cross-owner tests. | Workbook request/error envelopes and OpenAPI/protocol projections indirectly. | tasksdecisions | high | `RS-08A` removed `net/http` and `httpapi`; old `Decode*Request` declarations do not exist. |
| `internal/modules/tasksdecisions/mutation_admission_test.go` | Proves strict admission, normalization, hashes, and closed failure variants. | Test-only helpers. | Owner source-mutation row. | Admission functions. | Its own unit test. | Request envelopes/idempotency. | tasksdecisions tests | low | Covers plain, collection-only, patch-limit, and collection-limit details. |
| `internal/modules/tasksdecisions/mutation_capabilities.go` | Defines injected capabilities and typed stored mutation results. | Idempotency types/errors, capability interfaces, dependencies. | Workbook assembly and mutation facade. | PostgreSQL, Incidents, Records, Links, Revisions, Collaboration, conflict tokens. | Capability/composition/adapter/store tests. | Mutation/replay/revision/event behavior. | tasksdecisions | high | Application assembly supplies peers. |
| `internal/modules/tasksdecisions/mutation_capabilities_test.go` | Proves operation-kind mismatch fails before source mutation. | Test-only fake. | Owner test family. | Mutation facade/idempotency capability. | Its own unit test. | Replay failure behavior. | tasksdecisions tests | low | Verification evidence only. |
| `internal/modules/tasksdecisions/mutation_composition_test.go` | Proves complete injection and one shared mutation facade. | Test-only consumers/fakes. | Owner test family. | Mutation/import contributions. | Its own unit tests. | Composition boundary. | tasksdecisions tests | low | Prevents split coordinators. |
| `internal/modules/tasksdecisions/mutations.go` | Adapts public field aliases to policy/source operations and syncs direct Decision links. | Policy aliases and `TaskDecisionRecordFieldKey`. | Mutation/import and tests. | Source, policy, Links. | Reference, store, admission tests. | Decision-reference field/link contract. | tasksdecisions | medium | Retain semantic adapter. |
| `internal/modules/tasksdecisions/projectionprovider/contribution.go` | Constructs Task Request and Decision projection contribution. | `NewContribution`. | Projection assembly/tests. | Owner projection sources and `workbookprojection`. | Projection tests. | Projection provider index. | tasksdecisions | medium | Thin public constructor. |
| `internal/modules/tasksdecisions/publication.go` | Derives revision conflict facts and appends atomic `record_changed` intent. | Package-private facade helpers. | Create, patch, conflict, import, supersede. | Revisions and Collaboration capabilities. | Store/conflict/supersession/replay tests. | Revision/WebSocket semantics. | tasksdecisions | high | Side effects remain transaction-bound. |
| `internal/modules/tasksdecisions/recovery_state.go` | Declares authoritative recovery tables. | `NewRecoveryContribution`. | Recovery assembly/catalog. | Recovery declaration contract. | Recovery tests. | Recovery state catalog. | tasksdecisions | medium | No recovery algorithm. |
| `internal/modules/tasksdecisions/reporting_contribution.go` | Exposes owner Reporting fact provider. | Contribution type, constructor, provider/collection methods. | Server/reporting composition. | Reporting contract/provider and Task projection reader. | Reporting/server tests. | Reporting provider contract. | tasksdecisions | medium | Reporting owns rendering/lifecycle. |
| `internal/modules/tasksdecisions/revision_provider_contribution.go` | Contributes rollback/delete-restore providers and snapshot metadata. | `NewRevisionContribution`. | Revision assembly. | Owner providers and Revisions contract. | Revision/provider tests. | Snapshot schema IDs/view routes. | tasksdecisions | medium | Revisions validates full catalog. |
| `internal/modules/tasksdecisions/supersede_facade.go` | Coordinates atomic explicit Decision supersession. | View ID, request/command/result/fact/error types, `SupersedeDecision`. | Workbook action adapter/route. | Source/policy, Records, Links, Revisions, Projections, Collaboration, PostgreSQL. | Supersession, Workbook, composition, replay tests. | Supersede route, Decision view/event contracts. | tasksdecisions | high | Preserve exact effect and lock ordering. |
| `internal/modules/tasksdecisions/task_decisions_store_test.go` | Service-backed lifecycle, supersession, link, projection, replay, consistency, and atomic-effect evidence. | Test-only helpers/assertions. | Owner test family. | Mutation facade, Workbook catalog, PostgreSQL harness. | Its own tests. | Most owner observable contracts. | tasksdecisions tests | low | Primary behavior-freeze evidence. |
| `internal/modules/tasksdecisions/workbook_conflict.go` | Adapts conflict resolution to generic conflict-token coordination. | `ConflictCommand`, `ResolveConflict`. | Workbook conflict adapter/route. | Conflict tokens, idempotency, projections, revision snapshots. | Conflict/store/Workbook tests. | Conflict route/response. | tasksdecisions | high | Source target loading remains owner-local. |
| `internal/modules/tasksdecisions/workbook_facade.go` | Defines mutation contracts and coordinates create/patch, references, collections, conflicts, source writes, revisions, projections, replay, and publication. | View ID, `MutationFacade`, request/command/result/conflict types, constructor, `Create`, `Patch`. | Workbook assembly, import helpers, tests. | PostgreSQL, source/policy, Incidents, Records, Links, Projections, Revisions, Collaboration, conflict tokens. | Admission, composition, store, route, idempotency, browser tests. | Create/patch routes, views, conflicts, replay, OpenAPI/protocol. | tasksdecisions | high | Legitimate facade; same-package decomposition remains pending. |
| `internal/modules/tasksdecisions/workbookprojection/contribution.go` | Defines typed projection inputs, ports, descriptors, and surface intents for both views. | Input/page/fact types, interfaces, `Ports`, `Contribution`, constructor/descriptors/intents. | Projection assembly/adapters/query engine. | Projection provider contract and view registry. | Contribution/projection assembly tests. | Projection index, both canonical views/generated contracts. | tasksdecisions | high | Semantic input remains source-owned. |
| `internal/modules/tasksdecisions/workbookprojection/contribution_test.go` | Proves required sources, descriptors, intents, copies, and typed source retention. | Test-only fake. | Owner test family. | Projection contribution. | Its own unit tests. | Provider descriptor/intent contracts. | tasksdecisions tests | low | Verification evidence only. |

Authorized adjacent changed surfaces are exhaustively accounted for here:

| Path | Current responsibility | `RS-08A` change | Contract risk | Owner |
| --- | --- | --- | --- | --- |
| `internal/app/workbookassembly/action_adapters.go` | Workbook action composition. | Supersede admission now maps `AdmissionFailure` directly. | Supersede invalid-payload behavior. | Workbook application assembly |
| `internal/app/workbookassembly/taskdecision_adapters.go` | Create, patch, and conflict provider composition. | Uses four new admission functions and one exhaustive semantic-to-Workbook mapper. | All mutation invalid-payload variants. | Workbook application assembly |
| `internal/app/workbookassembly/taskdecision_admission_test.go` | Application translation evidence. | New dedicated four-variant adapter test. | Semantic identity preservation. | tasksdecisions/Workbook verification |
| `internal/modules/assessments/assessment_contract_test.go` | Cross-owner strict-object contract evidence. | Tasks/Decisions branches call the new admission surface. | Shared admission parity. | Assessments verification |
| `internal/modules/workbook/mutation_contract.go` | Closed Workbook mutation-failure vocabulary and HTTP translation. | Adds collection-context-only constructor without synthetic counts. | Core 01 error detail shape. | Workbook |
| `internal/modules/workbook/mutation_failure_test.go` | Workbook failure serialization evidence. | New exact plain/collection/count mapping matrix. | Status, code, message, details and omissions. | Workbook verification |
| `internal/modules/workbook/owner_admission_contract_test.go` | Cross-owner admission conformance. | Tasks/Decisions branches assert semantic failure accessors. | Strictness and reason identity. | Workbook verification |
| `internal/modules/workbook/timestamp_contract_test.go` | Timestamp admission conformance. | Tasks/Decisions branches call the new admission surface. | Timestamp normalization/nullability. | Workbook verification |
| `tools/backend_module_boundaries.json` | Authored backend boundary policy. | Adds the three-import Tasks/Decisions transport prohibition. | Boundary regression. | Backend boundary tooling |
| `tools/test_families/module.tasksdecisions.json` | Authored owner verification routing. | Adds one adapter-translation unit row. | Harness accounting. | Testing Harness |
| `tools/execution_topology_render_index.json` | Generated topology render index. | Refreshed by `make generate` from the authored row. | Generated drift/provenance. | Testing Harness generator |
| `docs/handoffs/tasksdecisions-module-refactor-tracker.md` | Execution tracker and handoff. | Revised to this normative, evidence-backed state. | Handoff completeness. | Documentation |

## 3. Module Boundary Diagnosis

The target is a legitimate source-owner facade, mutation coordinator, and set
of owner-local semantic adapters. It is mixed at the root-file cohesion level,
but it is not a cross-domain catch-all, frontend controller, grid-vendor layer,
or generic transport/persistence owner.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Task Request/Decision policy, lifecycle, and source facts | `internal/policy`, `internal/source` | tasksdecisions | keep | Core 02 and source-table boundaries | First-class Coordination objects. |
| Create, patch, conflict, and supersede transaction coordination | Root facade files | tasksdecisions | split | Workbook boundary and injected capabilities | Split only within the package; retain one transaction coordinator. |
| HTTP routes, authorization, and final envelopes | Workbook and app adapters | Workbook/application | keep | Core 01/04 and Workbook boundary | MUST NOT move into tasksdecisions. |
| Semantic JSON admission and canonical hashes | `mutation_admission.go`, `admission_failure.go` | tasksdecisions | split | Source-owner admission boundary | Transport-neutral seam is complete; operation file decomposition remains pending. |
| Admission-to-public-failure translation | `internal/app/workbookassembly` | Workbook application assembly | keep | `RS-08A` adapter and Workbook failure vocabulary | Exact mapping is defined in Section 4. |
| Projection source reads/intent | Owner providers and `workbookprojection` | tasksdecisions | keep | Projections boundary | Physical projection storage remains in Projections. |
| Revision rollback/delete-restore adapters | Owner providers/contribution | tasksdecisions | keep | Revisions boundary | Revisions coordinates and validates the catalog. |
| Import, Reporting, Incident Bundle, and Recovery contributions | Owner adapters | tasksdecisions | keep | Adopted subsystem owners and live assemblies | Coordinating subsystems retain cross-owner orchestration. |
| `record_changed` intent creation | `publication.go` | tasksdecisions producer; Collaboration transport | keep | Atomic publication capability | Socket authorization/replay remains Collaboration-owned. |
| Frontend workflow, saved views, view registry, selectors, grid adapter | Outside target | Existing frontend/Workbook/Views owners | keep | Live imports and controller code | No frontend source change is required. |
| Timeline, Evidence, Links, Entities/Indicators domains | Named modules | Their named owners | keep | Narrow injected capabilities and boundary manifest | References do not transfer ownership. |

## 4. Public Contract and Behavior Freeze Map

### Admission interface

`RS-08A` is intentionally breaking only for the private-to-repository Go API.
No compatibility aliases, wrappers, or dual error seam MAY exist.

| Requirement ID | Operation | Required signature | Success result | Semantic failure result |
| --- | --- | --- | --- | --- |
| AD-REQ-001 | Create | `AdmitCreateJSON(viewSchemaID string, reader io.Reader) (CreateRequest, *AdmissionFailure)` | Existing normalized `CreateRequest`; nil failure. | Zero request plus non-nil closed failure. |
| AD-REQ-002 | Patch | `AdmitPatchJSON(reader io.Reader) (PatchRequest, *AdmissionFailure)` | Existing normalized `PatchRequest`; nil failure. | Zero request plus non-nil closed failure. |
| AD-REQ-003 | Conflict resolution | `AdmitConflictResolveJSON(reader io.Reader, token string, claims ConflictClaims) (ConflictResolveRequest, *AdmissionFailure)` | Existing normalized request; nil failure. | Zero request plus non-nil closed failure. |
| AD-REQ-004 | Decision supersede | `AdmitSupersedeJSON(reader io.Reader) (SupersedeRequest, *AdmissionFailure)` | Existing normalized request; nil failure. | Zero request plus non-nil closed failure. |
| AD-REQ-005 | Retired API | `DecodeCreateRequest`, `DecodePatchRequest`, `DecodeConflictResolveRequest`, and `DecodeSupersedeRequest` MUST be absent. | Not applicable. | A remaining declaration or call site fails acceptance. |

### Closed failure interface and defaults

| Requirement ID | Member | Required behavior | Absent/default behavior |
| --- | --- | --- | --- |
| AF-REQ-001 | Private state/constructors | Only Tasks/Decisions admission code MAY populate semantic details. | The type exposes no mutable detail field. |
| AF-REQ-002 | `Error() string` | MUST return exactly `tasksdecisions: invalid mutation admission`. | Same safe text for every variant; raw input MUST NOT be disclosed. |
| AF-REQ-003 | `Field() string` | Returns the public validation field identity. | Returns `""` for a nil receiver. |
| AF-REQ-004 | `ReasonCode() string` | Returns the closed public reason code. | Returns `""` for a nil receiver. |
| AF-REQ-005 | `CollectionFieldKey() (string, bool)` | Returns collection context only when present. | Returns `"", false`; it MUST NOT synthesize context. |
| AF-REQ-006 | `CountLimit() (requestedCount int, maxCount int, ok bool)` | Returns the raw requested count and configured maximum as one paired variant. | Returns `0, 0, false`; one count MUST NOT be present without the other. |

### Adapter mapping

Every semantic failure variant MUST map to exactly one Workbook failure. The
Workbook route layer remains the only producer of the public HTTP error.

| Admission variant | Presence | Workbook constructor | Public details | Required omissions |
| --- | --- | --- | --- | --- |
| Plain | `field`, `reason_code` | `InvalidPayloadFailure` | `field`, `reason_code` | `field_key`, `requested_count`, `max_count` |
| Collection context only | `field`, `reason_code`, `field_key` | `InvalidPayloadCollectionFailure` | `field`, `reason_code`, `field_key` | `requested_count`, `max_count` |
| Patch count limit | `field`, `reason_code`, count pair | `InvalidPayloadLimitFailure` | `field`, `reason_code`, `requested_count`, `max_count` | `field_key` |
| Collection count limit | `field`, `reason_code`, `field_key`, count pair | `InvalidPayloadLimitFailure` | All five members | None |

All four rows produce HTTP `400`, code `invalid_mutation_payload`, message
`invalid mutation payload`, and the common Core 01 error envelope. In
particular, `empty_collection_actions` MUST include `field_key` and MUST omit
both count members.

### Frozen observable contracts

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Workbook create, patch, conflict-resolution, and supersede routes | Workbook routes; tasksdecisions providers | Core 01, OpenAPI owner, route catalog | Workbook/source/browser tests | Exact admission mapping, unchanged route response and rejection behavior | high | Paths, methods, status, envelopes, precedence MUST remain exact. |
| Collaboration `GET /ws/v1/incidents/{incident_id}` and `record_changed` | Collaboration | Collaboration implementation/protocol | Collaboration plus mutation-effect tests | Preserve committed intent count, order, row version, authorization, replay | high | Target emits intent only. |
| Task Request defaults, lifecycle, references, links | tasksdecisions | Core 02, policy/source/views | Admission/store/reference tests | Preserve defaults, transitions, timestamps, link sync, incident/type validation | high | No generalized workflow. |
| Decision lifecycle and supersession | tasksdecisions | Core 02 and supersede facade | Store/terminal/supersession tests | Preserve explicit-only supersession, consistency, replay, effects | high | No generalized approval gate. |
| Strict object admission | tasksdecisions semantic owner | Core 01 and `strictjson` | Admission/owner contract tests | Reject non-object, duplicate members at any depth, trailing data, unknown members, invalid nullability | high | Validation precedence remains unchanged. |
| Patch and collection bounds | Core 01; tasksdecisions admission | Core 01 AC-299 | Admission and Workbook mapping tests | Raw 33/32 and 65/64 counts; no truncation; empty collection has no counts | high | Count validation precedes replay/write. |
| Normalization and canonical replay hashes | tasksdecisions | Admission/hash code and idempotency contract | Admission/idempotency/store tests | Preserve stable UUID admission, normalization, canonical bytes, operation-kind replay | high | Internal API rename MUST NOT alter hashes. |
| Revision/change-set and projection refresh | Revisions/Projections coordinators; owner facts/input | Adopted boundaries | Store/revision/projection tests | Preserve ordering, fields, refresh timing, atomicity | high | No transaction change in `RS-08A`. |
| Canonical views and import targets | Views/Workbook and Imports; source semantics here | Both view schemas and import targets | Contract/import/frontend/browser tests | Preserve stable IDs, fields, enums, writeback classes, target behavior | high | Saved views remain outside target. |
| Projection-provider descriptors | Projections plus owner contribution | Provider index | Contribution/assembly tests | Preserve both descriptors, source paging, rebuild intents | medium | Generated index is downstream. |
| Incident Bundle family and Recovery catalog | Coordinators plus owner adapters | Source catalog/recovery catalog | Portability/recovery tests | Preserve paths, exact rows, invariants, atomicity, table declarations | high | No generic coordinator ownership moves. |
| Generated protocol/OpenAPI/view projections | Authored contract owners/generators | Generated artifact policy | Drift/OpenAPI/frontend tests | No hand edits; no public generated delta from `RS-08A` | high | Only execution-topology render index changed through generation. |
| Authorization and zero rejected effects | Core 04 and Workbook/application | Route guards and transaction tests | Route/security/store tests | Preserve auth-first precedence and prove rejected admission has no durable or event effects | high | Semantic admission MUST NOT authorize. |
| Frontend controller/selectors/grid adapter | Frontend Workbook shell/grid adapter | Live frontend code | Vitest/browser owner rows | Browser regression only; frontend boundary check only if frontend is touched | medium | No frontend file changed. |
| Harness accounting | Testing Harness | Owner manifest/catalog/topology | Twelve owner rows after `RS-08A` | Dedicated adapter translation row plus drift/policy checks | medium | Verification accounting is not runtime architecture. |
| Entities/Indicators, Evidence, Timeline behavior | Named owners | No target ownership found | Their owner tests | None beyond shared broad regression | low | Outside this refactor. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Source SQL is confined to owner-private packages. | Source/projection/rollback/portability packages and table allowlists | high if moved | `intentional/no_action` | tasksdecisions | Preserve private ownership. |
| Projection, revision, reporting, import, portability, recovery, and publication semantics are source contributions to coordinating subsystems. | Adopted boundaries and live assemblies | low as implemented | `intentional/no_action` | split semantic/coordinator ownership | Do not relocate providers to reduce directory size. |
| Generic peers are injected by application assembly. | `MutationDependencies` and assemblies | low | `intentional/no_action` | application assemblies | Preserve narrow ports and one facade. |
| `workbook_facade.go` remains a large mixed coordinator file. | Exact declarations and call flow | medium | `should_fix` | tasksdecisions | Decompose within package after characterization; preserve exported surface/effect order. |
| `mutation_admission.go` still combines four operations, canonicalization, and hashes. | Exact declarations | medium | `should_fix` | tasksdecisions | Split by operation in later `RS-03`; preserve the new interface and hashes. |
| Tasks/Decisions admission previously imported HTTP transport types. | Pre-slice diff; new `AdmissionFailure`; adapter mapper; boundary rule | high if reintroduced | `must_fix` | tasksdecisions semantics; Workbook translation | Resolved by `RS-08A`; retain the import prohibition and no dual seam. |
| Field knowledge repeats across policy, admission, source updates, projection reads, and view contracts. | Switches, allowlists, SQL scans, view schemas | medium | `should_fix` | tasksdecisions plus view owners | Add parity evidence before consolidation; never read Markdown/generated JSON at runtime. |
| Route authorization remains outside owner semantics. | Workbook guards, app adapters, capability injection | high | `intentional/no_action` | Core 04/Workbook | Preserve precedence and fail-closed attribution. |
| Generated artifacts remain downstream projections. | Generated policy and `make generate` diff | high | `intentional/no_action` | generators/authored inputs | Never hand-edit. |
| No frontend, grid-vendor, saved-view, Timeline, Evidence, Entities/Indicators, or test-only production coupling exists. | Target import scan and exact source inspection | low | `intentional/no_action` | existing owners | Preserve separation. |
| Cross-module relocation or any public behavior change lacks present authorization. | `RS-08B` scope | high | `defer` | affected adopted owners | Require later owner evidence and explicit authorization. |

No unresolved current-state owner violation remains after `RS-08A`.

## 6. Refactor Workstreams

The dependency graph is
`WF-00 -> {WF-01, WF-02} -> {WF-03, WF-04} -> WF-05 -> WF-06 -> WF-07 -> WF-08`.

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap | root | none | WF-01, WF-02 | Bind authority, revision, baseline, scope, and history. | Tracker, owner docs, Git state | Revision/digest and baseline roots | Authorization and release states are explicit. |
| WF-01 | Target inventory | parallel | WF-00 | WF-03, WF-04 | Account for every target file and surface. | All 38 target files/callers | Exact source inspection | Section 2 is exhaustive. |
| WF-02 | Contract-owner mapping | parallel | WF-00 | WF-03, WF-04 | Assign routes, events, schemas, providers, frontend, and harness contracts. | Owners, contracts, assemblies | Exact owner/live inspection | Semantic and coordinator ownership are distinct. |
| WF-03 | Characterization gap analysis | parallel | WF-01, WF-02 | WF-05 | Prove behavior affected by the next slice without asserting file layout. | Owner, Workbook, app, browser tests | Owner/service slices | Required gaps have tests or documented no-op findings. |
| WF-04 | Boundary/coupling scan | parallel | WF-01, WF-02 | WF-05 | Classify coupling and enforce transport neutrality. | Imports and boundary manifest | `make backend-module-boundary-check` | Findings have exact action/classification. |
| WF-05 | Facade/seam ownership design | chain | WF-03, WF-04 | WF-06 | Retain source ownership, define transport mapping, and plan same-package cohesion. | Admission/facade/assembly/Workbook contracts | Interface/mapping review | No owner/interface choice is implicit. |
| WF-06 | Slice sequencing | chain | WF-05 | WF-07 | Order independently reversible, behavior-preserving changes. | Root package/tests | Per-slice gates | Each slice has stop, rollback, and completion criteria. |
| WF-07 | Harness/accounting | chain | WF-06 | WF-08 | Route actual tests and refresh downstream topology. | Owner manifest/generated topology | Generation/drift/policy | Rows account for evidence only. |
| WF-08 | Validation/final handoff | chain | WF-07 | none | Complete the applicable ladder and leave reproducible evidence. | Changed files/tracker | Section 8 | Results, failures, skips, risks, next action are current. |

## 7. Proposed Refactor Slice Plan

Every slice MUST be independently revertible. A behavior-changing discovery
outside `RS-08A` MUST stop the slice and move to `RS-08B`.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| RS-00 | none | Establish revision-bound owner, service-backed, and boundary baseline. | No production files | False baseline or environmental failure | All then-current owner rows | Owner slice; service-backed slice; backend boundary | No code rollback; retain failures. | All three baseline roots are recorded and passing, or work stops. |
| RS-01 | RS-00 | **SUPERSEDED by `RM-03`.** Historical characterization proposal only. | Existing owner tests/accounting if topology changes | Asserting implementation details | Black-box owner behavior | Owner slices | Revert unsupported tests/accounting only. | Governed by Section 13. |
| RS-02 | RS-01 | **SUPERSEDED by `RM-06`.** Historical facade-decomposition proposal only. | Root tasksdecisions files | Transaction/lock/effect order, replay, conflicts, references, projections | Owner/service/Workbook/browser evidence | Owner slices; boundary; browser when route-adjacent | Revert file split as one unit. | Governed by Section 13. |
| RS-03 | RS-02 | **SUPERSEDED by `RM-07`.** Historical admission-decomposition proposal only. | Admission files/tests | Strictness, limits, precedence, normalized hashes | Admission/owner/Workbook contract tests | Owner slice; `make test-fast` | Revert decomposition only; no durable migration. | Governed by Section 13. |
| RS-04 | RS-03 | **SUPERSEDED by `RM-09`.** Historical conditional supersession review only. | Supersede/publication/capabilities | Revisions, links, projections, events, replay | Supersession/store/Collaboration tests | Service slice; boundary | Revert helper extraction independently. | Governed by Section 13. |
| RS-05 | RS-04 | **SUPERSEDED by `RM-04..RM-05`.** Historical mapping-cleanup proposal only. | Policy/admission/source/projection/tests | Field/default drift | Owner semantic parity | Owner slice; drift; JSON shape | Revert cleanup if ownership becomes less explicit. | Governed by Section 13. |
| RS-06 | RS-05 | **SUPERSEDED by `RM-10`.** Historical harness-accounting proposal only. | Authored test manifest and generated topology | False accounting/hand edits | Harness contract | Generate; drift; policy; JSON shape | Revert authored input/generated refresh together. | Governed by Section 13. |
| RS-07 | RS-06 | **SUPERSEDED by `RM-11`.** Historical final-validation proposal only. | Tracker/evidence only | Missed cross-surface regression | All affected owner/consumer evidence | Boundary; browser; `test-fast`; `agent-finalize`; risk-based `check` | Revert first failing slice, not tests/contracts. | Governed by Section 13. |
| RS-08A | RS-00 and explicit 2026-08-25 seam authorization | Replace only the Tasks/Decisions HTTP error seam with `AdmissionFailure`, four breaking `Admit*JSON` functions, direct Workbook mapping, transport-import rule, and dedicated test row. | Tasks/Decisions admission; Workbook assembly/failure contract; tests; boundary/harness inputs | Public invalid-payload shape, auth precedence, hashes, rejected effects | Four admissions; four failure variants; Core 01 limits/shape; owner/service/browser tests | Full Section 8 seam ladder | Revert source, adapter, tests, authored manifest, and generated index as one slice; no schema/data rollback. | `AD-REQ-*`, `AF-REQ-*`, and Section 12 seam criteria pass. |
| RS-08B | Separate later authorization | **SUPERSEDED by the explicitly authorized and scoped `RM-00..RM-11` ledger.** It remains historical evidence that unplanned observable behavior changes require separate owner review. | Affected owners/contracts/implementation | Intentional observable change | Owner-specified new evidence | Section 13 | Do not broaden beyond the scoped remediation ledger. | Governed by Section 13. |

## 8. Validation Plan

### Fail-closed outcomes and slice gates

| Outcome | Required action |
| --- | --- |
| Selected affected assertion fails | Stop. Diagnose before broadening or changing another slice. |
| Harness/configuration/infrastructure fails | Repair and rerun. Record it as infrastructure, never as a product pass. |
| Clearly unrelated pre-existing product failure | Record exact row, class, root, relatedness analysis, and owner acceptance; never claim it passed. |
| Owner/catalog drift | Repair authored inputs and regenerate before using owner evidence. |
| Boundary failure | Stop; MUST NOT widen an allowlist as refactor convenience. |
| Generated/OpenAPI drift from an unintended owner input | Stop and revert the active slice. |
| Owner contradiction | Record `BLOCKED: owner contradiction`; do not choose. |

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make test-slice OWNER=module.tasksdecisions` | Twelve routed owner rows after `RS-08A` | yes | Baseline and post-slice roots are recorded below. |
| integration | `make service-backed-test-slice OWNER=module.tasksdecisions` | Service-backed source/import/portability/adapter evidence | yes | Baseline and post-slice roots are recorded below. |
| e2e/browser | `make browser-e2e-webserver-backed` | Workbook Task/Decision browser flows | no | Required and run because the seam is response-adjacent. |
| generated drift | `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check` | Authored harness refresh, generated projections/policy, JSON shape | no | `make generate` changed only the topology render index downstream of the new row. |
| import-boundary/static | `make backend-module-boundary-check` | Module imports and new Tasks/Decisions transport prohibition | yes | Frontend import check is not applicable because no frontend file/import changed. |
| OpenAPI compatibility | `make openapi-compatibility-check` | Public OpenAPI compatibility | no | Required for the response-adjacent seam. |
| documentation | `make lint-markdown` | Tracker Markdown | no | Run after the final tracker update. |
| focused broader | `make test-fast` | Fast repository regression surface | no | Run after focused gates. |
| full check | `make agent-finalize`; `make check` | Final handoff maintenance and standard full check | no | `agent-finalize` precedes the risk-based full check. |

### Retained evidence

| Phase | Command | Result | Retained root or artifact |
| --- | --- | --- | --- |
| Pre-slice baseline | `make backend-module-boundary-check` | pass | `.cartulary/test-results/20260825T215807Z-p270417` |
| Pre-slice baseline | `make test-slice OWNER=module.tasksdecisions` | pass | `.cartulary/test-results/20260825T215812Z-p270805` |
| Pre-slice baseline | `make service-backed-test-slice OWNER=module.tasksdecisions` | pass | `.cartulary/test-results/20260825T215903Z-p310638` |
| Formatting | `make format` | pass | `.cartulary/test-results/20260825T221023Z-p354137` |
| Harness refresh | `make generate` | pass | `.cartulary/test-results/20260825T221038Z-p358028` |
| Post-slice owner | `make test-slice OWNER=module.tasksdecisions` | pass, 19/19 units | `.cartulary/test-results/20260825T221058Z-p360930` |
| Post-slice integration | `make service-backed-test-slice OWNER=module.tasksdecisions` | pass, 15/15 units | `.cartulary/test-results/20260825T221152Z-p402364` |
| Post-slice boundary | `make backend-module-boundary-check` | pass, 3/3 units | `.cartulary/test-results/20260825T221244Z-p442377` |
| Post-slice generated drift | `make generate-drift` | pass, 4/4 units | `.cartulary/test-results/20260825T221251Z-p442791` |
| Post-slice generated policy | `make generated-artifact-policy-check` | pass, 3/3 units | `.cartulary/test-results/20260825T221251Z-p442824` |
| Post-slice JSON shape | `make json-shape-check` | pass, 3/3 units | `.cartulary/test-results/20260825T221251Z-p442818` |
| Post-slice OpenAPI | `make openapi-compatibility-check` | pass, 4/4 units | `.cartulary/test-results/20260825T221304Z-p446556` |
| Post-slice browser | `make browser-e2e-webserver-backed` | pass, 58/58 units | `.cartulary/test-results/20260825T221304Z-p446712` |
| Post-slice broader | `make test-fast` | pass, 431/431 units | `.cartulary/test-results/20260825T221713Z-p502851` |
| Documentation | `make lint-markdown` | pass before and after final failure disposition and changed-file accounting | `.cartulary/test-results/20260825T222850Z-p513701`; `.cartulary/test-results/20260825T223621Z-p638899`; `.cartulary/test-results/20260825T223755Z-p640588` |
| Finalization | `make agent-finalize` | pass, 1/1 unit | `.cartulary/test-results/20260825T222943Z-p514796` |
| Full check | `make check` | fail, 656/657 units; unrelated existing-dependency vulnerability `GO-2026-6253` in `github.com/moby/go-archive v0.2.0`; `go-vulncheck` alone failed and no dependency file changed | `.cartulary/test-results/20260825T223002Z-p517705`; findings at `unit-artifacts/target-go-vulncheck/govulncheck-findings.json` |

No product validation result is inferred from a command that was not run.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| TD-001 | Establish scope, authority, revision binding, and state model | WF-00 | DONE | none | Section 1 | Authority and distinct completion states are explicit. |
| TD-002 | Inventory all 38 target files and direct surfaces | WF-01 | DONE | TD-001 | Section 2 | Every target file has a concrete row. |
| TD-003 | Map behavior and supporting-subsystem ownership | WF-02 | DONE | TD-001 | Sections 3-4 | Every public risk has an owner/test posture. |
| TD-004 | Classify coupling and boundary findings | WF-04 | DONE | TD-002, TD-003 | Section 5 | Every finding uses the controlled classification. |
| TD-005 | Establish owner/service/boundary baseline | WF-03 | DONE | TD-002, TD-003 | Three pre-slice roots in Section 8 | Production release gate passed before code movement. |
| TD-006 | Characterize and implement transport-neutral admission failures | WF-03, WF-05 | DONE | TD-005 | Admission/failure/adapter tests | Four variants and exact mappings are proved. |
| TD-007 | Decompose create/patch/conflict facade within package | WF-05, WF-06 | SUPERSEDED | TD-006 | `RM-06` | Exported surface and behavior remain exact. |
| TD-008 | Decompose admission files around the new seam | WF-05, WF-06 | SUPERSEDED | TD-007 | `RM-07` | Admission outcomes and hashes remain exact. |
| TD-009 | Review supersession/publication cohesion | WF-05, WF-06 | SUPERSEDED | TD-008 | `RM-09` | Atomic transaction ownership remains explicit. |
| TD-010 | Add adapter test row and refresh generated topology | WF-07 | DONE | TD-006 | Authored manifest; generated render index; drift results | Twelve owner rows route real tests; no hand edits. |
| TD-011 | Enforce Tasks/Decisions transport neutrality | WF-04 | DONE | TD-006 | Boundary manifest and passing root | Forbidden imports fail the boundary check. |
| TD-012 | Complete `RS-08A` final validation and handoff | WF-08 | DONE | TD-010, TD-011 | Section 8 and current handoff | Markdown and finalization pass; the unrelated full-check failure is precisely dispositioned without a false pass claim. |
| TD-013 | Make any remaining behavior or ownership change | WF-05 | SUPERSEDED | Separate authorization | Scoped `RM-*` work; no public behavior change authorized | Later owner-complete authorization still controls any unplanned behavior change. |
| TD-014 | Complete broader same-package refactor | WF-08 | SUPERSEDED | TD-007, TD-008, TD-009 | `RM-00..RM-11` | Section 13 is the sole active completion ledger. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-25T17:20:30-04:00 | Codex / tracker-plan-2026-08-25 | Planning inventory complete; implementation not started | Inspected framework, domain, Core 00-04, adopted boundary decisions, Reporting and Testing Harness NLSpecs; touched only this tracker | `sed`, `rg`, `git status`, `stat`, `date` | Authority and clean write scope confirmed; no owner contradiction found | RB-001 | Obtain explicit implementation authorization, then start RS-00. |
| 2026-08-25T18:25:00-04:00 | Codex / admission-seam-2026-08-25 | `RB-001` administratively closed; `RS-08A` implemented; broader refactor incomplete | Inspected request-bound revision/digest, analysis notes, NLSpec guidance, owner/live sources; touched authorized seam, tests, harness inputs/generated index, and tracker | `git status`, `sha256sum`, `sed`, `rg`, Make targets in Section 8 | Scope remained seam-specific; no owner contradiction or unauthorized behavior change found | none for `RS-08A` | Finish final tracker validation, then leave `RS-02` as the next implementation slice. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-25T17:20:30-04:00 | Codex / tracker-plan-2026-08-25 | Legitimate source-owner boundary; same-package cohesion work proposed | Inspected all 37 target files, direct app assemblies, Workbook/Collaboration routes, Projections consumers, and backend boundary manifest; touched only this tracker | `rg --files`, declaration/import searches, exact `sed` reads | No `must_fix` boundary violation; large facade/admission files classified `should_fix` | RB-001 | Establish baseline and implement RS-02 only in a later task. |
| 2026-08-25T18:25:00-04:00 | Codex / admission-seam-2026-08-25 | Source owner exposes only transport-neutral admission failure; Workbook assembly maps transport behavior | Added `admission_failure.go`; changed admission and Workbook adapters; added transport-import rule | `rg`; `make backend-module-boundary-check` | Old decoders/call sites and target HTTP imports are absent; boundary passed at the Section 8 root | none | Preserve rule; later split files without changing the new surface. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-25T17:20:30-04:00 | Codex / tracker-plan-2026-08-25 | Frontend remains a consumer of generated Workbook contracts; no frontend move planned | Inspected Coordination workflow controller, mutation command ports, surface policies, and registry; touched only this tracker | Targeted `rg` and `sed` | No target-owned frontend shell, saved-view controller, selector, or grid-vendor code found | none beyond RB-001 | Run frontend checks only if a later slice changes frontend or public shapes. |
| 2026-08-25T18:25:00-04:00 | Codex / admission-seam-2026-08-25 | No frontend source/import/public generated shape changed | No frontend file touched; browser consumer surface inspected through existing owner route | `make browser-e2e-webserver-backed`; generated/OpenAPI checks | Browser passed; frontend import check skipped as not applicable | none | Run frontend import/type checks only if a future slice touches frontend or public generated shapes. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-25T17:20:30-04:00 | Codex / tracker-plan-2026-08-25 | Contract freeze map complete; no owner input or generated output changed | Inspected both view schemas, projection provider index, import targets, Incident Bundle source catalog, recovery catalog, Workbook OpenAPI owner, and generated artifact policy; touched only this tracker | `jq`, `rg`, `sed` | Authored owners and downstream generated surfaces identified; hand-edit prohibition recorded | RB-001 | Use drift targets after later contract-sensitive work; do not edit generated files. |
| 2026-08-25T18:25:00-04:00 | Codex / admission-seam-2026-08-25 | Public contract frozen; collection-context mapping now conforms exactly to Core 01 | Changed Workbook typed failure constructor/mapping tests; generated only `tools/execution_topology_render_index.json` through Make | `make generate`; drift/policy/JSON-shape/OpenAPI targets | Generation and every contract/drift gate passed at Section 8 roots; no protocol/OpenAPI/view generated delta | none | Preserve exact omission of count members for `empty_collection_actions`. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-25T17:20:30-04:00 | Codex / tracker-plan-2026-08-25 | Eleven owner rows and validation ladder discovered; no product suite run | Inspected target tests, owner test-family manifest, verification owner, test catalog, and public Make target surface; touched only this tracker | `make task-guide ROLE=module-author OWNER=module.tasksdecisions`; `make explain-test-owner OWNER=module.tasksdecisions`; searched `make help-all` output | Discovery commands completed; six rows are service-backed; no validation success claimed | RB-001 and fresh baseline not yet run | Run RS-00 in the authorized implementation session. |
| 2026-08-25T18:25:00-04:00 | Codex / admission-seam-2026-08-25 | Twelve owner rows include dedicated application-adapter translation | Updated admission/cross-owner tests; added Workbook mapping and app adapter tests; added authored owner row and generated index | `make explain-test-owner OWNER=module.tasksdecisions`; owner/service slices; browser; `test-fast`; generation/drift targets; full check | Explanation reports 12 rows and six service-backed; every affected gate passes; full check alone fails unrelated `go-vulncheck` | RB-002 is outside seam scope | Retain the row for future seam regression; remediate the repository dependency separately. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-25T17:20:30-04:00 | Codex / tracker-plan-2026-08-25 | Route authorization and source validity remain distinct | Inspected Core 04, Workbook route guards, app adapters, mutation capabilities, member and record validators; touched only this tracker | Targeted `rg` and `sed` | No authorization logic was found in an incorrect target layer; actor and incident checks are injected or source-semantic | RB-001 | Preserve failure precedence and attribution in characterization and later slices. |
| 2026-08-25T18:25:00-04:00 | Codex / admission-seam-2026-08-25 | Semantic admission cannot construct transport/auth failures; route authorization remains outside owner | Inspected guards/effect tests; changed only semantic failure and application translation surfaces | Owner/service/browser/boundary tests; `make check` diagnostics | Admission security evidence passes; full check found unrelated existing `GO-2026-6253` test-infrastructure dependency exposure | RB-002 outside seam scope | Retain transport rule; remediate the dependency in a separately authorized maintenance task. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-25T17:20:30-04:00 | Codex / tracker-plan-2026-08-25 | Tracker prepared for implementation handoff | Inspected current worktree and tracker scope; touched only this tracker | `git status --short --branch`; file inventory and timestamp commands | Highest risks are transaction/effect ordering, hash stability, field-map drift, and generated hand-edit risk | RB-001 | Later authorized agent begins with RS-00 and updates all relevant handoff tables with actual results. |
| 2026-08-25T18:25:00-04:00 | Codex / admission-seam-2026-08-25 | `RS-08A` is complete with one rollback unit; broader file decomposition remains pending | Inspected current diff, generated outputs, and full-check finding; touched files listed by Git plus this tracker | `git diff --stat`; `git status`; Section 8 validation; `make explain-run`; `jq` finding inspection | No schema, dependency, migration, frontend, durable compatibility layer, or product behavior change; `GO-2026-6253` is present in the bound base dependency graph | RB-002 outside current scope | Next authorized work is `RS-02`; dependency remediation is separate; `RS-08B` remains deferred. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Production implementation was outside the original planning-only task. | Administrative closure must not be mistaken for baseline release, slice completion, or full refactor completion. | The 2026-08-25 requests bound `RS-08A` and Iteration 2 to exact repository/tracker revisions; baseline, serial checkpoint, and final evidence are retained below. | CLOSED - both authorized iterations are complete; further observable behavior change remains out of scope. |
| RB-002 | Repository-wide `make check` historically failed `go-vulncheck` for `GO-2026-6253` in indirect `github.com/moby/go-archive v0.2.0`. | The finding was security-relevant and prevented an all-green full repository gate. | `RM-02` used Go module tooling to resolve `go-archive v0.3.3` while retaining Testcontainers `v0.42.0`; `RM-11` retained passing vulnerability and full-check evidence. | CLOSED - the resolved graph is `v0.3.3`, `go-vulncheck` passes, and final `make check` passes 658/658. |

There is no open blocker for historical `RS-08A` or Iteration 2. The scoped
Section 13 ledger supersedes `RS-08B`, but any unplanned observable behavior
change still requires owner review.

## 12. Binary Completion Criteria

| Acceptance ID | Traceable requirement | Binary criterion | Evidence | Status |
| --- | --- | --- | --- | --- |
| TD-AC-001 | TD-REQ-001, TD-REQ-002 | Authority order, revision binding, state distinctions, and owner-contradiction stop rule are explicit. | Sections 1, 8, 11 | PASS |
| TD-AC-002 | TD-REQ-003 | All 38 target files and every discovered public contract risk have an owner and test posture. | Sections 2-4 | PASS |
| TD-AC-003 | AD-REQ-001 through AD-REQ-005 | Exactly four `Admit*JSON` functions exist and every old Tasks/Decisions decoder declaration/call site is absent. | Source inspection; owner/test-fast roots | PASS |
| TD-AC-004 | AF-REQ-001 through AF-REQ-006 | `AdmissionFailure` has private state, fixed safe text, exact accessors, and paired optional counts with fail-closed defaults. | `admission_failure.go`; admission tests | PASS |
| TD-AC-005 | TD-REQ-003, adapter table | Plain, collection-only, patch-limit, and collection-limit variants map exactly; empty collection omits both count members. | Adapter and Workbook mapping tests; owner/test-fast roots | PASS |
| TD-AC-006 | TD-REQ-003 | Strict-object rejection, unknown members, nullability, stable UUID admission, normalization, and canonical hashes remain characterized. | Admission/cross-owner tests; owner root | PASS |
| TD-AC-007 | TD-REQ-003 | Raw patch 33/32 and collection 65/64 counts are reported without truncation; empty actions fail with collection context only. | Admission and Workbook failure tests | PASS |
| TD-AC-008 | TD-REQ-004, TD-REQ-005 | Authorization precedence and zero rejected effects remain passing; transaction, revision, projection, replay, and event behavior are unchanged. | Service-backed and browser roots | PASS |
| TD-AC-009 | TD-REQ-007 | Tasks/Decisions production code imports none of the three prohibited HTTP/auth packages. | Boundary rule and passing boundary root | PASS |
| TD-AC-010 | TD-REQ-006, TD-REQ-008 | Dedicated adapter row exists, twelve owner rows are accounted for, and downstream topology was generated through Make with clean drift/policy/shape. | Authored manifest, generated index, generation/drift roots | PASS |
| TD-AC-011 | TD-REQ-009, TD-REQ-010 | Every required workflow/slice has dependencies, stop conditions, rollback, validation, and exact completion evidence. | Sections 6-10 | PASS |
| TD-AC-012 | TD-REQ-003 | OpenAPI and browser-visible Workbook behavior remain compatible. | OpenAPI and browser roots | PASS |
| TD-AC-013 | TD-REQ-010 | Markdown and finalization pass; the risk-based full check passes or has a precise unrelated-failure disposition without a false success claim. | Section 8 final rows and RB-002 | PASS |
| TD-AC-014 | Scope posture | No migration, dependency, frontend source, public contract input, or persistent compatibility layer was introduced. | Git diff and handoff | PASS |

`RS-08A` is complete because every acceptance row is `PASS`; `make check` was
not claimed as passing and `RB-002` remained explicit at that historical
checkpoint. Section 13 now governs the separately authorized broader
remediation; the superseded `RS-*` rows are not active work.

## 13. Iteration 2 Remediation

### Authority and checkpoint protocol

The 2026-08-25 remediation request authorizes only the serial work in this
section against repository commit
`dd06b39a1ec901b0c610c87606546324552693da` and pre-iteration tracker SHA-256
`889a4b1d812d6d8df7c5bff7b1166ea0d2c493bd44d7516a357cb75391c297d6`.
Completed `RS-08A` evidence remains historical. The older pending `RS-*` and
`TD-*` proposals marked `SUPERSEDED` above are not completion claims.

The rows below execute in order. After each row, the session MUST record the
exact diff, commands, retained roots, failures, skips, rollback unit, and exit
evidence here and MUST pass `make lint-markdown` before activating the next
row. An owner contradiction, unintended public-contract drift, affected test
failure, boundary failure, generated drift, or unresolved security failure
sets the active row to `BLOCKED` and stops later work.

### Remediation ledger

| ID | Workstream | Status | Depends on | Rollback unit | Exit evidence |
| --- | --- | --- | --- | --- | --- |
| RM-00 | Tracker rebaseline | DONE | none | Section 13 authorization, ledger, and baseline evidence | Owner 19/19, service-backed 15/15, and boundary 3/3 pass; vulnerability baseline reproduces only blocking `GO-2026-6253`. |
| RM-01 | Specification cleanup | DONE | RM-00 | Core owner, conformance, trace, and tracker edits | `REQ-01-671` and `AC-565` are bidirectionally traced; the orphaned Decision field is absent from owner requirements. |
| RM-02 | Archive dependency remediation | DONE | RM-01 | Module-managed `go.mod` and `go.sum` change | Resolved graph uses `go-archive v0.3.3`; vulnerability, service, toolchain, and fast-suite gates pass. |
| RM-03 | Characterization freeze | DONE | RM-02 | New or strengthened owner tests | Four exact hashes and source-before-target supersession intent ordering are frozen; exported-surface and Reporting inventories are recorded. |
| RM-04 | Source-catalog contract and generation | DONE | RM-03 | Authored catalog/schema/generator plus generated output | Exact two-surface, 20-direct-field, three-collection-field contract generates and drifts cleanly. |
| RM-05 | Runtime catalog integration | DONE | RM-04 | Owner-private catalog, consumers, tests, and authored harness accounting | Construction fails closed and duplicate routing maps are removed within the defined boundary. |
| RM-06 | Facade and test decomposition | DONE | RM-05 | Same-package source/test split and export cleanup | Mixed coordinator files and accidental exports are removed without behavior drift. |
| RM-07 | Admission decomposition | DONE | RM-06 | Admission source split | Four admission entry points and exact request hashes remain unchanged. |
| RM-08 | Projection boundary cleanup | DONE | RM-07 | Root contribution, caller updates, wrapper deletion, boundary evidence | Deleted wrapper has no callers and affected owner/boundary evidence passes. |
| RM-09 | Supersession/publication cohesion | DONE | RM-08 | Contract/file organization and focused tests | One mutation facade visibly retains every atomic supersession effect. |
| RM-10 | Harness accounting audit | DONE | RM-09 | Authored test-family changes and generated topology, if needed | Every changed selector resolves; generation, drift, policy, shape, and harness checks pass. |
| RM-11 | Validation and handoff completion | DONE | RM-10 | Final tracker evidence only | Full required ladder passes and every remediation row is complete or an evidenced no-op. |

### Workstream checkpoints

| Time | Workstream | Bound input | Changed files | Commands and retained roots | Result and next action |
| --- | --- | --- | --- | --- | --- |
| 2026-08-25T19:12:18-04:00 | RM-00 | HEAD `dd06b39a1ec901b0c610c87606546324552693da`; tracker `889a4b1d812d6d8df7c5bff7b1166ea0d2c493bd44d7516a357cb75391c297d6`; clean `main` | Tracker rebaseline in progress | Baselines pending | `RM-00` active; do not begin `RM-01`. |
| 2026-08-25T19:16:14-04:00 | RM-00 | Same bound HEAD; expected tracker-only dirty state | `docs/handoffs/tasksdecisions-module-refactor-tracker.md` | `make task-guide ROLE=module-author OWNER=module.tasksdecisions`; boundary pass `.cartulary/test-results/20260825T231351Z-p654381`; owner pass `.cartulary/test-results/20260825T231357Z-p654755`; service-backed pass `.cartulary/test-results/20260825T231448Z-p694756`; vulnerability baseline fail `.cartulary/test-results/20260825T231540Z-p734696`; Markdown pass `.cartulary/test-results/20260825T231635Z-p735930` | Baseline released. Vulnerability artifact confirms blocking `GO-2026-6253` through `go-archive v0.2.0`; `RM-01` may begin. |
| 2026-08-25T19:20:32-04:00 | RM-01 | RM-00 checkpoint complete | Core 01, Core 02, Core 04, Appendix F, and this tracker | Owner search confirms the removed field had no other product occurrence; `make lint-markdown` pass `.cartulary/test-results/20260825T232015Z-p738873` | Added `REQ-01-671` and `AC-565`, updated base claim and bidirectional trace, removed only the orphaned requirement, and confirmed no domain, NLSpec, schema, API, or data migration. `RM-02` may begin after checkpoint lint. |
| 2026-08-25T19:25:11-04:00 | RM-02 | RM-01 checkpoint complete | `go.mod`, module-managed `go.sum`, and this tracker | `go list -m all` resolves `go-archive v0.3.3` and Testcontainers `v0.42.0`; vulnerability pass `.cartulary/test-results/20260825T232132Z-p741382`; service-backed pass `.cartulary/test-results/20260825T232142Z-p742107`; toolchain pass `.cartulary/test-results/20260825T232235Z-p783919`; fast pass 431/431 `.cartulary/test-results/20260825T232244Z-p784411` | `RB-002` closed. Required transitive floors are recorded; there is no product, schema, or data migration. `RM-03` may begin after Markdown lint. |
| 2026-08-25T19:29:40-04:00 | RM-03 | RM-02 checkpoint complete | `mutation_admission_test.go`, `task_decisions_store_test.go`, and this tracker | `make format` pass `.cartulary/test-results/20260825T232723Z-p830616`; owner pass 19/19 `.cartulary/test-results/20260825T232731Z-p834517`; service-backed pass 15/15 `.cartulary/test-results/20260825T232822Z-p874872`; `git diff --check` pass | Create `d0a52c…e1c51`, patch `f15969…0d9b3`, conflict `1b3c4a…c0763`, and supersede `6f287b…9877d` are exact goldens. Supersession proves source-then-target intent order, two revisions, two intents, replay stability, and rejected-effect absence. Current exports were inventoried for `RM-06`; server assembly plus existing Reporting boundary/integration coverage make new Reporting behavior tests a no-op. `RM-04` may begin after Markdown lint. |
| 2026-08-25T19:43:26-04:00 | RM-04 | RM-03 checkpoint complete | `contracts/index.json`; `contracts/tasksdecisions/source-catalog.v1.json` and schema; `tools/contractgen/{main.go,validation.go,tasksdecisions_source_catalog.go,tasksdecisions_source_catalog_test.go}`; `tools/{generate_drift_scratch_inputs.json,execution_topology_render_index.json}`; `tools/harness/generated-artifacts/check-json-shapes.mjs`; generated `internal/gen/contracttasksdecisions/{artifacts_gen.go,source_catalog_gen.go}` and extension provenance; this tracker | Focused positive/negative generator tests pass; `make generate` pass `.cartulary/test-results/20260825T234014Z-p935046`; generation drift 4/4 `.cartulary/test-results/20260825T234046Z-p938521`; generated policy 3/3 `.cartulary/test-results/20260825T234057Z-p941479`; JSON shape 3/3 `.cartulary/test-results/20260825T234033Z-p938052`; fast suite 431/431 `.cartulary/test-results/20260825T234104Z-p942006`; checkpoint Markdown pass `.cartulary/test-results/20260825T234349Z-p983746` | The catalog generates exact `2/20/3` facts, joins canonical view facts, rejects missing/duplicate/cross-surface/read-only/unsafe/stale-reference/collection disagreements, and carries no wire or database migration. The first JSON-shape run exposed missing family registration (`.cartulary/test-results/20260825T233752Z-p933220`); the authored checker and generated topology were corrected and the gate passed. Extension output changed only because it binds the edited generator-source digests. Rollback unit is the authored catalog/schema/generator registration plus generated projections. `RM-05` may begin. |
| 2026-08-25T20:03:25-04:00 | RM-05 | RM-04 checkpoint complete | New `internal/modules/tasksdecisions/internal/sourcecatalog/{catalog.go,catalog_test.go}`; mutation/import/conflict/revision/projection consumers; private source direct-update and touch mechanics; `internal/app/revisionassembly/revisions.go`; `internal/testutil/collaborationsupport/assertions.go`; `tools/{backend_module_boundaries.json,test_families/module.tasksdecisions.json,execution_topology_render_index.json}`; strengthened supersession test; this tracker | `make generate` pass `.cartulary/test-results/20260825T235822Z-p1093100`; generation drift 4/4 `.cartulary/test-results/20260825T235857Z-p1096558`; generated policy 3/3 `.cartulary/test-results/20260825T235912Z-p1099531`; JSON shape 3/3 `.cartulary/test-results/20260825T235919Z-p1099971`; final owner 20/20 `.cartulary/test-results/20260826T000221Z-p1185875`; final service-backed 15/15 `.cartulary/test-results/20260826T000127Z-p1145573`; final boundary 3/3 `.cartulary/test-results/20260826T000313Z-p1226206`; checkpoint Markdown `.cartulary/test-results/20260826T000405Z-p1226918`; `git diff --check` pass | Direct SQL, conflict snapshots, surface/record lookup, reference roles and mirror links, collection admission, revision routes, import record types, and projection descriptors now consume one cached immutable catalog. Production construction is fallible; runtime negatives cover missing, duplicate, cross-surface, read-only, unsafe, stale-view, reference, and collection drift. Remaining view switches dispatch distinct typed task/decision algorithms and are intentionally not generic routing maps. Boundary run `.cartulary/test-results/20260825T235719Z-p1088326` exposed the missing private-catalog allowlist and disallowed owner-test SQL; both were repaired through authored policy/test support. Service run `.cartulary/test-results/20260825T235956Z-p1100577` then caught a missing test import; it was corrected before final evidence. Rollback unit is catalog loading plus all atomically updated consumers and the authored owner row. No wire, database, hash, or data migration. `RM-06` may begin. |
| 2026-08-25T20:22:58-04:00 | RM-06 | RM-05 checkpoint complete | Deleted `workbook_facade.go` and added `mutation_{contracts,construction,create,patch,references,collections,transactions,conflicts}.go`; deleted `task_decisions_store_test.go` and added task, decision/supersession, and shared store-test files; added `exported_surface_test.go`; updated projection-provider characterization references, boundary policy, authored owner selector, generated topology, and this tracker | Final format 2/2 `.cartulary/test-results/20260826T002304Z-p1471312`; owner 20/20 `.cartulary/test-results/20260826T001333Z-p1256190`; service-backed 15/15 `.cartulary/test-results/20260826T001426Z-p1297790`; Workbook 66/66 `.cartulary/test-results/20260826T001519Z-p1337829`; Projections 15/15 `.cartulary/test-results/20260826T002210Z-p1453954`; browser webserver-backed 58/58 `.cartulary/test-results/20260826T001747Z-p1397376`; boundary 3/3 `.cartulary/test-results/20260826T001739Z-p1396935`; generation drift 4/4 `.cartulary/test-results/20260826T002155Z-p1450992`; checkpoint Markdown `.cartulary/test-results/20260826T002338Z-p1475463`; `git diff --check` pass | Mutation files are 32–332 lines and behavior-cohesive; store evidence is split into 321-line task, 485-line decision/supersession, and 385-line shared-support units with every routed test name retained. The AST allowlist fixes the complete root package surface. `TaskDecisionRecordFieldKey`, `TaskCreateParams`, `DecisionCreateParams`, `TaskLifecycleState`, `DecisionMachineState`, `IsRecordRefCollectionField`, and `AllowsCollectionOp` are absent from `go doc`, with no aliases or shims. File-only moves created no new verification identity; only the new allowlist selector was added. Projection characterization references now name the real split files. Rollback unit is the source/test split, export cleanup, and metadata updates together. No observable or persistent migration. `RM-07` may begin. |
| 2026-08-25T20:28:01-04:00 | RM-07 | RM-06 checkpoint complete | Deleted `mutation_admission.go`; added `admission_{contracts,create,patch,conflict,supersession,values,hash}.go`; this tracker | Format 2/2 `.cartulary/test-results/20260826T002535Z-p1478001`; exact hash/allowlist diagnostic pass; owner 20/20 `.cartulary/test-results/20260826T002606Z-p1482548`; fast suite 432/432 `.cartulary/test-results/20260826T002659Z-p1523669`; focused Workbook admission mapping pass; checkpoint Markdown `.cartulary/test-results/20260826T002832Z-p1532428`; `git diff --check` pass | Admission files are 23–181 lines and separated by operation, shared decoding, and canonical hashing. The four public `Admit*JSON` signatures and `AdmissionFailure` remain exact. Goldens remain create `d0a52c…e1c51`, patch `f15969…0d9b3`, conflict `1b3c4a…c0763`, and supersede `6f287b…9877d`; normalization equivalence, meaningful differences, strict error precedence, limits, reference decoding, adapter mapping, and replay remain covered. No second hasher or compatibility layer exists. Rollback unit is the file-only admission split. No observable or persistent migration. `RM-08` may begin. |
| 2026-08-25T20:34:02-04:00 | RM-08 | RM-07 checkpoint complete | Added root `projection_contribution.go`; deleted `projectionprovider/contribution.go` and its empty directory; updated projection assembly, projection manifest test, projection test support, Projections boundary allowlist, backend boundary policy, root export allowlist, generated topology, and this tracker | Final format 2/2 `.cartulary/test-results/20260826T003247Z-p1600550`; `make generate` pass `.cartulary/test-results/20260826T003038Z-p1538949`; Tasks/Decisions 20/20 `.cartulary/test-results/20260826T003055Z-p1541875`; Projections 15/15 `.cartulary/test-results/20260826T003301Z-p1604683`; boundary 3/3 `.cartulary/test-results/20260826T003343Z-p1621367`; checkpoint Markdown `.cartulary/test-results/20260826T003430Z-p1622086`; deleted-path search, directory absence, and `git diff --check` pass | `tasksdecisions.NewProjectionContribution()` is the sole root constructor and the internal projection provider may now be imported only by that file. `workbookprojection` remains the typed cross-owner contract. Projections run `.cartulary/test-results/20260826T003148Z-p1583236` exposed one stale final-topology expectation for the new root import; the allowlist was corrected and fully rerun. No shim, forwarding package, provider identity, descriptor fact, row behavior, or persistent migration changed. Rollback unit is the root constructor, all callers, wrapper deletion, and boundary expectations together. `RM-09` may begin. |
| 2026-08-25T20:45:46-04:00 | RM-09 | RM-08 checkpoint complete | Moved supersession public contracts into `mutation_contracts.go`; renamed `supersede_facade.go` to `mutation_supersede.go`; retained `publication.go`; strengthened `decision_mutation_store_test.go`; added the injectable test-composition seam in `internal/testutil/appsupport/workbook.go`; updated `tools/backend_module_boundaries.json`; this tracker | Final format 2/2 `.cartulary/test-results/20260826T004327Z-p1760596`; owner 20/20 `.cartulary/test-results/20260826T004345Z-p1764718`; service-backed 15/15 `.cartulary/test-results/20260826T004437Z-p1805572`; Collaboration 32/32 `.cartulary/test-results/20260826T003928Z-p1710402`; boundary 3/3 `.cartulary/test-results/20260826T004527Z-p1845822`; checkpoint Markdown `.cartulary/test-results/20260826T004614Z-p1846503`; `git diff --check` and deleted-filename search pass | `MutationFacade.SupersedeDecision` remains the sole 374-line transaction owner; the 38-line publication helper remains shared and narrow. Public contracts now sit with other mutations, with no signature or export change. Injected failures before each of the two ordered Collaboration publications prove rollback of the link, both record/lifecycle updates, change set, revisions, any partial intent, and idempotency result; same-request retries execute once and preserve source-before-target ordering. No second service, compatibility shim, observable behavior change, or persistent migration was introduced. Rollback unit is the contract move, file rename, boundary path, and rollback characterization together. `RM-10` may begin. |
| 2026-08-25T20:48:26-04:00 | RM-10 | RM-09 checkpoint complete | Audited `tools/test_families/module.tasksdecisions.json`, all Tasks/Decisions Go test declarations, and generated topology; this tracker; no new authored selector or generated output was required in this slice | `make explain-test-owner OWNER=module.tasksdecisions` reports 13 real rows and six service-backed rows; generate pass `.cartulary/test-results/20260826T004724Z-p1848917`; drift 4/4 `.cartulary/test-results/20260826T004736Z-p1851818`; generated policy 3/3 `.cartulary/test-results/20260826T004749Z-p1854780`; JSON shape 3/3 `.cartulary/test-results/20260826T004753Z-p1855219`; harness contract 2/2 `.cartulary/test-results/20260826T004801Z-p1855727`; checkpoint Markdown `.cartulary/test-results/20260826T004901Z-p1856561`; `git diff --check` pass | The source-catalog tests have their dedicated owner row and the AST export allowlist extends the existing composition row. Store and admission decomposition retained existing test names, so file-only moves add no verification identity. The RM-09 rollback assertions are subtests beneath the already routed Decision lifecycle test. `make generate` was idempotent and every selector resolved through the harness contract. There were no failures, skips, or topology edits in this audit. Rollback unit is this tracker checkpoint only. `RM-11` may begin. |
| 2026-08-25T21:12:00-04:00 | RM-11 | RM-10 checkpoint complete | This tracker only; all product, contract, generator, test, dependency, and boundary edits were checkpointed by their owning earlier workstreams | Tasks/Decisions owner 20/20 `.cartulary/test-results/20260826T004956Z-p1858655` and service-backed 15/15 `.cartulary/test-results/20260826T005045Z-p1898756`; Workbook 66/66 `.cartulary/test-results/20260826T005134Z-p1938847`; Projections 15/15 `.cartulary/test-results/20260826T005356Z-p1997153`; Revisions 27/27 `.cartulary/test-results/20260826T005439Z-p2013968`; Collaboration 32/32 `.cartulary/test-results/20260826T005547Z-p2059842`; Reporting 5/5 `.cartulary/test-results/20260826T005719Z-p2107866`; Imports 23/23 `.cartulary/test-results/20260826T005808Z-p2124356`; boundary 3/3 `.cartulary/test-results/20260826T005926Z-p2167188`; generation drift 4/4 `.cartulary/test-results/20260826T005933Z-p2167547`; generated policy 3/3 `.cartulary/test-results/20260826T005945Z-p2170504`; JSON shape 3/3 `.cartulary/test-results/20260826T005949Z-p2170937`; OpenAPI 4/4 `.cartulary/test-results/20260826T005958Z-p2171387`; browser 58/58 `.cartulary/test-results/20260826T010006Z-p2172036`; fast 432/432 `.cartulary/test-results/20260826T010417Z-p2225632`; vulnerability 4/4 `.cartulary/test-results/20260826T010516Z-p2232727`; targeted gosec 4/4 `.cartulary/test-results/20260826T010537Z-p2233578`; build 7/7 `.cartulary/test-results/20260826T010559Z-p2263851`; finalization 1/1 `.cartulary/test-results/20260826T010627Z-p2297673`; full check 658/658 `.cartulary/test-results/20260826T010657Z-p2300699`; final Markdown `.cartulary/test-results/20260826T011409Z-p2429471`; repository search and `git diff --check` pass | The complete required ladder passes, including final vulnerability and full-check gates; no affected failure was dispositioned. `agent-finalize` ran without `RESULTS_DIR`, so retained-run maintenance was skipped. The worktree began clean and contains only the intended Iteration 2 changes; no schema, data, public API, frontend, or compatibility-shim migration exists. The dependency graph alone moves to `go-archive v0.3.3` with module-managed transitives. Rollback is by the independent units recorded in `RM-00..RM-10`; this completion checkpoint is tracker-only and the iteration is closed. |

### Iteration 2 acceptance

| Acceptance ID | Binary criterion | Evidence | Status |
| --- | --- | --- | --- |
| RM-AC-001 | The orphaned Decision classification token is repository-absent, and its former requirement remains coherent without replacement. | Repository-wide search; `RM-01` owner and trace review | PASS |
| RM-AC-002 | `REQ-01-671` and `AC-565` are bidirectionally traced and the base profile includes the acceptance criterion. | Core 01, Core 04, Appendix F; final check | PASS |
| RM-AC-003 | One versioned catalog defines exactly two surfaces, 20 direct fields, and three collections; generation and production construction fail closed on drift. | `RM-04..RM-05`; generated facts; negative generator/runtime tests | PASS |
| RM-AC-004 | Mutation/store and admission sources are behavior-cohesive; the seven retired exports and projection wrapper are absent without shims. | AST allowlist; deleted-path searches; `RM-06..RM-08` | PASS |
| RM-AC-005 | Four admission entry points preserve strict decoding, failure precedence, normalization, and exact canonical hash bytes. | Four hash goldens; owner, Workbook, and fast-suite roots | PASS |
| RM-AC-006 | Supersession remains one transaction-owning facade with one narrow publication helper; ordered effects, rollback, and replay are atomic. | Injected publication failures; Tasks/Decisions, Collaboration, and service-backed roots | PASS |
| RM-AC-007 | The resolved archive module is `v0.3.3`, Testcontainers remains `v0.42.0`, and security/full-check gates contain no historical advisory. | `go list -m`; vulnerability, gosec, and full-check roots | PASS |
| RM-AC-008 | Every added test has a real owner selector or existing routed parent; generated topology and harness contracts agree. | 13 owner rows/six service-backed; generation, drift, policy, shape, harness evidence | PASS |
| RM-AC-009 | Routes, OpenAPI, browser behavior, storage, hashes, authorization, lifecycle, events, and persisted-data posture remain compatible; no schema/data migration exists. | Affected owner ladder, OpenAPI, browser, build, finalization, full check, Git review | PASS |

### Iteration 2 decisions and migration posture

| Decision | Disposition |
| --- | --- |
| Public behavior | Preserve routes, wire shapes, view IDs and fields, lifecycle, authorization, transactions, events, and persisted data. |
| Orphaned Decision classification | Removed from the specification with no replacement, schema change, or data migration. |
| Field routing | Introduce one versioned Tasks/Decisions source catalog and fail closed on incomplete projections. |
| Internal compatibility | Remove explicitly retired internal aliases and the projection wrapper without shims; update repository callers atomically. |
| Supersession | Retain one atomic `MutationFacade` coordinator and the narrow publication helper. |
| Dependency | Retain Testcontainers `v0.42.0`; resolve `github.com/moby/go-archive` to `v0.3.3` through Go module tooling. |
| Domain and NLSpecs | `docs/domain.md` remains unchanged; Core 01, Core 02, and Core 04 already own the required normative scopes. |
