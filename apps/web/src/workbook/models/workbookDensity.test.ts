import { describe, expect, it } from "vitest";
import { resolveEffectiveWorkbookDensity } from "./workbookDensity";
import {
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "./workbookSurfaceRegistry";

describe("workbook density resolution", () => {
  it("keeps Core null defaults while letting explicit account density override every workbook surface", () => {
    expect(resolveEffectiveWorkbookDensity(timelineViewSchemaId, null)).toBe(
      "compact",
    );
    expect(resolveEffectiveWorkbookDensity(hostsViewSchemaId, null)).toBe(
      "default",
    );
    expect(
      resolveEffectiveWorkbookDensity(timelineViewSchemaId, "comfortable"),
    ).toBe("comfortable");
    expect(
      resolveEffectiveWorkbookDensity(hostsViewSchemaId, "comfortable"),
    ).toBe("comfortable");
  });
});
