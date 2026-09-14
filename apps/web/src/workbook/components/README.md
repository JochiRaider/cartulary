# workbook/components/

[Parent](../README.md) · [Source overview](../../README.md)

Shared workbook surface facades, shell chrome, query controls, field editors, and recovery presentation.

Presentation receives behavior through props, local models, and semantic command
ports. Feature workflow ownership lives in [features](../features/README.md),
mutation state in [runtime](../runtime/README.md), and geometry in
[layout](../layout/README.md).

## Query and view controls

| File | Responsibility |
| --- | --- |
| [ActiveSurfaceSavedViewSelector.test.tsx](ActiveSurfaceSavedViewSelector.test.tsx) | Tests saved-view dialog focus restoration and late confirmation after surface changes. |
| [ActiveSurfaceSavedViewSelector.tsx](ActiveSurfaceSavedViewSelector.tsx) | Saved-view selector for the active workbook surface. |
| [SavedViewActionPanel.tsx](SavedViewActionPanel.tsx) | Saved-view create/update/delete action authoring and review controls. |
| [WorkbookActiveQueryChips.tsx](WorkbookActiveQueryChips.tsx) | Responsive canonical group/sort/filter chip presentation over semantic command descriptors. |
| [WorkbookColumnsControl.tsx](WorkbookColumnsControl.tsx) | Registered-focus semantic column visibility, ordering, and reset menu. |
| [WorkbookFiltersControl.tsx](WorkbookFiltersControl.tsx) | Registered-focus filter draft dialog with exact field/value parsing and invalid-draft feedback. |
| [WorkbookGridControls.test.tsx](WorkbookGridControls.test.tsx) | Tests filter-chip editing and roving keyboard navigation in active query controls. |
| [WorkbookGridControls.tsx](WorkbookGridControls.tsx) | Active-surface query-control composition over one transient reducer and semantic command port. |
| [workbookGridControlStyles.ts](workbookGridControlStyles.ts) | Shared workbook query/menu control frames, labels, inputs, and styles. |
| [WorkbookGroupControl.tsx](WorkbookGroupControl.tsx) | Exact contract-declared grouping selector. |
| [WorkbookShellViewBarControls.tsx](WorkbookShellViewBarControls.tsx) | Saved-view and query-control projections over the shell runtime's narrow snapshot and command boundary. |
| [WorkbookSortControl.tsx](WorkbookSortControl.tsx) | Complete ordered-sort add, direction, priority, removal, and limit menu. |
| [WorkbookViewBar.tsx](WorkbookViewBar.tsx) | Shared saved-view, query, inspector, and create control composition. |

## Surface and shell presentation

| File | Responsibility |
| --- | --- |
| [AssessmentWorkbookSurface.tsx](AssessmentWorkbookSurface.tsx) | Assessment presentation facade for rows, support selection, and semantic assessment commands. |
| [EntityWorkbookSurface.tsx](EntityWorkbookSurface.tsx) | Hosts and Identities presentation over entity query/edit/merge/continuity owners and borrowed retained ordinary creation. |
| [GenericWorkbookSurface.tsx](GenericWorkbookSurface.tsx) | Contract-backed grid presentation borrowing retained ordinary drafts, local recovery and source command ports. |
| [IncidentControlsDrawer.tsx](IncidentControlsDrawer.tsx) | Shell-level incident controls drawer presentation and focus boundary. |
| [SystemViewSwitcher.tsx](SystemViewSwitcher.tsx) | System-view switcher UI and grouped surface navigation. |
| [WorkbookActiveSurfaceFrame.tsx](WorkbookActiveSurfaceFrame.tsx) | Active-surface recovery boundary and mutually exclusive blocked/overflow/conflict presentation, preserving rejected editors' accessibility. |
| [WorkbookActiveSurfaceFrame.test.tsx](WorkbookActiveSurfaceFrame.test.tsx) | Keeps the original draft editor focusable and accessible as conflict recovery opens and closes. |
| [WorkbookActiveSurfacePresentation.tsx](WorkbookActiveSurfacePresentation.tsx) | Exact built-in or extension renderer selection with lazy extension lifecycle binding. |
| [WorkbookIncidentControlsPresentation.tsx](WorkbookIncidentControlsPresentation.tsx) | Incident-controls drawer content and lazy Import Assistant renderer selection. |
| [WorkbookPresenceMarkers.tsx](WorkbookPresenceMarkers.tsx) | Shared row-gutter and cell presence markers with design-owned capacity and overflow behavior. |
| [WorkbookShellSlots.tsx](WorkbookShellSlots.tsx) | Stable shell slot IDs, labels, and layout slot helpers. |
| [WorkbookShellTopBar.tsx](WorkbookShellTopBar.tsx) | Responsive built-in/system-surface navigation, registered menu focus, incident identity, presence, and account presentation. |

