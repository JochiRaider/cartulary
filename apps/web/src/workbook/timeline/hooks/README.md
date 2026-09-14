# workbook/timeline/hooks/

[Parent](../README.md) · [Source overview](../../../README.md)

Timeline query, mutation, inspector, owner-action, and interaction coordination.

Hooks execute owner-local effects over semantic capabilities and pure
[models](../models/README.md). Root assembly belongs in
[composition](../composition/README.md); captured action ownership belongs in
[actions](../actions/README.md).

## Interaction and continuity

| File | Responsibility |
| --- | --- |
| [useTimelineClipboardPasteController.ts](useTimelineClipboardPasteController.ts) | Captures and admits semantic table targets and preceding autosaves; preserves scalar paste with its editor/creation owner. |
| [useTimelineGridAnchorController.ts](useTimelineGridAnchorController.ts) | Resolves Timeline grid anchors, paste targets, selected cells, and focus anchors across committed and draft rows. |
| [useTimelineGridInteractions.ts](useTimelineGridInteractions.ts) | Coordinates Timeline grid refs, keyboard helpers, and grid interaction commands. |
| [useTimelineKeyboardController.ts](useTimelineKeyboardController.ts) | Owns Timeline scalar/collection editor keys, grid navigation/range commands, work-area shortcuts, event consumption, and focus priority. |
| [useTimelineViewportContinuityController.ts](useTimelineViewportContinuityController.ts) | Coordinates Timeline scroll snapshots, focus restoration, continuity tokens, and entity-refresh barriers. |

## Queries and committed rows

| File | Responsibility |
| --- | --- |
| [useTimelineCommittedRecordIdle.ts](useTimelineCommittedRecordIdle.ts) | Waits for one committed Timeline record to become mutation-idle and performs at most one authoritative refresh when its committed version is missing. |
| [useTimelineCommittedRows.ts](useTimelineCommittedRows.ts) | Derives committed Timeline row collections from row/runtime state. |
| [useTimelineRows.ts](useTimelineRows.ts) | Owns Timeline row state, the stable row ref, initial draft row, monotonic draft allocation, and semantic replace/update commands. |
| [useTimelineRowsLoader.ts](useTimelineRowsLoader.ts) | Executes the pure load machine around exact query, freshness, local-draft hydration, created-row pinning, access-loss, and continuity boundaries. |
| [useTimelineWorkbookRuntime.ts](useTimelineWorkbookRuntime.ts) | Reduces Timeline lifecycle state and translates shell-owned query commands into deterministic runtime transitions. |

## Mutation and recovery

| File | Responsibility |
| --- | --- |
| [useTimelineConflictProjectionAdapter.ts](useTimelineConflictProjectionAdapter.ts) | Timeline render-state adapter for shell-owned same-field conflict registration, projection, and resolution. |
| [useTimelineConflicts.ts](useTimelineConflicts.ts) | Coordinates Timeline same-field conflict state. |
| [useTimelineMutationCommands.ts](useTimelineMutationCommands.ts) | Applies pure scalar/collection admission plans, publishes optimistic rows, and queues exact Timeline owner envelopes. |
| [useTimelineMutationDriver.ts](useTimelineMutationDriver.ts) | Registers the exact Timeline row driver and applies owner-local admission, revalidation, settlement, conflict, discard, and accepted-result plans. |
| [useTimelineMutationRuntimeBindings.ts](useTimelineMutationRuntimeBindings.ts) | Lifecycle-registers concrete Timeline refresh, conflict-application, focus-restoration, and blocked-edit-discard commands with the workbook mutation runtime. |
| [useTimelinePendingSaves.ts](useTimelinePendingSaves.ts) | Coordinates Timeline pending-save queue runtime and replay admission. |
| [useTimelineSaveStatePresentation.ts](useTimelineSaveStatePresentation.ts) | Coordinates Timeline save-state labels, pending queue snapshot publication, refresh blocking, runtime drain requests, and beforeunload warning state. |
| [useTimelineSourceWriteCoordination.ts](useTimelineSourceWriteCoordination.ts) | Coordinates Timeline source-owner actions with pending record writes and committed versions. |

## Inspector and owner actions

| File | Responsibility |
| --- | --- |
| [useTimelineCreateRelatedWorkflow.ts](useTimelineCreateRelatedWorkflow.ts) | Coordinates Timeline inspector related-row workflow state, draft values, semantic creation, and Evidence linking. |
| [useTimelineEvidenceAttach.ts](useTimelineEvidenceAttach.ts) | Coordinates semantic Timeline Evidence attachment, validation feedback, and save/continuity handoff. |
| [useTimelineHistoryActions.ts](useTimelineHistoryActions.ts) | Coordinates semantic Timeline history load, rollback, delete, restore, preview, and confirmation actions. |
| [useTimelineHistoryState.ts](useTimelineHistoryState.ts) | Coordinates Timeline history panel and row-history state. |
| [useTimelineInspectorFeatureController.ts](useTimelineInspectorFeatureController.ts) | Owns fail-closed Timeline inspector feature routing, mutually exclusive workflow activation, cancellation, and lifecycle invalidation. |
| [useTimelineInspectorSelection.ts](useTimelineInspectorSelection.ts) | Coordinates selected Timeline row, inspector feedback and closure, row-bound feature invalidation, deleted-row history, and focus-safe inspector interactions. |
| [useTimelineMentionActions.ts](useTimelineMentionActions.ts) | Coordinates semantic Timeline mention resolution, entity creation, undo/review actions, and inspector updates. |
| [useTimelineMentions.ts](useTimelineMentions.ts) | Coordinates Timeline mention-resolution state and actions. |
| [useTimelineObservationSource.ts](useTimelineObservationSource.ts) | Prepares committed Timeline source text for Observation authoring and revalidates source changes. |
