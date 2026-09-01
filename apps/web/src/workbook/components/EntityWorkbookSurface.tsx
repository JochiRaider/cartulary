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
  dataTestIdSelector,
  entityInspectButtonTestId,
  entityInspectorTestId,
  entityMergeControlTestId,
  entityMergePreconditionDetailsTestId,
  entityReusableIdentifierItemTestId,
  entityReusableIdentifiersSectionTestId,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  genericEditFieldSelectTestId,
  genericEditRecordSelectTestId,
  genericEditSubmitTestId,
  genericEditValueTestId,
  gridActionsHeaderTestId,
  gridGroupRowTestId,
  gridShellTestId,
  timelinePreviewRowTestId,
  workbookInlineDraftRowTestId,
  workbookRowActionMenuButtonTestId,
} from "@cartulary/ui-contracts";
import type { InspectorDisabledCondition } from "@cartulary/view-contracts";
import {
  type InspectorFeatureGroup,
  requireViewContract,
} from "@cartulary/view-contracts";
import { MoreHorizontal, X } from "lucide-react";
import {
  type ReactNode,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
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
import { useEntityMergeController } from "../features/entities/useEntityMergeController";
import { useEntityTimelinePreview } from "../hooks/useEntityTimelinePreview";
import { InspectorCreateRelatedWorkflow } from "../inspector/InspectorCreateRelatedWorkflow";
import { WorkbookInspectorPublicError } from "../inspector/presentation/WorkbookInspectorFeedback";
import {
  WorkbookInspectorPanelSection,
  WorkbookInspectorShell,
} from "../inspector/presentation/WorkbookInspectorShell";
import type { WorkbookInspectorSubjectPresentation } from "../inspector/presentation/workbookInspectorPresentationModel";
import { useInspectorCreateRelatedWorkflow } from "../inspector/useInspectorCreateRelatedWorkflow";
import { useWorkbookInspectorCoordinator } from "../inspector/useWorkbookInspectorCoordinator";
import { WorkbookInspectorContextualActions } from "../inspector/WorkbookInspectorContextualActions";
import { WorkbookInspectorRecordHistory } from "../inspector/WorkbookInspectorRecordHistory";
import {
  type WorkbookInspectorErrorPresentation,
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
  selectWorkbookEditTarget,
  workbookCreationAvailable,
} from "../models/genericWorkbookModel";
import {
  workbookContractColumns,
  workbookGridRows,
} from "../models/workbookContractRows";
import {
  type WorkbookQueryLoadState,
  workbookGridDataState,
} from "../models/workbookGridState";
import { inspectorPanelIsDeclared } from "../models/workbookInspectorModel";
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
import { timelineRelationshipChipPresentation } from "../timeline/models/workbookMentionChips";
import { workbookClipboardPasteContract } from "../utils/workbookClipboard";
import { GenericMutationControl } from "./GenericMutationControl";
import { workbookGridEditorAdapter } from "./WorkbookGridEditorControl";
import { WorkbookCellPresenceMarker } from "./WorkbookPresenceMarkers";
import { WorkbookRelationshipChip } from "./WorkbookRelationshipChip";
import {
  type WorkbookConflictActivation,
  WorkbookSurfaceStatusStrip,
} from "./WorkbookStatusStrip";
import { WorkbookViewBar } from "./WorkbookViewBar";

const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);

type WorkbookMutationSaveState = "Syncing" | "Saved" | "Conflict";

