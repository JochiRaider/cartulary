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
| `networkFlow/networkFlowClient.ts` | Decoded Network Flow feature operations over the profile-and-route-scoped owner-neutral browser transport. |
| `networkFlow/networkFlowClient.test.ts` | Tests for extension route admission, cancellation, CSRF, and generated response decoding at the client boundary. |
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
| `services/browserApi.ts` | Owner-neutral browser JSON, multipart, and generated-operation transport with path/query derivation, response validation, CSRF, and public-error extraction. |
| `services/clientTransactionId.ts` | Secure prefixed client transaction ID generation using Web Crypto UUIDs or RFC 4122 v4 fallback formatting. |
| `extensions/extensionAvailability.ts` | Extension discovery state and typed packaged client-support registry consumption. |
| `services/httpTransport.ts` | Same-origin JSON and multipart transport mechanics for credentials, CSRF, cancellation, parsing, optional runtime decoding, and sanitized contract failures. |
| `services/importContractAdapter.ts` | Thin generated-protocol alias facade for Import and common-job operations and workflow resource types. |
| `services/importTargetContractAdapter.ts` | Thin generated-protocol facade for Import target discovery and target-specific mapping contracts. |
| `services/networkFlowContractAdapter.ts` | Thin post-decode Network Flow presentation type and decoder facade; contains no handwritten wire model. |
| `services/importTargetContractAdapter.test.ts` | Tests for strict Import target contract decoding and generated-protocol alignment. |
| `services/networkFlowContractAdapter.test.ts` | Network Flow claimed-profile and compiled-contract-major admission tests. |
| `services/workbookEvidence.ts` | Evidence upload/attach client helpers and evidence public-error mapping. |
| `services/browserApi.test.ts` | Tests for browser API base/path helpers. |
| `services/clientTransactionId.test.ts` | Tests for platform UUID use, secure-random fallback formatting, prefixes, and unavailable-crypto failure. |
| `services/clientTransactionIdPolicy.test.ts` | Static policy test excluding counters, clocks, and insecure randomness from browser mutation IDs. |
| `services/workbookEvidence.test.ts` | Tests for evidence client helpers and error mapping. |

## `shared/`

The `shared/` directory owns cross-feature helpers that are not workbook
specific.

| File | Responsibility |
| --- | --- |
| `shared/publicError.test.ts` | Direct tests for allowlisted public detail projection and unsafe-message sanitization. |
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
| `workbook/WorkbookShell.tsx` | Route-facing Workbook composition over incident-scoped lifecycle and presentation owners. |
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

### `workbook/adapters/`

Workbook adapters are the private generated-protocol boundary. Composition
creates incident-bound adapters once and injects semantic ports into query,
startup, saved-view, preference, incident, paste, and pending-mutation
consumers. The private clipboard port deliberately carries generated request
vocabulary and typed result rows/conflicts so surface owners can validate and
decode them; HTTP status codes, envelopes, routes, and transport failures do
not escape this directory.

| File | Responsibility |
| --- | --- |
| `workbook/adapters/WorkbookClipboardPastePort.ts` | Exact private generated-vocabulary transport capability shared by paste-capable Workbook surface owners. |
| `workbook/adapters/createWorkbookClipboardPasteAdapter.ts` | Executes the sole generated clipboard-paste operation with secure transaction identity and exact response-surface validation. |
| `workbook/adapters/createWorkbookClipboardPasteAdapter.test.ts` | Tests exact request projection, raw typed results, invalid input, response correlation, secure-ID failure, and transport containment. |
| `workbook/adapters/workbookOperationExecutor.ts` | Executes the closed Workbook operation-ID set and converts validated success/error envelopes to semantic outcomes. |
| `workbook/adapters/workbookProtocolTypes.ts` | Private type-only projection of the exact generated Workbook request vocabulary; prevents protocol imports from leaking into owner models, controllers, or runtime code. |
| `workbook/adapters/workbookAdapterResult.ts` | Normalizes operation outcomes into shared semantic port results. |
| `workbook/adapters/createWorkbookViewQueryAdapter.ts` | Executes one abortable Workbook query boundary with exact incident/schema correlation and contract-row normalization. |
| `workbook/adapters/createWorkbookViewQueryAdapter.test.ts` | Tests projected query requests, aborts, malformed responses, and cross-context rejection. |
| `workbook/adapters/createWorkbookPendingMutationAdapter.ts` | Executes queued create/patch units with dispatch identity, response correlation, and view-contract normalization. |
| `workbook/adapters/createWorkbookPendingMutationAdapter.test.ts` | Tests create/patch projection, identity correlation, accepted rows, and semantic failure outcomes. |
| `workbook/adapters/createWorkbookSavedViewAdapter.ts` | Executes explicit saved-view paging and CRUD behind incident-bound semantic outcomes. |
| `workbook/adapters/createWorkbookSavedViewAdapter.test.ts` | Tests paging progress, CRUD correlation, versions, immutability, and malformed responses. |
| `workbook/adapters/createWorkbookStartupAdapter.ts` | Loads validated Workbook startup state and correlated extension availability. |
| `workbook/adapters/createWorkbookIncidentAdapter.ts` | Loads validated incident identity and memberships with exact resource correlation. |
| `workbook/adapters/createWorkbookPreferenceAdapter.ts` | Loads and updates correlated current/default Workbook preferences. |
| `workbook/adapters/createWorkbookStartupAndIncidentAdapters.test.ts` | Tests startup, incident, membership, and preference operation boundaries. |

### `workbook/components/`

Shell-level presentation components live here. They should receive behavior
through props and local workbook models rather than own transport or domain
workflow logic.

