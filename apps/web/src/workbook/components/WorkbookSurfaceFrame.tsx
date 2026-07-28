import type {
  CSSProperties,
  KeyboardEvent,
  KeyboardEventHandler,
  MouseEventHandler,
  PointerEventHandler,
  ReactNode,
} from "react";
import { useLayoutEffect, useRef, useState } from "react";
import type { WorkbookChromeMode } from "../models/workbookResponsiveLayout";
import { statusStripStyle } from "../utils/workbookStyles";
import { WorkbookShellSlotRegion } from "./WorkbookShellSlots";

const inspectorDefaultWidthCssPx = 420;
const inspectorMinWidthCssPx = 360;
const inspectorMaxWidthCssPx = 560;
const inspectorKeyboardStepCssPx = 16;

export function WorkbookSurfaceFrame({
  chromeMode = "base",
  inspector,
  onRequestInspectorClose,
  primaryGrid,
  statusStrip,
  testId,
  viewBar,
  viewSchemaId,
  workAreaAriaLabel = "Workbook work area",
  workAreaOverlays,
  onWorkAreaContextMenu,
  onWorkAreaKeyDown,
}: {
  readonly chromeMode?: WorkbookChromeMode | undefined;
  readonly inspector?: ReactNode | undefined;
  readonly onRequestInspectorClose?: (() => void) | undefined;
  readonly primaryGrid: ReactNode;
  readonly statusStrip: ReactNode;
  readonly testId?: string | undefined;
  readonly viewBar: ReactNode;
  readonly viewSchemaId: string;
  readonly workAreaAriaLabel?: string | undefined;
  readonly workAreaOverlays?: ReactNode | undefined;
  readonly onWorkAreaContextMenu?: MouseEventHandler<HTMLElement> | undefined;
  readonly onWorkAreaKeyDown?: KeyboardEventHandler<HTMLElement> | undefined;
}) {
  const inspectorOpen = inspector !== undefined;
  const inspectorIsAdjacent = chromeMode === "base";
  const backgroundIsInert = inspectorOpen && !inspectorIsAdjacent;
  const [inspectorWidth, setInspectorWidth] = useState(
    inspectorDefaultWidthCssPx,
  );
  const pointerResizeRef = useRef<{
    readonly startClientX: number;
    readonly startWidth: number;
  } | null>(null);
  const inspectorWasOpenRef = useRef(false);
  const returnFocusRef = useRef<HTMLElement | null>(null);

  useLayoutEffect(() => {
    if (inspectorOpen && !inspectorWasOpenRef.current) {
      returnFocusRef.current =
        document.activeElement instanceof HTMLElement
          ? document.activeElement
          : null;
    }
    if (!inspectorOpen && inspectorWasOpenRef.current) {
      const returnFocus = returnFocusRef.current;
      window.requestAnimationFrame(() => {
        if (returnFocus?.isConnected) {
          returnFocus.focus({ preventScroll: true });
        }
      });
      returnFocusRef.current = null;
    }
    inspectorWasOpenRef.current = inspectorOpen;
  }, [inspectorOpen]);

  const clampInspectorWidth = (width: number) =>
    Math.min(inspectorMaxWidthCssPx, Math.max(inspectorMinWidthCssPx, width));
  const resizeInspectorFromKeyboard = (event: KeyboardEvent<HTMLElement>) => {
    let nextWidth: number | null = null;
    if (event.key === "ArrowLeft") {
      nextWidth = inspectorWidth + inspectorKeyboardStepCssPx;
    } else if (event.key === "ArrowRight") {
      nextWidth = inspectorWidth - inspectorKeyboardStepCssPx;
    } else if (event.key === "Home") {
      nextWidth = inspectorMinWidthCssPx;
    } else if (event.key === "End") {
      nextWidth = inspectorMaxWidthCssPx;
    }
    if (nextWidth === null) return;
    event.preventDefault();
    setInspectorWidth(clampInspectorWidth(nextWidth));
  };
  const beginPointerResize: PointerEventHandler<HTMLElement> = (event) => {
    pointerResizeRef.current = {
      startClientX: event.clientX,
      startWidth: inspectorWidth,
    };
    event.currentTarget.setPointerCapture(event.pointerId);
    event.preventDefault();
  };
  const continuePointerResize: PointerEventHandler<HTMLElement> = (event) => {
    const resize = pointerResizeRef.current;
    if (resize === null) return;
    setInspectorWidth(
      clampInspectorWidth(
        resize.startWidth + resize.startClientX - event.clientX,
      ),
    );
  };
  const finishPointerResize: PointerEventHandler<HTMLElement> = (event) => {
    pointerResizeRef.current = null;
    if (event.currentTarget.hasPointerCapture(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId);
    }
  };
  const closeInspectorFromEscape: KeyboardEventHandler<HTMLElement> = (
    event,
  ) => {
    if (
      event.key === "Escape" &&
      !event.defaultPrevented &&
      inspectorOpen &&
      onRequestInspectorClose
    ) {
      event.preventDefault();
      onRequestInspectorClose();
    }
  };
  const workAreaStyle = {
    ...workbookSurfaceWorkAreaStyle,
    gridTemplateColumns:
      inspectorOpen && inspectorIsAdjacent
        ? `minmax(0, 1fr) ${inspectorWidth}px`
        : "minmax(0, 1fr)",
  } satisfies CSSProperties;
  const inspectorSlotStyle = {
    ...workbookSurfaceInspectorSlotStyle,
    ...(inspectorIsAdjacent
      ? {
          position: "relative" as const,
          gridArea: "1 / 2",
          inlineSize: `${inspectorWidth}px`,
          minInlineSize: `${inspectorMinWidthCssPx}px`,
          maxInlineSize: `${inspectorMaxWidthCssPx}px`,
          boxShadow: "none",
        }
      : chromeMode === "narrow_desktop"
        ? workbookSurfaceNarrowInspectorSlotStyle
        : workbookSurfaceCompactInspectorSlotStyle),
  } satisfies CSSProperties;

  return (
    <section
      aria-label={`${workAreaAriaLabel} frame`}
      data-inspector-layout={
        inspectorOpen
          ? inspectorIsAdjacent
            ? "adjacent"
            : chromeMode === "narrow_desktop"
              ? "right_overlay"
              : "full_overlay"
          : "closed"
      }
      data-testid={testId}
      data-workbook-responsive-band={chromeMode}
      style={workbookSurfaceFrameStyle}
      onKeyDown={closeInspectorFromEscape}
    >
      <WorkbookShellSlotRegion
        slot="view-bar"
        style={workbookSurfaceViewBarStyle}
        viewSchemaId={viewSchemaId}
      >
        {viewBar}
      </WorkbookShellSlotRegion>
      <section
        aria-label={workAreaAriaLabel}
        style={workAreaStyle}
        onContextMenu={onWorkAreaContextMenu}
        onKeyDown={onWorkAreaKeyDown}
      >
        <WorkbookShellSlotRegion
          inert={backgroundIsInert}
          slot="primary-grid"
          style={workbookSurfacePrimaryGridSlotStyle}
          viewSchemaId={viewSchemaId}
        >
          {primaryGrid}
        </WorkbookShellSlotRegion>
        <div
          aria-hidden={backgroundIsInert || undefined}
          inert={backgroundIsInert || undefined}
          style={workbookSurfaceOverlayLayerStyle}
        >
          {workAreaOverlays}
        </div>
        {inspector === undefined ? null : (
          <WorkbookShellSlotRegion
            slot="inspector"
            style={inspectorSlotStyle}
            viewSchemaId={viewSchemaId}
          >
            {inspectorIsAdjacent ? (
              <hr
                aria-label="Resize inspector"
                aria-orientation="vertical"
                aria-valuemax={inspectorMaxWidthCssPx}
                aria-valuemin={inspectorMinWidthCssPx}
                aria-valuenow={inspectorWidth}
                aria-valuetext={`${inspectorWidth} pixels`}
                style={workbookSurfaceInspectorSeparatorStyle}
                tabIndex={0}
                onKeyDown={resizeInspectorFromKeyboard}
                onPointerCancel={finishPointerResize}
                onPointerDown={beginPointerResize}
                onPointerMove={continuePointerResize}
                onPointerUp={finishPointerResize}
              />
            ) : null}
            {inspector}
          </WorkbookShellSlotRegion>
        )}
      </section>
      <WorkbookShellSlotRegion
        slot="status-strip"
        style={workbookSurfaceStatusStripStyle}
        viewSchemaId={viewSchemaId}
      >
        {statusStrip}
      </WorkbookShellSlotRegion>
    </section>
  );
}

