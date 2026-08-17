import type { GridCellStateInput } from "@cartulary/grid-adapter";
import { requireViewContract } from "@cartulary/view-contracts";
import {
  type Dispatch,
  type SetStateAction,
  useCallback,
  useEffect,
  useMemo,
} from "react";
import type { SheetRef } from "../../../shared/sheetRef";
import type { WorkbookCollaborationCoordinator } from "../../collaboration/WorkbookCollaborationCoordinator";
import type { WorkbookQueryState } from "../../models/workbookQuery";
import { buildQueryRequest } from "../../models/workbookQuery";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { TimelineMutationCommandPorts } from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookPendingMutationPort } from "../../ports/WorkbookPendingMutationPort";
import type { WorkbookViewQueryPort } from "../../query/WorkbookViewQueryPort";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import type {
  WorkbookPendingQueueSnapshot,
  WorkbookPendingSavesRefs,
} from "../../runtime/workbookPendingReplayRuntime";
import { useTimelineCollaborationBindings } from "../collaboration/useTimelineCollaborationBindings";
import { useTimelinePresenceController } from "../collaboration/useTimelinePresenceController";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import { useTimelineCommittedRecordIdle } from "../hooks/useTimelineCommittedRecordIdle";
import { useTimelineMutationCommands } from "../hooks/useTimelineMutationCommands";
import { useTimelineMutationRuntimeBindings } from "../hooks/useTimelineMutationRuntimeBindings";
import { useTimelinePendingReplayController } from "../hooks/useTimelinePendingReplayController";
import { useTimelineRowsLoader } from "../hooks/useTimelineRowsLoader";
import type { TimelineViewportContinuityTarget } from "../hooks/useTimelineViewportContinuityController";
import type {
  PendingReplayRuntimeMeta,
  TimelineMutableRef,
  TimelineRowMutationEditorPort,
  TimelineRowStoreCommands,
} from "../models/timelineControllerPorts";
import type {
  TimelineContinuityRequirementName,
  TimelineSourceRecordEvidence,
} from "../models/timelineViewportContinuityModel";
import type {
  AutoResolutionNotice,
  DismissedMention,
} from "../models/workbookMentionChips";
import { reconcileDismissedMentionsForRow } from "../models/workbookMentionChips";
import {
  inputFocusKey,
  type WorkbookRow,
} from "../models/workbookTimelineModel";
import { useTimelineRowMutationCoordinator } from "../mutations/useTimelineRowMutationCoordinator";
import type { TimelineRecordActionPort } from "../ports/TimelineRecordActionPort";

const timelineContract = requireViewContract(timelineViewSchemaId);

