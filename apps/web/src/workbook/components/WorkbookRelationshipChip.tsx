import { relationshipChipTestId } from "@cartulary/ui-contracts";
import type { CSSProperties } from "react";
import type { WorkbookRelationshipChipPresentation } from "../models/workbookRelationshipChip";
import { visuallyHiddenStyle } from "../utils/workbookStyles";

export function WorkbookRelationshipChip({
  presentation,
}: {
  readonly presentation: WorkbookRelationshipChipPresentation;
}) {
  const {
    accessibleDetail,
    label,
    onSelect,
    selected,
    selectorIdentity,
    state,
  } = presentation;
  const isResolved = state === "resolved" || state === "auto_resolved";
  const isDismissed = state === "dismissed";
  const isAutoResolved = state === "auto_resolved";
  const chipStyle = {
    ...workbookRelationshipChipBaseStyle,
    ...(isDismissed
      ? dismissedChipStyle
      : isResolved
        ? isAutoResolved
          ? autoResolvedChipStyle
          : resolvedChipStyle
        : unresolvedChipStyle),
    ...(selected ? selectedChipStyle : null),
  };
  const stateLabel =
    state === "auto_resolved"
      ? "Auto-resolved"
      : state === "dismissed"
        ? "Dismissed"
        : state === "resolved"
          ? "Resolved"
          : "Unresolved";
  const accessibleLabel = `${stateLabel} ${label}${
    accessibleDetail === undefined ? "" : `; ${accessibleDetail}`
  }`;
  const content = (
    <WorkbookRelationshipChipContent
      detail={accessibleDetail}
      label={label}
      state={state}
    />
  );

  return onSelect === undefined ? (
    <span
      aria-label={accessibleLabel}
      data-relationship-chip="true"
      data-testid={relationshipChipTestId(selectorIdentity)}
      role="note"
      style={chipStyle}
      title={label}
    >
      {content}
    </span>
  ) : (
    <button
      aria-label={accessibleLabel}
      data-relationship-chip="true"
      data-testid={relationshipChipTestId(selectorIdentity)}
      tabIndex={0}
      style={chipStyle}
      title={label}
      type="button"
      onClick={onSelect}
    >
      {content}
    </button>
  );
}

function WorkbookRelationshipChipContent({
  detail,
  label,
  state,
}: {
  readonly detail?: string | undefined;
  readonly label: string;
  readonly state: WorkbookRelationshipChipPresentation["state"];
}) {
  const manual = state === "resolved" && detail === "manual resolution";
  const stateMarker =
    state === "auto_resolved"
      ? "A"
      : manual
        ? "M"
        : state === "resolved"
          ? "R"
          : state === "dismissed"
            ? "D"
            : "!";
  const stateText =
    state === "auto_resolved"
      ? "Auto"
      : manual
        ? "Manual"
        : state === "resolved"
          ? "Resolved"
          : state === "dismissed"
            ? "Dismissed"
            : "Unresolved";

  return (
    <>
      <span aria-hidden="true" style={chipStateMarkerStyle}>
        {stateMarker}
      </span>
      <span style={chipLabelStyle}>{label}</span>
      <span data-density-role="narrow-metadata" style={visuallyHiddenStyle}>
        {stateText}
      </span>
    </>
  );
}

export const workbookRelationshipChipBaseStyle = {
  display: "inline-flex",
  alignItems: "center",
  flex: "0 1 auto",
  gap: "0.25rem",
  borderRadius: "var(--ct-component-chip-rounded)",
  padding: "var(--ct-component-chip-padding)",
  font: "inherit",
  lineHeight: 1.2,
  maxWidth: "100%",
  minWidth: 0,
  overflow: "hidden",
  overflowWrap: "normal" as const,
  whiteSpace: "nowrap" as const,
};

const unresolvedChipStyle = {
  border: "1px dashed var(--ct-colors-semantic-caution)",
  background: "var(--ct-component-chip-backgroundColor)",
  color: "var(--ct-colors-semantic-caution)",
};

const resolvedChipStyle = {
  border: "var(--ct-component-chip-border)",
  background: "var(--ct-component-chip-backgroundColor)",
  color: "var(--ct-colors-ink)",
};

const autoResolvedChipStyle = {
  border: "var(--ct-component-chip-border)",
  background: "var(--ct-component-chip-backgroundColor)",
  color: "var(--ct-colors-semantic-info)",
};

const dismissedChipStyle = {
  border: "var(--ct-border-hairline)",
  background: "transparent",
  color: "var(--ct-colors-ink-tertiary)",
};

const selectedChipStyle = {
  boxShadow: "0 0 0 2px var(--ct-colors-accent)",
};

const chipStateMarkerStyle = {
  display: "inline-grid",
  placeItems: "center",
  flex: "0 0 auto",
  inlineSize: "1.05rem",
  blockSize: "1.05rem",
  borderRadius: "var(--ct-rounded-pill)",
  border: "var(--ct-border-hairline)",
  fontFamily: "var(--ct-typography-mono-fontFamily)",
  fontSize: "0.62rem",
  fontWeight: 700,
  lineHeight: 1,
} satisfies CSSProperties;

const chipLabelStyle = {
  minWidth: 0,
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap",
} satisfies CSSProperties;
