import type { GridCellAnchor, GridCellRange } from "@cartulary/grid-adapter";
import { coreAtomicIPType } from "../services/networkFlowIndicatorAdapter";
import type {
  NetworkFlowGraphEdge,
  NetworkFlowGraphResult,
  NetworkFlowGraphVertex,
  NetworkFlowIndicatorSelector,
  NetworkFlowRow,
  NetworkFlowRowRef,
} from "./networkFlowClient";
import type { NetworkFlowIndicatorLinkCandidate } from "./networkFlowIndicatorLinkOperation";

export type NetworkFlowLinkableFieldKey =
  | "network_flow.src_ip"
  | "network_flow.dst_ip";

export type NetworkFlowRowLinkSelection = {
  readonly candidateValue: string;
  readonly fieldKey: NetworkFlowLinkableFieldKey;
  readonly rows: readonly NetworkFlowRow[];
};

export function resolveNetworkFlowRowLinkSelection(options: {
  readonly activeAnchor: GridCellAnchor | null;
  readonly bindingSourceRowLimit: number;
  readonly cellRange: GridCellRange | null;
  readonly rows: readonly NetworkFlowRow[];
}): NetworkFlowRowLinkSelection | null {
  if (options.bindingSourceRowLimit < 1) {
    return null;
  }
  const fieldKey = linkableFieldKey(options.activeAnchor?.fieldKey);
  if (fieldKey === null) {
    return null;
  }
  const applicableRange =
    options.cellRange !== null &&
    rangeContainsAnchor(options.rows, options.cellRange, options.activeAnchor)
      ? options.cellRange
      : null;
  const selectedRows =
    applicableRange === null
      ? rowForAnchor(options.rows, options.activeAnchor)
      : rowsForRange(options.rows, applicableRange, fieldKey);
  if (
    selectedRows.length < 1 ||
    selectedRows.length > options.bindingSourceRowLimit
  ) {
    return null;
  }
  const candidateValue = String(selectedRows[0]?.[fieldKey] ?? "");
  if (
    coreAtomicIPType(candidateValue) === null ||
    selectedRows.some((row) => row[fieldKey] !== candidateValue)
  ) {
    return null;
  }
  return { candidateValue, fieldKey, rows: selectedRows };
}

export function networkFlowVertexLinkCandidate(
  graph: NetworkFlowGraphResult | null,
  vertex: NetworkFlowGraphVertex | null,
): NetworkFlowIndicatorLinkCandidate | null {
  if (graph === null || vertex === null) return null;
  const bindings = graph.vertex_selectors.filter(
    (binding) => binding.projected_vertex_id === vertex.vertex_id,
  );
  const semantic = bindings[0]?.selector;
  if (
    bindings.length !== 1 ||
    semantic === undefined ||
    !/^nfe_[a-f0-9]{64}$/u.test(semantic.source_vertex_id) ||
    coreAtomicIPType(semantic.endpoint_value) === null
  )
    return null;
  const selector: NetworkFlowIndicatorSelector = {
    kind: "graph_vertex",
    graph_query: graph.semantic_query,
    graph_query_digest: graph.graph_query_digest,
    vertex_id: semantic.source_vertex_id,
  };
  return {
    candidateValue: semantic.endpoint_value,
    key: JSON.stringify([
      selector,
      semantic.endpoint_value,
      graph.source_table_refs,
    ]),
    label: "Selected graph endpoint",
    selector,
    sourceRefs: [],
    sourceTableIds: graph.semantic_query.selected_table_ids,
    sourceTableRefs: graph.source_table_refs,
  };
}

export function networkFlowEdgeLinkCandidate(options: {
  readonly edge: NetworkFlowGraphEdge | null;
  readonly graph: NetworkFlowGraphResult | null;
  readonly fieldKey: NetworkFlowLinkableFieldKey;
}): NetworkFlowIndicatorLinkCandidate | null {
  const { edge, graph, fieldKey } = options;
  if (edge === null || graph === null) return null;
  const annotations = graph.edge_annotations.filter(
    (annotation) => annotation.projected_edge_id === edge.edge_id,
  );
  const semantic = annotations[0]?.selector;
  if (
    annotations.length !== 1 ||
    semantic?.kind !== "default_edge" ||
    !/^nff_[a-f0-9]{64}$/u.test(semantic.source_edge_id)
  )
    return null;
  const value =
    fieldKey === "network_flow.src_ip"
      ? semantic.source_endpoint_value
      : semantic.destination_endpoint_value;
  if (coreAtomicIPType(value) === null) return null;
  const selector: NetworkFlowIndicatorSelector = {
    kind: "graph_edge",
    graph_query: graph.semantic_query,
    graph_query_digest: graph.graph_query_digest,
    edge_id: semantic.source_edge_id,
    field_key: fieldKey,
  };
  return {
    candidateValue: value,
    key: JSON.stringify([selector, value, graph.source_table_refs]),
    label:
      fieldKey === "network_flow.src_ip"
        ? "Selected edge source endpoint"
        : "Selected edge destination endpoint",
    selector,
    sourceRefs: [],
    sourceTableIds: graph.semantic_query.selected_table_ids,
    sourceTableRefs: graph.source_table_refs,
  };
}

