# workbook/timeline/presentation/

[Parent](../README.md) · [Source overview](../../../README.md)

Timeline presentation derivation and stateless surface regions.

Presentation derives render models from the composition projection plus narrow
entity, layout, Indicator, and query-control inputs. It receives no mutation or
pending runtime, collaboration coordinator, authorization-recovery callback,
generated protocol, or browser service. The presentation hook owns visible-column
synchronization; the view and regions render statelessly.

## Files

| File | Responsibility |
| --- | --- |
| [TimelineWorkbookInspectorRegion.tsx](TimelineWorkbookInspectorRegion.tsx) | Stateless Timeline inspector and Indicator supplement region. |
| [TimelineWorkbookOverlayRegion.tsx](TimelineWorkbookOverlayRegion.tsx) | Stateless notices and row-context-menu overlay region. |
| [TimelineWorkbookView.tsx](TimelineWorkbookView.tsx) | Stateless Timeline surface layout and region assembly. |
| [TimelineWorkbookViewBarRegion.tsx](TimelineWorkbookViewBarRegion.tsx) | Stateless query, saved-view, Find/Clear, add-row and inspector-toggle controls; tag authoring lives in the work-area feedback leaf. |
| [useTimelineInspectorPresentation.tsx](useTimelineInspectorPresentation.tsx) | Derives Timeline inspector sections, action bindings, and feedback presentation. |
| [useTimelineWorkbookPresentation.tsx](useTimelineWorkbookPresentation.tsx) | Derives renderer, column, grid-row, load-state, inspector, status, view-bar, notice, and context-menu models. |

`useTimelineWorkbookPresentation` attaches a small subscription for unavailable
committed grid drafts. It reads exact source-owned registry values for local copy
or discard and conceals them with current authority. Inline review compares the
retained authoring baseline with the latest committed row, independently of the
vendor editor's captured render row.

Timeline Find is projected by the root composition. The view bar places its
control after query chips and before Inspector, while the shared control renders
the floating panel without another toolbar row. Grid registration binds the
Adapter presentation port; cell states add the orthogonal Find cue. These regions
do not execute queries, retain drafts or own search state.
