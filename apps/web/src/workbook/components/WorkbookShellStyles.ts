export const panelStyle = {
  boxSizing: "border-box" as const,
  display: "grid",
  gridTemplateRows: "auto minmax(0, 1fr)",
  width: "100%",
  blockSize: "100%",
  minBlockSize: 0,
  margin: 0,
  padding: 0,
  borderRadius: 0,
  background: "var(--ct-colors-canvas)",
  boxShadow: "none",
  border: 0,
  overflow: "hidden",
};

export const shellTopBarStyle = {
  boxSizing: "border-box" as const,
  display: "flex",
  alignItems: "center",
  flexWrap: "nowrap" as const,
  gap: "0.55rem",
  inlineSize: "100%",
  maxInlineSize: "100%",
  blockSize: "var(--ct-layout-topBarHeight)",
  minBlockSize: "var(--ct-layout-topBarHeight)",
  minWidth: 0,
  padding: "0 0.75rem",
  borderBottom: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  overflow: "visible",
};

export const shellTopBarUnsupportedStyle = {
  overflowX: "auto" as const,
};

export const shellTopBarActionsStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "flex-end",
  gap: "0.45rem",
  flex: "0 0 auto",
  minWidth: 0,
  order: 5,
};

export const shellTopBarValueStyle = {
  margin: 0,
  fontWeight: 650,
  color: "var(--ct-colors-ink)",
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
};

export const shellIncidentTitleStyle = {
  minWidth: 0,
  color: "var(--ct-colors-ink-muted)",
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
};

export const shellIncidentIdentityStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.45rem",
  flex: "0 1 11rem",
  minWidth: 0,
  overflow: "hidden",
};

export const currentUserSlotStyle = {
  display: "inline-flex",
  alignItems: "center",
  flex: "0 1 8rem",
  maxInlineSize: "8rem",
  minWidth: 0,
};

export const currentUserChipStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  flex: "0 0 auto",
  width: "2rem",
  height: "2rem",
  borderRadius: "var(--ct-rounded-pill)",
  border: "var(--ct-border-hairline)",
  color: "var(--ct-colors-ink)",
  background: "var(--ct-colors-surface-2)",
  fontSize: "0.82rem",
  fontWeight: 700,
};

export const shellContentRegionStyle = {
  position: "relative" as const,
  display: "grid",
  gridTemplateRows: "auto minmax(0, 1fr)",
  blockSize: "100%",
  minBlockSize: 0,
  minHeight: 0,
  minWidth: 0,
  overflow: "hidden",
};

export const shellContentNoticeStyle = {
  gridRow: 1,
  margin: 0,
  lineHeight: 1.5,
  color: "var(--ct-colors-ink-muted)",
  padding: "0.35rem 0.75rem",
  borderBottom: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
};

export const shellActiveSurfaceStyle = {
  display: "grid",
  gridTemplateRows: "minmax(0, 1fr)",
  gridRow: 2,
  blockSize: "100%",
  minBlockSize: 0,
  minHeight: 0,
  minWidth: 0,
  overflow: "hidden",
};

export const tabStripStyle = {
  display: "flex",
  alignItems: "stretch",
  gap: "0.2rem",
  flex: "0 1 auto",
  minWidth: 0,
  overflow: "hidden",
};

export const surfaceTabStyle = {
  borderRadius: 0,
  border: 0,
  borderBottom: "2px solid transparent",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  padding: "0 0.35rem",
  font: "inherit",
  cursor: "pointer",
  whiteSpace: "nowrap" as const,
  minBlockSize: "var(--ct-layout-topBarHeight)",
};

export const surfaceTabActiveStyle = {
  background: "transparent",
  color: "var(--ct-colors-ink)",
  borderBottomColor: "var(--ct-colors-accent)",
};

export const systemViewSlotStyle = {
  display: "inline-flex",
  alignItems: "center",
  gap: "0.35rem",
  flex: "0 1 auto",
  minWidth: 0,
  order: 4,
};

export const activeSystemViewTitleStyle = {
  color: "var(--ct-colors-ink)",
  fontSize: "0.86rem",
  fontWeight: 650,
  maxInlineSize: "6rem",
  minWidth: 0,
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
};

export const surfacesMenuFrameStyle = {
  position: "relative" as const,
  display: "inline-flex",
  flex: "0 0 auto",
};

export const surfaceMenuTriggerStyle = {
  borderRadius: "var(--ct-rounded-xs)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-ink)",
  padding: "0.35rem 0.55rem",
  font: "inherit",
  cursor: "pointer",
  whiteSpace: "nowrap" as const,
};

export const surfacesMenuStyle = {
  position: "absolute" as const,
  zIndex: 18,
  insetBlockStart: "calc(100% + 0.35rem)",
  insetInlineStart: 0,
  display: "grid",
  gap: "0.2rem",
  inlineSize: "min(16rem, 80vw)",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-md)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-popover)",
  padding: "0.45rem",
};

export const surfacesMenuItemStyle = {
  border: 0,
  borderRadius: "var(--ct-rounded-xs)",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  cursor: "pointer",
  font: "inherit",
  padding: "0.45rem 0.5rem",
  textAlign: "left" as const,
};

export const surfacesMenuItemSelectedStyle = {
  background: "var(--ct-colors-surface-3)",
  color: "var(--ct-colors-ink)",
  fontWeight: 700,
};
