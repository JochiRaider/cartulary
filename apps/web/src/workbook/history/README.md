# workbook/history/

[Parent](../README.md) · [Source overview](../../README.md)

Record History browsing, captured rollback/restore operations, acknowledgement, and recovery ownership.

This layer retains History operation state beyond inspector attachment.
[Inspector modules](../inspector/README.md) bind subjects, focus, and presentation;
[adapters](../adapters/README.md) own protocol validation. An accepted operation
receipt and successful projection refresh are separate observations.

The browsing model is the sole accepted-page representation. Reconciliation and
operation review retain complete `HistoryPage` values, including continuation
metadata; inspector attachment never maintains a second loaded-data phase.
Timeline shares one presentation state instance with its subject composition and
keeps its FIFO coordination port. Missing coordination fails action preparation
locally rather than issuing an uncoordinated command. Shared record identity and
display context live in `ports/WorkbookRecordSubject.ts`, independently of
inspector attachment and operation lifetime.

## Files

| File | Responsibility |
| --- | --- |
| [HistoryPageLookup.ts](HistoryPageLookup.ts) | Shared bounded read search, page validation, cancellation and authority fencing. |
| [workbookHistoryReview.ts](workbookHistoryReview.ts) | Minimal authorized read locator and three-page review policy, distinct from pending actions. |
| [WorkbookHistoryReview.tsx](WorkbookHistoryReview.tsx) | Explicit Recovery review using current row History and existing confirmed actions. |
| [HistoryActionLookup.ts](HistoryActionLookup.ts) | Incremental paged lookup for a retained History action without treating paused scans as absence. |
| [HistoryLookupFeedback.tsx](HistoryLookupFeedback.tsx) | History action lookup progress, paused-search, failure, and continuation feedback. |
| [historyOperationPresentation.ts](historyOperationPresentation.ts) | Projects History operation state into user-facing status. |
| [workbookHistoryBrowsing.ts](workbookHistoryBrowsing.ts) | Pure History browsing state transitions for initial, continuation, and refresh reads. |
| [WorkbookHistoryContext.ts](WorkbookHistoryContext.ts) | React access to History runtime, surface refresh, action permissions, and pending records. |
| [workbookHistoryItem.ts](workbookHistoryItem.ts) | Normalizes History items and builds stable rollback targets and pending action identities. |
| [WorkbookHistoryLocalStatus.tsx](WorkbookHistoryLocalStatus.tsx) | Record-local History operation status. |
| [workbookHistoryOperation.ts](workbookHistoryOperation.ts) | History authority, intents, captured attempts, receipts, semantic ports, and bindings. |
| [workbookHistoryPage.ts](workbookHistoryPage.ts) | History page requests, provenance, scope equality, and paging validation. |
| [WorkbookHistoryRecovery.tsx](WorkbookHistoryRecovery.tsx) | Retained History operation replay, lookup, and reconciliation recovery presentation. |
| [WorkbookRecordHistoryOwner.ts](WorkbookRecordHistoryOwner.ts) | Record History operation admission, captured attempts, acknowledgement, and recovery ownership. |

Shared semantic fixtures live in [application test support](../../testing/workbookHistoryTestSupport.ts). Production imports are forbidden by the authored import-boundary policy. Consumers import History item and response types directly from their adapter declaration owner; lookup state belongs to `HistoryPageLookup`. No forwarding exports are maintained.

## Tests

| File | Responsibility |
| --- | --- |
| [WorkbookBatchHistoryReview.test.tsx](WorkbookBatchHistoryReview.test.tsx) | Explicit off-window review, pruning lifetime, current action gates and existing reversal ownership. |
| [HistoryActionLookup.test.ts](HistoryActionLookup.test.ts) | Tests bounded History action lookup, explicit continuation, and failed-page retry. |
| [workbookHistoryBrowsing.characterization.test.tsx](workbookHistoryBrowsing.characterization.test.tsx) | Tests server-owned continuation and retained accepted History after refresh failure. |
| [workbookHistoryBrowsing.test.ts](workbookHistoryBrowsing.test.ts) | Tests short/empty page continuation, stable server ordering, deduplication, and eligibility refresh. |
| [WorkbookHistoryRecovery.test.tsx](WorkbookHistoryRecovery.test.tsx) | Tests retained acknowledgements, read-only refresh retry, and session-bound recovery visibility. |
| [WorkbookRecordHistoryOwner.test.ts](WorkbookRecordHistoryOwner.test.ts) | Tests History authority lifetime, retained operations after detachment, and late effects. |
