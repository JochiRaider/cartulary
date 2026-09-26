import type { GridHandle } from "@cartulary/grid-adapter";
import { workbookIncidentIdentityTestId } from "@cartulary/ui-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  type ParkedGridDraft,
  WorkbookParkedGridDrafts,
} from "./WorkbookParkedGridDrafts";

type Entry = { key: string; label: string; value: string | null };

const firstEntry: Entry = { key: "one", label: "One", value: "  first Ω  " };
const middleEntry: Entry = { key: "two", label: "Two", value: null };
const lastEntry: Entry = {
  key: "three",
  label: "Three",
  value: "  third 東京  ",
};
const entries: Entry[] = [firstEntry, middleEntry, lastEntry];

afterEach(() => {
  cleanup();
  document.body.replaceChildren();
});

function setup(initial: readonly Entry[] = entries) {
  const grid = document.createElement("div");
  grid.tabIndex = 0;
  document.body.append(grid);
  const requestFocus = vi.fn<GridHandle["requestFocus"]>(
    async (_target, options) => {
      if (options?.signal?.aborted) return "cancelled";
      grid.focus();
      return "focused";
    },
  );
  const gridHandleRef = {
    current: {
      getScrollElement: () => grid,
      requestFocus,
    } as unknown as GridHandle,
  };
  let current = [...initial];
  let scope = "surface-1";
  let view: ReturnType<typeof render>;
  const discard = vi.fn((key: string) => {
    current = current.filter((entry) => entry.key !== key);
    view.rerender(element());
  });
  const element = () => (
    <>
      <button type="button">External control</button>
      <WorkbookParkedGridDrafts
        drafts={current.map(
          (entry): ParkedGridDraft => ({
            ...entry,
            recordId: `record-${entry.key}`,
            reason: "Original field is unavailable.",
            discard: () => discard(entry.key),
          }),
        )}
        focusScopeKey={scope}
        gridHandleRef={gridHandleRef}
      />
      <button type="button">Following control</button>
    </>
  );
  view = render(element());
  const details = () => document.querySelector("details");
  const open = () => {
    const disclosure = details();
    if (!disclosure) throw new Error("Missing disclosure");
    disclosure.open = true;
  };
  const reconcile = (next: readonly Entry[]) => {
    current = [...next];
    view.rerender(element());
  };
  const changeScope = (next: string) => {
    scope = next;
    view.rerender(element());
  };
  return {
    changeScope,
    details,
    discard,
    grid,
    open,
    reconcile,
    requestFocus,
    view,
  };
}

