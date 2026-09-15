import {
  getReferenceFieldContract,
  type ViewFieldContract,
} from "@cartulary/view-contracts";
import type { WorkbookProtocolCollectionActions } from "../adapters/workbookProtocolTypes";
import {
  buildGenericPatchChange,
  type GenericCollectionMode,
} from "../models/genericWorkbookModel";
import {
  exactWorkbookReferenceId,
  normalizeWorkbookAuthoringText,
  validWorkbookTimestamp,
} from "../models/workbookAuthoringValues";

/** Existing-record admission is independent of create payload serialization. */
export function prepareWorkbookInspectorChange(
  field: ViewFieldContract,
  raw: string | null,
  mode: GenericCollectionMode,
  viewSchemaId: string,
) {
  if (!field.patchWritable || field.writeKind === "read_only")
    return { error: "This field is unavailable for editing." } as const;
  if (raw === null && !field.clearable)
    return { error: "This field cannot be cleared." } as const;
  if (field.writeKind === "direct_value" && field.directReferenceContractId) {
    if (raw !== null && !exactWorkbookReferenceId(raw))
      return {
        error: "Enter or select an exact reference identifier.",
      } as const;
    return { change: { field_key: field.fieldKey, value: raw } } as const;
  }
  const reference = getReferenceFieldContract(viewSchemaId, field.fieldKey);
  if (reference?.kind === "collection") {
    const values = raw?.split("\n") ?? [];
    if (
      !values.length ||
      values.length > 64 ||
      values.some(
        (value) =>
          !value || (mode === "add" && !exactWorkbookReferenceId(value)),
      )
    )
      return {
        error: "Provide 1 to 64 exact reference actions for this edit.",
      } as const;
    const actions: WorkbookProtocolCollectionActions["actions"][number][] =
      values.map((value) => {
        if (mode === "remove")
          return reference.removeOperation === "remove_party_ref"
            ? { op: "remove_party_ref", item_ref: value }
            : { op: "remove_record_ref", item_ref: value };
        return reference.addOperation === "add_party_ref"
          ? { op: "add_party_ref", party_id: value }
          : { op: "add_record_ref", linked_record_id: value };
      });
    const [first, ...rest] = actions;
    if (!first) return { error: "Select a reference action." } as const;
    const action_payload: WorkbookProtocolCollectionActions = {
      kind: "collection_actions_v1",
      actions: [first, ...rest],
    };
    return { change: { field_key: field.fieldKey, action_payload } } as const;
  }
  let value = raw ?? "";
  if (field.writeKind === "direct_value" && field.stringContractId) {
    const normalized = normalizeWorkbookAuthoringText(
      value,
      field.stringContractId,
    );
    if (normalized.error) return { error: normalized.error } as const;
    value = normalized.value;
  }
  if (
    value &&
    field.directReferenceContractId &&
    !exactWorkbookReferenceId(value)
  )
    return { error: "Select a valid reference." } as const;
  if (value && field.enumValues?.length && !field.enumValues.includes(value))
    return { error: "Select an available value." } as const;
  if (
    value &&
    field.directScalarContractId === "timestamp_instant_v1" &&
    !validWorkbookTimestamp(value)
  )
    return { error: "Enter a valid RFC3339 timestamp." } as const;
  const change = buildGenericPatchChange(field, value, mode, viewSchemaId);
  if (!change)
    return {
      error: "Provide a valid value, or clear this field if permitted.",
    } as const;
  if (change.action_payload) {
    for (const action of change.action_payload.actions) {
      for (const [key, item] of Object.entries(action)) {
        if (
          ["party_id", "linked_record_id"].includes(key) &&
          !exactWorkbookReferenceId(String(item))
        )
          return { error: "Select a valid item or reference." } as const;
        const contract =
          key === "alias_text"
            ? "alias_text_v1"
            : key === "tag_name"
              ? "tag_label_v1"
              : key === "risk_ref_text"
                ? "single_line_title_v1"
                : null;
        if (contract) {
          const normalized = normalizeWorkbookAuthoringText(
            String(item),
            contract,
          );
          if (normalized.error) return { error: normalized.error } as const;
          Object.assign(action, { [key]: normalized.value });
        }
      }
    }
  }
  return { change } as const;
}
