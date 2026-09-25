import { bindGridEditorControlReveal } from "@cartulary/grid-adapter";
import {
  draftRelationshipItemsTestId,
  draftTimelineCollectionInputTestId,
  relationshipItemsTestId,
  relationshipOverflowButtonTestId,
  timelineCollectionInputTestId,
} from "@cartulary/ui-contracts";
import {
  type KeyboardEvent,
  useCallback,
  useLayoutEffect,
  useRef,
  useSyncExternalStore,
} from "react";
import {
  WorkbookRelationshipChip,
  workbookRelationshipChipBaseStyle,
} from "../../components/WorkbookRelationshipChip";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import { projectTimelineCollectionPresentation } from "../models/timelineCollectionPresentation";
import {
  type CollectionFieldKey,
  inputFocusKey,
  type TimelineCollectionBinding,
  type TimelineScalarEditorSurface,
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
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
  readonly binding: TimelineCollectionBinding;
  readonly entityIndex: TimelineEntityIndex;
  readonly focusTargetRef?:
    | ((element: HTMLInputElement | null) => void)
    | undefined;
  readonly handleCollectionKeyDown: TimelineCollectionKeyDown;
  readonly handleSelectRow: (recordId: string) => void;
  readonly handleInspectCollection: (
    recordId: string,
    fieldKey: CollectionFieldKey,
    itemRef: string,
  ) => void;
  readonly label: string;
  readonly isInspectionControlTarget: (target: EventTarget | null) => boolean;
  readonly queueCollectionSave: TimelineCollectionSave;
  readonly readOnly: boolean;
  readonly registerInput: RegisterTimelineInput;
  readonly registerTrigger: (
    recordId: string,
    fieldKey: string,
    itemRef: string | null,
    element: HTMLElement | null,
  ) => void;
  readonly rememberReturnFocus: (
    recordId: string,
    fieldKey: string,
    itemRef: string | null,
  ) => void;
  readonly registerCollectionItem: (
    recordId: string,
    fieldKey: string,
    itemRef: string,
    element: HTMLElement | null,
  ) => void;
  readonly row: WorkbookRow;
  readonly surface: TimelineScalarEditorSurface;
  readonly updateTimelineSurfaceFocusAnchor: (
    recordId: string | null,
    fieldKey: string,
  ) => void;
};

