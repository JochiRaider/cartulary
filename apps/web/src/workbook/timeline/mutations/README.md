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

The retained Timeline mutation owner also exposes the file-draft promotion port.
Screenshot-only creation and ordinary first input share its existing creation
identity. [timelineFileDraftRecovery.test.ts](timelineFileDraftRecovery.test.ts)
verifies both orderings, detached acceptance and follow-on edits.

`createTimelineMutationDriver` receives accepted clear predecessors through the
Workbook driver registry. It advances later undispatched scalar baselines without
altering exact dispatched attempts or conflicted fields. The editor draft registry
retains newer authoring, including an editor still open at acknowledgement.
`useTimelineRowMutationCoordinator.applyAcceptedBatchRows` publishes a complete
batch receipt in one projection commit and preserves existing mention notices.

Queue admission refusal is distinct from a submitted operation's settlement.
The driver reports refusal to the command owner so its scalar/collection
deduplication entry can retire while the exact authoring revision remains.
A later explicit commit can retry that unsubmitted work; submitted failures and
uncertain requests keep their existing recovery and transaction identities.

A `client_txn_conflict` halt does not settle its logical editor command. Queue recovery owns
attention while the original command remains pending; Retry or Discard settles
it once. Sending a rejection followed by acceptance to a Promise loses the later
result and can leave an acknowledged editor blocking the next gesture.

Other terminal rejections settle immediately, including validation errors, so
Find and range navigation can return to the original edit. Discarding their
retained queue unit does not settle the same command twice.

Accepted background input and fresh-draft restoration delegate to the cancellable
viewport-continuity owner. Active draft-to-record editor transfer remains here,
including newer text and native selection. Presentation cancellation cannot alter
accepted versions, acknowledgements, exact replay or required refresh reads.
