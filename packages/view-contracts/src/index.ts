import { listViewSchemaArtifacts } from "@cartulary/protocol-ts";

export type SortEntry = {
  readonly fieldKey: string;
  readonly direction: "asc" | "desc";
};

export type ViewFieldContract = {
  readonly fieldKey: string;
  readonly label: string;
  readonly createWritable: boolean;
  readonly defaultHidden: boolean;
  readonly stringContractId: string | null;
  readonly directScalarContractId: string | null;
  readonly directReferenceContractId: string | null;
  readonly writeAction: string | null;
  readonly enumValues: readonly string[] | null;
  readonly headerSortFieldKey: string | null;
  readonly filterOps: readonly string[];
  readonly groupable: boolean;
  readonly sortable: boolean;
  readonly readKind: string;
  readonly writeKind: "read_only" | "direct_value" | "action_payload";
  readonly clearable: boolean;
  readonly conflictResolutionClass: string | null;
  readonly entityBindingMode: string | null;
};

export type ViewFieldCapability = {
  readonly editable: boolean;
  readonly filterable: boolean;
  readonly groupable: boolean;
  readonly sortable: boolean;
};

export type ViewContract = {
  readonly viewSchemaId: string;
  readonly title: string;
  readonly surfaceKind: string;
  readonly defaultHiddenFields: readonly string[];
  readonly defaultSort: readonly SortEntry[];
  readonly defaultVisibleFields: readonly string[];
  readonly filterFields: readonly string[];
  readonly fields: readonly ViewFieldContract[];
  readonly fieldMap: Readonly<Record<string, ViewFieldContract>>;
  readonly groupingFields: readonly string[];
  readonly permitsZeroFieldCreate: boolean;
  readonly sortableFieldMap: Readonly<Record<string, true>>;
  readonly filterableFieldMap: Readonly<Record<string, true>>;
  readonly groupableFieldMap: Readonly<Record<string, true>>;
  readonly sortFields: readonly string[];
  readonly sortNullOrder: "last";
  readonly technicalFields: readonly string[];
};

type RawField = {
  readonly clearable?: boolean;
  readonly conflict_resolution_class?: string | null;
  readonly create_writable?: boolean;
  readonly default_hidden?: boolean;
  readonly direct_reference_contract_id?: string | null;
  readonly direct_scalar_contract_id?: string | null;
  readonly entity_binding_mode?: string | null;
  readonly enum_values?: readonly string[] | null;
  readonly field_key?: string;
  readonly filter_ops?: readonly string[];
  readonly groupable?: boolean;
  readonly header_sort_field_key?: string | null;
  readonly label: string;
  readonly read_kind?: string;
  readonly sortable?: boolean;
  readonly string_contract_id?: string | null;
  readonly write_action?: string | null;
  readonly write_kind?: "read_only" | "direct_value" | "action_payload";
};

type RawSyntheticFilterPredicate = {
  readonly field_key?: string;
  readonly filter_ops?: readonly string[];
  readonly label: string;
};

type RawViewContract = {
  readonly default_hidden_fields?: readonly string[];
  readonly default_sort?: ReadonlyArray<{
    readonly field_key: string;
    readonly direction: "asc" | "desc";
  }>;
  readonly default_visible_fields?: readonly string[];
  readonly fields?: readonly RawField[];
  readonly filter_fields?: readonly string[];
  readonly grouping_fields?: readonly string[];
  readonly inline_create?: {
    readonly permits_zero_field_create?: boolean;
  };
  readonly sort_fields?: readonly string[];
  readonly sort_null_order?: "last";
  readonly surface_kind: string;
  readonly synthetic_filter_predicates?: readonly RawSyntheticFilterPredicate[];
  readonly technical_fields?: readonly string[];
  readonly title: string;
  readonly view_schema_id: string;
};

// FE-U-P0-02 exposes this narrow parser error contract for malformed
// view-schema adapter inputs that could otherwise drift away from field_key.
function viewContractInvariant(source: string, detail: string): never {
  throw new Error(`View contract invariant failed: ${source} ${detail}`);
}

function requireStableKey(
  value: unknown,
  source: string,
  label: string,
): string {
  if (typeof value !== "string" || value.trim() === "") {
    viewContractInvariant(source, `${label} must be a non-empty string`);
  }
  return value;
}

function stableKeySet(
  values: readonly string[],
  source: string,
  label: string,
): ReadonlySet<string> {
  const keys = new Set<string>();
  for (const [index, value] of values.entries()) {
    const fieldKey = requireStableKey(value, source, `${label}[${index + 1}]`);
    if (keys.has(fieldKey)) {
      viewContractInvariant(source, `${label} duplicate field_key ${fieldKey}`);
    }
    keys.add(fieldKey);
  }
  return keys;
}

function unionKeySet(
  ...sets: readonly ReadonlySet<string>[]
): ReadonlySet<string> {
  const keys = new Set<string>();
  for (const set of sets) {
    for (const key of set) {
      keys.add(key);
    }
  }
  return keys;
}

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

function truthMap(values: readonly string[]): Readonly<Record<string, true>> {
  return Object.freeze(
    Object.fromEntries(values.map((value) => [value, true])) as Record<
      string,
      true
    >,
  );
}

export function parseViewContractJSON(
  json: string,
  source = "view contract",
): ViewContract {
  let raw: RawViewContract;
  try {
    raw = JSON.parse(json) as RawViewContract;
  } catch (error) {
    viewContractInvariant(
      source,
      `contains invalid JSON: ${error instanceof Error ? error.message : String(error)}`,
    );
  }
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
    permitsZeroFieldCreate:
      raw.inline_create?.permits_zero_field_create ?? false,
    fields,
    fieldMap,
    sortableFieldMap: truthMap(sortFields),
    filterableFieldMap: truthMap(filterFields),
    groupableFieldMap: truthMap(groupingFields),
  });
}

const contracts = Object.freeze(
  listViewSchemaArtifacts()
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
    editable: field?.writeKind !== undefined && field.writeKind !== "read_only",
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
