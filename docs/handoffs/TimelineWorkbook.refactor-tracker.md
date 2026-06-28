# TimelineWorkbook Refactor Tracker and Handoff Planner

## 1. Session Header

| Field | Value |
| --- | --- |
| Target file | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx` |
| Target module / package / seam | Frontend Timeline workbook surface inside `/apps/web`; consumes `/packages/grid-adapter`, `/packages/view-contracts`, `/packages/ui-contracts`, and generated protocol-derived contracts |
| Planning mode | Planning artifact only; no production-code refactor, runtime behavior change, generated-file edit, selector change, route change, or UI redesign planned |
| Framework path used | `docs/handoffs/cartulary_modular_refactor_planning_framework.md` |
| Framework fallback note | `temp/current.md` exists but is an inspector patch plan, not the reusable modular refactor framework |
| Primary authority reminders | `AGENTS.md`; `docs/spec/00_document_set_status_and_precedence.md` through Core 04; `docs/testing-harness-nlspec.md` for harness mechanics; `docs/domain.md` for vocabulary and boundaries |
| Inspected-file count | 103 distinct paths by direct read or targeted search output; full file reads were narrower than search-based inspections |

## 2. Current-State Repository Scan

### Inspected Files Table

| Area | Inspected paths | Mode | Notes |
| --- | --- | --- | --- |
| Repository procedure | `AGENTS.md` | Read | Confirms authority, generated-file policy, command usage, validation handoff expectations |
| Framework and fallback | `docs/handoffs/cartulary_modular_refactor_planning_framework.md`; `temp/current.md` | Read | Framework is planning skeleton only; `temp/current.md` is not the reusable framework |
| Domain and owner docs | `docs/domain.md`; `docs/spec/00_document_set_status_and_precedence.md`; `docs/spec/01_architecture_storage_and_view_contracts.md`; `docs/spec/02_domain_model_schema_and_history.md`; `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`; `docs/spec/04_security_deployment_and_conformance.md`; `docs/testing-harness-nlspec.md`; `docs/guides/cartulary_frontend_implementation_testing_guide.md`; `docs/guides/cartulary-dev-guide.md` | Read / targeted search | Narrow owner citations are recorded in the closed gap evidence subsection |
| Target | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx` | Read / targeted declaration scan | 7065 lines; large hot-path controller |
| Timeline components | `TimelineCellEditors.tsx`; `TimelineConflictResolver.tsx`; `TimelineEvidencePanel.tsx`; `TimelineGridSurface.tsx`; `TimelineHistoryPanel.tsx`; `TimelineMentionsPanel.tsx`; `TimelinePresenceMarkers.tsx`; `TimelineRowActions.tsx`; `TimelineWorkbookGrid.tsx`; `TimelineWorkbookInspector.tsx`; `TimelineWorkbookNotices.tsx` | Declaration / import scan | Adjacent presentational and panel surfaces |
| Timeline hooks | `useTimelineCommittedRows.ts`; `useTimelineConflicts.ts`; `useTimelineEvidenceActions.ts`; `useTimelineGridInteractions.ts`; `useTimelineHistoryState.ts`; `useTimelineInspectorSelection.ts`; `useTimelineLiveUpdates.ts`; `useTimelineMentions.ts`; `useTimelinePendingSaves.ts`; `useTimelineRows.ts`; `useTimelineWorkbookRuntime.ts` | Declaration scan | Existing extraction seams around state, pending saves, live updates, history, selection, rows, and runtime state |
| Timeline models/services | `timelineConflictModel.ts`; `timelineRowsModel.ts`; `timelineViewportContinuityModel.ts`; `timelineHistoryModel.ts`; `workbookMentionChips.ts`; `workbookTimelineModel.ts`; `timelineMutationRequests.ts`; `workbookCollaborationMessages.ts`; `workbookSocketLifecycle.ts` | Declaration scan | Existing pure/model/service seams already carry some behavior out of the component |
| Shared app services/components/models/utils | `services/browserApi.ts`; `services/workbookApi.ts`; `services/workbookEvidence.ts`; `workbook/components/GenericMutationControl.tsx`; `WorkbookGridControls.tsx`; `WorkbookSheetToolbar.tsx`; `WorkbookStatusStrip.tsx`; `WorkbookSurfaceFrame.tsx`; `models/evidenceLifecycleViewModel.ts`; `genericWorkbookModel.ts`; `workbookInspectorModel.ts`; `workbookQuery.ts`; `workbookReferenceOptions.ts`; `workbookStartup.ts`; `workbookSurfaceRegistry.ts`; `utils/workbookClipboard.ts`; `workbookContinuity.ts`; `workbookGridFocus.tsx`; `workbookKeyboard.ts`; `workbookPendingQueue.ts`; `workbookPresence.ts`; `workbookStyles.ts`; `workbookValueFormat.ts`; `WorkbookShell.tsx` | Import / declaration / caller scan | Current app-owned workbook shell, query, pending queue, focus, continuity, presence, and evidence dependencies |
| Package seams | `packages/grid-adapter/src/index.tsx`; `packages/grid-adapter/src/css.d.ts`; `packages/view-contracts/src/index.ts`; `packages/ui-contracts/src/index.ts` | Read / targeted search | Direct `react-data-grid` import is adapter-owned; view and UI contract packages provide contract and selector facades |
| Contract and generated policy | `contracts/view-schemas/cartulary.view.timeline.v2.json`; `tools/generated_artifact_policy.json`; `tools/frontend_import_boundaries.json` | Read / targeted search | Timeline view schema is contract input; generated roots must not be edited; import-boundary policy enforces adapter boundary |
| Unit/integration support | `apps/web/src/testing/timelineWorkbookRenderTestSupport.tsx`; `apps/web/src/testing/timelineWorkbookTestSupport.ts`; `apps/web/src/testing/timelineWorkbookTestSupport.test.tsx` | Search / declaration scan | Timeline route/fetch helper coverage and render wrapper |
| Unit/integration tests | `App.test.tsx`; `fontRoles.test.tsx`; `App.landing.test.tsx`; `App.phase1.test.tsx`; `App.phase1.support.test.tsx`; `WorkbookShell.assessments.test.tsx`; `WorkbookShell.phase3.grid.test.tsx`; `WorkbookShell.phase3.autosave.test.tsx`; `WorkbookShell.phase3.payload.test.tsx`; `WorkbookShell.phase4.actionSequencing.test.tsx`; `WorkbookShell.phase4.saveState.test.tsx`; `WorkbookShell.phase4.support.test.tsx`; `WorkbookShell.phase4.timelineQuery.test.tsx`; `WorkbookShell.phase5.test.tsx`; `WorkbookShell.phase5.mentionChips.test.tsx`; `WorkbookShell.phase5.gridProvenance.test.tsx`; `WorkbookShell.phase6.test.tsx`; `WorkbookShell.phase7.test.tsx`; `WorkbookShell.phase8.query.test.tsx`; `WorkbookShell.phase9.inspector.test.tsx`; `WorkbookShell.phase9.sentinel.test.tsx`; `TimelineEvidencePanel.test.tsx`; `timelineConflictModel.test.ts`; `timelineRowsModel.test.ts`; `timelineViewportContinuityModel.test.ts`; `workbookTimelineModel.test.ts`; `workbookCollaborationMessages.test.ts`; `workbookSocketLifecycle.test.ts`; `workbookPendingQueue.test.ts`; `workbookClipboard.test.ts`; `workbookInspectorModel.test.ts`; `workbookQuery.test.ts`; `workbookSurfaceRegistry.test.ts`; `GridAdapter.phase9.anchor.test.ts` | Search / test-name scan | Existing characterization is broad but not necessarily per-slice closed |
| Browser, visual, a11y tests | `apps/web/e2e/frontend.phase3.spec.ts`; `frontend.phase4.public-route.spec.ts`; `frontend.phase4.timeline-query.spec.ts`; `frontend.phase6.collaboration.spec.ts`; `frontend.phase7.history.spec.ts`; `frontend.phase9.keyboard.spec.ts`; `frontend.phase9.inspector-actions.spec.ts`; `frontend.phase9.sentinel.spec.ts`; `workbook.visual.spec.ts`; `workbook.a11y.spec.ts`; `workbook.a11y-preflight.spec.ts` | Search / scenario scan | Browser suites exist for route, collaboration, history, inspector, keyboard, sentinel, visual, and a11y paths |

