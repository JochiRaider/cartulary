import { getViewContract } from "@cartulary/view-contracts";
import type { WorkbookRecoveryItem } from "../../shared/workbookRecoveryNavigation";
import type { WorkbookBatchSnapshot } from "./workbookBatchOperation";

export function workbookBatchRecoveryItems(
  state: WorkbookBatchSnapshot,
  conflictedBatches: ReadonlySet<string>,
): readonly WorkbookRecoveryItem[] {
  if (!state.authority) return [];
  const items: WorkbookRecoveryItem[] = state.entries.map((entry, order) => {
    const label =
      entry.plan.operation === "pasteWorkbookClipboard"
        ? "Paste"
        : entry.plan.request.kind === "fill_down_v1"
          ? "Fill"
          : entry.plan.request.kind === "clear_cells_v1"
            ? "Clear contents"
            : "Tag assignment";
    const conflicted = conflictedBatches.has(entry.id);
    const completed =
      entry.phase === "acknowledged" &&
      entry.reconciliation === "complete" &&
      !conflicted;
    const original =
      entry.plan.operation === "pasteWorkbookClipboard"
        ? entry.plan.request.clipboard_text
        : entry.plan.request.kind === "fill_down_v1"
          ? (entry.plan.request.value ?? "")
          : entry.plan.request.kind === "clear_cells_v1"
            ? `Clear ${entry.plan.request.field_keys?.length ?? 0} fields to null`
            : (entry.plan.request.tag_name ?? "");
    const previewSource = Array.from(
      original.slice(0, 256).replace(/\s+/gu, " ").trim(),
    );
    const preview =
      previewSource.slice(0, 72).join("") +
      (previewSource.length > 72 || original.length > 256 ? "…" : "");
    const stateSummary = conflicted
      ? "Conflicts need review"
      : entry.phase === "uncertain"
        ? "Outcome unconfirmed"
        : entry.phase === "rejected"
          ? "Review rejected action"
          : completed
            ? "Complete"
            : entry.receipt
              ? "Saved; refresh required"
              : "Waiting or applying batch";
    return {
      id: entry.id,
      label,
      summary: `${stateSummary} · ${entry.plan.recordIds.length} record${entry.plan.recordIds.length === 1 ? "" : "s"}${preview ? ` · ${preview}` : ""}`,
      origin:
        getViewContract(entry.plan.request.view_schema_id)?.title ?? "Workbook",
      sheetRef: { kind: "view_schema", id: entry.plan.request.view_schema_id },
      attention: completed
        ? "completed"
        : conflicted ||
            entry.phase === "uncertain" ||
            entry.phase === "rejected" ||
            entry.receipt
          ? "attention"
          : "progress",
      refreshViews:
        entry.receipt && entry.reconciliation !== "complete"
          ? [entry.plan.request.view_schema_id]
          : [],
      priority: conflicted ? 2 : 10,
      order,
    };
  });
  if (state.admissionError)
    items.push({
      id: "admission",
      label: "Batch review",
      summary: "Review the original range",
      origin: "Workbook",
      sheetRef: null,
      attention: "attention",
      order: -1,
    });
  return items;
}
