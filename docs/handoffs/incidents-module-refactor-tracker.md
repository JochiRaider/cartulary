# incidents Module Refactoring Tracker and Handoff

> **Completed forward plan (2026-08-28):** Section 18 records the completed
> Iteration 3 Mutation and Source-State Production Hardening effort. Sections 1
> through 17 remain immutable Iteration 1 and Iteration 2 history. Every authorized
> Iteration 3 workstream is `DONE`; the validation and handoff evidence is retained
> in section 18.13.

## 1. Scope and Source Posture

- **Target path:** `internal/modules/incidents`
- **Target label:** `incidents` (derived from the target path and normalized to lowercase kebab case)
- **Output path:** `docs/handoffs/incidents-module-refactor-tracker.md`
- **Repository snapshot:** commit `6e5eb78380cdabc7b41f6c0b05a932a1539dece7`; the worktree was clean and `main` was four commits ahead of `origin/main` when this planning session began.
- **Live execution snapshot:** implementation was authorized on 2026-08-18 from commit `6e5eb78380cdabc7b41f6c0b05a932a1539dece7`; `main` remained four commits ahead of `origin/main`, and this tracker remained the only staged path at activation.
- **Status:** Iteration 1 implementation is complete and every Iteration 1 workstream is `DONE`. Iteration 2 is planned in sections 13 through 17; only its document-activation workstream is authorized by the current request.
- **Execution boundary:** Iteration 1's completed production state remains frozen. The current document-only request may change only this tracker. Future Iteration 2 production code, tests, adopted specifications, authored contracts, SQL inputs, boundary policy, harness owner inputs, and generated projections remain `TODO` until separately authorized. Public HTTP/OpenAPI/WebSocket behavior, persisted source-boundary v1 bytes, persisted idempotency hashes, and database schema remain frozen.
- **Non-goals:** no database migration, frontend source change, durable outbox, multi-process delivery design, public route or payload change, or permanent internal compatibility shim is planned. Discovery requiring one stops the affected slice for owner review.
- **Authorization:** the Iteration 1 remediation authorized WF-00, SP-01, SP-02, and implementation slices RS-00, RS-01, RS-04, RS-05, RS-03, RS-02, RS-06, and RS-07 in that order, and those slices are complete. The current request authorizes only `I2-WF-00`; `I2-SP-01` and `I2-RS-00` through `I2-RS-07` are planned but not execution-authorized.
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

## 13. Iteration 2 Charter and Execution Ledger

Sections 1 through 12 are the closed, append-only Iteration 1 planning and handoff
record. This section begins Iteration 2 without reopening or reinterpreting any
Iteration 1 result.

- **Iteration target:** remove confirmed dead and legacy Incidents surfaces, make
  construction and dependency direction explicit, and close production-readiness
  gaps without widening public behavior.
- **Planning baseline:** clean `main` at
  `ffe49c02d40d04c5e55a5bbc8f4d95276e358f0d` on 2026-08-19.
- **Current authorization:** the 2026-08-19 implementation request authorizes
  `I2-SP-01` and `I2-RS-00` through `I2-RS-07` in the ledger order. Only the one
  active workstream may change its authorized areas; every successor remains gated
  on its predecessor's completed tracker checkpoint.
- **Compatibility boundary:** public HTTP, OpenAPI, WebSocket, session, error,
  validation-order, persisted idempotency, and database-schema behavior remains
  unchanged. Internal Go callers migrate to the final interfaces without aliases or
  forwarding shims.
- **Expansion gate:** discovery requiring a public contract change, database
  migration, frontend change, different persisted request hash, or broader Recovery
  generation policy MUST mark the affected workstream `BLOCKED: owner contradiction`
  or `BLOCKED: scope expansion` before dependent work begins.
- **Source posture:** `docs/domain.md` supplies vocabulary and owner navigation only.
  `docs/research/nlspec-spec.md` supplies writing and completeness guidance only.
  Neither document is an instruction source for this iteration, and neither requires
  an Iteration 2 edit.

Iteration 2 status values are closed. `TODO` means execution has not begun;
`IN_PROGRESS` is the only active workstream; `DONE` means every exit criterion and
tracker checkpoint passed; `BLOCKED` records the exact unmet owner or external
prerequisite. A successor MUST NOT begin before every predecessor is `DONE`.

The tracker update is the final action of every future workstream. The row, detailed
evidence, failures, residual risks, and next dependency MUST be current before
`make lint-markdown` and tracker-scoped `git diff HEAD --check` run. A failed exit
criterion stops all dependent work.

| Workstream | Status | Dependency | Authorized change | Exit evidence | Next dependency |
| --- | --- | --- | --- | --- | --- |
| `I2-WF-00` — append Iteration 2 plan | DONE | None | This tracker only | Sections 13–17 added; Markdown lint passed at `20260819T042615Z-p1430172`; tracker-scoped diff and tracker-only changed-path review passed | `I2-SP-01`, after separate execution authorization |
| `I2-SP-01` — close Recovery owner drift | DONE | `I2-WF-00` | Core 01 and Core 04 owner text only | Core 01 states exact 113/84 arithmetic and Graph classification; Core 04 names additive migrations 30–34; Markdown, traceability, and diff gates passed | `I2-RS-00` |
| `I2-RS-00` — characterization baseline | DONE | `I2-SP-01` | Tests, contracts needed to freeze current facts, and tracker | Six exact hash preimages/digests, existing constructor/finalizer/rollback and 11-operation behavior, zero-caller searches, and both existing Recovery generation identities pass | `I2-RS-01` |
| `I2-RS-01` — remove confirmed dead leaves | DONE | `I2-RS-00` | Incidents, affected classifiers, authored SQL, generated SQL, tests, boundary policy, and tracker | All five dead-leaf groups and generated/type-switch residue are absent; recurrence, affected-owner, harness, static, generation, and drift gates passed | `I2-RS-02` |
| `I2-RS-02` — internalize idempotency and helpers | DONE | `I2-RS-01` | Incidents callers, tests, and tracker | Final application signatures compile; persisted hash goldens and replay/conflict matrices pass; stale file and exports are gone | `I2-RS-03` |
| `I2-RS-03` — consolidate construction and test faults | DONE | `I2-RS-02` | Incidents construction, server/test composition, test support, boundary policy, and tracker | One fail-fast constructor remains and commit-fault injection is test-owned | `I2-RS-04` |
| `I2-RS-04` — harden HTTP composition | DONE | `I2-RS-03` | Incidents HTTP, server assembly, tests, boundary/harness inputs, and tracker | Exact dependencies are injected and all 11 operations retain parity | `I2-RS-05` |
| `I2-RS-05` — relocate bundle consumer port | DONE | `I2-RS-04` | Incidents, Incident Bundles, assembly, tests, boundary/harness inputs, and tracker | Incident Bundles owns its port/request ID; invalid input and prepared-value failures are deterministic; root ports file is gone | `I2-RS-06` |
| `I2-RS-06` — correct Recovery ownership | DONE | `I2-RS-05` | Workbook/Incidents contributions, Recovery contracts/runtime, assembly, tests, generated projections, policy, and tracker | Workbook owns preference tables and all three exact generations validate without reinterpretation | `I2-RS-07` |
| `I2-RS-07` — cleanup, validation, and handoff | DONE | `I2-RS-06` | Final residue, tests, generated outputs through owners, and tracker | Retired surfaces and drift are absent; retained exports/support are justified; focused, service-backed, browser, static, generation, finalization, and repository-wide validation pass | Complete |

## 14. Live Cleanup Inventory

This inventory records repository facts at the Iteration 2 planning baseline. A zero
production caller is evidence for removal, not independent authority to change public
behavior.

| ID | Current surface and location | Live callers or dependency | Disposition | Risk and owning workstream |
| --- | --- | --- | --- | --- |
| `I2-LC-001` | `IncidentCreateIdempotencyScope` in `incidents/lifecycle_access.go` | No caller | Delete | Dead compatibility vocabulary can imply a configurable scope that does not exist; `I2-RS-01`. |
| `I2-LC-002` | `(*Application).GetIncidentMembershipForUser` in `incident_application.go` | No caller | Delete | An unused read expands the application facade and future support burden; `I2-RS-01`. |
| `I2-LC-003` | `storetest.LookupUserByEmail` | No caller | Delete | Dead test support obscures the supported fixture surface; `I2-RS-01`. |
| `I2-LC-004` | Authored queries `EnsureIncidentOpenForMutation` and `ListIncidentMemberships` in `db/queries/incidents.sql` | No generated method call; the latter leaves a dead generated row-type switch in `repository.go` | Delete authored queries and dependent residue, then regenerate | Unused persistence APIs can be mistaken for supported transaction paths; `I2-RS-01`. |
| `I2-LC-005` | `incidents.ErrIncidentClosed` and classifier branches in Incidents HTTP, Entities, and Workbook | No Incidents production producer; actual incident closure denial is `admission.DenialIncidentClosed` | Delete sentinel and compatibility branches; retain `timeline.ErrIncidentClosed` | Dual error families can drift or mask the real admission contract; `I2-RS-01`. |
| `I2-LC-006` | Exported create/lifecycle/membership hash helpers and caller-supplied `requestHash []byte` application parameters | HTTP routes, Incidents store/test support, lifecycle contract tests, and commit-fault support construct or pass hashes | Derive privately inside application methods and remove arguments/exports | A caller can persist a hash that does not describe the normalized request; `I2-RS-02`. |
| `I2-LC-007` | `IncidentCreateBootstrap`, `DefaultIncidentCreateBootstrap`, `ApplyIncidentPatch`, and `WouldLeaveNoIncidentAdmins` | Incidents implementation plus one same-owner HTTP test | Privatize or inline; move direct helper evidence to the owning package | Unnecessary exports turn implementation choices into de facto APIs; `I2-RS-02` and `I2-RS-07`. |
| `I2-LC-008` | `NewApplicationWithOptions`, `ApplicationOptions`, `IncidentCreateCommitPort`, and `directIncidentCreateCommit` | Production uses the ordinary constructor; only Incidents tests and `internal/testutil/appsupport/commit_fault.go` use the commit seam | Replace with one dependency constructor and a test-owned database/transaction fault decorator | A production-only test hook weakens construction invariants and expands the root port surface; `I2-RS-03`. |
| `I2-LC-009` | `httpapi.RegisterRoutes(options ...RouteOptions)`, exported `Service`, and six exported `Decode*Request` functions | Production supplies exactly one option; decoder/service use is package-local | Use one exact dependency value and private service/decoders | Variadic zero-option compatibility always fails and hides required composition; `I2-RS-04`. |
| `I2-LC-010` | HTTP creates `admission.NewChecker` from platform dependencies; terminal coordinator interface lives in Incidents root | Server already owns composition; only HTTP consumes the coordinator interface | Inject a narrow admission capability and move the terminal interface to its consumer | Adapter construction in HTTP and producer-owned ports increase coupling; `I2-RS-04`. |
| `I2-LC-011` | Import-finalizer params/interface, unavailable-admin error, and bundle request-ID helper live under Incidents | Incident Bundles module, worker, server assembly, and Incidents store tests | Move the consumer contract/error to `incidentbundles/importfinalizerport`; keep implementation in Incidents; move request-ID derivation to Incident Bundles | Current placement forces Incident Bundles to depend on unrelated Incidents root surface; `I2-RS-05`. |
| `I2-LC-012` | Import finalizer substitutes `time.Now()` for zero `PublishedAt`; source-port apply uses `value.(sourceport.PreparedFiles)` | Current production callers supply a timestamp and adapter-prepared files | Reject zero IDs/time and return a controlled internal error for a wrong prepared value | Hidden clocks break deterministic commits and unchecked assertions can panic a worker; `I2-RS-05`. |
| `I2-LC-013` | Incidents recovery contribution claims `incident_workbook_preferences` and `user_workbook_preferences`; Workbook has no contribution | Recovery assembly registers Incidents but not Workbook | Move both declarations to Workbook and register its contribution | Persistence ownership and recovery ownership disagree; `I2-RS-06`. |
| `I2-LC-014` | Recovery admission recognizes current Graph v3 or one historical Graph v2 pair through branching and validates other artifacts against current state | Retained backups may bind either admitted generation | Replace the heuristic with a finite generation registry and add the pre-ownership-change Graph v3 generation | A catalog-owner-only digest change would reject retained backups or reinterpret them through current state; `I2-RS-06`. |
| `I2-LC-015` | Incidents source-port/resource builders and Incidents-specific mutation, performance, scenario, route, and store support | Active owner assembly and broad test families | Retain; remove only proven dead members | Moving cohesive owner behavior would reduce clarity without lowering coupling; regression review in `I2-RS-07`. |

The application constructor currently has thirteen caller files outside its definition:
server assembly; two Entities merge tests; Incidents membership-concurrency and store
tests; Incidents performance/store harnesses; Indicators and Revisions transaction
tests; Timeline store support; and shared app-scenario support. `I2-RS-03` MUST migrate
the complete compiler-discovered set rather than maintaining an overload.

## 15. Iteration 2 Remediation Decisions