### Direct Imports and Imported Symbols

| Source | Imported symbols / purpose |
| --- | --- |
| `@cartulary/grid-adapter` | `buildGridPresentationRows`, `navigateGridCellAnchor`, `reconcileRecordRows`, `resolveGridPasteTargets`, and grid types. This is the only allowed app-facing grid vendor seam. |
| `@cartulary/ui-contracts` | Selector and test-ID builders for conflict markers, draft cells, generic create, group rows, row gutter, sort header, mentions, relationships, Timeline collection/input/inspector/mutation/row-version/scalar-editor IDs, and `WorkbookSurface`. |
| `@cartulary/view-contracts` | `requireViewContract`, `resolveHeaderSortFieldKey`, `InspectorFeatureGroup`, `ViewContract`. |
| React / ReactDOM | State, effects, refs, memoization, transitions, events, `flushSync`. |
| App services | `apiPath`, `fetchJSON`, `parseErrorMessage`, `readEnvelope`, `createAndAttachEvidenceBlob`, `evidencePublicErrorMessage`. |
| Shared workbook components | Generic mutation control, grid controls, sheet toolbar, status strip, surface frame and frame styles. |
| Shared workbook models | Evidence count view model, generic-create model, inspector config, query state/request builders, reference options, sheet ref, surface registry IDs. |
| Shared workbook utils | Clipboard dimensions, viewport continuity, focus anchor type, keyboard command mapping, pending queue helpers, presence helpers, hidden style, grid value formatting. |
| Timeline hooks | Committed rows, conflicts, evidence action state, grid interactions, history state, inspector selection, live updates, mentions, pending saves, rows, runtime state. |
| Timeline models/services | Same-field conflict parsing, grid row projection, viewport continuity barriers, mention chips, Timeline field/payload/row models, mutation request builders, collaboration message builders, socket lifecycle state. |
| Timeline components | Cell editors, conflict resolver, evidence panel, grid surface, history panel, presence markers, row actions, inspector, notices. |

### Exported Symbols and Known Callers

| Export | Known callers / evidence | Contract sensitivity |
| --- | --- | --- |
| `TimelineWorkbook` | `WorkbookShell.tsx`; app tests; Timeline render test support; phase3/4/5/6/7/9 tests | Public app component surface; props must remain compatible |
| `TimelineWorkbookProps` | Test render helper via `ComponentProps<typeof TimelineWorkbook>`; production shell uses props at render site | Public TS prop contract |
| `SaveState`, `IncidentRole` | Target-local and test-facing types found by export scan | Do not rename without caller scan |
| `WorkbookRecordFreshnessDecision`, `WorkbookVersionedRecord` | Re-exported from Timeline model | Existing compatibility bridge |
| `buildAssessmentCreatePayload`, `buildCreatePayload`, `clipboardTextLooksTabular`, `confidenceScoreFromBand`, `createDraftRow`, `decideWorkbookRecordFreshness`, `ensureDraftRow`, `pendingReplayCapacity` | `WorkbookShell.tsx`, assessment tests, phase3/6/sentinel tests, model tests | Keep exports or migrate callers in a separate compatibility slice |
| `buildRecordRollbackTargetFromHistoryAction` | Used in target and history/rollback tests | History rollback contract helper |

### Adjacent Hooks, Components, and Utilities

| Seam | Current adjacent owner evidence | Extraction relevance |
| --- | --- | --- |
| Runtime/query state | `useTimelineWorkbookRuntime`, `workbookQuery`, `WorkbookGridControls`, `WorkbookSheetToolbar` | Query/sort/filter/group/save-view interaction guardrail |
| Row state and freshness | `useTimelineRows`, `useTimelineCommittedRows`, `timelineRowsModel`, `workbookTimelineModel` | Stable `record_id` and `row_version` identity |
| Pending saves | `useTimelinePendingSaves`, `workbookPendingQueue`, `timelineMutationRequests` | High-risk autosave/replay/conflict hot path |
| WebSocket/live updates | `useTimelineLiveUpdates`, `workbookCollaborationMessages`, `workbookSocketLifecycle` | Collaboration and live patch behavior |
| Inspector/history | `useTimelineInspectorSelection`, `useTimelineHistoryState`, `TimelineWorkbookInspector`, `TimelineHistoryPanel` | Default-closed inspector, row retarget, rollback/delete/restore |
| Mentions | `useTimelineMentions`, `workbookMentionChips`, `TimelineMentionsPanel` | Mention actions and auto-resolution notices |
| Grid/focus/paste | `TimelineGridSurface`, `TimelineWorkbookGrid`, `workbookGridFocus`, `workbookKeyboard`, `workbookContinuity`, `workbookClipboard`, `@cartulary/grid-adapter` | Must keep Cartulary anchor semantics and reject vendor-coordinate leakage |
| Evidence/create-related | `TimelineEvidencePanel`, `workbookEvidence`, generic-create model, surface registry IDs | Inspector workflow and route-binding-sensitive flows |

### Tests and Fixtures Found

