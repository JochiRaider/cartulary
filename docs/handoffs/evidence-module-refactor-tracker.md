# evidence Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

### 1.1 Target and change boundary

| Item | Normative planning value |
| --- | --- |
| Target path | `internal/modules/evidence` |
| Target label | `evidence`, derived from the final path segment and normalized to lowercase kebab case |
| Output path | `docs/handoffs/evidence-module-refactor-tracker.md` |
| Status | Planning and handoff only |
| Gap closure | Design and authority closure, not implementation completion |
| Future structural work | S-00 through S-06 remain unimplemented behavior-preserving work |
| Future behavior correction | S-07 remains unimplemented and separately authorized |
| Permitted writes in this revision | This tracker, Core 01, Core 04, and the Reporting NLSpec |
| Prohibited writes | Code, tests, migrations, contracts, generated artifacts, configuration, harness inputs, and git staging |

`MUST`, `MUST NOT`, `SHALL`, and `REQUIRED` in this tracker govern only the
refactor workflow, evidence gates, slice ordering, and handoff discipline. They
do not create product behavior. Adopted specifications remain the sole product
authority within their scopes. If this tracker conflicts with an adopted owner,
the owner controls and the affected work MUST record
`BLOCKED: owner contradiction`.

This tracker authorizes no implementation. A later implementation task MUST be
explicitly authorized and MUST recheck the live repository and adopted owners
before changing code, tests, migrations, contracts, generated artifacts,
configuration, or harness inputs.

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

Repository evidence inspected includes all 58 target files in Section 2 and
the exact live Evidence routes, admission, tokens, handles, Store, runtime,
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
| Core 02 and Core 03 | Existing product authority | Lifecycles, cleanup deadlines, timing, retention | Worker topology or claim schema |
| This tracker | Refactor guidance | Interfaces, dependencies, private worker design, characterization, rollback, acceptance gates | Product requirements |
| Testing Harness inputs | Verification accounting | Semantic identities, selectors, topology, retained evidence | Runtime architecture |
| Analysis notes and NLSpec research guide | Non-normative research | Recommendations and style input | Adopted behavior |
| Generated contracts/topology | Machine projection | Tool-produced outputs | Hand-authored authority or manual edits |

The framework/live mismatch remains a planning finding: the live module has
typed Workbook, Projections, Timeline, Revisions, Imports, Recovery, Reporting,
and Incident Bundle contributions beyond the framework's generic catalog.
Those seams MUST be judged by semantic ownership, not directory depth.

## 2. Current-State Repository Inventory

