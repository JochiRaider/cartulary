import { describe, expect, it } from "vitest";

import { collectionItems, requireItemByRawText } from "./mentions";

describe("mention collection envelopes", () => {
  it("preserves the collection-actions item envelope", () => {
    const items = collectionItems(
      {
        cells: {
          "timeline.host_refs": {
            value: {
              items: [
                { item_ref: "item-1", raw_text: " vpn   gateway" },
                null,
                "not-an-item",
              ],
            },
          },
        },
      },
      "timeline.host_refs",
    );

    expect(items).toEqual([{ item_ref: "item-1", raw_text: " vpn   gateway" }]);
    expect(requireItemByRawText(items, " vpn   gateway")).toMatchObject({
      item_ref: "item-1",
    });
  });

  it("treats omitted, scalar, array, and malformed item payloads as empty", () => {
    for (const value of [undefined, null, "item", [], {}, { items: null }]) {
      expect(
        collectionItems(
          {
            cells:
              value === undefined ? {} : { "timeline.host_refs": { value } },
          },
          "timeline.host_refs",
        ),
      ).toEqual([]);
    }
  });

  it("reports the exact raw-text lookup that failed", () => {
    expect(() => requireItemByRawText([], "missing value")).toThrow(
      "missing collection item raw_text=missing value",
    );
  });
});
