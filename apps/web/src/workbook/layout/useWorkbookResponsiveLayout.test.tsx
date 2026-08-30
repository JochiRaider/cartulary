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

afterEach(() => {
  restoreWindowProperty("visualViewport", originalVisualViewport);
  restoreWindowProperty("innerWidth", originalInnerWidth);
  restoreWindowProperty("innerHeight", originalInnerHeight);
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
