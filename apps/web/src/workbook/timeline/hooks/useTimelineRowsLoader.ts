import { reconcileRecordRows } from "@cartulary/grid-adapter";
import type { ViewContract } from "@cartulary/view-contracts";
import type { Dispatch, SetStateAction } from "react";
import { startTransition, useCallback, useMemo } from "react";
import { apiPath } from "../../../services/browserApi";
import { fetchJSON, readEnvelope } from "../../../services/workbookApi";
import {
  buildQueryRequest,
  type WorkbookQueryState,
} from "../../models/workbookQuery";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { DismissedMention } from "../models/workbookMentionChips";
import {
  createDraftRow,
  decideWorkbookRecordFreshness,
  inputFocusKey,
  normalizeTimelineFullRow,
  rowFromApi,
  timelineScalarBindings,
  timelineScalarEditorSurfaces,
  validateTimelineViewSchemaId,
  type WorkbookRow,
} from "../models/workbookTimelineModel";
import type { PendingReplayRuntimeMeta } from "./useTimelinePendingReplayController";
import type {
  TimelineMutableRef,
  TimelinePendingQueueRuntime,
  TimelinePendingRefreshBlockScope,
  TimelinePendingSavesRefs,
} from "./useTimelinePendingSaves";

export type LoadRowsOptions = {
  showLoading: boolean;
  freshnessRetryDepth?: number;
  viewportContinuityToken?: number;
};

type WorkbookQueryEnvelope = {
  data: {
    incident_id: string;
    view_schema_id: string;
    rows: unknown[];
  };
};

const maxTimelineFreshnessRetryDepth = 2;

