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
    .map((part) => Array.from(part)[0]?.toUpperCase() ?? "")
    .join("");
}

export function isPresenceRecord(value: unknown): value is PresenceRecord {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return false;
  }
  if (
    !("connection_id" in value) ||
    !("display_name" in value) ||
    !("expires_at" in value) ||
    !("mode" in value) ||
    !("observed_at" in value) ||
    !("sheet_ref" in value) ||
    !("user_id" in value)
  ) {
    return false;
  }
  const sheetRef = value.sheet_ref;
  return (
    typeof value.connection_id === "string" &&
    value.connection_id.length > 0 &&
    typeof value.user_id === "string" &&
    value.user_id.length > 0 &&
    typeof value.display_name === "string" &&
    typeof value.mode === "string" &&
    (value.mode === "viewing" ||
      value.mode === "editing" ||
      value.mode === "idle") &&
    typeof value.observed_at === "string" &&
    typeof value.expires_at === "string" &&
    isSheetRef(sheetRef) &&
    (!("record_id" in value) ||
      (typeof value.record_id === "string" && value.record_id.length > 0)) &&
    (!("field_key" in value) ||
      (typeof value.field_key === "string" &&
        value.field_key.length > 0 &&
        value.mode === "editing")) &&
    (sheetRef.kind !== "extension_workspace" ||
      (!("record_id" in value) && !("field_key" in value)))
  );
}