| Gap | Remediation and affected areas | Rationale and long-term benefit | Compatibility or migration impact | Risk if unresolved | Completion validation |
| --- | --- | --- | --- | --- | --- |
| `I2-GAP-001` — Recovery owner-spec drift | Before implementation, amend Core 01's catalog posture and REQ-01-647 from 113/83 with five Graph rebuildables to 113/84 with four Graph rebuildables; identify `network_flow_graph_views` as authoritative. Amend Core 04 AC-542 to immutable versions 1–29 plus additive versions 30–34 and Recovery 113/84. **Areas:** specification and tracker. | Restores one authoritative description of the already-shipped migration and Recovery projection without treating implementation as authority. Later ownership work starts from truthful invariants. | No runtime, SQL, migration, wire, or data change. Existing historical handoffs remain historical. | Owner text, validation code, and live contracts disagree, so later generation changes cannot be reviewed reliably. | Core text states `84 + 4 + 6 + 1 + 7 + 1 + 10 = 113`; `network_flow_graph_views` and Graph result classifications match live projections; AC-542 names head 34; traceability and Markdown checks pass. |
| `I2-GAP-002` — confirmed dead leaves | Delete `I2-LC-001` through `I2-LC-005`, change authored SQL before regeneration, and add boundary/static assertions where recurrence is plausible. **Areas:** implementation, authored SQL, generated SQL, tests, and boundary policy. | Removes misleading APIs and duplicate error vocabulary while shrinking persistence and test surfaces. | Internal compile-time removal only. Actual admission denial and Timeline closure behavior remain. No schema migration. | Future callers can adopt an unused or semantically obsolete path, making removal more costly and error handling inconsistent. | Repository searches find no retired symbol/query/generated method/type-switch branch; Incidents, Entities, and Workbook focused/service-backed slices plus boundary, lint, generation, and drift gates pass. |
| `I2-GAP-003` — caller-owned idempotency hashes | Derive hashes inside the three application mutations, remove public hash helpers and hash arguments, use closed private canonical payload structs, relocate patch/bootstrap helpers, and delete `lifecycle_access.go`. **Areas:** implementation and tests. | The application becomes the sole authority for the request-to-hash relationship; future callers cannot accidentally persist a mismatched digest. | Existing SHA-256 values, stored response JSON/status, replay, conflict, and actor scoping remain byte-compatible. No data migration. Internal callers change signatures. | A new caller can provide arbitrary or stale hash bytes and corrupt idempotency equivalence. | Pre/post exact-byte and digest goldens cover nullable/omitted fields and lifecycle action selection; fresh, replay, different-payload conflict, and different-actor matrices pass; no external hash helper remains. |
| `I2-GAP-004` — construction and production test seam | Introduce `ApplicationDependencies` and one error-returning constructor; reject nil and typed-nil dependencies before readiness. Replace the commit port with a test-owned `postgres.DB` decorator whose returned transaction rolls back and returns the retained commit-fault sentinel from `Commit`. **Areas:** implementation, test support, tests, and boundary policy. | Makes valid construction explicit and keeps fault injection at the infrastructure boundary it exercises. | Thirteen caller files migrate at compile time. Ordinary transaction behavior and commit-failure outcome remain unchanged. No compatibility constructor remains. | Partially initialized applications can reach serving, and production must permanently carry a test-only abstraction. | Nil/typed-nil matrices fail deterministically; server composition succeeds only with complete dependencies; commit-failure tests prove no incident, membership, preference, audit, or idempotency residue; retired constructor/seam searches are empty. |
| `I2-GAP-005` — HTTP composition and excess exports | Define exact `httpapi.Dependencies`, consumer-owned `Application`, `AdmissionChecker`, and `TerminalMutationCoordinator` capabilities; inject all three from server assembly; make service and decoders private; remove variadic registration. **Areas:** implementation, tests, assembly, harness, and boundary policy. | HTTP owns only transport concerns and declares the minimum application capabilities it consumes. Required dependencies become visible at composition. | All 11 public operations, OpenAPI operation IDs, error/envelope shapes, validation order, session sliding, and persisted replay payloads remain stable. Internal composition breaks intentionally. | Hidden adapter construction and broad exported helpers promote coupling and allow incomplete registration shapes. | Missing/nil/typed-nil dependency tests fail before binding; runtime/OpenAPI parity and all operation families pass; root ports no longer contain the terminal interface; package export audit passes. |
| `I2-GAP-006` — Incident Bundle port ownership and deterministic finalization | Add `internal/modules/incidentbundles/importfinalizerport`; move the params, interface, and unavailable-admin sentinel there; return `(importfinalizerport.Finalizer, error)` from construction; reject zero IDs/time; move request-ID derivation into Incident Bundles; replace the unchecked assertion; delete empty `incidents/ports.go`. **Areas:** implementation, tests, assembly, and boundary policy. | The consumer owns its narrow capability while Incidents retains atomic implementation. Explicit time and type validation make worker behavior deterministic and panic-free. | Bundle wire, job, transaction, membership, preferences, and audit behavior remain unchanged. Current production callers already provide nonzero values. Internal imports change with no shim. | Cross-owner root coupling persists; malformed internal inputs can call a hidden clock or crash a worker. | Create/import success, rollback, unavailable-admin, nil/typed-nil writer, zero-field, wrong-prepared-type, request-ID, replay, and worker tests pass; no old port symbol or root ports file remains. |
| `I2-GAP-007` — Workbook recovery ownership and generation compatibility | Add Workbook's contribution for the two preference tables, remove them from Incidents, register Workbook, and replace Recovery's two-branch admission with a schema-validated three-entry generation registry. **Areas:** specification projection, implementation, contracts, tests, generated outputs, assembly, and boundary policy. | Physical, semantic, and recovery ownership agree. Exact finite generations preserve justified backup compatibility without creating an open-ended legacy reader. | Table names and 113/84 membership are unchanged; contribution count becomes 30 and current catalog digest changes. New capture emits only the new generation. Exact pre-change Graph v3 and existing Graph v2 backups remain readable while retained. No data/schema migration. | Retained backups can become unrestorable or be silently validated against the wrong owner catalog; future Workbook changes remain coupled to Incidents. | Current generation and both historical pairs pass exact catalog, PostgreSQL unit, object-family, codec, and Graph-artifact tests; cross-pair mixtures and unknown generations fail before mutation; boundary checks forbid preference-table claims in Incidents. |
| `I2-GAP-008` — final residue and regression prevention | Privatize remaining same-package helpers, retain cohesive active source/test support, add recurrence policies, recalculate the final inventory, and close the tracker. **Areas:** implementation, tests, boundary/harness policy, generated artifacts, and documentation. | Leaves one comprehensible production architecture and an intentionally small exported surface without sacrificing valuable owner fixtures. | Internal compile-time cleanup only. No public product or persisted-state change. | Transitional names and accidental exports become long-lived compatibility obligations. | Retired-symbol/import scans, export inventory, all affected owner matrices, browser/static/harness/drift checks, finalization, and full check pass with every row `DONE`. |

Historical Recovery support is the only planned legacy retention. It has continuing
value because REQ-01-575 requires an exact decoder for a selected retained backup and
successful backup metadata remains retained for at least 30 days. An old generation
MUST NOT be removed on a calendar assumption: removal requires repository or
deployment evidence that no supported retained backup names it and a separate adopted
owner decision.

## 16. Target Interfaces, Invariants, and Workstream Detail

### 16.1 Application construction and idempotency

The final construction surface is exact:

```go
type ApplicationDependencies struct {
    Postgres            postgres.DB
    PreferenceBootstrap bootstrapport.Writer
}

func NewApplication(ApplicationDependencies) (*Application, error)
```

Both dependencies are required. A nil or typed-nil value has no default and fails
before an application can be registered or served. The application may continue to
construct its private repository, authentication store, and admission checker from the
validated PostgreSQL dependency.

The final mutation shapes omit caller-provided request hashes:

```go
CreateIncident(ctx, actor, request, requestID, now)
TransitionIncidentLifecycle(ctx, actor, incidentID, action, request, requestID, now)
CreateMembership(ctx, actor, incidentID, targetUser, request, requestID, now)
```

Each method derives SHA-256 from compact Go `encoding/json` bytes with no schema
prefix, whitespace suffix, or trailing newline. Private typed payloads declare members
in the exact persisted order:

- Create: `client_txn_id`, `current_phase`, `description`, `incident_key`,
  `primary_external_case_ref`, `severity`, `title`, `tlp`.
- Lifecycle: `action_route`, `base_incident_version`, `reason`.
- Membership: `client_txn_id`, `email`, `role`, `user_id`.

Omitted normalized nullable values retain their current JSON `null` representation.
The lifecycle action remains part of the hash; `close` and `reopen` with otherwise
equal request data are distinct. Stored hash, response bytes, response status, scope,
and replay precedence do not change.

### 16.2 HTTP consumer capabilities

HTTP registration becomes exactly `RegisterRoutes(Dependencies)`; there is no
variadic or zero-option form. `Dependencies` contains:

- an `Application` capability containing only `ListVisibleIncidents`,
  `CreateIncident`, `GetVisibleIncident`, `UpdateIncident`, `ListMemberships`,
  `CreateMembership`, `UpdateMembership`, and
  `ListAdministrativeAuditEvents`;
- an `AdmissionChecker` capability containing only `Check`; and
- a `TerminalMutationCoordinator` capability containing only incident lifecycle
  coordination and membership deletion coordination.

The coordinator lifecycle method adopts the application signature without a hash
argument. Server assembly constructs the neutral admission checker and injects all
three capabilities. HTTP may continue to construct transport-owned authentication and
pagination helpers from its platform dependency set. Missing, nil, or typed-nil owner
capabilities fail route composition before serving.

The HTTP service and all six incident/membership request decoders become private.
Their white-box tests remain in the `httpapi` package. Resource builders stay exported
from Incidents because both application persistence/audit logic and HTTP use them.

### 16.3 Incident Bundle finalizer port

`internal/modules/incidentbundles/importfinalizerport` owns the finalization params,
`Finalizer` interface, and `ErrInitialAdminUnavailable`. The interface remains:

```go
FinalizeIncidentBundleImportTx(context.Context, pgx.Tx, Params) error
```

Incidents supplies the implementation and returns
`(importfinalizerport.Finalizer, error)` from its constructor. The finalizer uses only
the caller-owned transaction. It never begins, commits, retries, upserts, calls a
clock, or changes authorization/audit scope. Nil/typed-nil preference writers and zero
incident ID, submitter ID, or publication timestamp fail before a query or write.

Incident Bundles privately derives `incident_bundle_import:<job UUID>`. Incidents
tests use explicit request IDs. A prepared value of the wrong internal type returns a
bounded internal error instead of panicking; it does not create a new public bundle
failure family.

### 16.4 Recovery generations and ownership

Workbook's contribution contains exactly `incident_workbook_preferences` and
`user_workbook_preferences`. Incidents retains exactly `incidents` and
`incident_memberships`. Recovery assembly registers both, increasing the contribution
count from 29 to 30 without changing the 113-table/84-required table set.

An authored, schema-validated Recovery generation registry is the sole admission
owner. Generated Go constants are derived from it; generated roots are never edited
directly. A closed generation record binds one exact catalog digest, codec-registry
digest, immutable catalog artifact, required table set, object-family set, and Graph
registry/binding artifacts. The registry contains exactly:

1. the new Workbook-owned current catalog, current codec registry, and Graph v3
   artifacts;
2. the frozen pre-change Incidents-owned workbook-preference catalog, current codec
   registry, and Graph v3 artifacts; and
3. the existing historical Graph v2 catalog, historical codec registry, and Graph v2
   artifacts.

The lookup key is the exact `(catalog digest, codec-registry digest)` pair. Duplicate
keys, multiple current entries, malformed artifacts, unknown pairs, or a catalog from
one entry combined with codecs or Graph artifacts from another fail before restore
mutation. Capture selects only entry 1. Restore uses the selected entry for PostgreSQL
unit count/order, object-family admission, catalog checks, and Graph artifacts; it does
not compare historical artifacts to current state or infer the Graph generation from
codec identity alone. Returned artifact bytes are immutable or defensively copied.

### 16.5 Slice execution and rollback

- `I2-SP-01` changes owner prose only and stops if live 113/84 or head 34 evidence is
  not reproducible.
- `I2-RS-00` adds evidence without changing production behavior. Any owner/behavior
  conflict blocks its dependent slice.
- `I2-RS-01` removes only proven dead leaves and regenerates immediately after the
  authored SQL edit.
- `I2-RS-02` migrates every caller atomically; it does not keep wrapper hash functions
  or overloads.
- `I2-RS-03` migrates the complete constructor caller set and proves the replacement
  fault wrapper before deleting the production seam.
- `I2-RS-04` composes final dependencies before deleting `RouteOptions`, exports, and
  the root terminal interface.
- `I2-RS-05` moves the port and all consumers in one slice; no type alias bridges old
  and new packages.
- `I2-RS-06` freezes the pre-change generation before changing contribution ownership,
  then proves old/current admission before deleting heuristic branches.
- `I2-RS-07` removes only residue made obsolete by prior passing slices and performs
  no unplanned feature work.

Rollback is workstream-local through Git. A failed slice restores its authored and
generated changes together; it MUST NOT preserve a partial compatibility wrapper,
rewrite persisted hashes, reinterpret a historical catalog, edit an immutable
migration, or advance the next tracker row.

## 17. Validation, Handoff, and Binary Completion

### 17.1 Per-workstream validation

Before selecting rows, use `make task-guide ROLE=module-author OWNER=<owner-id>` and
the repository explanation targets. Each implementation slice runs the focused and
service-backed owner slices for every changed owner. The minimum owner matrix is:

| Change family | Required owners |
| --- | --- |
| Dead SQL/errors and idempotency | `module.incidents`; add `module.entities` and `module.workbook` when their classifiers change |
| Constructor and HTTP composition | `module.incidents`, `app.server`, and every compiler-discovered caller owner, including Entities, Indicators, Revisions, and Timeline |
| Bundle finalizer port | `module.incidents`, `module.incidentbundles`, and `app.server` |
| Recovery ownership/generations | `module.incidents`, `module.workbook`, `module.recovery`, and `app.server`/Recovery assembly rows |

After an authored SQL, contract, or harness input changes, run `make generate` once
the input is coherent, then immediately run `make generate-drift`,
`make generated-artifact-policy-check`, and `make json-shape-check`. Every production
slice also runs `make backend-module-boundary-check` and `make lint-go`. Harness inputs
change only when selectors, ownership, or topology actually changes, followed by
`make harness-contract`.

Required scenario families include:

- exact create/lifecycle/membership hash preimages and digests for zero, null,
  non-null, replay, conflicting payload, different actor, and close/reopen action;
- nil and typed-nil construction, successful composition, and final-commit rollback
  without partial incident, membership, preference, audit, or idempotency rows;
- all 11 HTTP operations, validation precedence, concealment, paging, session sliding,
  Location, error/envelope shape, stored replay payload, and OpenAPI/runtime parity;
- bundle finalization success, rollback, missing/inactive/non-admin submitter, zero
  identity/time, nil writer, wrong prepared type, request-ID determinism, and worker
  replay;
- current Recovery capture/restore, exact pre-change Graph v3 restore, exact historical
  Graph v2 restore, cross-pair rejection, unknown generation rejection, contribution
  count 30, table set 113/84, and owner-boundary recurrence.

### 17.2 Final validation order

After all narrow workstream evidence passes:

1. Run focused and service-backed slices for Incidents, Incident Bundles, Workbook,
   Recovery, app.server, and every additional changed caller owner.
2. Run `make agent-finalize` with `RESULTS_DIR` unset unless a deliberately retained
   successful full-run root is supplied; record the retained-run skip when unset.
3. Run `make browser-e2e-webserver-backed`.
4. Run `make backend-module-boundary-check`.
5. Run `make lint-go`.
6. Run `make harness-contract`.
7. Run `make generate-drift`.
8. Run `make generated-artifact-policy-check`.
9. Run `make json-shape-check`.
10. Run `make lint-markdown`.
11. Run `make check`.
12. Run final repository-wide retired-symbol/import searches and `git diff --check`.

For this document-only `I2-WF-00` activation, the exact validation is limited to
`make lint-markdown`,
`git diff HEAD --check -- docs/handoffs/incidents-module-refactor-tracker.md`, and a
changed-path review proving that only this tracker changed and sections 1 through 12
remain Iteration 1 history.

### 17.3 Handoff record

Every workstream handoff records the baseline commit/worktree, files changed,
substantive decision, generated outputs, commands and result roots, every failed run
and disposition, skipped checks with reason, residual risk, rollback boundary, and
sole next dependency. `I2-RS-07` additionally records the final file/export inventory,
all three Recovery generation identities, the absence of a schema/data/public
contract migration, and confirmation that no internal compatibility shim remains.

### 17.4 Binary completion criteria

