import type { StableTestId } from "./selectorCore";
import {
  encodedTestId,
  encodeSelectorSegment,
  stableTestId,
} from "./selectorCore";

export type NetworkAnalysisSelector =
  | "accepted-grid"
  | "accepted-query-apply"
  | "accepted-query-clear"
  | "column-menu"
  | "contributor-close"
  | "contributor-grid"
  | "contributor-drawer"
  | "diagnostics-summary"
  | "delete-cancel"
  | "delete-confirm"
  | "delete-confirmation"
  | "delete-dialog"
  | "delete-trigger"
  | "filters"
  | "advanced-filters"
  | "graph-panel"
  | "graph-live-region"
  | "graph-scope"
  | "graph-surface-explore"
  | "graph-surface-saved"
  | "import-input"
  | "import-trigger"
  | "inspector"
  | "inspector-close"
  | "indicator-link-cancel"
  | "indicator-link-confirmation"
  | "indicator-link-dialog"
  | "indicator-link-existing-id"
  | "indicator-link-submit"
  | "layout-reset"
  | "load-fixture"
  | "mapping-apply"
  | "mapping-dialog"
  | "mapping-display-name"
  | "mapping-preview"
  | "mapping-preview-summary"
  | "mapping-profile"
  | "mapping-timestamp-mode"
  | "mapping-timezone"
  | "mapping-unknown-policy"
  | "mode-graph"
  | "mode-rejected"
  | "mode-rows"
  | "page-next"
  | "page-previous"
  | "page-status"
  | "refresh"
  | "rename-cancel"
  | "rename-dialog"
  | "rename-input"
  | "rename-submit"
  | "rename-trigger"
  | "rejected-grid"
  | "rejected-query-apply"
  | "rejected-query-clear"
  | "saved-graph-contributors"
  | "saved-graph-create"
  | "saved-graph-dialog"
  | "saved-graph-heading"
  | "saved-graph-name"
  | "saved-graph-result"
  | "saved-graphs"
  | "stale-state"
  | "status-strip"
  | "table-panel"
  | "tab"
  | "workspace"
  | "workspace-header";

export function networkAnalysisTestId(
  selector: NetworkAnalysisSelector,
): StableTestId {
  return stableTestId(`network-flow-analysis-${selector}`);
}

export function networkAnalysisTableTabTestId(
  networkFlowTableId: string,
): StableTestId {
  return stableTestId(
    encodedTestId("network-flow-table-tab", networkFlowTableId, "table_id"),
  );
}

export function networkAnalysisEdgeTestId(edgeId: string): StableTestId {
  return stableTestId(encodedTestId("network-flow-edge", edgeId, "edge_id"));
}

export function networkAnalysisVertexTestId(vertexId: string): StableTestId {
  return stableTestId(
    encodedTestId("network-flow-vertex", vertexId, "vertex_id"),
  );
}

export function networkAnalysisSavedGraphEdgeTestId(
  edgeId: string,
): StableTestId {
  return stableTestId(
    encodedTestId("network-flow-saved-graph-edge", edgeId, "edge_id"),
  );
}

export function networkAnalysisSavedGraphVertexTestId(
  vertexId: string,
): StableTestId {
  return stableTestId(
    encodedTestId("network-flow-saved-graph-vertex", vertexId, "vertex_id"),
  );
}

export function networkAnalysisSavedGraphTestId(
  graphViewId: string,
): StableTestId {
  return stableTestId(
    encodedTestId("network-flow-saved-graph", graphViewId, "graph_view_id"),
  );
}

export function networkAnalysisRowTestId(rowId: string): StableTestId {
  return stableTestId(encodedTestId("network-flow-row", rowId, "row_id"));
}

export function networkAnalysisRowCellTestId(
  rowId: string,
  fieldKey: string,
): StableTestId {
  return stableTestId(
    `${encodedTestId("network-flow-row-cell", rowId, "row_id")}-${encodeSelectorSegment(fieldKey, "field_key")}`,
  );
}

export function networkAnalysisDiagnosticTestId(
  diagnosticId: string,
): StableTestId {
  return stableTestId(
    encodedTestId("network-flow-diagnostic", diagnosticId, "diagnostic_id"),
  );
}

export function networkAnalysisDiagnosticCellTestId(
  diagnosticId: string,
  fieldKey: string,
): StableTestId {
  return stableTestId(
    `${encodedTestId("network-flow-diagnostic-cell", diagnosticId, "diagnostic_id")}-${encodeSelectorSegment(fieldKey, "field_key")}`,
  );
}

export function networkAnalysisColumnActionTestId(
  fieldKey: string,
  action: "move-earlier" | "move-later" | "toggle",
): StableTestId {
  return stableTestId(
    `network-flow-column-${encodeSelectorSegment(fieldKey, "field_key")}-${action}`,
  );
}

export function networkAnalysisMappingColumnTestId(
  sourceColumnOrdinal: number,
): StableTestId {
  if (!Number.isSafeInteger(sourceColumnOrdinal) || sourceColumnOrdinal < 1) {
    throw new Error("source_column_ordinal must be a positive safe integer");
  }
  return stableTestId(`network-flow-mapping-column-${sourceColumnOrdinal}`);
}
