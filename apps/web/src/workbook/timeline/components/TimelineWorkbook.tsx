import {
  buildGridPresentationRows,
  type GridCellAnchor,
  type GridColumn,
  type GridNavigationIntent,
  type GridPasteTargetResolution,
  type GridRow,
  navigateGridCellAnchor,
  reconcileRecordRows,
  resolveGridPasteTargets,
} from "@cartulary/grid-adapter";
import {
  conflictMarkerTestId,
  dataTestIdSelector,
  draftCellTestId,
  draftRelationshipItemsTestId,
  draftTimelineCollectionInputTestId,
  gridGroupRowTestId,
  gridRowGutterTestId,
  gridRowTestId,
  gridScrollportSelector,
  gridSortHeaderTestId,
  mentionItemTestId,
  relationshipItemsTestId,
  relationshipOverflowButtonTestId,
  rowCellTestId,
  timelineCollectionInputTestId,
  timelineInspectorSectionTestId,
  timelineMutationSubstrateReadyTestId,
  timelineRowVersionTestId,
  timelineScalarEditorTestId,
  type WorkbookSurface,
  workbookInlineDraftRowTestId,
} from "@cartulary/ui-contracts";
import {
  requireViewContract,
  resolveHeaderSortFieldKey,
} from "@cartulary/view-contracts";
import {
  type CSSProperties,
  type Dispatch,
  type ClipboardEvent as ReactClipboardEvent,
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
import { apiPath } from "../../../services/browserApi";
import {
  fetchJSON,
  parseErrorMessage,
  readEnvelope,
} from "../../../services/workbookApi";
import {
  createAndAttachEvidenceBlob,
  evidencePublicErrorMessage,
} from "../../../services/workbookEvidence";
import { WorkbookSheetToolbar } from "../../components/WorkbookSheetToolbar";
import { WorkbookStatusStrip } from "../../components/WorkbookStatusStrip";
import {
  WorkbookSurfaceFrame,
  workbookSurfaceGridShellStyle,
  workbookSurfaceOverlayPanelStyle,
} from "../../components/WorkbookSurfaceFrame";
import { buildEvidenceCountDisplayViewModel } from "../../models/evidenceLifecycleViewModel";
import {
  buildQueryRequest,
  defaultFilterDraft,
  emptyWorkbookQueryState,
  type FilterDraft,
  removeFilterField,
  type WorkbookQueryState,
} from "../../models/workbookQuery";
import type { WorkbookSheetRef } from "../../models/workbookStartup";
import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "../../models/workbookSurfaceRegistry";
import { clipboardGridDimensions } from "../../utils/workbookClipboard";
import {
  captureViewportAnchor,
  computeRestoredViewportScroll,
  isRectFullyVisibleWithinContainer,
  type ScrollPosition,
  type ViewportSnapshot,
} from "../../utils/workbookContinuity";
import type { WorkbookFocusAnchor } from "../../utils/workbookGridFocus";
import {
  mapWorkbookKeyboardCommand,
  type WorkbookKeyboardCommand,
} from "../../utils/workbookKeyboard";
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
} from "../../utils/workbookPendingQueue";
import {
  isPresenceRecord,
  presenceMatchesSheet,
  type WorkbookPresenceInput,
  type WorkbookPresenceMode,
} from "../../utils/workbookPresence";
import { visuallyHiddenStyle } from "../../utils/workbookStyles";
import { stringifyGridValue } from "../../utils/workbookValueFormat";
import { useTimelineCommittedRows } from "../hooks/useTimelineCommittedRows";
import { useTimelineConflicts } from "../hooks/useTimelineConflicts";
import { useTimelineEvidenceActions } from "../hooks/useTimelineEvidenceActions";
import {
  type TimelineGridInteractionRefs,
  useTimelineGridInteractions,
} from "../hooks/useTimelineGridInteractions";
import { useTimelineHistoryState } from "../hooks/useTimelineHistoryState";
import { useTimelineInspectorSelection } from "../hooks/useTimelineInspectorSelection";
import {
  type TimelineLiveUpdateRefs,
  type TimelinePresenceDraft,
  useTimelineLiveUpdates,
} from "../hooks/useTimelineLiveUpdates";
import { useTimelineMentions } from "../hooks/useTimelineMentions";
import {
  type TimelinePendingQueueRuntime,
  type TimelinePendingReplayAdmissionRequest,
  type TimelinePendingSavesRefs,
  timelineTabClientInstanceId,
  useTimelinePendingSaves,
} from "../hooks/useTimelinePendingSaves";
import { useTimelineRows } from "../hooks/useTimelineRows";
import { useTimelineWorkbookRuntime } from "../hooks/useTimelineWorkbookRuntime";
import { parseSameFieldConflict } from "../models/timelineConflictModel";
import {
  beginTimelineEntityRefreshBarrier,
  settleTimelineViewportContinuityBarrier,
  type TimelineEntityCatalogInput,
  type TimelineEntityRefreshSettleState,
  type TimelineViewportContinuityBarrier,
  timelineEntityRefreshExpectationForMention,
  timelineViewportContinuityBarrierSatisfied,
  withTimelineEntityRefreshExpectation,
} from "../models/timelineViewportContinuityModel";
import {
  type AutoResolutionNotice,
  buildAutoResolutionNotices,
  type CollectionItem,
  type DismissedMention,
  type InspectorMention,
} from "../models/workbookMentionChips";
import {
  applyViewRowPatch,
  buildAssessmentCreatePayload,
  buildAttachedEvidenceCreatePayload,
  buildAttachedEvidencePatchPayload,
  buildCollectionPatchIntent,
  buildCreatePayload,
  buildExpandedTimelineColumnWidths,
  buildScalarPatchIntent,
  type CollectionDraftKey,
  type CollectionFieldKey,
  clipboardTextLooksTabular,
  confidenceScoreFromBand,
  createDraftRow,
  createDraftRowForKey,
  decideWorkbookRecordFreshness,
  type EntityApiRow,
  ensureDraftRow,
  type FocusFieldKey,
  inputFocusKey,
  type LocalConflictState,
  materializePendingReplayPayload,
  normalizeTimelineFullRow,
  normalizeTimelinePatchCells,
  type RowValues,
  readTimelineCellValue,
  rowFromApi,
  type SameFieldConflictPayload,
  type TagCollectionItem,
  type TimelineCollectionBinding,
  type TimelineConflictResolution,
  type TimelinePatchCells,
  type TimelineScalarBinding,
  type TimelineScalarEditorSurface,
  timelineColumnWidth,
  timelineFieldBinding,
  timelineFocusFieldForFieldKey,
  timelineGroupLabel,
  timelineInspectorBindings,
  timelineRelationshipLabel,
  timelineScalarBindingForField,
  timelineScalarBindings,
  timelineScalarEditorSurfaces,
  timelineVisibleBindings,
  validateTimelineViewSchemaId,
  type WorkbookRecordFreshnessDecision,
  type WorkbookRow,
  type WorkbookVersionedRecord,
} from "../models/workbookTimelineModel";
import {
  buildMentionActionPayload,
  buildMentionPatchPayload,
  isRecordChangedMessage,
  type MentionResolutionAction,
  type RecordChangedPayload,
  shouldIgnoreSelfOriginatedRecordChange,
} from "../services/workbookShellPhase4";
import {
  createWorkbookSocketLifecycleState,
  type WorkbookSocketLifecycleAction,
  type WorkbookSocketLifecycleEffect,
} from "../services/workbookSocketLifecycle";
import {
  DraftRowCreateButton,
  mentionChipStateForItem,
  RelationshipChip,
  relationshipChipBaseStyle,
  relationshipItemLabel,
  TimelineScalarEditor,
} from "./TimelineCellEditors";
import { WorkbookGridControls } from "../../components/WorkbookGridControls";
import { TimelineConflictResolver } from "./TimelineConflictResolver";
import { TimelineEvidencePanel } from "./TimelineEvidencePanel";
import { TimelineGridSurface } from "./TimelineGridSurface";
import {
  type RecordHistoryData,
  type RecordHistoryItem,
  type RecordHistoryRollbackAction,
  type RowHistoryPendingAction,
  TimelineHistoryPanel,
} from "./TimelineHistoryPanel";
import {
  TimelineCellPresenceMarker,
  TimelineRowGutterContent,
} from "./TimelinePresenceMarkers";
import {
  TimelineRowContextMenu,
  type TimelineRowContextMenuPosition,
} from "./TimelineRowActions";
import { TimelineWorkbookInspector } from "./TimelineWorkbookInspector";
import {
  TimelineWorkbookNotices,
  timelinePendingQueueMessage,
} from "./TimelineWorkbookNotices";

export type { WorkbookRecordFreshnessDecision, WorkbookVersionedRecord };
export {
  buildAssessmentCreatePayload,
  buildCreatePayload,
  clipboardTextLooksTabular,
  confidenceScoreFromBand,
  createDraftRow,
  decideWorkbookRecordFreshness,
  ensureDraftRow,
  pendingReplayCapacity,
};

const timelineContract = requireViewContract(timelineViewSchemaId);
const timelineRowGutterWidth = 58;
const timelineVisibleFieldKeys = timelineVisibleBindings.map(
  (binding) => binding.fieldKey,
);

