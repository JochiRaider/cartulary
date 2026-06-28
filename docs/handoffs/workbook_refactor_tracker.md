# Workbook Refactor Tracker and Handoff Planner

## 1. Scope and source posture

Target directory: `apps/web/src/workbook`.

Resolved repo path: `/home/jochi/code/cartulary/apps/web/src/workbook`.

Path discrepancy: the prompt used `workbook/`; the live repo target is the longer repo-relative path `apps/web/src/workbook/`.

Candidate selection: a repo-wide scan excluding dependency, build, runtime-output, and temp directories found `apps/web/src/workbook` and `internal/modules/workbook`. The selected target is `apps/web/src/workbook` because it matches the requested path and contains `WorkbookShell.tsx`. The backend candidate `internal/modules/workbook` contains Go workbook module routes, stores, APIs, and tests; it is out of scope for this target-specific web tracker.

Planning-only status: this artifact is a refactoring tracker and handoff planner only. It does not authorize production-code edits, test edits, generated-file edits, package export changes, schema changes, config changes, harness-manifest changes, or behavior changes.

Source authority order used:

1. Core documents `docs/spec/00_document_set_status_and_precedence.md` through Core 04.
2. Core 05 only for claim-bearing timed, benchmark, fixture-sensitive, or publication evidence.
3. `docs/testing-harness-nlspec.md` for harness mechanics and Make-owned invocation.
4. `docs/domain.md` for workbook/domain vocabulary and concept boundaries.
5. `docs/design.md` for design-direction constraints only.
6. Adopted repo policy files such as `tools/generated_artifact_policy.json` and `tools/frontend_import_boundaries.json`.
7. Live repository code and tests for current implementation state.
8. Local modular refactor planning framework as planning structure only.

Framework source posture: `temp/current.md` exists, but it is an inspector-specific patch plan rather than the generic modular refactor framework. The generic framework was located at `docs/handoffs/cartulary_modular_refactor_planning_framework.md` and used as structure, not evidence of current state.

Source limits: no tests were run. Import, inventory, package, generated-policy, and Make-target claims below are based on static repo inspection and Make target explanation commands.

This artifact does not authorize behavior changes. Existing workbook behavior, test intent, route shapes, request payloads, UI selectors, saved-view/query semantics, shell layout, pending-save behavior, collaboration behavior, inspector behavior, and phase-accounting expectations remain stable unless a later implementation task explicitly authorizes change.

## 2. Current target inventory

