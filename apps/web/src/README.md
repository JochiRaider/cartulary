# apps/web/src Layout

This tree is organized by implementation ownership only. Directory placement
does not define product behavior, workbook-surface identity, route ownership,
or domain vocabulary. Product behavior remains owned by the core specs and
domain vocabulary remains owned by `docs/domain.md`.

Use direct relative imports. Do not add path aliases or app-local barrel
exports without updating this convention and the import-boundary checks.

Keep tests beside the module or surface they cover. Shared fixtures belong in
`testing/`. Test files and selectors use semantic owner and behavior identities;
they are not runtime architecture boundaries.

Generated artifacts and generated harness outputs remain owned by their source
manifests and must not be hand-edited.

## Coverage Scope

This README is a narrative implementation-ownership guide for authored files
under `apps/web/src`. Exact `.ts`/`.tsx` source accounting is owned by
`tools/frontend_source_ownership.json` and enforced by the `web.architecture`
policy tests. Executable checks must not read, stat, hash, or parse this
Markdown file. Directory headings describe the local implementation-ownership
boundary; representative table rows describe file-level responsibilities.

When a structurally significant file is added, removed, or moved, reconcile the
matching narrative section in the same change and update the machine ownership
manifest. Test-file rows should name the behavior or contract they pin, not the
mechanics of the test harness.

## Root Files

| File | Responsibility |
| --- | --- |
| `README.md` | Narrative implementation-ownership guide for `apps/web/src`; never an executable accounting input. |
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
| `app/api/publicHttpTypes.ts` | Public app-shell HTTP request and response type exports from the generated protocol facade. |
| `app/AuthGateway.tsx` | Authentication-state gate around app content and login/account readiness. |
| `app/debug/DebugHarnessShell.tsx` | Shared shell for development/debug harness pages. |
| `app/DeploymentAuditPanel.tsx` | Deployment administrative audit panel and audit-event formatting. |
| `app/IncidentImportPanel.tsx` | Incident bundle import panel and import-job polling controls. |
| `app/IncidentAdminPanel.tsx` | Incident administration panel UI for incident metadata, preferences, membership, and audit affordances. |
| `app/IncidentLanding.tsx` | Incident directory landing panel, search/filter controls, and create-incident dialog. |
| `app/LandingAdminDisplay.tsx` | Shared display helpers for landing/admin panels. |
| `app/LandingAdminLayout.tsx` | Landing and deployment-administration shell layout plus the account/application menu. |
| `app/debug/AuthenticationDebugHarness.tsx` | Authentication debug harness entrypoint for auth, account, and route readiness scenarios. |
| `app/debug/IncidentDirectoryDebugHarness.tsx` | Incident-directory debug harness entrypoint for incident setup and preference scenarios. |
| `app/referencePackAdminClient.ts` | Reference-pack administration HTTP client helpers. |
| `app/referencePackAdminModel.ts` | Reference-pack administration resource, query, paging, and session-shape types. |
| `app/ReferencePackAdminPanel.tsx` | Reference-pack administration panel UI for import, reload, cancellation, and job-status controls. |
| `app/landingAdminStyles.ts` | Shared style constants for landing/admin panels. |
| `app/landingAdminTypes.ts` | Shared landing/admin TypeScript model types. |
| `app/routeState.ts` | Pure app-route parsing and history URL construction helpers. |
| `app/useAppRouteRuntime.ts` | React hook for route state, popstate handling, and history writes. |
| `app/App.landing.test.tsx` | Landing-surface tests for app startup and landing interactions. |
| `app/App.auth.support.test.tsx` | Support tests for Authentication app harness behavior. |
| `app/App.auth.test.tsx` | Authentication app behavior tests for auth/account/route readiness. |
| `app/App.timeline-invalidation.support.test.tsx` | General app-shell behavior tests. |
| `app/IncidentAdminPanel.test.tsx` | Incident administration panel tests. |
| `app/ReferencePackAdminPanel.test.tsx` | Reference-pack administration panel tests. |
| `app/api/appShellClient.routeBoundary.test.ts` | Route-boundary tests for app-shell client helpers. |
| `app/fontBundle.test.ts` | Font bundle availability and packaging boundary tests. |
| `app/fontRoles.test.tsx` | Font-role presentation tests for app and workbook surfaces. |
| `app/otelBoundary.test.ts` | OpenTelemetry import and runtime-boundary tests. |
| `app/routeState.test.ts` | Route-state parsing and history write tests. |

## `collaboration/`

The collaboration directory owns the browser-tab incident session facade. It
owns transport lifecycle and typed publication only; feature interpreters own
Timeline and extension effects.

| File | Responsibility |
| --- | --- |
| `collaboration/IncidentCollaborationSession.tsx` | Incident-scoped WebSocket provider for hello/resume, private resume state, one sequence high-water mark, reconnect, heartbeat, safe decoding, presence publication, reset, revocation, and closure events. |
| `collaboration/IncidentCollaborationSession.test.tsx` | Tests for single-socket lifetime, surface presence changes, replay deduplication/gaps, heartbeat, and unknown-message safety. |

## `extensions/`

The `extensions/` directory owns client-side extension discovery and
availability state plus stable extension workspace identities. Feature
implementations consume this boundary but do not redefine availability.

| File | Responsibility |
| --- | --- |
| `extensions/ExtensionAvailabilityContext.tsx` | React provider and required-consumer hook for the extension availability controller. |
| `extensions/extensionAvailability.ts` | Strict support-registry and workspace-availability decoding plus generation-bound, fail-closed extension availability coordination. |
| `extensions/extensionWorkspaceIdentities.ts` | Stable extension profile, route-family, workspace, and sheet-reference identities. |
| `extensions/extensionAvailability.test.ts` | Tests for support-registry decoding, exact availability intersections, stale-response rejection, generation rollover, and fail-closed behavior. |

## `imports/`

The `imports/` directory owns browser-side Import workflow choreography.
Transport mechanics and generated protocol aliases remain in `services/`;
target-specific mapping interpretation remains with the consuming feature.

