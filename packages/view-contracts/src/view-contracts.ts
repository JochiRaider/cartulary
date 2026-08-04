import type { ViewSchemaSourceDocument } from "@cartulary/protocol-ts/view-schemas";
import {
  viewSchemaArtifacts,
  viewSchemaSourceDocumentDecoder,
} from "@cartulary/protocol-ts/view-schemas";
import { parseInspectorConfig } from "./inspector.js";
import {
  requireContractBoolean,
  requireContractObject,
  requireStableKey,
  stableKeyList,
  stableKeyMatrix,
  stableKeySet,
  truthMap,
  unionKeySet,
  viewContractInvariant,
  viewContractSourceInvariant,
} from "./invariants.js";
import type {
  SortEntry,
  ViewContract,
  ViewFieldCapability,
  ViewFieldContract,
} from "./types.js";

type RawViewContract = ViewSchemaSourceDocument;

function validateFieldKeyReferences(
  values: readonly string[],
  allowedKeys: ReadonlySet<string>,
  source: string,
  label: string,
) {
  for (const [index, value] of values.entries()) {
    const fieldKey = requireStableKey(value, source, `${label}[${index + 1}]`);
    if (!allowedKeys.has(fieldKey)) {
      viewContractInvariant(
        source,
        `${label} references unknown field_key ${fieldKey}`,
      );
    }
  }
}

function validateDefaultSortReferences(
  values: readonly SortEntry[],
  allowedKeys: ReadonlySet<string>,
  source: string,
) {
  for (const [index, entry] of values.entries()) {
    const fieldKey = requireStableKey(
      entry.fieldKey,
      source,
      `default_sort[${index + 1}].field_key`,
    );
    if (!allowedKeys.has(fieldKey)) {
      viewContractInvariant(
        source,
        `default_sort[${index + 1}].field_key references unknown field_key ${fieldKey}`,
      );
    }
  }
}

function validateHeaderSortReferences(
  fields: readonly ViewFieldContract[],
  knownFieldKeys: ReadonlySet<string>,
  sortFieldKeys: ReadonlySet<string>,
  source: string,
) {
  for (const [index, field] of fields.entries()) {
    const fieldKey = field.headerSortFieldKey;
    if (fieldKey === null) {
      continue;
    }
    const label = `fields[${index + 1}].header_sort_field_key`;
    if (!knownFieldKeys.has(fieldKey)) {
      viewContractInvariant(
        source,
        `${label} references unknown field_key ${fieldKey}`,
      );
    }
    if (!sortFieldKeys.has(fieldKey)) {
      viewContractInvariant(
        source,
        `${label} references non-sortable field_key ${fieldKey}`,
      );
    }
  }
}

function validateMinimumCreateFieldSets(
  values: readonly (readonly string[])[],
  allowedKeys: ReadonlySet<string>,
  source: string,
) {
  for (const [index, fieldSet] of values.entries()) {
    if (fieldSet.length === 0) {
      viewContractInvariant(
        source,
        `inline_create.minimum_create_field_sets[${index + 1}] must not be empty`,
      );
    }
    validateFieldKeyReferences(
      fieldSet,
      allowedKeys,
      source,
      `inline_create.minimum_create_field_sets[${index + 1}]`,
    );
  }
}

function parseInlineCreate(value: unknown, source: string) {
  const inlineCreate = requireContractObject(value, source, "inline_create");
  const allowedKeys = new Set([
    "minimum_create_field_sets",
    "permits_zero_field_create",
  ]);
  for (const key of Object.keys(inlineCreate)) {
    if (!allowedKeys.has(key)) {
      viewContractInvariant(source, `inline_create has unknown member ${key}`);
    }
  }
  return {
    minimumCreateFieldSets: stableKeyMatrix(
      inlineCreate.minimum_create_field_sets,
      source,
      "inline_create.minimum_create_field_sets",
    ),
    permitsZeroFieldCreate: requireContractBoolean(
      inlineCreate.permits_zero_field_create,
      source,
      "inline_create.permits_zero_field_create",
    ),
  };
}

