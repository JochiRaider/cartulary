# evidence Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

### 1.1 Target and change boundary

| Item | Normative planning value |
| --- | --- |
| Target path | `internal/modules/evidence` |
| Target label | `evidence`, derived from the final path segment and normalized to lowercase kebab case |
| Output path | `docs/handoffs/evidence-module-refactor-tracker.md` |
| Status | Iteration 1 S00-S10, Iteration 2 S11-S17, and Iteration 3 S18-S24 complete |
| Gap closure | Iteration 1 G01-G11, Iteration 2 G12-G18, and Iteration 3 G19-G24 closed |
| Implementation sequence | S00-S17 are completed history; S18-S24 are the controlling serial Iteration 3 sequence |
| Compatibility posture | Preserve normative behavior and data; remove obsolete internal Go surfaces, mixed-direction contracts, provider reach-through, and test compatibility scaffolding without shims |
| Permitted writes | S18 changes only this tracker; later explicitly authorized slices may change internal implementation, tests, authored verification inputs, generated outputs through Make, and required callers |
| Prohibited writes | Hand edits to generated roots, dependency lock files, and test/runtime dependencies on Markdown |

`MUST`, `MUST NOT`, `SHALL`, and `REQUIRED` in this tracker govern only the
refactor workflow, evidence gates, slice ordering, and handoff discipline. They
do not create product behavior. Adopted specifications remain the sole product
authority within their scopes. If this tracker conflicts with an adopted owner,
the owner controls and the affected work MUST record
`BLOCKED: owner contradiction`.

Iteration 1 implementation was explicitly authorized on 2026-08-11 and is
complete. Iteration 2 implementation was explicitly authorized on 2026-08-12;
S11 through S17 are complete. Each slice updated this tracker and passed
`make lint-markdown` before its successor began.
Iteration 3 planning was explicitly authorized on 2026-08-28. S18 was limited
to this tracker. Implementation was explicitly authorized on 2026-08-28;
S19 through S24 are complete and passed every serial tracker gate.
The Iteration 3 sequence applies the user's structural
cleanup principles: prefer durable boundaries, remove unnecessary internal
compatibility, and retain only behavior that improves the subsystem's future
state.
A slice that cannot satisfy its exit criteria MUST be marked `BLOCKED`; later
slices MUST NOT begin. Tests, generators, conformance, and runtime code MUST
NOT read this tracker or any other Markdown.

Sections 2 through 12 preserve the completed Iteration 1 inventory, decisions,
execution evidence, and handoff. Sections 13 through 22 preserve the completed
Iteration 2 production-readiness plan and ledger. Sections 23 through 32 are
the controlling Iteration 3 plan. A historical finding is not open work unless
the Iteration 3 delta inventory or gap matrix explicitly carries it forward.

### 1.2 Source hierarchy and evidence inspected

The workflow MUST apply this authority order:

1. Adopted subsystem NLSpecs for their named scopes.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. Domain vocabulary and implementation-support guides.
5. Current repository code and tests as implementation evidence.
6. Prior plans, handoffs, frameworks, and research as evidence only.

The initial session read the planning framework first. This revision preserved
that doctrine and used `temp/analysis-notes.md` and
`docs/research/nlspec-spec.md` only as non-authoritative inputs. Owner documents
inspected include Core 00 through Core 04, the Reporting and Testing Harness
NLSpecs, and `docs/domain.md`. Core 05 is inapplicable because no conformance
claim is published here.

Repository evidence inspected includes all 70 Iteration 1 target files in Section 2 and
the exact live Evidence routes, admission, tokens, handles, former Store and
current cohesive services, runtime,
Workbook/projection contributions, Projections adapters/storage, Collaboration
intents, Reporting provider, Imports, Incident Bundles, Recovery, assembly,
migrations, view-schema/OpenAPI owners, verification routing, test support, and
frontend consumers needed for the findings. Exact files were opened before a
search result was used as evidence.

### 1.3 Document-placement matrix

| Artifact | Authority class | Required content | Excluded content |
| --- | --- | --- | --- |
| Core 04 §2.0A | Product authority | Exact Evidence role matrix, rechecks, failure precedence, concealment, closed-incident posture, and issuance-to-use consequences | Go interfaces, scheduler internals, harness IDs |
| Core 01 Evidence route/handle sections | Product authority | Routes, envelopes, upload/access bindings, lifetimes, token states, consumption, error vocabularies | Duplicate role matrix |
| Reporting NLSpec §7.1.1 | Subsystem authority | Exact provider allowlist, forbidden values, default omission, eligibility, logical support identity | Storage design or route authorization |
| Core 02 and Core 03 | Existing product authority | Lifecycles, cleanup deadlines, retention, and durable coordination safety invariants | Application assembly, telemetry registry details, or a public Jobs surface |
| This tracker | Refactor guidance | Interfaces, dependencies, private worker design, characterization, rollback, acceptance gates | Product requirements |
| Testing Harness inputs | Verification accounting | Semantic identities, selectors, topology, retained evidence | Runtime architecture |
| Analysis notes and NLSpec research guide | Non-normative research | Recommendations and style input | Adopted behavior |
| Generated contracts/topology | Machine projection | Tool-produced outputs | Hand-authored authority or manual edits |

The framework/live mismatch remains a planning finding: the live module has
typed Workbook, Projections, Timeline, Revisions, Imports, Recovery, Reporting,
and Incident Bundle contributions beyond the framework's generic catalog.
Those seams MUST be judged by semantic ownership, not directory depth.

## 2. Current-State Repository Inventory

At the Iteration 1 handoff, the target contained 70 files: 48 production Go
files, 21 Go test files, and one inert placeholder. That historical inventory
is preserved below; the S16 completion record owns the final Iteration 2 layout.
Generated contracts are read-only dependencies or downstream outputs; none is
to be hand-edited.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/evidence/.gitkeep` | Historical directory placeholder | None | None | None | None | None | None | Low | Explicitly out of behavioral scope because it has no content or runtime effect. |
| `internal/modules/evidence/access_repository.go` | Reads access snapshots, persists opaque handles, and atomically redeems/revalidates them | Package-private `accessHandleRepository` | `AccessHandleService` only | PostgreSQL, evidence/blob/access-handle tables | `handles_test.go`, `evidence_integration_test.go`, `integration_test.go` | Core 01 handle contract; Evidence OpenAPI | Evidence persistence | Critical | Authorization-sensitive state is rechecked at redemption. |
| `internal/modules/evidence/api.go` | Strictly decodes blob, attach, and handle requests; computes idempotency hashes and response helpers | `BlobCreateRequest`, `AttachBlobRequest`, decode/hash functions | Evidence route handlers and tests | `httpapi`, JSON, hashing, media/filename helpers | `blob_create_test.go`, `attach_test.go`, `handles_test.go`, conformance/integration tests | Evidence OpenAPI and error registry | Evidence transport adapter | High | Public Go names are internal-repository API, not external wire authority. |
| `internal/modules/evidence/attach_test.go` | Unit and service-backed characterization for attach validation, concealment, uniqueness, and races | Test-only `BlobOptions`, `DurableCounts`, helpers, three tests | Go test and `module.evidence` harness rows | PostgreSQL test support, blob-lifecycle service | Self | Evidence view/lifecycle and uniqueness contracts | Evidence tests | Medium | Test-only exports are not production package design evidence. |
| `internal/modules/evidence/blob_create_test.go` | Characterizes object-blob create validation, idempotency, size ceilings, and preview payload limits | Four tests | Go test and Evidence harness rows | Route operations and blob-lifecycle test capabilities | Self | OpenAPI request/response and configured limit contracts | Evidence tests | Medium | Freezes rejection precedence and replay behavior. |
| `internal/modules/evidence/blobref/blobref.go` | Canonical physical object key and logical `object://` reference formatting/parsing | `StorageKeyParts` and key/ref parse, build, and validation functions | Evidence persistence, upload, portability, tests, object-store probe | UUID and string validation | `blobref_test.go`, portability and object-store tests | Storage-reference and non-leakage contracts | Evidence | High | Physical key and logical reference must remain distinct and never leak publicly. |
| `internal/modules/evidence/blobref/blobref_test.go` | Exact key/ref round-trip and rejection characterization | Two tests | Go test and Evidence unit row | `blobref` | Self | Core 01 storage-reference rules | Evidence tests | Low | Protects canonical `incidents/.../object-blobs/...` and `object://...` forms. |
| `internal/modules/evidence/cleanup.go` | Implements the durable cleanup claim engine, bounded sweeps, deletion deadlines, retries, state rechecks, metadata retention, and closed health snapshot | `CleanupObjectDeleter`, `CleanupSweepResult`, and `CleanupService.SweepFailedUnattachedBlobs` | Private Evidence cleanup dispatcher | PostgreSQL, typed object-store deletion, UUID/time | Cleanup integration and dispatcher tests | Core 02 cleanup policy, migration 31, Recovery transient-state contract | Evidence cleanup capability | Critical | The dispatcher activates this engine only after publication readiness. |
| `internal/modules/evidence/cleanup_dispatcher.go` | Owns the private periodic cleanup lifecycle and closed sweep observations | `CleanupDispatcherIdentity`, `CleanupSweeper`, `CleanupObserver`, `CleanupDispatcher` | Server application facade | Cleanup service, typed deleter, clock, timer, observer | `cleanup_dispatcher_test.go`, cleanup integration, server activation tests | Core 02 dispatcher identity/cadence and OpenTelemetry owner vocabulary | Evidence cleanup capability | Critical | Runs immediately after explicit start, then every 15 minutes, serializes sweeps, and stops cleanly. |
| `internal/modules/evidence/cleanup_dispatcher_test.go` | Proves dispatcher identity, cadence, delayed activation, cancellation, serialization, and shutdown | Two unit tests and private fakes | Evidence unit harness row | Cleanup dispatcher | Self | Authored Evidence cleanup activation row | Evidence tests | High | Uses a short private interval while asserting the production interval is exactly 15 minutes. |
| `internal/modules/evidence/cleanup_integration_test.go` | Service-backed safety matrix for durable cleanup coordination and multi-instance dispatch | One top-level integration test and private fixtures | Evidence service-backed harness row | PostgreSQL test support, fake typed deleter, cleanup service and dispatcher | Self | Authored Evidence test-family row and S08/S09 exit contracts | Evidence tests | Critical | Covers inert deployment, concurrency, restart, batching, retry, deadline, races, retention, and two active dispatchers; the full matrix passes in §§6.8-6.9. |
| `internal/modules/evidence/cleanup_objectstore.go` | Adapts cleanup to the typed object-store delete contract with one closed purpose | `NewCleanupObjectDeleter` | Server cleanup composition | `objectstore.Store`, typed delete capability, `PurposeEvidenceCleanup` | `cleanup_objectstore_test.go`, platform object-store tests | Core 02 typed cleanup purpose | Evidence storage adapter | Critical | Rejects legacy/untyped stores and preserves typed not-found classification. |
| `internal/modules/evidence/cleanup_objectstore_test.go` | Proves exact typed purpose, typed error preservation, and rejection of legacy-only stores | Three unit tests | Evidence unit harness row | Cleanup object-store adapter | Self | Authored Evidence cleanup activation row | Evidence tests | Medium | Prevents fallback to provider strings or ungoverned deletion. |
| `internal/modules/evidence/collaboration_intents.go` | Converts primary and affected-row changes into ordered Collaboration intents | Package-private helper | Mutation coordinator and blob-lifecycle mutation paths | Collaboration intent contract | Coordinator and integration/WebSocket tests | `record_changed` protocol semantics | Evidence producer; Collaboration transports | High | Evidence produces facts; it does not own the WebSocket hub. |
| `internal/modules/evidence/conflict_snapshot.go` | Maps Evidence view fields to revision conflict snapshot source keys | Package-private mapping helpers | Workbook conflict resolution | Revisions conflict/source contracts | Workbook conflict and conformance coverage | Evidence view schema and revision snapshot contract | Evidence | High | Field-key closure must not drift from the authored view schema. |
| `internal/modules/evidence/conformance_test.go` | Checks reason-code registry and Evidence public field-key closure | `ErrorRegistry`, helper functions, two tests | Go test and Evidence unit rows | Authored error registry and view schema loaders | Self | `contracts/errors/index.json`, Evidence view schema | Evidence tests | Medium | Reads authored contracts; generated output remains tool-owned. |
| `internal/modules/evidence/create_validation.go` | Enforces minimum create signal, reserved storage reference policy, lifecycle initialization, and direct patch admission | `ValidateWorkbookCreateParams`, `ValidateWorkbookDirectPatchChange` | Workbook/import mutation paths | Owner-local mutation types and blobref policy | Characterization, lifecycle, integration tests | Evidence view/create contract | Evidence | Critical | Authoritative Evidence mutation policy. |
| `internal/modules/evidence/create_validation_characterization_test.go` | Matrix for adopted create-signal and lifecycle rules | One test | Go test and Evidence unit row | Create validator | Self | Core 01/02/03 create contract | Evidence tests | Low | Baseline required before reshaping mutation code. |
| `internal/modules/evidence/deleterestore/provider.go` | Supplies Evidence source snapshot and restoration view resolution to Revisions delete/restore | `Source`, `NewSource`, provider methods | `RevisionProviderContribution` | Revisions delete/restore contracts | Provider contract and cross-owner Revisions tests | Revision provider/snapshot contracts | Evidence contribution; Revisions coordinates | Medium | Thin source-owner provider is appropriate. |
| `internal/modules/evidence/evidence_blob_uniqueness_migration_test.go` | Verifies the head migration contains the unique live blob association constraint | One integration test | Evidence service-backed row | PostgreSQL catalog | Self | Migration `00017_evidence.sql` | Evidence tests | Medium | Migration is authored source and must not be edited for a structural refactor. |
| `internal/modules/evidence/evidence_integration_test.go` | HTTP/service-backed tests for blob creation, handles, reason codes, strict members, and OpenAPI route presence | Seven tests | Evidence service-backed harness rows | Server harness, object store, database, public client | Self | Evidence OpenAPI, errors, generated protocol behavior | Evidence tests | High | Primary wire/security characterization. |
| `internal/modules/evidence/handles_test.go` | Unit characterization for issue idempotency, current-state recheck, disposition, and filename sanitation | Four tests | Evidence unit rows | Access repository/service test doubles | Self | Handle OpenAPI and security contract | Evidence tests | Medium | Includes single-use download and reusable preview expectations. |
| `internal/modules/evidence/import_create.go` | Adapts enabled Evidence imports to the owner-private source-mutation kernel in a caller-owned transaction | `ImportCreateCommand` alias and owner-private facade construction | Evidence owner runtime; narrow facade consumed by Imports | Imports owner facade and the Evidence mutation kernel | Import assembly and parity integration tests | `contracts/imports/view-targets.v1.json` | Evidence contribution; Imports coordinates | High | Imports receives only its published narrow owner facade and cannot reconstruct Evidence internals. |
| `internal/modules/evidence/import_projection.go` | Refreshes and reads Evidence projection results during import creation | Package-private helper | Evidence import facade | Evidence projection rows | Import tests | Evidence view/projection contract | Evidence contribution | High | Must preserve caller transaction and projection result semantics. |
| `internal/modules/evidence/incident_bundle_blob_portability.go` | Exports object bytes, rewrites/stages imported keys, and cleans up staged bytes on failure | `IncidentBundleBlobPortability` and methods | Evidence incident-bundle source port and portability assembly | Object store, blobref, incident portability contracts | Cross-owner portability tests | Incident-bundle source catalog | Evidence contribution; Incident Bundles coordinates | Critical | Cross-resource rollback and physical-key concealment are security/data-integrity sensitive. |
| `internal/modules/evidence/incident_bundle_portability.go` | Exports/imports Evidence, blob, access, and custody relations as deterministic NDJSON files | `ExportIncidentBundleFiles`, `ImportIncidentBundleFilesTx` | Evidence source port | PostgreSQL, incident portability types | Provider contract and cross-owner round-trip tests | Incident-bundle source catalog | Evidence contribution | Critical | Exact relation/path/version/attribution behavior must be frozen. |
| `internal/modules/evidence/incident_bundle_source_port.go` | Declares the Evidence incident-bundle source descriptor and callbacks | `NewIncidentBundleSourcePort` | Incident portability assembly | Incident Bundles source-port contract | Provider contract and cross-owner catalog tests | `contracts/incident-bundles/source_catalog.json` | Evidence contribution | Medium | Generic bundle coordination remains outside Evidence. |
| `internal/modules/evidence/incident_bundle_subtype_presence.go` | Contributes Evidence subtype-presence validation for Records portability | `IncidentBundleSubtypeContribution` | Incident portability assembly | Records subtype-presence contract | Provider contract and portability tests | Incident-bundle/Records invariants | Evidence contribution | Medium | Evidence owns subtype fact; Incident Bundles coordinates. |
| `internal/modules/evidence/initial_blob_create.go` | Validates, observes, finalizes, and associates an optional initial blob during atomic generic row creation | Package-private helpers | Workbook create facade | PostgreSQL, object store, blob slots, lifecycle policy | Atomic create, lifecycle, concealment tests | Generic Workbook create and Evidence lifecycle contracts | Evidence application service | Critical | Preserves one transaction and unique association while object observation is external. |
| `internal/modules/evidence/integration_test.go` | End-to-end module characterization for upload/attach, generic create, projection rebuild, WebSocket refresh, expiry, handle invalidation, and quarantine | Nine tests plus test-only event/count helpers | Evidence integration/service-backed rows | Full app support, object store, PostgreSQL, WebSocket client | Self | HTTP, Workbook, view, projection, collaboration, lifecycle contracts | Evidence tests | Critical | Contains current evidence for cross-owner atomic effects. |
| `internal/modules/evidence/internal/policy/failure.go` | Defines the owner-private finalize threshold and deterministic failure/cleanup schedule | `CleanupDelay`, finalize thresholds, `ScheduleFailure`, `FinalizeAttemptIsTerminal` | Blob lifecycle, pending timeout, cleanup candidates | Time only | Shared policy and service-backed failure tests | Core 02/03 failure and cleanup deadlines | Evidence policy | Critical | Fixes cleanup eligibility at failure plus 45 minutes. |
| `internal/modules/evidence/internal/policy/lifecycle.go` | Defines closed Evidence/blob vocabularies, transitions, lifecycle bridges, association disposition, and closed-incident mutation posture | Typed constants and owner-private policy functions | Ordinary mutation, import kernel, attach, finalization, quarantine, rollback | Time only | Shared policy, lifecycle, attach, rollback, import, integration tests | Core 02 lifecycle machines and Core 04 incident-state rules | Evidence policy | Critical | A linked available blob does not implicitly promote Evidence lifecycle. |
| `internal/modules/evidence/internal/policy/policy_test.go` | Exhaustively exercises lifecycle, association, incident, reference, and failure policy matrices | Test-only table suites | Evidence unit row | Owner-private policy | Self | Core 01/02/03 lifecycle/reference/timing rules | Evidence tests | Low | Central executable oracle for every policy caller. |
| `internal/modules/evidence/internal/policy/references.go` | Owns server-managed reference admission and persisted physical-key identity validation | Typed reason constants/errors and validation functions | Create/patch, rollback, routes, route dependencies | `blobref`, UUID | Shared policy, rollback, route/object-store tests | Core 01 storage-reference and concealment rules | Evidence policy | Critical | `blobref` remains syntax-only; association meaning stays here. |
| `internal/modules/evidence/lifecycle_test.go` | Proves Evidence-record and Object Blob lifecycles remain distinct | One test and test-only `Queryer` | Evidence unit row | Blob-lifecycle service, PostgreSQL stubs | Self | Core 02 lifecycle machines | Evidence tests | Medium | Essential boundary characterization. |
| `internal/modules/evidence/mutation_coordinator.go` | Implements the private source-mutation kernel that orders admission, source write, projection refresh, revisions, and Collaboration intents inside the caller transaction | Package-private kernel/command/result | Workbook and import mutation paths through `SourceMutationService` | Narrow ports from `mutation_ports.go` | Kernel-order and ordinary/import parity tests | Workbook, projection, revision, Collaboration contracts | Evidence application service | Critical | Caller owns transaction/idempotency while owner-controlled create behavior is shared. |
| `internal/modules/evidence/mutation_coordinator_test.go` | Fault-boundary test for exact mutation side-effect order | One test | Evidence unit row | Fake coordinator ports | Self | Mutation atomicity/effect-order contract | Evidence tests | Low | Primary non-database sequencing characterization. |
| `internal/modules/evidence/mutation_ports.go` | Defines narrow incident, Records, source, projection, revision, and collaboration mutation capabilities | Package-private interfaces/types | Mutation coordinator and composition | Peer-owner published types | Coordinator/composition/integration tests | No generated output directly | Evidence boundary layer | High | Existing boundary seam; must not become a generic transaction facade. |
| `internal/modules/evidence/objectstore_dependency_test.go` | Verifies public Evidence responses do not expose storage keys, refs, or object-store details | Test-only object-store admin plus one test | Evidence security/integration row | Server harness, object store, public clients | Self | OpenAPI public response and Core 04 concealment rules | Evidence tests | High | Security regression evidence, not production API. |
| `internal/modules/evidence/persistence_components.go` | Componentized blob slot/read/lifecycle, association, and cleanup-candidate SQL | Package-private repository components | Blob-lifecycle, source-mutation, and cleanup capabilities | PostgreSQL Evidence tables | Blob, lifecycle, and integration tests | Evidence schema/lifecycle contracts | Evidence persistence | Critical | Legitimate owner-local SQL is now injected only into the capability that uses it. |
| `internal/modules/evidence/projectionprovider/contribution.go` | Constructs the Evidence projection-provider contribution | `NewContribution` | Projection assembly | `workbookprojection` contribution | Projection contribution and assembly tests | `contracts/projection-providers/index.json` | Evidence contribution; Projections coordinates | Medium | Source owner constructs provider; Projections validates catalog and stores rows. |
| `internal/modules/evidence/projectionprovider/source.go` | Reads authoritative Evidence, blob, link, and record facts into typed projection input | `Source`, `NewSource`, reader methods | Evidence projection contribution | PostgreSQL and typed Evidence projection input | Projection and integration tests | Evidence view schema/projection provider contract | Evidence source reader | High | It must not own physical projection storage. |
| `internal/modules/evidence/provider_contract_test.go` | Exact assertions for source-owner recovery, revision, portability, subtype, and service-construction contracts | Two tests | Evidence unit row | Evidence contribution and cohesive-service constructors | Self | Recovery, revision, incident-bundle catalogs; composition boundary | Evidence tests | Medium | Missing dependency sets fail construction; the exact contribution-closure suite passed in S07. |
| `internal/modules/evidence/recovery_inventory.go` | Builds Evidence vNext object inventory entries for recovery validation | `VNextRecoveryObjectInventory` | Recovery/operator assembly and tooling | Object store, blobref, recovery contracts | Cross-owner recovery tests | Recovery object-inventory contracts | Evidence contribution; Recovery coordinates | High | Physical object existence/integrity is security and recoverability sensitive. |
| `internal/modules/evidence/recovery_state.go` | Declares Evidence authoritative tables for state recovery | `RecoveryStateContribution` | Recovery assembly | Recovery-state contribution contract | Provider contract and Recovery tests | Recovery state catalog/fixtures | Evidence contribution | Medium | Exact relation set must remain aligned with portability. |
| `internal/modules/evidence/recoveryprovider/provider.go` | Restores Evidence object bytes/state through a Recovery provider | `Provider`, `New`, provider methods | Operator recovery composition | PostgreSQL, object store, blobref, recovery provider contracts | Cross-owner Recovery/operator tests | Recovery verification contracts | Evidence contribution; Recovery/operator coordinates | Critical | Must preserve validation, cleanup, and no-raw-key behavior. |
| `internal/modules/evidence/reportingprovider/provider.go` | Supplies closed Evidence field facts and logical Evidence support targets for Reporting export | `CollectFieldsTx`, `CollectFactsTx`, `CollectLogicalSupportTargetsTx` | Reporting export materializer | PostgreSQL, Reporting provider contract | Exact Reporting provider and export-model tests | Reporting provider/content-class/support-ref contracts | Evidence contribution; Reporting coordinates | High | Constructs only the adopted twelve members; unknown columns and physical identifiers are omitted by default. |
| `internal/modules/evidence/revision_append_port.go` | Adapts Evidence mutation facts to the Revisions appender | Package-private adapter | Source and blob-lifecycle mutations | Revisions appender and snapshot types | Coordinator/integration tests | Revision/change-set contracts | Evidence contribution | High | Must retain caller transaction and exact source snapshot semantics. |
| `internal/modules/evidence/revision_provider_contribution.go` | Registers Evidence route, snapshot, delete/restore, and rollback capabilities with Revisions | `RevisionProviderContribution` | Revision assembly | Revisions provider contracts and Evidence providers | Provider contract and cross-owner Revisions tests | Revision provider catalog/snapshot contract | Evidence contribution; Revisions coordinates | High | Contribution identity and route coverage are stable internal contracts. |
| `internal/modules/evidence/rollbackprovider/provider.go` | Validates snapshot values and restores Evidence source rows during rollback | `Provider`, `NewProvider`, methods | Evidence revision contribution | PostgreSQL, Revisions rollback/source contracts | Local provider and Revisions rollback tests | Evidence revision snapshot contract | Evidence contribution | High | Lifecycle and reference validation duplicates some root policy. |
| `internal/modules/evidence/rollbackprovider/provider_test.go` | Unit tests snapshot presence/association and invalid lifecycle/reference rejection | Two tests | Evidence unit row | Rollback provider | Self | Revision snapshot and lifecycle contract | Evidence tests | Low | Characterizes fail-closed rollback admission. |
| `internal/modules/evidence/route_admission.go` | Performs route authorization, membership/role, CSRF, and incident admission | Package-private service helpers | Evidence handlers | `authn`, incident access repository, `httpapi` | HTTP, handle, object-store tests | Core 04 authorization and OpenAPI errors | Evidence transport edge | Critical | Admission is correctly edge-local; deployment admin alone is insufficient. |
| `internal/modules/evidence/route_dependencies.go` | Adapts route service dependencies, including object-store upload/read behavior | Package-private adapters | `Service` construction and routes | Object store, settings, narrow route capabilities | Blob/attach/handle/integration tests | Upload-target and payload-limit contracts | Evidence transport/platform adapter | High | Platform dependency is appropriate here, not in domain validation. |
| `internal/modules/evidence/route_errors.go` | Translates internal Evidence failures into stable public error/reason envelopes | Package-private mapping helpers | Evidence handlers | `httpapi`, owner errors | Conformance, HTTP, handle tests | `contracts/errors/index.json`, OpenAPI responses | Evidence transport adapter | Critical | Preserve concealment and precedence exactly. |
| `internal/modules/evidence/routes.go` | Registers and serves six Evidence HTTP operations | `Service`, `Settings`, `RouteService`, `WithRouteService`, `RegisterRoutes`, key-error classifier | Server assembly and route tests | Route admission/dependencies, `httpapi`, object store | Blob, attach, handle, integration, OpenAPI tests | Evidence OpenAPI owner and generated protocol surfaces | Evidence transport adapter | Critical | Thin over `RouteService`; boundary policy forbids private persistence and consumer orchestration here. |
| `internal/modules/evidence/runtime.go` | Validates and composes the server-only Evidence owner runtime and publishes narrow route, Workbook, Timeline, and Import capabilities | `OwnerRuntimeDependencies`, `TimelineAttachmentContribution`, `OwnerRuntime`, error-returning constructor | Server assembly and reusable test composition | Cohesive owner services, Revisions, Collaboration, Projections, object store | Server composition and provider tests | Runtime/boundary policy | Evidence application facade | High | The immutable runtime retains four private interface capabilities and no broad concrete owner root or panic-based construction. |
| `internal/modules/evidence/source_kernel.go` | Creates authoritative Records/Evidence source state, refreshes projection, and reads the result in caller transaction | Package-private kernel | Workbook/import mutation coordinator | PostgreSQL, Records/source/projection ports | Create/import/coordinator/integration tests | Records envelope and Evidence projection contracts | Evidence application/persistence boundary | Critical | Must not begin a generic transaction or orchestrate consumers. |
| `internal/modules/evidence/store.go` | Defines cohesive blob-lifecycle, access-handle, cleanup, source-mutation, and route-composition services plus shared result records | `BlobLifecycleService`, `AccessHandleService`, `CleanupService`, `RouteOperations`, dependency records, constructors, operation results | Owner runtime, Workbook/import facades, routes, focused tests | Capability-specific PostgreSQL repositories and narrow Revisions/Projections/Collaboration ports | Evidence owner/service-backed and construction tests | HTTP, lifecycle, view, projection, revision, collaboration contracts | Evidence application services | High | Filename is retained, but no `Store`, option surface, compatibility facade, or direct external test construction remains. Cleanup is reachable only through the private dispatcher. |
| `internal/modules/evidence/workbookprojection/row_test.go` | Canonical Evidence view-row mapping fixtures | Three tests | Evidence unit row | `workbookprojection.ViewRow` | Self | Evidence view schema | Evidence tests | Low | Covers rich/null fields and lifecycle/blob-state pairs at the owner boundary. |
| `internal/modules/evidence/timeline_facts.go` | Reads Evidence attachment facts for Timeline projection use | `TimelineFact`, `TimelineFactReader`, methods | Timeline assembly/collection reads | PostgreSQL, blobref | Timeline assembly/store tests | Timeline attachment presentation contract | Evidence contribution; Timeline coordinates | High | Evidence owns attachment facts; Timeline owns row composition/presentation. |
| `internal/modules/evidence/timeline_port.go` | Validates Timeline attachment references against Evidence records | Package-private implementation of published contribution | `NewTimelineAttachmentContribution`, Timeline assembly | PostgreSQL | Timeline assembly and workbook tests | Timeline attachment mutation contract | Evidence contribution | High | No Timeline implementation import; narrow source-owner port is appropriate. |
| `internal/modules/evidence/upload_token.go` | Creates, hashes, parses, and validates opaque upload capability tokens | Package-private token helpers | Blob create/upload routes and blob-lifecycle capability | Cryptography, UUID/time | Blob create and integration tests | Opaque same-origin upload-target contract | Evidence security adapter | Critical | Raw object-store credentials/keys must never replace this token. |
| `internal/modules/evidence/upload_token_test.go` | Proves version-2 capability parsing and version-1 fail-closed cutover | One test | Evidence unit row | Upload token helpers | Self | Opaque upload-target security contract | Evidence tests | Low | Added in S01; no legacy decoder is retained. |
| `internal/modules/evidence/workbook_conflict.go` | Resolves same-field conflict changes and revalidates Evidence source state | `WorkbookConflictCommand` and facade method | Workbook assembly/mutation API | Revisions conflict workflow, PostgreSQL, Evidence validation | Workbook/Revisions conflict tests | Generic conflict route and Evidence field contract | Evidence application contribution | Critical | Conflict token, row version, and source validation are observable. |
| `internal/modules/evidence/workbook_contribution.go` | Publishes the narrow owner-constructed Evidence create/patch/conflict facade to Workbook | `WorkbookContribution` interface | Evidence owner runtime; Workbook consumes contribution | Workbook facade contract only | Server/workbook composition and integration tests | Evidence view target and Workbook mutation contracts | Evidence contribution; Workbook coordinates | High | Consumers cannot construct or rewire the contribution. |
| `internal/modules/evidence/workbook_facade.go` | Owns Evidence generic create/patch transactions, idempotency, conflict, and delegation to the shared source-mutation kernel | Facade, commands, requests, results, conflict errors; owner-private error-returning construction | Evidence owner runtime and tests | PostgreSQL plus narrow owner capabilities | Validation, atomic create, conflict, parity, integration tests | Workbook routes, Evidence view schema, revisions/events | Evidence application facade | Critical | Ordinary creation and imported creation share one owner kernel; generic Workbook envelopes remain stable. |
| `internal/modules/evidence/workbook_mutations.go` | Defines writable field values, lifecycle patches, validation errors, and direct Evidence SQL mutations | Field/create/lifecycle/error types and `ValidLifecycleState` | Workbook facade, rollback/conflict validation | PostgreSQL, blobref, owner policy | Lifecycle/create/rollback/integration tests | Evidence view and lifecycle contracts | Evidence | Critical | Lifecycle vocabulary is repeated in rollback and frontend projections. |
| `internal/modules/evidence/workbookprojection/contribution.go` | Publishes typed Evidence projection input, canonical view-row mapper, rows/rebuilder ports, descriptor, and surface intent | `Rows`, `ProjectionInput`, `ViewRow`, source/rebuilder/ports/contribution types and constructors | Projection provider/assembly, Evidence facade, test support | Projection provider contract and Evidence view semantics | Canonical mapper, projection assembly/query tests | Projection-provider index and Evidence view schema | Evidence published language; Projections stores | High | Correct source/consumer boundary; all Evidence-specific row serialization now starts here. |
| `internal/modules/evidence/workbookprojection/contribution_test.go` | Validates required source, exact descriptor/intent, unknown-view rejection, and defensive copies | Four tests | Evidence unit row | Projection contribution | Self | Projection provider and view-schema contracts | Evidence tests | Low | Protects the published provider seam. |
| `internal/modules/evidence/workbookprojection/support_effects.go` | Publishes authoritative-subject input and neutral support-view effect records | `SupportProjectionEffectsTx` and closed input/result/change types | Evidence blob-lifecycle/runtime and Projections adapter | `pgx.Tx`, UUIDs, neutral patch/invalidate vocabulary | Evidence integration and Projections registry tests | Internal Evidence/Projections boundary | Evidence published language; Projections executes | High | Added in S03; contains no peer view IDs, SQL, or physical storage details. |

