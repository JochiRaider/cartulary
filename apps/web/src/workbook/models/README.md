# workbook/models/

`WorkbookGridDraftStore` retains exact ordinary grid authoring, explicit clear,
baseline and revision identity in the account/incident runtime. It is separate
from inspector drafts and captured mutation requests. `workbookGridEditValue`
constructs declared direct-value intent without rewriting raw authoring.

[Parent](../README.md) · [Source overview](../../README.md)

Pure workbook request, row, query, registry, startup, saved-view, and presentation models.

Models consume typed package facades and represent workbook-local decisions.
App workflow state remains in [app](../../app/README.md), operation lifetime in
[runtime](../runtime/README.md), and Timeline-specific models in
[Timeline models](../timeline/models/README.md).

## Record and mutation models

| File | Responsibility |
| --- | --- |
| [WorkbookGridDraftStore.ts](WorkbookGridDraftStore.ts) | Retained raw grid input, revision-specific validation, original baseline, explicit clear and authorization lifetime. |
| [WorkbookGridDraftStore.test.ts](WorkbookGridDraftStore.test.ts) | Exact input, stale completion, review and account/incident retirement evidence. |
| [workbookGridEditValue.ts](workbookGridEditValue.ts) | Declared grid-only scalar and reference validation without changing unfinished text. |
| [workbookGridEditValue.test.ts](workbookGridEditValue.test.ts) | Field-family input, explicit-null and non-editable-field dispositions. |
| [assessmentWorkbookModel.test.ts](assessmentWorkbookModel.test.ts) | Tests contract-backed Assessment defaults, creation payloads, confidence, and support models. |
| [assessmentWorkbookModel.ts](assessmentWorkbookModel.ts) | Assessment workbook draft, payload, confidence-band, and support-row helpers. |
| [entityClipboardPastePlan.test.ts](entityClipboardPastePlan.test.ts) | Tests Entity scalar routing, exact-origin all-create requests, and fail-closed target/authority handling. |
| [entityClipboardPastePlan.ts](entityClipboardPastePlan.ts) | Pure Entity scalar-versus-batch paste planning with current-record, writable-field, grouping, create-capability, and exact-surface admission. |
| [entityCellPresentation.ts](entityCellPresentation.ts) | Shared committed renderer labels and independent readable Find fragments for Hosts and Identities. |
| [entityCellPresentation.test.ts](entityCellPresentation.test.ts) | Exact fallbacks, collection fragment boundaries and metadata/placeholder exclusions for both schemas. |
| [entityIdentifierClasses.ts](entityIdentifierClasses.ts) | Entity identifier field classification used by merge review. |
| [entityMergePlan.test.ts](entityMergePlan.test.ts) | Tests owner-aligned identifier normalization, promotion ordering, duplicates, and alias policy. |
| [entityMergePlan.ts](entityMergePlan.ts) | Pure Entity merge identifier outcomes and review explanation. |
| [entityWorkbookModel.test.ts](entityWorkbookModel.test.ts) | Tests Entity identifier editability, row normalization, and workbook payload construction. |
| [entityWorkbookModel.ts](entityWorkbookModel.ts) | Host/identity entity row, merge-plan, grouping, and payload helpers. |
| [evidenceLifecycleViewModel.test.ts](evidenceLifecycleViewModel.test.ts) | Tests owner-valid Evidence lifecycle and object-blob upload combinations. |
| [evidenceLifecycleViewModel.ts](evidenceLifecycleViewModel.ts) | Evidence lifecycle display/count view-model helpers. |
| [genericWorkbookModel.test.ts](genericWorkbookModel.test.ts) | Tests generic creation/edit payloads, omission, explicit clears, and authoring minima. |
| [genericWorkbookModel.ts](genericWorkbookModel.ts) | Generic system-view create/edit payload, enum, validation, and row-label helpers. |
| [workbookClipboardPaste.test.ts](workbookClipboardPaste.test.ts) | Tests bounded paste columns/targets and exact generated paste-capable surface identities. |
| [workbookClipboardPaste.ts](workbookClipboardPaste.ts) | Exact generated paste-capable view parser plus bounded column/target constructors and semantic surface validation. |
| [workbookIncidentIdentity.ts](workbookIncidentIdentity.ts) | Incident identity normalization and loading-state model. |
| [workbookRequestDecoders.test.ts](workbookRequestDecoders.test.ts) | Tests exact creation, linked-note, collection, and patch request validation. |
| [workbookRequestDecoders.ts](workbookRequestDecoders.ts) | Fail-closed exact create, patch, linked-note, and collection-action request decoders with exhaustive generated action coverage. |

## Queries and surface registration

