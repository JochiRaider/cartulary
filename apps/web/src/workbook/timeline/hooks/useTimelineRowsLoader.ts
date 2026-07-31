import type { Dispatch, SetStateAction } from "react";
import { useCallback, useEffect, useRef } from "react";
import { flushSync } from "react-dom";
import type { WorkbookQueryState } from "../../models/workbookQuery";
import type {
  WorkbookPendingQueueRuntime,
  WorkbookPendingRefreshBlockScope,
  WorkbookPendingSavesRefs,
} from "../../runtime/workbookPendingReplayRuntime";
import { reconcileWorkbookRecordRows } from "../../utils/workbookRowReconciliation";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import type {
  PendingReplayRuntimeMeta,
  TimelineMutableRef,
} from "../models/timelineControllerPorts";
import type {
  TimelineSourceRecordEvidence,
  TimelineSourceRecordRequirement,
} from "../models/timelineViewportContinuityModel";
import { timelineSourceRecordRequirementSatisfied } from "../models/timelineViewportContinuityModel";
import type { DismissedMention } from "../models/workbookMentionChips";
import { decideWorkbookRecordFreshness } from "../models/workbookRecordFreshness";
import {
  createDraftRow,
  inputFocusKey,
  type WorkbookRow,
} from "../models/workbookTimelineModel";
import type { TimelineViewQueryPort } from "../ports/TimelineViewQueryPort";

export type LoadRowsOptions = {
  afterProjectionCommit?: () => void;
  showLoading: boolean;
  freshnessRetryDepth?: number;
  sourceRecordRequirement?: TimelineSourceRecordRequirement;
  viewportContinuityToken?: number;
};

const maxTimelineFreshnessRetryDepth = 2;

