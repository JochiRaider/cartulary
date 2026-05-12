import {
  type GridActionsColumn,
  type GridColumn,
  type GridRow,
  GridTable,
  GridViewport,
  reconcileRecordRows,
} from "@cartulary/grid-adapter";
import {
  cellPresenceMarkerTestId,
  conflictMarkerTestId,
  draftCellTestId,
  gridActionsHeaderTestId,
  gridGroupRowTestId,
  gridScrollportSelector,
  gridShellTestId,
  gridSortHeaderTestId,
  pendingQueueCountTestId,
  pendingQueueNoticeTestId,
  relationshipItemsTestId,
  rowCellTestId,
  rowInspectButtonTestId,
  rowPresenceMarkerTestId,
  saveStateTestId,
  type WorkbookSurface,
} from "@cartulary/ui-contracts";
import {
  listViewContracts,
  requireViewContract,
  resolveHeaderSortFieldKey,
  type ViewContract,
  type ViewFieldContract,
  visibleFields,
} from "@cartulary/view-contracts";
import {
  type FocusEvent as ReactFocusEvent,
  type KeyboardEvent as ReactKeyboardEvent,
  type MouseEvent as ReactMouseEvent,
  startTransition,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { IncidentAdminPanel } from "./IncidentAdminPanel";
import { WorkbookGridControls } from "./WorkbookGridControls";
import {
  captureViewportAnchor,
  computeRestoredViewportScroll,
  isRectFullyVisibleWithinContainer,
  type ScrollPosition,
  type ViewportSnapshot,
} from "./workbookContinuity";
import {
  applyFilterDraft,
  buildQueryRequest,
  defaultFilterDraft,
  emptyWorkbookQueryState,
  type FilterDraft,
  removeFilterField,
  toggleSortField,
  updateGroupBy,
  type WorkbookQueryState,
} from "./workbookQuery";
import {
  buildAutoResolutionNotices,
  buildInspectorMentions,
  buildMentionPatchPayload,
  isRecordChangedMessage,
  type RecordChangedPayload,
  readCollectionItems,
  shouldIgnoreSelfOriginatedRecordChange,
} from "./workbookShellPhase4";

const timelineViewSchemaId = "cartulary.view.timeline.v1";
const hostsViewSchemaId = "cartulary.view.hosts.v1";
const identitiesViewSchemaId = "cartulary.view.identities.v1";
const evidenceViewSchemaId = "cartulary.view.evidence.v1";
const notesViewSchemaId = "cartulary.view.notes.v1";
const assessmentsViewSchemaId = "cartulary.view.assessments.v1";
const taskRequestsViewSchemaId = "cartulary.view.task_requests.v1";
const decisionsViewSchemaId = "cartulary.view.decisions.v1";
const partiesViewSchemaId = "cartulary.view.parties.v1";
const commLogViewSchemaId = "cartulary.view.comm_log.v1";
const handoffViewSchemaId = "cartulary.view.handoff.v1";
const statusReviewViewSchemaId = "cartulary.view.status_review.v1";
const lessonViewSchemaId = "cartulary.view.lesson.v1";
const timelineContract = requireViewContract(timelineViewSchemaId);
const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);
const assessmentsContract = requireViewContract(assessmentsViewSchemaId);
const allWorkbookContracts = listViewContracts();
const builtInSurfaceIDs = [
  timelineViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  evidenceViewSchemaId,
  notesViewSchemaId,
] as const;
const csrfCookieName = "cartulary_csrf";
const csrfHeaderName = "X-CSRF-Token";

type SaveState = "Syncing" | "Saved" | "Conflict";
type SurfaceKind = string;
type EditableField =
  | "timeline.occurred_at"
  | "timeline.summary"
  | "timeline.details"
  | "timeline.source_text";
type RelationshipFieldKey = "timeline.host_refs" | "timeline.identity_refs";
type RelationshipDraftKey = "hostRefs" | "identityRefs";
type CollectionFieldKey = RelationshipFieldKey | "timeline.tags";
type CollectionDraftKey = RelationshipDraftKey | "tags";
type FocusFieldKey = keyof RowValues | CollectionDraftKey;
type IncidentRole = "viewer" | "editor" | "reviewer" | "admin" | "";
type AssessmentSubjectType = "host" | "identity";
type AssessmentConfidenceBand = "unset" | "low" | "medium" | "high";

type RowValues = {
  occurredAt: string;
  summary: string;
  details: string;
  sourceText: string;
};

type CollectionDrafts = {
  hostRefs: string;
  identityRefs: string;
  tags: string;
};

type CollectionItem = {
  itemRef: string;
  entityType: "host" | "identity";
  itemKind: "resolved_ref" | "unresolved_mention" | string;
  displayText: string;
  rawText: string;
  resolvedRecordId: string | null;
  resolutionMethod: string | null;
  autoResolved: boolean;
  provenance: string | null;
  confidence: number | null;
  matchedAliasText: string | null;
};

type TagCollectionItem = {
  itemRef: string;
  itemKind: "tag" | string;
  displayText: string;
  rawText: string;
};

type WorkbookRow = {
  key: string;
  recordId: string | null;
  rowVersion: number | null;
  captureState: string;
  values: RowValues;
  committedValues: RowValues;
  collectionValues: {
    hostRefs: CollectionItem[];
    identityRefs: CollectionItem[];
    tags: TagCollectionItem[];
  };
  collectionDrafts: CollectionDrafts;
  pendingSignature: string | null;
  rawRow: TimelineApiRow | null;
};

type TimelineWorkbookProps = {
  incidentId: string;
  apiBase?: string | undefined;
  hostEntities?: EntityRow[];
  identityEntities?: EntityRow[];
  entityIndex?: Record<string, EntityRow>;
  currentIncidentRole?: IncidentRole | null;
  onRefreshEntities?: () => Promise<void> | void;
};

type WorkbookShellProps = {
  incidentId: string;
  apiBase?: string | undefined;
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
};

type WorkbookQueryEnvelope = {
  data: {
    incident_id: string;
    view_schema_id: string;
    rows: TimelineApiRow[];
  };
};

type TimelineMutationEnvelope = {
  data: {
    view_schema_id: string;
    change_set_id?: string;
    row: TimelineApiRow;
  };
};

type SameFieldConflictPayload = {
  conflict_token: string;
  record_id: string;
  field_key: string;
  conflict_resolution_class: string;
  base_row_version: number;
  current_row_version: number;
  client_value: unknown;
  server_value: unknown;
  base_value?: unknown;
  server_updated_by?: string;
  server_updated_at?: string;
  suggested_merged_value?: unknown;
};

type LocalConflictState = {
  key: string;
  conflict: SameFieldConflictPayload;
  focusKey: string;
  localValue: unknown;
  mergedDraft: string;
};

type PendingReplayKind = "create" | "patch";
type PendingReplayStatus = "queued" | "in_flight";

type PendingReplayUnit = {
  id: string;
  kind: PendingReplayKind;
  rowKey: string;
  recordId: string | null;
  focusField: FocusFieldKey;
  focusKey: string;
  surface: TimelineScalarEditorSurface;
  method: "POST" | "PATCH";
  path: string;
  payload: Record<string, unknown>;
  clientTxnId: string;
  mutationSignature: string;
  coalesceKey: string;
  enqueueOrder: number;
  status: PendingReplayStatus;
  rowSnapshot: WorkbookRow;
  continueOnFreshDraft: boolean;
  detectAutoResolution: boolean;
  promoteToCommittedRowInspect: boolean;
  viewportContinuityToken: number;
};

type PendingQueueRuntime = {
  units: PendingReplayUnit[];
  haltedMessage: string | null;
  authPaused: boolean;
  overflowMessage: string | null;
  resetRefreshInFlight: boolean;
  replayScheduled: boolean;
};

type PendingQueueSnapshot = {
  queuedCount: number;
  inFlightCount: number;
  haltedMessage: string | null;
  authPaused: boolean;
  overflowMessage: string | null;
  resetRefreshInFlight: boolean;
};

export const pendingReplayCapacity = 64;

type TimelineActionEnvelope = {
  data: {
    record_id: string;
    incident_id: string;
    row_version: number;
    capture_state: string;
    change_set_id: string;
    reason: string | null;
    replacement_record_id: string | null;
  };
};

export type RecordHistoryRollbackAction =
  | "change_set"
  | "history_entry"
  | "row_restore";

export type RecordHistoryItem = {
  actor_user_id: string;
  committed_at: string;
  operation: string;
  diff_summary: {
    summary: string;
    units: Array<Record<string, unknown>>;
  };
  change_set_id: string;
  reversible: boolean;
  available_rollback_actions: RecordHistoryRollbackAction[];
  history_entry_ref?: string;
  revision_no?: number;
};

type RecordHistoryData = {
  incident_id: string;
  record_id: string;
  row_version: number;
  deleted: boolean;
  items: RecordHistoryItem[];
};

type RecordHistoryEnvelope = {
  data: RecordHistoryData;
};

type RecordDeleteRestoreEnvelope = {
  data: {
    record_id: string;
    incident_id: string;
    row_version: number;
    deleted: boolean;
    deleted_at: string | null;
    deleted_by_user_id: string | null;
    change_set_id: string;
  };
};

type RecordRollbackEnvelope = {
  data: {
    incident_id: string;
    record_id: string;
    row_version: number;
    target: Record<string, unknown>;
    target_change_set_id: string;
    rollback_change_set_id: string;
    affected_record_ids: string[];
  };
};

type RecordHistoryState = {
  recordId: string | null;
  status: "idle" | "loading" | "ready" | "error";
  data: RecordHistoryData | null;
  message: string | null;
};

export function buildRecordRollbackTargetFromHistoryAction(
  item: RecordHistoryItem,
  action: RecordHistoryRollbackAction,
): Record<string, unknown> | null {
  if (!item.available_rollback_actions.includes(action)) {
    return null;
  }
  if (action === "history_entry") {
    return typeof item.history_entry_ref === "string" &&
      item.history_entry_ref !== ""
      ? { kind: "history_entry", history_entry_ref: item.history_entry_ref }
      : null;
  }
  if (action === "change_set") {
    return { kind: "change_set", change_set_id: item.change_set_id };
  }
  return typeof item.revision_no === "number"
    ? { kind: "row_restore", restore_to_revision_no: item.revision_no }
    : null;
}

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

type EvidenceHandleEnvelope = {
  data: {
    href: string;
    method: "GET";
    filename: string;
    preview_kind?: string | null;
    content_type?: string | null;
  };
};

type ObjectBlobCreateEnvelope = {
  data: {
    object_blob_id: string;
    upload_target: {
      href: string;
      method?: string | null;
    };
  };
};

type EvidenceAttachBlobEnvelope = {
  data: {
    record_id: string;
    row_version: number;
  };
};

type EvidencePreviewState = {
  href: string;
  recordId: string;
  title: string;
  previewKind: string | null;
};

type TimelineApiRow = {
  record_id: string;
  row_version: number;
  cells: Record<string, { value: unknown }>;
  group_values?: Record<string, unknown>;
};

type EntityApiRow = TimelineApiRow;

type WorkbookSheetRef = {
  id: string;
  kind: "saved_view" | "view_schema";
};

type WorkbookPresenceMode = "editing" | "idle" | "viewing";

type WorkbookPresenceInput = {
  sheet_ref: WorkbookSheetRef;
  mode: WorkbookPresenceMode;
  record_id?: string;
  field_key?: string;
};

type PresenceRecord = WorkbookPresenceInput & {
  connection_id: string;
  display_name: string;
  expires_at: string;
  observed_at: string;
  user_id: string;
};

type EntityRow = {
  entityType: "host" | "identity";
  recordId: string;
  rowVersion: number;
  label: string;
  secondaryText: string;
  state: string;
  aliasTexts: string[];
  linkedEventCount: number;
  rawRow: EntityApiRow;
  identifiers: Array<{
    key: string;
    label: string;
    value: string;
  }>;
};

type AssessmentCreateDraft = {
  assessedAt: string;
  assessmentState: string;
  confidenceBand: AssessmentConfidenceBand;
  rationale: string;
  subjectRecordId: string;
  subjectType: AssessmentSubjectType;
  supportRecordIds: string[];
};

type GenericReferenceOption = {
  recordId: string;
  label: string;
  viewSchemaId: string;
};

type GenericReferenceOptions = {
  parties: GenericReferenceOption[];
  taskRequests: GenericReferenceOption[];
  decisions: GenericReferenceOption[];
  evidence: GenericReferenceOption[];
  notes: GenericReferenceOption[];
  timeline: GenericReferenceOption[];
  allRecords: GenericReferenceOption[];
};

type GenericCollectionMode = "add" | "remove";

type ViewportContinuityFollowup = "none" | "entity-refresh";

type ViewportContinuityTarget =
  | { kind: "row-inspect"; recordId: string }
  | { kind: "input"; focusKey: string }
  | { kind: "scroll-only" };

type ViewportContinuityRequest = {
  token: number;
  attemptVersion: number;
  target: ViewportContinuityTarget;
  preservedViewport: ViewportSnapshot | null;
  followup: ViewportContinuityFollowup;
  followupSettled: boolean;
  baselineHostEntities: EntityRow[];
  baselineIdentityEntities: EntityRow[];
};

type DismissedMention = {
  rowRecordId: string;
  fieldKey: RelationshipFieldKey;
  entityType: "host" | "identity";
  itemRef: string;
  rawText: string;
  resolvedRecordId: string | null;
  resolutionMethod: string | null;
  autoResolved: boolean;
};

type InspectorMention = DismissedMention & {
  status: "unresolved" | "resolved" | "dismissed";
  displayText: string;
  provenance: string | null;
  confidence: number | null;
  matchedAliasText: string | null;
};

type AutoResolutionNotice = {
  itemRef: string;
  rowRecordId: string;
  fieldKey: RelationshipFieldKey;
  entityType: "host" | "identity";
  rawText: string;
  resolvedRecordId: string;
  matchedAliasText: string | null;
};

type LoadRowsOptions = {
  showLoading: boolean;
  viewportContinuityToken?: number;
};

type MergePlanLine = {
  label: string;
  outcome: string;
};

type TimelineScalarBinding = {
  kind: "scalar";
  fieldKey: EditableField;
  key: keyof RowValues;
  multiline?: boolean;
};

type TimelineCollectionBinding = {
  kind: "collection";
  fieldKey: CollectionFieldKey;
  draftKey: CollectionDraftKey;
  collectionKind: "relationship" | "tag";
  entityType?: "host" | "identity";
};

type TimelineReadonlyBinding = {
  kind: "readonly";
  fieldKey: string;
};

type TimelineFieldBinding =
  | TimelineScalarBinding
  | TimelineCollectionBinding
  | TimelineReadonlyBinding;
type TimelineScalarEditorSurface = "grid" | "inspector";
const timelineScalarEditorSurfaces: readonly TimelineScalarEditorSurface[] = [
  "grid",
  "inspector",
];

const timelineScalarBindingIndex: Record<EditableField, TimelineScalarBinding> =
  {
    "timeline.occurred_at": {
      kind: "scalar",
      fieldKey: "timeline.occurred_at",
      key: "occurredAt",
    },
    "timeline.summary": {
      kind: "scalar",
      fieldKey: "timeline.summary",
      key: "summary",
    },
    "timeline.details": {
      kind: "scalar",
      fieldKey: "timeline.details",
      key: "details",
      multiline: true,
    },
    "timeline.source_text": {
      kind: "scalar",
      fieldKey: "timeline.source_text",
      key: "sourceText",
      multiline: true,
    },
  };

const timelineCollectionBindingIndex: Record<
  CollectionFieldKey,
  TimelineCollectionBinding
> = {
  "timeline.host_refs": {
    kind: "collection",
    fieldKey: "timeline.host_refs",
    draftKey: "hostRefs",
    collectionKind: "relationship",
    entityType: "host",
  },
  "timeline.identity_refs": {
    kind: "collection",
    fieldKey: "timeline.identity_refs",
    draftKey: "identityRefs",
    collectionKind: "relationship",
    entityType: "identity",
  },
  "timeline.tags": {
    kind: "collection",
    fieldKey: "timeline.tags",
    draftKey: "tags",
    collectionKind: "tag",
  },
};

const timelineInspectorEditableFields: readonly EditableField[] = [
  "timeline.details",
  "timeline.source_text",
];

function timelineFieldBinding(fieldKey: string): TimelineFieldBinding {
  if (fieldKey in timelineScalarBindingIndex) {
    return timelineScalarBindingIndex[fieldKey as EditableField];
  }
  if (fieldKey in timelineCollectionBindingIndex) {
    return timelineCollectionBindingIndex[fieldKey as CollectionFieldKey];
  }
  return {
    kind: "readonly",
    fieldKey,
  };
}

const timelineVisibleBindings: readonly TimelineFieldBinding[] = visibleFields(
  timelineContract,
).map((field) => timelineFieldBinding(field.fieldKey));
const timelineInspectorBindings: readonly TimelineScalarBinding[] =
  timelineInspectorEditableFields.map(
    (fieldKey) => timelineFieldBinding(fieldKey) as TimelineScalarBinding,
  );

function timelineScalarBindingForField(
  fieldKey: string,
): TimelineScalarBinding | null {
  const binding = timelineFieldBinding(fieldKey);
  return binding.kind === "scalar" ? binding : null;
}

function timelineColumnWidth(fieldKey: string): number {
  switch (fieldKey) {
    case "timeline.occurred_at":
    case "timeline.edited_at":
      return 180;
    case "timeline.summary":
      return 320;
    case "timeline.host_refs":
      return 300;
    case "timeline.identity_refs":
      return 320;
    case "timeline.evidence_count":
      return 112;
    case "timeline.tags":
      return 240;
    default:
      return 224;
  }
}

const mergeIdentifierFields: Record<
  EntityRow["entityType"],
  Array<{ key: string; label: string }>
> = {
  host: [
    { key: "host.aad_device_id", label: "AAD Device ID" },
    { key: "host.fqdn", label: "FQDN" },
    { key: "host.hostname", label: "Hostname" },
  ],
  identity: [
    { key: "identity.aad_object_id", label: "AAD Object ID" },
    { key: "identity.sid", label: "SID" },
    { key: "identity.upn", label: "UPN" },
    { key: "identity.email", label: "Email" },
    { key: "identity.sam_account_name", label: "SAM Account Name" },
  ],
};

function emptyValues(): RowValues {
  return {
    occurredAt: "",
    summary: "",
    details: "",
    sourceText: "",
  };
}

function emptyCollectionDrafts(): CollectionDrafts {
  return {
    hostRefs: "",
    identityRefs: "",
    tags: "",
  };
}

export function createDraftRow(index: number): WorkbookRow {
  return {
    key: `draft-${index}`,
    recordId: null,
    rowVersion: null,
    captureState: "rough",
    values: emptyValues(),
    committedValues: emptyValues(),
    collectionValues: {
      hostRefs: [],
      identityRefs: [],
      tags: [],
    },
    collectionDrafts: emptyCollectionDrafts(),
    pendingSignature: null,
    rawRow: null,
  };
}

function normalizeValue(value: string): string {
  return value.trim();
}

function readStringCell(
  row: TimelineApiRow | EntityApiRow,
  fieldKey: string,
): string {
  const raw = row.cells[fieldKey]?.value;
  return typeof raw === "string" ? raw : "";
}

function readNumberCell(row: EntityApiRow, fieldKey: string): number {
  const raw = row.cells[fieldKey]?.value;
  return typeof raw === "number" ? raw : 0;
}

function readCellValue(
  row: TimelineApiRow | EntityApiRow | null,
  fieldKey: string,
): unknown {
  return row?.cells[fieldKey]?.value ?? null;
}

function stringifyGridValue(value: unknown): string {
  if (typeof value === "string") {
    return value;
  }
  if (typeof value === "boolean") {
    return value ? "true" : "false";
  }
  if (typeof value === "number") {
    return String(value);
  }
  if (
    value &&
    typeof value === "object" &&
    !Array.isArray(value) &&
    "items" in value &&
    Array.isArray(value.items)
  ) {
    return value.items
      .map((item) => {
        if (!item || typeof item !== "object") {
          return null;
        }
        const object = item as Record<string, unknown>;
        return typeof object.display_text === "string"
          ? object.display_text
          : typeof object.raw_text === "string"
            ? object.raw_text
            : null;
      })
      .filter((item): item is string => item !== null)
      .join(", ");
  }
  return "";
}

function readTagItems(row: TimelineApiRow): TagCollectionItem[] {
  const raw = row.cells["timeline.tags"]?.value;
  const value =
    raw &&
    typeof raw === "object" &&
    !Array.isArray(raw) &&
    "items" in raw &&
    Array.isArray(raw.items)
      ? raw.items
      : [];
  return value
    .map((item, index) => {
      if (!item || typeof item !== "object") {
        return null;
      }
      const object = item as Record<string, unknown>;
      const rawText =
        typeof object.raw_text === "string"
          ? object.raw_text
          : typeof object.display_text === "string"
            ? object.display_text
            : "";
      if (rawText === "") {
        return null;
      }
      return {
        itemRef:
          typeof object.item_ref === "string"
            ? object.item_ref
            : `tag-item-${index}:${rawText}`,
        itemKind:
          typeof object.item_kind === "string" ? object.item_kind : "tag",
        displayText:
          typeof object.display_text === "string"
            ? object.display_text
            : rawText,
        rawText,
      };
    })
    .filter((item): item is TagCollectionItem => item !== null);
}

function timelineGroupLabel(row: WorkbookRow, fieldKey: string) {
  const value = stringifyGridValue(readCellValue(row.rawRow, fieldKey)).trim();
  return value === "" ? "Unassigned" : value;
}

function entityGroupLabel(row: EntityRow, fieldKey: string) {
  const value = stringifyGridValue(readCellValue(row.rawRow, fieldKey)).trim();
  return value === "" ? "Unassigned" : value;
}

function rowFromApi(row: TimelineApiRow): WorkbookRow {
  const values: RowValues = {
    occurredAt: readStringCell(row, "timeline.occurred_at"),
    summary: readStringCell(row, "timeline.summary"),
    details: readStringCell(row, "timeline.details"),
    sourceText: readStringCell(row, "timeline.source_text"),
  };

  return {
    key: row.record_id,
    recordId: row.record_id,
    rowVersion: row.row_version,
    captureState: readStringCell(row, "timeline.capture_state"),
    values,
    committedValues: values,
    collectionValues: {
      hostRefs: readCollectionItems(row, "timeline.host_refs"),
      identityRefs: readCollectionItems(row, "timeline.identity_refs"),
      tags: readTagItems(row),
    },
    collectionDrafts: emptyCollectionDrafts(),
    pendingSignature: null,
    rawRow: row,
  };
}

function applyViewRowPatch(
  row: TimelineApiRow,
  patch: NonNullable<
    RecordChangedPayload["affected_views"][number]["patch_cells"]
  >,
): TimelineApiRow {
  return {
    ...row,
    row_version: patch.row_version,
    cells: {
      ...row.cells,
      ...patch.cells,
    },
    ...(patch.group_values
      ? {
          group_values: {
            ...(row.group_values ?? {}),
            ...patch.group_values,
          },
        }
      : {}),
  };
}

function presenceMatchesSheet(
  presence: PresenceRecord,
  sheetRef: WorkbookSheetRef,
) {
  return (
    presence.sheet_ref.kind === sheetRef.kind &&
    presence.sheet_ref.id === sheetRef.id
  );
}

function displayInitials(displayName: string) {
  const parts = displayName.trim().split(/\s+/u).filter(Boolean);
  if (parts.length === 0) {
    return "?";
  }
  return parts
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? "")
    .join("");
}

function visiblePresence(records: readonly PresenceRecord[], limit: number) {
  return {
    shown: records.slice(0, limit),
    overflow: Math.max(0, records.length - limit),
  };
}

function resolveGridScrollElement(
  element: HTMLElement,
  surface: WorkbookSurface,
): HTMLElement {
  const selector = gridScrollportSelector();
  const scrollports = Array.from(
    element.querySelectorAll<HTMLElement>(selector),
  );
  if (scrollports.length !== 1) {
    throw new Error(
      `Expected ${surface} grid shell to contain exactly one ${selector} scrollport, received ${scrollports.length}`,
    );
  }
  const scrollport = scrollports[0];
  if (scrollport === undefined) {
    throw new Error(
      `Expected ${surface} grid shell to contain exactly one ${selector} scrollport, received 0`,
    );
  }
  return scrollport;
}

function isPresenceRecord(value: unknown): value is PresenceRecord {
  if (!value || typeof value !== "object") {
    return false;
  }
  const record = value as Record<string, unknown>;
  const sheetRef = record.sheet_ref;
  return (
    typeof record.connection_id === "string" &&
    typeof record.user_id === "string" &&
    typeof record.display_name === "string" &&
    typeof record.mode === "string" &&
    (record.mode === "viewing" ||
      record.mode === "editing" ||
      record.mode === "idle") &&
    typeof record.observed_at === "string" &&
    typeof record.expires_at === "string" &&
    !!sheetRef &&
    typeof sheetRef === "object" &&
    !Array.isArray(sheetRef) &&
    ((sheetRef as Record<string, unknown>).kind === "view_schema" ||
      (sheetRef as Record<string, unknown>).kind === "saved_view") &&
    typeof (sheetRef as Record<string, unknown>).id === "string" &&
    (record.record_id === undefined || typeof record.record_id === "string") &&
    (record.field_key === undefined || typeof record.field_key === "string")
  );
}

function socketIsOpen(socket: WebSocket) {
  return socket.readyState === WebSocket.OPEN;
}

function entityRowFromApi(
  row: EntityApiRow,
  entityType: EntityRow["entityType"],
): EntityRow {
  const labelField =
    entityType === "host" ? "host.display_name" : "identity.display_name";
  const secondaryCandidates =
    entityType === "host"
      ? ["host.hostname", "host.fqdn"]
      : ["identity.email", "identity.upn", "identity.sam_account_name"];
  const stateField =
    entityType === "host" ? "host.host_state" : "identity.identity_state";
  const aliasesField =
    entityType === "host" ? "host.aliases" : "identity.aliases";
  const linkedEventField =
    entityType === "host"
      ? "host.linked_event_count"
      : "identity.linked_event_count";
  const identifiers = mergeIdentifierFields[entityType]
    .map((field) => {
      const value = readStringCell(row, field.key);
      if (value === "") {
        return null;
      }
      return {
        key: field.key,
        label: field.label,
        value,
      };
    })
    .filter(
      (
        value,
      ): value is {
        key: string;
        label: string;
        value: string;
      } => value !== null,
    );
  const aliasItems = (() => {
    const raw = row.cells[aliasesField]?.value;
    if (
      !raw ||
      typeof raw !== "object" ||
      Array.isArray(raw) ||
      !("items" in raw) ||
      !Array.isArray(raw.items)
    ) {
      return [] as string[];
    }
    return raw.items
      .map((item) => {
        if (!item || typeof item !== "object") {
          return null;
        }
        const object = item as Record<string, unknown>;
        return typeof object.raw_text === "string" ? object.raw_text : null;
      })
      .filter((value): value is string => value !== null);
  })();
  const secondaryText =
    secondaryCandidates
      .map((field) => readStringCell(row, field))
      .find((value) => value !== "") ?? "";
  const label =
    readStringCell(row, labelField) || secondaryText || row.record_id;

  return {
    entityType,
    recordId: row.record_id,
    rowVersion: row.row_version,
    label,
    secondaryText,
    state: readStringCell(row, stateField),
    aliasTexts: aliasItems,
    linkedEventCount: readNumberCell(row, linkedEventField),
    rawRow: row,
    identifiers,
  };
}

export function ensureDraftRow(
  rows: WorkbookRow[],
  nextDraftIndex: number,
): WorkbookRow[] {
  if (rows.some((row) => row.recordId === null)) {
    return rows;
  }
  return [...rows, createDraftRow(nextDraftIndex)];
}

function ensureDraftRowWithFreshIndex(
  rows: WorkbookRow[],
  nextDraftIndex: () => number,
): {
  rows: WorkbookRow[];
  draftSummaryKey: string | null;
} {
  if (rows.some((row) => row.recordId === null)) {
    return {
      rows,
      draftSummaryKey: null,
    };
  }

  const draftIndex = nextDraftIndex();
  return {
    rows: [...rows, createDraftRow(draftIndex)],
    draftSummaryKey: inputFocusKey(`draft-${draftIndex}`, "summary"),
  };
}