| File | Responsibility |
| --- | --- |
| `workbook/models/workbookClipboardPaste.ts` | Exact generated paste-capable view parser plus bounded column/target constructors and semantic surface validation. |
| `workbook/models/entityClipboardPastePlan.ts` | Pure Entity scalar-versus-batch paste planning with current-record, writable-field, grouping, create-capability, and exact-surface admission. |
| `workbook/models/entityClipboardPastePlan.test.ts` | Tests Entity scalar routing, exact-origin all-create requests, and fail-closed target/authority handling. |
| `workbook/components/ActiveSurfaceSavedViewSelector.tsx` | Saved-view selector for the active workbook surface. |
| `workbook/components/AssessmentWorkbookSurface.tsx` | Assessment presentation facade for rows, support selection, and semantic assessment commands. |
| `workbook/components/EntityWorkbookSurface.tsx` | Hosts and Identities presentation facade over entity query, mutation, inspector, and continuity owners. |
| `workbook/features/entities/useEntityClipboardPasteController.ts` | Executes pure Entity paste plans through scalar mutation or the shared exact clipboard transport, then projects owner-local feedback and refresh. |
| `workbook/components/GenericMutationControl.tsx` | Generic row mutation controls for system-view surfaces. |
| `workbook/components/GenericWorkbookSurface.tsx` | Common contract-backed grid presentation. Domain mutation execution is injected through named owner command ports. |
| `workbook/components/IncidentControlsDrawer.tsx` | Shell-level incident controls drawer presentation and focus boundary. |
| `workbook/components/WorkbookRecordCandidatePicker.tsx` | Shared semantic record-candidate selection control for owner workflows. |
| `workbook/components/SystemViewSwitcher.tsx` | System-view switcher UI and grouped surface navigation. |
| `workbook/components/WorkbookConflictResolver.tsx` | Common typed conflict resolver and recovery presentation for every writable Base renderer. |
| `workbook/components/WorkbookActiveQueryChips.tsx` | Responsive canonical group/sort/filter chip presentation over semantic command descriptors. |
| `workbook/components/WorkbookActiveSurfaceFrame.tsx` | Active-surface recovery boundary and mutually exclusive blocked/overflow/conflict presentation. |
| `workbook/components/WorkbookActiveSurfacePresentation.tsx` | Exact built-in or extension renderer selection with lazy extension lifecycle binding. |
| `workbook/components/WorkbookColumnsControl.tsx` | Registered-focus semantic column visibility, ordering, and reset menu. |
| `workbook/components/WorkbookFiltersControl.tsx` | Registered-focus filter draft dialog with exact field/value parsing and invalid-draft feedback. |
| `workbook/components/WorkbookGridEditorControl.tsx` | Contract-field grid editor adapter, mutation controls, commit/cancel behavior, and editor-kind selection. |
| `workbook/components/WorkbookGridControls.tsx` | Active-surface query-control composition over one transient reducer and semantic command port. |
| `workbook/components/WorkbookGroupControl.tsx` | Exact contract-declared grouping selector. |
| `workbook/components/WorkbookIncidentControlsPresentation.tsx` | Incident-controls drawer content and lazy Import Assistant renderer selection. |
| `workbook/components/WorkbookSortControl.tsx` | Complete ordered-sort add, direction, priority, removal, and limit menu. |
| `workbook/inspector/presentation/` | Stateless inspector shell, section, semantic action, history, confirmation, and technical-detail presentation. |
| `workbook/components/WorkbookPresenceMarkers.tsx` | Shared row-gutter and cell presence markers with design-owned capacity and overflow behavior. |
| `workbook/components/WorkbookRelationshipChip.tsx` | Shared relationship-chip presentation over an explicit label, state, detail, selector identity, selection, and command model. |
| `workbook/components/WorkbookViewBar.tsx` | Shared saved-view, query, inspector, and create control composition. |
| `workbook/components/WorkbookShellSlots.tsx` | Stable shell slot IDs, labels, and layout slot helpers. |
| `workbook/components/WorkbookShellTopBar.tsx` | Responsive built-in/system-surface navigation, registered menu focus, incident identity, presence, and account presentation. |
| `workbook/components/WorkbookShellViewBarControls.tsx` | Saved-view and query-control projections over the shell runtime's narrow snapshot and command boundary. |
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

### `workbook/focus/`

Workbook overlay focus is coordinated through registered semantic item refs.
It must not depend on selector discovery, animation frames, or timing delays.

