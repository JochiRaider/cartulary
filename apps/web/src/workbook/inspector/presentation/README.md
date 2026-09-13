# workbook/inspector/presentation/

[Parent](../README.md) · [Source overview](../../../README.md)

Shared inspector shells, panels, actions, feedback, and History event rendering.

These components render declared subjects, panels, action bindings, and safe
feedback. They receive feature commands and presentation models rather than
owning source mutations or request lifetimes.

## Files

| File | Responsibility |
| --- | --- |
| [WorkbookHistoryPresentation.tsx](WorkbookHistoryPresentation.tsx) | Shared History list and event rendering from presentation models. |
| [WorkbookInspectorActions.tsx](WorkbookInspectorActions.tsx) | Inspector action groups, contextual actions, and accessible action buttons. |
| [WorkbookInspectorFeedback.tsx](WorkbookInspectorFeedback.tsx) | Inspector metadata, technical details, safe errors, feedback, and confirmation presentation. |
| [workbookInspectorPresentationModel.ts](workbookInspectorPresentationModel.ts) | Inspector action bindings, disabled reasons, technical fields, and History event presentation types. |
| [WorkbookInspectorShell.tsx](WorkbookInspectorShell.tsx) | Shared inspector shell and declared panel-section layout. |

## Tests

| File | Responsibility |
| --- | --- |
| [WorkbookInspectorPresentation.test.tsx](WorkbookInspectorPresentation.test.tsx) | Tests valid subject boundaries, ordered panels, deleted-record History, and explicit creation content. |
