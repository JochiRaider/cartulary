# recovery Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

- **Target path:** `internal/modules/recovery`
- **Target label:** `recovery`
- **Output path:** `docs/handoffs/recovery-module-refactor-tracker.md`
- **Status:** Planning and documentation only.
- **Allowed change in this session:** This tracker file only.
- **Non-goals:** No production refactor, test change, contract change, generated
  artifact change, package configuration change, migration, harness change, or
  behavior change is authorized.
- **Implementation posture:** Every implementation action described below requires
  a later, explicitly authorized task. Observable behavior is frozen by default.

### 1.1 Normative language and authority boundary

Within this tracker, **MUST**, **MUST NOT**, and **MAY** are closed normative
terms for the proposed refactor:

- **MUST** and **MUST NOT** identify a mandatory condition that a later
  authorized implementation is required to satisfy.
- **MAY** identifies optional behavior only when the same paragraph or table row
  defines the behavior when the option is omitted.
- `Default:` identifies behavior that applies when an authorized design leaves no
  narrower choice.
- `Current baseline:` identifies live implementation evidence that MUST be
  preserved during a behavior-preserving move.
- `READY FOR ADOPTION` means the design decision is complete but is not an
  adopted behavioral requirement.
- `READY FOR RS-00` means the design decision is complete and the next permitted
  action is the characterization-only slice.

This tracker is implementation-support evidence. It is not an adopted NLSpec,
Core owner section, machine contract, conformance declaration, or authorization
record. Proposed requirements in this tracker become implementation authority
only after they are promoted to the owner locations in section 5.1. If this
tracker conflicts with an adopted owner, the adopted owner governs and the
tracker MUST record `BLOCKED: owner contradiction`.

Every proposed interface and algorithm below defines observable inputs, outputs,
ordering, omission behavior, and failure consequences. Internal data structures,
source-file decomposition, concurrency mechanism, and SDK choice remain
intentional implementation freedom only when they cannot change observable
bytes, identifiers, ordering, errors, persistence, security, recovery, or
interoperability.

The source hierarchy used for this tracker is:

1. Adopted subsystem NLSpecs within their named scopes.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. Domain vocabulary and implementation-support guides.
5. Current code and tests.
6. Prior plans, handoffs, and the planning framework as evidence only.

The planning framework was read first and used as planning doctrine, not as proof
of current repository state. No owner contradiction was found. If later evidence
reveals one, the affected item must be recorded as `BLOCKED: owner contradiction`
without selecting a side.