## 3. Module Boundary Diagnosis

The target is a legitimate Evidence source-owner module, not an accidental
catch-all. It owns Evidence record and Object Blob lifecycles, source
persistence, attachment, custody, and secure access. Its root is nevertheless a
mixed-responsibility package, mutation coordinator, view/projection
orchestration layer, transport-adjacent adapter, and persistence-adjacent
adapter. It is not a frontend controller or grid-vendor layer.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Evidence source state, lifecycle, custody, and blob association | Root persistence/mutation and migration | Evidence | Keep | Core 01/02 and source SQL | Evidence and Object Blob remain distinct state machines. |
| Blob slot, upload, finalization, quarantine capability, cleanup eligibility | Store, routes, persistence | Evidence capability; private platform dispatch | Split | Core 01-03 and live methods | S04 decomposes capability ownership; S08-S09 add and activate cleanup. |
| Six HTTP operations | Route/admission/API files | Evidence transport edge using platform adapters | Keep | Core 01 and Core 04 §2.0A | Preserve thin `RouteService`. |
| Generic row create/patch/conflict/import | Workbook/Import facades | Evidence source contribution; consumers coordinate | Split | Live registrations and caller transactions | Preserve one source behavior. |
| Canonical Evidence row interpretation | Rows port plus local mapper | Canonical Rows capability | Move | Duplicate readers and `view_row_v1` | S-00 proves equality before S-01. |
| Timeline/Host/Identity support refresh selection | Hard-coded Store SQL/keys | Projections adapter | Move | Live support helper | Evidence supplies subjects; Projections selects plans/views/SQL. |
| Collaboration intents | Evidence coordinator/helpers | Evidence produces; Collaboration publishes | Keep | Current event types | Future adapter never publishes. |
| Source-owner consumer contributions | Typed provider packages | Evidence supplies; named consumer coordinates | Keep | Catalogs and assembly | Legitimate owner seams. |
| Saved views/view schema | Generic paths | Workbook/Saved Views/view-schema owners | Defer | No target-specific store | Preserve; invent nothing. |
| Frontend/grid state | Outside target | Web/grid adapter | Defer | Consumer scan | No backend move. |
| Test-util composition | Test support | Test support owners | Keep | Harness composition | Never a production dependency. |

### 3.1 Fixed projection-effects interface

S03 introduced this internal boundary with live repository types:

```go
type SupportProjectionEffectsTx interface {
    RefreshEvidenceAssociationEffects(
        context.Context,
        pgx.Tx,
        EvidenceAssociationEffectsInput,
    ) (EvidenceAssociationEffectsResult, error)
}
```

The closed records are:

```go
type EvidenceAssociationEffectsInput struct {
    IncidentID uuid.UUID
    Subjects   []EvidenceAssociationSubject
}

type EvidenceAssociationSubject struct {
    RecordID   uuid.UUID
    RecordType string
}

type SupportChangeKind string

const (
    SupportChangePatch      SupportChangeKind = "patch"
    SupportChangeInvalidate SupportChangeKind = "invalidate"
)

type EvidenceAssociationEffectsResult struct {
    Changes []EvidenceSupportRowChange
}

type EvidenceSupportRowChange struct {
    RecordID      uuid.UUID
    RowVersion    int64
    AffectedViews []EvidenceAffectedViewChange
}

type EvidenceAffectedViewChange struct {
    ViewSchemaID     string
    ChangeKind       SupportChangeKind
    ChangedFieldKeys []string
    Patch            *EvidenceViewRowPatch
}

type EvidenceViewRowPatch struct {
    RecordID    uuid.UUID
    RowVersion  int64
    Cells       map[string]any
    GroupValues map[string]any
}
```

`SupportChangeKind` is closed to `patch` and `invalidate`. Input subjects MUST
be unique and ordered by UUID then type. Results MUST be deterministic and
ordered by record UUID, view-schema ID, and field key. `Patch` is non-null
exactly for `patch` and contains canonical cells/group values; it is null for
`invalidate`.

Evidence determines affected authoritative subjects, calls inside the existing
caller-owned transaction, fails the full mutation on error, and produces
ordered Collaboration intents from neutral results. Projections selects
provider plans/views, owns all physical projection SQL, and returns neutral
effects. The adapter MUST NOT authorize, control transactions, publish, append
revisions, mutate Evidence/history, access object storage, expose SQL/table
details, or return partial success.

Primary Evidence row loading remains separate through canonical
`workbookprojection.Rows`. S03 updated internal callers atomically and provided
no compatibility shim. S05 carried this port into the validated
`OwnerRuntimeDependencies` composition contract without changing the boundary.

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Blob creation | Core 01; Evidence edge | Route/Store/OpenAPI | Blob tests | Exact validation, replay, roles, closure, targets, envelopes | Critical | No physical identity leaks. |
| Upload capability | Core 01 binding; Core 04 auth | Upload route/token/object store | Capability tests | Exact session/actor/incident/blob/contract/method/header/size/hash/time/lease checks | Critical | Recheck at use. |
| Attachment | Core 01/02; Evidence | Record route/Store/unique relation | Attach/race tests | Canonical row, version, replay, lifecycle, revisions, projections, intents, rollback | Critical | One caller transaction. |
| Handle issue/redeem | Core 01 state; Core 04 matrix | Access repository/routes | Handle tests | Full role/state/precedence/state-change matrix | Critical | Closed incidents retain valid reads. |
| Workbook row/query/conflict | Workbook transport; Evidence source | Contribution/schema/Rows | Cross-owner tests | Exact rows/defaults/query/version/conflict/replay | Critical | No bespoke create route. |
| Saved-view/schema | Workbook and schema owners | Generic Evidence surface | Cross-owner tests | Preserve IDs, fields, queries, selectors | High | No new behavior. |
| Lifecycles/storage | Core 02/03; Evidence | Source/migration/Store | Lifecycle tests | Every legal Evidence/blob combination and null state | Critical | State machines remain separate. |
| Projection/support effects | Evidence facts; Projections storage | Rows/support refresh | Projection/WebSocket tests | Atomicity suite | Critical | No peer layout in Evidence. |
| Collaboration | Collaboration transport; Evidence producer | Intents/socket | Event tests | Exact ordered patch/invalidate, no rollback/replay duplicate | Critical | No socket ownership move. |
| Revisions/delete/restore/rollback | Revisions; Evidence contribution | Coordinator/providers | Provider/rollback tests | Exact snapshots/history/round trip/atomicity | High | Restore never resurrects handle. |
| Authorization/concealment | Core 04 §2.0A | Matrix and route edge | Partial coverage | Six-operation matrix | Critical | Core amendment adoption gate. |
| Reporting | Reporting §7.1.1 | Live provider | Reporting tests | Redaction suite | Critical | Unknown fields omit. |
| Imports | Imports; Evidence facade | Two create paths | Import tests | Create parity suite | Critical | Full effect equality. |
| Incident Bundles | Bundle owner; Evidence provider | Portability/staging | Bundle tests | Staging cleanup suite | Critical | No physical locators. |
| Recovery | Recovery; Evidence provider | State/object inventory | Recovery tests | Exact inventory closure | Critical | Claims excluded. |
| Generated surfaces | Authored owners/generators | Read-only roots | Drift tests | Only if authored inputs change | High | Never hand-edit. |
| Frontend/grid | Web/UI owners | Consumers | Unit/browser rows | Run only if observable contract affected | High | No ownership move. |
| Harness | Testing Harness | Routing inputs | Harness contract | Map semantic suites | Medium | Accounting only. |

### 4.1 Canonical Evidence row parity

S02 made `workbookprojection.ProjectionInput` the single typed database fact
model and `workbookprojection.ViewRow` the sole Evidence-specific view-row
serializer. Transactional provider loads now read the authoritative source and
invoke that mapper; attach/quarantine mutation and revision values call the
same `Rows` port. The duplicate Store SQL/mapper and hard-coded linked count
were deleted.

The retained fixtures cover complete rich and null rows, UTC timestamp and UUID
normalization, lifecycle/blob-state pairs, initial and advanced versions, link
counts zero, one, and greater than one, tombstoned-link exclusion, immediate
post-link public projection reads, blob attachment, mutation/query exact row
parity, and existing replay/rollback paths. Timeline supplies affected
Evidence identities while the Evidence-owned attachment contribution performs
sorted unique primary projection refreshes in the caller transaction. Thus a
projection failure rolls back the link mutation rather than leaving a stale
count.

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Evidence is a legitimate source boundary. | Core/source/migration | Dissolution scatters lifecycle. | `intentional/no_action` | Evidence | Preserve. |
| Store combined persistence, lifecycle, projections, revisions, events, handles. | Former root Store/runtime | Broad blast radius. | `fixed/S04` | Evidence | Cohesive capabilities and construction evidence are recorded in §6.4. |
| Local mapper duplicated canonical Rows. | Store versus Rows | Row drift. | `fixed/S02` | Evidence/Rows | Canonical mapper and immediate link-count refresh are recorded in §6.2. |
| Evidence hard-codes peer projection layouts. | Support helper | Schema drift. | `fixed/S03` | Projections | Registry-owned effects and atomicity evidence are recorded in §6.3. |
| Evidence source SQL is legitimate. | Source reads/writes | Generic move erases ownership. | `intentional/no_action` | Evidence | Retain source SQL. |
| Platform imports are edge-local. | Current imports/guard | Domain leakage if spread. | `intentional/no_action` | Edge/platform | Retain static guard. |
| Workbook/Import composition is broad. | Former facades/assembly | Partial construction. | `fixed/S05` | Evidence owner runtime/narrow ports | Validated composition and shared-kernel parity evidence are recorded in §6.5. |
| Lifecycle/reference policy repeats. | Live/rollback code | Transition drift. | `fixed/S06` | Evidence | One private owner policy and cross-entry matrix are recorded in §6.6. |
| Reporting live implementation subtracts fields. | Provider SQL | Future leak. | `fixed/S07` | Reporting/Evidence | Closed construction, unknown-field omission, and export-model proof are recorded in §6.7. |
| Cleanup formerly had no production caller. | Caller search/Core deadlines | Unwired behavior. | `fixed/S09` | Evidence/app/storage/telemetry | Readiness-gated private dispatcher and exact lifecycle evidence are recorded in §6.9. |
| Cleanup needs durable claims. | Cross-resource race | Duplicate/false completion. | `fixed/S08` | Evidence runtime state | Exact §7.3 design and service-backed matrix pass in §6.8. |
| Providers are legitimate seams. | Catalogs/assembly | Ownership inversion. | `intentional/characterized` | Evidence/consumers | Retained after the five S07 owner suites proved closure; logical support targeting now uses an owner contribution instead of a cross-owner table read. |
| No backend grid-vendor import. | Import scan | Invented architecture. | `defer` | Frontend adapter | Reopen only with evidence. |
| Test helpers are test-owned. | Test support | Production leakage. | `intentional/no_action` | Test support | Preserve. |
| Generated artifacts are read-only. | Policy/current state | Drift. | `intentional/no_action` | Generators | Never hand-edit. |

## 6. Remediation Workstreams

The adopted Core and Reporting amendments are current. The former adoption
blocker is closed. Work is strictly serial so every slice leaves one reviewable
and validated state before the next begins.

| Slice | Gap | State | Depends on | Principal owner(s) | Rollback boundary | Exit evidence |
| --- | --- | --- | --- | --- | --- | --- |
| S00 | G01 authority/tracker rebaseline | DONE | None | Documentation/owners | Revert tracker checkpoint | Adopted owners reconciled; this sequence and ledger current. |
| S01 | G02 route authorization/upload leases | DONE | S00 | Evidence/Auth/Recovery/Web | Additive schema retained; application rollback fails new targets closed | AC-543 matrix, frontend credentials/CSRF/no-retry, old tokens rejected. |
| S02 | G03 canonical Evidence row | DONE | S01 | Evidence/Timeline/Projections | Revert mapper, attachment contribution, and link-refresh call together | One mapper; correct linked count; differential parity. |
| S03 | G04 projection-effects ownership | DONE | S02 | Evidence/Projections | Revert interface, adapter, and assembly atomically | Registry-owned deterministic effects; rollback emits none. |
| S04 | G05 Store decomposition | DONE | S03 | Evidence | Revert cohesive capability set together | No broad Store or compatibility facade. |
| S05 | G06 Workbook/Import composition | DONE | S04 | Evidence/Workbook/Imports | Revert owner runtime and consumer wiring together | Ordinary/import parity and typed construction errors. |
| S06 | G07 lifecycle/reference policy | DONE | S05 | Evidence | Revert policy package, callers, explicit browser transition, and fixtures together | One policy interpretation across entry points. |
| S07 | G08 provider/Reporting closure | DONE | S06 | Evidence/Reporting/consumers | Revert provider changes independently | Exact allowlist and five provider suites pass. |
| S08 | G09 durable cleanup engine | DONE | S07 | Evidence/Recovery/Migrations | Engine remains production-inert; retain additive schema | Exact cleanup, full Evidence service, Recovery inventory, migration, generation, and shape gates pass. |
| S09 | G10 cleanup activation/telemetry | DONE | S08 | Evidence/Object store/Telemetry/App | Stop/remove private dispatcher; expired claims recover | Cadence/shutdown/storage/telemetry matrix passes. |
| S10 | G11 validation and handoff | DONE | S09 | Cross-owner/Harness | No behavior change | Full command matrix passes; tracker closes all gaps. |

### 6.1 S01 completion record

S01 completed at 2026-08-11T19:42:54-04:00. Public Evidence route paths and
response schemas are unchanged. Upload targets now use version-2 signed claims
and one durable lease per blob, bound to actor, current session, incident, blob,
method, required headers, accepted contract, size, optional hash, issue/expiry
times, and one-way lease state. Authentication and CSRF precede capability
decoding; current visibility, role, binding, semantics, lifecycle,
version/idempotency, and object-store access follow the adopted order. Attach
preflight prevents object-store calls before current lifecycle, lease, row
version, and replay checks. Handle issuance and redemption now apply current
role membership before request diagnostics or byte access.

The frontend performs one same-origin credentialed upload with the current CSRF
token and exactly the server-issued upload headers. It fails locally if the CSRF
cookie is absent and does not retry an ambiguous or failed upload. Version-1
targets are rejected without a compatibility decoder.

| S01 evidence item | Recorded result |
| --- | --- |
| Principal authored/runtime files | `db/migrations/00030_evidence_upload_leases.sql`; Evidence route, admission, token, store, repositories, API, Recovery contribution, frontend upload service/helpers, and authored Evidence test-family input |
| Generated projections | Recovery contract Go, SQL model, migration history, schema-object ownership, and execution-topology render index produced through `make generate` |
| Migration impact | Additive version 30; immutable rebaseline remains through version 29; new table is `security_ephemeral` / `excluded_security_state` / `invalidate`; catalog is 112 total / 83 authoritative |
| Compatibility | Clean capability cutover. No version-1 decoder or lease backfill. Claimed/ambiguous targets never reopen. Existing pending legacy slots remain eligible for timeout and later cleanup. |
| Rollback | Drain old issuers before enabling version 2. Application rollback may leave migration 30 in place, but old and new targets are intentionally non-interoperable. |
| Remaining risk | Deployment must enforce the replica-drain cutover. Durable orphan deletion remains inactive until S08-S09; no later slice may weaken lease binding or one-shot state. |

| Command | Result | Run root or diagnostic evidence |
| --- | --- | --- |
| `make test-slice OWNER=module.evidence` | PASS, 57/57 rows | `.cartulary/test-results/20260811T232748Z-p359291` |
| `make service-backed-test-slice OWNER=module.evidence` | PASS, 46/46 rows | `.cartulary/test-results/20260811T232858Z-p390810` |
| AC-543 upload/preference row | PASS, authentication, CSRF, role, session, membership, closed state, binding, denial effects, permitted roles, and replay | `.cartulary/test-results/20260811T232716Z-p357922` |
| Legacy capability row | PASS, version-1 signed target rejected | `.cartulary/test-results/20260811T234212Z-p571048` |
| Affected frontend rows | PASS, Evidence upload, transport boundary, Workbook shell, and Timeline adapter | `.cartulary/test-results/20260811T233333Z-p463270`; `.cartulary/test-results/20260811T233340Z-p463800`; `.cartulary/test-results/20260811T233431Z-p468478`; `.cartulary/test-results/20260811T233718Z-p511793` |
| `make frontend-unit` | PASS, 364/364 units | `.cartulary/test-results/20260811T233824Z-p517026` |
| `make frontend-typecheck` | PASS, 2/2 units | `.cartulary/test-results/20260811T233809Z-p516502` |
| `make migration-drift` | PASS, 5/5 units | `.cartulary/test-results/20260811T234028Z-p556056` |
| `make generate`; `make generate-drift` | PASS | `.cartulary/test-results/20260811T234224Z-p572084`; `.cartulary/test-results/20260811T234310Z-p575389` |
| `make generated-artifact-policy-check` | PASS, 3/3 units | `.cartulary/test-results/20260811T234258Z-p574917` |
| `make lint-markdown` | PASS after S01 tracker checkpoint | `.cartulary/test-results/20260811T234433Z-p578557` |
| Diagnostic failures retained | Initial owner rows exposed grants, timestamp precision, direct-test lease setup, and stale frontend assumptions; generation exposed authored migration/recovery ordering gates. All were corrected and rerun. | `.cartulary/test-results/20260811T230605Z-p173374`; `.cartulary/test-results/20260811T231053Z-p221911`; `.cartulary/test-results/20260811T231329Z-p227109`; `.cartulary/test-results/20260811T231428Z-p235798`; `.cartulary/test-results/20260811T233010Z-p421020`; `.cartulary/test-results/20260811T233441Z-p469059` |

### 6.2 S02 completion record

S02 completed at 2026-08-11T20:09:18-04:00. The typed
`workbookprojection.ProjectionInput` is now the canonical Evidence database-row
model and `workbookprojection.ViewRow` is the only Evidence-specific serializer
of the public row shape. Projections transaction reads load authoritative source
facts through the provider and invoke that mapper. Store attach/quarantine
paths now use the same `Rows` port for revision and mutation values. The
parallel Store query, untyped row model, mapper, conversion helpers, and
hard-coded zero linked count were deleted.

The differential fixture exposed and closed a related stale-projection defect:
Timeline attachment link add/remove mutations changed the authoritative count
but did not refresh the persisted Evidence row. Timeline now passes only the
affected Evidence identities to the Evidence-owned attachment contribution;
that contribution sorts and deduplicates identities and refreshes the primary
Evidence projection through its injected Projections port in the same
transaction. Counts therefore move immediately from zero to one to two and
back to one when a link is tombstoned, with no whole-incident rebuild.

| S02 evidence item | Recorded result |
| --- | --- |
| Principal authored/runtime files | Evidence `workbookprojection` contribution/fixtures, projection provider adapter, Store row call sites, Evidence timeline attachment contribution, Timeline Evidence port/link mutation path, Timeline assembly adapter, affected test compositions, Evidence integration fixture, and authored Evidence test-family input |
| Deleted implementation | `internal/modules/evidence/store_test.go` moved to the owner boundary; `loadEvidenceRowTx`, `evidenceRowData`, `evidenceRowFromData`, `publicCellValue`, and `uuidFromDBValue` removed |
| Generated projections | Execution-topology render index refreshed through `make generate` after registering the canonical mapper row |
| Migration impact | None |
| Compatibility | Public routes and row schemas are unchanged. `evidence.linked_record_count` is intentionally corrected and now becomes current immediately after Timeline link changes; no stale-count shim is retained. |
| Rollback | Mapper call sites, Evidence attachment contribution method, Timeline port call, and registered fixture must revert together. No database rollback is involved. |
| Remaining risk | S03 still owns neutral support-view effect selection and removal of whole-incident support rebuilds; S02 changes only primary Evidence row freshness. |

| Command | Result | Run root or diagnostic evidence |
| --- | --- | --- |
| Canonical mapper row | PASS, rich/null fields plus lifecycle/blob-state matrix | `.cartulary/test-results/20260811T235647Z-p661988` |
| Attach/query/link-count differential row | PASS, zero/one/two/tombstoned counts and exact mutation/query parity | `.cartulary/test-results/20260812T000414Z-p738966` |
| `make test-slice OWNER=module.evidence` | PASS, 58/58 rows | `.cartulary/test-results/20260812T000452Z-p741598` |
| `make service-backed-test-slice OWNER=module.evidence` | PASS, 46/46 rows | `.cartulary/test-results/20260812T000559Z-p773644` |
| `make test-slice OWNER=module.timeline` | PASS, 68/68 rows | `.cartulary/test-results/20260812T000656Z-p801195` |
| `make service-backed-test-slice OWNER=module.timeline` | PASS, 46/46 rows | `.cartulary/test-results/20260812T000807Z-p833141` |
| `make backend-module-boundary-check` | PASS, 3/3 units | `.cartulary/test-results/20260812T000441Z-p741169` |
| `make generate`; `make generate-drift` | PASS | `.cartulary/test-results/20260812T000043Z-p728157`; `.cartulary/test-results/20260812T000059Z-p730944` |
| `make lint-markdown` | PASS after S02 tracker checkpoint | `.cartulary/test-results/20260812T001148Z-p861302` |
| Diagnostic failures retained | The first fixture correction distinguished requested-linked from available-attached Timeline counts. Expanded coverage then exposed the stale Evidence projection and one test-capability compile assumption; both were corrected without weakening assertions. | `.cartulary/test-results/20260811T235112Z-p587166`; `.cartulary/test-results/20260811T235654Z-p662393`; `.cartulary/test-results/20260811T235743Z-p667264` |

### 6.3 S03 completion record

S03 completed at 2026-08-11T20:38:41-04:00. Evidence now loads only the
authoritative active attachment subjects and passes their record IDs/types to
the neutral `SupportProjectionEffectsTx` boundary. Projections owns the
record-type registry, registered-view selection, before/after state loads,
exact row refresh, physical projection storage access, field-key vocabulary,
and deterministic effect construction. The primary Evidence `Rows` interface
no longer exposes support refresh. The former whole-incident Timeline, Host,
and Identity rebuild path was removed.

Current support changes remain sorted `invalidate` effects with no patch.
Collaboration now preserves the existing array-shaped protocol for one or many
affected views, orders them by view-schema ID, rejects duplicates, and retains
the established `remove` change kind for record deletion. Evidence does not
know any support view ID or field key. Projections trusts the authoritative
subject list supplied by Evidence and does not read the Records owner table.
Host and Identity before/after reads are behind Projections' private storage
adapter, which keeps the module boundary clean.

The integration fixture corrupts targeted and unrelated support rows, invokes
the effects boundary in a caller-owned transaction, checks the exact neutral
result, proves targeted rows refresh once, rolls the transaction back, and
proves every prior value remains. The subsequent real attach proves only the
three linked subjects refresh while unrelated Timeline, Host, and Identity
rows retain sentinel state. An unregistered Note subject produces no effect.

| S03 evidence item | Recorded result |
| --- | --- |
| Principal authored/runtime files | Evidence `workbookprojection` ports/effect records, Store/runtime/collaboration adaptation and integration fixture; Projections adapter, provider registry/factories, effect coordinator and private storage reads; server/test composition; Collaboration protocol/intent test; authored Collaboration test-family input |
| Generated projections | Execution-topology render index refreshed through `make generate`; no generated root was hand-edited |
| Residual S01 tooling closure | JSON-shape validation now consumes the canonical migration-history filename allocation instead of freezing the 29-file immutable baseline; migration owner 30 remains explicit. This closes the stale validator exposed by the S03 validation pass. |
| Migration impact | None in S03; migration 30 and its Recovery classification are unchanged |
| Compatibility | Public HTTP routes, schemas, and current single-view support invalidations are unchanged. The internal Collaboration payload parser now correctly accepts deterministic multi-view arrays while preserving `patch`, `invalidate`, and `remove`. No support-effects compatibility facade exists. |
| Rollback | Revert the Evidence port/call, Projections registry/coordinator/storage seam, assembly injection, and multi-view Collaboration adaptation together. No database rollback is involved. |
| Remaining risk | The broad Store risk identified at S03 was resolved in §6.4. Future support patches remain deliberately unimplemented until a registered provider supplies canonical patch state. |

| Command | Result | Run root or diagnostic evidence |
| --- | --- | --- |
| Exact association-effects integration row | PASS, exact subjects/views/fields, unregistered omission, unrelated-row preservation, stable order, and rollback | `.cartulary/test-results/20260812T003741Z-p1127894` |
| Projections provider-registry row | PASS, registry selection/order and fail-closed loader validation | `.cartulary/test-results/20260812T003741Z-p1127897` |
| `make test-slice OWNER=module.evidence` | PASS, 58/58 rows | `.cartulary/test-results/20260812T003059Z-p945140` |
| `make service-backed-test-slice OWNER=module.evidence` | PASS, 46/46 rows | `.cartulary/test-results/20260812T003251Z-p1011843` |
| `make test-slice OWNER=module.projections` | PASS, 17/17 rows | `.cartulary/test-results/20260812T003059Z-p945153` |
| `make service-backed-test-slice OWNER=module.projections` | PASS, 13/13 rows | `.cartulary/test-results/20260812T003251Z-p1011838` |
| Collaboration semantic/stateful rows | PASS, 14/14 units | `.cartulary/test-results/20260812T003511Z-p1066680` |
| `make test-slice OWNER=module.collaboration` | PASS, 44/44 rows | `.cartulary/test-results/20260812T003601Z-p1091697` |
| `make backend-module-boundary-check` | PASS, 3/3 units | `.cartulary/test-results/20260812T003735Z-p1127439` |
| `make generate`; `make generate-drift` | PASS | `.cartulary/test-results/20260812T004017Z-p1134823`; `.cartulary/test-results/20260812T004029Z-p1137549` |
| `make json-shape-check` | PASS, 3/3 units | `.cartulary/test-results/20260812T004029Z-p1137556` |
| `make lint-markdown` | PASS after S03 tracker checkpoint | `.cartulary/test-results/20260812T004257Z-p1141605` |
| Diagnostic failures retained | Initial runs exposed unmapped source-hydrated state reads, a primary-row assertion mixed into support effects, one expected-order typo, a dormant Collaboration helper, missing `remove` preservation, direct Records/projection-table boundary violations, and the stale 29-migration JSON validator. Each was structurally corrected and rerun. | `.cartulary/test-results/20260812T002209Z-p871055`; `.cartulary/test-results/20260812T002703Z-p919110`; `.cartulary/test-results/20260812T002751Z-p925345`; `.cartulary/test-results/20260812T002939Z-p935440`; `.cartulary/test-results/20260812T003350Z-p1040269`; `.cartulary/test-results/20260812T003601Z-p1091816`; `.cartulary/test-results/20260812T003841Z-p1133169` |

### 6.4 S04 completion record

S04 completed at 2026-08-11T20:55:59-04:00. The former Evidence `Store`
service locator, `NewStore`, variadic `StoreOption`, and option constructors were
removed without a compatibility facade. Owner composition now uses separate
source-mutation, blob-lifecycle, access-handle, cleanup, and route-operation
capabilities. Route operations embed only blob lifecycle and access handles;
cleanup and source mutation are not reachable from transport. Cleanup remains
an explicit, production-inert capability pending the durable S08 engine.

Workbook and Imports now consume the owner-private source-mutation service,
while server routes receive a composed `RouteOperations`. Focused tests use
application support to obtain the same narrow capabilities instead of
constructing Evidence internals. At the S04 checkpoint, the immutable
`OwnerRuntime` was characterized as three private interface fields: routes,
Workbook contribution, and Timeline attachment contribution. Each cohesive
constructor rejected its required missing dependencies with an error. S05
subsequently replaced the remaining panic-based broad composition and added the
narrow import capability, as recorded in §6.5.

| S04 evidence item | Recorded result |
| --- | --- |
| Principal implementation files | Evidence service definitions/constructors and receivers in `store.go`; owner runtime, route interface, Workbook/import facades, initial-blob access, test application support, direct Evidence tests, and server composition characterization |
| Removed compatibility surface | `Store`, `NewStore`, `StoreOption`, `WithWorkbookProjections`, `WithSupportProjectionEffects`, `WithRevisionAppender`, and `WithCollaborationIntents`; repository-wide Go search finds no Evidence caller or type reference |
| Authored verification input | Existing Evidence source-owner contribution row now also selects `TestEvidenceServiceConstructionRejectsIncompleteDependencies`; topology regenerated through `make generate` |
| Migration impact | None; migration 30 and Recovery classification remain unchanged |
| Compatibility | Public HTTP routes, request/response schemas, upload targets, authorization order, Workbook behavior, projection effects, and Collaboration events are unchanged. Internal direct construction breaks intentionally; no facade or option shim exists. |
| Rollback | Revert cohesive services, owner/runtime/facade wiring, test support, and caller migrations together. No database or public-client rollback is involved. |
| Remaining risk at the S04 checkpoint | Owner runtime and Workbook construction still panicked and Workbook/Imports accepted broad parallel dependency lists. G06/S05 closes that historical risk in §6.5; cleanup remains safely inert until S08-S09. |

| Command | Result | Run root or diagnostic evidence |
| --- | --- | --- |
| Constructor/source-owner row | PASS, all missing cohesive dependency cases rejected | `.cartulary/test-results/20260812T005336Z-p1217861` |
| Server narrow-runtime row | PASS, exact private capability surface | `.cartulary/test-results/20260812T005243Z-p1213395` |
| `make test-slice OWNER=module.evidence` | PASS, 58/58 rows | `.cartulary/test-results/20260812T005415Z-p1221972` |
| `make service-backed-test-slice OWNER=module.evidence` | PASS, 46/46 rows | `.cartulary/test-results/20260812T005415Z-p1221975` |
| `make backend-module-boundary-check` | PASS, 3/3 units | `.cartulary/test-results/20260812T005450Z-p1272458` |
| `make generate`; `make generate-drift` | PASS | `.cartulary/test-results/20260812T005347Z-p1218915`; `.cartulary/test-results/20260812T005450Z-p1272069` |
| `make json-shape-check` | PASS, 3/3 units | `.cartulary/test-results/20260812T005450Z-p1272119` |
| `make format`; `git diff --check` | PASS | `.cartulary/test-results/20260812T005329Z-p1214509`; command completed without findings |
| `make lint-markdown` | PASS after S04 tracker checkpoint | `.cartulary/test-results/20260812T005829Z-p1287406` |

### 6.5 S05 completion record

S05 completed at 2026-08-11T21:21:27-04:00. Evidence now owns one validated
`OwnerRuntimeDependencies` composition contract. Its constructor returns
descriptive errors for incomplete Postgres, conflict, Revisions, Collaboration,
object-store, or Projections dependencies; no runtime or Workbook constructor
panics. The immutable runtime publishes only four narrow capabilities: routes,
Workbook, Timeline attachments, and the Imports create facade.

Ordinary Workbook creation and imported Evidence creation now delegate to one
private source-mutation kernel. The kernel owns the common source row,
projection, revision, mutation, and Collaboration-intent sequence while each
consumer retains its transaction, idempotency, and orchestration responsibilities.
Imports no longer imports Evidence projection implementation or reconstructs
its dependency graph. Evidence constructs its own import contribution and
Imports accepts only `ownerfacade.ImportOwnerCreateFacade`.

The parity fixture creates equivalent normalized Evidence through the ordinary
and import entry points at the same requested time. It proves equal source
content and projection cells, one revision, one mutation, one Collaboration
intent, one projection row, and exact-replay idempotency without duplicate
Evidence records. Construction and assembly fixtures prove invalid dependency
sets fail with errors and consumer assembly has no broad Evidence projection
dependency.

| S05 evidence item | Recorded result |
| --- | --- |
| Principal implementation files | Evidence owner runtime, source-mutation service/kernel, Workbook facade/contribution, import facade/projection adapter, server assembly, Imports owner registry, and reusable application test support |
| Removed compatibility surface | Public `NewWorkbookContribution`, `NewWorkbookFacade`, and broad `NewImportCreateFacade`; Imports `EvidenceProjections` dependency; consumer-owned Evidence reconstruction; panic-based runtime construction |
| Authored verification input | Evidence owner-runtime constructor row, Imports parity selector, and Import assembly narrow-dependency selectors; topology regenerated only through `make generate` |
| Migration impact | None; migration 30 and its Recovery classification remain unchanged |
| Compatibility | Public HTTP, Workbook, import, row, revision, and Collaboration contracts are unchanged. Internal constructors break intentionally and have no compatibility facade. |
| Rollback | Revert `OwnerRuntimeDependencies`, the shared kernel, server/test composition, Imports registry wiring, and authored selectors together. No database or public-client rollback is involved. |
| Remaining risk | Lifecycle/reference policy remains duplicated across ordinary, rollback, and supporting paths. This is exclusively G07/S06; cleanup remains safely inert until S08-S09. |

