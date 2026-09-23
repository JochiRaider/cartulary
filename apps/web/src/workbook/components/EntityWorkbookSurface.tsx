import {
  type GridActionsColumn,
  type GridCellAnchor,
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
  entityInspectButtonTestId,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridActionsHeaderTestId,
  gridGroupRowTestId,
  gridShellTestId,
  workbookInlineDraftRowTestId,
  workbookRowActionMenuButtonTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import { MoreHorizontal } from "lucide-react";
import {
  type ReactNode,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import type { WorkbookClipboardPastePort } from "../adapters/WorkbookClipboardPastePort";
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
import { useEntityClipboardPasteController } from "../features/entities/useEntityClipboardPasteController";
import { useEntityWorkbookInspectorComposition } from "../features/entities/useEntityWorkbookInspectorComposition";
import { OrdinaryCreateControl } from "../features/ordinary/OrdinaryCreateControl";
import { OrdinaryCreateNotice } from "../features/ordinary/OrdinaryCreateNotice";
import { useOrdinaryCreateDraft } from "../features/ordinary/useOrdinaryCreateDraft";
import {
  useWorkbookFind,
  type WorkbookFindFocusLoan,
} from "../find/useWorkbookFind";
import { WorkbookFindControl } from "../find/WorkbookFindControl";
import { useWorkbookSemanticGridFocus } from "../hooks/useWorkbookSemanticGridFocus";
import { useRetainedInspectorRow } from "../inspector/useRetainedInspectorRow";
import { WorkbookExplicitPatchRecovery } from "../inspector/WorkbookExplicitPatchRecovery";
import type {
  WorkbookInspectorErrorPresentation,
  WorkbookInspectorFeedback,
} from "../inspector/workbookInspectorErrorModel";
import { workbookInspectorLocalErrorPresentation } from "../inspector/workbookInspectorErrorModel";
import { useWorkbookColumnSizingBinding } from "../layout/useWorkbookColumnSizingBinding";
import type { WorkbookSurfaceLayoutOwner } from "../layout/useWorkbookLayoutFacade";
import {
  WorkbookSurfaceLayout,
  workbookGridWithNoticeStyle,
  workbookSurfaceGridShellStyle,
} from "../layout/WorkbookSurfaceLayout";
import {
  applyWorkbookLayoutToColumns,
  workbookFrozenDataColumnPrefix,
} from "../layout/workbookColumnLayout";
import {
  entityCellPresentation,
  entityFindText,
} from "../models/entityCellPresentation";
import {
  type EntityRow,
  entityContractColumnWidth,
  entityRowFromApi,
} from "../models/entityWorkbookModel";
import {
  genericCellLabel,
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
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import {
  useWorkbookQueryPresentation,
  useWorkbookQueryRestart,
} from "../query/WorkbookQueryBrowsingContext";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import type { WorkbookViewQueryPort } from "../query/WorkbookViewQueryPort";
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
  WorkbookObservedStatusStrip,
} from "./WorkbookStatusStrip";
import {
  WorkbookViewBar,
  type WorkbookViewBarWorkingSetBinding,
} from "./WorkbookViewBar";

const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);

export type EntityWorkbookSurfaceProps = {
  readonly sheetRef: SheetRef;
  readonly incidentId: string;
  readonly clipboardPaste: WorkbookClipboardPastePort;
  continuityResetKey: string;
  entityType: EntityRow["entityType"];
  inspectorResetKey: string;
  gridEntryFocus: WorkbookGridEntryFocusOwner;
  viewBarWorkingSet: WorkbookViewBarWorkingSetBinding;
  layout: WorkbookSurfaceLayoutOwner;
  onSortChange: (sort: WorkbookQueryState["sort"]) => void;
  queryState: WorkbookQueryState;
  rows: EntityRow[];
  currentIncidentRole: WorkbookIncidentRole | null;
  currentUserId: string | null;
  entityIndex: Record<string, EntityRow>;
  onRefreshEntities: (options?: {
    readonly requireAcceptance?: boolean;
  }) => Promise<void>;
  onAuthorityUncertain?: (() => void) | undefined;
  loadState: WorkbookQueryLoadState;
  mutationRuntime: WorkbookMutationRuntime;
  onActivateConflict?: WorkbookConflictActivation | undefined;
  collaborationProjection: WorkbookCollaborationCoordinator;
  onClearFilters: () => void;
  viewQuery: WorkbookViewQueryPort;
};

function entityCellContent(
  entityType: EntityRow["entityType"],
  row: EntityRow,
  fieldKey: string,
): ReactNode {
  const presentation = entityCellPresentation(row, fieldKey);
  const aliasesField = `${entityType}.aliases`;
  if (fieldKey === aliasesField) {
    return row.aliasTexts.length > 0 ? (
      <div style={entityAliasListStyle}>
        {row.aliasTexts.map((alias) => (
          <span key={alias} style={tagChipStyle}>
            {alias}
          </span>
        ))}
      </div>
    ) : (
      "No aliases"
    );
  }
  return presentation.text;
}

export function EntityWorkbookSurface({
  sheetRef,
  incidentId,
  clipboardPaste: clipboardPastePort,
  continuityResetKey,
  entityType,
  inspectorResetKey,
  gridEntryFocus,
  viewBarWorkingSet,
  layout,
  rows,
  queryState,
  onSortChange,
  currentIncidentRole,
  currentUserId,
  entityIndex,
  onRefreshEntities,
  onAuthorityUncertain,
  loadState,
  mutationRuntime,
  onActivateConflict,
  collaborationProjection,
  onClearFilters,
  viewQuery,
}: EntityWorkbookSurfaceProps) {
  const {
    commands: { onColumnReorder, onColumnSizingIntent },
    snapshot: {
      chromeMode,
      density,
      incidentClosed,
      interactionMode,
      showStatusPresence,
      state: layoutState,
    },
  } = layout;
  const [selectedRecordId, setSelectedRecordId] = useState<string | null>(null);
  const inspectorContinuityTokenRef = useRef<WorkbookContinuityToken | null>(
    null,
  );
  const continuityPortRef = useRef<WorkbookContinuityPort | null>(null);
  const [createDraft, setCreateDraft] = useOrdinaryCreateDraft(
    mutationRuntime.ordinaryCreate,
    entityType === "host" ? hostsContract : identitiesContract,
  );
  const [mutationError, setMutationError] =
    useState<WorkbookInspectorErrorPresentation | null>(null);
  const [entityActionFeedback, setEntityActionFeedback] =
    useState<WorkbookInspectorFeedback | null>(null);
  const collaboration = useWorkbookCollaborationCoordinator(
    collaborationProjection,
  );

  const selectedVisible = rows.find((row) => row.recordId === selectedRecordId);
  const selectedPatch = selectedRecordId
    ? mutationRuntime.explicitPatches.latestRow(selectedRecordId)
    : null;
  const selectedCreate = selectedRecordId
    ? mutationRuntime.ordinaryCreate.latestRow(selectedRecordId)
    : null;
  const selectedObservations = useMemo(
    () => [
      selectedVisible,
      selectedPatch ? entityRowFromApi(selectedPatch, entityType) : null,
      selectedCreate ? entityRowFromApi(selectedCreate, entityType) : null,
    ],
    [selectedVisible, selectedPatch, selectedCreate, entityType],
  );
  const selectedEntity = useRetainedInspectorRow({
    recordId: selectedRecordId,
    rows: selectedObservations,
    sourceRow: (row) => row.rawRow,
    scope:
      mutationRuntime.recordReadScope?.actorId === currentUserId
        ? mutationRuntime.recordReadScope
        : null,
    readable: !!currentIncidentRole && loadState.kind !== "permission_denied",
  });
  const canMerge =
    currentIncidentRole === "reviewer" || currentIncidentRole === "admin";
  const contract = entityType === "host" ? hostsContract : identitiesContract;
  const surface: string = contract.viewSchemaId;
  const draftRowRecordId = `${surface}:draft-row`;
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

  const entityAnchorColumns = useMemo<readonly GridColumn<EntityRow>[]>(
    () =>
      workbookContractColumns<EntityRow>({
        contract,
        surface,
        widthForField: entityContractColumnWidth,
      }),
    [contract, surface],
  );
  const visibleEntityAnchorColumns = useMemo(
    () =>
      applyWorkbookLayoutToColumns(contract, entityAnchorColumns, layoutState),
    [contract, entityAnchorColumns, layoutState],
  );
  const draftEntityRawRow = useMemo<WorkbookQueryRow>(
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
  const draftEntityRow = useMemo<EntityRow>(
    () => entityRowFromApi(draftEntityRawRow, entityType),
    [draftEntityRawRow, entityType],
  );
  const entityGridRows = useMemo<readonly GridDataRow<EntityRow>[]>(
    () =>
      workbookGridRows({
        renderGutter: (recordId, ordinal) => (
          <WorkbookRowGutterContent
            recordId={recordId}
            ordinal={ordinal}
            presences={presenceForRow(collaboration.presence, recordId)}
          />
        ),
        getRecordId: (row: EntityRow) => row.recordId,
        getRowVersion: (row: EntityRow) => row.rowVersion,
        rows,
        surface,
      }),
    [rows, surface, collaboration.presence],
  );
  const entityDraftRow = useMemo<GridDraftRow<EntityRow> | undefined>(
    () =>
      !showOrdinaryDraft
        ? undefined
        : {
            kind: "draft",
            data: draftEntityRow,
            gutterContent: "+",
            gutterLabel: "Draft row",
            testId: workbookInlineDraftRowTestId(surface),
          },
    [showOrdinaryDraft, draftEntityRow, surface],
  );
  const grouping = useMemo<GridGroupingDescriptor<EntityRow> | null>(() => {
    const fieldKey = queryState.groupBy;
    if (fieldKey === null) {
      return null;
    }
    return {
      fieldKey,
      compareValues: compareWorkbookGroupValues,
      formatLabel: (value) => (value === null ? null : String(value)),
      getTestId: (groupFieldKey, _value, label) =>
        label === null
          ? undefined
          : gridGroupRowTestId(surface, groupFieldKey, label),
      getValue: (row) => workbookGroupValue(row.rawRow, fieldKey),
      label: contract.fieldMap[fieldKey]?.label ?? fieldKey,
    };
  }, [contract.fieldMap, queryState.groupBy, surface]);
  const gridHandleRef = useRef<GridHandle | null>(null);
  useWorkbookColumnSizingBinding({
    columns: entityAnchorColumns,
    commands: layout.commands,
    gridHandleRef,
  });
  const entityFocus = useWorkbookGridContinuity({
    columns: visibleEntityAnchorColumns,
    continuityResetKey,
    gridHandleRef,
    viewSchemaId: surface,
  });
  continuityPortRef.current = entityFocus.port;
  const entityInspector = useEntityWorkbookInspectorComposition({
    sheetRef,
    canMerge,
    contract,
    currentIncidentRole,
    entityActionFeedback,
    entityIndex,
    entityType,
    incidentClosed,
    inspectorResetKey,
    interactionMode,
    mutationError,
    mutationRuntime,
    onClearSurfaceSelection: () => {
      continuityPortRef.current?.clear();
      setSelectedRecordId(null);
    },
    onRefreshEntities,
    onAuthorityUncertain,
    onRestoreFocus: () => {
      const token = inspectorContinuityTokenRef.current;
      inspectorContinuityTokenRef.current = null;
      if (token !== null) continuityPortRef.current?.restore(token);
    },

    rows,
    selectedEntity,
    setEntityActionFeedback,
    setMutationError,
    setSelectedRecordId,
    viewQuery,
  });
  const browsing = useWorkbookQueryPresentation();
  const readFindText = useCallback(
    (anchor: GridCellAnchor) => {
      if (anchor.rowIdentity.kind !== "core_record") return [];
      const id = anchor.rowIdentity.recordId;
      const row = rows.find((row) => row.recordId === id);
      return row ? entityFindText(row, anchor.fieldKey, contract) : [];
    },
    [rows, contract],
  );
  const find = useWorkbookFind({
    lifetimeKey: `${incidentId}:${surface}`,
    configurationKey: JSON.stringify([
      continuityResetKey,
      sheetRef,
      queryState,
      layoutState,
    ]),
    browser: browsing?.find(surface),
    authorization: collaborationProjection,
    readable: !!currentIncidentRole && loadState.kind !== "permission_denied",
    stale: loadState.kind === "stale_error",
    readText: readFindText,
    captureFocus: (target): WorkbookFindFocusLoan | null => {
      const element = target ?? document.activeElement;
      const inspectorLoan = entityInspector.captureFindFocus(element);
      if (inspectorLoan) return inspectorLoan;
      if (
        !(element instanceof HTMLElement) ||
        !gridHandleRef.current?.getScrollElement()?.contains(element) ||
        !(
          element instanceof HTMLInputElement ||
          element instanceof HTMLTextAreaElement ||
          element instanceof HTMLSelectElement
        )
      )
        return null;
      const handle = gridHandleRef.current;
      return {
        restore: () => {
          if (gridHandleRef.current !== handle || !element.isConnected)
            return false;
          element.focus();
          return document.activeElement === element;
        },
      };
    },
  });
  const focusEntityDraft = useCallback(() => {
    const firstWritableField = createFields[0];
    if (!firstWritableField || !canCreateRows) return;
    void gridHandleRef.current?.requestFocus({
      kind: "draft",
      fieldKey: firstWritableField.fieldKey,
    });
  }, [canCreateRows, createFields]);
  const restartQuery = useWorkbookQueryRestart(surface, onRefreshEntities);
  const dataState = workbookGridDataState({
    emptyAction: canCreateRows
      ? { label: "Add row", onInvoke: focusEntityDraft }
      : undefined,
    emptyMessage: `No ${entityType === "host" ? "hosts" : "identities"} have been added.`,
    loadState,
    onClearFilters,
    onRetry: () => void restartQuery(),
    queryState,
    rowCount: entityGridRows.length,
    surfaceLabel: contract.title,
  });
  const gridEntryDraftFieldKeys = useMemo(
    () =>
      entityDraftRow === undefined
        ? []
        : visibleEntityAnchorColumns
            .filter((column) =>
              createFields.some((field) => field.fieldKey === column.fieldKey),
            )
            .map((column) => column.fieldKey),
    [createFields, entityDraftRow, visibleEntityAnchorColumns],
  );
  const surfaceCallbacks = useRef({
    onRefreshEntities,
  });
  surfaceCallbacks.current = {
    onRefreshEntities,
  };
  useEffect(
    () =>
      mutationRuntime.registerSurface(
        contract.viewSchemaId,
        () =>
          surfaceCallbacks.current.onRefreshEntities({
            requireAcceptance: true,
          }),
        async () => {
          await surfaceCallbacks.current.onRefreshEntities({
            requireAcceptance: true,
          });
        },
      ),
    [contract.viewSchemaId, mutationRuntime],
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
      const current = rows.find((row) => row.recordId === target.recordId);
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
      mutationRuntime.gridDrafts.update(identity, current.rawRow, draftValue);
      if (!mutationRuntime.gridDrafts.canAuthor())
        return {
          kind: "rejected_mutation",
          message: "Editing is unavailable with the current authorization.",
        };
      if (
        !mutationRuntime.hasPendingGridWrite(target.recordId) &&
        mutationRuntime.gridDrafts.staleFields(identity, current.rawRow).length
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
      setMutationError(null);
      const outcome = mutationRuntime.enqueuePatch({
        sheetRef,
        baseline: current.rawRow,
        baseRowVersion: target.baseRowVersion,
        changes: [change],
        fieldKey,
        focusKey: `${target.recordId}:${fieldKey}`,
        localValue: draftValue,
        recordId: target.recordId,
        rowLabel: genericRowLabel(contract, current.rawRow),
        surfaceLabel: contract.title,
        viewSchemaId: contract.viewSchemaId,
      });
      return outcome.kind === "admitted" ? outcome.completion : outcome;
    },
    [contract, mutationRuntime, rows, sheetRef],
  );
  const writablePasteFieldKeys = useMemo(
    () =>
      new Set(
        contract.fields
          .filter((field) => field.gridEditable)
          .map((field) => field.fieldKey),
      ),
    [contract.fields],
  );
  const { handlePaste: handleEntityPaste } = useEntityClipboardPasteController({
    canCreateRows,
    clipboardPaste: clipboardPastePort,
    commitGridEdit,
    grouped: grouping !== null,
    rows,
    setActionFeedback: setEntityActionFeedback,
    setMutationError,
    viewSchemaId: contract.viewSchemaId,
    writableFieldKeys: writablePasteFieldKeys,
  });
  const clipboardPaste = useMemo(
    () =>
      workbookClipboardPasteContract(handleEntityPaste, (message) => {
        setMutationError(workbookInspectorLocalErrorPresentation(message));
      }),
    [handleEntityPaste],
  );
  const entityColumns: readonly GridColumn<EntityRow>[] =
    visibleEntityAnchorColumns.map((column) => {
      const field = contract.fieldMap[column.fieldKey];
      const displayedValue = (row: EntityRow) => {
        const local = mutationRuntime.visibleEdit(
          contract.viewSchemaId,
          row.recordId,
          column.fieldKey,
        );
        return local === undefined
          ? row.rawRow.cells[column.fieldKey]?.value
          : local;
      };
      return {
        ...column,
        contractWritable: field?.gridEditable === true,
        draftWritable: field?.createWritable === true,
        getClipboardValue: (row: EntityRow) => {
          const value = displayedValue(row);
          return field?.readKind === "collection"
            ? genericCellLabel(value)
            : value;
        },
        editor:
          field?.gridEditable === true
            ? workbookGridEditorAdapter({
                drafts: mutationRuntime.gridDrafts,
                viewSchemaId: contract.viewSchemaId,
                readRow: (row: EntityRow) => row.rawRow,
                readCurrentRow: (row: EntityRow) =>
                  mutationRuntime.explicitPatches.latestRow(row.recordId) ??
                  row.rawRow,
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
                readValue: (row: EntityRow) => displayedValue(row),
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
        isCellContentCommitted: (row) =>
          mutationRuntime.visibleEdit(
            contract.viewSchemaId,
            row.recordId,
            column.fieldKey,
          ) === undefined,
        renderCell: ({ row }) => {
          const visibleEdit = mutationRuntime.visibleEdit(
            contract.viewSchemaId,
            row.recordId,
            column.fieldKey,
          );
          return (
            <WorkbookContinuityCell
              continuity={entityFocus.port}
              fieldKey={column.fieldKey}
              recordId={row.recordId}
              viewSchemaId={contract.viewSchemaId}
            >
              <WorkbookPresenceCellLayout
                marker={
                  <WorkbookCellPresenceMarker
                    fieldKey={column.fieldKey}
                    fieldLabel={field?.label ?? column.fieldKey}
                    presences={collaborationProjection.editingPresenceForCell(
                      row.recordId,
                      column.fieldKey,
                    )}
                    recordId={row.recordId}
                  />
                }
              >
                {visibleEdit === undefined
                  ? entityCellContent(entityType, row, column.fieldKey)
                  : genericCellLabel(visibleEdit)}
              </WorkbookPresenceCellLayout>
            </WorkbookContinuityCell>
          );
        },
      };
    });
  const registerGridHandle = useWorkbookSemanticGridFocus({
    dataRows: entityGridRows,
    dataState,
    draftFieldKeys: gridEntryDraftFieldKeys,
    focusOwner: gridEntryFocus,
    gridHandleRef,
    visibleColumns: entityColumns,
    viewSchemaId: surface,
  });
  const registerFindGrid = useCallback(
    (handle: GridHandle | null) => {
      registerGridHandle(handle);
      find.bindGrid(handle);
    },
    [registerGridHandle, find.bindGrid],
  );
  const entityActionsColumn: GridActionsColumn<EntityRow> = {
    headerTestId: gridActionsHeaderTestId(surface),
    label: "",
    width: 76,
    minWidth: 76,
    renderDraftCell: () => (
      <button
        data-testid={genericCreateSubmitTestId(contract.viewSchemaId)}
        disabled={
          !canCreateRows ||
          mutationRuntime.ordinaryCreate.busy(contract.viewSchemaId)
        }
        style={secondaryActionButtonStyle}
        type="button"
        onClick={() => {
          void submitEntityCreate();
        }}
      >
        Commit
      </button>
    ),
    renderCell: ({ data: row }) => (
      <span
        data-testid={workbookRowActionMenuButtonTestId(surface, row.recordId)}
      >
        <button
          data-testid={entityInspectButtonTestId(entityType, row.recordId)}
          aria-label={`Inspect ${row.label}`}
          style={rowMenuButtonStyle}
          type="button"
          onClick={() => {
            inspectorContinuityTokenRef.current = entityFocus.port.capture({
              fieldKey:
                entityFocus.snapshot.anchor?.fieldKey ??
                contract.fields[0]?.fieldKey ??
                "",
              recordId: row.recordId,
              viewSchemaId: contract.viewSchemaId,
            });
            entityInspector.openForRecord(row.recordId);
          }}
        >
          <MoreHorizontal aria-hidden="true" size={16} />
        </button>
      </span>
    ),
  };

  useEffect(() => {
    if (selectedRecordId === null || selectedEntity !== null) {
      return;
    }
    setSelectedRecordId(null);
  }, [selectedEntity, selectedRecordId]);

  async function submitEntityCreate() {
    await mutationRuntime.ordinaryCreate.submit(contract.viewSchemaId);
  }

  return (
    <WorkbookSurfaceLayout
      chromeMode={chromeMode}
      inspector={entityInspector.node}
      onRequestInspectorClose={() => {
        entityInspector.close();
      }}
      primaryGrid={
        <div style={{ ...workbookGridWithNoticeStyle, minWidth: 0 }}>
          <div>
            <WorkbookUnavailableGridDrafts
              store={mutationRuntime.gridDrafts}
              contract={contract}
              recordIds={rows.map((row) => row.recordId)}
              fieldKeys={entityColumns.map((column) => column.fieldKey)}
            />
            <WorkbookExplicitPatchRecovery
              owner={mutationRuntime.explicitPatches}
              viewSchemaId={contract.viewSchemaId}
            />
            <OrdinaryCreateNotice
              owner={mutationRuntime.ordinaryCreate}
              view={contract.viewSchemaId}
            />
          </div>
          <GridViewport
            blockSizing="fill"
            style={gridShellStyle}
            testId={gridShellTestId(surface)}
          >
            <SemanticDataGrid
              frozenDataColumnPrefix={workbookFrozenDataColumnPrefix(
                layoutState,
              )}
              rowGutter={workbookPresenceRowGutter}
              ref={registerFindGrid}
              activeRowIdentity={
                selectedRecordId === null
                  ? null
                  : { kind: "core_record", recordId: selectedRecordId }
              }
              allowPasteCreateRows
              actionsColumn={entityActionsColumn}
              columns={entityColumns}
              dataState={dataState}
              getCellState={({ anchor }) => ({
                findMatch: find.cellMatch(anchor),
              })}
              density={density}
              draftRow={entityDraftRow}
              grouping={grouping}
              interactionMode={interactionMode}
              onActiveCellChange={(anchor) => {
                const recordId =
                  anchor?.rowIdentity.kind === "core_record"
                    ? anchor.rowIdentity.recordId
                    : null;
                if (!find.isApplyingFocus()) setSelectedRecordId(recordId);
                if (recordId === null || anchor === null) {
                  entityFocus.port.clear();
                } else {
                  entityFocus.port.select({
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
              onColumnSizingIntent={onColumnSizingIntent}
              clipboardPaste={clipboardPaste}
              onSortChange={onSortChange}
              dataRows={entityGridRows}
              sort={queryState.sort}
              surface={{ kind: "view_schema", viewSchemaId: surface }}
            />
          </GridViewport>
        </div>
      }
      statusStrip={
        <WorkbookObservedStatusStrip
          presence={collaboration.presence.header}
          source={mutationRuntime.statusSource}
          sheetRef={sheetRef}
          chromeMode={chromeMode}
          onActivateConflict={onActivateConflict}
          showPresence={showStatusPresence}
          workbookFocusAnchor={entityFocus.snapshot.anchor}
        />
      }
      viewBar={
        <WorkbookViewBar
          addRowDisabled={!canCreateRows}
          chromeMode={chromeMode}
          findControls={
            <WorkbookFindControl
              binding={find.control}
              chromeMode={chromeMode}
            />
          }
          workingSet={viewBarWorkingSet}
          onAddRow={focusEntityDraft}
          onInspectorToggle={() => {
            inspectorContinuityTokenRef.current = entityFocus.port.capture();
            entityInspector.open();
          }}
          surface={surface}
        />
      }
      viewSchemaId={surface}
    />
  );
}

const gridShellStyle = {
  ...workbookSurfaceGridShellStyle,
};

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

const rowMenuButtonStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  width: "1.75rem",
  height: "1.75rem",
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  cursor: "pointer",
};

const draftCellPlaceholderStyle = {
  color: "var(--ct-colors-ink-subtle)",
};

const relationshipChipStyle = {
  display: "inline-flex",
  alignItems: "center",
  gap: "0.35rem",
  borderRadius: "var(--ct-component-chip-rounded)",
  padding: "var(--ct-component-chip-padding)",
  font: "inherit",
  lineHeight: 1.2,
  maxWidth: "100%",
  minWidth: 0,
  overflowWrap: "anywhere" as const,
};

const entityAliasListStyle = {
  display: "flex",
  flexWrap: "wrap" as const,
  gap: "0.35rem",
};

const tagChipStyle = {
  ...relationshipChipStyle,
  border: "var(--ct-component-chip-border)",
  background: "var(--ct-component-chip-backgroundColor)",
  color: "var(--ct-component-chip-textColor)",
};
