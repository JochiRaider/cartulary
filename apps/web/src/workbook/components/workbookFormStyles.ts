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

export const workbookQuietCommandStyle = {
  ...workbookTypography("button"),
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  gap: "var(--ct-spacing-xs)",
  border: "1px solid transparent",
  borderRadius: "var(--ct-component-button-quiet-rounded)",
  background:
    "var(--ct-command-background, var(--ct-component-button-quiet-backgroundColor))",
  color:
    "var(--ct-command-foreground, var(--ct-component-button-quiet-textColor))",
  padding: "var(--ct-component-button-quiet-padding)",
  minBlockSize: "var(--ct-component-button-quiet-minBlockSize)",
  cursor: "pointer",
  whiteSpace: "nowrap",
} satisfies CSSProperties;

// Only quiet commands consume these local state variables. Feature admission is
// still expressed by each native control's disabled/aria-disabled/busy state.
export const workbookCommandStateStyles = `
.cartulary-shell .workbook-query-summary:empty { display: none; }
.cartulary-shell :is(button, summary):hover:not(:disabled):not([aria-disabled="true"]) {
  --ct-command-background: var(--ct-component-button-quiet-hoverBackgroundColor);
  --ct-command-foreground: var(--ct-colors-ink);
}
.cartulary-shell :is(button, summary):is(:disabled, [aria-disabled="true"]) {
  --ct-command-foreground: var(--ct-component-button-quiet-disabledTextColor);
  cursor: not-allowed;
}
.cartulary-shell button[aria-busy="true"] {
  --ct-command-background: var(--ct-component-button-quiet-hoverBackgroundColor);
  cursor: progress;
}
.cartulary-shell :is(button, summary, select):focus-visible {
  outline: var(--ct-component-focus-ring-border);
  outline-offset: var(--ct-component-focus-ring-offset);
}
`;

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