| Command | Result | Run root or diagnostic evidence |
| --- | --- | --- |
| Ordinary/import parity row | PASS, normalized source/projection/effects equality and replay uniqueness | `.cartulary/test-results/20260812T012058Z-p1428081` |
| Owner-runtime/kernel construction rows | PASS, complete composition and every required missing dependency rejected | `.cartulary/test-results/20260812T011451Z-p1321584` |
| Imports assembly narrow-dependency row | PASS, no consumer-owned Evidence reconstruction | `.cartulary/test-results/20260812T011451Z-p1321578` |
| Server narrow-runtime row | PASS, exact four-capability owner runtime | `.cartulary/test-results/20260812T011451Z-p1321596` |
| `make test-slice OWNER=module.evidence` | PASS, 58/58 rows | `.cartulary/test-results/20260812T011531Z-p1324790` |
| `make service-backed-test-slice OWNER=module.evidence` | PASS, 46/46 rows | `.cartulary/test-results/20260812T011638Z-p1356684` |
| `make service-backed-test-slice OWNER=module.imports` | PASS, 29/29 units | `.cartulary/test-results/20260812T011748Z-p1384705` |
| Workbook composition row | PASS | `.cartulary/test-results/20260812T011834Z-p1411248` |
| `make backend-module-boundary-check` | PASS, 3/3 units | `.cartulary/test-results/20260812T012030Z-p1424182` |
| `make format`; `make generate`; `make generate-drift` | PASS | `.cartulary/test-results/20260812T012013Z-p1417772`; `.cartulary/test-results/20260812T012016Z-p1421064`; `.cartulary/test-results/20260812T012030Z-p1423950` |
| `make json-shape-check`; `git diff --check` | PASS | `.cartulary/test-results/20260812T012030Z-p1423985`; command completed without findings |
| `make lint-markdown` | PASS after S05 tracker checkpoint | `.cartulary/test-results/20260812T012335Z-p1430741` |
| Diagnostic failures retained | The first parity run exposed implicit request-time divergence and was corrected with one explicit owner input. Initial boundary/drift/shape runs exposed a forbidden consumer identity literal and stale generated topology after the final authored selector change; both were structurally corrected and rerun. One malformed focused command used a suite alias instead of the catalog row ID and changed no state. | `.cartulary/test-results/20260812T011112Z-p1307353`; `.cartulary/test-results/20260812T011854Z-p1413216`; `.cartulary/test-results/20260812T011853Z-p1413030`; `.cartulary/test-results/20260812T011854Z-p1413054` |

### 6.6 S06 completion record

S06 completed at 2026-08-11T22:07:42-04:00. The new Evidence-private
`internal/policy` package is the single owner interpretation of Evidence and
Object Blob lifecycle vocabularies and transitions, lifecycle/blob bridges,
association admission, closed-incident mutation posture, server-managed
references, persisted object-key identity, finalize exhaustion, and failure
scheduling. Ordinary mutation, imported creation, rollback, blob association,
finalization, quarantine, pending timeout, route access, and cleanup candidate
selection now call this policy instead of maintaining parallel rules.

`blobref` is again limited to logical-reference and physical-key syntax. A
malformed value beginning with the reserved `object://` scheme is treated as a
server-managed reference and rejected rather than escaping owner validation.
Rollback validates both lifecycle/blob bridges and the logical reference/blob
identity association through the same policy as ordinary mutation.

Attachment no longer changes `evidence.lifecycle_state` implicitly. It links
the finalized blob and exposes its independent upload state; an explicit,
versioned Workbook patch moves Evidence to `available`. This is an
owner-directed correction, not preserved legacy behavior. The browser client,
direct E2E helpers, and test composition perform the explicit transition. A
lifecycle patch also invokes the Projections-owned support-effects boundary in
the same transaction so registered support counts remain current without
restoring whole-incident coupling. Failure eligibility is now exactly
`failed_at + 45 minutes`, leaving one 15-minute dispatcher interval before the
adopted one-hour deletion deadline.

| S06 evidence item | Recorded result |
| --- | --- |
| Principal implementation files | Evidence private policy package; create validation, Workbook mutation/facade, rollback provider, route/dependency, persistence, blob-lifecycle, and `blobref` callers; browser mutation command; E2E/shared fixtures; affected integration tests |
| Removed duplicate behavior | Local lifecycle/blob validators, transition and bridge switches, route-local persisted-key error/validation, rollback-local state vocabularies, and `blobref.IsServerManagedStorageRef` |
| Authored verification input | Two Evidence unit rows cover the shared policy matrix and rollback bridge/reference behavior; topology was regenerated through `make generate` |
| Migration impact | None; S06 changes policy and timing only. Migration 30 and its Recovery classification remain unchanged. |
| Compatibility | Public routes and schemas remain unchanged. Blob attachment intentionally preserves the current Evidence lifecycle; callers that need handle-eligible Evidence now issue an explicit `available` patch. No auto-promotion shim remains. |
| Rollback | Revert the policy package, all production callers, explicit client transition, projection refresh on lifecycle patch, and fixtures together. No database rollback is involved. |
| Remaining risk | Reporting allowlist and provider closure remain G08/S07. Cleanup is still production-inert; S08 must add durable claims before S09 activation. |

| Command | Result | Run root or diagnostic evidence |
| --- | --- | --- |
| Shared policy row | PASS, lifecycle/blob matrices, bridges, incident posture, references, failure threshold, and 45-minute schedule | `.cartulary/test-results/20260812T013518Z-p1445307` |
| Rollback shared-policy row | PASS, invalid states, bridge violations, malformed/mismatched logical refs, and valid association | `.cartulary/test-results/20260812T013518Z-p1445324` |
| Focused lifecycle/create/attach/access/quarantine rows | PASS | `.cartulary/test-results/20260812T013518Z-p1445365`; `.cartulary/test-results/20260812T013518Z-p1445397`; `.cartulary/test-results/20260812T013657Z-p1456576`; `.cartulary/test-results/20260812T013657Z-p1456581`; `.cartulary/test-results/20260812T013657Z-p1456592` |
| Corrected route/projection/WebSocket rows | PASS | `.cartulary/test-results/20260812T014750Z-p1547788`; `.cartulary/test-results/20260812T014804Z-p1549090`; `.cartulary/test-results/20260812T014521Z-p1536749` |
| Browser/a11y/stateful regression rows | PASS in the focused six-row run; five selected behavior rows passed before the visual timing assertion was corrected | `.cartulary/test-results/20260812T020128Z-p1663534` |
| Visual explicit-transition row | PASS after awaiting the completed two-mutation workflow | `.cartulary/test-results/20260812T020336Z-p1696009` |
| `make test-slice OWNER=module.evidence` | PASS, 60/60 rows | `.cartulary/test-results/20260812T020415Z-p1719303` |
| `make service-backed-test-slice OWNER=module.evidence` | PASS, 46/46 rows | `.cartulary/test-results/20260812T020521Z-p1750627` |
| Ordinary/import parity row | PASS, shared kernel remains equivalent after policy consolidation | `.cartulary/test-results/20260812T020652Z-p1778837` |
| `make frontend-unit`; `make frontend-typecheck` | PASS, 364/364 and 2/2 units | `.cartulary/test-results/20260812T015423Z-p1589451`; `.cartulary/test-results/20260812T020118Z-p1663119` |
| `make backend-module-boundary-check` | PASS, 3/3 units | `.cartulary/test-results/20260812T020703Z-p1780638` |
| `make generate-drift`; `make json-shape-check`; `git diff --check` | PASS | `.cartulary/test-results/20260812T020705Z-p1780959`; `.cartulary/test-results/20260812T020713Z-p1783739`; command completed without findings |
| `make lint-markdown` | PASS after S06 tracker checkpoint and final evidence-root insertion | `.cartulary/test-results/20260812T021006Z-p1785471`; `.cartulary/test-results/20260812T021042Z-p1786567` |
| Diagnostic failures retained | Initial service-backed runs exposed stale attachment fixtures and the prohibited implicit lifecycle promotion. The first full owner rerun then identified the remaining backend helper, browser workflow, shared E2E fixture, and visual completion timing assumptions. Each was corrected to perform or await the explicit owner transition. One focused Imports command used a guessed row ID, failed before execution, and changed no state. | `.cartulary/test-results/20260812T013535Z-p1448489`; `.cartulary/test-results/20260812T013535Z-p1448482`; `.cartulary/test-results/20260812T013535Z-p1448499`; `.cartulary/test-results/20260812T013856Z-p1464589`; `.cartulary/test-results/20260812T013856Z-p1464608`; `.cartulary/test-results/20260812T014832Z-p1550507`; `.cartulary/test-results/20260812T015650Z-p1629143` |

### 6.7 S07 completion record

S07 completed at 2026-08-11T22:36:03-04:00. The Evidence Reporting
provider now constructs its value with the adopted twelve named members. It no
longer serializes a whole database row and subtracts known private fields.
Unknown future columns, physical object identity, capability material, and
deleted Evidence therefore remain private by construction.

Evidence now contributes its eligible logical support targets as
`/evidence/{record_id}`. Reporting composes that owner contribution with the
Links-owned relationship read; Links remains generic and no longer reads the
Records owner table to infer a destination subtype. Reporting then normalizes
the contributed path to the adopted `evidence:{evidence_id}` identity. This
keeps selection and redaction with their semantic owners and admits future
logical target families without another cross-owner join.

The five named behavioral suites are registered through authored test-family
inputs. The Reporting suite mutates only its isolated test database with a
future private Evidence column and proves exact key equality, null/timestamp
preservation, deleted-row exclusion, deterministic facts, logical support
identity, forbidden-value absence in both facts and the persisted export model,
and all-or-nothing rejection of malformed provider output. Imports retains the
ordinary/import kernel parity proof. Incident Bundles retains its broad failure
family and proves staged object cleanup. Recovery proves exact current relation
classification and exactly-once available-object inventories through both
current and vNext provider seams. Evidence proves exact owner contributions and
typed construction rejection.

The Incident Bundles suite also exposed two test/implementation durability
issues. Records explicitly closes its envelope cursor before invoking subtype
providers on the same transaction. The failure-job assertion now allows the
governed 30-second object-store mutation deadline plus processing margin and
retains the last observed job state on timeout. No production storage deadline
or retry policy changed.

| S07 evidence item | Recorded result |
| --- | --- |
| Principal implementation files | Evidence Reporting provider; Reporting materializer and exact provider tests; Links Reporting provider; Records incident-bundle validator; Incident Bundles wait diagnostics; Recovery inventory test |
| Authored verification inputs | Reporting redaction, Imports parity, Incident Bundles staging cleanup, Recovery inventory closure, and Evidence provider-contribution selectors; Incident Bundle fixture names updated with the renamed suite |
| Generated projections | Execution-topology render index regenerated only through `make generate`; no generated root was hand-edited |
| Migration impact | None; migration 30 and its Recovery classification remain unchanged |
| Compatibility | Public Evidence and Reporting routes/schemas remain unchanged. Accidental Reporting fields outside the adopted allowlist disappear intentionally. Internal Links support-ref collection now receives logical target contributions instead of inspecting Records. |
| Rollback | Revert the closed Evidence provider query, logical-target contribution/composition, Links resolver input, exact suites, and authored routing together. No database rollback is involved. |
| Remaining risk | Cleanup is still production-inert. S08 must add the durable transient claim state and engine before S09 activation; provider closure is now a retained regression gate. |

| Command | Result | Run root or diagnostic evidence |
| --- | --- | --- |
| `reporting.evidence_provider_redaction` | PASS, exact allowlist, future-field omission, deleted exclusion, logical identity, export-model absence, malformed-output atomicity | `.cartulary/test-results/20260812T023334Z-p1853426` |
| `imports.evidence_create_parity` | PASS, shared ordinary/import mutation behavior and replay | `.cartulary/test-results/20260812T022345Z-p1810786` |
| `incident_bundles.evidence_blob_staging_cleanup` | PASS, real export/import failures, staged-object cleanup, and no residual staging | `.cartulary/test-results/20260812T023110Z-p1841594` |
| `recovery.evidence_inventory_closure` | PASS, exact relations/classification and current/vNext available-object inventories | `.cartulary/test-results/20260812T022329Z-p1807831` |
| `module.evidence.provider_contribution_closure` | PASS, exact Incident Bundle, Recovery, Revisions, subtype, service, and owner-runtime contributions | `.cartulary/test-results/20260812T022358Z-p1812186` |
| `make test-slice OWNER=module.reporting` | PASS, 14/14 units after logical-target composition | `.cartulary/test-results/20260812T023524Z-p1862898` |
| Records portability integration row | PASS after cursor-lifetime correction | `.cartulary/test-results/20260812T023227Z-p1848371` |
| `make backend-module-boundary-check` | PASS, 3/3 units | `.cartulary/test-results/20260812T023517Z-p1862481` |
| `make generate`; `make generate-drift`; `make json-shape-check` | PASS | `.cartulary/test-results/20260812T023147Z-p1842982`; `.cartulary/test-results/20260812T023355Z-p1854702`; `.cartulary/test-results/20260812T023410Z-p1857546` |
| `make format`; `git diff --check` | PASS | `.cartulary/test-results/20260812T023509Z-p1859061`; command completed without findings |
| `make lint-markdown` | PASS after the S07 tracker checkpoint | `.cartulary/test-results/20260812T023814Z-p1867908` |
| Diagnostic failures retained | The first Reporting run exposed a test-package import cycle and was split into external integration plus internal validation tests. Incident Bundles then exposed its 10-second assertion horizon against a 30-second object-store mutation contract. The boundary gate rejected a provisional Links-to-Records join and drove the owner-contributed target design. All were corrected and rerun without weakening product assertions. | `.cartulary/test-results/20260812T022155Z-p1800777`; `.cartulary/test-results/20260812T022409Z-p1813265`; `.cartulary/test-results/20260812T022805Z-p1831170`; `.cartulary/test-results/20260812T023420Z-p1858080` |

### 6.8 S08 completion record

S08 reached an implementation-complete but unvalidated checkpoint on
2026-08-11T23:10:21-04:00. Migration 31 adds the Evidence-private durable
cleanup claim relation, bounded claim/retry/completion metadata, indexes, and
database guards that reject blob-state or Evidence-association mutation while
a claim exists. The cleanup capability first fails expired pending slots,
claims at most 100 eligible rows with `FOR UPDATE SKIP LOCKED`, uses 5-minute
leases, deletes bytes outside the transaction under a 1-minute context,
retries after 1, 5, and repeated 15-minute delays, rechecks claim ownership and
current failed/unattached state before completion, and hard-deletes metadata
only after 7 days. It remains absent from server assembly and therefore inert.

Recovery classifies `evidence_blob_cleanup_claims` as
`transient`/`excluded_transient`/`invalidate`; the catalog is now exactly
113/83. Core 01, Core 02, and Core 04 projections, migration allocation,
schema ownership validation, Recovery contracts, exact owner tests, and
authored Evidence routing were updated. Generated outputs were produced only
through `make generate`.

The service-backed matrix covers inert partial deployment, two-instance claim
exclusion, lifecycle/association races, restart after lease expiry, batches
larger than 100, deterministic unbounded retries, typed not-found, the delete
deadline, pending-timeout compliance, and seven-day metadata retention. The
first three exact attempts on 2026-08-11 failed before test execution because
Docker container inspection exceeded the readiness deadline. After the host
was repaired, the matrix reached product execution on 2026-08-12. That run
identified an untyped nullable-timestamp test fixture and, because the failed
subtest left one eligible row, a contaminated batch assertion. Explicit SQL
types fixed the fixture. The focused Recovery run then detected that the owner
contribution appended transient cleanup claims before the pre-existing
security-ephemeral upload leases; the implementation was reordered to match
the exact owner contract. No assertion or required behavior was weakened.

S08 completed at 2026-08-12T18:39:56-04:00. The exact cleanup matrix, full
Evidence service-backed owner slice, owner contribution closure, exact
Recovery inventory, migration drift, generated drift, and JSON shape checks
all pass. The engine remains absent from server assembly and production-inert
until S09.

| S08 evidence item | Recorded result |
| --- | --- |
| Principal implementation files | Migration 31; `internal/modules/evidence/cleanup.go`; cleanup service/repository simplification; durable cleanup integration matrix |
| Specification and contract projections | Core 01 catalog/migration counts; Core 02 lifecycle cleanup; Core 04 AC-155/AC-542; Recovery catalog/schema/index/fixture; schema-ownership schema and validator |
| Authored verification input | New `module.evidence.integration.durable_cleanup_claim_engine_5be2d5d937` row; generated topology refreshed through Make |
| Migration impact | Additive migration 31. Existing eligible failed/unattached blobs enter policy naturally. Downgrade leaves no deletion caller until S09 activation; retain the additive table across application rollback. |
| Compatibility | No public API change. Cleanup remains production-inert. Recovery authoritative content remains 83 tables; one excluded transient relation raises total catalog cardinality to 113. |
| Rollback | Revert engine, migration 31, Core/Recovery/schema-owner projections, and authored test row together before activation; after S09, stop the dispatcher first and retain claim state for safe lease recovery. |
| Remaining risk | Cleanup remains production-inert. S09 must govern activation, typed object-store deletion, lifecycle startup/shutdown, and closed telemetry before bytes can be deleted operationally. |

| Command | Result | Run root or diagnostic evidence |
| --- | --- | --- |
| `make format` | PASS | `.cartulary/test-results/20260812T030012Z-p1905614` |
| `make generate`; `make generate-drift` | PASS | `.cartulary/test-results/20260812T030344Z-p1933125`; `.cartulary/test-results/20260812T031021Z-p1942960` |
| `make migration-drift` | PASS, 5/5 units | `.cartulary/test-results/20260812T030236Z-p1926624` |
| `make json-shape-check` | PASS, 3/3 units after adding migration 31 and `excluded_transient` to schema-ownership validation | `.cartulary/test-results/20260812T030355Z-p1935892` |
| Exact Evidence policy/unit row | PASS; package including the durable cleanup matrix compiles | `.cartulary/test-results/20260812T030221Z-p1925993` |
| Exact Recovery catalog unit rows | PASS, 2/2 units at 113/83 | `.cartulary/test-results/20260812T030807Z-p1939708` |
| `make build-server` | PASS; S08 engine is not assembled | `.cartulary/test-results/20260812T030203Z-p1913886` |
| `make backend-module-boundary-check` | PASS, 3/3 units | `.cartulary/test-results/20260812T030819Z-p1940453` |
| Exact service-backed cleanup row, attempt 1 | BLOCKED before test execution: Docker PostgreSQL inspect/readiness timeout | `.cartulary/test-results/20260812T030040Z-p1911762` |
| Exact service-backed cleanup row, attempt 2 | BLOCKED before test execution: same Docker PostgreSQL inspect/readiness timeout | `.cartulary/test-results/20260812T030608Z-p1937384` |
| Exact service-backed cleanup row, attempt 3 | BLOCKED before test execution: same Docker PostgreSQL inspect/readiness timeout | `.cartulary/test-results/20260812T030836Z-p1940897` |
| `make services-up` diagnostic | BLOCKED after PostgreSQL provisioning: external unproven listener owns `127.0.0.1:8333` | `.cartulary/test-results/20260812T030721Z-p1938780` |
| `git diff --check` | PASS | Command completed without findings. |
| `make lint-markdown` | PASS after the S08 blocked checkpoint and 66-file inventory refresh | `.cartulary/test-results/20260812T031427Z-p1947222` |
| Exact cleanup row after host repair | PASS, 3/3 work units and all cleanup scenarios | `.cartulary/test-results/20260812T223643Z-p12933` |
| Full `service-backed-test-slice OWNER=module.evidence` | PASS, 47/47 work units | `.cartulary/test-results/20260812T223701Z-p14224` |
| Exact Evidence provider-contribution closure | PASS, 1/1 after contribution-order correction | `.cartulary/test-results/20260812T223905Z-p51642` |
| Exact Recovery Evidence inventory closure | PASS, 3/3 work units | `.cartulary/test-results/20260812T223913Z-p52380` |
| `make migration-drift`; `make generate-drift`; `make json-shape-check` | PASS, 5/5, 4/4, and 3/3 work units | `.cartulary/test-results/20260812T223931Z-p54180`; `.cartulary/test-results/20260812T223939Z-p56796`; `.cartulary/test-results/20260812T223947Z-p59565` |
| Retained post-repair diagnostics | First live cleanup run exposed only fixture parameter typing and consequent row contamination; first Recovery run exposed exact contribution ordering. Both were corrected and rerun without weakening assertions. | `.cartulary/test-results/20260812T223532Z-p7049`; `.cartulary/test-results/20260812T223812Z-p44816` |
| `make lint-markdown` | PASS after the S08 completion checkpoint | `.cartulary/test-results/20260812T224107Z-p60426` |

G09 is closed. S09 may begin only after this completed checkpoint passes
`make lint-markdown`.

### 6.9 S09 completion record

S09 completed at 2026-08-12T19:08:22-04:00. The Evidence-private dispatcher
uses the exact identity
`evidence.failed_unattached_blob_cleanup.v1`, remains stopped throughout
application construction and bootstrap, starts only after publication readiness,
runs its first sweep immediately after activation, and then runs every 15
minutes. Sweeps are serialized within each process, durable claims coordinate
multiple processes, and application cleanup cancels and joins an in-flight
sweep before the object store and telemetry providers close. No common Jobs
registration, route, catalog row, or public control surface was added.

Cleanup deletion is acquired only through the typed object-store capability
with purpose `evidence_cleanup`. Typed not-found is idempotent success; typed
transient failures remain retryable through the S08 claim schedule. The
dispatcher emits a closed observation rather than identifiers or raw errors.
The server observer binds that observation to the adopted Evidence telemetry
scope and the four closed measures for operation/result count, duration,
overdue blob count, and oldest eligible age. Metric and log attributes are
limited to closed `operation`, `result`, and `error_class` values. Incident,
record/blob, object-key, filename, hash, capability, claim-token, and raw-error
values cannot enter the helper output.

S09 also closed a process-test defect uncovered by full server validation. The
test upload helper still used the pre-S01 unauthenticated retry behavior, and
the scenario assumed blob attachment implicitly promoted an Evidence row from
`requested` to `available`. The helper now sends the issued method and headers
once with current session and CSRF state. The scenario performs the owner-
required explicit lifecycle patch before expecting a visible Timeline count,
and asserts both the authoritative active link and the mutation/query
projection values. Product behavior and the S06 lifecycle policy were not
weakened.

| S09 evidence item | Recorded result |
| --- | --- |
| Principal implementation files | `cleanup_dispatcher.go`; `cleanup_objectstore.go`; `cleanup.go`; server runtime and telemetry adapter; platform object-store and telemetry contracts |
| Specification and contract projections | Core 02 dispatcher/storage contract; Core 04 AC-155 activation evidence; adopted OpenTelemetry Evidence scope, closed vocabulary, and four metric rows |
| Authored verification input | New Evidence dispatcher/storage row, server activation/telemetry row, and platform object-store typed-purpose row; generated topology refreshed through Make |
| Migration impact | None in S09. Migration 31 claim state remains authoritative for coordination; existing eligible failed/unattached blobs begin cleanup after activation. |
| Compatibility | No public route, envelope, common Jobs surface, or client schema changes. Operations begin deleting only owner-eligible orphaned bytes. |
| Rollback | Stop and remove the private dispatcher registration before rolling back application code. Retain migration 31 so outstanding claims expire safely and can be resumed by a later compatible deployment. |
| Remaining risk | Operational deletion is now active. S10 must prove the complete cross-owner, browser, security, generated-artifact, migration, and broad repository matrix without weakening the closed telemetry or typed storage boundary. |

| Command | Result | Run root or diagnostic evidence |
| --- | --- | --- |
| Exact Evidence dispatcher and typed-storage unit row | PASS | `.cartulary/test-results/20260812T225453Z-p72880` |
| Exact server activation and telemetry unit row | PASS after moving composition behind bootstrap | `.cartulary/test-results/20260812T225832Z-p144323` |
| Exact platform object-store typed-purpose row | PASS | `.cartulary/test-results/20260812T225515Z-p76354` |
| Exact platform telemetry registry/accessor rows | PASS | `.cartulary/test-results/20260812T225518Z-p76810` |
| Exact live cleanup row including two-dispatcher ownership | PASS | `.cartulary/test-results/20260812T225528Z-p78220` |
| Process Evidence upload/attachment/lifecycle/projection row | PASS, 7/7 units | `.cartulary/test-results/20260812T230435Z-p218356` |
| `make test-slice OWNER=module.evidence` | PASS, 62/62 work units | `.cartulary/test-results/20260812T225543Z-p79443` |
| `make test-slice OWNER=app.server` | PASS, 45/45 work units | `.cartulary/test-results/20260812T230504Z-p238154` |
| `make test-slice OWNER=platform.telemetry` | PASS, 10/10 work units | `.cartulary/test-results/20260812T230541Z-p264681` |
| `make test-slice OWNER=platform.objectstore` | PASS, 6/6 work units | `.cartulary/test-results/20260812T230546Z-p265851` |
| `make service-backed-test-slice OWNER=module.evidence` | PASS, 47/47 work units | `.cartulary/test-results/20260812T230559Z-p267434` |
| `make service-backed-test-slice OWNER=app.server` | PASS, 34/34 work units | `.cartulary/test-results/20260812T230659Z-p299259` |
| `make build-server` | PASS | `.cartulary/test-results/20260812T230733Z-p321787` |
| `make backend-module-boundary-check` | PASS, 3/3 work units | `.cartulary/test-results/20260812T230744Z-p333165` |
| `make generate-drift`; `make generated-artifact-policy-check` | PASS, 4/4 and 3/3 work units | `.cartulary/test-results/20260812T230746Z-p333527`; `.cartulary/test-results/20260812T230753Z-p336299` |
| `make migration-drift`; `make json-shape-check` | PASS, 5/5 and 3/3 work units | `.cartulary/test-results/20260812T230754Z-p336705`; `.cartulary/test-results/20260812T230802Z-p339348` |
| `git diff --check`; `make lint-markdown` | PASS | Diff check completed without findings; `.cartulary/test-results/20260812T231107Z-p340796` |
| Retained diagnostics | First broad server run exposed pre-bootstrap cleanup composition and the stale process upload/lifecycle fixture; both were corrected and rerun without changing normative behavior. | `.cartulary/test-results/20260812T225646Z-p112497`; `.cartulary/test-results/20260812T225844Z-p144912` |

G10 is closed. S10 may begin only after this completed checkpoint passes
`make lint-markdown`.

### 6.10 S10 completion record

S10 completed at 2026-08-12T20:00:26-04:00. The final owner-first matrix,
frontend and browser suites, generated and migration drift, harness contracts,
security checks, finalize gate, fast graph, and full repository graph all pass.
Every required gap is closed, no stale adoption blocker remains, and the
adopted Evidence authorization, lifecycle, projection, Reporting, Recovery,
cleanup, storage, and telemetry contracts agree with their executable
projections and implementation.

Final validation exposed closure defects rather than new product-policy
questions. A Recovery process fixture still assumed attachment implicitly made
Evidence available; it now performs the required explicit lifecycle change.
Four migration characterizations still treated immutable version 29 as the
live catalog head; current count, head, and authored-object allocation now come
from the migration-history projection while version 29 remains the immutable
boundary. The two new table composite types now explicitly revoke PUBLIC
usage. A Reference Data degradation fixture now exercises the authenticated,
CSRF-bound, issued-header upload contract. Static analysis findings in the new
composition, cleanup, policy, and projection paths were removed without
changing public behavior. Exact source hashes and migration-evidence bytes were
then refreshed from the final generated catalog.

The first Incident Bundles owner run encountered a transient external object-
store `operation unavailable` response while creating its fixture. Its exact
row and full owner rerun passed without a code change, so the retained failure
is classified as external infrastructure evidence, not an excluded product
failure. `platform.telemetry` has no service-backed catalog rows; its required
closure is the passing ten-row unit owner suite plus the server/Evidence
service-backed telemetry paths.

| S10 evidence item | Recorded result |
| --- | --- |
| Principal S10 closure files | Database Migrations catalog/remediation/readiness/admission tests; Operator migration-evidence fixture; PostgreSQL test harness hash fixture; Reference Data Evidence-upload workflow; Evidence/Projections cleanup and effect static-analysis corrections; migrations 30-31; generated migration/schema projections through Make |
| Specification and contract reconciliation | Core 01/02/04, Reporting, Recovery, and OpenTelemetry owner projections remain aligned with S01-S09 implementation; no S10 owner amendment was required |
| Migration impact | Final pre-handoff migrations remain additive versions 30 and 31. PUBLIC usage is revoked on both new composite types. Version 29 remains immutable; live head/count are manifest-derived. Any disposable environment that applied an earlier uncommitted draft of 30/31 must be recreated before validation because the final migration hashes changed. |
| Compatibility | No new S10 public route, envelope, schema, token decoder, Store facade, or client behavior. The clean upload-capability cutover and explicit lifecycle transition remain intentional. |
| Deployment handoff | Drain old issuer replicas before enabling the version-2 upload issuer. Apply migrations 30 and 31 before starting the new application; cleanup activates only after readiness. |
| Remaining risk | No unresolved specification or implementation gap. Operational rollout must honor replica drain/migration ordering and retain ordinary object-store/telemetry monitoring. |

| Command | Result | Run root or diagnostic evidence |
| --- | --- | --- |
| `make test-slice OWNER=module.evidence` | PASS, 62/62 | `.cartulary/test-results/20260812T231206Z-p342862` |
| Projections, Imports, Reporting unit owner slices | PASS, 17/17; 40/40; 14/14 | `.cartulary/test-results/20260812T231301Z-p371786`; `.cartulary/test-results/20260812T231314Z-p375770`; `.cartulary/test-results/20260812T231356Z-p403552` |
| Recovery, Incident Bundles, Revisions unit owner slices | PASS, 38/38; 26/26; 56/56 | `.cartulary/test-results/20260812T231623Z-p483732`; `.cartulary/test-results/20260812T231904Z-p527936`; `.cartulary/test-results/20260812T231914Z-p529120` |
| Platform Jobs and Telemetry unit owner slices | PASS, 18/18; 10/10 | `.cartulary/test-results/20260812T232006Z-p564022`; `.cartulary/test-results/20260812T232021Z-p568767` |
| Evidence, Projections, Imports service-backed owner slices | PASS, 47/47; 13/13; 29/29 | `.cartulary/test-results/20260812T232036Z-p570671`; `.cartulary/test-results/20260812T232137Z-p600456`; `.cartulary/test-results/20260812T232206Z-p602648` |
| Reporting, Recovery, Incident Bundles service-backed owner slices | PASS, 7/7; 26/26; 15/15 | `.cartulary/test-results/20260812T232307Z-p628243`; `.cartulary/test-results/20260812T232317Z-p629825`; `.cartulary/test-results/20260812T232347Z-p663307` |
| Revisions and Platform Jobs service-backed owner slices | PASS, 40/40; 14/14 | `.cartulary/test-results/20260812T232357Z-p664490`; `.cartulary/test-results/20260812T232439Z-p689396` |
| Platform Telemetry service-backed selection | NOT APPLICABLE: owner catalog has no service-backed row; unit and server/Evidence service paths pass | Unit root above; server/Evidence roots in §§6.9-6.10 |
| `make frontend-unit`; `make frontend-typecheck` | PASS, 364/364; 2/2 | `.cartulary/test-results/20260812T232456Z-p690758`; `.cartulary/test-results/20260812T232642Z-p729626` |
| `make backend-module-boundary-check` | PASS, 3/3 | `.cartulary/test-results/20260812T232652Z-p730102` |
| `make harness-evidence-contract`; `make harness-contract` | PASS, 2/2 combined graph | `.cartulary/test-results/20260812T232701Z-p730989` |
| `make generate-drift`; `make generated-artifact-policy-check` | PASS, 4/4; 3/3 | `.cartulary/test-results/20260812T235250Z-p1123926`; `.cartulary/test-results/20260812T232712Z-p734187` |
| `make migration-drift`; `make json-shape-check` | PASS, 5/5; 3/3 | `.cartulary/test-results/20260812T235300Z-p1126716`; `.cartulary/test-results/20260812T232721Z-p737328` |
| `make go-gosec-targeted` | PASS, 4/4 | `.cartulary/test-results/20260812T232731Z-p737860` |
| `make browser-e2e-stateful`; `make browser-e2e-webserver-backed` | PASS, 36/36; 62/62 | `.cartulary/test-results/20260812T232748Z-p765768`; `.cartulary/test-results/20260812T232946Z-p792037` |
| Final `make agent-finalize` | PASS, 1/1 | `.cartulary/test-results/20260812T235312Z-p1129575` |
| Final `make test-fast` | PASS, 355/355 | `.cartulary/test-results/20260812T235328Z-p1132301` |
| Final `make check` | PASS, 755/755 | `.cartulary/test-results/20260812T235540Z-p1195713` |
| Final tracker `git diff --check`; `make lint-markdown` | PASS | Diff check completed without findings; `.cartulary/test-results/20260813T000254Z-p1336544` |
| Exact S10 closure reruns | PASS: Database Migrations 5/5 service and 3/3 unit; PostgreSQL ACL 3/3; Reference Data 3/3; Operator 1/1 | `.cartulary/test-results/20260812T235121Z-p1118017`; `.cartulary/test-results/20260812T235234Z-p1122923`; `.cartulary/test-results/20260812T235150Z-p1119800`; `.cartulary/test-results/20260812T235203Z-p1121058`; `.cartulary/test-results/20260812T235240Z-p1123521` |
| Retained S10 diagnostics | Recovery fixture corrected; Incident Bundles external failure passed exact/full rerun; first `test-fast` exposed four stale migration fixtures; first `check` exposed closure defects described above | `.cartulary/test-results/20260812T231408Z-p407970`; `.cartulary/test-results/20260812T231712Z-p519957`; `.cartulary/test-results/20260812T233400Z-p847685`; `.cartulary/test-results/20260812T234130Z-p938571` |