## Editing and reference controls

| File | Responsibility |
| --- | --- |
| [GenericMutationControl.test.tsx](GenericMutationControl.test.tsx) | Tests field-control descriptors, variants, options, sizing, and form/grid presentation. |
| [GenericMutationControl.tsx](GenericMutationControl.tsx) | Renders descriptor-driven field editors for generic form and grid mutation surfaces. |
| [genericMutationControlModel.ts](genericMutationControlModel.ts) | Pure field-control descriptors for generic form and grid mutation editors. |
| [WorkbookAuthoringReferenceControl.tsx](WorkbookAuthoringReferenceControl.tsx) | Shared staged/paged reference picker, with compact grid popover geometry and keyboard exit; selection identities belong to the authoring owner. |
| [WorkbookGridEditorControl.tsx](WorkbookGridEditorControl.tsx) | Contract-field grid editor adapter, mutation controls, commit/cancel behavior, and editor-kind selection. |
| [WorkbookRecordCandidatePicker.tsx](WorkbookRecordCandidatePicker.tsx) | Shared semantic record-candidate selection control for owner workflows. |
| [WorkbookRelationshipChip.test.tsx](WorkbookRelationshipChip.test.tsx) | Tests relationship-chip state details, semantic selectors, and optional selection behavior. |
| [WorkbookRelationshipChip.tsx](WorkbookRelationshipChip.tsx) | Shared relationship-chip presentation over an explicit label, state, detail, selector identity, selection, and command model. |

## Status and recovery

| File | Responsibility |
| --- | --- |
| [RecoverySurface.tsx](RecoverySurface.tsx) | Shared presentation frame for retained-operation recovery content. |
| [SavedViewRecovery.tsx](SavedViewRecovery.tsx) | Saved-view operation recovery presentation with stable recovery identities. |
| [WorkbookEditRecoveryPanel.tsx](WorkbookEditRecoveryPanel.tsx) | Workbook edit recovery panel for retained pending work and explicit recovery actions. |
| [WorkbookQueueOverflowNotice.tsx](WorkbookQueueOverflowNotice.tsx) | Pending mutation queue overflow feedback and recovery affordances. |
| [WorkbookSameFieldConflictResolver.tsx](WorkbookSameFieldConflictResolver.tsx) | Same-field conflict comparison and explicit resolution controls. |
| [WorkbookSaveAnnouncements.tsx](WorkbookSaveAnnouncements.tsx) | Accessible announcements derived from workbook save and recovery state. |
| [WorkbookStatusStrip.tsx](WorkbookStatusStrip.tsx) | Status strip presentation for save/load/selection state. |

## Batch recovery

| File | Responsibility |
| --- | --- |
| [WorkbookBatchRecovery.tsx](WorkbookBatchRecovery.tsx) | Compact non-color batch status, original-input recovery, keyboard Retry, reads-only Retry refresh and activation of retained per-cell conflict groups. |

The presentation subscribes to retained runtime state. Opening batch recovery
closes conflict presentation; reviewing conflicts activates the existing resolver
and focus owner. Routine acceptance opens no panel.