| Path | Exists | Kind | Current responsibility | Target responsibility | Public/runtime/test/generated | Imports of concern | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `apps/web/src/workbook/` | yes | directory | Workbook shell, tests, submodules, timeline, utilities. | Target boundary for workbook shell and handoff to facades. | runtime/test, generated=no | Cross-root/timeline coupling present. | Selected target. |
| `apps/web/src/workbook/components/` | yes | directory | Workbook UI subcomponents. | Shell-adjacent components only. | runtime, generated=no | UI/grid/view facade usage in children. | Keep shell composition-oriented. |
| `apps/web/src/workbook/hooks/` | yes | directory | Workbook shell hooks and data orchestration. | Shell state/runtime seams. | runtime, generated=no | View-contract and timeline model coupling. | Candidate for staged extraction. |
| `apps/web/src/workbook/models/` | yes | directory | Workbook domain/view models. | Pure workbook models and facade adapters. | runtime/test, generated=no | View/protocol/grid facade types. | Good target for non-React seams. |
| `apps/web/src/workbook/timeline/` | yes | directory | Timeline workbook implementation. | Timeline-specific workbook surface and seams. | runtime/test, generated=no | Heavy root-workbook imports. | Contains largest responsibility concentration. |
| `apps/web/src/workbook/timeline/components/` | yes | directory | Timeline UI and grid components. | Timeline presentation components. | runtime/test, generated=no | Grid/UI/view facades, root utils. | `TimelineWorkbook.tsx` is oversized. |
| `apps/web/src/workbook/timeline/hooks/` | yes | directory | Timeline runtime hooks. | Timeline runtime orchestration seams. | runtime, generated=no | Root model/util imports. | Existing seam foundation. |
| `apps/web/src/workbook/timeline/models/` | yes | directory | Timeline pure models and tests. | Timeline model ownership. | runtime/test, generated=no | View-contract facade usage. | Keep pure where possible. |
| `apps/web/src/workbook/timeline/services/` | yes | directory | Timeline mutation/collaboration services. | Service adapters for timeline runtime. | runtime/test, generated=no | Route/payload semantics risk. | Preserve payload behavior. |
| `apps/web/src/workbook/utils/` | yes | directory | Workbook utilities and tests. | Shared workbook utilities only. | runtime/test, generated=no | Grid facade types in focus/keyboard. | Existing utility seams. |
| `apps/web/src/workbook/WorkbookShell.tsx` | yes | runtime entry | Shell coordinator; composes surfaces, saved-view/query controls, incident controls, generic surfaces, assessment/entity support, and Timeline entrypoints. | Thin shell coordinator delegating to hooks/models/facades. | runtime public entry, generated=no | Grid/UI/view facades; large shell exports used by tests. | 3069 lines. No direct generated or `react-data-grid` import found. |
| `apps/web/src/workbook/WorkbookShell.assessments.test.tsx` | yes | test | Assessment workbook behavior tests. | Characterization evidence. | test, generated=no | Runtime imports from `./WorkbookShell`; testing helpers. | Keep phase-free assessment traceability. |
| `apps/web/src/workbook/WorkbookShell.phase3.autosave.test.tsx` | yes | test | Phase 3 autosave and pending-save characterization. | Characterization evidence for autosave/pending seams. | test, generated=no | Grid test-support mock; runtime shell exports. | Do not reorganize only for phase filename. |
| `apps/web/src/workbook/WorkbookShell.phase3.grid.test.tsx` | yes | test | Phase 3 grid/create behavior characterization. | Characterization evidence for grid/create behavior. | test, generated=no | Imports `gridAdapterVendor`; asserts vendor string. | Test-only vendor awareness. |
| `apps/web/src/workbook/WorkbookShell.phase3.payload.test.tsx` | yes | test | Phase 3 request payload characterization. | Payload contract evidence. | test, generated=no | Grid test-support mock; fetch helpers. | Protect request JSON shapes. |
| `apps/web/src/workbook/WorkbookShell.phase4.actionSequencing.test.tsx` | yes | test | Phase 4 action sequencing tests. | Action sequencing characterization. | test, generated=no | Grid test-support mock; timeline helpers. | Protect ordering. |
| `apps/web/src/workbook/WorkbookShell.phase4.saveState.test.tsx` | yes | test | Phase 4 save-state tests. | Save-state characterization. | test, generated=no | Grid test-support mock; shell exports. | Protect pending/saved labels. |
| `apps/web/src/workbook/WorkbookShell.phase4.support.test.tsx` | yes | test | Phase 4 support tests for Timeline/workbook interactions. | Timeline/workbook interaction characterization. | test, generated=no | Runtime shell imports. | Support surface evidence. |
| `apps/web/src/workbook/WorkbookShell.phase4.timelineQuery.test.tsx` | yes | test | Timeline query and view-row normalization tests. | Query normalization characterization. | test, generated=no | Grid test-support mock; fetch helpers. | Protect query payload and row normalization. |
| `apps/web/src/workbook/WorkbookShell.phase5.gridProvenance.test.tsx` | yes | test | Grid provenance and contract-row tests for Phase 5 surfaces. | Grid provenance characterization. | test, generated=no | Grid test-support mock; UI selectors. | Protect provenance metadata. |
| `apps/web/src/workbook/WorkbookShell.phase5.mentionChips.test.ts` | yes | test | Mention chip model tests. | Model characterization. | test, generated=no | Timeline mention model. | Pure model evidence. |
| `apps/web/src/workbook/WorkbookShell.phase5.test.tsx` | yes | test | Phase 5 workbook/evidence behavior tests. | Evidence behavior characterization. | test, generated=no | Runtime shell imports; helpers. | Protect evidence panel/actions. |
| `apps/web/src/workbook/WorkbookShell.phase6.test.tsx` | yes | test | Phase 6 collaboration/session and pending behavior tests. | Collaboration characterization. | test, generated=no | Grid test-support mock; WebSocket helpers. | Protect session/pending behavior. |
| `apps/web/src/workbook/WorkbookShell.phase7.test.tsx` | yes | test | Phase 7 history and inspector behavior tests. | History/inspector characterization. | test, generated=no | Grid test-support mock; shell exports. | Protect inspector/history flow. |
| `apps/web/src/workbook/WorkbookShell.phase8.query.test.tsx` | yes | test | Phase 8 saved-view/query behavior tests. | Saved-view/query characterization. | test, generated=no | Grid test-support helpers; fetch helpers. | Protect saved view and query semantics. |
| `apps/web/src/workbook/WorkbookShell.phase9.inspector.test.tsx` | yes | test | Phase 9 inspector and row-local action tests. | Inspector/action characterization. | test, generated=no | Grid test-support mock; shell exports. | Protect row-local action UX. |
| `apps/web/src/workbook/WorkbookShell.phase9.sentinel.test.tsx` | yes | test | Phase 9 sentinel, focus, and continuity tests. | Focus/continuity characterization. | test, generated=no | Grid facade types; shell exports. | Protect keyboard/focus continuity. |
| `apps/web/src/workbook/WorkbookShell.surfaces.test.tsx` | yes | test | Multi-surface workbook shell tests. | Multi-surface characterization. | test, generated=no | Runtime shell imports; test title mentions generated data. | No generated import found. |
| `apps/web/src/workbook/components/ActiveSurfaceSavedViewSelector.tsx` | yes | runtime component | Active-surface saved-view selector. | Saved-view control component. | runtime, generated=no | UI contract facade. | Preserve selector/test ID behavior. |
| `apps/web/src/workbook/components/GenericMutationControl.tsx` | yes | runtime component | Generic mutation controls. | Generic surface control component. | runtime, generated=no | UI contract facade. | Keep mutation semantics external. |
| `apps/web/src/workbook/components/GenericWorkbookSurface.tsx` | yes | runtime component | Generic workbook grid surface. | Generic surface presentation/orchestration seam. | runtime, generated=no | Grid facade, protocol facade, UI/view facades. | 1719 lines; extraction candidate. |
| `apps/web/src/workbook/components/SystemViewSwitcher.tsx` | yes | runtime component | System-view switching UI. | System-view control component. | runtime, generated=no | UI/view contract facades. | Preserve `view_schema_id` semantics. |
| `apps/web/src/workbook/components/WorkbookGridControls.tsx` | yes | runtime component | Grid control UI. | Thin grid-control component. | runtime, generated=no | UI contract facade. | No vendor import found. |
| `apps/web/src/workbook/components/WorkbookInspectorFeatureGroups.tsx` | yes | runtime component | Inspector feature grouping UI. | Inspector presentation component. | runtime, generated=no | UI/view contract facades. | Preserve feature-group visibility. |
| `apps/web/src/workbook/components/WorkbookSheetToolbar.tsx` | yes | runtime component | Workbook toolbar. | Shell toolbar component. | runtime, generated=no | UI contract facade. | Preserve labels/selectors. |
| `apps/web/src/workbook/components/WorkbookShellSlots.tsx` | yes | runtime component | Shell layout slots. | Shell layout component. | runtime, generated=no | UI contract facade. | Layout behavior at risk. |
| `apps/web/src/workbook/components/WorkbookStatusStrip.tsx` | yes | runtime component | Status strip. | Save/status display component. | runtime, generated=no | UI contract facade. | Save-state wording risk. |
| `apps/web/src/workbook/components/WorkbookSurfaceFrame.tsx` | yes | runtime component | Surface frame/chrome. | Surface frame component. | runtime, generated=no | UI contract facade. | Preserve multi-surface layout. |
| `apps/web/src/workbook/hooks/useAssessmentSupportRows.ts` | yes | runtime hook | Assessment support rows. | Assessment support seam. | runtime, generated=no | View-contract facade. | Preserve assessment row semantics. |
| `apps/web/src/workbook/hooks/useEntityTimelinePreview.ts` | yes | runtime hook | Entity timeline preview. | Entity preview seam. | runtime, generated=no | Timeline/root model types. | Cross-subfolder coupling. |
| `apps/web/src/workbook/hooks/useGenericReferenceOptions.ts` | yes | runtime hook | Generic reference options. | Reference option seam. | runtime, generated=no | View-contract facade. | Good facade use. |
| `apps/web/src/workbook/hooks/useWorkbookIncidentIdentity.ts` | yes | runtime hook | Incident identity state. | Incident identity seam. | runtime, generated=no | Workbook model. | Preserve route/identity semantics. |
| `apps/web/src/workbook/hooks/useWorkbookResponsiveLayout.ts` | yes | runtime hook | Responsive shell layout state. | Layout seam. | runtime, generated=no | Workbook model. | Preserve responsive thresholds. |
| `apps/web/src/workbook/hooks/useWorkbookShellRuntime.ts` | yes | runtime hook | Workbook shell runtime, startup, saved views, query state. | Primary runtime orchestration seam. | runtime, generated=no | View-contract facade; app routes. | 873 lines; saved-view/query extraction candidate. |
| `apps/web/src/workbook/models/assessmentWorkbookModel.test.ts` | yes | test | Assessment model tests. | Model characterization. | test, generated=no | Local model imports. | Keep near model. |
| `apps/web/src/workbook/models/assessmentWorkbookModel.ts` | yes | runtime model | Assessment workbook model. | Assessment pure model. | runtime, generated=no | View-contract facade. | Preserve row mapping. |
| `apps/web/src/workbook/models/entityWorkbookModel.test.ts` | yes | test | Entity model tests. | Model characterization. | test, generated=no | Local model imports. | Keep near model. |
| `apps/web/src/workbook/models/entityWorkbookModel.ts` | yes | runtime model | Entity workbook model. | Entity pure model. | runtime, generated=no | Timeline/root type coupling. | Cross-surface seam candidate. |
| `apps/web/src/workbook/models/evidenceLifecycleViewModel.test.ts` | yes | test | Evidence lifecycle model tests. | Model characterization. | test, generated=no | Local model imports. | Keep near model. |
| `apps/web/src/workbook/models/evidenceLifecycleViewModel.ts` | yes | runtime model | Evidence lifecycle view model. | Evidence pure model. | runtime, generated=no | View-contract facade. | Preserve evidence states. |
| `apps/web/src/workbook/models/genericWorkbookModel.test.ts` | yes | test | Generic model tests. | Model characterization. | test, generated=no | Local model imports. | Payload evidence. |
| `apps/web/src/workbook/models/genericWorkbookModel.ts` | yes | runtime model | Generic create/patch/model helpers. | Generic pure model and payload builder. | runtime, generated=no | View-contract facade. | Payload-shape risk. |
| `apps/web/src/workbook/models/workbookContractRows.ts` | yes | runtime model | Contract-to-grid row helpers. | Facade-backed contract row adapter. | runtime, generated=no | Grid and view-contract facades. | Grid row seam candidate. |
| `apps/web/src/workbook/models/workbookDensity.test.ts` | yes | test | Density model tests. | Model characterization. | test, generated=no | Local model imports. | Keep near model. |
| `apps/web/src/workbook/models/workbookDensity.ts` | yes | runtime model | Density model. | Pure density model. | runtime, generated=no | Grid facade and protocol facade. | Protocol dependency posture question. |
| `apps/web/src/workbook/models/workbookIncidentIdentity.ts` | yes | runtime model | Incident identity derivation. | Pure identity model. | runtime, generated=no | None found. | Preserve route identity. |
| `apps/web/src/workbook/models/workbookInspectorModel.test.ts` | yes | test | Inspector model tests. | Model characterization. | test, generated=no | Local model imports. | Keep near model. |
| `apps/web/src/workbook/models/workbookInspectorModel.ts` | yes | runtime model | Inspector model. | Pure inspector state model. | runtime, generated=no | View-contract facade. | Inspector seam foundation. |
| `apps/web/src/workbook/models/workbookQuery.test.ts` | yes | test | Query model tests. | Query characterization. | test, generated=no | Local model imports. | Protect query JSON. |
| `apps/web/src/workbook/models/workbookQuery.ts` | yes | runtime model | Query JSON, saved-view query/layout helpers, normalization. | Query facade/model seam. | runtime, generated=no | View-contract facade. | 641 lines; duplicate-query seam candidate. |
| `apps/web/src/workbook/models/workbookReferenceOptions.test.ts` | yes | test | Reference option model tests. | Model characterization. | test, generated=no | Local model imports. | Keep near model. |
| `apps/web/src/workbook/models/workbookReferenceOptions.ts` | yes | runtime model | Reference option modeling. | Pure reference model. | runtime, generated=no | View-contract facade. | Preserve option labels. |
| `apps/web/src/workbook/models/workbookResponsiveLayout.test.ts` | yes | test | Responsive layout tests. | Layout characterization. | test, generated=no | Local model imports. | Preserve layout thresholds. |
| `apps/web/src/workbook/models/workbookResponsiveLayout.ts` | yes | runtime model | Responsive layout model. | Pure layout model. | runtime, generated=no | None found. | Keep pure. |
| `apps/web/src/workbook/models/workbookSavedViewRuntime.test.ts` | yes | test | Saved-view runtime tests. | Saved-view characterization. | test, generated=no | Local model imports. | Protect runtime selection. |
| `apps/web/src/workbook/models/workbookSavedViewRuntime.ts` | yes | runtime model | Saved-view runtime model. | Pure saved-view runtime seam. | runtime, generated=no | View-contract facade. | Saved-view seam foundation. |
| `apps/web/src/workbook/models/workbookSavedViews.test.ts` | yes | test | Saved-view model tests. | Saved-view characterization. | test, generated=no | Local model imports. | Preserve labels/defaults. |
| `apps/web/src/workbook/models/workbookSavedViews.ts` | yes | runtime model | Saved-view helpers. | Pure saved-view model. | runtime, generated=no | View-contract facade. | Preserve system/user view boundary. |
| `apps/web/src/workbook/models/workbookStartup.test.ts` | yes | test | Startup model tests. | Startup characterization. | test, generated=no | Local model imports. | Preserve initial surface behavior. |
| `apps/web/src/workbook/models/workbookStartup.ts` | yes | runtime model | Workbook startup model. | Pure startup model. | runtime, generated=no | View-contract facade. | Startup semantics risk. |
| `apps/web/src/workbook/models/workbookSurfaceRegistry.test.ts` | yes | test | Surface registry tests. | Registry characterization. | test, generated=no | Local model imports. | Preserve surface ids. |
| `apps/web/src/workbook/models/workbookSurfaceRegistry.ts` | yes | runtime model | Surface registry. | Surface registry model. | runtime, generated=no | View-contract facade. | Preserve `view_schema_id`. |
| `apps/web/src/workbook/timeline/components/TimelineCellEditors.tsx` | yes | runtime component | Timeline cell editors. | Timeline editor components. | runtime, generated=no | UI/view contracts. | Editing behavior risk. |
| `apps/web/src/workbook/timeline/components/TimelineConflictResolver.tsx` | yes | runtime component | Conflict resolution UI. | Conflict presentation component. | runtime, generated=no | UI contracts. | Preserve conflict actions. |
| `apps/web/src/workbook/timeline/components/TimelineEvidencePanel.test.tsx` | yes | test | Evidence panel tests. | Component characterization. | test, generated=no | Testing helpers; local component. | Evidence UI coverage. |
| `apps/web/src/workbook/timeline/components/TimelineEvidencePanel.tsx` | yes | runtime component | Timeline evidence panel. | Evidence presentation component. | runtime, generated=no | UI contracts. | Preserve evidence actions. |
| `apps/web/src/workbook/timeline/components/TimelineGridSurface.tsx` | yes | runtime component | Timeline grid surface. | Timeline grid presentation seam. | runtime, generated=no | Grid/UI facades. | No vendor import found. |
| `apps/web/src/workbook/timeline/components/TimelineHistoryPanel.tsx` | yes | runtime component | Timeline history panel. | History presentation component. | runtime, generated=no | UI contracts. | Preserve history UX. |
| `apps/web/src/workbook/timeline/components/TimelineMentionsPanel.tsx` | yes | runtime component | Mentions panel. | Mention presentation component. | runtime, generated=no | UI contracts. | Preserve mention chips. |
| `apps/web/src/workbook/timeline/components/TimelinePresenceMarkers.tsx` | yes | runtime component | Presence markers. | Presence presentation component. | runtime, generated=no | UI contracts. | Preserve collaboration UX. |
| `apps/web/src/workbook/timeline/components/TimelineRowActions.tsx` | yes | runtime component | Row-local actions. | Row action presentation component. | runtime, generated=no | UI contracts. | Inspector/action sequencing risk. |
| `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx` | yes | runtime component | Timeline workbook runtime, grid, pending saves, mutations, evidence, history, inspector, focus. | Decomposed timeline coordinator delegating to hooks/models/services. | runtime, generated=no | Grid/UI/view facades; extensive root-workbook imports. | 7093 lines; highest-risk extraction area. |
| `apps/web/src/workbook/timeline/components/TimelineWorkbookGrid.tsx` | yes | runtime component | Timeline grid wrapper. | Thin grid component. | runtime, generated=no | Grid/UI facades. | Preserve grid props. |
| `apps/web/src/workbook/timeline/components/TimelineWorkbookInspector.tsx` | yes | runtime component | Timeline inspector. | Inspector presentation component. | runtime, generated=no | UI/view contracts. | Preserve inspector feature groups. |
| `apps/web/src/workbook/timeline/components/TimelineWorkbookNotices.tsx` | yes | runtime component | Timeline notices. | Notice presentation component. | runtime, generated=no | UI contracts. | Preserve save/conflict notices. |
| `apps/web/src/workbook/timeline/hooks/useTimelineCommittedRows.ts` | yes | runtime hook | Committed row derivation. | Timeline row seam. | runtime, generated=no | Timeline models. | Preserve committed/pending split. |
| `apps/web/src/workbook/timeline/hooks/useTimelineConflicts.ts` | yes | runtime hook | Conflict state. | Conflict seam. | runtime, generated=no | Timeline models. | Preserve conflict lifecycle. |
| `apps/web/src/workbook/timeline/hooks/useTimelineEvidenceActions.ts` | yes | runtime hook | Evidence action state. | Evidence action seam. | runtime, generated=no | Timeline/root services. | Payload risk. |
| `apps/web/src/workbook/timeline/hooks/useTimelineGridInteractions.ts` | yes | runtime hook | Grid interaction state. | Grid-interaction seam. | runtime, generated=no | Grid facade. | Good facade use. |
| `apps/web/src/workbook/timeline/hooks/useTimelineHistoryState.ts` | yes | runtime hook | History state. | History seam. | runtime, generated=no | Timeline models. | Preserve history selection. |
| `apps/web/src/workbook/timeline/hooks/useTimelineInspectorSelection.ts` | yes | runtime hook | Inspector selection state. | Inspector selection seam. | runtime, generated=no | Timeline models. | Preserve default-closed behavior. |
| `apps/web/src/workbook/timeline/hooks/useTimelineLiveUpdates.ts` | yes | runtime hook | Live update handling. | Collaboration/live-update seam. | runtime, generated=no | Timeline services/models. | Preserve session behavior. |
| `apps/web/src/workbook/timeline/hooks/useTimelineMentions.ts` | yes | runtime hook | Mentions state. | Mention seam. | runtime, generated=no | Timeline models. | Preserve chip model. |
| `apps/web/src/workbook/timeline/hooks/useTimelinePendingSaves.ts` | yes | runtime hook | Pending-save state. | Autosave/pending seam. | runtime, generated=no | Pending queue utility. | Existing seam; still local orchestration in component. |
| `apps/web/src/workbook/timeline/hooks/useTimelineRows.ts` | yes | runtime hook | Timeline row state. | Row-state seam. | runtime, generated=no | Timeline models. | Preserve row normalization. |
| `apps/web/src/workbook/timeline/hooks/useTimelineWorkbookRuntime.ts` | yes | runtime hook | Timeline runtime wrapper. | Runtime coordination seam. | runtime, generated=no | Timeline hooks/models. | Candidate to absorb component logic. |
| `apps/web/src/workbook/timeline/models/timelineConflictModel.test.ts` | yes | test | Conflict model tests. | Model characterization. | test, generated=no | Local model imports. | Keep near model. |
| `apps/web/src/workbook/timeline/models/timelineConflictModel.ts` | yes | runtime model | Conflict model. | Pure conflict model. | runtime, generated=no | None found. | Preserve conflict status semantics. |
| `apps/web/src/workbook/timeline/models/timelineRowsModel.test.ts` | yes | test | Timeline row model tests. | Model characterization. | test, generated=no | Local model imports. | Preserve normalization. |
| `apps/web/src/workbook/timeline/models/timelineRowsModel.ts` | yes | runtime model | Timeline row model. | Pure row model. | runtime, generated=no | View-contract facade. | Row contract seam. |
| `apps/web/src/workbook/timeline/models/timelineViewportContinuityModel.test.ts` | yes | test | Viewport continuity model tests. | Continuity characterization. | test, generated=no | Local model imports. | Preserve focus continuity. |
| `apps/web/src/workbook/timeline/models/timelineViewportContinuityModel.ts` | yes | runtime model | Viewport continuity model. | Pure continuity model. | runtime, generated=no | None found. | Continuity seam foundation. |
| `apps/web/src/workbook/timeline/models/workbookMentionChips.ts` | yes | runtime model | Mention chip model. | Pure mention model. | runtime, generated=no | None found. | Covered by phase5 test. |
| `apps/web/src/workbook/timeline/models/workbookTimelineModel.test.ts` | yes | test | Timeline model tests. | Model characterization. | test, generated=no | Local model imports. | Payload/model evidence. |
| `apps/web/src/workbook/timeline/models/workbookTimelineModel.ts` | yes | runtime model | Timeline payload builders and row modeling. | Pure timeline model and payload builders. | runtime, generated=no | View-contract facade. | 897 lines; payload-shape risk. |
| `apps/web/src/workbook/timeline/services/timelineMutationRequests.ts` | yes | runtime service | Timeline mutation request helpers. | Mutation service adapter. | runtime, generated=no | App fetch/route semantics. | Preserve payload/routes. |
| `apps/web/src/workbook/timeline/services/workbookCollaborationMessages.test.ts` | yes | test | Collaboration message tests. | Service characterization. | test, generated=no | Local service imports. | Preserve wire message handling. |
| `apps/web/src/workbook/timeline/services/workbookCollaborationMessages.ts` | yes | runtime service | Collaboration message parsing/modeling. | Collaboration service model. | runtime, generated=no | None found. | Preserve socket payload semantics. |
| `apps/web/src/workbook/timeline/services/workbookSocketLifecycle.test.ts` | yes | test | Socket lifecycle tests. | Service characterization. | test, generated=no | Local service imports. | Preserve lifecycle behavior. |
| `apps/web/src/workbook/timeline/services/workbookSocketLifecycle.ts` | yes | runtime service | Socket lifecycle helpers. | Collaboration lifecycle service. | runtime, generated=no | None found. | Preserve reconnect/cleanup. |
| `apps/web/src/workbook/utils/GridAdapter.phase9.anchor.test.ts` | yes | test | Grid adapter anchor tests. | Grid anchor characterization. | test, generated=no | Grid facade types. | Test-only grid adapter contract. |
| `apps/web/src/workbook/utils/workbookClipboard.test.ts` | yes | test | Clipboard utility tests. | Utility characterization. | test, generated=no | Local utility imports. | Preserve paste detection. |
| `apps/web/src/workbook/utils/workbookClipboard.ts` | yes | runtime utility | Clipboard helpers. | Pure utility. | runtime, generated=no | None found. | Exported through shell tests. |
| `apps/web/src/workbook/utils/workbookContinuity.test.ts` | yes | test | Continuity utility tests. | Utility characterization. | test, generated=no | Local utility imports. | Preserve continuity behavior. |
| `apps/web/src/workbook/utils/workbookContinuity.ts` | yes | runtime utility | Continuity helpers. | Pure continuity utility. | runtime, generated=no | None found. | Focus/scroll risk. |
| `apps/web/src/workbook/utils/workbookGridFocus.tsx` | yes | runtime utility | Grid focus helpers. | Grid-focus utility via facade. | runtime, generated=no | Grid/UI facades. | No vendor import found. |
| `apps/web/src/workbook/utils/workbookKeyboard.test.ts` | yes | test | Keyboard utility tests. | Utility characterization. | test, generated=no | Local utility imports. | Preserve keyboard semantics. |
| `apps/web/src/workbook/utils/workbookKeyboard.ts` | yes | runtime utility | Keyboard helpers. | Pure keyboard utility. | runtime, generated=no | Grid facade type. | Preserve grid key behavior. |
| `apps/web/src/workbook/utils/workbookPendingQueue.test.ts` | yes | test | Pending queue tests. | Pending queue characterization. | test, generated=no | Local utility imports. | High-value autosave evidence. |
| `apps/web/src/workbook/utils/workbookPendingQueue.ts` | yes | runtime utility | Pending queue, replay, save-state, conflict parsing. | Autosave/pending core model. | runtime, generated=no | None found. | 1309 lines; critical seam. |
| `apps/web/src/workbook/utils/workbookPresence.ts` | yes | runtime utility | Presence helpers. | Pure presence utility. | runtime, generated=no | None found. | Collaboration UX risk. |
| `apps/web/src/workbook/utils/workbookStyles.ts` | yes | runtime utility | Workbook style helpers/classes. | Shared workbook styling utility. | runtime, generated=no | None found. | Preserve design tokens/selectors. |
| `apps/web/src/workbook/utils/workbookValueFormat.test.ts` | yes | test | Value formatting tests. | Utility characterization. | test, generated=no | Local utility imports. | Preserve display formatting. |
| `apps/web/src/workbook/utils/workbookValueFormat.ts` | yes | runtime utility | Value formatting helpers. | Pure formatting utility. | runtime, generated=no | None found. | Preserve visible text. |

