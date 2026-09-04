import { workbookLayoutMetrics } from "@cartulary/ui-contracts";

export type WorkbookChromeMode =
  | "base"
  | "below_supported_minimum"
  | "compact_desktop"
  | "narrow_desktop";

export type WorkbookBlockMode =
  | "base_height"
  | "compact_height"
  | "short_height";

function finiteCssPx(value: number): number {
  return Number.isFinite(value) ? value : 0;
}

export function selectWorkbookChromeMode(
  widthCssPx: number,
): WorkbookChromeMode {
  const width = finiteCssPx(widthCssPx);
  const metrics = workbookLayoutMetrics(width);
  if (width >= metrics.baseMinWidthCssPx) {
    return "base";
  }
  if (width >= metrics.narrowMinWidthCssPx) {
    return "narrow_desktop";
  }
  if (width >= metrics.compactMinWidthCssPx) {
    return "compact_desktop";
  }
  return "below_supported_minimum";
}

export function selectWorkbookBlockMode(
  heightCssPx: number,
): WorkbookBlockMode {
  const height = finiteCssPx(heightCssPx);
  const metrics = workbookLayoutMetrics(0);
  if (height >= metrics.baseMinHeightCssPx) {
    return "base_height";
  }
  if (height >= metrics.compactMinHeightCssPx) {
    return "compact_height";
  }
  return "short_height";
}

export function workbookQueryChipCapacity(
  chromeMode: WorkbookChromeMode,
): number {
  if (chromeMode === "base") return 3;
  if (chromeMode === "narrow_desktop") return 2;
  return 0;
}
