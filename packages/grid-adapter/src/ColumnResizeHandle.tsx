import { useEffect, useRef } from "react";
import { normalizeMeasuredColumnWidth } from "./columnSizing";
import type { GridColumnSizingIntent } from "./core";
import { elementCssScale } from "./viewportGeometry";

export function ColumnResizeHandle({
  fieldKey,
  label,
  minWidth,
  maxWidth,
  onIntent,
  onStart,
}: {
  readonly fieldKey: string;
  readonly label: string;
  readonly minWidth: number | undefined;
  readonly maxWidth: number | undefined;
  readonly onIntent: (intent: GridColumnSizingIntent) => void;
  readonly onStart: () => void;
}) {
  const drag = useRef<{
    startX: number;
    width: number;
    scale: number;
    direction: number;
  } | null>(null);
  useEffect(
    () => () => {
      drag.current = null;
    },
    [],
  );
  return (
    <button
      type="button"
      tabIndex={-1}
      aria-hidden="true"
      data-grid-column-resize={fieldKey}
      data-grid-sizing-exclude="true"
      className="cartulary-grid-column-resize"
      title={`Resize ${label}; double-click to fit visible content`}
      onClick={(event) => {
        event.preventDefault();
        event.stopPropagation();
      }}
      onDoubleClick={(event) => {
        event.preventDefault();
        event.stopPropagation();
        onIntent({ kind: "fit_visible", fieldKey });
      }}
      onMouseDown={(event) => {
        event.preventDefault();
        event.stopPropagation();
      }}
      onPointerDown={(event) => {
        if (event.button !== 0) return;
        event.preventDefault();
        event.stopPropagation();
        const header = event.currentTarget.closest<HTMLElement>(
          '[role="columnheader"]',
        );
        if (!header) return;
        const scale = elementCssScale(header);
        drag.current = {
          startX: event.clientX,
          width: header.getBoundingClientRect().width / scale,
          scale,
          direction: getComputedStyle(header).direction === "rtl" ? -1 : 1,
        };
        onStart();
        event.currentTarget.setPointerCapture(event.pointerId);
      }}
      onPointerMove={(event) => {
        const current = drag.current;
        if (!current) return;
        event.preventDefault();
        event.stopPropagation();
        const widthPx = normalizeMeasuredColumnWidth(
          current.width +
            (current.direction * (event.clientX - current.startX)) /
              current.scale,
          { fieldKey, minWidth, maxWidth },
        );
        if (widthPx !== null)
          onIntent({ kind: "set_width", fieldKey, widthPx });
      }}
      onPointerUp={(event) => {
        drag.current = null;
        event.currentTarget.releasePointerCapture(event.pointerId);
      }}
      onLostPointerCapture={() => {
        drag.current = null;
      }}
    />
  );
}