| File | Responsibility |
| --- | --- |
| `imports/importCoordinator.ts` | Paginated discovery, preview, mapping approval, selection, apply, polling, and cancellation orchestration over validated Import/job operations. |
| `imports/importCoordinator.test.ts` | Characterization for opaque cursor traversal, exact envelopes, preview/approval separation, stale fingerprints, CSRF, and public errors. |

## `networkFlow/`

Network Flow is an extension feature. It owns Network Analysis behavior and
typed presentation adaptation, but it is not a Base Profile workbook surface
or `view_schema` owner. Workbook composition consumes it only through
`workbook/features/NetworkFlowFeature.tsx`.

| File | Responsibility |
| --- | --- |
| `networkFlow/NetworkAnalysisWorkspace.tsx` | Network Analysis presentation/composition facade over feature-specific controllers. |
| `networkFlow/NetworkFlowGridLoadFixture.tsx` | Debug-only deterministic supported-load fixture composed from the production Network Flow grid components. |
| `networkFlow/NetworkFlowMappingModal.tsx` | Explicit ordinal-aware Network Flow mapping, safe preview, and approval dialog. |
| `networkFlow/NetworkFlowQueryControls.tsx` | Accepted-row and rejected-row filter, sort, time-window, and reset controls. |
| `networkFlow/NetworkFlowSemanticGrid.tsx` | Semantic accepted-row, rejected-row, and contributor grids with layout controls, selection, focus recovery, and inspector presentation. |
| `networkFlow/networkFlowBoundaryPolicy.test.ts` | Static enforcement for controller composition, generated-decoder ownership, and browser projection-input exclusion. |
| `networkFlow/networkFlowClient.ts` | Decoded Network Flow feature operations; browser routing remains owned by the app/workbook route seam. |
| `networkFlow/networkFlowCollaborationInterpreter.ts` | Feature-local interpreter for decoded Network Flow extension invalidation/removal events. |
| `networkFlow/networkFlowController.ts` | Pure Network Flow table selection and refresh/removal state reducer. |
| `networkFlow/networkFlowErrors.ts` | Feature-local authorization-loss classification shared by query controllers. |
| `networkFlow/networkFlowImportModel.ts` | Generated-registry mapping suggestions and candidate construction for discovered source ordinals. |
| `networkFlow/networkFlowIndicatorLinkModel.ts` | Semantic grid-selection validation and immutable row-selector construction for indicator links. |
| `networkFlow/networkFlowPresentation.tsx` | Extension-owned grid metadata, semantic row projections, value formatting, diagnostic localization, and column compilation. |
| `networkFlow/networkFlowQueryModel.ts` | Accepted/rejected query request compilation, continuation construction, and owner-identity result reconciliation. |
| `networkFlow/useNetworkFlowCollaborationController.ts` | Owner invalidation/removal effects over the shared collaboration session. |
| `networkFlow/useNetworkFlowExtensionEvents.ts` | Network Flow subscription adapter over the shared incident collaboration session. |
| `networkFlow/useNetworkFlowGraphController.ts` | Graph and contributor query state, cancellation, selection, and stale-result rejection. |
| `networkFlow/useNetworkFlowGridLayout.ts` | Session-lifetime Network Flow column visibility, order, width, and reset state. |
| `networkFlow/useNetworkFlowImportController.ts` | Staged Network Flow discovery, preview, fingerprint-bound approval/apply, and returned-table selection controller. |
| `networkFlow/useNetworkFlowIndicatorLinkController.ts` | Explicit graph-edge indicator-link command controller. |
| `networkFlow/useNetworkFlowModalFocus.ts` | Modal initial focus, focus trapping, Escape dismissal, and focus restoration hook. |
| `networkFlow/useNetworkFlowPagedQuery.ts` | Abortable cursor paging, previous/next history, invalid-cursor recovery, and protected-state clearing hook. |
| `networkFlow/useNetworkFlowRejectedRowsController.ts` | Rejected-row query state and cancellation controller. |
| `networkFlow/useNetworkFlowRowsController.ts` | Accepted table-row query state and cancellation controller. |
| `networkFlow/useNetworkFlowTableController.ts` | Table discovery, active selection, load state, and access-loss controller. |
| `networkFlow/NetworkAnalysisWorkspace.test.tsx` | Network Analysis workspace behavior tests. |
| `networkFlow/NetworkFlowSemanticGrid.test.tsx` | Semantic-grid accessibility tests for focus recovery and keyboard column reordering announcements. |
| `networkFlow/networkFlowCollaborationInterpreter.test.ts` | Network Flow collaboration event admission tests. |
| `networkFlow/networkFlowController.test.ts` | Network Flow controller lifecycle tests. |
| `networkFlow/networkFlowErrors.test.ts` | Tests for structured Network Flow error preservation, safe messages, and lifecycle/access-loss classification. |
| `networkFlow/networkFlowImportModel.test.ts` | Ordinal identity, registry suggestion, policy-accounting, and timestamp candidate tests. |
| `networkFlow/networkFlowIndicatorLinkModel.test.ts` | Tests for semantic IP-cell/range selection, row-ref construction, and link-admission limits. |
| `networkFlow/networkFlowPresentation.test.tsx` | Tests for extension grid metadata, row identities, formatting, diagnostic localization, grouping, and projection reuse. |
| `networkFlow/networkFlowQueryModel.test.ts` | Tests for exact initial and continuation queries plus owner-identity reconciliation. |
| `networkFlow/useNetworkFlowExtensionEvents.test.tsx` | Shared-session Network Flow reconnect and sequence-deduplication tests. |
| `networkFlow/useNetworkFlowGridLayout.test.tsx` | Tests for Network Flow session layout mutation, reset, and remount persistence. |
| `networkFlow/useNetworkFlowPagedQuery.test.tsx` | Tests for cursor history, invalid-cursor recovery, request cancellation, and late-response rejection. |

## `services/`

The `services/` directory owns browser transport helpers and small client
facades used by app/workbook code. It should not encode domain behavior beyond
request/response handling already owned by specs and backend contracts.

