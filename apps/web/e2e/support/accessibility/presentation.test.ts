// @vitest-environment jsdom
import type { Page } from "@playwright/test";
import { afterEach, describe, expect, it, vi } from "vitest";
import { withPresentation } from "./presentation";

function fixture(failRemove = false) {
  // jsdom does not implement CSS zoom; supply that CSSOM property while
  // retaining the real document/style lifecycle for all other observations.
  const style = document.documentElement.style;
  let zoom = "",
    priority = "";
  const set = style.setProperty.bind(style),
    get = style.getPropertyValue.bind(style),
    pri = style.getPropertyPriority.bind(style),
    remove = style.removeProperty.bind(style);
  vi.spyOn(style, "setProperty").mockImplementation(
    (name, value, important) => {
      if (name === "zoom") {
        zoom = value ?? "";
        priority = important ?? "";
      } else set(name, value, important);
    },
  );
  vi.spyOn(style, "getPropertyValue").mockImplementation((name) =>
    name === "zoom" ? zoom : get(name),
  );
  vi.spyOn(style, "getPropertyPriority").mockImplementation((name) =>
    name === "zoom" ? priority : pri(name),
  );
  vi.spyOn(style, "removeProperty").mockImplementation((name) => {
    if (name !== "zoom") return remove(name);
    const old = zoom;
    zoom = "";
    priority = "";
    return old;
  });
  let viewport = { width: 640, height: 480 };
  document.documentElement.style.setProperty("zoom", "125%", "important");
  document.documentElement.dataset.visualSnapshot = "prior";
  const page = {
    viewportSize: () => ({ ...viewport }),
    isClosed: () => false,
    setViewportSize: vi.fn(async (next) => {
      viewport = next;
    }),
    evaluate: async (fn: (arg: unknown) => unknown, arg: unknown) => fn(arg),
    addStyleTag: async ({ content }: { content: string }) => {
      const style = document.createElement("style");
      style.textContent = content;
      document.head.append(style);
      return {
        evaluate: async (fn: (node: HTMLElement) => void) => {
          if (failRemove) throw new Error("style removal failed");
          fn(style);
        },
      };
    },
  };
  return page as unknown as Page;
}
afterEach(() => {
  vi.restoreAllMocks();
  document.documentElement.removeAttribute("style");
  document.documentElement.removeAttribute("data-visual-snapshot");
  document.head.querySelectorAll("style").forEach((node) => {
    node.remove();
  });
  document.body.replaceChildren();
});
describe("temporary presentation lifetime", () => {
  it("restores exact prior viewport styles after observation failure", async () => {
    const page = fixture();
    const primary = new Error("capture failed");

    await expect(
      withPresentation(
        page,
        {
          zoom: "200%",
          textSpacing: "p { line-height: 1.5 }",
          visualSnapshot: true,
          viewport: { width: 1280, height: 720 },
        },
        async () => {
          throw primary;
        },
      ),
    ).rejects.toBe(primary);
    expect(page.viewportSize()).toEqual({ width: 640, height: 480 });
    expect(document.documentElement.style.getPropertyValue("zoom")).toBe(
      "125%",
    );
    expect(document.documentElement.style.getPropertyPriority("zoom")).toBe(
      "important",
    );
    expect(document.documentElement.dataset.visualSnapshot).toBe("prior");
    expect(document.head.querySelectorAll("style")).toHaveLength(0);
  });
  it("continues restoration when cleanup fails and preserves the primary failure", async () => {
    const page = fixture(true),
      primary = new Error("product failure"),
      report = vi.fn(async () => {});
    await expect(
      withPresentation(
        page,
        {
          zoom: "200%",
          textSpacing: "p {}",
          viewport: { width: 100, height: 100 },
        },
        async () => {
          throw primary;
        },
        report,
      ),
    ).rejects.toBe(primary);
    expect(report).toHaveBeenCalledOnce();
    expect(page.viewportSize()).toEqual({ width: 640, height: 480 });
    expect(document.documentElement.style.getPropertyValue("zoom")).toBe(
      "125%",
    );
    await expect(
      withPresentation(page, { textSpacing: "p {}" }, async () => {}, report),
    ).rejects.toThrow("Temporary presentation restoration failed");
  });
});
