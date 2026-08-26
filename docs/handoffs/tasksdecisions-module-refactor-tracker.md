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
| Current status | `RS-08A`, Iteration 2, and Iteration 3, "Legacy Removal and Production Readiness," are complete historical work. `I3-00..I3-08` are closed with passing acceptance evidence. |
| Administrative authorization | The 2026-08-25 implementation request activates Iteration 3 against `main` commit `6bfc77b10cf0cba13991e3b94c833d8c84aaed51`, pre-implementation tracker SHA-256 `b3efdf265499dde020ab6234489ec2282e342878adb8af28b9f4a1dcd63b8047`, and a worktree whose only pre-existing change was the user-owned staged Section 14 planning draft in this file. |
| Presently authorized production change | Serial workstreams `I3-00..I3-08` in Section 14 after `I3-00` releases fresh baselines. The adopted Projections implementation-boundary decision and its non-normative Appendix I explanation may be amended before the projection code cutover. |
| Pending planned work | None for Iteration 3; `I3-00..I3-08` are `DONE`. |
| Non-goals | No public route, wire schema, lifecycle, authorization, event, transaction, database schema, frontend behavior, persisted-data, dependency, domain-vocabulary, or behavioral-owner change; no compatibility wrapper for intentionally retired internal Go APIs. |
| Generated files | Generated outputs MAY change only through `make generate`; generated roots MUST NOT be hand-edited. |

The following states are distinct and MUST NOT be conflated:

| State | Meaning | Current disposition |
| --- | --- | --- |
| Planning adoption | A document-only request records a decision-complete future iteration without authorizing its implementation. | Closed for Iteration 3 when the implementation request activated `I3-00`. |
| Administrative authorization | An explicit implementation request authorizes bounded production work against identified repository and tracker revisions. | Granted for Iteration 3 against the binding recorded above. |
| Baseline release | The owner, service-backed, affected-owner, backend-boundary, and full-check baselines pass before production edits. | Released for Iteration 3 by `I3-00`; no production edit preceded the baseline. |
| Slice progress | One independently revertible slice has its specified implementation and evidence. | `I3-00..I3-08` are `DONE`. |
| Refactor completion | Every required slice is `DONE` or an evidenced `NO-OP` and final handoff evidence is complete. | Reached for Iterations 2 and 3. |

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

Sections 2 through 12 preserve the Iteration 1 planning baseline and execution
history. Section 13 preserves the completed Iteration 2 evidence. Neither is a
current-state inventory or active plan; Section 14 is the sole current plan.

## 2. Iteration 1 Baseline Repository Inventory (Historical)

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

## 14. Iteration 3 Legacy Removal and Production Readiness

### Planning authority and baseline

The 2026-08-25 planning request authorized adoption of this section. The
planning inspection was performed on clean branch `main` at commit
`6bfc77b10cf0cba13991e3b94c833d8c84aaed51`, with pre-update tracker SHA-256
`91f28fa265b1cd9d2b385bacdb70aaca56673068ef753c69b8b7e5da52ce3052`.
These values establish planning provenance, not an implementation binding.

The 2026-08-25 implementation request activates `I3-00` against the commit,
tracker digest, and staged planning-draft state recorded in Section 1. It
authorizes the serial ledger after fresh baselines pass. It also resolves the
planning contradiction identified before execution: the adopted Projections
implementation-boundary decision MUST be amended before removing the
Tasks/Decisions-specific rebuild port, and Appendix I MUST be reconciled with
that decision. No production edit is permitted until `I3-00` is `DONE`.

`docs/domain.md` and behavioral owner specifications require no change. This
iteration changes no domain vocabulary or owner-required behavior. It amends
one adopted implementation-topology decision, reconciles its non-normative
explanation, removes dead, duplicated, misleading, or unnecessarily broad
repository-internal Go surfaces, and decomposes mixed implementation and test
units.

### Checkpoint protocol

The allowed Iteration 3 states are `PLANNED`, `IN_PROGRESS`, `BLOCKED`, `DONE`,
`NO-OP`, and `SUPERSEDED`. Every ledger row starts as `PLANNED`; no row is
active at adoption time.

The workstreams MUST execute strictly in ledger order and MUST NOT overlap.
After completing each workstream, and before activating the next, the session
MUST update this tracker with:

- the bound input commit and tracker digest;
- exact changed files and the independently revertible rollback unit;
- every command, retained result root, failure, repair, rerun, and skipped
  check with its reason;
- compatibility and migration posture;
- acceptance evidence and the exact next slice; and
- a passing `make lint-markdown` checkpoint.

An owner contradiction, unintended observable-contract drift, affected test
failure, backend-boundary failure, generated drift, unresolved security
failure, or unexplained worktree change MUST set the active row to `BLOCKED`
and stop later work. An affected failure MUST NOT be dispositioned as
acceptable. `NO-OP` is permitted only when repository evidence proves that a
planned change is already satisfied and records the validating commands.

Generated roots and generated topology MUST change only through
`make generate`. File-only moves MUST NOT create new verification identities.
Authored test rows may change only when a package or selected test identity
actually changes.

### Current findings and remediation decisions

Planning inspection found 56 Go files and 8,855 Go lines below the owner root.
The largest remaining mixed units are the 563-line Decision mutation test, the
415-line Incident Bundle portability implementation, and the 410-line rollback
provider. The active projection contract also exposes unused rebuild surfaces
and combines opposite integration directions under the Workbook-specific
package name. Repository-wide symbol tracing found the specific dead or
duplicated surfaces governed below; it did not justify removing any adopted
Task Request or Decision feature.

#### I3-G1 Directional projection boundary

- **Remediation and areas:** Replace
  `internal/modules/tasksdecisions/workbookprojection` with
  `projectioncontract` for source-directed inputs, source readers, and the
  projection `Contribution`, and `projectionports` for consumer-directed
  mutation and Reporting ports. Update implementation, the code-backed
  provider manifest, application assembly, boundary policy, tests, and harness
  accounting together.
- **Runtime contract:** `projectionports.MutationRows` contains only
  `RefreshTaskRequestTx`, `LoadTaskRequestTx`, `RefreshDecisionTx`, and
  `LoadDecisionTx`. `projectionports.ReportingReader` directly declares the
  two existing derived-fact collection methods and retains the current Task
  and Decision derived-fact value shapes.
