# Phase 8 Coverage Ledger

This ledger is generated from `tools/phase8_test_map.json`. Update the manifest row metadata first, then regenerate this file.

- Scope: Links, tags, saved views, sorting, filtering, grouping, startup selection, and projection-backed query semantics.
- Normative owners: Core 01 §3.3.4; Core 01 §3.3.5.2; Core 01 §7.4; Core 01 §8; Core 02 §11–§12; Core 03 §2 and §14; Core 04 §2.
- Authority: `tools/phase8_test_map.json` is the enforced Phase 8 traceability source. This ledger is a rendered companion and does not control the mechanical row inventory.
- Phase 8 is active and rows must remain backed by direct evidence in their declared execution layer before Phase 8 exit.
- Every authoritative Phase 8 row has exactly one declared execution dependency from the approved evidence layers: backend_unit, backend_store, backend_integration, frontend_unit, or browser_functional.
- Phase 7 handoff boundaries are preserved: typed tag rollback, AC-184, and AC-185 are Phase 8-owned; merge-specific rollback and Phase 9 keyboard or clipboard behavior remain non-claims.
- Existing Phase 3 through Phase 7 tests may support substrate confidence but are forbidden from claiming Phase 8 row identifiers.

## Authoritative Execution

- `backend-unit` carries authoritative pure backend Phase 8 rows for saved-view scope vocabulary, query contract validation, grouping semantics, normalization, and view-schema discovery.
- `backend-store` carries authoritative service-backed Phase 8 unit rows for links and tags, saved-view create or patch behavior, workbook preferences, and row wire-family semantics.
- `backend-integration` carries authoritative service-backed Phase 8 integration rows for saved-view lifecycle, startup fallback, link and tag atomic consequences, and live-authorized cursor pagination.
- `frontend-unit` carries authoritative `U-8-GRID-01` grid control evidence through Vitest.
- `browser-e2e-webserver-backed` carries authoritative `E-8-*` browser-functional rows.

## Support-Only Execution

- `internal/modules/workbook/phase8_row_wire_test.go` runs through `backend-integration-support` with `TestSupportPhase8Integration_` and is forbidden from claiming `I-8-*` identifiers.
- Phase 8 workbook backend integration support covers exact-token full_text, strict prefix, and null-last ordering readiness for the browser-owned E-8-04 row.
- Phase 3 through Phase 7 workbook, projection, mutation, collaboration, evidence, history, rollback, visual, and browser tests remain support-only for Phase 8.
- Support-only carryover files are listed in forbidden_id_files so they cannot accidentally claim Phase 8 IDs.

## Unit

| Row | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- |
| `U-8-01` | `internal/modules/links/phase8_links_tags_test.go::TestPhase8_TypedLinksAndTags_U_8_01` | `backend_store` | Typed links and tags persist as structured rows and use the closed base relationship vocabulary. | Saved-view, startup, query, cursor, and browser workflows are covered by separate Phase 8 rows. |
| `U-8-02` | `internal/modules/savedviews/phase8_savedviews_test.go::TestPhase8_SavedViewCreateDefaults_U_8_02`, `TestPhase8_SavedViewOpenAPICreateInputIsLenient_U_8_02` | `backend_store` | Ordinary saved-view create defaults omitted scope to private, rejects ordinary system create, persists exactly one view_schema_id, and documents lenient create input with canonical resource output in OpenAPI. | Patch, duplicate, delete, startup selection, and browser workflow behavior are covered by separate rows. |
| `U-8-03` | `internal/modules/savedviews/phase8_savedviews_test.go::TestPhase8_SavedViewScopeVocabulary_U_8_03` | `backend_unit` | Saved-view scope vocabulary is exactly private, shared, and system; obsolete team is rejected. | Persistence and authorization consequences are covered by service-backed rows. |
| `U-8-04` | `internal/modules/savedviews/phase8_savedviews_test.go::TestPhase8_SavedViewPatchContract_U_8_04` | `backend_store` | Ordinary saved-view patch enforces the mutable-field contract, rejects immutable mutation, and preserves version and timestamp on no-op patches. | Create defaults, lifecycle integration, and browser workflows are covered by separate rows. |
| `U-8-05` | `internal/modules/incidents/phase8_workbook_startup_test.go::TestPhase8_WorkbookPreferencePointers_U_8_05` | `backend_store` | Workbook preference routes keep home_sheet_ref and default_sheet_ref separate, preserve no-op timestamps, enforce default-pointer admin policy, and clear a hidden saved-view home pointer before startup fallback continues. | Saved-view lifecycle persistence and browser-visible startup behavior are covered by separate rows. |
| `U-8-06` | `internal/platform/viewquery/phase8_query_contract_test.go::TestPhase8_StableFieldKeysOnly_U_8_06` | `backend_unit` | View-query filters, sort, and group_by accept stable field_key values only and reject invalid operators, argument shapes, duplicate keys, and invalid grouping keys. | Persistence, query execution, cursor pagination, and browser controls are covered by separate rows. |
| `U-8-07` | `internal/platform/viewquery/phase8_query_contract_test.go::TestPhase8_TimelineGroupingWhitelist_U_8_07` | `backend_unit` | Timeline grouping permits only the declared whitelist and never serializes group headers as writable rows. | Browser rendering of grouping and full query execution are covered by separate rows. |
| `U-8-08` | `internal/platform/viewquery/phase8_query_contract_test.go::TestPhase8_QueryNormalizationMeta_U_8_08` | `backend_unit` | Sort and filter ceilings, canonical normalization, meta.query, default-tail expansion, and grouping omission semantics follow the Phase 8 query contract. | Saved-view persistence and browser workflows are covered by separate rows. |
| `U-8-09` | `internal/modules/viewschemas/phase8_viewschema_test.go::TestPhase8_ViewSchemaDiscovery_U_8_09` | `backend_unit` | View-schema discovery exposes exact sorting and grouping contracts, null-last ordering, and no client-sortable record_id. | Runtime query validation and browser control payloads are covered by separate rows. |
| `U-8-10` | `internal/modules/workbook/phase8_row_wire_test.go::TestPhase8_RowWireFamilies_I_8_04_RowWireContract` | `backend_integration` | Full-row and sparse-patch wire families preserve hidden writable fields, authoritative nulls, and canonical changed_field_keys plus affected_views metadata. | Cursor continuation and browser workflows are covered by separate rows. |
| `U-8-GRID-01` | `apps/web/src/WorkbookShell.phase8.query.test.tsx::Phase 8 U-8-GRID-01 emits stable schema keys for sort, filter, and group query controls` | `frontend_unit` | Grid sort, filter, and group controls send only stable contract keys and never labels, vendor indexes, projection-table names, or storage-table names. | Server query validation and browser-functional workflows are covered by separate rows. |

