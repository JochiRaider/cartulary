import {
  draftRelationshipItemsTestId,
  draftTimelineCollectionInputTestId,
  relationshipItemsTestId,
  relationshipOverflowButtonTestId,
  timelineCollectionInputTestId,
} from "@cartulary/ui-contracts";
import {
  WorkbookRelationshipChip,
  workbookRelationshipChipBaseStyle,
} from "../../components/WorkbookRelationshipChip";
import { visuallyHiddenStyle } from "../../utils/workbookStyles";
import {
  projectTimelineCollectionPresentation,
  type TimelineCollectionPresentation,
} from "../models/timelineCollectionPresentation";
import {
  inputFocusKey,
  type TimelineCollectionBinding,
} from "../models/timelineFieldRegistry";
import type { WorkbookRow } from "../models/timelineRowModel";
import type {
  RegisterTimelineInput,
  TimelineCollectionKeyDown,
  TimelineCollectionSave,
  TimelineEntityIndex,
} from "./TimelineWorkbookRendererTypes";
import { inputStyle } from "./TimelineWorkbookStyles";

type TimelineCollectionCellProps = {
  readonly activateCollectionInput: (focusKey: string) => void;
  readonly activeCollectionInputKey: string | null;
  readonly binding: TimelineCollectionBinding;
  readonly deactivateCollectionInput: (focusKey: string) => void;
  readonly entityIndex: TimelineEntityIndex;
  readonly focusTargetRef?: (element: HTMLInputElement | null) => void;
  readonly handleCollectionInputChange: (
    focusKey: string,
    value: string,
  ) => void;
  readonly handleCollectionKeyDown: TimelineCollectionKeyDown;
  readonly handleSelectMention: (recordId: string, itemRef: string) => void;
  readonly handleSelectRow: (recordId: string) => void;
  readonly label: string;
  readonly openInspectorForRow: (recordId: string) => void;
  readonly queueCollectionSave: TimelineCollectionSave;
  readonly readOnly: boolean;
  readonly registerInput: RegisterTimelineInput;
  readonly resolveInputElement: (
    focusKey: string,
  ) => HTMLInputElement | HTMLTextAreaElement | null;
  readonly row: WorkbookRow;
  readonly updateTimelineSurfaceFocusAnchor: (
    recordId: string | null,
    fieldKey: string,
  ) => void;
};

function TimelineCollectionSummary({
  fieldKey,
  handleSelectMention,
  label,
  openInspectorForRow,
  presentation,
  recordId,
}: {
  readonly fieldKey: string;
  readonly handleSelectMention: (recordId: string, itemRef: string) => void;
  readonly label: string;
  readonly openInspectorForRow: (recordId: string) => void;
  readonly presentation: TimelineCollectionPresentation;
  readonly recordId: string | null;
}) {
  const visible =
    presentation.visibleItems.length < 1 ? (
      <span style={emptyRelationshipStyle}>No items</span>
    ) : presentation.kind === "relationship" ? (
      presentation.visibleItems.map((item) => (
        <WorkbookRelationshipChip
          key={item.itemRef}
          presentation={{
            ...item.chip,
            onSelect: () => {
              if (recordId !== null)
                handleSelectMention(recordId, item.itemRef);
            },
          }}
        />
      ))
    ) : (
      presentation.visibleItems.map((item) => (
        <span key={item.itemRef} style={tagChipStyle} title={item.displayText}>
          {item.displayText}
        </span>
      ))
    );
  if (presentation.hiddenItemCount < 1) return visible;
  const overflowLabel = `${presentation.hiddenItemCount} more ${label.toLowerCase()}`;
  const overflowRecordId = presentation.overflowRecordId;
  return (
    <>
      {visible}
      {overflowRecordId === null ? (
        <span
          aria-label={overflowLabel}
          role="note"
          style={collectionOverflowStyle}
          title={overflowLabel}
        >
          +{presentation.hiddenItemCount}
        </span>
      ) : (
        <button
          aria-label={`Inspect ${overflowLabel}`}
          data-testid={relationshipOverflowButtonTestId(
            overflowRecordId,
            fieldKey,
          )}
          style={collectionOverflowButtonStyle}
          title={`Inspect ${overflowLabel}`}
          type="button"
          onClick={(event) => {
            event.stopPropagation();
            if (presentation.firstHiddenItemRef === null) {
              openInspectorForRow(overflowRecordId);
            } else {
              handleSelectMention(
                overflowRecordId,
                presentation.firstHiddenItemRef,
              );
            }
          }}
          onKeyDown={(event) => event.stopPropagation()}
        >
          +{presentation.hiddenItemCount}
        </button>
      )}
      <span style={visuallyHiddenStyle}>
        {presentation.hiddenLabels.join(" ")}
      </span>
    </>
  );
}

