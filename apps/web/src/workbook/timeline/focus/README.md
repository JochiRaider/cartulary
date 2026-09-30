# workbook/timeline/focus/

[Parent](../README.md) · [Source overview](../../../README.md)

Timeline inspector element registration and subject-bound panel/mention focus.

Focus requests are tied to the captured canonical inspector subject and
registered elements, including panel and mention identities.

Mention-list arrows keep their group-local neighbor calculation in the panel and
borrow focus through this registry with nearest placement. The shared Inspector
navigator owns availability and body reveal; the panel has no second button map.
Mention focus requests are synchronous and are never retained for late registration.

## Files

| File | Responsibility |
| --- | --- |
| [timelineInspectorElementRegistry.ts](timelineInspectorElementRegistry.ts) | Registers Timeline inspector panel/mention elements and resolves subject-bound semantic focus. |

## Tests

| File | Responsibility |
| --- | --- |
| [timelineInspectorElementRegistry.test.ts](timelineInspectorElementRegistry.test.ts) | Tests panel and mention focus is restricted to the captured canonical inspector subject. |
