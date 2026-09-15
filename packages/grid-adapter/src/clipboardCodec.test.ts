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
    for (const html of [
      '<table data-cartulary-clipboard="2"><tr><td>a</td></tr></table>',
      '<table><tr><td colspan="2">a</td></tr></table>',
      "<table><tr><td><table><tr><td>a</td></tr></table></td></tr></table>",
      '<table><tr><td><img src="https://invalid.test/no-fetch"></td></tr></table>',
      "<table><tr><td><script>alert(1)</script></td></tr></table>",
      "<table><tr><td>a</td></tr><tr><td>b</td><td>c</td></tr></table>",
    ])
      expect(
        decodeGridClipboard({
          "text/html": html,
          "text/plain": "must not fall back",
        }),
      ).toMatchObject({ kind: "failure" });
    expect(
      decodeDelimitedClipboard("x".repeat(clipboardLimits.maxBytes), "tsv")
        .kind,
    ).toBe("decoded");
    expect(
      decodeGridClipboard({
        "text/plain": "x".repeat(clipboardLimits.maxBytes + 1),
      }),
    ).toMatchObject({ reason: "too_large" });
    expect(
      decodeDelimitedClipboard(Array(65).fill("x").join("\t"), "tsv"),
    ).toMatchObject({ reason: "too_many_columns" });
    expect(
      decodeDelimitedClipboard(Array(502).fill("x").join("\n"), "tsv"),
    ).toMatchObject({ reason: "too_many_rows" });
  });
});
