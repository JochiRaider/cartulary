import {
  rowHistoryOpenButtonTestId,
  rowInspectButtonTestId,
  timelineRowMarkReviewedButtonTestId,
  timelineRowReplacementInputTestId,
  timelineRowSupersedeButtonTestId,
  workbookRowActionMenuButtonTestId,
} from "@cartulary/ui-contracts";
import { MoreHorizontal } from "lucide-react";
import { useState } from "react";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookRow } from "../models/workbookTimelineModel";
import { DraftRowCreateButton } from "./TimelineCellEditors";

export type TimelineRowActionsProps = {
  readonly replacementDraft: string;
  readonly row: WorkbookRow;
  readonly onCreateBlankDraftRow: (row: WorkbookRow) => void;
  readonly onInspectRow: (recordId: string) => void;
  readonly onMarkReviewed: (rowKey: string) => void;
  readonly onOpenHistory: (recordId: string) => void;
  readonly onReplacementDraftChange: (rowKey: string, value: string) => void;
  readonly onSupersede: (rowKey: string) => void;
};

export function TimelineRowActions({
  replacementDraft,
  row,
  onCreateBlankDraftRow,
  onInspectRow,
  onMarkReviewed,
  onOpenHistory,
  onReplacementDraftChange,
  onSupersede,
}: TimelineRowActionsProps) {
  const [isOpen, setIsOpen] = useState(false);
  if (row.recordId === null) {
    return (
      <div style={actionStackStyle}>
        <DraftRowCreateButton onCreate={onCreateBlankDraftRow} row={row} />
      </div>
    );
  }

  return (
    <div style={menuShellStyle}>
      <button
        aria-expanded={isOpen}
        aria-label="Row actions"
        data-testid={workbookRowActionMenuButtonTestId(
          timelineViewSchemaId,
          row.recordId,
        )}
        style={rowMenuButtonStyle}
        type="button"
        onClick={() => {
          setIsOpen((current) => !current);
        }}
      >
        <MoreHorizontal aria-hidden="true" size={16} />
      </button>
      <div
        aria-hidden={!isOpen}
        style={{
          ...actionPopoverStyle,
          ...(isOpen ? null : hiddenActionPopoverStyle),
        }}
      >
        <button
          data-testid={rowInspectButtonTestId(row.recordId)}
          style={timelineActionButtonStyle}
          type="button"
          onClick={() => {
            onInspectRow(row.recordId ?? "");
          }}
        >
          Inspect
        </button>
        <button
          data-testid={rowHistoryOpenButtonTestId(row.recordId ?? "")}
          style={timelineActionButtonStyle}
          type="button"
          onClick={() => {
            onOpenHistory(row.recordId ?? "");
          }}
        >
          History
        </button>
        <button
          data-testid={timelineRowMarkReviewedButtonTestId(row.recordId)}
          disabled={
            row.captureState === "reviewed" || row.captureState === "superseded"
          }
          style={timelineActionButtonStyle}
          type="button"
          onClick={() => {
            onMarkReviewed(row.key);
          }}
        >
          Mark reviewed
        </button>
        <input
          data-testid={timelineRowReplacementInputTestId(row.recordId)}
          placeholder="Replacement record id"
          style={timelineReplacementInputStyle}
          type="text"
          value={replacementDraft}
          onChange={(event) => {
            onReplacementDraftChange(row.key, event.target.value);
          }}
        />
        <button
          data-testid={timelineRowSupersedeButtonTestId(row.recordId)}
          disabled={
            row.captureState === "superseded" || replacementDraft.trim() === ""
          }
          style={timelineActionButtonStyle}
          type="button"
          onClick={() => {
            onSupersede(row.key);
          }}
        >
          Supersede
        </button>
      </div>
    </div>
  );
}

const actionStackStyle = {
  display: "grid",
  gap: "0.5rem",
};

const menuShellStyle = {
  position: "relative" as const,
  display: "grid",
  placeItems: "center",
  minWidth: 0,
};

const rowMenuButtonStyle = {
  display: "inline-grid",
  placeItems: "center",
  inlineSize: "1.75rem",
  blockSize: "1.75rem",
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink-muted)",
  cursor: "pointer",
};

const actionPopoverStyle = {
  position: "absolute" as const,
  insetBlockStart: "calc(100% + 0.25rem)",
  insetInlineEnd: 0,
  zIndex: 20,
  display: "grid",
  gap: "0.35rem",
  minWidth: "13rem",
  padding: "0.45rem",
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  boxShadow: "var(--ct-elevation-popover)",
};

const hiddenActionPopoverStyle = {
  visibility: "hidden" as const,
  pointerEvents: "none" as const,
};

const inputStyle = {
  boxSizing: "border-box" as const,
  display: "block",
  minWidth: 0,
  width: "100%",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  padding: "0.65rem 0.75rem",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
};

const replacementInputStyle = {
  ...inputStyle,
  fontSize: "0.9rem",
};

const timelineReplacementInputStyle = {
  ...replacementInputStyle,
  boxSizing: "border-box" as const,
  fontSize: "0.82rem",
  width: "100%",
};

const actionButtonStyle = {
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  padding: "0.55rem 0.9rem",
  font: "inherit",
  cursor: "pointer",
};

const timelineActionButtonStyle = {
  ...actionButtonStyle,
  boxSizing: "border-box" as const,
  fontSize: "0.85rem",
  lineHeight: 1.1,
  padding: "0.45rem 0.3rem",
  width: "100%",
};
