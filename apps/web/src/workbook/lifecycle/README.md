# workbook/lifecycle/

[Parent](../README.md) · [Source overview](../../README.md)

Shared typed invalidation reasons for workbook, query, mutation, and extension lifetimes.

This leaf defines reason unions only. It contains no mutable state, event bus,
or generic clear-all operation.

## Files

| File | Responsibility |
| --- | --- |
| [workbookInvalidation.ts](workbookInvalidation.ts) | Shared typed Workbook and extension invalidation reason unions for applicable lifecycle owners. |
