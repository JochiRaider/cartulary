import { describe, expect, it } from "vitest";
import { decodeWorkbookClipboardInput } from "./workbookClipboard";

describe("workbookClipboard", () => {
  it("decodes explicit CSV and plain TSV without punctuation guessing", () => {
    expect(
      decodeWorkbookClipboardInput({ "text/plain": "a,b\nc,d" }),
    ).toMatchObject({ format: "tsv", values: [["a,b"], ["c,d"]] });
    expect(
      decodeWorkbookClipboardInput({ "text/csv": '"a,b","c""d"' }),
    ).toMatchObject({ format: "csv", values: [["a,b", 'c"d']] });
    expect(
      decodeWorkbookClipboardInput({ "text/plain": "a\tb\r\nc\t" }),
    ).toMatchObject({
      values: [
        ["a", "b"],
        ["c", ""],
      ],
    });
  });
  it("preserves explicit empty records and rejects malformed or missing cells", () => {
    expect(decodeWorkbookClipboardInput({ "text/plain": "" })).toEqual({
      kind: "noop",
    });
    expect(decodeWorkbookClipboardInput({ "text/csv": "" })).toMatchObject({
      kind: "failure",
      reason: "empty_table",
    });
    expect(
      decodeWorkbookClipboardInput({ "text/plain": "\na\n\n" }),
    ).toMatchObject({ values: [[""], ["a"], [""]] });
    expect(
      decodeWorkbookClipboardInput({ "text/csv": '"unclosed,a,b' }),
    ).toMatchObject({ reason: "malformed_quotes" });
    expect(
      decodeWorkbookClipboardInput({ "text/plain": "a\tb\nc" }),
    ).toMatchObject({ reason: "ragged_rows" });
  });
  it("enforces the owner row limit before destination planning", () => {
    expect(
      decodeWorkbookClipboardInput({
        "text/plain": Array(500).fill("x").join("\n"),
      }),
    ).toMatchObject({ kind: "table" });
    expect(
      decodeWorkbookClipboardInput({
        "text/plain": Array(501).fill("x").join("\n"),
      }),
    ).toMatchObject({ reason: "too_many_rows" });
  });
  it("keeps scalar comma text out of interactive tabular dispatch", () => {
    for (const text of ["Hello, world", 'a"b', '"a""b"']) {
      expect(decodeWorkbookClipboardInput({ "text/plain": text })).toEqual({
        kind: "scalar",
        rawText: text,
        value: text,
      });
    }
    expect(
      decodeWorkbookClipboardInput({ "text/plain": '"one,two"\nthree' }),
    ).toMatchObject({ format: "tsv", values: [["one,two"], ["three"]] });
  });
});