G11 is closed. S00-S10 are complete and the overall effort is `DONE`.

## 7. Gap Ledger and Fixed Decisions

| Gap | Classification | Remediation and durable benefit | Compatibility/migration | Risk if unresolved | Validation criterion |
| --- | --- | --- | --- | --- | --- |
| G01 | CLOSED in S00: documentation/authority | Rebaseline this tracker against adopted owners so execution cannot preserve superseded behavior. | Documentation-only. | Closed; regression would restore contradictory or falsely blocked execution. | Owner and tracker status agree. |
| G02 | CLOSED in S01: implementation/migration/recovery/frontend/tests | Core 04 Tables 2-A through 2-C are enforced and session-bound, accepted-contract-bound, one-shot upload leases are durable. | Same routes/envelopes; clean token cutover; no decoder/backfill; pending legacy slots time out. | Closed; regression would restore unauthorized/cross-session upload, CSRF, enumeration, or state bypass. | Passing owner/service slices, AC-543 cases, frontend tests, and legacy-token rejection recorded in §6.1. |
| G03 | CLOSED in S02: implementation/internal interfaces/tests/harness | One typed source model and canonical row mapper serve provider transaction reads, source mutations, revisions, and projection parity; link mutations refresh exact affected Evidence rows transactionally. | Intentional observable count correction; no shim or wire change. | Closed; regression would reintroduce divergent rows, stale counts, or revision/query drift. | Rich/null/state fixtures, 0/1/2/tombstone differential parity, Evidence/Timeline owner suites, boundaries, and drift pass in §6.2. |
| G04 | CLOSED in S03: internal interface/implementation/tests | Evidence emits authoritative subjects; Projections owns registry selection, before/after loads, exact refresh, storage, and deterministic neutral effects. | Current support effects remain `invalidate`; no wire patch or compatibility facade. | Closed; regression would restore cross-owner layout coupling, unrelated rebuilds, or partial effects. | Exact affected/unaffected subjects, registered views, stable order, omission, and atomic rollback pass in §6.3. |
| G05 | CLOSED in S04: implementation/internal interfaces/tests/harness | Cohesive source mutation, blob lifecycle, access handles, cleanup, and route operations replace the service locator; constructors validate required capability sets. | Intentional internal break; all callers migrated directly; no facade or options shim. | Closed; regression would restore service-locator growth, broad test composition, and accidental route access to cleanup/source mutation. | No Evidence `Store`/`NewStore`/`StoreOption` or legacy option remains; owner/service/boundary/construction evidence passes in §6.4. |
| G06 | CLOSED in S05: internal interfaces/implementation/tests/harness | One validated owner-runtime contract constructs all Evidence contributions, and one private mutation kernel serves ordinary/import creation. | Intentional internal constructor break; consumers migrated directly; no compatibility facade. | Closed; regression would restore import drift, missing effects, initialization panics, or cyclic consumer reconstruction. | Parity, typed construction failures, exact runtime surface, Imports assembly, owner/service suites, boundaries, and drift pass in §6.5. |
| G07 | CLOSED in S06: implementation/frontend/tests/harness | One Evidence-private policy owns lifecycle/blob vocabularies and transitions, bridges, association, reference identity, closed-incident posture, finalize exhaustion, and failure scheduling across ordinary/import/rollback/route/cleanup entry points. | Attachment no longer auto-promotes lifecycle; clients use an explicit versioned transition; no policy or promotion shim. No migration. | Closed; regression would admit divergent lifecycle/reference state or miss the one-hour cleanup bound. | Shared matrices, rollback/import parity, owner/service, browser/a11y/visual/stateful, boundary, and drift evidence pass in §6.6. |
| G08 | CLOSED in S07: implementation/internal interfaces/tests/harness | Evidence uses closed twelve-member Reporting construction and contributes logical support targets; Reporting coordinates generic Links resolution; all five owner suites are registered. | Accidental extra fields disappear intentionally; no route/schema or migration change. | Closed; regression would restore future-field disclosure, physical/cross-owner coupling, or unverified provider drift. | Exact keys/nulls/times, unknown/forbidden absence through export, deleted exclusion, logical identity, malformed-output atomicity, provider suites, boundary, and drift evidence pass in §6.7. |
| G09 | CLOSED in S08: specification projection/migration/recovery/implementation/tests | Transient durable cleanup claims, bounded leases/retries/rechecks, and seven-day metadata retention are implemented and routed. | Additive schema; existing eligible orphans enter policy and the validated engine is activated by S09. | Closed; regression could orphan, duplicate, race, or miss cleanup deadlines. | Exact/full Evidence service, Recovery inventory, contribution, migration, generation, and shape evidence passes in §6.8. |
| G10 | CLOSED in S09: specification/implementation/assembly/telemetry/tests | Private readiness-gated 15-minute dispatch, typed deletion, orderly shutdown, and closed safe telemetry activate the durable engine without a Jobs surface. | Deletes only eligible failed unattached bytes; no public route/schema or migration change in S09. | Closed; regression would leave the engine inert, race shutdown, or expose ungoverned storage/telemetry behavior. | Exact/full Evidence and server, platform telemetry/object-store, process lifecycle/projection, boundary, generation, migration, and shape evidence passes in §6.9. |
| G11 | CLOSED in S10: tests/generated projections/docs/handoff | Reconciled every owner, drift, browser, security, harness, fast, and full-repository gate with retained diagnostic and success evidence. | No public compatibility change; final pre-handoff migration hashes and exact fixtures reflect the completed additive schema. | Closed; regression would restore hidden drift, stale catalog assumptions, unsafe ACL defaults, or unauditable completion. | Owner-first matrix, frontend/browser, harness, security, drift, finalize, 355/355 fast, and 755/755 full checks pass in §6.10. |

Fixed decisions:

- Public Evidence HTTP routes and response schemas remain unchanged.
- Upload token version cutover is clean: no legacy decoder, lease backfill, or
  same-slot retry after a claim begins.
- New database objects use next-available authored migrations; historical
  migrations are not edited.
- Upload leases are `security_ephemeral`/`excluded_security_state`/`invalidate`.
- Cleanup claims are `transient`/`excluded_transient`/`invalidate`.
- Generated projections are changed only through authored inputs and Make.
- Provider packages remain unless executable evidence proves ownership inversion.
- `QuarantineBlob` remains separate and cleanup never invokes it.
- Object Blob association never promotes Evidence lifecycle implicitly; handle
  eligibility requires an explicit owner lifecycle transition.

### 7.1 Projection-effects characterization

The S03 association-effects integration row freezes exact support record IDs,
versions, registered view-schema set, field keys, invalidation shape, event
order, unregistered omission, unrelated-row preservation, and rollback. The
sets are exact, sorted, unique, and all-or-nothing. Provider failure remains a
caller-transaction error; no partial effect is published.

### 7.2 Provider behavioral suites

Consumer owners retain behavioral-test ownership; Evidence supplies fixtures,
contributions, and fault hooks.

| Suite | Owner | Required cases |
| --- | --- | --- |
| `reporting.evidence_provider_redaction` | Reporting | Exact twelve-member allowlist; forbidden values absent at every depth; deterministic; new field omitted; deleted/ineligible excluded; logical support refs; malformed contribution rejected without partial output. |
| `imports.evidence_create_parity` | Imports | Ordinary/import equality for source, envelope, create signal, defaults, times, parties, ref admission, provenance, history, projection, intents, result, row, first success/replay; failure after envelope, source, initial blob, projection, revision, apply journal, and unit outcome rolls back; post-commit restart returns one result. |
| `incident_bundles.evidence_blob_staging_cleanup` | Incident Bundles | Deterministic files/order/digest paths; missing, path/byte digest, size, duplicate, cross-incident failures; partial staging, cancel before/after, prepare, commit, restart residue, cleanup retry, pre-existing final, success/no residue; no locator leak; never delete committed final. |
| `recovery.evidence_inventory_closure` | Recovery | Exact relation set once; no projection/session/handle/upload/cleanup claim; objects once with incident/logical ID/size/digest/ref; missing/extra/duplicate/changed/unparseable fail; restore blocks readiness on mismatch; successful same-set restore/rebuild/lifecycle/version/count/hash; partial target not ready. |
| `module.evidence.provider_contribution_closure` | Evidence | Stable IDs, exact constructors, typed nil rejection, deterministic descriptors, exact relations/paths, no consumer orchestration, no generated/descriptor SQL, no physical projection access, all-or-nothing failure. |

All five suites are `DONE` in §6.7. Their authored catalog rows are the
execution identities; the dotted names above remain stable semantic suite
names for this handoff and future maintenance.

### 7.3 S08/S09 private durable-claim cleanup

Cleanup is private maintenance. It MUST create no public job resource, route,
Workbook row, Collaboration event, revision, or incident audit entry. Its
stable private identity is `evidence.failed_unattached_blob_cleanup.v1`.

| Parameter/concern | Fixed requirement |
| --- | --- |
| Ownership | Evidence owns candidates/state/recheck/delete request/completion and its private periodic dispatcher; application assembly activates after readiness; the object-store adapter deletes; the telemetry boundary instruments. |
| First sweep | Immediately after serving readiness. |
| Interval/configuration | Fixed 15 minutes; no new deployment configuration. |
| Batch/order | 100, ordered `cleanup_due_at ASC, object_blob_id ASC`; drain further batches while eligible work remains. |
| Failed due time | `cleanup_due_at = failed_at + 45 minutes`. |
| Runtime relation | `evidence_blob_cleanup_claims`, excluded from Recovery authoritative state. |
| Claim fields | Blob identity, unique claim token, claimed/expiry timestamps, attempt count, next-attempt timestamp. |
| Claim lease | 5 minutes. |
| Deletion deadline | 1 minute. |
| Retry | 1, 5, then 15 minutes; continue at 15 minutes without terminal exhaustion. |
| Retention | Hard-delete only after metadata has remained queryable at least seven days. |
| Quarantine | `QuarantineBlob` remains dormant and is never invoked by cleanup. |

Pending rows past `pending_expires_at` first transition atomically to
`failed/pending_timeout` with failure time and due time 45 minutes later.
Cleanup eligibility requires failed state, null `cleaned_up_at`, due time
reached, no Evidence association, and no live competing claim. Pending,
available, quarantined, attached, portability-final, and cleaned rows are
ineligible.

Candidates MUST be claimed transactionally with
`FOR UPDATE SKIP LOCKED`; this construct is permitted only for queue-like
claiming, never inventory or lifecycle proof. Object deletion occurs outside
the claim transaction. Typed object-not-found is idempotent success. A short
completion transaction MUST revalidate claim token plus failed/unattached state
before `cleaned_up_at`.

Association, finalization, quarantine, replacement, import, and restore MUST
not proceed while an unexpired claim exists. Shutdown stops claims, drains or
cancels bounded calls, writes no false completion, and leaves restart-eligible
work. Restart recovers stale claims. There is no terminal retry exhaustion.

Safe telemetry is limited to sweep state, candidates, cleaned/already-absent,
state-changed skips, closed deletion class, completion failure, overdue count
and oldest age, and duration. It MUST exclude incident/Evidence/blob IDs,
storage refs/keys, filenames, hashes, raw errors, credentials, and content.

