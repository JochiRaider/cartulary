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
| [evidenceFileOperation.ts](evidenceFileOperation.ts) | Single-file admission, captured upload/finalization attempts and full receipt contracts. |
| [EvidenceUploadSession.ts](EvidenceUploadSession.ts) | Retains the File and validated slot; confines the opaque session-bound target to one transfer. |
| [EvidenceFileFinalization.ts](EvidenceFileFinalization.ts) | Retains immutable finalization attempts and accepted receipts independently of presentation. |
| [WorkbookEvidenceAttachmentOwner.ts](WorkbookEvidenceAttachmentOwner.ts) | Existing-Evidence attachment, source review, exact recovery and read reconciliation. |
| [WorkbookTimelineFileOwner.ts](WorkbookTimelineFileOwner.ts) | Atomic file-backed Evidence creation and original Timeline association with separate receipts. |
| [timelineFileOperation.ts](timelineFileOperation.ts) | Timeline source identity, captured link receipts and the original draft promotion port. |
| [EvidenceAttachmentContext.ts](EvidenceAttachmentContext.ts) | Supplies both incident-retained file owners to grid and inspector presentations. |
| [EvidenceFileRecovery.tsx](EvidenceFileRecovery.tsx) | Compact local stage feedback and explicit review, resume, fresh upload, discard and refresh controls. |
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
| [evidenceFileRecovery.test.ts](evidenceFileRecovery.test.ts) | Stage uncertainty, complete receipts, source identity, retirement and current-authority recovery. |
| [timelineRelatedEvidenceAuthoring.test.tsx](timelineRelatedEvidenceAuthoring.test.tsx) | Tests source-bound Evidence drafts, authoring minima, exact omission, and presentation detachment. |
| [timelineRelatedEvidenceRecovery.test.tsx](timelineRelatedEvidenceRecovery.test.tsx) | Tests duplicate reservation, independent creation/link receipts, and link-only recovery. |
| [useEvidenceWorkbookBindings.test.tsx](useEvidenceWorkbookBindings.test.tsx) | Tests latest preview intent and stale preview rejection across subject and authority changes. |

## Ordinary creation seam

| File | Responsibility |
| --- | --- |
| [evidenceOrdinaryCreate.ts](evidenceOrdinaryCreate.ts) | Ordinary metadata minimum signals, explicit timestamp intent and upload/lifecycle boundaries. |

File work survives permitted presentation detachment in the incident/account
runtime. Accepted bytes and Evidence custody remain independent. No reload or
persistent-storage recovery is promised. Long transfers are separate from short
coordinated record writes. The Timeline mutation owner resolves the original
draft key through ordinary creation or screenshot-only creation; this directory
does not own a second Timeline creation queue.

RelatedEvidencePartyControl uses bounded Workbook discovery for two independent
single Party identities. Its parent retains labels only for currently referenced
collector/source identities. Candidate reads and retries never modify Party text,
create Evidence, replay creation, or run the separate linking operation.
