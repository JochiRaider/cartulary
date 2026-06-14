import {
  draftRowCreateButtonTestId,
  relationshipChipTestId,
} from "@cartulary/ui-contracts";
import {
  type ClipboardEvent as ReactClipboardEvent,
  type FocusEvent as ReactFocusEvent,
  type KeyboardEvent as ReactKeyboardEvent,
  type MouseEvent as ReactMouseEvent,
  useEffect,
  useRef,
  useState,
} from "react";
import type {
  CollectionItem,
  InspectorMention,
  MentionChipState,
} from "./workbookMentionChips";
import type {
  FocusFieldKey,
  RowValues,
  WorkbookRow,
} from "./workbookTimelineModel";

type TimelineScalarEditorSurface = "grid" | "inspector";

export function relationshipItemLabel(
  item: CollectionItem | InspectorMention,
  entityIndex: Record<string, { label: string }>,
) {
  if ("status" in item && item.status === "dismissed") {
    return item.displayText || item.rawText;
  }
  if (item.resolvedRecordId) {
    const resolvedEntity = entityIndex[item.resolvedRecordId];
    if (resolvedEntity) {
      return resolvedEntity.label;
    }
  }
  return item.displayText || item.rawText;
}

export function mentionChipStateForItem(
  item: CollectionItem | InspectorMention,
): MentionChipState {
  if ("chipState" in item) {
    return item.chipState;
  }
  if (item.itemKind !== "resolved_ref") {
    return "unresolved";
  }
  if (item.autoResolved) {
    return "auto-resolved";
  }
  if (item.resolutionMethod === "explicit_resolve_route") {
    return "manual-resolution";
  }
  return "resolved";
}

export function RelationshipChip({
  item,
  entityIndex,
  onSelect,
  selected = false,
}: {
  item: CollectionItem | InspectorMention;
  entityIndex: Record<string, { label: string }>;
  onSelect?: () => void;
  selected?: boolean;
}) {
  const label = relationshipItemLabel(item, entityIndex);
  const chipState = mentionChipStateForItem(item);
  const isResolved =
    chipState === "resolved" ||
    chipState === "auto-resolved" ||
    chipState === "manual-resolution";
  const isDismissed = chipState === "dismissed";
  const isAutoResolved = chipState === "auto-resolved";
  const chipStyle = {
    ...relationshipChipStyle,
    ...(isDismissed
      ? dismissedChipStyle
      : isResolved
        ? isAutoResolved
          ? autoResolvedChipStyle
          : resolvedChipStyle
        : unresolvedChipStyle),
    ...(selected ? selectedChipStyle : null),
  };
  const labelPrefix =
    chipState === "manual-resolution"
      ? "Resolved"
      : chipState === "auto-resolved"
        ? "Auto-resolved"
        : chipState === "dismissed"
          ? "Dismissed"
          : chipState === "resolved"
            ? "Resolved"
            : "Unresolved";
  const stateDetail =
    chipState === "manual-resolution"
      ? "; manual resolution"
      : chipState === "auto-resolved" && item.matchedAliasText
        ? `; matched ${item.matchedAliasText}`
        : "";
  const accessibleLabel = `${labelPrefix} ${label}${stateDetail}`;
  const content = (
    <RelationshipChipContent
      chipState={chipState}
      isAutoResolved={isAutoResolved}
      isDismissed={isDismissed}
      isResolved={isResolved}
      label={label}
    />
  );

  return onSelect ? (
    <button
      aria-label={accessibleLabel}
      data-testid={relationshipChipTestId(item.itemRef)}
      tabIndex={0}
      style={chipStyle}
      type="button"
      onClick={onSelect}
    >
      {content}
    </button>
  ) : (
    <span
      aria-label={accessibleLabel}
      data-testid={relationshipChipTestId(item.itemRef)}
      role="note"
      style={chipStyle}
    >
      {content}
    </span>
  );
}

function RelationshipChipContent({
  chipState,
  isAutoResolved,
  isDismissed,
  isResolved,
  label,
}: {
  chipState: MentionChipState;
  isAutoResolved: boolean;
  isDismissed: boolean;
  isResolved: boolean;
  label: string;
}) {
  return (
    <>
      <span>{label}</span>
      {isAutoResolved ? (
        <span data-density-role="narrow-metadata" style={chipMetaStyle}>
          Auto
        </span>
      ) : null}
      {chipState === "manual-resolution" ? (
        <span data-density-role="narrow-metadata" style={chipMetaStyle}>
          Manual
        </span>
      ) : null}
      {chipState === "resolved" ? (
        <span data-density-role="narrow-metadata" style={chipMetaStyle}>
          Resolved
        </span>
      ) : null}
      {!isResolved && !isDismissed ? (
        <span data-density-role="narrow-metadata" style={chipMetaStyle}>
          Unresolved
        </span>
      ) : null}
      {isDismissed ? (
        <span data-density-role="narrow-metadata" style={chipMetaStyle}>
          Dismissed
        </span>
      ) : null}
    </>
  );
}