| File | Responsibility |
| --- | --- |
| `services/browserApi.ts` | Browser API base/path helpers for app-local HTTP calls. |
| `services/clientTransactionId.ts` | Secure prefixed client transaction ID generation using Web Crypto UUIDs or RFC 4122 v4 fallback formatting. |
| `services/extensionContractAdapter.ts` | Thin generated-protocol facade for extension resource types and packaged contract-artifact parsing. |
| `services/httpTransport.ts` | Same-origin JSON and multipart transport mechanics for credentials, CSRF, cancellation, parsing, optional runtime decoding, and sanitized contract failures. |
| `services/importContractAdapter.ts` | Thin generated-protocol alias facade for Import and common-job operations and workflow resource types. |
| `services/importTargetContractAdapter.ts` | Thin generated-protocol facade for Import target discovery and target-specific mapping contracts. |
| `services/networkFlowContractAdapter.ts` | Thin post-decode Network Flow presentation type and decoder facade; contains no handwritten wire model. |
| `services/importTargetContractAdapter.test.ts` | Tests for strict Import target contract decoding and generated-protocol alignment. |
| `services/networkFlowContractAdapter.test.ts` | Network Flow claimed-profile and compiled-contract-major admission tests. |
| `services/workbookApi.ts` | Workbook HTTP helper utilities, envelope parsing, abort/query runtime helpers, and user-facing error extraction. |
| `services/workbookEvidence.ts` | Evidence upload/attach client helpers and evidence public-error mapping. |
| `services/browserApi.test.ts` | Tests for browser API base/path helpers. |
| `services/clientTransactionId.test.ts` | Tests for platform UUID use, secure-random fallback formatting, prefixes, and unavailable-crypto failure. |
| `services/clientTransactionIdPolicy.test.ts` | Static policy test excluding counters, clocks, and insecure randomness from browser mutation IDs. |
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
| `testing/TimelineWorkbookRuntimeFixture.tsx` | Production-composed Timeline workbook runtime fixture with configurable shell, layout, query, entity, and interaction inputs. |
| `testing/appShellTestSupport.ts` | Shared app-shell test helpers and fixtures. |
| `testing/extensionAvailabilityTestSupport.ts` | Deterministic ready extension-availability controller fixture for workbook and feature tests. |
| `testing/fetchMockTestSupport.ts` | Fetch mock helpers for unit and integration-style frontend tests. |
| `testing/selectorContractPolicy.test.ts` | Selector ownership policy test that guards raw `data-testid` literals and shared selector facade usage. |
| `testing/sourceOwnershipPolicy.test.ts` | Exact non-Markdown frontend source-ownership parity and manifest-shape policy. |
| `testing/transportBoundaryPolicy.test.ts` | Static same-origin transport policy; raw fetch is limited to the shared transport and server-issued Evidence upload target. |
| `testing/testSetup.dom.ts` | DOM-specific Vitest setup. |
| `testing/testSetup.ts` | Common Vitest setup for frontend tests. |
| `testing/timelineWorkbookTestSupport.ts` | Shared Timeline workbook fixture helpers, route mocks, and row builders for tests. |
| `testing/timelineWorkbookRenderTestSupport.tsx` | Shared Timeline render helpers used by component characterization tests. |
| `testing/timelineWorkbookTestSupport.test.tsx` | Tests for Timeline workbook test-support helpers. |
| `testing/workbookInspectorTestSupport.ts` | Entity-inspector readiness waits and diagnostics keyed by stable surface, record, and row-version identity. |
| `testing/workbookInspectorTestSupport.test.tsx` | Tests for delayed entity-inspector readiness and safe subject diagnostics. |

## `workbook/`

The `workbook/` directory owns the workbook shell, shell-level tests, and
workbook subfolders. Runtime code in this directory should use package facades
for grid, protocol, UI contracts, and view contracts rather than generated
internals.

| File | Responsibility |
| --- | --- |
| `workbook/WorkbookShell.tsx` | Workbook shell coordinator. Composes surfaces, saved-view/query controls, incident controls, generic surfaces, assessment/entity support, and Timeline entrypoints. |
| `workbook/WorkbookShell.assessments.test.tsx` | Assessment workbook behavior tests. |
| `workbook/WorkbookShell.autosave.test.tsx` | Timeline mutation autosave and pending-save characterization. |
| `workbook/WorkbookShell.grid.test.tsx` | Timeline mutation grid/create behavior characterization. |
| `workbook/WorkbookShell.payload.test.tsx` | Timeline mutation request payload characterization. |
| `workbook/WorkbookShell.actionSequencing.test.tsx` | Entity linking action sequencing tests. |
| `workbook/WorkbookShell.saveState.test.tsx` | Entity linking save-state tests. |
| `workbook/WorkbookShell.support.test.tsx` | Entity linking support tests for Timeline/workbook interactions. |
| `workbook/WorkbookShell.timelineQuery.test.tsx` | Timeline query and view-row normalization tests. |
| `workbook/WorkbookShell.gridProvenance.test.tsx` | Grid provenance and contract-row tests for Evidence lifecycle surfaces. |
| `workbook/WorkbookShell.mentionChips.test.ts` | Mention chip model tests. |
| `workbook/WorkbookShell.evidence.test.tsx` | Evidence lifecycle workbook/evidence behavior tests. |
| `workbook/WorkbookShell.collaboration.test.tsx` | Collaboration collaboration/session and pending behavior tests. |
| `workbook/WorkbookShell.history.test.tsx` | History and revision history and inspector behavior tests. |
| `workbook/WorkbookShell.query.test.tsx` | Saved-view and query saved-view/query behavior tests. |
| `workbook/WorkbookShell.inspector.test.tsx` | Workbook interaction inspector and row-local action tests. |
| `workbook/WorkbookShell.sentinel.test.tsx` | Workbook interaction sentinel, focus, and continuity tests. |
| `workbook/WorkbookShell.surfaces.test.tsx` | Multi-surface workbook shell tests. |

### `workbook/components/`

Shell-level presentation components live here. They should receive behavior
through props and local workbook models rather than own transport or domain
workflow logic.

