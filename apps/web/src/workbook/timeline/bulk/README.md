# workbook/timeline/bulk/

[Parent](../README.md) · [Source overview](../../../README.md)

Timeline clear, bulk tagging and fill workflows over stable record and field identities.

Controllers capture current-page record identities and submitted values. The
Workbook batch owner retains admitted operations, overlap ordering, receipts,
conflicts and recovery independently of controller lifetime.

## Files

| File | Responsibility |
| --- | --- |
| [TimelineBulkTagAuthoring.ts](TimelineBulkTagAuthoring.ts) | Retains exact raw command text/revision across presentation detachment; conceals suspended authority and fences obsolete bindings. |
| [useTimelineBulkTagController.ts](useTimelineBulkTagController.ts) | Owns authorized accepted-query selection and pruning, supplies committed readable checkbox context, revalidates the complete intended set and submits native delivery identity to retained ownership. |
| [useTimelineFillController.ts](useTimelineFillController.ts) | Plans keyboard/pointer scalar fill over stable records and fields with existing target restrictions; captures source value and preceding autosaves. |

## Tests

| File | Responsibility |
| --- | --- |
| [timelineBulkTagInteraction.test.tsx](../timelineBulkTagInteraction.test.tsx) | Tests membership/readiness separation, captured submission, delivery identity, admission failure and leaf draft retention. |
| [useTimelineFillController.test.tsx](useTimelineFillController.test.tsx) | Tests ordered scalar fill planning, unsupported-target rejection and semantic admission to retained recovery. |

`useTimelineClearController.ts` owns the shared selected-cell clear planner and
admission for Delete and the view-bar action. `useTimelineClearController.test.tsx`
covers exact membership, capabilities, drafts and retained dispatch.

`createTimelineBulkTagReadiness.ts` reads existing draft revisions and pending
queue/conflict ownership. It permits captured pending saves and explains failed
or unsubmitted selected edits without flushing or mirroring them. `WorkbookTimelineMutationOwner` retains the raw tag value through its independently
observable `TimelineBulkTagAuthoring` slot. The leaf observes it directly and keeps
focus, native caret and revision-scoped feedback local.
