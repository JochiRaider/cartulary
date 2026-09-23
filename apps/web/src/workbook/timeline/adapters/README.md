# workbook/timeline/adapters/

[Parent](../README.md) · [Source overview](../../../README.md)

Timeline operation transport, protocol validation, and private grid/runtime integration adapters.

Operation adapters derive transport from generated bindings and expose
owner-specific semantic outcomes. The directory also contains private adapters
for grid commits, editor continuity, projection commits, and runtime transaction
tracking.

## Files

| File | Responsibility |
| --- | --- |
| [createTimelineBulkTagCommandAdapter.ts](createTimelineBulkTagCommandAdapter.ts) | Maps source-owned tag targets/value into the retained batch admission contract. |
| [createTimelineCandidateReader.ts](createTimelineCandidateReader.ts) | Adapts validated workbook queries to Timeline capture-action candidate pages. |
| [createTimelineMentionCandidateReader.ts](createTimelineMentionCandidateReader.ts) | Adapts Entity workbook queries to mention-resolution candidate pages. |
| [createTimelineMentionEntityCreationAdapter.ts](createTimelineMentionEntityCreationAdapter.ts) | Creates a host or identity from a validated Timeline mention through an exact generated operation. |
| [createTimelineMentionResolutionAdapter.ts](createTimelineMentionResolutionAdapter.ts) | Resolves one versioned entity-mention action and validates its exact source and mention response identities. |
| [createTimelineMentionSourceReader.ts](createTimelineMentionSourceReader.ts) | Reads and validates authoritative Timeline sources for mention actions. |
| [createTimelineRecordActionAdapter.ts](createTimelineRecordActionAdapter.ts) | Executes and normalizes Timeline review and supersede actions. |
| [createTimelineRowMutationEditorAdapter.ts](createTimelineRowMutationEditorAdapter.ts) | Translates active editor promotion, cancellation and explicit recovery focus; background accepted restoration stays with viewport continuity. |
| [createTimelineScalarGridCommitAdapter.ts](createTimelineScalarGridCommitAdapter.ts) | Adapts Grid Adapter scalar commits to the Timeline scalar-save command and exact settlement promise. |
| [createTimelineSocketTransactionAdapter.ts](createTimelineSocketTransactionAdapter.ts) | Adapts Timeline accepted/action transaction tracking to the shell-lifetime runtime ledger. |
| [timelineCaptureProtocol.ts](timelineCaptureProtocol.ts) | Private generated Timeline review/supersession receipt contracts. |
| [timelineEvidenceRequestBuilders.ts](timelineEvidenceRequestBuilders.ts) | Constructs exact Evidence creation and Timeline attachment patch requests. |
| [timelineMentionProtocol.ts](timelineMentionProtocol.ts) | Validates complete mention-operation receipts and their correlated source/target identities. |
| [timelineProjectionCommitAdapter.ts](timelineProjectionCommitAdapter.ts) | Sole synchronous projection-commit boundary used when focus or continuity must observe the committed Timeline row tree. |

## Tests

| File | Responsibility |
| --- | --- |
| [createTimelineActionAdapters.test.ts](createTimelineActionAdapters.test.ts) | Characterizes history, record action, mention, and Evidence attachment protocol boundaries. |
| [timelineCaptureTransport.test.ts](timelineCaptureTransport.test.ts) | Tests bounded candidate pages and capture receipt identity, version, reason, and replacement correlation. |
| [timelineMentionProtocol.test.ts](timelineMentionProtocol.test.ts) | Tests complete mention receipts, raw text, opaque selectors, required nulls, and uncertain malformed success. |
