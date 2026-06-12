import {
  buildGridPresentationRows,
  type GridActionsColumn,
  type GridCellAnchor,
  type GridColumn,
  type GridNavigationIntent,
  type GridRow,
  GridTable,
  GridViewport,
  navigateGridCellAnchor,
  reconcileRecordRows,
  resolveGridPasteTargets,
} from "@cartulary/grid-adapter";
import type {
  ErrorEnvelope,
  EvidenceAttachBlobEnvelope,
  EvidenceAttachBlobRequest,
  EvidenceHandleEnvelope,
  EvidenceHandleIssueRequest,
  ObjectBlobCreateEnvelope,
  ObjectBlobCreateRequest,
  ObjectBlobUploadTarget,
} from "@cartulary/protocol-ts";
import {
  assessmentCreatePanelTestId,
  autoResolutionNoticeTestId,
  autoResolutionReviewButtonTestId,
  autoResolutionUndoButtonTestId,
  cellPresenceMarkerTestId,
  conflictMarkerTestId,
  currentIncidentRoleTestId,
  dataTestIdSelector,
  draftCellTestId,
  draftRelationshipItemsTestId,
  draftRowCreateButtonTestId,
  draftTimelineCollectionInputTestId,
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
  gridRowGutterTestId,
  gridRowTestId,
  gridScrollportSelector,
  gridShellTestId,
  gridSortHeaderTestId,
  incidentControlsPanelTestId,
  incidentControlsTriggerTestId,
  mentionCreateEntityButtonTestId,
  mentionDismissButtonTestId,
  mentionItemTestId,
  mentionResolveExistingButtonTestId,
  mentionResolveTargetSelectTestId,
  mentionRestoreUnresolvedButtonTestId,
  pasteConflictItemTestId,
  pendingQueueCountTestId,
  pendingQueueNoticeTestId,
  phase1LandingTestId,
  phase1RouteTestId,
  relationshipChipTestId,
  relationshipItemsTestId,
  rowCellTestId,
  rowHistoryActionTestId,
  rowHistoryDeleteButtonTestId,
  rowHistoryItemTestId,
  rowHistoryLoadingTestId,
  rowHistoryMessageTestId,
  rowHistoryOpenButtonTestId,
  rowHistoryOpenInspectorButtonTestId,
  rowHistoryOpenSelectedButtonTestId,
  rowHistoryPanelTestId,
  rowHistoryRestoreButtonTestId,
  rowInspectButtonTestId,
  rowPresenceMarkerTestId,
  savedViewCreateButtonTestId,
  savedViewDeleteButtonTestId,
  savedViewDuplicateButtonTestId,
  savedViewNameInputTestId,
  savedViewOptionTestId,
  savedViewScopeSelectTestId,
  savedViewSelectorTestId,
  savedViewSetDefaultButtonTestId,
  savedViewSetHomeButtonTestId,
  savedViewStatusTestId,
  savedViewUpdateButtonTestId,
  saveStateTestId,
  statusStripQueueCountTestId,
  surfaceTabTestId,
  systemViewSwitcherGroupTestId,
  systemViewSwitcherMenuTestId,
  systemViewSwitcherOptionTestId,
  systemViewSwitcherTriggerTestId,
  timelineCollectionInputTestId,
  timelineDraftEvidenceAttachSectionTestId,
  timelineDraftEvidenceFileInputTestId,
  timelineEvidenceAttachSectionTestId,
  timelineEvidenceFileInputTestId,
  timelineInspectorSectionTestId,
  timelineMutationSubstrateReadyTestId,
  timelinePreviewRowTestId,
  timelineRowMarkReviewedButtonTestId,
  timelineRowReplacementInputTestId,
  timelineRowSupersedeButtonTestId,
  timelineRowVersionTestId,
  timelineScalarEditorTestId,
  type WorkbookShellSlot,
  type WorkbookSurface,
  workbookShellReadyTestId,
  workbookShellSlotLabel,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import {
  normalizeViewRowPatchV1,
  normalizeViewRowV1,
  requireViewContract,
  resolveHeaderSortFieldKey,
  type ViewContract,
  type ViewFieldContract,
  visibleFields,
} from "@cartulary/view-contracts";
import {
  type CSSProperties,
  type Dispatch,
  Fragment,
  type ClipboardEvent as ReactClipboardEvent,
  type FocusEvent as ReactFocusEvent,
  type KeyboardEvent as ReactKeyboardEvent,
  type MouseEvent as ReactMouseEvent,
  type ReactNode,
  type SetStateAction,
  startTransition,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { flushSync } from "react-dom";
import {
  apiPath,
  csrfCookieName,
  csrfHeaderName,
  readCookie,
} from "./browserApi";
import {
  buildEvidenceCountDisplayViewModel,
  buildEvidenceLifecycleViewModel,
  type EvidenceLifecycleViewModel,
} from "./evidenceLifecycleViewModel";
import { IncidentAdminPanel } from "./IncidentAdminPanel";
import { publicErrorStatusText } from "./publicError";
import { WorkbookGridControls } from "./WorkbookGridControls";
import {
  captureViewportAnchor,
  computeRestoredViewportScroll,
  isRectFullyVisibleWithinContainer,
  type ScrollPosition,
  type ViewportSnapshot,
} from "./workbookContinuity";
import {
  normalizeWorkbookViewRows,
  workbookContractColumns,
} from "./workbookContractRows";
import {
  mapWorkbookKeyboardCommand,
  type WorkbookKeyboardCommand,
} from "./workbookKeyboard";
import {
  type AutoResolutionNotice,
  buildAutoResolutionNotices,
  buildInspectorMentions,
  type CollectionItem,
  type DismissedMention,
  type InspectorMention,
  type MentionChipState,
  type RelationshipFieldKey,
  readCollectionItems,
} from "./workbookMentionChips";
import {
  buildStableMutationSignature,
  deriveWorkbookSaveState,
  type PendingReplayUnitInput,
  type PendingReplayUnitState,
  parsePendingReplayPublicError,
  pendingReplayCapacity,
  sameFieldConflictQueueKey,
  WorkbookPendingQueueModel,
  type WorkbookSaveStateConflictAnchor,
} from "./workbookPendingQueue";
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
  workbookLayoutStateFromSavedViewLayoutJson,
  workbookQueryStateFromSavedViewQueryJson,
} from "./workbookQuery";
import {
  buildMentionActionPayload,
  buildMentionPatchPayload,
  isRecordChangedMessage,
  type MentionResolutionAction,
  type RecordChangedPayload,
  shouldIgnoreSelfOriginatedRecordChange,
} from "./workbookShellPhase4";
import {
  createWorkbookSocketLifecycleState,
  reduceWorkbookSocketLifecycle,
  type WorkbookSocketLifecycleAction,
  type WorkbookSocketLifecycleEffect,
} from "./workbookSocketLifecycle";
import {
  normalizeWorkbookStartupSelection,
  type WorkbookSheetRef,
  workbookStartupQueryFromURLParams,
} from "./workbookStartup";
import {
  assessmentsViewSchemaId,
  commLogViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  findingsViewSchemaId,
  forensicKeywordsViewSchemaId,
  handoffViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  investigativeQueriesViewSchemaId,
  isStandardizedWorkbookViewSchemaId,
  knownWorkbookViewSchemaId,
  lessonViewSchemaId,
  listSystemWorkbookSurfaceGroups,
  listWorkbookSurfaceRegistryEntries,
  notesViewSchemaId,
  partiesViewSchemaId,
  requiredBuiltInWorkbookSurfaceIds,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "./workbookSurfaceRegistry";

export { pendingReplayCapacity };

const timelineContract = requireViewContract(timelineViewSchemaId);
const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);
const assessmentsContract = requireViewContract(assessmentsViewSchemaId);
const allWorkbookContracts = listWorkbookSurfaceRegistryEntries().map(
  (entry) => entry.contract,
);
const systemWorkbookSurfaceGroups = listSystemWorkbookSurfaceGroups();
const systemViewSwitcherEntries = systemWorkbookSurfaceGroups.flatMap((group) =>
  group.entries.map((entry) => ({
    contract: entry.contract,
    groupToken: group.token,
    viewSchemaId: entry.viewSchemaId,
  })),
);

type SaveState = "Syncing" | "Saved" | "Conflict";
type FilterDraftSetter = Dispatch<SetStateAction<FilterDraft>>;
type MutationErrorSetter = Dispatch<SetStateAction<string | null>>;
type MutationStateSetter = Dispatch<SetStateAction<SaveState>>;
type WorkbookQueryStateSetter = Dispatch<SetStateAction<WorkbookQueryState>>;
type SurfaceKind = string;
type EditableField =
  | "timeline.occurred_at"
  | "timeline.summary"
  | "timeline.details"
  | "timeline.source_text";
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
  viewSchemaId: string | null;
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

function rowStillHasAutoResolvedNotice(
  row: WorkbookRow,
  notice: AutoResolutionNotice,
) {
  if (row.recordId !== notice.rowRecordId) {
    return false;
  }
  const item = [
    ...row.collectionValues.hostRefs,
    ...row.collectionValues.identityRefs,
  ].find((candidate) => candidate.itemRef === notice.itemRef);
  return (
    item?.itemKind === "resolved_ref" &&
    item.autoResolved &&
    item.resolvedRecordId !== null
  );
}

type TimelineWorkbookProps = {
  incidentId: string;
  apiBase?: string | undefined;
  sheetRef?: WorkbookSheetRef | undefined;
  reloadToken?: number | undefined;
  savedViewSelector?: ReactNode | undefined;
  filterDraft?: FilterDraft | undefined;
  onFilterDraftChange?: FilterDraftSetter | undefined;
  onQueryStateChange?: WorkbookQueryStateSetter | undefined;
  queryState?: WorkbookQueryState | undefined;
  hostEntities?: EntityRow[];
  identityEntities?: EntityRow[];
  entityIndex?: Record<string, EntityRow>;
  currentIncidentRole?: IncidentRole | null;
  onRefreshEntities?: () => Promise<void> | void;
};

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

const workbookShellId = workbookShellReadyTestId();

function WorkbookShellSlotRegion({
  children,
  slot,
  style,
  viewSchemaId,
}: {
  readonly children: ReactNode;
  readonly slot: WorkbookShellSlot;
  readonly style?: CSSProperties | undefined;
  readonly viewSchemaId?: string | undefined;
}) {
  return (
    <section
      aria-label={workbookShellSlotLabel(slot)}
      data-testid={workbookShellSlotTestId(slot)}
      data-view-schema-id={viewSchemaId}
      data-workbook-shell-id={workbookShellId}
      style={style}
    >
      {children}
    </section>
  );
}

type WorkbookQueryEnvelope = {
  data: {
    incident_id: string;
    view_schema_id: string;
    rows: unknown[];
  };
};

type SavedViewResource = {
  saved_view_id: string;
  view_schema_id: string;
  display_name: string;
  scope: "private" | "shared" | "system";
  query_json: unknown;
  layout_json: unknown;
  owner_user_id: string | null;
  saved_view_version: number;
};

type SavedViewListEnvelope = {
  data: {
    saved_views: SavedViewResource[];
  };
  meta?: {
    paging?: {
      has_more?: boolean;
      next_cursor?: string | null;
    };
  };
};

type SavedViewEnvelope = {
  data: SavedViewResource;
};

function workbookContractForViewSchemaId(viewSchemaId: string): ViewContract {
  return (
    allWorkbookContracts.find(
      (contract) => contract.viewSchemaId === viewSchemaId,
    ) ?? timelineContract
  );
}

function normalizeSavedViewResource(value: unknown): SavedViewResource | null {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return null;
  }
  const record = value as Record<string, unknown>;
  if (
    typeof record.saved_view_id !== "string" ||
    typeof record.view_schema_id !== "string" ||
    typeof record.display_name !== "string" ||
    !isStandardizedWorkbookViewSchemaId(record.view_schema_id)
  ) {
    return null;
  }
  const scope = normalizeSavedViewScope(record.scope);
  if (scope === null) {
    return null;
  }
  const version =
    typeof record.saved_view_version === "number" &&
    Number.isSafeInteger(record.saved_view_version)
      ? record.saved_view_version
      : 0;
  return {
    saved_view_id: record.saved_view_id,
    view_schema_id: record.view_schema_id,
    display_name: record.display_name,
    scope,
    query_json: record.query_json ?? {},
    layout_json: record.layout_json ?? {},
    owner_user_id:
      typeof record.owner_user_id === "string" ? record.owner_user_id : null,
    saved_view_version: version,
  };
}

function normalizeSavedViewScope(
  value: unknown,
): SavedViewResource["scope"] | null {
  return value === "private" || value === "shared" || value === "system"
    ? value
    : null;
}

function canMutateSavedView(
  savedView: SavedViewResource | null,
  currentUserId: string | null,
  currentIncidentRole: IncidentRole | null,
): boolean {
  if (savedView === null || savedView.scope === "system") {
    return false;
  }
  if (currentIncidentRole === "admin") {
    return true;
  }
  return (
    currentUserId !== null &&
    savedView.owner_user_id !== null &&
    savedView.owner_user_id === currentUserId
  );
}

function savedViewQueryJsonForPersistence(
  contract: ViewContract,
  value: unknown,
) {
  return buildSavedViewQueryJson(
    contract,
    workbookQueryStateFromSavedViewQueryJson(contract, value),
  );
}

function savedViewLayoutJsonForPersistence(
  contract: ViewContract,
  value: unknown,
) {
  return buildSavedViewLayoutJson(
    contract,
    workbookLayoutStateFromSavedViewLayoutJson(contract, value),
  );
}

type TimelineMutationEnvelope = {
  data: {
    view_schema_id: string;
    change_set_id?: string;
    row: unknown;
  };
};

type MentionActionEnvelope = {
  data: {
    incident_id: string;
    entity_mention: {
      entity_mention_id: string;
      source_record_id: string;
      source_field_key: string;
      entity_type: "host" | "identity" | string;
      raw_text: string;
      resolution_status: "unresolved" | "resolved" | "dismissed" | string;
      resolved_record_id: string | null;
      row_version: number;
      resolution_method: string | null;
    };
    source_record: {
      record_id: string;
      row_version: number;
    };
    change_set_id: string;
  };
};

