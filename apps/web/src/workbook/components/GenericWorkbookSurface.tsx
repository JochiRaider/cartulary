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
  gridGroupRowTestId,
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
import {
  useWorkbookGridContinuity,
  WorkbookContinuityCell,
} from "../continuity/useWorkbookGridContinuity";
import type {
  WorkbookContinuityPort,
  WorkbookContinuityToken,
} from "../continuity/workbookContinuityPort";
import { useGenericWorkbookInspectorComposition } from "../features/generic/useGenericWorkbookInspectorComposition";
import { useGenericSurfaceMutationController } from "../hooks/useGenericSurfaceMutationController";
import { useOwnerReferenceOptions } from "../hooks/useOwnerReferenceOptions";
import { useWorkbookSemanticGridFocus } from "../hooks/useWorkbookSemanticGridFocus";
import type { WorkbookSurfaceLayoutOwner } from "../layout/useWorkbookLayoutFacade";
import {
  WorkbookSurfaceLayout,
  workbookSurfaceGridShellStyle,
} from "../layout/WorkbookSurfaceLayout";
import { applyWorkbookLayoutToColumns } from "../layout/workbookColumnLayout";
import {
  buildGenericPatchChange,
  genericCellLabelForField,
  genericContractColumnWidth,
  genericRowLabel,
  initialGenericCreateDraft,
  workbookCreationAvailable,
} from "../models/genericWorkbookModel";
import {
  workbookContractColumns,
  workbookGridRows,
} from "../models/workbookContractRows";
import type { WorkbookGridEntryFocusOwner } from "../models/workbookGridEntryFocus";
import {
  type WorkbookQueryLoadState,
  workbookGridDataState,
} from "../models/workbookGridState";
import type { WorkbookQueryState } from "../models/workbookQuery";
import { requireWorkbookSurfaceRegistration } from "../models/workbookSurfaceRegistration";
import type { WorkbookMutationCommandPorts } from "../mutations/workbookMutationCommandPorts";
import type { WorkbookIncidentPort } from "../ports/WorkbookIncidentPort";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import { useWorkbookMutationRuntime } from "../runtime/useWorkbookMutationRuntime";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import type { ReferenceQueryBrokerPort } from "../services/referenceQueryBroker";
import { workbookClipboardPasteContract } from "../utils/workbookClipboard";
import { GenericMutationControl } from "./GenericMutationControl";
import { workbookGridEditorAdapter } from "./WorkbookGridEditorControl";
import { WorkbookCellPresenceMarker } from "./WorkbookPresenceMarkers";
import {
  type WorkbookConflictActivation,
  WorkbookSurfaceStatusStrip,
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
  readonly referenceQueryBroker: ReferenceQueryBrokerPort;
  readonly collaborationProjection: WorkbookCollaborationCoordinator;
  readonly sheetRef: SheetRef;
  readonly onClearFilters: () => void;
  readonly onIncidentAccessLost?: (() => void) | undefined;
  readonly onRefresh: () => Promise<void> | void;
  readonly onSortChange: (sort: WorkbookQueryState["sort"]) => void;
  readonly queryState: WorkbookQueryState;
  readonly rows: WorkbookQueryRow[];
};