| Service-backed case | Required result |
| --- | --- |
| Startup overdue | First post-readiness sweep cleans. |
| Cadence | No controlled-clock interval exceeds 15 minutes. |
| Batch drain | More than 100 rows drain stably within deadline. |
| Pending timeout | Failed by expiry plus 15 minutes; due equals failure plus 45. |
| One-hour deadline | Every admitted failed/unattached object cleaned by failure plus one hour. |
| Attached failed/inconsistent | No deletion. |
| Available/quarantined | No deletion. |
| Claim/lifecycle conflict | Associate/finalize/quarantine/replace/import/restore cannot race. |
| Competing worker | One effective claim; duplicate delete harmless. |
| Restart stale claim | Recovered and completed once. |
| Transient object error | No completion; retry policy within deadline. |
| Typed not found | Idempotent completion. |
| One-minute delete timeout | Canceled; no false completion. |
| Completion DB failure | Retry sees absent object and completes. |
| State change before delete | Skip; no object call. |
| State change before completion | Reject stale completion. |
| Shutdown during claim | Recoverable after drain/cancel. |
| Retention | No hard delete; seven-day queryability. |
| Recovery | Claim relation absent from authoritative sets. |
| Telemetry | Required safe facts present; forbidden values absent. |
| Quarantine guard | Cleanup never calls `QuarantineBlob`. |
| Public-effect guard | No job/route/row/event/revision/audit effect. |

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make test-slice OWNER=module.evidence` | Evidence fast rows | yes for S00 through S07 | Baseline first. |
| integration | `make service-backed-test-slice OWNER=module.evidence` | Evidence service-backed | yes | Per slice and S07. |
| consumer integration | `make test-slice OWNER=<owner>` and `make service-backed-test-slice OWNER=<owner>` | Projections/Imports/Reporting/Recovery/Incident Bundles/Revisions/jobs/telemetry | per affected slice | Resolve exact rows through `make task-guide`. |
| e2e/browser | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | Observable consumers | conditional | Only if affected. |
| generated drift | `make generate-drift`; `make generated-artifact-policy-check` | Generated policy | conditional | Never hand-edit. |
| import/static | `make backend-module-boundary-check` | Evidence/Projections | yes for structural slices | Especially S02-S06. |
| harness | `make harness-contract` | Accounting | conditional | Only if routing changes. |
| full | `make test-fast`; `make check` | Repository | final implementation | Run `make agent-finalize` before broad implementation gates. |
| docs | `git diff --check`; `make lint-markdown` | Tracker and owner docs | every slice | Tracker checkpoint precedes the next slice. |

S10 additionally runs frontend unit/typecheck, module boundaries,
`harness-evidence-contract`, `harness-contract`, generated and migration drift,
JSON shape checks, targeted gosec, stateful and webserver-backed browser suites,
`agent-finalize`, `test-fast`, and `check` in the order recorded by the
authorized plan.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| TR-001 | Authority/scope | S00 | DONE | None | Sections 1/6/7 | Adopted/current baseline and implementation authority recorded. |
| TR-002 | 70-file current inventory | WF-01 | DONE | TR-001 | Section 2 | Every file accounted. |
| TR-003 | Boundary diagnosis | WF-02/WF-04 | DONE | TR-002 | Sections 3/5 | Legitimate/mixed/misplaced distinguished. |
| TR-004 | Core 04 matrix text | WF-02 | DONE | TR-001 | Core 04 §2.0A/AC-543 | Exact authority authored. |
| TR-005 | Core 01 bindings/xrefs | WF-02 | DONE | TR-004 | Core 01 | No duplicate matrix. |
| TR-006 | Reporting allowlist | WF-02 | DONE | TR-001 | Reporting §7.1.1 | Exact allowlist/default authored. |
| TR-007 | Adopt owner amendments | S00 | DONE | TR-004-TR-006 | Adopted/current owner front matter | Baseline current. |
| TR-008 | S00 rebaseline | S00 | DONE | TR-007 | Sections 1/6/7 and handoff log | Tracker authoritative and linted. |
| TR-009 | S01 authorization/upload leases | S01 | DONE | TR-008 | §6.1 AC-543, lease, frontend, migration, and drift evidence | Security matrix passes. |
| TR-010 | S02 canonical mapper | S02 | DONE | TR-009 | §6.2 mapper/link-count parity evidence | One correct mapper and current linked count. |
| TR-011 | S03 projection effects | S03 | DONE | TR-010 | §6.3 atomic effects and boundary evidence | Projections owns support mechanics. |
| TR-012 | S04 Store decomposition | S04 | DONE | TR-011 | §6.4 cohesive-service, caller, construction, owner, and boundary evidence | No broad Store. |
| TR-013 | S05 composition | S05 | DONE | TR-012 | §6.5 validated composition and import parity | One owner runtime. |
| TR-014 | S06 policy | S06 | DONE | TR-013 | §6.6 shared lifecycle/reference/failure matrices and cross-entry evidence | One policy. |
| TR-015 | S07 providers/Reporting | S07 | DONE | TR-014 | §6.7 exact allowlist, logical-target contribution, five provider suites, boundary, and drift evidence | Exact allowlist/closure. |
| TR-016 | S08 cleanup engine | S08 | DONE | TR-015 | §6.8 exact/full Evidence, Recovery inventory, migration, generation, and shape evidence | Durable engine is validated and remains inert pending S09. |
| TR-017 | S09 cleanup activation | S09 | DONE | TR-016 | §6.9 dispatcher/storage/telemetry/server/process evidence | Active safe cleanup. |
| TR-019 | S10 validation/handoff | S10 | DONE | TR-017 | §6.10 and handoff log | Full closure recorded. |
| TR-018 | Frontend/grid move | WF-05 | DROPPED | TR-003 | Repository evidence | Reopen only with new authority/evidence. |

## 10. Session Handoff Log

The initial planning rows are preserved session history. Every later session MUST append and MUST NOT overwrite an earlier row except to remove a demonstrated duplicate. The initial session touched only this tracker. The NLSpec-grade revision session touched only the four documentation files authorized in Section 1.

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-11T06:38:40-04:00 | Codex / initial planning session | Planning inventory complete; implementation unauthorized | Inspected framework, Domain, Core 00-04, Reporting and Testing Harness NLSpecs; touched only this tracker | `sed` owner/framework reads; `rg --files docs`; `rg -n` authority searches; `git status --short --branch` | Target and authority posture confirmed; no Evidence NLSpec and no owner contradiction found | None for tracker completion | A future authorized agent starts at WF-00 and rechecks the commit/owners. |
| 2026-08-11T18:27:01-04:00 | Codex / NLSpec-grade revision | Design and owner authority closed in documentation; implementation unauthorized | Inspected analysis notes, NLSpec guide, tracker, Core 01, Core 04, Reporting NLSpec, and targeted live repository evidence; touched four authorized docs | Targeted `sed`/`rg`; `git status --short --branch`; documentation patching | Core 04 owns the matrix, Core 01 owns route/capability behavior, and Reporting owns the provider allowlist | Owner amendments require adoption into the implementation baseline and S-00 evidence | Adopt the amendments, then authorize S-00 separately. |
| 2026-08-11T18:55:55-04:00 | Codex / S00 implementation | Authority and tracker rebaselined; S00 complete | Rechecked adopted Core/Reporting owners, worktree, tracker, and implementation hotspots; changed tracker | `git status`; targeted `sed`/`rg`; `make lint-markdown` | Adoption blocker removed; G01-G11 and S00-S10 are controlling | None | Begin S01 only after this checkpoint passes lint. |
| 2026-08-11T19:42:54-04:00 | Codex / S01 implementation | Authorization and upload-capability correction complete | Inspected and changed the files recorded in §6.1; generated projections changed only through Make | §6.1 command matrix | G02 closed with clean version-2 cutover and passing retained evidence | None | Begin S02 canonical row mapping after Markdown lint. |
| 2026-08-11T20:09:18-04:00 | Codex / S02 implementation | Canonical row and linked-count remediation complete | Inspected Evidence Store/source/projection paths plus Timeline link mutation and assembly paths; changed files recorded in §6.2 | §6.2 command matrix | G03 closed; S00-S02 are complete | None | Begin S03 projection-effects ownership after Markdown lint. |
| 2026-08-11T20:38:41-04:00 | Codex / S03 implementation | Registry-owned support projection effects complete | Inspected and changed the Evidence/Projections/Collaboration boundary, private projection storage, assembly/test capabilities, authored test routing, and residual migration validator recorded in §6.3 | §6.3 command matrix | G04 and RB-003 closed; S00-S03 are complete | None | Begin S04 Store decomposition after Markdown lint. |
| 2026-08-11T20:55:59-04:00 | Codex / S04 implementation | Cohesive Evidence owner capabilities complete | Inspected and changed every former Store constructor/receiver/caller plus owner runtime, Workbook/import facades, test support, server characterization, and authored Evidence routing recorded in §6.4 | §6.4 command matrix and repository-wide symbol searches | G05 closed; S00-S04 are complete with no compatibility facade | None | Begin S05 owner composition only after Markdown lint. |
| 2026-08-11T21:21:27-04:00 | Codex / S05 implementation | Validated Evidence owner composition and shared create kernel complete | Changed owner runtime, Workbook/import construction, mutation kernel, server/test support, Imports registry/tests, and authored routing recorded in §6.5 | §6.5 command matrix and repository-wide constructor/import searches | G06 closed; S00-S05 are complete with no consumer reconstruction or panic-based owner construction | None | Begin S06 policy consolidation only after Markdown lint. |
| 2026-08-11T22:07:42-04:00 | Codex / S06 implementation | Shared lifecycle/reference/failure policy complete | Changed the Evidence-private policy package, every policy caller, explicit browser transition, support-effect lifecycle refresh, fixtures, and authored routing recorded in §6.6 | §6.6 command matrix and repository-wide duplicate-policy searches | G07 closed; S00-S06 are complete with one owner policy and no implicit attachment promotion | None | Begin S07 provider/Reporting closure only after Markdown lint. |
| 2026-08-11T22:36:03-04:00 | Codex / S07 implementation | Closed Reporting provider and five consumer-owner suites complete | Changed Evidence/Links/Reporting provider composition, exact tests, Records portability cursor lifetime, Incident Bundles diagnostics, Recovery inventory evidence, authored catalogs, generated topology, and this tracker | §6.7 command matrix and closed-construction/cross-owner searches | G08 and RB-005 closed; S00-S07 are complete with no subtractive Evidence Reporting value or Links-to-Records join | None | Begin S08 durable cleanup engine only after Markdown lint. |
| 2026-08-11T23:10:21-04:00 | Codex / S08 implementation | S08 implementation present but exit gate blocked; S09 unstarted | Changed Core 01/02/04 projections, migration/recovery/schema-owner contracts, Evidence cleanup engine/tests, authored routing, generated outputs through Make, and this tracker | §6.8 command matrix | No owner contradiction; serial stop rule applies because required service evidence never started | Shared Docker inspect/readiness timeout repeated three times | Restore disposable-container readiness, rerun the exact cleanup row and full Evidence service slice, then close S08 before S09. |
| 2026-08-12T19:08:22-04:00 | Codex / S09 implementation | Private cleanup activation and governed telemetry complete | Changed Core 02/04 and OpenTelemetry owners; Evidence engine/dispatcher/storage adapter; server lifecycle/observer; platform storage/telemetry contracts; process fixture; authored routing; generated topology through Make; and this tracker | §6.9 command matrix | G10 closed; S00-S09 are complete with no public Jobs surface or unsafe telemetry vocabulary | None | Begin S10 final validation only after Markdown lint. |
| 2026-08-12T20:00:26-04:00 | Codex / S10 validation and handoff | Overall remediation `DONE` | Reconciled all adopted owners, executable projections, migrations, implementation, authored routing, generated artifacts, validation evidence, and this tracker | §6.10 owner-first and full command matrix | G01-G11 and S00-S10 closed; no stale blocker or owner contradiction remains | None | Deploy only with the recorded migration ordering and old-replica drain. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-11T06:38:40-04:00 | Codex / initial planning session | Legitimate Evidence owner; mixed root package; future slices unstarted | Inspected all 58 target files, app assemblies, importers, migrations, and `tools/backend_module_boundaries.json`; touched only tracker | `find`/`rg --files`; targeted `sed`; `rg -n` callers/imports/symbols/cleanup uses | Store/row/projection-effect hotspots and intentional provider seams mapped | RB-001 through RB-003 block affected implementation slices | Execute S-00 characterization after later authorization. |
| 2026-08-11T18:27:01-04:00 | Codex / NLSpec-grade revision | S-01 through S-06 are fully specified future behavior-preserving slices | Inspected Evidence runtime, Store, Rows port, support-refresh SQL, Projections adapters, assembly, and providers; touched authorized docs only | Targeted `sed`/`rg` for live types, callers, SQL, and constructors | Canonical-row oracle and `SupportProjectionEffectsTx` interface fixed; no shim allowed | S-00 has not run | Run S-00; permit S-01 only after zero mismatches. |
| 2026-08-11T18:55:55-04:00 | Codex / S00 implementation | Serial implementation authorized | Reconciled Evidence inventory and caller findings with the remediation plan | Targeted source and caller searches | Existing inventory remains the baseline; new route, row, projection, Store, policy, provider, and cleanup gaps recorded | S01 validation remains | Implement S01 security before structural work. |
| 2026-08-11T19:42:54-04:00 | Codex / S01 implementation | Route pipeline and durable one-shot upload lease implemented | Changed Evidence admission/routes/API/token/store/repositories and added attach preflight plus lease persistence | Evidence owner and service-backed slices; focused AC-543/legacy rows | No denied upload/attach path reaches storage before current authorization, lifecycle, contract, version, and replay gates | Broad Store remains intentionally for S04 | S02 replaces duplicate row mapping without changing S01 gates. |
| 2026-08-11T20:09:18-04:00 | Codex / S02 implementation | One Evidence row mapping boundary remains | Moved serialization to `workbookprojection.ViewRow`; routed provider and Store transaction loads through typed source input; added exact affected-row refresh for Timeline attachment links | Evidence/Timeline owner and service-backed slices; module-boundary check | Store duplicate mapper and hard-coded linked count removed; link add/remove refreshes exact Evidence rows atomically | Broad Store and support-view mechanics remain assigned to S04 and S03 respectively | Implement S03 without reintroducing Evidence view knowledge. |
| 2026-08-11T20:38:41-04:00 | Codex / S03 implementation | Evidence publishes subjects; Projections owns all support mechanics | Removed support refresh from primary Rows, added neutral effects port, registry selection, exact refresh/state reads, and private storage access; adapted deterministic Collaboration arrays | Evidence/Projections/Collaboration suites and module-boundary check | No Evidence production code names a support view/field/table; no whole-incident support rebuild remains; rollback is exact | Broad Store remains intentionally assigned to S04 | Decompose Store into cohesive owner capabilities without restoring a facade. |
| 2026-08-11T20:55:59-04:00 | Codex / S04 implementation | No broad Evidence owner root remains | Replaced Store construction/receivers with source-mutation, blob-lifecycle, access-handle, cleanup, and route capabilities; migrated server and focused tests directly | Evidence owner/service suites, exact construction/runtime rows, repository symbol search, module-boundary check | Routes cannot reach cleanup/source mutation; every required dependency fails closed; no old constructor, option, type, or facade remains | S05 still owns the broad panic-based Workbook/Import composition lists | Introduce validated `OwnerRuntimeDependencies` and one create kernel in S05. |
| 2026-08-11T21:21:27-04:00 | Codex / S05 implementation | Evidence alone composes its owner runtime and contributions | Added validated dependencies, four narrow runtime capabilities, and one private ordinary/import mutation kernel; removed consumer projection wiring and public construction | Evidence/Imports owner suites, exact construction/assembly rows, parity, module-boundary check | Missing dependencies return errors; Imports cannot import Evidence implementation/projection packages or reconstruct owner state | Lifecycle/reference rules remain G07/S06 | Move all owner policy interpretations behind one Evidence-private package. |
| 2026-08-11T22:07:42-04:00 | Codex / S06 implementation | Evidence owns one private lifecycle/reference/failure policy | Added policy package and migrated ordinary/import/rollback/association/finalize/quarantine/route/timeout callers; `blobref` is syntax-only | Shared policy/rollback rows, Evidence and Imports suites, module-boundary check | Duplicate switches/errors removed; support effects remain Projections-owned and atomic on explicit lifecycle patches | Cleanup claims do not yet exist | Preserve the policy boundary while adding Reporting/provider closure in S07. |
| 2026-08-11T22:36:03-04:00 | Codex / S07 implementation | Evidence owns redacted facts and logical target identity; Reporting coordinates generic relationships | Replaced subtractive Evidence JSON, added logical-target contribution, removed provisional Links-to-Records read, and retained provider package seams | Reporting and five named suites; Records portability; module-boundary check | Exact allowlist and owner boundaries pass; provider packages remain because executable evidence confirms their cohesion | Cleanup claims do not yet exist | Preserve closed output and owner-contributed targets while implementing S08 claims. |
| 2026-08-11T23:10:21-04:00 | Codex / S08 implementation | Durable cleanup coordination is owner-private and production-inert | Added narrow cleanup sweep capability, transactional claim/complete/retry/retention paths, and database race guards; removed direct candidate deletion | Build, unit compile, Recovery catalog units, boundary, drift, and exact service attempts | Structural and non-service gates pass; no server assembly or public capability exposes cleanup | Service-backed engine semantics remain unexecuted because suite startup failed externally | Do not wire a dispatcher; rerun §6.8 service gates after Docker readiness returns. |
| 2026-08-12T19:08:22-04:00 | Codex / S09 implementation | Cleanup is active behind a narrow private runtime boundary | Added Evidence dispatcher and typed delete adapter, composed them after bootstrap, activated after publication readiness, and registered reverse-order shutdown; added no common Jobs dependency | Exact/full Evidence/server/platform suites, process flow, build, boundaries, and drift gates in §6.9 | One serialized dispatcher per process plus durable claims provide multi-instance safety; typed storage and closed observer isolate dependencies | None | Preserve this capability boundary through S10 broad validation. |
| 2026-08-12T20:00:26-04:00 | Codex / S10 validation and handoff | Backend boundaries and cross-owner behavior fully validated | Corrected manifest-driven migration characterizations, final ACLs, stale owner fixtures, and static-analysis findings; otherwise inspected the complete S01-S09 backend delta | Owner/service slices, boundary, finalize, 355/355 fast, and 755/755 full graphs in §6.10 | No broad Store, consumer-owned composition, support-view coupling, duplicate policy/mapper, or inert cleanup path remains | None | Preserve owner-first suites and the full graph as regression evidence. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-11T06:38:40-04:00 | Codex / initial planning session | Frontend is an external consumer, not part of the target move | Inspected Evidence service, Workbook bindings, lifecycle/access presentation, Timeline attachment adapter, and browser Evidence fixtures/uploads; touched only tracker | Targeted `sed` and `rg -n` for Evidence fields/actions/selectors/vendor imports | No backend grid-vendor dependency and no supported frontend relocation found | None unless a later slice changes wire/event/schema behavior | Run relevant frontend/browser rows only when a backend seam affects them. |
| 2026-08-11T18:27:01-04:00 | Codex / NLSpec-grade revision | Frontend remains an external consumer; no frontend slice planned | Reused and reviewed initial frontend evidence; touched authorized docs only | Tracker and owner review | No Evidence ownership transfer or grid-vendor work authorized | None for structural planning | Run browser consumers only when a later backend slice can affect them. |
| 2026-08-11T18:55:55-04:00 | Codex / S00 implementation | Frontend is in scope for S01 transport correction only | Inspected Workbook Evidence upload service and browser upload helpers | Targeted `rg`/`sed` | Current upload omits credentials and retries; S01 must add CSRF/session transport and remove retry | S01 tests | Preserve public upload-target shape. |
| 2026-08-11T19:42:54-04:00 | Codex / S01 implementation | Browser upload transport conforms to the current-session route | Changed Workbook Evidence upload service/tests, Timeline/Workbook consumers, transport boundary, and E2E upload helper/tests | Frontend unit/typecheck and affected owner rows in §6.1 | Credentials included, current CSRF required, issued headers preserved, retry removed | None | No frontend work is expected in S02 unless row fixtures expose a consumer mismatch. |
| 2026-08-11T20:09:18-04:00 | Codex / S02 implementation | No frontend change required | Compared HTTP mutation and query rows through existing JSON consumers | Service-backed differential fixture | Public row shape is unchanged; corrected linked counts arrive through the existing cell | None | S03 remains internal unless neutral effects expose an existing consumer defect. |
| 2026-08-11T20:38:41-04:00 | Codex / S03 implementation | No frontend source change required | Exercised real attach stream and Collaboration stateful consumer | Evidence integration and Collaboration owner rows | Existing single-view invalidations and record-removal events remain wire-compatible | None | S04 is internal; rerun frontend only if decomposition changes an observable seam. |
| 2026-08-11T20:55:59-04:00 | Codex / S04 implementation | Frontend remains unchanged | Verified decomposition below the existing route/Workbook interfaces | Full Evidence owner and service-backed matrices | HTTP and event behavior remains covered and unchanged | None | S05 remains internal unless parity evidence exposes an observable defect. |
| 2026-08-11T21:21:27-04:00 | Codex / S05 implementation | Frontend remains unchanged | Verified ordinary/import parity below stable public HTTP and Workbook interfaces | Evidence and Imports owner/service-backed matrices | No route, schema, event, or frontend dependency changed | None | S06 remains owner-private unless policy evidence exposes an existing wire defect. |
| 2026-08-11T22:07:42-04:00 | Codex / S06 implementation | Browser attachment uses the explicit Evidence lifecycle transition | Updated mutation commands, Workbook fixtures, shared/direct E2E upload fixtures, and visual readiness timing | Frontend unit/typecheck plus browser, stateful, accessibility, and visual owner rows | Public route/schema shape remains stable; attach preserves lifecycle and a versioned patch establishes `available` | None | S07 frontend work is unnecessary unless Reporting/provider evidence identifies a consumer defect. |
| 2026-08-11T22:36:03-04:00 | Codex / S07 implementation | Frontend remains unchanged | Verified Reporting/provider closure below stable public Evidence and Reporting routes | Reporting exact suite and full owner slice | No wire, browser, or frontend contract changed | None | S08 remains backend/recovery work; frontend changes require separately demonstrated impact. |
| 2026-08-11T23:10:21-04:00 | Codex / S08 implementation | Frontend remains unchanged; no cleanup surface exists | Verified the engine is absent from server assembly and retains existing public routes and schemas | `make build-server`; repository assembly search | Partial deployment cannot initiate cleanup and no client compatibility path changed | S08 service gate blocked externally | Resume backend validation only; frontend remains out of S08 unless live evidence reveals an observable defect. |
| 2026-08-12T19:08:22-04:00 | Codex / S09 implementation | No frontend cleanup surface or schema change | Corrected the retained process helper to use the S01 current-session/CSRF one-shot upload contract and the S06 explicit lifecycle transition | Exact process row and full app-server owner/service slices | Real client behavior remains unchanged; the stale process fixture now exercises the normative flow through Timeline projections | None | Run the complete frontend/browser matrix in S10. |
| 2026-08-12T20:00:26-04:00 | Codex / S10 validation and handoff | Frontend and browser closure complete | Revalidated Workbook transport, explicit lifecycle consumers, process fixtures, stateful browser behavior, and webserver-backed flows; no new frontend production change in S10 | Frontend 364/364, typecheck 2/2, stateful 36/36, webserver-backed 62/62 | Issued headers, same-session credentials, CSRF, no upload retry, and existing public schemas remain intact | None | No follow-up frontend migration is required. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-11T06:38:40-04:00 | Codex / initial planning session | Six HTTP operations and typed view/provider surfaces frozen; no contract edit planned | Inspected Evidence OpenAPI source, view schema, errors, projection/import/incident/recovery catalogs, and generated projections read-only; touched only tracker | `jq` OpenAPI path/operation extraction; `rg -n` contract/generated references | Structural plan requires no generated or authored contract change | A discovered wire/schema change would stop the structural slice | Use Make generators/drift only after separately justified authored input changes. |
| 2026-08-11T18:27:01-04:00 | Codex / NLSpec-grade revision | Authorization and Reporting authority gaps closed in owner text; generated artifacts untouched | Inspected Core 01 routes/handles, Core 04 authorization, Reporting boundary, and live provider; touched three owner docs plus tracker | Targeted reads and documentation patching | No route, role, token, error vocabulary, public resource, contract artifact, or generated file added | Adoption and executable characterization remain future gates | Adopt owners and run S-00; never hand-edit generated projections. |
| 2026-08-11T18:55:55-04:00 | Codex / S00 implementation | Additive migrations and authored recovery inputs authorized | Rechecked generated-artifact policy, migration posture, and recovery classifications | Targeted owner/tool reads | Lease and claim classifications fixed; generated outputs remain Make-owned | Per-slice drift gates | Add next-available migrations only in S01/S08. |
| 2026-08-11T19:42:54-04:00 | Codex / S01 implementation | Migration 30 and Recovery projection current | Added authored migration/recovery/catalog/schema inputs; generator now admits contiguous post-baseline migrations with explicit owners | `make generate`; migration/generate drift; generated-artifact policy check | Version 30 projected; immutable boundary stays 29; Recovery catalog 112/83 | Deployment replica drain is operational handoff | S02 should not change generated contracts unless the canonical mapper exposes an owner-input defect. |
| 2026-08-11T20:09:18-04:00 | Codex / S02 implementation | No product contract or migration change | Added one authored Evidence test-family row and regenerated topology through Make | `make generate`; `make generate-drift` | Generated topology is current; no generated root was hand-edited | None | S03 may change only internal interfaces unless owner evidence requires more. |
| 2026-08-11T20:38:41-04:00 | Codex / S03 implementation | Internal effect contract and generated topology current | Added the Evidence/Projections internal port and activated the Collaboration intent fixture in authored routing; made schema ownership validation consume migration history | `make generate`; `make generate-drift`; `make json-shape-check` | No public schema or migration change; migration 30 validation is forward-compatible by filename allocation | None | S04 should not alter contracts unless decomposition exposes a genuine owner-input defect. |
| 2026-08-11T20:55:59-04:00 | Codex / S04 implementation | Internal construction contract and topology current | Extended an authored Evidence unit row with cohesive-constructor failure cases; regenerated topology through Make | `make generate`; `make generate-drift`; `make json-shape-check` | No public contract, migration, Recovery input, or generated root was manually changed | None | S05 may change only owner composition interfaces and authored verification routing. |
| 2026-08-11T21:21:27-04:00 | Codex / S05 implementation | Owner composition and generated topology current | Updated authored Evidence, Imports, and Assessments selectors; regenerated the execution-topology index through Make | `make generate`; `make generate-drift`; `make json-shape-check` | No public contract, migration, or Recovery input changed; no generated output was hand-edited | None | S06 changes only Evidence-private policy and authored tests unless an owner projection defect is demonstrated. |
| 2026-08-11T22:07:42-04:00 | Codex / S06 implementation | Policy test routing and generated topology current | Added two authored Evidence selectors and regenerated topology only through Make | `make generate`; `make generate-drift`; `make json-shape-check` | No product contract, migration, Recovery input, or hand-edited generated root in S06 | None | S07 must register provider suites through authored catalog inputs. |
| 2026-08-11T22:36:03-04:00 | Codex / S07 implementation | Five provider-suite routes and generated topology current | Updated authored Reporting, Imports, Incident Bundles, Recovery, and Evidence family inputs plus Incident Bundle fixture names; regenerated through Make | `make generate`; `make generate-drift`; `make json-shape-check` | No migration, public contract, or hand-edited generated root; all named suites resolve through authored catalog rows | None | S08 adds only the next migration and corresponding authored Recovery/schema inputs before regeneration. |
| 2026-08-11T23:10:21-04:00 | Codex / S08 implementation | Migration 31 and 113/83 Recovery/schema projections are current | Added authored migration, Core/Recovery inputs, schema-owner validators/schema, and Evidence family row; regenerated tool-owned outputs through Make | `make generate`; `make generate-drift`; `make migration-drift`; `make json-shape-check` | All generation, migration, Recovery unit, and shape gates pass; no historical migration or generated root was hand-edited | Live migration/service matrix blocked before execution by Docker readiness | Retain additive migration while blocked; rerun exact live rows before S09. |
| 2026-08-12T19:08:22-04:00 | Codex / S09 implementation | Cleanup execution and telemetry owner projections are current | Amended Core 02/04 and the adopted OpenTelemetry owner; added three authored family rows; regenerated topology through Make | `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make migration-drift`; `make json-shape-check` | No migration or public contract changed in S09; all generated/drift gates pass and no generated root was hand-edited | None | Preserve authored-input ownership while running S10 drift and harness closure. |
| 2026-08-12T20:00:26-04:00 | Codex / S10 validation and handoff | Final migration and generated projections agree | Hardened the two new composite-type ACLs, made live catalog facts manifest-driven, refreshed migration/schema projections through Make, and updated exact catalog fixtures | `make generate`; generate/artifact/migration/JSON drift gates; full `check` | Additive migrations 30-31 are final, immutable boundary 29 is preserved, Recovery cardinality is 113, and no generated root was hand-edited | None | Recreate any disposable database that applied an earlier uncommitted draft of migrations 30-31. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-11T06:38:40-04:00 | Codex / initial planning session | Current owner topology mapped: 45 rows, 34 service-backed; tests not executed in planning | Inspected all target tests, Evidence verification owner/family, test support, and browser fixtures; touched only tracker | `make help-all`; `make task-guide ROLE=module-author OWNER=module.evidence`; `make explain-test-owner OWNER=module.evidence`; targeted source reads | Canonical narrow commands discovered; characterization gaps named | Cleanup job owner command still requires live discovery | In S-00, select exact live row IDs and establish a passing baseline. |
| 2026-08-11T18:27:01-04:00 | Codex / NLSpec-grade revision | Five provider suites and two differential suites specified; no tests or harness inputs changed | Inspected analysis recommendations and owner command posture; touched authorized docs only | Targeted reads; no product test, generator, or harness command | Acceptance matrices cover success, failure, rollback, replay, restart, redaction, determinism, and set equality | Suites do not yet exist or have not been proven exact | Consumer owners add/map suites later; Evidence supplies fixtures/contributions. |
| 2026-08-11T18:55:55-04:00 | Codex / S00 implementation | Per-slice validation routing fixed | Rechecked live task-guide commands for affected owners | `make task-guide`; tracker update | Narrow owner slices precede broader S10 gates | Test suites remain to implement | S01 adds and runs route authorization evidence. |
| 2026-08-11T19:42:54-04:00 | Codex / S01 implementation | AC-543 and frontend coverage registered and passing | Added service-backed upload matrix and legacy-token proof; updated authored Evidence family selectors | §6.1 pass/failure roots | Exact current-use denials preserve issued lease/no-object state; permitted roles complete once; replay fails closed | Final broad cross-owner matrix remains S10 | Add S02 differential mapper fixtures before deleting duplicate paths. |
| 2026-08-11T20:09:18-04:00 | Codex / S02 implementation | Canonical mapper and differential suites registered and passing | Added owner mapper fixtures and expanded upload/attach projection integration coverage | §6.2 Evidence and Timeline command roots | Rich/null/state mapping and 0/1/2/tombstone count parity are executable; diagnostic failures retained | Final broad cross-owner matrix remains S10 | Add S03 exact deterministic effect and rollback evidence before changing support refresh. |
| 2026-08-11T20:38:41-04:00 | Codex / S03 implementation | Atomic effect, registry, protocol, and boundary evidence passing | Expanded existing Evidence integration and Projections registry rows; activated Collaboration intent test | §6.3 retained roots | Exact targeted/unaffected behavior, deterministic neutral results, multi-view arrays, remove compatibility, and rollback are executable | Final broad cross-owner matrix remains S10 | S04 must preserve these rows while removing Store construction. |
| 2026-08-11T20:55:59-04:00 | Codex / S04 implementation | Cohesive construction and all retained behavior pass | Migrated former Store-focused tests to narrow app-support/runtime capabilities and added missing-dependency cases to the source-owner row | §6.4 retained roots | Evidence 58/58, service-backed 46/46, exact server/construction rows, boundary, topology, and shape checks pass | Ordinary/import parity is deliberately deferred to S05 | Add parity evidence before replacing broad owner dependency lists. |
| 2026-08-11T21:21:27-04:00 | Codex / S05 implementation | Composition failures and ordinary/import parity are executable | Added explicit requested-time parity plus exact durable/effect/replay assertions; expanded runtime and Imports assembly construction evidence | §6.5 retained roots | Evidence 58/58, service-backed 46/46, Imports service-backed 29/29, parity, boundaries, topology, and shape checks pass | Full five-suite provider closure remains S07 | Consolidate policy in S06 without splitting the shared mutation kernel. |
| 2026-08-11T22:07:42-04:00 | Codex / S06 implementation | Cross-entry policy and explicit lifecycle behavior are executable | Added exhaustive private policy and rollback cases; updated ordinary/import/service/browser fixtures to the explicit transition | §6.6 retained roots | Evidence 60/60, service-backed 46/46, frontend 364/364, Imports parity, browser surfaces, boundaries, topology, and shape checks pass | Full five-suite provider closure remains S07 | Implement the exact Reporting allowlist and provider suites. |
| 2026-08-11T22:36:03-04:00 | Codex / S07 implementation | Exact Reporting and provider closure is executable | Added unknown-column/deleted/forbidden/export-model fixtures, malformed-output atomicity, current/vNext Recovery inventory closure, subtype contribution closure, and stable suite identities | §6.7 retained roots | All five named suites, Reporting 14/14, Records portability, boundary, topology, and shape gates pass | Final broad cross-owner matrix remains S10 | Build the production-inert durable cleanup engine and service-backed race matrix in S08. |
| 2026-08-11T23:10:21-04:00 | Codex / S08 implementation | Full cleanup matrix authored and routed but not executed | Added exact inert/concurrency/restart/batch/retry/not-found/timeout/race/deadline/retention suite | Three exact `service-backed-test-slice` attempts plus non-service gates in §6.8 | Test package compiles; each service attempt failed during disposable PostgreSQL startup before Go execution | Repeated Docker inspect/readiness timeout is the sole S08 exit blocker | Rerun the exact row first when infrastructure changes; then full Evidence service and Recovery inventory rows. |
| 2026-08-12T19:08:22-04:00 | Codex / S09 implementation | Activation, typed storage, telemetry, process, and multi-instance evidence pass | Added dispatcher/storage unit suites, server readiness/telemetry tests, platform typed-purpose coverage, and two-dispatcher service-backed coverage | §6.9 exact/full command matrix | Evidence 62/62 and 47/47, app server 45/45 and 34/34, telemetry 10/10, object store 6/6, and process 7/7 pass | Final repository-wide matrix remains S10 | Run the S10 owner, frontend, browser, harness, security, drift, finalize, fast, and broad checks. |
| 2026-08-12T20:00:26-04:00 | Codex / S10 validation and handoff | Complete reproducible validation basis retained | Ran every required owner, service, frontend, browser, harness, security, drift, finalize, fast, and full gate; retained both diagnostic and passing roots | §6.10 command matrix | Required owner rows pass; `test-fast` is 355/355 and `check` is 755/755; the only transient external fixture failure passed unchanged on exact/full rerun | None | Use §6.10 roots for review and future regression comparison. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-11T06:38:40-04:00 | Codex / initial planning session | Route admission, handle recheck, concealment, and opaque-target seams identified | Inspected route admission/errors/dependencies, access repository, upload tokens, blobref, Core 04, and security tests; touched only tracker | Targeted `sed` and `rg -n` for membership/role/handle/storage identifiers | Current checks are edge-local and intentionally guarded | RB-004 before route/access restructuring | Add a six-operation role/membership matrix if existing cross-owner evidence is not exact. |
| 2026-08-11T18:27:01-04:00 | Codex / NLSpec-grade revision | Exact six-operation role, recheck, concealment, precedence, and state-change rules authored in Core 04 | Inspected Core 01/04 and live Evidence admission, tokens, handles, errors, tests; touched authorized docs only | Targeted reads and documentation patching | Design resolved without a new public token, role, route, error, or resource | Core 04 adoption and table-driven characterization remain required | Pass `module.evidence.route_authorization_concealment_matrix` before route restructuring. |
| 2026-08-11T18:55:55-04:00 | Codex / S00 implementation | Adopted authorization defect is next | Rechecked Core 04 REQ-04-156 and frontend/server upload paths | Targeted `rg`/`sed` | Exact gate order, clean token cutover, and one-shot lease are fixed decisions | S01 AC-543 evidence | Implement S01 before mapper or Store changes. |
| 2026-08-11T19:42:54-04:00 | Codex / S01 implementation | Adopted Evidence authorization defect closed | Changed six-operation route admission/order where required and bound uploads to current actor/session plus durable lease | AC-543 service-backed coverage, handle state-change suite, frontend CSRF tests, legacy rejection | Old/cross-session/replayed targets fail closed; viewer upload denied; deployment admin without membership is insufficient | Replica drain must precede deployment cutover | Preserve this gate order through S02-S09. |
| 2026-08-11T20:09:18-04:00 | Codex / S02 implementation | S01 authorization posture preserved | Structural row/link refresh changes remained inside the caller transaction and did not alter HTTP admission | Full Evidence owner/service-backed suites | AC-543 regression rows remain green; denied paths and capability behavior are unchanged | Replica drain remains an operational S01 handoff item | Preserve security gates while moving support effects in S03. |
| 2026-08-11T20:38:41-04:00 | Codex / S03 implementation | S01 authorization posture and concealment preserved | Support effects consume only authoritative subject IDs/types after the owning Evidence mutation; no new route or storage access exists | Full Evidence owner/service-backed suites | No authorization, capability, object-store, or public error behavior changed | Replica drain remains an operational S01 handoff item | Preserve route gates while decomposing Store in S04. |
| 2026-08-11T20:55:59-04:00 | Codex / S04 implementation | Authorization capabilities are isolated from cleanup and source mutation | Route composition embeds only blob lifecycle and access handles; upload/session gates and access rechecks are unchanged | Full Evidence owner/service-backed suites | AC-543 regression rows remain green; cleanup/source mutation cannot be acquired through `RouteService` | Replica drain remains an operational S01 handoff item | Preserve the narrow route facade through S05 composition. |
| 2026-08-11T21:21:27-04:00 | Codex / S05 implementation | Authorization and route capability boundaries remain unchanged | Composed routes only through the validated owner runtime; Imports receives no route/blob/access capability | Full Evidence owner/service-backed suites and Imports parity | AC-543 coverage remains green and shared create behavior introduces no transport authority | Replica drain remains an operational S01 handoff item | Preserve authorization gates while centralizing policy in S06. |
| 2026-08-11T22:07:42-04:00 | Codex / S06 implementation | Authorization gates and reference concealment share owner policy | Centralized closed-incident, malformed reserved reference, and persisted key identity decisions without moving route admission order | Full Evidence owner/service-backed suites and object-store contract rows | AC-543 remains green; malformed/mismatched references fail before backend access; no storage identifier leaks | Replica drain remains an operational S01 handoff item | Preserve route order and policy calls through S07. |
| 2026-08-11T22:36:03-04:00 | Codex / S07 implementation | Reporting disclosure is closed without changing route authorization | Proved forbidden storage/capability classes and a synthetic future column cannot enter facts or the persisted export model | Exact Reporting suite and full Reporting owner slice | No physical object identity or deleted Evidence is exposed; S01 route gates are untouched | Replica drain remains an operational S01 handoff item | S08 claim state must remain private and omitted from Reporting and public diagnostics. |
| 2026-08-11T23:10:21-04:00 | Codex / S08 implementation | Claim state is private, excluded from authoritative Recovery, and mutation-guarded | Added closed claim/failure states, opaque tokens, database guards, and no public/log/Reporting projection | Recovery/catalog/shape/boundary checks; exact live suite authored | No public identifier, raw error, object key, or claim token is emitted by the engine surface | Live race and typed-error assertions blocked before execution by Docker readiness | Preserve private state and do not activate telemetry/storage dispatch until S08 passes. |
| 2026-08-12T19:08:22-04:00 | Codex / S09 implementation | Active cleanup retains private state and closed observations | Bound deletion to typed `evidence_cleanup`; observer emits only closed operation/result/error classes and aggregate health values; no IDs, keys, capabilities, hashes, filenames, or raw errors | Exact telemetry/storage/server tests plus full Evidence/app-server/platform slices | Typed not-found is idempotent; transient failures retry; forbidden sentinels do not enter metrics or logs | Replica-drain requirement from S01 remains an operational deployment item | Retain authorization and telemetry regression gates in S10. |
| 2026-08-12T20:00:26-04:00 | Codex / S10 validation and handoff | Security, disclosure, storage, and telemetry closure complete | Revalidated AC-543, clean token cutover, explicit Reporting allowlist, private claims/leases, typed deletion, safe observations, and final database ACLs | Evidence/Reporting/Recovery/platform suites, targeted gosec, browser flows, and full `check` | Old/cross-session/replayed uploads fail closed; forbidden fields/identifiers remain absent; PUBLIC cannot use new table composite types | Old-replica drain remains a deployment prerequisite, not an implementation blocker | Drain old replicas, apply migrations, then start the new issuer/runtime. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-11T06:38:40-04:00 | Codex / initial planning session | Tracker ready for implementation handoff; all implementation rows remain unstarted | Inspected tracker evidence set; touched only tracker | `date -Iseconds`; `git status --short --branch`; `git diff --check`; `make lint-markdown` | Worktree was clean before tracker creation; inventory counts matched 58/58; diff and Markdown checks passed; transient ignored lint artifacts were removed; no production change made | RB-001 is a behavior-authorization blocker; RB-002 through RB-005 are evidence gates | Recheck repository delta, authorize S-00 only, and update this log before any test/code edit. |
| 2026-08-11T18:27:01-04:00 | Codex / NLSpec-grade revision | Gap closure means design/authority closure only; S-00 through S-07 remain unimplemented | Inspected all named material; touched only four authorized docs | `git diff HEAD --check`; `make lint-markdown`; inventory, history, vocabulary, authority, and scope scans | Documentation lint and all tracker invariants passed; closure ledger distinguishes READY gates from the separately blocked correction | S-07 needs later authorization; S-00 absent; owner amendments need adoption | Authorize only S-00 or separately authorize S-07. |
| 2026-08-11T18:55:55-04:00 | Codex / S00 implementation | S00 checkpoint prepared | Changed tracker only | `git diff --check`; `make lint-markdown` | G01 closed; S01 is the only permitted next slice | None after lint | Begin S01. |
| 2026-08-11T19:42:54-04:00 | Codex / S01 implementation | S01 checkpoint prepared | Runtime, migration, frontend, tests, authored contract/tooling inputs, generated outputs, and tracker recorded in §6.1 | `git diff --check` and §6.1 matrix; Markdown lint follows this checkpoint | G02 closed; diagnostics retained; no required S01 gap remains | Cleanup remains deliberately inactive until S08-S09 | Begin S02 only after this tracker passes Markdown lint. |
| 2026-08-11T20:09:18-04:00 | Codex / S02 implementation | S02 checkpoint prepared | Canonical mapper, Store/provider call sites, Evidence/Timeline attachment projection boundary, tests, authored routing, generated topology, and tracker recorded in §6.2 | `git diff --check` and §6.2 matrix; Markdown lint follows this checkpoint | G03 closed; no duplicate mapper or stale linked-count path remains | Cleanup remains deliberately inactive until S08-S09 | Begin S03 only after this tracker passes Markdown lint. |
| 2026-08-11T20:38:41-04:00 | Codex / S03 implementation | S03 checkpoint prepared | Neutral effects boundary, registry/state-load/storage implementation, exact integration evidence, protocol adaptation, harness input, and tracker recorded in §6.3 | `git diff --check` and §6.3 matrix; Markdown lint follows this checkpoint | G04 closed; no whole-incident support rebuild or Evidence-owned support layout remains | Cleanup remains deliberately inactive until S08-S09 | Begin S04 only after this tracker passes Markdown lint. |
| 2026-08-11T20:55:59-04:00 | Codex / S04 implementation | S04 checkpoint prepared | Cohesive services, callers, tests, authored routing, generated topology, and tracker recorded in §6.4 | `git diff --check` and §6.4 matrix; Markdown lint follows this checkpoint | G05 closed; no broad Store or compatibility surface remains | Cleanup remains deliberately inactive until S08-S09; broad composition is G06/S05 | Begin S05 only after this tracker passes Markdown lint. |
| 2026-08-11T21:21:27-04:00 | Codex / S05 implementation | S05 checkpoint prepared | Validated runtime, shared mutation kernel, narrow consumer wiring, parity tests, authored routing, generated topology, and tracker recorded in §6.5 | `git diff --check` and §6.5 matrix; Markdown lint follows this checkpoint | G06 closed; no panic-based broad composition or consumer-owned Evidence reconstruction remains | Cleanup remains deliberately inactive until S08-S09; duplicated policy is G07/S06 | Begin S06 only after this tracker passes Markdown lint. |
| 2026-08-11T22:07:42-04:00 | Codex / S06 implementation | S06 checkpoint prepared | Private policy, all callers, explicit frontend transition, cross-entry tests, authored routing, and tracker recorded in §6.6 | `git diff --check` and §6.6 matrix; Markdown lint follows this checkpoint | G07 closed; no duplicate owner policy or implicit attachment promotion remains | Cleanup remains deliberately inactive until S08-S09; provider closure is G08/S07 | Begin S07 only after this tracker passes Markdown lint. |
| 2026-08-11T22:36:03-04:00 | Codex / S07 implementation | S07 checkpoint prepared | Closed Reporting construction, logical-target owner contribution, five named suites, Records cursor fix, authored routing, generated topology, and tracker recorded in §6.7 | `git diff --check` and §6.7 matrix; Markdown lint follows this checkpoint | G08 and RB-005 closed; no subtractive Evidence JSON or cross-owner Links classification remains | Cleanup remains deliberately inactive until S08-S09 | Begin S08 only after this tracker passes Markdown lint. |
| 2026-08-11T23:10:21-04:00 | Codex / S08 implementation | S08 checkpoint is explicitly blocked; S09/S10 unstarted | Implementation, migration, Recovery/schema projections, exact suite, generated outputs, and §6.8 tracker record | §6.8 command matrix; `git diff --check`; Markdown lint follows | Non-service gates pass and production remains inert | Three identical external Docker readiness failures before test execution | Resume only after shared Docker readiness changes; rerun exact/full service gates and update tracker before any S09 work. |
| 2026-08-12T18:39:56-04:00 | Codex / S08 resumed validation | S08 complete; S09 authorized only after tracker lint | Corrected cleanup fixture SQL typing and Recovery contribution order; changed Evidence cleanup integration test, Evidence recovery contribution, and this tracker | §6.8 resumed command matrix | Exact cleanup, full Evidence service, owner contribution, Recovery inventory, migration, generation, and shape gates pass | Cleanup remains inert until S09 activation | Lint this checkpoint, then begin S09 only. |
| 2026-08-12T19:08:22-04:00 | Codex / S09 implementation | S09 checkpoint prepared | Dispatcher, typed storage, server activation/telemetry, owner specifications, process fixture, authored routing, generated topology, and this tracker recorded in §6.9 | `git diff --check` and §6.9 command matrix; Markdown lint follows this checkpoint | G10 closed; cleanup is active, bounded, multi-instance-safe, observable, and private | Final cross-owner and broad repository closure remains S10 | Begin S10 only after this tracker passes Markdown lint. |
| 2026-08-12T20:00:26-04:00 | Codex / S10 validation and handoff | Overall tracker and implementation complete | Full S00-S10 worktree and §6.10 evidence; final tracker reconciliation | §6.10 matrix; `git diff --check`; final `make lint-markdown` | No unresolved required gap, blocker, drift, or failing gate remains | None | Handoff complete; follow deployment ordering in §6.10. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Cleanup activation is destructive and requires its complete safety matrix. | It activates deletion, migration, scheduling, and retries. | S08 exact and full Evidence service-backed matrices must pass before S09 starts. | `CLOSED`: exact/full Evidence, Recovery inventory, and drift gates pass in §6.8. |
| RB-002 | Canonical row parity lacked executable proof. | S02 could not remove duplicate interpretation safely. | S02 differential row suite. | `CLOSED`: §6.2 mapper and linked-count parity evidence passes. |
| RB-003 | Support-projection ownership requires executable atomicity proof. | S03 changes cross-owner transaction effects. | §3.1 interface and atomicity suite. | `CLOSED`: §6.3 exact selection, unrelated-row, deterministic-order, and rollback evidence passes. |
| RB-004 | Evidence route behavior did not satisfy adopted Core 04. | Disclosure and authorization defects required correction before structural work. | S01 AC-543 matrix. | `CLOSED`: §6.1 evidence passes; preserve as a regression gate. |
| RB-005 | Consumer behavior is not yet proven by the named suites. | Rearrangement can leak, lose, or diverge state. | Five §7.2 suites. | `CLOSED`: all five authored owner suites and the Reporting owner aggregate pass in §6.7. |

No owner contradiction is currently established. Any later contradiction MUST
be recorded as `BLOCKED: owner contradiction` and the affected work MUST stop.

## 12. Binary Completion Criteria

- **PASS:** All 70 current target files remain inventoried: 48 production Go,
  21 test Go, and one `.gitkeep`.
- **PASS:** Every public route, handle, row/query, saved-view, lifecycle,
  projection, authorization, revision, Collaboration, Reporting, Imports,
  portability, Recovery, generated, frontend, and harness risk has owner/test
  posture.
- **PASS:** Core 04 §2.0A solely owns the authorization matrix.
- **PASS:** Core 01 owns routes, bindings, lifetimes, states, envelopes, errors.
- **PASS:** Reporting owns the exact twelve-member allowlist/default omission.
- **PASS:** Every workflow has dependencies, validation, and checkpoint.
- **PASS:** S00-S10 are complete in the required serial sequence.
- **PASS:** S08 exact/full service-backed safety, Recovery inventory, and drift
  gates pass; S09 activation/storage/telemetry/server gates pass; S10 owner,
  browser, security, harness, drift, fast, and full repository gates pass.
- **PASS:** Implementation authority covers every slice; destructive cleanup
  activated only after S08 evidence and is governed by the S09 private runtime.
- **PASS:** The projection interface uses `pgx.Tx`/`uuid.UUID`, closed change
  kinds/records, deterministic order, correct ownership, no shim/partial result.
- **PASS:** Row parity defines oracle, fixtures, exact comparisons, zero gate.
- **PASS:** Cleanup fixes schedule, due time, claim, lease, deletion deadline,
  retries, lifecycle exclusion, restart, shutdown, telemetry, retention, tests.
- **PASS:** Provider suites cover success, failure, rollback, replay, restart
  where applicable, redaction, determinism, and set equality.
- **PASS:** Unknown commands have a reason/discovery method; none invented.
- **PASS:** Every initial handoff row remains and every table has a new row.
- **PASS:** Framework mismatch and document-placement matrix are explicit.
- **PASS:** Product completion is established by §6.10, and this tracker records
  every slice `DONE` with no unresolved required gap or blocker.

## 13. Iteration 2 Delta Inventory

Iteration 2 starts from the completed S10 state and is a production-readiness
cleanup, not a feature phase. It considers only accidental Go surface,
fragmented owner construction, untyped Evidence object-store fallbacks,
transitional terminology and files, and permanent evidence that prevents those
conditions from returning. The Iteration 1 behavior and security freeze remains
in force.

| Current path or surface | Current evidence | Iteration 2 disposition | Reason |
| --- | --- | --- | --- |
| `internal/modules/evidence/.gitkeep` | The directory contains production and test files, but the original empty-directory marker remains. | Delete in S16. | It has no runtime or documentation value and falsely describes the package state. |
| Root exported surface | Request decoders, request hashes, service implementations, operation DTOs, validators, cleanup internals, and constructors are exported despite having no non-test consumer outside Evidence. | Account for every export in S12; unexport or remove accidental members in S13 and S15; enforce the final set with an AST guard. | Capitalization has become an accidental compatibility mechanism and obscures the intended owner boundary. |
| Route composition | `RegisterRoutes(settings, WithRouteService(evidenceOwner.RouteService()))` is the only production construction path; `RouteOption` and `WithRouteService` exist only to inject one required capability. | Replace with `OwnerRuntime.RouteRegistrar(Settings)` in S13 and remove the optional injection surface. | A required application capability must not be represented as an optional variadic route setting. |
| Cleanup composition | Server assembly separately constructs the cleanup service, typed deleter, observer, and dispatcher before constructing `OwnerRuntime`. | Make cleanup part of the validated owner runtime in S13 while preserving server-owned activation after readiness and reverse-order shutdown. | Split construction permits an incomplete owner runtime and duplicates dependency validation. |
| Test-only contribution constructor | `NewTimelineAttachmentContribution` is documented as a focused-test helper and has no production caller. | Remove in S13; cross-owner tests use one complete reusable owner-runtime fixture. | Production APIs must not exist solely to let tests bypass the owner composition path. |
| Route object-store adapter | Evidence prefers `objectstore.TypedStore` but falls back to `PutObject`, `ReadObject`, and `StatObject`. | Require typed storage at composition and delete every fallback in S14. | The fallback bypasses closed purposes and permits less-governed storage implementations in a security-sensitive path. |
| Incident-bundle blob portability | `IncidentBundleBlobPortability` exposes a public `objectstore.Store` field and calls untyped read, write, and delete methods. | Replace literal construction with a fallible typed constructor and closed operation purposes in S14. | Bundle staging and cleanup need the same typed error and purpose governance as ordinary Evidence operations. |
| `store.go` | The former broad `Store` is gone, but the 1,285-line filename still combines blob lifecycle, access handles, cleanup, route operations, source mutation, DTOs, errors, and constructors. | Split along capability boundaries in S16 after surface reduction. | The obsolete name and mixed responsibilities make future changes harder to locate and review. |
| `routes.go`, `api.go`, and `workbook_facade.go` | Blob routes, handle routes, registration, decoding, mutation types, create, and patch orchestration remain in large mixed files. | Split by cohesive behavior in S16 without adding a new module or transport owner. | Smaller owner-local units reduce change risk without inventing package boundaries. |
| Large integration suites | `integration_test.go` and `evidence_integration_test.go` mix distinct route, lifecycle, projection, Collaboration, and cleanup behavior. | Split by existing semantic behavior in S16 while retaining test names and verification ownership. | Test layout should explain the contracts it protects and must not create false production seams. |
| Transitional terminology | Production comments still refer to S09; tests and filenames use `legacy`, `characterization`, and phase-era wording. | Replace with current contract language in S16. Retain the underlying version-1 rejection and migration-history assertions. | Historical wording hides continuing invariants, while deleting the negative evidence would weaken security and upgrade safety. |
| Workbook-oriented root names | Evidence publishes `WorkbookContribution`, `Workbook*Request`, `Workbook*Command`, `WorkbookMutationResult`, and `WorkbookFieldValue`. | Rename atomically to Evidence-native mutation names in S15, without aliases. | Workbook owns generic coordination; the Evidence owner API should describe Evidence mutations rather than its current consumer. |
| Provider packages | Projection, Reporting, Recovery, Revisions, rollback, and delete/restore packages have live owner or assembly consumers and passed the Iteration 1 provider suites. | Retain unless executable evidence establishes an ownership inversion. | Package movement by appearance would add churn without removing coupling. |
| `workbookprojection` published language | The package is an active cross-owner Projections contract with live row, contribution, and support-effect consumers. | Defer redesign to a separately authorized cross-owner iteration. | It is active behavior, not dead code, and changing it would dominate this cleanup. |

The intended final cross-package surface consists of owner runtime construction
and accessors, Evidence-native mutation contracts consumed by Workbook,
Timeline facts and attachment validation, cleanup lifecycle observations,
source-owner contribution constructors, `blobref`, and the active
`workbookprojection` language. S12 freezes exact membership from repository
reachability rather than relying on this summary alone.

## 14. Iteration 2 Compatibility and Deletion Posture

Iteration 2 preserves public Evidence routes, HTTP and OpenAPI shapes, frontend
requests and responses, authorization and concealment precedence, upload
capability version 2 behavior, idempotency rows, lifecycle transitions,
revision and Collaboration ordering, Reporting allowlists, incident-bundle
formats, Recovery identities, cleanup deadlines, telemetry vocabulary, and
stored data.

Internal Go APIs intentionally break without aliases, compatibility facades,
or deprecation shims. Unused exports, test-only production constructors,
variadic injection of a single required route dependency, untyped Evidence
storage fallbacks, public dependency fields, historical phase comments, file
layout, and Workbook-oriented names inside the Evidence owner API are not
compatibility contracts.

No database, data migration, Core owner, Domain, OpenAPI, frontend, Reporting,
telemetry-vocabulary, generated contract, or public behavior change is planned.
Domain vocabulary is unchanged. `docs/research/nlspec-spec.md` informs planning
completeness and binary acceptance criteria only; it is not Evidence product
authority. If implementation discovers a required normative behavior change,
migration, or owner contradiction, the affected slice is blocked and returned
to specification planning instead of expanding tactically.

The negative version-1 upload-capability fixture, typed-storage rejection, and
historical migration-index assertions remain because they provide continuing
security and upgrade value. Their names will describe the protected invariant,
not a supported legacy mode. No version-1 decoder, untyped Evidence fallback,
or migration compatibility path is retained.

## 15. Iteration 2 Gap Matrix

| Gap | Areas | Remediation | Rationale | Expected long-term benefit | Compatibility or migration impact | Risk if unresolved | Validation criteria |
| --- | --- | --- | --- | --- | --- | --- | --- |
| G12 — The completed Iteration 1 plan is still presented as the active execution posture. | Tracker and documentation. | Preserve S00-S10 and Sections 2-12 as historical evidence; add a delta-only controlling Iteration 2 with S11-S17, G12-G18, validation, handoff, deferrals, and binary criteria. | Rewriting completed evidence would lose auditability, while leaving it controlling would direct future work at closed gaps. | One unambiguous current ledger with intact historical proof. | Documentation only. | A later session can repeat closed work, miss the current deletion policy, or misread a historical blocker as live. | Section 1 points to the new iteration; S11-S17 are serial and internally consistent; Markdown and diff gates pass. |
| G13 — Accidental and test-only exports enlarge the Evidence API. | Implementation, tests, authored verification routing, and boundary policy. | Inventory every root and owner-facing subpackage export; classify it as retain, rename, unexport, or remove; add an AST-backed exact export guard; migrate private-service tests in-package and cross-owner tests to a complete owner-runtime fixture. | Current capitalization exposes implementation types and constructors merely because external-package tests use them. | A smaller deliberate API that resists accidental growth and makes dependency direction visible. | Internal Go source break only; no aliases. | Test helpers and owner-private DTOs become de facto production contracts, increasing coupling and blocking later decomposition. | Every remaining export has a production consumer or named contribution role; old symbols have zero references; exact export and consumer compilation tests pass. |
| G14 — Routes and cleanup are composed outside the complete owner runtime. | Owner runtime, server assembly, test support, telemetry binding, and tests. | Make `OwnerRuntime` construct route, mutation, attachment, import, and cleanup capabilities from one validated dependency set; expose a route registrar and cleanup dispatcher; retain server-owned start-after-readiness and shutdown; remove optional route injection and standalone internal constructors. | A successful owner construction should prove that every production capability is complete and cannot be forgotten by a consumer. | One fallible composition path, simpler assembly, and consistent test construction. | Internal constructor and accessor break only. | Partial composition can silently omit cleanup, duplicate validation, or let tests exercise a topology production never uses. | Missing-dependency matrix, server composition contract, readiness/start/shutdown tests, and searches for external internal-service construction pass. |
| G15 — Evidence still admits untyped object-store behavior. | Evidence routes, Workbook initial-blob flow, cleanup, Incident Bundles, server assembly, object-store tests, and telemetry tests. | Require `objectstore.TypedStore` before readiness; remove route and portability fallbacks; use `product_upload`, `product_read`, `migration_copy`, `staged_cleanup`, and `evidence_cleanup` for their exact operations; make bundle portability construction fallible and private-field based. | Typed requests carry closed purpose and error semantics that untyped calls cannot enforce. | One governed storage path with fail-fast capability validation and reliable telemetry/error classification. | Internal dependency type break; production stores already implement the typed contract; no data migration. | A permissive store can bypass purpose policy, emit provider-specific errors, or perform unsafe bundle cleanup. | Untyped-only stores fail construction; exact-purpose assertions, typed not-found behavior, route flows, bundle round trip, staged cleanup, Recovery, and platform telemetry pass; no untyped method call remains under Evidence. |
| G16 — Workbook terminology and owner-private types leak through the root API. | Evidence mutation API, Workbook adapters, Imports, tests, and export guard. | Rename the live owner contract to `MutationContribution`, `CreateRequest`, `PatchRequest`, `PatchChange`, `CreateCommand`, `PatchCommand`, `ConflictCommand`, `MutationResult`, and `FieldValue`; make the facade, internal parameters, validators, request decoders, hashes, services, DTOs, and import aliases private where no cross-owner consumer exists. | The owner API should model Evidence concepts and expose only information required for independent consumers. | Clearer owner language and lower future rename/decomposition cost. | Internal Go source break without aliases; operation strings, persisted JSON, errors, and wire behavior remain exact. | Consumer vocabulary becomes embedded in Evidence and owner-private changes require broad caller migrations. | Exact export guard, zero old-name references, stored idempotency replay, public error translation, ordinary/import parity, and affected owner suites pass. |
| G17 — Transitional files, tests, comments, and placeholders reduce cohesion. | Implementation layout, tests, harness routing, and documentation comments. | Delete `.gitkeep`; split `store.go`, route/API files, mutation facade, and integration evidence along existing capability boundaries; rename characterization and legacy-oriented tests to contract language; remove phase-era comments while preserving test identities and invariants. | The former implementation sequence should not remain encoded in the steady-state package structure. | Smaller review units, clearer ownership, and safer future feature expansion. | Structural only; no new production package or verification owner. | Mixed files encourage duplicate sequencing and stale comments mislead maintainers about active behavior. | No placeholder, `store.go`, transitional production comment, false package seam, or duplicate sequencing remains; test declarations and semantic routing stay complete. |
| G18 — Final deletion and compatibility evidence is not yet assembled. | Validation, generated state, security, documentation, and handoff. | Run owner-first and cross-owner slices, storage/security/browser gates, generated and migration drift, finalization, broad checks, build, and release verification; record all roots and compatibility conclusions. | Local compilation cannot prove cross-owner compatibility or that removed surfaces have no hidden consumer. | Reproducible production-readiness evidence for future maintenance. | None beyond the internal breaks already assigned to earlier slices. | Dead paths, stale harness routing, or behavior drift can survive an apparently successful cleanup. | S17 matrix passes, removed surfaces have no shim or consumer, G12-G18 are closed, and tracker state, evidence, risks, and completion checklist agree. |

## 16. Iteration 2 Execution Policy and Tracker Gate

S11 through S17 are strictly serial. Only one slice may be `ACTIVE` or
`VALIDATING`. `READY` means the slice is fully planned but implementation has
not begun. After validating a slice and before activating its successor, this
tracker MUST record status, gap changes, files changed, substantive decisions,
compatibility impact, commands and results, run roots, retained evidence,
risks, blockers, handoff state, and binary criteria. `make lint-markdown` is the
last gate at every checkpoint.

A failed required check marks the slice `BLOCKED`; later slices do not start.
Failures are returned to the slice that owns the changed behavior rather than
receiving tactical cleanup in S17. A discovered normative behavior change,
database migration, or owner contradiction also blocks the slice until the
adopted owner and this plan are corrected.

Generated roots and generated execution topology are changed only through
their authored inputs and public Make targets. Tests, tools, generators,
conformance, and runtime code MUST NOT read or derive facts from this tracker or
other Markdown. Each slice has an atomic rollback boundary and must preserve
the last completed checkpoint evidence.

## 17. Iteration 2 Serial Workstreams

| Slice | Workstream | Depends on | Status | Rollback boundary | Exit criterion |
| --- | --- | --- | --- | --- | --- |
| S11 | Iteration 2 tracker rebaseline | S10 | DONE | Revert only the Iteration 2 tracker additions and Section 1 status update. | Controlling inventory, compatibility, gaps, workstreams, validation, handoff, deferrals, and criteria are recorded; diff and Markdown gates pass. |
| S12 | Characterization and exported-surface accounting | S11 | DONE | Revert only additive characterization, export guard, and authored verification routing. | Every later removal has zero production consumers or a named migration; current contracts and export membership are executable. |
| S13 | Owner runtime and application composition | S12 | DONE | Revert owner runtime, server/test assembly, and route/cleanup constructor changes together. | One validated runtime constructs every Evidence capability; no external caller constructs owner internals. |
| S14 | Typed object-storage cutover | S13 | DONE | Revert typed port, Evidence adapters, Incident Bundle adapters, and their caller changes together. | Every Evidence storage operation is typed with its exact purpose and untyped-only construction fails closed. |
| S15 | Module-native API minimization | S14 | DONE | Revert owner API renames, consumer migrations, private-symbol changes, and export guard update together. | The final export allowlist passes; old names, aliases, and accidental exports are absent. |
| S16 | Cohesion and historical-residue cleanup | S15 | DONE | Revert file/test moves and terminology changes without changing S15 behavior. | Cohesive files and tests retain one implementation of each sequence and all verification identities. |
| S17 | Final validation and handoff | S16 | DONE | Return any correction to its owning slice; S17 changes only final evidence and tracker state. | Required owner, cross-owner, security, generated-state, browser, broad, build, and release gates pass and G12-G18 close. |

### S11 — Iteration 2 tracker rebaseline

- **Areas:** tracker and documentation.
- **Remediation:** Update Section 1; preserve Sections 2-12 as Iteration 1;
  add Sections 13-22 as the controlling delta inventory, compatibility policy,
  gaps, execution rules, workstreams, validation, ledger, handoff, deferrals,
  and checklist.
- **Compatibility:** Documentation only. Domain vocabulary is unchanged. No
  runtime, test, contract, generated, migration, schema, or frontend behavior
  changes.
- **Validation:** `git diff --check --
  docs/handoffs/evidence-module-refactor-tracker.md`, `make lint-markdown`, and
  final status review.
- **Exit:** G12 is closed, S12 is `READY` but inactive, and S13-S17 remain
  `PENDING`.

### S12 — Characterization and exported-surface accounting

- **Areas:** tests, authored verification routing, generated topology through
  Make, boundary policy, and tracker.
- **Remediation:** Inventory exports in the root, `blobref`,
  `workbookprojection`, and provider packages using actual non-test consumers.
  Classify every export as retain, rename, unexport, or remove. Add an
  AST-backed exact allowlist and an authored `module.evidence.api_minimization`
  verification identity. Freeze HTTP/OpenAPI behavior, authorization order,
  error mapping, stored idempotency replay, Collaboration effects, cleanup
  identity/cadence, Reporting output, incident-bundle bytes, Recovery
  contributions, and typed-storage purposes before structural edits.
- **Compatibility:** Additive tests and harness accounting only. Existing
  semantic identities remain; generated routing changes only through authored
  inputs and Make.
- **Validation:** Evidence unit/service-backed slices; affected Workbook,
  Timeline, Imports, Incident Bundles, Recovery, Projections, Revisions, server,
  object-store, and telemetry rows; harness and generated drift gates.
- **Exit:** Every S13-S16 change has executable retained behavior and a complete
  reachability disposition; no planned deletion depends solely on a text
  search.

### S13 — Owner runtime and application composition

The final post-S15 application-facing shape is:

```go
type OwnerRuntimeDependencies struct {
    // Existing required owner dependencies.
    ObjectStore     objectstore.TypedStore
    CleanupObserver CleanupObserver
    Now             func() time.Time
}

