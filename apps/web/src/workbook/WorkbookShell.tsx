import {
  buildGridPresentationRows,
  type GridActionsColumn,
  type GridColumn,
  type GridDensity,
  type GridRow,
  GridTable,
  GridViewport,
  reconcileRecordRows,
  resolveGridPasteTargets,
} from "@cartulary/grid-adapter";
import {
  assessmentCreatePanelTestId,
  dataTestIdSelector,
  entityInspectButtonTestId,
  entityInspectorTestId,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  genericEditFieldSelectTestId,
  genericEditRecordSelectTestId,
  genericEditSubmitTestId,
  genericEditValueTestId,
  gridActionsHeaderTestId,
  gridGroupRowTestId,
  gridShellTestId,
  gridSortHeaderTestId,
  type IncidentControlsSection,
  incidentControlsCloseButtonTestId,
  incidentControlsPanelTestId,
  surfaceTabTestId,
  timelinePreviewRowTestId,
  type WorkbookSurface,
  workbookIncidentIdentityTestId,
  workbookInlineDraftRowTestId,
  workbookInspectorCloseButtonTestId,
  workbookResponsiveBandTestId,
  workbookRowActionMenuButtonTestId,
  workbookShellReadyTestId,
  workbookSurfacesMenuOptionTestId,
  workbookSurfacesMenuTestId,
  workbookSurfacesMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import {
  requireViewContract,
  resolveHeaderSortFieldKey,
  visibleFields,
} from "@cartulary/view-contracts";
import { MoreHorizontal, X } from "lucide-react";
import {
  type CSSProperties,
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
import { apiPath } from "../services/browserApi";
import {
  abortLatestQuery,
  beginLatestQuery,
  fetchJSON,
  handleWorkbookLoadFailure,
  isAbortError,
  type LatestQueryRuntime,
  parseErrorMessage,
  readEnvelope,
} from "../services/workbookApi";
import type {
  WorkbookAccountApplicationMenuProps,
  WorkbookAccountModel,
  WorkbookIncidentControlsMenuItem,
  WorkbookIncidentControlsRendererProps,
  WorkbookIncidentRole,
  WorkbookIncidentSnapshot,
} from "../shared/workbookShellContracts";
import { ActiveSurfaceSavedViewSelector } from "./components/ActiveSurfaceSavedViewSelector";
import { GenericMutationControl } from "./components/GenericMutationControl";
import { GenericWorkbookSurface } from "./components/GenericWorkbookSurface";
import { SystemViewSwitcher } from "./components/SystemViewSwitcher";
import { WorkbookGridControls } from "./components/WorkbookGridControls";
import {
  type InspectorDisabledToken,
  WorkbookInspectorPanelSection,
} from "./components/WorkbookInspectorFeatureGroups";
import { WorkbookSheetToolbar } from "./components/WorkbookSheetToolbar";
import {
  WorkbookShellSlotRegion,
  workbookShellId,
} from "./components/WorkbookShellSlots";
import { WorkbookSurfaceStatusStrip } from "./components/WorkbookStatusStrip";
import {
  WorkbookSurfaceFrame,
  workbookSurfaceGridShellStyle,
  workbookSurfaceInspectorPanelStyle,
  workbookSurfaceOverlayPanelStyle,
} from "./components/WorkbookSurfaceFrame";
import { useAssessmentSupportRows } from "./hooks/useAssessmentSupportRows";
import { useEntityTimelinePreview } from "./hooks/useEntityTimelinePreview";
import { useWorkbookIncidentIdentity } from "./hooks/useWorkbookIncidentIdentity";
import { useWorkbookPendingGridFocus } from "./hooks/useWorkbookPendingGridFocus";
import { useWorkbookResponsiveLayout } from "./hooks/useWorkbookResponsiveLayout";
import { useWorkbookShellRuntime } from "./hooks/useWorkbookShellRuntime";
import {
  assessmentColumnWidth,
  initialAssessmentDraft,
  isAssessmentConfidenceBand,
  supportRowLabel,
} from "./models/assessmentWorkbookModel";
import {
  buildMergePlan,
  type EntityRow,
  entityContractColumnWidth,
  entityGroupLabel,
  entityRowFromApi,
} from "./models/entityWorkbookModel";
import {
  buildGenericCreatePayload,
  buildGenericPatchChange,
  enumValuesFor,
  genericCellLabel,
  genericCreateMinimumMessage,
  genericRowLabel,
  initialGenericCreateDraft,
  parseMutationError,
  selectWorkbookEditTarget,
} from "./models/genericWorkbookModel";
import {
  normalizeWorkbookViewRows,
  workbookContractColumns,
  workbookGridRows,
} from "./models/workbookContractRows";
import {
  type AccountDensityMode,
  resolveEffectiveWorkbookDensity,
} from "./models/workbookDensity";
import type { WorkbookIncidentIdentity } from "./models/workbookIncidentIdentity";
import {
  inspectorNoRowState,
  inspectorPanelIsDeclared,
  selectInspectorConfig,
} from "./models/workbookInspectorModel";
import {
  buildQueryRequest,
  toggleSortField,
  type WorkbookQueryState,
} from "./models/workbookQuery";
import { emptyGenericReferenceOptions } from "./models/workbookReferenceOptions";
import {
  assessmentsViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  notesViewSchemaId,
  requiredBuiltInWorkbookSurfaceIds,
  timelineViewSchemaId,
} from "./models/workbookSurfaceRegistry";
import { RelationshipChip } from "./timeline/components/TimelineCellEditors";
import { TimelineWorkbook } from "./timeline/components/TimelineWorkbook";
import {
  type AssessmentCreateDraft,
  buildAssessmentCreatePayload,
  type EntityApiRow,
} from "./timeline/models/workbookTimelineModel";
import {
  clipboardGridDimensions,
  clipboardTextLooksTabular,
} from "./utils/workbookClipboard";
import {
  FocusableWorkbookCell,
  useWorkbookGridFocus,
} from "./utils/workbookGridFocus";
import { displayInitials } from "./utils/workbookPresence";

export type { WorkbookIncidentIdentity } from "./models/workbookIncidentIdentity";
export type {
  WorkbookAccountApplicationMenuProps,
  WorkbookAccountModel,
  WorkbookIncidentControlsMenuItem,
  WorkbookIncidentControlsRendererProps,
  WorkbookIncidentRole,
  WorkbookIncidentSnapshot,
};

const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);
const assessmentsContract = requireViewContract(assessmentsViewSchemaId);
const incidentControlsMenuItems = [
  {
    section: "summary",
    label: "Summary and preferences",
    description: "Incident summary and workbook defaults",
  },
  {
    section: "incident-fields",
    label: "Promoted fields",
    description: "TLP, phase, and external case",
  },
  {
    section: "memberships",
    label: "Memberships",
    description: "Incident access and roles",
  },
  {
    section: "membership-audit",
    label: "Membership audit",
    description: "Incident membership changes",
  },
] as const satisfies ReadonlyArray<{
  readonly description: string;
  readonly label: string;
  readonly section: IncidentControlsSection;
}>;
const defaultIncidentControlsMenuItem = incidentControlsMenuItems[0];

function requireIncidentControlsMenuItem(section: IncidentControlsSection) {
  return (
    incidentControlsMenuItems.find((item) => item.section === section) ??
    defaultIncidentControlsMenuItem
  );
}

type SaveState = "Syncing" | "Saved" | "Conflict";
type MutationErrorSetter = Dispatch<SetStateAction<string | null>>;
type MutationStateSetter = Dispatch<SetStateAction<SaveState>>;
type IncidentRole = WorkbookIncidentRole;

type WorkbookShellProps = {
  incidentId: string;
  apiBase?: string | undefined;
  account?: WorkbookAccountModel | undefined;
  accountDensityMode?: AccountDensityMode | undefined;
  accountApplicationMenu?:
    | ((props: WorkbookAccountApplicationMenuProps) => ReactNode)
    | undefined;
  currentUserLabel?: string | undefined;
  initialIncidentIdentity?: WorkbookIncidentIdentity | undefined;
  onIncidentSnapshot?:
    | ((incident: WorkbookIncidentSnapshot) => void)
    | undefined;
  onIncidentAccessLost?: (() => void) | undefined;
  renderIncidentControls?:
    | ((props: WorkbookIncidentControlsRendererProps) => ReactNode)
    | undefined;
};

type TimelineMutationEnvelope = {
  data: {
    view_schema_id: string;
    change_set_id?: string;
    row: unknown;
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
  density,
  entityType,
  inspectorResetKey,
  savedViewSelector,
  rows,
  onToggleSort,
  queryState,
  currentIncidentRole,
  entityIndex,
  onRefreshEntities,
}: {
  incidentId: string;
  apiBase?: string | undefined;
  density: GridDensity;
  entityType: EntityRow["entityType"];
  inspectorResetKey: string;
  savedViewSelector?: ReactNode | undefined;
  onToggleSort: (fieldKey: string) => void;
  queryState: WorkbookQueryState;
  rows: EntityRow[];
  currentIncidentRole: IncidentRole | null;
  entityIndex: Record<string, EntityRow>;
  onRefreshEntities: () => Promise<void>;
}) {
  const [selectedRecordId, setSelectedRecordId] = useState<string | null>(null);
  const [isInspectorOpen, setIsInspectorOpen] = useState(false);
  const [mergeCandidateId, setMergeCandidateId] = useState<string>("");
  const [mergeReason, setMergeReason] = useState("Merge duplicate entity");
  const [mergeMessage, setMergeMessage] = useState<string | null>(null);
  const [editRecordId, setEditRecordId] = useState("");
  const [editFieldKey, setEditFieldKey] = useState("");
  const [editValue, setEditValue] = useState("");
  const [createDraft, setCreateDraft] = useState<Record<string, string>>(() =>
    initialGenericCreateDraft(
      entityType === "host" ? hostsContract : identitiesContract,
      null,
    ),
  );
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
    setEditRecordId("");
    setEditFieldKey("");
    setEditValue("");
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
      return;
    }
    setMergeMessage(null);
  }, [selectedEntityRecordKey]);

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
    const result = await fetchJSON<ViewMutationEnvelope>(
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
                  <p data-testid="merge-message" style={bodyStyle}>
                    {mergeMessage}
                  </p>
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

function AssessmentWorkbookSurface({
  apiBase,
  assessmentRows,
  currentIncidentRole,
  density,
  inspectorResetKey,
  savedViewSelector,
  hostRows,
  identityRows,
  incidentId,
  loadError,
  onRefreshAssessmentRows,
  onToggleSort,
  queryState,
}: {
  apiBase?: string | undefined;
  assessmentRows: EntityApiRow[];
  currentIncidentRole: IncidentRole | null;
  density: GridDensity;
  inspectorResetKey: string;
  savedViewSelector?: ReactNode | undefined;
  hostRows: EntityRow[];
  identityRows: EntityRow[];
  incidentId: string;
  loadError: string | null;
  onRefreshAssessmentRows: () => Promise<void>;
  onToggleSort: (fieldKey: string) => void;
  queryState: WorkbookQueryState;
}) {
  const [draft, setDraft] = useState<AssessmentCreateDraft>(() =>
    initialAssessmentDraft(assessmentsContract),
  );
  const [isInspectorOpen, setIsInspectorOpen] = useState(false);
  const inspectorConfig = selectInspectorConfig(assessmentsContract);
  const showWorkflowPanel = inspectorPanelIsDeclared(
    inspectorConfig,
    "workflow",
  );
  const supportRows = useAssessmentSupportRows({ apiBase, incidentId });
  const [message, setMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const assessmentInspectorDisabledTokens = useMemo(
    () => new Set<InspectorDisabledToken>(),
    [],
  );
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
  const gridRows: readonly GridRow<EntityApiRow>[] = workbookGridRows({
    getRecordId: (row) => row.record_id,
    rows: assessmentRows,
    surface: assessmentsViewSchemaId,
  });
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
    if (inspectorResetKey === "") {
      return;
    }
    setIsInspectorOpen(false);
    setDraft(initialAssessmentDraft(assessmentsContract));
    setMessage(null);
  }, [inspectorResetKey]);

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
    <WorkbookSurfaceFrame
      inspector={
        isInspectorOpen && showWorkflowPanel ? (
          <aside
            data-testid={assessmentCreatePanelTestId()}
            style={inspectorShellStyle}
          >
            <div style={inspectorHeaderStyle}>
              <div style={inspectorTitleRowStyle}>
                <div>
                  <p style={eyebrowStyle}>Create</p>
                  <h2 style={inspectorTitleStyle}>Append assessment</h2>
                </div>
                <button
                  aria-label="Close inspector"
                  data-testid={workbookInspectorCloseButtonTestId(
                    assessmentsViewSchemaId,
                  )}
                  style={inspectorCloseButtonStyle}
                  type="button"
                  onClick={() => {
                    setIsInspectorOpen(false);
                  }}
                >
                  <X aria-hidden="true" size={16} />
                </button>
              </div>
            </div>
            {inspectorConfig.panels.map((panel) => (
              <WorkbookInspectorPanelSection
                config={inspectorConfig}
                disabledTokens={assessmentInspectorDisabledTokens}
                key={panel.panelId}
                panelId={panel.panelId}
              />
            ))}
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
                  {enumValuesFor(
                    assessmentsContract,
                    "assessment.subject_type",
                    ["host", "identity"],
                  ).map((value) => (
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
        ) : undefined
      }
      primaryGrid={
        <GridViewport
          blockSizing="fill"
          style={gridShellStyle}
          testId={gridShellTestId(assessmentsViewSchemaId)}
        >
          <GridTable
            columns={columns}
            density={density}
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
      }
      statusStrip={
        <WorkbookSurfaceStatusStrip
          mutationState="Saved"
          workbookFocusAnchor={assessmentFocus.anchor}
        />
      }
      viewBar={
        <WorkbookSheetToolbar
          addRowDisabled={!canCreate}
          leading={savedViewSelector}
          onAddRow={() => {
            setIsInspectorOpen(true);
          }}
          onInspectorToggle={() => {
            setIsInspectorOpen(true);
          }}
          surface={assessmentsViewSchemaId}
        />
      }
      viewSchemaId={assessmentsViewSchemaId}
      workAreaOverlays={
        loadError ? (
          <p
            data-testid="assessment-surface-load-error"
            style={surfaceNoticeOverlayStyle}
          >
            {loadError}
          </p>
        ) : undefined
      }
    />
  );
}

export function WorkbookShell({
  incidentId,
  apiBase,
  account,
  accountDensityMode,
  accountApplicationMenu,
  currentUserLabel,
  initialIncidentIdentity,
  onIncidentSnapshot,
  onIncidentAccessLost,
  renderIncidentControls,
}: WorkbookShellProps) {
  const responsiveLayout = useWorkbookResponsiveLayout();
  const responsiveBand = responsiveLayout.chromeMode;
  const surfaceSelectionVersionRef = useRef(0);
  const workbookRuntime = useWorkbookShellRuntime({
    apiBase,
    incidentId,
    onIncidentAccessLost,
    surfaceSelectionVersionRef,
  });
  const {
    activeContract,
    activeQueryControls,
    activeSavedViewModified,
    assessmentQueryState,
    genericQueryState,
    hostQueryState,
    identityQueryState,
    pendingGridFocusSurface,
    savedViews,
    sheetReloadToken,
    startupSheetRef,
    surface,
    timelineQueryState,
  } = workbookRuntime.snapshot;
  const effectiveDensity = useMemo(
    () => resolveEffectiveWorkbookDensity(surface, accountDensityMode),
    [accountDensityMode, surface],
  );
  const {
    createSavedView,
    deleteSavedView,
    duplicateSavedView,
    selectSavedView,
    selectWorkbookSurface,
    setAssessmentQueryState,
    setGenericQueryState,
    setHostQueryState,
    setIdentityQueryState,
    setPendingGridFocusSurface,
    setTimelineQueryState,
    setWorkbookDefaultSheetRef,
    setWorkbookHomeSheetRef,
    updateSavedView,
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
  const [currentUserId, setCurrentUserId] = useState<string | null>(
    () => account?.user_id ?? null,
  );
  const [currentIncidentRole, setCurrentIncidentRole] =
    useState<IncidentRole | null>(null);
  const { incidentIdentity, incidentIdentityError } =
    useWorkbookIncidentIdentity({
      apiBase,
      incidentId,
      initialIncidentIdentity,
      onIncidentAccessLost,
      onIncidentSnapshot,
    });
  const [controlsDrawerSection, setControlsDrawerSection] =
    useState<IncidentControlsSection | null>(null);
  const [lastControlsSection, setLastControlsSection] =
    useState<IncidentControlsSection>("summary");
  const controlsReturnFocusTargetRef = useRef<HTMLElement | null>(null);
  const controlsDrawerCloseRef = useRef<HTMLButtonElement | null>(null);
  const [surfacesMenuOpen, setSurfacesMenuOpen] = useState(false);
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
  useEffect(() => {
    if (account?.user_id) {
      setCurrentUserId(account.user_id);
    }
  }, [account?.user_id]);

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
    setGenericRows([]);
    setGenericLoadError(null);
  }, [startupSheetRef.kind]);

  useEffect(() => {
    void sheetReloadToken;
    void loadGenericSurface();
  }, [loadGenericSurface, sheetReloadToken]);

  useEffect(() => {
    void sheetReloadToken;
    void loadAssessmentSurface();
  }, [loadAssessmentSurface, sheetReloadToken]);

  useWorkbookPendingGridFocus({
    pendingGridFocusSurface,
    setPendingGridFocusSurface,
    surface,
  });

  const deferControlsFocus = useCallback(
    (resolveTarget: () => HTMLElement | null) => {
      window.setTimeout(() => {
        resolveTarget()?.focus({ preventScroll: true });
      }, 0);
    },
    [],
  );

  const closeControlsDrawer = useCallback(
    (options: { readonly restoreTriggerFocus: boolean }) => {
      setControlsDrawerSection(null);
      if (options.restoreTriggerFocus) {
        deferControlsFocus(() => controlsReturnFocusTargetRef.current);
      }
    },
    [deferControlsFocus],
  );

  const openControlsDrawer = useCallback(
    (
      section: IncidentControlsSection,
      returnFocusTarget?: HTMLElement | null,
    ) => {
      controlsReturnFocusTargetRef.current = returnFocusTarget ?? null;
      setLastControlsSection(section);
      setControlsDrawerSection(section);
    },
    [],
  );

  useEffect(() => {
    if (controlsDrawerSection === null) {
      return;
    }
    deferControlsFocus(() => controlsDrawerCloseRef.current);
  }, [controlsDrawerSection, deferControlsFocus]);

  const activeControlsMenuItem =
    controlsDrawerSection === null
      ? requireIncidentControlsMenuItem(lastControlsSection)
      : requireIncidentControlsMenuItem(controlsDrawerSection);
  const activeSurfaceIsBuiltIn = requiredBuiltInWorkbookSurfaceIds.some(
    (viewSchemaId) => viewSchemaId === surface,
  );
  const activeSystemSurfaceTitle = activeSurfaceIsBuiltIn
    ? null
    : activeContract.title;
  const incidentKeyLabel = incidentIdentity?.incident_key ?? "Incident";
  const incidentTitleLabel = incidentIdentity?.title ?? "Loading incident";
  const accountDisplayName =
    account?.display_name ?? currentUserLabel ?? "Unknown user";
  const accountTitle = account
    ? `${account.display_name}${account.is_deployment_admin ? " (deployment administrator)" : ""}`
    : accountDisplayName;

  const activeSavedViewSelector = (
    <ActiveSurfaceSavedViewSelector
      activeViewSchemaId={surface}
      currentIncidentRole={currentIncidentRole}
      currentUserId={currentUserId}
      isModified={activeSavedViewModified}
      onCreateSavedView={createSavedView}
      onDeleteSavedView={deleteSavedView}
      onDuplicateSavedView={duplicateSavedView}
      onResetToSavedView={selectSavedView}
      onSelectBaseSurface={selectWorkbookSurface}
      onSelectSavedView={selectSavedView}
      onSetDefaultSheetRef={setWorkbookDefaultSheetRef}
      onSetHomeSheetRef={setWorkbookHomeSheetRef}
      onUpdateSavedView={updateSavedView}
      savedViews={savedViews}
      selectedSheetRef={startupSheetRef}
    />
  );
  const workbookAccountApplicationMenu = accountApplicationMenu?.({
    currentIncidentRole,
    incidentControls: {
      activeSection: lastControlsSection,
      items: incidentControlsMenuItems,
      onSelectSection: openControlsDrawer,
    },
  });
  const incidentControlsDrawer =
    controlsDrawerSection === null
      ? null
      : (renderIncidentControls?.({
          activeSection: controlsDrawerSection,
          apiBase,
          currentIncidentRole,
          incidentId,
          onIncidentAccessLost,
          onIncidentSnapshot,
          onSessionRoleChange: loadSessionRole,
        }) ?? null);
  const inspectorResetKey = `${surface}:${startupSheetRef.kind}:${startupSheetRef.id}:${sheetReloadToken}`;

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
        style={{
          ...shellTopBarStyle,
          ...(responsiveBand === "below_supported_minimum"
            ? shellTopBarUnsupportedStyle
            : null),
        }}
        viewSchemaId={surface}
      >
        <div
          data-testid={workbookIncidentIdentityTestId()}
          style={shellIncidentIdentityStyle}
          title={
            incidentIdentity === null
              ? (incidentIdentityError ?? "Loading incident")
              : `${incidentIdentity.incident_key} ${incidentIdentity.title}`
          }
        >
          <strong style={shellTopBarValueStyle}>{incidentKeyLabel}</strong>
          <span style={shellIncidentTitleStyle}>{incidentTitleLabel}</span>
        </div>
        <span
          aria-hidden="true"
          data-testid={workbookResponsiveBandTestId()}
          data-workbook-block-mode={responsiveLayout.blockMode}
          data-workbook-responsive-band={responsiveBand}
          hidden
        />
        {responsiveBand === "base" ? (
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
                    ...(surface === viewSchemaID
                      ? surfaceTabActiveStyle
                      : null),
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
        ) : (
          <div style={surfacesMenuFrameStyle}>
            <button
              aria-controls={
                surfacesMenuOpen ? workbookSurfacesMenuTestId() : undefined
              }
              aria-expanded={surfacesMenuOpen}
              aria-haspopup="menu"
              data-testid={workbookSurfacesMenuTriggerTestId()}
              style={surfaceMenuTriggerStyle}
              type="button"
              onClick={() => {
                setSurfacesMenuOpen((current) => !current);
              }}
            >
              Surfaces
            </button>
            {surfacesMenuOpen ? (
              <div
                data-testid={workbookSurfacesMenuTestId()}
                id={workbookSurfacesMenuTestId()}
                role="menu"
                style={surfacesMenuStyle}
              >
                {requiredBuiltInWorkbookSurfaceIds.map((viewSchemaID) => {
                  const contract = requireViewContract(viewSchemaID);
                  const isSelected = surface === viewSchemaID;
                  return (
                    <button
                      key={viewSchemaID}
                      aria-checked={isSelected}
                      data-testid={workbookSurfacesMenuOptionTestId(
                        viewSchemaID,
                      )}
                      data-view-schema-id={viewSchemaID}
                      role="menuitemradio"
                      style={{
                        ...surfacesMenuItemStyle,
                        ...(isSelected ? surfacesMenuItemSelectedStyle : null),
                      }}
                      type="button"
                      onClick={() => {
                        setSurfacesMenuOpen(false);
                        selectWorkbookSurface(viewSchemaID);
                      }}
                    >
                      {contract.title}
                    </button>
                  );
                })}
              </div>
            ) : null}
          </div>
        )}
        <div style={systemViewSlotStyle}>
          <SystemViewSwitcher
            activeViewSchemaId={surface}
            onSelect={(viewSchemaId) => {
              selectWorkbookSurface(viewSchemaId, {
                focusFirstGridTarget: true,
              });
            }}
          />
          {activeSystemSurfaceTitle ? (
            <span style={activeSystemViewTitleStyle}>
              {activeSystemSurfaceTitle}
            </span>
          ) : null}
        </div>
        {responsiveBand === "below_supported_minimum" ? null : (
          <div style={topBarQuerySlotStyle}>
            <WorkbookGridControls
              contract={activeQueryControls.contract}
              filterDraft={activeQueryControls.filterDraft}
              onApplyFilter={activeQueryControls.onApplyFilter}
              onClearAll={activeQueryControls.onClearAll}
              onFilterDraftChange={activeQueryControls.onFilterDraftChange}
              onGroupByChange={activeQueryControls.onGroupByChange}
              onRemoveFilter={activeQueryControls.onRemoveFilter}
              onToggleSort={activeQueryControls.onToggleSort}
              queryState={activeQueryControls.queryState}
              surface={activeQueryControls.surface}
            />
          </div>
        )}
        <div style={shellTopBarActionsStyle}>
          <div style={currentUserSlotStyle}>
            {workbookAccountApplicationMenu ?? (
              <span style={currentUserChipStyle} title={accountTitle}>
                {displayInitials(accountDisplayName)}
              </span>
            )}
          </div>
        </div>
      </WorkbookShellSlotRegion>

      <div style={shellContentRegionStyle}>
        {entityLoadError ? (
          <p data-testid="entity-load-error" style={shellContentNoticeStyle}>
            {entityLoadError}
          </p>
        ) : null}

        <div style={shellActiveSurfaceStyle}>
          {surface === timelineViewSchemaId ? (
            <TimelineWorkbook
              apiBase={apiBase}
              currentIncidentRole={currentIncidentRole}
              currentUserId={currentUserId}
              density={effectiveDensity}
              entityIndex={entityIndex}
              hostEntities={hostRows}
              identityEntities={identityRows}
              incidentId={incidentId}
              inspectorResetKey={inspectorResetKey}
              onQueryStateChange={setTimelineQueryState}
              onRefreshEntities={loadEntities}
              queryState={timelineQueryState}
              reloadToken={sheetReloadToken}
              renderInlineQueryControls={false}
              savedViewSelector={activeSavedViewSelector}
              sheetRef={startupSheetRef}
            />
          ) : surface === hostsViewSchemaId ||
            surface === identitiesViewSchemaId ? (
            <EntityWorkbookSurface
              apiBase={apiBase}
              currentIncidentRole={currentIncidentRole}
              density={effectiveDensity}
              entityIndex={entityIndex}
              entityType={surface === hostsViewSchemaId ? "host" : "identity"}
              incidentId={incidentId}
              inspectorResetKey={inspectorResetKey}
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
                surface === hostsViewSchemaId
                  ? hostQueryState
                  : identityQueryState
              }
              rows={surface === hostsViewSchemaId ? hostRows : identityRows}
              savedViewSelector={activeSavedViewSelector}
            />
          ) : surface === assessmentsViewSchemaId ? (
            <AssessmentWorkbookSurface
              apiBase={apiBase}
              assessmentRows={assessmentRows}
              currentIncidentRole={currentIncidentRole}
              density={effectiveDensity}
              hostRows={hostRows}
              identityRows={identityRows}
              incidentId={incidentId}
              inspectorResetKey={inspectorResetKey}
              loadError={assessmentLoadError}
              onRefreshAssessmentRows={loadAssessmentSurface}
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
              density={effectiveDensity}
              incidentId={incidentId}
              inspectorResetKey={inspectorResetKey}
              loadError={genericLoadError}
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
        </div>

        {controlsDrawerSection !== null ? (
          <section
            aria-labelledby="incident-controls-panel-title"
            data-testid={incidentControlsPanelTestId()}
            data-workbook-shell-region="support"
            id={incidentControlsPanelTestId()}
            role="dialog"
            style={supportRegionStyle}
            onKeyDown={(event) => {
              if (event.key === "Escape") {
                event.preventDefault();
                closeControlsDrawer({ restoreTriggerFocus: true });
              }
            }}
          >
            <header style={supportRegionHeaderStyle}>
              <div>
                <p style={supportRegionEyebrowStyle}>Controls</p>
                <h2
                  id="incident-controls-panel-title"
                  style={supportRegionTitleStyle}
                >
                  {activeControlsMenuItem.label}
                </h2>
              </div>
              <button
                ref={controlsDrawerCloseRef}
                aria-label="Close incident controls"
                data-testid={incidentControlsCloseButtonTestId()}
                style={supportRegionCloseButtonStyle}
                type="button"
                onClick={() => {
                  closeControlsDrawer({ restoreTriggerFocus: true });
                }}
              >
                <X aria-hidden="true" size={16} />
              </button>
            </header>
            <div style={supportRegionBodyStyle}>{incidentControlsDrawer}</div>
          </section>
        ) : null}
      </div>
    </section>
  );
}

const panelStyle = {
  boxSizing: "border-box" as const,
  display: "grid",
  gridTemplateRows: "auto minmax(0, 1fr)",
  width: "100%",
  blockSize: "100%",
  minBlockSize: 0,
  margin: 0,
  padding: 0,
  borderRadius: 0,
  background: "var(--ct-colors-canvas)",
  boxShadow: "none",
  border: 0,
  overflow: "hidden",
};

const shellTopBarStyle = {
  boxSizing: "border-box" as const,
  display: "flex",
  alignItems: "center",
  flexWrap: "nowrap" as const,
  gap: "0.55rem",
  inlineSize: "100%",
  maxInlineSize: "100%",
  blockSize: "var(--ct-layout-topBarHeight)",
  minBlockSize: "var(--ct-layout-topBarHeight)",
  minWidth: 0,
  padding: "0 0.75rem",
  borderBottom: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  overflow: "visible",
};

const shellTopBarUnsupportedStyle = {
  overflowX: "auto" as const,
};

const shellTopBarActionsStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "flex-end",
  gap: "0.45rem",
  flex: "0 0 auto",
  minWidth: 0,
  order: 5,
};

