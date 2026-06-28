import type { CSSProperties } from "react";

export const landingAdminShellStyle: CSSProperties = {
  width: "100%",
  minHeight: "100vh",
  display: "grid",
  gridTemplateRows: "auto minmax(0, 1fr)",
  background: "var(--ct-colors-canvas)",
  overflow: "hidden",
};

export const incidentDirectoryShellStyle: CSSProperties = {
  ...landingAdminShellStyle,
  overflow: "auto",
};

export const landingAdminHeaderStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "minmax(12rem, 1fr) auto minmax(14rem, 1fr)",
  gap: "var(--ct-spacing-lg)",
  alignItems: "center",
  minHeight: "var(--ct-layout-topBarHeight)",
  padding: "var(--ct-spacing-sm) var(--ct-spacing-lg)",
  borderBottom: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
};

export const brandBlockStyle: CSSProperties = {
  minWidth: 0,
};

export const landingEyebrowStyle: CSSProperties = {
  margin: 0,
  fontSize: "0.72rem",
  letterSpacing: 0,
  textTransform: "uppercase",
  color: "var(--ct-colors-accent)",
};

export const landingAdminTitleStyle: CSSProperties = {
  margin: "0.18rem 0 0",
  fontSize: "var(--ct-typography-surface-title-fontSize)",
  lineHeight: "var(--ct-typography-surface-title-lineHeight)",
};

export const landingAdminHeaderMetaStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "repeat(2, minmax(9rem, auto))",
  gap: "var(--ct-spacing-lg)",
  margin: 0,
};

export const landingAdminMetaValueStyle: CSSProperties = {
  margin: "0.2rem 0 0",
  color: "var(--ct-colors-ink)",
  fontWeight: 700,
};

export const landingToolbarLabelStyle: CSSProperties = {
  margin: 0,
  fontSize: "0.68rem",
  letterSpacing: 0,
  textTransform: "uppercase",
  color: "var(--ct-colors-ink-subtle)",
};

export const landingAccountNavStyle: CSSProperties = {
  display: "flex",
  flexWrap: "wrap",
  justifyContent: "flex-end",
  gap: "var(--ct-spacing-xs)",
};

export const accountMenuAnchorStyle: CSSProperties = {
  position: "relative",
  display: "inline-flex",
  justifyContent: "flex-end",
  minInlineSize: 0,
  margin: 0,
  padding: 0,
  border: 0,
};

export const accountMenuTriggerStyle: CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  maxWidth: "min(9rem, 100%)",
  minWidth: 0,
  gap: "0.45rem",
  padding: "0.5rem 0.65rem",
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink)",
  fontWeight: 700,
  cursor: "pointer",
};

export const accountMenuTriggerTextStyle: CSSProperties = {
  minWidth: 0,
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap",
};

export const accountMenuStyle: CSSProperties = {
  position: "absolute",
  insetBlockStart: "calc(100% + 0.35rem)",
  insetInlineEnd: 0,
  zIndex: 20,
  minWidth: "18rem",
  display: "grid",
  gap: "0.2rem",
  padding: "0.4rem",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-2)",
  boxShadow: "var(--ct-elevation-panel)",
};

export const accountMenuItemStyle: CSSProperties = {
  width: "100%",
  padding: "0.55rem 0.65rem",
  border: "none",
  borderRadius: "var(--ct-rounded-sm)",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  font: "inherit",
  fontWeight: 700,
  textAlign: "left",
  cursor: "pointer",
};

export const accountMenuStatusItemStyle: CSSProperties = {
  width: "100%",
  padding: "0.55rem 0.65rem",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-ink-muted)",
  fontWeight: 700,
  textAlign: "left",
};

export const accountSubmenuStyle: CSSProperties = {
  display: "grid",
  gap: "0.2rem",
  padding: "0.2rem 0 0.25rem 0.55rem",
};

