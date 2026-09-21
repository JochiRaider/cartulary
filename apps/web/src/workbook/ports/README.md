# workbook/ports/

[Parent](../README.md) · [Source overview](../../README.md)

Semantic workbook read/write capabilities and shared outcomes used across owner boundaries.

These capabilities expose semantic requests and outcomes without HTTP routes,
envelopes, status inspection, or API-base coordinates. The generated-vocabulary
clipboard exception remains private to [adapters](../adapters/README.md).

## Files

| File | Responsibility |
| --- | --- |
| [WorkbookAuthoringReadPort.ts](WorkbookAuthoringReadPort.ts) | Semantic authoring context, candidate pages, queries, record reads, and authority capabilities. |
| [WorkbookIncidentPort.ts](WorkbookIncidentPort.ts) | Incident identity and membership capabilities. |
| [WorkbookPendingMutationPort.ts](WorkbookPendingMutationPort.ts) | Shared queued create/patch execution capability over committed versions and semantic mutation units. |
| [WorkbookRecordSubject.ts](WorkbookRecordSubject.ts) | Read and operation subject identity shared by History and inspector attachment. |
| [WorkbookPortResult.ts](WorkbookPortResult.ts) | Shared accepted/aborted/authentication/authorization/stale/retryable/terminal semantic result union. |
| [WorkbookPreferencePort.ts](WorkbookPreferencePort.ts) | Current-user and incident-default preference capabilities. |
| [WorkbookSavedViewPort.ts](WorkbookSavedViewPort.ts) | Schema-scoped page discovery, addressed authorized resource reads, accepted-response CRUD and the injected cancellable observation capability shared by read/operation owners. |
| [WorkbookTimelineActionRuntimePort.ts](WorkbookTimelineActionRuntimePort.ts) | Narrow workbook runtime capability required by Timeline action owners. |

WorkbookReferenceReadPort separates record, Party and incident-member identities from presentation. It carries canonical record query metadata and paging without granting mutation admission. Collection item_ref removal remains a separate action contract.

## Authoring candidate discovery

| File | Responsibility |
| --- | --- |
| [WorkbookCandidateReadPort.ts](WorkbookCandidateReadPort.ts) | Typed candidate observation and query capability, independent of selection. |

## Accepted record observations

`WorkbookQueryRow` carries optional source evidence alongside its transport-shaped
fields. Query and mutation adapters attach record identity, accepted version and
the existing runtime account/session/incident scope with its read-revocation epoch.
An unobserved row is not eligible for retained inspector reading. Normalizers and
presentation conversions preserve matching evidence without granting authority.
Read revocation is distinct from mutation coordination resumption, query-window
eviction, incident closure and readable role changes. A delayed write receipt keeps
its dispatch scope and independent recovery lifetime. A delayed read crossing a
revocation boundary is rejected before presentation can accept it.
