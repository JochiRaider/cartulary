import { getViewContract } from "@cartulary/view-contracts";
import type { WorkbookRecoveryItem } from "../../shared/workbookRecoveryNavigation";
import type { WorkbookBatchSnapshot } from "./workbookBatchOperation";
import { workbookBatchOutcome } from "./workbookBatchOutcome";

export function workbookBatchRecoveryItems(
  state: WorkbookBatchSnapshot,
  conflictedBatches: ReadonlyMap<string, number>,
): readonly WorkbookRecoveryItem[] {
  if (!state.authority) return [];
  const items: WorkbookRecoveryItem[] = state.entries.map((entry, order) => {
    const outcome = workbookBatchOutcome(
      entry,
      conflictedBatches.get(entry.id) ?? 0,
    );
    return {
      id: entry.id,
      label: outcome.label,
      summary: `${outcome.summary} · ${outcome.requestedScope}${outcome.inputPreview ? ` · Input: ${outcome.inputPreview}` : ""}${outcome.refresh === "required" ? " · Refresh required" : ""}`,
      origin:
        getViewContract(entry.plan.request.view_schema_id)?.title ?? "Workbook",
      sheetRef: { kind: "view_schema", id: entry.plan.request.view_schema_id },
      attention: outcome.attention,
      refreshViews:
        outcome.refresh !== "none" ? [entry.plan.request.view_schema_id] : [],
      priority: outcome.unresolvedConflictCount ? 2 : 10,
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