The target contains 58 files: 42 production Go files, 15 Go test files, and one inert placeholder. Every file is inventoried below. Generated contracts are read-only dependencies or downstream outputs; none is to be hand-edited.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/evidence/.gitkeep` | Historical directory placeholder | None | None | None | None | None | None | Low | Explicitly out of behavioral scope because it has no content or runtime effect. |
| `internal/modules/evidence/access_repository.go` | Reads access snapshots, persists opaque handles, and atomically redeems/revalidates them | Package-private `Store` methods | `routes.go` through `Store` | PostgreSQL, evidence/blob/access-handle tables | `handles_test.go`, `evidence_integration_test.go`, `integration_test.go` | Core 01 handle contract; Evidence OpenAPI | Evidence persistence | Critical | Authorization-sensitive state is rechecked at redemption. |
| `internal/modules/evidence/api.go` | Strictly decodes blob, attach, and handle requests; computes idempotency hashes and response helpers | `BlobCreateRequest`, `AttachBlobRequest`, decode/hash functions | Evidence route handlers and tests | `httpapi`, JSON, hashing, media/filename helpers | `blob_create_test.go`, `attach_test.go`, `handles_test.go`, conformance/integration tests | Evidence OpenAPI and error registry | Evidence transport adapter | High | Public Go names are internal-repository API, not external wire authority. |
| `internal/modules/evidence/attach_test.go` | Unit and service-backed characterization for attach validation, concealment, uniqueness, and races | Test-only `BlobOptions`, `DurableCounts`, helpers, three tests | Go test and `module.evidence` harness rows | PostgreSQL test support, Evidence store | Self | Evidence view/lifecycle and uniqueness contracts | Evidence tests | Medium | Test-only exports are not production package design evidence. |
| `internal/modules/evidence/blob_create_test.go` | Characterizes object-blob create validation, idempotency, size ceilings, and preview payload limits | Four tests | Go test and Evidence harness rows | Evidence service/store test doubles | Self | OpenAPI request/response and configured limit contracts | Evidence tests | Medium | Freezes rejection precedence and replay behavior. |
| `internal/modules/evidence/blobref/blobref.go` | Canonical physical object key and logical `object://` reference formatting/parsing | `StorageKeyParts` and key/ref parse, build, and validation functions | Evidence persistence, upload, portability, tests, object-store probe | UUID and string validation | `blobref_test.go`, portability and object-store tests | Storage-reference and non-leakage contracts | Evidence | High | Physical key and logical reference must remain distinct and never leak publicly. |
| `internal/modules/evidence/blobref/blobref_test.go` | Exact key/ref round-trip and rejection characterization | Two tests | Go test and Evidence unit row | `blobref` | Self | Core 01 storage-reference rules | Evidence tests | Low | Protects canonical `incidents/.../object-blobs/...` and `object://...` forms. |
| `internal/modules/evidence/collaboration_intents.go` | Converts primary and affected-row changes into ordered Collaboration intents | Package-private helper | Mutation coordinator, Workbook/store mutation paths | Collaboration intent contract | Coordinator and integration/WebSocket tests | `record_changed` protocol semantics | Evidence producer; Collaboration transports | High | Evidence produces facts; it does not own the WebSocket hub. |
| `internal/modules/evidence/conflict_snapshot.go` | Maps Evidence view fields to revision conflict snapshot source keys | Package-private mapping helpers | Workbook conflict resolution | Revisions conflict/source contracts | Workbook conflict and conformance coverage | Evidence view schema and revision snapshot contract | Evidence | High | Field-key closure must not drift from the authored view schema. |
| `internal/modules/evidence/conformance_test.go` | Checks reason-code registry and Evidence public field-key closure | `ErrorRegistry`, helper functions, two tests | Go test and Evidence unit rows | Authored error registry and view schema loaders | Self | `contracts/errors/index.json`, Evidence view schema | Evidence tests | Medium | Reads authored contracts; generated output remains tool-owned. |
| `internal/modules/evidence/create_validation.go` | Enforces minimum create signal, reserved storage reference policy, lifecycle initialization, and direct patch admission | `ValidateWorkbookCreateParams`, `ValidateWorkbookDirectPatchChange` | Workbook/import mutation paths | Owner-local mutation types and blobref policy | Characterization, lifecycle, integration tests | Evidence view/create contract | Evidence | Critical | Authoritative Evidence mutation policy. |
| `internal/modules/evidence/create_validation_characterization_test.go` | Matrix for adopted create-signal and lifecycle rules | One test | Go test and Evidence unit row | Create validator | Self | Core 01/02/03 create contract | Evidence tests | Low | Baseline required before reshaping mutation code. |
| `internal/modules/evidence/deleterestore/provider.go` | Supplies Evidence source snapshot and restoration view resolution to Revisions delete/restore | `Source`, `NewSource`, provider methods | `RevisionProviderContribution` | Revisions delete/restore contracts | Provider contract and cross-owner Revisions tests | Revision provider/snapshot contracts | Evidence contribution; Revisions coordinates | Medium | Thin source-owner provider is appropriate. |
| `internal/modules/evidence/evidence_blob_uniqueness_migration_test.go` | Verifies the head migration contains the unique live blob association constraint | One integration test | Evidence service-backed row | PostgreSQL catalog | Self | Migration `00017_evidence.sql` | Evidence tests | Medium | Migration is authored source and must not be edited for a structural refactor. |
| `internal/modules/evidence/evidence_integration_test.go` | HTTP/service-backed tests for blob creation, handles, reason codes, strict members, and OpenAPI route presence | Seven tests | Evidence service-backed harness rows | Server harness, object store, database, public client | Self | Evidence OpenAPI, errors, generated protocol behavior | Evidence tests | High | Primary wire/security characterization. |
| `internal/modules/evidence/handles_test.go` | Unit characterization for issue idempotency, current-state recheck, disposition, and filename sanitation | Four tests | Evidence unit rows | Access repository/service test doubles | Self | Handle OpenAPI and security contract | Evidence tests | Medium | Includes single-use download and reusable preview expectations. |
| `internal/modules/evidence/import_create.go` | Adapts enabled Evidence view imports to a caller-owned transaction and Evidence create behavior | `ImportCreateCommand` alias, `NewImportCreateFacade` | Import assembly owner registry | Imports owner facade, Records, Projections, Revisions, Collaboration | Import assembly and integration tests | `contracts/imports/view-targets.v1.json` | Evidence contribution; Imports coordinates | High | Concrete construction is broad; public import behavior belongs to Imports. |
| `internal/modules/evidence/import_projection.go` | Refreshes and reads Evidence projection results during import creation | Package-private helper | Evidence import facade | Evidence projection rows | Import tests | Evidence view/projection contract | Evidence contribution | High | Must preserve caller transaction and projection result semantics. |
| `internal/modules/evidence/incident_bundle_blob_portability.go` | Exports object bytes, rewrites/stages imported keys, and cleans up staged bytes on failure | `IncidentBundleBlobPortability` and methods | Evidence incident-bundle source port and portability assembly | Object store, blobref, incident portability contracts | Cross-owner portability tests | Incident-bundle source catalog | Evidence contribution; Incident Bundles coordinates | Critical | Cross-resource rollback and physical-key concealment are security/data-integrity sensitive. |
| `internal/modules/evidence/incident_bundle_portability.go` | Exports/imports Evidence, blob, access, and custody relations as deterministic NDJSON files | `ExportIncidentBundleFiles`, `ImportIncidentBundleFilesTx` | Evidence source port | PostgreSQL, incident portability types | Provider contract and cross-owner round-trip tests | Incident-bundle source catalog | Evidence contribution | Critical | Exact relation/path/version/attribution behavior must be frozen. |
| `internal/modules/evidence/incident_bundle_source_port.go` | Declares the Evidence incident-bundle source descriptor and callbacks | `NewIncidentBundleSourcePort` | Incident portability assembly | Incident Bundles source-port contract | Provider contract and cross-owner catalog tests | `contracts/incident-bundles/source_catalog.json` | Evidence contribution | Medium | Generic bundle coordination remains outside Evidence. |
| `internal/modules/evidence/incident_bundle_subtype_presence.go` | Contributes Evidence subtype-presence validation for Records portability | `IncidentBundleSubtypeContribution` | Incident portability assembly | Records subtype-presence contract | Provider contract and portability tests | Incident-bundle/Records invariants | Evidence contribution | Medium | Evidence owns subtype fact; Incident Bundles coordinates. |
| `internal/modules/evidence/initial_blob_create.go` | Validates, observes, finalizes, and associates an optional initial blob during atomic generic row creation | Package-private helpers | Workbook create facade | PostgreSQL, object store, blob slots, lifecycle policy | Atomic create, lifecycle, concealment tests | Generic Workbook create and Evidence lifecycle contracts | Evidence application service | Critical | Preserves one transaction and unique association while object observation is external. |
| `internal/modules/evidence/integration_test.go` | End-to-end module characterization for upload/attach, generic create, projection rebuild, WebSocket refresh, expiry, handle invalidation, and quarantine | Nine tests plus test-only event/count helpers | Evidence integration/service-backed rows | Full app support, object store, PostgreSQL, WebSocket client | Self | HTTP, Workbook, view, projection, collaboration, lifecycle contracts | Evidence tests | Critical | Contains current evidence for cross-owner atomic effects. |
| `internal/modules/evidence/lifecycle_test.go` | Proves Evidence-record and Object Blob lifecycles remain distinct | One test and test-only `Queryer` | Evidence unit row | Store/lifecycle code, PostgreSQL stubs | Self | Core 02 lifecycle machines | Evidence tests | Medium | Essential boundary characterization. |
| `internal/modules/evidence/mutation_coordinator.go` | Orders incident admission, source write, projection refresh, revisions, and collaboration intents within caller transaction | Package-private coordinator | Workbook and import mutation paths | Narrow ports from `mutation_ports.go` | Coordinator and integration tests | Workbook, projection, revision, collaboration contracts | Evidence application service | Critical | Caller owns transaction/idempotency; ordering is observable. |
| `internal/modules/evidence/mutation_coordinator_test.go` | Fault-boundary test for exact mutation side-effect order | One test | Evidence unit row | Fake coordinator ports | Self | Mutation atomicity/effect-order contract | Evidence tests | Low | Primary non-database sequencing characterization. |
| `internal/modules/evidence/mutation_ports.go` | Defines narrow incident, Records, source, projection, revision, and collaboration mutation capabilities | Package-private interfaces/types | Mutation coordinator and composition | Peer-owner published types | Coordinator/composition/integration tests | No generated output directly | Evidence boundary layer | High | Existing boundary seam; must not become a generic transaction facade. |
| `internal/modules/evidence/objectstore_dependency_test.go` | Verifies public Evidence responses do not expose storage keys, refs, or object-store details | Test-only object-store admin plus one test | Evidence security/integration row | Server harness, object store, public clients | Self | OpenAPI public response and Core 04 concealment rules | Evidence tests | High | Security regression evidence, not production API. |
| `internal/modules/evidence/persistence_components.go` | Componentized blob slot/read/lifecycle, association, and cleanup-candidate SQL | Package-private repository components | `Store` and owner composition | PostgreSQL Evidence tables | Store, lifecycle, integration tests | Evidence schema/lifecycle contracts | Evidence persistence | Critical | Legitimate direct SQL, but broad root composition obscures individual capabilities. |
| `internal/modules/evidence/projectionprovider/contribution.go` | Constructs the Evidence projection-provider contribution | `NewContribution` | Projection assembly | `workbookprojection` contribution | Projection contribution and assembly tests | `contracts/projection-providers/index.json` | Evidence contribution; Projections coordinates | Medium | Source owner constructs provider; Projections validates catalog and stores rows. |
| `internal/modules/evidence/projectionprovider/source.go` | Reads authoritative Evidence, blob, link, and record facts into typed projection input | `Source`, `NewSource`, reader methods | Evidence projection contribution | PostgreSQL and typed Evidence projection input | Projection and integration tests | Evidence view schema/projection provider contract | Evidence source reader | High | It must not own physical projection storage. |
| `internal/modules/evidence/provider_contract_test.go` | Exact assertions for source-owner recovery, revision, portability, and subtype contributions | One test | Evidence unit row | Evidence contribution constructors | Self | Recovery, revision, incident-bundle catalogs | Evidence tests | Medium | Provides partial provider closure, not full runtime proof. |
| `internal/modules/evidence/recovery_inventory.go` | Builds Evidence vNext object inventory entries for recovery validation | `VNextRecoveryObjectInventory` | Recovery/operator assembly and tooling | Object store, blobref, recovery contracts | Cross-owner recovery tests | Recovery object-inventory contracts | Evidence contribution; Recovery coordinates | High | Physical object existence/integrity is security and recoverability sensitive. |
| `internal/modules/evidence/recovery_state.go` | Declares Evidence authoritative tables for state recovery | `RecoveryStateContribution` | Recovery assembly | Recovery-state contribution contract | Provider contract and Recovery tests | Recovery state catalog/fixtures | Evidence contribution | Medium | Exact relation set must remain aligned with portability. |
| `internal/modules/evidence/recoveryprovider/provider.go` | Restores Evidence object bytes/state through a Recovery provider | `Provider`, `New`, provider methods | Operator recovery composition | PostgreSQL, object store, blobref, recovery provider contracts | Cross-owner Recovery/operator tests | Recovery verification contracts | Evidence contribution; Recovery/operator coordinates | Critical | Must preserve validation, cleanup, and no-raw-key behavior. |
| `internal/modules/evidence/reportingprovider/provider.go` | Supplies redacted Evidence fields/facts for Reporting export | `CollectFieldsTx`, `CollectFactsTx` | Reporting export materializer | PostgreSQL, Reporting provider contract | Reporting boundary/export tests | Reporting provider/content-class contracts | Evidence contribution; Reporting coordinates | High | Deliberately excludes blob hash, object ID, and storage ref. |
| `internal/modules/evidence/revision_append_port.go` | Adapts Evidence mutation facts to the Revisions appender | Package-private adapter | Workbook/import/store mutations | Revisions appender and snapshot types | Coordinator/integration tests | Revision/change-set contracts | Evidence contribution | High | Must retain caller transaction and exact source snapshot semantics. |
| `internal/modules/evidence/revision_provider_contribution.go` | Registers Evidence route, snapshot, delete/restore, and rollback capabilities with Revisions | `RevisionProviderContribution` | Revision assembly | Revisions provider contracts and Evidence providers | Provider contract and cross-owner Revisions tests | Revision provider catalog/snapshot contract | Evidence contribution; Revisions coordinates | High | Contribution identity and route coverage are stable internal contracts. |
| `internal/modules/evidence/rollbackprovider/provider.go` | Validates snapshot values and restores Evidence source rows during rollback | `Provider`, `NewProvider`, methods | Evidence revision contribution | PostgreSQL, Revisions rollback/source contracts | Local provider and Revisions rollback tests | Evidence revision snapshot contract | Evidence contribution | High | Lifecycle and reference validation duplicates some root policy. |
| `internal/modules/evidence/rollbackprovider/provider_test.go` | Unit tests snapshot presence/association and invalid lifecycle/reference rejection | Two tests | Evidence unit row | Rollback provider | Self | Revision snapshot and lifecycle contract | Evidence tests | Low | Characterizes fail-closed rollback admission. |
| `internal/modules/evidence/route_admission.go` | Performs route authorization, membership/role, CSRF, and incident admission | Package-private service helpers | Evidence handlers | `authn`, incident access repository, `httpapi` | HTTP, handle, object-store tests | Core 04 authorization and OpenAPI errors | Evidence transport edge | Critical | Admission is correctly edge-local; deployment admin alone is insufficient. |
| `internal/modules/evidence/route_dependencies.go` | Adapts route service dependencies, including object-store upload/read behavior | Package-private adapters | `Service` construction and routes | Object store, settings, store capabilities | Blob/attach/handle/integration tests | Upload-target and payload-limit contracts | Evidence transport/platform adapter | High | Platform dependency is appropriate here, not in domain validation. |
| `internal/modules/evidence/route_errors.go` | Translates internal Evidence failures into stable public error/reason envelopes | Package-private mapping helpers | Evidence handlers | `httpapi`, owner errors | Conformance, HTTP, handle tests | `contracts/errors/index.json`, OpenAPI responses | Evidence transport adapter | Critical | Preserve concealment and precedence exactly. |
| `internal/modules/evidence/routes.go` | Registers and serves six Evidence HTTP operations | `Service`, `Settings`, `RouteService`, `WithRouteService`, `RegisterRoutes`, key-error classifier | Server assembly and route tests | Route admission/dependencies, `httpapi`, object store | Blob, attach, handle, integration, OpenAPI tests | Evidence OpenAPI owner and generated protocol surfaces | Evidence transport adapter | Critical | Thin over `RouteService`; boundary policy forbids private persistence and consumer orchestration here. |
| `internal/modules/evidence/runtime.go` | Composes the server-only Evidence owner runtime and publishes narrow Workbook/Timeline contributions | `TimelineAttachmentContribution`, `OwnerRuntime`, constructors | Server assembly and test composition | Store, revisions, collaboration, projections, object store | Server composition and provider tests | Runtime/boundary policy | Evidence application facade | High | `OwnerRuntime` is intentionally server-constructed only. |
| `internal/modules/evidence/source_kernel.go` | Creates authoritative Records/Evidence source state, refreshes projection, and reads the result in caller transaction | Package-private kernel | Workbook/import mutation coordinator | PostgreSQL, Records/source/projection ports | Create/import/coordinator/integration tests | Records envelope and Evidence projection contracts | Evidence application/persistence boundary | Critical | Must not begin a generic transaction or orchestrate consumers. |
| `internal/modules/evidence/store.go` | Broad coordinator for blob slots, attach/quarantine/cleanup, projections, revisions, collaboration, and access handles | Error/result/data types, options, `NewStore`, exported `Store` methods | Routes, owner runtime, tests; cross-package production construction is forbidden | PostgreSQL, object store, projections, revisions, collaboration, auth | Most Evidence unit/integration tests | HTTP, lifecycle, view, projection, revision, collaboration contracts | Mixed Evidence persistence/application service | Critical | Largest coupling hotspot. `QuarantineBlob` and cleanup are tested but have no production caller. |
| `internal/modules/evidence/store_test.go` | Unit mapping test for source data to Evidence projection row | One test | Evidence unit row | Package-private row mapper | Self | Evidence view schema | Evidence tests | Low | Exposes the duplicate row-rendering seam discussed in Section 5. |
| `internal/modules/evidence/timeline_facts.go` | Reads Evidence attachment facts for Timeline projection use | `TimelineFact`, `TimelineFactReader`, methods | Timeline assembly/collection reads | PostgreSQL, blobref | Timeline assembly/store tests | Timeline attachment presentation contract | Evidence contribution; Timeline coordinates | High | Evidence owns attachment facts; Timeline owns row composition/presentation. |
| `internal/modules/evidence/timeline_port.go` | Validates Timeline attachment references against Evidence records | Package-private implementation of published contribution | `NewTimelineAttachmentContribution`, Timeline assembly | PostgreSQL | Timeline assembly and workbook tests | Timeline attachment mutation contract | Evidence contribution | High | No Timeline implementation import; narrow source-owner port is appropriate. |
| `internal/modules/evidence/upload_token.go` | Creates, hashes, parses, and validates opaque upload capability tokens | Package-private token helpers | Blob create/upload routes and store | Cryptography, UUID/time | Blob create and integration tests | Opaque same-origin upload-target contract | Evidence security adapter | Critical | Raw object-store credentials/keys must never replace this token. |
| `internal/modules/evidence/workbook_conflict.go` | Resolves same-field conflict changes and revalidates Evidence source state | `WorkbookConflictCommand` and facade method | Workbook assembly/mutation API | Revisions conflict workflow, PostgreSQL, Evidence validation | Workbook/Revisions conflict tests | Generic conflict route and Evidence field contract | Evidence application contribution | Critical | Conflict token, row version, and source validation are observable. |
| `internal/modules/evidence/workbook_contribution.go` | Publishes the narrow Evidence create/patch/conflict facade to Workbook | `WorkbookContribution`, `NewWorkbookContribution` | Workbook assembly and app test support | Workbook facade, revisions, projections, conflict idempotency | Server/workbook composition and integration tests | Evidence view target and Workbook mutation contracts | Evidence contribution; Workbook coordinates | High | Appropriate public seam; construction accepts many dependencies. |
| `internal/modules/evidence/workbook_facade.go` | Owns Evidence generic create/patch transaction, idempotency, conflict, source/revision/projection/collaboration sequencing | Facade, commands, requests, results, conflict errors, constructor | Workbook contribution, import path, tests | PostgreSQL and several peer-owner ports/stores | Validation, atomic create, conflict, integration tests | Workbook routes, Evidence view schema, revisions/events | Mixed Evidence application facade | Critical | Another major coupling hotspot; preserve generic Workbook envelopes. |
| `internal/modules/evidence/workbook_mutations.go` | Defines writable field values, lifecycle patches, validation errors, and direct Evidence SQL mutations | Field/create/lifecycle/error types and `ValidLifecycleState` | Workbook facade, rollback/conflict validation | PostgreSQL, blobref, owner policy | Lifecycle/create/rollback/integration tests | Evidence view and lifecycle contracts | Evidence | Critical | Lifecycle vocabulary is repeated in rollback and frontend projections. |
| `internal/modules/evidence/workbookprojection/contribution.go` | Publishes typed Evidence projection input, rows/rebuilder ports, descriptor, and surface intent | `Rows`, input/page/source/rebuilder/ports/contribution types and constructors | Projection provider/assembly, Evidence facade, test support | Projection provider contract and Evidence view semantics | Local contribution, projection assembly/query tests | Projection-provider index and Evidence view schema | Evidence published language; Projections stores | High | Correct source/consumer boundary; generated view facts remain authoritative. |
| `internal/modules/evidence/workbookprojection/contribution_test.go` | Validates required source, exact descriptor/intent, unknown-view rejection, and defensive copies | Four tests | Evidence unit row | Projection contribution | Self | Projection provider and view-schema contracts | Evidence tests | Low | Protects the published provider seam. |

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
| Blob slot, upload, finalization, quarantine capability, cleanup eligibility | Store, routes, persistence | Evidence capability; platform jobs dispatch | Split | Core 01-03 and live methods | S-07 is independent and separately authorized. |
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

