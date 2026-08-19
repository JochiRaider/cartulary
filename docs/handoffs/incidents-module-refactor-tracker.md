# incidents Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

- **Target path:** `internal/modules/incidents`
- **Target label:** `incidents` (derived from the target path and normalized to lowercase kebab case)
- **Output path:** `docs/handoffs/incidents-module-refactor-tracker.md`
- **Repository snapshot:** commit `6e5eb78380cdabc7b41f6c0b05a932a1539dece7`; the worktree was clean and `main` was four commits ahead of `origin/main` when this planning session began.
- **Live execution snapshot:** implementation was authorized on 2026-08-18 from commit `6e5eb78380cdabc7b41f6c0b05a932a1539dece7`; `main` remained four commits ahead of `origin/main`, and this tracker remained the only staged path at activation.
- **Status:** implementation complete. Every authorized workstream is `DONE`; this tracker remains the controlling execution and handoff artifact.
- **Execution boundary:** production code, tests, adopted specifications, authored contracts, SQL inputs, boundary policy, harness owner inputs, and generated projections may change only where the approved workstream requires them. Public HTTP/OpenAPI/WebSocket behavior, persisted source-boundary v1 bytes, and database schema remain frozen.
- **Non-goals:** no database migration, frontend source change, durable outbox, multi-process delivery design, public route or payload change, or permanent internal compatibility shim is planned. Discovery requiring one stops the affected slice for owner review.
- **Authorization:** the remediation request authorizes WF-00, SP-01, SP-02, and implementation slices RS-00, RS-01, RS-04, RS-05, RS-03, RS-02, RS-06, and RS-07 in that order.
- **NLSpec posture:** this tracker uses NLSpec-style normative language to make the proposed refactor reproducible. It is not an adopted NLSpec, does not amend an owner document, and cannot authorize a public behavior change.
- **Revision posture:** `temp/analysis-notes.md` and `docs/research/nlspec-spec.md` were inspected as research and writing guidance only. Their recommendations are accepted only where they agree with adopted owners and live repository evidence.

The key words **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY** are normative only for a later authorized refactor that elects to execute this tracker. **MUST** and **MUST NOT** identify binary conformance requirements. **SHOULD** and **SHOULD NOT** require a recorded, owner-compatible reason for deviation. **MAY** identifies an optional implementation choice that does not change a frozen contract. An invalid or omitted enum, role set, lifecycle rule, dependency, or result disposition has no implicit default and MUST fail closed.

| Requirement ID | Normative posture |
| --- | --- |
| `INC-REF-001` | A later implementation MUST preserve every owner-defined observable behavior in section 4 unless a separate authorized owner change explicitly supersedes it. |
| `INC-REF-002` | A package path or interface proposed by this tracker MUST be treated as implementation-support architecture, not as a new public product contract. |
| `INC-REF-003` | When adopted owner text and repository behavior conflict, the affected slice MUST stop with `BLOCKED: owner contradiction`; it MUST NOT choose a side. |
| `INC-REF-004` | Generated outputs MUST be changed only by changing their authored owner inputs and running the repository-owned generator. |
| `INC-REF-005` | Every migrated consumer MUST retain an explicit owner reference, role set, lifecycle rule, concealment outcome, transaction posture, and executable evidence row. |
| `INC-REF-006` | Implementation changes MUST remain within the active workstream, satisfy its exit criteria, and be recorded here before a dependent workstream begins. |

### 1.1 Execution ledger

Status values are closed: `TODO` has not begun, `IN_PROGRESS` is the sole active
workstream, `DONE` has satisfied every exit criterion, and `BLOCKED` identifies an
owner contradiction or external prerequisite. A dependent workstream MUST NOT begin
until every predecessor is `DONE`. Updating this ledger and the relevant detailed
rows is the final action of each workstream before Markdown and tracker-diff checks.

| Workstream | Status | Evidence and validation | Residual risk or blocker | Next dependency |
| --- | --- | --- | --- | --- |
| WF-00 — activate and correct tracker | DONE | Live Git snapshot recorded; staged history preserved; Markdown and tracker-diff gates passed (`20260818T225058Z-p43345`) | None | SP-01 |
| SP-01 — close `OG-001` | DONE | Core 02 REQ-02-143b adopted; Core 01 references it; AC-266 and Revisions schema/vectors added; Markdown/JSON/hash/diff gates passed (`20260818T225336Z-p45891`, `20260818T225410Z-p46849`) | None | SP-02 |
| SP-02 — close `OG-002` | DONE | Core 01 REQ-01-592 and Core 04 REQ-04-017/AC-162/AC-419 adopt post-commit ordering, no-effect outcomes, target isolation, and process-local delivery; Markdown/diff gates passed (`20260818T225551Z-p48786`) | None | RS-00 |
| RS-00 — characterization baseline | DONE | Exact source-boundary and Collaboration isolation tests added; 70-method admission ledger complete; Incidents and four collaborator focused/service-backed baselines plus static/harness/drift gates pass | None; two test-authoring failures were corrected and retained below | RS-01 |
| RS-01 — neutral admission and transaction safety | DONE | Typed fail-closed admission added; all 70 production consumer methods migrated; legacy facade removed; incident and membership mutations reauthorize in the caller transaction; cross-admin deletion is serialized; focused/service-backed consumer matrix and static/harness/drift gates pass | None; transient test, infrastructure, allowlist, lint, and harness-count failures were corrected or cleanly rerun and are retained below | RS-04 |
| RS-04 — Workbook persistence relocation | DONE | Workbook-owned `bootstrapport` and `startup/postgres` repository/session/writer implemented; create/import dependencies injected; old Incidents package removed; authored SQL and generated Go projection renamed; focused/service-backed Incidents, Workbook, and server assembly plus static/drift gates pass | None; one expected broad boundary-token mismatch was narrowed and is retained below | RS-05 |
| RS-05 — Revisions source-boundary port | DONE | Revisions-owned resolver selects and serializes the v1 boundary; dual characterization passed before deletion; Reporting receives the injected resolver; Incidents provider now exposes incident metadata only; owner rows and boundary policy prevent persistence regression | None; the expected dedicated-fixture harness count was updated and retained below | RS-03 |
| RS-03 — post-commit effect coordinator | DONE | Incidents returns closed fresh/replay commit results keyed by the committed audit UUID; assembled coordinator invokes effects only after fresh commits; Collaboration hub is incident/user indexed with no revocation database lookup; focused/service-backed Incidents, Collaboration, and server slices plus boundary, lint, and harness gates pass | None; one test query column and the expected transaction-fixture count were corrected and retained below | RS-02 |
| RS-02 — Incidents HTTP extraction | DONE | Registration, auth/session handling, decoding, paging, response/error mapping, audit-list transport, and white-box tests moved to `incidents/httpapi`; application and terminal coordinator are injected; all 11 operations and OpenAPI/runtime parity pass; root production files import no HTTP packages | None; selector paths, one fixture classification, and generated topology were updated from authored inputs | RS-06 |
| RS-06 — compatibility/test-support cleanup | DONE | Searches find no retired admission, Workbook, Reporting, effect, or root-route surface; regression boundary rules forbid their return; Incidents-specific mutation, performance, route, scenario, and store support remains local; all affected owner matrices and static/harness gates pass | None | RS-07 |
| RS-07 — validation and handoff | DONE | Obsolete `.gitkeep` removed; final 45-file inventory recorded; generation, every focused/service-backed owner matrix, browser `58/58`, boundary `3/3`, harness `2/2`, drift/policy/JSON/Markdown, finalization `1/1`, and full check `626/626` pass | None; one Reporting run failed in object-store preflight and passed cleanly on rerun; retained-run checks were intentionally skipped because `RESULTS_DIR` was unset | Complete |

The source hierarchy used here is:

1. Adopted subsystem NLSpecs, within each document's named boundary.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication. It is not applicable to this planning-only tracker because no such claim is published.
4. Domain vocabulary and implementation-support guides for terminology, package boundaries, and harness mechanics.
5. Current repository code and tests for current implementation state.
6. Prior plans, handoffs, and the planning framework as evidence and doctrine only.

| Source class | Permitted assertion in this tracker | Prohibited use |
| --- | --- | --- |
| Adopted subsystem NLSpec | Named-subsystem behavior, terminology, ownership, and acceptance requirements | Applying its rules outside the named subsystem |
| Core 00 through Core 04 | Cross-subsystem implementation-conformance behavior and security outcomes | Replacing a more specific adopted subsystem owner |
| Core 05 | Timed or fixture-sensitive publication requirements only | Treating this planning document as claim-bearing publication |
| Domain vocabulary and implementation guides | Vocabulary, navigation, package mechanics, command discovery, and execution support | Creating public behavior not owned by Core or an adopted NLSpec |
| Live repository code and tests | Current implementation, dependency, interface, query, test, and generated-projection state | Treating current code as authority when it contradicts an adopted owner |
| Planning framework, prior handoffs, `analysis-notes.md`, and `nlspec-spec.md` | Planning doctrine, research evidence, candidate designs, and writing discipline | Treating recommendations or examples as proof of repository state or adopted behavior |

No owner contradiction was found. If later owner review exposes one, the affected work must be marked `BLOCKED: owner contradiction` rather than resolved by this tracker.

Owner and support documents inspected:

