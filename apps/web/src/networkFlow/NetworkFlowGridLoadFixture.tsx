import type { GridCellAnchor, GridCellRange } from "@cartulary/grid-adapter";
import { networkAnalysisTestId } from "@cartulary/ui-contracts";
import { type CSSProperties, useCallback, useMemo, useState } from "react";
import type {
  NetworkFlowContributor,
  NetworkFlowDiagnostic,
  NetworkFlowRow,
  NetworkFlowTable,
} from "../services/networkFlowContractAdapter";
import {
  NetworkFlowAcceptedGrid,
  NetworkFlowContributorGrid,
  NetworkFlowRejectedGrid,
} from "./NetworkFlowSemanticGrid";
import {
  reconcileNetworkFlowContributors,
  reconcileNetworkFlowDiagnostics,
  reconcileNetworkFlowRows,
} from "./networkFlowQueryModel";

type FixtureSurface = "accepted" | "contributors" | "rejected";

export function NetworkFlowGridLoadFixture() {
  const logicalRowCount = fixtureRowCount();
  const [surface, setSurface] = useState<FixtureSurface>("accepted");
  const [rows, setRows] = useState<readonly NetworkFlowRow[]>(() =>
    fixtureRows(logicalRowCount),
  );
  const [diagnostics, setDiagnostics] = useState<
    readonly NetworkFlowDiagnostic[]
  >(() => fixtureDiagnostics(logicalRowCount));
  const [contributors, setContributors] = useState<
    readonly NetworkFlowContributor[]
  >(() => fixtureContributors(logicalRowCount));
  const tables = useMemo(() => fixtureTables(), []);
  const [refreshCount, setRefreshCount] = useState(0);
  const [selectionSummary, setSelectionSummary] = useState("No selection");
  const handleSelectionChange = useCallback(
    (active: GridCellAnchor | null, range: GridCellRange | null) => {
      setSelectionSummary(
        active === null
          ? "No selection"
          : `${anchorSummary(active)}; range ${range === null ? "none" : `${anchorSummary(range.start)} to ${anchorSummary(range.end)}`}`,
      );
    },
    [],
  );

  const refreshEquivalentResources = () => {
    setRows((current) =>
      reconcileNetworkFlowRows(current, fixtureRows(logicalRowCount)),
    );
    setDiagnostics((current) =>
      reconcileNetworkFlowDiagnostics(
        current,
        fixtureDiagnostics(logicalRowCount),
      ),
    );
    setContributors((current) =>
      reconcileNetworkFlowContributors(
        current,
        fixtureContributors(logicalRowCount),
      ),
    );
    setRefreshCount((current) => current + 1);
  };

  return (
    <section
      aria-label="Network Flow supported-load fixture"
      data-logical-row-count={logicalRowCount}
      data-testid={networkAnalysisTestId("load-fixture")}
      style={fixtureStyle}
    >
      <header style={headerStyle}>
        <div>
          <strong>Network Flow supported-load fixture</strong>
          <div>
            {logicalRowCount.toLocaleString("en-US")} deterministic resources;
            refresh generation {refreshCount}
          </div>
        </div>
        <button type="button" onClick={refreshEquivalentResources}>
          Refresh equivalent resources
        </button>
        <output aria-label="Fixture semantic selection">
          {selectionSummary}
        </output>
      </header>
      <nav aria-label="Fixture grid surface" style={surfaceControlsStyle}>
        {(["accepted", "rejected", "contributors"] as const).map(
          (candidate) => (
            <button
              aria-pressed={surface === candidate}
              key={candidate}
              type="button"
              onClick={() => setSurface(candidate)}
            >
              {fixtureSurfaceLabel(candidate)}
            </button>
          ),
        )}
      </nav>
      <div style={gridHostStyle}>
        {surface === "accepted" ? (
          <NetworkFlowAcceptedGrid
            filtered={false}
            loadState="ready"
            resetKey="supported-load-fixture"
            rows={rows}
            sort={[]}
            onResetQuery={() => undefined}
            onRetry={() => undefined}
            onSelectionChange={handleSelectionChange}
            onSortChange={() => undefined}
          />
        ) : null}
        {surface === "rejected" ? (
          <NetworkFlowRejectedGrid
            diagnostics={diagnostics}
            filtered={false}
            loadState="ready"
            resetKey="supported-load-fixture"
            onResetQuery={() => undefined}
            onRetry={() => undefined}
          />
        ) : null}
        {surface === "contributors" ? (
          <NetworkFlowContributorGrid
            contributors={contributors}
            loadState="ready"
            tables={tables}
            onRetry={() => undefined}
          />
        ) : null}
      </div>
    </section>
  );
}

