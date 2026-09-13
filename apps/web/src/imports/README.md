# imports/

[Source overview / parent](../README.md)

Workbook Import workflow ownership, captured requests, mapping, and job observation.

Transport and generated protocol aliases live in
[services](../services/README.md). Mapping interpretation stays with the
consuming feature. The [app bindings](../app/README.md) retain workflow lifetime,
and the [Import Assistant](../workbook/features/README.md) supplies presentation.

## Files

| File | Responsibility |
| --- | --- |
| [importCoordinator.ts](importCoordinator.ts) | Paginated discovery, preview, mapping approval, selection, apply, polling, and cancellation orchestration over validated Import/job operations. |
| [importRequests.ts](importRequests.ts) | Captures immutable upload and write attempts with scope and transaction identity for replay. |
| [WorkbookImportController.ts](WorkbookImportController.ts) | Workbook import workflow owner for upload, mapping, approval, selection, apply, and recovery. |
| [workbookImportMapping.ts](workbookImportMapping.ts) | Workbook mapping drafts, advisory suggestions, validation, request construction, and clipboard rectangles. |
| [workbookImportState.ts](workbookImportState.ts) | Workbook import session, unit, and operation state with apply blockers and outcome projections. |

## Tests

| File | Responsibility |
| --- | --- |
| [importCoordinator.test.ts](importCoordinator.test.ts) | Characterization for opaque cursor traversal, exact envelopes, preview/approval separation, stale fingerprints, CSRF, and public errors. |
| [importJobObservation.test.ts](importJobObservation.test.ts) | Tests bounded import job reads, explicit observation recovery, aborts, and regression rejection. |
| [WorkbookImportController.test.ts](WorkbookImportController.test.ts) | Tests import-stage authority fencing, advisory mapping suggestions, and duplicate destination errors. |
