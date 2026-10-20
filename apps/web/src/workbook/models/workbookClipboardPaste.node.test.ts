import type { GridPasteTargetResolution } from "@cartulary/grid-adapter";
import { describe, expect, it } from "vitest";
import {
  workbookPasteColumns,
  workbookPasteResolutionMatchesSurface,
  workbookPasteTargets,
  workbookPasteViewSchemaId,
} from "./workbookClipboardPaste";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  timelineViewSchemaId,
} from "./workbookSurfaceRegistry";

describe("workbookClipboardPaste", () => {
  it("constructs only bounded non-empty generated columns and targets", () => {
    expect(workbookPasteColumns([])).toBeNull();
    expect(
      workbookPasteColumns(["timeline.date_entered_text", " "]),
    ).toBeNull();
    expect(workbookPasteColumns(["timeline.date_entered_text"])).toEqual([
      "timeline.date_entered_text",
    ]);
    expect(workbookPasteTargets([])).toBeNull();
    expect(
      workbookPasteTargets([
        { base_row_version: 0, kind: "record", record_id: "record-1" },
      ]),
    ).toBeNull();
    expect(
      workbookPasteTargets([
        { base_row_version: 3, kind: "record", record_id: "record-1" },
        { kind: "create" },
      ]),
    ).toEqual([
      { base_row_version: 3, kind: "record", record_id: "record-1" },
      { kind: "create" },
    ]);
  });

  it("parses exactly the generated paste-capable view identities", () => {
    expect(workbookPasteViewSchemaId(timelineViewSchemaId)).toBe(
      timelineViewSchemaId,
    );
    expect(workbookPasteViewSchemaId(hostsViewSchemaId)).toBe(
      hostsViewSchemaId,
    );
    expect(workbookPasteViewSchemaId(identitiesViewSchemaId)).toBe(
      identitiesViewSchemaId,
    );
    expect(workbookPasteViewSchemaId("cartulary.view.evidence.v1")).toBeNull();
  });

  it("rejects any target planned for another surface", () => {
    const resolution: GridPasteTargetResolution = {
      columns: ["timeline.activity_synopsis_text"],
      rowTargets: [
        {
          kind: "record",
          mutationIdentity: {
            baseRowVersion: 2,
            kind: "core_row_version",
          },
          rowIdentity: { kind: "core_record", recordId: "record-1" },
          surface: { kind: "view_schema", viewSchemaId: timelineViewSchemaId },
        },
      ],
    };
    expect(
      workbookPasteResolutionMatchesSurface(resolution, timelineViewSchemaId),
    ).toBe(true);
    expect(
      workbookPasteResolutionMatchesSurface(
        {
          ...resolution,
          rowTargets: [
            {
              createIndex: 0,
              kind: "create",
              surface: { kind: "view_schema", viewSchemaId: hostsViewSchemaId },
            },
          ],
        },
        timelineViewSchemaId,
      ),
    ).toBe(false);
  });
});
