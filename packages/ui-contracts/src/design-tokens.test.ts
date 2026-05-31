import { describe, expect, it } from "vitest";
import {
  cartularyDefaultThemeId,
  cartularyDesignThemeCssText,
  cartularyDesignTokenVars,
} from ".";

describe("Cartulary design token exports", () => {
  it("exposes the adopted default theme as generated presentation tokens", () => {
    expect(cartularyDefaultThemeId).toBe("dark_graphite");
    expect(cartularyDesignTokenVars["--ct-colors-accent"]).toBe("#FACC15");
    expect(cartularyDesignTokenVars["--ct-border-hairline"]).toBe(
      "1px solid #2B313C",
    );
    expect(cartularyDesignTokenVars["--ct-component-inspector-padding"]).toBe(
      "16px",
    );
  });

  it("renders CSS variables without unresolved object values", () => {
    expect(cartularyDesignThemeCssText).toContain(
      '[data-cartulary-theme="dark_graphite"]',
    );
    expect(cartularyDesignThemeCssText).toContain(
      "--ct-component-button-primary-backgroundColor: #FACC15;",
    );
    expect(cartularyDesignThemeCssText).not.toContain("[object Object]");
  });
});