## 3. Boundary contract for `workbook/`

`workbook/` owns shell composition, shell-level state orchestration, workbook surface routing, saved-view/query controls, incident controls, inspector mounting, and handoff to package facades.

`workbook/` must not own grid vendor mechanics that belong in `@cartulary/grid-adapter`.

`workbook/` must not import generated internals directly in runtime code when a package facade exists.

`workbook/` must not become the owner of protocol generation, UI-contract generation, view-schema generation, backend route semantics, authorization behavior, or persistence semantics.

`workbook/` tests may remain phase-characterization tests unless later repo inspection proves a better owner and an implementation task explicitly authorizes a test-support move.

Runtime code may use public package facades such as `@cartulary/grid-adapter`, `@cartulary/protocol-ts`, `@cartulary/ui-contracts`, and `@cartulary/view-contracts`. Generated roots remain downstream artifacts and must not be hand-edited.

## 4. Public behavior and contracts at risk

| Contract surface | Files involved | Existing test evidence | Risk if refactored poorly | Characterization needed | Status |
| --- | --- | --- | --- | --- | --- |
| Grid creation/editing | `WorkbookShell.tsx`, `GenericWorkbookSurface.tsx`, `TimelineWorkbook.tsx`, grid utilities/models | `WorkbookShell.phase3.grid.test.tsx`, `GridAdapter.phase9.anchor.test.ts` | Lost create-row behavior, wrong edit affordance, focus breakage. | Run targeted phase3 grid and phase9 anchor tests for grid slices. | Covered by characterization; not run. |
| Autosave/pending save | `TimelineWorkbook.tsx`, `useTimelinePendingSaves.ts`, `workbookPendingQueue.ts` | `WorkbookShell.phase3.autosave.test.tsx`, `workbookPendingQueue.test.ts` | Dropped pending writes, duplicate replays, incorrect save indicators. | Preserve pending queue signatures and replay semantics. | Covered by characterization; not run. |
| Payload shape | `workbookTimelineModel.ts`, `genericWorkbookModel.ts`, `timelineMutationRequests.ts`, `workbookQuery.ts` | `WorkbookShell.phase3.payload.test.tsx`, model tests | Backend route payload regressions. | Snapshot/request-body assertions before and after slice. | Covered by characterization; not run. |
| Action sequencing | `TimelineWorkbook.tsx`, pending queue, mutation services | `WorkbookShell.phase4.actionSequencing.test.tsx` | Out-of-order writes, stale row action state. | Run action sequencing after any queue or mutation seam. | Covered by characterization; not run. |
| Save state | `WorkbookStatusStrip.tsx`, `workbookPendingQueue.ts`, `TimelineWorkbook.tsx` | `WorkbookShell.phase4.saveState.test.tsx` | Incorrect saved/saving/error state. | Preserve state derivation labels and transitions. | Covered by characterization; not run. |
| Timeline query normalization | `workbookQuery.ts`, `TimelineWorkbook.tsx`, `timelineRowsModel.ts` | `WorkbookShell.phase4.timelineQuery.test.tsx`, `workbookQuery.test.ts` | Wrong row projection or query request JSON. | Keep query JSON and normalized rows stable. | Covered by characterization; not run. |
| Grid provenance | `workbookContractRows.ts`, `TimelineWorkbookGrid.tsx`, `GenericWorkbookSurface.tsx` | `WorkbookShell.phase5.gridProvenance.test.tsx` | Lost contract-row provenance or wrong cell metadata. | Run grid provenance after contract-row refactors. | Covered by characterization; not run. |
| Mention chips | `workbookMentionChips.ts`, `TimelineMentionsPanel.tsx`, `useTimelineMentions.ts` | `WorkbookShell.phase5.mentionChips.test.ts` | Broken mention rendering/counting. | Keep pure model test as guard. | Covered by characterization; not run. |
| Evidence behavior | `TimelineEvidencePanel.tsx`, `useTimelineEvidenceActions.ts`, `workbookTimelineModel.ts` | `WorkbookShell.phase5.test.tsx`, `TimelineEvidencePanel.test.tsx` | Broken attach/link/open evidence flows. | Preserve route payloads and visible actions. | Covered by characterization; not run. |
| Collaboration/session/pending behavior | `workbookCollaborationMessages.ts`, `workbookSocketLifecycle.ts`, `useTimelineLiveUpdates.ts`, `TimelineWorkbook.tsx` | `WorkbookShell.phase6.test.tsx`, service tests | Socket lifecycle or remote update regressions. | Run phase6 and service tests after collaboration seams. | Covered by characterization; not run. |
| History | `TimelineHistoryPanel.tsx`, `useTimelineHistoryState.ts`, `TimelineWorkbook.tsx` | `WorkbookShell.phase7.test.tsx` | Wrong history row, preview, or confirm flow. | Preserve row history request/preview/confirm contracts. | Covered by characterization; not run. |
| Inspector | `TimelineWorkbookInspector.tsx`, `workbookInspectorModel.ts`, `useTimelineInspectorSelection.ts`, `WorkbookInspectorFeatureGroups.tsx` | `WorkbookShell.phase7.test.tsx`, `WorkbookShell.phase9.inspector.test.tsx`, `workbookInspectorModel.test.ts` | Default-open/default-closed regressions, stale inspector selection. | Keep inspector model and row-local action tests. | Covered by characterization; not run. |
| Saved-view/query behavior | `useWorkbookShellRuntime.ts`, `workbookSavedViews.ts`, `workbookSavedViewRuntime.ts`, `workbookQuery.ts` | `WorkbookShell.phase8.query.test.tsx`, saved-view model tests | Wrong active view, query persistence, URL/query mismatch. | Run phase8 and saved-view model tests for query slices. | Covered by characterization; not run. |
| Sentinel/focus/continuity | `workbookGridFocus.tsx`, `workbookContinuity.ts`, `timelineViewportContinuityModel.ts`, `TimelineWorkbook.tsx` | `WorkbookShell.phase9.sentinel.test.tsx`, continuity/model tests | Lost focus restoration or viewport continuity. | Run phase9 sentinel plus utility/model tests. | Covered by characterization; not run. |
| Multi-surface behavior | `WorkbookShell.tsx`, `WorkbookSurfaceFrame.tsx`, `workbookSurfaceRegistry.ts`, `GenericWorkbookSurface.tsx` | `WorkbookShell.surfaces.test.tsx`, registry tests | Wrong surface routing or shell layout. | Run surfaces and registry tests after shell changes. | Covered by characterization; not run. |

