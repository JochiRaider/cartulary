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
  coordinationWorkflowTestId,
  dataTestIdSelector,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  genericEditActionSelectTestId,
  genericEditFieldSelectTestId,
  genericEditRecordSelectTestId,
  genericEditSubmitTestId,
  genericEditValueTestId,
  genericWorkbookTestId,
  gridActionsHeaderTestId,
  gridGroupRowTestId,
  gridShellTestId,
  workbookInlineDraftRowTestId,
  workbookInspectorCloseButtonTestId,
} from "@cartulary/ui-contracts";
import type {
  InspectorFeatureGroup,
  ViewContract,
} from "@cartulary/view-contracts";
import { X } from "lucide-react";
import {
  type CSSProperties,
  type ReactNode,
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
import { CoordinationWorkflowBindings } from "../features/coordination/CoordinationWorkflowBindings";
import { useEvidenceWorkbookBindings } from "../features/evidence/useEvidenceWorkbookBindings";
import { IndicatorInspectorWorkflow } from "../features/indicators/IndicatorInspectorWorkflow";
import {
  type IndicatorInspectorHandler,
  resolveIndicatorInspectorHandler,
} from "../features/indicators/indicatorInspectorHandlers";
import { useGenericPartyLinkWorkflow } from "../features/parties/useGenericPartyLinkWorkflow";
import { useGenericSurfaceMutationController } from "../hooks/useGenericSurfaceMutationController";
import { useOwnerReferenceOptions } from "../hooks/useOwnerReferenceOptions";
import { InspectorCreateRelatedWorkflow } from "../inspector/InspectorCreateRelatedWorkflow";
import { useInspectorCreateRelatedWorkflow } from "../inspector/useInspectorCreateRelatedWorkflow";
import { useWorkbookInspectorCoordinator } from "../inspector/useWorkbookInspectorCoordinator";
import { WorkbookInspectorRecordHistory } from "../inspector/WorkbookInspectorRecordHistory";
import type { WorkbookSurfaceLayoutOwner } from "../layout/useWorkbookLayoutFacade";
import {
  WorkbookSurfaceLayout,
  workbookSurfaceGridShellStyle,
  workbookSurfaceInspectorPanelStyle,
} from "../layout/WorkbookSurfaceLayout";
import { applyWorkbookLayoutToColumns } from "../layout/workbookColumnLayout";
import {
  buildGenericPatchChange,
  type GenericCollectionMode,
  genericCellLabelForField,
  genericCollectionItems,
  genericCollectionSupportsRemove,
  genericContractColumnWidth,
  genericCreateMinimumMessage,
  genericRowLabel,
  initialGenericCreateDraft,
  partyLinkPairsForContract,
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
import {
  inspectorPanelIsDeclared,
  selectInspectorConfig,
} from "../models/workbookInspectorModel";
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
import {
  type InspectorDisabledToken,
  WorkbookInspectorPanelSection,
} from "./WorkbookInspectorFeatureGroups";
import { WorkbookCellPresenceMarker } from "./WorkbookPresenceMarkers";
import {
  type WorkbookConflictActivation,
  WorkbookSurfaceStatusStrip,
} from "./WorkbookStatusStrip";
import { WorkbookViewBar } from "./WorkbookViewBar";

export type ContractWorkbookSurfaceProps = {
  readonly contract: ViewContract;
  readonly continuityResetKey: string;
  readonly currentUserId: string | null;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly incidentPort: WorkbookIncidentPort;
  readonly inspectorResetKey: string;
  readonly queryControls?: ReactNode | undefined;
  readonly savedViewSelector?: ReactNode | undefined;
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
  queryControls,
  savedViewSelector,
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
  const inspectorConfig = selectInspectorConfig(contract);
  const showDetailsPanel = inspectorPanelIsDeclared(inspectorConfig, "details");
  const showRelationshipsPanel = inspectorPanelIsDeclared(
    inspectorConfig,
    "relationships",
  );
  const showWorkflowPanel = inspectorPanelIsDeclared(
    inspectorConfig,
    "workflow",
  );
  const draftRowRecordId = `${surface}:draft-row`;
  const inspectorContinuityTokenRef = useRef<WorkbookContinuityToken | null>(
    null,
  );
  const continuityPortRef = useRef<WorkbookContinuityPort | null>(null);
  const editableFields = useMemo(
    () => contract.fields.filter((field) => field.writeKind !== "read_only"),
    [contract],
  );
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
  const [editFieldKey, setEditFieldKey] = useState("");
  const [editValue, setEditValue] = useState("");
  const [linkedNoteSourceRecordId, setLinkedNoteSourceRecordId] = useState("");
  const [indicatorInspectorHandler, setIndicatorInspectorHandler] =
    useState<IndicatorInspectorHandler | null>(null);
  const [editCollectionMode, setEditCollectionMode] =
    useState<GenericCollectionMode>("add");
  const { referenceLoadError, referenceOptions, refreshReferenceOptions } =
    useOwnerReferenceOptions({
      incidentPort,
      onIncidentAccessLost,
      referenceQueryBroker,
      viewSchemaId: contract.viewSchemaId,
    });
  const {
    beginMutation,
    clearMutationError,
    completeGenericMutation,
    markMutationConflict,
    markMutationSaved,
    mutationError,
    mutationState,
    rejectMutationFailure,
    setValidationError,
    submitPatchMutation,
  } = useGenericSurfaceMutationController({
    mutationCommands: mutationCommands.generic,
    mutationRuntime,
    onRefresh,
    refreshReferenceOptions,
    surfaceLabel: contract.title,
  });
  const sharedMutation = useWorkbookMutationRuntime(
    mutationRuntime,
    contract.viewSchemaId,
  );
  const collaboration = useWorkbookCollaborationCoordinator(
    collaborationProjection,
  );
  const presentedMutationState =
    mutationState === "Saved" ? sharedMutation.primaryLabel : mutationState;
  const inspectorSubjectRow =
    rows.find((row) => row.record_id === editRecordId) ?? null;
  const inspector = useWorkbookInspectorCoordinator({
    actionPorts: {
      clearLocalForm: () => {
        setEditValue("");
        setLinkedNoteSourceRecordId("");
        setEditCollectionMode("add");
        setPartyLinkExistingPartyId("");
        clearMutationError();
      },
      clearLifecycleState: () => {
        continuityPortRef.current?.clear();
      },
      clearSelection: () => {
        setEditRecordId("");
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
      inspectorSubjectRow === null
        ? null
        : {
            recordId: inspectorSubjectRow.record_id,
            rowVersion: inspectorSubjectRow.row_version,
            viewSchemaId: contract.viewSchemaId,
          },
  });
  const isInspectorOpen = inspector.snapshot.isOpen;
  const inspectorInvalidationKey = `${contract.viewSchemaId}:${inspector.snapshot.invalidationGeneration}`;
  const isNotesSurface = ownerBindings.includes("linked_note_create");
  const partyLinkPairs = useMemo(
    () => partyLinkPairsForContract(contract),
    [contract],
  );

  useEffect(() => {
    setCreateDraft((current) => {
      const defaults = initialGenericCreateDraft(contract, currentUserId);
      return { ...defaults, ...current };
    });
  }, [contract, currentUserId]);

  const ownerRecordActions = useEvidenceWorkbookBindings({
    mutationCommands: mutationCommands.evidence,
    mutation: { beginMutation, markMutationConflict, markMutationSaved },
    onRefresh,
    ownerBindings,
    resetKey: inspectorInvalidationKey,
  });

  const submitCreate = useCallback(async () => {
    if (!canCreateRows) return;
    if (
      !mutationCommands.generic.canCreateRecord({
        contract,
        draft: createDraft,
      })
    ) {
      setValidationError(genericCreateMinimumMessage(contract));
      return;
    }
    beginMutation();
    const result = await mutationCommands.generic.createRecord({
      contract,
      draft: createDraft,
      linkedNoteSourceRecordId:
        isNotesSurface && linkedNoteSourceRecordId !== ""
          ? linkedNoteSourceRecordId
          : "",
    });
    if (result.kind === "rejected") {
      rejectMutationFailure(result.failure);
      return;
    }
    setCreateDraft(initialGenericCreateDraft(contract, currentUserId));
    setLinkedNoteSourceRecordId("");
    await completeGenericMutation();
  }, [
    beginMutation,
    completeGenericMutation,
    contract,
    canCreateRows,
    createDraft,
    currentUserId,
    isNotesSurface,
    linkedNoteSourceRecordId,
    mutationCommands,
    rejectMutationFailure,
    setValidationError,
  ]);

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
    if (!ownerRecordActions.hasRecordActions && !canCreateRows) {
      return undefined;
    }
    return {
      headerTestId: gridActionsHeaderTestId(surface),
      label: "",
      width: ownerRecordActions.actionsWidth,
      renderDraftCell: () => (
        <button
          data-testid={
            isInspectorOpen
              ? undefined
              : genericCreateSubmitTestId(contract.viewSchemaId)
          }
          disabled={mutationState === "Syncing"}
          style={secondaryActionButtonStyle}
          type="button"
          onClick={() => {
            void submitCreate();
          }}
        >
          Commit
        </button>
      ),
      renderCell: ({ data: row }) => {
        return ownerRecordActions.renderRecordActions(row);
      },
    };
  }, [
    contract.viewSchemaId,
    isInspectorOpen,
    mutationState,
    ownerRecordActions,
    surface,
    submitCreate,
    canCreateRows,
  ]);
  const { row: selectedEditRow, field: selectedEditField } =
    selectWorkbookEditTarget({
      fieldKey: editFieldKey,
      fields: editableFields,
      getRecordId: (row: WorkbookQueryRow) => row.record_id,
      recordId: editRecordId,
      rows,
    });
  const selectedEditCollectionItems =
    selectedEditRow !== null && selectedEditField !== null
      ? genericCollectionItems(selectedEditRow, selectedEditField.fieldKey)
      : [];
  const createRelatedWorkflow = useInspectorCreateRelatedWorkflow({
    currentUserId,
    mutationCommands: mutationCommands.timeline.related,
    onCreated: refreshReferenceOptions,
    onMessage: (message) => {
      if (message === null) clearMutationError();
      else setValidationError(message);
    },
    selectedSubject:
      inspectorSubjectRow === null
        ? null
        : {
            cells: inspectorSubjectRow.cells,
            recordId: inspectorSubjectRow.record_id,
            rowVersion: inspectorSubjectRow.row_version,
          },
  });
  const genericInspectorDisabledTokens = useMemo(() => {
    const tokens = new Set<InspectorDisabledToken>();
    if (selectedEditRow === null) tokens.add("no_row_selected");
    else tokens.add("record_not_deleted");
    tokens.add("rollback_target_unavailable");
    tokens.add("pivot_target_unavailable");
    if (interactionMode.kind === "read_only") tokens.add("incident_closed");
    return tokens;
  }, [interactionMode.kind, selectedEditRow]);
  const executeInspectorRecordLifecycle = useCallback(
    async (featureGroup: InspectorFeatureGroup): Promise<boolean> => {
      if (
        featureGroup.routeBinding.kind !== "record_action" ||
        featureGroup.routeBinding.owner !== "record_delete_route"
      ) {
        return false;
      }
      if (inspectorSubjectRow === null) {
        setValidationError("Select a saved row before running this action.");
        return true;
      }
      beginMutation();
      const outcome = await mutationCommands.records.execute({
        action: "delete",
        baseRowVersion: inspectorSubjectRow.row_version,
        reason: `Deleted from the ${contract.title} inspector`,
        recordId: inspectorSubjectRow.record_id,
      });
      if (outcome.kind === "rejected") {
        rejectMutationFailure(outcome.failure);
        return true;
      }
      setEditRecordId("");
      inspector.commands.completeAction();
      await completeGenericMutation();
      return true;
    },
    [
      beginMutation,
      completeGenericMutation,
      contract.title,
      inspector.commands,
      inspectorSubjectRow,
      mutationCommands.records,
      rejectMutationFailure,
      setValidationError,
    ],
  );
  const handleInspectorFeatureAction = useCallback(
    (featureGroup: InspectorFeatureGroup) => {
      const indicatorHandler = resolveIndicatorInspectorHandler(
        contract.viewSchemaId,
        featureGroup,
      );
      setIndicatorInspectorHandler(indicatorHandler);
      if (indicatorHandler !== null) return;
      if (createRelatedWorkflow.commands.begin(featureGroup)) return;
      if (featureGroup.routeBinding.kind === "record_action") {
        void executeInspectorRecordLifecycle(featureGroup).then((handled) => {
          if (!handled) {
            setValidationError(
              `${featureGroup.label}: use the owner controls in this inspector section.`,
            );
          }
        });
        return;
      }
      if (featureGroup.routeBinding.kind === "record_patch") {
        if (inspectorSubjectRow !== null) {
          setEditRecordId(inspectorSubjectRow.record_id);
        }
        const actionField = contract.fields.find(
          (field) => field.writeAction === featureGroup.routeBinding.actionKey,
        );
        if (actionField !== undefined) {
          setEditFieldKey(actionField.fieldKey);
        }
        setValidationError(
          `${featureGroup.label}: the selected row edit controls are ready below.`,
        );
        return;
      }
      setValidationError(
        `${featureGroup.label}: use the owner controls in this inspector section.`,
      );
    },
    [
      contract.viewSchemaId,
      contract.fields,
      createRelatedWorkflow.commands,
      executeInspectorRecordLifecycle,
      inspectorSubjectRow,
      setValidationError,
    ],
  );

  useEffect(() => {
    if (selectedEditField?.writeKind !== "action_payload") {
      setEditCollectionMode("add");
    } else if (
      !genericCollectionSupportsRemove(selectedEditField.fieldKey) &&
      editCollectionMode === "remove"
    ) {
      setEditCollectionMode("add");
    }
  }, [editCollectionMode, selectedEditField]);

  useEffect(() => {
    if (selectedEditRow === null || selectedEditField === null) {
      setEditValue("");
      return;
    }
    if (selectedEditField.writeKind === "action_payload") {
      setEditValue("");
      return;
    }
    const value = selectedEditRow.cells[selectedEditField.fieldKey]?.value;
    setEditValue(value === null || value === undefined ? "" : String(value));
  }, [selectedEditField, selectedEditRow]);

  const submitEdit = async () => {
    if (selectedEditRow === null || selectedEditField === null) {
      setValidationError("invalid_mutation_payload");
      return;
    }
    const change = buildGenericPatchChange(
      selectedEditField,
      editValue,
      editCollectionMode,
      contract.viewSchemaId,
    );
    if (change === null) {
      setValidationError(
        "Provide a value, or leave clearable fields empty to clear them.",
      );
      return;
    }
    const payload = await submitPatchMutation({
      baseRowVersion: selectedEditRow.row_version,
      changes: [change],
      purpose: "generic-patch",
      recordId: selectedEditRow.record_id,
      viewSchemaId: contract.viewSchemaId,
    });
    if (payload === null) {
      return;
    }
    setEditValue("");
    await completeGenericMutation();
  };

  const submitPartyLinkPatch = async (
    changes: Array<Record<string, unknown>>,
    txnPrefix: string,
  ) => {
    if (selectedEditRow === null) {
      setValidationError("Select a row before changing a party link.");
      return false;
    }
    const payload = await submitPatchMutation({
      baseRowVersion: selectedEditRow.row_version,
      changes,
      purpose: txnPrefix,
      recordId: selectedEditRow.record_id,
      viewSchemaId: contract.viewSchemaId,
    });
    if (payload === null) {
      return false;
    }
    await completeGenericMutation();
    return true;
  };
  const {
    clearPartyBoth,
    clearPartyLink,
    clearPartyText,
    createPartyFromText,
    linkExistingParty,
    partialCompletionMessage,
    partyLinkExistingPartyId,
    retryCreatedPartyLink,
    selectedPartyLinkPair,
    setPartyLinkExistingPartyId,
    setPartyLinkPairKey,
  } = useGenericPartyLinkWorkflow({
    mutation: { beginMutation, rejectMutationFailure, setValidationError },
    mutationCommands: mutationCommands.generic,
    originViewSchemaId: contract.viewSchemaId,
    partyLinkPairs,
    resetKey: inspectorInvalidationKey,
    selectedRow: selectedEditRow,
    submitLinkPatch: submitPartyLinkPatch,
  });

  const focusDraftRow = useCallback(() => {
    const firstWritableField = createFields[0];
    if (!firstWritableField || !canCreateRows) {
      return;
    }
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

  return (
    <WorkbookSurfaceLayout
      chromeMode={chromeMode}
      inspector={
        isInspectorOpen ? (
          <section style={genericMutationPanelStyle}>
            <div style={inspectorTitleRowStyle}>
              <div>
                <p style={eyebrowStyle}>Inspector</p>
                <h2 style={inspectorTitleStyle}>Workbook actions</h2>
              </div>
              <button
                aria-label="Close inspector"
                data-testid={workbookInspectorCloseButtonTestId(surface)}
                style={inspectorCloseButtonStyle}
                type="button"
                onClick={() => {
                  inspector.commands.close({ restoreFocus: true });
                }}
              >
                <X aria-hidden="true" size={16} />
              </button>
            </div>
            {inspectorConfig.panels.map((panel) => (
              <WorkbookInspectorPanelSection
                config={inspectorConfig}
                currentIncidentRole={currentIncidentRole}
                disabledTokens={genericInspectorDisabledTokens}
                key={panel.panelId}
                panelId={panel.panelId}
                subjectRecordId={selectedEditRow?.record_id ?? null}
                subjectRowVersion={selectedEditRow?.row_version ?? null}
                onFeatureAction={handleInspectorFeatureAction}
              >
                {indicatorInspectorHandler?.panelId === panel.panelId &&
                selectedEditRow !== null ? (
                  <IndicatorInspectorWorkflow
                    action={indicatorInspectorHandler.action}
                    indicatorRecordId={selectedEditRow.record_id}
                    port={mutationCommands.indicators}
                    rowVersion={selectedEditRow.row_version}
                    onMutationCommitted={onRefresh}
                  />
                ) : null}
                {createRelatedWorkflow.snapshot.workflow?.featureGroup
                  .panelId === panel.panelId ? (
                  <InspectorCreateRelatedWorkflow
                    referenceOptions={referenceOptions}
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
                    commands={mutationCommands.records}
                    subject={
                      inspectorSubjectRow === null
                        ? null
                        : {
                            recordId: inspectorSubjectRow.record_id,
                            rowVersion: inspectorSubjectRow.row_version,
                          }
                    }
                    onMessage={setValidationError}
                    onRefresh={onRefresh}
                  />
                ) : null}
                {panel.panelId === "evidence" && inspectorSubjectRow !== null
                  ? ownerRecordActions.renderRecordActions(inspectorSubjectRow)
                  : null}
              </WorkbookInspectorPanelSection>
            ))}
            {isNotesSurface ? (
              <label
                htmlFor={genericWorkbookTestId("note-source-record")}
                style={labelStyle}
              >
                Linked source for draft row
                <select
                  data-testid={genericWorkbookTestId("note-source-record")}
                  id={genericWorkbookTestId("note-source-record")}
                  style={selectStyle}
                  value={linkedNoteSourceRecordId}
                  onChange={(event) => {
                    setLinkedNoteSourceRecordId(event.target.value);
                  }}
                >
                  <option value="">None</option>
                  {referenceOptions.noteSourceRecords.map((option) => (
                    <option key={option.recordId} value={option.recordId}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </label>
            ) : null}

            {showWorkflowPanel &&
            canCreateRows &&
            draftInspectorFields.length > 0 ? (
              <div style={genericDraftInspectorFieldsStyle}>
                {draftInspectorFields.map((field) => {
                  const controlId = `generic-create-inspector-${field.fieldKey}`;
                  return (
                    <label
                      htmlFor={controlId}
                      key={field.fieldKey}
                      style={labelStyle}
                    >
                      {field.label}
                      <GenericMutationControl
                        collectionMode="add"
                        field={field}
                        id={controlId}
                        referenceOptions={referenceOptions}
                        testId={genericCreateFieldTestId(field.fieldKey)}
                        value={createDraft[field.fieldKey] ?? ""}
                        onChange={(value) => {
                          setCreateDraft((current) => ({
                            ...current,
                            [field.fieldKey]: value,
                          }));
                        }}
                      />
                    </label>
                  );
                })}
              </div>
            ) : null}

            {showWorkflowPanel && canCreateRows ? (
              <button
                data-testid={genericCreateSubmitTestId(contract.viewSchemaId)}
                disabled={mutationState === "Syncing"}
                style={secondaryActionButtonStyle}
                type="button"
                onClick={() => {
                  void submitCreate();
                }}
              >
                Commit draft row
              </button>
            ) : null}

            {showDetailsPanel &&
            rows.length > 0 &&
            selectedEditField !== null ? (
              <div style={genericEditRowStyle}>
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
                    <option key={row.record_id} value={row.record_id}>
                      {genericRowLabel(contract, row)}
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
                  {editableFields.map((field) => (
                    <option key={field.fieldKey} value={field.fieldKey}>
                      {field.label}
                    </option>
                  ))}
                </select>
                {selectedEditField.writeKind === "action_payload" &&
                genericCollectionSupportsRemove(selectedEditField.fieldKey) ? (
                  <select
                    aria-label="Collection edit action"
                    data-testid={genericEditActionSelectTestId(
                      contract.viewSchemaId,
                    )}
                    style={selectStyle}
                    value={editCollectionMode}
                    onChange={(event) => {
                      setEditCollectionMode(
                        event.target.value === "remove" ? "remove" : "add",
                      );
                      setEditValue("");
                    }}
                  >
                    <option value="add">Add</option>
                    <option value="remove">Remove</option>
                  </select>
                ) : null}
                <GenericMutationControl
                  collectionItems={selectedEditCollectionItems}
                  collectionMode={editCollectionMode}
                  field={selectedEditField}
                  referenceOptions={referenceOptions}
                  testId={genericEditValueTestId(contract.viewSchemaId)}
                  value={editValue}
                  onChange={setEditValue}
                />
                <button
                  data-testid={genericEditSubmitTestId(contract.viewSchemaId)}
                  disabled={mutationState === "Syncing"}
                  style={actionButtonStyle}
                  type="button"
                  onClick={() => {
                    void submitEdit();
                  }}
                >
                  Update
                </button>
              </div>
            ) : null}

            {showRelationshipsPanel &&
            partyLinkPairs.length > 0 &&
            selectedEditRow !== null ? (
              <div style={genericEditRowStyle}>
                <select
                  aria-label="Party link field"
                  data-testid={coordinationWorkflowTestId("party-pair")}
                  style={selectStyle}
                  value={selectedPartyLinkPair?.key ?? ""}
                  onChange={(event) => {
                    setPartyLinkPairKey(event.target.value);
                  }}
                >
                  {partyLinkPairs.map((pair) => (
                    <option key={pair.key} value={pair.key}>
                      {pair.label}
                    </option>
                  ))}
                </select>
                <select
                  aria-label="Existing party"
                  data-testid={coordinationWorkflowTestId("party-existing")}
                  style={selectStyle}
                  value={partyLinkExistingPartyId}
                  onChange={(event) => {
                    setPartyLinkExistingPartyId(event.target.value);
                  }}
                >
                  <option value="">Party</option>
                  {referenceOptions.parties.map((option) => (
                    <option key={option.recordId} value={option.recordId}>
                      {option.label}
                    </option>
                  ))}
                </select>
                <button
                  data-testid={coordinationWorkflowTestId(
                    "party-create-from-text",
                  )}
                  disabled={mutationState === "Syncing"}
                  style={secondaryActionButtonStyle}
                  type="button"
                  onClick={() => {
                    void createPartyFromText();
                  }}
                >
                  Create party from text
                </button>
                <button
                  data-testid={coordinationWorkflowTestId(
                    "party-link-existing",
                  )}
                  disabled={mutationState === "Syncing"}
                  style={secondaryActionButtonStyle}
                  type="button"
                  onClick={() => {
                    void linkExistingParty();
                  }}
                >
                  Link existing party
                </button>
                <button
                  data-testid={coordinationWorkflowTestId("party-clear-link")}
                  disabled={mutationState === "Syncing"}
                  style={secondaryActionButtonStyle}
                  type="button"
                  onClick={() => {
                    void clearPartyLink();
                  }}
                >
                  Clear party link
                </button>
                <button
                  data-testid={coordinationWorkflowTestId("party-clear-text")}
                  disabled={mutationState === "Syncing"}
                  style={secondaryActionButtonStyle}
                  type="button"
                  onClick={() => {
                    void clearPartyText();
                  }}
                >
                  Clear party text
                </button>
                <button
                  data-testid={coordinationWorkflowTestId("party-clear-both")}
                  disabled={mutationState === "Syncing"}
                  style={secondaryActionButtonStyle}
                  type="button"
                  onClick={() => {
                    void clearPartyBoth();
                  }}
                >
                  Clear both
                </button>
                {partialCompletionMessage === null ? null : (
                  <div>
                    <p
                      data-testid={coordinationWorkflowTestId(
                        "party-partial-completion",
                      )}
                      role="status"
                    >
                      {partialCompletionMessage}
                    </p>
                    <button
                      data-testid={coordinationWorkflowTestId(
                        "party-retry-created-link",
                      )}
                      disabled={mutationState === "Syncing"}
                      style={secondaryActionButtonStyle}
                      type="button"
                      onClick={() => {
                        void retryCreatedPartyLink();
                      }}
                    >
                      Retry link to created party
                    </button>
                  </div>
                )}
              </div>
            ) : null}

            {showWorkflowPanel ? (
              <CoordinationWorkflowBindings
                contract={contract}
                disabled={mutationState === "Syncing"}
                mutation={{
                  beginMutation,
                  completeGenericMutation,
                  rejectMutationFailure,
                  setValidationError,
                }}
                mutationCommands={mutationCommands.coordination}
                ownerBindings={ownerBindings}
                referenceOptions={referenceOptions}
                resetKey={inspectorInvalidationKey}
                rows={rows}
              />
            ) : null}

            {referenceLoadError ? (
              <p
                data-testid={genericWorkbookTestId("reference-load-error")}
                style={bodyStyle}
              >
                {referenceLoadError}
              </p>
            ) : null}

            {mutationError ? (
              <p style={genericErrorTextStyle}>{mutationError}</p>
            ) : null}
          </section>
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
          mutationError={mutationError ?? sharedMutation.secondaryMessage}
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
          queryControls={queryControls}
          savedViewControls={savedViewSelector}
          onAddRow={focusDraftRow}
          onInspectorToggle={() => {
            inspectorContinuityTokenRef.current = genericFocus.port.capture();
            inspector.commands.open();
          }}
          surface={surface}
        />
      }
      viewSchemaId={surface}
      workAreaOverlays={ownerRecordActions.overlay}
    />
  );
}

const eyebrowStyle = {
  margin: 0,
  fontSize: "0.78rem",
  letterSpacing: "0.12em",
  textTransform: "uppercase" as const,
  color: "var(--ct-colors-accent)",
};

const bodyStyle = {
  margin: 0,
  lineHeight: 1.5,
  color: "var(--ct-colors-ink-muted)",
};

const gridShellStyle = {
  ...workbookSurfaceGridShellStyle,
} satisfies CSSProperties;

const genericMutationPanelStyle = {
  ...workbookSurfaceInspectorPanelStyle,
  display: "grid",
  alignContent: "start",
  gap: "0.75rem",
  background: "var(--ct-colors-surface-2)",
};

const genericEditRowStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(0, 1fr)",
  gap: "0.6rem",
  alignItems: "stretch",
};

const genericDraftInspectorFieldsStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(12rem, 1fr))",
  gap: "0.75rem",
  alignItems: "end",
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

const draftCellPlaceholderStyle = {
  color: "var(--ct-colors-ink-subtle)",
};

const genericErrorTextStyle = {
  margin: 0,
  color: "var(--ct-colors-semantic-conflict)",
  fontWeight: 700,
};

const labelStyle = {
  display: "grid",
  gap: "0.4rem",
  fontSize: "0.95rem",
  color: "var(--ct-colors-ink-muted)",
};

const inspectorTitleRowStyle = {
  display: "flex",
  alignItems: "start",
  justifyContent: "space-between",
  gap: "1rem",
};

const inspectorCloseButtonStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  width: "1.9rem",
  height: "1.9rem",
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  cursor: "pointer",
};

const inspectorTitleStyle = {
  margin: 0,
  fontSize: "1.25rem",
};

const selectStyle = {
  ...inputStyle,
  appearance: "auto" as const,
};
