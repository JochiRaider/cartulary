import type { ViewContract } from "@cartulary/view-contracts";
import type { Dispatch, SetStateAction } from "react";
import { useCallback, useMemo } from "react";
import { flushSync } from "react-dom";
import { apiPath } from "../../../services/browserApi";
import {
  fetchWorkbookJSON,
  parseErrorMessage,
  readEnvelope,
  workbookLoadFailureIsAccessLoss,
} from "../../../services/workbookApi";
import {
  buildQueryRequest,
  type WorkbookQueryState,
} from "../../models/workbookQuery";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type {
  WorkbookPendingQueueRuntime,
  WorkbookPendingRefreshBlockScope,
  WorkbookPendingSavesRefs,
} from "../../runtime/workbookPendingReplayRuntime";
import { reconcileWorkbookRecordRows } from "../../utils/workbookRowReconciliation";
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
  normalizeTimelineFullRow,
  rowFromApi,
  timelineScalarBindings,
  timelineScalarEditorSurfaces,
  validateTimelineViewSchemaId,
  type WorkbookRow,
} from "../models/workbookTimelineModel";

export type LoadRowsOptions = {
  afterProjectionCommit?: () => void;
  showLoading: boolean;
  freshnessRetryDepth?: number;
  sourceRecordRequirement?: TimelineSourceRecordRequirement;
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
  committedRowsChangedSince,
  currentCommittedTimelineRow,
  finishRefreshInFlight,
  failViewportContinuity,
  hasLoadedRows,
  incidentId,
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
  scalarDraftValuesRef,
  setDismissedMentionsByRow,
  setIsInitialLoading,
  setIsRefreshing,
  setLoadError,
  setRefreshError,
  setRows,
  timelineContract,
}: {
  readonly acceptCommittedTimelineRows: (rows: readonly WorkbookRow[]) => void;
  readonly advanceViewportContinuity: (
    token?: number,
    options?: { sourceRecord?: TimelineSourceRecordEvidence },
  ) => void;
  readonly apiBase?: string | undefined;
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
  readonly scalarDraftValuesRef: TimelineMutableRef<Map<string, string>>;
  readonly setDismissedMentionsByRow: Dispatch<
    SetStateAction<Record<string, DismissedMention[]>>
  >;
  readonly setIsInitialLoading: (loading: boolean) => void;
  readonly setIsRefreshing: (refreshing: boolean) => void;
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
        setLoadError(null);
      }

      const result = await fetchWorkbookJSON<WorkbookQueryEnvelope>(queryPath, {
        method: "POST",
        body: queryBody,
      });

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

      if (!result.ok) {
        setIsRefreshing(false);
        if (options.viewportContinuityToken !== undefined) {
          failViewportContinuity(options.viewportContinuityToken);
        }
        const message = parseErrorMessage(result.payload);
        if (workbookLoadFailureIsAccessLoss(message)) {
          rowsRef.current = [];
          setRows([]);
          setLoadError(message);
          setIsInitialLoading(false);
          onIncidentAccessLost?.();
          return;
        }
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
        setIsRefreshing(false);
        if (options.viewportContinuityToken !== undefined) {
          failViewportContinuity(options.viewportContinuityToken);
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
          draftValueForFocusKey: (focusKey) =>
            scalarDraftValuesRef.current.get(focusKey),
          nextDraftIndex,
        });
      acceptCommittedTimelineRows(committedRows);
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
      queryBody,
      queryPath,
      refreshTimelineRowsAfterStaleResult,
      rowsRef,
      scalarDraftValuesRef,
      settleProjectionObligationsFromCurrentRows,
      setDismissedMentionsByRow,
      setIsInitialLoading,
      setIsRefreshing,
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
