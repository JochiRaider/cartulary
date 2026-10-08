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
| [WorkbookBooleanFilterOperand.tsx](WorkbookBooleanFilterOperand.tsx) | Shared native boolean select and checkbox choices over FilterDraft; no query or retained state owner. |
| [WorkbookEnumFilterOperand.tsx](WorkbookEnumFilterOperand.tsx) | Shared native enum equality choices and explicit literal editor over FilterDraft; local disclosure only. |
| [ActiveSurfaceSavedViewSelector.test.tsx](ActiveSurfaceSavedViewSelector.test.tsx) | Tests saved-view dialog focus restoration and late confirmation after surface changes. |
| [ActiveSurfaceSavedViewSelector.tsx](ActiveSurfaceSavedViewSelector.tsx) | Saved-view selector for the active workbook surface. |
| [SavedViewBrowser.tsx](SavedViewBrowser.tsx) | Bounded saved-view discovery, explicit resource activation and keyboard/focus behavior in a compact popover. |
| [SavedViewActionPanel.tsx](SavedViewActionPanel.tsx) | Saved-view create/update/delete action authoring and review controls. |
| [WorkbookActiveQueryChips.tsx](WorkbookActiveQueryChips.tsx) | Responsive canonical group/sort/filter chip presentation over semantic command descriptors. |
| [WorkbookColumnsControl.tsx](WorkbookColumnsControl.tsx) | Registered-focus semantic column visibility, ordering, and reset menu. |
| [WorkbookFiltersControl.tsx](WorkbookFiltersControl.tsx) | Registered-focus filter draft dialog with exact field/value parsing and invalid-draft feedback. |
| [WorkbookGridControls.test.tsx](WorkbookGridControls.test.tsx) | Tests filter-chip editing and roving keyboard navigation in active query controls. |
| [WorkbookGridControls.tsx](WorkbookGridControls.tsx) | Active-surface query-control composition over one transient reducer and semantic command port. |
| [workbookGridControlStyles.ts](workbookGridControlStyles.ts) | Shared workbook query/menu control frames, labels, inputs, and styles. |
| [WorkbookGroupControl.tsx](WorkbookGroupControl.tsx) | Exact contract-declared grouping selector. |
| [WorkbookShellViewBarControls.tsx](WorkbookShellViewBarControls.tsx) | Stateless query-to-view-bar binding shared by the shell and Timeline fixture; shell-only Saved View, preset, subject and visibility composition. |
| [WorkbookSortControl.tsx](WorkbookSortControl.tsx) | Complete ordered-sort add, direction, priority, removal, and limit menu. |
| [WorkbookViewBar.tsx](WorkbookViewBar.tsx) | Shared saved-view, query, inspector, and create control composition. |

Query and View options composition keeps one command/draft owner. Panels use
native manual top-layer popovers and the existing viewport-placement port with
an explicit live anchor. This keeps nested editors outside parent scroll clipping
at browser zoom while retaining DOM containment, source-owned dismissal and
registered keyboard focus. Saved View discovery retains its independent browsing
lifetime and placement.

## Surface and shell presentation

| File | Responsibility |
| --- | --- |
| [AssessmentWorkbookSurface.tsx](AssessmentWorkbookSurface.tsx) | Assessment presentation facade for rows, support selection, and semantic assessment commands. |
| [EntityWorkbookSurface.tsx](EntityWorkbookSurface.tsx) | Hosts and Identities presentation over entity query/edit/merge/continuity owners and borrowed retained ordinary creation. |
| [GenericWorkbookSurface.tsx](GenericWorkbookSurface.tsx) | Contract-backed grid presentation borrowing retained ordinary drafts, local recovery and source command ports. |
| [IncidentControlsDrawer.tsx](IncidentControlsDrawer.tsx) | Shell-level incident controls drawer presentation and focus boundary. |
| [SystemViewSwitcher.tsx](SystemViewSwitcher.tsx) | System-view switcher UI and grouped surface navigation. |
| [WorkbookActiveSurfaceFrame.tsx](WorkbookActiveSurfaceFrame.tsx) | Active-surface recovery boundary and desktop built-in tab panels, preserving rejected editors' accessibility and the single active surface mount. |
| [WorkbookActiveSurfaceFrame.test.tsx](WorkbookActiveSurfaceFrame.test.tsx) | Keeps the original editor accessible and checks grouped conflict selection, drafts, feedback and focus through delayed resolution and projection refresh. |
| [WorkbookActiveSurfacePresentation.tsx](WorkbookActiveSurfacePresentation.tsx) | Exact built-in or extension renderer selection with lazy extension lifecycle binding. |
| [WorkbookIncidentControlsPresentation.tsx](WorkbookIncidentControlsPresentation.tsx) | Incident-controls drawer content and lazy Import Assistant renderer selection. |
| [WorkbookPresenceMarkers.tsx](WorkbookPresenceMarkers.tsx) | Shared row-gutter and cell presence markers with design-owned capacity and overflow behavior. |
| [WorkbookShellSlots.tsx](WorkbookShellSlots.tsx) | Stable shell slot IDs, labels, and layout slot helpers. |
| [WorkbookShellTopBar.test.tsx](WorkbookShellTopBar.test.tsx) | Tests desktop manual navigation, explicit activation, and responsive selector focus ownership. |
| [WorkbookShellTopBar.tsx](WorkbookShellTopBar.tsx) | Manual desktop built-in tab navigation, responsive selector focus ownership, registered menu focus, incident identity, presence, and account presentation. |
| [WorkbookIncidentIdentityDisclosure.tsx](WorkbookIncidentIdentityDisclosure.tsx) | Shared native incident h1 around the full-identity disclosure button; the disclosed region is outside the heading. Retains pointer, keyboard, Escape and focus departure behavior. |
| [WorkbookIncidentIdentityDisclosure.test.tsx](WorkbookIncidentIdentityDisclosure.test.tsx) | Heading identity updates, uniqueness, disclosure keyboard operation, Escape focus return and outside dismissal. |

