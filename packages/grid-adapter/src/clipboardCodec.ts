import type { GridClipboardInput } from "./core";

/** Machine projection of the adopted clipboard representation contract. */
export const clipboardLimits = {
  maxBytes: 8_388_608,
  maxRows: 500,
  maxColumns: 64,
} as const;
export type ClipboardFormat = "csv" | "tsv";
export type ClipboardRepresentations = Readonly<
  Partial<
    Record<
      "text/html" | "text/tab-separated-values" | "text/csv" | "text/plain",
      string
    >
  >
>;
export type ClipboardFailure = {
  readonly kind: "failure";
  readonly reason:
    | "empty_table"
    | "malformed_quotes"
    | "ragged_rows"
    | "too_many_rows"
    | "too_many_columns"
    | "too_large"
    | "unsupported_table";
  readonly message: string;
};
export type ClipboardDecodeResult =
  | GridClipboardInput
  | ClipboardFailure
  | { readonly kind: "noop" };

const messages: Record<ClipboardFailure["reason"], string> = {
  empty_table: "The clipboard table is empty.",
  malformed_quotes: "The clipboard table has invalid quotation marks.",
  ragged_rows: "Every clipboard row must have the same number of cells.",
  too_many_rows: "Paste supports up to 500 rows.",
  too_many_columns: "Paste supports up to 64 columns.",
  too_large: "Clipboard content exceeds the 8 MiB limit.",
  unsupported_table: "This clipboard table format is not supported.",
};
export function clipboardFailure(
  reason: ClipboardFailure["reason"],
): ClipboardFailure {
  return { kind: "failure", reason, message: messages[reason] };
}

/** Count UTF-8 without allocating another copy of potentially hostile text. */
export function clipboardTextWithinLimit(text: string): boolean {
  let bytes = 0;
  for (let i = 0; i < text.length; i++) {
    const code = text.charCodeAt(i);
    if (code < 0x80) bytes++;
    else if (code < 0x800) bytes += 2;
    else if (
      code >= 0xd800 &&
      code <= 0xdbff &&
      i + 1 < text.length &&
      text.charCodeAt(i + 1) >= 0xdc00 &&
      text.charCodeAt(i + 1) <= 0xdfff
    ) {
      bytes += 4;
      i++;
    } else bytes += 3;
    if (bytes > clipboardLimits.maxBytes) return false;
  }
  return true;
}

export function decodeDelimitedClipboard(
  text: string,
  format: ClipboardFormat,
):
  | {
      readonly kind: "decoded";
      readonly values: readonly (readonly string[])[];
    }
  | ClipboardFailure {
  if (!clipboardTextWithinLimit(text)) return clipboardFailure("too_large");
  if (!text.length) return clipboardFailure("empty_table");
  const delimiter = format === "csv" ? "," : "\t";
  const rows: string[][] = [];
  let row: string[] = [];
  let index = 0;
  while (true) {
    let value: string;
    if (text[index] === '"') {
      const start = ++index;
      while (true) {
        if (index === text.length) return clipboardFailure("malformed_quotes");
        if (text[index] !== '"') {
          index++;
          continue;
        }
        if (text[index + 1] === '"') {
          index += 2;
          continue;
        }
        value = text.slice(start, index).replace(/""/g, '"');
        index++;
        break;
      }
    } else {
      const start = index;
      while (
        index < text.length &&
        text[index] !== delimiter &&
        text[index] !== "\r" &&
        text[index] !== "\n"
      ) {
        if (text[index] === '"') return clipboardFailure("malformed_quotes");
        index++;
      }
      value = text.slice(start, index);
    }
    row.push(value);
    if (row.length > clipboardLimits.maxColumns)
      return clipboardFailure("too_many_columns");
    const next = text[index];
    if (next === delimiter) {
      index++;
      continue;
    }
    if (next !== undefined && next !== "\r" && next !== "\n")
      return clipboardFailure("malformed_quotes");
    if (rows.length && row.length !== rows[0]?.length)
      return clipboardFailure("ragged_rows");
    rows.push(row);
    if (rows.length > clipboardLimits.maxRows + 1)
      return clipboardFailure("too_many_rows");
    row = [];
    if (next === undefined) break;
    index += next === "\r" && text[index + 1] === "\n" ? 2 : 1;
    if (index === text.length) break;
  }
  return { kind: "decoded", values: rows };
}