- `AGENTS.md`
- `docs/handoffs/cartulary_modular_refactor_planning_framework.md` (read first)
- `docs/domain.md`
- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/reporting-subsystem-nlspec.md`
- `docs/extension-subsystem-nlspec.md`
- `docs/network-flow-activity-nlspec.md`
- `docs/report-composition-nlspec.md`
- `docs/testing-harness-nlspec.md`
- `docs/guides/cartulary_implementation_testing_guide.md`
- `docs/research/nlspec-spec.md`
- `temp/analysis-notes.md`

Repository evidence inspected:

- Every file under `internal/modules/incidents`, inventoried individually in section 2.
- Composition roots and adapters: `internal/app/server/runtime_assembly.go`, `internal/app/workbookassembly/startup.go`, `internal/app/recoveryassembly/state_catalog.go`, and `internal/app/incidentportabilityassembly/catalog.go`.
- Neighboring boundary examples and consumers: `internal/modules/indicators/httpapi/routes.go`, `internal/modules/revisions/httpapi/routes.go`, `internal/modules/reporting/export_materializer.go`, `internal/modules/reporting/boundary_guard_test.go`, `internal/modules/workbook/startup/api.go`, and `internal/modules/workbook/startup/store.go`.
- Authored contracts and SQL: `contracts/openapi-source/owners/module.incidents/openapi.json`, `contracts/incident-bundles/source_catalog.json`, `contracts/incident-bundles/traceability.json`, `contracts/verification/owners/module.incidents.json`, `db/queries/incidents.sql`, and `db/queries/incidents_workbook_preferences.sql`.
- Harness ownership: `tools/test_catalog_owner.json` and `tools/test_families/module.incidents.json`.
- Representative frontend consumers: `apps/web/src/app/IncidentLanding.tsx`, `apps/web/src/app/IncidentAdminPanel.tsx`, `apps/web/src/collaboration/IncidentCollaborationSession.tsx`, `apps/web/src/workbook/WorkbookShell.tsx`, and `apps/web/src/workbook/lifecycle/workbookInvalidation.ts`.
- Generated projections were inspected only as downstream evidence: `internal/gen/administrativeauditregistry/registry_gen.go`, `internal/gen/contractopenapi/artifacts_gen.go`, `internal/gen/contractrecovery/artifacts_gen.go`, `internal/gen/openapioperations/catalog_gen.go`, `internal/gen/sql/incidents.sql.go`, `internal/gen/sql/incidents_workbook_preferences.sql.go`, `internal/gen/sql/models.go`, `packages/protocol-ts/src/generated/core-http-types.ts`, `packages/protocol-ts/src/generated/core-http-validators.ts`, and `packages/protocol-ts/src/generated/http-operation-bindings.ts`.
- Boundary and exact-behavior evidence added during the NLSpec-grade revision: `tools/backend_module_boundaries.json`, `internal/modules/workbook/startup/**`, `internal/modules/reporting/store.go`, `internal/modules/incidents/reportingprovider/provider.go`, `internal/modules/incidents/access.go`, `internal/modules/incidents/lifecycle_access.go`, `internal/modules/incidents/audit.go`, `internal/platform/administrativeaudit/audit.go`, incident lifecycle and membership application/route methods, and live admission call sites across peer modules.

## 2. Activation-State Repository Inventory

The target contained 43 files at tracker activation. This baseline inventory is retained
as planning history; workstream results and current disposition are recorded in the
execution ledger and handoff log. Generated artifacts named below are affected
surfaces, not edit targets.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/incidents/.gitkeep` | Historical empty-directory placeholder | None | None | None | None | None | None | low | Obsolete now that the directory is populated; cleanup only in a later slice. |
| `internal/modules/incidents/access.go` | Activation-time HTTP-shaped admission facade; deleted in RS-01 | Retired: `Access`, `AccessService`, `NewAccess` | All production callers migrated to `incidents/admission` in RS-01 | None after deletion | Replacement admission and consumer tests | Boundary policy bans the retired surface | `incidents/admission` | resolved | No compatibility alias remains. |
| `internal/modules/incidents/api.go` | HTTP request decoding, normalization, hashing inputs, resource building, and API error mapping | Request DTOs; decode/build helpers; `WouldLeaveNoIncidentAdmins` | Incident routes and contract/unit tests | platform HTTP API, auth types, list query, UUID | `request_test.go`, `unit_test.go`, lifecycle and OpenAPI tests | Incident OpenAPI owner and generated protocol/operation bindings (affected) | `incidents/httpapi` candidate | high | Protocol and transport logic is in the root package. |
| `internal/modules/incidents/application.go` | Application composition, transaction/policy dependencies, and administrative-audit listing | `Application`, `NewApplication`, `NewApplicationWithOptions` | Server assembly, workbook/import/bundle coordination, performance fixtures, and many integration tests | auth store, repository, workbook preference bootstrap, commit port | Store, integration, lifecycle, membership, and performance tests | SQL and audit registry projections (affected) | `incidents` | high | Legitimate facade, but currently constructs several concrete adapters. |
| `internal/modules/incidents/audit.go` | Maps incident and membership mutations to deployment-local administrative audit events | Package-private mapping/appending helpers | Incident and membership applications; import finalization | platform administrative audit, UUID/time | Extra, store, lifecycle, and membership-audit tests | Administrative-audit registry (affected) | `incidents` owner contribution | medium | Must remain distinct from Revisions change sets. |
| `internal/modules/incidents/extra_integration_test.go` | Extra store-backed audit, replay, and failure characterization | Test-only | Go test runner; module incident integration family | App support, PostgreSQL, incident APIs | Self | Harness row evidence only | `incidents` tests | low | Retained behavior evidence. |
| `internal/modules/incidents/http_conformance_test.go` | Live HTTP authorization, route, error, and envelope conformance | Test-only | Go test runner; module incident integration family | Server test support, auth, HTTP | Self | OpenAPI/protocol behavior evidence | `incidents/httpapi` tests candidate | medium | Must follow transport extraction. |
| `internal/modules/incidents/import_finalization.go` | Final incident-bundle publication bootstrap for membership, preferences, and audit | `NewIncidentBundleImportFinalizer`, `ImportBundleRequestID` | Incident portability assembly and bundle import flow | `pgx.Tx`, auth, workbook preferences, audit | Store and bundle integration tests | Incident-bundle source/traceability contracts (affected) | `incidents` orchestration | high | Cross-concern work is required to remain atomic. |
| `internal/modules/incidents/incident_application.go` | Incident list/create/patch/close/reopen orchestration, idempotency, and audit | Methods on `Application` | Routes, fixtures, bundle/import coordination, tests | Repository, auth, list query, preferences, audit, commit port | Store, integration, lifecycle, pagination, request tests | Incident OpenAPI and SQL projections (affected) | `incidents` | high | Core incident lifecycle behavior. |
| `internal/modules/incidents/incident_bundle_portability.go` | Incident resource export/import codec and validation | `ExportIncidentBundleIncident`, `ImportIncidentBundleIncidentTx` | Incident bundle source port | Incident portability contracts, `pgx.Tx`, generated SQL | Bundle round-trip and route integration tests | Incident-bundle source catalog and traceability (affected) | `incidents` owner adapter | medium | Narrow source-owner contribution. |
| `internal/modules/incidents/incident_bundle_source_port.go` | Registers the incidents source contribution with bundle portability | `NewIncidentBundleSourcePort` | Incident portability assembly | `incidentbundles/sourceport`, portability codec | Bundle integration tests | `contracts/incident-bundles/source_catalog.json` (affected) | `incidents` owner adapter | medium | Intentional cross-owner port. |
| `internal/modules/incidents/integration_test.go` | Broad live incident and membership route behavior | Test-only | Go test runner; module incident integration rows | Server app support, HTTP/auth fixtures | Self | OpenAPI and harness evidence | `incidents` tests | low | Covers ordinary end-to-end backend behavior. |
| `internal/modules/incidents/inventory_helpers_test.go` | Test route/audit inventory helpers | Test-only helper surface | Incident tests | Route inventory and test support | Incident test files | Harness/route inventory evidence | `incidents` tests | low | Test-only composition. |
| `internal/modules/incidents/lifecycle_access.go` | Create/lifecycle hashes, patch application, and shared HTTP-shaped incident admission | `IncidentCreateBootstrap`, hash/patch helpers, `IncidentAccessError`, `RequireIncidentMembership`, `RequireIncidentRole` | Routes and numerous peer modules | `Access`, platform `httpapi.APIError`, UUID/time | Unit, lifecycle, route, and peer-module tests | OpenAPI error behavior (affected) | Split between `incidents` admission and HTTP adapters | high | Semantic policy and transport error representation are coupled. |
| `internal/modules/incidents/lifecycle_contract_test.go` | Lifecycle request/OpenAPI contract alignment | Test-only | Go test runner; module incident unit row | Request decoders and OpenAPI source | Self | OpenAPI owner/protocol projections | `incidents/httpapi` tests candidate | low | Contract freeze evidence. |
| `internal/modules/incidents/lifecycle_integration_test.go` | Close/reopen authorization, idempotency, audit, and closed behavior | Test-only | Go test runner; module incident integration row | Server test support and incidents routes | Self | OpenAPI/harness evidence | `incidents` tests | medium | Protects a high-risk lifecycle boundary. |
| `internal/modules/incidents/membership_application.go` | Membership list/create/update/delete orchestration, idempotency, last-admin protection, and audit | Methods on `Application` | Routes, fixtures, tests, import/bootstrap flows | Repository, auth, audit, commit port | Store, integration, audit, and request tests | Membership OpenAPI and SQL projections (affected) | `incidents` membership sub-area | high | Legitimate incidents concern; authorization is currently largely route-adjacent. |
| `internal/modules/incidents/membership_audit_contract_test.go` | Membership-audit OpenAPI and exact-shape checks | Test-only | Go test runner; module incident unit row | OpenAPI source and audit handlers | Self | OpenAPI and audit registry projections | `incidents/httpapi` tests candidate | low | Contract alignment evidence. |
| `internal/modules/incidents/membership_audit_handlers.go` | Membership-audit HTTP query, paging, role gate, and resource mapping | Package-private `Service` handlers/helpers | `routes.go` | platform administrative audit, pagination, HTTP API | Audit contract and integration tests | Audit/OpenAPI projections (affected) | `incidents/httpapi` candidate | high | Owner-specific transport over a platform audit substrate. |
| `internal/modules/incidents/membership_audit_integration_test.go` | Scope safety, authorization, paging, and audit visibility | Test-only | Go test runner; module incident integration row | Server/auth test support | Self | Audit/OpenAPI/harness evidence | `incidents` tests | medium | Protects hidden-not-found and role behavior. |
| `internal/modules/incidents/models.go` | Incident/membership records, pagination positions, mutation results, and typed version conflict | Public record/result/error types | Applications, routes, adapters, tests, and consumers | UUID/time | Unit, store, route, and integration tests | Generated SQL models and protocol resources (indirectly affected) | `incidents` | medium | Transport-neutral models should remain at the owner boundary. |
| `internal/modules/incidents/mutation_admission_lock_integration_test.go` | Concurrent incident mutation admission/locking evidence | Test-only | Go test runner; module incident store row | PostgreSQL, application, transaction helpers | Self | Harness evidence | `incidents` tests | medium | Guards serialization and same-incident mutation behavior. |
| `internal/modules/incidents/openapi_contract_test.go` | Runtime route-to-OpenAPI parity and schema checks | Test-only | Go test runner; module incident unit row | OpenAPI source, route registration | Self | OpenAPI operation and protocol projections | `incidents/httpapi` tests candidate | medium | Must move with route ownership, not be weakened. |
| `internal/modules/incidents/pagination_integration_test.go` | Live incident and membership keyset pagination behavior | Test-only | Go test runner; module incident integration family | Server/auth test support, routes | Self | OpenAPI/harness evidence | `incidents` tests | medium | Freezes ordering, cursor binding, and live authorization. |
| `internal/modules/incidents/ports.go` | Application bootstrap, commit, import finalization, collaboration, and route option ports | `PreferenceBootstrapPort`, `IncidentCreateCommitPort`, `IncidentBundleImportFinalizer`, `CollaborationSessionPort`, option structs | Application, routes, server and portability assembly, tests | `pgx.Tx`, UUID/time | Store, route, and bundle tests | None directly | `incidents` | high | Several ports expose transaction mechanics or transport composition. |
| `internal/modules/incidents/recovery_state.go` | Declares incidents-owned recovery state | `RecoveryStateContribution` | Recovery assembly | Recovery state catalog | Recovery inventory tests | Recovery contract registry (affected) | `incidents` owner contribution | medium | Keep as owner evidence, not runtime architecture. |
| `internal/modules/incidents/reportingprovider/provider.go` | Builds the Incidents-owned Reporting metadata snapshot | `GetIncidentSnapshotTx` | Reporting export materializer | `pgx.Tx`, Incidents persistence, Reporting export port | Reporting and Revisions source-boundary tests | Reporting contracts (affected) | Incidents snapshot adapter | resolved | RS-05 removed all Revisions table, selection, serialization, and hash knowledge. |
| `internal/modules/incidents/repository.go` | Private sqlc-backed incident and membership persistence | Package-private repository/session methods | Incident and membership applications, access, import finalizer | `postgres.DB`, `pgx.Tx`, generated incident queries | Store and integration tests | `internal/gen/sql/incidents.sql.go` (affected) | Private `incidents` persistence adapter | medium | Direct SQL is hidden behind the owner facade and is intentional. |
| `internal/modules/incidents/request_test.go` | Decoder, normalization, request hash, resource, and guard tests | Test-only | Go test runner; module incident unit/support rows | `api.go`, lifecycle helpers | Self | OpenAPI/protocol behavior evidence | `incidents/httpapi` tests candidate | low | Move with protocol decoding if extracted. |
| `internal/modules/incidents/routes.go` | Route registration, authentication, CSRF/session handling, authorization, dispatch, mutation responses, and collaboration notification | `Service`, `RegisterRoutes` | Server runtime assembly and HTTP tests | Application, Access, auth, HTTP API, pagination, collaboration port | HTTP conformance, integration, lifecycle, audit, and unit tests | OpenAPI operation catalog and TypeScript bindings (affected) | `incidents/httpapi` candidate | high | Post-commit notifications are currently initiated here. |
| `internal/modules/incidents/store_test.go` | Store atomicity, bootstrap, idempotency, version conflict, import finalization, and membership guard evidence | Test-only | Go test runner; module incident store rows | PostgreSQL harness, application, test support | Self | SQL/harness evidence | `incidents` tests | medium | Central persistence characterization suite. |
| `internal/modules/incidents/support_test.go` | Shared incident test setup and support fixtures | Test-only helper surface | Root incident test files | Test containers/app support/auth | Incident test suite | Harness evidence | `incidents` tests | low | Test-only package support. |
| `internal/modules/incidents/testsupport/mutationtest/mutations.go` | Incidents-specific queries and assertions over administrative-audit and incident mutation artifacts | `MutationOwner`, `AuditEventRecord`, `Database`, selectors and assertion helpers | Incidents-local test packages only | SQL/PostgreSQL test handles, testing package, Incidents event vocabulary | Incidents integration tests | Harness evidence only | Incidents test support | low | Direct-import evidence confirms high cohesion; retain in place. |
| `internal/modules/incidents/testsupport/performancefixture/production.go` | Production-backed incident fixture application for performance harnesses | `ProductionApplication`, `NewProductionApplication` | App-support performance fixture assembly | Incident application, auth, PostgreSQL | Performance/harness tests | Performance fixture profile catalog (affected) | `incidents` test support | low | Owner-specific fixture adapter. |
| `internal/modules/incidents/testsupport/performancefixture/provider.go` | Harness fixture provider abstraction and descriptor binding | `Application`, `Provider`, `New` | Performance fixture assembly | Harness fixture contracts | Performance/harness tests | Performance fixture profile catalog (affected) | `incidents` test support | low | Owner-specific published test helper. |
| `internal/modules/incidents/testsupport/routetest/routes.go` | Canonical public/control route inventories for incident and membership families | Four exported route inventory functions | Incident, auth, and browser-support tests | Route inventory test utility | Route inventory and conformance tests | OpenAPI/harness evidence | `incidents` test support | low | Semantic owner fixture; keep. |
| `internal/modules/incidents/testsupport/scenariotest/incidents.go` | HTTP scenario helpers for incident and membership mutations | Exported create/patch/delete helpers | Many module integration tests | HTTP/auth flow test utilities | Cross-module integration tests | Harness evidence | `incidents` test support | low | Incident-specific shared fixture; keep. |
| `internal/modules/incidents/testsupport/storetest/harness.go` | PostgreSQL-backed incident store test harness | `StoreHarness`, `StartStore` | Incident and peer store tests | PostgreSQL test utility, incident application | Store/integration tests | Harness evidence | `incidents` test support | low | Incident-specific harness wrapper. |
| `internal/modules/incidents/testsupport/storetest/store.go` | Direct incident/user/membership fixture creation, replay snapshots, counts, and seeds | Exported store fixture/query helpers | Many module integration tests | PostgreSQL/sql, auth, incident application | Cross-module store and integration tests | SQL/harness evidence | `incidents` test support | medium | Keep semantic fixtures; review raw generic count/query helpers with `mutationtest`. |
| `internal/modules/incidents/transaction_participant.go` | Incident-row lock participant for cross-owner atomic Network Flow/import work | `TransactionParticipant`, `NewTransactionParticipant` | Server assembly and Network Flow store | `pgx.Tx`, generated incident query | Network Flow store and incident lock tests | Network Flow contract evidence (affected) | `incidents` owner contribution | medium | Narrow source-owner adapter matching Network Flow's private transaction port. |
| `internal/modules/incidents/unit_test.go` | Focused access/error/decoder and helper unit behavior | Test-only | Go test runner; module incident unit rows | Incident public helpers | Self | OpenAPI/harness evidence | `incidents` tests | low | Retained characterization. |
| `internal/modules/incidents/workbookpreferences/bootstrap.go` | Activation-time bootstrap implementation; deleted in RS-04 | Retired: `Bootstrap`, `NewBootstrap` | Callers now consume Workbook's `bootstrapport.Writer` | None after deletion | Incident create/import tests | Replaced by `workbook_startup_preferences.sql.go` | `workbook/startup/postgres` | resolved | No default Incidents-owned constructor remains. |
| `internal/modules/incidents/workbookpreferences/repository.go` | Activation-time preference repository/session; deleted in RS-04 | Retired repository and session | Workbook assembly now constructs Workbook's PostgreSQL repository/session | None after deletion | Workbook startup and incident store tests | Replaced by `workbook_startup_preferences.sql.go` | `workbook/startup/postgres` | resolved | Ordinary preference persistence and bootstrap share one semantic owner. |

### 2.1 Final-State Repository Inventory

RS-07 recalculated the live target after cleanup. The final target contains 45 files:
25 root files, 2 admission files, 10 HTTP adapter files, 1 Reporting provider, and
7 Incidents-specific test-support files. The deleted `.gitkeep` and retired activation
paths remain only in the historical inventory above.

| Final path | Final responsibility and disposition |
| --- | --- |
| `internal/modules/incidents/admission/admission.go` | Transport-neutral, typed, fail-closed incident admission. |
| `internal/modules/incidents/admission/admission_test.go` | Exact role-set, denial, malformed-value, lifecycle, and nil-dependency evidence. |
| `internal/modules/incidents/application.go` | Transport-neutral application composition and dependency validation. |
| `internal/modules/incidents/audit.go` | Incidents-owned administrative-audit event construction and committed event keys. |
| `internal/modules/incidents/contracts.go` | Transport-neutral application commands and resources. |
| `internal/modules/incidents/extra_integration_test.go` | Store-backed audit, replay, and failure characterization. |
| `internal/modules/incidents/http_conformance_test.go` | Live HTTP authorization, error, and envelope conformance. |
| `internal/modules/incidents/idempotency.go` | Stable request hashing and stored HTTP replay serialization. |
| `internal/modules/incidents/import_finalization.go` | Atomic incident-bundle finalization through injected Workbook bootstrap. |
| `internal/modules/incidents/incident_application.go` | Incident list/create/patch/close/reopen orchestration and typed commit results. |
| `internal/modules/incidents/incident_bundle_portability.go` | Incidents-owned bundle codec and validation contribution. |
| `internal/modules/incidents/incident_bundle_source_port.go` | Narrow Incidents source registration for bundle portability. |
| `internal/modules/incidents/integration_test.go` | Broad incident and membership runtime behavior evidence. |
| `internal/modules/incidents/inventory_helpers_test.go` | Incidents test route and audit inventory helpers. |
| `internal/modules/incidents/lifecycle_access.go` | Transport-neutral create hashing, patch application, and lifecycle helpers. |
| `internal/modules/incidents/lifecycle_integration_test.go` | Close/reopen authorization, replay, audit, and lifecycle evidence. |
| `internal/modules/incidents/membership_application.go` | Serialized membership mutations, transactional reauthorization, last-admin protection, and typed terminal results. |
| `internal/modules/incidents/membership_audit_integration_test.go` | Membership-audit scope, authorization, paging, and visibility evidence. |
| `internal/modules/incidents/membership_concurrency_integration_test.go` | Concurrent cross-admin safety and deadlock regression evidence. |
| `internal/modules/incidents/models.go` | Transport-neutral records, pagination positions, conflicts, and closed terminal dispositions. |
| `internal/modules/incidents/mutation_admission_lock_integration_test.go` | Transaction-time admission and incident-row locking evidence. |
| `internal/modules/incidents/pagination_integration_test.go` | Live incident and membership keyset pagination evidence. |
| `internal/modules/incidents/ports.go` | Narrow application-owned persistence, bootstrap, commit, and contribution ports. |
| `internal/modules/incidents/recovery_state.go` | Incidents-owned recovery-state contribution. |
| `internal/modules/incidents/repository.go` | Private sqlc-backed incident and membership persistence. |
| `internal/modules/incidents/store_test.go` | Atomicity, bootstrap, idempotency, conflict, import, and membership persistence evidence. |
| `internal/modules/incidents/transaction_participant.go` | Narrow incident-row lock participant for cross-owner atomic work. |
| `internal/modules/incidents/httpapi/api.go` | Owner-local HTTP DTO decoding, normalization, resource building, and denial/error mapping. |
| `internal/modules/incidents/httpapi/lifecycle_contract_test.go` | Lifecycle request and OpenAPI contract alignment. |
| `internal/modules/incidents/httpapi/location_test.go` | Pure HTTP Location derivation evidence. |
| `internal/modules/incidents/httpapi/membership_audit_contract_test.go` | Membership-audit HTTP and OpenAPI shape evidence. |
| `internal/modules/incidents/httpapi/membership_audit_handlers.go` | Membership-audit query, paging, authorization mapping, and HTTP resources. |
| `internal/modules/incidents/httpapi/openapi_contract_test.go` | Runtime registration and all-operation OpenAPI parity. |
| `internal/modules/incidents/httpapi/request_test.go` | Decoder, validation order, hashing, resource, and replay behavior evidence. |
| `internal/modules/incidents/httpapi/routes.go` | All 11 route bindings, authentication/session handling, dispatch, and response mapping. |
| `internal/modules/incidents/httpapi/support_test.go` | HTTP adapter white-box test support. |
| `internal/modules/incidents/httpapi/unit_test.go` | Focused HTTP adapter behavior evidence. |
| `internal/modules/incidents/reportingprovider/provider.go` | Incidents metadata-only Reporting provider; no Revisions persistence access. |
| `internal/modules/incidents/testsupport/mutationtest/mutations.go` | Incidents-specific mutation and administrative-audit assertions. |
| `internal/modules/incidents/testsupport/performancefixture/production.go` | Production-backed incident performance fixture application. |
| `internal/modules/incidents/testsupport/performancefixture/provider.go` | Incidents performance fixture provider abstraction. |
| `internal/modules/incidents/testsupport/routetest/routes.go` | Canonical Incidents route inventories. |
| `internal/modules/incidents/testsupport/scenariotest/incidents.go` | Shared Incidents HTTP scenario helpers. |
| `internal/modules/incidents/testsupport/storetest/harness.go` | PostgreSQL-backed Incidents store harness. |
| `internal/modules/incidents/testsupport/storetest/store.go` | Incidents-specific store seeds, counts, and replay fixtures. |

## 3. Module Boundary Diagnosis

At activation, the root was not an accidental home for every incident-scoped feature:
entities, indicators, evidence, links, timeline, saved views, projections, workbook
interaction, and grid behavior had distinct implementations elsewhere. The legitimate
incident/membership facade was nevertheless colocated with HTTP transport, shared
admission, private persistence, owner contributions, Workbook preference persistence,
and test support. The final state resolves the improper edges: transport and admission
have dedicated Incidents subpackages, Workbook and Revisions own their persistence,
terminal effects are application-assembled, and only cohesive private persistence,
owner contributions, and Incidents-specific test support remain.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Incident collection, create, metadata patch, close, and reopen | Root application/repository files | `incidents` | keep | Core 01 incident resource/lifecycle requirements; live application and tests | Legitimate central orchestration. |
| Membership lifecycle and creator/admin bootstrap | Root membership application and import finalizer | `incidents` membership sub-area | keep | Core 01 membership routes/bootstrap; Core 04 role rules | May be internally separated without inventing a new owner. |
| HTTP decoding, response/error mapping, and route registration | `api.go`, `routes.go`, audit handlers | `incidents/httpapi` | split | Existing `indicators/httpapi` and `revisions/httpapi` patterns; platform owns generic HTTP plumbing | Preserve all operation IDs and envelopes. |
| Incident membership/open-state policy | `access.go`, `lifecycle_access.go` | `incidents` admission facade | split | Wide peer-module caller set; Core 04 owns semantics | Policy stays incident-owned; HTTP-shaped errors do not. |
| Private incident and membership persistence | `repository.go` | Private incidents adapter | keep | SQL is hidden behind the application/access facade | Do not move merely because it uses PostgreSQL. |
| Workbook startup preference persistence | `incidents/workbookpreferences` | `workbook/startup` persistence adapter | move | `workbook/startup` already owns the records, interfaces, routes, and behavior; app assembly adapts this package | Retain an injected transaction bootstrap port for incident creation/import. |
| Incident administrative audit projection | `audit.go`, membership audit handlers | Incidents owner contribution plus `incidents/httpapi` read adapter | split | Domain/Core distinguish deployment-local audit from incident history | Generic audit storage remains platform-owned. |
| Incident bundle source and import publication | Bundle port/codec/finalizer files | Incidents owner contribution | keep | Incident bundle source catalog and app assembly | Cross-owner coordination is intentional. |
| Reporting incident snapshot | `incidents/reportingprovider` | Incidents source adapter | keep | Reporting boundary guard explicitly permits it | Snapshot reading is legitimate. |
| Reporting source high-water derived from Revisions change sets | `incidents/reportingprovider` | `internal/modules/revisions/sourceboundary` | move | Reporting needs an immutable boundary; Revisions owns change sets | The design is resolved; `OG-001` gates implementation. |
| Recovery state declaration | `recovery_state.go` | Incidents owner contribution | keep | Recovery assembly consumes owner contributions | Evidence/catalog role only. |
| Network Flow incident row lock | `transaction_participant.go` | Incidents owner contribution | keep | Adopted Network Flow contract and private participant port | Do not generalize transaction ownership. |
| Incidents mutation/audit test queries | `testsupport/mutationtest` | Incidents test support | keep | All direct imports are Incidents-local and assertions encode Incidents event vocabulary | Moving would create misleading generic ownership. |
| Incident scenario/store/route/performance fixtures | Other `testsupport/**` | Incidents test support | keep | Helpers create or exercise incident-owned state | Published test-only semantic surface. |
| Frontend shell/controller and collaboration client state | `apps/web`, not the target | Web/workbook/collaboration owners | defer | Representative consumers inspected; no implementation in target | Preserve indirect contracts only. |
| Entities, indicators, evidence, links, timeline, saved views, projections, and grid integration | Other modules, not the target | Their existing owners | defer | Caller imports use incidents access; no corresponding runtime implementation is in this tree | Do not infer ownership from route prefixes or visible tabs. |

**Repository/framework finding:** the planning framework correctly warns against accepting a directory as a permanent module boundary, but its generic candidate lists are not current-state proof. Live evidence shows that incidents is already a legitimate incident/membership orchestration owner. The adaptation is therefore to deepen and narrow that facade, not dissolve the module or move unrelated incident-scoped features into it.

### 3.1 Resolved owner map

| Current responsibility | Required final owner | Disposition | Normative result |
| --- | --- | --- | --- |
| Incident and membership lifecycle, creator bootstrap, import finalization | `incidents` | keep | Incidents MUST retain the containing application transactions and policy. |
| HTTP registration, decoding, response/error mapping, and paging | `internal/modules/incidents/httpapi` | split | All 11 operation bindings and wire outcomes MUST remain unchanged. |
| Visibility, membership, role, and open-state admission | `internal/modules/incidents/admission` | split | Admission MUST return transport-neutral values and errors. |
| Private incident/member SQL | Private Incidents persistence | keep | Private PostgreSQL use is intentional and MUST NOT be generalized. |
| Workbook preference persistence | `internal/modules/workbook/startup/postgres` | move | Incidents MUST retain only a Workbook-owned bootstrap capability. |
| Reporting source-boundary selection/serialization | `internal/modules/revisions/sourceboundary` | move | Reporting MUST consume a Revisions-owned caller-transaction port. |
| Close/removal terminal-effect orchestration | `internal/app/collaborationassembly/incidenteffects` | split | Incidents and Collaboration MUST remain mutually independent owner packages. |
| Bundle, recovery, Network Flow, and incident-specific test contributions | Existing Incidents owner adapters | keep | Narrow owner contributions MUST remain owner-created. |
| Incidents mutation/audit assertions | Incidents test support | keep | Direct imports and event vocabulary are Incidents-specific; no generic relocation is justified. |
| Frontend, grid, projection, saved-view, entity, indicator, evidence, link, and timeline logic | Existing owners outside target | defer | No relocation is authorized by this tracker. |

### 3.2 Workbook Startup bootstrap contract

`RB-001` is implemented. The Workbook-owned leaf contract
`internal/modules/workbook/startup/bootstrapport` and PostgreSQL implementation
`internal/modules/workbook/startup/postgres` use this interface shape:

```go
package bootstrapport

type InitialPreferenceInput struct {
	IncidentID      uuid.UUID
	UserID          uuid.UUID
	CommitTimestamp time.Time
}

type Writer interface {
	InsertInitialTx(context.Context, pgx.Tx, InitialPreferenceInput) error
}
```

| Requirement ID | Requirement |
| --- | --- |
| `INC-REF-WB-001` | The transaction is caller-owned. The writer MUST NOT begin, commit, roll back, or create a savepoint. |
| `INC-REF-WB-002` | One call MUST insert one incident preference with `default_sheet_ref = NULL` and one user preference with `home_sheet_ref = NULL`. |
| `INC-REF-WB-003` | Both rows MUST use the supplied incident, user, and commit timestamp. The implementation MUST NOT call its own clock. |
| `INC-REF-WB-004` | The operation is insert-only. It MUST NOT update, upsert, repair, or interpret a uniqueness violation as replay. |
| `INC-REF-WB-005` | The writer MUST perform no authorization, audit append, collaboration effect, startup fallback, sheet-reference validation, ordinary preference read/write, or idempotency lookup. |
| `INC-REF-WB-006` | Incident create and import finalization MUST perform replay detection before invoking the writer and MUST include its inserts in their existing atomic transaction. |
| `INC-REF-WB-007` | No relation, column, constraint, migration, SQL behavior, or public Workbook route changes in this ownership move. |

The required dependency direction is Incidents application and Incident Portability finalization to `bootstrapport`; `workbook/startup/postgres` implements that contract; application assembly constructs and injects it. Incidents MUST NOT import `workbook/startup/postgres`, and Workbook Startup persistence MUST NOT import the Incidents application.

| Future authored boundary change | Exact requirement |
| --- | --- |
| `workbook-no-storage-construction` | Permit generated SQL and PostgreSQL construction only below `internal/modules/workbook/startup/postgres/**`; retain the prohibition for every other Workbook path. |
| `incidents-no-workbook-startup-or-platform-ws` | Replace the broad Workbook Startup prohibition with an exact allowance for `workbook/startup/bootstrapport`; retain the platform WebSocket prohibition. |
| Owner-port allowlist | Permit the bootstrap contract only from Incidents create/finalization, Incident Portability finalization, its implementation, assembly, and tests. No other production importer is allowed. |

### 3.3 Revisions source-boundary contract

`RB-002` is implemented in `internal/modules/revisions/sourceboundary` with this
interface shape:

```go
package sourceboundary

type ResolveInput struct {
	IncidentID      uuid.UUID
	IncidentVersion int64
}

type Boundary struct {
	Token         string
	CanonicalJSON []byte
}

type Resolver interface {
	ResolveCurrentTx(context.Context, pgx.Tx, ResolveInput) (Boundary, error)
}
```

| Requirement ID | Requirement |
| --- | --- |
| `INC-REF-SB-001` | The caller owns the transaction. The resolver MUST NOT begin, commit, roll back, or change its isolation/access mode. |
| `INC-REF-SB-002` | The incident read and Revisions lookup MUST use the same existing `REPEATABLE READ` snapshot. That transaction is not read-only because Reporting subsequently writes snapshot-admission state. |
| `INC-REF-SB-003` | Revisions alone reads `change_sets`; the supplied incident version remains Incidents-owned. Reporting receives no table name, query handle, SQL fragment, or generated Revisions row. |
| `INC-REF-SB-004` | Visible committed change sets MUST be ordered by `created_at DESC`, then canonical `change_set_id DESC`, and at most one row is selected. |
| `INC-REF-SB-005` | Canonical JSON MUST contain exactly `incident_id`, `incident_version`, `latest_change_set_id`, and `latest_change_set_created_at`. Both latest-change-set members MUST be JSON `null` when no visible row exists; fields MUST NOT be omitted. |
| `INC-REF-SB-006` | `Token` MUST equal `cartulary.source_boundary.v1:` plus lowercase hexadecimal SHA-256 of the exact canonical bytes. Returned bytes MUST be immutable or defensively copied. |
| `INC-REF-SB-007` | Administrative-audit events MUST NOT participate. Exact replay MUST reuse the boundary resolved for the original admitted job. |

`OG-001` is an authority prerequisite: Core 02 must adopt the exact canonical member names and deterministic tie rule before the direct Incidents-provider query is removed. Until then, RS-05 remains implementation-gated even though its package and interface design are resolved.

### 3.4 Neutral admission contract

`RB-003` is implemented. `internal/modules/incidents/admission` exposes one concrete,
transport-neutral checker. Consumers own any narrower local interface they require;
the Incidents transport and each peer HTTP owner map denials to their existing public
error family.

```go
package admission

type Role uint8
const (
	RoleViewer Role = iota + 1
	RoleEditor
	RoleReviewer
	RoleAdmin
)

type RoleSet uint8
const (
	AllowViewer RoleSet = 1 << iota
	AllowEditor
	AllowReviewer
	AllowAdmin
)

type Lifecycle uint8
const (
	LifecycleAny Lifecycle = iota + 1
	LifecycleOpen
)

type IncidentStatus uint8
const (
	IncidentStatusActive IncidentStatus = iota + 1
	IncidentStatusClosed
)

type Requirement struct {
	AllowedRoles RoleSet
	Lifecycle    Lifecycle
}

type Grant struct {
	Role           Role
	IncidentStatus IncidentStatus
}

type DenialCode string
const (
	DenialNotVisible DenialCode = "not_visible"
	DenialInsufficientRole DenialCode = "insufficient_role"
	DenialIncidentClosed DenialCode = "incident_closed"
)

type Checker struct { /* private PostgreSQL dependency */ }

