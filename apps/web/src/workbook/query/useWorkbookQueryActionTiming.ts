import { useCallback } from "react";
import { useWorkbookBrowsingRegistry } from "./WorkbookQueryBrowsingContext";

/** Acceptance belongs to the user command; a staged read or React updater is too late. */
export function useWorkbookQueryActionTiming(viewSchemaId: string) {
  const browsing = useWorkbookBrowsingRegistry();
  return useCallback(
    (operation: "sort" | "filter" | "group") => {
      performance.mark(`cartulary.workbook.${operation}_change_accepted`, {
        detail: {
          viewSchemaId,
          afterRevision:
            browsing.find(viewSchemaId)?.getSnapshot().resultRevision ?? 0,
        },
      });
    },
    [browsing, viewSchemaId],
  );
}
