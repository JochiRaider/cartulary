import type {
  WorkbookGridQueryCommand,
  WorkbookGridQueryControlProjection,
} from "../models/workbookGridQueryControls";
import { visuallyHiddenStyle } from "../utils/workbookStyles";
import {
  clearButtonStyle,
  controlButtonStyle,
} from "./workbookGridControlStyles";

export function WorkbookActiveQueryChips({
  condensed,
  onCommand,
  projection,
}: {
  readonly condensed: boolean;
  readonly onCommand: (command: WorkbookGridQueryCommand) => void;
  readonly projection: WorkbookGridQueryControlProjection;
}) {
  return (
    <div
      aria-label="Active query chips"
      hidden={projection.visibleChipCapacity === 0}
      role="toolbar"
      style={{
        ...chipRailStyle,
        ...(condensed ? condensedChipRailStyle : null),
      }}
    >
      <span style={visuallyHiddenStyle}>Active query chips</span>
      {projection.visibleChips.map((chip) => (
        <button
          key={chip.key}
          aria-label={chip.label}
          data-testid={chip.testId}
          style={chipButtonStyle}
          title={chip.label}
          type="button"
          onClick={() => {
            onCommand(chip.command);
          }}
        >
          <span style={chipLabelStyle}>{chip.label}</span>
        </button>
      ))}
      {projection.chips.length > 1 && projection.hiddenChips.length === 0 ? (
        <button
          style={clearButtonStyle}
          type="button"
          onClick={() => {
            onCommand({ kind: "query_clear" });
          }}
        >
          Clear all
        </button>
      ) : null}
    </div>
  );
}

const chipRailStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.3rem",
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
const condensedChipRailStyle = { gap: "0.2rem" };
const chipButtonStyle = {
  ...controlButtonStyle,
  background: "var(--ct-colors-surface-3)",
  justifyContent: "flex-start",
  flex: "1 1 0",
  minInlineSize: "1.8rem",
  maxInlineSize: "7rem",
  overflow: "hidden",
};
const chipLabelStyle = {
  display: "block",
  minInlineSize: 0,
  maxInlineSize: "100%",
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
};
