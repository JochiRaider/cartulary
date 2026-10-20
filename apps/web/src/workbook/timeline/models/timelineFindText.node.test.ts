import { describe, expect, it } from "vitest";
import { timelineFindText } from "./timelineFindText";
import { createDraftRow, rowFromApi } from "./timelineRowModel";

function row(values: Record<string, unknown>) {
  return rowFromApi({
    record_id: "record",
    row_version: 1,
    view_schema_id: "cartulary.view.timeline.v2",
    cells: Object.fromEntries(
      Object.entries(values).map(([key, value]) => [key, { value }]),
    ),
  });
}

describe("Timeline Find text", () => {
  it("reads saved scalar text despite pending authoring and excludes drafts placeholders and unknown metadata", () => {
    const saved = row({
      "timeline.activity_synopsis_text": "saved text",
      secret: "private metadata",
    });
    const pending = {
      ...saved,
      pendingSignature: "pending",
      values: { ...saved.values, activitySynopsisText: "local text" },
    };
    expect(
      timelineFindText(pending, "timeline.activity_synopsis_text", {}),
    ).toEqual(["saved text"]);
    expect(timelineFindText(pending, "timeline.raw_activity_text", {})).toEqual(
      [],
    );
    expect(timelineFindText(pending, "secret", {})).toEqual([]);
    expect(
      timelineFindText(
        row({
          "timeline.activity_synopsis_text": {
            items: [{ raw_text: "transport metadata" }],
          },
        }),
        "timeline.activity_synopsis_text",
        {},
      ),
    ).toEqual([]);
    expect(
      timelineFindText(
        createDraftRow(1),
        "timeline.activity_synopsis_text",
        {},
      ),
    ).toEqual([]);
  });
  it("searches only the committed displayed collection label without overflow or authoring", () => {
    const saved = row({
      "timeline.tags": {
        items: [
          {
            item_ref: "one",
            item_kind: "tag",
            raw_text: "first",
            display_text: "First label",
          },
          {
            item_ref: "two",
            item_kind: "tag",
            raw_text: "hidden",
            display_text: "Hidden label",
          },
        ],
      },
    });
    const pending = {
      ...saved,
      collectionDrafts: { ...saved.collectionDrafts, tags: "new tag" },
      collectionValues: { ...saved.collectionValues, tags: [] },
    };
    expect(timelineFindText(pending, "timeline.tags", {})).toEqual([
      "First label",
    ]);
    expect(
      timelineFindText(
        row({ "timeline.tags": { items: [] } }),
        "timeline.tags",
        {},
      ),
    ).toEqual([]);
  });
  it("uses formatted committed values as independent fragments and omits unavailable evidence fallbacks", () => {
    const saved = row({
      "timeline.evidence_count": 3,
      "timeline.has_evidence": true,
      "timeline.edited_at": "2026-09-17T12:00:00Z",
    });
    expect(timelineFindText(saved, "timeline.evidence_count", {})).toEqual([
      "3",
      "true",
    ]);
    expect(timelineFindText(saved, "timeline.edited_at", {})).toEqual([
      "2026-09-17T12:00:00Z",
    ]);
    expect(
      timelineFindText(
        row({ "timeline.evidence_count": -1 }),
        "timeline.evidence_count",
        {},
      ),
    ).toEqual([]);
  });
});
