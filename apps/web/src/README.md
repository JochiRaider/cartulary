# apps/web/src Layout

This tree is organized by implementation ownership only. Directory placement
does not define product behavior, workbook-surface identity, route ownership,
or domain vocabulary. Product behavior remains owned by the core specs and
domain vocabulary remains owned by `docs/domain.md`.

Use direct relative imports. Do not add path aliases or app-local barrel
exports without updating this convention and the import-boundary checks.

Keep tests beside the module or surface they cover. Shared fixtures belong in
`testing/`. Phase-named test files are coverage/harness accounting labels only;
they are not runtime architecture boundaries.

Generated artifacts and generated harness outputs remain owned by their source
manifests and must not be hand-edited.

## Root Files

| File | Responsibility |
| --- | --- |
| `README.md` | Local implementation-ownership map for `apps/web/src`. Update it when files are added, removed, or moved in this tree. |
| `main.tsx` | Vite browser entry shim. Mounts the React application root and should stay thin. |

## `app/`

The `app/` directory owns the application shell, route-level browser entry
surfaces, authentication gateway, landing/admin surfaces, debug harness
entrypoints, and app-shell tests. It should not own workbook internals.

| File | Responsibility |
| --- | --- |
| `app/App.tsx` | Top-level application composition and route/surface selection for the web app. |
| `app/AppRoot.tsx` | Root React wrapper that connects app-level providers and the rendered `App`. |
| `app/AccountAdministrationPanels.tsx` | Account-security and deployment-user administration panels. |
| `app/AccountSettingsPanels.tsx` | Account profile and appearance settings panels. |
| `app/api/appShellClient.ts` | App-shell client helpers for auth, account, deployment administration, and extension profile requests. |
| `app/AuthGateway.tsx` | Authentication-state gate around app content and login/account readiness. |
| `app/debug/DebugHarnessShell.tsx` | Shared shell for development/debug harness pages. |
| `app/DeploymentAuditPanel.tsx` | Deployment administrative audit panel and audit-event formatting. |
| `app/IncidentImportPanel.tsx` | Incident bundle import panel and import-job polling controls. |
| `app/IncidentAdminPanel.tsx` | Incident administration panel UI for incident metadata, preferences, membership, and audit affordances. |
| `app/IncidentLanding.tsx` | Incident directory landing panel, search/filter controls, and create-incident dialog. |
| `app/LandingAdminDisplay.tsx` | Shared display helpers for landing/admin panels. |
| `app/LandingAdminLayout.tsx` | Landing and deployment-administration shell layout plus the account/application menu. |
| `app/LandingAdminSurface.tsx` | Compatibility re-export for landing/admin modules; new runtime imports should prefer the cohesive modules directly. |
| `app/debug/Phase1Harness.tsx` | Phase 1 debug harness entrypoint for auth, account, and route readiness scenarios. |
| `app/debug/Phase2Harness.tsx` | Phase 2 debug harness entrypoint for incident setup and preference scenarios. |
| `app/referencePackAdminClient.ts` | Reference-pack administration HTTP client helpers. |
| `app/referencePackAdminModel.ts` | Reference-pack administration resource, query, paging, and session-shape types. |
| `app/ReferencePackAdminPanel.tsx` | Reference-pack administration panel UI for import, reload, cancellation, and job-status controls. |
| `app/landingAdminStyles.ts` | Shared style constants for landing/admin panels. |
| `app/landingAdminTypes.ts` | Shared landing/admin TypeScript model types. |
| `app/routeState.ts` | Pure app-route parsing and history URL construction helpers. |
| `app/useAppRouteRuntime.ts` | React hook for route state, popstate handling, and history writes. |
| `app/App.landing.test.tsx` | Landing-surface tests for app startup and landing interactions. |
| `app/App.phase1.support.test.tsx` | Support tests for Phase 1 app harness behavior. |
| `app/App.phase1.test.tsx` | Phase 1 app behavior tests for auth/account/route readiness. |
| `app/App.test.tsx` | General app-shell behavior tests. |
| `app/IncidentAdminPanel.test.tsx` | Incident administration panel tests. |
| `app/ReferencePackAdminPanel.test.tsx` | Reference-pack administration panel tests. |
| `app/api/appShellClient.routeBoundary.test.ts` | Route-boundary tests for app-shell client helpers. |
| `app/fontBundle.test.ts` | Font bundle availability and packaging boundary tests. |
| `app/fontRoles.test.tsx` | Font-role presentation tests for app and workbook surfaces. |
| `app/otelBoundary.test.ts` | OpenTelemetry import and runtime-boundary tests. |
| `app/routeState.test.ts` | Route-state parsing and history write tests. |