func (*OwnerRuntime) RouteRegistrar(Settings) httpapi.RouteRegistrar
func (*OwnerRuntime) CleanupDispatcher() *CleanupDispatcher
func (*OwnerRuntime) MutationContribution() MutationContribution
```

S13 retains `objectstore.Store` and `WorkbookContribution()` temporarily so
that the typed-storage cutover and mutation-language rename remain atomic S14
and S15 rollback boundaries. S14 narrows the dependency to
`objectstore.TypedStore`; S15 installs `MutationContribution()` and its renamed
contract.

- **Areas:** Evidence owner runtime, routes, cleanup, server assembly, test
  support, Workbook/Timeline/Imports composition, and tests.
- **Remediation:** Construct blob lifecycle, access handles, route operations,
  mutation facade, Timeline attachment contribution, Imports facade, cleanup
  engine, typed deleter, and dispatcher once in `NewOwnerRuntime`. Require and
  validate every dependency. Keep server responsibility for calling `Start`
  only after publication readiness and for reverse-order shutdown. Remove
  `RouteOption`, `WithRouteService`, `RouteService`, standalone
  `RegisterRoutes`, public service/deleter/dispatcher constructors, and
  `NewTimelineAttachmentContribution`. Cross-owner tests use one complete
  reusable owner-runtime fixture; focused private-service tests move into the
  Evidence package.
- **Compatibility:** Internal constructor/accessor break only; no public route,
  lifecycle, cadence, or telemetry change.
- **Validation:** Complete missing-dependency matrix, exact runtime surface,
  server startup/readiness/shutdown, cleanup activation, cross-owner
  composition, owner slices, boundary, and zero-construction searches.
- **Exit:** A successful owner construction proves all capabilities are
  available; no server or test caller reconstructs Evidence internals.

### S14 — Typed object-storage cutover

- **Areas:** Evidence routes and Workbook mutation, cleanup, Incident Bundles,
  server assembly, object-store contracts, telemetry, tests, and tracker.
- **Remediation:** Require `objectstore.TypedStore` before readiness. Delete
  `PutObject`, `ReadObject`, `StatObject`, and `DeleteObject` fallback paths.
  Use `product_upload` for upload writes and upload observation,
  `product_read` for availability, preview, download, and attachment reads,
  `migration_copy` for bundle export/import bytes, `staged_cleanup` for failed
  import staging, and `evidence_cleanup` for durable orphan deletion. Replace
  `IncidentBundleBlobPortability{ObjectStore: ...}` with a fallible constructor
  and private typed-store field. Preserve typed not-found and retry/error
  classification.
- **Compatibility:** Internal dependency break; production adapters already
  provide typed capabilities. No object key, byte, bundle, route, or data
  migration change.
- **Validation:** Untyped-only negative fixture, exact purpose calls, typed
  failure/not-found matrix, upload/attach/handle flows, bundle round trip and
  rollback cleanup, Recovery, object-store telemetry, browser-backed Evidence,
  boundary, and static absence of untyped calls below Evidence.
- **Exit:** Every Evidence-owned storage access is typed and purpose-bound; an
  incomplete adapter fails before readiness.

### S15 — Module-native API minimization

- **Areas:** Evidence root API, Workbook and Imports adapters, Timeline and
  application assembly, tests, export guard, and tracker.
- **Remediation:** Rename without aliases:
  `WorkbookContribution` to `MutationContribution`,
  `WorkbookCreateRequest` to `CreateRequest`,
  `WorkbookPatchRequest` to `PatchRequest`,
  `WorkbookPatchChange` to `PatchChange`,
  `WorkbookCreateCommand` to `CreateCommand`,
  `WorkbookPatchCommand` to `PatchCommand`,
  `WorkbookConflictCommand` to `ConflictCommand`,
  `WorkbookMutationResult` to `MutationResult`, and
  `WorkbookFieldValue` to `FieldValue`. Make the facade, source/blob/access and
  cleanup services, route handler, operation DTOs, internal request decoders
  and hashes, validators, parameters, and import alias private where no
  cross-owner consumer exists.
- **Retained surface:** `Settings`, owner runtime construction and accessors,
  Evidence-native mutation commands/results/errors consumed by Workbook,
  Timeline facts and attachment capability, cleanup observer/dispatcher
  lifecycle, legitimate contribution constructors, `blobref`, and
  `workbookprojection`.
- **Compatibility:** Intentional internal Go API break without aliases.
  Operation strings, persisted JSON, public errors, transaction order, and wire
  behavior remain exact.
- **Validation:** Export guard, zero old-name/alias search, existing idempotency
  row replay, public error translation, ordinary/import parity, cross-owner
  compilation, owner slices, boundary, harness, and generated drift.
- **Exit:** Every remaining export is deliberate and every owner-private symbol
  is private; provider packages remain only where behavior proves a legitimate
  seam.

### S16 — Cohesion and historical-residue cleanup

- **Areas:** implementation layout, tests, authored routing if file selectors
  change, comments, and tracker.
- **Remediation:** Delete `.gitkeep`. Replace `store.go` with cohesive files for
  blob lifecycle, access handles, cleanup state, route operations, and source
  mutation. Split route registration, blob routes, handle routes, and request
  decoding. Split mutation types/composition, create, patch, conflict,
  idempotency, and shared transaction behavior. Split the large integration
  suites along current semantic tests without creating a production or
  test-only subpackage. Rename characterization and legacy-oriented tests to
  contract language; remove S00-S09 and stale inert-until-S09 comments.
- **Retained evidence:** Immutable migration-history checks, version-1 upload
  rejection, typed-store rejection, test names required by semantic routing,
  and one implementation of every transaction sequence.
- **Compatibility:** Structural only.
- **Validation:** Evidence and affected owner slices, exact test-declaration and
  routing parity, export/import/boundary guards, absence searches, harness
  contract, generated drift when routing changes, and `make test-fast`.
- **Exit:** No `.gitkeep`, `store.go`, phase-era production comment, false
  package boundary, or duplicate sequence remains.

### S17 — Final validation and handoff

- **Areas:** validation, generated state, security, tracker, documentation, and
  handoff.
- **Remediation:** Run the Section 18 matrix from narrow owner checks through
  release verification. Return a failure to its owning slice. Reconcile the
  final export set, owner runtime, typed storage purposes, filenames, test
  routing, generated outputs, compatibility conclusions, retained roots,
  risks, and checklist.
- **Compatibility:** No new behavior or migration may be introduced in S17.
- **Exit:** Every required gate passes, removed APIs have no alias or consumer,
  G12-G18 are closed, and the tracker is internally consistent and marked
  `DONE`.

## 18. Iteration 2 Validation Matrix

| Layer | Command or scenario | Required point |
| --- | --- | --- |
| Tracker checkpoint | `git diff --check -- docs/handoffs/evidence-module-refactor-tracker.md`; `make lint-markdown`; status review | S11 and every later slice. |
| Owner guidance | `make task-guide ROLE=module-author OWNER=<owner-id>` | Before selecting or changing verification routing. |
| Formatting | `make format` | Every implementation slice that changes Go; never for tracker-only S11. |
| Evidence | `make test-slice OWNER=module.evidence`; `make service-backed-test-slice OWNER=module.evidence` | S12-S17. |
| Application/server | Matching unit and service-backed slices for `app.server` | S13, S14, and S17. |
| Mutation consumers | Matching unit and service-backed slices for `module.workbook`, `module.timeline`, and `module.imports` | S12-S16 as affected and S17. |
| Storage/portability | Matching unit and service-backed slices for `module.incidentbundles`, `module.recovery`, and `platform.objectstore` | S12, S14, and S17. |
| Projection/history | Matching unit and service-backed slices for `module.projections` and `module.revisions` | S12, S15, and S17. |
| Telemetry | Matching unit and service-backed slices for `platform.telemetry` | S13, S14, and S17. |
| Export surface | Exact AST allowlist; no removed name, alias, external internal-service constructor, public dependency field, or test-only production helper | S12-S16 and S17. |
| Behavior freeze | HTTP/OpenAPI, AC-543 precedence, idempotency replay, Collaboration ordering, cleanup cadence/deadlines, Reporting allowlist, bundle bytes, Recovery set | Before and after the owning structural slice. |
| Typed storage | Untyped-only rejection; exact operation purposes; typed error/not-found behavior; no untyped Evidence method call | S14 and S17. |
| Frontend | `make frontend-unit`; `make frontend-typecheck` | S14 and S17. |
| Boundaries and harness | `make backend-module-boundary-check`; `make harness-evidence-contract`; `make harness-contract` | Every changed-boundary or routing slice and S17. |
| Generated state | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make migration-drift` | Changed-input slices and S17. |
| Security/browser | `make go-gosec-targeted`; `make browser-e2e-stateful`; `make browser-e2e-webserver-backed` | S14 and S17. |
| Final repository | `make agent-finalize`; `make test-fast`; `make check`; `make build`; `make release-check` | S17. |

S17 reruns the unit and service-backed owner slices for Evidence, server,
Workbook, Timeline, Imports, Incident Bundles, Recovery, Projections,
Revisions, object store, and telemetry before broad checks. `make
agent-finalize` uses a retained successful `RESULTS_DIR` when available;
otherwise the tracker records that retained-run maintenance was skipped because
`RESULTS_DIR` was unset. A broad unrelated failure is recorded with its run
root and ownership evidence rather than silently excluded.

### S17 final evidence reconciliation

| Evidence group | Passing result and retained roots |
| --- | --- |
| Unit owner graph | Evidence 63/63 at `20260813T034157Z-p438215`; server 45/45 at `20260813T022136Z-p2941599`; Workbook 100/100 at `20260813T022136Z-p2941574`; Timeline 68/68 at `20260813T022721Z-p3137607`; Imports 40/40 at `20260813T022409Z-p3037603`; Projections 17/17 at `20260813T022409Z-p3037607`; Revisions 56/56 at `20260813T022409Z-p3037622`; Recovery 38/38 at `20260813T022520Z-p3094976`; Reporting 14/14 at `20260813T022520Z-p3094975`; Collaboration 44/44 at `20260813T022721Z-p3137605`; object store 6/6 at `20260813T022721Z-p3137606`; telemetry 10/10 at `20260813T022911Z-p3199906`; the exact Incident Bundles affected row 3/3 at `20260813T024547Z-p3735353`. |
| Service-backed owner graph | Evidence 47/47 at `20260813T034157Z-p438218`; server 34/34 at `20260813T022911Z-p3199930`; Workbook 66/66, Timeline 46/46, and Imports 29/29 at `20260813T023014Z-p3251079`, `20260813T023014Z-p3251078`, and `20260813T023014Z-p3251098`; Projections 13/13, Incident Bundles 15/15, and Recovery 26/26 at `20260813T023219Z-p3336081`, `20260813T023219Z-p3336086`, and `20260813T023219Z-p3336085`; Reporting 7/7, Revisions 40/40, and Collaboration 33/33 at `20260813T023259Z-p3372199`, `20260813T023259Z-p3372194`, and `20260813T023259Z-p3372205`; object store 5/5 at `20260813T023437Z-p3425269`. Telemetry publishes no service-backed rows. |
| Frontend and browser | Frontend unit 364/364 and typecheck 2/2 at `20260813T023449Z-p3427797` and `20260813T023449Z-p3427798`; stateful browser 36/36 and webserver-backed browser 62/62 at `20260813T023534Z-p3460373` and `20260813T023534Z-p3460371`; support browser 20/20 with one port lane at `20260813T030725Z-p4182573`; final release reran the complete browser graph at `20260813T035357Z-p806092`. |
| Security, boundaries, and harness | Targeted Go security 4/4 at `20260813T023449Z-p3427930`; backend boundary 3/3 at `20260813T023534Z-p3460324`; harness contract 2/2 at `20260813T023935Z-p3540183`. |
| Generated and migration state | Generate drift 4/4 at `20260813T023935Z-p3539965`; migration drift 5/5 at `20260813T023935Z-p3539995`; generated-artifact policy 3/3 and JSON shape 3/3 at `20260813T023951Z-p3545901` and `20260813T023951Z-p3545913`. |
| Final repository | `agent-finalize` 1/1 at `20260813T034458Z-p500768`; `test-fast` 355/355 at `20260813T034520Z-p503683`; build 7/7 at `20260813T034520Z-p503732`; check 755/755 at `20260813T035234Z-p726834`; release check 927/927 at `20260813T035357Z-p806092`. |

The broad Incident Bundles unit graph twice reached an unavailable external
object-store seed at `20260813T022520Z-p3094983` and
`20260813T024406Z-p3678376`; its exact affected row and complete
service-backed graph passed unchanged. The first release run at
`20260813T025217Z-p3945849` exposed an Entities support fixture that omitted
the adopted upload session and explicit Evidence lifecycle transition, plus a
browser support timeout. The fixture was corrected at its S13 consumer
boundary, the exact Entities row passed 3/3 at
`20260813T030406Z-p4152697`, browser support passed 20/20 with serialized port
lanes, and the complete final release passed. The first final check's unrelated
Jobs shutdown timing miss at `20260813T034726Z-p587456` passed 3/3 unchanged at
`20260813T035209Z-p725462`; the complete check then passed.

## 19. Iteration 2 Slice Checkpoint Ledger

| Slice | Status | Files changed | Decisions and compatibility | Commands and evidence | Risks, blockers, and next action |
| --- | --- | --- | --- | --- | --- |
| S11 | DONE | `docs/handoffs/evidence-module-refactor-tracker.md` only | Preserved S00-S10 and Sections 2-12; made Sections 13-22 controlling; recorded domain vocabulary unchanged, no product/specification/migration change, no implementation start, and no compatibility shims for future internal removals. | `git diff --check -- docs/handoffs/evidence-module-refactor-tracker.md` passed; `make lint-markdown` passed at `.cartulary/test-results/20260813T002119Z-p1363390`. | No implementation was performed. S12 is ready but inactive; S13-S17 remain pending. |
| S12 | DONE | Evidence exported-surface test; Evidence verification contract and authored family manifest; generated execution-topology render index; this tracker | Classified every current root, `blobref`, `workbookprojection`, internal-policy, and provider-package export as retain, rename, unexport, or remove; added an exact AST lock and negative fixture; preserved all product and data behavior; clarified the final API versus the S13-S15 atomic boundaries. | `make format` passed at `20260813T003836Z-p1380171`; `make generate` passed at `20260813T003640Z-p1375531`; the exact export row passed at `20260813T003840Z-p1383502`; affected unit and service-backed owner graphs passed through `20260813T005746Z-p1870116`; boundary, harness, drift, generated-policy, and JSON-shape gates passed through `20260813T005811Z-p1875935`; tracker diff passed and `make lint-markdown` passed at `20260813T005913Z-p1876750`. | One Incident Bundles unit row initially failed because its fixture could not seed the object store at `20260813T004450Z-p1514192`; the exact row passed unchanged at `20260813T004531Z-p1517928`. `platform.telemetry` has no service-backed rows; its unit graph passed. No stable blocker; S13 is ready. |
| S13 | DONE | Evidence owner runtime, route and cleanup composition; server assembly and lifecycle test; complete cross-owner test support; exact export guard; affected tests; this tracker | `OwnerRuntime` now validates the cleanup observer and clock, constructs routes, mutations, Timeline/Imports capabilities, cleanup service/deleter/dispatcher, and retains captured route dependencies. Server owns only readiness activation and reverse-order close. Removed production route options, standalone route registration, test-only Timeline constructor, and public internal-service constructors. S13 deliberately retains broad `objectstore.Store` and Workbook names for the S14/S15 rollback boundaries; no wire, data, schema, telemetry vocabulary, or migration change occurred. | Format passed at `20260813T012628Z-p2201542`; Evidence unit 63/63 at `20260813T011421Z-p1923034` and service-backed 47/47 at `20260813T011542Z-p1957629`; server unit 45/45 at `20260813T011752Z-p2013838` and service-backed 34/34 at `20260813T012255Z-p2142499`; Workbook 100/100 at `20260813T011838Z-p2040621`, Timeline 68/68 at `20260813T012052Z-p2083276`, Imports 40/40 at `20260813T012206Z-p2115072`, and telemetry 10/10 at `20260813T012342Z-p2165059`; boundary passed at `20260813T012632Z-p2204921`; harness Evidence contract passed. | A broad Entities run reached 45/46 but its unrelated shared support integration row used no upload session and received the established `session_required` result at `20260813T012353Z-p2167373`; the changed Entity composition unit compiled and passed. No stable S13 blocker; S14 is ready. |
| S14 | DONE | Evidence owner/runtime, routes, initial-blob path, cleanup deleter, and Incident Bundle blob port; Incident Bundles builder/importer/worker/route/import-provider composition; server typed-capability derivation; test support and exact export guard; this tracker | Evidence now accepts only `objectstore.TypedStore`; every product upload/read, migration copy, staged cleanup, and durable Evidence cleanup request carries its closed purpose. Server validates the raw capability after bootstrap and before owner/readiness assembly while retaining broad `Options.ObjectStore` for unrelated owners. One fallible blob-port instance is injected through all Incident Bundle operations. No key, byte, bundle, stored data, public error, schema, or migration changed. | Format passed at `20260813T013822Z-p2372483`; Evidence unit 63/63 at `20260813T013157Z-p2240526` and service-backed 47/47 at `20260813T013609Z-p2341783`; server exact typed/runtime rows passed at `20260813T013834Z-p2375910` and service-backed 34/34 at `20260813T014007Z-p2380581`; Incident Bundles exact retry passed at `20260813T013906Z-p2376662` and service-backed 15/15 at `20260813T013935Z-p2378214`; object store 6/6 and telemetry 10/10 passed through `20260813T014104Z-p2408519`; boundary passed at `20260813T014110Z-p2411061`; static searches found no platform untyped Evidence storage call or operation-local Incident Bundle port literal. | The first Incident Bundles unit graph attempt failed while seeding its external object-store fixture at `20260813T013520Z-p2335331`; the exact affected row passed unchanged. No stable blocker; S15 is ready. |
| S15 | DONE | Evidence mutation API and owner runtime; Workbook, Imports, Timeline, server, and test assembly consumers; root services, DTOs, decoders, validators, cleanup internals, and import adapter; `blobref`; exact export guard; affected tests; this tracker | Renamed the Evidence cross-owner mutation language without aliases and removed the Workbook-oriented accessor. Unexported owner-private services, route implementation, operation/request types, hashes, validators, parameters, import alias, and dispatcher test hooks. Removed unused panic/boolean `blobref` helpers and migrated tests to error-returning APIs. Durable operation strings, stored replay JSON, public errors, transaction/effect order, wire behavior, data, schema, and migrations are unchanged. | Format passed at `20260813T014547Z-p2415205`; exact export row passed at `20260813T014804Z-p2451915`; Evidence unit 63/63 at `20260813T014952Z-p2453288` and service-backed 47/47 at `20260813T015109Z-p2484682`; Workbook 100/100 at `20260813T015109Z-p2484671`, Imports 40/40 at `20260813T015109Z-p2484701`, server 45/45 at `20260813T015341Z-p2581216`, Revisions 56/56 at `20260813T015601Z-p2640834`, and Projections 17/17 at `20260813T015905Z-p2704814`; boundary 3/3 and generated drift 4/4 passed through `20260813T015905Z-p2704973`; the Evidence harness contract passed; zero-reference and diff searches passed. | Timeline's Go work completed but its browser measurement lifecycle failed fixture preflight at `20260813T015601Z-p2640822`; Incident Bundles reached the known unavailable external object-store seed at `20260813T015601Z-p2640830`. Neither path consumes a retired Evidence identifier, and all direct mutation consumers compile and pass. No stable S15 blocker; S16 is ready. |
| S16 | DONE | Evidence blob lifecycle, access handles, cleanup state, route operations/registration/blob/handle/errors, request decoding, mutation facade/create/patch/conflict/shared/source operations, integration suites, test names, provider references, boundary/probe/test-support inputs, generated topology, and this tracker | Deleted `.gitkeep`; replaced catch-all `store.go`; decomposed routes and mutation sequencing; split the two mixed integration suites into upload, attachment/authorization, projection/Collaboration, access-handle, route, lifecycle, and quarantine units. Replaced phase-era and legacy terminology with continuing contract language while retaining version-1 rejection, typed-store rejection, migration history, semantic row IDs, and the provider contract's required `CharacterizationRefs` field name. No production behavior, wire/data/schema, operation string, or migration changed. | Format passed at `20260813T021316Z-p2748854`; generation passed at `20260813T021319Z-p2752213`; exact export guard passed at `20260813T021206Z-p2747792`; Evidence unit 63/63 and service-backed 47/47 passed at `20260813T021338Z-p2755132` and `20260813T021338Z-p2755134`; boundary 3/3, harness 2/2, generated drift 4/4, JSON shape 3/3, and generated policy 3/3 passed through `20260813T021500Z-p2820576`; final `make test-fast` passed 355/355 at `20260813T021809Z-p2887824`; absence, exact-sequence, selector, and diff searches passed. | The first fast graph exposed a missing Evidence import in an Import Assembly fixture retained from S15; the exact failed row passed after restoring its required `ViewSchemaID` dependency at `20260813T021758Z-p2887295`, then the full graph passed. No stable blocker; S17 is ready. |
| S17 | DONE | Final owner/cross-owner validation; S13 consumer fixture correction; S14 lint cleanup; final export, storage, layout, routing, generated-state, compatibility, and handoff reconciliation; this tracker | The final production export set is exact and every member has an independent owner or contribution role. `OwnerRuntime` remains the single complete composition path. Typed purposes remain `product_upload`, `product_read`, `migration_copy`, `staged_cleanup`, and `evidence_cleanup`; no untyped Evidence platform-store call or operation-local Incident Bundle port remains. The final capability-oriented layout and semantic verification identities agree with generated topology. No public route, OpenAPI, frontend protocol, telemetry vocabulary, schema, stored-data, key/byte, bundle-format, or Domain change occurred; intentional internal Go breaks have no shim. | Final Evidence unit 63/63 and service-backed 47/47 passed at `20260813T034157Z-p438215` and `20260813T034157Z-p438218`; the exact export lock passed at `20260813T033628Z-p371103`. The complete owner matrix, frontend, browser, security, boundary, harness, generation/migration drift, generated-policy, and JSON-shape evidence is retained in the S17 handoff below. `make agent-finalize` passed at `20260813T034458Z-p500768`; `make test-fast` passed 355/355 at `20260813T034520Z-p503683`; `make build` passed 7/7 at `20260813T034520Z-p503732`; final `make check` passed 755/755 at `20260813T035234Z-p726834`; final `make release-check` passed 927/927 at `20260813T035357Z-p806092`. Static zero-reference, typed-storage, constructor-literal, transitional-file, and diff searches passed. | No stable blocker or remaining implementation work. A Jobs shutdown timing assertion missed once in the first final check at `20260813T034726Z-p587456`, passed 3/3 unchanged at `20260813T035209Z-p725462`, and the full check then passed. Retained-run maintenance was skipped because no source-digest-compatible canonical check root existed before broad validation; `agent-finalize` passed with `RESULTS_DIR` unset. G12-G18 are closed. |