S-02 MUST introduce this future internal boundary with live repository types:

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
`workbookprojection.Rows`. S-02 MUST add
`SupportProjectionEffects SupportProjectionEffectsTx` to a future
`OwnerRuntimeDependencies`, update internal callers atomically, and provide no
compatibility shim.

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

S-00 MUST treat the canonical Evidence projection reader as oracle.
`module.evidence.attach_row_projection_parity` MUST capture the current
attach/quarantine mapper, read the same row through `workbookprojection.Rows`
inside the same transaction, commit, query the public route, and capture
committed `record_changed` effects.

Fixtures MUST cover link counts 0, 1, and greater than 1 with deleted links
excluded; every legal lifecycle/blob combination; null/present time, party,
storage, and hash fields; initial/advanced versions; first success/replay; and
injected projection/history/publication failures. Compare exact `view_row_v1`
serialization, complete cells, null versus omission, group values, order,
duplicates, linked counts, public query output, and event patch/invalidate.

Zero mismatches permit S-01. Implementation drift is repaired toward the
canonical reader. Owner conflict yields `BLOCKED: owner contradiction`.

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Evidence is a legitimate source boundary. | Core/source/migration | Dissolution scatters lifecycle. | `intentional/no_action` | Evidence | Preserve. |
| Store combines persistence, lifecycle, projections, revisions, events, handles. | Root Store/runtime | Broad blast radius. | `should_fix` | Evidence | S-03 after S-00. |
| Local mapper duplicates canonical Rows. | Store versus Rows | Row drift. | `must_fix` | Evidence/Rows | S-00 then S-01. |
| Evidence hard-codes peer projection layouts. | Support helper | Schema drift. | `must_fix` | Projections | S-00 then S-02. |
| Evidence source SQL is legitimate. | Source reads/writes | Generic move erases ownership. | `intentional/no_action` | Evidence | Retain source SQL. |
| Platform imports are edge-local. | Current imports/guard | Domain leakage if spread. | `intentional/no_action` | Edge/platform | Retain static guard. |
| Workbook/Import composition is broad. | Facades/assembly | Partial construction. | `should_fix` | Assembly/narrow ports | S-04 atomic update. |
| Lifecycle/reference policy repeats. | Live/rollback code | Transition drift. | `should_fix` | Evidence | S-05. |
| Reporting live implementation subtracts fields. | Provider SQL | Future leak. | `must_fix` | Reporting/Evidence | Adopt allowlist; S-06 proof. |
| Cleanup has no production caller. | Caller search/Core deadlines | Unwired behavior. | `must_fix` | Evidence/jobs/assembly | S-07 independent and blocked. |
| Cleanup needs durable claims. | Cross-resource race | Duplicate/false completion. | `must_fix` | Evidence runtime state | Exact §7.3 design. |
| Providers are legitimate seams. | Catalogs/assembly | Ownership inversion. | `intentional/no_action` | Evidence/consumers | Characterize, retain. |
| No backend grid-vendor import. | Import scan | Invented architecture. | `defer` | Frontend adapter | Reopen only with evidence. |
| Test helpers are test-owned. | Test support | Production leakage. | `intentional/no_action` | Test support | Preserve. |
| Generated artifacts are read-only. | Policy/current state | Drift. | `intentional/no_action` | Generators | Never hand-edit. |

