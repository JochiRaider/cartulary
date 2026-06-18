import type {
  CSSProperties,
  KeyboardEventHandler,
  MouseEventHandler,
  ReactNode,
} from "react";
import { statusStripStyle } from "../utils/workbookStyles";
import { WorkbookShellSlotRegion } from "./WorkbookShellSlots";

export function WorkbookSurfaceFrame({
  header,
  inspector,
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
  readonly header?: ReactNode | undefined;
  readonly inspector?: ReactNode | undefined;
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
  return (
    <section data-testid={testId} style={workbookSurfaceFrameStyle}>
      {header === undefined ? null : (
        <div style={surfaceHeaderStyle}>{header}</div>
      )}
      <WorkbookShellSlotRegion
        slot="view-bar"
        style={workbookSurfaceViewBarStyle}
        viewSchemaId={viewSchemaId}
      >
        {viewBar}
      </WorkbookShellSlotRegion>
      <section
        aria-label={workAreaAriaLabel}
        style={workbookSurfaceWorkAreaStyle}
        onContextMenu={onWorkAreaContextMenu}
        onKeyDown={onWorkAreaKeyDown}
      >
        <WorkbookShellSlotRegion
          slot="primary-grid"
          style={workbookSurfacePrimaryGridSlotStyle}
          viewSchemaId={viewSchemaId}
        >
          {primaryGrid}
        </WorkbookShellSlotRegion>
        {workAreaOverlays}
        {inspector === undefined ? null : (
          <WorkbookShellSlotRegion
            slot="inspector"
            style={workbookSurfaceInspectorSlotStyle}
            viewSchemaId={viewSchemaId}
          >
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
    "auto var(--ct-layout-viewBarHeight) minmax(0, 1fr) var(--ct-layout-statusStripHeight)",
  blockSize: "100%",
  minBlockSize: 0,
  minHeight: 0,
  minWidth: 0,
  overflow: "hidden",
  background: "var(--ct-colors-canvas)",
} satisfies CSSProperties;

const surfaceHeaderStyle = {
  gridRow: 1,
  minBlockSize: 0,
  minWidth: 0,
  overflow: "hidden",
} satisfies CSSProperties;

const workbookSurfaceViewBarStyle = {
  position: "relative" as const,
  zIndex: 4,
  gridRow: 2,
  display: "block",
  minBlockSize: "var(--ct-layout-viewBarHeight)",
  minWidth: 0,
  margin: 0,
  padding: 0,
  border: 0,
  borderRadius: 0,
  background: "var(--ct-colors-surface-1)",
  overflowX: "visible" as const,
} satisfies CSSProperties;

const workbookSurfaceWorkAreaStyle = {
  position: "relative" as const,
  gridRow: 3,
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

const workbookSurfaceStatusStripStyle = {
  ...statusStripStyle,
  gridRow: 4,
  minBlockSize: "var(--ct-layout-statusStripHeight)",
} satisfies CSSProperties;
