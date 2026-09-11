import type { CSSProperties } from "react";
export const observationStack = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  minInlineSize: 0,
  overflowWrap: "anywhere",
} satisfies CSSProperties;
export const observationField = {
  ...observationStack,
  gap: "var(--ct-spacing-xs)",
} satisfies CSSProperties;
export const observationInput = {
  inlineSize: "100%",
  minInlineSize: 0,
  boxSizing: "border-box",
  font: "inherit",
  colorScheme: "dark",
  color: "var(--ct-colors-ink)",
  background: "var(--ct-component-text-input-backgroundColor)",
  border: "var(--ct-component-text-input-border)",
  borderRadius: "var(--ct-component-text-input-rounded)",
  padding: "var(--ct-spacing-xs) var(--ct-spacing-sm)",
} satisfies CSSProperties;
export const observationText = {
  margin: 0,
  whiteSpace: "pre-wrap",
} satisfies CSSProperties;