function rowWithMaterializedScalarDrafts(
  row: WorkbookRow,
  draftValueForFocusKey: (focusKey: string) => string | undefined,
  preferred?: {
    readonly field: keyof RowValues;
    readonly value: string | undefined;
  },
): WorkbookRow {
  let nextValues: RowValues | null = null;
  for (const binding of Object.values(timelineScalarBindingIndex)) {
    let draftValue =
      preferred?.field === binding.key ? preferred.value : undefined;
    if (draftValue === undefined) {
      for (const surface of timelineScalarEditorSurfaces) {
        draftValue = draftValueForFocusKey(
          inputFocusKey(row.key, binding.key, surface),
        );
        if (draftValue !== undefined) {
          break;
        }
      }
    }
    if (draftValue === undefined || draftValue === row.values[binding.key]) {
      continue;
    }
    nextValues ??= { ...row.values };
    nextValues[binding.key] = draftValue;
  }
  return nextValues === null ? row : { ...row, values: nextValues };
}

function reconcileCommittedRowsWithLocalDrafts({
  currentRows,
  incomingRows,
  draftValueForFocusKey,
  nextDraftIndex,
}: {
  readonly currentRows: WorkbookRow[];
  readonly incomingRows: WorkbookRow[];
  readonly draftValueForFocusKey: (focusKey: string) => string | undefined;
  readonly nextDraftIndex: () => number;
}): {
  readonly committedRows: WorkbookRow[];
  readonly rows: WorkbookRow[];
} {
  const committedRows = [
    ...reconcileRecordRows(
      currentRows.filter((row) => row.recordId !== null),
      incomingRows,
    ),
  ];
  const localDraftRows = currentRows
    .filter((row) => row.recordId === null)
    .map((row) => rowWithMaterializedScalarDrafts(row, draftValueForFocusKey));

  return {
    committedRows,
    rows: ensureDraftRowWithFreshIndex(
      [...committedRows, ...localDraftRows],
      nextDraftIndex,
    ).rows,
  };
}

function buildCollectionActions(rawInput: string) {
  const actions = rawInput
    .split(/\r?\n/u)
    .filter((segment) => segment.trim() !== "")
    .map((rawText) => ({
      op: "add_token",
      raw_text: rawText,
    }));
  if (actions.length < 1) {
    return null;
  }
  return {
    kind: "collection_actions_v1",
    actions,
  };
}

type BuildCreatePayloadOptions = {
  allowZeroFieldCreate?: boolean;
};

export function buildCreatePayload(
  row: WorkbookRow,
  clientTxnId: string,
  options: BuildCreatePayloadOptions = {},
) {
  const payload: Record<string, unknown> = {
    client_txn_id: clientTxnId,
  };

  for (const field of Object.values(timelineScalarBindingIndex)) {
    const normalized = normalizeValue(row.values[field.key]);
    if (normalized !== "") {
      payload[field.fieldKey] = normalized;
    }
  }

  for (const field of Object.values(timelineCollectionBindingIndex)) {
    const actions = buildCollectionActions(
      row.collectionDrafts[field.draftKey],
    );
    if (actions !== null) {
      payload[field.fieldKey] = actions;
    }
  }

  if (Object.keys(payload).length < 2 && !options.allowZeroFieldCreate) {
    return null;
  }
  return payload;
}

export function confidenceScoreFromBand(
  band: AssessmentConfidenceBand,
): number | null {
  switch (band) {
    case "low":
      return 25;
    case "medium":
      return 55;
    case "high":
      return 85;
    default:
      return null;
  }
}

export function buildAssessmentCreatePayload(
  draft: AssessmentCreateDraft,
  clientTxnId: string,
): Record<string, unknown> | null {
  const subjectRecordId = normalizeValue(draft.subjectRecordId);
  const assessmentState = normalizeValue(draft.assessmentState);
  const rationale = normalizeValue(draft.rationale);
  if (subjectRecordId === "" || assessmentState === "" || rationale === "") {
    return null;
  }

  const payload: Record<string, unknown> = {
    client_txn_id: clientTxnId,
    "assessment.subject_ref": subjectRecordId,
    "assessment.subject_type": draft.subjectType,
    "assessment.assessment_state": assessmentState,
    "assessment.confidence_score": confidenceScoreFromBand(
      draft.confidenceBand,
    ),
    "assessment.rationale": rationale,
  };

  const assessedAt = normalizeValue(draft.assessedAt);
  if (assessedAt !== "") {
    payload["assessment.assessed_at"] = assessedAt;
  }

  const supportRecordIds = Array.from(
    new Set(
      draft.supportRecordIds
        .map((recordId) => normalizeValue(recordId))
        .filter((recordId) => recordId !== ""),
    ),
  );
  if (supportRecordIds.length > 0) {
    payload["assessment.support_refs"] = {
      kind: "collection_actions_v1",
      actions: supportRecordIds.map((recordId) => ({
        op: "add_record_ref",
        linked_record_id: recordId,
      })),
    };
  }

  return payload;
}

function buildScalarPatchPayload(row: WorkbookRow, clientTxnId: string) {
  const changes = Object.values(timelineScalarBindingIndex)
    .map((field) => {
      const current = normalizeValue(row.values[field.key]);
      const committed = normalizeValue(row.committedValues[field.key]);
      if (current === committed) {
        return null;
      }
      return {
        field_key: field.fieldKey,
        value: current === "" ? null : current,
      };
    })
    .filter(
      (change): change is { field_key: EditableField; value: string | null } =>
        change !== null,
    )
    .sort((left, right) => left.field_key.localeCompare(right.field_key));

  if (changes.length < 1 || row.rowVersion === null) {
    return null;
  }

  return {
    view_schema_id: timelineViewSchemaId,
    base_row_version: row.rowVersion,
    client_txn_id: clientTxnId,
    changes,
  };
}

function buildCollectionPatchPayload(
  row: WorkbookRow,
  fieldKey: CollectionFieldKey,
  draftValue: string,
  clientTxnId: string,
) {
  const actionPayload = buildCollectionActions(draftValue);
  if (row.rowVersion === null || actionPayload === null) {
    return null;
  }

  return {
    view_schema_id: timelineViewSchemaId,
    base_row_version: row.rowVersion,
    client_txn_id: clientTxnId,
    changes: [
      {
        field_key: fieldKey,
        action_payload: actionPayload,
      },
    ],
  };
}

function buildAttachedEvidenceCreatePayload(
  evidenceRecordId: string,
  clientTxnId: string,
) {
  return {
    client_txn_id: clientTxnId,
    "timeline.attached_evidence_ids": {
      kind: "collection_actions_v1",
      actions: [
        {
          op: "add_record_ref",
          linked_record_id: evidenceRecordId,
        },
      ],
    },
  };
}

function buildAttachedEvidencePatchPayload(
  row: WorkbookRow,
  evidenceRecordId: string,
  clientTxnId: string,
) {
  if (row.rowVersion === null) {
    return null;
  }
  return {
    view_schema_id: timelineViewSchemaId,
    base_row_version: row.rowVersion,
    client_txn_id: clientTxnId,
    changes: [
      {
        field_key: "timeline.attached_evidence_ids",
        action_payload: {
          kind: "collection_actions_v1",
          actions: [
            {
              op: "add_record_ref",
              linked_record_id: evidenceRecordId,
            },
          ],
        },
      },
    ],
  };
}

// Dedup queued autosaves by the logical mutation payload, not the per-request txn id.
function buildStableMutationSignature(payload: unknown): string {
  if (!payload || typeof payload !== "object" || Array.isArray(payload)) {
    return JSON.stringify(payload);
  }

  const { client_txn_id: _clientTxnID, ...stablePayload } = payload as Record<
    string,
    unknown
  >;
  return JSON.stringify(stablePayload);
}

function readEnvelope<T>(payload: unknown): T {
  return payload as T;
}

function readCookie(name: string): string | null {
  if (typeof document === "undefined") {
    return null;
  }

  const prefix = `${name}=`;
  for (const segment of document.cookie.split(";")) {
    const trimmed = segment.trim();
    if (trimmed.startsWith(prefix)) {
      return decodeURIComponent(trimmed.slice(prefix.length));
    }
  }
  return null;
}

async function fetchJSON<T>(
  input: RequestInfo | URL,
  init?: RequestInit,
): Promise<{
  ok: boolean;
  status: number;
  payload:
    | T
    | { error?: { code?: string; message?: string; details?: unknown } };
}> {
  const method = (init?.method ?? "GET").toUpperCase();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(init?.headers as Record<string, string> | undefined),
  };
  if (method !== "GET" && method !== "HEAD" && method !== "OPTIONS") {
    const csrfToken = readCookie(csrfCookieName);
    if (csrfToken !== null && csrfToken !== "") {
      headers[csrfHeaderName] = csrfToken;
    }
  }

  const response = await fetch(input, {
    credentials: "include",
    ...init,
    headers,
  });
  const payload = (await response.json()) as
    | T
    | { error?: { code?: string; message?: string; details?: unknown } };
  return { ok: response.ok, status: response.status, payload };
}

function apiPath(base: string | undefined, path: string): string {
  const trimmedBase = (base ?? "").trim();
  if (trimmedBase === "") {
    return path;
  }
  return `${trimmedBase.replace(/\/$/, "")}${path}`;
}

function websocketPath(base: string | undefined, path: string): string {
  const trimmedBase = (base ?? "").trim();
  if (trimmedBase === "") {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    return `${protocol}//${window.location.host}${path}`;
  }

  const target = new URL(trimmedBase);
  target.protocol = target.protocol === "https:" ? "wss:" : "ws:";
  target.pathname = path;
  target.search = "";
  target.hash = "";
  return target.toString();
}