| File | Responsibility |
| --- | --- |
| [workbookColumnSizing.ts](workbookColumnSizing.ts) | Shared integer width validation downstream of the authored design projection. |
| [workbookContractRows.ts](workbookContractRows.ts) | Contract-backed row normalization and grid-column materialization helpers for workbook surfaces. |
| [workbookGridQueryControls.test.ts](workbookGridQueryControls.test.ts) | Tests ordered-sort lifecycle, duplicate/limit rejection, and reference-preserving no-ops. |
| [workbookGridQueryControls.ts](workbookGridQueryControls.ts) | Pure query-control projection, closure-free command descriptors, exact controlled-value parsers, ordered-sort commands, and surface-keyed transient reducer. |
| [workbookQuery.test.ts](workbookQuery.test.ts) | Tests declared query operators, argument shapes, sort, and grouping construction. |
| [workbookQuery.ts](workbookQuery.ts) | Workbook query, filter, sort, grouping, and request-building helpers. |
| [workbookSurfaceQueryRuntime.ts](workbookSurfaceQueryRuntime.ts) | Resolves view-schema contracts and maps them to the appropriate workbook query owner slot. |
| [workbookSurfaceRegistration.test.ts](workbookSurfaceRegistration.test.ts) | Tests for policy registration completeness, uniqueness, and extension-workspace exclusion. |
| [workbookSurfaceRegistration.ts](workbookSurfaceRegistration.ts) | Exact `view_schema_id` registration and bounded-context policy registry for all exposed workbook schemas. |
| [workbookSurfaceRegistry.test.ts](workbookSurfaceRegistry.test.ts) | Tests for workbook surface registry invariants. |
| [workbookSurfaceRegistry.ts](workbookSurfaceRegistry.ts) | Built-in/system/optional workbook surface registry and stable view-schema IDs. |

## Inspector and presentation models

| File | Responsibility |
| --- | --- |
| [workbookGridEntryFocus.ts](workbookGridEntryFocus.ts) | Generation-keyed grid-entry focus request and exact-acknowledgement model. |
| [workbookGridState.ts](workbookGridState.ts) | Contract-grid load-state presentation and incident-role interaction-mode helpers. |
| [workbookInspectorModel.test.ts](workbookInspectorModel.test.ts) | Tests immutable inspector configuration, default-closed state, subjects, and declared no-row behavior. |
| [workbookInspectorModel.ts](workbookInspectorModel.ts) | Pure inspector state machine for default-closed state, semantic subjects, active panels, no-row state, and invalidation generations. |
| [workbookRelationshipChip.ts](workbookRelationshipChip.ts) | Cross-surface relationship-chip presentation contract with no Timeline interpretation. |
| [workbookShellPresentation.ts](workbookShellPresentation.ts) | Pure account, active-system-surface, and Network Analysis presentation decisions. |
| [workbookViewBarWorkingSet.ts](workbookViewBarWorkingSet.ts) | View-bar control ordering, query-entry identities, panel state, and available action models. |

## Saved views and startup

| File | Responsibility |
| --- | --- |
| [workbookSavedViewControl.test.ts](workbookSavedViewControl.test.ts) | Tests saved-view resource states, active-surface filtering, and editable scope validation. |
| [workbookSavedViewControl.ts](workbookSavedViewControl.ts) | Saved-view control resource states, action feedback, and active-surface projections. |
| [workbookSavedViewPaginationMachine.test.ts](workbookSavedViewPaginationMachine.test.ts) | Tests saved-view page correlation, ordered accumulation, and terminal-page publication. |
| [workbookSavedViewPaginationMachine.ts](workbookSavedViewPaginationMachine.ts) | Pure saved-view pagination admission, page validation, accumulation, and completion. |
| [workbookSavedViewRuntime.test.ts](workbookSavedViewRuntime.test.ts) | Tests saved-view identity updates, display ordering, selection, and dirty-state helpers. |
| [workbookSavedViewRuntime.ts](workbookSavedViewRuntime.ts) | Saved-view runtime selection, dirty-state, and command helpers. |
| [workbookSavedViews.test.ts](workbookSavedViews.test.ts) | Tests saved-view normalization and persistence preserve saved-view and view-schema identities. |
| [workbookSavedViews.ts](workbookSavedViews.ts) | Saved-view resource normalization and payload helpers. |
| [workbookStartup.test.ts](workbookStartup.test.ts) | Tests startup preference/fallback ordering by stable sheet reference. |
| [workbookStartup.ts](workbookStartup.ts) | Workbook startup candidate, selected sheet reference, and fallback resolution helpers. |

## Ordinary creation seam

| File | Responsibility |
| --- | --- |
| [workbookAuthoringValues.ts](workbookAuthoringValues.ts) | Neutral declared writable-string and timestamp normalization, separate from raw authoring. |

[workbookSavedValues.ts](workbookSavedValues.ts) compares canonical saved field
values for independent draft review and explicit patch preparation.

[WorkbookLocalDraftStore.ts](WorkbookLocalDraftStore.ts) retains scalar editor values
for one source surface in a live Workbook runtime. It holds no mounted controls or
source-specific capture behavior; runtime retirement clears its values.
