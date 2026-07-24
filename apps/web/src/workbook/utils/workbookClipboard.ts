import type {
  GridClipboardDimensions,
  GridClipboardInput,
  GridClipboardPasteContract,
} from "@cartulary/grid-adapter";

export function parseClipboardTable(text: string): string[][] {
  const normalized = text.replace(/\r\n/g, "\n").replace(/\r/g, "\n");
  const trimmed = normalized.replace(/\n+$/u, "");
  if (trimmed === "") return [[""]];
  const delimiter = trimmed.includes("\t") ? "\t" : ",";
  const rows: string[][] = [];
  let row: string[] = [];
  let cell = "";
  let quoted = false;
  for (let index = 0; index < trimmed.length; index += 1) {
    const char = trimmed[index];
    if (char === '"') {
      if (quoted && trimmed[index + 1] === '"') {
        cell += '"';
        index += 1;
      } else {
        quoted = !quoted;
      }
      continue;
    }
    if (!quoted && char === delimiter) {
      row.push(cell);
      cell = "";
      continue;
    }
    if (!quoted && char === "\n") {
      row.push(cell);
      rows.push(row);
      row = [];
      cell = "";
      continue;
    }
    cell += char;
  }
  row.push(cell);
  rows.push(row);
  const columnCount = rows.reduce(
    (maximum, candidate) => Math.max(maximum, candidate.length),
    1,
  );
  return rows.map((candidate) => [
    ...candidate,
    ...Array.from({ length: columnCount - candidate.length }, () => ""),
  ]);
}

export function clipboardGridDimensions(text: string): GridClipboardDimensions {
  const rows = parseClipboardTable(text);
  return {
    columnCount: rows[0]?.length ?? 1,
    rowCount: rows.length,
  };
}

export function clipboardTextLooksTabular(text: string): boolean {
  return text.includes("\n") || text.includes("\r") || text.includes("\t");
}

export function decodeWorkbookClipboardInput(
  rawText: string,
): GridClipboardInput {
  if (!clipboardTextLooksTabular(rawText)) {
    return { kind: "scalar", rawText, value: rawText };
  }
  return {
    format: rawText.includes("\t") ? "tsv" : "csv",
    kind: "table",
    rawText,
    values: parseClipboardTable(rawText),
  };
}

export function workbookClipboardPasteContract(
  onPaste: GridClipboardPasteContract["onPaste"],
): GridClipboardPasteContract {
  return { decode: decodeWorkbookClipboardInput, onPaste };
}
