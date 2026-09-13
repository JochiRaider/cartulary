# workbook/features/generic/

[Parent](../README.md) · [Source overview](../../../README.md)

Contract-backed generic surface creation commands and inspector composition.

Generic presentation consumes declared surface policies and semantic commands.
Source-specific workflows remain with their feature owners; this directory
assembles their bindings into the generic inspector.

## Files

| File | Responsibility |
| --- | --- |
| [createGenericMutationCommandPort.ts](createGenericMutationCommandPort.ts) | Builds semantic mutation commands for contract-backed generic workbook surfaces. |
| [genericCreateRequestBuilder.ts](genericCreateRequestBuilder.ts) | Constructs generic record-creation requests from surface policy and draft values. |
| [GenericWorkbookInspector.tsx](GenericWorkbookInspector.tsx) | Generic contract-surface inspector facade over composition and presentation. |
| [GenericWorkbookInspectorPresentation.tsx](GenericWorkbookInspectorPresentation.tsx) | Renders generic inspector sections from the feature composition's presentation model. |
| [useGenericCreateDraft.ts](useGenericCreateDraft.ts) | React state and commands for generic surface create drafts. |
| [useGenericWorkbookInspectorComposition.tsx](useGenericWorkbookInspectorComposition.tsx) | Assembles generic inspector fields, actions, relationships, and owner workflow bindings. |
