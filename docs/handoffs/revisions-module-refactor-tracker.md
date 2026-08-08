# revisions Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

This tracker is normative implementation guidance for a later Revisions
refactor. It records required gates, defaults, interfaces, evidence, and
handoff state. It is not an adopted behavioral owner and it is not
authorization to perform the refactor described below.

| Item | Value |
| --- | --- |
| Target path | `internal/modules/revisions` |
| Normalized target label | `revisions` |
| Output path | `docs/handoffs/revisions-module-refactor-tracker.md` |
| Document class | Implementation-support refactor tracker and handoff |
| Document revision | 2, NLSpec-style planning revision |
| Repository baseline | Commit `65d9a6009fc0a0e1cdc2f8f31bc9d965a873ca6f`; the existing tracker was the sole changed path at this revision session start |
| Planning status | Planning and documentation only |
| Allowed change | This tracker file only |
| Non-goals | No production code, tests, contracts, generated artifacts, package configuration, migrations, harness inputs, or runtime behavior changes |
| Later authorization | Every characterization or production slice requires a separate authorized task |

### 1.1 Normative interpretation

The requirements in this tracker constrain later refactor work only within the
behavior and ownership already established by adopted owners. They do not
create product behavior, supersede an owner, or constitute conformance
authority.

| Term | Meaning in this tracker |
| --- | --- |
| `MUST` | The stated condition is mandatory for an authorized refactor slice or for closing the named gate. |
| `MUST NOT` | The stated condition is prohibited for an authorized refactor slice. |
| `MAY` | The condition is optional only when the requirement states the controlling choice and the omission default. |
| Omission default | If this tracker does not expressly authorize a behavior, contract, schema, storage, configuration, ownership, or interface change, that change is not authorized. Existing observable behavior MUST be preserved. |
| Owner conflict default | An adopted owner prevails. Work MUST stop, and Section 11 MUST record `BLOCKED: owner contradiction`; this tracker MUST NOT select a side. |
| Unknown-evidence default | The tracker MUST use `TODO: RS-00 evidence required` and MUST NOT infer an answer from a filename, target label, JSON key, or historical phase name. |

Normative requirements use stable identifiers of the form `RMR-REQ-###`.
Binary acceptance criteria use stable identifiers of the form `RMR-AC-###`.
Rationale and repository findings explain the requirements but are not
independent authorization.

### 1.2 Definitions

| Term | Closed definition |
| --- | --- |
| Authoritative snapshot | The `{record, source}` value returned by the applicable source owner's `DeleteRestoreSource.SnapshotTx` from authoritative envelope and source tables in the caller-owned PostgreSQL transaction. |
| Projection-derived value | Any value read from a projection or view row, including labels, counts, chips, joined summaries, preview handles, grouping, sorting, or layout fields. Projection-derived values are disposable and are not source truth. |
| Observable consequence | Any HTTP or WebSocket result, stored history value, opaque selector, rollback outcome, projection refresh, authorization result, portability result, generated contract surface, frontend behavior, or harness evidence mapping listed in Section 4. |
| Target kind | The exact persisted `change_set_mutations.target_kind` token. Similar source concepts with different exact tokens are distinct target kinds. |
| Target-semantics contribution | One immutable, source-constructed contribution for one exact target kind that supplies history association/addressability and, when applicable, rollback planning/application semantics to Revisions. |
| Source owner | The module that owns the authoritative record or non-row mutation vocabulary and fixed SQL needed to interpret and invert that target. |
| Exact compatibility | Equality at all four layers in Section 4.4 after representation-only differences are normalized as expressly permitted there. |

### 1.3 Owner and artifact placement

| Material | Authority and permitted content | Material that MUST remain elsewhere |
| --- | --- | --- |
| Adopted Core 00-04 | Product behavior, ownership, security, storage, history, rollback, and binary conformance requirements | Go paths and symbol names, fixture hashes, task sequencing, run roots, and session status |
| Core 05 | Claim-bearing timed or fixture-sensitive publication only | This tracker-only task, which publishes no Core 05 claim |
| RS-00 characterization evidence | Executed old/new comparisons, target matrices, fixtures, failure results, and run roots | New product behavior or ownership decisions |
| This tracker | Refactor decisions, defaults, interfaces, authorization gates, work sequencing, blockers, and handoff state | Adoption authority and proof that unexecuted characterization passed |
| Future implementation guide or authorized task | Exact Go types, packages, constructors, adapter wiring, and permitted file set | Changes beyond the named authorization |
| Authored verification inputs | Evidence ownership and test selection | Runtime architecture or behavioral requirements |
| Generated projections | Machine output from adopted owners and authored inputs through Make-owned generation | Hand edits or independent requirements |

The source hierarchy used for this audit is:

1. Adopted subsystem NLSpecs for their named subsystem. No adopted
   Revisions-specific subsystem NLSpec was found.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication. It
   does not apply to this tracker because no such claim is published.
4. Domain vocabulary and implementation-support guidance.
5. Current repository code and tests.
6. Prior plans, handoffs, and the planning framework as evidence only.

Owner and guidance documents inspected:

- `docs/spec/00_document_set_status_and_precedence.md`, especially
  REQ-00-067;
- `docs/spec/01_architecture_storage_and_view_contracts.md`, including the
  Revisions route requirements, REQ-01-649, and REQ-01-650;
