import { requireViewContract } from "@cartulary/view-contracts";
import { taskViewId } from "../workbook/features/coordination/taskLifecycleModel";
import type {
  ExplicitPatchAuthority,
  ExplicitPatchIntent,
} from "../workbook/runtime/WorkbookExplicitPatchOwner";
export const taskAuthority: ExplicitPatchAuthority = {
  actorId: "00000000-0000-4000-8000-000000000010",
  sessionIdentity: "task-session",
  incidentId: "00000000-0000-4000-8000-000000000001",
  role: "editor",
  closed: false,
};
export const taskRecordId = "00000000-0000-4000-8000-000000000410";
export function taskRow(version = 7, status = "open") {
  return {
    record_id: taskRecordId,
    row_version: version,
    view_schema_id: taskViewId,
    cells: {
      ...Object.fromEntries(
        requireViewContract(taskViewId).fields.map((field) => [
          field.fieldKey,
          { value: null as unknown },
        ]),
      ),
      "task.title": { value: "Collect evidence" },
      "task.status": { value: status },
      "task.owner_user_id": { value: taskAuthority.actorId },
      "task.blocked_reason": { value: status === "blocked" ? "Waiting" : null },
      "task.completed_at": {
        value: status === "done" ? "2026-09-11T12:00:00Z" : null,
      },
    },
  };
}
export function taskIntent(): ExplicitPatchIntent {
  return {
    baseline: taskRow(),
    changes: [
      { field_key: "task.status", value: "blocked" },
      { field_key: "task.blocked_reason", value: "Waiting" },
    ],
    purpose: "task-lifecycle",
    sheetRef: { kind: "view_schema", id: taskViewId },
    surfaceLabel: "Task Requests",
  };
}
export function taskReceipt() {
  return {
    changeSetId: "00000000-0000-4000-8000-000000000510",
    row: taskRow(8, "blocked"),
    viewSchemaId: taskViewId,
  };
}