| File | Responsibility |
| --- | --- |
| `workbook/components/ActiveSurfaceSavedViewSelector.tsx` | Saved-view selector for the active workbook surface. |
| `workbook/components/AssessmentWorkbookSurface.tsx` | Assessment presentation facade for rows, support selection, and semantic assessment commands. |
| `workbook/components/EntityWorkbookSurface.tsx` | Hosts and Identities presentation facade over entity query, mutation, inspector, and continuity owners. |
| `workbook/components/GenericMutationControl.tsx` | Generic row mutation controls for system-view surfaces. |
| `workbook/components/GenericWorkbookSurface.tsx` | Common contract-backed grid presentation. Domain mutation execution is injected through named owner command ports. |
| `workbook/components/IncidentControlsDrawer.tsx` | Shell-level incident controls drawer presentation and focus boundary. |
| `workbook/components/WorkbookRecordCandidatePicker.tsx` | Shared semantic record-candidate selection control for owner workflows. |
| `workbook/components/SystemViewSwitcher.tsx` | System-view switcher UI and grouped surface navigation. |
| `workbook/components/WorkbookConflictResolver.tsx` | Common typed conflict resolver and recovery presentation for every writable Base renderer. |
| `workbook/components/WorkbookGridEditorControl.tsx` | Contract-field grid editor adapter, mutation controls, commit/cancel behavior, and editor-kind selection. |
| `workbook/components/WorkbookGridControls.tsx` | Reusable workbook grid filter/sort/grouping control shell. |
| `workbook/components/WorkbookInspectorFeatureGroups.tsx` | Inspector feature-group renderer and disabled-state presentation helpers. |
| `workbook/components/WorkbookPresenceMarkers.tsx` | Shared row-gutter and cell presence markers with design-owned capacity and overflow behavior. |
| `workbook/components/WorkbookViewBar.tsx` | Shared saved-view, query, inspector, and create control composition. |
| `workbook/components/WorkbookShellSlots.tsx` | Stable shell slot IDs, labels, and layout slot helpers. |
| `workbook/components/WorkbookStatusStrip.tsx` | Status strip presentation for save/load/selection state. |

### `workbook/surfaces/`

The surfaces composition facade selects a registered renderer by stable
`view_schema_id`. Concrete surface imports remain private to this boundary;
the shell supplies cohesive owner snapshots and command ports.

| File | Responsibility |
| --- | --- |
| `workbook/surfaces/WorkbookSurfacesFacade.tsx` | Registration-driven active-surface selection and adaptation of view-state, query, mutation, collaboration, inspector, continuity, and layout owners into concrete renderer inputs. |

### `workbook/evidence/`

Workbook-owned Evidence presentation adapts transport outcomes and lifecycle
view models for Workbook surfaces. It performs no transport.

| File | Responsibility |
| --- | --- |
| `workbook/evidence/evidenceAccessPresentation.ts` | Evidence access message severity and live-region presentation mapping. |
| `workbook/evidence/evidenceAccessPresentation.test.ts` | Tests for blocking and non-blocking Evidence live-region outcomes. |

### `workbook/hooks/`

Shell-level hooks coordinate reusable runtime behavior. Hooks in this folder
should remain workbook-generic unless their name clearly scopes them to a
specific surface.

| File | Responsibility |
| --- | --- |
| `workbook/hooks/useAssessmentSupportCandidates.ts` | Loads and normalizes Timeline support candidates needed by assessment workflows. |
| `workbook/hooks/useEntityTimelinePreview.ts` | Loads Timeline preview rows for entity-related workbook workflows. |
| `workbook/hooks/useGenericSurfaceMutationController.ts` | Contract-surface mutation state, conflict admission, and refresh coordination over the generic command port. |
| `workbook/hooks/useIncidentControlsDrawer.ts` | Incident controls drawer state, selection, and focus restoration. |
| `workbook/hooks/useOwnerReferenceOptions.ts` | Resolves only the active bounded-context policy's reference requirements through the generic broker. |
| `workbook/hooks/useWorkbookIncidentIdentity.ts` | Resolves incident identity/loading state for the workbook shell. |
| `workbook/hooks/useWorkbookPendingGridFocus.ts` | Restores the requested first grid target after a surface transition. |
| `workbook/hooks/useWorkbookProjectionRefreshController.test.tsx` | Direct tests for initial and sheet-triggered projection refresh ownership. |
| `workbook/hooks/useWorkbookProjectionRefreshController.ts` | Initial session/entity and sheet-triggered projection refresh coordinator. |
| `workbook/hooks/useWorkbookQueryController.test.tsx` | Direct tests for exact-view-schema query-state isolation. |
| `workbook/hooks/useWorkbookQueryController.ts` | Schema-keyed query/filter controller and active query-control adapter. |
| `workbook/hooks/useWorkbookSavedViewController.test.tsx` | Direct tests for saved-view loading and selection precedence. |
| `workbook/hooks/useWorkbookSavedViewController.ts` | Saved-view list, CRUD, persistence, and active-selection controller. |
| `workbook/hooks/useWorkbookShellRuntime.ts` | Thin shell runtime facade composing startup, saved-view, and query controllers. |
| `workbook/hooks/useWorkbookStartupController.test.tsx` | Direct tests for workbook selection, focus intent, versioning, and URL state. |
| `workbook/hooks/useWorkbookStartupController.ts` | Startup/sheet identity, URL history, focus intent, and workbook-preference controller. |

### `workbook/inspector/`

| Path | Responsibility |
| --- | --- |
| `workbook/inspector/useWorkbookInspectorCoordinator.ts` | Schema-bound inspector lifecycle coordinator for explicit open/close, stable row subjects, ordered feature invalidation, action completion, and focus restoration ports. |
| `workbook/inspector/useWorkbookInspectorCoordinator.test.tsx` | Direct tests for retargeting, lifecycle invalidation, action completion, idempotent close, and focus restoration. |

### `workbook/lifecycle/`

The lifecycle leaf owns typed invalidation reasons only. It contains no state,
event bus, or generic clear-all operation.

