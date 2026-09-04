import {
  workbookAddRowButtonTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import { Plus, SlidersHorizontal } from "lucide-react";
import type { CSSProperties, ReactNode, Ref } from "react";
import type { WorkbookChromeMode } from "../layout/workbookResponsiveLayout";
import {
  ActiveSurfaceSavedViewSelector,
  type ActiveSurfaceSavedViewSelectorProps,
} from "./ActiveSurfaceSavedViewSelector";
import {
  WorkbookGridControls,
  type WorkbookGridControlsProps,
} from "./WorkbookGridControls";

export type WorkbookViewBarWorkingSetBinding = {
  readonly query: Omit<WorkbookGridControlsProps, "chromeMode"> | null;
  readonly savedView: Omit<
    ActiveSurfaceSavedViewSelectorProps,
    "chromeMode"
  > | null;
};

type WorkbookViewBarProps = {
  readonly addRowDisabled?: boolean | undefined;
  readonly addRowLabel?: string | undefined;
  readonly chromeMode?: WorkbookChromeMode | undefined;
  readonly inspectorButtonRef?: Ref<HTMLButtonElement> | undefined;
  readonly onAddRow?: (() => void) | undefined;
  readonly onInspectorToggle?: (() => void) | undefined;
  readonly supplementalControls?: ReactNode | undefined;
  readonly surface: string;
  readonly workingSet?: WorkbookViewBarWorkingSetBinding | undefined;
};

export function WorkbookViewBar({
  addRowDisabled = false,
  addRowLabel = "Add row",
  chromeMode = "base",
  inspectorButtonRef,
  onAddRow,
  onInspectorToggle,
  supplementalControls,
  surface,
  workingSet,
}: WorkbookViewBarProps) {
  const compactActions =
    chromeMode === "compact_desktop" ||
    chromeMode === "below_supported_minimum";
  return (
    <section
      aria-label="Workbook query and action controls"
      data-chrome-mode={chromeMode}
      style={viewBarStyle}
    >
      <div style={controlRailStyle}>
        {workingSet?.savedView ? (
          <div style={savedViewAllocationStyleFor(chromeMode)}>
            <ActiveSurfaceSavedViewSelector
              {...workingSet.savedView}
              chromeMode={chromeMode}
            />
          </div>
        ) : null}
        {workingSet?.query ? (
          <div style={queryAllocationStyle}>
            <WorkbookGridControls
              {...workingSet.query}
              chromeMode={chromeMode}
            />
          </div>
        ) : null}
        {supplementalControls !== undefined && supplementalControls !== null ? (
          <div style={supplementalAllocationStyle}>{supplementalControls}</div>
        ) : null}
      </div>
      <div style={rightRailStyle}>
        {onInspectorToggle ? (
          <button
            aria-label="Open inspector"
            data-testid={workbookInspectorToggleTestId(surface)}
            ref={inspectorButtonRef}
            style={toolbarButtonStyle}
            title={compactActions ? "Open inspector" : undefined}
            type="button"
            onClick={onInspectorToggle}
          >
            <SlidersHorizontal aria-hidden="true" size={16} />
            {compactActions ? null : "Inspector"}
          </button>
        ) : null}
        {onAddRow ? (
          <button
            aria-label={addRowLabel}
            data-testid={workbookAddRowButtonTestId(surface)}
            disabled={addRowDisabled}
            style={primaryToolbarButtonStyle}
            title={compactActions ? addRowLabel : undefined}
            type="button"
            onClick={onAddRow}
          >
            <Plus aria-hidden="true" size={16} />
            {compactActions ? null : addRowLabel}
          </button>
        ) : null}
      </div>
    </section>
  );
}

const compactInputStyle = {
  borderRadius: "var(--ct-rounded-xs)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-ink)",
  font: "inherit",
  minHeight: "1.75rem",
  padding: "0.22rem 0.45rem",
} satisfies CSSProperties;

const toolbarButtonStyle = {
  ...compactInputStyle,
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  gap: "0.35rem",
  cursor: "pointer",
  whiteSpace: "nowrap",
} satisfies CSSProperties;

const primaryToolbarButtonStyle = {
  ...toolbarButtonStyle,
  borderColor: "var(--ct-colors-accent-active)",
  background: "var(--ct-colors-accent)",
  color: "var(--ct-colors-on-accent)",
  fontWeight: 700,
} satisfies CSSProperties;

const viewBarStyle = {
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

const controlRailStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.45rem",
  inlineSize: "100%",
  maxInlineSize: "100%",
  minWidth: 0,
  overflow: "visible",
} satisfies CSSProperties;

function savedViewAllocationStyleFor(
  chromeMode: WorkbookChromeMode,
): CSSProperties {
  if (chromeMode === "below_supported_minimum") return { display: "none" };
  const inlineSize =
    chromeMode === "base"
      ? "min(var(--ct-layout-viewBarSavedViewMaxInlineSize), 24vw)"
      : chromeMode === "narrow_desktop"
        ? "10rem"
        : "9rem";
  return {
    display: "flex",
    alignItems: "center",
    flex: `0 1 ${inlineSize}`,
    inlineSize,
    maxInlineSize: "var(--ct-layout-viewBarSavedViewMaxInlineSize)",
    minInlineSize: 0,
    overflow: "visible",
  };
}

const queryAllocationStyle = {
  display: "flex",
  alignItems: "center",
  flex: "1 1 0",
  inlineSize: "100%",
  maxInlineSize: "100%",
  minInlineSize: 0,
  overflow: "visible",
} satisfies CSSProperties;

const supplementalAllocationStyle = {
  display: "flex",
  alignItems: "center",
  flex: "0 1 auto",
  minInlineSize: 0,
  overflow: "visible",
} satisfies CSSProperties;

const rightRailStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "end",
  gap: "0.45rem",
  flex: "0 0 auto",
  minInlineSize: "max-content",
} satisfies CSSProperties;
