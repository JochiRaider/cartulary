# workbook/collaboration/

[Parent](../README.md) · [Source overview](../../README.md)

Workbook interpretation of decoded collaboration events, presence, reset, and authorization recovery.

The [browser session](../../collaboration/README.md) owns WebSocket mechanics.
This layer interprets decoded events and coordinates workbook state.
Authorization recovery is injected by app composition; these modules do not
own authentication transport.

## Files

| File | Responsibility |
| --- | --- |
| [useWorkbookCollaborationCoordinator.ts](useWorkbookCollaborationCoordinator.ts) | React external-store subscription, Strict Mode-safe lifetime lease, and collaboration-session adapter for the coordinator. |
| [workbookAuthorizationRecoveryMachine.ts](workbookAuthorizationRecoveryMachine.ts) | Pure generation-keyed authorization recovery, retry, role, and stale-settlement transitions. |
| [WorkbookCollaborationCoordinator.ts](WorkbookCollaborationCoordinator.ts) | Sole shell-lifetime effect coordinator for decoded collaboration plans, injected timing, authorization recovery, reset settlement, and active-surface reconciliation. |
| [workbookCollaborationEventPlan.ts](workbookCollaborationEventPlan.ts) | Pure closed routing from session events to Workbook-owned effects. |
| [workbookCollaborationInvalidationPlan.ts](workbookCollaborationInvalidationPlan.ts) | Pure ordered invalidation plans for reset, authorization, role, closure, and disposal transitions. |
| [workbookCollaborationMessages.ts](workbookCollaborationMessages.ts) | Workbook presence and Mention action payload helpers. |
| [workbookCollaborationResetMachine.ts](workbookCollaborationResetMachine.ts) | Pure session-and-surface-keyed reset admission model. |
| [workbookCollaborationTiming.ts](workbookCollaborationTiming.ts) | Clock and cancellable scheduler capabilities used by the coordinator effect shell. |
| [workbookPresencePresentation.ts](workbookPresencePresentation.ts) | Projects deduplicated, scoped workbook presence into display modes and user summaries. |
| [workbookPresenceProjection.ts](workbookPresenceProjection.ts) | Pure exact-key presence snapshot/delta projection and active-sheet derivation. |
| [workbookPresencePublicationMachine.ts](workbookPresencePublicationMachine.ts) | Pure bounded debounce and stale-callback rejection for outgoing presence. |
| [workbookSurfacePort.ts](workbookSurfacePort.ts) | Active-surface identity and live-row reconciliation capabilities. |

## Tests

| File | Responsibility |
| --- | --- |
| [WorkbookCollaborationCoordinator.test.ts](WorkbookCollaborationCoordinator.test.ts) | Tests for presence, reset, cleanup ordering, authorization recovery, role downgrade, access loss, and late-work rejection. |
| [workbookCollaborationMessages.test.ts](workbookCollaborationMessages.test.ts) | Tests for Base, saved-view, and extension-workspace presence message construction. |
| [workbookCollaborationTransitionModels.test.ts](workbookCollaborationTransitionModels.test.ts) | Tests closed event routing and ordered reset, authorization, role, closure, and disposal plans. |
