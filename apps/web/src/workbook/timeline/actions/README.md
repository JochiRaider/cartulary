# workbook/timeline/actions/

[Parent](../README.md) · [Source overview](../../../README.md)

Reviewed Timeline capture actions and mention resolution, including optional Entity creation and recovery.

Capture actions and mention operations retain reviewed subjects, exact attempts,
and accepted receipts beyond the active inspector. Optional Entity creation and
mention resolution keep independent recovery state.

## Files

| File | Responsibility |
| --- | --- |
| [reconcileTimelineCaptureReceipt.ts](reconcileTimelineCaptureReceipt.ts) | Reconciles accepted Timeline capture actions against record and History materialization. |
| [reconcileTimelineMentionReceipt.ts](reconcileTimelineMentionReceipt.ts) | Reconciles mention receipts with current source, mention, and Entity projections. |
| [TimelineCandidatePort.ts](TimelineCandidatePort.ts) | Semantic paged Timeline candidate capability for capture-action review. |
| [timelineCaptureActionModel.ts](timelineCaptureActionModel.ts) | Timeline capture authority, action subjects, reviewed values, and eligibility models. |
| [timelineCaptureOwnerFor.ts](timelineCaptureOwnerFor.ts) | Resolves the Timeline capture-action owner for a workbook mutation runtime. |
| [TimelineCaptureRecovery.tsx](TimelineCaptureRecovery.tsx) | Retained Timeline review/supersession replay and reconciliation recovery controls. |
| [timelineMentionAuthority.ts](timelineMentionAuthority.ts) | Derives Timeline mention operation authority from the current workbook binding. |
| [TimelineMentionCandidatePort.ts](TimelineMentionCandidatePort.ts) | Semantic Entity candidate pages for Timeline mention resolution. |
| [timelineMentionCreationModel.ts](timelineMentionCreationModel.ts) | Mention-driven Entity creation reviews, captured attempts, receipts, and operation state. |
| [timelineMentionOperationModel.ts](timelineMentionOperationModel.ts) | Mention subjects, authority, actions, reviews, attempts, and retained resolution state. |
| [timelineMentionOwnerFor.ts](timelineMentionOwnerFor.ts) | Resolves the Timeline mention-operation owner for a workbook mutation runtime. |
| [TimelineMentionRecovery.tsx](TimelineMentionRecovery.tsx) | Separate entity-creation and mention-resolution recovery for retained operations. |
| [TimelineSupersessionEditor.tsx](TimelineSupersessionEditor.tsx) | Timeline supersession reason authoring, replacement selection, and explicit confirmation. |
| [useTimelineCandidates.ts](useTimelineCandidates.ts) | Paged Timeline replacement candidate reads with authority and request-lifetime fencing. |
| [useTimelineCaptureActions.ts](useTimelineCaptureActions.ts) | React binding for reviewed Timeline capture actions and retained operation state. |
| [useTimelineMentionCandidates.ts](useTimelineMentionCandidates.ts) | Paged eligible Entity discovery for Timeline mention resolution. |
| [WorkbookTimelineCaptureActionOwner.ts](WorkbookTimelineCaptureActionOwner.ts) | Timeline review/supersession admission, captured attempts, acknowledgement, and reconciliation ownership. |
| [WorkbookTimelineMentionOperationOwner.ts](WorkbookTimelineMentionOperationOwner.ts) | Owns mention resolution and optional Entity creation with independent attempts and receipts. |

## Tests

| File | Responsibility |
| --- | --- |
| [reconcileTimelineMentionReceipt.test.ts](reconcileTimelineMentionReceipt.test.ts) | Tests mention reconciliation against newer corrections, stale sources, and detached reads. |
| [timelineMentionCandidates.test.tsx](timelineMentionCandidates.test.tsx) | Tests selected Entity identity, accessible dismissal, and independently paged eligible targets. |
| [TimelineSupersessionEditor.test.tsx](TimelineSupersessionEditor.test.tsx) | Tests capture-action eligibility, authored reasons, frozen review, and optional replacement. |
| [WorkbookTimelineCaptureActionOwner.test.ts](WorkbookTimelineCaptureActionOwner.test.ts) | Tests synchronous capture reservation, preparation ordering, and renewed review for changed versions. |
| [WorkbookTimelineMentionOperationOwner.test.ts](WorkbookTimelineMentionOperationOwner.test.ts) | Tests retained entity-creation receipts and independent mention-link recovery without duplicate creation. |

Mention presentation refresh carries the exact accepted receipt and its required
source version. The presentation binding lends a continuity token only to that
receipt's captured operation; another mention on the same row cannot borrow it.
The row loader commits a qualifying projection before publishing continuity
evidence. A same-source inspector version update invalidates review without
cancelling eligible completion. Entity-creation refresh remains independent of
mention-link refresh. Terminal refresh failure retains the receipt and settles
at an eligible visible fallback; read recovery does not replay the mutation or
revive cancelled focus. Cancellation ownership remains attached while the Grid
Adapter awaits a virtualized focus target.
