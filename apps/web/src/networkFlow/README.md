# networkFlow/

[Source overview / parent](../README.md)

Network Analysis extension presentation, analytical tables, imports, graph exploration, and retained operations.

Network Flow is an extension feature, not a Base Profile workbook surface or
a `view_schema` owner. Workbook integration uses the
[feature entry points](../workbook/features/README.md): lazy
[NetworkFlowFeature.tsx](../workbook/features/NetworkFlowFeature.tsx) presentation
and [NetworkFlowOperations.ts](../workbook/features/NetworkFlowOperations.ts)
for persistent operation ownership.

Extension availability comes from [extensions](../extensions/README.md);
the shared [collaboration session](../collaboration/README.md) owns the socket.
Measurement fixtures live in [measurement](../measurement/README.md).

## Graph exploration

| File | Responsibility |
| --- | --- |
| [explorationContinuity.test.tsx](explorationContinuity.test.tsx) | Tests exploration selection retention, hidden-read pausing, and stale contributor rejection. |
| [explorationFocus.test.tsx](explorationFocus.test.tsx) | Tests safe focus withdrawal on graph removal and superseded semantic focus intents. |
| [explorationTestFixtures.ts](explorationTestFixtures.ts) | Deterministic graph exploration fixture for continuity and focus tests. |
| [networkFlowExplorationNavigation.test.ts](networkFlowExplorationNavigation.test.ts) | Tests bucket transitions, off-page selection, and semantic reveal/close focus planning. |
| [networkFlowExplorationNavigation.ts](networkFlowExplorationNavigation.ts) | Pure graph exploration indexing, vertex/edge pagination, selection, and focus plans. |
| [NetworkFlowExplorationPanel.tsx](NetworkFlowExplorationPanel.tsx) | Graph exploration presentation for vertices, edges, semantic selection, and contributor pages. |

## Indicator linking

| File | Responsibility |
| --- | --- |
| [IndicatorLinkDialog.tsx](IndicatorLinkDialog.tsx) | Indicator-link review dialog and retained-operation recovery presentation. |
| [IndicatorLinkTargetDiscovery.ts](IndicatorLinkTargetDiscovery.ts) | Paged compatible Indicator discovery with authority-bound snapshots and stale-read fencing. |
| [NetworkFlowIndicatorLinkController.ts](NetworkFlowIndicatorLinkController.ts) | Indicator-link draft, reviewed write, settlement, and recovery ownership. |
| [networkFlowIndicatorLinkModel.test.ts](networkFlowIndicatorLinkModel.test.ts) | Tests for semantic IP-cell/range selection, row-ref construction, and link-admission limits. |
| [networkFlowIndicatorLinkModel.ts](networkFlowIndicatorLinkModel.ts) | Semantic grid-selection validation and immutable row-selector construction for indicator links. |
| [networkFlowIndicatorLinkOperation.test.ts](networkFlowIndicatorLinkOperation.test.ts) | Tests Indicator target discovery, stale-read fencing, and captured link-operation recovery. |
| [networkFlowIndicatorLinkOperation.ts](networkFlowIndicatorLinkOperation.ts) | Indicator-link authority, immutable candidates and attempts, admission, and failure models. |
| [useNetworkFlowIndicatorLinkController.ts](useNetworkFlowIndicatorLinkController.ts) | React subscription and selection-context binding for the retained Indicator-link controller. |
| [useNetworkFlowIndicatorLinkOwner.ts](useNetworkFlowIndicatorLinkOwner.ts) | React lifetime binding for Indicator-link operation ownership and target discovery. |

## Workspace and collaboration