function tabClientInstanceId(): string {
  const key = "cartulary.client_instance_id";
  try {
    const existing = window.sessionStorage.getItem(key);
    if (existing) {
      return existing;
    }
    const created =
      window.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random()}`;
    window.sessionStorage.setItem(key, created);
    return created;
  } catch {
    return `${Date.now()}-${Math.random()}`;
  }
}

function workbookPresence(
  presence: {
    fieldKey: string | null;
    mode: WorkbookPresenceMode;
    recordId: string | null;
  } = { fieldKey: null, mode: "viewing", recordId: null },
): WorkbookPresenceInput {
  const input: WorkbookPresenceInput = {
    sheet_ref: {
      kind: "view_schema",
      id: timelineViewSchemaId,
    },
    mode: presence.mode,
  };
  if (presence.recordId !== null) {
    input.record_id = presence.recordId;
  }
  if (presence.mode === "editing" && presence.fieldKey !== null) {
    input.field_key = presence.fieldKey;
  }
  return input;
}

function inputFocusKey(
  rowKey: string,
  field: FocusFieldKey,
  surface: TimelineScalarEditorSurface = "grid",
) {
  return `${rowKey}:${field}:${surface}`;
}

function sanitizeTestId(value: string) {
  return value.replace(/[^a-zA-Z0-9_-]+/gu, "-");
}

function focusTestIdForKey(focusKey: string) {
  const [rowKey, fieldKey, surface = "grid"] = focusKey.split(":");
  if (!rowKey || !fieldKey) {
    return null;
  }
  const fieldSuffix =
    fieldKey === "hostRefs" || fieldKey === "identityRefs"
      ? `${fieldKey}-input`
      : fieldKey;
  const surfaceSuffix = surface === "inspector" ? "-inspector" : "";
  if (rowKey.startsWith("draft-")) {
    return `draft-row-${fieldSuffix}${surfaceSuffix}`;
  }
  return `row-${rowKey}-${fieldSuffix}${surfaceSuffix}`;
}

function relationshipItemLabel(
  item: CollectionItem | InspectorMention,
  entityIndex: Record<string, EntityRow>,
) {
  if ("status" in item && item.status === "dismissed") {
    return item.displayText || item.rawText;
  }
  if (item.resolvedRecordId) {
    const resolvedEntity = entityIndex[item.resolvedRecordId];
    if (resolvedEntity) {
      return resolvedEntity.label;
    }
  }
  return item.displayText || item.rawText;
}

function timelineRelationshipLabel(fieldKey: RelationshipFieldKey) {
  return fieldKey === "timeline.identity_refs" ? "Identities" : "Hosts";
}

function pruneDismissedMentions(
  dismissedMentionsByRow: Record<string, DismissedMention[]>,
  row: WorkbookRow,
) {
  if (row.recordId === null) {
    return dismissedMentionsByRow;
  }

  const activeRefs = new Set(
    [
      ...row.collectionValues.hostRefs,
      ...row.collectionValues.identityRefs,
    ].map((item) => item.itemRef),
  );
  const next = { ...dismissedMentionsByRow };
  const current = next[row.recordId] ?? [];
  const remaining = current.filter((item) => !activeRefs.has(item.itemRef));
  if (remaining.length < 1) {
    delete next[row.recordId];
    return next;
  }
  next[row.recordId] = remaining;
  return next;
}

function compareValue(value: string) {
  return value.trim().toLowerCase();
}

function buildMergePlan(survivor: EntityRow, loser: EntityRow) {
  const identifierLines: MergePlanLine[] = mergeIdentifierFields[
    survivor.entityType
  ].flatMap((field) => {
    const survivorValue =
      survivor.identifiers.find((identifier) => identifier.key === field.key)
        ?.value ?? "";
    const loserValue =
      loser.identifiers.find((identifier) => identifier.key === field.key)
        ?.value ?? "";
    if (survivorValue === "" && loserValue === "") {
      return [];
    }
    if (survivorValue === "" && loserValue !== "") {
      return [{ label: field.label, outcome: `Promote ${loserValue}` }];
    }
    if (
      survivorValue !== "" &&
      loserValue !== "" &&
      compareValue(survivorValue) === compareValue(loserValue)
    ) {
      return [{ label: field.label, outcome: `Duplicate no-op ${loserValue}` }];
    }
    if (survivorValue !== "" && loserValue !== "") {
      return [
        {
          label: field.label,
          outcome: `Conflict survivor=${survivorValue} loser=${loserValue}`,
        },
      ];
    }
    return [{ label: field.label, outcome: `Survivor keeps ${survivorValue}` }];
  });

  const survivorAliases = new Set(survivor.aliasTexts.map(compareValue));
  const aliasesToCopy = loser.aliasTexts.filter(
    (value) => !survivorAliases.has(compareValue(value)),
  );
  const duplicateAliases = loser.aliasTexts.filter((value) =>
    survivorAliases.has(compareValue(value)),
  );

  return {
    identifierLines,
    aliasesToCopy,
    duplicateAliases,
    provenanceOnlySummary: "Not exposed on this surface.",
    dependencySummary:
      survivor.linkedEventCount > 0 || loser.linkedEventCount > 0
        ? `Linked events visible on surface: survivor=${survivor.linkedEventCount}, loser=${loser.linkedEventCount}.`
        : "Dependency counts are not exposed on this surface.",
  };
}

function parseErrorMessage(payload: unknown) {
  if (!payload || typeof payload !== "object" || !("error" in payload)) {
    return "Request failed.";
  }
  const error = payload.error;
  if (!error || typeof error !== "object") {
    return "Request failed.";
  }
  if ("code" in error && typeof error.code === "string") {
    if (
      "details" in error &&
      error.details &&
      typeof error.details === "object" &&
      "reason_code" in error.details &&
      typeof error.details.reason_code === "string"
    ) {
      return `${error.code}: ${error.details.reason_code}`;
    }
    return error.code;
  }
  if ("message" in error && typeof error.message === "string") {
    return error.message;
  }
  return "Request failed.";
}

function formatHistoryTimestamp(value: string): string {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toISOString();
}

function RelationshipChip({
  item,
  entityIndex,
  onSelect,
  selected = false,
}: {
  item: CollectionItem | InspectorMention;
  entityIndex: Record<string, EntityRow>;
  onSelect?: () => void;
  selected?: boolean;
}) {
  const label = relationshipItemLabel(item, entityIndex);
  const isInspectorItem = "status" in item;
  const isResolved = isInspectorItem
    ? item.status === "resolved"
    : item.itemKind === "resolved_ref";
  const isDismissed = isInspectorItem ? item.status === "dismissed" : false;
  const isAutoResolved = item.autoResolved;
  const chipStyle = {
    ...relationshipChipStyle,
    ...(isDismissed
      ? dismissedChipStyle
      : isResolved
        ? isAutoResolved
          ? autoResolvedChipStyle
          : resolvedChipStyle
        : unresolvedChipStyle),
    ...(selected ? selectedChipStyle : null),
  };
  const labelPrefix = isDismissed
    ? "Dismissed"
    : isResolved
      ? isAutoResolved
        ? "Auto-resolved"
        : "Resolved"
      : "Unresolved";

  return onSelect ? (
    <button
      aria-label={`${labelPrefix} ${label}`}
      data-testid={`chip-${sanitizeTestId(item.itemRef)}`}
      style={chipStyle}
      type="button"
      onClick={onSelect}
    >
      <span>{label}</span>
      {isAutoResolved ? <span style={chipMetaStyle}>Auto</span> : null}
      {!isResolved && !isDismissed ? (
        <span style={chipMetaStyle}>Mention</span>
      ) : null}
    </button>
  ) : (
    <span
      aria-label={`${labelPrefix} ${label}`}
      data-testid={`chip-${sanitizeTestId(item.itemRef)}`}
      role="note"
      style={chipStyle}
    >
      <span>{label}</span>
      {isAutoResolved ? <span style={chipMetaStyle}>Auto</span> : null}
    </span>
  );
}

function DraftRowCreateButton({
  onCreate,
  row,
}: {
  readonly onCreate: (row: WorkbookRow) => void;
  readonly row: WorkbookRow;
}) {
  const createBlankRow = (
    event:
      | ReactKeyboardEvent<HTMLButtonElement>
      | ReactMouseEvent<HTMLButtonElement>,
  ) => {
    if (event.currentTarget.disabled) {
      return;
    }
    event.preventDefault();
    event.stopPropagation();
    onCreate(row);
  };

  return (
    <button
      data-testid="draft-row-create"
      disabled={row.pendingSignature !== null}
      style={actionButtonStyle}
      type="button"
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") {
          createBlankRow(event);
        }
      }}
      onMouseDown={createBlankRow}
    >
      Create blank row
    </button>
  );
}

function TimelineScalarEditor({
  accessibleLabel,
  blockedByConflict,
  committedValue,
  controlId,
  dataTestId,
  draftValue,
  field,
  multiline,
  onBlurCommit,
  onDraftChange,
  onEditModeChange,
  onFocusRecord,
  onKeyCommit,
  onPasteCommit,
  registerInput,
  presenceFieldKey,
  rowKey,
  rowRecordId,
  surface,
}: {
  readonly accessibleLabel?: string | undefined;
  readonly blockedByConflict?: boolean | undefined;
  readonly committedValue: string;
  readonly controlId: string;
  readonly dataTestId: string;
  readonly draftValue?: string | undefined;
  readonly field: keyof RowValues;
  readonly multiline?: boolean | undefined;
  readonly onBlurCommit: (
    rowKey: string,
    field: keyof RowValues,
    surface: TimelineScalarEditorSurface,
    value: string,
  ) => void;
  readonly onDraftChange: (
    rowKey: string,
    field: keyof RowValues,
    surface: TimelineScalarEditorSurface,
    value: string,
  ) => void;
  readonly onEditModeChange: (
    recordId: string | null,
    fieldKey: string,
    editing: boolean,
  ) => void;
  readonly onFocusRecord: (recordId: string) => void;
  readonly onKeyCommit: (
    event: ReactKeyboardEvent<HTMLInputElement | HTMLTextAreaElement>,
    rowKey: string,
    field: keyof RowValues,
    surface: TimelineScalarEditorSurface,
  ) => void;
  readonly onPasteCommit: (
    rowKey: string,
    field: keyof RowValues,
    surface: TimelineScalarEditorSurface,
  ) => void;
  readonly registerInput: (
    rowKey: string,
    field: FocusFieldKey,
    surface: TimelineScalarEditorSurface,
    element: HTMLInputElement | HTMLTextAreaElement | null,
  ) => void;
  readonly presenceFieldKey: string;
  readonly rowKey: string;
  readonly rowRecordId: string | null;
  readonly surface: TimelineScalarEditorSurface;
}) {
  const displayValue = draftValue ?? committedValue;
  const [editorValue, setEditorValue] = useState(displayValue);
  const hasActiveEditRef = useRef(false);

  useEffect(() => {
    if (!hasActiveEditRef.current || draftValue === undefined) {
      setEditorValue(displayValue);
    }
  }, [displayValue, draftValue]);

  const handleFocus = () => {
    hasActiveEditRef.current = true;
    if (rowRecordId) {
      onFocusRecord(rowRecordId);
    }
    onEditModeChange(rowRecordId, presenceFieldKey, true);
  };
  const handleChange = (value: string) => {
    setEditorValue(value);
    onDraftChange(rowKey, field, surface, value);
  };
  const handleBlur = (
    event: ReactFocusEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    hasActiveEditRef.current = false;
    onEditModeChange(rowRecordId, presenceFieldKey, false);
    onDraftChange(rowKey, field, surface, event.currentTarget.value);
    if (blockedByConflict) {
      return;
    }
    onBlurCommit(rowKey, field, surface, event.currentTarget.value);
  };
  const handleKeyDown = (
    event: ReactKeyboardEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    onKeyCommit(event, rowKey, field, surface);
  };
  const handlePaste = () => {
    onPasteCommit(rowKey, field, surface);
  };
  const inputRef = (element: HTMLInputElement | HTMLTextAreaElement | null) => {
    registerInput(rowKey, field, surface, element);
  };

  if (multiline) {
    return (
      <textarea
        aria-label={accessibleLabel}
        data-testid={dataTestId}
        id={controlId}
        ref={inputRef}
        rows={3}
        style={textareaStyle}
        value={editorValue}
        onBlur={handleBlur}
        onChange={(event) => {
          handleChange(event.target.value);
        }}
        onFocus={handleFocus}
        onKeyDown={handleKeyDown}
        onPaste={handlePaste}
      />
    );
  }

  return (
    <input
      aria-label={accessibleLabel}
      data-testid={dataTestId}
      id={controlId}
      ref={inputRef}
      style={inputStyle}
      type="text"
      value={editorValue}
      onBlur={handleBlur}
      onChange={(event) => {
        handleChange(event.target.value);
      }}
      onFocus={handleFocus}
      onKeyDown={handleKeyDown}
      onPaste={handlePaste}
    />
  );
}

export function TimelineWorkbook({
  incidentId,
  apiBase,
  hostEntities = [],
  identityEntities = [],
  entityIndex = {},
  currentIncidentRole = "",
  onRefreshEntities,
}: TimelineWorkbookProps) {
  const [rows, setRows] = useState<WorkbookRow[]>(() => [createDraftRow(1)]);
  const [isInitialLoading, setIsInitialLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [refreshError, setRefreshError] = useState<string | null>(null);
  const [saveState, setSaveState] = useState<SaveState>("Saved");
  const [conflictQueue, setConflictQueue] = useState<
    Record<string, LocalConflictState>
  >({});
  const [activeConflictKey, setActiveConflictKey] = useState<string | null>(
    null,
  );
  const [replacementDrafts, setReplacementDrafts] = useState<
    Record<string, string>
  >({});
  const [selectedRowId, setSelectedRowId] = useState<string | null>(null);
  const [selectedMentionRef, setSelectedMentionRef] = useState<string | null>(
    null,
  );
  const [currentPresence, setCurrentPresence] = useState<{
    fieldKey: string | null;
    mode: WorkbookPresenceMode;
    recordId: string | null;
  }>({ fieldKey: null, mode: "viewing", recordId: null });
  const [presenceRecords, setPresenceRecords] = useState<PresenceRecord[]>([]);
  const [selectedResolveTargetId, setSelectedResolveTargetId] = useState("");
  const [inspectorMessage, setInspectorMessage] = useState<string | null>(null);
  const [rowHistory, setRowHistory] = useState<RecordHistoryState>({
    recordId: null,
    status: "idle",
    data: null,
    message: null,
  });
  const [dismissedMentionsByRow, setDismissedMentionsByRow] = useState<
    Record<string, DismissedMention[]>
  >({});
  const [autoResolutionNotices, setAutoResolutionNotices] = useState<
    AutoResolutionNotice[]
  >([]);
  const [queryState, setQueryState] = useState<WorkbookQueryState>(() =>
    emptyWorkbookQueryState(),
  );
  const [filterDraft, setFilterDraft] = useState<FilterDraft>(() =>
    defaultFilterDraft(timelineContract),
  );
  const draftCounterRef = useRef(2);
  const clientTxnRef = useRef(1);
  const pendingOpsRef = useRef(0);
  const pendingSignaturesRef = useRef(new Map<string, string>());
  const pendingSocketTxnTimeoutsRef = useRef(new Map<string, number>());
  const saveQueueRef = useRef(Promise.resolve());
  const pendingQueueRef = useRef<PendingQueueRuntime>({
    units: [],
    haltedMessage: null,
    authPaused: false,
    overflowMessage: null,
    resetRefreshInFlight: false,
    replayScheduled: false,
  });
  const [pendingQueueSnapshot, setPendingQueueSnapshot] =
    useState<PendingQueueSnapshot>({
      queuedCount: 0,
      inFlightCount: 0,
      haltedMessage: null,
      authPaused: false,
      overflowMessage: null,
      resetRefreshInFlight: false,
    });
  const pendingReplayOrderRef = useRef(1);
  const pendingReplayTimerRef = useRef<number | null>(null);
  const pendingReplayAuthRetryRef = useRef<number | null>(null);
  const appliedStreamSeqRef = useRef(new Set<number>());
  const rowsRef = useRef(rows);
  const conflictQueueRef = useRef<Record<string, LocalConflictState>>({});
  const activeSocketRef = useRef<WebSocket | null>(null);
  const socketEstablishedRef = useRef(false);
  const socketConnectionIDRef = useRef<string | null>(null);
  const presenceUpdateTimerRef = useRef<number | null>(null);
  const currentPresenceRef = useRef(currentPresence);
  const hasLoadedRowsRef = useRef(false);
  const loadSequenceRef = useRef(0);
  const scalarDraftValuesRef = useRef(new Map<string, string>());
  const loadRowsRef = useRef<(options: LoadRowsOptions) => Promise<void>>(
    async () => undefined,
  );
  const socketResumeTokenRef = useRef<string | null>(null);
  const socketLastSeenStreamSeqRef = useRef(0);
  const socketClientInstanceIdRef = useRef<string | null>(null);
  const rowInputRefs = useRef(
    new Map<string, HTMLInputElement | HTMLTextAreaElement>(),
  );
  const gridShellRef = useRef<HTMLDivElement | null>(null);
  const viewportContinuityTokenRef = useRef(1);
  const [viewportContinuityRequest, setViewportContinuityRequest] =
    useState<ViewportContinuityRequest | null>(null);

  const setConflictQueueState = useCallback(
    (
      updater: (
        current: Record<string, LocalConflictState>,
      ) => Record<string, LocalConflictState>,
    ) => {
      setConflictQueue((current) => {
        const next = updater(current);
        conflictQueueRef.current = next;
        return next;
      });
    },
    [],
  );

  const computeSaveState = useCallback(
    (
      pending: PendingQueueRuntime,
      conflicts: Record<string, LocalConflictState> = conflictQueueRef.current,
    ): SaveState => {
      if (
        Object.keys(conflicts).length > 0 ||
        pending.overflowMessage !== null ||
        pending.haltedMessage !== null
      ) {
        return "Conflict";
      }
      if (
        pending.units.length > 0 ||
        pending.resetRefreshInFlight ||
        pendingOpsRef.current > 0
      ) {
        return "Syncing";
      }
      return "Saved";
    },
    [],
  );

  useEffect(() => {
    currentPresenceRef.current = currentPresence;
    if (
      activeSocketRef.current === null ||
      !socketEstablishedRef.current ||
      !socketIsOpen(activeSocketRef.current)
    ) {
      return;
    }
    if (presenceUpdateTimerRef.current !== null) {
      window.clearTimeout(presenceUpdateTimerRef.current);
    }
    presenceUpdateTimerRef.current = window.setTimeout(() => {
      presenceUpdateTimerRef.current = null;
      const target = activeSocketRef.current;
      if (
        target === null ||
        !socketEstablishedRef.current ||
        !socketIsOpen(target)
      ) {
        return;
      }
      target.send(
        JSON.stringify({
          type: "presence_update",
          payload: {
            presence: workbookPresence(currentPresenceRef.current),
          },
        }),
      );
    }, 150);
  }, [currentPresence]);

  const publishPendingQueueState = useCallback(() => {
    const pending = pendingQueueRef.current;
    setPendingQueueSnapshot({
      queuedCount: pending.units.filter((unit) => unit.status === "queued")
        .length,
      inFlightCount: pending.units.filter((unit) => unit.status === "in_flight")
        .length,
      haltedMessage: pending.haltedMessage,
      authPaused: pending.authPaused,
      overflowMessage: pending.overflowMessage,
      resetRefreshInFlight: pending.resetRefreshInFlight,
    });
    setSaveState(computeSaveState(pending));
  }, [computeSaveState]);

  const queryPath = useMemo(
    () =>
      apiPath(
        apiBase,
        `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`,
      ),
    [apiBase, incidentId],
  );
  const queryBody = useMemo(
    () => JSON.stringify(buildQueryRequest(timelineContract, queryState)),
    [queryState],
  );
  const changeSocketURL = useMemo(
    () => websocketPath(apiBase, `/ws/v1/incidents/${incidentId}`),
    [apiBase, incidentId],
  );
  const activeSheetRef = useMemo<WorkbookSheetRef>(
    () => ({ kind: "view_schema", id: timelineViewSchemaId }),
    [],
  );
  const nextDraftIndex = useCallback(() => {
    const value = draftCounterRef.current;
    draftCounterRef.current += 1;
    return value;
  }, []);

  const selectedRow = useMemo(
    () =>
      rows.find(
        (row) => row.recordId !== null && row.recordId === selectedRowId,
      ) ?? null,
    [rows, selectedRowId],
  );
  const draftRow = useMemo(
    () => rows.find((row) => row.recordId === null) ?? null,
    [rows],
  );
  const dismissedForSelectedRow = selectedRow?.recordId
    ? (dismissedMentionsByRow[selectedRow.recordId] ?? [])
    : [];
  const inspectorMentions = useMemo(
    () =>
      buildInspectorMentions(selectedRow ?? undefined, dismissedForSelectedRow),
    [dismissedForSelectedRow, selectedRow],
  );
  const selectedMention =
    inspectorMentions.find((item) => item.itemRef === selectedMentionRef) ??
    inspectorMentions[0] ??
    null;
  const canManageMentions =
    currentIncidentRole === "editor" ||
    currentIncidentRole === "reviewer" ||
    currentIncidentRole === "admin";
  const activeSheetPresenceRecords = useMemo(
    () =>
      [...presenceRecords]
        .filter((presence) => presenceMatchesSheet(presence, activeSheetRef))
        .filter(
          (presence) =>
            presence.connection_id !== socketConnectionIDRef.current,
        )
        .sort((left, right) => {
          const byName = left.display_name.localeCompare(right.display_name);
          return byName === 0
            ? left.connection_id.localeCompare(right.connection_id)
            : byName;
        }),
    [activeSheetRef, presenceRecords],
  );

  const presenceForRow = useCallback(
    (recordId: string | null) =>
      recordId === null
        ? []
        : activeSheetPresenceRecords.filter(
            (presence) => presence.record_id === recordId,
          ),
    [activeSheetPresenceRecords],
  );

  const editingPresenceForCell = useCallback(
    (recordId: string | null, fieldKey: string) =>
      recordId === null
        ? []
        : activeSheetPresenceRecords.filter(
            (presence) =>
              presence.record_id === recordId &&
              presence.field_key === fieldKey &&
              presence.mode === "editing",
          ),
    [activeSheetPresenceRecords],
  );

  const applyQueryFilter = useCallback(() => {
    setQueryState((current) => applyFilterDraft(current, filterDraft));
    setFilterDraft((current) => ({
      ...current,
      booleanValue: "",
      value: "",
    }));
  }, [filterDraft]);

  const handleQueryGroupByChange = useCallback((groupBy: string | null) => {
    setQueryState((current) =>
      updateGroupBy(timelineContract, current, groupBy),
    );
  }, []);

  const handleQuerySortToggle = useCallback((fieldKey: string) => {
    setQueryState((current) =>
      toggleSortField(timelineContract, current, fieldKey),
    );
  }, []);

  const currentGridScrollSnapshot = useCallback(() => {
    const element = gridShellRef.current;
    if (!element) {
      return null;
    }
    const scrollElement = resolveGridScrollElement(element, "timeline");
    return {
      top: scrollElement.scrollTop,
      left: scrollElement.scrollLeft,
    };
  }, []);

  const currentGridViewportSnapshot = useCallback(
    (target: HTMLElement | null = null): ViewportSnapshot | null => {
      const gridShell = gridShellRef.current;
      const scroll = currentGridScrollSnapshot();
      if (gridShell === null || scroll === null) {
        return null;
      }
      return {
        scroll,
        anchor:
          target === null
            ? null
            : captureViewportAnchor(
                resolveGridScrollElement(
                  gridShell,
                  "timeline",
                ).getBoundingClientRect(),
                target.getBoundingClientRect(),
              ),
      };
    },
    [currentGridScrollSnapshot],
  );

  const trackPendingSocketTxn = useCallback((clientTxnId: string) => {
    const existingTimeout =
      pendingSocketTxnTimeoutsRef.current.get(clientTxnId);
    if (existingTimeout !== undefined) {
      window.clearTimeout(existingTimeout);
    }
    const timeoutId = window.setTimeout(() => {
      pendingSocketTxnTimeoutsRef.current.delete(clientTxnId);
    }, 30_000);
    pendingSocketTxnTimeoutsRef.current.set(clientTxnId, timeoutId);
  }, []);

  const resolvePendingSocketTxn = useCallback(
    (clientTxnId: string | null | undefined) => {
      if (!clientTxnId) {
        return false;
      }

      const timeoutId = pendingSocketTxnTimeoutsRef.current.get(clientTxnId);
      if (timeoutId === undefined) {
        return false;
      }

      window.clearTimeout(timeoutId);
      pendingSocketTxnTimeoutsRef.current.delete(clientTxnId);
      return true;
    },
    [],
  );

  const restoreGridScroll = useCallback(
    (preservedScroll: ScrollPosition | null) => {
      const gridShell = gridShellRef.current;
      if (gridShell === null || preservedScroll === null) {
        return;
      }
      const scrollElement = resolveGridScrollElement(gridShell, "timeline");
      scrollElement.scrollTop = preservedScroll.top;
      scrollElement.scrollLeft = preservedScroll.left;
      window.requestAnimationFrame(() => {
        const currentGridShell = gridShellRef.current;
        if (currentGridShell === null) {
          return;
        }
        const currentScrollElement = resolveGridScrollElement(
          currentGridShell,
          "timeline",
        );
        currentScrollElement.scrollTop = preservedScroll.top;
        currentScrollElement.scrollLeft = preservedScroll.left;
      });
    },
    [],
  );

  const restoreGridViewportForElement = useCallback(
    (
      resolveElement: () =>
        | HTMLButtonElement
        | HTMLInputElement
        | HTMLTextAreaElement
        | null,
      preservedViewport: ViewportSnapshot | null,
    ) => {
      const element = resolveElement();
      if (element === null) {
        return false;
      }
      // Continuity restores the previous scroll position first, then applies
      // only the extra delta needed to keep the target fully visible.
      const currentViewport =
        preservedViewport ??
        ({
          scroll: currentGridScrollSnapshot(),
          anchor: null,
        } satisfies ViewportSnapshot);
      const preservedScroll = currentViewport.scroll;
      window.focus();
      element.focus({ preventScroll: true });
      restoreGridScroll(preservedScroll);
      const restoreViewportGeometryNow = () => {
        const currentGridShell = gridShellRef.current;
        const currentElement = resolveElement();
        if (
          currentGridShell === null ||
          preservedScroll === null ||
          currentElement === null ||
          !currentElement.isConnected
        ) {
          return false;
        }
        const scrollElement = resolveGridScrollElement(
          currentGridShell,
          "timeline",
        );
        const restoredScroll = computeRestoredViewportScroll({
          preservedScroll,
          currentScroll: {
            top: scrollElement.scrollTop,
            left: scrollElement.scrollLeft,
          },
          preservedAnchor: currentViewport.anchor,
          containerRect: scrollElement.getBoundingClientRect(),
          elementRect: currentElement.getBoundingClientRect(),
        });
        restoreGridScroll(restoredScroll);
        const updatedGridShell = gridShellRef.current;
        const updatedElement = resolveElement();
        if (
          updatedGridShell === null ||
          updatedElement === null ||
          !updatedElement.isConnected
        ) {
          return false;
        }
        return isRectFullyVisibleWithinContainer(
          resolveGridScrollElement(
            updatedGridShell,
            "timeline",
          ).getBoundingClientRect(),
          updatedElement.getBoundingClientRect(),
        );
      };
      const restoredNow = restoreViewportGeometryNow();
      const restoreViewportGeometry = (attempt: number) => {
        window.requestAnimationFrame(() => {
          if (restoreViewportGeometryNow()) {
            return;
          }
          if (attempt < 6) {
            restoreViewportGeometry(attempt + 1);
          }
        });
      };
      restoreViewportGeometry(0);
      return document.activeElement === element && restoredNow;
    },
    [currentGridScrollSnapshot, restoreGridScroll],
  );

  const resolveInputElement = useCallback((focusKey: string) => {
    const selectorTestId = focusTestIdForKey(focusKey);
    const selector =
      selectorTestId === null
        ? null
        : document.querySelector<HTMLInputElement | HTMLTextAreaElement>(
            `[data-testid="${selectorTestId}"]`,
          );
    return selector ?? rowInputRefs.current.get(focusKey) ?? null;
  }, []);

  const resolveViewportContinuityElement = useCallback(
    (target: ViewportContinuityTarget) => {
      switch (target.kind) {
        case "row-inspect":
          return document.querySelector<HTMLButtonElement>(
            `[data-testid="${rowInspectButtonTestId(target.recordId)}"]`,
          );
        case "input":
          return resolveInputElement(target.focusKey);
        case "scroll-only":
          return null;
      }
    },
    [resolveInputElement],
  );

  const beginViewportContinuity = useCallback(
    (
      target: ViewportContinuityTarget,
      options: { followup?: ViewportContinuityFollowup } = {},
    ) => {
      const token = viewportContinuityTokenRef.current;
      viewportContinuityTokenRef.current += 1;
      const followup = options.followup ?? "none";
      setViewportContinuityRequest({
        token,
        attemptVersion: 0,
        target,
        preservedViewport: currentGridViewportSnapshot(
          resolveViewportContinuityElement(target),
        ),
        followup,
        followupSettled: followup === "none",
        baselineHostEntities: hostEntities,
        baselineIdentityEntities: identityEntities,
      });
      return token;
    },
    [
      currentGridViewportSnapshot,
      hostEntities,
      identityEntities,
      resolveViewportContinuityElement,
    ],
  );

  const settleViewportContinuityFollowup = useCallback((token: number) => {
    setViewportContinuityRequest((current) => {
      if (!current || current.token !== token) {
        return current;
      }
      return {
        ...current,
        followupSettled: true,
        attemptVersion: current.attemptVersion + 1,
      };
    });
  }, []);

  const clearViewportContinuity = useCallback((token: number) => {
    setViewportContinuityRequest((current) =>
      current?.token === token ? null : current,
    );
  }, []);

  const advanceViewportContinuity = useCallback(
    (
      token: number | undefined,
      options: {
        target?: ViewportContinuityTarget | null;
      } = {},
    ) => {
      if (token === undefined) {
        return;
      }
      setViewportContinuityRequest((current) => {
        if (current === null || current.token !== token) {
          return current;
        }
        return {
          ...current,
          attemptVersion: current.attemptVersion + 1,
          target: options.target ?? current.target,
        };
      });
    },
    [],
  );

  const tryRestoreViewportContinuity = useCallback(
    (continuity: ViewportContinuityRequest) => {
      if (continuity.target.kind === "scroll-only") {
        restoreGridScroll(continuity.preservedViewport?.scroll ?? null);
        return true;
      }
      return restoreGridViewportForElement(
        () => resolveViewportContinuityElement(continuity.target),
        continuity.preservedViewport,
      );
    },
    [
      resolveViewportContinuityElement,
      restoreGridScroll,
      restoreGridViewportForElement,
    ],
  );

  const shouldHoldViewportContinuity = useCallback(
    (continuity: ViewportContinuityRequest) => {
      if (continuity.followup !== "entity-refresh") {
        return false;
      }
      if (!continuity.followupSettled) {
        return true;
      }
      return (
        continuity.baselineHostEntities === hostEntities &&
        continuity.baselineIdentityEntities === identityEntities
      );
    },
    [hostEntities, identityEntities],
  );

  const applyRowMutation = useCallback(
    (
      rowKey: string,
      envelope: TimelineMutationEnvelope,
      options: {
        continueOnFreshDraft?: boolean;
        detectAutoResolution?: boolean;
        promoteToCommittedRowInspect?: boolean;
        viewportContinuityToken?: number;
      } = {},
    ) => {
      const previousRow = rowsRef.current.find(
        (candidate) => candidate.key === rowKey,
      );
      const committed = rowFromApi(envelope.data.row);
      const nextRows = rowsRef.current.map((row) =>
        row.key === rowKey ? committed : row,
      );
      const hydrated = ensureDraftRowWithFreshIndex(nextRows, nextDraftIndex);
      const hydratedRows = hydrated.rows;

      rowsRef.current = hydratedRows;
      setRows(hydratedRows);

      if (committed.recordId !== null) {
        setDismissedMentionsByRow((current) =>
          pruneDismissedMentions(current, committed),
        );
      }
      if (options.detectAutoResolution !== false) {
        const notices = buildAutoResolutionNotices(previousRow, committed);
        if (notices.length > 0) {
          setAutoResolutionNotices((current) => {
            const knownRefs = new Set(current.map((notice) => notice.itemRef));
            return [
              ...current,
              ...notices.filter((notice) => !knownRefs.has(notice.itemRef)),
            ];
          });
        }
      }
      if (
        selectedRowId !== null &&
        previousRow?.recordId !== null &&
        previousRow !== undefined &&
        previousRow.recordId === selectedRowId
      ) {
        setSelectedRowId(committed.recordId);
      }
      const nextViewportTarget =
        options.continueOnFreshDraft && hydrated.draftSummaryKey
          ? ({
              kind: "input",
              focusKey: hydrated.draftSummaryKey,
            } satisfies ViewportContinuityTarget)
          : options.promoteToCommittedRowInspect && committed.recordId !== null
            ? ({
                kind: "row-inspect",
                recordId: committed.recordId,
              } satisfies ViewportContinuityTarget)
            : null;
      advanceViewportContinuity(options.viewportContinuityToken, {
        target: nextViewportTarget,
      });
    },
    [advanceViewportContinuity, nextDraftIndex, selectedRowId],
  );

  const currentCommittedTimelineRow = useCallback((recordId: string) => {
    const row = rowsRef.current.find(
      (candidate) => candidate.recordId === recordId,
    );
    return row && row.rowVersion !== null ? row : null;
  }, []);

  const waitForPendingReplayForRecord = useCallback(
    async (recordId: string) => {
      for (;;) {
        const pending = pendingQueueRef.current;
        const hasPendingRecordWork = pending.units.some(
          (unit) => unit.recordId === recordId,
        );
        if (!hasPendingRecordWork) {
          return true;
        }
        if (pending.authPaused || pending.haltedMessage !== null) {
          return false;
        }
        await new Promise((resolve) => window.setTimeout(resolve, 16));
      }
    },
    [],
  );

  const loadRows = useCallback(
    async (options: LoadRowsOptions) => {
      const requestSequence = loadSequenceRef.current + 1;
      loadSequenceRef.current = requestSequence;

      if (options.showLoading && !hasLoadedRowsRef.current) {
        setIsInitialLoading(true);
      }
      if (hasLoadedRowsRef.current) {
        setRefreshError(null);
      } else {
        setLoadError(null);
      }

      const result = await fetchJSON<WorkbookQueryEnvelope>(queryPath, {
        method: "POST",
        body: queryBody,
      });

      if (requestSequence !== loadSequenceRef.current) {
        return;
      }

      if (!result.ok) {
        if (options.viewportContinuityToken !== undefined) {
          clearViewportContinuity(options.viewportContinuityToken);
        }
        const message = "Timeline projection load failed.";
        if (hasLoadedRowsRef.current) {
          setRefreshError(message);
        } else {
          setLoadError(message);
          setIsInitialLoading(false);
        }
        return;
      }

      const envelope = readEnvelope<WorkbookQueryEnvelope>(result.payload);
      const { committedRows, rows: hydratedRows } =
        reconcileCommittedRowsWithLocalDrafts({
          currentRows: rowsRef.current,
          incomingRows: envelope.data.rows.map(rowFromApi),
          draftValueForFocusKey: (focusKey) =>
            scalarDraftValuesRef.current.get(focusKey),
          nextDraftIndex,
        });
      startTransition(() => {
        rowsRef.current = hydratedRows;
        setRows(hydratedRows);
      });
      advanceViewportContinuity(options.viewportContinuityToken);
      setDismissedMentionsByRow((current) => {
        const next = { ...current };
        for (const row of committedRows) {
          if (row.recordId === null) {
            continue;
          }
          Object.assign(next, pruneDismissedMentions(next, row));
        }
        return next;
      });
      setSaveState(computeSaveState(pendingQueueRef.current));
      hasLoadedRowsRef.current = true;
      setLoadError(null);
      setRefreshError(null);
      setIsInitialLoading(false);
    },
    [
      advanceViewportContinuity,
      clearViewportContinuity,
      computeSaveState,
      nextDraftIndex,
      queryBody,
      queryPath,
    ],
  );

  loadRowsRef.current = loadRows;

  const applyRecordChangedPatch = useCallback(
    (payload: RecordChangedPayload) => {
      const affectedView = payload.affected_views.find(
        (view) => view.view_schema_id === timelineViewSchemaId,
      );
      if (
        affectedView?.change_kind !== "patch" ||
        affectedView.patch_cells === undefined ||
        affectedView.patch_cells.record_id !== payload.record_id
      ) {
        return false;
      }

      const patch = affectedView.patch_cells;
      let patched = false;
      const nextRows = rowsRef.current.map((row) => {
        if (row.recordId !== patch.record_id || row.rawRow === null) {
          return row;
        }
        patched = true;
        const nextRawRow = applyViewRowPatch(row.rawRow, patch);
        const committed = rowFromApi(nextRawRow);
        return {
          ...committed,
          collectionDrafts: row.collectionDrafts,
          pendingSignature: row.pendingSignature,
        };
      });
      if (!patched) {
        return false;
      }
      rowsRef.current = nextRows;
      setRows(nextRows);
      return true;
    },
    [],
  );

  useEffect(() => {
    void loadRows({ showLoading: true });
  }, [loadRows]);

  useLayoutEffect(() => {
    if (
      viewportContinuityRequest === null ||
      viewportContinuityRequest.attemptVersion < 1
    ) {
      return;
    }
    let cancelled = false;
    const restoreTarget = (attempt: number) => {
      if (cancelled) {
        return;
      }
      if (!tryRestoreViewportContinuity(viewportContinuityRequest)) {
        if (attempt < 60) {
          window.setTimeout(() => {
            restoreTarget(attempt + 1);
          }, 50);
        }
        return;
      }
      if (shouldHoldViewportContinuity(viewportContinuityRequest)) {
        return;
      }
      clearViewportContinuity(viewportContinuityRequest.token);
    };
    restoreTarget(0);
    return () => {
      cancelled = true;
    };
  }, [
    clearViewportContinuity,
    shouldHoldViewportContinuity,
    tryRestoreViewportContinuity,
    viewportContinuityRequest,
  ]);

  useEffect(() => {
    return () => {
      for (const timeoutId of pendingSocketTxnTimeoutsRef.current.values()) {
        window.clearTimeout(timeoutId);
      }
      pendingSocketTxnTimeoutsRef.current.clear();
      if (pendingReplayTimerRef.current !== null) {
        window.clearTimeout(pendingReplayTimerRef.current);
        pendingReplayTimerRef.current = null;
      }
      if (pendingReplayAuthRetryRef.current !== null) {
        window.clearTimeout(pendingReplayAuthRetryRef.current);
        pendingReplayAuthRetryRef.current = null;
      }
    };
  }, []);

  useEffect(() => {
    const hasUnsavedRuntimeWork =
      pendingQueueSnapshot.queuedCount > 0 ||
      pendingQueueSnapshot.inFlightCount > 0 ||
      Object.keys(conflictQueue).length > 0;
    if (!hasUnsavedRuntimeWork) {
      return;
    }
    const warnBeforeUnload = (event: BeforeUnloadEvent) => {
      event.preventDefault();
      event.returnValue = "";
    };
    window.addEventListener("beforeunload", warnBeforeUnload);
    return () => {
      window.removeEventListener("beforeunload", warnBeforeUnload);
    };
  }, [
    conflictQueue,
    pendingQueueSnapshot.inFlightCount,
    pendingQueueSnapshot.queuedCount,
  ]);

  useEffect(() => {
    if (selectedRowId === null) {
      return;
    }
    if (!rows.some((row) => row.recordId === selectedRowId)) {
      setSelectedRowId(null);
    }
  }, [rows, selectedRowId]);

  useEffect(() => {
    if (inspectorMentions.length < 1) {
      if (selectedMentionRef !== null) {
        setSelectedMentionRef(null);
      }
      setSelectedResolveTargetId("");
      return;
    }
    if (
      selectedMentionRef !== null &&
      inspectorMentions.some((item) => item.itemRef === selectedMentionRef)
    ) {
      return;
    }
    const [firstMention] = inspectorMentions;
    if (firstMention) {
      setSelectedMentionRef(firstMention.itemRef);
    }
    setSelectedResolveTargetId("");
  }, [inspectorMentions, selectedMentionRef]);

  const nextClientTxnId = useCallback(() => {
    const value = clientTxnRef.current;
    clientTxnRef.current += 1;
    return `timeline-client-${value}`;
  }, []);

  const beginSave = useCallback(() => {
    pendingOpsRef.current += 1;
    setSaveState(computeSaveState(pendingQueueRef.current));
  }, [computeSaveState]);

  const finishSave = useCallback(
    (nextState: SaveState) => {
      pendingOpsRef.current = Math.max(0, pendingOpsRef.current - 1);
      if (nextState === "Conflict") {
        setSaveState("Conflict");
        return;
      }
      setSaveState(computeSaveState(pendingQueueRef.current));
    },
    [computeSaveState],
  );

  const restoreConflictFocus = useCallback((focusKey: string) => {
    window.setTimeout(() => {
      const testId = focusTestIdForKey(focusKey);
      const element =
        testId === null
          ? null
          : (document.querySelector(
              `[data-testid="${testId}"]`,
            ) as HTMLElement | null);
      element?.focus();
    }, 0);
  }, []);

  const updateSaveStateForConflicts = useCallback(
    (nextQueue: Record<string, LocalConflictState>) => {
      setSaveState(computeSaveState(pendingQueueRef.current, nextQueue));
    },
    [computeSaveState],
  );

  const registerSameFieldConflict = useCallback(
    (
      conflict: SameFieldConflictPayload,
      focusKey: string,
      surface: TimelineScalarEditorSurface,
    ) => {
      const queueKey = `${conflict.record_id}:${conflict.field_key}`;
      const binding = timelineScalarBindingForField(conflict.field_key);
      if (binding !== null && typeof conflict.client_value === "string") {
        scalarDraftValuesRef.current.set(
          inputFocusKey(conflict.record_id, binding.key, surface),
          conflict.client_value,
        );
      }
      if (binding !== null) {
        setRows((current) => {
          const nextRows = current.map((row) => {
            if (row.recordId !== conflict.record_id) {
              return row;
            }
            const serverText =
              typeof conflict.server_value === "string"
                ? conflict.server_value
                : "";
            return {
              ...row,
              rowVersion: conflict.current_row_version,
              values: { ...row.values, [binding.key]: serverText },
              committedValues: {
                ...row.committedValues,
                [binding.key]: serverText,
              },
              pendingSignature: null,
            };
          });
          rowsRef.current = nextRows;
          return nextRows;
        });
      }
      setConflictQueueState((current) => {
        const next = {
          ...current,
          [queueKey]: {
            key: queueKey,
            conflict,
            focusKey,
            localValue: conflict.client_value,
            mergedDraft:
              typeof conflict.suggested_merged_value === "string"
                ? conflict.suggested_merged_value
                : typeof conflict.server_value === "string"
                  ? conflict.server_value
                  : "",
          },
        };
        updateSaveStateForConflicts(next);
        return next;
      });
      setActiveConflictKey(queueKey);
    },
    [setConflictQueueState, updateSaveStateForConflicts],
  );

  const handleMutationConflict = useCallback(
    (
      payload: unknown,
      rowKey: string,
      focusField: FocusFieldKey,
      surface: TimelineScalarEditorSurface,
    ) => {
      const conflict = parseSameFieldConflict(payload);
      if (conflict === null) {
        return false;
      }
      registerSameFieldConflict(
        conflict,
        inputFocusKey(rowKey, focusField, surface),
        surface,
      );
      return true;
    },
    [registerSameFieldConflict],
  );

  const setScalarEditorDraftValue = useCallback(
    (
      rowKey: string,
      field: keyof RowValues,
      surface: TimelineScalarEditorSurface,
      value: string,
    ) => {
      scalarDraftValuesRef.current.set(
        inputFocusKey(rowKey, field, surface),
        value,
      );
    },
    [],
  );

  const rowWithScalarEditorDrafts = useCallback(
    (
      row: WorkbookRow,
      preferred?: {
        readonly field: keyof RowValues;
        readonly value: string | undefined;
      },
    ): WorkbookRow => {
      return rowWithMaterializedScalarDrafts(
        row,
        (focusKey) => scalarDraftValuesRef.current.get(focusKey),
        preferred,
      );
    },
    [],
  );

  const clearSubmittedScalarEditorDraftValuesForRow = useCallback(
    (rowKey: string, submittedValues: RowValues) => {
      for (const binding of Object.values(timelineScalarBindingIndex)) {
        for (const surface of timelineScalarEditorSurfaces) {
          const focusKey = inputFocusKey(rowKey, binding.key, surface);
          if (
            scalarDraftValuesRef.current.get(focusKey) ===
            submittedValues[binding.key]
          ) {
            scalarDraftValuesRef.current.delete(focusKey);
          }
        }
      }
    },
    [],
  );

  const registerInput = useCallback(
    (
      rowKey: string,
      field: FocusFieldKey,
      surface: TimelineScalarEditorSurface,
      element: HTMLInputElement | HTMLTextAreaElement | null,
    ) => {
      const key = inputFocusKey(rowKey, field, surface);
      if (element === null) {
        rowInputRefs.current.delete(key);
        return;
      }
      rowInputRefs.current.set(key, element);
    },
    [],
  );

  const clearPendingSignatureForUnit = useCallback(
    (unit: PendingReplayUnit) => {
      if (
        pendingSignaturesRef.current.get(unit.rowKey) === unit.mutationSignature
      ) {
        pendingSignaturesRef.current.delete(unit.rowKey);
      }
      const nextRows = rowsRef.current.map((row) =>
        row.key === unit.rowKey &&
        row.pendingSignature === unit.mutationSignature
          ? { ...row, pendingSignature: null }
          : row,
      );
      rowsRef.current = nextRows;
      setRows(nextRows);
    },
    [],
  );

  const mergeCollectionActionPayload = useCallback(
    (existing: unknown, next: unknown) => {
      if (
        existing &&
        typeof existing === "object" &&
        !Array.isArray(existing) &&
        next &&
        typeof next === "object" &&
        !Array.isArray(next) &&
        "kind" in existing &&
        "kind" in next &&
        existing.kind === "collection_actions_v1" &&
        next.kind === "collection_actions_v1" &&
        "actions" in existing &&
        "actions" in next &&
        Array.isArray(existing.actions) &&
        Array.isArray(next.actions)
      ) {
        return {
          kind: "collection_actions_v1",
          actions: [...existing.actions, ...next.actions],
        };
      }
      return next;
    },
    [],
  );

  const mergePendingPayload = useCallback(
    (
      existing: Record<string, unknown>,
      next: Record<string, unknown>,
      kind: PendingReplayKind,
    ) => {
      if (kind === "create") {
        const merged = { ...existing };
        for (const [key, value] of Object.entries(next)) {
          if (key === "client_txn_id") {
            continue;
          }
          merged[key] = mergeCollectionActionPayload(merged[key], value);
        }
        return merged;
      }

      const existingChanges = Array.isArray(existing.changes)
        ? (existing.changes as Array<Record<string, unknown>>)
        : [];
      const nextChanges = Array.isArray(next.changes)
        ? (next.changes as Array<Record<string, unknown>>)
        : [];
      const mergedByField = new Map<string, Record<string, unknown>>();
      for (const change of existingChanges) {
        const fieldKey = change.field_key;
        if (typeof fieldKey === "string") {
          mergedByField.set(fieldKey, { ...change });
        }
      }
      for (const change of nextChanges) {
        const fieldKey = change.field_key;
        if (typeof fieldKey !== "string") {
          continue;
        }
        const existingChange = mergedByField.get(fieldKey);
        if (
          existingChange &&
          "action_payload" in existingChange &&
          "action_payload" in change
        ) {
          mergedByField.set(fieldKey, {
            ...existingChange,
            action_payload: mergeCollectionActionPayload(
              existingChange.action_payload,
              change.action_payload,
            ),
          });
          continue;
        }
        mergedByField.set(fieldKey, { ...change });
      }
      return {
        ...existing,
        changes: Array.from(mergedByField.values()).sort((left, right) =>
          String(left.field_key).localeCompare(String(right.field_key)),
        ),
      };
    },
    [mergeCollectionActionPayload],
  );

  const schedulePendingReplay = useCallback(() => {
    const pending = pendingQueueRef.current;
    if (pending.replayScheduled) {
      return;
    }
    pending.replayScheduled = true;
    pendingReplayTimerRef.current = window.setTimeout(() => {
      pendingReplayTimerRef.current = null;
      void replayPendingQueueRef.current();
    }, 0);
  }, []);

  const schedulePendingReplayRetry = useCallback(() => {
    const pending = pendingQueueRef.current;
    if (pending.replayScheduled) {
      return;
    }
    pending.replayScheduled = true;
    pendingReplayTimerRef.current = window.setTimeout(() => {
      pendingReplayTimerRef.current = null;
      void replayPendingQueueRef.current();
    }, 1000);
  }, []);

  const scheduleAuthRecoveryProbe = useCallback(() => {
    if (pendingReplayAuthRetryRef.current !== null) {
      return;
    }
    pendingReplayAuthRetryRef.current = window.setTimeout(async () => {
      pendingReplayAuthRetryRef.current = null;
      if (!pendingQueueRef.current.authPaused) {
        return;
      }
      try {
        const result = await fetchJSON<SessionEnvelope>(
          apiPath(apiBase, "/api/v1/auth/session"),
        );
        if (!result.ok) {
          scheduleAuthRecoveryProbe();
          return;
        }
        pendingQueueRef.current.authPaused = false;
        pendingQueueRef.current.haltedMessage = null;
        publishPendingQueueState();
        schedulePendingReplay();
      } catch {
        scheduleAuthRecoveryProbe();
      }
    }, 1000);
  }, [apiBase, publishPendingQueueState, schedulePendingReplay]);

  const replayPendingQueueRef = useRef<() => Promise<void>>(async () => {
    return undefined;
  });

  const enqueuePendingReplayUnit = useCallback(
    (unit: PendingReplayUnit) => {
      const pending = pendingQueueRef.current;
      pending.overflowMessage = null;
      pending.haltedMessage = null;
      const duplicate = pending.units.some(
        (candidate) =>
          candidate.rowKey === unit.rowKey &&
          candidate.mutationSignature === unit.mutationSignature,
      );
      if (duplicate) {
        clearViewportContinuity(unit.viewportContinuityToken);
        return;
      }

      const lastUnit = pending.units[pending.units.length - 1];
      const canCoalesce =
        lastUnit !== undefined &&
        lastUnit.status === "queued" &&
        lastUnit.kind === unit.kind &&
        lastUnit.coalesceKey === unit.coalesceKey &&
        (unit.kind === "create" ||
          (unit.recordId !== null && lastUnit.recordId === unit.recordId));
      if (canCoalesce) {
        lastUnit.payload = mergePendingPayload(
          lastUnit.payload,
          unit.payload,
          unit.kind,
        );
        lastUnit.rowSnapshot = unit.rowSnapshot;
        lastUnit.focusField = unit.focusField;
        lastUnit.focusKey = unit.focusKey;
        lastUnit.surface = unit.surface;
        lastUnit.mutationSignature = buildStableMutationSignature(
          lastUnit.payload,
        );
        pendingSignaturesRef.current.set(
          lastUnit.rowKey,
          lastUnit.mutationSignature,
        );
        clearViewportContinuity(unit.viewportContinuityToken);
        publishPendingQueueState();
        schedulePendingReplay();
        return;
      }

      if (pending.units.length >= pendingReplayCapacity) {
        pending.overflowMessage =
          "Local pending queue is full. The current edit remains unsaved.";
        clearPendingSignatureForUnit(unit);
        clearViewportContinuity(unit.viewportContinuityToken);
        publishPendingQueueState();
        return;
      }

      pending.units.push(unit);
      pendingSignaturesRef.current.set(unit.rowKey, unit.mutationSignature);
      publishPendingQueueState();
      schedulePendingReplay();
    },
    [
      clearPendingSignatureForUnit,
      clearViewportContinuity,
      mergePendingPayload,
      publishPendingQueueState,
      schedulePendingReplay,
    ],
  );

  const shouldRetryPendingResult = useCallback((status: number) => {
    return status === 0 || status === 408 || status === 425 || status >= 500;
  }, []);

  replayPendingQueueRef.current = async () => {
    const pending = pendingQueueRef.current;
    pending.replayScheduled = false;
    if (
      pending.authPaused ||
      pending.haltedMessage !== null ||
      pending.resetRefreshInFlight ||
      Object.keys(conflictQueueRef.current).length > 0
    ) {
      publishPendingQueueState();
      return;
    }
    if (pending.units.some((unit) => unit.status === "in_flight")) {
      publishPendingQueueState();
      return;
    }
    const unit = pending.units.find(
      (candidate) => candidate.status === "queued",
    );
    if (!unit) {
      publishPendingQueueState();
      return;
    }

    const currentRow =
      unit.recordId === null
        ? rowsRef.current.find((row) => row.key === unit.rowKey)
        : rowsRef.current.find((row) => row.recordId === unit.recordId);
    if (
      unit.kind === "patch" &&
      currentRow?.rowVersion !== null &&
      currentRow?.rowVersion !== undefined
    ) {
      unit.payload = {
        ...unit.payload,
        base_row_version: currentRow.rowVersion,
      };
    }

    unit.status = "in_flight";
    publishPendingQueueState();
    trackPendingSocketTxn(unit.clientTxnId);

    let result: Awaited<
      ReturnType<typeof fetchJSON<TimelineMutationEnvelope>>
    > | null = null;
    try {
      result = await fetchJSON<TimelineMutationEnvelope>(unit.path, {
        method: unit.method,
        body: JSON.stringify(unit.payload),
      });
    } catch {
      resolvePendingSocketTxn(unit.clientTxnId);
      unit.status = "queued";
      publishPendingQueueState();
      schedulePendingReplayRetry();
      return;
    }

    if (!result.ok) {
      resolvePendingSocketTxn(unit.clientTxnId);
      if (result.status === 401 || result.status === 403) {
        unit.status = "queued";
        pending.authPaused = true;
        pending.haltedMessage = null;
        setRefreshError(
          "Authentication required before queued edits can replay.",
        );
        publishPendingQueueState();
        scheduleAuthRecoveryProbe();
        return;
      }

      if (
        handleMutationConflict(
          result.payload,
          unit.rowKey,
          unit.focusField,
          unit.surface,
        )
      ) {
        pending.units = pending.units.filter((candidate) => candidate !== unit);
        clearPendingSignatureForUnit(unit);
        publishPendingQueueState();
        return;
      }

      if (shouldRetryPendingResult(result.status)) {
        unit.status = "queued";
        publishPendingQueueState();
        schedulePendingReplayRetry();
        return;
      }

      unit.status = "queued";
      pending.haltedMessage = parseErrorMessage(result.payload);
      setRefreshError(pending.haltedMessage);
      publishPendingQueueState();
      return;
    }

    pending.units = pending.units.filter((candidate) => candidate !== unit);
    clearPendingSignatureForUnit(unit);
    const envelope = readEnvelope<TimelineMutationEnvelope>(result.payload);
    clearSubmittedScalarEditorDraftValuesForRow(
      unit.rowKey,
      unit.rowSnapshot.values,
    );
    applyRowMutation(unit.rowKey, envelope, {
      continueOnFreshDraft:
        unit.continueOnFreshDraft && unit.rowSnapshot.recordId === null,
      detectAutoResolution: unit.detectAutoResolution,
      promoteToCommittedRowInspect: unit.promoteToCommittedRowInspect,
      viewportContinuityToken: unit.viewportContinuityToken,
    });
    pending.authPaused = false;
    pending.haltedMessage = null;
    publishPendingQueueState();
    schedulePendingReplay();
  };

  useEffect(() => {
    if (incidentId.trim() === "") {
      return;
    }

    let closed = false;
    let socket: WebSocket | null = null;
    let reconnectTimer: number | null = null;
    const clientInstanceId =
      socketClientInstanceIdRef.current ?? tabClientInstanceId();
    socketClientInstanceIdRef.current = clientInstanceId;

    const scheduleReconnect = () => {
      if (closed || reconnectTimer !== null) {
        return;
      }
      reconnectTimer = window.setTimeout(() => {
        reconnectTimer = null;
        connect();
      }, 1000);
    };

    const sendSessionEstablishment = (target: WebSocket) => {
      const resumeToken = socketResumeTokenRef.current;
      if (resumeToken) {
        target.send(
          JSON.stringify({
            type: "resume",
            payload: {
              client_instance_id: clientInstanceId,
              resume_token: resumeToken,
              last_seen_stream_seq: socketLastSeenStreamSeqRef.current,
              presence: workbookPresence(currentPresenceRef.current),
            },
          }),
        );
        return;
      }
      target.send(
        JSON.stringify({
          type: "hello",
          payload: {
            client_instance_id: clientInstanceId,
            presence: workbookPresence(currentPresenceRef.current),
          },
        }),
      );
    };

    const applyPresenceSnapshot = (payload: Record<string, unknown>) => {
      const presences = Array.isArray(payload.presences)
        ? payload.presences
        : [];
      setPresenceRecords(
        presences.filter(isPresenceRecord).map((presence) => ({
          ...presence,
          sheet_ref: { ...presence.sheet_ref },
        })),
      );
    };

    const applyPresenceDelta = (payload: Record<string, unknown>) => {
      const deltaKind = payload.delta_kind;
      const presence = payload.presence;
      if (!presence || typeof presence !== "object") {
        return;
      }
      const candidate = presence as Record<string, unknown>;
      const connectionID = candidate.connection_id;
      if (typeof connectionID !== "string") {
        return;
      }
      setPresenceRecords((current) => {
        if (deltaKind === "remove") {
          return current.filter(
            (record) => record.connection_id !== connectionID,
          );
        }
        if (deltaKind !== "upsert" || !isPresenceRecord(candidate)) {
          return current;
        }
        const nextRecord = {
          ...candidate,
          sheet_ref: { ...candidate.sheet_ref },
        };
        const withoutExisting = current.filter(
          (record) => record.connection_id !== nextRecord.connection_id,
        );
        return [...withoutExisting, nextRecord];
      });
    };

    const handleMessage = (target: WebSocket, raw: unknown) => {
      if (!raw || typeof raw !== "object") {
        return;
      }
      const message = raw as {
        type?: string;
        stream_seq?: number;
        payload?: Record<string, unknown>;
      };
      if (message.type === "ping") {
        target.send(JSON.stringify({ type: "pong", payload: {} }));
        return;
      }
      if (message.type === "hello_ack" || message.type === "resume_ack") {
        socketEstablishedRef.current = true;
        const connectionID = message.payload?.connection_id;
        if (typeof connectionID === "string") {
          socketConnectionIDRef.current = connectionID;
        }
        const resumeToken = message.payload?.resume_token;
        if (typeof resumeToken === "string") {
          socketResumeTokenRef.current = resumeToken;
        }
        if (
          message.type === "resume_ack" &&
          message.payload?.status === "reset_required"
        ) {
          pendingQueueRef.current.resetRefreshInFlight = true;
          publishPendingQueueState();
          void loadRowsRef.current({ showLoading: false }).finally(() => {
            pendingQueueRef.current.resetRefreshInFlight = false;
            publishPendingQueueState();
            schedulePendingReplay();
          });
          return;
        }
        pendingQueueRef.current.authPaused = false;
        publishPendingQueueState();
        schedulePendingReplay();
        return;
      }
      if (message.type === "presence_snapshot") {
        applyPresenceSnapshot(message.payload ?? {});
        return;
      }
      if (message.type === "presence_delta") {
        applyPresenceDelta(message.payload ?? {});
        return;
      }
      if (message.type === "session_revoked") {
        socketResumeTokenRef.current = null;
        socketEstablishedRef.current = false;
        pendingQueueRef.current.authPaused = true;
        setRefreshError(
          "Authentication required before queued edits can replay.",
        );
        publishPendingQueueState();
        scheduleAuthRecoveryProbe();
        target.close();
        return;
      }
      if (
        shouldIgnoreSelfOriginatedRecordChange(raw, resolvePendingSocketTxn)
      ) {
        return;
      }
      if (!isRecordChangedMessage(raw)) {
        return;
      }
      if (typeof message.stream_seq === "number") {
        if (appliedStreamSeqRef.current.has(message.stream_seq)) {
          return;
        }
        const previousSeq = socketLastSeenStreamSeqRef.current;
        if (previousSeq > 0 && message.stream_seq > previousSeq + 1) {
          pendingQueueRef.current.resetRefreshInFlight = true;
          publishPendingQueueState();
          void loadRowsRef.current({ showLoading: false }).finally(() => {
            pendingQueueRef.current.resetRefreshInFlight = false;
            socketLastSeenStreamSeqRef.current = message.stream_seq ?? 0;
            appliedStreamSeqRef.current.add(message.stream_seq ?? 0);
            publishPendingQueueState();
            schedulePendingReplay();
          });
          return;
        }
        appliedStreamSeqRef.current.add(message.stream_seq);
        socketLastSeenStreamSeqRef.current = Math.max(
          previousSeq,
          message.stream_seq,
        );
      }
      const viewportContinuityToken = beginViewportContinuity({
        kind: "scroll-only",
      });
      if (applyRecordChangedPatch(message.payload as RecordChangedPayload)) {
        advanceViewportContinuity(viewportContinuityToken);
        return;
      }
      void loadRowsRef.current({
        showLoading: false,
        viewportContinuityToken,
      });
    };

    const connect = () => {
      if (closed) {
        return;
      }
      socket = new WebSocket(changeSocketURL);
      activeSocketRef.current = socket;
      socketEstablishedRef.current = false;
      socket.onopen = () => {
        if (socket) {
          sendSessionEstablishment(socket);
        }
      };
      socket.onmessage = (event) => {
        if (!socket) {
          return;
        }
        handleMessage(socket, JSON.parse(event.data) as unknown);
      };
      socket.onclose = scheduleReconnect;
      socket.onerror = () => {
        socket?.close();
      };
    };

    connect();
    return () => {
      closed = true;
      if (reconnectTimer !== null) {
        window.clearTimeout(reconnectTimer);
      }
      if (presenceUpdateTimerRef.current !== null) {
        window.clearTimeout(presenceUpdateTimerRef.current);
        presenceUpdateTimerRef.current = null;
      }
      activeSocketRef.current = null;
      socketEstablishedRef.current = false;
      socket?.close();
    };
  }, [
    advanceViewportContinuity,
    applyRecordChangedPatch,
    beginViewportContinuity,
    changeSocketURL,
    incidentId,
    publishPendingQueueState,
    resolvePendingSocketTxn,
    scheduleAuthRecoveryProbe,
    schedulePendingReplay,
  ]);

  const queueScalarSave = useCallback(
    (
      rowKey: string,
      focusField: keyof RowValues,
      options: {
        allowZeroFieldCreate?: boolean;
        continueOnFreshDraft: boolean;
        preserveInputFocus: boolean;
        surface: TimelineScalarEditorSurface;
      },
      currentValue?: string,
    ) => {
      const focusKey = inputFocusKey(rowKey, focusField, options.surface);
      const rowSnapshot = rowsRef.current.find(
        (candidate) => candidate.key === rowKey,
      );
      const snapshot =
        rowSnapshot === undefined
          ? undefined
          : rowWithScalarEditorDrafts(rowSnapshot, {
              field: focusField,
              value: scalarDraftValuesRef.current.get(focusKey) ?? currentValue,
            });
      if (!snapshot) {
        return;
      }
      const binding = Object.values(timelineScalarBindingIndex).find(
        (candidate) => candidate.key === focusField,
      );
      if (
        snapshot.recordId !== null &&
        binding &&
        conflictQueueRef.current[`${snapshot.recordId}:${binding.fieldKey}`]
      ) {
        return;
      }

      const clientTxnId = nextClientTxnId();
      const payload =
        snapshot.recordId === null
          ? buildCreatePayload(snapshot, clientTxnId, {
              allowZeroFieldCreate: options.allowZeroFieldCreate === true,
            })
          : buildScalarPatchPayload(snapshot, clientTxnId);
      if (payload === null) {
        scalarDraftValuesRef.current.delete(focusKey);
        return;
      }

      const mutationSignature = buildStableMutationSignature(payload);
      if (pendingSignaturesRef.current.get(rowKey) === mutationSignature) {
        return;
      }
      const viewportContinuityToken = beginViewportContinuity(
        options.preserveInputFocus
          ? {
              kind: "input",
              focusKey: inputFocusKey(rowKey, focusField, options.surface),
            }
          : {
              kind: "scroll-only",
            },
      );
      pendingSignaturesRef.current.set(rowKey, mutationSignature);

      startTransition(() => {
        setRows((current) => {
          const nextRows = current.map((row) =>
            row.key === rowKey
              ? { ...row, pendingSignature: mutationSignature }
              : row,
          );
          rowsRef.current = nextRows;
          return nextRows;
        });
      });

      const targetPath =
        snapshot.recordId === null
          ? apiPath(
              apiBase,
              `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`,
            )
          : apiPath(apiBase, `/api/v1/records/${snapshot.recordId}`);
      enqueuePendingReplayUnit({
        id: `pending-${clientTxnId}`,
        kind: snapshot.recordId === null ? "create" : "patch",
        rowKey,
        recordId: snapshot.recordId,
        focusField,
        focusKey,
        surface: options.surface,
        method: snapshot.recordId === null ? "POST" : "PATCH",
        path: targetPath,
        payload,
        clientTxnId,
        mutationSignature,
        coalesceKey:
          snapshot.recordId === null
            ? `draft:${rowKey}`
            : `record:${snapshot.recordId}`,
        enqueueOrder: pendingReplayOrderRef.current,
        status: "queued",
        rowSnapshot: snapshot,
        continueOnFreshDraft: options.continueOnFreshDraft,
        detectAutoResolution: false,
        promoteToCommittedRowInspect: false,
        viewportContinuityToken,
      });
      pendingReplayOrderRef.current += 1;
    },
    [
      apiBase,
      beginViewportContinuity,
      enqueuePendingReplayUnit,
      incidentId,
      nextClientTxnId,
      rowWithScalarEditorDrafts,
    ],
  );

  const queueCollectionSave = useCallback(
    (
      rowKey: string,
      fieldKey: CollectionFieldKey,
      focusField: CollectionDraftKey,
      draftValueOverride?: string,
    ) => {
      const snapshot = rowsRef.current.find(
        (candidate) => candidate.key === rowKey,
      );
      if (!snapshot) {
        return;
      }
      const draftValue =
        draftValueOverride ?? snapshot.collectionDrafts[focusField];
      const collectionSnapshot =
        draftValueOverride === undefined
          ? snapshot
          : {
              ...snapshot,
              collectionDrafts: {
                ...snapshot.collectionDrafts,
                [focusField]: draftValue,
              },
            };
      const effectiveSnapshot = rowWithScalarEditorDrafts(collectionSnapshot);
      const clientTxnId = nextClientTxnId();
      const payload =
        snapshot.recordId === null
          ? buildCreatePayload(effectiveSnapshot, clientTxnId)
          : buildCollectionPatchPayload(
              effectiveSnapshot,
              fieldKey,
              draftValue,
              clientTxnId,
            );
      if (payload === null) {
        return;
      }

      const mutationSignature = buildStableMutationSignature(payload);
      if (pendingSignaturesRef.current.get(rowKey) === mutationSignature) {
        return;
      }
      const viewportContinuityToken = beginViewportContinuity(
        snapshot.recordId === null
          ? {
              kind: "scroll-only",
            }
          : {
              kind: "row-inspect",
              recordId: snapshot.recordId,
            },
      );
      pendingSignaturesRef.current.set(rowKey, mutationSignature);
      startTransition(() => {
        setRows((current) => {
          const nextRows = current.map((row) =>
            row.key === rowKey
              ? { ...row, pendingSignature: mutationSignature }
              : row,
          );
          rowsRef.current = nextRows;
          return nextRows;
        });
      });

      const targetPath =
        snapshot.recordId === null
          ? apiPath(
              apiBase,
              `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`,
            )
          : apiPath(apiBase, `/api/v1/records/${snapshot.recordId}`);
      enqueuePendingReplayUnit({
        id: `pending-${clientTxnId}`,
        kind: snapshot.recordId === null ? "create" : "patch",
        rowKey,
        recordId: snapshot.recordId,
        focusField,
        focusKey: inputFocusKey(rowKey, focusField, "grid"),
        surface: "grid",
        method: snapshot.recordId === null ? "POST" : "PATCH",
        path: targetPath,
        payload,
        clientTxnId,
        mutationSignature,
        coalesceKey:
          snapshot.recordId === null
            ? `draft:${rowKey}`
            : `record:${snapshot.recordId}`,
        enqueueOrder: pendingReplayOrderRef.current,
        status: "queued",
        rowSnapshot: effectiveSnapshot,
        continueOnFreshDraft: snapshot.recordId === null,
        detectAutoResolution: true,
        promoteToCommittedRowInspect: snapshot.recordId === null,
        viewportContinuityToken,
      });
      pendingReplayOrderRef.current += 1;
    },
    [
      apiBase,
      beginViewportContinuity,
      enqueuePendingReplayUnit,
      incidentId,
      nextClientTxnId,
      rowWithScalarEditorDrafts,
    ],
  );

  const queueAction = useCallback(
    (rowKey: string, action: "mark-reviewed" | "supersede") => {
      const snapshot = rowsRef.current.find(
        (candidate) => candidate.key === rowKey,
      );
      const replacementRecordId =
        action === "supersede"
          ? normalizeValue(replacementDrafts[rowKey] ?? "")
          : null;
      if (
        !snapshot ||
        snapshot.recordId === null ||
        snapshot.rowVersion === null ||
        (action === "supersede" && replacementRecordId === "")
      ) {
        return;
      }

      const recordId = snapshot.recordId;
      const clientTxnId = nextClientTxnId();
      const viewportContinuityToken = beginViewportContinuity({
        kind: "row-inspect",
        recordId,
      });
      beginSave();
      saveQueueRef.current = saveQueueRef.current
        .catch(() => undefined)
        .then(async () => {
          if (!(await waitForPendingReplayForRecord(recordId))) {
            clearViewportContinuity(viewportContinuityToken);
            finishSave("Conflict");
            return;
          }
          const currentRow = currentCommittedTimelineRow(recordId);
          if (currentRow === null) {
            clearViewportContinuity(viewportContinuityToken);
            finishSave("Conflict");
            return;
          }
          const body =
            action === "mark-reviewed"
              ? {
                  base_row_version: currentRow.rowVersion,
                  client_txn_id: clientTxnId,
                  reason: "Reviewed from workbook",
                }
              : {
                  base_row_version: currentRow.rowVersion,
                  client_txn_id: clientTxnId,
                  reason: "Superseded from workbook",
                  replacement_record_id: replacementRecordId,
                };
          trackPendingSocketTxn(clientTxnId);
          const result = await fetchJSON<TimelineActionEnvelope>(
            apiPath(apiBase, `/api/v1/records/${recordId}/${action}`),
            {
              method: "POST",
              body: JSON.stringify(body),
            },
          );
          if (!result.ok) {
            resolvePendingSocketTxn(clientTxnId);
            clearViewportContinuity(viewportContinuityToken);
            finishSave("Conflict");
            return;
          }

          await loadRowsRef.current({
            showLoading: false,
            viewportContinuityToken,
          });
          finishSave("Saved");
        });
    },
    [
      apiBase,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      currentCommittedTimelineRow,
      finishSave,
      nextClientTxnId,
      replacementDrafts,
      resolvePendingSocketTxn,
      trackPendingSocketTxn,
      waitForPendingReplayForRecord,
    ],
  );

  const fetchRecordHistory = useCallback(
    async (recordId: string): Promise<RecordHistoryData | null> => {
      const result = await fetchJSON<RecordHistoryEnvelope>(
        apiPath(apiBase, `/api/v1/records/${recordId}/history`),
      );
      if (!result.ok) {
        setRowHistory({
          recordId,
          status: "error",
          data: null,
          message: parseErrorMessage(result.payload),
        });
        return null;
      }
      const envelope = readEnvelope<RecordHistoryEnvelope>(result.payload);
      setRowHistory({
        recordId,
        status: "ready",
        data: envelope.data,
        message: null,
      });
      return envelope.data;
    },
    [apiBase],
  );

  const openRowHistory = useCallback(
    (recordId: string) => {
      setSelectedRowId(recordId);
      setRowHistory({
        recordId,
        status: "loading",
        data: rowHistory.recordId === recordId ? rowHistory.data : null,
        message: null,
      });
      void fetchRecordHistory(recordId);
    },
    [fetchRecordHistory, rowHistory.data, rowHistory.recordId],
  );

  const currentHistoryRecordId =
    selectedRow?.recordId ?? rowHistory.data?.record_id ?? rowHistory.recordId;
  const currentHistoryRowVersion =
    selectedRow?.rowVersion ?? rowHistory.data?.row_version ?? null;
  const currentHistoryDeleted = rowHistory.data?.deleted === true;

  const submitRowHistoryDeleteRestore = useCallback(
    (operation: "delete" | "restore") => {
      const recordId = currentHistoryRecordId;
      if (recordId === null || recordId === undefined) {
        return;
      }
      const rowVersion =
        operation === "restore"
          ? rowHistory.data?.row_version
          : (currentCommittedTimelineRow(recordId)?.rowVersion ??
            currentHistoryRowVersion);
      if (typeof rowVersion !== "number") {
        setRowHistory((current) => ({
          ...current,
          message: "Missing row version for destructive action.",
        }));
        return;
      }
      const clientTxnId = nextClientTxnId();
      const viewportContinuityTarget: ViewportContinuityTarget =
        selectedRow?.recordId === recordId
          ? { kind: "row-inspect", recordId }
          : { kind: "scroll-only" };
      const viewportContinuityToken = beginViewportContinuity(
        viewportContinuityTarget,
      );
      beginSave();
      setRowHistory((current) => ({ ...current, message: null }));
      saveQueueRef.current = saveQueueRef.current
        .catch(() => undefined)
        .then(async () => {
          if (!(await waitForPendingReplayForRecord(recordId))) {
            clearViewportContinuity(viewportContinuityToken);
            finishSave("Conflict");
            return;
          }
          trackPendingSocketTxn(clientTxnId);
          const path =
            operation === "delete"
              ? `/api/v1/records/${recordId}`
              : `/api/v1/records/${recordId}/restore`;
          const result = await fetchJSON<RecordDeleteRestoreEnvelope>(
            apiPath(apiBase, path),
            {
              method: operation === "delete" ? "DELETE" : "POST",
              body: JSON.stringify({
                base_row_version: rowVersion,
                client_txn_id: clientTxnId,
                reason:
                  operation === "delete"
                    ? "Deleted from workbook history"
                    : "Restored from workbook history",
              }),
            },
          );
          if (!result.ok) {
            resolvePendingSocketTxn(clientTxnId);
            clearViewportContinuity(viewportContinuityToken);
            setRowHistory((current) => ({
              ...current,
              message: parseErrorMessage(result.payload),
            }));
            finishSave("Conflict");
            return;
          }
          await fetchRecordHistory(recordId);
          if (operation === "restore") {
            setSelectedRowId(recordId);
          }
          await loadRowsRef.current({
            showLoading: false,
            viewportContinuityToken,
          });
          finishSave("Saved");
        });
    },
    [
      apiBase,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      currentCommittedTimelineRow,
      currentHistoryRecordId,
      currentHistoryRowVersion,
      fetchRecordHistory,
      finishSave,
      nextClientTxnId,
      resolvePendingSocketTxn,
      rowHistory.data?.row_version,
      selectedRow?.recordId,
      trackPendingSocketTxn,
      waitForPendingReplayForRecord,
    ],
  );

  const submitRowHistoryRollback = useCallback(
    (item: RecordHistoryItem, action: RecordHistoryRollbackAction) => {
      const recordId = currentHistoryRecordId;
      const rowVersion =
        currentCommittedTimelineRow(recordId ?? "")?.rowVersion ??
        currentHistoryRowVersion;
      if (recordId === null || recordId === undefined) {
        return;
      }
      if (typeof rowVersion !== "number") {
        setRowHistory((current) => ({
          ...current,
          message: "Missing row version for rollback.",
        }));
        return;
      }
      const target = buildRecordRollbackTargetFromHistoryAction(item, action);
      if (target === null) {
        return;
      }
      const clientTxnId = nextClientTxnId();
      const viewportContinuityTarget: ViewportContinuityTarget =
        selectedRow?.recordId === recordId
          ? { kind: "row-inspect", recordId }
          : { kind: "scroll-only" };
      const viewportContinuityToken = beginViewportContinuity(
        viewportContinuityTarget,
      );
      beginSave();
      setRowHistory((current) => ({ ...current, message: null }));
      saveQueueRef.current = saveQueueRef.current
        .catch(() => undefined)
        .then(async () => {
          if (!(await waitForPendingReplayForRecord(recordId))) {
            clearViewportContinuity(viewportContinuityToken);
            finishSave("Conflict");
            return;
          }
          trackPendingSocketTxn(clientTxnId);
          const result = await fetchJSON<RecordRollbackEnvelope>(
            apiPath(apiBase, `/api/v1/records/${recordId}/rollback`),
            {
              method: "POST",
              body: JSON.stringify({
                base_row_version: rowVersion,
                client_txn_id: clientTxnId,
                reason: "Rollback from workbook history",
                target,
              }),
            },
          );
          if (!result.ok) {
            resolvePendingSocketTxn(clientTxnId);
            clearViewportContinuity(viewportContinuityToken);
            setRowHistory((current) => ({
              ...current,
              message: parseErrorMessage(result.payload),
            }));
            finishSave("Conflict");
            return;
          }
          await fetchRecordHistory(recordId);
          await loadRowsRef.current({
            showLoading: false,
            viewportContinuityToken,
          });
          finishSave("Saved");
        });
    },
    [
      apiBase,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      currentCommittedTimelineRow,
      currentHistoryRecordId,
      currentHistoryRowVersion,
      fetchRecordHistory,
      finishSave,
      nextClientTxnId,
      resolvePendingSocketTxn,
      selectedRow?.recordId,
      trackPendingSocketTxn,
      waitForPendingReplayForRecord,
    ],
  );

  function submitMentionAction(
    mention: InspectorMention,
    action: "resolve_item" | "dismiss_item" | "revert_to_unresolved",
    resolvedRecordId?: string,
  ) {
    const snapshot = rowsRef.current.find(
      (candidate) => candidate.recordId === mention.rowRecordId,
    );
    if (!snapshot || snapshot.recordId === null) {
      return;
    }

    const recordId = snapshot.recordId;
    const clientTxnId = nextClientTxnId();
    const viewportContinuityToken = beginViewportContinuity(
      {
        kind: "row-inspect",
        recordId,
      },
      {
        followup:
          action === "resolve_item" && resolvedRecordId === undefined
            ? "entity-refresh"
            : "none",
      },
    );
    beginSave();
    setInspectorMessage(null);
    saveQueueRef.current = saveQueueRef.current
      .catch(() => undefined)
      .then(async () => {
        if (!(await waitForPendingReplayForRecord(recordId))) {
          clearViewportContinuity(viewportContinuityToken);
          finishSave("Conflict");
          return;
        }
        const currentRow = currentCommittedTimelineRow(recordId);
        const payload =
          currentRow === null
            ? null
            : buildMentionPatchPayload(
                currentRow,
                mention,
                action,
                clientTxnId,
                resolvedRecordId,
              );
        if (currentRow === null || payload === null) {
          clearViewportContinuity(viewportContinuityToken);
          finishSave("Conflict");
          return;
        }
        trackPendingSocketTxn(clientTxnId);
        const result = await fetchJSON<TimelineMutationEnvelope>(
          apiPath(apiBase, `/api/v1/records/${recordId}`),
          {
            method: "PATCH",
            body: JSON.stringify(payload),
          },
        );
        if (!result.ok) {
          resolvePendingSocketTxn(clientTxnId);
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage(parseErrorMessage(result.payload));
          finishSave("Conflict");
          return;
        }

        if (action === "dismiss_item") {
          setDismissedMentionsByRow((current) => {
            const rowMentions = current[recordId] ?? [];
            return {
              ...current,
              [recordId]: [
                ...rowMentions.filter(
                  (item) => item.itemRef !== mention.itemRef,
                ),
                {
                  rowRecordId: recordId,
                  fieldKey: mention.fieldKey,
                  entityType: mention.entityType,
                  itemRef: mention.itemRef,
                  rawText: mention.rawText,
                  resolvedRecordId: mention.resolvedRecordId,
                  resolutionMethod: mention.resolutionMethod,
                  autoResolved: mention.autoResolved,
                },
              ],
            };
          });
        }
        if (action === "revert_to_unresolved") {
          setDismissedMentionsByRow((current) => {
            const rowMentions = (current[recordId] ?? []).filter(
              (item) => item.itemRef !== mention.itemRef,
            );
            if (rowMentions.length < 1) {
              const next = { ...current };
              delete next[recordId];
              return next;
            }
            return {
              ...current,
              [recordId]: rowMentions,
            };
          });
        }

        const envelope = readEnvelope<TimelineMutationEnvelope>(result.payload);
        applyRowMutation(currentRow.key, envelope, {
          detectAutoResolution: false,
          viewportContinuityToken,
        });
        finishSave("Saved");
        if (action === "resolve_item" && resolvedRecordId === undefined) {
          try {
            await onRefreshEntities?.();
          } finally {
            settleViewportContinuityFollowup(viewportContinuityToken);
          }
        }
      });
  }

  const handleBlur = useCallback(
    (
      rowKey: string,
      focusField: keyof RowValues,
      surface: TimelineScalarEditorSurface,
      currentValue: string,
    ) => {
      queueScalarSave(
        rowKey,
        focusField,
        {
          continueOnFreshDraft: false,
          preserveInputFocus: false,
          surface,
        },
        currentValue,
      );
    },
    [queueScalarSave],
  );

  const handleKeyDown = useCallback(
    (
      event: ReactKeyboardEvent<HTMLInputElement | HTMLTextAreaElement>,
      rowKey: string,
      focusField: keyof RowValues,
      surface: TimelineScalarEditorSurface,
    ) => {
      if (event.key === "Enter" || event.key === "Tab") {
        event.preventDefault();
        queueScalarSave(
          rowKey,
          focusField,
          {
            continueOnFreshDraft: true,
            preserveInputFocus: true,
            surface,
          },
          event.currentTarget.value,
        );
      }
    },
    [queueScalarSave],
  );

  const handleCollectionKeyDown = useCallback(
    (
      event: ReactKeyboardEvent<HTMLInputElement>,
      rowKey: string,
      fieldKey: CollectionFieldKey,
      draftKey: CollectionDraftKey,
    ) => {
      if (event.key === "Enter" || event.key === "Tab") {
        event.preventDefault();
        queueCollectionSave(
          rowKey,
          fieldKey,
          draftKey,
          event.currentTarget.value,
        );
      }
    },
    [queueCollectionSave],
  );

  const handlePaste = useCallback(
    (
      rowKey: string,
      focusField: keyof RowValues,
      surface: TimelineScalarEditorSurface,
    ) => {
      window.setTimeout(() => {
        const editor = rowInputRefs.current.get(
          inputFocusKey(rowKey, focusField, surface),
        );
        if (editor) {
          setScalarEditorDraftValue(rowKey, focusField, surface, editor.value);
        }
        queueScalarSave(
          rowKey,
          focusField,
          {
            continueOnFreshDraft: false,
            preserveInputFocus: true,
            surface,
          },
          editor?.value,
        );
      }, 0);
    },
    [queueScalarSave, setScalarEditorDraftValue],
  );

  const sendPresenceUpdate = useCallback((presence: typeof currentPresence) => {
    const target = activeSocketRef.current;
    if (
      target === null ||
      !socketEstablishedRef.current ||
      !socketIsOpen(target)
    ) {
      return;
    }
    target.send(
      JSON.stringify({
        type: "presence_update",
        payload: {
          presence: workbookPresence(presence),
        },
      }),
    );
  }, []);

  const handleSelectRow = useCallback(
    (recordId: string) => {
      setSelectedRowId(recordId);
      setInspectorMessage(null);
      if (currentPresenceRef.current.mode === "editing") {
        return;
      }
      const next = { fieldKey: null, mode: "viewing" as const, recordId };
      currentPresenceRef.current = next;
      setCurrentPresence(next);
      sendPresenceUpdate(next);
    },
    [sendPresenceUpdate],
  );

  const handleEditModePresence = useCallback(
    (recordId: string | null, fieldKey: string, editing: boolean) => {
      const next = editing
        ? { fieldKey, mode: "editing" as const, recordId }
        : {
            fieldKey: null,
            mode: "viewing" as const,
            recordId: recordId ?? currentPresenceRef.current.recordId,
          };
      currentPresenceRef.current = next;
      setCurrentPresence(next);
      sendPresenceUpdate(next);
    },
    [sendPresenceUpdate],
  );

  const handleSelectMention = useCallback(
    (rowRecordId: string, itemRef: string) => {
      setSelectedRowId(rowRecordId);
      setSelectedMentionRef(itemRef);
      setInspectorMessage(null);
    },
    [],
  );

  const handleCreateBlankDraftRow = useCallback(
    (row: WorkbookRow) => {
      const activeRow =
        rowsRef.current.find((candidate) => candidate.key === row.key) ??
        rowsRef.current.find((candidate) => candidate.recordId === null) ??
        row;
      queueScalarSave(activeRow.key, "summary", {
        allowZeroFieldCreate: true,
        continueOnFreshDraft: true,
        preserveInputFocus: false,
        surface: "grid",
      });
    },
    [queueScalarSave],
  );

  const createAndAttachEvidenceFile = useCallback(
    async (file: File): Promise<string> => {
      const title = normalizeValue(file.name) || "Workbook attachment";
      const createEvidence = await fetchJSON<ViewMutationEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${evidenceViewSchemaId}/rows`,
        ),
        {
          method: "POST",
          body: JSON.stringify({
            client_txn_id: nextClientTxnId(),
            "evidence.title": title,
            "evidence.collector_party_text": "Workbook upload",
          }),
        },
      );
      if (!createEvidence.ok) {
        throw new Error(parseErrorMessage(createEvidence.payload));
      }
      const evidenceEnvelope = readEnvelope<ViewMutationEnvelope>(
        createEvidence.payload,
      );
      const evidenceRecord = evidenceEnvelope.data.row;

      const createBlob = await fetchJSON<ObjectBlobCreateEnvelope>(
        apiPath(apiBase, "/api/v1/object-blobs"),
        {
          method: "POST",
          body: JSON.stringify({
            incident_id: incidentId,
            client_txn_id: nextClientTxnId(),
            byte_size: file.size,
            filename_hint: file.name || null,
            content_type_hint: file.type || null,
          }),
        },
      );
      if (!createBlob.ok) {
        throw new Error(parseErrorMessage(createBlob.payload));
      }
      const blobEnvelope = readEnvelope<ObjectBlobCreateEnvelope>(
        createBlob.payload,
      );
      const uploadHref =
        blobEnvelope.data.upload_target.href.startsWith("/") && apiBase
          ? apiPath(apiBase, blobEnvelope.data.upload_target.href)
          : blobEnvelope.data.upload_target.href;
      const upload = await fetch(uploadHref, {
        method: blobEnvelope.data.upload_target.method ?? "PUT",
        credentials: "include",
        headers: {
          "Content-Type": file.type || "application/octet-stream",
        },
        body: file,
      });
      if (!upload.ok) {
        throw new Error(`upload_failed_${upload.status}`);
      }

      const attach = await fetchJSON<EvidenceAttachBlobEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/evidence-records/${evidenceRecord.record_id}/attach-blob`,
        ),
        {
          method: "POST",
          body: JSON.stringify({
            object_blob_id: blobEnvelope.data.object_blob_id,
            base_row_version: evidenceRecord.row_version,
            client_txn_id: nextClientTxnId(),
          }),
        },
      );
      if (!attach.ok) {
        throw new Error(parseErrorMessage(attach.payload));
      }
      return evidenceRecord.record_id;
    },
    [apiBase, incidentId, nextClientTxnId],
  );

  const attachEvidenceFileToTimeline = useCallback(
    (target: WorkbookRow, file: File) => {
      const snapshot =
        rowsRef.current.find((candidate) => candidate.key === target.key) ??
        target;
      const viewportContinuityToken = beginViewportContinuity(
        snapshot.recordId === null
          ? { kind: "scroll-only" }
          : { kind: "row-inspect", recordId: snapshot.recordId },
      );
      beginSave();
      setInspectorMessage("Uploading evidence.");

      saveQueueRef.current = saveQueueRef.current
        .catch(() => undefined)
        .then(async () => {
          try {
            const evidenceRecordId = await createAndAttachEvidenceFile(file);
            const clientTxnId = nextClientTxnId();
            const payload =
              snapshot.recordId === null
                ? buildAttachedEvidenceCreatePayload(
                    evidenceRecordId,
                    clientTxnId,
                  )
                : buildAttachedEvidencePatchPayload(
                    snapshot,
                    evidenceRecordId,
                    clientTxnId,
                  );
            if (payload === null) {
              throw new Error("invalid_timeline_row");
            }

            trackPendingSocketTxn(clientTxnId);
            const targetPath =
              snapshot.recordId === null
                ? apiPath(
                    apiBase,
                    `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`,
                  )
                : apiPath(apiBase, `/api/v1/records/${snapshot.recordId}`);
            const result = await fetchJSON<TimelineMutationEnvelope>(
              targetPath,
              {
                method: snapshot.recordId === null ? "POST" : "PATCH",
                body: JSON.stringify(payload),
              },
            );
            if (!result.ok) {
              resolvePendingSocketTxn(clientTxnId);
              throw new Error(parseErrorMessage(result.payload));
            }

            const envelope = readEnvelope<TimelineMutationEnvelope>(
              result.payload,
            );
            applyRowMutation(snapshot.key, envelope, {
              continueOnFreshDraft: snapshot.recordId === null,
              promoteToCommittedRowInspect: snapshot.recordId === null,
              detectAutoResolution: false,
              viewportContinuityToken,
            });
            setInspectorMessage("Evidence attached.");
            finishSave("Saved");
          } catch (error) {
            clearViewportContinuity(viewportContinuityToken);
            setInspectorMessage(
              error instanceof Error
                ? error.message
                : "Evidence attach failed.",
            );
            finishSave("Conflict");
          }
        });
    },
    [
      apiBase,
      applyRowMutation,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      createAndAttachEvidenceFile,
      finishSave,
      incidentId,
      nextClientTxnId,
      resolvePendingSocketTxn,
      trackPendingSocketTxn,
    ],
  );

  const handleTimelineEvidenceFiles = useCallback(
    (target: WorkbookRow, files: FileList | File[]) => {
      const [file] = Array.from(files);
      if (!file) {
        return;
      }
      attachEvidenceFileToTimeline(target, file);
    },
    [attachEvidenceFileToTimeline],
  );

  const timelineBindingLabel = useCallback((fieldKey: string) => {
    return timelineContract.fieldMap[fieldKey]?.label ?? fieldKey;
  }, []);

  const timelineScalarControlId = useCallback(
    (
      row: WorkbookRow,
      binding: TimelineScalarBinding,
      surface: TimelineScalarEditorSurface,
    ) => {
      return ["timeline-editor", surface, row.key, binding.fieldKey]
        .map((value) => value.replace(/[^a-zA-Z0-9_-]+/g, "-"))
        .join("-");
    },
    [],
  );

  const renderTimelineScalarControl = useCallback(
    (
      row: WorkbookRow,
      binding: TimelineScalarBinding,
      surface: TimelineScalarEditorSurface,
      controlId: string,
    ) => {
      const label = timelineBindingLabel(binding.fieldKey);
      const gridAccessibleLabel =
        surface === "grid"
          ? `${label} ${row.recordId ?? "draft row"}`
          : undefined;
      const gridDataTestId =
        row.recordId === null
          ? draftCellTestId(binding.key)
          : rowCellTestId(row.recordId, binding.key);
      const dataTestId =
        surface === "grid" ? gridDataTestId : `${gridDataTestId}-inspector`;
      const conflictKey =
        row.recordId === null ? null : `${row.recordId}:${binding.fieldKey}`;
      const localConflict =
        conflictKey === null ? undefined : conflictQueue[conflictKey];
      const sameCellPresence = editingPresenceForCell(
        row.recordId,
        binding.fieldKey,
      );
      const sameCellVisible = visiblePresence(sameCellPresence, 1);
      return (
        <>
          <TimelineScalarEditor
            key={inputFocusKey(row.key, binding.key, surface)}
            accessibleLabel={gridAccessibleLabel}
            blockedByConflict={localConflict !== undefined}
            committedValue={row.values[binding.key]}
            controlId={controlId}
            dataTestId={dataTestId}
            draftValue={scalarDraftValuesRef.current.get(
              inputFocusKey(row.key, binding.key, surface),
            )}
            field={binding.key}
            multiline={binding.multiline}
            onEditModeChange={handleEditModePresence}
            registerInput={registerInput}
            presenceFieldKey={binding.fieldKey}
            rowKey={row.key}
            rowRecordId={row.recordId}
            surface={surface}
            onBlurCommit={handleBlur}
            onDraftChange={setScalarEditorDraftValue}
            onFocusRecord={handleSelectRow}
            onKeyCommit={handleKeyDown}
            onPasteCommit={handlePaste}
          />
          {localConflict && surface === "grid" ? (
            <button
              type="button"
              data-testid={conflictMarkerTestId(
                row.recordId ?? "draft",
                binding.fieldKey,
              )}
              style={conflictMarkerStyle}
              onClick={() => setActiveConflictKey(localConflict.key)}
            >
              Conflict
            </button>
          ) : null}
          {sameCellPresence.length > 0 && surface === "grid" ? (
            <span
              aria-label={`${sameCellPresence
                .map((presence) => presence.display_name)
                .join(", ")} editing ${timelineBindingLabel(binding.fieldKey)}`}
              data-testid={cellPresenceMarkerTestId(
                row.recordId ?? "draft",
                binding.fieldKey,
              )}
              role="img"
              style={cellPresenceStyle}
            >
              {sameCellVisible.shown.map((presence) =>
                displayInitials(presence.display_name),
              )}
              {sameCellVisible.overflow > 0
                ? ` +${sameCellVisible.overflow}`
                : ""}
            </span>
          ) : null}
        </>
      );
    },
    [
      conflictQueue,
      editingPresenceForCell,
      handleBlur,
      handleEditModePresence,
      handleKeyDown,
      handlePaste,
      handleSelectRow,
      registerInput,
      setScalarEditorDraftValue,
      timelineBindingLabel,
    ],
  );

  const renderTimelineGridEditor = useCallback(
    (row: WorkbookRow, binding: TimelineScalarBinding) => {
      return renderTimelineScalarControl(
        row,
        binding,
        "grid",
        timelineScalarControlId(row, binding, "grid"),
      );
    },
    [renderTimelineScalarControl, timelineScalarControlId],
  );

  const renderTimelineInspectorEditor = useCallback(
    (row: WorkbookRow, binding: TimelineScalarBinding) => {
      const controlId = timelineScalarControlId(row, binding, "inspector");
      return (
        <div key={binding.fieldKey} style={labelStyle}>
          <label htmlFor={controlId}>
            {timelineBindingLabel(binding.fieldKey)}
          </label>
          {renderTimelineScalarControl(row, binding, "inspector", controlId)}
        </div>
      );
    },
    [
      renderTimelineScalarControl,
      timelineBindingLabel,
      timelineScalarControlId,
    ],
  );

  const renderTimelineCollectionInput = useCallback(
    (row: WorkbookRow, binding: TimelineCollectionBinding) => {
      const label = timelineBindingLabel(binding.fieldKey);
      const items = row.collectionValues[binding.draftKey];
      return (
        <>
          <div
            data-testid={
              row.recordId === null
                ? draftCellTestId(`${binding.draftKey}-items`)
                : relationshipItemsTestId(row.recordId, binding.draftKey)
            }
            style={relationshipItemsWrapStyle}
          >
            {items.length > 0 ? (
              binding.collectionKind === "relationship" ? (
                items.map((item) => (
                  <RelationshipChip
                    key={item.itemRef}
                    entityIndex={entityIndex}
                    item={item as CollectionItem}
                    onSelect={() => {
                      if (row.recordId) {
                        handleSelectMention(row.recordId, item.itemRef);
                      }
                    }}
                  />
                ))
              ) : (
                items.map((item) => (
                  <span key={item.itemRef} style={tagChipStyle}>
                    {(item as TagCollectionItem).displayText}
                  </span>
                ))
              )
            ) : (
              <span style={emptyRelationshipStyle}>No items</span>
            )}
          </div>
          <input
            aria-label={`${label} ${row.recordId ?? "draft row"}`}
            data-testid={
              row.recordId === null
                ? draftCellTestId(`${binding.draftKey}-input`)
                : rowCellTestId(row.recordId, `${binding.draftKey}-input`)
            }
            key={`${row.key}:${binding.draftKey}:${row.rowVersion ?? "draft"}`}
            ref={(element) => {
              registerInput(row.key, binding.draftKey, "grid", element);
            }}
            style={inputStyle}
            type="text"
            defaultValue={row.collectionDrafts[binding.draftKey]}
            onBlur={(event) => {
              queueCollectionSave(
                row.key,
                binding.fieldKey,
                binding.draftKey,
                event.currentTarget.value,
              );
            }}
            onFocus={() => {
              if (row.recordId) {
                handleSelectRow(row.recordId);
              }
            }}
            onKeyDown={(event) => {
              handleCollectionKeyDown(
                event,
                row.key,
                binding.fieldKey,
                binding.draftKey,
              );
            }}
            placeholder={`Add ${label.toLowerCase()} token`}
          />
        </>
      );
    },
    [
      entityIndex,
      handleCollectionKeyDown,
      handleSelectRow,
      handleSelectMention,
      registerInput,
      timelineBindingLabel,
      queueCollectionSave,
    ],
  );

  const timelineColumns = useMemo<readonly GridColumn<WorkbookRow>[]>(
    () => [
      {
        fieldKey: "timeline.capture_state",
        headerTestId: gridSortHeaderTestId(
          "timeline",
          "timeline.capture_state",
        ),
        label: "State",
        width: 136,
        renderCell: (row) => {
          const rowPresence = presenceForRow(row.recordId);
          const rowPresenceVisible = visiblePresence(rowPresence, 3);
          return (
            <div style={stateCellStyle}>
              {rowPresence.length > 0 ? (
                <span
                  aria-label={`${rowPresence
                    .map((presence) => presence.display_name)
                    .join(", ")} focused on this row`}
                  data-testid={rowPresenceMarkerTestId(row.recordId ?? "draft")}
                  role="img"
                  style={rowPresenceStyle}
                >
                  {rowPresenceVisible.shown
                    .map((presence) => displayInitials(presence.display_name))
                    .join(" ")}
                  {rowPresenceVisible.overflow > 0
                    ? ` +${rowPresenceVisible.overflow}`
                    : ""}
                </span>
              ) : null}
              <span
                data-testid={
                  row.recordId === null
                    ? "draft-row-capture-state"
                    : rowCellTestId(row.recordId, "capture-state")
                }
              >
                {row.captureState}
              </span>
            </div>
          );
        },
        sortableFieldKey: "timeline.capture_state",
      },
      {
        fieldKey: "row_version",
        label: "Version",
        width: 96,
        renderCell: (row) => (
          <span
            data-testid={
              row.recordId === null
                ? "draft-row-version"
                : rowCellTestId(row.recordId, "row-version")
            }
          >
            {row.rowVersion ?? "new"}
          </span>
        ),
      },
      ...timelineVisibleBindings.map(
        (binding): GridColumn<WorkbookRow> => ({
          fieldKey: binding.fieldKey,
          headerTestId: gridSortHeaderTestId("timeline", binding.fieldKey),
          label: timelineBindingLabel(binding.fieldKey),
          width: timelineColumnWidth(binding.fieldKey),
          renderCell: (row) => {
            if (binding.kind === "scalar") {
              return renderTimelineGridEditor(row, binding);
            }
            if (binding.kind === "collection") {
              return renderTimelineCollectionInput(row, binding);
            }
            if (binding.fieldKey === "timeline.evidence_count") {
              const count = stringifyGridValue(
                readCellValue(row.rawRow, binding.fieldKey),
              );
              const hasEvidence = Boolean(
                readCellValue(row.rawRow, "timeline.has_evidence"),
              );
              return (
                <span style={timelineEvidenceCellStyle}>
                  <span
                    data-testid={
                      row.recordId === null
                        ? undefined
                        : rowCellTestId(row.recordId, binding.fieldKey)
                    }
                  >
                    {count === "" ? "0" : count}
                  </span>
                  {row.recordId === null ? null : (
                    <span
                      data-testid={rowCellTestId(
                        row.recordId,
                        "timeline.has_evidence",
                      )}
                      style={
                        hasEvidence
                          ? timelineEvidenceFlagOnStyle
                          : timelineEvidenceFlagOffStyle
                      }
                      title={
                        hasEvidence
                          ? "Timeline row has evidence"
                          : "Timeline row has no evidence"
                      }
                    >
                      {String(hasEvidence)}
                    </span>
                  )}
                </span>
              );
            }
            const text = stringifyGridValue(
              readCellValue(row.rawRow, binding.fieldKey),
            );
            return (
              <span
                data-testid={
                  row.recordId === null
                    ? undefined
                    : rowCellTestId(row.recordId, binding.fieldKey)
                }
                style={bodyStyle}
              >
                {text === "" ? "—" : text}
              </span>
            );
          },
          sortableFieldKey: resolveHeaderSortFieldKey(
            timelineContract,
            binding.fieldKey,
          ),
        }),
      ),
    ],
    [
      presenceForRow,
      renderTimelineCollectionInput,
      renderTimelineGridEditor,
      timelineBindingLabel,
    ],
  );

  const timelineActionsColumn = useMemo<GridActionsColumn<WorkbookRow>>(
    () => ({
      headerTestId: gridActionsHeaderTestId("timeline"),
      label: "Actions",
      minWidth: 296,
      width: 296,
      renderCell: ({ data: row }) =>
        row.recordId === null ? (
          <div style={actionStackStyle}>
            <DraftRowCreateButton
              onCreate={handleCreateBlankDraftRow}
              row={row}
            />
          </div>
        ) : (
          <div style={actionStackStyle}>
            <div style={timelineActionTopRowStyle}>
              <button
                data-testid={rowInspectButtonTestId(row.recordId)}
                style={timelineActionButtonStyle}
                type="button"
                onClick={() => {
                  handleSelectRow(row.recordId ?? "");
                }}
              >
                Inspect
              </button>
              <button
                data-testid={`row-history-open-${row.recordId}`}
                style={timelineActionButtonStyle}
                type="button"
                onClick={() => {
                  openRowHistory(row.recordId ?? "");
                }}
              >
                History
              </button>
            </div>
            <button
              data-testid={`row-${row.recordId}-mark-reviewed`}
              disabled={
                row.captureState === "reviewed" ||
                row.captureState === "superseded"
              }
              style={timelineActionButtonStyle}
              type="button"
              onClick={() => {
                queueAction(row.key, "mark-reviewed");
              }}
            >
              Mark reviewed
            </button>
            <input
              data-testid={`row-${row.recordId}-replacement-id`}
              placeholder="Replacement record id"
              style={timelineReplacementInputStyle}
              type="text"
              value={replacementDrafts[row.key] ?? ""}
              onChange={(event) => {
                const value = event.target.value;
                setReplacementDrafts((current) => ({
                  ...current,
                  [row.key]: value,
                }));
              }}
            />
            <button
              data-testid={`row-${row.recordId}-supersede`}
              disabled={
                row.captureState === "superseded" ||
                normalizeValue(replacementDrafts[row.key] ?? "") === ""
              }
              style={timelineActionButtonStyle}
              type="button"
              onClick={() => {
                queueAction(row.key, "supersede");
              }}
            >
              Supersede
            </button>
          </div>
        ),
    }),
    [
      handleCreateBlankDraftRow,
      handleSelectRow,
      openRowHistory,
      queueAction,
      replacementDrafts,
    ],
  );

  const timelineGridRows = useMemo<readonly GridRow<WorkbookRow>[]>(
    () =>
      rows.map((row) => ({
        key: row.key,
        recordId: row.recordId,
        data: row,
        onSelect: () => {
          if (row.recordId) {
            handleSelectRow(row.recordId);
          }
        },
        selected: row.recordId !== null && row.recordId === selectedRowId,
        testId:
          row.recordId === null ? undefined : `timeline-row-${row.recordId}`,
        variant: row.recordId === null ? "draft" : "default",
      })),
    [handleSelectRow, rows, selectedRowId],
  );

  const getTimelineGroupLabel = useCallback(
    (row: WorkbookRow, fieldKey: string) => timelineGroupLabel(row, fieldKey),
    [],
  );
  const getTimelineGroupRowTestId = useCallback(
    (fieldKey: string, value: string) =>
      gridGroupRowTestId("timeline", fieldKey, value),
    [],
  );

  function renderInspectorFieldEditors(row: WorkbookRow) {
    return (
      <section style={inspectorSectionStyle}>
        <h3 style={sectionTitleStyle}>Details</h3>
        <div style={inspectorActionStackStyle}>
          {timelineInspectorBindings.map((binding) =>
            renderTimelineInspectorEditor(row, binding),
          )}
        </div>
      </section>
    );
  }

  function renderEvidenceAttachSection(row: WorkbookRow) {
    const inputTestId =
      row.recordId === null
        ? "timeline-evidence-file-draft"
        : `timeline-evidence-file-${row.recordId}`;
    const evidenceCount = stringifyGridValue(
      readCellValue(row.rawRow, "timeline.evidence_count"),
    );
    return (
      <section
        data-testid={
          row.recordId === null
            ? "timeline-evidence-attach-draft"
            : `timeline-evidence-attach-${row.recordId}`
        }
        style={inspectorSectionStyle}
        aria-label="Timeline evidence attachment"
        onDragOver={(event) => {
          event.preventDefault();
        }}
        onDrop={(event) => {
          event.preventDefault();
          handleTimelineEvidenceFiles(row, event.dataTransfer.files);
        }}
        onPaste={(event) => {
          if (event.clipboardData.files.length > 0) {
            handleTimelineEvidenceFiles(row, event.clipboardData.files);
          }
        }}
      >
        <h3 style={sectionTitleStyle}>Evidence</h3>
        {row.recordId !== null ? (
          <p style={bodyStyle}>
            Attached evidence count:{" "}
            {evidenceCount === "" ? "0" : evidenceCount}
          </p>
        ) : null}
        <label style={labelStyle}>
          Attach file
          <input
            data-testid={inputTestId}
            style={inputStyle}
            type="file"
            accept="image/*,.txt,.pdf,text/plain,application/pdf"
            onChange={(event) => {
              handleTimelineEvidenceFiles(row, event.currentTarget.files ?? []);
              event.currentTarget.value = "";
            }}
          />
        </label>
      </section>
    );
  }

  function renderRowHistorySection() {
    const recordId = currentHistoryRecordId;
    const historyData = rowHistory.data;
    const selectedActiveRow =
      recordId !== null &&
      recordId !== undefined &&
      selectedRow?.recordId === recordId
        ? selectedRow
        : null;
    return (
      <section data-testid="row-history-panel" style={inspectorSectionStyle}>
        <div style={historySectionHeaderStyle}>
          <h3 style={sectionTitleStyle}>Row history</h3>
          {selectedActiveRow !== null ? (
            <button
              data-testid={`row-history-open-${selectedActiveRow.recordId}-inspector`}
              style={secondaryActionButtonStyle}
              type="button"
              onClick={() => {
                openRowHistory(selectedActiveRow.recordId ?? "");
              }}
            >
              Refresh history
            </button>
          ) : null}
        </div>
        {recordId !== null && recordId !== undefined ? (
          <p style={historyMetaStyle}>Record {recordId}</p>
        ) : null}
        {rowHistory.status === "idle" && selectedActiveRow !== null ? (
          <button
            data-testid="row-history-open-selected"
            style={actionButtonStyle}
            type="button"
            onClick={() => {
              openRowHistory(selectedActiveRow.recordId ?? "");
            }}
          >
            Open history
          </button>
        ) : null}
        {rowHistory.status === "loading" ? (
          <p data-testid="row-history-loading" style={bodyStyle}>
            Loading history...
          </p>
        ) : null}
        {rowHistory.message ? (
          <p data-testid="row-history-message" style={genericErrorTextStyle}>
            {rowHistory.message}
          </p>
        ) : null}
        {historyData !== null ? (
          <>
            <dl style={historyMetaGridStyle}>
              <div>
                <dt style={detailTermStyle}>Current row version</dt>
                <dd style={detailValueStyle}>{historyData.row_version}</dd>
              </div>
              <div>
                <dt style={detailTermStyle}>Deleted</dt>
                <dd style={detailValueStyle}>
                  {historyData.deleted ? "yes" : "no"}
                </dd>
              </div>
            </dl>
            <div style={inlineButtonRowStyle}>
              {selectedActiveRow !== null && !historyData.deleted ? (
                <button
                  data-testid="row-history-delete"
                  style={destructiveActionButtonStyle}
                  type="button"
                  onClick={() => {
                    submitRowHistoryDeleteRestore("delete");
                  }}
                >
                  Soft-delete row
                </button>
              ) : null}
              {historyData.deleted ? (
                <button
                  data-testid="row-history-restore"
                  style={actionButtonStyle}
                  type="button"
                  onClick={() => {
                    submitRowHistoryDeleteRestore("restore");
                  }}
                >
                  Restore row
                </button>
              ) : null}
            </div>
            <ol style={historyListStyle}>
              {historyData.items.map((item, index) => {
                const itemKey =
                  item.history_entry_ref ??
                  `${item.change_set_id}:${item.revision_no}:${item.operation}`;
                const actionButtons = item.available_rollback_actions.flatMap(
                  (action) => {
                    const target = buildRecordRollbackTargetFromHistoryAction(
                      item,
                      action,
                    );
                    if (target === null) {
                      return [];
                    }
                    const label =
                      action === "history_entry"
                        ? "Rollback entry"
                        : action === "change_set"
                          ? "Rollback change set"
                          : "Restore row fields";
                    return [
                      <button
                        data-testid={`row-history-action-${index}-${action}`}
                        key={action}
                        style={
                          action === "row_restore"
                            ? actionButtonStyle
                            : secondaryActionButtonStyle
                        }
                        type="button"
                        onClick={() => {
                          submitRowHistoryRollback(item, action);
                        }}
                      >
                        {label}
                      </button>,
                    ];
                  },
                );
                return (
                  <li
                    data-testid={`row-history-item-${index}`}
                    key={itemKey}
                    style={historyItemStyle}
                  >
                    <div style={historyItemHeaderStyle}>
                      <strong>{item.operation}</strong>
                      <time dateTime={item.committed_at}>
                        {formatHistoryTimestamp(item.committed_at)}
                      </time>
                    </div>
                    <dl style={detailListStyle}>
                      <div>
                        <dt style={detailTermStyle}>Actor</dt>
                        <dd style={detailValueStyle}>{item.actor_user_id}</dd>
                      </div>
                      <div>
                        <dt style={detailTermStyle}>Diff</dt>
                        <dd style={detailValueStyle}>
                          {item.diff_summary.summary}
                        </dd>
                      </div>
                      <div>
                        <dt style={detailTermStyle}>Change set</dt>
                        <dd style={detailValueStyle}>{item.change_set_id}</dd>
                      </div>
                    </dl>
                    {actionButtons.length > 0 ? (
                      <div style={inlineButtonRowStyle}>{actionButtons}</div>
                    ) : (
                      <p style={emptyRelationshipStyle}>No rollback action</p>
                    )}
                  </li>
                );
              })}
            </ol>
          </>
        ) : null}
      </section>
    );
  }

  const activeConflict =
    activeConflictKey === null
      ? null
      : (conflictQueue[activeConflictKey] ?? null);

  useEffect(() => {
    if (activeConflict === null) {
      return;
    }
    window.setTimeout(() => {
      (
        document.querySelector(
          '[data-testid="conflict-resolver-summary"]',
        ) as HTMLElement | null
      )?.focus();
    }, 0);
  }, [activeConflict]);

  const clearLocalConflict = useCallback(
    (conflict: LocalConflictState) => {
      setConflictQueueState((current) => {
        const next = { ...current };
        delete next[conflict.key];
        updateSaveStateForConflicts(next);
        return next;
      });
      setActiveConflictKey((current) =>
        current === conflict.key ? null : current,
      );
      restoreConflictFocus(conflict.focusKey);
      schedulePendingReplay();
    },
    [
      restoreConflictFocus,
      schedulePendingReplay,
      setConflictQueueState,
      updateSaveStateForConflicts,
    ],
  );

  const submitConflictResolution = useCallback(
    (
      conflict: LocalConflictState,
      resolutionKind: "keep_saved" | "use_unsaved" | "merged_value",
    ) => {
      const body: Record<string, unknown> = {
        conflict_token: conflict.conflict.conflict_token,
        resolution_kind: resolutionKind,
        client_txn_id: nextClientTxnId(),
      };
      if (resolutionKind === "use_unsaved") {
        body.resolved_value = conflict.localValue;
      } else if (resolutionKind === "merged_value") {
        body.resolved_value =
          conflict.conflict.conflict_resolution_class === "collection_review"
            ? conflict.localValue
            : conflict.mergedDraft;
      }
      setSaveState("Syncing");
      saveQueueRef.current = saveQueueRef.current
        .catch(() => undefined)
        .then(async () => {
          const result = await fetchJSON<TimelineMutationEnvelope>(
            apiPath(
              apiBase,
              `/api/v1/records/${conflict.conflict.record_id}/conflicts/${conflict.conflict.conflict_token}/resolve`,
            ),
            {
              method: "POST",
              body: JSON.stringify(body),
            },
          );
          if (!result.ok) {
            const refreshedConflict = parseSameFieldConflict(result.payload);
            if (refreshedConflict !== null) {
              registerSameFieldConflict(
                refreshedConflict,
                conflict.focusKey,
                "grid",
              );
              setSaveState("Conflict");
              return;
            }
            setSaveState("Conflict");
            return;
          }
          const envelope = readEnvelope<TimelineMutationEnvelope>(
            result.payload,
          );
          scalarDraftValuesRef.current.delete(conflict.focusKey);
          const binding = timelineScalarBindingForField(
            conflict.conflict.field_key,
          );
          if (binding !== null) {
            for (const surface of timelineScalarEditorSurfaces) {
              scalarDraftValuesRef.current.delete(
                inputFocusKey(
                  conflict.conflict.record_id,
                  binding.key,
                  surface,
                ),
              );
            }
          }
          applyRowMutation(conflict.conflict.record_id, envelope, {
            viewportContinuityToken: beginViewportContinuity({
              kind: "input",
              focusKey: conflict.focusKey,
            }),
          });
          clearLocalConflict(conflict);
        });
    },
    [
      apiBase,
      applyRowMutation,
      beginViewportContinuity,
      clearLocalConflict,
      nextClientTxnId,
      registerSameFieldConflict,
    ],
  );

  if (isInitialLoading) {
    return (
      <section style={panelStyle}>
        <p style={eyebrowStyle}>Timeline</p>
        <h1 style={headlineStyle}>Loading projection-backed rows.</h1>
      </section>
    );
  }

  if (loadError !== null) {
    return (
      <section style={panelStyle}>
        <p style={eyebrowStyle}>Timeline</p>
        <h1 style={headlineStyle}>Timeline load failed.</h1>
        <p style={bodyStyle}>{loadError}</p>
      </section>
    );
  }

  const headerPresence = visiblePresence(activeSheetPresenceRecords, 5);

  return (
    <section style={workbookStyle}>
      <header style={headerStyle}>
        <button
          aria-label="Blur timeline inputs"
          data-testid="timeline-blur-surface"
          tabIndex={-1}
          type="button"
          style={blurSurfaceButtonStyle}
          onMouseDown={(event) => {
            event.currentTarget.focus();
          }}
        />
        <div>
          <p style={eyebrowStyle}>Phase 3 Workbook</p>
          <h1 style={headlineStyle}>Timeline mutation substrate</h1>
          <p style={bodyStyle}>Incident {incidentId}</p>
        </div>
        <div style={statusClusterStyle}>
          <span style={statusLabelStyle}>Save State</span>
          <strong
            aria-live="polite"
            data-testid={saveStateTestId()}
            style={{
              ...statusValueStyle,
              color:
                saveState === "Conflict"
                  ? "rgb(145 30 30)"
                  : saveState === "Syncing"
                    ? "rgb(146 64 14)"
                    : "rgb(21 128 61)",
            }}
          >
            {saveState}
          </strong>
          <div
            aria-label={`${activeSheetPresenceRecords.length} collaborators present on this sheet`}
            data-testid="presence-header"
            role="status"
            style={headerPresenceStyle}
          >
            {headerPresence.shown.length === 0 ? (
              <span style={presenceEmptyStyle}>Presence</span>
            ) : (
              headerPresence.shown.map((presence) => (
                <span
                  key={presence.connection_id}
                  title={presence.display_name}
                  style={presenceAvatarStyle}
                >
                  {displayInitials(presence.display_name)}
                </span>
              ))
            )}
            {headerPresence.overflow > 0 ? (
              <span style={presenceOverflowStyle}>
                +{headerPresence.overflow}
              </span>
            ) : null}
          </div>
        </div>
      </header>

      {autoResolutionNotices.length > 0 ? (
        <aside style={noticeStackStyle}>
          {autoResolutionNotices.map((notice) => (
            <div
              key={notice.itemRef}
              data-testid={`auto-resolution-notice-${sanitizeTestId(notice.itemRef)}`}
              style={noticeCardStyle}
            >
              <p style={noticeTitleStyle}>Auto-resolved mention</p>
              <p style={bodyStyle}>
                Raw token <strong>{notice.rawText}</strong> matched{" "}
                <strong>
                  {entityIndex[notice.resolvedRecordId]?.label ??
                    notice.rawText}
                </strong>
                {notice.matchedAliasText ? (
                  <>
                    {" "}
                    via alias <strong>{notice.matchedAliasText}</strong>
                  </>
                ) : null}
                .
              </p>
              <div style={inlineButtonRowStyle}>
                <button
                  style={secondaryActionButtonStyle}
                  type="button"
                  onClick={() => {
                    const row = rowsRef.current.find(
                      (candidate) => candidate.recordId === notice.rowRecordId,
                    );
                    const activeItem =
                      row === undefined
                        ? undefined
                        : [
                            ...row.collectionValues.hostRefs,
                            ...row.collectionValues.identityRefs,
                          ].find(
                            (item) =>
                              item.itemRef === notice.itemRef &&
                              item.itemKind === "resolved_ref",
                          );
                    if (activeItem && row?.recordId) {
                      submitMentionAction(
                        {
                          rowRecordId: row.recordId,
                          fieldKey: notice.fieldKey,
                          entityType: activeItem.entityType,
                          itemRef: activeItem.itemRef,
                          rawText: activeItem.rawText,
                          resolvedRecordId: activeItem.resolvedRecordId,
                          resolutionMethod: activeItem.resolutionMethod,
                          autoResolved: activeItem.autoResolved,
                          status: "resolved",
                          displayText: activeItem.displayText,
                          provenance: activeItem.provenance,
                          confidence: activeItem.confidence,
                          matchedAliasText: activeItem.matchedAliasText,
                        },
                        "revert_to_unresolved",
                      );
                    }
                    setAutoResolutionNotices((current) =>
                      current.filter((item) => item.itemRef !== notice.itemRef),
                    );
                  }}
                >
                  Undo
                </button>
                <button
                  style={secondaryActionButtonStyle}
                  type="button"
                  onClick={() => {
                    handleSelectMention(notice.rowRecordId, notice.itemRef);
                    setAutoResolutionNotices((current) =>
                      current.filter((item) => item.itemRef !== notice.itemRef),
                    );
                  }}
                >
                  Review
                </button>
              </div>
            </div>
          ))}
        </aside>
      ) : null}

      {pendingQueueSnapshot.overflowMessage !== null ||
      pendingQueueSnapshot.haltedMessage !== null ||
      pendingQueueSnapshot.authPaused ||
      pendingQueueSnapshot.queuedCount + pendingQueueSnapshot.inFlightCount >
        0 ? (
        <aside
          data-testid={pendingQueueNoticeTestId()}
          role="status"
          style={noticeCardStyle}
        >
          <p style={noticeTitleStyle}>Queued edits</p>
          <p style={bodyStyle}>
            {pendingQueueSnapshot.overflowMessage ??
              pendingQueueSnapshot.haltedMessage ??
              (pendingQueueSnapshot.authPaused
                ? "Authentication is required before queued edits can replay."
                : "Queued edits are waiting to replay.")}
          </p>
          <p data-testid={pendingQueueCountTestId()} style={bodyStyle}>
            Pending units:{" "}
            {pendingQueueSnapshot.queuedCount +
              pendingQueueSnapshot.inFlightCount}
          </p>
        </aside>
      ) : null}

      <div style={splitShellStyle}>
        <div>
          {refreshError !== null ? (
            <aside
              data-testid="timeline-refresh-error"
              role="status"
              style={noticeCardStyle}
            >
              <p style={bodyStyle}>{refreshError}</p>
            </aside>
          ) : null}
          <WorkbookGridControls
            contract={timelineContract}
            filterDraft={filterDraft}
            onApplyFilter={applyQueryFilter}
            onFilterDraftChange={setFilterDraft}
            onGroupByChange={handleQueryGroupByChange}
            onRemoveFilter={(fieldKey) => {
              setQueryState((current) => removeFilterField(current, fieldKey));
            }}
            queryState={queryState}
            surface="timeline"
          />
          <GridViewport
            ref={gridShellRef}
            style={gridShellStyle}
            testId={gridShellTestId("timeline")}
          >
            <GridTable
              actionsColumn={timelineActionsColumn}
              columns={timelineColumns}
              getGroupLabel={getTimelineGroupLabel}
              getGroupRowTestId={getTimelineGroupRowTestId}
              groupBy={queryState.groupBy}
              onToggleSort={handleQuerySortToggle}
              rows={timelineGridRows}
              sort={queryState.sort}
            />
          </GridViewport>
          {activeConflict ? (
            <section
              data-testid="conflict-resolver"
              style={conflictResolverStyle}
              aria-label="Same-field conflict resolver"
            >
              <div data-testid="conflict-resolver-summary" tabIndex={-1}>
                <p style={eyebrowStyle}>Conflict</p>
                <h2 style={sectionTitleStyle}>
                  {timelineBindingLabel(activeConflict.conflict.field_key)}
                </h2>
                <p style={bodyStyle}>
                  This field changed before your edit was saved. Review the
                  saved value and your unsaved value.
                </p>
              </div>
              <div style={conflictResolverGridStyle}>
                <label style={labelStyle}>
                  Field key
                  <input
                    readOnly
                    style={inputStyle}
                    data-testid="conflict-field-key"
                    value={activeConflict.conflict.field_key}
                  />
                </label>
                <label style={labelStyle}>
                  Saved by
                  <input
                    readOnly
                    style={inputStyle}
                    data-testid="conflict-server-actor"
                    value={activeConflict.conflict.server_updated_by ?? ""}
                  />
                </label>
                <label style={labelStyle}>
                  Saved at
                  <input
                    readOnly
                    style={inputStyle}
                    data-testid="conflict-server-updated-at"
                    value={activeConflict.conflict.server_updated_at ?? ""}
                  />
                </label>
              </div>
              <div style={conflictResolverGridStyle}>
                <label style={labelStyle}>
                  Saved value
                  <textarea
                    readOnly
                    style={textareaStyle}
                    data-testid="conflict-server-value"
                    value={String(activeConflict.conflict.server_value ?? "")}
                  />
                </label>
                <label style={labelStyle}>
                  Your unsaved value
                  <textarea
                    readOnly
                    style={textareaStyle}
                    data-testid="conflict-local-value"
                    value={String(activeConflict.localValue ?? "")}
                  />
                </label>
                {activeConflict.conflict.conflict_resolution_class ===
                "text_compare_merge" ? (
                  <label style={labelStyle}>
                    Merged value
                    <textarea
                      style={textareaStyle}
                      data-testid="conflict-merged-value"
                      value={activeConflict.mergedDraft}
                      onChange={(event) => {
                        const value = event.currentTarget.value;
                        setConflictQueueState((current) => ({
                          ...current,
                          [activeConflict.key]: {
                            ...activeConflict,
                            mergedDraft: value,
                          },
                        }));
                      }}
                    />
                  </label>
                ) : null}
              </div>
              <div style={inlineButtonRowStyle}>
                <button
                  type="button"
                  style={secondaryActionButtonStyle}
                  data-testid="conflict-close"
                  onClick={() => {
                    setActiveConflictKey(null);
                    restoreConflictFocus(activeConflict.focusKey);
                  }}
                >
                  Close
                </button>
                <button
                  type="button"
                  style={secondaryActionButtonStyle}
                  data-testid="conflict-keep-saved"
                  onClick={() =>
                    submitConflictResolution(activeConflict, "keep_saved")
                  }
                >
                  Keep saved value
                </button>
                <button
                  type="button"
                  style={actionButtonStyle}
                  data-testid="conflict-use-unsaved"
                  onClick={() =>
                    submitConflictResolution(activeConflict, "use_unsaved")
                  }
                >
                  Use my unsaved value
                </button>
                {activeConflict.conflict.conflict_resolution_class ===
                "text_compare_merge" ? (
                  <button
                    type="button"
                    style={actionButtonStyle}
                    data-testid="conflict-use-merged"
                    onClick={() =>
                      submitConflictResolution(activeConflict, "merged_value")
                    }
                  >
                    Use merged value
                  </button>
                ) : null}
              </div>
            </section>
          ) : null}
        </div>

        <aside data-testid="timeline-inspector" style={inspectorShellStyle}>
          <div style={inspectorHeaderStyle}>
            <p style={eyebrowStyle}>Inspector</p>
            <h2 style={inspectorTitleStyle}>
              {selectedRow?.recordId
                ? `Timeline row ${selectedRow.recordId}`
                : currentHistoryDeleted && rowHistory.data?.record_id
                  ? `Deleted row ${rowHistory.data.record_id}`
                  : draftRow
                    ? "Draft timeline row"
                    : "Select a saved row"}
            </h2>
            <p style={bodyStyle}>
              Routine mention review and hidden-field editing stay on the
              workbook surface.
            </p>
          </div>
          {selectedRow?.recordId ? (
            <>
              {renderInspectorFieldEditors(selectedRow)}
              {renderEvidenceAttachSection(selectedRow)}
              {renderRowHistorySection()}
              <section style={inspectorSectionStyle}>
                <h3 style={sectionTitleStyle}>Mentions</h3>
                <div style={mentionGroupStyle}>
                  {["unresolved", "resolved", "dismissed"].map((status) => {
                    const group = inspectorMentions.filter(
                      (item) => item.status === status,
                    );
                    return (
                      <div key={status} style={mentionGroupColumnStyle}>
                        <p style={groupLabelStyle}>
                          {status === "dismissed"
                            ? "Dismissed"
                            : status === "resolved"
                              ? "Resolved"
                              : "Unresolved"}
                        </p>
                        {group.length > 0 ? (
                          group.map((item) => (
                            <button
                              key={item.itemRef}
                              data-testid={`mention-${sanitizeTestId(item.itemRef)}`}
                              style={{
                                ...mentionListButtonStyle,
                                ...(selectedMention?.itemRef === item.itemRef
                                  ? mentionListButtonSelectedStyle
                                  : null),
                              }}
                              type="button"
                              onClick={() => {
                                handleSelectMention(
                                  item.rowRecordId,
                                  item.itemRef,
                                );
                              }}
                            >
                              <RelationshipChip
                                entityIndex={entityIndex}
                                item={item}
                                selected={
                                  selectedMention?.itemRef === item.itemRef
                                }
                              />
                            </button>
                          ))
                        ) : (
                          <span style={emptyRelationshipStyle}>None</span>
                        )}
                      </div>
                    );
                  })}
                </div>
              </section>

              {selectedMention ? (
                <section style={inspectorSectionStyle}>
                  <h3 style={sectionTitleStyle}>Selected mention</h3>
                  <dl style={detailListStyle}>
                    <div>
                      <dt style={detailTermStyle}>Raw token</dt>
                      <dd style={detailValueStyle}>
                        {selectedMention.rawText}
                      </dd>
                    </div>
                    <div>
                      <dt style={detailTermStyle}>Field</dt>
                      <dd style={detailValueStyle}>
                        {timelineRelationshipLabel(selectedMention.fieldKey)}
                      </dd>
                    </div>
                    <div>
                      <dt style={detailTermStyle}>Status</dt>
                      <dd style={detailValueStyle}>{selectedMention.status}</dd>
                    </div>
                    <div>
                      <dt style={detailTermStyle}>Target</dt>
                      <dd style={detailValueStyle}>
                        {selectedMention.resolvedRecordId
                          ? relationshipItemLabel(selectedMention, entityIndex)
                          : "None"}
                      </dd>
                    </div>
                  </dl>

                  {selectedMention.status === "unresolved" ? (
                    <div style={inspectorActionStackStyle}>
                      <label style={labelStyle}>
                        Resolve to existing
                        <select
                          data-testid="inspector-resolve-target"
                          style={selectStyle}
                          value={selectedResolveTargetId}
                          onChange={(event) => {
                            const value = event.target.value;
                            setSelectedResolveTargetId(value);
                            if (value !== "") {
                              setInspectorMessage(`Selected ${value}`);
                            }
                          }}
                        >
                          <option value="">Select target</option>
                          {(selectedMention.entityType === "host"
                            ? hostEntities
                            : identityEntities
                          ).map((entity) => (
                            <option
                              key={entity.recordId}
                              value={entity.recordId}
                            >
                              {entity.label}
                            </option>
                          ))}
                        </select>
                      </label>
                      <div style={inlineButtonRowStyle}>
                        <button
                          style={secondaryActionButtonStyle}
                          type="button"
                          onClick={() => {
                            if (selectedResolveTargetId === "") {
                              setInspectorMessage("Select a target first.");
                              return;
                            }
                            submitMentionAction(
                              selectedMention,
                              "resolve_item",
                              selectedResolveTargetId,
                            );
                          }}
                        >
                          Resolve to existing
                        </button>
                        <button
                          style={secondaryActionButtonStyle}
                          type="button"
                          onClick={() => {
                            submitMentionAction(
                              selectedMention,
                              "resolve_item",
                            );
                          }}
                        >
                          {selectedMention.entityType === "host"
                            ? "Create host"
                            : "Create identity"}
                        </button>
                      </div>
                    </div>
                  ) : null}

                  {selectedMention.status === "resolved" ? (
                    <div style={inlineButtonRowStyle}>
                      {canManageMentions ? (
                        <button
                          style={secondaryActionButtonStyle}
                          type="button"
                          onClick={() => {
                            submitMentionAction(
                              selectedMention,
                              "dismiss_item",
                            );
                          }}
                        >
                          Dismiss
                        </button>
                      ) : null}
                      {canManageMentions ? (
                        <button
                          style={secondaryActionButtonStyle}
                          type="button"
                          onClick={() => {
                            submitMentionAction(
                              selectedMention,
                              "revert_to_unresolved",
                            );
                          }}
                        >
                          Revert to unresolved
                        </button>
                      ) : null}
                    </div>
                  ) : null}

                  {selectedMention.status === "dismissed" ? (
                    <div style={inlineButtonRowStyle}>
                      <button
                        style={secondaryActionButtonStyle}
                        type="button"
                        onClick={() => {
                          submitMentionAction(
                            selectedMention,
                            "revert_to_unresolved",
                          );
                        }}
                      >
                        Restore to unresolved
                      </button>
                    </div>
                  ) : null}
                </section>
              ) : null}
              {inspectorMessage ? (
                <p data-testid="timeline-inspector-message" style={bodyStyle}>
                  {inspectorMessage}
                </p>
              ) : null}
            </>
          ) : currentHistoryDeleted && rowHistory.data !== null ? (
            <>
              {renderRowHistorySection()}
              {inspectorMessage ? (
                <p data-testid="timeline-inspector-message" style={bodyStyle}>
                  {inspectorMessage}
                </p>
              ) : null}
            </>
          ) : (
            <>
              {draftRow ? renderInspectorFieldEditors(draftRow) : null}
              {draftRow ? renderEvidenceAttachSection(draftRow) : null}
              <p style={bodyStyle}>
                Pick a saved row to inspect unresolved, resolved, and dismissed
                mentions.
              </p>
              {inspectorMessage ? (
                <p data-testid="timeline-inspector-message" style={bodyStyle}>
                  {inspectorMessage}
                </p>
              ) : null}
            </>
          )}
        </aside>
      </div>
    </section>
  );
}

function EntityWorkbookSurface({
  incidentId,
  apiBase,
  entityType,
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
  const [timelinePreviewRows, setTimelinePreviewRows] = useState<WorkbookRow[]>(
    [],
  );

  const selectedEntity =
    rows.find((row) => row.recordId === selectedRecordId) ?? rows[0] ?? null;
  const canMerge =
    currentIncidentRole === "reviewer" || currentIncidentRole === "admin";
  const survivorLabel = selectedEntity?.label ?? "Select a record";
  const surface: WorkbookSurface =
    entityType === "host" ? "hosts" : "identities";
  const contract = entityType === "host" ? hostsContract : identitiesContract;
  const loserEntity =
    rows.find((row) => row.recordId === mergeCandidateId) ?? null;
  const mergePlan =
    selectedEntity && loserEntity
      ? buildMergePlan(selectedEntity, loserEntity)
      : null;
  const entityColumns: readonly GridColumn<EntityRow>[] = [
    {
      fieldKey:
        entityType === "host" ? "host.display_name" : "identity.display_name",
      headerTestId: gridSortHeaderTestId(
        surface,
        entityType === "host" ? "host.display_name" : "identity.display_name",
      ),
      label:
        contract.fieldMap[
          entityType === "host" ? "host.display_name" : "identity.display_name"
        ]?.label ?? "Name",
      width: 240,
      renderCell: (row) => row.label,
      sortableFieldKey:
        entityType === "host" ? "host.display_name" : "identity.display_name",
    },
    {
      fieldKey: entityType === "host" ? "host.hostname" : "identity.upn",
      headerTestId: gridSortHeaderTestId(
        surface,
        entityType === "host" ? "host.hostname" : "identity.upn",
      ),
      label:
        contract.fieldMap[
          entityType === "host" ? "host.hostname" : "identity.upn"
        ]?.label ?? "Primary",
      width: 260,
      renderCell: (row) => row.secondaryText || "None",
      sortableFieldKey:
        entityType === "host" ? "host.hostname" : "identity.upn",
    },
    {
      fieldKey: entityType === "host" ? "host.aliases" : "identity.aliases",
      label:
        contract.fieldMap[
          entityType === "host" ? "host.aliases" : "identity.aliases"
        ]?.label ?? "Aliases",
      width: 320,
      renderCell: (row) => (
        <div style={relationshipItemsWrapStyle}>
          {row.aliasTexts.length > 0 ? (
            row.aliasTexts.map((alias) => (
              <span key={alias} style={aliasChipStyle}>
                {alias}
              </span>
            ))
          ) : (
            <span style={emptyRelationshipStyle}>No aliases</span>
          )}
        </div>
      ),
    },
    {
      fieldKey:
        entityType === "host" ? "host.host_state" : "identity.identity_state",
      headerTestId: gridSortHeaderTestId(
        surface,
        entityType === "host" ? "host.host_state" : "identity.identity_state",
      ),
      label:
        contract.fieldMap[
          entityType === "host" ? "host.host_state" : "identity.identity_state"
        ]?.label ?? "State",
      width: 140,
      renderCell: (row) => row.state,
      sortableFieldKey:
        entityType === "host" ? "host.host_state" : "identity.identity_state",
    },
    {
      fieldKey: "row_version",
      label: "Version",
      width: 96,
      renderCell: (row) => row.rowVersion,
    },
  ];
  const entityActionsColumn: GridActionsColumn<EntityRow> = {
    headerTestId: gridActionsHeaderTestId(surface),
    label: "Actions",
    width: 176,
    renderCell: ({ data: row }) => (
      <button
        data-testid={`inspect-${entityType}-${row.recordId}`}
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
  const entityGridRows: readonly GridRow<EntityRow>[] = rows.map((row) => ({
    key: row.recordId,
    recordId: row.recordId,
    data: row,
    selected: row.recordId === selectedEntity?.recordId,
    testId: `${entityType}-row-${row.recordId}`,
  }));

  useEffect(() => {
    if (selectedEntity) {
      setSelectedRecordId(selectedEntity.recordId);
      return;
    }
    setSelectedRecordId(null);
  }, [selectedEntity]);

  const loadTimelinePreview = useCallback(
    async (recordId: string) => {
      const result = await fetchJSON<WorkbookQueryEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`,
        ),
        {
          method: "POST",
          body: JSON.stringify({}),
        },
      );
      if (!result.ok) {
        setTimelinePreviewRows([]);
        return;
      }
      const envelope = readEnvelope<WorkbookQueryEnvelope>(result.payload);
      const draftKey = entityType === "host" ? "hostRefs" : "identityRefs";
      const previewRows = envelope.data.rows
        .map(rowFromApi)
        .filter((row) =>
          row.collectionValues[draftKey].some(
            (item) => item.resolvedRecordId === recordId,
          ),
        );
      setTimelinePreviewRows(previewRows);
    },
    [apiBase, entityType, incidentId],
  );

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
    <section style={workbookStyle}>
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
      </header>

      <div style={splitShellStyle}>
        <div>
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
          data-testid={`${entityType}-inspector`}
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
                        data-testid={`timeline-preview-row-${row.recordId ?? row.key}`}
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
    initialAssessmentDraft(),
  );
  const [supportRows, setSupportRows] = useState<TimelineApiRow[]>([]);
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
  const columns: readonly GridColumn<EntityApiRow>[] = visibleFields(
    assessmentsContract,
  ).map((field) => ({
    fieldKey: field.fieldKey,
    headerTestId: gridSortHeaderTestId(assessmentsViewSchemaId, field.fieldKey),
    label: field.label,
    width: assessmentColumnWidth(field.fieldKey),
    renderCell: (row) => (
      <span data-testid={rowCellTestId(row.record_id, field.fieldKey)}>
        {genericCellLabel(row.cells[field.fieldKey]?.value)}
      </span>
    ),
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
      testId: `assessment-row-${row.record_id}`,
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

  useEffect(() => {
    let isCurrent = true;
    async function loadSupportRows() {
      const result = await fetchJSON<WorkbookQueryEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`,
        ),
        {
          method: "POST",
          body: JSON.stringify({}),
        },
      );
      if (!isCurrent) {
        return;
      }
      if (!result.ok) {
        setSupportRows([]);
        return;
      }
      const envelope = readEnvelope<WorkbookQueryEnvelope>(result.payload);
      setSupportRows(envelope.data.rows);
    }
    void loadSupportRows();
    return () => {
      isCurrent = false;
    };
  }, [apiBase, incidentId]);

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
        ...initialAssessmentDraft(),
        subjectType: current.subjectType,
        subjectRecordId: current.subjectRecordId,
      }));
      setMessage("Assessment created.");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <section style={workbookStyle}>
      <header style={headerStyle}>
        <div>
          <p style={eyebrowStyle}>System view</p>
          <h1 style={headlineStyle}>{assessmentsContract.title}</h1>
          <p style={bodyStyle}>Incident {incidentId}</p>
        </div>
      </header>

      <div style={splitShellStyle}>
        <div>
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
          data-testid="assessment-create-panel"
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
  const [editCollectionMode, setEditCollectionMode] =
    useState<GenericCollectionMode>("add");
  const [referenceOptions, setReferenceOptions] =
    useState<GenericReferenceOptions>(() => emptyGenericReferenceOptions());
  const [referenceLoadError, setReferenceLoadError] = useState<string | null>(
    null,
  );
  const [mutationError, setMutationError] = useState<string | null>(null);
  const [mutationState, setMutationState] = useState<SaveState>("Saved");
  const [evidenceMessageByRecordID, setEvidenceMessageByRecordID] = useState<
    Record<string, string>
  >({});
  const [evidencePreview, setEvidencePreview] =
    useState<EvidencePreviewState | null>(null);
  const isEvidenceSurface = contract.viewSchemaId === evidenceViewSchemaId;

  useEffect(() => {
    setCreateDraft((current) => {
      const defaults = initialGenericCreateDraft(contract, currentUserId);
      return { ...defaults, ...current };
    });
  }, [contract, currentUserId]);

  const refreshReferenceOptions = useCallback(async () => {
    setReferenceLoadError(null);
    const targetViewSchemaIds = [
      timelineViewSchemaId,
      partiesViewSchemaId,
      taskRequestsViewSchemaId,
      decisionsViewSchemaId,
      evidenceViewSchemaId,
      notesViewSchemaId,
    ];
    const loaded = await Promise.all(
      targetViewSchemaIds.map(async (viewSchemaId) => {
        const targetContract = requireViewContract(viewSchemaId);
        const result = await fetchJSON<ViewQueryEnvelope>(
          apiPath(
            apiBase,
            `/api/v1/incidents/${incidentId}/views/${viewSchemaId}/query`,
          ),
          {
            method: "POST",
            body: JSON.stringify(
              buildQueryRequest(targetContract, emptyWorkbookQueryState()),
            ),
          },
        );
        if (!result.ok) {
          throw new Error(parseErrorMessage(result.payload));
        }
        const envelope = readEnvelope<ViewQueryEnvelope>(result.payload);
        return [viewSchemaId, envelope.data.rows] as const;
      }),
    );
    const rowsByView = Object.fromEntries(loaded) as Record<
      string,
      EntityApiRow[]
    >;
    const next: GenericReferenceOptions = {
      parties: genericReferenceOptionsFromRows(
        partiesViewSchemaId,
        rowsByView[partiesViewSchemaId] ?? [],
      ),
      taskRequests: genericReferenceOptionsFromRows(
        taskRequestsViewSchemaId,
        rowsByView[taskRequestsViewSchemaId] ?? [],
      ),
      decisions: genericReferenceOptionsFromRows(
        decisionsViewSchemaId,
        rowsByView[decisionsViewSchemaId] ?? [],
      ),
      evidence: genericReferenceOptionsFromRows(
        evidenceViewSchemaId,
        rowsByView[evidenceViewSchemaId] ?? [],
      ),
      notes: genericReferenceOptionsFromRows(
        notesViewSchemaId,
        rowsByView[notesViewSchemaId] ?? [],
      ),
      timeline: genericReferenceOptionsFromRows(
        timelineViewSchemaId,
        rowsByView[timelineViewSchemaId] ?? [],
      ),
      allRecords: [],
    };
    next.allRecords = [
      ...next.timeline,
      ...next.evidence,
      ...next.notes,
      ...next.taskRequests,
      ...next.decisions,
      ...next.parties,
    ];
    setReferenceOptions(next);
  }, [apiBase, incidentId]);

  useEffect(() => {
    let isCurrent = true;
    refreshReferenceOptions().catch((error: unknown) => {
      if (!isCurrent) {
        return;
      }
      setReferenceOptions(emptyGenericReferenceOptions());
      setReferenceLoadError(
        error instanceof Error ? error.message : "Reference options failed.",
      );
    });
    return () => {
      isCurrent = false;
    };
  }, [refreshReferenceOptions]);

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
      const result = await fetchJSON<EvidenceHandleEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/evidence-records/${row.record_id}/${kind}-handle`,
        ),
        { method: "POST", body: JSON.stringify({}) },
      );
      if (!result.ok) {
        setEvidenceMessage(row.record_id, parseErrorMessage(result.payload));
        return;
      }
      const envelope = readEnvelope<EvidenceHandleEnvelope>(result.payload);
      const href =
        envelope.data.href.startsWith("/") && apiBase
          ? apiPath(apiBase, envelope.data.href)
          : envelope.data.href;
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
        const createBlob = await fetchJSON<ObjectBlobCreateEnvelope>(
          apiPath(apiBase, "/api/v1/object-blobs"),
          {
            method: "POST",
            body: JSON.stringify({
              incident_id: incidentId,
              client_txn_id: `evidence-blob-${Date.now()}`,
              byte_size: file.size,
              filename_hint: file.name || null,
              content_type_hint: file.type || null,
            }),
          },
        );
        if (!createBlob.ok) {
          throw new Error(parseErrorMessage(createBlob.payload));
        }
        const blobEnvelope = readEnvelope<ObjectBlobCreateEnvelope>(
          createBlob.payload,
        );
        const uploadHref =
          blobEnvelope.data.upload_target.href.startsWith("/") && apiBase
            ? apiPath(apiBase, blobEnvelope.data.upload_target.href)
            : blobEnvelope.data.upload_target.href;
        const upload = await fetch(uploadHref, {
          method: blobEnvelope.data.upload_target.method ?? "PUT",
          credentials: "include",
          headers: {
            "Content-Type": file.type || "application/octet-stream",
          },
          body: file,
        });
        if (!upload.ok) {
          throw new Error(`upload_failed_${upload.status}`);
        }
        const attach = await fetchJSON<EvidenceAttachBlobEnvelope>(
          apiPath(
            apiBase,
            `/api/v1/evidence-records/${row.record_id}/attach-blob`,
          ),
          {
            method: "POST",
            body: JSON.stringify({
              object_blob_id: blobEnvelope.data.object_blob_id,
              base_row_version: row.row_version,
              client_txn_id: `evidence-attach-${Date.now()}`,
            }),
          },
        );
        if (!attach.ok) {
          throw new Error(parseErrorMessage(attach.payload));
        }
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

  const columns: readonly GridColumn<EntityApiRow>[] = visibleFields(
    contract,
  ).map((field) => ({
    fieldKey: field.fieldKey,
    headerTestId: gridSortHeaderTestId(surface, field.fieldKey),
    label: field.label,
    width: field.defaultHidden ? 160 : 220,
    renderCell: (row) => (
      <span data-testid={rowCellTestId(row.record_id, field.fieldKey)}>
        {genericCellLabel(row.cells[field.fieldKey]?.value)}
      </span>
    ),
    sortableFieldKey: resolveHeaderSortFieldKey(contract, field.fieldKey),
  }));
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
        const uploadState = stringifyGridValue(
          row.cells["evidence.upload_state"]?.value,
        );
        const lifecycleState = stringifyGridValue(
          row.cells["evidence.lifecycle_state"]?.value,
        );
        const canAccess =
          uploadState === "available" &&
          (lifecycleState === "available" || lifecycleState === "released");
        const message =
          evidenceMessageByRecordID[row.record_id] ??
          (canAccess ? null : `Blocked: ${uploadState || "no blob"}`);
        return (
          <div style={actionStackStyle}>
            <div style={inlineButtonRowStyle}>
              <button
                data-testid={`evidence-preview-${row.record_id}`}
                disabled={!canAccess}
                style={actionButtonStyle}
                type="button"
                onClick={() => {
                  void issueEvidenceHandle(row, "preview");
                }}
              >
                Preview
              </button>
              <button
                data-testid={`evidence-download-${row.record_id}`}
                disabled={!canAccess}
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
                data-testid={`evidence-attach-file-${row.record_id}`}
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
                data-testid={`evidence-access-message-${row.record_id}`}
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
  const gridRows: readonly GridRow<EntityApiRow>[] = rows.map((row) => ({
    key: row.record_id,
    recordId: row.record_id,
    data: row,
    testId: `generic-row-${contract.viewSchemaId}-${row.record_id}`,
  }));
  const selectedEditRow =
    rows.find((row) => row.record_id === editRecordId) ?? null;
  const selectedEditField =
    writableFields.find((field) => field.fieldKey === editFieldKey) ??
    writableFields[0] ??
    null;
  const selectedEditCollectionItems =
    selectedEditRow !== null && selectedEditField !== null
      ? genericCollectionItems(selectedEditRow, selectedEditField.fieldKey)
      : [];

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
    setCreateDraft(initialGenericCreateDraft(contract, currentUserId));
    setMutationState("Saved");
    await onRefresh();
    await refreshReferenceOptions();
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
    setMutationState("Syncing");
    setMutationError(null);
    const result = await fetchJSON<ViewMutationEnvelope>(
      apiPath(apiBase, `/api/v1/records/${selectedEditRow.record_id}`),
      {
        method: "PATCH",
        body: JSON.stringify({
          view_schema_id: contract.viewSchemaId,
          base_row_version: selectedEditRow.row_version,
          client_txn_id: `generic-patch-${contract.viewSchemaId}-${Date.now()}`,
          changes: [change],
        }),
      },
    );
    if (!result.ok) {
      setMutationState("Conflict");
      setMutationError(parseMutationError(result.payload));
      return;
    }
    setEditValue("");
    setMutationState("Saved");
    await onRefresh();
    await refreshReferenceOptions();
  };

  return (
    <section style={workbookStyle}>
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
                  testId={`generic-create-field-${field.fieldKey}`}
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
          </div>
          <button
            data-testid={`generic-create-submit-${contract.viewSchemaId}`}
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
                data-testid={`generic-edit-record-${contract.viewSchemaId}`}
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
                data-testid={`generic-edit-field-${contract.viewSchemaId}`}
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
                  data-testid={`generic-edit-action-${contract.viewSchemaId}`}
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
                testId={`generic-edit-value-${contract.viewSchemaId}`}
                value={editValue}
                onChange={setEditValue}
              />
              <button
                data-testid={`generic-edit-submit-${contract.viewSchemaId}`}
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

      {loadError ? (
        <p data-testid="generic-surface-load-error" style={bodyStyle}>
          {loadError}
        </p>
      ) : null}

      {isEvidenceSurface && evidencePreview ? (
        <section
          data-testid="evidence-preview-panel"
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
            data-testid={`evidence-preview-frame-${evidencePreview.recordId}`}
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
    </section>
  );
}

function GenericMutationControl({
  collectionItems = [],
  collectionMode,
  field,
  id,
  referenceOptions,
  testId,
  value,
  onChange,
}: {
  collectionItems?: Array<{ itemRef: string; displayText: string }>;
  collectionMode: GenericCollectionMode;
  field: ViewFieldContract;
  id?: string;
  referenceOptions: GenericReferenceOptions;
  testId: string;
  value: string;
  onChange: (value: string) => void;
}) {
  if (field.writeKind === "action_payload") {
    if (collectionMode === "remove") {
      return (
        <select
          data-testid={testId}
          id={id}
          multiple
          size={Math.min(Math.max(collectionItems.length, 2), 6)}
          style={selectStyle}
          value={splitDraftValues(value)}
          onChange={(event) => {
            onChange(
              Array.from(event.currentTarget.selectedOptions)
                .map((option) => option.value)
                .join("\n"),
            );
          }}
        >
          {collectionItems.map((item) => (
            <option key={item.itemRef} value={item.itemRef}>
              {item.displayText}
            </option>
          ))}
        </select>
      );
    }

    const options = referenceOptionsForField(field, referenceOptions);
    if (genericFieldUsesReferenceOptions(field)) {
      return (
        <select
          data-testid={testId}
          id={id}
          multiple
          size={Math.min(Math.max(options.length, 2), 6)}
          style={selectStyle}
          value={splitDraftValues(value)}
          onChange={(event) => {
            onChange(
              Array.from(event.currentTarget.selectedOptions)
                .map((option) => option.value)
                .join("\n"),
            );
          }}
        >
          {options.map((option) => (
            <option key={option.recordId} value={option.recordId}>
              {option.label}
            </option>
          ))}
        </select>
      );
    }

    return (
      <textarea
        data-testid={testId}
        id={id}
        rows={3}
        style={textareaStyle}
        value={value}
        onChange={(event) => {
          onChange(event.target.value);
        }}
      />
    );
  }

  const referenceChoices = referenceOptionsForField(field, referenceOptions);
  if (genericFieldUsesReferenceOptions(field)) {
    return (
      <select
        data-testid={testId}
        id={id}
        style={selectStyle}
        value={value}
        onChange={(event) => {
          onChange(event.target.value);
        }}
      >
        <option value="">{field.clearable ? "None" : "Select"}</option>
        {referenceChoices.map((option) => (
          <option key={option.recordId} value={option.recordId}>
            {option.label}
          </option>
        ))}
      </select>
    );
  }

  if (field.enumValues && field.enumValues.length > 0) {
    return (
      <select
        data-testid={testId}
        id={id}
        style={selectStyle}
        value={value}
        onChange={(event) => {
          onChange(event.target.value);
        }}
      >
        <option value="">Select</option>
        {field.enumValues.map((option) => (
          <option key={option} value={option}>
            {option}
          </option>
        ))}
      </select>
    );
  }

  if (isMultilineGenericField(field)) {
    return (
      <textarea
        data-testid={testId}
        id={id}
        rows={3}
        style={textareaStyle}
        value={value}
        onChange={(event) => {
          onChange(event.target.value);
        }}
      />
    );
  }

  return (
    <input
      data-testid={testId}
      id={id}
      placeholder={
        field.directScalarContractId === "timestamp_instant_v1"
          ? "RFC3339 timestamp"
          : undefined
      }
      style={inputStyle}
      type="text"
      value={value}
      onChange={(event) => {
        onChange(event.target.value);
      }}
    />
  );
}

function genericCellLabel(value: unknown): string {
  if (value === null || value === undefined || value === "") {
    return "None";
  }
  if (typeof value === "string" || typeof value === "number") {
    return String(value);
  }
  if (typeof value === "boolean") {
    return value ? "Yes" : "No";
  }
  if (typeof value === "object" && value !== null && "items" in value) {
    const items = (value as { items?: unknown }).items;
    if (Array.isArray(items)) {
      return `${items.length} item${items.length === 1 ? "" : "s"}`;
    }
  }
  return JSON.stringify(value);
}

export function buildGenericCreatePayload(
  contract: ViewContract,
  draft: Record<string, string>,
  clientTxnId: string,
): Record<string, unknown> | null {
  if (!workbookCreateMinimumSatisfied(contract.viewSchemaId, draft)) {
    return null;
  }
  const payload: Record<string, unknown> = { client_txn_id: clientTxnId };
  const fields = contract.fields.filter(
    (field) => field.writeKind !== "read_only",
  );
  for (const field of fields) {
    const value = normalizeValue(draft[field.fieldKey] ?? "");
    if (field.writeKind === "action_payload") {
      if (value === "") {
        continue;
      }
      const actionPayload = buildGenericCollectionActions(field, value, "add");
      if (actionPayload !== null) {
        payload[field.fieldKey] = actionPayload;
      }
      continue;
    }

    if (value === "") {
      if (field.clearable) {
        payload[field.fieldKey] = null;
      }
      continue;
    }
    payload[field.fieldKey] = value;
  }
  return Object.keys(payload).length > 1 ? payload : null;
}

export function buildGenericPatchChange(
  field: ViewFieldContract,
  rawValue: string,
  collectionMode: GenericCollectionMode = "add",
): Record<string, unknown> | null {
  const value = normalizeValue(rawValue);
  if (field.writeKind === "action_payload") {
    const actionPayload = buildGenericCollectionActions(
      field,
      value,
      collectionMode,
    );
    return actionPayload === null
      ? null
      : { field_key: field.fieldKey, action_payload: actionPayload };
  }
  if (value === "" && !field.clearable) {
    return null;
  }
  return {
    field_key: field.fieldKey,
    value: value === "" && field.clearable ? null : value,
  };
}

function buildGenericCollectionActions(
  field: ViewFieldContract,
  rawValue: string,
  mode: GenericCollectionMode,
): Record<string, unknown> | null {
  const tokens = splitDraftValues(rawValue);
  if (tokens.length === 0) {
    return null;
  }
  const actions = tokens.map((value) => {
    if (field.fieldKey === "note.tags") {
      return { op: "add_token", raw_text: value };
    }
    if (isPartyRefCollection(field.fieldKey)) {
      return mode === "remove"
        ? { op: "remove_party_ref", item_ref: value }
        : { op: "add_party_ref", party_id: value };
    }
    if (field.fieldKey === "handoff.open_risk_refs") {
      return mode === "remove"
        ? { op: "remove_risk_ref", item_ref: value }
        : { op: "add_risk_ref", risk_ref_text: value };
    }
    return mode === "remove"
      ? { op: "remove_record_ref", item_ref: value }
      : { op: "add_record_ref", linked_record_id: value };
  });
  return { kind: "collection_actions_v1", actions };
}

function splitDraftValues(rawValue: string): string[] {
  return rawValue
    .split(/\r?\n/u)
    .map((value) => normalizeValue(value))
    .filter((value) => value !== "");
}

function workbookCreateMinimumSatisfied(
  viewSchemaId: string,
  draft: Record<string, string>,
): boolean {
  const has = (fieldKey: string) =>
    normalizeValue(draft[fieldKey] ?? "") !== "";
  switch (viewSchemaId) {
    case partiesViewSchemaId:
      return has("party.display_name") && has("party.party_kind");
    case notesViewSchemaId:
      return has("note.title") || has("note.body");
    case taskRequestsViewSchemaId:
      return has("task.title") && has("task.task_kind");
    case decisionsViewSchemaId:
      return (
        has("decision.summary") &&
        has("decision.decision_type") &&
        has("decision.rationale")
      );
    case evidenceViewSchemaId:
      return (
        has("evidence.title") ||
        has("evidence.storage_ref") ||
        has("evidence.collector_party_text") ||
        has("evidence.source_party_text")
      );
    case commLogViewSchemaId:
      return (
        has("comm_log.comm_type") &&
        has("comm_log.audience") &&
        has("comm_log.channel_or_meeting") &&
        has("comm_log.summary")
      );
    case handoffViewSchemaId:
      return (
        has("handoff.incoming_owner_user_id") &&
        has("handoff.current_state_summary")
      );
    case statusReviewViewSchemaId:
      return has("status_review.current_state_summary");
    case lessonViewSchemaId:
      return has("lesson.summary");
    default:
      return Object.values(draft).some((value) => normalizeValue(value) !== "");
  }
}

function genericCreateMinimumMessage(viewSchemaId: string): string {
  switch (viewSchemaId) {
    case partiesViewSchemaId:
      return "Display name and kind are required.";
    case notesViewSchemaId:
      return "Title or body is required.";
    case taskRequestsViewSchemaId:
      return "Title and task kind are required.";
    case decisionsViewSchemaId:
      return "Summary, decision type, and rationale are required.";
    case evidenceViewSchemaId:
      return "Evidence needs a title, storage ref, collector, or source.";
    case commLogViewSchemaId:
      return "Type, audience, channel or meeting, and summary are required.";
    case handoffViewSchemaId:
      return "Incoming owner and current state summary are required.";
    case statusReviewViewSchemaId:
      return "Current state summary is required.";
    case lessonViewSchemaId:
      return "Summary is required.";
    default:
      return "At least one value is required.";
  }
}

function initialGenericCreateDraft(
  contract: ViewContract,
  currentUserId: string | null,
): Record<string, string> {
  const draft: Record<string, string> = {};
  for (const field of contract.fields) {
    if (field.writeKind === "read_only") {
      continue;
    }
    if (
      currentUserId &&
      (field.fieldKey === "handoff.incoming_owner_user_id" ||
        field.fieldKey === "task.owner_user_id" ||
        field.fieldKey === "decision.owner_user_id" ||
        field.fieldKey === "status_review.review_owner_user_id" ||
        field.fieldKey === "lesson.owner_user_id")
    ) {
      draft[field.fieldKey] = currentUserId;
    }
  }
  return draft;
}

function emptyGenericReferenceOptions(): GenericReferenceOptions {
  return {
    parties: [],
    taskRequests: [],
    decisions: [],
    evidence: [],
    notes: [],
    timeline: [],
    allRecords: [],
  };
}

function genericReferenceOptionsFromRows(
  viewSchemaId: string,
  rows: EntityApiRow[],
): GenericReferenceOption[] {
  return rows.map((row) => ({
    recordId: row.record_id,
    label: genericRowLabel(requireViewContract(viewSchemaId), row),
    viewSchemaId,
  }));
}

function genericRowLabel(contract: ViewContract, row: EntityApiRow): string {
  const preferredFieldKeys = [
    "timeline.summary",
    "party.display_name",
    "task.title",
    "decision.summary",
    "evidence.title",
    "evidence.storage_ref",
    "note.title",
    "note.body",
    "comm_log.summary",
    "handoff.current_state_summary",
    "status_review.current_state_summary",
    "lesson.summary",
  ];
  for (const fieldKey of preferredFieldKeys) {
    if (!contract.fieldMap[fieldKey]) {
      continue;
    }
    const label = stringifyGridValue(row.cells[fieldKey]?.value).trim();
    if (label !== "") {
      return `${label} (${row.record_id})`;
    }
  }
  return row.record_id;
}

function referenceOptionsForField(
  field: ViewFieldContract,
  options: GenericReferenceOptions,
): GenericReferenceOption[] {
  if (field.directReferenceContractId === "same_incident_party_ref_v1") {
    return options.parties;
  }
  if (field.directReferenceContractId === "same_incident_decision_ref_v1") {
    return options.decisions;
  }
  if (isPartyRefCollection(field.fieldKey)) {
    return options.parties;
  }
  switch (field.fieldKey) {
    case "comm_log.decision_ids":
    case "handoff.open_decision_ids":
    case "status_review.open_decision_ids":
      return options.decisions;
    case "comm_log.action_task_ids":
    case "handoff.open_task_ids":
    case "status_review.blocked_task_ids":
    case "lesson.follow_up_task_ids":
      return options.taskRequests;
    case "status_review.pending_evidence_ids":
    case "lesson.evidence_refs":
      return options.evidence;
    case "task.linked_record_ids":
    case "decision.support_refs":
      return options.allRecords;
    default:
      return [];
  }
}

function genericFieldUsesReferenceOptions(field: ViewFieldContract): boolean {
  if (
    field.directReferenceContractId === "same_incident_party_ref_v1" ||
    field.directReferenceContractId === "same_incident_decision_ref_v1" ||
    isPartyRefCollection(field.fieldKey)
  ) {
    return true;
  }
  switch (field.fieldKey) {
    case "comm_log.decision_ids":
    case "handoff.open_decision_ids":
    case "status_review.open_decision_ids":
    case "comm_log.action_task_ids":
    case "handoff.open_task_ids":
    case "status_review.blocked_task_ids":
    case "lesson.follow_up_task_ids":
    case "status_review.pending_evidence_ids":
    case "lesson.evidence_refs":
    case "task.linked_record_ids":
    case "decision.support_refs":
      return true;
    default:
      return false;
  }
}

function isPartyRefCollection(fieldKey: string): boolean {
  return (
    fieldKey === "comm_log.audience_party_ids" ||
    fieldKey === "comm_log.attendee_party_ids"
  );
}

function genericCollectionSupportsRemove(fieldKey: string): boolean {
  return fieldKey !== "note.tags";
}

function genericCollectionItems(
  row: EntityApiRow,
  fieldKey: string,
): Array<{ itemRef: string; displayText: string }> {
  const value = row.cells[fieldKey]?.value;
  if (!value || typeof value !== "object" || !("items" in value)) {
    return [];
  }
  const rawItems = (value as { items?: unknown }).items;
  if (!Array.isArray(rawItems)) {
    return [];
  }
  return rawItems.flatMap((item) => {
    if (!item || typeof item !== "object") {
      return [];
    }
    const raw = item as Record<string, unknown>;
    const itemRef = typeof raw.item_ref === "string" ? raw.item_ref : "";
    if (itemRef === "") {
      return [];
    }
    const displayText =
      typeof raw.display_text === "string" && raw.display_text.trim() !== ""
        ? raw.display_text
        : itemRef;
    return [{ itemRef, displayText }];
  });
}

function isMultilineGenericField(field: ViewFieldContract): boolean {
  return (
    field.stringContractId === "multiline_body_v1" ||
    field.fieldKey.endsWith(".body") ||
    field.fieldKey.endsWith(".notes") ||
    field.fieldKey.endsWith(".rationale") ||
    field.fieldKey.endsWith("_summary") ||
    field.fieldKey.endsWith(".details")
  );
}

function parseMutationError(payload: unknown): string {
  const base = parseErrorMessage(payload);
  if (!payload || typeof payload !== "object" || !("error" in payload)) {
    return base;
  }
  const error = payload.error;
  if (!error || typeof error !== "object" || !("conflict" in error)) {
    return base;
  }
  const conflict = error.conflict;
  if (!conflict || typeof conflict !== "object") {
    return base;
  }
  const fieldKey =
    "field_key" in conflict && typeof conflict.field_key === "string"
      ? conflict.field_key
      : null;
  return fieldKey ? `${base}: ${fieldKey}` : base;
}

function parseSameFieldConflict(
  payload: unknown,
): SameFieldConflictPayload | null {
  if (!payload || typeof payload !== "object" || !("error" in payload)) {
    return null;
  }
  const error = payload.error;
  if (
    !error ||
    typeof error !== "object" ||
    !("code" in error) ||
    error.code !== "same_field_conflict" ||
    !("conflict" in error)
  ) {
    return null;
  }
  const conflict = error.conflict;
  if (!conflict || typeof conflict !== "object") {
    return null;
  }
  const object = conflict as Record<string, unknown>;
  if (
    typeof object.conflict_token !== "string" ||
    typeof object.record_id !== "string" ||
    typeof object.field_key !== "string" ||
    typeof object.conflict_resolution_class !== "string" ||
    typeof object.base_row_version !== "number" ||
    typeof object.current_row_version !== "number"
  ) {
    return null;
  }
  const parsed: SameFieldConflictPayload = {
    conflict_token: object.conflict_token,
    record_id: object.record_id,
    field_key: object.field_key,
    conflict_resolution_class: object.conflict_resolution_class,
    base_row_version: object.base_row_version,
    current_row_version: object.current_row_version,
    client_value: object.client_value,
    server_value: object.server_value,
    base_value: object.base_value,
    suggested_merged_value: object.suggested_merged_value,
  };
  if (typeof object.server_updated_by === "string") {
    parsed.server_updated_by = object.server_updated_by;
  }
  if (typeof object.server_updated_at === "string") {
    parsed.server_updated_at = object.server_updated_at;
  }
  return parsed;
}

function enumValuesFor(
  contract: ViewContract,
  fieldKey: string,
  fallback: readonly string[],
): readonly string[] {
  return contract.fieldMap[fieldKey]?.enumValues ?? fallback;
}

function isAssessmentConfidenceBand(
  value: string,
): value is AssessmentConfidenceBand {
  return (
    value === "unset" ||
    value === "low" ||
    value === "medium" ||
    value === "high"
  );
}

function initialAssessmentDraft(): AssessmentCreateDraft {
  const [assessmentState = "unknown"] = enumValuesFor(
    assessmentsContract,
    "assessment.assessment_state",
    ["unknown", "suspected", "confirmed", "disproven", "cleared"],
  );
  const confidenceBand = enumValuesFor(
    assessmentsContract,
    "assessment.confidence_band",
    ["unset", "low", "medium", "high"],
  ).find(isAssessmentConfidenceBand);
  return {
    assessedAt: "",
    assessmentState,
    confidenceBand: confidenceBand ?? "unset",
    rationale: "",
    subjectRecordId: "",
    subjectType: "host",
    supportRecordIds: [],
  };
}

function assessmentColumnWidth(fieldKey: string): number {
  switch (fieldKey) {
    case "assessment.subject_ref":
      return 300;
    case "assessment.rationale":
      return 360;
    case "assessment.assessed_at":
      return 210;
    case "assessment.assessor":
      return 300;
    default:
      return 180;
  }
}

function supportRowLabel(row: TimelineApiRow): string {
  const summary = readStringCell(row, "timeline.summary");
  if (summary !== "") {
    return summary;
  }
  return row.record_id;
}

function surfaceSlug(viewSchemaID: string): string {
  return viewSchemaID.replace(/^cartulary\.view\./, "").replace(/\.v1$/, "");
}

export function WorkbookShell({
  incidentId,
  apiBase,
  onIncidentSnapshot,
  onIncidentAccessLost,
}: WorkbookShellProps) {
  const params = useMemo(() => new URLSearchParams(window.location.search), []);
  const initialViewSchemaID = useMemo(() => {
    const explicit = params.get("view_schema_id");
    if (
      explicit &&
      allWorkbookContracts.some(
        (contract) => contract.viewSchemaId === explicit,
      )
    ) {
      return explicit;
    }
    return timelineViewSchemaId;
  }, [params]);
  const [surface, setSurface] = useState<SurfaceKind>(initialViewSchemaID);
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
          body: JSON.stringify(buildQueryRequest(contract, queryState)),
        },
      );
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const envelope = readEnvelope<ViewQueryEnvelope>(result.payload);
      return envelope.data.rows.map((row) => entityRowFromApi(row, entityType));
    },
    [apiBase, incidentId],
  );

  const loadEntities = useCallback(async () => {
    setEntityLoadError(null);
    try {
      const [nextHosts, nextIdentities] = await Promise.all([
        queryEntityView(hostsViewSchemaId, "host", hostQueryState),
        queryEntityView(identitiesViewSchemaId, "identity", identityQueryState),
      ]);
      setHostRows((current) => [...reconcileRecordRows(current, nextHosts)]);
      setIdentityRows((current) => [
        ...reconcileRecordRows(current, nextIdentities),
      ]);
    } catch (error) {
      const message =
        error instanceof Error ? error.message : "Entity load failed.";
      if (
        typeof message === "string" &&
        (message.includes("incident_not_found") ||
          message.includes("authorization_denied"))
      ) {
        onIncidentAccessLost?.();
      }
      setEntityLoadError(message);
    }
  }, [
    hostQueryState,
    identityQueryState,
    onIncidentAccessLost,
    queryEntityView,
  ]);

  const applyHostFilter = useCallback(() => {
    setHostQueryState((current) => applyFilterDraft(current, hostFilterDraft));
    setHostFilterDraft((current) => ({
      ...current,
      booleanValue: "",
      value: "",
    }));
  }, [hostFilterDraft]);

  const applyIdentityFilter = useCallback(() => {
    setIdentityQueryState((current) =>
      applyFilterDraft(current, identityFilterDraft),
    );
    setIdentityFilterDraft((current) => ({
      ...current,
      booleanValue: "",
      value: "",
    }));
  }, [identityFilterDraft]);

  const isSpecializedSurface =
    surface === timelineViewSchemaId ||
    surface === hostsViewSchemaId ||
    surface === identitiesViewSchemaId ||
    surface === assessmentsViewSchemaId;

  const loadAssessmentSurface = useCallback(async () => {
    if (surface !== assessmentsViewSchemaId) {
      return;
    }
    setAssessmentLoadError(null);
    try {
      const result = await fetchJSON<ViewQueryEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${assessmentsViewSchemaId}/query`,
        ),
        {
          method: "POST",
          body: JSON.stringify(
            buildQueryRequest(assessmentsContract, assessmentQueryState),
          ),
        },
      );
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const envelope = readEnvelope<ViewQueryEnvelope>(result.payload);
      setAssessmentRows(envelope.data.rows);
    } catch (error) {
      const message =
        error instanceof Error ? error.message : "Assessment load failed.";
      if (
        typeof message === "string" &&
        (message.includes("incident_not_found") ||
          message.includes("authorization_denied"))
      ) {
        onIncidentAccessLost?.();
      }
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
      return;
    }
    setGenericLoadError(null);
    try {
      const result = await fetchJSON<ViewQueryEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${surface}/query`,
        ),
        {
          method: "POST",
          body: JSON.stringify(
            buildQueryRequest(activeContract, genericQueryState),
          ),
        },
      );
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const envelope = readEnvelope<ViewQueryEnvelope>(result.payload);
      setGenericRows(envelope.data.rows);
    } catch (error) {
      const message =
        error instanceof Error ? error.message : "Surface load failed.";
      if (
        typeof message === "string" &&
        (message.includes("incident_not_found") ||
          message.includes("authorization_denied"))
      ) {
        onIncidentAccessLost?.();
      }
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

  useEffect(() => {
    void Promise.all([loadEntities(), loadSessionRole()]);
  }, [loadEntities, loadSessionRole]);

  useEffect(() => {
    setGenericQueryState(emptyWorkbookQueryState());
    setGenericFilterDraft(defaultFilterDraft(activeContract));
    setGenericRows([]);
    setGenericLoadError(null);
  }, [activeContract]);

  useEffect(() => {
    void loadGenericSurface();
  }, [loadGenericSurface]);

  useEffect(() => {
    void loadAssessmentSurface();
  }, [loadAssessmentSurface]);

  useEffect(() => {
    const next = new URLSearchParams(window.location.search);
    next.set("incident_id", incidentId);
    next.set("view_schema_id", surface);
    next.delete("surface");
    window.history.replaceState({}, "", `/?${next.toString()}`);
  }, [incidentId, surface]);

  return (
    <section style={panelStyle}>
      <div style={heroStyle}>
        <p style={eyebrowStyle}>Cartulary</p>
        <h1 style={headlineStyle}>Timeline workbook shell</h1>
        <p style={bodyStyle}>
          Phase 3 workbook behavior stays intact while Phase 4 adds entity
          mention review, stub creation, merge review, and same-surface
          auto-resolution disclosure.
        </p>
      </div>

      <div style={toolbarStyle}>
        <div style={tabStripStyle}>
          {builtInSurfaceIDs.map((viewSchemaID) => {
            const contract = requireViewContract(viewSchemaID);
            return (
              <button
                key={viewSchemaID}
                data-testid={`surface-tab-${surfaceSlug(viewSchemaID)}`}
                style={{
                  ...surfaceTabStyle,
                  ...(surface === viewSchemaID ? surfaceTabActiveStyle : null),
                }}
                type="button"
                onClick={() => {
                  setSurface(viewSchemaID);
                }}
              >
                {contract.title}
              </button>
            );
          })}
          <label style={systemViewSelectLabelStyle}>
            System view
            <select
              data-testid="system-view-selector"
              style={selectStyle}
              value={
                builtInSurfaceIDs.includes(
                  surface as (typeof builtInSurfaceIDs)[number],
                )
                  ? ""
                  : surface
              }
              onChange={(event) => {
                if (event.target.value) {
                  setSurface(event.target.value);
                }
              }}
            >
              <option value="">Select view</option>
              {allWorkbookContracts
                .filter(
                  (contract) =>
                    !builtInSurfaceIDs.includes(
                      contract.viewSchemaId as (typeof builtInSurfaceIDs)[number],
                    ),
                )
                .map((contract) => (
                  <option
                    key={contract.viewSchemaId}
                    value={contract.viewSchemaId}
                  >
                    {contract.title}
                  </option>
                ))}
            </select>
          </label>
        </div>
        <div style={roleBadgeStyle}>
          Current incident role: {currentIncidentRole || "viewer"}
        </div>
      </div>

      <IncidentAdminPanel
        apiBase={apiBase}
        currentIncidentRole={currentIncidentRole}
        incidentId={incidentId}
        onIncidentAccessLost={onIncidentAccessLost}
        onIncidentSnapshot={onIncidentSnapshot}
        onSessionRoleChange={loadSessionRole}
      />

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
          onRefreshEntities={loadEntities}
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
            setAssessmentQueryState((current) =>
              applyFilterDraft(current, assessmentFilterDraft),
            );
            setAssessmentFilterDraft((current) => ({
              ...current,
              booleanValue: "",
              value: "",
            }));
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
            setGenericQueryState((current) =>
              applyFilterDraft(current, genericFilterDraft),
            );
            setGenericFilterDraft((current) => ({
              ...current,
              booleanValue: "",
              value: "",
            }));
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
        />
      )}
    </section>
  );
}

