# workbook/timeline/

[Parent](../README.md) · [Source overview](../../README.md)

Timeline composition, presentation, actions, editing, and semantic runtime integration.

[TimelineWorkbook.tsx](components/TimelineWorkbook.tsx) is the public facade and
sole root composition-hook caller. The private composer assembles focused
owners and publishes a narrow presentation projection:

`TimelineWorkbook → root composition → presentation model → stateless view`

Generic workbook behavior stays in the [parent workbook](../README.md).
Root tests characterize composition boundaries and cross-directory interactions.

`WorkbookRow.rawRow` contains canonical accepted cells. Nullable and absent cells
must remain distinct from empty text. `values`, `committedValues` and retained
drafts are display/authoring projections; they never reconstruct the canonical
row. Committed-version tracking strips local presentation state without rewriting
accepted cells. Discard recovery reapplies remaining FIFO work only to its local
authoring projection, preserving the canonical baseline for Details, History and
the next explicit edit.

## Subdirectories

| Directory | Responsibility |
| --- | --- |
| [actions/](actions/README.md) | Reviewed Timeline capture actions and mention resolution, including optional Entity creation and recovery. |
| [adapters/](adapters/README.md) | Timeline operation transport, protocol validation, and private grid/runtime integration adapters. |
| [bulk/](bulk/README.md) | Timeline bulk tagging and fill workflows over stable record and field identities. |
| [collaboration/](collaboration/README.md) | Timeline active-surface registration, decoded event effects, and semantic presence publication. |
| [components/](components/README.md) | Timeline grid, cells, editors, inspector sections, and renderer composition. |
| [composition/](composition/README.md) | Private assembly of Timeline query, mutation, inspector, interaction, and presentation capabilities. |
| [editing/](editing/README.md) | Timeline scalar editor drafts and registered inputs keyed by semantic row, field, and surface identity. |
| [focus/](focus/README.md) | Timeline inspector element registration and subject-bound panel/mention focus. |
| [hooks/](hooks/README.md) | Timeline query, mutation, inspector, owner-action, and interaction coordination. |
| [models/](models/README.md) | Pure Timeline row, mutation, load, action, keyboard, and continuity models. |
| [mutations/](mutations/README.md) | Monotonic admission and projection of Timeline query, mutation, replay, and live-event results. |
| [ports/](ports/README.md) | Narrow Timeline bulk-tag, attachment, mention, and capture-action capabilities. |
| [presentation/](presentation/README.md) | Timeline presentation derivation and stateless surface regions. |

## Tests

| File | Responsibility |
| --- | --- |
| [timelineCaptureActions.characterization.test.ts](timelineCaptureActions.characterization.test.ts) | Tests History placement, distinct capture confirmation policies, and exact supersession reason omission. |
| [timelineCompositionArchitecture.test.ts](timelineCompositionArchitecture.test.ts) | Enforces the slim public root, sole composition-hook caller, stateless presentation regions, forbidden-capability exclusion, and visible-column synchronization ownership. |
| [useTimelineCommittedRecordIdle.test.tsx](useTimelineCommittedRecordIdle.test.tsx) | Characterizes bounded refresh when a committed record version is temporarily missing. |
| [useTimelineCompositionLifecycle.test.tsx](useTimelineCompositionLifecycle.test.tsx) | Characterizes grid measurement/observer cleanup and inspector selection/continuity reset ownership. |
| [useTimelineCreateRelatedWorkflow.test.tsx](useTimelineCreateRelatedWorkflow.test.tsx) | Tests retained Evidence creation across navigation and preservation of the original unsent draft. |
| [useTimelineInspectorFeatureController.test.tsx](useTimelineInspectorFeatureController.test.tsx) | Characterizes canonical feature routing, fail-closed behavior, cancellation, exclusivity, and lifecycle reset. |
| [useTimelineInspectorLifecycle.test.tsx](useTimelineInspectorLifecycle.test.tsx) | Characterizes the shared explicit/layout inspector close command. |
| [useTimelineKeyboardController.test.tsx](useTimelineKeyboardController.test.tsx) | Characterizes scalar/collection commit, navigation, range, draft, inspector, work-area, and event-consumption keyboard ownership. |
| [useTimelineMentionActions.test.tsx](useTimelineMentionActions.test.tsx) | Characterizes auto-resolution undo identity, committed-version refresh/continuity sequencing, and rejection behavior. |
| [useTimelineMutationRuntimeBindings.test.tsx](useTimelineMutationRuntimeBindings.test.tsx) | Tests stable mounted registration, current callback dispatch and unmount cleanup. |
| [useTimelineObservationSource.test.tsx](useTimelineObservationSource.test.tsx) | Tests exact committed Observation source text and rejection after saved or local source changes. |
| [useTimelineRowActionMenu.test.tsx](useTimelineRowActionMenu.test.tsx) | Covers menu validity and cancellation across semantic focus fallback requests. |
| [useTimelineRows.test.tsx](useTimelineRows.test.tsx) | Characterizes initial draft-row identity, stable row refs, and monotonic draft allocation. |
| [useTimelineSurfaceFoundation.test.tsx](useTimelineSurfaceFoundation.test.tsx) | Characterizes stable adapter, row/query, pending-save, and semantic foundation identities. |

Conflict cell state and local recovery buttons project the retained Workbook
conflict store, including batch groups. Timeline's local conflict draft queue
remains with scalar mutation coordination; it does not own batch presentation
or determine which retained batch conflicts remain actionable.

## Cell-range interaction

Timeline explicitly enables Adapter contiguous selection. Mutation composition
supplies a scope key from the incident/authority lifetime, sheet and accepted
canonical query; pending query overrides do not replace selection identity.
Adapter owns the sole completed `GridCellRange` and tentative pointer state.
Range completion focuses its endpoint and preserves inspector context and bulk
checkbox state. Scalar sessions retain raw text, validation, deduplicated commit
and rejected-draft recovery. Existing clipboard and one-column fill planners
retain their stable-ID, row-version, authorization and grouped-fill boundaries.
Selection creates no records or bulk operations.