| File | Responsibility |
| --- | --- |
| `workbook/focus/useRegisteredOverlayNavigation.ts` | Registered-ref opening focus, enabled-item traversal, Escape handling, and trigger/fallback restoration for workbook menus and overlays. |
| `workbook/focus/useRegisteredOverlayNavigation.test.tsx` | Focus opening, traversal, disabled-item skipping, close, subject-change, and restoration tests. |

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
| `workbook/hooks/useWorkbookProjectionRefreshController.test.tsx` | Direct tests for initial and sheet-triggered projection refresh ownership. |
| `workbook/hooks/useWorkbookProjectionRefreshController.ts` | Initial session/entity and sheet-triggered projection refresh coordinator. |
| `workbook/hooks/useWorkbookQueryController.test.tsx` | Direct tests for exact-view-schema query-state isolation. |
| `workbook/hooks/useWorkbookQueryController.ts` | Schema-keyed query/filter controller and active query-control adapter. |
| `workbook/hooks/useWorkbookSavedViewController.test.tsx` | Direct tests for saved-view loading and selection precedence. |
| `workbook/hooks/useWorkbookSavedViewController.ts` | Saved-view list, CRUD, persistence, and active-selection controller. |
| `workbook/hooks/useWorkbookSemanticGridFocus.test.tsx` | Direct tests for semantic grid-entry focus order, lifecycle readiness, and stale-request handling. |
| `workbook/hooks/useWorkbookSemanticGridFocus.ts` | Resolves generation-keyed grid-entry requests through the mounted semantic grid handle. |
| `workbook/hooks/useWorkbookAuthorizationState.ts` | Incident authorization subject and explicit session-role recovery lifecycle. |
| `workbook/hooks/useWorkbookCollaborationLifecycle.ts` | Collaboration invalidation, session projection, and exact active-surface registration. |
| `workbook/hooks/useWorkbookExtensionAvailability.test.tsx` | Effect-owned discovery and controller subject-lifetime tests. |
| `workbook/hooks/useWorkbookExtensionAvailability.ts` | Shell-lifetime extension controller, effect-owned discovery, invalidation, and render revision. |
| `workbook/hooks/useWorkbookRecoveryFocus.ts` | Deterministic focus transfer across blocked-edit, overflow, and same-field conflict recovery. |
| `workbook/hooks/useWorkbookShellInfrastructure.ts` | Incident-scoped adapters, registry-owned mutation runtime, command ports, and disposable reference broker. |
| `workbook/hooks/useWorkbookShellRuntime.ts` | Startup, saved-view, query, and layout-state composition facade. |
| `workbook/hooks/useWorkbookSurfaceQueries.ts` | Surface query loading, invalidation, facade projection, and collaboration port selection. |
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
| `workbook/models/workbookGridEntryFocus.ts` | Generation-keyed grid-entry focus request and exact-acknowledgement model. |
| `workbook/models/workbookGridQueryControls.ts` | Pure query-control projection, closure-free command descriptors, exact controlled-value parsers, ordered-sort commands, and surface-keyed transient reducer. |
| `workbook/models/workbookShellPresentation.ts` | Pure account, active-system-surface, and Network Analysis presentation decisions. |
| `workbook/models/workbookGridState.ts` | Contract-grid load-state presentation and incident-role interaction-mode helpers. |
| `workbook/models/workbookIncidentIdentity.ts` | Incident identity normalization and loading-state model. |
| `workbook/models/workbookInspectorModel.ts` | Pure inspector state machine for default-closed state, semantic subjects, active panels, no-row state, and invalidation generations. |
| `workbook/models/workbookQuery.ts` | Workbook query, filter, sort, grouping, and request-building helpers. |
| `workbook/models/workbookReferenceOptions.ts` | Reference option normalization and lookup helpers. |
| `workbook/models/workbookRelationshipChip.ts` | Cross-surface relationship-chip presentation contract with no Timeline interpretation. |
| `workbook/models/workbookRequestDecoders.ts` | Fail-closed exact create, patch, linked-note, and collection-action request decoders with exhaustive generated action coverage. |
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
| `workbook/policies/workbookApplicationShortcuts.ts` | Pure capability-gated Workbook application shortcuts for quick link, Evidence preview, History, and inspector close. |
| `workbook/policies/workbookApplicationShortcuts.test.ts` | Exhaustive application-shortcut admission and event-consumption tests. |
| `workbook/policies/workbookSurfaceOwnershipPolicy.test.ts` | Static policy-purity, owner-completeness, and common-surface boundary checks. |
| `workbook/policies/workbookSurfacePolicy.ts` | Pure `WorkbookSurfacePolicy` types, immutable defaults, and stable reference declarations. |

### `workbook/services/`

Workbook services execute generic mechanics declared by owners. They do not
choose domain references or authorize access.

| File | Responsibility |
| --- | --- |
| `workbook/services/referenceQueryBroker.ts` | Creates instance-scoped, incident/authorization-bound reference-query ports with shared-consumer deduplication, typed invalidation, abort ownership, and idempotent disposal. |
| `workbook/services/referenceQueryBroker.test.ts` | Tests in-flight deduplication, shared-consumer cancellation, two-shell isolation, context binding, invalidation, teardown, and late-result rejection. |

### `workbook/ports/`

Workbook ports expose semantic requests and outcomes only. They contain no
generated DTOs, HTTP envelopes, route construction, status inspection, or API
base coordinates.

| File | Responsibility |
| --- | --- |
| `workbook/ports/WorkbookPortResult.ts` | Shared accepted/aborted/authentication/authorization/stale/retryable/terminal semantic result union. |
| `workbook/ports/WorkbookIncidentPort.ts` | Incident identity and membership capabilities. |
| `workbook/ports/WorkbookPreferencePort.ts` | Current-user and incident-default preference capabilities. |
| `workbook/ports/WorkbookSavedViewPort.ts` | Explicit saved-view page listing and accepted-response CRUD capabilities. |
| `workbook/ports/WorkbookPendingMutationPort.ts` | Shared queued create/patch execution capability over committed versions and semantic mutation units. |

### `workbook/query/`

Workbook query owners bind request admission, live reconciliation, protected
state cleanup, and teardown to the Workbook instance that consumes them.
Shared query rows are application contracts and do not belong to Timeline.

| File | Responsibility |
| --- | --- |
| `workbook/query/WorkbookQueryRow.ts` | Shared schema-keyed view-query row shape below Timeline ownership. |
| `workbook/query/WorkbookViewQueryPort.ts` | Shared abortable query capability returning correlated, contract-normalized rows. |
| `workbook/query/workbookLatestRequest.ts` | Instance-local latest-request sequencing, supersession abort, and current-result admission. |
| `workbook/query/workbookLatestRequest.test.ts` | Tests exclusive latest-request ownership and supersession aborts. |
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
| `workbook/startup/WorkbookStartupPort.ts` | Semantic startup query inputs and correlated startup selection/availability outcomes. |
| `workbook/startup/useWorkbookStartupAdmission.ts` | Startup request and response admission boundary with selection-version, incident, ordinal, availability, and cancellation guards. |
| `workbook/startup/useWorkbookStartupAdmission.test.tsx` | Startup admission characterization for precedence, overlap, extension availability, fallbacks, teardown, and late work. |

