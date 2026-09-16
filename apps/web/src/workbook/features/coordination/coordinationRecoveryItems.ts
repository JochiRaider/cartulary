import { getViewContract } from "@cartulary/view-contracts";
import type { WorkbookRecoveryItem } from "../../../shared/workbookRecoveryNavigation";
import type { WorkbookCoordinationCreateOwner } from "./WorkbookCoordinationCreateOwner";

export function coordinationRecoveryItems(
  state: ReturnType<WorkbookCoordinationCreateOwner["getSnapshot"]>,
): readonly WorkbookRecoveryItem[] {
  if (!state.authority) return [];
  const items = new Map<string, WorkbookRecoveryItem>();
  const draft = state.draft;
  if (draft)
    items.set(String(draft.id), {
      id: String(draft.id),
      label: "Coordination draft",
      summary: state.preparing
        ? "Preparing submission"
        : state.needsReview
          ? "Review retained authoring"
          : "Draft retained",
      origin: draft.source?.label || "Coordination",
      sheetRef: draft.origin.sheetRef,
      attention: state.preparing
        ? "progress"
        : state.needsReview
          ? "attention"
          : "draft",
      order: draft.id,
    });
  for (const entry of state.entries) {
    if (entry.phase === "rejected" || entry.refresh === "complete") continue;
    const original = entry.attempt.review.draft;
    items.set(String(original.id), {
      id: String(original.id),
      label: "Coordination creation",
      summary: entry.receipt
        ? "Saved; refresh required"
        : entry.phase === "uncertain"
          ? "Outcome unconfirmed"
          : "Creating Coordination",
      origin:
        original.source?.label ||
        getViewContract("cartulary.view.lessons.v1")?.title ||
        "Coordination",
      sheetRef: original.origin.sheetRef,
      refreshViews: entry.receipt
        ? [
            original.target.viewSchemaId,
            ...(original.source ? [original.source.viewSchemaId] : []),
            ...entry.observations.flatMap((event) =>
              event.payload.affected_views.map((view) => view.view_schema_id),
            ),
          ]
        : [],
      attention:
        entry.phase === "uncertain" || entry.receipt ? "attention" : "progress",
      order: original.id,
    });
  }
  return [...items.values()];
}
