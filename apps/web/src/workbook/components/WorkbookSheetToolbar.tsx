import {
  workbookAddRowButtonTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import { Plus, Search, SlidersHorizontal } from "lucide-react";
import type { CSSProperties, ReactNode } from "react";

type WorkbookSheetToolbarProps = {
  readonly addRowDisabled?: boolean | undefined;
  readonly addRowLabel?: string | undefined;
  readonly leading?: ReactNode | undefined;
  readonly onAddRow?: (() => void) | undefined;
  readonly onInspectorToggle?: (() => void) | undefined;
  readonly surface: string;
};

export function WorkbookSheetToolbar({
  addRowDisabled = false,
  addRowLabel = "Add row",
  leading,
  onAddRow,
  onInspectorToggle,
  surface,
}: WorkbookSheetToolbarProps) {
  return (
    <div style={toolbarStyle}>
      <div style={leftRailStyle}>{leading}</div>
      <div style={rightRailStyle}>
        {onInspectorToggle ? (
          <button
            aria-label="Open inspector"
            data-testid={workbookInspectorToggleTestId(surface)}
            style={toolbarButtonStyle}
            type="button"
            onClick={onInspectorToggle}
          >
            <SlidersHorizontal aria-hidden="true" size={16} />
            Inspector
          </button>
        ) : null}
        {onAddRow ? (
          <button
            data-testid={workbookAddRowButtonTestId(surface)}
            disabled={addRowDisabled}
            style={primaryToolbarButtonStyle}
            type="button"
            onClick={onAddRow}
          >
            <Plus aria-hidden="true" size={16} />
            {addRowLabel}
          </button>
        ) : null}
      </div>
    </div>
  );
}

export const compactInputStyle = {
  borderRadius: "var(--ct-rounded-xs)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-ink)",
  font: "inherit",
  minHeight: "1.75rem",
  padding: "0.22rem 0.45rem",
} satisfies CSSProperties;

export const toolbarButtonStyle = {
  ...compactInputStyle,
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  gap: "0.35rem",
  cursor: "pointer",
  whiteSpace: "nowrap",
} satisfies CSSProperties;

export const primaryToolbarButtonStyle = {
  ...toolbarButtonStyle,
  borderColor: "var(--ct-colors-accent-active)",
  background: "var(--ct-colors-accent)",
  color: "var(--ct-colors-on-accent)",
  fontWeight: 700,
} satisfies CSSProperties;

const toolbarStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(0, 1fr) auto",
  alignItems: "center",
  gap: "0.5rem",
  blockSize: "var(--ct-layout-viewBarHeight)",
  minWidth: 0,
  padding: "0 var(--ct-spacing-sm)",
  borderBlockEnd: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  overflow: "visible",
} satisfies CSSProperties;

const leftRailStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.45rem",
  minWidth: 0,
  overflow: "visible",
} satisfies CSSProperties;

const rightRailStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "end",
  gap: "0.45rem",
  minWidth: 0,
} satisfies CSSProperties;

const toolbarStatusStyle = {
  display: "inline-flex",
  alignItems: "center",
  gap: "0.3rem",
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.82rem",
  whiteSpace: "nowrap",
} satisfies CSSProperties;

export function WorkbookToolbarSearchLabel({
  children,
}: {
  readonly children: ReactNode;
}) {
  return (
    <span style={toolbarStatusStyle}>
      <Search aria-hidden="true" size={16} />
      {children}
    </span>
  );
}