- `docs/spec/02_domain_model_schema_and_history.md`, especially REQ-02-204
  through REQ-02-218 and retained-history requirements;
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`,
  especially REQ-03-066 and its Revisions conflict-capability ownership;
- `docs/spec/04_security_deployment_and_conformance.md`, especially AC-525
  through AC-529 and REQ-04-069 plus REQ-04-147 through REQ-04-149;
- `docs/spec/05_claim_publication_and_benchmark_reproducibility.md` for the
  non-applicability determination above;
- `docs/domain.md` for Revisions vocabulary and owner navigation;
- `docs/handoffs/cartulary_modular_refactor_planning_framework.md` as
  planning doctrine; and

Repository evidence inspected includes every tracked file under
`internal/modules/revisions`, the Revisions application and server assembly,
configuration assembly, representative source-owner provider contributions,
the Timeline frontend history port/adapter/model/actions/panel, authored
OpenAPI and Revisions contracts, Incident Bundle row contracts, WebSocket
contracts, backend boundary policy, verification ownership, test-family
routing, and public Make target guidance.

The live package contains 73 tracked Go files: 47 production files and 26 test
files, totaling 14,497 lines. The framework's generic candidate state and the
archived tracker's earlier inventory do not match that live shape. The live
repository already contains dedicated HTTP, conflict, history, and rollback
components created by the earlier refactor. This tracker preserves those
completed boundaries and plans only from the remaining live findings.

No owner-document contradiction was found. Repository deviations from adopted
owners are recorded as implementation findings rather than owner conflicts.

## 2. Current-State Repository Inventory

The table contains one row for every tracked file under the target. “Contracts”
names the observable or generated surface affected by the file; it does not
mean that the file owns or may hand-edit a generated artifact.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/revisions/appender.go` | Appends change sets, mutation entries, record revisions, and transactional Collaboration intents | `Appender`, append parameter types, `HistoricalIntentPolicy`, `IntentAppender`, constructor and append methods | Revision assembly and source-owner mutation facades | PostgreSQL transaction, Collaboration, concrete Records store, record-view catalog | Revisions integration/delete/rollback tests and source-owner mutation suites | Revision rows and `record_changed` payload consequences | Revisions, with Records and Collaboration ports | high | Legitimate facade, but its concrete Records lookup bypasses the injected application port. |
| `internal/modules/revisions/application_boundary_test.go` | Static application/HTTP boundary evidence | `TestRevisionsApplicationAndHTTPBoundaries_Unit` | `module.revisions` support-unit row | Go source inspection | Self | Verification `http_application_boundary` and facade closure | Revisions tests | medium | Protects the completed root-versus-HTTP split. |
| `internal/modules/revisions/application_ports.go` | Transport-neutral command, transaction, authorization, idempotency, and record-envelope ports | `ActorID`, `TransactionRunner`, `CommandAuthorizer`, `IdempotencyPort`, `RecordEnvelopePort`, command/query DTOs | Command service, HTTP adapter, revision assembly | UUID, PostgreSQL transaction types | Command service and application-boundary tests | Internal application interface behind four HTTP operations | Revisions | high | Consumer-owned port surface; no public wire change is planned. |
| `internal/modules/revisions/attribution_resolver_registry.go` | Validates profile-scoped imported-actor resolver contributions | `ExtensionClaim`, `AttributionResolverRegistry` and methods | Revision assembly and Incident Bundle integration | Imported-attribution resolver port | Registry and portability tests | Incident Bundle attribution and profile claims | Revisions with Incident Bundles contribution | high | Deliberately avoids direct sidecar-table joins. |
| `internal/modules/revisions/attribution_resolver_registry_test.go` | Resolver admission and duplicate-profile evidence | Two registry tests and resolver fake | `module.revisions` portability/support rows | Registry API | Self | Attribution verification mapping | Revisions tests | low | Test-only. |
| `internal/modules/revisions/command_models.go` | Normalizes command requests, rollback selectors, hashes, and route keys | `DeleteRestoreRequest`, `RollbackTarget`, `RollbackRequest`, hash helpers | Command service and HTTP decoding | JSON, SHA-256, UUID | Command, route, and rollback tests | Delete/restore/rollback request identity and idempotency | Revisions | high | Internal model mirrors stable wire semantics without owning HTTP decoding. |
| `internal/modules/revisions/command_service.go` | Thin public application facade for history, delete, restore, and rollback | `CommandService`, dependencies, constructor, four methods | Revisions HTTP adapter and revision assembly | Application ports, provider catalogs, history/mutation coordinators | `command_service_test.go` plus route/integration suites | Four Revisions OpenAPI operations | Revisions | high | Legitimate thin facade. |
| `internal/modules/revisions/command_service_test.go` | Constructor closure, provider completeness, and validation-catalog evidence | Three principal tests and dependency fakes | `module.revisions.facade` and provider rows | Application interfaces and provider contracts | Self | Facade and provider verification IDs | Revisions tests | medium | Test-only. |
| `internal/modules/revisions/command_store.go` | Private dependency container used by command components | `ImportedAttributionResolver` interface | Command service and private coordinators | Application ports, catalogs, projection and appender capabilities | Command/integration tests indirectly | None directly; supports public command behavior | Revisions | medium | Private composition state, not a second public facade. |
| `internal/modules/revisions/conflict_field_provider.go` | Startup adapter from global view-schema registry to immutable conflict descriptors | `ConflictFieldProvider`, `NewViewSchemaConflictFieldProvider` | Nine source-owner provider contribution constructors | `internal/platform/viewschema`, Revisions conflicts descriptors | Field-resolver and revision-assembly tests | View-schema writeability/conflict classes | Application composition candidate | high | Startup-only, but platform/global-registry access remains in the Revisions root. |
| `internal/modules/revisions/conflicts/configuration.go` | Revisions configuration overlay, normalization, findings, and manifest-read errors | `Configuration`, findings/errors, overlay and validation helpers | Config assembly and server startup | Path normalization and Revisions config vocabulary | Conflict-token/config and server/config tests | Deployment config v2 Revisions namespace | Revisions configuration semantics | high | Secure file reading remains in server assembly. |
| `internal/modules/revisions/conflicts/conflict_token.go` | Issues and verifies opaque v3 conflict tokens | Token claims/binding/codec, constructor, hash helper | Workbook/source-owner conflict flows | AES-GCM, key ring, clock/entropy injection | Token lifecycle tests and source-owner conflict suites | Conflict token v3 and `same_field_conflict` | Revisions conflicts | very high | Generic conflict mechanic explicitly owned by Revisions. |
| `internal/modules/revisions/conflicts/conflict_token_test.go` | Token sealing, expiry, rotation, entropy, key isolation, and config evidence | Five tests and error reader | Revisions conflict verification rows | Token/config/key-ring APIs | Self | Key-ring schema and v3 lifecycle | Revisions tests | medium | Test-only. |
| `internal/modules/revisions/conflicts/conflict_window.go` | Pure descriptor-driven revision-window conflict calculation | Conflict-window types and build/decode/diff helpers | Source-owner workbook conflict handlers | Immutable field descriptors and revision rows | Resolver/window tests and source-owner conflict suites | Same-field conflict field/window behavior | Revisions conflicts | high | Does not read the global view registry at runtime. |
| `internal/modules/revisions/conflicts/consolidation_test.go` | Guards one generic conflict package and absence of source-owner imports | `TestConflictCapabilitiesAreConsolidated_Unit` | Revisions conflicts verification row | Filesystem/source inspection | Self | Conflict capability consolidation evidence | Revisions tests | low | Test-only architecture evidence. |
| `internal/modules/revisions/conflicts/field_resolver.go` | Immutable field descriptor and resolver catalogs | Descriptor, contribution, resolver, catalog types and constructors | Revision assembly and conflict consumers | Source-contributed copied descriptors | Resolver tests and revision-assembly tests | View-schema field/conflict contract | Revisions conflicts | high | Intended runtime replacement for global lookup. |
| `internal/modules/revisions/conflicts/field_resolver_test.go` | Catalog closure, immutability, fail-closed, and no-runtime-global-read evidence | Three tests | Revisions conflict verification rows | Resolver API and source inspection | Self | Conflict field-resolver verification | Revisions tests | medium | Does not cover physical placement of the startup adapter. |
| `internal/modules/revisions/conflicts/keep_saved.go` | Revisionless `keep_saved` idempotency transaction coordination | Command/idempotency/target/result types, `KeepSaved`, decoder | Workbook conflict resolution flows | Injected transaction runner, idempotency, semantic loader | Source-owner conflict and workbook tests | Explicit conflict-resolution behavior | Revisions conflicts | very high | Source owners retain state loading and revalidation. |
| `internal/modules/revisions/conflicts/key_ring.go` | Strict key-ring JSON parsing, secret resolution, rotation admission, and purpose registration | `ConflictTokenKeyRing`, parse options, parse functions | Server startup and conflict test support | Environment map or process fallback, secret-purpose registry | Conflict-token tests and server startup tests | `contracts/revisions/conflict-token-key-ring.v1.schema.json` | Revisions config with platform startup adapter | very high | Optional `os.LookupEnv` fallback is deferred for ownership review. |
| `internal/modules/revisions/conflicts/revision_window_reader.go` | Reads ordered record revisions inside a caller transaction | `RevisionWindowReader`, constructor, row/error types | Source-owner conflict mutation paths | PostgreSQL and Revisions history tables | `revision_window_reader_test.go` and conflict suites | Conflict revision window | Revisions persistence adapter | high | Owner-table SQL is intentional. |
| `internal/modules/revisions/conflicts/text_merge.go` | Deterministic line-based merge suggestion | `SuggestedTextMergeValue` | Source-owner conflict response construction | Pure string logic | `text_merge_test.go` | Optional merged-value conflict member | Revisions conflicts | medium | Generic mechanic explicitly owned by Revisions. |
| `internal/modules/revisions/conflicts/text_merge_test.go` | Clean/non-clean text merge characterization | Two tests | Revisions conflict verification row | Text merge helper | Self | Text-merge contract evidence | Revisions tests | low | Test-only. |
| `internal/modules/revisions/delete_restore_source_catalog.go` | Fail-closed record-type to delete/restore provider catalog | Registration/catalog types, constructor, lookup | Command service and delete/restore coordinator | Consumer-owned delete/restore port | Catalog, command, delete/restore tests | Admitted record-type provider coverage | Revisions | high | Complete catalog is required by REQ-01-650. |
| `internal/modules/revisions/delete_restore_source_catalog_test.go` | Missing/duplicate/typed-nil catalog evidence | Catalog test and source fake | Revisions provider row | Delete/restore port/catalog | Self | Delete/restore source-port verification | Revisions tests | low | Test-only. |
| `internal/modules/revisions/delete_restore_store.go` | Atomic delete/restore coordination across auth, locks, source state, history, projection, events, and idempotency | Typed errors/results and private command-store methods | Command service | Application ports, source catalog, projections, Appender | Delete/restore, lock, security, integration, and browser tests | Delete/restore routes and consequences | Revisions mutation coordinator | very high | Legitimate coordination; source semantics stay behind providers. |
| `internal/modules/revisions/delete_restore_test.go` | Provider matrix and delete/restore precondition/atomicity evidence | Four tests plus test state helpers | Delete/restore and provider verification rows | App support, source owners, PostgreSQL | Self | Delete/restore route, history, event, and adapter contracts | Revisions tests | high | Service-backed test-only surface. |
| `internal/modules/revisions/deleterestorecontract/source.go` | Consumer-owned source port for authoritative snapshot and delete-state effects | `DeleteRestoreSource`, `ScanSnapshot` | Source-owner delete/restore adapters and Revisions catalog | Caller transaction, record envelope/source values | Provider and delete/restore suites | Internal source-owner contract | Revisions contract package | high | Must remain narrow and transaction-neutral. |
| `internal/modules/revisions/history_attribution.go` | Decorates history with imported actor attribution | Private decorator method | History service | Imported-attribution resolver | History component and portability tests | History actor attribution | Revisions | medium | Generic decoration only. |
| `internal/modules/revisions/history_components_test.go` | Unit and service-backed evidence for history decomposition | Three tests and attribution fake | History component/repository verification rows | History private components and PostgreSQL | Self | History decomposition verification | Revisions tests | medium | Test-only. |
| `internal/modules/revisions/history_helpers_test.go` | Shared same-package history test helpers | No exported surface | Revisions history tests | Test-only fixtures/helpers | History test files | None | Revisions tests | low | Test-only helper; not production support. |
| `internal/modules/revisions/history_materializer.go` | Builds stable history items, ordering, summaries, and addressability | Private materializer/assembler methods | History service | JSON, stable hashing, hard-coded target kinds and source field keys | History components, history, rollback, indicator tests | History item refs, ordering, action eligibility | Revisions plus source-owner metadata | high | Contains central source-specific switches that should move behind contributions. |
| `internal/modules/revisions/history_model.go` | Internal history record/item model and HTTP resource projection | `RecordHistoryRecord`, `RecordHistoryItem`, `Resource` | History service and HTTP adapter | UUID/time and generic resource maps | History route/component tests | History response item shape | Revisions | high | Public response shape must remain stable. |
| `internal/modules/revisions/history_repository.go` | Queries Revisions history tables and persists stable entry refs | Private repository methods | History service and rollback action evaluator | PostgreSQL; Revisions-owned tables; target-shape JSON matching | History repository, route, pagination, retention tests | Retained history and stable selector behavior | Revisions persistence adapter | high | SQL ownership is valid; source-specific JSON interpretation needs capability review. |
| `internal/modules/revisions/history_rollback_actions.go` | Computes currently legal rollback actions for history items | Private evaluator method | History service | Rollback query/planner and provider catalogs | History/rollback tests | `reversible` and `available_rollback_actions` | Revisions | high | Must preserve current legality and selector behavior. |
| `internal/modules/revisions/history_service.go` | Coordinates history record lookup, materialization, attribution, actions, and pagination inputs | `ErrRecordNotFound` and private history-store methods | Command service | Transaction runner, repository, materializer, attribution, rollback evaluator | History component/route/integration tests | GET history behavior | Revisions | high | Legitimate internal service. |
| `internal/modules/revisions/history_test.go` | History envelope/OpenAPI/ref/retention characterization | Four tests | Store/support verification rows | App support and OpenAPI test facade | Self | History route and retained-history contracts | Revisions tests | high | Includes negative absence of purge/retention routes. |
| `internal/modules/revisions/http_security_precedence_test.go` | End-to-end malformed/unauthorized mutation precedence | One integration test | Security-precedence verification row | HTTP app harness, auth, PostgreSQL | Self | Delete/restore/rollback security order | Revisions tests with Auth collaboration | high | Service-backed security evidence. |
| `internal/modules/revisions/httpapi/delete_restore_decoding.go` | Strict delete/restore HTTP body decoding | Private decoder | `httpapi/routes.go` | JSON and application request model | Route, security, delete/restore tests | Delete/restore request envelope | Revisions HTTP adapter | high | Transport-only code stays outside the Revisions root package. |
| `internal/modules/revisions/httpapi/rollback_decoding.go` | Strict rollback selector-union decoding | Private decoder | `httpapi/routes.go` | JSON and application rollback model | Rollback and security tests | Rollback request union | Revisions HTTP adapter | high | Preserve strict member and selector behavior. |
| `internal/modules/revisions/httpapi/routes.go` | Registers and maps four Revisions HTTP operations | `Service`, `RegisterRoutes` | Server runtime assembly | Platform auth/HTTP/pagination plus Records and Incidents adapters | History/delete/rollback/security/integration/browser suites | Revisions OpenAPI operations and envelopes | Revisions HTTP adapter | very high | Intentional transport-adjacent adapter; platform concerns terminate here. |
| `internal/modules/revisions/imported_attribution_boundary_test.go` | Guards against direct Incident Bundle sidecar joins | One boundary test | Portability attribution verification row | Source inspection | Self | Attribution boundary evidence | Revisions tests | low | Test-only. |
| `internal/modules/revisions/incident_bundle_allocator_test.go` | Sequence allocation, locking, repair ordering, and hardening evidence | Five tests | Portability allocator/atomicity rows | PostgreSQL and migration-owned functions | Self | Revisions Incident Bundle allocation contract | Revisions tests | high | Service-backed portability evidence. |
| `internal/modules/revisions/incident_bundle_portability.go` | Deterministic export/import codecs, validation, writes, and sequence lifecycle for Revisions rows | Begin/finish sequence functions plus source-port internals | Incident Bundles source catalog/runtime | PostgreSQL, Incident Bundles ports, attribution registry | Portability and incident-bundle integration suites | Three Revisions NDJSON members and six invariants | Revisions portability adapter | very high | Owns only Revisions history rows, not whole-bundle orchestration. |
| `internal/modules/revisions/incident_bundle_portability_test.go` | Determinism, invariant priority, attribution, and atomic failure evidence | Four tests and portability fakes | Portability verification rows | Incident Bundle harness and PostgreSQL | Self | Revisions bundle versions 1/2 | Revisions tests | high | Service-backed test-only surface. |
| `internal/modules/revisions/incident_bundle_source_port.go` | Declares Revisions source family, files, versions, dependencies, and invariants | `NewIncidentBundleSourcePort` | Incident Bundle application composition | Incident Bundles source-port API | Portability tests and source-catalog checks | `contracts/incident-bundles/source_catalog.json` | Revisions portability adapter | high | Fixed source descriptor, not dynamic relation metadata. |
| `internal/modules/revisions/incident_bundle_validation_catalog.go` | Builds immutable source-owner current-row validation catalog | Validation catalog/envelope reader types and constructor | Revision assembly and portability import | Provider contributions and source snapshots | Command-service and portability tests | Terminal history reconstruction validation | Revisions with source-owner contributions | high | Must not use projections as current-source truth. |
| `internal/modules/revisions/indicator_children_test.go` | Observation/state-interval history and rollback integration evidence | One integration test | Revisions integration row | Indicators, app support, Collaboration | Self | Indicator-child history/rollback/events | Revisions tests with Indicators | high | Test-only cross-owner characterization. |
| `internal/modules/revisions/integration_test.go` | Cross-cutting delete/restore/rollback/history/merge/retention consequences | Five integration tests | Revisions integration rows | Full application support, PostgreSQL, WebSocket test support | Self | HTTP, history, projection, events, retention | Revisions tests | very high | Primary service-backed behavior freeze. |
| `internal/modules/revisions/json_value.go` | Internal JSON marshaling helper for nullable stored values | No exported surface | Appender and rollback/history persistence | `encoding/json` | Integration and rollback suites indirectly | Stored revision/mutation JSON | Revisions | medium | Preserve stored null/JSON semantics. |
| `internal/modules/revisions/locks_test.go` | Destructive-operation lock precedence and protected-set evidence | One unit test | Revisions store row | Locking fakes/private coordinator | Self | Restore/rollback contention behavior | Revisions tests | medium | Test-only. |
| `internal/modules/revisions/nonrow_provider_catalog.go` | Fail-closed non-row rollback provider catalog | Registration/catalog types, constructor, lookup | Command service and rollback components | `rollbackcontract` | Non-row catalog and rollback suites | Non-row target coverage | Revisions | high | Generic lookup is intentional. |
| `internal/modules/revisions/nonrow_provider_catalog_test.go` | Catalog closure and exact apply-result validation | Two tests and provider fake | Provider/rollback rows | Non-row contract/catalog | Self | Non-row rollback provider evidence | Revisions tests | low | Test-only. |
| `internal/modules/revisions/projection_port.go` | Consumer port for projection rebuild/load and record snapshot helper | `ProjectionRebuilder`, `ProjectionServices` | Command store, delete/restore, rollback, revision assembly | Projection adapter and delete/restore source port | Revision-assembly and integration suites indirectly | Projection refresh and revision snapshot shape | Revisions consumer port; Projections implementation | high | `snapshotRecordTx` currently prefers a projection row, contrary to the no-projection-truth posture. |
| `internal/modules/revisions/provider_contributions.go` | Closed owner vocabulary, record/non-row contributions, and complete catalog validation | Owner enum, contribution types, policies, validator | All source-owner contribution constructors and revision assembly | Delete/restore and rollback contracts | Command, provider, revision-assembly tests | Record-type, target-kind, and view-route admission | Revisions | high | Closed completeness maps are intentional; semantic switches outside the catalog are not. |
| `internal/modules/revisions/record_view_catalog.go` | Immutable record-type/variant to public view-schema catalog | Descriptor/catalog types, constructor, descriptors, resolver | Revision assembly and Appender | Provider routes plus projection/view descriptors | Revision-assembly and event tests | `affected_views` and view-schema identities | Revisions with source-owner contributions | high | Generic routing catalog; source owners declare variants/routes. |
| `internal/modules/revisions/recovery_state.go` | Declares authoritative Revisions tables to Recovery | `RecoveryStateContribution` | Recovery application assembly | Recovery state contract | Recovery catalog tests indirectly | Recovery-state catalog | Revisions | medium | Intentional owner contribution. |
| `internal/modules/revisions/revision_window_reader_test.go` | PostgreSQL ordering and UTC normalization evidence | One integration test | Revision-window verification row | Revision window reader/PostgreSQL | Self | Conflict revision-window contract | Revisions tests | medium | Service-backed test-only surface. |
| `internal/modules/revisions/rollback_apply.go` | Applies finalized row/non-row inverse steps and builds snapshots | Private rollback applier methods | Rollback coordinator | Source-owner providers, Records/projection ports, hard-coded target-kind families | Rollback component/failure/integration suites | Rollback source, history, projection, and event consequences | Revisions plus source-owner capabilities | very high | Contains source-kind branching that should move behind provider capabilities. |
| `internal/modules/revisions/rollback_changed_fields_test.go` | Ensures rollback events include every public cell delta | One test | Rollback verification row | Record-view catalog and rollback helpers | Self | `record_changed.changed_field_keys` | Revisions tests | medium | Test-only behavior freeze. |
| `internal/modules/revisions/rollback_components_test.go` | Pure plan determinism, canonical lock order, and decomposition boundaries | Three tests | Rollback-decomposition verification row | Private planner/locker and source inspection | Self | Rollback architecture evidence | Revisions tests | medium | Test-only. |
| `internal/modules/revisions/rollback_coordinator.go` | Owns caller-transaction rollback orchestration and stage order | Private `RollbackRecord` method | Command service | Query, planner, locker, applier, publication, result components | Rollback and integration suites | Rollback atomicity and ordering | Revisions | very high | Legitimate mutation coordinator. |
| `internal/modules/revisions/rollback_failure_injection_test.go` | Proves rollback atomicity across publication stages | One integration test | Rollback atomicity verification row | PostgreSQL failure injection and application harness | Self | No partial source/history/projection/idempotency effects | Revisions tests | very high | Service-backed test-only evidence. |
| `internal/modules/revisions/rollback_locker.go` | Canonicalizes and acquires protected record locks | Private locker methods | Rollback coordinator | Records envelope port | Lock/component/rollback tests | Destructive lock order | Revisions with Records port | high | Generic lock coordination is intentional. |
| `internal/modules/revisions/rollback_model.go` | Internal rollback target/plan/error/result model | `RollbackPreconditionError`, `RollbackResult`, `RollbackRecordChange` | All rollback components and command facade | UUID/time and generic values | Rollback suites | Rollback error/result behavior | Revisions | high | Keep wire-neutral. |
| `internal/modules/revisions/rollback_planner.go` | Finalizes and validates rollback plans and provider identities | Private planner/helpers | Rollback coordinator and history action evaluator | Non-row catalog plus hard-coded first-class kinds/type mapping | Component/provider/integration tests | Rollback target legality and plan order | Revisions plus source-owner metadata | very high | Central record-type/target-kind switches violate the desired source-owner boundary. |
| `internal/modules/revisions/rollback_providers.go` | Adapts row and non-row provider lookups for rollback components | Private provider helpers | Query/planner/applier | Row/non-row catalogs and contracts | Provider/rollback tests | Source-owner rollback dispatch | Revisions | high | Generic dispatch seam. |
| `internal/modules/revisions/rollback_publication.go` | Centralizes history append, projection refresh, and live-change construction after inverse apply | Private publication methods | Rollback coordinator/applier | Appender, projection port, record-view catalog | Rollback failure/integration/event tests | Revision rows, projection refresh, `record_changed` | Revisions coordination | very high | Must retain transaction ordering and failure atomicity. |
| `internal/modules/revisions/rollback_query_repository.go` | Loads selectors, mutations, companions, snapshots, and current target descriptions | Private query repository methods | Rollback coordinator and history action evaluator | Revisions SQL, provider catalogs, source-specific target rules | Repository/component/rollback/integration tests | History selectors and rollback candidate sets | Revisions persistence plus source-owner metadata | very high | SQL is owner-valid; special target handling should be provider-driven. |
| `internal/modules/revisions/rollback_result.go` | Builds stable command result from applied record changes | Private result builder | Rollback coordinator | Rollback model | Rollback and integration tests | Rollback success envelope | Revisions | high | Preserve stable affected-record ordering. |
| `internal/modules/revisions/rollback_test.go` | Selector-union, provider-family, history/projection, and event characterization | Selector test plus test request helpers | Rollback/store verification rows | Full app harness and source-owner fixtures | Self | Rollback route and selector union | Revisions tests | very high | Broad service-backed behavior freeze. |
| `internal/modules/revisions/rollbackcontract/contract.go` | Consumer-owned row/non-row source rollback ports and value objects | Target/request/descriptor/result types and provider interfaces | Source-owner rollback providers and Revisions catalogs | Caller transaction and generic values | Provider and rollback suites | Internal source-owner rollback contract | Revisions contract package | high | Source owners retain validation, inverse application, and persistence. |
| `internal/modules/revisions/row_provider_catalog.go` | Fail-closed record-type to row rollback provider catalog | Registration/catalog types, constructor, lookup | Command service and rollback components | `rollbackcontract` | Row catalog and rollback suites | Row-provider coverage | Revisions | high | Generic lookup is intentional. |
| `internal/modules/revisions/row_provider_catalog_test.go` | Duplicate/missing row-provider catalog evidence | One test and provider fake | Provider verification row | Row provider contract/catalog | Self | Row rollback provider evidence | Revisions tests | low | Test-only. |

