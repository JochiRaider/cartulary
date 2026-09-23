# workbook/hooks/

`ordinaryInspectorRecovery.test.tsx` characterizes ordinary explicit patch response
loss and acknowledgement independent of presentation refresh.

[Parent](../README.md) · [Source overview](../../README.md)

React coordination for shell lifetime, queries, startup, saved views, imports, and shared surface behavior.

These hooks bind React to workbook owners and compose shell-wide behavior.
Timeline-specific controllers live under
[Timeline hooks](../timeline/hooks/README.md).

## Queries, startup, and saved views

| File | Responsibility |
| --- | --- |
| [useActiveSurfaceSavedViewActions.ts](useActiveSurfaceSavedViewActions.ts) | Binds active-surface saved-view intents to retained action and recovery ownership. |
| [useWorkbookProjectionRefreshController.test.tsx](useWorkbookProjectionRefreshController.test.tsx) | Direct tests for initial and sheet-triggered projection refresh ownership. |
| [useWorkbookProjectionRefreshController.ts](useWorkbookProjectionRefreshController.ts) | Initial session/entity and sheet-triggered projection refresh coordinator. |
| [useWorkbookQueryController.test.tsx](useWorkbookQueryController.test.tsx) | Direct tests for exact-view-schema query-state isolation. |
| [useWorkbookQueryController.ts](useWorkbookQueryController.ts) | Schema-keyed query/filter controller and active query-control adapter. |
| [useWorkbookSavedViewController.test.tsx](useWorkbookSavedViewController.test.tsx) | Direct tests for saved-view loading and selection precedence. |
| [useWorkbookSavedViewController.ts](useWorkbookSavedViewController.ts) | Binds active query/layout configuration and selection effects to the session-owned saved-view controller. |
| [useWorkbookStartupController.test.tsx](useWorkbookStartupController.test.tsx) | Direct tests for workbook selection, focus intent, versioning, and URL state. |
| [useWorkbookStartupController.ts](useWorkbookStartupController.ts) | Startup/sheet identity, URL history, focus intent, and workbook-preference controller. |
| [useWorkbookSurfaceQueries.ts](useWorkbookSurfaceQueries.ts) | Surface query loading, ordinary accepted-version integration, invalidation and collaboration port selection. |

## Surface and import bindings

| File | Responsibility |
| --- | --- |
| [useEntityTimelinePreview.ts](useEntityTimelinePreview.ts) | Loads Timeline preview rows for entity-related workbook workflows. |
| [useGenericSurfaceMutationController.ts](useGenericSurfaceMutationController.ts) | Existing-record mutation state and refresh coordination; ordinary creation is borrowed from its retained owner. |
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
| [useWorkbookShellInfrastructure.ts](useWorkbookShellInfrastructure.ts) | Pure mounted adapters, command ports and reference broker over the committed application-owned mutation runtime. |
| [useWorkbookShellRuntime.ts](useWorkbookShellRuntime.ts) | Startup, saved-view, query, and layout-state composition facade. |

`useGenericSurfaceMutationController` submits existing-record edits through the
retained explicit owner. The one-shot generic and Entity patch ports are retired;
creation and specialized route commands keep their own admission boundaries.
The accepted baseline supplies record identity and version; presentation does
not report another mutation around the retained operation. The generic hook
subscribes to the selected record's blocking boolean, so unrelated retained
operations cannot trigger its render. Local pending controls
read the selected record's retained admission/recovery state, while aggregate save
status counts authoritative settlement independently of follow-up reads.

Committed grid denials bind to the existing collaboration authorization recovery
through `useWorkbookCollaborationLifecycle`. `useWorkbookIncidentIdentity` reacts
to incident closure with its version-fenced identity read; reopening uses the
retained incident lifecycle owner. Neither hook owns raw editor drafts.

Workbook query continuation uses a stable incident-wide invalidation callback
with current sheet readers. `useWorkbookProjectionRefreshController` observes
initial authorization separately from Entity query changes, preventing reference
broker replacement from creating a read/recovery loop. Grid focus bindings retain
semantic anchors at departure and detach editor presentation for explicit browsing.

`useWorkbookRecoveryFocus.ts` routes status actions by stable semantic cause and
records explicit resolver focus intent. Shared recovery presentation owns closing,
completion fallback and inspector/dialog coordination; mutation snapshots contain
no disclosure state.

## Authoring candidate discovery

| File | Responsibility |
| --- | --- |
| [useWorkbookCandidateDiscovery.ts](useWorkbookCandidateDiscovery.ts) | React attachment and semantic authority/target/query fencing for candidate discovery. |

useWorkbookAuthoringInventory uses the typed discovery lifecycle for the finite
implemented-surface inventory; member selection uses the membership read directly.
The accumulating useWorkbookCandidates hook is retired. The shell supplies account,
session, incident and authority generation through WorkbookCandidateAuthorityContext,
including authoring attached through recovery panels.