const shellTopBarValueStyle = {
  margin: 0,
  fontWeight: 650,
  color: "var(--ct-colors-ink)",
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
};

const shellIncidentTitleStyle = {
  minWidth: 0,
  color: "var(--ct-colors-ink-muted)",
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
};

const shellIncidentIdentityStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.45rem",
  flex: "0 1 11rem",
  minWidth: 0,
  overflow: "hidden",
};

const currentUserSlotStyle = {
  display: "inline-flex",
  alignItems: "center",
  flex: "0 1 8rem",
  maxInlineSize: "8rem",
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

const shellContentRegionStyle = {
  position: "relative" as const,
  display: "grid",
  gridTemplateRows: "auto minmax(0, 1fr)",
  blockSize: "100%",
  minBlockSize: 0,
  minHeight: 0,
  minWidth: 0,
  overflow: "hidden",
};

const shellContentNoticeStyle = {
  gridRow: 1,
  margin: 0,
  lineHeight: 1.5,
  color: "var(--ct-colors-ink-muted)",
  padding: "0.35rem 0.75rem",
  borderBottom: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
};

const shellActiveSurfaceStyle = {
  display: "grid",
  gridTemplateRows: "minmax(0, 1fr)",
  gridRow: 2,
  blockSize: "100%",
  minBlockSize: 0,
  minHeight: 0,
  minWidth: 0,
  overflow: "hidden",
};

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

const noticeTitleStyle = {
  margin: 0,
  fontSize: "0.95rem",
  fontWeight: 600,
};

const tabStripStyle = {
  display: "flex",
  alignItems: "stretch",
  gap: "0.2rem",
  flex: "0 1 auto",
  minWidth: 0,
  overflow: "hidden",
};

const surfaceTabStyle = {
  borderRadius: 0,
  border: 0,
  borderBottom: "2px solid transparent",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  padding: "0 0.35rem",
  font: "inherit",
  cursor: "pointer",
  whiteSpace: "nowrap" as const,
  minBlockSize: "var(--ct-layout-topBarHeight)",
};

const surfaceTabActiveStyle = {
  background: "transparent",
  color: "var(--ct-colors-ink)",
  borderBottomColor: "var(--ct-colors-accent)",
};

const systemViewSlotStyle = {
  display: "inline-flex",
  alignItems: "center",
  gap: "0.35rem",
  flex: "0 1 auto",
  minWidth: 0,
  order: 4,
};

const activeSystemViewTitleStyle = {
  color: "var(--ct-colors-ink)",
  fontSize: "0.86rem",
  fontWeight: 650,
  maxInlineSize: "6rem",
  minWidth: 0,
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
};

const topBarQuerySlotStyle = {
  display: "flex",
  alignItems: "center",
  flex: "1 1 14rem",
  boxSizing: "border-box" as const,
  minWidth: 0,
  minInlineSize: 0,
  overflow: "visible",
  order: 4,
};

const surfacesMenuFrameStyle = {
  position: "relative" as const,
  display: "inline-flex",
  flex: "0 0 auto",
};

const surfaceMenuTriggerStyle = {
  borderRadius: "var(--ct-rounded-xs)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-ink)",
  padding: "0.35rem 0.55rem",
  font: "inherit",
  cursor: "pointer",
  whiteSpace: "nowrap" as const,
};

const surfacesMenuStyle = {
  position: "absolute" as const,
  zIndex: 18,
  insetBlockStart: "calc(100% + 0.35rem)",
  insetInlineStart: 0,
  display: "grid",
  gap: "0.2rem",
  inlineSize: "min(16rem, 80vw)",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-md)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-popover)",
  padding: "0.45rem",
};

