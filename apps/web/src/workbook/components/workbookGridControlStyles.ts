import type { CSSProperties } from "react";

export const menuFrameStyle = {
  position: "relative" as const,
  display: "inline-flex",
};

export const fixedMenuFrameStyle = {
  ...menuFrameStyle,
  flex: "0 0 auto",
  minInlineSize: "max-content",
};

export const controlButtonStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  gap: "0.3rem",
  borderRadius: "var(--ct-rounded-xs)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-ink)",
  padding: "0.28rem 0.5rem",
  font: "inherit",
  cursor: "pointer",
  minBlockSize: "1.8rem",
  whiteSpace: "nowrap" as const,
};

export const immutableControlLabelStyle = {
  flex: "0 0 auto",
  whiteSpace: "nowrap" as const,
};

export const dynamicControlValueStyle = {
  display: "block",
  minInlineSize: 0,
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
};

export const inputStyle = {
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  padding: "0.35rem 0.5rem",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
  boxSizing: "border-box" as const,
  minInlineSize: "12rem",
  minBlockSize: "1.9rem",
};

export const selectStyle = {
  ...inputStyle,
  minInlineSize: "9rem",
};

export const stackedLabelStyle = {
  display: "grid",
  gap: "0.3rem",
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.82rem",
};

export const menuStyle = {
  position: "absolute" as const,
  zIndex: 20,
  insetBlockStart: "calc(100% + 0.35rem)",
  insetInlineStart: 0,
  display: "grid",
  gap: "0.2rem",
  inlineSize: "min(var(--ct-layout-viewBarOverlayMaxInlineSize), 92vw)",
  maxBlockSize: "70dvh",
  overflowY: "auto" as const,
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-md)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-popover)",
  padding: "0.45rem",
};

export const menuItemStyle = {
  border: 0,
  borderRadius: "var(--ct-rounded-xs)",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  cursor: "pointer",
  font: "inherit",
  padding: "0.45rem 0.5rem",
  textAlign: "left" as const,
  overflowWrap: "anywhere" as const,
};

export const menuItemSelectedStyle = {
  background: "var(--ct-colors-surface-3)",
  color: "var(--ct-colors-ink)",
  fontWeight: 700,
};

export const filterValidationStyle = {
  margin: 0,
  color: "var(--ct-colors-semantic-conflict)",
  fontSize: "0.8rem",
  fontWeight: 700,
};

export const queryListStyle = {
  display: "grid",
  gap: "0.25rem",
  borderBlockStart: "var(--ct-border-hairline)",
  paddingBlockStart: "0.45rem",
} satisfies CSSProperties;

export const queryListButtonStyle = {
  ...menuItemStyle,
  maxInlineSize: "100%",
  whiteSpace: "normal" as const,
  overflowWrap: "anywhere" as const,
} satisfies CSSProperties;

export const secondaryButtonStyle = {
  ...controlButtonStyle,
  background: "transparent",
};

export const primaryButtonStyle = {
  ...controlButtonStyle,
  borderColor: "var(--ct-colors-accent-active)",
  background: "var(--ct-colors-accent)",
  color: "var(--ct-colors-on-accent)",
  fontWeight: 700,
};

export const clearButtonStyle = {
  ...controlButtonStyle,
  flex: "0 0 auto",
  borderColor: "transparent",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
};