## 5. Import and dependency findings

| Finding ID | File | Import or dependency | Classification | Risk | Proposed future action | Blocking? |
| --- | --- | --- | --- | --- | --- | --- |
| WB-IMP-01 | `apps/web/src/workbook/**` runtime files | Direct generated-internal imports | intentional | Static scan found no direct runtime imports from generated paths. | Keep using package facades; rerun `make frontend-import-boundary-check` after implementation. | no |
| WB-IMP-02 | `apps/web/src/workbook/**` runtime files | Direct `react-data-grid` imports | intentional | Static scan found no direct runtime vendor imports. | Preserve vendor isolation in `@cartulary/grid-adapter`. | no |
| WB-IMP-03 | Multiple runtime files | `@cartulary/grid-adapter` facade imports | defer | Facade imports are allowed, but grid mechanics remain spread across shell/timeline/generic surfaces. | Narrow grid row/focus/interaction seams without importing vendor APIs. | no |
| WB-IMP-04 | `GenericWorkbookSurface.tsx`, `workbookDensity.ts` | Runtime imports from `@cartulary/protocol-ts`; `apps/web/package.json` lists it under `devDependencies` | unknown | Runtime dependency posture may be inconsistent with actual usage. | Verify intended package dependency owner before moving exports or dependency declarations. | no |
| WB-IMP-05 | Multiple runtime files | `@cartulary/ui-contracts` and `@cartulary/view-contracts` facade imports | intentional | Facade usage is intended; risk is behavior drift if adapters move. | Keep imports on public package names and preserve selector/view contracts. | no |
| WB-IMP-06 | `WorkbookShell.tsx` and runtime files | Runtime imports from test helpers | intentional | Static scan found no runtime imports from `apps/web/src/testing`, Testing Library, or Vitest. | Preserve runtime/test-support separation. | no |
| WB-IMP-07 | Shell tests | Tests import runtime internals from `./WorkbookShell` and package test-support | defer | Moving exports can break characterization tests even when behavior is stable. | Add explicit test entrypoint plan before changing shell exports. | no |
| WB-IMP-08 | Root workbook and timeline files | Cross-workbook subfolder imports | should-fix | Timeline/root coupling makes safe extraction harder. | Introduce behavior-preserving seams around timeline runtime, shell runtime, and shared models. | no |
| WB-IMP-09 | `TimelineWorkbook.tsx`, `WorkbookShell.tsx`, `GenericWorkbookSurface.tsx` | Large responsibility concentration | should-fix | High regression risk from broad edits. | Use small slices around pure models/hooks before component moves. | no |
| WB-IMP-10 | `WorkbookShell.tsx`, `GenericWorkbookSurface.tsx`, `workbookContractRows.ts` | Grid-row/model builder logic spread across files | should-fix | Inconsistent provenance or grid row behavior. | Consolidate through existing model/facade boundaries after characterization. | no |
| WB-IMP-11 | `useWorkbookShellRuntime.ts`, `TimelineWorkbook.tsx`, `workbookQuery.ts` | Saved-view/query request and normalization logic split across shell and timeline | should-fix | Query JSON or active-surface behavior can drift. | Plan query seam around `workbookQuery.ts` while preserving phase8 evidence. | no |
| WB-IMP-12 | `TimelineWorkbook.tsx`, `useTimelinePendingSaves.ts`, `workbookPendingQueue.ts` | Pending/autosave setup split between model, hook, and component | should-fix | Duplicate queue orchestration can cause pending-save regressions. | Move only clearly modelable orchestration into hook/model seams. | no |
| WB-IMP-13 | `TimelineWorkbook.tsx`, inspector/history/focus utilities and hooks | Inspector/history/focus behavior concentrated in timeline component | should-fix | UX regressions in row-local action, history preview, focus continuity. | Extract after phase7 and phase9 characterization is green. | no |
| WB-IMP-14 | Workbook shell tests and `apps/web/src/testing/*` | Repeated shell/test setup and grid test-support mocks | defer | Test refactors can obscure characterization value. | Consolidate only after runtime seams stabilize. | no |
| WB-IMP-15 | `docs/archive/apps_web_src_workbook_refactor_tracker.md` | Prior archived tracker exists | defer | Archive may be stale if treated as current evidence. | Use only as historical support; live repo scan remains source of current state. | no |

