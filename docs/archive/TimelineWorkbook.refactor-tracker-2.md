# TimelineWorkbook refactor tracker

## 1. Session header

| Field | Value |
| --- | --- |
| Session timestamp | 2026-06-28 13:00:07 EDT |
| Branch / commit | `main` / `57f6ccae139393b6d87a1fafd5e470dd97876762` |
| Dirty tree state | Pre-implementation scan was clean. WS-0 updates this tracker first; later workstreams may add authored frontend files only. |
| Target file | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx` |
| Target module / package / seam | Frontend Timeline workbook surface inside `/apps/web`; controller and UI composition seam for the Timeline workbook hot path. |
| Execution mode | Implementation run using this tracker as the controlling workstream artifact. Public routes, generated files, contracts, selectors, view schemas, and data migrations remain frozen unless a spec-first blocker is discovered. |
| Framework path used | `docs/handoffs/cartulary_modular_refactor_planning_framework.md` |
| Framework path rejected | `temp/current.md` exists, but its title is `Detailed patch plan: inspector workflow closure`; it is not the reusable modular refactor framework for this target. |
| Source limits and unseen files | Mechanical broad inventory across targeted Timeline/frontend contract/backend owner roots counted 218 files. Target is 4,919 lines. Target review remains slice-local by responsibility ranges. Backend route/security owner audit was completed for route-sensitive workstream planning; no public route or authorization change is planned. |

## 2. Current-state repository scan

### Inspected files table

| Area | Files inspected or path-listed | Evidence gathered |
| --- | --- | --- |
| Framework and owner docs | `temp/current.md`; `docs/handoffs/cartulary_modular_refactor_planning_framework.md`; `docs/domain.md`; `docs/design.md`; `docs/testing-harness-nlspec.md`; `docs/spec/01_architecture_storage_and_view_contracts.md`; `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`; `docs/spec/04_security_deployment_and_conformance.md` | Framework shape, authority posture, domain vocabulary, design/harness boundaries, workbook contract owners. |
| Target component | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx` | 4,919 lines; exports `SaveState`, `IncidentRole`, `TimelineWorkbookProps`, and `TimelineWorkbook`; high-risk hot-path controller. |
| Timeline components | `TimelineCellEditors.tsx`; `TimelineConflictResolver.tsx`; `TimelineEvidencePanel.tsx`; `TimelineGridSurface.tsx`; `TimelineHistoryPanel.tsx`; `TimelineMentionsPanel.tsx`; `TimelinePresenceMarkers.tsx`; `TimelineRowActions.tsx`; `TimelineWorkbookGrid.tsx`; `TimelineWorkbookInspector.tsx`; `TimelineWorkbookNotices.tsx` | Existing panel/grid/editor seams and remaining render-builder overlap with target. |
| Timeline hooks | `useTimelineCommittedRows.ts`; `useTimelineConflicts.ts`; `useTimelineCreateRelatedWorkflow.ts`; `useTimelineEvidenceActions.ts`; `useTimelineEvidenceAttach.ts`; `useTimelineGridAnchorController.ts`; `useTimelineGridInteractions.ts`; `useTimelineHistoryActions.ts`; `useTimelineHistoryState.ts`; `useTimelineInspectorSelection.ts`; `useTimelineLiveUpdateController.ts`; `useTimelineLiveUpdates.ts`; `useTimelineMentionActions.ts`; `useTimelineMentions.ts`; `useTimelinePendingReplayController.ts`; `useTimelinePendingSaves.ts`; `useTimelineRows.ts`; `useTimelineWorkbookRuntime.ts` | Existing state/effect seams for rows, conflicts, evidence, history, inspector selection, live updates, mentions, pending saves, grid interactions, and query runtime. |
| Timeline models and services | `timelineConflictModel.ts`; `timelineHistoryModel.ts`; `timelineRowsModel.ts`; `timelineViewportContinuityModel.ts`; `workbookMentionChips.ts`; `workbookTimelineModel.ts`; `timelineMutationRequests.ts`; `workbookCollaborationMessages.ts`; `workbookSocketLifecycle.ts` | Current helper ownership for row freshness, conflict parsing, bindings, payload builders, collaboration messages, and socket lifecycle. |
| Callers and test support | `apps/web/src/workbook/WorkbookShell.tsx`; `apps/web/src/testing/timelineWorkbookRenderTestSupport.tsx`; `apps/web/src/testing/timelineWorkbookTestSupport.ts`; `apps/web/src/testing/timelineWorkbookTestSupport.test.tsx` | Production caller and shared Timeline workbook render/fetch/socket helpers. |
| Existing characterization tests | `WorkbookShell.phase3.grid.test.tsx`; `WorkbookShell.phase3.autosave.test.tsx`; `WorkbookShell.phase3.payload.test.tsx`; `WorkbookShell.phase4.actionSequencing.test.tsx`; `WorkbookShell.phase4.saveState.test.tsx`; `WorkbookShell.phase4.support.test.tsx`; `WorkbookShell.phase4.timelineQuery.test.tsx`; `WorkbookShell.phase5.test.tsx`; `WorkbookShell.phase6.test.tsx`; `WorkbookShell.phase7.test.tsx`; `WorkbookShell.phase8.query.test.tsx`; `WorkbookShell.phase9.inspector.test.tsx`; `WorkbookShell.phase9.sentinel.test.tsx` | Phase-characterized workbook behavior across create/edit/paste/query/actions/evidence/collaboration/history/inspector/keyboard. |
| Unit tests for adjacent seams | `TimelineEvidencePanel.test.tsx`; `timelineConflictModel.test.ts`; `timelineRowsModel.test.ts`; `timelineViewportContinuityModel.test.ts`; `workbookTimelineModel.test.ts`; `workbookCollaborationMessages.test.ts`; `workbookSocketLifecycle.test.ts` | Existing lower-level coverage for evidence panel, conflict parsing, row freshness, continuity, binding/payload helpers, messages, and socket reducer. |
| Contract and generated policy inputs | `contracts/view-schemas/cartulary.view.timeline.v2.json`; `contracts/view-schemas/index.json`; `packages/view-contracts/src/index.ts`; `packages/view-contracts/src/index.test.ts`; `packages/ui-contracts/src/index.ts`; `packages/ui-contracts/src/index.test.ts`; `tools/generated_artifact_policy.json`; `tools/frontend_import_boundaries.json` | Timeline view-schema identity, inspector config, feature-group registry, selector builders, generated roots, and import boundary rules. |
| Grid adapter seam | `packages/grid-adapter/src/index.tsx`; `packages/grid-adapter/src/test-support.tsx`; `packages/grid-adapter/package.json` | Direct `react-data-grid` import and CSS singleton live in the adapter; app tests use adapter facade/test support. |

### Direct imports and imported symbols