## `services/`

The `services/` directory owns browser transport helpers and small client
facades used by app/workbook code. It should not encode domain behavior beyond
request/response handling already owned by specs and backend contracts.

| File | Responsibility |
| --- | --- |
| `services/browserApi.ts` | Browser API base/path helpers for app-local HTTP calls. |
| `services/workbookApi.ts` | Workbook HTTP helper utilities, envelope parsing, abort/query runtime helpers, and user-facing error extraction. |
| `services/workbookEvidence.ts` | Evidence upload/attach client helpers and evidence public-error mapping. |
| `services/browserApi.test.ts` | Tests for browser API base/path helpers. |
| `services/workbookApi.test.ts` | Tests for workbook API envelope, abort, and error helpers. |
| `services/workbookEvidence.test.ts` | Tests for evidence client helpers and error mapping. |

## `shared/`

The `shared/` directory owns cross-feature helpers that are not workbook
specific.

| File | Responsibility |
| --- | --- |
| `shared/publicError.ts` | Shared public-error normalization helpers. |
| `shared/workbookSheetRef.ts` | Shared workbook sheet-reference contract and runtime guard. |
| `shared/workbookShellContracts.ts` | Shared app/workbook shell contracts for account identity, application menu handoff, and incident-controls renderer props. |

## `testing/`

The `testing/` directory owns Vitest setup, reusable app/workbook fixtures, and
policy tests. Runtime application code must not import this directory.

| File | Responsibility |
| --- | --- |
| `testing/appShellTestSupport.ts` | Shared app-shell test helpers and fixtures. |
| `testing/fetchMockTestSupport.ts` | Fetch mock helpers for unit and integration-style frontend tests. |
| `testing/selectorContractPolicy.test.ts` | Selector ownership policy test that guards raw `data-testid` literals and shared selector facade usage. |
| `testing/testSetup.dom.ts` | DOM-specific Vitest setup. |
| `testing/testSetup.ts` | Common Vitest setup for frontend tests. |
| `testing/timelineWorkbookTestSupport.ts` | Shared Timeline workbook fixture helpers, route mocks, and row builders for tests. |
| `testing/timelineWorkbookTestSupport.test.tsx` | Tests for Timeline workbook test-support helpers. |

## `workbook/`

The `workbook/` directory owns the workbook shell, shell-level tests, and
workbook subfolders. Runtime code in this directory should use package facades
for grid, protocol, UI contracts, and view contracts rather than generated
internals.

| File | Responsibility |
| --- | --- |
| `workbook/WorkbookShell.tsx` | Workbook shell coordinator. Composes surfaces, saved-view/query controls, incident controls, generic surfaces, assessment/entity support, and Timeline entrypoints. |
| `workbook/WorkbookShell.assessments.test.tsx` | Assessment workbook behavior tests. |
| `workbook/WorkbookShell.phase3.autosave.test.tsx` | Phase 3 autosave and pending-save characterization. |
| `workbook/WorkbookShell.phase3.grid.test.tsx` | Phase 3 grid/create behavior characterization. |
| `workbook/WorkbookShell.phase3.payload.test.tsx` | Phase 3 request payload characterization. |
| `workbook/WorkbookShell.phase4.actionSequencing.test.tsx` | Phase 4 action sequencing tests. |
| `workbook/WorkbookShell.phase4.saveState.test.tsx` | Phase 4 save-state tests. |
| `workbook/WorkbookShell.phase4.support.test.tsx` | Phase 4 support tests for Timeline/workbook interactions. |
| `workbook/WorkbookShell.phase4.timelineQuery.test.tsx` | Timeline query and view-row normalization tests. |
| `workbook/WorkbookShell.phase5.gridProvenance.test.tsx` | Grid provenance and contract-row tests for Phase 5 surfaces. |
| `workbook/WorkbookShell.phase5.mentionChips.test.ts` | Mention chip model tests. |
| `workbook/WorkbookShell.phase5.test.tsx` | Phase 5 workbook/evidence behavior tests. |
| `workbook/WorkbookShell.phase6.test.tsx` | Phase 6 collaboration/session and pending behavior tests. |
| `workbook/WorkbookShell.phase7.test.tsx` | Phase 7 history and inspector behavior tests. |
| `workbook/WorkbookShell.phase8.query.test.tsx` | Phase 8 saved-view/query behavior tests. |
| `workbook/WorkbookShell.phase9.inspector.test.tsx` | Phase 9 inspector and row-local action tests. |
| `workbook/WorkbookShell.phase9.sentinel.test.tsx` | Phase 9 sentinel, focus, and continuity tests. |
| `workbook/WorkbookShell.surfaces.test.tsx` | Multi-surface workbook shell tests. |

