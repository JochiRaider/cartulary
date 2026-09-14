import type {
  ViewContract,
  ViewFieldContract,
} from "@cartulary/view-contracts";
import { genericCreateMinimumMessage } from "../../models/genericWorkbookModel";
import {
  exactWorkbookReferenceId,
  normalizeWorkbookAuthoringText,
  validWorkbookTimestamp,
} from "../../models/workbookAuthoringValues";
import { decodeCreateViewRowRequest } from "../../models/workbookRequestDecoders";
import { requireWorkbookSurfaceRegistration } from "../../models/workbookSurfaceRegistration";
import type {
  OrdinaryCreatePreparation,
  OrdinaryCreateValues,
} from "./ordinaryCreateContract";

/** Field serialization is neutral. Source contributions own minima and additional admission. */
export function prepareOrdinaryCreateFields(
  contract: ViewContract,
  values: OrdinaryCreateValues,
  clientTxnId: string,
  options: {
    scalar?: (
      field: ViewFieldContract,
      raw: string,
    ) => { value: string; error?: string };
    validate?: (
      request: Record<string, unknown>,
      errors: Record<string, string>,
    ) => void;
    minimum?: (request: Readonly<Record<string, unknown>>) => boolean;
  } = {},
): OrdinaryCreatePreparation {
  const errors: Record<string, string> = {};
  const request: Record<string, unknown> = { client_txn_id: clientTxnId };
  const actions = requireWorkbookSurfaceRegistration(contract.viewSchemaId)
    .policy.collectionActions;
  for (const [key, raw] of Object.entries(values)) {
    const field = contract.fieldMap[key];
    if (!field?.createWritable) {
      errors[key] = "This field is unavailable for ordinary creation.";
      continue;
    }
    if (raw === null) {
      if (field.clearable && field.writeKind !== "action_payload")
        request[key] = null;
      else errors[key] = `${field.label} cannot be cleared.`;
      continue;
    }
    if (field.writeKind === "action_payload") {
      const kind = actions[key] ?? "record";
      const parts = raw.split(/\r\n?|\n/u).filter((value) => value !== "");
      if (parts.length > 64) {
        errors[key] = "Choose at most 64 values.";
        continue;
      }
      const prepared = [];
      for (const part of parts) {
        if (kind === "record" || kind === "party") {
          if (!exactWorkbookReferenceId(part)) {
            errors[key] = "Choose a valid reference.";
            continue;
          }
          prepared.push(
            kind === "party"
              ? { op: "add_party_ref", party_id: part }
              : { op: "add_record_ref", linked_record_id: part },
          );
        } else {
          const text = normalizeWorkbookAuthoringText(
            part,
            kind === "alias"
              ? "alias_text_v1"
              : kind === "tag"
                ? "tag_label_v1"
                : "single_line_title_v1",
          );
          if (text.error) {
            errors[key] = text.error;
            continue;
          }
          if (text.value)
            prepared.push(
              kind === "alias"
                ? { op: "add_alias", alias_text: text.value }
                : kind === "tag"
                  ? { op: "add_tag", tag_name: text.value }
                  : { op: "add_risk_ref", risk_ref_text: text.value },
            );
        }
      }
      if (prepared.length)
        request[key] = { kind: "collection_actions_v1", actions: prepared };
      continue;
    }
    if (raw === "" && field.clearable) {
      request[key] = null;
      continue;
    }
    if (field.directReferenceContractId) {
      if (!exactWorkbookReferenceId(raw))
        errors[key] = `Choose ${field.label.toLowerCase()}.`;
      else request[key] = raw;
    } else if (field.directScalarContractId === "timestamp_instant_v1") {
      if (!validWorkbookTimestamp(raw))
        errors[key] = "Enter an RFC 3339 timestamp with a timezone.";
      else request[key] = raw;
    } else if (field.enumValues) {
      if (!field.enumValues.includes(raw))
        errors[key] = `Choose ${field.label.toLowerCase()}.`;
      else request[key] = raw;
    } else if (field.readKind === "boolean") {
      if (raw !== "true" && raw !== "false")
        errors[key] = "Choose true or false.";
      else request[key] = raw === "true";
    } else if (field.readKind === "number") {
      if (
        !/^-?\d+$/u.test(raw) ||
        !Number.isSafeInteger(Number(raw)) ||
        Number(raw) < 0 ||
        Number(raw) > 100
      )
        errors[key] = "Enter a whole number from 0 to 100.";
      else request[key] = Number(raw);
    } else {
      const normalized = field.stringContractId
        ? normalizeWorkbookAuthoringText(raw, field.stringContractId)
        : (options.scalar?.(field, raw) ?? { value: raw });
      if (normalized.error) errors[key] = normalized.error;
      else if (normalized.value === "") {
        if (field.clearable) request[key] = null;
        // Empty nonnullable optional values are omitted; minima are evaluated below.
      } else request[key] = normalized.value;
    }
  }
  const has = (key: string) => request[key] != null && request[key] !== "";
  const minimum = options.minimum
    ? options.minimum(request)
    : contract.minimumCreateFieldSets.some((set) => set.every(has)) ||
      contract.permitsZeroFieldCreate;
  if (!minimum) errors.minimum = genericCreateMinimumMessage(contract);
  if (!contract.createCapable)
    errors.contract = "Creation is unavailable on this sheet.";
  options.validate?.(request, errors);
  const decoded = Object.keys(errors).length
    ? null
    : decodeCreateViewRowRequest(contract, request);
  if (!decoded && !Object.keys(errors).length)
    errors.contract =
      "These values do not satisfy this sheet's creation contract.";
  return { request: decoded, errors };
}
