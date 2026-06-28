# TimelineWorkbook Refactor Tracker

## 1. Session Header

| Field | Value |
| --- | --- |
| Date | 2026-06-28 |
| Branch | `main` |
| Commit | `d318fe30047de05a5c1dbad74910fce7726f5d96` |
| Dirty tree state | Tracked tree clean at inspection time: `git status --short --branch` reported `## main...origin/main`. |
| Target file | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx` |
| Target module/package/seam | Frontend Timeline workbook surface inside `/apps/web` |
| Planning mode | Planning-only. Do not implement the refactor in this artifact. |
| Framework path used | `docs/handoffs/cartulary_modular_refactor_planning_framework.md` |
| Default framework source limit | `temp/current.md` exists but is empty, so it is not usable framework evidence. |
| Primary owner references inspected | `docs/domain.md`, Core 00 through Core 04 targeted sections, Core 01 WebSocket sections, `docs/testing-harness-nlspec.md`, `docs/design.md`, Timeline view-schema contract. |
| Source limits and unseen files | Timeline adjacent files were inventoried and declaration-scanned unless listed as directly inspected or named in the slice entry inspection manifest. No production validation targets were executed during planning. |

This artifact treats the reusable framework as planning structure only. It is not proof of current repository state unless confirmed in the repository scan below.

## 2. Current-State Repository Scan

### Inspected Files Table

| Path or command | Method | In-scope evidence | Source limit |
| --- | --- | --- | --- |
| `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx` | Direct read and symbol searches | Primary file exists and is 3001 lines. Imports, exports, hook use, route identity, paste payload, and row-version handling were inspected. | Full behavior was not line-by-line proven; complex blocks need local inspection before editing. |
| `docs/handoffs/cartulary_modular_refactor_planning_framework.md` | Direct read | Source for top-level tracker, WF map, workstreams, handoff template, and RF-AC rows. | Framework is planning structure only. |
| `temp/current.md` | File-size check | Present but 0 bytes. | Unusable as framework content. |
| `docs/domain.md` | Direct and targeted read | Timeline vocabulary, stable workbook identity terms, saved-view boundary, evidence and mention concept boundaries. | Vocabulary reference only, not product conformance proof. |
| `docs/spec/00_document_set_status_and_precedence.md` | Targeted read | Core 00 through Core 04 are current implementation-conformance owners; Core 05 is only for claim-bearing publication evidence. | Targeted sections only. |
| `docs/spec/01_architecture_storage_and_view_contracts.md` | Targeted read | Public API identity, route roots, view-row envelopes, query route shape, paste identity, generated contract boundaries. | Targeted sections only. |
| `docs/spec/02_domain_model_schema_and_history.md` | Targeted read | `timeline_event`, record envelope, saved-view query constraints, mention/evidence/link boundaries. | Targeted sections only. |
| `docs/spec/03_workbook_interaction_collaboration_and_workflows.md` | Targeted read | Grid-first behavior, save/conflict rules, inspector lifecycle, sorting/filtering/grouping, collaboration, Timeline route usage. | Targeted sections only. |
| `docs/spec/04_security_deployment_and_conformance.md` | Targeted read and owner mapping | Security, authorization, session, incident-role, evidence-access, inspector-action, and WebSocket rules mapped for future slices. | No remaining source limit for planner use; if implementation inspection finds contradiction, split a spec-aligned remediation task before refactoring. |
| `docs/spec/05_claim_publication_and_benchmark_reproducibility.md` | Targeted read | Visual/a11y/timed evidence must not be promoted into product conformance unless the Core 05 publication boundary is satisfied. | Not otherwise relevant to this refactor. |
| `docs/testing-harness-nlspec.md` | Direct and targeted read | Make targets are canonical invocation; harness failure classes; visual/a11y evidence boundaries; generated roots must not be hand-edited. | Harness mechanics only. |
| `docs/design.md` | Direct and targeted read | Timeline design direction: dense workbook surface, grid-primary editing, inspector adjacent, evidence/presence semantic states. | Design direction only, not product conformance. |
| `contracts/view-schemas/cartulary.view.timeline.v2.json` | Direct read and JSON parse | Timeline `view_schema_id`, fields, sort/filter/grouping fields, inline-create policy, inspector config, feature groups. | Contract-derived input; do not hand-edit generated consumers. |
| `packages/view-contracts/src/index.ts` | Targeted search | Parses `inspector_config`; declares expected Timeline feature group registry. | Not fully read. |
| `packages/ui-contracts/src/index.ts` | Import/user scan | Selector and test-id builders are imported by the target and must stay stable. | Not fully read. |
| `tools/generated_artifact_policy.json` | Direct read | Generated roots and generated files identified. | None for this artifact. |
| `tools/frontend_import_boundaries.json` | Direct read and boundary search | Direct `react-data-grid` imports restricted to `packages/grid-adapter/src/**`. | None for boundary planning. |
| `apps/web/src/workbook/timeline/hooks/useTimelineGridAnchorController.ts` | Direct read | Existing grid-anchor and paste-target seam uses Cartulary anchors, not vendor coordinates. | Read for seam planning, not edited. |
| `apps/web/src/workbook/timeline/components/TimelineWorkbookRenderers.tsx` | Direct read | Existing renderer seam owns Timeline grid columns/renderers and selector/test-id use. | Read for seam planning, not edited. |
| `apps/web/src/workbook/WorkbookShell.tsx` | Targeted read | Production caller renders `TimelineWorkbook` and controls query state, saved-view selector, sheet ref, density, entities, role, and refresh inputs. | Targeted lines only. |
| `tools/phase3_test_map.json` through `tools/phase9_test_map.json` | Targeted searches | Phase maps identify existing harness accounting for grid, save state, query, evidence, collaboration, history, inspector, keyboard, and anchor evidence. | Phase identity is not runtime structure. |
| `make help-all` and `make explain-target ...` | Command discovery only | Public Make targets and expected artifacts discovered for validation planning. | No verification target was run. |

### Direct Imports And Imported Symbols

`TimelineWorkbook.tsx` imports these direct surfaces:

| Source | Imported symbols or purpose | Contract note |
| --- | --- | --- |
| `@cartulary/grid-adapter` | `GridCellAnchor`, `GridColumn`, `GridDensity`, `GridRow` | Allowed package seam. Do not import `react-data-grid` directly in `/apps/web`. |
| `@cartulary/ui-contracts` | `dataTestIdSelector`, `draftCellTestId`, `genericCreateFieldTestId`, `genericCreateSubmitTestId`, `gridGroupRowTestId`, `gridRowGutterTestId`, `gridScrollportSelector`, `rowCellTestId`, `timelineInspectorSectionTestId`, `timelineMutationSubstrateReadyTestId`, `WorkbookSurface` | Stable selector/test-id contract. |
| `@cartulary/view-contracts` | `requireViewContract`, `ViewContract` | Runtime contract adapter for view-schema inputs. |
| `react`, `react-dom` | Hooks, event types, `flushSync` | Component/runtime glue. |
| Services | `apiPath`, `fetchJSON`, `readEnvelope` | API path and envelope handling. |
| Shared workbook components | `GenericMutationControl`, `WorkbookGridControls`, `WorkbookSheetToolbar`, `WorkbookStatusStrip`, `WorkbookSurfaceFrame` | Workbook shell/control composition. |
| Shared workbook models | Evidence count display model, inspector config, query state, reference options, startup sheet ref, surface registry IDs | Query/view/schema/inspector contract glue. |
| Shared workbook utils | Viewport continuity, focus anchor, keyboard command mapping, save-state derivation, pending queue conflict anchors, presence sheet matching | Hot-path behavior; move only behind characterized seams. |
| Timeline hooks | `useTimelineCommittedRows`, `useTimelineConflictResolverCoordinator`, `useTimelineConflicts`, `useTimelineCreateRelatedWorkflow`, `useTimelineEvidenceActions`, `useTimelineEvidenceAttach`, `useTimelineGridAnchorController`, `useTimelineGridInteractions`, `useTimelineHistoryActions`, `useTimelineHistoryState`, `useTimelineInspectorEscape`, `useTimelineInspectorLifecycle`, `useTimelineInspectorRowInteractions`, `useTimelineInspectorSelection`, `useTimelineLiveUpdateController`, `useTimelineLiveUpdates`, `useTimelineMentionActions`, `useTimelineMentions`, `useTimelineMutationCommands`, `useTimelinePendingReplayController`, `useTimelinePendingSaves`, `useTimelineRows`, `useTimelineRowsLoader`, `useTimelineWorkbookRuntime` | Many seams already exist. New extraction should fit these patterns. |
| Timeline models/services | Timeline rows, viewport continuity, mention chips, workbook timeline model, mutation envelope, collaboration payload, socket lifecycle | Row identity, conflict, mutation, and live-update contracts. |
| Timeline components/styles | Cell editors, conflict resolver, evidence panel, grid surface, history panel, presence row gutter, row context menu, inspector, notices, renderers, style constants | Presentation and inspector glue. No redesign planned. |

### Exported Symbols And Known Callers

| Export | Current known callers | Stability note |
| --- | --- | --- |
| `SaveState` | Runtime state shape in target and tests. | Must remain `"Syncing"`, `"Saved"`, or `"Conflict"` unless owner docs change. |
| `IncidentRole` | Target props and tests. | Role values must not drift. |
| `TimelineWorkbookProps` | `apps/web/src/testing/timelineWorkbookRenderTestSupport.tsx`, tests via `ComponentProps`, production caller props. | Public component prop shape; do not narrow or rename without a separate owner decision. |
| `TimelineWorkbook` | Production caller `apps/web/src/workbook/WorkbookShell.tsx`; direct tests in app/workbook suites; render helper in `apps/web/src/testing/timelineWorkbookRenderTestSupport.tsx`. | Primary target component; keep observable behavior unchanged. |

Known direct production caller:

- `apps/web/src/workbook/WorkbookShell.tsx`

Known test/helper callers include:

- `apps/web/src/testing/timelineWorkbookRenderTestSupport.tsx`
- `apps/web/src/testing/timelineWorkbookTestSupport.ts`
- `apps/web/src/app/App.test.tsx`
- `apps/web/src/app/fontRoles.test.tsx`
- `apps/web/src/workbook/WorkbookShell.phase3.*.test.tsx`
- `apps/web/src/workbook/WorkbookShell.phase4.*.test.tsx`
- `apps/web/src/workbook/WorkbookShell.phase5*.test.*`
- `apps/web/src/workbook/WorkbookShell.phase6.test.tsx`
- `apps/web/src/workbook/WorkbookShell.phase7.test.tsx`
- `apps/web/src/workbook/WorkbookShell.phase8.query.test.tsx`
- `apps/web/src/workbook/WorkbookShell.phase9.*.test.tsx`

### Adjacent Hooks, Components, And Utilities

The Timeline directory currently contains 51 files. The primary file and two adjacent seam files were directly inspected; the rest were inventoried and symbol-scanned.

Directly inspected adjacent files:

- `apps/web/src/workbook/timeline/hooks/useTimelineGridAnchorController.ts`
- `apps/web/src/workbook/timeline/components/TimelineWorkbookRenderers.tsx`
- `apps/web/src/workbook/WorkbookShell.tsx`

Inventory-only adjacent Timeline components:

- `TimelineCellEditors.tsx`
- `TimelineConflictResolver.tsx`
- `TimelineEvidencePanel.tsx`
- `TimelineGridSurface.tsx`
- `TimelineHistoryPanel.tsx`
- `TimelineMentionsPanel.tsx`
- `TimelinePresenceMarkers.tsx`
- `TimelineRowActions.tsx`
- `TimelineWorkbookGrid.tsx`
- `TimelineWorkbookInspector.tsx`
- `TimelineWorkbookNotices.tsx`
- `TimelineWorkbookStyles.ts`

Inventory-only adjacent Timeline hooks:

- `useTimelineCommittedRows.ts`
- `useTimelineConflictResolverCoordinator.ts`
- `useTimelineConflicts.ts`
- `useTimelineCreateRelatedWorkflow.ts`
- `useTimelineEvidenceActions.ts`
- `useTimelineEvidenceAttach.ts`
- `useTimelineGridInteractions.ts`
- `useTimelineHistoryActions.ts`
- `useTimelineHistoryState.ts`
- `useTimelineInspectorSelection.ts`
- `useTimelineLiveUpdateController.ts`
- `useTimelineLiveUpdates.ts`
- `useTimelineMentionActions.ts`
- `useTimelineMentions.ts`
- `useTimelineMutationCommands.ts`
- `useTimelinePendingReplayController.ts`
- `useTimelinePendingSaves.ts`
- `useTimelineRows.ts`
- `useTimelineRowsLoader.ts`
- `useTimelineWorkbookRuntime.ts`

Inventory-only adjacent Timeline models/services:

- `timelineConflictModel.ts`
- `timelineHistoryModel.ts`
- `timelineRowsModel.ts`
- `timelineViewportContinuityModel.ts`
- `workbookMentionChips.ts`
- `workbookTimelineModel.ts`
- `timelineMutationRequests.ts`
- `workbookCollaborationMessages.ts`
- `workbookSocketLifecycle.ts`

Before editing any inventory-only adjacent file, use the slice entry inspection manifest below and record the directly inspected files in the future handoff record.

### Slice Entry Inspection Manifest

| Slice | Required pre-edit file inspection | Contracts to reread | Default evidence to map | Stop condition |
| --- | --- | --- | --- | --- |
| `TW-SL-01` save state and pending presentation | `TimelineWorkbook.tsx`, `useTimelinePendingSaves.ts`, `workbookPendingQueue.ts`, `TimelineWorkbookNotices.tsx`, `WorkbookStatusStrip` caller surface | Core 03 save-state and pending-queue rules; Core 04 session/auth failure rules | `WorkbookShell.phase4.saveState.test.tsx`, `WorkbookShell.phase3.autosave.test.tsx`, `WorkbookShell.phase4.actionSequencing.test.tsx`, `workbookPendingQueue.test.ts` | Stop if precedence among pending work, conflicts, overflow, auth pause, refresh block, or beforeunload is not covered by existing behavior tests. |
| `TW-SL-02` viewport and focus continuity | `TimelineWorkbook.tsx`, `useTimelineGridAnchorController.ts`, `useTimelineGridInteractions.ts`, `timelineViewportContinuityModel.ts`, `workbookContinuity.ts`, `workbookKeyboard.ts` | Core 03 grid-first focus, active row, and keyboard behavior; grid-adapter boundary policy | `timelineViewportContinuityModel.test.ts`, `workbookContinuity.test.ts`, `workbookKeyboard.test.ts`, `GridAdapter.phase9.anchor.test.ts`, `WorkbookShell.phase9.sentinel.test.tsx` | Stop if the extraction needs direct grid-vendor coordinates or new selector semantics in `/apps/web`. |
| `TW-SL-03` clipboard paste dispatch | `TimelineWorkbook.tsx`, `useTimelineGridAnchorController.ts`, `workbookClipboard.ts`, `timelineMutationRequests.ts`, `workbookTimelineModel.ts` | Core 01 paste identity and route envelope; Core 03 grid paste/focus behavior | `workbookClipboard.test.ts`, `GridAdapter.phase9.anchor.test.ts`, `WorkbookShell.phase3.payload.test.tsx`, `WorkbookShell.phase9.sentinel.test.tsx`, browser-backed paste rows when route dispatch changes | Stop if route body, target identity, conflict handling, or scalar comma behavior cannot remain spec-aligned. |
| `TW-SL-04` inspector render/action seam | `TimelineWorkbook.tsx`, `TimelineWorkbookInspector.tsx`, `TimelineEvidencePanel.tsx`, `TimelineHistoryPanel.tsx`, `TimelineRowActions.tsx`, `workbookInspectorModel.ts`, Timeline view-schema contract | Core 03 inspector lifecycle and feature groups; Core 04 inspector authorization and no external enrichment | `WorkbookShell.phase9.inspector.test.tsx`, `TimelineEvidencePanel.test.tsx`, `WorkbookShell.phase7.test.tsx`, `WorkbookShell.phase4.support.test.tsx` | Stop if extraction would treat control visibility as authorization, change feature-group keys, rename selectors, or redesign the UI. |
| `TW-SL-05` presence and live updates | `TimelineWorkbook.tsx`, `useTimelineLiveUpdates.ts`, `useTimelineLiveUpdateController.ts`, `workbookCollaborationMessages.ts`, `workbookSocketLifecycle.ts`, `contracts/ws/index.schema.json` | Core 01 WebSocket contract; Core 03 presence/live update rules; Core 04 session, Origin, membership, and revocation rules | `workbookCollaborationMessages.test.ts`, `workbookSocketLifecycle.test.ts`, `WorkbookShell.phase6.test.tsx`, `phase6.collaboration` browser rows when socket semantics change | Stop if wire shape, resume semantics, authorization, or session handling would change without an owner-spec decision. |
| `TW-SL-06` assembler cleanup | `TimelineWorkbook.tsx` plus every helper created by completed prior slices | All contracts touched by prior slices | Per-slice validation evidence plus `frontend-import-boundary-check` | Stop if cleanup starts changing public props, routes, selectors, view-schema IDs, field keys, or phase accounting. |

### Tests And Fixtures Found

Existing Timeline-local tests:

- `apps/web/src/workbook/timeline/components/TimelineEvidencePanel.test.tsx`
- `apps/web/src/workbook/timeline/models/timelineConflictModel.test.ts`
- `apps/web/src/workbook/timeline/models/timelineRowsModel.test.ts`
- `apps/web/src/workbook/timeline/models/timelineViewportContinuityModel.test.ts`
- `apps/web/src/workbook/timeline/models/workbookTimelineModel.test.ts`
- `apps/web/src/workbook/timeline/services/workbookCollaborationMessages.test.ts`
- `apps/web/src/workbook/timeline/services/workbookSocketLifecycle.test.ts`

Existing shared workbook tests relevant to this target:

- `apps/web/src/workbook/utils/workbookClipboard.test.ts`
- `apps/web/src/workbook/utils/workbookContinuity.test.ts`
- `apps/web/src/workbook/utils/workbookKeyboard.test.ts`
- `apps/web/src/workbook/utils/workbookPendingQueue.test.ts`
- `apps/web/src/workbook/utils/GridAdapter.phase9.anchor.test.ts`

Existing WorkbookShell and Timeline integration tests found:

- `apps/web/src/workbook/WorkbookShell.phase3.grid.test.tsx`
- `apps/web/src/workbook/WorkbookShell.phase3.autosave.test.tsx`
- `apps/web/src/workbook/WorkbookShell.phase3.payload.test.tsx`
- `apps/web/src/workbook/WorkbookShell.phase4.actionSequencing.test.tsx`
- `apps/web/src/workbook/WorkbookShell.phase4.saveState.test.tsx`
- `apps/web/src/workbook/WorkbookShell.phase4.support.test.tsx`
- `apps/web/src/workbook/WorkbookShell.phase4.timelineQuery.test.tsx`
- `apps/web/src/workbook/WorkbookShell.phase5.test.tsx`
- `apps/web/src/workbook/WorkbookShell.phase5.mentionChips.test.ts`
- `apps/web/src/workbook/WorkbookShell.phase5.gridProvenance.test.tsx`
- `apps/web/src/workbook/WorkbookShell.phase6.test.tsx`
- `apps/web/src/workbook/WorkbookShell.phase7.test.tsx`
- `apps/web/src/workbook/WorkbookShell.phase8.query.test.tsx`
- `apps/web/src/workbook/WorkbookShell.phase9.inspector.test.tsx`
- `apps/web/src/workbook/WorkbookShell.phase9.sentinel.test.tsx`

Existing browser, visual, and a11y tests found:

- `apps/web/e2e/workbook.a11y.spec.ts`
- `apps/web/e2e/workbook.a11y-preflight.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts`
- `apps/web/e2e/frontend.phase4.timeline-query.spec.ts`
- `apps/web/e2e/phase8.workbook.spec.ts` and related phase browser specs found by inventory.

Fixture and helper surfaces found:

- `apps/web/src/testing/timelineWorkbookTestSupport.ts`
- `apps/web/src/testing/timelineWorkbookRenderTestSupport.tsx`
- Visual snapshots under `apps/web/e2e/workbook.visual.spec.ts-snapshots/`

### Generated Artifacts Or Contract-Derived Inputs Found

Contract-derived inputs relevant to this target:

- `contracts/view-schemas/cartulary.view.timeline.v2.json`
- `contracts/view-schemas/index.json`
- `packages/view-contracts/src/index.ts`
- `packages/ui-contracts/src/index.ts`
- `contracts/ws/index.schema.json`
- `contracts/openapi/cartulary.openapi.yaml`

Generated roots and files that must not be hand-edited:

- `internal/gen/**`
- `packages/protocol-ts/src/generated/**`
- `packages/ui-contracts/src/generated/**`
- `tools/task_surface.generated.mk`

Timeline view-schema facts inspected:

- `view_schema_id`: `cartulary.view.timeline.v2`
- `surface_kind`: `built_in_sheet`
- `source_record_types`: `timeline_event`
- `technical_fields`: `record_id`, `row_version`
- `inline_create.permits_zero_field_create`: `true`
- Default sort: `timeline.activity_sort_ts ASC`, `record_id ASC`
- Grouping fields: `timeline.date_entered_sort_day`, `timeline.activity_time_pair_state`, `timeline.capture_state`, `timeline.has_evidence`, `timeline.has_unresolved_mentions`
- Inspector config: `default_open: false`, `no_row_state: no_row_selected`, panels `details`, `relationships`, `evidence`, `history`, `workflow`
- Inspector feature-group count: 27

### Validation Commands Discovered

Discovered public Make targets:

- `make lint-markdown`
- `make frontend-typecheck`
- `make frontend-unit`
- `make frontend-import-boundary-check`
- `make browser-e2e-webserver-backed`
- `make browser-e2e-stateful`
- `make browser-e2e-a11y-preflight`
- `make browser-e2e-visual`
- `make generated-artifact-policy-check`
- `make json-shape-check`
- `make agent-finalize`

Target guidance was inspected for `frontend-unit`, `frontend-typecheck`, `frontend-import-boundary-check`, `lint-markdown`, `browser-e2e-webserver-backed`, `browser-e2e-stateful`, `browser-e2e-a11y-preflight`, `browser-e2e-visual`, `generated-artifact-policy-check`, and `json-shape-check`. Future implementation sessions must explain any additional target with `make explain-target TARGET=<target> DETAIL=summary` before invoking it.

## 3. Responsibility Diagnosis For TimelineWorkbook.tsx

### Current Responsibilities

`TimelineWorkbook.tsx` currently acts as the Timeline workbook surface assembler and hot-path controller. It coordinates:

- Runtime query state, saved-view inputs, sheet ref, role, density, and entity catalogs.
- Row loading, row normalization, draft row creation, committed row high-water marks, and stale row-version rejection.
- Inline edit, keyboard command handling, scalar/collection draft state, and draft input refs.
- Tabular clipboard paste dispatch to the Timeline paste route.
- Save-state presentation, pending queue refresh blocks, pending replay, conflict anchors, and before-unload protection.
- Same-field conflict registration and conflict resolver coordination.
- Viewport continuity, grid focus restoration, selection anchors, and scroll restoration.
- Inspector selection, default-closed lifecycle, row retargeting, history, evidence, relationships, mentions, workflow actions, and related-row creation.
- Collaboration socket lifecycle, live row patches, presence filtering, and edit-mode presence updates.
- Renderer and component composition for `WorkbookSurfaceFrame`, `TimelineGridSurface`, inspector, toolbar, status strip, notices, context menu, and conflict resolver.

### Likely Overreach

The component has become responsible for multiple internal controllers that can be separated without changing behavior:

- Save-state and pending-presentation bridge logic.
- Viewport/focus continuity retries and DOM scroll restoration.
- Clipboard paste route dispatch and response/conflict handling.
- Presence projection and edit-mode update dispatch.
- Inspector subsection render factories and action wiring.
- Scalar draft/input registry glue that is not inherently presentation.

### Stable Public Contracts It Participates In

- `TimelineWorkbookProps`, `SaveState`, and `IncidentRole`.
- `timelineViewSchemaId` and `cartulary.view.timeline.v2`.
- `record_id`, `row_version`, `base_row_version`, and `field_key` identity.
- View-row query, row creation, record patch, and clipboard-paste routes.
- Timeline view-schema field bindings, sort/filter/group capabilities, and inspector feature-group keys.
- Saved-view interaction through controlled query state, sheet ref, and selector props.
- Save-state labels and conflict-state presentation.
- Stable selectors/test IDs from `@cartulary/ui-contracts`.
- Grid adapter Cartulary anchors from `@cartulary/grid-adapter`.
- WebSocket presence and row-change payload behavior.
- Harness phase accounting, without using phase identity in runtime code.

### Risk Classification

Risk: HIGH.

Reason: the target owns hot-path workbook behavior across create/edit/paste, row-version safety, conflicts, inspector actions, collaboration, focus/scroll continuity, saved-view query state, and selector/test evidence. Refactors must be small, behavior-preserving, and validated after each risky move.

### Seams That Appear Extractable

- Save-state and pending-presentation controller.
- Viewport/focus continuity coordinator.
- Clipboard paste dispatch controller.
- Inspector section render factories.
- Presence projection/edit-mode controller.
- Final assembler cleanup after the above.

### Seams That Must Stay Local Until More Evidence Exists

- `TimelineWorkbookProps` and top-level `TimelineWorkbook` export.
- Direct view-contract constants and `timelineContract` setup.
- `applyRowMutation` and row commit acceptance flow until mutation/pending/conflict interplay is characterized.
- `waitForCommittedRecordIdle` until paste/conflict replay evidence is confirmed.
- WebSocket lifecycle refs and live-update reducers unless a collaboration-specific slice is selected.
- `recordWorkbookTiming` until timing and measurement ownership is verified.
- Any generated contract, selector, route, field key, authorization, or phase-map surface.

## 4. Owner-Contract Map

| Contract surface | Drift risk | Owner source | Current evidence | Guardrail |
| --- | --- | --- | --- | --- |
| Grid-first row creation, inline edit, paste, keyboard, and focus | Hot-path user workflow could shift from grid-first behavior or lose focus/scroll continuity. | Core 03, Core 01, grid-adapter boundary, tests. | Target uses `useTimelineGridAnchorController`, `useTimelineGridInteractions`, `mapWorkbookKeyboardCommand`, and clipboard-paste route. | Keep create/edit/paste grid-first; validate with frontend unit and browser-backed tests for risky slices. |
| `record_id`, `row_version`, `base_row_version`, `field_key` identity | Wrong identity breaks conflict safety, stale update rejection, or row patch application. | Core 01, Core 02, Core 03, Timeline view schema. | Target checks stale row versions and sends paste targets with record/base version and field keys. | Do not replace stable IDs with labels, row order, storage names, or vendor coordinates. |
| Active `view_schema_id` and Timeline field bindings | Field/key drift breaks query, edit, grouping, inspector, and tests. | `contracts/view-schemas/cartulary.view.timeline.v2.json`, `@cartulary/view-contracts`, Core 01. | Target imports `timelineViewSchemaId`, `requireViewContract`, Timeline scalar/collection bindings. | No view-schema ID, field-key, sort/filter/group, or generated consumer changes in this refactor. |
| Query state, sorting, filtering, grouping, saved-view interaction | Saved-view or query state could become runtime identity or overwrite canonical surface identity. | Core 03, Core 01, `workbookQuery`, `WorkbookShell`. | `WorkbookShell` passes controlled query state and saved-view selector; target uses runtime hook and grid controls. | Preserve controlled prop behavior and saved-view query-only boundary. |
| Pending save, conflict, and sync state | Save labels, conflict anchors, and pending replay can regress silently. | Core 03, `workbookPendingQueue`, phase tests. | Target uses `deriveWorkbookSaveState`, conflict queues, pending refresh blocks, resolver coordinator. | Preserve `Syncing`/`Saved`/`Conflict`, same-field conflict anchoring, pending replay, and before-unload semantics. |
| Inspector default-closed and row retargeting | Inspector could open unexpectedly, retain invalid row, or route actions against stale rows. | Core 03, Timeline `inspector_config`, `workbookInspectorModel`. | Contract has `default_open: false`, `no_row_selected`, 5 panels, 27 feature groups. Target uses inspector selection/lifecycle hooks. | Preserve default-closed, no-row state, retarget/clear behavior, feature-group key checks, and selectors. |
| Collaboration, presence, and live updates | Presence might leak across sheets or live patches could lower committed row versions. | Core 01, Core 03, `contracts/ws/index.schema.json`. | Target filters presence by sheet/record/field and rejects stale live row versions. | Preserve sheet matching, self-connection suppression, high-water row version, and WebSocket route/event semantics. |
| Generated contract usage | Manual edits or regenerated drift can invalidate downstream TypeScript contracts. | Generated artifact policy, view-contract package. | Generated roots listed; target imports contract adapters and UI selectors. | Do not hand-edit generated roots or generated task surface. |
| Harness/test accounting | Refactor could encode phase identity or misrepresent visual/a11y evidence. | `docs/testing-harness-nlspec.md`, phase maps, Core 05 boundary. | Phase maps discovered for existing evidence. | Keep phase identity out of production runtime; report product vs harness/config/infra failures separately. |

## 5. Refactor Tracker

| ID | Work item | Workstream | Status | Depends on | Owner | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- | --- |
| T-001 | Define target module and scope | scope | DONE | none | `/apps/web` Timeline workbook surface | This artifact names `TimelineWorkbook.tsx` and excludes production behavior changes. | One module/seam and exclusions are explicit. |
| T-002 | Inspect current repo state | discovery | DONE | T-001 | Planning session | Current-state scan, imports, tests, contracts, generated policy, and Make target guidance in this file. | Relevant files, imports, tests, generated paths, and commands are listed with source limits. |
| T-003 | Map owner contracts | contracts | DONE | T-002 | Core 00-04, Timeline contract, harness NLSpec | Owner-contract map in section 4. | Public behavior and owner docs are mapped. |
| T-004 | Freeze characterization evidence | tests | DONE | T-003 | Implementation session | Existing tests listed; per-slice assertions were validated through the slice gates in section 14. | Existing and missing characterization tests are known for each completed slice. |
| T-005 | Plan boundary guardrails | architecture | DONE | T-003 | `/apps/web`, `/packages/grid-adapter`, generated policy owners | Boundary guardrails in section 9. | Import, generated, selector, phase, and evidence guardrails are defined. |
| T-006 | Plan behavior-preserving moves | implementation | DONE | T-004,T-005 | Future implementation session | Candidate slices in section 7. | Smallest safe move sequence is defined. |
| T-007 | Plan validation loop | validation | DONE | T-006 | Harness NLSpec and Make target owners | Validation plan in section 10. | Cheapest sufficient validation targets are named. |
| T-008 | Update docs/contracts if required | docs | DONE | T-003 | Future implementation session | This planner is the only planned doc artifact; no owner docs or generated contracts are planned for the refactor. | No generated file or owner contract edit is required for behavior-preserving slices. |
| T-009 | Execute or hand off | handoff | DONE | T-006,T-007,T-008 | Implementation session | Final implementation log in section 14. | The refactor slices, validation, and handoff are complete. |

Status values: `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`, `DROPPED`.

## 6. Workflow Dependency Map

| Workflow | Name | Class | Required previous workflows | Required subsequent workflows | Status | Target-specific dependency, evidence, and exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap | root | none | WF-01 | DONE | Evidence: branch, commit, target, framework, source limits recorded. Exit: session header complete. |
| WF-01 | Current-state repository scan | chain | WF-00 | WF-02, WF-03 | DONE | Evidence: file inventory, imports, tests, generated paths, validation commands. Exit: relevant unseen files are marked. |
| WF-02 | Module/package ownership inventory | chain | WF-01 | WF-04, WF-05 | DONE | Evidence: `/apps/web` owns surface orchestration; `/packages/grid-adapter` owns vendor grid integration. Exit: touched slice files were directly inspected. |
| WF-03 | Public contract freeze | chain | WF-01 | WF-04, WF-05 | DONE | Evidence: owner-contract map. Exit: drift-prone route, row identity, view schema, inspector, collaboration, generated, and harness surfaces are frozen. |
| WF-04 | Refactor slice selection | chain | WF-02, WF-03 | WF-05, WF-06 | DONE | Dependency: WF-02/WF-03. Evidence: `TW-SL-01` through `TW-SL-06` were selected and completed sequentially. Exit: each slice stayed reviewable and tracker-scoped. |
| WF-05 | Characterization test plan | chain | WF-03, WF-04 | WF-09 | DONE | Dependency: selected slice. Evidence: section 14 maps each slice to validation evidence. Exit: no additional missing characterization tests were required. |
| WF-06 | Boundary guardrail plan | chain | WF-02, WF-04 | WF-09 | DONE | Dependency: selected slice. Evidence: import boundary and generated-file guardrails stayed satisfied. Exit: guardrails are encoded in final validation. |
| WF-08 | Frontend package seam plan | parallel | WF-04, WF-05, WF-06 | WF-09 | DONE | Dependency: selected frontend slice. Evidence: extractions stayed under Timeline hooks/components. Exit: no public API or package boundary changed. |
| WF-09 | Execution checkpoint plan | chain | WF-05 plus WF-08 | WF-10 | DONE | Dependency: characterization and guardrails. Evidence: each slice was validated before the next started. Exit: implementation completed. |
| WF-10 | Validation and harness accounting plan | chain | WF-09 | WF-11 | DONE | Dependency: checkpoint plan. Evidence: section 14 records Make target results and classified skipped broader gates. Exit: validation results are recorded. |
| WF-11 | Documentation/generated-artifact plan | parallel | WF-03, WF-09 | WF-12 | DONE | Dependency: contract freeze. Evidence: no generated files or owner specs were touched. Exit: generated-artifact policy remains satisfied. |
| WF-12 | Cleanup and anti-drift plan | chain | WF-10, WF-11 | WF-13 | DONE | Dependency: validation and doc plan. Evidence: `TW-SL-06` cleanup and final validation completed. Exit: no phase-shaped runtime structure or broad refactor leftovers were introduced. |
| WF-13 | Handoff and next-slice bootstrap | chain | WF-12 | none | DONE | Dependency: validated implementation slice. Evidence: final section 14 handoff. Exit: next actor can resume from the final record. |

## 7. Candidate Refactor Slices

Each slice must be behavior-preserving and small enough for one reviewable implementation pass. Select only one slice per implementation session unless validation evidence is very strong and the reviewer explicitly allows combining.

| Slice ID | Objective | Files likely touched | Prerequisites | Public behavior expected unchanged | Characterization evidence required before edit | Validation command | Rollback note | Stop condition |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| TW-SL-01 | Extract save-state and pending-presentation bridge logic from `TimelineWorkbook.tsx` into a focused hook/controller. | `TimelineWorkbook.tsx`; likely `useTimelinePendingSaves.ts` or a new Timeline hook. | Direct-inspect pending-save hook, pending queue utility, save-state tests. | Save labels, pending queue messages, conflict anchors, refresh blocking, pending replay, and before-unload behavior. | Existing `WorkbookShell.phase4.saveState.test.tsx`, `workbookPendingQueue.test.ts`, autosave/action-sequencing tests; add a behavior-named unit only if no existing assertion covers pending/conflict/overflow/auth-pause/refresh-block/beforeunload precedence. | `make frontend-unit`; `make frontend-typecheck`. | Revert new hook and restore inline logic in `TimelineWorkbook.tsx`. | Stop if save-state precedence among pending ops, conflict anchors, and refresh blocks is unclear. |
| TW-SL-02 | Extract viewport/focus continuity coordinator around remaining scroll, focus, and anchor restoration glue. | `TimelineWorkbook.tsx`; likely `useTimelineGridAnchorController.ts`, `useTimelineGridInteractions.ts`, or a new hook. | Direct-inspect grid interaction hook, continuity model/tests, keyboard/anchor tests. | Selected record, focused cell, grid scroll, draft focus, inspector retargeting, and no vendor coordinate leakage. | Existing `timelineViewportContinuityModel.test.ts`, `workbookContinuity.test.ts`, `GridAdapter.phase9.anchor.test.ts`, Phase 9 sentinel/inspector tests. | `make frontend-unit`; `make frontend-import-boundary-check`; `make frontend-typecheck`. | Remove extracted coordinator and restore current local functions. | Stop if a DOM selector or vendor-grid coordinate is needed outside `@cartulary/grid-adapter`. |
| TW-SL-03 | Extract tabular paste dispatch controller while preserving current target resolution and route payload. | `TimelineWorkbook.tsx`; likely `useTimelineGridAnchorController.ts` and a new paste controller/helper. | Direct-inspect clipboard helper, paste route tests, phase maps, Core 01 paste contract. | `/api/v1/incidents/{incident_id}/views/{view_schema_id}/clipboard-paste`, `client_txn_id`, `start_field_key`, target record/base version, conflict registration, scalar fallback paste, and focus restore. | Existing `workbookClipboard.test.ts`, Phase 3 payload/grid tests, Phase 9 sentinel/anchor tests, browser-backed paste evidence. | `make frontend-unit`; `make frontend-import-boundary-check`; `make frontend-typecheck`; broaden to `make browser-e2e-webserver-backed` if dispatch behavior changes. | Restore existing `handlePaste` block and remove helper. | Stop if route envelope, target shape, or conflict handling cannot be kept byte-for-byte equivalent. |
| TW-SL-04 | Extract inspector section render factories and action wiring from the main component. | `TimelineWorkbook.tsx`; likely `TimelineWorkbookInspector.tsx`, `TimelineEvidencePanel.tsx`, `TimelineHistoryPanel.tsx`, or a new component/helper. | Direct-inspect inspector component, evidence/history panels, feature-group usage, inspector tests. | Default-closed inspector, `no_row_selected`, row retargeting, active panel, feature-group keys, selectors, and current UI layout. | Existing `WorkbookShell.phase9.inspector.test.tsx`, `TimelineEvidencePanel.test.tsx`, Phase 7 history tests, visual/a11y only if layout changes. | `make frontend-unit`; `make frontend-typecheck`; add `make browser-e2e-a11y-preflight` if rendered structure changes. | Restore local render functions in `TimelineWorkbook.tsx`. | Stop if extraction implies UI redesign, selector rename, or feature-group behavior change. |
| TW-SL-05 | Extract presence projection and edit-mode presence dispatch into a focused hook/controller. | `TimelineWorkbook.tsx`; likely `useTimelineLiveUpdates.ts`, `useTimelineLiveUpdateController.ts`, or a new hook. | Direct-inspect live update hooks/services and collaboration tests. | Presence sheet filtering, self-connection suppression, record/cell matching, display ordering, edit/viewing updates, and stale row-version protection. | Existing `workbookCollaborationMessages.test.ts`, `workbookSocketLifecycle.test.ts`, `WorkbookShell.phase6.test.tsx`, visual presence snapshots as design evidence only. | `make frontend-unit`; `make frontend-typecheck`; broaden to `make browser-e2e-stateful` if WebSocket behavior is touched. | Restore local presence selectors and dispatch calls. | Stop if socket protocol or authorization semantics need Core 04 decisions. |
| TW-SL-06 | Final assembler cleanup after prior slices so `TimelineWorkbook.tsx` remains a composition root. | `TimelineWorkbook.tsx`; any new hooks created by prior slices. | Complete at least one earlier slice with green validation. | All public props, routes, selectors, view schema IDs, field keys, save state, inspector, query, and collaboration behavior. | Existing per-slice characterization plus import boundary evidence. | `make frontend-unit`; `make frontend-typecheck`; `make frontend-import-boundary-check`; `make agent-finalize` before broader verification. | Revert cleanup-only commit or restore prior local helper layout. | Stop if cleanup starts changing behavior or crossing package boundaries. |

## 8. Characterization Test Plan

Existing evidence must be used before adding tests. New tests should name observed behavior, not implementation details. A selected slice must map the named tests below to the exact assertions it relies on before editing.

| Slice | Behaviors to preserve | Existing tests found first | Conditional additional evidence |
| --- | --- | --- | --- |
| `TW-SL-01` save-state and pending presentation | Save-state precedence for pending work, conflicts, queue overflow, auth pause, refresh block, pending replay, beforeunload, and secondary messages. | `WorkbookShell.phase4.saveState.test.tsx`; `WorkbookShell.phase3.autosave.test.tsx`; `WorkbookShell.phase4.actionSequencing.test.tsx`; `WorkbookShell.phase6.test.tsx`; `workbookPendingQueue.test.ts`; `timelineConflictModel.test.ts`. | Add a behavior-named unit/component test only if no existing assertion covers a precedence branch the extraction moves. |
| `TW-SL-02` viewport and focus continuity | Selected record, focused cell, scroll position, draft focus, keyboard movement, invalid anchor clearing, and no vendor-coordinate dependency. | `timelineViewportContinuityModel.test.ts`; `workbookContinuity.test.ts`; `workbookKeyboard.test.ts`; `GridAdapter.phase9.anchor.test.ts`; `WorkbookShell.phase9.sentinel.test.tsx`; relevant `WorkbookShell.phase3.grid.test.tsx` stale refresh cases. | Add coverage only for a moved behavior that lacks an existing focus or continuity assertion. |
| `TW-SL-03` clipboard paste dispatch | Stable paste anchors by `record_id`/`field_key`, route payload with `view_schema_id`, `client_txn_id`, `start_field_key`, create targets, existing-record `base_row_version`, scalar comma handling, and grouped conflict selection continuity. | `workbookClipboard.test.ts`; `GridAdapter.phase9.anchor.test.ts`; `WorkbookShell.phase3.payload.test.tsx`; `WorkbookShell.phase3.autosave.test.tsx`; `WorkbookShell.phase9.sentinel.test.tsx`. | Use `browser-e2e-webserver-backed` for rendered dispatch evidence if the slice touches route dispatch rather than pure helper placement. |
| `TW-SL-04` inspector render/action seam | Default-closed inspector, `no_row_selected`, row retargeting, active panel, feature-group keys, evidence/history/read actions, related-row creation, server auth failure as authoritative, and unchanged selectors. | `WorkbookShell.phase9.inspector.test.tsx`; `TimelineEvidencePanel.test.tsx`; `WorkbookShell.phase7.test.tsx`; `WorkbookShell.phase4.support.test.tsx`; view-schema contract parsing tests where feature groups are relevant. | Run a11y preflight if focus or rendered structure changes; run visual only for intentional layout-affecting movement. |
| `TW-SL-05` presence and live updates | Sheet-ref matching, `connection_id` keyed presence, self-connection suppression, row/cell matching, stale row-version rejection, sequence gap reset/requery, session revocation handling, and no save-state changes from ambient presence. | `workbookCollaborationMessages.test.ts`; `workbookSocketLifecycle.test.ts`; `WorkbookShell.phase6.test.tsx`; backend/platform Phase 6 socket evidence for protocol behavior. | Run `make browser-e2e-stateful` when socket/session semantics, auth pause, or stateful collaboration behavior is touched. |
| `TW-SL-06` assembler cleanup | Public props, routes, selectors, view-schema IDs, field keys, save-state behavior, inspector behavior, query/saved-view behavior, package boundaries, and generated-file boundaries. | All relevant tests from completed prior slices; `frontend-import-boundary-check`; `frontend-typecheck`; `frontend-unit`. | Add no new tests for cleanup-only movement unless a previously untested behavior is moved into a new public helper. |
| Conditional visual/a11y readiness | Rendered structure, focus order, inspector/grid layout, save-state strip, presence markers, and conflict presentation. | `workbook.a11y-preflight.spec.ts`; `workbook.a11y.spec.ts`; `workbook.visual.spec.ts`; existing snapshots. | Treat visual/a11y as design or readiness evidence only unless an owner document separately makes it product conformance evidence. |

## 9. Boundary Guardrails

| Guardrail | Current evidence | Required check |
| --- | --- | --- |
| `/apps/web` must not learn `react-data-grid` vendor coordinate semantics. | Boundary scan found app code imports `@cartulary/grid-adapter`, not `react-data-grid`; direct vendor import is in `packages/grid-adapter/src/index.tsx`. | Run `make frontend-import-boundary-check` after any grid/focus/paste extraction. |
| Direct `react-data-grid` imports must remain inside `/packages/grid-adapter`. | `tools/frontend_import_boundaries.json` enforces `frontend-grid-vendor-boundary`. | Do not add vendor imports or CSS imports in `/apps/web`. |
| Generated files must not be hand-edited. | Generated roots and `tools/task_surface.generated.mk` recorded from `tools/generated_artifact_policy.json`. | Do not edit generated roots; run `make generated-artifact-policy-check` if generation boundaries are touched. |
| Selectors/test IDs must remain stable or be updated through the owning package. | Target imports selector builders from `@cartulary/ui-contracts`. | Do not inline or rename selector/test-id strings in the refactor. |
| Phase identity must not become production runtime structure. | Phase maps are harness accounting only. | Do not import phase maps or phase labels into runtime code. |
| Visual/a11y evidence must not be promoted into product conformance unless an owner says so. | `docs/design.md`, Core 05 boundary, and harness NLSpec separate evidence classes. | Report visual/a11y as design/readiness evidence only. |
| Public route shapes, wire envelopes, view-schema IDs, field keys, authorization, hot-path behavior, selectors, and harness accounting must not drift. | Core docs, contracts, and user non-goals. | Any change to these surfaces is out of scope and must be split into a separate owner-doc decision. |

## 10. Validation Plan

Use the cheapest narrow validation first, then broaden based on the selected slice.

| Scope | Command | When to run | Failure classification |
| --- | --- | --- | --- |
| Planner-only Markdown | `make lint-markdown` | After creating or editing this handoff. | Usually docs/harness formatting unless the target excludes `docs/handoffs`; report exact target result. |
| Frontend typing | `make frontend-typecheck` | After any TypeScript extraction. | Product code failure if type errors are in touched authored files; otherwise classify by target summary. |
| Frontend unit/interaction tests | `make frontend-unit` | After every implementation slice. | Product failure if Timeline/workbook behavior regresses; otherwise classify harness/config/infra using summary artifacts. |
| Import boundaries | `make frontend-import-boundary-check` | After any grid, focus, paste, selector, or package-boundary change. | Product architecture failure if `/apps/web` imports grid vendor or generated internals. |
| Generated/artifact boundaries | `make generated-artifact-policy-check` | Only if generated roots, codegen inputs, or generated policy are touched. | Harness/policy failure unless production code attempted manual generated edit. |
| JSON contract shape | `make json-shape-check` | Only if JSON contracts or manifests are touched. | Contract/data failure. Refactor slices should not need this. |
| Browser webserver-backed | `make browser-e2e-webserver-backed` | For paste, keyboard/focus, inspector, query, or row mutation risk beyond unit coverage. | Product/harness/config/infra based on run summary. |
| Browser stateful | `make browser-e2e-stateful` | For collaboration, presence, live updates, auth, pending replay, or service-backed state. | Product/harness/config/infra based on run summary. |
| A11y preflight | `make browser-e2e-a11y-preflight` | If rendered inspector/grid structure or focus behavior changes. | Accessibility readiness evidence, not product conformance by itself. |
| Visual | `make browser-e2e-visual` | Only if rendered layout is touched and reviewer expects visual evidence. | Design/readiness evidence unless owner/Core 05 publication boundary is separately satisfied. |
| End-of-run | `make agent-finalize` or `make agent-finalize RESULTS_DIR=<successful full warm check run root>` | Before broader end-of-run verification or when retaining successful run evidence. | Report skipped retained-run maintenance if `RESULTS_DIR` is unset. |

Baseline command sequence for every implementation slice:

1. Run `make task-guide ROLE=feature-dev PHASE=<phase>` using the phase most closely matching the selected slice evidence: phase3 for grid/autosave/payload, phase4 for save state/action sequencing/query, phase6 for collaboration/presence/pending replay, phase7 for history/rollback, phase9 for inspector/keyboard/grid anchors.
2. Run `make explain-target TARGET=frontend-typecheck DETAIL=summary`, `make explain-target TARGET=frontend-unit DETAIL=summary`, and `make explain-target TARGET=frontend-import-boundary-check DETAIL=summary`; explain any broader target before invoking it.
3. Run narrow gates after code movement: `make frontend-typecheck`, `make frontend-unit`, and `make frontend-import-boundary-check`.
4. Add conditional broader gates by touched surface: `make browser-e2e-webserver-backed` for paste, keyboard/focus, inspector, query, or row mutation risk; `make browser-e2e-stateful` for collaboration, presence, auth, session revocation, or pending replay; `make browser-e2e-a11y-preflight` for focus/rendered structure; `make browser-e2e-visual` only for layout-affecting changes.
5. Run `make generated-artifact-policy-check` or `make json-shape-check` only when their owner inputs are touched; behavior-preserving Timeline slices should not need either.
6. Run `make agent-finalize` before broader end-of-run verification. If retaining prior successful full warm check evidence, pass `RESULTS_DIR=<successful full warm check run root>`; otherwise report retained-run maintenance skipped because `RESULTS_DIR` was unset.

## 11. Workstream Notes

### Scope And Evidence

- Primary target is exactly `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`.
- Target seam is the frontend Timeline workbook surface inside `/apps/web`.
- Adjacent files are in scope only when imports, props, hooks, route/query/mutation contracts, grid adapter usage, view-schema/field-key contracts, inspector behavior, tests, generated contracts, selectors, or validation commands require them.
- This planner records 26 directly inspected or parsed paths plus Timeline directory inventories and Make target discovery.
- Each implementation slice must update the handoff with the slice ID, selected phase evidence, and files directly inspected during that slice. Use the slice entry inspection manifest rather than inspecting all adjacent files by default.

### Contracts And Docs

- Core 00 through Core 04 own current product conformance behavior; Core 05 is only for claim-bearing timed, benchmark, fixture-sensitive, or publication evidence.
- `docs/domain.md` owns vocabulary and concept boundaries, especially Timeline rough capture, saved-view identity, field-key identity, mentions, evidence, and workbook-native coordination.
- `docs/testing-harness-nlspec.md` owns harness mechanics and Make invocation.
- `docs/design.md` is design direction only.
- Core 04 security rules relevant to future Timeline slices are mapped for planner use: public API and WebSocket auth use opaque server-managed sessions (`REQ-04-002`); cookie-authenticated state-changing HTTP requests require fail-closed CSRF protection (`REQ-04-003`); bootstrap tokens are not ordinary sessions and are not valid on `/ws/v1/*` (`REQ-04-084`); incident WebSocket membership loss terminates the affected subscription and future incident checks fail closed (`REQ-04-017`); current incident roles are exactly `viewer`, `editor`, `reviewer`, and `admin` (`REQ-04-021`); record access inherits incident access (`REQ-04-022`); API routes, evidence handle issuance/redemption, jobs, and WebSocket incident subscriptions re-derive authorization at request time (`REQ-04-023`); saved-view scope does not widen or narrow underlying incident data access (`REQ-04-025`); `deployment_admin` is not an incident-data, preview/download, export, job, or WebSocket bypass (`REQ-04-028`, `REQ-04-029`); inspector visibility is not authorization and inspector actions must re-derive authorization (`REQ-04-127`); browser WebSocket `Origin` must match `application.public_origin` before joining an incident stream (`REQ-04-110`).
- Core 01 WebSocket rules relevant to future Timeline slices are mapped for planner use: the public route is `GET /ws/v1/incidents/{incident_id}`; the bounded message families are `hello`, `resume`, `hello_ack`, `resume_ack`, `presence_snapshot`, `presence_delta`, `presence_update`, `record_changed`, `job_progress`, `ping`, terminal `error`, and `session_revoked`; `presence_snapshot` is keyed by `connection_id`; `record_changed` is keyed by `incident_id`, `record_id`, `row_version`, `changed_field_keys[]`, and `affected_views[]`; browser WebSocket upgrades and session-establishment messages validate `Origin`; session or membership revocation sends `session_revoked` and closes the socket; incident close sends terminal `error` with `code = incident_closed`.
- No normative spec edit is planned for these cleanup notes. If implementation inspection reveals behavior that contradicts Core 00 through Core 04, split that into a separate spec-aligned remediation task rather than hiding it inside a refactor.

### Frontend Package Seam

- `/apps/web` owns workbook shell/controllers and browser state.
- `/packages/grid-adapter` owns direct grid vendor integration and vendor coordinate translation.
- `/packages/view-contracts` owns runtime adapters around view-schema contracts.
- `/packages/ui-contracts` owns selector/test-id builders.
- Current app boundary uses `@cartulary/grid-adapter`; do not cross into direct `react-data-grid`.

### Tests And Harness

- Existing tests cover much of the target through phase suites and Timeline-local model/service tests.
- Tests must describe observed behavior: save state, row identity, conflict anchoring, paste target resolution, inspector lifecycle, presence filtering, focus restoration.
- Phase maps remain harness accounting and must not be imported by production runtime.
- Pick exact tests from the characterization matrix after selecting a slice. Add a new test only when the selected slice moves an observed behavior that existing assertions do not cover; name that test after the behavior, not the extracted helper.

### Generated Artifacts

- No generated file changes are planned.
- Do not hand-edit `internal/gen/**`, `packages/protocol-ts/src/generated/**`, `packages/ui-contracts/src/generated/**`, or `tools/task_surface.generated.mk`.
- Timeline view-schema JSON is an authored contract input, but this refactor should not edit it.

### Risks And Blockers

- HIGH risk due to hot-path workbook behavior and many contract surfaces.
- BLOCKED for implementation if a slice requires changing public routes, wire envelopes, view-schema IDs, field keys, auth behavior, selectors, generated files, or harness accounting.
- Directly inspect only the inventory-only adjacent files named by the selected slice entry manifest before touching them, then record that inspection in the handoff.
- Establish the validation baseline before editing production code by running the command-sequence discovery in section 10 and recording the selected targets in the handoff.

## 12. Session Handoff Template

### Filled Initial Handoff Record

| Field | Value |
| --- | --- |
| Date | 2026-06-28 |
| Branch/commit | `main` / `d318fe30047de05a5c1dbad74910fce7726f5d96` |
| Target file/seam | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`; frontend Timeline workbook surface inside `/apps/web` |
| Files changed in prior Plan Mode session | None. Plan Mode prevented writing. |
| Files changed in artifact implementation session | `docs/handoffs/TimelineWorkbook.refactor-tracker.md` only. |
| Current finding | Target exists and is 3001 lines. Many hooks/components already exist; remaining overreach is mostly orchestration glue around save state, focus/continuity, paste, inspector, presence, and assembler cleanup. The remediation pass resolved open-ended planning TODOs by mapping Core 04/Core 01 security ownership, slice-entry inspections, characterization tests, and validation baselines. |
| Framework decision | Use `docs/handoffs/cartulary_modular_refactor_planning_framework.md`; record `temp/current.md` as present but empty. |
| Current workflow state | WF-00 and WF-01 complete for planning. WF-03 contract freeze recorded. WF-04 slice selection is next. |
| Commands run | `git status --short --branch`; `git rev-parse HEAD`; `wc -l`; `rg`; `find`; `sed`; `node` JSON parses; `make explain-target` for selected targets; targeted Core 04 and Core 01 WebSocket reads for remediation. |
| Validation run | `make lint-markdown` passed after writing this artifact and after the TODO-gap remediation pass. |
| Blockers/source limits | No open-ended Core 04 source limit remains for planner use. Inventory-only adjacent files must be directly inspected according to the selected slice entry manifest before editing. |
| Next recommended workflow | Choose one slice in WF-04, freeze characterization in WF-05, then plan guardrails in WF-06/WF-08 before production edits. |

### Blank Future Handoff Record

| Field | Value |
| --- | --- |
| Date | TODO: |
| Branch/commit | TODO: |
| Selected workflow and slice | TODO: |
| Files directly inspected this session | TODO: |
| Files changed | TODO: |
| Public behavior expected unchanged | TODO: |
| Characterization evidence before edit | TODO: |
| Implementation summary | TODO: |
| Validation commands and results | TODO: |
| Product failures | TODO: |
| Harness/config/infra failures | TODO: |
| Generated files touched | TODO: should be none unless owner/codegen workflow is explicit. |
| Boundary guardrails checked | TODO: |
| Source limits or unseen files remaining | TODO: |
| Next recommended workflow | TODO: |

## 13. Binary Acceptance Criteria

| ID | Criterion | Status | Evidence |
| --- | --- | --- | --- |
| RF-AC-001 | Exactly one primary target file and seam are named. | PASS | `TimelineWorkbook.tsx`; frontend Timeline workbook surface inside `/apps/web`. |
| RF-AC-002 | All inspected in-scope files are listed and all relevant unseen files are marked. | PASS | Section 2 lists inspected files, inventory-only adjacent files, and the slice-entry inspection manifest. |
| RF-AC-003 | Every drift-prone public contract is mapped. | PASS | Section 4 maps grid, row identity, view schema, query/saved views, pending/conflict state, inspector, collaboration, generated contracts, and harness accounting. |
| RF-AC-004 | Behavior-preserving refactors are separated from behavior changes. | PASS | Candidate slices are explicitly behavior-preserving; public route/wire/schema/auth/selector changes are out of scope. |
| RF-AC-005 | Characterization coverage is stated. | PASS | Section 8 maps each candidate slice to existing tests and conditional behavior-named additional evidence. |
| RF-AC-006 | Checkpoint sequence includes validation after risky moves. | PASS | Sections 6, 7, and 10 require validation after each selected slice. |
| RF-AC-007 | Frontend package boundaries are preserved. | PASS | Section 9 preserves `/apps/web`, `/packages/grid-adapter`, `/packages/view-contracts`, and `/packages/ui-contracts` boundaries. |
| RF-AC-008 | No generated file hand-edit is planned. | PASS | Sections 2, 9, and 11 list generated roots and prohibit manual edits. |
| RF-AC-009 | No phase-shaped runtime dependency is introduced. | PASS | Sections 4, 9, and 11 state phase maps are harness accounting only. |
| RF-AC-010 | Handoff is sufficient for restart. | PASS | Section 12 provides filled and blank handoff records plus next workflow. |

## 14. Implementation Execution Log

### TW-SL-01 Save State And Pending Presentation

| Field | Value |
| --- | --- |
| Status | DONE |
| Date | 2026-06-28 |
| Direct inspection before edit | `TimelineWorkbook.tsx`, `useTimelinePendingSaves.ts`, `workbookPendingQueue.ts`, `TimelineWorkbookNotices.tsx`, `WorkbookStatusStrip.tsx` |
| Contract reread | Core 03 save-state and pending-queue owner rows via tracker owner map; Core 04 session/auth failure rows via tracker owner map. No spec contradiction found before edit. |
| Characterization evidence selected | `WorkbookShell.phase4.saveState.test.tsx`, `WorkbookShell.phase3.autosave.test.tsx`, `WorkbookShell.phase4.actionSequencing.test.tsx`, `workbookPendingQueue.test.ts` through `make frontend-unit`. |
| Boundary guardrails | Internal hook only; no generated roots, public props, routes, selectors, view-schema IDs, field keys, auth behavior, or phase maps are touched. |
| Implementation summary | Added `useTimelineSaveStatePresentation.ts`; moved save-state derivation, pending snapshot publication, refresh-block begin/finish refs, `beginSave`/`finishSave`, and beforeunload protection out of `TimelineWorkbook.tsx`. |
| Validation | Initial run failed because the extracted local-conflict mapping used a nonexistent `payload`; fixed to preserve the existing `entry.anchor` mapping. Final `make frontend-typecheck` passed with run root `.cartulary/test-results/20260628T185352Z-p2780911/frontend-typecheck`; final `make frontend-unit` passed with run root `.cartulary/test-results/20260628T185352Z-p2780987/frontend-unit`. |
| Generated files touched | None. |
| Next slice | `TW-SL-02` viewport and focus continuity. |

### TW-SL-02 Viewport And Focus Continuity

| Field | Value |
| --- | --- |
| Status | DONE |
| Date | 2026-06-28 |
| Direct inspection before edit | `TimelineWorkbook.tsx`, `useTimelineGridAnchorController.ts`, `useTimelineGridInteractions.ts`, `timelineViewportContinuityModel.ts`, `workbookContinuity.ts`, `workbookKeyboard.ts` |
| Contract reread | Core 03 grid-first focus and keyboard behavior via tracker owner map; grid-adapter boundary policy via tracker guardrails. No spec contradiction found before edit. |
| Characterization evidence selected | `timelineViewportContinuityModel.test.ts`, `workbookContinuity.test.ts`, `workbookKeyboard.test.ts`, `GridAdapter.phase9.anchor.test.ts`, `WorkbookShell.phase9.sentinel.test.tsx` through `make frontend-unit`; `make frontend-import-boundary-check` for vendor boundary. |
| Boundary guardrails | New code must stay inside `/apps/web` Timeline hooks and import only `@cartulary/grid-adapter`, not `react-data-grid`; no selector/test-id changes. |
| Implementation summary | Added `useTimelineViewportContinuityController.ts`; moved input resolution, scroll snapshots, focus/viewport restoration, continuity token transitions, barrier settling, and restoration effect out of `TimelineWorkbook.tsx`. |
| Validation | Initial run failed because the extracted scrollport helper filtered/fell back instead of preserving the prior exact-scrollport requirement; fixed to match existing behavior. Final `make frontend-typecheck` passed with run root `.cartulary/test-results/20260628T190016Z-p2787476/frontend-typecheck`; `make frontend-import-boundary-check` passed with run root `.cartulary/test-results/20260628T190016Z-p2787498/frontend-import-boundary-check`; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T190016Z-p2787735/frontend-unit`. |
| Generated files touched | None. |
| Next slice | `TW-SL-03` clipboard paste dispatch. |

### TW-SL-03 Clipboard Paste Dispatch

| Field | Value |
| --- | --- |
| Status | DONE |
| Date | 2026-06-28 |
| Direct inspection before edit | `TimelineWorkbook.tsx`, `useTimelineGridAnchorController.ts`, `workbookClipboard.ts`, `timelineMutationRequests.ts`, `workbookTimelineModel.ts` |
| Contract reread | Core 01 paste identity and route envelope via tracker owner map; Core 03 grid paste/focus behavior via tracker owner map. No spec contradiction found before edit. |
| Characterization evidence selected | `workbookClipboard.test.ts`, `GridAdapter.phase9.anchor.test.ts`, `WorkbookShell.phase3.payload.test.tsx`, `WorkbookShell.phase9.sentinel.test.tsx` through `make frontend-unit`; `make frontend-import-boundary-check` for grid boundary. |
| Boundary guardrails | Preserve `/api/v1/incidents/{incident_id}/views/{view_schema_id}/clipboard-paste`, `view_schema_id`, `client_txn_id`, `start_field_key`, `columns`, target identities, conflict registration, scalar fallback paste, and focus restore. |
| Implementation summary | Added `useTimelineClipboardPasteController.ts`; moved tabular paste route dispatch, target payload construction, conflict registration, scalar fallback paste scheduling, and post-paste focus/refresh handling out of `TimelineWorkbook.tsx`. |
| Validation | `make frontend-typecheck` passed with run root `.cartulary/test-results/20260628T190434Z-p2791237/frontend-typecheck`; `make frontend-import-boundary-check` passed with run root `.cartulary/test-results/20260628T190434Z-p2791259/frontend-import-boundary-check`; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T190434Z-p2791524/frontend-unit`. Browser webserver-backed validation was not run because route dispatch semantics and payload shape were moved without modification and narrow evidence passed. |
| Generated files touched | None. |
| Next slice | `TW-SL-04` inspector render/action seam. |

### TW-SL-04 Inspector Render And Action Seam

| Field | Value |
| --- | --- |
| Status | DONE |
| Date | 2026-06-28 |
| Direct inspection before edit | `TimelineWorkbook.tsx`, `TimelineWorkbookInspector.tsx`, `TimelineEvidencePanel.tsx`, `TimelineHistoryPanel.tsx`, `TimelineRowActions.tsx`, `workbookInspectorModel.ts`, `contracts/view-schemas/cartulary.view.timeline.v2.json` |
| Contract reread | Core 03 inspector lifecycle/feature-group behavior and Core 04 inspector authorization boundary via tracker owner map. No spec contradiction found before edit. |
| Characterization evidence selected | `WorkbookShell.phase9.inspector.test.tsx`, `TimelineEvidencePanel.test.tsx`, `WorkbookShell.phase7.test.tsx`, `WorkbookShell.phase4.support.test.tsx` through `make frontend-unit`. |
| Boundary guardrails | Preserve default-closed/no-row behavior in `TimelineWorkbookInspector.tsx`, feature-group keys, selectors, and current layout. Extraction must not imply authorization from control visibility. |
| Implementation summary | Added `TimelineWorkbookInspectorSections.tsx`; moved operational-text editors, relationship editors, evidence attach section, create-related workflow section, and row-history section factories out of `TimelineWorkbook.tsx`. `TimelineWorkbookInspector.tsx` layout and selectors are unchanged. |
| Validation | `make frontend-typecheck` passed with run root `.cartulary/test-results/20260628T190745Z-p2794547/frontend-typecheck`; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T190756Z-p2794935/frontend-unit`. A11y preflight was not run because the rendered inspector structure/focus order was not intentionally changed. |
| Generated files touched | None. |
| Next slice | `TW-SL-05` presence and live updates. |

### TW-SL-05 Presence And Live Updates

| Field | Value |
| --- | --- |
| Status | DONE |
| Date | 2026-06-28 |
| Direct inspection before edit | `TimelineWorkbook.tsx`, `useTimelineLiveUpdates.ts`, `useTimelineLiveUpdateController.ts`, `workbookCollaborationMessages.ts`, `workbookSocketLifecycle.ts`, `contracts/ws/index.schema.json`, `workbookPresence.ts` |
| Contract reread | Core 01 WebSocket route/message ownership, Core 03 presence/live-update interaction behavior, and Core 04 session/origin/membership/revocation rules via tracker owner map. No spec contradiction found before edit. |
| Characterization evidence selected | `workbookCollaborationMessages.test.ts`, `workbookSocketLifecycle.test.ts`, `WorkbookShell.phase6.test.tsx` through `make frontend-unit`. |
| Boundary guardrails | Leave socket wire shape, resume semantics, authorization/session behavior, stale row-version handling, and live update reducer behavior unchanged. Extract only active-sheet filtering, self-connection suppression, row/cell presence lookup, and edit-mode presence draft dispatch. |
| Implementation summary | Added `useTimelinePresenceProjection.ts`; moved active-sheet presence filtering, self-connection suppression, row/cell presence lookups, and edit-mode presence draft/send logic out of `TimelineWorkbook.tsx`. Socket protocol and lifecycle hooks were unchanged. |
| Validation | `make frontend-typecheck` passed with run root `.cartulary/test-results/20260628T191106Z-p2797891/frontend-typecheck`; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T191106Z-p2797956/frontend-unit`. Stateful browser validation was not run because WebSocket wire/session behavior was not changed. |
| Generated files touched | None. |
| Next slice | `TW-SL-06` assembler cleanup and final validation. |

### TW-SL-06 Assembler Cleanup And Final Validation

| Field | Value |
| --- | --- |
| Status | DONE |
| Date | 2026-06-28 |
| Direct inspection before edit | `TimelineWorkbook.tsx`; all helpers created by `TW-SL-01` through `TW-SL-05`; tracker handoff sections. |
| Contract reread | Prior slice owner maps remain controlling. No public props, routes, selectors, view-schema IDs, field keys, generated files, auth behavior, or phase accounting are in scope for cleanup. |
| Characterization evidence selected | Per-slice evidence plus `make frontend-import-boundary-check`, `make frontend-typecheck`, `make frontend-unit`, `make lint-markdown`, and `make agent-finalize`. |
| Boundary guardrails | Remove stale imports/local glue only; do not cross package boundaries or introduce phase-shaped runtime structure. |
| Implementation summary | Ran assembler cleanup after all five extractions; kept `TimelineWorkbook.tsx` as the composition root while moving focused logic into Timeline-local hooks/components. Target file is now 2160 lines, down from 3001 at planning inspection. Ran targeted Biome formatting on the six touched frontend files after `lint-biome` reported import/format diagnostics. |
| Validation | Final `make frontend-typecheck` passed with run root `.cartulary/test-results/20260628T191353Z-p2804144/frontend-typecheck`; final `make frontend-unit` passed with run root `.cartulary/test-results/20260628T191353Z-p2804146/frontend-unit`; final `make frontend-import-boundary-check` passed with run root `.cartulary/test-results/20260628T191353Z-p2804188/frontend-import-boundary-check`; final `make lint-markdown` passed with run root `.cartulary/test-results/20260628T191702Z-p2809699/lint-markdown`; final `make lint-biome` passed with run root `.cartulary/test-results/20260628T191354Z-p2804612/lint-biome`; `make agent-finalize` passed with run root `.cartulary/test-results/20260628T191531Z-p2808708/agent-finalize`. |
| Product failures | None remaining. Earlier slice-local failures were fixed before advancing: `TW-SL-01` conflict-anchor mapping and `TW-SL-02` scrollport exactness. |
| Harness/config/infra failures | None remaining. Initial `lint-biome` diagnostics were style/import organization diagnostics and were resolved with targeted formatting. |
| Generated files touched | None. `make agent-finalize` reported `generated=unchanged files=0`; retained-run maintenance was skipped because `RESULTS_DIR` was unset. |
| Skipped broader checks | `browser-e2e-webserver-backed`, `browser-e2e-stateful`, and `browser-e2e-a11y-preflight` were not run because no route semantics, socket/session behavior, or rendered focus/layout structure intentionally changed beyond code movement covered by unit/type/import-boundary gates. |
| Final handoff | Refactor implementation complete. No owner spec, generated artifact, public API, route, selector, view-schema ID, field key, auth behavior, or phase accounting changes were made. |