| Area | Existing evidence found | Notes |
| --- | --- | --- |
| Grid row identity and query | `WorkbookShell.phase3.grid.test.tsx`; `WorkbookShell.phase4.timelineQuery.test.tsx`; `timelineRowsModel.test.ts`; `workbookTimelineModel.test.ts` | Covers `record_id`, `row_version`, field cells, sorted/filtered/grouped behavior |
| Autosave and pending saves | `WorkbookShell.phase3.autosave.test.tsx`; `WorkbookShell.phase4.actionSequencing.test.tsx`; `WorkbookShell.phase4.saveState.test.tsx`; `workbookPendingQueue.test.ts` | Covers save labels, queued units, action sequencing, capacity |
| Clipboard/keyboard/focus/sentinels | `WorkbookShell.phase9.sentinel.test.tsx`; `GridAdapter.phase9.anchor.test.ts`; `workbookClipboard.test.ts`; browser phase9 specs | Critical before grid/paste/focus extraction |
| Inspector/history/rollback | `WorkbookShell.phase7.test.tsx`; `WorkbookShell.phase9.inspector.test.tsx`; browser phase7 and phase9 inspector specs | Critical before history or create-related extraction |
| Collaboration/presence/live updates | `WorkbookShell.phase6.test.tsx`; `workbookCollaborationMessages.test.ts`; `workbookSocketLifecycle.test.ts`; `frontend.phase6.collaboration.spec.ts` | Critical before WebSocket/live-update extraction |
| Evidence/mentions | `WorkbookShell.phase5.test.tsx`; `WorkbookShell.phase5.mentionChips.test.tsx`; `TimelineEvidencePanel.test.tsx`; `WorkbookShell.phase4.support.test.tsx` | Critical before evidence or mention extraction |
| Selector/test ID contracts | `packages/ui-contracts/src/index.ts`; UI-contract uses in tests; `tools/frontend_import_boundaries.json` | Selector stability must be preserved or changed through owner package |
| Visual/a11y readiness | `workbook.visual.spec.ts`; `workbook.a11y.spec.ts`; `workbook.a11y-preflight.spec.ts` | Design/support evidence only unless owner promotes claim boundary |

### Generated Artifacts or Contract-Derived Inputs Found

| Surface | Evidence | Rule |
| --- | --- | --- |
| Generated roots | `tools/generated_artifact_policy.json` lists `internal/gen`, `packages/protocol-ts/src/generated`, `packages/ui-contracts/src/generated` | Do not hand-edit |
| Generated topology/task surface | `tools/task_surface.generated.mk` and generated topology outputs are policy-owned | Do not hand-edit |
| Timeline view schema | `contracts/view-schemas/cartulary.view.timeline.v2.json` | Authored contract input; update only if an owner change requires it, not for behavior-preserving refactor |
| View contracts | `packages/view-contracts/src/index.ts` consumes generated protocol artifacts and exposes Timeline feature groups | Do not treat app code as owner of field keys or feature IDs |
| UI contracts | `packages/ui-contracts/src/index.ts` owns selector/test-ID builders | Stable selectors stay here |
| Import boundaries | `tools/frontend_import_boundaries.json` enforces direct `react-data-grid` containment | Validate with `make frontend-import-boundary-check` |

### Closed Gap Evidence: Owner Citations, Saved Views, and Timeline Feature Groups

#### Core Owner Citation Closure

- Core 00 authority and precedence: `docs/spec/00_document_set_status_and_precedence.md:5-17`, `:24-35`, `:95-110`.
- Core 00 saved-view, startup, view-schema, inspector, identifier, and invariant ownership rows: `docs/spec/00_document_set_status_and_precedence.md:133-134`, `:145`, `:154`, `:226-236`, `:263-280`.
- Core 01 saved-view and workbook-preference contract owner evidence: `docs/spec/01_architecture_storage_and_view_contracts.md:2079-2217`.
- Core 01 presence, live-update, inspector, workbook-surface, and feature-registry owner evidence: `docs/spec/01_architecture_storage_and_view_contracts.md:3800-3832`, `:3895-3905`, `:4041-4214`, `:4359-4430`, `:4500`.
- Core 03 saved-view interaction, startup, query/grouping, and presence owner evidence: `docs/spec/03_workbook_interaction_collaboration_and_workflows.md:81-123`, `:181-247`, `:258-286`, `:744-762`, `:1871-1905`, `:1984-1985`.
- Core 04 saved-view security and conformance evidence: `docs/spec/04_security_deployment_and_conformance.md:167-172`, `:790-799`, `:1058-1077`, `:1774-1777`.
- Domain/design/harness boundaries: `docs/domain.md:100-110`, `:121-123`, `:139-140`, `:246-258`, `:430-438`, `:570-604`; `docs/design.md:241-255`, `:1505`, `:1546-1548`, `:1558-1564`, `:1637-1644`, `:1712-1723`; `docs/testing-harness-nlspec.md:315-325`.

#### Saved-View Interaction Evidence Closure

- Owner conclusion: saved views are incident-bound configurations over immutable `view_schema_id`; `saved_view` sheet refs are distinct from canonical `view_schema` base surfaces; saved-view scope controls discoverability/mutability only, not row or evidence access.
- Implementation ownership: `useWorkbookShellRuntime.ts` owns saved-view state and commands (`apps/web/src/workbook/hooks/useWorkbookShellRuntime.ts:122-127`, `:192-265`, `:305-444`, `:552-645`, `:797-810`).
- Model ownership: saved-view resource normalization, mutability, canonical query/layout, identity fallback, and modified-state logic live in `workbookSavedViews.ts:10-118`, `workbookSavedViewRuntime.ts:38-118`, and `workbookQuery.ts:196-258`.
- Selector ownership: `ActiveSurfaceSavedViewSelector.tsx:28-112`, `:140-185`, `:253-435` filters saved views to active `view_schema_id`, emits `saved_view` versus `view_schema` sheet refs, and guards create/update/delete/duplicate/home/default actions.
- Shell integration: `WorkbookShell.tsx:1684-1718`, `:2097-2114`, `:2308-2325` composes the selector and passes controlled query/sheet state into `TimelineWorkbook`.
- Timeline participation: `TimelineWorkbook.tsx:357-365`, `:992-1000`, `:1340-1450`, `:2379-2383`, `:3246-3338`, `:4625-4678`, `:6655-6699` consumes `savedViewSelector`, `sheetRef`, `reloadToken`, `queryState`, and `onQueryStateChange`; it does not own saved-view lifecycle.
- Test evidence: `WorkbookShell.phase8.query.test.tsx:171-390`, `apps/web/e2e/phase2.spec.ts:535-686`, `apps/web/e2e/phase8.workbook.support.spec.ts:55-244`, `apps/web/e2e/phase9.inspector-actions.spec.ts:487-536`, and `workbookInspectorModel.test.ts:54-82` cover saved-view query/layout portability, active-surface filtering, `sheet_ref.kind="saved_view"`, home/default pointers, and inspector inheritance/default-closed behavior.

#### Timeline Feature-Group Inventory Closure

Inventory source: `contracts/view-schemas/cartulary.view.timeline.v2.json`. The feature-group count is `27` by `.inspector_config.feature_groups | length`; this is authored contract input and must not be treated as generated output.

