import type {
  ApplyWorkbookBulkMutationRequest,
  AttachBlobToEvidenceRecordResponse,
  CollectionActionsV1,
  CreateObjectBlobSlotResponse,
  CreateRecordLinkedNoteRequest,
  CreateRecordLinkedNoteResponse,
  CreateViewRowRequest,
  CreateViewRowResponse,
  ListNoteAssociationsResponse,
  MergeEntityRecordResponse,
  MutateNoteAssociationsRequest,
  MutateNoteAssociationsResponse,
  PasteWorkbookClipboardRequest,
  PatchRecordRequest,
  QueryWorkbookViewRequest,
  ResolveEntityMentionResponse,
  ResolveRecordSameFieldConflictRequest,
  ResolveRecordSameFieldConflictResponse,
} from "@cartulary/protocol-ts/http";

/**
 * Private exact protocol types for Workbook application code. Runtime protocol
 * access remains confined to adapters; owner logic consumes these type-only
 * projections and proves unknown values through the request decoders.
 */
export type WorkbookProtocolCollectionActions = CollectionActionsV1;
export type WorkbookProtocolAttachBlobReceipt =
  AttachBlobToEvidenceRecordResponse;
export type WorkbookProtocolBlobSlotReceipt = CreateObjectBlobSlotResponse;
export type WorkbookProtocolCreateLinkedNoteRequest =
  CreateRecordLinkedNoteRequest;
export type WorkbookProtocolCreateLinkedNoteReceipt =
  CreateRecordLinkedNoteResponse;
export type WorkbookProtocolCreateViewRowRequest = CreateViewRowRequest;
export type WorkbookProtocolPatchRecordRequest = PatchRecordRequest;
export type WorkbookProtocolQueryViewRequest = QueryWorkbookViewRequest;
export type WorkbookProtocolResolveConflictRequest =
  ResolveRecordSameFieldConflictRequest;
export type WorkbookProtocolMergeReceipt = MergeEntityRecordResponse["data"];

export type WorkbookProtocolMentionReceipt =
  ResolveEntityMentionResponse["data"];
export type WorkbookProtocolCreateViewRowReceipt = CreateViewRowResponse;
export type WorkbookProtocolConflictResolutionReceipt =
  ResolveRecordSameFieldConflictResponse;

export type WorkbookProtocolBulkRequest = ApplyWorkbookBulkMutationRequest;
export type WorkbookProtocolPasteRequest = PasteWorkbookClipboardRequest;

export type WorkbookProtocolNoteAssociationsRequest =
  MutateNoteAssociationsRequest;
export type WorkbookProtocolNoteAssociationsReceipt =
  MutateNoteAssociationsResponse;
export type WorkbookProtocolNoteAssociationsPage = ListNoteAssociationsResponse;
