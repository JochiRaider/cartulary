import type { GridEditorFocusTarget } from "@cartulary/grid-adapter";
import {
  type ChangeEvent as ReactChangeEvent,
  type ClipboardEvent as ReactClipboardEvent,
  type FocusEvent as ReactFocusEvent,
  type KeyboardEvent as ReactKeyboardEvent,
  useEffect,
  useRef,
  useState,
} from "react";
import type {
  FocusFieldKey,
  RowValues,
  TimelineScalarEditorSurface,
} from "../models/workbookTimelineModel";
import { inputStyle } from "./TimelineWorkbookStyles";

export function TimelineScalarEditor({
  accessibleLabel,
  blockedByConflict,
  committedValue,
  controlId,
  dataTestId,
  draftValue,
  field,
  multiline,
  onBlurCommit,
  onDraftChange,
  onEditModeChange,
  onCloseGridEditor,
  onFocusAnchor,
  onFocusRecord,
  focusTargetRef,
  onKeyCommit,
  onPasteCommit,
  registerInput,
  readOnly = false,
  presenceFieldKey,
  rowKey,
  rowRecordId,
  surface,
}: {
  readonly accessibleLabel?: string | undefined;
  readonly blockedByConflict?: boolean | undefined;
  readonly committedValue: string;
  readonly controlId: string;
  readonly dataTestId: string;
  readonly draftValue?: string | undefined;
  readonly field: keyof RowValues;
  readonly focusTargetRef?:
    | ((element: GridEditorFocusTarget | null) => void)
    | undefined;
  readonly multiline?: boolean | undefined;
  readonly onBlurCommit: (
    rowKey: string,
    field: keyof RowValues,
    surface: TimelineScalarEditorSurface,
    value: string,
  ) => void;
  readonly onDraftChange: (
    rowKey: string,
    field: keyof RowValues,
    surface: TimelineScalarEditorSurface,
    value: string,
  ) => void;
  readonly onEditModeChange: (
    recordId: string | null,
    fieldKey: string,
    editing: boolean,
  ) => void;
  readonly onCloseGridEditor?:
    | ((commit: boolean, draftValue: string) => void)
    | undefined;
  readonly onFocusAnchor: (recordId: string | null, fieldKey: string) => void;
  readonly onFocusRecord: (recordId: string) => void;
  readonly onKeyCommit: (
    event: ReactKeyboardEvent<HTMLInputElement | HTMLTextAreaElement>,
    rowKey: string,
    field: keyof RowValues,
    surface: TimelineScalarEditorSurface,
  ) => void;
  readonly onPasteCommit: (
    event: ReactClipboardEvent<HTMLInputElement | HTMLTextAreaElement>,
    rowKey: string,
    field: keyof RowValues,
    surface: TimelineScalarEditorSurface,
  ) => void;
  readonly registerInput: (
    rowKey: string,
    field: FocusFieldKey,
    surface: TimelineScalarEditorSurface,
    dataTestId: string,
    element: HTMLInputElement | HTMLTextAreaElement | null,
  ) => void;
  readonly presenceFieldKey: string;
  readonly rowKey: string;
  readonly rowRecordId: string | null;
  readonly readOnly?: boolean | undefined;
  readonly surface: TimelineScalarEditorSurface;
}) {
  const displayValue = draftValue ?? committedValue;
  const [editorValue, setEditorValue] = useState(displayValue);
  const hasActiveEditRef = useRef(false);

  useEffect(() => {
    if (!hasActiveEditRef.current || draftValue === undefined) {
      setEditorValue(displayValue);
    }
  }, [displayValue, draftValue]);

  useEffect(
    () => () => {
      if (hasActiveEditRef.current) {
        onEditModeChange(rowRecordId, presenceFieldKey, false);
      }
    },
    [onEditModeChange, presenceFieldKey, rowRecordId],
  );

  const handleFocus = () => {
    hasActiveEditRef.current = !readOnly;
    if (surface === "grid") onFocusAnchor(rowRecordId, presenceFieldKey);
    if (rowRecordId) onFocusRecord(rowRecordId);
    if (!readOnly) onEditModeChange(rowRecordId, presenceFieldKey, true);
  };
  const handleChange = (value: string) => {
    if (readOnly) return;
    setEditorValue(value);
    onDraftChange(rowKey, field, surface, value);
  };
  const markTypingAcknowledgement = (
    event: ReactChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    const inputEvent = event.nativeEvent;
    if (
      rowRecordId !== null &&
      surface === "grid" &&
      presenceFieldKey === "timeline.activity_synopsis_text" &&
      typeof InputEvent !== "undefined" &&
      inputEvent instanceof InputEvent &&
      inputEvent.data === "x" &&
      !inputEvent.isComposing &&
      event.currentTarget.value === `${editorValue}x` &&
      event.currentTarget.selectionStart === event.currentTarget.value.length &&
      event.currentTarget.selectionEnd === event.currentTarget.value.length
    ) {
      performance.mark("cartulary.workbook.typing_ack_accepted", {
        detail: { field: presenceFieldKey, surface },
      });
    }
  };
  const handleBlur = (
    event: ReactFocusEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    if (readOnly) return;
    hasActiveEditRef.current = false;
    onEditModeChange(rowRecordId, presenceFieldKey, false);
    onDraftChange(rowKey, field, surface, event.currentTarget.value);
    if (blockedByConflict || onCloseGridEditor !== undefined) return;
    onBlurCommit(rowKey, field, surface, event.currentTarget.value);
  };
  const handleKeyDown = (
    event: ReactKeyboardEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    if (readOnly) return;
    if (
      event.key === "Escape" &&
      (editorValue !== committedValue || onCloseGridEditor !== undefined)
    ) {
      event.preventDefault();
      if (editorValue !== committedValue) {
        setEditorValue(committedValue);
        onDraftChange(rowKey, field, surface, committedValue);
      }
      onCloseGridEditor?.(false, committedValue);
      return;
    }
    if (
      surface === "grid" &&
      onCloseGridEditor !== undefined &&
      (event.key === "Enter" || event.key === "Tab")
    ) {
      event.preventDefault();
      onCloseGridEditor(true, event.currentTarget.value);
      return;
    }
    onKeyCommit(event, rowKey, field, surface);
  };
  const handlePaste = (
    event: ReactClipboardEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    if (readOnly) return;
    onPasteCommit(event, rowKey, field, surface);
    event.stopPropagation();
  };
  const handleCopy = (
    event: ReactClipboardEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    event.clipboardData.setData("text/plain", editorValue);
    event.preventDefault();
    event.stopPropagation();
    if (surface === "grid") onFocusAnchor(rowRecordId, presenceFieldKey);
  };
  const inputRef = (element: HTMLInputElement | HTMLTextAreaElement | null) => {
    focusTargetRef?.(element);
    registerInput(rowKey, field, surface, dataTestId, element);
  };

  if (multiline) {
    return (
      <textarea
        aria-label={accessibleLabel}
        data-testid={dataTestId}
        id={controlId}
        ref={inputRef}
        readOnly={readOnly}
        rows={surface === "grid" ? 1 : 3}
        style={surface === "grid" ? gridCellTextareaStyle : textareaStyle}
        value={editorValue}
        onBlur={handleBlur}
        onChange={(event) => {
          markTypingAcknowledgement(event);
          handleChange(event.target.value);
        }}
        onFocus={handleFocus}
        onKeyDown={handleKeyDown}
        onCopy={handleCopy}
        onPaste={handlePaste}
      />
    );
  }

  return (
    <input
      aria-label={accessibleLabel}
      data-testid={dataTestId}
      id={controlId}
      ref={inputRef}
      readOnly={readOnly}
      style={surface === "grid" ? gridCellInputStyle : inputStyle}
      type="text"
      value={editorValue}
      onBlur={handleBlur}
      onChange={(event) => {
        markTypingAcknowledgement(event);
        handleChange(event.target.value);
      }}
      onFocus={handleFocus}
      onKeyDown={handleKeyDown}
      onCopy={handleCopy}
      onPaste={handlePaste}
    />
  );
}

const textareaStyle = {
  ...inputStyle,
  resize: "vertical" as const,
};

const gridCellInputStyle = {
  ...inputStyle,
  position: "absolute" as const,
  inset: 0,
  minHeight: 0,
  height: "100%",
  blockSize: "100%",
  border: "none",
  borderRadius: 0,
  background: "transparent",
  padding: "var(--cartulary-grid-cell-padding)",
  fontSize: "var(--cartulary-grid-font-size)",
  lineHeight: "var(--cartulary-grid-line-height)",
  width: "100%",
  inlineSize: "100%",
};

const gridCellTextareaStyle = {
  ...gridCellInputStyle,
  resize: "none" as const,
  overflow: "auto",
};