All 73 tracked target files are represented above; no file is marked out of
scope.

## 3. Module Boundary Diagnosis

The live target is a real Revisions bounded context. `CommandService` is a
legitimate thin application facade, and the package hides substantial history,
conflict, destructive-mutation, and rollback complexity. The overall target is
also a mixed-responsibility package because it contains transport-adjacent and
persistence-adjacent adapters alongside generic domain/application
coordination. Those adapters are not automatically misplaced: the adopted
owner split permits Revisions-owned HTTP and persistence adapters provided
platform concerns terminate at the adapter and source semantics remain with
their owners.

It is therefore best described as:

- a legitimate thin application/service facade;
- a view/projection orchestration consumer, but not a projection owner;
- a transport-adjacent adapter through `httpapi`;
- a persistence-adjacent adapter for Revisions-owned tables;
- a mutation coordinator for delete, restore, rollback, and revisionless
  conflict outcomes; and
- a mixed-responsibility package with residual misplaced source semantics and
  platform lookup coupling.

It is not supported by repository evidence as a frontend shell/controller,
grid-vendor integration, saved-view owner, tabular-ingest owner, or a general
catch-all for Timeline, Entities, Indicators, Evidence, Links, or other source
modules.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Thin Revisions command facade | `command_service.go`, `application_ports.go` | Revisions | keep | Four transport-neutral methods and assembly adapters | Preserve a small public surface. |
| Change-set and record-revision append | `appender.go` | Revisions | keep | Core 00/02 and Revisions table ownership | Split only the concrete Records dependency. |
| Records envelope lookup for Collaboration intent | `appender.go` | Records port injected by revision assembly | move | Direct `records.NewStore().LoadEnvelopeTx` versus REQ-01-649 | Behavior stays in Revisions; dependency construction moves. |
| History query/materialization | `history_*` files | Revisions generic mechanics | split | Core 02 and current component boundaries | Keep components private; extract only source-specific target semantics. |
| Source-specific history addressability and JSON reference keys | `history_materializer.go`, `history_repository.go` | Respective source owners via immutable contributions | move | Target-kind switches and keys such as `src_record_id`, `source_record_id`, and `indicator_record_id` | Must preserve every current selector and affected-record rule. |
| Delete/restore coordination | `delete_restore_store.go` and consumer contract/catalog | Revisions coordinator plus source-owner providers | keep | REQ-01-650 and complete catalog | Source owners retain preconditions and source persistence. |
| Rollback planning and transaction coordination | `rollback_*` files | Revisions generic mechanics | keep | REQ-02-218 and current decomposition | Existing component split is valid. |
| Source-kind mapping and special rollback handling | Rollback planner/apply/query components | Respective source owners via provider metadata/capabilities | move | Hard-coded first-class and non-row target branches | Closed completeness catalogs remain in Revisions. |
| Projection refresh | `projection_port.go`, mutation coordinators | Projections behind a Revisions consumer port | keep | Injected `ProjectionServices` | Refresh is a consequence, never source truth. |
| Projection-first history snapshot | `projection_port.go:snapshotRecordTx` | Authoritative source snapshot provider | split | Projection load precedes `DeleteRestoreSource.SnapshotTx`; REQ-02-218 forbids projection-derived fallback | Change only after exact snapshot compatibility is characterized. |
| Conflict token, revision window, merge, and `keep_saved` | `conflicts/**` | Revisions conflicts | keep | REQ-03-066 and AC-529 | Generic mechanics; source owners retain current state and mutation. |
| View-schema-to-conflict descriptor adapter | `conflict_field_provider.go` | Application composition/platform adapter | move | Root imports process-global `platform/viewschema`; revision assembly already admits that dependency | Runtime catalog remains immutable. |
| Key-ring environment fallback | `conflicts/key_ring.go` | Revisions config parser or server startup adapter | defer | Explicit env map exists, but nil falls back to `os.LookupEnv` | Compatibility and config-owner evidence are insufficient for a safe move now. |
| Revisions HTTP operations | `httpapi/**` | Revisions HTTP adapter | keep | AC-528/AC-529 and dedicated package | No route or guard-order change. |
| Revisions-table SQL | History, rollback query, conflict reader, portability, and appender files | Revisions persistence adapters | keep | Backend SQL allowlist permits only Revisions history/envelope tables | Do not replace fixed SQL with dynamic relation metadata. |
| Incident Bundle Revisions members | Incident-bundle source/portability/validation files | Revisions source adapter under Incident Bundles orchestration | keep | Source family `revisions`, three row files, six invariants | Not general import/tabular-ingest ownership. |
| Recovery-state contribution | `recovery_state.go` | Revisions contribution to Recovery | keep | Four authoritative Revisions tables | Recovery orchestrates the aggregate catalog. |
| Collaboration WebSocket delivery | Outside target; target appends intents | Collaboration | keep | `IntentAppender` and `record_changed` contract | Revisions owns transactional intent facts, not `/ws/v1/` transport. |
| Frontend history/controller state | `apps/web/src/workbook/timeline/**` | Web application/Timeline workbook | keep | History port, adapter, actions, model, and inspector panel | External consumer only; no frontend move is proposed. |

## 4. Public Contract and Behavior Freeze Map

