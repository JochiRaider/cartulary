# workbook/adapters/

[Parent](../README.md) · [Source overview](../../README.md)

Private workbook protocol adapters, captured operation transport, response validation, and semantic outcome conversion.

Composition creates incident-bound adapters and injects semantic capabilities
into queries, startup, preferences, saved views, and mutation owners. Transport
status, routes, envelopes, and failures are converted here.

[WorkbookClipboardPastePort.ts](WorkbookClipboardPastePort.ts) is the private
paste capability that intentionally carries generated request vocabulary and
typed rows/conflicts for surface-specific validation. General semantic
capabilities live in [ports](../ports/README.md).

## Feature mutation transport

| File | Responsibility |
| --- | --- |
| [createAssessmentAppendTransport.ts](createAssessmentAppendTransport.ts) | Executes captured Assessment append requests and validates correlated operation receipts. |
| [createContextualCreateTransport.ts](createContextualCreateTransport.ts) | Executes reviewed contextual Task/Decision creation through generated operations. |
| [createCoordinationCreateTransport.ts](createCoordinationCreateTransport.ts) | Executes captured coordination creation requests and validates semantic receipts. |
| [createIndicatorCreateTransport.test.ts](createIndicatorCreateTransport.test.ts) | Tests canonical creation request constraints, exact replay bytes, and complete receipt validation. |
| [createIndicatorCreateTransport.ts](createIndicatorCreateTransport.ts) | Executes captured canonical Indicator creation and validates correlated accepted receipts. |
| [createIndicatorLifecycleAdapter.test.ts](createIndicatorLifecycleAdapter.test.ts) | Tests complete lifecycle receipts, uncertain malformed success, and exact ordered-support replay. |
| [createIndicatorLifecycleAdapter.ts](createIndicatorLifecycleAdapter.ts) | Executes Indicator lifecycle requests and validates complete operation receipts. |
| [createNoteCreateTransport.ts](createNoteCreateTransport.ts) | Executes captured Note creation requests and validates semantic receipts. |
| [createObservationTransport.test.ts](createObservationTransport.test.ts) | Tests exact Observation child requests, omission, replay, provenance, and receipt correlation. |
| [createObservationTransport.ts](createObservationTransport.ts) | Executes Observation capture/transition requests and validates complete correlated receipts. |
| [createPartyCreationTransport.ts](createPartyCreationTransport.ts) | Captures Party creation requests and executes them with typed receipts and outcomes. |
| [createTimelineRelatedEvidenceTransport.ts](createTimelineRelatedEvidenceTransport.ts) | Executes separately captured Evidence creation and Timeline linking operations. |
| [sendWorkbookRecordMutation.ts](sendWorkbookRecordMutation.ts) | Sends captured workbook record mutations and normalizes accepted rows and failures. |
| [workbookRecordPatchTransport.ts](workbookRecordPatchTransport.ts) | Captures exact record patches and normalizes validated mutation rows and outcomes. |

## Authoring and candidate reads

| File | Responsibility |
| --- | --- |
| [createAssessmentCandidateReader.ts](createAssessmentCandidateReader.ts) | Adapts authorized workbook queries to Assessment subject and support candidate pages. |
| [createContextualCreateReader.ts](createContextualCreateReader.ts) | Exposes the contextual creation reader through the shared workbook authoring adapter. |
| [createDecisionCandidateReader.ts](createDecisionCandidateReader.ts) | Reads paged Decision replacement candidates with exact workbook context validation. |
| [createIndicatorLifecycleReader.ts](createIndicatorLifecycleReader.ts) | Reads and validates paged Indicator lifecycle resources and support candidates. |
| [createNoteCreateReader.ts](createNoteCreateReader.ts) | Adapts workbook authoring reads for Note source and reference discovery. |
| [createObservationReader.ts](createObservationReader.ts) | Reads Observation resources and candidates with validated paging. |
| [createPartyLinkReader.ts](createPartyLinkReader.ts) | Adapts workbook authoring queries to paged Party linking candidates. |
| [createWorkbookAuthoringReader.ts](createWorkbookAuthoringReader.ts) | Shared adapter for authorized workbook authoring discovery and record reads. |
| [readWorkbookAuthoringRecord.ts](readWorkbookAuthoringRecord.ts) | Reads and validates one authoring record within its workbook and source-owner context. |

## Workbook operation adapters

