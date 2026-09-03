import type {
  GridClipboardInput,
  GridPasteTargetResolution,
} from "@cartulary/grid-adapter";
import { describe, expect, it } from "vitest";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import {
  type TimelinePasteAuthority,
  timelinePastePlanAdmission,
  timelinePasteRequestTargetsMatchResolution,
  timelinePasteTargetPlansMatch,
} from "./timelineClipboardPastePlan";

const editable: TimelinePasteAuthority = {
  canCreateRows: true,
  editable: true,
  grouped: false,
};
const input: Extract<GridClipboardInput, { readonly kind: "table" }> = {
  format: "tsv",
  kind: "table",
  rawText: "Alpha\tSource\nBeta\tSource",
  values: [
    ["Alpha", "Source"],
    ["Beta", "Source"],
  ],
};
const resolution: GridPasteTargetResolution = {
  columns: ["timeline.activity_synopsis_text", "timeline.data_source_text"],
  rowTargets: [
    {
      kind: "record",
      mutationIdentity: { baseRowVersion: 2, kind: "core_row_version" },
      rowIdentity: { kind: "core_record", recordId: "record-1" },
      surface: { kind: "view_schema", viewSchemaId: timelineViewSchemaId },
    },
    {
      createIndex: 0,
      kind: "create",
      surface: { kind: "view_schema", viewSchemaId: timelineViewSchemaId },
    },
  ],
};

describe("timelineClipboardPastePlan", () => {
  it("admits exact editable mixed batches", () => {
    expect(timelinePastePlanAdmission(resolution, input, editable)).toEqual({
      kind: "accepted",
    });
  });

  it("fails closed for authority, grouping, field, shape, and surface drift", () => {
    expect(
      timelinePastePlanAdmission(resolution, input, {
        ...editable,
        editable: false,
      }),
    ).toEqual({ kind: "rejected", reason: "read_only" });
    expect(
      timelinePastePlanAdmission(resolution, input, {
        ...editable,
        grouped: true,
      }),
    ).toEqual({ kind: "rejected", reason: "grouped_create" });
    expect(
      timelinePastePlanAdmission(resolution, input, {
        ...editable,
        canCreateRows: false,
      }),
    ).toEqual({ kind: "rejected", reason: "create_unavailable" });
    expect(
      timelinePastePlanAdmission(
        {
          ...resolution,
          columns: ["timeline.tags", "timeline.data_source_text"],
        },
        input,
        editable,
      ),
    ).toEqual({ kind: "rejected", reason: "invalid_fields" });
    expect(
      timelinePastePlanAdmission(
        { ...resolution, rowTargets: resolution.rowTargets.slice(0, 1) },
        input,
        editable,
      ),
    ).toEqual({ kind: "rejected", reason: "invalid_shape" });
    const recordTarget = resolution.rowTargets[0];
    if (recordTarget === undefined) throw new Error("expected record target");
    expect(
      timelinePastePlanAdmission(
        {
          ...resolution,
          rowTargets: [
            recordTarget,
            {
              createIndex: 0,
              kind: "create",
              surface: {
                kind: "view_schema",
                viewSchemaId: "cartulary.view.hosts.v1",
              },
            },
          ],
        },
        input,
        editable,
      ),
    ).toEqual({ kind: "rejected", reason: "wrong_surface" });
  });

  it("matches only stable fields, surfaces, row identities, and create positions", () => {
    expect(timelinePasteTargetPlansMatch(resolution, resolution)).toBe(true);
    const createTarget = resolution.rowTargets[1];
    if (createTarget === undefined) throw new Error("expected create target");
    expect(
      timelinePasteTargetPlansMatch(resolution, {
        ...resolution,
        rowTargets: [
          {
            kind: "record",
            mutationIdentity: {
              baseRowVersion: 9,
              kind: "core_row_version",
            },
            rowIdentity: { kind: "core_record", recordId: "record-2" },
            surface: {
              kind: "view_schema",
              viewSchemaId: timelineViewSchemaId,
            },
          },
          createTarget,
        ],
      }),
    ).toBe(false);
    expect(
      timelinePasteTargetPlansMatch(resolution, {
        ...resolution,
        columns: [
          "timeline.data_source_text",
          "timeline.activity_synopsis_text",
        ],
      }),
    ).toBe(false);
  });

  it("matches dispatch targets to the final record identities and versions", () => {
    expect(
      timelinePasteRequestTargetsMatchResolution(
        [
          { base_row_version: 2, kind: "record", record_id: "record-1" },
          { kind: "create" },
        ],
        resolution,
      ),
    ).toBe(true);
    expect(
      timelinePasteRequestTargetsMatchResolution(
        [
          { base_row_version: 1, kind: "record", record_id: "record-1" },
          { kind: "create" },
        ],
        resolution,
      ),
    ).toBe(false);
  });
});
