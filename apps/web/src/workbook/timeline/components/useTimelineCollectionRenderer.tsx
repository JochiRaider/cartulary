import {
  dataTestIdSelector,
  draftRelationshipItemsTestId,
  draftTimelineCollectionInputTestId,
  relationshipItemsTestId,
  relationshipOverflowButtonTestId,
  timelineCollectionInputTestId,
} from "@cartulary/ui-contracts";
import type { Dispatch, SetStateAction } from "react";
import { useCallback } from "react";
import {
  WorkbookRelationshipChip,
  workbookRelationshipChipBaseStyle,
} from "../../components/WorkbookRelationshipChip";
import { visuallyHiddenStyle } from "../../utils/workbookStyles";
import {
  type CollectionItem,
  relationshipItemLabel,
  timelineRelationshipChipPresentation,
} from "../models/workbookMentionChips";
import {
  inputFocusKey,
  type TagCollectionItem,
  type TimelineCollectionBinding,
  type WorkbookRow,
} from "../models/workbookTimelineModel";
import type {
  RegisterTimelineInput,
  TimelineCollectionKeyDown,
  TimelineCollectionSave,
  TimelineEntityIndex,
} from "./TimelineWorkbookRendererTypes";
import { inputStyle } from "./TimelineWorkbookStyles";

