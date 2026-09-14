# workbook/layout/

[Parent](../README.md) · [Source overview](../../README.md)

Workbook density, responsive layout, column geometry, surface sizing, and shared style slots.

Surface owners consume semantic layout snapshots and commands. Viewport and
column geometry belong here rather than being inferred from row counts or
repeated viewport subtraction in surface components.

## Files

| File | Responsibility |
| --- | --- |
| [useWorkbookColumnLayoutController.ts](useWorkbookColumnLayoutController.ts) | Per-schema column state, active column commands, saved-layout application, and startup resets. |
| [useWorkbookLayoutFacade.ts](useWorkbookLayoutFacade.ts) | Composition facade for effective density, responsive mode, interaction mode, column state, and surface layout commands. |
| [useWorkbookResponsiveLayout.ts](useWorkbookResponsiveLayout.ts) | Viewport subscription and semantic responsive-layout snapshot. |
| [workbookColumnLayout.ts](workbookColumnLayout.ts) | Contract-normalized column ordering, visibility, width, movement, and materialization helpers. |
| [workbookDensity.ts](workbookDensity.ts) | Account density preference resolution. |
| [workbookResponsiveLayout.ts](workbookResponsiveLayout.ts) | Responsive layout classification and surface-band helpers. |
| [workbookShellStyles.ts](workbookShellStyles.ts) | Shared shell chrome, work-area, viewport-overlay, and responsive style slots. |
| [WorkbookWorkAreaOverlay.tsx](WorkbookWorkAreaOverlay.tsx) | Shared recovery host, containing bounds, entry focus, stacking, and internal scrolling for Coordination and Notes. |
| [WorkbookSurfaceLayout.tsx](WorkbookSurfaceLayout.tsx) | Shared work-area, independently scrolling grid/inspector slots, overlay geometry, resize behavior, and focus restoration. |

## Tests

| File | Responsibility |
| --- | --- |
| [useWorkbookResponsiveLayout.test.tsx](useWorkbookResponsiveLayout.test.tsx) | Tests viewport fallbacks and effective root width under zoom. |
| [workbookDensity.test.ts](workbookDensity.test.ts) | Tests for effective Workbook density. |
| [workbookLayoutPolicy.test.ts](workbookLayoutPolicy.test.ts) | Architecture checks prohibiting viewport-subtraction, row-count geometry, synthetic rows, and surface-private minimum-height workarounds. |
| [workbookResponsiveLayout.test.ts](workbookResponsiveLayout.test.ts) | Tests for responsive classification and query-control capacity. |
