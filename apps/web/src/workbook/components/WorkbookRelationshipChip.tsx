import { relationshipChipTestId } from "@cartulary/ui-contracts";
import type { CSSProperties, KeyboardEvent, RefCallback } from "react";
import {
  relationshipChipAccessibleName,
  type WorkbookRelationshipChipPresentation,
} from "../models/workbookRelationshipChip";

export function WorkbookRelationshipChip({
  presentation,
  elementRef,
  onKeyDown,
  tabIndex = 0,
  expanded = false,
  decorative = false,
}: {
  readonly presentation: WorkbookRelationshipChipPresentation;
  readonly elementRef?: RefCallback<HTMLButtonElement> | undefined;
  readonly onKeyDown?:
    | ((event: KeyboardEvent<HTMLButtonElement>) => void)
    | undefined;
  readonly tabIndex?: number | undefined;
  readonly expanded?: boolean | undefined;
  readonly decorative?: boolean | undefined;
}) {
  const { label, onSelect, selected, selectorIdentity, state } = presentation;
  const chipStyle = {
    ...workbookRelationshipChipBaseStyle,
    ...(expanded ? null : { paddingBlock: 0 }),
    ...(state === "dismissed"
      ? dismissedChipStyle
      : state === "auto_resolved"
        ? autoResolvedChipStyle
        : state === "resolved"
          ? resolvedChipStyle
          : unresolvedChipStyle),
    ...(selected ? selectedChipStyle : null),
  };
  const marker =
    state === "unresolved"
      ? "?"
      : state === "auto_resolved"
        ? "auto"
        : state === "dismissed"
          ? "dismissed"
          : null;
  const content = (
    <>
      {marker === null ? null : (
        <span aria-hidden="true" style={chipStateMarkerStyle}>
          {marker}
        </span>
      )}
      <span style={expanded ? expandedChipLabelStyle : chipLabelStyle}>
        {label}
      </span>
    </>
  );
  const accessibleName = relationshipChipAccessibleName(presentation);
  return onSelect === undefined ? (
    <span
      aria-label={decorative ? undefined : accessibleName}
      data-relationship-chip="true"
      data-testid={relationshipChipTestId(selectorIdentity)}
      role="note"
      style={chipStyle}
    >
      {content}
    </span>
  ) : (
    <button
      aria-label={accessibleName}
      aria-pressed={selected}
      data-relationship-chip="true"
      data-testid={relationshipChipTestId(selectorIdentity)}
      ref={elementRef}
      tabIndex={tabIndex}
      style={chipStyle}
      type="button"
      onClick={(event) => {
        event.stopPropagation();
        onSelect();
      }}
      onKeyDown={(event) => {
        onKeyDown?.(event);
        if (event.key !== "Escape" && event.key !== "Tab")
          event.stopPropagation();
      }}
    >
      {content}
    </button>
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

export function WorkbookRelationshipChipDetails({
  presentation,
}: {
  readonly presentation: WorkbookRelationshipChipPresentation;
}) {
  const resolutionLabels = {
    manual: "Manual",
    auto: "Automatic",
    import: "Imported",
    system: "System",
  } as const;
  const entries = [
    ["Entity type", presentation.entityType],
    ["Source text", presentation.rawText],
    [
      "Status",
      presentation.state === "auto_resolved"
        ? "Auto-resolved"
        : presentation.state,
    ],
    [
      "Target",
      presentation.targetRecordId === null ? "None" : presentation.label,
    ],
    ...(presentation.previousTarget === null
      ? []
      : [["Prior target", presentation.previousTarget.label]]),
    [
      "Resolution",
      presentation.resolution.method === null
        ? "Not available"
        : resolutionLabels[presentation.resolution.method],
    ],
    ...(presentation.resolution.matchedAliasText === null
      ? []
      : [["Matched alias", presentation.resolution.matchedAliasText]]),
    ...(presentation.resolution.provenance === null
      ? []
      : [["Provenance", presentation.resolution.provenance]]),
    ...(presentation.resolution.confidence === null
      ? []
      : [["Confidence", String(presentation.resolution.confidence)]]),
  ];
  return (
    <dl style={{ margin: 0, display: "grid", gap: "0.4rem", minWidth: 0 }}>
      {entries.map(([term, value]) => (
        <div key={term}>
          <dt style={{ color: "var(--ct-colors-ink-muted)" }}>{term}</dt>
          <dd
            style={{
              margin: 0,
              whiteSpace: "pre-wrap",
              overflowWrap: "anywhere",
            }}
          >
            {value}
          </dd>
        </div>
      ))}
    </dl>
  );
}

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
  fontFamily: "var(--ct-typography-mono-fontFamily)",
  fontSize: "inherit",
  fontWeight: 700,
  lineHeight: 1,
} satisfies CSSProperties;

const chipLabelStyle = {
  minWidth: 0,
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap",
} satisfies CSSProperties;

const expandedChipLabelStyle = {
  minWidth: 0,
  whiteSpace: "pre-wrap",
  overflowWrap: "anywhere",
} satisfies CSSProperties;