| Feature group | Panel | Route binding | Role | Mutates / confirms | Target / seed | Evidence |
| --- | --- | --- | --- | --- | --- | --- |
| `details.read` | `details` | `panel_read` / `current_row_projection` | none | false / false | none | `contracts/view-schemas/cartulary.view.timeline.v2.json:121-131` |
| `relationships.read` | `relationships` | `panel_read` / `current_row_projection` | none | false / false | none | `contracts/view-schemas/cartulary.view.timeline.v2.json:139-149` |
| `evidence.read` | `evidence` | `panel_read` / `current_row_projection` | none | false / false | none | `contracts/view-schemas/cartulary.view.timeline.v2.json:157-167` |
| `history.read` | `history` | `panel_read` / `record_history_route` | none | false / false | none | `contracts/view-schemas/cartulary.view.timeline.v2.json:175-185` |
| `record.delete` | `history` | `record_action` / `record_delete_route` | `editor` | true / true | none | `contracts/view-schemas/cartulary.view.timeline.v2.json:193-204` |
| `record.restore` | `history` | `record_action` / `record_restore_route` | `reviewer` | true / true | none | `contracts/view-schemas/cartulary.view.timeline.v2.json:217-228` |
| `history.rollback` | `history` | `record_action` / `record_rollback_route` | `reviewer` | true / true | none | `contracts/view-schemas/cartulary.view.timeline.v2.json:241-252` |
| `entity_mentions.resolve` | `relationships` | `entity_mention_action` / `entity_mention_resolve_route` | `editor` | true / false | none | `contracts/view-schemas/cartulary.view.timeline.v2.json:264-275` |
| `entity_mentions.create_host` | `relationships` | `entity_mention_action` / `entity_mention_resolve_route` | `editor` | true / true | none | `contracts/view-schemas/cartulary.view.timeline.v2.json:286-297` |
| `entity_mentions.create_identity` | `relationships` | `entity_mention_action` / `entity_mention_resolve_route` | `editor` | true / true | none | `contracts/view-schemas/cartulary.view.timeline.v2.json:308-319` |
| `entity_mentions.dismiss` | `relationships` | `entity_mention_action` / `entity_mention_resolve_route` | `editor` | true / false | none | `contracts/view-schemas/cartulary.view.timeline.v2.json:330-341` |
| `entity_mentions.restore` | `relationships` | `entity_mention_action` / `entity_mention_resolve_route` | `editor` | true / false | none | `contracts/view-schemas/cartulary.view.timeline.v2.json:352-363` |
| `indicator.observations.manage` | `workflow` | `record_patch` / `record_patch_route` | `editor` | true / false | none | `contracts/view-schemas/cartulary.view.timeline.v2.json:374-385` |
| `relationships.manage` | `relationships` | `record_patch` / `record_patch_route` | `editor` | true / false | none | `contracts/view-schemas/cartulary.view.timeline.v2.json:396-407` |
| `evidence.attach_blob` | `evidence` | `record_action` / `evidence_attach_blob_route` | `editor` | true / false | none | `contracts/view-schemas/cartulary.view.timeline.v2.json:418-429` |
| `evidence.preview_handle` | `evidence` | `evidence_access` / `evidence_preview_handle_route` | none | false / false | none | `contracts/view-schemas/cartulary.view.timeline.v2.json:440-451` |
| `evidence.download_handle` | `evidence` | `evidence_access` / `evidence_download_handle_route` | none | false / false | none | `contracts/view-schemas/cartulary.view.timeline.v2.json:461-472` |
| `timeline.mark_reviewed` | `history` | `record_action` / `record_mark_reviewed_route` | `reviewer` | true / false | none | `contracts/view-schemas/cartulary.view.timeline.v2.json:481-492` |
| `timeline.supersede` | `history` | `record_action` / `record_supersede_route` | `reviewer` | true / true | none | `contracts/view-schemas/cartulary.view.timeline.v2.json:503-514` |
| `create_related.note` | `workflow` | `view_row_create` / `view_row_create_route` | `editor` | true / false | `cartulary.view.notes.v1`; no seed | `contracts/view-schemas/cartulary.view.timeline.v2.json:525-537` |
| `create_related.task_request` | `workflow` | `view_row_create` / `view_row_create_route` | `editor` | true / false | `cartulary.view.task_requests.v1`; `task.linked_record_ids:selected_record_id` | `contracts/view-schemas/cartulary.view.timeline.v2.json:547-561` |
| `create_related.decision` | `workflow` | `view_row_create` / `view_row_create_route` | `editor` | true / false | `cartulary.view.decisions.v1`; `decision.support_refs:selected_record_id` | `contracts/view-schemas/cartulary.view.timeline.v2.json:576-590` |
| `create_related.evidence` | `workflow` | `view_row_create` / `view_row_create_route` | `editor` | true / false | `cartulary.view.evidence.v1`; no seed | `contracts/view-schemas/cartulary.view.timeline.v2.json:605-617` |
| `create_related.comm_log` | `workflow` | `view_row_create` / `view_row_create_route` | `editor` | true / false | `cartulary.view.comm_log.v1`; no seed | `contracts/view-schemas/cartulary.view.timeline.v2.json:627-639` |
| `create_related.handoff` | `workflow` | `view_row_create` / `view_row_create_route` | `editor` | true / false | `cartulary.view.handoff.v1`; no seed | `contracts/view-schemas/cartulary.view.timeline.v2.json:649-661` |
| `create_related.status_review` | `workflow` | `view_row_create` / `view_row_create_route` | `editor` | true / false | `cartulary.view.status_review.v1`; no seed | `contracts/view-schemas/cartulary.view.timeline.v2.json:671-683` |
| `create_related.lesson` | `workflow` | `view_row_create` / `view_row_create_route` | `editor` | true / false | `cartulary.view.lesson.v1`; no seed | `contracts/view-schemas/cartulary.view.timeline.v2.json:693-705` |

## 3. Responsibility Diagnosis for `TimelineWorkbook.tsx`

### Current Responsibilities

- Composes the Timeline workbook surface under `WorkbookSurfaceFrame` with grid, inspector, status strip, toolbar, notices, row actions, conflict resolver, evidence panel, and editor controls.
- Loads Timeline rows through `/api/v1/incidents/{incidentId}/views/{view_schema_id}/query`, builds query request state, validates `view_schema_id`, normalizes `view_row_v1` cells, reconciles committed rows with local drafts, and handles stale query races.
- Owns local row state, draft row creation, row freshness checks, committed row version tracking, and pending viewport continuity.
- Coordinates inline edit, collection edit, explicit row create, autosave, pending replay admission, retry, auth recovery probing, same-field conflict handling, save-state labels, and conflict resolution.
- Coordinates keyboard navigation, paste target resolution, active focus anchors, scroll/viewport continuity, grouped-row rejection, and draft-row paste creation through Cartulary grid-adapter abstractions.
- Coordinates WebSocket session lifecycle, hello/resume/ping/pong, presence snapshots/deltas, record-change handling, self-originated change suppression, sequence-gap refresh, and sparse live patches.
- Coordinates inspector selection, default-closed behavior, no-row state, row retargeting, history panel, rollback/delete/restore, mentions, evidence attach, and create-related feature groups.
- Re-exports several model helpers for existing shell/tests, creating a compatibility surface wider than the component itself.

### Likely Overreach