| File | Responsibility |
| --- | --- |
| [createWorkbookClipboardPasteAdapter.test.ts](createWorkbookClipboardPasteAdapter.test.ts) | Tests exact request projection, raw typed results, invalid input, response correlation, secure-ID failure, and transport containment. |
| [createWorkbookClipboardPasteAdapter.ts](createWorkbookClipboardPasteAdapter.ts) | Executes the sole generated clipboard-paste operation with secure transaction identity and exact response-surface validation. |
| [createWorkbookDecisionSupersessionAdapter.test.ts](createWorkbookDecisionSupersessionAdapter.test.ts) | Tests complete supersession receipts, exact replay with current credentials, and candidate boundaries. |
| [createWorkbookDecisionSupersessionAdapter.ts](createWorkbookDecisionSupersessionAdapter.ts) | Executes captured Decision supersession requests through validated generated operations. |
| [createWorkbookEntityMergeAdapter.test.ts](createWorkbookEntityMergeAdapter.test.ts) | Tests exact Entity merge replay, current CSRF, full receipts, and uncertain malformed outcomes. |
| [createWorkbookEntityMergeAdapter.ts](createWorkbookEntityMergeAdapter.ts) | Executes captured Entity merges and validates both participants' receipt identities and versions. |
| [createWorkbookIncidentAdapter.ts](createWorkbookIncidentAdapter.ts) | Loads validated incident identity and memberships with exact resource correlation. |
| [createWorkbookPendingMutationAdapter.test.ts](createWorkbookPendingMutationAdapter.test.ts) | Tests create/patch projection, identity correlation, accepted rows, and semantic failure outcomes. |
| [createWorkbookPendingMutationAdapter.ts](createWorkbookPendingMutationAdapter.ts) | Executes queued create/patch units with dispatch identity, response correlation, and view-contract normalization. |
| [createWorkbookPreferenceAdapter.test.ts](createWorkbookPreferenceAdapter.test.ts) | Tests all preference operations, explicit nullable bodies, CSRF, and resource correlation. |
| [createWorkbookPreferenceAdapter.ts](createWorkbookPreferenceAdapter.ts) | Loads and updates correlated current-user home and incident-default workbook preferences. |
| [createWorkbookRecordHistoryAdapter.test.ts](createWorkbookRecordHistoryAdapter.test.ts) | Tests retained History mutation bytes, full receipts, and uncertain transport or malformed outcomes. |
| [createWorkbookRecordHistoryAdapter.ts](createWorkbookRecordHistoryAdapter.ts) | Adapts record History reads and captured rollback/restore operations to semantic outcomes. |
| [createWorkbookSavedViewAdapter.test.ts](createWorkbookSavedViewAdapter.test.ts) | Tests paging progress, CRUD correlation, versions, immutability, and malformed responses. |
| [createWorkbookSavedViewAdapter.ts](createWorkbookSavedViewAdapter.ts) | Executes explicit saved-view paging and CRUD behind incident-bound semantic outcomes. |
| [createWorkbookStartupAdapter.ts](createWorkbookStartupAdapter.ts) | Loads validated Workbook startup state and correlated extension availability. |
| [createWorkbookStartupAndIncidentAdapters.test.ts](createWorkbookStartupAndIncidentAdapters.test.ts) | Tests startup, incident, membership, and preference operation boundaries. |
| [createWorkbookViewQueryAdapter.test.ts](createWorkbookViewQueryAdapter.test.ts) | Tests projected query requests, aborts, malformed responses, and cross-context rejection. |
| [createWorkbookViewQueryAdapter.ts](createWorkbookViewQueryAdapter.ts) | Executes one abortable Workbook query boundary with exact incident/schema correlation and contract-row normalization. |
| [WorkbookClipboardPastePort.ts](WorkbookClipboardPastePort.ts) | Exact private generated-vocabulary transport capability shared by paste-capable Workbook surface owners. |
| [workbookOperationExecutor.ts](workbookOperationExecutor.ts) | Executes the closed Workbook operation-ID set and converts validated success/error envelopes to semantic outcomes. |

## Contracts and result validation

| File | Responsibility |
| --- | --- |
| [entityIdentifierNormalization.ts](entityIdentifierNormalization.ts) | Adapts Entity identifier and alias normalization and normalized-value comparison. |
| [indicatorCreateProtocol.ts](indicatorCreateProtocol.ts) | Private generated contract facade for canonical Indicator creation. |
| [indicatorLifecycleProtocol.ts](indicatorLifecycleProtocol.ts) | Private generated lifecycle constraints, values, intervals, and receipt types. |
| [observationContracts.test.ts](observationContracts.test.ts) | Tests typed Observation source-span and transition contracts exclude caller-authored provenance. |
| [observationProtocol.ts](observationProtocol.ts) | Private generated contract facade for Observation operations and receipts. |
| [workbookAdapterResult.ts](workbookAdapterResult.ts) | Normalizes operation outcomes into shared semantic port results. |
| [workbookHistoryResponse.ts](workbookHistoryResponse.ts) | Private record History response types and paging adaptation. |
| [workbookOperationContract.ts](workbookOperationContract.ts) | Closed workbook operation IDs and typed executor request/response contracts. |
| [workbookOperationErrorPolicy.test.ts](workbookOperationErrorPolicy.test.ts) | Tests malformed-envelope rejection and exact conflict, authorization, and stale-target classification. |
| [workbookOperationErrorPolicy.ts](workbookOperationErrorPolicy.ts) | Classifies decoded workbook operation failures into semantic owner outcomes. |
| [workbookProtocolTypes.ts](workbookProtocolTypes.ts) | Private type-only projection of the exact generated Workbook request vocabulary; prevents protocol imports from leaking into owner models, controllers, or runtime code. |
| [workbookPublicErrorDecoder.ts](workbookPublicErrorDecoder.ts) | Decodes and sanitizes public workbook errors at the adapter boundary. |
