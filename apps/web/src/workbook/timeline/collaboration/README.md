# workbook/timeline/collaboration/

[Parent](../README.md) · [Source overview](../../../README.md)

Timeline active-surface registration, decoded event effects, and semantic presence publication.

Bindings use the [workbook coordinator](../../collaboration/README.md) active
surface port. The shared browser session owns socket lifetime, sequencing,
and transport; Timeline owns its row effects and semantic presence.

## Files

| File | Responsibility |
| --- | --- |
| [TimelineCollaborationBoundary.tsx](TimelineCollaborationBoundary.tsx) | Owns the Timeline collaboration session boundary and optional coordinator-session attachment around the presentation root. |
| [useTimelineCollaborationBindings.ts](useTimelineCollaborationBindings.ts) | Owns Timeline active-surface and transaction-resolver registration, sparse row patching, refresh/reset effects, presence publication, and lifecycle teardown. |
| [useTimelinePresenceController.ts](useTimelinePresenceController.ts) | Derives Timeline row/cell presence and publishes semantic viewing/editing transitions for the active sheet. |

## Tests

| File | Responsibility |
| --- | --- |
| [useTimelineCollaborationBindings.test.tsx](useTimelineCollaborationBindings.test.tsx) | Characterizes live/stale admission, fail-closed sparse patches, gap refresh, access invalidation, surface switching, and teardown. |
| [useTimelinePresenceController.test.tsx](useTimelinePresenceController.test.tsx) | Characterizes stable record/field presence derivation, coherent viewing/editing publication, and lifecycle reset without stale publication. |