describe("WorkbookParkedGridDrafts focus continuity", () => {
  it("moves first and middle discard to the next retained text without changing surviving values", () => {
    const h = setup();
    h.open();
    const first = screen.getByRole("button", { name: "Discard One draft" });
    first.focus();
    fireEvent.click(first);
    expect(screen.getByRole("textbox", { name: "Retained Two" })).toBe(
      document.activeElement,
    );
    expect(screen.getByText("Explicit clear")).toBeTruthy();
    expect(
      screen.getByRole<HTMLTextAreaElement>("textbox", {
        name: "Retained Three",
      }).value,
    ).toBe("  third 東京  ");
    const middle = screen.getByRole("button", { name: "Discard Two draft" });
    middle.focus();
    fireEvent.click(middle);
    expect(screen.getByRole("textbox", { name: "Retained Three" })).toBe(
      document.activeElement,
    );
    expect(h.discard).toHaveBeenCalledTimes(2);
    expect(h.requestFocus).not.toHaveBeenCalled();
  });

  it("moves last-position removal to the previous retained text", () => {
    const h = setup();
    h.open();
    const last = screen.getByRole("button", { name: "Discard Three draft" });
    last.focus();
    fireEvent.click(last);
    expect(screen.getByRole("textbox", { name: "Retained Two" })).toBe(
      document.activeElement,
    );
    expect(
      screen.getByRole<HTMLTextAreaElement>("textbox", { name: "Retained One" })
        .value,
    ).toBe("  first Ω  ");
  });

  it("returns final draft removal and focused summary loss to the grid root", async () => {
    const h = setup([firstEntry]);
    h.open();
    const discard = screen.getByRole("button", { name: "Discard One draft" });
    discard.focus();
    fireEvent.click(discard);
    await waitFor(() => expect(document.activeElement).toBe(h.grid));
    expect(h.details()).toBeNull();
    expect(h.requestFocus).toHaveBeenCalledWith(
      { kind: "root" },
      expect.objectContaining({ preserveSelection: true }),
    );

    h.reconcile([firstEntry]);
    h.open();
    const summary = screen.getByText("Unsaved cells (1)");
    summary.focus();
    h.reconcile([]);
    await waitFor(() => expect(document.activeElement).toBe(h.grid));
  });

  it("handles focused reconciliation and preserves focus during background removal", () => {
    const h = setup();
    h.open();
    screen.getByRole("textbox", { name: "Retained Two" }).focus();
    h.reconcile([firstEntry, lastEntry]);
    expect(screen.getByRole("textbox", { name: "Retained Three" })).toBe(
      document.activeElement,
    );
    h.changeScope("surface-2");
    h.reconcile([firstEntry]);
    expect(screen.getByRole("textbox", { name: "Retained One" })).toBe(
      document.activeElement,
    );
    const external = screen.getByRole("button", { name: "External control" });
    external.focus();
    h.reconcile([lastEntry]);
    expect(document.activeElement).toBe(external);
    expect(h.requestFocus).not.toHaveBeenCalled();
    h.reconcile([]);
    expect(document.activeElement).toBe(external);
    expect(h.requestFocus).not.toHaveBeenCalled();
  });

  it("does not open a closed disclosure for a background update", () => {
    const h = setup([firstEntry]);
    const external = screen.getByRole("button", { name: "External control" });
    external.focus();
    h.reconcile(entries);
    expect(h.details()?.open).toBe(false);
    expect(document.activeElement).toBe(external);
  });

  it("uses the visible shell fallback when the grid root is unavailable", async () => {
    const h = setup([firstEntry]);
    h.requestFocus.mockResolvedValue("unavailable");
    const navigation = document.createElement("nav");
    navigation.setAttribute("aria-label", "Built-in workbook surfaces");
    const selector = document.createElement("button");
    selector.setAttribute("aria-current", "page");
    selector.getClientRects = () =>
      [new DOMRect(0, 0, 50, 25)] as unknown as DOMRectList;
    navigation.append(selector);
    document.body.append(navigation);
    h.open();
    const discard = screen.getByRole("button", { name: "Discard One draft" });
    discard.focus();
    fireEvent.click(discard);
    await waitFor(() => expect(document.activeElement).toBe(selector));

    h.reconcile([firstEntry]);
    navigation.remove();
    const incident = document.createElement("div");
    incident.dataset.testid = workbookIncidentIdentityTestId();
    const identity = document.createElement("button");
    identity.getClientRects = () =>
      [new DOMRect(0, 0, 50, 25)] as unknown as DOMRectList;
    incident.append(identity);
    document.body.append(incident);
    h.open();
    const second = screen.getByRole("button", { name: "Discard One draft" });
    second.focus();
    fireEvent.click(second);
    await waitFor(() => expect(document.activeElement).toBe(identity));
  });

  it("cancels pending grid fallback for newer focus, scope change, and unmount", async () => {
    const h = setup([firstEntry]);
    let resolve!: (value: "unavailable") => void;
    h.requestFocus.mockImplementation(
      () =>
        new Promise((complete) => {
          resolve = complete;
        }),
    );
    h.open();
    const discard = screen.getByRole("button", { name: "Discard One draft" });
    discard.focus();
    fireEvent.click(discard);
    const signal = h.requestFocus.mock.calls[0]?.[1]?.signal;
    expect(signal?.aborted).toBe(false);
    const external = screen.getByRole("button", { name: "External control" });
    external.focus();
    expect(signal?.aborted).toBe(true);
    await act(async () => resolve("unavailable"));
    expect(document.activeElement).toBe(external);

    h.reconcile([firstEntry]);
    h.open();
    const second = screen.getByRole("button", { name: "Discard One draft" });
    second.focus();
    fireEvent.click(second);
    const secondSignal = h.requestFocus.mock.calls[1]?.[1]?.signal;
    h.changeScope("surface-2");
    expect(secondSignal?.aborted).toBe(true);
    h.view.unmount();
    await act(async () => resolve("unavailable"));
    expect(document.activeElement).not.toBe(h.grid);
  });
});