| Import source | Imported symbols |
| --- | --- |
| `@cartulary/grid-adapter` | `GridCellAnchor`, `GridColumn`, `GridDensity`, `GridRow`, `reconcileRecordRows` |
| `@cartulary/ui-contracts` | `conflictMarkerTestId`, `dataTestIdSelector`, `draftCellTestId`, `draftRelationshipItemsTestId`, `draftTimelineCollectionInputTestId`, `genericCreateFieldTestId`, `genericCreateSubmitTestId`, `gridGroupRowTestId`, `gridRowGutterTestId`, `gridScrollportSelector`, `gridSortHeaderTestId`, `mentionItemTestId`, `relationshipItemsTestId`, `relationshipOverflowButtonTestId`, `rowCellTestId`, `timelineCollectionInputTestId`, `timelineInspectorSectionTestId`, `timelineMutationSubstrateReadyTestId`, `timelineScalarEditorTestId`, `WorkbookSurface` |
| `@cartulary/view-contracts` | `requireViewContract`, `resolveHeaderSortFieldKey`, `ViewContract` |
| `react` / `react-dom` | React types and hooks, `startTransition`, `flushSync` |
| App services/components/models/utils | `apiPath`, `fetchJSON`, `readEnvelope`, shared workbook controls, surface frame styles, query helpers, inspector config selector, startup/surface registry, continuity, keyboard, pending queue, presence, value formatting |
| Timeline hooks | committed rows, conflicts, create-related workflow, evidence actions/attach, grid anchor/controller/interactions, history actions/state, inspector selection, live update controller/updates, mention actions/mentions, pending replay/saves, rows, workbook runtime |
| Timeline models/services | conflict parser, row builder, viewport continuity helpers, mention chips, Timeline row/binding/payload helpers, mutation request builders, collaboration payload types, socket lifecycle |
| Timeline components | cell editors, conflict resolver, evidence panel, grid surface, history panel, presence markers, row actions, inspector, notices |

### Exported symbols and known callers

| Export | Known callers |
| --- | --- |
| `SaveState` | Public type from the target; no direct production caller found in scan. |
| `IncidentRole` | Public type from the target; used through `TimelineWorkbookProps` and tests passing `currentIncidentRole`. |
| `TimelineWorkbookProps` | Used by `timelineWorkbookRenderTestSupport.tsx` via `ComponentProps<typeof TimelineWorkbook>`. |
| `TimelineWorkbook` | Production: `WorkbookShell.tsx`. Tests/support: app tests, font role test, phase3/4/5/6/7/9 workbook tests, `timelineWorkbookRenderTestSupport.tsx`. |

### Adjacent hooks, components, and utilities

| Seam | Existing adjacent ownership |
| --- | --- |
| Grid facade | `TimelineGridSurface.tsx`, `TimelineWorkbookGrid.tsx`, `workbookGridFocus.tsx`, `workbookKeyboard.ts`, `@cartulary/grid-adapter`. |
| Rendering/editor pieces | `TimelineCellEditors.tsx`, `TimelinePresenceMarkers.tsx`, `TimelineWorkbookInspector.tsx`, `TimelineHistoryPanel.tsx`, `TimelineEvidencePanel.tsx`, `TimelineConflictResolver.tsx`, `TimelineWorkbookNotices.tsx`. |
| State/effect hooks | Existing hooks own many snapshots and commands, but `TimelineWorkbook.tsx` still coordinates route calls, pending-save submission, render builders, conflict resolver dispatch, query refresh, selection, and inspector lifecycle. |
| Models/services | `workbookTimelineModel.ts`, `timelineRowsModel.ts`, `timelineConflictModel.ts`, `timelineMutationRequests.ts`, `workbookCollaborationMessages.ts`, `workbookSocketLifecycle.ts`. |

### Tests and fixtures found

| Behavior area | Existing tests or fixtures |
| --- | --- |
| Grid row identity, draft preservation, stale query handling | `WorkbookShell.phase3.grid.test.tsx`; `timelineRowsModel.test.ts`; `workbookTimelineModel.test.ts` |
| Autosave, paste, payloads, save-state labels | `WorkbookShell.phase3.autosave.test.tsx`; `WorkbookShell.phase3.payload.test.tsx`; `workbookPendingQueue.test.ts` found by path scan |
| Action sequencing and relationship/mention support | `WorkbookShell.phase4.actionSequencing.test.tsx`; `WorkbookShell.phase4.support.test.tsx`; `WorkbookShell.phase4.saveState.test.tsx`; `WorkbookShell.phase4.timelineQuery.test.tsx` |
| Evidence counts and attach behavior | `WorkbookShell.phase5.test.tsx`; `TimelineEvidencePanel.test.tsx` |
| Presence, live updates, conflict resolution, pending replay | `WorkbookShell.phase6.test.tsx`; `workbookCollaborationMessages.test.ts`; `workbookSocketLifecycle.test.ts`; `timelineConflictModel.test.ts` |
| History, delete/restore, rollback | `WorkbookShell.phase7.test.tsx`; `timelineHistoryModel.ts` path scan |
| Query/sort/filter/group/saved-view behavior | `WorkbookShell.phase8.query.test.tsx`; `workbookQuery.test.ts` and saved-view model tests found by path scan |
| Inspector and keyboard/paste sentinels | `WorkbookShell.phase9.inspector.test.tsx`; `WorkbookShell.phase9.sentinel.test.tsx`; `GridAdapter.phase9.anchor.test.ts` found by path scan |
| Fixtures/helpers | `timelineWorkbookTestSupport.ts`; `timelineWorkbookRenderTestSupport.tsx`; route/fetch/socket helper functions and selector helpers. |

### Generated artifacts or contract-derived inputs found

| Input | Status |
| --- | --- |
| `tools/generated_artifact_policy.json` | Generated roots are `internal/gen`, `packages/protocol-ts/src/generated`, `packages/ui-contracts/src/generated`; generated file is `tools/task_surface.generated.mk`. Do not hand-edit. |
| `contracts/view-schemas/cartulary.view.timeline.v2.json` | Authored contract-derived input for Timeline surface identity, default sort/filter/group fields, zero-field create, and inspector config. |
| `packages/view-contracts/src/index.ts` | Contract adapter and inspector feature registry; includes exact Timeline feature-group list for `cartulary.view.timeline.v2`. |
| `packages/ui-contracts/src/index.ts` | Runtime-safe selectors and test IDs used by target and tests. Generated design tokens are downstream and must not be hand-edited. |
| `tools/frontend_import_boundaries.json` | Enforces adapter-only `react-data-grid` imports and singleton stylesheet ownership. |

### Validation commands discovered

| Command | Purpose | Status |
| --- | --- | --- |
| `make explain-target TARGET=frontend-unit DETAIL=summary` | Discover frontend unit target coverage and artifacts. | Discovered; not a behavior validation run. |
| `make explain-target TARGET=frontend-typecheck DETAIL=summary` | Discover frontend typecheck target. | Discovered; not executed as validation. |
| `make explain-target TARGET=frontend-import-boundary-check DETAIL=summary` | Discover boundary check target. | Discovered; not executed as validation. |
| `make task-guide ROLE=feature-dev PHASE=phase3` through `phase9` | Discover narrow phase validation by behavior surface. | Discovered. |
| `make generated-artifact-policy-check` | Docs/artifact safety check for generated policy. | TODO: run after artifact creation. |
| `make lint-markdown` | Markdown validation for authored docs. | TODO: run after artifact creation. |
| `make frontend-unit`, `make frontend-typecheck`, `make frontend-import-boundary-check` | Narrow implementation validations after future refactor slices. | TODO: run only after production edits. |

## 3. Responsibility diagnosis for `TimelineWorkbook.tsx`

### Current responsibilities

