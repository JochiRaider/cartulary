import {
  type GridActionsColumn,
  type GridCellPasteIntent,
  type GridColumn,
  type GridDataRow,
  type GridDensity,
  type GridDraftRow,
  type GridEditCommitOutcome,
  type GridGroupingDescriptor,
  type GridHandle,
  type GridInteractionMode,
  GridViewport,
  SemanticDataGrid,
} from "@cartulary/grid-adapter";
import {
  dataTestIdSelector,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  genericEditActionSelectTestId,
  genericEditFieldSelectTestId,
  genericEditRecordSelectTestId,
  genericEditSubmitTestId,
  genericEditValueTestId,
  gridActionsHeaderTestId,
  gridGroupRowTestId,
  gridShellTestId,
  type WorkbookSurface,
  workbookInlineDraftRowTestId,
  workbookInspectorCloseButtonTestId,
} from "@cartulary/ui-contracts";
import type { ViewContract } from "@cartulary/view-contracts";
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
import { apiPath, clientTxnID } from "../../services/browserApi";
import {
  fetchWorkbookJSON,
  parseErrorMessage,
  readEnvelope,
} from "../../services/workbookApi";
import { CoordinationWorkflowBindings } from "../features/coordination/CoordinationWorkflowBindings";
import { useEvidenceWorkbookBindings } from "../features/evidence/useEvidenceWorkbookBindings";
import {
  useGenericSurfaceMutationController,
  type GenericViewMutationEnvelope as ViewMutationEnvelope,
} from "../hooks/useGenericSurfaceMutationController";
import { useOwnerReferenceOptions } from "../hooks/useOwnerReferenceOptions";
import {
  buildGenericCreatePayload,
  buildGenericPatchChange,
  extractEmailFromPartyText,
  type GenericCollectionMode,
  genericCellLabelForField,
  genericCollectionItems,
  genericCollectionSupportsRemove,
  genericContractColumnWidth,
  genericCreateMinimumMessage,
  genericRowLabel,
  initialGenericCreateDraft,
  normalizeGenericTextValue,
  partyLinkPairsForContract,
  selectWorkbookEditTarget,
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
import {
  applyWorkbookLayoutToColumns,
  type WorkbookResolvedLayoutState,
} from "../models/workbookLayout";
import type { WorkbookQueryState } from "../models/workbookQuery";
import { requireWorkbookSurfaceRegistration } from "../models/workbookSurfaceRegistration";
import { partiesViewSchemaId } from "../models/workbookSurfaceRegistry";
import type { EntityApiRow } from "../timeline/models/workbookTimelineModel";
import { workbookClipboardPasteContract } from "../utils/workbookClipboard";
import {
  FocusableWorkbookCell,
  useWorkbookGridFocus,
} from "../utils/workbookGridFocus";
import { GenericMutationControl } from "./GenericMutationControl";
import { workbookGridEditorAdapter } from "./WorkbookGridEditorControl";
import {
  type InspectorDisabledToken,
  WorkbookInspectorPanelSection,
} from "./WorkbookInspectorFeatureGroups";
import { WorkbookSheetToolbar } from "./WorkbookSheetToolbar";
import { WorkbookSurfaceStatusStrip } from "./WorkbookStatusStrip";
import {
  WorkbookSurfaceFrame,
  workbookSurfaceGridShellStyle,
  workbookSurfaceInspectorPanelStyle,
} from "./WorkbookSurfaceFrame";

export type ContractWorkbookSurfaceProps = {
  readonly apiBase?: string | undefined;
  readonly authorizationEpoch: string;
  readonly contract: ViewContract;
  readonly currentUserId: string | null;
  readonly density: GridDensity;
  readonly inspectorResetKey: string;
  readonly savedViewSelector?: ReactNode | undefined;
  readonly incidentId: string;
  readonly loadState: WorkbookQueryLoadState;
  readonly interactionMode: GridInteractionMode;
  readonly onClearFilters: () => void;
  readonly onRefresh: () => Promise<void> | void;
  readonly layoutState: WorkbookResolvedLayoutState;
  readonly onColumnReorder: (
    sourceFieldKey: string,
    targetFieldKey: string,
  ) => void;
  readonly onColumnWidthChange: (fieldKey: string, width: number) => void;
  readonly onSortChange: (sort: WorkbookQueryState["sort"]) => void;
  readonly queryState: WorkbookQueryState;
  readonly rows: EntityApiRow[];
};

export function ContractWorkbookSurface({
  apiBase,
  authorizationEpoch,
  contract,
  currentUserId,
  density,
  inspectorResetKey,
  savedViewSelector,
  incidentId,
  loadState,
  interactionMode,
  onClearFilters,
  layoutState,
  onColumnReorder,
  onColumnWidthChange,
  onRefresh,
  onSortChange,
  queryState,
  rows,
}: ContractWorkbookSurfaceProps) {
  const surface = contract.viewSchemaId as WorkbookSurface;
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
  const [isInspectorOpen, setIsInspectorOpen] = useState(false);
  const writableFields = useMemo(
    () => contract.fields.filter((field) => field.writeKind !== "read_only"),
    [contract],
  );
  const [createDraft, setCreateDraft] = useState<Record<string, string>>(() =>
    initialGenericCreateDraft(contract, currentUserId),
  );
  const [editRecordId, setEditRecordId] = useState("");
  const [editFieldKey, setEditFieldKey] = useState("");
  const [editValue, setEditValue] = useState("");
  const [linkedNoteSourceRecordId, setLinkedNoteSourceRecordId] = useState("");
  const [editCollectionMode, setEditCollectionMode] =
    useState<GenericCollectionMode>("add");
  const [partyLinkPairKey, setPartyLinkPairKey] = useState("");
  const [partyLinkExistingPartyId, setPartyLinkExistingPartyId] = useState("");
  const { referenceLoadError, referenceOptions, refreshReferenceOptions } =
    useOwnerReferenceOptions({
      apiBase,
      authorizationEpoch,
      incidentId,
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
    rejectMutationPayload,
    setValidationError,
    submitPatchMutation,
  } = useGenericSurfaceMutationController({
    apiBase,
    onRefresh,
    refreshReferenceOptions,
  });
  const isNotesSurface = ownerBindings.includes("linked_note_create");
  const partyLinkPairs = useMemo(
    () => partyLinkPairsForContract(contract),
    [contract],
  );

  useEffect(() => {
    if (inspectorResetKey === "") {
      return;
    }
    setIsInspectorOpen(false);
    setEditRecordId("");
    setEditFieldKey("");
    setEditValue("");
    setLinkedNoteSourceRecordId("");
    setEditCollectionMode("add");
    setPartyLinkExistingPartyId("");
    clearMutationError();
  }, [clearMutationError, inspectorResetKey]);

  useEffect(() => {
    setCreateDraft((current) => {
      const defaults = initialGenericCreateDraft(contract, currentUserId);
      return { ...defaults, ...current };
    });
  }, [contract, currentUserId]);

  useEffect(() => {
    setPartyLinkPairKey((current) => {
      if (partyLinkPairs.some((pair) => pair.key === current)) {
        return current;
      }
      return partyLinkPairs[0]?.key ?? "";
    });
  }, [partyLinkPairs]);

  const ownerRecordActions = useEvidenceWorkbookBindings({
    apiBase,
    incidentId,
    mutation: { beginMutation, markMutationConflict, markMutationSaved },
    onRefresh,
    ownerBindings,
    resetKey: inspectorResetKey,
  });

  const submitCreate = useCallback(async () => {
    if (interactionMode.kind === "read_only") return;
    const payload = buildGenericCreatePayload(
      contract,
      createDraft,
      clientTxnID(`generic-create-${contract.viewSchemaId}`),
    );
    if (payload === null) {
      setValidationError(genericCreateMinimumMessage(contract.viewSchemaId));
      return;
    }
    beginMutation();
    const createPath =
      isNotesSurface && linkedNoteSourceRecordId !== ""
        ? `/api/v1/records/${linkedNoteSourceRecordId}/linked-notes`
        : `/api/v1/incidents/${incidentId}/views/${contract.viewSchemaId}/rows`;
    const result = await fetchWorkbookJSON<ViewMutationEnvelope>(
      apiPath(apiBase, createPath),
      { method: "POST", body: JSON.stringify(payload) },
    );
    if (!result.ok) {
      rejectMutationPayload(result.payload);
      return;
    }
    setCreateDraft(initialGenericCreateDraft(contract, currentUserId));
    setLinkedNoteSourceRecordId("");
    await completeGenericMutation<ViewMutationEnvelope>(result.payload);
  }, [
    apiBase,
    beginMutation,
    completeGenericMutation,
    contract,
    createDraft,
    currentUserId,
    incidentId,
    isNotesSurface,
    interactionMode.kind,
    linkedNoteSourceRecordId,
    rejectMutationPayload,
    setValidationError,
  ]);

  const anchorColumns = useMemo<readonly GridColumn<EntityApiRow>[]>(
    () =>
      workbookContractColumns<EntityApiRow>({
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
      beginMutation();
      const result = await fetchWorkbookJSON<ViewMutationEnvelope>(
        apiPath(apiBase, `/api/v1/records/${target.recordId}`),
        {
          method: "PATCH",
          body: JSON.stringify({
            view_schema_id: contract.viewSchemaId,
            base_row_version: target.baseRowVersion,
            client_txn_id: clientTxnID(`grid-edit-${contract.viewSchemaId}`),
            changes: [change],
          }),
        },
      );
      if (!result.ok) {
        rejectMutationPayload(result.payload);
        const message = parseErrorMessage(result.payload);
        if (result.status === 409) return { kind: "conflict", message };
        if (result.status === 404) return { kind: "stale_target", message };
        if (result.status === 400) {
          return { kind: "validation_error", message };
        }
        return { kind: "rejected_mutation", message };
      }
      await completeGenericMutation<ViewMutationEnvelope>(result.payload);
      return { kind: "accepted" };
    },
    [
      apiBase,
      beginMutation,
      completeGenericMutation,
      contract,
      rejectMutationPayload,
      rows,
    ],
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
    return writableFields.filter((field) => !gridFieldKeys.has(field.fieldKey));
  }, [visibleAnchorColumns, writableFields]);
  const draftApiRow = useMemo<EntityApiRow>(
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
  const gridRecordRows = useMemo<readonly GridDataRow<EntityApiRow>[]>(
    () =>
      workbookGridRows({
        getRecordId: (row: EntityApiRow) => row.record_id,
        getRowVersion: (row: EntityApiRow) => row.row_version,
        rows,
        surface,
      }),
    [rows, surface],
  );
  const gridDraftRow = useMemo<GridDraftRow<EntityApiRow> | undefined>(
    () =>
      writableFields.length === 0
        ? undefined
        : {
            kind: "draft",
            data: draftApiRow,
            gutterContent: "+",
            gutterLabel: "Draft row",
            testId: workbookInlineDraftRowTestId(surface),
          },
    [draftApiRow, surface, writableFields.length],
  );
  const grouping = useMemo<GridGroupingDescriptor<EntityApiRow> | null>(() => {
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
  const genericFocus = useWorkbookGridFocus({
    columns: visibleAnchorColumns,
    gridHandleRef,
    surface,
  });
  const columns: readonly GridColumn<EntityApiRow>[] = visibleAnchorColumns.map(
    (column) => {
      const field = contract.fieldMap[column.fieldKey];
      return {
        ...column,
        contractWritable: field?.gridEditable === true,
        getClipboardValue: (row: EntityApiRow) => {
          const value = row.cells[column.fieldKey]?.value;
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
                readValue: (row: EntityApiRow) =>
                  row.cells[field.fieldKey]?.value,
                referenceOptions,
              })
            : undefined,
        renderDraftCell: () => {
          const writableField =
            writableFields.find(
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
            <FocusableWorkbookCell
              fieldKey={column.fieldKey}
              focus={genericFocus}
              recordId={row.record_id}
            >
              {genericCellLabelForField(
                surface,
                column.fieldKey,
                row.cells[column.fieldKey]?.value,
              )}
            </FocusableWorkbookCell>
          );
        },
      };
    },
  );
  const rowActionsColumn = useMemo<
    GridActionsColumn<EntityApiRow> | undefined
  >(() => {
    if (!ownerRecordActions.hasRecordActions && writableFields.length === 0) {
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
    writableFields.length,
  ]);
  const { row: selectedEditRow, field: selectedEditField } =
    selectWorkbookEditTarget({
      fieldKey: editFieldKey,
      fields: writableFields,
      getRecordId: (row: EntityApiRow) => row.record_id,
      recordId: editRecordId,
      rows,
    });
  const selectedPartyLinkPair =
    partyLinkPairs.find((pair) => pair.key === partyLinkPairKey) ??
    partyLinkPairs[0] ??
    null;
  const selectedEditCollectionItems =
    selectedEditRow !== null && selectedEditField !== null
      ? genericCollectionItems(selectedEditRow, selectedEditField.fieldKey)
      : [];
  const genericInspectorDisabledTokens = useMemo(
    () =>
      new Set<InspectorDisabledToken>(
        selectedEditRow === null ? ["no_row_selected"] : [],
      ),
    [selectedEditRow],
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
      clientTxnId: clientTxnID(`generic-patch-${contract.viewSchemaId}`),
      recordId: selectedEditRow.record_id,
      viewSchemaId: contract.viewSchemaId,
    });
    if (payload === null) {
      return;
    }
    setEditValue("");
    await completeGenericMutation<ViewMutationEnvelope>(payload);
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
      clientTxnId: clientTxnID(`${txnPrefix}-${contract.viewSchemaId}`),
      recordId: selectedEditRow.record_id,
      viewSchemaId: contract.viewSchemaId,
    });
    if (payload === null) {
      return false;
    }
    await completeGenericMutation<ViewMutationEnvelope>(payload);
    return true;
  };

  const createPartyFromText = async () => {
    if (selectedEditRow === null || selectedPartyLinkPair === null) {
      setValidationError("Select a row and party field first.");
      return;
    }
    const rawText = normalizeGenericTextValue(
      String(
        selectedEditRow.cells[selectedPartyLinkPair.textFieldKey]?.value ?? "",
      ),
    );
    if (rawText === "") {
      setValidationError("Party text is empty.");
      return;
    }
    beginMutation();
    const createPayload: Record<string, unknown> = {
      client_txn_id: clientTxnID(`party-from-text-${contract.viewSchemaId}`),
      "party.display_name": rawText,
      "party.party_kind": "person",
    };
    const email = extractEmailFromPartyText(rawText);
    if (email !== null) {
      createPayload["party.primary_email"] = email;
    }
    const createResult = await fetchWorkbookJSON<ViewMutationEnvelope>(
      apiPath(
        apiBase,
        `/api/v1/incidents/${incidentId}/views/${partiesViewSchemaId}/rows`,
      ),
      { method: "POST", body: JSON.stringify(createPayload) },
    );
    if (!createResult.ok) {
      rejectMutationPayload(createResult.payload);
      return;
    }
    const partyID = readEnvelope<ViewMutationEnvelope>(createResult.payload)
      .data.row.record_id;
    await submitPartyLinkPatch(
      [{ field_key: selectedPartyLinkPair.refFieldKey, value: partyID }],
      "party-link-created",
    );
  };

  const linkExistingParty = async () => {
    if (selectedPartyLinkPair === null || partyLinkExistingPartyId === "") {
      setValidationError("Select an existing party.");
      return;
    }
    await submitPartyLinkPatch(
      [
        {
          field_key: selectedPartyLinkPair.refFieldKey,
          value: partyLinkExistingPartyId,
        },
      ],
      "party-link-existing",
    );
  };

  const clearPartyLink = async () => {
    if (selectedPartyLinkPair === null) {
      setValidationError("Select a party field first.");
      return;
    }
    await submitPartyLinkPatch(
      [{ field_key: selectedPartyLinkPair.refFieldKey, value: null }],
      "party-clear-link",
    );
  };

  const clearPartyText = async () => {
    if (selectedPartyLinkPair === null) {
      setValidationError("Select a party field first.");
      return;
    }
    await submitPartyLinkPatch(
      [{ field_key: selectedPartyLinkPair.textFieldKey, value: null }],
      "party-clear-text",
    );
  };

  const clearPartyBoth = async () => {
    if (selectedPartyLinkPair === null) {
      setValidationError("Select a party field first.");
      return;
    }
    await submitPartyLinkPatch(
      [
        { field_key: selectedPartyLinkPair.textFieldKey, value: null },
        { field_key: selectedPartyLinkPair.refFieldKey, value: null },
      ],
      "party-clear-both",
    );
  };

  const focusDraftRow = useCallback(() => {
    const firstWritableField = writableFields[0];
    if (!firstWritableField || interactionMode.kind === "read_only") {
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
  }, [interactionMode.kind, writableFields]);
  const dataState = workbookGridDataState({
    emptyAction:
      writableFields.length > 0 && interactionMode.kind === "editable"
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
    <WorkbookSurfaceFrame
      inspector={
        isInspectorOpen && writableFields.length > 0 ? (
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
                  setIsInspectorOpen(false);
                }}
              >
                <X aria-hidden="true" size={16} />
              </button>
            </div>
            {inspectorConfig.panels.map((panel) => (
              <WorkbookInspectorPanelSection
                config={inspectorConfig}
                disabledTokens={genericInspectorDisabledTokens}
                key={panel.panelId}
                panelId={panel.panelId}
              />
            ))}
            {isNotesSurface ? (
              <label
                htmlFor="generic-create-note-source-record"
                style={labelStyle}
              >
                Linked source for draft row
                <select
                  data-testid="generic-create-note-source-record"
                  id="generic-create-note-source-record"
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

            {showWorkflowPanel && draftInspectorFields.length > 0 ? (
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

            {showWorkflowPanel ? (
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
                  {writableFields.map((field) => (
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
                  data-testid="party-link-pair"
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
                  data-testid="party-link-existing-party"
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
                  data-testid="party-link-create-from-text"
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
                  data-testid="party-link-link-existing"
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
                  data-testid="party-link-clear-link"
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
                  data-testid="party-link-clear-text"
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
                  data-testid="party-link-clear-both"
                  disabled={mutationState === "Syncing"}
                  style={secondaryActionButtonStyle}
                  type="button"
                  onClick={() => {
                    void clearPartyBoth();
                  }}
                >
                  Clear both
                </button>
              </div>
            ) : null}

            {showWorkflowPanel ? (
              <CoordinationWorkflowBindings
                apiBase={apiBase}
                contract={contract}
                disabled={mutationState === "Syncing"}
                mutation={{
                  beginMutation,
                  completeGenericMutation,
                  rejectMutationPayload,
                  setValidationError,
                }}
                ownerBindings={ownerBindings}
                referenceOptions={referenceOptions}
                resetKey={inspectorResetKey}
                rows={rows}
              />
            ) : null}

            {referenceLoadError ? (
              <p data-testid="generic-reference-load-error" style={bodyStyle}>
                {referenceLoadError}
              </p>
            ) : null}

            {mutationError ? (
              <p style={genericErrorTextStyle}>{mutationError}</p>
            ) : null}
          </section>
        ) : undefined
      }
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
            onActiveCellChange={(anchor) =>
              genericFocus.update(
                anchor?.rowIdentity.kind === "core_record"
                  ? anchor.rowIdentity.recordId
                  : null,
                anchor?.fieldKey ?? "",
              )
            }
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
          mutationError={mutationError}
          mutationState={mutationState}
          workbookFocusAnchor={genericFocus.anchor}
        />
      }
      viewBar={
        <WorkbookSheetToolbar
          addRowDisabled={
            writableFields.length === 0 || interactionMode.kind === "read_only"
          }
          leading={savedViewSelector}
          onAddRow={focusDraftRow}
          onInspectorToggle={() => {
            setIsInspectorOpen(true);
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
