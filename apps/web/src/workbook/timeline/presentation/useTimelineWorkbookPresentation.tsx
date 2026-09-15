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
import {
  useCallback,
  useLayoutEffect,
  useMemo,
  useRef,
  useSyncExternalStore,
} from "react";
import { WorkbookParkedGridDrafts } from "../../components/WorkbookParkedGridDrafts";
import { WorkbookRowGutterContent } from "../../components/WorkbookPresenceMarkers";
import { EvidenceFileRecovery } from "../../features/evidence/EvidenceFileRecovery";
import { admitEvidenceFile } from "../../features/evidence/evidenceFileOperation";
import { useWorkbookSemanticGridFocus } from "../../hooks/useWorkbookSemanticGridFocus";
import { applyWorkbookLayoutToColumns } from "../../layout/workbookColumnLayout";
import {
  type WorkbookQueryLoadState,
  workbookGridDataState,
} from "../../models/workbookGridState";
import { workbookInspectorStateIsOpen } from "../../models/workbookInspectorModel";
import {
  defaultFilterDraft,
  removeFilterField,
} from "../../models/workbookQuery";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { DraftRowCreateButton } from "../components/TimelineDraftRowActions";
import { timelinePendingQueueMessage } from "../components/TimelineWorkbookNotices";
import { useTimelineWorkbookRenderers } from "../components/TimelineWorkbookRenderers";
import {
  timelineGridShellStyle,
  timelineRowGutterWidth,
} from "../components/TimelineWorkbookStyles";
import type { TimelineWorkbookCompositionResult } from "../composition/useTimelineWorkbookComposition";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import { timelineRelationshipLabel } from "../models/timelineFieldRegistry";
import { timelineGroupLabel } from "../models/timelineLayoutPolicy";
import type { WorkbookRow } from "../models/timelineRowModel";
import { buildTimelineGridRows } from "../models/timelineRowsModel";
import type { TimelineWorkbookSurfaceRuntime } from "../models/timelineWorkbookSurfaceRuntime";
import { useTimelineInspectorPresentation } from "./useTimelineInspectorPresentation";

const timelineContract = requireViewContract(timelineViewSchemaId);
const timelineInspectorConfig = timelineContract.inspectorConfig;

