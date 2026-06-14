import type { WorkbookSheetRef } from "./workbookStartup";

export type WorkbookPresenceMode = "editing" | "idle" | "viewing";

export type WorkbookPresenceInput = {
  sheet_ref: WorkbookSheetRef;
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
  sheetRef: WorkbookSheetRef,
) {
  return (
    presence.sheet_ref.kind === sheetRef.kind &&
    presence.sheet_ref.id === sheetRef.id
  );
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
  return {
    shown: records.slice(0, limit),
    overflow: Math.max(0, records.length - limit),
  };
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
