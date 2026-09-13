# workbook/timeline/bulk/

[Parent](../README.md) · [Source overview](../../../README.md)

Timeline bulk tagging and fill workflows over stable record and field identities.

Controllers own one multi-record workflow at a time. They use current-page
stable record identities and semantic commands; transport and DOM state stay
behind injected boundaries.

## Files

| File | Responsibility |
| --- | --- |
| [useTimelineBulkTagController.ts](useTimelineBulkTagController.ts) | Owns Timeline tag-selection state, eligibility/pruning, tag draft, deduplicated semantic submission, refresh, bounded conflict copy, and late-completion invalidation. |
| [useTimelineFillController.ts](useTimelineFillController.ts) | Plans and dispatches versioned Timeline fill commands against stable record/field identities while preserving mutation admission and focus continuity. |

## Tests

| File | Responsibility |
| --- | --- |
| [useTimelineBulkTagController.test.tsx](useTimelineBulkTagController.test.tsx) | Characterizes current-page stable-ID selection, pruning, versioned submission, conflicts, authorization loss, and draft retention. |
| [useTimelineFillController.test.tsx](useTimelineFillController.test.tsx) | Characterizes ordered fill planning, atomic rejection, semantic dispatch, save/refresh sequencing, focus restoration, and conflict activation. |