export const accountSubmenuItemStyle: CSSProperties = {
  display: "grid",
  gap: "0.16rem",
  width: "100%",
  padding: "0.5rem 0.6rem",
  border: "none",
  borderRadius: "var(--ct-rounded-sm)",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  font: "inherit",
  textAlign: "left",
  cursor: "pointer",
};

export const accountSubmenuItemSelectedStyle: CSSProperties = {
  ...accountSubmenuItemStyle,
  background: "var(--ct-colors-surface-3)",
  color: "var(--ct-colors-ink)",
};

export const accountSubmenuItemLabelStyle: CSSProperties = {
  fontWeight: 700,
};

export const accountSubmenuItemDescriptionStyle: CSSProperties = {
  color: "var(--ct-colors-ink-subtle)",
  fontSize: "0.76rem",
};

export const accountMenuSeparatorStyle: CSSProperties = {
  height: 1,
  margin: "0.25rem 0",
  background: "var(--ct-colors-border-muted)",
};

export const landingAccountNavButtonStyle: CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  gap: "0.42rem",
  padding: "0.5rem 0.65rem",
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink-muted)",
  fontWeight: 700,
  cursor: "pointer",
};

export const landingAccountNavButtonSelectedStyle: CSSProperties = {
  ...landingAccountNavButtonStyle,
  border: "var(--ct-border-strong)",
  color: "var(--ct-colors-ink)",
  background: "var(--ct-colors-surface-3)",
};

export const landingAdminWorkspaceStyle: CSSProperties = {
  minHeight: 0,
  display: "grid",
  gridTemplateColumns: "16rem minmax(0, 1fr)",
  background: "var(--ct-colors-canvas)",
};

export const landingAdminMenuStyle: CSSProperties = {
  minHeight: 0,
  padding: "var(--ct-spacing-md)",
  borderRight: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  overflow: "auto",
};

export const landingAdminMenuItemsStyle: CSSProperties = {
  display: "grid",
  gap: "var(--ct-spacing-lg)",
};

export const menuGroupStyle: CSSProperties = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
};

export const menuGroupTitleStyle: CSSProperties = {
  margin: 0,
  padding: "0 var(--ct-spacing-xs)",
  fontSize: "0.68rem",
  letterSpacing: "0.14em",
  textTransform: "uppercase",
  color: "var(--ct-colors-ink-subtle)",
};

export const menuGroupItemsStyle: CSSProperties = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
};

export const landingAdminMenuItemStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "1.1rem minmax(0, 1fr)",
  gap: "var(--ct-spacing-sm)",
  alignItems: "start",
  width: "100%",
  minWidth: 0,
  padding: "var(--ct-spacing-sm)",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  textAlign: "left",
  cursor: "pointer",
};

export const landingAdminMenuItemSelectedStyle: CSSProperties = {
  ...landingAdminMenuItemStyle,
  border: "var(--ct-border-strong)",
  background: "var(--ct-colors-surface-3)",
  color: "var(--ct-colors-ink)",
  boxShadow: "inset 3px 0 0 var(--ct-colors-accent)",
};

export const landingAdminMenuItemTextStyle: CSSProperties = {
  display: "grid",
  gap: "0.2rem",
  minWidth: 0,
};

export const landingAdminMenuItemLabelStyle: CSSProperties = {
  fontWeight: 700,
  overflowWrap: "anywhere",
};

export const landingAdminMenuItemDescriptionStyle: CSSProperties = {
  color: "var(--ct-colors-ink-subtle)",
  fontSize: "0.76rem",
  overflowWrap: "anywhere",
};

export const landingAdminContentStyle: CSSProperties = {
  minWidth: 0,
  minHeight: 0,
  overflow: "auto",
  width: "100%",
  maxWidth: "1600px",
  justifySelf: "center",
  boxSizing: "border-box",
};