| File | Responsibility |
| --- | --- |
| [NetworkAnalysisWorkspace.test.tsx](NetworkAnalysisWorkspace.test.tsx) | Network Analysis workspace behavior tests. |
| [NetworkAnalysisWorkspace.tsx](NetworkAnalysisWorkspace.tsx) | Network Analysis presentation/composition facade over feature-specific controllers. |
| [networkFlowBoundaryPolicy.test.ts](networkFlowBoundaryPolicy.test.ts) | Static enforcement for controller composition, generated-decoder ownership, and browser projection-input exclusion. |
| [networkFlowClient.test.ts](networkFlowClient.test.ts) | Tests for extension route admission, cancellation, CSRF, and generated response decoding at the client boundary. |
| [networkFlowClient.ts](networkFlowClient.ts) | Decoded Network Flow feature operations over the profile-and-route-scoped owner-neutral browser transport. |
| [networkFlowCollaborationInterpreter.test.ts](networkFlowCollaborationInterpreter.test.ts) | Network Flow collaboration event admission tests. |
| [networkFlowCollaborationInterpreter.ts](networkFlowCollaborationInterpreter.ts) | Feature-local interpreter for decoded Network Flow extension invalidation/removal events. |
| [networkFlowController.test.ts](networkFlowController.test.ts) | Network Flow controller lifecycle tests. |
| [networkFlowController.ts](networkFlowController.ts) | Pure Network Flow table selection and refresh/removal state reducer. |
| [networkFlowErrors.test.ts](networkFlowErrors.test.ts) | Tests for structured Network Flow error preservation, safe messages, and lifecycle/access-loss classification. |
| [networkFlowErrors.ts](networkFlowErrors.ts) | Feature-local authorization-loss classification shared by query controllers. |
| [networkFlowWorkspaceStatus.ts](networkFlowWorkspaceStatus.ts) | Derives Network Analysis workspace status from operation and query state. |
| [useNetworkFlowCollaborationController.ts](useNetworkFlowCollaborationController.ts) | Owner invalidation/removal effects over the shared collaboration session. |
| [useNetworkFlowExtensionEvents.test.tsx](useNetworkFlowExtensionEvents.test.tsx) | Shared-session Network Flow reconnect and sequence-deduplication tests. |
| [useNetworkFlowExtensionEvents.ts](useNetworkFlowExtensionEvents.ts) | Network Flow subscription adapter over the shared incident collaboration session. |

## Grid and control presentation

| File | Responsibility |
| --- | --- |
| [NetworkFlowControls.test.tsx](NetworkFlowControls.test.tsx) | Tests native control behavior, visual variants, selected semantics, and accessible icon names. |
| [NetworkFlowControls.tsx](NetworkFlowControls.tsx) | Shared Network Flow buttons, inputs, and chrome presentation primitives. |
| [networkFlowPresentation.test.tsx](networkFlowPresentation.test.tsx) | Tests for extension grid metadata, row identities, formatting, diagnostic localization, grouping, and projection reuse. |
| [networkFlowPresentation.tsx](networkFlowPresentation.tsx) | Extension-owned grid metadata, semantic row projections, value formatting, diagnostic localization, and column compilation. |
| [NetworkFlowSemanticGrid.test.tsx](NetworkFlowSemanticGrid.test.tsx) | Semantic-grid accessibility tests for focus recovery and keyboard column reordering announcements. |
| [NetworkFlowSemanticGrid.tsx](NetworkFlowSemanticGrid.tsx) | Semantic accepted-row, rejected-row, and contributor grids with layout controls, selection, focus recovery, and inspector presentation. |
| [useNetworkFlowGridLayout.test.tsx](useNetworkFlowGridLayout.test.tsx) | Tests for Network Flow session layout mutation, reset, and remount persistence. |
| [useNetworkFlowGridLayout.ts](useNetworkFlowGridLayout.ts) | Session-lifetime Network Flow column visibility, order, width, and reset state. |
| [useNetworkFlowModalFocus.ts](useNetworkFlowModalFocus.ts) | Saved-graph modal focus, Escape and return, coordinated with shell recovery. |

## Import and mapping

