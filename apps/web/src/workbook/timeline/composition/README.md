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

## Query continuation

The surface foundation owns the shared committed-record capability used by the
inspector and mutation coordinator. The query loader stages bounded browser
windows before Timeline projection acknowledges them. Browsing metadata and
checkpoints remain under Workbook query ownership.

Collection authoring shares its revision key across the grid and inspector.
Settlement clears the captured revision through one mounted draft registry, or
through the retained owner when detached. An earlier socket observation cannot
leave accepted text in an input or clear typing entered after dispatch.

Row-menu admission/lifetime is independent of Inspector selection. Composition
supplies the accepted query, incident and authorization scope with the semantic
grid handle; only explicit Inspect/History activation selects its destination.
Capture actions retain their existing source-owner preflight and mutation paths.

Deferred viewport continuity receives a stable observation port for the accepted
incident/surface/query scope and live collaboration/mutation authority. Pending
query replacements retain the accepted scope. Accepted fresh-draft and input
restoration use the continuity command; active editor promotion stays with the
mutation editor owner to preserve newer authoring and native selection.
