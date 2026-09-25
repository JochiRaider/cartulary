import type { GridCellStateInput } from "@cartulary/grid-adapter";
import { requireViewContract } from "@cartulary/view-contracts";
import { useCallback, useEffect, useMemo } from "react";
import { type SheetRef, sheetRefKey } from "../../../shared/sheetRef";
import type { WorkbookCollaborationCoordinator } from "../../collaboration/WorkbookCollaborationCoordinator";
import type { WorkbookQueryState } from "../../models/workbookQuery";
import { buildQueryRequest } from "../../models/workbookQuery";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { TimelineMutationCommandPorts } from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookViewQueryPort } from "../../query/WorkbookViewQueryPort";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import { useTimelineCollaborationBindings } from "../collaboration/useTimelineCollaborationBindings";
import { useTimelinePresenceController } from "../collaboration/useTimelinePresenceController";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import { useTimelineCommittedRecordIdle } from "../hooks/useTimelineCommittedRecordIdle";
import type { useTimelineCommittedRows } from "../hooks/useTimelineCommittedRows";
import { useTimelineMutationCommands } from "../hooks/useTimelineMutationCommands";
import { useTimelineMutationDriver } from "../hooks/useTimelineMutationDriver";
import { useTimelineMutationRuntimeBindings } from "../hooks/useTimelineMutationRuntimeBindings";
import { useTimelineRowsLoader } from "../hooks/useTimelineRowsLoader";
import type { TimelineViewportContinuityTarget } from "../hooks/useTimelineViewportContinuityController";
import type { TimelineAcceptedContinuity } from "../models/timelineAcceptedMutationEffects";
import type {
  TimelineMutableRef,
  TimelineRowMutationEditorPort,
  TimelineRowStoreCommands,
} from "../models/timelineControllerPorts";
import { inputFocusKey } from "../models/timelineFieldRegistry";
import type { TimelinePendingSavesRefs } from "../models/timelinePendingSaves";
import type { WorkbookRow } from "../models/timelineRowModel";
import type {
  TimelineContinuityRequirementName,
  TimelineSourceRecordEvidence,
} from "../models/timelineViewportContinuityModel";
import { useTimelineRowMutationCoordinator } from "../mutations/useTimelineRowMutationCoordinator";

const timelineContract = requireViewContract(timelineViewSchemaId);

type TimelineMutationCompositionInput = {
  readonly collaborationProjection: WorkbookCollaborationCoordinator;
  readonly foundation: {
    readonly committedRows: ReturnType<
      typeof useTimelineCommittedRows
    >["commands"];
    readonly editorDraftRegistry: TimelineEditorDraftRegistry;
    readonly nextDraftIndex: () => number;
    readonly loadAccessLost: boolean;
    readonly pendingSavesRefs: TimelinePendingSavesRefs;
    readonly recordWorkbookTiming: (
      name: string,
      details?: Record<string, unknown>,
    ) => void;
    readonly rowStoreCommands: TimelineRowStoreCommands;
    readonly rowsRef: TimelineMutableRef<WorkbookRow[]>;
    readonly setInitialLoadGenerationKey: (generationKey: number) => void;
    readonly setIsInitialLoading: (loading: boolean) => void;
    readonly setIsRefreshing: (refreshing: boolean) => void;
    readonly setLoadAccessLost: (lost: boolean) => void;
    readonly setLoadError: (message: string | null) => void;
    readonly setRefreshError: (message: string | null) => void;
    readonly setMutationError: (message: string | null) => void;
  };
  readonly grid: {
    readonly advanceViewportContinuity: (
      token?: number,
      options?: {
        readonly sourceRecord?: TimelineSourceRecordEvidence;
        readonly target?: TimelineViewportContinuityTarget | null;
      },
    ) => void;
    readonly beginViewportContinuity: (
      target: TimelineViewportContinuityTarget,
      options?: {
        readonly requirements?: readonly TimelineContinuityRequirementName[];
      },
    ) => number;
    readonly clearViewportContinuity: (token: number) => void;
    readonly completeAcceptedViewportContinuity: (
      token: number | undefined,
      continuity: TimelineAcceptedContinuity,
    ) => void;
    readonly editorPort: TimelineRowMutationEditorPort;
    readonly failViewportContinuity: (token: number) => void;
    readonly viewportContinuityRequest: { readonly token: number } | null;
  };
  readonly incident: {
    readonly continuityResetKey: string;
    readonly id: string;
    readonly reloadToken: number;
    readonly sheetRef: SheetRef;
  };
  readonly inspector: {
    readonly selectedRowId: string | null;
    readonly selectRow: (recordId: string | null) => void;
  };
  readonly mutationCommands: TimelineMutationCommandPorts;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly onAuthorityUncertain: (() => void) | undefined;
  readonly query: {
    readonly queryState: WorkbookQueryState;
    readonly viewQuery: WorkbookViewQueryPort;
  };
};

