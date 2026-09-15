import type { ViewFieldContract } from "@cartulary/view-contracts";
import type { WorkbookProtocolPatchRecordRequest } from "../adapters/workbookProtocolTypes";
import {
  exactWorkbookReferenceId,
  normalizeWorkbookAuthoringText,
  validWorkbookTimestamp,
} from "./workbookAuthoringValues";

/** Validation constructs intent; it never rewrites the retained raw draft. */
export function workbookGridEditChange(
  field: ViewFieldContract,
  raw: string | null,
): WorkbookProtocolPatchRecordRequest["changes"][number] | null {
  if (!field.gridEditable || field.writeKind !== "direct_value") return null;
  if (raw === null)
    return field.clearable ? { field_key: field.fieldKey, value: null } : null;
  let value: string | number | boolean | null = raw;
  if (field.directScalarContractId === "timestamp_instant_v1") {
    if (!validWorkbookTimestamp(raw)) return null;
  } else if (field.directReferenceContractId) {
    if (!exactWorkbookReferenceId(raw)) return null;
  } else if (field.enumValues?.length) {
    if (!field.enumValues.includes(raw)) return null;
  } else if (field.readKind === "number") {
    if (!/^-?\d+$/u.test(raw) || !Number.isSafeInteger(Number(raw)))
      return null;
    value = Number(raw);
    if (
      field.fieldKey === "finding.confidence_score" &&
      (value < 0 || value > 100)
    )
      return null;
  } else if (field.readKind === "boolean") {
    if (raw !== "true" && raw !== "false") return null;
    value = raw === "true";
  } else if (
    field.stringContractId &&
    field.stringContractId !== "timeline_visible_text_v1"
  ) {
    const normalized = normalizeWorkbookAuthoringText(
      raw,
      field.stringContractId,
    );
    if (normalized.error) return null;
    value = normalized.value;
    if (value === "") {
      if (!field.clearable) return null;
      value = null;
    }
  } else if (raw === "" && !field.clearable) return null;
  return { field_key: field.fieldKey, value };
}