type TimelineClipboardPasteEnvelope = {
  data: {
    view_schema_id: string;
    change_set_id?: string;
    rows: unknown[];
    conflicts?: SameFieldConflictPayload[];
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
  anchor: WorkbookSaveStateConflictAnchor;
  conflict: SameFieldConflictPayload;
  focusKey: string;
  localValue: unknown;
  mergedDraft: string;
};

type PasteConflictGroupState = {
  keys: string[];
};

type PendingReplayRuntimeMeta = {
  focusField: FocusFieldKey;
  focusKey: string;
  surface: TimelineScalarEditorSurface;
  rowSnapshot: WorkbookRow;
  continueOnFreshDraft: boolean;
  detectAutoResolution: boolean;
  promoteToCommittedRowInspect: boolean;
  viewportContinuityToken: number;
};

type PendingReplayAdmissionRequest = PendingReplayUnitInput &
  PendingReplayRuntimeMeta;

type PendingQueueRuntime = {
  model: WorkbookPendingQueueModel;
  metaByUnitId: Map<string, PendingReplayRuntimeMeta>;
  refreshBlockedRecordIds: Map<string, number>;
  refreshInFlightCount: number;
  refreshReplayBlockAllCount: number;
  resetRefreshInFlight: boolean;
  replayScheduled: boolean;
};

type PendingQueueRuntimeSnapshot = {
  queuedCount: number;
  inFlightCount: number;
  haltedMessage: string | null;
  authPaused: boolean;
  overflowMessage: string | null;
  resetRefreshInFlight: boolean;
};

type PendingRefreshReplayBlockScope =
  | { kind: "all" }
  | { kind: "record"; recordId: string }
  | { kind: "none" };

function refreshBlocksRecord(
  pending: PendingQueueRuntime,
  recordId: string | null,
): boolean {
  if (pending.resetRefreshInFlight !== true) {
    return false;
  }
  if (pending.refreshReplayBlockAllCount > 0) {
    return true;
  }
  return recordId !== null && pending.refreshBlockedRecordIds.has(recordId);
}

function refreshBlocksPendingUnit(
  pending: PendingQueueRuntime,
  unit: PendingReplayUnitState,
): boolean {
  return refreshBlocksRecord(pending, unit.recordId);
}

function saveStateConflictAnchorsFromLocalConflicts(
  conflicts: Record<string, LocalConflictState>,
): WorkbookSaveStateConflictAnchor[] {
  return Object.values(conflicts).map((entry) => ({ ...entry.anchor }));
}

function saveStateConflictAnchorFromPayload(
  conflict: SameFieldConflictPayload,
): WorkbookSaveStateConflictAnchor {
  return {
    record_id: conflict.record_id,
    field_key: conflict.field_key,
    base_row_version: conflict.base_row_version,
    current_row_version: conflict.current_row_version,
  };
}

type WorkbookTimingEvent = {
  at: number;
  name: string;
  [key: string]: unknown;
};

declare global {
  interface Window {
    __cartularyWorkbookTimingProbe?:
      | {
          events: WorkbookTimingEvent[];
          mark?: (event: WorkbookTimingEvent) => void;
        }
      | undefined;
  }
}

const maxTimelineFreshnessRetryDepth = 2;

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

export type RecordHistoryRollbackAction =
  | "change_set"
  | "history_entry"
  | "row_restore";

export type RecordHistoryItem = {
  actor_user_id: string;
  committed_at: string;
  history_item_ref: string;
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

const rowHistoryRollbackActionOrder = [
  "history_entry",
  "change_set",
  "row_restore",
] as const satisfies readonly RecordHistoryRollbackAction[];

export function buildRecordRollbackTargetFromHistoryAction(
  item: RecordHistoryItem,
  action: RecordHistoryRollbackAction,
): Record<string, unknown> | null {
  if (!item.available_rollback_actions.includes(action)) {
    return null;
  }
  if (action === "history_entry") {
    return typeof item.history_entry_ref === "string" &&
      item.history_entry_ref.trim() !== ""
      ? { kind: "history_entry", history_entry_ref: item.history_entry_ref }
      : null;
  }
  if (action === "change_set") {
    return typeof item.change_set_id === "string" &&
      item.change_set_id.trim() !== ""
      ? { kind: "change_set", change_set_id: item.change_set_id }
      : null;
  }
  return isPositiveInteger(item.revision_no)
    ? { kind: "row_restore", restore_to_revision_no: item.revision_no }
    : null;
}

function normalizeRecordHistoryData(
  data: RecordHistoryData,
): RecordHistoryData {
  if (!Array.isArray(data.items)) {
    throw new Error("row history items must be an array");
  }
  const seenItemRefs = new Set<string>();
  for (const item of data.items) {
    if (!isNonEmptyString(item.history_item_ref)) {
      throw new Error("row history item is missing history_item_ref");
    }
    if (seenItemRefs.has(item.history_item_ref)) {
      throw new Error("row history item has duplicate history_item_ref");
    }
    seenItemRefs.add(item.history_item_ref);
    if (!isNonEmptyString(item.change_set_id)) {
      throw new Error("row history item is missing change_set_id");
    }
    if (
      item.history_entry_ref !== undefined &&
      !isNonEmptyString(item.history_entry_ref)
    ) {
      throw new Error("row history item has invalid history_entry_ref");
    }
    if (
      item.revision_no !== undefined &&
      !isPositiveInteger(item.revision_no)
    ) {
      throw new Error("row history item has invalid revision_no");
    }
    validateRowHistoryActions(item);
  }
  return data;
}

function validateRowHistoryActions(item: RecordHistoryItem) {
  if (!Array.isArray(item.available_rollback_actions)) {
    throw new Error("row history actions must be an array");
  }
  let previousIndex = -1;
  for (const action of item.available_rollback_actions as unknown[]) {
    if (!isRowHistoryRollbackAction(action)) {
      throw new Error("row history action token is invalid");
    }
    const actionIndex = rowHistoryRollbackActionOrder.indexOf(action);
    if (actionIndex <= previousIndex) {
      throw new Error("row history actions are not canonical");
    }
    previousIndex = actionIndex;
    if (buildRecordRollbackTargetFromHistoryAction(item, action) === null) {
      throw new Error("row history action is missing its selector");
    }
  }
}

function isRowHistoryRollbackAction(
  value: unknown,
): value is RecordHistoryRollbackAction {
  return (
    value === "history_entry" ||
    value === "change_set" ||
    value === "row_restore"
  );
}

function isNonEmptyString(value: unknown): value is string {
  return typeof value === "string" && value.trim() !== "";
}

function isPositiveInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value > 0;
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

type ViewApiRow = {
  record_id: string;
  row_version: number;
  cells: Record<string, { value: unknown }>;
  group_values?: Record<string, unknown>;
};

type TimelineApiRow = ViewApiRow & {
  view_schema_id: string;
};

type EntityApiRow = ViewApiRow;

type WorkbookStartupEnvelope = {
  data?: unknown;
};

type WorkbookPresenceMode = "editing" | "idle" | "viewing";

type WorkbookPresenceInput = {
  sheet_ref: WorkbookSheetRef;
  mode: WorkbookPresenceMode;
  record_id?: string;
  field_key?: string;
};

type WorkbookFocusAnchor = GridCellAnchor & {
  readonly surface: WorkbookSurface;
};

type WorkbookGridFocusRuntime = {
  readonly anchor: WorkbookFocusAnchor | null;
  readonly navigate: (
    current: GridCellAnchor,
    intent: GridNavigationIntent,
  ) => void;
  readonly update: (recordId: string | null, fieldKey: string) => void;
};

type PresenceRecord = WorkbookPresenceInput & {
  connection_id: string;
  display_name: string;
  expires_at: string;
  observed_at: string;
  user_id: string;
};

function formatWorkbookFocusAnchor(anchor: WorkbookFocusAnchor | null) {
  return anchor === null
    ? "cleared"
    : `${anchor.surface}:${anchor.recordId}:${anchor.fieldKey}`;
}

function useWorkbookGridFocus<Row>({
  columns,
  getGroupLabel,
  groupBy,
  rows,
  surface,
}: {
  readonly columns: readonly GridColumn<Row>[];
  readonly getGroupLabel?: (
    row: Row,
    fieldKey: string,
  ) => string | null | undefined;
  readonly groupBy?: string | null | undefined;
  readonly rows: readonly GridRow<Row>[];
  readonly surface: WorkbookSurface;
}): WorkbookGridFocusRuntime {
  const [anchor, setAnchor] = useState<WorkbookFocusAnchor | null>(null);

  const update = useCallback(
    (recordId: string | null, fieldKey: string) => {
      if (
        recordId === null ||
        recordId.trim() === "" ||
        !columns.some((column) => column.fieldKey === fieldKey)
      ) {
        setAnchor(null);
        return;
      }
      setAnchor({ fieldKey, recordId, surface });
    },
    [columns, surface],
  );

  const navigate = useCallback(
    (current: GridCellAnchor, intent: GridNavigationIntent) => {
      const nextAnchor = navigateGridCellAnchor({
        columns,
        current,
        intent,
        presentationRows: buildGridPresentationRows({
          getGroupLabel,
          groupBy,
          rows,
        }),
      });
      if (nextAnchor === null) {
        setAnchor(null);
        return;
      }
      setAnchor({ ...nextAnchor, surface });
      window.setTimeout(() => {
        const element = document.querySelector<HTMLElement>(
          dataTestIdSelector(
            rowCellTestId(nextAnchor.recordId, nextAnchor.fieldKey),
          ),
        );
        element?.focus({ preventScroll: true });
      }, 0);
    },
    [columns, getGroupLabel, groupBy, rows, surface],
  );

  return { anchor, navigate, update };
}

function WorkbookFocusAnchorStatus({
  anchor,
}: {
  readonly anchor: WorkbookFocusAnchor | null;
}) {
  return (
    <span data-testid="workbook-focus-anchor" style={visuallyHiddenStyle}>
      {formatWorkbookFocusAnchor(anchor)}
    </span>
  );
}

function FocusableWorkbookCell({
  children,
  fieldKey,
  focus,
  onPaste,
  recordId,
}: {
  readonly children: ReactNode;
  readonly fieldKey: string;
  readonly focus: WorkbookGridFocusRuntime;
  readonly onPaste?: (
    event: ReactClipboardEvent<HTMLElement>,
    anchor: { readonly fieldKey: string; readonly recordId: string },
  ) => void;
  readonly recordId: string;
}) {
  return (
    // biome-ignore lint/a11y: The grid adapter owns the gridcell role; this wrapper is only the focus/paste anchor inside that cell.
    <span
      data-testid={rowCellTestId(recordId, fieldKey)}
      onFocus={() => {
        focus.update(recordId, fieldKey);
      }}
      onPaste={
        onPaste
          ? (event) => {
              onPaste(event, { fieldKey, recordId });
            }
          : undefined
      }
      onKeyDown={(event) => {
        const command = mapWorkbookKeyboardCommand(event);
        if (command.preventDefault) {
          event.preventDefault();
        }
        if (command.kind === "navigate") {
          focus.navigate({ fieldKey, recordId }, command.intent);
        }
      }}
      style={focusableCellStyle}
      // biome-ignore lint/a11y/noNoninteractiveTabindex: The grid adapter owns cell semantics; this wrapper keeps keyboard focus anchored inside the rendered cell.
      tabIndex={0}
    >
      {children}
    </span>
  );
}

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
  hosts: GenericReferenceOption[];
  identities: GenericReferenceOption[];
  notes: GenericReferenceOption[];
  timeline: GenericReferenceOption[];
  noteSourceRecords: GenericReferenceOption[];
  allRecords: GenericReferenceOption[];
};

type PartyLinkPair = {
  key: string;
  label: string;
  textFieldKey: string;
  refFieldKey: string;
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

type LoadRowsOptions = {
  showLoading: boolean;
  freshnessRetryDepth?: number;
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

export type WorkbookVersionedRecord = {
  readonly recordId: string | null;
  readonly rowVersion: number | null;
};

export type WorkbookRecordFreshnessDecision = {
  readonly comparable: boolean;
  readonly stale: boolean;
};

export function decideWorkbookRecordFreshness(
  incoming: WorkbookVersionedRecord,
  knownRowVersion: number | null | undefined,
): WorkbookRecordFreshnessDecision {
  if (
    incoming.recordId === null ||
    incoming.rowVersion === null ||
    knownRowVersion === null ||
    knownRowVersion === undefined
  ) {
    return {
      comparable: false,
      stale: false,
    };
  }
  return {
    comparable: true,
    stale: incoming.rowVersion < knownRowVersion,
  };
}

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

function timelineFocusFieldForFieldKey(fieldKey: string): FocusFieldKey | null {
  const binding = timelineFieldBinding(fieldKey);
  if (binding.kind === "scalar") {
    return binding.key;
  }
  if (binding.kind === "collection") {
    return binding.draftKey;
  }
  return null;
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

function entityContractColumnWidth(field: ViewFieldContract): number {
  switch (field.fieldKey) {
    case "host.display_name":
    case "identity.display_name":
      return 240;
    case "host.hostname":
    case "identity.upn":
    case "identity.email":
    case "identity.sam_account_name":
      return 260;
    case "host.aliases":
    case "identity.aliases":
      return 320;
    case "host.linked_event_count":
    case "identity.linked_event_count":
    case "host.evidence_count":
    case "identity.evidence_count":
      return 140;
    case "host.edited_at":
    case "identity.edited_at":
      return 180;
    case "host.host_state":
    case "identity.identity_state":
    case "identity.mfa_state":
    case "identity.reset_status":
      return 150;
    default:
      return 200;
  }
}

function genericContractColumnWidth(field: ViewFieldContract): number {
  if (field.fieldKey.endsWith(".body") || field.fieldKey.endsWith(".summary")) {
    return 320;
  }
  if (
    field.fieldKey.endsWith(".edited_at") ||
    field.fieldKey.endsWith(".updated_at") ||
    field.fieldKey.endsWith(".timestamp_utc")
  ) {
    return 180;
  }
  if (field.readKind === "collection") {
    return 260;
  }
  return field.defaultHidden ? 160 : 220;
}

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

function handleWorkbookLoadFailure(
  error: unknown,
  fallbackMessage: string,
  onIncidentAccessLost: (() => void) | undefined,
): string {
  const message =
    typeof error === "string"
      ? error
      : error instanceof Error
        ? error.message
        : fallbackMessage;
  if (
    message.includes("incident_not_found") ||
    message.includes("authorization_denied")
  ) {
    onIncidentAccessLost?.();
  }
  return message;
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
    viewSchemaId: timelineViewSchemaId,
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

function createDraftRowForKey(rowKey: string): WorkbookRow | null {
  if (!rowKey.startsWith("draft-")) {
    return null;
  }
  return { ...createDraftRow(0), key: rowKey };
}

function normalizeValue(value: string): string {
  return value.trim();
}

function recordWorkbookTiming(
  name: string,
  details: Record<string, unknown> = {},
) {
  const probe =
    typeof window === "undefined"
      ? undefined
      : window.__cartularyWorkbookTimingProbe;
  if (probe === undefined) {
    return;
  }
  const event = {
    at: performance.now(),
    name,
    ...details,
  };
  probe.events.push(event);
  probe.mark?.(event);
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

function validateTimelineViewSchemaId(value: unknown, source: string) {
  if (value !== timelineViewSchemaId) {
    throw new Error(
      `Timeline view row envelope failed: ${source} view_schema_id must be ${timelineViewSchemaId}.`,
    );
  }
}

function materializeTimelineCells(
  cells: Readonly<Record<string, { readonly value: unknown }>>,
): Record<string, { value: unknown }> {
  return Object.fromEntries(
    Object.entries(cells).map(([fieldKey, cell]) => [
      fieldKey,
      { value: cell.value },
    ]),
  );
}

function normalizeTimelineFullRow(
  row: unknown,
  source: string,
): TimelineApiRow {
  const normalized = normalizeViewRowV1(timelineContract, row, source);
  return {
    view_schema_id: normalized.viewSchemaId,
    record_id: normalized.recordId,
    row_version: normalized.rowVersion,
    cells: materializeTimelineCells(normalized.cells),
    ...(normalized.groupValues === undefined
      ? {}
      : { group_values: { ...normalized.groupValues } }),
  };
}

type TimelinePatchCells = NonNullable<
  RecordChangedPayload["affected_views"][number]["patch_cells"]
>;

function normalizeTimelinePatchCells(
  patch: unknown,
  source: string,
): TimelinePatchCells {
  const normalized = normalizeViewRowPatchV1(timelineContract, patch, source);
  return {
    record_id: normalized.recordId,
    row_version: normalized.rowVersion,
    cells: materializeTimelineCells(normalized.cells),
    ...(normalized.groupValues === undefined
      ? {}
      : { group_values: { ...normalized.groupValues } }),
  };
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
    viewSchemaId: row.view_schema_id,
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
  patch: TimelinePatchCells,
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
        if (typeof object.raw_text === "string") {
          return object.raw_text;
        }
        if (typeof object.alias_text === "string") {
          return object.alias_text;
        }
        return typeof object.display_text === "string"
          ? object.display_text
          : null;
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
  const currentCommittedByRecordId = new Map(
    currentRows
      .filter(
        (row): row is WorkbookRow & { recordId: string } =>
          row.recordId !== null,
      )
      .map((row) => [row.recordId, row]),
  );
  const committedRows = reconcileRecordRows(
    currentRows.filter((row) => row.recordId !== null),
    incomingRows,
  ).map((row) => {
    let rowWithLocalState = row;
    if (row.recordId === null) {
      return row;
    }
    const current = currentCommittedByRecordId.get(row.recordId);
    if (current !== undefined) {
      if (
        current.pendingSignature !== null ||
        current.collectionDrafts.hostRefs !== "" ||
        current.collectionDrafts.identityRefs !== "" ||
        current.collectionDrafts.tags !== ""
      ) {
        rowWithLocalState = {
          ...rowWithLocalState,
          collectionDrafts: current.collectionDrafts,
          pendingSignature: current.pendingSignature,
        };
      }
    }
    return rowWithMaterializedScalarDrafts(
      rowWithLocalState,
      draftValueForFocusKey,
    );
  });
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

function buildCollectionActions(
  fieldKey: CollectionFieldKey,
  rawInput: string,
) {
  const actions = rawInput
    .split(/\r?\n/u)
    .filter((segment) => segment.trim() !== "")
    .map((rawText) =>
      fieldKey === "timeline.tags"
        ? {
            op: "add_tag",
            tag_name: rawText,
          }
        : {
            op: "add_token",
            raw_text: rawText,
          },
    );
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
      field.fieldKey,
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

function buildScalarPatchIntent(row: WorkbookRow, clientTxnId: string) {
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

  if (changes.length < 1) {
    return null;
  }

  return {
    view_schema_id: timelineViewSchemaId,
    client_txn_id: clientTxnId,
    changes,
  };
}

function buildCollectionPatchIntent(
  fieldKey: CollectionFieldKey,
  draftValue: string,
  clientTxnId: string,
) {
  const actionPayload = buildCollectionActions(fieldKey, draftValue);
  if (actionPayload === null) {
    return null;
  }

  return {
    view_schema_id: timelineViewSchemaId,
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

function materializePendingReplayPayload(
  unit: PendingReplayUnitState,
  currentRow: WorkbookRow | undefined,
) {
  if (unit.kind === "create") {
    return unit.payloadIntent;
  }
  if (currentRow?.rowVersion === null || currentRow?.rowVersion === undefined) {
    return null;
  }
  return {
    ...unit.payloadIntent,
    base_row_version: currentRow.rowVersion,
  };
}

export function clipboardTextLooksTabular(text: string): boolean {
  return text.includes("\n") || text.includes("\r") || text.includes("\t");
}

function clipboardGridDimensions(text: string): {
  readonly columnCount: number;
  readonly rowCount: number;
} {
  const rows = parseClipboardTableForDimensions(text);
  const columnCount = rows.reduce((max, row) => Math.max(max, row.length), 1);
  return {
    columnCount,
    rowCount: Math.max(1, rows.length),
  };
}

function parseClipboardTableForDimensions(text: string): string[][] {
  const normalized = text.replace(/\r\n/g, "\n").replace(/\r/g, "\n");
  const trimmed = normalized.replace(/\n+$/, "");
  if (trimmed === "") {
    return [[""]];
  }
  const delimiter = trimmed.includes("\t") ? "\t" : ",";
  const rows: string[][] = [];
  let row: string[] = [];
  let cell = "";
  let quoted = false;
  for (let index = 0; index < trimmed.length; index += 1) {
    const char = trimmed[index];
    if (char === '"') {
      if (quoted && trimmed[index + 1] === '"') {
        cell += '"';
        index += 1;
      } else {
        quoted = !quoted;
      }
      continue;
    }
    if (!quoted && char === delimiter) {
      row.push(cell);
      cell = "";
      continue;
    }
    if (!quoted && char === "\n") {
      row.push(cell);
      rows.push(row);
      row = [];
      cell = "";
      continue;
    }
    cell += char;
  }
  row.push(cell);
  rows.push(row);
  return rows;
}

function readEnvelope<T>(payload: unknown): T {
  return payload as T;
}

type LatestQueryRuntime = {
  controller: AbortController | null;
  sequence: number;
};

type LatestQueryRequest = {
  isCurrent: () => boolean;
  signal: AbortSignal;
};

function beginLatestQuery(runtime: {
  current: LatestQueryRuntime;
}): LatestQueryRequest {
  const previousController = runtime.current.controller;
  const controller = new AbortController();
  const sequence = runtime.current.sequence + 1;
  runtime.current = { controller, sequence };
  previousController?.abort();

  return {
    signal: controller.signal,
    isCurrent: () =>
      runtime.current.sequence === sequence &&
      runtime.current.controller === controller &&
      !controller.signal.aborted,
  };
}

function abortLatestQuery(runtime: { current: LatestQueryRuntime }) {
  runtime.current.controller?.abort();
  runtime.current = {
    controller: null,
    sequence: runtime.current.sequence + 1,
  };
}

function isAbortError(error: unknown): boolean {
  return error instanceof DOMException
    ? error.name === "AbortError"
    : error instanceof Error && error.name === "AbortError";
}

async function fetchJSON<T>(
  input: RequestInfo | URL,
  init?: RequestInit,
  options: {
    onJSONParsed?: () => void;
    onResponse?: (response: Response) => void;
  } = {},
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
  const requestURL = input instanceof Request ? input.url : String(input);
  if (
    window.__cartularyWorkbookTimingProbe !== undefined &&
    method === "POST" &&
    requestURL.includes("/views/cartulary.view.timeline.v1/rows")
  ) {
    headers["X-Cartulary-Timing-Debug"] = "1";
  }

  const response = await fetch(input, {
    credentials: "include",
    ...init,
    headers,
  });
  options.onResponse?.(response);
  const payload = (await response.json()) as
    | T
    | {
        error?: {
          code?: string;
          message?: string;
          retryable?: unknown;
          conflict?: unknown;
          details?: unknown;
        };
      };
  options.onJSONParsed?.();
  return { ok: response.ok, status: response.status, payload };
}

async function uploadObjectBlobTarget(
  apiBase: string | undefined,
  uploadTarget: ObjectBlobUploadTarget,
  file: File,
): Promise<void> {
  const uploadHref =
    uploadTarget.href.startsWith("/") && apiBase
      ? apiPath(apiBase, uploadTarget.href)
      : uploadTarget.href;
  const headers = new Headers(uploadTarget.headers ?? undefined);
  let hasContentType = false;
  headers.forEach((_value, key) => {
    if (key.toLowerCase() === "content-type") {
      hasContentType = true;
    }
  });
  if (!hasContentType) {
    headers.set("Content-Type", file.type || "application/octet-stream");
  }
  let lastStatus = 0;
  let lastDetail = "";
  for (let attempt = 0; attempt < 3; attempt += 1) {
    const upload = await fetch(uploadHref, {
      method: uploadTarget.method ?? "PUT",
      credentials: "omit",
      headers,
      body: file,
    });
    if (upload.ok) {
      return;
    }
    lastStatus = upload.status;
    lastDetail = await readUploadFailureDetail(upload);
    if (attempt < 2 && isRetryableUploadFailure(upload.status, lastDetail)) {
      await sleep(200 * (attempt + 1));
      continue;
    }
    break;
  }
  throw new Error(
    lastDetail === ""
      ? `upload_failed_${lastStatus}`
      : `upload_failed_${lastStatus}: ${lastDetail}`,
  );
}

async function createAndAttachEvidenceBlob({
  apiBase,
  attachClientTxnId,
  baseRowVersion,
  createClientTxnId,
  evidenceRecordId,
  file,
  incidentId,
}: {
  readonly apiBase: string | undefined;
  readonly attachClientTxnId: () => string;
  readonly baseRowVersion: number;
  readonly createClientTxnId: () => string;
  readonly evidenceRecordId: string;
  readonly file: File;
  readonly incidentId: string;
}): Promise<void> {
  const createBlobRequest = {
    incident_id: incidentId,
    client_txn_id: createClientTxnId(),
    byte_size: file.size,
    filename_hint: file.name || null,
    content_type_hint: file.type || null,
  } satisfies ObjectBlobCreateRequest;
  const createBlob = await fetchJSON<ObjectBlobCreateEnvelope>(
    apiPath(apiBase, "/api/v1/object-blobs"),
    {
      method: "POST",
      body: JSON.stringify(createBlobRequest),
    },
  );
  if (!createBlob.ok) {
    throw new Error(evidencePublicErrorMessage(createBlob.payload));
  }
  const blobEnvelope = readEnvelope<ObjectBlobCreateEnvelope>(
    createBlob.payload,
  );
  await uploadObjectBlobTarget(apiBase, blobEnvelope.data.upload_target, file);

  const attachRequest = {
    object_blob_id: blobEnvelope.data.object_blob_id,
    base_row_version: baseRowVersion,
    client_txn_id: attachClientTxnId(),
  } satisfies EvidenceAttachBlobRequest;
  const attach = await fetchJSON<EvidenceAttachBlobEnvelope>(
    apiPath(
      apiBase,
      `/api/v1/evidence-records/${evidenceRecordId}/attach-blob`,
    ),
    {
      method: "POST",
      body: JSON.stringify(attachRequest),
    },
  );
  if (!attach.ok) {
    throw new Error(evidencePublicErrorMessage(attach.payload));
  }
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

async function readUploadFailureDetail(response: Response): Promise<string> {
  try {
    return (await response.text()).replace(/\s+/g, " ").slice(0, 300);
  } catch {
    return "";
  }
}

function isRetryableUploadFailure(status: number, detail: string): boolean {
  if (status !== 503 && status !== 504) {
    return false;
  }
  return detail === "" || detail.includes('"retryable":true');
}

function sleep(milliseconds: number): Promise<void> {
  return new Promise((resolve) => window.setTimeout(resolve, milliseconds));
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
  sheetRef: WorkbookSheetRef = {
    kind: "view_schema",
    id: timelineViewSchemaId,
  },
): WorkbookPresenceInput {
  const input: WorkbookPresenceInput = {
    sheet_ref: { ...sheetRef },
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

function mentionChipStateForItem(
  item: CollectionItem | InspectorMention,
): MentionChipState {
  if ("chipState" in item) {
    return item.chipState;
  }
  if (item.itemKind !== "resolved_ref") {
    return "unresolved";
  }
  if (item.autoResolved) {
    return "auto-resolved";
  }
  if (item.resolutionMethod === "explicit_resolve_route") {
    return "manual-resolution";
  }
  return "resolved";
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

const rawEvidenceStorageDetailPatterns = [
  /\bhttps?:\/\//iu,
  /\bs3:\/\//iu,
  /\bobject:\/\//iu,
  /\bminio\b/iu,
  /\bseaweedfs?\b/iu,
  /\bbucket(?:\b|_)/iu,
  /\bobject[-_ ]?store(?:\b|_)/iu,
  /\bstorage[-_ ]?backend(?:\b|_)/iu,
  /\b(?:storage|object)[-_ ]?key(?:\b|_)/iu,
  /\bobject[-_ ]?blob[-_ ]?storage[-_ ]?key(?:\b|_)/iu,
  /\/(?:home|var|tmp|usr|app|workspace|data|mnt|srv)\//iu,
  /\b(?:local|filesystem|s3|gcs|azure)[-_ ]?backend\b/iu,
] as const;

function containsRawEvidenceStorageDetail(value: string): boolean {
  return rawEvidenceStorageDetailPatterns.some((pattern) =>
    pattern.test(value),
  );
}

function safeEvidencePublicText(value: unknown): string | null {
  if (
    typeof value !== "string" &&
    typeof value !== "number" &&
    typeof value !== "boolean"
  ) {
    return null;
  }
  const text = String(value).trim();
  if (
    text === "" ||
    text.length > 240 ||
    containsRawEvidenceStorageDetail(text)
  ) {
    return null;
  }
  return text;
}

function evidencePublicError(payload: unknown): ErrorEnvelope["error"] | null {
  if (!payload || typeof payload !== "object" || !("error" in payload)) {
    return null;
  }
  const error = payload.error;
  if (!error || typeof error !== "object") {
    return null;
  }
  return error as ErrorEnvelope["error"];
}

function evidencePublicErrorMessage(
  payload: unknown,
  fallback = "Evidence request failed.",
): string {
  const error = evidencePublicError(payload);
  if (error === null) {
    return fallback;
  }
  const reason =
    error.details && typeof error.details === "object"
      ? safeEvidencePublicText(error.details.reason_code)
      : null;
  const code = safeEvidencePublicText(error.code);
  if (code !== null && reason !== null) {
    return `${code}: ${reason}`;
  }
  if (reason !== null) {
    return reason;
  }
  const message = safeEvidencePublicText(error.message);
  if (message !== null) {
    return message;
  }
  if (code !== null) {
    return code;
  }
  const statusText = safeEvidencePublicText(
    publicErrorStatusText({ status: error.status }, error.status),
  );
  return statusText ?? fallback;
}

function evidenceAccessMessageLiveRegion(
  message: string,
  evidenceAccess: EvidenceLifecycleViewModel,
): { ariaLive: "assertive" | "polite"; role: "alert" | "status" } {
  const normalized = message.toLowerCase();
  const isBlockingMessage =
    evidenceAccess.messageTone === "danger" ||
    evidenceAccess.stateKey === "blocked" ||
    evidenceAccess.stateKey === "failed" ||
    evidenceAccess.stateKey === "inconsistent" ||
    evidenceAccess.stateKey === "preview_blocked" ||
    evidenceAccess.stateKey === "public_error" ||
    /\b(?:blob_failed|evidence_access_unavailable|failed|inconsistent|quarantined|unavailable|unsupported_preview)\b/u.test(
      normalized,
    );
  return isBlockingMessage
    ? { ariaLive: "assertive", role: "alert" }
    : { ariaLive: "polite", role: "status" };
}

function resolvePublicEvidenceHandleHref(href: string): string | null {
  const trimmed = href.trim();
  if (
    containsRawEvidenceStorageDetail(trimmed) ||
    !/^\/api\/v1\/evidence-handles\/[A-Za-z0-9._~!$&'()*+,;=:@%-]+(?:\?[A-Za-z0-9._~!$&'()*+,;=:@%/?-]*)?$/u.test(
      trimmed,
    )
  ) {
    return null;
  }
  return trimmed;
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
  const chipState = mentionChipStateForItem(item);
  const isResolved =
    chipState === "resolved" ||
    chipState === "auto-resolved" ||
    chipState === "manual-resolution";
  const isDismissed = chipState === "dismissed";
  const isAutoResolved = chipState === "auto-resolved";
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
  const labelPrefix =
    chipState === "manual-resolution"
      ? "Resolved"
      : chipState === "auto-resolved"
        ? "Auto-resolved"
        : chipState === "dismissed"
          ? "Dismissed"
          : chipState === "resolved"
            ? "Resolved"
            : "Unresolved";
  const stateDetail =
    chipState === "manual-resolution"
      ? "; manual resolution"
      : chipState === "auto-resolved" && item.matchedAliasText
        ? `; matched ${item.matchedAliasText}`
        : "";
  const accessibleLabel = `${labelPrefix} ${label}${stateDetail}`;
  const content = (
    <RelationshipChipContent
      chipState={chipState}
      isAutoResolved={isAutoResolved}
      isDismissed={isDismissed}
      isResolved={isResolved}
      label={label}
    />
  );

  return onSelect ? (
    <button
      aria-label={accessibleLabel}
      data-testid={relationshipChipTestId(item.itemRef)}
      tabIndex={0}
      style={chipStyle}
      type="button"
      onClick={onSelect}
    >
      {content}
    </button>
  ) : (
    <span
      aria-label={accessibleLabel}
      data-testid={relationshipChipTestId(item.itemRef)}
      role="note"
      style={chipStyle}
    >
      {content}
    </span>
  );
}

function RelationshipChipContent({
  chipState,
  isAutoResolved,
  isDismissed,
  isResolved,
  label,
}: {
  chipState: MentionChipState;
  isAutoResolved: boolean;
  isDismissed: boolean;
  isResolved: boolean;
  label: string;
}) {
  return (
    <>
      <span>{label}</span>
      {isAutoResolved ? (
        <span data-density-role="narrow-metadata" style={chipMetaStyle}>
          Auto
        </span>
      ) : null}
      {chipState === "manual-resolution" ? (
        <span data-density-role="narrow-metadata" style={chipMetaStyle}>
          Manual
        </span>
      ) : null}
      {chipState === "resolved" ? (
        <span data-density-role="narrow-metadata" style={chipMetaStyle}>
          Resolved
        </span>
      ) : null}
      {!isResolved && !isDismissed ? (
        <span data-density-role="narrow-metadata" style={chipMetaStyle}>
          Unresolved
        </span>
      ) : null}
      {isDismissed ? (
        <span data-density-role="narrow-metadata" style={chipMetaStyle}>
          Dismissed
        </span>
      ) : null}
    </>
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
      data-testid={draftRowCreateButtonTestId()}
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
  onFocusAnchor,
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
  readonly onFocusAnchor: (recordId: string | null, fieldKey: string) => void;
  readonly onFocusRecord: (recordId: string) => void;
  readonly onKeyCommit: (
    event: ReactKeyboardEvent<HTMLInputElement | HTMLTextAreaElement>,
    rowKey: string,
    field: keyof RowValues,
    surface: TimelineScalarEditorSurface,
  ) => void;
  readonly onPasteCommit: (
    event: ReactClipboardEvent<HTMLInputElement | HTMLTextAreaElement>,
    rowKey: string,
    field: keyof RowValues,
    surface: TimelineScalarEditorSurface,
  ) => void;
  readonly registerInput: (
    rowKey: string,
    field: FocusFieldKey,
    surface: TimelineScalarEditorSurface,
    dataTestId: string,
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
    if (surface === "grid") {
      onFocusAnchor(rowRecordId, presenceFieldKey);
    }
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
  const handlePaste = (
    event: ReactClipboardEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    onPasteCommit(event, rowKey, field, surface);
  };
  const inputRef = (element: HTMLInputElement | HTMLTextAreaElement | null) => {
    registerInput(rowKey, field, surface, dataTestId, element);
  };

  if (multiline) {
    return (
      <textarea
        aria-label={accessibleLabel}
        data-testid={dataTestId}
        id={controlId}
        ref={inputRef}
        rows={3}
        style={surface === "grid" ? gridCellTextareaStyle : textareaStyle}
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
      style={surface === "grid" ? gridCellInputStyle : inputStyle}
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
  sheetRef,
  reloadToken = 0,
  savedViewSelector,
  filterDraft: controlledFilterDraft,
  onFilterDraftChange,
  onQueryStateChange,
  queryState: controlledQueryState,
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
  const [saveStateSecondaryMessage, setSaveStateSecondaryMessage] = useState<
    string | null
  >(null);
  const [conflictQueue, setConflictQueue] = useState<
    Record<string, LocalConflictState>
  >({});
  const [activeConflictKey, setActiveConflictKey] = useState<string | null>(
    null,
  );
  const [pasteConflictGroup, setPasteConflictGroup] =
    useState<PasteConflictGroupState | null>(null);
  const [replacementDrafts, setReplacementDrafts] = useState<
    Record<string, string>
  >({});
  const [selectedRowId, setSelectedRowId] = useState<string | null>(null);
  const [workbookFocusAnchor, setWorkbookFocusAnchor] =
    useState<WorkbookFocusAnchor | null>(null);
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
  const [uncontrolledQueryState, setUncontrolledQueryState] =
    useState<WorkbookQueryState>(() => emptyWorkbookQueryState());
  const [uncontrolledFilterDraft, setUncontrolledFilterDraft] =
    useState<FilterDraft>(() => defaultFilterDraft(timelineContract));
  const queryState = controlledQueryState ?? uncontrolledQueryState;
  const setQueryState = onQueryStateChange ?? setUncontrolledQueryState;
  const filterDraft = controlledFilterDraft ?? uncontrolledFilterDraft;
  const setFilterDraft = onFilterDraftChange ?? setUncontrolledFilterDraft;
  const draftCounterRef = useRef(2);
  const clientTxnRef = useRef(1);
  const pendingOpsRef = useRef(0);
  const pendingSignaturesRef = useRef(new Map<string, string>());
  const collectionKeyboardCommitRef = useRef(new Map<string, string>());
  const pendingSocketTxnTimeoutsRef = useRef(new Map<string, number>());
  const saveQueueRef = useRef(Promise.resolve());
  const socketClientInstanceIdRef = useRef<string | null>(null);
  if (socketClientInstanceIdRef.current === null) {
    socketClientInstanceIdRef.current = tabClientInstanceId();
  }
  const pendingQueueRef = useRef<PendingQueueRuntime>({
    model: new WorkbookPendingQueueModel({
      incidentId,
      clientInstanceId: socketClientInstanceIdRef.current,
    }),
    metaByUnitId: new Map(),
    refreshBlockedRecordIds: new Map(),
    refreshInFlightCount: 0,
    refreshReplayBlockAllCount: 0,
    resetRefreshInFlight: false,
    replayScheduled: false,
  });
  const [pendingQueueSnapshot, setPendingQueueSnapshot] =
    useState<PendingQueueRuntimeSnapshot>({
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
  const rowsRef = useRef(rows);
  const committedTimelineRowsRef = useRef(new Map<string, WorkbookRow>());
  const committedTimelineRowVersionsRef = useRef(new Map<string, number>());
  const committedTimelineRowsEpochRef = useRef(0);
  const conflictQueueRef = useRef<Record<string, LocalConflictState>>({});
  const activeSocketRef = useRef<WebSocket | null>(null);
  const socketLifecycleRef = useRef(createWorkbookSocketLifecycleState());
  const socketEstablishedRef = useRef(false);
  const socketConnectionIDRef = useRef<string | null>(null);
  const presenceUpdateTimerRef = useRef<number | null>(null);
  const currentPresenceRef = useRef(currentPresence);
  const workbookFocusAnchorRef = useRef<WorkbookFocusAnchor | null>(null);
  const hasLoadedRowsRef = useRef(false);
  const loadSequenceRef = useRef(0);
  const scalarDraftValuesRef = useRef(new Map<string, string>());
  const loadRowsRef = useRef<(options: LoadRowsOptions) => Promise<void>>(
    async () => undefined,
  );
  const schedulePendingReplayRef = useRef<() => void>(() => undefined);
  const socketReconnectAfterAuthRef = useRef<(() => void) | null>(null);
  const socketResumeTokenRef = useRef<string | null>(null);
  const socketLastSeenStreamSeqRef = useRef(0);
  const rowInputRefs = useRef(
    new Map<string, HTMLInputElement | HTMLTextAreaElement>(),
  );
  const rowInputTestIdsRef = useRef(new Map<string, string>());
  const timelineAnchorColumnsRef = useRef<readonly GridColumn<WorkbookRow>[]>(
    [],
  );
  const timelineAnchorRowsRef = useRef<readonly GridRow<WorkbookRow>[]>([]);
  const gridShellRef = useRef<HTMLDivElement | null>(null);
  const viewportContinuityTokenRef = useRef(1);
  const [viewportContinuityRequest, setViewportContinuityRequest] =
    useState<ViewportContinuityRequest | null>(null);

  useLayoutEffect(() => {
    rowsRef.current = rows;
  }, [rows]);

  const syncSocketLifecycleRefs = useCallback(() => {
    const state = socketLifecycleRef.current;
    socketEstablishedRef.current = state.established;
    socketConnectionIDRef.current = state.connectionId;
    socketResumeTokenRef.current = state.resumeToken;
    socketLastSeenStreamSeqRef.current = state.lastSeenStreamSeq;
  }, []);

  const dispatchSocketLifecycle = useCallback(
    (
      action: WorkbookSocketLifecycleAction,
    ): WorkbookSocketLifecycleEffect[] => {
      const reduction = reduceWorkbookSocketLifecycle(
        socketLifecycleRef.current,
        action,
      );
      socketLifecycleRef.current = reduction.state;
      syncSocketLifecycleRefs();
      return reduction.effects;
    },
    [syncSocketLifecycleRefs],
  );
  const dispatchSocketLifecycleRef = useRef(dispatchSocketLifecycle);
  dispatchSocketLifecycleRef.current = dispatchSocketLifecycle;

  const pruneAutoResolutionNoticesForRows = useCallback(
    (committedRows: readonly WorkbookRow[]) => {
      if (committedRows.length < 1) {
        return;
      }
      setAutoResolutionNotices((current) =>
        current.filter((notice) => {
          const row = committedRows.find(
            (candidate) => candidate.recordId === notice.rowRecordId,
          );
          return (
            row === undefined || rowStillHasAutoResolvedNotice(row, notice)
          );
        }),
      );
    },
    [],
  );

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

  const computeSaveStatePresentation = useCallback(
    (
      pending: PendingQueueRuntime,
      conflicts: Record<string, LocalConflictState> = conflictQueueRef.current,
    ) => {
      const snapshot = pending.model.snapshot();
      return deriveWorkbookSaveState({
        authPaused: snapshot.authPaused,
        halted: snapshot.halted,
        overflow: snapshot.overflow,
        sameFieldConflicts: snapshot.sameFieldConflicts,
        localDraftConflicts:
          saveStateConflictAnchorsFromLocalConflicts(conflicts),
        queuedCount: snapshot.queuedCount,
        inFlightCount: snapshot.inFlightCount,
        refreshPaused: pending.resetRefreshInFlight,
        pendingMutationCount: pendingOpsRef.current,
      });
    },
    [],
  );

  const publishSaveStatePresentation = useCallback(
    (
      pending: PendingQueueRuntime,
      conflicts: Record<string, LocalConflictState> = conflictQueueRef.current,
    ) => {
      const presentation = computeSaveStatePresentation(pending, conflicts);
      setSaveState(presentation.primaryLabel);
      setSaveStateSecondaryMessage(presentation.secondaryMessage);
      return presentation;
    },
    [computeSaveStatePresentation],
  );

  const activeSheetRef = useMemo<WorkbookSheetRef>(
    () => sheetRef ?? { kind: "view_schema", id: timelineViewSchemaId },
    [sheetRef],
  );
  const activeSheetRuntimeRef = useRef(activeSheetRef);
  activeSheetRuntimeRef.current = activeSheetRef;

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
            presence: workbookPresence(
              currentPresenceRef.current,
              activeSheetRuntimeRef.current,
            ),
          },
        }),
      );
    }, 150);
  }, [currentPresence]);

  const updateWorkbookFocusAnchor = useCallback(
    (anchor: WorkbookFocusAnchor | null) => {
      workbookFocusAnchorRef.current = anchor;
      setWorkbookFocusAnchor(anchor);
    },
    [],
  );

  const updateTimelineFocusAnchor = useCallback(
    (recordId: string | null, fieldKey: string) => {
      if (
        recordId === null ||
        recordId.trim() === "" ||
        !timelineAnchorColumnsRef.current.some(
          (column) => column.fieldKey === fieldKey,
        )
      ) {
        updateWorkbookFocusAnchor(null);
        return;
      }
      updateWorkbookFocusAnchor({
        fieldKey,
        recordId,
        surface: timelineViewSchemaId,
      });
    },
    [updateWorkbookFocusAnchor],
  );

  const publishPendingQueueState = useCallback(() => {
    const pending = pendingQueueRef.current;
    const snapshot = pending.model.snapshot();
    setPendingQueueSnapshot({
      queuedCount: snapshot.queuedCount,
      inFlightCount: snapshot.inFlightCount,
      haltedMessage: snapshot.halted?.message ?? null,
      authPaused: snapshot.authPaused,
      overflowMessage: snapshot.overflow?.message ?? null,
      resetRefreshInFlight: pending.resetRefreshInFlight,
    });
    publishSaveStatePresentation(pending);
  }, [publishSaveStatePresentation]);
  const publishPendingQueueStateRef = useRef(publishPendingQueueState);
  publishPendingQueueStateRef.current = publishPendingQueueState;

  const beginRefreshInFlight = useCallback(
    (scope: PendingRefreshReplayBlockScope) => {
      const pending = pendingQueueRef.current;
      pending.refreshInFlightCount += 1;
      if (scope.kind === "all") {
        pending.refreshReplayBlockAllCount += 1;
      } else if (scope.kind === "record") {
        pending.refreshBlockedRecordIds.set(
          scope.recordId,
          (pending.refreshBlockedRecordIds.get(scope.recordId) ?? 0) + 1,
        );
      }
      pending.resetRefreshInFlight = pending.refreshInFlightCount > 0;
      publishPendingQueueState();
    },
    [publishPendingQueueState],
  );

  const finishRefreshInFlight = useCallback(
    (scope: PendingRefreshReplayBlockScope) => {
      const pending = pendingQueueRef.current;
      pending.refreshInFlightCount = Math.max(
        0,
        pending.refreshInFlightCount - 1,
      );
      if (scope.kind === "all") {
        pending.refreshReplayBlockAllCount = Math.max(
          0,
          pending.refreshReplayBlockAllCount - 1,
        );
      } else if (scope.kind === "record") {
        const currentCount = pending.refreshBlockedRecordIds.get(
          scope.recordId,
        );
        if (currentCount === undefined || currentCount <= 1) {
          pending.refreshBlockedRecordIds.delete(scope.recordId);
        } else {
          pending.refreshBlockedRecordIds.set(scope.recordId, currentCount - 1);
        }
      }
      pending.resetRefreshInFlight = pending.refreshInFlightCount > 0;
      publishPendingQueueState();
      schedulePendingReplayRef.current();
    },
    [publishPendingQueueState],
  );
  const beginRefreshInFlightRef = useRef(beginRefreshInFlight);
  beginRefreshInFlightRef.current = beginRefreshInFlight;
  const finishRefreshInFlightRef = useRef(finishRefreshInFlight);
  finishRefreshInFlightRef.current = finishRefreshInFlight;

  useEffect(() => {
    const clientInstanceId =
      socketClientInstanceIdRef.current ?? tabClientInstanceId();
    socketClientInstanceIdRef.current = clientInstanceId;
    const scope = pendingQueueRef.current.model.scope;
    if (
      scope.incidentId === incidentId &&
      scope.clientInstanceId === clientInstanceId
    ) {
      return;
    }
    pendingQueueRef.current = {
      model: new WorkbookPendingQueueModel({ incidentId, clientInstanceId }),
      metaByUnitId: new Map(),
      refreshBlockedRecordIds: new Map(),
      refreshInFlightCount: 0,
      refreshReplayBlockAllCount: 0,
      resetRefreshInFlight: false,
      replayScheduled: false,
    };
    socketLifecycleRef.current = createWorkbookSocketLifecycleState();
    syncSocketLifecycleRefs();
    pendingSignaturesRef.current.clear();
    publishPendingQueueState();
  }, [incidentId, publishPendingQueueState, syncSocketLifecycleRefs]);

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
    applyFilterDraftToQuery(setQueryState, setFilterDraft, filterDraft);
  }, [filterDraft, setFilterDraft, setQueryState]);

  const handleQueryGroupByChange = useCallback(
    (groupBy: string | null) => {
      setQueryState((current) =>
        updateGroupBy(timelineContract, current, groupBy),
      );
    },
    [setQueryState],
  );

  const handleQuerySortToggle = useCallback(
    (fieldKey: string) => {
      setQueryState((current) =>
        toggleSortField(timelineContract, current, fieldKey),
      );
    },
    [setQueryState],
  );

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
  const resolvePendingSocketTxnRef = useRef(resolvePendingSocketTxn);
  resolvePendingSocketTxnRef.current = resolvePendingSocketTxn;

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
    const selectorTestId = rowInputTestIdsRef.current.get(focusKey) ?? null;
    const selector =
      selectorTestId === null
        ? null
        : document.querySelector<HTMLInputElement | HTMLTextAreaElement>(
            dataTestIdSelector(selectorTestId),
          );
    if (selector !== null) {
      return selector;
    }
    const [rowKey, fieldKey, surface] = focusKey.split(":");
    const scalarBinding = Object.values(timelineScalarBindingIndex).find(
      (binding) => binding.key === fieldKey,
    );
    if (
      rowKey !== undefined &&
      surface === "grid" &&
      scalarBinding !== undefined
    ) {
      const fallbackTestId = rowKey.startsWith("draft-")
        ? draftCellTestId(scalarBinding.fieldKey)
        : rowCellTestId(rowKey, scalarBinding.fieldKey);
      const fallback = document.querySelector<
        HTMLInputElement | HTMLTextAreaElement
      >(dataTestIdSelector(fallbackTestId));
      if (fallback !== null) {
        return fallback;
      }
    }
    return rowInputRefs.current.get(focusKey) ?? null;
  }, []);

  const resolveTimelineAnchorElement = useCallback(
    (anchor: GridCellAnchor) => {
      const focusField = timelineFocusFieldForFieldKey(anchor.fieldKey);
      if (focusField !== null) {
        const inputElement = resolveInputElement(
          inputFocusKey(anchor.recordId, focusField, "grid"),
        );
        if (inputElement !== null) {
          return inputElement;
        }
      }
      const testId =
        anchor.fieldKey === "timeline.capture_state"
          ? gridRowGutterTestId(timelineViewSchemaId, anchor.recordId)
          : anchor.fieldKey === "row_version"
            ? timelineRowVersionTestId(anchor.recordId)
            : rowCellTestId(anchor.recordId, anchor.fieldKey);
      return document.querySelector<HTMLElement>(dataTestIdSelector(testId));
    },
    [resolveInputElement],
  );

  const restoreTimelineFocusAnchor = useCallback(
    (anchor: GridCellAnchor) => {
      const element = resolveTimelineAnchorElement(anchor);
      if (element === null) {
        return false;
      }
      if (!element.hasAttribute("tabindex")) {
        element.tabIndex = -1;
      }
      element.focus({ preventScroll: true });
      return document.activeElement === element;
    },
    [resolveTimelineAnchorElement],
  );

  const resolveViewportContinuityElement = useCallback(
    (target: ViewportContinuityTarget) => {
      switch (target.kind) {
        case "row-inspect":
          return document.querySelector<HTMLButtonElement>(
            dataTestIdSelector(rowInspectButtonTestId(target.recordId)),
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
  const beginViewportContinuityRef = useRef(beginViewportContinuity);
  beginViewportContinuityRef.current = beginViewportContinuity;

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
  const advanceViewportContinuityRef = useRef(advanceViewportContinuity);
  advanceViewportContinuityRef.current = advanceViewportContinuity;

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

  const knownTimelineRowVersion = useCallback((recordId: string) => {
    return committedTimelineRowVersionsRef.current.get(recordId);
  }, []);

  const currentCommittedTimelineRow = useCallback((recordId: string) => {
    return (
      committedTimelineRowsRef.current.get(recordId) ??
      rowsRef.current.find(
        (candidate) =>
          candidate.recordId === recordId && candidate.rowVersion !== null,
      ) ??
      null
    );
  }, []);

  const acceptCommittedTimelineRow = useCallback(
    (
      row: WorkbookRow,
    ): { row: WorkbookRow; accepted: boolean; stale: boolean } => {
      if (row.recordId === null || row.rowVersion === null) {
        return { row, accepted: false, stale: false };
      }
      const currentVersion = knownTimelineRowVersion(row.recordId);
      if (decideWorkbookRecordFreshness(row, currentVersion).stale) {
        return {
          row: currentCommittedTimelineRow(row.recordId) ?? row,
          accepted: false,
          stale: true,
        };
      }
      if (currentVersion !== row.rowVersion) {
        committedTimelineRowsEpochRef.current += 1;
      }
      committedTimelineRowVersionsRef.current.set(row.recordId, row.rowVersion);
      committedTimelineRowsRef.current.set(row.recordId, row);
      return { row, accepted: true, stale: false };
    },
    [currentCommittedTimelineRow, knownTimelineRowVersion],
  );

  const acceptCommittedTimelineRows = useCallback(
    (committedRows: readonly WorkbookRow[]) => {
      for (const row of committedRows) {
        acceptCommittedTimelineRow(row);
      }
    },
    [acceptCommittedTimelineRow],
  );

  const isStaleTimelineRowVersion = useCallback(
    (recordId: string, rowVersion: number) =>
      decideWorkbookRecordFreshness(
        { recordId, rowVersion },
        knownTimelineRowVersion(recordId),
      ).stale,
    [knownTimelineRowVersion],
  );

  const acceptTimelineRecordVersion = useCallback(
    (recordId: string, rowVersion: number) => {
      if (isStaleTimelineRowVersion(recordId, rowVersion)) {
        return { accepted: false, stale: true };
      }
      const existing = currentCommittedTimelineRow(recordId);
      if (existing === null) {
        const currentVersion = knownTimelineRowVersion(recordId);
        if (currentVersion !== rowVersion) {
          committedTimelineRowsEpochRef.current += 1;
        }
        committedTimelineRowVersionsRef.current.set(recordId, rowVersion);
        return { accepted: true, stale: false };
      }
      const accepted = acceptCommittedTimelineRow({
        ...existing,
        rowVersion,
        rawRow:
          existing.rawRow === null
            ? null
            : {
                ...existing.rawRow,
                row_version: rowVersion,
              },
      });
      return { accepted: accepted.accepted, stale: accepted.stale };
    },
    [
      acceptCommittedTimelineRow,
      currentCommittedTimelineRow,
      isStaleTimelineRowVersion,
      knownTimelineRowVersion,
    ],
  );

  const acceptTimelineActionResult = useCallback(
    (result: TimelineActionEnvelope["data"]) => {
      const existing =
        committedTimelineRowsRef.current.get(result.record_id) ??
        rowsRef.current.find((row) => row.recordId === result.record_id) ??
        null;
      if (existing === null) {
        acceptTimelineRecordVersion(result.record_id, result.row_version);
        return;
      }
      const accepted = acceptCommittedTimelineRow({
        ...existing,
        rowVersion: result.row_version,
        captureState: result.capture_state,
        rawRow:
          existing.rawRow === null
            ? null
            : {
                ...existing.rawRow,
                row_version: result.row_version,
                cells: {
                  ...existing.rawRow.cells,
                  "timeline.capture_state": {
                    value: result.capture_state,
                  },
                },
              },
      });
      return accepted;
    },
    [acceptCommittedTimelineRow, acceptTimelineRecordVersion],
  );

  const latestCommittedTimelineRow = useCallback(
    (recordId: string) => {
      const visibleRow = rowsRef.current.find(
        (candidate) => candidate.recordId === recordId,
      );
      const knownVersion =
        committedTimelineRowVersionsRef.current.get(recordId);
      if (
        visibleRow !== undefined &&
        visibleRow.rowVersion !== null &&
        (knownVersion === undefined || visibleRow.rowVersion >= knownVersion)
      ) {
        acceptCommittedTimelineRow(visibleRow);
        return visibleRow;
      }

      const committedRow = committedTimelineRowsRef.current.get(recordId);
      if (
        committedRow !== undefined &&
        committedRow.rowVersion !== null &&
        (knownVersion === undefined || committedRow.rowVersion >= knownVersion)
      ) {
        return committedRow;
      }
      return null;
    },
    [acceptCommittedTimelineRow],
  );

  const latestCommittedRowVersion = useCallback(
    (recordId: string) => {
      const row = latestCommittedTimelineRow(recordId);
      return (
        committedTimelineRowVersionsRef.current.get(recordId) ??
        row?.rowVersion ??
        null
      );
    },
    [latestCommittedTimelineRow],
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
      let previousRow = rowsRef.current.find(
        (candidate) => candidate.key === rowKey,
      );
      validateTimelineViewSchemaId(
        envelope.data.view_schema_id,
        "mutation response",
      );
      const responseRow = normalizeTimelineFullRow(
        envelope.data.row,
        "mutation response row",
      );
      recordWorkbookTiming("apply_row_mutation_start", {
        rowKey,
        recordId: responseRow.record_id,
        rowVersion: responseRow.row_version,
      });
      const incomingCommitted = rowFromApi(responseRow);
      const accepted = acceptCommittedTimelineRow(incomingCommitted);
      const committed = accepted.row;
      let draftSummaryKey: string | null = null;
      flushSync(() => {
        setRows((current) => {
          previousRow =
            current.find((candidate) => candidate.key === rowKey) ??
            current.find(
              (candidate) => candidate.recordId === committed.recordId,
            ) ??
            previousRow;
          let replaced = false;
          let nextRows = current.map((row) => {
            if (
              row.key !== rowKey &&
              (committed.recordId === null ||
                row.recordId !== committed.recordId)
            ) {
              return row;
            }
            replaced = true;
            return committed;
          });
          if (!replaced) {
            const draftIndex = nextRows.findIndex(
              (row) => row.recordId === null,
            );
            nextRows =
              draftIndex === -1
                ? [...nextRows, committed]
                : [
                    ...nextRows.slice(0, draftIndex),
                    committed,
                    ...nextRows.slice(draftIndex),
                  ];
          }
          const hydrated = ensureDraftRowWithFreshIndex(
            nextRows,
            nextDraftIndex,
          );
          draftSummaryKey = hydrated.draftSummaryKey;
          rowsRef.current = hydrated.rows;
          return hydrated.rows;
        });
      });

      if (committed.recordId !== null) {
        setDismissedMentionsByRow((current) =>
          pruneDismissedMentions(current, committed),
        );
        pruneAutoResolutionNoticesForRows([committed]);
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
        options.continueOnFreshDraft && draftSummaryKey
          ? ({
              kind: "input",
              focusKey: draftSummaryKey,
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
      if (nextViewportTarget?.kind === "input") {
        resolveInputElement(nextViewportTarget.focusKey)?.focus({
          preventScroll: true,
        });
      }
      recordWorkbookTiming("apply_row_mutation_end", {
        rowKey,
        recordId: committed.recordId,
        rowVersion: committed.rowVersion,
      });
    },
    [
      acceptCommittedTimelineRow,
      advanceViewportContinuity,
      nextDraftIndex,
      pruneAutoResolutionNoticesForRows,
      resolveInputElement,
      selectedRowId,
    ],
  );

  const waitForCommittedRecordIdle = useCallback(
    async (
      recordId: string,
      options: {
        readonly fallbackRowVersion?: number | null | undefined;
        readonly refreshIfMissing?: boolean;
      } = {},
    ): Promise<{ row: WorkbookRow | null; rowVersion: number } | null> => {
      let attemptedRefresh = false;
      for (;;) {
        const pending = pendingQueueRef.current;
        const snapshot = pending.model.snapshot();
        const hasPendingRecordWork = snapshot.units.some(
          (unit) => unit.recordId === recordId,
        );
        if (
          snapshot.authPaused ||
          snapshot.halted !== null ||
          snapshot.overflow !== null ||
          snapshot.sameFieldConflicts.length > 0 ||
          Object.keys(conflictQueueRef.current).length > 0
        ) {
          return null;
        }
        if (!hasPendingRecordWork && !refreshBlocksRecord(pending, recordId)) {
          const row = latestCommittedTimelineRow(recordId);
          const rowVersion =
            latestCommittedRowVersion(recordId) ?? options.fallbackRowVersion;
          if (typeof rowVersion === "number") {
            return { row, rowVersion };
          }
          if (
            options.refreshIfMissing !== false &&
            attemptedRefresh === false
          ) {
            attemptedRefresh = true;
            await loadRowsRef.current({ showLoading: false });
            continue;
          }
          return null;
        }
        await new Promise((resolve) => window.setTimeout(resolve, 16));
      }
    },
    [latestCommittedRowVersion, latestCommittedTimelineRow],
  );

  const refreshTimelineRowsAfterStaleResult = useCallback(
    async (options: LoadRowsOptions) => {
      const nextDepth = (options.freshnessRetryDepth ?? 0) + 1;
      if (nextDepth > maxTimelineFreshnessRetryDepth) {
        if (options.viewportContinuityToken !== undefined) {
          clearViewportContinuity(options.viewportContinuityToken);
        }
        return false;
      }

      const refreshScope: PendingRefreshReplayBlockScope = { kind: "all" };
      beginRefreshInFlight(refreshScope);
      try {
        const retryOptions: LoadRowsOptions = {
          showLoading: false,
          freshnessRetryDepth: nextDepth,
        };
        if (options.viewportContinuityToken !== undefined) {
          retryOptions.viewportContinuityToken =
            options.viewportContinuityToken;
        }
        await loadRowsRef.current(retryOptions);
      } finally {
        finishRefreshInFlight(refreshScope);
      }
      return true;
    },
    [beginRefreshInFlight, clearViewportContinuity, finishRefreshInFlight],
  );

  const freshTimelineRowsForQueryResult = useCallback(
    (incomingRows: readonly WorkbookRow[]) => {
      let hasStaleRows = false;
      const rows: WorkbookRow[] = [];
      for (const row of incomingRows) {
        if (
          row.recordId !== null &&
          decideWorkbookRecordFreshness(
            row,
            knownTimelineRowVersion(row.recordId),
          ).stale
        ) {
          hasStaleRows = true;
          const current = currentCommittedTimelineRow(row.recordId);
          if (current !== null) {
            rows.push(current);
          }
          continue;
        }
        rows.push(row);
      }
      return { hasStaleRows, rows };
    },
    [currentCommittedTimelineRow, knownTimelineRowVersion],
  );

  const loadRows = useCallback(
    async (options: LoadRowsOptions) => {
      const requestSequence = loadSequenceRef.current + 1;
      loadSequenceRef.current = requestSequence;
      const queryStartEpoch = committedTimelineRowsEpochRef.current;

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

      if (queryStartEpoch !== committedTimelineRowsEpochRef.current) {
        const refreshed = await refreshTimelineRowsAfterStaleResult(options);
        if (!refreshed && !hasLoadedRowsRef.current) {
          setIsInitialLoading(false);
        }
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

      let incomingRows: WorkbookRow[];
      try {
        const envelope = readEnvelope<WorkbookQueryEnvelope>(result.payload);
        validateTimelineViewSchemaId(
          envelope.data.view_schema_id,
          "query response",
        );
        incomingRows = envelope.data.rows.map((row, index) =>
          rowFromApi(
            normalizeTimelineFullRow(row, `query response rows[${index}]`),
          ),
        );
      } catch {
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
      const incomingFreshness = freshTimelineRowsForQueryResult(incomingRows);
      if (incomingFreshness.hasStaleRows) {
        const refreshed = await refreshTimelineRowsAfterStaleResult(options);
        if (refreshed) {
          return;
        }
      }
      const { committedRows, rows: hydratedRows } =
        reconcileCommittedRowsWithLocalDrafts({
          currentRows: rowsRef.current,
          incomingRows: incomingFreshness.rows,
          draftValueForFocusKey: (focusKey) =>
            scalarDraftValuesRef.current.get(focusKey),
          nextDraftIndex,
        });
      acceptCommittedTimelineRows(committedRows);
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
      pruneAutoResolutionNoticesForRows(committedRows);
      publishSaveStatePresentation(pendingQueueRef.current);
      hasLoadedRowsRef.current = true;
      setLoadError(null);
      setRefreshError(null);
      setIsInitialLoading(false);
    },
    [
      advanceViewportContinuity,
      acceptCommittedTimelineRows,
      clearViewportContinuity,
      pruneAutoResolutionNoticesForRows,
      publishSaveStatePresentation,
      freshTimelineRowsForQueryResult,
      nextDraftIndex,
      queryBody,
      queryPath,
      refreshTimelineRowsAfterStaleResult,
    ],
  );

  loadRowsRef.current = loadRows;

  const applyRecordChangedPatch = useCallback(
    (payload: RecordChangedPayload) => {
      if (isStaleTimelineRowVersion(payload.record_id, payload.row_version)) {
        return true;
      }
      acceptTimelineRecordVersion(payload.record_id, payload.row_version);
      const affectedView = payload.affected_views.find(
        (view) => view.view_schema_id === timelineViewSchemaId,
      );
      if (
        affectedView?.change_kind !== "patch" ||
        affectedView.patch_cells === undefined
      ) {
        return false;
      }

      let patch: TimelinePatchCells;
      try {
        patch = normalizeTimelinePatchCells(
          affectedView.patch_cells,
          "record_changed patch_cells",
        );
      } catch {
        return false;
      }
      if (patch.record_id !== payload.record_id) {
        return false;
      }
      if (isStaleTimelineRowVersion(patch.record_id, patch.row_version)) {
        return true;
      }
      acceptTimelineRecordVersion(patch.record_id, patch.row_version);
      let patched = false;
      const nextRows = rowsRef.current.map((row) => {
        if (row.recordId !== patch.record_id || row.rawRow === null) {
          return row;
        }
        patched = true;
        const nextRawRow = applyViewRowPatch(row.rawRow, patch);
        const committed = rowFromApi(nextRawRow);
        const accepted = acceptCommittedTimelineRow(committed);
        return {
          ...accepted.row,
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
    [
      acceptCommittedTimelineRow,
      acceptTimelineRecordVersion,
      isStaleTimelineRowVersion,
    ],
  );
  const applyRecordChangedPatchRef = useRef(applyRecordChangedPatch);
  applyRecordChangedPatchRef.current = applyRecordChangedPatch;

  useEffect(() => {
    // Keep saved-view reselection and shell-level refreshes observable here even
    // though loadRows owns the actual query inputs.
    void reloadToken;
    void loadRows({ showLoading: true });
  }, [loadRows, reloadToken]);

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
    publishSaveStatePresentation(pendingQueueRef.current);
  }, [publishSaveStatePresentation]);

  const finishSave = useCallback(
    (nextState: SaveState) => {
      pendingOpsRef.current = Math.max(0, pendingOpsRef.current - 1);
      if (nextState === "Conflict") {
        setSaveState("Conflict");
        setSaveStateSecondaryMessage("Conflict requires review.");
        return;
      }
      publishSaveStatePresentation(pendingQueueRef.current);
    },
    [publishSaveStatePresentation],
  );

  const restoreConflictFocus = useCallback(
    (focusKey: string) => {
      window.setTimeout(() => {
        resolveInputElement(focusKey)?.focus();
      }, 0);
    },
    [resolveInputElement],
  );

  const updateSaveStateForConflicts = useCallback(
    (nextQueue: Record<string, LocalConflictState>) => {
      publishSaveStatePresentation(pendingQueueRef.current, nextQueue);
    },
    [publishSaveStatePresentation],
  );

  const registerSameFieldConflict = useCallback(
    (
      conflict: SameFieldConflictPayload,
      focusKey: string,
      surface: TimelineScalarEditorSurface,
    ) => {
      const queueKey = sameFieldConflictQueueKey(conflict);
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
        const existing = current[queueKey];
        const mergedDraft =
          existing?.mergedDraft ??
          (typeof conflict.suggested_merged_value === "string"
            ? conflict.suggested_merged_value
            : typeof conflict.server_value === "string"
              ? conflict.server_value
              : "");
        const next = {
          ...current,
          [queueKey]: {
            key: queueKey,
            anchor: saveStateConflictAnchorFromPayload(conflict),
            conflict,
            focusKey,
            localValue:
              conflict.client_value === undefined
                ? existing?.localValue
                : conflict.client_value,
            mergedDraft,
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
      dataTestId: string,
      element: HTMLInputElement | HTMLTextAreaElement | null,
    ) => {
      const key = inputFocusKey(rowKey, field, surface);
      if (element === null) {
        rowInputRefs.current.delete(key);
        rowInputTestIdsRef.current.delete(key);
        return;
      }
      rowInputTestIdsRef.current.set(key, dataTestId);
      rowInputRefs.current.set(key, element);
    },
    [],
  );

  const clearPendingSignatureForUnit = useCallback(
    (unit: {
      readonly rowKey: string;
      readonly mutationSignature?: string;
    }) => {
      if (unit.mutationSignature === undefined) {
        return;
      }
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

  const schedulePendingReplayAfter = useCallback((delayMs: number) => {
    const pending = pendingQueueRef.current;
    if (pending.replayScheduled) {
      return;
    }
    pending.replayScheduled = true;
    pendingReplayTimerRef.current = window.setTimeout(() => {
      pendingReplayTimerRef.current = null;
      void replayPendingQueueRef.current();
    }, delayMs);
  }, []);
  const schedulePendingReplay = useCallback(() => {
    schedulePendingReplayAfter(0);
  }, [schedulePendingReplayAfter]);
  schedulePendingReplayRef.current = schedulePendingReplay;

  const schedulePendingReplayRetry = useCallback(() => {
    schedulePendingReplayAfter(1000);
  }, [schedulePendingReplayAfter]);

  const scheduleAuthRecoveryProbe = useCallback(() => {
    if (pendingReplayAuthRetryRef.current !== null) {
      return;
    }
    pendingReplayAuthRetryRef.current = window.setTimeout(async () => {
      pendingReplayAuthRetryRef.current = null;
      if (!pendingQueueRef.current.model.snapshot().authPaused) {
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
        dispatchSocketLifecycle({ type: "auth_recovered" });
        pendingQueueRef.current.model.resumeAfterAuthRecovery();
        publishPendingQueueState();
        schedulePendingReplay();
        socketReconnectAfterAuthRef.current?.();
      } catch {
        scheduleAuthRecoveryProbe();
      }
    }, 1000);
  }, [
    apiBase,
    dispatchSocketLifecycle,
    publishPendingQueueState,
    schedulePendingReplay,
  ]);
  const scheduleAuthRecoveryProbeRef = useRef(scheduleAuthRecoveryProbe);
  scheduleAuthRecoveryProbeRef.current = scheduleAuthRecoveryProbe;

  const replayPendingQueueRef = useRef<() => Promise<void>>(async () => {
    return undefined;
  });

  const requestPendingReplay = useCallback(
    (reason: string) => {
      const pending = pendingQueueRef.current;
      const snapshot = pending.model.snapshot();
      const candidate = pending.model.peekNextQueued();
      const readyForImmediateDrain =
        !snapshot.authPaused &&
        snapshot.halted === null &&
        snapshot.sameFieldConflicts.length === 0 &&
        candidate !== null &&
        !refreshBlocksPendingUnit(pending, candidate.unit) &&
        Object.keys(conflictQueueRef.current).length === 0 &&
        snapshot.inFlightCount === 0 &&
        snapshot.queuedCount > 0;
      if (!readyForImmediateDrain) {
        schedulePendingReplay();
        return;
      }
      if (pendingReplayTimerRef.current !== null) {
        window.clearTimeout(pendingReplayTimerRef.current);
        pendingReplayTimerRef.current = null;
      }
      pending.replayScheduled = false;
      recordWorkbookTiming("pending_replay_drain_immediate", { reason });
      void replayPendingQueueRef.current();
    },
    [schedulePendingReplay],
  );

  const enqueuePendingReplayUnit = useCallback(
    (unit: PendingReplayAdmissionRequest) => {
      const pending = pendingQueueRef.current;
      const {
        focusField,
        focusKey,
        surface,
        rowSnapshot,
        continueOnFreshDraft,
        detectAutoResolution,
        promoteToCommittedRowInspect,
        viewportContinuityToken,
        ...input
      } = unit;
      const meta: PendingReplayRuntimeMeta = {
        focusField,
        focusKey,
        surface,
        rowSnapshot,
        continueOnFreshDraft,
        detectAutoResolution,
        promoteToCommittedRowInspect,
        viewportContinuityToken,
      };
      const admission = pending.model.admit(input);
      if (admission.status === "duplicate") {
        clearViewportContinuity(unit.viewportContinuityToken);
        publishPendingQueueState();
        return;
      }
      if (admission.status === "refused") {
        clearPendingSignatureForUnit(unit);
        clearViewportContinuity(unit.viewportContinuityToken);
        publishPendingQueueState();
        return;
      }

      pending.metaByUnitId.set(admission.unit.id, meta);
      pendingSignaturesRef.current.set(
        admission.unit.rowKey,
        admission.unit.mutationSignature,
      );
      recordWorkbookTiming("pending_unit_admitted", {
        clientTxnId: admission.unit.clientTxnId,
        kind: admission.unit.kind,
        rowKey: admission.unit.rowKey,
      });
      publishPendingQueueState();
      requestPendingReplay(
        admission.status === "coalesced" ? "coalesced_unit" : "admitted_unit",
      );
    },
    [
      clearPendingSignatureForUnit,
      clearViewportContinuity,
      publishPendingQueueState,
      requestPendingReplay,
    ],
  );

  replayPendingQueueRef.current = async () => {
    const pending = pendingQueueRef.current;
    pending.replayScheduled = false;
    const snapshot = pending.model.snapshot();
    if (
      snapshot.authPaused ||
      snapshot.halted !== null ||
      snapshot.sameFieldConflicts.length > 0 ||
      Object.keys(conflictQueueRef.current).length > 0
    ) {
      publishPendingQueueState();
      return;
    }
    const candidate = pending.model.peekNextQueued();
    if (candidate === null) {
      publishPendingQueueState();
      return;
    }
    const unit = candidate.unit;
    if (refreshBlocksPendingUnit(pending, unit)) {
      publishPendingQueueState();
      return;
    }
    const meta = pending.metaByUnitId.get(unit.id);
    if (meta === undefined) {
      const dispatch = pending.model.markDispatched(unit.id);
      if (dispatch !== null) {
        const settlement = pending.model.settleDispatched({
          ok: false,
          status: 0,
          error: {
            code: "pending_runtime_metadata_missing",
            message: "Queued edit metadata is missing.",
          },
        });
        if (settlement.outcome === "halted") {
          setRefreshError(settlement.halt.message);
        }
      }
      publishPendingQueueState();
      return;
    }

    const currentRow =
      unit.recordId === null
        ? rowsRef.current.find((row) => row.key === unit.rowKey)
        : (latestCommittedTimelineRow(unit.recordId) ??
          rowsRef.current.find((row) => row.recordId === unit.recordId));
    const dispatchPayload = materializePendingReplayPayload(unit, currentRow);
    if (dispatchPayload === null) {
      publishPendingQueueState();
      schedulePendingReplayRetry();
      return;
    }

    const dispatch = pending.model.markDispatched(unit.id);
    if (dispatch === null) {
      publishPendingQueueState();
      return;
    }
    const dispatchedUnit = dispatch.unit;
    publishPendingQueueState();
    trackPendingSocketTxn(dispatchedUnit.clientTxnId);

    let result: Awaited<
      ReturnType<typeof fetchJSON<TimelineMutationEnvelope>>
    > | null = null;
    try {
      recordWorkbookTiming("pending_fetch_start", {
        clientTxnId: unit.clientTxnId,
        kind: unit.kind,
        rowKey: unit.rowKey,
      });
      result = await fetchJSON<TimelineMutationEnvelope>(
        dispatchedUnit.path,
        {
          method: dispatchedUnit.method,
          body: JSON.stringify(dispatchPayload),
        },
        {
          onJSONParsed: () => {
            recordWorkbookTiming("pending_fetch_json_parsed", {
              clientTxnId: unit.clientTxnId,
              kind: unit.kind,
              rowKey: unit.rowKey,
            });
          },
          onResponse: (response) => {
            recordWorkbookTiming("pending_fetch_response", {
              clientTxnId: unit.clientTxnId,
              kind: unit.kind,
              rowKey: unit.rowKey,
              serverTiming: response.headers.get("server-timing") ?? "",
              status: response.status,
            });
          },
        },
      );
    } catch {
      resolvePendingSocketTxn(dispatchedUnit.clientTxnId);
      pending.model.settleDispatched({
        ok: false,
        status: 0,
        error: {
          code: "transport_failure",
          message: "Transport failure",
          retryable: true,
        },
      });
      publishPendingQueueState();
      schedulePendingReplayRetry();
      return;
    }

    if (!result.ok) {
      resolvePendingSocketTxn(dispatchedUnit.clientTxnId);
      const publicError = parsePendingReplayPublicError(result.payload);
      const settlement = pending.model.settleDispatched({
        ok: false,
        status: result.status,
        error: publicError,
      });
      if (settlement.outcome === "auth_paused") {
        setRefreshError(
          "Authentication required before queued edits can replay.",
        );
        publishPendingQueueState();
        scheduleAuthRecoveryProbe();
        return;
      }

      if (settlement.outcome === "same_field_conflict") {
        if (
          handleMutationConflict(
            result.payload,
            settlement.unit.rowKey,
            meta.focusField,
            meta.surface,
          )
        ) {
          pending.metaByUnitId.delete(settlement.unit.id);
          clearPendingSignatureForUnit(settlement.unit);
          publishPendingQueueState();
          return;
        }
        setRefreshError(
          publicError.message ?? parseErrorMessage(result.payload),
        );
        publishPendingQueueState();
        return;
      }

      if (settlement.outcome === "retryable_failure") {
        publishPendingQueueState();
        schedulePendingReplayRetry();
        return;
      }

      if (settlement.outcome === "halted") {
        setRefreshError(settlement.halt.message);
      } else {
        setRefreshError(parseErrorMessage(result.payload));
      }
      publishPendingQueueState();
      return;
    }

    recordWorkbookTiming("pending_result_apply_start", {
      clientTxnId: dispatchedUnit.clientTxnId,
      kind: dispatchedUnit.kind,
      rowKey: dispatchedUnit.rowKey,
    });
    let envelope: TimelineMutationEnvelope;
    let appliedRow: { record_id: string; row_version: number };
    try {
      envelope = readEnvelope<TimelineMutationEnvelope>(result.payload);
      const responseRow = normalizeTimelineFullRow(
        envelope.data.row,
        "pending mutation response row",
      );
      appliedRow = {
        record_id: responseRow.record_id,
        row_version: responseRow.row_version,
      };
      clearSubmittedScalarEditorDraftValuesForRow(
        dispatchedUnit.rowKey,
        meta.rowSnapshot.values,
      );
      applyRowMutation(dispatchedUnit.rowKey, envelope, {
        continueOnFreshDraft:
          meta.continueOnFreshDraft && meta.rowSnapshot.recordId === null,
        detectAutoResolution: meta.detectAutoResolution,
        promoteToCommittedRowInspect: meta.promoteToCommittedRowInspect,
        viewportContinuityToken: meta.viewportContinuityToken,
      });
    } catch (error) {
      recordWorkbookTiming("pending_result_apply_error", {
        clientTxnId: dispatchedUnit.clientTxnId,
        kind: dispatchedUnit.kind,
        message: error instanceof Error ? error.message : String(error),
        rowKey: dispatchedUnit.rowKey,
      });
      const settlement = pending.model.settleDispatched({
        ok: false,
        status: 0,
        error: {
          code: "client_apply_error",
          message: error instanceof Error ? error.message : String(error),
        },
      });
      if (settlement.outcome === "halted") {
        setRefreshError(settlement.halt.message);
      }
      publishPendingQueueState();
      return;
    }
    recordWorkbookTiming("pending_result_apply_end", {
      clientTxnId: dispatchedUnit.clientTxnId,
      kind: dispatchedUnit.kind,
      recordId: appliedRow.record_id,
      rowKey: dispatchedUnit.rowKey,
      rowVersion: appliedRow.row_version,
    });
    const successResult = {
      ok: true,
      row: appliedRow,
      ...(envelope.data.change_set_id === undefined
        ? {}
        : { change_set_id: envelope.data.change_set_id }),
    } as const;
    const settlement = pending.model.settleDispatched(successResult);
    if (settlement.outcome === "success") {
      pending.metaByUnitId.delete(settlement.unit.id);
      clearPendingSignatureForUnit(settlement.unit);
    }
    publishPendingQueueState();
    requestPendingReplay("unit_completed");
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
      if (
        closed ||
        reconnectTimer !== null ||
        socketLifecycleRef.current.reconnectSuppressed
      ) {
        return;
      }
      reconnectTimer = window.setTimeout(() => {
        reconnectTimer = null;
        connect();
      }, 1000);
    };
    socketReconnectAfterAuthRef.current = scheduleReconnect;

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
              presence: workbookPresence(
                currentPresenceRef.current,
                activeSheetRuntimeRef.current,
              ),
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
            presence: workbookPresence(
              currentPresenceRef.current,
              activeSheetRuntimeRef.current,
            ),
          },
        }),
      );
    };

    const requestSocketLifecycleRefresh = (
      options: Omit<LoadRowsOptions, "showLoading"> = {},
      refreshScope: PendingRefreshReplayBlockScope = { kind: "all" },
    ) => {
      beginRefreshInFlightRef.current(refreshScope);
      void loadRowsRef
        .current({ showLoading: false, ...options })
        .finally(() => {
          finishRefreshInFlightRef.current(refreshScope);
        });
    };

    const applySocketLifecycleEffects = (
      effects: readonly WorkbookSocketLifecycleEffect[],
      target: WebSocket,
    ) => {
      for (const effect of effects) {
        switch (effect.kind) {
          case "pause_for_auth_recovery":
            pendingQueueRef.current.model.pauseForAuthRecovery();
            setRefreshError(
              "Authentication required before queued edits can replay.",
            );
            publishPendingQueueStateRef.current();
            break;
          case "schedule_auth_recovery_probe":
            scheduleAuthRecoveryProbeRef.current();
            break;
          case "close_socket":
            target.close();
            break;
          case "request_refresh":
            if (
              effect.reason === "reset_required" ||
              effect.reason === "sequence_gap"
            ) {
              requestSocketLifecycleRefresh();
            }
            break;
          case "resume_pending_replay":
            pendingQueueRef.current.model.resumeAfterAuthRecovery();
            publishPendingQueueStateRef.current();
            schedulePendingReplayRef.current();
            break;
          case "apply_record_change":
          case "ignore_duplicate_sequence":
          case "suppress_reconnect":
            break;
        }
      }
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
        const effects = dispatchSocketLifecycleRef.current({
          type: "session_ack",
          messageType: message.type,
          ...(message.payload === undefined
            ? {}
            : { payload: message.payload }),
        });
        applySocketLifecycleEffects(effects, target);
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
        const effects = dispatchSocketLifecycleRef.current({
          type: "session_revoked",
        });
        applySocketLifecycleEffects(effects, target);
        return;
      }
      if (
        shouldIgnoreSelfOriginatedRecordChange(
          raw,
          resolvePendingSocketTxnRef.current,
        )
      ) {
        return;
      }
      if (!isRecordChangedMessage(raw)) {
        return;
      }
      const recordChangedPayload = message.payload as RecordChangedPayload;
      const streamEffects = dispatchSocketLifecycleRef.current({
        type: "record_changed_received",
        message: {
          ...(typeof message.stream_seq === "number"
            ? { stream_seq: message.stream_seq }
            : {}),
          payload: recordChangedPayload,
        },
      });
      for (const effect of streamEffects) {
        if (effect.kind === "ignore_duplicate_sequence") {
          return;
        }
        if (
          effect.kind === "request_refresh" &&
          effect.reason === "sequence_gap"
        ) {
          applySocketLifecycleEffects([effect], target);
          return;
        }
      }
      const viewportContinuityToken = beginViewportContinuityRef.current({
        kind: "scroll-only",
      });
      const applied = applyRecordChangedPatchRef.current(recordChangedPayload);
      const followupEffects = dispatchSocketLifecycleRef.current({
        type: "record_change_result",
        applied,
      });
      if (applied) {
        advanceViewportContinuityRef.current(viewportContinuityToken);
        return;
      }
      if (
        followupEffects.some(
          (effect) =>
            effect.kind === "request_refresh" &&
            effect.reason === "record_change_requery",
        )
      ) {
        requestSocketLifecycleRefresh(
          {
            viewportContinuityToken,
          },
          { kind: "record", recordId: recordChangedPayload.record_id },
        );
      }
    };

    const connect = () => {
      if (closed) {
        return;
      }
      socket = new WebSocket(changeSocketURL);
      activeSocketRef.current = socket;
      dispatchSocketLifecycleRef.current({ type: "socket_connecting" });
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
      socket.onclose = (event) => {
        if (
          event.code === 1008 &&
          (event.reason === "session_revoked" ||
            event.reason === "authorization_denied")
        ) {
          const effects = dispatchSocketLifecycleRef
            .current({
              type: "authorization_closed",
            })
            .filter((effect) => effect.kind !== "close_socket");
          applySocketLifecycleEffects(effects, socket as WebSocket);
        }
        scheduleReconnect();
      };
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
      socketReconnectAfterAuthRef.current = null;
      dispatchSocketLifecycleRef.current({ type: "socket_connecting" });
      socket?.close();
    };
  }, [changeSocketURL, incidentId]);

  const enqueueAutosaveReplayForPendingMutation = useCallback(
    ({
      clientTxnId,
      continueOnFreshDraft,
      detectAutoResolution,
      focusField,
      focusKey,
      mutationSignature,
      payloadIntent,
      promoteToCommittedRowInspect,
      rowKey,
      rowSnapshot,
      surface,
      viewportContinuityToken,
      visibleEdit,
    }: {
      clientTxnId: string;
      continueOnFreshDraft: boolean;
      detectAutoResolution: boolean;
      focusField: FocusFieldKey;
      focusKey: string;
      mutationSignature: string;
      payloadIntent: PendingReplayUnitInput["payloadIntent"];
      promoteToCommittedRowInspect: boolean;
      rowKey: string;
      rowSnapshot: WorkbookRow;
      surface: TimelineScalarEditorSurface;
      viewportContinuityToken: number;
      visibleEdit?: PendingReplayUnitInput["visibleEdit"];
    }) => {
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
        rowSnapshot.recordId === null
          ? apiPath(
              apiBase,
              `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`,
            )
          : apiPath(apiBase, `/api/v1/records/${rowSnapshot.recordId}`);
      const clientInstanceId =
        socketClientInstanceIdRef.current ?? tabClientInstanceId();
      socketClientInstanceIdRef.current = clientInstanceId;
      enqueuePendingReplayUnit({
        id: `pending-${clientTxnId}`,
        kind: rowSnapshot.recordId === null ? "create" : "patch",
        source: "autosave",
        incidentId,
        clientInstanceId,
        viewSchemaId: timelineViewSchemaId,
        rowKey,
        recordId: rowSnapshot.recordId,
        focusField,
        focusKey,
        surface,
        method: rowSnapshot.recordId === null ? "POST" : "PATCH",
        path: targetPath,
        payloadIntent,
        clientTxnId,
        mutationSignature,
        coalesceKey:
          rowSnapshot.recordId === null
            ? `draft:${rowKey}`
            : `record:${rowSnapshot.recordId}`,
        enqueueOrder: pendingReplayOrderRef.current,
        operationClass: "hot_path",
        status: "queued",
        ...(visibleEdit === undefined ? {} : { visibleEdit }),
        rowSnapshot,
        continueOnFreshDraft,
        detectAutoResolution,
        promoteToCommittedRowInspect,
        viewportContinuityToken,
      });
      pendingReplayOrderRef.current += 1;
    },
    [apiBase, enqueuePendingReplayUnit, incidentId],
  );

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
      const requestedRowSnapshot = rowsRef.current.find(
        (candidate) => candidate.key === rowKey,
      );
      const rowSnapshot =
        requestedRowSnapshot ?? createDraftRowForKey(rowKey) ?? undefined;
      const effectiveRowKey = rowSnapshot?.key ?? rowKey;
      const focusKey = inputFocusKey(
        effectiveRowKey,
        focusField,
        options.surface,
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
          : buildScalarPatchIntent(snapshot, clientTxnId);
      if (payload === null) {
        scalarDraftValuesRef.current.delete(focusKey);
        return;
      }

      const mutationSignature = buildStableMutationSignature(payload);
      if (
        pendingSignaturesRef.current.get(effectiveRowKey) === mutationSignature
      ) {
        return;
      }
      const visibleEdit =
        binding === undefined
          ? undefined
          : {
              rowKey: effectiveRowKey,
              fieldKey: binding.fieldKey,
              value: snapshot.values[focusField],
            };
      const viewportContinuityToken = beginViewportContinuity(
        options.preserveInputFocus
          ? {
              kind: "input",
              focusKey: inputFocusKey(
                effectiveRowKey,
                focusField,
                options.surface,
              ),
            }
          : {
              kind: "scroll-only",
            },
      );
      enqueueAutosaveReplayForPendingMutation({
        clientTxnId,
        continueOnFreshDraft: options.continueOnFreshDraft,
        detectAutoResolution: false,
        focusField,
        focusKey,
        mutationSignature,
        payloadIntent: payload,
        promoteToCommittedRowInspect: false,
        rowKey: effectiveRowKey,
        surface: options.surface,
        rowSnapshot: snapshot,
        viewportContinuityToken,
        visibleEdit,
      });
    },
    [
      beginViewportContinuity,
      enqueueAutosaveReplayForPendingMutation,
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
      source: "keyboard" | "blur" = "blur",
    ) => {
      const focusKey = inputFocusKey(rowKey, focusField, "grid");
      const rowSnapshot = rowsRef.current.find(
        (candidate) => candidate.key === rowKey,
      );
      if (!rowSnapshot) {
        return;
      }
      const draftValue =
        draftValueOverride ?? rowSnapshot.collectionDrafts[focusField];
      const priorKeyboardCommitValue =
        collectionKeyboardCommitRef.current.get(focusKey);
      if (source === "blur") {
        collectionKeyboardCommitRef.current.delete(focusKey);
        if (priorKeyboardCommitValue === draftValue) {
          return;
        }
      } else {
        collectionKeyboardCommitRef.current.set(focusKey, draftValue);
      }
      const snapshot =
        rowSnapshot.recordId === null
          ? rowSnapshot
          : (latestCommittedTimelineRow(rowSnapshot.recordId) ?? rowSnapshot);
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
          : buildCollectionPatchIntent(fieldKey, draftValue, clientTxnId);
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
      enqueueAutosaveReplayForPendingMutation({
        clientTxnId,
        continueOnFreshDraft: snapshot.recordId === null,
        detectAutoResolution: true,
        focusField,
        focusKey,
        mutationSignature,
        payloadIntent: payload,
        promoteToCommittedRowInspect: snapshot.recordId === null,
        rowKey,
        surface: "grid",
        rowSnapshot: effectiveSnapshot,
        viewportContinuityToken,
        visibleEdit: {
          rowKey,
          fieldKey,
          value: draftValue,
        },
      });
    },
    [
      beginViewportContinuity,
      enqueueAutosaveReplayForPendingMutation,
      latestCommittedTimelineRow,
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
          const idleRecord = await waitForCommittedRecordIdle(recordId);
          if (idleRecord === null) {
            clearViewportContinuity(viewportContinuityToken);
            finishSave("Conflict");
            return;
          }
          const body =
            action === "mark-reviewed"
              ? {
                  base_row_version: idleRecord.rowVersion,
                  client_txn_id: clientTxnId,
                  reason: "Reviewed from workbook",
                }
              : {
                  base_row_version: idleRecord.rowVersion,
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

          const envelope = readEnvelope<TimelineActionEnvelope>(result.payload);
          acceptTimelineActionResult(envelope.data);
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
      acceptTimelineActionResult,
      finishSave,
      nextClientTxnId,
      replacementDrafts,
      resolvePendingSocketTxn,
      trackPendingSocketTxn,
      waitForCommittedRecordIdle,
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
      let historyData: RecordHistoryData;
      try {
        const envelope = readEnvelope<RecordHistoryEnvelope>(result.payload);
        historyData = normalizeRecordHistoryData(envelope.data);
      } catch {
        setRowHistory({
          recordId,
          status: "error",
          data: null,
          message: "Invalid row history response.",
        });
        return null;
      }
      acceptTimelineRecordVersion(recordId, historyData.row_version);
      setRowHistory({
        recordId,
        status: "ready",
        data: historyData,
        message: null,
      });
      return historyData;
    },
    [acceptTimelineRecordVersion, apiBase],
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

  const submitRowHistoryMutation = useCallback(
    ({
      idleOptions,
      missingVersionMessage,
      onSuccess,
      recordId,
      request,
      viewportContinuityTarget,
    }: {
      idleOptions?: {
        readonly fallbackRowVersion?: number | null | undefined;
        readonly refreshIfMissing?: boolean;
      };
      missingVersionMessage: string;
      onSuccess: (
        payload: unknown,
        viewportContinuityToken: number,
      ) => Promise<void>;
      recordId: string;
      request: (
        baseRowVersion: number,
        clientTxnId: string,
      ) => Promise<{ ok: boolean; payload: unknown }>;
      viewportContinuityTarget: ViewportContinuityTarget;
    }) => {
      const clientTxnId = nextClientTxnId();
      const viewportContinuityToken = beginViewportContinuity(
        viewportContinuityTarget,
      );
      beginSave();
      setRowHistory((current) => ({ ...current, message: null }));
      saveQueueRef.current = saveQueueRef.current
        .catch(() => undefined)
        .then(async () => {
          const idleRecord = await waitForCommittedRecordIdle(
            recordId,
            idleOptions,
          );
          if (idleRecord === null) {
            clearViewportContinuity(viewportContinuityToken);
            setRowHistory((current) => ({
              ...current,
              message: missingVersionMessage,
            }));
            finishSave("Conflict");
            return;
          }
          trackPendingSocketTxn(clientTxnId);
          const result = await request(idleRecord.rowVersion, clientTxnId);
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
          await onSuccess(result.payload, viewportContinuityToken);
          finishSave("Saved");
        });
    },
    [
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      finishSave,
      nextClientTxnId,
      resolvePendingSocketTxn,
      trackPendingSocketTxn,
      waitForCommittedRecordIdle,
    ],
  );

  const submitRowHistoryDeleteRestore = useCallback(
    (operation: "delete" | "restore") => {
      const recordId = currentHistoryRecordId;
      if (recordId === null || recordId === undefined) {
        return;
      }
      const viewportContinuityTarget: ViewportContinuityTarget =
        selectedRow?.recordId === recordId
          ? { kind: "row-inspect", recordId }
          : { kind: "scroll-only" };
      submitRowHistoryMutation({
        idleOptions: {
          fallbackRowVersion:
            operation === "restore"
              ? rowHistory.data?.row_version
              : currentHistoryRowVersion,
          refreshIfMissing: operation !== "restore",
        },
        missingVersionMessage: "Missing row version for destructive action.",
        recordId,
        viewportContinuityTarget,
        request: (baseRowVersion, clientTxnId) => {
          const path =
            operation === "delete"
              ? `/api/v1/records/${recordId}`
              : `/api/v1/records/${recordId}/restore`;
          return fetchJSON<RecordDeleteRestoreEnvelope>(
            apiPath(apiBase, path),
            {
              method: operation === "delete" ? "DELETE" : "POST",
              body: JSON.stringify({
                base_row_version: baseRowVersion,
                client_txn_id: clientTxnId,
                reason:
                  operation === "delete"
                    ? "Deleted from workbook history"
                    : "Restored from workbook history",
              }),
            },
          );
        },
        onSuccess: async (payload, viewportContinuityToken) => {
          const envelope = readEnvelope<RecordDeleteRestoreEnvelope>(payload);
          acceptTimelineRecordVersion(recordId, envelope.data.row_version);
          await fetchRecordHistory(recordId);
          if (operation === "restore") {
            setSelectedRowId(recordId);
          }
          await loadRowsRef.current({
            showLoading: false,
            viewportContinuityToken,
          });
        },
      });
    },
    [
      apiBase,
      acceptTimelineRecordVersion,
      currentHistoryRecordId,
      currentHistoryRowVersion,
      fetchRecordHistory,
      rowHistory.data?.row_version,
      selectedRow?.recordId,
      submitRowHistoryMutation,
    ],
  );

  const submitRowHistoryRollback = useCallback(
    (item: RecordHistoryItem, action: RecordHistoryRollbackAction) => {
      const recordId = currentHistoryRecordId;
      if (recordId === null || recordId === undefined) {
        return;
      }
      const target = buildRecordRollbackTargetFromHistoryAction(item, action);
      if (target === null) {
        return;
      }
      const viewportContinuityTarget: ViewportContinuityTarget =
        selectedRow?.recordId === recordId
          ? { kind: "row-inspect", recordId }
          : { kind: "scroll-only" };
      submitRowHistoryMutation({
        idleOptions: {
          fallbackRowVersion: currentHistoryRowVersion,
        },
        missingVersionMessage: "Missing row version for rollback.",
        recordId,
        viewportContinuityTarget,
        request: (baseRowVersion, clientTxnId) =>
          fetchJSON<RecordRollbackEnvelope>(
            apiPath(apiBase, `/api/v1/records/${recordId}/rollback`),
            {
              method: "POST",
              body: JSON.stringify({
                base_row_version: baseRowVersion,
                client_txn_id: clientTxnId,
                reason: "Rollback from workbook history",
                target,
              }),
            },
          ),
        onSuccess: async (payload, viewportContinuityToken) => {
          const envelope = readEnvelope<RecordRollbackEnvelope>(payload);
          acceptTimelineRecordVersion(recordId, envelope.data.row_version);
          await fetchRecordHistory(recordId);
          await loadRowsRef.current({
            showLoading: false,
            viewportContinuityToken,
          });
        },
      });
    },
    [
      apiBase,
      acceptTimelineRecordVersion,
      currentHistoryRecordId,
      currentHistoryRowVersion,
      fetchRecordHistory,
      selectedRow?.recordId,
      submitRowHistoryMutation,
    ],
  );

  function submitMentionAction(
    mention: InspectorMention,
    action: MentionResolutionAction,
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
        const idleRecord = await waitForCommittedRecordIdle(recordId);
        if (idleRecord === null || idleRecord.row === null) {
          clearViewportContinuity(viewportContinuityToken);
          finishSave("Conflict");
          return;
        }
        const currentRow = idleRecord.row;
        if (action === "resolve_item" && resolvedRecordId === undefined) {
          const payload = buildMentionPatchPayload(
            currentRow,
            mention,
            action,
            clientTxnId,
            resolvedRecordId,
          );
          if (payload === null) {
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

          const envelope = readEnvelope<TimelineMutationEnvelope>(
            result.payload,
          );
          applyRowMutation(currentRow.key, envelope, {
            detectAutoResolution: false,
            viewportContinuityToken,
          });
          finishSave("Saved");
          try {
            await onRefreshEntities?.();
          } finally {
            settleViewportContinuityFollowup(viewportContinuityToken);
          }
          return;
        }

        const activeItem = [
          ...currentRow.collectionValues.hostRefs,
          ...currentRow.collectionValues.identityRefs,
        ].find((item) => item.itemRef === mention.itemRef);
        const currentMention =
          activeItem === undefined
            ? mention
            : {
                ...mention,
                entityType: activeItem.entityType,
                rawText: activeItem.rawText,
                resolvedRecordId: activeItem.resolvedRecordId,
                mentionRowVersion: activeItem.mentionRowVersion,
                resolutionMethod: activeItem.resolutionMethod,
                autoResolved: activeItem.autoResolved,
                displayText: activeItem.displayText,
                provenance: activeItem.provenance,
                confidence: activeItem.confidence,
                matchedAliasText: activeItem.matchedAliasText,
                anchor: {
                  ...mention.anchor,
                  targetEntityRecordId: activeItem.resolvedRecordId,
                },
              };
        const mentionID =
          currentMention.anchor.entityMentionId ??
          (currentMention.itemRef.startsWith("entity_mention:")
            ? currentMention.itemRef.slice("entity_mention:".length) || null
            : null);
        if (mentionID === null) {
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage("Missing entity mention identifier.");
          finishSave("Conflict");
          return;
        }
        const payload = buildMentionActionPayload(
          currentMention,
          action,
          clientTxnId,
          resolvedRecordId,
        );
        if (payload === null) {
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage("Missing mention row version.");
          finishSave("Conflict");
          return;
        }
        trackPendingSocketTxn(clientTxnId);
        const result = await fetchJSON<MentionActionEnvelope>(
          apiPath(apiBase, `/api/v1/entity-mentions/${mentionID}/resolve`),
          {
            method: "POST",
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

        const envelope = readEnvelope<MentionActionEnvelope>(result.payload);
        const entityMention = envelope.data.entity_mention;
        if (action === "dismiss_item") {
          setDismissedMentionsByRow((current) => {
            const rowMentions = current[recordId] ?? [];
            return {
              ...current,
              [recordId]: [
                ...rowMentions.filter(
                  (item) => item.itemRef !== currentMention.itemRef,
                ),
                {
                  rowRecordId: recordId,
                  fieldKey:
                    entityMention.source_field_key === "timeline.identity_refs"
                      ? "timeline.identity_refs"
                      : currentMention.fieldKey,
                  entityType:
                    entityMention.entity_type === "identity"
                      ? "identity"
                      : currentMention.entityType,
                  itemRef: currentMention.itemRef,
                  rawText: entityMention.raw_text || currentMention.rawText,
                  resolvedRecordId: currentMention.resolvedRecordId,
                  mentionRowVersion: entityMention.row_version,
                  resolutionMethod:
                    entityMention.resolution_method ??
                    currentMention.resolutionMethod,
                  autoResolved: currentMention.autoResolved,
                  displayText: currentMention.displayText,
                  priorTargetEntityRecordId:
                    currentMention.anchor.targetEntityRecordId ??
                    currentMention.priorTargetEntityRecordId ??
                    currentMention.resolvedRecordId,
                  provenance: currentMention.provenance,
                  confidence: currentMention.confidence,
                  matchedAliasText: currentMention.matchedAliasText,
                },
              ],
            };
          });
        }
        if (action === "revert_to_unresolved") {
          setDismissedMentionsByRow((current) => {
            const rowMentions = (current[recordId] ?? []).filter(
              (item) => item.itemRef !== currentMention.itemRef,
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

        await loadRowsRef.current({
          showLoading: false,
          viewportContinuityToken,
        });
        const restoreMentionActionFocus = () => {
          resolveViewportContinuityElement({
            kind: "row-inspect",
            recordId,
          })?.focus({ preventScroll: true });
        };
        restoreMentionActionFocus();
        window.requestAnimationFrame(restoreMentionActionFocus);
        window.setTimeout(restoreMentionActionFocus, 0);
        finishSave("Saved");
      });
  }

  const currentTimelineAnchorFor = useCallback(
    (rowKey: string, fieldKey: string): GridCellAnchor | null => {
      const row = rowsRef.current.find((candidate) => candidate.key === rowKey);
      if (row?.recordId === null || row?.recordId === undefined) {
        updateWorkbookFocusAnchor(null);
        return null;
      }
      const anchor = {
        fieldKey,
        recordId: row.recordId,
      };
      updateTimelineFocusAnchor(anchor.recordId, anchor.fieldKey);
      return anchor;
    },
    [updateTimelineFocusAnchor, updateWorkbookFocusAnchor],
  );

  const navigateTimelineFocusAnchor = useCallback(
    (current: GridCellAnchor, intent: GridNavigationIntent) => {
      const nextAnchor = navigateGridCellAnchor({
        columns: timelineAnchorColumnsRef.current,
        current,
        intent,
        presentationRows: buildGridPresentationRows({
          getGroupLabel: (row, fieldKey) => timelineGroupLabel(row, fieldKey),
          groupBy: queryState.groupBy,
          rows: timelineAnchorRowsRef.current,
        }),
      });
      if (nextAnchor === null) {
        updateWorkbookFocusAnchor(null);
        return;
      }
      updateTimelineFocusAnchor(nextAnchor.recordId, nextAnchor.fieldKey);
      window.setTimeout(() => {
        restoreTimelineFocusAnchor(nextAnchor);
      }, 0);
    },
    [
      queryState.groupBy,
      restoreTimelineFocusAnchor,
      updateTimelineFocusAnchor,
      updateWorkbookFocusAnchor,
    ],
  );

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

  const handleTimelineCommonKeyboardCommand = useCallback(
    (command: WorkbookKeyboardCommand, anchor: GridCellAnchor | null) => {
      if (anchor === null) {
        return false;
      }
      if (command.kind === "open-history") {
        openRowHistory(anchor.recordId);
        return true;
      }
      if (command.kind === "close-inspector") {
        setSelectedRowId(null);
        setSelectedMentionRef(null);
        setInspectorMessage(null);
        setRowHistory({
          recordId: null,
          status: "idle",
          data: null,
          message: null,
        });
        restoreTimelineFocusAnchor(anchor);
        return true;
      }
      if (command.kind === "preview-linked-evidence") {
        setSelectedRowId(anchor.recordId);
        setInspectorMessage(
          "Linked evidence preview is unavailable for this row.",
        );
        restoreTimelineFocusAnchor(anchor);
        return true;
      }
      return false;
    },
    [openRowHistory, restoreTimelineFocusAnchor],
  );

  const handleKeyDown = useCallback(
    (
      event: ReactKeyboardEvent<HTMLInputElement | HTMLTextAreaElement>,
      rowKey: string,
      focusField: keyof RowValues,
      surface: TimelineScalarEditorSurface,
    ) => {
      const priorGridAnchor = workbookFocusAnchorRef.current;
      if (
        surface === "inspector" &&
        event.key === "Escape" &&
        priorGridAnchor?.surface === timelineViewSchemaId
      ) {
        event.preventDefault();
        restoreTimelineFocusAnchor(priorGridAnchor);
        return;
      }
      const binding = Object.values(timelineScalarBindingIndex).find(
        (candidate) => candidate.key === focusField,
      );
      const fieldKey = binding?.fieldKey ?? focusField;
      const anchor = currentTimelineAnchorFor(rowKey, fieldKey);
      const command = mapWorkbookKeyboardCommand(event, {
        closeInspector:
          anchor !== null &&
          (surface === "inspector" ||
            selectedRowId !== null ||
            rowHistory.recordId !== null ||
            rowHistory.status !== "idle"),
        history: anchor !== null,
        previewLinkedEvidence:
          anchor !== null && event.currentTarget.value === "",
      });
      if (command.preventDefault) {
        event.preventDefault();
      }
      if (
        command.kind === "navigate" &&
        anchor === null &&
        (command.intent.key === "Enter" || command.intent.key === "Tab")
      ) {
        recordWorkbookTiming("key_commit_accepted", {
          field: focusField,
          key: command.intent.key,
          rowKey,
          surface,
        });
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
        return;
      }
      if (command.kind === "navigate" && anchor !== null) {
        if (command.intent.key === "Enter" || command.intent.key === "Tab") {
          recordWorkbookTiming("key_commit_accepted", {
            field: focusField,
            key: command.intent.key,
            recordId: anchor.recordId,
            rowKey,
            surface,
          });
          queueScalarSave(
            rowKey,
            focusField,
            {
              continueOnFreshDraft: true,
              preserveInputFocus: false,
              surface,
            },
            event.currentTarget.value,
          );
        }
        navigateTimelineFocusAnchor(anchor, command.intent);
        return;
      }
      if (handleTimelineCommonKeyboardCommand(command, anchor)) {
        return;
      }
    },
    [
      currentTimelineAnchorFor,
      handleTimelineCommonKeyboardCommand,
      navigateTimelineFocusAnchor,
      queueScalarSave,
      rowHistory.recordId,
      rowHistory.status,
      restoreTimelineFocusAnchor,
      selectedRowId,
    ],
  );

  const handleCollectionKeyDown = useCallback(
    (
      event: ReactKeyboardEvent<HTMLInputElement>,
      rowKey: string,
      fieldKey: CollectionFieldKey,
      draftKey: CollectionDraftKey,
    ) => {
      const anchor = currentTimelineAnchorFor(rowKey, fieldKey);
      const command = mapWorkbookKeyboardCommand(event, {
        closeInspector:
          anchor !== null &&
          (selectedRowId !== null ||
            rowHistory.recordId !== null ||
            rowHistory.status !== "idle"),
        history: anchor !== null,
        previewLinkedEvidence:
          anchor !== null && event.currentTarget.value === "",
        quickLink:
          anchor !== null &&
          (fieldKey === "timeline.host_refs" ||
            fieldKey === "timeline.identity_refs"),
      });
      if (command.preventDefault) {
        event.preventDefault();
      }
      if (command.kind === "navigate" && anchor !== null) {
        if (command.intent.key === "Enter" || command.intent.key === "Tab") {
          queueCollectionSave(
            rowKey,
            fieldKey,
            draftKey,
            event.currentTarget.value,
            "keyboard",
          );
        }
        navigateTimelineFocusAnchor(anchor, command.intent);
        return;
      }
      if (handleTimelineCommonKeyboardCommand(command, anchor)) {
        return;
      }
      if (command.kind === "quick-link" && anchor !== null) {
        const row = rowsRef.current.find(
          (candidate) => candidate.recordId === anchor.recordId,
        );
        const mention =
          row === undefined
            ? undefined
            : [
                ...row.collectionValues.hostRefs,
                ...row.collectionValues.identityRefs,
              ].find((item) => item.itemKind !== "resolved_ref");
        if (mention !== undefined) {
          setSelectedRowId(anchor.recordId);
          setSelectedMentionRef(mention.itemRef);
          setInspectorMessage(null);
        } else {
          setSelectedRowId(anchor.recordId);
          setInspectorMessage(
            "No unresolved mention is available for quick link.",
          );
        }
        return;
      }
      if (event.key === "Enter" || event.key === "Tab") {
        event.preventDefault();
        queueCollectionSave(
          rowKey,
          fieldKey,
          draftKey,
          event.currentTarget.value,
          "keyboard",
        );
      }
    },
    [
      currentTimelineAnchorFor,
      handleTimelineCommonKeyboardCommand,
      navigateTimelineFocusAnchor,
      queueCollectionSave,
      rowHistory.recordId,
      rowHistory.status,
      selectedRowId,
    ],
  );

  const handlePaste = useCallback(
    (
      event: ReactClipboardEvent<HTMLInputElement | HTMLTextAreaElement>,
      rowKey: string,
      focusField: keyof RowValues,
      surface: TimelineScalarEditorSurface,
    ) => {
      const clipboardText = event.clipboardData?.getData("text/plain") ?? "";
      const binding = Object.values(timelineScalarBindingIndex).find(
        (candidate) => candidate.key === focusField,
      );
      const fieldKey = binding?.fieldKey ?? focusField;
      if (
        surface === "grid" &&
        binding !== undefined &&
        clipboardTextLooksTabular(clipboardText)
      ) {
        const anchor = currentTimelineAnchorFor(rowKey, fieldKey);
        if (anchor !== null) {
          const dimensions = clipboardGridDimensions(clipboardText);
          const presentationRows = buildGridPresentationRows({
            getGroupLabel: (row, groupFieldKey) =>
              timelineGroupLabel(row, groupFieldKey),
            groupBy: queryState.groupBy,
            rows: timelineAnchorRowsRef.current,
          });
          const targetResolution = resolveGridPasteTargets({
            columns: timelineAnchorColumnsRef.current,
            current: anchor,
            pastedColumnCount: dimensions.columnCount,
            pastedRowCount: dimensions.rowCount,
            presentationRows,
          });
          if (targetResolution !== null) {
            event.preventDefault();
            const clientTxnId = nextClientTxnId();
            const viewportContinuityToken = beginViewportContinuity({
              kind: "input",
              focusKey: inputFocusKey(rowKey, focusField, surface),
            });
            beginSave();
            saveQueueRef.current = saveQueueRef.current
              .catch(() => undefined)
              .then(async () => {
                const rowTargetPayload: Array<
                  | { readonly kind: "create" }
                  | {
                      readonly base_row_version: number;
                      readonly kind: "record";
                      readonly record_id: string;
                    }
                > = [];
                for (const target of targetResolution.rowTargets) {
                  if (target.kind === "create") {
                    rowTargetPayload.push({ kind: "create" });
                    continue;
                  }
                  const idleRecord = await waitForCommittedRecordIdle(
                    target.recordId,
                  );
                  if (idleRecord === null) {
                    clearViewportContinuity(viewportContinuityToken);
                    finishSave("Conflict");
                    return;
                  }
                  rowTargetPayload.push({
                    kind: "record",
                    record_id: target.recordId,
                    base_row_version: idleRecord.rowVersion,
                  });
                }
                trackPendingSocketTxn(clientTxnId);
                const result = await fetchJSON<TimelineClipboardPasteEnvelope>(
                  apiPath(
                    apiBase,
                    `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/clipboard-paste`,
                  ),
                  {
                    method: "POST",
                    body: JSON.stringify({
                      view_schema_id: timelineViewSchemaId,
                      client_txn_id: clientTxnId,
                      clipboard_text: clipboardText,
                      format: clipboardText.includes("\t") ? "tsv" : "csv",
                      start_field_key: fieldKey,
                      columns: targetResolution.columns,
                      targets: rowTargetPayload,
                    }),
                  },
                );
                resolvePendingSocketTxn(clientTxnId);
                if (!result.ok) {
                  clearViewportContinuity(viewportContinuityToken);
                  finishSave("Conflict");
                  return;
                }
                const envelope = readEnvelope<TimelineClipboardPasteEnvelope>(
                  result.payload,
                );
                const pasteConflictKeys: string[] = [];
                for (const conflict of envelope.data.conflicts ?? []) {
                  const conflictBinding = timelineScalarBindingForField(
                    conflict.field_key,
                  );
                  const queueKey = sameFieldConflictQueueKey(conflict);
                  pasteConflictKeys.push(queueKey);
                  registerSameFieldConflict(
                    conflict,
                    inputFocusKey(
                      conflict.record_id,
                      conflictBinding?.key ?? focusField,
                      "grid",
                    ),
                    "grid",
                  );
                }
                if (pasteConflictKeys.length > 1) {
                  setPasteConflictGroup({ keys: pasteConflictKeys });
                  setActiveConflictKey(pasteConflictKeys[0] ?? null);
                } else if (pasteConflictKeys.length === 0) {
                  setPasteConflictGroup(null);
                }
                await loadRowsRef.current({
                  showLoading: false,
                  viewportContinuityToken,
                });
                restoreTimelineFocusAnchor(anchor);
                finishSave(
                  envelope.data.conflicts && envelope.data.conflicts.length > 0
                    ? "Conflict"
                    : "Saved",
                );
              });
            return;
          }
        }
      }
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
    [
      apiBase,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      currentTimelineAnchorFor,
      finishSave,
      incidentId,
      nextClientTxnId,
      queryState.groupBy,
      queueScalarSave,
      registerSameFieldConflict,
      resolvePendingSocketTxn,
      restoreTimelineFocusAnchor,
      setScalarEditorDraftValue,
      trackPendingSocketTxn,
      waitForCommittedRecordIdle,
    ],
  );

  const sendPresenceUpdate = useCallback(
    (presence: typeof currentPresence) => {
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
            presence: workbookPresence(presence, activeSheetRef),
          },
        }),
      );
    },
    [activeSheetRef],
  );

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
        rowsRef.current.find((candidate) => candidate.key === row.key) ?? row;
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
        throw new Error(evidencePublicErrorMessage(createEvidence.payload));
      }
      const evidenceEnvelope = readEnvelope<ViewMutationEnvelope>(
        createEvidence.payload,
      );
      const evidenceRecord = evidenceEnvelope.data.row;

      await createAndAttachEvidenceBlob({
        apiBase,
        attachClientTxnId: nextClientTxnId,
        baseRowVersion: evidenceRecord.row_version,
        createClientTxnId: nextClientTxnId,
        evidenceRecordId: evidenceRecord.record_id,
        file,
        incidentId,
      });
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
            const effectiveSnapshot =
              snapshot.recordId === null
                ? snapshot
                : (await waitForCommittedRecordIdle(snapshot.recordId))?.row;
            if (effectiveSnapshot === null || effectiveSnapshot === undefined) {
              throw new Error("invalid_timeline_row");
            }
            const evidenceRecordId = await createAndAttachEvidenceFile(file);
            const clientTxnId = nextClientTxnId();
            const payload =
              effectiveSnapshot.recordId === null
                ? buildAttachedEvidenceCreatePayload(
                    evidenceRecordId,
                    clientTxnId,
                  )
                : buildAttachedEvidencePatchPayload(
                    effectiveSnapshot,
                    evidenceRecordId,
                    clientTxnId,
                  );
            if (payload === null) {
              throw new Error("invalid_timeline_row");
            }

            trackPendingSocketTxn(clientTxnId);
            const targetPath =
              effectiveSnapshot.recordId === null
                ? apiPath(
                    apiBase,
                    `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`,
                  )
                : apiPath(
                    apiBase,
                    `/api/v1/records/${effectiveSnapshot.recordId}`,
                  );
            const result = await fetchJSON<TimelineMutationEnvelope>(
              targetPath,
              {
                method: effectiveSnapshot.recordId === null ? "POST" : "PATCH",
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
            applyRowMutation(effectiveSnapshot.key, envelope, {
              continueOnFreshDraft: effectiveSnapshot.recordId === null,
              promoteToCommittedRowInspect: effectiveSnapshot.recordId === null,
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
      waitForCommittedRecordIdle,
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
      const dataTestId = timelineScalarEditorTestId({
        fieldKey: binding.fieldKey,
        recordId: row.recordId,
        surface,
      });
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
            onFocusAnchor={updateTimelineFocusAnchor}
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
      updateTimelineFocusAnchor,
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
      const collectionInputTestId =
        row.recordId === null
          ? draftTimelineCollectionInputTestId(binding.fieldKey)
          : timelineCollectionInputTestId(row.recordId, binding.fieldKey);
      return (
        <>
          <div
            data-testid={
              row.recordId === null
                ? draftRelationshipItemsTestId(binding.fieldKey)
                : relationshipItemsTestId(row.recordId, binding.fieldKey)
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
            data-testid={collectionInputTestId}
            key={`${row.key}:${binding.draftKey}:${row.rowVersion ?? "draft"}`}
            ref={(element) => {
              registerInput(
                row.key,
                binding.draftKey,
                "grid",
                collectionInputTestId,
                element,
              );
            }}
            style={gridCellInputStyle}
            type="text"
            defaultValue={row.collectionDrafts[binding.draftKey]}
            onChange={(event) => {
              const focusKey = inputFocusKey(row.key, binding.draftKey, "grid");
              if (
                collectionKeyboardCommitRef.current.get(focusKey) !==
                event.currentTarget.value
              ) {
                collectionKeyboardCommitRef.current.delete(focusKey);
              }
            }}
            onBlur={(event) => {
              queueCollectionSave(
                row.key,
                binding.fieldKey,
                binding.draftKey,
                event.currentTarget.value,
              );
            }}
            onFocus={() => {
              updateTimelineFocusAnchor(row.recordId, binding.fieldKey);
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
      updateTimelineFocusAnchor,
    ],
  );

  const timelineColumns = useMemo<readonly GridColumn<WorkbookRow>[]>(
    () =>
      timelineVisibleBindings.map(
        (binding): GridColumn<WorkbookRow> => ({
          fieldKey: binding.fieldKey,
          headerTestId: gridSortHeaderTestId(
            timelineViewSchemaId,
            binding.fieldKey,
          ),
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
              const countDisplay = buildEvidenceCountDisplayViewModel({
                projectedCount: readCellValue(row.rawRow, binding.fieldKey),
                projectedHasEvidence: readCellValue(
                  row.rawRow,
                  "timeline.has_evidence",
                ),
              });
              return (
                <span
                  data-evidence-count-state={countDisplay.stateKey}
                  style={timelineEvidenceCellStyle}
                >
                  <span
                    data-testid={
                      row.recordId === null
                        ? undefined
                        : rowCellTestId(row.recordId, binding.fieldKey)
                    }
                  >
                    {countDisplay.displayCount}
                  </span>
                  {row.recordId === null ? null : (
                    <span
                      data-testid={rowCellTestId(
                        row.recordId,
                        "timeline.has_evidence",
                      )}
                      style={
                        countDisplay.hasEvidence
                          ? timelineEvidenceFlagOnStyle
                          : timelineEvidenceFlagOffStyle
                      }
                      title={
                        countDisplay.hasEvidence
                          ? "Timeline row has evidence"
                          : "Timeline row has no evidence"
                      }
                    >
                      {String(countDisplay.hasEvidence)}
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
    [
      renderTimelineCollectionInput,
      renderTimelineGridEditor,
      timelineBindingLabel,
    ],
  );

  const timelineRowGutter = useMemo(
    () => ({
      label: "",
      width: 58,
      minWidth: 58,
    }),
    [],
  );

  const timelineActionsColumn = useMemo<GridActionsColumn<WorkbookRow>>(
    () => ({
      headerTestId: gridActionsHeaderTestId(timelineViewSchemaId),
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
                data-testid={rowHistoryOpenButtonTestId(row.recordId ?? "")}
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
              data-testid={timelineRowMarkReviewedButtonTestId(row.recordId)}
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
              data-testid={timelineRowReplacementInputTestId(row.recordId)}
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
              data-testid={timelineRowSupersedeButtonTestId(row.recordId)}
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
      rows.map((row, index) => {
        const rowPresence = presenceForRow(row.recordId);
        const rowPresenceVisible = visiblePresence(rowPresence, 2);
        const ordinal = row.recordId === null ? "+" : String(index + 1);
        return {
          key: row.key,
          recordId: row.recordId,
          data: row,
          gutterContent: (
            <span style={rowGutterContentStyle}>
              <span aria-hidden="true">{ordinal}</span>
              {rowPresence.length > 0 ? (
                <span
                  aria-label={`${rowPresence
                    .map((presence) => presence.display_name)
                    .join(", ")} focused on this row`}
                  data-testid={rowPresenceMarkerTestId(row.recordId ?? "draft")}
                  role="img"
                  style={rowGutterPresenceStyle}
                >
                  {rowPresenceVisible.shown
                    .map((presence) => displayInitials(presence.display_name))
                    .join("")}
                  {rowPresenceVisible.overflow > 0
                    ? `+${rowPresenceVisible.overflow}`
                    : ""}
                </span>
              ) : null}
            </span>
          ),
          gutterLabel: ordinal,
          gutterTestId:
            row.recordId === null
              ? undefined
              : gridRowGutterTestId(timelineViewSchemaId, row.recordId),
          onSelect: () => {
            if (row.recordId) {
              handleSelectRow(row.recordId);
            }
          },
          selected: row.recordId !== null && row.recordId === selectedRowId,
          testId:
            row.recordId === null
              ? undefined
              : gridRowTestId(timelineViewSchemaId, row.recordId),
          variant: row.recordId === null ? "draft" : "default",
        };
      }),
    [handleSelectRow, presenceForRow, rows, selectedRowId],
  );

  useLayoutEffect(() => {
    timelineAnchorColumnsRef.current = timelineColumns;
    timelineAnchorRowsRef.current = timelineGridRows;
  }, [timelineColumns, timelineGridRows]);

  const getTimelineGroupLabel = useCallback(
    (row: WorkbookRow, fieldKey: string) => timelineGroupLabel(row, fieldKey),
    [],
  );
  const getTimelineGroupRowTestId = useCallback(
    (fieldKey: string, value: string) =>
      gridGroupRowTestId(timelineViewSchemaId, fieldKey, value),
    [],
  );

  function renderInspectorFieldEditors(row: WorkbookRow) {
    return (
      <section
        data-testid={timelineInspectorSectionTestId("details")}
        style={inspectorSectionStyle}
      >
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
        ? timelineDraftEvidenceFileInputTestId()
        : timelineEvidenceFileInputTestId(row.recordId);
    const countDisplay = buildEvidenceCountDisplayViewModel({
      projectedCount: readCellValue(row.rawRow, "timeline.evidence_count"),
      projectedHasEvidence: readCellValue(row.rawRow, "timeline.has_evidence"),
    });
    return (
      <section
        data-testid={timelineInspectorSectionTestId("evidence")}
        data-evidence-count-state={countDisplay.stateKey}
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
        <div
          data-testid={
            row.recordId === null
              ? timelineDraftEvidenceAttachSectionTestId()
              : timelineEvidenceAttachSectionTestId(row.recordId)
          }
        >
          <h3 style={sectionTitleStyle}>Evidence</h3>
          {row.recordId !== null ? (
            <p style={bodyStyle}>
              Attached evidence count: {countDisplay.displayCount}
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
                handleTimelineEvidenceFiles(
                  row,
                  event.currentTarget.files ?? [],
                );
                event.currentTarget.value = "";
              }}
            />
          </label>
        </div>
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
      <section
        data-testid={timelineInspectorSectionTestId("history")}
        style={inspectorSectionStyle}
      >
        <div data-testid={rowHistoryPanelTestId()}>
          <div style={historySectionHeaderStyle}>
            <h3 style={sectionTitleStyle}>Row history</h3>
            {selectedActiveRow !== null ? (
              <button
                data-testid={rowHistoryOpenInspectorButtonTestId(
                  selectedActiveRow.recordId ?? "",
                )}
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
              data-testid={rowHistoryOpenSelectedButtonTestId()}
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
            <p data-testid={rowHistoryLoadingTestId()} style={bodyStyle}>
              Loading history...
            </p>
          ) : null}
          {rowHistory.message ? (
            <p
              data-testid={rowHistoryMessageTestId()}
              style={genericErrorTextStyle}
            >
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
                    data-testid={rowHistoryDeleteButtonTestId()}
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
                    data-testid={rowHistoryRestoreButtonTestId()}
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
                {historyData.items.map((item) => {
                  const itemAnchor = {
                    historyItemRef: item.history_item_ref,
                  };
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
                          data-testid={rowHistoryActionTestId({
                            ...itemAnchor,
                            action,
                          })}
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
                      data-testid={rowHistoryItemTestId(itemAnchor)}
                      key={item.history_item_ref}
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
        </div>
      </section>
    );
  }

  const activeConflict =
    activeConflictKey === null
      ? null
      : (conflictQueue[activeConflictKey] ?? null);
  const activePasteConflictKeys =
    pasteConflictGroup?.keys.filter((key) => conflictQueue[key]) ?? [];
  const activePasteConflictIndex =
    activeConflictKey === null
      ? -1
      : activePasteConflictKeys.indexOf(activeConflictKey);
  const showPasteConflictNavigator =
    activePasteConflictKeys.length > 1 && activePasteConflictIndex >= 0;

  useEffect(() => {
    if (activeConflict === null) {
      return;
    }
    window.setTimeout(() => {
      (
        document.querySelector(
          dataTestIdSelector("conflict-resolver-summary"),
        ) as HTMLElement | null
      )?.focus();
    }, 0);
  }, [activeConflict]);

  const closeConflictResolver = useCallback(
    (conflict: LocalConflictState) => {
      setActiveConflictKey(null);
      restoreConflictFocus(conflict.focusKey);
    },
    [restoreConflictFocus],
  );

  useEffect(() => {
    if (!activeConflict) {
      return;
    }
    const handleConflictResolverEscape = (event: KeyboardEvent) => {
      if (event.key !== "Escape") {
        return;
      }
      event.preventDefault();
      closeConflictResolver(activeConflict);
    };
    document.addEventListener("keydown", handleConflictResolverEscape);
    return () => {
      document.removeEventListener("keydown", handleConflictResolverEscape);
    };
  }, [activeConflict, closeConflictResolver]);

  const clearLocalConflict = useCallback(
    (conflict: LocalConflictState) => {
      pendingQueueRef.current.model.clearSameFieldConflict(conflict.key);
      setConflictQueueState((current) => {
        const next = { ...current };
        delete next[conflict.key];
        updateSaveStateForConflicts(next);
        return next;
      });
      setActiveConflictKey((current) =>
        current === conflict.key ? null : current,
      );
      setPasteConflictGroup((current) => {
        if (current === null || !current.keys.includes(conflict.key)) {
          return current;
        }
        const keys = current.keys.filter((key) => key !== conflict.key);
        return keys.length > 1 ? { keys } : null;
      });
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
      setSaveStateSecondaryMessage("Workbook edits are syncing.");
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
              setSaveStateSecondaryMessage("Conflict requires review.");
              return;
            }
            setSaveState("Conflict");
            setSaveStateSecondaryMessage("Conflict requires review.");
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
    <section
      data-testid={timelineMutationSubstrateReadyTestId()}
      style={workbookStyle}
    >
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

      {autoResolutionNotices.length > 0 ? (
        <aside style={noticeStackStyle}>
          {autoResolutionNotices.map((notice) => (
            <div
              key={notice.itemRef}
              data-testid={autoResolutionNoticeTestId(notice.itemRef)}
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
                  data-testid={autoResolutionUndoButtonTestId(notice.itemRef)}
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
                          mentionRowVersion: activeItem.mentionRowVersion,
                          resolutionMethod: activeItem.resolutionMethod,
                          autoResolved: activeItem.autoResolved,
                          status: "resolved",
                          chipState: mentionChipStateForItem(activeItem),
                          anchor: {
                            recordId: row.recordId,
                            fieldKey: notice.fieldKey,
                            itemRef: activeItem.itemRef,
                            entityMentionId: activeItem.itemRef.startsWith(
                              "entity_mention:",
                            )
                              ? activeItem.itemRef.slice(
                                  "entity_mention:".length,
                                ) || null
                              : null,
                            targetEntityRecordId: activeItem.resolvedRecordId,
                          },
                          sourceKind: "entity_mention",
                          isActiveRelationshipValue: true,
                          priorTargetEntityRecordId: null,
                          displayText: activeItem.displayText,
                          provenance: activeItem.provenance,
                          confidence: activeItem.confidence,
                          matchedAliasText: activeItem.matchedAliasText,
                        },
                        "revert_to_unresolved",
                      );
                    }
                  }}
                >
                  Undo
                </button>
                <button
                  data-testid={autoResolutionReviewButtonTestId(notice.itemRef)}
                  style={secondaryActionButtonStyle}
                  type="button"
                  onClick={() => {
                    handleSelectMention(notice.rowRecordId, notice.itemRef);
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
          <WorkbookShellSlotRegion
            slot="view-bar"
            style={viewBarStyle}
            viewSchemaId={timelineViewSchemaId}
          >
            {savedViewSelector}
            <WorkbookGridControls
              contract={timelineContract}
              filterDraft={filterDraft}
              onApplyFilter={applyQueryFilter}
              onClearAll={() => {
                setQueryState(emptyWorkbookQueryState());
                setFilterDraft(defaultFilterDraft(timelineContract));
              }}
              onFilterDraftChange={setFilterDraft}
              onGroupByChange={handleQueryGroupByChange}
              onRemoveFilter={(fieldKey) => {
                setQueryState((current) =>
                  removeFilterField(current, fieldKey),
                );
              }}
              queryState={queryState}
              surface={timelineViewSchemaId}
            />
          </WorkbookShellSlotRegion>
          <WorkbookShellSlotRegion
            slot="primary-grid"
            viewSchemaId={timelineViewSchemaId}
          >
            <GridViewport
              ref={gridShellRef}
              style={timelineGridShellStyle}
              testId={gridShellTestId(timelineViewSchemaId)}
            >
              <GridTable
                actionsColumn={timelineActionsColumn}
                columns={timelineColumns}
                getGroupLabel={getTimelineGroupLabel}
                getGroupRowTestId={getTimelineGroupRowTestId}
                groupBy={queryState.groupBy}
                onToggleSort={handleQuerySortToggle}
                rowGutter={timelineRowGutter}
                rows={timelineGridRows}
                sort={queryState.sort}
              />
              <div aria-hidden="true" style={visuallyHiddenStyle}>
                {rows.map((row) => (
                  <Fragment key={`${row.key}-metadata`}>
                    <span
                      data-testid={
                        row.recordId === null
                          ? draftCellTestId("timeline.capture_state")
                          : rowCellTestId(
                              row.recordId,
                              "timeline.capture_state",
                            )
                      }
                    >
                      {row.captureState}
                    </span>
                    <span
                      data-testid={
                        row.recordId === null
                          ? draftCellTestId("row_version")
                          : rowCellTestId(row.recordId, "row_version")
                      }
                    >
                      {row.rowVersion ?? "new"}
                    </span>
                  </Fragment>
                ))}
              </div>
            </GridViewport>
          </WorkbookShellSlotRegion>
          {activeConflict ? (
            <section
              data-testid="conflict-resolver"
              data-conflict-record-id={activeConflict.anchor.record_id}
              data-conflict-field-key={activeConflict.anchor.field_key}
              data-conflict-base-row-version={
                activeConflict.anchor.base_row_version
              }
              data-conflict-current-row-version={
                activeConflict.anchor.current_row_version
              }
              style={conflictResolverStyle}
              aria-label="Same-field conflict resolver"
              onKeyDown={(event) => {
                if (event.key === "Escape") {
                  event.preventDefault();
                  event.stopPropagation();
                  closeConflictResolver(activeConflict);
                  return;
                }
                const summary = event.currentTarget.querySelector(
                  dataTestIdSelector("conflict-resolver-summary"),
                );
                if (event.key === "Enter" && event.target === summary) {
                  event.preventDefault();
                  event.stopPropagation();
                }
              }}
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
              {showPasteConflictNavigator ? (
                <nav
                  aria-label="Paste conflict navigator"
                  data-testid="paste-conflict-navigator"
                  style={noticeCardStyle}
                >
                  <p data-testid="paste-conflict-position" style={bodyStyle}>
                    {activePasteConflictIndex + 1} of{" "}
                    {activePasteConflictKeys.length}
                  </p>
                  <div style={inlineButtonRowStyle}>
                    <button
                      data-testid="paste-conflict-previous"
                      disabled={activePasteConflictIndex <= 0}
                      onClick={() => {
                        const previousKey =
                          activePasteConflictKeys[activePasteConflictIndex - 1];
                        if (previousKey) {
                          setActiveConflictKey(previousKey);
                        }
                      }}
                      style={secondaryActionButtonStyle}
                      type="button"
                    >
                      Previous
                    </button>
                    <button
                      data-testid="paste-conflict-next"
                      disabled={
                        activePasteConflictIndex >=
                        activePasteConflictKeys.length - 1
                      }
                      onClick={() => {
                        const nextKey =
                          activePasteConflictKeys[activePasteConflictIndex + 1];
                        if (nextKey) {
                          setActiveConflictKey(nextKey);
                        }
                      }}
                      style={secondaryActionButtonStyle}
                      type="button"
                    >
                      Next
                    </button>
                  </div>
                  <div style={inlineButtonRowStyle}>
                    {activePasteConflictKeys.map((key, index) => {
                      const queued = conflictQueue[key];
                      if (!queued) {
                        return null;
                      }
                      return (
                        <button
                          aria-current={
                            key === activeConflictKey ? "true" : undefined
                          }
                          data-testid={pasteConflictItemTestId(key)}
                          key={key}
                          onClick={() => setActiveConflictKey(key)}
                          style={
                            key === activeConflictKey
                              ? actionButtonStyle
                              : secondaryActionButtonStyle
                          }
                          type="button"
                        >
                          {index + 1}.{" "}
                          {timelineBindingLabel(queued.conflict.field_key)}
                        </button>
                      );
                    })}
                  </div>
                </nav>
              ) : null}
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
                    closeConflictResolver(activeConflict);
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

        <WorkbookShellSlotRegion
          slot="inspector"
          viewSchemaId={timelineViewSchemaId}
        >
          <aside
            aria-label="Timeline inspector"
            data-testid="timeline-inspector"
            style={inspectorShellStyle}
          >
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
            </div>
            {selectedRow?.recordId ? (
              <>
                {renderInspectorFieldEditors(selectedRow)}
                {renderEvidenceAttachSection(selectedRow)}
                {renderRowHistorySection()}
                <section
                  data-testid={timelineInspectorSectionTestId("relationships")}
                  style={inspectorSectionStyle}
                >
                  <h3 style={sectionTitleStyle}>Mentions</h3>
                  <div style={mentionGroupStyle}>
                    {["unresolved", "resolved", "dismissed"].map((status) => {
                      const group = inspectorMentions.filter(
                        (item) => item.status === status,
                      );
                      return (
                        <div key={status} style={mentionGroupColumnStyle}>
                          <p
                            data-density-role="narrow-metadata"
                            style={groupLabelStyle}
                          >
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
                                data-testid={mentionItemTestId(item.itemRef)}
                                tabIndex={0}
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
                        <dd style={detailValueStyle}>
                          {selectedMention.status}
                        </dd>
                      </div>
                      <div>
                        <dt style={detailTermStyle}>Target</dt>
                        <dd style={detailValueStyle}>
                          {selectedMention.resolvedRecordId
                            ? relationshipItemLabel(
                                selectedMention,
                                entityIndex,
                              )
                            : "None"}
                        </dd>
                      </div>
                    </dl>

                    {selectedMention.status === "unresolved" ? (
                      <div style={inspectorActionStackStyle}>
                        <label style={labelStyle}>
                          Resolve to existing
                          <select
                            data-testid={mentionResolveTargetSelectTestId()}
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
                            data-testid={mentionResolveExistingButtonTestId()}
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
                            data-testid={mentionCreateEntityButtonTestId(
                              selectedMention.entityType,
                            )}
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
                      <div style={inspectorActionStackStyle}>
                        {canManageMentions ? (
                          <label style={labelStyle}>
                            Correct target
                            <select
                              data-testid={mentionResolveTargetSelectTestId()}
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
                        ) : null}
                        <div style={inlineButtonRowStyle}>
                          {canManageMentions ? (
                            <button
                              data-testid={mentionResolveExistingButtonTestId()}
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
                              Correct target
                            </button>
                          ) : null}
                          {canManageMentions ? (
                            <button
                              data-testid={mentionDismissButtonTestId()}
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
                              data-testid={mentionRestoreUnresolvedButtonTestId()}
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
                      </div>
                    ) : null}

                    {selectedMention.status === "dismissed" ? (
                      <div style={inlineButtonRowStyle}>
                        <button
                          data-testid={mentionRestoreUnresolvedButtonTestId()}
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
                  Pick a saved row to inspect unresolved, resolved, and
                  dismissed mentions.
                </p>
                {inspectorMessage ? (
                  <p data-testid="timeline-inspector-message" style={bodyStyle}>
                    {inspectorMessage}
                  </p>
                ) : null}
              </>
            )}
          </aside>
        </WorkbookShellSlotRegion>
      </div>

      <WorkbookShellSlotRegion
        slot="status-strip"
        style={statusStripStyle}
        viewSchemaId={timelineViewSchemaId}
      >
        <span style={statusStripItemStyle}>
          <span aria-hidden="true" style={statusIconStyle(saveState)} />
          <strong
            aria-live="polite"
            aria-label="Save state"
            data-density-role="narrow-metadata"
            data-testid={saveStateTestId()}
            role="status"
          >
            {saveState}
          </strong>
        </span>
        {saveStateSecondaryMessage !== null ? (
          <span style={statusStripSecondaryItemStyle}>
            {saveStateSecondaryMessage}
          </span>
        ) : null}
        <span style={statusStripItemStyle}>
          Queue{" "}
          <span data-testid={statusStripQueueCountTestId()}>
            {pendingQueueSnapshot.queuedCount +
              pendingQueueSnapshot.inFlightCount}
          </span>
        </span>
        <div
          aria-label={`${activeSheetPresenceRecords.length} collaborators present on this sheet`}
          data-testid="presence-header"
          role="status"
          style={statusStripPresenceStyle}
        >
          <span>Presence</span>
          {headerPresence.shown.length === 0 ? (
            <span style={presenceEmptyStyle}>0</span>
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
        <span style={statusStripSpacerStyle} />
        <span style={statusStripMutedItemStyle}>{incidentId}</span>
        <WorkbookFocusAnchorStatus anchor={workbookFocusAnchor} />
      </WorkbookShellSlotRegion>
    </section>
  );
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
  const [timelinePreviewRows, setTimelinePreviewRows] = useState<WorkbookRow[]>(
    [],
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
      try {
        validateTimelineViewSchemaId(
          envelope.data.view_schema_id,
          "timeline preview query response",
        );
      } catch {
        setTimelinePreviewRows([]);
        return;
      }
      const draftKey = entityType === "host" ? "hostRefs" : "identityRefs";
      let previewRows: WorkbookRow[];
      try {
        previewRows = envelope.data.rows
          .map((row, index) =>
            rowFromApi(
              normalizeTimelineFullRow(
                row,
                `timeline preview query rows[${index}]`,
              ),
            ),
          )
          .filter((row) =>
            row.collectionValues[draftKey].some(
              (item) => item.resolvedRecordId === recordId,
            ),
          );
      } catch {
        setTimelinePreviewRows([]);
        return;
      }
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
      try {
        validateTimelineViewSchemaId(
          envelope.data.view_schema_id,
          "assessment support query response",
        );
        setSupportRows(
          envelope.data.rows.map((row, index) =>
            normalizeTimelineFullRow(
              row,
              `assessment support query rows[${index}]`,
            ),
          ),
        );
      } catch {
        setSupportRows([]);
      }
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

  const refreshReferenceOptions = useCallback(async () => {
    setReferenceLoadError(null);
    const targetViewSchemaIds = [
      timelineViewSchemaId,
      hostsViewSchemaId,
      identitiesViewSchemaId,
      partiesViewSchemaId,
      taskRequestsViewSchemaId,
      decisionsViewSchemaId,
      evidenceViewSchemaId,
      notesViewSchemaId,
      findingsViewSchemaId,
      investigativeQueriesViewSchemaId,
      forensicKeywordsViewSchemaId,
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
      hosts: genericReferenceOptionsFromRows(
        hostsViewSchemaId,
        rowsByView[hostsViewSchemaId] ?? [],
      ),
      identities: genericReferenceOptionsFromRows(
        identitiesViewSchemaId,
        rowsByView[identitiesViewSchemaId] ?? [],
      ),
      notes: genericReferenceOptionsFromRows(
        notesViewSchemaId,
        rowsByView[notesViewSchemaId] ?? [],
      ),
      timeline: genericReferenceOptionsFromRows(
        timelineViewSchemaId,
        rowsByView[timelineViewSchemaId] ?? [],
      ),
      noteSourceRecords: [],
      allRecords: [],
    };
    next.noteSourceRecords = [
      ...next.timeline,
      ...next.hosts,
      ...next.identities,
      ...next.evidence,
    ];
    next.allRecords = [
      ...next.timeline,
      ...next.hosts,
      ...next.identities,
      ...next.evidence,
      ...next.notes,
      ...genericReferenceOptionsFromRows(
        findingsViewSchemaId,
        rowsByView[findingsViewSchemaId] ?? [],
      ),
      ...genericReferenceOptionsFromRows(
        investigativeQueriesViewSchemaId,
        rowsByView[investigativeQueriesViewSchemaId] ?? [],
      ),
      ...genericReferenceOptionsFromRows(
        forensicKeywordsViewSchemaId,
        rowsByView[forensicKeywordsViewSchemaId] ?? [],
      ),
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
    </section>
  );
}

function GenericMultiSelectControl({
  ariaLabel,
  id,
  options,
  testId,
  value,
  onChange,
}: {
  readonly ariaLabel: string;
  readonly id: string | undefined;
  readonly options: readonly {
    readonly label: string;
    readonly value: string;
  }[];
  readonly testId: string;
  readonly value: string;
  readonly onChange: (value: string) => void;
}) {
  return (
    <select
      aria-label={ariaLabel}
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
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
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
  const controlLabel = `${field.label} value`;
  if (field.writeKind === "action_payload") {
    if (collectionMode === "remove") {
      return (
        <GenericMultiSelectControl
          ariaLabel={controlLabel}
          id={id}
          options={collectionItems.map((item) => ({
            label: item.displayText,
            value: item.itemRef,
          }))}
          testId={testId}
          value={value}
          onChange={onChange}
        />
      );
    }

    const options = referenceOptionsForField(field, referenceOptions);
    if (genericFieldUsesReferenceOptions(field)) {
      return (
        <GenericMultiSelectControl
          ariaLabel={controlLabel}
          id={id}
          options={options.map((option) => ({
            label: option.label,
            value: option.recordId,
          }))}
          testId={testId}
          value={value}
          onChange={onChange}
        />
      );
    }

    return (
      <textarea
        aria-label={controlLabel}
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
        aria-label={controlLabel}
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
        aria-label={controlLabel}
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

  if (field.readKind === "boolean") {
    return (
      <input
        aria-label={controlLabel}
        data-testid={testId}
        id={id}
        style={inputStyle}
        type="checkbox"
        checked={value === "true"}
        onChange={(event) => {
          onChange(event.target.checked ? "true" : "false");
        }}
      />
    );
  }

  if (field.readKind === "number") {
    return (
      <input
        aria-label={controlLabel}
        data-testid={testId}
        id={id}
        style={inputStyle}
        type="number"
        value={value}
        onChange={(event) => {
          onChange(event.target.value);
        }}
      />
    );
  }

  if (isMultilineGenericField(field)) {
    return (
      <textarea
        aria-label={controlLabel}
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
      aria-label={controlLabel}
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
      const labels = collectionItemLabels(items);
      if (labels.length > 0) {
        return labels.join(", ");
      }
      return "None";
    }
  }
  return JSON.stringify(value);
}

function genericCellLabelForField(
  surface: WorkbookSurface,
  fieldKey: string,
  value: unknown,
): string {
  if (
    surface === evidenceViewSchemaId &&
    fieldKey === "evidence.storage_ref" &&
    typeof value === "string" &&
    /^object:\/\/[0-9a-f-]+$/iu.test(value.trim())
  ) {
    return "Managed object";
  }
  return genericCellLabel(value);
}

function collectionItemLabels(items: readonly unknown[]): string[] {
  return items.flatMap((item) => {
    if (!item || typeof item !== "object" || Array.isArray(item)) {
      return [];
    }
    const raw = item as Record<string, unknown>;
    const candidates = [
      raw.display_text,
      raw.alias_text,
      raw.tag_name,
      raw.raw_text,
      raw.linked_record_id,
      raw.record_id,
      raw.item_ref,
    ];
    const label = candidates.find(
      (value): value is string =>
        typeof value === "string" && value.trim() !== "",
    );
    return label === undefined ? [] : [label];
  });
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
    const payloadValue = genericDirectPayloadValue(field, value);
    if (payloadValue === invalidGenericPayloadValue) {
      return null;
    }
    payload[field.fieldKey] = payloadValue;
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
  const payloadValue =
    value === "" && field.clearable
      ? null
      : genericDirectPayloadValue(field, value);
  if (payloadValue === invalidGenericPayloadValue) {
    return null;
  }
  return {
    field_key: field.fieldKey,
    value: payloadValue,
  };
}

const invalidGenericPayloadValue = Symbol("invalid generic payload value");

function genericDirectPayloadValue(
  field: ViewFieldContract,
  value: string,
): string | number | boolean | typeof invalidGenericPayloadValue {
  if (field.readKind === "number") {
    if (!/^-?\d+$/u.test(value)) {
      return invalidGenericPayloadValue;
    }
    const parsed = Number.parseInt(value, 10);
    return Number.isSafeInteger(parsed) ? parsed : invalidGenericPayloadValue;
  }
  if (field.readKind === "boolean") {
    if (value === "true") {
      return true;
    }
    if (value === "false") {
      return false;
    }
    return invalidGenericPayloadValue;
  }
  return value;
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
      return mode === "remove"
        ? { op: "remove_tag", item_ref: value }
        : { op: "add_tag", tag_name: value };
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
    case findingsViewSchemaId:
      return has("finding.statement");
    case investigativeQueriesViewSchemaId:
      return (
        has("investigative_query.platform") &&
        has("investigative_query.purpose") &&
        has("investigative_query.query_text")
      );
    case forensicKeywordsViewSchemaId:
      return has("forensic_keyword.pattern") && has("forensic_keyword.reason");
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
    case findingsViewSchemaId:
      return "Statement is required.";
    case investigativeQueriesViewSchemaId:
      return "Platform, purpose, and query text are required.";
    case forensicKeywordsViewSchemaId:
      return "Pattern and reason are required.";
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
      (field.fieldKey === "task.owner_user_id" ||
        field.fieldKey === "decision.owner_user_id" ||
        field.fieldKey === "finding.owner_user_id" ||
        field.fieldKey === "status_review.review_owner_user_id" ||
        field.fieldKey === "lesson.owner_user_id")
    ) {
      draft[field.fieldKey] = currentUserId;
    }
    if (field.fieldKey === "finding.kind") {
      draft[field.fieldKey] = "finding";
    }
    if (field.fieldKey === "finding.state") {
      draft[field.fieldKey] = "open";
    }
    if (field.fieldKey === "forensic_keyword.match_mode") {
      draft[field.fieldKey] = "literal";
    }
    if (field.fieldKey === "forensic_keyword.case_sensitive") {
      draft[field.fieldKey] = "false";
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
    hosts: [],
    identities: [],
    notes: [],
    timeline: [],
    noteSourceRecords: [],
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

function partyLinkPairsForContract(contract: ViewContract): PartyLinkPair[] {
  const hasField = (fieldKey: string) => Boolean(contract.fieldMap[fieldKey]);
  const pairs: PartyLinkPair[] = [];
  if (
    hasField("evidence.collector_party_text") &&
    hasField("evidence.collector_party_id")
  ) {
    pairs.push({
      key: "evidence.collector_party_text:evidence.collector_party_id",
      label: "Collector",
      textFieldKey: "evidence.collector_party_text",
      refFieldKey: "evidence.collector_party_id",
    });
  }
  if (
    hasField("evidence.source_party_text") &&
    hasField("evidence.source_party_id")
  ) {
    pairs.push({
      key: "evidence.source_party_text:evidence.source_party_id",
      label: "Source",
      textFieldKey: "evidence.source_party_text",
      refFieldKey: "evidence.source_party_id",
    });
  }
  if (
    hasField("task.requester_party_text") &&
    hasField("task.requester_party_id")
  ) {
    pairs.push({
      key: "task.requester_party_text:task.requester_party_id",
      label: "Requester",
      textFieldKey: "task.requester_party_text",
      refFieldKey: "task.requester_party_id",
    });
  }
  return pairs;
}

function extractEmailFromPartyText(value: string): string | null {
  const match = value.match(/[^\s<>@]+@[^\s<>@]+/u);
  return match?.[0] ?? null;
}

function genericRowLabel(contract: ViewContract, row: EntityApiRow): string {
  const preferredFieldKeys = [
    "timeline.summary",
    "host.display_name",
    "host.hostname",
    "identity.display_name",
    "identity.upn",
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
    "finding.statement",
    "investigative_query.purpose",
    "investigative_query.query_text",
    "forensic_keyword.pattern",
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
    case "decision.affected_record_ids":
    case "finding.supporting_refs":
    case "finding.contradictory_refs":
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
    case "decision.affected_record_ids":
    case "finding.supporting_refs":
    case "finding.contradictory_refs":
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

function genericCollectionSupportsRemove(_fieldKey: string): boolean {
  return true;
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

function SystemViewSwitcher({
  activeViewSchemaId,
  onSelect,
}: {
  readonly activeViewSchemaId: string;
  readonly onSelect: (viewSchemaId: string) => void;
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(0);
  const containerRef = useRef<HTMLFieldSetElement | null>(null);
  const blurCloseTimerRef = useRef<number | null>(null);
  const deferredFocusTimerRef = useRef<number | null>(null);
  const triggerRef = useRef<HTMLButtonElement | null>(null);
  const optionRefs = useRef(new Map<string, HTMLButtonElement>());
  const activeSystemEntryIndex = systemViewSwitcherEntries.findIndex(
    (entry) => entry.viewSchemaId === activeViewSchemaId,
  );
  const activeSystemEntry =
    activeSystemEntryIndex === -1
      ? null
      : (systemViewSwitcherEntries[activeSystemEntryIndex] ?? null);

  const clearBlurCloseTimer = useCallback(() => {
    if (blurCloseTimerRef.current !== null) {
      window.clearTimeout(blurCloseTimerRef.current);
      blurCloseTimerRef.current = null;
    }
  }, []);

  const clearDeferredFocusTimer = useCallback(() => {
    if (deferredFocusTimerRef.current !== null) {
      window.clearTimeout(deferredFocusTimerRef.current);
      deferredFocusTimerRef.current = null;
    }
  }, []);

  useEffect(
    () => () => {
      clearBlurCloseTimer();
      clearDeferredFocusTimer();
    },
    [clearBlurCloseTimer, clearDeferredFocusTimer],
  );

  const deferFocus = useCallback(
    (focusTarget: () => HTMLElement | null) => {
      clearDeferredFocusTimer();
      deferredFocusTimerRef.current = window.setTimeout(() => {
        deferredFocusTimerRef.current = null;
        focusTarget()?.focus({ preventScroll: true });
      }, 0);
    },
    [clearDeferredFocusTimer],
  );

  const focusOption = useCallback(
    (index: number) => {
      const entry = systemViewSwitcherEntries[index];
      if (entry === undefined) {
        return;
      }
      deferFocus(() => optionRefs.current.get(entry.viewSchemaId) ?? null);
    },
    [deferFocus],
  );

  const scheduleBlurClose = useCallback(() => {
    clearBlurCloseTimer();
    blurCloseTimerRef.current = window.setTimeout(() => {
      blurCloseTimerRef.current = null;
      const activeElement = document.activeElement;
      if (
        activeElement instanceof Node &&
        containerRef.current?.contains(activeElement)
      ) {
        return;
      }
      setIsOpen(false);
    }, 0);
  }, [clearBlurCloseTimer]);

  const openMenu = useCallback(() => {
    clearBlurCloseTimer();
    const nextIndex =
      activeSystemEntryIndex === -1 ? 0 : activeSystemEntryIndex;
    setActiveIndex(nextIndex);
    setIsOpen(true);
    focusOption(nextIndex);
  }, [activeSystemEntryIndex, clearBlurCloseTimer, focusOption]);

  const closeMenu = useCallback(
    (options: { readonly restoreTriggerFocus: boolean }) => {
      clearBlurCloseTimer();
      clearDeferredFocusTimer();
      setIsOpen(false);
      if (options.restoreTriggerFocus) {
        deferFocus(() => triggerRef.current);
      }
    },
    [clearBlurCloseTimer, clearDeferredFocusTimer, deferFocus],
  );

  const handleInternalPointerDown = useCallback(() => {
    clearBlurCloseTimer();
  }, [clearBlurCloseTimer]);

  const moveOptionFocus = useCallback(
    (nextIndex: number) => {
      const optionCount = systemViewSwitcherEntries.length;
      if (optionCount === 0) {
        return;
      }
      const wrappedIndex = (nextIndex + optionCount) % optionCount;
      setActiveIndex(wrappedIndex);
      focusOption(wrappedIndex);
    },
    [focusOption],
  );

  const selectOption = useCallback(
    (viewSchemaId: string) => {
      if (viewSchemaId === "") {
        return;
      }
      clearBlurCloseTimer();
      clearDeferredFocusTimer();
      setIsOpen(false);
      onSelect(viewSchemaId);
    },
    [clearBlurCloseTimer, clearDeferredFocusTimer, onSelect],
  );

  const handleOptionKeyDown = useCallback(
    (event: ReactKeyboardEvent<HTMLButtonElement>, index: number) => {
      switch (event.key) {
        case "ArrowDown":
          event.preventDefault();
          moveOptionFocus(index + 1);
          break;
        case "ArrowUp":
          event.preventDefault();
          moveOptionFocus(index - 1);
          break;
        case "Home":
          event.preventDefault();
          moveOptionFocus(0);
          break;
        case "End":
          event.preventDefault();
          moveOptionFocus(systemViewSwitcherEntries.length - 1);
          break;
        case "Escape":
          event.preventDefault();
          closeMenu({ restoreTriggerFocus: true });
          break;
        case "Enter":
          event.preventDefault();
          selectOption(systemViewSwitcherEntries[index]?.viewSchemaId ?? "");
          break;
        default:
          break;
      }
    },
    [closeMenu, moveOptionFocus, selectOption],
  );

  return (
    <fieldset
      aria-label="System view switcher"
      ref={containerRef}
      style={systemViewSwitcherStyle}
      onPointerDownCapture={handleInternalPointerDown}
      onBlur={(event) => {
        const nextFocus = event.relatedTarget;
        if (
          nextFocus instanceof Node &&
          containerRef.current?.contains(nextFocus)
        ) {
          return;
        }
        scheduleBlurClose();
      }}
    >
      <button
        ref={triggerRef}
        aria-controls={isOpen ? systemViewSwitcherMenuTestId() : undefined}
        aria-expanded={isOpen}
        aria-haspopup="menu"
        aria-label="System views"
        data-testid={systemViewSwitcherTriggerTestId()}
        data-view-schema-id={activeSystemEntry?.viewSchemaId ?? ""}
        style={systemViewSwitcherTriggerStyle}
        type="button"
        onClick={() => {
          if (isOpen) {
            closeMenu({ restoreTriggerFocus: false });
            return;
          }
          openMenu();
        }}
        onKeyDown={(event) => {
          if (event.key === "Enter" || event.key === " ") {
            event.preventDefault();
            openMenu();
          }
          if (event.key === "Escape" && isOpen) {
            event.preventDefault();
            closeMenu({ restoreTriggerFocus: true });
          }
        }}
      >
        <span>System views</span>
        <span aria-hidden="true" style={systemViewSwitcherValueStyle}>
          {activeSystemEntry?.contract.title ?? "Select view"}
        </span>
      </button>
      {isOpen ? (
        <div
          data-testid={systemViewSwitcherMenuTestId()}
          id={systemViewSwitcherMenuTestId()}
          role="menu"
          style={systemViewSwitcherMenuStyle}
        >
          {systemWorkbookSurfaceGroups.map((group) => (
            <fieldset
              key={group.token}
              aria-label={group.label}
              data-testid={systemViewSwitcherGroupTestId(group.token)}
              style={systemViewSwitcherGroupStyle}
            >
              <legend style={systemViewSwitcherGroupLabelStyle}>
                {group.label}
              </legend>
              {group.entries.map((entry) => {
                const optionIndex = systemViewSwitcherEntries.findIndex(
                  (option) => option.viewSchemaId === entry.viewSchemaId,
                );
                const isSelected = entry.viewSchemaId === activeViewSchemaId;
                return (
                  <button
                    key={entry.viewSchemaId}
                    ref={(node) => {
                      if (node === null) {
                        optionRefs.current.delete(entry.viewSchemaId);
                        return;
                      }
                      optionRefs.current.set(entry.viewSchemaId, node);
                    }}
                    aria-checked={isSelected}
                    data-testid={systemViewSwitcherOptionTestId(
                      group.token,
                      entry.viewSchemaId,
                    )}
                    data-view-schema-id={entry.viewSchemaId}
                    role="menuitemradio"
                    style={{
                      ...systemViewSwitcherOptionStyle,
                      ...(isSelected
                        ? systemViewSwitcherOptionSelectedStyle
                        : null),
                    }}
                    tabIndex={optionIndex === activeIndex ? 0 : -1}
                    type="button"
                    onMouseDown={(event) => {
                      event.preventDefault();
                    }}
                    onClick={() => {
                      selectOption(entry.viewSchemaId);
                    }}
                    onKeyDown={(event) => {
                      handleOptionKeyDown(event, optionIndex);
                    }}
                  >
                    {entry.contract.title}
                  </button>
                );
              })}
            </fieldset>
          ))}
        </div>
      ) : null}
    </fieldset>
  );
}

function ActiveSurfaceSavedViewSelector({
  activeViewSchemaId,
  currentIncidentRole,
  currentUserId,
  savedViews,
  selectedSheetRef,
  onCreateSavedView,
  onDeleteSavedView,
  onDuplicateSavedView,
  onSelectBaseSurface,
  onSelectSavedView,
  onSetDefaultSheetRef,
  onSetHomeSheetRef,
  onUpdateSavedView,
}: {
  readonly activeViewSchemaId: string;
  readonly currentIncidentRole: IncidentRole | null;
  readonly currentUserId: string | null;
  readonly savedViews: readonly SavedViewResource[];
  readonly selectedSheetRef: WorkbookSheetRef;
  readonly onCreateSavedView: (input: {
    readonly displayName: string;
    readonly scope: "private" | "shared";
  }) => Promise<SavedViewResource>;
  readonly onDeleteSavedView: (savedView: SavedViewResource) => Promise<void>;
  readonly onDuplicateSavedView: (
    savedView: SavedViewResource,
  ) => Promise<SavedViewResource>;
  readonly onSelectBaseSurface: (viewSchemaId: string) => void;
  readonly onSelectSavedView: (savedView: SavedViewResource) => void;
  readonly onSetDefaultSheetRef: () => Promise<void>;
  readonly onSetHomeSheetRef: () => Promise<void>;
  readonly onUpdateSavedView: (
    savedView: SavedViewResource,
    input: {
      readonly displayName: string;
      readonly scope: "private" | "shared";
    },
  ) => Promise<SavedViewResource>;
}) {
  const [displayName, setDisplayName] = useState("Saved view");
  const [scope, setScope] = useState<"private" | "shared">("private");
  const [status, setStatus] = useState("");
  const activeSavedViews = useMemo(
    () =>
      savedViews.filter(
        (savedView) => savedView.view_schema_id === activeViewSchemaId,
      ),
    [activeViewSchemaId, savedViews],
  );
  const selectedSavedViewId =
    selectedSheetRef.kind === "saved_view" &&
    activeSavedViews.some(
      (savedView) => savedView.saved_view_id === selectedSheetRef.id,
    )
      ? selectedSheetRef.id
      : "";
  const selectedSavedView =
    activeSavedViews.find(
      (savedView) => savedView.saved_view_id === selectedSavedViewId,
    ) ?? null;
  const selectedSavedViewMutable = canMutateSavedView(
    selectedSavedView,
    currentUserId,
    currentIncidentRole,
  );
  const trimmedDisplayName = displayName.trim();

  useEffect(() => {
    if (selectedSavedView === null) {
      setDisplayName("Saved view");
      setScope("private");
      return;
    }
    setDisplayName(selectedSavedView.display_name);
    setScope(selectedSavedView.scope === "shared" ? "shared" : "private");
  }, [selectedSavedView]);

  const runSavedViewAction = async (
    action: () => Promise<void>,
    successMessage: string,
  ) => {
    setStatus("");
    try {
      await action();
      setStatus(successMessage);
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "Saved view failed.");
    }
  };

  return (
    <div style={savedViewControlGroupStyle}>
      <label style={savedViewSelectorFrameStyle}>
        <span style={savedViewSelectorLabelStyle}>View:</span>
        <select
          aria-label="Saved view"
          data-active-view-schema-id={activeViewSchemaId}
          data-selected-saved-view-id={selectedSavedViewId}
          data-selected-sheet-ref-kind={
            selectedSavedViewId === "" ? "view_schema" : "saved_view"
          }
          data-testid={savedViewSelectorTestId(activeViewSchemaId)}
          style={savedViewSelectStyle}
          value={selectedSavedViewId}
          onChange={(event) => {
            const nextSavedViewId = event.currentTarget.value;
            setStatus("");
            if (nextSavedViewId === "") {
              onSelectBaseSurface(activeViewSchemaId);
              return;
            }
            const savedView = activeSavedViews.find(
              (candidate) => candidate.saved_view_id === nextSavedViewId,
            );
            if (savedView !== undefined) {
              onSelectSavedView(savedView);
            }
          }}
        >
          <option value="">Base view</option>
          {activeSavedViews.map((savedView) => (
            <option
              key={savedView.saved_view_id}
              data-saved-view-id={savedView.saved_view_id}
              data-testid={savedViewOptionTestId(
                activeViewSchemaId,
                savedView.saved_view_id,
              )}
              data-view-schema-id={activeViewSchemaId}
              value={savedView.saved_view_id}
            >
              {savedView.display_name}
            </option>
          ))}
        </select>
      </label>
      <input
        aria-label="Saved view name"
        data-testid={savedViewNameInputTestId(activeViewSchemaId)}
        style={savedViewNameInputStyle}
        type="text"
        value={displayName}
        onChange={(event) => {
          setDisplayName(event.currentTarget.value);
        }}
      />
      <select
        aria-label="Saved view scope"
        data-testid={savedViewScopeSelectTestId(activeViewSchemaId)}
        style={savedViewScopeSelectStyle}
        value={scope}
        onChange={(event) => {
          setScope(
            event.currentTarget.value === "shared" ? "shared" : "private",
          );
        }}
      >
        <option value="private">Private</option>
        <option value="shared">Shared</option>
      </select>
      <button
        data-testid={savedViewCreateButtonTestId(activeViewSchemaId)}
        disabled={trimmedDisplayName === ""}
        style={secondaryActionButtonStyle}
        type="button"
        onClick={() => {
          void runSavedViewAction(async () => {
            await onCreateSavedView({
              displayName: trimmedDisplayName,
              scope,
            });
          }, "Saved view created.");
        }}
      >
        Create
      </button>
      {selectedSavedView ? (
        <>
          <button
            data-testid={savedViewDuplicateButtonTestId(
              activeViewSchemaId,
              selectedSavedView.saved_view_id,
            )}
            style={secondaryActionButtonStyle}
            type="button"
            onClick={() => {
              void runSavedViewAction(async () => {
                await onDuplicateSavedView(selectedSavedView);
              }, "Saved view duplicated.");
            }}
          >
            Duplicate
          </button>
          <button
            data-testid={savedViewUpdateButtonTestId(
              activeViewSchemaId,
              selectedSavedView.saved_view_id,
            )}
            disabled={!selectedSavedViewMutable || trimmedDisplayName === ""}
            style={secondaryActionButtonStyle}
            type="button"
            onClick={() => {
              void runSavedViewAction(async () => {
                await onUpdateSavedView(selectedSavedView, {
                  displayName: trimmedDisplayName,
                  scope,
                });
              }, "Saved view updated.");
            }}
          >
            Update
          </button>
          <button
            data-testid={savedViewDeleteButtonTestId(
              activeViewSchemaId,
              selectedSavedView.saved_view_id,
            )}
            disabled={!selectedSavedViewMutable}
            style={secondaryActionButtonStyle}
            type="button"
            onClick={() => {
              void runSavedViewAction(async () => {
                await onDeleteSavedView(selectedSavedView);
              }, "Saved view deleted.");
            }}
          >
            Delete
          </button>
        </>
      ) : null}
      <button
        data-testid={savedViewSetHomeButtonTestId(activeViewSchemaId)}
        style={secondaryActionButtonStyle}
        type="button"
        onClick={() => {
          void runSavedViewAction(onSetHomeSheetRef, "Home view updated.");
        }}
      >
        Home
      </button>
      <button
        data-testid={savedViewSetDefaultButtonTestId(activeViewSchemaId)}
        disabled={currentIncidentRole !== "admin"}
        style={secondaryActionButtonStyle}
        type="button"
        onClick={() => {
          void runSavedViewAction(
            onSetDefaultSheetRef,
            "Default view updated.",
          );
        }}
      >
        Default
      </button>
      <span
        aria-live="polite"
        data-testid={savedViewStatusTestId(activeViewSchemaId)}
        style={savedViewStatusStyle}
      >
        {status}
      </span>
    </div>
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
  const [surface, setSurface] = useState<SurfaceKind>(initialViewSchemaID);
  const [startupSheetRef, setStartupSheetRef] = useState<WorkbookSheetRef>(
    () => ({ kind: "view_schema", id: initialViewSchemaID }),
  );
  const surfaceSelectionVersionRef = useRef(0);
  const [sheetReloadToken, setSheetReloadToken] = useState(0);
  const [pendingGridFocusSurface, setPendingGridFocusSurface] = useState<
    string | null
  >(null);
  const [hostRows, setHostRows] = useState<EntityRow[]>([]);
  const [identityRows, setIdentityRows] = useState<EntityRow[]>([]);
  const [entityLoadError, setEntityLoadError] = useState<string | null>(null);
  const [genericRows, setGenericRows] = useState<EntityApiRow[]>([]);
  const [genericLoadError, setGenericLoadError] = useState<string | null>(null);
  const [assessmentRows, setAssessmentRows] = useState<EntityApiRow[]>([]);
  const [assessmentLoadError, setAssessmentLoadError] = useState<string | null>(
    null,
  );
  const [savedViews, setSavedViews] = useState<SavedViewResource[]>([]);
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
  const upsertSavedView = useCallback((savedView: SavedViewResource) => {
    setSavedViews((current) => {
      const next = current.filter(
        (candidate) => candidate.saved_view_id !== savedView.saved_view_id,
      );
      return [...next, savedView].sort((left, right) =>
        left.display_name.localeCompare(right.display_name),
      );
    });
  }, []);
  const selectWorkbookSurface = useCallback(
    (
      viewSchemaID: string,
      options: { readonly focusFirstGridTarget?: boolean } = {},
    ) => {
      const nextSurface = knownWorkbookViewSchemaId(viewSchemaID);
      surfaceSelectionVersionRef.current += 1;
      setSurface(nextSurface);
      setStartupSheetRef({ kind: "view_schema", id: nextSurface });
      if (options.focusFirstGridTarget) {
        setPendingGridFocusSurface(nextSurface);
      }
    },
    [],
  );
  const selectSavedView = useCallback(
    (savedView: SavedViewResource) => {
      const nextSurface = knownWorkbookViewSchemaId(savedView.view_schema_id);
      const contract = workbookContractForViewSchemaId(nextSurface);
      applyQueryStateForSurface(
        nextSurface,
        workbookQueryStateFromSavedViewQueryJson(
          contract,
          savedView.query_json,
        ),
      );
      surfaceSelectionVersionRef.current += 1;
      setSurface(nextSurface);
      setStartupSheetRef({
        kind: "saved_view",
        id: savedView.saved_view_id,
      });
      setSheetReloadToken((current) => current + 1);
    },
    [applyQueryStateForSurface],
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
      setSavedViews((current) =>
        current.filter(
          (candidate) => candidate.saved_view_id !== savedView.saved_view_id,
        ),
      );
      if (
        startupSheetRef.kind === "saved_view" &&
        startupSheetRef.id === savedView.saved_view_id
      ) {
        const baseSurface = knownWorkbookViewSchemaId(savedView.view_schema_id);
        setSurface(baseSurface);
        setStartupSheetRef({ kind: "view_schema", id: baseSurface });
        setSheetReloadToken((current) => current + 1);
      }
    },
    [apiBase, incidentId, startupSheetRef],
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
          workbookQueryStateFromSavedViewQueryJson(
            contract,
            startupSavedView.query_json,
          ),
        );
      }
      setSurface(nextSurface);
      setStartupSheetRef({ ...startup.selectedSheetRef });
    };
    void loadStartup();
    return () => {
      cancelled = true;
    };
  }, [apiBase, applyQueryStateForSurface, incidentId, params, upsertSavedView]);

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
          setSavedViews([]);
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
        setSavedViews(nextSavedViews);
      }
    };

    void loadSavedViews();
    return () => {
      cancelled = true;
    };
  }, [apiBase, incidentId, onIncidentAccessLost]);

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
  }, [pendingGridFocusSurface, surface]);

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

const workbookStyle = {
  position: "relative" as const,
  display: "grid",
  gridTemplateRows: "minmax(0, 1fr) var(--ct-layout-statusStripHeight)",
  minHeight: "calc(100vh - var(--ct-layout-topBarHeight))",
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

const blurSurfaceButtonStyle = {
  position: "absolute" as const,
  insetBlockStart: 0,
  insetInlineStart: 0,
  zIndex: 2,
  border: 0,
  padding: 0,
  margin: 0,
  inlineSize: 1,
  blockSize: 1,
  opacity: 0,
  background: "transparent",
  color: "transparent",
  cursor: "default",
};

const visuallyHiddenStyle = {
  position: "absolute" as const,
  inlineSize: 1,
  blockSize: 1,
  overflow: "hidden",
  clipPath: "inset(50%)",
};

function statusIconStyle(saveState: SaveState): CSSProperties {
  return {
    display: "inline-block",
    inlineSize: "0.55rem",
    blockSize: "0.55rem",
    borderRadius: "var(--ct-rounded-pill)",
    background:
      saveState === "Conflict"
        ? "var(--ct-colors-semantic-conflict)"
        : saveState === "Syncing"
          ? "var(--ct-colors-semantic-caution)"
          : "var(--ct-colors-semantic-success)",
  };
}

const statusStripStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.8rem",
  minHeight: "var(--ct-layout-statusStripHeight)",
  padding: "0 1rem",
  borderTop: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.82rem",
};

const statusStripItemStyle = {
  display: "inline-flex",
  alignItems: "center",
  gap: "0.35rem",
  whiteSpace: "nowrap" as const,
};

const statusStripMutedItemStyle = {
  ...statusStripItemStyle,
  color: "var(--ct-colors-ink-subtle)",
};

const statusStripSecondaryItemStyle = {
  ...statusStripMutedItemStyle,
  minWidth: 0,
  maxWidth: "min(34rem, 42vw)",
  overflow: "hidden",
  textOverflow: "ellipsis",
};

const statusStripSpacerStyle = {
  flex: "1 1 auto",
};

const statusStripPresenceStyle = {
  ...statusStripItemStyle,
  gap: "0.25rem",
};

const presenceAvatarStyle = {
  display: "inline-grid",
  placeItems: "center",
  width: "1.5rem",
  height: "1.5rem",
  borderRadius: "var(--ct-rounded-pill)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-semantic-presence-self)",
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
  color: "var(--ct-colors-ink-muted)",
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

const timelineGridShellStyle = {
  ...gridShellStyle,
  blockSize:
    "calc(100vh - var(--ct-layout-topBarHeight) - var(--ct-layout-statusStripHeight) - var(--ct-layout-viewBarHeight) - 32px)",
  minBlockSize: "24rem",
};

const focusableCellStyle = {
  display: "block",
  lineHeight: "1.25rem",
  minHeight: "1.25rem",
  minWidth: "100%",
  maxWidth: "100%",
  outlineOffset: "2px",
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
};

const rowGutterContentStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  gap: "0.2rem",
  minWidth: 0,
};

const rowGutterPresenceStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  minWidth: "1rem",
  height: "1rem",
  borderRadius: "var(--ct-rounded-pill)",
  border: "var(--ct-border-hairline)",
  color: "var(--ct-colors-semantic-presence-other)",
  fontSize: "0.62rem",
  lineHeight: 1,
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

const gridCellInputStyle = {
  ...inputStyle,
  minHeight: "1.35rem",
  borderColor: "transparent",
  background: "transparent",
  padding: 0,
  color: "var(--ct-colors-ink)",
  lineHeight: 1.2,
};

const gridCellTextareaStyle = {
  ...gridCellInputStyle,
  resize: "none" as const,
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
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
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
  background: "var(--ct-colors-surface-3)",
};

const destructiveActionButtonStyle = {
  ...actionButtonStyle,
  borderColor: "var(--ct-colors-semantic-destructive)",
  background: "transparent",
  color: "var(--ct-colors-semantic-destructive)",
};

const conflictMarkerStyle = {
  ...secondaryActionButtonStyle,
  position: "absolute" as const,
  insetBlockStart: "4px",
  insetInlineEnd: "6px",
  boxSizing: "border-box" as const,
  minHeight: 0,
  height: "18px",
  margin: 0,
  borderColor: "var(--ct-colors-semantic-conflict)",
  color: "var(--ct-colors-semantic-conflict)",
  background: "var(--ct-colors-surface-2)",
  padding: "0 0.35rem",
  fontSize: "0.68rem",
  lineHeight: 1,
};

const cellPresenceStyle = {
  position: "absolute" as const,
  insetBlockStart: "4px",
  insetInlineEnd: "6px",
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  width: "fit-content",
  minHeight: 0,
  height: "18px",
  margin: 0,
  borderRadius: "var(--ct-rounded-pill)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-semantic-presence-other)",
  padding: "0 0.35rem",
  fontSize: "0.68rem",
  fontWeight: 700,
  lineHeight: 1,
};

const conflictResolverStyle = {
  display: "grid",
  gap: "0.75rem",
  padding: "1rem",
  borderTop: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
};

const conflictResolverGridStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(12rem, 1fr))",
  gap: "0.75rem",
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
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-semantic-success)",
};

const timelineEvidenceFlagOffStyle = {
  ...timelineEvidenceFlagBaseStyle,
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
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

const historySectionHeaderStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  gap: "0.75rem",
  flexWrap: "wrap" as const,
};

const historyMetaStyle = {
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
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
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
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
  borderRadius: "var(--ct-component-chip-rounded)",
  padding: "var(--ct-component-chip-padding)",
  font: "inherit",
  lineHeight: 1.2,
  maxWidth: "100%",
  minWidth: 0,
  overflowWrap: "anywhere" as const,
};

const unresolvedChipStyle = {
  border: "1px dashed var(--ct-colors-semantic-caution)",
  background: "var(--ct-component-chip-backgroundColor)",
  color: "var(--ct-colors-semantic-caution)",
};

const resolvedChipStyle = {
  border: "var(--ct-component-chip-border)",
  background: "var(--ct-component-chip-backgroundColor)",
  color: "var(--ct-colors-ink)",
};

const autoResolvedChipStyle = {
  border: "var(--ct-component-chip-border)",
  background: "var(--ct-component-chip-backgroundColor)",
  color: "var(--ct-colors-semantic-info)",
};

const dismissedChipStyle = {
  border: "var(--ct-border-hairline)",
  background: "transparent",
  color: "var(--ct-colors-ink-tertiary)",
};

const selectedChipStyle = {
  boxShadow: "0 0 0 2px var(--ct-colors-accent)",
};

const chipMetaStyle = {
  fontSize: "0.72rem",
  textTransform: "uppercase" as const,
  letterSpacing: "0.04em",
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

const emptyRelationshipStyle = {
  color: "var(--ct-colors-ink-tertiary)",
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
  color: "var(--ct-colors-ink-subtle)",
};

const mentionListButtonStyle = {
  border: "none",
  background: "transparent",
  color: "var(--ct-colors-ink)",
  padding: 0,
  textAlign: "left" as const,
  cursor: "pointer",
};

const mentionListButtonSelectedStyle = {
  boxShadow: "0 0 0 2px var(--ct-colors-accent)",
  outline: "2px solid transparent",
  outlineOffset: "2px",
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
  color: "var(--ct-colors-ink-muted)",
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
  borderRadius: "var(--ct-rounded-lg)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
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

const systemViewSwitcherStyle = {
  position: "relative" as const,
  minWidth: "10rem",
  border: 0,
  margin: 0,
  padding: 0,
};

const systemViewSwitcherTriggerStyle = {
  ...surfaceTabStyle,
  width: "100%",
  display: "flex",
  justifyContent: "space-between",
  alignItems: "center",
  gap: "0.75rem",
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  padding: "0.45rem 0.65rem",
};

const systemViewSwitcherValueStyle = {
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.85rem",
  fontWeight: 500,
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
};

const systemViewSwitcherMenuStyle = {
  position: "absolute" as const,
  zIndex: 10,
  insetBlockStart: "calc(100% + 0.35rem)",
  insetInlineStart: 0,
  width: "min(26rem, 80vw)",
  maxHeight: "28rem",
  overflowY: "auto" as const,
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-md)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-popover)",
  padding: "0.5rem",
};

const systemViewSwitcherGroupStyle = {
  display: "grid",
  gap: "0.25rem",
  padding: "0.35rem 0",
  border: 0,
  margin: 0,
  minInlineSize: 0,
};

const systemViewSwitcherGroupLabelStyle = {
  ...eyebrowStyle,
  margin: "0.2rem 0.45rem",
  padding: 0,
};

const systemViewSwitcherOptionStyle = {
  border: "0",
  borderRadius: "var(--ct-rounded-sm)",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  cursor: "pointer",
  font: "inherit",
  padding: "0.5rem 0.55rem",
  textAlign: "left" as const,
};

const systemViewSwitcherOptionSelectedStyle = {
  background: "var(--ct-colors-surface-3)",
  color: "var(--ct-colors-ink)",
  fontWeight: 700,
};

const savedViewSelectorFrameStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.4rem",
  flex: "0 0 auto",
  minWidth: "10rem",
};

const savedViewControlGroupStyle = {
  display: "flex",
  alignItems: "center",
  flexWrap: "wrap" as const,
  gap: "0.4rem",
  flex: "0 0 auto",
  minWidth: 0,
};

const savedViewSelectorLabelStyle = {
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.82rem",
  whiteSpace: "nowrap" as const,
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

const savedViewSelectStyle = {
  ...selectStyle,
  borderRadius: "var(--ct-rounded-sm)",
  padding: "0.42rem 0.55rem",
  maxWidth: "10rem",
};

const savedViewNameInputStyle = {
  ...inputStyle,
  width: "10rem",
  minWidth: "8rem",
  padding: "0.42rem 0.55rem",
};

const savedViewScopeSelectStyle = {
  ...savedViewSelectStyle,
  maxWidth: "7rem",
};

const savedViewStatusStyle = {
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.78rem",
  maxWidth: "12rem",
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
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