| Criterion | Planned result |
| --- | --- |
| Sections 1–12 remain a truthful Iteration 1 record and sections 13–17 control Iteration 2. | `I2-WF-00` exit |
| Core 01, Core 04, migrations, and Recovery projections agree on head 34 and 113/84. | `I2-SP-01` exit |
| Every removed symbol/query has zero live caller and a recurrence check where material. | `I2-RS-01` exit |
| The application alone derives byte-compatible idempotency hashes. | `I2-RS-02` exit |
| One fail-fast application constructor exists and production contains no commit-fault port. | `I2-RS-03` exit |
| HTTP has exact injected owner capabilities, private mechanics, and unchanged 11-operation behavior. | `I2-RS-04` exit |
| Incident Bundles owns its finalizer port/request ID and malformed internal values cannot panic or call a hidden clock. | `I2-RS-05` exit |
| Workbook owns both preference tables and exactly three immutable Recovery generations are admitted without reinterpretation. | `I2-RS-06` exit |
| No public route/wire, persisted hash, schema, migration, frontend, or domain-vocabulary change occurred. | Final handoff |
| No retired internal alias, overload, shim, symbol, import, query, generated residue, or unexplained drift remains. | `I2-RS-07` exit |
| Every required focused, service-backed, browser, static, harness, drift, finalization, and broad check passes or is explicitly blocked before dependent work. | `I2-RS-07` exit |

The 2026-08-19 implementation request authorizes the remaining Iteration 2 ledger.
Execution remains serial: one row may be `IN_PROGRESS`, and no successor begins until
its predecessor's tracker checkpoint and exit gates pass.

### 17.5 `I2-WF-00` completion evidence

| Time | Baseline and scope | Change | Validation | Result and next action |
| --- | --- | --- | --- | --- |
| 2026-08-19T00:23:21-04:00 | Clean `main` at `ffe49c02d40d04c5e55a5bbc8f4d95276e358f0d`; document-only authorization | Updated the section 1 posture and appended sections 13–17. No Iteration 1 ledger/history, production, test, owner, SQL, contract, generated, harness, migration, frontend, domain, or research file changed. | `make lint-markdown` passed at `.cartulary/test-results/20260819T042615Z-p1430172`; tracker-scoped `git diff HEAD --check`; `git status --short`; tracker-only diff review | `I2-WF-00` is `DONE`. Only this tracker is modified, sections 1–12 remain the Iteration 1 record, and `I2-SP-01` is the sole planned successor but remains unauthorized and `TODO`. |

### 17.6 `I2-SP-01` completion evidence

| Time | Baseline and scope | Change | Validation | Result and next action |
| --- | --- | --- | --- | --- |
| 2026-08-19T00:39:49-04:00 | `main` at `ffe49c02d40d04c5e55a5bbc8f4d95276e358f0d`; preserved the staged `I2-WF-00` tracker addition and changed only the two adopted owners plus the tracker | Core 01 now classifies 84 authoritative tables, four Graph result tables as rebuildable, and `network_flow_graph_views` as authoritative; its arithmetic is `84 + 4 + 6 + 1 + 7 + 1 + 10 = 113`. Core 04 AC-542 now names immutable migrations 1–29, additive migrations 30–34, and Recovery 113/84. No runtime, SQL, contract, generated, harness, migration, frontend, domain, or historical Iteration 1 content changed. | Migration-head and Recovery-limit shell assertions; owner-text and projection searches; `git diff HEAD --check`; `make lint-markdown` passed at `.cartulary/test-results/20260819T043911Z-p1437347` before the final tracker checkpoint, followed by the required final Markdown and tracker-diff gates | `I2-SP-01` is `DONE`. No failure, skipped technical check, owner contradiction, compatibility impact, or residual risk remains. Rollback is limited to the two owner-text corrections and this checkpoint. `I2-RS-00` is the sole successor. |

### 17.7 `I2-RS-00` completion evidence

| Time | Baseline and scope | Change | Validation | Result and next action |
| --- | --- | --- | --- | --- |
| 2026-08-19T00:49:43-04:00 | `I2-SP-01` complete on `main` at `ffe49c02d40d04c5e55a5bbc8f4d95276e358f0d`; retained the staged tracker baseline | Added test-only Incidents golden evidence for six exact compact JSON preimages and SHA-256 values: create with null and populated optionals, close, reopen, membership by user ID, and membership by email. Pinned the current pre-Workbook-ownership catalog digest `04174fbf70585af5afb5ace702b797f7512f655005b8502b02f4b41700a966a6`, Graph v3 registry/binding digests `61c3f7348c4df2bee3e969c905c91c9857082cf2839a0b57104e40339e3e16d3`/`6ec244d0b82466a18adbdb82554be29f5e4baac384175538acbc92e56f14b8d5`, and historical Graph v2 catalog/registry/binding digests `ce0a1f4053a9ce156273e4adf40c8b4185fa616170eadd6a860500d0b24fd22f`/`e75697ef1f6b5a197d299746fd42d2bf07afcd2e1c9d187a6fe695bca3096730`/`235c69bbc0e5d4f25f3fab7b1f2b8c30ba6370bfc65abcba75822007802621b9`. Existing focused matrices retain valid construction, finalizer publication time, final-commit rollback residue, 11-operation HTTP, and exact historical v2 dispatch evidence. Production behavior and contracts did not change. Static searches confirmed `I2-LC-001` through `I2-LC-004` have definitions/generated residue but no consumers, and `incidents.ErrIncidentClosed` has classifiers but no Incidents producer. | `make format` passed at `.cartulary/test-results/20260819T044312Z-p1440709`; `make test-slice OWNER=module.incidents` passed 26/26 at `.cartulary/test-results/20260819T044331Z-p1445424`; corrected the first Recovery golden attempt after `.cartulary/test-results/20260819T044332Z-p1445434` proved raw file hashes differ from canonical JSON digests; `make test-slice OWNER=module.recovery` then passed 24/24 at `.cartulary/test-results/20260819T044551Z-p1539559`; service-backed Incidents and Recovery slices each passed 19/19 at `.cartulary/test-results/20260819T044712Z-p1589755` and `.cartulary/test-results/20260819T044712Z-p1589757`; `make backend-module-boundary-check` passed at `.cartulary/test-results/20260819T044848Z-p1681858`; `make lint-go` passed; dead-surface and OpenAPI inventory searches; `make lint-markdown` passed at `.cartulary/test-results/20260819T044905Z-p1688425` before the final checkpoint, followed by the required final Markdown and tracker-diff gates | `I2-RS-00` is `DONE`. The failed Recovery run was a test-authoring error and is retained above; canonical constants corrected it with no product change. Contract generation/drift was skipped because no authored contract changed. Residual risk is limited to the intentionally retained dead surfaces, which the next slice removes. Rollback is the three test edits and this checkpoint. `I2-RS-01` is the sole successor. |

### 17.8 `I2-RS-01` completion evidence

| Time | Baseline and scope | Change | Validation | Result and next action |
| --- | --- | --- | --- | --- |
| 2026-08-19T00:59:20-04:00 | `I2-RS-00` complete on `main` at `ffe49c02d40d04c5e55a5bbc8f4d95276e358f0d`; retained all preceding staged and unstaged Iteration 2 work | Removed `IncidentCreateIdempotencyScope`, `GetIncidentMembershipForUser`, `storetest.LookupUserByEmail`, authored `EnsureIncidentOpenForMutation` and `ListIncidentMemberships`, their generated SQL methods/params/rows, the repository row switch, `incidents.ErrIncidentClosed`, and its Incidents HTTP, Entities, and Workbook classifiers. Removed the newly unused Entities/Workbook imports. Retained Timeline's independently produced and consumed sentinel. Added exact boundary-policy recurrence rules. Generated output changed only `internal/gen/sql/incidents.sql.go`; no schema, migration, public API, persisted behavior, frontend, domain, or compatibility shim changed. | `make generate` passed at `.cartulary/test-results/20260819T045100Z-p1690751`; `make format` passed at `.cartulary/test-results/20260819T045313Z-p1713574`; repository-wide production/generated retired-symbol searches passed; Incidents, Entities, and Workbook focused slices passed 26/26, 32/32, and 67/67 at `.cartulary/test-results/20260819T045336Z-p1728829`, `.cartulary/test-results/20260819T045336Z-p1728812`, and `.cartulary/test-results/20260819T045336Z-p1728825`; their service-backed slices passed 19/19, 29/29, and 39/39 at `.cartulary/test-results/20260819T045620Z-p1880738`, `.cartulary/test-results/20260819T045620Z-p1880740`, and `.cartulary/test-results/20260819T045620Z-p1880743`; `make generate-drift`, artifact policy, and JSON shape passed at `.cartulary/test-results/20260819T045223Z-p1697997`, `.cartulary/test-results/20260819T045233Z-p1708197`, and `.cartulary/test-results/20260819T045234Z-p1708606`; final boundary and Harness gates passed at `.cartulary/test-results/20260819T045903Z-p2030140` and `.cartulary/test-results/20260819T045903Z-p2030171`; `make lint-go` passed after removing unused imports; `make lint-markdown` passed at `.cartulary/test-results/20260819T045903Z-p2030165` before the final checkpoint, followed by the required final Markdown and tracker-diff gates | `I2-RS-01` is `DONE`. An intermediate `make lint-go` failed because the removed classifiers left unused imports; the imports were deleted and the rerun passed. No technical check was skipped and no residual dead-leaf risk remains. Rollback is the authored SQL, regenerated SQL, source removals, recurrence policy, and this checkpoint as one slice. `I2-RS-02` is the sole successor. |

### 17.9 `I2-RS-02` completion evidence

| Time | Baseline and scope | Change | Validation | Result and next action |
| --- | --- | --- | --- | --- |
| 2026-08-19T01:20:27-04:00 | `I2-RS-01` complete on `main` at `ffe49c02d40d04c5e55a5bbc8f4d95276e358f0d`; retained all preceding Iteration 2 work and changed Incidents mutation callers, their direct tests/support, Collaboration coordination, affected owner test callers, boundary policy, Harness routing, and this tracker | Added private typed create, lifecycle, and membership payloads in the required member order and made the application derive compact-JSON SHA-256 values. Removed all three caller hash parameters and exported hash helpers, inlined create bootstrap, privatized patch and last-admin mechanics, moved their white-box evidence to the Incidents package, and deleted `lifecycle_access.go`. Membership creation now rejects a missing selector or a normalized selector/target mismatch before transaction start. Migrated every compiler-discovered mutation caller without an alias or overload. Renamed the remaining HTTP test/row so it describes decoder evidence only and added recurrence policy for the retired surface. No generated output, public route/wire, status, stored response, persisted digest, schema, migration, frontend, or domain-vocabulary behavior changed. | Exact preimage/digest and private-helper row passed at `.cartulary/test-results/20260819T050956Z-p2127473`; full Incidents passed 27/27 at `.cartulary/test-results/20260819T051436Z-p2324385`; Entities, Indicators, Revisions, Collaboration, and server focused slices passed 32/32, 13/13, 27/27, 33/33, and 34/34 at `.cartulary/test-results/20260819T051006Z-p2128014`, `.cartulary/test-results/20260819T051006Z-p2128017`, `.cartulary/test-results/20260819T051006Z-p2128022`, `.cartulary/test-results/20260819T051205Z-p2239132`, and `.cartulary/test-results/20260819T051205Z-p2239134`; their service-backed matrices passed Incidents 19/19, Entities 29/29, Indicators 7/7, Revisions 20/20, Collaboration 23/23, and server 17/17 at `.cartulary/test-results/20260819T051602Z-p2367764`, `.cartulary/test-results/20260819T051602Z-p2367771`, `.cartulary/test-results/20260819T051602Z-p2367778`, `.cartulary/test-results/20260819T051802Z-p2476570`, `.cartulary/test-results/20260819T051802Z-p2476590`, and `.cartulary/test-results/20260819T051802Z-p2476625`; renamed decoder row passed at `.cartulary/test-results/20260819T051958Z-p2602655`; `make format` passed at `.cartulary/test-results/20260819T050707Z-p2073570`; final boundary and Harness gates passed at `.cartulary/test-results/20260819T050742Z-p2081875` and `.cartulary/test-results/20260819T051958Z-p2602816`; `make lint-go` and production retired-symbol searches passed | `I2-RS-02` is `DONE`. Early format/lint attempts exposed one stale Harness selector, unused imports, and unused test helpers; each was corrected and all reruns passed. Generation/drift checks were skipped because no SQL, contract, schema, or generated source changed in this slice. The remaining `requestHash` parameters belong to other owners and are outside this gap. Residual risk is limited to constructor invalid-state and production commit-seam concerns assigned to the successor. Rollback is the private hash/helper implementation, all migrated callers and tests, boundary/Harness inputs, and this checkpoint as one atomic slice. `I2-RS-03` is the sole successor. |

### 17.10 `I2-RS-03` completion evidence

| Time | Baseline and scope | Change | Validation | Result and next action |
| --- | --- | --- | --- | --- |
| 2026-08-19T01:48:10-04:00 | `I2-RS-02` complete on `main` at `ffe49c02d40d04c5e55a5bbc8f4d95276e358f0d`; retained all preceding Iteration 2 work and changed Incidents construction, all 13 compiler-discovered external caller files, server/test composition, test support, boundary/Harness policy, and this tracker | Replaced both constructors with `NewApplication(ApplicationDependencies) (*Application, error)` and the exact `Postgres`/`PreferenceBootstrap` dependency fields. Construction rejects nil and typed-nil Postgres first, then nil and typed-nil bootstrap, before repositories, auth, or admission services are built. Migrated every external caller atomically. Removed `ApplicationOptions`, `NewApplicationWithOptions`, `IncidentCreateCommitPort`, `directIncidentCreateCommit`, and the application commit field. Incident creation now commits its transaction directly. Final-commit fault injection is a test-support `postgres.DB` decorator that returns a wrapped `pgx.Tx`; its `Commit` rolls back and returns `ErrIncidentCreateCommitFault`. Added constructor-order/typed-nil evidence and recurrence rules for the retired surfaces. No generated output, public route/wire, transaction semantics, persisted response/hash, schema, migration, frontend, domain-vocabulary, or compatibility constructor changed or was added. | `make format` passed at `.cartulary/test-results/20260819T052628Z-p2639678`; focused Incidents, server, Entities, Indicators, Revisions, and Timeline slices passed 27/27, 24/24, 32/32, 13/13, 27/27, and 48/48 at `.cartulary/test-results/20260819T052717Z-p2653047`, `.cartulary/test-results/20260819T052717Z-p2653058`, `.cartulary/test-results/20260819T053252Z-p2792188`, `.cartulary/test-results/20260819T053434Z-p2844346`, `.cartulary/test-results/20260819T053515Z-p2859681`, and `.cartulary/test-results/20260819T053622Z-p2903891`; their service-backed matrices passed 19/19, 17/17, 29/29, 7/7, 20/20, and 29/29 at `.cartulary/test-results/20260819T054102Z-p2959451`, `.cartulary/test-results/20260819T054102Z-p2959459`, `.cartulary/test-results/20260819T054544Z-p3095010`, `.cartulary/test-results/20260819T054544Z-p3095006`, `.cartulary/test-results/20260819T054544Z-p3095003`, and `.cartulary/test-results/20260819T054102Z-p2959473`; constructor guards and the atomic rollback test, including incident, membership, both preference, audit, and idempotency residue, passed in the Incidents root; final boundary and Harness gates passed at `.cartulary/test-results/20260819T052641Z-p2650505` and `.cartulary/test-results/20260819T052646Z-p2650958`; `make lint-go`, retired-symbol searches, exact-signature search, 13-caller inventory, and `git diff --check` passed | `I2-RS-03` is `DONE`. Two early lint attempts exposed a missing retained `context` import and one migrated test redeclaration; both implementation errors were corrected and lint reran cleanly. A parallel Timeline run at `.cartulary/test-results/20260819T052717Z-p2653050` had 48/48 rows pass but its visual helper failed to start alongside other browser-bearing owners; the Harness classified it as an infrastructure fixture error, and the isolated full rerun passed. Generation/drift was skipped because this slice changed no SQL, contract, schema, or generated source. Residual risk is limited to HTTP composition/export concerns assigned to the successor. Rollback is the dependency constructor, all caller migrations, test-owned decorator, tests/policy, and this checkpoint as one atomic slice. `I2-RS-04` is the sole successor. |

