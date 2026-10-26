import {
  gridShellTestId,
  workbookViewBarQueryControlsTestId,
} from "@cartulary/ui-contracts";
import { cleanup, render, screen } from "@testing-library/react";
import { useEffect, useState } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "../workbook/models/workbookSurfaceRegistry";

import {
  changeInputValue,
  waitForWorkbookRows,
  workbookAsyncTimeoutMs,
} from "./timelineWorkbookTestSupport";

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  vi.useRealTimers();
});

function IdentityRows({
  recordId,
  surface = timelineViewSchemaId,
}: {
  readonly recordId: string;
  readonly surface?: string;
}) {
  return (
    <section>
      <input
        aria-label="Unrelated row editor"
        defaultValue="private-authoring-value"
      />
      <div
        data-testid={workbookViewBarQueryControlsTestId(timelineViewSchemaId)}
      />
      <table data-testid={gridShellTestId(timelineViewSchemaId)}>
        <tbody>
          {surface === timelineViewSchemaId ? (
            // biome-ignore lint/a11y/noRedundantRoles: identity fixtures exercise the actual row selectors.
            <tr data-grid-record-id={recordId} role="row">
              <td>private-row-payload</td>
            </tr>
          ) : null}
        </tbody>
      </table>
      {surface !== timelineViewSchemaId ? (
        <table data-testid={gridShellTestId(surface)}>
          <tbody>
            {/* biome-ignore lint/a11y/noRedundantRoles: identity fixtures exercise the actual row selectors. */}
            <tr data-grid-record-id={recordId} role="row">
              <td>private-row-payload</td>
            </tr>
          </tbody>
        </table>
      ) : null}
    </section>
  );
}

function DelayedWorkbookRows() {
  const [mounted, setMounted] = useState(false);
  useEffect(() => {
    const handle = window.setTimeout(() => setMounted(true), 20);
    return () => window.clearTimeout(handle);
  }, []);

  return (
    <section>
      <div
        data-testid={workbookViewBarQueryControlsTestId(timelineViewSchemaId)}
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
  it("waits for row identity despite identical text without changing focus or dispatching actions", async () => {
    vi.useFakeTimers();
    const { container, rerender, getByRole } = render(
      <IdentityRows recordId="record-other" />,
    );
    const editor = getByRole("textbox");
    editor.focus();
    const focus = vi.spyOn(HTMLElement.prototype, "focus");
    const click = vi.spyOn(HTMLElement.prototype, "click");
    const dispatch = vi.spyOn(EventTarget.prototype, "dispatchEvent");
    const ready = vi.fn();
    const observation = waitForWorkbookRows({
      container,
      expectedRecordIds: ["record-expected"],
      surface: timelineViewSchemaId,
    }).then(ready);
    await vi.advanceTimersByTimeAsync(50);
    expect(ready).not.toHaveBeenCalled();
    rerender(<IdentityRows recordId="record-expected" />);
    await vi.advanceTimersByTimeAsync(50);
    await observation;
    expect(ready).toHaveBeenCalledOnce();
    expect(document.activeElement).toBe(editor);
    expect(focus).not.toHaveBeenCalled();
    expect(click).not.toHaveBeenCalled();
    expect(dispatch).not.toHaveBeenCalled();
  });

  it("rejects a matching row on the wrong surface until the correct surface hydrates", async () => {
    vi.useFakeTimers();
    const { container, rerender } = render(
      <IdentityRows recordId="record-expected" surface={hostsViewSchemaId} />,
    );
    const ready = vi.fn();
    const observation = waitForWorkbookRows({
      container,
      expectedRecordIds: ["record-expected"],
      surface: timelineViewSchemaId,
    }).then(ready);
    await vi.advanceTimersByTimeAsync(50);
    expect(ready).not.toHaveBeenCalled();
    rerender(<IdentityRows recordId="record-expected" />);
    await vi.advanceTimersByTimeAsync(50);
    await observation;
    expect(ready).toHaveBeenCalledOnce();
  });

  it("times out persistent row mismatch within the configured budget without payload values", async () => {
    vi.useFakeTimers();
    const { container } = render(<IdentityRows recordId="record-other" />);
    const ready = vi.fn();
    const rejected = vi.fn();
    const observation = waitForWorkbookRows({
      container,
      expectedRecordIds: ["record-expected"],
      surface: timelineViewSchemaId,
    }).then(ready, rejected);
    // Existing shell controls settle first; the identity wait retains its own configured deadline.
    await vi.advanceTimersByTimeAsync(50);
    expect(ready).not.toHaveBeenCalled();
    expect(rejected).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(workbookAsyncTimeoutMs);
    await observation;
    expect(rejected).toHaveBeenCalledOnce();
    const diagnostic = String(rejected.mock.calls[0]?.[0]);
    expect(diagnostic).toContain("record-expected");
    expect(diagnostic).toContain("record-other");
    expect(diagnostic).not.toContain("private-row-payload");
    expect(diagnostic).not.toContain("private-authoring-value");
  });

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
