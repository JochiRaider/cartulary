# Phase 9 Coverage Ledger

This ledger is generated from `tools/phase9_test_map.json`. Update the manifest row metadata first, then regenerate this file.

- Scope: keyboard contract, clipboard and bulk-edit behaviors, Notes, Indicators, Parties, Assessments, Task Requests, Decisions, coordination surfaces, optional standardized surfaces when exposed, and remaining analyst-work workbook surfaces.
- Normative owners: Core 01 §7.4 and §19; Core 02 §10 and §19; Core 03 §2, §11, §13, and §16–§20; Core 04 §2.
- Authority: `tools/phase9_test_map.json` is the enforced Phase 9 traceability source. This ledger is a rendered companion and does not control the mechanical row inventory.
- Sprint 0 activates Phase 9 as a selectable harness phase using explicit blocker sentinels; Sprint 1 replaces keyboard dispatch, grid-anchor, and keyboard browser rows with direct evidence.
- Every authoritative Phase 9 row has exactly one declared execution dependency from the approved evidence layers: backend_unit, backend_store, backend_integration, frontend_unit, or browser_functional.
- Sprint 1 row-level evidence can be green while aggregate Phase 9 targets remain red from later-row blocker sentinels; that status does not complete broader Phase 9.
- Remaining Sprint 0 sentinel tests intentionally fail when selected and must not be treated as implemented behavior.
- Optional standardized Findings, Investigative Queries, and Forensic Keywords rows remain blocker sentinels until later work records direct coverage for exposed surfaces or an owner-cited N/A interpretation.

## Authoritative Execution

- `backend-unit` selects Phase 9 blocker sentinels for pure backend clipboard, timestamp, direct-reference, and registry rows until real implementation evidence replaces them.
- `backend-store` selects Phase 9 blocker sentinels for Notes, Indicators, Parties, Assessments, Tasks and Decisions, coordination, optional standardized surfaces, and manual relationship collection rows until real implementation evidence replaces them.
- `backend-integration` selects Phase 9 blocker sentinels for clipboard persistence, workbook projection, and party-link helper rows until real implementation evidence replaces them.
- `frontend-unit` selects direct Sprint 1 evidence for keyboard command and grid-anchor rows while later frontend Phase 9 rows remain blocker sentinels until real implementation evidence replaces them.
- `browser-e2e-webserver-backed` selects direct Sprint 1 keyboard and shared-grid browser evidence while later browser-functional workbook workflows remain blocker sentinels until real implementation evidence replaces them.

## Support-Only Execution

- Phase 4 workbook generic, assessment, Party, Indicator, and coordination smoke evidence remains support-only substrate and cannot claim Phase 9 completion or satisfy Phase 9 row ownership by itself.
- Phase 6 grid, collaboration, pending-queue, and conflict marker evidence remains support-only substrate and cannot claim Phase 9 keyboard, clipboard, or shared-grid completion.
- Phase 8 query, saved-view, grouping, sorting, filtering, and startup evidence remains support-only substrate and cannot claim Phase 9 workbook behavior.
- Support-only carryover files are listed in forbidden_id_files so they cannot accidentally claim Phase 9 identifiers.

## Unit