Owner documents inspected:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`, in full.
- `docs/spec/00_document_set_status_and_precedence.md`, for authority and
  projection posture.
- Relevant recovery, storage, restore, projection, and operator sections of
  `docs/spec/01_architecture_storage_and_view_contracts.md`.
- Relevant recovery metadata, audit, and restored-record sections of
  `docs/spec/02_domain_model_schema_and_history.md`.
- Relevant restored workbook/query behavior in
  `docs/spec/03_application_behavior_and_user_surfaces.md`.
- Relevant backup, restore, operator, encryption, and route-absence sections of
  `docs/spec/04_security_deployment_and_conformance.md`.
- Core 05 claim-publication applicability in
  `docs/spec/05_measurement_and_claim_publication.md`; no claim-bearing
  publication is in this target's current surface.
- Recovery vocabulary and owner navigation in `docs/domain.md`.
- Backup binding and codec requirements in
  `docs/extension-subsystem-nlspec.md`, within the Extension subsystem only.

Informative planning inputs inspected:

- `temp/analysis-notes.md`, as the source of the recommended design closures and
  document-placement analysis.
- `docs/research/nlspec-spec.md`, as writing guidance for behavioral
  completeness, interface precision, defaults, mapping tables, and binary
  acceptance criteria.

Neither informative input supersedes an adopted owner. External research and
environment-specific links present in the analysis notes are not copied into
this tracker and do not support a normative requirement here.

Repository evidence inspected:

- Every file listed in the inventory below.
- Inbound composition and caller files under `internal/app/extensionassembly`,
  `internal/app/operator`, `internal/app/recoveryassembly`, and
  `internal/app/timelineassembly`.
- `internal/modules/projections/rebuild.go` and its tests,
  `internal/platform/administrativeaudit/audit_integration_test.go`,
  `tools/recoverybrowserrestore/main.go`, and the browser restore fixture.
- `db/queries/recovery.sql`,
  `db/migrations/00003_deployment_admin_and_recovery.sql`, and generated
  `internal/gen/sql/recovery.sql.go`.
- Extension backup registry and codec inputs under `contracts/extensions`, and
  their generated Go projection under `internal/gen/contractextensions`.
- `contracts/verification/owners/module.recovery.json`,
  `tools/test_families/module.recovery.json`, `tools/test_catalog_owner.json`,
  `tools/execution_topology_manifest.json`,
  `tools/backend_module_boundaries.json`, and
  `tools/schema_object_ownership_manifest.json`.

The target path exists. The normalized label is exactly `recovery`. The framework
and live repository differ in an important way: the framework describes a
cohesive Recovery orchestration owner, while the current package also contains
CLI transport, persistence, concrete workbook querying, platform configuration,
and cross-owner object-store migration mechanics. That mismatch is an
architectural planning finding, not an implementation decision.

## 2. Current-State Repository Inventory

All 29 files under `internal/modules/recovery` are in scope and inventoried.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `capture.go` | Captures Postgres and object-store snapshots, writes backup artifacts and integrity manifests, and classifies authoritative tables. | `CaptureService`, `BackupStorage`, artifact and manifest DTOs, capture/decode helpers, `IsAuthoritativePostgresSnapshotTable` | Operator operations and Recovery tests | Platform Postgres and object store; Recovery store and extension catalog | Backup metadata, object-store artifact, restore, and Extension contract tests | Recovery SQL projection; extension backup registry | Recovery orchestration, with source-owner snapshot ports | high | The hand-maintained authoritative-table predicate spans module-owned data. |
| `catalog.go` | Selects the latest durable retained backup and verification-due candidates after proof checks. | `BackupCatalog`, selection and diagnostic DTOs, constructor and selection methods | Restore runner, verification runner, operator operations, tests | Recovery store, backup storage, extension catalog | Backup metadata and restore tests | Recovery SQL projection and integrity artifacts | Recovery | high | Selection order, 24-hour freshness, and proof checks are observable. |
| `encryption.go` | Loads the recovery key and provides AES-GCM envelopes for backup storage and operator journal payloads. | `RecoveryEncryptionKey`, encryption proof and journal envelope DTOs, key parsing/loading, encrypted storage constructor | Recovery assembly, operator journal, tests | Cryptographic standard library; `BackupStorage` | Backup metadata and journal tests | Envelope schema identifiers persisted by recovery migration | Recovery security primitive | high | Key handling and fail-closed behavior must not change. |
| `extensions.go` | Builds immutable extension backup bindings, codecs, physical-table claims, and binding proofs. | `ExtensionBackupCodec`, `ExtensionBackupBinding`, `ExtensionBackupCatalog`, `ExtensionPristineMetadata`, `ExtensionBindingProof` | Extension assembly, capture, restore, tests | Extension contract projection supplied by application assembly | Extension catalog, contract, and internal tests | `cartulary.extension_backup_registry.v1` and extension codec inputs | Recovery orchestration using Extension-owned contributions | high | Exact current or historical codec and numeric order are frozen. |
| `object_store_backup_artifacts.go` | Builds SeaweedFS backup manifests, redacted summaries, and restore-verification artifacts; reads blob IDs by storage reference. | Object-store backup and verification artifact DTOs; capture, encode, decode, validate, and lookup helpers | Operator operations, verification flow, tests | Platform object store and Postgres; raw `object_blobs` lookup | Object-store artifact tests and restore verification paths | JSON artifact schema identifiers implemented in Go | Recovery artifacts, with Evidence lookup port | high | Canonical JSON, redaction, digest, and storage-reference behavior are contracts. |
| `object_store_migration.go` | Models and executes MinIO-to-SeaweedFS migration state, copy ledgers, validation, target probes, and rollback evidence. | Migration state, event, ledger, validation, diagnostic, probe, rollback, and artifact DTOs plus encode/decode/copy/validate helpers | Operator operations, dedicated migration tests | Platform Postgres and object store; Evidence `blobref`; raw Evidence/object blob SQL | Migration unit and preservation integration tests | Six existing migration artifact schemas implemented in Go | Proposed `module.objectstoremigration`, with Evidence and platform ports | critical | RB-001 is design-complete and ready for owner adoption; production relocation remains blocked. |
| `restore.go` | Admits a stopped-empty target, verifies backup artifacts, restores Postgres and object storage, validates extensions, rebuilds projections, checks consistency, and gates readiness. | `RestoreRunner`, `RestoreTarget`, gates/observer, result and consistency DTOs, reset and decode helpers | Operator operations, browser restore helper, tests | Platform Postgres/object store; Recovery store/catalog; `restorecontract` | Restore owner, projection, Extension contract, object-byte, and browser tests | Recovery SQL, backup artifacts, projection contract | Recovery orchestration | critical | Order and fail-closed readiness semantics are public operational behavior. |
| `store.go` | Wraps generated sqlc queries for backup sets and verification runs and enforces retention, freshness, and selection rules. | `Store`, `BackupSet`, `RestoreVerificationRun`, verification state and typed selection errors | Capture, catalog, restore, verification, operator operations, tests | Platform Postgres and `internal/gen/sql` | Metadata unit and integration tests | `db/queries/recovery.sql` to generated sqlc | Recovery persistence adapter | high | Exported concrete sqlc-facing store should be hidden behind semantic services. |
| `verification.go` | Runs latest, specific, and due restore verification and records pass/fail state and verification artifacts. | Verification runner, request/result/probe surfaces and verification-basis helpers | Operator operations and tests | Recovery catalog, restore runner, store, backup storage, workbook probe | Metadata, restore owner, workbook probe, and browser tests | Recovery SQL and restore-verification artifact | Recovery | critical | Seven-day or basis-change due behavior and atomic completion are frozen. |
| `workbook_probe.go` | Selects the lowest incident UUID by text order and executes the Timeline default query as the restore workbook probe. | `WorkbookProjectionQuery`, `RestoreVerificationWorkbookProbe`, constructor and probe method | Timeline assembly and verification flow | Platform Postgres and view-schema query interface; hard-coded Timeline view ID | Workbook probe and browser restore tests | Timeline view-schema contract consumed at runtime | Timeline registration plus Workbook executor, composed for Recovery | high | RB-002 fixes exact ownership and parity behavior; the current request is recorded in Appendix A. |
| `operatorcli/cli.go` | Parses the five operator commands, validates flags/timeouts, emits JSON/JSONL, maps results and errors, and runs operations. | Command/result/error/progress/outcome DTOs, `Operations`, `Runner`, parse/validation/mapping helpers | Operator application root and process tests | Recovery root types; standard flag, JSON, and I/O packages | Operator process tests | Operator command wire contract implemented in Go | Operator CLI adapter plus Recovery application ports | critical | Operations are coupled to CLI DTOs; error mapping relies on error text. |
| `operatorcli/journal.go` | Encrypts and persists operator recovery journal records and emits safe administrative-audit summaries. | `JournalStore`, `JournalRecord`, constructor and record method | Operator operations and journal tests | Recovery encryption; platform administrative audit and Postgres | Journal test and administrative-audit integration test | Recovery journal SQL/schema and audit vocabulary | Recovery journal port with platform adapters | high | Audit secrecy and encryption must remain exact. |
| `operatorops/operations.go` | Coordinates backup, inspect, restore, verification, migration, deployment loading, locks, journal, audit, and progress. | Deployment and binding DTOs, loader/factory types, `Service`, operation methods | Operator application composition and tests | Recovery root, `operatorcli`, `restorecontract`, platform Postgres/object-store settings and factories | Operator application and process tests | Operator CLI, recovery artifacts, projection contract | Recovery application service with operator adapter | critical | Semantic operations directly depend on transport DTOs and platform settings. |
| `restorecontract/projections.go` | Defines the Recovery-owned request/result boundary for projection rebuilds. | Projection scope/status/readiness/provider/table/message DTOs and `ProjectionRebuilder` | Recovery restore, operator/timeline assembly, Projections implementation and tests | Standard context/time only | Projection contract and Projections rebuild tests | Recovery-to-Projections contract | Recovery contract; Projections implements it | high | This owner direction is explicit in Core 01 and should remain. |
| `backup_metadata_integration_test.go` | Verifies real Postgres-backed backup metadata persistence and latest selection. | Go test package surface only | Go test and owner test family | Testcontainers/test support, Recovery store | Self | Recovery SQL and migration exercised | Recovery tests | medium | Service-backed owner row exists. |
| `backup_metadata_store_test.go` | Characterizes metadata shape, retention, due selection, durability, encryption, and latest-backup rules. | Go test package surface only | Go test and owner test family | Recovery services and fakes | Self | Recovery artifact and SQL semantics | Recovery tests | high | Atomic verification completion is not explicitly selected by the current owner row. |
| `extension_catalog_test.go` | Supplies generated extension catalog fixtures and helpers for tests. | Test helper surface only; no `Test*` entry point | Extension Recovery tests | Generated extension contract projection | Extension contract tests | Generated extension backup registry | Recovery test support using Extension contracts | medium | Helper-only file; not production test-util. |
| `extension_recovery_contract_test.go` | Characterizes binding proofs, stopped-empty admission, and rejection before mutation. | Three integration tests | Go test and owner family | Recovery capture/restore and extension fixtures | Self | Extension registry and recovery artifacts | Recovery and Extension contract tests | critical | Exact pre-mutation failure order is frozen. |
| `extensions_internal_test.go` | Characterizes exact codec selection, canonical ordering, and shared-table rejection. | Two unit tests | Go test and owner family | Recovery extension catalog | Self | Extension codec and binding projection | Recovery and Extension contract tests | high | Current owner row covers these tests by prefix. |
| `object_store_backup_artifacts_test.go` | Characterizes canonical manifests, redacted summaries, duplicate rejection, and Seaweed capture. | Three support tests | Go test; broader support targets | Recovery artifact helpers and object-store fake | Self | Object-store artifact schemas | Recovery tests | high | Current `module.recovery` family does not explicitly route these tests. |
| `object_store_migration_preservation_integration_test.go` | Verifies pass and mismatch evidence while preserving object bytes and keys. | Two service-backed migration tests | `make seaweedfs-migration-preservation` | Recovery migration helpers and service-backed object stores | Self | Migration evidence artifacts | Deferred cross-owner migration tests | critical | Separate target exists; do not infer permanent ownership from it. |
| `object_store_migration_test.go` | Characterizes migration state, quiescence, validation, copy ledger, and blob-reference preflight. | Six support tests | Go test; broader support targets | Recovery migration helpers and fakes | Self | Migration artifact schemas | Deferred cross-owner migration tests | critical | Not explicitly routed under current `module.recovery` owner rows. |
| `operatorcli/journal_test.go` | Verifies encrypted journal payloads and safe audit summaries. | One unit test | Go test | Operator journal, Recovery encryption, DB/audit fakes | Self | Journal envelope and audit contract | Recovery security tests | high | Not explicitly selected by current Recovery owner rows. |
| `operatortest/operator_process_test.go` | Exercises object-store initialization and all five standalone operator commands as process contracts. | Six process tests | Owner process family and Go test | Built operator process, deployment services, JSON wire | Self | Operator CLI contract and recovery artifacts | Recovery operator process tests | critical | Five Recovery commands are owner-routed; object-store init is separately scoped. |
| `recovery_sentinel_test.go` | Statically rejects authored public backup, restore, and verification HTTP/WS route literals. | One unit test | Owner unit family | Authored Go source inventory | Self | Public route inventory | Recovery contract test | critical | Route absence is intentional and normative. |
| `restore_owner_integration_test.go` | Characterizes latest selection, restore order, missing artifacts, and fail-closed readiness. | Three unit/integration tests | Recovery owner integration rows | Recovery restore with fakes | Self | Backup and restore artifacts | Recovery tests | critical | Existing owner rows select these behaviors. |
| `restore_projection_contract_test.go` | Characterizes structured projection requests and fail-closed projection readiness. | Two unit tests | Go test and broader Recovery tests | Recovery restore and fake `ProjectionRebuilder` | Self | Recovery-to-Projections contract | Recovery and Projections contract tests | critical | Not explicitly routed under current Recovery owner rows. |
| `restore_test.go` | Characterizes latest tie-break selection and restorable object bytes. | Two unit tests | Go test and broader Recovery tests | Recovery store, catalog, artifacts, fakes | Self | Backup artifact formats | Recovery tests | high | Not explicitly routed under current Recovery owner rows. |
| `workbook_probe_test.go` | Characterizes deterministic incident selection and built-in workbook query verification. | One unit test | Go test and browser behavior | Recovery workbook probe, DB/query fakes | Self | Timeline view-schema usage | Workbook/Timeline adapter test for Recovery | high | Not explicitly routed under current Recovery owner rows. |

## 3. Module Boundary Diagnosis

The target is a legitimate Recovery application and service owner, but not a
legitimate permanent boundary for every current responsibility. It is a
mixed-responsibility package containing a recovery orchestration layer,
persistence-adjacent adapter, transport-adjacent adapter, mutation coordinator,
view/projection verification adapter, platform integration, and operational
object-store migration. It is not a frontend shell, saved-view owner,
grid-vendor integration layer, or public HTTP/WebSocket transport.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Backup capture, durable catalog, retention, and proof verification | Root Recovery package | Recovery | keep | Core 01 recovery requirements; capture, catalog, store, and tests | Hide persistence mechanics without changing behavior. |
| Recovery encryption and fail-closed key use | `encryption.go` | Recovery security primitive | keep | Core 04 and encryption tests | Platform key loading may remain an adapter. |
| Restore admission, orchestration, consistency, and readiness | `restore.go` | Recovery | keep | Core 01/Core 04 and restore tests | Preserve stopped-empty admission before artifact reads or mutation. |
| Extension backup binding and codec proof validation | `extensions.go` | Recovery orchestration consuming Extension-owned contributions | keep | Extension NLSpec and generated extension registry | Recovery must not redefine Extension facts. |
| Projection rebuild request/result contract | `restorecontract` | Recovery contract | keep | Core 01 explicitly assigns the adapter contract to Recovery | Projections continues to own implementation. |
| Concrete Timeline workbook verification query | `workbook_probe.go`; wired by Timeline assembly | Timeline registration and Workbook executor composed by Recovery assembly | move | Hard-coded `cartulary.view.timeline.v2`, exact default query, and Timeline assembly caller | Recovery MUST retain only deterministic incident selection, orchestration, skip rules, and failure mapping. |
| CLI parsing, JSON/JSONL encoding, flags, and exit mapping | `operatorcli/cli.go` | Operator transport adapter | split | Exact operator process tests and Core 01/Core 04 command contract | Preserve the exact five-command wire surface. |
| Backup/restore/verify application operations | `operatorops/operations.go` | Recovery application service | split | Operations call Recovery semantics but consume CLI DTOs and platform settings | Introduce closed Recovery command/result/progress/error contracts later. |
| Encrypted journal persistence and administrative audit emission | `operatorcli/journal.go` | Recovery journal port plus Postgres/audit adapters | split | Direct platform imports and security tests | Preserve audit redaction and envelope semantics. |
| Generated-sqlc persistence | `store.go` | Recovery persistence adapter behind private semantic ports | split | Direct `internal/gen/sql` use | Do not move the authored SQL owner without authority. |
| Authoritative table capture and raw restore | `capture.go`, `restore.go` | Source-owner `cartulary.recovery_state_contribution.v1` registry aggregated by Recovery assembly | split | Hand-maintained table predicate, 92-table live baseline, and raw cross-module SQL | The runtime registry MUST be owner-authored; tooling manifests MUST remain validation-only. |
| Evidence blob and Revisions consistency checks | Artifact and restore code | Evidence/Revisions-owned narrow capabilities | move | Raw storage-reference and table access | Recovery retains orchestration of the checks. |
| MinIO-to-SeaweedFS migration | `object_store_migration.go` | Proposed `module.objectstoremigration` application boundary | move after adoption | Cross-owner SQL, blob key policy, object copy, probes, and separate harness target | Evidence owns reference validation; platform object storage owns mechanics; deployment configuration owns activation. |
| Public browser/API authorization | No production Recovery route | No Recovery public transport owner | keep | Core 04 and route-absence sentinel | Intentional absence; no `deployment_admin` capability is to be invented. |
| Frontend shell, saved views, grid adapter, and collaboration state | Not present in target | Existing frontend and domain owners | defer | No production inbound or outbound dependency | Only restored data and browser verification are indirect risks. |

### 3.1 Object-store migration boundary

**Proposed owner:** `module.objectstoremigration`.

**Default implementation placement:**

- semantic application service:
  `internal/modules/objectstoremigration`;
- application composition:
  `internal/app/objectstoremigrationassembly`;
- Evidence capability implementation:
  an Evidence-owned package selected by the Evidence owner;
- transport mechanics:
  `internal/platform/objectstore`;
- binding activation:
  a deployment-configuration adapter selected by the configuration owner.

The boundary is an in-process application service in the existing deployable. It
MUST NOT create a network listener, a separately deployed service, a sixth
standardized Recovery command, or an HTTP/WebSocket route.

| Responsibility | Required owner | Required interface consequence |
| --- | --- | --- |
| Admission, immutable plan, state machine, ledger, resume, aggregate validation, cutover readiness, rollback coordination, and private artifacts | `module.objectstoremigration` | Owns orchestration and state; MUST NOT parse Evidence identity or perform backend SDK operations directly. |
| Blob eligibility, logical `storage_ref`, physical-reference legality, blob/incident/key agreement, lifecycle eligibility, duplicate detection, and no-reference-rewrite proof | `module.evidence` | Supplies `BlobMigrationCatalog`; MUST complete validation before the first target mutation. |
| Source/target probe, stat, transfer, byte-stream SHA-256, conditional create, retryable transport classification, and deletion of migration-owned staged objects | `platform.objectstore` | Supplies `ObjectMigrationBackend`; MUST NOT update migration state or activate deployment binding. |
| Effective object-store binding activation | deployment configuration/cutover adapter | Supplies `ObjectStoreBindingActivator`; MUST require a valid completion proof. |
| Backup capture/catalog, coherent restore, restore verification, and five standardized commands | `module.recovery` | Recovery MUST NOT own migration mechanics after the adopted move. |

The proposed typed boundary is:

```go
type BlobMigrationCatalog interface {
    BuildMigrationPlan(
        context.Context,
        BuildMigrationPlanRequest,
    ) (BlobMigrationPlan, error)
    ValidatePhysicalReferences(
        context.Context,
        ValidatePhysicalReferencesRequest,
    ) (PhysicalReferenceValidationReport, error)
}