const panelStyle = {
  width: "min(96rem, 100%)",
  margin: "0 auto",
  padding: "2rem",
  borderRadius: "1.5rem",
  background: "rgb(255 252 247 / 0.92)",
  boxShadow: "0 24px 80px rgb(29 78 70 / 0.12)",
  border: "1px solid rgb(185 204 196 / 0.8)",
};

const heroStyle = {
  marginBottom: "1.5rem",
};

const toolbarStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "1rem",
  alignItems: "center",
  flexWrap: "wrap" as const,
  marginBottom: "1rem",
};

const workbookStyle = {
  marginTop: "1.5rem",
};

const headerStyle = {
  position: "relative" as const,
  display: "flex",
  justifyContent: "space-between",
  gap: "1rem",
  alignItems: "flex-start",
  marginBottom: "1rem",
};

const blurSurfaceButtonStyle = {
  position: "absolute" as const,
  inset: 0,
  zIndex: 1,
  border: 0,
  padding: 0,
  margin: 0,
  background: "transparent",
  color: "transparent",
  cursor: "default",
};

const statusClusterStyle = {
  display: "grid",
  gap: "0.25rem",
  minWidth: "8rem",
  textAlign: "right" as const,
};

const statusLabelStyle = {
  fontSize: "0.75rem",
  letterSpacing: "0.08em",
  textTransform: "uppercase" as const,
  color: "rgb(81 110 103)",
};

