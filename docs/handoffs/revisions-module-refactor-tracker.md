# revisions Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

This tracker is the controlling execution ledger for the authorized Revisions
remediation. It records required gates, defaults, interfaces, evidence, and
handoff state. It is not an adopted behavioral owner; adopted specifications
continue to own behavior and contracts.

| Item | Value |
| --- | --- |
| Target path | `internal/modules/revisions` |
| Normalized target label | `revisions` |
| Output path | `docs/handoffs/revisions-module-refactor-tracker.md` |
| Document class | Implementation-support refactor tracker and handoff |
| Document revision | 4, completed remediation and validation handoff |
| Repository baseline | Clean commit `b88ce7b5f43a465c1f817dfc9fa9068bbafef7ac`; 73 tracked files under `internal/modules/revisions` |
| Execution status | COMPLETE; WS-00 through WS-12 are `DONE`, every active blocker is closed, and no implementation slice remains |
| Allowed change | The specification, contracts, authored implementation, tests, migration, documentation, and harness inputs named by WS-00 through WS-12; generated output only through Make-owned generation |
| Non-goals | No HTTP, WebSocket, OpenAPI operation, UI selector, conflict-token wire-format, or frontend-port redesign; no legacy snapshot reader, dual path, fallback, shadow write, feature flag, historical-shape inference, or hand edit of generated output |
| Authorization | The user-authorized remediation plan dated 2026-08-09 governs WS-00 through WS-12, subject to adopted-owner precedence and each workstream checkpoint |

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
| Omission default | If this tracker does not expressly authorize a behavior, contract, schema, storage, configuration, ownership, or interface change, that change is not authorized. Existing behavior is retained only when adopted owners require it or it has explicit continuing value in the approved compatibility posture. |
| Owner conflict default | An adopted owner prevails. Work MUST stop, and Section 11 MUST record `BLOCKED: owner contradiction`; this tracker MUST NOT select a side. |
| Unknown-evidence default | The tracker MUST record missing evidence as a blocking workstream item and MUST NOT infer an answer from a filename, target label, JSON key, or historical phase name. |

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

The clean WS-00 baseline contained 73 tracked Go files: 47 production files
and 26 test files, totaling 14,497 lines. The completed worktree contains 75
files after the exact additions and removals in Section 2.1. The framework's
generic candidate state and the
archived tracker's earlier inventory do not match that live shape. The live
repository already contains dedicated HTTP, conflict, history, and rollback
components created by the earlier refactor. This tracker preserves those
completed boundaries and plans only from the remaining live findings.

No owner-document contradiction was found. Repository deviations from adopted
owners are recorded as implementation findings rather than owner conflicts.

### 1.4 Authorized correction and checkpoint posture

This execution deliberately corrects the history representation instead of
preserving accidental projection-derived or schema-less snapshots. Existing
pre-production databases with mutation rows are reset rather than migrated.
Schema-less snapshots receive no reader, translator, backfill, compatibility
alias, or shape inference. Incident Bundles containing schema-less revision
snapshots fail closed; version-1 and version-2 bundles produced after the
cutover remain supported with canonical snapshot envelopes.

The following external consequences retain continuing value and therefore
remain exact: public operations and envelopes, authorization precedence,
opaque history selectors, Collaboration event consequences, conflict-token
version 3, and current Incident Bundle versions. Internal Go callers migrate
within their owning workstream and receive no deprecation shim.

Every workstream uses this mandatory checkpoint before the next workstream
begins:

1. Update this tracker with status, files changed, substantive result,
   compatibility impact, validation commands and run roots, blockers,
   rollback posture, residual risk, and next workstream.
2. Mark the completed workstream `DONE`; unresolved validation leaves it
   `BLOCKED` and prohibits the next production change.
3. Run `make lint-markdown`, `git diff --check`, and status/diff review after
   the tracker edit.
4. Preserve prior handoff rows and append a timestamped checkpoint row.

## 2. Baseline Repository Inventory

The table contains one row for every tracked file under the WS-00 baseline.
Section 2.1 applies the exact final-topology delta. “Contracts” names the
observable or generated surface affected by the file; it does not mean that
the file owns or may hand-edit a generated artifact.

