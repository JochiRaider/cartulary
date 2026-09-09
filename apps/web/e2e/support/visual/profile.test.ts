import { describe, expect, it } from "vitest";
import { assertVisualPresentation, type VisualPresentation } from "./profile";

describe("visual presentation declarations", () => {
  const workbook: VisualPresentation = {
    surface_kind: "workbook_shell",
    density_id: "compact",
    theme_id: "dark_graphite",
  };
  it("rejects missing empty unknown and mismatched applicable observations", () => {
    for (const density_id of [undefined, null, "", "unknown", "comfortable"]) {
      expect(() =>
        assertVisualPresentation(workbook, {
          density_id,
          theme_id: "dark_graphite",
        }),
      ).toThrowError(/declared surface/);
    }
    for (const theme_id of [undefined, null, "", "unknown"]) {
      expect(() =>
        assertVisualPresentation(workbook, { density_id: "compact", theme_id }),
      ).toThrowError(/declared surface/);
    }
  });
  it("permits inapplicable density only on an application shell", () => {
    const application = {
      ...workbook,
      surface_kind: "application_shell" as const,
      density_id: null,
    };
    expect(() =>
      assertVisualPresentation(application, application),
    ).not.toThrow();
    expect(() =>
      assertVisualPresentation({ ...workbook, density_id: null }, application),
    ).toThrow();
    expect(() =>
      assertVisualPresentation(
        { ...application, density_id: "compact" },
        workbook,
      ),
    ).toThrow();
  });
  it("accepts every declared workbook density and classifies invalid evidence", () => {
    for (const density_id of ["compact", "default", "comfortable"] as const) {
      const declaration = { ...workbook, density_id };
      expect(() =>
        assertVisualPresentation(declaration, declaration),
      ).not.toThrow();
    }
    try {
      assertVisualPresentation(workbook, { density_id: "", theme_id: "" });
    } catch (error) {
      expect(error).toMatchObject({ name: "CartularyVisualCaptureError" });
    }
  });
});