No public HTTP, WebSocket, OpenAPI, database, generated TypeScript, or UI
selector change is intended. Internal Go interfaces may later change only to
improve dependency direction and source-owner contribution semantics.

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `GET /api/v1/records/{record_id}/history` / `getRecordHistory` | Revisions HTTP/history | Core 01, Revisions OpenAPI source, `httpapi/routes.go` | History unit/OpenAPI/ref/retention, pagination integration, browser history | Preserve strict query, hidden visibility, newest-first order, cursor binding, resource shape, opaque refs, and retained visibility | high | No purge or retention-control routes exist. |
| `DELETE /api/v1/records/{record_id}` / `deleteRecord` | Revisions coordinator; source owners for blockers/state | Core 01 and OpenAPI source | Delete adapter matrix, preconditions, security, integration, browser | Preserve optimistic concurrency, blocker tuples, idempotency, tombstone, history, projection, and remove event | very high | Success remains the common `200` envelope. |
| `POST /api/v1/records/{record_id}/restore` / `restoreRecord` | Revisions coordinator; source owners for state | Core 01 and OpenAPI source | Restore preconditions, lock, security, integration, browser | Preserve reviewer/admin gate, nowait lock precedence, replay, projection, and invalidate event | very high | Protected set is the target record. |
| `POST /api/v1/records/{record_id}/rollback` / `rollbackRecord` | Revisions generic coordinator; source-owner providers | Core 01/02 and OpenAPI source | Selector union, provider families, locks, failure injection, integration, browser | Preserve selector union, canonical locks, fresh-state revalidation, reverse apply, atomicity, and result order | very high | Source mutation semantics must remain behind providers. |
| Authentication/CSRF/visibility/role/body precedence | Auth and Revisions HTTP adapter | Core 01/03/04, AC-528, `http_security_precedence_test.go` | Dedicated precedence integration test and route suites | Retain no-write assertions for every rejected stage | very high | Unauthorized malformed requests reveal no later-stage detail. |
| Change sets and mutation entries | Revisions | Core 00/02, `appender.go`, owned tables | Source-owner mutation tests, Revisions integration and portability | Characterize append count/order, attribution, target values, and no partial writes before moving dependencies | high | Immutable append-only substrate. |
| Record revisions and row snapshots | Revisions history; source owners for authoritative current state | Core 00/02, Appender and snapshot helper | History, rollback, conflict-window, portability, source-owner suites | Compare projection-derived and authoritative-source snapshots for every admitted record family before RS-02 | very high | Current projection-first behavior is an owner mismatch, not a compatibility decision. |
| Stable history items and rollback actions | Revisions generic mechanics plus source-owner legality | Core 02 and history materializer/evaluator | History component/ref/retention, indicator-child, rollback suites | Freeze every `history_item_ref`, `history_entry_ref`, action order, target association, and ineligibility result | very high | Source-kind extraction must not alter selectors. |
| Source-owner row and non-row inverse behavior | Artifacts, Assessments, Entities, Evidence, Indicators, Links, Parties, Tasks/Decisions, Timeline | Provider contributions and consumer-owned contracts | Provider catalog rows and source-owner rollback tests | Add/retain exact provider-capability tests before deleting central branches | very high | Revisions coordinates but does not define source vocabulary. |
| Projection rebuild and refresh consequences | Projections implementation; Revisions coordinates calls | Projection consumer port and Core 01/02 | Delete/restore/rollback integration, source-owner query tests | Assert rebuild ordering, rollback on failure, and no projection-as-history truth | high | Projection tables remain disposable derived state. |
| `record_changed` WebSocket consequence | Collaboration delivery; Revisions/source transactions create intents | Core 01/03/04, WS contract, Appender | Revisions integration, Collaboration E2E, source-owner mutation suites | Preserve one canonical post-commit event per affected live record, array order, replay identity, and no event on failure/import | very high | Revisions owns no WebSocket route. |
| Same-field conflict and v3 token | Revisions generic conflicts; source owners for fields/state/mutation | REQ-03-066, AC-526/AC-529, Revisions key-ring contract | Token, resolver, revision-window, text-merge, workbook/source-owner tests | Preserve opaque `cft3`, bindings, TTL/skew, uniform error, immutable descriptors, and revisionless `keep_saved` | very high | No v2 or authentication-master fallback. |
| View-schema mapping and `affected_views` | Source owners declare routes; Revisions immutable catalog; Projections owns descriptors | Record-view and conflict-field catalogs | Revision-assembly, resolver, changed-fields, integration tests | Freeze record variant routing, writable/conflict classes, and affected-view identities while moving startup adapter | high | Saved-view lifecycle is not involved. |
| Idempotency and destructive locks | Platform idempotency/Records locks behind Revisions ports | Core 01 and coordinator code | Lock, replay, stale-state, failure-injection, integration tests | Preserve exact replay/divergent-replay outcomes and lock acquisition order | very high | No nested transaction may be introduced. |
| Revisions Incident Bundle source family | Revisions source adapter; Incident Bundles orchestrates | Source catalog and three row schemas | Allocator, deterministic round trip, invariant priority, atomic failure, incident-bundle route integration | Preserve versions 1/2, exact rows, order, invariant priority, attribution, sequence repair, and no live events | very high | Generated projections are downstream only. |
| Recovery contribution | Revisions tables; Recovery aggregate | `recovery_state.go` and recovery fixtures/catalog | Recovery catalog and broader recovery checks | Preserve classification of four authoritative history tables | medium | No schema or migration change is planned. |
| Frontend Timeline history port and inspector | Web application/Timeline workbook | Timeline port, generated-operation adapter, history model/actions/panel, UI contracts | Frontend units; `history.spec.ts`, inspector, visual, and accessibility scenarios | Run browser rows when wire behavior, selectors, refresh, or event consequences are touched | high | No direct Revisions frontend package or grid-vendor dependency exists. |
| Generated OpenAPI/protocol/UI surfaces | Authored owner inputs and generators | Revisions OpenAPI source, generated protocol-ts, ui-contract selectors | OpenAPI compatibility, frontend units/browser | No generated change for structural slices; regenerate only from an authorized owner-input change | high | Never hand-edit generated roots. |
| Harness/test accounting | Verification owners and authored test-family catalog | 22 verification entries and 43 active `module.revisions` rows; 28 service-backed per owner explanation | Harness contract and owner slices | Add authored rows only for genuinely new evidence, then regenerate topology through Make | medium | Routing is evidence accounting, not runtime architecture. |

### 4.1 Refactor-wide normative requirements

| ID | Requirement | Owner basis | Planned evidence |
| --- | --- | --- | --- |
| RMR-REQ-001 | Every authorized slice MUST preserve every observable consequence in the freeze map unless that slice names a separately authorized owner change and compatibility plan. | Core 00-04 | Pre-change characterization plus focused and broad post-change evidence |
| RMR-REQ-002 | No slice in this tracker authorizes a public HTTP, WebSocket, OpenAPI, database, TypeScript, UI-selector, portability, configuration, or generated interface change. | Core 01-04 | Contract drift, browser, portability, and generated-artifact checks when applicable |
| RMR-REQ-003 | Revisions coordinators MUST retain caller-owned transaction semantics. A provider MUST NOT begin, commit, roll back, or replace the caller transaction. | Core 01/02 | Transactional failure injection and provider unit tests |
| RMR-REQ-004 | Authentication, CSRF, visibility, role, body validation, idempotency, lock, and fresh-state checks MUST retain their adopted precedence. An unauthorized request MUST NOT invoke target-semantics providers. | Core 01/03/04 | Security-precedence integration evidence and negative provider-call assertions |
| RMR-REQ-005 | History MUST remain immutable and append-only. Delete, restore, rollback, and source mutations MUST retain their current change-set, mutation, revision, and attribution facts. | Core 00/02 | History, rollback, source-owner, and portability evidence |
| RMR-REQ-006 | Every existing history item reference, history entry reference, target association, logical grouping rule, action order, and rollback selector MUST remain stable. | Core 02 | Exact old/new history and rollback comparisons |
| RMR-REQ-007 | Projection refresh and `record_changed` consequences MUST remain transactionally ordered, failure-atomic, and suppressed for imports as currently owned. | Core 01/03/04 | Integration, Collaboration, and failure-injection evidence |
| RMR-REQ-008 | Incident Bundle versions 1 and 2 MUST retain exact Revisions row interpretation, ordering, invariants, attribution, and no-publication behavior. | Core 01/02 and Incident Bundle owner | Deterministic round-trip and atomic-import evidence |
| RMR-REQ-009 | Generated files MUST NOT be edited by hand. Authored owner or verification inputs MAY change only when an authorized slice requires them; omission means no authored or generated change. | Repository generated-artifact policy | Make-owned drift and policy checks when applicable |
| RMR-REQ-010 | Verification maps and phase rows MUST be treated only as evidence routing. They MUST NOT be used as runtime architecture or proof of requirement completeness. | Testing Harness owner and repository procedure | Review of authored mapping changes |

### 4.2 Authoritative snapshot contract

| ID | Requirement |
| --- | --- |
| RMR-REQ-100 | For every new revision snapshot, Revisions MUST use only the applicable source owner's authoritative snapshot returned from the current source state in the caller-owned transaction. |
| RMR-REQ-101 | Revisions MUST NOT query a projection row as a snapshot source, fallback, augmentation, or compatibility authority. A missing source snapshot MUST fail through the existing typed error path; it MUST NOT fall back to a projection. |
| RMR-REQ-102 | An authoritative snapshot MUST contain envelope identity, record type, row version, deletion state, authoritative source-row state required for row restoration, and source-owned scalar or structured extension values. |
| RMR-REQ-103 | An authoritative snapshot MUST exclude projection-only labels, counts, chips, joined summaries, preview handles, sort/group/layout state, binary bytes or object material, authorization state, UI state, and external enrichment. |
| RMR-REQ-104 | Separately revisioned collection targets, including links, tags, mentions, indicator observations, indicator state intervals, and evidence associations, MUST NOT be folded into a row snapshot. Their own target semantics remain authoritative. |
| RMR-REQ-105 | Active and deleted records MUST use the same snapshot shape domain. The snapshot MUST express the current row version and the complete deletion tuple, including null deletion members when active. |
| RMR-REQ-106 | Structural comparison MUST distinguish absent, explicit `null`, empty, and non-empty members. It MUST compare JSON member names, types, values, array ordering, nested values, and custom attributes. |
| RMR-REQ-107 | JSON object key order and insignificant serialized whitespace MAY differ. That is the only default representation tolerance. Duplicate object members MUST be rejected before persistence. |
| RMR-REQ-108 | The old projection-first reader and candidate authoritative reader MUST observe the same database state through the same supplied transaction for every parity comparison. |
| RMR-REQ-109 | RS-02 MUST stop at the first non-representation difference and classify it using Section 4.5. It MUST NOT add a dual reader, shadow writer, feature flag, compatibility fallback, or projection rescue path. |
| RMR-REQ-110 | Exact parity authorizes a single-cutover replacement of the old source-selection path. Removal of the old path and introduction of the new path MUST occur in the same RS-02 slice. |
| RMR-REQ-111 | If incompatible historical snapshot interpretations must coexist, a separately authorized owner change MUST introduce an explicit persisted discriminator named `snapshot_schema_id`. Key-presence inference is prohibited. |
| RMR-REQ-112 | A `snapshot_schema_id` change MUST also version any affected Incident Bundle row or payload interpretation. Existing bundle versions MUST NOT be silently reinterpreted. |

### 4.3 Closed snapshot characterization matrix

The record-family column is closed over the ten currently admitted record
families. RS-00 MUST stop and expand the matrix if live evidence identifies an
additional admitted family or variant.

| Record family | Required variants and boundary cases | Required lifecycle cases | Association exclusions |
| --- | --- | --- | --- |
| `timeline_event` | Every currently admitted Timeline event source shape | Active, before delete, after delete, before restore, after restore | Separately revisioned links, tags, and mentions |
| `host` | Stub and canonical | Active, deleted, delete transition, restore transition | Aliases, mentions, preserved identifiers, links, and tags |
| `identity` | Stub and canonical | Active, deleted, delete transition, restore transition | Aliases, mentions, preserved identifiers, links, and tags |
| `party` | Current Party source shape | Active, deleted, delete transition, restore transition | Links, tags, mentions, and other separately revisioned relations |
| `indicator` | Indicator row with observations absent/present and intervals absent/present | Active, deleted, delete transition, restore transition | Observations and state intervals |
| `artifact` | `comm_log`, `finding`, `forensic_keyword`, `handoff`, `investigative_query`, `lesson`, `note`, and `status_review` | Active, deleted, delete transition, restore transition | Links, tags, mentions, and binary/object material |
| `task_request` | Current Task Request source shape | Active, deleted, delete transition, restore transition | Links, tags, mentions, and separately revisioned relations |
| `decision` | Current Decision source shape | Active, deleted, delete transition, restore transition | Links, tags, mentions, and separately revisioned relations |
| `evidence` | Evidence object absent and present | Active, deleted, delete transition, restore transition | Evidence associations and binary/object material |
| `assessment` | Host subject and identity subject | Active, deleted, delete transition, restore transition | Subject-owned collections and separately revisioned relations |

Every row above MUST cover the following cross-product dimensions to the
extent that the source family admits them. “Not admitted” MUST be proved by the
source-owner contract or test; it MUST NOT be assumed.

| Dimension | Required cases |
| --- | --- |
| Presence | Required, nullable, absent, empty string, empty object, empty array, and non-empty values |
| Value domain | Unicode, multiline text, timestamps, booleans, numeric boundaries, UUIDs, nested objects, ordered arrays, and nested `custom_attrs` |
| Attribution | Local actor and imported actor attribution where admitted |
| Operation | Create, scalar patch, structured patch, soft delete, restore, and row rollback where admitted |
| Adjacent mutation | Each separately revisioned non-row target that can occur next to the row mutation |
| Failure | Missing row, malformed persisted JSON if representable, stale state, provider error, projection failure, history failure, and transaction rollback |
| Portability | Live write, export, version 1 import, version 2 import, sequence repair, and post-import history read |

### 4.4 Exact compatibility layers

| Layer | Equality requirement |
| --- | --- |
| A: structural snapshot | Exact member presence, nullability, JSON type, value, array order, nested value, and custom-attribute equality after only the representation tolerance in RMR-REQ-107 |
| B: history materialization | Exact item count/order, operation and diff units, fields, targets, summaries, associations, identifiers, reversibility, ordered actions, references, and pagination behavior |
| C: rollback | Exact selector resolution, candidate/companion/protected sets, lock order, post-lock revalidation, inverse order, restored values, affected-record order, versions, changed fields, projection refresh, event consequences, and error mapping |
| D: external consequence | Exact HTTP history and rollback envelopes, frontend rendering/actions, WebSocket consequences, delete/restore view effects, portability, conflict revision windows, and no-publication-on-failure/import behavior |

### 4.5 Difference classification and mandatory action