### `workbook/components/`

Shell-level presentation components live here. They should receive behavior
through props and local workbook models rather than own transport or domain
workflow logic.

| File | Responsibility |
| --- | --- |
| `workbook/components/ActiveSurfaceSavedViewSelector.tsx` | Saved-view selector for the active workbook surface. |
| `workbook/components/GenericMutationControl.tsx` | Generic row mutation controls for system-view surfaces. |
| `workbook/components/GenericWorkbookSurface.tsx` | Generic workbook surface renderer for contract-backed non-Timeline surfaces. |
| `workbook/components/SystemViewSwitcher.tsx` | System-view switcher UI and grouped surface navigation. |
| `workbook/components/WorkbookGridControls.tsx` | Reusable workbook grid filter/sort/grouping control shell. |
| `workbook/components/WorkbookInspectorFeatureGroups.tsx` | Inspector feature-group renderer and disabled-state presentation helpers. |
| `workbook/components/WorkbookSheetToolbar.tsx` | Workbook sheet toolbar composition. |
| `workbook/components/WorkbookShellSlots.tsx` | Stable shell slot IDs, labels, and layout slot helpers. |
| `workbook/components/WorkbookStatusStrip.tsx` | Status strip presentation for save/load/selection state. |
| `workbook/components/WorkbookSurfaceFrame.tsx` | Shared surface frame and style primitives for workbook grid/inspector layouts. |

### `workbook/hooks/`

Shell-level hooks coordinate reusable runtime behavior. Hooks in this folder
should remain workbook-generic unless their name clearly scopes them to a
specific surface.

| File | Responsibility |
| --- | --- |
| `workbook/hooks/useAssessmentSupportRows.ts` | Loads and normalizes Timeline support rows needed by assessment workflows. |
| `workbook/hooks/useEntityTimelinePreview.ts` | Loads Timeline preview rows for entity-related workbook workflows. |
| `workbook/hooks/useGenericReferenceOptions.ts` | Loads reference options used by generic workbook surface create/edit controls. |
| `workbook/hooks/useWorkbookIncidentIdentity.ts` | Resolves incident identity/loading state for the workbook shell. |
| `workbook/hooks/useWorkbookResponsiveLayout.ts` | Coordinates shell responsive-layout state from viewport measurements. |
| `workbook/hooks/useWorkbookShellRuntime.ts` | Shell runtime hook for startup, active surface, saved views, query state, and runtime commands. |

### `workbook/models/`

Workbook models are pure or near-pure TypeScript helpers for request payloads,
view models, state machines, and contract-backed workbook decisions. They may
use `@cartulary/view-contracts` facades, but app workflow state stays in the
app.

| File | Responsibility |
| --- | --- |
| `workbook/models/assessmentWorkbookModel.ts` | Assessment workbook draft, payload, confidence-band, and support-row helpers. |
| `workbook/models/entityWorkbookModel.ts` | Host/identity entity row, merge-plan, grouping, and payload helpers. |
| `workbook/models/evidenceLifecycleViewModel.ts` | Evidence lifecycle display/count view-model helpers. |
| `workbook/models/genericWorkbookModel.ts` | Generic system-view create/edit payload, enum, validation, and row-label helpers. |
| `workbook/models/workbookContractRows.ts` | Contract-backed row normalization and grid-column materialization helpers for workbook surfaces. |
| `workbook/models/workbookDensity.ts` | Account density preference resolution. |
| `workbook/models/workbookIncidentIdentity.ts` | Incident identity normalization and loading-state model. |
| `workbook/models/workbookInspectorModel.ts` | Workbook inspector state, reducer, panel, and feature-group helpers. |
| `workbook/models/workbookQuery.ts` | Workbook query, filter, sort, grouping, and request-building helpers. |
| `workbook/models/workbookReferenceOptions.ts` | Reference option normalization and lookup helpers. |
| `workbook/models/workbookResponsiveLayout.ts` | Responsive layout classification and surface-band helpers. |
| `workbook/models/workbookSavedViewRuntime.ts` | Saved-view runtime selection, dirty-state, and command helpers. |
| `workbook/models/workbookSavedViews.ts` | Saved-view resource normalization and payload helpers. |
| `workbook/models/workbookStartup.ts` | Workbook startup candidate, selected sheet reference, and fallback resolution helpers. |
| `workbook/models/workbookSurfaceRegistry.ts` | Built-in/system/optional workbook surface registry and stable view-schema IDs. |
| `workbook/models/assessmentWorkbookModel.test.ts` | Tests for assessment workbook model helpers. |
| `workbook/models/entityWorkbookModel.test.ts` | Tests for entity workbook model helpers. |
| `workbook/models/evidenceLifecycleViewModel.test.ts` | Tests for evidence lifecycle view-model helpers. |
| `workbook/models/genericWorkbookModel.test.ts` | Tests for generic workbook model helpers. |
| `workbook/models/workbookDensity.test.ts` | Tests for workbook density resolution. |
| `workbook/models/workbookInspectorModel.test.ts` | Tests for workbook inspector state/model helpers. |
| `workbook/models/workbookQuery.test.ts` | Tests for workbook query helpers. |
| `workbook/models/workbookReferenceOptions.test.ts` | Tests for reference-option helpers. |
| `workbook/models/workbookResponsiveLayout.test.ts` | Tests for responsive layout model helpers. |
| `workbook/models/workbookSavedViewRuntime.test.ts` | Tests for saved-view runtime helpers. |
| `workbook/models/workbookSavedViews.test.ts` | Tests for saved-view normalization and payload helpers. |
| `workbook/models/workbookStartup.test.ts` | Tests for workbook startup resolution. |
| `workbook/models/workbookSurfaceRegistry.test.ts` | Tests for workbook surface registry invariants. |