- **Removal:** Delete the old `Rows`, `Ports`, `Rebuilder`, and `TaskReader`
  interfaces, package-level `Descriptors` and `SurfaceIntents`, aggregate
  `TaskDecisionPorts`, and every unused task-specific rebuild method on the
  Projections runtime and row adapter. Do not add a forwarding package, alias,
  or replacement aggregate.
- **Rationale and long-term benefit:** The two packages make integration
  direction and ownership explicit, apply least authority to mutation and
  import callers, and prevent dead restore APIs from becoming a compatibility
  obligation. New projection capabilities can grow without expanding one
  Workbook-named kitchen-sink contract.
- **Compatibility and migration:** This is an atomic repository-internal Go
  import and signature break. Provider IDs, capabilities, rebuild ordering,
  view-schema IDs, semantic intents, projection tables, query behavior, and
  persisted rows remain unchanged. The provider manifest changes only its
  Tasks/Decisions facade package path.
- **Risk if unresolved:** Mutation, Imports, Reporting, Recovery, and
  Projections remain unnecessarily coupled, and unused rebuild entry points
  may be mistaken for supported APIs or evolve inconsistently with the generic
  catalog-driven rebuild path.
- **Validation:** The old directory and imports are absent; exact export locks
  pass for both replacement packages; mutation refresh/load and Reporting fact
  behavior remain equivalent; generic incident and restore rebuild tests pass;
  provider-manifest, Projections, Reporting, Imports, Workbook, Recovery,
  boundary, generation, and harness evidence is green.

#### I3-G2 Root contribution facade cleanup

- **Remediation and areas:** Replace `IncidentBundleContribution` and
  `NewIncidentBundleContribution` with `NewIncidentBundleSourcePort` and
  `IncidentBundleSubtypeContribution`. Make the Reporting implementation
  private and make `NewReportingContribution` return
  `exportprovider.FieldProvider`. Rename `NewRecoveryContribution` to
  `RecoveryStateContribution`. Remove the owner-local `ImportCreateCommand`
  alias and use the Imports owner-facade command type internally. Update root
  implementation, application assembly, boundary tests, and affected owner
  tests atomically.
- **Rationale and long-term benefit:** Each root constructor will expose one
  owner contribution without publishing an otherwise-unused concrete wrapper
  or local alias. Composition roots receive the narrow capability they need,
  making future owner contributions independently extensible.
- **Compatibility and migration:** Repository-internal Go callers break and
  are updated in the same slice. There are no compatibility shims and no wire,
  database, data, or provider-identity changes.
- **Risk if unresolved:** Thin concrete wrappers and combined contributions
  accumulate accidental API commitments and force unrelated consumers to
  depend on one aggregate shape.
- **Validation:** Removed declarations have zero definitions and callers;
  Incident Bundles, Reporting, Recovery, Imports, server composition, root
  export locks, and boundary checks pass with identical provider outputs and
  recovery-state declarations.

#### I3-G3 Mutation indirection and duplicate conflict type

- **Remediation and areas:** Remove `MemberReferenceCapability`,
  `NewMemberReferenceCapability`, the corresponding dependency and facade
  fields, and Workbook assembly injection. Mutation admission calls the
  existing owner-private member-reference validator directly, as Imports
  already does. Replace `SupersedeRowVersionConflictError` with the identical
  common `RowVersionConflictError`. Rename `workbook_conflict.go` to
  `mutation_conflict_resolution.go`; retain `mutation_conflicts.go` for
  conflict construction helpers.
- **Rationale and long-term benefit:** Owner-internal validation no longer
  leaves the module solely to be injected back into the same owner. One row
  conflict type eliminates duplicate error evolution and gives every mutation
  path one stable internal vocabulary. Mutation files describe owner behavior
  instead of a downstream adapter.
- **Compatibility and migration:** Internal constructor and error-type callers
  change atomically, with no aliases. Validation order, error details,
  idempotency, transaction ownership, rollback, supersession ordering, and
  Workbook failure mapping remain exact.
- **Risk if unresolved:** Duplicate errors can drift in fields or mapping, and
  the self-injected validator creates needless composition coupling and a false
  extension point.
- **Validation:** Create and patch member-reference cases preserve success and
  failure precedence; supersession maps the common row conflict identically;
  owner, service-backed, Workbook, Imports, Collaboration, replay, rollback,
  and export-surface checks pass.

#### I3-G4 Test topology and dead support

- **Remediation and areas:** Make shared external-test helpers private, move
  the misplaced Task state helper into shared support, remove the unused
  terminal-state catalog parameter and its dead construction, and split
  Decision mutation coverage from supersession coverage without renaming
  routed tests. Update authored selectors only if a selected test identity
  actually changes.
- **Rationale and long-term benefit:** Test support stops presenting accidental
  API-like names, behavior suites become easier to navigate, and dead setup no
  longer obscures real dependencies.
- **Compatibility and migration:** Test-only source organization changes. Test
  names, assertions, service fixtures, and harness identities remain stable.
- **Risk if unresolved:** Future maintainers can mistake test helpers for
  reusable contracts, while oversized suites and unused construction hide
  ownership and make regressions harder to localize.
- **Validation:** All former test names still resolve, no dead parameter or
  construction remains, owner/service-backed slices pass, and harness contract
  reports every selector as real.

#### I3-G5 Portability and rollback cohesion

- **Remediation and areas:** Replace the Incident Bundle `portability.go`
  catch-all with cohesive export, portable-value decoding, import preparation,
  and import application units. Replace rollback `provider.go` with task
  provider, decision provider, and shared value-decoding units. Keep the same
  packages and do not introduce generic reflection, dynamic dispatch, or a
  shared cross-owner framework.
- **Rationale and long-term benefit:** Export, strict admission, persistence,
  Task lifecycle restoration, and Decision machine restoration evolve in
  isolated files while retaining high owner cohesion.
- **Compatibility and migration:** Same-package code movement only. Bundle
  paths, NDJSON bytes, invariant identifiers, attribution, SQL, snapshot
  schemas, reference validation, and rollback results remain exact.
- **Risk if unresolved:** Changes to one portability phase or record family
  continue to require editing mixed high-risk files, increasing regression and
  review cost.
- **Validation:** Byte and ordering fixtures, malformed/duplicate/cross-
  incident negatives, attribution, atomic import, nullable rollback,
  lifecycle, reference, revision, and service-backed tests all pass.

#### I3-G6 Durable boundary and completion enforcement

