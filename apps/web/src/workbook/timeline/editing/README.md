# workbook/timeline/editing/

[Parent](../README.md) · [Source overview](../../../README.md)

Timeline scalar editor drafts and registered inputs keyed by semantic row, field, and surface identity.

Editor state uses semantic record, field, and surface identity. Authoritative
refresh and grid coordinates do not replace those identities or erase retained
invalid text.

The neutral retained store keeps authoring revisions and scalar baselines.
Grid and inspector materialization use their own context. The original eligible
cell resumes its raw text; query omission only removes mounted input references.
Accepted predecessors advance only their written baseline fields, and settlement
clears only captured revisions. Independent saved changes require local review.

## Files

| File | Responsibility |
| --- | --- |
| [useTimelineEditorDraftRegistry.ts](useTimelineEditorDraftRegistry.ts) | Binds retained local values to currently mounted input references and row subscriptions; clears only submitted values on acknowledgement. |

## Tests

| File | Responsibility |
| --- | --- |
| [useTimelineEditorDraftRegistry.test.tsx](useTimelineEditorDraftRegistry.test.tsx) | Characterizes invalid-text preservation, grid/inspector separation, submitted-value cleanup, semantic input registration, row removal, surface detachment, and runtime retirement. |

Predecessor acknowledgement advances only baseline fields that still match that
predecessor's original authored values. A subsequent explicit review establishes
its own baseline and is not undone by an older receipt.

Observation source readiness checks both independent grid and inspector drafts.
Separating their submission payloads does not make either unsaved source eligible
for a specialized observation action.