export function useTimelineCollectionRenderer({
  activeCollectionInputKey,
  entityIndex,
  handleCollectionInputChange,
  handleCollectionKeyDown,
  handleSelectMention,
  handleSelectRow,
  openInspectorForRow,
  queueCollectionSave,
  readOnly,
  registerInput,
  setActiveCollectionInputKey,
  timelineBindingLabel,
  updateTimelineSurfaceFocusAnchor,
}: {
  readonly activeCollectionInputKey: string | null;
  readonly entityIndex: TimelineEntityIndex;
  readonly handleCollectionInputChange: (
    focusKey: string,
    value: string,
  ) => void;
  readonly handleCollectionKeyDown: TimelineCollectionKeyDown;
  readonly handleSelectMention: (recordId: string, itemRef: string) => void;
  readonly handleSelectRow: (recordId: string) => void;
  readonly openInspectorForRow: (recordId: string) => void;
  readonly queueCollectionSave: TimelineCollectionSave;
  readonly readOnly: boolean;
  readonly registerInput: RegisterTimelineInput;
  readonly setActiveCollectionInputKey: Dispatch<SetStateAction<string | null>>;
  readonly timelineBindingLabel: (fieldKey: string) => string;
  readonly updateTimelineSurfaceFocusAnchor: (
    recordId: string | null,
    fieldKey: string,
  ) => void;
}) {
  return useCallback(
    (row: WorkbookRow, binding: TimelineCollectionBinding) => {
      const label = timelineBindingLabel(binding.fieldKey);
      const items = row.collectionValues[binding.draftKey];
      const collectionInputTestId =
        row.recordId === null
          ? draftTimelineCollectionInputTestId(binding.fieldKey)
          : timelineCollectionInputTestId(row.recordId, binding.fieldKey);
      const collectionFocusKey = inputFocusKey(
        row.key,
        binding.draftKey,
        "grid",
      );
      const isCollectionInputActive =
        activeCollectionInputKey === collectionFocusKey ||
        row.collectionDrafts[binding.draftKey] !== "";
      const visibleItems = items.slice(0, 1);
      const hiddenItems = items.slice(1);
      const hiddenItemCount = Math.max(0, items.length - visibleItems.length);
      const activateCollectionInput = () => {
        if (readOnly) return;
        setActiveCollectionInputKey(collectionFocusKey);
        window.requestAnimationFrame(() => {
          document
            .querySelector<HTMLInputElement>(
              dataTestIdSelector(collectionInputTestId),
            )
            ?.focus({ preventScroll: true });
        });
      };
      const relationshipOverflowRecordId =
        binding.collectionKind === "relationship" ? row.recordId : null;
      return (
        <fieldset
          aria-label={`${label} collection cell`}
          style={collectionCellStyle}
          onClick={(event) => {
            if (readOnly) return;
            const target = event.target;
            if (
              target instanceof HTMLElement &&
              target.closest("[data-relationship-chip='true']") !== null
            ) {
              return;
            }
            activateCollectionInput();
          }}
          onKeyDown={(event) => {
            if (readOnly || (event.key !== "Enter" && event.key !== "F2")) {
              return;
            }
            event.preventDefault();
            activateCollectionInput();
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
            {items.length > 0 ? (
              binding.collectionKind === "relationship" ? (
                visibleItems.map((item) => (
                  <WorkbookRelationshipChip
                    key={item.itemRef}
                    presentation={{
                      ...timelineRelationshipChipPresentation({
                        entityIndex,
                        item: item as CollectionItem,
                      }),
                      onSelect: () => {
                        if (row.recordId) {
                          handleSelectMention(row.recordId, item.itemRef);
                        }
                      },
                    }}
                  />
                ))
              ) : (
                visibleItems.map((item) => (
                  <span
                    key={item.itemRef}
                    style={tagChipStyle}
                    title={(item as TagCollectionItem).displayText}
                  >
                    {(item as TagCollectionItem).displayText}
                  </span>
                ))
              )
            ) : (
              <span style={emptyRelationshipStyle}>No items</span>
            )}
            {hiddenItemCount > 0 ? (
              <>
                {relationshipOverflowRecordId !== null ? (
                  <button
                    aria-label={`Inspect ${hiddenItemCount} more ${label.toLowerCase()}`}
                    data-testid={relationshipOverflowButtonTestId(
                      relationshipOverflowRecordId,
                      binding.fieldKey,
                    )}
                    style={collectionOverflowButtonStyle}
                    title={`Inspect ${hiddenItemCount} more ${label.toLowerCase()}`}
                    type="button"
                    onClick={(event) => {
                      event.stopPropagation();
                      const firstHiddenItem = hiddenItems[0];
                      if (firstHiddenItem !== undefined) {
                        handleSelectMention(
                          relationshipOverflowRecordId,
                          firstHiddenItem.itemRef,
                        );
                      } else {
                        openInspectorForRow(relationshipOverflowRecordId);
                      }
                    }}
                    onKeyDown={(event) => {
                      event.stopPropagation();
                    }}
                  >
                    +{hiddenItemCount}
                  </button>
                ) : (
                  <span
                    aria-label={`${hiddenItemCount} more ${label.toLowerCase()}`}
                    role="note"
                    style={collectionOverflowStyle}
                    title={`${hiddenItemCount} more ${label.toLowerCase()}`}
                  >
                    +{hiddenItemCount}
                  </span>
                )}
                <span style={visuallyHiddenStyle}>
                  {hiddenItems
                    .map((item) =>
                      binding.collectionKind === "relationship"
                        ? relationshipItemLabel(
                            item as CollectionItem,
                            entityIndex,
                          )
                        : (item as TagCollectionItem).displayText,
                    )
                    .join(" ")}
                </span>
              </>
            ) : null}
          </div>
          <input
            aria-label={`${label} ${row.recordId ?? "draft row"}`}
            data-testid={collectionInputTestId}
            key={`${row.key}:${binding.draftKey}:${row.rowVersion ?? "draft"}`}
            ref={(element) => {
              registerInput(
                row.key,
                binding.draftKey,
                "grid",
                collectionInputTestId,
                element,
              );
            }}
            readOnly={readOnly}
            tabIndex={isCollectionInputActive ? 0 : -1}
            style={
              isCollectionInputActive
                ? collectionCellInputStyle
                : collectionCellInactiveInputStyle
            }
            type="text"
            defaultValue={row.collectionDrafts[binding.draftKey]}
            onChange={(event) => {
              if (!readOnly) {
                handleCollectionInputChange(
                  inputFocusKey(row.key, binding.draftKey, "grid"),
                  event.currentTarget.value,
                );
              }
            }}
            onBlur={(event) => {
              if (readOnly) return;
              queueCollectionSave(
                row.key,
                binding.fieldKey,
                binding.draftKey,
                event.currentTarget.value,
              );
              if (event.currentTarget.value.trim() === "") {
                setActiveCollectionInputKey((current) =>
                  current === collectionFocusKey ? null : current,
                );
              }
            }}
            onFocus={() => {
              setActiveCollectionInputKey(collectionFocusKey);
              updateTimelineSurfaceFocusAnchor(row.recordId, binding.fieldKey);
              if (row.recordId) handleSelectRow(row.recordId);
            }}
            onKeyDown={(event) => {
              if (!readOnly) {
                handleCollectionKeyDown(
                  event,
                  row.key,
                  binding.fieldKey,
                  binding.draftKey,
                );
              }
            }}
            placeholder={
              isCollectionInputActive
                ? `Add ${label.toLowerCase()} token`
                : undefined
            }
          />
        </fieldset>
      );
    },
    [
      activeCollectionInputKey,
      entityIndex,
      handleCollectionInputChange,
      handleCollectionKeyDown,
      handleSelectMention,
      handleSelectRow,
      openInspectorForRow,
      queueCollectionSave,
      readOnly,
      registerInput,
      setActiveCollectionInputKey,
      timelineBindingLabel,
      updateTimelineSurfaceFocusAnchor,
    ],
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
  fontSize: "0.78rem",
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

const collectionCellInactiveInputStyle = {
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
