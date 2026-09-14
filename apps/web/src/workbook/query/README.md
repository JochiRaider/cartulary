# workbook/query/

[Parent](../README.md) · [Source overview](../../README.md)

Instance-owned surface queries, committed-record capabilities, latest-request admission, and live-row reconciliation.

Queries bind admission, live reconciliation, protected-state cleanup, and
teardown to the consuming workbook instance. Shared query-row contracts belong
here and are consumed by Timeline as well as other surfaces.

## Files

| File | Responsibility |
| --- | --- |
| [entityLiveEventPatchPlanner.ts](entityLiveEventPatchPlanner.ts) | Plans monotonic Entity query-row patches from decoded live events. |
| [useAssessmentSurfaceQuery.ts](useAssessmentSurfaceQuery.ts) | Assessment-owned query construction, admission, live reconciliation, refresh, access cleanup, and cancellation. |
| [useEntitySurfaceQuery.ts](useEntitySurfaceQuery.ts) | Instance-owned dual host/identity loading, indexing, live patching, refresh, cancellation, and protected-state cleanup. |
| [useGenericSurfaceQuery.ts](useGenericSurfaceQuery.ts) | Schema-keyed generic query admission, schema matching, Notes normalization, live reconciliation, stale-row retention, access cleanup, and cancellation. |
| [WorkbookCommittedRecordPort.ts](WorkbookCommittedRecordPort.ts) | Semantic capability for reading and observing authoritative committed workbook records. |
| [workbookLatestRequest.ts](workbookLatestRequest.ts) | Instance-local latest-request sequencing, supersession abort, and current-result admission. |
| [WorkbookQueryRow.ts](WorkbookQueryRow.ts) | Shared schema-keyed view-query row shape below Timeline ownership. |
| [workbookQueryRowPatch.ts](workbookQueryRowPatch.ts) | Pure sparse-patch application for shared query rows. |
| [WorkbookViewQueryPort.ts](WorkbookViewQueryPort.ts) | Shared abortable query capability returning correlated, contract-normalized rows. |

## Tests

| File | Responsibility |
| --- | --- |
| [ordinaryCreateQuery.test.tsx](ordinaryCreateQuery.test.tsx) | Ordinary accepted-version floors, filtered membership and concealment of obsolete reads. |
| [entityLiveEventPatchPlanner.test.ts](entityLiveEventPatchPlanner.test.ts) | Tests exact newer Entity patches and reference-preserving stale-event no-ops. |
| [useAssessmentSurfaceQuery.test.tsx](useAssessmentSurfaceQuery.test.tsx) | Direct rapid-filter, stale-error, live-patch, access-loss, inactive-surface, and teardown characterization. |
| [useEntitySurfaceQuery.test.tsx](useEntitySurfaceQuery.test.tsx) | Direct dual-load, stale-result, live-patch, access-loss, explicit cleanup, and teardown characterization. |
| [useGenericSurfaceQuery.test.tsx](useGenericSurfaceQuery.test.tsx) | Direct Notes normalization, schema-switch, mismatched-envelope, stale-row, access-loss, inactive-surface, and teardown characterization. |
| [workbookLatestRequest.test.ts](workbookLatestRequest.test.ts) | Tests exclusive latest-request ownership and supersession aborts. |