type BlobMigrationPlan struct {
    PlanID           string
    ConsistencyPoint string
    PlanDigestSHA256 string
    Entries          []BlobMigrationEntry
}

type BlobMigrationEntry struct {
    EntryID           string
    BlobID            string
    IncidentID        string
    LogicalStorageRef string
    PhysicalAddress   PrivateObjectAddress
    ExpectedSizeBytes int64
    ExpectedSHA256    *string
    LifecycleClass    string
}

type ObjectMigrationBackend interface {
    ProbeSource(context.Context, BindingRef) (SourceProbe, error)
    ProbeTarget(context.Context, BindingRef) (TargetProbe, error)
    Stat(context.Context, BindingRef, PrivateObjectAddress) (ObjectStat, error)
    Transfer(context.Context, ObjectTransferRequest) (ObjectTransferProof, error)
    Verify(context.Context, ObjectVerificationRequest) (ObjectVerificationProof, error)
    DeleteMigrationOwnedObject(
        context.Context,
        DeleteMigrationObjectRequest,
    ) error
}

type MigrationLedgerStore interface {
    Load(
        context.Context,
        string,
    ) (ObjectStoreMigrationRun, ObjectStoreMigrationCopyLedger, bool, error)
    Append(
        context.Context,
        ObjectStoreMigrationState,
        ObjectStoreMigrationRun,
    ) error
}

type MigrationArtifactStore interface {
    Put(
        context.Context,
        string,
        []byte,
        string,
    ) (ObjectStoreMigrationArtifactRef, error)
    Get(context.Context, ObjectStoreMigrationArtifactRef) ([]byte, error)
}

type ObjectStoreBindingActivator interface {
    Activate(
        context.Context,
        ObjectStoreMigrationCompletionProof,
    ) error
}
```

`PrivateObjectAddress` MAY contain a bucket and bucket-relative key internally.
It MUST NOT be operator input, public output, an ordinary log field, or Evidence
identity. `ExpectedSHA256=nil` means no pre-existing Evidence-owned digest is
available; it does not waive transfer verification. The migration service MUST
still compute source and target SHA-256 values over the exact byte streams.

Migration behavior MUST satisfy these rules:

1. Source and target bindings MUST differ. The target MUST be empty or contain
   only objects proved to have been staged by the same migration ID and immutable
   plan digest.
2. Evidence validation and write-quiescence proof MUST pass before target
   mutation. Missing credentials, malformed keys, duplicate addresses, identity
   mismatch, missing references, or an unprovable target state MUST fail closed.
3. A migration ID reused with the same plan digest resumes. Reuse with a
   different digest MUST fail before transfer. A completed ledger item MAY be
   reused only after its target proof still validates.
4. The default MinIO-to-SeaweedFS mapping preserves the bucket name and
   bucket-relative key exactly. `evidence.storage_ref` and user-authored external
   locators MUST NOT be rewritten, normalized, or reinterpreted. Zero-byte
   objects are valid.
5. Source and target lengths and independently computed SHA-256 values MUST
   match. ETag MAY be retained as diagnostic metadata but MUST NOT prove byte
   identity. A conflicting target object MUST fail closed and MUST NOT be
   overwritten as implicit repair.
6. Binding activation MUST require a completion proof bound to the migration ID,
   plan digest, source and target bindings, completed ledger, and passing
   validation. Default: if current cutover is manual or partially automated, the
   refactor preserves that behavior and does not add automation.
7. Before cutover, rollback MAY delete only target objects recorded as created by
   that migration. It MUST NOT delete source, pre-existing, or unledgered
   objects. After cutover, automatic destructive rollback is forbidden.
8. Plans, ledgers, validation extracts, buckets, keys, storage references, and
   blob identities MUST remain operator-private and encrypted at rest.
   Shareable summaries MUST contain only counts, safe IDs, status, durations,
   and closed error classes.

The ownership move MUST preserve the six schema IDs and the complete state,
event, copy-status, validation-status, and JSON member inventories in Appendix
C. A token or artifact redesign requires a separate contract-change
authorization.

### 3.2 Restore workbook probe boundary

The proposed owner allocation is:

| Responsibility | Required owner |
| --- | --- |
| Base registration and exact Timeline query semantics | `module.timeline` |
| Registry validation and execution through ordinary workbook query behavior | `module.workbook` |
| Deterministic restored-incident selection, invocation, skip rules, and failure mapping | `module.recovery` |
| Registry construction and injection | `internal/app/recoveryassembly` |

Recovery MUST depend on this interface and MUST NOT import a Timeline or
view-schema implementation package:

```go
type RestoreWorkbookProbe interface {
    Probe(
        context.Context,
        RestoreWorkbookProbeRequest,
    ) (RestoreWorkbookProbeResult, error)
}
```

`RestoreWorkbookProbeRequest` carries the restored deployment reference and the
Recovery-selected incident ID. It MUST NOT carry a Timeline schema ID, query
fields, visible labels, package names, or fixture names. The implementation
mechanism of the request/result records is otherwise intentionally unspecified
until the registration contract is adopted.

The proposed closed registration is:

```json
{
  "schema_id": "cartulary.restore_workbook_probe_registration.v1",
  "registration_id": "cartulary.restore_probe.timeline.v1",
  "contract_major": 1,
  "profile_id": "base",
  "owner_module": "timeline",
  "view_schema_id": "cartulary.view.timeline.v2",
  "incident_selector_id": "timeline.restore_probe.incident_id_text_asc.v1",
  "query_request": {
    "filters": [],
    "sort": [
      {
        "field_key": "timeline.activity_sort_ts",
        "direction": "asc"
      },
      {
        "field_key": "record_id",
        "direction": "asc"
      }
    ]
  },
  "eligible_source_row_predicate_id": null,
  "row_requirement": "zero_rows_allowed",
  "status": "active"
}
```

All members are required except `query_request.group_by`, which MUST be omitted
for the current registration. Explicit `null` for `group_by` is invalid.
`eligible_source_row_predicate_id=null` is the only nullable member and means no
source-row predicate requires a returned row. Exactly one active Base
registration for contract major `1` MUST exist.

Duplicate registration IDs, multiple Base defaults, unknown majors, inactive
view schemas, unresolved owners, malformed query requests, and unknown tokens
MUST fail application startup or restore-verification admission before query
execution. Recovery MUST NOT select a registration by package, path, fixture,
display label, or registry position.

Current baseline and failure behavior are defined once in Appendix A. A
transport, schema lookup, query execution, or owner-declared required-row
failure MUST produce `verification_failed` with
`reason_code='workbook_probe_failed'`. A zero-row result is success for the
current registration.

### 3.3 Source-owner recovery-state boundary

Source owners MUST provide immutable declarative contributions. They MUST NOT
receive arbitrary Recovery database handles or executable callback freedom. The
proposed v1 shape is:

```go
type RecoveryStateContributionV1 struct {
    SchemaID       string
    ContributionID string
    OwnerID         string
    ContractMajor   uint16
    Dependencies    []string
    Units           []RecoveryStateUnitV1
}