export type SaveState = "Syncing" | "Saved" | "Conflict";
type FilterDraftSetter = Dispatch<SetStateAction<FilterDraft>>;
type WorkbookQueryStateSetter = Dispatch<SetStateAction<WorkbookQueryState>>;
export type IncidentRole = "viewer" | "editor" | "reviewer" | "admin" | "";
type TimelineInspectorHistorySubject =
  | {
      readonly kind: "live";
      readonly recordId: string;
      readonly rowVersion: number | null;
    }
  | {
      readonly kind: "deleted";
      readonly recordId: string;
      readonly rowVersion: number;
    }
  | { readonly kind: "draft" }
  | { readonly kind: "none" };

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

export type TimelineWorkbookProps = {
  incidentId: string;
  apiBase?: string | undefined;
  sheetRef?: WorkbookSheetRef | undefined;
  reloadToken?: number | undefined;
  renderInlineQueryControls?: boolean | undefined;
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

type WorkbookQueryEnvelope = {
  data: {
    incident_id: string;
    view_schema_id: string;
    rows: unknown[];
  };
};

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

type TimelinePasteTargetResolution = {
  anchor: GridCellAnchor | null;
  targetResolution: GridPasteTargetResolution;
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

type PendingReplayAdmissionRequest =
  TimelinePendingReplayAdmissionRequest<PendingReplayRuntimeMeta>;

type PendingQueueRuntime =
  TimelinePendingQueueRuntime<PendingReplayRuntimeMeta>;

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

type ViewMutationEnvelope = {
  data: {
    view_schema_id: string;
    change_set_id: string;
    row: EntityApiRow;
  };
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

type ViewportContinuityTarget =
  | { kind: "row-inspect"; recordId: string }
  | { kind: "input"; focusKey: string }
  | { kind: "scroll-only" };

type TimelineRowContextMenuState = {
  readonly position: TimelineRowContextMenuPosition;
  readonly recordId: string;
};

type ViewportContinuityRequest = {
  token: number;
  attemptVersion: number;
  target: ViewportContinuityTarget;
  preservedViewport: ViewportSnapshot | null;
  barrier: TimelineViewportContinuityBarrier;
};

function isCollectionDraftKey(
  focusField: FocusFieldKey,
): focusField is CollectionDraftKey {
  return (
    focusField === "hostRefs" ||
    focusField === "identityRefs" ||
    focusField === "tags"
  );
}

type LoadRowsOptions = {
  showLoading: boolean;
  freshnessRetryDepth?: number;
  viewportContinuityToken?: number;
};

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

function socketIsOpen(socket: WebSocket) {
  return socket.readyState === WebSocket.OPEN;
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
  for (const binding of timelineScalarBindings) {
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

function timelineClipboardShouldDispatchTabular(
  fieldKey: string,
  clipboardText: string,
) {
  if (clipboardTextLooksTabular(clipboardText)) {
    return true;
  }
  return (
    fieldKey === "timeline.occurred_at" &&
    clipboardGridDimensions(clipboardText).columnCount > 1
  );
}

function timelinePasteColumnsFromStart(
  columns: readonly GridColumn<WorkbookRow>[],
  startFieldKey: string,
  pastedColumnCount: number,
) {
  const startColumnIndex = columns.findIndex(
    (column) => column.fieldKey === startFieldKey,
  );
  if (startColumnIndex < 0) {
    return null;
  }
  const targetColumns = columns
    .slice(startColumnIndex, startColumnIndex + pastedColumnCount)
    .map((column) => column.fieldKey);
  return targetColumns.length < 1 ? null : targetColumns;
}

function resolveDraftTimelinePasteTargets({
  columns,
  pastedColumnCount,
  pastedRowCount,
  startFieldKey,
}: {
  readonly columns: readonly GridColumn<WorkbookRow>[];
  readonly pastedColumnCount: number;
  readonly pastedRowCount: number;
  readonly startFieldKey: string;
}): GridPasteTargetResolution | null {
  const targetColumns = timelinePasteColumnsFromStart(
    columns,
    startFieldKey,
    pastedColumnCount,
  );
  if (targetColumns === null || pastedRowCount < 1) {
    return null;
  }
  return {
    columns: targetColumns,
    rowTargets: Array.from({ length: pastedRowCount }, (_, createIndex) => ({
      createIndex,
      kind: "create" as const,
    })),
  };
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

export function TimelineWorkbook({
  incidentId,
  apiBase,
  sheetRef,
  reloadToken = 0,
  renderInlineQueryControls = true,
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
  const entityCatalogInput = useMemo(
    () =>
      ({
        hostEntities,
        identityEntities,
      }) satisfies TimelineEntityCatalogInput,
    [hostEntities, identityEntities],
  );
  const timelineRuntime = useTimelineWorkbookRuntime({
    controlledFilterDraft,
    controlledQueryState,
    onFilterDraftChange,
    onQueryStateChange,
  });
  const {
    isInitialLoading,
    loadError,
    refreshError,
    saveState,
    saveStateSecondaryMessage,
    setIsInitialLoading,
    setLoadError,
    setRefreshError,
    setSaveState,
    setSaveStateSecondaryMessage,
  } = timelineRuntime.lifecycle;
  const {
    applyQueryFilter,
    filterDraft,
    handleQueryGroupByChange,
    handleQuerySortToggle,
    queryState,
    setFilterDraft,
    setQueryState,
  } = timelineRuntime.query;
  const initialTimelineRows = useMemo(() => [createDraftRow(1)], []);
  const rowsRef = useRef<WorkbookRow[]>(initialTimelineRows);
  const draftCounterRef = useRef(2);
  const timelineRows = useTimelineRows({ draftCounterRef, rowsRef });
  const { rows } = timelineRows.snapshot;
  const { setRows } = timelineRows.commands;
  const conflictQueueRef = useRef<Record<string, LocalConflictState>>({});
  const timelineConflicts = useTimelineConflicts({ conflictQueueRef });
  const { activeConflictKey, conflictQueue, pasteConflictGroup } =
    timelineConflicts.snapshot;
  const { setActiveConflictKey, setConflictQueueState, setPasteConflictGroup } =
    timelineConflicts.commands;
  const timelineMentions = useTimelineMentions();
  const {
    autoResolutionNotices,
    dismissedMentionsByRow,
    selectedMentionRef,
    selectedResolveTargetId,
  } = timelineMentions.snapshot;
  const {
    setAutoResolutionNotices,
    setDismissedMentionsByRow,
    setSelectedMentionRef,
    setSelectedResolveTargetId,
  } = timelineMentions.commands;
  const timelineEvidenceActions = useTimelineEvidenceActions();
  const { inspectorMessage } = timelineEvidenceActions.snapshot;
  const { setInspectorMessage } = timelineEvidenceActions.commands;
  const activeSocketRef = useRef<WebSocket | null>(null);
  const socketLifecycleRef = useRef(createWorkbookSocketLifecycleState());
  const socketEstablishedRef = useRef(false);
  const socketConnectionIDRef = useRef<string | null>(null);
  const presenceUpdateTimerRef = useRef<number | null>(null);
  const currentPresenceRef = useRef<TimelinePresenceDraft>({
    fieldKey: null,
    mode: "viewing",
    recordId: null,
  });
  const socketReconnectAfterAuthRef = useRef<(() => void) | null>(null);
  const socketResumeTokenRef = useRef<string | null>(null);
  const socketLastSeenStreamSeqRef = useRef(0);
  const dispatchSocketLifecycleRef = useRef<
    (action: WorkbookSocketLifecycleAction) => WorkbookSocketLifecycleEffect[]
  >(() => []);
  const timelineLiveUpdateRefs: TimelineLiveUpdateRefs = {
    activeSocketRef,
    currentPresenceRef,
    dispatchSocketLifecycleRef,
    presenceUpdateTimerRef,
    socketConnectionIDRef,
    socketEstablishedRef,
    socketLastSeenStreamSeqRef,
    socketLifecycleRef,
    socketReconnectAfterAuthRef,
    socketResumeTokenRef,
  };
  const timelineLiveUpdates = useTimelineLiveUpdates({
    refs: timelineLiveUpdateRefs,
  });
  const { currentPresence, presenceRecords } = timelineLiveUpdates.snapshot;
  const {
    dispatchSocketLifecycle,
    setCurrentPresence,
    setPresenceRecords,
    syncSocketLifecycleRefs,
  } = timelineLiveUpdates.commands;
  const pendingOpsRef = useRef(0);
  const pendingSignaturesRef = useRef(new Map<string, string>());
  const collectionKeyboardCommitRef = useRef(new Map<string, string>());
  const pendingSocketTxnTimeoutsRef = useRef(new Map<string, number>());
  const saveQueueRef = useRef(Promise.resolve());
  const socketClientInstanceIdRef = useRef<string | null>(null);
  if (socketClientInstanceIdRef.current === null) {
    socketClientInstanceIdRef.current = timelineTabClientInstanceId();
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
  const pendingReplayOrderRef = useRef(1);
  const pendingReplayTimerRef = useRef<number | null>(null);
  const pendingReplayAuthRetryRef = useRef<number | null>(null);
  const schedulePendingReplayRef = useRef<() => void>(() => undefined);
  const timelinePendingSaveRefs: TimelinePendingSavesRefs<PendingReplayRuntimeMeta> =
    {
      collectionKeyboardCommitRef,
      pendingOpsRef,
      pendingQueueRef,
      pendingReplayAuthRetryRef,
      pendingReplayOrderRef,
      pendingReplayTimerRef,
      pendingSignaturesRef,
      pendingSocketTxnTimeoutsRef,
      saveQueueRef,
      schedulePendingReplayRef,
      socketClientInstanceIdRef,
    };
  const timelinePendingSaves =
    useTimelinePendingSaves<PendingReplayRuntimeMeta>({
      refs: timelinePendingSaveRefs,
    });
  const { pendingQueueSnapshot } = timelinePendingSaves.snapshot;
  const { setPendingQueueSnapshot } = timelinePendingSaves.commands;
  const workbookFocusAnchorRef = useRef<WorkbookFocusAnchor | null>(null);
  const rowInputRefs = useRef(
    new Map<string, HTMLInputElement | HTMLTextAreaElement>(),
  );
  const rowInputTestIdsRef = useRef(new Map<string, string>());
  const timelineAnchorColumnsRef = useRef<readonly GridColumn<WorkbookRow>[]>(
    [],
  );
  const timelineAnchorRowsRef = useRef<readonly GridRow<WorkbookRow>[]>([]);
  const gridShellRef = useRef<HTMLDivElement | null>(null);
  const [timelineGridShellWidth, setTimelineGridShellWidth] = useState(0);
  const [isInspectorOpen, setIsInspectorOpen] = useState(false);
  const viewportContinuityTokenRef = useRef(1);
  const timelineGridInteractionRefs: TimelineGridInteractionRefs = {
    gridShellRef,
    rowInputRefs,
    rowInputTestIdsRef,
    timelineAnchorColumnsRef,
    timelineAnchorRowsRef,
    viewportContinuityTokenRef,
    workbookFocusAnchorRef,
  };
  const timelineGridInteractions =
    useTimelineGridInteractions<ViewportContinuityRequest>({
      refs: timelineGridInteractionRefs,
    });
  const { viewportContinuityRequest, workbookFocusAnchor } =
    timelineGridInteractions.snapshot;
  const { setViewportContinuityRequest, setWorkbookFocusAnchor } =
    timelineGridInteractions.commands;

  useLayoutEffect(() => {
    const gridShell = gridShellRef.current;
    if (gridShell === null) {
      return;
    }
    const updateGridShellWidth = (width: number) => {
      const measuredWidth = Math.max(0, Math.floor(width));
      setTimelineGridShellWidth((current) =>
        current === measuredWidth ? current : measuredWidth,
      );
    };
    const updateFromElement = () => {
      updateGridShellWidth(gridShell.clientWidth);
    };
    const scheduleUpdateFromElement = () => {
      updateFromElement();
      window.requestAnimationFrame(updateFromElement);
    };

    updateFromElement();
    window.addEventListener("resize", scheduleUpdateFromElement);
    window.visualViewport?.addEventListener(
      "resize",
      scheduleUpdateFromElement,
    );

    if (typeof ResizeObserver === "undefined") {
      return () => {
        window.removeEventListener("resize", scheduleUpdateFromElement);
        window.visualViewport?.removeEventListener(
          "resize",
          scheduleUpdateFromElement,
        );
      };
    }

    const observer = new ResizeObserver(() => {
      scheduleUpdateFromElement();
    });
    observer.observe(gridShell);
    observer.observe(document.documentElement);
    return () => {
      window.removeEventListener("resize", scheduleUpdateFromElement);
      window.visualViewport?.removeEventListener(
        "resize",
        scheduleUpdateFromElement,
      );
      observer.disconnect();
    };
  }, []);

  const [replacementDrafts, setReplacementDrafts] = useState<
    Record<string, string>
  >({});
  const [rowContextMenu, setRowContextMenu] =
    useState<TimelineRowContextMenuState | null>(null);
  useLayoutEffect(() => {
    const gridShell = gridShellRef.current;
    if (gridShell === null) {
      return;
    }
    const measuredWidth = Math.max(0, Math.floor(gridShell.clientWidth));
    setTimelineGridShellWidth((current) =>
      current === measuredWidth ? current : measuredWidth,
    );
  }, []);

  const [activeCollectionInputKey, setActiveCollectionInputKey] = useState<
    string | null
  >(null);
  const timelineInspectorSelection = useTimelineInspectorSelection({
    currentIncidentRole,
    dismissedMentionsByRow,
    rows,
    selectedMentionRef,
  });
  const {
    canManageMentions,
    draftRow,
    inspectorMentions,
    selectedMention,
    selectedRow,
    selectedRowId,
  } = timelineInspectorSelection.snapshot;
  const { setSelectedRowId } = timelineInspectorSelection.commands;
  const timelineHistoryState = useTimelineHistoryState();
  const { rowHistory, rowHistoryPendingAction } = timelineHistoryState.snapshot;
  const { setRowHistory, setRowHistoryPendingAction } =
    timelineHistoryState.commands;
  const currentHistoryRecordIdRef = useRef<string | null>(null);
  const rowHistoryRequestSeqRef = useRef(0);
  const clientTxnRef = useRef(1);
  const timelineCommittedRows = useTimelineCommittedRows({ rowsRef });
  const {
    acceptCommittedTimelineRow,
    acceptCommittedTimelineRows,
    acceptTimelineActionResult,
    acceptTimelineRecordVersion,
    beginLoad: beginTimelineRowsLoad,
    committedRowsChangedSince,
    currentCommittedTimelineRow,
    hasLoadedRows,
    isCurrentLoadSequence,
    isStaleTimelineRowVersion,
    knownTimelineRowVersion,
    latestCommittedRowVersion,
    latestCommittedTimelineRow,
    markRowsLoaded,
  } = timelineCommittedRows.commands;
  const scalarDraftValuesRef = useRef(new Map<string, string>());
  const loadRowsRef = useRef<(options: LoadRowsOptions) => Promise<void>>(
    async () => undefined,
  );

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
    [setAutoResolutionNotices],
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
    [computeSaveStatePresentation, setSaveState, setSaveStateSecondaryMessage],
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
    [setWorkbookFocusAnchor],
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
  }, [publishSaveStatePresentation, setPendingQueueSnapshot]);
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
      socketClientInstanceIdRef.current ?? timelineTabClientInstanceId();
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

  const handleResolveTargetChange = useCallback(
    (value: string) => {
      setSelectedResolveTargetId(value);
      if (value !== "") {
        setInspectorMessage(`Selected ${value}`);
      }
    },
    [setSelectedResolveTargetId, setInspectorMessage],
  );
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
      resolveElement: () => HTMLElement | null,
      preservedViewport: ViewportSnapshot | null,
    ) => {
      // Continuity restores the previous scroll position first, then applies
      // only the extra delta needed to keep the target fully visible.
      const currentViewport =
        preservedViewport ??
        ({
          scroll: currentGridScrollSnapshot(),
          anchor: null,
        } satisfies ViewportSnapshot);
      const preservedScroll = currentViewport.scroll;
      const focusResolvedElement = () => {
        const element = resolveElement();
        if (element === null || !element.isConnected) {
          return false;
        }
        if (!element.hasAttribute("tabindex")) {
          element.tabIndex = -1;
        }
        element.focus({ preventScroll: true });
        return document.activeElement === element;
      };
      window.focus();
      const focusedNow = focusResolvedElement();
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
        const fullyVisible = isRectFullyVisibleWithinContainer(
          resolveGridScrollElement(
            updatedGridShell,
            "timeline",
          ).getBoundingClientRect(),
          updatedElement.getBoundingClientRect(),
        );
        return focusResolvedElement() && fullyVisible;
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
      return focusedNow && restoredNow;
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
    const scalarBinding = timelineScalarBindings.find(
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
      const binding = timelineFieldBinding(anchor.fieldKey);
      if (binding.kind === "collection") {
        const collectionInput = document.querySelector<HTMLInputElement>(
          dataTestIdSelector(
            timelineCollectionInputTestId(anchor.recordId, anchor.fieldKey),
          ),
        );
        if (collectionInput !== null) {
          return collectionInput;
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
          return (
            document.querySelector<HTMLElement>(
              dataTestIdSelector(
                rowCellTestId(target.recordId, "timeline.summary"),
              ),
            ) ??
            document.querySelector<HTMLElement>(
              dataTestIdSelector(
                gridRowGutterTestId(timelineViewSchemaId, target.recordId),
              ),
            )
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
      options: { barrier?: TimelineViewportContinuityBarrier } = {},
    ) => {
      const token = viewportContinuityTokenRef.current;
      viewportContinuityTokenRef.current += 1;
      setViewportContinuityRequest({
        token,
        attemptVersion: 0,
        target,
        preservedViewport: currentGridViewportSnapshot(
          resolveViewportContinuityElement(target),
        ),
        barrier: options.barrier ?? null,
      });
      return token;
    },
    [
      currentGridViewportSnapshot,
      resolveViewportContinuityElement,
      setViewportContinuityRequest,
    ],
  );
  const beginViewportContinuityRef = useRef(beginViewportContinuity);
  beginViewportContinuityRef.current = beginViewportContinuity;

  const settleViewportContinuityBarrier = useCallback(
    (token: number, refreshState: TimelineEntityRefreshSettleState) => {
      setViewportContinuityRequest((current) => {
        if (!current || current.token !== token) {
          return current;
        }
        return {
          ...current,
          barrier: settleTimelineViewportContinuityBarrier(
            current.barrier,
            refreshState,
          ),
          attemptVersion: current.attemptVersion + 1,
        };
      });
    },
    [setViewportContinuityRequest],
  );

  const clearViewportContinuity = useCallback(
    (token: number) => {
      setViewportContinuityRequest((current) =>
        current?.token === token ? null : current,
      );
    },
    [setViewportContinuityRequest],
  );

  const advanceViewportContinuity = useCallback(
    (
      token: number | undefined,
      options: {
        barrier?: TimelineViewportContinuityBarrier;
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
          barrier:
            options.barrier === undefined ? current.barrier : options.barrier,
          target: options.target ?? current.target,
        };
      });
    },
    [setViewportContinuityRequest],
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
      return !timelineViewportContinuityBarrierSatisfied(
        continuity.barrier,
        entityCatalogInput,
      );
    },
    [entityCatalogInput],
  );

  const applyRowMutation = useCallback(
    (
      rowKey: string,
      envelope: TimelineMutationEnvelope,
      options: {
        continueOnFreshDraft?: boolean;
        clearActiveCollectionFocusKey?: string | undefined;
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
        if (options.clearActiveCollectionFocusKey !== undefined) {
          setActiveCollectionInputKey((current) =>
            current === options.clearActiveCollectionFocusKey ? null : current,
          );
        }
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
      return committed;
    },
    [
      acceptCommittedTimelineRow,
      advanceViewportContinuity,
      nextDraftIndex,
      pruneAutoResolutionNoticesForRows,
      resolveInputElement,
      selectedRowId,
      setDismissedMentionsByRow,
      setAutoResolutionNotices,
      setRows,
      setSelectedRowId,
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
      const { queryStartEpoch, requestSequence } = beginTimelineRowsLoad();

      if (options.showLoading && !hasLoadedRows()) {
        setIsInitialLoading(true);
      }
      if (hasLoadedRows()) {
        setRefreshError(null);
      } else {
        setLoadError(null);
      }

      const result = await fetchJSON<WorkbookQueryEnvelope>(queryPath, {
        method: "POST",
        body: queryBody,
      });

      if (!isCurrentLoadSequence(requestSequence)) {
        return;
      }

      if (committedRowsChangedSince(queryStartEpoch)) {
        const refreshed = await refreshTimelineRowsAfterStaleResult(options);
        if (!refreshed && !hasLoadedRows()) {
          setIsInitialLoading(false);
        }
        return;
      }

      if (!result.ok) {
        if (options.viewportContinuityToken !== undefined) {
          clearViewportContinuity(options.viewportContinuityToken);
        }
        const message = "Timeline projection load failed.";
        if (hasLoadedRows()) {
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
        if (hasLoadedRows()) {
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
      markRowsLoaded();
      setLoadError(null);
      setRefreshError(null);
      setIsInitialLoading(false);
    },
    [
      advanceViewportContinuity,
      acceptCommittedTimelineRows,
      beginTimelineRowsLoad,
      clearViewportContinuity,
      committedRowsChangedSince,
      pruneAutoResolutionNoticesForRows,
      publishSaveStatePresentation,
      freshTimelineRowsForQueryResult,
      hasLoadedRows,
      isCurrentLoadSequence,
      markRowsLoaded,
      nextDraftIndex,
      queryBody,
      queryPath,
      refreshTimelineRowsAfterStaleResult,
      setIsInitialLoading,
      setLoadError,
      setRefreshError,
      setDismissedMentionsByRow,
      setRows,
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
      setRows,
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
      window.requestAnimationFrame(() => {
        if (cancelled) {
          return;
        }
        if (shouldHoldViewportContinuity(viewportContinuityRequest)) {
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
        clearViewportContinuity(viewportContinuityRequest.token);
      });
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
      const deletedHistoryMatchesSelectedRow =
        rowHistory.data?.deleted === true &&
        rowHistory.data.record_id === selectedRowId;
      const previousAnchor = workbookFocusAnchorRef.current;
      setSelectedRowId(null);
      setSelectedMentionRef(null);
      setSelectedResolveTargetId("");
      setRowHistory((current) => {
        if (
          current.recordId !== selectedRowId ||
          current.data?.deleted === true
        ) {
          return current;
        }
        rowHistoryRequestSeqRef.current += 1;
        return {
          recordId: null,
          status: "idle",
          data: null,
          message: null,
        };
      });
      setRowHistoryPendingAction((current) =>
        current?.recordId === selectedRowId ? null : current,
      );
      setInspectorMessage(
        deletedHistoryMatchesSelectedRow
          ? "Selected row was deleted."
          : "Selected row is no longer available.",
      );
      if (deletedHistoryMatchesSelectedRow) {
        return;
      }
      window.setTimeout(() => {
        const fallbackFieldKey =
          previousAnchor?.surface === timelineViewSchemaId
            ? previousAnchor.fieldKey
            : "timeline.summary";
        const fallbackRow = rows.find((row) => row.recordId !== null);
        if (fallbackRow?.recordId) {
          if (
            restoreTimelineFocusAnchor({
              fieldKey: fallbackFieldKey,
              recordId: fallbackRow.recordId,
            })
          ) {
            return;
          }
        }
        const gridShell = gridShellRef.current;
        if (gridShell !== null) {
          if (!gridShell.hasAttribute("tabindex")) {
            gridShell.tabIndex = -1;
          }
          gridShell.focus({ preventScroll: true });
        }
      }, 0);
    }
  }, [
    restoreTimelineFocusAnchor,
    rowHistory.data,
    rows,
    selectedRowId,
    setSelectedMentionRef,
    setInspectorMessage,
    setSelectedResolveTargetId,
    setRowHistory,
    setRowHistoryPendingAction,
    setSelectedRowId,
  ]);

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
  }, [
    inspectorMentions,
    selectedMentionRef,
    setSelectedMentionRef,
    setSelectedResolveTargetId,
  ]);

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
    [publishSaveStatePresentation, setSaveState, setSaveStateSecondaryMessage],
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
    [
      setConflictQueueState,
      updateSaveStateForConflicts,
      setActiveConflictKey,
      setRows,
    ],
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
      for (const binding of timelineScalarBindings) {
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
    [setRows],
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
        clearActiveCollectionFocusKey:
          meta.surface === "grid" && isCollectionDraftKey(meta.focusField)
            ? meta.focusKey
            : undefined,
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
      socketClientInstanceIdRef.current ?? timelineTabClientInstanceId();
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
  }, [changeSocketURL, incidentId, setRefreshError, setPresenceRecords]);

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
        socketClientInstanceIdRef.current ?? timelineTabClientInstanceId();
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
    [apiBase, enqueuePendingReplayUnit, incidentId, setRows],
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
      const binding = timelineScalarBindings.find(
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
    async (
      recordId: string,
      options: {
        readonly retainedData?: RecordHistoryData | null;
        readonly setLoading?: boolean;
      } = {},
    ): Promise<RecordHistoryData | null> => {
      const requestSeq = rowHistoryRequestSeqRef.current + 1;
      rowHistoryRequestSeqRef.current = requestSeq;
      if (options.setLoading === true) {
        setRowHistoryPendingAction(null);
        setRowHistory({
          recordId,
          status: "loading",
          data:
            options.retainedData?.record_id === recordId
              ? options.retainedData
              : null,
          message: null,
        });
      }
      const result = await fetchJSON<RecordHistoryEnvelope>(
        apiPath(apiBase, `/api/v1/records/${recordId}/history`),
      );
      if (rowHistoryRequestSeqRef.current !== requestSeq) {
        return null;
      }
      if (!result.ok) {
        setRowHistoryPendingAction(null);
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
        if (historyData.record_id !== recordId) {
          throw new Error("row history response record mismatch");
        }
      } catch {
        if (rowHistoryRequestSeqRef.current !== requestSeq) {
          return null;
        }
        setRowHistoryPendingAction(null);
        setRowHistory({
          recordId,
          status: "error",
          data: null,
          message: "Invalid row history response.",
        });
        return null;
      }
      if (rowHistoryRequestSeqRef.current !== requestSeq) {
        return null;
      }
      acceptTimelineRecordVersion(recordId, historyData.row_version);
      setRowHistoryPendingAction(null);
      setRowHistory({
        recordId,
        status: "ready",
        data: historyData,
        message: null,
      });
      return historyData;
    },
    [
      acceptTimelineRecordVersion,
      apiBase,
      setRowHistory,
      setRowHistoryPendingAction,
    ],
  );

  const openRowHistory = useCallback(
    (recordId: string) => {
      setSelectedRowId(recordId);
      setIsInspectorOpen(true);
      void fetchRecordHistory(recordId, {
        retainedData: rowHistory.recordId === recordId ? rowHistory.data : null,
        setLoading: true,
      });
    },
    [
      fetchRecordHistory,
      rowHistory.data,
      rowHistory.recordId,
      setSelectedRowId,
    ],
  );

  const matchedRowHistoryData =
    rowHistory.data !== null &&
    rowHistory.data.record_id === rowHistory.recordId
      ? rowHistory.data
      : null;
  const deletedRowHistoryData =
    matchedRowHistoryData?.deleted === true ? matchedRowHistoryData : null;
  const selectedLiveRecordId = selectedRow?.recordId ?? null;
  const deletedRowIsActiveSubject =
    deletedRowHistoryData !== null &&
    (selectedLiveRecordId === null ||
      selectedLiveRecordId === deletedRowHistoryData.record_id);
  const inspectorHistorySubject: TimelineInspectorHistorySubject =
    deletedRowIsActiveSubject && deletedRowHistoryData !== null
      ? {
          kind: "deleted",
          recordId: deletedRowHistoryData.record_id,
          rowVersion: deletedRowHistoryData.row_version,
        }
      : selectedLiveRecordId !== null
        ? {
            kind: "live",
            recordId: selectedLiveRecordId,
            rowVersion: selectedRow?.rowVersion ?? null,
          }
        : draftRow !== null
          ? { kind: "draft" }
          : { kind: "none" };
  const currentHistoryRecordId =
    inspectorHistorySubject.kind === "live" ||
    inspectorHistorySubject.kind === "deleted"
      ? inspectorHistorySubject.recordId
      : null;
  currentHistoryRecordIdRef.current = currentHistoryRecordId;
  const currentHistoryRowVersion =
    inspectorHistorySubject.kind === "live" ||
    inspectorHistorySubject.kind === "deleted"
      ? inspectorHistorySubject.rowVersion
      : null;
  const currentHistoryDeleted = inspectorHistorySubject.kind === "deleted";
  const activeHistoryLiveRecordId =
    inspectorHistorySubject.kind === "live"
      ? inspectorHistorySubject.recordId
      : null;

  useEffect(() => {
    if (
      activeHistoryLiveRecordId === null ||
      rowHistory.status === "idle" ||
      rowHistory.recordId === activeHistoryLiveRecordId
    ) {
      return;
    }
    void fetchRecordHistory(activeHistoryLiveRecordId, {
      retainedData:
        rowHistory.recordId === activeHistoryLiveRecordId
          ? rowHistory.data
          : null,
      setLoading: true,
    });
  }, [
    activeHistoryLiveRecordId,
    fetchRecordHistory,
    rowHistory.data,
    rowHistory.recordId,
    rowHistory.status,
  ]);

  const clearRowHistory = useCallback(() => {
    rowHistoryRequestSeqRef.current += 1;
    setRowHistoryPendingAction(null);
    setRowHistory({
      recordId: null,
      status: "idle",
      data: null,
      message: null,
    });
  }, [setRowHistory, setRowHistoryPendingAction]);

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
      setRowHistoryPendingAction(null);
      setRowHistory((current) =>
        current.recordId === recordId ? { ...current, message: null } : current,
      );
      saveQueueRef.current = saveQueueRef.current
        .catch(() => undefined)
        .then(async () => {
          const idleRecord = await waitForCommittedRecordIdle(
            recordId,
            idleOptions,
          );
          if (idleRecord === null) {
            clearViewportContinuity(viewportContinuityToken);
            setRowHistory((current) =>
              current.recordId === recordId
                ? {
                    ...current,
                    message: missingVersionMessage,
                  }
                : current,
            );
            finishSave("Conflict");
            return;
          }
          trackPendingSocketTxn(clientTxnId);
          const result = await request(idleRecord.rowVersion, clientTxnId);
          if (!result.ok) {
            resolvePendingSocketTxn(clientTxnId);
            clearViewportContinuity(viewportContinuityToken);
            setRowHistory((current) =>
              current.recordId === recordId
                ? {
                    ...current,
                    message: parseErrorMessage(result.payload),
                  }
                : current,
            );
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
      setRowHistory,
      setRowHistoryPendingAction,
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
          fallbackRowVersion: currentHistoryRowVersion,
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
          if (currentHistoryRecordIdRef.current === recordId) {
            await fetchRecordHistory(recordId);
          }
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
      selectedRow?.recordId,
      setSelectedRowId,
      submitRowHistoryMutation,
    ],
  );

  const submitRowHistoryRollbackTarget = useCallback(
    (
      pending: Extract<RowHistoryPendingAction, { readonly kind: "rollback" }>,
    ) => {
      const { recordId, target } = pending;
      if (recordId.trim() === "") {
        return;
      }
      const viewportContinuityTarget: ViewportContinuityTarget =
        selectedRow?.recordId === recordId
          ? { kind: "row-inspect", recordId }
          : { kind: "scroll-only" };
      submitRowHistoryMutation({
        idleOptions: {
          fallbackRowVersion:
            currentHistoryRecordId === recordId
              ? currentHistoryRowVersion
              : pending.rowVersion,
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
          if (currentHistoryRecordIdRef.current === recordId) {
            await fetchRecordHistory(recordId);
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
      selectedRow?.recordId,
      submitRowHistoryMutation,
    ],
  );

  const previewRowHistoryDeleteRestore = useCallback(
    (operation: "delete" | "restore") => {
      const recordId = currentHistoryRecordId;
      if (recordId === null || recordId === undefined) {
        return;
      }
      setRowHistoryPendingAction({
        kind: "destructive",
        operation,
        recordId,
        rowVersion: currentHistoryRowVersion,
      });
      setRowHistory((current) =>
        current.recordId === recordId ? { ...current, message: null } : current,
      );
    },
    [
      currentHistoryRecordId,
      currentHistoryRowVersion,
      setRowHistory,
      setRowHistoryPendingAction,
    ],
  );

  const previewRowHistoryRollback = useCallback(
    (item: RecordHistoryItem, action: RecordHistoryRollbackAction) => {
      const recordId = currentHistoryRecordId;
      if (recordId === null || recordId === undefined) {
        return;
      }
      const target = buildRecordRollbackTargetFromHistoryAction(item, action);
      if (target === null) {
        return;
      }
      setRowHistoryPendingAction({
        kind: "rollback",
        action,
        historyItemRef: item.history_item_ref,
        recordId,
        rowVersion: currentHistoryRowVersion,
        target,
      });
      setRowHistory((current) =>
        current.recordId === recordId ? { ...current, message: null } : current,
      );
    },
    [
      currentHistoryRecordId,
      currentHistoryRowVersion,
      setRowHistory,
      setRowHistoryPendingAction,
    ],
  );

  const confirmRowHistoryPendingAction = useCallback(() => {
    const pending = rowHistoryPendingAction;
    if (pending === null) {
      return;
    }
    if (pending.kind === "destructive") {
      submitRowHistoryDeleteRestore(pending.operation);
      return;
    }
    submitRowHistoryRollbackTarget(pending);
  }, [
    rowHistoryPendingAction,
    submitRowHistoryDeleteRestore,
    submitRowHistoryRollbackTarget,
  ]);

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
    const requiresEntityRefresh =
      action === "resolve_item" && resolvedRecordId === undefined;
    const entityRefreshBarrier = requiresEntityRefresh
      ? beginTimelineEntityRefreshBarrier(entityCatalogInput)
      : null;
    const viewportContinuityToken = beginViewportContinuity(
      {
        kind: "row-inspect",
        recordId,
      },
      {
        barrier: entityRefreshBarrier,
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
        if (requiresEntityRefresh) {
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
          const committed = applyRowMutation(currentRow.key, envelope, {
            detectAutoResolution: false,
            viewportContinuityToken,
          });
          const expectedEntity = timelineEntityRefreshExpectationForMention(
            committed,
            mention.itemRef,
          );
          if (expectedEntity !== null) {
            advanceViewportContinuity(viewportContinuityToken, {
              barrier: withTimelineEntityRefreshExpectation(
                entityRefreshBarrier,
                expectedEntity,
              ),
            });
          }
          finishSave("Saved");
          let refreshState: TimelineEntityRefreshSettleState = "complete";
          try {
            if (onRefreshEntities === undefined) {
              refreshState = "terminal";
            } else {
              await onRefreshEntities();
            }
          } catch (error) {
            refreshState = "terminal";
            throw error;
          } finally {
            settleViewportContinuityBarrier(
              viewportContinuityToken,
              refreshState,
            );
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

  const resolveTimelinePasteTargetResolution = useCallback(
    (
      rowKey: string,
      fieldKey: string,
      clipboardText: string,
    ): TimelinePasteTargetResolution | null => {
      if (!timelineClipboardShouldDispatchTabular(fieldKey, clipboardText)) {
        return null;
      }

      const dimensions = clipboardGridDimensions(clipboardText);
      const row = rowsRef.current.find((candidate) => candidate.key === rowKey);
      const isDraftTarget =
        row?.recordId === null ||
        (row === undefined && rowKey.startsWith("draft-"));

      if (isDraftTarget) {
        updateWorkbookFocusAnchor(null);
        const targetResolution = resolveDraftTimelinePasteTargets({
          columns: timelineAnchorColumnsRef.current,
          pastedColumnCount: dimensions.columnCount,
          pastedRowCount: dimensions.rowCount,
          startFieldKey: fieldKey,
        });
        return targetResolution === null
          ? null
          : { anchor: null, targetResolution };
      }

      const recordId = row?.recordId;
      if (recordId === undefined || recordId === null) {
        return null;
      }

      const anchor = {
        fieldKey,
        recordId,
      };
      const presentationRows = buildGridPresentationRows({
        getGroupLabel: (candidate, groupFieldKey) =>
          timelineGroupLabel(candidate, groupFieldKey),
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
      if (targetResolution === null) {
        return null;
      }

      updateTimelineFocusAnchor(anchor.recordId, anchor.fieldKey);
      return { anchor, targetResolution };
    },
    [queryState.groupBy, updateTimelineFocusAnchor, updateWorkbookFocusAnchor],
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
      const restoredNow = restoreTimelineFocusAnchor(nextAnchor);
      window.setTimeout(() => {
        if (restoredNow) {
          return;
        }
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
        clearRowHistory();
        restoreTimelineFocusAnchor(anchor);
        return true;
      }
      if (command.kind === "preview-linked-evidence") {
        setSelectedRowId(anchor.recordId);
        setIsInspectorOpen(true);
        setInspectorMessage(
          "Linked evidence preview is unavailable for this row.",
        );
        restoreTimelineFocusAnchor(anchor);
        return true;
      }
      return false;
    },
    [
      clearRowHistory,
      openRowHistory,
      restoreTimelineFocusAnchor,
      setSelectedMentionRef,
      setInspectorMessage,
      setSelectedRowId,
    ],
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
      const binding = timelineScalarBindings.find(
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
          setIsInspectorOpen(true);
          setSelectedMentionRef(mention.itemRef);
          setInspectorMessage(null);
        } else {
          setSelectedRowId(anchor.recordId);
          setIsInspectorOpen(true);
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
      setSelectedMentionRef,
      setInspectorMessage,
      setSelectedRowId,
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
      const binding = timelineScalarBindings.find(
        (candidate) => candidate.key === focusField,
      );
      const fieldKey = binding?.fieldKey ?? focusField;
      if (surface === "grid" && binding !== undefined) {
        const pasteTargetResolution = resolveTimelinePasteTargetResolution(
          rowKey,
          fieldKey,
          clipboardText,
        );
        if (pasteTargetResolution !== null) {
          event.preventDefault();
          const { anchor, targetResolution } = pasteTargetResolution;
          const clientTxnId = nextClientTxnId();
          const viewportContinuityToken = beginViewportContinuity(
            anchor === null
              ? { kind: "scroll-only" }
              : {
                  kind: "input",
                  focusKey: inputFocusKey(rowKey, focusField, surface),
                },
          );
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
              if (anchor !== null) {
                restoreTimelineFocusAnchor(anchor);
              }
              finishSave(
                envelope.data.conflicts && envelope.data.conflicts.length > 0
                  ? "Conflict"
                  : "Saved",
              );
            });
          return;
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
      finishSave,
      incidentId,
      nextClientTxnId,
      queueScalarSave,
      registerSameFieldConflict,
      resolvePendingSocketTxn,
      resolveTimelinePasteTargetResolution,
      restoreTimelineFocusAnchor,
      setScalarEditorDraftValue,
      trackPendingSocketTxn,
      waitForCommittedRecordIdle,
      setActiveConflictKey,
      setPasteConflictGroup,
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
    [
      sendPresenceUpdate,
      setInspectorMessage,
      setCurrentPresence,
      setSelectedRowId,
    ],
  );

  const openInspectorForRow = useCallback(
    (recordId: string) => {
      handleSelectRow(recordId);
      setIsInspectorOpen(true);
    },
    [handleSelectRow],
  );

  const timelineRowForEventTarget = useCallback(
    (target: EventTarget | null) => {
      if (!(target instanceof Element)) {
        return null;
      }
      const rowElement = target.closest<HTMLElement>("[data-grid-record-id]");
      const recordId = rowElement?.dataset.gridRecordId ?? "";
      if (recordId === "") {
        return null;
      }
      return (
        rowsRef.current.find((candidate) => candidate.recordId === recordId) ??
        null
      );
    },
    [],
  );

  const openTimelineRowContextMenu = useCallback(
    (row: WorkbookRow, position: TimelineRowContextMenuPosition) => {
      if (row.recordId === null) {
        return;
      }
      handleSelectRow(row.recordId);
      setRowContextMenu({
        position,
        recordId: row.recordId,
      });
    },
    [handleSelectRow],
  );

  const handleTimelineGridContextMenu = useCallback(
    (event: ReactMouseEvent<HTMLDivElement>) => {
      const row = timelineRowForEventTarget(event.target);
      if (row?.recordId === null || row?.recordId === undefined) {
        return;
      }
      event.preventDefault();
      event.stopPropagation();
      openTimelineRowContextMenu(row, {
        x: event.clientX,
        y: event.clientY,
      });
    },
    [openTimelineRowContextMenu, timelineRowForEventTarget],
  );

  const handleTimelineGridContextKeyDown = useCallback(
    (event: ReactKeyboardEvent<HTMLDivElement>) => {
      if (
        !(
          event.key === "ContextMenu" ||
          (event.key === "F10" && event.shiftKey)
        )
      ) {
        return;
      }
      const row = timelineRowForEventTarget(event.target);
      if (row?.recordId === null || row?.recordId === undefined) {
        return;
      }
      const targetElement =
        event.target instanceof Element
          ? event.target.closest<HTMLElement>(
              "[role='gridcell'], [role='rowheader'], [data-grid-record-id]",
            )
          : null;
      const targetRect = targetElement?.getBoundingClientRect();
      const fallbackRect = event.currentTarget.getBoundingClientRect();
      event.preventDefault();
      event.stopPropagation();
      openTimelineRowContextMenu(row, {
        x: targetRect ? targetRect.left + 12 : fallbackRect.left + 16,
        y: targetRect ? targetRect.top + 12 : fallbackRect.top + 16,
      });
    },
    [openTimelineRowContextMenu, timelineRowForEventTarget],
  );

  useEffect(() => {
    if (rowContextMenu === null) {
      return;
    }
    if (!rows.some((row) => row.recordId === rowContextMenu.recordId)) {
      setRowContextMenu(null);
    }
  }, [rowContextMenu, rows]);

  const focusDraftRow = useCallback(() => {
    const draftSummary = document.querySelector<HTMLInputElement>(
      dataTestIdSelector(draftCellTestId("timeline.summary")),
    );
    draftSummary?.focus({ preventScroll: false });
  }, []);

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
    [sendPresenceUpdate, setCurrentPresence],
  );

  const handleSelectMention = useCallback(
    (rowRecordId: string, itemRef: string) => {
      setSelectedRowId(rowRecordId);
      setSelectedMentionRef(itemRef);
      setInspectorMessage(null);
      setIsInspectorOpen(true);
      window.requestAnimationFrame(() => {
        window.requestAnimationFrame(() => {
          document
            .querySelector<HTMLButtonElement>(
              dataTestIdSelector(mentionItemTestId(itemRef)),
            )
            ?.focus({ preventScroll: true });
        });
      });
    },
    [setSelectedMentionRef, setInspectorMessage, setSelectedRowId],
  );

  function handleUndoAutoResolutionNotice(notice: AutoResolutionNotice) {
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
            entityMentionId: activeItem.itemRef.startsWith("entity_mention:")
              ? activeItem.itemRef.slice("entity_mention:".length) || null
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
  }

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
      setInspectorMessage,
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
          {surface === "grid" ? (
            <TimelineCellPresenceMarker
              fieldKey={binding.fieldKey}
              fieldLabel={timelineBindingLabel(binding.fieldKey)}
              presences={sameCellPresence}
              recordId={row.recordId}
            />
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
      setActiveConflictKey,
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
      const collectionFocusKey = inputFocusKey(
        row.key,
        binding.draftKey,
        "grid",
      );
      const isCollectionInputActive =
        activeCollectionInputKey === collectionFocusKey ||
        row.collectionDrafts[binding.draftKey] !== "";
      const visibleItemLimit = 1;
      const visibleItems = items.slice(0, visibleItemLimit);
      const hiddenItems = items.slice(visibleItemLimit);
      const hiddenItemCount = Math.max(0, items.length - visibleItems.length);
      const activateCollectionInput = () => {
        setActiveCollectionInputKey(collectionFocusKey);
        window.requestAnimationFrame(() => {
          document
            .querySelector<HTMLInputElement>(
              dataTestIdSelector(collectionInputTestId),
            )
            ?.focus({ preventScroll: true });
        });
      };
      const relationshipOverflowRecordId =
        binding.collectionKind === "relationship" ? row.recordId : null;
      return (
        <fieldset
          aria-label={`${label} collection cell`}
          style={collectionCellStyle}
          onClick={(event) => {
            const target = event.target;
            if (
              target instanceof HTMLElement &&
              target.closest("[data-relationship-chip='true']") !== null
            ) {
              return;
            }
            activateCollectionInput();
          }}
          onKeyDown={(event) => {
            if (event.key !== "Enter" && event.key !== "F2") {
              return;
            }
            event.preventDefault();
            activateCollectionInput();
          }}
        >
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
                visibleItems.map((item) => (
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
                visibleItems.map((item) => (
                  <span
                    key={item.itemRef}
                    style={tagChipStyle}
                    title={(item as TagCollectionItem).displayText}
                  >
                    {(item as TagCollectionItem).displayText}
                  </span>
                ))
              )
            ) : (
              <span style={emptyRelationshipStyle}>No items</span>
            )}
            {hiddenItemCount > 0 ? (
              <>
                {relationshipOverflowRecordId !== null ? (
                  <button
                    aria-label={`Inspect ${hiddenItemCount} more ${label.toLowerCase()}`}
                    data-testid={relationshipOverflowButtonTestId(
                      relationshipOverflowRecordId,
                      binding.fieldKey,
                    )}
                    style={collectionOverflowButtonStyle}
                    title={`Inspect ${hiddenItemCount} more ${label.toLowerCase()}`}
                    type="button"
                    onClick={(event) => {
                      event.stopPropagation();
                      const firstHiddenItem = hiddenItems[0];
                      if (firstHiddenItem !== undefined) {
                        handleSelectMention(
                          relationshipOverflowRecordId,
                          firstHiddenItem.itemRef,
                        );
                      } else {
                        openInspectorForRow(relationshipOverflowRecordId);
                      }
                    }}
                    onKeyDown={(event) => {
                      event.stopPropagation();
                    }}
                  >
                    +{hiddenItemCount}
                  </button>
                ) : (
                  <span
                    aria-label={`${hiddenItemCount} more ${label.toLowerCase()}`}
                    role="note"
                    style={collectionOverflowStyle}
                    title={`${hiddenItemCount} more ${label.toLowerCase()}`}
                  >
                    +{hiddenItemCount}
                  </span>
                )}
                <span style={visuallyHiddenStyle}>
                  {hiddenItems
                    .map((item) =>
                      binding.collectionKind === "relationship"
                        ? relationshipItemLabel(
                            item as CollectionItem,
                            entityIndex,
                          )
                        : (item as TagCollectionItem).displayText,
                    )
                    .join(" ")}
                </span>
              </>
            ) : null}
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
            tabIndex={isCollectionInputActive ? 0 : -1}
            style={
              isCollectionInputActive
                ? collectionCellInputStyle
                : collectionCellInactiveInputStyle
            }
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
              if (event.currentTarget.value.trim() === "") {
                setActiveCollectionInputKey((current) =>
                  current === collectionFocusKey ? null : current,
                );
              }
            }}
            onFocus={() => {
              setActiveCollectionInputKey(collectionFocusKey);
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
            placeholder={
              isCollectionInputActive
                ? `Add ${label.toLowerCase()} token`
                : undefined
            }
          />
        </fieldset>
      );
    },
    [
      activeCollectionInputKey,
      entityIndex,
      handleCollectionKeyDown,
      handleSelectRow,
      handleSelectMention,
      openInspectorForRow,
      registerInput,
      timelineBindingLabel,
      queueCollectionSave,
      updateTimelineFocusAnchor,
    ],
  );

  const timelineColumnWidths = useMemo(
    () =>
      buildExpandedTimelineColumnWidths({
        actionsColumnWidth: 0,
        fieldKeys: timelineVisibleFieldKeys,
        gridShellWidth: timelineGridShellWidth,
        rowGutterWidth: timelineRowGutterWidth,
      }),
    [timelineGridShellWidth],
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
          width:
            timelineColumnWidths[binding.fieldKey] ??
            timelineColumnWidth(binding.fieldKey),
          renderCell: (row) => {
            if (binding.kind === "scalar") {
              return renderTimelineGridEditor(row, binding);
            }
            if (binding.kind === "collection") {
              return renderTimelineCollectionInput(row, binding);
            }
            if (binding.fieldKey === "timeline.evidence_count") {
              const countDisplay = buildEvidenceCountDisplayViewModel({
                projectedCount: readTimelineCellValue(
                  row.rawRow,
                  binding.fieldKey,
                ),
                projectedHasEvidence: readTimelineCellValue(
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
              readTimelineCellValue(row.rawRow, binding.fieldKey),
            );
            return (
              <span
                data-testid={
                  row.recordId === null
                    ? undefined
                    : rowCellTestId(row.recordId, binding.fieldKey)
                }
                style={
                  binding.fieldKey === "timeline.edited_at"
                    ? timelineTimestampCellStyle
                    : bodyStyle
                }
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
      timelineColumnWidths,
      timelineBindingLabel,
    ],
  );

  const timelineRowGutter = useMemo(
    () => ({
      label: "",
      width: timelineRowGutterWidth,
      minWidth: timelineRowGutterWidth,
    }),
    [],
  );

  const timelineGridRows = useMemo<readonly GridRow<WorkbookRow>[]>(
    () =>
      rows.map((row, index) => {
        const rowPresence = presenceForRow(row.recordId);
        const ordinal = row.recordId === null ? "+" : String(index + 1);
        return {
          key: row.key,
          recordId: row.recordId,
          data: row,
          gutterContent:
            row.recordId === null ? (
              <DraftRowCreateButton
                row={row}
                onCreate={handleCreateBlankDraftRow}
              />
            ) : (
              <TimelineRowGutterContent
                ordinal={ordinal}
                presences={rowPresence}
                recordId={row.recordId}
              />
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
              ? workbookInlineDraftRowTestId(timelineViewSchemaId)
              : gridRowTestId(timelineViewSchemaId, row.recordId),
          variant: row.recordId === null ? "draft" : "default",
        };
      }),
    [
      handleCreateBlankDraftRow,
      handleSelectRow,
      presenceForRow,
      rows,
      selectedRowId,
    ],
  );

  useLayoutEffect(() => {
    timelineAnchorColumnsRef.current = timelineColumns;
    timelineAnchorRowsRef.current = timelineGridRows;
  }, [timelineColumns, timelineGridRows]);

  const activeRowContextMenuRow = useMemo(
    () =>
      rowContextMenu === null
        ? null
        : (rows.find((row) => row.recordId === rowContextMenu.recordId) ??
          null),
    [rowContextMenu, rows],
  );

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
    const countDisplay = buildEvidenceCountDisplayViewModel({
      projectedCount: readTimelineCellValue(
        row.rawRow,
        "timeline.evidence_count",
      ),
      projectedHasEvidence: readTimelineCellValue(
        row.rawRow,
        "timeline.has_evidence",
      ),
    });
    return (
      <TimelineEvidencePanel
        countDisplay={countDisplay}
        row={row}
        onFilesSelected={handleTimelineEvidenceFiles}
      />
    );
  }

  function renderRowHistorySection() {
    return (
      <TimelineHistoryPanel
        currentRecordId={currentHistoryRecordId}
        history={rowHistory}
        pendingAction={rowHistoryPendingAction}
        selectedActiveRowRecordId={
          inspectorHistorySubject.kind === "live"
            ? inspectorHistorySubject.recordId
            : null
        }
        onCancelPendingAction={() => {
          setRowHistoryPendingAction(null);
        }}
        onConfirmPendingAction={confirmRowHistoryPendingAction}
        onOpenHistory={openRowHistory}
        onPreviewDeleteRestore={previewRowHistoryDeleteRestore}
        onPreviewRollback={previewRowHistoryRollback}
      />
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
    [restoreConflictFocus, setActiveConflictKey],
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
      setPasteConflictGroup,
      setActiveConflictKey,
    ],
  );

  const submitConflictResolution = useCallback(
    (
      conflict: LocalConflictState,
      resolutionKind: TimelineConflictResolution,
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
      setSaveState,
      setSaveStateSecondaryMessage,
    ],
  );

  const handleConflictMergedDraftChange = useCallback(
    (conflictKey: string, value: string) => {
      setConflictQueueState((current) => {
        const conflict = current[conflictKey];
        if (conflict === undefined) {
          return current;
        }
        return {
          ...current,
          [conflictKey]: {
            ...conflict,
            mergedDraft: value,
          },
        };
      });
    },
    [setConflictQueueState],
  );

  const pendingQueueDisplayMessage =
    timelinePendingQueueMessage(pendingQueueSnapshot);
  const visibleRefreshError =
    refreshError !== null && refreshError !== pendingQueueDisplayMessage
      ? refreshError
      : null;

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

  return (
    <WorkbookSurfaceFrame
      inspector={
        isInspectorOpen ? (
          <TimelineWorkbookInspector
            canManageMentions={canManageMentions}
            currentHistoryDeleted={currentHistoryDeleted}
            draftRow={draftRow}
            entityIndex={entityIndex}
            getRelationshipLabel={timelineRelationshipLabel}
            hostEntities={hostEntities}
            identityEntities={identityEntities}
            inspectorMessage={inspectorMessage}
            inspectorMentions={inspectorMentions}
            onClose={() => {
              setIsInspectorOpen(false);
              clearRowHistory();
            }}
            onResolveTargetChange={handleResolveTargetChange}
            onSelectMention={handleSelectMention}
            onSetInspectorMessage={setInspectorMessage}
            onSubmitMentionAction={submitMentionAction}
            renderEvidenceAttachSection={renderEvidenceAttachSection}
            renderInspectorFieldEditors={renderInspectorFieldEditors}
            renderRowHistorySection={renderRowHistorySection}
            rowHistoryRecordId={
              currentHistoryDeleted ? currentHistoryRecordId : null
            }
            selectedMention={selectedMention}
            selectedResolveTargetId={selectedResolveTargetId}
            selectedRow={selectedRow}
          />
        ) : undefined
      }
      primaryGrid={
        <TimelineGridSurface
          columns={timelineColumns}
          getGroupLabel={getTimelineGroupLabel}
          getGroupRowTestId={getTimelineGroupRowTestId}
          groupBy={queryState.groupBy}
          onToggleSort={handleQuerySortToggle}
          ref={gridShellRef}
          rowGutter={timelineRowGutter}
          rows={rows}
          sort={queryState.sort}
          style={timelineGridShellStyle}
          timelineGridRows={timelineGridRows}
        />
      }
      statusStrip={
        <WorkbookStatusStrip
          activeSheetPresenceRecords={activeSheetPresenceRecords}
          inFlightCount={pendingQueueSnapshot.inFlightCount}
          queuedCount={pendingQueueSnapshot.queuedCount}
          saveState={saveState}
          saveStateSecondaryMessage={saveStateSecondaryMessage}
          workbookFocusAnchor={workbookFocusAnchor}
        />
      }
      testId={timelineMutationSubstrateReadyTestId()}
      viewBar={
        <WorkbookSheetToolbar
          leading={
            <>
              {savedViewSelector}
              {renderInlineQueryControls ? (
                <WorkbookGridControls
                  contract={timelineContract}
                  defaultFilterPopoverOpen
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
                  onToggleSort={handleQuerySortToggle}
                  queryState={queryState}
                  surface={timelineViewSchemaId}
                />
              ) : null}
            </>
          }
          onAddRow={focusDraftRow}
          onInspectorToggle={() => {
            setIsInspectorOpen(true);
          }}
          surface={timelineViewSchemaId}
        />
      }
      viewSchemaId={timelineViewSchemaId}
      workAreaAriaLabel="Timeline row interaction layer"
      workAreaOverlays={
        <>
          <TimelineWorkbookNotices
            autoResolutionNotices={autoResolutionNotices}
            entityIndex={entityIndex}
            inspectorOpen={isInspectorOpen}
            onReviewAutoResolution={handleSelectMention}
            onUndoAutoResolution={handleUndoAutoResolutionNotice}
            pendingQueueSnapshot={pendingQueueSnapshot}
          />
          {visibleRefreshError !== null ? (
            <aside
              data-testid="timeline-refresh-error"
              role="status"
              style={timelineNoticeOverlayStyle}
            >
              <p style={bodyStyle}>{visibleRefreshError}</p>
            </aside>
          ) : null}
          {rowContextMenu === null ? null : (
            <TimelineRowContextMenu
              position={rowContextMenu.position}
              replacementDraft={
                activeRowContextMenuRow === null
                  ? ""
                  : (replacementDrafts[activeRowContextMenuRow.key] ?? "")
              }
              row={activeRowContextMenuRow}
              onClose={() => {
                setRowContextMenu(null);
              }}
              onInspectRow={openInspectorForRow}
              onMarkReviewed={(rowKey) => {
                queueAction(rowKey, "mark-reviewed");
              }}
              onOpenHistory={openRowHistory}
              onReplacementDraftChange={(rowKey, value) => {
                setReplacementDrafts((current) => ({
                  ...current,
                  [rowKey]: value,
                }));
              }}
              onSupersede={(rowKey) => {
                queueAction(rowKey, "supersede");
              }}
            />
          )}
          {activeConflict ? (
            <TimelineConflictResolver
              activeConflict={activeConflict}
              activeConflictKey={activeConflictKey}
              activePasteConflictIndex={activePasteConflictIndex}
              activePasteConflictKeys={activePasteConflictKeys}
              conflictQueue={conflictQueue}
              getFieldLabel={timelineBindingLabel}
              onClose={closeConflictResolver}
              onMergedDraftChange={handleConflictMergedDraftChange}
              onSelectConflictKey={(conflictKey) => {
                setActiveConflictKey(conflictKey);
              }}
              onSubmit={submitConflictResolution}
              showPasteConflictNavigator={showPasteConflictNavigator}
            />
          ) : null}
        </>
      }
      onWorkAreaContextMenu={handleTimelineGridContextMenu}
      onWorkAreaKeyDown={handleTimelineGridContextKeyDown}
    />
  );
}

const panelStyle = {
  boxSizing: "border-box" as const,
  width: "100%",
  blockSize: "100%",
  minBlockSize: 0,
  margin: 0,
  padding: 0,
  borderRadius: 0,
  background: "var(--ct-colors-canvas)",
  boxShadow: "none",
  border: 0,
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

const timelineTimestampCellStyle = {
  display: "block",
  minWidth: 0,
  maxWidth: "100%",
  margin: 0,
  overflow: "hidden",
  overflowWrap: "normal" as const,
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
  wordBreak: "normal" as const,
  lineHeight: "var(--ct-typography-grid-cell-lineHeight)",
  color: "var(--ct-colors-ink-muted)",
};

const timelineGridShellStyle = {
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

const gridCellInputStyle = {
  ...inputStyle,
  minHeight: "1.35rem",
  borderColor: "transparent",
  background: "transparent",
  padding: 0,
  color: "var(--ct-colors-ink)",
  lineHeight: 1.2,
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

const sectionTitleStyle = {
  margin: 0,
  fontSize: "1rem",
};

const relationshipItemsWrapStyle = {
  display: "flex",
  alignItems: "center",
  flex: "0 1 auto",
  flexWrap: "nowrap" as const,
  gap: "0.2rem",
  marginBottom: 0,
  maxWidth: "100%",
  minWidth: 0,
  overflow: "hidden",
  whiteSpace: "nowrap" as const,
};

const tagChipStyle = {
  ...relationshipChipBaseStyle,
  flex: "0 1 auto",
  minWidth: 0,
  border: "var(--ct-component-chip-border)",
  background: "var(--ct-component-chip-backgroundColor)",
  color: "var(--ct-component-chip-textColor)",
  textOverflow: "ellipsis",
};

const emptyRelationshipStyle = {
  display: "inline-block",
  minWidth: 0,
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
  color: "var(--ct-colors-ink-tertiary)",
  fontSize: "0.78rem",
};

const collectionCellStyle = {
  position: "relative" as const,
  display: "flex",
  alignItems: "center",
  gap: "0.25rem",
  margin: 0,
  minWidth: 0,
  maxWidth: "100%",
  minBlockSize: "1.2rem",
  padding: 0,
  border: 0,
  overflow: "hidden",
  whiteSpace: "nowrap" as const,
};

const collectionCellInputStyle = {
  ...gridCellInputStyle,
  flex: "1 1 4.5rem",
  minWidth: "4.25rem",
  inlineSize: "auto",
  minHeight: "1.2rem",
  paddingInline: "0.1rem",
};

const collectionCellInactiveInputStyle = {
  ...gridCellInputStyle,
  position: "absolute" as const,
  insetBlockStart: 0,
  insetInlineStart: 0,
  inlineSize: 1,
  blockSize: 1,
  minHeight: 0,
  padding: 0,
  opacity: 0,
  pointerEvents: "none" as const,
};

const collectionOverflowStyle = {
  flex: "0 0 auto",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-pill)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink-muted)",
  fontFamily: "var(--ct-typography-mono-fontFamily)",
  fontSize: "0.68rem",
  fontWeight: 700,
  lineHeight: 1.1,
  padding: "0.12rem 0.35rem",
};

const collectionOverflowButtonStyle = {
  ...collectionOverflowStyle,
  appearance: "none" as const,
  cursor: "pointer",
};

const inspectorActionStackStyle = {
  display: "grid",
  gap: "0.75rem",
};

const noticeCardStyle = {
  position: "absolute" as const,
  zIndex: 7,
  insetBlockStart:
    "calc(var(--ct-layout-viewBarHeight) + var(--ct-spacing-sm))",
  insetInlineEnd: "var(--ct-spacing-sm)",
  maxInlineSize: "min(36rem, calc(100% - var(--ct-spacing-xl)))",
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  padding: "0.85rem 1rem",
  display: "grid",
  gap: "0.5rem",
  boxShadow: "var(--ct-elevation-popover)",
};

const timelineNoticeOverlayStyle = {
  ...noticeCardStyle,
  ...workbookSurfaceOverlayPanelStyle,
  insetBlockStart: "var(--ct-spacing-sm)",
  insetInlineEnd: "auto",
} satisfies CSSProperties;
