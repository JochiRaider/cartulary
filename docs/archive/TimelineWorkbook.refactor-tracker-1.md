# TimelineWorkbook Refactor Tracker and Handoff Planner

## 1. Session Header

| Field | Value |
| --- | --- |
| Target file | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx` |
| Target module / package / seam | Frontend Timeline workbook surface inside `/apps/web`; consumes `/packages/grid-adapter`, `/packages/view-contracts`, `/packages/ui-contracts`, and generated protocol-derived contracts |
| Execution mode | Remediation control artifact for the TimelineWorkbook implementation slices; update after every completed workstream and before starting the next |
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
| Browser, visual, a11y tests | `apps/web/e2e/frontend.phase4.public-route.spec.ts`; `frontend.phase4.timeline-query.spec.ts`; `phase6.collaboration.spec.ts`; `phase7.history.spec.ts`; `phase9.keyboard.spec.ts`; `phase9.inspector-actions.spec.ts`; `phase9.sentinel.spec.ts`; `workbook.visual.spec.ts`; `workbook.a11y.spec.ts`; `workbook.a11y-preflight.spec.ts` | Search / scenario scan | Browser suites exist for route, collaboration, history, inspector, keyboard, sentinel, visual, and a11y paths; older `frontend.phase6/7/9.*` names are stale and must not be used |

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
| T-004 | READY | Public contract freeze | Owner-contract map; route/query/pending/inspector/collab tests | Contract map above; W0 remediation confirms no public-interface change unless Core owner contradiction is found | Every drift-prone contract has explicit owner, evidence, and no-change rule |
| T-005 | READY | Select first reviewable refactor slice | Candidate slices below | Slice table below; W2 is the first production slice after W1 characterization | One slice chosen with prerequisites, validation, rollback, and stop condition |
| T-006 | READY | Characterization test plan | Existing phase/unit/browser tests | Characterization plan below; W1 execution ledger must map exact scenario names before production edits | Tests cover observed behavior before any production edit |
| T-007 | READY | Boundary guardrail plan | Import-boundary policy, generated-artifact policy, selector owners | Guardrails below; actual e2e paths corrected in W0 | Guardrails are runnable or explicitly TODO with discovery step |
| T-008 | READY | Validation and harness accounting plan | Make targets, frontend guide, harness NLSpec | Validation plan below; W0 confirms Make-owned wrappers and target discovery are required | Narrow-to-broad command sequence and failure classification are recorded |
| T-009 | TODO | Handoff and next-slice bootstrap | Filled session handoff and blank template | Handoff section below | Another agent can continue without rediscovering target state |

## 6. Workflow Dependency Map

| Workflow | Status | Dependency | Evidence | Exit condition |
| --- | --- | --- | --- | --- |
| WF-00 session/source bootstrap | DONE | AGENTS, git status, commit, framework path, target existence | Session header | Planner records exact checkout and source limits |
| WF-01 current-state repository scan | DONE | Target imports/exports/callers/tests/packages/docs | Scan tables | In-scope inspected files listed; unseen files marked TODO |
| WF-02 module/package ownership inventory | DONE | Frontend guide, dev guide, package imports, generated policy | Closed gap evidence, ownership notes, and boundary scan | Narrow line citations and saved-view ownership evidence are recorded |
| WF-03 public contract freeze | READY | Owner-contract map | Contract map and W0 no-change rule | All route, field, selector, grid, pending, inspector, collaboration, harness contracts frozen |
| WF-04 refactor slice selection | READY | Tracker, risk diagnosis, characterization evidence | Candidate slices and W2-first sequencing | Pick one small behavior-preserving slice and record stop condition |
| WF-05 characterization test plan | READY | Existing phase/unit/browser tests | W1 characterization ledger must name exact scenarios before production edits | Required behavior observations are named before production edits |
| WF-06 boundary guardrail plan | READY | Import-boundary policy, generated policy, ui-contract ownership | Guardrails and corrected browser inventory | Runnables and manual checks selected |
| WF-08 frontend package seam plan | READY | `/apps/web`, `/packages/grid-adapter`, `/packages/view-contracts`, `/packages/ui-contracts` | Package seam notes | `/apps/web` keeps app/controller ownership and does not learn vendor semantics |
| WF-09 execution checkpoint plan | READY | Slice selection and validation plan | Candidate slice stop conditions and execution ledger below | Each risky move has checkpoint validation before continuing |
| WF-10 validation and harness accounting plan | READY | Make targets, browser suites, frontend guide, harness NLSpec | Validation plan with corrected e2e names | Narrow-to-broad validation sequence and failure classification recorded |
| WF-11 documentation/generated-artifact plan | READY | Generated policy, contract inputs, docs owner hierarchy | Generated artifact section | No generated hand-edit; docs-only changes use docs/harness checks |
| WF-12 cleanup and anti-drift plan | READY | Boundary guardrails, validation, selector and generated checks | Guardrails and acceptance criteria | Refactor leaves no phase-shaped runtime dependency or stale compatibility bridge |
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

## 7A. Execution Workstream Ledger

Each implementation workstream MUST update this ledger after validation and before the next workstream starts. The tracker is the controlling handoff artifact for this remediation.

| Workstream | Status | Tracker checkpoint | Files changed | Contract decision | Validation | Next step |
| --- | --- | --- | --- | --- | --- | --- |
| W0 tracker/spec-control cleanup | DONE | Corrected stale browser spec paths, promoted ready tracker rows, and recorded execution sequencing | `docs/handoffs/TimelineWorkbook.refactor-tracker.md` | No product contract change; public route, selector, field, generated, saved-view, auth, and inspector contracts remain frozen | `make lint-markdown` passed; `make generated-artifact-policy-check` passed with run root `.cartulary/test-results/20260628T150753Z-p2421872` | W1 characterization mapping |
| W1 characterization mapping | DONE | Mapped exact unit and browser scenario names to TW-SL-01..TW-SL-08 before production edits | `docs/handoffs/TimelineWorkbook.refactor-tracker.md` | No product contract change expected | `make explain-target TARGET=browser-e2e-webserver-backed DETAIL=summary` passed; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T150906Z-p2422784`; `make lint-markdown` passed | W2 history extraction |
| W2 history extraction | DONE | Added `useTimelineHistoryActions.ts` for history fetch, retarget refresh, delete/restore, rollback preview/action, and retained-row state; `TimelineWorkbook.tsx` now composes the hook and re-exports the rollback target helper for compatibility | `apps/web/src/workbook/timeline/hooks/useTimelineHistoryActions.ts`; `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`; `docs/handoffs/TimelineWorkbook.refactor-tracker.md` | No route/envelope/selector change; `/api/v1/records/{record_id}/history`, delete, restore, and rollback payload behavior preserved | `make frontend-typecheck` passed; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T151633Z-p2429486` | W3 create-related workflow |
| W3 create-related workflow extraction | DONE | Added `useTimelineCreateRelatedWorkflow.ts` for feature-group validation, seed binding, create-row submission, evidence linkback, workflow invalidation, draft updates, and cancellation; component keeps render-only workflow form | `apps/web/src/workbook/timeline/hooks/useTimelineCreateRelatedWorkflow.ts`; `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`; `docs/handoffs/TimelineWorkbook.refactor-tracker.md` | No generic workflow route or inspector-state record family; existing view-row create and Timeline patch routes preserved | `make frontend-typecheck` passed; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T152259Z-p2433510` | W4 mention action extraction |
| W4 mention action extraction | DONE | Added `useTimelineMentionActions.ts` for resolve, create-host/create-identity patch flow, dismiss, restore, dismissed mention state, entity refresh barriers, notices, and focus continuity; component delegates inspector mention actions to the hook | `apps/web/src/workbook/timeline/hooks/useTimelineMentionActions.ts`; `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`; `docs/handoffs/TimelineWorkbook.refactor-tracker.md` | No entity-mention route, row-version, selector, or focus-continuity contract change | Initial `make frontend-typecheck` failed on stale `InspectorMention` import, fixed; rerun passed. `make frontend-unit` passed with run root `.cartulary/test-results/20260628T153005Z-p2438045` | W5 evidence attach extraction |
| W5 evidence attach extraction | DONE | Added `useTimelineEvidenceAttach.ts` for evidence row create, object blob upload/attach, Timeline linkback, pending transaction handling, save state, and viewport continuity; added service-level attach-error sanitization and folded coverage into the existing mapped public-error test | `apps/web/src/workbook/timeline/hooks/useTimelineEvidenceAttach.ts`; `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`; `apps/web/src/services/workbookEvidence.ts`; `apps/web/src/services/workbookEvidence.test.ts`; `docs/handoffs/TimelineWorkbook.refactor-tracker.md` | No evidence route, selector, `view_schema_id`, or storage API contract change; caught attach errors now pass through the existing public-text filter before inspector display | `make frontend-typecheck` passed; initial `make frontend-unit` passed tests but failed accounting on one unmapped new test, fixed by folding assertions into mapped evidence public-error coverage; rerun passed with run root `.cartulary/test-results/20260628T153628Z-p2444232` | W6 grid anchor extraction |
| W6 grid anchor extraction | DONE | Added `useTimelineGridAnchorController.ts` for Cartulary `GridCellAnchor` lookup, tabular paste target resolution, selector-based focus restoration, and grid-adapter navigation; component keeps paste mutation execution but no longer owns adapter navigation/paste resolution helpers | `apps/web/src/workbook/timeline/hooks/useTimelineGridAnchorController.ts`; `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`; `docs/handoffs/TimelineWorkbook.refactor-tracker.md` | No direct `react-data-grid` app dependency, selector change, or vendor-coordinate contract; app layer continues to depend only on `@cartulary/grid-adapter` and UI selector contracts | Initial `make frontend-typecheck` failed on stale `timelineRowVersionTestId` import and unused hook return, fixed; rerun passed. `make frontend-import-boundary-check` passed with run root `.cartulary/test-results/20260628T154212Z-p2448520`; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T154220Z-p2448893` | W7 pending replay extraction |
| W7 pending replay extraction | DONE | Added `useTimelinePendingReplayController.ts` for replay timers, auth-recovery probing, admission/duplicate/refusal handling, FIFO drain, refresh/conflict blocking, dispatch settlement, metadata cleanup, and row-application continuation; component now supplies row/conflict/socket callbacks and consumes returned `enqueuePendingReplayUnit`/`schedulePendingReplay` | `apps/web/src/workbook/timeline/hooks/useTimelinePendingReplayController.ts`; `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`; `docs/handoffs/TimelineWorkbook.refactor-tracker.md` | No save-state label, queue capacity 64, replay identity, auth-recovery path, conflict halt, or pending socket transaction contract change | Initial `make frontend-typecheck` failed on stale component imports/type alias and exact-optional hook options, fixed; rerun passed. `make frontend-unit` passed with run root `.cartulary/test-results/20260628T155125Z-p2454116` | W8 live-update extraction |
| W8 live-update extraction | DONE | Added `useTimelineLiveUpdateController.ts` for presence send/debounce, socket connect/reconnect/cleanup, hello/resume, ping/pong, auth-close handling, auth recovery, presence snapshot/delta, self-origin suppression, sequence-gap refresh, sparse record patch application, and replay resume effects | `apps/web/src/workbook/timeline/hooks/useTimelineLiveUpdateController.ts`; `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`; `docs/handoffs/TimelineWorkbook.refactor-tracker.md` | No `/ws/v1` protocol, auth close, presence, origin suppression, sparse patch, sequence-gap, or refresh-block contract change | `make frontend-typecheck` passed; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T155727Z-p2457938`; `make browser-e2e-stateful` passed with run root `.cartulary/test-results/20260628T155750Z-p2459670` | W9 presentation/export cleanup |
| W9 presentation/export cleanup | DONE | Migrated the remaining rollback helper caller to `useTimelineHistoryActions.ts`; removed the helper/type re-export block from `TimelineWorkbook.tsx`; ran Biome safe formatting/import organization across touched frontend files; remaining imports from `TimelineWorkbook.tsx` are React component usage | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`; `apps/web/src/workbook/WorkbookShell.phase7.test.tsx`; formatted touched hook/test files; `docs/handoffs/TimelineWorkbook.refactor-tracker.md` | `TimelineWorkbook`, `TimelineWorkbookProps`, `SaveState`, and `IncidentRole` remain stable; model/history/pending/clipboard helpers are owned by their model/hook/service modules, not the component barrel | `make frontend-typecheck` passed; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T160404Z-p2479403`; initial `make lint-biome` failed on formatting/import organization and one extra hook dependency, fixed; rerun passed | W10 final validation |
| W10 final validation and handoff | DONE | Recorded final dirty state, validation run roots, skipped checks, failures fixed during the run, and restart instructions | `docs/handoffs/TimelineWorkbook.refactor-tracker.md` | No generated hand-edit; retained-run maintenance was not reused because `RESULTS_DIR` was unset | `make agent-finalize` passed with run root `.cartulary/test-results/20260628T160507Z-p2481337`; `make test-fast` passed with run root `.cartulary/test-results/20260628T160521Z-p2482418`; `make lint-markdown` passed after final tracker edits | Complete |

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

