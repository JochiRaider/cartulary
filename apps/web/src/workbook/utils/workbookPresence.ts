import type { SheetRef } from "../../shared/sheetRef";
import { isSheetRef, sheetRefsEqual } from "../../shared/sheetRef";

export type WorkbookPresenceMode = "editing" | "idle" | "viewing";

export type WorkbookPresenceInput = {
  sheet_ref: SheetRef;
  mode: WorkbookPresenceMode;
  record_id?: string;
  field_key?: string;
};

export type PresenceRecord = WorkbookPresenceInput & {
  connection_id: string;
  display_name: string;
  expires_at: string;
  observed_at: string;
  user_id: string;
};

export function presenceMatchesSheet(
  presence: PresenceRecord,
  sheetRef: SheetRef,
) {
  return sheetRefsEqual(presence.sheet_ref, sheetRef);
}

export function displayInitials(displayName: string) {
  const parts = displayName.trim().split(/\s+/u).filter(Boolean);
  if (parts.length === 0) {
    return "?";
  }
  return parts
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? "")
    .join("");
}

export function visiblePresence(
  records: readonly PresenceRecord[],
  limit: number,
) {
  const sorted = [...records].sort(comparePresenceRecord);
  return {
    shown: sorted.slice(0, limit),
    overflow: Math.max(0, records.length - limit),
  };
}

function comparePresenceRecord(left: PresenceRecord, right: PresenceRecord) {
  const displayNameComparison = left.display_name.localeCompare(
    right.display_name,
    undefined,
    { sensitivity: "base" },
  );
  if (displayNameComparison !== 0) {
    return displayNameComparison;
  }
  const userIdComparison = left.user_id.localeCompare(right.user_id);
  if (userIdComparison !== 0) {
    return userIdComparison;
  }
  return left.connection_id.localeCompare(right.connection_id);
}

export function isPresenceRecord(value: unknown): value is PresenceRecord {
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
    isSheetRef(sheetRef) &&
    (record.record_id === undefined || typeof record.record_id === "string") &&
    (record.field_key === undefined || typeof record.field_key === "string")
  );
}
