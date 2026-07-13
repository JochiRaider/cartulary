import {
  buildGridPresentationRows,
  type GridActionsColumn,
  type GridColumn,
  type GridDensity,
  type GridRow,
  GridTable,
  GridViewport,
  resolveGridPasteTargets,
} from "@cartulary/grid-adapter";
import {
  dataTestIdSelector,
  entityInspectButtonTestId,
  entityInspectorTestId,
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
  type WorkbookSurface,
  workbookInlineDraftRowTestId,
  workbookInspectorCloseButtonTestId,
  workbookRowActionMenuButtonTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import { MoreHorizontal, X } from "lucide-react";
import {
  type ClipboardEvent as ReactClipboardEvent,
  type ReactNode,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { apiPath } from "../../services/browserApi";
import {
  fetchWorkbookJSON,
  parseErrorMessage,
  readEnvelope,
} from "../../services/workbookApi";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import { useEntityTimelinePreview } from "../hooks/useEntityTimelinePreview";
import {
  buildMergePlan,
  type EntityRow,
  entityContractColumnWidth,
  entityGroupLabel,
  entityRowFromApi,
} from "../models/entityWorkbookModel";
import {
  buildGenericCreatePayload,
  buildGenericPatchChange,
  genericCellLabel,
  genericCreateMinimumMessage,
  genericRowLabel,
  initialGenericCreateDraft,
  parseMutationError,
  selectWorkbookEditTarget,
} from "../models/genericWorkbookModel";
import {
  workbookContractColumns,
  workbookGridRows,
} from "../models/workbookContractRows";
import {
  inspectorNoRowState,
  inspectorPanelIsDeclared,
  selectInspectorConfig,
} from "../models/workbookInspectorModel";
import {
  submitWorkbookPatchMutation,
  type ViewMutationEnvelope,
  type WorkbookMutationSaveState,
} from "../models/workbookMutations";
import type { WorkbookQueryState } from "../models/workbookQuery";
import { emptyGenericReferenceOptions } from "../models/workbookReferenceOptions";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import { RelationshipChip } from "../timeline/components/TimelineCellEditors";
import type { EntityApiRow } from "../timeline/models/workbookTimelineModel";
import {
  clipboardGridDimensions,
  clipboardTextLooksTabular,
} from "../utils/workbookClipboard";
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
} from "./WorkbookSurfaceFrame";

const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);

type EntityClipboardPasteEnvelope = {
  data: {
    view_schema_id: string;
    change_set_id: string;
    rows: EntityApiRow[];
  };
};

type MergeEnvelope = {
  data: {
    incident_id: string;
    record_type: "host" | "identity";
    survivor_record_id: string;
    loser_record_id: string;
    survivor_row_version: number;
    loser_row_version: number;
    change_set_id: string;
    merged_into_record_id: string;
    merge_summary: {
      record_type: string;
      repointed_mention_resolution_count: number;
      repointed_link_count: number;
      deduped_link_count: number;
      repointed_tag_count: number;
      deduped_tag_count: number;
      repointed_assessment_count: number;
      suggestion_aliases_copied_count: number;
      suggestion_alias_duplicate_noop_count: number;
      provenance_only_retained_count: number;
      exact_match_classes: Array<{
        identifier_class: string;
        promoted_count: number;
        carried_count: number;
        duplicate_noop_count: number;
        blocked_conflict_count: number;
        provenance_only_count: number;
        suggestion_only_count: number;
      }>;
    };
  };
};

type MergePreconditionDetailLine = {
  label: string;
  value: string;
};

export type EntityWorkbookSurfaceProps = {
  incidentId: string;
  apiBase?: string | undefined;
  density: GridDensity;
  entityType: EntityRow["entityType"];
  inspectorResetKey: string;
  savedViewSelector?: ReactNode | undefined;
  onToggleSort: (fieldKey: string) => void;
  queryState: WorkbookQueryState;
  rows: EntityRow[];
  currentIncidentRole: WorkbookIncidentRole | null;
  entityIndex: Record<string, EntityRow>;
  onRefreshEntities: () => Promise<void>;
};

