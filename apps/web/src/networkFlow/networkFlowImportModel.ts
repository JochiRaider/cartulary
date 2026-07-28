import type { DiscoveredImportColumn } from "../imports/importCoordinator";
import type { NetworkFlowMappingCandidate } from "../services/networkFlowContractAdapter";
import { networkFlowMappingMetadata } from "../services/networkFlowContractAdapter";

export const ignoredColumnChoice = "__ignore__";

export type NetworkFlowTimestampMode =
  | "rfc3339"
  | "epoch_seconds"
  | "epoch_milliseconds"
  | "netflow_sys_uptime_milliseconds";

export type NetworkFlowUnknownColumnPolicy =
  | "preserve_unmapped_raw"
  | "reject_unmapped_columns"
  | "ignore_unmapped_columns";

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

const sourceProfile = networkFlowMappingMetadata.source_profiles[0];

export const networkFlowMappingFields = sourceProfile.fields.filter(
  (field) =>
    field.requirement === "required" ||
    field.requirement === "optional_map_when_present",
);

export const networkFlowRequiredFieldKeys = sourceProfile.fields
  .filter((field) => field.requirement === "required")
  .map((field) => field.field_key);

export function createNetworkFlowMappingDraft(
  columns: readonly DiscoveredImportColumn[],
): NetworkFlowMappingDraft {
  const fieldOwnerOrdinals = new Map<string, number>();
  const collisionOrdinals = new Set<number>();
  const columnChoices: Record<number, string | null> = {};
  for (const column of columns) {
    const matchKey = sourceAliasMatchKey(
      importHeaderText(column.source_header_text),
    );
    const suggested = networkFlowMappingFields.find((field) =>
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
    netflowExportTimeMode: "rfc3339",
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
    const field = networkFlowMappingFields.find(
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
    source_profile_id: draft.sourceProfileId as "cisco_sna_netflow_csv_v1",
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
  return networkFlowRequiredFieldKeys.filter((fieldKey) => mapped.has(fieldKey))
    .length;
}

export function networkFlowMappingDraftReadyForPreview(
  draft: NetworkFlowMappingDraft,
): boolean {
  return (
    draft.unresolvedAliasCollisionOrdinals.length === 0 &&
    (draft.timestampMode !== "netflow_sys_uptime_milliseconds" ||
      (draft.netflowExportTimeColumnOrdinal !== null &&
        draft.netflowExporterUptimeColumnOrdinal !== null &&
        draft.netflowExportTimeColumnOrdinal !==
          draft.netflowExporterUptimeColumnOrdinal))
  );
}

export function sourceColumnLabel(column: DiscoveredImportColumn): string {
  const header = importHeaderText(column.source_header_text).trim();
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
        schema_id: "cartulary.network_flow.timestamp_profile.v1",
        mode: "epoch_seconds",
        precision: "seconds",
      };
    case "epoch_milliseconds":
      return {
        schema_id: "cartulary.network_flow.timestamp_profile.v1",
        mode: "epoch_milliseconds",
        precision: "milliseconds",
      };
    case "netflow_sys_uptime_milliseconds":
      return {
        schema_id: "cartulary.network_flow.timestamp_profile.v1",
        mode: "netflow_sys_uptime_milliseconds",
        precision: "milliseconds",
        netflow_export_time_column_ordinal:
          draft.netflowExportTimeColumnOrdinal ?? 0,
        netflow_export_time_mode: draft.netflowExportTimeMode,
        netflow_exporter_uptime_at_export_column_ordinal:
          draft.netflowExporterUptimeColumnOrdinal ?? 0,
      };
    default:
      return {
        schema_id: "cartulary.network_flow.timestamp_profile.v1",
        mode: "rfc3339",
        precision: "microseconds",
        timezone: draft.timezone.trim() === "" ? null : draft.timezone.trim(),
        timezone_ruleset_id:
          draft.timezone.trim() === "" || draft.timezone.trim() === "UTC"
            ? null
            : "tzdb-2026c",
        ambiguous_local_time_policy: "reject",
        local_time_gap_policy: "reject",
      };
  }
}

function sourceAliasMatchKey(value: string): string {
  return value
    .trim()
    .replace(/[A-Z]/gu, (character) => character.toLowerCase());
}