## 6. Workstream tracker

| ID | Workstream | Status | Dependencies | Files / surfaces | Validation | Handoff notes |
| -- | ---------- | ------ | ------------ | ---------------- | ---------- | ------------- |
| WB-WS-01 | Scope and authority verification. | done | None | Core/docs/framework/domain/harness policy | Static source inspection only | `temp/current.md` discrepancy recorded. |
| WB-WS-02 | Inventory and import-boundary scan. | done | WB-WS-01 | `apps/web/src/workbook/**` | Static `rg`/file scan only | No direct runtime generated/vendor imports found. |
| WB-WS-03 | WorkbookShell responsibility decomposition analysis. | TODO | WB-WS-02 | `WorkbookShell.tsx`, shell components/hooks | Targeted tests before edits | Start with non-React model/hook seams. |
| WB-WS-04 | Grid facade usage and vendor-boundary cleanup planning. | TODO | WB-WS-02 | Grid-related components/utils/models | `make frontend-import-boundary-check`; grid tests | Keep vendor mechanics in `@cartulary/grid-adapter`. |
| WB-WS-05 | Protocol/UI/view-contract facade usage planning. | TODO | WB-WS-02 | Protocol/UI/view facade imports | Import-boundary check; typecheck | Verify `@cartulary/protocol-ts` dependency posture. |
| WB-WS-06 | Test characterization and fixture ownership planning. | TODO | WB-WS-02 | `WorkbookShell.*.test.*`, model tests, testing helpers | Narrow `make frontend-unit VITEST_FLAGS=...` | Preserve phase filenames unless later authorized. |
| WB-WS-07 | Autosave, pending, save-state, and action-sequencing seam planning. | TODO | WB-WS-03, WB-WS-06 | Pending queue, timeline runtime, save-state components | Phase3/4 tests and pending queue tests | Highest payload/order risk. |
| WB-WS-08 | Inspector, history, sentinel, focus, and continuity seam planning. | TODO | WB-WS-03, WB-WS-06 | Timeline inspector/history/focus utilities | Phase7/9 tests and model tests | Preserve default-closed and focus behavior. |
| WB-WS-09 | Saved-view/query and multi-surface seam planning. | TODO | WB-WS-03, WB-WS-06 | Runtime/query/surface registry | Phase8/surfaces/model tests | Preserve `view_schema_id` and query JSON. |
| WB-WS-10 | Final validation and implementation-handoff packaging. | done | WB-WS-01 through WB-WS-09 planning inputs | This artifact | No tests run | Ready as planning handoff; implementation still needs validation. |

