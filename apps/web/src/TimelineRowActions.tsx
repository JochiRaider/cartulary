import {
  rowHistoryOpenButtonTestId,
  rowInspectButtonTestId,
  timelineRowMarkReviewedButtonTestId,
  timelineRowReplacementInputTestId,
  timelineRowSupersedeButtonTestId,
} from "@cartulary/ui-contracts";
import { DraftRowCreateButton } from "./TimelineCellEditors";
import type { WorkbookRow } from "./workbookTimelineModel";

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
  if (row.recordId === null) {
    return (
      <div style={actionStackStyle}>
        <DraftRowCreateButton onCreate={onCreateBlankDraftRow} row={row} />
      </div>
    );
  }

  return (
    <div style={actionStackStyle}>
      <div style={timelineActionTopRowStyle}>
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
      </div>
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
  );
}

const actionStackStyle = {
  display: "grid",
  gap: "0.5rem",
};

const timelineActionTopRowStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(2, minmax(0, 1fr))",
  gap: "0.35rem",
  alignItems: "center",
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
