import type { SheetRef } from "@cartulary/protocol-ts/http";

export type { SheetRef } from "@cartulary/protocol-ts/http";

const exactKeys = (record: Record<string, unknown>, expected: string[]) => {
  const keys = Object.keys(record).sort();
  return (
    keys.length === expected.length &&
    keys.every((key, index) => key === expected[index])
  );
};

const nonEmptyString = (value: unknown): value is string =>
  typeof value === "string" && value.trim() !== "";

export function isSheetRef(value: unknown): value is SheetRef {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return false;
  }
  const record = value as Record<string, unknown>;
  if (record.kind === "view_schema" || record.kind === "saved_view") {
    return exactKeys(record, ["id", "kind"]) && nonEmptyString(record.id);
  }
  if (record.kind === "extension_workspace") {
    return (
      exactKeys(record, ["extension_profile_id", "kind", "workspace_key"]) &&
      nonEmptyString(record.extension_profile_id) &&
      nonEmptyString(record.workspace_key)
    );
  }
  return false;
}

export function sheetRefKey(sheetRef: SheetRef): string {
  if (sheetRef.kind === "extension_workspace") {
    return `${sheetRef.kind}:${sheetRef.extension_profile_id}:${sheetRef.workspace_key}`;
  }
  return `${sheetRef.kind}:${sheetRef.id}`;
}

export function sheetRefsEqual(left: SheetRef, right: SheetRef): boolean {
  return sheetRefKey(left) === sheetRefKey(right);
}