### `workbook/timeline/`

Timeline-specific implementation lives under `workbook/timeline/`. This folder
owns Timeline composition, presentation, controllers, models, and semantic
ports/adapters. Generic workbook behavior belongs one level up unless it is
truly Timeline-specific. `TimelineWorkbook.tsx` is the public facade and sole
composition-hook caller; the private root composer assembles focused owners and
publishes a narrow presentation projection. Feature-specific coordination
belongs in the composition, hook, model, port, or adapter that owns its
lifecycle, while render-only structure belongs under `presentation/` or
`components/`.

The intended dependency direction is:

`TimelineWorkbook` → root composition → presentation model → stateless view

#### Timeline root-level tests

Cross-folder tests remain at the Timeline root when they characterize a
composition boundary or a controller interaction spanning more than one
implementation folder.

| File | Responsibility |
| --- | --- |
| `workbook/timeline/timelineCompositionArchitecture.test.ts` | Enforces the slim public root, sole composition-hook caller, stateless presentation regions, forbidden-capability exclusion, and visible-column synchronization ownership. |
| `workbook/timeline/useTimelineCommittedRecordIdle.test.tsx` | Characterizes bounded refresh when a committed record version is temporarily missing. |
| `workbook/timeline/useTimelineCompositionLifecycle.test.tsx` | Characterizes grid measurement/observer cleanup and inspector selection/continuity reset ownership. |
| `workbook/timeline/useTimelineInspectorFeatureController.test.tsx` | Characterizes canonical feature routing, fail-closed behavior, cancellation, exclusivity, and lifecycle reset. |
| `workbook/timeline/useTimelineInspectorLifecycle.test.tsx` | Characterizes the shared explicit/layout inspector close command. |
| `workbook/timeline/useTimelineKeyboardController.test.tsx` | Characterizes scalar/collection commit, navigation, range, draft, inspector, work-area, and event-consumption keyboard ownership. |
| `workbook/timeline/useTimelineMentionActions.test.tsx` | Characterizes auto-resolution undo identity, committed-version refresh/continuity sequencing, and rejection behavior. |
| `workbook/timeline/useTimelineMutationRuntimeBindings.test.tsx` | Characterizes concrete mutation-runtime command registration, replacement, and cleanup. |
| `workbook/timeline/useTimelineRows.test.tsx` | Characterizes initial draft-row identity, stable row refs, and monotonic draft allocation. |
| `workbook/timeline/useTimelineSurfaceFoundation.test.tsx` | Characterizes stable adapter, row/query, pending-save, and semantic foundation identities. |

#### `workbook/timeline/composition/`

Timeline composition assembles existing semantic owners without importing
presentation, generated protocol bindings, or browser services. Only the root
composer may import sibling composition owners; lower Timeline layers do not
depend back on composition.

| File | Responsibility |
| --- | --- |
| `workbook/timeline/composition/useTimelineGridEnvironment.ts` | Owns Timeline grid refs, rounded width observation and cleanup, focus and anchor commands, viewport continuity, and row-mutation editor adaptation. |
| `workbook/timeline/composition/useTimelineInspectorStateComposition.ts` | Owns Timeline selection, inspector open/reset continuity, invalidation state, feedback, and row-history state. |
| `workbook/timeline/composition/useTimelineInspectorWorkflowComposition.ts` | Owns Timeline feature routing, related-record, history, mention, Evidence, row-menu, close, and Escape workflows. |
| `workbook/timeline/composition/useTimelineInteractionComposition.ts` | Owns Timeline bulk, keyboard, paste, fill, scalar/collection commit, draft, collection-focus, and recovery-focus interactions. |
| `workbook/timeline/composition/useTimelineMutationComposition.ts` | Owns the explicit Timeline query admission, row mutation, replay, runtime registration, collaboration, presence, conflict, and save command graph. |
| `workbook/timeline/composition/useTimelineSurfaceFoundation.ts` | Owns semantic Timeline adapter construction, query/lifecycle foundations, row and mention state, pending-save refs, editor drafts, and guarded workbook timing. |
| `workbook/timeline/composition/useTimelineWorkbookComposition.ts` | Private root composer that invokes Timeline composition owners in dependency order, returns grouped capabilities, and publishes the capability-limited presentation projection. |

#### `workbook/timeline/presentation/`

Timeline presentation derives render-only models from the composition-owned
presentation projection plus narrow entity, layout, Indicator, and query-control
inputs. It receives no mutation runtime, pending runtime, collaboration
coordinator, authorization recovery callback, generated protocol, or browser
service. The presentation hook owns the sole visible-column synchronization
effect; the view and focused regions are stateless.

| File | Responsibility |
| --- | --- |
| `workbook/timeline/presentation/TimelineWorkbookInspectorRegion.tsx` | Stateless Timeline inspector and Indicator supplement region. |
| `workbook/timeline/presentation/TimelineWorkbookOverlayRegion.tsx` | Stateless notices and row-context-menu overlay region. |
| `workbook/timeline/presentation/TimelineWorkbookView.tsx` | Stateless Timeline surface layout and region assembly. |
| `workbook/timeline/presentation/TimelineWorkbookViewBarRegion.tsx` | Stateless query, saved-view, add-row, inspector-toggle, and bulk-action controls. |
| `workbook/timeline/presentation/useTimelineWorkbookPresentation.tsx` | Derives renderer, column, grid-row, load-state, inspector, status, view-bar, notice, and context-menu models. |