| File | Responsibility |
| --- | --- |
| [networkFlowImportAdmission.test.ts](networkFlowImportAdmission.test.ts) | Tests analytical claim revalidation after shared queue admission and before upload dispatch. |
| [NetworkFlowImportController.test.ts](NetworkFlowImportController.test.ts) | Tests single upload admission, preview/approval separation, and stale preview rejection. |
| [NetworkFlowImportController.ts](NetworkFlowImportController.ts) | Network Flow import workflow owner for captured attempts, preview, approval, apply, and recovery. |
| [networkFlowImportModel.test.ts](networkFlowImportModel.test.ts) | Ordinal identity, registry suggestion, policy-accounting, and timestamp candidate tests. |
| [networkFlowImportModel.ts](networkFlowImportModel.ts) | Generated-registry mapping suggestions and candidate construction for discovered source ordinals. |
| [networkFlowImportState.ts](networkFlowImportState.ts) | Network Flow import stages, captured attempts, jobs, previews, approvals, and handoff state. |
| [NetworkFlowImportSurface.tsx](NetworkFlowImportSurface.tsx) | Network Flow import stage controls and retained-operation recovery presentation. |
| [NetworkFlowMappingPanel.tsx](NetworkFlowMappingPanel.tsx) | Owner-controlled ordinal mapping, preview and approval inside the shell recovery panel. |
| [useNetworkFlowImportController.ts](useNetworkFlowImportController.ts) | React presentation binding for the persistent Network Flow import workflow owner. |

## Queries and paging

| File | Responsibility |
| --- | --- |
| [networkFlowPageRecovery.test.tsx](networkFlowPageRecovery.test.tsx) | Tests contributor replacement fencing and consistent accepted/rejected-row page recovery. |
| [networkFlowQueryAuthoring.test.ts](networkFlowQueryAuthoring.test.ts) | Tests exact query predicate round trips, unsigned integer bounds, and scalar validation. |
| [NetworkFlowQueryControls.test.tsx](NetworkFlowQueryControls.test.tsx) | Tests unchanged Apply preserves protocol lists and complete counter predicates. |
| [NetworkFlowQueryControls.tsx](NetworkFlowQueryControls.tsx) | Accepted-row and rejected-row filter, sort, time-window, and reset controls. |
| [networkFlowQueryModel.test.ts](networkFlowQueryModel.test.ts) | Tests for exact initial and continuation queries plus owner-identity reconciliation. |
| [networkFlowQueryModel.ts](networkFlowQueryModel.ts) | Accepted/rejected query request compilation, continuation construction, and owner-identity result reconciliation. |
| [NetworkFlowQueryPagination.test.tsx](NetworkFlowQueryPagination.test.tsx) | Tests committed-page feedback and distinct destination, refresh, and recovery actions. |
| [NetworkFlowQueryPagination.tsx](NetworkFlowQueryPagination.tsx) | Paging controls and feedback that distinguish committed results from pending or failed destinations. |
| [networkFlowQueryValues.ts](networkFlowQueryValues.ts) | Strict query scalar, timestamp, and IP parsing with exact scalar comparison. |
| [useNetworkFlowGraphController.ts](useNetworkFlowGraphController.ts) | Graph and contributor query state, cancellation, selection, and stale-result rejection. |
| [useNetworkFlowPagedQuery.test.tsx](useNetworkFlowPagedQuery.test.tsx) | Tests for cursor history, invalid-cursor recovery, request cancellation, and late-response rejection. |
| [useNetworkFlowPagedQuery.ts](useNetworkFlowPagedQuery.ts) | Abortable cursor paging, previous/next history, invalid-cursor recovery, and protected-state clearing hook. |
| [useNetworkFlowQueryAuthoring.test.ts](useNetworkFlowQueryAuthoring.test.ts) | Tests captured query attempts, stale outcomes, and rejection attribution without draft loss. |
| [useNetworkFlowQueryAuthoring.ts](useNetworkFlowQueryAuthoring.ts) | Binds query drafts, captured Apply attempts, accepted settings, and field feedback to React. |
| [useNetworkFlowRejectedRowsController.ts](useNetworkFlowRejectedRowsController.ts) | Rejected-row query state and cancellation controller. |
| [useNetworkFlowRowsController.ts](useNetworkFlowRowsController.ts) | Accepted table-row query state and cancellation controller. |

## Saved graphs