### `workbook/timeline/`

Timeline-specific implementation lives under `workbook/timeline/`. This folder
owns Timeline presentation, Timeline hooks, Timeline models, and Timeline
services. Generic workbook behavior belongs one level up unless it is truly
Timeline-specific.

#### `workbook/timeline/components/`

| File | Responsibility |
| --- | --- |
| `workbook/timeline/components/TimelineCellEditors.tsx` | Timeline scalar editors, draft-row create button, and relationship chip presentation. |
| `workbook/timeline/components/TimelineConflictResolver.tsx` | Same-field conflict resolver presentation. |
| `workbook/timeline/components/TimelineEvidencePanel.tsx` | Timeline inspector evidence panel and evidence actions UI. |
| `workbook/timeline/components/TimelineGridSurface.tsx` | Timeline grid surface wrapper around the grid adapter boundary. |
| `workbook/timeline/components/TimelineHistoryPanel.tsx` | Timeline row history, rollback, delete, restore, and history action presentation. |
| `workbook/timeline/components/TimelineMentionsPanel.tsx` | Timeline mention-resolution inspector panel. |
| `workbook/timeline/components/TimelinePresenceMarkers.tsx` | Timeline row/cell presence marker presentation. |
| `workbook/timeline/components/TimelineRowActions.tsx` | Timeline row action/context-menu presentation. |
| `workbook/timeline/components/TimelineWorkbook.tsx` | Timeline hot-path controller for rows, pending saves, collaboration, grid continuity, inspector routing, create/edit, conflict, evidence, history, and related-row workflows. |
| `workbook/timeline/components/TimelineWorkbookGrid.tsx` | Timeline grid renderer and row/cell composition. |
| `workbook/timeline/components/TimelineWorkbookInspector.tsx` | Timeline inspector shell, panel tabs, selected-row state presentation, and inspector messages. |
| `workbook/timeline/components/TimelineWorkbookNotices.tsx` | Timeline notices, pending queue messages, and save-state presentation. |
| `workbook/timeline/components/TimelineEvidencePanel.test.tsx` | Tests for Timeline evidence panel behavior. |

#### `workbook/timeline/hooks/`

| File | Responsibility |
| --- | --- |
| `workbook/timeline/hooks/useTimelineCommittedRows.ts` | Derives committed Timeline row collections from row/runtime state. |
| `workbook/timeline/hooks/useTimelineConflicts.ts` | Coordinates Timeline same-field conflict state. |
| `workbook/timeline/hooks/useTimelineEvidenceActions.ts` | Coordinates Timeline evidence attach/preview/download action state. |
| `workbook/timeline/hooks/useTimelineGridInteractions.ts` | Coordinates Timeline grid refs, focus anchors, viewport continuity, keyboard helpers, and interaction commands. |
| `workbook/timeline/hooks/useTimelineHistoryState.ts` | Coordinates Timeline history panel and row-history state. |
| `workbook/timeline/hooks/useTimelineInspectorSelection.ts` | Coordinates selected Timeline row and inspector selection state. |
| `workbook/timeline/hooks/useTimelineLiveUpdates.ts` | Coordinates Timeline WebSocket/live-update side effects and presence/session callbacks. |
| `workbook/timeline/hooks/useTimelineMentions.ts` | Coordinates Timeline mention-resolution state and actions. |
| `workbook/timeline/hooks/useTimelinePendingSaves.ts` | Coordinates Timeline pending-save queue runtime and replay admission. |
| `workbook/timeline/hooks/useTimelineRows.ts` | Coordinates Timeline row state, draft rows, and row reconciliation. |
| `workbook/timeline/hooks/useTimelineWorkbookRuntime.ts` | Coordinates Timeline query/runtime state derived from workbook shell inputs. |

