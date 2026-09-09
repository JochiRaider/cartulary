import type {
  DiscoveredImportColumn,
  DiscoveredImportPreview,
  DiscoveredImportUnit,
  ImportMappingRequest,
  ImportSessionResource,
} from "../services/importContractAdapter";

export type NetworkFlowImportDiscovery = {
  readonly sessionId: string;
  readonly session: ImportSessionResource;
  readonly unit: DiscoveredImportUnit;
  readonly preview: DiscoveredImportPreview;
};
export function networkFlowApprovalRequest(
  discovery: Pick<NetworkFlowImportDiscovery, "preview">,
  candidate: {
    readonly target_kind: string;
    readonly extension_profile_id: string;
    readonly owner_mapping_schema_id: string;
    readonly owner_mapping: Record<string, unknown>;
  },
  transactionId: string,
): ImportMappingRequest {
  return {
    ...candidate,
    client_txn_id: transactionId,
    header_row_ref: discovery.preview.header_row_ref,
    data_start_row_ref: discovery.preview.data_start_row_ref,
    source_columns: discovery.preview.columns.map((column) => ({
      ...column,
      field_key: null,
      entity_binding_mode: null,
      transform_id: null,
      transform_options: {},
      empty_value_policy: "omit_field",
    })),
  };
}
export function networkFlowApprovedPreviewMatches(
  unit: DiscoveredImportUnit,
  fingerprint: string,
) {
  return unit.mapping_fingerprint === fingerprint;
}

import type { NetworkFlowMappingCandidate } from "../services/networkFlowContractAdapter";
import {
  networkFlowMappingMetadata,
  networkFlowTimestampMetadata,
} from "../services/networkFlowContractAdapter";

export const ignoredColumnChoice = "__ignore__";

export type NetworkFlowTimestampMode =
  NetworkFlowMappingCandidate["timestamp_profile"]["mode"];

export type NetworkFlowUnknownColumnPolicy = NonNullable<
  NetworkFlowMappingCandidate["unknown_column_policy"]
>;

export type NetworkFlowMappingDraft = {
  readonly sourceProfileId: string;
  readonly unknownColumnPolicy: NetworkFlowUnknownColumnPolicy;
  readonly timestampMode: NetworkFlowTimestampMode;
  readonly timezone: string;
  readonly displayNameOverride: string;
  readonly netflowExportTimeColumnOrdinal: number | null;
  readonly netflowExportTimeMode:
    | "rfc3339"
    | "epoch_seconds"
    | "epoch_milliseconds";
  readonly netflowExporterUptimeColumnOrdinal: number | null;
  readonly columnChoices: Readonly<Record<number, string | null>>;
  readonly unresolvedAliasCollisionOrdinals: readonly number[];
};

export function networkFlowSourceProfile(id: string) {
  return networkFlowMappingMetadata.source_profiles.find(
    (profile) =>
      profile.source_profile_id === id &&
      profile.conformance_status === "required_v1",
  );
}

export const networkFlowMappingFields = (profileId: string) =>
  networkFlowSourceProfile(profileId)?.fields.filter(
    (field) =>
      field.requirement === "required" ||
      field.requirement === "optional_map_when_present",
  ) ?? [];

export const networkFlowRequiredFieldKeys = (profileId: string) =>
  networkFlowMappingFields(profileId)
    .filter((field) => field.requirement === "required")
    .map((field) => field.field_key);

export function createNetworkFlowMappingDraft(
  columns: readonly DiscoveredImportColumn[],
  profileId: string = networkFlowMappingMetadata.source_profiles[0]
    .source_profile_id,
): NetworkFlowMappingDraft {
  const sourceProfile = networkFlowSourceProfile(profileId);
  if (!sourceProfile) throw new Error("Unsupported source profile");
  const fieldOwnerOrdinals = new Map<string, number>();
  const collisionOrdinals = new Set<number>();
  const columnChoices: Record<number, string | null> = {};
  for (const column of columns) {
    const matchKey = sourceAliasMatchKey(
      importHeaderText(column.source_header_text),
    );
    const suggested = networkFlowMappingFields(profileId).find((field) =>
      field.aliases.some((alias) => sourceAliasMatchKey(alias) === matchKey),
    );
    const ordinal = column.source_column_ordinal;
    const existingOwner =
      suggested === undefined
        ? undefined
        : fieldOwnerOrdinals.get(suggested.field_key);
    if (suggested !== undefined && existingOwner !== undefined) {
      columnChoices[existingOwner] = null;
      columnChoices[ordinal] = null;
      collisionOrdinals.add(existingOwner);
      collisionOrdinals.add(ordinal);
      continue;
    }
    columnChoices[ordinal] = suggested?.field_key ?? null;
    if (suggested !== undefined) {
      fieldOwnerOrdinals.set(suggested.field_key, ordinal);
    }
  }
  return {
    sourceProfileId: sourceProfile.source_profile_id,
    unknownColumnPolicy: sourceProfile.default_unknown_column_policy,
    timestampMode: sourceProfile.default_timestamp_profile.mode,
    timezone: sourceProfile.default_timestamp_profile.timezone ?? "",
    displayNameOverride: "",
    netflowExportTimeColumnOrdinal: null,
    netflowExportTimeMode: networkFlowTimestampMetadata.exportTimeModes[0],
    netflowExporterUptimeColumnOrdinal: null,
    columnChoices,
    unresolvedAliasCollisionOrdinals: [...collisionOrdinals].sort(
      (left, right) => left - right,
    ),
  };
}