#### `workbook/timeline/adapters/`

Timeline adapters are the private protocol boundary. They derive transport from
generated operation bindings, validate success contracts, and expose only
owner-specific semantic outcomes to Timeline controllers.

| File | Responsibility |
| --- | --- |
| `workbook/timeline/adapters/createTimelineActionAdapters.test.ts` | Characterizes history, record action, mention, and Evidence attachment protocol boundaries. |
| `workbook/timeline/adapters/createTimelineEvidenceAttachmentAdapter.ts` | Creates an uploaded Evidence object and row, then links it to Timeline with stable transaction identity. |
| `workbook/timeline/adapters/createTimelineHistoryAdapter.ts` | Loads validated record history and executes delete, restore, and rollback operations. |
| `workbook/timeline/adapters/createTimelineMentionAdapter.ts` | Creates mention target entities and resolves mention actions through generated operations. |
| `workbook/timeline/adapters/createTimelineRecordActionAdapter.ts` | Executes and normalizes Timeline review and supersede actions. |
| `workbook/timeline/adapters/createTimelineRowMutationEditorAdapter.ts` | Translates semantic row-mutation editor commands into grid and continuity operations. |
| `workbook/timeline/adapters/createTimelineScalarGridCommitAdapter.ts` | Adapts Grid Adapter scalar commits to the Timeline scalar-save command and exact settlement promise. |
| `workbook/timeline/adapters/createTimelineSocketTransactionAdapter.ts` | Adapts Timeline accepted/action transaction tracking to the shell-lifetime runtime ledger. |
| `workbook/timeline/adapters/timelineProjectionCommitAdapter.ts` | Sole synchronous projection-commit boundary used when focus or continuity must observe the committed Timeline row tree. |

#### `workbook/timeline/bulk/`

Timeline bulk controllers own stable record-identity selection and one adopted
multi-record workflow without depending on transport or DOM state.

| File | Responsibility |
| --- | --- |
| `workbook/timeline/bulk/useTimelineBulkTagController.test.tsx` | Characterizes current-page stable-ID selection, pruning, versioned submission, conflicts, authorization loss, and draft retention. |
| `workbook/timeline/bulk/useTimelineBulkTagController.ts` | Owns Timeline tag-selection state, eligibility/pruning, tag draft, deduplicated semantic submission, refresh, bounded conflict copy, and late-completion invalidation. |
| `workbook/timeline/bulk/useTimelineFillController.test.tsx` | Characterizes ordered fill planning, atomic rejection, semantic dispatch, save/refresh sequencing, focus restoration, and conflict activation. |
| `workbook/timeline/bulk/useTimelineFillController.ts` | Plans and dispatches versioned Timeline fill commands against stable record/field identities while preserving mutation admission and focus continuity. |

#### `workbook/timeline/collaboration/`

Timeline collaboration bindings adapt the shell coordinator's active-surface
port without owning WebSocket transport, sequencing, or authorization policy.

| File | Responsibility |
| --- | --- |
| `workbook/timeline/collaboration/TimelineCollaborationBoundary.tsx` | Owns the Timeline collaboration session boundary and optional coordinator-session attachment around the presentation root. |
| `workbook/timeline/collaboration/useTimelineCollaborationBindings.test.tsx` | Characterizes live/stale admission, fail-closed sparse patches, gap refresh, access invalidation, surface switching, and teardown. |
| `workbook/timeline/collaboration/useTimelineCollaborationBindings.ts` | Owns Timeline active-surface and transaction-resolver registration, sparse row patching, refresh/reset effects, presence publication, and lifecycle teardown. |
| `workbook/timeline/collaboration/useTimelinePresenceController.test.tsx` | Characterizes stable record/field presence derivation, coherent viewing/editing publication, and lifecycle reset without stale publication. |
| `workbook/timeline/collaboration/useTimelinePresenceController.ts` | Derives Timeline row/cell presence and publishes semantic viewing/editing transitions for the active sheet. |

#### `workbook/timeline/components/`

