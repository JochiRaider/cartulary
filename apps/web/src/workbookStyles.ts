import type { CSSProperties } from "react";

export const visuallyHiddenStyle = {
  position: "absolute",
  inlineSize: 1,
  blockSize: 1,
  overflow: "hidden",
  clipPath: "inset(50%)",
} satisfies CSSProperties;

export const focusableCellStyle = {
  display: "block",
  lineHeight: "1.25rem",
  minHeight: "1.25rem",
  minWidth: "100%",
  maxWidth: "100%",
  outlineOffset: "2px",
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap",
} satisfies CSSProperties;

export function statusIconStyle(saveState: string): CSSProperties {
  return {
    display: "inline-block",
    inlineSize: "0.55rem",
    blockSize: "0.55rem",
    borderRadius: "var(--ct-rounded-pill)",
    background:
      saveState === "Conflict"
        ? "var(--ct-colors-semantic-conflict)"
        : saveState === "Syncing"
          ? "var(--ct-colors-semantic-caution)"
          : "var(--ct-colors-semantic-success)",
  };
}

export const statusStripStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.8rem",
  minHeight: "var(--ct-layout-statusStripHeight)",
  padding: "0 1rem",
  borderTop: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.82rem",
} satisfies CSSProperties;

export const statusStripItemStyle = {
  display: "inline-flex",
  alignItems: "center",
  gap: "0.35rem",
  whiteSpace: "nowrap",
} satisfies CSSProperties;

export const statusStripMutedItemStyle = {
  ...statusStripItemStyle,
  color: "var(--ct-colors-ink-subtle)",
} satisfies CSSProperties;

export const statusStripSecondaryItemStyle = {
  ...statusStripMutedItemStyle,
  minWidth: 0,
  maxWidth: "min(34rem, 42vw)",
  overflow: "hidden",
  textOverflow: "ellipsis",
} satisfies CSSProperties;

export const statusStripSpacerStyle = {
  flex: "1 1 auto",
} satisfies CSSProperties;

export const statusStripPresenceStyle = {
  ...statusStripItemStyle,
  gap: "0.25rem",
} satisfies CSSProperties;

export const presenceAvatarStyle = {
  display: "inline-grid",
  placeItems: "center",
  width: "1.5rem",
  height: "1.5rem",
  borderRadius: "var(--ct-rounded-pill)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-semantic-presence-self)",
  fontSize: "0.72rem",
  fontWeight: 700,
} satisfies CSSProperties;

export const presenceOverflowStyle = {
  ...presenceAvatarStyle,
  width: "auto",
  minWidth: "1.5rem",
  paddingInline: "0.35rem",
} satisfies CSSProperties;

export const presenceEmptyStyle = {
  fontSize: "0.75rem",
  color: "var(--ct-colors-ink-muted)",
} satisfies CSSProperties;