export function withNetworkFlowColumnChoice(
  draft: NetworkFlowMappingDraft,
  ordinal: number,
  choice: string | null,
): NetworkFlowMappingDraft {
  const columnChoices = { ...draft.columnChoices };
  if (
    choice !== null &&
    choice !== ignoredColumnChoice &&
    Object.entries(columnChoices).some(
      ([candidateOrdinal, candidate]) =>
        Number(candidateOrdinal) !== ordinal && candidate === choice,
    )
  ) {
    for (const [candidateOrdinal, candidate] of Object.entries(columnChoices)) {
      if (candidate === choice) {
        columnChoices[Number(candidateOrdinal)] = null;
      }
    }
  }
  columnChoices[ordinal] = choice;
  return {
    ...draft,
    columnChoices,
    unresolvedAliasCollisionOrdinals:
      choice === null
        ? draft.unresolvedAliasCollisionOrdinals
        : draft.unresolvedAliasCollisionOrdinals.filter(
            (candidate) => candidate !== ordinal,
          ),
  };
}

export function buildNetworkFlowMappingCandidate(
  draft: NetworkFlowMappingDraft,
  columns: readonly DiscoveredImportColumn[],
): NetworkFlowMappingCandidate {
  const sourceProfile = networkFlowSourceProfile(draft.sourceProfileId);
  if (!sourceProfile) throw new Error("Unsupported source profile");
  const fieldMappings: NetworkFlowMappingCandidate["field_mappings"] = [];
  for (const column of columns) {
    const choice = draft.columnChoices[column.source_column_ordinal] ?? null;
    if (
      choice === ignoredColumnChoice ||
      (choice === null &&
        draft.unknownColumnPolicy === "ignore_unmapped_columns")
    ) {
      fieldMappings.push({
        mapping_kind: "ignored_source_column",
        source_column_ordinal: column.source_column_ordinal,
        ignore_reason: "user_ignored",
      });
      continue;
    }
    if (choice === null) {
      continue;
    }
    const field = networkFlowMappingFields(draft.sourceProfileId).find(
      (candidate) => candidate.field_key === choice,
    );
    if (
      field === undefined ||
      field.transform_id === null ||
      field.empty_value_policy === null
    ) {
      continue;
    }
    fieldMappings.push({
      mapping_kind: "source_column",
      field_key: field.field_key,
      source_column_ordinal: column.source_column_ordinal,
      transform_id: field.transform_id,
      empty_value_policy: field.empty_value_policy,
      combinability: "single_source_only",
    });
  }

  const timestampProfile = timestampProfileFromDraft(draft);
  return {
    target_kind: networkFlowMappingMetadata.target_kind,
    target_table_schema_id: networkFlowMappingMetadata.target_table_schema_id,
    source_profile_id: sourceProfile.source_profile_id,
    parser_profile_id: sourceProfile.parser_profile_id,
    unknown_column_policy: draft.unknownColumnPolicy,
    ...(draft.displayNameOverride.trim() === ""
      ? {}
      : { display_name_override: draft.displayNameOverride.trim() }),
    timestamp_profile: timestampProfile,
    field_mappings: fieldMappings,
  };
}

export function mappedRequiredFieldCount(
  draft: NetworkFlowMappingDraft,
): number {
  const mapped = new Set(Object.values(draft.columnChoices));
  return networkFlowRequiredFieldKeys(draft.sourceProfileId).filter(
    (fieldKey) => mapped.has(fieldKey),
  ).length;
}

export function networkFlowMappingDraftReadyForPreview(
  draft: NetworkFlowMappingDraft,
): boolean {
  return (
    networkFlowSourceProfile(draft.sourceProfileId) !== undefined &&
    draft.unresolvedAliasCollisionOrdinals.length === 0 &&
    (draft.timestampMode !== "netflow_sys_uptime_milliseconds" ||
      (draft.netflowExportTimeColumnOrdinal !== null &&
        draft.netflowExporterUptimeColumnOrdinal !== null &&
        draft.netflowExportTimeColumnOrdinal !==
          draft.netflowExporterUptimeColumnOrdinal))
  );
}

