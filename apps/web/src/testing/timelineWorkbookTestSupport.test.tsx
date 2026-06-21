import {
  gridShellTestId,
  workbookTopBarQueryControlsTestId,
} from "@cartulary/ui-contracts";
import { cleanup, render, screen } from "@testing-library/react";
import { useEffect, useState } from "react";
import { afterEach, describe, expect, it } from "vitest";

import {
  changeInputValue,
  timelineViewSchemaId,
  waitForWorkbookRows,
} from "./timelineWorkbookTestSupport";

afterEach(() => {
  cleanup();
});

function DelayedWorkbookRows() {
  const [mounted, setMounted] = useState(false);
  useEffect(() => {
    const handle = window.setTimeout(() => setMounted(true), 20);
    return () => window.clearTimeout(handle);
  }, []);

  return (
    <section>
      <div
        data-testid={workbookTopBarQueryControlsTestId(timelineViewSchemaId)}
      />
      <table data-testid={gridShellTestId(timelineViewSchemaId)}>
        <tbody>
          {mounted ? (
            // biome-ignore lint/a11y/noRedundantRoles: shared grid selectors intentionally target explicit row roles.
            <tr data-grid-record-id="record-delayed" role="row">
              <td>delayed</td>
            </tr>
          ) : null}
        </tbody>
      </table>
    </section>
  );
}

function ControlledInput() {
  const [value, setValue] = useState("Pending");
  return (
    <input
      data-testid="controlled-input"
      onChange={(event) => setValue(event.currentTarget.value)}
      value={value}
    />
  );
}

describe("timeline workbook test support", () => {
  it("waits for delayed workbook row identity with the shared bounded helper", async () => {
    const { container } = render(<DelayedWorkbookRows />);

    await waitForWorkbookRows({
      container,
      expectedRecordIds: ["record-delayed"],
      surface: timelineViewSchemaId,
    });
  });

  it("replaces controlled input values without interleaving with the existing value", async () => {
    render(<ControlledInput />);
    const input = screen.getByTestId("controlled-input") as HTMLInputElement;

    await changeInputValue(input, "Pending browser fact");

    expect(input.value).toBe("Pending browser fact");
  });
});
