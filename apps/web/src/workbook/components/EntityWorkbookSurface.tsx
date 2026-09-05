import {
  type GridActionsColumn,
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
import { useWorkbookSemanticGridFocus } from "../hooks/useWorkbookSemanticGridFocus";
import {
  type WorkbookInspectorErrorPresentation,
  type WorkbookInspectorFeedback,
  workbookInspectorErrorPresentation,
  workbookInspectorLocalErrorPresentation,
} from "../inspector/workbookInspectorErrorModel";
import type { WorkbookSurfaceLayoutOwner } from "../layout/useWorkbookLayoutFacade";
import {
  WorkbookSurfaceLayout,
  workbookSurfaceGridShellStyle,
} from "../layout/WorkbookSurfaceLayout";
import { applyWorkbookLayoutToColumns } from "../layout/workbookColumnLayout";
import {
  type EntityRow,
  entityContractColumnWidth,
  entityGroupLabel,
  entityRowFromApi,
} from "../models/entityWorkbookModel";
import {
  buildGenericPatchChange,
  genericCellLabel,
  genericCreateMinimumMessage,
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
import { emptyGenericReferenceOptions } from "../models/workbookReferenceOptions";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import type {
  EntityMutationCommandPort,
  RecordRouteCommandPort,
  TimelineRelatedRecordPort,
} from "../mutations/workbookMutationCommandPorts";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import type { WorkbookViewQueryPort } from "../query/WorkbookViewQueryPort";
import { useWorkbookMutationRuntime } from "../runtime/useWorkbookMutationRuntime";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import { workbookClipboardPasteContract } from "../utils/workbookClipboard";
import { GenericMutationControl } from "./GenericMutationControl";
import { workbookGridEditorAdapter } from "./WorkbookGridEditorControl";
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

const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);

export type EntityWorkbookSurfaceProps = {
  readonly sheetRef: SheetRef;
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
  onRefreshEntities: () => Promise<void>;
  loadState: WorkbookQueryLoadState;
  mutationRuntime: WorkbookMutationRuntime;
  mutationCommands: EntityMutationCommandPort;
  onActivateConflict?: WorkbookConflictActivation | undefined;
  recordMutationCommands: RecordRouteCommandPort;
  relatedMutationCommands: TimelineRelatedRecordPort;
  collaborationProjection: WorkbookCollaborationCoordinator;
  onClearFilters: () => void;
  viewQuery: WorkbookViewQueryPort;
};

function entityCellContent(
  entityType: EntityRow["entityType"],
  row: EntityRow,
  fieldKey: string,
): ReactNode {
  const displayField =
    entityType === "host" ? "host.display_name" : "identity.display_name";
  const primaryField = entityType === "host" ? "host.hostname" : "identity.upn";
  const stateField =
    entityType === "host" ? "host.host_state" : "identity.identity_state";
  const aliasesField =
    entityType === "host" ? "host.aliases" : "identity.aliases";
  if (fieldKey === displayField) {
    return row.label;
  }
  if (fieldKey === primaryField) {
    return row.secondaryText || "None";
  }
  if (fieldKey === stateField) {
    return row.state;
  }
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
  if (fieldKey === "row_version") {
    return String(row.rowVersion);
  }
  return genericCellLabel(row.rawRow.cells[fieldKey]?.value);
}

export function EntityWorkbookSurface({
  sheetRef,
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
  loadState,
  mutationRuntime,
  mutationCommands,
  onActivateConflict,
  recordMutationCommands,
  relatedMutationCommands,
  collaborationProjection,
  onClearFilters,
  viewQuery,
}: EntityWorkbookSurfaceProps) {
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
  const [selectedRecordId, setSelectedRecordId] = useState<string | null>(null);
  const inspectorContinuityTokenRef = useRef<WorkbookContinuityToken | null>(
    null,
  );
  const continuityPortRef = useRef<WorkbookContinuityPort | null>(null);
  const [createDraft, setCreateDraft] = useState<Record<string, string>>(() =>
    initialGenericCreateDraft(
      entityType === "host" ? hostsContract : identitiesContract,
      null,
    ),
  );
  const [mutationError, setMutationError] =
    useState<WorkbookInspectorErrorPresentation | null>(null);
  const [entityActionFeedback, setEntityActionFeedback] =
    useState<WorkbookInspectorFeedback | null>(null);
  const [mutationPending, setMutationPending] = useState(false);
  const sharedMutation = useWorkbookMutationRuntime(mutationRuntime, sheetRef);
  const collaboration = useWorkbookCollaborationCoordinator(
    collaborationProjection,
  );

  const selectedEntity =
    rows.find((row) => row.recordId === selectedRecordId) ?? null;
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
  const entityReferenceOptions = useMemo(
    () => emptyGenericReferenceOptions(),
    [],
  );
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
      !canCreateRows
        ? undefined
        : {
            kind: "draft",
            data: draftEntityRow,
            gutterContent: "+",
            gutterLabel: "Draft row",
            testId: workbookInlineDraftRowTestId(surface),
          },
    [canCreateRows, draftEntityRow, surface],
  );
  const grouping = useMemo<GridGroupingDescriptor<EntityRow> | null>(() => {
    const fieldKey = queryState.groupBy;
    if (fieldKey === null) {
      return null;
    }
    return {
      fieldKey,
      formatLabel: (value) => (value === null ? null : String(value)),
      getTestId: (groupFieldKey, _value, label) =>
        label === null
          ? undefined
          : gridGroupRowTestId(surface, groupFieldKey, label),
      getValue: (row) => entityGroupLabel(row, fieldKey),
      label: contract.fieldMap[fieldKey]?.label ?? fieldKey,
    };
  }, [contract.fieldMap, queryState.groupBy, surface]);
  const gridHandleRef = useRef<GridHandle | null>(null);
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
    currentUserId,
    entityActionFeedback,
    entityIndex,
    entityType,
    incidentClosed,
    inspectorResetKey,
    interactionMode,
    mutationCommands,
    mutationError,
    mutationRuntime,
    mutationPending,
    onClearSurfaceSelection: () => {
      continuityPortRef.current?.clear();
      setSelectedRecordId(null);
      setCreateDraft(initialGenericCreateDraft(contract, null));
    },
    onRefreshEntities,
    onResetOwnerState: () => {
      setCreateDraft(initialGenericCreateDraft(contract, null));
    },
    onRestoreFocus: () => {
      const token = inspectorContinuityTokenRef.current;
      inspectorContinuityTokenRef.current = null;
      if (token !== null) continuityPortRef.current?.restore(token);
    },
    recordMutationCommands,
    relatedMutationCommands,
    rows,
    selectedEntity,
    setEntityActionFeedback,
    setMutationError,
    setMutationPending,
    setSelectedRecordId,
    viewQuery,
  });
  const focusEntityDraft = useCallback(() => {
    const firstWritableField = createFields[0];
    if (!firstWritableField || !canCreateRows) return;
    window.setTimeout(() => {
      gridHandleRef.current?.focusDraftCell(firstWritableField.fieldKey);
    }, 0);
  }, [canCreateRows, createFields]);
  const dataState = workbookGridDataState({
    emptyAction: canCreateRows
      ? { label: "Add row", onInvoke: focusEntityDraft }
      : undefined,
    emptyMessage: `No ${entityType === "host" ? "hosts" : "identities"} have been added.`,
    loadState,
    onClearFilters,
    onRetry: () => void onRefreshEntities(),
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
  useEffect(
    () =>
      mutationRuntime.registerSurface(
        contract.viewSchemaId,
        onRefreshEntities,
        async (_payload, conflict) => {
          await onRefreshEntities();
          window.setTimeout(() => {
            entityFocus.port.focus({
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
              entityFocus.port.focus({
                fieldKey: anchor.fieldKey,
                recordId: conflict.conflict.record_id,
                viewSchemaId: contract.viewSchemaId,
              });
            }
          }, 0);
        },
      ),
    [
      contract.viewSchemaId,
      entityFocus.port,
      mutationRuntime,
      onRefreshEntities,
    ],
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
      const current = rows.find((row) => row.recordId === target.recordId);
      if (
        current === undefined ||
        current.rowVersion !== target.baseRowVersion
      ) {
        return {
          kind: "stale_target",
          message: "The record changed before this edit was submitted.",
        };
      }
      const change = buildGenericPatchChange(field, draftValue);
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
      if (outcome.kind !== "accepted") {
        setMutationError(
          workbookInspectorLocalErrorPresentation(outcome.message),
        );
      } else {
        setSelectedRecordId(target.recordId);
      }
      return outcome;
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
  const beginMutation = useCallback(
    () => mutationRuntime.beginExplicitMutation(),
    [mutationRuntime],
  );
  const { handlePaste: handleEntityPaste } = useEntityClipboardPasteController({
    beginMutation,
    canCreateRows,
    clipboardPaste: clipboardPastePort,
    commitGridEdit,
    entityType,
    grouped: grouping !== null,
    onRefreshEntities,
    rows,
    setActionFeedback: setEntityActionFeedback,
    setMutationError,
    setSelectedRecordId,
    viewSchemaId: contract.viewSchemaId,
    writableFieldKeys: writablePasteFieldKeys,
  });
  const clipboardPaste = useMemo(
    () =>
      workbookClipboardPasteContract((intent) => {
        void handleEntityPaste(intent);
      }),
    [handleEntityPaste],
  );
  const entityColumns: readonly GridColumn<EntityRow>[] =
    visibleEntityAnchorColumns.map((column) => {
      const field = contract.fieldMap[column.fieldKey];
      return {
        ...column,
        contractWritable: field?.gridEditable === true,
        getClipboardValue: (row: EntityRow) => {
          const value =
            mutationRuntime.visibleEdit(
              contract.viewSchemaId,
              row.recordId,
              column.fieldKey,
            ) ?? row.rawRow.cells[column.fieldKey]?.value;
          return field?.readKind === "collection"
            ? genericCellLabel(value)
            : value;
        },
        editor:
          field?.gridEditable === true
            ? workbookGridEditorAdapter({
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
                readValue: (row: EntityRow) =>
                  mutationRuntime.visibleEdit(
                    contract.viewSchemaId,
                    row.recordId,
                    field.fieldKey,
                  ) ?? row.rawRow.cells[field.fieldKey]?.value,
                referenceOptions: entityReferenceOptions,
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
              referenceOptions={entityReferenceOptions}
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
  const entityActionsColumn: GridActionsColumn<EntityRow> = {
    headerTestId: gridActionsHeaderTestId(surface),
    label: "",
    width: 76,
    minWidth: 76,
    renderDraftCell: () => (
      <button
        data-testid={genericCreateSubmitTestId(contract.viewSchemaId)}
        disabled={mutationPending}
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
    if (
      selectedRecordId === null ||
      rows.some((row) => row.recordId === selectedRecordId)
    ) {
      return;
    }
    setSelectedRecordId(null);
  }, [rows, selectedRecordId]);

  async function submitEntityCreate() {
    if (!canCreateRows) return;
    if (!mutationCommands.canCreateRecord({ contract, draft: createDraft })) {
      setMutationError(
        workbookInspectorLocalErrorPresentation(
          genericCreateMinimumMessage(contract),
        ),
      );
      return;
    }
    setMutationPending(true);
    setMutationError(null);
    const finishMutation = mutationRuntime.beginExplicitMutation();
    try {
      const result = await mutationCommands.createRecord({
        contract,
        draft: createDraft,
      });
      if (result.kind === "rejected") {
        setMutationPending(false);
        setMutationError(workbookInspectorErrorPresentation(result.failure));
        return;
      }
      setCreateDraft(initialGenericCreateDraft(contract, null));
      await onRefreshEntities();
      setSelectedRecordId(result.value.row.record_id);
      setMutationPending(false);
    } finally {
      finishMutation();
    }
  }

  return (
    <WorkbookSurfaceLayout
      chromeMode={chromeMode}
      inspector={entityInspector.node}
      onRequestInspectorClose={() => {
        entityInspector.close();
      }}
      primaryGrid={
        <GridViewport
          blockSizing="fill"
          style={gridShellStyle}
          testId={gridShellTestId(surface)}
        >
          <SemanticDataGrid
            rowGutter={workbookPresenceRowGutter}
            ref={registerGridHandle}
            activeRowIdentity={
              selectedRecordId === null
                ? null
                : { kind: "core_record", recordId: selectedRecordId }
            }
            allowPasteCreateRows
            actionsColumn={entityActionsColumn}
            columns={entityColumns}
            columnWidths={layoutState.columnWidths}
            dataState={dataState}
            density={density}
            draftRow={entityDraftRow}
            grouping={grouping}
            interactionMode={interactionMode}
            onActiveCellChange={(anchor) => {
              const recordId =
                anchor?.rowIdentity.kind === "core_record"
                  ? anchor.rowIdentity.recordId
                  : null;
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
            onColumnWidthChange={onColumnWidthChange}
            clipboardPaste={clipboardPaste}
            onSortChange={onSortChange}
            dataRows={entityGridRows}
            sort={queryState.sort}
            surface={{ kind: "view_schema", viewSchemaId: surface }}
          />
        </GridViewport>
      }
      statusStrip={
        <WorkbookStatusStrip
          presence={collaboration.presence.header}
          status={sharedMutation}
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
