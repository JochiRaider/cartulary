import type { CSSProperties } from "react";
export const buttonRowStyle: CSSProperties = {
  display: "flex",
  flexWrap: "wrap",
  gap: "0.75rem",
  marginTop: "1rem",
};

export const buttonStyle: CSSProperties = {
  padding: "var(--ct-component-button-primary-padding)",
  borderRadius: "var(--ct-component-button-primary-rounded)",
  border: "none",
  background: "var(--ct-component-button-primary-backgroundColor)",
  color: "var(--ct-component-button-primary-textColor)",
  fontWeight: 600,
  cursor: "pointer",
};

export const cardHeaderStyle: CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  gap: "1rem",
  alignItems: "center",
  marginBottom: "1rem",
};

export const cardStyle: CSSProperties = {
  padding: "1.4rem",
  borderRadius: "var(--ct-rounded-lg)",
  background: "var(--ct-colors-surface-2)",
  border: "var(--ct-border-hairline)",
  color: "var(--ct-colors-ink)",
  minWidth: 0,
  boxSizing: "border-box",
};

export const detailGridStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(12rem, 1fr))",
  gap: "0.85rem",
  marginTop: "1rem",
  minWidth: 0,
};

export const errorStyle: CSSProperties = {
  margin: "0.25rem 0 0",
  minHeight: "1.25rem",
  color: "var(--ct-colors-semantic-conflict)",
  fontWeight: 600,
};

export const formGridStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "minmax(0, 1fr)",
  gap: "0.75rem",
  marginTop: "1rem",
  minWidth: 0,
};

export const inputStyle: CSSProperties = {
  boxSizing: "border-box",
  width: "100%",
  maxWidth: "100%",
  minWidth: 0,
  marginTop: "0.35rem",
  padding: "var(--ct-component-text-input-padding)",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  color: "var(--ct-component-text-input-textColor)",
  fontSize: "0.95rem",
};

export const labelBlockStyle: CSSProperties = {
  fontSize: "0.84rem",
  fontWeight: 600,
  color: "var(--ct-colors-ink-muted)",
  minWidth: 0,
};

export const labelStyle: CSSProperties = {
  display: "block",
  fontSize: "0.72rem",
  letterSpacing: "0.12em",
  textTransform: "uppercase",
  color: "var(--ct-colors-ink-subtle)",
  marginBottom: "0.35rem",
};

export const monoTextStyle: CSSProperties = {
  fontFamily: "var(--ct-typography-mono-fontFamily)",
  overflowWrap: "anywhere",
  minWidth: 0,
};

export const secondaryButtonStyle: CSSProperties = {
  ...buttonStyle,
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  border: "var(--ct-component-button-secondary-border)",
};

export const sectionTitleStyle: CSSProperties = {
  margin: "0.35rem 0 0",
  fontSize: "1.15rem",
};

export const statusTextStyle: CSSProperties = {
  margin: "1rem 0 0",
  minHeight: "1.5rem",
  color: "var(--ct-colors-ink-muted)",
};

export const subsectionTitleStyle: CSSProperties = {
  margin: 0,
  fontWeight: 700,
  color: "var(--ct-colors-ink)",
};
