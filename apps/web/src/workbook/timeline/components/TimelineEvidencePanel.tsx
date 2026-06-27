import {
  timelineEvidenceAttachSectionTestId,
  timelineEvidenceFileInputTestId,
  timelineInspectorSectionTestId,
} from "@cartulary/ui-contracts";
import type { WorkbookRow } from "../models/workbookTimelineModel";

export type TimelineEvidenceCountDisplay = {
  readonly displayCount: string;
  readonly stateKey: string;
};

export type TimelineEvidencePanelProps = {
  readonly countDisplay: TimelineEvidenceCountDisplay;
  readonly row: WorkbookRow;
  readonly onFilesSelected: (
    row: WorkbookRow,
    files: FileList | File[],
  ) => void;
};

export function TimelineEvidencePanel({
  countDisplay,
  row,
  onFilesSelected,
}: TimelineEvidencePanelProps) {
  const recordId = row.recordId;
  if (recordId === null) {
    return null;
  }

  return (
    <section
      data-testid={timelineInspectorSectionTestId("evidence")}
      data-evidence-count-state={countDisplay.stateKey}
      style={inspectorSectionStyle}
      aria-label="Timeline evidence attachment"
      onDragOver={(event) => {
        event.preventDefault();
      }}
      onDrop={(event) => {
        event.preventDefault();
        onFilesSelected(row, event.dataTransfer.files);
      }}
      onPaste={(event) => {
        if (event.clipboardData.files.length > 0) {
          onFilesSelected(row, event.clipboardData.files);
        }
      }}
    >
      <div data-testid={timelineEvidenceAttachSectionTestId(recordId)}>
        <h3 style={sectionTitleStyle}>Evidence</h3>
        <p style={bodyStyle}>
          Attached evidence count: {countDisplay.displayCount}
        </p>
        <label style={labelStyle}>
          Attach file
          <input
            data-testid={timelineEvidenceFileInputTestId(recordId)}
            style={inputStyle}
            type="file"
            accept="image/*,.txt,.pdf,text/plain,application/pdf"
            onChange={(event) => {
              onFilesSelected(row, event.currentTarget.files ?? []);
              event.currentTarget.value = "";
            }}
          />
        </label>
      </div>
    </section>
  );
}

const bodyStyle = {
  margin: 0,
  lineHeight: 1.5,
  color: "var(--ct-colors-ink-muted)",
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

const labelStyle = {
  display: "grid",
  gap: "0.4rem",
  fontSize: "0.95rem",
  color: "var(--ct-colors-ink-muted)",
};

const inspectorSectionStyle = {
  display: "grid",
  gap: "0.75rem",
  marginBottom: "1rem",
};

const sectionTitleStyle = {
  margin: 0,
  fontSize: "1rem",
};
