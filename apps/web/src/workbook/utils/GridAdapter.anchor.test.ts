import type {
  GridCellAnchor,
  GridColumn,
  GridDataRow,
  GridDraftRow,
  GridHandle,
} from "@cartulary/grid-adapter";
import { SemanticDataGrid } from "@cartulary/grid-adapter/test-support";
import { cleanup, render } from "@testing-library/react";
import { createElement, createRef } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

type AnchorRow = {
  readonly label: string;
  readonly state: string | null | undefined;
};

function savedRow(
  recordId: string,
  state: string | null | undefined,
): GridDataRow<AnchorRow> {
  return {
    kind: "data",
    mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
    rowIdentity: { kind: "core_record", recordId },
    data: {
      label: recordId,
      state,
    },
  };
}

function draftRow(key: string): GridDraftRow<AnchorRow> {
  return {
    kind: "draft",
    data: {
      label: key,
      state: "draft",
    },
    testId: key,
  };
}

const columns: readonly GridColumn<AnchorRow>[] = [
  { fieldKey: "task.title", label: "Title", renderCell: () => null },
  { fieldKey: "task.status", label: "Status", renderCell: () => null },
  { fieldKey: "task.priority", label: "Priority", renderCell: () => null },
];

const surface = { kind: "view_schema", viewSchemaId: "test.view" } as const;

function anchor(recordId: string, fieldKey: string): GridCellAnchor {
  return {
    fieldKey,
    rowIdentity: { kind: "core_record", recordId },
    surface,
  };
}

function renderConsumerGrid(options?: {
  readonly draft?: GridDraftRow<AnchorRow> | undefined;
  readonly grouped?: boolean | undefined;
  readonly onActiveCellChange?:
    | ((anchor: GridCellAnchor | null) => void)
    | undefined;
  readonly rows?: readonly GridDataRow<AnchorRow>[] | undefined;
}) {
  const handle = createRef<GridHandle>();
  const rows = options?.rows ?? [
    savedRow("record-1", "open"),
    savedRow("record-2", "done"),
  ];
  render(
    createElement(SemanticDataGrid<AnchorRow>, {
      columns,
      dataRows: rows,
      draftRow: options?.draft,
      grouping:
        options?.grouped === true
          ? {
              fieldKey: "state",
              formatLabel: (value) => (value === null ? null : String(value)),
              getValue: (row) => row.state ?? null,
            }
          : null,
      onActiveCellChange: options?.onActiveCellChange,
      ref: handle,
      surface,
    }),
  );
  if (handle.current === null) {
    throw new Error("Expected the grid consumer handle to be bound");
  }
  return handle.current;
}

afterEach(cleanup);

describe("adapter-owned semantic handle evidence", () => {
  it("focuses stable record_id and field_key anchors through the consumer handle", () => {
    const onActiveCellChange = vi.fn();
    const handle = renderConsumerGrid({ onActiveCellChange });
    const target = anchor("record-2", "task.status");

    expect(handle.focusAnchor(target)).toBe(true);
    expect(onActiveCellChange).toHaveBeenLastCalledWith(target);
  });

  it("rejects invalid row, field, surface, and recordless targets", () => {
    const handle = renderConsumerGrid({
      draft: draftRow("draft-1"),
      grouped: true,
    });

    expect(handle.focusAnchor(anchor("missing", "task.status"))).toBe(false);
    expect(handle.focusAnchor(anchor("record-1", "missing"))).toBe(false);
    expect(handle.focusAnchor(anchor("", "task.title"))).toBe(false);
    expect(
      handle.focusAnchor({
        ...anchor("record-1", "task.title"),
        surface: { kind: "view_schema", viewSchemaId: "wrong.view" },
      }),
    ).toBe(false);
  });

  it("resolves Arrow, Tab, Enter, and Shift+Enter navigation through adapter anchors", () => {
    const handle = renderConsumerGrid({
      rows: [
        savedRow("record-1", "open"),
        savedRow("record-2", "open"),
        savedRow("record-3", "open"),
      ],
    });

    expect(
      handle.moveFocus(anchor("record-1", "task.title"), {
        key: "ArrowRight",
      }),
    ).toEqual(anchor("record-1", "task.status"));
    expect(
      handle.moveFocus(anchor("record-1", "task.status"), { key: "Tab" }),
    ).toEqual(anchor("record-1", "task.priority"));
    expect(
      handle.moveFocus(anchor("record-1", "task.status"), { key: "Enter" }),
    ).toEqual(anchor("record-2", "task.status"));
    expect(
      handle.moveFocus(anchor("record-2", "task.status"), {
        key: "Enter",
        shiftKey: true,
      }),
    ).toEqual(anchor("record-1", "task.status"));
  });

  it("exposes no vendor-selection resolver on the semantic handle", () => {
    const handle = renderConsumerGrid();

    expect(Object.keys(handle)).not.toContain("resolveVendorSelection");
    expect(handle.focusAnchor(anchor("record-2", "task.status"))).toBe(true);
  });
});
