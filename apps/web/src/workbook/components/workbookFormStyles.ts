import type { CSSProperties } from "react";

// Form controls share authored component states; grid sizing remains with the grid.
export const workbookFormInputStyle = {
  boxSizing: "border-box",
  display: "block",
  minWidth: 0,
  width: "100%",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  padding: "var(--ct-component-text-input-padding)",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
} satisfies CSSProperties;
export const workbookFormFieldStackStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  minWidth: 0,
} satisfies CSSProperties;
export const workbookFormFieldsStyle = {
  ...workbookFormFieldStackStyle,
  gap: "var(--ct-spacing-sm)",
} satisfies CSSProperties;
export const workbookFormMessageStyle = { margin: 0 } satisfies CSSProperties;
