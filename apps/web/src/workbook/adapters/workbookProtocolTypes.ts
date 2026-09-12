import type {
  AttachBlobToEvidenceRecordRequest,
  CollectionActionsV1,
  CreateObjectBlobSlotRequest,
  CreateRecordLinkedNoteRequest,
  CreateViewRowRequest,
  CreateViewRowResponse,
  MergeEntityRecordRequest,
  MergeEntityRecordResponse,
  PatchRecordRequest,
  QueryWorkbookViewRequest,
  ResolveEntityMentionResponse,
  ResolveRecordSameFieldConflictRequest,
} from "@cartulary/protocol-ts/http";

/**
 * Private exact protocol types for Workbook application code. Runtime protocol
 * access remains confined to adapters; owner logic consumes these type-only
 * projections and proves unknown values through the request decoders.
 */
export type WorkbookProtocolCollectionActions = CollectionActionsV1;
export type WorkbookProtocolAttachBlobRequest =
  AttachBlobToEvidenceRecordRequest;
export type WorkbookProtocolCreateObjectBlobSlotRequest =
  CreateObjectBlobSlotRequest;
export type WorkbookProtocolCreateLinkedNoteRequest =
  CreateRecordLinkedNoteRequest;
export type WorkbookProtocolCreateViewRowRequest = CreateViewRowRequest;
export type WorkbookProtocolPatchRecordRequest = PatchRecordRequest;
export type WorkbookProtocolQueryViewRequest = QueryWorkbookViewRequest;
export type WorkbookProtocolResolveConflictRequest =
  ResolveRecordSameFieldConflictRequest;

export type WorkbookProtocolMergeRequest = MergeEntityRecordRequest;
export type WorkbookProtocolMergeReceipt = MergeEntityRecordResponse["data"];

export type WorkbookProtocolMentionReceipt =
  ResolveEntityMentionResponse["data"];
export type WorkbookProtocolCreateViewRowReceipt = CreateViewRowResponse;