export function encodeClipboardTable(
  values: readonly (readonly string[])[],
  format: ClipboardFormat = "tsv",
): string {
  const delimiter = format === "csv" ? "," : "\t";
  return values
    .map((row) =>
      row
        .map((value) =>
          value === "" || value.includes(delimiter) || /[\r\n"]/u.test(value)
            ? `"${value.replace(/"/g, '""')}"`
            : value,
        )
        .join(delimiter),
    )
    .join("\n");
}

export function clipboardRepresentations(
  data: Pick<DataTransfer, "types" | "getData">,
): ClipboardRepresentations {
  const result: Partial<Record<keyof ClipboardRepresentations, string>> = {};
  for (const type of [
    "text/html",
    "text/tab-separated-values",
    "text/csv",
    "text/plain",
  ] as const) {
    if (data.types.includes(type)) result[type] = data.getData(type);
  }
  return result;
}

const marker = "data-cartulary-clipboard";
const intentMarker = "data-cartulary-clipboard-intent";
const formulaPrefix = /^[=+\-@\t\r]/u;
function escapeHTML(text: string): string {
  return text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/\r/g, "&#13;");
}
export function encodeGridClipboard(
  values: readonly (readonly string[])[],
): ClipboardRepresentations | ClipboardFailure {
  if (!values.length || !values[0]?.length)
    return clipboardFailure("empty_table");
  if (values.length > clipboardLimits.maxRows)
    return clipboardFailure("too_many_rows");
  if (values[0].length > clipboardLimits.maxColumns)
    return clipboardFailure("too_many_columns");
  const scalar = values.length === 1 && values[0]?.length === 1;
  const opening = `<table ${marker}="1" ${intentMarker}="${scalar ? "scalar" : "table"}">`;
  // Count exact export sizes before allocating escaped strings.
  let htmlBytes = opening.length + "</table>".length;
  let plainBytes = values.length - 1;
  for (const row of values) {
    if (row.length !== values[0].length) return clipboardFailure("ragged_rows");
    htmlBytes += 9;
    plainBytes += row.length - 1;
    for (const value of row) {
      const neutralize = formulaPrefix.test(value);
      htmlBytes += 9 + Number(neutralize || value.startsWith("'"));
      plainBytes += Number(neutralize);
      if (!scalar && (value === "" || /[\t\r\n"]/u.test(value)))
        plainBytes += 2;
      for (const char of value) {
        const cp = char.codePointAt(0) ?? 0;
        const bytes = cp < 0x80 ? 1 : cp < 0x800 ? 2 : cp < 0x10000 ? 3 : 4;
        htmlBytes +=
          char === "&" || char === "\r"
            ? 5
            : char === "<" || char === ">"
              ? 4
              : bytes;
        plainBytes += bytes + Number(!scalar && char === '"');
        if (Math.max(htmlBytes, plainBytes) > clipboardLimits.maxBytes)
          return clipboardFailure("too_large");
      }
    }
  }
  const portable = values.map((row) =>
    row.map((value) => (formulaPrefix.test(value) ? `'${value}` : value)),
  );
  const plain = scalar
    ? (portable[0]?.[0] ?? "")
    : encodeClipboardTable(portable);
  const html = `${opening}${values
    .map(
      (row) =>
        `<tr>${row
          .map((value) => {
            const escaped =
              value.startsWith("'") || formulaPrefix.test(value)
                ? `'${value}`
                : value;
            return `<td>${escapeHTML(escaped)}</td>`;
          })
          .join("")}</tr>`,
    )
    .join("")}</table>`;
  if (!clipboardTextWithinLimit(plain) || !clipboardTextWithinLimit(html))
    return clipboardFailure("too_large");
  return { "text/plain": plain, "text/html": html };
}

const cellWrappers = new Set([
  "SPAN",
  "FONT",
  "B",
  "I",
  "U",
  "S",
  "STRONG",
  "EM",
  "SMALL",
  "SUB",
  "SUP",
  "A",
]);
function htmlCellText(cell: Element): string | null {
  let text = "";
  // Iterative traversal avoids call-stack exhaustion from hostile nesting.
  const nodes = Array.from(cell.childNodes).reverse();
  while (nodes.length) {
    const node = nodes.pop();
    if (!node) break;
    if (node.nodeType === Node.TEXT_NODE) {
      text += node.textContent ?? "";
      continue;
    }
    if (node.nodeType === Node.COMMENT_NODE) continue;
    if (!(node instanceof Element)) return null;
    if (node.tagName === "BR") {
      text += "\n";
      continue;
    }
    if (!cellWrappers.has(node.tagName) || node.hasAttribute("hidden"))
      return null;
    for (let index = node.childNodes.length - 1; index >= 0; index--) {
      const child = node.childNodes[index];
      if (child) nodes.push(child);
    }
  }
  return text;
}

function decodeHTML(html: string): ClipboardDecodeResult | null {
  if (!clipboardTextWithinLimit(html)) return clipboardFailure("too_large");
  const template = document.createElement("template");
  template.innerHTML = html;
  const tables = template.content.querySelectorAll("table");
  if (!tables.length)
    return template.content.querySelector(`[${marker}]`)
      ? clipboardFailure("unsupported_table")
      : null;
  if (tables.length !== 1) return clipboardFailure("unsupported_table");
  const table = tables[0];
  if (!table) return clipboardFailure("unsupported_table");
  const marked = table.hasAttribute(marker);
  const intent = table.getAttribute(intentMarker);
  if (
    marked &&
    (table.getAttribute(marker) !== "1" ||
      (intent !== "scalar" && intent !== "table"))
  )
    return clipboardFailure("unsupported_table");
  if (
    table.querySelector(
      "script,style,iframe,img,object,embed,svg,math,input,textarea,select,template",
    )
  )
    return clipboardFailure("unsupported_table");
  // Table structure may carry formatting, but no omitted data-bearing nodes.
  for (const parent of [
    table,
    ...table.querySelectorAll("thead,tbody,tfoot,tr,colgroup"),
  ]) {
    const allowed =
      parent.tagName === "TR"
        ? ["TD", "TH"]
        : parent.tagName === "COLGROUP"
          ? ["COL"]
          : parent.tagName === "TABLE"
            ? ["THEAD", "TBODY", "TFOOT", "TR", "COLGROUP", "COL"]
            : ["TR"];
    for (const node of parent.childNodes) {
      if (node.nodeType === Node.COMMENT_NODE) continue;
      if (node.nodeType === Node.TEXT_NODE && !node.textContent?.trim())
        continue;
      if (!(node instanceof Element) || !allowed.includes(node.tagName))
        return clipboardFailure("unsupported_table");
    }
  }
  const values: string[][] = [];
  for (const row of table.rows) {
    if (values.length >= clipboardLimits.maxRows + 1)
      return clipboardFailure("too_many_rows");
    const cells: string[] = [];
    for (const cell of row.cells) {
      if (cells.length >= clipboardLimits.maxColumns)
        return clipboardFailure("too_many_columns");
      if (cell.colSpan !== 1 || cell.rowSpan !== 1)
        return clipboardFailure("unsupported_table");
      let text = htmlCellText(cell);
      if (text === null) return clipboardFailure("unsupported_table");
      if (marked) {
        if (
          text.startsWith("'") &&
          (text[1] === "'" || formulaPrefix.test(text.slice(1)))
        )
          text = text.slice(1);
        else if (text.startsWith("'") || formulaPrefix.test(text))
          return clipboardFailure("unsupported_table");
      }
      cells.push(text);
    }
    if (!cells.length || (values.length && cells.length !== values[0]?.length))
      return clipboardFailure("ragged_rows");
    values.push(cells);
  }
  if (!values.length) return clipboardFailure("empty_table");
  const scalar = values.length === 1 && values[0]?.length === 1;
  if (marked && (intent === "scalar") !== scalar)
    return clipboardFailure("unsupported_table");
  if (scalar)
    return {
      kind: "scalar",
      value: values[0]?.[0] ?? "",
      rawText: values[0]?.[0] ?? "",
    };
  const rawText = encodeClipboardTable(values);
  if (!clipboardTextWithinLimit(rawText)) return clipboardFailure("too_large");
  return {
    kind: "table",
    format: "tsv",
    rawText,
    values,
    headerMode: marked ? "none" : "auto",
  };
}

export function decodeGridClipboard(
  offered: ClipboardRepresentations,
): ClipboardDecodeResult {
  if (offered["text/html"] !== undefined) {
    const result = decodeHTML(offered["text/html"]);
    if (result) return result;
  }
  for (const [type, format] of [
    ["text/tab-separated-values", "tsv"],
    ["text/csv", "csv"],
    ["text/plain", "tsv"],
  ] as const) {
    const text = offered[type];
    if (text === undefined) continue;
    if (!clipboardTextWithinLimit(text)) return clipboardFailure("too_large");
    if (type === "text/plain" && !text.length) return { kind: "noop" };
    if (type === "text/plain" && !/[\t\r\n]/u.test(text))
      return { kind: "scalar", rawText: text, value: text };
    const result = decodeDelimitedClipboard(text, format);
    if (result.kind === "failure") return result;
    return {
      kind: "table",
      format,
      rawText: text,
      values: result.values,
      headerMode: "auto",
    };
  }
  return { kind: "noop" };
}
