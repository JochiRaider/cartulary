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

export type InspectorPanelId =
  | "details"
  | "relationships"
  | "evidence"
  | "history"
  | "workflow";

export type InspectorRouteBindingKind =
  | "panel_read"
  | "view_row_create"
  | "record_patch"
  | "record_action"
  | "entity_mention_action"
  | "evidence_access"
  | "surface_pivot";

export type InspectorDisabledCondition =
  | "no_row_selected"
  | "incident_closed"
  | "authorization_lost"
  | "row_version_changed"
  | "record_deleted"
  | "record_merged"
  | "evidence_preview_unavailable"
  | "merge_target_unavailable";

export type InspectorSeedSourceKind =
  | "selected_record_id"
  | "selected_field_value"
  | "literal";

export type InspectorPanel = {
  readonly panelId: InspectorPanelId;
  readonly label: string;
};

export type InspectorRouteBinding = {
  readonly kind: InspectorRouteBindingKind;
  readonly targetViewSchemaId?: string | undefined;
  readonly actionKey?: string | undefined;
};

export type InspectorSeedSource = {
  readonly kind: InspectorSeedSourceKind;
  readonly sourceFieldKey?: string | undefined;
  readonly value?: unknown;
};

export type InspectorSeedBinding = {
  readonly targetFieldKey: string;
  readonly source: InspectorSeedSource;
};

export type InspectorFeatureGroup = {
  readonly featureGroupKey: string;
  readonly panelId: InspectorPanelId;
  readonly label: string;
  readonly minimumIncidentRole:
    | "viewer"
    | "editor"
    | "reviewer"
    | "admin"
    | null;
  readonly mutates: boolean;
  readonly requiresConfirmation: boolean;
  readonly routeBinding: InspectorRouteBinding;
  readonly seedBindings: readonly InspectorSeedBinding[];
  readonly disabledWhen: readonly InspectorDisabledCondition[];
};

export type InspectorConfig = {
  readonly inspectorConfigSchemaId: "cartulary.inspector_config.v1";
  readonly viewSchemaId: string;
  readonly defaultOpen: false;
  readonly subjectBinding: {
    readonly kind: "selected_record";
  };
  readonly noRowState: "no_row_selected";
  readonly unsupportedFeatureBehavior: "omit_feature";
  readonly panels: readonly InspectorPanel[];
  readonly featureGroups: readonly InspectorFeatureGroup[];
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
  readonly inspectorConfig: InspectorConfig;
  readonly permitsZeroFieldCreate: boolean;
  readonly requiredReferencePackKeys: readonly string[];
  readonly sortableFieldMap: Readonly<Record<string, true>>;
  readonly filterableFieldMap: Readonly<Record<string, true>>;
  readonly groupableFieldMap: Readonly<Record<string, true>>;
  readonly sortFields: readonly string[];
  readonly sortNullOrder: "last";
  readonly technicalFields: readonly string[];
};

export type ViewRowCellV1 = {
  readonly value: unknown;
};

export type ViewRowV1 = {
  readonly record_id: string;
  readonly row_version: number;
  readonly cells: Readonly<Record<string, ViewRowCellV1>>;
  readonly group_values?: Readonly<Record<string, unknown>> | undefined;
  readonly view_schema_id?: string | undefined;
};