export const workbookSurfaceGridShellStyle = {
  boxSizing: "border-box" as const,
  inlineSize: "100%",
  blockSize: "100%",
  minBlockSize: 0,
  minWidth: 0,
  overflow: "hidden",
  overflowAnchor: "none" as const,
  borderRadius: 0,
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
} satisfies CSSProperties;

export const workbookSurfaceInspectorPanelStyle = {
  boxSizing: "border-box" as const,
  inlineSize: "100%",
  blockSize: "100%",
  minBlockSize: 0,
  overflow: "auto",
  overflowAnchor: "none" as const,
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-component-inspector-border)",
  background: "var(--ct-component-inspector-backgroundColor)",
  color: "var(--ct-component-inspector-textColor)",
  padding: "var(--ct-spacing-panel-padding)",
} satisfies CSSProperties;

export const workbookSurfaceOverlayPanelStyle = {
  position: "absolute" as const,
  zIndex: 7,
  top: "var(--ct-spacing-sm)",
  left: "var(--ct-spacing-sm)",
  inlineSize: "min(34rem, calc(100% - var(--ct-spacing-xl)))",
  maxBlockSize: "min(28rem, calc(100% - var(--ct-spacing-xl)))",
  overflow: "auto",
  boxSizing: "border-box" as const,
} satisfies CSSProperties;

