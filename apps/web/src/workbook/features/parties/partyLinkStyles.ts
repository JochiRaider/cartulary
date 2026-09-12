import type { CSSProperties } from "react";

export const partyInputStyle = {
  boxSizing: "border-box",
  display: "block",
  minWidth: 0,
  width: "100%",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  padding: "0.65rem 0.75rem",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
} satisfies CSSProperties;

export const partyStackStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  minWidth: 0,
  margin: 0,
} satisfies CSSProperties;
