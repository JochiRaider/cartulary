export type ClipboardGridDimensions = {
  readonly columnCount: number;
  readonly rowCount: number;
};

export function parseClipboardTable(text: string): string[][] {
  return parseGridClipboardTable(text);
}

export function clipboardGridDimensions(text: string): ClipboardGridDimensions {
  return gridClipboardDimensions(text);
}

export function clipboardTextLooksTabular(text: string): boolean {
  return text.includes("\n") || text.includes("\r") || text.includes("\t");
}

import {
  gridClipboardDimensions,
  parseGridClipboardTable,
} from "@cartulary/grid-adapter";
