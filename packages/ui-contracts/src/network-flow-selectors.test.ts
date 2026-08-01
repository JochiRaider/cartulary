import { describe, expect, it } from "vitest";

import {
  networkAnalysisColumnActionTestId,
  networkAnalysisDiagnosticCellTestId,
  networkAnalysisDiagnosticTestId,
  networkAnalysisEdgeTestId,
  networkAnalysisMappingColumnTestId,
  networkAnalysisRowCellTestId,
  networkAnalysisRowTestId,
  networkAnalysisTableTabTestId,
  networkAnalysisTestId,
  networkAnalysisVertexTestId,
} from "./index";

const networkAnalysisSelectors = [
  "accepted-grid",
  "accepted-query-apply",
  "accepted-query-clear",
  "column-menu",
  "contributor-close",
  "contributor-grid",
  "contributor-drawer",
  "diagnostics-summary",
  "delete-cancel",
  "delete-confirm",
  "delete-confirmation",
  "delete-dialog",
  "delete-trigger",
  "filters",
  "graph-panel",
  "graph-live-region",
  "graph-scope",
  "import-input",
  "import-trigger",
  "inspector",
  "inspector-close",
  "indicator-link-cancel",
  "indicator-link-confirmation",
  "indicator-link-dialog",
  "indicator-link-existing-id",
  "indicator-link-submit",
  "layout-reset",
  "load-fixture",
  "mapping-apply",
  "mapping-dialog",
  "mapping-display-name",
  "mapping-preview",
  "mapping-preview-summary",
  "mapping-profile",
  "mapping-timestamp-mode",
  "mapping-timezone",
  "mapping-unknown-policy",
  "mode-graph",
  "mode-rejected",
  "mode-rows",
  "page-next",
  "page-previous",
  "page-status",
  "refresh",
  "rename-cancel",
  "rename-dialog",
  "rename-input",
  "rename-submit",
  "rename-trigger",
  "rejected-grid",
  "rejected-query-apply",
  "rejected-query-clear",
  "stale-state",
  "status-strip",
  "table-panel",
  "tab",
  "workspace",
  "workspace-header",
] as const;

describe("@cartulary/ui-contracts Network Flow selectors", () => {
  it("preserves every closed Network Analysis selector in exact order", () => {
    expect(networkAnalysisSelectors.map(networkAnalysisTestId)).toEqual(
      networkAnalysisSelectors.map(
        (selector) => `network-flow-analysis-${selector}`,
      ),
    );
  });

  it("encodes stable Network Flow table, graph, row, and diagnostic identities", () => {
    expect(networkAnalysisTableTabTestId("table/1")).toBe(
      "network-flow-table-tab-table%2F1",
    );
    expect(networkAnalysisEdgeTestId("edge:1")).toBe(
      "network-flow-edge-edge%3A1",
    );
    expect(networkAnalysisVertexTestId("vertex 1")).toBe(
      "network-flow-vertex-vertex%201",
    );
    expect(networkAnalysisRowTestId("row/1")).toBe("network-flow-row-row%2F1");
    expect(networkAnalysisRowCellTestId("row/1", "source.address")).toBe(
      "network-flow-row-cell-row%2F1-source.address",
    );
    expect(networkAnalysisDiagnosticTestId("diag:1")).toBe(
      "network-flow-diagnostic-diag%3A1",
    );
    expect(
      networkAnalysisDiagnosticCellTestId("diag:1", "diagnostic.message"),
    ).toBe("network-flow-diagnostic-cell-diag%3A1-diagnostic.message");
  });

  it("builds ordered column actions and mapping-column identities", () => {
    expect(
      (["move-earlier", "move-later", "toggle"] as const).map((action) =>
        networkAnalysisColumnActionTestId("source.address", action),
      ),
    ).toEqual([
      "network-flow-column-source.address-move-earlier",
      "network-flow-column-source.address-move-later",
      "network-flow-column-source.address-toggle",
    ]);
    expect([1, 2, 12].map(networkAnalysisMappingColumnTestId)).toEqual([
      "network-flow-mapping-column-1",
      "network-flow-mapping-column-2",
      "network-flow-mapping-column-12",
    ]);
  });

  it("fails closed for malformed Network Flow identity and ordinal inputs", () => {
    expect(() => networkAnalysisTableTabTestId("")).toThrow(
      "Invalid table_id selector token: ",
    );
    expect(() => networkAnalysisEdgeTestId(" ")).toThrow(
      "Invalid edge_id selector token:  ",
    );
    expect(networkAnalysisRowCellTestId("row-1", "Source address")).toBe(
      "network-flow-row-cell-row-1-Source%20address",
    );
    expect(() => networkAnalysisColumnActionTestId("", "toggle")).toThrow(
      "Invalid field_key selector token: ",
    );
    for (const ordinal of [0, -1, 1.5, Number.MAX_SAFE_INTEGER + 1]) {
      expect(() => networkAnalysisMappingColumnTestId(ordinal)).toThrow(
        "source_column_ordinal must be a positive safe integer",
      );
    }
  });
});
