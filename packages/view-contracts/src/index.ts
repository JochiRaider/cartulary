import {
  getViewSchemaRegistryContract,
  listViewSchemaArtifacts,
} from "@cartulary/protocol-ts";

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
  readonly gridEditable: boolean;
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

export type InspectorRouteBindingOwner =
  | "current_row_projection"
  | "view_query_route"
  | "view_row_create_route"
  | "record_patch_route"
  | "record_mark_reviewed_route"
  | "record_supersede_route"
  | "record_delete_route"
  | "record_restore_route"
  | "record_history_route"
  | "record_rollback_route"
  | "record_merge_route"
  | "entity_mention_resolve_route"
  | "evidence_attach_blob_route"
  | "evidence_preview_handle_route"
  | "evidence_download_handle_route";

export type InspectorDisabledCondition =
  | "no_row_selected"
  | "incident_closed"
  | "authorization_lost"
  | "row_version_changed"
  | "record_deleted"
  | "record_merged"
  | "evidence_preview_unavailable"
  | "merge_target_unavailable"
  | "record_not_deleted"
  | "rollback_target_unavailable"
  | "party_text_unavailable"
  | "pivot_target_unavailable";

export type InspectorSuccessResultBehavior =
  | "preserve_selected_row"
  | "retarget_selected_row"
  | "clear_to_no_row_selected"
  | "surface_pivot";

export type InspectorFailureResultBehavior =
  | "show_same_shell_error_preserve_selection"
  | "show_same_shell_error_invalidate_pending_action"
  | "show_same_shell_error_clear_subject";

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
  readonly owner: InspectorRouteBindingOwner;
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
  readonly successResultBehavior: InspectorSuccessResultBehavior;
  readonly failureResultBehavior: InspectorFailureResultBehavior;
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
  readonly minimumCreateFieldSets: readonly (readonly string[])[];
  readonly permitsZeroFieldCreate: boolean;
  readonly requiredReferencePackKeys: readonly string[];
  readonly sortableFieldMap: Readonly<Record<string, true>>;
  readonly filterableFieldMap: Readonly<Record<string, true>>;
  readonly groupableFieldMap: Readonly<Record<string, true>>;
  readonly sortFields: readonly string[];
  readonly sortNullOrder: "last";
  readonly technicalFields: readonly string[];
};

export type WorkbookSurfaceStatus =
  | "required_built_in_sheet"
  | "required_system_view"
  | "standardized_optional_workbook_surface";

export type WorkbookSurfaceKind = "built_in_sheet" | "system_view";

export type WorkbookSurfaceContract = {
  readonly contract: ViewContract;
  readonly requiredReferencePackKeys: readonly string[];
  readonly sourceRecordTypes: readonly string[];
  readonly surfaceKind: WorkbookSurfaceKind;
  readonly surfaceStatus: WorkbookSurfaceStatus;
  readonly title: string;
  readonly viewSchemaId: string;
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
  readonly grid_editable?: boolean;
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
  readonly failure_result_behavior?: string;
  readonly feature_group_key?: string;
  readonly label?: string;
  readonly minimum_incident_role?: string | null;
  readonly mutates?: boolean;
  readonly panel_id?: string;
  readonly requires_confirmation?: boolean;
  readonly route_binding?: RawInspectorRouteBinding;
  readonly seed_bindings?: readonly RawInspectorSeedBinding[];
  readonly success_result_behavior?: string;
};

