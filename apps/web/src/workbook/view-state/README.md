# workbook/view-state/

[Parent](../README.md) · [Source overview](../../README.md)

Instance-local, schema-keyed workbook query state and defaults.

Query state and defaults belong to a workbook instance and exact schema.
There are no module-global mutable defaults or cross-shell stores.

## Files

| File | Responsibility |
| --- | --- |
| [useWorkbookQueryState.ts](useWorkbookQueryState.ts) | Instance-owned schema-keyed query defaults, reducer state, reset, and update operations. |
