export type WorkbookChromeMode =
  | "base"
  | "below_supported_minimum"
  | "compact_desktop"
  | "narrow_desktop";

export type WorkbookBlockMode =
  | "base_height"
  | "compact_height"
  | "short_height";

const baseChromeMinWidthCssPx = 1280;
const narrowChromeMinWidthCssPx = 1024;
const compactChromeMinWidthCssPx = 768;
const baseBlockMinHeightCssPx = 720;
const compactBlockMinHeightCssPx = 640;

function finiteCssPx(value: number): number {
  return Number.isFinite(value) ? value : 0;
}

export function selectWorkbookChromeMode(
  widthCssPx: number,
): WorkbookChromeMode {
  const width = finiteCssPx(widthCssPx);
  if (width >= baseChromeMinWidthCssPx) {
    return "base";
  }
  if (width >= narrowChromeMinWidthCssPx) {
    return "narrow_desktop";
  }
  if (width >= compactChromeMinWidthCssPx) {
    return "compact_desktop";
  }
  return "below_supported_minimum";
}

export function selectWorkbookBlockMode(
  heightCssPx: number,
): WorkbookBlockMode {
  const height = finiteCssPx(heightCssPx);
  if (height >= baseBlockMinHeightCssPx) {
    return "base_height";
  }
  if (height >= compactBlockMinHeightCssPx) {
    return "compact_height";
  }
  return "short_height";
}
