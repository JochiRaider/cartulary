# workbook/timeline/components/

[Parent](../README.md) · [Source overview](../../../README.md)

Timeline grid, cells, editors, inspector sections, and renderer composition.

[TimelineWorkbook.tsx](TimelineWorkbook.tsx) delegates composition and rendering.
Focused cells, editors, inspector sections, and renderer factories consume
semantic models and commands. Surface layout regions live in
[presentation](../presentation/README.md).

## Files

| File | Responsibility |
| --- | --- |
| [TimelineBulkTagControl.tsx](TimelineBulkTagControl.tsx) | Owns mounted raw tag authoring and local feedback; observes captured batch outcomes without changing selection or grid focus. |
| [TimelineCollectionCell.tsx](TimelineCollectionCell.tsx) | Focused relationship/tag summary, overflow, and collection-draft cell presentation over discriminated models. |
| [TimelineDraftRowActions.tsx](TimelineDraftRowActions.tsx) | Timeline draft-row create and evidence-attachment actions. |
| [TimelineEvidencePanel.tsx](TimelineEvidencePanel.tsx) | Timeline inspector evidence panel and evidence actions UI. |
| [TimelineHistoryPanel.tsx](TimelineHistoryPanel.tsx) | Timeline row history, rollback, delete, restore, and history action presentation. |
| [TimelineMentionActionControls.tsx](TimelineMentionActionControls.tsx) | Mention action, Entity creation, and operation-recovery feedback controls. |
| [TimelineMentionsPanel.tsx](TimelineMentionsPanel.tsx) | Timeline mention-resolution inspector panel. |
| [TimelineRowActions.tsx](TimelineRowActions.tsx) | Timeline row action/context-menu presentation. |
| [TimelineScalarEditor.tsx](TimelineScalarEditor.tsx) | Timeline scalar input/textarea editing, commit, presence, clipboard, and grid-editor lifecycle behavior. |
| [TimelineWorkbook.tsx](TimelineWorkbook.tsx) | Public Timeline facade that retains the collaboration boundary and delegates grouped composition, presentation derivation, and stateless rendering. |
| [TimelineWorkbookGrid.tsx](TimelineWorkbookGrid.tsx) | Timeline grid renderer, grouped row table wrapper, hidden contract metadata cells, and grid test-ID placement. |
| [TimelineWorkbookInspector.tsx](TimelineWorkbookInspector.tsx) | Timeline inspector shell, panel tabs, disabled-state presentation, selected-row state, and inspector messages. |
| [TimelineWorkbookInspectorSections.tsx](TimelineWorkbookInspectorSections.tsx) | Timeline inspector section factories for field editors, relationships, evidence attach, related-row creation, and row history. |
| [TimelineWorkbookNotices.tsx](TimelineWorkbookNotices.tsx) | Retained auto-resolution disclosure list and local actions in the bounded above-grid feedback region. |
| [TimelineWorkbookRenderers.tsx](TimelineWorkbookRenderers.tsx) | Stable private facade composing scalar, collection, and column renderer owners. |
| [TimelineWorkbookRendererTypes.ts](TimelineWorkbookRendererTypes.ts) | Private renderer command and output types shared by the Timeline renderer workstreams. |
| [TimelineWorkbookStyles.ts](TimelineWorkbookStyles.ts) | Timeline-specific style constants shared by Timeline workbook components. |
| [useTimelineCollectionRenderer.tsx](useTimelineCollectionRenderer.tsx) | Narrow renderer factory that binds Timeline collection commands to the focused collection-cell component. |
| [useTimelineColumnAssembly.tsx](useTimelineColumnAssembly.tsx) | Timeline contract column ordering, widths, editors, clipboard values, and evidence/read-only cell assembly. |
| [useTimelineScalarRenderers.tsx](useTimelineScalarRenderers.tsx) | Timeline scalar grid/inspector controls, read cells, presence, and conflict markers. |

## Tests

| File | Responsibility |
| --- | --- |
| [TimelineCollectionCell.test.tsx](TimelineCollectionCell.test.tsx) | Tests hidden collection-item inspection without editing or committing pending text. |
| [TimelineEvidencePanel.test.tsx](TimelineEvidencePanel.test.tsx) | Tests for Timeline evidence panel behavior. |
| [TimelineScalarEditor.test.tsx](TimelineScalarEditor.test.tsx) | Characterizes controlled scalar drafts, read-only behavior, presence publication, and commit lifecycle. |

Timeline enables Grid Adapter's optional cyclic range entry. The Adapter owns
semantic traversal, range retention, F2 activation and acceptance-gated departure.
Timeline retains field authorization and mutation ownership. Managed scalar
focus publishes its semantic anchor and presence without selecting Inspector
context; stationary pointer inspection remains explicit. Multiline Shift+Enter
and composition remain local text input.