### 17.11 `I2-RS-04` completion evidence

| Time | Baseline and scope | Change | Validation | Result and next action |
| --- | --- | --- | --- | --- |
| 2026-08-19T01:59:02-04:00 | `I2-RS-03` complete on `main` at `ffe49c02d40d04c5e55a5bbc8f4d95276e358f0d`; retained all preceding Iteration 2 work and changed Incidents HTTP, server composition, tests, boundary/Harness inputs, and this tracker | Replaced variadic `RegisterRoutes(options ...RouteOptions)` with `RegisterRoutes(Dependencies)`. HTTP now owns exact `Application`, `AdmissionChecker`, and `TerminalMutationCoordinator` capabilities containing only methods used by its 11 operations. Server assembly explicitly supplies the Incidents application, an admission checker, and Collaboration terminal-effects coordinator. Registration rejects nil and typed-nil fields in dependency order before loading transport configuration or binding routes. Removed root `incidents.TerminalMutationCoordinator`, `RouteOptions`, exported `Service`, and all six exported `Decode*Request` functions; service and decoder mechanics are private while root resource builders remain exported. Auth/session, key/cursor construction, list/paging, request parsing, envelopes, errors, and response mechanics remain transport-owned. No public HTTP/OpenAPI/WebSocket behavior, status, validation precedence, concealment, session sliding, paging, Location, persisted response/hash, schema, migration, frontend, or compatibility shim changed. | Exact missing/nil/typed-nil dependency row passed at `.cartulary/test-results/20260819T055245Z-p3222424`; focused Incidents, server, and Collaboration slices passed 27/27, 24/24, and 32/32 at `.cartulary/test-results/20260819T055314Z-p3224422`, `.cartulary/test-results/20260819T055447Z-p3268362`, and `.cartulary/test-results/20260819T055545Z-p3307522`; their service-backed matrices passed 19/19, 17/17, and 23/23 at `.cartulary/test-results/20260819T055710Z-p3353483`, `.cartulary/test-results/20260819T055710Z-p3353491`, and `.cartulary/test-results/20260819T055710Z-p3353486`; `make format` passed at `.cartulary/test-results/20260819T055223Z-p3207110`; final boundary and Harness gates passed at `.cartulary/test-results/20260819T055245Z-p3222579` and `.cartulary/test-results/20260819T055245Z-p3222573`; `make lint-go`, exported-surface and retired-symbol/import searches, and `git diff --check` passed | `I2-RS-04` is `DONE`. The first format attempt rejected an overlong authored Harness row ID; the row was shortened without changing its selector and all gates then passed. Generation/drift was skipped because no SQL, schema, contract, or generated source changed. Residual risk is limited to bundle consumer-port ownership and finalizer input determinism assigned to the successor. Rollback is the HTTP capability boundary, server composition, tests/policy, and this checkpoint as one atomic slice. `I2-RS-05` is the sole successor. |

### 17.12 `I2-RS-05` completion evidence

| Time | Baseline and scope | Change | Validation | Result and next action |
| --- | --- | --- | --- | --- |
| 2026-08-19T02:13:12-04:00 | `I2-RS-04` complete on `main` at `ffe49c02d40d04c5e55a5bbc8f4d95276e358f0d`; retained all preceding Iteration 2 work and changed the Incidents provider, Incident Bundles module/source/worker, server assembly, tests, boundary/Harness inputs, and this tracker | Added consumer-owned `incidentbundles/importfinalizerport` with `Params`, `Finalizer`, and `ErrInitialAdminUnavailable`; migrated module, importer, worker, server, and tests without aliases. Incidents still provides the atomic implementation, but its constructor now returns `(importfinalizerport.Finalizer, error)` and rejects nil/typed-nil Workbook writers. Finalization rejects nil/typed-nil transactions and zero incident ID, submitter ID, or publication time before repository construction or database access; the hidden clock fallback is gone and caller publication time is normalized to UTC. Request-ID derivation is private and deterministic in Incident Bundles. The Incidents source port replaces its unchecked prepared-value assertion with a checked conversion wrapping `sourceport.ErrInvalidCatalog`. Deleted the obsolete root `incidents/ports.go`, moved the unavailable-admin sentinel, removed stale allowlist entries, and added exact import/recurrence policy. No public bundle wire/job/error mapping, transaction/audit behavior, schema, migration, persisted hash, frontend, or compatibility bridge changed. | Focused finalizer success/unavailable/rollback and invalid dependency/transaction/zero-value evidence passed 3/3 at `.cartulary/test-results/20260819T060658Z-p3500836`; wrong prepared-type evidence passed at `.cartulary/test-results/20260819T060658Z-p3500850`; deterministic request-ID evidence passed at `.cartulary/test-results/20260819T060658Z-p3500838`; full Incidents, Incident Bundles, and server slices passed 27/27, 8/8, and 24/24 at `.cartulary/test-results/20260819T060759Z-p3518222`, `.cartulary/test-results/20260819T060922Z-p3561849`, and `.cartulary/test-results/20260819T061021Z-p3576865`; their service-backed matrices passed 19/19, 6/6, and 17/17 at `.cartulary/test-results/20260819T061127Z-p3616033`, `.cartulary/test-results/20260819T061127Z-p3616031`, and `.cartulary/test-results/20260819T061127Z-p3616038`; `make format` passed at `.cartulary/test-results/20260819T060459Z-p3482847`; final boundary and Harness gates passed at `.cartulary/test-results/20260819T060635Z-p3500252` and `.cartulary/test-results/20260819T060556Z-p3499037`; `make lint-go`, old symbol/import/file searches, hidden-clock/assertion searches, and `git diff --check` passed | `I2-RS-05` is `DONE`. The first boundary run at `.cartulary/test-results/20260819T060556Z-p3499014` showed that the existing Incident Bundles facade prefix rule also governs its new port subpackage; the three exact Incidents provider/test importers were added and the rerun passed. Generation/drift was skipped because no SQL, schema, contract, or generated source changed. Residual risk is limited to Recovery table ownership and multi-generation interpretation assigned to the successor. Rollback is the consumer port, provider/consumer migration, deterministic validation/request ID, checked conversion, tests/policy, and this checkpoint as one atomic slice. `I2-RS-06` is the sole successor. |

### 17.13 `I2-RS-06` completion evidence

| Time | Baseline and scope | Change | Validation | Result and next action |
| --- | --- | --- | --- | --- |
| 2026-08-19T03:22:12-04:00 | `I2-RS-05` complete on `main` at `ffe49c02d40d04c5e55a5bbc8f4d95276e358f0d`; retained all preceding staged and unstaged Iteration 2 work and changed Workbook/Incidents recovery contributions, Recovery and Graph contracts/runtime, recovery/operator assembly, tests, database-ownership and boundary/Harness inputs, generated projections, and this tracker | Moved `incident_workbook_preferences` and `user_workbook_preferences` from Incidents to a new Workbook contribution, registered Workbook, and raised the exact catalog contribution count to 30 while retaining 113 authored and 84 required tables. Added schema-validated `cartulary.recovery_generation_registry.v1`, immutable current/pre-ownership-v3/historical-v2 catalogs and Graph bindings, exact artifact digests/counts, and generated typed projections. Capture accepts only the current entry. Restore, durability, PostgreSQL unit order, object families, invalidation, consistency, Graph dispatch, replay, and verification basis/evidence select one entry by the exact catalog/codec pair; due-verification computes the expected basis per retained backup rather than repeatedly treating historical backups as changed. Unknown and cross-generation pairs or Graph mixtures fail before target mutation. Frozen catalogs validate their own shape/digest instead of current state. Database ownership generation now derives its closed owner vocabulary from projected entries, so Workbook ownership is reproducible. Generated outputs are `internal/gen/contractrecovery/recovery_generation_registry_gen.go`, updated Recovery/Graph artifact projections, schema-ownership and topology projections. No schema/data migration, immutable SQL edit, public HTTP/OpenAPI/WebSocket or frontend change, persisted Incidents hash/response change, compatibility shim, or open-ended legacy reader was introduced. | Focused Incidents, Workbook, Recovery, Graph Projection, operator, and server slices passed 27/27, 67/67, 24/24, 8/8, 12/12, and 24/24 at `.cartulary/test-results/20260819T070642Z-p27336`, `.cartulary/test-results/20260819T070406Z-p4164503`, `.cartulary/test-results/20260819T070147Z-p4078725`, `.cartulary/test-results/20260819T071344Z-p219125`, `.cartulary/test-results/20260819T070318Z-p4129699`, and `.cartulary/test-results/20260819T072233Z-p336628`; their service-backed matrices passed 19/19, 39/39, 19/19, 6/6, 9/9, and 17/17 at `.cartulary/test-results/20260819T071210Z-p176131`, `.cartulary/test-results/20260819T070946Z-p120324`, `.cartulary/test-results/20260819T070825Z-p70557`, `.cartulary/test-results/20260819T071445Z-p234278`, `.cartulary/test-results/20260819T071538Z-p249276`, and `.cartulary/test-results/20260819T072335Z-p376084`. Final `make generate`, drift, artifact-policy, JSON-shape, and Harness gates passed at `.cartulary/test-results/20260819T072121Z-p328827`, `.cartulary/test-results/20260819T072134Z-p331761`, `.cartulary/test-results/20260819T072144Z-p334707`, `.cartulary/test-results/20260819T072148Z-p335156`, and `.cartulary/test-results/20260819T072154Z-p335702`; boundary passed at `.cartulary/test-results/20260819T071839Z-p293959`; `make format`, `make lint-go`, repository diff checks, contribution/table inventories, and retired-heuristic searches passed. The generation registry digest is `09cb6e77f710577a9b95c0ea21c4c603ef92167c1f7ef5e7a60aad393accae34`. Exact identities are: current Workbook/v3 catalog `96ab9cac942a3729afcefa47a02bbfe910a2c09af0fb25ee32f7b610b6352055`, codec `8fa8c539eabd71ce38b0808ee7176261e32e144ae8fbe0c66ca9bce35f907d47`, Graph registry `61c3f7348c4df2bee3e969c905c91c9857082cf2839a0b57104e40339e3e16d3`, binding `113056a35ec55e42532fca7fd15f557450cd585cd54fb70c657ea1bfb4b61673`; pre-ownership v3 catalog `04174fbf70585af5afb5ace702b797f7512f655005b8502b02f4b41700a966a6`, the same v3 codec/registry, binding `6ec244d0b82466a18adbdb82554be29f5e4baac384175538acbc92e56f14b8d5`; historical v2 catalog `ce0a1f4053a9ce156273e4adf40c8b4185fa616170eadd6a860500d0b24fd22f`, codec `20807db2017de12c86732a11912effdebed32b8a691b032113675f2bc6129352`, registry `e75697ef1f6b5a197d299746fd42d2bf07afcd2e1c9d187a6fe695bca3096730`, binding `235c69bbc0e5d4f25f3fab7b1f2b8c30ba6370bfc65abcba75822007802621b9`. | `I2-RS-06` is `DONE`. Failed runs were retained and resolved as follows: `.cartulary/test-results/20260819T062634Z-p3718200` exposed a newline-versus-canonical Graph-binding digest; `.cartulary/test-results/20260819T063533Z-p3731075` and `.cartulary/test-results/20260819T063722Z-p3785163` exposed one stale test variable and an over-strict partial JSON decode; `.cartulary/test-results/20260819T065211Z-p3871832`, `.cartulary/test-results/20260819T065454Z-p3932253`, and `.cartulary/test-results/20260819T065746Z-p3982187` exposed legacy process capture and two unsuitable failure fixtures, replaced with current capture plus a determinate workbook-probe failure; `.cartulary/test-results/20260819T071634Z-p286706` and `.cartulary/test-results/20260819T072043Z-p328026` exposed the missing generated Workbook owner vocabulary and its source-generator omission; `.cartulary/test-results/20260819T071803Z-p289239` exposed a lexical generated-root-write false positive in a test import; `.cartulary/test-results/20260819T071901Z-p310582` exposed four generator error-string style violations. Every disposition was an implementation/test/Harness correction and passed on rerun; no owner contradiction or scope expansion occurred. Browser, `agent-finalize`, and repository-wide `make check` are deliberately deferred to the authorized final successor. Residual risk is limited to final residue/export review and broad regression validation. Rollback is the three immutable generation artifacts and registry/schema/index, ownership contributions/assembly, generator/generated projections, exact-generation runtime/Graph dispatch, verification/replay scheduling, tests/policy, and this checkpoint as one atomic slice. `I2-RS-07` is the sole successor. |

### 17.14 `I2-RS-07` completion evidence

