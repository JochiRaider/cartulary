import type { WorkbookInspectorErrorPresentation } from "../../inspector/workbookInspectorErrorModel";
import type { WorkbookRow } from "./workbookTimelineModel";

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

export type RecordHistoryData = {
  incident_id: string;
  record_id: string;
  row_version: number;
  deleted: boolean;
  items: RecordHistoryItem[];
};

const rollbackActionOrder = [
  "history_entry",
  "change_set",
  "row_restore",
] as const satisfies readonly RecordHistoryRollbackAction[];

export function normalizeRecordHistoryData(
  data: RecordHistoryData,
): RecordHistoryData | null {
  const seen = new Set<string>();
  for (const item of data.items) {
    if (
      item.history_item_ref.trim() === "" ||
      item.change_set_id.trim() === "" ||
      seen.has(item.history_item_ref)
    ) {
      return null;
    }
    seen.add(item.history_item_ref);
    let previous = -1;
    for (const action of item.available_rollback_actions) {
      const index = rollbackActionOrder.indexOf(action);
      if (index <= previous || !validItemSelector(item, action)) return null;
      previous = index;
    }
  }
  return data;
}

function validItemSelector(
  item: RecordHistoryItem,
  action: RecordHistoryRollbackAction,
): boolean {
  if (action === "history_entry") {
    return (
      typeof item.history_entry_ref === "string" &&
      item.history_entry_ref !== ""
    );
  }
  if (action === "change_set") return item.change_set_id !== "";
  return Number.isInteger(item.revision_no) && (item.revision_no ?? 0) > 0;
}

export type RecordHistoryState = {
  recordId: string | null;
  status: "idle" | "loading" | "ready" | "error";
  data: RecordHistoryData | null;
  error: WorkbookInspectorErrorPresentation | null;
};

export type RowHistoryPendingAction =
  | {
      kind: "rollback";
      action: RecordHistoryRollbackAction;
      historyItemRef: string;
      recordId: string;
      rowVersion: number | null;
      target: Record<string, unknown>;
    }
  | {
      kind: "destructive";
      operation: "delete" | "restore";
      recordId: string;
      rowVersion: number | null;
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

function isPositiveInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value > 0;
}

export type TimelineInspectorHistorySubject =
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

type TimelineHistoryStateLike = {
  readonly recordId: string | null;
  readonly data: {
    readonly deleted?: boolean | undefined;
    readonly record_id: string;
    readonly row_version: number;
  } | null;
};

export function selectTimelineInspectorHistorySubject({
  draftRow,
  rowHistory,
  selectedRow,
}: {
  readonly draftRow: WorkbookRow | null;
  readonly rowHistory: TimelineHistoryStateLike;
  readonly selectedRow: WorkbookRow | null;
}): TimelineInspectorHistorySubject {
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
  if (deletedRowIsActiveSubject && deletedRowHistoryData !== null) {
    return {
      kind: "deleted",
      recordId: deletedRowHistoryData.record_id,
      rowVersion: deletedRowHistoryData.row_version,
    };
  }
  if (selectedLiveRecordId !== null) {
    return {
      kind: "live",
      recordId: selectedLiveRecordId,
      rowVersion: selectedRow?.rowVersion ?? null,
    };
  }
  if (draftRow !== null) {
    return { kind: "draft" };
  }
  return { kind: "none" };
}