export type NetworkFlowMappingIssue = {
  readonly field: string;
  readonly message: string;
};
export function networkFlowMappingIssues(
  draft: NetworkFlowMappingDraft,
): readonly NetworkFlowMappingIssue[] {
  const profile = networkFlowSourceProfile(draft.sourceProfileId);
  if (!profile)
    return [
      {
        field: "sourceProfileId",
        message: "Choose a supported source profile.",
      },
    ];
  const issues: NetworkFlowMappingIssue[] = [];
  const choices = Object.values(draft.columnChoices);
  for (const field of networkFlowRequiredFieldKeys(draft.sourceProfileId)) {
    if (!choices.includes(field))
      issues.push({
        field,
        message: `Map required field ${field.replace("network_flow.", "")}.`,
      });
  }
  for (const [ordinal, field] of Object.entries(draft.columnChoices)) {
    if (
      field &&
      field !== ignoredColumnChoice &&
      (!networkFlowMappingFields(draft.sourceProfileId).some(
        (f) => f.field_key === field,
      ) ||
        choices.filter((v) => v === field).length > 1)
    ) {
      issues.push({
        field: `column:${ordinal}`,
        message: "Choose one supported target for this column.",
      });
    }
    if (
      field === null &&
      draft.unknownColumnPolicy === "reject_unmapped_columns"
    )
      issues.push({
        field: `column:${ordinal}`,
        message:
          "Map or explicitly ignore this column under the selected policy.",
      });
  }
  if (
    !profile.supported_timestamp_modes.some(
      (mode) => mode === draft.timestampMode,
    ) ||
    !networkFlowMappingDraftReadyForPreview(draft)
  )
    issues.push({
      field: "timestampMode",
      message:
        "Resolve alias collisions and choose valid, distinct timestamp source columns.",
    });
  if (draft.timestampMode === "netflow_sys_uptime_milliseconds") {
    for (const ordinal of [
      draft.netflowExportTimeColumnOrdinal,
      draft.netflowExporterUptimeColumnOrdinal,
    ]) {
      if (
        ordinal === null ||
        !Object.hasOwn(draft.columnChoices, ordinal) ||
        ["network_flow.flow_start_utc", "network_flow.flow_end_utc"].includes(
          draft.columnChoices[ordinal] ?? "",
        )
      )
        issues.push({
          field: "timestampMode",
          message:
            "Timestamp context columns must exist and differ from event-time columns.",
        });
    }
  }
  if (
    !profile.supported_unknown_column_policies.some(
      (p) => p === draft.unknownColumnPolicy,
    )
  )
    issues.push({
      field: "unknownColumnPolicy",
      message: "Choose a supported unknown-column policy.",
    });
  return issues;
}

export function sourceColumnLabel(column: DiscoveredImportColumn): string {
  const header = importHeaderText(column.source_header_text);
  return `${header === "" ? "(unnamed)" : header} · column ${column.source_column_ordinal}`;
}

function importHeaderText(
  value: DiscoveredImportColumn["source_header_text"],
): string {
  return value === null ? "" : String(value);
}

function timestampProfileFromDraft(
  draft: NetworkFlowMappingDraft,
): NetworkFlowMappingCandidate["timestamp_profile"] {
  switch (draft.timestampMode) {
    case "epoch_seconds":
      return {
        ...networkFlowTimestampMetadata.defaults.epoch_seconds,
      };
    case "epoch_milliseconds":
      return {
        ...networkFlowTimestampMetadata.defaults.epoch_milliseconds,
      };
    case "netflow_sys_uptime_milliseconds":
      return {
        ...networkFlowTimestampMetadata.defaults
          .netflow_sys_uptime_milliseconds,
        netflow_export_time_column_ordinal:
          draft.netflowExportTimeColumnOrdinal ?? 0,
        netflow_export_time_mode: draft.netflowExportTimeMode,
        netflow_exporter_uptime_at_export_column_ordinal:
          draft.netflowExporterUptimeColumnOrdinal ?? 0,
      };
    default:
      return {
        ...networkFlowTimestampMetadata.defaults.rfc3339,
        timezone: draft.timezone.trim() === "" ? null : draft.timezone.trim(),
        timezone_ruleset_id:
          draft.timezone.trim() === "" || draft.timezone.trim() === "UTC"
            ? null
            : networkFlowTimestampMetadata.timezoneRulesetId,
      };
  }
}

function sourceAliasMatchKey(value: string): string {
  return value
    .trim()
    .replace(/[A-Z]/gu, (character) => character.toLowerCase());
}
