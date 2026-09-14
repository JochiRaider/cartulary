import { viewSchemaRegistry } from "@cartulary/protocol-ts/view-schemas";
import type { ViewContract } from "@cartulary/view-contracts";

// Project protocol fixtures from typed contracts; object key order is deliberately
// reversed to exercise structural comparison rather than serialization identity.
export function publicWorkbookSchema(contract: ViewContract) {
  return {
    view_schema_id: contract.viewSchemaId,
    surface_kind: contract.surfaceKind,
    title: contract.title,
    source_record_types:
      viewSchemaRegistry.view_schemas.find(
        (view) => view.view_schema_id === contract.viewSchemaId,
      )?.source_record_types ?? [],
    technical_fields: contract.technicalFields,
    required_reference_pack_keys: contract.requiredReferencePackKeys,
    default_sort: snakeKeys(contract.defaultSort),
    sort_fields: contract.sortFields,
    sort_null_order: contract.sortNullOrder,
    filter_fields: contract.filterFields,
    synthetic_filter_predicates: [],
    grouping_fields: contract.groupingFields,
    create_capable: contract.createCapable,
    create_inputs: snakeKeys(contract.createInputs),
    inline_create: {
      minimum_create_field_sets: contract.minimumCreateFieldSets,
      permits_zero_field_create: contract.permitsZeroFieldCreate,
    },
    inspector_config: snakeKeys(contract.inspectorConfig),
    fields: contract.fields.map(
      ({
        writeAction: _writeAction,
        patchWritable: _patchWritable,
        ...field
      }) => snakeKeys(field),
    ),
  };
}
function snakeKeys(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(snakeKeys);
  if (value && typeof value === "object")
    return Object.fromEntries(
      Object.entries(value)
        .reverse()
        .map(([key, child]) => [
          key.replace(/[A-Z]/g, (letter) => `_${letter.toLowerCase()}`),
          snakeKeys(child),
        ]),
    );
  return value;
}
