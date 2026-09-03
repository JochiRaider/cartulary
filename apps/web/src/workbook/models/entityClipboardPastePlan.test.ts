import type {
  GridCellPasteIntent,
  GridPasteTargetResolution,
} from "@cartulary/grid-adapter";
import { describe, expect, it } from "vitest";
import { entityClipboardPastePlan } from "./entityClipboardPastePlan";

const surface = "cartulary.view.hosts.v1";
const resolution: GridPasteTargetResolution = {
  columns: ["host.display_name", "host.hostname"],
  rowTargets: [
    {
      kind: "record",
      mutationIdentity: { baseRowVersion: 3, kind: "core_row_version" },
      rowIdentity: { kind: "core_record", recordId: "host-1" },
      surface: { kind: "view_schema", viewSchemaId: surface },
    },
  ],
};

function intent(
  input: GridCellPasteIntent["input"],
  targetResolution: GridPasteTargetResolution = resolution,
): GridCellPasteIntent {
  return {
    input,
    range: {
      end: {
        fieldKey: "host.hostname",
        rowIdentity: { kind: "core_record", recordId: "host-1" },
        surface: { kind: "view_schema", viewSchemaId: surface },
      },
      start: {
        fieldKey: "host.display_name",
        rowIdentity: { kind: "core_record", recordId: "host-1" },
        surface: { kind: "view_schema", viewSchemaId: surface },
      },
    },
    target: {
      fieldKey: "host.display_name",
      mutationIdentity: { baseRowVersion: 3, kind: "core_row_version" },
      rowIdentity: { kind: "core_record", recordId: "host-1" },
      surface: { kind: "view_schema", viewSchemaId: surface },
    },
    targetResolution,
  };
}

const authority = {
  canCreateRows: true,
  grouped: false,
  rows: [{ recordId: "host-1", rowVersion: 3 }],
  viewSchemaId: surface,
  writableFieldKeys: new Set(["host.display_name", "host.hostname"]),
};

describe("entity clipboard paste plan", () => {
  it("keeps scalar paste on the existing-record mutation path", () => {
    expect(
      entityClipboardPastePlan(
        intent({ kind: "scalar", rawText: "alpha,beta", value: "alpha,beta" }),
        authority,
      ),
    ).toEqual({
      fieldKey: "host.display_name",
      kind: "scalar",
      target: { baseRowVersion: 3, recordId: "host-1" },
      value: "alpha,beta",
    });
  });

  it("builds an all-create batch while proving source targets are current", () => {
    const plan = entityClipboardPastePlan(
      intent({
        format: "tsv",
        kind: "table",
        rawText: "alpha\thost-a",
        values: [["alpha", "host-a"]],
      }),
      authority,
    );
    expect(plan).toEqual({
      input: {
        clipboard_text: "alpha\thost-a",
        columns: ["host.display_name", "host.hostname"],
        format: "tsv",
        start_field_key: "host.display_name",
        targets: [{ kind: "create" }],
        view_schema_id: surface,
      },
      kind: "batch",
    });
  });

  it("fails closed for stale targets, grouping, and unavailable fields", () => {
    const tableIntent = intent({
      format: "csv",
      kind: "table",
      rawText: "alpha,host-a",
      values: [["alpha", "host-a"]],
    });
    expect(
      entityClipboardPastePlan(tableIntent, {
        ...authority,
        rows: [{ recordId: "host-1", rowVersion: 4 }],
      }),
    ).toMatchObject({ kind: "rejected" });
    expect(
      entityClipboardPastePlan(tableIntent, { ...authority, grouped: true }),
    ).toMatchObject({ kind: "rejected" });
    expect(
      entityClipboardPastePlan(tableIntent, {
        ...authority,
        writableFieldKeys: new Set(["host.display_name"]),
      }),
    ).toMatchObject({ kind: "rejected" });
  });
});
