# workbook/timeline/composition/

[Parent](../README.md) · [Source overview](../../../README.md)

Private assembly of Timeline query, mutation, inspector, interaction, and presentation capabilities.

Only the root composer imports sibling composition owners. Lower Timeline
layers do not depend back on composition. This directory assembles semantic
owners without importing presentation, generated protocol bindings, or browser
services.

## Files

| File | Responsibility |
| --- | --- |
| [useTimelineGridEnvironment.ts](useTimelineGridEnvironment.ts) | Owns Timeline grid refs, rounded width observation and cleanup, focus and anchor commands, viewport continuity, and row-mutation editor adaptation. |
| [useTimelineInspectorStateComposition.ts](useTimelineInspectorStateComposition.ts) | Owns Timeline selection, inspector open/reset continuity, invalidation state, feedback, and row-history state. |
| [useTimelineInspectorWorkflowComposition.ts](useTimelineInspectorWorkflowComposition.ts) | Owns Timeline feature routing, related-record, history, mention, Evidence, row-menu, close, and Escape workflows. |
| [useTimelineInteractionComposition.ts](useTimelineInteractionComposition.ts) | Owns Timeline bulk, keyboard, paste, fill, scalar/collection commit, draft, collection-focus, and recovery-focus interactions. |
| [useTimelineMutationComposition.ts](useTimelineMutationComposition.ts) | Owns the explicit Timeline query admission, row mutation, replay, runtime registration, collaboration, presence, conflict, and save command graph. |
| [useTimelineSurfaceFoundation.ts](useTimelineSurfaceFoundation.ts) | Owns semantic Timeline adapter construction, query/lifecycle foundations, row and mention state, pending-save refs, editor drafts, and guarded workbook timing. |
| [useTimelineWorkbookComposition.ts](useTimelineWorkbookComposition.ts) | Private root composer that invokes Timeline composition owners in dependency order, returns grouped capabilities, and publishes the capability-limited presentation projection. |
