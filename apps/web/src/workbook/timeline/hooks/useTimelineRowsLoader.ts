import { requireViewContract } from "@cartulary/view-contracts";
import { useCallback, useEffect, useRef } from "react";
import {
  requireWorkbookSurfaceAcceptance,
  WorkbookSurfaceRefreshError,
} from "../../collaboration/workbookSurfacePort";
import type { WorkbookQueryState } from "../../models/workbookQuery";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { workbookFailureLifecycle } from "../../ports/WorkbookPortResult";
import {
  useWorkbookBrowsingRead,
  useWorkbookQueryBrowser,
} from "../../query/WorkbookQueryBrowsingContext";
import type { WorkbookViewQueryPort } from "../../query/WorkbookViewQueryPort";
import {
  abortLatestQuery,
  beginLatestQuery,
  type LatestQueryRuntime,
} from "../../query/workbookLatestRequest";
import type { WorkbookPendingRefreshBlockScope } from "../../runtime/workbookPendingReplayRuntime";
import { reconcileWorkbookRecordRows } from "../../utils/workbookRowReconciliation";
import { commitTimelineProjection } from "../adapters/timelineProjectionCommitAdapter";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import type {
  TimelineMutableRef,
  TimelineRowStoreCommands,
} from "../models/timelineControllerPorts";
import {
  createTimelineLoadState,
  type TimelineLoadEffect,
  type TimelineLoadEvent,
  type TimelineLoadIdentity,
  type TimelineLoadState,
  type TimelineLoadSubject,
  timelineFreshnessRetryLimit,
  timelineLoadIdentitiesEqual,
  transitionTimelineLoad,
} from "../models/timelineLoadMachine";
import {
  normalizeTimelineFullRow,
  rowFromApi,
  type WorkbookRow,
} from "../models/timelineRowModel";
import { ensureTimelineDraftRow } from "../models/timelineRowsModel";
import type {
  TimelineSourceRecordEvidence,
  TimelineSourceRecordRequirement,
} from "../models/timelineViewportContinuityModel";
import { timelineSourceRecordRequirementSatisfied } from "../models/timelineViewportContinuityModel";
import { decideWorkbookRecordFreshness } from "../models/workbookRecordFreshness";

type LoadRowsOptions = {
  requireAcceptance?: boolean;
  afterProjectionCommit?: () => void;
  showLoading: boolean;
  freshnessRetryDepth?: number;
  sourceRecordRequirement?: TimelineSourceRecordRequirement;
  viewportContinuityToken?: number;
};

type TimelineRowsLoaderInput = {
  readonly acceptCommittedTimelineRows: (rows: readonly WorkbookRow[]) => void;
  readonly advanceViewportContinuity: (
    token?: number,
    options?: { sourceRecord?: TimelineSourceRecordEvidence },
  ) => void;
  readonly beginRefreshInFlight: (
    scope: WorkbookPendingRefreshBlockScope,
  ) => () => void;
  readonly beginTimelineRowsLoad: () => {
    readonly queryStartEpoch: number;
    readonly requestSequence: number;
  };
  readonly currentCreatedRowPresentationRecordId: () => string | null;
  readonly currentCommittedTimelineRow: (
    recordId: string,
  ) => WorkbookRow | null;
  readonly currentMutationEpoch: () => number;
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
  readonly failViewportContinuity: (token: number) => void;
  readonly hasLoadedRows: () => boolean;
  readonly isCurrentLoadSequence: (requestSequence: number) => boolean;
  readonly knownTimelineRowVersion: (
    recordId: string,
  ) => number | null | undefined;
  readonly loadIdentity: TimelineLoadIdentity;
  readonly markRowsLoaded: () => void;
  readonly nextDraftIndex: () => number;
  readonly onAuthorityUncertain?: (() => void) | undefined;
  readonly publishSaveStatePresentation: () => void;
  readonly queryState: WorkbookQueryState;
  readonly rowStoreCommands: TimelineRowStoreCommands;
  readonly rowsRef: TimelineMutableRef<WorkbookRow[]>;
  readonly setInitialLoadGenerationKey: (generationKey: number) => void;
  readonly setIsInitialLoading: (loading: boolean) => void;
  readonly setIsRefreshing: (refreshing: boolean) => void;
  readonly setLoadAccessLost: (lost: boolean) => void;
  readonly setLoadError: (message: string | null) => void;
  readonly setRefreshError: (message: string | null) => void;
  readonly viewQuery: WorkbookViewQueryPort;
};

