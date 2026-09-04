import { workbookQueryEntryTestId } from "@cartulary/ui-contracts";
import { type KeyboardEvent, type RefObject, useEffect, useState } from "react";
import type {
  WorkbookGridQueryCommand,
  WorkbookGridQueryControlProjection,
  WorkbookQueryChip,
} from "../models/workbookGridQueryControls";
import { visuallyHiddenStyle } from "../utils/workbookStyles";
import { controlButtonStyle } from "./workbookGridControlStyles";

export function WorkbookActiveQueryChips({
  activeKey,
  condensed,
  entryRefs,
  onActivate,
  onCommand,
  onFallbackFocus,
  onRovingEntryChange,
  projection,
  surface,
}: {
  readonly activeKey: string | null;
  readonly condensed: boolean;
  readonly entryRefs: RefObject<Map<string, HTMLButtonElement>>;
  readonly onActivate: (
    chip: WorkbookQueryChip,
    element: HTMLButtonElement,
  ) => void;
  readonly onCommand: (command: WorkbookGridQueryCommand) => void;
  readonly onFallbackFocus: (chip: WorkbookQueryChip) => void;
  readonly onRovingEntryChange: (entryKey: string | null) => void;
  readonly projection: WorkbookGridQueryControlProjection;
  readonly surface: string;
}) {
  const chips = projection.visibleChips;
  const [describedKey, setDescribedKey] = useState<string | null>(null);
  const resolvedActiveKey = chips.some((chip) => chip.key === activeKey)
    ? activeKey
    : (chips[0]?.key ?? null);

  useEffect(() => {
    if (resolvedActiveKey === activeKey) return;
    onRovingEntryChange(resolvedActiveKey);
  }, [activeKey, onRovingEntryChange, resolvedActiveKey]);

  const focusAt = (index: number) => {
    const chip = chips[index];
    if (chip === undefined) return;
    onRovingEntryChange(chip.key);
    entryRefs.current.get(chip.key)?.focus({ preventScroll: true });
  };

  const deleteChip = (chip: WorkbookQueryChip) => {
    const index = chips.findIndex((candidate) => candidate.key === chip.key);
    const fallback = chips[index + 1] ?? chips[index - 1];
    onRovingEntryChange(fallback?.key ?? null);
    onCommand(chip.removeCommand);
    if (fallback !== undefined) {
      queueMicrotask(() => entryRefs.current.get(fallback.key)?.focus());
    } else {
      queueMicrotask(() => onFallbackFocus(chip));
    }
  };

  const onChipKeyDown = (
    event: KeyboardEvent<HTMLButtonElement>,
    chip: WorkbookQueryChip,
  ) => {
    const index = chips.findIndex((candidate) => candidate.key === chip.key);
    switch (event.key) {
      case "ArrowLeft":
      case "ArrowUp":
        event.preventDefault();
        focusAt(index <= 0 ? chips.length - 1 : index - 1);
        return;
      case "ArrowRight":
      case "ArrowDown":
        event.preventDefault();
        focusAt(index >= chips.length - 1 ? 0 : index + 1);
        return;
      case "Home":
        event.preventDefault();
        focusAt(0);
        return;
      case "End":
        event.preventDefault();
        focusAt(chips.length - 1);
        return;
      case "Delete":
        event.preventDefault();
        deleteChip(chip);
        return;
      case "Enter":
      case " ":
        event.preventDefault();
        onActivate(chip, event.currentTarget);
        return;
    }
  };

  return (
    <div
      aria-label="Active query chips"
      hidden={projection.visibleChipCapacity === 0 || chips.length === 0}
      role="toolbar"
      style={{
        ...chipRailStyle,
        ...(condensed ? condensedChipRailStyle : null),
      }}
    >
      <span style={visuallyHiddenStyle}>Active query chips</span>
      {chips.map((chip) => {
        const testId = workbookQueryEntryTestId(
          surface,
          chip.identity.kind,
          chip.identity.fieldKey,
        );
        const tooltipId = `${testId}-description`;
        return (
          <span key={chip.key} style={chipFrameStyle}>
            <button
              ref={(element) => {
                if (element === null) entryRefs.current.delete(chip.key);
                else entryRefs.current.set(chip.key, element);
              }}
              aria-describedby={tooltipId}
              aria-label={chip.accessibleName}
              data-query-entry-key={chip.key}
              data-testid={testId}
              style={chipButtonStyle}
              tabIndex={chip.key === resolvedActiveKey ? 0 : -1}
              type="button"
              onBlur={() => setDescribedKey(null)}
              onClick={(event) => onActivate(chip, event.currentTarget)}
              onFocus={() => {
                onRovingEntryChange(chip.key);
                setDescribedKey(chip.key);
              }}
              onKeyDown={(event) => onChipKeyDown(event, chip)}
              onMouseEnter={() => setDescribedKey(chip.key)}
              onMouseLeave={() => setDescribedKey(null)}
            >
              <span aria-hidden="true" style={chipTokenStyle}>
                {chip.token}
              </span>
              <span aria-hidden="true" style={chipDetailStyle}>
                {chip.detail}
              </span>
            </button>
            <span
              id={tooltipId}
              role="tooltip"
              style={
                describedKey === chip.key
                  ? chipTooltipStyle
                  : visuallyHiddenStyle
              }
            >
              {chip.accessibleName}
            </span>
          </span>
        );
      })}
    </div>
  );
}

const chipRailStyle = {
  display: "flex",
  alignItems: "center",
  gap: "var(--ct-spacing-xs)",
  border: 0,
  padding: 0,
  margin: 0,
  inlineSize: "100%",
  maxInlineSize: "100%",
  minWidth: 0,
  minInlineSize: 0,
  flex: "1 1 auto",
  overflow: "visible",
};
const condensedChipRailStyle = { gap: "var(--ct-spacing-xxs)" };
const chipFrameStyle = {
  position: "relative" as const,
  display: "inline-flex",
  flex: "1 1 var(--ct-layout-viewBarQueryChipMaxInlineSize)",
  minInlineSize: "var(--ct-layout-viewBarQueryChipMinInlineSize)",
  maxInlineSize: "var(--ct-layout-viewBarQueryChipMaxInlineSize)",
};
const chipButtonStyle = {
  ...controlButtonStyle,
  background: "var(--ct-colors-surface-3)",
  justifyContent: "flex-start",
  inlineSize: "100%",
  minInlineSize: 0,
  maxInlineSize: "100%",
  overflow: "hidden",
};
const chipTokenStyle = {
  flex: "0 0 auto",
  color: "var(--ct-colors-ink)",
  fontWeight: 700,
};
const chipDetailStyle = {
  display: "block",
  minInlineSize: 0,
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
};
const chipTooltipStyle = {
  position: "absolute" as const,
  zIndex: 40,
  insetBlockStart: "calc(100% + var(--ct-spacing-xs))",
  insetInlineStart: 0,
  inlineSize: "max-content",
  maxInlineSize: "min(var(--ct-layout-viewBarOverlayMaxInlineSize), 88vw)",
  padding: "var(--ct-spacing-xs) var(--ct-spacing-sm)",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-xs)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink)",
  boxShadow: "var(--ct-elevation-panel)",
  overflowWrap: "anywhere" as const,
};