export function useTimelineMutationComposition({
  collaborationProjection,
  foundation,
  grid,
  incident,
  inspector,
  mutationCommands,
  mutationRuntime,
  onAuthorityUncertain,
  query,
}: TimelineMutationCompositionInput) {
  const timelineQueryIdentity = useMemo(
    () => JSON.stringify(buildQueryRequest(timelineContract, query.queryState)),
    [query.queryState],
  );
  const timelineSurfaceIdentity = sheetRefKey(incident.sheetRef);
  const timelineLoadIdentity = useMemo(
    () => ({
      incidentId: incident.id,
      queryIdentity: timelineQueryIdentity,
      surfaceIdentity: timelineSurfaceIdentity,
    }),
    [incident.id, timelineQueryIdentity, timelineSurfaceIdentity],
  );
  const createdRowPresentationScopeKey = useMemo(
    () => `${incident.continuityResetKey}:${timelineQueryIdentity}`,
    [incident.continuityResetKey, timelineQueryIdentity],
  );
  const rowMutations = useTimelineRowMutationCoordinator({
    committedRows: foundation.committedRows,
    sheetRef: incident.sheetRef,
    completeAcceptedViewportContinuity: grid.completeAcceptedViewportContinuity,
    createdRowPresentationScopeKey,
    editorDraftRegistry: foundation.editorDraftRegistry,
    editorPort: grid.editorPort,
    mutationRuntime,
    nextDraftIndex: foundation.nextDraftIndex,
    pendingSavesRefs: foundation.pendingSavesRefs,
    rowsRef: foundation.rowsRef,
    selectedRowId: inspector.selectedRowId,
    rowStoreCommands: foundation.rowStoreCommands,
    setSelectedRowId: inspector.selectRow,
  });
  const { activeConflict, commonConflicts } = rowMutations.snapshot;
  const conflictQueue = useMemo(
    () =>
      Object.fromEntries(
        commonConflicts
          .filter((entry) => entry.origin.viewSchemaId === timelineViewSchemaId)
          .map((entry) => [entry.key, entry]),
      ),
    [commonConflicts],
  );
  const conflictCellKeys = useMemo(
    () =>
      new Set(
        commonConflicts
          .filter((entry) => entry.origin.viewSchemaId === timelineViewSchemaId)
          .map(
            (entry) =>
              `${entry.conflict.record_id}\u0000${entry.conflict.field_key}`,
          ),
      ),
    [commonConflicts],
  );
  const getCellState = useCallback(
    ({
      fieldKey,
      recordId,
    }: {
      readonly fieldKey: string;
      readonly recordId: string;
    }): GridCellStateInput => ({
      conflicted: conflictCellKeys.has(`${recordId}\u0000${fieldKey}`),
    }),
    [conflictCellKeys],
  );
  useEffect(() => {
    if (
      grid.viewportContinuityRequest === null ||
      !commonConflicts.some(
        (entry) => entry.origin.viewSchemaId === timelineViewSchemaId,
      )
    ) {
      return;
    }
    grid.clearViewportContinuity(grid.viewportContinuityRequest.token);
  }, [
    commonConflicts,
    grid.clearViewportContinuity,
    grid.viewportContinuityRequest,
  ]);

  const queryAdmission = rowMutations.ports.queryAdmission;
  const { loadRows, browser, browsing } = useTimelineRowsLoader({
    acceptCommittedTimelineRows:
      rowMutations.commands.acceptCommittedTimelineRows,
    advanceViewportContinuity: grid.advanceViewportContinuity,
    beginRefreshInFlight: rowMutations.commands.beginRefreshInFlight,
    beginTimelineRowsLoad: queryAdmission.beginLoad,
    currentCreatedRowPresentationRecordId:
      queryAdmission.currentCreatedRowPresentationRecordId,
    currentCommittedTimelineRow: queryAdmission.currentCommittedTimelineRow,
    currentMutationEpoch: queryAdmission.currentMutationEpoch,
    failViewportContinuity: grid.failViewportContinuity,
    hasLoadedRows: queryAdmission.hasLoadedRows,
    isCurrentLoadSequence: queryAdmission.isCurrentLoadSequence,
    knownTimelineRowVersion: queryAdmission.knownTimelineRowVersion,
    markRowsLoaded: queryAdmission.markRowsLoaded,
    loadIdentity: timelineLoadIdentity,
    nextDraftIndex: foundation.nextDraftIndex,
    onAuthorityUncertain,
    publishSaveStatePresentation:
      rowMutations.commands.publishSaveStatePresentation,
    queryState: query.queryState,
    rowsRef: foundation.rowsRef,
    editorDraftRegistry: foundation.editorDraftRegistry,
    setIsInitialLoading: foundation.setIsInitialLoading,
    setInitialLoadGenerationKey: foundation.setInitialLoadGenerationKey,
    setIsRefreshing: foundation.setIsRefreshing,
    setLoadAccessLost: foundation.setLoadAccessLost,
    setLoadError: foundation.setLoadError,
    setRefreshError: foundation.setRefreshError,
    rowStoreCommands: foundation.rowStoreCommands,
    viewQuery: query.viewQuery,
  });
  const waitForCommittedRecordIdle = useTimelineCommittedRecordIdle({
    mutationRuntime,
    latestCommittedRowVersion: rowMutations.commands.latestCommittedRowVersion,
    latestCommittedTimelineRow:
      rowMutations.commands.latestCommittedTimelineRow,
    loadRows,
  });
  const activeSheetRef = useMemo<SheetRef>(
    () =>
      incident.sheetRef ?? { kind: "view_schema", id: timelineViewSchemaId },
    [incident.sheetRef],
  );
  const refreshRowsForCollaboration = useCallback(
    (options?: { readonly requireAcceptance?: boolean }) =>
      loadRows({ showLoading: false, ...options }),
    [loadRows],
  );
  const collaboration = useTimelineCollaborationBindings({
    activeSheetRef,
    admission: rowMutations.ports.collaborationAdmission,
    beginRowsLoad: queryAdmission.beginLoad,
    clearCommittedRows: foundation.committedRows.clearProtectedRows,
    collaborationProjection,
    refreshRows: refreshRowsForCollaboration,
    resolveClientTxn: rowMutations.commands.resolvePendingSocketTxn,
    rowsRef: foundation.rowsRef,
    rowStoreCommands: foundation.rowStoreCommands,
  });

  useEffect(() => {
    void incident.reloadToken;
    void loadRows({ showLoading: true });
  }, [incident.reloadToken, loadRows]);

  const replay = useTimelineMutationDriver({
    sheetRef: activeSheetRef,
    applyAcceptedRowMutation: rowMutations.commands.applyAcceptedRowMutation,
    settleEditorRevisions: foundation.editorDraftRegistry.settleRevisions,
    acceptEditorPredecessor: foundation.editorDraftRegistry.acceptPredecessor,
    batchAuthoring: foundation.editorDraftRegistry.batch,
    captureEditorDrafts: foundation.editorDraftRegistry.captureRow,
    clearViewportContinuity: grid.clearViewportContinuity,
    conflictQueueRef: rowMutations.refs.conflictQueueRef,
    registerMutationConflict: (
      conflict,
      rowKey,
      focusField,
      surface,
      refresh,
      originSheetRef,
      draftRevisions,
    ) => {
      rowMutations.commands.registerSameFieldConflict(
        conflict,
        inputFocusKey(rowKey, focusField, surface),
        surface,
        refresh,
        originSheetRef,
        draftRevisions,
      );
      return true;
    },
    latestCommittedTimelineRow:
      rowMutations.commands.latestCommittedTimelineRow,
    loadRows,
    mutationCommands: mutationCommands.identity,
    mutationRuntime,
    pendingSavesRefs: foundation.pendingSavesRefs,
    postMutationQueryRefreshRequired:
      query.queryState.filters.length > 0 ||
      query.queryState.sort.length > 0 ||
      query.queryState.groupBy !== null,
    recordWorkbookTiming: foundation.recordWorkbookTiming,
    rowsRef: foundation.rowsRef,
    requestAuthorizationRecovery:
      collaboration.commands.requestAuthorizationRecovery,
    setRefreshError: foundation.setRefreshError,
    setMutationError: foundation.setMutationError,
    rowStoreCommands: foundation.rowStoreCommands,
  });
  useTimelineMutationRuntimeBindings({
    applyAcceptedRowMutation: rowMutations.commands.applyAcceptedRowMutation,
    applyAcceptedBatchRows: rowMutations.commands.applyAcceptedBatchRows,
    discardBlockedEdit: replay.discardBlockedEdit,
    editorDraftRegistry: foundation.editorDraftRegistry,
    editorPort: grid.editorPort,
    loadRows,
    mutationRuntime,
  });
  const presence = useTimelinePresenceController({
    presence: collaboration.snapshot.presence,
    publishPresence: collaboration.commands.publishPresence,
    resetKey: `${incident.continuityResetKey}:${
      foundation.loadAccessLost ? "access-lost" : "authorized"
    }`,
  });
  const nextClientTxnId = useCallback(
    () => mutationCommands.identity.createLogicalActionId(),
    [mutationCommands.identity],
  );
  const captureActionBlocksRecord = useCallback(
    (recordId: string) => mutationRuntime.timelineActionBlocksRecord(recordId),
    [mutationRuntime],
  );
  const mutations = useTimelineMutationCommands({
    captureActionBlocksRecord,
    beginViewportContinuity: grid.beginViewportContinuity,
    clearViewportContinuity: grid.clearViewportContinuity,
    clientInstanceId: mutationRuntime.scope.clientInstanceId,
    conflictQueueRef: rowMutations.refs.conflictQueueRef,
    editorDraftRegistry: foundation.editorDraftRegistry,
    enqueuePendingReplayUnit: replay.enqueuePendingReplayUnit,
    incidentId: incident.id,
    latestCommittedTimelineRow:
      rowMutations.commands.latestCommittedTimelineRow,
    nextClientTxnId,
    pendingSavesRefs: foundation.pendingSavesRefs,
    rowsRef: foundation.rowsRef,
    rowStoreCommands: foundation.rowStoreCommands,
  });

  return {
    commands: {
      collaboration: collaboration.commands,
      identity: { nextClientTxnId },
      mutation: mutations.commands,
      presence: presence.commands,
      query: { loadRows, browser },
      replay,
      save: rowMutations.commands,
    },
    ports: {
      activeSheetRef,
      queryAdmission,
      waitForCommittedRecordIdle,
    },
    refs: rowMutations.refs,
    snapshot: {
      browsing,
      cellRangeScopeKey: JSON.stringify([
        incident.id,
        incident.continuityResetKey,
        timelineSurfaceIdentity,
        browsing.canonicalQuery,
      ]),
      collaboration: collaboration.snapshot,
      conflict: {
        activeConflict,
        commonConflicts,
        conflictQueue,
        getCellState,
      },
      presence: presence.snapshot,
    },
  };
}