## 6. Refactor Workstreams

Core authorization closure MUST precede S-00. S-00 MUST precede S-01 through
S-06. S-07 is independent of structural completion and MUST NOT be folded into
a structural slice.

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Source/authority bootstrap | root | None | WF-01, WF-02 | Recheck target, worktree, adoption, allowed scope | Tracker/owners/target | Status and owner reads | Record commit, adoption, scope, slice. |
| WF-01 | Inventory | chain | WF-00 | WF-03, WF-04 | Reconcile files/callers/dependencies | Target/importers | File scan/boundary check | Explain delta from 58. |
| WF-02 | Owner closure | parallel | WF-00 | WF-03, WF-05 | Confirm Core 04/Reporting adopted; freeze Core 01 | Owners/contracts | Exact review | Stop on contradiction. |
| WF-03 | S-00 gate | chain | WF-01, WF-02 | WF-05, WF-06 | Establish differential/security/provider baselines | Tests/authorized routing | Owner slices and discovered consumer targets | All S-00 suites pass. |
| WF-04 | Coupling scan | parallel | WF-01, WF-02 | WF-05 | Recheck boundaries | Target/assembly/manifest | Boundary check | Every finding classified. |
| WF-05 | Redesign freeze | chain | WF-02, WF-03, WF-04 | WF-06 | Freeze interfaces/transactions/order/rollback | Rows/effects/runtime/providers | Sections 3-5 review | No unresolved interface. |
| WF-06 | Structural execution | chain | WF-03, WF-05 | WF-07, WF-08 | Execute S-01 through S-06 separately | Slice files | Narrow gates | Update tracker each slice. |
| WF-07 | Accounting | chain | WF-03 and affected slice | WF-08 | Reconcile semantic evidence only | Authored harness inputs | Harness/drift | No runtime inference. |
| WF-07C | Cleanup correction | parallel | WF-00 and separate S-07 authorization | WF-08 | Implement S-07 independently | Evidence/jobs/assembly/migration/tests | Evidence plus discovered jobs target | Full cleanup matrix passes. |
| WF-08 | Validation/handoff | chain | WF-06/WF-07 or authorized WF-07C | None | Reconcile gates and risks | Changed scope/tracker | Section 8 | Record results/rollback/skips. |

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 | WF-00 through WF-04; adopted Core 04/Reporting amendments | Characterization only | Evidence/consumer tests; authorized routing only | Freezing accidents | Parity, atomicity, authorization, five provider suites | Evidence owner slices; discover consumer targets | Revert tests/routing | All evidence passes; zero row mismatch. |
| S-01 | S-00 | Replace duplicate mapper with canonical Rows | Store/private composition | Exact row/event/replay | S-00 parity/query/rollback | Evidence owner slices | Restore mapper | One oracle; exact behavior. |
| S-02 | S-00 | Add §3.1 port and Projections adapter | Evidence/Projections/assembly | Subjects/versions/views/ordering/atomicity | Atomicity/fault suite | Service-backed/boundary | Revert atomically | No peer layout; no shim/partial success. |
| S-03 | S-00, S-01, S-02 | Split Store into private capabilities | Store/persistence/access/runtime | Locks/idempotency/handles | Existing suites plus nil/fault cases | Owner slices/boundary | Reversible capability commits | Cohesive single owners. |
| S-04 | S-00, S-02 | Narrow Workbook/Import composition | Facades/assembly | Defaults/provenance/blob/conflict/effects | Import parity/create tests | Evidence/Imports targets/boundary | Revert facade+assembly | Contracted parity passes. |
| S-05 | S-00 | Consolidate lifecycle/reference admission | Validation/live/rollback/blobref | Transitions/refs/snapshots | Lifecycle/rollback matrices | Evidence owner slices | Restore validators | One identical interpretation. |
| S-06 | S-00 | Close/reorganize provider seams only where useful | Provider packages/consumer assembly | IDs/redaction/paths/inventory/replay | Five suites plus contracts | Evidence and discovered consumers | Revert provider-by-provider | Exact behaviors pass. |
| S-07 | Separate later authorization; WF-07C | Implement §7.3 private cleanup | Evidence/jobs/assembly/migration/tests | Deletion/timing/retry/shutdown | Complete §7.3 matrix | Evidence service-backed; `TODO: discover current platform-jobs owner command` | Disable/revert code+migration; retain evidence | Deadline/safety/restart/telemetry pass; no public effect. |

