# workbook/timeline/bulk/

[Parent](../README.md) · [Source overview](../../../README.md)

Timeline bulk tagging and fill workflows over stable record and field identities.

Controllers capture current-page record identities and submitted values. The
Workbook batch owner retains admitted operations, overlap ordering, receipts,
conflicts and recovery independently of controller lifetime.

## Files

| File | Responsibility |
| --- | --- |
| [useTimelineBulkTagController.ts](useTimelineBulkTagController.ts) | Owns current-page selection, eligibility/pruning, tag draft and same-frame delivery guarding; submits prepared intent to retained ownership. |
| [useTimelineFillController.ts](useTimelineFillController.ts) | Plans keyboard/pointer scalar fill over stable records and fields with existing target restrictions; captures source value and preceding autosaves. |

## Tests

| File | Responsibility |
| --- | --- |
| [useTimelineBulkTagController.test.tsx](useTimelineBulkTagController.test.tsx) | Tests current-page IDs, pruning, captured submission, same-frame delivery guarding, authorization and draft retention. |
| [useTimelineFillController.test.tsx](useTimelineFillController.test.tsx) | Tests ordered scalar fill planning, unsupported-target rejection and semantic admission to retained recovery. |