const surfacesMenuItemStyle = {
  border: 0,
  borderRadius: "var(--ct-rounded-xs)",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  cursor: "pointer",
  font: "inherit",
  padding: "0.45rem 0.5rem",
  textAlign: "left" as const,
};

const surfacesMenuItemSelectedStyle = {
  background: "var(--ct-colors-surface-3)",
  color: "var(--ct-colors-ink)",
  fontWeight: 700,
};

const supportRegionStyle = {
  position: "absolute" as const,
  zIndex: 12,
  top: "var(--ct-spacing-md)",
  right: "var(--ct-spacing-md)",
  bottom: "var(--ct-spacing-md)",
  display: "grid",
  gridTemplateRows: "auto minmax(0, 1fr)",
  inlineSize: "min(52rem, calc(100% - var(--ct-spacing-xl)))",
  maxInlineSize: "calc(100% - var(--ct-spacing-xl))",
  minBlockSize: 0,
  overflow: "hidden",
  boxSizing: "border-box" as const,
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-drawer)",
};

const supportRegionHeaderStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  gap: "var(--ct-spacing-md)",
  minWidth: 0,
  padding: "0.7rem 0.85rem",
  borderBottom: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
};

const supportRegionEyebrowStyle = {
  ...eyebrowStyle,
  margin: 0,
};

const supportRegionTitleStyle = {
  margin: "0.15rem 0 0",
  fontSize: "1rem",
  lineHeight: 1.2,
};

const supportRegionCloseButtonStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  flex: "0 0 auto",
  width: "1.75rem",
  height: "1.75rem",
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  cursor: "pointer",
};

const supportRegionBodyStyle = {
  minBlockSize: 0,
  minWidth: 0,
  overflow: "auto",
  padding: "var(--ct-spacing-md)",
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
