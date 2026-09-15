# workbook/utils/

[Parent](../README.md) · [Source overview](../../README.md)

Reusable workbook clipboard, queue, presence, reconciliation, recovery, formatting, and style helpers.

Keep Timeline-specific helpers in its [models](../timeline/models/README.md)
and cross-feature helpers in [shared](../../shared/README.md). These utilities
serve workbook consumers without owning a feature workflow.

## Files

| File | Responsibility |
| --- | --- |
| [freezeWorkbookValue.ts](freezeWorkbookValue.ts) | Recursively freezes captured workbook values for immutable operation state. |
| [workbookClipboard.ts](workbookClipboard.ts) | Adapts the Grid Adapter representation decoder and source error callback; applies the non-header 500-row bound. |
| [workbookEditRecoveryPresentation.ts](workbookEditRecoveryPresentation.ts) | Projects pending edit and conflict state into workbook recovery presentation. |
| [workbookPendingQueue.ts](workbookPendingQueue.ts) | Pending-save queue capacity, save-state, replay, conflict, and public-error helpers. |
| [workbookPresence.ts](workbookPresence.ts) | Presence input/type helpers and presence matching helpers. |
| [workbookRowReconciliation.ts](workbookRowReconciliation.ts) | Record-identity and row-version reconciliation that preserves unchanged row references. |
| [workbookStatusSecondary.ts](workbookStatusSecondary.ts) | Selects scoped secondary workbook status and actions by generated priority. |
| [workbookStyles.ts](workbookStyles.ts) | Shared workbook style primitives. |
| [workbookValueFormat.ts](workbookValueFormat.ts) | Grid/workbook value formatting helpers. |

## Tests

| File | Responsibility |
| --- | --- |
| [GridAdapter.anchor.test.ts](GridAdapter.anchor.test.ts) | Workbook interaction grid-adapter anchor behavior tests. |
| [workbookClipboard.test.ts](workbookClipboard.test.ts) | Tests explicit representation dispatch, scalar comma/quote preservation, deterministic rejection and row bounds. |
| [workbookPendingQueue.test.ts](workbookPendingQueue.test.ts) | Tests queue scope isolation, capacity, replay, and settlement behavior. |
| [workbookPresence.test.tsx](workbookPresence.test.tsx) | Tests unique-user presence counts and activity-first display ordering. |
| [workbookRowReconciliation.test.ts](workbookRowReconciliation.test.ts) | Tests for sparse row replacement, removal, drafts, and row-version reference reuse. |
| [workbookStatusSecondary.test.ts](workbookStatusSecondary.test.ts) | Tests generated secondary-status priority and exclusion of inactive-surface candidates. |
| [workbookValueFormat.test.ts](workbookValueFormat.test.ts) | Tests grid value formatting preserves primitive display strings. |