| Row | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- |
| `U-9-01` | `apps/web/src/workbookKeyboard.test.ts::Phase 9 U-9-01 maps required workbook keys without hidden paste macro behavior`, `Phase 9 U-9-01 fails closed when optional shortcut actions are unavailable` | `frontend_unit` | Keyboard command mapping covers Arrow keys, Enter, Shift+Enter, Tab, Ctrl+V, Ctrl+K, Space, Alt+H, and Esc without hidden paste macro behavior. | Ctrl+V is classified as paste intent only; Sprint 2 owns paste and bulk-edit semantics. |
| `U-9-02` | `internal/modules/workbook/phase9_sprint0_sentinel_test.go::TestPhase9_Sprint0Blocker_U_9_02` | `backend_unit` | Sprint 0 blocker sentinel for paste and bulk ingest grouping; this is not behavior completion evidence. | Implementing clipboard and bulk-ingest behavior is later Phase 9 work. |
| `U-9-03` | `internal/modules/workbook/phase9_sprint0_sentinel_test.go::TestPhase9_Sprint0Blocker_U_9_03` | `backend_store` | Sprint 0 blocker sentinel for artifact-backed Notes rows; this is not behavior completion evidence. | Implementing Notes behavior is later Phase 9 work. |
| `U-9-04` | `internal/modules/entities/phase9_sprint0_sentinel_test.go::TestPhase9_Sprint0Blocker_U_9_04` | `backend_store` | Sprint 0 blocker sentinel for canonical Indicator rows; this is not behavior completion evidence. | Implementing Indicator behavior is later Phase 9 work. |
| `U-9-05` | `internal/modules/entities/phase9_sprint0_sentinel_test.go::TestPhase9_Sprint0Blocker_U_9_05` | `backend_store` | Sprint 0 blocker sentinel for Party exact-match reuse and raw text preservation; this is not behavior completion evidence. | Implementing Party text-plus-link behavior is later Phase 9 work. |
| `U-9-06` | `internal/modules/assessments/phase9_sprint0_sentinel_test.go::TestPhase9_Sprint0Blocker_U_9_06` | `backend_store` | Sprint 0 blocker sentinel for append-only compromise assessments; this is not behavior completion evidence. | Implementing assessment behavior is later Phase 9 work. |
| `U-9-07` | `internal/modules/workbook/phase9_sprint0_sentinel_test.go::TestPhase9_Sprint0Blocker_U_9_07` | `backend_store` | Sprint 0 blocker sentinel for Task Requests and Decisions; this is not behavior completion evidence. | Implementing Task Request and Decision behavior is later Phase 9 work. |
| `U-9-08` | `internal/modules/workbook/phase9_sprint0_sentinel_test.go::TestPhase9_Sprint0Blocker_U_9_08` | `backend_store` | Sprint 0 blocker sentinel for coordination surfaces; this is not behavior completion evidence. | Implementing coordination surfaces is later Phase 9 work. |
| `U-9-09` | `internal/modules/workbook/phase9_sprint0_sentinel_test.go::TestPhase9_Sprint0Blocker_U_9_09` | `backend_store` | Sprint 0 blocker sentinel for optional standardized surfaces when exposed; this is not behavior completion evidence and does not pre-claim optional coverage. | Later Phase 9 work must replace this with direct optional-surface evidence or owner-cited N/A coverage. |
| `U-9-10` | `internal/modules/workbook/phase9_sprint0_sentinel_test.go::TestPhase9_Sprint0Blocker_U_9_10` | `backend_unit` | Sprint 0 blocker sentinel for timestamp_instant_v1 behavior; this is not behavior completion evidence. | Implementing timestamp validation and client-local draft preservation is later Phase 9 work. |
| `U-9-11` | `internal/modules/workbook/phase9_sprint0_sentinel_test.go::TestPhase9_Sprint0Blocker_U_9_11` | `backend_unit` | Sprint 0 blocker sentinel for direct-reference field behavior; this is not behavior completion evidence. | Implementing direct-reference validation and text-plus-ref independence is later Phase 9 work. |
| `U-9-12` | `internal/modules/workbook/phase9_sprint0_sentinel_test.go::TestPhase9_Sprint0Blocker_U_9_12` | `backend_store` | Sprint 0 blocker sentinel for manual relationship collection confidence behavior; this is not behavior completion evidence. | Implementing manual relationship collection behavior is later Phase 9 work. |
| `U-9-13` | `internal/modules/workbook/phase9_sprint0_sentinel_test.go::TestPhase9_Sprint0Blocker_U_9_13` | `backend_unit` | Sprint 0 blocker sentinel for artifact-backed variant registry closure; this is not behavior completion evidence. | Implementing Notes and coordination registry closure is later Phase 9 work. |
| `U-9-GRID-01` | `apps/web/src/GridAdapter.phase9.anchor.test.ts::Phase 9 U-9-GRID-01 resolves vendor coordinates to stable record_id and field_key anchors`, `Phase 9 U-9-GRID-01 clears invalid row, field, group row, and recordless draft targets`, `Phase 9 U-9-GRID-01 resolves Arrow, Tab, Enter, and Shift+Enter navigation through adapter anchors`, `Phase 9 U-9-GRID-01 keeps vendor selection separate until translated by the adapter contract` | `frontend_unit` | The grid adapter translates vendor row, column, selection, and keyboard navigation coordinates into stable Cartulary anchors by record_id and field_key, clears invalid or presentation-only targets, and keeps vendor selection separate until translated through the adapter contract. | Sorted or filtered paste anchor translation remains owned by I-9-GRID-01 in Sprint 2. |

## Integration

| Row | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- |
| `I-9-01` | `internal/modules/workbook/phase9_sprint0_sentinel_test.go::TestPhase9_Sprint0Blocker_I_9_01` | `backend_integration` | Sprint 0 blocker sentinel for multi-row clipboard paste persistence; this is not behavior completion evidence. | Implementing ordered clipboard mutations and origin distinctions is later Phase 9 work. |
| `I-9-02` | `internal/modules/workbook/phase9_sprint0_sentinel_test.go::TestPhase9_Sprint0Blocker_I_9_02` | `backend_integration` | Sprint 0 blocker sentinel for required surface persistence and projection queries; this is not behavior completion evidence. | Implementing required surface persistence and query behavior is later Phase 9 work. |
| `I-9-03` | `internal/modules/workbook/phase9_sprint0_sentinel_test.go::TestPhase9_Sprint0Blocker_I_9_03` | `backend_integration` | Sprint 0 blocker sentinel for party-link helper field behavior; this is not behavior completion evidence. | Implementing party-link helper behavior is later Phase 9 work. |
| `I-9-GRID-01` | `apps/web/src/WorkbookShell.phase9.sentinel.test.tsx::Phase 9 I-9-GRID-01 Sprint 0 blocker sentinel` | `frontend_unit` | Sprint 0 blocker sentinel for sorted or filtered paste anchor translation; this is not behavior completion evidence. | Implementing visual-order paste translation is later Phase 9 work. |