- Route orchestration for query, history, delete/restore/rollback, conflict resolution, mentions, evidence attach, create-related rows, and clipboard paste remains concentrated in one component.
- Pending replay scheduling and settlement are partially extracted but still deeply entangled with component refs, row mutation, refresh, and save-state presentation.
- WebSocket/live-update handling is partially modelized but effect wiring, row application, presence update, and refresh decisions still live in the component.
- Grid focus, keyboard, paste, inspector retargeting, and viewport restoration are cross-cutting enough to make single-slice refactors risky without characterization gates.
- Public re-exports from the component preserve legacy coupling that should not be removed inside behavior-preserving slices unless callers are migrated in a dedicated compatibility slice.

### Stable Public Contracts It Participates In

- `TimelineWorkbookProps` and existing re-exported helper/type names.
- Public `/api/v1/` and `/ws/v1/` routes used by query, row create/patch, conflict resolve, record actions, history, delete, restore, rollback, mentions, evidence, auth recovery, and clipboard paste.
- `cartulary.view.timeline.v2`, `view_schema_id`, field keys, feature group keys, route bindings, inspector config, visible/default/sort/filter/group field semantics.
- `record_id`, `row_version`, `base_row_version`, `field_key`, `client_txn_id`, `sheet_ref`, and `saved_view_id` where applicable.
- Selector and test-ID builders from `@cartulary/ui-contracts`.
- Grid-adapter Cartulary anchor semantics and the prohibition on direct `react-data-grid` app imports.
- Harness accounting through frontend phase maps, browser specs, generated ledgers, and Make targets.

### Risk Classification

High. The target is a 7065-line hot-path controller touching row identity, mutation envelopes, keyboard/paste behavior, pending saves, live collaboration, inspector actions, evidence, history, generated/contract-derived field keys, selectors, and browser-test-visible workflows.

### Seams That Appear Extractable

- History request/action orchestration around existing `useTimelineHistoryState` and `TimelineHistoryPanel`.
- Create-related inspector workflow controller around feature-group route bindings and generic-create helpers.
- Mention action orchestration around existing mention state/model helpers.
- Evidence attach orchestration around existing evidence service and Timeline payload builders.
- Grid keyboard/paste/focus orchestration around existing adapter and utility helpers.
- Pending replay controller around existing `WorkbookPendingQueueModel`, pending-save hook, and mutation service.
- WebSocket/live-update controller around existing collaboration message helpers and lifecycle service.
- Late-stage presentation cleanup for inline render helpers and local style objects after behavior-heavy seams are stable.

### Seams That Must Stay Local Until More Evidence Exists

- `TimelineWorkbookProps` and `WorkbookSurfaceFrame` composition.
- Route shapes, wire envelopes, feature-group route-binding decisions, field keys, `view_schema_id`, generated contract usage, and selector/test-ID usage.
- Cross-cutting focus/viewport continuity that spans keyboard, paste, history, evidence, mentions, conflicts, and live updates.
- Save-state labels, pending queue capacity, duplicate suppression, auth recovery, same-field conflict halting, and replay settlement.
- Any direct vendor coordinate semantics; `/apps/web` must continue to receive Cartulary adapter abstractions only.

## 4. Owner-Contract Map

| Contract surface that could drift | Current evidence | Owner / guardrail | Refactor risk |
| --- | --- | --- | --- |
| Grid-first row creation, inline edit, paste, keyboard/focus behavior | Target imports grid-adapter helpers; phase3/phase9 unit and browser tests; frontend guide grid-adapter boundary | Core 03 behavior, Core 01 routes/contracts, grid-adapter package boundary | High |
| `record_id` / `row_version` / `field_key` identity | Domain stable identifiers; Timeline view schema; phase3/4/6/7/9 tests | Core 01/Core 03/domain; app must key mutations by stable IDs | High |
| Active `view_schema_id` and Timeline field bindings | `timelineViewSchemaId`, `requireViewContract`, `cartulary.view.timeline.v2`, `workbookTimelineModel` bindings | Contract input and generated protocol-derived facade; no generated hand-edit | High |
| Query state, sorting, filtering, grouping, saved-view interaction | `workbookQuery`, `useWorkbookShellRuntime`, `workbookSavedViews`, `workbookSavedViewRuntime`, `ActiveSurfaceSavedViewSelector`, `WorkbookShell`, `savedViewSelector` prop, phase2/8/9 tests | Core 01 saved-view routes/startup/view-schema contracts; Core 03 saved-view/query/grouping behavior; Core 04 saved-view authorization and inspector ACs | Medium-high; keep saved-view lifecycle in shell/runtime and let Timeline consume controlled query/sheet state |
| Pending save, conflict, sync state | `workbookPendingQueue`, `useTimelinePendingSaves`, `timelineMutationRequests`, phase3/4/6 tests | Core 03 editing/conflict behavior and app-owned pending replay | High |
| Inspector default-closed and row retargeting | Timeline view schema `inspector_config.default_open=false`; `useTimelineInspectorSelection`; phase7/phase9 tests | Core 03 inspector behavior and view contract | High |
| Collaboration, presence, live updates | `workbookCollaborationMessages`, `workbookSocketLifecycle`, `useTimelineLiveUpdates`, phase6 tests/e2e | Core 01 WebSocket routes/envelopes; Core 03 collaboration behavior; Core 04 auth/origin security | High |
| Generated contract usage | `packages/protocol-ts/src/generated`, `packages/ui-contracts/src/generated`, `packages/view-contracts`, `packages/ui-contracts` | Generated artifact policy and contract owners | Medium-high |
| Harness/test accounting | `docs/testing-harness-nlspec.md`, frontend phase maps, browser specs, Make targets | Harness NLSpec owns mechanics; phase identity is evidence/accounting only | Medium |
| Visual/a11y evidence | Visual/a11y specs and frontend guide | Design/support evidence only unless Core 05 claim boundary is separately satisfied | Medium |

## 5. Refactor Tracker

| ID | Status | Objective | Dependency | Evidence | Exit condition |
| --- | --- | --- | --- | --- | --- |
| T-001 | DONE | Establish target-specific source baseline | AGENTS, framework, target file, current git state | Session header and inspected-file table above | Baseline records target, seam, commit, dirty state, framework path, and source limits |
| T-002 | DONE | Current-state repository scan | Target imports/exports, adjacent hooks/components/models/services/tests/packages | Current-state scan tables above | All inspected in-scope files are listed or grouped with source limits |
| T-003 | DONE | Module/package ownership inventory | Domain doc, frontend guide, dev guide, grid-adapter package, app source | Closed gap evidence, ownership notes, and boundary scan above | Owner citations, package seams, generated policy, and saved-view controller ownership are recorded |
| T-004 | TODO | Public contract freeze | Owner-contract map; route/query/pending/inspector/collab tests | Contract map above | Every drift-prone contract has explicit owner, evidence, and no-change rule |
| T-005 | TODO | Select first reviewable refactor slice | Candidate slices below | Slice table below | One slice chosen with prerequisites, validation, rollback, and stop condition |
| T-006 | TODO | Characterization test plan | Existing phase/unit/browser tests | Characterization plan below | Tests cover observed behavior before any production edit |
| T-007 | TODO | Boundary guardrail plan | Import-boundary policy, generated-artifact policy, selector owners | Guardrails below | Guardrails are runnable or explicitly TODO with discovery step |
| T-008 | TODO | Validation and harness accounting plan | Make targets, frontend guide, harness NLSpec | Validation plan below | Narrow-to-broad command sequence and failure classification are recorded |
| T-009 | TODO | Handoff and next-slice bootstrap | Filled session handoff and blank template | Handoff section below | Another agent can continue without rediscovering target state |