const statusValueStyle = {
  fontSize: "1rem",
};

const headerPresenceStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "flex-end",
  gap: "0.25rem",
  minHeight: "1.5rem",
};

const presenceAvatarStyle = {
  display: "inline-grid",
  placeItems: "center",
  width: "1.5rem",
  height: "1.5rem",
  borderRadius: "999px",
  border: "1px solid rgb(56 126 104)",
  background: "rgb(236 253 245)",
  color: "rgb(20 83 45)",
  fontSize: "0.72rem",
  fontWeight: 700,
};

const presenceOverflowStyle = {
  ...presenceAvatarStyle,
  width: "auto",
  minWidth: "1.5rem",
  paddingInline: "0.35rem",
};

const presenceEmptyStyle = {
  fontSize: "0.75rem",
  color: "rgb(87 112 104)",
};

const eyebrowStyle = {
  margin: 0,
  fontSize: "0.78rem",
  letterSpacing: "0.12em",
  textTransform: "uppercase" as const,
  color: "rgb(87 112 104)",
};

const headlineStyle = {
  margin: "0.35rem 0 0.5rem",
  fontSize: "2rem",
  lineHeight: 1.1,
};

const bodyStyle = {
  margin: 0,
  lineHeight: 1.5,
  color: "rgb(53 79 72)",
};

