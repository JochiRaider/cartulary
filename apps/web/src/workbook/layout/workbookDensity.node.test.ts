import { describe, expect, it } from "vitest";
import {
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import { resolveEffectiveWorkbookDensity } from "./workbookDensity";

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