| Time | Baseline and scope | Change | Validation | Result and next action |
| --- | --- | --- | --- | --- |
| 2026-08-19T04:38:23-04:00 | `I2-RS-06` complete on `main` at `ffe49c02d40d04c5e55a5bbc8f4d95276e358f0d`; retained every preceding staged and unstaged Iteration 2 change and audited final Incidents, Incident Bundles, Recovery, generated SQL, boundary/Harness policy, owner matrices, and changed paths | Removed the now-incompatible SQL/store selector that evaluated all retained backups against one current verification basis; generation-aware catalog scheduling remains the sole path. Made both raw Incidents bundle codec helpers private and changed Incident Bundles to obtain canonical `data/incident.json` through the registered source port, preserving its bytes, incident-key propagation, and incident-not-found mapping. Renamed the last three tests carrying retired helper spellings and migrated their authored selectors. Added recurrence rules for the deleted Recovery selector and bundle codec exports. Regenerated `internal/gen/sql/recovery.sql.go` and affected Harness topology output. Final Incidents inventory is 45 files: 25 production and 20 test, with 114 exported-declaration lines; final Recovery inventory is 47 files: 24 production and 23 test, with 326 exported-declaration lines. Surviving Incidents exports are justified application capabilities and records, source registration, resource builders shared with HTTP, transaction participation, Recovery and Reporting contributions, finalizer construction, and cohesive owner-specific test support. Recovery's generation identity is consumed across its application boundary; generated generation entries are also consumed by Graph Projection. No unused compatibility export remains. | Final focused/service-backed pairs passed for Incidents 27/27 and 19/19 at `.cartulary/test-results/20260819T082455Z-p1617673` and `.cartulary/test-results/20260819T082617Z-p1661332`; Incident Bundles 8/8 and 6/6 at `.cartulary/test-results/20260819T073736Z-p633288` and `.cartulary/test-results/20260819T073834Z-p648545`; Recovery 24/24 and 19/19 at `.cartulary/test-results/20260819T072917Z-p424365` and `.cartulary/test-results/20260819T073053Z-p476243`; Entities 32/32 and 29/29 at `.cartulary/test-results/20260819T074435Z-p721301` and `.cartulary/test-results/20260819T074623Z-p773231`; Workbook 67/67 and 39/39 at `.cartulary/test-results/20260819T074808Z-p825149` and `.cartulary/test-results/20260819T075017Z-p881078`; Indicators 13/13 and 7/7 at `.cartulary/test-results/20260819T075244Z-p936691` and `.cartulary/test-results/20260819T075321Z-p952010`; Graph Projection 8/8 and 6/6 at `.cartulary/test-results/20260819T075359Z-p966854` and `.cartulary/test-results/20260819T075458Z-p981908`; operator 12/12 and 9/9 at `.cartulary/test-results/20260819T075546Z-p996748` and `.cartulary/test-results/20260819T075621Z-p1031155`; Revisions 27/27 and 20/20 at `.cartulary/test-results/20260819T075655Z-p1065089` and `.cartulary/test-results/20260819T075757Z-p1109270`; Timeline 48/48 and 29/29 at `.cartulary/test-results/20260819T075859Z-p1151939` and `.cartulary/test-results/20260819T080333Z-p1208019`; Collaboration 32/32 and 23/23 at `.cartulary/test-results/20260819T080806Z-p1263616` and `.cartulary/test-results/20260819T080928Z-p1309811`; server 24/24 and 17/17 at `.cartulary/test-results/20260819T081050Z-p1355282` and `.cartulary/test-results/20260819T081145Z-p1394021`. Final `make format` and `make generate` passed at `.cartulary/test-results/20260819T082439Z-p1611166` and `.cartulary/test-results/20260819T082443Z-p1614711`; `make agent-finalize` passed 1/1 at `.cartulary/test-results/20260819T082743Z-p1704060` with `RESULTS_DIR` unset and the expected retained-run maintenance skip; browser passed 58/58 at `.cartulary/test-results/20260819T082759Z-p1706930`; boundary 3/3, Harness 2/2, drift 4/4, generated policy 3/3, JSON shape 3/3, and Markdown passed at `.cartulary/test-results/20260819T083206Z-p1758330`, `.cartulary/test-results/20260819T083214Z-p1763380`, `.cartulary/test-results/20260819T083227Z-p1763835`, `.cartulary/test-results/20260819T083235Z-p1766717`, `.cartulary/test-results/20260819T083236Z-p1767121`, and `.cartulary/test-results/20260819T083239Z-p1767596`; `make lint-go` passed; terminal `make check` passed 627/627 at `.cartulary/test-results/20260819T083246Z-p1768497`; exact repository-wide retired-symbol/import/generated-residue scans, ownership scan, export review, changed-path review, and `git diff HEAD --check` passed. The mandatory post-checkpoint Markdown and tracker-scoped diff gates also pass. Registry digest remains `09cb6e77f710577a9b95c0ea21c4c603ef92167c1f7ef5e7a60aad393accae34`. Exact generation identities remain: current Workbook/v3 catalog/codec/Graph registry/binding `96ab9cac942a3729afcefa47a02bbfe910a2c09af0fb25ee32f7b610b6352055` / `8fa8c539eabd71ce38b0808ee7176261e32e144ae8fbe0c66ca9bce35f907d47` / `61c3f7348c4df2bee3e969c905c91c9857082cf2839a0b57104e40339e3e16d3` / `113056a35ec55e42532fca7fd15f557450cd585cd54fb70c657ea1bfb4b61673`; pre-ownership v3 `04174fbf70585af5afb5ace702b797f7512f655005b8502b02f4b41700a966a6` / the same v3 codec and registry / `6ec244d0b82466a18adbdb82554be29f5e4baac384175538acbc92e56f14b8d5`; historical v2 `ce0a1f4053a9ce156273e4adf40c8b4185fa616170eadd6a860500d0b24fd22f` / `20807db2017de12c86732a11912effdebed32b8a691b032113675f2bc6129352` / `e75697ef1f6b5a197d299746fd42d2bf07afcd2e1c9d187a6fe695bca3096730` / `235c69bbc0e5d4f25f3fab7b1f2b8c30ba6370bfc65abcba75822007802621b9`. | `I2-RS-07` and Iteration 2 are `DONE`. The first final Entities run at `.cartulary/test-results/20260819T074012Z-p667242` passed 31/32 and classified the unstarted unit as infrastructure `fixture_error` because `cartulary-test-services start-suite` exited; its isolated rerun passed 32/32. An initial `make explain-run ... DETAIL=rows` investigation used an unsupported detail value; supported summary/log views confirmed the infrastructure classification. The first post-rename `make format` failed before a run root because one authored selector list was not ASCII-sorted; ordering was corrected and every downstream gate reran successfully. Earlier broad acceptance roots were superseded after the test-name cleanup. No required check was skipped; only retained-run maintenance was intentionally skipped because `RESULTS_DIR` was unset. The only retained legacy behavior is the two exact historical Recovery generations, removable only through a separate adopted owner decision backed by repository or deployment evidence. No public HTTP/OpenAPI/WebSocket/session/frontend behavior, persisted response or idempotency hash, schema/data migration, domain vocabulary, alias, overload, forwarding shim, open-ended reader, or unexplained drift remains. Residual risk is limited to operational stewardship of those deliberately retained historical backups. Rollback is the obsolete Recovery SQL/store selector removal, private bundle-codec migration, tests/Harness/policy, regenerated outputs, and this checkpoint as one atomic final slice; preceding completed workstreams retain their recorded rollback boundaries. There is no successor. |

Iteration 2 is complete. Sections 13 through 17 contain no active or pending row.

## 18. Iteration 3: Mutation and Source-State Production Hardening

### 18.1 Charter, authority, and execution boundary

Iteration 3 removes accidental compatibility surfaces and permissive production
seams that survived the first two structural iterations. Its target state is an
Incidents package whose mutation inputs cannot be fabricated in invalid states,
whose audit facts and time are explicit, whose application facade exposes only
application behavior, and whose Incident Bundle and Recovery source state fails
closed against one owner-derived manifest.

The planning direction for this iteration is:

- prefer structural fixes over tactical patches;
- treat future phase growth as a core constraint;
- remove behavior that creates an unnecessary compatibility burden;
- carry a feature forward only when it materially improves the future state; and
- make the subsystem easier to understand, test, extend, and maintain.

Those principles select among owner-compatible implementations; they do not amend
product requirements. Adopted subsystem NLSpecs and Core owner sections remain the
normative behavioral authority. `docs/domain.md` supplies vocabulary and owner
navigation only. `docs/research/nlspec-spec.md` supplies research and writing
guidance only. The prior tracker sections are repository history, not instructions
for this iteration.

The current implementation request authorizes `I3-WF-01`, `I3-RS-00`, `I3-SP-01`,
and `I3-RS-01` through `I3-RS-07` in that exact order. At most one row may be
`IN_PROGRESS`, and each workstream MUST record completion evidence in this tracker
before its successor begins. The existing staged tracker snapshot is user-owned and
MUST NOT be reset, unstaged, overwritten wholesale, or committed automatically.

### 18.2 Current baseline and live inventory

The planning baseline is clean `main` at commit
`57b9805c2d4ac4f2b5003b7f03beea74b18aa344` (`Links Iteration 3`, committed
2026-08-28T08:12:04-04:00).

| Baseline fact | Current evidence |
| --- | --- |
| Incidents Go files | 45 total: 25 production and 20 test |
| Owner routing | 47 rows currently route to `module.incidents` |
| Focused owner baseline | `make test-slice OWNER=module.incidents` passed 27/27 at `.cartulary/test-results/20260828T122009Z-p1521184` |
| Backend boundary baseline | `make backend-module-boundary-check` passed 3/3 at `.cartulary/test-results/20260828T122010Z-p1521266` |
| Repository state at discovery | Clean worktree; local `main` and `origin/main` at the same baseline commit |

The live cleanup inventory is:

| Surface | Current problem | Iteration 3 disposition |
| --- | --- | --- |
| Six mutation request structs and `OptionalNullableString` | Exported mutable values let non-HTTP callers bypass the strict transport decoder and construct contradictory states. | Replace atomically with opaque Incidents-owned admission values; delete the structs and old decoders without aliases. |
| Lifecycle `action string` | The application and Collaboration coordinator accept values outside close/reopen and derive behavior from an open string. | Introduce a closed `LifecycleAction`, bind it into lifecycle admission, and remove the separate string parameter. |
| Audit construction | A hidden `time.Now`, open event/source strings, an implicit public-source default, and `membershipRole(any)` make audit state nondeterministic and permissive. | Require the command time and typed facts at every call site; delete fallback inference and defaults. |
| `Application.GetMembership` | The exported method is used by application replay logic and tests, not by a production consumer capability. | Use the private repository for replay and exact test support for test lookup; remove the method. |
| `Application.ListAdministrativeAuditEvents` | The method only forwards to the platform administrative-audit query for one HTTP handler. | Give the HTTP service a private reader over its existing PostgreSQL dependency; remove the forwarding method and capability. |
| Terminal mutation commit | Exported fields and a public `Validate` method permit callers to fabricate invalid disposition/effect-key pairs. | Make state private and constructor-valid; expose read-only classification/effect-key methods. |
| Incident Bundle source state | Import accepts a generic map, checks only minimal columns, and relies on `jsonb_populate_record`/`SELECT *` behavior that can ignore unknown data or drift with table shape. | Validate an exact typed source row and apply explicit columns from one Incidents manifest. |
| Recovery contribution | Recovery tables and bundle-source facts are separately authored and their constructors cannot report manifest defects. | Derive both from the same validated manifest and make construction fallible. |
| Test support and historical Recovery generations | Much of the support surface is actively shared; two historical generations are owner-required compatibility. | Retain live cohesive support and both historical generations. Delete only members proven dead by exact reachability. |

The existing non-fallible `admission.Checker` remains in scope and MUST NOT be
reopened by this iteration. Root resource builders remain justified shared
Incidents/HTTP capabilities unless a later reachability pass proves otherwise.

### 18.3 Product, persistence, and artifact freeze

For every input valid at this baseline, Iteration 3 MUST preserve:

- all HTTP paths, methods, operation IDs, status codes, Location behavior, success
  envelopes, public errors, fields, reason codes, validation priority, and strict
  JSON behavior;
- authentication, authorization, incident concealment, role and lifecycle policy,
  pagination/cursor scope, session sliding, and WebSocket behavior;
- normalized request-hash preimages and digests, idempotency scopes, persisted
  response status/JSON, replay ordering, conflict behavior, and fresh/replay outcome;
- database schema, authored SQL semantics, lock order, transaction boundaries,
  optimistic concurrency, audit action codes and projection shapes, and committed
  terminal effect identity;
- Collaboration close and membership-revocation effects and their post-commit-only
  ordering;
- Incident Bundle v3 paths and valid exported bytes, actor remapping, source version,
  publication behavior, Reporting snapshots, and Network Flow integration; and
- the Recovery two-table Incidents contribution and all three currently admitted
  Recovery generations.

Two owner closures precede production edits: Core 01 MUST enumerate the exact
Incident Bundle v3 incident row and invariant precedence, and Core 04 MUST define one
recorded UTC mutation timestamp shared by the mutation/source rows and every
raw/projected audit row made observable by a successful commit. Import finalization
MUST use its existing `PublishedAt`; every interactive mutation MUST let the
Incidents application sample its required clock exactly once immediately before a
fresh transaction. These closures change no public field, schema, or timestamp
format.

No schema/data migration, frontend change, durable outbox, public route, new product
feature, generalized compatibility layer, or owner change beyond the two named
closures is planned. Discovery of another required change in those categories marks the affected row
`BLOCKED: owner contradiction or scope expansion` before implementation.

### 18.4 Target mutation interfaces

The root Incidents package MUST own these opaque, nonzero-valid values:

```go
type IncidentCreateAdmission struct { /* private */ }
type IncidentPatchAdmission struct { /* private */ }
type IncidentLifecycleAdmission struct { /* private */ }
type MembershipCreateAdmission struct { /* private */ }
type MembershipPatchAdmission struct { /* private */ }
type MembershipDeleteAdmission struct { /* private */ }

func AdmitIncidentCreateJSON(io.Reader) (IncidentCreateAdmission, *AdmissionError)
func AdmitIncidentPatchJSON(io.Reader) (IncidentPatchAdmission, *AdmissionError)
func AdmitIncidentLifecycleJSON(LifecycleAction, io.Reader) (IncidentLifecycleAdmission, *AdmissionError)
func AdmitMembershipCreateJSON(io.Reader) (MembershipCreateAdmission, *AdmissionError)
func AdmitMembershipPatchJSON(io.Reader) (MembershipPatchAdmission, *AdmissionError)
func AdmitMembershipDeleteJSON(io.Reader) (MembershipDeleteAdmission, *AdmissionError)
```

`AdmissionError` MUST implement `error` and expose only `Field() (string, bool)` and
`ReasonCode() string`. It MUST NOT import or return an HTTP type, status, public error
code, response envelope, or handler concern. Each HTTP handler maps the constructor
it invoked to its existing error family:

| Constructor family | Frozen public error code |
| --- | --- |
| Incident create | `invalid_incident_create` |
| Incident patch | `invalid_incident_patch` |
| Incident lifecycle | `invalid_incident_lifecycle_request` |
| Membership create, patch, and delete | `invalid_mutation_payload` |

The constructors MUST preserve the current unknown-field rejection, single-JSON-value
rule, empty-body behavior, field validation order, normalization, length limits, null
versus omission behavior, role catalog, TLP catalog, reason handling, and request-hash
bytes. Invalid or zero `LifecycleAction` values MUST fail admission. The only valid
actions are `LifecycleActionClose` and `LifecycleActionReopen`, with frozen string
forms `close` and `reopen` used in persisted hash and audit data.

Opaque admissions expose only read access needed outside the root package:

- every idempotent admission exposes `ClientTxnID() string`;
- lifecycle admission exposes `Action() LifecycleAction`;
- membership-create admission exposes mutually exclusive
  `TargetUserID() (uuid.UUID, bool)` and `TargetEmail() (string, bool)` selectors; and
- all other normalized mutation state remains private to the Incidents application.

The application method signatures replace each old request type one-for-one.
Lifecycle transition receives the bound admission without a separate action.
The application receives one required clock through its dependencies and samples it
only after admission, authorization, conflict/no-op, and replay resolution and
immediately before a fresh transaction. Callers do not supply command timestamps.
Zero admission values and zero sampled timestamps MUST fail before transaction
acquisition, mutation writes, audit insertion, or Collaboration effects.

