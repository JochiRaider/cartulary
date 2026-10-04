import type { ViewContract } from "@cartulary/view-contracts";
import {
  coreIndicatorCreateConstraints as constraints,
  coreIndicatorTypes,
} from "../../adapters/indicatorCreateProtocol";
import { workbookCreationAvailable } from "../../models/genericWorkbookModel";
import type { IndicatorObservation } from "../../mutations/workbookMutationCommandPorts";
import {
  freezeObservation,
  observationIndicatorView,
} from "./observationModel";

export {
  constraints as indicatorCreateConstraints,
  coreIndicatorTypes as indicatorCreateTypes,
};
export type IndicatorCreateValues = Readonly<Record<string, string>>;
export type IndicatorCreateDraft = Readonly<{
  observationId: string;
  revision: number;
  values: IndicatorCreateValues;
}>;
export type IndicatorCreateErrors = Readonly<Record<string, string>>;
const includes = (values: readonly string[], value: string) =>
  values.includes(value);
export function indicatorValueKinds(type: string): readonly string[] {
  return (
    (
      constraints.valueKindsByType as Readonly<
        Record<string, readonly string[]>
      >
    )[type] ?? []
  );
}
export function indicatorCreateAvailable(contract: ViewContract) {
  return (
    contract.viewSchemaId === observationIndicatorView &&
    workbookCreationAvailable(contract) &&
    constraints.requiredFields.every((key) =>
      contract.fields.some(
        (field) => field.fieldKey === key && field.createWritable,
      ),
    )
  );
}
export function indicatorCreateSeed(
  observation: IndicatorObservation,
): IndicatorCreateValues {
  const type = observation.parsed_indicator_type ?? "";
  return freezeObservation({
    "indicator.indicator_type": includes(coreIndicatorTypes, type) ? type : "",
    "indicator.value_kind": includes(constraints.atomicTypes, type)
      ? "atomic"
      : "",
    "indicator.display_value": observation.normalized_candidate ?? "",
  });
}
export function changeIndicatorCreateValue(
  values: IndicatorCreateValues,
  key: string,
  value: string,
): IndicatorCreateValues {
  if (key === "indicator.indicator_type" && value !== values[key]) {
    return freezeObservation({
      [key]: value,
      "indicator.value_kind": includes(constraints.atomicTypes, value)
        ? "atomic"
        : "",
    });
  }
  return freezeObservation({ ...values, [key]: value });
}
export function indicatorCreateErrors(
  contract: ViewContract,
  values: IndicatorCreateValues,
): IndicatorCreateErrors {
  const errors: Record<string, string> = {};
  const value = (key: string) => values[key] ?? "";
  if (!indicatorCreateAvailable(contract))
    errors.contract =
      "Canonical creation is unavailable for the active Indicator contract.";
  for (const key of constraints.requiredFields)
    if (!value(key)) errors[key] = "Required.";
  const type = value("indicator.indicator_type"),
    kind = value("indicator.value_kind");
  if (type && !includes(coreIndicatorTypes, type))
    errors["indicator.indicator_type"] = "Choose a supported Indicator type.";
  if (kind && !includes(constraints.valueKinds, kind))
    errors["indicator.value_kind"] = "Choose a supported value kind.";
  if (type && kind && !indicatorValueKinds(type).includes(kind))
    errors["indicator.value_kind"] =
      "Choose a value kind supported by this Indicator type.";
  const [algorithm, hash] = constraints.hashFields;
  if (Boolean(value(algorithm)) !== Boolean(value(hash))) {
    errors[value(algorithm) ? hash : algorithm] =
      "Supply both hash algorithm and hash value, or omit both.";
  }
  if (
    includes(constraints.hashForbiddenTypes, type) &&
    (value(algorithm) || value(hash))
  )
    errors[algorithm] = "IP Indicators do not admit hash fields.";
  if (
    value(hash) &&
    !new RegExp(constraints.hashValuePattern, "u").test(value(hash))
  )
    errors[hash] = "Use hexadecimal hash digits.";
  for (const key of ["indicator.display_value", "indicator.normalized_value"]) {
    if (Array.from(value(key)).length > constraints.maximumValueScalars)
      errors[key] =
        `Use at most ${constraints.maximumValueScalars} characters.`;
  }
  for (const [key, raw] of Object.entries(values)) {
    if (raw.includes("\0")) errors[key] = "Remove the null character.";
    if (
      raw &&
      !contract.fields.some(
        (field) => field.fieldKey === key && field.createWritable,
      )
    )
      errors[key] = "This field cannot be supplied on create.";
  }
  return errors;
}
export function indicatorCreateRequest(
  contract: ViewContract,
  values: IndicatorCreateValues,
  id: string,
) {
  if (Object.keys(indicatorCreateErrors(contract, values)).length) return null;
  const request: Record<string, unknown> & { client_txn_id: string } = {
    client_txn_id: id,
  };
  for (const field of contract.fields) {
    const raw = values[field.fieldKey];
    if (field.createWritable && raw !== undefined && raw !== "")
      request[field.fieldKey] = raw;
  }
  return request;
}