export const surfacePanelStyle: CSSProperties = {
  minWidth: 0,
  minHeight: 0,
  display: "grid",
  gap: "var(--ct-spacing-md)",
  padding: "var(--ct-spacing-md)",
  color: "var(--ct-colors-ink)",
};

export const surfaceHeaderStyle: CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  alignItems: "center",
  gap: "var(--ct-spacing-md)",
  padding: "0 0 var(--ct-spacing-sm)",
  borderBottom: "var(--ct-border-hairline)",
};

export const headerActionRowStyle: CSSProperties = {
  display: "flex",
  alignItems: "center",
  gap: "var(--ct-spacing-sm)",
};

export const sectionEyebrowStyle: CSSProperties = {
  margin: 0,
  fontSize: "0.72rem",
  letterSpacing: "0.14em",
  textTransform: "uppercase",
  color: "var(--ct-colors-ink-subtle)",
};

export const sectionTitleStyle: CSSProperties = {
  margin: "0.2rem 0 0",
  fontSize: "1.25rem",
};

export const subsectionTitleStyle: CSSProperties = {
  margin: "0.2rem 0 0",
  fontSize: "var(--ct-typography-section-heading-fontSize)",
};

export const incidentWorkspaceStyle: CSSProperties = {
  minWidth: 0,
  minHeight: 0,
  display: "grid",
  gridTemplateColumns: "minmax(0, 1fr)",
  gap: "var(--ct-spacing-md)",
  alignItems: "start",
};

export const incidentDirectoryStyle: CSSProperties = {
  minWidth: 0,
  display: "grid",
  gap: "var(--ct-spacing-md)",
};

export const toolbarGridStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "minmax(18rem, 1fr) minmax(10rem, 14rem)",
  gap: "var(--ct-spacing-sm)",
  alignItems: "end",
};

export const labelBlockStyle: CSSProperties = {
  display: "grid",
  gap: "0.35rem",
  minWidth: 0,
  fontSize: "0.82rem",
  fontWeight: 700,
  color: "var(--ct-colors-ink-muted)",
};

export const searchInputShellStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "1rem minmax(0, 1fr)",
  gap: "0.55rem",
  alignItems: "center",
  padding: "0.72rem 0.85rem",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
};

export const searchInputStyle: CSSProperties = {
  minWidth: 0,
  width: "100%",
  border: "none",
  outline: "none",
  padding: 0,
  background: "transparent",
  color: "var(--ct-component-text-input-textColor)",
};

export const inputStyle: CSSProperties = {
  boxSizing: "border-box",
  width: "100%",
  maxWidth: "100%",
  minWidth: 0,
  padding: "var(--ct-component-text-input-padding)",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  color: "var(--ct-component-text-input-textColor)",
  fontSize: "0.92rem",
};

export const textAreaStyle: CSSProperties = {
  ...inputStyle,
  minHeight: "6rem",
  resize: "vertical",
};

export const formGridStyle: CSSProperties = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
};

export const segmentedFormStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "minmax(16rem, 24rem) auto",
  gap: "var(--ct-spacing-md)",
  alignItems: "end",
  marginTop: "var(--ct-spacing-md)",
};

export const detailsStyle: CSSProperties = {
  marginTop: "0.25rem",
};

export const detailsSummaryStyle: CSSProperties = {
  cursor: "pointer",
  color: "var(--ct-colors-ink-muted)",
  fontWeight: 700,
};

export const dialogBackdropStyle: CSSProperties = {
  position: "fixed",
  inset: 0,
  zIndex: 40,
  display: "grid",
  placeItems: "center",
  padding: "var(--ct-spacing-lg)",
  background: "rgba(10, 13, 18, 0.68)",
};

export const createDialogStyle: CSSProperties = {
  width: "min(48rem, 100%)",
  maxHeight: "calc(100vh - 3rem)",
  overflow: "auto",
  display: "grid",
  gap: "var(--ct-spacing-md)",
  padding: "var(--ct-spacing-lg)",
  border: "var(--ct-border-strong)",
  borderRadius: "var(--ct-rounded-md)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-panel)",
};

