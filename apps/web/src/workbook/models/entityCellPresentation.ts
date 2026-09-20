import type { ViewContract } from "@cartulary/view-contracts";
import type { EntityRow } from "./entityWorkbookModel";
import { genericCellPresentation } from "./genericWorkbookModel";

/** Committed entity renderer semantics, before clipping; never local edit overlays. */
export function entityCellPresentation(row: EntityRow, fieldKey: string) {
  const prefix = row.entityType;
  if (fieldKey === `${prefix}.display_name`) {
    const named = row.rawRow.cells[fieldKey]?.value;
    return {
      text: row.label,
      fragments: named || row.secondaryText ? [row.label] : [],
    };
  }
  if (fieldKey === `${prefix}.${prefix === "host" ? "hostname" : "upn"}`)
    return {
      text: row.secondaryText || "None",
      fragments: row.secondaryText ? [row.secondaryText] : [],
    };
  if (
    fieldKey ===
    `${prefix}.${prefix === "host" ? "host_state" : "identity_state"}`
  )
    return { text: row.state, fragments: row.state ? [row.state] : [] };
  if (fieldKey === `${prefix}.aliases`)
    return {
      text: row.aliasTexts.length ? row.aliasTexts.join(", ") : "No aliases",
      fragments: row.aliasTexts,
    };
  if (fieldKey === "row_version")
    return { text: String(row.rowVersion), fragments: [] };
  return genericCellPresentation(row.rawRow.cells[fieldKey]?.value);
}
export function entityFindText(
  row: EntityRow,
  fieldKey: string,
  contract: ViewContract,
): readonly string[] {
  if (
    fieldKey === "record_id" ||
    fieldKey === "row_version" ||
    !contract.fieldMap[fieldKey]
  )
    return [];
  return entityCellPresentation(row, fieldKey).fragments;
}