| Path | Baseline responsibility and completed-slice annotation | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/revisions/appender.go` | Appends change sets, mutation entries, record revisions, and transactional Collaboration intents | `Appender`, append parameter types, `HistoricalIntentPolicy`, `IntentAppender`, constructor and append methods | Revision assembly and source-owner mutation facades | PostgreSQL transaction, Collaboration, injected Records envelope reader, record-view catalog | Revisions integration/delete/rollback tests and source-owner mutation suites | Revision rows and `record_changed` payload consequences | Revisions, with Records and Collaboration ports | high | WS-03 injects the transaction-bound Records reader; Revisions no longer imports or constructs the concrete Records store. |
| `internal/modules/revisions/application_boundary_test.go` | Static application/HTTP boundary evidence | `TestRevisionsApplicationAndHTTPBoundaries_Unit` | `module.revisions` support-unit row | Go source inspection | Self | Verification `http_application_boundary` and facade closure | Revisions tests | medium | Protects the completed root-versus-HTTP split. |
| `internal/modules/revisions/application_ports.go` | Transport-neutral command, transaction, authorization, idempotency, and record-envelope ports | `ActorID`, `TransactionRunner`, `CommandAuthorizer`, `IdempotencyPort`, `RecordEnvelopePort`, command/query DTOs | Command service, HTTP adapter, revision assembly | UUID, PostgreSQL transaction types | Command service and application-boundary tests | Internal application interface behind four HTTP operations | Revisions | high | Consumer-owned port surface; no public wire change is planned. |
| `internal/modules/revisions/attribution_resolver_registry.go` | Validates profile-scoped imported-actor resolver contributions | `ExtensionClaim`, `AttributionResolverRegistry` and methods | Revision assembly and Incident Bundle integration | Imported-attribution resolver port | Registry and portability tests | Incident Bundle attribution and profile claims | Revisions with Incident Bundles contribution | high | Deliberately avoids direct sidecar-table joins. |
| `internal/modules/revisions/attribution_resolver_registry_test.go` | Resolver admission and duplicate-profile evidence | Two registry tests and resolver fake | `module.revisions` portability/support rows | Registry API | Self | Attribution verification mapping | Revisions tests | low | Test-only. |
| `internal/modules/revisions/command_models.go` | Normalizes command requests, rollback selectors, hashes, and route keys | `DeleteRestoreRequest`, `RollbackTarget`, `RollbackRequest`, hash helpers | Command service and HTTP decoding | JSON, SHA-256, UUID | Command, route, and rollback tests | Delete/restore/rollback request identity and idempotency | Revisions | high | Internal model mirrors stable wire semantics without owning HTTP decoding. |
| `internal/modules/revisions/command_service.go` | Thin public application facade for history, delete, restore, and rollback | `CommandService`, dependencies, constructor, four methods | Revisions HTTP adapter and revision assembly | Application ports, provider catalogs, history/mutation coordinators | `command_service_test.go` plus route/integration suites | Four Revisions OpenAPI operations | Revisions | high | Legitimate thin facade. |
| `internal/modules/revisions/command_service_test.go` | Constructor closure, provider completeness, and validation-catalog evidence | Three principal tests and dependency fakes | `module.revisions.facade` and provider rows | Application interfaces and provider contracts | Self | Facade and provider verification IDs | Revisions tests | medium | Test-only. |
| `internal/modules/revisions/command_store.go` | Private dependency container used by command components | `ImportedAttributionResolver` interface | Command service and private coordinators | Application ports, catalogs, projection and appender capabilities | Command/integration tests indirectly | None directly; supports public command behavior | Revisions | medium | Private composition state, not a second public facade. |
| `internal/modules/revisions/conflict_field_provider.go` | Removed in WS-04; revision assembly now performs the startup-only view-schema conversion | None | None | None | Field-resolver and revision-assembly tests | View-schema writeability/conflict classes | Application composition | resolved | Revisions retains only the immutable copied descriptor catalog. |
| `internal/modules/revisions/conflicts/configuration.go` | Revisions configuration overlay, normalization, findings, and manifest-read errors | `Configuration`, findings/errors, overlay and validation helpers | Config assembly and server startup | Path normalization and Revisions config vocabulary | Conflict-token/config and server/config tests | Deployment config v2 Revisions namespace | Revisions configuration semantics | high | Secure file reading remains in server assembly. |
| `internal/modules/revisions/conflicts/conflict_token.go` | Issues and verifies opaque v3 conflict tokens | Token claims/binding/codec, constructor, hash helper | Workbook/source-owner conflict flows | AES-GCM, key ring, clock/entropy injection | Token lifecycle tests and source-owner conflict suites | Conflict token v3 and `same_field_conflict` | Revisions conflicts | very high | Generic conflict mechanic explicitly owned by Revisions. |
| `internal/modules/revisions/conflicts/conflict_token_test.go` | Token sealing, expiry, rotation, entropy, key isolation, and config evidence | Five tests and error reader | Revisions conflict verification rows | Token/config/key-ring APIs | Self | Key-ring schema and v3 lifecycle | Revisions tests | medium | Test-only. |
| `internal/modules/revisions/conflicts/conflict_window.go` | Pure descriptor-driven revision-window conflict calculation | Conflict-window types and build/decode/diff helpers | Source-owner workbook conflict handlers | Immutable field descriptors and revision rows | Resolver/window tests and source-owner conflict suites | Same-field conflict field/window behavior | Revisions conflicts | high | Does not read the global view registry at runtime. |
| `internal/modules/revisions/conflicts/consolidation_test.go` | Guards one generic conflict package and absence of source-owner imports | `TestConflictCapabilitiesAreConsolidated_Unit` | Revisions conflicts verification row | Filesystem/source inspection | Self | Conflict capability consolidation evidence | Revisions tests | low | Test-only architecture evidence. |
| `internal/modules/revisions/conflicts/field_resolver.go` | Immutable field descriptor and resolver catalogs | Descriptor, contribution, resolver, catalog types and constructors | Revision assembly and conflict consumers | Source-contributed copied descriptors | Resolver tests and revision-assembly tests | View-schema field/conflict contract | Revisions conflicts | high | Intended runtime replacement for global lookup. |
| `internal/modules/revisions/conflicts/field_resolver_test.go` | Catalog closure, immutability, fail-closed, and no-runtime-global-read evidence | Three tests | Revisions conflict verification rows | Resolver API and source inspection | Self | Conflict field-resolver verification | Revisions tests | medium | Does not cover physical placement of the startup adapter. |
| `internal/modules/revisions/conflicts/keep_saved.go` | Revisionless `keep_saved` idempotency transaction coordination | Command/idempotency/target/result types, `KeepSaved`, decoder | Workbook conflict resolution flows | Injected transaction runner, idempotency, semantic loader | Source-owner conflict and workbook tests | Explicit conflict-resolution behavior | Revisions conflicts | very high | Source owners retain state loading and revalidation. |
| `internal/modules/revisions/conflicts/key_ring.go` | Strict key-ring JSON parsing, explicit secret resolution, rotation admission, and purpose registration | `ConflictTokenKeyRing`, parse options, parse functions | Server startup and conflict test support | Explicit environment snapshot, secret-purpose registry | Conflict-token tests and server startup tests | `contracts/revisions/conflict-token-key-ring.v1.schema.json` | Revisions config with platform startup adapter | very high | WS-05 rejects nil input; server assembly owns process-environment capture and explicit empty input never falls back. |
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

All 73 baseline target files are represented above; no baseline file is marked
out of scope.

### 2.1 Final topology delta

The completed path set is the 73-file baseline minus five removed paths plus
seven added paths, for exactly 75 files. Read together, the baseline table and
this delta inventory every file in the completed target.

| Disposition | Path | Final responsibility or reason |
| --- | --- | --- |
| added | `internal/modules/revisions/candidate_semantics_test.go` | Candidate and exact compiled-catalog admission evidence. |
| added | `internal/modules/revisions/canonical_snapshot.go` | Opaque captured snapshots and canonical envelope validation. |
| added | `internal/modules/revisions/history_associations_migration_test.go` | Fresh/reset-required migration, index, and association-fact evidence. |
| added | `internal/modules/revisions/rollback_provider_validation.go` | Generic validation for catalog-selected row and non-row providers. |
| added | `internal/modules/revisions/target_semantics.go` | Immutable history facets, dispatch classes, and unified target catalog. |
| added | `internal/modules/revisions/target_semantics_nonrow_catalog_test.go` | Non-row contribution closure and dispatch evidence. |
| added | `internal/modules/revisions/target_semantics_row_catalog_test.go` | Row contribution closure and record-provider admission evidence. |
| removed | `internal/modules/revisions/conflict_field_provider.go` | Startup view-schema conversion moved to revision assembly in WS-04. |
| removed | `internal/modules/revisions/nonrow_provider_catalog.go` | Superseded by the unified immutable target-semantics catalog in WS-10. |
| removed | `internal/modules/revisions/nonrow_provider_catalog_test.go` | Replaced by unified non-row catalog evidence. |
| removed | `internal/modules/revisions/row_provider_catalog.go` | Superseded by the unified immutable target-semantics catalog in WS-10. |
| removed | `internal/modules/revisions/row_provider_catalog_test.go` | Replaced by unified row catalog evidence. |

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

The completed target is therefore best described as:

- a legitimate thin application/service facade;
- a view/projection orchestration consumer, but not a projection owner;
- a transport-adjacent adapter through `httpapi`;
- a persistence-adjacent adapter for Revisions-owned tables;
- a mutation coordinator for delete, restore, rollback, and revisionless
  conflict outcomes; and
- a cohesive bounded context whose source semantics arrive through immutable
  owner contributions and whose platform dependencies terminate at application
  or HTTP adapters.

It is not supported by repository evidence as a frontend shell/controller,
grid-vendor integration, saved-view owner, tabular-ingest owner, or a general
catch-all for Timeline, Entities, Indicators, Evidence, Links, or other source
modules.

The table below preserves the WS-00 diagnosis and planned disposition. Section
5 and the append-only workstream ledger record the completed result.

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

No public HTTP, WebSocket, OpenAPI operation, generated TypeScript API, or UI
selector changes are authorized. The adopted corrective storage changes are
limited to canonical snapshot envelopes, indexed history associations, and
normalized conflict facts in migration 61; internal Go interfaces change only
to improve dependency direction and source-owner contribution semantics.
References to “current” behavior in this freeze map describe the WS-01
characterization baseline; the workstream ledger records its preserved or
intentionally corrected final disposition.

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
| Recovery contribution | Revisions tables; Recovery aggregate | `recovery_state.go` and recovery fixtures/catalog | Recovery catalog and broader recovery checks | Preserve classification of five authoritative history tables | medium | Migration 61 adds the conflict-fact table to the owner contribution. |
| Frontend Timeline history port and inspector | Web application/Timeline workbook | Timeline port, generated-operation adapter, history model/actions/panel, UI contracts | Frontend units; `history.spec.ts`, inspector, visual, and accessibility scenarios | Run browser rows when wire behavior, selectors, refresh, or event consequences are touched | high | No direct Revisions frontend package or grid-vendor dependency exists. |
| Generated OpenAPI/protocol/UI surfaces | Authored owner inputs and generators | Revisions OpenAPI source, generated protocol-ts, ui-contract selectors | OpenAPI compatibility, frontend units/browser | No generated change for structural slices; regenerate only from an authorized owner-input change | high | Never hand-edit generated roots. |
| Harness/test accounting | Verification owners and authored test-family catalog | Final authored owner/family catalogs and generated topology | Harness contract and owner slices | Keep every new seam selectable and regenerate topology through Make | medium | Routing is evidence accounting, not runtime architecture. |

### 4.1 Refactor-wide normative requirements

| ID | Requirement | Owner basis | Planned evidence |
| --- | --- | --- | --- |
| RMR-REQ-001 | Every authorized slice MUST preserve every observable consequence in the freeze map unless that slice names a separately authorized owner change and compatibility plan. | Core 00-04 | Pre-change characterization plus focused and broad post-change evidence |
| RMR-REQ-002 | No slice authorizes a public HTTP, WebSocket, OpenAPI operation, TypeScript API, UI-selector, bundle-version, or configuration behavior change. Storage changes are limited to the adopted canonical envelope, indexed association, normalized conflict-fact, and reset posture; generated changes MUST come only from amended owner inputs. | Core 01-04 | Contract drift, browser, portability, migration, and generated-artifact checks |
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
| RMR-REQ-110 | Characterization MUST select either representation-only parity or an adopted corrective cutover. The old source-selection path MUST be removed in the same cutover; no dual reader, fallback, shadow writer, or feature flag may remain. |
| RMR-REQ-111 | Every canonical snapshot MUST carry an admitted `snapshot_schema_id`. Historical interpretations MUST NOT coexist or be inferred from key presence; schema-less history requires reset and fails closed. |
| RMR-REQ-112 | Current Incident Bundle versions retain their exact outer rows and admit canonical snapshot envelopes only. A schema-less snapshot MUST be rejected rather than silently reinterpreted, translated, or assigned an inferred version. |

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

WS-01 executed all ten source-owner snapshot cases in
`TestDeleteRestoreConcreteSourceAdapterMatrix_Integration`. Each authoritative
snapshot contains the envelope and source objects with matching record
identity. Every projection-backed case returns a flat row, not `{record,
source}`, and is structurally different; unsupported view-query cases take the
existing source fallback. This is an intentional `projection_only_member` plus
`source_member_missing_from_projection` correction, not parity. The approved
canonical envelope and reset posture resolve the representation break while
the public consequence layers remain frozen.

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

WS-01 records outcome B for the stored representation. The 2026-08-09
authorization already supplies the required owner-amendment, reset, and
no-legacy-reader posture; WS-02 adopts that authority and WS-08 performs the
single cutover. Layers B through D remain exact compatibility gates.

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
producers and consumers. The cells below record live mutation, provider,
repository, and test evidence captured by WS-01; they are characterization,
not authority for the replacement design.

| Target kind | Semantic owner | Target ID shape | Primary history record | Additional associated records | Logical item and addressability | Selector inputs | Provider evidence |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `assessment` | Assessments | UUID record ID | Same assessment record | None | Target-ID grouping; one history entry when addressed by the assessment record | Exact persisted target kind/ID and current history fields | Assessment row provider and history/rollback tests |
| `entity_alias` | Entities | `entity_alias:<uuid>` or current content-derived logical ID | Owning Host or Identity by mutation value | None | Target-identity grouping; currently not individually addressable | Alias ID or normalized owner/value identity | Entity collection provider, patch/merge producers, and collection rollback tests |
| `entity_mention` | Entities | UUID mention ID | Current `source_record_id` association | None; resolved records are not current history associations | Mention target grouping; one history entry from the source-record association | Mention ID, source record, source field, entity type, and retained resolution state | Mention provider, history repository predicate, and rollback tests |
| `entity_preserved_identifier` | Entities | Current content-derived logical ID | Owning Host or Identity by mutation value | None | Target-identity grouping; currently not individually addressable and individual rollback requires the change set | Owner/type/value identity | Entity collection provider and merge/repoint tests |
| `evidence` | Evidence | UUID record ID | Same Evidence record | None | Target-ID grouping; one history entry when addressed by the Evidence record | Exact persisted target kind/ID and current history fields | Evidence row provider and history/rollback tests |
| `host` | Entities | UUID record ID | Same Host record | None | Target-ID grouping; one history entry when addressed by the Host record | Exact persisted target kind/ID and current history fields | Host row provider and history/rollback tests |
| `identity` | Entities | UUID record ID | Same Identity record | None | Target-ID grouping; one history entry when addressed by the Identity record | Exact persisted target kind/ID and current history fields | Identity row provider and history/rollback tests |
| `indicator` | Indicators | UUID record ID | Same Indicator record | None | Target-ID grouping; one history entry when addressed by the Indicator record | Exact persisted target kind/ID and current history fields | Indicator row provider and history/rollback tests |
| `indicator_observation` | Indicators | UUID observation ID | Current `source_record_id` association | Resolved Indicator when currently associated | Current observation grouping and addressability | Exact observation and association inputs | Indicator non-row provider, association SQL, and rollback tests |
| `indicator_state_interval` | Indicators | UUID interval ID | Current `indicator_record_id` association | None | Interval target grouping; one history entry from the Indicator association | Interval ID and Indicator ID | Indicator child provider, association SQL, and rollback tests |
| `record` | Applicable source owner selected through the admitted record-type catalog; Revisions owns generic dispatch | UUID record ID | Same record | None | Generic target-ID row grouping; one history entry | Persisted token/ID plus envelope record type | Artifacts, Parties, Task Requests, Decisions, Evidence, imports, and generic delete/restore producers |
| `record_link` | Links | UUID link ID | Current source-record association | Current destination-record association | Current link grouping and addressability | Exact link ID, source, destination, and relation inputs | Link non-row provider, history association SQL, and rollback tests |
| `record_tag` | Links | Raw tag UUID or `record_tag:<record_id>:<tag_id>` | Current `record_id` association | None | Both identities retain target grouping; one history entry from the owning-record association | Tag ID plus record/tag membership identity | Tag provider, merge producer, association SQL, and rollback tests |
| `timeline_record` | Timeline | UUID record ID | Same Timeline Event record | None | Target-ID grouping; one history entry when addressed by the Timeline record | Exact persisted target kind/ID and current history fields | Timeline row provider and history/rollback tests |

| Target kind | Companion rule | Protected-set rule | Current-state rule | Inverse order | Changed-field rule | Affected-record order | Deleted-target behavior | Portability behavior |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `assessment` | None | Assessment record | Source-owned assessment row plus envelope | One row inverse; reverse sequence within whole-set rollback | Preserve current row diff fields | Canonical record-ID order | Deleted rows retain current row-provider eligibility | Exact persisted mutation and revision round trip |
| `entity_alias` | None | Owning Host or Identity; whole change set required | Source-owned alias state | Reverse sequence in whole-set rollback; create tombstones, delete restores | Preserve current alias fields | Canonical owning-record order | Current collection-provider stale and eligibility rules | Exact persisted mutation round trip |
| `entity_mention` | Every same-change-set `record_link` is an individual-entry companion | Source record | Source-owned mention and companion state | Mention inverse, then companions in ascending original sequence; whole-set uses global reverse sequence | Mention/source-field and link fields | Canonical source-record order | Current mention/link provider eligibility | Exact mention and companion round trip |
| `entity_preserved_identifier` | None | Owning Host or Identity; whole change set required | Source-owned preserved-identifier state | Reverse sequence in whole-set rollback; create tombstones | Preserve current identifier fields | Canonical owning-record order | Create is reversible only through whole-set selection | Exact persisted mutation round trip |
| `evidence` | None | Evidence record | Source-owned Evidence row plus envelope | One row inverse; reverse sequence within whole-set rollback | Preserve current row diff fields | Canonical record-ID order | Deleted rows retain current row-provider eligibility | Exact mutation, revision, and evidence-object round trip |
| `host` | None | Host record | Source-owned Host row plus envelope | One row inverse; reverse sequence within whole-set rollback | Preserve current row diff fields | Canonical record-ID order | Deleted rows retain current row-provider eligibility | Exact mutation and revision round trip |
| `identity` | None | Identity record | Source-owned Identity row plus envelope | One row inverse; reverse sequence within whole-set rollback | Preserve current row diff fields | Canonical record-ID order | Deleted rows retain current row-provider eligibility | Exact mutation and revision round trip |
| `indicator` | Observations and intervals remain separate, not implicit companions | Indicator record | Source-owned Indicator row plus envelope | One row inverse; reverse sequence within whole-set rollback | Preserve current row diff fields | Canonical record-ID order | Deleted rows retain current row-provider eligibility | Exact mutation and revision round trip |
| `indicator_observation` | Source record and before/after resolved Indicator records are affected, not separate mutation companions | Canonical union of source and resolved Indicator record IDs | Source-owned observation state | One provider inverse; global reverse sequence in whole-set rollback | Preserve current observation fields | Canonical record-ID order | Create tombstones; resolve restores prior resolution | Exact observation round trip |
| `indicator_state_interval` | Indicator is the affected record, not a separate mutation companion | Indicator record | Source-owned interval state | One provider inverse; global reverse sequence in whole-set rollback | Preserve current interval fields | Canonical record-ID order | Create tombstones under current provider rules | Exact interval round trip |
| `record` | Selected row provider supplies source companions, if any | Target record remains protected where current operation requires it | Source-owned row snapshot plus envelope | Preserve current row-provider inverse order | Preserve exact changed-field calculation | Preserve current result order | Preserve current deleted-row eligibility | Exact mutation and revision round trip for every producing family |
| `record_link` | Attached-evidence links expose same-change-set row targets as atomic companions | Source and destination records | Source-owned fixed-SQL link state | One provider inverse; global reverse sequence in whole-set rollback | Preserve current link fields | Canonical endpoint-ID order | Create tombstones and delete restores after endpoint validation | Exact link round trip |
| `record_tag` | None; both target-ID forms identify the same membership | Owning record | Source-owned fixed-SQL tag membership state | One provider inverse; global reverse sequence in whole-set rollback | Preserve current tag fields | Canonical owning-record order | Create tombstones; patch/delete restore | Exact tag round trip without normalizing historical IDs |
| `timeline_record` | None | Timeline record | Source-owned Timeline row plus envelope | One row inverse; reverse sequence within whole-set rollback | Preserve current row diff fields | Canonical record-ID order | Deleted rows retain current row-provider eligibility | Exact mutation and revision round trip |

### 4.9 RB-002 closure evidence

RB-002 remained open until tests proved all of the following:

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

WS-10 closed the implementation evidence gate for RMR-AC-200 through
RMR-AC-208. The focused Revisions, service-backed, nine source-owner, catalog
negative, failure-injection, security-precedence, lock-order, and static
boundary evidence cited in the WS-10 checkpoint proves generic row/non-row
dispatch and exact owner behavior. WS-12 closed RMR-AC-209 through the clean
browser, broad, and release graphs in its handoff row.

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Appender bypasses the injected Records boundary for an envelope lookup | Resolved in WS-03: Appender consumes `RecordEnvelopeTxReader`; revision assembly constructs and adapts the Records store | Concrete dependency could evade composition tests and broaden record authority | `fixed` | Records port injected by `internal/app/revisionassembly` | Static import/construction checks plus event, atomicity, and boundary suites pass. |
| History snapshot helper prefers projections to source snapshots | Resolved in WS-08: canonical source-owner captures are the only row-history shape, and live projection material is supplied separately | Projection-derived history could corrupt retained truth and couple it to disposable views | `fixed` | Source-owner `DeleteRestoreSource.SnapshotTx` through typed capture | Canonical/schema-less negative, rollback, portability, Collaboration, and static projection-read evidence pass. |
| History addressability embeds source target kinds and JSON field names | Resolved in WS-09: immutable history facets derive indexed association and addressability arrays before persistence | Central JSON interpretation could drift and scale poorly | `fixed` | Source owners through the compiled target-semantics catalog | Indexed generic lookup, deterministic bundle recomputation, migration admission, and static source-key evidence pass. |
| Rollback components embed source-kind families and special cases | Resolved in WS-10: planner, repository, and applier dispatch only on catalog-provided `row` or `non_row`; owner providers select companions and invert source state | Central switches could partially admit new targets or use stale source assumptions | `fixed` | Immutable target catalog plus row/non-row source providers | Fourteen-entry closure, catalog negatives, exact rollback suites, owner slices, and prohibited-switch searches pass. |
| Revisions root owns a process-global view-schema lookup adapter | Resolved in WS-04: revision assembly resolves declared IDs and copies descriptors before runtime construction | Startup composition was physically coupled to global state | `fixed` | `internal/app/revisionassembly` platform adapter | Static dependency checks and immutable field-resolver catalog tests pass. |
| Key-ring parsing optionally reads process environment | Resolved in WS-05: the parser requires an explicit map and server assembly snapshots `os.Environ` only when runtime options omit `Env` | Hidden ambient runtime input complicated deterministic configuration tests | `fixed` | Revisions config parser and server startup adapter | Nil/empty/hostile/explicit resolution and startup slices pass; no `os` import remains in Revisions. |
| Revisions reads and writes only owner history tables | Backend SQL allowlist and inspected repositories | Low if fixed SQL and allowlist remain exact | `intentional/no_action` | Revisions persistence adapters | Preserve the allowlist and reject dynamic relation metadata. |
| Dedicated HTTP package imports platform auth/HTTP/pagination and injected visibility adapters | `httpapi/**`, application boundary test, AC-528/AC-529 | Low while source-owner construction terminates in application assembly | `intentional/no_action` | Revisions HTTP adapter | Preserve four-operation mapping and guard precedence; WS-03 removed its concrete Records construction. |
| Collaboration intent creation remains transaction-bound and delivery-neutral | `Appender` uses injected `IntentAppender`; runtime delivery is outside target | Low if event payload and transaction ordering stay frozen | `intentional/no_action` | Revisions facts plus Collaboration delivery | Preserve one intent per affected live revision and import suppression. |
| Provider completeness uses a closed catalog | WS-10 removed handwritten record/target owner maps; `provider_contributions.go` compiles exact closure from the generated snapshot and target registries | Low while adopted projections and source contributions must agree before serving | `fixed` | Revisions aggregate catalog downstream of adopted contract projections | Keep fail-closed generated-registry closure and reject duplicate, missing, unexpected, cross-owner, or typed-nil contributions. |
| Frontend consumes generated HTTP operations and UI selectors outside the target | Timeline history adapter/model/actions/panel and E2E rows | Low for backend-only structural changes; high if wire/event behavior drifts | `intentional/no_action` | Web application/Timeline workbook | Do not move frontend or grid code; retain browser characterization. |
| Archived tracker claimed the earlier decomposition was complete while the WS-00 baseline retained residual seams | Archived tracker versus the WS-00 Appender, projection, and source-kind evidence | Trusting stale completion could have produced an inaccurate plan | `fixed` | Current tracker | This ledger records the corrected baseline, every serial slice, and final topology. |

Generated outputs were regenerated only through Make after their authored
owner inputs changed. Final generation drift and generated-artifact policy
checks pass; generated output must never be edited directly.

## 6. Authorized Remediation Workstreams

Each row is a separate, serial workstream. Its tracker checkpoint is a hard
dependency of the following row.

| Workstream | Phase | Depends on | Scope | Exit criteria | Status |
| --- | --- | --- | --- | --- | --- |
| WS-00 | Authority and evidence | none | Rebaseline this ledger to the live repository and record the authorized correction/reset posture. | Exact baseline, 73-path inventory, compatibility posture, workstream ledger, Markdown lint, and diff review are recorded. | DONE |
| WS-01 | Authority and evidence | WS-00 | Complete record-family snapshots and the fourteen-token target matrix; freeze valuable public, portability, event, and security consequences. | No unresolved matrix cells; intentional preservation and correction are evidence-backed. | DONE |
| WS-02 | Authority and evidence | WS-01 | Adopt the Revisions boundary decision; amend Core 00/01/02/04, domain navigation, authored contracts, and generated projections. | Snapshot schemas, target registry, associations, dispatch, reset policy, and acceptance criteria have one owner and executable projections. | DONE |
| WS-03 | Independent boundaries | WS-02 | Inject a transaction-bound Records envelope reader into Appender. | No concrete Records construction/import in Revisions; event and atomicity suites pass. | DONE |
| WS-04 | Independent boundaries | WS-03 | Resolve and copy view-schema conflict descriptors in revision assembly. | No Revisions production import of `platform/viewschema`; catalog negative and conflict suites pass. | DONE |
| WS-05 | Independent boundaries | WS-04 | Replace ambient conflict-key environment reads with explicit server-assembled input. | No `os` import in Revisions; empty, hostile, rotation, and startup cases pass. | DONE |
| WS-06 | Typed transition | WS-05 | Add candidate typed snapshot capture, separate live-change input, target facets, compiled catalog, and candidate tests. | Candidate APIs are fail-closed and coexist only as a bounded transition. | DONE |
| WS-07A | Source contribution | WS-06 | Migrate Assessments. | Owner snapshot schemas, target semantics, rollback behavior, callers, and owner/Revisions slices pass. | DONE |
| WS-07B | Source contribution | WS-07A | Migrate Parties. | Same owner-slice exit criteria. | DONE |
| WS-07C | Source contribution | WS-07B | Migrate Artifacts. | Same owner-slice exit criteria. | DONE |
| WS-07D | Source contribution | WS-07C | Migrate Evidence. | Same owner-slice exit criteria. | DONE |
| WS-07E | Source contribution | WS-07D | Migrate Timeline. | Same owner-slice exit criteria. | DONE |
| WS-07F | Source contribution | WS-07E | Migrate Tasks/Decisions. | Same owner-slice exit criteria. | DONE |
| WS-07G | Source contribution | WS-07F | Migrate Links. | Same owner-slice exit criteria. | DONE |
| WS-07H | Source contribution | WS-07G | Migrate Indicators. | Same owner-slice exit criteria. | DONE |
| WS-07I | Source contribution | WS-07H | Migrate Entities. | Ten record families, nine owners, and fourteen target kinds form a complete candidate catalog; every revision caller uses typed captures. | DONE |
| WS-08 | Atomic cutover | WS-07I | Require canonical snapshots and separate event material; remove arbitrary/projection-first snapshot APIs. | Canonical snapshots are universal; old runtime paths are deleted; reset prerequisite is documented. | DONE |
| WS-09 | Atomic cutover | WS-08 | Add indexed association facts, generic history lookup, and deterministic bundle-import recomputation. | Arrays are complete/indexed; no target-specific JSON history predicate remains. | DONE |
| WS-10 | Atomic cutover | WS-09 | Activate provider-driven rollback and remove source-kind branches/special cases. | Revisions branches only on generic row/non-row dispatch; exact public rollback consequences pass. | DONE |
| WS-11 | Evidence reconciliation | WS-10 | Reconcile authored verification, test-family, boundary, ownership, contract-index, and migration-history inputs; regenerate through Make. | Harness, generated-artifact, JSON-shape, migration, and boundary checks pass. | DONE |
| WS-12 | Validation and handoff | WS-11 | Run focused, cross-owner, browser, drift, security, migration, finalization, broad, and release checks; close this ledger. | All required checks pass or a related failure is rooted and blocking; no implementation slice remains. | DONE |

## 7. Superseded Planning Slice Reference

The RS-00 through RS-05 material below is retained as historical design and
test discovery only. Section 6 is the controlling execution sequence. The
2026-08-09 user authorization supersedes the old separate-authorization and
parity-preservation gates wherever they conflict with the approved correction
and reset posture.

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
| RS-00 | Separate characterization authorization | Execute the closed snapshot, target-semantics, wire, projection, Collaboration, portability, security, and frontend matrices without production changes | Revisions and source-owner tests, fixtures/test support, applicable Timeline frontend/E2E, necessary authored verification input, this tracker | False baseline if behavior is inferred, old/new state differs, or unknown cells are guessed | All RMR-REQ-100 through RMR-REQ-215 characterization cases and RMR-AC-200 through RMR-AC-209 evidence | `make test-slice OWNER=module.revisions`; `make service-backed-test-slice OWNER=module.revisions`; applicable existing commands in Section 8 | Revert characterization-only additions that assert behavior not owned by adopted requirements; retain only valid evidence | Sections 4.3 and 4.8 contain proven values and run roots; every difference is classified; RB-001/RB-002 statuses are updated without overstating closure. |
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

WS-12 executed every applicable layer above plus security, migration, recovery,
and release verification. Section 10 records the exact final roots, corrected
intermediate failures, retained-run posture, and final worktree review.

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
| REM-000 | Execute the authorized remediation | WS-00 through WS-12 | DONE | RT-010 and user authorization dated 2026-08-09 | Sections 1.4, 6, 8, 10, 11, and 12 | Every workstream completed its mandatory checkpoint; WS-12 confirms that no implementation slice remains. |

The `IMP-*` rows are preserved planning history and are superseded by
`REM-000` plus the Section 6 ledger. Their old `BLOCKED` and `DEFERRED` states
must not be used as current execution gates.

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
| 2026-08-07 21:26 EDT | Codex / NLSpec-style tracker revision | Characterization matrices and binary evidence requirements are specified but unexecuted | Inspected analysis notes, NLSpec writing doctrine, current test/harness inventory, and Make guidance; touched only this tracker | `rg`, `sed`, `make lint-markdown`, inventory/traceability shell checks, `git diff --check` | Documentation checks passed; Markdown run root `.cartulary/test-results/20260808T012855Z-p2134753`; product tests were not run because only documentation changed | RB-003 prevents RS-00 test or harness changes | A separately authorized RS-00 task MUST record exact Make run roots and replace the open evidence cells. |

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

### Authorized remediation checkpoints

This table is append-only. It is the current per-workstream handoff ledger;
the earlier topical tables remain historical evidence.

| Time | Workstream | Status | Files changed | Substantive result and compatibility | Validation and run roots | Blockers, rollback, residual risk | Next workstream |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-09 15:40 EDT | WS-00 | DONE | `docs/handoffs/revisions-module-refactor-tracker.md` | Rebased the execution ledger to clean `b88ce7b5`; authorized canonical correction and pre-production reset replace stale parity and separate-authorization gates. No runtime compatibility changed. | `make lint-markdown` passed at `.cartulary/test-results/20260809T194403Z-p1578786`; `git diff --check`, status review, and diff review passed. | Rollback is this documentation edit; characterization evidence remains the only residual risk. | WS-01 characterization. |
| 2026-08-09 15:55 EDT | WS-01 | DONE | `internal/modules/revisions/delete_restore_test.go`, `internal/modules/revisions/history_components_test.go`, this tracker | Executed all ten authoritative source snapshots, proved projection-backed rows are a different shape, and completed exact current facts for fourteen target kinds. Stored snapshot correction is intentionally incompatible; retained public, rollback, portability, event, and security consequences remain unchanged. | `make format` passed at `.cartulary/test-results/20260809T195259Z-p1703196`; `make test-slice OWNER=module.revisions` passed 54/54 at `.cartulary/test-results/20260809T195408Z-p1733938`; `make service-backed-test-slice OWNER=module.revisions` passed 39/39 at `.cartulary/test-results/20260809T195315Z-p1706487`. Earlier evidence-test failures at `.cartulary/test-results/20260809T195014Z-p1636810` and `.cartulary/test-results/20260809T195137Z-p1675214` were corrected test setup assumptions. | Test-only rollback is removal of the two characterization cases; no runtime rollback needed. WS-02 must make the correction authoritative before production changes. | WS-02 specification adoption. |
| 2026-08-09 16:02 EDT | WS-02 | DONE | `docs/decisions/revisions-module-boundary.md`, Core 00/01/02/04, `docs/domain.md`, `contracts/revisions/*`, generated Revisions artifacts, this tracker | Adopted the internal boundary through REQ-00-071, canonical snapshot/target/history semantics through REQ-02-265, portability/reset rules through REQ-01-659, and exact AC-529 conformance. Added ten snapshot schema IDs and fourteen target registry entries. Public routes and bundle outer rows remain unchanged; schema-less history is intentionally rejected. | `make generate` passed at `.cartulary/test-results/20260809T200110Z-p1763330`; `make generate-drift` 4/4 at `.cartulary/test-results/20260809T200128Z-p1765902`; `make generated-artifact-policy-check` 3/3 at `.cartulary/test-results/20260809T200139Z-p1768705`; `make json-shape-check` 3/3 at `.cartulary/test-results/20260809T200143Z-p1769135`; `make lint-markdown` passed at `.cartulary/test-results/20260809T200149Z-p1769632`; Revisions slice 54/54 at `.cartulary/test-results/20260809T200201Z-p1770496`. | Rollback requires reverting the adopted decision, coordinated Core amendments, authored contracts, and generated projection together. Residual implementation risk is now bounded by the adopted catalog closure. | WS-03 Records envelope injection. |
| 2026-08-09 16:11 EDT | WS-03 | DONE | `internal/modules/revisions/application_ports.go`, `appender.go`, `httpapi/routes.go`; `internal/app/revisionassembly/revisions.go`; `internal/app/server/runtime_assembly.go`; Entities merge test composition; this tracker | Added narrow transaction-only and read-only envelope ports, injected the existing application adapter into Appender and HTTP composition, and removed all concrete Records imports/construction from Revisions production code. Constructor callers migrated in the slice with no shim. Public operations, visibility/error precedence, live event payloads, and transaction ownership remain unchanged. | `make format` passed at `.cartulary/test-results/20260809T200930Z-p1899643`; Revisions slice passed 54/54 at `.cartulary/test-results/20260809T200934Z-p1902808`; service-backed slice passed 39/39 at `.cartulary/test-results/20260809T201025Z-p1933231`; boundary check passed 3/3 at `.cartulary/test-results/20260809T201110Z-p1958342`; Markdown lint passed at `.cartulary/test-results/20260809T201226Z-p1959251`; static production search found no Records import or construction. An intermediate compile failure at `.cartulary/test-results/20260809T200839Z-p1869686` identified and removed two stale HTTP sentinel references. | Revert the ports, adapter factory, constructor signature, and callers together. Residual risk is limited to later snapshot/event separation, which retains this injected reader. No blocker remains. | WS-04 view-schema adapter relocation. |
| 2026-08-09 16:15 EDT | WS-04 | DONE | Removed `internal/modules/revisions/conflict_field_provider.go`; updated Revisions provider contributions, all source-owner contribution constructors, revision assembly and its catalog tests, and this tracker | Removed `ConflictFieldProvider` from the contribution contract. Revision assembly now resolves each declared view-schema ID, copies writable/value-kind/conflict-class facts, and builds the immutable resolver before serving. Runtime conflict-token and field semantics are unchanged; this is an internal cutover with no shim. | `make format` passed at `.cartulary/test-results/20260809T201408Z-p1961749`; Revisions slice passed 54/54 at `.cartulary/test-results/20260809T201411Z-p1964916`; boundary check passed 3/3 at `.cartulary/test-results/20260809T201508Z-p2002378`; static searches found no Revisions production import of `platform/viewschema` and no remaining provider symbol. | Revert assembly conversion, contribution shape, and deleted adapter together. Descriptor correctness remains owned by the view-schema registry and is snapshotted at startup; no ambient registry access survives at runtime. No blocker remains. | WS-05 explicit key-ring environment resolution. |
| 2026-08-09 16:20 EDT | WS-05 | DONE | `internal/modules/revisions/conflicts/key_ring.go`, conflict-token tests; `internal/app/server/runtime.go`, `runtime_assembly.go`, environment snapshot test; this tracker | Removed ambient environment reads from conflict-key parsing. Nil parser input now fails closed; a non-nil empty map remains empty even when the host contains a matching secret. Server runtime composition snapshots process entries once only when `Options.Env` is nil, preserving explicit injected maps. Environment names, key material format, rotation, fixture policy, and admitted startup behavior remain unchanged. | `make format` passed at `.cartulary/test-results/20260809T201826Z-p2005243`; Revisions slice passed 54/54 at `.cartulary/test-results/20260809T201829Z-p2008414`; server slice passed 43/43 at `.cartulary/test-results/20260809T201923Z-p2043540`; server service-backed slice passed 34/34 at `.cartulary/test-results/20260809T202002Z-p2073422`; boundary check passed 3/3 at `.cartulary/test-results/20260809T202031Z-p2095836`; static search found no production `os` use in Revisions. | Revert parser admission and server snapshot normalization together. Embedded callers that formerly passed nil directly must now supply an explicit map; this intentional internal migration removes host-secret ambiguity. No blocker remains. | WS-06 candidate typed substrate. |
| 2026-08-09 16:30 EDT | WS-06 | DONE | New `internal/modules/revisions/canonical_snapshot.go`, `target_semantics.go`, and candidate tests; updated Appender, provider contributions, revision assembly, and direct Entities test composition; this tracker | Added opaque schema-tagged snapshot capture, typed record-mutation/revision append APIs, and a separate `LiveRecordChange` input. Added pure direct/field history facets and an immutable catalog compiler that validates exact requirements, dispatch class, owner, record admission, fields, addressability, typed-nil providers, duplicates, unknowns, and missing entries. Revision assembly builds the partial capture catalog; no production caller uses the candidate append surface yet. | `make format` passed at `.cartulary/test-results/20260809T202757Z-p2099521`; Revisions slice passed 54/54 at `.cartulary/test-results/20260809T202803Z-p2102759`; service-backed slice passed 39/39 at `.cartulary/test-results/20260809T202909Z-p2139507`; boundary check passed 3/3 at `.cartulary/test-results/20260809T202954Z-p2164566`; `git diff --check` passed. | Revert the candidate files, Appender methods/constructor field, contribution fields, and assembly construction together. The intentionally temporary old/new append surfaces may coexist only through WS-07I and must be removed in WS-08; they are not fallback paths. No blocker remains. | WS-07A Assessments contribution. |
| 2026-08-09 16:37 EDT | WS-07A | DONE | Assessment contribution; workbook/import revision adapters; Assessment merge effects and assembly; Entities merge transitional carrier/dispatch and focused composition test; this tracker | Declared `cartulary.revisions.snapshot.assessment.v1` and the `assessment` row target. Assessment create and import now persist captured source snapshots while supplying projection rows only as live-change material. Cross-owner Entity merge captures Assessment source state before and after repoint and appends the typed row mutation; existing rollback provider, associations, protected set, ordering, and public results remain exact. New history is intentionally incompatible with pre-reset schema-less rows. | Assessment slice passed 26/26 at `.cartulary/test-results/20260809T203417Z-p2197399`; Assessment service-backed slice passed 19/19 at `.cartulary/test-results/20260809T203500Z-p2223550`; Revisions slice passed 54/54 at `.cartulary/test-results/20260809T203535Z-p2247088`; Revisions service-backed slice passed 39/39 at `.cartulary/test-results/20260809T203627Z-p2279747`; format passed at `.cartulary/test-results/20260809T203413Z-p2194226`; `git diff --check` passed. An intermediate compile failure at `.cartulary/test-results/20260809T203332Z-p2170537` caught a misplaced transitional field copy and was corrected. | Revert the schema declaration, create/import adapters, merge capture dependency, and typed carrier together. Generic delete/restore and rollback still use the old active path until WS-08/WS-10; no Assessment caller falls back between snapshot forms. No blocker remains. | WS-07B Parties contribution. |
| 2026-08-09 16:42 EDT | WS-07B | DONE | Party contribution, revision port, workbook create/reuse/patch flows, Party import path, shared captured import finalizer, and this tracker | Declared `cartulary.revisions.snapshot.party.v1`. Party create, reuse, patch, and import now persist only captured `{snapshot_schema_id, record, source}` values; projection rows remain response and live Collaboration material. Added an explicit captured import finalizer beside the bounded old transition function so unmigrated owners do not receive an implicit behavior change. Public Party results, conflict handling, idempotency, and rollback provider behavior remain exact. | Party slice passed 14/14 at `.cartulary/test-results/20260809T203953Z-p2334302`; Party service-backed slice passed 14/14 at `.cartulary/test-results/20260809T204030Z-p2359474`; Revisions slice passed 54/54 at `.cartulary/test-results/20260809T204101Z-p2381874`; Revisions service-backed slice passed 39/39 at `.cartulary/test-results/20260809T204154Z-p2416656`; format passed at `.cartulary/test-results/20260809T203950Z-p2331131`. An intermediate compile failure at `.cartulary/test-results/20260809T203845Z-p2309339` exposed the shared import interface and led to the explicit captured finalizer. | Revert Party schema/port/callers and the captured finalizer together. The old shared finalizer remains only for unmigrated owners and is deleted in WS-08 after their slices. No Party snapshot fallback exists. | WS-07C Artifacts contribution. |
| 2026-08-09 16:53 EDT | WS-07C | DONE | Artifact contribution, source snapshot, revision port, mutation kernel, workbook/import flows, source-owned conflict mapping, generic canonical conflict projector, Party conflict reconstruction, and this tracker | Declared `cartulary.revisions.snapshot.artifact.v1`; Artifact create, patch, and import now persist captured canonical snapshots and supply projection rows only as live-change material. The Artifact aggregate snapshot includes base and subtype-owned source facts but excludes associations and derived projection facts. Conflict windows now reconstruct writable cells from canonical snapshots through immutable source-owner mappings, preserving conflict-token v3 and same-field behavior without retaining projections. Record-link mutations remain separately revisioned. | Format passed at `.cartulary/test-results/20260809T205018Z-p2452166`; Artifact slice passed 11/11 at `.cartulary/test-results/20260809T205021Z-p2455344`; Artifact service-backed slice passed 7/7 at `.cartulary/test-results/20260809T205039Z-p2459034`; Party slice passed 14/14 at `.cartulary/test-results/20260809T205041Z-p2460294`; Party service-backed slice passed 14/14 at `.cartulary/test-results/20260809T205115Z-p2483808`; Revisions slice passed 54/54 at `.cartulary/test-results/20260809T205148Z-p2506298`; Revisions service-backed slice passed 39/39 at `.cartulary/test-results/20260809T205242Z-p2540325`. An intermediate Artifact failure at `.cartulary/test-results/20260809T204410Z-p2446410` exposed projection-shaped conflict-window coupling and drove the canonical reconstruction seam. | Revert the Artifact schema/callers and canonical conflict projector together. Schema-less or projection-shaped snapshots have no canonical-projector fallback; the temporary legacy conflict helper remains only for unmigrated Evidence and Tasks/Decisions and is deleted in WS-08. Collection associations still require their WS-07G/WS-09 semantics. No blocker remains. | WS-07D Evidence contribution. |
| 2026-08-09 17:00 EDT | WS-07D | DONE | Evidence contribution, revision port, mutation coordinator and test, workbook/import flows, blob attach/quarantine flows, conflict projector, and this tracker | Declared `cartulary.revisions.snapshot.evidence.v1`. All Evidence revision-producing paths capture authoritative source state in the caller transaction and persist typed canonical mutations/revisions; projection rows are limited to responses, changed-field calculation, and live Collaboration intents. Multi-record quarantine captures the complete before set before applying changes and appends deterministic per-record canonical pairs. Same-field conflict behavior is reconstructed from source-owned scalar mappings. | Format passed at `.cartulary/test-results/20260809T205609Z-p2567564`; Evidence slice passed 57/57 at `.cartulary/test-results/20260809T205612Z-p2570742`; Evidence service-backed slice passed 46/46 at `.cartulary/test-results/20260809T205722Z-p2605060`; Revisions slice passed 54/54 at `.cartulary/test-results/20260809T205817Z-p2632592`; Revisions service-backed slice passed 39/39 at `.cartulary/test-results/20260809T205908Z-p2664715`; static search found no old Evidence append/finalizer/conflict-window calls; `git diff --check` passed. | Revert the Evidence schema, port, callers, and projector together. Custody events remain separately owned source facts and are not folded into row snapshots. Schema-less history has no fallback. No blocker remains. | WS-07E Timeline contribution. |
| 2026-08-09 17:22 EDT | WS-07E | DONE | Timeline contribution and revision port; create, patch, clipboard-paste, and lifecycle paths; revision-window change-set identity; canonical conflict projector; Timeline snapshot assertions; this tracker | Declared `cartulary.revisions.snapshot.timeline_event.v1`. Every Timeline row revision path now captures authoritative envelope/source state and persists typed canonical mutation/revision pairs while retaining hydrated projection rows only for responses and live Collaboration consequences. Separately revisioned links, tags, and mentions remain outside row snapshots; Timeline conflict handling combines canonical scalar reconstruction with owner mutation/source-table facts for collection conflicts. Existing selectors, conflict-token v3 fields, collection values, event effects, supersedes behavior, and rollback registrations remain exact. | Format passed at `.cartulary/test-results/20260809T211811Z-p2819773`; focused collection-conflict row passed 3/3 at `.cartulary/test-results/20260809T211820Z-p2823032`; Timeline slice passed 68/68 at `.cartulary/test-results/20260809T211835Z-p2825202`; Timeline service-backed slice passed 46/46 at `.cartulary/test-results/20260809T211949Z-p2857524`; Revisions slice passed 54/54 at `.cartulary/test-results/20260809T212050Z-p2884960`; Revisions service-backed slice passed 39/39 at `.cartulary/test-results/20260809T212148Z-p2919264`; Markdown lint passed at `.cartulary/test-results/20260809T212316Z-p2944861`; static search found no old Timeline row-revision or projection conflict-window call. Intermediate failures at `.cartulary/test-results/20260809T210351Z-p2695809`, `.cartulary/test-results/20260809T211151Z-p2737693`, `.cartulary/test-results/20260809T211425Z-p2778444`, and `.cartulary/test-results/20260809T211725Z-p2817053` exposed projection-shaped assertions, UUID/text parameter inference, collection-change discovery, and the exact empty collection value shape; each was corrected before closure. | Revert the Timeline schema declaration, revision-port methods, typed callers, revision-window change-set field, projector, and collection-conflict fact derivation together. Separately revisioned associations are intentionally not folded into canonical row history. Schema-less history has no fallback and requires the approved reset. No blocker remains. | WS-07F Tasks/Decisions contribution. |
| 2026-08-09 17:37 EDT | WS-07F | DONE | Tasks/Decisions contribution; workbook create, patch, conflict, and supersession flows; import finalization; immutable conflict projectors; Links mutation-value adapter; this tracker | Declared `cartulary.revisions.snapshot.task_request.v1` and `cartulary.revisions.snapshot.decision.v1`. Task Request and Decision row callers now persist captured canonical snapshots and send projection rows only as live-change material. Scalar conflict windows reconstruct from source-owned mappings, and keep-saved admission uses the immutable resolver rather than the global view registry. Workbook relationship collections and the Task-to-Decision direct-reference link now append deterministic `record_link` mutations; supersession retains its separately revisioned link and exact two-record event ordering. Imports use captured finalization. | Format passed at `.cartulary/test-results/20260809T213322Z-p3008707`; Tasks/Decisions slice passed 20/20 at `.cartulary/test-results/20260809T213332Z-p3011967`; service-backed slice passed 17/17 at `.cartulary/test-results/20260809T213410Z-p3037477`; Links slice passed 14/14 at `.cartulary/test-results/20260809T213444Z-p3060073`; Links service-backed slice passed 14/14 at `.cartulary/test-results/20260809T213514Z-p3084127`; Revisions slice passed 54/54 at `.cartulary/test-results/20260809T213550Z-p3106694`; Revisions service-backed slice passed 39/39 at `.cartulary/test-results/20260809T213649Z-p3139689`; static search found no old Tasks/Decisions row-revision, projection conflict-window, or legacy import-finalizer call. | Revert the two schema declarations, captured caller changes, conflict projectors, immutable resolver use, and Links mutation-value wrapper together. Schema-less history has no fallback. Import row sequencing still admits one primary row mutation per input row; WS-07G owns any additional imported link-mutation allocation required by the Links target semantics. No blocker remains for the typed row contribution. | WS-07G Links contribution. |
| 2026-08-09 17:49 EDT | WS-07G | DONE | Links provider contribution and facet test; record/party/tag collection mutation-value APIs; Task direct-reference import sequencing; Artifact collection callers and linked-note evidence; this tracker | `record_link` now contributes the canonical source/destination association facet and `record_tag` contributes the owning-record facet, both individually addressable and bound to the existing Links rollback provider. Workbook Task/Decision and Artifact collection writes persist deterministic full Links-owned mutation values; Task imports allocate a contiguous owner-effect range through the existing import sequencer. Projection associations remain outside row snapshots. Public rows, link/tag item references, rollback targets, selectors, and live events remain unchanged; newly complete internal change sets contain additional non-row entries where projection snapshots formerly hid relationship changes. | Format passed at `.cartulary/test-results/20260809T214438Z-p3205666`; Tasks/Decisions slice passed 20/20 at `.cartulary/test-results/20260809T214445Z-p3208916`; Artifact slice passed 11/11 at `.cartulary/test-results/20260809T214527Z-p3234098`; Links slice passed 14/14 at `.cartulary/test-results/20260809T214542Z-p3236863`; Links service-backed slice passed 14/14 at `.cartulary/test-results/20260809T214616Z-p3260555`; Artifact service-backed slice passed 7/7 at `.cartulary/test-results/20260809T214649Z-p3283131`; Tasks/Decisions service-backed slice passed 17/17 at `.cartulary/test-results/20260809T214658Z-p3284464`; Revisions slice passed 54/54 at `.cartulary/test-results/20260809T214735Z-p3306977`; Revisions service-backed slice passed 39/39 at `.cartulary/test-results/20260809T214832Z-p3339868`. The intermediate Artifact failure at `.cartulary/test-results/20260809T214224Z-p3195696` correctly exposed the newly explicit tag mutation in linked-note accounting; the test now proves one row, one contextual link, and one tag entry. | Revert the Links history facets, mutation-value adapters, caller result plumbing, import range allocation, and updated linked-note assertion together. Old schema-less/projection-folded relationship history is intentionally unsupported. Collection-only changes now carry explicit rollback/addressability facts and no unresolved Links producer fallback remains. | WS-07H Indicators contribution. |
| 2026-08-09 17:56 EDT | WS-07H | DONE | Indicators contribution and facet test; revision port; create/reuse/import, observation, and lifecycle callers; affected-record snapshot coordination and atomicity fake; this tracker | Declared `cartulary.revisions.snapshot.indicator.v1` and the direct `indicator` row target. Indicator row mutations and revisions now persist typed authoritative snapshots while projection rows remain live-change and response material. Observation and state-interval entries retain their source-owned non-row values and now contribute exact, sorted history associations; every affected record is captured before and after the child change, including cross-owner observation sources. Public operations, replay behavior, affected versions, live events, and existing rollback providers remain exact. | Format passed at `.cartulary/test-results/20260809T215343Z-p3368007`; Indicators slice passed 27/27 at `.cartulary/test-results/20260809T215350Z-p3371299`; Indicators service-backed slice passed 13/13 at `.cartulary/test-results/20260809T215411Z-p3378751`; Revisions slice passed 54/54 at `.cartulary/test-results/20260809T215425Z-p3380258`; Revisions service-backed slice passed 39/39 at `.cartulary/test-results/20260809T215519Z-p3412885`; static search found no old Indicator row-revision or legacy import-finalizer call. | Revert the Indicator schema/facets, revision port, typed callers, capture coordination, and atomicity fake together. During the bounded transition, an observation sourced from an as-yet-unmigrated Entity type cannot capture that source; WS-07I closes the remaining catalog gap. Schema-less history has no fallback. No blocker remains. | WS-07I Entities contribution. |
| 2026-08-09 18:17 EDT | WS-07I | DONE | Entities contribution and facet test; Host/Identity create, import, clipboard, patch, merge, and Mention-to-Timeline callers; merge/mention revision ports; candidate closure test; Links/Entities Timeline history-fact ports; this tracker | Declared canonical Host and Identity snapshot schemas and direct row targets. All Host/Identity row callers, both merge participants, and cross-owner Timeline mention effects now persist typed source snapshots; projection values remain live responses or separately published intents. Entity mention, alias, and preserved-identifier facets supply exact association and addressability rules. The live contributions now close ten record families, nine owners, and fourteen target kinds. Timeline conflict history obtains Links and mention facts through source-owner ports instead of reading foreign tables. Public create/reuse/patch/merge/mention results, ordering, invalidations, replay, and rollback providers remain exact. | Format passed at `.cartulary/test-results/20260809T221422Z-p3643679`; focused mention lifecycle passed 3/3 at `.cartulary/test-results/20260809T220510Z-p3483505`; Entities slice passed 45/45 at `.cartulary/test-results/20260809T220526Z-p3485030`; Entities service-backed slice passed 42/42 at `.cartulary/test-results/20260809T220652Z-p3517397`; Revisions slice passed 54/54 at `.cartulary/test-results/20260809T221057Z-p3613070`; Revisions service-backed slice passed 39/39 at `.cartulary/test-results/20260809T220934Z-p3584403`; Timeline slice passed 68/68 at `.cartulary/test-results/20260809T221431Z-p3647325`; Timeline service-backed slice passed 46/46 at `.cartulary/test-results/20260809T221543Z-p3683820`; backend boundary passed 3/3 at `.cartulary/test-results/20260809T221426Z-p3646937`. The intermediate Entities failure at `.cartulary/test-results/20260809T220224Z-p3445502` exposed a projection-equality assertion and was replaced by canonical schema/identity/version evidence. The intermediate boundary failure at `.cartulary/test-results/20260809T221146Z-p3641824` exposed foreign Links reads and a comment-parser collision; source-owner ports and unambiguous wording closed both. | Revert the two Entity schemas/facets, typed caller and port changes, canonical lifecycle assertion, closure test, and source-owner Timeline fact ports together. No production source owner retains an arbitrary row-revision append or legacy import finalizer; generic delete/restore and rollback are the only old Revisions paths and are removed in WS-08/WS-10. No blocker remains. | WS-08 canonical snapshot/event cutover. |
| 2026-08-09 18:51 EDT | WS-08 | DONE | Revisions Appender, canonical snapshot validation, live-record port, delete/restore, rollback publication/application, Incident Bundle validation and fixtures; projection/revision/server assembly; source-owner rollback readers and tests; import owner facade; this tracker | Canonical `{snapshot_schema_id, record, source}` captures are now the only row-history write and read shape. Removed arbitrary row maps, legacy row append/finalizer APIs, projection-first history reads, and source-provider projection/direct fallbacks. `ProjectionServices` is rebuild-only; explicitly separate `LiveRecordReader` material preserves Collaboration changed fields and events without entering history. Delete, restore, rollback, and bundle import fail closed on schema-less history after existing authorization checks. Current version-1/version-2 bundles remain supported when canonical; routes, opaque selectors, conflict-token v3, rollback results, event ordering, and public row shapes remain exact. | Final format passed at `.cartulary/test-results/20260809T224608Z-p3953818`; Revisions slice passed 54/54 at `.cartulary/test-results/20260809T224238Z-p3918980`; service-backed slice passed 39/39 at `.cartulary/test-results/20260809T224616Z-p3957083`; fast suite passed 353/353 at `.cartulary/test-results/20260809T224708Z-p3984381`; backend boundary passed 3/3 at `.cartulary/test-results/20260809T224936Z-p4051738`; Markdown lint passed at `.cartulary/test-results/20260809T225141Z-p4052989`; static searches and `git diff --check` passed. Corrected intermediate compile/fixture/live-material failures are rooted at `.cartulary/test-results/20260809T222617Z-p3720437`, `.cartulary/test-results/20260809T222709Z-p3748318`, `.cartulary/test-results/20260809T222806Z-p3780168`, `.cartulary/test-results/20260809T222914Z-p3820389`, `.cartulary/test-results/20260809T223704Z-p3894020`, `.cartulary/test-results/20260809T223739Z-p3899038`, and `.cartulary/test-results/20260809T223858Z-p3907420`. | Revert the cutover as one unit; no legacy reader or runtime fallback is retained. A destructive local database reset was not executed: isolated service-backed databases passed, while any persistent database containing old mutation rows remains reset-required and WS-09 will enforce that admission. Remaining storage risk is the absent indexed association columns. RB-001 is closed. | WS-09 indexed history-semantics cutover. |
| 2026-08-09 19:33 EDT | WS-09 | DONE | Migration `00061`; target-semantics contract loader/catalog; Revisions Appender, history repository/materializer, Incident Bundle validation/import, application assembly; history, rollback, portability, migration, and bundle fixtures; Evidence contribution; this tracker | Every mutation write now derives sorted, unique `history_record_ids` and `history_entry_record_ids` from the immutable target catalog before persistence. History lookup uses only the GIN-indexed association array and persisted addressability; generic Revisions SQL/materialization no longer interprets source JSON keys. Bundle import recomputes the facts from admitted semantics while export retains the exact portable row members. Migration 61 succeeds on an empty database and fails before schema change with an explicit reset-required diagnostic when legacy mutation rows exist. Public history ordering, item/entry references, actions, bundle versions, and live consequences remain exact. | Final format passed at `.cartulary/test-results/20260809T233253Z-p183118`; Revisions slice passed 54/54 at `.cartulary/test-results/20260809T232909Z-p124505`; service-backed slice passed 39/39 at `.cartulary/test-results/20260809T233003Z-p152865`; focused catalog/history/portability rows passed 6/6 at `.cartulary/test-results/20260809T233306Z-p186795`; backend boundary passed 3/3 at `.cartulary/test-results/20260809T233300Z-p186402`; the focused repository/migration path passed 3/3 at `.cartulary/test-results/20260809T232232Z-p47307`; static source-vocabulary/JSON-predicate search and `git diff --check` passed. Corrected intermediate compile, Goose statement-boundary, SQL parameter-type, fixture-addressability, missing Evidence contribution, and boundary-placement failures are rooted at `.cartulary/test-results/20260809T230601Z-p4083960`, `.cartulary/test-results/20260809T231658Z-p4169441`, `.cartulary/test-results/20260809T232116Z-p13266`, `.cartulary/test-results/20260809T232301Z-p48872`, `.cartulary/test-results/20260809T232523Z-p80138`, `.cartulary/test-results/20260809T232802Z-p116953`, and `.cartulary/test-results/20260809T233136Z-p182014`. | Revert migration, catalog loading, derived writes, generic query, and import recomputation together. No reset, backfill, inference, or destructive developer-database action was performed. `make migration-drift` now reaches the intentional migration-history manifest mismatch at `.cartulary/test-results/20260809T231755Z-p4189328`; WS-11 owns that authored manifest/test-routing reconciliation after WS-10 fixes the final topology. No product blocker remains for WS-10. | WS-10 provider-driven rollback cutover. |
| 2026-08-09 20:04 EDT | WS-10 | DONE | Unified target-semantics and snapshot-registry catalogs; Revisions rollback contract, planner, query repository, applier, command composition, and portability validation; Links attached-evidence and Entities mention providers; application revision assembly; catalog, rollback, and isolated owner tests; removed row/non-row rollback catalogs; this tracker | One immutable catalog now supplies exact target lookup, row/non-row dispatch, admitted record providers, owner non-row providers, history facets, and default row restoration targets. Generic Revisions rollback branches only on `row` versus `non_row`; it loads siblings generically while owner providers select companions and whole-set coupling. The applier invokes catalog-selected providers and no longer contains source-kind apply paths. Handwritten record/target owner maps were removed in favor of generated snapshot/target registry closure. Current tokens, opaque selectors, companion order, affected/protected records, lock order, errors, results, events, and version identifiers remain exact; no old runtime dispatch path or compatibility shim remains. | Final format passed at `.cartulary/test-results/20260810T000235Z-p560494`; server build passed at `.cartulary/test-results/20260809T235255Z-p279666`; Revisions slice passed 54/54 at `.cartulary/test-results/20260809T235308Z-p291296`; service-backed Revisions passed 39/39 at `.cartulary/test-results/20260809T235418Z-p327995`; source-owner slices passed: Assessments 26/26 at `.cartulary/test-results/20260809T235522Z-p353592`, Parties 14/14 at `.cartulary/test-results/20260809T235558Z-p380048`, Artifacts 11/11 at `.cartulary/test-results/20260809T235629Z-p402786`, Evidence 57/57 at `.cartulary/test-results/20260809T235639Z-p405493`, Timeline 68/68 at `.cartulary/test-results/20260809T235742Z-p439334`, Tasks/Decisions 20/20 at `.cartulary/test-results/20260809T235851Z-p471236`, Links 14/14 at `.cartulary/test-results/20260809T235925Z-p495742`, Indicators 27/27 at `.cartulary/test-results/20260809T235955Z-p519786`, and Entities 45/45 at `.cartulary/test-results/20260810T000239Z-p563705`; backend boundary passed 3/3 at `.cartulary/test-results/20260810T000415Z-p598157`; Markdown lint passed at `.cartulary/test-results/20260810T000552Z-p599252`. Static searches found no old catalog symbols, source-kind rollback cases, handwritten owner maps, or source target literals in generic production coordination. The initial Entities run failed 44/45 at `.cartulary/test-results/20260810T000014Z-p526923` because an isolated Assessments-only test used the exact production snapshot-catalog constructor; an explicit-requirements candidate constructor corrected the fixture without weakening production closure. | Revert the unified catalog, generic coordinator, owner companion facets, and composition changes together; deleted source dispatch files cannot be restored independently. No public or persisted compatibility change is introduced beyond the already-authorized canonical/reset posture. Final browser, generated, harness, migration-manifest, and broad validation remain intentionally assigned to WS-11/WS-12. RB-002 implementation closure is complete. | WS-11 verification and harness reconciliation. |
| 2026-08-09 20:18 EDT | WS-11 | DONE | Revisions verification contract and 44-row family; App Server, Entities, Indicators, and Links family manifests; migration-history manifest; Make-generated execution-topology render index; renamed target-semantics catalog tests; migration assertion; this tracker | Reconciled the final topology without making evidence authoritative at runtime. Every newly authored test now has an exact selectable row; the unified catalog tests no longer carry retired row/non-row catalog names. Added dedicated migration-61 verification for fresh admission, canonical array/index facts, reset-required rejection, and failure-before-schema-change. Source-owner contribution and explicit server-environment tests execute in their owning slices. Migration 61 is recorded with its exact checksum. Generated files were changed only by `make generate`; no boundary or schema-owner allowlist expansion was needed. No product compatibility changed. | `make generate` passed at `.cartulary/test-results/20260810T001302Z-p614460`; generate drift passed 4/4 at `.cartulary/test-results/20260810T001321Z-p617010`; generated-artifact policy passed 3/3 at `.cartulary/test-results/20260810T001332Z-p619807`; JSON shape passed 3/3 at `.cartulary/test-results/20260810T001338Z-p620276`; migration drift passed 5/5 at `.cartulary/test-results/20260810T001346Z-p620747`; Revisions passed 55/55 at `.cartulary/test-results/20260810T001555Z-p668652` and service-backed 40/40 at `.cartulary/test-results/20260810T001647Z-p697243`; focused migration passed 3/3 at `.cartulary/test-results/20260810T001512Z-p662388`; new App Server, Entities, Indicators, and Links rows each passed 1/1 at `.cartulary/test-results/20260810T001533Z-p664791`, `.cartulary/test-results/20260810T001538Z-p665746`, `.cartulary/test-results/20260810T001542Z-p666493`, and `.cartulary/test-results/20260810T001547Z-p667499`; harness contract passed 2/2 at `.cartulary/test-results/20260810T001737Z-p722524`; backend boundary passed 3/3 at `.cartulary/test-results/20260810T001744Z-p723076`; format passed at `.cartulary/test-results/20260810T001754Z-p723550`. Initial generation failures at `.cartulary/test-results/20260810T001012Z-p605620`, `.cartulary/test-results/20260810T001044Z-p608361`, and `.cartulary/test-results/20260810T001119Z-p611057` exposed authored verification ordering, cross-package selector placement, and selector ordering and were corrected. The first routed Revisions run failed 54/55 at `.cartulary/test-results/20260810T001357Z-p623540` because the new migration assertion used `pg_am.name`; changing it to `pg_am.amname` closed the dormant test defect. | Revert authored verification/family rows, migration manifest entry, and generated topology index together; do not retain a selector for a removed test. No product blocker remains. A destructive developer-database reset was not run; persistent pre-remediation databases with existing mutation rows still require the documented reset. WS-12 owns browser, security, finalization, broad, release, and final worktree/handoff evidence. | WS-12 validation and handoff completion. |
| 2026-08-09 21:51 EDT | WS-12 | DONE | Final Revisions, conflict, workbook, Incident Bundle, migration, recovery, specification, contract, generated, ownership, boundary, and harness corrections; this tracker | Validation found that canonical source snapshots alone cannot reconstruct collection-valued conflict cells. Added normalized immutable per-field conflict facts to migration 61 and the append/window path, while keeping them outside row history, rollback, projection, and portable bundle authority. Recovery, schema ownership, migration history, Core 01/02/04, AC-529, and generated projections now carry that final table. Canonical snapshots, indexed generic history, provider rollback, explicit environment resolution, and immutable assembly catalogs remain the only runtime paths. Public routes, selectors, authorization order, history items, rollback results, Collaboration consequences, conflict-token v3, and canonical bundle versions remain exact. | Final Revisions passed 55/55 at `.cartulary/test-results/20260810T010435Z-p1284719`; browser webserver-backed passed 62/62 at `.cartulary/test-results/20260810T001952Z-p731068` and stateful passed 36/36 at `.cartulary/test-results/20260810T002355Z-p784899`; vulnerability and targeted security checks passed 4/4 at `.cartulary/test-results/20260810T002554Z-p811479` and 4/4 at `.cartulary/test-results/20260810T002630Z-p812418`; generation passed at `.cartulary/test-results/20260810T010223Z-p1233658`, drift passed 4/4 at `.cartulary/test-results/20260810T010602Z-p1320408`, generated policy passed 3/3 at `.cartulary/test-results/20260810T010616Z-p1323217`, JSON shape passed 3/3 at `.cartulary/test-results/20260810T010726Z-p1325396`, migration drift passed 5/5 at `.cartulary/test-results/20260810T010734Z-p1325867`, and backend boundary passed 3/3 at `.cartulary/test-results/20260810T010817Z-p1329494`; finalization passed at `.cartulary/test-results/20260810T010858Z-p1330891`, fast tests passed 353/353 at `.cartulary/test-results/20260810T010912Z-p1333628`, the clean broad check passed 748/748 at `.cartulary/test-results/20260810T011809Z-p1553251`, and the clean release check passed 919/919 at `.cartulary/test-results/20260810T014140Z-p2023969`. Markdown lint passed at `.cartulary/test-results/20260810T015708Z-p2197464`; static acceptance searches and Go lint passed. | No product blocker remains. The implementation-related first broad failure at `.cartulary/test-results/20260810T002932Z-p910485` drove the conflict-fact and recovery corrections. Later Jobs timing and object-store/browser transients at `.cartulary/test-results/20260810T011128Z-p1399200`, `.cartulary/test-results/20260810T012008Z-p1632506`, and `.cartulary/test-results/20260810T013124Z-p1830516` were isolated green and superseded by clean full runs. No destructive database reset was executed; operators must reset any persistent pre-remediation database containing mutations. Schema-less snapshots and bundles remain intentionally unsupported. `RESULTS_DIR` was unset, so retained successful-run maintenance was skipped. The intended remediation change set is the only worktree delta. | None; remediation and handoff are complete, and no further implementation slice remains. |

## 11. Open Questions and Blockers

No `BLOCKED: owner contradiction` entry is required because the inspected owner
documents agree. The following historical items are resolved or converted into
active workstream gates by the approved remediation posture.

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Authoritative source snapshots replace projection-derived/schema-less history. | Public consequences could drift even though storage correction is intentional. | WS-01 characterization, adopted owner amendments in WS-02, and cutover evidence in WS-08. | `CLOSED`; WS-08 completed the authorized correction and retained-consequence validation without a legacy path. |
| RB-002 | Immutable source-contributed target semantics replace central source vocabulary. | An incomplete matrix could change associations, selectors, companions, protected records, or inverse ordering. | Complete the fourteen-token matrix in WS-01, adopt it in WS-02, and prove exact closure through WS-07I and WS-10. | `CLOSED`; WS-09 completed generic indexed history semantics and WS-10 completed catalog-driven dispatch, owner companion selection, protected/affected sets, inverse ordering, and old-path deletion. |
| RB-003 | The former documentation-only gate. | It previously prohibited all implementation. | User authorization dated 2026-08-09, with Section 1.4 checkpoints. | `CLOSED / SUPERSEDED`. |

### 11.1 Binary blocker closure

| ID | Blocker | Closure condition |
| --- | --- | --- |
| RMR-AC-100 | RB-001 | Every Section 4.3 matrix cell is executed against old and candidate readers in the same transaction; every Section 4.4 layer passes; every difference is classified; outcome A, B, or C is recorded with run roots. |
| RMR-AC-101 | RB-001 | Outcome A is the only behavior-preserving closure. It requires no difference except `representation_only`, no fallback machinery, and explicit evidence for every public and portability consequence. |
| RMR-AC-102 | RB-001 | Outcome B remains blocked until adopted owners authorize the difference and define migration, versioning, compatibility, and any `snapshot_schema_id` behavior. |
| RMR-AC-103 | RB-001 | Outcome C remains blocked until the defective owner adapter is repaired separately and the complete matrix is rerun. |
| RMR-AC-210 | RB-002 | Every formerly open Section 4.8 cell is replaced by cited live evidence, the 14-token set is proven complete, and no source behavior is inferred. |
| RMR-AC-211 | RB-002 | RMR-AC-200 through RMR-AC-209 all pass with exact run roots and static-search evidence. |
| RMR-AC-300 | RB-003 | PASS: Section 1 records the baseline, allowed scope, posture, exclusions, checkpoints, and handoff requirements. |
| RMR-AC-301 | RB-003 | PASS: Section 6 keeps characterization, specification, owner migrations, cutovers, and completion as serial workstreams with independent checkpoints. |
| RMR-AC-302 | RB-003 | PASS: Section 1.4 enumerates the authorized storage correction, reset posture, rejected legacy shapes, and retained public consequences. |

## 12. Binary Completion Criteria

Tracker completion now requires implementation closure. WS-12 cannot be marked
`DONE` while a required slice, blocker, validation result, reset instruction,
or residual risk is missing.

### 12.1 Requirement traceability

| Requirement or criterion range | Controlling owner | Workstream or slice | Required evidence | Current state |
| --- | --- | --- | --- | --- |
| RMR-REQ-001 through RMR-REQ-010 | Core 00-04 and Testing Harness owner within their scopes | All workflows and slices | Contract freeze map, owner-selected tests, boundaries, generated policy | PASS; final focused, boundary, generated, broad, and release evidence is recorded in the WS-12 handoff row. |
| RMR-REQ-100 through RMR-REQ-112 | Core 02 REQ-02-265, plus Core 01/03/04 for external consequences | WS-01, WS-02, WS-06 through WS-09 | Sections 4.3-4.6, history/rollback/conflict/portability/browser evidence | PASS; canonical snapshots and separate normalized conflict facts passed focused, browser, broad, and release verification. |
| RMR-REQ-200 through RMR-REQ-215 | Core 01 REQ-01-659, Core 02 REQ-02-265, Core 04 AC-529, and source owners | WS-01, WS-02, WS-06 through WS-10 | Section 4.8 and RMR-AC-200 through RMR-AC-209 | PASS; fourteen target kinds, indexed association facts, and provider-driven rollback passed catalog, owner, broad, and release verification. |
| RMR-AC-100 through RMR-AC-103 | Same as snapshot requirements | WS-01, WS-02, and WS-08 | Same-state old/new comparisons and classified outcomes | PASS through authorized outcome B: WS-01 proved the shape defect, WS-02 adopted canonical correction/reset semantics, and WS-08 removed legacy paths. |
| RMR-AC-200 through RMR-AC-211 | Same as target-semantics requirements | WS-01, WS-06 through WS-10, and WS-12 | Exact matrix, catalog negatives, equivalence, failure, security, static, and final system evidence | PASS; owner and Revisions slices plus clean 748-unit broad and 919-unit release graphs confirm closure. |
| RMR-AC-300 through RMR-AC-302 | User task authorization under adopted owners | WS-00 and every workstream checkpoint | Section 1 authorization and Section 6 sequencing | PASS; WS-00 through WS-12 completed serial checkpoints and no implementation slice remains. |

### 12.2 Tracker-document completion

| ID | Criterion | Status | Evidence |
| --- | --- | --- | --- |
| RMR-AC-001 | Every file in `internal/modules/revisions` is inventoried or explicitly out of scope. | PASS | Section 2 contains the exact 73-file baseline and Section 2.1 records five removals plus seven additions; the completed target has exactly 75 files. |
| RMR-AC-002 | Every discovered public contract risk has an owner and test posture. | PASS | Section 4 maps HTTP, authorization, history, source providers, projection, WebSocket, conflicts, portability, recovery, frontend, generated, and harness surfaces. |
| RMR-AC-003 | Every proposed workflow has dependencies and exit criteria. | PASS | Section 6 defines WS-00 through WS-12 with serial dependencies, exit criteria, and mandatory checkpoints. |
| RMR-AC-004 | Every preserved behavior or authorized correction has an explicit compatibility posture. | PASS | Section 1.4 freezes valuable external consequences and authorizes canonical-history correction plus pre-production reset. |
| RMR-AC-005 | Validation commands are discovered or marked unresolved with a reason. | PASS | Section 8 records Make-owned focused, browser, drift, boundary, and broader commands. |
| RMR-AC-006 | Contradictions are marked `BLOCKED: owner contradiction`. | PASS | No owner contradiction was found; repository deviations are findings and Section 11 defines the mandatory conflict posture. |
| RMR-AC-007 | Repository/framework mismatches are recorded as planning findings. | PASS | Sections 1, 3, and 5 distinguish the 73-file live decomposition from generic and archived evidence. |
| RMR-AC-008 | Snapshot sourcing has a complete decision rule, explicit defaults, closed matrices, difference taxonomy, and binary outcomes without claiming parity. | PASS | Sections 4.2-4.6 and RB-001. |
| RMR-AC-009 | Target semantics has an unambiguous two-facet interface, closed ownership, fail-closed defaults, exhaustive current-token set, prohibited alternatives, and binary tests without guessing open evidence. | PASS | Sections 4.7-4.9 and RB-002. |
| RMR-AC-010 | Characterization and production have separate, complete authorization templates and stop conditions. | PASS | Sections 7.1, 7.2, and RB-003. |
| RMR-AC-011 | Handoff sections preserve prior history and contain a current row for all seven workstream-specific tables. | PASS | Section 10 retains the seven prior rows and adds exactly one current-session row to each table. |
| RMR-AC-012 | Markdown lint, diff hygiene, exact inventory comparison, requirement traceability, and worktree-scope checks pass. | PASS | WS-12 reran Markdown lint after completing this ledger; exact inventory, static acceptance searches, `git diff --check`, diff review, and final status review passed with only the intended remediation change set present. |

The remediation is complete. Persistent databases containing pre-remediation
mutation rows require the documented reset before startup; no destructive reset
was performed during this work. No legacy reader, translator, fallback, shadow
path, or historical-shape inference remains. No further implementation
workstream remains.