- **Remediation and areas:** Upgrade the AST surface lock to include exported
  constants, variables, types, functions, and methods in the owner root and
  both new projection packages. Add a negative fixture, cohesive declaration
  assertions, and absence checks for retired symbols, catch-all files, and the
  old package path. Reconcile backend boundaries, provider metadata, authored
  test routing, and generated topology.
- **Rationale and long-term benefit:** Production readiness becomes an
  enforced repository property rather than a one-time review conclusion.
- **Compatibility and migration:** Test, boundary-policy, machine-metadata,
  harness, and generated-topology changes only. File moves do not create new
  test identities.
- **Risk if unresolved:** Dead surfaces or stale paths can return silently,
  exported methods can bypass the current top-level-only guard, and moved tests
  can lose owner routing.
- **Validation:** Negative AST fixtures fail as designed; retired-path and
  declaration searches are empty; `make generate`, drift, generated-policy,
  JSON-shape, boundary, and harness-contract gates pass.

### Bound internal interface disposition

`I3-01` freezes the complete current top-level export inventory below. A name
listed as retained remains intentionally public in the same package unless a
later row explicitly names its destination. No unlisted compatibility alias or
forwarder is permitted.

| Current root export | Iteration 3 disposition |
| --- | --- |
| `AdmissionFailure`, `AdmitConflictResolveJSON`, `AdmitCreateJSON`, `AdmitPatchJSON`, `AdmitSupersedeJSON`, `CollectionAction`, `CollectionActionPayload`, `ConflictClaims`, `ConflictCommand`, `ConflictResolveRequest`, `ConflictResolveRequestHash`, `CreateCommand`, `CreateRequest`, `CreateRequestHash`, `DecisionsViewSchemaID`, `ErrClientTxnConflict`, `ErrIdempotencyNotFound`, `ErrStoredMutationKindMismatch`, `FieldValue`, `IdempotencyCapability`, `IdempotencyKey`, `IdempotencyRecord`, `ImportDependencies`, `ImportLinkCapability`, `ImportRecordEnvelopeCapability`, `ImportRevisionCapability`, `IncidentStateCapability`, `LifecycleValidationError`, `LinkCapability`, `MutationDependencies`, `MutationFacade`, `MutationResult`, `NewImportContribution`, `NewMutationContribution`, `NewProjectionContribution`, `NewReportingContribution`, `NewRevisionContribution`, `NewStoredCreateResult`, `NewStoredDecisionSupersessionResult`, `NewStoredPatchResult`, `OptionalConflictValue`, `PatchChange`, `PatchCommand`, `PatchRequest`, `PatchRequestHash`, `RecordEnvelopeCapability`, `RevisionCapability`, `RowVersionConflictError`, `SameFieldConflict`, `SameFieldConflictError`, `StoredDecisionSupersessionResult`, `StoredMutationCreate`, `StoredMutationDecisionSupersession`, `StoredMutationKind`, `StoredMutationPatch`, `StoredMutationResult`, `StoredRowMutationResult`, `SupersedeCommand`, `SupersedeFacts`, `SupersedeMutationResult`, `SupersedeRequest`, `SupersedeRequestHash`, `TaskRequestsViewSchemaID`, `ValidationError` | Retain. `NewReportingContribution` keeps its name but returns the narrow `exportprovider.FieldProvider` interface instead of an exported concrete implementation. |
| `ImportCreateCommand` | Delete; use `ownerfacade.ImportOwnerCreateCommand` internally. |
| `IncidentBundleContribution`, `NewIncidentBundleContribution` | Delete; replace the aggregate constructor with `NewIncidentBundleSourcePort` and `IncidentBundleSubtypeContribution`. |
| `MemberReferenceCapability`, `NewMemberReferenceCapability` | Delete; call owner-private member validation directly. |
| `NewRecoveryContribution` | Rename to `RecoveryStateContribution` without an alias. |
| `ReportingContribution` | Make private; do not replace it with another exported concrete type. |
| `SupersedeRowVersionConflictError` | Delete; supersession returns the common `RowVersionConflictError`. |

| Current `workbookprojection` export | Iteration 3 disposition |
| --- | --- |
| `TaskRequestProjectionInput`, `TaskRequestProjectionInputPage`, `TaskRequestSourceReader`, `DecisionProjectionInput`, `DecisionProjectionInputPage`, `DecisionSourceReader`, `Contribution`, `NewContribution` | Move without semantic change to `projectioncontract`. |
| `Contribution.ProjectionContribution`, `Contribution.TaskRequestSource`, `Contribution.DecisionSource` | Retain as the three accessors on `projectioncontract.Contribution`. |
| `TaskDerivedFact`, `DecisionDerivedFact` | Move without semantic change to `projectionports`. |
| `Rows` | Replace with four-method `projectionports.MutationRows`; omit both transaction rebuild methods. |
| `Reader`, `TaskReader` | Replace with one directly declared, two-method `projectionports.ReportingReader`. |
| `Ports`, `Rebuilder` | Delete without replacement; expose mutation rows and Reporting reader separately and retain generic catalog-driven rebuild. |
| `Descriptors`, `SurfaceIntents` | Delete package-level test conveniences; inspect the constructed contribution's immutable facts. |
| `tasksdecisions/workbookprojection` | Delete after the atomic move; retain no forwarding directory or import. |

The intentional breaks above are internal to the repository. There is no HTTP,
OpenAPI, TypeScript, WebSocket, database, persisted-data, request-hash,
lifecycle, authorization, frontend, dependency, domain, or adopted-owner
migration. All callers MUST be updated atomically, and no compatibility alias,
forwarder, or dual path may remain.

### Iteration 3 ledger

