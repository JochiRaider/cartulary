# workbook/savedviews/

[Parent](../README.md) · [Source overview](../../README.md)

Bounded discovery, addressed resource observations and captured create/update/delete operations with retained acknowledgement and recovery.

Saved-view operation lifetime is independent of selector/dialog attachment.
Identity, draft, attempt, receipt, and refresh state remain explicit. The
[component controls](../components/README.md) supply action and recovery UI.

## Files

| File | Responsibility |
| --- | --- |
| [SavedViewDiscovery.ts](SavedViewDiscovery.ts) | Retains one schema-scoped page of 50 candidates and ten previous page-start cursors; no automatic traversal. |
| [SavedViewResourceObserver.ts](SavedViewResourceObserver.ts) | Retains independently authorized resources only for selection, activation, operation and inspected preference pointers. |
| [savedViewOperationModel.ts](savedViewOperationModel.ts) | Saved-view authority, subjects, drafts, intents, immutable attempts, and operation snapshots. |
| [WorkbookSavedViewController.ts](WorkbookSavedViewController.ts) | Selection admission and create/update/delete ownership with retained attempts and recovery. |

## Tests

| File | Responsibility |
| --- | --- |
| [WorkbookSavedViewController.test.ts](WorkbookSavedViewController.test.ts) | Tests exact saved-view write capture, synchronous admission, and receipts after presentation detaches. |

Opening the browser revalidates its retained page. Previous refetches a retained
checkpoint; First and Refresh start a fresh chain only after acceptance. Dismissal
cancels browsing and activation. Schema/authority replacement retires discovery.
Candidate focus and paging never apply query/layout; explicit activation validates
the addressed resource and the captured working revision. Recovery observes its
captured ID; uncertain creation is never resolved by catalog matching.
