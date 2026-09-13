import { boundedRead } from "../../services/asyncObservation";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import type { WorkbookAuthoringReadPort } from "../ports/WorkbookAuthoringReadPort";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";

/** Resolve original identity through the public view, independently of active filters. */
export function readWorkbookAuthoringRecord(
  reader: WorkbookAuthoringReadPort,
  viewSchemaId: string,
  recordId: string,
  signal: AbortSignal,
): Promise<WorkbookQueryRow | null> {
  return boundedRead(async (observed) => {
    let cursor: string | null = null;
    const visited = new Set<string>();
    do {
      const result = await reader.page({
        viewSchemaId,
        cursor,
        signal: observed,
        queryState: emptyWorkbookQueryState(),
      });
      if (observed.aborted || result.kind !== "accepted")
        throw new Error(
          "Current records could not be verified. Retry the read.",
        );
      const page = result.value;
      const row = page.candidates.find(
        (candidate) => candidate.recordId === recordId,
      )?.row;
      if (row) return row;
      if (!page.hasMore) return null;
      cursor = page.nextCursor;
      if (!cursor || visited.has(cursor))
        throw new Error("Record paging changed. Retry the read.");
      visited.add(cursor);
    } while (!observed.aborted);
    throw new Error("Record read interrupted.");
  }, signal);
}
