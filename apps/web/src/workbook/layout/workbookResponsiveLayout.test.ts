import { describe, expect, it } from "vitest";

import {
  selectWorkbookBlockMode,
  selectWorkbookChromeMode,
  workbookQueryChipCapacity,
} from "./workbookResponsiveLayout";

describe("workbook responsive layout model", () => {
  it.each([
    [1280, "base"],
    [1440, "base"],
    [1279, "narrow_desktop"],
    [1024, "narrow_desktop"],
    [1023, "compact_desktop"],
    [768, "compact_desktop"],
    [767, "below_supported_minimum"],
    [Number.NaN, "below_supported_minimum"],
  ] as const)("selects chrome mode from width %s", (widthCssPx, expectedMode) => {
    expect(selectWorkbookChromeMode(widthCssPx)).toBe(expectedMode);
  });

  it.each([
    [720, "base_height"],
    [900, "base_height"],
    [719, "compact_height"],
    [640, "compact_height"],
    [639, "short_height"],
    [Number.NaN, "short_height"],
  ] as const)("selects block mode from height %s", (heightCssPx, expectedMode) => {
    expect(selectWorkbookBlockMode(heightCssPx)).toBe(expectedMode);
  });

  it.each([
    1280, 1440,
  ] as const)("keeps wide chrome in base mode across short heights at %s CSS px", (widthCssPx) => {
    expect(selectWorkbookChromeMode(widthCssPx)).toBe("base");
    expect(selectWorkbookBlockMode(560)).toBe("short_height");
  });

  it.each([
    ["base", 3],
    ["narrow_desktop", 2],
    ["compact_desktop", 0],
    ["below_supported_minimum", 0],
  ] as const)("caps visible query chips in %s mode at %i", (chromeMode, expectedCapacity) => {
    expect(workbookQueryChipCapacity(chromeMode)).toBe(expectedCapacity);
  });
});
