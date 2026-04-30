import { viewSchemaArtifacts } from "@cartulary/protocol-ts/generated";

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
  readonly field_key: string;
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
  readonly surface_kind: string;
  readonly technical_fields?: readonly string[];
  readonly title: string;
  readonly view_schema_id: string;
};

function truthMap(values: readonly string[]): Readonly<Record<string, true>> {
  return Object.freeze(
    Object.fromEntries(values.map((value) => [value, true])) as Record<
      string,
      true
    >,
  );
}

function parseContract(json: string): ViewContract {
  const raw = JSON.parse(json) as RawViewContract;
  const fields = Object.freeze(
    (raw.fields ?? []).map(
      (field): ViewFieldContract => ({
        fieldKey: field.field_key,
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
      }),
    ),
  );
  const fieldMap = Object.freeze(
    Object.fromEntries(
      fields.map((field) => [field.fieldKey, field]),
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
  const filterFields = Object.freeze([...(raw.filter_fields ?? [])]);
  const groupingFields = Object.freeze([...(raw.grouping_fields ?? [])]);
  const technicalFields = Object.freeze([...(raw.technical_fields ?? [])]);

  return Object.freeze({
    viewSchemaId: raw.view_schema_id,
    title: raw.title,
    surfaceKind: raw.surface_kind,
    defaultHiddenFields,
    defaultSort,
    defaultVisibleFields,
    sortFields,
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
  viewSchemaArtifacts
    .filter((artifact) => !artifact.path.endsWith("/index.json"))
    .map((artifact) => parseContract(artifact.json)),
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
  return contract.defaultVisibleFields
    .map((fieldKey) => contract.fieldMap[fieldKey])
    .filter((field): field is ViewFieldContract => field !== undefined);
}
