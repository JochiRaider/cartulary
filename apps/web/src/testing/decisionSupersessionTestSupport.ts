import type { SupersedeRecordRequest } from "@cartulary/protocol-ts/http";
import type { RecordChangedPayload } from "../workbook/collaboration/workbookCollaborationMessages";
import {
  type DecisionAuthority,
  type DecisionSupersessionReview,
  decisionViewId,
  reviewedDecision,
} from "../workbook/features/coordination/decisionSupersessionModel";
import type { DecisionSupersessionReceipt } from "../workbook/features/coordination/decisionSupersessionOperation";
import type { WorkbookQueryRow } from "../workbook/query/WorkbookQueryRow";

export const decisionAuthority: DecisionAuthority = {
  actorId: "00000000-0000-4000-8000-000000000010",
  sessionIdentity: "decision-session",
  incidentId: "00000000-0000-4000-8000-000000000001",
  role: "reviewer",
  closed: false,
};
export const decisionTargetId = "00000000-0000-4000-8000-000000000420";
export const decisionReplacementId = "00000000-0000-4000-8000-000000000421";
export function decisionRow(
  id = decisionTargetId,
  status = "proposed",
  version = 4,
): WorkbookQueryRow {
  return {
    record_id: id,
    row_version: version,
    cells: {
      "decision.summary": { value: "Duplicate summary" },
      "decision.status": { value: status },
      "decision.owner_user_id": { value: decisionAuthority.actorId },
      "decision.decided_at": { value: "2026-09-10T12:00:00Z" },
      "decision.is_superseded": { value: status === "superseded" },
      "decision.supersedes_record_id": { value: null },
    },
  };
}
export function decisionReview(
  status = "proposed",
  generation = 1,
): DecisionSupersessionReview {
  return {
    target: reviewedDecision(
      decisionRow(decisionTargetId, status),
      decisionAuthority.incidentId,
    ),
    replacement: reviewedDecision(
      decisionRow(decisionReplacementId, "approved", 6),
      decisionAuthority.incidentId,
    ),
    reason: "Later evidence",
    authority: decisionAuthority,
    authorityGeneration: generation,
    originSurface: "decisions",
    lifecycleKey: "selected-target",
  };
}
export function decisionReceipt(
  status = "proposed",
): DecisionSupersessionReceipt {
  return {
    view_schema_id: "cartulary.view.decisions.v1",
    change_set_id: "00000000-0000-4000-8000-000000000511",
    target_record_id: decisionTargetId,
    superseding_record_id: decisionReplacementId,
    target_row_version: 5,
    superseding_row_version: 7,
    target_status: status === "executed" ? "executed" : "superseded",
    reason: "Later evidence",
  };
}

export function decisionRequest(clientTxnId: string): SupersedeRecordRequest {
  return {
    base_row_version: 4,
    client_txn_id: clientTxnId,
    replacement_record_id: decisionReplacementId,
    reason: "Later evidence",
  };
}

export function decisionRecordChanged(): RecordChangedPayload {
  return {
    record_id: decisionTargetId,
    row_version: 9,
    change_set_id: "change",
    client_txn_id: "external",
    actor_user_id: decisionAuthority.actorId,
    changed_field_keys: ["decision.status"],
    affected_views: [
      {
        view_schema_id: decisionViewId,
        change_kind: "patch" as const,
        patch_cells: {
          record_id: decisionTargetId,
          row_version: 9,
          cells: { "decision.status": { value: "approved" } },
        },
      },
    ],
  };
}
