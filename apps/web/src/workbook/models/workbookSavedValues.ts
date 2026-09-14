import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";

function canonical(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(canonical);
  if (value !== null && typeof value === "object")
    return Object.fromEntries(
      Object.entries(value)
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([key, item]) => [key, canonical(item)]),
    );
  return value ?? null;
}
export function workbookSavedFieldEqual(
  a: WorkbookQueryRow,
  b: WorkbookQueryRow,
  field: string,
): boolean {
  return (
    JSON.stringify(canonical(a.cells[field]?.value)) ===
    JSON.stringify(canonical(b.cells[field]?.value))
  );
}