function anchorSummary(anchor: GridCellAnchor): string {
  const resourceID =
    anchor.rowIdentity.kind === "extension_resource"
      ? anchor.rowIdentity.resourceId
      : anchor.rowIdentity.recordId;
  return `${resourceID}:${anchor.fieldKey}`;
}

function fixtureRowCount(): number {
  const value = new URLSearchParams(window.location.search).get("fixture_rows");
  return value === "100" ? 100 : 1_000;
}

function fixtureRows(count: number): NetworkFlowRow[] {
  return Array.from({ length: count }, (_, index) => {
    const ordinal = index + 1;
    const suffix = String(ordinal).padStart(4, "0");
    return {
      mapping_fingerprint: "a".repeat(64),
      network_flow_row_id: `nfr_load_${suffix}`,
      network_flow_table_id: `nft_load_${(index % 4) + 1}`,
      source_row_number: ordinal,
      unmapped_raw: {},
      "network_flow.application_label": `application-${ordinal % 17}`,
      "network_flow.bytes_count": String(ordinal * 1_024),
      "network_flow.dst_ip": `203.0.113.${(index % 250) + 1}`,
      "network_flow.dst_port": 8_000 + (index % 100),
      "network_flow.exporter_id": `exporter-${index % 8}`,
      "network_flow.flow_end_utc": "2026-07-16T00:01:00Z",
      "network_flow.flow_start_utc": "2026-07-16T00:00:00Z",
      "network_flow.input_interface": `ingress-${index % 4}`,
      "network_flow.ip_protocol": 6,
      "network_flow.observation_source_ref": {
        import_unit_id: `niu_load_${(index % 4) + 1}`,
      },
      "network_flow.output_interface": `egress-${index % 4}`,
      "network_flow.packets_count": String(ordinal * 4),
      "network_flow.src_ip": `192.0.2.${(index % 250) + 1}`,
      "network_flow.src_port": 1_000 + (index % 1_000),
      "network_flow.tcp_flags": index % 256,
    } as NetworkFlowRow;
  });
}

function fixtureDiagnostics(count: number): NetworkFlowDiagnostic[] {
  return Array.from({ length: count }, (_, index) => {
    const ordinal = index + 1;
    const suffix = String(ordinal).padStart(4, "0");
    return {
      actual_value: null,
      diagnostic_id: `nfd_load_${suffix}`,
      error_code: "network_flow_invalid_ip",
      field_key: "network_flow.src_ip",
      limit_name: null,
      limit_value: null,
      message: "The value is not a valid IP address.",
      message_args: {},
      message_key:
        "network_flow.diagnostic.network_flow_invalid_ip.invalid_ipv4",
      raw_header_sha256: null,
      raw_value_sha256: null,
      reason_code: "invalid_ipv4",
      safe_sample: `invalid-ip-${ordinal}`,
      source_column_ordinal: 3,
      source_row_number: ordinal,
    } as NetworkFlowDiagnostic;
  });
}

function fixtureContributors(count: number): NetworkFlowContributor[] {
  return fixtureRows(count).map((row) => ({
    row,
    row_ref: {
      mapping_fingerprint: row.mapping_fingerprint,
      network_flow_row_id: row.network_flow_row_id,
      network_flow_table_id: row.network_flow_table_id,
      source_row_number: row.source_row_number,
    },
  }));
}

function fixtureTables(): NetworkFlowTable[] {
  return Array.from({ length: 4 }, (_, index) => ({
    display_name: `Load fixture table ${index + 1}`,
    network_flow_table_id: `nft_load_${index + 1}`,
  })) as NetworkFlowTable[];
}

function fixtureSurfaceLabel(surface: FixtureSurface): string {
  switch (surface) {
    case "accepted":
      return "Accepted rows";
    case "rejected":
      return "Rejected diagnostics";
    case "contributors":
      return "Graph contributors";
  }
}

const fixtureStyle = {
  blockSize: "min(48rem, calc(100vh - 2rem))",
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  gridTemplateRows: "auto auto minmax(0, 1fr)",
  inlineSize: "min(96rem, calc(100vw - 2rem))",
  minBlockSize: "36rem",
  minWidth: 0,
} satisfies CSSProperties;

const headerStyle = {
  alignItems: "center",
  display: "flex",
  gap: "var(--ct-spacing-md)",
  justifyContent: "space-between",
} satisfies CSSProperties;

const surfaceControlsStyle = {
  display: "flex",
  gap: "var(--ct-spacing-sm)",
} satisfies CSSProperties;

const gridHostStyle = {
  display: "grid",
  minBlockSize: 0,
  minWidth: 0,
} satisfies CSSProperties;
