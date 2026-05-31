// @vitest-environment jsdom

import { gridShellTestId, rowCellTestId } from "@cartulary/ui-contracts";
import { describe, expect, it } from "vitest";

import { findCommittedRowSummaryInRoot } from "./helpers";

const timelineViewSchemaId = "cartulary.view.timeline.v1";

describe("committed row measurement predicates", () => {
  it("matches only visible committed rows with stable record and row version", () => {
    document.body.innerHTML = `
      <div data-testid="${gridShellTestId(timelineViewSchemaId)}">
        <div role="row" data-grid-record-id="">
          <input data-testid="draft-timeline.summary" value="Timing sample" />
        </div>
        <div role="row" data-grid-record-id="record-1">
          <input data-testid="${rowCellTestId("record-1", "timeline.summary")}" value="Timing sample" />
          <span data-testid="${rowCellTestId("record-1", "row_version")}">2</span>
        </div>
      </div>
    `;

    expect(
      findCommittedRowSummaryInRoot(document, {
        expectedSummary: "Timing sample",
        surface: timelineViewSchemaId,
      }),
    ).toEqual({ recordId: "record-1", rowVersion: 2 });
  });

  it("rejects draft rows, missing versions, and mismatched summaries", () => {
    document.body.innerHTML = `
      <div data-testid="${gridShellTestId(timelineViewSchemaId)}">
        <div role="row" data-grid-record-id="">
          <input data-testid="draft-timeline.summary" value="Timing sample" />
        </div>
        <div role="row" data-grid-record-id="record-2">
          <input data-testid="${rowCellTestId("record-2", "timeline.summary")}" value="Other sample" />
          <span data-testid="${rowCellTestId("record-2", "row_version")}">1</span>
        </div>
        <div role="row" data-grid-record-id="record-3">
          <input data-testid="${rowCellTestId("record-3", "timeline.summary")}" value="Timing sample" />
          <span data-testid="${rowCellTestId("record-3", "row_version")}">new</span>
        </div>
      </div>
    `;

    expect(
      findCommittedRowSummaryInRoot(document, {
        expectedSummary: "Timing sample",
        surface: timelineViewSchemaId,
      }),
    ).toBeNull();
  });
});
