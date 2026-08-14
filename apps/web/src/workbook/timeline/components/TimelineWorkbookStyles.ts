import type { CSSProperties } from "react";
import { workbookSurfaceGridShellStyle } from "../../layout/WorkbookSurfaceLayout";

export const timelineRowGutterWidth = 58;

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

export const inputStyle = {
  boxSizing: "border-box" as const,
  display: "block",
  minWidth: 0,
  width: "100%",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  padding: "0.65rem 0.75rem",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
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
