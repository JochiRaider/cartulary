# workbook/savedviews/

[Parent](../README.md) · [Source overview](../../README.md)

Saved-view listing and captured create/update/delete operations with retained acknowledgement and recovery.

Saved-view operation lifetime is independent of selector/dialog attachment.
Identity, draft, attempt, receipt, and refresh state remain explicit. The
[component controls](../components/README.md) supply action and recovery UI.

## Files

| File | Responsibility |
| --- | --- |
| [loadSavedViewList.ts](loadSavedViewList.ts) | Loads a complete validated saved-view list through the semantic paged port. |
| [savedViewOperationModel.ts](savedViewOperationModel.ts) | Saved-view authority, subjects, drafts, intents, immutable attempts, and operation snapshots. |
| [WorkbookSavedViewController.ts](WorkbookSavedViewController.ts) | Saved-view catalog and create/update/delete ownership with retained attempts and recovery. |

## Tests

| File | Responsibility |
| --- | --- |
| [WorkbookSavedViewController.test.ts](WorkbookSavedViewController.test.ts) | Tests exact saved-view write capture, synchronous admission, and receipts after presentation detaches. |
