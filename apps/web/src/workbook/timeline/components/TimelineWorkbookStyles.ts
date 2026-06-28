import type { CSSProperties } from "react";
import {
  workbookSurfaceGridShellStyle,
  workbookSurfaceOverlayPanelStyle,
} from "../../components/WorkbookSurfaceFrame";

export const timelineRowGutterWidth = 58;

export const panelStyle = {
  boxSizing: "border-box" as const,
  width: "100%",
  blockSize: "100%",
  minBlockSize: 0,
  margin: 0,
  padding: 0,
  borderRadius: 0,
  background: "var(--ct-colors-canvas)",
  boxShadow: "none",
  border: 0,
};

export const eyebrowStyle = {
  margin: 0,
  fontSize: "0.78rem",
  letterSpacing: "0.12em",
  textTransform: "uppercase" as const,
  color: "var(--ct-colors-accent)",
};

export const headlineStyle = {
  margin: "0.35rem 0 0.5rem",
  fontSize: "2rem",
  lineHeight: 1.1,
};

export const bodyStyle = {
  margin: 0,
  lineHeight: 1.5,
  color: "var(--ct-colors-ink-muted)",
};

export const timelineGridShellStyle = {
  ...workbookSurfaceGridShellStyle,
} satisfies CSSProperties;

export const actionButtonStyle = {
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  padding: "0.55rem 0.9rem",
  font: "inherit",
  cursor: "pointer",
};

export const secondaryActionButtonStyle = {
  ...actionButtonStyle,
  background: "var(--ct-colors-surface-3)",
};

export const inlineButtonRowStyle = {
  display: "flex",
  flexWrap: "wrap" as const,
  gap: "0.5rem",
};

export const labelStyle = {
  display: "grid",
  gap: "0.4rem",
  fontSize: "0.95rem",
  color: "var(--ct-colors-ink-muted)",
};

export const inspectorSectionStyle = {
  display: "grid",
  gap: "0.75rem",
  marginBottom: "1rem",
};

export const sectionTitleStyle = {
  margin: 0,
  fontSize: "1rem",
};

export const inspectorActionStackStyle = {
  display: "grid",
  gap: "0.75rem",
};

const noticeCardStyle = {
  position: "absolute" as const,
  zIndex: 7,
  insetBlockStart:
    "calc(var(--ct-layout-viewBarHeight) + var(--ct-spacing-sm))",
  insetInlineEnd: "var(--ct-spacing-sm)",
  maxInlineSize: "min(36rem, calc(100% - var(--ct-spacing-xl)))",
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  padding: "0.85rem 1rem",
  display: "grid",
  gap: "0.5rem",
  boxShadow: "var(--ct-elevation-popover)",
};

export const timelineNoticeOverlayStyle = {
  ...noticeCardStyle,
  ...workbookSurfaceOverlayPanelStyle,
  insetBlockStart: "var(--ct-spacing-sm)",
  insetInlineEnd: "auto",
} satisfies CSSProperties;
