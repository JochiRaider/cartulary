import type {
  GridColumn,
  GridDataRow,
  GridSurfaceIdentity,
} from "@cartulary/grid-adapter";
import {
  networkAnalysisDiagnosticTestId,
  networkAnalysisRowTestId,
} from "@cartulary/ui-contracts";
import type { CSSProperties, ReactNode } from "react";
import {
  type NetworkFlowDiagnostic,
  type NetworkFlowRow,
  networkFlowPresentationMetadata,
} from "../services/networkFlowContractAdapter";
import {
  networkAnalysisWorkspaceKey,
  networkFlowActivityProfileId,
} from "./networkFlowClient";

export type NetworkFlowGridSchemaId =
  | "network_flow.accepted_rows.v1"
  | "network_flow.rejected_rows.v1"
  | "network_flow.graph_contributors.v1";

export type NetworkFlowPresentationColumn = {
  readonly copyable: boolean;
  readonly default_order: number;
  readonly default_visible: boolean;
  readonly default_width_px: number;
  readonly field_key: string;
  readonly filter_operators: readonly string[];
  readonly inspector_only: boolean;
  readonly label_key: string;
  readonly link_contexts: readonly string[];
  readonly minimum_width_px: number;
  readonly renderer_kind: string;
  readonly sortable: boolean;
  readonly value_kind: string;
};

export function networkFlowGridSurface(
  gridSchemaId: NetworkFlowGridSchemaId,
): Extract<GridSurfaceIdentity, { readonly kind: "extension_grid" }> {
  return {
    kind: "extension_grid",
    extensionProfileId: networkFlowActivityProfileId,
    workspaceKey: networkAnalysisWorkspaceKey,
    gridSchemaId,
  };
}

export function networkFlowPresentationColumns(
  gridSchemaId: Exclude<
    NetworkFlowGridSchemaId,
    "network_flow.graph_contributors.v1"
  >,
): readonly NetworkFlowPresentationColumn[] {
  const schema = networkFlowPresentationMetadata.grid_schemas.find(
    (candidate) => candidate.grid_schema_id === gridSchemaId,
  );
  if (schema === undefined || !("columns" in schema)) {
    throw new Error(`Missing Network Flow presentation schema ${gridSchemaId}`);
  }
  return [...schema.columns].sort(
    (left, right) => left.default_order - right.default_order,
  );
}

export function compileNetworkFlowColumns<Row extends object>(options: {
  readonly gridSchemaId:
    | "network_flow.accepted_rows.v1"
    | "network_flow.rejected_rows.v1";
  readonly orderedVisibleFieldKeys: readonly string[];
  readonly widths: Readonly<Record<string, number>>;
}): readonly GridColumn<Row>[] {
  const metadata = networkFlowPresentationColumns(options.gridSchemaId);
  const byFieldKey = new Map(
    metadata.map((column) => [column.field_key, column]),
  );
  return options.orderedVisibleFieldKeys.flatMap((fieldKey) => {
    const column = byFieldKey.get(fieldKey);
    if (column === undefined || column.inspector_only) {
      return [];
    }
    return [
      {
        fieldKey: column.field_key,
        label: networkFlowColumnLabel(column.label_key),
        minWidth: column.minimum_width_px,
        width: options.widths[column.field_key] ?? column.default_width_px,
        valueKind: column.value_kind,
        ...(column.sortable
          ? { sortableFieldKey: column.field_key }
          : { sortDisabled: true }),
        align: networkFlowColumnAlignment(column.renderer_kind),
        renderCell: ({ row }: { readonly row: Row }) =>
          renderNetworkFlowCell(row, column),
        getClipboardValue: (row: Row) =>
          column.copyable ? networkFlowClipboardValue(row, column) : "",
      } satisfies GridColumn<Row>,
    ];
  });
}

export function networkFlowRowsForGrid(
  rows: readonly NetworkFlowRow[],
): readonly GridDataRow<NetworkFlowRow>[] {
  return rows.map((row) => ({
    kind: "data",
    rowIdentity: {
      kind: "extension_resource",
      extensionProfileId: networkFlowActivityProfileId,
      resourceKind: "network_flow_row",
      resourceId: row.network_flow_row_id,
    },
    data: row,
    gutterContent: row.source_row_number,
    gutterLabel: `Source row ${row.source_row_number}`,
    testId: networkAnalysisRowTestId(row.network_flow_row_id),
  }));
}

