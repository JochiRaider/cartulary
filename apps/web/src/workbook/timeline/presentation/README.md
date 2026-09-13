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
| [TimelineWorkbookViewBarRegion.tsx](TimelineWorkbookViewBarRegion.tsx) | Stateless query, saved-view, add-row, inspector-toggle, and bulk-action controls. |
| [useTimelineInspectorPresentation.tsx](useTimelineInspectorPresentation.tsx) | Derives Timeline inspector sections, action bindings, and feedback presentation. |
| [useTimelineWorkbookPresentation.tsx](useTimelineWorkbookPresentation.tsx) | Derives renderer, column, grid-row, load-state, inspector, status, view-bar, notice, and context-menu models. |