- Assembles the Timeline workbook runtime: query state, filter draft, active sheet reference, density, save-state labels, and row reload triggers.
- Owns row state and draft row reconciliation around query results, stale responses, local high-water row versions, and materialized drafts.
- Coordinates grid-first create, inline edit, blur/keyboard/paste commits, collection edits, row actions, and autosave/pending replay.
- Coordinates same-field conflict parsing, local conflict state, conflict marker anchors, resolver focus, and conflict resolution requests.
- Coordinates presence, WebSocket lifecycle effects, sparse live patches, self-originated invalidation suppression, and pending replay interaction.
- Coordinates inspector open/close/default-closed behavior, selected-row retargeting, relationship/evidence/history/workflow sections, and row-local actions.
- Builds or invokes route payloads for Timeline query, row create, record patch, clipboard paste, record actions, conflict resolution, evidence, mentions, and history.
- Renders the Timeline grid surface, scalar and collection editors, evidence count cells, inspector field editors, relationship editors, notices, context menu, conflict resolver, toolbar, and status strip.

### Likely overreach

- Inline render builders remain in the target: `renderTimelineScalarControl`, `renderTimelineGridEditor`, `renderTimelineInspectorEditor`, `renderTimelineCollectionInput`, inspector field/relationship/evidence/workflow/history section renderers.
- Orchestration remains concentrated in the target around `loadRows`, `freshTimelineRowsForQueryResult`, `applyRowMutation`, `waitForCommittedRecordIdle`, `queueScalarSave`, `queueCollectionSave`, `queueAction`, `handlePaste`, selection/context menu handlers, and `submitConflictResolution`.
- Route path assembly and request sequencing are still close to rendering, even though payload builders and some hooks already exist.
- The target remains a hot-path hub for query, pending replay, live updates, inspector, history, evidence, mentions, focus, paste, and rendering.

### Stable public contracts it participates in

- `view_schema_id = cartulary.view.timeline.v2`; stable Timeline field keys and inspector feature-group keys.
- `record_id`, `row_version`, `base_row_version`, `field_key`, and `client_txn_id` identity semantics.
- Grid-first creation, inline edit, paste, keyboard, focus, and selection behavior.
- Query state for sort/filter/group and saved-view interaction.
- Pending save, conflict, and sync state labels and details.
- Inspector default-closed, explicit activation, no-row state, and row retargeting.
- Collaboration/presence/live update behavior and self-originated message handling.
- Generated/derived contract usage through `@cartulary/view-contracts`, `@cartulary/ui-contracts`, and view-schema JSON.
- Harness/test accounting around phase maps and Make targets.

### Risk classification

| Area | Risk | Reason |
| --- | --- | --- |
| Overall target | High | Hot-path surface with create/edit/paste/query/pending/conflict/live/inspector behavior and broad phase coverage. |
| Presentation-only render extraction | Medium | Selector, focus, and editor callback wiring can drift even without route changes. |
| Pending save or query extraction | High | Row-version, coalescing, stale result, and save-state behavior are heavily characterized and user-visible. |
| Inspector/context/history extraction | High | Default-closed, row retargeting, history, rollback, and authorization-adjacent actions are contract-sensitive. |
| Style/constants extraction | Low to Medium | Low if exact movement only; medium if layout or tokens change. |

### Seams that appear extractable

- Grid and inspector render builders as Timeline-local presentational helpers.
- Autosave and pending mutation command coordination as a hook that receives current row/query/focus dependencies.
- Query/load/stale refresh coordination as a hook that preserves exact route paths and result freshness rules.
- Selection, row context menu, and inspector lifecycle as a hook that emits current callbacks and state.
- Conflict resolver coordination as a hook over existing conflict parser and pending queue anchors.
- Local style constants after behavioral seams are stable.

### Seams that must stay local until more evidence exists

- Route shape and envelope changes for create, patch, paste, actions, history, rollback, evidence, mentions, and conflict resolution.
- Authorization and `currentIncidentRole` behavior.
- View-schema field and inspector feature-group interpretation.
- Direct grid-vendor behavior or adapter coordinate semantics.
- Browser-only layout, visual, a11y, and focus behavior not covered by unit tests.
- Backend route implementations and generated contract regeneration.

## 4. Owner-contract map

| Drift-prone contract surface | Owner / evidence | Current target participation | Refactor guard |
| --- | --- | --- | --- |
| Grid-first row creation, inline edit, paste, keyboard/focus behavior | Core 03 workbook interaction; Core 01 clipboard/query contracts; phase3 and phase9 tests | Draft row creation, scalar/collection editors, `handlePaste`, keyboard mapping, focus anchors, grid selectors | Preserve grid-first behavior; reject group/presentation/vendor-coordinate-only paste targets; validate with phase3 and phase9. |
| `record_id` / `row_version` / `field_key` identity | `docs/domain.md`; Core 01/03; phase3 grid tests | Row identity, stale acceptance, local drafts, mutation payloads, conflict anchors | Never use visible row index, labels, or grid coordinates as mutation identity. |
| Active `view_schema_id` and Timeline field bindings | `contracts/view-schemas/cartulary.view.timeline.v2.json`; `@cartulary/view-contracts`; `workbookTimelineModel.ts` | `timelineContract`, `timelineVisibleBindings`, scalar/collection bindings, default sort/filter/group | Keep `cartulary.view.timeline.v2` and field keys unchanged; no view-schema edits. |
| Query state, sorting, filtering, grouping, saved-view interaction | Core 01/03; phase4 timeline query and phase8 query tests | `useTimelineWorkbookRuntime`, `buildQueryRequest`, `loadRows`, toolbar/query controls | Preserve request body, latest-result-wins behavior, active surface state, and saved-view semantics. |
| Pending save/conflict/sync state | Core 03; `workbookPendingQueue.ts`; phase3 autosave, phase4 save-state, phase6 tests | Pending replay queue runtime, save-state derivation, conflict anchors, queue notices | Preserve labels, capacity, coalescing, retry/halt, and same-field conflict extraction. |
| Inspector default-closed and row retargeting | Core 03; view-schema inspector config; phase9 inspector tests | Explicit inspector open, no-row state, active sheet reset, row refresh retarget | Preserve default closed; open only through explicit controls; keep record-bound retargeting. |
| Collaboration/presence/live updates | Core 03; `workbookCollaborationMessages.ts`; `workbookSocketLifecycle.ts`; phase6 tests | WebSocket lifecycle, presence drafts, sparse patches, invalidation/reload | Preserve session messages, self-origin suppression, sequence gaps, and pending replay interaction. |
| Generated contract usage | Generated artifact policy; view/ui contract packages | Imports selector builders and view contract helpers | Do not hand-edit generated roots; update owner inputs before generated outputs in future work. |
| Harness/test accounting | `docs/testing-harness-nlspec.md`; Make target discovery | Phase-slice and frontend unit targets establish characterization evidence | Use Make targets; distinguish product failures from harness/config/infra failures. |
| Route/security ownership | Core 01 route families, Core 03 interaction algorithms, Core 04 authorization model, backend route modules | Timeline query/create/patch/paste/conflict/action/evidence/mention/history code calls route-owned behavior | Frontend slices may reorganize call sites but must not change route paths, request envelopes, role assumptions, or server-derived authorization behavior. |

### Route and security owner map

