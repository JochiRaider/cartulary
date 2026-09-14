# workbook/mutations/

[Parent](../README.md) · [Source overview](../../README.md)

Semantic mutation command assembly, secure action identity, write coordination, and operation outcomes.

Assembly owns logical-action identity and exact owner request construction.
Presentation receives named semantic command ports rather than a transaction-ID
provider or generic mutation transport. Ordinary creates use the retained
`features/ordinary` owner; superseded one-shot create methods are removed.

## Files

| File | Responsibility |
| --- | --- |
| [createWorkbookMutationCommandPorts.ts](createWorkbookMutationCommandPorts.ts) | Incident-scoped command assembly and exact owner request construction over the private secure-ID port. |
| [entityRecordWriteBoundary.ts](entityRecordWriteBoundary.ts) | Semantic coordination boundary for writes targeting Entity records. |
| [secureTransactionId.ts](secureTransactionId.ts) | Private browser assembly adapter for secure transaction-ID creation. |
| [workbookConflictResolutionAdapter.ts](workbookConflictResolutionAdapter.ts) | Executes reviewed workbook conflict resolutions through semantic mutation commands. |
| [workbookMutationAuthority.ts](workbookMutationAuthority.ts) | Actor, incident, and session authority contract for workbook mutations. |
| [workbookMutationCommandPorts.ts](workbookMutationCommandPorts.ts) | Named Timeline, generic, entity, assessment, Evidence, and coordination command-port contracts. |
| [workbookOperationOutcome.ts](workbookOperationOutcome.ts) | Shared semantic workbook operation outcomes and field-level failures. |

## Tests

| File | Responsibility |
| --- | --- |
| [createWorkbookMutationCommandPorts.test.ts](createWorkbookMutationCommandPorts.test.ts) | Exact owner payload, logical identity, and local secure-random failure tests. |