### Slice Characterization Map

| Slice | Unit characterization | Browser characterization | Pre-edit gap decision |
| --- | --- | --- | --- |
| TW-SL-01 history | `WorkbookShell.phase7.test.tsx`: opens row-centric history, retargets open history, clears stale rollback previews, builds rollback targets only from advertised actions/selectors, submits single-entry rollback, uses active/tombstone row versions; `WorkbookShell.phase9.inspector.test.tsx`: history and rollback preview/action public route contracts | `phase7.history.spec.ts`: E-7-01, E-7-01b, E-7-02, E-7-03, E-7-04, E-7-05; `phase9.inspector-actions.spec.ts`: FE-I-P9-01, FE-E-P9-01 | Covered; add tests only if extraction changes stale request or retarget state shape |
| TW-SL-02 create-related workflow | `WorkbookShell.phase9.inspector.test.tsx`: creates related Task Request from emitted seed bindings, creates related Evidence and links it back; `workbookInspectorModel.test.ts` and `packages/view-contracts/src/index.test.ts` validate feature-group keys/config | `phase9.sentinel.spec.ts`: FE-E-P9-03 create-related stays in workbook shell, FE-B-P10-01 coordination surfaces; `phase9.inspector-actions.spec.ts`: FE-E-P9-02 inspector config/default-closed behavior | Covered for Timeline Task Request/Evidence and browser all create-related surfaces; add table test only if hook introduces per-key branching |
| TW-SL-03 mentions | `WorkbookShell.phase4.support.test.tsx`: create-from-mention payload/action payloads, mention notices, resolve continuity, create-from-mention refresh continuity, undo row-version, dismiss/restore continuity; `WorkbookShell.phase5.mentionChips.test.tsx` chip state model | `phase4.mentions.resolve.spec.ts`: E-4-01 resolves and creates entities from Timeline mentions; `frontend.phase5.mention-lifecycle.spec.ts` covers lifecycle | Covered; add tests only for newly isolated branch helpers |
| TW-SL-04 evidence attach | `WorkbookShell.phase5.test.tsx`: evidence counts without navigation; `TimelineEvidencePanel.test.tsx`; `workbookEvidence.test.ts`: public errors, public handle routes, live regions, create/upload/attach row-version concurrency | `phase5.evidence.spec.ts`: E-5-01 through E-5-05; `phase9.inspector-actions.spec.ts`: evidence panel/action coverage | Covered; add tests only if extraction changes linkback or public-error helpers |
| TW-SL-05 grid/paste/focus | `WorkbookShell.phase9.sentinel.test.tsx`: Cartulary anchors, invalid-anchor clearing, Enter/Shift+Enter/Tab, sorted paste rows, scalar CSV, draft paste, multi-row paste, group conflict continuity, overflow create anchors, group/presentation rejection, vendor-coordinate rejection | `phase9.keyboard.spec.ts`: keyboard/clipboard/focus contracts; `phase9.sentinel.spec.ts`: large paste and grouped paste conflict scenarios; `frontend.phase4.public-route.spec.ts`: paste public route behavior | Covered; run import-boundary check after extraction |
| TW-SL-06 pending replay | `WorkbookShell.phase3.autosave.test.tsx`, `WorkbookShell.phase4.actionSequencing.test.tsx`, `WorkbookShell.phase4.saveState.test.tsx`, `WorkbookShell.phase6.test.tsx`, and `workbookPendingQueue.test.ts` cover labels, FIFO, capacity 64, retry, auth recovery, same-field conflict halt, settlement, coalescing, terminal failures | `phase6.collaboration.spec.ts`: FE-I-P7-01, FE-E-P7-01, E-6-05; `frontend.phase4.public-route.spec.ts`: pending/replay public route behavior | Covered; no queue behavior changes permitted |
| TW-SL-07 live updates | `WorkbookShell.phase6.test.tsx`: keyed presence, sparse live patches, conflict focus, resolver outcomes; `workbookCollaborationMessages.test.ts`; `workbookSocketLifecycle.test.ts` sequence gap, duplicate sequence, auth close, stale row requery | `phase6.collaboration.spec.ts`: FE-E-P7-01, E-6-01, E-6-02, E-6-04 | Covered; run stateful browser suite if socket effect wiring changes materially |
| TW-SL-08 presentation/export cleanup | Existing component render suites plus typecheck; visual/a11y only for layout or accessibility changes | `workbook.visual.spec.ts` and `workbook.a11y.spec.ts` are support/design evidence only | Covered for non-visual moves; broaden only if rendered structure changes |

