# workbook/continuity/

[Parent](../README.md) · [Source overview](../../README.md)

Semantic grid focus, selection, and viewport continuity through workbook-private adapter bindings.

Public continuity contracts use stable schema, record, and field identities.
DOM nodes, grid coordinates, viewport geometry, and captured restore state stay
inside private continuity/grid-adapter implementations.

## Files

| File | Responsibility |
| --- | --- |
| [gridViewportContinuity.ts](gridViewportContinuity.ts) | Private viewport anchor capture, restoration, and rectangle visibility helpers used by Timeline continuity. |
| [useWorkbookGridContinuity.tsx](useWorkbookGridContinuity.tsx) | Workbook-private translation between the semantic continuity port and grid-adapter focus, scrolling, clipboard, and viewport behavior. |
| [workbookContinuityPort.ts](workbookContinuityPort.ts) | Semantic capture, focus, selection, clear, one-shot restore, and disposal contract over stable schema, record, and field identities. |

## Tests

| File | Responsibility |
| --- | --- |
| [gridViewportContinuity.test.ts](gridViewportContinuity.test.ts) | Tests for viewport capture, restoration, and visibility helpers. |
| [workbookContinuityPort.test.ts](workbookContinuityPort.test.ts) | Tests for opaque capture tokens, stable semantic identities, one-shot restoration, and idempotent cleanup. |