`TerminalMutationCommit` retains the two valid constructors but makes disposition and
effect key private. It exposes `IsNewCommit() bool`, `IsReplay() bool`, and
`EffectKey() (uuid.UUID, bool)`. `Validate`, open disposition constants, and direct
field reads are removed in the same slice. A new commit always carries one nonzero
audit UUID; a replay never carries an effect key.

### 18.5 Audit and facade target state

The private audit fact MUST require:

- actor, target, and incident identities appropriate to the event;
- a closed private event kind;
- an explicit platform public source (`api` or `system` as currently selected);
- explicit `OccurredAt`, normalized to UTC;
- current request/client transaction/reason values with unchanged omission rules;
- raw before/after payloads with unchanged JSON shapes; and
- explicit optional before/after membership roles for the three membership
  projections.

The Incidents event source remains the fixed private value `incidents`; callers MUST
NOT supply it. Event-kind-to-action-code mapping is exhaustive. A membership event
with incomplete projection identities or role facts fails before either audit row is
written. `membershipRole(any)`, its marshal/unmarshal fallback, computed lifecycle
event strings, and the empty-public-source default are deleted.

Every mutation and import finalizer MUST use one occurrence time for its raw journal
row and public projection. Fresh success retains the current transaction ordering;
replay, rejected/no-op mutations, rollback, and commit failure retain their current
no-new-effect behavior.

`CreateMembership` MUST use the private repository for its replay lookup. Tests that
need direct lookup MAY receive one exact helper under Incidents test support, but no
production compatibility method. The membership-audit HTTP service MUST construct a
private reader over `DependencySet.PostgresHandle()` and call the platform query
through that reader. The HTTP `Application` capability then loses the administrative
audit forwarding method.

### 18.6 Source-state manifest and exact validation

Add `internal/modules/incidents/internal/sourcestate` as the sole Incidents manifest
owner for:

- bundle source path `data/incident.json` and its current source version;
- ordered incident columns `id`, `incident_key`, `incident_key_canonical`, `title`,
  `description`, `status`, `severity`, `tlp`, `current_phase`,
  `primary_external_case_ref`, `created_by_user_id`, `created_at`, `updated_at`,
  `updated_by_user_id`, `incident_version`, and `closed_at`;
- ordered invariant IDs `incident.source_identity_admitted`,
  `incident.exact_shape`, `incident.identity_key_lifecycle`, and
  `incident.attribution_version`; and
- Recovery relations `incidents` and `incident_memberships`.

Manifest construction MUST reject an empty/duplicate path, column, invariant, or
Recovery relation. `NewIncidentBundleSourcePort() (sourceport.Port, error)` and
`RecoveryStateContribution() (recoverystate.Contribution, error)` MUST derive their
descriptors from this manifest. Incident portability assembly, revision assembly,
server composition, recovery assembly, and every compiler-discovered caller MUST
propagate construction failure with contextual startup errors. No old overload or
non-fallible wrapper remains.

`Prepare` MUST use the repository strict-object decoder with number preservation,
require one object containing exactly the 16 manifest keys, reject duplicate members
and trailing JSON, and validate in this order:

1. Internal catalog, registered-path/version, operation, and port-binding defects are
   typed internal failures and MUST NOT masquerade as source invariants.
2. `incident.source_identity_admitted`: the row has one canonical nonzero `id` and
   it agrees with the source incident identity and import context.
3. `incident.exact_shape`: no missing, unknown, or duplicate member exists and every
   value has its exact JSON type and nullability.
4. `incident.identity_key_lifecycle`: UUID and import context agree; incident key and
   canonical key agree; status is `active` or `closed`; `closed_at` coheres with that
   status; bounded key/title/metadata/TLP values are valid.
5. `incident.attribution_version`: actor UUIDs admitted by the actor catalog,
   canonical UTC timestamps, timestamp
   relationships, and a positive integer incident version are valid.

Prepared state MUST be a private concrete value bound to the source port identity,
operation, source version, incident, and validated row. `Apply` MUST reject any other
prepared type and use an explicit 16-column insert after the existing actor remap;
it MUST NOT use `jsonb_populate_record`, `INSERT ... SELECT *`, or table-order
inference. `Validate` MUST require exactly one stored incident row and compare its
identity, key/canonical key, lifecycle/closed time, attribution, timestamps, and
version with the prepared row. All failures occur before publication and map to the
declared invariant that owns the failed fact.

Valid export behavior remains byte-compatible, but export MUST select and serialize
the 16 manifest fields explicitly while producing the characterized v3 bytes. A
whole-row `to_jsonb` projection is not an acceptable versioned wire contract.

### 18.7 Iteration requirements

| Requirement ID | Binary requirement |
| --- | --- |
| `INC-I3-001` | Sections 1 through 17 remain immutable history; section 18 is the sole forward plan. |
| `INC-I3-002` | Characterization precedes production edits and freezes every behavior/artifact named in section 18.3. |
| `INC-I3-003` | Core 01 owns the exact 16-member v3 row and invariant precedence before strict codec implementation. |
| `INC-I3-004` | Core 04 owns the one-time-per-command committed timestamp rule before audit implementation. |
| `INC-I3-005` | Only opaque admitted mutation values reach the application; zero or fabricated state fails before effects. |
| `INC-I3-006` | Existing public mutation validation, normalized hashes, replay bytes, and errors remain exact. |
| `INC-I3-007` | Lifecycle action is closed and bound once; no application/coordinator action string remains. |
| `INC-I3-008` | Each successful command uses one explicit UTC time for mutation and audit state. |
| `INC-I3-009` | Audit kinds, sources, and membership role facts are typed and exhaustive with no generic fallback. |
| `INC-I3-010` | The application facade exposes no test-only membership lookup or administrative-audit forwarding method. |
| `INC-I3-011` | Terminal commit state cannot be invalidly fabricated and retains exact post-commit effect semantics. |
| `INC-I3-012` | One immutable manifest derives Incidents bundle and Recovery facts and both constructors fail closed. |
| `INC-I3-013` | Source prepare rejects missing, unknown, duplicate, malformed, incoherent, unbound, or stale incident data before apply. |
| `INC-I3-014` | Source apply names every column, export names every wire member, and validation compares exact stored state. |
| `INC-I3-015` | No alias, overload, wrapper, parallel decoder, hidden clock, or open-ended historical reader is added. |
| `INC-I3-016` | Only exact dead surfaces are removed; live test support and owner-required Recovery generations remain. |
| `INC-I3-017` | Generated artifacts change only through authored inputs and repository-owned generation targets. |
| `INC-I3-018` | An owner contradiction or required public/schema/scope expansion blocks the affected row before implementation. |

### 18.8 Serial execution ledger

The status vocabulary remains `TODO`, `IN_PROGRESS`, `DONE`, and `BLOCKED`. At most
one row may be `IN_PROGRESS`; no successor begins before its predecessor is `DONE`.

| Workstream | Status | Scope and exit evidence | Sole successor |
| --- | --- | --- | --- |
| `I3-WF-00` — tracker activation | DONE | Added the current-plan banner and section 18; Markdown passed at `.cartulary/test-results/20260828T124717Z-p1573854`; tracker diff and changed-path checks prove sections 1–17 unchanged. | `I3-WF-01` |
| `I3-WF-01` — authorized-plan correction | DONE | Corrected only future Iteration 3 material, preserved the staged snapshot and historical sections, and installed the serial owner/implementation/checkpoint plan; Markdown and both staged/working-tree diff checks passed. | `I3-RS-00` |
| `I3-RS-00` — characterization and reachability | DONE | Retained existing mutation/hash/replay/audit/effect/rollback coverage, added exact valid-v3 incident source descriptor/byte evidence, and recorded all migration call sites without changing production behavior. | `I3-SP-01` |
| `I3-SP-01` — owner and projection closure | DONE | Adopted the exact incident-row/invariant and committed-command-time requirements and projected the row schema, source catalog, traceability, claim selectors, and runtime descriptor. | `I3-RS-01` |
| `I3-RS-01` — opaque mutation admission | DONE | Replaced all six mutable request paths with strict Incidents-owned opaque admissions, closed lifecycle actions, private normalized hashes, and fully migrated callers without compatibility aliases. | `I3-RS-02` |
| `I3-RS-02` — audit facts and terminal effects | DONE | Made the application clock, audit facts, and terminal commits closed and explicit; proved one committed timestamp and post-commit-only effects across all mutation families. | `I3-RS-03` |
| `I3-RS-03` — facade contraction | DONE | Moved replay to the private repository and administrative-audit projection reading to a private HTTP reader; removed both application methods and proved no production references remain. | `I3-RS-04` |
| `I3-RS-04` — source manifest and fallible composition | DONE | Added one validated, once-loaded neutral source-state catalog; derived defensive portability/Recovery projections and propagated contextual construction failures through both assemblies. | `I3-RS-05` |
| `I3-RS-05` — strict source codec | DONE | Replaced table-shaped source handling with a strictly decoded, operation-bound 16-member row, explicit apply/export SQL, exact post-remap validation, and routed adversarial/compatibility evidence. | `I3-RS-06` |
| `I3-RS-06` — dead-surface cleanup | DONE | Removed the last open HTTP lifecycle parameter and unnecessary exported strict-decoder error type; proved no compatibility implementation remains and added exact boundary recurrence rules. | `I3-RS-07` |
| `I3-RS-07` — validation and handoff | DONE | Completed all required owner, static, generation, migration, API, security, broad, browser, release, documentation, recurrence, and repository-integrity gates with exact handoff evidence. | Complete |

### 18.9 Workstream execution detail

#### `I3-RS-00` — characterization and reachability

- Add table-driven unit evidence for every accepted/rejected body, validation field,
  reason code, and public error mapping across all six mutations.
- Freeze normalized request-hash preimages/digests and persisted fresh/replay response
  status/JSON, including omission, null, selector, actor, and close/reopen cases.
- Freeze raw/projected audit pairs, action codes, roles, timestamps, request/client
  identities, effect UUIDs, no-op/replay behavior, and transaction rollback.
- Freeze current valid Incident Bundle v3 bytes and add negative rows for every
  missing/unknown/type/identity/key/lifecycle/attribution/version failure with the
  deterministic invariant priority.
- Record exact callers for request types, action strings, terminal fields,
  application forwarding methods, source/Recovery constructors, root exports, and
  every candidate test helper. A candidate without exact zero-caller evidence is
  retained.

#### `I3-SP-01` — owner and typed-projection closure

- Add Core 01 `REQ-01-673` and `AC-566` for the exact 16-member Incident Bundle v3
  row, canonical lexical forms, nullability, closed enums, and invariant precedence.
- Add Core 04 `REQ-04-161` and `AC-567` for one UTC timestamp sampled immediately
  before transactional execution and made observable only by successful commit.
- Add `incident.row.v1.schema.json`, the source-catalog `schema_id`, traceability,
  owner-verification projections, and only the generated outputs owned by those
  inputs. This tracker remains evidence, not specification authority.

#### `I3-RS-01` — opaque mutation admission

- Implement transport-neutral admissions and migrate HTTP, application, idempotency,
  Collaboration coordination, store tests, and test support in one slice.
- Keep target-user resolution after membership-create admission and before the
  application call; bind the normalized selector to the resolved identity exactly as
  today.
- Derive request hashes only inside Incidents from admitted normalized state. Tests
  may assert hashes but MUST NOT receive an override seam or public hash builder.
- Delete every old request definition and HTTP decoder in the same slice. Repository
  searches and an AST/boundary rule prevent their return.

#### `I3-RS-02` — audit facts and terminal effects

- Require the Incidents application clock, sample it once immediately before each
  fresh mutation transaction, and use the admitted import publication time for
  finalization; reject a zero sample before transaction acquisition.
- Replace generic audit strings/maps used for projection inference with closed facts,
  retaining raw before/after JSON and public projection bytes.
- Close terminal commit state and migrate coordinator/tests to query methods.
- Do not contract the facade in this slice; that is the sole scope of `I3-RS-03`.

#### `I3-RS-03` — facade contraction

- Replace application replay use of `GetMembership` with the private repository and
  give tests exact store support without a production compatibility method.
- Give the membership-audit HTTP service a private projection reader over its
  existing PostgreSQL dependency and remove the application forwarding capability.
- Prove by package-boundary and reachability checks that transport projection policy
  and repository details did not leak into the application facade.

#### `I3-RS-04` — source manifest and fallible composition

- Add and unit-test the private manifest before changing constructors.
- Make source and Recovery constructors fallible and migrate all assemblies before
  enabling strict codec validation; startup must fail before listener registration.
- Derive ordered bundle and Recovery facts from the same immutable catalog and expose
  defensive copies only.

#### `I3-RS-05` — strict source codec

- Replace permissive source prepare/apply with bound typed state and explicit SQL;
  replace whole-row export with an explicit 16-member projection.
- Validate actor admission, exact stored post-remap state, affected-row equality,
  valid-byte compatibility, publication, rollback, and invariant precedence.
- Add owner-routed unit/integration rows through authored Harness inputs when the
  existing selector surface cannot express the new evidence; regenerate only then.

#### `I3-RS-06` — exact cleanup and recurrence prevention

- Repeat production and test call-graph/export scans after all migrations.
- Delete only zero-caller helpers, obsolete fixtures, compatibility wording, and
  superseded boundary allowances. Do not relocate cohesive support merely to reduce
  file/export counts.
- Add exact subtree rules forbidding retired DTO/decoder names, action strings at the
  application boundary, `time.Now` in Incidents audit/application code, generic audit
  role extraction, non-fallible source constructors, `jsonb_populate_record`, and
  wildcard source inserts.

#### `I3-RS-07` — final validation and handoff

- Record final production/test/export inventory, source-manifest facts, retained
  Recovery generations, substantive changes, generated outputs, command roots, every
  failure and disposition, skipped checks, residual risk, and rollback boundary.
- Close the iteration only after all binary criteria and recurrence searches pass.
  A nearly complete or budget-limited run remains active; it is not marked `DONE`.

### 18.10 Verification routing

Before each slice, run `make task-guide ROLE=module-author OWNER=<owner-id>` and use
`make explain-test-owner`, `make explain-target`, or `make target-plan` to select the
narrowest evidence. The minimum owner routing is:

| Slice | Required focused and service-backed owners |
| --- | --- |
| `I3-RS-00` | `module.incidents`, plus the characterized Collaboration, Incident Bundles, Recovery, Workbook, and server rows |
| `I3-SP-01` | `module.incidents`, `module.incidentbundles`, and every owner whose authored verification projection changes |
| `I3-RS-01` | `module.incidents`, `module.collaboration`, `app.server`, and compiler-discovered direct callers |
| `I3-RS-02` | `module.incidents`, `module.collaboration`, `app.server`, and administrative-audit consumer rows |
| `I3-RS-03` | `module.incidents`, `app.server`, and administrative-audit consumer rows |
| `I3-RS-04` | `module.incidents`, `module.incidentbundles`, `module.recovery`, `module.workbook`, `app.server`, and affected assembly rows |
| `I3-RS-05` | `module.incidents`, `module.incidentbundles`, `module.recovery`, and service-backed import/export rows |
| `I3-RS-06` | Every owner changed by cleanup, boundary policy, or Harness routing |
| `I3-RS-07` | All owners changed during Iteration 3 |