## 7. Workflow dependency plan

Dependency graph:

`WF-WB-00 -> WF-WB-01 -> {WF-WB-02, WF-WB-03}`

`{WF-WB-02, WF-WB-03} -> {WF-WB-04, WF-WB-05}`

`{WF-WB-04, WF-WB-05} -> WF-WB-06 -> WF-WB-07 -> WF-WB-08 -> WF-WB-09`

| Workflow ID | Goal | Depends on | Enables | Inputs | Output | Validation | Stop condition |
| --- | --- | --- | --- | --- | --- | --- | --- |
| WF-WB-00 | Session setup and handoff initialization. | None | WF-WB-01 | User request, AGENTS instructions, repo root | Planning session posture and no-mutation guard | Confirm repo root and planning-only status | Framework/source posture recorded. |
| WF-WB-01 | Resolve actual `workbook/` path and owner docs. | WF-WB-00 | WF-WB-02, WF-WB-03 | `temp/current.md`, framework, domain, core, harness docs | Resolved target path and authority order | Static doc/path inspection | Target path and source limits recorded. |
| WF-WB-02 | Verify file inventory and subfolder structure. | WF-WB-01 | WF-WB-04, WF-WB-05 | `rg --files`, `find` under target | Complete target inventory | All supplied files represented; extra files classified | Inventory table has every supplied and discovered target file. |
| WF-WB-03 | Scan imports and generated-internal usage. | WF-WB-01 | WF-WB-04, WF-WB-06 | Import-boundary policy, package manifests, source imports | Import/dependency finding table | `rg` for facade/vendor/generated/test-helper terms | Direct runtime generated/vendor findings recorded. |
| WF-WB-04 | Map `WorkbookShell.tsx` responsibilities and extraction seams. | WF-WB-02, WF-WB-03 | WF-WB-06, WF-WB-07 | Shell, runtime hook, components, timeline coordinator | Responsibility and seam map | Static inspection plus existing test map | Shell seams can be sliced without deciding behavior changes. |
| WF-WB-05 | Map test coverage by behavior, phase, and contract surface. | WF-WB-02, WF-WB-03 | WF-WB-07, WF-WB-08 | `WorkbookShell.*.test.*`, model/service tests, test helpers | Behavior/test matrix | Narrow test commands discovered | Each contract surface has evidence or `TODO`. |
| WF-WB-06 | Identify facade-boundary refactor candidates. | WF-WB-04, WF-WB-05 | WF-WB-07 | Import findings, package facades, responsibility map | Facade-oriented candidate list | Import-boundary check planned | Candidates avoid generated/vendor direct imports. |
| WF-WB-07 | Define safe implementation slices. | WF-WB-06 | WF-WB-08 | Candidate list and behavior risk table | Behavior-preserving implementation slices | Each slice names required tests | Every slice has dependencies and risk. |
| WF-WB-08 | Define validation command matrix. | WF-WB-07 | WF-WB-09 | Make target explanations, package scripts, test paths | Validation matrix | Commands are discovered or marked `TODO: verify` | No invented commands remain unmarked. |
| WF-WB-09 | Prepare continuation handoff. | WF-WB-08 | Later implementation task | This tracker, handoff logs, questions/blockers | Implementation-ready handoff plan | Acceptance criteria checklist | Later agent can start without repeating planning scan. |

## 8. Proposed implementation slices for a later task

| Slice ID | Name | Depends on | Files likely touched | Change type | Behavior preserved | Required tests | Risk | Notes |
| -------- | ---- | ---------- | -------------------- | ----------- | ------------------ | -------------- | ---- | ----- |
| WB-SL-01 | Shell export and characterization guardrail | WF-WB-05 | `WorkbookShell.tsx`, shell tests only if authorized | Test-entrypoint planning | yes | Relevant existing shell tests | medium | First decide which exports are supported test seams before moving code. |
| WB-SL-02 | Grid-row/focus seam tightening | WB-SL-01 | `workbookContractRows.ts`, `workbookGridFocus.tsx`, grid surface components | Facade-boundary cleanup | yes | Phase3 grid, phase5 grid provenance, phase9 sentinel, anchor test | high | Do not import `react-data-grid`; keep facade imports only. |
| WB-SL-03 | Protocol/UI/view facade audit and dependency posture | WF-WB-03 | `apps/web/package.json` only if owner approves; runtime imports if needed | Boundary/dependency cleanup | yes, unless dependency declaration changes need owner approval | Typecheck, import-boundary check | medium | Verify `@cartulary/protocol-ts` runtime dependency status before editing package metadata. |
| WB-SL-04 | Pending/autosave runtime seam | WB-SL-01 | `TimelineWorkbook.tsx`, `useTimelinePendingSaves.ts`, `workbookPendingQueue.ts` | Responsibility extraction | yes | Phase3 autosave, phase4 action sequencing, phase4 save state, pending queue tests | high | Preserve replay order, conflict parsing, save-state labels. |
| WB-SL-05 | Timeline payload service seam | WB-SL-04 | `workbookTimelineModel.ts`, `timelineMutationRequests.ts`, `TimelineWorkbook.tsx` | Payload/helper extraction | yes | Phase3 payload, phase4 action sequencing, model tests | high | Payload shapes and route calls must remain byte-for-byte compatible where asserted. |
| WB-SL-06 | Inspector/history/focus seam | WB-SL-01 | `TimelineWorkbook.tsx`, inspector/history hooks, focus/continuity utilities | Responsibility extraction | yes | Phase7, phase9 inspector, phase9 sentinel, inspector/continuity model tests | high | Preserve default-closed inspector and row-local action behavior. |
| WB-SL-07 | Saved-view/query runtime seam | WB-SL-01 | `useWorkbookShellRuntime.ts`, `workbookQuery.ts`, saved-view models, `TimelineWorkbook.tsx` | Query seam consolidation | yes | Phase8 query, phase4 timelineQuery, surfaces, query/saved-view model tests | high | Preserve active surface, URL/query state, `query_json`, and `layout_json`. |
| WB-SL-08 | Generic surface controller seam | WB-SL-07 | `GenericWorkbookSurface.tsx`, generic models/components | Responsibility extraction | yes | Generic model tests, surfaces, phase3 grid/payload as applicable | medium | Prioritize controller/model split over cosmetic file moves. |
| WB-SL-09 | Test-support consolidation after runtime seams | WB-SL-02 through WB-SL-08 | `apps/web/src/testing/*`, shell tests if authorized | Test support cleanup | yes | Full targeted shell suite plus frontend unit | medium | Defer until runtime seams stabilize; preserve phase-characterization traceability. |

