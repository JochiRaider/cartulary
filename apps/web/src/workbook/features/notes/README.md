# workbook/features/notes/

[Parent](../README.md) · [Source overview](../../../README.md)

Source-bound and Notes-sheet authoring with retained atomic Note creation and recovery.

Inspector source attachment and Notes-sheet authoring use the same operation
owner without silently retargeting retained contextual drafts. Transport and
source discovery use [workbook adapters](../../adapters/README.md).

## Files

| File | Responsibility |
| --- | --- |
| [NoteCreateContext.ts](NoteCreateContext.ts) | React context and sheet attachment for retained Note creation ownership. |
| [NoteCreateForm.tsx](NoteCreateForm.tsx) | Note authoring form with source context and reviewed submission. |
| [noteCreateModel.ts](noteCreateModel.ts) | Note source identities, drafts, read capabilities, validation, and request preparation. |
| [noteCreateOperation.ts](noteCreateOperation.ts) | Note creation reviews, immutable attempts, receipts, outcomes, and owner entries. |
| [NoteCreateRecovery.tsx](NoteCreateRecovery.tsx) | Retained Note creation replay and reconciliation recovery controls. |
| [NoteSheetAuthoring.tsx](NoteSheetAuthoring.tsx) | Notes sheet authoring bound to the Note creation owner. |
| [NoteSourceControl.tsx](NoteSourceControl.tsx) | Note source selection and source-context presentation. |
| [useNoteCreateAttachment.ts](useNoteCreateAttachment.ts) | Attaches Note presentation to retained source-bound creation state. |
| [WorkbookNoteCreateOwner.ts](WorkbookNoteCreateOwner.ts) | Source-bound Note drafts, atomic creation attempts, replay, and reconciliation ownership. |

## Tests

| File | Responsibility |
| --- | --- |
| [noteCreateAuthoring.test.tsx](noteCreateAuthoring.test.tsx) | Tests contextual Note sources, retained inspector drafts, and independent sheet authoring. |
| [noteCreateRecovery.test.tsx](noteCreateRecovery.test.tsx) | Tests Note receipt integrity, newer source evidence, and renewed review after authority restoration. |