export const dialogHeaderStyle: CSSProperties = {
  display: "flex",
  alignItems: "flex-start",
  justifyContent: "space-between",
  gap: "var(--ct-spacing-md)",
};

export const iconButtonStyle: CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  width: "2rem",
  height: "2rem",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink)",
  cursor: "pointer",
};

export const buttonRowEndStyle: CSSProperties = {
  display: "flex",
  justifyContent: "flex-end",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-sm)",
};

export const primaryButtonStyle: CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  gap: "0.45rem",
  padding: "var(--ct-component-button-primary-padding)",
  borderRadius: "var(--ct-component-button-primary-rounded)",
  border: "none",
  background: "var(--ct-component-button-primary-backgroundColor)",
  color: "var(--ct-component-button-primary-textColor)",
  fontWeight: 700,
  cursor: "pointer",
};

export const secondaryButtonStyle: CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  gap: "0.45rem",
  padding: "var(--ct-component-button-secondary-padding)",
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  fontWeight: 700,
  cursor: "pointer",
};

export const countPillStyle: CSSProperties = {
  margin: 0,
  minWidth: "2.5rem",
  padding: "0.45rem 0.75rem",
  borderRadius: "var(--ct-rounded-pill)",
  background: "var(--ct-colors-surface-3)",
  textAlign: "center",
  fontWeight: 700,
};

export const inlineStatusStyle: CSSProperties = {
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
};

export const emptyStateStyle: CSSProperties = {
  margin: "var(--ct-spacing-md) 0 0",
  color: "var(--ct-colors-ink-muted)",
};

export const tableShellStyle: CSSProperties = {
  minWidth: 0,
  overflowX: "auto",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-1)",
};

export const dataTableStyle: CSSProperties = {
  width: "100%",
  borderCollapse: "collapse",
  minWidth: "62rem",
};

export const tableHeaderCellStyle: CSSProperties = {
  padding: "0.7rem 0.85rem",
  borderBottom: "var(--ct-border-hairline)",
  color: "var(--ct-colors-ink-subtle)",
  fontSize: "0.68rem",
  letterSpacing: "0.12em",
  textTransform: "uppercase",
  textAlign: "left",
  whiteSpace: "nowrap",
};

export const tableHeaderActionCellStyle: CSSProperties = {
  ...tableHeaderCellStyle,
  textAlign: "right",
};

export const tableRowStyle: CSSProperties = {
  borderBottom: "var(--ct-border-hairline)",
};

export const tableCellStyle: CSSProperties = {
  padding: "0.75rem 0.85rem",
  color: "var(--ct-colors-ink-muted)",
  verticalAlign: "top",
  fontSize: "0.84rem",
};

export const primaryCellStyle: CSSProperties = {
  ...tableCellStyle,
  color: "var(--ct-colors-ink)",
  minWidth: "14rem",
};

export const tableActionCellStyle: CSSProperties = {
  ...tableCellStyle,
  textAlign: "right",
};

export const tableLinkButtonStyle: CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  gap: "0.35rem",
  padding: "0.25rem 0",
  border: "none",
  background: "transparent",
  color: "var(--ct-colors-ink)",
  fontWeight: 700,
  cursor: "pointer",
};

export const strongTextStyle: CSSProperties = {
  margin: 0,
  color: "var(--ct-colors-ink)",
  fontWeight: 700,
  overflowWrap: "anywhere",
};

export const metadataTextStyle: CSSProperties = {
  margin: "0.24rem 0 0",
  color: "var(--ct-colors-ink-subtle)",
  fontSize: "0.78rem",
  overflowWrap: "anywhere",
};

export const monoMutedStyle: CSSProperties = {
  margin: 0,
  color: "var(--ct-colors-ink-subtle)",
  fontFamily: "var(--ct-typography-mono-fontFamily)",
  fontSize: "0.78rem",
  overflowWrap: "anywhere",
};