const timelineContract = requireViewContract(timelineViewSchemaId);
const allRefreshScope: WorkbookPendingRefreshBlockScope = { kind: "all" };

function retryOptions(
  options: LoadRowsOptions,
  freshnessRetryDepth: number,
): LoadRowsOptions {
  const next: LoadRowsOptions = {
    ...options,
    freshnessRetryDepth,
    showLoading: false,
  };
  if (options.afterProjectionCommit !== undefined) {
    next.afterProjectionCommit = options.afterProjectionCommit;
  }
  if (options.sourceRecordRequirement !== undefined) {
    next.sourceRecordRequirement = options.sourceRecordRequirement;
  }
  if (options.viewportContinuityToken !== undefined) {
    next.viewportContinuityToken = options.viewportContinuityToken;
  }
  return next;
}

function convergenceFailureMessage(options: LoadRowsOptions): string {
  const requirement = options.sourceRecordRequirement;
  return requirement === undefined
    ? "Timeline projection did not converge."
    : `Timeline row ${requirement.recordId} did not reach version ${requirement.minimumRowVersion}.`;
}

function currentSourceRecordEvidence(
  rows: readonly WorkbookRow[],
  requirement: TimelineSourceRecordRequirement | undefined,
): TimelineSourceRecordEvidence | null {
  if (requirement === undefined) return null;
  const sourceRow = rows.find(
    (row) => row.recordId === requirement.recordId && row.rowVersion !== null,
  );
  if (
    sourceRow?.recordId === null ||
    sourceRow?.recordId === undefined ||
    sourceRow.rowVersion === null
  ) {
    return null;
  }
  const evidence = {
    recordId: sourceRow.recordId,
    rowVersion: sourceRow.rowVersion,
  };
  return timelineSourceRecordRequirementSatisfied(requirement, evidence)
    ? evidence
    : null;
}

function sourceEvidenceFromCommittedRows(
  rows: readonly WorkbookRow[],
  requirement: TimelineSourceRecordRequirement | undefined,
): TimelineSourceRecordEvidence | undefined {
  return currentSourceRecordEvidence(rows, requirement) ?? undefined;
}