export function parseViewContractJSON(
  json: string,
  source = "view contract",
): ViewContract {
  let value: unknown;
  try {
    value = JSON.parse(json) as unknown;
  } catch {
    viewContractSourceInvariant(source, "$", "invalid_json");
  }
  const decoded = viewSchemaSourceDocumentDecoder.decode(value);
  if (!decoded.ok) {
    viewContractSourceInvariant(
      source,
      decoded.error.instancePath === ""
        ? "$"
        : `$${decoded.error.instancePath}`,
      decoded.error.reasonCategory,
    );
  }
  const raw: RawViewContract = decoded.value;
  requireStableKey(raw.view_schema_id, source, "view_schema_id");
  const fields = Object.freeze(
    (raw.fields ?? []).map((field, index): ViewFieldContract => {
      const fieldKey = requireStableKey(
        field.field_key,
        source,
        `fields[${index + 1}].field_key`,
      );
      return {
        fieldKey,
        label: field.label,
        createWritable: field.create_writable ?? false,
        defaultHidden: field.default_hidden ?? false,
        stringContractId: field.string_contract_id ?? null,
        directScalarContractId: field.direct_scalar_contract_id ?? null,
        directReferenceContractId: field.direct_reference_contract_id ?? null,
        writeAction: field.write_action ?? null,
        enumValues:
          field.enum_values === null || field.enum_values === undefined
            ? null
            : Object.freeze([...field.enum_values]),
        headerSortFieldKey: field.header_sort_field_key ?? null,
        filterOps: Object.freeze([...(field.filter_ops ?? [])]),
        groupable: field.groupable ?? false,
        sortable: field.sortable ?? false,
        readKind: field.read_kind ?? "text",
        writeKind: field.write_kind ?? "read_only",
        gridEditable: requireContractBoolean(
          field.grid_editable,
          source,
          `fields[${index + 1}].grid_editable`,
        ),
        clearable: field.clearable ?? false,
        conflictResolutionClass: field.conflict_resolution_class ?? null,
        entityBindingMode: field.entity_binding_mode ?? null,
      };
    }),
  );
  const syntheticFilterFields = Object.freeze(
    (raw.synthetic_filter_predicates ?? []).map(
      (field, index): ViewFieldContract => {
        const fieldKey = requireStableKey(
          field.field_key,
          source,
          `synthetic_filter_predicates[${index + 1}].field_key`,
        );
        return {
          fieldKey,
          label: field.label,
          createWritable: false,
          defaultHidden: true,
          stringContractId: null,
          directScalarContractId: null,
          directReferenceContractId: null,
          writeAction: null,
          enumValues: null,
          headerSortFieldKey: null,
          filterOps: Object.freeze([...(field.filter_ops ?? [])]),
          groupable: false,
          sortable: false,
          readKind: "synthetic_filter",
          writeKind: "read_only",
          gridEditable: false,
          clearable: false,
          conflictResolutionClass: null,
          entityBindingMode: null,
        };
      },
    ),
  );
  const fieldKeySet = stableKeySet(
    fields.map((field) => field.fieldKey),
    source,
    "fields",
  );
  const syntheticFieldKeySet = stableKeySet(
    syntheticFilterFields.map((field) => field.fieldKey),
    source,
    "synthetic_filter_predicates",
  );
  const duplicateSyntheticField = syntheticFilterFields.find((field) =>
    fieldKeySet.has(field.fieldKey),
  );
  if (duplicateSyntheticField !== undefined) {
    viewContractInvariant(
      source,
      `synthetic_filter_predicates duplicate field_key ${duplicateSyntheticField.fieldKey}`,
    );
  }
  const fieldMapKeySet = unionKeySet(fieldKeySet, syntheticFieldKeySet);
  const fieldMap = Object.freeze(
    Object.fromEntries(
      [...fields, ...syntheticFilterFields].map((field) => [
        field.fieldKey,
        field,
      ]),
    ) as Record<string, ViewFieldContract>,
  );
  const defaultSort = Object.freeze(
    (raw.default_sort ?? []).map(
      (entry): SortEntry => ({
        fieldKey: entry.field_key,
        direction: entry.direction,
      }),
    ),
  );
  const defaultVisibleFields = Object.freeze([
    ...(raw.default_visible_fields ?? []),
  ]);
  const defaultHiddenFields = Object.freeze([
    ...(raw.default_hidden_fields ?? []),
  ]);
  const sortFields = Object.freeze([...(raw.sort_fields ?? [])]);
  const sortNullOrder = raw.sort_null_order ?? "last";
  const filterFields = Object.freeze([
    ...(raw.filter_fields ?? []),
    ...syntheticFilterFields.map((field) => field.fieldKey),
  ]);
  const groupingFields = Object.freeze([...(raw.grouping_fields ?? [])]);
  const technicalFields = Object.freeze([...(raw.technical_fields ?? [])]);
  const inlineCreate = parseInlineCreate(raw.inline_create, source);
  const minimumCreateFieldSets = inlineCreate.minimumCreateFieldSets;
  const requiredReferencePackKeys = stableKeyList(
    raw.required_reference_pack_keys,
    source,
    "required_reference_pack_keys",
  );
  const inspectorConfig = parseInspectorConfig(
    raw.inspector_config,
    raw.view_schema_id,
    source,
  );
  const technicalFieldKeySet = stableKeySet(
    technicalFields,
    source,
    "technical_fields",
  );
  const fieldOrTechnicalKeySet = unionKeySet(
    fieldMapKeySet,
    technicalFieldKeySet,
  );
  validateFieldKeyReferences(
    defaultVisibleFields,
    fieldMapKeySet,
    source,
    "default_visible_fields",
  );
  validateFieldKeyReferences(
    defaultHiddenFields,
    fieldOrTechnicalKeySet,
    source,
    "default_hidden_fields",
  );
  validateFieldKeyReferences(sortFields, fieldMapKeySet, source, "sort_fields");
  validateFieldKeyReferences(
    raw.filter_fields ?? [],
    fieldMapKeySet,
    source,
    "filter_fields",
  );
  validateFieldKeyReferences(
    groupingFields,
    fieldMapKeySet,
    source,
    "grouping_fields",
  );
  validateDefaultSortReferences(defaultSort, fieldOrTechnicalKeySet, source);
  validateHeaderSortReferences(
    fields,
    fieldMapKeySet,
    new Set(sortFields),
    source,
  );
  validateMinimumCreateFieldSets(
    minimumCreateFieldSets,
    fieldMapKeySet,
    source,
  );

  return Object.freeze({
    viewSchemaId: raw.view_schema_id,
    title: raw.title,
    surfaceKind: raw.surface_kind,
    defaultHiddenFields,
    defaultSort,
    defaultVisibleFields,
    sortFields,
    sortNullOrder,
    filterFields,
    groupingFields,
    technicalFields,
    inspectorConfig,
    minimumCreateFieldSets,
    permitsZeroFieldCreate: inlineCreate.permitsZeroFieldCreate,
    requiredReferencePackKeys,
    fields,
    fieldMap,
    sortableFieldMap: truthMap(sortFields),
    filterableFieldMap: truthMap(filterFields),
    groupableFieldMap: truthMap(groupingFields),
  });
}

