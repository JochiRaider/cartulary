import { describe, expect, it } from "vitest";
import {
  clipboardGridDimensions,
  clipboardTextLooksTabular,
  parseClipboardTable,
} from "./workbookClipboard";

describe("workbookClipboard", () => {
  it("parses empty, CRLF, tabular, and comma-delimited payloads", () => {
    expect(parseClipboardTable("")).toEqual([[""]]);
    expect(parseClipboardTable("\r\n")).toEqual([[""]]);
    expect(parseClipboardTable("a\tb\r\nc\t")).toEqual([
      ["a", "b"],
      ["c", ""],
    ]);
    expect(parseClipboardTable("a,b\nc,d\n\n")).toEqual([
      ["a", "b"],
      ["c", "d"],
    ]);
  });

  it("preserves quote-aware parser behavior used by Timeline and Shell paste dimensions", () => {
    expect(parseClipboardTable('"a,b","c""d"\n"line\nbreak",e')).toEqual([
      ["a,b", 'c"d'],
      ["line\nbreak", "e"],
    ]);
    expect(parseClipboardTable('"unclosed,a,b')).toEqual([["unclosed,a,b"]]);
  });

  it("computes dimensions without enforcing an implementation-local maximum", () => {
    const text = Array.from({ length: 4 }, (_, row) =>
      Array.from({ length: row + 1 }, (_, column) => `${row}:${column}`).join(
        "\t",
      ),
    ).join("\n");
    expect(clipboardGridDimensions(text)).toEqual({
      columnCount: 4,
      rowCount: 4,
    });
    expect(clipboardGridDimensions("single")).toEqual({
      columnCount: 1,
      rowCount: 1,
    });
  });

  it("keeps scalar comma text out of interactive tabular dispatch", () => {
    expect(clipboardTextLooksTabular("one,two")).toBe(false);
    expect(clipboardTextLooksTabular("one\ttwo")).toBe(true);
    expect(clipboardTextLooksTabular("one\ntwo")).toBe(true);
    expect(clipboardTextLooksTabular("one\rtwo")).toBe(true);
  });
});
