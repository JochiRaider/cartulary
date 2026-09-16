import {
  type GridActionsColumn,
  type GridCellPasteIntent,
  type GridColumn,
  type GridDataRow,
  type GridDraftRow,
  type GridEditCommitOutcome,
  type GridGroupingDescriptor,
  type GridHandle,
  GridViewport,
  SemanticDataGrid,
} from "@cartulary/grid-adapter";
import {
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridActionsHeaderTestId,
  gridDataCellsSelector,
  gridGroupRowTestId,
  gridSavedRowsSelector,
  gridShellTestId,
  workbookInlineDraftRowTestId,
} from "@cartulary/ui-contracts";
import type { ViewContract } from "@cartulary/view-contracts";
import {
  type CSSProperties,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import { useWorkbookCollaborationCoordinator } from "../collaboration/useWorkbookCollaborationCoordinator";
import type { WorkbookCollaborationCoordinator } from "../collaboration/WorkbookCollaborationCoordinator";
import { presenceForRow } from "../collaboration/workbookPresencePresentation";
import {
  useWorkbookGridContinuity,
  WorkbookContinuityCell,
} from "../continuity/useWorkbookGridContinuity";
import type {
  WorkbookContinuityPort,
  WorkbookContinuityToken,
} from "../continuity/workbookContinuityPort";
import { TaskPatchRecovery } from "../features/coordination/TaskPatchRecovery";
import {
  taskGuardFields,
  taskPatchErrors,
  taskViewId,
} from "../features/coordination/taskLifecycleModel";
import { useGenericCreateDraft } from "../features/generic/useGenericCreateDraft";
import { useGenericWorkbookInspectorComposition } from "../features/generic/useGenericWorkbookInspectorComposition";
import { OrdinaryCreateControl } from "../features/ordinary/OrdinaryCreateControl";
import { OrdinaryCreateNotice } from "../features/ordinary/OrdinaryCreateNotice";
import { useGenericSurfaceMutationController } from "../hooks/useGenericSurfaceMutationController";
import { useWorkbookSemanticGridFocus } from "../hooks/useWorkbookSemanticGridFocus";
import { WorkbookExplicitPatchRecovery } from "../inspector/WorkbookExplicitPatchRecovery";
import type { WorkbookSurfaceLayoutOwner } from "../layout/useWorkbookLayoutFacade";
import {
  WorkbookSurfaceLayout,
  workbookGridWithNoticeStyle,
  workbookSurfaceGridShellStyle,
} from "../layout/WorkbookSurfaceLayout";
import { applyWorkbookLayoutToColumns } from "../layout/workbookColumnLayout";
import {
  genericCellLabelForField,
  genericContractColumnWidth,
  genericRowLabel,
  workbookCreationAvailable,
} from "../models/genericWorkbookModel";
import {
  workbookContractColumns,
  workbookGridRows,
} from "../models/workbookContractRows";
import { workbookGridEditChange } from "../models/workbookGridEditValue";
import type { WorkbookGridEntryFocusOwner } from "../models/workbookGridEntryFocus";
import {
  type WorkbookQueryLoadState,
  workbookGridDataState,
} from "../models/workbookGridState";
import type { WorkbookQueryState } from "../models/workbookQuery";
import {
  compareWorkbookGroupValues,
  workbookGroupValue,
} from "../models/workbookQuery";
import { requireWorkbookSurfaceRegistration } from "../models/workbookSurfaceRegistration";
import type { WorkbookMutationCommandPorts } from "../mutations/workbookMutationCommandPorts";
import type { WorkbookIncidentPort } from "../ports/WorkbookIncidentPort";
import { useWorkbookQueryRestart } from "../query/WorkbookQueryBrowsingContext";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import { useWorkbookMutationRuntime } from "../runtime/useWorkbookMutationRuntime";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import { workbookClipboardPasteContract } from "../utils/workbookClipboard";
import { workbookGridEditorAdapter } from "./WorkbookGridEditorControl";
import { WorkbookUnavailableGridDrafts } from "./WorkbookParkedGridDrafts";
import {
  WorkbookCellPresenceMarker,
  WorkbookPresenceCellLayout,
  WorkbookRowGutterContent,
  workbookPresenceRowGutter,
} from "./WorkbookPresenceMarkers";
import {
  type WorkbookConflictActivation,
  WorkbookStatusStrip,
} from "./WorkbookStatusStrip";
import {
  WorkbookViewBar,
  type WorkbookViewBarWorkingSetBinding,
} from "./WorkbookViewBar";

export type ContractWorkbookSurfaceProps = {
  readonly contract: ViewContract;
  readonly continuityResetKey: string;
  readonly currentUserId: string | null;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly incidentPort: WorkbookIncidentPort;
  readonly inspectorResetKey: string;
  readonly gridEntryFocus: WorkbookGridEntryFocusOwner;
  readonly viewBarWorkingSet: WorkbookViewBarWorkingSetBinding;
  readonly layout: WorkbookSurfaceLayoutOwner;
  readonly loadState: WorkbookQueryLoadState;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly mutationCommands: WorkbookMutationCommandPorts;
  readonly onActivateConflict?: WorkbookConflictActivation | undefined;
  readonly collaborationProjection: WorkbookCollaborationCoordinator;
  readonly sheetRef: SheetRef;
  readonly onClearFilters: () => void;
  readonly onAuthorityUncertain?: (() => void) | undefined;
  readonly onRefresh: (options?: {
    readonly requireAcceptance?: boolean;
  }) => Promise<void> | void;
  readonly onSortChange: (sort: WorkbookQueryState["sort"]) => void;
  readonly queryState: WorkbookQueryState;
  readonly rows: WorkbookQueryRow[];
};

export function ContractWorkbookSurface({
  contract,
  continuityResetKey,
  currentIncidentRole,
  currentUserId,
  inspectorResetKey,
  gridEntryFocus,
  viewBarWorkingSet,
  layout,
  loadState,
  mutationRuntime,
  mutationCommands,
  onActivateConflict,
  collaborationProjection,
  sheetRef,
  onClearFilters,
  onRefresh,
  onSortChange,
  queryState,
  rows,
}: ContractWorkbookSurfaceProps) {
  const {
    commands: { onColumnReorder, onColumnWidthChange },
    snapshot: {
      chromeMode,
      density,
      incidentClosed,
      interactionMode,
      showStatusPresence,
      state: layoutState,
    },
  } = layout;
  const surface = contract.viewSchemaId;
  const registration = requireWorkbookSurfaceRegistration(
    contract.viewSchemaId,
  );
  const { ownerBindings } = registration.policy;
  const draftRowRecordId = `${surface}:draft-row`;
  const inspectorContinuityTokenRef = useRef<WorkbookContinuityToken | null>(
    null,
  );
  const continuityPortRef = useRef<WorkbookContinuityPort | null>(null);
  const createFields = useMemo(
    () => contract.fields.filter((field) => field.createWritable),
    [contract],
  );
  const canCreateRows =
    interactionMode.kind === "editable" && workbookCreationAvailable(contract);
  const showOrdinaryDraft =
    canCreateRows ||
    Object.keys(
      mutationRuntime.ordinaryCreate.getSnapshot().schemas[
        contract.viewSchemaId
      ]?.draft.values ?? {},
    ).length > 0;
  const [createDraft, setCreateDraft, draftDisabled] = useGenericCreateDraft(
    contract,
    currentUserId,
    mutationRuntime.ordinaryCreate,
  );
  const [editRecordId, setEditRecordId] = useState("");
  const mutationController = useGenericSurfaceMutationController({
    mutationRuntime,
    selectedRecordId: editRecordId,
    surfaceLabel: contract.title,
    sheetRef,
  });
  const { setValidationError } = mutationController;
  const sharedMutation = useWorkbookMutationRuntime(mutationRuntime, sheetRef);
  const collaboration = useWorkbookCollaborationCoordinator(
    collaborationProjection,
  );

  const anchorColumns = useMemo<readonly GridColumn<WorkbookQueryRow>[]>(
    () =>
      workbookContractColumns<WorkbookQueryRow>({
        contract,
        surface,
        widthForField: genericContractColumnWidth,
      }),
    [contract, surface],
  );
  const commitGridEdit = useCallback(
    async (
      fieldKey: string,
      draftValue: string | null,
      target: {
        readonly baseRowVersion: number;
        readonly recordId: string;
      },
    ): Promise<GridEditCommitOutcome> => {
      const field = contract.fieldMap[fieldKey];
      if (field?.gridEditable !== true) {
        return {
          kind: "rejected_mutation",
          message: "This field is not grid-editable.",
        };
      }
      const current = rows.find((row) => row.record_id === target.recordId);
      if (current === undefined) {
        return {
          kind: "stale_target",
          message: "The record changed before this edit was submitted.",
        };
      }
      const identity = {
        viewSchemaId: contract.viewSchemaId,
        recordId: target.recordId,
        fieldKey,
      };
      mutationRuntime.gridDrafts.update(identity, current, draftValue);
      if (!mutationRuntime.gridDrafts.canAuthor())
        return {
          kind: "rejected_mutation",
          message: "Editing is unavailable with the current authorization.",
        };
      const dependencies =
        contract.viewSchemaId === taskViewId &&
        taskGuardFields.some((key) => key === fieldKey)
          ? taskGuardFields
          : [];
      if (
        !mutationRuntime.hasPendingGridWrite(target.recordId) &&
        mutationRuntime.gridDrafts.staleFields(identity, current, dependencies)
          .length
      )
        return {
          kind: "stale_target",
          message:
            "Review the changed saved value before committing this draft.",
        };
      const change = workbookGridEditChange(field, draftValue);
      if (change === null) {
        return {
          kind: "validation_error",
          message:
            "Enter a valid value, or clear only a field that permits null.",
        };
      }
      if (contract.viewSchemaId === taskViewId) {
        const error = taskPatchErrors(current, [change])[0];
        if (error) return { kind: "validation_error", message: error.message };
      }
      const admission = mutationRuntime.enqueuePatch({
        sheetRef,
        baseline: current,
        dependencies,
        baseRowVersion: target.baseRowVersion,
        changes: [change],
        fieldKey,
        focusKey: `${target.recordId}:${fieldKey}`,
        localValue: draftValue,
        recordId: target.recordId,
        rowLabel: genericRowLabel(contract, current),
        surfaceLabel: contract.title,
        viewSchemaId: contract.viewSchemaId,
      });
      return admission.kind === "admitted" ? admission.completion : admission;
    },
    [contract, mutationRuntime, rows, sheetRef],
  );
  const visibleAnchorColumns = useMemo(
    () => applyWorkbookLayoutToColumns(contract, anchorColumns, layoutState),
    [anchorColumns, contract, layoutState],
  );
  const handleGridPaste = useCallback(
    async (intent: GridCellPasteIntent) => {
      const values =
        intent.input.kind === "scalar"
          ? [[intent.input.value]]
          : intent.input.values;
      const resolution = intent.targetResolution;
      const rowTarget = resolution?.rowTargets[0];
      if (
        values.length !== 1 ||
        values[0]?.length !== 1 ||
        resolution?.columns.length !== 1 ||
        resolution.rowTargets.length !== 1 ||
        rowTarget?.kind !== "record"
      ) {
        setValidationError(
          "This surface accepts one pasted grid-editable value at a time.",
        );
        return false;
      }
      const outcome = await commitGridEdit(
        resolution.columns[0] ?? intent.target.fieldKey,
        values[0]?.[0] ?? "",
        {
          baseRowVersion: rowTarget.mutationIdentity.baseRowVersion,
          recordId: rowTarget.rowIdentity.recordId,
        },
      );
      if (outcome.kind !== "accepted") setValidationError(outcome.message);
      return outcome.kind === "accepted";
    },
    [commitGridEdit, setValidationError],
  );
  const clipboardPaste = useMemo(
    () => workbookClipboardPasteContract(handleGridPaste, setValidationError),
    [handleGridPaste, setValidationError],
  );
  const draftInspectorFields = useMemo(() => {
    const gridFieldKeys = new Set(
      visibleAnchorColumns.map((column) => column.fieldKey),
    );
    return createFields.filter((field) => !gridFieldKeys.has(field.fieldKey));
  }, [createFields, visibleAnchorColumns]);
  const genericInspector = useGenericWorkbookInspectorComposition({
    sheetRef,
    canCreateRows,
    contract,
    createDraft,
    draftDisabled,
    currentIncidentRole,
    currentUserId,
    draftInspectorFields,
    density,
    incidentClosed,
    inspectorResetKey,
    interactionMode,
    mutation: mutationController,
    mutationCommands,
    onClearSurfaceSelection: () => {
      continuityPortRef.current?.clear();
      setEditRecordId("");
    },
    onRefresh,
    onRestoreFocus: () => {
      const token = inspectorContinuityTokenRef.current;
      inspectorContinuityTokenRef.current = null;
      if (token !== null) void continuityPortRef.current?.restore(token);
    },
    onRestoreEvidenceFocus: (recordId) => {
      const fieldKey = visibleAnchorColumns[0]?.fieldKey;
      if (fieldKey !== undefined)
        void continuityPortRef.current?.focus({
          viewSchemaId: contract.viewSchemaId,
          recordId,
          fieldKey,
        });
    },
    onSelectRecord: setEditRecordId,
    ownerBindings,
    rows,
    selectedRecordId: editRecordId,
    setCreateDraft,
  });
  const draftApiRow = useMemo<WorkbookQueryRow>(
    () => ({
      record_id: draftRowRecordId,
      row_version: 0,
      cells: Object.fromEntries(
        contract.fields.map((field) => [
          field.fieldKey,
          { value: createDraft[field.fieldKey] ?? "" },
        ]),
      ),
    }),
    [contract.fields, createDraft, draftRowRecordId],
  );
  const gridRecordRows = useMemo<readonly GridDataRow<WorkbookQueryRow>[]>(
    () =>
      workbookGridRows({
        renderGutter: (recordId, ordinal) => (
          <WorkbookRowGutterContent
            recordId={recordId}
            ordinal={ordinal}
            presences={presenceForRow(collaboration.presence, recordId)}
          />
        ),
        getRecordId: (row: WorkbookQueryRow) => row.record_id,
        getRowVersion: (row: WorkbookQueryRow) => row.row_version,
        rows,
        surface,
      }),
    [rows, surface, collaboration.presence],
  );
  const gridDraftRow = useMemo<GridDraftRow<WorkbookQueryRow> | undefined>(
    () =>
      !showOrdinaryDraft
        ? undefined
        : {
            kind: "draft",
            data: draftApiRow,
            gutterContent: "+",
            gutterLabel: "Draft row",
            testId: workbookInlineDraftRowTestId(surface),
          },
    [showOrdinaryDraft, draftApiRow, surface],
  );
  const grouping =
    useMemo<GridGroupingDescriptor<WorkbookQueryRow> | null>(() => {
      const fieldKey = queryState.groupBy;
      if (fieldKey === null) {
        return null;
      }
      return {
        fieldKey,
        compareValues: compareWorkbookGroupValues,
        formatLabel: (value) =>
          genericCellLabelForField(surface, fieldKey, value),
        getTestId: (groupFieldKey, _value, label) =>
          label === null
            ? undefined
            : gridGroupRowTestId(surface, groupFieldKey, label),
        getValue: (row) => workbookGroupValue(row, fieldKey),
        label: contract.fieldMap[fieldKey]?.label ?? fieldKey,
      };
    }, [contract.fieldMap, queryState.groupBy, surface]);
  const gridHandleRef = useRef<GridHandle | null>(null);
  const genericFocus = useWorkbookGridContinuity({
    columns: visibleAnchorColumns,
    continuityResetKey,
    gridHandleRef,
    viewSchemaId: surface,
  });
  continuityPortRef.current = genericFocus.port;
  useEffect(
    () =>
      mutationRuntime.registerSurface(
        contract.viewSchemaId,
        async () => {
          await onRefresh({ requireAcceptance: true });
        },
        async () => {
          await onRefresh({ requireAcceptance: true });
        },
      ),
    [contract.viewSchemaId, mutationRuntime, onRefresh],
  );
  const columns: readonly GridColumn<WorkbookQueryRow>[] =
    visibleAnchorColumns.map((column) => {
      const field = contract.fieldMap[column.fieldKey];
      const displayedValue = (row: WorkbookQueryRow) => {
        const local = mutationRuntime.visibleEdit(
          contract.viewSchemaId,
          row.record_id,
          column.fieldKey,
        );
        return local === undefined ? row.cells[column.fieldKey]?.value : local;
      };
      return {
        ...column,
        contractWritable: field?.gridEditable === true,
        draftWritable: field?.createWritable === true,
        getClipboardValue: (row: WorkbookQueryRow) => {
          const value = displayedValue(row);
          return field?.readKind === "collection"
            ? genericCellLabelForField(surface, column.fieldKey, value)
            : value;
        },
        editor:
          field?.gridEditable === true
            ? workbookGridEditorAdapter({
                drafts: mutationRuntime.gridDrafts,
                viewSchemaId: contract.viewSchemaId,
                readRow: (row: WorkbookQueryRow) => row,
                readCurrentRow: (row: WorkbookQueryRow) =>
                  mutationRuntime.explicitPatches.latestRow(row.record_id) ??
                  row,
                collaboration: collaborationProjection,
                commit: (draftValue, target) =>
                  commitGridEdit(field.fieldKey, draftValue, {
                    baseRowVersion: target.mutationIdentity.baseRowVersion,
                    recordId:
                      target.rowIdentity.kind === "core_record"
                        ? target.rowIdentity.recordId
                        : "",
                  }),
                field,
                readValue: (row: WorkbookQueryRow) => displayedValue(row),
              })
            : undefined,
        renderDraftCell: ({ focusTargetRef }) => {
          const writableField =
            createFields.find(
              (candidate) => candidate.fieldKey === column.fieldKey,
            ) ?? null;
          if (writableField === null) {
            return <span style={draftCellPlaceholderStyle}>-</span>;
          }
          return (
            <OrdinaryCreateControl
              owner={mutationRuntime.ordinaryCreate}
              contract={contract}
              disabled={draftDisabled}
              collectionMode="add"
              field={writableField}
              focusTargetRef={focusTargetRef}
              surface="grid"
              testId={genericCreateFieldTestId(writableField.fieldKey)}
              value={createDraft[writableField.fieldKey] ?? ""}
              onChange={(value) => {
                setCreateDraft((current) => ({
                  ...current,
                  [writableField.fieldKey]: value,
                }));
              }}
            />
          );
        },
        renderCell: ({ row }) => {
          return (
            <WorkbookContinuityCell
              continuity={genericFocus.port}
              fieldKey={column.fieldKey}
              recordId={row.record_id}
              viewSchemaId={contract.viewSchemaId}
            >
              <WorkbookPresenceCellLayout
                marker={
                  <WorkbookCellPresenceMarker
                    fieldKey={column.fieldKey}
                    fieldLabel={field?.label ?? column.fieldKey}
                    presences={collaborationProjection.editingPresenceForCell(
                      row.record_id,
                      column.fieldKey,
                    )}
                    recordId={row.record_id}
                  />
                }
              >
                {genericCellLabelForField(
                  surface,
                  column.fieldKey,
                  displayedValue(row),
                )}
              </WorkbookPresenceCellLayout>
            </WorkbookContinuityCell>
          );
        },
      };
    });
  const rowActionsColumn = useMemo<
    GridActionsColumn<WorkbookQueryRow> | undefined
  >(() => {
    if (
      !genericInspector.ownerRecordActions.hasRecordActions &&
      !canCreateRows
    ) {
      return undefined;
    }
    return {
      headerTestId: gridActionsHeaderTestId(surface),
      label: "",
      width: genericInspector.ownerRecordActions.actionsWidth,
      renderDraftCell: () => (
        <button
          data-testid={
            genericInspector.isOpen
              ? undefined
              : genericCreateSubmitTestId(contract.viewSchemaId)
          }
          disabled={
            draftDisabled ||
            mutationRuntime.ordinaryCreate.busy(contract.viewSchemaId)
          }
          style={secondaryActionButtonStyle}
          type="button"
          onClick={() => {
            void genericInspector.submitCreate();
          }}
        >
          Commit
        </button>
      ),
      renderCell: ({ data: row }) => {
        return genericInspector.ownerRecordActions.renderRowActions(row);
      },
    };
  }, [
    contract.viewSchemaId,
    genericInspector,
    mutationRuntime,
    draftDisabled,
    surface,
    canCreateRows,
  ]);

  const focusDraftRow = useCallback(() => {
    const firstWritableField = createFields[0];
    if (!firstWritableField || !canCreateRows) {
      return;
    }
    void gridHandleRef.current?.requestFocus({
      kind: "draft",
      fieldKey: firstWritableField.fieldKey,
    });
  }, [canCreateRows, createFields]);
  const restartQuery = useWorkbookQueryRestart(
    contract.viewSchemaId,
    onRefresh,
  );
  const dataState = workbookGridDataState({
    emptyAction: canCreateRows
      ? { label: "Add row", onInvoke: focusDraftRow }
      : undefined,
    emptyMessage: `No ${contract.title.toLocaleLowerCase()} records are available.`,
    loadState,
    onClearFilters,
    onRetry: () => void restartQuery(),
    queryState,
    rowCount: gridRecordRows.length,
    surfaceLabel: contract.title,
  });
  const gridEntryDraftFieldKeys = useMemo(
    () =>
      gridDraftRow === undefined
        ? []
        : visibleAnchorColumns
            .filter((column) =>
              createFields.some((field) => field.fieldKey === column.fieldKey),
            )
            .map((column) => column.fieldKey),
    [createFields, gridDraftRow, visibleAnchorColumns],
  );
  const registerGridHandle = useWorkbookSemanticGridFocus({
    dataRows: gridRecordRows,
    dataState,
    draftFieldKeys: gridEntryDraftFieldKeys,
    focusOwner: gridEntryFocus,
    gridHandleRef,
    visibleColumns: columns,
    viewSchemaId: surface,
  });
  return (
    <WorkbookSurfaceLayout
      chromeMode={chromeMode}
      inspector={genericInspector.node}
      onRequestPreviewClose={genericInspector.ownerRecordActions.closePreview}
      onRequestInspectorClose={() => {
        genericInspector.close();
      }}
      primaryGrid={
        <div
          onFocusCapture={(event) => {
            // A retained grid caret can revisit the same Task after filter
            // departure. Observe real focus through adapter-owned semantic
            // identity even when its active-cell notification deduplicates.
            if (
              contract.viewSchemaId !== taskViewId ||
              !(event.target instanceof HTMLElement)
            )
              return;
            const cell = event.target.closest<HTMLElement>(
              gridDataCellsSelector(),
            );
            const recordId = cell?.closest<HTMLElement>(gridSavedRowsSelector())
              ?.dataset.gridRecordId;
            const fieldKey = cell?.querySelector<HTMLElement>(
              "[data-grid-field-key]",
            )?.dataset.gridFieldKey;
            if (
              !cell ||
              !event.currentTarget.contains(cell) ||
              !recordId ||
              !fieldKey ||
              !rows.some((row) => row.record_id === recordId)
            )
              return;
            setEditRecordId(recordId);
            genericFocus.port.select({
              recordId,
              fieldKey,
              viewSchemaId: contract.viewSchemaId,
            });
          }}
          style={workbookGridWithNoticeStyle}
        >
          <div>
            <WorkbookUnavailableGridDrafts
              store={mutationRuntime.gridDrafts}
              contract={contract}
              recordIds={rows.map((row) => row.record_id)}
              fieldKeys={columns.map((column) => column.fieldKey)}
            />
            <WorkbookExplicitPatchRecovery
              owner={mutationRuntime.explicitPatches}
              viewSchemaId={contract.viewSchemaId}
            />
            <OrdinaryCreateNotice
              owner={mutationRuntime.ordinaryCreate}
              view={contract.viewSchemaId}
            />
            {contract.viewSchemaId === taskViewId ? (
              <TaskPatchRecovery
                owner={mutationRuntime.explicitPatches}
                rows={rows}
              />
            ) : null}
          </div>
          <GridViewport
            blockSizing="fill"
            style={gridShellStyle}
            testId={gridShellTestId(surface)}
          >
            <SemanticDataGrid
              rowGutter={workbookPresenceRowGutter}
              ref={registerGridHandle}
              actionsColumn={rowActionsColumn}
              columns={columns}
              columnWidths={layoutState.columnWidths}
              dataState={dataState}
              density={density}
              fillViewportInline={
                genericInspector.ownerRecordActions.hasRecordActions
              }
              draftRow={gridDraftRow}
              grouping={grouping}
              interactionMode={interactionMode}
              onActiveCellChange={(anchor) => {
                const recordId =
                  anchor?.rowIdentity.kind === "core_record"
                    ? anchor.rowIdentity.recordId
                    : null;
                setEditRecordId(recordId ?? "");
                if (recordId === null || anchor === null) {
                  genericFocus.port.clear();
                } else {
                  genericFocus.port.select({
                    fieldKey: anchor.fieldKey,
                    recordId,
                    viewSchemaId: contract.viewSchemaId,
                  });
                }
                collaborationProjection.publishFocusedCell(
                  recordId,
                  anchor?.fieldKey ?? null,
                );
              }}
              onColumnReorder={onColumnReorder}
              onColumnWidthChange={onColumnWidthChange}
              clipboardPaste={clipboardPaste}
              onSortChange={onSortChange}
              dataRows={gridRecordRows}
              sort={queryState.sort}
              surface={{ kind: "view_schema", viewSchemaId: surface }}
            />
          </GridViewport>
        </div>
      }
      statusStrip={
        <WorkbookStatusStrip
          presence={collaboration.presence.header}
          status={sharedMutation}
          chromeMode={chromeMode}
          onActivateConflict={onActivateConflict}
          showPresence={showStatusPresence}
          workbookFocusAnchor={genericFocus.snapshot.anchor}
        />
      }
      viewBar={
        <WorkbookViewBar
          addRowDisabled={!canCreateRows}
          chromeMode={chromeMode}
          workingSet={viewBarWorkingSet}
          onAddRow={focusDraftRow}
          onInspectorToggle={() => {
            inspectorContinuityTokenRef.current = genericFocus.port.capture();
            genericInspector.open();
          }}
          surface={surface}
        />
      }
      viewSchemaId={surface}
      workAreaOverlays={genericInspector.ownerRecordActions.overlay}
      workAreaAnnouncements={genericInspector.ownerRecordActions.announcements}
    />
  );
}

const gridShellStyle = {
  ...workbookSurfaceGridShellStyle,
} satisfies CSSProperties;

const actionButtonStyle = {
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  padding: "0.55rem 0.9rem",
  font: "inherit",
  cursor: "pointer",
};

const secondaryActionButtonStyle = {
  ...actionButtonStyle,
  background: "var(--ct-colors-surface-3)",
};

const draftCellPlaceholderStyle = {
  color: "var(--ct-colors-ink-subtle)",
};
