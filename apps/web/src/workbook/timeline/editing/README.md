# workbook/timeline/editing/

[Parent](../README.md) · [Source overview](../../../README.md)

Timeline scalar editor drafts and registered inputs keyed by semantic row, field, and surface identity.

Editor state uses semantic record, field, and surface identity. Authoritative
refresh and grid coordinates do not replace those identities or erase retained
invalid text.

## Files

| File | Responsibility |
| --- | --- |
| [useTimelineEditorDraftRegistry.ts](useTimelineEditorDraftRegistry.ts) | Binds retained local values to currently mounted input references and row subscriptions; clears only submitted values on acknowledgement. |

## Tests

| File | Responsibility |
| --- | --- |
| [useTimelineEditorDraftRegistry.test.tsx](useTimelineEditorDraftRegistry.test.tsx) | Characterizes invalid-text preservation, grid/inspector separation, submitted-value cleanup, semantic input registration, row removal, surface detachment, and runtime retirement. |