## 9. Validation matrix

| Validation ID | Command | Scope | When to run | Expected evidence | If failing |
| ------------- | ------- | ----- | ----------- | ----------------- | ---------- |
| WB-VAL-01 | `make explain-target TARGET=frontend-unit DETAIL=summary` | Target discovery | Planning or before implementation | Public target summary and artifact expectations | Re-check Make target surface before running tests. |
| WB-VAL-02 | `make explain-target TARGET=frontend-import-boundary-check DETAIL=summary` | Import-boundary discovery | Before boundary slices | Public target summary | Re-check `tools/frontend_import_boundaries.json`. |
| WB-VAL-03 | `make task-guide ROLE=feature-dev PHASE=phase9` | Phase verification planning | Before phase9-adjacent work | Narrow and broad recommended targets | Use guide output to choose narrower phase target. |
| WB-VAL-04 | `make frontend-unit VITEST_FLAGS='apps/web/src/workbook/WorkbookShell.assessments.test.tsx'` | Assessment shell behavior | Assessment/shell slices | Targeted frontend unit evidence | Inspect failing assertion and fetch/mock artifacts if emitted. |
| WB-VAL-05 | `make frontend-unit VITEST_FLAGS='apps/web/src/workbook/WorkbookShell.phase3.autosave.test.tsx'` | Autosave/pending | Pending/autosave slices | Targeted frontend unit evidence | Treat as slice blocker until pending semantics are restored. |
| WB-VAL-06 | `make frontend-unit VITEST_FLAGS='apps/web/src/workbook/WorkbookShell.phase3.grid.test.tsx'` | Grid/create | Grid slices | Targeted frontend unit evidence | Check grid facade props and create-row behavior. |
| WB-VAL-07 | `make frontend-unit VITEST_FLAGS='apps/web/src/workbook/WorkbookShell.phase3.payload.test.tsx'` | Payload shape | Payload/mutation slices | Targeted frontend unit evidence | Compare request-body expectations before changing payload builders. |
| WB-VAL-08 | `make frontend-unit VITEST_FLAGS='apps/web/src/workbook/WorkbookShell.phase4.actionSequencing.test.tsx'` | Action sequencing | Pending/mutation slices | Targeted frontend unit evidence | Inspect action order and queue replay. |
| WB-VAL-09 | `make frontend-unit VITEST_FLAGS='apps/web/src/workbook/WorkbookShell.phase4.saveState.test.tsx'` | Save state | Pending/save-state slices | Targeted frontend unit evidence | Check status derivation and visible save labels. |
| WB-VAL-10 | `make frontend-unit VITEST_FLAGS='apps/web/src/workbook/WorkbookShell.phase4.support.test.tsx'` | Timeline/workbook support | Timeline shell slices | Targeted frontend unit evidence | Inspect shell/timeline interaction regressions. |
| WB-VAL-11 | `make frontend-unit VITEST_FLAGS='apps/web/src/workbook/WorkbookShell.phase4.timelineQuery.test.tsx'` | Timeline query normalization | Query slices | Targeted frontend unit evidence | Check query request and normalized row behavior. |
| WB-VAL-12 | `make frontend-unit VITEST_FLAGS='apps/web/src/workbook/WorkbookShell.phase5.gridProvenance.test.tsx'` | Grid provenance | Grid contract-row slices | Targeted frontend unit evidence | Check provenance metadata and contract rows. |
| WB-VAL-13 | `make frontend-unit VITEST_FLAGS='apps/web/src/workbook/WorkbookShell.phase5.mentionChips.test.ts'` | Mention chips | Mention/model slices | Targeted frontend unit evidence | Check chip model changes. |
| WB-VAL-14 | `make frontend-unit VITEST_FLAGS='apps/web/src/workbook/WorkbookShell.phase5.test.tsx'` | Evidence behavior | Evidence slices | Targeted frontend unit evidence | Check evidence panel/actions and payloads. |
| WB-VAL-15 | `make frontend-unit VITEST_FLAGS='apps/web/src/workbook/WorkbookShell.phase6.test.tsx'` | Collaboration/session/pending | Collaboration slices | Targeted frontend unit evidence | Check WebSocket/session mocks and pending transitions. |
| WB-VAL-16 | `make frontend-unit VITEST_FLAGS='apps/web/src/workbook/WorkbookShell.phase7.test.tsx'` | History/inspector | History/inspector slices | Targeted frontend unit evidence | Check history request/preview/confirm and inspector state. |
| WB-VAL-17 | `make frontend-unit VITEST_FLAGS='apps/web/src/workbook/WorkbookShell.phase8.query.test.tsx'` | Saved-view/query | Query/saved-view slices | Targeted frontend unit evidence | Check active view and persisted query behavior. |
| WB-VAL-18 | `make frontend-unit VITEST_FLAGS='apps/web/src/workbook/WorkbookShell.phase9.inspector.test.tsx'` | Inspector/row-local actions | Inspector slices | Targeted frontend unit evidence | Check row-local action selectors and inspector mounting. |
| WB-VAL-19 | `make frontend-unit VITEST_FLAGS='apps/web/src/workbook/WorkbookShell.phase9.sentinel.test.tsx'` | Sentinel/focus/continuity | Focus/continuity slices | Targeted frontend unit evidence | Check focus restoration and sentinel behavior. |
| WB-VAL-20 | `make frontend-unit VITEST_FLAGS='apps/web/src/workbook/WorkbookShell.surfaces.test.tsx'` | Multi-surface shell | Shell/surface slices | Targeted frontend unit evidence | Check surface routing and shell layout. |
| WB-VAL-21 | `make frontend-unit VITEST_FLAGS='apps/web/src/workbook/models'` | Workbook model tests | Model-only slices | Targeted frontend unit evidence | Narrow to failing model file. |
| WB-VAL-22 | `make frontend-unit VITEST_FLAGS='apps/web/src/workbook/timeline/models'` | Timeline model tests | Timeline model slices | Targeted frontend unit evidence | Narrow to failing model file. |
| WB-VAL-23 | `make frontend-unit VITEST_FLAGS='apps/web/src/workbook/timeline/services'` | Timeline service tests | Collaboration/mutation service slices | Targeted frontend unit evidence | Inspect service wire semantics. |
| WB-VAL-24 | `make frontend-unit VITEST_FLAGS='apps/web/src/workbook/utils'` | Workbook utility tests | Utility/focus/pending slices | Targeted frontend unit evidence | Narrow to failing utility. |
| WB-VAL-25 | `make frontend-unit` | All frontend unit tests | After multiple workbook slices | Frontend unit target evidence | Use failing test paths to reduce scope. |
| WB-VAL-26 | `make frontend-typecheck` | Frontend typecheck | After exported type or facade changes | Typecheck evidence | Fix type/export drift before behavior tests. |
| WB-VAL-27 | `make frontend-import-boundary-check` | Frontend import boundaries | After import/facade changes | Import-boundary evidence | Fix direct vendor/generated/test-helper imports. |
| WB-VAL-28 | `make lint-biome` | Frontend lint/format check | After frontend source edits | Lint evidence | Apply repo formatter only if authorized by implementation task. |
| WB-VAL-29 | `make phase-slice PHASE=phase3` through `make phase-slice PHASE=phase9` | Phase-level evidence | When a slice touches phase-owned behavior | Phase slice evidence | Use phase guide/explain output to narrow related failures. |
| WB-VAL-30 | `make service-backed-slice PHASE=phase9` | Browser/service-backed phase9 evidence | After focus/inspector UX changes needing browser coverage | Service-backed evidence | Inspect browser artifacts and do not claim visual/accessibility conformance unless target ran. |
| WB-VAL-31 | `make generated-artifact-policy-check` | Generated artifact policy | After any generated-root adjacency | Policy evidence | Stop if generated files were hand-edited. |
| WB-VAL-32 | `make json-shape-check` | JSON/schema shape checks | After contract-shape-adjacent changes | JSON-shape evidence | Re-check owner inputs before schema changes. |
| WB-VAL-33 | `make agent-finalize` | End-of-run maintenance | Before broad implementation handoff | Finalize evidence or retained-run note | Report skipped retained-run maintenance if `RESULTS_DIR` is unset. |
| WB-VAL-34 | `make test-fast` | Broader fast gate | Before finishing substantial implementation | Fast test/check evidence | Report failing target and run root. |
| WB-VAL-35 | `make check` | Broad local check | Before major handoff if risk warrants | Check evidence | Report failing target and related artifacts. |
| WB-VAL-36 | `pnpm --dir apps/web exec vitest run <path>` | Developer convenience only | Local diagnosis when Make wrapper is insufficient | Non-canonical Vitest output | Prefer Make evidence for handoff; mark direct run as convenience. |

