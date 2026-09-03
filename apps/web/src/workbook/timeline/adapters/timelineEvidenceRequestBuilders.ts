import type {
  PatchRecordRequest,
  TimelineCreateRequest,
} from "@cartulary/protocol-ts/http";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookRow } from "../models/timelineRowModel";

export function buildAttachedEvidenceCreateRequest(
  evidenceRecordId: string,
  clientTxnId: string,
): TimelineCreateRequest {
  return {
    client_txn_id: clientTxnId,
    "timeline.attached_evidence_ids": {
      kind: "collection_actions_v1",
      actions: [
        {
          op: "add_record_ref",
          linked_record_id: evidenceRecordId,
        },
      ],
    },
  };
}

export function buildAttachedEvidencePatchRequest(
  row: WorkbookRow,
  evidenceRecordId: string,
  clientTxnId: string,
): PatchRecordRequest | null {
  if (row.rowVersion === null) return null;
  return {
    view_schema_id: timelineViewSchemaId,
    base_row_version: row.rowVersion,
    client_txn_id: clientTxnId,
    changes: [
      {
        field_key: "timeline.attached_evidence_ids",
        action_payload: {
          kind: "collection_actions_v1",
          actions: [
            {
              op: "add_record_ref",
              linked_record_id: evidenceRecordId,
            },
          ],
        },
      },
    ],
  };
}
