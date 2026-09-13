# workbook/timeline/models/

[Parent](../README.md) · [Source overview](../../../README.md)

Pure Timeline row, mutation, load, action, keyboard, and continuity models.

These models separate pure admission, normalization, reconciliation, and
presentation decisions from React effects. They consume stable owner identities
and typed contracts rather than grid coordinates.

## Mutation and replay

| File | Responsibility |
| --- | --- |
| [timelineAcceptedMutationEffects.ts](timelineAcceptedMutationEffects.ts) | Pure post-acceptance selection, notice, created-row, and continuity effect planning. |
| [timelineConflictState.ts](timelineConflictState.ts) | Timeline-local same-field and grouped-paste conflict state types. |
| [timelineDiscardedReconciliation.ts](timelineDiscardedReconciliation.ts) | Pure discarded-unit reconciliation that reapplies later same-row work in FIFO order. |
| [timelineMutationDriverPlans.test.ts](timelineMutationDriverPlans.test.ts) | Tests exact owner-unit replay admission, committed versions, and mutation settlement plans. |
| [timelineMutationDriverPlans.ts](timelineMutationDriverPlans.ts) | Pure Timeline replay admission, settlement, discard, and accepted-projection decisions. |
| [timelineMutationIntents.ts](timelineMutationIntents.ts) | Exact scalar, collection-action, and draft-create mutation intent construction. |
| [timelineMutationModels.test.ts](timelineMutationModels.test.ts) | Pure intent, deduplication, acceptance, discard, version-ledger, and discriminated-collection evidence. |
| [timelineMutationQueueAdmission.ts](timelineMutationQueueAdmission.ts) | Pure scalar/collection no-op, conflict, duplicate, and exact queue-admission decisions. |
| [timelinePendingSaves.ts](timelinePendingSaves.ts) | Timeline-local pending signature, replay-order, and serial-save references. |

## Rows, versions, and loading

| File | Responsibility |
| --- | --- |
| [timelineAcceptedProjection.ts](timelineAcceptedProjection.ts) | Pure accepted-row replacement, insertion, and bottom-draft projection. |
| [timelineCommittedVersionLedger.ts](timelineCommittedVersionLedger.ts) | Monotonic committed row/version high-water ledger with reference-preserving acceptance. |
| [timelineLoadMachine.test.ts](timelineLoadMachine.test.ts) | Exhaustive load-subject, lifecycle, mutation-race, retry-bound, failure, access-loss, and obligation-join transition evidence. |
| [timelineLoadMachine.ts](timelineLoadMachine.ts) | Pure incident/surface/query/generation/mutation-epoch/source-obligation transitions with explicit load effects. |
| [timelineRowModel.ts](timelineRowModel.ts) | Timeline row envelope decoding, normalization, materialization, and sparse-patch application. |
| [timelineRowsModel.test.ts](timelineRowsModel.test.ts) | Tests for Timeline grid-row materialization. |
| [timelineRowsModel.ts](timelineRowsModel.ts) | Timeline row collection helpers and row-state utilities. |
| [timelineWorkbookRuntime.test.ts](timelineWorkbookRuntime.test.ts) | Deterministic lifecycle transition traces for load, refresh, save, conflict, and recovery state. |
| [timelineWorkbookSurfaceRuntime.ts](timelineWorkbookSurfaceRuntime.ts) | Required shell-owned Timeline composition contract for incident, query, entity, layout, and access-loss services. |
| [workbookRecordFreshness.test.ts](workbookRecordFreshness.test.ts) | Tests for comparable and non-comparable row-version freshness decisions. |
| [workbookRecordFreshness.ts](workbookRecordFreshness.ts) | Pure durable-identity and row-version freshness comparison leaf. |

## Interaction and presentation

| File | Responsibility |
| --- | --- |
| [timelineBulkTagPlan.test.ts](timelineBulkTagPlan.test.ts) | Covers authorization, capability, stable selection, current versions, partial rejection, and stale settlement. |
| [timelineBulkTagPlan.ts](timelineBulkTagPlan.ts) | Pure current-page stable-target validation and subject-keyed bulk-tag settlement policy. |
| [timelineClipboardPastePlan.test.ts](timelineClipboardPastePlan.test.ts) | Tests exact editable paste batches, target identity, and fail-closed authority/shape validation. |
| [timelineClipboardPastePlan.ts](timelineClipboardPastePlan.ts) | Pure Timeline paste authority, shape, field, surface, and stable-target admission policy. |
| [timelineCollectionPresentation.ts](timelineCollectionPresentation.ts) | Discriminated relationship/tag collection items, overflow identity, and accessible hidden labels. |
| [timelineControllerPorts.ts](timelineControllerPorts.ts) | Neutral capability-port, row-store, committed-record-idle, context-menu-position, and replay contracts shared by isolated Timeline controllers. |
| [timelineFieldRegistry.ts](timelineFieldRegistry.ts) | Exhaustive Timeline scalar, collection, readonly, inspector, and focus binding registry. |
| [timelineKeyboardIntentModel.test.ts](timelineKeyboardIntentModel.test.ts) | Tests scalar and collection keyboard intent mapping independently of focus or save effects. |
| [timelineKeyboardIntentModel.ts](timelineKeyboardIntentModel.ts) | Pure mapping of editor and work-area key events to semantic Timeline intents. |
| [timelineLayoutPolicy.ts](timelineLayoutPolicy.ts) | Pure Timeline grouping labels and base/expanded column-width policy. |
| [timelineModelBoundaries.test.ts](timelineModelBoundaries.test.ts) | Tests for Timeline row, payload, binding, normalization, and display helpers. |
| [timelineViewportContinuityModel.test.ts](timelineViewportContinuityModel.test.ts) | Tests for viewport continuity and refresh barrier helpers. |
| [timelineViewportContinuityModel.ts](timelineViewportContinuityModel.ts) | Timeline viewport continuity and entity-refresh barrier helpers. |

## Owner action models

| File | Responsibility |
| --- | --- |
| [timelineEvidenceAttachmentPlan.test.ts](timelineEvidenceAttachmentPlan.test.ts) | Covers Evidence action authorization, capability, surface, selection, identity, and dispatchability checks. |
| [timelineEvidenceAttachmentPlan.ts](timelineEvidenceAttachmentPlan.ts) | Pure selected-target revalidation for Evidence materialization and Timeline linking. |
| [timelineHistoryModel.ts](timelineHistoryModel.ts) | Selects the canonical live/deleted Timeline History subject across selection, deletion, and restore. |
| [timelineMentionActionPlan.test.ts](timelineMentionActionPlan.test.ts) | Covers Mention source/version refresh, typed target validation, access loss, surface changes, and entity-create admission. |
| [timelineMentionActionPlan.ts](timelineMentionActionPlan.ts) | Builds an exact mention-action subject from a committed Timeline row and public mention identity. |
| [workbookMentionChips.ts](workbookMentionChips.ts) | Mention chip state, relationship-field keys, and mention display helpers. |