| Difference class | Definition | Mandatory action |
| --- | --- | --- |
| `representation_only` | Only JSON object key order or insignificant serialized whitespace differs | Record the normalization and continue. This class alone does not block parity. |
| `projection_only_member` | The old snapshot contains a value owned only by a projection | Stop RS-02. Perform ownership and observability review and obtain behavior-change authorization. |
| `source_member_missing_from_projection` | Authoritative state is absent from the old snapshot | Stop RS-02. Repair or version the contract before cutover. |
| `normalization_mismatch` | Old and candidate paths normalize the same source value differently | Stop RS-02. Assign the normalization owner and repair the defective path. |
| `association_or_selector_change` | History grouping, association, action ordering, or opaque selector differs | Stop RS-02 and resolve the difference under RB-002. |
| `rollback_change` | Candidate snapshots change eligibility, ordering, restored state, results, or errors | Stop RS-02. Obtain owner review and any required Core change. |
| `public_or_portability_change` | HTTP, WebSocket, frontend, conflict, generated, or Incident Bundle behavior differs | Stop RS-02. Obtain explicit owner and compatibility authorization. |

### 4.6 RB-001 outcome gate

| Outcome | Required evidence | Result |
| --- | --- | --- |
| A: exact semantic parity | All matrix cells and all four compatibility layers pass with only recorded `representation_only` differences | Mark RB-001 `CLOSED`; authorize a single-cutover RS-02; remove the projection-first path without compatibility machinery. |
| B: projection-only behavior is observable | One or more non-representation differences are owned or externally observable | Keep RB-001 open; stop RS-02; require controlling-owner amendment plus an explicit migration/versioning and compatibility plan. |
| C: one path is defective | Evidence identifies a source-owner or projection adapter defect | Keep RB-001 open; repair the defective owner path in a separately authorized slice; rerun the full parity matrix without fallback. |

### 4.7 Target-semantics capability contract

The following shape is implementation-facing guidance for RS-03. Exact Go
package placement belongs to the later authorized implementation, but that
implementation MUST preserve these responsibilities and prohibitions.

```go
type TargetSemanticsContribution struct {
	TargetKind string
	History    HistoryTargetSemantics
	Rollback   RollbackTargetSemantics
}

type HistoryTargetSemantics interface {
	DescribeMutation(StoredMutation) (HistoryTargetDescription, error)
}

type RollbackTargetSemantics interface {
	PlanRollbackTx(context.Context, Transaction, RollbackTargetRequest) (RollbackTargetPlan, error)
	ApplyInverseTx(context.Context, Transaction, RollbackTargetPlan) (RollbackApplyResult, error)
}
```

`Transaction` in this shape means the caller-supplied transaction capability;
it does not authorize a provider to create or own a transaction. Existing
`DeleteRestoreSource`, `rollbackcontract.RowSourceProvider`, and
`rollbackcontract.NonRowTargetProvider` signatures remain unchanged by
default. RS-03 MAY adapt those providers behind the new facets. It MUST NOT
widen them merely to move central branching.

| ID | Requirement |
| --- | --- |
| RMR-REQ-200 | Revisions MUST own one immutable consumer-facing target-semantics catalog, keyed by exact target kind and composed from source-owner implementations before serving. |
| RMR-REQ-201 | Every persistable target kind MUST resolve to exactly one admitted history facet. A rollback facet MUST also be present unless the contribution declares the target permanently non-reversible and RS-00 proves that declaration matches current behavior. |
| RMR-REQ-202 | A history facet MUST be pure. It MUST NOT query a database, mutate state, use a transaction, or inspect platform, transport, projection, or UI state. |
| RMR-REQ-203 | A history description MUST return the exact target kind and target ID, primary associated record, additional associated records, owner-scoped logical-item identity, closed addressability, selector inputs, companion references, source-owned diff classification, and display-neutral semantic metadata. |
| RMR-REQ-204 | A history description MUST NOT return labels, components, routes, `history_entry_ref`, SQL, relation names, view or projection rows, callback maps, untyped extension payloads, or source-package runtime strings. |
| RMR-REQ-205 | Revisions MUST continue to generate, encode, persist, resolve, and lifetime-stabilize opaque history and rollback selectors. A source owner supplies semantic inputs only. |
| RMR-REQ-206 | A rollback plan MUST return reversibility or a typed ineligibility reason, affected and protected record IDs, companion targets, deterministic inverse steps and order, current and expected state descriptions, changed fields, and source-owned inverse input. |
| RMR-REQ-207 | A rollback provider MUST use only the caller transaction and source-owned fixed SQL. It MAY revalidate source state under locks and apply only its source-owned inverse. |
| RMR-REQ-208 | A provider MUST NOT own authorization, HTTP decoding, idempotency, history append, projection refresh, Collaboration intent publication, network or object-store calls, dynamic SQL, relation names supplied at runtime, or transaction completion. |
| RMR-REQ-209 | Revisions MUST retain selector lookup, history SQL, canonical ordering, lock aggregation, transaction coordination, idempotency, history publication, projection refresh, Collaboration consequences, result ordering, and public error mapping. |
| RMR-REQ-210 | Catalog construction MUST reject an empty catalog, a missing required kind, a duplicate kind, an unknown kind, a typed-nil facet, an incomplete contribution, non-deterministic output, cross-owner behavior, and relation-bearing or dynamic-SQL metadata before serving. |
| RMR-REQ-211 | An unknown target kind MUST yield no association, addressability, companion, rollback, or action behavior. If a producer can persist that kind, startup admission MUST fail until a complete contribution exists. |
| RMR-REQ-212 | The default for a future target kind is fail-closed. It is neither individually addressable nor reversible and MUST NOT be persisted until its owner contribution and tests are admitted. |
| RMR-REQ-213 | RS-03 MUST NOT replace current switches with another central switch, a target-to-relation map, SQL fragments, reflection, a global initialization registry, a generic hook bus, a chain of responsibility, a callback map, or a legacy fallback. |
| RMR-REQ-214 | Revisions MUST canonicalize, copy, deduplicate, and order all provider-supplied record and target sets before locking or publication. It MUST revalidate after locks and MUST fail atomically. |
| RMR-REQ-215 | Provider failures MUST map to existing safe typed errors. Logs MUST NOT contain authored content, raw snapshot values, conflict material, or secrets. |

The returned value fields have the following closed ownership. A field not
listed here is omitted by default and requires a later authorized design
change.

| Value | Required members | Owner and constraints |
| --- | --- | --- |
| `HistoryTargetDescription` | Exact target kind/ID; primary record; ordered additional records; logical-item key; addressability enum; selector inputs; ordered companions; diff classification; display-neutral metadata | Source owner supplies semantic facts; Revisions validates and canonicalizes all record identifiers and selector inputs. |
| Addressability enum | Exactly `not_individually_addressable` or `single_history_entry` unless an adopted owner later adds a value | Source owner selects the current behavior; unknown or omitted values fail admission. |
| `RollbackTargetPlan` | Eligibility or typed ineligibility; ordered affected/protected records; ordered companions; deterministic inverse steps; current/expected state; changed fields; inverse input | Source owner computes source semantics; Revisions canonicalizes locks and owns orchestration. |
| `RollbackApplyResult` | Applied target facts, ordered affected records, changed fields, and source versions needed by the existing public result | Source owner reports only applied source facts; Revisions owns history, projections, events, and public encoding. |

### 4.8 Closed current target-kind matrix

The target-kind column is closed over exact tokens found in current Revisions
producers and consumers. RS-00 MUST prove every remaining cell from live
mutations, providers, and tests. A `TODO` cell is an evidence blocker, not an
invitation to choose new behavior.

| Target kind | Semantic owner | Target ID shape | Primary history record | Additional associated records | Logical item and addressability | Selector inputs | Provider evidence |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `assessment` | Assessments | UUID record ID | Same assessment record | TODO: RS-00 evidence required | TODO: RS-00 evidence required | Exact persisted target kind/ID and current history fields | Assessment row provider and history/rollback tests |
| `entity_alias` | Entities | UUID-prefixed or current content-derived logical ID | Owning Host or Identity | TODO: RS-00 evidence required | Current alias grouping; not individually addressable unless existing evidence proves otherwise | Exact current alias identity inputs | Entity non-row provider and alias mutation tests |
| `entity_mention` | Entities | UUID mention ID | Current `source_record_id` association | Mentioned entity when currently associated | Current mention grouping and action posture | Exact mention ID and association inputs | Entity non-row provider, history JSON association, and rollback tests |
| `entity_preserved_identifier` | Entities | Current content-derived logical ID | Owning Host or Identity | TODO: RS-00 evidence required | Current preserved-identifier grouping; create-only behavior retained | Exact current logical identity inputs | Entity non-row provider and merge/repoint tests |
| `evidence` | Evidence | UUID record ID | Same Evidence record | TODO: RS-00 evidence required | TODO: RS-00 evidence required | Exact persisted target kind/ID and current history fields | Evidence row provider and history/rollback tests |
| `host` | Entities | UUID record ID | Same Host record | TODO: RS-00 evidence required | TODO: RS-00 evidence required | Exact persisted target kind/ID and current history fields | Host row provider and history/rollback tests |
| `identity` | Entities | UUID record ID | Same Identity record | TODO: RS-00 evidence required | TODO: RS-00 evidence required | Exact persisted target kind/ID and current history fields | Identity row provider and history/rollback tests |
| `indicator` | Indicators | UUID record ID | Same Indicator record | TODO: RS-00 evidence required | TODO: RS-00 evidence required | Exact persisted target kind/ID and current history fields | Indicator row provider and history/rollback tests |
| `indicator_observation` | Indicators | UUID observation ID | Current `source_record_id` association | Resolved Indicator when currently associated | Current observation grouping and addressability | Exact observation and association inputs | Indicator non-row provider, association SQL, and rollback tests |
| `indicator_state_interval` | Indicators | UUID interval ID | Current `indicator_record_id` association | TODO: RS-00 evidence required | Current interval grouping and addressability | Exact interval and indicator inputs | Indicator non-row provider, association SQL, and rollback tests |
| `record` | Applicable source owner selected through the admitted record-type catalog; Revisions owns generic dispatch | UUID record ID | Same record | TODO: RS-00 evidence required | Generic row history rule; exact action posture requires RS-00 evidence | Exact persisted token/ID plus admitted record type | Artifacts, Parties, Task Requests, Decisions, and any other live producer; exact producer set requires RS-00 confirmation |
| `record_link` | Links | UUID link ID | Current source-record association | Current destination-record association | Current link grouping and addressability | Exact link ID, source, destination, and relation inputs | Link non-row provider, history association SQL, and rollback tests |
| `record_tag` | Links | Current raw UUID or `record_tag:<record_id>:<tag_id>` composite | Current `record_id` association | TODO: RS-00 evidence required | Both live identity forms MUST retain their current grouping and addressability | Exact current record/tag identity inputs | Tag non-row provider, merge producer, history association SQL, and rollback tests |
| `timeline_record` | Timeline | UUID record ID | Same Timeline Event record | TODO: RS-00 evidence required | TODO: RS-00 evidence required | Exact persisted target kind/ID and current history fields | Timeline row provider and history/rollback tests |