type TimelineMutationCompositionInput = {
  readonly collaborationProjection: WorkbookCollaborationCoordinator;
  readonly foundation: {
    readonly clearActiveCollectionInputKey: (focusKey: string) => void;
    readonly editorDraftRegistry: TimelineEditorDraftRegistry;
    readonly nextDraftIndex: () => number;
    readonly loadAccessLost: boolean;
    readonly pendingQueueSnapshot: WorkbookPendingQueueSnapshot;
    readonly pendingSavesRefs: WorkbookPendingSavesRefs<PendingReplayRuntimeMeta>;
    readonly recordActionPort: TimelineRecordActionPort;
    readonly recordWorkbookTiming: (
      name: string,
      details?: Record<string, unknown>,
    ) => void;
    readonly rowStoreCommands: TimelineRowStoreCommands;
    readonly rowsRef: TimelineMutableRef<WorkbookRow[]>;
    readonly setAutoResolutionNotices: Dispatch<
      SetStateAction<AutoResolutionNotice[]>
    >;
    readonly setDismissedMentionsByRow: Dispatch<
      SetStateAction<Record<string, DismissedMention[]>>
    >;
    readonly setInitialLoadGenerationKey: (generationKey: number) => void;
    readonly setIsInitialLoading: (loading: boolean) => void;
    readonly setIsRefreshing: (refreshing: boolean) => void;
    readonly setLoadAccessLost: (lost: boolean) => void;
    readonly setLoadError: (message: string | null) => void;
    readonly setPendingQueueSnapshot: (
      snapshot: WorkbookPendingQueueSnapshot,
    ) => void;
    readonly setRefreshError: (message: string | null) => void;
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
  readonly onIncidentAccessLost: (() => void) | undefined;
  readonly pendingMutationPort: WorkbookPendingMutationPort;
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
  onIncidentAccessLost,
  pendingMutationPort,
  query,
}: TimelineMutationCompositionInput) {
  const createdRowPresentationScopeKey = useMemo(
    () =>
      `${incident.continuityResetKey}:${JSON.stringify(
        buildQueryRequest(timelineContract, query.queryState),
      )}`,
    [incident.continuityResetKey, query.queryState],
  );
  const rowMutations = useTimelineRowMutationCoordinator({
    advanceViewportContinuity: grid.advanceViewportContinuity,
    clearActiveCollectionInputKey: foundation.clearActiveCollectionInputKey,
    clearViewportContinuity: grid.clearViewportContinuity,
    createdRowPresentationScopeKey,
    editorDraftRegistry: foundation.editorDraftRegistry,
    editorPort: grid.editorPort,
    mutationRuntime,
    nextDraftIndex: foundation.nextDraftIndex,
    pendingQueueSnapshot: foundation.pendingQueueSnapshot,
    pendingSavesRefs: foundation.pendingSavesRefs,
    rowsRef: foundation.rowsRef,
    selectedRowId: inspector.selectedRowId,
    setAutoResolutionNotices: foundation.setAutoResolutionNotices,
    setDismissedMentionsByRow: foundation.setDismissedMentionsByRow,
    setPendingQueueSnapshot: foundation.setPendingQueueSnapshot,
    rowStoreCommands: foundation.rowStoreCommands,
    setSelectedRowId: inspector.selectRow,
  });
  const { activeConflict, commonMutationSnapshot, conflictQueue } =
    rowMutations.snapshot;
  const conflictCellKeys = useMemo(
    () =>
      new Set(
        Object.values(conflictQueue).map(
          (entry) => `${entry.anchor.record_id}\u0000${entry.anchor.field_key}`,
        ),
      ),
    [conflictQueue],
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
      !commonMutationSnapshot.conflicts.some(
        (entry) => entry.origin.viewSchemaId === timelineViewSchemaId,
      )
    ) {
      return;
    }
    grid.clearViewportContinuity(grid.viewportContinuityRequest.token);
  }, [
    commonMutationSnapshot.conflicts,
    grid.clearViewportContinuity,
    grid.viewportContinuityRequest,
  ]);

  const queryAdmission = rowMutations.ports.queryAdmission;
  const { loadRows } = useTimelineRowsLoader({
    acceptCommittedTimelineRows:
      rowMutations.commands.acceptCommittedTimelineRows,
    advanceViewportContinuity: grid.advanceViewportContinuity,
    beginRefreshInFlight: rowMutations.commands.beginRefreshInFlight,
    beginTimelineRowsLoad: queryAdmission.beginLoad,
    committedRowsChangedSince: queryAdmission.committedRowsChangedSince,
    currentCreatedRowPresentationRecordId:
      queryAdmission.currentCreatedRowPresentationRecordId,
    currentCommittedTimelineRow: queryAdmission.currentCommittedTimelineRow,
    finishRefreshInFlight: rowMutations.commands.finishRefreshInFlight,
    failViewportContinuity: grid.failViewportContinuity,
    hasLoadedRows: queryAdmission.hasLoadedRows,
    isCurrentLoadSequence: queryAdmission.isCurrentLoadSequence,
    knownTimelineRowVersion: queryAdmission.knownTimelineRowVersion,
    markRowsLoaded: queryAdmission.markRowsLoaded,
    nextDraftIndex: foundation.nextDraftIndex,
    onIncidentAccessLost,
    pendingSavesRefs: foundation.pendingSavesRefs,
    pruneAutoResolutionNoticesForRows:
      rowMutations.commands.pruneAutoResolutionNoticesForRows,
    pruneDismissedMentionsForRow: reconcileDismissedMentionsForRow,
    publishSaveStatePresentation:
      rowMutations.commands.publishSaveStatePresentation,
    queryState: query.queryState,
    rowsRef: foundation.rowsRef,
    editorDraftRegistry: foundation.editorDraftRegistry,
    setDismissedMentionsByRow: foundation.setDismissedMentionsByRow,
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
    conflictQueueRef: rowMutations.refs.conflictQueueRef,
    latestCommittedRowVersion: rowMutations.commands.latestCommittedRowVersion,
    latestCommittedTimelineRow:
      rowMutations.commands.latestCommittedTimelineRow,
    loadRows,
    pendingSavesRefs: foundation.pendingSavesRefs,
  });
  const activeSheetRef = useMemo<SheetRef>(
    () =>
      incident.sheetRef ?? { kind: "view_schema", id: timelineViewSchemaId },
    [incident.sheetRef],
  );
  const refreshRowsForCollaboration = useCallback(
    () => loadRows({ showLoading: false }),
    [loadRows],
  );
  const collaboration = useTimelineCollaborationBindings({
    activeSheetRef,
    admission: rowMutations.ports.collaborationAdmission,
    beginRowsLoad: queryAdmission.beginLoad,
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

  const replay = useTimelinePendingReplayController({
    applyAcceptedRowMutation: rowMutations.commands.applyAcceptedRowMutation,
    clearSubmittedScalarEditorDraftValuesForRow:
      foundation.editorDraftRegistry.clearSubmittedRow,
    clearViewportContinuity: grid.clearViewportContinuity,
    conflictQueueRef: rowMutations.refs.conflictQueueRef,
    registerMutationConflict: (
      conflict,
      rowKey,
      focusField,
      surface,
      refresh,
    ) => {
      rowMutations.commands.registerSameFieldConflict(
        conflict,
        inputFocusKey(rowKey, focusField, surface),
        surface,
        refresh,
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
    publishPendingQueueState: rowMutations.commands.publishPendingQueueState,
    reconcileDiscardedPendingUnit:
      rowMutations.commands.reconcileDiscardedPendingUnit,
    recordWorkbookTiming: foundation.recordWorkbookTiming,
    resolvePendingSocketTxn: rowMutations.commands.resolvePendingSocketTxn,
    rowsRef: foundation.rowsRef,
    requestAuthorizationRecovery:
      collaboration.commands.requestAuthorizationRecovery,
    setRefreshError: foundation.setRefreshError,
    rowStoreCommands: foundation.rowStoreCommands,
    pendingMutationPort,
    trackPendingSocketTxn: rowMutations.commands.trackPendingSocketTxn,
  });
  useTimelineMutationRuntimeBindings({
    applyAcceptedRowMutation: rowMutations.commands.applyAcceptedRowMutation,
    discardBlockedEdit: replay.discardBlockedEdit,
    editorDraftRegistry: foundation.editorDraftRegistry,
    editorPort: grid.editorPort,
    loadRows,
    mutationRuntime,
  });
  const presence = useTimelinePresenceController({
    presenceRecords: collaboration.snapshot.activeSheetPresenceRecords,
    publishPresence: collaboration.commands.publishPresence,
    resetKey: `${incident.continuityResetKey}:${
      foundation.loadAccessLost ? "access-lost" : "authorized"
    }`,
  });
  const nextClientTxnId = useCallback(
    () => mutationCommands.identity.createLogicalActionId(),
    [mutationCommands.identity],
  );
  const mutations = useTimelineMutationCommands({
    acceptTimelineActionResult:
      rowMutations.commands.acceptTimelineActionResult,
    beginSave: rowMutations.commands.beginSave,
    beginViewportContinuity: grid.beginViewportContinuity,
    clearViewportContinuity: grid.clearViewportContinuity,
    clientInstanceId: mutationRuntime.scope.clientInstanceId,
    conflictQueueRef: rowMutations.refs.conflictQueueRef,
    editorDraftRegistry: foundation.editorDraftRegistry,
    enqueueSaveWork: rowMutations.commands.enqueueSaveWork,
    enqueuePendingReplayUnit: replay.enqueuePendingReplayUnit,
    finishSave: rowMutations.commands.finishSave,
    incidentId: incident.id,
    latestCommittedTimelineRow:
      rowMutations.commands.latestCommittedTimelineRow,
    loadRows,
    nextClientTxnId,
    pendingSavesRefs: foundation.pendingSavesRefs,
    recordActionPort: foundation.recordActionPort,
    resolvePendingSocketTxn: rowMutations.commands.resolvePendingSocketTxn,
    rowsRef: foundation.rowsRef,
    rowStoreCommands: foundation.rowStoreCommands,
    trackPendingSocketTxn: rowMutations.commands.trackPendingSocketTxn,
    waitForCommittedRecordIdle,
  });

  return {
    commands: {
      collaboration: collaboration.commands,
      identity: { nextClientTxnId },
      mutation: mutations.commands,
      presence: presence.commands,
      query: { loadRows },
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
      collaboration: collaboration.snapshot,
      conflict: {
        activeConflict,
        commonMutationSnapshot,
        conflictQueue,
        getCellState,
      },
      mutation: mutations.snapshot,
      presence: presence.snapshot,
    },
  };
}
