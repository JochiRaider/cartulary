# workbook/features/parties/

[Parent](../README.md) · [Source overview](../../../README.md)

Party candidate discovery, independent reference-pair edits, and staged creation/linking recovery.

Party creation and source reference changes are separate operations with
independent receipts. Candidate discovery preserves Party identity through
paging; recovery can finish linking without repeating accepted creation.

## Files

| File | Responsibility |
| --- | --- |
| [PartyLinkControls.tsx](PartyLinkControls.tsx) | Party pair selection, independent clearing, and new-Party authoring controls. |
| [partyLinkModel.ts](partyLinkModel.ts) | Supported Party reference pairs, actions, reviewed values, and source patch preparation. |
| [PartyLinkPanel.tsx](PartyLinkPanel.tsx) | Party linking panel composing candidate selection and source-pair actions. |
| [PartyLinkRecovery.tsx](PartyLinkRecovery.tsx) | Party creation and source-patch status with independent recovery actions. |
| [partyLinkStyles.ts](partyLinkStyles.ts) | Shared Party linking input and layout styles. |
| [useGenericPartyLinkWorkflow.ts](useGenericPartyLinkWorkflow.ts) | Binds generic workbook inspectors to Party linking ownership and commands. |
| [usePartyCandidates.ts](usePartyCandidates.ts) | Paged Party candidate discovery with retained results and explicit retry. |
| [WorkbookPartyLinkOperationOwner.ts](WorkbookPartyLinkOperationOwner.ts) | Coordinates Party creation and source linking as separately captured and recoverable operations. |

## Tests

| File | Responsibility |
| --- | --- |
| [partyLinkControls.test.tsx](partyLinkControls.test.tsx) | Tests independent pair edits, deliberate Party authoring, and candidate-page retry. |
| [WorkbookPartyLinkOperationOwner.test.ts](WorkbookPartyLinkOperationOwner.test.ts) | Tests Party/source receipt retention, coordinated admission, and link-only recovery after creation. |