| File | Responsibility |
| --- | --- |
| `workbook/lifecycle/workbookInvalidation.ts` | Shared typed Workbook and extension invalidation reason unions for applicable lifecycle owners. |

### `workbook/layout/`

Workbook layout owns responsive classification, density, column geometry,
shell/work-area sizing, inspector geometry, and shared style slots. Surface
owners consume semantic snapshots and commands without deriving geometry from
row counts or viewport subtraction.

| File | Responsibility |
| --- | --- |
| `workbook/layout/WorkbookSurfaceLayout.tsx` | Shared work-area, independently scrolling grid/inspector slots, overlay geometry, resize behavior, and focus restoration. |
| `workbook/layout/useWorkbookColumnLayoutController.ts` | Per-schema column state, active column commands, saved-layout application, and startup resets. |
| `workbook/layout/useWorkbookLayoutFacade.ts` | Composition facade for effective density, responsive mode, interaction mode, column state, and surface layout commands. |
| `workbook/layout/useWorkbookResponsiveLayout.ts` | Viewport subscription and semantic responsive-layout snapshot. |
| `workbook/layout/workbookColumnLayout.ts` | Contract-normalized column ordering, visibility, width, movement, and materialization helpers. |
| `workbook/layout/workbookDensity.ts` | Account density preference resolution. |
| `workbook/layout/workbookResponsiveLayout.ts` | Responsive layout classification and surface-band helpers. |
| `workbook/layout/workbookShellStyles.ts` | Shared shell chrome, work-area, viewport-overlay, and responsive style slots. |
| `workbook/layout/workbookLayoutPolicy.test.ts` | Architecture checks prohibiting viewport-subtraction, row-count geometry, synthetic rows, and surface-private minimum-height workarounds. |
| `workbook/layout/workbookDensity.test.ts` | Tests for effective Workbook density. |
| `workbook/layout/workbookResponsiveLayout.test.ts` | Tests for responsive classification and query-control capacity. |

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
| `workbook/models/workbookGridState.ts` | Contract-grid load-state presentation and incident-role interaction-mode helpers. |
| `workbook/models/workbookIncidentIdentity.ts` | Incident identity normalization and loading-state model. |
| `workbook/models/workbookInspectorModel.ts` | Pure inspector state machine for default-closed state, semantic subjects, active panels, no-row state, and invalidation generations. |
| `workbook/models/workbookQuery.ts` | Workbook query, filter, sort, grouping, and request-building helpers. |
| `workbook/models/workbookReferenceOptions.ts` | Reference option normalization and lookup helpers. |
| `workbook/models/workbookSavedViewRuntime.ts` | Saved-view runtime selection, dirty-state, and command helpers. |
| `workbook/models/workbookSavedViews.ts` | Saved-view resource normalization and payload helpers. |
| `workbook/models/workbookStartup.ts` | Workbook startup candidate, selected sheet reference, and fallback resolution helpers. |
| `workbook/models/workbookSurfaceRegistry.ts` | Built-in/system/optional workbook surface registry and stable view-schema IDs. |
| `workbook/models/workbookSurfaceQueryRuntime.ts` | Latest-query cancellation and context-token runtime for workbook projection loads. |
| `workbook/models/workbookSurfaceRegistration.ts` | Exact `view_schema_id` registration and bounded-context policy registry for all exposed workbook schemas. |
| `workbook/models/assessmentWorkbookModel.test.ts` | Tests for assessment workbook model helpers. |
| `workbook/models/entityWorkbookModel.test.ts` | Tests for entity workbook model helpers. |
| `workbook/models/evidenceLifecycleViewModel.test.ts` | Tests for evidence lifecycle view-model helpers. |
| `workbook/models/genericWorkbookModel.test.ts` | Tests for generic workbook model helpers. |
| `workbook/models/workbookInspectorModel.test.ts` | Tests for workbook inspector state/model helpers. |
| `workbook/models/workbookQuery.test.ts` | Tests for workbook query helpers. |
| `workbook/models/workbookReferenceOptions.test.ts` | Tests for reference-option helpers. |
| `workbook/models/workbookSavedViewRuntime.test.ts` | Tests for saved-view runtime helpers. |
| `workbook/models/workbookSavedViews.test.ts` | Tests for saved-view normalization and payload helpers. |
| `workbook/models/workbookStartup.test.ts` | Tests for workbook startup resolution. |
| `workbook/models/workbookSurfaceRegistry.test.ts` | Tests for workbook surface registry invariants. |
| `workbook/models/workbookSurfaceRegistration.test.ts` | Tests for policy registration completeness, uniqueness, and extension-workspace exclusion. |

### `workbook/mutations/`

Workbook mutation assembly owns secure logical-action identity and exact
owner-specific request construction. Presentation receives named semantic
command ports; it never receives the transaction-ID provider or a generic
mutation transport.

| File | Responsibility |
| --- | --- |
| `workbook/mutations/secureTransactionId.ts` | Private browser assembly adapter for secure transaction-ID creation. |
| `workbook/mutations/workbookMutationCommandPorts.ts` | Named Timeline, generic, entity, assessment, Evidence, and coordination command-port contracts. |
| `workbook/mutations/createWorkbookMutationCommandPorts.ts` | Incident-scoped command assembly and exact owner request construction over the private secure-ID port. |
| `workbook/mutations/createWorkbookMutationCommandPorts.test.ts` | Exact owner payload, logical identity, and local secure-random failure tests. |

### `workbook/features/`

Feature facades are the only workbook-shell entrypoints for extension-owned
workspaces. They may compose feature internals but must not move extension
identity into the Base surface registry.

| File | Responsibility |
| --- | --- |
| `workbook/features/ImportAssistantFeature.tsx` | Availability-gated workbook import assistant for discovery, ordinal mapping, approval, unit selection, apply/cancel, and result navigation. |
| `workbook/features/ImportAssistantFeature.test.tsx` | Import assistant discovery, approval, cancellation, and returned-selection characterization. |
| `workbook/features/NetworkFlowFeature.tsx` | Workbook/app-facing Network Flow facade for workspace rendering, debug-fixture composition, and stable extension identity. |
| `workbook/features/coordination/CoordinationWorkflowBindings.tsx` | Coordination-owned task lifecycle and decision supersession presentation over semantic commands. |
| `workbook/features/evidence/useEvidenceWorkbookBindings.tsx` | Evidence-owned access, preview, download, and semantic attachment binding for the contract surface. |