export function DraftRowCreateButton({
  onCreate,
  row,
}: {
  readonly onCreate: (row: WorkbookRow) => void;
  readonly row: WorkbookRow;
}) {
  const createBlankRow = (
    event:
      | ReactKeyboardEvent<HTMLButtonElement>
      | ReactMouseEvent<HTMLButtonElement>,
  ) => {
    if (event.currentTarget.disabled) {
      return;
    }
    event.preventDefault();
    event.stopPropagation();
    onCreate(row);
  };

  return (
    <button
      data-testid={draftRowCreateButtonTestId()}
      disabled={row.pendingSignature !== null}
      style={actionButtonStyle}
      type="button"
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") {
          createBlankRow(event);
        }
      }}
      onMouseDown={createBlankRow}
    >
      Create blank row
    </button>
  );
}

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
  onFocusAnchor,
  onFocusRecord,
  onKeyCommit,
  onPasteCommit,
  registerInput,
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

  const handleFocus = () => {
    hasActiveEditRef.current = true;
    if (surface === "grid") {
      onFocusAnchor(rowRecordId, presenceFieldKey);
    }
    if (rowRecordId) {
      onFocusRecord(rowRecordId);
    }
    onEditModeChange(rowRecordId, presenceFieldKey, true);
  };
  const handleChange = (value: string) => {
    setEditorValue(value);
    onDraftChange(rowKey, field, surface, value);
  };
  const handleBlur = (
    event: ReactFocusEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    hasActiveEditRef.current = false;
    onEditModeChange(rowRecordId, presenceFieldKey, false);
    onDraftChange(rowKey, field, surface, event.currentTarget.value);
    if (blockedByConflict) {
      return;
    }
    onBlurCommit(rowKey, field, surface, event.currentTarget.value);
  };
  const handleKeyDown = (
    event: ReactKeyboardEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    if (event.key === "Escape" && editorValue !== committedValue) {
      event.preventDefault();
      setEditorValue(committedValue);
      onDraftChange(rowKey, field, surface, committedValue);
      return;
    }
    onKeyCommit(event, rowKey, field, surface);
  };
  const handlePaste = (
    event: ReactClipboardEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    onPasteCommit(event, rowKey, field, surface);
  };
  const handleCopy = (
    event: ReactClipboardEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    event.clipboardData.setData("text/plain", editorValue);
    event.preventDefault();
    if (surface === "grid") {
      onFocusAnchor(rowRecordId, presenceFieldKey);
    }
  };
  const inputRef = (element: HTMLInputElement | HTMLTextAreaElement | null) => {
    registerInput(rowKey, field, surface, dataTestId, element);
  };

  if (multiline) {
    return (
      <textarea
        aria-label={accessibleLabel}
        data-testid={dataTestId}
        id={controlId}
        ref={inputRef}
        rows={3}
        style={surface === "grid" ? gridCellTextareaStyle : textareaStyle}
        value={editorValue}
        onBlur={handleBlur}
        onChange={(event) => {
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
      style={surface === "grid" ? gridCellInputStyle : inputStyle}
      type="text"
      value={editorValue}
      onBlur={handleBlur}
      onChange={(event) => {
        handleChange(event.target.value);
      }}
      onFocus={handleFocus}
      onKeyDown={handleKeyDown}
      onCopy={handleCopy}
      onPaste={handlePaste}
    />
  );
}

const actionButtonStyle = {
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  padding: "0.55rem 0.9rem",
  font: "inherit",
  cursor: "pointer",
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

const textareaStyle = {
  ...inputStyle,
  resize: "vertical" as const,
};

const gridCellInputStyle = {
  ...inputStyle,
  minHeight: "1.35rem",
  border: "none",
  background: "transparent",
  padding: 0,
  lineHeight: "1.25rem",
  width: "100%",
};

const gridCellTextareaStyle = {
  ...gridCellInputStyle,
  resize: "vertical" as const,
  minHeight: "3.5rem",
};

const relationshipChipStyle = {
  display: "inline-flex",
  alignItems: "center",
  gap: "0.35rem",
  borderRadius: "var(--ct-component-chip-rounded)",
  padding: "var(--ct-component-chip-padding)",
  font: "inherit",
  lineHeight: 1.2,
  maxWidth: "100%",
  minWidth: 0,
  overflowWrap: "anywhere" as const,
};

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

const chipMetaStyle = {
  fontSize: "0.72rem",
  textTransform: "uppercase" as const,
  letterSpacing: "0.04em",
};
