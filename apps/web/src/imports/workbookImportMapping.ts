import type {
  DiscoveredImportPreview,
  DiscoveredImportUnit,
  ImportMappingRequest,
} from "../services/importContractAdapter";
import { workbookImportTargets } from "../services/importTargetContractAdapter";

export type WorkbookMappingDraft = {
  readonly targetViewSchemaId: string;
  readonly fields: Readonly<Record<number, string>>;
  readonly dirty: boolean;
};
export type ImportRectangle = {
  readonly startRow: number;
  readonly startColumn: number;
  readonly endRow: number;
  readonly endColumn: number;
};
const token = (text: string | number | boolean | null) =>
  String(text ?? "")
    .toLowerCase()
    .replace(/[^a-z0-9]/g, "");
export function createWorkbookMappingDraft(
  preview: DiscoveredImportPreview,
  unit: DiscoveredImportUnit,
): WorkbookMappingDraft {
  const approved = unit.approved_mapping;
  if (approved && "target_view_schema_id" in approved)
    return {
      targetViewSchemaId: approved.target_view_schema_id,
      fields: Object.fromEntries(
        approved.source_columns.map((c) => [
          c.source_column_ordinal,
          c.field_key ?? "",
        ]),
      ),
      dirty: false,
    };
  const headers = new Set(
    preview.columns.map((c) => token(c.source_header_text)),
  );
  const targets = workbookImportTargets.map((target) => ({
    target,
    score: target.fields.filter(
      (f) =>
        headers.has(token(f.label)) ||
        headers.has(token(f.fieldKey.split(".").at(-1) ?? "")),
    ).length,
  }));
  targets.sort((a, b) => b.score - a.score);
  return suggestWorkbookMapping(
    preview,
    targets[0]?.target.contract.viewSchemaId ?? "",
  );
}
export function suggestWorkbookMapping(
  preview: DiscoveredImportPreview,
  targetViewSchemaId: string,
): WorkbookMappingDraft {
  const target = workbookImportTargets.find(
    (t) => t.contract.viewSchemaId === targetViewSchemaId,
  );
  const fields: Record<number, string> = {},
    used = new Set<string>();
  for (const column of preview.columns) {
    const header = token(column.source_header_text);
    const field =
      header === ""
        ? undefined
        : target?.fields.find(
            (f) =>
              !used.has(f.fieldKey) &&
              (token(f.label) === header ||
                token(f.fieldKey.split(".").at(-1) ?? "") === header),
          );
    if (field) {
      fields[column.source_column_ordinal] = field.fieldKey;
      used.add(field.fieldKey);
    }
  }
  return { targetViewSchemaId, fields, dirty: true };
}
export function workbookMappingErrors(
  draft: WorkbookMappingDraft,
  preview: DiscoveredImportPreview,
): Readonly<Record<number, string>> {
  const target = workbookImportTargets.find(
    (t) => t.contract.viewSchemaId === draft.targetViewSchemaId,
  );
  const errors: Record<number, string> = {},
    used = new Map<string, number>();
  for (const column of preview.columns) {
    const ordinal = column.source_column_ordinal,
      key = draft.fields[ordinal];
    if (!key) {
      if (
        target?.semantics.default_unknown_column_policy === "reject_if_unmapped"
      )
        errors[ordinal] = "Map this column for the selected target.";
      continue;
    }
    if (!target?.fields.some((f) => f.fieldKey === key))
      errors[ordinal] = "This field does not permit import values.";
    const earlier = used.get(key);
    if (earlier !== undefined) {
      errors[earlier] = "Each target field can be mapped only once.";
      errors[ordinal] = errors[earlier];
    }
    used.set(key, ordinal);
  }
  return errors;
}
export function workbookMappingRequest(
  unit: DiscoveredImportUnit,
  preview: DiscoveredImportPreview,
  draft: WorkbookMappingDraft,
  transactionId: string,
): ImportMappingRequest | null {
  const target = workbookImportTargets.find(
    (t) => t.contract.viewSchemaId === draft.targetViewSchemaId,
  );
  if (!target || Object.keys(workbookMappingErrors(draft, preview)).length)
    return null;
  const approved = unit.approved_mapping;
  return {
    client_txn_id: transactionId,
    target_view_schema_id: target.contract.viewSchemaId,
    unknown_column_policy: target.semantics.default_unknown_column_policy,
    header_row_ref: unit.header_row_ref,
    data_start_row_ref: unit.data_start_row_ref,
    source_columns: preview.columns.map((column) => {
      const field = target.fields.find(
        (f) => f.fieldKey === draft.fields[column.source_column_ordinal],
      );
      const previous =
        approved &&
        "target_view_schema_id" in approved &&
        approved.target_view_schema_id === target.contract.viewSchemaId
          ? approved.source_columns.find(
              (c) =>
                c.source_column_ordinal === column.source_column_ordinal &&
                c.field_key === field?.fieldKey,
            )
          : undefined;
      return {
        ...column,
        field_key: field?.fieldKey ?? null,
        entity_binding_mode: field?.entityBindingMode ?? null,
        transform_id: previous?.transform_id ?? null,
        transform_options: previous?.transform_options ?? {},
        empty_value_policy: previous?.empty_value_policy ?? "omit_field",
      };
    }),
  };
}

/** Source rectangle geometry only; no workbook parsing or mapping execution. */
export function importRectangle(value: string): ImportRectangle | null {
  const match = /^([A-Z]+)([1-9][0-9]*)(?::([A-Z]+)([1-9][0-9]*))?$/.exec(
    value,
  );
  if (!match) return null;
  const column = (text: string) =>
    [...text].reduce((n, c) => n * 26 + c.charCodeAt(0) - 64, 0);
  const rect = {
    startRow: Number(match[2]),
    startColumn: column(match[1] ?? ""),
    endRow: Number(match[4] ?? match[2]),
    endColumn: column(match[3] ?? match[1] ?? ""),
  };
  return Object.values(rect).every(Number.isSafeInteger) &&
    rect.startRow <= rect.endRow &&
    rect.startColumn <= rect.endColumn
    ? rect
    : null;
}
export function validImportRegion(
  rect: ImportRectangle,
  unit: DiscoveredImportUnit,
): boolean {
  const base = importRectangle(unit.source_rect_a1);
  return (
    unit.locator_kind === "xlsx_used_range" &&
    base !== null &&
    Object.values(rect).every((v) => Number.isSafeInteger(v) && v > 0) &&
    rect.startRow >= base.startRow &&
    rect.startColumn >= base.startColumn &&
    rect.endRow <= base.endRow &&
    rect.endColumn <= base.endColumn &&
    rect.endRow > rect.startRow &&
    rect.endColumn >= rect.startColumn
  );
}
export function overlappingImportUnits(
  units: readonly DiscoveredImportUnit[],
): readonly string[] {
  const overlaps = new Set<string>();
  for (let i = 0; i < units.length; i++)
    for (let j = i + 1; j < units.length; j++) {
      const a = units[i],
        b = units[j];
      if (!a || !b || a.locator.sheet_name !== b.locator.sheet_name) continue;
      const ra = importRectangle(a.source_rect_a1),
        rb = importRectangle(b.source_rect_a1);
      if (
        !ra ||
        !rb ||
        (ra.startColumn <= rb.endColumn &&
          rb.startColumn <= ra.endColumn &&
          ra.startRow <= rb.endRow &&
          rb.startRow <= ra.endRow)
      ) {
        overlaps.add(a.import_unit_id);
        overlaps.add(b.import_unit_id);
      }
    }
  return [...overlaps];
}
