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

The shell keeps its compact title and Close control outside the single
`data-inspector-scroll-body`. Full record context and technical metadata belong
in that body. Semantic focus targets scroll within it; outer slot geometry,
separator behavior and focus restoration retain their layout owners.

`WorkbookInspectorPanelContent.tsx` renders explicit data/access states without
owning requests or authoring. Contextual action groups share descriptions by
reason identity and parameters, never by wording. Feature owners retain closed
cause vocabularies; shared rendering knows only their presentation contributions.
Buttons use authored component tokens. Form controls share
`../../components/workbookFormStyles.ts`; grid-specific sizing remains separate.