## 10. Handoff log

### Handoff: shell coordination

| Time | Session | State | Files inspected | Decisions | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-06-28 00:38:21 EDT | planning artifact write on `main@52f2ce2` | Planning complete; artifact written; no production/test/generated code edits. | `WorkbookShell.tsx`, `useWorkbookShellRuntime.ts`, shell components, `TimelineWorkbook.tsx`, `GenericWorkbookSurface.tsx` | Preserve behavior; keep phase tests as characterization; slice shell work through models/hooks first. | None for planning. | Choose first behavior-preserving shell seam and run narrow characterization before editing. |

### Handoff: grid and package facades

| Time | Session | State | Imports inspected | Boundary findings | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-06-28 00:38:21 EDT | planning artifact write on `main@52f2ce2` | Static import scan complete. | `@cartulary/grid-adapter`, `react-data-grid`, grid test-support, grid utilities/components | No direct runtime `react-data-grid` import found under target; test-only vendor assertion exists in phase3 grid test. | None for planning. | Preserve vendor boundary; use `make frontend-import-boundary-check` after grid slices. |

### Handoff: protocol/UI/view contracts

| Time | Session | State | Imports inspected | Generated-internal findings | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-06-28 00:38:21 EDT | planning artifact write on `main@52f2ce2` | Static facade scan complete. | `@cartulary/protocol-ts`, `@cartulary/ui-contracts`, `@cartulary/view-contracts`, generated path segments | No direct runtime generated-internal import found under target. `@cartulary/protocol-ts` is runtime-imported while listed as an app devDependency. | `TODO:` verify dependency posture owner before package metadata edits. | Keep imports on public facades and confirm dependency declaration before implementation. |

### Handoff: tests and characterization

| Time | Session | State | Tests inspected | Coverage notes | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-06-28 00:38:21 EDT | planning artifact write on `main@52f2ce2` | Test inventory mapped; tests not run. | All supplied `WorkbookShell.*.test.*`; model/service/utility tests under target; `apps/web/src/testing/*` helpers | Phase tests cover grid, autosave, payload, action sequencing, save state, timeline query, evidence, collaboration, history, query, inspector, sentinel, surfaces. | None for planning; green status unknown because tests were not run. | For each slice, run the exact targeted `make frontend-unit VITEST_FLAGS=...` tests first and after. |

### Handoff: inspector/history/sentinel/focus

| Time | Session | State | Files inspected | UX risks | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-06-28 00:38:21 EDT | planning artifact write on `main@52f2ce2` | Seam candidates identified. | `TimelineWorkbook.tsx`, `TimelineWorkbookInspector.tsx`, `TimelineHistoryPanel.tsx`, inspector/history hooks, focus/continuity utilities/models | Default-closed inspector, row-local action routing, history preview, focus restoration, viewport continuity. | None for planning. | Start only after phase7 and phase9 characterization commands are selected. |

### Handoff: saved-view/query/multi-surface

| Time | Session | State | Files inspected | Contract risks | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-06-28 00:38:21 EDT | planning artifact write on `main@52f2ce2` | Query/surface seams identified. | `useWorkbookShellRuntime.ts`, `workbookQuery.ts`, saved-view models, surface registry, `TimelineWorkbook.tsx`, `WorkbookShell.surfaces.test.tsx` | Active surface selection, `view_schema_id`, `query_json`, `layout_json`, URL/query state, timeline query normalization. | None for planning. | Use phase8, phase4 timelineQuery, surfaces, and model tests for any query/surface extraction. |

## 11. Open questions and blockers

| Question ID | Question | Why it matters | Owner needed | Current evidence | Status |
| --- | --- | --- | --- | --- | --- |
| WB-Q-01 | Should future workbook sessions prefer `docs/handoffs/cartulary_modular_refactor_planning_framework.md` when `temp/current.md` exists but is not the generic framework? | Prevents accidental use of an inspector-specific patch plan as general framework. | Planning/session owner | `temp/current.md` exists but is titled as an inspector workflow closure plan. | TODO |
| WB-Q-02 | Should `@cartulary/protocol-ts` be a runtime dependency of `apps/web` if runtime workbook files keep importing it? | Package metadata may not match runtime import usage. | Frontend/package owner | `apps/web/package.json` lists it under `devDependencies`; runtime files import the facade. | TODO |
| WB-Q-03 | Which exports from `WorkbookShell.tsx` are intentional public test seams? | Moving exports can break characterization tests even if UI behavior is unchanged. | Frontend/workbook owner | Shell tests import runtime entries and helpers from `./WorkbookShell`. | TODO |
| WB-Q-04 | Should any future workbook-wide refactor coordinate with backend `internal/modules/workbook` ownership, or remain web-only unless explicitly scoped? | Repo-wide scan found a backend workbook module in addition to the web workbook target. | Frontend/backend module owners | `internal/modules/workbook` exists and contains Go workbook APIs, routes, stores, and tests. | TODO |

| Blocker ID | Blocker | Affected workflow | Evidence | Required resolution | Status |
| --- | --- | --- | --- | --- | --- |
| WB-B-01 | No implementation validation has been run. | WF-WB-08 and later implementation | Only static inspection and Make explanation/task-guide commands were run. | Run targeted validation for the chosen slice before and after code edits. | TODO |
| WB-B-02 | No owner contradiction found during planning scan. | None | Core/harness/domain/framework/live-code posture was compatible for this planning task. | Not applicable unless a later source conflict appears. | not_applicable |

## 12. Binary acceptance criteria

The planning artifact status:

- Every supplied file is represented in the verified inventory: satisfied.
- Every discovered file under the resolved target directory is classified: satisfied for the static target scan.
- Direct runtime imports from generated internals are identified or explicitly marked absent: satisfied by static scan; none found.
- Direct runtime imports from grid vendor APIs are identified or explicitly marked absent: satisfied by static scan; none found.
- Test files are mapped to behavior surfaces and phase-characterization value: satisfied.
- Each proposed workflow has dependencies, outputs, validation, and stop condition: satisfied.
- Each proposed implementation slice is behavior-preserving or marked as requiring owner approval: satisfied.
- Every validation command is discovered from the repo or marked `TODO: verify`: satisfied; commands are Make-owned unless explicitly labeled developer convenience.
- Every unresolved authority or repo-state uncertainty is listed as a question or blocker: satisfied.
- The handoff log is sufficient for a later implementation agent to start without redoing the planning scan: satisfied for target-local planning.

Artifact path: `temp/workbook_refactor_tracker.md`.

Highest-risk findings:

- `TimelineWorkbook.tsx` and `WorkbookShell.tsx` concentrate high-risk shell, timeline, mutation, inspector, focus, and query behavior.
- No direct runtime `react-data-grid` or generated-internal imports were found under `apps/web/src/workbook`, but grid mechanics still need to stay behind `@cartulary/grid-adapter`.
- Phase tests are important characterization assets and should not be reorganized merely because of phase-number filenames.
- Runtime imports from `@cartulary/protocol-ts` exist while `apps/web` lists that package as a dev dependency; dependency posture needs owner verification.
- Root workbook and timeline subfolders are heavily coupled, so future work should use small behavior-preserving seams rather than broad moves.

Readiness: this plan is ready as an implementation handoff planner. It is not a green-light validation report because tests were not run.