const splitShellStyle = {
  display: "grid",
  gap: "1rem",
  alignItems: "start",
  gridTemplateColumns: "minmax(0, 3fr) minmax(20rem, 1.25fr)",
};

const gridShellStyle = {
  overflow: "hidden",
  overflowAnchor: "none" as const,
  borderRadius: "1rem",
  border: "1px solid rgb(199 214 207)",
  background: "rgb(255 255 255 / 0.82)",
  blockSize: "min(70vh, 46rem)",
  minBlockSize: "18rem",
};

const stateCellStyle = {
  display: "grid",
  gap: "0.35rem",
  alignItems: "start",
};

const rowPresenceStyle = {
  display: "inline-flex",
  alignItems: "center",
  width: "fit-content",
  maxWidth: "100%",
  borderRadius: "999px",
  border: "1px solid rgb(56 126 104)",
  background: "rgb(236 253 245)",
  color: "rgb(20 83 45)",
  padding: "0.15rem 0.4rem",
  fontSize: "0.72rem",
  fontWeight: 700,
};

const actionStackStyle = {
  display: "grid",
  gap: "0.5rem",
};

const timelineActionTopRowStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(2, minmax(0, 1fr))",
  gap: "0.35rem",
  alignItems: "center",
};

const genericMutationPanelStyle = {
  display: "grid",
  gap: "0.75rem",
  padding: "1rem",
  borderRadius: "1rem",
  border: "1px solid rgb(199 214 207)",
  background: "rgb(247 250 248)",
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
  borderRadius: "0.75rem",
  border: "1px solid rgb(192 205 198)",
  background: "rgb(255 255 255)",
  padding: "0.65rem 0.75rem",
  font: "inherit",
  color: "inherit",
};