func NewChecker(postgres.DB) *Checker
func (*Checker) Check(context.Context, uuid.UUID, uuid.UUID, Requirement) (Grant, error)
func (*Checker) CheckTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, Requirement) (Grant, error)
func (*Checker) RequireOpenTx(context.Context, pgx.Tx, uuid.UUID) error
```

`MembershipReader` and `VisibleIncidentReader` MAY be separate narrow interfaces when a characterized consumer needs owner data beyond `Grant`; `Grant` MUST NOT be widened speculatively.

| Requirement ID | Requirement |
| --- | --- |
| `INC-REF-AD-001` | Roles are an explicit set, not an ordinal hierarchy. Authorization MUST use bit membership and MUST NOT compare role strings or enum order. |
| `INC-REF-AD-002` | Zero or unknown `RoleSet` bits, a zero/unknown lifecycle value, malformed stored role/status, or a nil required dependency MUST fail closed as an internal programming/data error. No implicit default exists. |
| `INC-REF-AD-003` | Every check MUST re-read authoritative current state. A `Grant` is not a cacheable capability token. |
| `INC-REF-AD-004` | A non-member or hidden incident MUST produce `not_visible`; deployment-administrator status MUST NOT bypass incident membership or role requirements. |
| `INC-REF-AD-005` | Transactional source mutation MUST perform its open-state and, where owned there, role recheck inside the caller transaction. Route-time checks alone are insufficient. |
| `INC-REF-AD-006` | Storage, cancellation, and internal failures remain ordinary errors and MUST NOT be converted to a grant or a policy denial. |

Named role sets are exact unions: `member = viewer|editor|reviewer|admin`, `editor-reviewer-admin`, `reviewer-admin`, `editor-admin`, and `admin`. There is no inheritance between them.

| Consumer/operation family | Exact allowed roles | Lifecycle | Concealment | Transaction recheck | New capability | Migration state |
| --- | --- | --- | --- | --- | --- | --- |
| Incidents list/get and membership collection read | member | any | concealed `404` for addressed hidden resource | no | `Checker` or reader | removed |
| Incident metadata patch | reviewer-admin | open | hidden before visible role denial | application transaction | `TxChecker` | removed |
| Close/reopen and membership administration/audit | admin | any; transition rules remain owner-defined | hidden `404`, then visible `403` | application transaction | `Checker` plus owner application | removed |
| Collaboration connect/resume | member | open for writable subscription | concealed | connect/resume | `Checker` plus `VisibleIncidentReader` | removed |
| Entities read/administrative routes | reviewer-admin where the current route requires it | current route rule | concealed | owner mutation stores require open | `Checker` plus `OpenStateTxChecker` | removed |
| Evidence preview/download handle use | member | any | concealed | handle/use revalidation | `Checker` or reader | removed |
| Evidence source/blob mutation | editor-reviewer-admin | open | concealed before visible role denial | yes | `TxChecker`/`OpenStateTxChecker` | removed |
| Imports read/session inspection | member | any | concealed | no | `Checker` | removed |
| Imports region/apply/source mutation | editor-reviewer-admin | open | concealed before visible role denial | yes | `TxChecker` | removed |
| Indicators read | member | any | concealed | no | `Checker` | removed |
| Indicators mutation | editor-reviewer-admin | open | concealed before visible role denial | yes | `Checker` plus `OpenStateTxChecker` | removed |
| Network Flow read | member | any | concealed | no | `Checker` | removed |
| Network Flow write families | editor-admin or reviewer-admin exactly as the owner route states | open where source state changes | concealed before visible role denial | yes | `Checker` plus owner transaction participant | removed |
| Report Composition read/mutate | member for read; editor-admin for mutation | owner-defined | concealed before visible role denial | where transactional | `Checker`/`TxChecker` | removed |
| Reporting read/create/cancel/remove | member for read; editor-reviewer-admin for create/cancel; admin for removal | owner-defined | concealed before visible role denial | snapshot transaction where applicable | `Checker` | removed |
| Revisions read/commands | member for read; reviewer-admin for commands; soft delete additionally allows editor | open for commands | concealed before visible role denial | yes for commands | `Checker`/`TxChecker` | removed |
| Saved Views | member | any | concealed | owner-defined | `Checker` or reader | removed |
| Timeline read/review/time-profile | member for read; reviewer-admin for review; admin for time-profile mutation | open for source mutation | concealed before visible role denial | yes for mutation | `Checker` plus `OpenStateTxChecker` | removed |
| Workbook read/preferences and source mutation | member for read; current exact editor-reviewer-admin, reviewer-admin, or admin set for mutation | any for read; open for source mutation | concealed before visible role denial | yes for source mutation | `Checker`, reader, `OpenStateTxChecker` | removed |
| Job API and Incident Bundle incident-scoped access | Current owner policy plus membership | any | owner-specific not-found concealment | finalizer rechecks prerequisites | `MembershipReader` | removed |
| Artifacts, Parties, Tasks/Decisions, and generic row mutations | role authorized by their owning edge | open | owner-defined | yes | `OpenStateTxChecker` only | removed |

The RS-00 direct-call inventory below is the admission migration ledger. Each row is
one production consumer method; a method with both read and mutation branches records
both capabilities in one row. The call column is the RS-00 baseline call retained for
audit history. Every row reached `removed` only after its legacy import/call was gone
and its owner tests passed against neutral admission.

| ID | Production consumer method | RS-00 baseline call | Requirement posture | Migration state |
| --- | --- | --- | --- | --- |
| AM-001 | `timelineassembly.incidentAdapter.EnsureOpenTx` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-002 | `revisionassembly.commandAuthorizerAdapter.AuthorizeCommandTx` | `AuthorizeMutationTx` | caller-supplied exact roles, open, transaction-bound | removed |
| AM-003 | `indicatorObservationService.createManualObservation` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-004 | `indicatorObservationService.transitionObservation` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-005 | `indicators/httpapi.Service.handleObservationAction` | `RequireIncidentRole` | editor-reviewer-admin, concealed | removed |
| AM-006 | `indicators/httpapi.Service.visibleEnvelope` | `RequireIncidentRole` / `RequireIncidentMembership` | mutation editor-reviewer-admin; read member; concealed | removed |
| AM-007 | `reportcomposition.Service.requireIncidentMembership` | `RequireIncidentMembership` | member, concealed | removed |
| AM-008 | `reportcomposition.Service.requireIncidentRole` | `RequireIncidentRole` | call-site exact roles, concealed | removed |
| AM-009 | `indicatorCreateService.createIndicatorRow` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-010 | `indicators.Store.FindOrCreateIndicatorParticipantTx` | `EnsureOpenTx` | open, caller transaction | removed |
| AM-011 | `indicatorLifecycleService.appendInterval` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-012 | `parties.WorkbookFacade.loadConflictTarget` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-013 | `reporting.ApplicationService.requireIncidentMembership` | `RequireIncidentMembership` | member, concealed | removed |
| AM-014 | `reporting.ApplicationService.requireIncidentRole` | `RequireIncidentRole` | call-site exact roles, concealed | removed |
| AM-015 | `parties.WorkbookFacade.Create` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-016 | `parties.WorkbookFacade.Patch` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-017 | `networkflow.Service.requireIncidentMembership` | `RequireIncidentMembership` | member, concealed | removed |
| AM-018 | `networkflow.Service.requireIncidentRole` | `RequireIncidentRole` | editor-admin or reviewer-admin by route, concealed | removed |
| AM-019 | `workbook.Service.requireIncidentMembership` | `RequireIncidentMembership` | member, concealed | removed |
| AM-020 | `workbook.Service.requireIncidentRole` | `RequireIncidentRole` | call-site exact roles, concealed | removed |
| AM-021 | `imports.Store.ValidateExtensionApplyPreconditionsTx` | `EnsureOpenTx` | open, caller transaction | removed |
| AM-022 | `imports.Store.CreateOperatorRegion` | `AuthorizeMutationTx` | editor-reviewer-admin, open, transaction-bound | removed |
| AM-023 | `imports.Service.requireIncidentMembership` | `RequireIncidentMembership` | member, concealed | removed |
| AM-024 | `imports.Service.requireIncidentRole` | `RequireIncidentRole` | call-site exact roles, concealed | removed |
| AM-025 | `imports.Store.StartApply` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-026 | `imports.Service.applyUnit` | `AuthorizeMutationTx` | editor-reviewer-admin, open, transaction-bound | removed |
| AM-027 | `jobapi.Service.authorizeJob` | `GetIncidentMembershipForUser` | incident-scoped job membership, concealed | removed |
| AM-028 | `tasksdecisions.MutationFacade.loadConflictTarget` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-029 | `tasksdecisions.MutationFacade.SupersedeDecision` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-030 | `incidentbundles.service.handleBundleMember` | `GetIncidentMembershipForUser` | member, bundle-not-found concealment | removed |
| AM-031 | `incidentbundles.service.handleExport` | `GetIncidentMembershipForUser` | member, concealed | removed |
| AM-032 | `tasksdecisions.MutationFacade.Create` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-033 | `tasksdecisions.MutationFacade.Patch` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-034 | `collaboration.Service.incidentClosed` | `GetVisibleIncident` | member plus lifecycle read, concealed | removed |
| AM-035 | `collaboration.Service.requireIncidentMembership` | `RequireIncidentMembership` | member, concealed | removed |
| AM-036 | `revisions/httpapi.Service.requireIncidentMembership` | `RequireIncidentMembership` | member, concealed | removed |
| AM-037 | `artifacts.MutationFacade.executeCreateTx` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-038 | `artifacts.MutationFacade.Patch` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-039 | `artifacts.MutationFacade.loadConflictTarget` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-040 | `savedviews.service.requireIncidentMembership` | `RequireIncidentMembership` | member, concealed | removed |
| AM-041 | `timeline.store.createPerformanceFixtureRows` | `EnsureOpenTx` | open, transaction-bound fixture | removed |
| AM-042 | `timeline.store.createRow` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-043 | `timeline.store.applyPatch` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-044 | `timeline/admission.Service.requireIncidentMembership` | `RequireIncidentMembership` | member, concealed | removed |
| AM-045 | `timeline/admission.Service.requireIncidentRole` | `RequireIncidentRole` | call-site exact roles, concealed | removed |
| AM-046 | `timeline.store.applyOwnerBatchV1` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-047 | `evidence.mutationFacade.Patch` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-048 | `evidence.mutationFacade.Create` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-049 | `evidence.mutationFacade.loadConflictTarget` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-050 | `evidence.blobLifecycleService.CreateBlobSlot` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-051 | `evidence.blobLifecycleService.ClaimUploadLease` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-052 | `evidence.blobLifecycleService.PreflightAttachBlob` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-053 | `evidence.blobLifecycleService.AttachBlob` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-054 | `evidence.blobLifecycleService.QuarantineBlob` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-055 | `evidence.routeAdmission.visibleIncident` | `GetVisibleIncident` | member, concealed | removed |
| AM-056 | `evidence.routeAdmission.requireRole` | `RequireIncidentRole` | call-site exact roles, concealed | removed |
| AM-057 | `incidents.Service.handleIncidentsMember` | `GetVisibleIncident` | member, concealed | removed |
| AM-058 | `incidents.Service.requireIncidentMembership` | `GetIncidentMembershipForUser` | member, concealed | removed |
| AM-059 | `incidents.RequireIncidentMembership` | `GetIncidentMembershipForUser` | member, shared HTTP mapping | removed |
| AM-060 | `evidence.evidenceSourceMutationKernel.createTx` | `EnsureOpenTx` | open, caller transaction | removed |
| AM-061 | `entities.Store.MergeEntity` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-062 | `entities.Service.requireIncidentRole` | `RequireIncidentRole` | call-site exact roles, concealed | removed |
| AM-063 | `hostidentity.Store.PatchEntityRow` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-064 | `hostidentity.Store.loadConflictTarget` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-065 | `hostidentity.Store.ApplyClipboardPastePlan` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-066 | `hostidentity.Store.CreateHostRow` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-067 | `hostidentity.Store.CreateIdentityRow` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-068 | `timeline.store.applyAction` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-069 | `timeline.store.PutTimeConversionProfile` | `EnsureOpenTx` | open, transaction-bound | removed |
| AM-070 | `entities.Store.ApplyMentionAction` | `EnsureOpenTx` | open, transaction-bound | removed |

| Neutral outcome | HTTP mapping owned by `incidents/httpapi` | Default |
| --- | --- | --- |
| `not_visible` | Existing concealed incident `404` family | Never disclose membership or incident existence. |
| `insufficient_role` | Existing `403 authorization_denied` family | Preserve owner-required role details only. |
| `incident_closed` | Existing owner-defined `409 incident_closed` family | Applies only where the owner requires an open incident. |
| Invalid requirement or malformed stored enum | Sanitized `500` | Fail closed; emit no internal value. |
| Unexpected persistence/cancellation failure | Existing sanitized internal/cancellation handling | Never convert to a denial or grant. |

### 3.5 Post-commit collaboration effects

`RB-004` is design-resolved to `internal/app/collaborationassembly/incidenteffects`. Incidents owns mutation and commit results. Collaboration owns current-socket selection, terminal wire objects, delivery attempts, socket closure, and presence removal. Only application assembly may depend on both.

```go
type TerminalMutationDisposition string
const (
	TerminalMutationNewCommit TerminalMutationDisposition = "new_commit"
	TerminalMutationReplay TerminalMutationDisposition = "replay"
)