export function TimelineCollectionCell(props: TimelineCollectionCellProps) {
  const { binding, label, row } = props;
  const collectionFocusKey = inputFocusKey(row.key, binding.draftKey, "grid");
  const presentation = projectTimelineCollectionPresentation({
    binding,
    entityIndex: props.entityIndex,
    row,
  });
  const isInputActive =
    props.activeCollectionInputKey === collectionFocusKey ||
    row.collectionDrafts[binding.draftKey] !== "";
  const activateInput = () => {
    if (props.readOnly) return;
    props.activateCollectionInput(collectionFocusKey);
    props
      .resolveInputElement(collectionFocusKey)
      ?.focus({ preventScroll: true });
  };
  return (
    <fieldset
      aria-label={`${label} collection cell`}
      style={collectionCellStyle}
      onClick={(event) => {
        if (
          props.readOnly ||
          (event.target instanceof HTMLElement &&
            event.target.closest("[data-relationship-chip='true']") !== null)
        ) {
          return;
        }
        activateInput();
      }}
      onKeyDown={(event) => {
        if (props.readOnly || (event.key !== "Enter" && event.key !== "F2"))
          return;
        event.preventDefault();
        activateInput();
      }}
    >
      <div
        data-testid={
          row.recordId === null
            ? draftRelationshipItemsTestId(binding.fieldKey)
            : relationshipItemsTestId(row.recordId, binding.fieldKey)
        }
        style={relationshipItemsWrapStyle}
      >
        <TimelineCollectionSummary
          fieldKey={binding.fieldKey}
          handleSelectMention={props.handleSelectMention}
          label={label}
          openInspectorForRow={props.openInspectorForRow}
          presentation={presentation}
          recordId={row.recordId}
        />
      </div>
      <input
        aria-label={`${label} ${row.recordId ?? "draft row"}`}
        data-testid={
          row.recordId === null
            ? draftTimelineCollectionInputTestId(binding.fieldKey)
            : timelineCollectionInputTestId(row.recordId, binding.fieldKey)
        }
        key={`${row.key}:${binding.draftKey}:${row.rowVersion ?? "draft"}`}
        ref={(element) => {
          props.focusTargetRef?.(element);
          props.registerInput(row.key, binding.draftKey, "grid", element);
        }}
        readOnly={props.readOnly}
        tabIndex={isInputActive ? 0 : -1}
        style={isInputActive ? collectionCellInputStyle : inactiveInputStyle}
        type="text"
        defaultValue={row.collectionDrafts[binding.draftKey]}
        onChange={(event) => {
          if (!props.readOnly)
            props.handleCollectionInputChange(
              collectionFocusKey,
              event.currentTarget.value,
            );
        }}
        onBlur={(event) => {
          if (props.readOnly) return;
          props.queueCollectionSave(
            row.key,
            binding.fieldKey,
            binding.draftKey,
            event.currentTarget.value,
          );
          if (event.currentTarget.value.trim() === "")
            props.deactivateCollectionInput(collectionFocusKey);
        }}
        onFocus={() => {
          props.activateCollectionInput(collectionFocusKey);
          props.updateTimelineSurfaceFocusAnchor(
            row.recordId,
            binding.fieldKey,
          );
          if (row.recordId !== null) props.handleSelectRow(row.recordId);
        }}
        onKeyDown={(event) => {
          if (!props.readOnly) {
            props.handleCollectionKeyDown(
              event,
              row.key,
              binding.fieldKey,
              binding.draftKey,
            );
          }
        }}
        placeholder={
          isInputActive ? `Add ${label.toLowerCase()} token` : undefined
        }
      />
    </fieldset>
  );
}

const gridCellInputStyle = {
  ...inputStyle,
  minHeight: 0,
  minBlockSize: 0,
  borderColor: "transparent",
  background: "transparent",
  padding: "var(--cartulary-grid-cell-padding)",
  color: "var(--ct-colors-ink)",
  fontSize: "var(--cartulary-grid-font-size)",
  lineHeight: "var(--cartulary-grid-line-height)",
};
const relationshipItemsWrapStyle = {
  display: "flex",
  alignItems: "center",
  flex: "0 1 auto",
  flexWrap: "nowrap" as const,
  gap: "0.2rem",
  marginBottom: 0,
  maxWidth: "100%",
  minWidth: 0,
  overflow: "hidden",
  whiteSpace: "nowrap" as const,
};
const tagChipStyle = {
  ...workbookRelationshipChipBaseStyle,
  flex: "0 1 auto",
  minWidth: 0,
  border: "var(--ct-component-chip-border)",
  background: "var(--ct-component-chip-backgroundColor)",
  color: "var(--ct-component-chip-textColor)",
  textOverflow: "ellipsis",
};
const emptyRelationshipStyle = {
  display: "inline-block",
  minWidth: 0,
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
  color: "var(--ct-colors-ink-tertiary)",
  fontSize: "var(--cartulary-grid-font-size)",
  lineHeight: "var(--cartulary-grid-line-height)",
};
const collectionCellStyle = {
  position: "relative" as const,
  display: "flex",
  alignItems: "center",
  gap: "0.25rem",
  margin: 0,
  minWidth: 0,
  maxWidth: "100%",
  minBlockSize: 0,
  blockSize: "100%",
  padding: 0,
  border: 0,
  fontSize: "var(--cartulary-grid-font-size)",
  lineHeight: "var(--cartulary-grid-line-height)",
  overflow: "hidden",
  whiteSpace: "nowrap" as const,
};
const collectionCellInputStyle = {
  ...gridCellInputStyle,
  flex: "1 1 4.5rem",
  minWidth: "4.25rem",
  inlineSize: "auto",
  blockSize: "100%",
  paddingBlock: 0,
  paddingInline: "0.1rem",
};
const inactiveInputStyle = {
  ...gridCellInputStyle,
  position: "absolute" as const,
  insetBlockStart: 0,
  insetInlineStart: 0,
  inlineSize: 1,
  blockSize: 1,
  minHeight: 0,
  padding: 0,
  opacity: 0,
  pointerEvents: "none" as const,
};
const collectionOverflowStyle = {
  flex: "0 0 auto",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-pill)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink-muted)",
  fontFamily: "var(--ct-typography-mono-fontFamily)",
  fontSize: "0.68rem",
  fontWeight: 700,
  lineHeight: 1.1,
  padding: "0.12rem 0.35rem",
};
const collectionOverflowButtonStyle = {
  ...collectionOverflowStyle,
  appearance: "none" as const,
  cursor: "pointer",
};
