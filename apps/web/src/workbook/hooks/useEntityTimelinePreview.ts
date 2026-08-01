import { requireViewContract } from "@cartulary/view-contracts";
import { useCallback, useEffect, useRef, useState } from "react";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import type { WorkbookViewQueryPort } from "../query/WorkbookViewQueryPort";
import {
  abortLatestQuery,
  beginLatestQuery,
  type LatestQueryRuntime,
} from "../query/workbookLatestRequest";
import {
  normalizeTimelineFullRow,
  rowFromApi,
  type WorkbookRow,
} from "../timeline/models/workbookTimelineModel";

const timelineContract = requireViewContract(timelineViewSchemaId);

export function useEntityTimelinePreview({
  entityType,
  viewQuery,
}: {
  readonly entityType: "host" | "identity";
  readonly viewQuery: WorkbookViewQueryPort;
}) {
  const [timelinePreviewRows, setTimelinePreviewRows] = useState<WorkbookRow[]>(
    [],
  );
  const queryRuntimeRef = useRef<LatestQueryRuntime>({
    controller: null,
    sequence: 0,
  });

  const clearTimelinePreview = useCallback(() => {
    abortLatestQuery(queryRuntimeRef);
    setTimelinePreviewRows([]);
  }, []);

  const loadTimelinePreview = useCallback(
    async (recordId: string) => {
      const request = beginLatestQuery(queryRuntimeRef);
      setTimelinePreviewRows([]);
      const result = await viewQuery.query({
        contract: timelineContract,
        queryState: emptyWorkbookQueryState(),
        signal: request.signal,
      });
      if (!request.isCurrent() || result.kind === "aborted") {
        return;
      }
      if (result.kind === "rejected") {
        setTimelinePreviewRows([]);
        return;
      }
      const draftKey = entityType === "host" ? "hostRefs" : "identityRefs";
      let previewRows: WorkbookRow[];
      try {
        previewRows = result.value.rows
          .map((row, index) =>
            rowFromApi(
              normalizeTimelineFullRow(
                row,
                `timeline preview query rows[${index}]`,
              ),
            ),
          )
          .filter((row) =>
            row.collectionValues[draftKey].some(
              (item) => item.resolvedRecordId === recordId,
            ),
          );
      } catch {
        setTimelinePreviewRows([]);
        return;
      }
      if (request.isCurrent()) {
        setTimelinePreviewRows(previewRows);
      }
    },
    [entityType, viewQuery],
  );

  useEffect(
    () => () => {
      abortLatestQuery(queryRuntimeRef);
    },
    [],
  );

  return {
    clearTimelinePreview,
    loadTimelinePreview,
    timelinePreviewRows,
  };
}