export function ContractWorkbookSurface({
  contract,
  continuityResetKey,
  currentIncidentRole,
  currentUserId,
  incidentPort,
  inspectorResetKey,
  gridEntryFocus,
  viewBarWorkingSet,
  layout,
  loadState,
  mutationRuntime,
  mutationCommands,
  onActivateConflict,
  referenceQueryBroker,
  collaborationProjection,
  sheetRef: _sheetRef,
  onClearFilters,
  onIncidentAccessLost,
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
  const [createDraft, setCreateDraft] = useState<Record<string, string>>(() =>
    initialGenericCreateDraft(contract, currentUserId),
  );
  const [editRecordId, setEditRecordId] = useState("");
  const { referenceLoadError, referenceOptions, refreshReferenceOptions } =
    useOwnerReferenceOptions({
      incidentPort,
      onIncidentAccessLost,
      referenceQueryBroker,
      viewSchemaId: contract.viewSchemaId,
    });
  const mutationController = useGenericSurfaceMutationController({
    mutationCommands: mutationCommands.generic,
    mutationRuntime,
    onRefresh,
    refreshReferenceOptions,
    surfaceLabel: contract.title,
  });
  const { mutationError, mutationState, setValidationError } =
    mutationController;
  const sharedMutation = useWorkbookMutationRuntime(
    mutationRuntime,
    contract.viewSchemaId,
  );
  const collaboration = useWorkbookCollaborationCoordinator(
    collaborationProjection,
  );
  const presentedMutationState =
    mutationState === "Saved" ? sharedMutation.primaryLabel : mutationState;

  useEffect(() => {
    setCreateDraft((current) => {
      const defaults = initialGenericCreateDraft(contract, currentUserId);
      return { ...defaults, ...current };
    });
  }, [contract, currentUserId]);

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
      draftValue: string,
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
      if (
        current === undefined ||
        current.row_version !== target.baseRowVersion
      ) {
        return {
          kind: "stale_target",
          message: "The record changed before this edit was submitted.",
        };
      }
      const change = buildGenericPatchChange(
        field,
        draftValue,
        "add",
        contract.viewSchemaId,
      );
      if (change === null) {
        return {
          kind: "validation_error",
          message:
            "Enter a valid value, or clear only a field that permits null.",
        };
      }
      return mutationRuntime.enqueuePatch({
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
    },
    [contract, mutationRuntime, rows],
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
        return;
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
    },
    [commitGridEdit, setValidationError],
  );
  const clipboardPaste = useMemo(
    () =>
      workbookClipboardPasteContract((intent) => {
        void handleGridPaste(intent);
      }),
    [handleGridPaste],
  );
  const draftInspectorFields = useMemo(() => {
    const gridFieldKeys = new Set(
      visibleAnchorColumns.map((column) => column.fieldKey),
    );
    return createFields.filter((field) => !gridFieldKeys.has(field.fieldKey));
  }, [createFields, visibleAnchorColumns]);
  const genericInspector = useGenericWorkbookInspectorComposition({
    canCreateRows,
    contract,
    createDraft,
    currentIncidentRole,
    currentUserId,
    draftInspectorFields,
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
      if (token !== null) continuityPortRef.current?.restore(token);
    },
    onSelectRecord: setEditRecordId,
    ownerBindings,
    referenceLoadError,
    referenceOptions,
    refreshReferenceOptions,
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
        getRecordId: (row: WorkbookQueryRow) => row.record_id,
        getRowVersion: (row: WorkbookQueryRow) => row.row_version,
        rows,
        surface,
      }),
    [rows, surface],
  );
  const gridDraftRow = useMemo<GridDraftRow<WorkbookQueryRow> | undefined>(
    () =>
      !canCreateRows
        ? undefined
        : {
            kind: "draft",
            data: draftApiRow,
            gutterContent: "+",
            gutterLabel: "Draft row",
            testId: workbookInlineDraftRowTestId(surface),
          },
    [canCreateRows, draftApiRow, surface],
  );
  const grouping =
    useMemo<GridGroupingDescriptor<WorkbookQueryRow> | null>(() => {
      const fieldKey = queryState.groupBy;
      if (fieldKey === null) {
        return null;
      }
      return {
        fieldKey,
        formatLabel: (value) =>
          genericCellLabelForField(surface, fieldKey, value),
        getTestId: (groupFieldKey, _value, label) =>
          label === null
            ? undefined
            : gridGroupRowTestId(surface, groupFieldKey, label),
        getValue: (row) => {
          const value = row.cells[fieldKey]?.value;
          return value === null ||
            typeof value === "boolean" ||
            typeof value === "number" ||
            typeof value === "string"
            ? value
            : null;
        },
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
        onRefresh,
        async (_payload, conflict) => {
          await onRefresh();
          window.setTimeout(() => {
            genericFocus.port.focus({
              fieldKey: conflict.conflict.field_key,
              recordId: conflict.conflict.record_id,
              viewSchemaId: contract.viewSchemaId,
            });
          }, 0);
        },
        (conflict) => {
          window.setTimeout(() => {
            const anchor = {
              fieldKey: conflict.conflict.field_key,
              rowIdentity: {
                kind: "core_record" as const,
                recordId: conflict.conflict.record_id,
              },
              surface: {
                kind: "view_schema" as const,
                viewSchemaId: contract.viewSchemaId,
              },
            };
            if (
              !gridHandleRef.current?.activateEdit(anchor, {
                value: conflict.localValue,
              })
            ) {
              genericFocus.port.focus({
                fieldKey: anchor.fieldKey,
                recordId: conflict.conflict.record_id,
                viewSchemaId: contract.viewSchemaId,
              });
            }
          }, 0);
        },
      ),
    [contract.viewSchemaId, genericFocus.port, mutationRuntime, onRefresh],
  );
  const columns: readonly GridColumn<WorkbookQueryRow>[] =
    visibleAnchorColumns.map((column) => {
      const field = contract.fieldMap[column.fieldKey];
      return {
        ...column,
        contractWritable: field?.gridEditable === true,
        getClipboardValue: (row: WorkbookQueryRow) => {
          const value =
            mutationRuntime.visibleEdit(
              contract.viewSchemaId,
              row.record_id,
              column.fieldKey,
            ) ?? row.cells[column.fieldKey]?.value;
          return field?.readKind === "collection"
            ? genericCellLabelForField(surface, column.fieldKey, value)
            : value;
        },
        editor:
          field?.gridEditable === true
            ? workbookGridEditorAdapter({
                commit: (draftValue, target) =>
                  commitGridEdit(field.fieldKey, draftValue, {
                    baseRowVersion: target.mutationIdentity.baseRowVersion,
                    recordId:
                      target.rowIdentity.kind === "core_record"
                        ? target.rowIdentity.recordId
                        : "",
                  }),
                field,
                readValue: (row: WorkbookQueryRow) =>
                  mutationRuntime.visibleEdit(
                    contract.viewSchemaId,
                    row.record_id,
                    field.fieldKey,
                  ) ?? row.cells[field.fieldKey]?.value,
                referenceOptions,
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
            <GenericMutationControl
              collectionMode="add"
              field={writableField}
              focusTargetRef={focusTargetRef}
              referenceOptions={referenceOptions}
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
              {genericCellLabelForField(
                surface,
                column.fieldKey,
                mutationRuntime.visibleEdit(
                  contract.viewSchemaId,
                  row.record_id,
                  column.fieldKey,
                ) ?? row.cells[column.fieldKey]?.value,
              )}
              <WorkbookCellPresenceMarker
                fieldKey={column.fieldKey}
                fieldLabel={field?.label ?? column.fieldKey}
                presences={collaborationProjection.editingPresenceForCell(
                  row.record_id,
                  column.fieldKey,
                )}
                recordId={row.record_id}
              />
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
          disabled={mutationState === "Syncing"}
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
        return genericInspector.ownerRecordActions.renderRecordActions(row);
      },
    };
  }, [
    contract.viewSchemaId,
    genericInspector,
    mutationState,
    surface,
    canCreateRows,
  ]);

  const focusDraftRow = useCallback(() => {
    const firstWritableField = createFields[0];
    if (!firstWritableField || !canCreateRows) {
      return;
    }
    window.setTimeout(() => {
      gridHandleRef.current?.focusDraftCell(firstWritableField.fieldKey);
    }, 0);
  }, [canCreateRows, createFields]);
  const dataState = workbookGridDataState({
    emptyAction: canCreateRows
      ? { label: "Add row", onInvoke: focusDraftRow }
      : undefined,
    emptyMessage: `No ${contract.title.toLocaleLowerCase()} records are available.`,
    loadState,
    onClearFilters,
    onRetry: () => void onRefresh(),
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
      onRequestInspectorClose={() => {
        genericInspector.close();
      }}
      primaryGrid={
        <GridViewport
          blockSizing="fill"
          style={gridShellStyle}
          testId={gridShellTestId(surface)}
        >
          <SemanticDataGrid
            ref={registerGridHandle}
            actionsColumn={rowActionsColumn}
            columns={columns}
            columnWidths={layoutState.columnWidths}
            dataState={dataState}
            density={density}
            draftRow={gridDraftRow}
            grouping={grouping}
            interactionMode={interactionMode}
            onActiveCellChange={(anchor) => {
              const recordId =
                anchor?.rowIdentity.kind === "core_record"
                  ? anchor.rowIdentity.recordId
                  : null;
              if (recordId !== null) {
                setEditRecordId(recordId);
              }
              if (recordId === null || anchor === null) {
                genericFocus.port.clear();
              } else {
                genericFocus.port.select({
                  fieldKey: anchor.fieldKey,
                  recordId,
                  viewSchemaId: contract.viewSchemaId,
                });
              }
              collaborationProjection.publishPresence({
                fieldKey: null,
                mode: recordId === null ? "idle" : "viewing",
                recordId,
              });
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
      }
      statusStrip={
        <WorkbookSurfaceStatusStrip
          activeSheetPresenceRecords={collaboration.activeSheetPresenceRecords}
          mutationError={
            mutationError?.primaryMessage ?? sharedMutation.secondaryMessage
          }
          mutationState={presentedMutationState}
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
