# workbook/timeline/mutations/

[Parent](../README.md) · [Source overview](../../../README.md)

Monotonic admission and projection of Timeline query, mutation, replay, and live-event results.

The coordinator admits query, mutation, replay, conflict, and live-event results
against committed-version evidence before presentation observes them.

## Files

| File | Responsibility |
| --- | --- |
| [useTimelineRowMutationCoordinator.ts](useTimelineRowMutationCoordinator.ts) | Applies accepted/discarded plans and composes committed-version, conflict, transaction, save-state, collaboration, and continuity owners. |

## Tests

| File | Responsibility |
| --- | --- |
| [useTimelineRowMutationCoordinator.test.tsx](useTimelineRowMutationCoordinator.test.tsx) | Characterizes accepted/stale action and mutation admission, query/action races, conflict state partitions, and committed-version high-water behavior. |