`workbookFormStyles.ts` supplies the shared quiet command treatment. Feature
owners still supply admission, pending and recovery state. Saved-view Startup
disclosure state controls presentation only; preference operations and saved-view
confirmation remain in their existing controllers.

## Editing and reference controls

| File | Responsibility |
| --- | --- |
| [GenericMutationControl.test.tsx](GenericMutationControl.test.tsx) | Tests field-control descriptors, variants, options, sizing, and form/grid presentation. |
| [GenericMutationControl.tsx](GenericMutationControl.tsx) | Renders descriptor-driven field editors for generic form and grid mutation surfaces. |
| [genericMutationControlModel.ts](genericMutationControlModel.ts) | Pure field-control descriptors for generic form and grid mutation editors. |
| [WorkbookAuthoringReferenceControl.tsx](WorkbookAuthoringReferenceControl.tsx) | Shared staged/paged reference picker, with compact grid popover geometry and keyboard exit; selection identities belong to the authoring owner. |
| [WorkbookGridEditorControl.tsx](WorkbookGridEditorControl.tsx) | Contract-field grid editor adapter, mutation controls, commit/cancel behavior, and editor-kind selection. |
| [WorkbookReferenceControl.tsx](WorkbookReferenceControl.tsx) | Existing-record exact-ID input, staged top-layer picker, and authorized selected presentation; no mutation owner. |
| [WorkbookReferenceControl.test.tsx](WorkbookReferenceControl.test.tsx) | Inspector staging, explicit acceptance, read-only retry, exact collection payloads, and committed-label updates. |
| [WorkbookRecordCandidatePicker.tsx](WorkbookRecordCandidatePicker.tsx) | Native single-target candidate control retained for Party linking and single authoring. |
| [WorkbookMultiCandidatePicker.tsx](WorkbookMultiCandidatePicker.tsx) | Bounded checkbox presentation with complete wrapped labels and stable identities; explicit identity toggles only. |
| [WorkbookRelationshipChip.test.tsx](WorkbookRelationshipChip.test.tsx) | Tests relationship-chip state details, semantic selectors, and optional selection behavior. |
| [WorkbookRelationshipChip.tsx](WorkbookRelationshipChip.tsx) | Shared relationship-chip presentation over an explicit label, state, detail, selector identity, selection, and command model. |

## Status and recovery

| File | Responsibility |
| --- | --- |
| [WorkbookRecoveryPanel.tsx](WorkbookRecoveryPanel.tsx) | Compact Recovery entry, grouped discovery, one attached detail, focus transitions and accessible return to the grid. |
| [SavedViewRecovery.tsx](SavedViewRecovery.tsx) | Saved-view operation recovery presentation with stable recovery identities. |
| [WorkbookEditRecoveryPanel.tsx](WorkbookEditRecoveryPanel.tsx) | Workbook edit recovery panel for retained pending work and explicit recovery actions. |
| [WorkbookQueueOverflowNotice.tsx](WorkbookQueueOverflowNotice.tsx) | Pending mutation queue overflow feedback and recovery affordances. |
| [WorkbookSameFieldConflictResolver.tsx](WorkbookSameFieldConflictResolver.tsx) | Same-field comparison, grouped navigation and explicit resolution controls with conflict-specific pending/feedback and revocable focus intent. |
| [WorkbookSaveAnnouncements.tsx](WorkbookSaveAnnouncements.tsx) | Accessible announcements derived from workbook save and recovery state. |
| [WorkbookStatusStrip.tsx](WorkbookStatusStrip.tsx) | Status strip presentation for save/load/selection state. |

## Batch recovery

| File | Responsibility |
| --- | --- |
| [WorkbookBatchRecovery.tsx](WorkbookBatchRecovery.tsx) | Compact non-color batch status, original-input recovery, keyboard Retry, reads-only Retry refresh and activation of retained per-cell conflict groups. |