export const unsetValueStyle: CSSProperties = {
  color: "var(--ct-colors-ink-subtle)",
  fontStyle: "italic",
};

export const activeBadgeStyle: CSSProperties = {
  display: "inline-flex",
  padding: "0.18rem 0.5rem",
  borderRadius: "var(--ct-rounded-pill)",
  color: "var(--ct-colors-semantic-success)",
  background:
    "color-mix(in srgb, var(--ct-colors-semantic-success) 14%, transparent)",
  fontWeight: 700,
};

export const closedBadgeStyle: CSSProperties = {
  ...activeBadgeStyle,
  color: "var(--ct-colors-ink-muted)",
  background: "var(--ct-colors-surface-3)",
};

export const statusTextStyle: CSSProperties = {
  margin: "0.2rem 0 0",
  minHeight: "1.4rem",
  color: "var(--ct-colors-ink-muted)",
};

export const errorTextStyle: CSSProperties = {
  margin: 0,
  minHeight: "1.25rem",
  color: "var(--ct-colors-semantic-conflict)",
  fontWeight: 700,
};

export const publicErrorStyle: CSSProperties = {
  marginTop: "0.25rem",
};

export const errorMessageStyle: CSSProperties = {
  margin: 0,
  minHeight: "1.25rem",
  color: "var(--ct-colors-semantic-conflict)",
};

export const errorDetailStyle: CSSProperties = {
  margin: "0.2rem 0 0",
  minHeight: "1.25rem",
  color: "var(--ct-colors-semantic-conflict)",
  overflowWrap: "anywhere",
};

export const definitionPanelStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "minmax(12rem, 24rem) minmax(12rem, 24rem) auto",
  gap: "var(--ct-spacing-md)",
  alignItems: "end",
  marginTop: "var(--ct-spacing-md)",
};

export const definitionLabelStyle: CSSProperties = {
  display: "block",
  color: "var(--ct-colors-ink-subtle)",
  fontSize: "0.72rem",
  letterSpacing: "0.12em",
  textTransform: "uppercase",
  marginBottom: "0.35rem",
};

export const definitionValueStyle: CSSProperties = {
  minHeight: "2.45rem",
  display: "flex",
  alignItems: "center",
  padding: "0.3rem 0",
  color: "var(--ct-colors-ink)",
  overflowWrap: "anywhere",
};

export const auditFilterGridStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(13rem, 1fr))",
  gap: "var(--ct-spacing-sm)",
};

export const auditDetailCellStyle: CSSProperties = {
  padding: 0,
  borderBottom: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
};

export const auditDetailPanelStyle: CSSProperties = {
  display: "grid",
  gap: "var(--ct-spacing-md)",
  padding: "var(--ct-spacing-md)",
};

export const auditDetailMetaGridStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(14rem, 1fr))",
  gap: "var(--ct-spacing-sm)",
};

export const auditChangeTableStyle: CSSProperties = {
  ...dataTableStyle,
  minWidth: "44rem",
};

export const redactedBadgeStyle: CSSProperties = {
  display: "inline-flex",
  padding: "0.08rem 0.38rem",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-pill)",
  color: "var(--ct-colors-semantic-caution)",
  fontWeight: 700,
};

export const nullValueStyle: CSSProperties = {
  color: "var(--ct-colors-ink-subtle)",
  fontStyle: "italic",
};

export const jobPanelStyle: CSSProperties = {
  padding: "var(--ct-spacing-md)",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-1)",
};

export const visuallyHiddenStyle: CSSProperties = {
  position: "absolute",
  width: "1px",
  height: "1px",
  padding: 0,
  margin: "-1px",
  overflow: "hidden",
  clip: "rect(0, 0, 0, 0)",
  whiteSpace: "nowrap",
  border: 0,
};
