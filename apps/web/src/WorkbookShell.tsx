import {
  buildGridPresentationRows,
  type GridActionsColumn,
  type GridColumn,
  type GridRow,
  GridTable,
  GridViewport,
  reconcileRecordRows,
  resolveGridPasteTargets,
} from "@cartulary/grid-adapter";
import type {
  EvidenceHandleEnvelope,
  EvidenceHandleIssueRequest,
} from "@cartulary/protocol-ts";
import {
  assessmentCreatePanelTestId,
  currentIncidentRoleTestId,
  dataTestIdSelector,
  entityInspectButtonTestId,
  entityInspectorTestId,
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
  gridRowTestId,
  gridShellTestId,
  gridSortHeaderTestId,
  incidentControlsPanelTestId,
  incidentControlsTriggerTestId,
  phase1LandingTestId,
  phase1RouteTestId,
  saveStateTestId,
  surfaceTabTestId,
  timelinePreviewRowTestId,
  type WorkbookSurface,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import {
  requireViewContract,
  resolveHeaderSortFieldKey,
  type ViewContract,
  visibleFields,
} from "@cartulary/view-contracts";
import {
  type Dispatch,
  type ClipboardEvent as ReactClipboardEvent,
  type ReactNode,
  type SetStateAction,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { ActiveSurfaceSavedViewSelector } from "./ActiveSurfaceSavedViewSelector";
import {
  assessmentColumnWidth,
  initialAssessmentDraft,
  isAssessmentConfidenceBand,
  supportRowLabel,
} from "./assessmentWorkbookModel";
import { apiPath } from "./browserApi";
import {
  buildMergePlan,
  type EntityRow,
  entityContractColumnWidth,
  entityGroupLabel,
  entityRowFromApi,
} from "./entityWorkbookModel";
import { buildEvidenceLifecycleViewModel } from "./evidenceLifecycleViewModel";
import { GenericMutationControl } from "./GenericMutationControl";
import {
  buildGenericCreatePayload,
  buildGenericPatchChange,
  enumValuesFor,
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
} from "./genericWorkbookModel";
import { IncidentAdminPanel } from "./IncidentAdminPanel";
import { SystemViewSwitcher } from "./SystemViewSwitcher";
import { RelationshipChip } from "./TimelineCellEditors";
import {
  buildRecordRollbackTargetFromHistoryAction,
  TimelineWorkbook,
} from "./TimelineWorkbook";
import { useAssessmentSupportRows } from "./useAssessmentSupportRows";
import { useEntityTimelinePreview } from "./useEntityTimelinePreview";
import { useGenericReferenceOptions } from "./useGenericReferenceOptions";
import { useWorkbookShellRuntime } from "./useWorkbookShellRuntime";
import { WorkbookGridControls } from "./WorkbookGridControls";
import { WorkbookShellSlotRegion, workbookShellId } from "./WorkbookShellSlots";
import {
  abortLatestQuery,
  beginLatestQuery,
  fetchJSON,
  handleWorkbookLoadFailure,
  isAbortError,
  type LatestQueryRuntime,
  parseErrorMessage,
  readEnvelope,
} from "./workbookApi";
import { clipboardGridDimensions } from "./workbookClipboard";
import {
  normalizeWorkbookViewRows,
  workbookContractColumns,
} from "./workbookContractRows";
import {
  createAndAttachEvidenceBlob,
  evidenceAccessMessageLiveRegion,
  evidencePublicErrorMessage,
  resolvePublicEvidenceHandleHref,
} from "./workbookEvidence";
import {
  FocusableWorkbookCell,
  useWorkbookGridFocus,
  WorkbookFocusAnchorStatus,
} from "./workbookGridFocus";
import { pendingReplayCapacity } from "./workbookPendingQueue";
import { displayInitials } from "./workbookPresence";
import {
  applyFilterDraft,
  buildQueryRequest,
  buildSavedViewLayoutJson,
  buildSavedViewQueryJson,
  defaultFilterDraft,
  emptyWorkbookQueryState,
  type FilterDraft,
  removeFilterField,
  toggleSortField,
  updateGroupBy,
  type WorkbookQueryState,
} from "./workbookQuery";
import { emptyGenericReferenceOptions } from "./workbookReferenceOptions";
import { savedViewQueryStateForRuntime } from "./workbookSavedViewRuntime";
import {
  normalizeSavedViewResource,
  type SavedViewEnvelope,
  type SavedViewListEnvelope,
  type SavedViewResource,
  savedViewLayoutJsonForPersistence,
  savedViewQueryJsonForPersistence,
} from "./workbookSavedViews";
import {
  normalizeWorkbookStartupSelection,
  workbookStartupQueryFromURLParams,
} from "./workbookStartup";
import {
  statusIconStyle,
  statusStripItemStyle,
  statusStripStyle,
} from "./workbookStyles";
import {
  assessmentsViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  knownWorkbookViewSchemaId,
  listWorkbookSurfaceRegistryEntries,
  notesViewSchemaId,
  partiesViewSchemaId,
  requiredBuiltInWorkbookSurfaceIds,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "./workbookSurfaceRegistry";
import {
  type AssessmentCreateDraft,
  buildAssessmentCreatePayload,
  buildCreatePayload,
  clipboardTextLooksTabular,
  confidenceScoreFromBand,
  createDraftRow,
  decideWorkbookRecordFreshness,
  type EntityApiRow,
  ensureDraftRow,
  type WorkbookRecordFreshnessDecision,
  type WorkbookVersionedRecord,
} from "./workbookTimelineModel";
import { stringifyGridValue } from "./workbookValueFormat";

export type {
  RecordHistoryItem,
  RecordHistoryRollbackAction,
} from "./TimelineHistoryPanel";
export type { TimelineWorkbookProps } from "./TimelineWorkbook";
export type { WorkbookRecordFreshnessDecision, WorkbookVersionedRecord };
export {
  buildAssessmentCreatePayload,
  buildCreatePayload,
  buildRecordRollbackTargetFromHistoryAction,
  clipboardTextLooksTabular,
  confidenceScoreFromBand,
  createDraftRow,
  decideWorkbookRecordFreshness,
  ensureDraftRow,
  pendingReplayCapacity,
  TimelineWorkbook,
};

const timelineContract = requireViewContract(timelineViewSchemaId);
const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);
const assessmentsContract = requireViewContract(assessmentsViewSchemaId);
const allWorkbookContracts = listWorkbookSurfaceRegistryEntries().map(
  (entry) => entry.contract,
);

type SaveState = "Syncing" | "Saved" | "Conflict";
type FilterDraftSetter = Dispatch<SetStateAction<FilterDraft>>;
type MutationErrorSetter = Dispatch<SetStateAction<string | null>>;
type MutationStateSetter = Dispatch<SetStateAction<SaveState>>;
type WorkbookQueryStateSetter = Dispatch<SetStateAction<WorkbookQueryState>>;
type IncidentRole = "viewer" | "editor" | "reviewer" | "admin" | "";

type WorkbookShellProps = {
  incidentId: string;
  apiBase?: string | undefined;
  currentUserLabel?: string | undefined;
  onIncidentSnapshot?:
    | ((incident: {
        incident_id: string;
        incident_key: string;
        title: string;
        description: string | null;
        severity: string | null;
        tlp: string | null;
        current_phase: string | null;
        primary_external_case_ref: string | null;
        incident_version: number;
      }) => void)
    | undefined;
  onIncidentAccessLost?: (() => void) | undefined;
  onReturnToLanding?: (() => void) | undefined;
};

function workbookContractForViewSchemaId(viewSchemaId: string): ViewContract {
  return (
    allWorkbookContracts.find(
      (contract) => contract.viewSchemaId === viewSchemaId,
    ) ?? timelineContract
  );
}

type TimelineMutationEnvelope = {
  data: {
    view_schema_id: string;
    change_set_id?: string;
    row: unknown;
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

type SessionEnvelope = {
  data: {
    user_id: string;
    memberships: Array<{
      incident_id: string;
      role: IncidentRole;
    }>;
  };
};

type ViewQueryEnvelope = {
  data: {
    incident_id: string;
    view_schema_id: string;
    rows: EntityApiRow[];
  };
};

type ViewMutationEnvelope = {
  data: {
    view_schema_id: string;
    change_set_id: string;
    row: EntityApiRow;
  };
};

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

type EvidencePreviewState = {
  href: string;
  recordId: string;
  title: string;
  previewKind: string | null;
};

type WorkbookStartupEnvelope = {
  data?: unknown;
};

function buildWorkbookGridRows<Row>({
  getRecordId,
  rows,
  selectedRecordId,
  surface,
}: {
  readonly getRecordId: (row: Row) => string;
  readonly rows: readonly Row[];
  readonly selectedRecordId?: string | null | undefined;
  readonly surface: WorkbookSurface;
}): readonly GridRow<Row>[] {
  return rows.map((row) => {
    const recordId = getRecordId(row);
    return {
      key: recordId,
      recordId,
      data: row,
      ...(selectedRecordId === null || selectedRecordId === undefined
        ? {}
        : { selected: recordId === selectedRecordId }),
      testId: gridRowTestId(surface, recordId),
    };
  });
}

function selectWorkbookEditTarget<
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

function clearAppliedFilterDraft(current: FilterDraft): FilterDraft {
  return {
    ...current,
    booleanValue: "",
    value: "",
  };
}

function applyFilterDraftToQuery(
  setQueryState: WorkbookQueryStateSetter,
  setFilterDraft: FilterDraftSetter,
  draft: FilterDraft,
): void {
  setQueryState((current) => applyFilterDraft(current, draft));
  setFilterDraft(clearAppliedFilterDraft);
}

function normalizeValue(value: string): string {
  return value.trim();
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

async function submitViewRecordPatch({
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

async function submitWorkbookPatchMutation({
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
  readonly setMutationError: MutationErrorSetter;
  readonly setMutationState: MutationStateSetter;
  readonly viewSchemaId: string;
}) {
  setMutationState("Syncing");
  setMutationError(null);
  const result = await submitViewRecordPatch({
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

function EntityWorkbookSurface({
  incidentId,
  apiBase,
  entityType,
  savedViewSelector,
  filterDraft,
  onApplyFilter,
  onFilterDraftChange,
  onGroupByChange,
  onRemoveFilter,
  rows,
  onToggleSort,
  queryState,
  currentIncidentRole,
  entityIndex,
  onRefreshEntities,
}: {
  incidentId: string;
  apiBase?: string | undefined;
  entityType: EntityRow["entityType"];
  savedViewSelector?: ReactNode | undefined;
  filterDraft: FilterDraft;
  onApplyFilter: () => void;
  onFilterDraftChange: (draft: FilterDraft) => void;
  onGroupByChange: (groupBy: string | null) => void;
  onRemoveFilter: (fieldKey: string) => void;
  onToggleSort: (fieldKey: string) => void;
  queryState: WorkbookQueryState;
  rows: EntityRow[];
  currentIncidentRole: IncidentRole | null;
  entityIndex: Record<string, EntityRow>;
  onRefreshEntities: () => Promise<void>;
}) {
  const [selectedRecordId, setSelectedRecordId] = useState<string | null>(null);
  const [mergeCandidateId, setMergeCandidateId] = useState<string>("");
  const [mergeReason, setMergeReason] = useState("Merge duplicate entity");
  const [mergeMessage, setMergeMessage] = useState<string | null>(null);
  const [editRecordId, setEditRecordId] = useState("");
  const [editFieldKey, setEditFieldKey] = useState("");
  const [editValue, setEditValue] = useState("");
  const [mutationError, setMutationError] = useState<string | null>(null);
  const [mutationState, setMutationState] = useState<SaveState>("Saved");
  const { loadTimelinePreview, timelinePreviewRows } = useEntityTimelinePreview(
    {
      apiBase,
      entityType,
      incidentId,
    },
  );

  const selectedEntity =
    rows.find((row) => row.recordId === selectedRecordId) ?? rows[0] ?? null;
  const canMerge =
    currentIncidentRole === "reviewer" || currentIncidentRole === "admin";
  const survivorLabel = selectedEntity?.label ?? "Select a record";
  const contract = entityType === "host" ? hostsContract : identitiesContract;
  const surface: WorkbookSurface = contract.viewSchemaId;
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
  const entityAnchorColumns = useMemo<readonly GridColumn<EntityRow>[]>(
    () =>
      workbookContractColumns<EntityRow>({
        contract,
        surface,
        widthForField: entityContractColumnWidth,
      }),
    [contract, surface],
  );
  const entityGridRows = buildWorkbookGridRows({
    getRecordId: (row: EntityRow) => row.recordId,
    rows,
    selectedRecordId: selectedEntity?.recordId ?? null,
    surface,
  });
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
      const result = await fetchJSON<EntityClipboardPasteEnvelope>(
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
      renderCell: (row) => (
        <FocusableWorkbookCell
          fieldKey={column.fieldKey}
          focus={entityFocus}
          onPaste={handleEntityPaste}
          recordId={row.recordId}
        >
          {entityCellContent(entityType, row, column.fieldKey)}
        </FocusableWorkbookCell>
      ),
    }));
  const entityActionsColumn: GridActionsColumn<EntityRow> = {
    headerTestId: gridActionsHeaderTestId(surface),
    label: "Actions",
    width: 176,
    renderCell: ({ data: row }) => (
      <button
        data-testid={entityInspectButtonTestId(entityType, row.recordId)}
        style={actionButtonStyle}
        type="button"
        onClick={() => {
          setSelectedRecordId(row.recordId);
          setMergeMessage(null);
        }}
      >
        Inspect
      </button>
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
    if (selectedEntity) {
      setSelectedRecordId(selectedEntity.recordId);
      return;
    }
    setSelectedRecordId(null);
  }, [selectedEntity]);

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

  async function confirmMerge() {
    if (!selectedEntity || !loserEntity) {
      return;
    }
    setMergeMessage(null);
    const result = await fetchJSON<MergeEnvelope>(
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
      return;
    }

    const envelope = readEnvelope<MergeEnvelope>(result.payload);
    setMergeMessage(
      `Merged ${loserEntity.label} into ${selectedEntity.label} (${envelope.data.merge_summary.record_type}).`,
    );
    await onRefreshEntities();
    await loadTimelinePreview(selectedEntity.recordId);
    setSelectedRecordId(selectedEntity.recordId);
    setMergeCandidateId("");
  }

  return (
    <section style={surfaceStackStyle}>
      <header style={headerStyle}>
        <div>
          <p style={eyebrowStyle}>
            {entityType === "host" ? "Hosts" : "Identities"}
          </p>
          <h1 style={headlineStyle}>
            {entityType === "host" ? "Hosts surface" : "Identities surface"}
          </h1>
          <p style={bodyStyle}>Incident {incidentId}</p>
        </div>
        <div style={roleBadgeStyle} data-testid="generic-mutation-state">
          {mutationState}
        </div>
      </header>
      <WorkbookFocusAnchorStatus anchor={entityFocus.anchor} />

      <div style={splitShellStyle}>
        <div>
          <WorkbookShellSlotRegion
            slot="view-bar"
            style={viewBarStyle}
            viewSchemaId={surface}
          >
            {savedViewSelector}
            <WorkbookGridControls
              contract={contract}
              filterDraft={filterDraft}
              onApplyFilter={onApplyFilter}
              onFilterDraftChange={onFilterDraftChange}
              onGroupByChange={onGroupByChange}
              onRemoveFilter={onRemoveFilter}
              queryState={queryState}
              surface={surface}
            />
          </WorkbookShellSlotRegion>
          {editableEntityFields.length > 0 && rows.length > 0 ? (
            <section style={genericMutationPanelStyle}>
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
              {mutationError ? <p style={bodyStyle}>{mutationError}</p> : null}
            </section>
          ) : null}
          <GridViewport
            style={gridShellStyle}
            testId={gridShellTestId(surface)}
          >
            <GridTable
              actionsColumn={entityActionsColumn}
              columns={entityColumns}
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
        </div>

        <aside
          data-testid={entityInspectorTestId(entityType)}
          style={inspectorShellStyle}
        >
          <div style={inspectorHeaderStyle}>
            <p style={eyebrowStyle}>Inspector</p>
            <h2 style={inspectorTitleStyle}>{survivorLabel}</h2>
            <p style={bodyStyle}>
              Merge review stays inside the workbook shell.
            </p>
          </div>
          {selectedEntity ? (
            <>
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

              {canMerge ? (
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
              ) : (
                <section style={inspectorSectionStyle}>
                  <h3 style={sectionTitleStyle}>Merge</h3>
                  <p style={bodyStyle}>
                    Merge is available to reviewer or admin roles.
                  </p>
                </section>
              )}

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
                          {row.values.summary || "Untitled row"}
                        </p>
                        <div style={relationshipItemsWrapStyle}>
                          {row.collectionValues[
                            entityType === "host" ? "hostRefs" : "identityRefs"
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
                <p data-testid="merge-message" style={bodyStyle}>
                  {mergeMessage}
                </p>
              ) : null}
            </>
          ) : (
            <p style={bodyStyle}>No active records on this surface.</p>
          )}
        </aside>
      </div>
    </section>
  );
}

function AssessmentWorkbookSurface({
  apiBase,
  assessmentRows,
  currentIncidentRole,
  savedViewSelector,
  filterDraft,
  hostRows,
  identityRows,
  incidentId,
  loadError,
  onApplyFilter,
  onFilterDraftChange,
  onGroupByChange,
  onRefreshAssessmentRows,
  onRemoveFilter,
  onToggleSort,
  queryState,
}: {
  apiBase?: string | undefined;
  assessmentRows: EntityApiRow[];
  currentIncidentRole: IncidentRole | null;
  savedViewSelector?: ReactNode | undefined;
  filterDraft: FilterDraft;
  hostRows: EntityRow[];
  identityRows: EntityRow[];
  incidentId: string;
  loadError: string | null;
  onApplyFilter: () => void;
  onFilterDraftChange: (draft: FilterDraft) => void;
  onGroupByChange: (groupBy: string | null) => void;
  onRefreshAssessmentRows: () => Promise<void>;
  onRemoveFilter: (fieldKey: string) => void;
  onToggleSort: (fieldKey: string) => void;
  queryState: WorkbookQueryState;
}) {
  const [draft, setDraft] = useState<AssessmentCreateDraft>(() =>
    initialAssessmentDraft(assessmentsContract),
  );
  const supportRows = useAssessmentSupportRows({ apiBase, incidentId });
  const [message, setMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const subjectRows = draft.subjectType === "host" ? hostRows : identityRows;
  const canCreate =
    currentIncidentRole === "editor" ||
    currentIncidentRole === "reviewer" ||
    currentIncidentRole === "admin";
  const stateOptions = enumValuesFor(
    assessmentsContract,
    "assessment.assessment_state",
    ["unknown", "suspected", "confirmed", "disproven", "cleared"],
  );
  const confidenceBandOptions = enumValuesFor(
    assessmentsContract,
    "assessment.confidence_band",
    ["unset", "low", "medium", "high"],
  ).filter(isAssessmentConfidenceBand);
  const anchorColumns: readonly GridColumn<EntityApiRow>[] = visibleFields(
    assessmentsContract,
  ).map((field) => ({
    fieldKey: field.fieldKey,
    headerTestId: gridSortHeaderTestId(assessmentsViewSchemaId, field.fieldKey),
    label: field.label,
    width: assessmentColumnWidth(field.fieldKey),
    renderCell: () => null,
    sortableFieldKey: resolveHeaderSortFieldKey(
      assessmentsContract,
      field.fieldKey,
    ),
  }));
  const gridRows: readonly GridRow<EntityApiRow>[] = assessmentRows.map(
    (row) => ({
      key: row.record_id,
      recordId: row.record_id,
      data: row,
      testId: gridRowTestId(assessmentsViewSchemaId, row.record_id),
    }),
  );
  const assessmentFocus = useWorkbookGridFocus({
    columns: anchorColumns,
    getGroupLabel: (row, fieldKey) =>
      genericCellLabel(row.cells[fieldKey]?.value),
    groupBy: queryState.groupBy,
    rows: gridRows,
    surface: assessmentsViewSchemaId,
  });
  const columns: readonly GridColumn<EntityApiRow>[] = anchorColumns.map(
    (field) => ({
      ...field,
      renderCell: (row) => (
        <FocusableWorkbookCell
          fieldKey={field.fieldKey}
          focus={assessmentFocus}
          recordId={row.record_id}
        >
          {genericCellLabel(row.cells[field.fieldKey]?.value)}
        </FocusableWorkbookCell>
      ),
    }),
  );

  useEffect(() => {
    setDraft((current) => {
      if (
        current.subjectRecordId !== "" &&
        subjectRows.some((row) => row.recordId === current.subjectRecordId)
      ) {
        return current;
      }
      return {
        ...current,
        subjectRecordId: subjectRows[0]?.recordId ?? "",
      };
    });
  }, [subjectRows]);

  async function submitAssessment() {
    const payload = buildAssessmentCreatePayload(
      draft,
      `assessment-${Date.now()}`,
    );
    if (payload === null) {
      setMessage("Subject, state, and rationale are required.");
      return;
    }

    setIsSubmitting(true);
    setMessage(null);
    try {
      const result = await fetchJSON<TimelineMutationEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${assessmentsViewSchemaId}/rows`,
        ),
        {
          method: "POST",
          body: JSON.stringify(payload),
        },
      );
      if (!result.ok) {
        setMessage(parseErrorMessage(result.payload));
        return;
      }
      await onRefreshAssessmentRows();
      setDraft((current) => ({
        ...initialAssessmentDraft(assessmentsContract),
        subjectType: current.subjectType,
        subjectRecordId: current.subjectRecordId,
      }));
      setMessage("Assessment created.");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <section style={surfaceStackStyle}>
      <header style={headerStyle}>
        <div>
          <p style={eyebrowStyle}>System view</p>
          <h1 style={headlineStyle}>{assessmentsContract.title}</h1>
          <p style={bodyStyle}>Incident {incidentId}</p>
        </div>
      </header>
      <WorkbookFocusAnchorStatus anchor={assessmentFocus.anchor} />

      <div style={splitShellStyle}>
        <div>
          <WorkbookShellSlotRegion
            slot="view-bar"
            style={viewBarStyle}
            viewSchemaId={assessmentsViewSchemaId}
          >
            {savedViewSelector}
            <WorkbookGridControls
              contract={assessmentsContract}
              filterDraft={filterDraft}
              onApplyFilter={onApplyFilter}
              onFilterDraftChange={onFilterDraftChange}
              onGroupByChange={onGroupByChange}
              onRemoveFilter={onRemoveFilter}
              queryState={queryState}
              surface={assessmentsViewSchemaId}
            />
          </WorkbookShellSlotRegion>
          {loadError ? (
            <p data-testid="assessment-surface-load-error" style={bodyStyle}>
              {loadError}
            </p>
          ) : null}
          <GridViewport
            style={gridShellStyle}
            testId={gridShellTestId(assessmentsViewSchemaId)}
          >
            <GridTable
              columns={columns}
              getGroupLabel={(row, fieldKey) =>
                genericCellLabel(row.cells[fieldKey]?.value)
              }
              getGroupRowTestId={(fieldKey, value) =>
                gridGroupRowTestId(assessmentsViewSchemaId, fieldKey, value)
              }
              groupBy={queryState.groupBy}
              onToggleSort={onToggleSort}
              rows={gridRows}
              sort={queryState.sort}
            />
          </GridViewport>
        </div>

        <aside
          data-testid={assessmentCreatePanelTestId()}
          style={inspectorShellStyle}
        >
          <div style={inspectorHeaderStyle}>
            <p style={eyebrowStyle}>Create</p>
            <h2 style={inspectorTitleStyle}>Append assessment</h2>
          </div>
          <div style={inspectorSectionStyle}>
            <label style={labelStyle}>
              Subject type
              <select
                data-testid="assessment-create-subject-type"
                style={selectStyle}
                value={draft.subjectType}
                onChange={(event) => {
                  const subjectType =
                    event.target.value === "identity" ? "identity" : "host";
                  const nextRows =
                    subjectType === "host" ? hostRows : identityRows;
                  setDraft((current) => ({
                    ...current,
                    subjectType,
                    subjectRecordId: nextRows[0]?.recordId ?? "",
                  }));
                }}
              >
                {enumValuesFor(assessmentsContract, "assessment.subject_type", [
                  "host",
                  "identity",
                ]).map((value) => (
                  <option key={value} value={value}>
                    {value}
                  </option>
                ))}
              </select>
            </label>

            <label style={labelStyle}>
              Subject
              <select
                data-testid="assessment-create-subject"
                style={selectStyle}
                value={draft.subjectRecordId}
                onChange={(event) => {
                  setDraft((current) => ({
                    ...current,
                    subjectRecordId: event.target.value,
                  }));
                }}
              >
                <option value="">Select subject</option>
                {subjectRows.map((row) => (
                  <option key={row.recordId} value={row.recordId}>
                    {row.label}
                  </option>
                ))}
              </select>
            </label>

            <label style={labelStyle}>
              State
              <select
                data-testid="assessment-create-state"
                style={selectStyle}
                value={draft.assessmentState}
                onChange={(event) => {
                  setDraft((current) => ({
                    ...current,
                    assessmentState: event.target.value,
                  }));
                }}
              >
                {stateOptions.map((value) => (
                  <option key={value} value={value}>
                    {value}
                  </option>
                ))}
              </select>
            </label>

            <label style={labelStyle}>
              Confidence
              <select
                data-testid="assessment-create-confidence-band"
                style={selectStyle}
                value={draft.confidenceBand}
                onChange={(event) => {
                  const confidenceBand = isAssessmentConfidenceBand(
                    event.target.value,
                  )
                    ? event.target.value
                    : "unset";
                  setDraft((current) => ({
                    ...current,
                    confidenceBand,
                  }));
                }}
              >
                {confidenceBandOptions.map((value) => (
                  <option key={value} value={value}>
                    {value}
                  </option>
                ))}
              </select>
            </label>

            <label style={labelStyle}>
              Rationale
              <textarea
                data-testid="assessment-create-rationale"
                rows={4}
                style={textareaStyle}
                value={draft.rationale}
                onChange={(event) => {
                  setDraft((current) => ({
                    ...current,
                    rationale: event.target.value,
                  }));
                }}
              />
            </label>

            <label style={labelStyle}>
              Assessed
              <input
                data-testid="assessment-create-assessed-at"
                placeholder="RFC3339 timestamp"
                style={inputStyle}
                type="text"
                value={draft.assessedAt}
                onChange={(event) => {
                  setDraft((current) => ({
                    ...current,
                    assessedAt: event.target.value,
                  }));
                }}
              />
            </label>

            <label style={labelStyle}>
              Support refs
              <select
                data-testid="assessment-create-support-refs"
                multiple
                size={Math.min(Math.max(supportRows.length, 2), 5)}
                style={selectStyle}
                value={draft.supportRecordIds}
                onChange={(event) => {
                  const supportRecordIds = Array.from(
                    event.currentTarget.selectedOptions,
                  ).map((option) => option.value);
                  setDraft((current) => ({
                    ...current,
                    supportRecordIds,
                  }));
                }}
              >
                {supportRows.map((row) => (
                  <option key={row.record_id} value={row.record_id}>
                    {supportRowLabel(row)}
                  </option>
                ))}
              </select>
            </label>

            <button
              data-testid="assessment-create-submit"
              disabled={!canCreate || isSubmitting}
              style={secondaryActionButtonStyle}
              type="button"
              onClick={() => {
                void submitAssessment();
              }}
            >
              Create assessment
            </button>
            {message ? (
              <p data-testid="assessment-create-message" style={bodyStyle}>
                {message}
              </p>
            ) : null}
          </div>
        </aside>
      </div>
    </section>
  );
}

function GenericWorkbookSurface({
  apiBase,
  contract,
  currentUserId,
  savedViewSelector,
  filterDraft,
  incidentId,
  loadError,
  onApplyFilter,
  onFilterDraftChange,
  onGroupByChange,
  onRemoveFilter,
  onRefresh,
  onToggleSort,
  queryState,
  rows,
}: {
  apiBase?: string | undefined;
  contract: ViewContract;
  currentUserId: string | null;
  savedViewSelector?: ReactNode | undefined;
  filterDraft: FilterDraft;
  incidentId: string;
  loadError: string | null;
  onApplyFilter: () => void;
  onFilterDraftChange: (draft: FilterDraft) => void;
  onGroupByChange: (groupBy: string | null) => void;
  onRemoveFilter: (fieldKey: string) => void;
  onRefresh: () => Promise<void> | void;
  onToggleSort: (fieldKey: string) => void;
  queryState: WorkbookQueryState;
  rows: EntityApiRow[];
}) {
  const surface = contract.viewSchemaId as WorkbookSurface;
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
  const [mutationState, setMutationState] = useState<SaveState>("Saved");
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
      const handleRequest = {} satisfies EvidenceHandleIssueRequest;
      const result = await fetchJSON<EvidenceHandleEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/evidence-records/${row.record_id}/${kind}-handle`,
        ),
        { method: "POST", body: JSON.stringify(handleRequest) },
      );
      if (!result.ok) {
        setEvidenceMessage(
          row.record_id,
          evidencePublicErrorMessage(result.payload, "Evidence access failed."),
        );
        return;
      }
      const envelope = readEnvelope<EvidenceHandleEnvelope>(result.payload);
      const href = resolvePublicEvidenceHandleHref(envelope.data.href);
      if (href === null) {
        setEvidenceMessage(row.record_id, "Evidence handle is unavailable.");
        return;
      }
      if (kind === "preview") {
        setEvidencePreview({
          href,
          recordId: row.record_id,
          title:
            stringifyGridValue(row.cells["evidence.title"]?.value).trim() ||
            row.record_id,
          previewKind: envelope.data.preview_kind ?? null,
        });
        setEvidenceMessage(row.record_id, "Preview loaded inline.");
        return;
      }

      const anchor = document.createElement("a");
      anchor.href = href;
      anchor.download = envelope.data.filename || "evidence";
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

  const anchorColumns = useMemo<readonly GridColumn<EntityApiRow>[]>(
    () =>
      workbookContractColumns<EntityApiRow>({
        contract,
        surface,
        widthForField: genericContractColumnWidth,
      }),
    [contract, surface],
  );
  const gridRows = buildWorkbookGridRows({
    getRecordId: (row: EntityApiRow) => row.record_id,
    rows,
    surface,
  });
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
      renderCell: (row) => (
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
      ),
    }),
  );
  const evidenceActionsColumn = useMemo<
    GridActionsColumn<EntityApiRow> | undefined
  >(() => {
    if (!isEvidenceSurface) {
      return undefined;
    }
    return {
      headerTestId: gridActionsHeaderTestId(surface),
      label: "Access",
      width: 208,
      renderCell: ({ data: row }) => {
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
    evidenceMessageByRecordID,
    isEvidenceSurface,
    issueEvidenceHandle,
    surface,
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

  const completeGenericMutation = async <TEnvelope,>(payload: unknown) => {
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
  };

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

  const submitCreate = async () => {
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
  };

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
    const payload = await submitWorkbookPatchMutation({
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
    const payload = await submitWorkbookPatchMutation({
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

  return (
    <section style={surfaceStackStyle}>
      <header style={headerStyle}>
        <div>
          <p style={eyebrowStyle}>
            {contract.surfaceKind === "built_in_sheet"
              ? "Built-in sheet"
              : "System view"}
          </p>
          <h1 style={headlineStyle}>{contract.title}</h1>
          <p style={bodyStyle}>Incident {incidentId}</p>
        </div>
        <div style={roleBadgeStyle} data-testid="generic-mutation-state">
          {mutationState}
        </div>
      </header>
      <WorkbookFocusAnchorStatus anchor={genericFocus.anchor} />

      {writableFields.length > 0 ? (
        <section style={genericMutationPanelStyle}>
          <div style={genericFormGridStyle}>
            {writableFields.map((field) => (
              <label
                key={field.fieldKey}
                htmlFor={`generic-create-input-${field.fieldKey}`}
                style={labelStyle}
              >
                {field.label}
                <GenericMutationControl
                  collectionMode="add"
                  field={field}
                  id={`generic-create-input-${field.fieldKey}`}
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
            ))}
            {isNotesSurface ? (
              <label
                htmlFor="generic-create-note-source-record"
                style={labelStyle}
              >
                Linked source
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
          </div>
          <button
            data-testid={genericCreateSubmitTestId(contract.viewSchemaId)}
            disabled={mutationState === "Syncing"}
            style={actionButtonStyle}
            type="button"
            onClick={() => {
              void submitCreate();
            }}
          >
            Create
          </button>

          {rows.length > 0 && selectedEditField !== null ? (
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

          {partyLinkPairs.length > 0 && selectedEditRow !== null ? (
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

          {isTaskRequestSurface && rows.length > 0 ? (
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

          {isDecisionSurface && rows.length > 1 ? (
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
            <p
              data-testid="generic-mutation-error"
              style={genericErrorTextStyle}
            >
              {mutationError}
            </p>
          ) : null}
        </section>
      ) : null}

      <WorkbookShellSlotRegion
        slot="view-bar"
        style={viewBarStyle}
        viewSchemaId={surface}
      >
        {savedViewSelector}
        <WorkbookGridControls
          contract={contract}
          filterDraft={filterDraft}
          onApplyFilter={onApplyFilter}
          onFilterDraftChange={onFilterDraftChange}
          onGroupByChange={onGroupByChange}
          onRemoveFilter={onRemoveFilter}
          queryState={queryState}
          surface={surface}
        />
      </WorkbookShellSlotRegion>

      {loadError ? (
        <p data-testid="generic-surface-load-error" style={bodyStyle}>
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
            data-testid={evidencePreviewFrameTestId(evidencePreview.recordId)}
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

      <GridViewport style={gridShellStyle} testId={gridShellTestId(surface)}>
        <GridTable
          actionsColumn={evidenceActionsColumn}
          columns={columns}
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

      <WorkbookShellSlotRegion
        slot="status-strip"
        style={statusStripStyle}
        viewSchemaId={surface}
      >
        <span style={statusStripItemStyle}>
          <span aria-hidden="true" style={statusIconStyle(mutationState)} />
          <strong
            aria-live="polite"
            aria-label="Save state"
            data-density-role="narrow-metadata"
            data-testid={saveStateTestId()}
            role="status"
          >
            {mutationState}
          </strong>
        </span>
      </WorkbookShellSlotRegion>
    </section>
  );
}

export function WorkbookShell({
  incidentId,
  apiBase,
  currentUserLabel,
  onIncidentSnapshot,
  onIncidentAccessLost,
  onReturnToLanding,
}: WorkbookShellProps) {
  const params = useMemo(() => new URLSearchParams(window.location.search), []);
  const initialViewSchemaID = useMemo(() => {
    const explicit = params.get("view_schema_id");
    return explicit
      ? knownWorkbookViewSchemaId(explicit)
      : timelineViewSchemaId;
  }, [params]);
  const surfaceSelectionVersionRef = useRef(0);
  const workbookRuntime = useWorkbookShellRuntime({
    initialViewSchemaId: initialViewSchemaID,
    surfaceSelectionVersionRef,
  });
  const {
    pendingGridFocusSurface,
    savedViews,
    sheetReloadToken,
    startupSheetRef,
    surface,
  } = workbookRuntime.snapshot;
  const {
    applyStartupIdentity,
    deleteSavedViewIdentity,
    replaceSavedViews,
    selectSavedViewIdentity,
    selectWorkbookSurface,
    setPendingGridFocusSurface,
    upsertSavedView,
  } = workbookRuntime.commands;
  const [hostRows, setHostRows] = useState<EntityRow[]>([]);
  const [identityRows, setIdentityRows] = useState<EntityRow[]>([]);
  const [entityLoadError, setEntityLoadError] = useState<string | null>(null);
  const [genericRows, setGenericRows] = useState<EntityApiRow[]>([]);
  const [genericLoadError, setGenericLoadError] = useState<string | null>(null);
  const [assessmentRows, setAssessmentRows] = useState<EntityApiRow[]>([]);
  const [assessmentLoadError, setAssessmentLoadError] = useState<string | null>(
    null,
  );
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);
  const [currentIncidentRole, setCurrentIncidentRole] =
    useState<IncidentRole | null>(null);
  const [incidentControlsOpen, setIncidentControlsOpen] = useState(false);
  const [timelineQueryState, setTimelineQueryState] =
    useState<WorkbookQueryState>(() => emptyWorkbookQueryState());
  const [timelineFilterDraft, setTimelineFilterDraft] = useState<FilterDraft>(
    () => defaultFilterDraft(timelineContract),
  );
  const [hostQueryState, setHostQueryState] = useState<WorkbookQueryState>(() =>
    emptyWorkbookQueryState(),
  );
  const [identityQueryState, setIdentityQueryState] =
    useState<WorkbookQueryState>(() => emptyWorkbookQueryState());
  const [hostFilterDraft, setHostFilterDraft] = useState<FilterDraft>(() =>
    defaultFilterDraft(hostsContract),
  );
  const [identityFilterDraft, setIdentityFilterDraft] = useState<FilterDraft>(
    () => defaultFilterDraft(identitiesContract),
  );
  const [assessmentQueryState, setAssessmentQueryState] =
    useState<WorkbookQueryState>(() => emptyWorkbookQueryState());
  const [assessmentFilterDraft, setAssessmentFilterDraft] =
    useState<FilterDraft>(() => defaultFilterDraft(assessmentsContract));
  const activeContract = useMemo(
    () =>
      allWorkbookContracts.find(
        (contract) => contract.viewSchemaId === surface,
      ) ?? timelineContract,
    [surface],
  );
  const [genericQueryState, setGenericQueryState] =
    useState<WorkbookQueryState>(() => emptyWorkbookQueryState());
  const [genericFilterDraft, setGenericFilterDraft] = useState<FilterDraft>(
    () => defaultFilterDraft(activeContract),
  );
  const entityQueryRuntimeRef = useRef<LatestQueryRuntime>({
    controller: null,
    sequence: 0,
  });
  const assessmentQueryRuntimeRef = useRef<LatestQueryRuntime>({
    controller: null,
    sequence: 0,
  });
  const genericQueryRuntimeRef = useRef<LatestQueryRuntime>({
    controller: null,
    sequence: 0,
  });
  const currentQueryStateForSurface = useCallback(
    (viewSchemaId: string): WorkbookQueryState => {
      if (viewSchemaId === timelineViewSchemaId) {
        return timelineQueryState;
      }
      if (viewSchemaId === hostsViewSchemaId) {
        return hostQueryState;
      }
      if (viewSchemaId === identitiesViewSchemaId) {
        return identityQueryState;
      }
      if (viewSchemaId === assessmentsViewSchemaId) {
        return assessmentQueryState;
      }
      return genericQueryState;
    },
    [
      assessmentQueryState,
      genericQueryState,
      hostQueryState,
      identityQueryState,
      timelineQueryState,
    ],
  );
  const applyQueryStateForSurface = useCallback(
    (viewSchemaId: string, queryState: WorkbookQueryState) => {
      const contract = workbookContractForViewSchemaId(viewSchemaId);
      if (viewSchemaId === timelineViewSchemaId) {
        setTimelineQueryState(queryState);
        setTimelineFilterDraft(defaultFilterDraft(timelineContract));
        return;
      }
      if (viewSchemaId === hostsViewSchemaId) {
        setHostQueryState(queryState);
        setHostFilterDraft(defaultFilterDraft(hostsContract));
        return;
      }
      if (viewSchemaId === identitiesViewSchemaId) {
        setIdentityQueryState(queryState);
        setIdentityFilterDraft(defaultFilterDraft(identitiesContract));
        return;
      }
      if (viewSchemaId === assessmentsViewSchemaId) {
        setAssessmentQueryState(queryState);
        setAssessmentFilterDraft(defaultFilterDraft(assessmentsContract));
        return;
      }
      setGenericQueryState(queryState);
      setGenericFilterDraft(defaultFilterDraft(contract));
    },
    [],
  );
  const selectSavedView = useCallback(
    (savedView: SavedViewResource) => {
      const nextSurface = knownWorkbookViewSchemaId(savedView.view_schema_id);
      const contract = workbookContractForViewSchemaId(nextSurface);
      applyQueryStateForSurface(
        nextSurface,
        savedViewQueryStateForRuntime(contract, savedView),
      );
      selectSavedViewIdentity(savedView);
    },
    [applyQueryStateForSurface, selectSavedViewIdentity],
  );

  const createSavedView = useCallback(
    async (input: {
      readonly displayName: string;
      readonly scope: "private" | "shared";
    }) => {
      const contract = activeContract;
      const queryState = currentQueryStateForSurface(contract.viewSchemaId);
      const result = await fetchJSON<SavedViewEnvelope>(
        apiPath(apiBase, `/api/v1/incidents/${incidentId}/saved-views`),
        {
          method: "POST",
          body: JSON.stringify({
            view_schema_id: contract.viewSchemaId,
            display_name: input.displayName,
            scope: input.scope,
            query_json: buildSavedViewQueryJson(contract, queryState),
            layout_json: buildSavedViewLayoutJson(contract),
          }),
        },
      );
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const savedView = normalizeSavedViewResource(
        readEnvelope<SavedViewEnvelope>(result.payload).data,
      );
      if (savedView === null) {
        throw new Error("Saved-view create returned an invalid resource.");
      }
      upsertSavedView(savedView);
      selectSavedView(savedView);
      return savedView;
    },
    [
      activeContract,
      apiBase,
      currentQueryStateForSurface,
      incidentId,
      selectSavedView,
      upsertSavedView,
    ],
  );

  const duplicateSavedView = useCallback(
    async (source: SavedViewResource) => {
      const contract = workbookContractForViewSchemaId(source.view_schema_id);
      const result = await fetchJSON<SavedViewEnvelope>(
        apiPath(apiBase, `/api/v1/incidents/${incidentId}/saved-views`),
        {
          method: "POST",
          body: JSON.stringify({
            view_schema_id: source.view_schema_id,
            display_name: `${source.display_name} Copy`,
            scope: "private",
            query_json: savedViewQueryJsonForPersistence(
              contract,
              source.query_json,
            ),
            layout_json: savedViewLayoutJsonForPersistence(
              contract,
              source.layout_json,
            ),
          }),
        },
      );
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const savedView = normalizeSavedViewResource(
        readEnvelope<SavedViewEnvelope>(result.payload).data,
      );
      if (savedView === null) {
        throw new Error("Saved-view duplicate returned an invalid resource.");
      }
      upsertSavedView(savedView);
      selectSavedView(savedView);
      return savedView;
    },
    [apiBase, incidentId, selectSavedView, upsertSavedView],
  );

  const updateSavedView = useCallback(
    async (
      savedView: SavedViewResource,
      input: {
        readonly displayName: string;
        readonly scope: "private" | "shared";
      },
    ) => {
      const contract = workbookContractForViewSchemaId(
        savedView.view_schema_id,
      );
      const queryState = currentQueryStateForSurface(savedView.view_schema_id);
      const result = await fetchJSON<SavedViewEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/saved-views/${savedView.saved_view_id}`,
        ),
        {
          method: "PATCH",
          body: JSON.stringify({
            base_saved_view_version: savedView.saved_view_version,
            display_name: input.displayName,
            scope: input.scope,
            query_json: buildSavedViewQueryJson(contract, queryState),
            layout_json: buildSavedViewLayoutJson(contract),
          }),
        },
      );
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const updated = normalizeSavedViewResource(
        readEnvelope<SavedViewEnvelope>(result.payload).data,
      );
      if (updated === null) {
        throw new Error("Saved-view update returned an invalid resource.");
      }
      upsertSavedView(updated);
      return updated;
    },
    [apiBase, currentQueryStateForSurface, incidentId, upsertSavedView],
  );

  const deleteSavedView = useCallback(
    async (savedView: SavedViewResource) => {
      const result = await fetchJSON<Record<string, unknown>>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/saved-views/${savedView.saved_view_id}`,
        ),
        { method: "DELETE" },
      );
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      deleteSavedViewIdentity(savedView, startupSheetRef);
    },
    [apiBase, deleteSavedViewIdentity, incidentId, startupSheetRef],
  );

  const setWorkbookHomeSheetRef = useCallback(async () => {
    const result = await fetchJSON<Record<string, unknown>>(
      apiPath(
        apiBase,
        `/api/v1/incidents/${incidentId}/workbook-preferences/me`,
      ),
      {
        method: "PUT",
        body: JSON.stringify({ home_sheet_ref: startupSheetRef }),
      },
    );
    if (!result.ok) {
      throw new Error(parseErrorMessage(result.payload));
    }
  }, [apiBase, incidentId, startupSheetRef]);

  const setWorkbookDefaultSheetRef = useCallback(async () => {
    const result = await fetchJSON<Record<string, unknown>>(
      apiPath(
        apiBase,
        `/api/v1/incidents/${incidentId}/workbook-preferences/default`,
      ),
      {
        method: "PUT",
        body: JSON.stringify({ default_sheet_ref: startupSheetRef }),
      },
    );
    if (!result.ok) {
      throw new Error(parseErrorMessage(result.payload));
    }
  }, [apiBase, incidentId, startupSheetRef]);

  const entityIndex = useMemo(() => {
    const index: Record<string, EntityRow> = {};
    for (const row of [...hostRows, ...identityRows]) {
      index[row.recordId] = row;
    }
    return index;
  }, [hostRows, identityRows]);

  const loadSessionRole = useCallback(async () => {
    const result = await fetchJSON<SessionEnvelope>(
      apiPath(apiBase, "/api/v1/auth/session"),
    );
    if (!result.ok) {
      setCurrentUserId(null);
      setCurrentIncidentRole("");
      onIncidentAccessLost?.();
      return;
    }
    const envelope = readEnvelope<SessionEnvelope>(result.payload);
    setCurrentUserId(envelope.data.user_id || null);
    const membership =
      envelope.data.memberships.find(
        (entry) => entry.incident_id === incidentId,
      ) ?? null;
    if (membership === null) {
      onIncidentAccessLost?.();
    }
    setCurrentIncidentRole(membership?.role ?? "");
  }, [apiBase, incidentId, onIncidentAccessLost]);

  const queryEntityView = useCallback(
    async (
      viewSchemaId: string,
      entityType: EntityRow["entityType"],
      queryState: WorkbookQueryState,
      signal: AbortSignal,
    ) => {
      const contract =
        viewSchemaId === hostsViewSchemaId ? hostsContract : identitiesContract;
      const result = await fetchJSON<ViewQueryEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${viewSchemaId}/query`,
        ),
        {
          method: "POST",
          signal,
          body: JSON.stringify(buildQueryRequest(contract, queryState)),
        },
      );
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const envelope = readEnvelope<ViewQueryEnvelope>(result.payload);
      if (envelope.data.view_schema_id !== viewSchemaId) {
        throw new Error(
          `Entity surface load returned ${envelope.data.view_schema_id} for ${viewSchemaId}.`,
        );
      }
      return normalizeWorkbookViewRows(
        contract,
        envelope.data.rows,
        `${viewSchemaId} query response`,
      ).map((row) => entityRowFromApi(row, entityType));
    },
    [apiBase, incidentId],
  );

  const loadEntities = useCallback(async () => {
    const request = beginLatestQuery(entityQueryRuntimeRef);
    setEntityLoadError(null);
    try {
      const [nextHosts, nextIdentities] = await Promise.all([
        queryEntityView(
          hostsViewSchemaId,
          "host",
          hostQueryState,
          request.signal,
        ),
        queryEntityView(
          identitiesViewSchemaId,
          "identity",
          identityQueryState,
          request.signal,
        ),
      ]);
      if (!request.isCurrent()) {
        return;
      }
      setHostRows((current) => [...reconcileRecordRows(current, nextHosts)]);
      setIdentityRows((current) => [
        ...reconcileRecordRows(current, nextIdentities),
      ]);
    } catch (error) {
      if (!request.isCurrent() || isAbortError(error)) {
        return;
      }
      const message = handleWorkbookLoadFailure(
        error,
        "Entity load failed.",
        onIncidentAccessLost,
      );
      setEntityLoadError(message);
    }
  }, [
    hostQueryState,
    identityQueryState,
    onIncidentAccessLost,
    queryEntityView,
  ]);

  const applyHostFilter = useCallback(() => {
    applyFilterDraftToQuery(
      setHostQueryState,
      setHostFilterDraft,
      hostFilterDraft,
    );
  }, [hostFilterDraft]);

  const applyIdentityFilter = useCallback(() => {
    applyFilterDraftToQuery(
      setIdentityQueryState,
      setIdentityFilterDraft,
      identityFilterDraft,
    );
  }, [identityFilterDraft]);

  const isSpecializedSurface =
    surface === timelineViewSchemaId ||
    surface === hostsViewSchemaId ||
    surface === identitiesViewSchemaId ||
    surface === assessmentsViewSchemaId;

  const loadAssessmentSurface = useCallback(async () => {
    if (surface !== assessmentsViewSchemaId) {
      abortLatestQuery(assessmentQueryRuntimeRef);
      return;
    }
    const request = beginLatestQuery(assessmentQueryRuntimeRef);
    setAssessmentLoadError(null);
    try {
      const result = await fetchJSON<ViewQueryEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${assessmentsViewSchemaId}/query`,
        ),
        {
          method: "POST",
          signal: request.signal,
          body: JSON.stringify(
            buildQueryRequest(assessmentsContract, assessmentQueryState),
          ),
        },
      );
      if (!request.isCurrent()) {
        return;
      }
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const envelope = readEnvelope<ViewQueryEnvelope>(result.payload);
      setAssessmentRows(envelope.data.rows);
    } catch (error) {
      if (!request.isCurrent() || isAbortError(error)) {
        return;
      }
      const message = handleWorkbookLoadFailure(
        error,
        "Assessment load failed.",
        onIncidentAccessLost,
      );
      setAssessmentLoadError(message);
      setAssessmentRows([]);
    }
  }, [
    apiBase,
    assessmentQueryState,
    incidentId,
    onIncidentAccessLost,
    surface,
  ]);

  const loadGenericSurface = useCallback(async () => {
    if (isSpecializedSurface) {
      abortLatestQuery(genericQueryRuntimeRef);
      return;
    }
    const requestedSurface = surface;
    const request = beginLatestQuery(genericQueryRuntimeRef);
    setGenericLoadError(null);
    try {
      const result = await fetchJSON<ViewQueryEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${requestedSurface}/query`,
        ),
        {
          method: "POST",
          signal: request.signal,
          body: JSON.stringify(
            buildQueryRequest(activeContract, genericQueryState),
          ),
        },
      );
      if (!request.isCurrent()) {
        return;
      }
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const envelope = readEnvelope<ViewQueryEnvelope>(result.payload);
      if (envelope.data.view_schema_id !== requestedSurface) {
        throw new Error(
          `Surface load returned ${envelope.data.view_schema_id} for ${requestedSurface}.`,
        );
      }
      setGenericRows(
        requestedSurface === notesViewSchemaId
          ? normalizeWorkbookViewRows(
              activeContract,
              envelope.data.rows,
              `${requestedSurface} query response`,
            )
          : envelope.data.rows,
      );
    } catch (error) {
      if (!request.isCurrent() || isAbortError(error)) {
        return;
      }
      const message = handleWorkbookLoadFailure(
        error,
        "Surface load failed.",
        onIncidentAccessLost,
      );
      setGenericLoadError(message);
      setGenericRows([]);
    }
  }, [
    activeContract,
    apiBase,
    genericQueryState,
    incidentId,
    isSpecializedSurface,
    onIncidentAccessLost,
    surface,
  ]);

  useEffect(
    () => () => {
      abortLatestQuery(entityQueryRuntimeRef);
      abortLatestQuery(assessmentQueryRuntimeRef);
      abortLatestQuery(genericQueryRuntimeRef);
    },
    [],
  );

  useEffect(() => {
    let cancelled = false;
    const startupQuery = workbookStartupQueryFromURLParams(params);
    const selectionVersionAtRequest = surfaceSelectionVersionRef.current;
    const loadStartup = async () => {
      const result = await fetchJSON<WorkbookStartupEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/workbook-startup${startupQuery}`,
        ),
      );
      if (cancelled || !result.ok) {
        return;
      }
      const envelope = readEnvelope<WorkbookStartupEnvelope>(result.payload);
      const startup = normalizeWorkbookStartupSelection(envelope.data);
      if (!startup) {
        return;
      }
      if (selectionVersionAtRequest !== surfaceSelectionVersionRef.current) {
        return;
      }
      const nextSurface = knownWorkbookViewSchemaId(
        startup.selectedViewSchemaId,
      );
      const startupSavedView = normalizeSavedViewResource(
        startup.selectedSavedView,
      );
      if (
        startup.selectedSheetRef.kind === "saved_view" &&
        startupSavedView !== null &&
        startupSavedView.saved_view_id === startup.selectedSheetRef.id
      ) {
        const contract = workbookContractForViewSchemaId(nextSurface);
        upsertSavedView(startupSavedView);
        applyQueryStateForSurface(
          nextSurface,
          savedViewQueryStateForRuntime(contract, startupSavedView),
        );
      }
      applyStartupIdentity({
        sheetRef: startup.selectedSheetRef,
        viewSchemaId: nextSurface,
      });
    };
    void loadStartup();
    return () => {
      cancelled = true;
    };
  }, [
    apiBase,
    applyQueryStateForSurface,
    applyStartupIdentity,
    incidentId,
    params,
    upsertSavedView,
  ]);

  useEffect(() => {
    let cancelled = false;
    const nextSavedViews: SavedViewResource[] = [];
    const loadSavedViews = async () => {
      let cursorToken: string | null = null;
      do {
        const query = new URLSearchParams({ limit: "100" });
        if (cursorToken !== null) {
          query.set("cursor_token", cursorToken);
        }
        const result = await fetchJSON<SavedViewListEnvelope>(
          apiPath(
            apiBase,
            `/api/v1/incidents/${incidentId}/saved-views?${query.toString()}`,
          ),
        );
        if (cancelled) {
          return;
        }
        if (!result.ok) {
          handleWorkbookLoadFailure(
            parseErrorMessage(result.payload),
            "Saved views load failed.",
            onIncidentAccessLost,
          );
          replaceSavedViews([]);
          return;
        }

        const envelope = readEnvelope<SavedViewListEnvelope>(result.payload);
        for (const savedView of envelope.data.saved_views) {
          const normalized = normalizeSavedViewResource(savedView);
          if (normalized !== null) {
            nextSavedViews.push(normalized);
          }
        }
        const paging = envelope.meta?.paging;
        cursorToken =
          paging?.has_more === true && paging.next_cursor
            ? paging.next_cursor
            : null;
      } while (cursorToken !== null);

      if (!cancelled) {
        replaceSavedViews(nextSavedViews);
      }
    };

    void loadSavedViews();
    return () => {
      cancelled = true;
    };
  }, [apiBase, incidentId, onIncidentAccessLost, replaceSavedViews]);

  useEffect(() => {
    void Promise.all([loadEntities(), loadSessionRole()]);
  }, [loadEntities, loadSessionRole]);

  useEffect(() => {
    if (sheetReloadToken === 0) {
      return;
    }
    void loadEntities();
  }, [loadEntities, sheetReloadToken]);

  useEffect(() => {
    if (startupSheetRef.kind === "saved_view") {
      return;
    }
    setGenericQueryState(emptyWorkbookQueryState());
    setGenericFilterDraft(defaultFilterDraft(activeContract));
    setGenericRows([]);
    setGenericLoadError(null);
  }, [activeContract, startupSheetRef.kind]);

  useEffect(() => {
    void sheetReloadToken;
    void loadGenericSurface();
  }, [loadGenericSurface, sheetReloadToken]);

  useEffect(() => {
    void sheetReloadToken;
    void loadAssessmentSurface();
  }, [loadAssessmentSurface, sheetReloadToken]);

  useEffect(() => {
    const next = new URLSearchParams(window.location.search);
    next.set("incident_id", incidentId);
    if (startupSheetRef.kind === "saved_view") {
      next.delete("view_schema_id");
      next.set("sheet_ref_kind", startupSheetRef.kind);
      next.set("sheet_ref_id", startupSheetRef.id);
    } else {
      next.set("view_schema_id", surface);
      next.delete("sheet_ref_kind");
      next.delete("sheet_ref_id");
    }
    next.delete("surface");
    window.history.replaceState({}, "", `/?${next.toString()}`);
  }, [incidentId, startupSheetRef, surface]);

  useEffect(() => {
    if (
      pendingGridFocusSurface === null ||
      pendingGridFocusSurface !== surface
    ) {
      return;
    }

    let cancelled = false;
    let timer: number | null = null;
    let attempt = 0;
    const focusFirstTarget = () => {
      if (cancelled) {
        return;
      }
      const gridShell = document.querySelector<HTMLElement>(
        dataTestIdSelector(gridShellTestId(pendingGridFocusSurface)),
      );
      const focusTarget = gridShell?.querySelector<HTMLElement>(
        '[role="row"][data-grid-record-id] [role="gridcell"] [data-testid][tabindex="0"], [role="row"][data-grid-record-id] [role="gridcell"] button:not([disabled]), [role="row"][data-grid-record-id] [role="gridcell"] input:not([disabled]), [role="row"][data-grid-record-id] [role="gridcell"] select:not([disabled]), [role="row"][data-grid-record-id] [role="gridcell"] textarea:not([disabled]), [role="row"][data-grid-record-id] [role="gridcell"] a[href]',
      );
      if (focusTarget) {
        focusTarget.focus({ preventScroll: true });
        setPendingGridFocusSurface((current) =>
          current === pendingGridFocusSurface ? null : current,
        );
        return;
      }
      attempt += 1;
      if (attempt < 30) {
        timer = window.setTimeout(focusFirstTarget, 50);
      }
    };

    timer = window.setTimeout(focusFirstTarget, 0);
    return () => {
      cancelled = true;
      if (timer !== null) {
        window.clearTimeout(timer);
      }
    };
  }, [pendingGridFocusSurface, setPendingGridFocusSurface, surface]);

  const activeSavedViewSelector = (
    <ActiveSurfaceSavedViewSelector
      activeViewSchemaId={surface}
      currentIncidentRole={currentIncidentRole}
      currentUserId={currentUserId}
      onCreateSavedView={createSavedView}
      onDeleteSavedView={deleteSavedView}
      onDuplicateSavedView={duplicateSavedView}
      onSelectBaseSurface={selectWorkbookSurface}
      onSelectSavedView={selectSavedView}
      onSetDefaultSheetRef={setWorkbookDefaultSheetRef}
      onSetHomeSheetRef={setWorkbookHomeSheetRef}
      onUpdateSavedView={updateSavedView}
      savedViews={savedViews}
      selectedSheetRef={startupSheetRef}
    />
  );

  return (
    <section
      aria-label="Workbook shell"
      data-active-view-schema-id={surface}
      data-testid={workbookShellReadyTestId()}
      data-workbook-shell-id={workbookShellId}
      style={panelStyle}
    >
      <WorkbookShellSlotRegion
        slot="top-bar"
        style={shellTopBarStyle}
        viewSchemaId={surface}
      >
        <div style={shellIncidentIdentityStyle}>
          <span style={shellTopBarLabelStyle}>Incident:</span>
          <strong style={shellTopBarValueStyle}>{incidentId}</strong>
        </div>
        <nav aria-label="Built-in workbook surfaces" style={tabStripStyle}>
          {requiredBuiltInWorkbookSurfaceIds.map((viewSchemaID) => {
            const contract = requireViewContract(viewSchemaID);
            return (
              <button
                aria-current={surface === viewSchemaID ? "page" : undefined}
                key={viewSchemaID}
                data-testid={surfaceTabTestId(viewSchemaID)}
                data-view-schema-id={viewSchemaID}
                data-workbook-tab-index={String(
                  requiredBuiltInWorkbookSurfaceIds.indexOf(viewSchemaID),
                )}
                style={{
                  ...surfaceTabStyle,
                  ...(surface === viewSchemaID ? surfaceTabActiveStyle : null),
                }}
                type="button"
                onClick={() => {
                  selectWorkbookSurface(viewSchemaID);
                }}
              >
                {contract.title}
              </button>
            );
          })}
        </nav>
        <div style={systemViewSlotStyle}>
          <SystemViewSwitcher
            activeViewSchemaId={surface}
            onSelect={(viewSchemaId) => {
              selectWorkbookSurface(viewSchemaId, {
                focusFirstGridTarget: true,
              });
            }}
          />
        </div>
        <div style={shellTopBarActionsStyle}>
          <span
            data-testid={phase1RouteTestId("workbook-current-user")}
            style={currentUserChipStyle}
            title={currentUserLabel ?? "Unknown user"}
          >
            {displayInitials(currentUserLabel ?? "Unknown user")}
          </span>
          <div data-testid={currentIncidentRoleTestId()} style={roleBadgeStyle}>
            Current incident role: {currentIncidentRole || "viewer"}
          </div>
          <button
            aria-controls={incidentControlsPanelTestId()}
            aria-expanded={incidentControlsOpen}
            data-testid={incidentControlsTriggerTestId()}
            style={secondaryActionButtonStyle}
            type="button"
            onClick={() => {
              setIncidentControlsOpen((current) => !current);
            }}
          >
            Controls
          </button>
          {onReturnToLanding ? (
            <button
              data-testid={phase1LandingTestId("return")}
              style={secondaryActionButtonStyle}
              type="button"
              onClick={onReturnToLanding}
            >
              Incidents
            </button>
          ) : null}
        </div>
      </WorkbookShellSlotRegion>

      {entityLoadError ? (
        <p data-testid="entity-load-error" style={bodyStyle}>
          {entityLoadError}
        </p>
      ) : null}

      {surface === timelineViewSchemaId ? (
        <TimelineWorkbook
          apiBase={apiBase}
          currentIncidentRole={currentIncidentRole}
          entityIndex={entityIndex}
          hostEntities={hostRows}
          identityEntities={identityRows}
          incidentId={incidentId}
          filterDraft={timelineFilterDraft}
          onFilterDraftChange={setTimelineFilterDraft}
          onQueryStateChange={setTimelineQueryState}
          onRefreshEntities={loadEntities}
          queryState={timelineQueryState}
          reloadToken={sheetReloadToken}
          savedViewSelector={activeSavedViewSelector}
          sheetRef={startupSheetRef}
        />
      ) : surface === hostsViewSchemaId ||
        surface === identitiesViewSchemaId ? (
        <EntityWorkbookSurface
          apiBase={apiBase}
          currentIncidentRole={currentIncidentRole}
          entityIndex={entityIndex}
          entityType={surface === hostsViewSchemaId ? "host" : "identity"}
          filterDraft={
            surface === hostsViewSchemaId
              ? hostFilterDraft
              : identityFilterDraft
          }
          incidentId={incidentId}
          onApplyFilter={
            surface === hostsViewSchemaId
              ? applyHostFilter
              : applyIdentityFilter
          }
          onFilterDraftChange={
            surface === hostsViewSchemaId
              ? setHostFilterDraft
              : setIdentityFilterDraft
          }
          onGroupByChange={(groupBy) => {
            if (surface === hostsViewSchemaId) {
              setHostQueryState((current) =>
                updateGroupBy(hostsContract, current, groupBy),
              );
              return;
            }
            setIdentityQueryState((current) =>
              updateGroupBy(identitiesContract, current, groupBy),
            );
          }}
          onRemoveFilter={(fieldKey) => {
            if (surface === hostsViewSchemaId) {
              setHostQueryState((current) =>
                removeFilterField(current, fieldKey),
              );
              return;
            }
            setIdentityQueryState((current) =>
              removeFilterField(current, fieldKey),
            );
          }}
          onRefreshEntities={loadEntities}
          onToggleSort={(fieldKey) => {
            if (surface === hostsViewSchemaId) {
              setHostQueryState((current) =>
                toggleSortField(hostsContract, current, fieldKey),
              );
              return;
            }
            setIdentityQueryState((current) =>
              toggleSortField(identitiesContract, current, fieldKey),
            );
          }}
          queryState={
            surface === hostsViewSchemaId ? hostQueryState : identityQueryState
          }
          rows={surface === hostsViewSchemaId ? hostRows : identityRows}
          savedViewSelector={activeSavedViewSelector}
        />
      ) : surface === assessmentsViewSchemaId ? (
        <AssessmentWorkbookSurface
          apiBase={apiBase}
          assessmentRows={assessmentRows}
          currentIncidentRole={currentIncidentRole}
          filterDraft={assessmentFilterDraft}
          hostRows={hostRows}
          identityRows={identityRows}
          incidentId={incidentId}
          loadError={assessmentLoadError}
          onApplyFilter={() => {
            applyFilterDraftToQuery(
              setAssessmentQueryState,
              setAssessmentFilterDraft,
              assessmentFilterDraft,
            );
          }}
          onFilterDraftChange={setAssessmentFilterDraft}
          onGroupByChange={(groupBy) => {
            setAssessmentQueryState((current) =>
              updateGroupBy(assessmentsContract, current, groupBy),
            );
          }}
          onRefreshAssessmentRows={loadAssessmentSurface}
          onRemoveFilter={(fieldKey) => {
            setAssessmentQueryState((current) =>
              removeFilterField(current, fieldKey),
            );
          }}
          onToggleSort={(fieldKey) => {
            setAssessmentQueryState((current) =>
              toggleSortField(assessmentsContract, current, fieldKey),
            );
          }}
          queryState={assessmentQueryState}
          savedViewSelector={activeSavedViewSelector}
        />
      ) : (
        <GenericWorkbookSurface
          key={activeContract.viewSchemaId}
          apiBase={apiBase}
          contract={activeContract}
          currentUserId={currentUserId}
          filterDraft={genericFilterDraft}
          incidentId={incidentId}
          loadError={genericLoadError}
          onApplyFilter={() => {
            applyFilterDraftToQuery(
              setGenericQueryState,
              setGenericFilterDraft,
              genericFilterDraft,
            );
          }}
          onFilterDraftChange={setGenericFilterDraft}
          onGroupByChange={(groupBy) => {
            setGenericQueryState((current) =>
              updateGroupBy(activeContract, current, groupBy),
            );
          }}
          onRemoveFilter={(fieldKey) => {
            setGenericQueryState((current) =>
              removeFilterField(current, fieldKey),
            );
          }}
          onRefresh={loadGenericSurface}
          onToggleSort={(fieldKey) => {
            setGenericQueryState((current) =>
              toggleSortField(activeContract, current, fieldKey),
            );
          }}
          queryState={genericQueryState}
          rows={genericRows}
          savedViewSelector={activeSavedViewSelector}
        />
      )}

      {incidentControlsOpen ? (
        <section
          data-testid={incidentControlsPanelTestId()}
          data-workbook-shell-region="support"
          id={incidentControlsPanelTestId()}
          style={supportRegionStyle}
        >
          <IncidentAdminPanel
            apiBase={apiBase}
            currentIncidentRole={currentIncidentRole}
            incidentId={incidentId}
            onIncidentAccessLost={onIncidentAccessLost}
            onIncidentSnapshot={onIncidentSnapshot}
            onSessionRoleChange={loadSessionRole}
          />
        </section>
      ) : null}
    </section>
  );
}

const panelStyle = {
  boxSizing: "border-box" as const,
  width: "100%",
  minHeight: "100vh",
  margin: 0,
  padding: 0,
  borderRadius: 0,
  background: "var(--ct-colors-canvas)",
  boxShadow: "none",
  border: 0,
};

const shellTopBarStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.75rem",
  minHeight: "var(--ct-layout-topBarHeight)",
  padding: "0 1rem",
  borderBottom: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
};

const shellTopBarActionsStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "flex-end",
  gap: "0.45rem",
  flex: "0 1 auto",
  minWidth: 0,
};

const shellTopBarLabelStyle = {
  margin: 0,
  fontSize: "0.85rem",
  color: "var(--ct-colors-ink-subtle)",
};

const shellTopBarValueStyle = {
  margin: 0,
  fontWeight: 650,
  color: "var(--ct-colors-ink)",
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
};

const shellIncidentIdentityStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.45rem",
  flex: "0 1 18rem",
  minWidth: 0,
};

const currentUserChipStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  flex: "0 0 auto",
  width: "2rem",
  height: "2rem",
  borderRadius: "var(--ct-rounded-pill)",
  border: "var(--ct-border-hairline)",
  color: "var(--ct-colors-ink)",
  background: "var(--ct-colors-surface-2)",
  fontSize: "0.82rem",
  fontWeight: 700,
};

const surfaceStackStyle = {
  position: "relative" as const,
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  minHeight: "calc(100vh - var(--ct-layout-topBarHeight))",
  alignContent: "start",
};

const headerStyle = {
  position: "relative" as const,
  display: "flex",
  justifyContent: "space-between",
  gap: "1rem",
  alignItems: "flex-start",
  marginBottom: "1rem",
};

const eyebrowStyle = {
  margin: 0,
  fontSize: "0.78rem",
  letterSpacing: "0.12em",
  textTransform: "uppercase" as const,
  color: "var(--ct-colors-accent)",
};

const headlineStyle = {
  margin: "0.35rem 0 0.5rem",
  fontSize: "2rem",
  lineHeight: 1.1,
};

const bodyStyle = {
  margin: 0,
  lineHeight: 1.5,
  color: "var(--ct-colors-ink-muted)",
};

const splitShellStyle = {
  display: "grid",
  gap: "var(--ct-spacing-shell-gap)",
  alignItems: "stretch",
  gridTemplateColumns:
    "minmax(0, 1fr) minmax(var(--ct-layout-inspectorMinWidth), var(--ct-layout-inspectorDefaultWidth))",
  minHeight: 0,
  padding: "var(--ct-spacing-sm)",
};

const viewBarStyle = {
  position: "relative" as const,
  zIndex: 4,
  display: "flex",
  alignItems: "center",
  flexWrap: "wrap" as const,
  gap: "0.75rem",
  minHeight: "var(--ct-layout-viewBarHeight)",
  marginBottom: "var(--ct-spacing-sm)",
  padding: "0 var(--ct-spacing-sm)",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-1)",
  overflowX: "visible" as const,
};

const gridShellStyle = {
  overflow: "hidden",
  overflowAnchor: "none" as const,
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
};

const actionStackStyle = {
  display: "grid",
  gap: "0.5rem",
};

const genericMutationPanelStyle = {
  position: "relative" as const,
  zIndex: 2,
  display: "grid",
  gap: "0.75rem",
  padding: "1rem",
  borderRadius: "var(--ct-rounded-lg)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
};

const genericFormGridStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(12rem, 1fr))",
  gap: "0.75rem",
};

const genericEditRowStyle = {
  display: "grid",
  gridTemplateColumns:
    "minmax(10rem, 1fr) minmax(10rem, 1fr) minmax(14rem, 2fr) auto",
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

const textareaStyle = {
  ...inputStyle,
  resize: "vertical" as const,
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
  display: "grid",
  gap: "0.75rem",
  margin: "1rem 0",
  padding: "1rem",
  borderRadius: "var(--ct-rounded-lg)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
};

const evidencePreviewHeaderStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "1rem",
  alignItems: "start",
};

const evidencePreviewFrameStyle = {
  width: "100%",
  minHeight: "28rem",
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

const inspectorShellStyle = {
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-component-inspector-border)",
  background: "var(--ct-component-inspector-backgroundColor)",
  color: "var(--ct-component-inspector-textColor)",
  padding: "var(--ct-spacing-panel-padding)",
  position: "sticky" as const,
  top: "calc(var(--ct-layout-topBarHeight) + var(--ct-spacing-sm))",
  maxHeight:
    "calc(100vh - var(--ct-layout-topBarHeight) - var(--ct-layout-statusStripHeight) - 16px)",
  overflow: "auto",
};

const inspectorHeaderStyle = {
  display: "grid",
  gap: "0.35rem",
  marginBottom: "1rem",
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

const sectionTitleStyle = {
  margin: 0,
  fontSize: "1rem",
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

const inlineButtonRowStyle = {
  display: "flex",
  gap: "0.5rem",
  flexWrap: "wrap" as const,
};

const noticeTitleStyle = {
  margin: 0,
  fontSize: "0.95rem",
  fontWeight: 600,
};

const tabStripStyle = {
  display: "flex",
  alignItems: "stretch",
  gap: "0.2rem",
  flex: "1 1 auto",
  minWidth: 0,
  overflow: "hidden",
};

const surfaceTabStyle = {
  borderRadius: 0,
  border: 0,
  borderBottom: "2px solid transparent",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  padding: "0.85rem 0.55rem 0.7rem",
  font: "inherit",
  cursor: "pointer",
  whiteSpace: "nowrap" as const,
};

const surfaceTabActiveStyle = {
  background: "transparent",
  color: "var(--ct-colors-ink)",
  borderBottomColor: "var(--ct-colors-accent)",
};

const systemViewSlotStyle = {
  flex: "0 0 11rem",
  minWidth: "10rem",
};

const roleBadgeStyle = {
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink-muted)",
  padding: "0.35rem 0.55rem",
  fontSize: "0.8rem",
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
};

const supportRegionStyle = {
  marginTop: "1rem",
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