Batch Retry keeps its initiating control focusable and stably named while the
retained owner dispatches exact replay or read-only refresh. Attachment-local
intent identifies the action and current authority; it does not retain a request,
receipt or second failure state. Newer keyboard, pointer, scroll, selection and
departure intent prevents settlement from reclaiming focus. If scrolling leaves
focus on a settled control, or Shift+Tab leaves focus on original input, that
focused presentation remains until focus departs so the shell's removed-focus
fallback does not reset the panel scroll or replace newer keyboard movement. Genuine
item pruning and authority withdrawal still use that fallback.

Conflict resolution admission remains with `WorkbookMutationRuntime`; the resolver
retains only attachment-local pending and feedback by conflict key and token.
Moving to another conflict retires the former submission's focus/scroll intent
without canceling its write. A delayed outcome may update its own conflict and
draft through the runtime, but cannot close Recovery or reclaim a newer editor.
An eligible selected resolution returns to its originating cell only after
semantic focus completes and its activation, authority and interaction intent
remain current. Explicit Close and ordinary Recovery-list fallback retain their
own focus behavior.

The presentation subscribes to retained runtime state. Opening batch recovery
attaches one batch identity; reviewing its conflicts uses the existing resolver
inside the same panel. Bounded original-input previews distinguish multiple
batches. Routine acceptance opens no panel.

[WorkbookParkedGridDrafts.tsx](WorkbookParkedGridDrafts.tsx) exposes readable,
copyable local grid text only when its original target is unavailable, with
source-owned eligibility and discard callbacks. It owns no draft state.

[WorkbookSurfaceRefreshNotice.tsx](WorkbookSurfaceRefreshNotice.tsx) exposes the
existing surface registry obligation as compact read-only recovery. It neither
replays mutations nor owns refresh debt. `GenericMutationControl` preserves unknown
grid enum/reference input visibly; inspector controls keep their own contract.

WorkbookReferenceControl keeps exact-ID typing in cells and stages candidate selection in a native top-layer popup. The popup marks its nested interaction boundary for Grid Adapter capture handlers. Accept commits through the existing grid interaction or updates only the inspector draft. Candidate reads and cancellation never dispatch writes.

Shell recovery consumers register safe summaries through the shared recovery
boundary. Their former independent triggers and panel placement are retired.
`WorkbookActiveSurfaceFrame` contributes FIFO/overflow/conflict obligations and
keeps urgent notices visible independently of the selected detail. Explicit
status actions use semantic keys, including parent batch/compound operations.
Opening never submits; closing never discards. Completed-notice suppression does
not delete a receipt. Inspector and workspace dialog activation detach recovery.

## Authoring discovery

WorkbookCandidateBrowsing presents bounded page navigation and typed read recovery;
it owns no selection or mutation. An initiating read control keeps its name and
keyboard position while the read is pending. Retry and exhausted page controls
remain focusable and unavailable while focused, then become natively disabled
when focus departs. A settled Retry stays mounted so pointer activation of the
next control keeps stable geometry. Local focus intent retires on newer input or discovery
scope changes; candidate admission and stale-response handling stay with the
discovery owner.

Authoring record pickers use WorkbookAuthoringReferencePicker for attachment-local
staging, WorkbookCandidateSelection for explicit identity toggles and selected-item
removal, WorkbookCandidateQueryControl for explicitly applied schema queries,
and WorkbookCandidateBrowsing for local read recovery. Selections remain separate
from the accepted page; callers supply their own maximum and semantic target.
Source-review callers alone request captured row versions. Apply never dispatches
a write. Ordinary collection reference controls and multiple-mode authoring/support
use WorkbookMultiCandidatePicker. Its checkboxes add or remove one staged identity;
focus movement and label inspection do not change selection or issue reads. Labels
and full identities wrap inside the existing bounded chooser. Count limits, retained
metadata, query/paging and acceptance remain with their existing owners. The first
enabled checkbox supplies selected-item removal focus fallback, revealed through
nested owned scrollports. Settled ordinary collection Retry stays mounted and
disabled after focus departs so the following pointer gesture keeps stable geometry.
WorkbookRecordCandidatePicker retains native single-target presentation for Party
linking, Assessment subjects, source review and single-reference authoring.

`WorkbookBatchRecordChoices.tsx` presents up to twenty returned Timeline records
with local search and explicit review activation. Its test verifies bounded
choices and absence of implicit activation. Receipt lifetime remains with the
batch owner; this presentation retains no independent inventory.

When responsive overflow retires a focused query chip, its existing menu-open
request carries the fallback control. The menu owner reveals the committed panel
before focusing that connected descendant; no hidden control receives focus.

Outer Query and View options menu requests apply only when that menu is enabled in the current chrome mode. A base-layout command cannot retain an invisible open that appears after later resizing; focused-chip fallback still opens its current compact Query panel before committed focus.