type RecoveryStateUnitV1 struct {
    UnitID                 string
    LogicalFamilyID        string
    StorageKind            StorageKind
    StateClass             StateClass
    BackupInclusion        BackupInclusion
    PhysicalRef            OpaquePhysicalRef
    CaptureCodecID         string
    CaptureCodecSHA256     string
    RestoreCodecID         string
    RestoreCodecSHA256     string
    RestoreOrderGroup      uint16
    PostRestoreValidatorID string
    RebuildAlgorithmID     *string
}
```

The schema ID MUST equal `cartulary.recovery_state_contribution.v1` and
`contract_major` MUST equal `1`. Closed tokens are:

| Field | Closed values | Omission/default behavior |
| --- | --- | --- |
| `storage_kind` | `postgres`, `object_store` | Required; no default. |
| `state_class` | `authoritative`, `derived` | Required; no default. |
| `backup_inclusion` | `required`, `excluded_rebuildable` | Required; no default. |
| `dependencies[]` | Contribution IDs | An owner-authored source MAY omit it; materialization MUST produce `[]`. The canonical contribution requires the array. Explicit `null` is invalid. |
| `units[]` | `RecoveryStateUnitV1` | Required and non-empty; explicit `null` is invalid. |
| `rebuild_algorithm_id` | Versioned algorithm ID or `null` | Authoritative state requires `backup_inclusion='required'` and `null`. Derived `excluded_rebuildable` state requires non-null. Derived required state MAY use null; omission materializes null. |

Set-like arrays MUST reject duplicates and use ascending UTF-8 byte ordering.
Units MUST use ascending `unit_id` UTF-8 byte order. Contributions MUST use
ascending `contribution_id` UTF-8 byte order. Registry serialization and digest
calculation MUST use the repository's adopted canonical JSON algorithm selected
by the future machine-contract owner; no algorithm may be inferred from this
tracker.

The registry MUST enforce:

- exactly one active owner and one required binding for every authoritative
  logical family;
- no physical reference claimed by two owners;
- resolved, unique, acyclic dependencies;
- known owners, majors, storage kinds, codecs, validators, and algorithms;
- derived exclusion only with a registered rebuild algorithm;
- fixed restore phases: Postgres authoritative state, object-store authoritative
  bytes, projection rebuild, then invariant validation/readiness;
- within-phase order by `restore_order_group`, then `contribution_id`, then
  `unit_id`, all ascending;
- one frozen registry and digest for an entire backup or restore operation;
- registry digest participation in the restore-verification basis;
- historical restore codec availability while any retained backup needs it;
- adaptation of Extension-owned state from the existing Extension registry,
  without a duplicate source of truth.

Unknown or duplicate facts, missing owner contributions, unresolved
dependencies, cycles, shared physical claims, unknown codecs, or unknown
validators MUST fail before backup publication or restore mutation.

Owner-authored versioned contribution inputs are the proposed authority. A
generated or code-backed registry is the runtime representation.
`tools/schema_object_ownership_manifest.json` and similar tooling inputs MUST
remain validation-only and MUST NOT be read at runtime. The current 92-table
baseline in Appendix B is characterization evidence, not a public physical
schema contract.

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| No public Recovery HTTP or WebSocket route | Recovery/Core 04 | Core 04 and static route inventory | `TestPublicRouteAbsenceStaticInventory_Unit` and server-process sentinel | Preserve both static inventories through moves | critical | Adding a route requires later authorization. |
| Exactly five Recovery operator commands | Recovery operator surface | Core 01/Core 04 and `operatorcli` | Five canonical process tests | Add table-driven parser/error/timeout characterization before splitting | critical | Preserve flags, defaults, JSON stdout, JSONL progress, exit codes, and timeouts. |
| Backup identity, consistency point, durability, 24-hour freshness, and 30-day retention | Recovery | Core 01, migration constraints, store/capture/catalog code | Metadata unit and integration tests | Baseline all tie-break, due-selection, and atomic-completion cases | critical | Table or selection changes require later authorization. |
| Encrypted backup storage and recovery journal | Recovery/Core 04 | Core 04, envelope schemas, encryption code | Wrong-key, encrypted-storage, and journal tests | Add typed-error and no-secret-output characterization | critical | No plaintext fallback. |
| Canonical integrity manifest, Seaweed manifest, and redacted summary | Recovery | Artifact encoders/validators | Object-store artifact tests | Route these tests to an explicit owner row | high | JSON shape, digest, duplicate-key rejection, and redaction are frozen. |
| Stopped-empty restore admission before reads or mutation | Recovery | Core 01/Core 04 and restore runner | Extension and restore owner tests | Preserve call-order assertions at the new facade | critical | Admission order is observable. |
| Postgres, object-store, then projection restore order | Recovery orchestration | Core 01 and restore runner | Restore readiness/order and projection tests | Add facade-level step and failure-boundary characterization | critical | Do not reorder or partially publish readiness. |
| Exact extension binding, numeric group/order, and current/historical codec | Extension facts consumed by Recovery | Extension NLSpec and generated registry | Extension catalog and contract tests | Preserve generated-registry fixture coverage | critical | Generated artifacts are changed only from authored inputs. |
| Structured projection rebuild request, provider result, and readiness | Recovery contract; Projections implementation | Core 01 and `restorecontract` | Recovery projection and Projections rebuild tests | Baseline both sides of the adapter before moving wiring | critical | `ProjectionRebuilder` remains Recovery-owned. |
| Verification every seven days or on basis change, with atomic pass/fail recording | Recovery | Core 01, SQL queries, verification/store code | Metadata due-selection and restore owner tests | Explicitly route due-selection and basis-change cases | critical | Preserve deterministic selection order. |
| Verification artifact and owner-defined built-in workbook query | Recovery orchestration; Timeline registration; Workbook executor | Core 01, artifact code, workbook probe | Workbook probe and browser restore test | Materialize Appendix A as the RS-00 canonical fixture | high | RB-002 is design-complete; exact current request and selector are frozen. |
| MinIO-to-SeaweedFS physical bucket/key and byte preservation | Proposed `module.objectstoremigration`, with Evidence/platform collaborators | Migration code and dedicated target | Migration unit and preservation tests | Preserve pass/mismatch, zero-byte, quiescence, redaction, resume, and rollback cases | critical | No Evidence `storage_ref` rewrite; no sixth Recovery command. |
| Encrypted operator journal and safe administrative-audit summary | Recovery plus platform audit | Journal code, migration, Core 04 | Journal and administrative-audit tests | Add typed operation/result mapping coverage | high | Authorization is deployment-local operator control, not a browser capability. |
| Generated Recovery SQL and extension backup registry | Source SQL and Extension contracts | Authored SQL/contracts and generated Go | Store/integration/Extension tests; drift targets | Run generated drift after any authorized input change | high | Never hand-edit generated files. |
| Entities, saved views, revisions, blobs, and restored workbook state | Source-owner contributions; Recovery orchestrates restoration | Restore SQL/artifacts, Appendix B, and browser test | Restore consistency and browser fixture | Generate and owner-review `cartulary.recovery_state_baseline.v1` before replacing raw access | high | No direct public row/query/mutation or saved-view contract is owned here. |
| Harness and evidence accounting | Verification/harness owners | Owner contract, test family, topology, and section 8.1 | Eleven current owner rows plus special migration target | Add exact rows in section 8.1 without duplicate ownership | medium | Rows verify behavior; they MUST NOT define runtime architecture. |

No generated TypeScript protocol or view contract dedicated to a public Recovery
route was found. The frontend relationship is limited to the browser restore
fixture and restored workbook behavior; no production frontend Recovery
controller, saved-view implementation, selector contract, or grid-vendor adapter
was found.

### 4.1 Current artifact and token freeze

The following identities are current implementation contracts. An ownership move
MUST preserve them byte-for-byte and token-for-token. A later adopted contract
MAY version them; omission of that authorization means no redesign is allowed.

| Artifact family | Current schema ID | Current role |
| --- | --- | --- |
| Backup integrity manifest | `cartulary.backup_integrity_manifest.v2` | Binds backup set, consistency point, restore anchors, artifact proofs, encryption proof, and Extension bindings. |
| Postgres snapshot | `cartulary.postgres_snapshot_artifact.v1` | Canonically ordered authoritative table rows. |
| Object-store snapshot | `cartulary.object_store_snapshot_artifact.v2` | Restorable object bytes, size, content type, and SHA-256. |
| Backup envelope | `cartulary.backup_artifact_envelope.v1` | AES-256-GCM backup artifact envelope. |
| Operator journal envelope | `cartulary.operator_recovery_journal_envelope.v1` | AES-256-GCM operator recovery journal payload. |
| Private object-store manifest | `cartulary.object_store_backup_manifest.v1` | Per-object private restore evidence and manifest digest. |
| Redacted object-store summary | `cartulary.object_store_backup_summary.v1` | Shareable derivative summary; never restore input. |
| Restore verification | `cartulary.restore_verification.v1` | Backup selection, object proofs, incident-open check, workbook probe, and result. |
| Migration run | `cartulary.object_store_migration_run.v1` | Migration state and event history. |
| Migration copy ledger | `cartulary.object_store_migration_copy_ledger.v1` | Per-object transfer outcomes and proofs. |
| Migration validation | `cartulary.object_store_migration_validation.v1` | Source/target identity, checks, diagnostics, and result. |
| Migration rollback | `cartulary.object_store_migration_rollback.v1` | Rollback boundary evidence. |
| Migration quiescence | `cartulary.object_store_migration_write_quiescence.v1` | Stopped process and closed-listener proof. |
| Migration target probe | `cartulary.object_store_migration_target_probe.v1` | Target probe evidence. |
| Extension backup registry | `cartulary.extension_backup_registry.v1` | Extension-owned bindings and exact codecs consumed by Recovery. |

### 4.2 Defaults, omissions, and failure consequences

| Boundary | Default or omitted input | Required result |
| --- | --- | --- |
| Backup freshness | No narrower deployment value is authorized. | Latest successful retained backup MUST be no older than 24 hours. |
| Retention | No narrower deployment value is authorized. | Successful backup metadata and restorable artifacts MUST be retained at least 30 days. |
| Verification cadence | No basis change occurs. | Full verification is due after seven days; a basis change makes it due immediately. |
| Restore target | Stopped-empty proof is absent or invalid. | Fail before artifact read or target mutation; readiness remains false. |
| Workbook restored incidents | No incidents exist. | Skip query only after restore, projection rebuild, and invariant validation pass. |
| Workbook query rows | Current registration returns zero rows. | Success because `row_requirement='zero_rows_allowed'`. |
| Workbook `group_by` | Omitted. | Execute with no grouping. Explicit `null` is invalid in the proposed registration. |
| Migration cutover | No adopted automation exists. | Preserve current manual or partial automation; do not activate automatically. |
| Migration expected digest | Evidence entry has no existing digest. | Compute source and target SHA-256; absence does not waive byte proof. |
| Migration rollback after cutover | Requested. | Automatic destructive rollback is forbidden; require separately admitted reverse migration or recovery procedure. |
| Unknown generated contract member or token | Encountered by a strict current decoder. | Fail closed before use; do not coerce, alias, trim, or case-fold. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Recovery is valid orchestration, but its package is not a thin facade. | Fourteen production files expose services, stores, wire DTOs, artifacts, platform settings, and migration state. | high | `must_fix` | Recovery application facade | Define a narrow semantic entry surface before moving adapters. |
| Operator operations depend directly on CLI DTOs, progress, and error mapping. | `operatorops` imports `operatorcli`; `MapError` matches error text. | critical | `must_fix` | Recovery service with operator CLI adapter | Specify typed commands/results/errors and characterize unchanged wire output. |
| Journal logic combines Recovery semantics, Postgres persistence, encryption, and administrative audit. | `operatorcli/journal.go` imports platform Postgres and audit packages. | high | `must_fix` | Recovery journal port plus platform adapters | Freeze encrypted payload and redacted audit output before splitting. |
| The exported store exposes generated sqlc persistence to Recovery services. | `Store` wraps `internal/gen/sql` and is accepted by capture/catalog/restore/verification constructors. | high | `should_fix` | Private Recovery persistence adapter | Introduce semantic persistence ports without changing SQL. |
| Authoritative snapshot membership is a hand-maintained Recovery predicate spanning source owners. | `IsAuthoritativePostgresSnapshotTable`, Appendix B, and raw capture/restore table operations. | critical | `must_fix` | Source-owner `cartulary.recovery_state_contribution.v1` registry | Materialize the exact 92-table baseline and prove set parity; tooling manifests remain validation-only. |
| Evidence blob lookups and migration preflight use raw SQL and Evidence key rules. | Raw `object_blobs`/evidence access and `evidence/blobref` import. | high | `must_fix` | Evidence-owned capability consumed by Recovery | Specify narrow lookup and validation ports with byte-for-byte characterization. |
| Concrete workbook verification is hard-coded to Timeline inside Recovery. | `workbook_probe.go` imports view-schema support and uses the exact Appendix A request; Timeline assembly wires it. | high | `should_fix` | Timeline registration and Workbook executor composed by Recovery assembly | Characterize Appendix A, then move only the concrete adapter. |
| Recovery semantics depend on platform Postgres/object-store settings and factories. | `RestoreTarget` and `operatorops.Deployment` expose platform types. | high | `should_fix` | Application composition and platform adapters | Introduce consumer-owned ports after facade characterization. |
| Object-store migration combines operational orchestration, Evidence physical-key policy, and platform copy mechanics. | Large migration state machine, raw blob access, object-store operations, and dedicated test target. | critical | `must_fix` | Proposed `module.objectstoremigration`, `module.evidence`, `platform.objectstore`, and deployment configuration | Adopt the section 3.1 owner split before relocation; retain current code as rollback until parity passes. |
| Direct grid-vendor, saved-view, collaboration, or frontend shell coupling was not found. | Target imports and inbound caller scan. | low | `intentional/no_action` | Existing owners | Retain as an explicit non-goal. |
| Public Recovery route and browser authorization logic are intentionally absent. | Core 04 and route sentinel. | critical | `intentional/no_action` | Deployment-local operator surface | Preserve absence; do not introduce a public capability check. |
| Generated files are downstream projections, not edit targets. | sqlc and extension registry generated roots. | high | `intentional/no_action` | Authored SQL/contracts and generators | Update only owner inputs in a later authorized task. |
| Some Recovery tests are not selected by current owner rows. | Comparison of live `Test*` functions with `module.recovery` family selectors. | medium | `should_fix` | Accountable owners in section 8.1 | Implement exact-symbol routing in RS-00; rows MUST NOT dictate package placement. |
| No test-only helper is imported by production Recovery code. | Import and package scan. | low | `intentional/no_action` | Recovery tests | Keep test fixtures out of production assembly. |

### 5.1 Required owner-document and contract placement

The tracker MUST NOT be treated as the permanent owner of the following
requirements. Production implementation MUST remain blocked until the applicable
owner material is adopted.

| Material | Required permanent location | Tracker disposition |
| --- | --- | --- |
| Object-store backend migration owner allocation | Core 00 §5.1 | Propose `module.objectstoremigration`; adopt before RS-06A. |
| Migration admission, immutable plan, replay, preservation, validation, cutover proof, and rollback | New Core 01 §12.4 adjacent to backup/restore | Promote section 3.1 semantics without package paths or physical names. |
| Base restore-probe registration identity, owner, deterministic selector, query, and registry semantics | Core 01 §12.2 | Promote Appendix A behavior; keep exact fixture in a machine contract. |
| Source-owner recovery-state contributions and complete-coverage proof | Core 01 §12.1–§12.2 | Promote section 3.3 logical rules, not physical table names. |
| Evidence logical reference, private physical address, lifecycle eligibility, and no-reference-rewrite consequence | Core 01 and Core 02 Evidence owner sections | Preserve Evidence identity; do not expose private addresses. |
| Deployment-local migration/recovery authority, exclusion, no-listener rule, secrets, encryption, audit, and route absence | Core 04 | Promote security consequences and binary acceptance. |
| Registration, contribution, baseline, plan, ledger, completion-proof, and artifact schemas | Versioned machine contracts or an adopted subsystem owner | Generate from owner inputs; never parse this tracker at runtime. |
| Exact table names, old/new parity set, query fixture, current test symbols, package paths, and import allowlists | Characterization appendices and implementation support | Appendices A through D; not public interoperability contracts. |
| Slice sequence, rollback units, and authorization record | This tracker | Maintain in sections 7, 9, and 10. |
| Claim-bearing performance statements | Core 05, only if introduced | Not applicable; this refactor proposes no public performance claim. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Reconfirm authority, clean scope, and tracker history before work. | Tracker and owner documents only | `git status --short`; source inspection | Authority posture and allowed changes recorded. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Maintain a complete file, caller, dependency, and artifact inventory. | `internal/modules/recovery/**` and direct callers | Repository searches and exact file reads | Every target file has an inventory row. |
| WF-02 | Contract-owner mapping | parallel | WF-01 | WF-05 | Map each observable contract to its normative and implementation owner. | Core/Extension owners, Recovery contracts, SQL, application assembly | Owner review; `make explain-test-owner OWNER=module.recovery` | Every freeze-map row has an owner and evidence. |
| WF-03 | Characterization and exact routing | parallel | WF-01 | WF-05, RS-00 | Materialize Appendices A–D as machine fixtures and account for exact test symbols without changing behavior. | Recovery/migration tests, browser fixture, owner manifests | Owner slices, migration preservation, selector audit | Every current baseline has exact executable evidence and no symbol is multiply owned. |
| WF-04 | Boundary and coupling scan | parallel | WF-01 | WF-05 | Classify transport, persistence, platform, projection, and cross-owner coupling. | Recovery production files and application assembly | `make backend-module-boundary-check` plus source inspection | Every finding has a classification and proposed owner. |
| WF-05 | Facade and ownership redesign plan | chain | WF-02, WF-03, WF-04 | WF-05A, WF-06 | Define the interfaces, defaults, errors, ordering, and owner directions in sections 3–5. | Recovery, migration, Workbook, Timeline, Evidence, platform, and application assembly | Two-implementer and recreatability review against Appendices A–D | Every design choice is closed or identified as intentional implementation freedom. |
| WF-05A | Adopted-owner gate | chain | WF-05 and completed RS-00 evidence | Production slices RS-03, RS-04, RS-06A | Promote required behavior to the owner locations in section 5.1. | Core owner sections and versioned machine-contract inputs named by separate authorization | Owner-document review and applicable drift targets | Adopted owners and machine schemas exist; the tracker is no longer the only statement of behavior. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07 | Sequence the smallest reversible behavior-preserving implementation slices. | Packages named by each slice | Per-slice commands in section 7 | Each slice has exact inputs, outputs, defaults, errors, authorization, rollback, and binary acceptance. |
| WF-07 | Harness/test/accounting update plan | chain | WF-06 and RS-00 | WF-08 | Update authored verification ownership only where section 8.1 requires it. | Verification owner inputs and authored topology inputs | Owner slices, drift checks, browser and migration targets | RB-004 evidence passes and generated outputs are produced only by generators. |
| WF-08 | Validation and final handoff | chain | WF-07 | none | Run narrow-to-broad validation and leave current evidence. | All authorized changed surfaces | `make agent-finalize` then required broader targets | Results, failures, run roots, skipped checks, and rollback state recorded. |

The planning dependency chain is:
`WF-00 → WF-01 → {WF-02, WF-03, WF-04} → WF-05 → WF-06 → WF-07 → WF-08`.
RS-00 MAY execute after WF-03 and WF-05 without adopted behavior changes.
RS-03, RS-04, and RS-06A MUST NOT execute until RS-00 passes and WF-05A
adopts their applicable owner requirements.

## 7. Proposed Refactor Slice Plan

All slices require a later authorized implementation task. They are
behavior-preserving unless the row explicitly says otherwise.

| Slice ID | Depends on | Normative input | Normative output | Default, omission, and error behavior | Intended change and likely packages | Required authorization | Validation | Rollback unit | Binary acceptance |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| RS-00 | WF-03, WF-05 | Live behavior, Appendices A–D, and current owner manifests | Characterization fixtures and exact owner/collaborator routing only | No production behavior may change; missing baseline data is an error, not a license to infer it | Add exact query/state/migration fixtures and section 8.1 routing in named tests and authored harness inputs | RS-00-only authorization; production Go, Core requirements, schemas, migrations, and package moves forbidden | Recovery unit/service-backed slices; browser restore; migration preservation; harness and generated drift | New tests, authored routing inputs, generated outputs produced from them, and tracker entries | Every Appendix baseline has executable parity evidence; every affected symbol is selected exactly once. |
| RS-01 | RS-00 | Existing five-command requests, CLI results/progress, and current errors | Closed Recovery command/result/progress/error values adapted to identical CLI bytes and exit codes | Current defaults, timeout bounds, JSON stdout, JSONL progress, and error vocabulary remain exact; unknown typed errors map to the existing closed fallback | Introduce Recovery application facade; adapt `operatorops` and `operatorcli` | Separate RS-01 authorization; no wire or command change | Recovery owner slice and `make build-operator` | New facade and adapter wiring as one reversible unit; old path retained until parity | Process evidence is unchanged and semantic operations no longer depend on CLI DTOs. |
| RS-02 | RS-01 | Current sqlc-backed metadata operations | Private Recovery persistence interfaces with identical records and transactions | 24-hour freshness, 30-day retention, deterministic ordering, and atomic completion remain exact; persistence errors retain current semantic mapping | Privatize `store.go` access across capture, catalog, restore, verification, and assembly | Separate RS-02 authorization; SQL and migrations forbidden | Recovery unit and service-backed slices | Interface adapters and constructor wiring; authored SQL untouched | No semantic service exposes sqlc types and all metadata tests remain unchanged. |
| RS-03 | RS-00, WF-05A | Adopted contribution schema, Appendix B baseline, source-owner contributions | Frozen registry, canonical digest, and capture/restore execution through registered algorithms | Missing/unknown/duplicate/cyclic facts fail before publication or mutation; no tooling runtime fallback | Add owner contributions, registry input/generation, capture/restore integration, Evidence/Revisions ports, and Recovery assembly | Separate RS-03 authorization naming source-owner and generated input/output paths; no table-membership change | Recovery service-backed slice, boundary check, generate drift, artifact policy | New registry path; retain current predicate/raw path until exact parity passes | Old and new sets, artifacts, digests, row identities, versions, revisions, audit, Evidence state, and bytes are equal. |
| RS-04 | RS-00, WF-05A | Adopted registration schema and Appendix A fixture | Timeline contribution, Workbook registry/executor, and Recovery-owned probe capability | Exactly one Base default; zero incidents skip at the defined stage; zero rows succeed; invalid registration fails before execution; query failures map exactly | Move concrete query from Recovery into Timeline/Workbook ownership and compose in Recovery assembly | Separate RS-04 authorization; no query, selector, view, browser, or error change | Recovery owner slice, Timeline/Workbook owner tests, browser restore | New registration/executor/composition; existing Recovery probe retained until parity | Exact fixture matches; Recovery imports no Timeline/view-schema implementation; browser behavior is unchanged. |
| RS-05 | RS-01, RS-02 | Typed Recovery operations and current journal/audit results | CLI adapter, journal persistence adapter, audit adapter, and typed error mapping | No plaintext fallback; unknown errors use current closed fallback; journal append and redacted audit behavior remain exact | Split CLI parse/encode and journal/audit mechanics from Recovery service | Separate RS-05 authorization; wire, journal schema, audit vocabulary, and encryption changes forbidden | Recovery slice, `make go-gosec-targeted`, `make build-operator` | CLI, journal, audit adapters and composition as one unit | Recovery service has no CLI dependency; process, encryption, and audit evidence is unchanged. |
| RS-06 | RS-03, RS-05 | Recovery semantic ports and current platform behavior | Postgres/object-store adapters injected at composition | Borrowed resources remain borrowed; created resources close in reverse order; storage keys, bytes, and readiness remain exact | Isolate platform types from capture, restore, artifacts, and deployment loading; exclude migration move | Separate RS-06 authorization; storage behavior and resource ownership changes forbidden | Recovery service-backed slice and boundary check | One platform adapter capability at a time | Semantic Recovery packages expose no platform settings/concrete stores and preserve storage evidence. |
| RS-06A | RS-06, WF-05A | Adopted migration owner, Appendix C tokens/artifacts, Evidence plan, platform backend | `module.objectstoremigration` service with completion proof and constrained rollback | Same-digest resume; different-digest pre-mutation failure; exact bucket/key preservation; no reference rewrite; no automatic cutover by default | Relocate/split migration across migration, Evidence, platform storage, secure artifacts, deployment cutover, and operator composition | Separate high-risk RS-06A authorization; no token, artifact, route, command, or schema redesign | Migration owner slice, `make seaweedfs-migration-preservation`, security and boundary checks | Entire current Recovery migration implementation retained until full parity | State/artifact tokens, bytes, keys, refs, failure order, resume, redaction, and rollback evidence match. |
| RS-07 | RS-04, RS-05, RS-06, RS-06A when adopted | Passing slice evidence and section 8.1 map | Authored boundary/verification inputs and generator-produced projections | A test has one accountable owner; collaborators do not duplicate ownership; generated files are never hand-edited | Update boundary guardrails and authored verification/topology inputs | Separate RS-07 authorization naming authored inputs and permitted generated outputs | Owner explanations/slices, boundary check, generate drift, artifact policy | Authored inputs plus their reproducible generated projections | Exact-once symbol routing, valid collaborator graph, public migration target selects preservation row, and no drift. |
| RS-08 | RS-07 | All parity evidence and no remaining legacy caller | Removal of superseded adapters plus final handoff | Cleanup is forbidden while any parity criterion or caller remains unresolved | Remove superseded paths and update tracker handoff | Separate RS-08 authorization; no behavior change | `make agent-finalize`, then `make check` and slice-specific gates | Cleanup change only; compatible adapters remain the rollback point | Required checks pass and the tracker records exact commands, results, run roots, and skips. |

Any proposed change to route absence, CLI wire shape, error vocabulary, artifact
schema, table membership, storage semantics, authorization, Extension codec,
projection request/result shape, database schema, or migration is
`requires later authorization`, even when discovered during a
behavior-preserving slice.

## 8. Validation Plan

### 8.1 Exact proposed verification ownership

The following rows are the decision-complete routing target. Exact top-level test
symbols are ASCII-sorted within each row. Generated row IDs, schedules, and
topology outputs MUST be produced by the harness generators and MUST NOT be
invented or hand-edited from this table.

| Proposed verification identity | Exact test symbols | Accountable owner | Collaborators | Runtime, resource, and fixture profile | `default_check` | Routing consequence |
| --- | --- | --- | --- | --- | --- | --- |
| `recovery.verification-selection-and-atomic-completion.v1` | `TestRestoreVerificationDueSelectionAndAtomicCompletion` | `module.recovery` | none | `default`; `go_transaction_heavy`; `postgres_transaction` | `true` | Add one Recovery service-backed row. |
| `recovery.object-store-backup-artifact-codec.v1` | `TestSupportObjectStoreBackup_DuplicateArtifactKeysRejected`; `TestSupportObjectStoreBackup_ObjectStoreBackupManifestCanonicalAndSummaryRedacted` | `module.recovery` | none | `none`; `go_balanced`; `none` | `true` | Add one cheap Recovery unit row. |
| `recovery.object-store-backup-capture.v1` | `TestSupportObjectStoreBackup_CaptureSeaweedFSS3BackupArtifactsFromObjectStore` | `module.recovery` | `platform.objectstore` | `none`; `go_io_heavy`; `none` | `true` | Add one filesystem-backed Recovery unit row. |
| `recovery.operator-journal-security.v1` | `TestOperatorRecoveryJournalEncryptsPayloadAndAuditIsSafe` | `module.recovery` | `platform.audit` | `default`; `go_transaction_heavy`; `postgres_transaction` | `true` | Add one Recovery security/service-backed row. |
| `recovery.restore-projection-contract.v1` | `TestRestoreProjectionRebuildReadinessFailsClosed`; `TestRestoreProjectionRebuildReceivesStructuredRequest` | `module.recovery` | `module.projections` | `default`; `go_clone_heavy`; `postgres_template_clone` | `true` | Add one Recovery service-backed collaborator row. |
| `recovery.restore-selection.v1` | `TestRestoreCandidateUsesObjectStoreBackupLatestTieBreakers` | `module.recovery` | none | `default`; `go_transaction_heavy`; `postgres_transaction` | `true` | Add one Recovery service-backed selection row. |
| `recovery.object-store-restorable-bytes.v1` | `TestObjectStoreSnapshotCarriesRestorableBytes` | `module.recovery` | `platform.objectstore` | `none`; `go_io_heavy`; `none` | `true` | Add one Recovery filesystem-backed unit row. |
| `recovery.restore-workbook-probe.v1` | `TestRestoreVerificationWorkbookProbe` | `module.recovery` | `module.timeline`, `module.workbook` | `default`; `go_transaction_heavy`; `postgres_transaction` | `true` | Preserve the cross-boundary postcondition under Recovery. |
| `timeline.restore-probe-registration.v1` | `TestRestoreWorkbookProbeRegistrationExactFixture_Unit` (new in RS-04) | `module.timeline` | `module.recovery`, `module.workbook` | `none`; `go_balanced`; `none` | `true` | Add only when the Timeline registration exists; assert Appendix A exactly. |
| `object-store-migration.state-machine.v1` | `TestSupportObjectStoreMigrationBlobReferencePreflight`; `TestSupportObjectStoreMigrationBlobReferencePreflightRejectsInvalidKeysBeforeBackendCalls`; `TestSupportObjectStoreMigrationCopyLedgerStatusesAndZeroByteObjects`; `TestSupportObjectStoreMigrationRunCanonicalStateGuardsAndRedaction`; `TestSupportObjectStoreMigrationValidationCanonicalDigestDuplicateKeysAndResult`; `TestSupportObjectStoreMigrationWriteQuiescenceRejectsOperatorAssertionOnly` | `module.objectstoremigration` | `module.evidence`, `platform.objectstore` | `none`; `go_balanced`; `none` | `true` | Create after owner adoption; do not place permanently under Recovery. |
| `object-store-migration.preservation.v1` | `TestSupportSeaweedFSMigrationPreservationMismatchEvidence`; `TestSupportSeaweedFSMigrationPreservationPassEvidence` | `module.objectstoremigration` | `module.evidence`, `platform.objectstore` | `default`; `go_io_heavy`; `service_stack` | `false` | The existing `make seaweedfs-migration-preservation` target MUST select this exact row after adoption. |

Additional routing rules:

- `extension_catalog_test.go` is helper-only and MUST NOT receive an executable
  row.
- The five canonical Recovery operator process tests remain accountable to
  `module.recovery`.
- `TestMVPObjectStoreInitOperatorCreatesConfiguredBucket` remains separately
  owned by the platform/object-store initialization contract.
- A collaborator identifies participation in the asserted postcondition. Import,
  package path, filename, or maintainer identity is insufficient.
- No test symbol may appear under two accountable owners. A special Make target
  is an execution entry, not an owner.

### 8.2 Canonical validation commands

These commands were discovered from the repository's public Make surface and
explanatory targets. No product validation suite was run during this
documentation-only session.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make test-slice OWNER=module.recovery` | Recovery owner rows that do not require service-backed execution | yes | Use `ROWS=` for the narrowest affected behavior after inspecting the owner map. |
| integration | `make service-backed-test-slice OWNER=module.recovery` | Recovery owner rows requiring Postgres/object-store/browser services | yes | Baseline before storage or restore changes; use exact rows where possible. |
| migration owner | `make test-slice OWNER=module.objectstoremigration` and `make service-backed-test-slice OWNER=module.objectstoremigration` | Future migration unit and preservation rows | no | Required after WF-05A creates the owner; before adoption, the owner is expected to be unknown. |
| e2e/browser | `make browser-e2e-webserver-backed` | Restored server and workbook surface/query fixture | yes | Required before moving workbook/projection wiring or restore composition. |
| object-store migration | `make seaweedfs-migration-preservation` | MinIO-to-SeaweedFS byte/key preservation and mismatch evidence | yes | Required only for migration or shared object-store adapter slices. |
| generated drift | `make generate-drift` and `make generated-artifact-policy-check` | Generated SQL, contract projections, and generated-root policy | no | Required after any later authorized authored input change. |
| migration drift | `make migration-drift` | Authored migration/schema projection | no | Run only if a later schema change is explicitly authorized. |
| import-boundary/static | `make backend-module-boundary-check` | Backend module dependency policy | yes | Run before design finalization and after each package move. |
| security | `make go-gosec-targeted` | Recovery encryption, journal, CLI, and storage-sensitive code | no | Required for RS-05, RS-06, and RS-06A. |
| operator build | `make build-operator` | Operator composition and CLI build | no | Required for CLI/operations slices. |
| finalization | `make agent-finalize` | Repository finalization and retained-run preparation | no | Run before broader end-of-run verification. |
| full check | `make check` | Repository-wide verification | no | Final broad check after narrow targets pass. |
| tracker Markdown | `make lint-markdown` | Documentation syntax and repository Markdown policy | no | Required for this tracker-only task. |

Command discovery performed:

- `make task-guide ROLE=module-author OWNER=module.recovery`
- `make explain-test-owner OWNER=module.recovery`
- `make explain-target TARGET=test-slice DETAIL=summary`
- `make explain-target TARGET=service-backed-test-slice DETAIL=summary`
- `make explain-target TARGET=backend-module-boundary-check DETAIL=summary`
- `make explain-target TARGET=generate-drift DETAIL=summary`
- `make explain-target TARGET=browser-e2e-webserver-backed DETAIL=summary`
- `make explain-target TARGET=check DETAIL=summary`
- Repository searches for target files, exported symbols, imports, callers,
  tests, SQL, contracts, generated projections, route handlers, and harness maps.

The current Recovery owner map contains 11 rows, 9 of them service-backed.
Section 8.1 fixes the intended disposition of every identified routing gap.
Before RS-00, the rows remain proposed and the existing broad or special targets
may execute some symbols without owner-level accounting. RS-00 MUST make the
routing exact without duplicating tests or using evidence ownership to dictate
runtime package architecture.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| RT-001 | Establish Recovery scope, authority, and planning-only posture | WF-00 | DONE | none | Section 1 and inspected owner documents | Authority order and allowed write are explicit. |
| RT-002 | Inventory all 29 target files and direct coupling | WF-01 | DONE | RT-001 | Section 2 and live repository scan | Every target file has a populated row. |
| RT-003 | Diagnose the current Recovery boundary | WF-04 | DONE | RT-002 | Sections 3 and 5 | Every discovered responsibility has a disposition. |
| RT-004 | Freeze observable Recovery contracts | WF-02 | DONE | RT-002 | Section 4 | Each discovered contract has an owner, evidence, and test posture. |
| RT-005 | Record exact workbook, state, migration-token, and test-symbol baselines | WF-03 | DONE | RT-004 | Appendices A–D | The three analysis-note evidence gaps are closed in the tracker without guessed behavior. |
| RT-006 | Execute characterization and exact owner-row accounting | WF-03 | TODO | RT-005 | RS-00 and section 8.1 | Machine fixtures and routing prove the tracker baselines without production changes. |
| RT-007 | Adopt object-store migration owner and behavior | WF-05A | TODO | RT-003, RT-005, RT-006 | Section 3.1 and section 5.1 placement | Core and machine-contract owners adopt the migration boundary before RS-06A. |
| RT-008 | Adopt workbook registration and query contract | WF-05A | TODO | RT-005, RT-006 | Section 3.2 and Appendix A | Core and machine contracts adopt the exact Base registration before RS-04. |
| RT-009 | Adopt source-owner recovery-state registry | WF-05A | TODO | RT-005, RT-006 | Section 3.3 and Appendix B | Core and machine contracts adopt the registry; source owners approve current contributions. |
| RT-010 | Approve closed Recovery facade and typed errors | WF-05 | TODO | RT-003, RT-004, RT-006 | RS-01 design handoff | CLI-independent interfaces and exact adapters are owner-reviewed. |
| RT-011 | Implement behavior-preserving production slices | WF-06 | BLOCKED | RT-007 through RT-010, RB-005 | Separate authorization records | Each named slice receives exact authority and all prior gates pass. |
| RT-012 | Update authored verification and boundary accounting | WF-07 | TODO | RT-006 and applicable production slice | Owner manifests and generated outputs | Section 8.1 routing is exact with no generated drift. |
| RT-013 | Run narrow-to-broad verification and final handoff | WF-08 | TODO | RT-011, RT-012 | Make results and updated session log | Required checks pass or failures are recorded with evidence. |

## 10. Session Handoff Log

The initial session created this file and had no earlier history to preserve.
This revision preserves that initial history and appends new entries. Times use
the repository session timezone, America/New_York.

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-29 09:28 EDT | Codex planning session | DONE | Inspected framework, Core recovery sections, domain vocabulary, Extension NLSpec, and live target. Touched only this tracker. | `sed`, `rg`, `git status --short --branch` | Target exists; label is `recovery`; no owner contradiction found. | Later implementation authorization | Obtain authorization only after owner blockers are resolved. |
| 2026-07-29 10:32 EDT | Codex NLSpec-style revision | DONE | Inspected `temp/analysis-notes.md`, NLSpec writing guidance, current tracker, and relevant live evidence. Touched only this tracker. | `sed`, `rg`, `git status --short --branch`, read-only baseline extraction, `make lint-markdown` | Proposed requirements are precise but remain implementation-support evidence pending owner adoption. Markdown lint passed with run root `.cartulary/test-results/20260729T144125Z-p1308166`. | RB-005 and WF-05A adoption gates | Authorize RS-00 only as the first implementation task. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-29 09:28 EDT | Codex planning session | DONE | Inspected all 14 production files, application assembly, Projections, Evidence references, and platform dependencies. Touched only this tracker. | Exported-symbol and inbound-import `rg` scans | Legitimate Recovery orchestration is mixed with transport, persistence, projection-query, platform, and migration concerns. | RB-001, RB-002, RB-003 | Review WF-05 interfaces and owner decisions. |
| 2026-07-29 10:32 EDT | Codex NLSpec-style revision | DONE | Re-inspected migration state/tokens, workbook probe/query defaults, table predicate, migrations, and ownership evidence. Touched only this tracker. | `sed`, `rg`, read-only 92-table owner grouping | RB-001 through RB-003 now have exact proposed owners, interfaces, defaults, and parity gates. | Owner adoption and machine fixtures remain TODO | Execute RS-00, then promote sections 3.1–3.3 to the owners in section 5.1. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-29 09:28 EDT | Codex planning session | DONE | Inspected the browser restore fixture and helper; touched only this tracker. | Route, caller, view-schema, and frontend-reference `rg` scans | No production Recovery frontend shell, saved-view, selector, collaboration, or grid adapter exists; restored workbook behavior is an indirect contract. | RB-002 for concrete workbook query ownership | Preserve the browser fixture through RS-04. |
| 2026-07-29 10:32 EDT | Codex NLSpec-style revision | DONE | Inspected exact Timeline view defaults and Recovery workbook probe/test. Touched only this tracker. | `sed`, targeted `rg` | Appendix A freezes the selector and query; no production frontend Recovery surface is introduced. | Timeline registration fixture and owner adoption | Preserve browser parity through RS-04. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-29 09:28 EDT | Codex planning session | DONE | Inspected Recovery SQL, migration, generated sqlc projection, Extension contract inputs/generated registry, and `restorecontract`. Touched only this tracker. | Contract and generated-artifact `rg` scans; `make explain-target TARGET=generate-drift DETAIL=summary` | Generated surfaces and authored owners are mapped; no public Recovery protocol route contract found. | Later authorization for any schema or contract change | Run generators only from authorized authored input changes. |
| 2026-07-29 10:32 EDT | Codex NLSpec-style revision | DONE | Inspected current Recovery and migration schema IDs, contribution patterns, and document authority. Touched only this tracker. | `rg`, `sed` | Exact current schemas are frozen; proposed registration/contribution/baseline contracts have owner placement and fail-closed rules. | WF-05A adoption; no machine contracts created in this task | Adopt owner inputs separately, then generate rather than hand-edit outputs. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-29 09:28 EDT | Codex planning session | DONE | Inspected all 15 target test/support files, owner contract, family map, catalog, topology, browser fixture, and migration target references. Touched only this tracker. | `make task-guide ROLE=module-author OWNER=module.recovery`; `make explain-test-owner OWNER=module.recovery`; relevant `make explain-target` commands | Eleven owner rows and several explicit routing gaps identified. Explanatory commands passed; no test suite was run. | RB-004 | Complete RS-00 before production refactoring. |
| 2026-07-29 10:32 EDT | Codex NLSpec-style revision | DONE | Re-inspected every live `Test*` symbol, current Recovery family, and migration preservation target. Touched only this tracker. | Exact-symbol `rg`; `make explain-target TARGET=seaweedfs-migration-preservation DETAIL=rows` | Section 8.1 fixes accountable owners, collaborators, profiles, `default_check`, and exact symbol arrays; current target has no owner row yet. | RS-00 and future `module.objectstoremigration` owner | Add exact rows without duplicate ownership. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-29 09:28 EDT | Codex planning session | DONE | Inspected encryption, CLI, journal, audit test, route sentinel, and Core 04 recovery requirements. Touched only this tracker. | Security, route, and authorization `rg` scans | Recovery is deployment-local; encryption, audit redaction, and public route absence are frozen. No browser capability is to be invented. | Later authorization for implementation | Add typed error/security characterization in RS-00 and validate RS-05 with gosec. |
| 2026-07-29 10:32 EDT | Codex NLSpec-style revision | DONE | Re-inspected migration proof, artifact redaction, journal schemas, and route absence. Touched only this tracker. | `sed`, `rg` | Migration secrecy, SHA-256 proof, constrained rollback, no-listener, no-sixth-command, and no-route requirements are explicit. | Core 04 adoption and later slice authorization | Characterize security behavior in RS-00; run targeted gosec for affected production slices. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-29 09:28 EDT | Codex planning session | BLOCKED | Inspected source-owner map, object-store migration, workbook probe wiring, and harness selectors. Touched only this tracker. | Ownership and caller `rg` scans | Four stable blockers are recorded; implementation remains unauthorized. | RB-001 through RB-004 | Resolve owner decisions, authorize RS-00 only, then update this log before work. |
| 2026-07-29 10:32 EDT | Codex NLSpec-style revision | BLOCKED | Inspected analysis closures, live baselines, owner placement, and authorization requirements. Touched only this tracker. | Repository inspection and mapping commands above | RB-001 design is ready for adoption; RB-002 through RB-004 are ready for RS-00; RB-005 remains blocked. | RS-00 authorization and later owner adoption | Issue an RS-00-only task using Appendix D. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Decision: proposed `module.objectstoremigration` owns orchestration; Evidence owns reference validation; platform object storage owns mechanics; deployment configuration owns activation; Recovery owns none of those mechanics. | Moving the current file as one unit would preserve mixed ownership; moving before adoption could change storage and Evidence semantics. | Adopt section 3.1 in the owner locations in section 5.1, then prove Appendix C parity. | READY FOR ADOPTION; production move blocked |
| RB-002 | Decision: Timeline owns the Base registration/query; Workbook owns registry/execution; Recovery owns selection/orchestration/failure mapping; Recovery assembly composes it. | The current concrete query is hard-coded in Recovery despite Core's owner-registration rule. | Materialize Appendix A in RS-00, adopt the registration contract, then execute RS-04. | READY FOR RS-00 |
| RB-003 | Decision: immutable owner-authored `cartulary.recovery_state_contribution.v1` contributions replace predicate/raw cross-owner knowledge. | Table membership spans source owners and tooling metadata cannot become runtime authority. | Materialize Appendix B as a machine baseline, obtain source-owner review, adopt the registry contract, then execute RS-03. | READY FOR RS-00 |
| RB-004 | Decision: section 8.1 defines exact accountable owners, collaborators, selectors, profiles, and `default_check` values. | Missing accounting can hide regressions, while duplicate ownership or architecture inferred from rows is also invalid. | Execute RS-00 and prove exact-once selection plus special-target mapping. | READY FOR RS-00 |
| RB-005 | Implementation authorization has not been granted by this tracker revision. | Every proposed code, test, contract, generated, migration, or harness change is outside this documentation-only write scope. | A later task satisfying Appendix D and naming exact slices and writable files. | BLOCKED |

## 12. Binary Completion Criteria

### 12.1 Tracker revision acceptance

- [x] Every file in `internal/modules/recovery` is inventoried; no target file is
  silently out of scope.
- [x] Every discovered public contract risk has an accountable owner proposal,
  evidence, exact default/failure posture, and required characterization.
- [x] Every workflow has dependencies, adoption gates, and a handoff exit
  checkpoint.
- [x] Every implementation slice specifies inputs, outputs, defaults, errors,
  authorization, validation, rollback unit, and binary acceptance.
- [x] Canonical validation commands are discovered; conditional commands state
  when they apply.
- [x] No owner contradiction was discovered; the required blocked posture is
  documented if one is found later.
- [x] The framework/live-repository mismatch is recorded as an architectural
  planning finding.
- [x] Appendix A records the exact current workbook selector and query.
- [x] Appendix B records all 92 included tables and every current exclusion.
- [x] Appendix C records the current migration schema and token vocabulary.
- [x] Section 8.1 records exact existing test symbols, owners, collaborators,
  execution profiles, and `default_check` decisions.
- [x] Handoff sections identify current evidence, touched-file scope, blockers,
  and the next action without requiring repository rediscovery.

### 12.2 Refactor definition of done

The Recovery refactor is complete only when every item below is true. An
unchecked item is not permission to omit the behavior.

- [ ] RS-00 fixtures reproduce Appendices A through C exactly and every section
  8.1 test symbol is selected once under one accountable owner.
- [ ] Core and machine-contract owners have adopted the section 5.1 material
  required by each production slice.
- [ ] The source-owner registry and old predicate produce exactly the same
  92-table set and artifact set before the predicate is removed.
- [ ] Backup bytes, canonical digests, restore order, record IDs, row versions,
  change sets, revisions, administrative audit, Evidence state, blob metadata,
  and required object bytes are unchanged.
- [ ] Duplicate/missing contribution facts, shared physical claims, unknown
  codecs, unresolved dependencies, and cycles fail before backup publication or
  restore mutation.
- [ ] The workbook registration emits Appendix A exactly, has one Base default,
  rejects conflicts, preserves zero-incident/zero-row behavior, and maps query
  failures to `workbook_probe_failed`.
- [ ] Recovery production packages import no Timeline or view-schema
  implementation package after RS-04.
- [ ] Migration preserves bucket names, keys, source/target byte lengths and
  SHA-256 values, zero-byte objects, and every Evidence `storage_ref`.
- [ ] Invalid Evidence references fail before object-store calls; same-digest
  resume succeeds; different-digest reuse fails before transfer.
- [ ] Pre-cutover rollback deletes only ledger-proved migration-owned target
  objects; no automatic destructive rollback occurs after cutover.
- [ ] The exact five Recovery commands and no-public-route posture remain
  unchanged.
- [ ] Runtime product behavior reads no file under `docs/` or `tools/`.
- [ ] Generated artifacts are produced only from authorized owner inputs and all
  required drift checks pass.
- [ ] Every production slice has a separate Appendix D authorization and a
  recorded rollback unit.
- [ ] Narrow owner slices, browser restore, migration preservation, security,
  boundary, generated-drift, and final repository checks pass or have a
  documented owner-approved exception.

This tracker is complete as a planning artifact. It does not authorize or report
completion of the Recovery implementation refactor.

## Appendix A. Exact Current Workbook Probe Baseline

This appendix is current implementation evidence for RS-00 and RS-04. It is not
an independent Workbook or Timeline owner.

### A.1 Incident selector

The current selector is exactly:

```sql
SELECT id
FROM incidents
ORDER BY id::text ASC
LIMIT 1
```

The selected incident is therefore the non-null incident `id` whose PostgreSQL
UUID text representation sorts first in ascending database text order. If the
query returns no row, Recovery returns probe success without invoking the
workbook query. A database error other than `pgx.ErrNoRows` fails the probe.

### A.2 Query request

The current probe resolves `cartulary.view.timeline.v2` from the code-backed
view-schema registry and passes its default query:

```json
{
  "filters": [],
  "sort": [
    {
      "field_key": "timeline.activity_sort_ts",
      "direction": "asc"
    },
    {
      "field_key": "record_id",
      "direction": "asc"
    }
  ]
}
```

`group_by` is omitted. The query uses the selected incident ID and exact view
schema ID `cartulary.view.timeline.v2`. The returned row collection is ignored;
an empty collection is success. Schema lookup or query error is failure.

### A.3 Parity fixture

RS-00 MUST serialize the selector ID, selector SQL semantics, view schema ID,
normalized query request, zero-incident result, and zero-row result into a
canonical characterization fixture. RS-04 MUST prove that the Timeline
registration emits that fixture exactly. No sort, filter, grouping, limit,
surface, incident selection, or required-row rule may change during the move.

## Appendix B. Exact Current Recovery-State Baseline

The proposed machine identity is
`cartulary.recovery_state_baseline.v1`. This human-readable rendering records
current behavior only. The future machine artifact MUST be generated from
authorized characterization inputs and reviewed by every source owner.

### B.1 Included Postgres tables

The current predicate includes 92 authored public base tables. The evidence-owner
column is derived from `tools/schema_object_ownership_manifest.json` for review
and grouping only; that tool file is not runtime authority.

| Evidence owner | Count | Current included tables |
| --- | --- | --- |
| `artifacts` | 5 | `artifact_findings`, `artifact_forensic_keywords`, `artifact_investigative_queries`, `artifacts`, `handoff_risk_refs` |
| `assessments` | 1 | `assessments` |
| `audit` | 1 | `administrative_audit_projections` |
| `auth` | 6 | `account_preferences`, `bootstrap_tokens`, `enterprise_auth_bindings`, `enterprise_auth_providers`, `enterprise_auth_transactions`, `users` |
| `collaboration` | 4 | `collaboration_event_intents`, `collaboration_incident_stream_cursors`, `collaboration_replay_events`, `collaboration_resume_tokens` |
| `deployment_admin` | 2 | `deployment_admin_audit_events`, `deployment_bootstrap_state` |
| `entities` | 5 | `entity_aliases`, `entity_mentions`, `entity_preserved_identifiers`, `hosts`, `identities` |
| `evidence` | 3 | `evidence`, `evidence_custody_events`, `object_blobs` |
| `extensions` | 6 | `extension_job_cancellation_observations`, `extension_job_commit_proofs`, `extension_migration_ledger`, `extension_staged_object_references`, `extension_staged_objects`, `extension_state_metadata` |
| `graphprojection` | 5 | `graph_projection_edges`, `graph_projection_idempotency`, `graph_projection_runs`, `graph_projection_vertices`, `graph_projection_views` |
| `imports` | 6 | `import_apply_journal`, `import_apply_unit_plans`, `import_sessions`, `import_source_streams`, `import_unit_apply_outcomes`, `import_units` |
| `incidentbundles` | 5 | `incident_bundle_exports`, `incident_bundle_imported_actors`, `incident_bundle_imported_attributions`, `incident_bundle_job_payloads`, `incident_bundle_manifest_files` |
| `incidents` | 4 | `incident_memberships`, `incident_workbook_preferences`, `incidents`, `user_workbook_preferences` |
| `indicators` | 3 | `indicator_observations`, `indicator_state_intervals`, `indicators` |
| `links` | 2 | `record_links`, `record_tags` |
| `networkflow` | 4 | `network_flow_indicator_bindings`, `network_flow_rejected_row_diagnostics`, `network_flow_rows`, `network_flow_tables` |
| `parties` | 1 | `parties` |
| `platform_jobs` | 1 | `jobs` |
| `recovery` | 1 | `operator_recovery_journal` |
| `reference_data` | 4 | `reference_pack_activation_state`, `reference_pack_attestations`, `reference_pack_job_payloads`, `reference_packs` |
| `reportcomposition` | 4 | `report_composition_preview_attempts`, `report_composition_release_bindings`, `report_composition_versions`, `report_compositions` |
| `reporting` | 8 | `reporting_composition_preview_output_files`, `reporting_composition_preview_outputs`, `reporting_job_payloads`, `reporting_release_approvals`, `reporting_releases`, `reporting_render_bundle_files`, `reporting_render_bundles`, `reporting_snapshots` |
| `revisions` | 5 | `change_set_mutations`, `change_sets`, `record_history_entry_refs`, `record_revisions`, `records` |
| `savedviews` | 1 | `saved_views` |
| `tasksdecisions` | 2 | `decisions`, `task_requests` |
| `timeline` | 3 | `timeline_events`, `timeline_source_provenance`, `timeline_time_conversion_profiles` |
| **Total** | **92** | Every table named above, with no additional table. |

### B.2 Current exclusions

The current predicate explicitly excludes these seven authored tables:

```text
backup_sets
evidence_access_handles
pending_totp_enrollments
restore_verification_runs
route_idempotency
schema_migration_lineage
user_sessions
```

It also excludes synthetic `goose_db_version` and every table whose normalized
name ends in `_grid_projection`. The ten currently authored suffix exclusions
are:

```text
artifact_grid_projection
assessment_grid_projection
decision_grid_projection
evidence_grid_projection
host_grid_projection
identity_grid_projection
indicator_grid_projection
party_grid_projection
task_request_grid_projection
timeline_grid_projection
```

The predicate trims surrounding whitespace, lowercases for classification,
rejects the empty name, applies the suffix and exact exclusions, and includes
every other public base table. Snapshot tables are emitted by ascending table
name. Rows in each table are converted to JSON and aggregated by ascending
`row_json::text`.

### B.3 Required parity

Before `IsAuthoritativePostgresSnapshotTable` or the raw restore path is removed:

- the old and new table sets MUST equal the 92 names in B.1 exactly;
- the new exclusion result MUST equal B.2 exactly;
- no table or object artifact may be added, omitted, or reclassified;
- backup artifact bytes and digests MUST remain unchanged;
- restore MUST preserve source row identities, versions, revisions, change sets,
  audit state, Evidence/blob state, and object bytes;
- derived grid projections MUST remain excluded and rebuild successfully;
- the registry digest MUST be stable across process runs;
- Extension bindings MUST remain byte-equivalent to the generated Extension
  registry.

## Appendix C. Exact Current Object-Store Migration Baseline

### C.1 Schema and token inventories

| Vocabulary | Exact current values |
| --- | --- |
| Schema IDs | `cartulary.object_store_migration_run.v1`, `cartulary.object_store_migration_copy_ledger.v1`, `cartulary.object_store_migration_validation.v1`, `cartulary.object_store_migration_rollback.v1`, `cartulary.object_store_migration_write_quiescence.v1`, `cartulary.object_store_migration_target_probe.v1` |
| Run states | `planned`, `preflighted`, `application_stopped`, `backup_captured`, `target_prepared`, `copying`, `copied`, `validating`, `cutover_ready`, `cutover_committed`, `post_cutover_verified`, `rolled_back`, `failed` |
| Events | `plan_created`, `preflight_passed`, `write_quiescence_verified`, `backup_captured`, `target_prepared`, `copy_started`, `copy_completed`, `validation_started`, `validation_passed`, `cutover_committed`, `post_cutover_verified`, `rollback_requested`, `blocking_failure` |
| Copy statuses | `copied`, `already_copied`, `missing_source`, `target_mismatch`, `unsupported_source_feature`, `error` |
| Validation statuses | `pass`, `missing_source`, `missing_target`, `size_mismatch`, `hash_mismatch`, `unsupported_source_feature`, `error` |
| Current source backend token | `minio_s3` |
| Current target backend token | `seaweedfs_s3` |
| Validation schema version | `1.0.0` |
| Current tool version | `cartulary-object-store-migration/2026-07-owner-cutover` |
| Quiescence proof | `proof_kind='process_stopped'`; `process_state` is `absent` or `stopped_by_supervisor`; both listener-closed booleans are `true` |

No item in this table may be renamed, aliased, reordered as a state transition,
or assigned new semantics during RS-06A.

### C.2 JSON member inventory

| Object | Exact current members and optionality |
| --- | --- |
| Run | Required or explicit-null members: `schema_id`, `run_id`, `created_at`, `updated_at`, `current_state`, `state_timestamps`, `events`, `operator_identity`, `source_endpoint_ref`, `target_endpoint_ref`, `source_bucket_ref`, `target_bucket_ref`, `backup_refs`, `probe_ref`, `copy_ledger_ref`, `validation_ref`, `rollback_ref`, `terminal_result`. |
| Event | `sequence`, `event`, `from_state`, `to_state`, `occurred_at`; `detail` is omitted when empty. |
| Backup refs | `backup_set_id`, `integrity_manifest`, `postgres_artifact`, `object_store_artifact`; object-store manifest and summary refs are omitted when absent. |
| Artifact ref | `key`, `sha256`, `size_bytes`, `content_type`. |
| Quiescence proof | `schema_id`, `proof_kind`, `checked_at`, `process_state`, `http_listener_closed`, `websocket_listener_closed`. |
| Copy ledger | `schema_id`, `run_id`, `source_backend`, `target_backend`, `source_bucket_ref`, `target_bucket_ref`, `object_count`, `status_counts`, `items`, `result`, `artifact_sha256`. |
| Copy item | Required: `sequence`, source/target redacted bucket/key refs, `source_size_bytes`, `source_sha256`, `status`, `reason_code`, `idempotency_key_sha256`. `object_blob_id`, `target_size_bytes`, and `target_sha256` are omitted when absent. |
| Validation | `schema_id`, `schema_version`, `validation_tool_version`, `run_id`, `started_at`, `completed_at`, source/target backend, snapshot, and bucket members, `incident_count`, `object_blob_count`, `objects_checked`, `preview_sample_checks`, `blocking_diagnostics`, `nonblocking_warnings`, `result`, `artifact_sha256`. |
| Validation object | Required identity, `storage_ref_sha256`, `status`, and `reason_code`; source/target size and SHA-256 members are omitted when absent. |
| Preview sample | `object_blob_id`, `incident_id`, `route_class`, `status`, `reason_code`. |
| Diagnostic | `diagnostic_id`, `severity`, `reason_code`, nullable `object_blob_id`, nullable `incident_id`, `message`, `refs`. |
| Rollback evidence | `schema_id`, `run_id`, `created_at`, `before_cutover_source_active`, `before_cutover_backup_retained`, `cutover_rollback_procedure`, `post_verification_rollback_closed`. |
| Target probe | `schema_id`, `run_id`, `started_at`, `completed_at`, `target_bucket_ref`, `probe_key_ref`, `result`, `sha256`. |

RS-00 MUST capture canonical examples for every object and each closed token.
RS-06A MUST prove that relocated encoders and decoders preserve the current
canonical bytes, strict duplicate/unknown-member rejection, redaction, and
digest behavior.

## Appendix D. Authorization Contract

The first implementation authorization MUST name RS-00 only.

```text
Authorized slice:
- RS-00 only.

Authorized purpose:
- Add behavior-characterization tests and canonical fixtures.
- Add or correct verification owner/collaborator rows.
- Update authored test-family and topology inputs.
- Update this tracker and handoff evidence.

Allowed files:
- Explicitly named Recovery *_test.go files.
- Explicitly named object-store migration *_test.go files.
- Explicitly named Workbook and Timeline probe test files.
- Exact contracts/verification/owners manifests.
- Exact authored test-family, catalog, and topology inputs.
- docs/handoffs/recovery-module-refactor-tracker.md.

Forbidden:
- Production Go.
- Database queries or migrations.
- Core behavioral requirements.
- Public API, CLI, artifact, token, error, or schema changes.
- Direct generated-file edits.
- Package moves.
- Observable behavior changes.

Required gates:
- Exact Recovery owner slices.
- Recovery service-backed owner slices.
- Browser restore baseline.
- Object-store migration preservation.
- Harness validation.
- Generated drift and generated-artifact policy.

Rollback unit:
- New characterization tests, authored routing inputs, their reproducible
  generated projections, and the corresponding tracker entry.
```

Every later authorization MUST contain:

1. exactly one slice ID or an explicitly justified atomic slice set;
2. exact allowed files or directories;
3. explicit allowed behavior changes, normally `none`;
4. the adopted requirements and machine contracts being implemented;
5. generated-source inputs and permitted generated outputs;
6. the rollback unit;
7. required narrow and broad validation;
8. explicit permission or prohibition for database, CLI, artifact, token, error,
   and schema changes;
9. a statement that every unspecified change is prohibited.

RS-04, RS-03, and RS-06A MUST receive separate authorizations after RS-00 and
their owner-adoption gates pass. RS-01, RS-02, RS-05, RS-06, RS-07, and RS-08
MUST follow the dependencies in section 7 and MUST NOT inherit authority from a
previous slice.
