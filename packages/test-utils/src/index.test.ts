import { describe, expect, it } from "vitest";

import {
  gridDraftRowSelector,
  gridSavedRowsSelector,
  gridShellTestId,
} from "./index";

describe("@cartulary/test-utils workbook row selectors", () => {
  it("targets saved and draft workbook rows when scoped through the grid shell", () => {
    document.body.innerHTML = `
      <div data-testid="${gridShellTestId("timeline")}">
        <div role="row">header row</div>
        <div role="row" data-grid-record-id="record-1">saved row</div>
        <div role="row" data-grid-record-id="">draft row</div>
      </div>
      <div data-testid="${gridShellTestId("hosts")}">
        <div role="row" data-grid-record-id="host-1">host row</div>
      </div>
    `;

    const timelineShell = document.querySelector(
      `[data-testid="${gridShellTestId("timeline")}"]`,
    );
    const hostsShell = document.querySelector(
      `[data-testid="${gridShellTestId("hosts")}"]`,
    );

    expect(timelineShell?.querySelectorAll(gridSavedRowsSelector())).toHaveLength(
      1,
    );
    expect(timelineShell?.querySelectorAll(gridDraftRowSelector())).toHaveLength(
      1,
    );
    expect(hostsShell?.querySelectorAll(gridSavedRowsSelector())).toHaveLength(
      1,
    );
    expect(hostsShell?.querySelectorAll(gridDraftRowSelector())).toHaveLength(0);
  });
});