S-00 through S-06 are behavior-preserving. S-07 is
`requires later authorization`.

### 7.1 Projection-effects characterization

`module.evidence.association_projection_effects_atomicity` MUST freeze exact
support record IDs, versions, view-schema set, field keys, patch/invalidate,
event order, replay, failure at each provider, and no post-rollback event.
Sets MUST be exact, sorted, unique, and all-or-nothing.

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

### 7.3 S-07 private durable-claim cleanup

S-07 is private maintenance. It MUST create no public job resource, route,
Workbook row, Collaboration event, revision, or incident audit entry. Its
stable private identity is `evidence.failed_unattached_blob_cleanup.v1`.

| Parameter/concern | Fixed requirement |
| --- | --- |
| Ownership | Evidence owns candidates/state/recheck/delete request/completion; platform jobs owns periodic singleton dispatch/cancellation/shutdown; application assembly registers; object-store adapter deletes; telemetry boundary instruments. |
| First sweep | Immediately after serving readiness. |
| Interval/configuration | Fixed 15 minutes; no new deployment configuration. |
| Batch/order | 100, ordered `cleanup_due_at ASC, object_blob_id ASC`; drain further batches while eligible work remains. |
| Failed due time | `cleanup_due_at = failed_at + 45 minutes`. |
| Runtime relation | `evidence_blob_cleanup_claims`, excluded from Recovery authoritative state. |
| Claim fields | Blob identity, unique claim token, claimed/expiry timestamps, attempt count, next-attempt timestamp. |
| Claim lease | 5 minutes. |
| Deletion deadline | 1 minute. |
| Retry | 1, 5, then 15 minutes; cap so no knowingly scheduled retry exceeds the one-hour deadline. |
| Retention | Never hard-delete metadata; queryable at least seven days. |
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
| unit | `make test-slice OWNER=module.evidence` | Evidence fast rows | yes for S-00 through S-06 | Baseline first. |
| integration | `make service-backed-test-slice OWNER=module.evidence` | Evidence service-backed | yes | Per slice and S-07. |
| consumer integration | `TODO: discover current owner commands with make task-guide and make explain-test-owner` | Reporting/Imports/Bundles/Recovery/jobs | yes when applicable | Explicit non-normative command-discovery marker. |
| e2e/browser | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | Observable consumers | conditional | Only if affected. |
| generated drift | `make generate-drift`; `make generated-artifact-policy-check` | Generated policy | conditional | Never hand-edit. |
| import/static | `make backend-module-boundary-check` | Evidence/Projections | yes for structural slices | Especially S-02-S-04/S-06. |
| harness | `make harness-contract` | Accounting | conditional | Only if routing changes. |
| full | `make test-fast`; `make check` | Repository | final implementation | Run `make agent-finalize` before broad implementation gates. |
| docs | `git diff --check -- <four docs>`; `make lint-markdown` | This task | yes | Product tests/generation/finalize/broad checks skipped. |