| Surface | Owner docs and implementation owners | Required handling for this refactor |
| --- | --- | --- |
| Workbook query/create/patch/paste/conflict routes | Core 01 §3.3.3 through §3.3.6, Core 03 §3 and §15; `internal/modules/workbook/routes.go`, `internal/modules/timeline/*` | Preserve route paths, envelopes, `client_txn_id`, `base_row_version`, `field_key`, conflict, and idempotency semantics. |
| Timeline lifecycle actions | Core 03 §6 and §15; `internal/modules/workbook/routes.go`, `internal/modules/timeline/*` | Preserve reviewer/admin role expectations, row-version sequencing, and row-refresh behavior. |
| Evidence and mention actions | Core 01 evidence and mention route families, Core 03 §12 and §16; `internal/modules/evidence/*`, `internal/modules/entities/*`, `internal/modules/timeline/*` | Preserve source-preserving visible text, row-bound actions, and same-incident target validation assumptions. |
| History/delete/restore/rollback | Core 01 record route families, Core 03 inspector workflow, Core 04 acceptance criteria; `internal/modules/revisions/*` | Preserve inspector invalidation, selected-row retargeting, and server-authoritative route failure handling. |
| Authorization and incident roles | Core 04 §2; backend `requireIncidentRole` gates in workbook/entities/evidence/revisions modules | Inspector visibility remains presentation only; every route invocation must continue to rely on server authorization. |

## 5. Refactor tracker

| ID | Work item | Workstream | Status | Depends on | Owner | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- | --- |
| T-001 | Define Timeline target module and scope | scope | DONE | none | `/apps/web` Timeline workbook surface | Session header and target declaration | Exactly one primary file and one seam are explicit. |
| T-002 | Inspect current repo state | discovery | DONE | T-001 | `/apps/web` | Current-state scan, imports, callers, tests, contracts | Relevant files, imports, tests, generated paths, and commands are listed with source limits. |
| T-003 | Map owner contracts | contracts | DONE | T-002 | Core 01/03/04, domain, harness, design | Owner-contract map | Public behavior and owner docs are mapped for drift-prone surfaces. |
| T-004 | Freeze characterization evidence | tests | DONE | T-003 | Current implementer | Characterization test plan plus per-slice pre/post validation gate | Existing coverage is mapped; each implementation workstream must record pre/post Make targets before the next workstream starts. |
| T-005 | Plan boundary guardrails | architecture | DONE | T-003 | `/apps/web`, `grid-adapter`, generated policy | Boundary guardrails section | Import/generated/selector/phase/design guardrails are explicit. |
| T-006 | Plan behavior-preserving moves | implementation | DONE | T-004,T-005 | Current implementer | Candidate slice table | `TW-RS-01` through `TW-RS-06` are selected in dependency order with validation and stop conditions. |
| T-007 | Plan validation loop | validation | DONE | T-006 | Current implementer | Validation plan and per-workstream checklist | Cheapest sufficient validation commands are named; each slice records command results before proceeding. |
| T-008 | Update docs/contracts if required | docs | DEFERRED | T-003 | Owner docs and contract inputs | No change currently planned | Remains deferred unless a future slice requires owner-doc or contract input changes before codegen. |
| T-009 | Execute or hand off | handoff | DONE | T-006,T-007,T-008 | Current planning artifact, then next actor | Session handoff section | All planned workstreams completed with final validation and handoff evidence. |

Status values: `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`, `DROPPED`.

## 6. Workflow dependency map

| Workflow | Name | Class | Status | Required previous workflows | Required subsequent workflows | Target-specific evidence and exit |
| --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap | root | DONE | none | WF-01 | Branch, commit, dirty state, framework path, and target existence recorded. |
| WF-01 | Current-state repository scan | chain | DONE | WF-00 | WF-02, WF-03 | Imports, exports, callers, adjacent files, tests, generated policy, and validation commands scanned. |
| WF-02 | Module/package ownership inventory | chain | DONE | WF-01 | WF-04, WF-08 | `/apps/web` owns Timeline workbook controller; `/packages/grid-adapter`, view-contracts, and ui-contracts boundaries recorded. |
| WF-03 | Public contract freeze | chain | DONE | WF-01 | WF-04, WF-05, WF-06 | Drift-prone contracts, backend route owners, and Core 04 authorization boundaries are mapped for the planned frontend-only refactor. |
| WF-04 | Refactor slice selection | chain | DONE | WF-02, WF-03 | WF-05, WF-09 | Selected `TW-RS-01` through `TW-RS-06` in order, one implementation pass per workstream. |
| WF-05 | Characterization test plan | chain | DONE | WF-03, WF-04 | WF-09, WF-10 | Existing tests are mapped by behavior; each workstream must run and record its selected Make targets. |
| WF-06 | Boundary guardrail plan | parallel | DONE | WF-03 | WF-08, WF-10 | Grid vendor, generated, selector, phase, and design evidence guardrails documented. |
| WF-08 | Frontend package seam plan | chain | DONE | WF-02, WF-06 | WF-09 | App code remains on package facades; no direct grid vendor or generated-root edits are planned. |
| WF-09 | Execution checkpoint plan | chain | DONE | WF-04, WF-05, WF-08 | WF-10, WF-12 | Tracker was updated after each completed workstream and before starting the next. |
| WF-10 | Validation and harness accounting plan | chain | DONE | WF-05, WF-06, WF-09 | WF-13 | Use public Make targets; record command, result, and failure classification in the workstream log. |
| WF-11 | Documentation/generated-artifact plan | parallel | DONE | WF-03 | WF-12 | This artifact is the only planned doc change; generated files are excluded. |
| WF-12 | Cleanup and anti-drift plan | chain | DONE | WF-09, WF-11 | WF-13 | Dead helper/import cleanup and formatting/lint checks completed; no unrelated refactors. |
| WF-13 | Handoff and next-slice bootstrap | chain | DONE | WF-10, WF-12 | none | Final handoff record added with validation and skipped-check rationale. |

## 7. Candidate refactor slices