export type EntityWorkbookSurfaceProps = {
  continuityResetKey: string;
  entityType: EntityRow["entityType"];
  inspectorResetKey: string;
  queryControls?: ReactNode | undefined;
  savedViewSelector?: ReactNode | undefined;
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
  continuityResetKey,
  entityType,
  inspectorResetKey,
  queryControls,
  savedViewSelector,
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
  const mergeResetRef = useRef<() => void>(() => undefined);
  const [editRecordId, setEditRecordId] = useState("");
  const [editFieldKey, setEditFieldKey] = useState("");
  const [editValue, setEditValue] = useState("");
  const [aliasDraft, setAliasDraft] = useState("");
  const aliasInputRef = useRef<HTMLInputElement | null>(null);
  const [createDraft, setCreateDraft] = useState<Record<string, string>>(() =>
    initialGenericCreateDraft(
      entityType === "host" ? hostsContract : identitiesContract,
      null,
    ),
  );
  const [mutationError, setMutationError] =
    useState<WorkbookInspectorErrorPresentation | null>(null);
  const [entityActionMessage, setEntityActionMessage] =
    useState<WorkbookInspectorErrorPresentation | null>(null);
  const [mutationState, setMutationState] =
    useState<WorkbookMutationSaveState>("Saved");
  const sharedMutation = useWorkbookMutationRuntime(
    mutationRuntime,
    entityType === "host"
      ? hostsContract.viewSchemaId
      : identitiesContract.viewSchemaId,
  );
  const collaboration = useWorkbookCollaborationCoordinator(
    collaborationProjection,
  );
  const presentedMutationState =
    mutationState === "Saved" ? sharedMutation.primaryLabel : mutationState;
  const { clearTimelinePreview, loadTimelinePreview, timelinePreviewRows } =
    useEntityTimelinePreview({
      entityType,
      viewQuery,
    });

  const selectedEntity =
    rows.find((row) => row.recordId === selectedRecordId) ?? null;
  const entityInspectorDisabledTokens = useMemo(() => {
    const tokens = new Set<InspectorDisabledCondition>();
    if (selectedEntity === null) tokens.add("no_row_selected");
    else tokens.add("record_not_deleted");
    tokens.add("rollback_target_unavailable");
    tokens.add("pivot_target_unavailable");
    if (rows.length < 2) tokens.add("merge_target_unavailable");
    if (incidentClosed) tokens.add("incident_closed");
    return tokens;
  }, [incidentClosed, rows.length, selectedEntity]);
  const canMerge =
    currentIncidentRole === "reviewer" || currentIncidentRole === "admin";
  const contract = entityType === "host" ? hostsContract : identitiesContract;
  const inspectorConfig = contract.inspectorConfig;
  const inspector = useWorkbookInspectorCoordinator({
    actionPorts: {
      resetOwnerState: ({ scope }) => {
        setEditRecordId("");
        setEditFieldKey("");
        setEditValue("");
        setAliasDraft("");
        setCreateDraft(initialGenericCreateDraft(contract, null));
        mergeResetRef.current();
        clearTimelinePreview();
        setEntityActionMessage(null);
        if (scope === "surface") {
          continuityPortRef.current?.clear();
          setSelectedRecordId(null);
        }
      },
      restoreFocus: () => {
        const token = inspectorContinuityTokenRef.current;
        inspectorContinuityTokenRef.current = null;
        if (token !== null) {
          continuityPortRef.current?.restore(token);
        }
      },
    },
    config: inspectorConfig,
    lifecycleKey: inspectorResetKey,
    subject:
      selectedEntity === null
        ? null
        : {
            recordId: selectedEntity.recordId,
            rowVersion: selectedEntity.rowVersion,
            viewSchemaId: contract.viewSchemaId,
          },
  });
  const isInspectorOpen = inspector.snapshot.isOpen;
  const showDetailsPanel = inspectorPanelIsDeclared(inspectorConfig, "details");
  const showRelationshipsPanel = inspectorPanelIsDeclared(
    inspectorConfig,
    "relationships",
  );
  const surface: string = contract.viewSchemaId;
  const draftRowRecordId = `${surface}:draft-row`;
  const createFields = useMemo(
    () => contract.fields.filter((field) => field.createWritable),
    [contract],
  );
  const canCreateRows =
    interactionMode.kind === "editable" && workbookCreationAvailable(contract);
  const editableEntityFields = useMemo(
    () => contract.fields.filter((field) => field.writeKind === "direct_value"),
    [contract],
  );
  const entityReferenceOptions = useMemo(
    () => emptyGenericReferenceOptions(),
    [],
  );
  const createRelatedWorkflow = useInspectorCreateRelatedWorkflow({
    currentUserId,
    mutationCommands: relatedMutationCommands,
    onCreated: onRefreshEntities,
    onMessage: (message) => {
      setEntityActionMessage(
        message === null
          ? null
          : workbookInspectorLocalErrorPresentation(message),
      );
    },
    selectedSubject:
      selectedEntity === null
        ? null
        : {
            cells: selectedEntity.rawRow.cells,
            recordId: selectedEntity.recordId,
            rowVersion: selectedEntity.rowVersion,
          },
  });
  const clearEntityMergeDrafts = useCallback(() => {
    setEditRecordId("");
    setEditFieldKey("");
    setEditValue("");
    setAliasDraft("");
  }, []);
  const merge = useEntityMergeController({
    canMerge,
    clearDrafts: clearEntityMergeDrafts,
    lifecycleResetKey: inspectorResetKey,
    loadSurvivorPreview: loadTimelinePreview,
    mutationCommands,
    onRefreshEntities,
    retargetSurvivor: setSelectedRecordId,
    rows,
    selectedEntity,
  });
  mergeResetRef.current = merge.commands.clearPlan;
  const {
    candidateId: mergeCandidateId,
    loser: loserEntity,
    message: mergeMessage,
    plan: mergePlan,
    preconditionDetails: mergePreconditionDetails,
    reason: mergeReason,
  } = merge.snapshot;
  const presentedEntityActionMessage = mergeMessage ?? entityActionMessage;
  const selectedEntityRecordKey = selectedEntity?.recordId ?? "";
  const selectedEntityPlanInvalidationKey =
    selectedEntity === null
      ? `none:${canMerge}`
      : `${selectedEntity.recordId}:${selectedEntity.rowVersion}:${canMerge}`;
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
        getRecordId: (row: EntityRow) => row.recordId,
        getRowVersion: (row: EntityRow) => row.rowVersion,
        rows,
        surface,
      }),
    [rows, surface],
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
  const focusEntityDraft = useCallback(() => {
    const firstWritableField = createFields[0];
    if (!firstWritableField || !canCreateRows) return;
    window.setTimeout(() => {
      document
        .querySelector<HTMLElement>(
          dataTestIdSelector(
            genericCreateFieldTestId(firstWritableField.fieldKey),
          ),
        )
        ?.focus({ preventScroll: true });
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
    [contract, mutationRuntime, rows],
  );
  const handleEntityPaste = useCallback(
    async (intent: GridCellPasteIntent) => {
      const clipboardText = intent.input.rawText;
      const targetResolution = intent.targetResolution;
      const values =
        intent.input.kind === "scalar"
          ? [[intent.input.value]]
          : intent.input.values;
      if (
        targetResolution === undefined ||
        targetResolution.columns.length === 0 ||
        targetResolution.rowTargets.length !== values.length
      ) {
        setMutationError(
          workbookInspectorLocalErrorPresentation(
            "Paste targets are incomplete or incompatible.",
          ),
        );
        return;
      }
      if (values.length === 1 && values[0]?.length === 1) {
        const rowTarget = targetResolution.rowTargets[0];
        if (rowTarget?.kind !== "record") {
          setMutationError(
            workbookInspectorLocalErrorPresentation(
              "Scalar paste requires an existing record target.",
            ),
          );
          return;
        }
        const outcome = await commitGridEdit(
          targetResolution.columns[0] ?? intent.target.fieldKey,
          values[0]?.[0] ?? "",
          {
            baseRowVersion: rowTarget.mutationIdentity.baseRowVersion,
            recordId: rowTarget.rowIdentity.recordId,
          },
        );
        if (outcome.kind !== "accepted") {
          setMutationError(
            workbookInspectorLocalErrorPresentation(outcome.message),
          );
        }
        return;
      }
      if (grouping !== null) {
        setMutationError(
          workbookInspectorLocalErrorPresentation(
            "Rectangular entity creation paste is unavailable while grouped.",
          ),
        );
        return;
      }
      if (!canCreateRows) {
        setMutationError(
          workbookInspectorLocalErrorPresentation(
            "Row creation is unavailable in the current view mode.",
          ),
        );
        return;
      }
      setEntityActionMessage(null);
      const result = await mutationCommands.pasteCreate({
        clipboardText,
        columns: targetResolution.columns,
        format: intent.input.kind === "table" ? intent.input.format : "csv",
        startFieldKey: intent.target.fieldKey,
        targetCount: targetResolution.rowTargets.length,
        viewSchemaId: contract.viewSchemaId,
      });
      if (result.kind === "rejected") {
        setEntityActionMessage(
          workbookInspectorErrorPresentation(result.failure),
        );
        return;
      }
      const firstRow = result.value.rows[0];
      await onRefreshEntities();
      if (firstRow) setSelectedRecordId(firstRow.record_id);
      setEntityActionMessage(
        workbookInspectorLocalErrorPresentation(
          `Paste applied to ${result.value.rows.length} ${entityType === "host" ? "host" : "identity"} row${result.value.rows.length === 1 ? "" : "s"}.`,
        ),
      );
    },
    [
      commitGridEdit,
      canCreateRows,
      contract.viewSchemaId,
      entityType,
      grouping,
      mutationCommands,
      onRefreshEntities,
    ],
  );
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
        renderDraftCell: () => {
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
              {visibleEdit === undefined
                ? entityCellContent(entityType, row, column.fieldKey)
                : genericCellLabel(visibleEdit)}
              <WorkbookCellPresenceMarker
                fieldKey={column.fieldKey}
                fieldLabel={field?.label ?? column.fieldKey}
                presences={collaborationProjection.editingPresenceForCell(
                  row.recordId,
                  column.fieldKey,
                )}
                recordId={row.recordId}
              />
            </WorkbookContinuityCell>
          );
        },
      };
    });
  const entityActionsColumn: GridActionsColumn<EntityRow> = {
    headerTestId: gridActionsHeaderTestId(surface),
    label: "",
    width: 76,
    minWidth: 76,
    renderDraftCell: () => (
      <button
        data-testid={genericCreateSubmitTestId(contract.viewSchemaId)}
        disabled={mutationState === "Syncing"}
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
            setSelectedRecordId(row.recordId);
            setEntityActionMessage(null);
            merge.commands.reset();
            inspectorContinuityTokenRef.current = entityFocus.port.capture({
              fieldKey:
                entityFocus.snapshot.anchor?.fieldKey ??
                contract.fields[0]?.fieldKey ??
                "",
              recordId: row.recordId,
              viewSchemaId: contract.viewSchemaId,
            });
            inspector.commands.open();
          }}
        >
          <MoreHorizontal aria-hidden="true" size={16} />
        </button>
      </span>
    ),
  };
  const { row: selectedEditRow, field: selectedEditField } =
    selectWorkbookEditTarget({
      fieldKey: editFieldKey,
      fields: editableEntityFields,
      getRecordId: (row: EntityRow) => row.recordId,
      recordId: editRecordId,
      rows,
    });

  useEffect(() => {
    if (selectedEntityRecordKey === "") {
      clearTimelinePreview();
    }
  }, [clearTimelinePreview, selectedEntityRecordKey]);

  useEffect(() => {
    void selectedEntityPlanInvalidationKey;
    if (!isInspectorOpen || selectedEntityRecordKey === "") {
      clearTimelinePreview();
      return;
    }
    void loadTimelinePreview(selectedEntityRecordKey);
  }, [
    clearTimelinePreview,
    isInspectorOpen,
    loadTimelinePreview,
    selectedEntityPlanInvalidationKey,
    selectedEntityRecordKey,
  ]);

  useEffect(() => {
    if (
      selectedRecordId === null ||
      rows.some((row) => row.recordId === selectedRecordId)
    ) {
      return;
    }
    setSelectedRecordId(null);
  }, [rows, selectedRecordId]);

  useEffect(() => {
    if (selectedEditRow === null || selectedEditField === null) {
      setEditValue("");
      return;
    }
    const value =
      selectedEditRow.rawRow.cells[selectedEditField.fieldKey]?.value;
    setEditValue(value === null || value === undefined ? "" : String(value));
  }, [selectedEditField, selectedEditRow]);

  async function submitEntityEdit() {
    if (selectedEditRow === null || selectedEditField === null) {
      setMutationError(
        workbookInspectorLocalErrorPresentation("invalid_mutation_payload"),
      );
      return;
    }
    const change = buildGenericPatchChange(selectedEditField, editValue);
    if (change === null) {
      setMutationError(
        workbookInspectorLocalErrorPresentation(
          "Provide a value, or leave clearable fields empty to clear them.",
        ),
      );
      return;
    }
    const finishMutation = mutationRuntime.beginExplicitMutation();
    try {
      setMutationState("Syncing");
      setMutationError(null);
      const result = await mutationCommands.patchRecord({
        baseRowVersion: selectedEditRow.rowVersion,
        changes: [change],
        purpose: "entity-patch",
        recordId: selectedEditRow.recordId,
        viewSchemaId: contract.viewSchemaId,
      });
      if (result.kind === "rejected") {
        if (result.failure.kind === "same_field_conflict") {
          mutationRuntime.registerConflict({
            conflict: result.failure.conflict,
            focusKey: `${selectedEditRow.recordId}:${selectedEditField.fieldKey}`,
            rowLabel: selectedEditRow.label,
            surfaceLabel: contract.title,
            viewSchemaId: contract.viewSchemaId,
          });
        }
        setMutationState("Conflict");
        setMutationError(workbookInspectorErrorPresentation(result.failure));
        return;
      }
      await onRefreshEntities();
      setSelectedRecordId(selectedEditRow.recordId);
      setMutationState("Saved");
    } finally {
      finishMutation();
    }
  }

  async function submitAliasActions(
    actions: Array<
      | { op: "add_alias"; alias_text: string }
      | { op: "remove_alias"; item_ref: string }
    >,
  ) {
    if (selectedEntity === null) {
      setMutationError(
        workbookInspectorLocalErrorPresentation("invalid_mutation_payload"),
      );
      return;
    }
    const aliasFieldKey =
      entityType === "host" ? "host.aliases" : "identity.aliases";
    const finishMutation = mutationRuntime.beginExplicitMutation();
    try {
      setMutationState("Syncing");
      setMutationError(null);
      const result = await mutationCommands.patchRecord({
        baseRowVersion: selectedEntity.rowVersion,
        changes: [
          {
            field_key: aliasFieldKey,
            action_payload: { kind: "collection_actions_v1", actions },
          },
        ],
        purpose: `entity-alias-${selectedEntity.recordId}`,
        recordId: selectedEntity.recordId,
        viewSchemaId: contract.viewSchemaId,
      });
      if (result.kind === "rejected") {
        if (result.failure.kind === "same_field_conflict") {
          mutationRuntime.registerConflict({
            conflict: result.failure.conflict,
            focusKey: `${selectedEntity.recordId}:${aliasFieldKey}`,
            rowLabel: selectedEntity.label,
            surfaceLabel: contract.title,
            viewSchemaId: contract.viewSchemaId,
          });
        }
        setMutationState("Conflict");
        setMutationError(workbookInspectorErrorPresentation(result.failure));
        return;
      }
      setAliasDraft("");
      await onRefreshEntities();
      setSelectedRecordId(selectedEntity.recordId);
      setMutationState("Saved");
      requestAnimationFrame(() => aliasInputRef.current?.focus());
    } finally {
      finishMutation();
    }
  }

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
    setMutationState("Syncing");
    setMutationError(null);
    const finishMutation = mutationRuntime.beginExplicitMutation();
    try {
      const result = await mutationCommands.createRecord({
        contract,
        draft: createDraft,
      });
      if (result.kind === "rejected") {
        setMutationState("Conflict");
        setMutationError(workbookInspectorErrorPresentation(result.failure));
        return;
      }
      setCreateDraft(initialGenericCreateDraft(contract, null));
      await onRefreshEntities();
      setSelectedRecordId(result.value.row.record_id);
      setMutationState("Saved");
    } finally {
      finishMutation();
    }
  }

  async function executeEntityRecordLifecycle(
    featureGroup: InspectorFeatureGroup,
  ): Promise<boolean> {
    if (
      featureGroup.routeBinding.kind !== "record_action" ||
      featureGroup.routeBinding.owner !== "record_delete_route"
    ) {
      return false;
    }
    if (selectedEntity === null) {
      setEntityActionMessage(
        workbookInspectorLocalErrorPresentation(
          "Select a saved row before running this action.",
        ),
      );
      return true;
    }
    const finishMutation = mutationRuntime.beginExplicitMutation();
    try {
      setMutationState("Syncing");
      setMutationError(null);
      const outcome = await recordMutationCommands.execute({
        action: "delete",
        baseRowVersion: selectedEntity.rowVersion,
        reason: `Deleted from the ${contract.title} inspector`,
        recordId: selectedEntity.recordId,
      });
      if (outcome.kind === "rejected") {
        setMutationState("Conflict");
        setMutationError(workbookInspectorErrorPresentation(outcome.failure));
        return true;
      }
      setSelectedRecordId(null);
      inspector.commands.completeAction();
      await onRefreshEntities();
      setMutationState("Saved");
      return true;
    } finally {
      finishMutation();
    }
  }

  const inspectorSubjectPresentation: WorkbookInspectorSubjectPresentation | null =
    selectedEntity === null
      ? null
      : {
          label: selectedEntity.label,
          recordId: selectedEntity.recordId,
          rowVersion: selectedEntity.rowVersion,
          stateLabel: selectedEntity.state,
          surfaceLabel: contract.title,
        };

  return (
    <WorkbookSurfaceLayout
      chromeMode={chromeMode}
      inspector={
        isInspectorOpen ? (
          <WorkbookInspectorShell
            accessibleLabel={`${contract.title} inspector`}
            noRowHeading={`${contract.title} inspector`}
            subject={inspectorSubjectPresentation}
            testId={entityInspectorTestId(entityType)}
            viewSchemaId={contract.viewSchemaId}
            onClose={() => {
              inspector.commands.close({ restoreFocus: true });
            }}
          >
            {inspectorConfig.panels.map((panel) => (
              <WorkbookInspectorPanelSection
                config={inspectorConfig}
                key={panel.panelId}
                panelId={panel.panelId}
              >
                {inspectorSubjectPresentation === null ? null : (
                  <WorkbookInspectorContextualActions
                    config={inspectorConfig}
                    currentIncidentRole={currentIncidentRole}
                    disabledTokens={entityInspectorDisabledTokens}
                    featureGroups={inspectorConfig.featureGroups.filter(
                      (featureGroup) =>
                        featureGroup.panelId === panel.panelId &&
                        (featureGroup.routeBinding.kind === "view_row_create" ||
                          (featureGroup.routeBinding.kind === "record_action" &&
                            featureGroup.routeBinding.owner ===
                              "record_delete_route")),
                    )}
                    subject={inspectorSubjectPresentation}
                    onAction={(featureGroup) => {
                      if (createRelatedWorkflow.commands.begin(featureGroup)) {
                        return;
                      }
                      void executeEntityRecordLifecycle(featureGroup);
                    }}
                  />
                )}
                {createRelatedWorkflow.snapshot.workflow?.featureGroup
                  .panelId === panel.panelId ? (
                  <InspectorCreateRelatedWorkflow
                    referenceOptions={entityReferenceOptions}
                    state={createRelatedWorkflow.snapshot.workflow}
                    onCancel={createRelatedWorkflow.commands.cancel}
                    onSubmit={() => {
                      void createRelatedWorkflow.commands.submit();
                    }}
                    onUpdateDraft={createRelatedWorkflow.commands.updateDraft}
                  />
                ) : null}
                {panel.panelId === "history" ? (
                  <WorkbookInspectorRecordHistory
                    canMutate={
                      interactionMode.kind === "editable" &&
                      currentIncidentRole !== null &&
                      currentIncidentRole !== "viewer"
                    }
                    commands={recordMutationCommands}
                    subject={
                      selectedEntity === null
                        ? null
                        : {
                            recordId: selectedEntity.recordId,
                            rowVersion: selectedEntity.rowVersion,
                          }
                    }
                    onMessage={(message) => {
                      setEntityActionMessage(
                        workbookInspectorLocalErrorPresentation(message),
                      );
                    }}
                    onRefresh={onRefreshEntities}
                  />
                ) : null}
              </WorkbookInspectorPanelSection>
            ))}
            {showDetailsPanel &&
            editableEntityFields.length > 0 &&
            rows.length > 0 ? (
              <section style={inspectorSectionStyle}>
                <h3 style={sectionTitleStyle}>Edit cell</h3>
                <div style={inspectorControlStackStyle}>
                  <select
                    data-testid={genericEditRecordSelectTestId(
                      contract.viewSchemaId,
                    )}
                    style={selectStyle}
                    value={editRecordId}
                    onChange={(event) => {
                      setEditRecordId(event.target.value);
                    }}
                  >
                    <option value="">Row</option>
                    {rows.map((row) => (
                      <option key={row.recordId} value={row.recordId}>
                        {genericRowLabel(contract, row.rawRow)}
                      </option>
                    ))}
                  </select>
                  <select
                    data-testid={genericEditFieldSelectTestId(
                      contract.viewSchemaId,
                    )}
                    style={selectStyle}
                    value={editFieldKey}
                    onChange={(event) => {
                      setEditFieldKey(event.target.value);
                    }}
                  >
                    <option value="">Field</option>
                    {editableEntityFields.map((field) => (
                      <option key={field.fieldKey} value={field.fieldKey}>
                        {field.label}
                      </option>
                    ))}
                  </select>
                  {selectedEditField ? (
                    <GenericMutationControl
                      collectionMode="add"
                      field={selectedEditField}
                      referenceOptions={entityReferenceOptions}
                      testId={genericEditValueTestId(contract.viewSchemaId)}
                      value={editValue}
                      onChange={setEditValue}
                    />
                  ) : null}
                  <button
                    data-testid={genericEditSubmitTestId(contract.viewSchemaId)}
                    disabled={mutationState === "Syncing"}
                    style={actionButtonStyle}
                    type="button"
                    onClick={() => {
                      void submitEntityEdit();
                    }}
                  >
                    Update
                  </button>
                </div>
                {mutationError ? (
                  <WorkbookInspectorPublicError error={mutationError} />
                ) : null}
              </section>
            ) : null}
            {showDetailsPanel && selectedEntity ? (
              <section style={inspectorSectionStyle}>
                <h3 style={sectionTitleStyle}>Aliases</h3>
                <div style={entityAliasListStyle}>
                  {selectedEntity.aliases.map((alias) => (
                    <span key={alias.itemRef} style={tagChipStyle}>
                      {alias.displayText}
                      <button
                        aria-label={`Remove alias ${alias.displayText}`}
                        disabled={mutationState === "Syncing"}
                        style={aliasRemoveButtonStyle}
                        type="button"
                        onClick={() => {
                          void submitAliasActions([
                            { op: "remove_alias", item_ref: alias.itemRef },
                          ]);
                        }}
                      >
                        <X aria-hidden="true" size={12} />
                      </button>
                    </span>
                  ))}
                </div>
                <div style={aliasAddRowStyle}>
                  <input
                    ref={aliasInputRef}
                    aria-label="Alias text"
                    maxLength={256}
                    style={inputStyle}
                    value={aliasDraft}
                    onChange={(event) => setAliasDraft(event.target.value)}
                  />
                  <button
                    disabled={
                      mutationState === "Syncing" || aliasDraft.trim() === ""
                    }
                    style={secondaryActionButtonStyle}
                    type="button"
                    onClick={() => {
                      void submitAliasActions([
                        { op: "add_alias", alias_text: aliasDraft },
                      ]);
                    }}
                  >
                    Add alias
                  </button>
                </div>
              </section>
            ) : null}
            {selectedEntity ? (
              <>
                {showDetailsPanel ? (
                  <section style={inspectorSectionStyle}>
                    <h3 style={sectionTitleStyle}>Identifiers</h3>
                    <ul style={flatListStyle}>
                      {selectedEntity.identifiers.length > 0 ? (
                        selectedEntity.identifiers.map((identifier) => (
                          <li key={identifier.key}>
                            {identifier.label}: {identifier.value}
                          </li>
                        ))
                      ) : (
                        <li>No exact-match identifiers visible.</li>
                      )}
                    </ul>
                  </section>
                ) : null}

                {showDetailsPanel ? (
                  <section
                    data-testid={entityReusableIdentifiersSectionTestId(
                      entityType,
                      selectedEntity.recordId,
                    )}
                    style={reusableIdentifierSectionStyle}
                  >
                    <div style={sectionHeadingRowStyle}>
                      <h3 style={sectionTitleStyle}>Reusable identifiers</h3>
                      <span style={readOnlyBadgeStyle}>Read-only</span>
                    </div>
                    <ul style={flatListStyle}>
                      {selectedEntity.reusableIdentifiers.length > 0 ? (
                        selectedEntity.reusableIdentifiers.map((identifier) => (
                          <li
                            data-testid={entityReusableIdentifierItemTestId(
                              entityType,
                              selectedEntity.recordId,
                              identifier.itemRef,
                            )}
                            key={identifier.itemRef}
                          >
                            {identifier.label}: {identifier.displayText}
                          </li>
                        ))
                      ) : (
                        <li>No reusable identifiers carried forward.</li>
                      )}
                    </ul>
                  </section>
                ) : null}

                {showRelationshipsPanel && canMerge ? (
                  <section style={inspectorSectionStyle}>
                    <h3 style={sectionTitleStyle}>Merge</h3>
                    <label style={labelStyle}>
                      Merge loser
                      <select
                        data-testid={entityMergeControlTestId("loser-record")}
                        style={selectStyle}
                        value={mergeCandidateId}
                        onChange={(event) => {
                          setEntityActionMessage(null);
                          merge.commands.selectCandidate(event.target.value);
                        }}
                      >
                        <option value="">Select duplicate</option>
                        {rows
                          .filter(
                            (row) => row.recordId !== selectedEntity.recordId,
                          )
                          .map((row) => (
                            <option key={row.recordId} value={row.recordId}>
                              {row.label}
                            </option>
                          ))}
                      </select>
                    </label>
                    <label style={labelStyle}>
                      Merge reason
                      <input
                        data-testid={entityMergeControlTestId("reason")}
                        style={inputStyle}
                        type="text"
                        value={mergeReason}
                        onChange={(event) => {
                          merge.commands.setReason(event.target.value);
                        }}
                      />
                    </label>
                    {loserEntity && mergePlan ? (
                      <div
                        data-testid={entityMergeControlTestId("plan")}
                        style={mergePlanStyle}
                      >
                        <p style={noticeTitleStyle}>
                          Survivor {selectedEntity.label} absorbs loser{" "}
                          {loserEntity.label}
                        </p>
                        <p style={bodyStyle}>
                          Survivor record {selectedEntity.recordId}
                          <br />
                          Loser record {loserEntity.recordId}
                        </p>
                        <ul style={flatListStyle}>
                          {mergePlan.identifierLines.map((line) => (
                            <li key={`${line.label}:${line.outcome}`}>
                              {line.label}: {line.outcome}
                            </li>
                          ))}
                          <li>
                            Aliases to copy:{" "}
                            {mergePlan.aliasesToCopy.length > 0
                              ? mergePlan.aliasesToCopy.join(", ")
                              : "none"}
                          </li>
                          <li>
                            Alias duplicate no-op:{" "}
                            {mergePlan.duplicateAliases.length > 0
                              ? mergePlan.duplicateAliases.join(", ")
                              : "none"}
                          </li>
                          <li>
                            Provenance-only values:{" "}
                            {mergePlan.provenanceOnlySummary}
                          </li>
                          <li>{mergePlan.dependencySummary}</li>
                        </ul>
                        <button
                          data-testid={entityMergeControlTestId("confirm")}
                          style={secondaryActionButtonStyle}
                          type="button"
                          onClick={() => {
                            setEntityActionMessage(null);
                            void merge.commands.confirm();
                          }}
                        >
                          Confirm merge
                        </button>
                      </div>
                    ) : (
                      <button
                        data-testid={entityMergeControlTestId("start")}
                        style={secondaryActionButtonStyle}
                        type="button"
                        onClick={() => {
                          setEntityActionMessage(null);
                          merge.commands.start();
                        }}
                      >
                        Start merge
                      </button>
                    )}
                  </section>
                ) : showRelationshipsPanel ? (
                  <section style={inspectorSectionStyle}>
                    <h3 style={sectionTitleStyle}>Merge</h3>
                    <p style={bodyStyle}>
                      Merge is available to reviewer or admin roles.
                    </p>
                  </section>
                ) : null}

                {timelinePreviewRows.length > 0 ? (
                  <section style={inspectorSectionStyle}>
                    <h3 style={sectionTitleStyle}>Dependent Timeline</h3>
                    <div style={timelinePreviewStackStyle}>
                      {timelinePreviewRows.map((row) => (
                        <article
                          key={row.recordId ?? row.key}
                          data-testid={
                            row.recordId === null
                              ? undefined
                              : timelinePreviewRowTestId(row.recordId)
                          }
                          style={timelinePreviewCardStyle}
                        >
                          <p style={noticeTitleStyle}>
                            {row.values.activitySynopsisText || "Untitled row"}
                          </p>
                          <div style={relationshipItemsWrapStyle}>
                            {row.collectionValues[
                              entityType === "host"
                                ? "hostRefs"
                                : "identityRefs"
                            ].map((item) => (
                              <WorkbookRelationshipChip
                                key={item.itemRef}
                                presentation={timelineRelationshipChipPresentation(
                                  { entityIndex, item },
                                )}
                              />
                            ))}
                          </div>
                        </article>
                      ))}
                    </div>
                  </section>
                ) : null}

                {presentedEntityActionMessage ? (
                  <div style={mergeMessageBlockStyle}>
                    <WorkbookInspectorPublicError
                      error={presentedEntityActionMessage}
                      testId={entityMergeControlTestId("message")}
                    />
                    {mergePreconditionDetails.length > 0 ? (
                      <ul
                        data-testid={entityMergePreconditionDetailsTestId(
                          entityType,
                          selectedEntity.recordId,
                        )}
                        style={flatListStyle}
                      >
                        {mergePreconditionDetails.map((line) => (
                          <li key={line.key}>
                            {line.label}: {line.value}
                          </li>
                        ))}
                      </ul>
                    ) : null}
                  </div>
                ) : null}
              </>
            ) : null}
          </WorkbookInspectorShell>
        ) : undefined
      }
      onRequestInspectorClose={() => {
        inspector.commands.close({ restoreFocus: true });
      }}
      primaryGrid={
        <GridViewport
          blockSizing="fill"
          style={gridShellStyle}
          testId={gridShellTestId(surface)}
        >
          <SemanticDataGrid
            ref={gridHandleRef}
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
            dataRows={entityGridRows}
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
          workbookFocusAnchor={entityFocus.snapshot.anchor}
        />
      }
      viewBar={
        <WorkbookViewBar
          addRowDisabled={!canCreateRows}
          chromeMode={chromeMode}
          queryControls={queryControls}
          savedViewControls={savedViewSelector}
          onAddRow={focusEntityDraft}
          onInspectorToggle={() => {
            inspectorContinuityTokenRef.current = entityFocus.port.capture();
            inspector.commands.open();
          }}
          surface={surface}
        />
      }
      viewSchemaId={surface}
    />
  );
}