### TODO Tests Only If Needed

- Add slice-local characterization only when a selected extraction moves an observed behavior without existing direct coverage.
- Prefer behavior names in tests, for example "preserves selected history row after refresh", not implementation names such as "calls extracted hook".
- Do not add phase-shaped production structures to satisfy tests.

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
| Keyboard, paste, focus, grid sentinel | `make browser-e2e-webserver-backed` after frontend unit coverage; use `make explain-target TARGET=browser-e2e-webserver-backed DETAIL=summary` to confirm current scenario selection before rerun |
| Pending replay, auth recovery, collaboration, live updates | `make browser-e2e-stateful` after phase6 unit coverage |
| Inspector, history, create-related, evidence | `make browser-e2e-webserver-backed` or targeted browser phase target discovered through `make explain-target`; current files include `phase7.history.spec.ts` and `phase9.inspector-actions.spec.ts` |
| Visual layout changes | `make browser-e2e-visual`; evidence remains design/support unless owner says otherwise |
| Accessibility surface changes | `make browser-e2e-a11y`; preflight only when phase map requires it |
| Broad end-of-run | `make agent-finalize`; then `make test-fast` or `make check` only when risk or owner process requires it |

### Unknown Command TODOs

- Use `make task-guide ROLE=feature-dev PHASE=phaseN` or frontend equivalent discovery before broad browser reruns.
- Use `make explain-target TARGET=<target> DETAIL=summary` before rerunning expensive browser suites.
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

