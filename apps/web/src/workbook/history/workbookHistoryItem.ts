import type {
  RecordHistoryData,
  RecordHistoryItem,
} from "./workbookHistoryPage";
export type RecordHistoryRollbackAction =
  RecordHistoryItem["available_rollback_actions"][number];

export type RecordHistoryRollbackTarget =
  | { readonly kind: "history_entry"; readonly history_entry_ref: string }
  | { readonly kind: "change_set"; readonly change_set_id: string }
  | { readonly kind: "row_restore"; readonly restore_to_revision_no: number };

export type WorkbookRecordHistoryPendingAction =
  | {
      readonly kind: "rollback";
      readonly action: RecordHistoryRollbackAction;
      readonly historyItemRef: string;
      readonly recordId: string;
      readonly rowVersion: number;
      readonly target: RecordHistoryRollbackTarget;
    }
  | {
      readonly kind: "destructive";
      readonly operation: "delete" | "restore";
      readonly recordId: string;
      readonly rowVersion: number;
    };

const rollbackActionOrder = [
  "history_entry",
  "change_set",
  "row_restore",
] as const satisfies readonly RecordHistoryRollbackAction[];
export function normalizeRecordHistoryData(
  data: RecordHistoryData,
): RecordHistoryData | null {
  if (data.record_id.trim() === "" || !isPositiveInteger(data.row_version)) {
    return null;
  }
  const seen = new Set<string>();
  for (const item of data.items) {
    if (
      (!item.reversible && item.available_rollback_actions.length > 0) ||
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

export function buildRecordRollbackTargetFromHistoryAction(
  item: RecordHistoryItem,
  action: RecordHistoryRollbackAction,
): RecordHistoryRollbackTarget | null {
  if (!item.available_rollback_actions.includes(action)) return null;
  switch (action) {
    case "history_entry":
      return typeof item.history_entry_ref === "string" &&
        item.history_entry_ref.trim() !== ""
        ? { history_entry_ref: item.history_entry_ref, kind: "history_entry" }
        : null;
    case "change_set":
      return item.change_set_id.trim() === ""
        ? null
        : { change_set_id: item.change_set_id, kind: "change_set" };
    case "row_restore":
      return isPositiveInteger(item.revision_no)
        ? { kind: "row_restore", restore_to_revision_no: item.revision_no }
        : null;
  }
}

function validItemSelector(
  item: RecordHistoryItem,
  action: RecordHistoryRollbackAction,
): boolean {
  switch (action) {
    case "history_entry":
      return (
        typeof item.history_entry_ref === "string" &&
        item.history_entry_ref.trim() !== ""
      );
    case "change_set":
      return item.change_set_id.trim() !== "";
    case "row_restore":
      return isPositiveInteger(item.revision_no);
  }
}

function isPositiveInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value > 0;
}

export function historyTargetEqual(
  left: RecordHistoryRollbackTarget,
  right: RecordHistoryRollbackTarget,
): boolean {
  if (left.kind !== right.kind) return false;
  switch (left.kind) {
    case "history_entry":
      return (
        right.kind === "history_entry" &&
        left.history_entry_ref === right.history_entry_ref
      );
    case "change_set":
      return (
        right.kind === "change_set" &&
        left.change_set_id === right.change_set_id
      );
    case "row_restore":
      return (
        right.kind === "row_restore" &&
        left.restore_to_revision_no === right.restore_to_revision_no
      );
  }
}

export function historyItemContentEqual(
  a: RecordHistoryItem,
  b: RecordHistoryItem,
): boolean {
  // Eligibility and revision selectors are current-state observations. The
  // committed content and any previously issued entry selector are immutable.
  return (
    a.actor_user_id === b.actor_user_id &&
    a.committed_at === b.committed_at &&
    a.history_item_ref === b.history_item_ref &&
    a.operation === b.operation &&
    a.change_set_id === b.change_set_id &&
    canonical(a.diff_summary) === canonical(b.diff_summary) &&
    (a.history_entry_ref === undefined ||
      a.history_entry_ref === b.history_entry_ref)
  );
}
function canonical(value: unknown): string {
  if (Array.isArray(value)) return `[${value.map(canonical).join(",")}]`;
  if (value && typeof value === "object")
    return `{${Object.entries(value)
      .sort(([a], [b]) => (a < b ? -1 : a > b ? 1 : 0))
      .map(([key, child]) => `${JSON.stringify(key)}:${canonical(child)}`)
      .join(",")}}`;
  return JSON.stringify(value) ?? "null";
}
