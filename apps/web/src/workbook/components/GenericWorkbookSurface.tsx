import {
  type GridActionsColumn,
  type GridColumn,
  type GridDensity,
  type GridRow,
  GridTable,
  GridViewport,
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
  useState,
} from "react";
import { apiPath } from "../../services/browserApi";
import { fetchWorkbookJSON, readEnvelope } from "../../services/workbookApi";
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
  genericCellLabel,
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
  inspectorPanelIsDeclared,
  selectInspectorConfig,
} from "../models/workbookInspectorModel";
import type { WorkbookQueryState } from "../models/workbookQuery";
import { requireWorkbookSurfaceRegistration } from "../models/workbookSurfaceRegistration";
import { partiesViewSchemaId } from "../models/workbookSurfaceRegistry";
import type { EntityApiRow } from "../timeline/models/workbookTimelineModel";
import {
  FocusableWorkbookCell,
  useWorkbookGridFocus,
} from "../utils/workbookGridFocus";
import { GenericMutationControl } from "./GenericMutationControl";
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
  workbookSurfaceOverlayPanelStyle,
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
  readonly loadError: string | null;
  readonly onRefresh: () => Promise<void> | void;
  readonly onToggleSort: (fieldKey: string) => void;
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
  loadError,
  onRefresh,
  onToggleSort,
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
    const payload = buildGenericCreatePayload(
      contract,
      createDraft,
      `generic-create-${contract.viewSchemaId}-${Date.now()}`,
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
  const draftInspectorFields = useMemo(() => {
    const gridFieldKeys = new Set(
      anchorColumns.map((column) => column.fieldKey),
    );
    return writableFields.filter((field) => !gridFieldKeys.has(field.fieldKey));
  }, [anchorColumns, writableFields]);
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
  const gridRows = useMemo<readonly GridRow<EntityApiRow>[]>(() => {
    const savedRows = workbookGridRows({
      getRecordId: (row: EntityApiRow) => row.record_id,
      rows,
      surface,
    });
    if (writableFields.length === 0) {
      return savedRows;
    }
    return [
      ...savedRows,
      {
        key: draftRowRecordId,
        recordId: null,
        data: draftApiRow,
        gutterContent: "+",
        gutterLabel: "Draft row",
        testId: workbookInlineDraftRowTestId(surface),
        variant: "draft",
      },
    ];
  }, [draftApiRow, draftRowRecordId, rows, surface, writableFields.length]);
  const genericFocus = useWorkbookGridFocus({
    columns: anchorColumns,
    getGroupLabel: (row, fieldKey) =>
      genericCellLabelForField(surface, fieldKey, row.cells[fieldKey]?.value),
    groupBy: queryState.groupBy,
    rows: gridRows,
    surface,
  });
  const columns: readonly GridColumn<EntityApiRow>[] = anchorColumns.map(
    (field) => ({
      ...field,
      renderCell: (row) => {
        if (row.record_id === draftRowRecordId) {
          const writableField =
            writableFields.find(
              (candidate) => candidate.fieldKey === field.fieldKey,
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
        }
        return (
          <FocusableWorkbookCell
            fieldKey={field.fieldKey}
            focus={genericFocus}
            recordId={row.record_id}
          >
            {genericCellLabelForField(
              surface,
              field.fieldKey,
              row.cells[field.fieldKey]?.value,
            )}
          </FocusableWorkbookCell>
        );
      },
    }),
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
      renderCell: ({ data: row }) => {
        if (row.record_id === draftRowRecordId) {
          return (
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
          );
        }
        return ownerRecordActions.renderRecordActions(row);
      },
    };
  }, [
    contract.viewSchemaId,
    draftRowRecordId,
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
      clientTxnId: `generic-patch-${contract.viewSchemaId}-${Date.now()}`,
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
      clientTxnId: `${txnPrefix}-${contract.viewSchemaId}-${Date.now()}`,
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
      client_txn_id: `party-from-text-${contract.viewSchemaId}-${Date.now()}`,
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
    if (!firstWritableField) {
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
  }, [writableFields]);

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
          <GridTable
            actionsColumn={rowActionsColumn}
            columns={columns}
            density={density}
            getGroupLabel={(row, fieldKey) =>
              genericCellLabel(row.cells[fieldKey]?.value)
            }
            getGroupRowTestId={(fieldKey, value) =>
              gridGroupRowTestId(surface, fieldKey, value)
            }
            groupBy={queryState.groupBy}
            onToggleSort={onToggleSort}
            rows={gridRows}
            sort={queryState.sort}
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
          addRowDisabled={writableFields.length === 0}
          leading={savedViewSelector}
          onAddRow={focusDraftRow}
          onInspectorToggle={() => {
            setIsInspectorOpen(true);
          }}
          surface={surface}
        />
      }
      viewSchemaId={surface}
      workAreaOverlays={
        <>
          {loadError ? (
            <p
              data-testid="generic-surface-load-error"
              style={surfaceNoticeOverlayStyle}
            >
              {loadError}
            </p>
          ) : null}
          {ownerRecordActions.overlay}
        </>
      }
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

const surfaceNoticeOverlayStyle = {
  ...workbookSurfaceOverlayPanelStyle,
  margin: 0,
  padding: "0.85rem 1rem",
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink-muted)",
  boxShadow: "var(--ct-elevation-popover)",
} satisfies CSSProperties;

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
