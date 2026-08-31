import { describe, expect, it } from "vitest";
import {
  cartularyDefaultThemeId,
  cartularyDesignThemeCssText,
  cartularyDesignTokenReference,
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
    expect(cartularyDesignTokenVars["--ct-density-compact-rowHeight"]).toBe(
      "24px",
    );
    expect(cartularyDesignTokenVars["--ct-density-compact-cellPadding"]).toBe(
      "2px 5px",
    );
    expect(cartularyDesignTokenVars["--ct-density-compact-fontSize"]).toBe(
      "12px",
    );
    expect(cartularyDesignTokenVars["--ct-density-compact-lineHeight"]).toBe(
      "1.2",
    );
    expect(cartularyDesignTokenVars["--ct-density-default-rowHeight"]).toBe(
      "32px",
    );
    expect(cartularyDesignTokenVars["--ct-density-default-fontSize"]).toBe(
      "13px",
    );
    expect(cartularyDesignTokenVars["--ct-density-comfortable-rowHeight"]).toBe(
      "40px",
    );
    expect(cartularyDesignTokenVars["--ct-density-comfortable-fontSize"]).toBe(
      "14px",
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

  it("builds typed references to authored design-token variables", () => {
    expect(cartularyDesignTokenReference("--ct-colors-ink")).toBe(
      "var(--ct-colors-ink)",
    );
    expect(
      cartularyDesignTokenReference(
        "--ct-component-text-input-backgroundColor",
      ),
    ).toBe("var(--ct-component-text-input-backgroundColor)");
  });

  it("exposes explicit font role tokens without changing the default UI or mono stacks", () => {
    expect(cartularyDesignTokenVars["--ct-typography-ui-fontFamily"]).toMatch(
      /^Inter,/u,
    );
    expect(cartularyDesignTokenVars["--ct-typography-grid-fontFamily"]).toMatch(
      /^Inter,/u,
    );
    expect(
      cartularyDesignTokenVars["--ct-typography-grid-cell-fontFamily"],
    ).toMatch(/^Inter,/u);
    expect(cartularyDesignTokenVars["--ct-typography-mono-fontFamily"]).toMatch(
      /^JetBrains Mono,/u,
    );
    expect(
      cartularyDesignTokenVars["--ct-typography-ui-fontFamily"],
    ).not.toContain("Geist");
    expect(
      cartularyDesignTokenVars["--ct-typography-mono-fontFamily"],
    ).not.toContain("Geist Mono");
    expect(
      cartularyDesignTokenVars["--ct-typography-alternate-ui-fontFamily"],
    ).toMatch(/^Geist,/u);
    expect(
      cartularyDesignTokenVars["--ct-typography-alternate-mono-fontFamily"],
    ).toMatch(/^Geist Mono,/u);
    expect(
      cartularyDesignTokenVars["--ct-typography-report-narrative-fontFamily"],
    ).toMatch(/^Source Serif 4,/u);
    expect(
      cartularyDesignTokenVars["--ct-typography-accessible-reading-fontFamily"],
    ).toMatch(/^Atkinson Hyperlegible,/u);
    expect(
      cartularyDesignTokenVars["--ct-typography-compact-metadata-fontFamily"],
    ).toMatch(/^IBM Plex Sans Condensed,/u);
  });
});
