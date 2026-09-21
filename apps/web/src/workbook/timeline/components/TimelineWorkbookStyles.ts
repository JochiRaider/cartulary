import type { CSSProperties } from "react";
import {
  workbookFormButtonStyle,
  workbookFormFieldStackStyle,
  workbookFormFieldsStyle,
  workbookFormInputStyle,
  workbookTypography,
} from "../../components/workbookFormStyles";
import { workbookSurfaceGridShellStyle } from "../../layout/WorkbookSurfaceLayout";

export const timelineRowGutterWidth = 58;

export const bodyStyle = {
  ...workbookTypography("ui"),
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
};

export const timelineGridBodyStyle = {
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
  fontSize: "var(--cartulary-grid-font-size)",
  lineHeight: "var(--cartulary-grid-line-height)",
};

export const timelineGridShellStyle = {
  ...workbookSurfaceGridShellStyle,
} satisfies CSSProperties;

export const actionButtonStyle = {
  ...workbookFormButtonStyle,
  cursor: "pointer",
};

export const inputStyle = workbookFormInputStyle;

export const secondaryActionButtonStyle = {
  ...actionButtonStyle,
  background: "var(--ct-colors-surface-3)",
};

export const labelStyle = {
  ...workbookFormFieldStackStyle,
  ...workbookTypography("metadata"),
  color: "var(--ct-colors-ink-muted)",
};

export const inspectorSectionStyle = {
  ...workbookFormFieldsStyle,
  marginBottom: "var(--ct-spacing-sm)",
};

export const sectionTitleStyle = {
  ...workbookTypography("section-heading"),
  margin: 0,
};

export const inspectorActionStackStyle = workbookFormFieldsStyle;