const bodyStyle = {
  margin: 0,
  lineHeight: 1.5,
  color: "var(--ct-colors-ink-muted)",
};

const gridShellStyle = {
  ...workbookSurfaceGridShellStyle,
};

const inputStyle = {
  boxSizing: "border-box" as const,
  display: "block",
  minWidth: 0,
  width: "100%",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  padding: "0.65rem 0.75rem",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
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

const labelStyle = {
  display: "grid",
  gap: "0.4rem",
  fontSize: "0.95rem",
  color: "var(--ct-colors-ink-muted)",
};

const inspectorSectionStyle = {
  display: "grid",
  gap: "0.75rem",
  marginBottom: "1rem",
};

const inspectorControlStackStyle = {
  display: "grid",
  gap: "0.65rem",
};

const sectionTitleStyle = {
  margin: 0,
  fontSize: "1rem",
};

const sectionHeadingRowStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  gap: "0.5rem",
};

const reusableIdentifierSectionStyle = {
  ...inspectorSectionStyle,
  borderInlineStart: "var(--ct-border-strong)",
  paddingInlineStart: "0.75rem",
};

const readOnlyBadgeStyle = {
  border: "var(--ct-border-hairline)",
  borderRadius: "999px",
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.75rem",
  lineHeight: 1,
  padding: "0.2rem 0.45rem",
};