| File | Responsibility |
| --- | --- |
| `workbook/timeline/components/TimelineCollectionCell.tsx` | Focused relationship/tag summary, overflow, and collection-draft cell presentation over discriminated models. |
| `workbook/timeline/components/TimelineDraftRowActions.tsx` | Timeline draft-row create and evidence-attachment actions. |
| `workbook/timeline/components/TimelineEvidencePanel.test.tsx` | Tests for Timeline evidence panel behavior. |
| `workbook/timeline/components/TimelineEvidencePanel.tsx` | Timeline inspector evidence panel and evidence actions UI. |
| `workbook/timeline/components/TimelineHistoryPanel.tsx` | Timeline row history, rollback, delete, restore, and history action presentation. |
| `workbook/timeline/components/TimelineMentionsPanel.tsx` | Timeline mention-resolution inspector panel. |
| `workbook/timeline/components/TimelineRowActions.tsx` | Timeline row action/context-menu presentation. |
| `workbook/timeline/components/TimelineScalarEditor.test.tsx` | Characterizes controlled scalar drafts, read-only behavior, presence publication, and commit lifecycle. |
| `workbook/timeline/components/TimelineScalarEditor.tsx` | Timeline scalar input/textarea editing, commit, presence, clipboard, and grid-editor lifecycle behavior. |
| `workbook/timeline/components/TimelineWorkbook.tsx` | Public Timeline facade that retains the collaboration boundary and delegates grouped composition, presentation derivation, and stateless rendering. |
| `workbook/timeline/components/TimelineWorkbookGrid.tsx` | Timeline grid renderer, grouped row table wrapper, hidden contract metadata cells, and grid test-ID placement. |
| `workbook/timeline/components/TimelineWorkbookInspector.tsx` | Timeline inspector shell, panel tabs, disabled-state presentation, selected-row state, and inspector messages. |
| `workbook/timeline/components/TimelineWorkbookInspectorSections.tsx` | Timeline inspector section factories for field editors, relationships, evidence attach, related-row creation, and row history. |
| `workbook/timeline/components/TimelineWorkbookNotices.tsx` | Timeline notice overlay for auto-resolution notices, pending queue messages, and queued-edit counts. |
| `workbook/timeline/components/TimelineWorkbookRendererTypes.ts` | Private renderer command and output types shared by the Timeline renderer workstreams. |
| `workbook/timeline/components/TimelineWorkbookRenderers.tsx` | Stable private facade composing scalar, collection, and column renderer owners. |
| `workbook/timeline/components/TimelineWorkbookStyles.ts` | Timeline-specific style constants shared by Timeline workbook components. |
| `workbook/timeline/components/useTimelineCollectionRenderer.tsx` | Narrow renderer factory that binds Timeline collection commands to the focused collection-cell component. |
| `workbook/timeline/components/useTimelineColumnAssembly.tsx` | Timeline contract column ordering, widths, editors, clipboard values, and evidence/read-only cell assembly. |
| `workbook/timeline/components/useTimelineScalarRenderers.tsx` | Timeline scalar grid/inspector controls, read cells, presence, and conflict markers. |

#### `workbook/timeline/editing/`

Timeline editor state is keyed by semantic record, field, and surface identity so
authoritative refreshes and grid implementation details cannot erase or address it
by visual coordinates.

| File | Purpose |
| --- | --- |
| `workbook/timeline/editing/useTimelineEditorDraftRegistry.test.tsx` | Characterizes invalid-text preservation, grid/inspector separation, submitted-value cleanup, semantic input registration, row removal, and schema invalidation. |
| `workbook/timeline/editing/useTimelineEditorDraftRegistry.ts` | Owns scalar editor drafts and input references by semantic row/field/surface identity for one Timeline schema generation. |

#### `workbook/timeline/hooks/`

| File | Responsibility |
| --- | --- |
| `workbook/timeline/hooks/useTimelineClipboardPasteController.ts` | Coordinates semantic Timeline paste targets/outcomes, conflict registration, scalar fallback, and post-paste focus/viewport restoration. |
| `workbook/timeline/hooks/useTimelineCommittedRecordIdle.ts` | Waits for one committed Timeline record to become mutation-idle and performs at most one authoritative refresh when its committed version is missing. |
| `workbook/timeline/hooks/useTimelineCommittedRows.ts` | Derives committed Timeline row collections from row/runtime state. |
| `workbook/timeline/hooks/useTimelineConflicts.ts` | Coordinates Timeline same-field conflict state. |
| `workbook/timeline/hooks/useTimelineCreateRelatedWorkflow.ts` | Coordinates Timeline inspector related-row workflow state, draft values, semantic creation, and Evidence linking. |
| `workbook/timeline/hooks/useTimelineEvidenceAttach.ts` | Coordinates semantic Timeline Evidence attachment, validation feedback, and save/continuity handoff. |
| `workbook/timeline/hooks/useTimelineGridAnchorController.ts` | Resolves Timeline grid anchors, paste targets, selected cells, and focus anchors across committed and draft rows. |
| `workbook/timeline/hooks/useTimelineGridInteractions.ts` | Coordinates Timeline grid refs, keyboard helpers, and grid interaction commands. |
| `workbook/timeline/hooks/useTimelineHistoryActions.ts` | Coordinates semantic Timeline history load, rollback, delete, restore, preview, and confirmation actions. |
| `workbook/timeline/hooks/useTimelineHistoryState.ts` | Coordinates Timeline history panel and row-history state. |
| `workbook/timeline/hooks/useTimelineInspectorFeatureController.ts` | Owns fail-closed Timeline inspector feature routing, mutually exclusive workflow activation, cancellation, and lifecycle invalidation. |
| `workbook/timeline/hooks/useTimelineInspectorSelection.ts` | Coordinates selected Timeline row, inspector feedback and closure, row-bound feature invalidation, deleted-row history, and focus-safe inspector interactions. |
| `workbook/timeline/hooks/useTimelineKeyboardController.ts` | Owns Timeline scalar/collection editor keys, grid navigation/range commands, work-area shortcuts, event consumption, and focus priority. |
| `workbook/timeline/hooks/useTimelineMentionActions.ts` | Coordinates semantic Timeline mention resolution, entity creation, undo/review actions, and inspector updates. |
| `workbook/timeline/hooks/useTimelineMentions.ts` | Coordinates Timeline mention-resolution state and actions. |
| `workbook/timeline/hooks/useTimelineMutationCommands.ts` | Applies pure scalar/collection admission plans, publishes optimistic rows, and queues exact Timeline owner envelopes. |
| `workbook/timeline/hooks/useTimelineMutationRuntimeBindings.ts` | Lifecycle-registers concrete Timeline refresh, conflict-application, focus-restoration, and blocked-edit-discard commands with the workbook mutation runtime. |
| `workbook/timeline/hooks/useTimelineMutationDriver.ts` | Registers the exact Timeline row driver and applies owner-local admission, revalidation, settlement, conflict, discard, and accepted-result plans. |
| `workbook/timeline/hooks/useTimelinePendingSaves.ts` | Coordinates Timeline pending-save queue runtime and replay admission. |
| `workbook/timeline/hooks/useTimelineRows.ts` | Owns Timeline row state, the stable row ref, initial draft row, monotonic draft allocation, and semantic replace/update commands. |
| `workbook/timeline/hooks/useTimelineRowsLoader.ts` | Executes the pure load machine around exact query, freshness, local-draft hydration, created-row pinning, access-loss, and continuity boundaries. |
| `workbook/timeline/hooks/useTimelineSaveStatePresentation.ts` | Coordinates Timeline save-state labels, pending queue snapshot publication, refresh blocking, runtime drain requests, and beforeunload warning state. |
| `workbook/timeline/hooks/useTimelineConflictProjectionAdapter.ts` | Timeline render-state adapter for shell-owned same-field conflict registration, projection, and resolution. |
| `workbook/timeline/hooks/useTimelineViewportContinuityController.ts` | Coordinates Timeline scroll snapshots, focus restoration, continuity tokens, and entity-refresh barriers. |
| `workbook/timeline/hooks/useTimelineWorkbookRuntime.ts` | Reduces Timeline lifecycle state and translates shell-owned query commands into deterministic runtime transitions. |