export function useTimelineRowsLoader(input: TimelineRowsLoaderInput) {
  const {
    acceptCommittedTimelineRows,
    advanceViewportContinuity,
    beginRefreshInFlight,
    beginTimelineRowsLoad,
    currentCreatedRowPresentationRecordId,
    currentCommittedTimelineRow,
    currentMutationEpoch,
    editorDraftRegistry,
    failViewportContinuity,
    hasLoadedRows,
    isCurrentLoadSequence,
    knownTimelineRowVersion,
    loadIdentity,
    markRowsLoaded,
    nextDraftIndex,
    onAuthorityUncertain,
    publishSaveStatePresentation,
    queryState,
    rowStoreCommands: { replaceRows },
    rowsRef,
    setInitialLoadGenerationKey,
    setIsInitialLoading,
    setIsRefreshing,
    setLoadAccessLost,
    setLoadError,
    setRefreshError,
    viewQuery,
  } = input;
  const { browser, snapshot: browsing } = useWorkbookQueryBrowser(
    viewQuery,
    timelineViewSchemaId,
  );
  const queryRuntimeRef = useRef<LatestQueryRuntime>({
    controller: null,
    sequence: 0,
  });
  const machineRef = useRef<TimelineLoadState>(
    createTimelineLoadState(loadIdentity, currentMutationEpoch()),
  );

  const dispatchLoadEvent = useCallback((event: TimelineLoadEvent) => {
    const transition = transitionTimelineLoad(machineRef.current, event);
    machineRef.current = transition.state;
    return transition.effects;
  }, []);

  const publishLoadStatus = useCallback(
    (effect: Extract<TimelineLoadEffect, { kind: "publish_status" }>) => {
      switch (effect.status) {
        case "idle":
          setIsInitialLoading(false);
          setIsRefreshing(false);
          setLoadAccessLost(false);
          setLoadError(null);
          setRefreshError(null);
          return;
        case "initial_loading":
          if (effect.requestGeneration !== null) {
            setInitialLoadGenerationKey(effect.requestGeneration);
          }
          setIsInitialLoading(true);
          setIsRefreshing(false);
          setLoadAccessLost(false);
          setLoadError(null);
          return;
        case "refreshing":
          setIsRefreshing(true);
          setRefreshError(null);
          return;
        case "ready":
          setIsInitialLoading(false);
          setIsRefreshing(false);
          setLoadAccessLost(false);
          setLoadError(null);
          setRefreshError(null);
          return;
        case "stale_error":
          setIsInitialLoading(false);
          setIsRefreshing(false);
          setLoadAccessLost(false);
          if (effect.hasLoadedRows) {
            setRefreshError(effect.message);
          } else {
            setLoadError(effect.message);
          }
          return;
        case "unavailable":
          setIsInitialLoading(false);
          setIsRefreshing(false);
          setLoadAccessLost(true);
          setLoadError(effect.message);
      }
    },
    [
      setInitialLoadGenerationKey,
      setIsInitialLoading,
      setIsRefreshing,
      setLoadAccessLost,
      setLoadError,
      setRefreshError,
    ],
  );

  const applyLifecycleEffects = useCallback(
    (effects: readonly TimelineLoadEffect[], options?: LoadRowsOptions) => {
      for (const effect of effects) {
        switch (effect.kind) {
          case "publish_status":
            publishLoadStatus(effect);
            break;
          case "clear_protected_rows":
            browser.invalidate();
            rowsRef.current = [];
            replaceRows([]);
            onAuthorityUncertain?.();
            break;
          case "fail_continuity":
            if (options?.viewportContinuityToken !== undefined) {
              failViewportContinuity(options.viewportContinuityToken);
            }
            break;
          case "commit":
          case "request":
          case "retry":
          case "settle_obligation":
            break;
        }
      }
    },
    [
      failViewportContinuity,
      onAuthorityUncertain,
      publishLoadStatus,
      replaceRows,
      rowsRef,
      browser,
    ],
  );

  useEffect(() => {
    // Callback identity may change while a live row update is accepted. Only
    // a different query subject cancels the active refresh or its retry.
    if (timelineLoadIdentitiesEqual(machineRef.current.identity, loadIdentity))
      return;
    abortLatestQuery(queryRuntimeRef);
    const effects = dispatchLoadEvent({
      hasLoadedRows: hasLoadedRows(),
      identity: loadIdentity,
      kind: "subject_changed",
      mutationEpoch: currentMutationEpoch(),
    });
    applyLifecycleEffects(effects);
  }, [
    applyLifecycleEffects,
    currentMutationEpoch,
    dispatchLoadEvent,
    hasLoadedRows,
    loadIdentity.incidentId,
    loadIdentity.queryIdentity,
    loadIdentity.surfaceIdentity,
    loadIdentity,
  ]);

  useEffect(
    () => () => {
      abortLatestQuery(queryRuntimeRef);
    },
    [],
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
          if (current !== null) rows.push(current);
        } else {
          rows.push(row);
        }
      }
      if (
        sourceRecordRequirement !== undefined &&
        sourceEvidenceFromCommittedRows(rows, sourceRecordRequirement) ===
          undefined &&
        sourceEvidenceFromCommittedRows(
          [
            currentCommittedTimelineRow(sourceRecordRequirement.recordId),
          ].filter((row): row is WorkbookRow => row !== null),
          sourceRecordRequirement,
        ) === undefined
      ) {
        hasStaleRows = true;
      }
      return { hasStaleRows, rows };
    },
    [currentCommittedTimelineRow, knownTimelineRowVersion],
  );

  const settleProjectionObligation = useCallback(
    (options: LoadRowsOptions, evidence: TimelineSourceRecordEvidence) => {
      commitTimelineProjection(() => {
        options.afterProjectionCommit?.();
      }, true);
      advanceViewportContinuity(options.viewportContinuityToken, {
        sourceRecord: evidence,
      });
    },
    [advanceViewportContinuity],
  );

  const commitAcceptedRows = useCallback(
    (incomingRows: readonly WorkbookRow[], options: LoadRowsOptions) => {
      const { committedRows, rows: hydratedRows } =
        reconcileCommittedRowsWithLocalDrafts({
          currentRows: rowsRef.current,
          incomingRows,
          materializeRow: editorDraftRegistry.materializeRow,
          nextDraftIndex,
          pinnedCommittedRow: (() => {
            const recordId = currentCreatedRowPresentationRecordId();
            return recordId === null
              ? null
              : currentCommittedTimelineRow(recordId);
          })(),
        });
      acceptCommittedTimelineRows(committedRows);
      editorDraftRegistry.retainRows(
        new Set(hydratedRows.map((row) => row.key)),
      );
      rowsRef.current = hydratedRows;
      commitTimelineProjection(() => {
        replaceRows(hydratedRows);
        options.afterProjectionCommit?.();
        publishSaveStatePresentation();
        markRowsLoaded();
      }, options.viewportContinuityToken !== undefined);
      const sourceRecord =
        sourceEvidenceFromCommittedRows(
          committedRows,
          options.sourceRecordRequirement,
        ) ??
        (options.sourceRecordRequirement === undefined
          ? undefined
          : sourceEvidenceFromCommittedRows(
              [
                currentCommittedTimelineRow(
                  options.sourceRecordRequirement.recordId,
                ),
              ].filter((row): row is WorkbookRow => row !== null),
              options.sourceRecordRequirement,
            ));
      advanceViewportContinuity(
        options.viewportContinuityToken,
        sourceRecord === undefined ? {} : { sourceRecord },
      );
    },
    [
      acceptCommittedTimelineRows,
      advanceViewportContinuity,
      currentCommittedTimelineRow,
      currentCreatedRowPresentationRecordId,
      editorDraftRegistry,
      markRowsLoaded,
      nextDraftIndex,
      publishSaveStatePresentation,
      replaceRows,
      rowsRef,
    ],
  );

  const loadRows = useCallback(
    async function loadTimelineRows(options: LoadRowsOptions) {
      const retryDepth = options.freshnessRetryDepth ?? 0;
      const { queryStartEpoch, requestSequence } = beginTimelineRowsLoad();
      const subject: TimelineLoadSubject = {
        ...loadIdentity,
        mutationEpoch: queryStartEpoch,
        requestGeneration: requestSequence,
        sourceVersionObligation: options.sourceRecordRequirement ?? null,
      };
      applyLifecycleEffects(
        dispatchLoadEvent({
          hasLoadedRows: hasLoadedRows(),
          kind: "start",
          retryDepth,
          showLoading: options.showLoading,
          subject,
        }),
        options,
      );
      const request = beginLatestQuery(queryRuntimeRef);
      browser.observeRows(
        rowsRef.current.flatMap((row) =>
          row.rawRow === null ? [] : [row.rawRow],
        ),
      );
      const result = await browser.query(
        {
          contract: timelineContract,
          queryState,
          signal: request.signal,
        },
        { recoveryDepth: retryDepth },
      );

      const retryStaleResult = async (
        retryable: boolean,
        evidence: TimelineSourceRecordEvidence | null,
      ) => {
        if (result.kind === "accepted")
          browser.reject(result.value, {
            kind: "stale_target",
            message:
              "The query is older than an accepted change. Refresh current rows.",
          });
        const sharedRetryDepth = Math.max(
          retryDepth,
          browser.recoveryAttemptsUsed(),
        );
        const effects = dispatchLoadEvent({
          kind: "stale_result",
          obligationSatisfied: evidence !== null,
          retryDepth: sharedRetryDepth,
          retryable,
          subject,
        });
        applyLifecycleEffects(effects, options);
        if (effects.some((effect) => effect.kind === "settle_obligation")) {
          if (evidence !== null) settleProjectionObligation(options, evidence);
          return;
        }
        const retry = effects.find((effect) => effect.kind === "retry");
        if (retry !== undefined) {
          const finishRefresh = beginRefreshInFlight(allRefreshScope);
          try {
            await loadTimelineRows(retryOptions(options, retry.retryDepth));
          } finally {
            finishRefresh();
          }
          return;
        }
        if (options.requireAcceptance)
          throw new WorkbookSurfaceRefreshError({
            kind: "unavailable",
            failure: "transient",
          });
        if (retryable && sharedRetryDepth >= timelineFreshnessRetryLimit) {
          applyLifecycleEffects(
            dispatchLoadEvent({
              hasLoadedRows: hasLoadedRows(),
              kind: "retry_exhaustion",
              message: convergenceFailureMessage(options),
              subject,
            }),
            options,
          );
        }
      };

      const obligationEvidence =
        currentSourceRecordEvidence(
          rowsRef.current,
          options.sourceRecordRequirement,
        ) ??
        (options.sourceRecordRequirement === undefined
          ? null
          : currentSourceRecordEvidence(
              [
                currentCommittedTimelineRow(
                  options.sourceRecordRequirement.recordId,
                ),
              ].filter((row): row is WorkbookRow => row !== null),
              options.sourceRecordRequirement,
            ));
      if (!request.isCurrent() || !isCurrentLoadSequence(requestSequence)) {
        if (options.requireAcceptance)
          requireWorkbookSurfaceAcceptance({ kind: "aborted" });
        await retryStaleResult(
          options.sourceRecordRequirement !== undefined,
          obligationEvidence,
        );
        return;
      }
      const mutationEpoch = currentMutationEpoch();
      if (mutationEpoch !== subject.mutationEpoch) {
        dispatchLoadEvent({ kind: "accepted_mutation", mutationEpoch });
        await retryStaleResult(true, obligationEvidence);
        return;
      }
      if (result.kind === "aborted") {
        if (options.requireAcceptance) requireWorkbookSurfaceAcceptance(result);
        return;
      }
      if (result.kind === "rejected") {
        const event: TimelineLoadEvent =
          workbookFailureLifecycle(result.failure).kind ===
          "authority_unavailable"
            ? {
                kind: "access_loss",
                message: result.failure.message,
                subject,
              }
            : {
                hasLoadedRows: hasLoadedRows(),
                kind: "failure",
                message: result.failure.message,
                subject,
              };
        applyLifecycleEffects(dispatchLoadEvent(event), options);
        if (options.requireAcceptance) requireWorkbookSurfaceAcceptance(result);
        return;
      }
      let incomingRows: WorkbookRow[];
      try {
        incomingRows = result.value.rows.map((row, index) =>
          rowFromApi(
            normalizeTimelineFullRow(row, `query response rows[${index}]`),
          ),
        );
      } catch {
        browser.reject(result.value, {
          kind: "invalid_contract",
          message: "Timeline projection load failed.",
        });
        if (options.requireAcceptance)
          throw new WorkbookSurfaceRefreshError({
            kind: "unavailable",
            failure: "contract",
          });
        applyLifecycleEffects(
          dispatchLoadEvent({
            hasLoadedRows: hasLoadedRows(),
            kind: "failure",
            message: "Timeline projection load failed.",
            subject,
          }),
          options,
        );
        return;
      }
      const freshness = freshTimelineRowsForQueryResult(
        incomingRows,
        options.sourceRecordRequirement,
      );
      if (freshness.hasStaleRows) {
        browser.reject(result.value, {
          kind: "stale_target",
          message:
            "The query is older than an accepted change. Refresh current rows.",
        });
        await retryStaleResult(true, obligationEvidence);
        return;
      }
      const effects = dispatchLoadEvent({ kind: "success", subject });
      if (
        effects.some((effect) => effect.kind === "commit") &&
        browser.accept(result.value)
      ) {
        commitAcceptedRows(freshness.rows, options);
      }
      if (
        options.requireAcceptance &&
        !effects.some((effect) => effect.kind === "commit")
      )
        requireWorkbookSurfaceAcceptance({ kind: "aborted" });
      applyLifecycleEffects(effects, options);
    },
    [
      applyLifecycleEffects,
      beginRefreshInFlight,
      beginTimelineRowsLoad,
      commitAcceptedRows,
      currentMutationEpoch,
      currentCommittedTimelineRow,
      dispatchLoadEvent,
      freshTimelineRowsForQueryResult,
      hasLoadedRows,
      isCurrentLoadSequence,
      loadIdentity,
      queryState,
      rowsRef,
      settleProjectionObligation,
      browser,
    ],
  );

  const browse = useCallback(
    () => loadRows({ showLoading: false }),
    [loadRows],
  );
  useWorkbookBrowsingRead(timelineViewSchemaId, browse);
  return { loadRows, browser, browsing };
}

