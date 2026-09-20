# workbook/layout/

[Parent](../README.md) · [Source overview](../../README.md)

Workbook density, responsive layout, column geometry, surface sizing, and shared style slots.

Surface owners consume semantic layout snapshots and commands. Viewport and
column geometry belong here rather than being inferred from row counts or
repeated viewport subtraction in surface components.

## Files

| File | Responsibility |
| --- | --- |
| [useWorkbookColumnLayoutController.ts](useWorkbookColumnLayoutController.ts) | React subscription and active-schema commands for the sole working-layout owner. |
| [WorkbookColumnLayoutController.ts](WorkbookColumnLayoutController.ts) | Sole working-layout store, sparse width and semantic freeze commands, mounted measurement capabilities and cancellation across configuration replacement. |
| [useWorkbookColumnSizingBinding.ts](useWorkbookColumnSizingBinding.ts) | Supplies source defaults and the mounted neutral GridHandle sizing capability; never owns widths or vendor nodes. |
| [useWorkbookLayoutFacade.ts](useWorkbookLayoutFacade.ts) | Composition facade for effective density, responsive mode, interaction mode, column state, and surface layout commands. |
| [useWorkbookResponsiveLayout.ts](useWorkbookResponsiveLayout.ts) | Viewport subscription and semantic responsive-layout snapshot. |
| [workbookColumnLayout.ts](workbookColumnLayout.ts) | Contract-normalized column ordering, visibility, width, semantic frozen prefix, movement, and materialization helpers. |
| [workbookDensity.ts](workbookDensity.ts) | Account density preference resolution. |
| [workbookResponsiveLayout.ts](workbookResponsiveLayout.ts) | Responsive layout classification and surface-band helpers. |
| [workbookShellStyles.ts](workbookShellStyles.ts) | Shared shell chrome, work-area, viewport-overlay, and responsive style slots. |
| [Shared work-area overlay](../../shared/WorkbookWorkAreaOverlay.tsx) | Workbook and Network Analysis recovery host, bounds and internal scrolling. |
| [WorkbookSurfaceLayout.tsx](WorkbookSurfaceLayout.tsx) | Shared work-area, bounded contextual feedback, independently scrolling grid/inspector slots, overlay geometry, resize behavior, and focus restoration. |

## Tests

`WorkbookColumnLayoutController.test.ts` covers portable bounds, sparse
restoration, numeric validation and obsolete measurement rejection.

| File | Responsibility |
| --- | --- |
| [useWorkbookResponsiveLayout.test.tsx](useWorkbookResponsiveLayout.test.tsx) | Tests viewport fallbacks and effective root width under zoom. |
| [workbookDensity.test.ts](workbookDensity.test.ts) | Tests for effective Workbook density. |
| [workbookLayoutPolicy.test.ts](workbookLayoutPolicy.test.ts) | Architecture checks prohibiting viewport-subtraction, row-count geometry, synthetic rows, and surface-private minimum-height workarounds. |
| [workbookResponsiveLayout.test.ts](workbookResponsiveLayout.test.ts) | Tests for responsive classification and query-control capacity. |

A new Workbook surface builds all semantic columns with its declared defaults,
marks committed presentation eligibility, and calls `useWorkbookColumnSizingBinding`
with its existing GridHandle ref and layout commands. Apply the working layout
through `applyWorkbookLayoutToColumns` and send header intents to
`onColumnSizingIntent`. The view bar consumes `commands.sizing`; it does not
import vendor coordinates or retain a second width map.


Freezing is authored only as a nullable boundary field in the complete semantic
order. All shared Core-grid surfaces project its visible prefix to Grid Adapter.
The existing mounted capability binding also subscribes to effective placement;
viewport suspension is never written back into authored state or saved-view dirty
comparison. Reset Columns clears the boundary; saved-view Reset restores it.
