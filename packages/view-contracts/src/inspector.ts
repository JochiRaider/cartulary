import type {
  ViewSchemaSourceInspectorConfig,
  ViewSchemaSourceInspectorFeatureGroup,
  ViewSchemaSourceInspectorRouteBinding,
  ViewSchemaSourceInspectorSeedBinding,
} from "@cartulary/protocol-ts/view-schemas";
import { viewInspectorRegistry } from "@cartulary/protocol-ts/view-schemas";
import {
  hasOwn,
  requireContractBoolean,
  requireContractObject,
  requireEnumValue,
  requireFieldKey,
  requireInspectorKey,
  requireStableKey,
  viewContractInvariant,
} from "./invariants.js";
import type {
  InspectorConfig,
  InspectorDisabledCondition,
  InspectorFeatureGroup,
  InspectorPanel,
  InspectorPanelId,
  InspectorRouteBinding,
  InspectorSeedBinding,
  InspectorSeedSource,
} from "./types.js";

type RawInspectorConfig = ViewSchemaSourceInspectorConfig;
type RawInspectorPanel = ViewSchemaSourceInspectorConfig["panels"][number];
type RawInspectorFeatureGroup = ViewSchemaSourceInspectorFeatureGroup;
type RawInspectorRouteBinding = ViewSchemaSourceInspectorRouteBinding;
type RawInspectorSeedBinding = ViewSchemaSourceInspectorSeedBinding;

const inspectorPanelIds = viewInspectorRegistry.vocabularies.panels;
const inspectorRouteBindingKinds =
  viewInspectorRegistry.vocabularies.route_kinds;
const inspectorRouteBindingOwners =
  viewInspectorRegistry.vocabularies.route_owners;
const inspectorDisabledConditions =
  viewInspectorRegistry.vocabularies.disabled_conditions;
const inspectorSuccessResultBehaviors =
  viewInspectorRegistry.vocabularies.success_result_behaviors;
const inspectorFailureResultBehaviors =
  viewInspectorRegistry.vocabularies.failure_result_behaviors;
const inspectorSeedSourceKinds =
  viewInspectorRegistry.vocabularies.seed_source_kinds;
const incidentRoles = viewInspectorRegistry.vocabularies.incident_roles;
const inspectorFeatureRegistryByViewSchemaId =
  viewInspectorRegistry.view_feature_keys;
const specializedInspectorFeatures = viewInspectorRegistry.specialized_features;

export function parseInspectorConfig(
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
        routeBinding: parseInspectorRouteBinding(
          group.route_binding,
          source,
          `${label}.route_binding`,
        ),
        seedBindings: parseInspectorSeedBindings(
          group.seed_bindings,
          source,
          `${label}.seed_bindings`,
        ),
        disabledWhen: parseInspectorDisabledConditions(
          group.disabled_when,
          source,
          `${label}.disabled_when`,
        ),
        successResultBehavior: requireEnumValue(
          group.success_result_behavior,
          inspectorSuccessResultBehaviors,
          source,
          `${label}.success_result_behavior`,
        ),
        failureResultBehavior: requireEnumValue(
          group.failure_result_behavior,
          inspectorFailureResultBehaviors,
          source,
          `${label}.failure_result_behavior`,
        ),
      });
    }),
  );
  validateInspectorFeatureRegistry(
    configViewSchemaId,
    featureGroups,
    source,
    "inspector_config.feature_groups",
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
    owner: requireEnumValue(
      route.owner,
      inspectorRouteBindingOwners,
      source,
      `${label}.owner`,
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
        targetFieldKey: requireFieldKey(
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
      : requireFieldKey(
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

function validateInspectorFeatureRegistry(
  viewSchemaId: string,
  featureGroups: readonly InspectorFeatureGroup[],
  source: string,
  label: string,
) {
  const expected =
    inspectorFeatureRegistryByViewSchemaId[
      viewSchemaId as keyof typeof inspectorFeatureRegistryByViewSchemaId
    ];
  if (expected === undefined) {
    return;
  }
  if (featureGroups.length !== expected.length) {
    viewContractInvariant(
      source,
      `${label} must contain exactly ${expected.length} ordered feature groups for ${viewSchemaId}, got ${featureGroups.length}`,
    );
  }
  expected.forEach((key, index) => {
    if (featureGroups[index]?.featureGroupKey !== key) {
      viewContractInvariant(
        source,
        `${label}[${index + 1}].feature_group_key must be ${key} for ${viewSchemaId}`,
      );
    }
  });
  for (const specialized of specializedInspectorFeatures) {
    if (specialized.view_schema_id !== viewSchemaId) continue;
    const group = featureGroups.find(
      (candidate) =>
        candidate.featureGroupKey === specialized.feature_group_key,
    );
    if (group === undefined) {
      viewContractInvariant(
        source,
        `${label} missing specialized feature_group_key ${specialized.feature_group_key} for ${viewSchemaId}`,
      );
    }
    const route = group.routeBinding;
    const matches =
      group.panelId === specialized.panel_id &&
      group.minimumIncidentRole === specialized.minimum_incident_role &&
      group.mutates === specialized.mutates &&
      group.requiresConfirmation === specialized.requires_confirmation &&
      route.kind === specialized.route_binding_kind &&
      route.owner === specialized.route_binding_owner &&
      route.actionKey === specialized.action_key &&
      route.targetViewSchemaId === undefined &&
      group.seedBindings.length === 0 &&
      group.disabledWhen.length === specialized.disabled_when.length &&
      group.disabledWhen.every(
        (condition, index) => condition === specialized.disabled_when[index],
      ) &&
      group.successResultBehavior === specialized.success_result_behavior &&
      group.failureResultBehavior === specialized.failure_result_behavior;
    if (!matches) {
      viewContractInvariant(
        source,
        `${label} specialized feature_group_key ${specialized.feature_group_key} does not match the owner registry`,
      );
    }
  }
}