export function reconcileCommittedRowsWithLocalDrafts({
  currentRows,
  incomingRows,
  materializeRow,
  nextDraftIndex,
  pinnedCommittedRow = null,
}: {
  readonly currentRows: readonly WorkbookRow[];
  readonly incomingRows: readonly WorkbookRow[];
  readonly materializeRow: (row: WorkbookRow) => WorkbookRow;
  readonly nextDraftIndex: () => number;
  readonly pinnedCommittedRow?: WorkbookRow | null;
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
  const incomingRowsWithPresentationPin =
    pinnedCommittedRow?.recordId !== null &&
    pinnedCommittedRow?.recordId !== undefined &&
    !incomingRows.some((row) => row.recordId === pinnedCommittedRow.recordId)
      ? [...incomingRows, pinnedCommittedRow]
      : incomingRows;
  const committedRows = reconcileWorkbookRecordRows(
    currentRows.filter((row) => row.recordId !== null),
    incomingRowsWithPresentationPin,
  ).map((row) => {
    let rowWithLocalState = row;
    if (row.recordId === null) return row;
    const current = currentCommittedByRecordId.get(row.recordId);
    if (
      current !== undefined &&
      (current.pendingSignature !== null ||
        current.collectionDrafts.hostRefs !== "" ||
        current.collectionDrafts.identityRefs !== "" ||
        current.collectionDrafts.tags !== "")
    ) {
      rowWithLocalState = {
        ...rowWithLocalState,
        collectionDrafts: current.collectionDrafts,
        pendingSignature: current.pendingSignature,
      };
    }
    return materializeRow(rowWithLocalState);
  });
  const localDraftRows = currentRows
    .filter((row) => row.recordId === null)
    .map(materializeRow);

  return {
    committedRows,
    rows: ensureTimelineDraftRow({
      nextDraftIndex,
      rows: [...committedRows, ...localDraftRows],
    }).rows,
  };
}