type TerminalMutationCommit struct {
	Disposition TerminalMutationDisposition
	EffectKey uuid.UUID
}

type TerminalNotifier interface {
	NotifyIncidentClosed(context.Context, uuid.UUID, uuid.UUID)
	NotifyIncidentMembershipRevoked(context.Context, uuid.UUID, uuid.UUID, uuid.UUID)
}
```

For `CloseResult`, `CommitNew` requires a nonzero `EffectKey`; `CommitReplay` requires `EffectKey == uuid.Nil`. No other disposition is valid. `MembershipRemovalResult` represents only a new successful deletion and always requires a nonzero effect key. Repeating deletion of a missing membership remains the current `404 membership_not_found`; the refactor MUST NOT introduce delete idempotency, a replay result, or a no-op result. The effect key is the administrative-audit event UUID committed by the same mutation; it is not a Revisions change-set identifier.

| Mutation outcome | Coordinator invocation | Terminal effect |
| --- | --- | --- |
| New incident close commit | One invocation after commit | Every current incident socket: terminal `error`, `code=incident_closed`, `retryable=false`, then close |
| Exact close replay | None | None |
| New membership deletion commit | One invocation after commit | Removed user's sockets for that incident: `session_revoked`, `reason_code=incident_access_revoked`, then close |
| Repeated/missing membership deletion | None; current error returned | None |
| Role update, legal no-op, rollback, commit failure, or pre-commit error | None | None |

| Requirement ID | Requirement |
| --- | --- |
| `INC-REF-CE-001` | No terminal effect may occur before the authoritative transaction commits. |
| `INC-REF-CE-002` | Production HTTP and non-HTTP callers MUST use the assembled coordinator for close and membership removal; they MUST NOT call raw mutation methods and omit the effect. |
| `INC-REF-CE-003` | Within the single active application process, a current affected connection receives at most one terminal message for one effect key. Collaboration indexes live connections by incident and user, removes matching registrations before/with close, and performs no fallible database lookup. No durable delivery claim is made. |
| `INC-REF-CE-004` | Membership removal MUST preserve the account session and access to every other incident. A role change MUST NOT trigger full revocation unless an adopted owner later requires it. |
| `INC-REF-CE-005` | The process-local notifier has no ordinary error result. Socket write failure follows existing connection cleanup and telemetry and MUST NOT roll back or reinterpret the committed mutation. No retry, readiness gate, or new failure response is introduced. |
| `INC-REF-CE-006` | Crash-survivable delivery, a durable outbox, terminal-effect replay, or changed wire semantics requires later behavior authorization and owner adoption. |

`OG-002` is an authority prerequisite: Core 01/Core 04 must adopt fresh-commit-only,
post-commit, incident-targeted effects and the current process-local delivery boundary.
The durable fix removes the fallible membership-session database lookup through the
incident-user hub index; it deliberately does not invent retry, readiness, or
degradation semantics outside the single-active-process topology.

## 4. Public Contract and Behavior Freeze Map

### 4.1 Exact HTTP operation map

| Operation | Current owner | Frozen behavior summary | Required characterization | Risk |
| --- | --- | --- | --- | --- |
| `GET /api/v1/incidents` | Core 01/Core 04 | Envelope, live membership filtering, query validation, keyset order, session sliding | Membership changes between pages | high |
| `POST /api/v1/incidents` | Core 01/Core 04 | Normalization, actor-scoped idempotency, atomic incident/admin/preferences/audit bootstrap | Failure after each participant and exact replay | high |
| `GET /api/v1/incidents/{incident_id}` | Core 01/Core 04 | Current membership, concealed `404`, deployment-admin non-bypass | Non-member, inactive user, deployment-admin-only | high |
| `PATCH /api/v1/incidents/{incident_id}` | Core 01/Core 04 | Exact fields, validation order, version conflict, no-op, open guard, audit | Denial and validation precedence | high |
| `POST /api/v1/incidents/{incident_id}/close` | Core 01/Core 03/Core 04 | Admin, version/idempotency, committed close/audit, replay, post-commit effect | Commit failure, replay, socket cleanup | high |
| `POST /api/v1/incidents/{incident_id}/reopen` | Core 01/Core 03/Core 04 | Admin, version/idempotency, legal transition, audit, no close effect | Replay and illegal transition | high |
| `GET /api/v1/incidents/{incident_id}/membership-audit-events` | Core 01/Core 04 | Admin, concealed `404`, visible `403`, filters, order, paging, shape | Role and cursor matrix | high |
| `GET /api/v1/incidents/{incident_id}/memberships` | Core 01/Core 04 | Current member visibility, pagination, closed-incident access | Membership loss during paging | high |
| `POST /api/v1/incidents/{incident_id}/memberships` | Core 01/Core 04 | Admin, target resolution, validation, idempotency, audit | Replay and inactive/missing target | high |
| `PATCH /api/v1/incidents/{incident_id}/memberships/{user_id}` | Core 01/Core 04 | Admin, version conflict, last-admin guard, no-op/audit semantics | Concurrent admin count and no-op | high |
| `DELETE /api/v1/incidents/{incident_id}/memberships/{user_id}` | Core 01/Core 04 | Admin, base version, last-admin guard, `204`, missing/repeated `404`, audit, post-commit revocation | Commit failure, repeated delete, target isolation | high |

All 11 operations MUST retain their request and response envelopes, validation and normalization order, authorization outcomes, version/change-set semantics, pagination, storage atomicity, audit behavior, session sliding, and generated operation bindings. The target owns no WebSocket route.

### 4.2 Complete contract map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `GET /api/v1/incidents` | Core 01; `module.incidents` projection | OpenAPI owner, routes, list application | Integration, pagination, browser incident directory | Preserve live membership re-evaluation, keyset ordering, query validation, session slide, and envelopes | high | No cached authorization snapshot. |
| `POST /api/v1/incidents` | Core 01/Core 04; incidents application | OpenAPI, create decoder/application, SQL | Unit, store, integration, browser create | Preserve normalization, idempotency scope/hash, atomic incident/admin/preferences/audit bootstrap, replay, and commit-failure behavior | high | Creator becomes admin atomically. |
| `GET /api/v1/incidents/{incident_id}` | Core 01/Core 04 | OpenAPI, routes, access | HTTP conformance and integration | Preserve hidden `404` for non-members and deployment-admin non-bypass | high | Visible incident membership is required. |
| `PATCH /api/v1/incidents/{incident_id}` | Core 01/Core 04 | OpenAPI, decoder, application | Request, store, integration, frontend | Preserve patchable fields, unknown/reserved field rejection, no-op semantics, version conflicts, closed guard, audit, and session behavior | high | Saved-view and workbook preference fields remain forbidden. |
| `POST .../{incident_id}/close` and `/reopen` | Core 01/Core 03/Core 04 | OpenAPI, lifecycle application/routes | Lifecycle unit/integration, frontend/workbook tests | Preserve admin authorization, idempotency, audit, state transition, replay, and post-commit close notification | high | Close is read-only lifecycle, not delete or purge. |
| `GET .../{incident_id}/membership-audit-events` | Core 01/Core 04 | OpenAPI, audit handler, platform audit projection | Audit contract/integration and frontend tests | Preserve admin gate, hidden `404`, insufficient-role `403`, filters, ordering, paging, and resource shape | high | Deployment-local audit only. |
| `GET|POST .../{incident_id}/memberships` | Core 01/Core 04 | OpenAPI, routes, membership application | Request, store, integration, browser admin tests | Preserve current-role authorization, target resolution, exact request validation, idempotency, audit, paging, and closed-incident administration | high | Closed incidents still permit membership administration. |
| `PATCH|DELETE .../memberships/{user_id}` | Core 01/Core 04 | OpenAPI, routes, membership application | Unit, store, integration, browser tests | Preserve version conflicts, last-admin guard, PATCH no-op behavior, DELETE missing/repeated `404`, audit, and post-commit target-user revocation signal | high | DELETE has no client transaction identifier and no replay result; account session remains valid after incident membership loss. |
| Collaboration close and membership-revocation effects | Collaboration owner consumes incidents signals | `CollaborationSessionPort`, route notifications, server assembly, web collaboration session | Collaboration socket/stream tests, lifecycle and browser tests | Add or retain explicit no-notify-before-commit, no duplicate notify on replay, and correct target/incident signal tests | high | No WebSocket route is owned by incidents. |
| Incident access used by entity, evidence, import, saved-view, reporting, timeline, workbook, and other routes | Core 04 semantics; incidents adapter | `Access`, shared require helpers, inbound import scan | Peer route/integration tests and incident access unit row | Preserve current membership/role re-derivation, hidden/forbidden outcomes, open-state admission, and transaction-bound checks | high | Interface refactor has a wide blast radius. |
| Workbook startup preferences and initial references | Core 01/Core 03; workbook startup behavior | Preference SQL, incidents bootstrap, workbook startup ports and assembly | Incident store, workbook startup, frontend workbook tests | Preserve creator/import defaults, no-op/update semantics, canonical sheet refs, closed readability, and transaction atomicity | high | Physical ownership may move; behavior may not. |
| Incident bundle `data/incident.json` and final publication | Core 01 and incident-bundle contracts | Source catalog, traceability, codec, source port, finalizer | Bundle round-trip/atomicity/actor tests | Preserve schema version, digest/validation, initial admin, preference bootstrap, audit, and single publication transaction | high | Generated/source catalogs are not hand edited. |
| Reporting incident snapshot and source boundary | Reporting NLSpec plus incidents/Revisions owner data | Reporting provider and export materializer | Reporting integration/evidence provider tests | Characterize exact incident snapshot fields and high-water selection before changing provider dependencies | medium | Direct change-set read is a boundary concern, not permission to alter snapshot semantics. |
| Administrative audit versus Revisions history | Core 01/Core 02/Core 04 | Audit mapping, Reporting provider, domain vocabulary | Audit, store, reporting, and mutation helper tests | Preserve that incident/membership admin audit is deployment-local and not a Revisions change set | high | No conversion between histories. |
| Generated OpenAPI, protocol, SQL, audit, and recovery surfaces | Authored contract/SQL owners | OpenAPI owner, SQL inputs, registries, generated files | OpenAPI parity, drift, JSON-shape, recovery tests | Regenerate only after owner-input changes and require drift checks | high | Never hand-edit generated roots. |
| Frontend incident landing/admin/workbook behavior | Web/workbook/collaboration owners | Representative TSX consumers and generated protocol bindings | Five Vitest and seven Playwright owner rows plus collaborator tests | Preserve incident selection, membership administration, closed shell state, and collaboration invalidation | medium | No grid-adapter or UI-selector implementation exists in the target. |
| Harness/test accounting | Testing Harness owner | Verification owner, test catalog, family manifest | 41 rows: 29 Go, 7 Playwright, 5 Vitest | Update only if test identity, selector, ownership, or evidence routing changes | medium | Evidence accounting does not define runtime architecture. |
| Projection refresh and saved-view/view-schema behavior | Projection, saved-view, and workbook owners | Absence from target plus peer consumers | Peer module and workbook suites | Preserve current incident admission and closed/read behavior; do not add a refresh hook without owner evidence | medium | No direct projection-refresh implementation was found in incidents. |

### 4.3 Requirement-to-evidence map

| Requirement group | Minimum executable evidence before movement | Harness/accounting posture |
| --- | --- | --- |
| `INC-REF-WB-*` | Incident create/import success, failure rollback, exact replay, and ordinary preference regression | Preserve Incidents and Workbook owner rows. |
| `INC-REF-SB-*` | Null/one/many/tied change sets, same-snapshot concurrency, canonical byte/token goldens, and replay | Preserve Reporting/Revisions rows; add identity only if required. |
| `INC-REF-AD-*` | Role-set table, concealment, malformed enums, deployment-admin non-bypass, per-consumer route checks, and transaction checks | Consumer rows remain with their behavior owners. |
| `INC-REF-CE-*` | Zero before commit, one coordinator call for new commit, zero for replay/rollback, target isolation, and socket-failure cleanup | Preserve Collaboration and Incidents rows. |
| HTTP operation map | Runtime/OpenAPI parity and all 11 operation behavior families | `module.incidents` remains the evidence-accounting owner. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Shared incident admission returned platform `httpapi.APIError` | Activation-time `lifecycle_access.go` and peer route imports | Transport semantics leaked through a domain-owner facade and made migration broad | `resolved_rs01` | `incidents/admission`; owner-local HTTP mappings pending package extraction only | RS-01 removed the legacy facade and migrated all 70 consumers; RS-02 will move the remaining Incidents-owned transport mapping without changing outcomes. |
| Collaboration notifications were initiated in HTTP route handlers after application calls | Activation-time `routes.go`, `CollaborationSessionPort`, server assembly | Non-HTTP callers could omit required effects; replay/commit ordering could drift | `resolved_rs03` | `collaborationassembly/incidenteffects` | RS-03 added fresh/replay results, committed effect keys, composition validation, and incident/user target isolation. |
| Owner HTTP handlers and decoding lived in the root package | Activation-time `api.go`, `routes.go`, `membership_audit_handlers.go`; neighboring `httpapi` packages | Root facade was wide and transport-adjacent | `resolved_rs02` | `incidents/httpapi` | RS-02 moved transport mechanics and white-box tests while retaining all 11 operation identities and behavior. |
| `Access` exposed concrete PostgreSQL and `pgx.Tx` mechanics | Activation-time `access.go`; 70-method inbound caller scan | Storage mechanics became a public cross-module dependency | `resolved_rs01` | Concrete transport-neutral checker with consumer-owned narrow interfaces | Legacy calls and imports are absent; transactional consumers use `CheckTx` or `RequireOpenTx`. |
| Workbook preference persistence was nested under incidents | Activation-time `incidents/workbookpreferences`, Workbook ports, and app adapter | Semantic ownership and physical ownership diverged | `resolved_rs04` | `workbook/startup/postgres` behind `bootstrapport` | RS-04 moved repository/session/bootstrap, injected the narrow writer, renamed authored SQL, and removed the old package without a schema or SQL-semantic change. |
| Reporting provider read Revisions `change_sets` directly | Activation-time `reportingprovider/provider.go` | Incidents adapter knew another owner's persistence representation | `resolved_rs05` | `revisions/sourceboundary` injected into Reporting composition | Dual characterization passed before the old query and duplicate serializer were deleted; a boundary rule now forbids Incidents production access to `change_sets`. |
| Incidents mutation/audit queries are published under Incidents test support | Direct imports of `testsupport/mutationtest`; Incidents event vocabulary | The package is cohesive and no cross-owner coupling exists | `intentional/no_action` | Incidents test support | Retain; add a boundary rule only if future non-Incidents production or test callers appear. |
| Private incident/member SQL remains behind owner application/access adapters | `repository.go`, authored queries | Low if encapsulation remains; high if exposed during refactor | `intentional/no_action` | Incidents private persistence | Preserve facade and transaction behavior. |
| Bundle, recovery, Reporting snapshot, and Network Flow lock adapters are source-owner contributions | Source ports, assemblies, adopted subsystem boundaries | Moving them to generic coordinators would invert ownership | `intentional/no_action` | Incidents owner contribution packages | Keep narrow; change only dependency leakage identified separately. |
| Incident-specific route, scenario, store, and performance fixtures are shared | `testsupport/**` callers | Test imports are broad but semantically coherent | `intentional/no_action` | Incidents test support | Keep incident-specific setup close to the owner. |
| Generated files project authored contracts | Generated artifact policy and discovered generated surfaces | Hand edits would drift and violate repository policy | `intentional/no_action` | Contract generators | Change authored inputs, then use `make generate` in a later authorized task. |
| `.gitkeep` remained in a populated directory | Activation inventory | No runtime risk; created inventory noise | `resolved_rs07` | Repository maintenance | RS-07 deleted the obsolete placeholder and recalculated the 45-file final inventory. |
| Frontend, grid, projections, saved views, entities, indicators, evidence, links, and timeline are not implemented in the target | Direct target inspection and inbound caller scan | Moving based on route names or UI labels would invent ownership | `defer` | Existing owners | Preserve contracts; propose no relocation without new evidence. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | None | WF-01, WF-02 | Fix target, authority, snapshot, constraints, and handoff artifact | This tracker; owner documents | `make lint-markdown` for tracker edits | Scope, authority, and only-write rule are explicit. |
| WF-01 | Complete target and caller inventory | parallel | WF-00 | WF-03, WF-04 | Maintain one live row per target file and map inbound/outbound edges | `internal/modules/incidents/**`; assemblies and consumers | `find`, `rg`; later owner slice tests | All 43 files and material caller groups are accounted for. |
| WF-02 | Contract-owner mapping | parallel | WF-00 | WF-03, WF-04 | Map routes, auth, audit, collaboration, portability, preferences, reporting, generated surfaces, and harness rows to owners | Core/NLSpecs, contracts, generated evidence, SQL, harness manifests | OpenAPI parity and drift targets in section 8 | Every observable risk has an owner and evidence location. |
| WF-03 | Characterization test gap analysis | parallel | WF-01, WF-02 | WF-05 | Determine which current behaviors need tests before movement | Incident and collaborator tests; owner family manifest | `make test-slice OWNER=module.incidents`; service-backed slice | Gaps are either backed by exact tests or tracked as implementation prerequisites. |
| WF-04 | Boundary and coupling scan | parallel | WF-01, WF-02 | WF-05 | Classify transport, persistence, cross-owner, generated, and test-support couplings | Access, routes, ports, preferences, reporting provider, test support | `make backend-module-boundary-check`; `make lint-go`; dependency review | Every finding has classification, owner, and planning action. |
| WF-05 | Facade and ownership redesign | chain | WF-03, WF-04 | WF-06 | Apply the four resolved designs without changing behavior | Incidents root/admission/http adapters; Workbook/Revisions ports; application assembly | Characterization plus owner review | RB-001 through RB-004 are design-resolved; SP-01/SP-02 closed the owner gates. |
| WF-06 | Safe slice sequencing | chain | WF-05 | WF-07 | Turn the redesign into reversible, behavior-preserving slices | Files named per slice in section 7 | Per-slice owner tests and drift checks | Each slice has dependency, rollback, and binary exit criterion. |
| WF-07 | Harness/test/accounting update plan | chain | WF-06 | WF-08 | Preserve evidence routing and update authored manifests only if identities change | Tests; verification/test-family owner inputs | Harness validation plus owner slice explanations | No phase map or row is used as runtime architecture; conditional edits are explicit. |
| WF-08 | Validation and final handoff | chain | WF-07 | None | Run narrow-to-broad verification and leave continuation evidence | All authorized slice files and this tracker | `make agent-finalize`, then `make check` after narrow targets | Results, failures, artifacts, rollback state, and remaining blockers are recorded. |

## 7. Proposed Refactor Slice Plan

All default slices are behavior-preserving. Any discovered need to change a route, envelope, authorization outcome, persistence semantic, event semantic, generated contract, or UI state is **requires later authorization** and MUST stop before implementation.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| RS-00 | None | Capture current route, admission, audit, effect, bundle, preference, and source-boundary behavior; production changes are prohibited | Existing tests only | Encoding desired rather than current behavior | Section 4 matrices, including repeated membership delete and canonical boundary bytes | Incident unit/service slices and affected collaborator rows | Revert characterization only | Every later risk has passing baseline evidence or an explicit authority gate. |
| RS-01 | RS-00 | Add neutral admission enums/denials, narrow readers and transaction capabilities, and migrate callers through compatibility stages; HTTP/package moves are prohibited | Access/lifecycle, new admission package, callers, assembly | Concealment, role sets, rechecks, precedence | Per-consumer matrix, malformed input, role-set, closed/open, transaction tests | Incident and affected owner slices; boundary check; `make lint-go` | Keep legacy facade until each caller batch passes; revert a batch independently | Every production consumer is migrated; no domain/application consumer receives HTTP errors. |
| RS-04 | RS-00 | Create Workbook bootstrap contract/adapter, migrate create/import, move ordinary persistence, update authored boundary rules; schema/SQL behavior changes are prohibited | Workbook Startup, Incidents preferences/application/finalizer, assembly, boundary manifest | Atomicity, replay, timestamps, null defaults, cycles | `INC-REF-WB-*` and ordinary preference routes | Incident/Workbook service slices; boundary check; generated drift | Preserve old adapter until all callers pass; revert without migration | Workbook owns persistence, Incidents imports only bootstrap port, behavior unchanged. |
| RS-05 | RS-00, OG-001 | Add Revisions resolver and inject it into Reporting; changing canonical bytes/token or snapshot is prohibited | Revisions sourceboundary, Incidents provider, Reporting store/materializer, assembly | Immutable snapshot, deterministic latest row, byte drift | `INC-REF-SB-*` matrix | Reporting/Revisions/Incidents service slices; boundary check; generated drift | Keep direct query until byte-for-byte dual characterization passes | Incidents no longer reads Revisions persistence and bytes/tokens match. |
| RS-03 | RS-01, OG-002 | Add typed commit results and assembled post-commit coordinator; index the local hub by incident/user; durable delivery and new wire effects are prohibited | Incidents application/audit, Collaboration notifier/hub, assembly, HTTP adapter | Commit timing, close replay, delete semantics, socket cleanup, target isolation | `INC-REF-CE-*` matrix | Incident/Collaboration unit/service slices; affected browser rows | Retain route notifier until coordinator parity, then remove in this slice | Only new commits invoke coordinator; current wire and mutation outcomes remain. |
| RS-02 | RS-01, RS-03 | Extract registration, decoding, admission-to-HTTP mapping, paging, and response building; remove compatibility after all callers migrate | Incidents API/routes/audit handlers/tests, server assembly | All 11 bindings, envelopes, validation/session ordering | Operation map, OpenAPI parity, HTTP conformance, browser evidence | Incident unit/service slices; generated drift; boundary check | Keep old registrar until full parity; revert as one package move | Root Incidents owns no HTTP mechanics; all 11 operations remain compatible. |
| RS-06 | RS-02, RS-04, RS-05 | Remove all temporary admission, Workbook, Reporting, effect, and route shims; retain Incidents-specific test helpers | Final package imports, compatibility aliases, boundary owner inputs | Cycles, dual APIs, evidence identity | Retired-symbol searches, affected owner slices, and harness contract | Affected owner slices; `make harness-contract`; `make lint-go` | Revert individual boundary rules only with recorded evidence | No compatibility surface remains; Incidents-specific fixtures and mutation assertions stay local. |
| RS-07 | RS-06 | Remove `.gitkeep`, regenerate only from changed owner inputs, conditionally update harness routing, and finalize | Placeholder; authored/generated inputs only if needed | Drift, stale imports, missing evidence | All retained owner/collaborator evidence | Drift/policy/JSON checks; `make agent-finalize`; `make check` | Revert cleanup, owner inputs, and projections separately | No placeholder/import/drift remains; final results are recorded. |

The dependency graph is `RS-00 -> {RS-01, RS-04, RS-05}`, `RS-01 -> RS-03`, `{RS-01, RS-03} -> RS-02`, `{RS-02, RS-04, RS-05} -> RS-06`, and `RS-06 -> RS-07`. RS-04 and RS-05 MAY proceed in parallel with RS-01 after RS-00; no other dependency may be omitted.

## 8. Validation Plan

These commands were discovered through repository-owned help and explanation targets.
At planning time, no product target had run. The authorized implementation subsequently
executed the narrow-to-broad sequence below; section 8.1 is the final RS-07 evidence.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| documentation | `make lint-markdown` | Authored Markdown | yes | Required for this tracker-only revision. |
| unit | `make test-slice OWNER=module.incidents` | Focused owner rows runnable through the normal incident owner slice, including Go/Vitest unit evidence | yes | Use `ROWS=<row-id,...>` from `tools/test_families/module.incidents.json` for a narrower slice when a workflow touches only named rows. |
| integration | `make service-backed-test-slice OWNER=module.incidents` | PostgreSQL/browser-backed incident owner rows | yes | Establish baseline before high-risk application, persistence, or authorization changes. |
| e2e/browser | `make browser-e2e-webserver-backed` | Webserver-backed incident directory, administration, collaboration, and workbook behavior | no | Required before completion of slices that affect routes, collaboration, or frontend-visible contracts; narrow to owner row IDs when practical. |
| generated drift | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check` | Generated Go/TypeScript, authored contract projections, and JSON shapes | yes | Baseline drift is required before owner-input changes; never hand-edit generated roots. |
| import-boundary/static | `make backend-module-boundary-check`; `make lint-go` | Exact owner-port, forbidden-import, and Go static boundaries | yes | Add `make frontend-import-boundary-check` only if a later authorized slice touches frontend files. |
| harness | `make harness-contract` | Owner/test routing and generated topology | no | Required only if test identity or routing changes. |
| full check | `make agent-finalize`, then `make check` | Repository-wide finalization and broad check | no | Required before final implementation handoff, after narrow targets pass. |

### 8.1 RS-07 Final Validation Evidence

All commands below ran from the repository root through public Make targets. Paired
owner entries mean the focused command ran first and the service-backed command ran
second. Run-root values are relative to `.cartulary/test-results/`.

| Command or owner | Final result | Run root or evidence |
| --- | --- | --- |
| `make generate` | PASS; authored inputs regenerated without hand edits | `20260819T023526Z-p3995368` |
| `module.incidents` focused / service-backed | PASS `26/26`; PASS `19/19` | `20260819T023556Z-p4001279`; `20260819T023716Z-p4043851` |
| `module.workbook` focused / service-backed | PASS `67/67`; PASS `39/39` | `20260819T023836Z-p4086350`; `20260819T024047Z-p4141974` |
| `module.revisions` focused / service-backed | PASS `27/27`; PASS `20/20` | `20260819T024256Z-p3397`; `20260819T024358Z-p46247` |
| `module.reporting` focused / service-backed | PASS `5/5`; PASS `4/4` | `20260819T025108Z-p209954`; `20260819T024742Z-p103960` |
| `module.collaboration` focused / service-backed | PASS `32/32`; PASS `23/23` | `20260819T024821Z-p118983`; `20260819T024943Z-p164483` |
| `module.artifacts` focused / service-backed | PASS `7/7`; PASS `3/3` | `20260819T025154Z-p224680`; `20260819T025230Z-p239744` |
| `module.entities` focused / service-backed | PASS `32/32`; PASS `29/29` | `20260819T025306Z-p254355`; `20260819T025445Z-p306566` |
| `module.evidence` focused / service-backed | PASS `35/35`; PASS `25/25` | `20260819T025624Z-p358322`; `20260819T025740Z-p406970` |
| `module.imports` focused / service-backed | PASS `22/22`; PASS `14/14` | `20260819T025901Z-p454197`; `20260819T030010Z-p495113` |
| `module.incidentbundles` focused / service-backed | PASS `8/8`; PASS `6/6` | `20260819T030205Z-p535218`; `20260819T030300Z-p550359` |
| `module.indicators` focused / service-backed | PASS `13/13`; PASS `7/7` | `20260819T030355Z-p565170`; `20260819T030432Z-p580770` |
| `module.jobapi` focused / service-backed | PASS `3/3`; PASS `3/3` | `20260819T030515Z-p595632`; `20260819T030552Z-p610279` |
| `module.networkflow` focused / service-backed | PASS `35/35`; PASS `30/30` | `20260819T030629Z-p624903`; `20260819T030839Z-p679675` |
| `module.parties` focused / service-backed | PASS `12/12`; PASS `12/12` | `20260819T031047Z-p733978`; `20260819T031133Z-p771524` |
| `module.reportcomposition` focused / service-backed | PASS `3/3`; PASS `3/3` | `20260819T031244Z-p809397`; `20260819T031319Z-p824018` |
| `module.savedviews` focused / service-backed | PASS `25/25`; PASS `24/24` | `20260819T031354Z-p838631`; `20260819T031507Z-p885708` |
| `module.tasksdecisions` focused / service-backed | PASS `18/18`; PASS `15/15` | `20260819T031624Z-p932693`; `20260819T031710Z-p970793` |
| `module.timeline` focused / service-backed | PASS `48/48`; PASS `29/29` | `20260819T031757Z-p1008505`; `20260819T032234Z-p1065849` |
| `app.server` focused / service-backed | PASS `24/24`; PASS `17/17` | `20260819T032711Z-p1121448`; `20260819T032807Z-p1159978` |
| `make browser-e2e-webserver-backed` | PASS `58/58` | `20260819T032906Z-p1198142` |
| `make backend-module-boundary-check` | PASS `3/3` | `20260819T033314Z-p1249297` |
| `make lint-go` | PASS | Make target exited zero; no separate graph root emitted. |
| `make harness-contract` | PASS `2/2` | `20260819T033321Z-p1253964` |
| `make generate-drift` | PASS `4/4` | `20260819T033338Z-p1254491` |
| `make generated-artifact-policy-check` | PASS `3/3` | `20260819T033347Z-p1257365` |
| `make json-shape-check` | PASS `3/3` | `20260819T033348Z-p1257769` |
| `make lint-markdown` | PASS | `20260819T033351Z-p1258251` |
| `env -u RESULTS_DIR make agent-finalize` | PASS `1/1`; no generated mutation | `20260819T033356Z-p1259059` |
| `make check` | PASS `626/626` | `20260819T033413Z-p1261931` |

The first final focused Reporting attempt failed `4/5` at
`20260819T024458Z-p88772` with harness classification `infra/fixture_error` because
object-store readiness expired before the product row began. The service-backed run
and a clean focused rerun passed, so no implementation correction or unresolved risk
is attributed to that failure.

`agent-finalize` was intentionally invoked without `RESULTS_DIR`. Its finalizer summary
records `results-dir-not-provided`; retained canonical-evidence and scheduler-drift
maintenance were skipped, while JSON shape, catalog tier, and generated-structure
checks passed. No successful retained warm-check root was deliberately reused.

After the RS-07 tracker closure update, `make lint-markdown` and
`git diff HEAD --check -- docs/handoffs/incidents-module-refactor-tracker.md` MUST pass
before handoff. No later workstream exists.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| IT-001 | Framework-first authority review | WF-00 | DONE | None | Section 1 | Framework and research are correctly subordinate. |
| IT-002 | Target/snapshot/execution-boundary confirmation | WF-00 | DONE | IT-001 | Git and 43-file inventory | Live authorization and staged tracker posture are explicit. |
| IT-003 | File/caller inventory | WF-01 | DONE | IT-002 | Section 2 | All 43 activation files and all 45 final files are mapped with material callers and dispositions. |
| IT-004 | Owner/contract/generated/harness mapping | WF-02 | DONE | IT-002 | Sections 1 and 4 | All 11 operations and non-route contracts are mapped. |
| IT-005 | NLSpec-grade boundary diagnosis | WF-04 | DONE | IT-003, IT-004 | Sections 3 and 5 | Exact owners and dispositions are recorded. |
| IT-006 | Resolve Workbook ownership design (`RB-001`) | WF-05 | DONE | IT-005 | `INC-REF-WB-*` | Interface, dependency, defaults, boundary edits, tests, and rollback are explicit. |
| IT-007 | Resolve Revisions source-boundary design (`RB-002`) | WF-05 | DONE | IT-005 | `INC-REF-SB-*` | Interface, bytes, selection, snapshot, tests, and rollback are explicit. |
| IT-008 | Resolve neutral admission design (`RB-003`) | WF-05 | DONE | IT-005 | `INC-REF-AD-*` | Nonordinal roles, mappings, migration, and invalid defaults are explicit. |
| IT-009 | Resolve collaboration ownership design (`RB-004`) | WF-05 | DONE | IT-005 | `INC-REF-CE-*` | Result variants, effect key, invocation matrix, and delivery boundary are explicit. |
| IT-010 | Complete characterization baseline | WF-03 | DONE | IT-003, IT-004 | RS-00 | Exact source bytes, repeatable-read visibility, replay, effects, admission edges, and owner baselines pass. |
| IT-011 | Adopt exact Core 02 source-boundary details | WF-05 | DONE | IT-007 | OG-001 | Owner text, schema, and vectors adopt exact bytes, selection, and replay rules. |
| IT-012 | Adopt Core 01/Core 04 terminal-effect semantics | WF-05 | DONE | IT-009 | OG-002 | Owner text defines fresh-commit ordering, target isolation, and process-local delivery. |
| IT-013 | Implement neutral admission migration | WF-06 | DONE | IT-010 | RS-01 | All 70 consumers use neutral admission; legacy symbols are absent; owner and structural gates pass. |
| IT-014 | Implement Workbook persistence relocation | WF-06 | DONE | IT-010 | RS-04 | Workbook owns preference persistence; injected create/import bootstrap remains atomic and behavior-compatible. |
| IT-015 | Implement Revisions source-boundary port | WF-06 | DONE | IT-010, IT-011 | RS-05 | Direct Incidents persistence read and duplicate Reporting serialization are removed with exact byte/token parity. |
| IT-016 | Implement post-commit effects coordinator | WF-06 | DONE | IT-012, IT-013 | RS-03 | Fresh/replay audit-key results, post-commit ordering, target isolation, composition validation, and socket effects pass. |
| IT-017 | Extract HTTP adapter and remove compatibility | WF-06 | DONE | IT-013, IT-016 | RS-02 | All 11 operations pass from `incidents/httpapi`; root application results and imports are transport-neutral. |
| IT-018 | Remove compatibility residue and retain cohesive test support | WF-06 | DONE | IT-014, IT-015, IT-017 | RS-06 | Retired surfaces are absent and boundary-forbidden; Incidents-specific fixtures remain. |
| IT-019 | Update harness only if identity/routing changes | WF-07 | DONE | IT-018 | Test-family owner inputs | RS-02 selector paths and one fixture classification were updated from authored inputs; topology and harness contract pass. |
| IT-020 | Final implementation validation/handoff | WF-08 | DONE | IT-018, IT-019 | RS-07 and section 8.1 | Narrow, owner-matrix, browser, structural, drift, finalization, and `626/626` broad results are recorded. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-18T17:23:50-04:00 | Codex / tracker creation | Framework-first planning complete; tracker created as the only write | Inspected `AGENTS.md`, framework, domain, Core 00-04, relevant adopted NLSpecs and guide; touched only this tracker | `sed`, `rg`, `find`, `git status --short --branch`, `git rev-parse HEAD` | Authority order and clean snapshot confirmed; no owner contradiction found | None for tracker completion | Obtain later authorization before any implementation. |
| 2026-08-18T18:07:14-04:00 | Codex / NLSpec-grade revision | Research recommendations reconciled with adopted owners and live state | Inspected `nlspec-spec.md`, `analysis-notes.md`, framework, tracker, Core/owners; touched only this tracker | `sed`, `rg`, `find`, `git status`, `git rev-parse`, `date` | Normative posture and two owner-authoring gates recorded; no owner contradiction | OG-001 and OG-002 gate affected implementation only | Begin RS-00 in a later authorized task. |
| 2026-08-18T18:48:03-04:00 | Codex / WF-00 activation | Approved remediation activated; live snapshot and execution ledger added without replacing or unstaging prior tracker history | Inspected live Git state, tracker claims, and direct `mutationtest` imports; touched only this tracker | `git status`, `git diff --cached --stat`, `rg`, `sed`, `date`, documentation gates | Stale test-support and notifier-failure premises corrected; implementation boundary is explicit | None after tracker gates | Begin SP-01. |
| 2026-08-18T18:54:15-04:00 | Codex / SP-01 | Exact v1 source-boundary ownership adopted and projected | Core 01, Core 02, Core 04 AC-266, `contracts/revisions` schema/vector/index, tracker | `make lint-markdown`; `make json-shape-check`; vector hash verification; scoped diff checks | Exact zero/one/many/tied bytes and tokens pass; repeatable-read and replay rules are explicit | None; `OG-001` closed | Begin SP-02. |
| 2026-08-18T18:56:20-04:00 | Codex / SP-02 | Fresh-commit terminal-effect semantics adopted | Core 01 REQ-01-592; Core 04 REQ-04-017, AC-162, and AC-419; tracker | `make lint-markdown`; scoped diff checks | No-effect outcomes, post-commit ordering, incident/user isolation, effect identity, and process-local delivery are explicit | None; `OG-002` closed | Begin RS-00. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-18T17:23:50-04:00 | Codex / tracker creation | All 43 target files inventoried; legitimate facade plus mixed responsibilities diagnosed | Inspected every file under `internal/modules/incidents`, app assemblies, neighboring `httpapi` patterns, workbook startup, reporting consumer; touched only this tracker | `find`, `wc`, `rg`, `sed` | Access, transport, collaboration, preferences, reporting, owner adapters, and test-support boundaries mapped | RB-001 through RB-004 block affected implementation slices | Begin RS-00 characterization in a later task. |
| 2026-08-18T18:07:14-04:00 | Codex / NLSpec-grade revision | Four package/interface decisions are design-complete | Re-read target access, lifecycle, applications, routes, preferences, provider, assembly patterns, audit IDs, boundary manifest; touched only tracker | `rg`, `sed`, `jq` | Exact Workbook, Revisions, admission, HTTP, and effects ownership/dependency rules recorded | OG-001 and OG-002 only | Characterize, then execute the section 7 dependency graph. |
| 2026-08-18T21:17:19-04:00 | Codex / RS-04 | Workbook preference ownership relocation complete | Added `workbook/startup/bootstrapport` and `workbook/startup/postgres`; changed Incidents application/finalizer/options, Workbook and server assembly, SQL/codegen boundary inputs, generated SQL projection, tests, and boundary policy; deleted `incidents/workbookpreferences` | Incidents `26/26` and `19/19`; Workbook `67/67` and `39/39`; app.server `24/24` and `17/17`; boundary `3/3`; Go lint, harness `2/2`, drift `4/4`, artifact policy `3/3`, and JSON `3/3` pass | Initial boundary run `20260819T010530Z-p2309733` exposed the expected broad Workbook token rule; exact Incidents capability-owner paths were added while root/implementation imports remain forbidden | Begin RS-05. |
| 2026-08-18T21:59:49-04:00 | Codex / RS-03 | Post-commit terminal-effect ownership complete | Added typed Incidents commit dispositions and committed audit effect keys; added `collaborationassembly/incidenteffects`; injected one shared application and coordinator; reindexed Collaboration revocation by incident/user; removed notifier database access and route-owned notifications; added owner rows and boundary rules | Incidents focused `26/26` (`20260819T015130Z-p3014684`) and service `19/19` (`20260819T015518Z-p3144939`); Collaboration focused `32/32` (`20260819T015253Z-p3058864`) and service `23/23` (`20260819T015638Z-p3187351`); app.server `24/24` (`20260819T015419Z-p3105939`) and `17/17` (`20260819T015800Z-p3232689`); boundary `3/3`, Go lint, and harness `2/2` (`20260819T015916Z-p3270928`) pass | Initial audit-identity row used the projection column name against the raw table (`20260819T014823Z-p2978739`); initial harness run recorded the expected transaction-row count delta (`20260819T015111Z-p3013869`). Both authored expectations were corrected. | Begin RS-02. |
| 2026-08-18T22:17:05-04:00 | Codex / RS-02 | Incidents HTTP adapter extraction complete | Added `incidents/httpapi`; moved registrar, handlers, decoder/error mapping, paging, membership-audit transport, and six white-box test files; added transport-neutral root contracts/idempotency support; removed root status/location results; updated server composition, owner selectors, fixture policy, boundary policy, and generated topology | Incidents focused `26/26` (`20260819T021156Z-p3334776`) and service `19/19` (`20260819T021322Z-p3379200`); app.server `24/24` (`20260819T021441Z-p3421782`) and `17/17` (`20260819T021536Z-p3460444`); boundary `3/3`, harness `2/2`, Go lint, generation, drift `4/4`, artifact policy `3/3`, and JSON `3/3` pass | Intermediate Go lint correctly rejected incomplete symbol movement while the package split was in progress; no owner contradiction or public behavior change was found | Begin RS-06. |
| 2026-08-18T22:34:20-04:00 | Codex / RS-06 | Compatibility and test-support cleanup complete | Audited retired symbols/imports; added regression rules for effect routing, root registrar/status/location, old Workbook adapter/SQL name, and old Reporting resolver; retained Incidents-local mutation, performance, route, scenario, and store support | Incidents `26/26` and `19/19`; Workbook `67/67` and `39/39`; Revisions `27/27` and `20/20`; Reporting `5/5` and `4/4`; Collaboration `32/32` and `23/23`; app.server `24/24` and `17/17`; boundary `3/3`, harness `2/2`, Go lint, format, and repository diff checks pass | None; no compatibility shim or generic test-support relocation remains | Begin RS-07. |
| 2026-08-18T23:41:00-04:00 | Codex / RS-07 | Refactor and final handoff complete | Deleted obsolete `.gitkeep`; recalculated the 45-file final target; reviewed all specification, contract, SQL, generated, assembly, module, test, boundary, and harness changes; preserved the user-staged tracker history | Section 8.1 records all final results; `make check` passed `626/626` at `20260819T033413Z-p1261931` | None; no migration, frontend source, public HTTP/OpenAPI/WebSocket, persisted v1 byte, or deployment-topology change occurred | Handoff complete; no next workstream. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-18T17:23:50-04:00 | Codex / tracker creation | Frontend is an indirect consumer, not part of the target implementation | Inspected incident landing/admin, collaboration session, workbook shell and invalidation files; touched only this tracker | `rg`, `sed` | Incident selection, administration, closed state, and collaboration effects identified; no target-owned grid/UI selector code found | None for planning | Preserve existing Vitest/Playwright behavior during backend slices. |
| 2026-08-18T18:07:14-04:00 | Codex / NLSpec-grade revision | Frontend remains indirect evidence only | Reused live frontend consumer and owner-row evidence; touched only tracker | `rg`, `jq` | No frontend, grid, projection, selector, or saved-view relocation authorized | None | Run affected browser rows only when a later slice creates visible-behavior risk. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-18T17:23:50-04:00 | Codex / tracker creation | Eleven OpenAPI operations and authored/generated projections mapped | Inspected incident OpenAPI owner, bundle catalog/traceability, verification owner, SQL inputs, generated OpenAPI/protocol/SQL/audit/recovery evidence; touched only this tracker | `jq`, `rg`, `sed`, `find`; `make explain-target TARGET=generate-drift DETAIL=summary`, generated policy and JSON-shape explanations | Authored inputs and generated outputs distinguished; no generated file was edited | None for tracker; RB-002 affects future Reporting port design | Run drift targets only in an authorized implementation or validation task. |
| 2026-08-18T18:07:14-04:00 | Codex / NLSpec-grade revision | Exact route and source-boundary freeze maps authored | Inspected OpenAPI evidence, Reporting transaction/provider, current canonical members and token path; touched only tracker | `rg`, `sed`, `jq` | Eleven operations retained; current `latest_*` keys and read/write repeatable-read transaction corrected from research | OG-001 | Adopt owner details and characterize bytes before removing the direct query. |
| 2026-08-18T21:37:42-04:00 | Codex / RS-05 | Revisions source-boundary ownership complete | Added `revisions/sourceboundary` resolver and owner rows; injected it through Reporting and server composition; moved selection characterization to Revisions; removed Incidents `change_sets` query and duplicate Reporting state/serialization; updated Reporting and backend boundary guards plus generated topology | Dual path row `3/3` (`20260819T012049Z-p2604602`); new Revisions rows `1/1` and `3/3`; full Revisions `27/27` and `20/20`, Reporting `5/5` and `4/4`, Incidents `26/26` and `19/19`, app.server `24/24` and `17/17`; boundary, Go lint, harness, drift, policy, and JSON gates pass | First formatter run rejected unsorted authored rows and harness run `20260819T013644Z-p2939159` reported the expected fixture-count delta; both authored inputs were corrected | Begin RS-03. |
| 2026-08-18T23:41:00-04:00 | Codex / RS-07 | Authored and generated surfaces are reproducible | Final authored changes include Core 01/02/04 owner text, Revisions schemas/vectors/index, renamed Workbook SQL, boundary policy, owner-family routing, and fixture policy; generated changes are the Revisions contract projection, renamed Workbook sqlc output, and execution-topology index | `make generate`, drift `4/4`, artifact policy `3/3`, JSON shape `3/3`, finalizer generated structure, and full check pass | None; generated roots were changed only by repository generators | Handoff complete. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-18T17:23:50-04:00 | Codex / tracker creation | Owner routing contains 41 rows: 29 Go, 7 Playwright, 5 Vitest | Inspected incident tests, shared test-support callers, verification owner, test catalog, test family; touched only this tracker | `make help`; `make task-guide ROLE=module-author OWNER=module.incidents`; `make explain-test-owner OWNER=module.incidents`; target explanations for owner slices, browser, lint, drift, and check; `jq`, `rg` | Canonical validation commands discovered; explanation targets passed; no test/lint/drift/product target was run | Characterization gap audit remains TODO | Run `make lint-markdown` for this tracker, then preserve owner rows for RS-00. |
| 2026-08-18T18:07:14-04:00 | Codex / NLSpec-grade revision | Requirement-to-evidence map and slice acceptance are explicit | Re-read test family, target tests, and cross-module test-support callers; touched only tracker | `jq`, `rg`, `find`; `make lint-markdown` | 41-row accounting preserved; Markdown lint passed; harness changes remain conditional | RS-00 executable characterization remains | Add tests only in a later authorized implementation. |
| 2026-08-18T19:19:30-04:00 | Codex / RS-00 | Characterization baseline complete | Added Reporting source-boundary goldens/repeatable-read test, strengthened Collaboration target isolation, updated Reporting test-family selector and generated topology; recorded 70 production admission consumer methods | Incidents `24/24` and `18/18`; Workbook `67/67` and `39/39`; Revisions `25/25` and `19/19`; Reporting `5/5` and `4/4`; Collaboration `31/31` and `23/23`; boundary, lint, harness, generation/drift/policy/JSON gates pass | Initial runs `20260818T225946Z-p55817` and `20260818T230050Z-p71015` failed from duplicate fixture UUID and PostgreSQL microsecond precision in the new test; both were test-authoring issues corrected before passing `20260818T230202Z-p85969` | Begin RS-01. |
| 2026-08-18T20:57:21-04:00 | Codex / RS-01 | Neutral admission and transaction safety complete | Added `incidents/admission`, migrated all 70 production consumer methods, removed `access.go` and root HTTP-shaped admission helpers, added mutation reauthorization and serialized membership administration, added concurrency and fail-closed tests, and updated boundary/harness owner inputs plus generated topology | Incidents focused `26/26` (`20260818T234543Z-p677237`) and service `19/19` (`20260818T234709Z-p721104`); every touched collaborator focused and service-backed owner passed; boundary `3/3` (`20260819T005219Z-p2252281`), `make lint-go`, harness `2/2` (`20260819T005403Z-p2272380`), drift `4/4`, policy `3/3`, and JSON `3/3` pass | Evidence/Entities exposed and corrected an over-restrictive read role set; two object-store readiness failures passed on clean rerun; exact-import allowlists, two dead helpers, and the fixture-policy count were corrected. No public or schema change occurred. | Begin RS-04. |
| 2026-08-18T21:59:49-04:00 | Codex / RS-03 | Fresh-commit effect matrix and incident/user isolation complete | Added coordinator unit coverage, committed-audit identity integration coverage, and incident/user hub isolation assertions; updated Collaboration and Incidents owner manifests plus transaction fixture approval/count | New coordinator row `1/1` (`20260819T014820Z-p2978325`); new audit-identity row `3/3` (`20260819T014938Z-p2994515`); full owner and assembly results are recorded above; public WebSocket messages, close replay, delete `404`, and account/other-incident access remain unchanged | No blocker; final browser coverage remains in RS-07's required broad sequence | Begin RS-02. |
| 2026-08-18T22:17:05-04:00 | Codex / RS-02 | HTTP test routing and parity evidence complete | Retained existing row IDs and verification IDs while moving 11 selector groups to `incidents/httpapi`; added an explicit route-composition row; reclassified the pure Location derivation row from PostgreSQL transaction to no-fixture unit evidence; regenerated topology | Focused and service-backed Incidents and server results are recorded above; `make harness-contract`, `make generate-drift`, generated policy, and JSON shape checks pass | Webserver-backed browser coverage is intentionally retained for the mandatory RS-07 final sequence | Begin RS-06. |
| 2026-08-18T22:34:20-04:00 | Codex / RS-06 | Retired-path and retained-support evidence complete | Searched every approved compatibility family across production code and SQL; enumerated retained Incidents test-support files; exercised all affected source and assembly owners after enabling regression boundaries | All focused/service-backed counts and run roots are recorded in the backend row above; searches returned no retired production match and `make harness-contract` passes | None | Begin RS-07 validation and handoff completion. |
| 2026-08-18T23:41:00-04:00 | Codex / RS-07 | Final owner, browser, structural, and broad evidence complete | Exercised Incidents, Workbook, Revisions, Reporting, Collaboration, every admission consumer, and server composition in focused/service-backed modes; then browser, boundary, lint, harness, drift, policy, JSON, Markdown, finalizer, and full-check gates | Section 8.1 records every result and run root; the only failure was Reporting object-store preflight `20260819T024458Z-p88772`, followed by clean focused `5/5` and service `4/4` passes | Retained-run-only finalizer checks skipped because `RESULTS_DIR` was unset; no product check skipped | Handoff complete. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-18T17:23:50-04:00 | Codex / tracker creation | Current route-time membership/role checks and error distinctions mapped | Inspected Core 04, incident access/admission/routes, membership audit handlers, relevant peer route callers and tests; touched only this tracker | `rg`, `sed` | Hidden `404`, visible insufficient-role `403`, deployment-admin non-bypass, live reauthorization, and closed-state admission frozen | RB-003 and RB-004 | Design neutral admission and post-commit effects without changing outcomes. |
| 2026-08-18T18:07:14-04:00 | Codex / NLSpec-grade revision | Admission design uses explicit nonordinal role sets and closed invalid defaults | Inspected every production `Access` method caller and current role arguments; touched only tracker | Targeted `rg -n -C`, `sed` | Neutral denial/HTTP mapping and consumer-family matrix authored | Per-method RS-00 matrix remains executable prerequisite | Characterize precedence, then migrate callers in batches. |
| 2026-08-18T20:57:21-04:00 | Codex / RS-01 | Admission is transport-neutral and fail-closed | Inspected and changed Incidents admission/application code plus every production consumer recorded in AM-001 through AM-070 | Focused and service-backed owner slices; retired-symbol searches; boundary, lint, harness, and drift gates | Exact role unions, denial precedence, malformed/unknown values, nil dependencies, deployment-admin non-bypass, caller-transaction rechecks, and cross-admin serialization pass | None | Preserve neutral admission while relocating Workbook persistence in RS-04. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-18T17:23:50-04:00 | Codex / tracker creation | Planning artifact is ready for implementation handoff | Inspected evidence listed in section 1; touched only this tracker | Repository discovery and Make explanation commands listed above | No production refactor performed; behavior-changing proposals are outside authorization | RB-001, RB-002, RB-003, RB-004 | Resolve blockers as each affected slice is authorized; start with RS-00. |
| 2026-08-18T18:07:14-04:00 | Codex / NLSpec-grade revision | RB-001 through RB-004 are design-resolved; two owner-authoring gates remain | Inspected all evidence named in section 1; touched only tracker | Discovery commands above; `make lint-markdown`; `git diff HEAD --check -- docs/handoffs/incidents-module-refactor-tracker.md` | Markdown and tracker-scoped diff checks passed; no production/test/contract/generated/harness change | OG-001, OG-002 | Adopt owner text, then begin RS-00; do not start RS-03/RS-05 before their gates. |
| 2026-08-18T20:57:21-04:00 | Codex / RS-01 | RS-01 exit criteria satisfied; RS-04 is the sole next workstream | Reviewed final symbol/import searches, owner matrices, generated inputs, and retained run evidence | All RS-01 commands above; tracker Markdown and diff gates run after this update | No legacy admission path remains; public HTTP/OpenAPI/WebSocket and database schema remain unchanged | None | Begin RS-04 only after the tracker gates pass. |
| 2026-08-18T23:41:00-04:00 | Codex / RS-07 | No open implementation risk or dependent work remains | Reviewed final inventory, all ledger rows, open gates, compatibility searches, generated drift, validation evidence, and staged tracker preservation | Every workstream and IT item is `DONE`; section 8.1 and section 12 are complete | None | Preserve run roots and this tracker as the handoff record. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Workbook preference persistence owner and transaction dependency | Prevents ownership inversion while retaining atomic create/import | Section 3.2 contract and RS-04 evidence | DONE in RS-04. |
| RB-002 | Revisions-owned Reporting source boundary | Removes physical coupling without changing immutable snapshot bytes | Section 3.3 contract, OG-001, and RS-05 evidence | DONE in RS-05. |
| RB-003 | Broad HTTP-shaped admission compatibility | Prevents a flag day and authorization drift | Section 3.4 types, mappings, consumer matrix, and RS-00 evidence | DONE: neutral admission/legacy removal completed in RS-01 and physical HTTP extraction completed in RS-02. |
| RB-004 | Ownership of close/removal terminal effects | Ensures non-HTTP callers cannot omit post-commit effects | Section 3.5 coordinator and OG-002 adoption | DONE in RS-03; raw production application calls are boundary-forbidden outside the coordinator. |
| OG-001 | Core 02 exact source-boundary JSON and tie-breaking ownership | Stable public hashes and immutable replay require one exact owner | Core 02 REQ-02-143b, Core 01 reference, AC-266, and `contracts/revisions/source-boundary*` | DONE in SP-01. |
| OG-002 | Core 01/Core 04 terminal-effect ordering, target isolation, and process-local delivery | Effects must follow a fresh commit and remain incident/user scoped | Core 01 REQ-01-592; Core 04 REQ-04-017, AC-162, and AC-419 | DONE in SP-02. |

No owner contradiction was found. SP-01 and SP-02 closed OG-001 and OG-002 without
changing public wire behavior or deployment topology.

## 12. Binary Completion Criteria

| Criterion | Result | Evidence |
| --- | --- | --- |
| Every file in `internal/modules/incidents` is inventoried or explicitly out of scope. | PASS | Section 2 preserves all 43 activation rows and inventories all 45 final files; `.gitkeep` is deleted. |
| Every discovered public contract risk has an owner and test posture. | PASS | Section 4 maps routes, collaboration, admission, preferences, portability, reporting, generated surfaces, frontend effects, and harness evidence. |
| Every proposed workflow has dependencies and exit criteria. | PASS | Section 6 records predecessors, successors, validation, and handoff checkpoints for WF-00 through WF-08. |
| Every proposed implementation slice preserves owner-required public compatibility and removes unjustified internal legacy surfaces. | PASS | Section 7 freezes public behavior while authorizing structural implementation changes. |
| Validation commands are discovered or marked `TODO` with a reason. | PASS | Section 8 contains Make-owned commands for every requested layer. |
| Contradictions are marked `BLOCKED: owner contradiction`. | PASS | No owner contradiction was found; section 1 records the required response if one appears. |
| Repository/framework mismatches are recorded as planning findings. | PASS | Section 3 records that the live target is a legitimate orchestration owner despite the framework's non-presumption rule. |
| Handoff sections are current enough for another agent to continue without rediscovery. | PASS | Section 10 records the snapshot, evidence, commands, results, blockers, and next actions by workstream. |
| All 11 Incidents OpenAPI operations are individually frozen. | PASS | Section 4.1 contains one exact method/path row per operation. |
| Each resolved boundary has an interface, dependency direction, explicit invalid/default behavior, failure semantics, evidence, and rollback. | PASS | Sections 3.2 through 3.5 and section 7. |
| Admission roles are nonordinal and mapped to consumers. | PASS | Section 3.4 defines explicit sets, fail-closed values, neutral outcomes, and consumer families. |
| Reporting semantics match the live transaction. | PASS | `INC-REF-SB-002` requires the same repeatable-read snapshot without incorrectly requiring a read-only transaction. |
| Membership deletion retains current non-idempotent behavior. | PASS | Section 3.5 and the DELETE operation row require missing/repeated deletion to remain `404`. |
| Collaboration delivery claims are process- and connection-bounded. | PASS | `INC-REF-CE-003` and `INC-REF-CE-006` exclude durable delivery. |
| Prior handoff history is preserved and current execution evidence is append-only. | PASS | Section 10 retains earlier rows and appends the WF-00 activation record. |
| Two independent implementers can derive the same package edges, interfaces, defaults, dependency graph, and acceptance outcomes. | PASS | Stable requirement IDs, mapping tables, authority gates, and binary slice criteria define the result. |
| Every authorized workstream and implementation item is complete. | PASS | The execution ledger and IT-001 through IT-020 contain only `DONE`; no dependent work or compatibility shim remains. |
| Final validation is reproducible and complete. | PASS | Section 8.1 records generation, all owner matrices, browser, structural, drift, finalization, and `make check` `626/626` evidence, including the infrastructure retry and deliberate retained-run skip. |

The Incidents module specification and implementation remediation is complete. No
owner contradiction, public contract change, database migration, frontend source
change, durable delivery expansion, or retained internal compatibility shim remains.