export function networkFlowDiagnosticsForGrid(
  diagnostics: readonly NetworkFlowDiagnostic[],
): readonly GridDataRow<NetworkFlowDiagnostic>[] {
  return diagnostics.map((diagnostic) => ({
    kind: "data",
    rowIdentity: {
      kind: "extension_resource",
      extensionProfileId: networkFlowActivityProfileId,
      resourceKind: "network_flow_rejected_row",
      resourceId: diagnostic.diagnostic_id,
    },
    data: diagnostic,
    testId: networkAnalysisDiagnosticTestId(diagnostic.diagnostic_id),
  }));
}

export function networkFlowClipboardValue(
  row: object,
  column: NetworkFlowPresentationColumn,
): string | number {
  if (column.renderer_kind === "diagnostic_message") {
    return localizedNetworkFlowDiagnosticMessage(row as NetworkFlowDiagnostic);
  }
  const value = networkFlowFieldValue(row, column.field_key);
  if (value === null || value === undefined) {
    return "";
  }
  if (typeof value === "object") {
    return canonicalJSON(value);
  }
  return typeof value === "number" ? value : String(value);
}

export function networkFlowDisplayValue(
  row: object,
  column: NetworkFlowPresentationColumn,
): string {
  const canonical = networkFlowClipboardValue(row, column);
  if (canonical === "") {
    return "—";
  }
  const value = networkFlowFieldValue(row, column.field_key);
  switch (column.renderer_kind) {
    case "decimal_counter":
      return formatDecimalCounter(String(canonical));
    case "ip_protocol":
      return formatIPProtocol(Number(value));
    case "tcp_flags":
      return formatTCPFlags(Number(value));
    case "source_column_ordinal":
      return value === null ? "Row" : String(value);
    default:
      return String(canonical);
  }
}

export function localizedNetworkFlowDiagnosticMessage(
  diagnostic: NetworkFlowDiagnostic,
): string {
  const keyParts = diagnostic.message_key.match(
    /^network_flow\.diagnostic\.([a-z0-9_]+)\.([a-z0-9_]+)$/u,
  );
  const localized =
    keyParts?.[1] === diagnostic.error_code &&
    keyParts?.[2] === diagnostic.reason_code
      ? diagnosticMessages[keyParts[1]]
      : undefined;
  if (localized !== undefined) {
    return localized;
  }
  return diagnostic.message.trim() === ""
    ? "The source value could not be accepted."
    : diagnostic.message;
}

export function networkFlowColumnLabel(labelKey: string): string {
  return columnLabels[labelKey] ?? labelKey;
}

function renderNetworkFlowCell(
  row: object,
  column: NetworkFlowPresentationColumn,
): ReactNode {
  const canonical = networkFlowClipboardValue(row, column);
  const display = networkFlowDisplayValue(row, column);
  const label = networkFlowColumnLabel(column.label_key);
  return (
    <span
      style={cellValueStyle}
      title={canonical === "" ? "No value" : String(canonical)}
    >
      <span aria-hidden="true">{display}</span>
      <span style={visuallyHiddenStyle}>
        {label}: {canonical === "" ? "No value" : canonical}
      </span>
    </span>
  );
}

function networkFlowFieldValue(row: object, fieldKey: string): unknown {
  return (row as Record<string, unknown>)[fieldKey];
}

function networkFlowColumnAlignment(rendererKind: string): "left" | "right" {
  return numericRenderers.has(rendererKind) ? "right" : "left";
}

function canonicalJSON(value: object): string {
  if (Array.isArray(value)) {
    return `[${value.map((item) => canonicalJSONValue(item)).join(",")}]`;
  }
  return `{${Object.keys(value)
    .sort()
    .map(
      (key) =>
        `${JSON.stringify(key)}:${canonicalJSONValue(
          (value as Record<string, unknown>)[key],
        )}`,
    )
    .join(",")}}`;
}

function canonicalJSONValue(value: unknown): string {
  return value !== null && typeof value === "object"
    ? canonicalJSON(value)
    : JSON.stringify(value);
}

