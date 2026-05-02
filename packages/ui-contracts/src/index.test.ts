import { describe, expect, it } from "vitest";

import {
  gridDraftRowSelector,
  gridSavedRowsSelector,
  gridShellTestId,
  rowInspectorFieldTestId,
} from "./index";

describe("@cartulary/ui-contracts workbook row selectors", () => {
  it("derives inspector field ids from the stable row cell id", () => {
    expect(rowInspectorFieldTestId("record-1", "details")).toBe(
      "row-record-1-details-inspector",
    );
  });

  it("targets saved and draft workbook rows when scoped through the grid shell", () => {
    expect(gridShellTestId("timeline")).toBe("timeline-grid-shell");
    expect(gridSavedRowsSelector()).toBe(
      '[role="row"][data-grid-record-id]:not([data-grid-record-id=""])',
    );
    expect(gridDraftRowSelector()).toBe('[role="row"][data-grid-record-id=""]');
  });
});
