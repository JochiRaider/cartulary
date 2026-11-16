import { describe, expect, it } from "vitest";
import fixtureCorpus from "../../../contracts/tabularingest/clipboard.v1.json";
import {
  clipboardLimits,
  decodeDelimitedClipboard,
  decodeGridClipboard,
  encodeClipboardTable,
  encodeGridClipboard,
} from "./clipboardCodec";

describe("clipboard representation", () => {
  it("preserves values and rejects invalid representations", () => {
    const corpus = fixtureCorpus as {
      limits: typeof clipboardLimits;
      cases: {
        id: string;
        format: "tsv" | "csv";
        text: string;
        values?: string[][];
        error?: string;
      }[];
    };
    expect(clipboardLimits).toEqual(corpus.limits);
    for (const fixture of corpus.cases) {
      const result = decodeDelimitedClipboard(fixture.text, fixture.format);
      expect(result, fixture.id).toEqual(
        fixture.error
          ? expect.objectContaining({ kind: "failure", reason: fixture.error })
          : { kind: "decoded", values: fixture.values },
      );
      if (fixture.values)
        expect(
          decodeDelimitedClipboard(
            encodeClipboardTable(fixture.values, fixture.format),
            fixture.format,
          ),
        ).toEqual({ kind: "decoded", values: fixture.values });
    }
    for (const values of [
      [["a,b"], ["c,d"]],
      [['a"b']],
      [[""]],
      [[""], [""]],
      [
        [
          "  界😀001  ",
          "=1+1",
          "'=literal",
          "\tvalue",
          "\rvalue",
          "a\r\nb\nc\td",
        ],
        ["", "-4", "+4", "@x", "''", "2026-09-15"],
      ],
    ]) {
      const offered = encodeGridClipboard(values);
      expect(offered).not.toHaveProperty("kind");
      if ("kind" in offered) continue;
      const decoded = decodeGridClipboard(offered);
      expect(decoded).toMatchObject(
        values.length === 1 && values[0]?.length === 1
          ? { kind: "scalar", value: values[0][0] }
          : { kind: "table", values, headerMode: "none" },
      );
    }
    expect(
      decodeGridClipboard({ "text/plain": 'Hello, "world"' }),
    ).toMatchObject({ kind: "scalar", value: 'Hello, "world"' });
    expect(decodeGridClipboard({ "text/plain": "" })).toEqual({ kind: "noop" });
    expect(
      decodeGridClipboard({
        "text/html": "<b>text</b>",
        "text/plain": "fallback",
      }),
    ).toMatchObject({ value: "fallback" });
    expect(
      decodeGridClipboard({
        "text/html": "<table><tr><td>  a<br>b &amp; c</td></tr></table>",
        "text/plain": "wrong",
      }),
    ).toMatchObject({ value: "  a\nb & c" });
    for (const [html, reason] of [
      [
        '<table data-cartulary-clipboard="2"><tr><td>a</td></tr></table>',
        "unsupported_table",
      ],
      ['<table><tr><td colspan="2">a</td></tr></table>', "unsupported_table"],
      [
        "<table><tr><td><table><tr><td>a</td></tr></table></td></tr></table>",
        "unsupported_table",
      ],
      [
        '<table><tr><td><img src="https://invalid.test/no-fetch"></td></tr></table>',
        "unsupported_table",
      ],
      [
        "<table><tr><td><script>alert(1)</script></td></tr></table>",
        "unsupported_table",
      ],
      [
        "<table><tr><td>a</td></tr><tr><td>b</td><td>c</td></tr></table>",
        "ragged_rows",
      ],
    ] as const)
      expect(
        decodeGridClipboard({
          "text/html": html,
          "text/plain": "must not fall back",
        }),
        html,
      ).toMatchObject({ kind: "failure", reason });
    // The byte bound is paired at both entry points, and counted in UTF-8
    // bytes rather than UTF-16 code units (a four-byte scalar is two units).
    const atLimit = "x".repeat(clipboardLimits.maxBytes);
    const fourByteAtLimit = "😀".repeat(clipboardLimits.maxBytes / 4);
    expect(decodeDelimitedClipboard(atLimit, "tsv").kind).toBe("decoded");
    expect(decodeDelimitedClipboard(`${atLimit}x`, "tsv")).toMatchObject({
      kind: "failure",
      reason: "too_large",
    });
    expect(decodeGridClipboard({ "text/plain": atLimit }).kind).toBe("scalar");
    expect(decodeGridClipboard({ "text/plain": `${atLimit}x` })).toMatchObject({
      kind: "failure",
      reason: "too_large",
    });
    expect(decodeGridClipboard({ "text/plain": fourByteAtLimit }).kind).toBe(
      "scalar",
    );
    expect(
      decodeGridClipboard({ "text/plain": `${fourByteAtLimit}x` }),
    ).toMatchObject({ kind: "failure", reason: "too_large" });
    expect(
      decodeDelimitedClipboard(Array(65).fill("x").join("\t"), "tsv"),
    ).toMatchObject({ reason: "too_many_columns" });
    expect(
      decodeDelimitedClipboard(Array(502).fill("x").join("\n"), "tsv"),
    ).toMatchObject({ reason: "too_many_rows" });
    // Exact bounds: 64 columns and 500 rows plus one header row are admitted.
    expect(
      decodeDelimitedClipboard(Array(64).fill("x").join("\t"), "tsv").kind,
    ).toBe("decoded");
    expect(
      decodeDelimitedClipboard(Array(501).fill("x").join("\n"), "tsv").kind,
    ).toBe("decoded");
    const bounded = Array.from({ length: 500 }, () => Array(64).fill("x"));
    expect(encodeGridClipboard(bounded)).not.toHaveProperty("kind");
    expect(
      encodeGridClipboard([...bounded, Array(64).fill("x")]),
    ).toMatchObject({
      reason: "too_many_rows",
    });
    expect(encodeGridClipboard([Array(65).fill("x")])).toMatchObject({
      reason: "too_many_columns",
    });
  });
});
