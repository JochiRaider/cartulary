import { workbookInspectorStateIsOpen } from "../../models/workbookInspectorModel";
import { useTimelineCaptureActions } from "../actions/useTimelineCaptureActions";
import { useTimelineFind } from "../hooks/useTimelineFind";
import { useTimelineObservationSource } from "../hooks/useTimelineObservationSource";
import { useTimelineSourceWriteCoordination } from "../hooks/useTimelineSourceWriteCoordination";
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
    committedRecords: foundation.ports.committedRows,
    continuity: grid.ports.continuity,
    currentIncidentRole: runtime.incident.currentRole,
    dismissedMentionsByRow: foundation.snapshot.mentions.dismissedMentionsByRow,
    observedMentions: foundation.snapshot.mentions.observedMentions,
    inspectorResetKey: runtime.incident.inspectorResetKey,
    rows: foundation.snapshot.rows,
    selectedMentionRef: foundation.snapshot.mentions.selectedMentionRef,
    workbookFocusAnchorRef: grid.refs.workbookFocusAnchor,
  });
  const mutation = useTimelineMutationComposition({
    collaborationProjection: runtime.collaborationProjection,
    foundation: {
      committedRows: foundation.ports.committedRows,
      editorDraftRegistry: foundation.refs.editorDraftRegistry,
      loadAccessLost: foundation.snapshot.lifecycle.loadAccessLost,
      nextDraftIndex: foundation.commands.rows.allocateDraftIndex,
      pendingSavesRefs: foundation.refs.pendingSaves,
      recordWorkbookTiming: foundation.commands.recordTiming,
      rowStoreCommands: foundation.commands.rows,
      rowsRef: foundation.refs.rows,
      setAutoResolutionNotices:
        foundation.commands.mentions.setAutoResolutionNotices,
      setInitialLoadGenerationKey:
        foundation.commands.lifecycle.setInitialLoadGenerationKey,
      setIsInitialLoading: foundation.commands.lifecycle.setIsInitialLoading,
      setIsRefreshing: foundation.commands.lifecycle.setIsRefreshing,
      setLoadAccessLost: foundation.commands.lifecycle.setLoadAccessLost,
      setLoadError: foundation.commands.lifecycle.setLoadError,
      setRefreshError: foundation.commands.lifecycle.setRefreshError,
      setMutationError: foundation.commands.lifecycle.setMutationError,
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
    onAuthorityUncertain: runtime.onAuthorityUncertain,
    query: {
      queryState: foundation.snapshot.query.queryState,
      viewQuery: runtime.query.viewQuery,
    },
  });
  const workflow = useTimelineInspectorWorkflowComposition({
    mentionOwner: foundation.ports.mentionOwner,
    mentionCandidates: foundation.ports.mentionCandidates,
    earlierSaves: foundation.refs.pendingSaves.saveQueueRef,
    activeSheetRef: mutation.ports.activeSheetRef,
    foundation: {
      evidenceAttachmentPort: foundation.ports.evidenceAttachment,
      loadAccessLost: foundation.snapshot.lifecycle.loadAccessLost,
      selectedTargetId: foundation.snapshot.mentions.selectedResolveTargetId,
      rows: foundation.snapshot.rows,
      rowsRef: foundation.refs.rows,
      selectedMentionRef: foundation.snapshot.mentions.selectedMentionRef,
      setSelectedMentionRef: foundation.commands.mentions.setSelectedMentionRef,
      setSelectedResolveTargetId:
        foundation.commands.mentions.setSelectedResolveTargetId,
    },
    grid: {
      gridShellRef: grid.refs.gridShell,

      restoreTimelineFocusAnchor:
        grid.commands.anchors.restoreTimelineFocusAnchor,

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
      commands: {
        acceptTimelineRecordVersion:
          mutation.commands.save.acceptTimelineRecordVersion,
        enqueueSaveWork: mutation.commands.save.enqueueSaveWork,
      },
      loadRows: mutation.commands.query.loadRows,
      publishViewingPresence: mutation.commands.presence.publishViewingPresence,
      waitForCommittedRecordIdle: mutation.ports.waitForCommittedRecordIdle,
    },
    onAuthorityUncertain: runtime.onAuthorityUncertain,
  });
  const interaction = useTimelineInteractionComposition({
    foundation: {
      editorDraftRegistry: foundation.refs.editorDraftRegistry,
      bulkTagPort: foundation.ports.bulkTag,
      clipboardPastePort: foundation.ports.clipboardPaste,
      pendingSavesRefs: foundation.refs.pendingSaves,
      recordTiming: foundation.commands.recordTiming,
      rows: foundation.snapshot.rows,
      rowsRef: foundation.refs.rows,
      setOperationError: foundation.commands.lifecycle.setOperationError,
      setSelectedMentionRef: foundation.commands.mentions.setSelectedMentionRef,
    },
    grid: {
      currentTimelineAnchorFor: grid.commands.anchors.currentTimelineAnchorFor,
      focusDraftRow: grid.commands.focusDraftRow,
      prepareTimelineCollectionNavigation:
        grid.commands.anchors.prepareTimelineCollectionNavigation,
      navigateTimelineDraftFocus:
        grid.commands.anchors.navigateTimelineDraftFocus,
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
      activateConflict: mutation.commands.save.activateConflict,
      commitScalarGridEdit: mutation.commands.mutation.commitScalarGridEdit,
      mutationCommands: runtime.mutationCommands,
      queueCollectionSave: mutation.commands.mutation.queueCollectionSave,
      queueScalarSave: mutation.commands.mutation.queueScalarSave,
    },
    queryState: foundation.snapshot.query.queryState,
    role: runtime.incident.currentRole,
    workflow: {
      handleTimelineGridContextKeyDown:
        workflow.commands.rowInteractions.handleTimelineGridContextKeyDown,
      openRowHistory: workflow.commands.history.openRowHistory,
      timelineRowForEventTarget:
        workflow.commands.rowInteractions.timelineRowForEventTarget,
    },
  });

  useTimelineSourceWriteCoordination({
    committedRow: foundation.ports.committedRows.currentCommittedTimelineRow,
    owner: runtime.mutationRuntime.timelineFiles,
    rows: foundation.refs.rows,
    drafts: foundation.refs.editorDraftRegistry,
    available: !foundation.snapshot.lifecycle.loadAccessLost,
    waitForIdle: mutation.ports.waitForCommittedRecordIdle,
  });
  useTimelineSourceWriteCoordination({
    committedRow: foundation.ports.committedRows.currentCommittedTimelineRow,
    owner: runtime.mutationRuntime.timelineRelatedEvidence,
    rows: foundation.refs.rows,
    drafts: foundation.refs.editorDraftRegistry,
    available: !foundation.snapshot.lifecycle.loadAccessLost,
    waitForIdle: mutation.ports.waitForCommittedRecordIdle,
  });
  useTimelineSourceWriteCoordination({
    committedRow: foundation.ports.committedRows.currentCommittedTimelineRow,
    owner: runtime.mutationRuntime.noteCreate,
    rows: foundation.refs.rows,
    drafts: foundation.refs.editorDraftRegistry,
    available: !foundation.snapshot.lifecycle.loadAccessLost,
    waitForIdle: mutation.ports.waitForCommittedRecordIdle,
  });
  useTimelineSourceWriteCoordination({
    committedRow: foundation.ports.committedRows.currentCommittedTimelineRow,
    owner: runtime.mutationRuntime.coordinationCreate,
    rows: foundation.refs.rows,
    drafts: foundation.refs.editorDraftRegistry,
    available: !foundation.snapshot.lifecycle.loadAccessLost,
    waitForIdle: mutation.ports.waitForCommittedRecordIdle,
  });
  const captureActions = useTimelineCaptureActions({
    runtime: runtime.mutationRuntime,
    selectedRow: inspector.snapshot.selection.selectedRow,
    selectedId: inspector.snapshot.selection.selectedRowId,
    isOpen: workbookInspectorStateIsOpen(inspector.snapshot.lifecycle),
    deleted:
      inspector.snapshot.selection.selectedRowWorkflowSubject?.kind !== "live",
    concealed: foundation.snapshot.lifecycle.loadAccessLost,
    originKey: runtime.incident.inspectorResetKey,
    rowsRef: foundation.refs.rows,
    drafts: foundation.refs.editorDraftRegistry,
    pending: foundation.refs.pendingSaves,
    openHistory: workflow.commands.history.openRowHistory,
    waitForIdle: mutation.ports.waitForCommittedRecordIdle,
    loadRows: mutation.commands.query.loadRows,
    acceptReceipt: (receipt, baseVersion) => {
      const data = receipt.data;
      mutation.commands.save.acceptTimelineActionResult({
        captureState: data.capture_state,
        changeSetId: data.change_set_id,
        incidentId: data.incident_id,
        reason: data.reason,
        recordId: data.record_id,
        replacementRecordId: data.replacement_record_id,
        rowVersion: data.row_version,
      });
      foundation.commands.rows.updateRows((rows) =>
        rows.map((row) =>
          row.recordId === data.record_id &&
          row.rawRow &&
          row.rowVersion === baseVersion
            ? {
                ...row,
                captureState: data.capture_state,
                rowVersion: data.row_version,
                rawRow: {
                  ...row.rawRow,
                  row_version: data.row_version,
                  cells: {
                    ...row.rawRow.cells,
                    "timeline.capture_state": { value: data.capture_state },
                    "timeline.replacement_record_id": {
                      value: data.replacement_record_id,
                    },
                  },
                },
              }
            : row,
        ),
      );
    },
  });
  const observationSource = useTimelineObservationSource({
    runtime: runtime.mutationRuntime,
    selectedRow: inspector.snapshot.selection.selectedRow,
    available:
      workbookInspectorStateIsOpen(inspector.snapshot.lifecycle) &&
      workflow.snapshot.indicatorHandler?.action ===
        "indicator.observations.manage" &&
      inspector.snapshot.selection.selectedRowWorkflowSubject?.kind ===
        "live" &&
      !foundation.snapshot.lifecycle.loadAccessLost,
    rowsRef: foundation.refs.rows,
    drafts: foundation.refs.editorDraftRegistry,
    waitForIdle: mutation.ports.waitForCommittedRecordIdle,
  });
  const find = useTimelineFind({
    runtime,
    browser: mutation.commands.query.browser,
    rows: foundation.snapshot.rows,
    registry: foundation.refs.editorDraftRegistry,
    accessLost: foundation.snapshot.lifecycle.loadAccessLost,
    stale: foundation.snapshot.lifecycle.refreshError !== null,
    interruptViewportContinuity:
      grid.commands.viewportContinuity.interruptViewportContinuity,
    queueScalarSave: mutation.commands.mutation.queueScalarSave,
    queueCollectionSave: mutation.commands.mutation.queueCollectionSave,
  });
  const presentation = {
    find,
    observationSource,
    captureActions,
    fileOwner: runtime.mutationRuntime.timelineFiles,
    foundation: {
      commands: {
        query: foundation.commands.query,
      },
      refs: {
        editorDraftRegistry: foundation.refs.editorDraftRegistry,
        rows: foundation.refs.rows,
      },
      snapshot: {
        initialLoadGenerationKey: foundation.snapshot.initialLoadGenerationKey,
        lifecycle: foundation.snapshot.lifecycle,
        mentions: foundation.snapshot.mentions,
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
      statusSource: runtime.mutationRuntime.statusSource,
      sheetRef: runtime.incident.sheetRef,
      commands: {
        beginMutation: mutation.commands.save.beginSave,
        presence: {
          publishEditModePresence:
            mutation.commands.presence.publishEditModePresence,
        },
        query: {
          loadRows: mutation.commands.query.loadRows,
          browser: mutation.commands.query.browser,
        },
      },
      snapshot: {
        cellRangeScopeKey: mutation.snapshot.cellRangeScopeKey,
        collaboration: {
          presence: mutation.snapshot.collaboration.presence,
        },
        conflict: {
          conflictQueue: mutation.snapshot.conflict.conflictQueue,
          getCellState: mutation.snapshot.conflict.getCellState,
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
