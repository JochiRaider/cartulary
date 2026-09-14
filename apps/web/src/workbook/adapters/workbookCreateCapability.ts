import type { HTTPOperationResponse } from "@cartulary/protocol-ts/http";
import type { ViewContract } from "@cartulary/view-contracts";

/** Shared creation discovery facts; source context and feature routes remain separate. */
export function workbookCreateCapabilityMatches(
  schema: HTTPOperationResponse<"getViewSchema">["data"],
  target: ViewContract,
): boolean {
  return (
    schema.create_capable &&
    schema.view_schema_id === target.viewSchemaId &&
    schema.inline_create.permits_zero_field_create ===
      target.permitsZeroFieldCreate &&
    JSON.stringify(schema.inline_create.minimum_create_field_sets) ===
      JSON.stringify(target.minimumCreateFieldSets) &&
    JSON.stringify(
      schema.create_inputs.map((input) => [
        input.input_key,
        input.value_contract_id,
        input.required,
        input.nullable,
      ]),
    ) ===
      JSON.stringify(
        target.createInputs.map((input) => [
          input.inputKey,
          input.valueContractId,
          input.required,
          input.nullable,
        ]),
      ) &&
    schema.fields.length === target.fields.length &&
    new Set(schema.fields.map((field) => field.field_key)).size ===
      schema.fields.length &&
    schema.fields.every((field) => {
      const local = target.fieldMap[field.field_key];
      return (
        local &&
        field.create_writable === local.createWritable &&
        field.clearable === local.clearable &&
        field.write_kind === local.writeKind &&
        field.read_kind === local.readKind &&
        field.string_contract_id === local.stringContractId &&
        field.direct_scalar_contract_id === local.directScalarContractId &&
        field.direct_reference_contract_id ===
          local.directReferenceContractId &&
        JSON.stringify(field.enum_values) === JSON.stringify(local.enumValues)
      );
    })
  );
}
