import { act, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import {
  currentWorkbookViewportSize,
  useWorkbookResponsiveLayout,
} from "./useWorkbookResponsiveLayout";

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
  });

  it("uses the effective root inline size for zoomed workbook chrome", () => {
    setWindowProperty("visualViewport", undefined);
    setWindowProperty("innerWidth", 1440);
    setWindowProperty("innerHeight", 900);
    setProperty(document.documentElement, "clientWidth", 1440);
    document.documentElement.style.zoom = "200%";

    expect(currentWorkbookViewportSize()).toEqual({
      height: 900,
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
