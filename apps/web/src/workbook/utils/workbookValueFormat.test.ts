import { describe, expect, it } from "vitest";
import { stringifyGridValue } from "./workbookValueFormat";

describe("workbookValueFormat.stringifyGridValue", () => {
  it("preserves primitive display strings from duplicate grid helpers", () => {
    expect(stringifyGridValue("  untrimmed  ")).toBe("  untrimmed  ");
    expect(stringifyGridValue(true)).toBe("true");
    expect(stringifyGridValue(false)).toBe("false");
    expect(stringifyGridValue(42)).toBe("42");
    expect(stringifyGridValue(Number.POSITIVE_INFINITY)).toBe("Infinity");
    expect(stringifyGridValue(Number.NaN)).toBe("NaN");
  });

  it("keeps unsupported and omitted values empty", () => {
    expect(stringifyGridValue(null)).toBe("");
    expect(stringifyGridValue(undefined)).toBe("");
    expect(stringifyGridValue([])).toBe("");
    expect(stringifyGridValue({})).toBe("");
    expect(stringifyGridValue(new Date("2026-04-24T12:00:00Z"))).toBe("");
  });

  it("joins collection display text without inventing labels", () => {
    expect(
      stringifyGridValue({
        items: [
          { display_text: "Host A", raw_text: "host-a" },
          { raw_text: "identity-b" },
          { display_text: "" },
          null,
          { ignored: true },
        ],
      }),
    ).toBe("Host A, identity-b, ");
  });
});