#### `workbook/timeline/models/`

| File | Responsibility |
| --- | --- |
| `workbook/timeline/models/timelineAcceptedMutationEffects.ts` | Pure post-acceptance selection, notice, created-row, and continuity effect planning. |
| `workbook/timeline/models/timelineAcceptedProjection.ts` | Pure accepted-row replacement, insertion, and bottom-draft projection. |
| `workbook/timeline/models/timelineClipboardPastePlan.ts` | Pure Timeline paste authority, shape, field, surface, and stable-target admission policy. |
| `workbook/timeline/models/timelineCollectionPresentation.ts` | Discriminated relationship/tag collection items, overflow identity, and accessible hidden labels. |
| `workbook/timeline/models/timelineCommittedVersionLedger.ts` | Monotonic committed row/version high-water ledger with reference-preserving acceptance. |
| `workbook/timeline/models/timelineConflictState.ts` | Timeline-local same-field and grouped-paste conflict state types. |
| `workbook/timeline/models/timelineControllerPorts.ts` | Neutral capability-port, row-store, committed-record-idle, context-menu-position, and replay contracts shared by isolated Timeline controllers. |
| `workbook/timeline/models/timelineDiscardedReconciliation.ts` | Pure discarded-unit reconciliation that reapplies later same-row work in FIFO order. |
| `workbook/timeline/models/timelineFieldRegistry.ts` | Exhaustive Timeline scalar, collection, readonly, inspector, and focus binding registry. |
| `workbook/timeline/models/timelineHistoryModel.ts` | Timeline row-history normalization, pending-action labels, and history operation helpers. |
| `workbook/timeline/models/timelineLayoutPolicy.ts` | Pure Timeline grouping labels and base/expanded column-width policy. |
| `workbook/timeline/models/timelineLoadMachine.test.ts` | Exhaustive load-subject, lifecycle, mutation-race, retry-bound, failure, access-loss, and obligation-join transition evidence. |
| `workbook/timeline/models/timelineLoadMachine.ts` | Pure incident/surface/query/generation/mutation-epoch/source-obligation transitions with explicit load effects. |
| `workbook/timeline/models/timelineMutationDriverPlans.ts` | Pure Timeline replay admission, settlement, discard, and accepted-projection decisions. |
| `workbook/timeline/models/timelineMutationIntents.ts` | Exact scalar, collection-action, and draft-create mutation intent construction. |
| `workbook/timeline/models/timelineMutationModels.test.ts` | Pure intent, deduplication, acceptance, discard, version-ledger, and discriminated-collection evidence. |
| `workbook/timeline/models/timelineMutationQueueAdmission.ts` | Pure scalar/collection no-op, conflict, duplicate, and exact queue-admission decisions. |
| `workbook/timeline/models/timelinePendingSaves.ts` | Timeline-local pending signature, replay-order, and serial-save references. |
| `workbook/timeline/models/timelineRowModel.ts` | Timeline row envelope decoding, normalization, materialization, and sparse-patch application. |
| `workbook/timeline/models/timelineRowsModel.ts` | Timeline row collection helpers and row-state utilities. |
| `workbook/timeline/models/timelineWorkbookFeaturePolicy.test.ts` | Characterizes canonical related-row/Indicator feature tuples and fail-closed rejection of altered or unsupported tuples. |
| `workbook/timeline/models/timelineWorkbookFeaturePolicy.ts` | Defines canonical Timeline inspector feature tuples and fail-closed semantic routing. |
| `workbook/timeline/models/timelineWorkbookSurfaceRuntime.ts` | Required shell-owned Timeline composition contract for incident, query, entity, layout, and access-loss services. |
| `workbook/timeline/models/timelineViewportContinuityModel.ts` | Timeline viewport continuity and entity-refresh barrier helpers. |
| `workbook/timeline/models/workbookMentionChips.ts` | Mention chip state, relationship-field keys, and mention display helpers. |
| `workbook/timeline/models/workbookRecordFreshness.ts` | Pure durable-identity and row-version freshness comparison leaf. |
| `workbook/timeline/models/timelineRowsModel.test.ts` | Tests for Timeline grid-row materialization. |
| `workbook/timeline/models/timelineWorkbookRuntime.test.ts` | Deterministic lifecycle transition traces for load, refresh, save, conflict, and recovery state. |
| `workbook/timeline/models/timelineViewportContinuityModel.test.ts` | Tests for viewport continuity and refresh barrier helpers. |
| `workbook/timeline/models/workbookRecordFreshness.test.ts` | Tests for comparable and non-comparable row-version freshness decisions. |
| `workbook/timeline/models/timelineModelBoundaries.test.ts` | Tests for Timeline row, payload, binding, normalization, and display helpers. |