| Slice ID | Objective | Files likely touched | Prerequisites | Public behavior expected unchanged | Characterization evidence required before edit | Validation command | Rollback note | Stop condition |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| TW-RS-01 | Extract Timeline grid/inspector render builders into a Timeline-local presentational module. | `TimelineWorkbook.tsx`; new or adjacent render helper under `components/`; possibly no test file if only prop threading moves. | WF-03, WF-05; confirm selectors and editor callbacks. | Test IDs, focus, labels, column widths, collection activation/overflow, inspector sections, save behavior. | Phase3 grid/autosave, phase4 support, phase9 inspector/sentinel tests already identified; TODO: run relevant narrow tests before edit. | `make frontend-unit`; `make frontend-typecheck`; `make frontend-import-boundary-check`. | Revert helper file and restore inline render builders. | Stop if helper needs to own mutation, pending queue refs, route calls, or grid vendor details. |
| TW-RS-02 | Extract autosave and pending mutation command coordination into a hook. | `TimelineWorkbook.tsx`; new hook under `timeline/hooks`; maybe unit tests around pure command inputs. | TW-RS-01 optional; freeze pending queue contracts. | Client txn IDs, paths, payload intents, coalescing, pending signatures, focus continuity, base row version, save-state labels. | Phase3 autosave/grid/payload; phase4 action sequencing/save-state; phase6 pending queue tests. | `make frontend-unit`; targeted phase3/4/6 frontend unit tests; `make frontend-typecheck`. | Restore `queueScalarSave`, `queueCollectionSave`, and pending replay wiring to target. | Stop if extraction changes request envelopes, conflict admission, replay order, or save-state copy. |
| TW-RS-03 | Extract query/load/stale refresh coordinator. | `TimelineWorkbook.tsx`; new hook under `timeline/hooks`; possibly `workbookQuery` tests if helper boundaries shift. | Freeze query/saved-view contract and latest-result-wins behavior. | Query path/body, sort/filter/group, saved-view interaction, stale local high-water behavior, draft preservation, query errors, dismissed mention pruning. | Phase3 stale query/live patch tests; phase4 timeline query; phase8 query tests; workbook query model tests. | `make frontend-unit`; `make frontend-typecheck`; phase8 slice if behavior risk remains. | Restore `loadRows` and freshness coordination to target. | Stop if saved-view/query contract, route body, or active surface state changes. |
| TW-RS-04 | Extract selection, context menu, and inspector lifecycle hook. | `TimelineWorkbook.tsx`; new hook under `timeline/hooks`; possibly `TimelineRowActions.tsx` prop typing. | Freeze inspector default-closed and row-retargeting contracts. | Explicit inspector open, no-row state, active sheet reset, selected-row message, presence update, focus restore, context menu draft-row rejection. | Phase4 context menu/support; phase7 history retarget; phase9 inspector tests. | `make frontend-unit`; `make frontend-typecheck`; phase7/9 slice if route/focus risk remains. | Restore selection/context/inspector state and handlers to target. | Stop if hook persists inspector/saved-view state or changes panel availability. |
| TW-RS-05 | Extract conflict resolver coordinator. | `TimelineWorkbook.tsx`; new hook under `timeline/hooks`; maybe `timelineConflictModel` tests if parse ownership changes. | Freeze same-field conflict and pending replay unblock behavior. | Conflict payload, route path, resolver focus, save-state conflict labels, local draft merge, active conflict navigation, paste conflict state. | Phase6 conflict tests; phase7 conflict resolver/history interactions; phase9 grouped paste conflicts; conflict model tests. | `make frontend-unit`; `make frontend-typecheck`; phase6/9 slice if changed. | Restore resolver state and submit/clear handlers to target. | Stop if conflict detection semantics, server payload shape, or focus return changes. |
| TW-RS-06 | Extract local styles/constants only after behavior seams are stable. | `TimelineWorkbook.tsx`; new style/constants module. | TW-RS-01 through TW-RS-05 as applicable; ensure no behavior callbacks move. | Layout, tokens, density, typography, inspector dimensions, visual/a11y evidence classification. | Frontend unit; TODO: visual/a11y only if style movement changes layout risk. | `make frontend-unit`; `make frontend-typecheck`; TODO: `make browser-e2e-visual` only for layout-affecting changes. | Restore constants inline. | Stop if token/layout changes, UI redesign begins, or visual diff appears intentional. |

## 8. Characterization test plan

| Behavior to preserve | Existing evidence first | Missing or TODO evidence |
| --- | --- | --- |
| Timeline grid binds rows by stable `record_id` and `row_version` and cells by `field_key`. | `WorkbookShell.phase3.grid.test.tsx`; `timelineRowsModel.test.ts`; `workbookTimelineModel.test.ts`. | TODO: rerun before any row/query/focus extraction. |
| Draft create and inline edit survive refresh and do not depend on visible row index. | `WorkbookShell.phase3.grid.test.tsx`; `WorkbookShell.phase3.autosave.test.tsx`. | None identified for planning. |
| Autosave on Enter/Tab/blur/paste preserves exact save-state labels and conflict behavior. | `WorkbookShell.phase3.autosave.test.tsx`; `WorkbookShell.phase4.saveState.test.tsx`; `WorkbookShell.phase6.test.tsx`. | TODO: add targeted characterization only if extraction reveals untested callback ordering. |
| Query refresh, sort/filter/group, and saved-view interaction keep latest valid result and visible controls. | `WorkbookShell.phase3.grid.test.tsx`; `WorkbookShell.phase4.timelineQuery.test.tsx`; `WorkbookShell.phase8.query.test.tsx`; `workbookQuery.test.ts` found by scan. | TODO: inspect `WorkbookShell.phase8.query.test.tsx` line by line before query coordinator extraction. |
| Relationship/mention chips and auto-resolution notices preserve continuity. | `WorkbookShell.phase4.support.test.tsx`; `workbookMentionChips.ts` and related tests found by scan. | None identified for planning. |
| Evidence counts and evidence attach behavior stay row-bound. | `WorkbookShell.phase5.test.tsx`; `TimelineEvidencePanel.test.tsx`. | TODO: inspect attach hook tests before changing evidence section wiring. |
| Presence, sparse live patches, session revocation, and pending replay are stable. | `WorkbookShell.phase6.test.tsx`; `workbookCollaborationMessages.test.ts`; `workbookSocketLifecycle.test.ts`. | None identified for planning. |
| History, rollback preview/action, delete, and restore preserve public route contracts and row versions. | `WorkbookShell.phase7.test.tsx`; `WorkbookShell.phase9.inspector.test.tsx`. | TODO: inspect backend route tests before any route-sensitive history extraction. |
| Inspector default-closed, explicit activation, no-row state, retargeting, and section anchors are stable. | `WorkbookShell.phase9.inspector.test.tsx`; view-schema inspector config; `packages/view-contracts/src/index.ts`. | None identified for planning. |
| Keyboard, paste, and anchor behavior reject group/presentation/vendor-coordinate-only targets. | `WorkbookShell.phase9.sentinel.test.tsx`; `GridAdapter.phase9.anchor.test.ts` found by path scan. | None identified for planning. |
| Visual/a11y layout stays unchanged. | Design owner document and browser visual/a11y targets discovered. | TODO: run visual/a11y only if a slice changes layout, tokens, geometry, focus modality, or inspector overlay behavior. |

## 9. Boundary guardrails

- `/apps/web` must not learn `react-data-grid` vendor coordinate semantics; app code may use Cartulary grid anchors and the `@cartulary/grid-adapter` facade only.
- Direct `react-data-grid` imports and `react-data-grid/lib/styles.css` ownership must remain inside `/packages/grid-adapter`; preserve `make frontend-import-boundary-check`.
- Generated roots declared by `tools/generated_artifact_policy.json` must not be hand-edited: `internal/gen/**`, `packages/protocol-ts/src/generated/**`, `packages/ui-contracts/src/generated/**`, and `tools/task_surface.generated.mk`.
- Selectors and test IDs from `@cartulary/ui-contracts` must remain stable or be changed only through the owning package and contract/test updates.
- Phase identity must remain harness/accounting structure and must not become production runtime structure.
- Visual and accessibility evidence must not be promoted into product conformance unless an owner document establishes that boundary.
- Keep feature behavior out of `cmd/*`; this target is frontend `/apps/web` only.
- Do not alter public route shapes, wire envelopes, authorization behavior, view-schema IDs, field keys, generated contracts, saved-view envelopes, or harness accounting.

## 10. Validation plan