### `workbook/policies/`

Pure bounded-context policy declarations own exact-schema create defaults,
advisory minimums, collection semantics, reference requirements, refresh
consequences, public-error presentation, and owner UI-binding declarations.
They must not perform transport or authorization.

| File | Responsibility |
| --- | --- |
| `workbook/policies/artifactSurfacePolicies.ts` | Notes, Findings, Investigative Queries, and Forensic Keywords policy owner. |
| `workbook/policies/assessmentSurfacePolicies.ts` | Assessments policy owner. |
| `workbook/policies/captureTimelineSurfacePolicies.ts` | Timeline surface policy owner. |
| `workbook/policies/coordinationSurfacePolicies.ts` | Parties, Tasks, Decisions, Communications, Handoff, Status Review, and Lessons policy owner. |
| `workbook/policies/entitiesObservationsSurfacePolicies.ts` | Hosts, Identities, and Indicators policy owner. |
| `workbook/policies/evidenceSurfacePolicies.ts` | Evidence policy owner. |
| `workbook/policies/workbookSurfaceOwnershipPolicy.test.ts` | Static policy-purity, owner-completeness, and common-surface boundary checks. |
| `workbook/policies/workbookSurfacePolicy.ts` | Pure `WorkbookSurfacePolicy` types, immutable defaults, and stable reference declarations. |

### `workbook/services/`

Workbook services execute generic mechanics declared by owners. They do not
choose domain references or authorize access.

| File | Responsibility |
| --- | --- |
| `workbook/services/referenceQueryBroker.ts` | Creates instance-scoped, incident/authorization-bound reference-query ports with shared-consumer deduplication, typed invalidation, abort ownership, and idempotent disposal. |
| `workbook/services/referenceQueryBroker.test.ts` | Tests in-flight deduplication, shared-consumer cancellation, two-shell isolation, context binding, invalidation, teardown, and late-result rejection. |

### `workbook/query/`

Workbook query owners bind request admission, live reconciliation, protected
state cleanup, and teardown to the Workbook instance that consumes them.
Shared query rows are application contracts and do not belong to Timeline.

| File | Responsibility |
| --- | --- |
| `workbook/query/WorkbookQueryRow.ts` | Shared schema-keyed view-query row shape below Timeline ownership. |
| `workbook/query/workbookQueryRowPatch.ts` | Pure sparse-patch application for shared query rows. |
| `workbook/query/useAssessmentSurfaceQuery.ts` | Assessment-owned query construction, admission, live reconciliation, refresh, access cleanup, and cancellation. |
| `workbook/query/useAssessmentSurfaceQuery.test.tsx` | Direct rapid-filter, stale-error, live-patch, access-loss, inactive-surface, and teardown characterization. |
| `workbook/query/useEntitySurfaceQuery.ts` | Instance-owned dual host/identity loading, indexing, live patching, refresh, cancellation, and protected-state cleanup. |
| `workbook/query/useEntitySurfaceQuery.test.tsx` | Direct dual-load, stale-result, live-patch, access-loss, explicit cleanup, and teardown characterization. |
| `workbook/query/useGenericSurfaceQuery.ts` | Schema-keyed generic query admission, schema matching, Notes normalization, live reconciliation, stale-row retention, access cleanup, and cancellation. |
| `workbook/query/useGenericSurfaceQuery.test.tsx` | Direct Notes normalization, schema-switch, mismatched-envelope, stale-row, access-loss, inactive-surface, and teardown characterization. |

### `workbook/startup/`

Workbook startup admission owns transport ordering, availability reservation,
saved-view hydration precedence, and stale-result rejection. Selection state
and availability state remain behind injected ports.

| File | Responsibility |
| --- | --- |
| `workbook/startup/useWorkbookStartupAdmission.ts` | Startup request and response admission boundary with selection-version, incident, ordinal, availability, and cancellation guards. |
| `workbook/startup/useWorkbookStartupAdmission.test.tsx` | Startup admission characterization for precedence, overlap, extension availability, fallbacks, teardown, and late work. |

### `workbook/timeline/`

Timeline-specific implementation lives under `workbook/timeline/`. This folder
owns Timeline presentation, Timeline hooks, Timeline models, and Timeline
services. Generic workbook behavior belongs one level up unless it is truly
Timeline-specific. `TimelineWorkbook.tsx` is the Timeline composition root;
feature-specific coordination should live in the focused components, hooks,
models, or services below.

#### `workbook/timeline/components/`

| File | Responsibility |
| --- | --- |
| `workbook/timeline/components/TimelineCellEditors.tsx` | Timeline scalar editors, draft-row create button, and relationship chip presentation. |
| `workbook/timeline/components/TimelineEvidencePanel.test.tsx` | Tests for Timeline evidence panel behavior. |
| `workbook/timeline/components/TimelineEvidencePanel.tsx` | Timeline inspector evidence panel and evidence actions UI. |
| `workbook/timeline/components/TimelineGridSurface.tsx` | Timeline grid surface wrapper around the grid adapter boundary. |
| `workbook/timeline/components/TimelineHistoryPanel.tsx` | Timeline row history, rollback, delete, restore, and history action presentation. |
| `workbook/timeline/components/TimelineMentionsPanel.tsx` | Timeline mention-resolution inspector panel. |
| `workbook/timeline/components/TimelineRowActions.tsx` | Timeline row action/context-menu presentation. |
| `workbook/timeline/components/TimelineWorkbook.tsx` | Timeline composition root. Wires runtime state, focused controllers, grid, inspector, notices, and shell callbacks while leaving feature-specific logic in local hooks/components. |
| `workbook/timeline/components/TimelineWorkbookGrid.tsx` | Timeline grid renderer, grouped row table wrapper, hidden contract metadata cells, and grid test-ID placement. |
| `workbook/timeline/components/TimelineWorkbookInspector.tsx` | Timeline inspector shell, panel tabs, disabled-state presentation, selected-row state, and inspector messages. |
| `workbook/timeline/components/TimelineWorkbookInspectorSections.tsx` | Timeline inspector section factories for field editors, relationships, evidence attach, related-row creation, and row history. |
| `workbook/timeline/components/TimelineWorkbookNotices.tsx` | Timeline notice overlay for auto-resolution notices, pending queue messages, and queued-edit counts. |
| `workbook/timeline/components/TimelineWorkbookRenderers.tsx` | Timeline grid and inspector editor renderer factory, column materialization, relationship controls, and cell presence wiring. |
| `workbook/timeline/components/TimelineWorkbookStyles.ts` | Timeline-specific style constants shared by Timeline workbook components. |