#### `workbook/timeline/mutations/`

Timeline row state admits query, mutation, replay, conflict, and live-event
results through one high-water coordinator before presentation observes them.

| File | Purpose |
| --- | --- |
| `workbook/timeline/mutations/useTimelineRowMutationCoordinator.test.tsx` | Characterizes accepted/stale action and mutation admission, query/action races, conflict state partitions, and committed-version high-water behavior. |
| `workbook/timeline/mutations/useTimelineRowMutationCoordinator.ts` | Applies accepted/discarded plans and composes committed-version, conflict, transaction, save-state, collaboration, and continuity owners. |

#### `workbook/timeline/ports/`

Each Timeline owner receives only its required capability. There is no broad
Timeline action facade and no port exposes routes, status codes, raw payloads,
or transport coordinates.

| File | Responsibility |
| --- | --- |
| `workbook/timeline/ports/TimelineEvidenceAttachmentPort.ts` | Semantic file attachment and Timeline row outcomes. |
| `workbook/timeline/ports/TimelineHistoryPort.ts` | Semantic history query, delete/restore, and rollback capabilities. |
| `workbook/timeline/ports/TimelineMentionPort.ts` | Semantic mention entity creation and resolution capabilities. |
| `workbook/timeline/ports/TimelineRecordActionPort.ts` | Semantic Timeline review and supersede capability. |

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
| `workbook/runtime/WorkbookMutationRuntime.ts` | Shell-lifetime facade over queue coordination, conflicts, surface registration, managed patches, retry, transaction settlement, and status projection. |
| `workbook/runtime/WorkbookClientTransactionLedger.ts` | Bounded client-transaction identity retention and pending-queue settlement lookup. |
| `workbook/runtime/WorkbookConflictStore.ts` | Conflict registration, compatible draft preservation, refresh commands, and panel state. |
| `workbook/runtime/WorkbookManagedPatchDriver.ts` | Managed-patch admission, transport dispatch, settlement, local projection, conflict registration, and refresh. |
| `workbook/runtime/WorkbookMutationDriverRegistry.ts` | Closed managed-patch/Timeline-row owner envelopes, exact driver registration, duplicate rejection, and absence-safe dispatch. |
| `workbook/runtime/WorkbookRetryScheduler.ts` | Injected, single-flight retry scheduling and cancellation. |
| `workbook/runtime/WorkbookRuntimeLifecycle.ts` | Listener lifetime, coalesced drain notification, and terminal disposal. |
| `workbook/runtime/WorkbookSurfaceRegistry.ts` | Surface command registration, replacement-safe cleanup, and retained refresh debt. |
| `workbook/runtime/workbookMutationStatusProjector.ts` | Pure queue/conflict/explicit-operation status and save-state projection. |
| `workbook/runtime/workbookRuntimePorts.ts` | Clock and scheduler ports with the browser composition defaults. |
| `workbook/runtime/workbookPendingMutationSettlement.ts` | Maps semantic mutation failures to common queue settlement outcomes. |
| `workbook/runtime/workbookPendingReplayRuntime.ts` | Pending-replay runtime state, admission contracts, and refresh barriers. |
| `workbook/runtime/workbookConflictModel.ts` | Same-field conflict parsing and common envelope types. |
| `workbook/runtime/workbookLifecycleModel.ts` | Shared load, refresh, save, conflict, and recovery lifecycle reducer. |
| `workbook/runtime/useWorkbookMutationRuntime.ts` | React external-store subscription hook for shell-owned workbook mutation state. |
| `workbook/runtime/WorkbookMutationRuntime.test.ts` | Tests for shell-lifetime queue retention, autosave, refresh debt, conflicts, and mutation coordination. |
| `workbook/runtime/WorkbookRuntimeResponsibilities.test.ts` | Deterministic tests for responsibility boundaries, injected time/scheduling, registration cleanup, conflict drafts, transaction settlement, and disposal. |
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
| `workbook/utils/workbookPendingQueue.ts` | Pending-save queue capacity, save-state, replay, conflict, and public-error helpers. |
| `workbook/utils/workbookPresence.ts` | Presence input/type helpers and presence matching helpers. |
| `workbook/utils/workbookRowReconciliation.ts` | Record-identity and row-version reconciliation that preserves unchanged row references. |
| `workbook/utils/workbookStyles.ts` | Shared workbook style primitives. |
| `workbook/utils/workbookValueFormat.ts` | Grid/workbook value formatting helpers. |
| `workbook/utils/GridAdapter.anchor.test.ts` | Workbook interaction grid-adapter anchor behavior tests. |
| `workbook/utils/workbookClipboard.test.ts` | Tests for clipboard helpers. |
| `workbook/utils/workbookPendingQueue.test.ts` | Tests for pending-queue helpers. |
| `workbook/utils/workbookRowReconciliation.test.ts` | Tests for sparse row replacement, removal, drafts, and row-version reference reuse. |
| `workbook/utils/workbookValueFormat.test.ts` | Tests for value formatting helpers. |

### `workbook/view-state/`

Workbook view-state hooks own instance-scoped query and layout state. They
contain no module-global mutable defaults or cross-shell store.

| File | Responsibility |
| --- | --- |
| `workbook/view-state/useWorkbookQueryState.ts` | Instance-owned schema-keyed query defaults, reducer state, reset, and update operations. |