## 6. Workflow Dependency Map

| Workflow | Status | Dependency | Evidence | Exit condition |
| --- | --- | --- | --- | --- |
| WF-00 session/source bootstrap | DONE | AGENTS, git status, commit, framework path, target existence | Session header | Planner records exact checkout and source limits |
| WF-01 current-state repository scan | DONE | Target imports/exports/callers/tests/packages/docs | Scan tables | In-scope inspected files listed; unseen files marked TODO |
| WF-02 module/package ownership inventory | DONE | Frontend guide, dev guide, package imports, generated policy | Closed gap evidence, ownership notes, and boundary scan | Narrow line citations and saved-view ownership evidence are recorded |
| WF-03 public contract freeze | TODO | Owner-contract map | Contract map | All route, field, selector, grid, pending, inspector, collaboration, harness contracts frozen |
| WF-04 refactor slice selection | TODO | Tracker, risk diagnosis, characterization evidence | Candidate slices | Pick one small behavior-preserving slice and record stop condition |
| WF-05 characterization test plan | TODO | Existing phase/unit/browser tests | Test plan | Required behavior observations are named before production edits |
| WF-06 boundary guardrail plan | TODO | Import-boundary policy, generated policy, ui-contract ownership | Guardrails | Runnables and manual checks selected |
| WF-08 frontend package seam plan | TODO | `/apps/web`, `/packages/grid-adapter`, `/packages/view-contracts`, `/packages/ui-contracts` | Package seam notes | `/apps/web` keeps app/controller ownership and does not learn vendor semantics |
| WF-09 execution checkpoint plan | TODO | Slice selection and validation plan | Candidate slice stop conditions | Each risky move has checkpoint validation before continuing |
| WF-10 validation and harness accounting plan | TODO | Make targets, browser suites, frontend guide, harness NLSpec | Validation plan | Narrow-to-broad validation sequence and failure classification recorded |
| WF-11 documentation/generated-artifact plan | TODO | Generated policy, contract inputs, docs owner hierarchy | Generated artifact section | No generated hand-edit; docs-only changes use docs/harness checks |
| WF-12 cleanup and anti-drift plan | TODO | Boundary guardrails, validation, selector and generated checks | Guardrails and acceptance criteria | Refactor leaves no phase-shaped runtime dependency or stale compatibility bridge |
| WF-13 handoff and next-slice bootstrap | TODO | Handoff template | Handoff section | Future session can resume at selected workflow with evidence and TODOs |

## 7. Candidate Refactor Slices

| Slice ID | Objective | Files likely touched | Prerequisites | Public behavior expected unchanged | Characterization evidence required before edit | Validation command | Rollback note | Stop condition |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| TW-SL-01 | Extract history fetch/delete/restore/rollback orchestration into a hook/service | `TimelineWorkbook.tsx`; new `useTimelineHistoryActions.ts` or service under `apps/web/src/workbook/timeline` | WF-03, WF-05; confirm existing history route tests | History routes, envelopes, row retargeting, rollback target selection, inspector default-closed behavior | Phase7 history tests; phase9 inspector history/rollback tests; browser phase7 if route choreography changes | `make frontend-unit`; broaden to `make browser-e2e-webserver-backed` for browser route risk | Delete new hook/service and restore local callbacks | Stop if payload shape, stale request handling, or selected-row retarget semantics need behavior decisions |
| TW-SL-02 | Extract create-related inspector workflow controller | `TimelineWorkbook.tsx`; new `useTimelineCreateRelatedWorkflow.ts`; possibly model helper file | WF-03; 27-row Timeline feature-group inventory recorded above; characterize create-related tests before edit | Feature group route-binding guard, seed bindings, create target routes, evidence linkback route, selectors | Phase9 inspector create-related tests; generic create model tests | `make frontend-unit`; `make frontend-typecheck` | Reinline controller and keep local state names | Stop if evidence linkback behavior is unclear or implementation would invent behavior outside declared feature groups |
| TW-SL-03 | Extract mention action orchestration | `TimelineWorkbook.tsx`; new `useTimelineMentionActions.ts`; possibly mention service helper | Characterize resolve/dismiss/restore/create-host/create-identity behavior | Mention action routes, row mutation, dismissed persistence, auto-resolution notices, focus continuity | Phase4 support tests; phase5 mention chips/tests; phase9 inspector tests | `make frontend-unit`; browser phase9 inspector if focus/selection path changes | Reinline mention callbacks | Stop if row refresh, continuity, or undo behavior cannot be preserved with explicit inputs |
| TW-SL-04 | Extract evidence attach orchestration | `TimelineWorkbook.tsx`; new `useTimelineEvidenceAttach.ts` or service | Characterize evidence count/link behavior and errors | Evidence blob create/attach, Timeline row create/patch payloads, inspector messages, selectors | Phase5 evidence tests; `TimelineEvidencePanel.test.tsx`; phase9 inspector where relevant | `make frontend-unit`; browser evidence flow only if route flow changes | Reinline evidence functions | Stop if evidence route ownership or error-publication behavior is unclear |
| TW-SL-05 | Extract grid keyboard/paste anchor controller without vendor semantics | `TimelineWorkbook.tsx`; new grid interaction controller/model under `/apps/web`; no direct RDG imports | WF-06 guardrails; phase9 sentinel locked | Cartulary anchor semantics, paste target resolution, grouped/presentation row rejection, draft paste creation, focus restore | Phase9 sentinel; `GridAdapter.phase9.anchor.test.ts`; phase3 autosave/grid; browser phase9 keyboard/sentinel if needed | `make frontend-unit`; `make frontend-import-boundary-check`; browser phase9 if high risk | Reinline controller and remove new file | Stop if implementation needs vendor coordinates, DOM row indexes, or RDG internals |
| TW-SL-06 | Extract pending replay controller | `TimelineWorkbook.tsx`; new pending replay controller/hook around existing pending-save hook/model | Complete earlier lower-risk slices; characterization for queue and conflict states | Capacity 64, duplicate suppression, auth recovery, save labels, conflict halt, settlement order | Phase3 autosave; phase4 action sequencing/save state; phase6 tests; `workbookPendingQueue.test.ts` | `make frontend-unit`; `make browser-e2e-stateful` if auth/replay browser behavior changes | Reinline pending replay orchestration | Stop if refactor requires changing save-state labels, queue capacity, or conflict behavior |
| TW-SL-07 | Extract WebSocket/live update controller | `TimelineWorkbook.tsx`; new live-update controller/hook around existing lifecycle/message services | Characterize socket session, presence, sequence gap, live patches | WebSocket URL, hello/resume/ping/pong, presence, self-originated suppression, sequence-gap refresh | Phase6 tests; `workbookCollaborationMessages.test.ts`; `workbookSocketLifecycle.test.ts`; browser collaboration spec | `make frontend-unit`; `make browser-e2e-stateful` for collaboration risk | Reinline effect wiring | Stop if Core 04/auth/origin/session semantics are touched |
| TW-SL-08 | Move presentation-only render helpers/styles after behavior seams | `TimelineWorkbook.tsx`; existing/new presentational component files | Behavior-heavy slices stable; visual impact reviewed | No UI redesign, no selector/test-ID change, no layout behavior change | Unit tests plus visual/a11y only if rendered layout surfaces change | `make frontend-typecheck`; `make frontend-unit`; `make lint-biome`; `make browser-e2e-visual` only when visual surface changes | Reinline render helpers/styles | Stop if change becomes a redesign or requires selector churn |