## Browser E2E

| Row | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- |
| `E-9-01` | `apps/web/e2e/phase9.keyboard.spec.ts::Phase 9 E-9-01 keyboard shortcuts keep workbook grid anchors without module switching` | `browser_functional` | Required Sprint 1 keyboard shortcuts operate on the live Timeline workbook grid without route or module switching: Arrow keys, Enter, Shift+Enter, Tab, Ctrl+V paste intent, Ctrl+K same-surface quick-link or resolve, Space same-surface linked-evidence preview state, Alt+H row history, and Esc inspector or history close. | Ctrl+V remains paste-intent only and does not claim Sprint 2 paste behavior. |
| `E-9-02` | `apps/web/e2e/phase9.sentinel.spec.ts::Phase 9 E-9-02 Sprint 0 blocker sentinel` | `browser_functional` | Sprint 0 blocker sentinel for representative Timeline paste workflow; this is not behavior completion evidence. | Implementing browser paste workflow behavior is later Phase 9 work. |
| `E-9-03` | `apps/web/e2e/phase9.sentinel.spec.ts::Phase 9 E-9-03 Sprint 0 blocker sentinel` | `browser_functional` | Sprint 0 blocker sentinel for Notes tab browser workflow; this is not behavior completion evidence. | Implementing Notes browser workflow behavior is later Phase 9 work. |
| `E-9-04` | `apps/web/e2e/phase9.sentinel.spec.ts::Phase 9 E-9-04 Sprint 0 blocker sentinel` | `browser_functional` | Sprint 0 blocker sentinel for Party create/link browser workflow; this is not behavior completion evidence. | Implementing Party browser workflow behavior is later Phase 9 work. |
| `E-9-05` | `apps/web/e2e/phase9.sentinel.spec.ts::Phase 9 E-9-05 Sprint 0 blocker sentinel` | `browser_functional` | Sprint 0 blocker sentinel for assessment history and filters browser workflow; this is not behavior completion evidence. | Implementing assessment browser workflow behavior is later Phase 9 work. |
| `E-9-06` | `apps/web/e2e/phase9.sentinel.spec.ts::Phase 9 E-9-06 Sprint 0 blocker sentinel` | `browser_functional` | Sprint 0 blocker sentinel for Task, Decision, and coordination browser workflows; this is not behavior completion evidence. | Implementing Task, Decision, and coordination browser workflow behavior is later Phase 9 work. |
| `E-9-07` | `apps/web/e2e/phase9.sentinel.spec.ts::Phase 9 E-9-07 Sprint 0 blocker sentinel` | `browser_functional` | Sprint 0 blocker sentinel for optional standardized surfaces when exposed; this is not behavior completion evidence and does not pre-claim optional coverage. | Later Phase 9 work must replace this with direct optional-surface browser evidence or owner-cited N/A coverage. |
| `E-9-08` | `apps/web/e2e/phase9.sentinel.spec.ts::Phase 9 E-9-08 Sprint 0 blocker sentinel` | `browser_functional` | Sprint 0 blocker sentinel for workbook surface registry browser exposure; this is not behavior completion evidence. | Implementing registry browser exposure behavior is later Phase 9 work. |
| `E-9-GRID-01` | `apps/web/e2e/phase9.keyboard.spec.ts::Phase 9 E-9-GRID-01 shared grid keyboard anchors stay stable across workbook cells` | `browser_functional` | Shared grid keyboard anchor behavior remains stable across Timeline cells, renderer-family surfaces, the Assessments system view, and the required Task Requests generic system view using stable record_id and field_key anchors. | Later Phase 9 paste, bulk-edit, presence, save-state, conflict-marker, and optional standardized surface behavior remain blocked by their own browser-functional sentinels. |

## Shared Harness Coverage

| Harness | Phase 9 evidence |
| --- | --- |
| Active manifest ownership | `tools/phase9_test_map.json` records every authoritative guide row and marks earlier phase carryover as non-owning support. |
| Blocker sentinel boundary | Sprint 0 sentinels are selectable only to prevent false completion claims; each sentinel must fail until replaced by real Phase 9 behavior evidence. |
| Generated ledger | `docs/testing/phase9_coverage_ledger.md` is generated from this manifest and must not be hand-edited. |
| Schedule boundary | Phase 9 is active because coherent selectable row symbols and titles exist, but any remaining sentinel blocks Phase 9 completion claims. |

## Support-Only Evidence

- Existing Phase 4 backend, frontend, and browser evidence for Parties, Indicators, Assessments, coordination, and generic workbook routes remains previous-phase support only.
- Existing Phase 6 backend, frontend, and browser evidence for collaboration, conflict, presence, pending queue, and grid marker behavior remains previous-phase support only.
- Existing Phase 8 backend, frontend, and browser evidence for query controls, saved views, sorting, filtering, grouping, startup selection, and projection-backed query semantics remains previous-phase support only.
- Phase 9 Sprint 0 sentinels are intentional blockers and are not implemented behavior.