export function useTimelineRowsLoader({
  acceptCommittedTimelineRows,
  advanceViewportContinuity,
  beginRefreshInFlight,
  beginTimelineRowsLoad,
  committedRowsChangedSince,
  currentCommittedTimelineRow,
  finishRefreshInFlight,
  failViewportContinuity,
  hasLoadedRows,
  isCurrentLoadSequence,
  knownTimelineRowVersion,
  loadRowsRef,
  markRowsLoaded,
  nextDraftIndex,
  onIncidentAccessLost,
  pendingSavesRefsRef,
  pruneAutoResolutionNoticesForRows,
  pruneDismissedMentionsForRow,
  publishSaveStatePresentation,
  queryState,
  rowsRef,
  editorDraftRegistry,
  setDismissedMentionsByRow,
  setIsInitialLoading,
  setIsRefreshing,
  setLoadAccessLost,
  setLoadError,
  setRefreshError,
  setRows,
  timelineViewQuery,
}: {
  readonly acceptCommittedTimelineRows: (rows: readonly WorkbookRow[]) => void;
  readonly advanceViewportContinuity: (
    token?: number,
    options?: { sourceRecord?: TimelineSourceRecordEvidence },
  ) => void;
  readonly beginRefreshInFlight: (
    scope: WorkbookPendingRefreshBlockScope,
  ) => void;
  readonly beginTimelineRowsLoad: () => {
    readonly queryStartEpoch: number;
    readonly requestSequence: number;
  };
  readonly committedRowsChangedSince: (epoch: number) => boolean;
  readonly currentCommittedTimelineRow: (
    recordId: string,
  ) => WorkbookRow | null;
  readonly finishRefreshInFlight: (
    scope: WorkbookPendingRefreshBlockScope,
  ) => void;
  readonly failViewportContinuity: (token: number) => void;
  readonly hasLoadedRows: () => boolean;
  readonly isCurrentLoadSequence: (requestSequence: number) => boolean;
  readonly knownTimelineRowVersion: (
    recordId: string,
  ) => number | null | undefined;
  readonly loadRowsRef: TimelineMutableRef<
    (options: LoadRowsOptions) => Promise<void>
  >;
  readonly markRowsLoaded: () => void;
  readonly nextDraftIndex: () => number;
  readonly onIncidentAccessLost?: (() => void) | undefined;
  readonly pendingSavesRefsRef: TimelineMutableRef<
    WorkbookPendingSavesRefs<PendingReplayRuntimeMeta>
  >;
  readonly pruneAutoResolutionNoticesForRows: (
    rows: readonly WorkbookRow[],
  ) => void;
  readonly pruneDismissedMentionsForRow: (
    dismissedMentionsByRow: Record<string, DismissedMention[]>,
    row: WorkbookRow,
  ) => Record<string, DismissedMention[]>;
  readonly publishSaveStatePresentation: (
    pending: WorkbookPendingQueueRuntime<PendingReplayRuntimeMeta>,
  ) => void;
  readonly queryState: WorkbookQueryState;
  readonly rowsRef: TimelineMutableRef<WorkbookRow[]>;
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
  readonly setDismissedMentionsByRow: Dispatch<
    SetStateAction<Record<string, DismissedMention[]>>
  >;
  readonly setIsInitialLoading: (loading: boolean) => void;
  readonly setIsRefreshing: (refreshing: boolean) => void;
  readonly setLoadAccessLost: (lost: boolean) => void;
  readonly setLoadError: (message: string | null) => void;
  readonly setRefreshError: (message: string | null) => void;
  readonly setRows: Dispatch<SetStateAction<WorkbookRow[]>>;
  readonly timelineViewQuery: TimelineViewQueryPort;
}) {
  const activeQueryRef = useRef<AbortController | null>(null);

  useEffect(
    () => () => {
      activeQueryRef.current?.abort();
      activeQueryRef.current = null;
    },
    [],
  );

  const refreshTimelineRowsAfterStaleResult = useCallback(
    async (options: LoadRowsOptions) => {
      const nextDepth = (options.freshnessRetryDepth ?? 0) + 1;
      if (nextDepth > maxTimelineFreshnessRetryDepth) {
        if (options.viewportContinuityToken !== undefined) {
          failViewportContinuity(options.viewportContinuityToken);
        }
        return false;
      }

      const refreshScope: WorkbookPendingRefreshBlockScope = { kind: "all" };
      beginRefreshInFlight(refreshScope);
      try {
        const retryOptions: LoadRowsOptions = {
          showLoading: false,
          freshnessRetryDepth: nextDepth,
        };
        if (options.afterProjectionCommit !== undefined) {
          retryOptions.afterProjectionCommit = options.afterProjectionCommit;
        }
        if (options.sourceRecordRequirement !== undefined) {
          retryOptions.sourceRecordRequirement =
            options.sourceRecordRequirement;
        }
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
      failViewportContinuity,
      finishRefreshInFlight,
      loadRowsRef,
    ],
  );

  const freshTimelineRowsForQueryResult = useCallback(
    (
      incomingRows: readonly WorkbookRow[],
      sourceRecordRequirement?: TimelineSourceRecordRequirement,
    ) => {
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
      if (
        sourceRecordRequirement !== undefined &&
        !rows.some(
          (row) =>
            row.recordId !== null &&
            row.rowVersion !== null &&
            timelineSourceRecordRequirementSatisfied(sourceRecordRequirement, {
              recordId: row.recordId,
              rowVersion: row.rowVersion,
            }),
        )
      ) {
        hasStaleRows = true;
      }
      return { hasStaleRows, rows };
    },
    [currentCommittedTimelineRow, knownTimelineRowVersion],
  );

  const settleProjectionObligationsFromCurrentRows = useCallback(
    (options: LoadRowsOptions) => {
      const requirement = options.sourceRecordRequirement;
      if (requirement === undefined) {
        return false;
      }
      const sourceRow = rowsRef.current.find(
        (row) =>
          row.recordId === requirement.recordId && row.rowVersion !== null,
      );
      if (
        sourceRow?.recordId === null ||
        sourceRow?.recordId === undefined ||
        sourceRow.rowVersion === null ||
        !timelineSourceRecordRequirementSatisfied(requirement, {
          recordId: sourceRow.recordId,
          rowVersion: sourceRow.rowVersion,
        })
      ) {
        return false;
      }

      // A newer concurrent query may already have committed the exact source
      // version this mutation requires. Join that projection instead of
      // dropping the mutation's semantic follow-up under last-query-wins
      // scheduling.
      flushSync(() => {
        options.afterProjectionCommit?.();
      });
      advanceViewportContinuity(options.viewportContinuityToken, {
        sourceRecord: {
          recordId: sourceRow.recordId,
          rowVersion: sourceRow.rowVersion,
        },
      });
      return true;
    },
    [advanceViewportContinuity, rowsRef],
  );

  const loadRows = useCallback(
    async (options: LoadRowsOptions) => {
      const { queryStartEpoch, requestSequence } = beginTimelineRowsLoad();

      if (options.showLoading && !hasLoadedRows()) {
        setIsInitialLoading(true);
      }
      if (hasLoadedRows()) {
        setIsRefreshing(true);
        setRefreshError(null);
      } else {
        setLoadAccessLost(false);
        setLoadError(null);
      }

      activeQueryRef.current?.abort();
      const queryAbort = new AbortController();
      activeQueryRef.current = queryAbort;
      const result = await timelineViewQuery.query({
        queryState,
        signal: queryAbort.signal,
      });
      if (activeQueryRef.current === queryAbort) {
        activeQueryRef.current = null;
      }

      if (!isCurrentLoadSequence(requestSequence)) {
        if (settleProjectionObligationsFromCurrentRows(options)) {
          return;
        }
        if (options.sourceRecordRequirement !== undefined) {
          await refreshTimelineRowsAfterStaleResult(options);
        }
        return;
      }

      if (committedRowsChangedSince(queryStartEpoch)) {
        setIsRefreshing(false);
        if (settleProjectionObligationsFromCurrentRows(options)) {
          return;
        }
        const refreshed = await refreshTimelineRowsAfterStaleResult(options);
        if (!refreshed && !hasLoadedRows()) {
          setIsInitialLoading(false);
        }
        return;
      }

      if (result.kind === "aborted") {
        return;
      }

      if (result.kind === "rejected") {
        setIsRefreshing(false);
        if (options.viewportContinuityToken !== undefined) {
          failViewportContinuity(options.viewportContinuityToken);
        }
        const message = result.failure.message;
        if (
          result.failure.kind === "authentication_required" ||
          result.failure.kind === "authorization_lost" ||
          result.failure.kind === "stale_target"
        ) {
          editorDraftRegistry.clearAll();
          rowsRef.current = [];
          setRows([]);
          setLoadAccessLost(true);
          setLoadError(message);
          setIsInitialLoading(false);
          onIncidentAccessLost?.();
          return;
        }
        if (hasLoadedRows()) {
          setRefreshError(message);
        } else {
          setLoadAccessLost(false);
          setLoadError(message);
          setIsInitialLoading(false);
        }
        return;
      }

      const incomingRows = [...result.value.rows];
      setLoadAccessLost(false);
      const incomingFreshness = freshTimelineRowsForQueryResult(
        incomingRows,
        options.sourceRecordRequirement,
      );
      if (incomingFreshness.hasStaleRows) {
        const refreshed = await refreshTimelineRowsAfterStaleResult(options);
        if (refreshed) {
          return;
        }
        setIsRefreshing(false);
        setRefreshError(
          options.sourceRecordRequirement === undefined
            ? "Timeline projection did not converge."
            : `Timeline row ${options.sourceRecordRequirement.recordId} did not reach version ${options.sourceRecordRequirement.minimumRowVersion}.`,
        );
        return;
      }
      const { committedRows, rows: hydratedRows } =
        reconcileCommittedRowsWithLocalDrafts({
          currentRows: rowsRef.current,
          incomingRows: incomingFreshness.rows,
          materializeRow: editorDraftRegistry.materializeRow,
          nextDraftIndex,
        });
      acceptCommittedTimelineRows(committedRows);
      editorDraftRegistry.retainRows(
        new Set(hydratedRows.map((row) => row.key)),
      );
      rowsRef.current = hydratedRows;
      const commitProjectionAndFollowUps = () => {
        setRows(hydratedRows);
        options.afterProjectionCommit?.();
        setDismissedMentionsByRow((current) => {
          let next = current;
          for (const row of committedRows) {
            if (row.recordId === null) {
              continue;
            }
            next = pruneDismissedMentionsForRow(next, row);
          }
          return next;
        });
        pruneAutoResolutionNoticesForRows(committedRows);
        publishSaveStatePresentation(
          pendingSavesRefsRef.current.pendingQueueRef.current,
        );
        markRowsLoaded();
        setIsRefreshing(false);
        setLoadError(null);
        setRefreshError(null);
        setIsInitialLoading(false);
      };
      if (options.viewportContinuityToken === undefined) {
        commitProjectionAndFollowUps();
      } else {
        // Continuity-bearing callers restore focus as soon as loadRows
        // resolves. Commit the authoritative row tree and every same-surface
        // follow-up before advancing the continuity generation.
        flushSync(() => {
          commitProjectionAndFollowUps();
        });
      }
      const committedSourceRecord =
        options.sourceRecordRequirement === undefined
          ? undefined
          : committedRows.find(
              (row) =>
                row.recordId === options.sourceRecordRequirement?.recordId,
            );
      const sourceRecord =
        committedSourceRecord?.recordId === null ||
        committedSourceRecord?.recordId === undefined ||
        committedSourceRecord.rowVersion === null
          ? undefined
          : ({
              recordId: committedSourceRecord.recordId,
              rowVersion: committedSourceRecord.rowVersion,
            } satisfies TimelineSourceRecordEvidence);
      advanceViewportContinuity(
        options.viewportContinuityToken,
        sourceRecord === undefined ? {} : { sourceRecord },
      );
    },
    [
      acceptCommittedTimelineRows,
      advanceViewportContinuity,
      beginTimelineRowsLoad,
      committedRowsChangedSince,
      failViewportContinuity,
      freshTimelineRowsForQueryResult,
      hasLoadedRows,
      isCurrentLoadSequence,
      markRowsLoaded,
      nextDraftIndex,
      onIncidentAccessLost,
      pendingSavesRefsRef,
      pruneAutoResolutionNoticesForRows,
      pruneDismissedMentionsForRow,
      publishSaveStatePresentation,
      queryState,
      refreshTimelineRowsAfterStaleResult,
      rowsRef,
      editorDraftRegistry,
      settleProjectionObligationsFromCurrentRows,
      setDismissedMentionsByRow,
      setIsInitialLoading,
      setIsRefreshing,
      setLoadAccessLost,
      setLoadError,
      setRefreshError,
      setRows,
      timelineViewQuery,
    ],
  );

  loadRowsRef.current = loadRows;

  return { loadRows };
}

function reconcileCommittedRowsWithLocalDrafts({
  currentRows,
  incomingRows,
  materializeRow,
  nextDraftIndex,
}: {
  readonly currentRows: WorkbookRow[];
  readonly incomingRows: WorkbookRow[];
  readonly materializeRow: (row: WorkbookRow) => WorkbookRow;
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
  const committedRows = reconcileWorkbookRecordRows(
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
    return materializeRow(rowWithLocalState);
  });
  const localDraftRows = currentRows
    .filter((row) => row.recordId === null)
    .map(materializeRow);

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
