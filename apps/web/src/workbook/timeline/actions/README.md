# workbook/timeline/actions/

[Parent](../README.md) · [Source overview](../../../README.md)

Reviewed Timeline capture actions and mention resolution, including optional Entity creation and recovery.

Capture actions and mention operations retain reviewed subjects, exact attempts,
and accepted receipts beyond the active inspector. Optional Entity creation and
mention resolution keep independent recovery state.

Timeline mention target discovery owns one authorized Entity search page separately
from the displayed Entity sheet. The adapter supplies the generated policy: one
page of 100 candidates, ten previous cursor checkpoints, and 250ms settled typing.
The owner captures each immutable search/read request; composition defers reads,
replacement aborts the old chain, and late responses cannot enter the new scope.
Retry repeats the captured failed read. Previous/Next navigate pages; Restart
reaches history outside the checkpoint window through a cursor-free chain.

The minimal selected observation is retained separately across same-query paging
and recoverable failures. Editing search invalidates it; Restart marks it stale
until observed again. Current authority loss conceals targets through the
workbook authority owner. Selection never resolves a mention: explicit actions
remain with their mutation and recovery owner.

The controlled combobox owns popup and keyboard presentation only. Read status,
paging and recovery remain in the labelled chooser region, outside listbox
options. Disappearing read controls return focus only while that interaction
still owns focus; newer navigation keeps the analyst's destination. Unrelated
native pickers and shared single-page candidate services retain their consumers.

The [WAI-ARIA combobox pattern](https://www.w3.org/WAI/ARIA/apg/patterns/combobox/)
guides manual acceptance: the active option has popup selection semantics while
the source owner changes its selected identity only after acceptance. DOM focus
stays in the input through active-descendant navigation. Native text editing and
Escape dismissal precede outer inspector handling.

Supersession Review preparation reserves its row through the capture owner but
does not admit a write, create a transaction identity, or appear as action
recovery. The editor keeps Review mounted and focusable with a stable name and
busy state while the owner checks the committed target and optional replacement.
Only the still-current attachment, review inputs, authority, and interaction may
open confirmation and transfer focus. Newer navigation retires that presentation
transition; another explicit Review is required. Cancel or detachment aborts the
authoring preparation without cancelling an earlier admitted action. Confirm
alone enters the retained mutation and exact-replay lifetime.

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
| [useTimelineMentionCandidates.ts](useTimelineMentionCandidates.ts) | Accumulated eligible Entity pages, exact failed-read identity, restart and authority/scope fencing for mention resolution. |
| [WorkbookTimelineCaptureActionOwner.ts](WorkbookTimelineCaptureActionOwner.ts) | Timeline review/supersession admission, captured attempts, acknowledgement, and reconciliation ownership. |
| [WorkbookTimelineMentionOperationOwner.ts](WorkbookTimelineMentionOperationOwner.ts) | Owns mention resolution and optional Entity creation with independent attempts and receipts. |

## Tests

| File | Responsibility |
| --- | --- |
| [reconcileTimelineMentionReceipt.test.ts](reconcileTimelineMentionReceipt.test.ts) | Tests mention reconciliation against newer corrections, stale sources, and detached reads. |
| [timelineMentionCandidates.test.tsx](timelineMentionCandidates.test.tsx) | Tests selected Entity identity, exact retry/restart, stale eligibility, focus continuity, authority and independently paged targets. |
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
