# shared/

[Source overview / parent](../README.md)

Cross-feature validation, public errors, semantic contracts, and interaction helpers.

These helpers serve multiple features. Workbook-specific coordination stays
in [workbook](../workbook/README.md). Overlay navigation uses registered semantic
element refs; it does not discover focus targets through selectors or timers.

## Files

| File | Responsibility |
| --- | --- |
| [auditReadValues.ts](auditReadValues.ts) | Shared audit filter fields, timestamp normalization, validation, and JSON display formatting. |
| [authorizationRecovery.ts](authorizationRecovery.ts) | Semantic authorization-recovery capability and outcomes shared by app and workbook owners. |
| [displayName.ts](displayName.ts) | Shared display-name validation. |
| [incidentResource.ts](incidentResource.ts) | Incident-resource validation and ordering of authoritative observations. |
| [publicError.ts](publicError.ts) | Shared public-error normalization helpers. |
| [publicErrorPresentation.ts](publicErrorPresentation.ts) | Maps structured public errors and operation families to safe presentation. |
| [sheetRef.ts](sheetRef.ts) | Workbook sheet-reference validation, stable keys, and semantic equality. |
| [useRegisteredOverlayNavigation.ts](useRegisteredOverlayNavigation.ts) | Overlay keyboard navigation and focus restoration through registered semantic element refs. |
| [useTransientMessageController.ts](useTransientMessageController.ts) | Transient-message visibility timer with interaction pauses and explicit dismissal. |
| [workbookShellContracts.ts](workbookShellContracts.ts) | Shared app/workbook shell contracts for account identity, application menu handoff, and incident-controls renderer props. |

## Tests

| File | Responsibility |
| --- | --- |
| [publicError.test.ts](publicError.test.ts) | Direct tests for allowlisted public detail projection and unsafe-message sanitization. |
| [publicErrorPresentation.test.ts](publicErrorPresentation.test.ts) | Tests structured error presentation by operation context without inspecting human messages. |
| [useRegisteredOverlayNavigation.test.tsx](useRegisteredOverlayNavigation.test.tsx) | Tests registered overlay focus reconciliation, restoration cancellation, and removed items. |
| [useTransientMessageController.test.tsx](useTransientMessageController.test.tsx) | Tests visible-time dismissal, hover/focus pausing, and transient-message lifetime. |
