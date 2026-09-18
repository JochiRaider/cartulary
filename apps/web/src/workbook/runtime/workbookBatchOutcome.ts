import { getViewContract } from "@cartulary/view-contracts";
import type { RecoveryAttention } from "../../shared/workbookRecoveryNavigation";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import type { WorkbookBatchEntry } from "./workbookBatchOperation";

export type WorkbookBatchOutcome = {
  readonly kind:
    | "waiting"
    | "applying"
    | "uncertain"
    | "rejected"
    | "saved"
    | "saved_with_conflicts"
    | "conflicts_only"
    | "no_op";
  readonly label: string;
  readonly summary: string;
  readonly detail: string;
  readonly requestedScope: string;
  readonly originalInput: string;
  readonly inputPreview: string;
  readonly originalConflictCount: number;
  readonly unresolvedConflictCount: number;
  readonly returnedRecordCount: number;
  readonly attention: RecoveryAttention;
  readonly refresh: "none" | "pending" | "refreshing" | "required";
  readonly canReview: boolean;
};

/** One fact projection for list and detail. Receipt rows never inventory a change set. */
export function workbookBatchOutcome(
  entry: WorkbookBatchEntry,
  unresolvedConflictCount: number,
): WorkbookBatchOutcome {
  const { plan, receipt } = entry;
  const label =
    plan.operation === "pasteWorkbookClipboard"
      ? "Paste"
      : plan.request.kind === "fill_down_v1"
        ? "Fill"
        : plan.request.kind === "clear_cells_v1"
          ? "Clear contents"
          : "Tag assignment";
  const contract = getViewContract(plan.request.view_schema_id);
  const fieldLabel = (key: string) => contract?.fieldMap[key]?.label ?? "Field";
  const originalInput =
    plan.operation === "pasteWorkbookClipboard"
      ? plan.request.clipboard_text
      : plan.request.kind === "clear_cells_v1"
        ? `Clear ${plan.request.field_keys?.map(fieldLabel).join(", ") ?? "selected fields"}`
        : plan.request.kind === "fill_down_v1"
          ? (plan.request.value ?? "")
          : (plan.request.tag_name ?? "");
  const previewSource = Array.from(
    originalInput.slice(0, 256).replace(/\s+/gu, " ").trim(),
  );
  const inputPreview =
    previewSource.slice(0, 72).join("") +
    (previewSource.length > 72 || originalInput.length > 256 ? "…" : "");
  const targets = plan.request.targets;
  const creates = targets.filter(
    (target) => "kind" in target && target.kind === "create",
  ).length;
  const fields =
    plan.operation === "applyWorkbookBulkMutation"
      ? plan.request.kind === "clear_cells_v1"
        ? ` · ${plan.request.field_keys?.length ?? 0} field${plan.request.field_keys?.length === 1 ? "" : "s"}`
        : plan.request.kind === "fill_down_v1"
          ? ` · ${fieldLabel(plan.request.field_key ?? "")}`
          : ""
      : "";
  const requestedScope = `Requested: ${targets.length} target row${targets.length === 1 ? "" : "s"}${creates ? ` (${creates} new)` : ""}${fields}`;
  const originalConflictCount = receipt?.conflicts.length ?? 0;
  const returnedRecordCount = new Set(receipt?.rows.map((row) => row.record_id))
    .size;
  const kind: WorkbookBatchOutcome["kind"] = receipt
    ? receipt.changeSetId
      ? unresolvedConflictCount
        ? "saved_with_conflicts"
        : "saved"
      : originalConflictCount
        ? "conflicts_only"
        : "no_op"
    : entry.phase === "uncertain"
      ? "uncertain"
      : entry.phase === "rejected"
        ? "rejected"
        : entry.phase === "waiting"
          ? "waiting"
          : "applying";
  const summary = {
    waiting: "Awaiting earlier work",
    applying: "Applying changes",
    uncertain: "Outcome uncertain",
    rejected: "Action rejected",
    saved: "Saved changes",
    saved_with_conflicts: "Saved changes; conflicts need review",
    conflicts_only: "No batch changes saved",
    no_op: "No changes were needed",
  }[kind];
  const refresh =
    receipt && entry.reconciliation !== "complete"
      ? entry.reconciliation
      : "none";
  const completed =
    receipt !== null && refresh === "none" && unresolvedConflictCount === 0;
  const detail =
    kind === "uncertain"
      ? "The result is unknown. Retry checks the original action without duplicating accepted work."
      : kind === "rejected"
        ? (entry.failure?.message ??
          "The batch was rejected. Review the original range.")
        : `${summary}.${returnedRecordCount ? ` ${returnedRecordCount} record${returnedRecordCount === 1 ? "" : "s"} returned.` : ""}${originalConflictCount ? ` ${originalConflictCount} original conflict${originalConflictCount === 1 ? "" : "s"}; ${unresolvedConflictCount} unresolved.` : unresolvedConflictCount ? ` ${unresolvedConflictCount} conflicts remain unresolved.` : ""}`;
  return {
    kind,
    label,
    summary,
    detail,
    requestedScope,
    originalInput,
    inputPreview,
    originalConflictCount,
    unresolvedConflictCount,
    returnedRecordCount,
    refresh,
    attention: completed
      ? "completed"
      : receipt ||
          unresolvedConflictCount ||
          kind === "uncertain" ||
          kind === "rejected"
        ? "attention"
        : "progress",
    canReview: Boolean(
      receipt?.changeSetId &&
        returnedRecordCount &&
        receipt.viewSchemaId === timelineViewSchemaId,
    ),
  };
}
