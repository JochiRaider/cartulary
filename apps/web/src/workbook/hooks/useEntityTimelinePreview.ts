import { requireViewContract } from "@cartulary/view-contracts";
import { useCallback, useEffect, useRef, useState } from "react";
import { requireWorkbookSurfaceAcceptance } from "../collaboration/workbookSurfacePort";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import { workbookFailureLifecycle } from "../ports/WorkbookPortResult";
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
} from "../timeline/models/timelineRowModel";

const timelineContract = requireViewContract(timelineViewSchemaId);

export function useEntityTimelinePreview({
  entityType,
  viewQuery,
  onAuthorityUncertain,
}: {
  readonly entityType: "host" | "identity";
  readonly onAuthorityUncertain?: (() => void) | undefined;
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
    async (
      recordId: string,
      options?: { readonly requireAcceptance?: boolean },
    ) => {
      const request = beginLatestQuery(queryRuntimeRef);
      setTimelinePreviewRows([]);
      const result = await viewQuery.query({
        contract: timelineContract,
        queryState: emptyWorkbookQueryState(),
        signal: request.signal,
      });
      if (!request.isCurrent() || result.kind === "aborted") {
        if (options?.requireAcceptance)
          requireWorkbookSurfaceAcceptance({ kind: "aborted" });
        return;
      }
      if (result.kind === "rejected") {
        setTimelinePreviewRows([]);
        if (
          workbookFailureLifecycle(result.failure).kind ===
          "authority_unavailable"
        )
          onAuthorityUncertain?.();
        if (options?.requireAcceptance)
          requireWorkbookSurfaceAcceptance(result);
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
        if (options?.requireAcceptance)
          throw new Error("Timeline preview could not be verified");
        return;
      }
      if (request.isCurrent()) {
        setTimelinePreviewRows(previewRows);
      }
    },
    [entityType, viewQuery, onAuthorityUncertain],
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
