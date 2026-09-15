# workbook/query/

[Parent](../README.md) · [Source overview](../../README.md)

Instance-owned surface queries, committed-record capabilities, latest-request admission, and live-row reconciliation.

Queries bind admission, live reconciliation, protected-state cleanup, and
teardown to the consuming workbook instance. Shared query-row contracts belong
here and are consumed by Timeline as well as other surfaces.

## Files

| File | Responsibility |
| --- | --- |
| [useAssessmentSurfaceQuery.ts](useAssessmentSurfaceQuery.ts) | Assessment-owned query construction, admission, live reconciliation, refresh, access cleanup, and cancellation. |
| [useEntitySurfaceQuery.ts](useEntitySurfaceQuery.ts) | Independent Hosts and Identities reads with Entity-owned conversion and partial indexing. |
| [useGenericSurfaceQuery.ts](useGenericSurfaceQuery.ts) | Schema-keyed generic query admission, schema matching, Notes normalization, live reconciliation, stale-row retention, access cleanup, and cancellation. |
| [WorkbookCommittedRecordPort.ts](WorkbookCommittedRecordPort.ts) | Semantic capabilities for authoritative committed records and complete accepted mutation receipts, independent of subsequent query reads. |
| [workbookLatestRequest.ts](workbookLatestRequest.ts) | Instance-local latest-request sequencing, supersession abort, and current-result admission. |
| [WorkbookQueryRow.ts](WorkbookQueryRow.ts) | Shared schema-keyed view-query row shape below Timeline ownership. |
| [workbookQueryRowPatch.ts](workbookQueryRowPatch.ts) | Pure sparse-patch application for shared query rows. |
| [WorkbookViewQueryPort.ts](WorkbookViewQueryPort.ts) | Shared abortable query capability returning correlated, contract-normalized rows. |
| [WorkbookQueryBrowser.ts](WorkbookQueryBrowser.ts) | Staged complete-result admission, bounded windows, request checkpoints, continuation and restart. |
| [WorkbookQueryBrowsingControls.tsx](WorkbookQueryBrowsingControls.tsx) | Explicit keyboard continuation, earlier checkpoints, cursor-free refresh, and local query Retry/Revert. |
| [WorkbookQueryBrowsingContext.tsx](WorkbookQueryBrowsingContext.tsx) | Workbook-scoped browsing lifetimes and accepted query presentation. |
| [workbookQueryMetadata.ts](workbookQueryMetadata.ts) | Canonical metadata validation and authored-intent preservation. |

## Tests

| File | Responsibility |
| --- | --- |
| [workbookQueryMetadata.test.ts](workbookQueryMetadata.test.ts) | Canonical query correlation, required paging, authored-sort preservation and continuation mismatch rejection. |
| [WorkbookQueryBrowser.test.ts](WorkbookQueryBrowser.test.ts) | Window bounds, opaque cursors, checkpoint exhaustion, atomic admission, supersession and recovery. |
| [ordinaryCreateQuery.test.tsx](ordinaryCreateQuery.test.tsx) | Ordinary accepted-version floors, filtered membership and concealment of obsolete reads. |
| [useAssessmentSurfaceQuery.test.tsx](useAssessmentSurfaceQuery.test.tsx) | Direct rapid-filter, stale-error, live-patch, access-loss, inactive-surface, and teardown characterization. |
| [useEntitySurfaceQuery.test.tsx](useEntitySurfaceQuery.test.tsx) | Independent-sheet reads, reference revalidation, stale-result, live-patch, access-loss, cleanup and teardown. |
| [useGenericSurfaceQuery.test.tsx](useGenericSurfaceQuery.test.tsx) | Direct Notes normalization, schema-switch, mismatched-envelope, stale-row, access-loss, inactive-surface, and teardown characterization. |
| [workbookLatestRequest.test.ts](workbookLatestRequest.test.ts) | Tests exclusive latest-request ownership and supersession aborts. |

## Browsing lifetime

Each active surface retains at most three pages of 100 records and twenty
request-only return checkpoints. Explicit refresh retires the old cursor chain,
including after a failed restart. Live invalidations coalesce and re-read only
the current window. Selection refers to loaded query members; drafts, inspectors
and mutation receipts remain independent.

The incident collaboration coordinator keeps a stable invalidation binding while
sheet readers change. Initial authorization recovery has its own effect; neither
reference-broker replacement nor a sheet query restarts that authority loop.
Grouping uses full-row `group_values`, with owner-supplied bucket comparison;
server row ordering remains unchanged within buckets.

Startup authorization may replace an in-flight reference broker. The Entity
reference reader completes that unfinished obligation through the replacement;
an already accepted reference set does not trigger another eager read.

Timeline mention rendering retains a named, bounded Host/Identity observation in useEntityReferenceRows. It reads the existing query port directly, separately from authored Entity sheet queries. It no longer depends on the ordinary reference broker. Specialized mention resolution continues to own its own candidate workflow.
