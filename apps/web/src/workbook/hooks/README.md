# workbook/hooks/

[Parent](../README.md) · [Source overview](../../README.md)

React coordination for shell lifetime, queries, startup, saved views, imports, and shared surface behavior.

These hooks bind React to workbook owners and compose shell-wide behavior.
Timeline-specific controllers live under
[Timeline hooks](../timeline/hooks/README.md).

## Queries, startup, and saved views

| File | Responsibility |
| --- | --- |
| [useActiveSurfaceSavedViewActions.ts](useActiveSurfaceSavedViewActions.ts) | Binds active-surface saved-view intents to retained action and recovery ownership. |
| [useOwnerReferenceOptions.ts](useOwnerReferenceOptions.ts) | Resolves only the active bounded-context policy's reference requirements through the generic broker. |
| [useWorkbookCandidates.ts](useWorkbookCandidates.ts) | Shared paged workbook candidate reads with query identity and stale-response handling. |
| [useWorkbookProjectionRefreshController.test.tsx](useWorkbookProjectionRefreshController.test.tsx) | Direct tests for initial and sheet-triggered projection refresh ownership. |
| [useWorkbookProjectionRefreshController.ts](useWorkbookProjectionRefreshController.ts) | Initial session/entity and sheet-triggered projection refresh coordinator. |
| [useWorkbookQueryController.test.tsx](useWorkbookQueryController.test.tsx) | Direct tests for exact-view-schema query-state isolation. |
| [useWorkbookQueryController.ts](useWorkbookQueryController.ts) | Schema-keyed query/filter controller and active query-control adapter. |
| [useWorkbookSavedViewController.test.tsx](useWorkbookSavedViewController.test.tsx) | Direct tests for saved-view loading and selection precedence. |
| [useWorkbookSavedViewController.ts](useWorkbookSavedViewController.ts) | Binds active query/layout configuration and selection effects to the session-owned saved-view controller. |
| [useWorkbookStartupController.test.tsx](useWorkbookStartupController.test.tsx) | Direct tests for workbook selection, focus intent, versioning, and URL state. |
| [useWorkbookStartupController.ts](useWorkbookStartupController.ts) | Startup/sheet identity, URL history, focus intent, and workbook-preference controller. |
| [useWorkbookSurfaceQueries.ts](useWorkbookSurfaceQueries.ts) | Surface query loading, invalidation, facade projection, and collaboration port selection. |

## Surface and import bindings

| File | Responsibility |
| --- | --- |
| [useEntityTimelinePreview.ts](useEntityTimelinePreview.ts) | Loads Timeline preview rows for entity-related workbook workflows. |
| [useGenericSurfaceMutationController.ts](useGenericSurfaceMutationController.ts) | Contract-surface mutation state, conflict admission, and refresh coordination over the generic command port. |
| [useNetworkFlowImportBinding.ts](useNetworkFlowImportBinding.ts) | Connects workbook surface state to app-owned Network Flow import lifetime. |
| [useWorkbookImportBinding.ts](useWorkbookImportBinding.ts) | Connects workbook surface state to app-owned workbook import lifetime. |

## Shell lifetime and interaction

| File | Responsibility |
| --- | --- |
| [useIncidentControlsDrawer.ts](useIncidentControlsDrawer.ts) | Incident controls drawer state, selection, and focus restoration. |
| [useWorkbookAuthorizationState.ts](useWorkbookAuthorizationState.ts) | Incident authorization subject and explicit session-role recovery lifecycle. |
| [useWorkbookCollaborationLifecycle.ts](useWorkbookCollaborationLifecycle.ts) | Collaboration invalidation, session projection, and exact active-surface registration. |
| [useWorkbookExtensionAvailability.test.tsx](useWorkbookExtensionAvailability.test.tsx) | Effect-owned discovery and controller subject-lifetime tests. |
| [useWorkbookExtensionAvailability.ts](useWorkbookExtensionAvailability.ts) | Shell-lifetime extension controller, effect-owned discovery, invalidation, and render revision. |
| [useWorkbookIncidentIdentity.ts](useWorkbookIncidentIdentity.ts) | Resolves incident identity/loading state for the workbook shell. |
| [useWorkbookRecoveryFocus.ts](useWorkbookRecoveryFocus.ts) | Deterministic focus transfer across blocked-edit, overflow, and same-field conflict recovery. |
| [useWorkbookSemanticGridFocus.test.tsx](useWorkbookSemanticGridFocus.test.tsx) | Direct tests for semantic grid-entry focus order, lifecycle readiness, and stale-request handling. |
| [useWorkbookSemanticGridFocus.ts](useWorkbookSemanticGridFocus.ts) | Resolves generation-keyed grid-entry requests through the mounted semantic grid handle. |
| [useWorkbookShellInfrastructure.ts](useWorkbookShellInfrastructure.ts) | Incident-scoped adapters, registry-owned mutation runtime, command ports, and disposable reference broker. |
| [useWorkbookShellRuntime.ts](useWorkbookShellRuntime.ts) | Startup, saved-view, query, and layout-state composition facade. |