This docs task MUST also prove no unresolved normative `TODO`, `TBD`,
`should`, `consider`, or `recommended` in added lines; 58 inventory rows and
all initial handoff rows remain; Core 04 alone owns the matrix; Core 01 owns
route/token behavior; Reporting owns its allowlist; the delta has only four
docs; and no staging occurred. The two `TODO` command cells are explicit
non-normative discovery markers required because live owner commands remain
unselected.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| TR-001 | Authority/scope | WF-00 | DONE | None | Section 1 | Planning authority closed; implementation unauthorized. |
| TR-002 | 58-file inventory | WF-01 | DONE | TR-001 | Section 2 | Every file accounted. |
| TR-003 | Boundary diagnosis | WF-02/WF-04 | DONE | TR-002 | Sections 3/5 | Legitimate/mixed/misplaced distinguished. |
| TR-004 | Core 04 matrix text | WF-02 | DONE | TR-001 | Core 04 §2.0A/AC-543 | Exact authority authored. |
| TR-005 | Core 01 bindings/xrefs | WF-02 | DONE | TR-004 | Core 01 | No duplicate matrix. |
| TR-006 | Reporting allowlist | WF-02 | DONE | TR-001 | Reporting §7.1.1 | Exact allowlist/default authored. |
| TR-007 | Adopt owner amendments | WF-00/WF-02 | BLOCKED | TR-004-TR-006 | Repository adoption process | Adopted/current baseline. |
| TR-008 | S-00 | WF-03 | TODO | TR-007 and later test authorization | Suite evidence | All pass; zero mismatch. |
| TR-009 | S-01 | WF-06 | TODO | TR-008 | Canonical reader | One oracle. |
| TR-010 | S-02 | WF-05/WF-06 | TODO | TR-008 | Fixed interface | Atomic, no shim. |
| TR-011 | S-03 | WF-06 | TODO | TR-009/TR-010 | Private capabilities | Exact behavior. |
| TR-012 | S-04 | WF-06 | TODO | TR-008/TR-010 | Import parity | Passes. |
| TR-013 | S-05 | WF-06 | TODO | TR-008 | Policy tests | One interpretation. |
| TR-014 | S-06 | WF-06 | TODO | TR-008 | Provider suites | Exact behaviors. |
| TR-015 | S-07 | WF-07C | BLOCKED | Separate later authorization | §7.3/RB-001 | Full matrix; no public effect. |
| TR-016 | Accounting | WF-07 | DEFERRED | Activated slices | Harness artifacts | Only real identity changes. |
| TR-017 | Final implementation handoff | WF-08 | TODO | Activated work | Section 8/log | Results/rollback/risks agree. |
| TR-018 | Frontend/grid move | WF-05 | DROPPED | TR-003 | Repository evidence | Reopen only with new authority/evidence. |

