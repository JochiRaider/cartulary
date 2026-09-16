import { useCallback, useMemo } from "react";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import type { WorkbookAuthoringReadPort } from "../ports/WorkbookAuthoringReadPort";
import type { WorkbookCandidateReader } from "../ports/WorkbookCandidateReadPort";
import { useWorkbookCandidateDiscovery } from "./useWorkbookCandidateDiscovery";

/** Finite implemented-surface inventory uses the same scoped read lifecycle. */
export function useWorkbookAuthoringInventory(
  reader: Pick<WorkbookAuthoringReadPort, "availableViews">,
  allowed: readonly string[],
  target: string,
  revision: number | string,
) {
  const allowedKey = JSON.stringify(allowed);
  const membershipOnly = allowed.every((view) => view === "incident_members");
  const read = useCallback<
    WorkbookCandidateReader<{ recordId: string; displayText: string }>
  >(
    async ({ signal, isCurrent }) => {
      const result = await reader.availableViews(signal);
      if (signal.aborted || isCurrent?.() === false) return { kind: "aborted" };
      if (result.kind !== "accepted") return result;
      const permitted = JSON.parse(allowedKey) as string[];
      return {
        kind: "accepted",
        value: {
          candidates: result.value
            .filter((view) => permitted.includes(view))
            .map((recordId) => ({ recordId, displayText: recordId })),
          hasMore: false,
          nextCursor: null,
        },
      };
    },
    [reader, allowedKey],
  );
  const query = useMemo(emptyWorkbookQueryState, []);
  const discovery = useWorkbookCandidateDiscovery(
    read,
    query,
    `${target}:surfaces`,
    revision,
    !membershipOnly,
  );
  const views: readonly string[] = membershipOnly
    ? allowed
    : (discovery.page?.candidates.map((item) => item.recordId) ?? []);
  return {
    discovery,
    membershipOnly,
    views,
  };
}