| ID | Workstream | Status | Depends on | Rollback unit | Exit evidence |
| --- | --- | --- | --- | --- | --- |
| I3-00 | Authorization rebind and baseline | DONE | none | Section 14 activation, bound inputs, and baseline record | Bound `main` commit, tracker digest, and staged draft state are recorded; Tasks/Decisions unit/service, all eight affected-owner unit/service pairs, boundary, and 658-unit full-check baselines pass before production edits. |
| I3-01 | Specification alignment and characterization freeze | DONE | I3-00 | Adopted decision, Appendix I, export disposition, and focused characterization tests | The adopted topology assigns Tasks/Decisions rebuild exclusively to generic catalog coordination; the complete top-level export inventory is dispositioned; contribution facts, generic rebuild order/failure, Reporting facts/failure, and existing mutation ordering are routed and green. |
| I3-02 | Directional projection-contract split | DONE | I3-01 | New packages, all callers, manifest, boundaries, tests, and topology | Exact source and consumer contracts compile across all callers; mutation and Reporting are exposed separately; generic rebuild consumers pass; old package and task-specific rebuild APIs are absent. |
| I3-03 | Contribution facade cleanup | DONE | I3-02 | Root contribution APIs and every application caller | Incident Bundle source/subtype capabilities are independent, Reporting returns an interface, Recovery uses owner-standard naming, Imports uses the owner command directly, and every legacy wrapper/alias is absent. |
| I3-04 | Mutation API and error cleanup | DONE | I3-03 | Mutation dependencies, validator calls, error type, adapter, and file rename | Member validation stays owner-private, all row conflicts use one type, the conflict-resolution filename is owner-semantic, and mutation/failure behavior is equivalent. |
| I3-05 | Test topology cleanup | DONE | I3-04 | Test support and behavior-file organization | Helpers are private, dead setup is gone, behavior files are cohesive, preserved test names pass, and every authored selector remains real. |
| I3-06 | Portability and rollback decomposition | DONE | I3-05 | Same-package provider source splits | Mixed provider files are absent; bundle bytes, SQL, validation, attribution, lifecycle, references, and rollback behavior remain exact. |
| I3-07 | Boundary and harness accounting | DONE | I3-06 | Export locks, absence guards, boundary policy, authored routing, and generated topology | Root and replacement-package surfaces are exact; retired paths are rejected; generation, drift, policy, shape, boundary, and harness gates pass. |
| I3-08 | Validation and handoff completion | DONE | I3-07 | Final tracker evidence only | Every row is `DONE` or an evidenced `NO-OP`; the final ladder passes, the worktree contains only intended changes, and the iteration is unambiguously closed. |

### Iteration 3 workstream checkpoints

