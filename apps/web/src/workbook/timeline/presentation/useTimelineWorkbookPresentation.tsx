import type {
  GridCellAnchor,
  GridDataRow,
  GridHandle,
  GridRowStateInput,
} from "@cartulary/grid-adapter";
import {
  gridGroupRowTestId,
  timelineMutationSubstrateReadyTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import {
  type MouseEvent,
  useCallback,
  useLayoutEffect,
  useMemo,
  useSyncExternalStore,
} from "react";
import { WorkbookParkedGridDrafts } from "../../components/WorkbookParkedGridDrafts";
import { WorkbookRowGutterContent } from "../../components/WorkbookPresenceMarkers";
import { useWorkbookSemanticGridFocus } from "../../hooks/useWorkbookSemanticGridFocus";
import { useWorkbookColumnSizingBinding } from "../../layout/useWorkbookColumnSizingBinding";
import {
  applyWorkbookLayoutToColumns,
  workbookFrozenDataColumnPrefix,
} from "../../layout/workbookColumnLayout";
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
import { useWorkbookQueryRestart } from "../../query/WorkbookQueryBrowsingContext";
import { DraftRowCreateButton } from "../components/TimelineDraftRowActions";
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
  const { foundation, grid, inspector, interaction, mutation, workflow, find } =
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
      onColumnSizingIntent: handleColumnSizingIntent,
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
      operationError,
    },
  } = foundation.snapshot;
  const {
    applyQueryFilter,
    handleQueryGroupByChange,
    handleQuerySortChange,
    setFilterDraft,
    setQueryState,
  } = foundation.commands.query;
  const { filterDraft, queryState: requestedQueryState } =
    foundation.snapshot.query;
  const queryState =
    mutation.commands.query.browser?.presentationQuery(requestedQueryState) ??
    requestedQueryState;
  const rows = foundation.snapshot.rows;
  const fileOwner = composition.fileOwner;
  const getTimelineRowState = useCallback(
    (row: GridDataRow<WorkbookRow>): GridRowStateInput => ({
      pending: row.data.pendingSignature !== null,
    }),
    [],
  );
  const editorDraftRegistry = foundation.refs.editorDraftRegistry;
  const currentRows = foundation.refs.rows;
  const readCurrentRow = useCallback(
    (row: WorkbookRow) =>
      currentRows.current.find((candidate) => candidate.key === row.key) ?? row,
    [currentRows],
  );
  const gridShellRef = grid.refs.gridShell;
  const timelineGridShellWidth = grid.snapshot.gridShellWidth;
  const { workbookFocusAnchor } = grid.snapshot;
  const updateTimelineSurfaceFocusAnchor =
    grid.commands.updateTimelineSurfaceFocusAnchor;

  const {
    activateConflictCell,
    commitScalarGridEdit,
    handleBlur,
    handleCollectionKeyDown,
    handleKeyDown,
    queueCollectionSave,
  } = interaction.commands.editor;
  const {
    clipboardPaste: timelineClipboardPaste,
    focusDraftRow,
    handleCreateBlankDraftRow,
    handleFillCells,
    handleClearCells,
    handleWorkAreaKeyDown: handleTimelineWorkAreaKeyDown,
  } = interaction.commands.grid;
  const timelineBulkSelection = interaction.snapshot.bulk.gridSelection;
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
  const indicatorInspectorHandler = workflow.snapshot.indicatorHandler;
  const { handleFeatureAction: handleInspectorFeatureAction } =
    workflow.commands.feature;

  const closeInspector = workflow.commands.closeInspector;

  const { conflictQueue, getCellState } = mutation.snapshot.conflict;
  const findCellMatch = find.cellMatch;
  const getFindCellState = useCallback(
    (input: { recordId: string; fieldKey: string }) => ({
      ...getCellState(input),
      findMatch: findCellMatch({
        surface: { kind: "view_schema", viewSchemaId: timelineViewSchemaId },
        rowIdentity: { kind: "core_record", recordId: input.recordId },
        fieldKey: input.fieldKey,
      }),
    }),
    [getCellState, findCellMatch],
  );

  const { loadRows } = mutation.commands.query;
  const presence = mutation.snapshot.collaboration.presence.header;
  const { editingPresenceForCell, presenceForRow } = mutation.snapshot.presence;
  const { publishEditModePresence: handleEditModePresence } =
    mutation.commands.presence;
  const rowContextMenu = workflow.snapshot.rowMenu;
  const { handleTimelineGridContextMenu, handleTimelineGridPointerDown } =
    workflow.commands.rowMenu;
  const {
    handleInspectCollection,
    handleSelectMention,
    handleSelectRow,
    openInspectorForRow,
    requestRowPanelFocus,
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

  const { renderTimelineCollectionInput, timelineColumns } =
    useTimelineWorkbookRenderers({
      activateConflictCell,
      elementRegistry: inspector.ports.elements,
      handleInspectCollection,
      commitScalarGridEdit,
      conflictQueue,
      editorDraftRegistry,
      editingPresenceForCell,
      entityIndex,
      gridShellWidth: timelineGridShellWidth,
      handleBlur,
      handleCollectionKeyDown,
      handleEditModePresence,
      handleKeyDown,
      handleSelectRow,
      queueCollectionSave,
      readOnly: interactionMode.kind === "read_only",
      readCurrentRow,
      rowGutterWidth: timelineRowGutterWidth,
      timelineContract,
      updateTimelineSurfaceFocusAnchor,
    });
  useWorkbookColumnSizingBinding({
    columns: timelineColumns,
    commands: layout.commands,
    gridHandleRef: grid.refs.gridHandle,
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
      detailsOwner: composition.inspectorDetails,
      rowHistory,
    },
  });

  const timelineLoadState: WorkbookQueryLoadState = isInitialLoading
    ? { generationKey: initialLoadGenerationKey, kind: "initial_loading" }
    : loadError !== null
      ? loadAccessLost
        ? { kind: "permission_denied", message: loadError }
        : { kind: "unavailable", message: loadError }
      : isRefreshing
        ? { kind: "refreshing" }
        : refreshError !== null
          ? { kind: "stale_error", message: refreshError }
          : { kind: "ready" };
  const handleClearFilters = useCallback(() => {
    setQueryState((current) =>
      current.filters.length === 0 ? current : { ...current, filters: [] },
    );
    setFilterDraft(defaultFilterDraft(timelineContract));
  }, [setFilterDraft, setQueryState]);
  const restartQuery = useWorkbookQueryRestart(timelineViewSchemaId);
  const handleRetry = useCallback(() => {
    void restartQuery();
  }, [restartQuery]);
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
  const registerFindGrid = useCallback(
    (handle: GridHandle | null) => {
      registerTimelineGridHandle(handle);
      find.bindGrid(handle);
    },
    [registerTimelineGridHandle, find.bindGrid],
  );
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
      operationFeedback:
        !loadAccessLost && currentIncidentRole && operationError ? (
          <div
            role="alert"
            aria-label="Timeline operation feedback"
            style={{ padding: "var(--ct-spacing-xs) var(--ct-spacing-sm)" }}
          >
            {operationError.message}
          </div>
        ) : null,
      parkedDrafts:
        !loadAccessLost && currentIncidentRole ? (
          <TimelineParkedGridDrafts
            registry={editorDraftRegistry}
            canEdit={interactionMode.kind === "editable" && !incidentClosed}
            rows={rows}
            fieldKeys={visibleTimelineColumns.map((column) => column.fieldKey)}
          />
        ) : null,
      fileOwner: !loadAccessLost && currentIncidentRole ? fileOwner : undefined,
      onFilesSelected: handleTimelineEvidenceFiles,
      onFileAdmission: fileOwner.reportAdmission,
      fileEditorRegistry: editorDraftRegistry,
      activeRecordId: selectedRowId,
      bulkSelection: timelineBulkSelection,
      clipboardPaste: timelineClipboardPaste,
      cellRangeScopeKey: mutation.snapshot.cellRangeScopeKey,
      columns: visibleTimelineColumns,
      frozenDataColumnPrefix: workbookFrozenDataColumnPrefix(layoutState),
      dataState: timelineDataState,
      density,
      getCellState: getFindCellState,
      getGroupLabel: getTimelineGroupLabel,
      getGroupRowTestId: getTimelineGroupRowTestId,
      getRowState: getTimelineRowState,
      groupBy: queryState.groupBy,
      interactionMode,
      onActiveCellChange: handleActiveCellChange,
      onColumnReorder: handleColumnReorder,
      onColumnSizingIntent: handleColumnSizingIntent,
      onFillCells: handleFillCells,
      onClearCells: handleClearCells,
      onSelectRecord: handleSelectRow,
      onSortChange: handleQuerySortChange,
      ref: registerFindGrid,
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
      onWorkAreaPointerDown: handleTimelineGridPointerDown,
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
              ...rowContextMenu,
              reviewDisabledReason: captureActions.reason(
                rowContextMenu.row,
                "mark-reviewed",
              ),
              supersedeDisabledReason: captureActions.reason(
                rowContextMenu.row,
                "supersede",
              ),
              onInspectRow: (recordId: string) => {
                openInspectorForRow(recordId);
                requestRowPanelFocus(recordId, "details");
              },
              onMarkReviewed: (rowKey: string) => {
                captureActions.start(rowKey, "mark-reviewed");
              },
              onOpenHistory: (recordId: string) => {
                openRowHistory(recordId);
                requestRowPanelFocus(recordId, "history");
              },
              onSupersede: (rowKey: string) => {
                captureActions.start(rowKey, "supersede");
              },
            },
    },
    notices: {
      owner: workflow.commands.mentions.owner,
      entityIndex,
      density,
      onReviewAutoResolution: async (
        notice: import("../actions/WorkbookTimelineMentionOperationOwner").AutoResolutionDisclosure,
      ) => {
        await workflow.commands.mentions.prepareDisclosureReview(notice);
        handleSelectMention(notice.rowRecordId, notice.itemRef);
      },
      onUndoAutoResolution: handleUndoAutoResolutionNotice,
    },
    status: {
      presence,
      onActivateConflict,
      source: mutation.statusSource,
      sheetRef: mutation.sheetRef,
      chromeMode,
      showPresence: showStatusPresence,
      workbookFocusAnchor,
    },
    bulkTag: interaction.ports.bulkTag,
    viewBar: {
      find: find.control,
      onClearContents: (event: MouseEvent<HTMLButtonElement>) =>
        handleClearCells(
          grid.refs.gridHandle.current?.captureClearIntent?.(
            event.nativeEvent,
          ) ?? null,
        ),
      addRowDisabled: interactionMode.kind === "read_only",
      chromeMode,
      workingSet: renderInlineQueryControls
        ? {
            savedView: viewBarWorkingSet?.savedView ?? null,
            query: {
              contract: timelineContract,
              defaultFilterPopoverOpen: true,
              filterDraft,
              layoutState,
              sizing: layout.commands.sizing,
              freezing: layout.commands.freezing,
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
              requestedSort: requestedQueryState.sort,
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