#### `workbook/timeline/hooks/`

| File | Responsibility |
| --- | --- |
| `workbook/timeline/hooks/useTimelineClipboardPasteController.ts` | Coordinates Timeline clipboard paste dispatch, payload construction, conflict registration, scalar fallback, and post-paste focus/viewport restoration. |
| `workbook/timeline/hooks/useTimelineCommittedRows.ts` | Derives committed Timeline row collections from row/runtime state. |
| `workbook/timeline/hooks/useTimelineConflicts.ts` | Coordinates Timeline same-field conflict state. |
| `workbook/timeline/hooks/useTimelineCreateRelatedWorkflow.ts` | Coordinates Timeline inspector related-row create workflow state, draft values, and payload submission. |
| `workbook/timeline/hooks/useTimelineEvidenceActions.ts` | Coordinates Timeline evidence attach/preview/download action state. |
| `workbook/timeline/hooks/useTimelineEvidenceAttach.ts` | Coordinates Timeline evidence file attachment, validation feedback, and reload/save-state handoff. |
| `workbook/timeline/hooks/useTimelineGridAnchorController.ts` | Resolves Timeline grid anchors, paste targets, selected cells, and focus anchors across committed and draft rows. |
| `workbook/timeline/hooks/useTimelineGridInteractions.ts` | Coordinates Timeline grid refs, keyboard helpers, and grid interaction commands. |
| `workbook/timeline/hooks/useTimelineHistoryActions.ts` | Coordinates Timeline history rollback, delete, restore, preview, and confirmation actions. |
| `workbook/timeline/hooks/useTimelineHistoryState.ts` | Coordinates Timeline history panel and row-history state. |
| `workbook/timeline/hooks/useTimelineInspectorSelection.ts` | Coordinates selected Timeline row, row-bound feature invalidation, deleted-row history, and focus-safe inspector interactions. |
| `workbook/timeline/hooks/useTimelineMentionActions.ts` | Coordinates Timeline mention resolution, undo/review actions, and related inspector selection updates. |
| `workbook/timeline/hooks/useTimelineMentions.ts` | Coordinates Timeline mention-resolution state and actions. |
| `workbook/timeline/hooks/useTimelineMutationCommands.ts` | Coordinates Timeline scalar and relationship mutation commands, pending-save admission, and save lifecycle callbacks. |
| `workbook/timeline/hooks/useTimelinePendingReplayController.ts` | Coordinates pending-save replay admission, HTTP replay dispatch, socket transaction tracking, and reload scheduling. |
| `workbook/timeline/hooks/useTimelinePendingSaves.ts` | Coordinates Timeline pending-save queue runtime and replay admission. |
| `workbook/timeline/hooks/useTimelineRows.ts` | Coordinates Timeline row state, draft rows, and row reconciliation. |
| `workbook/timeline/hooks/useTimelineRowsLoader.ts` | Coordinates Timeline row loading, query aborts, runtime status, and row reconciliation callbacks. |
| `workbook/timeline/hooks/useTimelineSaveStatePresentation.ts` | Coordinates Timeline save-state labels, pending queue snapshot publication, refresh blocking, replay scheduling, and beforeunload warning state. |
| `workbook/timeline/hooks/useTimelineConflictProjectionAdapter.ts` | Timeline render-state adapter for shell-owned same-field conflict registration, projection, and resolution. |
| `workbook/timeline/hooks/useTimelineViewportContinuityController.ts` | Coordinates Timeline scroll snapshots, focus restoration, continuity tokens, and entity-refresh barriers. |
| `workbook/timeline/hooks/useTimelineWorkbookRuntime.ts` | Reduces Timeline lifecycle state and translates shell-owned query commands into deterministic runtime transitions. |

#### `workbook/timeline/models/`

| File | Responsibility |
| --- | --- |
| `workbook/timeline/models/timelineControllerPorts.ts` | Neutral capability-port contracts shared by isolated Timeline controllers. |
| `workbook/timeline/models/timelineHistoryModel.ts` | Timeline row-history normalization, pending-action labels, and history operation helpers. |
| `workbook/timeline/models/timelineRowsModel.ts` | Timeline row collection helpers and row-state utilities. |
| `workbook/timeline/models/timelineWorkbookSurfaceRuntime.ts` | Required shell-owned Timeline composition contract for incident, query, entity, layout, and access-loss services. |
| `workbook/timeline/models/timelineViewportContinuityModel.ts` | Timeline viewport continuity and entity-refresh barrier helpers. |
| `workbook/timeline/models/workbookMentionChips.ts` | Mention chip state, relationship-field keys, and mention display helpers. |
| `workbook/timeline/models/workbookRecordFreshness.ts` | Pure durable-identity and row-version freshness comparison leaf. |
| `workbook/timeline/models/workbookTimelineModel.ts` | Timeline row model, field bindings, payload builders, normalization, patch intents, and display helpers. |
| `workbook/timeline/models/timelineRowsModel.test.ts` | Tests for Timeline grid-row materialization. |
| `workbook/timeline/models/timelineWorkbookRuntime.test.ts` | Deterministic lifecycle transition traces for load, refresh, save, conflict, and recovery state. |
| `workbook/timeline/models/timelineViewportContinuityModel.test.ts` | Tests for viewport continuity and refresh barrier helpers. |
| `workbook/timeline/models/workbookRecordFreshness.test.ts` | Tests for comparable and non-comparable row-version freshness decisions. |
| `workbook/timeline/models/workbookTimelineModel.test.ts` | Tests for Timeline row, payload, binding, normalization, and display helpers. |

