export type WorkbookSheetRef = {
  readonly id: string;
  readonly kind: "saved_view" | "view_schema";
};

export function isWorkbookSheetRef(value: unknown): value is WorkbookSheetRef {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return false;
  }
  const record = value as Record<string, unknown>;
  return (
    (record.kind === "view_schema" || record.kind === "saved_view") &&
    typeof record.id === "string" &&
    record.id.trim() !== ""
  );
}