function mergePreconditionDetailLines(
  payload: unknown,
): MergePreconditionDetailLine[] {
  if (!payload || typeof payload !== "object" || !("error" in payload)) {
    return [];
  }
  const error = payload.error;
  if (
    !error ||
    typeof error !== "object" ||
    !("details" in error) ||
    !error.details ||
    typeof error.details !== "object" ||
    Array.isArray(error.details)
  ) {
    return [];
  }
  const details = error.details as Record<string, unknown>;
  const fields: Array<readonly [string, string]> = [
    ["reason_code", "Reason"],
    ["record_type", "Record type"],
    ["identifier_class", "Identifier class"],
    ["normalized_value", "Normalized value"],
    ["blocking_record_id", "Blocking record"],
    ["survivor_record_id", "Survivor record"],
    ["loser_record_id", "Loser record"],
    ["survivor_base_row_version", "Survivor supplied version"],
    ["loser_base_row_version", "Loser supplied version"],
    ["survivor_current_row_version", "Survivor current version"],
    ["loser_current_row_version", "Loser current version"],
  ];
  return fields.flatMap(([key, label]) => {
    const value = details[key];
    if (typeof value === "number") {
      return [{ label, value: String(value) }];
    }
    return typeof value === "string" && value.trim() !== ""
      ? [{ label, value }]
      : [];
  });
}

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
  incidentId,
  apiBase,
  density,
  entityType,
  inspectorResetKey,
  savedViewSelector,
  rows,
  queryState,
  onToggleSort,
  currentIncidentRole,
  entityIndex,
  onRefreshEntities,
}: EntityWorkbookSurfaceProps) {
  const [selectedRecordId, setSelectedRecordId] = useState<string | null>(null);
  const [isInspectorOpen, setIsInspectorOpen] = useState(false);
  const [mergeCandidateId, setMergeCandidateId] = useState<string>("");
  const [mergeReason, setMergeReason] = useState("Merge duplicate entity");
  const [mergeMessage, setMergeMessage] = useState<string | null>(null);
  const [mergePreconditionDetails, setMergePreconditionDetails] = useState<
    MergePreconditionDetailLine[]
  >([]);
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
  const [mutationError, setMutationError] = useState<string | null>(null);
  const [mutationState, setMutationState] =
    useState<WorkbookMutationSaveState>("Saved");
  const { clearTimelinePreview, loadTimelinePreview, timelinePreviewRows } =
    useEntityTimelinePreview({
      apiBase,
      entityType,
      incidentId,
    });

  const selectedEntity =
    rows.find((row) => row.recordId === selectedRecordId) ?? null;
  const entityInspectorDisabledTokens = useMemo(
    () =>
      new Set<InspectorDisabledToken>(
        selectedEntity === null ? ["no_row_selected"] : [],
      ),
    [selectedEntity],
  );
  const canMerge =
    currentIncidentRole === "reviewer" || currentIncidentRole === "admin";
  const survivorLabel = selectedEntity?.label ?? "Select a record";
  const contract = entityType === "host" ? hostsContract : identitiesContract;
  const inspectorConfig = selectInspectorConfig(contract);
  const showDetailsPanel = inspectorPanelIsDeclared(inspectorConfig, "details");
  const showRelationshipsPanel = inspectorPanelIsDeclared(
    inspectorConfig,
    "relationships",
  );
  const surface: WorkbookSurface = contract.viewSchemaId;
  const draftRowRecordId = `${surface}:draft-row`;
  const writableFields = useMemo(
    () => contract.fields.filter((field) => field.writeKind !== "read_only"),
    [contract],
  );
  const editableEntityFields = useMemo(
    () => contract.fields.filter((field) => field.writeKind === "direct_value"),
    [contract],
  );
  const entityReferenceOptions = useMemo(
    () => emptyGenericReferenceOptions(),
    [],
  );
  const loserEntity =
    rows.find((row) => row.recordId === mergeCandidateId) ?? null;
  const mergePlan =
    selectedEntity && loserEntity
      ? buildMergePlan(selectedEntity, loserEntity)
      : null;
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
  const draftEntityRawRow = useMemo<EntityApiRow>(
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
  const entityGridRows = useMemo<readonly GridRow<EntityRow>[]>(() => {
    const savedRows = workbookGridRows({
      getRecordId: (row: EntityRow) => row.recordId,
      rows,
      selectedRecordId: selectedEntity?.recordId ?? null,
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
        data: draftEntityRow,
        gutterContent: "+",
        gutterLabel: "Draft row",
        testId: workbookInlineDraftRowTestId(surface),
        variant: "draft",
      },
    ];
  }, [
    draftEntityRow,
    draftRowRecordId,
    rows,
    selectedEntity?.recordId,
    surface,
    writableFields.length,
  ]);
  const entityFocus = useWorkbookGridFocus({
    columns: entityAnchorColumns,
    getGroupLabel: (row, fieldKey) => entityGroupLabel(row, fieldKey),
    groupBy: queryState.groupBy,
    rows: entityGridRows,
    surface,
  });
  const handleEntityPaste = useCallback(
    async (
      event: ReactClipboardEvent<HTMLElement>,
      anchor: { readonly fieldKey: string; readonly recordId: string },
    ) => {
      const clipboardText = event.clipboardData?.getData("text/plain") ?? "";
      if (!clipboardTextLooksTabular(clipboardText)) {
        return;
      }
      const dimensions = clipboardGridDimensions(clipboardText);
      const presentationRows = buildGridPresentationRows({
        getGroupLabel: (row, fieldKey) => entityGroupLabel(row, fieldKey),
        groupBy: queryState.groupBy,
        rows: entityGridRows,
      });
      const targetResolution = resolveGridPasteTargets({
        columns: entityAnchorColumns,
        current: anchor,
        pastedColumnCount: dimensions.columnCount,
        pastedRowCount: dimensions.rowCount,
        presentationRows,
      });
      if (targetResolution === null) {
        return;
      }

      event.preventDefault();
      setMergeMessage(null);
      setMergePreconditionDetails([]);
      const result = await fetchWorkbookJSON<EntityClipboardPasteEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${contract.viewSchemaId}/clipboard-paste`,
        ),
        {
          method: "POST",
          body: JSON.stringify({
            view_schema_id: contract.viewSchemaId,
            client_txn_id: `${contract.viewSchemaId}-paste-${Date.now()}`,
            clipboard_text: clipboardText,
            format: clipboardText.includes("\t") ? "tsv" : "csv",
            start_field_key: anchor.fieldKey,
            columns: targetResolution.columns,
            targets: targetResolution.rowTargets.map(() => ({
              kind: "create",
            })),
          }),
        },
      );
      if (!result.ok) {
        setMergeMessage(parseErrorMessage(result.payload));
        return;
      }
      const envelope = readEnvelope<EntityClipboardPasteEnvelope>(
        result.payload,
      );
      const firstRow = envelope.data.rows[0];
      await onRefreshEntities();
      if (firstRow) {
        setSelectedRecordId(firstRow.record_id);
      }
      setMergeMessage(
        `Paste applied to ${envelope.data.rows.length} ${entityType === "host" ? "host" : "identity"} row${envelope.data.rows.length === 1 ? "" : "s"}.`,
      );
    },
    [
      apiBase,
      contract.viewSchemaId,
      entityAnchorColumns,
      entityGridRows,
      entityType,
      incidentId,
      onRefreshEntities,
      queryState.groupBy,
    ],
  );
  const entityColumns: readonly GridColumn<EntityRow>[] =
    entityAnchorColumns.map((column) => ({
      ...column,
      renderCell: (row) => {
        if (row.recordId === draftRowRecordId) {
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
        }
        return (
          <FocusableWorkbookCell
            fieldKey={column.fieldKey}
            focus={entityFocus}
            onPaste={handleEntityPaste}
            recordId={row.recordId}
          >
            {entityCellContent(entityType, row, column.fieldKey)}
          </FocusableWorkbookCell>
        );
      },
    }));
  const entityActionsColumn: GridActionsColumn<EntityRow> = {
    headerTestId: gridActionsHeaderTestId(surface),
    label: "",
    width: 76,
    minWidth: 76,
    renderCell: ({ data: row }) =>
      row.recordId === draftRowRecordId ? (
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
      ) : (
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
              setMergeMessage(null);
              setMergePreconditionDetails([]);
              setIsInspectorOpen(true);
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
    if (inspectorResetKey === "") {
      return;
    }
    setIsInspectorOpen(false);
    setMergeCandidateId("");
    setMergeMessage(null);
    setMergePreconditionDetails([]);
    setEditRecordId("");
    setEditFieldKey("");
    setEditValue("");
    setAliasDraft("");
    setCreateDraft(initialGenericCreateDraft(contract, null));
  }, [contract, inspectorResetKey]);

  useEffect(() => {
    if (selectedEntityPlanInvalidationKey === "") {
      return;
    }
    setMergeCandidateId("");
  }, [selectedEntityPlanInvalidationKey]);

  useEffect(() => {
    if (selectedEntityRecordKey === "") {
      clearTimelinePreview();
      return;
    }
    setMergeMessage(null);
    setMergePreconditionDetails([]);
  }, [clearTimelinePreview, selectedEntityRecordKey]);

  useEffect(() => {
    if (!isInspectorOpen || selectedEntityRecordKey === "") {
      clearTimelinePreview();
      return;
    }
    void loadTimelinePreview(selectedEntityRecordKey);
  }, [
    clearTimelinePreview,
    isInspectorOpen,
    loadTimelinePreview,
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
      setMutationError("invalid_mutation_payload");
      return;
    }
    const change = buildGenericPatchChange(selectedEditField, editValue);
    if (change === null) {
      setMutationError(
        "Provide a value, or leave clearable fields empty to clear them.",
      );
      return;
    }
    const payload = await submitWorkbookPatchMutation({
      apiBase,
      baseRowVersion: selectedEditRow.rowVersion,
      changes: [change],
      clientTxnId: `entity-patch-${contract.viewSchemaId}-${Date.now()}`,
      recordId: selectedEditRow.recordId,
      setMutationError,
      setMutationState,
      viewSchemaId: contract.viewSchemaId,
    });
    if (payload === null) {
      return;
    }
    await onRefreshEntities();
    setSelectedRecordId(selectedEditRow.recordId);
    setMutationState("Saved");
  }

  async function submitAliasActions(
    actions: Array<
      | { op: "add_alias"; alias_text: string }
      | { op: "remove_alias"; item_ref: string }
    >,
  ) {
    if (selectedEntity === null) {
      setMutationError("invalid_mutation_payload");
      return;
    }
    const aliasFieldKey =
      entityType === "host" ? "host.aliases" : "identity.aliases";
    const payload = await submitWorkbookPatchMutation({
      apiBase,
      baseRowVersion: selectedEntity.rowVersion,
      changes: [
        {
          field_key: aliasFieldKey,
          action_payload: { kind: "collection_actions_v1", actions },
        },
      ],
      clientTxnId: `entity-alias-${selectedEntity.recordId}-${Date.now()}`,
      recordId: selectedEntity.recordId,
      setMutationError,
      setMutationState,
      viewSchemaId: contract.viewSchemaId,
    });
    if (payload === null) {
      return;
    }
    setAliasDraft("");
    await onRefreshEntities();
    setSelectedRecordId(selectedEntity.recordId);
    setMutationState("Saved");
    requestAnimationFrame(() => aliasInputRef.current?.focus());
  }

  async function submitEntityCreate() {
    const payload = buildGenericCreatePayload(
      contract,
      createDraft,
      `entity-create-${contract.viewSchemaId}-${Date.now()}`,
    );
    if (payload === null) {
      setMutationError(genericCreateMinimumMessage(contract.viewSchemaId));
      return;
    }
    setMutationState("Syncing");
    setMutationError(null);
    const result = await fetchWorkbookJSON<ViewMutationEnvelope>(
      apiPath(
        apiBase,
        `/api/v1/incidents/${incidentId}/views/${contract.viewSchemaId}/rows`,
      ),
      { method: "POST", body: JSON.stringify(payload) },
    );
    if (!result.ok) {
      setMutationState("Conflict");
      setMutationError(parseMutationError(result.payload));
      return;
    }
    const envelope = readEnvelope<ViewMutationEnvelope>(result.payload);
    setCreateDraft(initialGenericCreateDraft(contract, null));
    await onRefreshEntities();
    setSelectedRecordId(envelope.data.row.record_id);
    setMutationState("Saved");
  }

  async function confirmMerge() {
    if (!selectedEntity || !loserEntity) {
      return;
    }
    setMergeMessage(null);
    setMergePreconditionDetails([]);
    const result = await fetchWorkbookJSON<MergeEnvelope>(
      apiPath(apiBase, `/api/v1/records/${selectedEntity.recordId}/merge`),
      {
        method: "POST",
        body: JSON.stringify({
          loser_record_id: loserEntity.recordId,
          survivor_base_row_version: selectedEntity.rowVersion,
          loser_base_row_version: loserEntity.rowVersion,
          client_txn_id: `merge-${Date.now()}`,
          reason: mergeReason,
        }),
      },
    );
    if (!result.ok) {
      setMergeMessage(parseErrorMessage(result.payload));
      setMergePreconditionDetails(mergePreconditionDetailLines(result.payload));
      return;
    }

    const envelope = readEnvelope<MergeEnvelope>(result.payload);
    setMergePreconditionDetails([]);
    setMergeMessage(
      `Merged ${loserEntity.label} into ${selectedEntity.label} (${envelope.data.record_type}).`,
    );
    await onRefreshEntities();
    await loadTimelinePreview(selectedEntity.recordId);
    setSelectedRecordId(selectedEntity.recordId);
    setMergeCandidateId("");
  }

  return (
    <WorkbookSurfaceFrame
      inspector={
        isInspectorOpen ? (
          <aside
            data-testid={entityInspectorTestId(entityType)}
            style={inspectorShellStyle}
          >
            <div style={inspectorHeaderStyle}>
              <div style={inspectorTitleRowStyle}>
                <div>
                  <p style={eyebrowStyle}>Inspector</p>
                  <h2 style={inspectorTitleStyle}>{survivorLabel}</h2>
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
              <p style={bodyStyle}>
                Merge review stays inside the workbook shell.
              </p>
            </div>
            {inspectorConfig.panels.map((panel) => (
              <WorkbookInspectorPanelSection
                config={inspectorConfig}
                disabledTokens={entityInspectorDisabledTokens}
                key={panel.panelId}
                panelId={panel.panelId}
              />
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
                  <p style={bodyStyle}>{mutationError}</p>
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
                        data-testid="merge-loser-record"
                        style={selectStyle}
                        value={mergeCandidateId}
                        onChange={(event) => {
                          setMergeCandidateId(event.target.value);
                          setMergeMessage(null);
                          setMergePreconditionDetails([]);
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
                        data-testid="merge-reason"
                        style={inputStyle}
                        type="text"
                        value={mergeReason}
                        onChange={(event) => {
                          setMergeReason(event.target.value);
                        }}
                      />
                    </label>
                    {loserEntity && mergePlan ? (
                      <div data-testid="merge-plan" style={mergePlanStyle}>
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
                          data-testid="merge-confirm"
                          style={secondaryActionButtonStyle}
                          type="button"
                          onClick={() => {
                            void confirmMerge();
                          }}
                        >
                          Confirm merge
                        </button>
                      </div>
                    ) : (
                      <button
                        data-testid="merge-start"
                        style={secondaryActionButtonStyle}
                        type="button"
                        onClick={() => {
                          setMergeMessage(
                            "Select a loser to review the merge plan.",
                          );
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
                              <RelationshipChip
                                key={item.itemRef}
                                entityIndex={entityIndex}
                                item={item}
                              />
                            ))}
                          </div>
                        </article>
                      ))}
                    </div>
                  </section>
                ) : null}

                {mergeMessage ? (
                  <div style={mergeMessageBlockStyle}>
                    <p data-testid="merge-message" style={bodyStyle}>
                      {mergeMessage}
                    </p>
                    {mergePreconditionDetails.length > 0 ? (
                      <ul
                        data-testid={entityMergePreconditionDetailsTestId(
                          entityType,
                          selectedEntity.recordId,
                        )}
                        style={flatListStyle}
                      >
                        {mergePreconditionDetails.map((line) => (
                          <li key={line.label}>
                            {line.label}: {line.value}
                          </li>
                        ))}
                      </ul>
                    ) : null}
                  </div>
                ) : null}
              </>
            ) : (
              <p style={bodyStyle}>{inspectorNoRowState(inspectorConfig)}</p>
            )}
          </aside>
        ) : undefined
      }
      primaryGrid={
        <GridViewport
          blockSizing="fill"
          style={gridShellStyle}
          testId={gridShellTestId(surface)}
        >
          <GridTable
            actionsColumn={entityActionsColumn}
            columns={entityColumns}
            density={density}
            getGroupLabel={(row, fieldKey) => entityGroupLabel(row, fieldKey)}
            getGroupRowTestId={(fieldKey, value) =>
              gridGroupRowTestId(surface, fieldKey, value)
            }
            groupBy={queryState.groupBy}
            onToggleSort={onToggleSort}
            rows={entityGridRows}
            sort={queryState.sort}
          />
        </GridViewport>
      }
      statusStrip={
        <WorkbookSurfaceStatusStrip
          mutationError={mutationError}
          mutationState={mutationState}
          workbookFocusAnchor={entityFocus.anchor}
        />
      }
      viewBar={
        <WorkbookSheetToolbar
          addRowDisabled={writableFields.length === 0}
          leading={savedViewSelector}
          onAddRow={() => {
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
          }}
          onInspectorToggle={() => {
            setIsInspectorOpen(true);
          }}
          surface={surface}
        />
      }
      viewSchemaId={surface}
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

const inspectorShellStyle = {
  ...workbookSurfaceInspectorPanelStyle,
};

const inspectorHeaderStyle = {
  display: "grid",
  gap: "0.35rem",
  marginBottom: "1rem",
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