const mergeMessageBlockStyle = {
  display: "grid",
  gap: "0.4rem",
};

const relationshipItemsWrapStyle = {
  display: "flex",
  flexWrap: "wrap" as const,
  gap: "0.4rem",
  marginBottom: "0.55rem",
  maxWidth: "100%",
  minWidth: 0,
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

const aliasRemoveButtonStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  border: 0,
  background: "transparent",
  color: "inherit",
  cursor: "pointer",
  padding: 0,
};

const aliasAddRowStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(0, 1fr) auto",
  gap: "0.5rem",
};

const noticeTitleStyle = {
  margin: 0,
  fontSize: "0.95rem",
  fontWeight: 600,
};

const selectStyle = {
  ...inputStyle,
  appearance: "auto" as const,
};

const mergePlanStyle = {
  borderRadius: "var(--ct-rounded-lg)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  padding: "0.9rem",
  display: "grid",
  gap: "0.65rem",
};

const flatListStyle = {
  margin: 0,
  paddingLeft: "1.2rem",
  display: "grid",
  gap: "0.35rem",
};

const timelinePreviewStackStyle = {
  display: "grid",
  gap: "0.75rem",
};

const timelinePreviewCardStyle = {
  borderRadius: "var(--ct-rounded-lg)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  padding: "0.85rem",
  display: "grid",
  gap: "0.55rem",
};
