# workbook/features/

[Parent](../README.md) · [Source overview](../../README.md)

Workbook feature entry points and owner-specific authoring, inspector, and recovery workflows.

The root contains Import Assistant presentation and the workbook entry points
for Network Analysis. Child directories own source-specific workflows consumed
by the shell and inspectors. Extension workspace identities remain outside the
Base surface registry.

## Subdirectories

| Directory | Responsibility |
| --- | --- |
| [assessments/](assessments/README.md) | Assessment subject/support discovery, append authoring, inspector composition, and recovery. |
| [coordination/](coordination/README.md) | Task lifecycle, Decision supersession, and contextual or coordination record-creation workflows. |
| [entities/](entities/README.md) | Hosts and Identities inspector composition, Entity merge ownership, and clipboard paste. |
| [evidence/](evidence/README.md) | Evidence access and attachment workflows, including staged Timeline-related Evidence creation and linking. |
| [generic/](generic/README.md) | Contract-backed generic surface creation commands and inspector composition. |
| [indicators/](indicators/README.md) | Indicator creation and lifecycle, Observation capture/resolution, and operation reconciliation. |
| [notes/](notes/README.md) | Source-bound and Notes-sheet authoring with retained atomic Note creation and recovery. |
| [parties/](parties/README.md) | Party candidate discovery, independent reference-pair edits, and staged creation/linking recovery. |

## Files

| File | Responsibility |
| --- | --- |
| [ImportAssistantFeature.tsx](ImportAssistantFeature.tsx) | Availability-gated workbook import assistant for discovery, ordinal mapping, approval, unit selection, apply/cancel, and result navigation. |
| [ImportOperationNotice.tsx](ImportOperationNotice.tsx) | Import operation status, recovery actions, and shared assistant layout styles. |
| [ImportUnitCard.tsx](ImportUnitCard.tsx) | One discovered import unit's mapping, approval, selection, and outcome controls. |
| [NetworkFlowFeature.tsx](NetworkFlowFeature.tsx) | Lazy Network Analysis presentation entrypoint. |
| [NetworkFlowOperations.ts](NetworkFlowOperations.ts) | Workbook/app-facing persistent Network Flow operation ownership. |

## Tests

| File | Responsibility |
| --- | --- |
| [ImportAssistantFeature.test.tsx](ImportAssistantFeature.test.tsx) | Import assistant discovery, approval, cancellation, and returned-selection characterization. |