export function useTimelineRowsLoader({
  acceptCommittedTimelineRows,
  advanceViewportContinuity,
  apiBase,
  beginRefreshInFlight,
  beginTimelineRowsLoad,
  clearViewportContinuity,
  committedRowsChangedSince,
  currentCommittedTimelineRow,
  finishRefreshInFlight,
  hasLoadedRows,
  incidentId,
  isCurrentLoadSequence,
  knownTimelineRowVersion,
  loadRowsRef,
  markRowsLoaded,
  nextDraftIndex,
  pendingSavesRefsRef,
  pruneAutoResolutionNoticesForRows,
  pruneDismissedMentionsForRow,
  publishSaveStatePresentation,
  queryState,
  rowsRef,
  scalarDraftValuesRef,
  setDismissedMentionsByRow,
  setIsInitialLoading,
  setLoadError,
  setRefreshError,
  setRows,
  timelineContract,
}: {
  readonly acceptCommittedTimelineRows: (rows: readonly WorkbookRow[]) => void;
  readonly advanceViewportContinuity: (token?: number) => void;
  readonly apiBase?: string | undefined;
  readonly beginRefreshInFlight: (
    scope: TimelinePendingRefreshBlockScope,
  ) => void;
  readonly beginTimelineRowsLoad: () => {
    readonly queryStartEpoch: number;
    readonly requestSequence: number;
  };
  readonly clearViewportContinuity: (token: number) => void;
  readonly committedRowsChangedSince: (epoch: number) => boolean;
  readonly currentCommittedTimelineRow: (
    recordId: string,
  ) => WorkbookRow | null;
  readonly finishRefreshInFlight: (
    scope: TimelinePendingRefreshBlockScope,
  ) => void;
  readonly hasLoadedRows: () => boolean;
  readonly incidentId: string;
  readonly isCurrentLoadSequence: (requestSequence: number) => boolean;
  readonly knownTimelineRowVersion: (
    recordId: string,
  ) => number | null | undefined;
  readonly loadRowsRef: TimelineMutableRef<
    (options: LoadRowsOptions) => Promise<void>
  >;
  readonly markRowsLoaded: () => void;
  readonly nextDraftIndex: () => number;
  readonly pendingSavesRefsRef: TimelineMutableRef<
    TimelinePendingSavesRefs<PendingReplayRuntimeMeta>
  >;
  readonly pruneAutoResolutionNoticesForRows: (
    rows: readonly WorkbookRow[],
  ) => void;
  readonly pruneDismissedMentionsForRow: (
    dismissedMentionsByRow: Record<string, DismissedMention[]>,
    row: WorkbookRow,
  ) => Record<string, DismissedMention[]>;
  readonly publishSaveStatePresentation: (
    pending: TimelinePendingQueueRuntime<PendingReplayRuntimeMeta>,
  ) => void;
  readonly queryState: WorkbookQueryState;
  readonly rowsRef: TimelineMutableRef<WorkbookRow[]>;
  readonly scalarDraftValuesRef: TimelineMutableRef<Map<string, string>>;
  readonly setDismissedMentionsByRow: Dispatch<
    SetStateAction<Record<string, DismissedMention[]>>
  >;
  readonly setIsInitialLoading: (loading: boolean) => void;
  readonly setLoadError: (message: string | null) => void;
  readonly setRefreshError: (message: string | null) => void;
  readonly setRows: Dispatch<SetStateAction<WorkbookRow[]>>;
  readonly timelineContract: ViewContract;
}) {
  const queryPath = useMemo(
    () =>
      apiPath(
        apiBase,
        `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`,
      ),
    [apiBase, incidentId],
  );
  const queryBody = useMemo(
    () => JSON.stringify(buildQueryRequest(timelineContract, queryState)),
    [queryState, timelineContract],
  );

  const refreshTimelineRowsAfterStaleResult = useCallback(
    async (options: LoadRowsOptions) => {
      const nextDepth = (options.freshnessRetryDepth ?? 0) + 1;
      if (nextDepth > maxTimelineFreshnessRetryDepth) {
        if (options.viewportContinuityToken !== undefined) {
          clearViewportContinuity(options.viewportContinuityToken);
        }
        return false;
      }

      const refreshScope: TimelinePendingRefreshBlockScope = { kind: "all" };
      beginRefreshInFlight(refreshScope);
      try {
        const retryOptions: LoadRowsOptions = {
          showLoading: false,
          freshnessRetryDepth: nextDepth,
        };
        if (options.viewportContinuityToken !== undefined) {
          retryOptions.viewportContinuityToken =
            options.viewportContinuityToken;
        }
        await loadRowsRef.current(retryOptions);
      } finally {
        finishRefreshInFlight(refreshScope);
      }
      return true;
    },
    [
      beginRefreshInFlight,
      clearViewportContinuity,
      finishRefreshInFlight,
      loadRowsRef,
    ],
  );

  const freshTimelineRowsForQueryResult = useCallback(
    (incomingRows: readonly WorkbookRow[]) => {
      let hasStaleRows = false;
      const rows: WorkbookRow[] = [];
      for (const row of incomingRows) {
        if (
          row.recordId !== null &&
          decideWorkbookRecordFreshness(
            row,
            knownTimelineRowVersion(row.recordId),
          ).stale
        ) {
          hasStaleRows = true;
          const current = currentCommittedTimelineRow(row.recordId);
          if (current !== null) {
            rows.push(current);
          }
          continue;
        }
        rows.push(row);
      }
      return { hasStaleRows, rows };
    },
    [currentCommittedTimelineRow, knownTimelineRowVersion],
  );

  const loadRows = useCallback(
    async (options: LoadRowsOptions) => {
      const { queryStartEpoch, requestSequence } = beginTimelineRowsLoad();

      if (options.showLoading && !hasLoadedRows()) {
        setIsInitialLoading(true);
      }
      if (hasLoadedRows()) {
        setRefreshError(null);
      } else {
        setLoadError(null);
      }

      const result = await fetchJSON<WorkbookQueryEnvelope>(queryPath, {
        method: "POST",
        body: queryBody,
      });

      if (!isCurrentLoadSequence(requestSequence)) {
        return;
      }

      if (committedRowsChangedSince(queryStartEpoch)) {
        const refreshed = await refreshTimelineRowsAfterStaleResult(options);
        if (!refreshed && !hasLoadedRows()) {
          setIsInitialLoading(false);
        }
        return;
      }

      if (!result.ok) {
        if (options.viewportContinuityToken !== undefined) {
          clearViewportContinuity(options.viewportContinuityToken);
        }
        const message = "Timeline projection load failed.";
        if (hasLoadedRows()) {
          setRefreshError(message);
        } else {
          setLoadError(message);
          setIsInitialLoading(false);
        }
        return;
      }

      let incomingRows: WorkbookRow[];
      try {
        const envelope = readEnvelope<WorkbookQueryEnvelope>(result.payload);
        validateTimelineViewSchemaId(
          envelope.data.view_schema_id,
          "query response",
        );
        incomingRows = envelope.data.rows.map((row, index) =>
          rowFromApi(
            normalizeTimelineFullRow(row, `query response rows[${index}]`),
          ),
        );
      } catch {
        if (options.viewportContinuityToken !== undefined) {
          clearViewportContinuity(options.viewportContinuityToken);
        }
        const message = "Timeline projection load failed.";
        if (hasLoadedRows()) {
          setRefreshError(message);
        } else {
          setLoadError(message);
          setIsInitialLoading(false);
        }
        return;
      }
      const incomingFreshness = freshTimelineRowsForQueryResult(incomingRows);
      if (incomingFreshness.hasStaleRows) {
        const refreshed = await refreshTimelineRowsAfterStaleResult(options);
        if (refreshed) {
          return;
        }
      }
      const { committedRows, rows: hydratedRows } =
        reconcileCommittedRowsWithLocalDrafts({
          currentRows: rowsRef.current,
          incomingRows: incomingFreshness.rows,
          draftValueForFocusKey: (focusKey) =>
            scalarDraftValuesRef.current.get(focusKey),
          nextDraftIndex,
        });
      acceptCommittedTimelineRows(committedRows);
      startTransition(() => {
        rowsRef.current = hydratedRows;
        setRows(hydratedRows);
      });
      advanceViewportContinuity(options.viewportContinuityToken);
      setDismissedMentionsByRow((current) => {
        const next = { ...current };
        for (const row of committedRows) {
          if (row.recordId === null) {
            continue;
          }
          Object.assign(next, pruneDismissedMentionsForRow(next, row));
        }
        return next;
      });
      pruneAutoResolutionNoticesForRows(committedRows);
      publishSaveStatePresentation(
        pendingSavesRefsRef.current.pendingQueueRef.current,
      );
      markRowsLoaded();
      setLoadError(null);
      setRefreshError(null);
      setIsInitialLoading(false);
    },
    [
      acceptCommittedTimelineRows,
      advanceViewportContinuity,
      beginTimelineRowsLoad,
      clearViewportContinuity,
      committedRowsChangedSince,
      freshTimelineRowsForQueryResult,
      hasLoadedRows,
      isCurrentLoadSequence,
      markRowsLoaded,
      nextDraftIndex,
      pendingSavesRefsRef,
      pruneAutoResolutionNoticesForRows,
      pruneDismissedMentionsForRow,
      publishSaveStatePresentation,
      queryBody,
      queryPath,
      refreshTimelineRowsAfterStaleResult,
      rowsRef,
      scalarDraftValuesRef,
      setDismissedMentionsByRow,
      setIsInitialLoading,
      setLoadError,
      setRefreshError,
      setRows,
    ],
  );

  loadRowsRef.current = loadRows;

  return { loadRows };
}

