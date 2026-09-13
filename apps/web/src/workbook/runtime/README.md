# workbook/runtime/

[Parent](../README.md) · [Source overview](../../README.md)

Workbook mutation lifetime, pending replay, conflict recovery, write coordination, and status projection.

The runtime coordinates pending work, explicit operations, conflicts, and
surface refresh obligations. It stays independent of Timeline implementation;
surface-specific commands enter through registered semantic capabilities.

## Files

| File | Responsibility |
| --- | --- |
| [useWorkbookMutationRuntime.ts](useWorkbookMutationRuntime.ts) | React external-store subscription hook for shell-owned workbook mutation state. |
| [WorkbookClientTransactionLedger.ts](WorkbookClientTransactionLedger.ts) | Bounded client-transaction identity retention and pending-queue settlement lookup. |
| [workbookConflictModel.ts](workbookConflictModel.ts) | Same-field conflict parsing and common envelope types. |
| [WorkbookConflictStore.ts](WorkbookConflictStore.ts) | Conflict registration, compatible draft preservation, refresh commands, and panel state. |
| [WorkbookExplicitPatchOwner.ts](WorkbookExplicitPatchOwner.ts) | Retained explicit record-patch ownership with guarded admission, replay, and recovery. |
| [workbookLifecycleModel.ts](workbookLifecycleModel.ts) | Shared load, refresh, save, conflict, and recovery lifecycle reducer. |
| [WorkbookManagedPatchDriver.ts](WorkbookManagedPatchDriver.ts) | Managed-patch admission, transport dispatch, settlement, local projection, conflict registration, and refresh. |
| [WorkbookMutationDriverRegistry.ts](WorkbookMutationDriverRegistry.ts) | Closed managed-patch/Timeline-row owner envelopes, exact driver registration, duplicate rejection, and absence-safe dispatch. |
| [WorkbookMutationRuntime.ts](WorkbookMutationRuntime.ts) | Shell-lifetime facade over queue coordination, conflicts, surface registration, managed patches, retry, transaction settlement, and status projection. |
| [WorkbookMutationRuntimeRegistry.ts](WorkbookMutationRuntimeRegistry.ts) | App-owned registry retaining one incident-scoped mutation runtime across shell detachment and retiring it on scope replacement. |
| [workbookMutationStatusProjector.ts](workbookMutationStatusProjector.ts) | Pure queue/conflict/explicit-operation status and save-state projection. |
| [workbookPendingMutationSettlement.ts](workbookPendingMutationSettlement.ts) | Maps semantic mutation failures to common queue settlement outcomes. |
| [workbookPendingReplayRuntime.ts](workbookPendingReplayRuntime.ts) | Pending-replay runtime state, admission contracts, and refresh barriers. |
| [WorkbookRetryScheduler.ts](WorkbookRetryScheduler.ts) | Injected, single-flight retry scheduling and cancellation. |
| [WorkbookRuntimeLifecycle.ts](WorkbookRuntimeLifecycle.ts) | Listener lifetime, coalesced drain notification, and terminal disposal. |
| [workbookRuntimePorts.ts](workbookRuntimePorts.ts) | Clock and scheduler ports with the browser composition defaults. |
| [WorkbookSurfaceRegistry.ts](WorkbookSurfaceRegistry.ts) | Surface command registration, replacement-safe cleanup, and retained refresh debt. |

## Tests

| File | Responsibility |
| --- | --- |
| [workbookConflictModel.test.ts](workbookConflictModel.test.ts) | Tests for conflict parsing, queue entries, resolution payloads, and collection actions. |
| [WorkbookEntityMergeAdmission.test.ts](WorkbookEntityMergeAdmission.test.ts) | Tests merge coordination with queued/direct writes and overlapping participant reservation. |
| [WorkbookExplicitPatchOwner.test.ts](WorkbookExplicitPatchOwner.test.ts) | Tests explicit Task patch reservation, guarded-field review, and exact uncertain replay. |
| [WorkbookMutationRuntime.test.ts](WorkbookMutationRuntime.test.ts) | Tests for shell-lifetime queue retention, autosave, refresh debt, conflicts, and mutation coordination. |
| [WorkbookRuntimeResponsibilities.test.ts](WorkbookRuntimeResponsibilities.test.ts) | Deterministic tests for responsibility boundaries, injected time/scheduling, registration cleanup, conflict drafts, transaction settlement, and disposal. |
