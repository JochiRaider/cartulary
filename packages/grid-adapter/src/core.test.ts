import { describe, expect, it } from "vitest";

import {
  buildGridPresentationRows,
  type GridPresentationRow,
  type GridRow,
} from "./core";

type HarnessRow = {
  readonly label: string;
  readonly state: string | null | undefined;
};

function gridRow(
  key: string,
  state: string | null | undefined,
): GridRow<HarnessRow> {
  return {
    key,
    recordId: key,
    data: {
      label: key,
      state,
    },
  };
}

function summarizeRows(
  rows: readonly GridPresentationRow<HarnessRow>[],
): readonly string[] {
  return rows.map((row) =>
    row.kind === "group"
      ? `group:${row.groupLabel ?? "null"}:${row.key}:${row.testId ?? "no-test-id"}`
      : `data:${row.key}`,
  );
}

describe("grid presentation rows", () => {
  it("maps rows directly to data presentation rows without grouping", () => {
    const rows = [gridRow("record-1", "open"), gridRow("record-2", "closed")];

    expect(
      buildGridPresentationRows({
        groupBy: null,
        rows,
      }),
    ).toEqual([
      {
        gridRow: rows[0],
        key: "record-1",
        kind: "data",
      },
      {
        gridRow: rows[1],
        key: "record-2",
        kind: "data",
      },
    ]);

    expect(
      buildGridPresentationRows({
        groupBy: "state",
        rows,
      }),
    ).toEqual([
      {
        gridRow: rows[0],
        key: "record-1",
        kind: "data",
      },
      {
        gridRow: rows[1],
        key: "record-2",
        kind: "data",
      },
    ]);
  });

  it("inserts group rows only when consecutive normalized labels change", () => {
    const rows = [
      gridRow("record-1", "open"),
      gridRow("record-2", "open"),
      gridRow("record-3", "reviewed"),
      gridRow("record-4", "reviewed"),
      gridRow("record-5", "open"),
    ];

    const presentationRows = buildGridPresentationRows({
      getGroupLabel: (row, fieldKey) =>
        fieldKey === "state" ? row.state : null,
      getGroupRowTestId: (fieldKey, value) => `group-${fieldKey}-${value}`,
      groupBy: "state",
      rows,
    });

    expect(summarizeRows(presentationRows)).toEqual([
      "group:open:group:state:open:0:group-state-open",
      "data:record-1",
      "data:record-2",
      "group:reviewed:group:state:reviewed:0:group-state-reviewed",
      "data:record-3",
      "data:record-4",
      "group:open:group:state:open:1:group-state-open",
      "data:record-5",
    ]);
  });

  it("normalizes empty group labels to the unassigned group without test IDs", () => {
    const rows = [
      gridRow("record-1", null),
      gridRow("record-2", undefined),
      gridRow("record-3", "   "),
      gridRow("record-4", ""),
      gridRow("record-5", "closed"),
    ];

    const presentationRows = buildGridPresentationRows({
      getGroupLabel: (row) => row.state,
      getGroupRowTestId: (fieldKey, value) => `group-${fieldKey}-${value}`,
      groupBy: "state",
      rows,
    });

    expect(summarizeRows(presentationRows)).toEqual([
      "data:record-1",
      "data:record-2",
      "data:record-3",
      "data:record-4",
      "group:closed:group:state:closed:0:group-state-closed",
      "data:record-5",
    ]);
  });

  it("creates an unassigned group when empty labels follow a non-empty group", () => {
    const rows = [
      gridRow("record-1", "open"),
      gridRow("record-2", " "),
      gridRow("record-3", null),
    ];

    const presentationRows = buildGridPresentationRows({
      getGroupLabel: (row) => row.state,
      getGroupRowTestId: (fieldKey, value) => `group-${fieldKey}-${value}`,
      groupBy: "state",
      rows,
    });

    expect(summarizeRows(presentationRows)).toEqual([
      "group:open:group:state:open:0:group-state-open",
      "data:record-1",
      "group:null:group:state:empty:0:no-test-id",
      "data:record-2",
      "data:record-3",
    ]);
  });

  it("trims labels before comparing groups and generating keys", () => {
    const rows = [
      gridRow("record-1", " open "),
      gridRow("record-2", "open"),
      gridRow("record-3", " reviewed "),
    ];

    const presentationRows = buildGridPresentationRows({
      getGroupLabel: (row) => row.state,
      getGroupRowTestId: (fieldKey, value) => `group-${fieldKey}-${value}`,
      groupBy: "state",
      rows,
    });

    expect(summarizeRows(presentationRows)).toEqual([
      "group:open:group:state:open:0:group-state-open",
      "data:record-1",
      "data:record-2",
      "group:reviewed:group:state:reviewed:0:group-state-reviewed",
      "data:record-3",
    ]);
  });
});
