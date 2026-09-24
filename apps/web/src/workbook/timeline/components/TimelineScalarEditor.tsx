import {
  clipboardTextWithinLimit,
  type GridEditorFocusTarget,
} from "@cartulary/grid-adapter";
import {
  type ClipboardEvent as ReactClipboardEvent,
  type FocusEvent as ReactFocusEvent,
  type FormEvent as ReactFormEvent,
  type KeyboardEvent as ReactKeyboardEvent,
  useCallback,
  useEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import type {
  FocusFieldKey,
  RowValues,
  TimelineScalarEditorSurface,
} from "../models/timelineFieldRegistry";
import { inputStyle } from "./TimelineWorkbookStyles";

export function TimelineScalarEditor({
  accessibleLabel,
  blockedByConflict,
  committedValue,
  controlId,
  dataTestId,
  editorDraftRegistry,
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
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
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
    input?: { readonly composing: boolean; readonly pasteCompleted: boolean },
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
  readonly registerInput: (
    rowKey: string,
    field: FocusFieldKey,
    surface: TimelineScalarEditorSurface,
    element: HTMLInputElement | HTMLTextAreaElement | null,
  ) => void;
  readonly presenceFieldKey: string;
  readonly rowKey: string;
  readonly rowRecordId: string | null;
  readonly readOnly?: boolean | undefined;
  readonly surface: TimelineScalarEditorSurface;
}) {
  const editorValue = useSyncExternalStore(
    editorDraftRegistry.subscribe,
    useCallback(
      () =>
        editorDraftRegistry.draftValue({ rowKey, field, surface }) ??
        committedValue,
      [editorDraftRegistry, rowKey, field, surface, committedValue],
    ),
  );
  const hasActiveEditRef = useRef(false);
  const [clipboardError, setClipboardError] = useState<string | null>(null);

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
    // Managed grid editors publish cell focus/presence. Explicit pointer
    // inspection is already owned by the Adapter's admitted click.
    if (rowRecordId && onCloseGridEditor === undefined)
      onFocusRecord(rowRecordId);
    if (!readOnly) onEditModeChange(rowRecordId, presenceFieldKey, true);
  };
  const markTypingAcknowledgement = (
    event: ReactFormEvent<HTMLInputElement | HTMLTextAreaElement>,
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
  const handleInput = (
    event: ReactFormEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    if (readOnly) return;
    markTypingAcknowledgement(event);
    const native = event.nativeEvent;
    const value = event.currentTarget.value;
    setClipboardError(null);
    onDraftChange(rowKey, field, surface, value, {
      composing:
        editorDraftRegistry.isComposing(rowKey) ||
        (native instanceof InputEvent && native.isComposing),
      pasteCompleted:
        native instanceof InputEvent && native.inputType === "insertFromPaste",
    });
  };
  const handleBlur = (
    event: ReactFocusEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    if (readOnly) return;
    hasActiveEditRef.current = false;
    onEditModeChange(rowRecordId, presenceFieldKey, false);
    if (
      blockedByConflict ||
      onCloseGridEditor !== undefined ||
      (event.relatedTarget instanceof Element &&
        event.relatedTarget.closest(
          '[data-grid-editor-external-action="true"]',
        ))
    )
      return;
    onBlurCommit(rowKey, field, surface, event.currentTarget.value);
  };
  const handleKeyDown = (
    event: ReactKeyboardEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    if (event.defaultPrevented) return;
    if (
      event.nativeEvent.isComposing ||
      (multiline && event.key === "Enter" && event.shiftKey)
    ) {
      event.stopPropagation();
      return;
    }
    if (readOnly) return;
    if (
      event.key === "Escape" &&
      (editorValue !== committedValue || onCloseGridEditor !== undefined)
    ) {
      event.preventDefault();
      if (editorValue !== committedValue) {
        onDraftChange(rowKey, field, surface, committedValue);
      }
      onCloseGridEditor?.(false, committedValue);
      return;
    }
    onKeyCommit(event, rowKey, field, surface);
  };
  const handlePaste = (
    event: ReactClipboardEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    event.stopPropagation();
    if (readOnly) return;
    // Only admission is ours; the browser owns insertion and its editing history.
    if (!clipboardTextWithinLimit(event.clipboardData.getData("text/plain"))) {
      event.preventDefault();
      setClipboardError("Clipboard text exceeds the 8 MiB limit.");
    }
  };
  const isolateClipboard = (
    event: ReactClipboardEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => event.stopPropagation();
  const inputRef = (element: HTMLInputElement | HTMLTextAreaElement | null) => {
    focusTargetRef?.(element);
    registerInput(rowKey, field, surface, element);
  };

  const controlProps = {
    "aria-label": accessibleLabel,
    "aria-describedby": clipboardError
      ? `${controlId}-clipboard-error`
      : undefined,
    "data-testid": dataTestId,
    id: controlId,
    ref: inputRef,
    readOnly,
    value: editorValue,
    onBlur: handleBlur,
    onInput: handleInput,
    onCompositionStart: (
      event: ReactFormEvent<HTMLInputElement | HTMLTextAreaElement>,
    ) => {
      if (readOnly) return;
      editorDraftRegistry.beginComposition(rowKey);
      onDraftChange(rowKey, field, surface, event.currentTarget.value, {
        composing: true,
        pasteCompleted: false,
      });
    },
    onCompositionEnd: (
      event: ReactFormEvent<HTMLInputElement | HTMLTextAreaElement>,
    ) => {
      if (!readOnly)
        onDraftChange(rowKey, field, surface, event.currentTarget.value, {
          composing: false,
          pasteCompleted: false,
        });
      editorDraftRegistry.endComposition(rowKey);
    },
    onFocus: handleFocus,
    onKeyDown: handleKeyDown,
    onCopy: isolateClipboard,
    onCut: isolateClipboard,
    onPaste: handlePaste,
  };
  return (
    <>
      {multiline ? (
        <textarea
          {...controlProps}
          rows={surface === "grid" ? 1 : 3}
          style={surface === "grid" ? gridCellTextareaStyle : textareaStyle}
        />
      ) : (
        <input
          {...controlProps}
          type="text"
          style={surface === "grid" ? gridCellInputStyle : inputStyle}
        />
      )}
      {clipboardError && (
        <span
          id={`${controlId}-clipboard-error`}
          role="alert"
          data-grid-editor-toolbar={surface === "grid" ? "true" : undefined}
        >
          {clipboardError}
        </span>
      )}
    </>
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