| Target kind | Companion rule | Protected-set rule | Current-state rule | Inverse order | Changed-field rule | Affected-record order | Deleted-target behavior | Portability behavior |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `assessment` | TODO: RS-00 evidence required | TODO: RS-00 evidence required | Source-owned assessment row plus envelope | TODO: RS-00 evidence required | Preserve current row diff fields | Preserve current order | Preserve current deleted-row eligibility | Exact persisted mutation and revision round trip |
| `entity_alias` | TODO: RS-00 evidence required | TODO: RS-00 evidence required | Source-owned alias state | TODO: RS-00 evidence required | Preserve current alias fields | Preserve current order | Preserve current eligibility | Exact persisted mutation round trip |
| `entity_mention` | Current companion `record_link` behavior MUST be preserved | TODO: RS-00 evidence required | Source-owned mention and companion state | Preserve current reverse apply order | Preserve current mention/link fields | Preserve current order | Preserve current eligibility | Exact mention and companion round trip |
| `entity_preserved_identifier` | TODO: RS-00 evidence required | TODO: RS-00 evidence required | Source-owned preserved-identifier state | Preserve current create-only ineligibility or inverse rule | Preserve current fields | Preserve current order | Preserve current eligibility | Exact persisted mutation round trip |
| `evidence` | TODO: RS-00 evidence required | TODO: RS-00 evidence required | Source-owned Evidence row plus envelope | TODO: RS-00 evidence required | Preserve current row diff fields | Preserve current order | Preserve current deleted-row eligibility | Exact mutation, revision, and evidence-object round trip |
| `host` | TODO: RS-00 evidence required | TODO: RS-00 evidence required | Source-owned Host row plus envelope | TODO: RS-00 evidence required | Preserve current row diff fields | Preserve current order | Preserve current deleted-row eligibility | Exact mutation and revision round trip |
| `identity` | TODO: RS-00 evidence required | TODO: RS-00 evidence required | Source-owned Identity row plus envelope | TODO: RS-00 evidence required | Preserve current row diff fields | Preserve current order | Preserve current deleted-row eligibility | Exact mutation and revision round trip |
| `indicator` | Observation and interval targets remain separate; exact companions require evidence | TODO: RS-00 evidence required | Source-owned Indicator row plus envelope | TODO: RS-00 evidence required | Preserve current row diff fields | Preserve current order | Preserve current deleted-row eligibility | Exact mutation and revision round trip |
| `indicator_observation` | Preserve current source/resolved-record companions | TODO: RS-00 evidence required | Source-owned observation state | Preserve current inverse order | Preserve current observation fields | Preserve current order | Preserve current eligibility | Exact observation round trip |
| `indicator_state_interval` | Preserve current Indicator companion | TODO: RS-00 evidence required | Source-owned interval state | Preserve current inverse order | Preserve current interval fields | Preserve current order | Preserve current eligibility | Exact interval round trip |
| `record` | Selected row provider supplies source companions, if any | Target record remains protected where current operation requires it | Source-owned row snapshot plus envelope | Preserve current row-provider inverse order | Preserve exact changed-field calculation | Preserve current result order | Preserve current deleted-row eligibility | Exact mutation and revision round trip for every producing family |
| `record_link` | Preserve source and destination record companions | TODO: RS-00 evidence required | Source-owned fixed-SQL link state | Preserve current inverse order | Preserve current link fields | Preserve current order | Preserve current endpoint-delete behavior | Exact link round trip |
| `record_tag` | Preserve current record/tag companions for both identity forms | TODO: RS-00 evidence required | Source-owned fixed-SQL tag membership state | Preserve current inverse order | Preserve current tag fields | Preserve current order | Preserve current record-delete behavior | Exact tag round trip without normalizing historical IDs |
| `timeline_record` | TODO: RS-00 evidence required | TODO: RS-00 evidence required | Source-owned Timeline row plus envelope | TODO: RS-00 evidence required | Preserve current row diff fields | Preserve current order | Preserve current deleted-row eligibility | Exact mutation and revision round trip |

### 4.9 RB-002 closure evidence

RB-002 MUST remain open until tests prove all of the following:

| ID | Required proof |
| --- | --- |
| RMR-AC-200 | Every persistable exact target kind has exactly one admitted history contribution and the closed token set matches all live producers. |
| RMR-AC-201 | Every target preserves current primary/additional association, logical grouping, addressability, selector inputs, item order, action order, and opaque selector results. |
| RMR-AC-202 | Every reversible target preserves candidate, companion, protected, and affected sets; canonical lock and inverse order; current-state revalidation; changed fields; versions; result order; and errors. |
| RMR-AC-203 | Empty, missing, duplicate, unknown, typed-nil, incomplete, non-deterministic, cross-owner, and relation-bearing contributions fail before serving. |
| RMR-AC-204 | Unknown targets and targets with absent rollback facets fail closed and publish no history action or partial mutation. |
| RMR-AC-205 | Failure injection at provider planning, locking, revalidation, inverse application, history append, projection refresh, event intent, and idempotency persistence leaves no partial state. |
| RMR-AC-206 | Unauthorized requests do not invoke a target provider; provider errors expose no authored values or unsafe implementation details. |
| RMR-AC-207 | Static boundary evidence finds no replacement switch, source JSON key, relation mapping, SQL fragment, reflection registry, hook bus, callback map, legacy fallback, or prohibited source/platform import in generic Revisions coordination. |
| RMR-AC-208 | Source-owner providers use fixed source-owned SQL, and Revisions continues to use fixed Revisions-owned SQL only. |
| RMR-AC-209 | HTTP, WebSocket, history, rollback, projection, frontend, portability, conflict, generated, and harness behavior remains unchanged. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Appender bypasses the injected Records boundary for an envelope lookup | `appender.go` calls `records.NewStore().LoadEnvelopeTx`; REQ-01-649 requires Revisions lookups through a Records-owned port | Concrete dependency can evade composition tests and broaden record authority | `must_fix` | Records port injected by `internal/app/revisionassembly` | Add a narrow reader dependency to `Appender`, migrate its sole production construction, and characterize event payloads. |
| History snapshot helper prefers projections to source snapshots | `projection_port.go:snapshotRecordTx`; REQ-02-218 rejects projection truth and projection-derived fallback | Revision JSON and rollback behavior can be shaped by disposable derived rows | `must_fix` | Source-owner `DeleteRestoreSource.SnapshotTx` | Compare shapes for every admitted record type; switch only when compatible, otherwise require explicit behavior-change authorization. |
| History addressability embeds source target kinds and JSON field names | `history_materializer.go` switch and JSON-key probes; related repository matching | New source families require central edits and can drift from owner schemas | `must_fix` | Source owners through immutable Revisions contribution metadata | Define owner-supplied association/addressability capability and freeze every existing selector before migration. |
| Rollback components embed source-kind families and special cases | `rollback_planner.go`, `rollback_apply.go`, and `rollback_query_repository.go` | Central switches can admit invalid targets or miss new owner semantics | `must_fix` | Row/non-row source providers plus immutable catalogs | Extend provider capabilities for classification, companion/association rules, and dispatch without exposing relation metadata. |
| Revisions root owns a process-global view-schema lookup adapter | `conflict_field_provider.go` imports `internal/platform/viewschema`; AC-529 rejects global-registry dependencies | Startup composition remains physically coupled to global state | `should_fix` | `internal/app/revisionassembly` platform adapter | Move the lookup adapter to composition and retain copied immutable descriptors at runtime. |
| Key-ring parsing optionally reads process environment | `conflicts/key_ring.go` falls back to `os.LookupEnv` when the supplied map is nil | Hidden ambient runtime input complicates deterministic configuration tests | `defer` | Revisions config parser or server startup adapter | Characterize nil-map compatibility and confirm owner placement before planning removal. |
| Revisions reads and writes only owner history tables | Backend SQL allowlist and inspected repositories | Low if fixed SQL and allowlist remain exact | `intentional/no_action` | Revisions persistence adapters | Preserve the allowlist and reject dynamic relation metadata. |
| Dedicated HTTP package imports platform auth/HTTP/pagination and concrete visibility adapters | `httpapi/**`, application boundary test, AC-528/AC-529 | Low while imports terminate at the adapter | `intentional/no_action` | Revisions HTTP adapter | Preserve four-operation mapping and guard precedence. |
| Collaboration intent creation remains transaction-bound and delivery-neutral | `Appender` uses injected `IntentAppender`; runtime delivery is outside target | Low if event payload and transaction ordering stay frozen | `intentional/no_action` | Revisions facts plus Collaboration delivery | Preserve one intent per affected live revision and import suppression. |
| Provider completeness uses closed record/target ownership maps | `provider_contributions.go` and catalog tests | Low; required fail-closed admission can be mistaken for source semantics | `intentional/no_action` | Revisions aggregate catalog | Keep completeness validation; move behavioral branches, not admission accounting. |
| Frontend consumes generated HTTP operations and UI selectors outside the target | Timeline history adapter/model/actions/panel and E2E rows | Low for backend-only structural changes; high if wire/event behavior drifts | `intentional/no_action` | Web application/Timeline workbook | Do not move frontend or grid code; retain browser characterization. |
| Archived tracker claims the earlier decomposition is complete while live residual seams remain | Archived tracker versus current `appender.go`, `projection_port.go`, and source-kind switches | Repeating completed work or trusting stale completion would produce an inaccurate plan | `intentional/no_action` | Current tracker | Preserve completed HTTP/conflict/history/rollback splits and track only live residual findings. |

No generated file is suspected of manual drift from this inspection. Any later
authored contract or harness change must precede Make-owned regeneration; the
generated output must never be edited directly.

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01, WF-02 | Fix target, authority, baseline, allowed write, and historical-evidence posture | This tracker only | `git status --short`; source review | Scope, baseline, and authorization boundary recorded. |
| WF-01 | Target inventory | chain | WF-00 | WF-03, WF-04 | Inventory every target file, caller, dependency, test, and contract surface | All 73 target files; assembly and consumer references | Inventory count and exact path comparison | Every target path has one Section 2 row. |
| WF-02 | Contract-owner mapping | parallel | WF-00 | WF-03, WF-04, WF-05 | Map adopted owners to HTTP, history, rollback, conflict, projection, portability, event, frontend, and generated contracts | Core owners, domain navigation, authored contracts, assembly | Owner-document review; `make explain-test-owner OWNER=module.revisions` | No contract is assigned from a filename or historical phase label. |
| WF-03 | Characterization test gap analysis | parallel | WF-01, WF-02 | WF-05, WF-06 | Define the RS-00 snapshot and target-semantics matrices without claiming unexecuted parity | Target tests, source-owner tests, frontend units/E2E, verification/test-family catalogs | `make task-guide ROLE=module-author OWNER=module.revisions` | Each contract risk has existing evidence, missing matrix cells, and a binary comparison rule. |
| WF-04 | Boundary/coupling scan | parallel | WF-01, WF-02 | WF-05, WF-06 | Classify Records, projection, source-kind, platform, SQL, Collaboration, frontend, and generated coupling | Target production files, revision assembly, backend boundary policy | `make explain-target TARGET=backend-module-boundary-check DETAIL=summary` | Every finding has classification, owner, and planning action. |
| WF-05 | Facade and ownership redesign plan | chain | WF-03, WF-04 | WF-06 | Preserve the thin facade while closing the Records, authoritative-snapshot, target-semantics, and composition designs | Appender, application ports, provider contracts/catalogs, projection helper, revision assembly | Design review against REQ-01-649, REQ-02-218, AC-529, and RMR-REQ-100 through RMR-REQ-215 | Interface responsibilities and prohibitions are explicit; remaining uncertainty is evidence, not design choice. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07, WF-08 | Order the smallest independently reversible behavior-preserving changes and their separate authorization gates | Section 7 packages | Per-slice focused commands | Every slice has dependency, stop rule, rollback, binary exit criteria, and one-slice-per-task posture. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-06 | WF-08 | Define when authored verification/boundary inputs and generated topology may change | Verification owner, test family, backend boundary policy, generated outputs | Harness/generation commands in Section 8 | No evidence row or generated output changes without an implementation need. |
| WF-08 | Validation and final handoff | chain | WF-01, WF-02, WF-03, WF-04, WF-05, WF-06, WF-07 | none | Close focused validation, broader checks, blocker state, authorization scope, and handoff | Implementation diff and this tracker | Focused owner slices, boundaries, applicable browser/drift checks, finalization | Another agent can continue without rediscovery and without inferring authorization. |

## 7. Proposed Refactor Slice Plan

All slices below are future implementation work. None is authorized by this
planning task. They preserve behavior unless a row explicitly says
`requires later authorization`.

### 7.1 Authorization 1: RS-00 characterization only

An RS-00 authorization MUST name the repository baseline and MAY permit only
characterization tests, fixtures, reusable test support, necessary authored
verification inputs, and tracker evidence. It MUST require these outputs:

1. the complete per-record-family snapshot parity matrix in Sections 4.3 and
   4.4;
2. the exact target-kind and rollback-semantics matrix in Section 4.8;
3. old/new comparisons made against the same state and transaction;
4. exact Make run roots and result summaries; and
5. a stop report for every non-representation difference or incomplete target
   contribution.

RS-00 MUST NOT change production code, public contracts, database schemas,
migrations, generated files, package configuration, runtime configuration, or
runtime behavior. It MUST NOT claim RB-001 or RB-002 closed unless every
applicable binary criterion passes.

### 7.2 Authorization 2: named production slice

A production authorization MUST state all fields below. An omitted field means
the authorization is incomplete and production work MUST NOT start.

| Required field | Normative content |
| --- | --- |
| Repository baseline | Exact commit and permitted pre-existing worktree changes |
| Permitted slice | Exactly one of RS-01 through RS-05 by default; RS-02 and RS-03 MUST always be separate tasks |
| Permitted paths | Closed authored-file set plus expressly permitted tests; generated paths MUST NOT be named as hand-edit targets |
| Behavior posture | `behavior-preserving`, unless each authorized difference and controlling owner amendment is enumerated |
| Preconditions | Required blocker closures, characterization run roots, and owner approvals |
| Stop conditions | First unexpected observable difference, owner contradiction, unlisted file need, failing admission rule, or unapproved schema/contract consequence |
| Validation | Exact focused, boundary, browser, drift, finalization, and broader Make targets required by risk |
| Handoff | Required tracker rows, result roots, residual blockers, and rollback state |

