import { sheetRefKey } from "../../../shared/sheetRef";
import type { TimelineWorkbookSurfaceRuntime } from "../models/timelineWorkbookSurfaceRuntime";
import { useTimelineGridEnvironment } from "./useTimelineGridEnvironment";
import { useTimelineInspectorStateComposition } from "./useTimelineInspectorStateComposition";
import { useTimelineInspectorWorkflowComposition } from "./useTimelineInspectorWorkflowComposition";
import { useTimelineInteractionComposition } from "./useTimelineInteractionComposition";
import { useTimelineMutationComposition } from "./useTimelineMutationComposition";
import { useTimelineSurfaceFoundation } from "./useTimelineSurfaceFoundation";

export function useTimelineWorkbookComposition({
  runtime,
}: {
  readonly runtime: TimelineWorkbookSurfaceRuntime;
}) {
  const foundation = useTimelineSurfaceFoundation({
    apiBase: runtime.incident.apiBase,
    clipboardPaste: runtime.clipboardPaste,
    incidentId: runtime.incident.id,
    mutationCommands: runtime.mutationCommands,
    mutationRuntime: runtime.mutationRuntime,
    query: runtime.query,
  });
  const grid = useTimelineGridEnvironment({
    continuityResetKey: runtime.incident.continuityResetKey,
    editorDraftRegistry: foundation.refs.editorDraftRegistry,
    rowsRef: foundation.refs.rows,
  });
  const inspector = useTimelineInspectorStateComposition({
    continuity: grid.ports.continuity,
    currentIncidentRole: runtime.incident.currentRole,
    dismissedMentionsByRow: foundation.snapshot.mentions.dismissedMentionsByRow,
    inspectorResetKey: runtime.incident.inspectorResetKey,
    rows: foundation.snapshot.rows,
    selectedMentionRef: foundation.snapshot.mentions.selectedMentionRef,
    workbookFocusAnchorRef: grid.refs.workbookFocusAnchor,
  });
  const mutation = useTimelineMutationComposition({
    collaborationProjection: runtime.collaborationProjection,
    foundation: {
      clearActiveCollectionInputKey:
        foundation.commands.editor.deactivateCollectionInput,
      editorDraftRegistry: foundation.refs.editorDraftRegistry,
      loadAccessLost: foundation.snapshot.lifecycle.loadAccessLost,
      nextDraftIndex: foundation.commands.rows.allocateDraftIndex,
      pendingQueueSnapshot: foundation.snapshot.pendingQueue,
      pendingSavesRefs: foundation.refs.pendingSaves,
      recordActionPort: foundation.ports.recordActions,
      recordWorkbookTiming: foundation.commands.recordTiming,
      rowStoreCommands: foundation.commands.rows,
      rowsRef: foundation.refs.rows,
      setAutoResolutionNotices:
        foundation.commands.mentions.setAutoResolutionNotices,
      setDismissedMentionsByRow:
        foundation.commands.mentions.setDismissedMentionsByRow,
      setInitialLoadGenerationKey:
        foundation.commands.lifecycle.setInitialLoadGenerationKey,
      setIsInitialLoading: foundation.commands.lifecycle.setIsInitialLoading,
      setIsRefreshing: foundation.commands.lifecycle.setIsRefreshing,
      setLoadAccessLost: foundation.commands.lifecycle.setLoadAccessLost,
      setLoadError: foundation.commands.lifecycle.setLoadError,
      setPendingQueueSnapshot:
        foundation.commands.pendingSaves.setPendingQueueSnapshot,
      setRefreshError: foundation.commands.lifecycle.setRefreshError,
    },
    grid: {
      advanceViewportContinuity:
        grid.commands.viewportContinuity.advanceViewportContinuity,
      beginViewportContinuity:
        grid.commands.viewportContinuity.beginViewportContinuity,
      clearViewportContinuity:
        grid.commands.viewportContinuity.clearViewportContinuity,
      editorPort: grid.ports.mutationEditor,
      failViewportContinuity:
        grid.commands.viewportContinuity.failViewportContinuity,
      viewportContinuityRequest: grid.snapshot.viewportContinuityRequest,
    },
    incident: {
      continuityResetKey: runtime.incident.continuityResetKey,
      id: runtime.incident.id,
      reloadToken: runtime.incident.reloadToken,
      sheetRef: runtime.incident.sheetRef,
    },
    inspector: {
      selectedRowId: inspector.snapshot.selection.selectedRowId,
      selectRow: inspector.commands.selectRow,
    },
    mutationCommands: runtime.mutationCommands,
    mutationRuntime: runtime.mutationRuntime,
    onIncidentAccessLost: runtime.onIncidentAccessLost,
    query: {
      queryState: foundation.snapshot.query.queryState,
      viewQuery: runtime.query.viewQuery,
    },
  });
  const workflow = useTimelineInspectorWorkflowComposition({
    activeSheetRef: mutation.ports.activeSheetRef,
    knownEntityTypes: new Map(
      Object.values(runtime.entities.index).map((entity) => [
        entity.recordId,
        entity.entityType,
      ]),
    ),
    foundation: {
      evidenceAttachmentPort: foundation.ports.evidenceAttachment,
      historyPort: foundation.ports.history,
      loadAccessLost: foundation.snapshot.lifecycle.loadAccessLost,
      mentionPorts: foundation.ports.mentions,
      rows: foundation.snapshot.rows,
      rowsRef: foundation.refs.rows,
      selectedMentionRef: foundation.snapshot.mentions.selectedMentionRef,
      setDismissedMentionsByRow:
        foundation.commands.mentions.setDismissedMentionsByRow,
      setSelectedMentionRef: foundation.commands.mentions.setSelectedMentionRef,
      setSelectedResolveTargetId:
        foundation.commands.mentions.setSelectedResolveTargetId,
    },
    grid: {
      beginViewportContinuity:
        grid.commands.viewportContinuity.beginViewportContinuity,
      clearViewportContinuity:
        grid.commands.viewportContinuity.clearViewportContinuity,
      gridShellRef: grid.refs.gridShell,
      requireViewportContinuitySourceRecord:
        grid.commands.viewportContinuity.requireViewportContinuitySourceRecord,
      restoreTimelineFocusAnchor:
        grid.commands.anchors.restoreTimelineFocusAnchor,
      settleViewportContinuityFollowUp:
        grid.commands.viewportContinuity.settleViewportContinuityFollowUp,
      workbookFocusAnchorRef: grid.refs.workbookFocusAnchor,
    },
    incident: {
      continuityResetKey: runtime.incident.continuityResetKey,
      currentRole: runtime.incident.currentRole,
      currentUserId: runtime.incident.currentUserId,
      incidentPort: runtime.incident.incidentPort,
      inspectorResetKey: runtime.incident.inspectorResetKey,
    },
    inspector: {
      elementRegistry: inspector.ports.elements,
      history: {
        commands: inspector.commands.history,
        snapshot: inspector.snapshot.history,
      },
      lifecycle: inspector.snapshot.lifecycle,
      publishFeedback: inspector.commands.publishFeedback,
      selectRow: inspector.commands.selectRow,
      selection: inspector.snapshot.selection,
      setOpen: inspector.commands.setOpen,
    },
    mutation: {
      activeConflict: mutation.snapshot.conflict.activeConflict,
      applyAcceptedRowMutation: mutation.commands.save.applyAcceptedRowMutation,
      commands: {
        acceptTimelineRecordVersion:
          mutation.commands.save.acceptTimelineRecordVersion,
        beginSave: mutation.commands.save.beginSave,
        enqueueSaveWork: mutation.commands.save.enqueueSaveWork,
        finishSave: mutation.commands.save.finishSave,
        nextClientTxnId: mutation.commands.identity.nextClientTxnId,
        resolvePendingSocketTxn: mutation.commands.save.resolvePendingSocketTxn,
        trackPendingSocketTxn: mutation.commands.save.trackPendingSocketTxn,
      },
      loadRows: mutation.commands.query.loadRows,
      publishViewingPresence: mutation.commands.presence.publishViewingPresence,
      waitForCommittedRecordIdle: mutation.ports.waitForCommittedRecordIdle,
    },
    mutationCommands: runtime.mutationCommands,
    onIncidentAccessLost: runtime.onIncidentAccessLost,
    onRefreshEntities: runtime.entities.refresh,
  });
  const interaction = useTimelineInteractionComposition({
    foundation: {
      activateCollectionInput:
        foundation.commands.editor.activateCollectionInput,
      activeCollectionInputKey:
        foundation.snapshot.editor.activeCollectionInputKey,
      bulkTagPort: foundation.ports.bulkTag,
      clipboardPastePort: foundation.ports.clipboardPaste,
      deactivateCollectionInput:
        foundation.commands.editor.deactivateCollectionInput,
      pendingSavesRefs: foundation.refs.pendingSaves,
      recordTiming: foundation.commands.recordTiming,
      rows: foundation.snapshot.rows,
      rowsRef: foundation.refs.rows,
      setRefreshError: foundation.commands.lifecycle.setRefreshError,
      setSelectedMentionRef: foundation.commands.mentions.setSelectedMentionRef,
    },
    grid: {
      beginViewportContinuity:
        grid.commands.viewportContinuity.beginViewportContinuity,
      clearViewportContinuity:
        grid.commands.viewportContinuity.clearViewportContinuity,
      currentTimelineAnchorFor: grid.commands.anchors.currentTimelineAnchorFor,
      focusDraftRow: grid.commands.focusDraftRow,
      navigateTimelineFocusAnchor:
        grid.commands.anchors.navigateTimelineFocusAnchor,
      resolveTimelinePasteTargetResolution:
        grid.commands.anchors.resolveTimelinePasteTargetResolution,
      restoreTimelineFocusAnchor:
        grid.commands.anchors.restoreTimelineFocusAnchor,
      timelineAnchorColumnsRef: grid.refs.timelineAnchorColumns,
      updateTimelineSurfaceFocusAnchor:
        grid.commands.updateTimelineSurfaceFocusAnchor,
      workbookFocusAnchorRef: grid.refs.workbookFocusAnchor,
    },
    inspector: {
      clearRowHistory: inspector.commands.history.clearRowHistory,
      elementRegistry: inspector.ports.elements,
      publishFeedback: inspector.commands.publishFeedback,
      rowHistory: inspector.snapshot.history.rowHistory,
      selectedRowId: inspector.snapshot.selection.selectedRowId,
      selectRow: inspector.commands.selectRow,
      setOpen: inspector.commands.setOpen,
    },
    interactionMode: runtime.layout.snapshot.interactionMode,
    loadAccessLost: foundation.snapshot.lifecycle.loadAccessLost,
    mutation: {
      applyClipboardResponseRows:
        mutation.commands.save.applyClipboardResponseRows,
      beginSave: mutation.commands.save.beginSave,
      commitScalarGridEdit: mutation.commands.mutation.commitScalarGridEdit,
      enqueueSaveWork: mutation.commands.save.enqueueSaveWork,
      finishSave: mutation.commands.save.finishSave,
      loadRows: mutation.commands.query.loadRows,
      mutationCommands: runtime.mutationCommands,
      queueCollectionSave: mutation.commands.mutation.queueCollectionSave,
      queueScalarSave: mutation.commands.mutation.queueScalarSave,
      registerSameFieldConflict:
        mutation.commands.save.registerSameFieldConflict,
      resolvePendingSocketTxn: mutation.commands.save.resolvePendingSocketTxn,
      setActiveConflictKey: mutation.commands.save.setActiveConflictKey,
      setPasteConflictGroup: mutation.commands.save.setPasteConflictGroup,
      trackPendingSocketTxn: mutation.commands.save.trackPendingSocketTxn,
      waitForCommittedRecordIdle: mutation.ports.waitForCommittedRecordIdle,
    },
    queryState: foundation.snapshot.query.queryState,
    role: runtime.incident.currentRole,
    surfaceKey: sheetRefKey(mutation.ports.activeSheetRef),
    workflow: {
      handleTimelineGridContextKeyDown:
        workflow.commands.rowInteractions.handleTimelineGridContextKeyDown,
      openRowHistory: workflow.commands.history.openRowHistory,
      timelineRowForEventTarget:
        workflow.commands.rowInteractions.timelineRowForEventTarget,
    },
  });

  const presentation = {
    foundation: {
      commands: {
        query: foundation.commands.query,
      },
      refs: {
        editorDraftRegistry: foundation.refs.editorDraftRegistry,
      },
      snapshot: {
        editor: foundation.snapshot.editor,
        initialLoadGenerationKey: foundation.snapshot.initialLoadGenerationKey,
        lifecycle: foundation.snapshot.lifecycle,
        mentions: foundation.snapshot.mentions,
        pendingQueue: foundation.snapshot.pendingQueue,
        query: foundation.snapshot.query,
        rows: foundation.snapshot.rows,
      },
    },
    grid: {
      commands: {
        registerVisibleColumns: grid.commands.registerVisibleColumns,
        updateTimelineSurfaceFocusAnchor:
          grid.commands.updateTimelineSurfaceFocusAnchor,
      },
      refs: {
        gridHandle: grid.refs.gridHandle,
        gridShell: grid.refs.gridShell,
      },
      snapshot: {
        gridShellWidth: grid.snapshot.gridShellWidth,
        workbookFocusAnchor: grid.snapshot.workbookFocusAnchor,
      },
    },
    inspector: {
      commands: {
        history: {
          cancelRowHistoryPendingAction:
            workflow.commands.history.cancelRowHistoryPendingAction,
        },
        publishFeedback: inspector.commands.publishFeedback,
        setOpen: inspector.commands.setOpen,
      },
      ports: {
        elements: inspector.ports.elements,
        restoreFocus: inspector.ports.restoreFocus,
      },
      snapshot: inspector.snapshot,
    },
    interaction: {
      commands: interaction.commands,
      snapshot: interaction.snapshot,
    },
    mutation: {
      commands: {
        mutation: {
          changeReplacementDraft:
            mutation.commands.mutation.changeReplacementDraft,
          queueAction: mutation.commands.mutation.queueAction,
        },
        presence: {
          publishEditModePresence:
            mutation.commands.presence.publishEditModePresence,
        },
        query: {
          loadRows: mutation.commands.query.loadRows,
        },
      },
      snapshot: {
        collaboration: {
          activeSheetPresenceRecords:
            mutation.snapshot.collaboration.activeSheetPresenceRecords,
        },
        conflict: {
          commonMutationSnapshot:
            mutation.snapshot.conflict.commonMutationSnapshot,
          conflictQueue: mutation.snapshot.conflict.conflictQueue,
          getCellState: mutation.snapshot.conflict.getCellState,
        },
        mutation: {
          replacementDrafts: mutation.snapshot.mutation.replacementDrafts,
        },
        presence: mutation.snapshot.presence,
      },
    },
    workflow: {
      commands: workflow.commands,
      snapshot: workflow.snapshot,
    },
  };

  return {
    foundation,
    grid,
    inspector,
    interaction,
    mutation,
    presentation,
    workflow,
  };
}

export type TimelineWorkbookCompositionResult = ReturnType<
  typeof useTimelineWorkbookComposition
>;
