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
  const [preview, setPreview] = useState<{
    recordId: string | null;
    rows: WorkbookRow[];
    hasData: boolean;
    state:
      | "initial_loading"
      | "ready"
      | "refreshing"
      | "stale_failure"
      | "unavailable";
    message: string | null;
  }>({
    recordId: null,
    rows: [],
    hasData: false,
    state: "initial_loading",
    message: null,
  });
  const queryRuntimeRef = useRef<LatestQueryRuntime>({
    controller: null,
    sequence: 0,
  });

  const clearTimelinePreview = useCallback(() => {
    abortLatestQuery(queryRuntimeRef);
    setPreview({
      recordId: null,
      rows: [],
      hasData: false,
      state: "initial_loading",
      message: null,
    });
  }, []);

  const loadTimelinePreview = useCallback(
    async (
      recordId: string,
      options?: { readonly requireAcceptance?: boolean },
    ) => {
      const request = beginLatestQuery(queryRuntimeRef);
      setPreview((current) =>
        current.recordId === recordId && current.hasData
          ? { ...current, state: "refreshing", message: null }
          : {
              recordId,
              rows: [],
              hasData: false,
              state: "initial_loading",
              message: null,
            },
      );
      const result = await viewQuery.query({
        contract: timelineContract,
        queryState: emptyWorkbookQueryState(),
        limit: 100,
        signal: request.signal,
      });
      if (!request.isCurrent() || result.kind === "aborted") {
        if (options?.requireAcceptance)
          requireWorkbookSurfaceAcceptance({ kind: "aborted" });
        return;
      }
      if (result.kind === "rejected") {
        setPreview((current) => ({
          ...current,
          state: current.hasData ? "stale_failure" : "unavailable",
          message: "Could not refresh the Timeline preview.",
        }));
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
        setPreview((current) => ({
          ...current,
          state: current.hasData ? "stale_failure" : "unavailable",
          message: "Could not refresh the Timeline preview.",
        }));
        if (options?.requireAcceptance)
          throw new Error("Timeline preview could not be verified");
        return;
      }
      if (request.isCurrent()) {
        setPreview({
          recordId,
          rows: previewRows,
          hasData: true,
          state: "ready",
          message: null,
        });
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
    timelinePreviewRows: preview.rows,
    timelinePreviewState: preview,
  };
}
