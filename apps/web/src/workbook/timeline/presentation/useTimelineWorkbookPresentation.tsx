import type {
  GridCellAnchor,
  GridDataRow,
  GridRowStateInput,
} from "@cartulary/grid-adapter";
import {
  gridGroupRowTestId,
  timelineMutationSubstrateReadyTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import { useCallback, useLayoutEffect, useMemo } from "react";
import { WorkbookRowGutterContent } from "../../components/WorkbookPresenceMarkers";
import { applyWorkbookLayoutToColumns } from "../../layout/workbookColumnLayout";
import {
  type WorkbookQueryLoadState,
  workbookGridDataState,
} from "../../models/workbookGridState";
import {
  defaultFilterDraft,
  emptyWorkbookQueryState,
  removeFilterField,
} from "../../models/workbookQuery";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { DraftRowCreateButton } from "../components/TimelineDraftRowActions";
import { useTimelineWorkbookInspectorSections } from "../components/TimelineWorkbookInspectorSections";
import { timelinePendingQueueMessage } from "../components/TimelineWorkbookNotices";
import { useTimelineWorkbookRenderers } from "../components/TimelineWorkbookRenderers";
import {
  timelineGridShellStyle,
  timelineRowGutterWidth,
} from "../components/TimelineWorkbookStyles";
import type { TimelineWorkbookCompositionResult } from "../composition/useTimelineWorkbookComposition";
import { buildTimelineGridRows } from "../models/timelineRowsModel";
import type { TimelineWorkbookSurfaceRuntime } from "../models/timelineWorkbookSurfaceRuntime";
import {
  timelineGroupLabel,
  timelineObservationSourceFields,
  timelineRelationshipLabel,
  type WorkbookRow,
} from "../models/workbookTimelineModel";

const timelineContract = requireViewContract(timelineViewSchemaId);
const timelineInspectorConfig = timelineContract.inspectorConfig;

export type TimelineWorkbookPresentationRuntime = {
  readonly currentIncidentRole: TimelineWorkbookSurfaceRuntime["incident"]["currentRole"];
  readonly indicatorWorkflow: TimelineWorkbookSurfaceRuntime["indicatorWorkflow"];
  readonly entities: Pick<
    TimelineWorkbookSurfaceRuntime["entities"],
    "hosts" | "identities" | "index"
  >;
  readonly layout: TimelineWorkbookSurfaceRuntime["layout"];
  readonly onActivateConflict: TimelineWorkbookSurfaceRuntime["onActivateConflict"];
  readonly queryControls: Pick<
    TimelineWorkbookSurfaceRuntime["query"],
    "renderInlineControls" | "savedViewSelector" | "viewBarQueryControls"
  >;
};

export function useTimelineWorkbookPresentation({
  composition,
  runtime,
}: {
  readonly composition: TimelineWorkbookCompositionResult["presentation"];
  readonly runtime: TimelineWorkbookPresentationRuntime;
}) {
  const { foundation, grid, inspector, interaction, mutation, workflow } =
    composition;
  const {
    currentIncidentRole,
    indicatorWorkflow,
    entities,
    layout,
    onActivateConflict,
    queryControls,
  } = runtime;
  const {
    renderInlineControls: renderInlineQueryControls,
    savedViewSelector,
    viewBarQueryControls,
  } = queryControls;
  const {
    hosts: hostEntities,
    identities: identityEntities,
    index: entityIndex,
  } = entities;
  const {
    commands: {
      onColumnHiddenChange: handleColumnHiddenChange,
      onColumnMove: handleColumnMove,
      onColumnReorder: handleColumnReorder,
      onColumnWidthChange: handleColumnWidthChange,
      onResetColumns: handleResetColumns,
    },
    snapshot: {
      chromeMode,
      density,
      incidentClosed,
      interactionMode,
      showStatusPresence,
      state: layoutState,
    },
  } = layout;
  const {
    initialLoadGenerationKey,
    lifecycle: {
      isInitialLoading,
      isRefreshing,
      loadAccessLost,
      loadError,
      refreshError,
    },
  } = foundation.snapshot;
  const {
    applyQueryFilter,
    handleQueryGroupByChange,
    handleQuerySortChange,
    setFilterDraft,
    setQueryState,
  } = foundation.commands.query;
  const { filterDraft, queryState } = foundation.snapshot.query;
  const rows = foundation.snapshot.rows;
  const getTimelineRowState = useCallback(
    (row: GridDataRow<WorkbookRow>): GridRowStateInput => ({
      pending: row.data.pendingSignature !== null,
    }),
    [],
  );
  const { autoResolutionNotices, selectedResolveTargetId } =
    foundation.snapshot.mentions;
  const pendingQueueSnapshot = foundation.snapshot.pendingQueue;
  const editorDraftRegistry = foundation.refs.editorDraftRegistry;
  const timelineGridHandleRef = grid.refs.gridHandle;
  const gridShellRef = grid.refs.gridShell;
  const timelineGridShellWidth = grid.snapshot.gridShellWidth;
  const { workbookFocusAnchor } = grid.snapshot;
  const updateTimelineSurfaceFocusAnchor =
    grid.commands.updateTimelineSurfaceFocusAnchor;

  const { activeCollectionInputKey } = interaction.snapshot.editor;
  const {
    activateConflictCell,
    activateCollectionInput,
    commitScalarGridEdit,
    deactivateCollectionInput,
    handleBlur,
    handleCollectionInputChange,
    handleCollectionKeyDown,
    handleKeyDown,
    handlePaste,
    queueCollectionSave,
  } = interaction.commands.editor;
  const {
    clipboardPaste: timelineClipboardPaste,
    focusDraftRow,
    handleCreateBlankDraftRow,
    handleFillCells,
    handleWorkAreaKeyDown: handleTimelineWorkAreaKeyDown,
  } = interaction.commands.grid;
  const {
    canAssign: canBulkTag,
    canSubmit: canSubmitBulkTag,
    gridSelection: timelineBulkSelection,
    message: bulkTagMessage,
    selectedRecordIds: selectedTimelineRecordIds,
    tagName: bulkTagName,
  } = interaction.snapshot.bulk;
  const { assignTag: assignTagToSelectedRows, changeTagName: setBulkTagName } =
    interaction.commands.bulk;
  const {
    canManageMentions,
    inspectorMessage,
    inspectorMentions,
    selectedMention,
    selectedRow,
    selectedRowId,
  } = inspector.snapshot.selection;
  const setInspectorMessage = inspector.commands.publishFeedback;
  const setIsInspectorOpen = inspector.commands.setOpen;
  const isInspectorOpen = inspector.snapshot.lifecycle.isOpen;
  const {
    currentHistoryDeleted,
    currentHistoryRecordId,
    inspectorHistorySubject,
    rowHistory,
    rowHistoryPendingAction,
  } = inspector.snapshot.history;
  const { setRowHistoryPendingAction } = inspector.commands.history;
  const handleResolveTargetChange = workflow.commands.resolveTargetChange;
  const createRelatedWorkflow = workflow.snapshot.createRelatedWorkflow;
  const timelineCreateRelatedReferenceOptions =
    workflow.snapshot.createRelatedReferenceOptions;
  const indicatorInspectorHandler = workflow.snapshot.indicatorHandler;
  const {
    cancelFeatureAction: cancelInspectorFeatureAction,
    handleFeatureAction: handleInspectorFeatureAction,
  } = workflow.commands.feature;
  const {
    submit: submitCreateRelatedWorkflow,
    updateDraft: updateCreateRelatedWorkflowDraft,
  } = workflow.commands.workflow;
  const closeInspector = workflow.commands.closeInspector;

  const { commonMutationSnapshot, conflictQueue, getCellState } =
    mutation.snapshot.conflict;
  const { loadRows } = mutation.commands.query;
  const activeSheetPresenceRecords =
    mutation.snapshot.collaboration.activeSheetPresenceRecords;
  const { editingPresenceForCell, presenceForRow } = mutation.snapshot.presence;
  const { publishEditModePresence: handleEditModePresence } =
    mutation.commands.presence;
  const { activeRowContextMenuRow, rowContextMenu } =
    workflow.snapshot.rowInteractions;
  const {
    closeRowContextMenu,
    handleSelectMention,
    handleSelectRow,
    handleTimelineGridContextMenu,
    openInspectorForRow,
  } = workflow.commands.rowInteractions;
  const { replacementDrafts } = mutation.snapshot.mutation;
  const { changeReplacementDraft, queueAction } = mutation.commands.mutation;
  const {
    confirmRowHistoryPendingAction,
    openRowHistory,
    previewRowHistoryDeleteRestore,
    previewRowHistoryRollback,
  } = workflow.commands.history;
  const {
    createEntityFromMention,
    handleUndoAutoResolutionNotice,
    submitMentionAction,
  } = workflow.commands.mentions;
  const handleTimelineEvidenceFiles = workflow.commands.evidence;

  const {
    renderTimelineCollectionInput,
    renderTimelineInspectorEditor,
    timelineColumns,
  } = useTimelineWorkbookRenderers({
    activateCollectionInput,
    activateConflictCell,
    activeCollectionInputKey,
    commitScalarGridEdit,
    conflictQueue,
    deactivateCollectionInput,
    editorDraftRegistry,
    editingPresenceForCell,
    entityIndex,
    gridShellWidth: timelineGridShellWidth,
    handleBlur,
    handleCollectionInputChange,
    handleCollectionKeyDown,
    handleEditModePresence,
    handleKeyDown,
    handlePaste,
    handleSelectMention,
    handleSelectRow,
    openInspectorForRow,
    queueCollectionSave,
    readOnly: interactionMode.kind === "read_only",
    rowGutterWidth: timelineRowGutterWidth,
    timelineContract,
    updateTimelineSurfaceFocusAnchor,
  });
  const visibleTimelineColumns = useMemo(
    () =>
      applyWorkbookLayoutToColumns(
        timelineContract,
        timelineColumns,
        layoutState,
      ),
    [layoutState, timelineColumns],
  );
  useLayoutEffect(() => {
    grid.commands.registerVisibleColumns(visibleTimelineColumns);
  }, [grid.commands, visibleTimelineColumns]);

  const timelineRowGutter = useMemo(
    () => ({
      label: "",
      width: timelineRowGutterWidth,
      minWidth: timelineRowGutterWidth,
    }),
    [],
  );
  const timelineGrid = useMemo(
    () =>
      buildTimelineGridRows({
        presenceForRow,
        renderDraftGutterContent: (row) => (
          <DraftRowCreateButton
            row={row}
            onCreate={handleCreateBlankDraftRow}
            onFilesSelected={handleTimelineEvidenceFiles}
          />
        ),
        renderSavedGutterContent: ({ ordinal, presences, recordId }) => (
          <WorkbookRowGutterContent
            ordinal={ordinal}
            presences={presences}
            recordId={recordId}
          />
        ),
        rows,
      }),
    [
      handleCreateBlankDraftRow,
      handleTimelineEvidenceFiles,
      presenceForRow,
      rows,
    ],
  );
  const timelineGridRows = timelineGrid.recordRows;
  const timelineDraftRow = timelineGrid.draftRow;
  const getTimelineGroupLabel = useCallback(
    (row: WorkbookRow, fieldKey: string) => timelineGroupLabel(row, fieldKey),
    [],
  );
  const getTimelineGroupRowTestId = useCallback(
    (fieldKey: string, value: string) =>
      gridGroupRowTestId(timelineViewSchemaId, fieldKey, value),
    [],
  );

  const {
    renderEvidenceAttachSection,
    renderInspectorFieldEditors,
    renderRelationshipEditors: renderInspectorRelationshipEditors,
    renderRowHistorySection,
    renderWorkflowSection: renderCreateRelatedWorkflowSection,
  } = useTimelineWorkbookInspectorSections({
    cancelCreateRelatedWorkflow: cancelInspectorFeatureAction,
    confirmRowHistoryPendingAction,
    createRelatedWorkflow,
    currentHistoryRecordId,
    handleTimelineEvidenceFiles,
    inspectorHistorySubject,
    openRowHistory,
    previewRowHistoryDeleteRestore,
    previewRowHistoryRollback,
    renderTimelineCollectionInput,
    renderTimelineInspectorEditor,
    rowHistory,
    rowHistoryPendingAction,
    setRowHistoryPendingAction,
    submitCreateRelatedWorkflow,
    timelineCreateRelatedReferenceOptions,
    updateCreateRelatedWorkflowDraft,
  });

  const pendingQueueDisplayMessage =
    timelinePendingQueueMessage(pendingQueueSnapshot);
  const visibleRefreshError =
    pendingQueueSnapshot.blockedEdit === null &&
    pendingQueueSnapshot.overflowMessage === null &&
    refreshError !== null &&
    refreshError !== pendingQueueDisplayMessage
      ? refreshError
      : null;
  const timelineLoadState: WorkbookQueryLoadState = isInitialLoading
    ? { generationKey: initialLoadGenerationKey, kind: "initial_loading" }
    : loadError !== null
      ? loadAccessLost
        ? { kind: "permission_denied", message: loadError }
        : { kind: "unavailable", message: loadError }
      : isRefreshing
        ? { kind: "refreshing" }
        : visibleRefreshError !== null
          ? { kind: "stale_error", message: visibleRefreshError }
          : { kind: "ready" };
  const handleClearFilters = useCallback(() => {
    setQueryState(emptyWorkbookQueryState());
    setFilterDraft(defaultFilterDraft(timelineContract));
  }, [setFilterDraft, setQueryState]);
  const handleRetry = useCallback(() => {
    void loadRows({ showLoading: true });
  }, [loadRows]);
  const timelineDataState = workbookGridDataState({
    emptyAction:
      interactionMode.kind === "editable"
        ? { label: "Add row", onInvoke: focusDraftRow }
        : undefined,
    emptyMessage: "No Timeline records have been added.",
    loadState: timelineLoadState,
    onClearFilters: handleClearFilters,
    onRetry: handleRetry,
    queryState,
    rowCount: timelineGridRows.length,
    surfaceLabel: timelineContract.title,
  });
  const handleActiveCellChange = useCallback(
    (anchor: GridCellAnchor | null) => {
      updateTimelineSurfaceFocusAnchor(
        anchor?.rowIdentity.kind === "core_record"
          ? anchor.rowIdentity.recordId
          : null,
        anchor?.fieldKey ?? "",
      );
    },
    [updateTimelineSurfaceFocusAnchor],
  );
  const handleRemoveFilter = useCallback(
    (fieldKey: string) => {
      setQueryState((current) => removeFilterField(current, fieldKey));
    },
    [setQueryState],
  );
  const handleInspectorToggle = useCallback(() => {
    setIsInspectorOpen(true);
  }, [setIsInspectorOpen]);

  return {
    grid: {
      activeRecordId: selectedRowId,
      bulkSelection: timelineBulkSelection,
      clipboardPaste: timelineClipboardPaste,
      columns: visibleTimelineColumns,
      columnWidths: layoutState.columnWidths,
      dataState: timelineDataState,
      density,
      getCellState,
      getGroupLabel: getTimelineGroupLabel,
      getGroupRowTestId: getTimelineGroupRowTestId,
      getRowState: getTimelineRowState,
      groupBy: queryState.groupBy,
      interactionMode,
      onActiveCellChange: handleActiveCellChange,
      onColumnReorder: handleColumnReorder,
      onColumnWidthChange: handleColumnWidthChange,
      onFillCells: handleFillCells,
      onSelectRecord: handleSelectRow,
      onSortChange: handleQuerySortChange,
      ref: timelineGridHandleRef,
      rowGutter: timelineRowGutter,
      rows,
      shellRef: gridShellRef,
      sort: queryState.sort,
      style: timelineGridShellStyle,
      timelineDraftRow,
      timelineGridRows,
    },
    inspector: isInspectorOpen
      ? {
          canManageMentions,
          currentHistoryDeleted,
          currentIncidentRole,
          incidentClosed,
          currentHistoryRecordId,
          entityIndex,
          getRelationshipLabel: timelineRelationshipLabel,
          hostEntities,
          identityEntities,
          indicatorInspectorHandler,
          indicatorWorkflow,
          inspectorConfig: timelineInspectorConfig,
          inspectorMentions,
          inspectorMessage,
          loadRows,
          onClose: closeInspector,
          onCreateEntityFromMention: createEntityFromMention,
          onFeatureAction: handleInspectorFeatureAction,
          onResolveTargetChange: handleResolveTargetChange,
          onSelectMention: handleSelectMention,
          onSetInspectorMessage: setInspectorMessage,
          onSubmitMentionAction: submitMentionAction,
          renderEvidenceAttachSection,
          renderInspectorFieldEditors,
          renderRelationshipEditors: renderInspectorRelationshipEditors,
          renderRowHistorySection,
          renderWorkflowSection: renderCreateRelatedWorkflowSection,
          rowHistoryRecordId: currentHistoryDeleted
            ? currentHistoryRecordId
            : null,
          rowHistoryRowVersion:
            currentHistoryDeleted &&
            rowHistory.data?.record_id === currentHistoryRecordId
              ? rowHistory.data.row_version
              : null,
          selectedMention,
          selectedResolveTargetId,
          selectedRow,
          sourceFields: timelineObservationSourceFields,
        }
      : null,
    layout: {
      chromeMode,
      onRequestInspectorClose: closeInspector,
      onWorkAreaContextMenu: handleTimelineGridContextMenu,
      onWorkAreaKeyDown: handleTimelineWorkAreaKeyDown,
      testId: timelineMutationSubstrateReadyTestId(),
      viewSchemaId: timelineViewSchemaId,
      workAreaAriaLabel: "Timeline row interaction layer",
    },
    overlays: {
      contextMenu:
        rowContextMenu === null
          ? null
          : {
              position: rowContextMenu.position,
              replacementDraft:
                activeRowContextMenuRow === null
                  ? ""
                  : (replacementDrafts[activeRowContextMenuRow.key] ?? ""),
              row: activeRowContextMenuRow,
              onClose: closeRowContextMenu,
              onInspectRow: openInspectorForRow,
              onMarkReviewed: (rowKey: string) => {
                queueAction(rowKey, "mark-reviewed");
              },
              onOpenHistory: openRowHistory,
              onReplacementDraftChange: changeReplacementDraft,
              onSupersede: (rowKey: string) => {
                queueAction(rowKey, "supersede");
              },
            },
      notices: {
        autoResolutionNotices,
        entityIndex,
        inspectorOpen: isInspectorOpen,
        onReviewAutoResolution: handleSelectMention,
        onUndoAutoResolution: handleUndoAutoResolutionNotice,
        pendingQueueSnapshot,
      },
    },
    status: {
      activeSheetPresenceRecords,
      inFlightCount: pendingQueueSnapshot.inFlightCount,
      onActivateConflict,
      queuedCount: pendingQueueSnapshot.queuedCount,
      saveState: commonMutationSnapshot.primaryLabel,
      saveStateSecondaryMessage: commonMutationSnapshot.secondaryMessage,
      showPresence: showStatusPresence,
      workbookFocusAnchor,
    },
    viewBar: {
      addRowDisabled: interactionMode.kind === "read_only",
      chromeMode,
      bulk:
        selectedTimelineRecordIds.size === 0
          ? null
          : {
              canAssign: canBulkTag,
              canSubmit: canSubmitBulkTag,
              message: bulkTagMessage,
              onAssign: assignTagToSelectedRows,
              onTagNameChange: setBulkTagName,
              selectedCount: selectedTimelineRecordIds.size,
              tagName: bulkTagName,
            },
      inlineQuery: renderInlineQueryControls
        ? {
            chromeMode,
            contract: timelineContract,
            defaultFilterPopoverOpen: true,
            filterDraft,
            layoutState,
            onApplyFilter: applyQueryFilter,
            onClearAll: handleClearFilters,
            onColumnHiddenChange: handleColumnHiddenChange,
            onColumnMove: handleColumnMove,
            onFilterDraftChange: setFilterDraft,
            onGroupByChange: handleQueryGroupByChange,
            onRemoveFilter: handleRemoveFilter,
            onResetColumns: handleResetColumns,
            onSortChange: handleQuerySortChange,
            queryState,
            surface: timelineViewSchemaId,
          }
        : null,
      onAddRow: focusDraftRow,
      onInspectorToggle: handleInspectorToggle,
      savedViewControls: savedViewSelector,
      surface: timelineViewSchemaId,
      viewBarQueryControls,
    },
  };
}

export type TimelineWorkbookPresentationModel = ReturnType<
  typeof useTimelineWorkbookPresentation
>;
