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
  evidenceAccessMessageTestId,
  evidenceAttachFileInputTestId,
  evidenceDownloadButtonTestId,
  evidencePreviewButtonTestId,
  evidencePreviewFrameTestId,
  evidencePreviewPanelTestId,
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
  type Dispatch,
  type ReactNode,
  type SetStateAction,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from "react";
import { apiPath } from "../../services/browserApi";
import { fetchJSON, readEnvelope } from "../../services/workbookApi";
import {
  createAndAttachEvidenceBlob,
  evidenceAccessMessageLiveRegion,
  issueEvidenceAccessHandle,
} from "../../services/workbookEvidence";
import { useGenericReferenceOptions } from "../hooks/useGenericReferenceOptions";
import { buildEvidenceLifecycleViewModel } from "../models/evidenceLifecycleViewModel";
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
  parseMutationError,
  partyLinkPairsForContract,
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
import {
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  notesViewSchemaId,
  partiesViewSchemaId,
  taskRequestsViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import type { EntityApiRow } from "../timeline/models/workbookTimelineModel";
import {
  FocusableWorkbookCell,
  useWorkbookGridFocus,
} from "../utils/workbookGridFocus";
import { stringifyGridValue } from "../utils/workbookValueFormat";
import { GenericMutationControl } from "./GenericMutationControl";
import {
  type InspectorDisabledToken,
  WorkbookInspectorPanelSection,
} from "./WorkbookInspectorFeatureGroups";
import { WorkbookSheetToolbar } from "./WorkbookSheetToolbar";
import {
  type WorkbookStatusSaveState,
  WorkbookSurfaceStatusStrip,
} from "./WorkbookStatusStrip";
import {
  WorkbookSurfaceFrame,
  workbookSurfaceGridShellStyle,
  workbookSurfaceInspectorPanelStyle,
  workbookSurfaceOverlayPanelStyle,
} from "./WorkbookSurfaceFrame";

type GenericMutationErrorSetter = Dispatch<SetStateAction<string | null>>;
type GenericMutationStateSetter = Dispatch<
  SetStateAction<WorkbookStatusSaveState>
>;

type ViewMutationEnvelope = {
  data: {
    view_schema_id: string;
    change_set_id: string;
    row: EntityApiRow;
  };
};

type DecisionSupersedeEnvelope = {
  data: {
    view_schema_id: string;
    change_set_id: string;
    target_record_id: string;
    superseding_record_id: string;
    target_row_version: number;
    superseding_row_version: number;
    target_status: string;
    reason: string;
  };
};

type EvidencePreviewState = {
  href: string;
  recordId: string;
  title: string;
  previewKind: string | null;
};

export type GenericWorkbookSurfaceProps = {
  readonly apiBase?: string | undefined;
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

function selectGenericEditTarget<
  Row,
  Field extends { readonly fieldKey: string },
>({
  fieldKey,
  fields,
  getRecordId,
  recordId,
  rows,
}: {
  readonly fieldKey: string;
  readonly fields: readonly Field[];
  readonly getRecordId: (row: Row) => string;
  readonly recordId: string;
  readonly rows: readonly Row[];
}): { readonly field: Field | null; readonly row: Row | null } {
  return {
    row: rows.find((row) => getRecordId(row) === recordId) ?? null,
    field:
      fields.find((field) => field.fieldKey === fieldKey) ?? fields[0] ?? null,
  };
}

function normalizeValue(value: string): string {
  return value.trim();
}

async function submitGenericRecordPatch({
  apiBase,
  baseRowVersion,
  changes,
  clientTxnId,
  recordId,
  viewSchemaId,
}: {
  readonly apiBase: string | undefined;
  readonly baseRowVersion: number;
  readonly changes: readonly Record<string, unknown>[];
  readonly clientTxnId: string;
  readonly recordId: string;
  readonly viewSchemaId: string;
}) {
  return fetchJSON<ViewMutationEnvelope>(
    apiPath(apiBase, `/api/v1/records/${recordId}`),
    {
      method: "PATCH",
      body: JSON.stringify({
        view_schema_id: viewSchemaId,
        base_row_version: baseRowVersion,
        client_txn_id: clientTxnId,
        changes,
      }),
    },
  );
}

async function submitGenericPatchMutation({
  apiBase,
  baseRowVersion,
  changes,
  clientTxnId,
  recordId,
  setMutationError,
  setMutationState,
  viewSchemaId,
}: {
  readonly apiBase: string | undefined;
  readonly baseRowVersion: number;
  readonly changes: readonly Record<string, unknown>[];
  readonly clientTxnId: string;
  readonly recordId: string;
  readonly setMutationError: GenericMutationErrorSetter;
  readonly setMutationState: GenericMutationStateSetter;
  readonly viewSchemaId: string;
}) {
  setMutationState("Syncing");
  setMutationError(null);
  const result = await submitGenericRecordPatch({
    apiBase,
    baseRowVersion,
    changes,
    clientTxnId,
    recordId,
    viewSchemaId,
  });
  if (!result.ok) {
    setMutationState("Conflict");
    setMutationError(parseMutationError(result.payload));
    return null;
  }
  return result.payload;
}

export function GenericWorkbookSurface({
  apiBase,
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
}: GenericWorkbookSurfaceProps) {
  const surface = contract.viewSchemaId as WorkbookSurface;
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
    useGenericReferenceOptions({ apiBase, incidentId });
  const [mutationError, setMutationError] = useState<string | null>(null);
  const [mutationState, setMutationState] =
    useState<WorkbookStatusSaveState>("Saved");
  const [evidenceMessageByRecordID, setEvidenceMessageByRecordID] = useState<
    Record<string, string>
  >({});
  const [evidencePreview, setEvidencePreview] =
    useState<EvidencePreviewState | null>(null);
  const isEvidenceSurface = contract.viewSchemaId === evidenceViewSchemaId;
  const isNotesSurface = contract.viewSchemaId === notesViewSchemaId;
  const isTaskRequestSurface =
    contract.viewSchemaId === taskRequestsViewSchemaId;
  const isDecisionSurface = contract.viewSchemaId === decisionsViewSchemaId;
  const [taskLifecycleRecordId, setTaskLifecycleRecordId] = useState("");
  const [taskLifecycleStatus, setTaskLifecycleStatus] = useState("blocked");
  const [taskLifecycleBlockedReason, setTaskLifecycleBlockedReason] =
    useState("");
  const [decisionSupersedeTargetId, setDecisionSupersedeTargetId] =
    useState("");
  const [decisionSupersedeReplacementId, setDecisionSupersedeReplacementId] =
    useState("");
  const [decisionSupersedeReason, setDecisionSupersedeReason] = useState("");
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
    setTaskLifecycleRecordId("");
    setTaskLifecycleBlockedReason("");
    setDecisionSupersedeTargetId("");
    setDecisionSupersedeReplacementId("");
    setDecisionSupersedeReason("");
    setEvidencePreview(null);
    setMutationError(null);
  }, [inspectorResetKey]);

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

  const setEvidenceMessage = useCallback(
    (recordId: string, message: string | null) => {
      setEvidenceMessageByRecordID((current) => {
        const next = { ...current };
        if (message === null) {
          delete next[recordId];
        } else {
          next[recordId] = message;
        }
        return next;
      });
    },
    [],
  );

  const issueEvidenceHandle = useCallback(
    async (row: EntityApiRow, kind: "preview" | "download") => {
      setEvidenceMessage(row.record_id, null);
      const handle = await issueEvidenceAccessHandle({
        apiBase,
        evidenceRecordId: row.record_id,
        kind,
      });
      if (!handle.ok) {
        setEvidenceMessage(row.record_id, handle.message);
        return;
      }
      if (kind === "preview") {
        setEvidencePreview({
          href: handle.href,
          recordId: row.record_id,
          title:
            stringifyGridValue(row.cells["evidence.title"]?.value).trim() ||
            row.record_id,
          previewKind: handle.previewKind,
        });
        setEvidenceMessage(row.record_id, "Preview loaded inline.");
        return;
      }

      const anchor = document.createElement("a");
      anchor.href = handle.href;
      anchor.download = handle.filename || "evidence";
      anchor.rel = "noopener";
      document.body.append(anchor);
      anchor.click();
      anchor.remove();
      setEvidenceMessage(row.record_id, "Download handle issued.");
    },
    [apiBase, setEvidenceMessage],
  );

  const attachEvidenceFile = useCallback(
    async (row: EntityApiRow, file: File) => {
      if (file.size <= 0) {
        setEvidenceMessage(row.record_id, "Evidence attach failed.");
        return;
      }
      setEvidenceMessage(row.record_id, "Uploading evidence.");
      setMutationState("Syncing");
      try {
        await createAndAttachEvidenceBlob({
          apiBase,
          attachClientTxnId: () => `evidence-attach-${Date.now()}`,
          baseRowVersion: row.row_version,
          createClientTxnId: () => `evidence-blob-${Date.now()}`,
          evidenceRecordId: row.record_id,
          file,
          incidentId,
        });
        setEvidenceMessage(row.record_id, "Evidence attached.");
        setMutationState("Saved");
        await onRefresh();
      } catch (error) {
        setEvidenceMessage(
          row.record_id,
          error instanceof Error ? error.message : "Evidence attach failed.",
        );
        setMutationState("Conflict");
      }
    },
    [apiBase, incidentId, onRefresh, setEvidenceMessage],
  );

  const completeGenericMutation = useCallback(
    async <TEnvelope,>(payload: unknown) => {
      const envelope = readEnvelope<TEnvelope>(payload);
      try {
        await onRefresh();
        await refreshReferenceOptions();
      } catch (error) {
        setMutationState("Conflict");
        setMutationError(
          error instanceof Error ? error.message : "Workbook refresh failed.",
        );
        return envelope;
      }
      setMutationState("Saved");
      return envelope;
    },
    [onRefresh, refreshReferenceOptions],
  );

  const submitCreate = useCallback(async () => {
    const payload = buildGenericCreatePayload(
      contract,
      createDraft,
      `generic-create-${contract.viewSchemaId}-${Date.now()}`,
    );
    if (payload === null) {
      setMutationError(genericCreateMinimumMessage(contract.viewSchemaId));
      return;
    }
    setMutationState("Syncing");
    setMutationError(null);
    const createPath =
      isNotesSurface && linkedNoteSourceRecordId !== ""
        ? `/api/v1/records/${linkedNoteSourceRecordId}/linked-notes`
        : `/api/v1/incidents/${incidentId}/views/${contract.viewSchemaId}/rows`;
    const result = await fetchJSON<ViewMutationEnvelope>(
      apiPath(apiBase, createPath),
      { method: "POST", body: JSON.stringify(payload) },
    );
    if (!result.ok) {
      setMutationState("Conflict");
      setMutationError(parseMutationError(result.payload));
      return;
    }
    setCreateDraft(initialGenericCreateDraft(contract, currentUserId));
    setLinkedNoteSourceRecordId("");
    await completeGenericMutation<ViewMutationEnvelope>(result.payload);
  }, [
    apiBase,
    completeGenericMutation,
    contract,
    createDraft,
    currentUserId,
    incidentId,
    isNotesSurface,
    linkedNoteSourceRecordId,
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
    if (!isEvidenceSurface && writableFields.length === 0) {
      return undefined;
    }
    return {
      headerTestId: gridActionsHeaderTestId(surface),
      label: "",
      width: isEvidenceSurface ? 208 : 76,
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
        if (!isEvidenceSurface) {
          return null;
        }
        const evidenceAccess = buildEvidenceLifecycleViewModel({
          evidenceLifecycleState: row.cells["evidence.lifecycle_state"]?.value,
          objectBlobUploadState: row.cells["evidence.upload_state"]?.value,
        });
        const message =
          evidenceMessageByRecordID[row.record_id] ?? evidenceAccess.message;
        const messageLiveRegion =
          message === null
            ? null
            : evidenceAccessMessageLiveRegion(message, evidenceAccess);
        return (
          <div
            data-evidence-state-key={evidenceAccess.stateKey}
            style={actionStackStyle}
          >
            <div style={inlineButtonRowStyle}>
              <button
                data-testid={evidencePreviewButtonTestId(row.record_id)}
                disabled={!evidenceAccess.canPreview}
                style={actionButtonStyle}
                type="button"
                onClick={() => {
                  void issueEvidenceHandle(row, "preview");
                }}
              >
                Preview
              </button>
              <button
                data-testid={evidenceDownloadButtonTestId(row.record_id)}
                disabled={!evidenceAccess.canDownload}
                style={actionButtonStyle}
                type="button"
                onClick={() => {
                  void issueEvidenceHandle(row, "download");
                }}
              >
                Download
              </button>
            </div>
            <label style={labelStyle}>
              Attach file
              <input
                data-testid={evidenceAttachFileInputTestId(row.record_id)}
                style={inputStyle}
                type="file"
                accept="image/*,.txt,.pdf,text/plain,application/pdf"
                onChange={(event) => {
                  const [file] = Array.from(event.currentTarget.files ?? []);
                  event.currentTarget.value = "";
                  if (file) {
                    void attachEvidenceFile(row, file);
                  }
                }}
              />
            </label>
            {message ? (
              <span
                aria-live={messageLiveRegion?.ariaLive}
                data-testid={evidenceAccessMessageTestId(row.record_id)}
                role={messageLiveRegion?.role}
                style={evidenceAccessMessageStyle}
              >
                {message}
              </span>
            ) : null}
          </div>
        );
      },
    };
  }, [
    attachEvidenceFile,
    contract.viewSchemaId,
    draftRowRecordId,
    evidenceMessageByRecordID,
    isEvidenceSurface,
    isInspectorOpen,
    issueEvidenceHandle,
    mutationState,
    surface,
    submitCreate,
    writableFields.length,
  ]);
  const { row: selectedEditRow, field: selectedEditField } =
    selectGenericEditTarget({
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
      setMutationError("invalid_mutation_payload");
      return;
    }
    const change = buildGenericPatchChange(
      selectedEditField,
      editValue,
      editCollectionMode,
    );
    if (change === null) {
      setMutationError(
        "Provide a value, or leave clearable fields empty to clear them.",
      );
      return;
    }
    const payload = await submitGenericPatchMutation({
      apiBase,
      baseRowVersion: selectedEditRow.row_version,
      changes: [change],
      clientTxnId: `generic-patch-${contract.viewSchemaId}-${Date.now()}`,
      recordId: selectedEditRow.record_id,
      setMutationError,
      setMutationState,
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
      setMutationError("Select a row before changing a party link.");
      return false;
    }
    const payload = await submitGenericPatchMutation({
      apiBase,
      baseRowVersion: selectedEditRow.row_version,
      changes,
      clientTxnId: `${txnPrefix}-${contract.viewSchemaId}-${Date.now()}`,
      recordId: selectedEditRow.record_id,
      setMutationError,
      setMutationState,
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
      setMutationError("Select a row and party field first.");
      return;
    }
    const rawText = normalizeValue(
      String(
        selectedEditRow.cells[selectedPartyLinkPair.textFieldKey]?.value ?? "",
      ),
    );
    if (rawText === "") {
      setMutationError("Party text is empty.");
      return;
    }
    setMutationState("Syncing");
    setMutationError(null);
    const createPayload: Record<string, unknown> = {
      client_txn_id: `party-from-text-${contract.viewSchemaId}-${Date.now()}`,
      "party.display_name": rawText,
      "party.party_kind": "person",
    };
    const email = extractEmailFromPartyText(rawText);
    if (email !== null) {
      createPayload["party.primary_email"] = email;
    }
    const createResult = await fetchJSON<ViewMutationEnvelope>(
      apiPath(
        apiBase,
        `/api/v1/incidents/${incidentId}/views/${partiesViewSchemaId}/rows`,
      ),
      { method: "POST", body: JSON.stringify(createPayload) },
    );
    if (!createResult.ok) {
      setMutationState("Conflict");
      setMutationError(parseMutationError(createResult.payload));
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
      setMutationError("Select an existing party.");
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
      setMutationError("Select a party field first.");
      return;
    }
    await submitPartyLinkPatch(
      [{ field_key: selectedPartyLinkPair.refFieldKey, value: null }],
      "party-clear-link",
    );
  };

  const clearPartyText = async () => {
    if (selectedPartyLinkPair === null) {
      setMutationError("Select a party field first.");
      return;
    }
    await submitPartyLinkPatch(
      [{ field_key: selectedPartyLinkPair.textFieldKey, value: null }],
      "party-clear-text",
    );
  };

  const clearPartyBoth = async () => {
    if (selectedPartyLinkPair === null) {
      setMutationError("Select a party field first.");
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

  const submitTaskLifecyclePatch = async () => {
    const target = rows.find((row) => row.record_id === taskLifecycleRecordId);
    if (!target) {
      setMutationError("Select a task row.");
      return;
    }
    const changes: Array<Record<string, unknown>> = [
      { field_key: "task.status", value: taskLifecycleStatus },
    ];
    if (taskLifecycleStatus === "blocked") {
      const reason = normalizeValue(taskLifecycleBlockedReason);
      if (reason === "") {
        setMutationError("Blocked tasks need a reason.");
        return;
      }
      changes.push({ field_key: "task.blocked_reason", value: reason });
    }
    setMutationState("Syncing");
    setMutationError(null);
    const result = await fetchJSON<ViewMutationEnvelope>(
      apiPath(apiBase, `/api/v1/records/${target.record_id}`),
      {
        method: "PATCH",
        body: JSON.stringify({
          view_schema_id: taskRequestsViewSchemaId,
          base_row_version: target.row_version,
          client_txn_id: `task-lifecycle-${Date.now()}`,
          changes,
        }),
      },
    );
    if (!result.ok) {
      setMutationState("Conflict");
      setMutationError(parseMutationError(result.payload));
      return;
    }
    if (taskLifecycleStatus !== "blocked") {
      setTaskLifecycleBlockedReason("");
    }
    await completeGenericMutation<ViewMutationEnvelope>(result.payload);
  };

  const submitDecisionSupersede = async () => {
    const target = rows.find(
      (row) => row.record_id === decisionSupersedeTargetId,
    );
    if (!target || decisionSupersedeReplacementId === "") {
      setMutationError("Select target and superseding decisions.");
      return;
    }
    if (target.record_id === decisionSupersedeReplacementId) {
      setMutationError("Select a different superseding decision.");
      return;
    }
    const reason = normalizeValue(decisionSupersedeReason);
    if (reason === "") {
      setMutationError("Reason is required.");
      return;
    }
    setMutationState("Syncing");
    setMutationError(null);
    const result = await fetchJSON<DecisionSupersedeEnvelope>(
      apiPath(apiBase, `/api/v1/records/${target.record_id}/supersede`),
      {
        method: "POST",
        body: JSON.stringify({
          base_row_version: target.row_version,
          client_txn_id: `decision-supersede-${Date.now()}`,
          replacement_record_id: decisionSupersedeReplacementId,
          reason,
        }),
      },
    );
    if (!result.ok) {
      setMutationState("Conflict");
      setMutationError(parseMutationError(result.payload));
      return;
    }
    setDecisionSupersedeReason("");
    await completeGenericMutation<DecisionSupersedeEnvelope>(result.payload);
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

            {showWorkflowPanel && isTaskRequestSurface && rows.length > 0 ? (
              <div style={genericEditRowStyle}>
                <select
                  aria-label="Task lifecycle row"
                  data-testid="task-lifecycle-target"
                  style={selectStyle}
                  value={taskLifecycleRecordId}
                  onChange={(event) => {
                    setTaskLifecycleRecordId(event.target.value);
                  }}
                >
                  <option value="">Task</option>
                  {rows.map((row) => (
                    <option key={row.record_id} value={row.record_id}>
                      {genericRowLabel(contract, row)}
                    </option>
                  ))}
                </select>
                <select
                  aria-label="Task lifecycle status"
                  data-testid="task-lifecycle-status"
                  style={selectStyle}
                  value={taskLifecycleStatus}
                  onChange={(event) => {
                    setTaskLifecycleStatus(event.target.value);
                  }}
                >
                  <option value="open">open</option>
                  <option value="in_progress">in_progress</option>
                  <option value="blocked">blocked</option>
                  <option value="done">done</option>
                  <option value="canceled">canceled</option>
                </select>
                <input
                  aria-label="Blocked reason"
                  data-testid="task-lifecycle-blocked-reason"
                  disabled={taskLifecycleStatus !== "blocked"}
                  style={inputStyle}
                  type="text"
                  value={taskLifecycleBlockedReason}
                  onChange={(event) => {
                    setTaskLifecycleBlockedReason(event.target.value);
                  }}
                />
                <button
                  data-testid="task-lifecycle-submit"
                  disabled={mutationState === "Syncing"}
                  style={secondaryActionButtonStyle}
                  type="button"
                  onClick={() => {
                    void submitTaskLifecyclePatch();
                  }}
                >
                  Apply task status
                </button>
              </div>
            ) : null}

            {showWorkflowPanel && isDecisionSurface && rows.length > 1 ? (
              <div style={genericEditRowStyle}>
                <select
                  aria-label="Superseded decision"
                  data-testid="decision-supersede-target"
                  style={selectStyle}
                  value={decisionSupersedeTargetId}
                  onChange={(event) => {
                    setDecisionSupersedeTargetId(event.target.value);
                  }}
                >
                  <option value="">Target</option>
                  {rows.map((row) => (
                    <option key={row.record_id} value={row.record_id}>
                      {genericRowLabel(contract, row)}
                    </option>
                  ))}
                </select>
                <select
                  aria-label="Superseding decision"
                  data-testid="decision-supersede-replacement"
                  style={selectStyle}
                  value={decisionSupersedeReplacementId}
                  onChange={(event) => {
                    setDecisionSupersedeReplacementId(event.target.value);
                  }}
                >
                  <option value="">Superseding</option>
                  {referenceOptions.decisions.map((option) => (
                    <option key={option.recordId} value={option.recordId}>
                      {option.label}
                    </option>
                  ))}
                </select>
                <input
                  aria-label="Decision supersession reason"
                  data-testid="decision-supersede-reason"
                  style={inputStyle}
                  type="text"
                  value={decisionSupersedeReason}
                  onChange={(event) => {
                    setDecisionSupersedeReason(event.target.value);
                  }}
                />
                <button
                  data-testid="decision-supersede-submit"
                  disabled={mutationState === "Syncing"}
                  style={secondaryActionButtonStyle}
                  type="button"
                  onClick={() => {
                    void submitDecisionSupersede();
                  }}
                >
                  Supersede decision
                </button>
              </div>
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
          {isEvidenceSurface && evidencePreview ? (
            <section
              data-testid={evidencePreviewPanelTestId()}
              style={evidencePreviewPanelStyle}
            >
              <div style={evidencePreviewHeaderStyle}>
                <div>
                  <p style={eyebrowStyle}>Preview</p>
                  <h2 style={sectionTitleStyle}>{evidencePreview.title}</h2>
                </div>
                <button
                  style={secondaryActionButtonStyle}
                  type="button"
                  onClick={() => {
                    setEvidencePreview(null);
                  }}
                >
                  Close
                </button>
              </div>
              <iframe
                data-testid={evidencePreviewFrameTestId(
                  evidencePreview.recordId,
                )}
                src={evidencePreview.href}
                style={evidencePreviewFrameStyle}
                title={`Evidence preview ${evidencePreview.title}`}
              />
              {evidencePreview.previewKind ? (
                <p style={evidenceAccessMessageStyle}>
                  {evidencePreview.previewKind}
                </p>
              ) : null}
            </section>
          ) : null}
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

const actionStackStyle = {
  display: "grid",
  gap: "0.5rem",
};

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

const evidenceAccessMessageStyle = {
  margin: 0,
  fontSize: "0.85rem",
  color: "var(--ct-colors-ink-muted)",
};

const evidencePreviewPanelStyle = {
  ...workbookSurfaceOverlayPanelStyle,
  display: "grid",
  gap: "0.75rem",
  padding: "1rem",
  borderRadius: "var(--ct-rounded-lg)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-popover)",
};

const evidencePreviewHeaderStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "1rem",
  alignItems: "start",
};

const evidencePreviewFrameStyle = {
  width: "100%",
  blockSize: "min(28rem, 34vh)",
  minHeight: "12rem",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-md)",
  background: "var(--ct-colors-surface-2)",
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

const sectionTitleStyle = {
  margin: 0,
  fontSize: "1rem",
};

const inlineButtonRowStyle = {
  display: "flex",
  gap: "0.5rem",
  flexWrap: "wrap" as const,
};

const selectStyle = {
  ...inputStyle,
  appearance: "auto" as const,
};