const textareaStyle = {
  ...inputStyle,
  resize: "vertical" as const,
};

const replacementInputStyle = {
  ...inputStyle,
  fontSize: "0.9rem",
};

const timelineReplacementInputStyle = {
  ...replacementInputStyle,
  boxSizing: "border-box" as const,
  fontSize: "0.82rem",
  width: "100%",
};

const actionButtonStyle = {
  borderRadius: "999px",
  border: "1px solid rgb(129 165 154)",
  background: "rgb(234 244 239)",
  color: "rgb(34 74 63)",
  padding: "0.55rem 0.9rem",
  font: "inherit",
  cursor: "pointer",
};

const timelineActionButtonStyle = {
  ...actionButtonStyle,
  boxSizing: "border-box" as const,
  fontSize: "0.85rem",
  lineHeight: 1.1,
  padding: "0.45rem 0.3rem",
  width: "100%",
};

const secondaryActionButtonStyle = {
  ...actionButtonStyle,
  background: "rgb(247 249 247)",
};

const destructiveActionButtonStyle = {
  ...actionButtonStyle,
  borderColor: "rgb(180 83 9)",
  background: "rgb(255 247 237)",
  color: "rgb(146 64 14)",
};

const conflictMarkerStyle = {
  ...secondaryActionButtonStyle,
  marginTop: "0.35rem",
  borderColor: "rgb(180 83 9)",
  color: "rgb(146 64 14)",
  background: "rgb(255 251 235)",
  padding: "0.35rem 0.6rem",
  fontSize: "0.85rem",
};

