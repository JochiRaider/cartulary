# workbook/features/evidence/

[Parent](../README.md) · [Source overview](../../../README.md)

Evidence access and attachment workflows, including staged Timeline-related Evidence creation and linking.

Object-blob slot creation/upload, Evidence attachment, and source linking keep
their own operation identities. Timeline-related Evidence retains creation and
linking receipts separately. Shared message severity belongs to
[workbook Evidence presentation](../../evidence/README.md).

## Files

| File | Responsibility |
| --- | --- |
| [createEvidenceAttachmentPort.ts](createEvidenceAttachmentPort.ts) | Coordinates object-blob upload, Evidence attachment, and lifecycle finalization through semantic operations. |
| [createUploadedEvidenceBlob.ts](createUploadedEvidenceBlob.ts) | Creates a validated object-blob slot and uploads bytes to its server-issued target. |
| [EvidenceAccessActions.tsx](EvidenceAccessActions.tsx) | Evidence preview/download actions with access feedback and shared control styles. |
| [RelatedEvidencePartyControl.tsx](RelatedEvidencePartyControl.tsx) | Party selection and linking controls for Timeline-related Evidence authoring. |
| [TimelineRelatedEvidenceContext.ts](TimelineRelatedEvidenceContext.ts) | React context exposing Timeline-related Evidence operation ownership. |
| [TimelineRelatedEvidenceForm.tsx](TimelineRelatedEvidenceForm.tsx) | Timeline-related Evidence metadata authoring and reviewed create/link controls. |
| [timelineRelatedEvidenceModel.ts](timelineRelatedEvidenceModel.ts) | Timeline-related Evidence draft, lifecycle-state choices, validation, and request preparation. |
| [timelineRelatedEvidenceOperation.ts](timelineRelatedEvidenceOperation.ts) | Staged related-Evidence reviews, immutable attempts, receipts, and transport outcomes. |
| [TimelineRelatedEvidenceRecovery.tsx](TimelineRelatedEvidenceRecovery.tsx) | Separate creation and linking recovery for retained Timeline-related Evidence work. |
| [useEvidenceWorkbookBindings.tsx](useEvidenceWorkbookBindings.tsx) | Evidence-owned access, preview, download, and semantic attachment binding for the contract surface. |
| [useTimelineRelatedEvidenceAttachment.ts](useTimelineRelatedEvidenceAttachment.ts) | Attaches Timeline-related Evidence presentation to retained source-bound operation state. |
| [WorkbookTimelineRelatedEvidenceOwner.ts](WorkbookTimelineRelatedEvidenceOwner.ts) | Owns staged Evidence creation and Timeline linking with independent captured attempts and receipts. |

## Tests

| File | Responsibility |
| --- | --- |
| [timelineRelatedEvidenceAuthoring.test.tsx](timelineRelatedEvidenceAuthoring.test.tsx) | Tests source-bound Evidence drafts, authoring minima, exact omission, and presentation detachment. |
| [timelineRelatedEvidenceRecovery.test.tsx](timelineRelatedEvidenceRecovery.test.tsx) | Tests duplicate reservation, independent creation/link receipts, and link-only recovery. |
| [useEvidenceWorkbookBindings.test.tsx](useEvidenceWorkbookBindings.test.tsx) | Tests latest preview intent and stale preview rejection across subject and authority changes. |
