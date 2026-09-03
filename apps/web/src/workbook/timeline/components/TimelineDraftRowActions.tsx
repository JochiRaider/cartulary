import {
  draftRowCreateButtonTestId,
  timelineDraftEvidenceAttachSectionTestId,
  timelineDraftEvidenceFileInputTestId,
} from "@cartulary/ui-contracts";
import { Paperclip } from "lucide-react";
import {
  type CSSProperties,
  type KeyboardEvent as ReactKeyboardEvent,
  type MouseEvent as ReactMouseEvent,
  useRef,
} from "react";
import type { WorkbookRow } from "../models/timelineRowModel";
import { actionButtonStyle } from "./TimelineWorkbookStyles";

export function DraftRowCreateButton({
  onCreate,
  onFilesSelected,
  row,
}: {
  readonly onCreate: (row: WorkbookRow) => void;
  readonly onFilesSelected?: (
    row: WorkbookRow,
    files: FileList | File[],
  ) => void;
  readonly row: WorkbookRow;
}) {
  const draftEvidenceInputRef = useRef<HTMLInputElement | null>(null);
  const createBlankRow = (
    event:
      | ReactKeyboardEvent<HTMLButtonElement>
      | ReactMouseEvent<HTMLButtonElement>,
  ) => {
    if (event.currentTarget.disabled) return;
    event.preventDefault();
    event.stopPropagation();
    onCreate(row);
  };
  const openDraftEvidencePicker = (
    event: ReactMouseEvent<HTMLButtonElement>,
  ) => {
    if (event.currentTarget.disabled) return;
    event.preventDefault();
    event.stopPropagation();
    draftEvidenceInputRef.current?.click();
  };

  return (
    <span
      data-testid={timelineDraftEvidenceAttachSectionTestId()}
      style={draftRowActionsStyle}
    >
      <button
        aria-label="Create timeline row"
        data-testid={draftRowCreateButtonTestId()}
        disabled={row.pendingSignature !== null}
        style={draftRowCreateButtonStyle}
        type="button"
        onKeyDown={(event) => {
          if (event.key === "Enter" || event.key === " ") {
            createBlankRow(event);
          }
        }}
        onMouseDown={createBlankRow}
      >
        +
      </button>
      {onFilesSelected ? (
        <>
          <button
            aria-label="Attach evidence to draft timeline row"
            disabled={row.pendingSignature !== null}
            style={{
              ...draftRowCreateButtonStyle,
              cursor: row.pendingSignature === null ? "pointer" : "not-allowed",
              opacity: row.pendingSignature === null ? 1 : 0.55,
            }}
            title="Attach evidence"
            type="button"
            onClick={openDraftEvidencePicker}
            onMouseDown={(event) => {
              event.stopPropagation();
            }}
          >
            <Paperclip aria-hidden="true" size={12} />
          </button>
          <input
            ref={draftEvidenceInputRef}
            data-testid={timelineDraftEvidenceFileInputTestId()}
            disabled={row.pendingSignature !== null}
            style={{ display: "none" }}
            type="file"
            accept="image/*,.txt,.pdf,text/plain,application/pdf"
            onChange={(event) => {
              onFilesSelected(row, event.currentTarget.files ?? []);
              event.currentTarget.value = "";
            }}
          />
        </>
      ) : null}
    </span>
  );
}

const draftRowCreateButtonStyle = {
  ...actionButtonStyle,
  display: "inline-grid",
  placeItems: "center",
  inlineSize: "1.4rem",
  blockSize: "1.4rem",
  minInlineSize: "1.4rem",
  padding: 0,
  fontSize: "0.9rem",
  fontWeight: 700,
  lineHeight: 1,
};

const draftRowActionsStyle = {
  display: "inline-grid",
  gridTemplateColumns: "repeat(2, 1.4rem)",
  gap: "0.2rem",
  alignItems: "center",
  justifyContent: "center",
} satisfies CSSProperties;
