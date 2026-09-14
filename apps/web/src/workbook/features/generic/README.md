# workbook/features/generic/

[Parent](../README.md) · [Source overview](../../../README.md)

Contract-backed generic surface patch commands and inspector composition.

Generic presentation consumes declared surface policies and semantic commands.
Source-specific workflows remain with their feature owners; this directory
assembles their bindings into the generic inspector.

## Files

| File | Responsibility |
| --- | --- |
| [createGenericMutationCommandPort.ts](createGenericMutationCommandPort.ts) | Builds existing-record patch commands; ordinary creation belongs to the retained owner. |
| [genericCreateRequestBuilder.ts](genericCreateRequestBuilder.ts) | Pure request helper retained for Party, Indicator observation and Timeline related creation consumers. |
| [GenericWorkbookInspector.tsx](GenericWorkbookInspector.tsx) | Generic contract-surface inspector facade over composition and presentation. |
| [GenericWorkbookInspectorPresentation.tsx](GenericWorkbookInspectorPresentation.tsx) | Renders generic inspector sections from the feature composition's presentation model. |
| [useGenericCreateDraft.ts](useGenericCreateDraft.ts) | Borrowed ordinary and Note authoring; no component-owned create draft. |
| [useGenericWorkbookInspectorComposition.tsx](useGenericWorkbookInspectorComposition.tsx) | Assembles generic inspector fields, actions, relationships, and owner workflow bindings. |
