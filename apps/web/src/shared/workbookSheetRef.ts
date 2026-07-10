export type WorkbookSheetRef =
  | {
      readonly id: string;
      readonly kind: "saved_view";
    }
  | {
      readonly id: string;
      readonly kind: "view_schema";
    }
  | {
      readonly extension_profile_id: string;
      readonly kind: "extension_workspace";
      readonly workspace_key: string;
    };

const exactKeys = (record: Record<string, unknown>, expected: string[]) => {
  const keys = Object.keys(record).sort();
  return (
    keys.length === expected.length &&
    keys.every((key, index) => key === expected[index])
  );
};

const nonEmptyString = (value: unknown): value is string =>
  typeof value === "string" && value.trim() !== "";

export function isWorkbookSheetRef(value: unknown): value is WorkbookSheetRef {
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

export function workbookSheetRefKey(sheetRef: WorkbookSheetRef): string {
  if (sheetRef.kind === "extension_workspace") {
    return `${sheetRef.kind}:${sheetRef.extension_profile_id}:${sheetRef.workspace_key}`;
  }
  return `${sheetRef.kind}:${sheetRef.id}`;
}

export function workbookSheetRefsEqual(
  left: WorkbookSheetRef,
  right: WorkbookSheetRef,
): boolean {
  return workbookSheetRefKey(left) === workbookSheetRefKey(right);
}