| Order | Command | Use | Failure classification |
| --- | --- | --- | --- |
| 1 | `make generated-artifact-policy-check` | Planning artifact safety and generated-root guardrail. | Harness/config failure if policy tooling fails without source relevance; product failure only if artifact points to illegal generated edits. |
| 2 | `make lint-markdown` | Authored Markdown validation. | Documentation formatting failure; not product runtime failure. |
| 3 | `make frontend-import-boundary-check` | Future implementation import guardrail around grid adapter/vendor imports. | Product boundary failure if new app import violates rule; harness/config failure if scanner infrastructure fails. |
| 4 | `make frontend-typecheck` | Future TypeScript seam validation. | Product/code failure if refactor causes type drift; config failure if toolchain unavailable. |
| 5 | `make frontend-unit` | Future broad frontend unit characterization. | Product/code failure if behavior tests fail; harness failure if phase artifact plumbing fails after tests pass. |
| 6 | `make phase-slice PHASE=phase3` | Future row/create/edit/paste/query/pending save changes. | Product failure for scenario regressions; service/harness failure if Postgres/object/browser stack fails unrelated to changed behavior. |
| 7 | `make phase-slice PHASE=phase4` through `phase9` as touched | Future evidence/mentions/collaboration/history/query/keyboard/inspector changes. | Product failure when touched behavior regresses; infrastructure failure when service/browser setup fails unrelated to changed code. |
| 8 | `make test-fast` | Broader local gate when implementation blast radius crosses multiple seams. | Product or harness classified by failing target summary. |
| 9 | `make check` | Full gate when review risk justifies browser stateful coverage. | Product or harness classified by scheduler summaries and run root. |
| 10 | `make agent-finalize` | End-of-run retained evidence and handoff hygiene before broad final verification. | Harness/reporting failure if finalize artifacts fail. |

## 11. Workstream notes

## 11A. Per-workstream contract-impact checklist

Every implementation workstream must fill this checklist in the workstream log before the next workstream starts.

| Checklist item | Required answer |
| --- | --- |
| Slice ID and behavior family | Name one `TW-RS-*` slice and the behavior family touched. |
| Owner contracts touched | Name Core/doc sections, route families, selectors, generated inputs, or state `none`. |
| Public compatibility | Confirm no route, envelope, selector, view-schema, generated artifact, or migration change; otherwise stop for spec-first remediation. |
| Backend validation needed | Name backend/service-backed targets when route/security behavior changes; otherwise state why frontend-only validation is sufficient. |
| Characterization before edit | Record Make targets run before edit, or record why retained evidence is not used. |
| Implementation summary | Describe the structural move and files changed. |
| Validation after edit | Record Make targets, results, and artifacts/run roots when available. |
| Failure classification | Classify failures as product, harness/config/infra, or unrelated pre-existing with evidence. |
| Tracker update | Confirm this tracker was updated before moving to the next workstream. |

## 11B. Workstream execution log

