// @vitest-environment node
import type { Locator } from "@playwright/test";
import { expect, it, vi } from "vitest";

const observe = vi.hoisted(() => ({
  count: 1,
  visible: true,
  name: "Save",
  calls: [] as string[],
}));
vi.mock("@playwright/test", () => ({
  expect: () => ({
    toHaveCount: async (expected: number) => {
      observe.calls.push("count");
      expect(observe.count).toBe(expected);
    },
    toBeVisible: async () => {
      observe.calls.push("visible");
      expect(observe.visible).toBe(true);
    },
    toHaveAccessibleName: async (expected: string) => {
      observe.calls.push("native-name");
      expect(observe.name).toBe(expected);
    },
  }),
}));

import { expectNamedControl } from "./assertions";

it("requires unique visible controls before delegating the exact expected name to Playwright", async () => {
  const locator = {} as Locator;
  for (const count of [0, 2]) {
    observe.count = count;
    observe.calls = [];
    await expect(expectNamedControl(locator, "Save")).rejects.toThrow();
    expect(observe.calls).toEqual(["count"]);
  }
  observe.count = 1;
  observe.visible = false;
  observe.calls = [];
  await expect(expectNamedControl(locator, "Save")).rejects.toThrow();
  expect(observe.calls).toEqual(["count", "visible"]);
  observe.visible = true;
  observe.calls = [];
  await expectNamedControl(locator, "Save");
  expect(observe.calls).toEqual(["count", "visible", "native-name"]);
  await expect(expectNamedControl(locator, "Unrelated text")).rejects.toThrow();
});