function formatDecimalCounter(value: string): string {
  try {
    return new Intl.NumberFormat("en-US").format(BigInt(value));
  } catch {
    return value;
  }
}

function formatIPProtocol(value: number): string {
  const label = protocolLabels[value];
  return label === undefined ? String(value) : `${label} (${value})`;
}

function formatTCPFlags(value: number): string {
  if (!Number.isInteger(value) || value < 0 || value > 255) {
    return String(value);
  }
  const labels = tcpFlagLabels
    .filter(([mask]) => (value & mask) !== 0)
    .map(([, label]) => label);
  return labels.length === 0 ? "None (0)" : `${labels.join(" ")} (${value})`;
}

const numericRenderers = new Set([
  "decimal_counter",
  "ip_protocol",
  "port",
  "source_column_ordinal",
  "source_row_number",
  "tcp_flags",
]);

const protocolLabels: Readonly<Record<number, string>> = {
  1: "ICMP",
  6: "TCP",
  17: "UDP",
  41: "IPv6",
  47: "GRE",
  50: "ESP",
  58: "ICMPv6",
};

const tcpFlagLabels: ReadonlyArray<readonly [number, string]> = [
  [1, "FIN"],
  [2, "SYN"],
  [4, "RST"],
  [8, "PSH"],
  [16, "ACK"],
  [32, "URG"],
  [64, "ECE"],
  [128, "CWR"],
];

const diagnosticMessages: Readonly<Record<string, string>> = {
  network_flow_csv_field_count_mismatch:
    "The row has a different number of columns than the header.",
  network_flow_end_before_start:
    "The flow end time occurs before the flow start time.",
  network_flow_invalid_counter: "The counter is not a valid unsigned value.",
  network_flow_invalid_ip: "The value is not a valid IP address.",
  network_flow_invalid_port: "The value is not a valid network port.",
  network_flow_invalid_protocol: "The value is not a valid IP protocol.",
  network_flow_invalid_timestamp: "The value is not a valid timestamp.",
  network_flow_resource_limit_exceeded:
    "The source value exceeds an allowed resource limit.",
};

const columnLabels: Readonly<Record<string, string>> = {
  "network_flow.column.source_row_number": "Source row",
  "network_flow.column.flow_start_utc": "Flow start",
  "network_flow.column.flow_end_utc": "Flow end",
  "network_flow.column.src_ip": "Source IP",
  "network_flow.column.src_port": "Source port",
  "network_flow.column.dst_ip": "Destination IP",
  "network_flow.column.dst_port": "Destination port",
  "network_flow.column.ip_protocol": "Protocol",
  "network_flow.column.bytes_count": "Bytes",
  "network_flow.column.packets_count": "Packets",
  "network_flow.column.input_interface": "Input interface",
  "network_flow.column.output_interface": "Output interface",
  "network_flow.column.exporter_id": "Exporter",
  "network_flow.column.tcp_flags": "TCP flags",
  "network_flow.column.application_label": "Application",
  "network_flow.column.source_column_ordinal": "Source column",
  "network_flow.column.field_key": "Field",
  "network_flow.column.error_code": "Error",
  "network_flow.column.reason_code": "Reason",
  "network_flow.column.message": "Message",
  "network_flow.inspector.row_id": "Row ID",
  "network_flow.inspector.table_id": "Table ID",
  "network_flow.inspector.mapping_fingerprint": "Mapping fingerprint",
  "network_flow.inspector.observation_source": "Observation source",
  "network_flow.inspector.unmapped_raw": "Unmapped source values",
  "network_flow.inspector.diagnostic_id": "Diagnostic ID",
  "network_flow.inspector.message_key": "Message key",
  "network_flow.inspector.safe_sample": "Safe sample",
  "network_flow.inspector.raw_value_sha256": "Raw value digest",
};

const cellValueStyle = {
  display: "block",
  minWidth: 0,
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap",
} satisfies CSSProperties;

const visuallyHiddenStyle = {
  blockSize: 1,
  clip: "rect(0 0 0 0)",
  clipPath: "inset(50%)",
  inlineSize: 1,
  overflow: "hidden",
  position: "absolute",
  whiteSpace: "nowrap",
} satisfies CSSProperties;