export type NormalizedViewRowV1 = {
  readonly recordId: string;
  readonly rowVersion: number;
  readonly viewSchemaId: string;
  readonly cells: Readonly<Record<string, ViewRowCellV1>>;
  readonly groupValues?: Readonly<Record<string, unknown>> | undefined;
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

type RawInspectorConfig = {
  readonly default_open?: boolean;
  readonly feature_groups?: readonly RawInspectorFeatureGroup[];
  readonly inspector_config_schema_id?: string;
  readonly no_row_state?: string;
  readonly panels?: readonly RawInspectorPanel[];
  readonly subject_binding?: {
    readonly kind?: string;
  };
  readonly unsupported_feature_behavior?: string;
  readonly view_schema_id?: string;
};

type RawInspectorPanel = {
  readonly label?: string;
  readonly panel_id?: string;
};

type RawInspectorFeatureGroup = {
  readonly disabled_when?: readonly string[];
  readonly feature_group_key?: string;
  readonly label?: string;
  readonly minimum_incident_role?: string | null;
  readonly mutates?: boolean;
  readonly panel_id?: string;
  readonly requires_confirmation?: boolean;
  readonly route_binding?: RawInspectorRouteBinding;
  readonly seed_bindings?: readonly RawInspectorSeedBinding[];
};

type RawInspectorRouteBinding = {
  readonly action_key?: string;
  readonly kind?: string;
  readonly target_view_schema_id?: string;
};

type RawInspectorSeedBinding = {
  readonly source?: RawInspectorSeedSource;
  readonly target_field_key?: string;
};

type RawInspectorSeedSource = {
  readonly kind?: string;
  readonly source_field_key?: string;
  readonly value?: unknown;
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
  readonly inspector_config?: RawInspectorConfig;
  readonly required_reference_pack_keys?: unknown;
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

function viewRowInvariant(source: string, detail: string): never {
  throw new Error(`View row invariant failed: ${source} ${detail}`);
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

function stableKeyList(
  value: unknown,
  source: string,
  label: string,
): readonly string[] {
  if (value === undefined) {
    return Object.freeze([]);
  }
  if (!Array.isArray(value)) {
    viewContractInvariant(source, `${label} must be an array`);
  }
  const keys = stableKeySet(value, source, label);
  return Object.freeze([...keys]);
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

const inspectorPanelIds = Object.freeze([
  "details",
  "relationships",
  "evidence",
  "history",
  "workflow",
] as const);

const inspectorRouteBindingKinds = Object.freeze([
  "panel_read",
  "view_row_create",
  "record_patch",
  "record_action",
  "entity_mention_action",
  "evidence_access",
  "surface_pivot",
] as const);

const inspectorDisabledConditions = Object.freeze([
  "no_row_selected",
  "incident_closed",
  "authorization_lost",
  "row_version_changed",
  "record_deleted",
  "record_merged",
  "evidence_preview_unavailable",
  "merge_target_unavailable",
] as const);

const inspectorSeedSourceKinds = Object.freeze([
  "selected_record_id",
  "selected_field_value",
  "literal",
] as const);

const incidentRoles = Object.freeze([
  "viewer",
  "editor",
  "reviewer",
  "admin",
] as const);

function parseInspectorConfig(
  raw: unknown,
  viewSchemaId: string,
  source: string,
): InspectorConfig {
  const config = requireContractObject(
    raw,
    source,
    "inspector_config",
  ) as RawInspectorConfig;
  const inspectorConfigSchemaId = requireEnumValue(
    config.inspector_config_schema_id,
    ["cartulary.inspector_config.v1"] as const,
    source,
    "inspector_config.inspector_config_schema_id",
  );
  const configViewSchemaId = requireStableKey(
    config.view_schema_id,
    source,
    "inspector_config.view_schema_id",
  );
  if (configViewSchemaId !== viewSchemaId) {
    viewContractInvariant(
      source,
      `inspector_config.view_schema_id must match ${viewSchemaId}`,
    );
  }
  const defaultOpen = requireContractBoolean(
    config.default_open,
    source,
    "inspector_config.default_open",
  );
  if (defaultOpen) {
    viewContractInvariant(
      source,
      "inspector_config.default_open must be false",
    );
  }
  const subjectBinding = requireContractObject(
    config.subject_binding,
    source,
    "inspector_config.subject_binding",
  );
  const subjectKind = requireEnumValue(
    subjectBinding.kind,
    ["selected_record"] as const,
    source,
    "inspector_config.subject_binding.kind",
  );
  const noRowState = requireEnumValue(
    config.no_row_state,
    ["no_row_selected"] as const,
    source,
    "inspector_config.no_row_state",
  );
  const unsupportedFeatureBehavior = requireEnumValue(
    config.unsupported_feature_behavior,
    ["omit_feature"] as const,
    source,
    "inspector_config.unsupported_feature_behavior",
  );

  if (!Array.isArray(config.panels) || config.panels.length === 0) {
    viewContractInvariant(
      source,
      "inspector_config.panels must be a non-empty array",
    );
  }
  if (config.panels.length > 5) {
    viewContractInvariant(
      source,
      "inspector_config.panels must contain at most 5 entries",
    );
  }
  const panelIds = new Set<InspectorPanelId>();
  const panels = Object.freeze(
    config.panels.map((rawPanel, index): InspectorPanel => {
      const panel = requireContractObject(
        rawPanel,
        source,
        `inspector_config.panels[${index + 1}]`,
      ) as RawInspectorPanel;
      const panelId = requireEnumValue(
        panel.panel_id,
        inspectorPanelIds,
        source,
        `inspector_config.panels[${index + 1}].panel_id`,
      );
      if (panelIds.has(panelId)) {
        viewContractInvariant(
          source,
          `inspector_config.panels duplicate panel_id ${panelId}`,
        );
      }
      panelIds.add(panelId);
      return Object.freeze({
        panelId,
        label: requireStableKey(
          panel.label,
          source,
          `inspector_config.panels[${index + 1}].label`,
        ),
      });
    }),
  );

  if (!Array.isArray(config.feature_groups)) {
    viewContractInvariant(
      source,
      "inspector_config.feature_groups must be an array",
    );
  }
  if (config.feature_groups.length > 64) {
    viewContractInvariant(
      source,
      "inspector_config.feature_groups must contain at most 64 entries",
    );
  }
  const featureKeys = new Set<string>();
  const featureGroups = Object.freeze(
    config.feature_groups.map((rawGroup, index): InspectorFeatureGroup => {
      const group = requireContractObject(
        rawGroup,
        source,
        `inspector_config.feature_groups[${index + 1}]`,
      ) as RawInspectorFeatureGroup;
      const label = `inspector_config.feature_groups[${index + 1}]`;
      const featureGroupKey = requireInspectorKey(
        group.feature_group_key,
        source,
        `${label}.feature_group_key`,
      );
      if (featureKeys.has(featureGroupKey)) {
        viewContractInvariant(
          source,
          `inspector_config.feature_groups duplicate feature_group_key ${featureGroupKey}`,
        );
      }
      featureKeys.add(featureGroupKey);
      const panelId = requireEnumValue(
        group.panel_id,
        inspectorPanelIds,
        source,
        `${label}.panel_id`,
      );
      if (!panelIds.has(panelId)) {
        viewContractInvariant(
          source,
          `${label}.panel_id references unknown panel_id ${panelId}`,
        );
      }
      const minimumIncidentRole =
        group.minimum_incident_role === null
          ? null
          : requireEnumValue(
              group.minimum_incident_role,
              incidentRoles,
              source,
              `${label}.minimum_incident_role`,
            );
      const routeBinding = parseInspectorRouteBinding(
        group.route_binding,
        source,
        `${label}.route_binding`,
      );
      const seedBindings = parseInspectorSeedBindings(
        group.seed_bindings,
        source,
        `${label}.seed_bindings`,
      );
      const disabledWhen = parseInspectorDisabledConditions(
        group.disabled_when,
        source,
        `${label}.disabled_when`,
      );
      return Object.freeze({
        featureGroupKey,
        panelId,
        label: requireStableKey(group.label, source, `${label}.label`),
        minimumIncidentRole,
        mutates: requireContractBoolean(
          group.mutates,
          source,
          `${label}.mutates`,
        ),
        requiresConfirmation: requireContractBoolean(
          group.requires_confirmation,
          source,
          `${label}.requires_confirmation`,
        ),
        routeBinding,
        seedBindings,
        disabledWhen,
      });
    }),
  );

  return Object.freeze({
    inspectorConfigSchemaId,
    viewSchemaId: configViewSchemaId,
    defaultOpen: false,
    subjectBinding: Object.freeze({ kind: subjectKind }),
    noRowState,
    unsupportedFeatureBehavior,
    panels,
    featureGroups,
  });
}

function parseInspectorRouteBinding(
  raw: unknown,
  source: string,
  label: string,
): InspectorRouteBinding {
  const route = requireContractObject(
    raw,
    source,
    label,
  ) as RawInspectorRouteBinding;
  const binding: InspectorRouteBinding = {
    kind: requireEnumValue(
      route.kind,
      inspectorRouteBindingKinds,
      source,
      `${label}.kind`,
    ),
  };
  if (route.target_view_schema_id !== undefined) {
    const action =
      route.action_key === undefined
        ? {}
        : {
            actionKey: requireInspectorKey(
              route.action_key,
              source,
              `${label}.action_key`,
            ),
          };
    return Object.freeze({
      ...binding,
      targetViewSchemaId: requireStableKey(
        route.target_view_schema_id,
        source,
        `${label}.target_view_schema_id`,
      ),
      ...action,
    });
  }
  if (route.action_key !== undefined) {
    return Object.freeze({
      ...binding,
      actionKey: requireInspectorKey(
        route.action_key,
        source,
        `${label}.action_key`,
      ),
    });
  }
  return Object.freeze(binding);
}

function parseInspectorSeedBindings(
  raw: unknown,
  source: string,
  label: string,
): readonly InspectorSeedBinding[] {
  if (!Array.isArray(raw)) {
    viewContractInvariant(source, `${label} must be an array`);
  }
  if (raw.length > 16) {
    viewContractInvariant(source, `${label} must contain at most 16 entries`);
  }
  return Object.freeze(
    raw.map((rawBinding, index): InspectorSeedBinding => {
      const binding = requireContractObject(
        rawBinding,
        source,
        `${label}[${index + 1}]`,
      ) as RawInspectorSeedBinding;
      return Object.freeze({
        targetFieldKey: requireStableKey(
          binding.target_field_key,
          source,
          `${label}[${index + 1}].target_field_key`,
        ),
        source: parseInspectorSeedSource(
          binding.source,
          source,
          `${label}[${index + 1}].source`,
        ),
      });
    }),
  );
}

function parseInspectorSeedSource(
  raw: unknown,
  source: string,
  label: string,
): InspectorSeedSource {
  const sourceObject = requireContractObject(raw, source, label);
  const kind = requireEnumValue(
    sourceObject.kind,
    inspectorSeedSourceKinds,
    source,
    `${label}.kind`,
  );
  const sourceFieldKey =
    sourceObject.source_field_key === undefined
      ? undefined
      : requireStableKey(
          sourceObject.source_field_key,
          source,
          `${label}.source_field_key`,
        );
  if (kind === "selected_field_value" && sourceFieldKey === undefined) {
    viewContractInvariant(
      source,
      `${label}.source_field_key is required for selected_field_value`,
    );
  }
  const base: InspectorSeedSource =
    sourceFieldKey === undefined ? { kind } : { kind, sourceFieldKey };
  if (kind === "literal") {
    if (!hasOwn(sourceObject, "value")) {
      viewContractInvariant(source, `${label}.value is required for literal`);
    }
    return Object.freeze({ ...base, value: sourceObject.value });
  }
  if (hasOwn(sourceObject, "value")) {
    return Object.freeze({ ...base, value: sourceObject.value });
  }
  return Object.freeze(base);
}

function parseInspectorDisabledConditions(
  raw: unknown,
  source: string,
  label: string,
): readonly InspectorDisabledCondition[] {
  if (!Array.isArray(raw)) {
    viewContractInvariant(source, `${label} must be an array`);
  }
  if (raw.length > 16) {
    viewContractInvariant(source, `${label} must contain at most 16 entries`);
  }
  const conditions = new Set<InspectorDisabledCondition>();
  return Object.freeze(
    raw.map((value, index) => {
      const condition = requireEnumValue(
        value,
        inspectorDisabledConditions,
        source,
        `${label}[${index + 1}]`,
      );
      if (conditions.has(condition)) {
        viewContractInvariant(
          source,
          `${label} duplicate condition ${condition}`,
        );
      }
      conditions.add(condition);
      return condition;
    }),
  );
}

function truthMap(values: readonly string[]): Readonly<Record<string, true>> {
  return Object.freeze(
    Object.fromEntries(values.map((value) => [value, true])) as Record<
      string,
      true
    >,
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function hasOwn(object: Record<string, unknown>, key: string): boolean {
  return Object.hasOwn(object, key);
}

function requireContractObject(
  value: unknown,
  source: string,
  label: string,
): Record<string, unknown> {
  if (!isRecord(value)) {
    viewContractInvariant(source, `${label} must be an object`);
  }
  return value;
}

function requireContractBoolean(
  value: unknown,
  source: string,
  label: string,
): boolean {
  if (typeof value !== "boolean") {
    viewContractInvariant(source, `${label} must be a boolean`);
  }
  return value;
}

function requireEnumValue<T extends string>(
  value: unknown,
  allowed: readonly T[],
  source: string,
  label: string,
): T {
  if (typeof value !== "string" || !allowed.includes(value as T)) {
    viewContractInvariant(
      source,
      `${label} must be one of ${allowed.join("|")}`,
    );
  }
  return value as T;
}

function requireInspectorKey(value: unknown, source: string, label: string) {
  const key = requireStableKey(value, source, label);
  if (!/^[a-z0-9_.]+$/.test(key)) {
    viewContractInvariant(
      source,
      `${label} must be ASCII lower snake or dotted key`,
    );
  }
  return key;
}

function requireRowObject(value: unknown, source: string, label: string) {
  if (!isRecord(value)) {
    viewRowInvariant(source, `${label} must be an object`);
  }
  return value;
}

function requireRowRecordId(value: unknown, source: string): string {
  if (typeof value !== "string" || value.trim() === "") {
    viewRowInvariant(source, "record_id must be a non-empty string");
  }
  return value;
}

function requireRowVersion(value: unknown, source: string): number {
  if (typeof value !== "number" || !Number.isSafeInteger(value) || value < 0) {
    viewRowInvariant(source, "row_version must be a non-negative safe integer");
  }
  return value;
}

function declaredDataFieldKeys(contract: ViewContract): ReadonlySet<string> {
  return new Set(contract.fields.map((field) => field.fieldKey));
}

function technicalFieldKeys(contract: ViewContract): ReadonlySet<string> {
  return new Set(contract.technicalFields);
}

function sanitizeViewRowCell(
  value: unknown,
  source: string,
  fieldKey: string,
): ViewRowCellV1 {
  const cell = requireRowObject(value, source, `cells.${fieldKey}`);
  if (!hasOwn(cell, "value")) {
    viewRowInvariant(source, `cell ${fieldKey} missing value`);
  }
  return Object.freeze({ value: cell.value });
}

function sanitizeGroupValues(
  value: unknown,
  source: string,
): Readonly<Record<string, unknown>> | undefined {
  if (value === undefined) {
    return undefined;
  }
  return Object.freeze({
    ...requireRowObject(value, source, "group_values"),
  });
}

function validateOptionalRowViewSchemaId(
  raw: Record<string, unknown>,
  contract: ViewContract,
  source: string,
) {
  const value = raw.view_schema_id;
  if (value === undefined) {
    return;
  }
  if (value !== contract.viewSchemaId) {
    viewRowInvariant(source, `view_schema_id must be ${contract.viewSchemaId}`);
  }
}

function normalizeViewRowCells({
  allowSparse,
  contract,
  rawCells,
  source,
}: {
  readonly allowSparse: boolean;
  readonly contract: ViewContract;
  readonly rawCells: unknown;
  readonly source: string;
}): Readonly<Record<string, ViewRowCellV1>> {
  const cells = requireRowObject(rawCells, source, "cells");
  const dataFields = declaredDataFieldKeys(contract);
  const technicalFields = technicalFieldKeys(contract);
  const normalized: Record<string, ViewRowCellV1> = {};

  for (const [fieldKey, cell] of Object.entries(cells)) {
    if (technicalFields.has(fieldKey)) {
      viewRowInvariant(source, `technical cell ${fieldKey} is not allowed`);
    }
    if (!dataFields.has(fieldKey)) {
      viewRowInvariant(source, `unknown cell ${fieldKey}`);
    }
    normalized[fieldKey] = sanitizeViewRowCell(cell, source, fieldKey);
  }

  if (!allowSparse) {
    for (const field of contract.fields) {
      if (!hasOwn(cells, field.fieldKey)) {
        viewRowInvariant(source, `missing cell ${field.fieldKey}`);
      }
    }
  }

  return Object.freeze(normalized);
}

export function normalizeViewRowV1(
  contract: ViewContract,
  row: unknown,
  source = contract.viewSchemaId,
): NormalizedViewRowV1 {
  const raw = requireRowObject(row, source, "view_row_v1");
  validateOptionalRowViewSchemaId(raw, contract, source);
  return Object.freeze({
    recordId: requireRowRecordId(raw.record_id, source),
    rowVersion: requireRowVersion(raw.row_version, source),
    viewSchemaId: contract.viewSchemaId,
    cells: normalizeViewRowCells({
      allowSparse: false,
      contract,
      rawCells: raw.cells,
      source,
    }),
    groupValues: sanitizeGroupValues(raw.group_values, source),
  });
}

export function normalizeViewRowPatchV1(
  contract: ViewContract,
  row: unknown,
  source = contract.viewSchemaId,
): NormalizedViewRowV1 {
  const raw = requireRowObject(row, source, "view_row_patch_v1");
  validateOptionalRowViewSchemaId(raw, contract, source);
  return Object.freeze({
    recordId: requireRowRecordId(raw.record_id, source),
    rowVersion: requireRowVersion(raw.row_version, source),
    viewSchemaId: contract.viewSchemaId,
    cells: normalizeViewRowCells({
      allowSparse: true,
      contract,
      rawCells: raw.cells,
      source,
    }),
    groupValues: sanitizeGroupValues(raw.group_values, source),
  });
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
    permitsZeroFieldCreate:
      raw.inline_create?.permits_zero_field_create ?? false,
    requiredReferencePackKeys,
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
