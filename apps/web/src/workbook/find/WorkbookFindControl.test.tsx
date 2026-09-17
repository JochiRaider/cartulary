import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
} from "@testing-library/react";
import { useRef, useSyncExternalStore } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { WorkbookFindControl } from "./WorkbookFindControl";
import { WorkbookFindController } from "./WorkbookFindController";

afterEach(() => {
  cleanup();
  vi.useRealTimers();
});
function fixture() {
  const move = vi.fn(async () => "focused" as const);
  const owner = new WorkbookFindController(move);
  owner.setSource({
    lifetimeKey: "incident",
    readable: true,
    unavailableReason: null,
    stale: false,
    navigationKey: "window",
    presentation: {
      surface: { kind: "view_schema", viewSchemaId: "timeline" },
      rowIdentities: [{ kind: "core_record", recordId: "r1" }],
      fieldKeys: ["a", "b"],
      revision: 1,
    },
    readText: () => ["needle"],
  });
  function Harness() {
    const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
    const hostRef = useRef<HTMLDivElement>(null);
    const inputRef = useRef<HTMLTextAreaElement>(null);
    return (
      <WorkbookFindControl
        chromeMode="base"
        binding={{
          snapshot,
          hostRef,
          inputRef,
          available: true,
          capture: () => {},
          open: () => owner.open(null),
          close: owner.close,
          changeTerm: (term) => owner.setTerm(term),
          changeCase: (value) => owner.setMatchCase(value),
          navigate: (direction) => owner.navigate(direction),
        }}
      />
    );
  }
  render(<Harness />);
  return { owner, move };
}
describe("Workbook Find controls", () => {
  it("labels scope counts and navigation without moving while typing and collapses after explicit movement", async () => {
    vi.useFakeTimers();
    const { owner, move } = fixture();
    fireEvent.click(
      screen.getByRole("button", { name: "Find in loaded rows" }),
    );
    const input = screen.getByRole("textbox", { name: "Find in loaded rows" });
    expect(document.activeElement).toBe(input);
    expect(
      screen.getByText(/Other incident rows were not searched/),
    ).toBeTruthy();
    fireEvent.change(input, { target: { value: "needle" } });
    await act(async () => {
      await vi.runAllTimersAsync();
    });
    expect(move).not.toHaveBeenCalled();
    expect(screen.getByRole("status").textContent).toContain(
      "2 matching cells",
    );
    fireEvent.keyDown(input, { key: "Enter", isComposing: true });
    expect(move).not.toHaveBeenCalled();
    await act(async () => {
      fireEvent.keyDown(input, { key: "Enter" });
    });
    expect(screen.queryByRole("textbox")).toBeNull();
    expect(owner.getSnapshot().term).toBe("needle");
    expect(screen.getByRole("status").textContent).toContain("1 of 2");
    fireEvent.click(
      screen.getByRole("button", { name: "Find in loaded rows" }),
    );
    await act(async () => {
      await vi.runAllTimersAsync();
    });
    await act(async () => {
      fireEvent.keyDown(screen.getByRole("textbox"), {
        key: "Enter",
        shiftKey: true,
      });
    });
    expect(screen.getByRole("status").textContent).toContain("Wrapped to end.");
  });
  it("keeps empty and zero-result feedback scoped and clears on Escape", async () => {
    vi.useFakeTimers();
    const { owner, move } = fixture();
    fireEvent.click(
      screen.getByRole("button", { name: "Find in loaded rows" }),
    );
    expect(
      screen.getByRole("button", { name: "Next" }).hasAttribute("disabled"),
    ).toBe(true);
    fireEvent.change(screen.getByRole("textbox"), {
      target: { value: "missing" },
    });
    await act(async () => {
      await vi.runAllTimersAsync();
    });
    expect(screen.getByRole("status").textContent).toBe(
      "No matches in loaded rows.",
    );
    fireEvent.keyDown(screen.getByRole("textbox"), { key: "Escape" });
    expect(owner.getSnapshot().term).toBe("");
    expect(screen.queryByRole("textbox")).toBeNull();
    expect(move).not.toHaveBeenCalled();
  });
});