Every production slice runs its focused/service-backed owner pairs,
`make backend-module-boundary-check`, the Go lint target selected by `make help-all`,
and `git diff --check`. Harness inputs change only for real ownership/selector/topology
changes and then require `make harness-contract`. Authored SQL, contract, or generator
input changes require `make generate`, followed by `make generate-drift`,
`make generated-artifact-policy-check`, and `make json-shape-check`.

Final validation order is:

1. focused and service-backed slices for every changed owner;
2. `make agent-finalize`, recording the expected retained-run skip when
   `RESULTS_DIR` is unset;
3. `make backend-module-boundary-check` and `make harness-contract`;
4. `make generate-drift`, `make generated-artifact-policy-check`, and
   `make json-shape-check`;
5. `make migration-drift` and `make openapi-compatibility-check`;
6. `make go-gosec-targeted` and `make go-vulncheck`;
7. `make test-fast` and `make check`;
8. `make browser-e2e` and `make browser-e2e-webserver-backed`;
9. `make release-check` and `make lint-markdown`; and
10. retired-symbol/import/permissive-source searches, changed-path review, and
    repository-wide `git diff --check`.

For document-only `I3-WF-00` and `I3-WF-01`, validation is limited to `make lint-markdown`,
`git diff --check -- docs/handoffs/incidents-module-refactor-tracker.md`, a
tracker-only changed-path review, and a diff review proving sections 1 through 17 are
unchanged.

### 18.11 Binary completion criteria and deferrals

| Criterion | Required result |
| --- | --- |
| Historical integrity | Sections 1–17 are unchanged and section 18 alone controls forward work. |
| Admission | Six opaque admission values are the only mutation inputs; invalid/zero state cannot reach effects. |
| Compatibility | Public validation/errors, hashes, stored replay bytes, valid bundle bytes, and transaction/effect ordering match characterization. |
| Lifecycle | Only close/reopen can be represented and the action is bound once. |
| Audit | One explicit UTC command time and closed audit facts produce the existing raw/public shapes. |
| Facade | No test-only membership read or platform audit forwarding method remains. |
| Terminal effects | Invalid commit dispositions cannot be constructed and fresh/replay notifications remain exact. |
| Source state | One validated manifest derives bundle/Recovery facts; prepare/apply/validate is exact, bound, explicit, and fail-closed. |
| Cleanup | Removed identifiers, aliases, overloads, hidden clocks, generic fallbacks, and permissive source operations have zero production references and recurrence guards. |
| Retained behavior | Active test support and all three owner-required Recovery generations remain intact. |
| Verification | Every required narrow, static, security, drift, browser, broad, and release gate passes or blocks the dependent row. |
| Handoff | Final inventory, evidence roots, failures, skips, residual risk, and rollback boundaries are recorded. |

Explicit deferrals are a database/schema migration, public API or frontend change,
new lifecycle states, new incident fields, durable/multi-process Collaboration
delivery, removal or reinterpretation of historical Recovery generations, redesign of
the neutral admission checker, and unrelated test-support relocation. Each requires a
separate owner-backed iteration.

### 18.12 `I3-WF-00` activation evidence

| Time | Baseline and scope | Change | Validation | Result and next action |
| --- | --- | --- | --- | --- |
| 2026-08-28T08:48:13-04:00 | Clean `main` at `57b9805c2d4ac4f2b5003b7f03beea74b18aa344`; document-only authorization with a 45-file Incidents inventory and 47 owner-routed rows | Added the current-plan banner and controlling Iteration 3 section after the completed Iteration 2 record. No production, test, owner, contract, SQL, generated, Harness, migration, frontend, domain, or research file changed. | Baseline `module.incidents` focused slice passed 27/27 at `.cartulary/test-results/20260828T122009Z-p1521184`; backend boundary passed 3/3 at `.cartulary/test-results/20260828T122010Z-p1521266`; `make lint-markdown` passed at `.cartulary/test-results/20260828T124717Z-p1573854`; tracker-scoped `git diff --check`, changed-path review, and zero-context diff review passed | `I3-WF-00` is `DONE`. Only this tracker is modified, sections 1–17 remain immutable history, and `I3-RS-00` is the sole planned successor but remains unauthorized and `TODO`. |

### 18.13 Workstream checkpoint protocol and execution evidence

For every remaining row, confirm the predecessor is `DONE`, mark only the current
row `IN_PROGRESS`, execute only that scope, and run its narrow gates. The final
repository change for the row MUST update this tracker with status, substantive
edits, exact commands and results, run roots, skips, residual risk, compatibility
impact, and rollback boundary. Then run Markdown lint and tracker-scoped diff-check
without modifying files. No successor starts until that checkpoint is `DONE`.