| Workstream | Status | Contract impact | Characterization before edit | Implementation and validation record | Next action |
| --- | --- | --- | --- | --- | --- |
| WS-0 Evidence and tracker cleanup | DONE | Documentation/test-process only; route/security owner map added. No public route, generated artifact, selector, view schema, or runtime change. | Discovery commands only; implementation characterization starts with WS-1. | Tracker baseline refreshed to `main` / `57f6ccae139393b6d87a1fafd5e470dd97876762`; broad targeted inventory counted 218 files; backend/security owner scan completed for planned route-sensitive slices; `make generated-artifact-policy-check` passed with run root `.cartulary/test-results/20260628T170139Z-p2598571`; `make lint-markdown` passed. | Start WS-1 `TW-RS-01` render builder extraction. |
| WS-1 `TW-RS-01` render builder extraction | DONE | Frontend-only internal refactor. No route, envelope, selector, view-schema, generated artifact, authorization, or migration change. Backend validation not needed because route behavior was untouched. | `make frontend-typecheck` passed; `make frontend-import-boundary-check` passed with run root `.cartulary/test-results/20260628T170207Z-p2599234`; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T170207Z-p2599263`. | Added `TimelineWorkbookRenderers.tsx` for scalar editor, collection editor, grid-column, and inspector editor render builders. `TimelineWorkbook.tsx` keeps state, callbacks, pending-save refs, route calls, and mutation sequencing. Post-edit `make frontend-typecheck` passed after fixing local type drift; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T171008Z-p2604497`; `make frontend-import-boundary-check` passed with run root `.cartulary/test-results/20260628T171008Z-p2604528`. `TimelineWorkbook.tsx` is now 4,279 lines; renderer module is 888 lines. | Start WS-2 `TW-RS-02` autosave and pending command coordination. |
| WS-2 `TW-RS-02` autosave and pending commands | DONE | Frontend-only internal refactor of existing command orchestration. Route paths, request envelopes, `client_txn_id`, `base_row_version`, coalescing, conflict admission, save-state labels, and server authorization behavior are unchanged. Backend validation not needed because route contracts were preserved. | `make frontend-typecheck` passed; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T171107Z-p2606809`; `make frontend-import-boundary-check` passed with run root `.cartulary/test-results/20260628T171107Z-p2606846`. | Added `useTimelineMutationCommands.ts` for `queueScalarSave`, `queueCollectionSave`, and `queueAction`. The hook receives existing refs/callbacks and keeps pending replay, payload builders, row-version sequencing, and route paths unchanged. Post-edit `make frontend-typecheck` passed after type-boundary fixes; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T171633Z-p2611711`; `make frontend-import-boundary-check` passed with run root `.cartulary/test-results/20260628T171633Z-p2611736`. `TimelineWorkbook.tsx` is now 3,894 lines; mutation-command hook is 534 lines. | Start WS-3 `TW-RS-03` query/load freshness coordination. |
| WS-3 `TW-RS-03` query/load freshness coordination | DONE | Frontend-only internal refactor of existing Timeline query/load orchestration. Query route path/body, saved-view query state, latest-result rejection, stale high-water retry, local draft preservation, dismissed mention pruning, and server authorization behavior are unchanged. Backend validation not needed because no route, envelope, generated contract, selector, or authorization assumption changed. | `make frontend-typecheck` passed; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T171728Z-p2614057`; `make frontend-import-boundary-check` passed with run root `.cartulary/test-results/20260628T171728Z-p2614076`. | Added `useTimelineRowsLoader.ts` for query path/body construction, result parsing, stale-result retry, local draft reconciliation, load/error state updates, viewport continuity advancement, and post-load save-state publication. Moved the original committed-row reconciliation semantics into the hook, including `reconcileRecordRows`, local pending signatures, collection drafts, scalar drafts, and draft-row materialization. Post-edit `make frontend-typecheck` first failed with run root `.cartulary/test-results/20260628T172328Z-p2618563` due hook boundary type mismatches; classified product/code and fixed by narrowing the viewport-continuity callback contract and accepting `undefined` row-version lookups. Rerun `make frontend-typecheck` passed; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T172400Z-p2619532`; `make frontend-import-boundary-check` passed with run root `.cartulary/test-results/20260628T172400Z-p2619538`. `TimelineWorkbook.tsx` is now 3,644 lines; row-loader hook is 458 lines. | Start WS-4 `TW-RS-04` selection, context menu, and inspector lifecycle. |
| WS-4 `TW-RS-04` selection, context menu, and inspector lifecycle | DONE | Frontend-only internal refactor of selected-row derivation, context-menu state, mention selection, inspector reset, selected-row disappearance, and inspector Escape handling. Stable selectors, row IDs, focus anchors, default-closed inspector behavior, route paths, envelopes, generated artifacts, view schema, and server authorization behavior are unchanged. Backend validation not needed because no route or security behavior changed. | `make frontend-typecheck` passed; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T172509Z-p2621910`; `make frontend-import-boundary-check` passed with run root `.cartulary/test-results/20260628T172509Z-p2621926`. | Expanded `useTimelineInspectorSelection.ts` to own row context menu state, active context row derivation, row selection/presence update commands, mention selection focus, selected-row disappearance retargeting, mention defaulting, reset-key cleanup, and inspector Escape cleanup. `TimelineWorkbook.tsx` now consumes hook snapshots/commands and keeps history/evidence/route actions in their existing hooks. Post-edit `make frontend-typecheck` passed; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T173253Z-p2627740`; `make frontend-import-boundary-check` passed with run root `.cartulary/test-results/20260628T173253Z-p2627762`. `TimelineWorkbook.tsx` is now 3,376 lines; inspector selection hook is 547 lines. | Start WS-5 `TW-RS-05` conflict resolver coordination. |
| WS-5 `TW-RS-05` conflict resolver coordination | DONE | Frontend-only internal refactor of conflict admission, active resolver derivation, paste conflict navigation, resolver focus, submit/clear flow, and pending replay unblock scheduling. Conflict route path, conflict-resolution payload, same-field conflict queue keys, row-version updates, scalar draft cleanup, and save-state labels are unchanged. Backend validation not needed because route/envelope/security semantics were preserved. | `make frontend-typecheck` passed; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T173355Z-p2630152`; `make frontend-import-boundary-check` passed with run root `.cartulary/test-results/20260628T173355Z-p2630186`. | Added `useTimelineConflictResolverCoordinator.ts` for `registerSameFieldConflict`, `handleMutationConflict`, active conflict snapshots, resolver Escape/focus effects, `clearLocalConflict`, `submitConflictResolution`, and merged-draft updates. `TimelineWorkbook.tsx` keeps the conflict state store and paste route flow but delegates resolver coordination. A scheduler ref breaks the previous pending-replay cycle without changing event-time behavior. Post-edit `make frontend-typecheck` first failed with run root `.cartulary/test-results/20260628T173759Z-p2633996` because the resolver still needed `activePasteConflictIndex`; classified product/code and fixed by returning that snapshot from the hook. Rerun `make frontend-typecheck` passed; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T173837Z-p2635011`; `make frontend-import-boundary-check` passed with run root `.cartulary/test-results/20260628T173837Z-p2635015`. `TimelineWorkbook.tsx` is now 3,098 lines; conflict coordinator hook is 437 lines. | Start WS-6 `TW-RS-06` styles and constants. |
| WS-6 `TW-RS-06` styles and constants | DONE | Frontend-only exact movement of Timeline workbook local style constants and row gutter width. CSS values, design tokens, layout, density, focus modality, route behavior, selectors, view schema, generated artifacts, and migrations are unchanged. Backend validation, visual, and a11y browser gates not needed because no geometry or accessibility behavior was intentionally changed. | `make frontend-typecheck` passed; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T173941Z-p2637400`; `make frontend-import-boundary-check` passed with run root `.cartulary/test-results/20260628T173941Z-p2637527`. | Added `TimelineWorkbookStyles.ts` and moved the used local style objects plus `timelineRowGutterWidth` into it. Post-edit `make frontend-typecheck` first failed with run root `.cartulary/test-results/20260628T174112Z-p2640311` because two workflow button styles were still referenced; classified product/code and fixed by exporting/importing those unchanged. Rerun `make frontend-typecheck` passed; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T174159Z-p2641356`; `make frontend-import-boundary-check` passed with run root `.cartulary/test-results/20260628T174159Z-p2641396`. `TimelineWorkbook.tsx` is now 3,004 lines; style module is 111 lines. | Start WS-7 validation and handoff completion. |
| WS-7 validation and handoff completion | DONE | Validation/documentation only after implementation slices. No public route, envelope, selector, view-schema, generated artifact, migration, design-token, or authorization change. | Retained per-slice evidence plus final post-format frontend checks. | Cleanup scan found no remaining component-owned conflict/style helpers. `make generated-artifact-policy-check` passed with run root `.cartulary/test-results/20260628T174351Z-p2644001`; `make lint-markdown` passed; `make lint-biome` first failed with run root `.cartulary/test-results/20260628T174351Z-p2644419` for formatting/import organization and was fixed with `make format` passing at `.cartulary/test-results/20260628T174410Z-p2645055`; rerun `make lint-biome` passed. Post-format `make frontend-typecheck` passed; `make frontend-unit` passed with run root `.cartulary/test-results/20260628T174433Z-p2647088`; `make frontend-import-boundary-check` passed with run root `.cartulary/test-results/20260628T174433Z-p2647118`. `make agent-finalize` passed with run root `.cartulary/test-results/20260628T174455Z-p2649565`; retained-run maintenance was skipped because `RESULTS_DIR` was unset. Broader `make test-fast` passed with run root `.cartulary/test-results/20260628T174507Z-p2650675` and 953 tests. Focus/a11y sanity `make browser-e2e-a11y-preflight` passed with run root `.cartulary/test-results/20260628T174654Z-p2680970` and 2 tests. Skipped `make check`, full browser visual, stateful, and service-specific browser suites because route/schema/generated contracts and layout geometry were unchanged, while `test-fast` and a11y preflight covered the refactor blast radius. Final `TimelineWorkbook.tsx` is 3,001 lines. | Effort complete; ready for review. |

### Scope and evidence notes

- Scope is exactly one primary file: `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`.
- Adjacent files are in scope only for imports, props, hooks, route/query/mutation contracts, grid-adapter usage, view-schema/field-key contracts, inspector behavior, tests, generated contracts, selectors, or validation commands.
- Full target review is performed slice-by-slice before each code move.
- Backend route implementation scan for planned route-sensitive frontend extraction is complete enough to identify owners; deeper backend inspection is required only if a slice changes route behavior.

### Contracts and docs notes

- Core 01 and Core 03 own most Timeline workbook behavior mapped here.
- Core 04 is relevant to authorization/security boundaries around incident role and route access; Core 04 §2 confirms inspector visibility is not authorization and routes re-derive current membership/role.
- `docs/domain.md` supplies vocabulary and identity boundaries.
- `docs/design.md` supplies design direction and token/layout constraints only; it is not Base Profile conformance evidence.
- `docs/testing-harness-nlspec.md` owns command invocation, scheduling, artifacts, cleanup, and gates.

### Frontend package seam notes

- `/apps/web` consumes grid behavior through `@cartulary/grid-adapter`; do not import `react-data-grid` directly.
- `@cartulary/view-contracts` and `@cartulary/ui-contracts` are contract-derived facade packages; keep selector and view-schema usage stable.
- Keep candidate helpers Timeline-local first unless a repeated cross-surface abstraction is proven.

### Tests and harness notes

- Existing coverage is strong but distributed across phase tests; future work must select narrow characterization by touched behavior.
- Use Make public targets, not direct `pnpm`, Vitest, Playwright, or raw script commands, for handoff-grade evidence.
- When a test command fails, record the target, run root or summary artifact, and whether failure appears product-related or harness/config/infra-related.

### Generated artifacts notes

- No generated file hand-edit is planned.
- Future contract changes, if any, must update owner inputs first and then run the appropriate generator/drift target.
- This refactor tracker is authored documentation, not generated output.

### Risks and blockers notes

