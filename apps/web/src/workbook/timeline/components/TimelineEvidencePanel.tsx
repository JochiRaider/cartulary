import {
  timelineEvidenceAttachSectionTestId,
  timelineEvidenceFileInputTestId,
  timelineInspectorSectionTestId,
} from "@cartulary/ui-contracts";
import type { WorkbookRow } from "../models/workbookTimelineModel";
import {
  bodyStyle,
  inputStyle,
  inspectorSectionStyle,
  labelStyle,
  sectionTitleStyle,
} from "./TimelineWorkbookStyles";

type TimelineEvidenceCountDisplay = {
  readonly displayCount: string;
  readonly stateKey: string;
};

type TimelineEvidencePanelProps = {
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
      tabIndex={-1}
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