const cellPresenceStyle = {
  display: "inline-flex",
  width: "fit-content",
  marginTop: "0.35rem",
  borderRadius: "999px",
  border: "1px solid rgb(37 99 235)",
  background: "rgb(239 246 255)",
  color: "rgb(30 64 175)",
  padding: "0.2rem 0.45rem",
  fontSize: "0.75rem",
  fontWeight: 700,
};

const conflictResolverStyle = {
  display: "grid",
  gap: "0.75rem",
  padding: "1rem",
  borderTop: "1px solid rgb(199 214 207)",
  background: "rgb(255 251 235)",
};

const conflictResolverGridStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(12rem, 1fr))",
  gap: "0.75rem",
};

const genericErrorTextStyle = {
  margin: 0,
  color: "rgb(153 27 27)",
  fontWeight: 700,
};

const evidenceAccessMessageStyle = {
  margin: 0,
  fontSize: "0.85rem",
  color: "rgb(87 109 103)",
};

const timelineEvidenceCellStyle = {
  display: "inline-flex",
  alignItems: "center",
  gap: "0.45rem",
  minWidth: 0,
};

const timelineEvidenceFlagBaseStyle = {
  borderRadius: "999px",
  padding: "0.15rem 0.42rem",
  fontSize: "0.72rem",
  lineHeight: 1.2,
};

const timelineEvidenceFlagOnStyle = {
  ...timelineEvidenceFlagBaseStyle,
  border: "1px solid rgb(73 143 108)",
  background: "rgb(225 245 233)",
  color: "rgb(29 90 59)",
};

const timelineEvidenceFlagOffStyle = {
  ...timelineEvidenceFlagBaseStyle,
  border: "1px solid rgb(178 191 184)",
  background: "rgb(242 246 243)",
  color: "rgb(54 74 67)",
};

const evidencePreviewPanelStyle = {
  display: "grid",
  gap: "0.75rem",
  margin: "1rem 0",
  padding: "1rem",
  borderRadius: "1rem",
  border: "1px solid rgb(199 214 207)",
  background: "rgb(255 255 255 / 0.86)",
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
  border: "1px solid rgb(199 214 207)",
  borderRadius: "0.75rem",
  background: "rgb(255 255 255)",
};

const labelStyle = {
  display: "grid",
  gap: "0.4rem",
  fontSize: "0.95rem",
  color: "rgb(52 79 72)",
};

const inspectorShellStyle = {
  borderRadius: "1rem",
  border: "1px solid rgb(199 214 207)",
  background: "rgb(248 251 249)",
  padding: "1rem",
  position: "sticky" as const,
  top: "1rem",
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

const historySectionHeaderStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  gap: "0.75rem",
  flexWrap: "wrap" as const,
};

const historyMetaStyle = {
  margin: 0,
  color: "rgb(87 109 103)",
  fontSize: "0.85rem",
  overflowWrap: "anywhere" as const,
};

const historyMetaGridStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(8rem, 1fr))",
  gap: "0.75rem",
  margin: 0,
};

const historyListStyle = {
  display: "grid",
  gap: "0.75rem",
  margin: 0,
  paddingInlineStart: "1.25rem",
};

const historyItemStyle = {
  display: "grid",
  gap: "0.65rem",
  padding: "0.75rem",
  borderRadius: "0.5rem",
  border: "1px solid rgb(212 224 218)",
  background: "rgb(255 255 255)",
};

const historyItemHeaderStyle = {
  display: "flex",
  alignItems: "baseline",
  justifyContent: "space-between",
  gap: "0.75rem",
  flexWrap: "wrap" as const,
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
  borderRadius: "999px",
  padding: "0.35rem 0.65rem",
  font: "inherit",
  lineHeight: 1.2,
  maxWidth: "100%",
  minWidth: 0,
  overflowWrap: "anywhere" as const,
};

const unresolvedChipStyle = {
  border: "1px dashed rgb(188 118 23)",
  background: "rgb(255 246 219)",
  color: "rgb(124 71 10)",
};

const resolvedChipStyle = {
  border: "1px solid rgb(73 143 108)",
  background: "rgb(225 245 233)",
  color: "rgb(29 90 59)",
};

const autoResolvedChipStyle = {
  border: "1px solid rgb(44 100 138)",
  background: "rgb(223 241 255)",
  color: "rgb(20 67 97)",
};

const dismissedChipStyle = {
  border: "1px solid rgb(128 128 128)",
  background: "rgb(240 240 240)",
  color: "rgb(76 76 76)",
};

const selectedChipStyle = {
  boxShadow: "0 0 0 2px rgb(31 94 70 / 0.22)",
};

const chipMetaStyle = {
  fontSize: "0.72rem",
  textTransform: "uppercase" as const,
  letterSpacing: "0.04em",
};

const aliasChipStyle = {
  ...relationshipChipStyle,
  border: "1px solid rgb(188 205 198)",
  background: "rgb(240 247 243)",
  color: "rgb(45 74 66)",
};

const tagChipStyle = {
  ...relationshipChipStyle,
  border: "1px solid rgb(178 191 184)",
  background: "rgb(242 246 243)",
  color: "rgb(54 74 67)",
};

const emptyRelationshipStyle = {
  color: "rgb(117 136 130)",
  fontSize: "0.9rem",
};

const mentionGroupStyle = {
  display: "grid",
  gap: "0.75rem",
};

const mentionGroupColumnStyle = {
  display: "grid",
  gap: "0.5rem",
};

const groupLabelStyle = {
  margin: 0,
  fontSize: "0.8rem",
  letterSpacing: "0.08em",
  textTransform: "uppercase" as const,
  color: "rgb(89 112 105)",
};

const mentionListButtonStyle = {
  border: "none",
  background: "transparent",
  padding: 0,
  textAlign: "left" as const,
  cursor: "pointer",
};

const mentionListButtonSelectedStyle = {
  outline: "none",
};

const detailListStyle = {
  display: "grid",
  gap: "0.75rem",
  margin: 0,
};

const detailTermStyle = {
  fontSize: "0.75rem",
  letterSpacing: "0.08em",
  textTransform: "uppercase" as const,
  color: "rgb(87 109 103)",
};

const detailValueStyle = {
  margin: "0.2rem 0 0",
};

const inlineButtonRowStyle = {
  display: "flex",
  gap: "0.5rem",
  flexWrap: "wrap" as const,
};

const inspectorActionStackStyle = {
  display: "grid",
  gap: "0.75rem",
};

const noticeStackStyle = {
  display: "grid",
  gap: "0.75rem",
  marginBottom: "1rem",
};

const noticeCardStyle = {
  borderRadius: "1rem",
  border: "1px solid rgb(158 194 214)",
  background: "rgb(240 248 255)",
  padding: "0.85rem 1rem",
  display: "grid",
  gap: "0.5rem",
};

const noticeTitleStyle = {
  margin: 0,
  fontSize: "0.95rem",
  fontWeight: 600,
};

const tabStripStyle = {
  display: "flex",
  gap: "0.5rem",
  flexWrap: "wrap" as const,
};

const surfaceTabStyle = {
  borderRadius: "999px",
  border: "1px solid rgb(176 194 187)",
  background: "rgb(248 251 249)",
  color: "rgb(42 77 67)",
  padding: "0.55rem 0.9rem",
  font: "inherit",
  cursor: "pointer",
};

const surfaceTabActiveStyle = {
  background: "rgb(34 74 63)",
  color: "rgb(250 252 251)",
  borderColor: "rgb(34 74 63)",
};

const systemViewSelectLabelStyle = {
  ...labelStyle,
  minWidth: "16rem",
};

const roleBadgeStyle = {
  borderRadius: "999px",
  background: "rgb(236 244 239)",
  color: "rgb(44 76 66)",
  padding: "0.45rem 0.8rem",
  fontSize: "0.9rem",
};

const selectStyle = {
  ...inputStyle,
  appearance: "auto" as const,
};

const mergePlanStyle = {
  borderRadius: "1rem",
  border: "1px solid rgb(205 219 213)",
  background: "rgb(255 255 255)",
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
  borderRadius: "0.9rem",
  border: "1px solid rgb(210 222 216)",
  background: "rgb(255 255 255)",
  padding: "0.85rem",
  display: "grid",
  gap: "0.55rem",
};