An authorization for one slice MUST NOT be treated as authorization for a
subsequent slice. RS-02 requires RB-001 outcome A. RS-03 requires every
RMR-AC-200 through RMR-AC-209 matrix precondition that can be established
before production movement; completion requires all of them.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| RS-00 | Separate characterization authorization | Execute the closed snapshot, target-semantics, wire, projection, Collaboration, portability, security, and frontend matrices without production changes | Revisions and source-owner tests, fixtures/test support, applicable Timeline frontend/E2E, necessary authored verification input, this tracker | False baseline if behavior is inferred, old/new state differs, or TODO cells are guessed | All RMR-REQ-100 through RMR-REQ-215 characterization cases and RMR-AC-200 through RMR-AC-209 evidence | `make test-slice OWNER=module.revisions`; `make service-backed-test-slice OWNER=module.revisions`; applicable existing commands in Section 8 | Revert characterization-only additions that assert behavior not owned by adopted requirements; retain only valid evidence | Sections 4.3 and 4.8 contain proven values and run roots; every difference is classified; RB-001/RB-002 statuses are updated without overstating closure. |
| RS-01 | RS-00 and separate RS-01 authorization | Add a narrow caller-transaction Records envelope-reader dependency to `Appender`; inject the existing revision-assembly Records adapter; remove direct construction/import of the concrete Records store | `internal/modules/revisions/appender.go`, application port if a narrower interface is needed, `internal/app/revisionassembly` | Event incident/type/delete state, changed fields, view routing, transaction ownership | Appender dependency validation, revision-assembly construction, source-owner revision append, delete/restore/rollback event suites | `make test-slice OWNER=module.revisions`; `make service-backed-test-slice OWNER=module.revisions`; `make backend-module-boundary-check` | Revert constructor and adapter changes together; no compatibility shim | No production Revisions path constructs a Records store; event and atomicity behavior is unchanged. |
| RS-02 | RS-00, RB-001 outcome A, separate RS-02 authorization | Make authoritative source snapshots the sole history/rollback input and remove projection-first snapshot sourcing in one cutover | `projection_port.go`, delete/restore providers, rollback publication/apply, affected source-owner tests | Stored revision JSON, history diffs, rollback restore values, changed fields, affected views | Complete Sections 4.3-4.6 evidence; retained history, rollback, conflict-window, portability, frontend, and event consequences | Focused owner rows plus `make service-backed-test-slice OWNER=module.revisions` and applicable Section 8 commands | Before parity, production stays unchanged. After authorization, revert the single cutover as one unit; no dual path, fallback, shadow mode, or feature flag | Projection reads are absent from authoritative snapshot selection and every frozen behavior passes. Any observable shape change stops the slice and `requires later authorization`. |
| RS-03 | RS-00, RS-02, completed Section 4.8 matrix, separate RS-03 authorization | Add the immutable two-facet target-semantics catalog, adapt source-owned providers, and remove central source-kind/JSON-key branches | Revisions provider contracts/catalogs, history materializer/repository, rollback planner/apply/query, nine source-owner contribution/provider packages | Stable history refs/actions, companion selection, affected/protected records, rollback legality/order, provider closure | All RMR-AC-200 through RMR-AC-209 tests, including catalog negatives, static prohibitions, security, and failure injection | Relevant owner rows via `make test-slice` and `make service-backed-test-slice`; `make backend-module-boundary-check` | Revert each capability family with its owner contribution; no switch, hook, dynamic relation, or old-path fallback may remain as a compatibility mechanism | Every live kind has one contribution, source semantics are absent from generic coordination, admission stays fail-closed, and RMR-AC-200 through RMR-AC-209 pass. |
| RS-04 | RS-00 and separate RS-04 authorization | Move the view-schema lookup/snapshot adapter from the Revisions root into revision assembly while retaining the immutable copied conflict-field catalog | `conflict_field_provider.go`, `internal/app/revisionassembly`, source-owner contribution construction as required | View-schema admission, writable flags, conflict classes, startup failure order | Resolver immutability/fail-closed tests, revision-assembly catalog tests, affected source-owner conflict tests | `make test-slice OWNER=module.revisions`; affected source-owner slices; `make backend-module-boundary-check` | Revert composition and contribution signature changes together | Production `internal/modules/revisions` no longer imports the global view-schema registry, and runtime conflict resolution remains immutable. |
| RS-05 | RS-01, RS-02, RS-03, RS-04 as applicable, plus separate RS-05 authorization | Update authored verification, test-family, or boundary inputs only for new evidence/interfaces; regenerate downstream topology through Make | `contracts/verification/owners/module.revisions.json`, `tools/test_families/module.revisions.json`, `tools/backend_module_boundaries.json`, generated outputs only through generators | False ownership, missing execution, generated drift, hand-edit risk | New rows only where prior evidence cannot select the changed seam | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make harness-contract` | Revert owner input and its generated projection together | Authored inputs and generated outputs agree; no generated file was manually edited. |
| RS-06 | RS-01, RS-02, RS-03, RS-04, RS-05 | Run focused, applicable browser, static, drift, and broad validation; update the tracker with exact run roots and remaining blockers | Validation artifacts and this tracker | Missed cross-owner, frontend, security, or generation regression | All retained owner rows; browser history/stateful rows when affected | Commands in Section 8, then `make agent-finalize` and broader checks according to risk | Revert the smallest failing structural slice; do not mask failures with compatibility paths | Required commands pass or failures are recorded with run roots and relation to the change; handoff is current. |

## 8. Validation Plan

The commands below were discovered from the live Make task surface and target
guidance. They are for a separately authorized characterization or production
task. Every executed command MUST be recorded with its result and run root
when one is emitted. A command not run MUST be reported as not run; it MUST NOT
be represented as passing.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make test-slice OWNER=module.revisions` | Owner-selected non-service-backed Revisions evidence | yes | Establish the RS-00 baseline and rerun after each structural slice. |
| integration | `make service-backed-test-slice OWNER=module.revisions` | Owner-selected PostgreSQL/application Revisions evidence | yes | Required for history, delete/restore, rollback, portability, or event consequences. |
| e2e/browser | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | History/inspector mutations, refresh, Collaboration, and token-cutover behavior | no | Required when a slice affects wire behavior, selectors, frontend refresh, events, or conflict lifecycle. |
| generated drift | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check` | Authored/generated OpenAPI, protocol, contract, and harness projections | no | Required whenever an authored contract, boundary, verification, or harness input changes. Never edit generated roots directly. |
| import-boundary/static | `make backend-module-boundary-check` | Revisions imports, source-table/field rules, SQL allowlist, and constructor ownership | yes | Required for every boundary slice; add `make frontend-import-boundary-check` only if frontend boundaries change. |
| full check | `make agent-finalize`; `make test-fast`; expand to `make check` | Repository-wide closure proportional to cross-owner risk | no | Run finalization before broader end-of-run verification. Supply `RESULTS_DIR` when retaining a successful full warm-check run; otherwise record that retained-run maintenance was skipped. |

For this tracker-only session, product tests and drift gates are unnecessary
because no product, contract, generated, or harness input changed. The required
session validation is `make lint-markdown`, followed by diff and status review.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| RT-001 | Fix scope, authority, baseline, and allowed write | WF-00 | DONE | none | Section 1 | Target and exclusions are explicit; later implementation is unauthorized. |
| RT-002 | Inventory all 73 live target files | WF-01 | DONE | RT-001 | Section 2 | Exact target path set and inventory rows match. |
| RT-003 | Map adopted owners and observable contracts | WF-02 | DONE | RT-001 | Sections 1 and 4 | Every discovered contract has an owner and test posture. |
| RT-004 | Identify characterization coverage and gaps | WF-03 | DONE | RT-002, RT-003 | Sections 4, 7, and 8 | Each future slice names tests to retain or add. |
| RT-005 | Classify live boundary and coupling findings | WF-04 | DONE | RT-002, RT-003 | Sections 3 and 5 | Each finding has evidence, risk, classification, owner, and action. |
| RT-006 | Define facade redesign and safe slice order | WF-05, WF-06 | DONE | RT-004, RT-005 | Sections 6 and 7 | Internal interface intent, dependencies, rollback, and exit criteria are decision-complete. |
| RT-007 | Discover validation and evidence-accounting commands | WF-07 | DONE | RT-006 | Section 8; Make target guidance | Commands are repository-owned and conditional scope is explicit. |
| RT-008 | Create and lint the tracker | WF-08 | DONE | RT-001 through RT-007 | This file; Markdown lint result | Tracker exists, lint passes, and only this file is changed. |
| RT-010 | Convert the tracker to NLSpec-style normative guidance | WF-02, WF-03, WF-05, WF-06 | DONE | RT-008 | `temp/analysis-notes.md`, `docs/research/nlspec-spec.md`, Sections 1, 4, 7, 11, and 12 | Normative requirements, defaults, interface boundaries, blocker dispositions, authorization gates, and binary acceptance criteria are complete and documentation validation passes. |
| IMP-000 | Execute RS-00 characterization matrices | RS-00 | BLOCKED | RT-010, separate RS-00 authorization | Sections 4.3, 4.4, 4.8, and 4.9 | Every matrix cell is evidence-backed, differences are classified, run roots are recorded, and RB-001/RB-002 statuses are updated. |
| IMP-001 | Inject Records envelope reader into Appender | RS-01 | DEFERRED | IMP-000, separate RS-01 authorization | Finding in Section 5 | Separate authorized implementation passes focused behavior and boundary checks. |
| IMP-002 | Remove projection-first history snapshot sourcing | RS-02 | BLOCKED | IMP-000, RB-001 outcome A, separate RS-02 authorization | RMR-REQ-100 through RMR-REQ-112 | Authoritative snapshots replace projection truth in one cutover and all four compatibility layers pass. |
| IMP-003 | Move source target semantics behind providers | RS-03 | BLOCKED | IMP-000, IMP-002, completed target matrix, separate RS-03 authorization | RMR-REQ-200 through RMR-REQ-215; RMR-AC-200 through RMR-AC-209 | Central semantic switches are absent, admission is fail-closed, and all history/rollback contracts pass. |
| IMP-004 | Move view-schema snapshot adapter to composition | RS-04 | DEFERRED | IMP-000, separate RS-04 authorization | Platform-coupling finding in Section 5 | Revisions root has no global view-schema dependency. |
| IMP-005 | Decide key-ring environment fallback ownership | Future planning | DEFERRED | Compatibility characterization, later authorization | Deferred finding in Section 5 | Owner and compatibility posture are explicit before code changes. |
| IMP-006 | Reconcile authored verification/boundary inputs if needed | RS-05 | DEFERRED | IMP-001 through IMP-004 | Section 8 | Authored and generated projections agree without manual generated edits. |
| RT-009 | Complete current planning handoff | WF-08 | DONE | RT-008 | Sections 10 through 12 | Session results, blockers, and binary criteria are current. |

## 10. Session Handoff Log

The requested output did not exist at the original planning session, so the
seven 19:47 EDT rows form the initial ledger. This revision preserves those
rows and appends one 21:26 EDT row to each table. The archived tracker remains
referenced evidence and is not copied into this ledger.

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 19:47 EDT | Codex / revisions planning audit | Scope and authority mapped from live owners; planning only | Inspected framework, domain, Core 00-05, archived tracker, live target; touched only this tracker | `sed`, `rg`, `git rev-parse`, `git status` | Target exists at the requested baseline; output was absent; no owner contradiction found | Production work requires later authorization | Tracker planning is complete; a later authorized task may begin with RS-00. |
| 2026-08-07 21:26 EDT | Codex / NLSpec-style tracker revision | Tracker requirements are subordinate implementation guidance with explicit normative defaults | Inspected `temp/analysis-notes.md`, `docs/research/nlspec-spec.md`, current tracker, adopted owners, and live repository evidence; touched only this tracker | `sed`, `rg`, `git rev-parse`, `git status`, `git ls-files`, `wc`, `date` | Baseline refreshed; no owner contradiction found; silence now authorizes no behavior or ownership change | RB-003 | Obtain separate RS-00 authorization before changing tests or evidence inputs. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 19:47 EDT | Codex / revisions planning audit | Legitimate Revisions bounded context with residual Records, projection-truth, source-switch, and view-schema seams | Inspected all 73 target files, revision/server/config/workbook/timeline assembly, source-owner contributions, backend boundary policy; touched only this tracker | `rg`, `sed`, `jq`, `git ls-files`, `wc`, `make explain-target TARGET=backend-module-boundary-check DETAIL=summary` | Current HTTP/conflict/history/rollback decomposition is retained; remaining findings are classified in Section 5 | RB-001, RB-002, RB-003 | Begin RS-00 only in a later authorized task. |
| 2026-08-07 21:26 EDT | Codex / NLSpec-style tracker revision | Snapshot authority and target-semantics interface decisions are closed; compatibility and exact target facts remain evidence work | Inspected projection-first snapshot helper, all ten authoritative snapshot adapters, existing delete/restore and rollback contracts, provider contributions, and source-specific history/rollback branches; touched only this tracker | `rg`, `sed`, `git ls-files` | Defined one-source snapshot rule, four-layer parity, two-facet catalog, fail-closed admission, and prohibited replacement patterns without changing code | RB-001 characterization, RB-002 target matrix, RB-003 authorization | Execute RS-00 matrices in a separately authorized characterization task. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 19:47 EDT | Codex / revisions planning audit | Frontend history is an external Timeline-workbook consumer, not part of the Go target | Inspected Timeline history port, generated-operation adapter, model, actions, inspector panel, and browser scenarios; touched only this tracker | `rg`, `sed` | No frontend shell, saved-view, or grid-vendor responsibility belongs in Revisions; browser behavior is frozen | None for planning; later wire/event changes require browser evidence | Preserve frontend files unless an authorized contract change requires them. |
| 2026-08-07 21:26 EDT | Codex / NLSpec-style tracker revision | Frontend remains an external contract consumer and an exact-compatibility layer | Inspected analysis recommendations and retained Timeline/frontend evidence from the current tracker; touched only this tracker | `sed`, `rg` | Frontend rendering, actions, refresh, selectors, and event consequences are explicit parity gates; no frontend change is proposed | RB-001 if any external consequence differs | Run applicable browser evidence only in an authorized RS-00 or affected production slice. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 19:47 EDT | Codex / revisions planning audit | Four HTTP operations, WS consequences, key-ring schema, portability rows, recovery contribution, and generated consumers mapped | Inspected Revisions OpenAPI source, Revisions contracts, Incident Bundle source/row schemas, WS contract, generated policy; touched only this tracker | `jq`, `rg`, `sed`, `make explain-target TARGET=generate-drift DETAIL=summary` | No authored or generated contract change is planned; generated roots remain untouched | Later behavior/schema change requires owner authorization first | Use RS-05 only if implementation adds real evidence or owner-input changes. |
| 2026-08-07 21:26 EDT | Codex / NLSpec-style tracker revision | Public and portability surfaces remain frozen; incompatible snapshot coexistence has an explicit conditional version rule | Inspected analysis notes, Revisions contract inventory, Incident Bundle posture, and generated-file policy; touched only this tracker | `sed`, `rg`, `git diff HEAD` | No contract or generated change occurred; `snapshot_schema_id` is permitted only after separate owner authorization and MUST NOT be inferred from keys | RB-001 outcome B if incompatibility is observable; RB-003 | Preserve authored and generated inputs unless a later authorization names the change. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 19:47 EDT | Codex / revisions planning audit | Verification ownership and future validation matrix discovered | Inspected all target tests, `module.revisions` verification owner and 43-row test family, task surface; touched only this tracker | `make help`, `make help-all`, `make task-guide ROLE=module-author OWNER=module.revisions`, `make explain-test-owner OWNER=module.revisions`, relevant `make explain-target` commands, `make lint-markdown` | Markdown lint passed at `.cartulary/test-results/20260807T235305Z-p2110213`; product tests were not run because this is documentation-only | None for tracker completion | Preserve the validation matrix for the later authorized implementation. |
| 2026-08-07 21:26 EDT | Codex / NLSpec-style tracker revision | Characterization matrices and binary evidence requirements are specified but unexecuted | Inspected analysis notes, NLSpec writing doctrine, current test/harness inventory, and Make guidance; touched only this tracker | `rg`, `sed`, `make lint-markdown`, inventory/traceability shell checks, `git diff --check` | Documentation checks passed; Markdown run root `.cartulary/test-results/20260808T012855Z-p2134753`; product tests were not run because only documentation changed | RB-003 prevents RS-00 test or harness changes | A separately authorized RS-00 task MUST record exact Make run roots and replace evidence TODOs. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 19:47 EDT | Codex / revisions planning audit | Mutation guard precedence, token v3 lifecycle, secret purpose, and startup admission are contract-frozen | Inspected Core 03/04, HTTP adapter/precedence test, conflict codec/key-ring/config, server startup preflight; touched only this tracker | `rg`, `sed` | Security behavior remains in scope as preserved observable behavior; no security implementation changed | Key-ring nil-env ownership is deferred, not a current blocker | Characterize before any authorized config-boundary change. |
| 2026-08-07 21:26 EDT | Codex / NLSpec-style tracker revision | Security and transaction ownership are explicit provider prohibitions and acceptance gates | Inspected authorization precedence evidence, provider boundaries, and analysis recommendations; touched only this tracker | `sed`, `rg` | Providers are prohibited from authorization, transaction ownership, unsafe errors/logs, dynamic SQL, and publication; unauthorized calls MUST not reach providers | RB-003; key-ring fallback remains deferred | Add negative provider-invocation and atomic failure evidence only under RS-00/RS-03 authorization. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 19:47 EDT | Codex / revisions planning audit | Tracker captures live residual work without reopening completed historical decomposition | Inspected live/archive delta and current worktree; touched only this tracker | `git log`, `git diff`, `git status`, source searches | Three stable blockers are recorded; no production refactor was performed | RB-001, RB-002, RB-003 | Close documentation validation; a future session starts with RS-00 after explicit authorization. |
| 2026-08-07 21:26 EDT | Codex / NLSpec-style tracker revision | Documentation decisions are complete; factual parity, the exact target matrix, and all implementation remain intentionally open | Inspected the revised tracker against notes, NLSpec doctrine, live source evidence, diff, and worktree; touched only this tracker | `make lint-markdown`, inventory/traceability checks, `git diff --check`, `git diff HEAD`, `git status --short` | Tracker uses precise normative language, explicit defaults, mapping tables, stop rules, and binary criteria; no production refactor occurred | RB-001, RB-002, RB-003 | Authorize RS-00 separately; do not start RS-02 or RS-03 from this tracker alone. |

## 11. Open Questions and Blockers

No `BLOCKED: owner contradiction` entry is required because the inspected owner
documents agree. The following items block safe implementation, not completion
of this planning tracker.

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | The authority decision is closed: authoritative source snapshots are the sole permitted input to new revision snapshots. Exact compatibility is not yet proven. | Removing projection truth can alter stored history JSON, diffs, rollback values, changed fields, portability, or frontend behavior | Separately authorized RS-00 execution of Sections 4.3-4.6; outcome B or C also requires controlling-owner and compatibility authorization | `DECISION CLOSED / CHARACTERIZATION REQUIRED`; RS-02 remains blocked |
| RB-002 | The interface design is closed: a separate immutable two-facet target-semantics catalog supplies source-owned semantics. The exact current matrix is incomplete. | Unproven rows can merely relocate a switch or change associations, opaque selectors, companions, affected/protected records, and inverse ordering | Separately authorized RS-00 completion of Section 4.8 and RMR-AC-200 through RMR-AC-209 | `DESIGN CLOSED / TARGET MATRIX REQUIRED`; RS-03 remains blocked |
| RB-003 | This task authorizes documentation only. It authorizes neither RS-00 characterization changes nor production implementation. | The sole permitted write is this tracker; tests, contracts, generated files, migrations, configuration, harness inputs, and runtime files are outside the task | A separate authorization conforming to Section 7.1 or 7.2 | `AWAITING IMPLEMENTATION AUTHORIZATION` |

### 11.1 Binary blocker closure

| ID | Blocker | Closure condition |
| --- | --- | --- |
| RMR-AC-100 | RB-001 | Every Section 4.3 matrix cell is executed against old and candidate readers in the same transaction; every Section 4.4 layer passes; every difference is classified; outcome A, B, or C is recorded with run roots. |
| RMR-AC-101 | RB-001 | Outcome A is the only behavior-preserving closure. It requires no difference except `representation_only`, no fallback machinery, and explicit evidence for every public and portability consequence. |
| RMR-AC-102 | RB-001 | Outcome B remains blocked until adopted owners authorize the difference and define migration, versioning, compatibility, and any `snapshot_schema_id` behavior. |
| RMR-AC-103 | RB-001 | Outcome C remains blocked until the defective owner adapter is repaired separately and the complete matrix is rerun. |
| RMR-AC-210 | RB-002 | Every Section 4.8 `TODO` cell is replaced by cited live evidence, the 14-token set is proven complete, and no source behavior is inferred. |
| RMR-AC-211 | RB-002 | RMR-AC-200 through RMR-AC-209 all pass with exact run roots and static-search evidence. |
| RMR-AC-300 | RB-003 | The authorization names its baseline, exactly permitted slice, paths, behavior posture, prerequisites, stop conditions, validation, and handoff requirements. |
| RMR-AC-301 | RB-003 | Characterization and production authorization are separate; authorization for one slice does not authorize another. |
| RMR-AC-302 | RB-003 | Any authorized behavior change enumerates the exact difference, controlling owner amendment, compatibility/migration posture, and additional tests. Omission means no behavior change. |

## 12. Binary Completion Criteria

Tracker completion and implementation closure are independent. This tracker
MAY be complete while RB-001 through RB-003 remain open, provided their
decisions, evidence gaps, defaults, and authorization gates are unambiguous.

### 12.1 Requirement traceability

| Requirement or criterion range | Controlling owner | Workstream or slice | Required evidence | Current state |
| --- | --- | --- | --- | --- |
| RMR-REQ-001 through RMR-REQ-010 | Core 00-04 and Testing Harness owner within their scopes | All workflows and slices | Contract freeze map, owner-selected tests, boundaries, generated policy | Defined; execution deferred |
| RMR-REQ-100 through RMR-REQ-112 | Core 02, plus Core 01/03/04 for external consequences | WF-03, RS-00, RS-02 | Sections 4.3-4.6, history/rollback/conflict/portability/browser evidence | Decision defined; characterization blocked by RB-003 |
| RMR-REQ-200 through RMR-REQ-215 | Core 01/02/04 and source owners within their scopes | WF-05, RS-00, RS-03 | Section 4.8 and RMR-AC-200 through RMR-AC-209 | Design defined; matrix execution blocked by RB-003 |
| RMR-AC-100 through RMR-AC-103 | Same as snapshot requirements | RS-00 and RS-02 gate | Same-state old/new comparisons and classified outcomes | Not executed |
| RMR-AC-200 through RMR-AC-211 | Same as target-semantics requirements | RS-00 and RS-03 gate | Exact matrix, catalog negatives, equivalence, failure, security, and static evidence | Not executed |
| RMR-AC-300 through RMR-AC-302 | User task authorization under adopted owners | Every future task | Complete Section 7.1 or 7.2 authorization | Awaiting authorization |

### 12.2 Tracker-document completion

| ID | Criterion | Status | Evidence |
| --- | --- | --- | --- |
| RMR-AC-001 | Every file in `internal/modules/revisions` is inventoried or explicitly out of scope. | PASS | Exact path-set comparison found 73 live paths and 73 identical Section 2 rows; none is excluded. |
| RMR-AC-002 | Every discovered public contract risk has an owner and test posture. | PASS | Section 4 maps HTTP, authorization, history, source providers, projection, WebSocket, conflicts, portability, recovery, frontend, generated, and harness surfaces. |
| RMR-AC-003 | Every proposed workflow has dependencies and exit criteria. | PASS | Section 6 defines WF-00 through WF-08 with predecessors, successors, validation, and checkpoints. |
| RMR-AC-004 | Every proposed implementation slice is behavior-preserving unless explicitly marked `requires later authorization`. | PASS | Sections 4 and 7 freeze behavior; RS-02 stops on any non-representation difference. |
| RMR-AC-005 | Validation commands are discovered or marked `TODO` with a reason. | PASS | Section 8 records Make-owned focused, browser, drift, boundary, and broader commands. |
| RMR-AC-006 | Contradictions are marked `BLOCKED: owner contradiction`. | PASS | No owner contradiction was found; repository deviations are findings and Section 11 defines the mandatory conflict posture. |
| RMR-AC-007 | Repository/framework mismatches are recorded as planning findings. | PASS | Sections 1, 3, and 5 distinguish the 73-file live decomposition from generic and archived evidence. |
| RMR-AC-008 | Snapshot sourcing has a complete decision rule, explicit defaults, closed matrices, difference taxonomy, and binary outcomes without claiming parity. | PASS | Sections 4.2-4.6 and RB-001. |
| RMR-AC-009 | Target semantics has an unambiguous two-facet interface, closed ownership, fail-closed defaults, exhaustive current-token set, prohibited alternatives, and binary tests without guessing open evidence. | PASS | Sections 4.7-4.9 and RB-002. |
| RMR-AC-010 | Characterization and production have separate, complete authorization templates and stop conditions. | PASS | Sections 7.1, 7.2, and RB-003. |
| RMR-AC-011 | Handoff sections preserve prior history and contain a current row for all seven workstream-specific tables. | PASS | Section 10 retains the seven prior rows and adds exactly one current-session row to each table. |
| RMR-AC-012 | Markdown lint, diff hygiene, exact inventory comparison, requirement traceability, and worktree-scope checks pass. | PASS | `make lint-markdown` passed at `.cartulary/test-results/20260808T012855Z-p2134753`; the path comparison, unique-ID audit, `git diff --check`, diff review, and final status review passed. |

No product validation is claimed by this documentation revision. No production
refactor was performed.

This tracker revision is complete. The repository diff is limited to this
tracker, and RB-001 through RB-003 remain implementation gates exactly as
recorded in Section 11.
