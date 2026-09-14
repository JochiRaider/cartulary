// @vitest-environment jsdom
import { gridShellTestId, rowCellTestId } from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import { findCommittedRowSummaryInRoot } from "../../measurement/timingSupport";

describe("committed row measurement predicates", () => {
  it("matches only visible committed rows with stable record and row version", () => {
    document.body.innerHTML = `
      <div data-testid="${gridShellTestId(timelineViewSchemaId)}">
        <div role="row" data-grid-record-id="">
          <input data-testid="draft-timeline.activity_synopsis_text" value="Timing sample" />
        </div>
        <div role="row" data-grid-record-id="record-1" data-grid-row-version="2">
          <input data-grid-field-key="timeline.activity_synopsis_text" data-testid="${rowCellTestId("record-1", "timeline.activity_synopsis_text")}" value="Timing sample" />
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
          <input data-testid="draft-timeline.activity_synopsis_text" value="Timing sample" />
        </div>
        <div role="row" data-grid-record-id="record-2" data-grid-row-version="1">
          <input data-grid-field-key="timeline.activity_synopsis_text" data-testid="${rowCellTestId("record-2", "timeline.activity_synopsis_text")}" value="Other sample" />
        </div>
        <div role="row" data-grid-record-id="record-3">
          <input data-grid-field-key="timeline.activity_synopsis_text" data-testid="${rowCellTestId("record-3", "timeline.activity_synopsis_text")}" value="Timing sample" />
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
  it("qualifies identity, committed version and visible field on one mounted row", () => {
    const grid = document.createElement("div");
    grid.dataset.testid = gridShellTestId(timelineViewSchemaId);
    const row = document.createElement("div");
    row.role = "row";
    row.dataset.gridRecordId = "expected";
    row.dataset.gridRowVersion = "2";
    const field = document.createElement("input");
    field.dataset.gridFieldKey = "timeline.activity_synopsis_text";
    field.value = "Sample";
    row.append(field);
    grid.append(row);
    document.body.replaceChildren(grid);
    const read = () =>
      findCommittedRowSummaryInRoot(document, {
        expectedSummary: "Sample",
        recordId: "expected",
        minimumRowVersion: 2,
        surface: timelineViewSchemaId,
      });
    expect(read()).toEqual({ recordId: "expected", rowVersion: 2 });
    row.dataset.gridRecordId = "wrong";
    expect(read()).toBeNull();
    row.dataset.gridRecordId = "expected";
    row.dataset.gridRowVersion = "1";
    expect(read()).toBeNull();
    row.dataset.gridRowVersion = "2";
    field.hidden = true;
    expect(read()).toBeNull();
    field.hidden = false;
    field.style.visibility = "hidden";
    expect(read()).toBeNull();
    field.style.visibility = "visible";
    field.dataset.gridFieldKey = "timeline.raw_activity_text";
    expect(read()).toBeNull();
    field.dataset.gridFieldKey = "timeline.activity_synopsis_text";
    const replacement = row.cloneNode(true) as HTMLElement;
    replacement.removeAttribute("data-grid-row-version");
    row.replaceWith(replacement);
    expect(read()).toBeNull();
    replacement.dataset.gridRowVersion = "3";
    expect(read()).toEqual({ recordId: "expected", rowVersion: 3 });
  });
});