### Filled Final Handoff Record

| Field | Value |
| --- | --- |
| Date | 2026-06-28 |
| Branch / commit | `main...origin/main [ahead 3]`; no commit was created by this remediation run |
| Dirty state at handoff | Modified: `apps/web/src/services/workbookEvidence.ts`, `apps/web/src/services/workbookEvidence.test.ts`, `apps/web/src/workbook/WorkbookShell.phase7.test.tsx`, `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`, and this tracker. Added: seven Timeline hooks under `apps/web/src/workbook/timeline/hooks/` for history, create-related, mention actions, evidence attach, grid anchors, pending replay, and live updates. |
| Primary target | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx` |
| Workflow completed | W0 through W10 remediation plan completed |
| Contract decisions frozen | No public route, `/ws/v1` protocol, `view_schema_id`, field key, record identity, row-version, selector, generated-root, saved-view, auth, or inspector feature-group contract changed |
| Files changed this session | `TimelineWorkbook.tsx`; `useTimelineHistoryActions.ts`; `useTimelineCreateRelatedWorkflow.ts`; `useTimelineMentionActions.ts`; `useTimelineEvidenceAttach.ts`; `useTimelineGridAnchorController.ts`; `useTimelinePendingReplayController.ts`; `useTimelineLiveUpdateController.ts`; `workbookEvidence.ts`; `workbookEvidence.test.ts`; `WorkbookShell.phase7.test.tsx`; tracker |
| Validation commands and results | Passed: `make lint-markdown`, `make generated-artifact-policy-check`, `make frontend-typecheck`, `make frontend-unit`, `make frontend-import-boundary-check`, `make lint-biome`, `make browser-e2e-stateful`, `make agent-finalize`, and `make test-fast`. Final broad run root: `.cartulary/test-results/20260628T160521Z-p2482418`; final tracker markdown lint also passed after handoff edits. |
| Product failures | None remaining. During execution, stale imports/type aliases, an unmapped standalone evidence test, formatting/import organization, and one extra hook dependency were fixed before final validation. |
| Harness/config/infra failures | None remaining. Retained successful-run maintenance was not reused because `RESULTS_DIR` was unset. |
| Skipped checks | `make check`, visual, and a11y suites were not run; W8 collaboration risk was covered by `make browser-e2e-stateful`, and final breadth was covered by `make test-fast`. |
| Remaining blockers | None known for this remediation plan. |
| Exact restart instructions | Review `git diff`, then run `make test-fast` or broader `make check` if additional non-Timeline changes are added before commit. Do not restart W0-W10 unless new contract/spec gaps are found. |

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