export function TimelineCollectionCell(props: TimelineCollectionCellProps) {
  const { binding, label, row, surface } = props;
  const cellRef = useRef<HTMLFieldSetElement>(null);
  const inputRef = useRef<HTMLInputElement | null>(null);
  const controls = useRef(new Map<string, HTMLButtonElement>());
  const suppressInspectionBlur = useRef(false);
  const composing = useRef(false);
  const focusOnActivation = useRef(false);
  const registry = props.editorDraftRegistry;
  const focusKey = inputFocusKey(row.key, binding.draftKey, surface);
  useSyncExternalStore(
    useCallback(
      (listener) => registry.subscribeInput(focusKey, listener),
      [registry, focusKey],
    ),
    useCallback(
      () => registry.collectionInputSnapshot(focusKey),
      [registry, focusKey],
    ),
  );
  const retainedDraft = registry.draftValue({
    rowKey: row.key,
    field: binding.draftKey,
    surface,
  });
  const retainDraft = (value: string, newRevision = false) =>
    registry.setDraft(
      { rowKey: row.key, field: binding.draftKey, surface },
      value,
      undefined,
      newRevision,
    );
  const presentation = projectTimelineCollectionPresentation({
    binding,
    entityIndex: props.entityIndex,
    row,
  });
  const isInspector = surface === "inspector";
  const draft =
    retainedDraft ??
    (surface === "grid" ? row.collectionDrafts[binding.draftKey] : "");
  const isInputActive =
    isInspector ||
    row.recordId === null ||
    (!props.readOnly && registry.isCollectionInputActive(focusKey));
  useLayoutEffect(() => {
    if (isInspector || !isInputActive || !inputRef.current) return;
    const reveal = bindGridEditorControlReveal(inputRef.current);
    return () => reveal?.dispose();
  }, [isInspector, isInputActive]);
  const { focusTargetRef, registerInput } = props;
  const register = useCallback(
    (element: HTMLInputElement | null) => {
      inputRef.current = element;
      focusTargetRef?.(element);
      registerInput(row.key, binding.draftKey, surface, element);
    },
    [focusTargetRef, registerInput, row.key, binding.draftKey, surface],
  );
  useLayoutEffect(
    () => () => registry.deactivateCollectionInput(focusKey),
    [registry, focusKey],
  );
  useLayoutEffect(() => {
    const input = inputRef.current;
    if (props.readOnly && !isInspector)
      registry.deactivateCollectionInput(focusKey);
    if (!isInputActive || !input || composing.current) return;
    if (input.value !== draft) input.value = draft;
    if (focusOnActivation.current && !props.readOnly) {
      focusOnActivation.current = false;
      input.focus({ preventScroll: true });
      input.setSelectionRange(input.value.length, input.value.length);
    }
    if (draft === "" && document.activeElement !== input)
      registry.deactivateCollectionInput(focusKey);
  }, [draft, isInputActive, props.readOnly, registry, focusKey, isInspector]);
  const activateInput = () => {
    if (props.readOnly) return;
    focusOnActivation.current = true;
    registry.activateCollectionInput(focusKey);
  };
  const inspect = (itemRef: string, triggerRef: string | null = itemRef) => {
    if (row.recordId === null) return;
    suppressInspectionBlur.current =
      document.activeElement === inputRef.current;
    if (!isInspector)
      props.rememberReturnFocus(row.recordId, binding.fieldKey, triggerRef);
    if (inputRef.current) retainDraft(inputRef.current.value);
    props.updateTimelineSurfaceFocusAnchor(row.recordId, binding.fieldKey);
    props.handleInspectCollection(row.recordId, binding.fieldKey, itemRef);
  };
  const registerControl = (
    key: string,
    element: HTMLButtonElement | null,
    itemRef: string | null = key,
  ) => {
    if (!isInspector && row.recordId !== null)
      props.registerTrigger(row.recordId, binding.fieldKey, itemRef, element);
    if (element === null) controls.current.delete(key);
    else controls.current.set(key, element);
  };
  const navigateControl = (
    event: KeyboardEvent<HTMLButtonElement>,
    key: string,
  ) => {
    if (event.key !== "ArrowLeft" && event.key !== "ArrowRight") return;
    const keys = [...controls.current.keys()];
    const index = keys.indexOf(key);
    const next = keys[index + (event.key === "ArrowLeft" ? -1 : 1)];
    if (next !== undefined) {
      event.preventDefault();
      controls.current.get(next)?.focus({ preventScroll: true });
    }
    event.stopPropagation();
  };
  const showSummary = !isInspector || presentation.kind === "tag";
  const tagItems =
    presentation.kind === "tag"
      ? isInspector
        ? presentation.items
        : presentation.visibleItems
      : [];
  return (
    <fieldset
      ref={cellRef}
      tabIndex={isInspector ? -1 : undefined}
      data-grid-sizing-draft={
        isInputActive || draft !== "" ? "true" : undefined
      }
      aria-label={`${label} collection ${isInspector ? "editor" : "cell"}`}
      style={isInspector ? inspectorCollectionStyle : collectionCellStyle}
    >
      {isInspector && binding.collectionKind === "tag" ? (
        <legend>{label}</legend>
      ) : null}
      {showSummary ? (
        <div
          data-testid={
            row.recordId === null
              ? draftRelationshipItemsTestId(binding.fieldKey)
              : relationshipItemsTestId(row.recordId, binding.fieldKey, surface)
          }
          style={
            isInspector
              ? inspectorCollectionItemsStyle
              : relationshipItemsWrapStyle
          }
          onPointerDownCapture={() => {
            if (document.activeElement === inputRef.current)
              suppressInspectionBlur.current = true;
          }}
        >
          {presentation.items.length === 0 && row.recordId !== null ? (
            <span style={emptyRelationshipStyle}>No items</span>
          ) : null}
          {presentation.kind === "relationship"
            ? presentation.visibleItems.map((item) => (
                <WorkbookRelationshipChip
                  key={item.itemRef}
                  elementRef={(element) =>
                    registerControl(item.itemRef, element)
                  }
                  onKeyDown={(event) => navigateControl(event, item.itemRef)}
                  presentation={{
                    ...item.chip,
                    ...(row.recordId === null
                      ? {}
                      : { onSelect: () => inspect(item.itemRef) }),
                  }}
                />
              ))
            : tagItems.map((item) =>
                isInspector ? (
                  <span
                    key={item.itemRef}
                    role="note"
                    tabIndex={-1}
                    aria-label={`Tag: ${item.displayText}`}
                    ref={(element) => {
                      if (row.recordId !== null)
                        props.registerCollectionItem(
                          row.recordId,
                          binding.fieldKey,
                          item.itemRef,
                          element,
                        );
                    }}
                    style={{
                      ...tagChipStyle,
                      whiteSpace: "pre-wrap",
                      overflowWrap: "anywhere",
                    }}
                  >
                    {item.displayText}
                  </span>
                ) : (
                  <button
                    key={item.itemRef}
                    type="button"
                    aria-label={`Inspect tag: ${item.displayText}`}
                    style={tagChipStyle}
                    ref={(element) => registerControl(item.itemRef, element)}
                    onClick={(event) => {
                      event.stopPropagation();
                      inspect(item.itemRef);
                    }}
                    onKeyDown={(event) => {
                      navigateControl(event, item.itemRef);
                      if (event.key !== "Tab" && event.key !== "Escape")
                        event.stopPropagation();
                    }}
                  >
                    {item.displayText}
                  </button>
                ),
              )}
          {!isInspector &&
          row.recordId !== null &&
          presentation.hiddenItemCount > 0 &&
          presentation.firstHiddenItemRef !== null ? (
            <button
              type="button"
              aria-label={`Inspect ${presentation.hiddenItemCount} more ${label.toLowerCase()}`}
              data-testid={relationshipOverflowButtonTestId(
                row.recordId,
                binding.fieldKey,
              )}
              style={collectionOverflowButtonStyle}
              ref={(element) => registerControl("overflow", element, null)}
              onClick={(event) => {
                event.stopPropagation();
                if (presentation.firstHiddenItemRef !== null)
                  inspect(presentation.firstHiddenItemRef, null);
              }}
              onKeyDown={(event) => {
                navigateControl(event, "overflow");
                if (event.key !== "Tab" && event.key !== "Escape")
                  event.stopPropagation();
              }}
            >
              +{presentation.hiddenItemCount}
            </button>
          ) : null}
        </div>
      ) : null}
      {!isInspector &&
      row.recordId !== null &&
      !props.readOnly &&
      !isInputActive ? (
        <button
          type="button"
          style={collectionOverflowButtonStyle}
          aria-label={`Add ${label.toLowerCase()} token`}
          onClick={(event) => {
            event.stopPropagation();
            activateInput();
          }}
          onKeyDown={(event) => {
            if (event.key === "F2") {
              event.preventDefault();
              activateInput();
            }
            if (event.key !== "Tab" && event.key !== "Escape")
              event.stopPropagation();
          }}
        >
          Add
        </button>
      ) : null}
      {!isInputActive && draft !== "" ? (
        <span
          style={emptyRelationshipStyle}
          role="note"
          aria-label={`${label} token draft retained`}
        >
          Draft
        </span>
      ) : null}
      {isInputActive ? (
        <input
          aria-label={`${label} ${row.recordId ?? "draft row"}`}
          data-testid={
            row.recordId === null
              ? draftTimelineCollectionInputTestId(binding.fieldKey)
              : timelineCollectionInputTestId(
                  row.recordId,
                  binding.fieldKey,
                  surface,
                )
          }
          key={focusKey}
          ref={register}
          readOnly={props.readOnly}
          style={isInspector ? inputStyle : collectionCellInputStyle}
          type="text"
          defaultValue={draft}
          onCompositionStart={(event) => {
            composing.current = true;
            if (!props.readOnly) retainDraft(event.currentTarget.value, true);
          }}
          onCompositionEnd={(event) => {
            composing.current = false;
            if (!props.readOnly) retainDraft(event.currentTarget.value);
          }}
          onInput={(event) => {
            if (!props.readOnly) {
              retainDraft(event.currentTarget.value, true);
            }
          }}
          onBlur={(event) => {
            if (props.readOnly || composing.current) return;
            const inspecting =
              suppressInspectionBlur.current ||
              props.isInspectionControlTarget(event.relatedTarget) ||
              (event.relatedTarget instanceof Element &&
                event.relatedTarget.closest(
                  '[data-grid-editor-external-action="true"]',
                ) !== null) ||
              (event.relatedTarget instanceof Node &&
                cellRef.current?.contains(event.relatedTarget));
            suppressInspectionBlur.current = false;
            if (inspecting) {
              retainDraft(event.currentTarget.value);
              return;
            }
            props.queueCollectionSave(
              row.key,
              binding.fieldKey,
              binding.draftKey,
              event.currentTarget.value,
              surface,
            );
            if (event.currentTarget.value.trim() === "")
              registry.deactivateCollectionInput(focusKey);
          }}
          onFocus={() => {
            if (!isInspector) registry.activateCollectionInput(focusKey);
            props.updateTimelineSurfaceFocusAnchor(
              row.recordId,
              binding.fieldKey,
            );
            if (row.recordId !== null) props.handleSelectRow(row.recordId);
          }}
          onKeyDown={(event) => {
            if (event.nativeEvent.isComposing || composing.current) {
              event.stopPropagation();
              return;
            }
            if (event.defaultPrevented) return;
            if (event.key === "Escape") {
              event.preventDefault();
              event.stopPropagation();
              if (!props.readOnly) {
                event.currentTarget.value = "";
                retainDraft("", true);
                registry.deactivateCollectionInput(focusKey);
              }
              if (isInspector) cellRef.current?.focus({ preventScroll: true });
              else if (row.recordId !== null)
                props.handleCollectionKeyDown(
                  event,
                  row.key,
                  binding.fieldKey,
                  binding.draftKey,
                  surface,
                );
              return;
            }
            if (!props.readOnly)
              props.handleCollectionKeyDown(
                event,
                row.key,
                binding.fieldKey,
                binding.draftKey,
                surface,
              );
            if (event.key !== "Tab" && event.key !== "Escape")
              event.stopPropagation();
          }}
          placeholder={
            isInputActive ? `Add ${label.toLowerCase()} token` : undefined
          }
        />
      ) : null}
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
  flex: "1 1 auto",
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
const collectionOverflowStyle = {
  flex: "0 0 auto",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-pill)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink-muted)",
  fontFamily: "var(--ct-typography-mono-fontFamily)",
  fontSize: "inherit",
  fontWeight: 700,
  lineHeight: "inherit",
  padding: "0 var(--ct-spacing-xs)",
};
const collectionOverflowButtonStyle = {
  ...collectionOverflowStyle,
  appearance: "none" as const,
  cursor: "pointer",
};

const inspectorCollectionStyle = {
  display: "grid",
  gap: "0.4rem",
  minWidth: 0,
  margin: 0,
  padding: 0,
  border: 0,
};
const inspectorCollectionItemsStyle = {
  display: "grid",
  gap: "0.4rem",
  minWidth: 0,
};