## 20. Iteration 2 Session Handoff

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-12T20:18:22-04:00 | Codex / Iteration 2 planning and S11 document update | S11 done; S12 ready but inactive | Inspected the completed Evidence tracker, Domain vocabulary/ownership guidance, NLSpec planning guidance, current Evidence exports and callers, owner runtime, routes, cleanup, storage adapters, Incident Bundle portability, server/test assembly, verification owner guidance, and the completed Artifacts Iteration 2 pattern; touched only this tracker | Targeted read-only `rg`, `find`, `sed`, caller/export counts, `make task-guide ROLE=module-author OWNER=module.evidence`; `git diff --check -- docs/handoffs/evidence-module-refactor-tracker.md`; `make lint-markdown` at `.cartulary/test-results/20260813T002119Z-p1363390` | The next iteration is bounded to accidental API, composition, typed storage, cohesion, and retained negative evidence; diff and Markdown gates pass; Domain vocabulary is unchanged and no implementation began | None | A later authorized implementation session may activate S12 only, add characterization/export accounting, validate it, and update this tracker before S13. |
| 2026-08-12T20:58:11-04:00 | Codex / S12 characterization and exported-surface accounting | S12 done; S13 ready but inactive | Added exact root/subpackage AST export accounting, its verification contract and authored test row, generated topology output, and this checkpoint | Format/generate; Evidence plus twelve affected owner unit graphs; available service-backed graphs; exact Incident Bundles retry; boundary, harness, generated drift/policy, and JSON-shape gates | Retained behavior and every planned API disposition are executable; no owner, Domain, public contract, schema, migration, or data change is required | No stable blocker; one transient object-store fixture failure cleared unchanged; telemetry has no service-backed row | Pass the S12 tracker diff and Markdown gates, then activate S13 only. |
| 2026-08-12T21:26:54-04:00 | Codex / S13 owner runtime and application composition | S13 done; S14 ready but inactive | Consolidated Evidence production construction, captured route dependencies, server cleanup lifecycle, cross-owner runtime fixtures, private-service test bridges, export accounting, and this checkpoint | Format; Evidence/server/Workbook/Timeline/Imports/telemetry owner graphs; boundary and Evidence harness contract; static zero-consumer searches | One fallible owner constructor now produces every Evidence application capability; routes cannot obtain a second database, clock, or object store; cleanup starts only after readiness and closes in reverse order | No stable blocker; one unrelated Entities support row omitted the established upload session and failed closed | Pass the S13 tracker diff and Markdown gates, then activate S14 only. |
| 2026-08-12T21:41:18-04:00 | Codex / S14 typed object-storage cutover | S14 done; S15 ready but inactive | Narrowed the complete Evidence runtime and all operation paths to typed storage; added the fallible private-field Incident Bundle blob port and injected it once through server, routes, worker, builder, importer, and import transactions; updated tests, export accounting, and this checkpoint | Format; Evidence/server/Incident Bundles/object-store/telemetry unit or exact graphs; Evidence/server/Incident Bundles service-backed graphs; boundary and static fallback/literal searches | Incomplete raw adapters fail before readiness; all owner operations carry the fixed purpose; typed errors and bundle rollback behavior remain green; no compatibility facade or data migration was added | No stable blocker; one transient Incident Bundles seed failure passed unchanged on exact retry | Pass the S14 tracker diff and Markdown gates, then activate S15 only. |
| 2026-08-12T21:58:55-04:00 | Codex / S15 module-native API minimization | S15 done; S16 ready but inactive | Renamed the complete Evidence mutation contract and accessor, privatized unconsumed implementation surfaces and test hooks, removed unused `blobref` convenience APIs, migrated all consumers and tests, finalized exact export accounting, and updated this checkpoint | Format; exact export row; Evidence unit/service graphs; Workbook, Imports, server, Revisions, and Projections owner graphs; boundary, harness, generated drift, zero-name, zero-helper, and diff searches | Every remaining export has a production role; retired names have no Evidence consumer or alias; persisted and public compatibility contracts remain exact | No stable blocker; Timeline browser preflight and Incident Bundles external seeding produced retained infrastructure diagnostics outside the changed API paths | Pass the S15 tracker diff and Markdown gates, then activate S16 only. |
| 2026-08-12T22:20:18-04:00 | Codex / S16 cohesion and residue cleanup | S16 done; S17 ready but inactive | Decomposed catch-all production and integration files, removed the placeholder, renamed current-contract tests and private source operations, updated authored routing/boundary/provider/probe inputs, regenerated projections, and updated this checkpoint | Format/generate; exact export; Evidence unit/service; boundary, harness, drift, policy, shape, absence/sequence/selector searches; 355/355 fast graph | Final package layout follows capability boundaries without a new package seam or duplicate create/patch path; verification identities and all observable behavior remain stable | No stable blocker; fast validation caught and cleared one missing Import Assembly test import from the S15 consumer migration | Pass the S16 tracker diff and Markdown gates, then activate S17 only. |
| 2026-08-13T00:08:14-04:00 | Codex / S17 final validation and handoff | S11-S17 done; G12-G18 closed | Reran the complete owner matrix and repository gates; corrected the owning S13 Entities support fixture and S14 lint residue; reconciled exports, runtime topology, typed purposes, layout, routing, generated state, compatibility, deferrals, diagnostics, and this final checkpoint | Exact export and Evidence owner graphs; all affected unit/service owner graphs; frontend, browser, security, boundary, harness, drift, policy, and shape gates; `agent-finalize`; 355/355 fast; 7/7 build; 755/755 check; 927/927 release; zero-reference/storage/file/literal/diff searches | The complete refactor is production-ready. Public and persisted behavior is unchanged; internal Go compatibility was intentionally removed without aliases; all final APIs and storage operations have deliberate roles. | No stable blocker. Transient external fixture, browser contention, and Jobs timing diagnostics are recorded with passing retries and broad closing roots. Retained-run maintenance was skipped because no current source-digest-compatible canonical check existed before broad validation. | No further Iteration 2 work. Preserve Sections 13-22 as the final ledger and start a new adopted-owner plan for any deferred platform or feature change. |

## 21. Iteration 2 Deferrals and Blockers

| ID | Decision or blocker | Why it matters | Trigger or required authority | Status |
| --- | --- | --- | --- | --- |
| DEF-01 | Project-wide retirement of the platform `objectstore.Store` interface is outside Iteration 2. | Other owners still consume the broad interface; this iteration needs only to make Evidence typed-only. | Separate platform-owned reachability and migration plan. | DEFERRED |
| DEF-02 | Redesign of `workbookprojection` is outside Iteration 2. | It is an active Evidence/Projections published language, not dead code. | Separate cross-owner contract and migration plan before a projection-language version change. | DEFERRED |
| DEF-03 | Provider relocation without evidence of an ownership inversion is outside Iteration 2. | Current providers express legitimate source-owner seams and passed closure suites. | Reopen only if executable boundary evidence shows consumer-owned or cyclic construction. | DEFERRED |
| DEF-04 | New routes, schemas, lifecycle tokens, telemetry vocabulary, Reporting fields, or Domain terms are outside Iteration 2. | This is structural cleanup; adding behavior would invalidate the compatibility freeze. | Adopted owner amendment and tracker replan. | DEFERRED |
| None | No current stable blocker or remaining Iteration 2 work. | All slices and closing validation gates passed under the existing owner authority. | Start a new adopted-owner plan before any deferred platform, vocabulary, schema, or feature change. | CLEAR |

Any later normative contradiction MUST be recorded as `BLOCKED: owner
contradiction`. A required database migration or public behavior change also
blocks the affected slice and returns it to planning; it is not absorbed into
an implementation cleanup.

## 22. Iteration 2 Binary Completion Criteria

- [x] S00-S10 and Sections 2-12 remain available as the completed Iteration 1
  record.
- [x] Section 1 identifies Iteration 2 S11-S17 as the controlling sequence.
- [x] The delta inventory distinguishes dead/accidental surface from active
  provider and projection contracts.
- [x] G12-G18 include remediation, areas, rationale, benefit, compatibility,
  unresolved risk, and binary validation.
- [x] The compatibility posture preserves public/data/security behavior and
  rejects aliases or untyped Evidence fallbacks.
- [x] Domain vocabulary is recorded unchanged and NLSpec research remains
  non-authoritative planning doctrine.
- [x] S11 changes only this tracker and passes its diff and Markdown gates.
- [x] S12 characterizes retained behavior and installs exact export accounting.
- [x] S13 makes `OwnerRuntime` the only complete production composition path.
- [x] S14 removes every untyped Evidence storage path and binds exact purposes.
- [x] S15 leaves only the deliberate module-native exported surface, without
  aliases.
- [x] S16 removes the dead placeholder, transitional structure/wording, and
  catch-all file layout without duplicating behavior.
- [x] S17 records passing owner, cross-owner, security, generated, browser,
  broad, build, and release evidence.
- [x] No removed identifier, constructor, package path, or fallback has a shim
  or remaining consumer.
- [x] G12-G18 are closed, no required blocker remains, and the final tracker,
  implementation, verification routing, evidence roots, handoff, and checklist
  agree.

## 23. Iteration 3 Delta Inventory

Iteration 3 begins from the clean repository state below. It does not reopen a
completed Iteration 1 or Iteration 2 finding unless this section or Section 25
explicitly carries that work forward.

| Baseline item | Iteration 3 value |
| --- | --- |
| Source commit | `cee1a9b81adc` (`Collaboration Contract Closure and Runtime Hardening`) |
| Baseline time | 2026-08-28 |
| Worktree | Clean before the S18 tracker edit |
| Evidence Go inventory | 92 files: 60 production and 32 test files |
| Evidence size | 8,367 production Go lines and 7,719 test Go lines |
| Change since Iteration 2 handoff | 12 later commits touched `internal/modules/evidence`; the Iteration 2 inventory and export label are historical rather than current-state evidence |
| Owner guidance | `make task-guide ROLE=module-author OWNER=module.evidence` |
| Owner routing | 53 rows, including 35 service-backed rows, from `make explain-test-owner OWNER=module.evidence` |
| Baseline boundary evidence | `make backend-module-boundary-check` passed 3/3 at `.cartulary/test-results/20260829T014857Z-p2272001` |
| Baseline owner diagnostic | `make test-slice OWNER=module.evidence` reached 34/35 at `.cartulary/test-results/20260829T014758Z-p2225607`; the remaining Go service row reported an object-store capability/readiness fixture error before a product assertion |

The following current-state findings control this iteration:

| Area | Current evidence | Iteration 3 disposition |
| --- | --- | --- |
| Export accounting | `exported_surface_test.go` still describes the final Iteration 2 boundary, and the authored row remains `module.evidence.surface.iteration2_exported_surface_reachability`. | Replace the phase label with a steady-state exact surface contract and classify every surviving export by a real production consumer or contribution role. |
| Mutation admission | Evidence exposes `DecodeCreateRequest`, `DecodePatchRequest`, `DecodeConflictResolveRequest`, mutable request DTOs, and separate hash functions returning transport `APIError` values. | Adopt opaque owner admissions and a transport-neutral failure contract; Workbook performs HTTP failure translation. |
| Mutation composition | The mutation facade constructs peer-owner auth, incident-admission, Records, and revision-history implementations internally and carries complete `authn.UserRecord` values although it uses only the actor ID. | Inject narrow consumer-owned capabilities at application composition and carry only `uuid.UUID` actor identity. |
| Mutation outcome | `MutationResult` publishes `Payload`, `StatusCode`, and `Replayed`, mixing Evidence facts with Workbook transport representation. | Return semantic row facts and a closed outcome; Workbook reconstructs the byte-compatible response. |
| Projection boundary | `workbookprojection` combines source contribution inputs, Projections-owned mutation rows, association effects, descriptor helpers, and an aggregate `Ports` object. | Replace it atomically with source-owned `projectioncontract` and consumer-facing `projectionports`; no forwarding package remains. |
| Provider topology | Projections, Reporting, Recovery, operator code, and tools import public-looking Evidence provider packages directly. Delete/restore and rollback packages are imported only through the Evidence root. | Construct source-owner contributions at the Evidence root and move concrete providers under `internal/providers`. |
| Dead surface | The Reporting `CollectFieldsTx` wrapper has no production caller. Bundle helpers, a bundle concrete type, projection helpers, `blobref.MaxStorageKeyBytes`, and multiple lifecycle error/reason exports have no external production role. | Delete true dead code and privatize live owner-local behavior after the exact S19 caller ledger proves each disposition. |
| Cohesion | `blob_lifecycle.go` is 1,120 lines, `request_decoding.go` is 532 lines, and `lifecycle_integration_test.go` is 1,131 lines. | Split by semantic responsibility after the interface cutovers so structure follows the final design rather than the transitional API. |
| Test compatibility | `service_test_bridge_test.go` publishes more than 30 aliases, constructors, and helpers solely from package `evidence` to package `evidence_test`. | Remove the bridge; private service tests use owner-package fixtures and external suites exercise production contributions or routes. |

S19 fixed the following complete current-surface caller ledger. The exact AST
keys remain executable in `exported_surface_test.go`; this table gives every
key a disposition without treating the current package shape as compatibility:

| Current surface | Production role and fixed disposition |
| --- | --- |
| Evidence root runtime, cleanup, timeline, recovery-state, revision contribution, settings, view-schema, lifecycle/error contracts, and `NewIncidentBundleSourcePort` | Retain because application composition, Timeline, Workbook error translation, Recovery, Revisions, or Incident Bundles has a production consumer. Method sets may narrow only with their consuming contract. |
| Evidence root request DTOs, decoders, request hashes, conflict claims/value, commands, and `MutationResult` | Replace atomically in S20 with opaque admissions, semantic commands/results, and owner-owned durability types. `MutationContribution` remains the semantic root contribution but changes contract. |
| Root `ExportIncidentBundleFiles`, `ImportIncidentBundleFilesTx`, and concrete `IncidentBundleBlobPortability` | Privatize in S22 behind the root source-port and a source-owned interface. |
| `blobref` | Retain the typed reference/key/parser contract used by Evidence policy and storage; privatize `MaxStorageKeyBytes` in S22 because it has no external production caller. |
| `internal/policy` | Retain as the owner-private cross-package policy contract; it is not a consumer API. |
| `workbookprojection` | Replace and delete in S21; source contribution types move to `projectioncontract`, consumer mutation/effect types move to `projectionports`, and no forwarding package remains. |
| `projectionprovider` | Replace with root construction plus a private S21 provider. |
| `reportingprovider`, `recoveryprovider`, `deleterestore`, and `rollbackprovider` | Replace with root contributions/private providers in S22. Delete the zero-caller Reporting `CollectFieldsTx` wrapper. |

Searches covered production, application, command/tool, reusable test-support,
and test roots. The ledger is deliberately grouped by current package because
the AST allowlist supplies the exhaustive symbol list and rejects an unlisted
addition.

`docs/domain.md` remains vocabulary and owner-navigation authority within its
stated boundary; this iteration adds no Domain term. The research document
`docs/research/nlspec-spec.md` supplies non-authoritative planning guidance and
does not authorize product behavior. The user's request and cleanup principles
define this refactoring task, subject to the adopted owners listed in Section
1.2.

## 24. Iteration 3 Compatibility and Deletion Posture

Iteration 3 is an internal structural refactor. The following observable and
persisted behavior is frozen:

| Contract | Required compatibility |
| --- | --- |
| Public HTTP and OpenAPI | Preserve the six Evidence operations, request/response bodies, status codes, error codes and details, route precedence, and generated frontend protocol. |
| Authorization and security | Preserve Core 04 roles, incident-state rechecks, concealment, failure precedence, upload capability v2, token lifetimes and consumption, and continuing rejection of legacy version-1 upload tokens. |
| Mutation durability | Preserve operation strings, request hashes, stored JSON, lookup-before-validation, idempotency conflicts and replay, transaction/effect ordering, row versions, change-set facts, and Collaboration publication order. |
| Evidence lifecycle | Preserve upload, attachment, quarantine, custody, access-handle, cleanup, and deletion/restore state machines and failure behavior. |
| Projection behavior | Preserve descriptors except the intentional facade package path, surface field sets, canonical rows, refresh/query/rebuild results, association effects, and ordering. |
| Reporting | Preserve the adopted 12-member Evidence allowlist, omission and redaction rules, logical support paths, support-reference behavior, and exclusion of physical/storage/capability values. |
| Portability and recovery | Preserve Incident Bundle paths and bytes, rollback behavior, recovery object identities, ordering, row counts, and object-store semantics. |
| Data and owners | No schema, migration, stored-data rewrite, lifecycle value, telemetry vocabulary, Reporting field, route, OpenAPI, generated protocol, or Domain vocabulary change. |

Internal Go compatibility is intentionally not frozen. Each cutover MUST be
atomic across its in-repository callers. A removed identifier or package MUST
have no alias, forwarding package, deprecated wrapper, fallback, dual-write,
or parallel implementation. Durable strings and stored representations are
retained because they are data contracts, not Go compatibility scaffolding.

Features are carried forward only when they preserve an adopted behavior or
provide a deliberate future-facing owner seam. Zero-production-caller helpers,
phase labels, broad aggregate ports, transport-shaped owner types, public
concrete providers, and test-only compatibility exports do not meet that bar.

## 25. Iteration 3 Gap Matrix

| ID | Gap and evidence | Remediation | Benefit | Compatibility and risk | Binary validation | Status |
| --- | --- | --- | --- | --- | --- | --- |
| G19 | The exact export lock, file inventory, and routed row still represent Iteration 2 although 12 later commits changed Evidence. | Built the current AST/caller ledger, renamed the routed surface identity to steady-state language, and re-ran the frozen owner behavior before removal. | Makes deletion decisions executable and prevents both accidental retention and accidental API loss. | Test-only and authored/generated verification changes; product behavior is unchanged. Caller discovery covered app, tool, support, and test roots. | Exact surface and characterization passed; Evidence passed 35/35 and service-backed Evidence passed 25/25. | CLOSED S19 |
| G20 | Mutation admission returned HTTP errors, commands carried a full user and route strings, results carried HTTP payload/status, and services constructed peer-owner stores. | Added opaque admissions, semantic failures/outcomes, closed stored mutation results, actor IDs, and injected narrow capabilities used by mutation and attachment flows. Workbook now owns HTTP translation and application composition owns peer adapters. | Separates Evidence policy from transport and composition, makes dependencies testable, and permits future owner growth without embedding peer implementations. | Exact wire, hashes, persisted JSON, errors, lifecycle replay route/status identities, and transaction ordering remain fixed. No schema, migration, stored-data, owner-specification, or Domain change occurred. | Admission/hash and stored-byte parity, replay/conflict, 201/200 mapping, rollback, ordering, typed-nil rejection, all Evidence/Workbook and affected owner rows, security, frontend, and browser checks passed. | CLOSED S20 |
| G21 | `workbookprojection` mixed source contribution and consumer-owned mutation/effect directions behind an aggregate port. | Created `projectioncontract` and `projectionports`, added the root contribution constructor, internalized the source provider, and exposed explicit mutation-row and association-effect capabilities from Projections and application composition. | Clarifies direction, removes Workbook naming from a Projections contract, and scales without another aggregate facade. | The descriptor facade path changed internally; all provider IDs, field sets, SQL, canonical rows, ordering, refresh, effect, and rebuild behavior remained exact. No compatibility package or alias remains. | The old packages and aggregate port are absent; descriptor, row, query, refresh, effects, restore, owner/consumer slices, boundaries, harness, generated state, and JSON shape passed. | CLOSED S21 |
| G22 | Consumers reached through Evidence into five provider packages, and several exported wrappers/constants/types had no production caller. | Added root contribution constructors, injected a narrow logical-target provider, moved concrete providers under `internal/providers`, deleted dead wrappers, and privatized live owner-local symbols. | Restores source-owner construction, reduces the supported surface, and removes compatibility burden. | Provider outputs, bundle bytes, recovery order, and revision behavior remain exact. Typed-nil providers now fail closed before use. | Root constructors and typed-nil guards pass; direct provider imports, retired packages, dead wrappers, and owner-local exports are absent; affected owner and build matrices pass. | CLOSED S22 |
| G23 | Large mixed-responsibility files and the test bridge obscured lifecycle ownership and forced private services through exported test aliases. | Split lifecycle and admission production code by responsibility, moved private tests to the owner package with owner-neutral fixtures, retained route/contribution tests externally, split integration coverage by concern, and deleted the bridge. | Improves comprehension, focused testing, and extension without creating a second production path. | Structural only. Semantic test identities, routes, durable behavior, and full-stack coverage remain unchanged. | The bridge and private-provider test imports are absent; declarations remain routed; Evidence unit/service suites, boundaries, harness contracts, generated drift, and fast checks pass. | CLOSED S23 |
| G24 | Iteration 2 closing evidence predated the final implementation and the prior owner baseline contained one fixture/readiness diagnostic. | Reconciled fresh owner and service-backed graphs, cross-owner integrations, generated state, security/browser checks, broad repository gates, final paths, exports, routing, and retained evidence. | Provides production-readiness evidence for the actual final source rather than a historical snapshot. | S24 introduced no behavior, schema, data, vocabulary, or specification change. One unrelated browser timeout was explicitly infrastructure-classified and passed alone. | Required narrow and broad commands pass; final exports, topology, routing, generated outputs, deferrals, retained roots, and checklist agree. | CLOSED S24 |

## 26. Iteration 3 Execution Policy and Tracker Gate

S18 through S24 are strictly serial. The user authorized their implementation
on 2026-08-28. S19 through S24 passed their tracker gates and are complete.
A session MUST activate one slice at a time and
MUST NOT begin a successor until the current slice meets its exit criteria and
records its checkpoint in Sections 29 and 30.

Every implementation slice MUST:

1. Re-read its adopted owners and the current affected interfaces before edit.
2. Use `make task-guide ROLE=module-author OWNER=<owner-id>` and the narrowest
   owner rows before broader verification.
3. Preserve the Section 24 freeze and use an atomic internal cutover without a
   compatibility facade.
4. Run formatting only for implementation files, update authored routing or
   policy inputs when topology changes, and generate outputs only through Make.
5. Run the required slice gates, update the gap and checkpoint ledgers, then
   pass `git diff --check` for touched files and `make lint-markdown` for this
   tracker.

If an adopted owner contradicts the plan, record `BLOCKED: owner
contradiction`. If implementation requires a route, schema, migration,
lifecycle, telemetry, Reporting-field, stored-data, or Domain change, stop and
return to planning. Such a change is not an incidental cleanup. A transient
service fixture failure is retained with its run root, exact row, and ownership
evidence; a reproducible owner fixture defect is repaired in the owning slice.

The rollback boundary is the active slice. Do not retain old and new package
paths or APIs simultaneously as a rollback mechanism. Restore the slice to its
pre-slice implementation if its invariants cannot be met, while preserving the
diagnostic evidence in this tracker.

## 27. Iteration 3 Serial Workstreams

### S18 — Iteration 3 tracker rebaseline

- **Status:** DONE. The Section 28 document gates passed; no implementation
  slice was activated.
- **Areas:** This tracker only.
- **Remediation:** Update Section 1, record the current commit, inventory,
  authority posture, baseline diagnostics, gaps, compatibility freeze, serial
  workstreams, validation, deferrals, and binary exit criteria. Preserve
  Sections 2 through 22 as completed history.
- **Decisions:** Activate prior DEF-02 only for the S21 projection cutover and
  prior DEF-03 only for the S21/S22 source-owner construction cutovers. Continue
  deferring project-wide `objectstore.Store` retirement and every feature or
  public-contract change.
- **Exit evidence:** Only this file changed. `git diff --check --
  docs/handoffs/evidence-module-refactor-tracker.md` passed and
  `make lint-markdown` passed at
  `.cartulary/test-results/20260829T021259Z-p2282284`. S19 is ready but inactive
  and no implementation began.

### S19 — Characterization and steady-state surface lock

- **Status:** DONE. The phase-neutral authored row and generated topology agree,
  current exports have fixed dispositions, and the prior readiness diagnostic
  cleared on a complete service-backed rerun.
- **Areas:** Evidence surface tests, characterization, authored Evidence test
  routing, generated topology through Make when the row identity changes, and
  tracker checkpoints.
- **Remediation:** Replace the Iteration-2-named export row with a steady-state
  exact AST surface contract. Inventory root, `blobref`, projection contracts,
  and provider packages. Give every export one fixed disposition: retained
  consumer contract, retained root contribution constructor, privatize, or
  delete. Add or tighten characterization for admission failures and hashes,
  typed-nil dependency rejection, stored replay bytes and operation tags,
  projection descriptors/rows/effects, Reporting fields and logical targets,
  recovery inventory, bundle portability, and lifecycle transaction order.
- **Baseline diagnostic:** Re-run the failed Evidence row from
  `20260829T014758Z-p2225607`. Repair it only if a reproducible owner fixture
  defect is proven; otherwise retain the precise external readiness diagnosis
  and demonstrate the affected exact product rows.
- **Compatibility:** Characterization and verification identity only; no
  production behavior changes.
- **Exit:** G19 is closed, every later removal has executable before-state
  evidence, the renamed row and generated topology agree, and S20 is ready.
- **Exit evidence:** Targeted surface/admission/effect-order/provider rows passed
  at `.cartulary/test-results/20260829T025727Z-p2299703`; service-backed Evidence
  passed 25/25 at `.cartulary/test-results/20260829T025738Z-p2300161`; complete
  Evidence passed 35/35 at `.cartulary/test-results/20260829T025853Z-p2349069`;
  boundary passed 3/3 at `.cartulary/test-results/20260829T030003Z-p2398696`;
  generation drift passed 4/4 at
  `.cartulary/test-results/20260829T030007Z-p2399491`; the tracker diff check
  and Markdown lint passed at
  `.cartulary/test-results/20260829T030140Z-p2403131`.

### S20 — Semantic mutation and lifecycle boundary

- **Status:** DONE. S20 closed G20 with one atomic internal cutover; S21 is now
  active.
- **Areas:** Evidence admission, commands/results, idempotency, mutation and
  attachment services, Workbook adapters, server/test composition, and exact
  surface accounting.
- **Admission contract:** Replace exported mutable request/hash mechanics with
  immutable `CreateAdmission`, `PatchAdmission`, and
  `ConflictResolveAdmission` values. `AdmitCreateJSON`, `AdmitPatchJSON`, and
  `AdmitConflictResolveJSON` return those values plus an Evidence-owned
  `AdmissionFailure`. The failure publishes only field, reason, collection,
  and count accessors required for Workbook translation. Zero-value admissions
  fail before persistence.
- **Command and result contract:** Commands carry `ActorUserID`, the admission,
  request ID, and time. Evidence owns its durable operation selection. Results
  carry the row, closed outcome (`created`, `updated`, `kept_saved`, or
  `replayed`), semantic IDs, row version, view schema, and changed field keys;
  they do not carry HTTP status or a prebuilt payload. Workbook recreates the
  exact existing response, including 201 create, 200 patch/conflict, change-set
  omission, and replay signaling.
- **Capabilities:** Add `MutationDependencies` with the minimum methods
  Evidence currently uses for incident-state admission, route idempotency,
  record-envelope insert/load/version advance, revision append/window,
  projection mutation rows, association effects, conflict fields and
  keep-saved idempotency, and Collaboration publication. Application
  composition supplies adapters. Mutation and attachment services do not
  construct `authn.Store`, incident admission, Records, or revision-history
  implementations.
- **Durability:** Introduce a closed operation-tagged stored-mutation result and
  typed idempotency key/record. Preserve the exact database JSON, operation
  strings, request hashes, lookup-before-validation, replay, conflict, commit,
  and publication order.
- **Compatibility:** Intentional internal Go break without aliases; public,
  persisted, lifecycle, authorization, and projection behavior remains exact.
- **Exit:** G20 is closed; the old decoders, hash helpers, full-user commands,
  transport-shaped results, and internal peer-store construction are absent;
  Evidence, Workbook, server, Records, Revisions, Collaboration, and affected
  Projections checks pass.
- **Completed surface:** `CreateAdmission`, `PatchAdmission`, and
  `ConflictResolveAdmission` now retain owner-validated immutable request and
  hash state. `AdmissionFailure`, `OperationID`, `MutationOutcome`,
  `MutationResult`, typed stored mutation records, and narrow mutation and
  lifecycle idempotency capabilities replace the former public DTO/hash/HTTP
  surface. Workbook application adapters construct incident, auth
  idempotency, Records, revision-history, projection, conflict, and
  Collaboration capabilities and reconstruct the unchanged HTTP response.
  Blob create/attach likewise consume injected incident, record-envelope, and
  lifecycle idempotency capabilities rather than constructing peer stores.
- **Compatibility result:** The durable operation strings
  `workbook.rows.create`, `workbook.records.patch`,
  `workbook.records.conflicts.resolve`, `object_blobs.create`, and
  `evidence.attach_blob`, their existing request hashes, response JSON/status
  records, external statuses/envelopes, lifecycle effects, and ordering were
  preserved. The old Go APIs were removed without aliases or dual paths.
- **Exit evidence:** Complete Evidence passed 36/36 at
  `.cartulary/test-results/20260829T033828Z-p2602015`; service-backed Evidence
  passed 25/25 at `.cartulary/test-results/20260829T033947Z-p2652666`;
  Workbook passed 68/68 at
  `.cartulary/test-results/20260829T034101Z-p2701491`; Projections passed 19/19
  at `.cartulary/test-results/20260829T034317Z-p2756463`; Revisions passed
  27/27 at `.cartulary/test-results/20260829T034401Z-p2774220`; Records passed
  8/8 at `.cartulary/test-results/20260829T035011Z-p2939335`; Collaboration
  passed 31/31 at `.cartulary/test-results/20260829T034659Z-p2841343`;
  `build-server`, boundary 3/3, generation drift 4/4, and targeted gosec 4/4
  passed at `20260829T034834Z-p2891263`, `20260829T034644Z-p2840878`,
  `20260829T034848Z-p2903751`, and `20260829T034859Z-p2906809`. Frontend
  typecheck and 390/390 unit checks passed at `20260829T034913Z-p2937101` and
  `20260829T034928Z-p2937660`; browser webserver-backed 60/60 and stateful
  34/34 passed at `20260829T035101Z-p2956380` and
  `20260829T035647Z-p3012807`. The tracker diff check passed and Markdown lint
  passed at `20260829T040011Z-p3061534`. The first Records run exposed only the related
  lexical boundary-token collision `EvidenceStored` at
  `20260829T034518Z-p2819242`; helper renaming fixed it before the clean rerun.

### S21 — Projection contract and port cutover

- **Status:** DONE. S21 closed G21 atomically; S22 is now active.
- **Areas:** Evidence projection language and provider, Projections adapters and
  runtime, projection assembly, server, Imports, Timeline, Recovery, tests,
  boundary policy, routing, and tracker.
- **Source contract:** Move projection inputs, paging, source reader,
  contribution, descriptor construction, surface intent, and canonical row
  mapping to `internal/modules/evidence/projectioncontract`. Add root
  `evidence.NewProjectionContribution`; its concrete source moves to
  `internal/modules/evidence/internal/providers/projection`.
- **Consumer ports:** Move row refresh/load and Evidence association-effect
  interfaces and DTOs to `internal/modules/evidence/projectionports`. Remove
  the aggregate `Ports`; Projections and projection assembly expose explicit
  Evidence mutation-row and association-effect capabilities.
- **Deletion:** Delete `internal/modules/evidence/workbookprojection` in the
  same cutover. Update the descriptor facade package, all imports, boundary
  allowlists, provider manifests, authored tests, and generated topology. Do
  not add a forwarding package or type aliases.
- **Compatibility:** Projection field sets, canonical rows, query/refresh,
  association effects, provider ordering, restore rebuild, and incident
  rebuild remain exact.
