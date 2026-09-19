// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  createSemanticFocusRequests,
  type GridFocusResolution,
} from "./semanticFocusRequest";

afterEach(() => document.body.replaceChildren());

describe("semantic focus requests", () => {
  it("keeps explicit cell focus out of roving descendant controls", async () => {
    const root = document.createElement("div");
    const cell = document.createElement("div");
    const chip = document.createElement("button");
    cell.append(chip);
    root.append(cell);
    document.body.append(root);
    root.addEventListener("focusin", (event) => {
      if (event.target === cell) chip.focus();
    });
    const nativeObserver = vi.fn();
    document.addEventListener("focusin", nativeObserver, true);
    try {
      const requests = createSemanticFocusRequests({
        prepare: () => {},
        resolve: () => ({ kind: "target", element: cell }),
      });
      expect(
        await requests.requestFocus({
          kind: "cell",
          anchor: {
            surface: { kind: "view_schema", viewSchemaId: "timeline" },
            rowIdentity: { kind: "core_record", recordId: "row" },
            fieldKey: "refs",
          },
        }),
      ).toBe("focused");
      expect(document.activeElement).toBe(cell);
      expect(nativeObserver).toHaveBeenCalledOnce();
      chip.focus();
      expect(document.activeElement).toBe(chip);
    } finally {
      document.removeEventListener("focusin", nativeObserver, true);
    }
  });

  it("waits for registration and acknowledges only the declared primary control", async () => {
    let resolution: GridFocusResolution = { kind: "pending" };
    const requests = createSemanticFocusRequests({
      prepare: vi.fn(),
      resolve: () => resolution,
    });
    const settled = vi.fn();
    const result = requests.requestFocus({
      kind: "draft",
      fieldKey: "create.only",
    });
    void result.then(settled);
    await Promise.resolve();
    expect(settled).not.toHaveBeenCalled();
    const input = document.createElement("input");
    document.body.append(input);
    resolution = { kind: "target", element: input };
    requests.refresh();
    expect(await result).toBe("focused");
    expect(document.activeElement).toBe(input);
    expect(settled).toHaveBeenCalledOnce();
  });

  it("cancels superseded and aborted requests without focusing their late controls", async () => {
    let resolution: GridFocusResolution = { kind: "pending" };
    const requests = createSemanticFocusRequests({
      prepare: () => {},
      resolve: () => resolution,
    });
    const first = requests.requestFocus({ kind: "draft", fieldKey: "first" });
    const abort = new AbortController();
    const second = requests.requestFocus(
      { kind: "draft", fieldKey: "second" },
      { signal: abort.signal },
    );
    expect(await first).toBe("cancelled");
    abort.abort();
    expect(await second).toBe("cancelled");
    const input = document.createElement("input");
    document.body.append(input);
    resolution = { kind: "target", element: input };
    requests.refresh();
    await Promise.resolve();
    expect(document.activeElement).not.toBe(input);
  });

  it("rejects unavailable and disabled targets and cancels on user navigation", async () => {
    const input = document.createElement("input");
    document.body.append(input);
    let resolution: GridFocusResolution = { kind: "target", element: input };
    const requests = createSemanticFocusRequests({
      prepare: () => {},
      resolve: () => resolution,
    });
    input.disabled = true;
    expect(
      await requests.requestFocus({ kind: "draft", fieldKey: "first" }),
    ).toBe("unavailable");
    input.disabled = false;
    const wrapper = document.createElement("div");
    wrapper.style.display = "none";
    document.body.append(wrapper);
    wrapper.append(input);
    expect(
      await requests.requestFocus({ kind: "draft", fieldKey: "first" }),
    ).toBe("unavailable");
    resolution = { kind: "pending" };
    const result = requests.requestFocus({ kind: "draft", fieldKey: "first" });
    document.dispatchEvent(new KeyboardEvent("keydown", { key: "Tab" }));
    expect(await result).toBe("cancelled");
    resolution = { kind: "unavailable" };
    expect(
      await requests.requestFocus({ kind: "draft", fieldKey: "removed" }),
    ).toBe("unavailable");
  });

  it("does not complete against a detached control and releases cancelled lifetimes", async () => {
    const input = document.createElement("input");
    document.body.append(input);
    let resolution: GridFocusResolution = { kind: "target", element: input };
    const requests = createSemanticFocusRequests({
      prepare: () => {},
      resolve: () => resolution,
    });
    const result = requests.requestFocus({ kind: "draft", fieldKey: "first" });
    await Promise.resolve();
    input.remove();
    resolution = { kind: "pending" };
    await Promise.resolve();
    requests.cancel();
    expect(await result).toBe("cancelled");
  });
});