| Time | Workstream | Change | Validation | Result and next action |
| --- | --- | --- | --- | --- |
| 2026-08-28T13:27:57-04:00 | `I3-RS-07` | Completed the end-to-end handoff audit. Final changed areas are the adopted Core 01/Core 04 and Appendix F owner text; Incident Bundle row schema, source catalog, and traceability projections; Incidents admission/application/audit/facade/source-state/codec/Recovery code; Collaboration and server composition; strict JSON support; migrated owner tests/test support; authored Harness and boundary inputs; the generator-produced execution-topology render index; and this tracker. No authored SQL, database migration/schema, frontend source, `docs/domain.md`, Recovery generation, public OpenAPI contract, or compatibility shim changed. The final source inventory retains one exact 16-member v3 row, one neutral immutable source/Recovery catalog, one designated Incident Bundles adapter, explicit insert/export/validation SQL, six opaque admission roots, closed audit/terminal/lifecycle facts, and active owner-routed characterization/adversarial evidence. The pre-existing staged tracker snapshot remained staged and untouched; no index operation or commit was performed. | `make agent-finalize` passed 1/1 at `.cartulary/test-results/20260828T162004Z-p4130266`; retained-run maintenance was intentionally skipped because `RESULTS_DIR` was unset (`results-dir-not-provided`). Final static/contract roots: boundary 3/3 `.cartulary/test-results/20260828T162045Z-p4133410`, Harness 2/2 `.cartulary/test-results/20260828T162045Z-p4133425`, drift 4/4 `.cartulary/test-results/20260828T162104Z-p4134338`, generated policy 3/3 `.cartulary/test-results/20260828T162104Z-p4134348`, JSON shape 3/3 `.cartulary/test-results/20260828T162104Z-p4134367`, migration drift 5/5 `.cartulary/test-results/20260828T162117Z-p4138122`, OpenAPI compatibility 4/4 `.cartulary/test-results/20260828T162117Z-p4138125`, targeted gosec 4/4 `.cartulary/test-results/20260828T162131Z-p4141887`, and vulnerability analysis 4/4 `.cartulary/test-results/20260828T162131Z-p4141888`. Final focused/service evidence passed for Incidents 28/28 and 19/19 (`.cartulary/test-results/20260828T161418Z-p4002742`, `.cartulary/test-results/20260828T161548Z-p4050578`), Incident Bundles 8/8 and 6/6 (`.cartulary/test-results/20260828T160257Z-p3887754`, `.cartulary/test-results/20260828T160525Z-p3950728`), Recovery 24/24 and 19/19 (`.cartulary/test-results/20260828T155104Z-p3626294`, `.cartulary/test-results/20260828T155451Z-p3742574`), Collaboration 32/32 and 23/23 (`.cartulary/test-results/20260828T162208Z-p4173415`, `.cartulary/test-results/20260828T162342Z-p28820`), server 24/24 and 17/17 (`.cartulary/test-results/20260828T162512Z-p77404`, `.cartulary/test-results/20260828T162613Z-p118895`), Workbook 66/66 and 37/37 (`.cartulary/test-results/20260828T162715Z-p160246`, `.cartulary/test-results/20260828T162932Z-p218892`), Entities 40/40 and 31/31 (`.cartulary/test-results/20260828T163148Z-p277391`, `.cartulary/test-results/20260828T163340Z-p332487`), Indicators 20/20 and 8/8 (`.cartulary/test-results/20260828T163537Z-p387584`, `.cartulary/test-results/20260828T163624Z-p404677`), Network Flow 35/35 and 30/30 (`.cartulary/test-results/20260828T163711Z-p421590`, `.cartulary/test-results/20260828T164155Z-p536937`), Revisions 27/27 and 20/20 (`.cartulary/test-results/20260828T164407Z-p594416`, `.cartulary/test-results/20260828T164517Z-p640403`), Saved Views 25/25 and 24/24 (`.cartulary/test-results/20260828T164629Z-p686139`, `.cartulary/test-results/20260828T164744Z-p736277`), and Timeline 52/52 and 29/29 (`.cartulary/test-results/20260828T164857Z-p786341`, `.cartulary/test-results/20260828T165336Z-p845103`). Broad/release roots passed: test-fast 439/439 `.cartulary/test-results/20260828T165820Z-p903952`, check 669/669 `.cartulary/test-results/20260828T165836Z-p904759`, browser 54/54 `.cartulary/test-results/20260828T170409Z-p1018897`, webserver-backed browser 58/58 `.cartulary/test-results/20260828T170918Z-p1092653`, and release 829/829 `.cartulary/test-results/20260828T171322Z-p1147075`. `make lint-markdown` passed before the final checkpoint at `.cartulary/test-results/20260828T172816Z-p1364558`; exact recurrence scans and working/staged repository diff checks passed. | `DONE`; every section 18.11 binary criterion is satisfied. The first Network Flow service-backed attempt at `.cartulary/test-results/20260828T163926Z-p479251` passed 27 units but its stateful browser pre-reset failed with `recovery_reset_truncate_failed` while PostgreSQL was saturated; the immediately isolated run passed 30/30, dispositioning this as unrelated fixture contention. No related failure or known residual product risk remains. Valid HTTP/OpenAPI behavior, request hashes, audit shapes, database schema, current Recovery identities, and valid bundle-v3 bytes remain compatible; strict rejection of malformed/defaulted command or bundle input is intentional. Roll back only as an owner-consistent unit spanning the adopted requirements, typed projections, implementation/caller migrations, tests, boundary/Harness inputs, generated topology index, and tracker; partial rollback would recreate specification or source-state drift. Iteration 3 is complete. |
| 2026-08-28T12:18:59-04:00 | `I3-RS-06` | Audited all Incidents production and test declarations/callers after the migrations. No retired DTO, nullable wrapper, old request-hash function, facade pass-through, terminal-result field/constant, permissive source helper, or obsolete compatibility-only test path remained, so no live helper or cohesive test support was deleted. Replaced the private HTTP lifecycle handler's arbitrary string with the closed `LifecycleAction` and made illegal-transition mapping exhaustive; made the strict JSON duplicate-member concrete error private while retaining its read-only query. Added source-boundary recurrence rules for all retired command/hash, audit/facade/terminal, permissive/non-fallible source, hidden mutation-clock, and open lifecycle-parameter forms. Retained the public OpenAPI schema name, active source characterization, current helpers, and all Recovery generations because each remains owner-required evidence or behavior. | Final `make format` passed at `.cartulary/test-results/20260828T161719Z-p4096986`; Incidents focused and service-backed slices passed 28/28 and 19/19 at `.cartulary/test-results/20260828T161418Z-p4002742` and `.cartulary/test-results/20260828T161548Z-p4050578`. `make backend-module-boundary-check` passed 3/3 at `.cartulary/test-results/20260828T161719Z-p4096925`; `make generate-drift` passed 4/4 at `.cartulary/test-results/20260828T161718Z-p4096631`; `make lint` passed 11/11 at `.cartulary/test-results/20260828T161719Z-p4097098`; `make test-fast` passed 439/439 at `.cartulary/test-results/20260828T161719Z-p4097115`; and `make json-shape-check` passed 3/3 at `.cartulary/test-results/20260828T161852Z-p4128418`. Exact repository scans returned no retired mutation type/hash declaration, facade/terminal/audit surface, permissive Incident source SQL, non-fallible Incidents source signature, hidden mutation clock, or open lifecycle parameter. Working and staged repository `git diff --check` passed. | `DONE`; HTTP payloads/statuses/errors, OpenAPI, hashes, database schema, valid v3 bytes, and Recovery generations are unchanged. No compatibility shim remains, and no generated artifact or authored Harness/SQL input changed in this slice. Residual risk is limited to failures that may be exposed by the mandatory repository-wide final matrix. Rollback is the typed private HTTP lifecycle parameter, private duplicate error type, boundary rules, and this checkpoint as one unit. `I3-RS-07` is the sole successor. |
| 2026-08-28T09:25:05-04:00 | `I3-WF-01` | Corrected the authorized future plan without changing Sections 1–17 or the completed `I3-WF-00` record. Split owner closure, audit/effect, facade, source-manifest, source-codec, cleanup, and final validation into serial rows; preserved the existing staged snapshot without index operations. | `make lint-markdown` passed at `.cartulary/test-results/20260828T132412Z-p1586717`; working-tree and staged tracker `git diff --check` passed; zero-context review located working-tree changes only in the banner and section 18. | `DONE`; no behavior or contract changed. Rollback is the unstaged Iteration 3 correction only. `I3-RS-00` is the sole successor. |
| 2026-08-28T09:35:28-04:00 | `I3-RS-00` | Added `incidents_source_characterization_test.go` to freeze the Incidents descriptor, v3 path/version/stable identity, four invariant registrations, exact canonical valid-row bytes, actor catalog, and prepared-value binding. Existing retained tests already freeze all six mutation validation families, create/lifecycle/membership hash digests, normalized replay, raw/projected audit before/after facts, rollback, close/reopen, and terminal effect identities. Reachability found mutable request symbols in 29 files, lifecycle transition strings in 6, terminal commit surface in 6, the two facade names in 10 repository files including unrelated Auth stores, and the Incidents source/recovery constructors plus permissive SQL in 8 exact sites. | `make format` passed at `.cartulary/test-results/20260828T132812Z-p1590291`; focused slices passed: `module.incidentbundles` 8/8 at `.cartulary/test-results/20260828T132831Z-p1594833`, `module.incidents` 27/27 at `.cartulary/test-results/20260828T132933Z-p1612048`, `module.collaboration` 32/32 at `.cartulary/test-results/20260828T133106Z-p1658050`, `app.server` 24/24 at `.cartulary/test-results/20260828T133106Z-p1658073`, and `module.workbook` 66/66 at `.cartulary/test-results/20260828T133106Z-p1658095`. The concurrent Recovery run had three PostgreSQL force-drop cleanup timeouts at `.cartulary/test-results/20260828T133106Z-p1658066`; isolated rerun passed 24/24 at `.cartulary/test-results/20260828T133406Z-p1860602`, classifying the first failure as concurrent harness resource contention unrelated to this test-only change. | `DONE`; production and public behavior are unchanged. Service-backed slices are deferred to the production workstreams that change their paths. Rollback is removal of the one characterization test and this evidence row. `I3-SP-01` is the sole successor. |
| 2026-08-28T09:44:26-04:00 | `I3-SP-01` | Adopted Core 01 `REQ-01-673` and `AC-566` for the exact 16-member Incident Bundle v3 singleton row, canonical forms, fixed invariant precedence, prepared binding, explicit apply/validate/export, actor remap, and no physical-table wire authority. Adopted Core 04 `REQ-04-161` and `AC-567` for one UTC recorded mutation time made observable only on commit. Updated Base/Incident-Portability claim selectors and Appendix F bidirectional/profile traceability. Added `incident.row.v1.schema.json`, its source-catalog `schema_id`, AC-566 machine traceability to existing Incidents and Incident Bundles verification IDs, and the matching runtime descriptor projection. `docs/domain.md`, database schema, generated roots, and the existing verification-ID registry remained unchanged because its incident-portability and actor IDs already own the new evidence. | Exact ID uniqueness and `jq empty` passed; `make lint-markdown` passed at `.cartulary/test-results/20260828T134230Z-p1917556`; `make json-shape-check` passed 3/3 at `.cartulary/test-results/20260828T134230Z-p1917256`; `make generate-drift` passed 4/4 at `.cartulary/test-results/20260828T134230Z-p1917246`; `make harness-contract` passed 2/2 at `.cartulary/test-results/20260828T134230Z-p1917575`; focused `module.incidentbundles` passed 8/8 at `.cartulary/test-results/20260828T134251Z-p1922084`; focused `module.incidents` passed 27/27 at `.cartulary/test-results/20260828T134251Z-p1922091`; repository `git diff --check` passed. | `DONE`; valid v3 bytes and public/database contracts are unchanged. Rollback is the two owner blocks, two criteria, Appendix F/claim edits, schema/catalog/traceability additions, and descriptor `schema_id` as one owner-projection unit. `I3-RS-01` is the sole successor. |
| 2026-08-28T10:16:48-04:00 | `I3-RS-01` | Replaced the six exported mutable mutation DTOs and `OptionalNullableString` with opaque admitted values whose state, optional-null representation, and SHA-256 digests remain private to Incidents. Added strict root factories over one duplicate/unknown/trailing-safe decoder, a closed close/reopen action, private admission errors with stable queries, and fail-fast zero-value guards. Migrated HTTP, application, Collaboration effects, commit-fault support, performance/store fixtures, and every compiler-discovered cross-module test caller without aliases or hash overrides. HTTP now only maps owner admission errors and resolves an already-normalized membership selector; normalization and hash golden evidence moved to Incidents. The pre-existing public OpenAPI schema name remains intentionally unchanged. | `make format` passed finally at `.cartulary/test-results/20260828T141615Z-p2302072`; focused slices passed: `module.incidents` 27/27 at `.cartulary/test-results/20260828T140259Z-p2048498`, `module.collaboration` 32/32 at `.cartulary/test-results/20260828T140430Z-p2094712`, and `app.server` 24/24 at `.cartulary/test-results/20260828T140600Z-p2143804`. `make backend-module-boundary-check` passed 3/3 at `.cartulary/test-results/20260828T140701Z-p2185688`; `make test-fast` first found one stale Saved Views import at `.cartulary/test-results/20260828T140704Z-p2186108`, then passed 438/438 at `.cartulary/test-results/20260828T140817Z-p2198557`. `make backend-unit` first found one stale Workbook import at `.cartulary/test-results/20260828T141021Z-p2209224`, then passed 144/144 with strict-framing, zero-admission, normalization, and admitted-hash tests at `.cartulary/test-results/20260828T141413Z-p2264874`. `make lint` first found staticcheck S1016 in the earlier characterization fixture at `.cartulary/test-results/20260828T141526Z-p2288174`; after the equivalent conversion it passed 11/11 at `.cartulary/test-results/20260828T141619Z-p2306159`. Retired-symbol, old-decoder/hash-helper, and open lifecycle-boundary scans returned zero matches; working and staged `git diff --check` passed. | `DONE`; public HTTP/OpenAPI behavior and all previously valid hash preimages remain byte-stable, while malformed or bypassed internal construction now fails before transaction/effect access. No database, frontend, bundle-byte, or generated-artifact change belongs to this slice. Residual timing/audit and terminal-result closure remains isolated to `I3-RS-02`. Rollback is the admission implementation plus its application/HTTP/Collaboration/test-support caller migration as one compile-atomic unit. `I3-RS-02` is the sole successor. |
| 2026-08-28T10:51:54-04:00 | `I3-RS-02` | Made the Incidents mutation clock a required application dependency and sampled it once, in UTC, immediately before each fresh transaction after admission/replay/no-op resolution; import finalization continues to use its admitted publication time. Every successful domain/source/bootstrap, raw audit, and projected audit write now receives that one occurrence time. Replaced open audit kind/source strings, implicit defaults, and JSON role inference with exhaustive private kinds, sources, and explicit optional before/after roles. Made terminal commit state private and constructor-valid, with read-only new/replay/effect queries; coordinators dispatch close and membership-revocation effects only after a new commit. Added exhaustive mapping/constructor tests and all-six-family timestamp evidence, including replay clock/audit stability. | Final `make format` passed at `.cartulary/test-results/20260828T144009Z-p2597297`; `make backend-unit` passed 144/144 at `.cartulary/test-results/20260828T144019Z-p2601361`; `make test-fast` passed 438/438 at `.cartulary/test-results/20260828T144151Z-p2627769`. Focused/service-backed pairs passed: `module.incidents` 27/27 at `.cartulary/test-results/20260828T144208Z-p2628860` and 19/19 at `.cartulary/test-results/20260828T144333Z-p2674830`; `module.collaboration` 32/32 at `.cartulary/test-results/20260828T144505Z-p2720403` and 23/23 at `.cartulary/test-results/20260828T144632Z-p2768986`; `app.server` 24/24 at `.cartulary/test-results/20260828T144759Z-p2817507` and 17/17 at `.cartulary/test-results/20260828T144854Z-p2858876`. `make backend-module-boundary-check` passed 3/3 at `.cartulary/test-results/20260828T145106Z-p2913386`; `make lint` passed 11/11 at `.cartulary/test-results/20260828T145108Z-p2913810`. Hidden-clock, obsolete terminal-field/constant, fallback-role, and open audit-call-site scans found no retired production path; working and staged `git diff --check` passed. | `DONE`; public HTTP/OpenAPI and audit field shapes remain unchanged, while successful new writes intentionally gain one exact command timestamp. Replay, rejection, no-op, rollback, commit failure, and post-commit response failure add no duplicate audit/effect state; no database migration, backfill, frontend, bundle-byte, or generated change belongs to this slice. Facade contraction remains isolated to `I3-RS-03`. Rollback is the application clock boundary, closed audit facts, closed terminal result, and their migrated callers/tests as one compile-atomic unit. `I3-RS-03` is the sole successor. |
| 2026-08-28T11:04:11-04:00 | `I3-RS-03` | Removed `Application.GetMembership` and changed membership-create replay to call the private repository directly. Added an unexported, projection-only HTTP membership-audit reader over the route `DependencySet` PostgreSQL handle, made route construction fail closed when that handle is absent, and removed `Application.ListAdministrativeAuditEvents` from both the concrete application and HTTP capability. Replaced external-package membership assertions with one exact Incidents test-support query; no production compatibility wrapper or transport policy was introduced. | `make format` passed finally at `.cartulary/test-results/20260828T145734Z-p2952669`. Focused/service-backed pairs passed: `module.incidents` 27/27 at `.cartulary/test-results/20260828T145738Z-p2956683` and 19/19 at `.cartulary/test-results/20260828T145909Z-p3002819`; `app.server` 24/24 at `.cartulary/test-results/20260828T150034Z-p3048471` and 17/17 at `.cartulary/test-results/20260828T150129Z-p3090174`. The first `make backend-unit` found one now-unused test variable at `.cartulary/test-results/20260828T145540Z-p2925942`; after its exact removal, the rerun passed 144/144 at `.cartulary/test-results/20260828T150229Z-p3131324`. `make backend-module-boundary-check` passed 3/3 at `.cartulary/test-results/20260828T150342Z-p3153629`; `make lint` passed 11/11 at `.cartulary/test-results/20260828T150344Z-p3154053`. Removed-method scans found no application definition or production call; the sole Incidents administrative-audit query is in the private HTTP reader and the sole replay lookup is private-repository access. Working and staged `git diff --check` passed. | `DONE`; public routes, responses, authorization, pagination, and stored projections are unchanged. The intentional internal compile-time removal narrows the facade; no schema, migration, frontend, generated, or bundle change belongs to this slice. Rollback is the two facade removals, private HTTP reader, and exact test-helper migration as one compile-atomic unit. `I3-RS-04` is the sole successor. |
| 2026-08-28T11:21:05-04:00 | `I3-RS-04` | Added `incidents/internal/sourcestate` as the once-loaded, validated owner of source family/contract/owner identity, singleton path and schema generation, the ordered 16-column row, fixed invariant precedence, and ordered `incidents`/`incident_memberships` Recovery relations. Its projections use defensive copies and neutral scalar types; only the Incidents root projects them into Incident Bundles and Recovery owner ports. Validation rejects missing, duplicate, reordered, unsafe, or generation-incoherent paths, columns, identities, invariants, and relations. Made both Incidents constructors fallible, added contract validation, migrated every caller, and wrapped failures with Incidents owner context in portability and Recovery assembly. Added source/Recovery agreement, immutability, malformed-catalog, and contextual-failure tests. | `make format` passed finally at `.cartulary/test-results/20260828T151913Z-p3401865`; final `make test-fast` passed 438/438 at `.cartulary/test-results/20260828T151916Z-p3405975`. Serial focused slices passed: `module.incidents` 27/27 at `.cartulary/test-results/20260828T151052Z-p3182864`, `module.incidentbundles` 8/8 at `.cartulary/test-results/20260828T151217Z-p3229290`, `module.recovery` 24/24 at `.cartulary/test-results/20260828T151313Z-p3246102`, `module.workbook` 66/66 at `.cartulary/test-results/20260828T151429Z-p3300288`, and `app.server` 24/24 at `.cartulary/test-results/20260828T151641Z-p3358983`; Recovery retained the frozen catalog digest. The first boundary run correctly rejected an Incident Bundles owner-port import from the internal catalog at `.cartulary/test-results/20260828T151745Z-p3400630`; after replacing it with neutral facts and root projection, `make backend-module-boundary-check` passed 3/3 at `.cartulary/test-results/20260828T152008Z-p3414718`. `make generate-drift` passed 4/4 at `.cartulary/test-results/20260828T152010Z-p3415064`; `make lint` passed 11/11 at `.cartulary/test-results/20260828T152019Z-p3418083`. Import scans prove the internal catalog has no Incident Bundles or Recovery owner-port coupling; constructor scans prove all production callers handle errors; working and staged `git diff --check` passed. | `DONE`; source identity, v3 path/schema, Recovery table membership, catalog digest, and all retained generation identities are unchanged. Startup now fails with owner context for invalid metadata before route/listener publication. No database, migration, frontend, bundle-byte, or generated-artifact change belongs to this slice. Codec-specific path/invariant consumers remain solely for replacement in `I3-RS-05`. Rollback is the neutral catalog, two fallible constructors, assembly propagation, and their tests/caller migrations as one compile-atomic unit. `I3-RS-05` is the sole successor. |
| 2026-08-28T12:08:28-04:00 | `I3-RS-05` | Replaced permissive Incident source preparation with a private typed row decoded through the shared strict-object boundary, including duplicate-member reporting, bounded UTF-8 input, exact 16-member types/nullability, canonical UUID/timestamp/version forms, lifecycle/key coherence, actor admission, and the adopted invariant precedence. Prepared state is bound to owner/port, path/schema, operation, incident, bundle version, and contract major. Apply now records both original actor attributions, remaps both stored actors, executes one explicit 16-column insert, and requires one affected row. Validation requires one row and compares all 16 stored facts after remap. Export uses an explicit fixed 16-member projection and existing canonical JSON. The codec has no Incident Bundles transport dependency; the single owner source-port adapter translates private catalog/binding/invariant results. Added adversarial codec, catalog, composition, timestamp, audit/admission, attribution, exact-state, and byte-stability tests to authored owner routing, then generated the topology projection. | `make generate` passed at `.cartulary/test-results/20260828T154544Z-p3477855`; final `make format` passed at `.cartulary/test-results/20260828T160059Z-p3835754`. The strict-codec row passed at `.cartulary/test-results/20260828T160108Z-p3839867`. Final focused slices passed: Incidents 28/28 at `.cartulary/test-results/20260828T160126Z-p3841132`, Incident Bundles 8/8 at `.cartulary/test-results/20260828T160257Z-p3887754`, and Recovery 24/24 at `.cartulary/test-results/20260828T155104Z-p3626294`. Final service-backed slices passed: Incidents 19/19 at `.cartulary/test-results/20260828T160358Z-p3904950`, Incident Bundles 6/6 at `.cartulary/test-results/20260828T160525Z-p3950728`, and Recovery 19/19 at `.cartulary/test-results/20260828T155451Z-p3742574`. `make harness-contract` and `make generate-drift` passed at `.cartulary/test-results/20260828T154737Z-p3488238` and `.cartulary/test-results/20260828T154737Z-p3488089`; `make json-shape-check` and `make go-gosec-targeted` passed at `.cartulary/test-results/20260828T155621Z-p3796307` and `.cartulary/test-results/20260828T155621Z-p3796679`; final boundary, lint, generated-policy, and test-fast gates passed 3/3, 11/11, 3/3, and 439/439 at `.cartulary/test-results/20260828T160117Z-p3840730`, `.cartulary/test-results/20260828T160647Z-p3968104`, `.cartulary/test-results/20260828T160647Z-p3967824`, and `.cartulary/test-results/20260828T160647Z-p3968086`. Production scans found no `jsonb_populate_record`, whole-row `to_jsonb`, wildcard incident select/insert, or codec-level source-port import; working and staged repository diff checks passed. | `DONE`; public HTTP/OpenAPI, database schema, Recovery generations, and valid bundle-v3 bytes remain compatible. Inputs formerly accepted through defaulting or table inference now fail intentionally. Initial generation attempts at `.cartulary/test-results/20260828T154350Z-p3468092` and `.cartulary/test-results/20260828T154453Z-p3471747` exposed authored row sorting/length defects; the first codec run at `.cartulary/test-results/20260828T154645Z-p3482616` exposed digit-only uppercase UUID fixtures; all were corrected. Concurrent focused runs at `.cartulary/test-results/20260828T154802Z-p3492413` and `.cartulary/test-results/20260828T154802Z-p3492420` raced on the shared service-image warm stamp and passed serially. The first final boundary run at `.cartulary/test-results/20260828T155621Z-p3796537` correctly rejected codec-level owner-port coupling; transport translation was isolated in the designated source-port adapter and the gate passed. No SQL generator input was needed because the fixed statements are private owner implementation; `make generate` was required and run for the authored Harness routing changes. Residual risk is limited to recurrence residue audited in `I3-RS-06`. Rollback is the strict codec/adapter, strict duplicate reporting, Incident Bundles actor-path exception, authored routing plus generated topology index, tests, and this checkpoint as one atomic unit. `I3-RS-06` is the sole successor. |