type RawInspectorRouteBinding = {
  readonly action_key?: string;
  readonly kind?: string;
  readonly owner?: string;
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
    readonly minimum_create_field_sets?: readonly (readonly string[])[];
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

// unit.stage-0.row-02 exposes this narrow parser error contract for malformed
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

function stableKeyMatrix(
  value: unknown,
  source: string,
  label: string,
): readonly (readonly string[])[] {
  if (value === undefined) {
    return Object.freeze([]);
  }
  if (!Array.isArray(value)) {
    viewContractInvariant(source, `${label} must be an array`);
  }
  return Object.freeze(
    value.map((item, index) =>
      stableKeyList(item, source, `${label}[${index + 1}]`),
    ),
  );
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

const inspectorRouteBindingOwners = Object.freeze([
  "current_row_projection",
  "view_query_route",
  "view_row_create_route",
  "record_patch_route",
  "record_mark_reviewed_route",
  "record_supersede_route",
  "record_delete_route",
  "record_restore_route",
  "record_history_route",
  "record_rollback_route",
  "record_merge_route",
  "entity_mention_resolve_route",
  "evidence_attach_blob_route",
  "evidence_preview_handle_route",
  "evidence_download_handle_route",
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
  "record_not_deleted",
  "rollback_target_unavailable",
  "party_text_unavailable",
  "pivot_target_unavailable",
] as const);

const inspectorSuccessResultBehaviors = Object.freeze([
  "preserve_selected_row",
  "retarget_selected_row",
  "clear_to_no_row_selected",
  "surface_pivot",
] as const);

const inspectorFailureResultBehaviors = Object.freeze([
  "show_same_shell_error_preserve_selection",
  "show_same_shell_error_invalidate_pending_action",
  "show_same_shell_error_clear_subject",
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

const inspectorFeatureRegistryByViewSchemaId = Object.freeze({
  "cartulary.view.assessments.v1": [
    "details.read",
    "relationships.read",
    "history.read",
    "record.delete",
    "record.restore",
    "history.rollback",
    "assessment.subject_pivot",
    "assessment.prior_history",
    "assessment.support_refs.manage",
    "evidence.refs.manage",
    "create_related.task_request",
    "create_related.decision",
  ],
  "cartulary.view.comm_log.v1": [
    "details.read",
    "relationships.read",
    "evidence.read",
    "history.read",
    "record.delete",
    "record.restore",
    "history.rollback",
    "comm.decisions.link",
    "comm.action_tasks.link",
    "comm.parties.manage",
    "comm.next_report.manage",
    "create_related.task_request",
    "create_related.status_review",
  ],
  "cartulary.view.decisions.v1": [
    "details.read",
    "relationships.read",
    "evidence.read",
    "history.read",
    "record.delete",
    "record.restore",
    "history.rollback",
    "decision.support_refs.manage",
    "decision.affected_records.manage",
    "decision.status.transition",
    "decision.supersede",
    "create_related.task_request",
    "create_related.comm_log",
    "create_related.status_review",
  ],
  "cartulary.view.evidence.v1": [
    "details.read",
    "relationships.read",
    "evidence.read",
    "history.read",
    "record.delete",
    "record.restore",
    "history.rollback",
    "evidence.preview_handle",
    "evidence.download_handle",
    "evidence.attach_blob",
    "party.collector.link",
    "party.source.link",
    "party.reference.clear",
    "relationships.manage",
    "surface_pivot.linked_records",
    "surface_pivot.timeline",
    "create_related.note",
    "create_related.task_request",
    "create_related.decision",
  ],
  "cartulary.view.findings.v1": [
    "details.read",
    "relationships.read",
    "evidence.read",
    "history.read",
    "record.delete",
    "record.restore",
    "history.rollback",
    "finding.support_refs.manage",
    "finding.contradictory_refs.manage",
    "finding.evidence_refs.manage",
    "finding.owner.manage",
    "finding.close_or_reopen",
    "create_related.task_request",
    "create_related.decision",
  ],
  "cartulary.view.forensic_keywords.v1": [
    "details.read",
    "relationships.read",
    "evidence.read",
    "history.read",
    "record.delete",
    "record.restore",
    "history.rollback",
    "keyword.evidence_refs.manage",
    "keyword.timeline_rows.link",
    "keyword.findings.link",
    "create_related.task_request",
  ],
  "cartulary.view.handoff.v1": [
    "details.read",
    "relationships.read",
    "history.read",
    "record.delete",
    "record.restore",
    "history.rollback",
    "handoff.acknowledge",
    "handoff.open_tasks.review",
    "handoff.open_decisions.review",
    "handoff.risks.review",
    "handoff.next_checks.manage",
    "create_related.task_request",
    "create_related.status_review",
  ],
  "cartulary.view.hosts.v1": [
    "details.read",
    "relationships.read",
    "evidence.read",
    "history.read",
    "record.delete",
    "record.restore",
    "history.rollback",
    "entity.aliases.read",
    "entity.relationships.manage",
    "entity.merge",
    "surface_pivot.timeline",
    "surface_pivot.evidence",
    "surface_pivot.assessments",
    "create_related.note",
    "create_related.task_request",
    "create_related.decision",
  ],
  "cartulary.view.identities.v1": [
    "details.read",
    "relationships.read",
    "evidence.read",
    "history.read",
    "record.delete",
    "record.restore",
    "history.rollback",
    "entity.aliases.read",
    "entity.relationships.manage",
    "entity.merge",
    "surface_pivot.timeline",
    "surface_pivot.evidence",
    "surface_pivot.assessments",
    "create_related.note",
    "create_related.task_request",
    "create_related.decision",
  ],
  "cartulary.view.indicators.v1": [
    "details.read",
    "relationships.read",
    "history.read",
    "record.delete",
    "record.restore",
    "history.rollback",
    "indicator.observations.pivot",
    "indicator.lifecycle.read",
    "relationships.manage",
    "create_related.task_request",
    "create_related.decision",
  ],
  "cartulary.view.investigative_queries.v1": [
    "details.read",
    "relationships.read",
    "evidence.read",
    "history.read",
    "record.delete",
    "record.restore",
    "history.rollback",
    "query.source.link",
    "query.result.link",
    "query.evidence_refs.manage",
    "query.findings.link",
    "create_related.task_request",
  ],
  "cartulary.view.lesson.v1": [
    "details.read",
    "relationships.read",
    "evidence.read",
    "history.read",
    "record.delete",
    "record.restore",
    "history.rollback",
    "lesson.followup_tasks.manage",
    "lesson.evidence_refs.manage",
    "lesson.owner.manage",
    "lesson.close_or_reopen",
    "create_related.task_request",
  ],
  "cartulary.view.notes.v1": [
    "details.read",
    "relationships.read",
    "evidence.read",
    "history.read",
    "record.delete",
    "record.restore",
    "history.rollback",
    "artifact.source_links.manage",
    "artifact.evidence_refs.manage",
    "artifact.tags.manage",
    "artifact.related_notes.manage",
    "surface_pivot.source_records",
    "create_related.task_request",
    "create_related.decision",
  ],
  "cartulary.view.parties.v1": [
    "details.read",
    "relationships.read",
    "history.read",
    "record.delete",
    "record.restore",
    "history.rollback",
    "party.usage_pivot.requester",
    "party.usage_pivot.collector_source",
    "party.usage_pivot.audience_attendee",
    "party.usage_pivot.owner_stakeholder",
    "party.reference.link",
    "party.reference.clear",
    "party.reference.clear_both",
  ],
  "cartulary.view.status_review.v1": [
    "details.read",
    "relationships.read",
    "evidence.read",
    "history.read",
    "record.delete",
    "record.restore",
    "history.rollback",
    "status_review.blocked_tasks.review",
    "status_review.pending_evidence.review",
    "status_review.open_decisions.review",
    "status_review.risks.review",
    "status_review.next_report.manage",
    "create_related.task_request",
    "create_related.comm_log",
  ],
  "cartulary.view.task_requests.v1": [
    "details.read",
    "relationships.read",
    "evidence.read",
    "history.read",
    "record.delete",
    "record.restore",
    "history.rollback",
    "task.links.manage",
    "task.requester_party.link",
    "task.requester_party.clear",
    "task.decision.link",
    "task.decision.clear",
    "task.status.transition",
    "create_related.comm_log",
    "create_related.status_review",
    "create_related.lesson",
  ],
  "cartulary.view.timeline.v2": [
    "details.read",
    "relationships.read",
    "evidence.read",
    "history.read",
    "record.delete",
    "record.restore",
    "history.rollback",
    "entity_mentions.resolve",
    "entity_mentions.create_host",
    "entity_mentions.create_identity",
    "entity_mentions.dismiss",
    "entity_mentions.restore",
    "indicator.observations.manage",
    "relationships.manage",
    "evidence.attach_blob",
    "evidence.preview_handle",
    "evidence.download_handle",
    "timeline.mark_reviewed",
    "timeline.supersede",
    "create_related.note",
    "create_related.task_request",
    "create_related.decision",
    "create_related.evidence",
    "create_related.comm_log",
    "create_related.handoff",
    "create_related.status_review",
    "create_related.lesson",
  ],
} as const satisfies Readonly<Record<string, readonly string[]>>);

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
      const successResultBehavior = requireEnumValue(
        group.success_result_behavior,
        inspectorSuccessResultBehaviors,
        source,
        `${label}.success_result_behavior`,
      );
      const failureResultBehavior = requireEnumValue(
        group.failure_result_behavior,
        inspectorFailureResultBehaviors,
        source,
        `${label}.failure_result_behavior`,
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
        successResultBehavior,
        failureResultBehavior,
      });
    }),
  );
  validateInspectorFeatureRegistry(
    configViewSchemaId,
    featureKeys,
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
  featureKeys: ReadonlySet<string>,
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
  if (featureKeys.size !== expected.length) {
    viewContractInvariant(
      source,
      `${label} must contain exactly ${expected.length} declared feature groups for ${viewSchemaId}, got ${featureKeys.size}`,
    );
  }
  for (const key of expected) {
    if (!featureKeys.has(key)) {
      viewContractInvariant(
        source,
        `${label} missing required feature_group_key ${key} for ${viewSchemaId}`,
      );
    }
  }
  const expectedSet = new Set<string>(expected);
  for (const key of featureKeys) {
    if (!expectedSet.has(key)) {
      viewContractInvariant(
        source,
        `${label} contains undeclared feature_group_key ${key} for ${viewSchemaId}`,
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
  if (!/^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/u.test(key)) {
    viewContractInvariant(
      source,
      `${label} must be ASCII lower snake or dotted key`,
    );
  }
  return key;
}

function requireFieldKey(value: unknown, source: string, label: string) {
  const key = requireStableKey(value, source, label);
  if (!/^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/u.test(key)) {
    viewContractInvariant(source, `${label} must be a stable field_key`);
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

function sanitizeGroupValues({
  allowSparse,
  contract,
  source,
  value,
}: {
  readonly allowSparse: boolean;
  readonly contract: ViewContract;
  readonly source: string;
  readonly value: unknown;
}): Readonly<Record<string, unknown>> | undefined {
  const groupingFields = new Set(contract.groupingFields);
  if (value === undefined) {
    if (!allowSparse && groupingFields.size > 0) {
      viewRowInvariant(source, "group_values is required for grouped schema");
    }
    return undefined;
  }
  if (groupingFields.size === 0) {
    viewRowInvariant(
      source,
      "group_values is not allowed for ungrouped schema",
    );
  }
  const raw = requireRowObject(value, source, "group_values");
  for (const fieldKey of Object.keys(raw)) {
    if (!groupingFields.has(fieldKey)) {
      viewRowInvariant(source, `unknown group_values field ${fieldKey}`);
    }
  }
  if (!allowSparse) {
    for (const fieldKey of groupingFields) {
      if (!hasOwn(raw, fieldKey)) {
        viewRowInvariant(source, `missing group_values field ${fieldKey}`);
      }
    }
  }
  return Object.freeze({ ...raw });
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
    groupValues: sanitizeGroupValues({
      allowSparse: false,
      contract,
      source,
      value: raw.group_values,
    }),
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
    groupValues: sanitizeGroupValues({
      allowSparse: true,
      contract,
      source,
      value: raw.group_values,
    }),
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
        gridEditable: field.grid_editable ?? false,
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
  const minimumCreateFieldSets = stableKeyMatrix(
    raw.inline_create?.minimum_create_field_sets,
    source,
    "inline_create.minimum_create_field_sets",
  );
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

const workbookSurfaceRegistryContract = getViewSchemaRegistryContract();

function sameOrderedValues(
  left: readonly string[],
  right: readonly string[],
): boolean {
  return (
    left.length === right.length &&
    left.every((value, index) => value === right[index])
  );
}

export function buildWorkbookSurfaceContracts(
  sourceContracts: readonly ViewContract[] = contracts,
): readonly WorkbookSurfaceContract[] {
  const contractsById = new Map(
    sourceContracts.map((contract) => [contract.viewSchemaId, contract]),
  );
  return Object.freeze(
    workbookSurfaceRegistryContract.view_schemas.flatMap((entry) => {
      const contract = contractsById.get(entry.view_schema_id);
      if (!contract) {
        if (entry.surface_status === "standardized_optional_workbook_surface") {
          return [];
        }
        throw new Error(
          `Missing workbook surface contract: ${entry.view_schema_id}`,
        );
      }
      if (contract.surfaceKind !== entry.surface_kind) {
        throw new Error(
          `Workbook surface ${entry.view_schema_id} has surface_kind ${contract.surfaceKind}, expected ${entry.surface_kind}`,
        );
      }
      if (
        !sameOrderedValues(
          contract.requiredReferencePackKeys,
          entry.required_reference_pack_keys,
        )
      ) {
        throw new Error(
          `Workbook surface ${entry.view_schema_id} required_reference_pack_keys do not match its registry entry`,
        );
      }
      return [
        Object.freeze({
          contract,
          requiredReferencePackKeys: Object.freeze([
            ...entry.required_reference_pack_keys,
          ]),
          sourceRecordTypes: Object.freeze([...entry.source_record_types]),
          surfaceKind: entry.surface_kind,
          surfaceStatus: entry.surface_status,
          title: entry.title,
          viewSchemaId: entry.view_schema_id,
        }),
      ];
    }),
  );
}

const workbookSurfaceContracts = buildWorkbookSurfaceContracts();
const workbookSurfaceContractIndex = new Map(
  workbookSurfaceContracts.map((entry) => [entry.viewSchemaId, entry]),
);

export function listWorkbookSurfaceContracts(): readonly WorkbookSurfaceContract[] {
  return workbookSurfaceContracts;
}

export function getWorkbookSurfaceContract(
  viewSchemaId: string,
): WorkbookSurfaceContract | undefined {
  return workbookSurfaceContractIndex.get(viewSchemaId);
}

export function requireWorkbookSurfaceContract(
  viewSchemaId: string,
): WorkbookSurfaceContract {
  const entry = getWorkbookSurfaceContract(viewSchemaId);
  if (!entry) {
    throw new Error(`Unknown workbook surface contract: ${viewSchemaId}`);
  }
  return entry;
}

function requiredRegistryViewSchemaId(viewSchemaId: string): string {
  const entry = workbookSurfaceRegistryContract.view_schemas.find(
    (candidate) => candidate.view_schema_id === viewSchemaId,
  );
  if (!entry) {
    throw new Error(`Missing workbook surface registry entry: ${viewSchemaId}`);
  }
  return entry.view_schema_id;
}

export const timelineViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.timeline.v2",
);
export const hostsViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.hosts.v1",
);
export const identitiesViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.identities.v1",
);
export const evidenceViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.evidence.v1",
);
export const notesViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.notes.v1",
);
export const indicatorsViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.indicators.v1",
);
export const assessmentsViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.assessments.v1",
);
export const taskRequestsViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.task_requests.v1",
);
export const decisionsViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.decisions.v1",
);
export const partiesViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.parties.v1",
);
export const commLogViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.comm_log.v1",
);
export const handoffViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.handoff.v1",
);
export const statusReviewViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.status_review.v1",
);
export const lessonViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.lesson.v1",
);
export const findingsViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.findings.v1",
);
export const investigativeQueriesViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.investigative_queries.v1",
);
export const forensicKeywordsViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.forensic_keywords.v1",
);

function workbookSurfaceIdsByStatus(
  status: WorkbookSurfaceStatus,
): readonly string[] {
  return Object.freeze(
    workbookSurfaceRegistryContract.view_schemas
      .filter((entry) => entry.surface_status === status)
      .map((entry) => entry.view_schema_id),
  );
}

export const requiredBuiltInWorkbookSurfaceIds = workbookSurfaceIdsByStatus(
  "required_built_in_sheet",
);
export const requiredSystemWorkbookSurfaceIds = workbookSurfaceIdsByStatus(
  "required_system_view",
);
export const optionalStandardizedWorkbookSurfaceIds =
  workbookSurfaceIdsByStatus("standardized_optional_workbook_surface");

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
