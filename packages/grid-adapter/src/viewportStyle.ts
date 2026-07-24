import type { CSSProperties } from "react";

import type { GridBlockSizing, GridChrome } from "./core";

export function resolveGridViewportStyle(
  style?: CSSProperties,
  chrome: GridChrome = "sheet",
  blockSizing: GridBlockSizing = "standalone",
): CSSProperties {
  return {
    background: "var(--ct-colors-surface-1)",
    blockSize: blockSizing === "fill" ? "100%" : "min(70vh, 46rem)",
    border: "var(--ct-border-hairline)",
    borderRadius: chrome === "framed" ? "var(--ct-rounded-sm)" : 0,
    boxSizing: blockSizing === "fill" ? "border-box" : undefined,
    color: "var(--ct-colors-ink)",
    minBlockSize: blockSizing === "fill" ? 0 : "18rem",
    minInlineSize: 0,
    overflow: "hidden",
    position: "relative",
    ...(blockSizing === "fill"
      ? { blockSize: "100%", display: "flex", flexDirection: "column" }
      : null),
    ...style,
  };
}
