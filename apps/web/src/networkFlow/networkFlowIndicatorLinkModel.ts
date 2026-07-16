import type { GridCellAnchor, GridCellRange } from "@cartulary/grid-adapter";
import type {
  NetworkFlowIndicatorSelector,
  NetworkFlowRow,
  NetworkFlowRowRef,
} from "./networkFlowClient";
import type { NetworkFlowIndicatorLinkCandidate } from "./useNetworkFlowIndicatorLinkController";

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
    candidateValue.trim() === "" ||
    selectedRows.some((row) => row[fieldKey] !== candidateValue)
  ) {
    return null;
  }
  return { candidateValue, fieldKey, rows: selectedRows };
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
    key: `${selector.kind}:${selection.fieldKey}:${selection.rows
      .map((row) => row.network_flow_row_id)
      .join(",")}:${selection.candidateValue}`,
    label:
      selection.rows.length === 1
        ? `Selected ${selection.fieldKey}`
        : `${selection.rows.length} selected rows`,
    selector,
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