## 8. Characterization Test Plan

### Existing Tests First

| Behavior to observe | Existing tests / fixtures | Gap |
| --- | --- | --- |
| Timeline query renders contract rows and preserves row identity | `WorkbookShell.phase3.grid.test.tsx`; `WorkbookShell.phase4.timelineQuery.test.tsx`; `timelineWorkbookTestSupport.ts` | No immediate gap for query/identity baseline |
| Grid-first row create, inline edit, autosave, blur/Enter/Tab, paste completion | `WorkbookShell.phase3.autosave.test.tsx`; `WorkbookShell.phase3.payload.test.tsx`; `WorkbookShell.phase3.grid.test.tsx` | Add only if selected slice moves untested branch |
| Keyboard/focus/paste sentinel behavior by `record_id` and `field_key` | `WorkbookShell.phase9.sentinel.test.tsx`; `GridAdapter.phase9.anchor.test.ts`; browser phase9 keyboard/sentinel | No grid extraction before these pass |
| Pending save/conflict/sync state | `WorkbookShell.phase4.saveState.test.tsx`; `WorkbookShell.phase6.test.tsx`; `workbookPendingQueue.test.ts` | TODO: confirm same-field conflict branch coverage before TW-SL-06 |
| Inspector default-closed, row retargeting, create-related, history/rollback | `WorkbookShell.phase7.test.tsx`; `WorkbookShell.phase9.inspector.test.tsx`; browser phase7/history and phase9 inspector | TODO: map exact tests to TW-SL-01 and TW-SL-02 before edits |
| Collaboration/presence/live updates | `WorkbookShell.phase6.test.tsx`; collaboration message/socket lifecycle tests; browser collaboration spec | No socket extraction before these pass |
| Evidence attach and evidence count | `WorkbookShell.phase5.test.tsx`; `TimelineEvidencePanel.test.tsx` | TODO: confirm attach error message cases before TW-SL-04 |
| Mention lifecycle and auto-resolution notices | `WorkbookShell.phase4.support.test.tsx`; `WorkbookShell.phase5.mentionChips.test.tsx`; inspector tests | TODO: confirm create-host/create-identity action branch coverage before TW-SL-03 |
| Selector/test-ID stability | UI-contract package tests/usage through phase suites | Do not add app-local selector literals |
| Visual and a11y posture | `workbook.visual.spec.ts`; `workbook.a11y.spec.ts`; `workbook.a11y-preflight.spec.ts` | Use only if rendered layout or accessibility surface changes; do not promote to product conformance without owner |

### TODO Tests Only If Needed

- TODO: Add slice-local characterization only when a selected extraction moves an observed behavior without existing direct coverage.
- TODO: Prefer behavior names in tests, for example "preserves selected history row after refresh", not implementation names such as "calls extracted hook".
- TODO: Do not add phase-shaped production structures to satisfy tests.

## 9. Boundary Guardrails

| Guardrail | Check | Expected result |
| --- | --- | --- |
| `/apps/web` must not learn `react-data-grid` vendor coordinate semantics | `make frontend-import-boundary-check`; code review for row/column index or RDG-specific imports in app extraction | Direct RDG imports remain only in `/packages/grid-adapter`; app uses Cartulary anchors |
| Direct `react-data-grid` imports remain inside `/packages/grid-adapter` | `rg -n "react-data-grid" apps packages tools docs --glob '!**/node_modules/**'`; authoritative Make target | Runtime direct import remains `packages/grid-adapter/src/index.tsx`; stylesheet import stays adapter-owned |
| Generated files are not hand-edited | `make generated-artifact-policy-check`; `git diff --name-only` for generated roots | No edits under `internal/gen`, `packages/protocol-ts/src/generated`, `packages/ui-contracts/src/generated`, generated topology files |
| Selectors/test IDs remain stable or are updated through owner package | Search for new app-local selector strings; review `@cartulary/ui-contracts` usage | No ad hoc selector/test-ID creation in extracted app code |
| Phase identity must not become production runtime structure | Review production code for `phase3`, `phase9`, frontend phase row IDs, or test-map assumptions | Phase labels stay in tests/harness/docs only |
| Visual/a11y evidence not promoted to product conformance without owner | Review docs/handoff wording and validation report | Visual/a11y remain design/support evidence unless Core 05 owner boundary is satisfied |
| Route shapes, wire envelopes, auth behavior, field keys stay frozen | Diff target and extracted files for route strings/payload builders/field-key literals | Behavior-preserving slices do not alter public contracts |

## 10. Validation Plan

### Cheapest Narrow Validations First

| Command | Use when | Failure classification |
| --- | --- | --- |
| `make frontend-typecheck` | Any TypeScript extraction or public type seam change | Product/code failure unless environment/tooling error is explicit |
| `make frontend-unit` | Any app/model/hook/service extraction | Product/code failure for behavior regressions; harness/config only if target cannot start or fixture setup fails unrelated to change |
| `make frontend-import-boundary-check` | Any grid, focus, keyboard, paste, or package-boundary-adjacent change | Product/boundary failure if direct import or forbidden dependency appears |
| `make lint-biome` | Any authored TS/TSX formatting or lint-sensitive extraction | Product/code hygiene unless tooling unavailable |
| `make generated-artifact-policy-check` | Any contract/package/generated-adjacent work or final refactor checkpoint | Product/process failure if generated roots changed by hand |

### Broader Validation By Risk

| Risk area | Broader command |
| --- | --- |
| Keyboard, paste, focus, grid sentinel | `make browser-e2e-webserver-backed` after frontend unit coverage; TODO: use `make explain-target TARGET=browser-e2e-webserver-backed DETAIL=summary` to confirm scenario selection before rerun |
| Pending replay, auth recovery, collaboration, live updates | `make browser-e2e-stateful` after phase6 unit coverage |
| Inspector, history, create-related, evidence | `make browser-e2e-webserver-backed` or targeted browser phase target discovered through `make explain-target` |
| Visual layout changes | `make browser-e2e-visual`; evidence remains design/support unless owner says otherwise |
| Accessibility surface changes | `make browser-e2e-a11y`; preflight only when phase map requires it |
| Broad end-of-run | `make agent-finalize`; then `make test-fast` or `make check` only when risk or owner process requires it |

### Unknown Command TODOs