const workbookSurfaceFrameStyle = {
  position: "relative" as const,
  display: "grid",
  gridTemplateRows:
    "var(--ct-layout-viewBarHeight) minmax(0, 1fr) var(--ct-layout-statusStripHeight)",
  blockSize: "100%",
  minBlockSize: 0,
  minHeight: 0,
  minWidth: 0,
  overflow: "hidden",
  background: "var(--ct-colors-canvas)",
} satisfies CSSProperties;

const workbookSurfaceViewBarStyle = {
  position: "relative" as const,
  zIndex: 9,
  gridRow: 1,
  display: "block",
  minBlockSize: "var(--ct-layout-viewBarHeight)",
  minWidth: 0,
  margin: 0,
  padding: 0,
  border: 0,
  borderRadius: 0,
  background: "var(--ct-colors-surface-1)",
  overflow: "visible",
} satisfies CSSProperties;

const workbookSurfaceWorkAreaStyle = {
  position: "relative" as const,
  gridRow: 2,
  display: "grid",
  gridTemplateRows: "minmax(0, 1fr)",
  inlineSize: "100%",
  blockSize: "100%",
  minBlockSize: 0,
  minHeight: 0,
  minWidth: 0,
  overflow: "hidden",
} satisfies CSSProperties;

const workbookSurfacePrimaryGridSlotStyle = {
  gridArea: "1 / 1",
  inlineSize: "100%",
  blockSize: "100%",
  minBlockSize: 0,
  minHeight: 0,
  minWidth: 0,
  overflow: "hidden",
} satisfies CSSProperties;

const workbookSurfaceOverlayLayerStyle = {
  display: "contents",
} satisfies CSSProperties;

const workbookSurfaceInspectorSlotStyle = {
  position: "absolute" as const,
  zIndex: 8,
  insetBlock: 0,
  insetInlineEnd: 0,
  inlineSize:
    "min(var(--ct-layout-inspectorDefaultWidth), calc(100% - var(--ct-spacing-xl)))",
  minInlineSize:
    "min(var(--ct-layout-inspectorMinWidth), calc(100% - var(--ct-spacing-xl)))",
  maxInlineSize: "var(--ct-layout-inspectorMaxWidth)",
  minBlockSize: 0,
  overflow: "hidden",
  boxShadow: "var(--ct-elevation-drawer)",
  boxSizing: "border-box" as const,
} satisfies CSSProperties;

const workbookSurfaceNarrowInspectorSlotStyle = {
  position: "absolute" as const,
  gridArea: "1 / 1",
  insetBlock: 0,
  insetInlineEnd: 0,
  inlineSize:
    "min(var(--ct-layout-inspectorDefaultWidth), calc(100% - var(--ct-spacing-xl)))",
  minInlineSize:
    "min(var(--ct-layout-inspectorMinWidth), calc(100% - var(--ct-spacing-xl)))",
  maxInlineSize: "var(--ct-layout-inspectorMaxWidth)",
  boxShadow: "var(--ct-elevation-drawer)",
} satisfies CSSProperties;

const workbookSurfaceCompactInspectorSlotStyle = {
  position: "absolute" as const,
  gridArea: "1 / 1",
  inset: 0,
  inlineSize: "100%",
  minInlineSize: 0,
  maxInlineSize: "none",
  boxShadow: "none",
} satisfies CSSProperties;

const workbookSurfaceInspectorSeparatorStyle = {
  position: "absolute" as const,
  zIndex: 2,
  insetBlock: 0,
  insetInlineStart: "-4px",
  inlineSize: "8px",
  blockSize: "100%",
  border: 0,
  borderInlineStart: "var(--ct-border-hairline)",
  background: "transparent",
  cursor: "col-resize",
  touchAction: "none",
} satisfies CSSProperties;

const workbookSurfaceStatusStripStyle = {
  ...statusStripStyle,
  gridRow: 3,
  minBlockSize: "var(--ct-layout-statusStripHeight)",
} satisfies CSSProperties;