## 10. Session Handoff Log

The initial planning rows are preserved session history. Every later session MUST append and MUST NOT overwrite an earlier row except to remove a demonstrated duplicate. The initial session touched only this tracker. The NLSpec-grade revision session touched only the four documentation files authorized in Section 1.

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-11T06:38:40-04:00 | Codex / initial planning session | Planning inventory complete; implementation unauthorized | Inspected framework, Domain, Core 00-04, Reporting and Testing Harness NLSpecs; touched only this tracker | `sed` owner/framework reads; `rg --files docs`; `rg -n` authority searches; `git status --short --branch` | Target and authority posture confirmed; no Evidence NLSpec and no owner contradiction found | None for tracker completion | A future authorized agent starts at WF-00 and rechecks the commit/owners. |
| 2026-08-11T18:27:01-04:00 | Codex / NLSpec-grade revision | Design and owner authority closed in documentation; implementation unauthorized | Inspected analysis notes, NLSpec guide, tracker, Core 01, Core 04, Reporting NLSpec, and targeted live repository evidence; touched four authorized docs | Targeted `sed`/`rg`; `git status --short --branch`; documentation patching | Core 04 owns the matrix, Core 01 owns route/capability behavior, and Reporting owns the provider allowlist | Owner amendments require adoption into the implementation baseline and S-00 evidence | Adopt the amendments, then authorize S-00 separately. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-11T06:38:40-04:00 | Codex / initial planning session | Legitimate Evidence owner; mixed root package; future slices unstarted | Inspected all 58 target files, app assemblies, importers, migrations, and `tools/backend_module_boundaries.json`; touched only tracker | `find`/`rg --files`; targeted `sed`; `rg -n` callers/imports/symbols/cleanup uses | Store/row/projection-effect hotspots and intentional provider seams mapped | RB-001 through RB-003 block affected implementation slices | Execute S-00 characterization after later authorization. |
| 2026-08-11T18:27:01-04:00 | Codex / NLSpec-grade revision | S-01 through S-06 are fully specified future behavior-preserving slices | Inspected Evidence runtime, Store, Rows port, support-refresh SQL, Projections adapters, assembly, and providers; touched authorized docs only | Targeted `sed`/`rg` for live types, callers, SQL, and constructors | Canonical-row oracle and `SupportProjectionEffectsTx` interface fixed; no shim allowed | S-00 has not run | Run S-00; permit S-01 only after zero mismatches. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-11T06:38:40-04:00 | Codex / initial planning session | Frontend is an external consumer, not part of the target move | Inspected Evidence service, Workbook bindings, lifecycle/access presentation, Timeline attachment adapter, and browser Evidence fixtures/uploads; touched only tracker | Targeted `sed` and `rg -n` for Evidence fields/actions/selectors/vendor imports | No backend grid-vendor dependency and no supported frontend relocation found | None unless a later slice changes wire/event/schema behavior | Run relevant frontend/browser rows only when a backend seam affects them. |
| 2026-08-11T18:27:01-04:00 | Codex / NLSpec-grade revision | Frontend remains an external consumer; no frontend slice planned | Reused and reviewed initial frontend evidence; touched authorized docs only | Tracker and owner review | No Evidence ownership transfer or grid-vendor work authorized | None for structural planning | Run browser consumers only when a later backend slice can affect them. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-11T06:38:40-04:00 | Codex / initial planning session | Six HTTP operations and typed view/provider surfaces frozen; no contract edit planned | Inspected Evidence OpenAPI source, view schema, errors, projection/import/incident/recovery catalogs, and generated projections read-only; touched only tracker | `jq` OpenAPI path/operation extraction; `rg -n` contract/generated references | Structural plan requires no generated or authored contract change | A discovered wire/schema change would stop the structural slice | Use Make generators/drift only after separately justified authored input changes. |
| 2026-08-11T18:27:01-04:00 | Codex / NLSpec-grade revision | Authorization and Reporting authority gaps closed in owner text; generated artifacts untouched | Inspected Core 01 routes/handles, Core 04 authorization, Reporting boundary, and live provider; touched three owner docs plus tracker | Targeted reads and documentation patching | No route, role, token, error vocabulary, public resource, contract artifact, or generated file added | Adoption and executable characterization remain future gates | Adopt owners and run S-00; never hand-edit generated projections. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-11T06:38:40-04:00 | Codex / initial planning session | Current owner topology mapped: 45 rows, 34 service-backed; tests not executed in planning | Inspected all target tests, Evidence verification owner/family, test support, and browser fixtures; touched only tracker | `make help-all`; `make task-guide ROLE=module-author OWNER=module.evidence`; `make explain-test-owner OWNER=module.evidence`; targeted source reads | Canonical narrow commands discovered; characterization gaps named | Cleanup job owner command still requires live discovery | In S-00, select exact live row IDs and establish a passing baseline. |
| 2026-08-11T18:27:01-04:00 | Codex / NLSpec-grade revision | Five provider suites and two differential suites specified; no tests or harness inputs changed | Inspected analysis recommendations and owner command posture; touched authorized docs only | Targeted reads; no product test, generator, or harness command | Acceptance matrices cover success, failure, rollback, replay, restart, redaction, determinism, and set equality | Suites do not yet exist or have not been proven exact | Consumer owners add/map suites later; Evidence supplies fixtures/contributions. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-11T06:38:40-04:00 | Codex / initial planning session | Route admission, handle recheck, concealment, and opaque-target seams identified | Inspected route admission/errors/dependencies, access repository, upload tokens, blobref, Core 04, and security tests; touched only tracker | Targeted `sed` and `rg -n` for membership/role/handle/storage identifiers | Current checks are edge-local and intentionally guarded | RB-004 before route/access restructuring | Add a six-operation role/membership matrix if existing cross-owner evidence is not exact. |
| 2026-08-11T18:27:01-04:00 | Codex / NLSpec-grade revision | Exact six-operation role, recheck, concealment, precedence, and state-change rules authored in Core 04 | Inspected Core 01/04 and live Evidence admission, tokens, handles, errors, tests; touched authorized docs only | Targeted reads and documentation patching | Design resolved without a new public token, role, route, error, or resource | Core 04 adoption and table-driven characterization remain required | Pass `module.evidence.route_authorization_concealment_matrix` before route restructuring. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-11T06:38:40-04:00 | Codex / initial planning session | Tracker ready for implementation handoff; all implementation rows remain unstarted | Inspected tracker evidence set; touched only tracker | `date -Iseconds`; `git status --short --branch`; `git diff --check`; `make lint-markdown` | Worktree was clean before tracker creation; inventory counts matched 58/58; diff and Markdown checks passed; transient ignored lint artifacts were removed; no production change made | RB-001 is a behavior-authorization blocker; RB-002 through RB-005 are evidence gates | Recheck repository delta, authorize S-00 only, and update this log before any test/code edit. |
| 2026-08-11T18:27:01-04:00 | Codex / NLSpec-grade revision | Gap closure means design/authority closure only; S-00 through S-07 remain unimplemented | Inspected all named material; touched only four authorized docs | `git diff HEAD --check`; `make lint-markdown`; inventory, history, vocabulary, authority, and scope scans | Documentation lint and all tracker invariants passed; closure ledger distinguishes READY gates from the separately blocked correction | S-07 needs later authorization; S-00 absent; owner amendments need adoption | Authorize only S-00 or separately authorize S-07. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Cleanup is designed but implementation is unauthorized. | It activates deletion, migration, scheduling, and retries. | Separate S-07 authorization and §7.3 evidence. | `BLOCKED`: design resolved; implementation pending S-07 authorization; does not block S-00-S-06. |
| RB-002 | Canonical row parity lacks executable proof. | S-01 cannot remove duplicate interpretation safely. | Attach-row parity suite. | `READY` for S-00 differential characterization. |
| RB-003 | Support-projection seam was undefined. | S-02 must preserve atomic neutral effects. | §3.1 interface and atomicity suite. | `READY` for S-00/S-02 with interface fixed. |
| RB-004 | Core 04 text is authored but not adopted/characterized in an implementation baseline. | Route changes can alter disclosure/precedence. | Adopt §2.1; pass authorization matrix. | Design resolved; `BLOCKED` until amendment adopted and characterized. |
| RB-005 | Consumer behavior was not proven by catalogs alone. | Rearrangement can leak/lose/diverge state. | Five §7.2 suites; Reporting amendment adoption. | Suites defined; Reporting work also depends on Reporting NLSpec amendment. |

No owner contradiction is currently established. Any later contradiction MUST
be recorded as `BLOCKED: owner contradiction` and the affected work MUST stop.

## 12. Binary Completion Criteria

- **PASS:** All 58 target files remain inventoried: 42 production Go, 15 test
  Go, and one `.gitkeep`.
- **PASS:** Every public route, handle, row/query, saved-view, lifecycle,
  projection, authorization, revision, Collaboration, Reporting, Imports,
  portability, Recovery, generated, frontend, and harness risk has owner/test
  posture.
- **PASS:** Core 04 §2.0A solely owns the authorization matrix.
- **PASS:** Core 01 owns routes, bindings, lifetimes, states, envelopes, errors.
- **PASS:** Reporting owns the exact twelve-member allowlist/default omission.
- **PASS:** Every workflow has dependencies, validation, and checkpoint.
- **PASS:** S-00 precedes behavior-preserving S-01-S-06.
- **PASS:** S-07 is independent and requires later authorization.
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
- **PASS:** No implementation or production completion is claimed.
