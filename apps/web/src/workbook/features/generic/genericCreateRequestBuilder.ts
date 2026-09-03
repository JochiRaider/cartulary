import type { ViewContract } from "@cartulary/view-contracts";
import {
  buildGenericPatchChange,
  normalizeGenericTextValue,
  workbookCreateMinimumSatisfied,
  workbookCreationAvailable,
} from "../../models/genericWorkbookModel";

type GenericCreateRequest = Record<string, unknown> & {
  readonly client_txn_id: string;
};

export function buildGenericCreateRequest(
  contract: ViewContract,
  draft: Readonly<Record<string, string>>,
  clientTxnId: string,
): GenericCreateRequest | null {
  if (
    !workbookCreationAvailable(contract) ||
    !workbookCreateMinimumSatisfied(contract, { ...draft })
  ) {
    return null;
  }
  const request: Record<string, unknown> = { client_txn_id: clientTxnId };
  for (const field of contract.fields.filter((entry) => entry.createWritable)) {
    const rawValue = draft[field.fieldKey] ?? "";
    const change = buildGenericPatchChange(
      field,
      rawValue,
      "add",
      contract.viewSchemaId,
    );
    if (change === null) {
      if (normalizeGenericTextValue(rawValue) !== "") return null;
      continue;
    }
    request[field.fieldKey] =
      "action_payload" in change ? change.action_payload : change.value;
  }
  for (const input of contract.createInputs) {
    const value = normalizeGenericTextValue(draft[input.inputKey] ?? "");
    if (value === "") {
      if (input.required) return null;
      continue;
    }
    request[input.inputKey] = value;
  }
  return Object.keys(request).length > 1 || contract.permitsZeroFieldCreate
    ? (request as GenericCreateRequest)
    : null;
}