#### `workbook/timeline/models/`

| File | Responsibility |
| --- | --- |
| `workbook/timeline/models/timelineConflictModel.ts` | Same-field conflict parsing and model helpers. |
| `workbook/timeline/models/timelineRowsModel.ts` | Timeline row collection helpers and row-state utilities. |
| `workbook/timeline/models/timelineViewportContinuityModel.ts` | Timeline viewport continuity and entity-refresh barrier helpers. |
| `workbook/timeline/models/workbookMentionChips.ts` | Mention chip state, relationship-field keys, and mention display helpers. |
| `workbook/timeline/models/workbookTimelineModel.ts` | Timeline row model, field bindings, payload builders, normalization, patch intents, freshness decisions, and display helpers. |
| `workbook/timeline/models/timelineConflictModel.test.ts` | Tests for Timeline conflict model helpers. |
| `workbook/timeline/models/timelineRowsModel.test.ts` | Tests for Timeline row model helpers. |
| `workbook/timeline/models/timelineViewportContinuityModel.test.ts` | Tests for viewport continuity and refresh barrier helpers. |
| `workbook/timeline/models/workbookTimelineModel.test.ts` | Tests for Timeline row, payload, binding, normalization, and freshness helpers. |

#### `workbook/timeline/services/`

Timeline services are behavior-named helpers for wire-message construction,
HTTP mutation dispatch, and socket lifecycle state. They are not backend route
owners.

| File | Responsibility |
| --- | --- |
| `workbook/timeline/services/timelineMutationRequests.ts` | Timeline pending-replay HTTP dispatch and timing event helper. |
| `workbook/timeline/services/workbookCollaborationMessages.ts` | Timeline collaboration message, presence payload, self-origin filtering, and mention action payload helpers. |
| `workbook/timeline/services/workbookSocketLifecycle.ts` | Workbook WebSocket lifecycle state machine and effect planner. |
| `workbook/timeline/services/workbookCollaborationMessages.test.ts` | Tests for collaboration/presence/session message helper output. |
| `workbook/timeline/services/workbookSocketLifecycle.test.ts` | Tests for WebSocket lifecycle state transitions and planned effects. |

### `workbook/utils/`

Workbook utilities are reusable helpers that are not specific enough to belong
under `timeline/` and not broad enough for `shared/`.

| File | Responsibility |
| --- | --- |
| `workbook/utils/workbookClipboard.ts` | Clipboard grid-shape and paste helpers. |
| `workbook/utils/workbookContinuity.ts` | Viewport anchor capture, restoration, and rectangle visibility helpers. |
| `workbook/utils/workbookGridFocus.tsx` | Workbook grid focus helpers and focusable-cell wrapper. |
| `workbook/utils/workbookKeyboard.ts` | Workbook keyboard command mapping. |
| `workbook/utils/workbookPendingQueue.ts` | Pending-save queue capacity, save-state, replay, conflict, and public-error helpers. |
| `workbook/utils/workbookPresence.ts` | Presence input/type helpers and presence matching helpers. |
| `workbook/utils/workbookStyles.ts` | Shared workbook style primitives. |
| `workbook/utils/workbookValueFormat.ts` | Grid/workbook value formatting helpers. |
| `workbook/utils/GridAdapter.phase9.anchor.test.ts` | Phase 9 grid-adapter anchor behavior tests. |
| `workbook/utils/workbookClipboard.test.ts` | Tests for clipboard helpers. |
| `workbook/utils/workbookContinuity.test.ts` | Tests for continuity helpers. |
| `workbook/utils/workbookKeyboard.test.ts` | Tests for keyboard command mapping. |
| `workbook/utils/workbookPendingQueue.test.ts` | Tests for pending-queue helpers. |
| `workbook/utils/workbookValueFormat.test.ts` | Tests for value formatting helpers. |