export function networkFlowRowLinkCandidate(
  selection: NetworkFlowRowLinkSelection,
): NetworkFlowIndicatorLinkCandidate {
  const selector: NetworkFlowIndicatorSelector =
    selection.rows.length === 1
      ? {
          kind: "row_field_value",
          network_flow_table_id: selection.rows[0]?.network_flow_table_id ?? "",
          network_flow_row_id: selection.rows[0]?.network_flow_row_id ?? "",
          field_key: selection.fieldKey,
        }
      : {
          kind: "row_refs",
          row_refs: rowRefs(selection.rows),
          field_key: selection.fieldKey,
        };
  return {
    candidateValue: selection.candidateValue,
    key: JSON.stringify([
      selector,
      rowRefs(selection.rows),
      selection.candidateValue,
    ]),
    label:
      selection.rows.length === 1
        ? selection.fieldKey === "network_flow.src_ip"
          ? "Selected source IP"
          : "Selected destination IP"
        : `${selection.rows.length} selected rows · ${selection.fieldKey === "network_flow.src_ip" ? "Source IP" : "Destination IP"}`,
    selector,
    sourceRefs: rowRefs(selection.rows),
    sourceTableRefs: [],
    sourceTableIds: [
      ...new Set(selection.rows.map((row) => row.network_flow_table_id)),
    ],
  };
}

function rowsForRange(
  rows: readonly NetworkFlowRow[],
  range: GridCellRange,
  fieldKey: NetworkFlowLinkableFieldKey,
): readonly NetworkFlowRow[] {
  if (range.start.fieldKey !== fieldKey || range.end.fieldKey !== fieldKey) {
    return [];
  }
  const startIndex = rowIndexForAnchor(rows, range.start);
  const endIndex = rowIndexForAnchor(rows, range.end);
  if (startIndex < 0 || endIndex < 0) {
    return [];
  }
  return rows.slice(
    Math.min(startIndex, endIndex),
    Math.max(startIndex, endIndex) + 1,
  );
}

function rangeContainsAnchor(
  rows: readonly NetworkFlowRow[],
  range: GridCellRange,
  anchor: GridCellAnchor | null,
): boolean {
  if (anchor === null) {
    return false;
  }
  const startIndex = rowIndexForAnchor(rows, range.start);
  const endIndex = rowIndexForAnchor(rows, range.end);
  const anchorIndex = rowIndexForAnchor(rows, anchor);
  return (
    startIndex >= 0 &&
    endIndex >= 0 &&
    anchorIndex >= Math.min(startIndex, endIndex) &&
    anchorIndex <= Math.max(startIndex, endIndex)
  );
}

function rowForAnchor(
  rows: readonly NetworkFlowRow[],
  anchor: GridCellAnchor | null,
): readonly NetworkFlowRow[] {
  const index = rowIndexForAnchor(rows, anchor);
  return index < 0 ? [] : [rows[index] as NetworkFlowRow];
}

function rowIndexForAnchor(
  rows: readonly NetworkFlowRow[],
  anchor: GridCellAnchor | null,
): number {
  if (anchor === null || !isNetworkFlowRowAnchor(anchor)) {
    return -1;
  }
  const resourceId = anchor.rowIdentity.resourceId;
  return rows.findIndex((row) => row.network_flow_row_id === resourceId);
}

function isNetworkFlowRowAnchor(
  anchor: GridCellAnchor,
): anchor is GridCellAnchor & {
  readonly rowIdentity: {
    readonly kind: "extension_resource";
    readonly extensionProfileId: string;
    readonly resourceKind: string;
    readonly resourceId: string;
  };
} {
  return (
    anchor.surface.kind === "extension_grid" &&
    anchor.surface.extensionProfileId === "network_flow_activity" &&
    anchor.surface.workspaceKey === "network_analysis" &&
    anchor.surface.gridSchemaId === "network_flow.accepted_rows.v1" &&
    anchor.rowIdentity.kind === "extension_resource" &&
    anchor.rowIdentity.extensionProfileId === "network_flow_activity" &&
    anchor.rowIdentity.resourceKind === "network_flow_row"
  );
}

function linkableFieldKey(
  value: string | undefined,
): NetworkFlowLinkableFieldKey | null {
  return value === "network_flow.src_ip" || value === "network_flow.dst_ip"
    ? value
    : null;
}

function rowRefs(
  rows: readonly NetworkFlowRow[],
): [NetworkFlowRowRef, ...NetworkFlowRowRef[]] {
  const [first, ...remaining] = rows.map(
    (row): NetworkFlowRowRef => ({
      network_flow_table_id: row.network_flow_table_id,
      network_flow_row_id: row.network_flow_row_id,
      source_row_number: row.source_row_number,
      mapping_fingerprint: row.mapping_fingerprint,
    }),
  );
  if (first === undefined) {
    throw new Error(
      "A Network Flow row-ref selector requires at least one row.",
    );
  }
  return [first, ...remaining];
}