export type TimelineWorkbookPresentationRuntime = {
  readonly currentIncidentRole: TimelineWorkbookSurfaceRuntime["incident"]["currentRole"];
  readonly gridEntryFocus: TimelineWorkbookSurfaceRuntime["gridEntryFocus"];
  readonly entities: Pick<
    TimelineWorkbookSurfaceRuntime["entities"],
    "hosts" | "identities" | "index"
  >;
  readonly layout: TimelineWorkbookSurfaceRuntime["layout"];
  readonly onActivateConflict: TimelineWorkbookSurfaceRuntime["onActivateConflict"];
  readonly queryControls: Pick<
    TimelineWorkbookSurfaceRuntime["query"],
    "renderInlineControls" | "viewBarWorkingSet"
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
    gridEntryFocus,
    entities,
    layout,
    onActivateConflict,
    queryControls,
  } = runtime;
  const { renderInlineControls: renderInlineQueryControls, viewBarWorkingSet } =
    queryControls;
  const { index: entityIndex } = entities;
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
  const fileOwner = composition.fileOwner;
  const files = useSyncExternalStore(
    fileOwner.subscribe,
    fileOwner.getSnapshot,
  );
  const fileMessage = useSyncExternalStore(
    fileOwner.subscribe,
    fileOwner.getAdmissionNotice,
  );
  const fileAnchor = useRef<GridCellAnchor | null>(null);
  const getTimelineRowState = useCallback(
    (row: GridDataRow<WorkbookRow>): GridRowStateInput => ({
      pending: row.data.pendingSignature !== null,
    }),
    [],
  );
  const { autoResolutionNotices } = foundation.snapshot.mentions;
  const pendingQueueSnapshot = foundation.snapshot.pendingQueue;
  const editorDraftRegistry = foundation.refs.editorDraftRegistry;
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
  const isInspectorOpen = workbookInspectorStateIsOpen(
    inspector.snapshot.lifecycle,
  );
  const {
    currentHistoryDeleted,
    currentHistoryRecordId,
    inspectorHistorySubject,
    rowHistory,
  } = inspector.snapshot.history;
  const { cancelRowHistoryPendingAction } = inspector.commands.history;
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
  const presence = mutation.snapshot.collaboration.presence.header;
  const { editingPresenceForCell, presenceForRow } = mutation.snapshot.presence;
  const { publishEditModePresence: handleEditModePresence } =
    mutation.commands.presence;
  const {
    activeRowContextMenuRow,
    rowContextMenu,
    rowContextMenuFallbackFocusRef,
    rowContextMenuReturnFocusRef,
  } = workflow.snapshot.rowInteractions;
  const {
    closeRowContextMenu,
    handleInspectCollection,
    handleSelectMention,
    handleSelectRow,
    handleTimelineGridContextMenu,
    openInspectorForRow,
  } = workflow.commands.rowInteractions;
  const captureActions = composition.captureActions;
  const {
    confirmRowHistoryPendingAction,
    openRowHistory,
    previewRowHistoryDeleteRestore,
    previewRowHistoryRollback,
    historyBrowsingControls,
  } = workflow.commands.history;
  const { handleUndoAutoResolutionNotice } = workflow.commands.mentions;
  const handleTimelineEvidenceFiles = workflow.commands.evidence;

  const {
    renderTimelineCollectionInput,
    renderTimelineInspectorEditor,
    timelineColumns,
  } = useTimelineWorkbookRenderers({
    activateCollectionInput,
    activateConflictCell,
    activeCollectionInputKey,
    elementRegistry: inspector.ports.elements,
    handleInspectCollection,
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
    handleSelectRow,
    queueCollectionSave,
    readOnly: interactionMode.kind === "read_only",
    readCurrentRow: (row) =>
      rows.find((candidate) => candidate.key === row.key) ?? row,
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

  const timelineInspector = useTimelineInspectorPresentation({
    currentHistoryDeleted,
    currentHistoryRecordId,
    isOpen: isInspectorOpen,
    model: {
      canManageMentions,
      currentHistoryDeleted,
      currentIncidentRole,
      elementRegistry: inspector.ports.elements,
      incidentClosed,
      currentHistoryRecordId,
      entityIndex,
      getRelationshipLabel: timelineRelationshipLabel,
      indicatorInspectorHandler,
      beginMutation: composition.mutation.commands.beginMutation,
      inspectorConfig: timelineInspectorConfig,
      inspectorMentions,
      inspectorMessage,
      loadRows,
      onClose: closeInspector,
      mentionActions: workflow.commands.mentions,
      captureEditor: captureActions.editor,
      captureResult: captureActions.result,
      additionalDisabledReasons: captureActions.additionalDisabledReasons,
      onFeatureAction: (
        capability: Parameters<typeof handleInspectorFeatureAction>[0],
      ) => {
        if (capability.kind === "timeline_capture")
          captureActions.activate(
            capability.featureGroup.featureGroupKey === "timeline.mark_reviewed"
              ? "mark-reviewed"
              : "supersede",
          );
        else handleInspectorFeatureAction(capability);
      },
      onSelectMention: handleSelectMention,
      onSetInspectorMessage: setInspectorMessage,
      selectedMention,
      selectedRow,
      observationSource: composition.observationSource,
    },
    sections: {
      cancelCreateRelatedWorkflow: cancelInspectorFeatureAction,
      cancelRowHistoryPendingAction,
      canMutateHistory:
        !incidentClosed &&
        currentIncidentRole !== null &&
        currentIncidentRole !== "viewer",
      confirmRowHistoryPendingAction,
      createRelatedWorkflow,
      handleTimelineEvidenceFiles,
      inspectorHistorySubject,
      openRowHistory,
      previewRowHistoryDeleteRestore,
      previewRowHistoryRollback,
      historyBrowsingControls,
      renderTimelineCollectionInput,
      renderTimelineInspectorEditor,
      rowHistory,
      submitCreateRelatedWorkflow,
      timelineCreateRelatedReferenceOptions,
      updateCreateRelatedWorkflowDraft,
    },
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
    setQueryState((current) =>
      current.filters.length === 0 ? current : { ...current, filters: [] },
    );
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
  const timelineDraftFieldKeys = useMemo(
    () =>
      timelineDraftRow === undefined
        ? []
        : visibleTimelineColumns
            .filter((column) => column.renderDraftCell !== undefined)
            .map((column) => column.fieldKey),
    [timelineDraftRow, visibleTimelineColumns],
  );
  const registerTimelineGridHandle = useWorkbookSemanticGridFocus({
    dataRows: timelineGridRows,
    dataState: timelineDataState,
    draftFieldKeys: timelineDraftFieldKeys,
    focusOwner: gridEntryFocus,
    gridHandleRef: grid.refs.gridHandle,
    visibleColumns: visibleTimelineColumns,
    viewSchemaId: timelineViewSchemaId,
  });
  const handleActiveCellChange = useCallback(
    (anchor: GridCellAnchor | null) => {
      fileAnchor.current = anchor;
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
      parkedDrafts:
        !loadAccessLost && currentIncidentRole ? (
          <TimelineParkedGridDrafts
            registry={editorDraftRegistry}
            canEdit={interactionMode.kind === "editable" && !incidentClosed}
            rows={rows}
            fieldKeys={visibleTimelineColumns.map((column) => column.fieldKey)}
          />
        ) : null,
      fileRecovery:
        !loadAccessLost && currentIncidentRole ? (
          <div style={{ maxBlockSize: "25%", overflow: "auto" }}>
            {fileMessage ? <div role="status">{fileMessage}</div> : null}
            {files.map((entry) => (
              <EvidenceFileRecovery
                {...entry}
                key={entry.key}
                source={entry.sourceLabel}
                onConfirmReview={() => fileOwner.confirmReview(entry.key)}
                onReview={() => void fileOwner.review(entry.key)}
                onResume={() => void fileOwner.resume(entry.key)}
                onFreshSlot={() => fileOwner.freshSlot(entry.key)}
                onNewId={() => fileOwner.newRequestId(entry.key)}
                onDiscard={() => fileOwner.discard(entry.key)}
                onRefresh={() => void fileOwner.refresh(entry.key)}
              />
            ))}
          </div>
        ) : null,
      onFilesSelected: (selected: File[], editorRowKey?: string) => {
        const admission = admitEvidenceFile(selected);
        if (admission.kind === "rejected") {
          fileOwner.reportAdmission(admission.message);
          return;
        }
        if (admission.kind === "empty") return;
        if (
          interactionMode.kind !== "editable" ||
          incidentClosed ||
          loadAccessLost
        )
          return;
        const identity = fileAnchor.current?.rowIdentity;
        const target = editorRowKey
          ? rows.find((row) => row.key === editorRowKey)
          : identity?.kind === "core_record"
            ? rows.find((row) => row.recordId === identity.recordId)
            : rows.find((row) => row.recordId === selectedRowId);
        if (!target) {
          fileOwner.reportAdmission(
            "Select the Timeline row or draft for this file.",
          );
          return;
        }
        fileOwner.begin(
          {
            label:
              target.values.activitySynopsisText ||
              target.values.rawActivityText ||
              "Timeline draft",
            key: target.key,
            recordId: target.recordId,
            rowVersion: target.rowVersion,
          },
          selected,
        );
      },
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
      ref: registerTimelineGridHandle,
      rowGutter: timelineRowGutter,
      rows,
      shellRef: gridShellRef,
      sort: queryState.sort,
      style: timelineGridShellStyle,
      timelineDraftRow,
      timelineGridRows,
    },
    inspector: timelineInspector,
    layout: {
      chromeMode,
      onRequestInspectorClose: closeInspector,
      restoreInspectorFocus: inspector.ports.restoreFocus,
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
              fallbackFocusTargetRef: rowContextMenuFallbackFocusRef,
              reviewDisabledReason: captureActions.reason(
                activeRowContextMenuRow,
                "mark-reviewed",
              ),
              supersedeDisabledReason: captureActions.reason(
                activeRowContextMenuRow,
                "supersede",
              ),
              row: activeRowContextMenuRow,
              returnFocusTargetRef: rowContextMenuReturnFocusRef,
              onClose: closeRowContextMenu,
              onInspectRow: openInspectorForRow,
              onMarkReviewed: (rowKey: string) => {
                captureActions.start(rowKey, "mark-reviewed");
              },
              onOpenHistory: openRowHistory,
              onSupersede: (rowKey: string) => {
                captureActions.start(rowKey, "supersede");
              },
            },
      notices: {
        canManageMentions:
          canManageMentions &&
          !incidentClosed &&
          interactionMode.kind !== "read_only",
        autoResolutionNotices,
        entityIndex,
        inspectorOpen: isInspectorOpen,
        onReviewAutoResolution: handleSelectMention,
        onUndoAutoResolution: handleUndoAutoResolutionNotice,
        pendingQueueSnapshot,
      },
    },
    status: {
      presence,
      onActivateConflict,
      status: commonMutationSnapshot,
      chromeMode,
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
      workingSet: renderInlineQueryControls
        ? {
            savedView: viewBarWorkingSet?.savedView ?? null,
            query: {
              contract: timelineContract,
              defaultFilterPopoverOpen: true,
              filterDraft,
              layoutState,
              onApplyFilter: applyQueryFilter,
              onClearFilters: handleClearFilters,
              onColumnHiddenChange: handleColumnHiddenChange,
              onColumnMove: handleColumnMove,
              onFilterDraftChange: setFilterDraft,
              onGroupByChange: handleQueryGroupByChange,
              onRemoveFilter: handleRemoveFilter,
              onResetColumns: handleResetColumns,
              onSortChange: handleQuerySortChange,
              queryState,
              surface: timelineViewSchemaId,
            },
          }
        : viewBarWorkingSet,
      onAddRow: focusDraftRow,
      onInspectorToggle: handleInspectorToggle,
      surface: timelineViewSchemaId,
    },
  };
}

export type TimelineWorkbookPresentationModel = ReturnType<
  typeof useTimelineWorkbookPresentation
>;

function TimelineParkedGridDrafts({
  registry,
  canEdit,
  rows,
  fieldKeys,
}: {
  registry: TimelineEditorDraftRegistry;
  canEdit: boolean;
  rows: readonly WorkbookRow[];
  fieldKeys: readonly string[];
}) {
  useSyncExternalStore(registry.subscribe, registry.getSnapshot);
  return (
    <WorkbookParkedGridDrafts
      drafts={registry.retainedGridDrafts().flatMap((draft) => {
        const reason = !canEdit
          ? "Editing is unavailable."
          : !rows.some((row) => row.key === draft.rowKey)
            ? "Original row is outside this result."
            : !fieldKeys.includes(draft.fieldKey)
              ? "Original field is unavailable."
              : null;
        return reason
          ? [
              {
                ...draft,
                recordId: draft.rowKey,
                label:
                  timelineContract.fieldMap[draft.fieldKey]?.label ??
                  draft.fieldKey,
                reason,
              },
            ]
          : [];
      })}
    />
  );
}