## Integration

| Row | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- |
| `I-8-01` | `internal/modules/savedviews/phase8_savedviews_test.go::TestPhase8_SavedViewLifecyclePersistence_I_8_01` | `backend_integration` | Saved-view create, update, duplicate, and delete persist normalized state and authorization consequences in the database. | Browser saved-view workflows are covered by E-8-01. |
| `I-8-02` | `internal/modules/incidents/phase8_workbook_startup_test.go::TestPhase8_WorkbookStartupFallback_I_8_02`, `TestPhase8_WorkbookStartupBaseSurfaceDoesNotRequireSavedView_I_8_02` | `backend_integration` | Startup selection fallback behaves correctly across valid, missing, hidden, and invalid saved-view refs. | Browser-visible startup behavior is covered by E-8-02. |
| `I-8-03` | `internal/modules/links/phase8_links_tags_test.go::TestPhase8_LinkTagProjectionHistoryQuery_I_8_03` | `backend_integration` | Link and tag mutations atomically update projections, history, and view-query results. | Saved-view lifecycle, startup fallback, cursor continuation, and browser workflows are covered by separate rows. |
| `I-8-04` | `internal/modules/workbook/phase8_row_wire_test.go::TestPhase8_LiveAuthorizedCursorPagination_I_8_04`, `TestPhase8_CursorContinuationRechecksAuthorization_I_8_04`, `TestPhase8_CursorContinuationRechecksMembership_I_8_04` | `backend_integration` | Cursor pagination re-derives live authorization, rejects tampering, returns fetch-time payloads, and lets fresh queries evaluate live state. | Browser exact search behavior is covered by E-8-04. |

## Browser E2E

| Row | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- |
| `E-8-01` | `apps/web/e2e/phase8.workbook.spec.ts::E-8-01 saved-view route foundation persists canonical state while browser lifecycle affordances remain pending` | `browser_functional` | The browser-authenticated public saved-view list/create route foundation exists and persists canonical state; full browser lifecycle affordances remain pending. | Patch, duplicate, delete, startup selection, and in-product saved-view browser controls remain outside Sprint 2. |
| `E-8-02` | `apps/web/e2e/phase8.workbook.spec.ts::E-8-02 workbook startup falls back to Timeline for an unsupported explicit sheet` | `browser_functional` | Browser startup honors explicit sheet, home, default, and Timeline fallback while clearing invalid pointers. | Backend startup selection is covered by U-8-05 and I-8-02. |
| `E-8-03` | `apps/web/e2e/phase8.workbook.spec.ts::E-8-03 browser Timeline sort, filter, and group controls submit stable query keys` | `browser_functional` | Browser Timeline sort, filter, and group reaches a stable viewport and deterministic grouping with no writable header rows. | Backend query validation and view-schema discovery are covered by U-8-06 through U-8-09. |
| `E-8-04` | `apps/web/e2e/phase8.workbook.spec.ts::E-8-04 browser Notes full_text and prefix queries remain exact` | `browser_functional` | Browser exact-token full_text and strict prefix behavior do not degrade to fuzzy or relevance-ranked search. | Backend cursor continuation is covered by I-8-04. |

## Shared Harness Coverage

| Harness | Phase 8 evidence |
| --- | --- |
| Active manifest ownership | `tools/phase8_test_map.json` records every authoritative guide row and marks previous phase carryover as non-owning support. |
| Generated ledger | `docs/testing/phase8_coverage_ledger.md` is generated from this manifest and must not be hand-edited. |
| Schedule boundary | Phase 8 is active only after selectable row symbols and titles exist, so generated schedules can include coherent Phase 8 work. |

## Support-Only Evidence

- Existing Phase 3 through Phase 7 backend tests under timeline, entities, evidence, workbook, collaboration, revisions, and WebSocket packages remain support-only for Phase 8.
- Existing Phase 3 through Phase 7 frontend and browser specs remain previous-phase evidence and are forbidden from claiming Phase 8 IDs.
- Incomplete Phase 8 rows must remain visible as blockers and cannot be counted as exit evidence.
