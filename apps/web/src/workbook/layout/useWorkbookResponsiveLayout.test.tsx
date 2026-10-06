import { workbookLayoutMetrics } from "@cartulary/ui-contracts";
import { act, render, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { WorkbookQueryBrowsingProvider } from "../query/WorkbookQueryBrowsingContext";
import {
  currentWorkbookViewportSize,
  useWorkbookResponsiveLayout,
} from "./useWorkbookResponsiveLayout";
import { WorkbookSurfaceLayout } from "./WorkbookSurfaceLayout";

const originalVisualViewport = Object.getOwnPropertyDescriptor(
  window,
  "visualViewport",
);
const originalInnerWidth = Object.getOwnPropertyDescriptor(
  window,
  "innerWidth",
);
const originalInnerHeight = Object.getOwnPropertyDescriptor(
  window,
  "innerHeight",
);
const originalRootClientWidth = Object.getOwnPropertyDescriptor(
  document.documentElement,
  "clientWidth",
);

afterEach(() => {
  restoreWindowProperty("visualViewport", originalVisualViewport);
  restoreWindowProperty("innerWidth", originalInnerWidth);
  restoreWindowProperty("innerHeight", originalInnerHeight);
  restoreProperty(
    document.documentElement,
    "clientWidth",
    originalRootClientWidth,
  );
  document.documentElement.style.zoom = "";
});

describe("workbook responsive viewport", () => {
  it("uses real window dimensions when visualViewport is unavailable", () => {
    setWindowProperty("visualViewport", undefined);
    setWindowProperty("innerWidth", 768);
    setWindowProperty("innerHeight", 640);

    expect(currentWorkbookViewportSize()).toEqual({
      height: 640,
      width: 768,
    });
    const { result } = renderHook(() => useWorkbookResponsiveLayout());
    expect(result.current).toEqual({
      blockMode: "compact_height",
      chromeMode: "compact_desktop",
    });

    setWindowProperty("innerWidth", 767);
    setWindowProperty("innerHeight", 639);
    act(() => window.dispatchEvent(new Event("resize")));
    expect(result.current).toEqual({
      blockMode: "short_height",
      chromeMode: "below_supported_minimum",
    });

    let width = 1280;
    let widthReads = 0;
    Object.defineProperty(window, "innerWidth", {
      configurable: true,
      get: () => {
        widthReads += 1;
        return width;
      },
    });
    const frame = (feedback: string) => (
      <WorkbookQueryBrowsingProvider>
        <WorkbookSurfaceLayout
          viewSchemaId="viewport-test"
          primaryGrid={<div>Grid</div>}
          inspector={<div>Inspector</div>}
          viewBar={<div>View bar</div>}
          statusStrip={<div>Status</div>}
          workAreaFeedback={<div>{feedback}</div>}
        />
      </WorkbookQueryBrowsingProvider>
    );
    const surface = render(frame("Ready"));
    const initialReads = widthReads;
    surface.rerender(frame("Saved"));
    expect(widthReads).toBe(initialReads);
    width = 768;
    act(() => window.dispatchEvent(new Event("resize")));
    expect(
      surface
        .getByRole("separator", { name: "Resize inspector" })
        .getAttribute("aria-valuemax"),
    ).toBe(
      String(workbookLayoutMetrics(width).inspectorEffectiveMaxWidthCssPx),
    );
    const resizedReads = widthReads;
    surface.rerender(frame("Updated"));
    expect(widthReads).toBe(resizedReads);
    surface.unmount();
  });

  it("uses the effective root inline size for zoomed workbook chrome", () => {
    setWindowProperty("visualViewport", undefined);
    setWindowProperty("innerWidth", 1440);
    setWindowProperty("innerHeight", 900);
    setProperty(document.documentElement, "clientWidth", 1440);
    document.documentElement.style.zoom = "200%";

    expect(currentWorkbookViewportSize()).toEqual({
      height: 450,
      width: 720,
    });
    const { result } = renderHook(() => useWorkbookResponsiveLayout());
    expect(result.current.chromeMode).toBe("below_supported_minimum");
  });
});

function setWindowProperty(key: string, value: unknown): void {
  Object.defineProperty(window, key, {
    configurable: true,
    value,
    writable: true,
  });
}

function restoreWindowProperty(
  key: string,
  descriptor: PropertyDescriptor | undefined,
): void {
  if (descriptor === undefined) {
    Reflect.deleteProperty(window, key);
    return;
  }
  Object.defineProperty(window, key, descriptor);
}

function setProperty(target: object, key: string, value: unknown): void {
  Object.defineProperty(target, key, {
    configurable: true,
    value,
  });
}

function restoreProperty(
  target: object,
  key: string,
  descriptor: PropertyDescriptor | undefined,
): void {
  if (descriptor === undefined) {
    Reflect.deleteProperty(target, key);
    return;
  }
  Object.defineProperty(target, key, descriptor);
}