#### `workbook/timeline/services/`

Timeline services are behavior-named helpers for wire-message construction,
HTTP mutation dispatch, and socket lifecycle state. They are not backend route
owners.

| File | Responsibility |
| --- | --- |
| `workbook/timeline/services/timelineMutationRequests.ts` | Timeline pending-replay HTTP dispatch and timing event helper. |

### `workbook/collaboration/`

Workbook collaboration modules reconcile decoded session events into
Workbook-owned state. Authorization recovery is injected from app
composition; these modules do not import authorization transport.

| File | Responsibility |
| --- | --- |
| `workbook/collaboration/WorkbookCollaborationCoordinator.ts` | Sole shell-lifetime interpreter for incident collaboration events, exact sheet presence, ordered typed cleanup, injected authorization recovery, and active-surface reconciliation. |
| `workbook/collaboration/workbookCollaborationMessages.ts` | Workbook presence, live-row message, self-origin filtering, and mention-action payload helpers. |
| `workbook/collaboration/workbookSurfacePort.ts` | Active-surface identity and live-row reconciliation capabilities. |
| `workbook/collaboration/useWorkbookCollaborationCoordinator.ts` | React external-store subscription, Strict Mode-safe lifetime lease, and collaboration-session adapter for the coordinator. |
| `workbook/collaboration/WorkbookCollaborationCoordinator.test.ts` | Tests for presence, reset, cleanup ordering, authorization recovery, role downgrade, access loss, and late-work rejection. |
| `workbook/collaboration/workbookCollaborationMessages.test.ts` | Tests for Base, saved-view, and extension-workspace presence message construction. |

### `workbook/runtime/`

Workbook runtime modules contain shell-lifetime mutation state and
renderer-neutral conflict and lifecycle contracts. Runtime modules MUST NOT
import Timeline implementation.

| File | Responsibility |
| --- | --- |
| `workbook/runtime/WorkbookMutationRuntime.ts` | Owns the incident/client-scoped pending queue across Base surface mounts. |
| `workbook/runtime/workbookPendingReplayRuntime.ts` | Pending-replay runtime state, admission contracts, and refresh barriers. |
| `workbook/runtime/workbookConflictModel.ts` | Same-field conflict parsing and common envelope types. |
| `workbook/runtime/workbookLifecycleModel.ts` | Shared load, refresh, save, conflict, and recovery lifecycle reducer. |
| `workbook/runtime/useWorkbookMutationRuntime.ts` | React external-store subscription hook for shell-owned workbook mutation state. |
| `workbook/runtime/WorkbookMutationRuntime.test.ts` | Tests for shell-lifetime queue retention, autosave, refresh debt, conflicts, and mutation coordination. |
| `workbook/runtime/workbookConflictModel.test.ts` | Tests for conflict parsing, queue entries, resolution payloads, and collection actions. |

### `workbook/continuity/`

Workbook continuity modules expose stable semantic focus and selection identities.
DOM nodes, grid coordinates, viewport geometry, and restore snapshots remain private
to the Workbook grid adapter.

| File | Responsibility |
| --- | --- |
| `workbook/continuity/workbookContinuityPort.ts` | Semantic capture, focus, selection, clear, one-shot restore, and disposal contract over stable schema, record, and field identities. |
| `workbook/continuity/useWorkbookGridContinuity.tsx` | Workbook-private translation between the semantic continuity port and grid-adapter focus, scrolling, clipboard, and viewport behavior. |
| `workbook/continuity/gridViewportContinuity.ts` | Private viewport anchor capture, restoration, and rectangle visibility helpers used by Timeline continuity. |
| `workbook/continuity/workbookContinuityPort.test.ts` | Tests for opaque capture tokens, stable semantic identities, one-shot restoration, and idempotent cleanup. |
| `workbook/continuity/gridViewportContinuity.test.ts` | Tests for viewport capture, restoration, and visibility helpers. |

### `workbook/utils/`

Workbook utilities are reusable helpers that are not specific enough to belong
under `timeline/` and not broad enough for `shared/`.

| File | Responsibility |
| --- | --- |
| `workbook/utils/workbookClipboard.ts` | Clipboard grid-shape and paste helpers. |
| `workbook/utils/workbookKeyboard.ts` | Workbook keyboard command mapping. |
| `workbook/utils/workbookPendingQueue.ts` | Pending-save queue capacity, save-state, replay, conflict, and public-error helpers. |
| `workbook/utils/workbookPresence.ts` | Presence input/type helpers and presence matching helpers. |
| `workbook/utils/workbookRowReconciliation.ts` | Record-identity and row-version reconciliation that preserves unchanged row references. |
| `workbook/utils/workbookStyles.ts` | Shared workbook style primitives. |
| `workbook/utils/workbookValueFormat.ts` | Grid/workbook value formatting helpers. |
| `workbook/utils/GridAdapter.anchor.test.ts` | Workbook interaction grid-adapter anchor behavior tests. |
| `workbook/utils/workbookClipboard.test.ts` | Tests for clipboard helpers. |
| `workbook/utils/workbookKeyboard.test.ts` | Tests for keyboard command mapping. |
| `workbook/utils/workbookPendingQueue.test.ts` | Tests for pending-queue helpers. |
| `workbook/utils/workbookRowReconciliation.test.ts` | Tests for sparse row replacement, removal, drafts, and row-version reference reuse. |
| `workbook/utils/workbookValueFormat.test.ts` | Tests for value formatting helpers. |

### `workbook/view-state/`

Workbook view-state hooks own instance-scoped query and layout state. They
contain no module-global mutable defaults or cross-shell store.

| File | Responsibility |
| --- | --- |
| `workbook/view-state/useWorkbookQueryState.ts` | Instance-owned schema-keyed query defaults, reducer state, reset, and update operations. |
