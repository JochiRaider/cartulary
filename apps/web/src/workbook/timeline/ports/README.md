# workbook/timeline/ports/

[Parent](../README.md) · [Source overview](../../../README.md)

Narrow Timeline bulk-tag, attachment, mention, and capture-action capabilities.

Each owner receives its required semantic capability. Ports do not expose
routes, status codes, raw transport envelopes, or API-base coordinates.

## Files

| File | Responsibility |
| --- | --- |
| [TimelineBulkTagCommandPort.ts](TimelineBulkTagCommandPort.ts) | Exact semantic multi-row tag assignment capability. |
| [TimelineMentionPort.ts](TimelineMentionPort.ts) | Separate semantic mention entity-creation and resolution capabilities. |
| [TimelineRecordActionPort.ts](TimelineRecordActionPort.ts) | Semantic Timeline review and supersede capability. |
