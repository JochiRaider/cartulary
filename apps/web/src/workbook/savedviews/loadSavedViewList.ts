import {
  acceptWorkbookSavedViewPage,
  startWorkbookSavedViewPagination,
} from "../models/workbookSavedViewPaginationMachine";
import type { SavedViewResource } from "../models/workbookSavedViews";
import type {
  SavedViewResult,
  WorkbookSavedViewPort,
} from "../ports/WorkbookSavedViewPort";

/** A list observation is published only after every page has passed validation. */
export async function loadSavedViewList(
  port: WorkbookSavedViewPort,
  signal: AbortSignal,
): Promise<SavedViewResult<readonly SavedViewResource[]>> {
  let pagination = startWorkbookSavedViewPagination(0);
  while (!signal.aborted) {
    const result = await port.listPage({
      cursorToken: pagination.nextCursor,
      limit: 100,
      signal,
    });
    if (result.kind !== "accepted") return result;
    const plan = acceptWorkbookSavedViewPage(pagination, result.value);
    if (plan.kind === "invalid")
      return {
        kind: "rejected",
        failure: { kind: "invalid_contract", message: plan.message },
      };
    if (plan.kind === "complete")
      return { kind: "accepted", value: plan.savedViews };
    pagination = plan.machine;
  }
  return {
    kind: "rejected",
    failure: {
      kind: "transport",
      message: "Saved-view list observation ended.",
    },
  };
}