function reconcileCommittedRowsWithLocalDrafts({
  currentRows,
  incomingRows,
  draftValueForFocusKey,
  nextDraftIndex,
}: {
  readonly currentRows: WorkbookRow[];
  readonly incomingRows: WorkbookRow[];
  readonly draftValueForFocusKey: (focusKey: string) => string | undefined;
  readonly nextDraftIndex: () => number;
}): {
  readonly committedRows: WorkbookRow[];
  readonly rows: WorkbookRow[];
} {
  const currentCommittedByRecordId = new Map(
    currentRows
      .filter(
        (row): row is WorkbookRow & { recordId: string } =>
          row.recordId !== null,
      )
      .map((row) => [row.recordId, row]),
  );
  const committedRows = reconcileRecordRows(
    currentRows.filter((row) => row.recordId !== null),
    incomingRows,
  ).map((row) => {
    let rowWithLocalState = row;
    if (row.recordId === null) {
      return row;
    }
    const current = currentCommittedByRecordId.get(row.recordId);
    if (current !== undefined) {
      if (
        current.pendingSignature !== null ||
        current.collectionDrafts.hostRefs !== "" ||
        current.collectionDrafts.identityRefs !== "" ||
        current.collectionDrafts.tags !== ""
      ) {
        rowWithLocalState = {
          ...rowWithLocalState,
          collectionDrafts: current.collectionDrafts,
          pendingSignature: current.pendingSignature,
        };
      }
    }
    return rowWithMaterializedScalarDrafts(
      rowWithLocalState,
      draftValueForFocusKey,
    );
  });
  const localDraftRows = currentRows
    .filter((row) => row.recordId === null)
    .map((row) => rowWithMaterializedScalarDrafts(row, draftValueForFocusKey));

  return {
    committedRows,
    rows: ensureDraftRowWithFreshIndex(
      [...committedRows, ...localDraftRows],
      nextDraftIndex,
    ).rows,
  };
}

function ensureDraftRowWithFreshIndex(
  rows: WorkbookRow[],
  nextDraftIndex: () => number,
): {
  rows: WorkbookRow[];
  draftSummaryKey: string | null;
} {
  if (rows.some((row) => row.recordId === null)) {
    return {
      rows,
      draftSummaryKey: null,
    };
  }

  const draftIndex = nextDraftIndex();
  return {
    rows: [...rows, createDraftRow(draftIndex)],
    draftSummaryKey: inputFocusKey(
      `draft-${draftIndex}`,
      "activitySynopsisText",
    ),
  };
}

function rowWithMaterializedScalarDrafts(
  row: WorkbookRow,
  draftValueForFocusKey: (focusKey: string) => string | undefined,
): WorkbookRow {
  let nextValues: WorkbookRow["values"] | null = null;
  for (const binding of timelineScalarBindings) {
    let draftValue: string | undefined;
    for (const surface of timelineScalarEditorSurfaces) {
      draftValue = draftValueForFocusKey(
        inputFocusKey(row.key, binding.key, surface),
      );
      if (draftValue !== undefined) {
        break;
      }
    }
    if (draftValue === undefined || draftValue === row.values[binding.key]) {
      continue;
    }
    nextValues ??= { ...row.values };
    nextValues[binding.key] = draftValue;
  }
  return nextValues === null ? row : { ...row, values: nextValues };
}
