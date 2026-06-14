export type ClipboardGridDimensions = {
  readonly columnCount: number;
  readonly rowCount: number;
};

export function parseClipboardTable(text: string): string[][] {
  const normalized = text.replace(/\r\n/g, "\n").replace(/\r/g, "\n");
  const trimmed = normalized.replace(/\n+$/, "");
  if (trimmed === "") {
    return [[""]];
  }
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
  return rows;
}

export function clipboardGridDimensions(text: string): ClipboardGridDimensions {
  const rows = parseClipboardTable(text);
  const columnCount = rows.reduce((max, row) => Math.max(max, row.length), 1);
  return {
    columnCount,
    rowCount: Math.max(1, rows.length),
  };
}

export function clipboardTextLooksTabular(text: string): boolean {
  return text.includes("\n") || text.includes("\r") || text.includes("\t");
}
