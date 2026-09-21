import type { CSSProperties } from "react";

export function workbookTypography(
  role:
    | "ui"
    | "surface-title"
    | "section-heading"
    | "metadata"
    | "compact-metadata"
    | "button"
    | "mono",
): CSSProperties {
  return {
    fontFamily: `var(--ct-typography-${role}-fontFamily)`,
    fontSize: `var(--ct-typography-${role}-fontSize)`,
    fontWeight: `var(--ct-typography-${role}-fontWeight)`,
    lineHeight: `var(--ct-typography-${role}-lineHeight)`,
    letterSpacing: `var(--ct-typography-${role}-letterSpacing)`,
  };
}

// Form controls share authored component states; grid sizing remains with the grid.
export const workbookFormInputStyle = {
  ...workbookTypography("ui"),
  boxSizing: "border-box",
  display: "block",
  minWidth: 0,
  width: "100%",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  padding: "var(--ct-component-text-input-padding)",
  color: "var(--ct-component-text-input-textColor)",
} satisfies CSSProperties;
export const workbookFormFieldStackStyle = {
  ...workbookTypography("ui"),
  overflowWrap: "anywhere",
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  minWidth: 0,
} satisfies CSSProperties;
export const workbookFormFieldsStyle = {
  ...workbookFormFieldStackStyle,
  gap: "var(--ct-spacing-sm)",
} satisfies CSSProperties;
export const workbookFormMessageStyle = { margin: 0 } satisfies CSSProperties;

export const workbookFormButtonStyle = {
  ...workbookTypography("button"),
  maxInlineSize: "100%",
  overflowWrap: "anywhere",
  border: "var(--ct-component-button-secondary-border)",
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  padding: "var(--ct-component-button-secondary-padding)",
} satisfies CSSProperties;

// Grid action controls share button states while fitting the selected row density.
export const workbookGridActionButtonStyle = {
  ...workbookFormButtonStyle,
  fontSize: "inherit",
  lineHeight: "inherit",
  padding: "0 var(--ct-spacing-xs)",
  minBlockSize: 0,
  maxBlockSize: "100%",
  whiteSpace: "nowrap",
} satisfies CSSProperties;

export const workbookFormGroupStyle = {
  ...workbookFormFieldsStyle,
  border: 0,
  padding: 0,
  margin: 0,
} satisfies CSSProperties;
export const workbookFormActionsStyle = {
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-sm)",
  alignItems: "start",
} satisfies CSSProperties;
export const workbookFormHeadingStyle = {
  ...workbookTypography("section-heading"),
  margin: 0,
  paddingBlockEnd: "var(--ct-spacing-xs)",
} satisfies CSSProperties;
export const workbookFormPreservedTextStyle = {
  ...workbookTypography("ui"),
  margin: 0,
  whiteSpace: "pre-wrap",
  overflowWrap: "anywhere",
} satisfies CSSProperties;
