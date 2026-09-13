# workbook/startup/

[Parent](../README.md) · [Source overview](../../README.md)

Workbook startup admission and commit planning across base surfaces, saved views, and extension availability.

Admission owns request ordering, availability reservation, saved-view hydration
precedence, and stale-result rejection. Selection and availability are supplied
through injected capabilities.

## Files

| File | Responsibility |
| --- | --- |
| [useWorkbookStartupAdmission.ts](useWorkbookStartupAdmission.ts) | Startup request and response admission boundary with selection-version, incident, ordinal, availability, and cancellation guards. |
| [workbookStartupAdmissionMachine.ts](workbookStartupAdmissionMachine.ts) | Pure workbook startup admission, cancellation, fallback, and commit planning. |
| [WorkbookStartupPort.ts](WorkbookStartupPort.ts) | Semantic startup query inputs and correlated startup selection/availability outcomes. |

## Tests

| File | Responsibility |
| --- | --- |
| [useWorkbookStartupAdmission.test.tsx](useWorkbookStartupAdmission.test.tsx) | Startup admission characterization for precedence, overlap, extension availability, fallbacks, teardown, and late work. |
| [workbookStartupAdmissionMachine.test.ts](workbookStartupAdmissionMachine.test.ts) | Tests startup identity and exact base, saved-view, extension, and availability commit plans. |