- **Exit:** G21 is closed; the old path and aggregate port are absent; Evidence,
  Projections, server, Workbook, Imports, Timeline, and Recovery gates pass.
- **Completed surface:** `projectioncontract` owns `ProjectionInput`, paging,
  `SourceReader`, `Contribution`, descriptor/surface intent, and canonical
  `ViewRow`. `projectionports` owns the explicit `MutationRows` and
  `AssociationEffects` capabilities and effect DTOs. Root
  `evidence.NewProjectionContribution` is the only production construction
  entry point; its concrete SQL source is private under
  `internal/providers/projection`. Projections/application composition now
  publish two named capabilities and no aggregate Evidence projection port.
- **Compatibility result:** The descriptor facade package intentionally moved
  to `internal/modules/evidence/projectioncontract`. The authored provider
  manifest and generated topology digest were updated; provider identity,
  authority modules, table, order, capabilities, canonical rows, SQL,
  transaction boundaries, association effects, and rebuild behavior are
  unchanged. `workbookprojection` and `projectionprovider` are absent without
  a forwarder, alias, or fallback.
- **Exit evidence:** Projections passed 19/19 at
  `.cartulary/test-results/20260829T041313Z-p3154408` and service-backed 12/12
  at `.cartulary/test-results/20260829T041357Z-p3171847`; Evidence passed 36/36
  at `.cartulary/test-results/20260829T041442Z-p3189127` and service-backed
  25/25 at `.cartulary/test-results/20260829T041603Z-p3239147`; Imports passed
  23/23 at `.cartulary/test-results/20260829T041717Z-p3287898`; Timeline passed
  53/53 at `.cartulary/test-results/20260829T041833Z-p3331927`; Workbook passed
  68/68 at `.cartulary/test-results/20260829T042413Z-p3406369`; Recovery passed
  24/24 at `.cartulary/test-results/20260829T042655Z-p3461310`; and
  `build-server` passed at `20260829T042311Z-p3389401`. Boundary 3/3,
  generation drift 4/4, JSON shape 3/3, harness contract 2/2, and generated
  artifact policy 3/3 passed at `20260829T041121Z-p3131277`,
  `20260829T042331Z-p3401934`, `20260829T042342Z-p3404870`,
  `20260829T042350Z-p3405379`, and `20260829T042407Z-p3405911`. Transitional
  runs `20260829T040716Z-p3075020`, `20260829T040835Z-p3105515`,
  `20260829T041052Z-p3130703`, and `20260829T041131Z-p3131671` exposed only
  related incomplete-cutover references or allowlists; each was fixed before
  its clean rerun.

### S22 — Provider and dead-surface cleanup

- **Areas:** Evidence root/provider topology, Reporting and reporting assembly,
  Recovery/operator/tool composition, Revisions, Incident Bundles, `blobref`,
  boundary policy, tests, and tracker.
- **Reporting contributions:** Add
  `NewReportingFieldContribution() exportprovider.FieldProvider` and
  `NewReportingLogicalTargetContribution()` returning a new narrow
  `exportprovider.LogicalSupportTargetProvider`. Change
  `reportingassembly.NewLinksProvider` to validate, sort, and merge providers
  while failing closed for nil, typed-nil, empty-key, duplicate-key, and
  conflicting-target inputs; Reporting and reporting assembly no longer
  import an Evidence provider package directly.
- **Recovery and revision contributions:** Add
  `NewRecoveryProvider(postgres.DB) recovery.EvidenceRecoveryProvider`. Keep the
  existing root revision contribution. Move delete/restore, rollback,
  projection, reporting, and recovery concrete implementations below
  `internal/providers` and update server, operator, tools, tests, and boundary
  policy atomically.
- **Deletions:** Remove the unused Reporting `CollectFieldsTx` wrapper. Make the
  Incident Bundle export/import functions and concrete blob-port type private.
  Unexport `blobref.MaxStorageKeyBytes` and live owner-local lifecycle
  errors/reason constants. Delete any zero-reachability symbol proven dead in
  S19; retain cross-owner errors such as blob visibility and object-store
  translation only where a production consumer remains.
- **Compatibility:** Reporting output, logical paths, provider keys, bundle
  bytes, recovery order/counts, revision behavior, and construction failures
  remain exact. No compatibility packages or deprecated wrappers remain.
- **Exit:** G22 is closed; all surviving exports have a production role; direct
  provider-package imports and retired names are absent; affected owner and
  tool/build checks pass.
- **Completed surface:** Root Evidence now constructs Reporting field and
  logical-target contributions, Recovery's `EvidenceRecoveryProvider`, the
  existing Revisions contribution, projection contribution, and the
  source-owned Incident Bundle portability interface. Concrete Reporting,
  Recovery, delete/restore, rollback, projection, and Incident Bundle
  providers are private under `internal/providers`; only their root files may
  import them. Reporting's narrow `exportprovider` package owns the logical
  target interface. Reporting assembly accepts any number of providers, sorts
  by stable key, merges deterministically, and rejects nil, typed-nil,
  empty-key, duplicate-key, and conflicting-target contributions. Server,
  operator, recovery tool, Reporting, Recovery tests, Revisions, and Incident
  Bundles use only root Evidence constructors or consumer interfaces.
- **Deleted/private surface:** Public `reportingprovider`, `recoveryprovider`,
  `deleterestore`, and `rollbackprovider` packages are gone. The unused
  Evidence Reporting `CollectFieldsTx`, exported Incident Bundle export/import
  helpers, public bundle concrete type, `blobref.MaxStorageKeyBytes`, six
  owner-local attachment reason constants, and four owner-local lifecycle
  sentinels are removed or private. The permanent surface test also requires
  the S21 `projectionprovider` and `workbookprojection` paths and all four S22
  provider paths to be physically absent.
- **Compatibility result:** Reporting still emits exactly the adopted twelve
  Evidence members and no forbidden storage/private fields; provider key
  `evidence`, logical `/evidence/<record-id>` targets, fallback behavior,
  recovery inventory order/count, revision snapshot/rollback behavior,
  Incident Bundle paths/bytes/staging/cleanup, HTTP behavior, SQL, data, and
  transaction ordering are unchanged. No schema, migration, stored-data,
  telemetry, owner specification, Domain, forwarding package, alias,
  deprecated wrapper, or dual path was added.
- **Exit evidence:** Evidence passed 36/36 at
  `.cartulary/test-results/20260829T044445Z-p3534112`, service-backed Evidence
  passed 25/25 at `.cartulary/test-results/20260829T045324Z-p3792687`, the final
  provider/typed-nil row passed at `20260829T050704Z-p4076962`, and the final
  absence row passed at `20260829T045930Z-p3962815`. Reporting passed 6/6 at
  `20260829T044610Z-p3585244`, service-backed 4/4 at
  `20260829T045239Z-p3775736`, and its final exact Evidence/provider subset
  passed 3/3 at `20260829T050352Z-p4053376`. Recovery passed 24/24 at
  `20260829T050043Z-p3988951` and service-backed 19/19 at
  `20260829T045544Z-p3858648`; Revisions passed 27/27 at
  `20260829T044953Z-p3711969` and service-backed 20/20 at
  `20260829T045716Z-p3912864`; Incident Bundles passed 8/8 at
  `20260829T045124Z-p3757917` and service-backed 6/6 at
  `20260829T045441Z-p3841494`. Server and operator builds passed at
  `20260829T045954Z-p3963443` and `20260829T050012Z-p3976158`. Final format,
  boundary, generation drift, generated policy, and JSON shape passed at
  `20260829T050655Z-p4072837`, `20260829T050342Z-p4052947`,
  `20260829T050205Z-p4043471`, `20260829T050219Z-p4046458`, and
  `20260829T050225Z-p4046914`; the tracker Markdown gate passed at
  `20260829T050943Z-p4079083`. Transitional Recovery run
  `20260829T044658Z-p3602747` exposed a related test import shadow and the
  first absence run `20260829T045904Z-p3961543` exposed retained empty S21
  directories; both were corrected before clean reruns. The first format
  graph classified an unsorted authored Reporting row as `artifact_error`;
  the row was sorted before the clean final format run.

### S23 — Cohesion and test-boundary cleanup

- **Areas:** Evidence lifecycle/admission implementation layout, integration
  suites, private fixtures, authored test routing, generated topology, exact
  surface/boundary guards, and tracker.
- **Production cohesion:** `blob_lifecycle.go` is now the 231-line dependency
  and type core; slot/lease, attachment, quarantine, custody, persistence, and
  projection-effect behavior live in cohesive files of 330 lines or fewer.
  Common mutation admission values remain centralized while create, patch,
  conflict, and failure behavior live in separate files. The lifecycle
  revision dependency was tightened to the existing narrow append port so
  owner tests do not construct a peer application.
- **Test boundary:** Deleted `service_test_bridge_test.go` and its more than 30
  aliases/helpers. Attachment, blob idempotency, handle-state, cleanup,
  quarantine-service, validation, and direct-decoding white-box coverage now
  uses package `evidence` with owner-neutral PostgreSQL fixtures and narrow
  local capability adapters. Genuine root-contribution, HTTP, authorization,
  projection, and composition coverage remains in package `evidence_test`.
  The local test projection source is obtained through
  `NewProjectionContribution`; no test imports the private provider.
- **Suite cohesion:** Full-stack test bodies are separated into boundary,
  projection/Collaboration, authorization, cleanup, and quarantine files.
  Shared fixtures are explicitly named `lifecycle_test_support_test.go` and
  contain no test body. The quarantine row gained the owner-private lifecycle
  test while retaining the two full-stack scenarios; generated topology was
  refreshed through `make generate`.
- **Boundary and compatibility result:** The authored Records-envelope policy
  now names the three split files that retain the pre-existing reads. The old
  bridge, direct private-provider test import, stale integration filename, and
  replacement aliases are absent. HTTP/OpenAPI, stored strings and JSON,
  hashing, replay, transaction ordering, SQL/schema/data, projection and
  Collaboration effects, lifecycle behavior, owner specifications, Domain,
  telemetry, and frontend behavior are unchanged.
- **Exit evidence:** Final formatting passed at
  `.cartulary/test-results/20260829T055125Z-p287959`; Evidence passed 36/36 at
  `.cartulary/test-results/20260829T055133Z-p292114`; service-backed Evidence
  passed 25/25 at `.cartulary/test-results/20260829T054726Z-p220410`; and the
  isolated quarantine row passed 3/3 at
  `.cartulary/test-results/20260829T054523Z-p153975`. Boundary, harness,
  generated drift, and JSON shape passed at
  `.cartulary/test-results/20260829T055252Z-p341877`,
  `.cartulary/test-results/20260829T055252Z-p341903`,
  `.cartulary/test-results/20260829T055252Z-p341627`, and
  `.cartulary/test-results/20260829T055252Z-p341657`. Generated-artifact policy
  passed at `.cartulary/test-results/20260829T055327Z-p346081`, the
  Evidence-specific harness contract exited zero, and `make test-fast` passed
  441/441 at `.cartulary/test-results/20260829T055327Z-p346344`. The tracker
  diff check passed and Markdown lint passed at
  `.cartulary/test-results/20260829T055550Z-p357820`.
- **Exit:** G23 is closed. No product or specification repair was required;
  S24 may begin after this tracker passes its diff and Markdown gate.

### S24 — Final validation and handoff

- **Areas:** Final source and export inventory, provider and dependency
  topology, behavior freeze, verification routing, generated state, security,
  browser behavior, build/release evidence, tracker, and handoff.
- **Final inventory:** Evidence contains 115 Go files: 82 production and 33
  test files, totaling 9,033 production and 7,803 test lines. Six private
  provider imports exist only in their matching Evidence root construction
  files. The permanent surface row is
  `module.evidence.surface.exported_surface_reachability`. The bridge and the
  `workbookprojection`, public projection, Reporting, Recovery,
  delete/restore, and rollback provider directories are physically absent.
- **Owner evidence:** Evidence, server, Workbook, Projections, Reporting,
  Recovery, Revisions, Incident Bundles, Records, Collaboration, Imports,
  Timeline, object storage, and telemetry passed 36/36, 24/24, 68/68, 19/19,
  6/6, 24/24, 27/27, 8/8, 8/8, 31/31, 23/23, 53/53, 6/6, and 2/2. The roots,
  in that order, are `20260829T063618Z-p1675040`,
  `20260829T055651Z-p360257`, `20260829T055651Z-p360246`,
  `20260829T055941Z-p527800`, `20260829T063618Z-p1675045`,
  `20260829T060025Z-p546088`, `20260829T060025Z-p546076`,
  `20260829T060025Z-p546085`, `20260829T060221Z-p679286`,
  `20260829T060221Z-p679290`, `20260829T060221Z-p679307`,
  `20260829T060221Z-p679320`, `20260829T060703Z-p846517`, and
  `20260829T060703Z-p846523` under `.cartulary/test-results/`.
- **Service-backed evidence:** The same owners through object storage passed
  25/25, 17/17, 39/39, 12/12, 4/4, 19/19, 20/20, 6/6, 5/5, 22/22,
  14/14, and 5/5 at roots `20260829T060748Z-p864203`,
  `20260829T060748Z-p864189`, `20260829T060748Z-p864199`,
  `20260829T060748Z-p864194`, `20260829T061017Z-p1026133`,
  `20260829T061017Z-p1026147`, `20260829T061017Z-p1026164`,
  `20260829T061017Z-p1026170`, `20260829T061207Z-p1158178`,
  `20260829T061207Z-p1158176`, `20260829T061207Z-p1158194`, and
  `20260829T061810Z-p1367848`. Timeline reached 28/30 at
  `20260829T061207Z-p1158182` solely because one pre-existing browser support
  row received an infrastructure timeout; that row passed alone 11/11 at
  `20260829T061714Z-p1326661`, and all later broad browser/release graphs
  passed. Telemetry has no authored service-backed row.
- **Architecture, security, and browser evidence:** Generation drift,
  generated-artifact policy, JSON shape, migration drift, and backend
  boundaries passed at `20260829T061907Z-p1385294`,
  `20260829T061907Z-p1385374`, `20260829T061907Z-p1385331`,
  `20260829T061907Z-p1385380`, and `20260829T061907Z-p1385714`. Harness,
  targeted gosec, frontend unit, and frontend typecheck passed at
  `20260829T061921Z-p1392815`, `20260829T061921Z-p1392940`,
  `20260829T061921Z-p1392760`, and `20260829T061921Z-p1392742`;
  `make harness-evidence-contract` also exited zero. Webserver-backed and
  stateful browser graphs passed 60/60 and 34/34 at
  `20260829T061948Z-p1425261` and `20260829T062557Z-p1481865`.
- **Broad and retained evidence:** `make test-fast` passed 441/441 at
  `20260829T062847Z-p1533416`; the corrected `make check` passed 671/671 at
  `20260829T063734Z-p1741972`; retained-root `make agent-finalize` passed at
  `20260829T064232Z-p1862244`; build passed 7/7 at
  `20260829T064252Z-p1865301`; and release passed 837/837 at
  `20260829T064312Z-p1899899`. The final tracker diff/Markdown gate passed at
  `20260829T065852Z-p2117282`, and documentation-compatible retained-root
  finalization passed at `20260829T065910Z-p2118309`.
- **Compatibility and exit:** S24 added no product behavior. Public HTTP and
  OpenAPI, authorization/concealment, durable mutation strings/hashes/JSON,
  transaction ordering, SQL/schema/data, lifecycle, projection, Reporting's
  exact twelve members, recovery, revision, bundle, telemetry, owner
  specifications, and Domain remain unchanged. G19-G24 are closed; no retired
  API/path has a caller, alias, forwarder, fallback, or dual implementation.

## 28. Iteration 3 Validation Matrix

| Layer | Command or scenario | Required point |
| --- | --- | --- |
| Document-only gate | `git diff --check -- docs/handoffs/evidence-module-refactor-tracker.md`; `make lint-markdown`; status review | S18 and every later tracker checkpoint. |
| Owner guidance | `make task-guide ROLE=module-author OWNER=<owner-id>`; `make explain-test-owner OWNER=<owner-id>` | Before selecting verification for an affected owner. |
| Formatting | `make format` | S19-S23 when authored Go or frontend source changes; never for S18. |
| Evidence owner | `make test-slice OWNER=module.evidence`; `make service-backed-test-slice OWNER=module.evidence` | S19-S24. |
| Mutation boundary | Malformed/unknown/limited and zero-value admissions; request-hash parity; stored JSON/operation parity; replay/conflict; 201/200 mapping; attachment rollback; lifecycle and transaction ordering | S19 characterization, S20 cutover, and S24. |
| Mutation consumers | Unit and service-backed slices for Workbook, server, Records, Revisions, Collaboration, Imports, Timeline, and Projections as affected | S20-S24. |
| Projection boundary | Descriptor/intent, canonical row, query, refresh, association effects, provider order, restore and incident rebuild; zero old import/path | S19, S21, and S24. |
| Provider boundary | Reporting 12-member allowlist and forbidden-value absence; logical support paths/fallback; bundle byte/path parity; recovery inventory/order; revision closure | S19, S22, and S24. |
| Export and absence | Exact AST allowlist; no retired symbol, direct provider import, aggregate projection port, test bridge, alias, forwarder, or deprecated wrapper | S19-S24. |
| Boundaries and harness | `make backend-module-boundary-check`; `make harness-evidence-contract`; `make harness-contract` | Every topology/routing slice and S24. |
| Generated state | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make migration-drift` | Changed-input slices and S24. |
| Frontend and browser | `make frontend-unit`; `make frontend-typecheck`; affected stateful and webserver-backed browser targets | S20 when adapter behavior changes and S24. |
| Security | `make go-gosec-targeted`; upload-capability, concealment, authorization, and failure-precedence scenarios | S20 and S24. |
| Final repository | `make agent-finalize`; `make test-fast`; `make check`; `make build`; `make release-check` | S24. |

S24 MUST rerun the unit and available service-backed owner graphs for Evidence,
server, Workbook, Projections, Reporting, Recovery, Revisions, Incident Bundles,
Records, Collaboration, Imports, Timeline, object store, and telemetry when
affected by the final dependency graph. `make agent-finalize` uses a retained
successful `RESULTS_DIR` when a source-digest-compatible full check exists;
otherwise the tracker records why retained-run maintenance was skipped.

## 29. Iteration 3 Slice Checkpoint Ledger

| Slice | Status | Intended or completed change | Required evidence | Blocker and next action |
| --- | --- | --- | --- | --- |
| S18 | DONE | This tracker only: Section 1 update and Sections 23-32 addition; no implementation, generated, contract, or other documentation change. | Tracker-only diff check passed; `make lint-markdown` passed at `20260829T021259Z-p2282284`; worktree review found only this tracker changed. | No blocker. S19 remains inactive pending later implementation authorization. |
| S19 | DONE | Renamed the permanent surface row, regenerated topology, fixed the caller/disposition ledger, and revalidated current behavior. | Targeted rows passed; service-backed Evidence 25/25; full Evidence 35/35; boundary 3/3; generation drift 4/4; tracker diff and Markdown lint passed. | No blocker. The former object-store readiness diagnostic did not reproduce. |
| S20 | DONE | Replaced mutable request/hash/HTTP mutation APIs with opaque admissions and semantic results; injected peer-owner mutation and lifecycle capabilities; moved exact HTTP and durable idempotency adaptation to application composition. | Evidence 36/36 and service-backed 25/25; Workbook 68/68; Projections 19/19; Revisions 27/27; Records 8/8; Collaboration 31/31; server build, boundary, drift, gosec, frontend, and browser checks passed. | No blocker. No specification, Domain, schema, migration, stored-data, telemetry, or external compatibility change. |
| S21 | DONE | Split source contribution language from consumer mutation/effect ports; added root construction/private source provider; removed aggregate ports and obsolete packages; updated descriptor, manifest, boundaries, routing, topology, and all callers. | Evidence and Projections unit/service graphs; Workbook, Imports, Timeline, Recovery; server build; boundary, harness, generated/drift, JSON, and absence checks passed. | No blocker. No specification, Domain, schema, SQL, stored-data, or external behavior change. |
| S22 | DONE | Added root Reporting/Recovery/provider construction, deterministic logical-target composition, private concrete providers, a source-owned bundle interface, typed-nil guards, and exact dead-package/symbol removal. | Evidence, Reporting, Recovery, Revisions, and Incident Bundle unit/service graphs; server/operator builds; format, surface, boundary, generation drift, generated policy, and JSON shape passed. | No blocker. Provider output, portability, recovery, revisions, HTTP, SQL, and data behavior remain exact. |
| S23 | DONE | Split lifecycle/admission code and lifecycle integration coverage by responsibility; removed the alias bridge; moved private tests to owner-neutral white-box fixtures; refreshed routing and boundary policy. | Evidence 36/36, service-backed 25/25, boundary 3/3, harness 2/2, drift 4/4, JSON 3/3, generated policy 3/3, and fast checks 441/441 passed. | No blocker. G23 is closed with structural-only compatibility. |
| S24 | DONE | Reconciled final source/export/provider inventory, every requested owner and service-backed graph, security/browser/drift evidence, retained full-check evidence, build, release, compatibility, and handoff. | Section 28 passed; `check` 671/671, build 7/7, and release 837/837 are retained with exact roots. | No blocker. Three discovered test/lint defects returned to S22/S23 and passed affected and broad reruns. Iteration 3 is complete. |

## 30. Iteration 3 Session Handoff

| Time | Agent/session | Current state | Files inspected or touched | Commands and evidence | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-28T22:09:48-04:00 | Codex / Iteration 3 planning and S18 document update | S18 done; S19 ready but inactive | Inspected the completed tracker, Domain and NLSpec research posture, live Evidence exports/callers, mutation/lifecycle composition, projection/provider/reporting/recovery seams, verification routing, and current git state; touched only this tracker | Read-only inventory and caller searches; `make task-guide ROLE=module-author OWNER=module.evidence`; `make explain-test-owner OWNER=module.evidence`; Evidence baseline 34/35 at `20260829T014758Z-p2225607`; boundary 3/3 at `20260829T014857Z-p2272001`; tracker diff passed; Markdown lint passed at `20260829T021259Z-p2282284` | Iteration 3 is bounded to current surface accounting, semantic mutation dependencies, projection direction, source-owner provider construction, dead code, cohesion, and final evidence; S18 passed with no product/specification/data change or implementation | No stable blocker. One baseline service row reached an object-store readiness fixture error before product assertions. | Leave S19 inactive. A later authorized implementation session may activate S19 only, reconcile the baseline diagnostic, install characterization, and update this tracker before S20. |
| 2026-08-28T23:00:50-04:00 | Codex / S19 characterization and surface lock | S19 done; S20 ready | Updated the steady-state surface test, authored Evidence row, generated topology index, and this tracker; inspected all direct provider imports and mutation/error consumers | Targeted characterization `20260829T025727Z-p2299703`; service-backed Evidence 25/25 `20260829T025738Z-p2300161`; full Evidence 35/35 `20260829T025853Z-p2349069`; boundary 3/3 `20260829T030003Z-p2398696`; generation drift 4/4 `20260829T030007Z-p2399491` | G19 closed with exhaustive AST accounting and grouped caller dispositions; no product behavior or normative owner changed | None. The prior object-store readiness failure did not reproduce. | Pass the tracker diff/Markdown gate, then activate S20 only. |
| 2026-08-29T00:00:00-04:00 | Codex / S20 semantic mutation and lifecycle boundary | S20 done; S21 active | Replaced the Evidence mutation admission, commands/results, idempotency, composition, Workbook adapters, and affected tests; injected lifecycle incident/record/idempotency capabilities; updated the authored Evidence row, generated topology, surface lock, and this tracker | Evidence 36/36 `20260829T033828Z-p2602015`; service-backed Evidence 25/25 `20260829T033947Z-p2652666`; Workbook 68/68 `20260829T034101Z-p2701491`; Projections 19/19 `20260829T034317Z-p2756463`; Revisions 27/27 `20260829T034401Z-p2774220`; Records 8/8 `20260829T035011Z-p2939335`; Collaboration 31/31 `20260829T034659Z-p2841343`; boundary/drift/gosec/frontend/browser roots recorded in S20 exit evidence | G20 closed. External HTTP, durable hashes/JSON/operation strings, data, lifecycle, projection, revision, and publication behavior are unchanged; old internal mutation APIs and peer construction are absent | None. One related lexical boundary-policy failure was fixed and cleanly rerun. | Pass the tracker diff/Markdown gate, then execute S21 only. |
| 2026-08-29T00:27:00-04:00 | Codex / S21 projection contract and port cutover | S21 done; S22 active | Added Evidence `projectioncontract`, `projectionports`, root contribution, and private projection source; removed `workbookprojection`, public `projectionprovider`, and aggregate ports; updated Projections/application/test composition, provider manifest, boundary policy, authored rows, generated topology, surface locks, and this tracker | Evidence/Projections unit and service-backed roots plus Workbook, Imports, Timeline, Recovery, server, boundary, harness, drift, JSON, and generated-policy roots are recorded in S21 exit evidence | G21 closed. Directional ownership is explicit, the old packages have zero references, and projection/data behavior is unchanged | None. Four related transitional compile/allowlist/policy failures were fixed before clean reruns. | Pass the tracker diff/Markdown gate, then execute S22 only. |
| 2026-08-29T01:07:52-04:00 | Codex / S22 provider and dead-surface cleanup | S22 done; S23 active | Added root Reporting and Recovery constructors, Reporting logical-target composition and its authored row, private Reporting/Recovery/delete-restore/rollback/bundle providers, typed-nil guards, and bundle source interface; removed public providers, dead wrapper/helpers/constants/errors; updated server/operator/tool/tests/boundaries/topology/surface and this tracker | Evidence/Reporting/Recovery/Revisions/Incident Bundle unit and service roots, exact provider/absence rows, server/operator builds, format, boundary, drift, generated policy, and JSON roots are recorded in S22 exit evidence | G22 closed. Provider reach-through and dead surfaces are absent; Reporting, recovery, revisions, bundle, HTTP, SQL, and data compatibility remain exact | None. One related Recovery test import shadow, two retained empty package directories, and one authored-row ordering diagnostic were corrected before clean reruns. | Pass the tracker diff/Markdown gate, then execute S23 only. |
| 2026-08-29T01:54:31-04:00 | Codex / S23 cohesion and test-boundary cleanup | S23 done; S24 active | Split lifecycle/admission production files and lifecycle integration bodies, renamed the shared support-only fixture file, removed the alias bridge, relocated owner-private tests, added narrow owner-local fixtures/adapters, updated the quarantine row, Records-envelope boundary paths, generated topology, and this tracker | Exact run roots are recorded in S23 exit evidence; final Evidence 36/36, service-backed 25/25, boundary, both harness contracts, generation drift, JSON shape, generated policy, and fast 441/441 checks passed | G23 closed. No alias bridge, private-provider test reach-through, duplicate implementation, route/data/specification change, or lost routed scenario remains | None. The initial import cycle, local SQL parameter mismatch, scheduler fixture-policy mismatch, and expired service-session diagnostic were corrected or owner-classified before clean reruns. | Pass the tracker diff/Markdown gate, then continue S24 validation only. |
| 2026-08-29T02:58:32-04:00 | Codex / S24 final validation and handoff | Iteration 3 done | Reconciled the final 115-file Evidence inventory, root/private provider topology, permanent surface/routing, all requested owner and service-backed graphs, drift/security/frontend/browser/broad/release evidence, compatibility conclusions, diagnostics, checklist, and this tracker | Exact roots are recorded in S24 exit evidence; final check passed 671/671, retained-root finalization passed, build passed 7/7, and release passed 837/837 | G24 and G19-G24 closed. No Evidence-owned failure, stale generated output, retired path/caller, compatibility shim, specification/data change, or unclassified diagnostic remains | None. A stale S23 test caller expectation and six dead helpers plus two S22 lint strings were corrected and broadly revalidated; one unrelated Timeline browser timeout passed in isolation and in later broad browser/release graphs. | Pass the final tracker diff/Markdown gate; hand off the completed Iteration 3 baseline. |

## 31. Iteration 3 Deferrals and Blockers

| ID | Decision or blocker | Iteration 3 treatment | Status |
| --- | --- | --- | --- |
| I3-DEF-01 | Project-wide retirement of `objectstore.Store` affects owners outside Evidence. | Keep Evidence typed-only and leave the platform-wide migration to a separate platform-owned plan. | DEFERRED |
| I3-DEF-02 | Prior DEF-02 deferred redesign of `workbookprojection`. | Completed as the internal S21 directional split supported by the adopted Projections boundary; no projection behavior or product language changed. | CLOSED S21 |
| I3-DEF-03 | Prior DEF-03 deferred provider relocation without ownership evidence. | Completed where live caller evidence showed consumer reach-through or root-only construction; S21/S22 provide atomic root constructors and private implementations. | CLOSED S22 |
| I3-DEF-04 | New routes, schemas, migrations, lifecycle values, upload tokens, telemetry vocabulary, Reporting fields, generated protocol, or Domain terms. | Excluded. Any such need blocks the affected slice and requires an adopted-owner amendment and new plan. | DEFERRED |
| I3-DIAG-01 | The S18 baseline Evidence slice reached 34/35 because one Go service row reported object-store capability/readiness fixture failure. | S19 reran the complete service-backed graph successfully at `20260829T025738Z-p2300161` and the full Evidence graph at `20260829T025853Z-p2349069`; no fixture or product repair was warranted. | CLOSED S19, TRANSIENT NOT REPRODUCED |
| I3-DIAG-02 | The first S23 Evidence graph failed 30/36 because same-package tests still imported downstream application composition, creating an Evidence import cycle. | Replaced downstream composition with owner-neutral database fixtures and narrow local capability adapters; the final graph passed 36/36 at `20260829T055133Z-p292114`. | CLOSED S23, TEST-BOUNDARY DEFECT |
| I3-DIAG-03 | Transitional S23 graphs exposed a local revision-adapter SQL parameter used as both text and UUID, a transaction fixture hardwired into a template-clone row, and an expired reusable test-service descriptor. | Gave the UUID array its own typed SQL parameter, made the local store honor the authored fixture policy, and refreshed the zero-borrower test-service session. Product assertions then passed; the readiness-only run remains `.cartulary/test-results/20260829T054012Z-p88011`. | CLOSED S23, FIXTURE DEFECTS; INFRA SESSION REFRESHED |
| I3-DIAG-04 | The first S24 Projections graph reached 18/19 because its caller matrix still expected the removed `EvidenceMutationRows()` use in the split lifecycle test. | Returned the stale expectation to S23, removed it after a repository-wide zero-caller scan, and passed Projections 19/19 plus harness 2/2 at `20260829T055941Z-p527800` and `20260829T055941Z-p527902`. | CLOSED S23 DURING S24, ROUTING ASSERTION |
| I3-DIAG-05 | Timeline service-backed validation reached 28/30 because the pre-existing full keyboard/clipboard browser row exceeded its 60-second limit under the aggregate load. | The harness classified the row as infrastructure timeout. It passed alone 11/11 at `20260829T061714Z-p1326661`; later webserver-backed 60/60, stateful 34/34, check 671/671, and release 837/837 graphs passed. | CLOSED S24, TRANSIENT INFRA TIMEOUT |
| I3-DIAG-06 | The first S24 `make check` reached 670/671 because staticcheck found two capitalized S22 Reporting error strings and six S23 helpers left unused after test relocation. | Returned the defects to S22/S23, normalized the internal strings, deleted the zero-caller helpers, passed affected Evidence/Reporting graphs, and passed the corrected check 671/671 at `20260829T063734Z-p1741972`. | CLOSED S22/S23 DURING S24, LINT CLEANUP |

No adopted-owner contradiction or stable implementation blocker remains.
Iteration 3 is complete.

## 32. Iteration 3 Binary Completion Criteria

- [x] Sections 2-12 remain the completed Iteration 1 record and Sections 13-22
  remain the completed Iteration 2 record.
- [x] Section 1 identifies Sections 23-32 and S18-S24 as the controlling
  Iteration 3 plan.
- [x] The baseline records commit `cee1a9b81adc`, the clean pre-edit worktree,
  the 92/60/32 file inventory, later Evidence changes, and exact diagnostic
  roots.
- [x] Domain vocabulary is unchanged; NLSpec research is explicitly
  non-authoritative; the user's request is distinguished from document
  content.
- [x] G19-G24 state remediation, benefit, compatibility risk, and binary
  validation.
- [x] The plan freezes public and persisted behavior and rejects internal
  aliases, forwarders, dual paths, deprecated wrappers, and fallbacks.
- [x] S18 changes only this tracker and passes its diff and Markdown gates.
- [x] S19 installs current characterization and steady-state surface/routing
  accounting and closes G19.
- [x] S20 installs semantic admission/results and injected narrow mutation and
  lifecycle capabilities and closes G20.
- [x] S21 replaces `workbookprojection` with directional contracts/ports and
  closes G21 without a forwarding package.
- [x] S22 constructs providers at the Evidence root, removes consumer
  reach-through and dead surface, and closes G22.
- [x] S23 removes the test bridge and mixed-responsibility layout while
  retaining one routed test for every frozen behavior and closes G23.
- [x] S24 records passing narrow and broad evidence, reconciles the final
  repository and tracker, and closes G24.
- [x] Every removed identifier and package has zero callers and no compatibility
  replacement; all retained exports have a deliberate production role.
- [x] G19-G24 are closed, no required blocker remains, and the final source,
  verification routing, generated outputs, evidence roots, deferrals, handoff,
  and checklist agree.