| File | Responsibility |
| --- | --- |
| [NetworkFlowSavedGraphPanel.tsx](NetworkFlowSavedGraphPanel.tsx) | Saved graph catalog, declaration editing, action review, and result controls. |
| [SavedGraphController.test.ts](SavedGraphController.test.ts) | Tests saved graph operation capture, duplicate admission, and bounded observation scheduling. |
| [SavedGraphController.ts](SavedGraphController.ts) | Saved graph catalog and captured lifecycle operations with job observation and explicit recovery. |
| [savedGraphObservation.test.ts](savedGraphObservation.test.ts) | Tests serial bounded job observation, scoped terminal references, and interrupted reads. |
| [savedGraphObservation.ts](savedGraphObservation.ts) | Validates saved graph job targets and performs bounded serial job observation. |
| [savedGraphOperation.ts](savedGraphOperation.ts) | Saved graph authority, intent, immutable attempts, and receipt correlation. |
| [savedGraphReadFailure.ts](savedGraphReadFailure.ts) | Classifies saved graph read failures by affected surface and recovery disposition. |
| [SavedGraphRecovery.test.ts](SavedGraphRecovery.test.ts) | Tests saved graph retirement during reads and definite cancellation before write dispatch. |
| [SavedGraphResultNavigation.test.ts](SavedGraphResultNavigation.test.ts) | Tests exact result continuity across rename, refresh, failed reads, and removed bindings. |
| [SavedGraphResultNavigation.ts](SavedGraphResultNavigation.ts) | Retains saved graph results and contributor context across navigation and refresh. |
| [savedGraphTestFixtures.ts](savedGraphTestFixtures.ts) | Saved graph authorities, resources, receipts, deferred reads, bindings, and result fixtures. |
| [savedGraphTransport.test.ts](savedGraphTransport.test.ts) | Tests exact saved graph replay, scoped job receipts, and malformed acknowledgement handling. |
| [useNetworkFlowSavedGraphController.ts](useNetworkFlowSavedGraphController.ts) | Workspace binding for saved graph discovery, selection, actions, and results. |
| [useNetworkFlowSavedGraphOwner.ts](useNetworkFlowSavedGraphOwner.ts) | React lifetime binding for persistent saved graph operation ownership. |

## Table lifecycle

| File | Responsibility |
| --- | --- |
| [NetworkFlowTableController.test.ts](NetworkFlowTableController.test.ts) | Tests table-operation role admission, duplicate suppression, captured bytes, and dispatch fencing. |
| [NetworkFlowTableController.ts](NetworkFlowTableController.ts) | Analytical table lifecycle draft, mutation, acknowledgement, and recovery ownership. |
| [NetworkFlowTableLifecycle.tsx](NetworkFlowTableLifecycle.tsx) | Analytical table lifecycle controls, review surfaces, and operation recovery. |
| [networkFlowTableOperation.ts](networkFlowTableOperation.ts) | Table action authority, display-name validation, captured requests, and operation outcomes. |
| [tableConsumerContinuity.test.tsx](tableConsumerContinuity.test.tsx) | Tests graph continuity through table rename/removal and fresh source selection on recompute. |
| [tableLifecycleTestFixtures.ts](tableLifecycleTestFixtures.ts) | Analytical table resources and actor/incident authority fixtures for lifecycle tests. |
| [useNetworkFlowTableController.ts](useNetworkFlowTableController.ts) | Table discovery, active selection, load state, and access-loss controller. |
| [useNetworkFlowTableOwner.ts](useNetworkFlowTableOwner.ts) | React lifetime binding for analytical table lifecycle operations. |

## Shell recovery contributions

Import, table lifecycle and Indicator linking publish safe metadata through the
shared recovery boundary. Their existing controllers retain authorization,
captured requests, receipts and action eligibility. Import recovery sequence,
table dialog identity and Indicator-link work identity stay stable through their
captured attempts. `networkFlowTableRecoveryItems.ts` and
`networkFlowIndicatorRecoveryItems.ts` project those owner snapshots; the Workbook
imports extension contributions through its existing `NetworkFlowOperations`
facade. Saved-graph recovery remains local; its dialogs detach shell recovery.
Background grid refresh restores focus only while that grid owns focus.

Recovery detaches the graph contributor drawer without clearing the selected
endpoint or invalidating a newly opened Indicator link draft. Explicit graph
selection reattaches its drawer and closes recovery through the common boundary.