- High risk: `TimelineWorkbook.tsx` currently couples rendering with query/load, pending save, conflict, live update, inspector, history, evidence, mention, and focus behavior.
- Implementation gate: each slice must run and record its selected characterization and validation commands before the following slice starts.
- Route-sensitive frontend slices may proceed only while preserving route paths, request envelopes, and server-derived authorization behavior.
- Exact broad inventory accounting was recomputed mechanically for targeted roots as 218 files.

## 12. Session handoff template

### Initial handoff record

| Field | Value |
| --- | --- |
| Date / actor | 2026-06-28 / Codex |
| Branch / commit | `main` / `217dfeb5fe18edae75bfef9329c99590c7f064dd` |
| Scope | Planning artifact for `TimelineWorkbook.tsx`; no production, test, generated, contract, selector, or runtime behavior edits. |
| Workflows completed | WF-00, WF-01, WF-02, WF-06, WF-11. |
| Workflows in progress | WF-03 public contract freeze; WF-13 handoff bootstrap. |
| Current findings | Target is a 4,919-line high-risk Timeline controller. Existing hooks/components/models already cover many seams, but render builders, query/load, pending save, selection/inspector lifecycle, and conflict resolver coordination remain concentrated in the target. |
| Evidence inspected | Framework, target imports/exports/symbols, adjacent Timeline files, phase tests, view/ui/grid contracts, generated policy, Make target guidance. |
| Validation run | Discovery commands only: `make explain-target ...` and `make task-guide ...`. TODO validation commands remain in section 10. |
| Blockers | No implementation characterization run; no backend route audit; full Core 04 and harness NLSpec not read line by line; exact manual inventory count should be recomputed if required. |
| Next recommended workflow | WF-03 Public Contract Freeze, then WF-04 Refactor Slice Selection. Start with TW-RS-01 as the lowest-risk first implementation slice. |

### Final handoff record

| Field | Value |
| --- | --- |
| Date / actor | 2026-06-28 / Codex |
| Branch / commit | `main` / `57f6ccae139393b6d87a1fafd5e470dd97876762` |
| Slice ID | WS-0 through WS-7 complete. |
| Files changed | `TimelineWorkbook.tsx`; `TimelineWorkbookRenderers.tsx`; `TimelineWorkbookStyles.ts`; `useTimelineMutationCommands.ts`; `useTimelineRowsLoader.ts`; `useTimelineInspectorSelection.ts`; `useTimelineConflictResolverCoordinator.ts`; this tracker. |
| Behavior intentionally unchanged | Public `TimelineWorkbook` props, `cartulary.view.timeline.v2`, selectors/test IDs, route paths, request/response envelopes, row identity, row-version semantics, pending queue behavior, conflict payloads, saved-view query semantics, inspector default-closed behavior, and server authorization assumptions. |
| Contract surfaces touched | Frontend internal ownership only; Core 01/03/04 route/security contracts, `docs/domain.md`, generated artifacts, view-schema inputs, and migrations unchanged. |
| Characterization run before edit | Each WS-1 through WS-6 recorded `make frontend-typecheck`, `make frontend-unit`, and `make frontend-import-boundary-check` before edits. |
| Implementation summary | Extracted render builders, mutation commands, row query loading, inspector/context lifecycle, conflict resolver coordination, and styles/constants into Timeline-local modules while keeping public contracts frozen. |
| Validation commands and results | Final post-format `make frontend-typecheck` passed; `make frontend-unit` passed at `.cartulary/test-results/20260628T174433Z-p2647088`; `make frontend-import-boundary-check` passed at `.cartulary/test-results/20260628T174433Z-p2647118`; `make generated-artifact-policy-check` passed at `.cartulary/test-results/20260628T174351Z-p2644001`; `make lint-markdown` passed; `make lint-biome` passed after `make format`; `make agent-finalize` passed at `.cartulary/test-results/20260628T174455Z-p2649565`; `make test-fast` passed at `.cartulary/test-results/20260628T174507Z-p2650675`; `make browser-e2e-a11y-preflight` passed at `.cartulary/test-results/20260628T174654Z-p2680970`. |
| Artifacts / run roots | See WS-0 through WS-7 execution log for per-slice run roots and failure classifications. |
| Product failures | Only local type/format diagnostics during extraction; all were fixed and rerun green. |
| Harness/config/infra failures | None unresolved. |
| Blockers or deferred evidence | `make check`, full browser visual, stateful, and service-specific browser suites skipped because no route/schema/generated contract or geometry change was made; `test-fast` plus browser a11y preflight were selected as sufficient broader gates. Retained-run maintenance skipped because `RESULTS_DIR` was unset for `make agent-finalize`. |
| Rollback notes | Revert the new Timeline-local modules and restore the corresponding inline blocks in `TimelineWorkbook.tsx`; no generated files, migrations, contracts, or lockfiles are involved. |
| Next recommended workflow | Review the extracted seams and keep future feature work in the new owner modules rather than re-expanding `TimelineWorkbook.tsx`. |

### Reusable future handoff blank

| Field | Value |
| --- | --- |
| Date / actor | TODO |
| Branch / commit | TODO |
| Slice ID | TODO |
| Files changed | TODO |
| Behavior intentionally unchanged | TODO |
| Contract surfaces touched | TODO |
| Characterization run before edit | TODO |
| Implementation summary | TODO |
| Validation commands and results | TODO |
| Artifacts / run roots | TODO |
| Product failures | TODO |
| Harness/config/infra failures | TODO |
| Blockers or deferred evidence | TODO |
| Rollback notes | TODO |
| Next recommended workflow | TODO |

## 13. Binary acceptance criteria

| RF-AC | Criterion | Status | Evidence / note |
| --- | --- | --- | --- |
| RF-AC-01 | Exactly one primary target file and seam named. | PASS | `TimelineWorkbook.tsx`; frontend Timeline workbook surface inside `/apps/web`. |
| RF-AC-02 | All inspected in-scope files listed. | PASS | Section 2 lists target, adjacent files, tests, contracts, and docs; broad archive hits are not in scope. |
| RF-AC-03 | All unseen relevant files marked. | PASS | TODOs identify full line audit, backend route audit, Core 04, harness NLSpec, and exact mechanical inventory count. |
| RF-AC-04 | Every drift-prone public contract mapped. | PASS | Section 4 covers grid, identity, view-schema, query, pending/conflict/sync, inspector, collaboration, generated contracts, and harness. |
| RF-AC-05 | Behavior-preserving refactors separated from behavior changes. | PASS | Candidate slices are internal and behavior-preserving; behavior changes are non-goals. |
| RF-AC-06 | Characterization coverage stated. | PASS | Section 8 maps existing tests first and marks TODO evidence only where needed. |
| RF-AC-07 | Checkpoint sequence with validation after risky moves. | PASS | Sections 6, 7, and 10 require one slice per pass and validation after edits. |
| RF-AC-08 | Frontend package boundaries preserved. | PASS | Section 9 protects `/apps/web`, `@cartulary/grid-adapter`, view contracts, and UI contracts. |
| RF-AC-09 | No generated file hand-edit planned. | PASS | Sections 9 and 11 explicitly exclude generated roots and generated files. |
| RF-AC-10 | No phase-shaped runtime dependency introduced. | PASS | Section 9 keeps phase identity in harness/accounting only. |
| RF-AC-11 | Handoff sufficient for restart. | PASS | Section 12 records findings, blockers, validation status, and next workflow. |
| RF-AC-12 | Target-specific owner-contract map includes harness/test accounting. | PASS | Section 4 includes generated contract usage and harness/test accounting. |