| Time | Workstream | Bound input | Changed files and rollback | Commands and retained roots | Result, compatibility, and next action |
| --- | --- | --- | --- | --- | --- |
| 2026-08-25T23:15:00-04:00 | I3-00 | HEAD `6bfc77b10cf0cba13991e3b94c833d8c84aaed51`; pre-implementation tracker SHA-256 `b3efdf265499dde020ab6234489ec2282e342878adb8af28b9f4a1dcd63b8047`; `main` worktree initially contained only the user-owned staged planning draft in this tracker | `docs/handoffs/tasksdecisions-module-refactor-tracker.md`; rollback is the Iteration 3 authorization rebind and baseline checkpoint only | `make task-guide ROLE=module-author OWNER=module.tasksdecisions` and `make explain-test-owner OWNER=module.tasksdecisions` pass; Tasks/Decisions unit `.cartulary/test-results/20260826T024800Z-p2464074` and service `.cartulary/test-results/20260826T024847Z-p2504318`; Projections unit `.cartulary/test-results/20260826T024939Z-p2544429` and service `.cartulary/test-results/20260826T025016Z-p2560974`; Workbook unit `.cartulary/test-results/20260826T025053Z-p2577487` and service `.cartulary/test-results/20260826T025306Z-p2635280`; Imports unit `.cartulary/test-results/20260826T025519Z-p2692987` and service `.cartulary/test-results/20260826T025628Z-p2735377`; Reporting unit `.cartulary/test-results/20260826T025737Z-p2777713` and service `.cartulary/test-results/20260826T025818Z-p2793925`; Recovery unit `.cartulary/test-results/20260826T025858Z-p2810171` and service `.cartulary/test-results/20260826T030012Z-p2862655`; Incident Bundles unit `.cartulary/test-results/20260826T030126Z-p2915134` and service `.cartulary/test-results/20260826T030222Z-p2931538`; Revisions unit `.cartulary/test-results/20260826T030318Z-p2947907` and service `.cartulary/test-results/20260826T030421Z-p2992852`; Collaboration unit `.cartulary/test-results/20260826T030524Z-p3037817` and service `.cartulary/test-results/20260826T030651Z-p3085624`; boundary `.cartulary/test-results/20260826T030821Z-p3133528`; full check 658/658 `.cartulary/test-results/20260826T030823Z-p3133960`; initial `make lint-markdown CARTULARY_TEST_RESULTS_DIR=... CARTULARY_TEST_RUN_ID=...` invocation rejected before execution because harness identity is not an allowed Make command-line input; corrected environment-scoped checkpoint Markdown `.cartulary/test-results/i3-00-checkpoint` | Baseline released before production edits. The only failure was the recorded checkpoint-command configuration error; no test or Markdown content failed, no check was skipped, and the supported rerun passed. This checkpoint changes planning and execution metadata only: no specification, product, schema, data, public contract, frontend, dependency, domain, or compatibility migration exists. After the recorded Markdown pass, activate `I3-01`; amend the adopted Projections implementation-boundary decision and Appendix I before any projection cutover. |
| 2026-08-25T23:26:05-04:00 | I3-01 | `I3-00` checkpoint complete against the same bound HEAD and preserved staged planning draft | `docs/decisions/projections-module-boundary.md`; `docs/spec/I_projection_authority_boundary_and_characterization.md`; `internal/modules/projections/internal/runtime/provider_registry_test.go`; new `internal/modules/tasksdecisions/reporting_contribution_test.go`; authored `tools/test_families/{module.projections,module.tasksdecisions}.json`; generated `tools/execution_topology_render_index.json`; this tracker. Rollback is the adopted-decision/Appendix amendment, characterization tests, authored selector changes, generated index, and checkpoint together. | Format 2/2 `.cartulary/test-results/20260826T032032Z-p3249756`; initial generate `.cartulary/test-results/20260826T032038Z-p3253680` and final generate `.cartulary/test-results/20260826T032425Z-p3335690`; Tasks/Decisions 20/20 `.cartulary/test-results/20260826T032100Z-p3256727`; initial Projections 15/15 `.cartulary/test-results/20260826T032150Z-p3297809` and final expanded selector 15/15 `.cartulary/test-results/20260826T032438Z-p3338630`; Reporting 5/5 `.cartulary/test-results/20260826T032234Z-p3314798`; initial drift/policy/shape/harness `.cartulary/test-results/20260826T032324Z-p3331151`, `.cartulary/test-results/20260826T032335Z-p3334108`, `.cartulary/test-results/20260826T032338Z-p3334535`, `.cartulary/test-results/20260826T032345Z-p3335049`; final drift `.cartulary/test-results/20260826T032523Z-p3355445`, policy `.cartulary/test-results/20260826T032531Z-p3358352`, shape `.cartulary/test-results/20260826T032532Z-p3358762`, and harness `.cartulary/test-results/20260826T032536Z-p3359241`; `git diff --check` pass; checkpoint Markdown `.cartulary/test-results/i3-01-checkpoint` | No command failed or was skipped. The second generation/gate pass incorporates the expanded projection-contribution characterization selector, not a repair. The adopted decision now authorizes generic-only Tasks/Decisions rebuild; Appendix I explains the directional contracts; every current top-level export is dispositioned. Generic incident/import rebuild order and first-failure propagation, immutable contribution facts, Reporting order/copy/failures, and existing source-before-target mutation ordering are protected. This is an internal topology amendment with no Core behavior, `docs/domain.md`, wire, schema, data, frontend, dependency, provider identity, or compatibility migration. After Markdown passes, activate `I3-02` for the atomic package and caller cutover. |
| 2026-08-25T23:53:02-04:00 | I3-02 | `I3-01` checkpoint complete against the same bound HEAD and intended cumulative worktree | Added `internal/modules/tasksdecisions/projectioncontract/{contribution.go,contribution_test.go}` and `projectionports/ports.go`; deleted `workbookprojection/{contribution.go,contribution_test.go}` and its directory; updated Tasks/Decisions projection construction, mutation/import/Reporting consumers and provider imports/tests; Projections adapter, runtime, query, storage, and tests; projection/import/workbook/server application assembly and test support; `contracts/projection-providers/index.json`; authored boundary and Projections test-family inputs; generated topology render index; this tracker. Rollback is the two-package split, every atomic caller/manifest/policy/selector migration, old-package deletion, and checkpoint together. | Format 2/2 `.cartulary/test-results/20260826T033220Z-p3363343`; Tasks/Decisions 20/20 `.cartulary/test-results/20260826T033227Z-p3367325` and service 15/15 `.cartulary/test-results/20260826T033445Z-p3429741`; Projections 15/15 `.cartulary/test-results/20260826T033324Z-p3409051` and service 11/11 `.cartulary/test-results/20260826T033531Z-p3470002`; generate `.cartulary/test-results/20260826T033414Z-p3426243`; boundary 3/3 `.cartulary/test-results/20260826T033432Z-p3429340`; Workbook 66/66 `.cartulary/test-results/20260826T033615Z-p3486565` and service 37/37 `.cartulary/test-results/20260826T033827Z-p3544848`; Imports 23/23 `.cartulary/test-results/20260826T034038Z-p3602591` and service 14/14 `.cartulary/test-results/20260826T034148Z-p3645421`; Reporting 5/5 `.cartulary/test-results/20260826T034304Z-p3687827` and service 4/4 `.cartulary/test-results/20260826T034344Z-p3704284`; Recovery 24/24 `.cartulary/test-results/20260826T034424Z-p3720557` and service 19/19 `.cartulary/test-results/20260826T034538Z-p3773833`; Incident Bundles 8/8 `.cartulary/test-results/20260826T034700Z-p3826429` and service 6/6 `.cartulary/test-results/20260826T034757Z-p3843047`; Revisions 27/27 `.cartulary/test-results/20260826T034854Z-p3859430` and service 20/20 `.cartulary/test-results/20260826T034959Z-p3905328`; drift `.cartulary/test-results/20260826T035113Z-p3950299`; generated policy `.cartulary/test-results/20260826T035121Z-p3953213`; JSON shape `.cartulary/test-results/20260826T035122Z-p3953623`; harness `.cartulary/test-results/20260826T035126Z-p3954095`; fast 432/432 `.cartulary/test-results/20260826T035144Z-p3954720`; corrected absence/export inventory search and `git diff --check` pass; checkpoint Markdown `.cartulary/test-results/i3-02-checkpoint` | `projectioncontract` now contains only the six source value/reader types plus `Contribution`, its constructor, and three accessors; `projectionports` contains only two fact types, four-method `MutationRows`, and two-method `ReportingReader`. Projection assembly exposes those ports separately. All task-specific adapter/runtime rebuild methods, aggregate ports, old imports, and old directory are absent; generic incident/import/Revisions/restore behavior is green. The first absence command incorrectly matched legitimate other-owner `Rows`/`Rebuilder` interfaces and returned nonzero; an owner-qualified rerun passed. No product test failed or was skipped. Provider IDs, order, capabilities, tables, schemas, facts, storage, queries, public contracts, data, and dependencies are unchanged; internal Go callers migrated atomically with no shim or migration. After Markdown passes, activate `I3-03` for root contribution cleanup. |
| 2026-08-26T00:05:30-04:00 | I3-03 | `I3-02` checkpoint complete against the same bound HEAD and intended cumulative worktree | `internal/modules/tasksdecisions/{incident_bundle_contribution.go,incident_bundle_source_port_test.go,reporting_contribution.go,recovery_state.go,import_create.go,exported_surface_test.go}`; `internal/app/{incidentportabilityassembly/catalog.go,recoveryassembly/state_catalog.go}`; this tracker. Rollback is the four root-facade changes, their exact callers/export lock, and checkpoint together. | Format 2/2 `.cartulary/test-results/20260826T035507Z-p3963227`; Tasks/Decisions 20/20 `.cartulary/test-results/20260826T035515Z-p3967170`; Incident Bundles 8/8 `.cartulary/test-results/20260826T035613Z-p4008568` and service 6/6 `.cartulary/test-results/20260826T035710Z-p4025180`; Reporting 5/5 `.cartulary/test-results/20260826T035806Z-p4041561` and service 4/4 `.cartulary/test-results/20260826T035845Z-p4057999`; Recovery 24/24 `.cartulary/test-results/20260826T035932Z-p4074295` and service 19/19 `.cartulary/test-results/20260826T040046Z-p4127635`; Imports 23/23 `.cartulary/test-results/20260826T040202Z-p4180505` and service 14/14 `.cartulary/test-results/20260826T040311Z-p29248`; boundary 3/3 `.cartulary/test-results/20260826T040428Z-p71796`; fast 432/432 `.cartulary/test-results/20260826T040430Z-p72217`; legacy-declaration/caller and replacement-call search plus `git diff --check` pass; checkpoint Markdown `.cartulary/test-results/i3-03-checkpoint` | No command failed or was skipped. Incident Bundle source and subtype contributions are independently composed; the Reporting implementation is private and returned only as `exportprovider.FieldProvider`; Recovery uses `RecoveryStateContribution`; import code directly accepts `ownerfacade.ImportOwnerCreateCommand`. `IncidentBundleContribution`, `NewIncidentBundleContribution`, exported `ReportingContribution`, `NewRecoveryContribution`, and `ImportCreateCommand` have no Tasks/Decisions declaration or caller. Provider output, subtype catalog, Recovery inventory, import atomicity, and server-wide compilation remain equivalent. This is an internal Go surface break with no wire, provider identity, schema, data, frontend, dependency, or compatibility migration. After Markdown passes, activate `I3-04` for mutation validation and conflict cleanup. |
| 2026-08-26T00:20:41-04:00 | I3-04 | `I3-03` checkpoint complete against the same bound HEAD and intended cumulative worktree | `internal/modules/tasksdecisions/{mutation_capabilities.go,mutation_construction.go,mutation_create.go,mutation_patch.go,mutation_contracts.go,mutation_supersede.go,mutation_composition_test.go,exported_surface_test.go}`; deleted `workbook_conflict.go` and added `mutation_conflict_resolution.go`; `internal/app/workbookassembly/{tasksdecisions_capabilities.go,action_adapters.go}`; this tracker. Rollback is the validator de-injection, common conflict type, filename move, Workbook mapping, export lock, and checkpoint together. | Format 2/2 `.cartulary/test-results/20260826T040803Z-p80750`; Tasks/Decisions 20/20 `.cartulary/test-results/20260826T040810Z-p84704` and service 15/15 `.cartulary/test-results/20260826T040907Z-p126193`; Workbook 66/66 `.cartulary/test-results/20260826T040953Z-p166426` and service 37/37 `.cartulary/test-results/20260826T041206Z-p224488`; Imports 23/23 `.cartulary/test-results/20260826T041424Z-p282287` and service 14/14 `.cartulary/test-results/20260826T041534Z-p325061`; Collaboration 32/32 `.cartulary/test-results/20260826T041643Z-p367431` and service 23/23 `.cartulary/test-results/20260826T041810Z-p415820`; boundary 3/3 `.cartulary/test-results/20260826T041941Z-p463821`; fast 432/432 `.cartulary/test-results/20260826T041943Z-p464242`; retired-symbol/file, direct-validator/common-conflict, and `git diff --check` searches pass; checkpoint Markdown `.cartulary/test-results/i3-04-checkpoint` | No command failed or was skipped. Create and patch call the existing owner-private member validator at the same reference-admission position; mutation dependencies and Workbook assembly no longer inject an owner capability back into its owner. Supersession now returns `RowVersionConflictError`, and the action adapter maps the same safe fields/status/code as before. `MemberReferenceCapability`, its constructor/fields, `SupersedeRowVersionConflictError`, and `workbook_conflict.go` are absent; `mutation_conflict_resolution.go` retains the exact conflict-resolution implementation. Member-reference precedence, source-before-target publication, replay, rollback, transaction ownership, and rejected-effect absence remain green. This internal cleanup has no wire, schema, data, frontend, dependency, or compatibility migration. After Markdown passes, activate `I3-05` for test topology cleanup. |
| 2026-08-26T00:30:53-04:00 | I3-05 | `I3-04` checkpoint complete against the same bound HEAD and intended cumulative worktree | `internal/modules/tasksdecisions/mutation_store_test_support_test.go`; deleted `decision_mutation_store_test.go`; added `decision_supersession_store_test.go` and `decision_lifecycle_store_test.go`; `task_mutation_store_test.go`; this tracker. Rollback is the private-helper rename, shared-support move, behavior-file split, dead catalog removal, and checkpoint together. | Initial format 2/2 `.cartulary/test-results/20260826T042558Z-p474405`; the first owner invocation used invalid `OWNER=tasksdecisions` and exited with a usage error before creating a run root; the corrected owner run exposed two mechanical compile errors and retained 18/20 evidence at `.cartulary/test-results/20260826T042616Z-p478673`; restoring the real `MutationFacade.Patch` method call and `PatchChange.Collection` field followed by format 2/2 `.cartulary/test-results/20260826T042730Z-p519847` produced Tasks/Decisions 20/20 `.cartulary/test-results/20260826T042739Z-p523792` and service 15/15 `.cartulary/test-results/20260826T042829Z-p564429`; harness 2/2 `.cartulary/test-results/20260826T042944Z-p604936`; boundary 3/3 `.cartulary/test-results/20260826T043000Z-p605527`; fast 432/432 `.cartulary/test-results/20260826T043006Z-p605978`; exact pre/post `Test*` inventory comparison, external-test export search, dead-catalog/file search, and `git diff --check` pass; checkpoint Markdown `.cartulary/test-results/i3-05-checkpoint` | Both failures were related to command naming or the mechanical helper rename and were corrected in-slice; no validation was skipped. All shared external-test helpers and types are private, including `taskState` in shared support. Supersession behavior and helpers live in `decision_supersession_store_test.go`; terminal lifecycle behavior lives in `decision_lifecycle_store_test.go`; the ignored catalog parameter and construction are absent. Every prior `Test*` identity is byte-for-byte preserved, so authored routing and generated topology required no change; the owner slice and harness prove selectors resolve. This is test-only organization with no production, wire, schema, data, frontend, dependency, or compatibility migration. After Markdown passes, activate `I3-06` for same-package portability and rollback decomposition. |
| 2026-08-26T00:45:20-04:00 | I3-06 | `I3-05` checkpoint complete against the same bound HEAD and intended cumulative worktree | Deleted `internal/modules/tasksdecisions/internal/providers/incidentbundle/portability.go` and added `export.go`, `portable_values.go`, `import_prepare.go`, and `import_apply.go`; deleted `internal/modules/tasksdecisions/internal/providers/rollback/provider.go` and added `task_provider.go`, `decision_provider.go`, and `value_decode.go`; `tools/backend_module_boundaries.json`; this tracker. Rollback is both same-package decompositions, the exact boundary allowlist update, and checkpoint together. | Format 2/2 `.cartulary/test-results/20260826T043632Z-p614610`; focused portability/rollback rows 4/4 `.cartulary/test-results/20260826T043643Z-p618571`; Tasks/Decisions 20/20 `.cartulary/test-results/20260826T043733Z-p635626` and service 15/15 `.cartulary/test-results/20260826T043822Z-p676417`; Incident Bundles 8/8 `.cartulary/test-results/20260826T043913Z-p716729` and service 6/6 `.cartulary/test-results/20260826T044014Z-p733413`; Revisions 27/27 `.cartulary/test-results/20260826T044115Z-p749893` and service 20/20 `.cartulary/test-results/20260826T044223Z-p795769`; the first boundary run correctly rejected stale exact file grants at 2/3 `.cartulary/test-results/20260826T044332Z-p840987`; after replacing them with least-authority new-file grants, boundary 3/3 `.cartulary/test-results/20260826T044413Z-p841908`; JSON shape 3/3 `.cartulary/test-results/20260826T044420Z-p842272`; fast 432/432 `.cartulary/test-results/20260826T044429Z-p842798`; catch-all path/policy searches and `git diff --check` pass; checkpoint Markdown `.cartulary/test-results/i3-06-checkpoint` | The boundary failure was related and corrected in-slice; no validation was skipped. Export, portable-value decoding, import preparation, and import application now change independently; Task lifecycle restoration, Decision machine restoration, and common value decoding likewise have distinct files. No declaration, SQL text, ordering, attribution, validation, or dispatch behavior changed. Exact bundle bytes/order, invariant selection, atomic import, nullable rollback, lifecycle, same-incident references, SQL effects, and negative cases remain green. Boundary grants name only files that require each capability; no generated or routing input changed, so generation was not required. This same-package cleanup has no public, schema, data, frontend, dependency, or compatibility migration. After Markdown passes, activate `I3-07` for durable boundary and harness accounting. |
| 2026-08-26T01:00:22-04:00 | I3-07 | `I3-06` checkpoint complete against the same bound HEAD and intended cumulative worktree | `internal/modules/tasksdecisions/exported_surface_test.go`; `internal/modules/tasksdecisions/projectioncontract/contribution.go`; `contracts/projection-providers/index.json`; regenerated `tools/execution_topology_render_index.json`; this tracker. Existing `tools/backend_module_boundaries.json` and authored family selectors were reconciled and required no further edit. Rollback is the AST/absence/cohesion lock, characterization reference, manifest projection, regenerated digest, and checkpoint together. | Format 2/2 `.cartulary/test-results/20260826T045324Z-p853327`; focused negative/export/cohesion/absence guard row 1/1 `.cartulary/test-results/20260826T045335Z-p857364`; initial generation passed `.cartulary/test-results/20260826T045345Z-p857973`, after which review identified that the source descriptor and authored provider index are separate projections; updating the authored index and rerunning generation passed `.cartulary/test-results/20260826T045451Z-p861734`; Tasks/Decisions 20/20 `.cartulary/test-results/20260826T045505Z-p864678` and service 15/15 `.cartulary/test-results/20260826T045558Z-p905697`; Projections 15/15 `.cartulary/test-results/20260826T045649Z-p946022` and service 11/11 `.cartulary/test-results/20260826T045737Z-p963093`; boundary 3/3 `.cartulary/test-results/20260826T045819Z-p979720`; generation drift 4/4 `.cartulary/test-results/20260826T045826Z-p980085`; generated policy 3/3 `.cartulary/test-results/20260826T045838Z-p983032`; JSON shape 3/3 `.cartulary/test-results/20260826T045842Z-p983466`; harness 2/2 `.cartulary/test-results/20260826T045849Z-p983968`; fast 432/432 `.cartulary/test-results/20260826T045907Z-p984626`; owner-qualified retired path/symbol/import searches and `git diff --check` pass; checkpoint Markdown `.cartulary/test-results/i3-07-checkpoint` | No Make target failed or was skipped. Two preliminary raw symbol searches exited nonzero because their scope included valid same-named APIs owned by Indicators, Parties, and Artifacts; the corrected owner-qualified search is empty. The routed AST lock now distinguishes constants, variables, types, functions, concrete qualified methods, interface methods, and interface embeddings across the owner root, `projectioncontract`, and `projectionports`; its synthetic unexpected-method fixture proves method drift is rejected. Exact declaration-file inventories protect the portability and rollback decomposition, and repository AST/path guards reject every retired Tasks/Decisions API, task-specific rebuild path, external-test helper export, old import, directory, and catch-all file. The provider index references the real supersession characterization file; boundary policy, authored selectors, generated topology, and harness agree. This verification change has no public, schema, data, frontend, dependency, or compatibility migration. After Markdown passes, activate `I3-08` for final validation and handoff completion. |
| 2026-08-26T01:39:16-04:00 | I3-08 | `I3-07` checkpoint complete against the same bound HEAD and intended cumulative worktree | Final validation changed only this tracker. The intended cumulative worktree is the adopted Projections boundary/Appendix amendment; the directional Tasks/Decisions packages and all atomic callers; root facade, mutation, test, portability, and rollback cleanup; boundary/provider/harness metadata; generated topology; and their tests. Rollback for this slice is the final evidence/status row only. | Tasks/Decisions 20/20 `.cartulary/test-results/20260826T050148Z-p991689` and service 15/15 `.cartulary/test-results/20260826T050234Z-p1032013`; Projections 15/15 `.cartulary/test-results/20260826T050328Z-p1072325` and service 11/11 `.cartulary/test-results/20260826T050405Z-p1088903`; Workbook 66/66 `.cartulary/test-results/20260826T050442Z-p1105470` and service 37/37 `.cartulary/test-results/20260826T050654Z-p1163333`; Imports 23/23 `.cartulary/test-results/20260826T050907Z-p1221155` and service 14/14 `.cartulary/test-results/20260826T051016Z-p1263640`; Reporting 5/5 `.cartulary/test-results/20260826T051125Z-p1306065` and service 4/4 `.cartulary/test-results/20260826T051206Z-p1322405`; Recovery 24/24 `.cartulary/test-results/20260826T051245Z-p1338672` and service 19/19 `.cartulary/test-results/20260826T051403Z-p1391567`; Incident Bundles 8/8 `.cartulary/test-results/20260826T051518Z-p1444189` and service 6/6 `.cartulary/test-results/20260826T051614Z-p1460612`; Revisions 27/27 `.cartulary/test-results/20260826T051710Z-p1477050` and service 20/20 `.cartulary/test-results/20260826T051816Z-p1522238`; Collaboration 32/32 `.cartulary/test-results/20260826T051918Z-p1567227` and service 23/23 `.cartulary/test-results/20260826T052045Z-p1615528`; boundary 3/3 `.cartulary/test-results/20260826T052216Z-p1663598`; drift 4/4 `.cartulary/test-results/20260826T052218Z-p1663944`; generated policy 3/3 `.cartulary/test-results/20260826T052227Z-p1666865`; JSON shape 3/3 `.cartulary/test-results/20260826T052228Z-p1667274`; harness 2/2 `.cartulary/test-results/20260826T052231Z-p1667768`; agent finalization 1/1 `.cartulary/test-results/20260826T052249Z-p1668339`; OpenAPI 4/4 `.cartulary/test-results/20260826T052306Z-p1671191`; browser 58/58 `.cartulary/test-results/20260826T052315Z-p1671871`; fast 432/432 `.cartulary/test-results/20260826T052725Z-p1725806`; vulnerability 4/4 `.cartulary/test-results/20260826T052741Z-p1726608`; targeted gosec 4/4 `.cartulary/test-results/20260826T052749Z-p1727437`; build 7/7 `.cartulary/test-results/20260826T052802Z-p1757353`; initial full check 657/658 `.cartulary/test-results/20260826T052824Z-p1791296`; exact unrelated Jobs telemetry rerun 3/3 `.cartulary/test-results/20260826T053335Z-p1910908`; final full check 658/658 `.cartulary/test-results/20260826T053422Z-p1927308`; independent repository-wide retired-surface audit passes; checkpoint Markdown `.cartulary/test-results/i3-08-checkpoint` | The initial full check's sole failure was the untouched `platform.jobs.integration.operational_telemetry` expiry-counter timing assertion; its exact service-backed row and the complete graph passed unchanged on rerun. `agent-finalize` intentionally received no `RESULTS_DIR` because the ladder places it before the final full check; retained-run selection, closure, and timing maintenance record `results-dir-not-provided`, while all non-retained finalization work passed. No affected validation, security gate, or required command was skipped. The final intended-worktree review preserves the user-owned staged tracker draft and contains no unrelated change; `docs/domain.md`, public contracts, authorization, lifecycle, transaction/data/schema behavior, dependencies, and frontend sources remain unchanged. No compatibility shim, alias, forwarding package, fallback, dual path, database reset, commit, deployment, or migration was introduced. Iteration 3 is closed. |

