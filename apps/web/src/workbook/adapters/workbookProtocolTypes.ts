import type {
  CollectionActionsV1,
  CreateRecordLinkedNoteRequest,
  CreateViewRowRequest,
  PatchRecordRequest,
  QueryWorkbookViewRequest,
  ResolveRecordSameFieldConflictRequest,
} from "@cartulary/protocol-ts/http";

/**
 * Private exact protocol types for Workbook application code. Runtime protocol
 * access remains confined to adapters; owner logic consumes these type-only
 * projections and proves unknown values through the request decoders.
 */
export type WorkbookProtocolCollectionActions = CollectionActionsV1;
export type WorkbookProtocolCreateLinkedNoteRequest =
  CreateRecordLinkedNoteRequest;
export type WorkbookProtocolCreateViewRowRequest = CreateViewRowRequest;
export type WorkbookProtocolPatchRecordRequest = PatchRecordRequest;
export type WorkbookProtocolQueryViewRequest = QueryWorkbookViewRequest;
export type WorkbookProtocolResolveConflictRequest =
  ResolveRecordSameFieldConflictRequest;