const contracts = Object.freeze(
  viewSchemaArtifacts
    .filter((artifact) => !artifact.path.endsWith("/index.json"))
    .map((artifact) => parseViewContractJSON(artifact.json, artifact.path)),
);

const contractIndex = Object.freeze(
  Object.fromEntries(
    contracts.map((contract) => [contract.viewSchemaId, contract]),
  ) as Record<string, ViewContract>,
);

export function listViewContracts(): readonly ViewContract[] {
  return contracts;
}

export function getViewContract(
  viewSchemaId: string,
): ViewContract | undefined {
  return contractIndex[viewSchemaId];
}

export function requireViewContract(viewSchemaId: string): ViewContract {
  const contract = getViewContract(viewSchemaId);
  if (!contract) {
    throw new Error(`Unknown view schema contract: ${viewSchemaId}`);
  }
  return contract;
}

export function resolveHeaderSortFieldKey(
  contract: ViewContract,
  fieldKey: string,
): string | null {
  const field = contract.fieldMap[fieldKey];
  if (!field) {
    return null;
  }
  return field.headerSortFieldKey ?? field.fieldKey;
}

export function fieldCapability(
  contract: ViewContract,
  fieldKey: string,
): ViewFieldCapability {
  const field = contract.fieldMap[fieldKey];
  return {
    editable: field?.gridEditable ?? false,
    filterable: (field?.filterOps.length ?? 0) > 0,
    groupable: field?.groupable ?? false,
    sortable: field?.sortable ?? false,
  };
}

export function visibleFields(
  contract: ViewContract,
): readonly ViewFieldContract[] {
  return contract.defaultVisibleFields.map((fieldKey) => {
    const field = contract.fieldMap[fieldKey];
    if (field === undefined) {
      viewContractInvariant(
        contract.viewSchemaId,
        `default_visible_fields references unknown field_key ${fieldKey}`,
      );
    }
    return field;
  });
}