### Slice verification posture

`I3-00` MUST begin with:

1. `make task-guide ROLE=module-author OWNER=module.tasksdecisions`
2. `make explain-test-owner OWNER=module.tasksdecisions`
3. `make test-slice OWNER=module.tasksdecisions`
4. `make service-backed-test-slice OWNER=module.tasksdecisions`
5. task-guide-directed affected-owner baselines for Projections, Workbook,
   Imports, Reporting, Recovery, Incident Bundles, Revisions, and
   Collaboration
6. `make backend-module-boundary-check`
7. `make check`

Each Go implementation slice MUST run `make format`, its narrow Tasks/Decisions
owner rows, and every directly affected owner row before checkpointing. A slice
that changes authored generation or test-routing inputs MUST also run
`make generate` and the corresponding drift, policy, shape, and harness gates
before it can be `DONE`.

### Final validation ladder

`I3-08` MUST run the following from the repository root, narrowest first:

1. `make test-slice OWNER=module.tasksdecisions`
2. `make service-backed-test-slice OWNER=module.tasksdecisions`
3. task-guide-directed unit and service-backed rows for
   `module.projections`, `module.workbook`, `module.imports`,
   `module.reporting`, `module.recovery`, `module.incidentbundles`,
   `module.revisions`, and `module.collaboration`