- TODO: Use `make task-guide ROLE=feature-dev PHASE=phaseN` or frontend equivalent discovery before broad browser reruns.
- TODO: Use `make explain-target TARGET=<target> DETAIL=summary` before rerunning expensive browser suites.
- TODO: Distinguish product regressions from harness/config/infra failures in every handoff record.

## 11. Workstream Notes

### Scope and Evidence

- The target exists and is the only primary file.
- Adjacent files are in scope only through imports, props, hooks, route/query/mutation contracts, grid-adapter usage, view-schema/field-key contracts, inspector behavior, tests, generated contracts, selectors, and validation commands.
- Historical archive files were encountered in searches but are not current-state proof.

### Contracts and Docs

- `docs/domain.md` confirms stable vocabulary and identity concepts such as `record_id`, `row_version`, `base_row_version`, `field_key`, `view_schema_id`, `sheet_ref`, `saved_view_id`, and `client_txn_id`.
- Core 01/Core 03/Core 04 and the testing harness NLSpec are line-cited above for this tracker; implementation slices must refresh citations only if they touch a different owner surface.
- `docs/design.md` and visual/a11y specs are not product-conformance proof by themselves.

### Frontend Package Seam

- `/apps/web` owns workbook shell/controller behavior and consumes `@cartulary/grid-adapter`.
- `/packages/grid-adapter` owns direct `react-data-grid` imports, stylesheet import, vendor-event translation, and coordinate containment.
- `/packages/view-contracts` owns TypeScript-consumable view-schema adapters.
- `/packages/ui-contracts` owns selector and test-ID builders.

### Tests and Harness

- Existing unit/integration and browser coverage is broad; each slice still needs a pre-edit characterization checkpoint naming observed behavior.
- Use Make targets rather than raw `pnpm`, Vitest, Playwright, or Biome commands.
- Frontend phase maps and generated ledgers are accounting evidence only and must not shape production runtime code.

### Generated Artifacts

- Do not hand-edit generated roots declared in `tools/generated_artifact_policy.json`.
- If a contract-derived output drifts, update owner input first, then run the owning generator/drift target.
- This planner does not require generated changes.

### Risks and Blockers

- High risk: pending replay, conflict resolution, keyboard/paste/focus, WebSocket/live updates, inspector/history/evidence flows.
- Closed on 2026-06-28: saved-view interaction evidence, Core owner line citations, and exact Timeline feature-group inventory were added to this tracker.
- Remaining blockers are slice-specific characterization gaps only; no open evidence blocker prevents WF-03/WF-05 planning.

## 12. Session Handoff Template

### Filled Initial Handoff Record

| Field | Value |
| --- | --- |
| Date | 2026-06-28 |
| Branch / commit | Remediation refresh: `main...origin/main [ahead 2]` at `4d2dcd33027d0680218c36a52f20bc30545d36e3`; initial planning refresh saw `cf48bc2034d8a768e6bf630bfba5feb90ddb2a5c` |
| Dirty state at handoff | Final remediation status is `AM docs/handoffs/TimelineWorkbook.refactor-tracker.md`; the file was already staged as added, and this remediation adds unstaged updates to that same planner artifact only. Initial refresh saw `D docs/handoffs/workbook_refactor_tracker.md` and `?? docs/archive/workbook_refactor_tracker.md`, which were no longer dirty after the mid-session local commit. |
| Primary target | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx` |
| Framework used | `docs/handoffs/cartulary_modular_refactor_planning_framework.md` |
| Current workflow | WF-03 public contract freeze should be next, followed by WF-05 characterization test plan |
| Current findings | Target is a 7065-line high-risk Timeline controller owning query, row identity, pending replay, conflict, grid keyboard/paste/focus, WebSocket/presence, inspector/history/mentions/evidence/create-related composition and orchestration |
| Best first slice candidate | `TW-SL-01` history orchestration or `TW-SL-02` create-related workflow; choose only after public contract freeze and characterization checkpoint |
| Blockers / TODOs | Previously open evidence blockers are closed; remaining TODOs are future slice characterization and target-discovery steps only |
| Validation results | `git diff --check -- docs/handoffs/TimelineWorkbook.refactor-tracker.md` passed; `make lint-markdown` passed; `make generated-artifact-policy-check` passed with final run root `.cartulary/test-results/20260628T145910Z-p2417981` |
| Do not change | Production code, runtime behavior, UI design, routes, wire envelopes, generated files, field keys, view-schema IDs, selectors, phase maps, harness accounting |

### Blank Future Handoff Record

| Field | Value |
| --- | --- |
| Date | TODO: |
| Branch / commit | TODO: |
| Dirty state at handoff | TODO: |
| Primary target | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx` |
| Framework used | `docs/handoffs/cartulary_modular_refactor_planning_framework.md` |
| Workflow completed | TODO: |
| Workflow next | TODO: |
| Files inspected this session | TODO: |
| Files changed this session | TODO: |
| Contract decisions frozen | TODO: |
| Slice selected or implemented | TODO: |
| Characterization evidence | TODO: |
| Validation commands and results | TODO: |
| Product failures | TODO: |
| Harness/config/infra failures | TODO: |
| Remaining blockers | TODO: |
| Exact restart instructions | TODO: |

## 13. Binary Acceptance Criteria

| RF-AC ID | Status | Criterion | Evidence / exit condition |
| --- | --- | --- | --- |
| RF-AC-TW-001 | DONE | Exactly one primary target file and seam named | Session header names `TimelineWorkbook.tsx` and `/apps/web` Timeline workbook surface |
| RF-AC-TW-002 | DONE | All inspected in-scope files listed | Current-state scan groups inspected paths and marks source limits |
| RF-AC-TW-003 | DONE | All unseen relevant files marked | Source limits now close saved-view, Core line citation, generated-root, and contract-tail evidence for this planner; future slice source reads remain explicit |
| RF-AC-TW-004 | DONE | Every drift-prone public contract mapped | Owner-contract map covers required surfaces and cites owner evidence for query/saved-view, grid, identity, inspector, collaboration, generated, harness, and visual/a11y surfaces |
| RF-AC-TW-005 | DONE | Behavior-preserving refactors separated from behavior changes | Candidate slices are behavior-preserving and stop on behavior decisions |
| RF-AC-TW-006 | DONE | Characterization coverage stated | Existing tests and TODO test gaps listed by observed behavior |
| RF-AC-TW-007 | DONE | Checkpoint sequence with validation after risky moves | Workflow map, candidate slices, and validation plan define checkpoints |
| RF-AC-TW-008 | DONE | Frontend package boundaries preserved | Boundary guardrails require adapter-only RDG use and `make frontend-import-boundary-check` |
| RF-AC-TW-009 | DONE | No generated file hand-edit planned | Generated policy section and guardrails prohibit generated edits |
| RF-AC-TW-010 | DONE | No phase-shaped runtime dependency introduced | Boundary guardrails keep phase identity in tests/harness/docs only |
| RF-AC-TW-011 | DONE | Handoff sufficient for restart | Filled and blank handoff records included |
| RF-AC-TW-012 | DONE | Visual/a11y evidence not promoted into product conformance | Guardrails and validation plan classify visual/a11y as design/support unless owner says otherwise |