4. `make backend-module-boundary-check`
5. `make generate-drift`
6. `make generated-artifact-policy-check`
7. `make json-shape-check`
8. `make harness-contract`
9. `make agent-finalize`, recording whether retained-run maintenance used a
   supplied `RESULTS_DIR`
10. `make openapi-compatibility-check`
11. `make browser-e2e-webserver-backed`
12. `make test-fast`
13. `make go-vulncheck`
14. `make go-gosec-targeted`
15. `make build`
16. `make check`
17. repository-wide retired-symbol, retired-file, and retired-import searches
18. `make lint-markdown`
19. `git diff --check`
20. final intended-worktree review

An affected failure, unexplained skip, stale generated artifact, unresolved
security finding, or compatibility shim prevents `I3-08` completion.

### Iteration 3 acceptance

| Acceptance ID | Binary criterion | Required evidence | Status |
| --- | --- | --- | --- |
| I3-AC-001 | The old projection package is absent; the source contribution and runtime consumer directions are represented by separate exact contracts. | Package/import absence, AST surface locks, Projections and affected-owner tests | PASS |
| I3-AC-002 | Unused task-specific rebuild interfaces and methods are absent while generic catalog-driven incident and restore rebuilding remains green. | Retired-symbol search, provider registry, incident rebuild, and restore tests | PASS |
| I3-AC-003 | Combined Incident Bundle, concrete Reporting, legacy Recovery naming, and import alias surfaces are absent without shims. | Root export lock, application composition, affected owner slices | PASS |
| I3-AC-004 | Member-reference validation is owner-private and direct, and supersession uses the common row-version conflict type with unchanged failure mapping. | Owner/service, Workbook, Imports, conflict, supersession, and rollback evidence | PASS |
| I3-AC-005 | Test helpers and behavior suites are private and cohesive; every preserved or changed selector resolves through authored routing. | AST/file assertions, owner slices, harness contract | PASS |
| I3-AC-006 | Portability and rollback catch-all files are absent, with exact bundle, invariant, attribution, SQL, reference, lifecycle, and rollback behavior. | Incident Bundle, Revisions, owner, and service-backed evidence | PASS |
| I3-AC-007 | Boundary policy, provider metadata, authored routing, and generated topology agree with the final package graph. | Generation, drift, policy, JSON-shape, boundary, and harness roots | PASS |
| I3-AC-008 | Public and persisted behavior remains compatible and no shim, schema, data, dependency, frontend, domain, or specification migration exists. | OpenAPI, browser, security, build, full check, searches, and Git review | PASS |
| I3-AC-009 | Every workstream has a complete checkpoint and the final worktree contains only intended changes. | Section 14 checkpoint log and final handoff | PASS |

### Completion and handoff posture

Iteration 3 is complete: every `I3-*` row is `DONE`, every `I3-AC-*` row is
`PASS`, all final gates are green, and this tracker records exact retained
roots and migration posture. The final handoff is closed through `I3-08`.
