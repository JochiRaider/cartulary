import type {
  GridColumn,
  GridEditCommitOutcome,
  GridEditorFocusTarget,
} from "@cartulary/grid-adapter";
import {
  conflictMarkerTestId,
  dataTestIdSelector,
  draftRelationshipItemsTestId,
  draftTimelineCollectionInputTestId,
  gridSortHeaderTestId,
  relationshipItemsTestId,
  relationshipOverflowButtonTestId,
  rowCellTestId,
  timelineCollectionInputTestId,
  timelineScalarEditorTestId,
} from "@cartulary/ui-contracts";
import {
  resolveHeaderSortFieldKey,
  type ViewContract,
} from "@cartulary/view-contracts";
import {
  type Dispatch,
  type ClipboardEvent as ReactClipboardEvent,
  type KeyboardEvent as ReactKeyboardEvent,
  type ReactNode,
  type RefObject,
  type SetStateAction,
  useCallback,
  useMemo,
} from "react";
import { buildEvidenceCountDisplayViewModel } from "../../models/evidenceLifecycleViewModel";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { PresenceRecord } from "../../utils/workbookPresence";
import { visuallyHiddenStyle } from "../../utils/workbookStyles";
import { stringifyGridValue } from "../../utils/workbookValueFormat";
import type { CollectionItem } from "../models/workbookMentionChips";
import {
  buildExpandedTimelineColumnWidths,
  type CollectionDraftKey,
  type CollectionFieldKey,
  type FocusFieldKey,
  inputFocusKey,
  type RowValues,
  readTimelineCellValue,
  type TagCollectionItem,
  type TimelineCollectionBinding,
  type TimelineScalarBinding,
  type TimelineScalarEditorSurface,
  timelineColumnWidth,
  timelineVisibleBindings,
  type WorkbookRow,
} from "../models/workbookTimelineModel";
import {
  RelationshipChip,
  relationshipChipBaseStyle,
  relationshipItemLabel,
  TimelineScalarEditor,
} from "./TimelineCellEditors";
import { TimelineCellPresenceMarker } from "./TimelinePresenceMarkers";

type EntityIndex = Record<string, { label: string }>;

type ScalarBlurCommit = (
  rowKey: string,
  field: keyof RowValues,
  surface: TimelineScalarEditorSurface,
  value: string,
) => void;

type ScalarDraftChange = (
  rowKey: string,
  field: keyof RowValues,
  surface: TimelineScalarEditorSurface,
  value: string,
) => void;

type ScalarKeyCommit = (
  event: ReactKeyboardEvent<HTMLInputElement | HTMLTextAreaElement>,
  rowKey: string,
  field: keyof RowValues,
  surface: TimelineScalarEditorSurface,
) => void;

type ScalarPasteCommit = (
  event: ReactClipboardEvent<HTMLInputElement | HTMLTextAreaElement>,
  rowKey: string,
  field: keyof RowValues,
  surface: TimelineScalarEditorSurface,
) => void;

type ScalarGridCommit = (
  rowKey: string,
  field: keyof RowValues,
  value: string,
) => Promise<GridEditCommitOutcome>;

type RegisterTimelineInput = (
  rowKey: string,
  field: FocusFieldKey,
  surface: TimelineScalarEditorSurface,
  dataTestId: string,
  element: HTMLInputElement | HTMLTextAreaElement | null,
) => void;

type CollectionSave = (
  rowKey: string,
  fieldKey: CollectionFieldKey,
  focusField: CollectionDraftKey,
  draftValueOverride?: string,
  source?: "keyboard" | "blur",
) => void;

type CollectionKeyDown = (
  event: ReactKeyboardEvent<HTMLInputElement>,
  rowKey: string,
  fieldKey: CollectionFieldKey,
  draftKey: CollectionDraftKey,
) => void;

export type TimelineWorkbookRenderers = {
  readonly renderTimelineCollectionInput: (
    row: WorkbookRow,
    binding: TimelineCollectionBinding,
  ) => ReactNode;
  readonly renderTimelineGridEditor: (
    row: WorkbookRow,
    binding: TimelineScalarBinding,
    closeGridEditor?:
      | ((commit: boolean, draftValue: string) => void)
      | undefined,
    focusTargetRef?:
      | ((element: GridEditorFocusTarget | null) => void)
      | undefined,
    controlledDraftValue?: string | undefined,
    onControlledDraftChange?: ((value: string) => void) | undefined,
  ) => ReactNode;
  readonly renderTimelineInspectorEditor: (
    row: WorkbookRow,
    binding: TimelineScalarBinding,
  ) => ReactNode;
  readonly timelineBindingLabel: (fieldKey: string) => string;
  readonly timelineColumns: readonly GridColumn<WorkbookRow>[];
};

export function useTimelineWorkbookRenderers({
  activeCollectionInputKey,
  conflictQueue,
  commitScalarGridEdit,
  editingPresenceForCell,
  entityIndex,
  gridShellWidth,
  handleBlur,
  handleCollectionInputChange,
  handleCollectionKeyDown,
  handleEditModePresence,
  handleKeyDown,
  handlePaste,
  handleSelectMention,
  handleSelectRow,
  openInspectorForRow,
  queueCollectionSave,
  registerInput,
  readOnly,
  rowGutterWidth,
  scalarDraftValuesRef,
  setScalarEditorDraftValue,
  setActiveCollectionInputKey,
  setActiveConflictKey,
  timelineContract,
  updateTimelineSurfaceFocusAnchor,
}: {
  readonly activeCollectionInputKey: string | null;
  readonly conflictQueue: Record<string, { readonly key: string }>;
  readonly commitScalarGridEdit: ScalarGridCommit;
  readonly editingPresenceForCell: (
    recordId: string | null,
    fieldKey: string,
  ) => readonly PresenceRecord[];
  readonly entityIndex: EntityIndex;
  readonly gridShellWidth: number;
  readonly handleBlur: ScalarBlurCommit;
  readonly handleCollectionInputChange: (
    focusKey: string,
    value: string,
  ) => void;
  readonly handleCollectionKeyDown: CollectionKeyDown;
  readonly handleEditModePresence: (
    recordId: string | null,
    fieldKey: string,
    editing: boolean,
  ) => void;
  readonly handleKeyDown: ScalarKeyCommit;
  readonly handlePaste: ScalarPasteCommit;
  readonly handleSelectMention: (recordId: string, itemRef: string) => void;
  readonly handleSelectRow: (recordId: string) => void;
  readonly openInspectorForRow: (recordId: string) => void;
  readonly queueCollectionSave: CollectionSave;
  readonly registerInput: RegisterTimelineInput;
  readonly readOnly: boolean;
  readonly rowGutterWidth: number;
  readonly scalarDraftValuesRef: RefObject<Map<string, string>>;
  readonly setScalarEditorDraftValue: ScalarDraftChange;
  readonly setActiveCollectionInputKey: Dispatch<SetStateAction<string | null>>;
  readonly setActiveConflictKey: Dispatch<SetStateAction<string | null>>;
  readonly timelineContract: ViewContract;
  readonly updateTimelineSurfaceFocusAnchor: (
    recordId: string | null,
    fieldKey: string,
  ) => void;
}): TimelineWorkbookRenderers {
  const timelineBindingLabel = useCallback(
    (fieldKey: string) =>
      timelineContract.fieldMap[fieldKey]?.label ?? fieldKey,
    [timelineContract],
  );

  const timelineScalarControlId = useCallback(
    (
      row: WorkbookRow,
      binding: TimelineScalarBinding,
      surface: TimelineScalarEditorSurface,
    ) => {
      return ["timeline-editor", surface, row.key, binding.fieldKey]
        .map((value) => value.replace(/[^a-zA-Z0-9_-]+/g, "-"))
        .join("-");
    },
    [],
  );

  const renderTimelineScalarControl = useCallback(
    (
      row: WorkbookRow,
      binding: TimelineScalarBinding,
      surface: TimelineScalarEditorSurface,
      controlId: string,
      closeGridEditor?:
        | ((commit: boolean, draftValue: string) => void)
        | undefined,
      focusTargetRef?:
        | ((element: GridEditorFocusTarget | null) => void)
        | undefined,
      controlledDraftValue?: string | undefined,
      onControlledDraftChange?: ((value: string) => void) | undefined,
    ) => {
      const label = timelineBindingLabel(binding.fieldKey);
      const gridAccessibleLabel =
        surface === "grid"
          ? `${label} ${row.recordId ?? "draft row"}`
          : undefined;
      const dataTestId = timelineScalarEditorTestId({
        fieldKey: binding.fieldKey,
        recordId: row.recordId,
        surface,
      });
      const conflictKey =
        row.recordId === null ? null : `${row.recordId}:${binding.fieldKey}`;
      const localConflict =
        conflictKey === null ? undefined : conflictQueue[conflictKey];
      return (
        <>
          <TimelineScalarEditor
            key={inputFocusKey(row.key, binding.key, surface)}
            accessibleLabel={gridAccessibleLabel}
            blockedByConflict={localConflict !== undefined}
            committedValue={row.values[binding.key]}
            controlId={controlId}
            dataTestId={dataTestId}
            draftValue={
              controlledDraftValue ??
              scalarDraftValuesRef.current?.get(
                inputFocusKey(row.key, binding.key, surface),
              )
            }
            field={binding.key}
            focusTargetRef={focusTargetRef}
            multiline={binding.multiline}
            onEditModeChange={handleEditModePresence}
            onCloseGridEditor={closeGridEditor}
            onFocusAnchor={updateTimelineSurfaceFocusAnchor}
            registerInput={registerInput}
            readOnly={readOnly}
            presenceFieldKey={binding.fieldKey}
            rowKey={row.key}
            rowRecordId={row.recordId}
            surface={surface}
            onBlurCommit={handleBlur}
            onDraftChange={(rowKey, field, editorSurface, value) => {
              setScalarEditorDraftValue(rowKey, field, editorSurface, value);
              onControlledDraftChange?.(value);
            }}
            onFocusRecord={handleSelectRow}
            onKeyCommit={handleKeyDown}
            onPasteCommit={handlePaste}
          />
          {localConflict && surface === "inspector" ? (
            <button
              type="button"
              data-testid={conflictMarkerTestId(
                row.recordId ?? "draft",
                binding.fieldKey,
              )}
              style={conflictMarkerStyle}
              onClick={() => setActiveConflictKey(localConflict.key)}
            >
              Conflict
            </button>
          ) : null}
        </>
      );
    },
    [
      conflictQueue,
      handleBlur,
      handleEditModePresence,
      handleKeyDown,
      handlePaste,
      handleSelectRow,
      registerInput,
      readOnly,
      scalarDraftValuesRef,
      setScalarEditorDraftValue,
      setActiveConflictKey,
      timelineBindingLabel,
      updateTimelineSurfaceFocusAnchor,
    ],
  );

  const renderTimelineGridEditor = useCallback(
    (
      row: WorkbookRow,
      binding: TimelineScalarBinding,
      closeGridEditor?:
        | ((commit: boolean, draftValue: string) => void)
        | undefined,
      focusTargetRef?:
        | ((element: GridEditorFocusTarget | null) => void)
        | undefined,
      controlledDraftValue?: string | undefined,
      onControlledDraftChange?: ((value: string) => void) | undefined,
    ) => {
      return renderTimelineScalarControl(
        row,
        binding,
        "grid",
        timelineScalarControlId(row, binding, "grid"),
        closeGridEditor,
        focusTargetRef,
        controlledDraftValue,
        onControlledDraftChange,
      );
    },
    [renderTimelineScalarControl, timelineScalarControlId],
  );

  const renderTimelineScalarCell = useCallback(
    (row: WorkbookRow, binding: TimelineScalarBinding) => {
      const conflictKey =
        row.recordId === null ? null : `${row.recordId}:${binding.fieldKey}`;
      const localConflict =
        conflictKey === null ? undefined : conflictQueue[conflictKey];
      const presences = editingPresenceForCell(row.recordId, binding.fieldKey);
      const text = stringifyGridValue(
        readTimelineCellValue(row.rawRow, binding.fieldKey),
      );
      return (
        <>
          <span
            data-testid={
              row.recordId === null
                ? undefined
                : rowCellTestId(row.recordId, binding.fieldKey)
            }
            style={bodyStyle}
          >
            {text === "" ? "—" : text}
          </span>
          {localConflict === undefined ? null : (
            <button
              data-grid-prevent-cell-edit="true"
              data-testid={conflictMarkerTestId(
                row.recordId ?? "draft",
                binding.fieldKey,
              )}
              style={conflictMarkerStyle}
              type="button"
              onClick={() => setActiveConflictKey(localConflict.key)}
            >
              Conflict
            </button>
          )}
          <TimelineCellPresenceMarker
            fieldKey={binding.fieldKey}
            fieldLabel={timelineBindingLabel(binding.fieldKey)}
            presences={presences}
            recordId={row.recordId}
          />
        </>
      );
    },
    [
      conflictQueue,
      editingPresenceForCell,
      setActiveConflictKey,
      timelineBindingLabel,
    ],
  );

  const renderTimelineInspectorEditor = useCallback(
    (row: WorkbookRow, binding: TimelineScalarBinding) => {
      const controlId = timelineScalarControlId(row, binding, "inspector");
      return (
        <div key={binding.fieldKey} style={labelStyle}>
          <label htmlFor={controlId}>
            {timelineBindingLabel(binding.fieldKey)}
          </label>
          {renderTimelineScalarControl(row, binding, "inspector", controlId)}
        </div>
      );
    },
    [
      renderTimelineScalarControl,
      timelineBindingLabel,
      timelineScalarControlId,
    ],
  );

  const renderTimelineCollectionInput = useCallback(
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
      const visibleItemLimit = 1;
      const visibleItems = items.slice(0, visibleItemLimit);
      const hiddenItems = items.slice(visibleItemLimit);
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
            if (readOnly) return;
            if (event.key !== "Enter" && event.key !== "F2") {
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
                  <RelationshipChip
                    key={item.itemRef}
                    entityIndex={entityIndex}
                    item={item as CollectionItem}
                    onSelect={() => {
                      if (row.recordId) {
                        handleSelectMention(row.recordId, item.itemRef);
                      }
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
              if (readOnly) return;
              handleCollectionInputChange(
                inputFocusKey(row.key, binding.draftKey, "grid"),
                event.currentTarget.value,
              );
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
              if (row.recordId) {
                handleSelectRow(row.recordId);
              }
            }}
            onKeyDown={(event) => {
              if (readOnly) return;
              handleCollectionKeyDown(
                event,
                row.key,
                binding.fieldKey,
                binding.draftKey,
              );
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

  const timelineColumnWidths = useMemo(
    () =>
      buildExpandedTimelineColumnWidths({
        actionsColumnWidth: 0,
        fieldKeys: timelineVisibleFieldKeys,
        gridShellWidth,
        rowGutterWidth,
      }),
    [gridShellWidth, rowGutterWidth],
  );

  const timelineColumns = useMemo<readonly GridColumn<WorkbookRow>[]>(
    () =>
      timelineVisibleBindings.map((binding): GridColumn<WorkbookRow> => {
        const renderCell = (row: WorkbookRow) => {
          if (binding.kind === "scalar") {
            return renderTimelineScalarCell(row, binding);
          }
          if (binding.kind === "collection") {
            return renderTimelineCollectionInput(row, binding);
          }
          if (binding.fieldKey === "timeline.evidence_count") {
            const countDisplay = buildEvidenceCountDisplayViewModel({
              projectedCount: readTimelineCellValue(
                row.rawRow,
                binding.fieldKey,
              ),
              projectedHasEvidence: readTimelineCellValue(
                row.rawRow,
                "timeline.has_evidence",
              ),
            });
            return (
              <span
                data-evidence-count-state={countDisplay.stateKey}
                style={timelineEvidenceCellStyle}
              >
                <span
                  data-testid={
                    row.recordId === null
                      ? undefined
                      : rowCellTestId(row.recordId, binding.fieldKey)
                  }
                >
                  {countDisplay.displayCount}
                </span>
                {row.recordId === null ? null : (
                  <span
                    data-testid={rowCellTestId(
                      row.recordId,
                      "timeline.has_evidence",
                    )}
                    style={
                      countDisplay.hasEvidence
                        ? timelineEvidenceFlagOnStyle
                        : timelineEvidenceFlagOffStyle
                    }
                    title={
                      countDisplay.hasEvidence
                        ? "Timeline row has evidence"
                        : "Timeline row has no evidence"
                    }
                  >
                    {String(countDisplay.hasEvidence)}
                  </span>
                )}
              </span>
            );
          }
          const text = stringifyGridValue(
            readTimelineCellValue(row.rawRow, binding.fieldKey),
          );
          return (
            <span
              data-testid={
                row.recordId === null
                  ? undefined
                  : rowCellTestId(row.recordId, binding.fieldKey)
              }
              style={
                binding.fieldKey === "timeline.edited_at"
                  ? timelineTimestampCellStyle
                  : bodyStyle
              }
            >
              {text === "" ? "—" : text}
            </span>
          );
        };
        return {
          contractWritable:
            timelineContract.fieldMap[binding.fieldKey]?.gridEditable === true,
          fieldKey: binding.fieldKey,
          getClipboardValue: (row) =>
            stringifyGridValue(
              readTimelineCellValue(row.rawRow, binding.fieldKey),
            ),
          headerTestId: gridSortHeaderTestId(
            timelineViewSchemaId,
            binding.fieldKey,
          ),
          label: timelineBindingLabel(binding.fieldKey),
          width:
            timelineColumnWidths[binding.fieldKey] ??
            timelineColumnWidth(binding.fieldKey),
          renderCell: ({ row }) => renderCell(row),
          renderDraftCell: ({ row }) =>
            binding.kind === "scalar"
              ? renderTimelineGridEditor(row, binding)
              : renderCell(row),
          editor:
            binding.kind === "scalar"
              ? {
                  ...(timelineContract.fieldMap[binding.fieldKey]?.clearable
                    ? { clearDraftValue: "" }
                    : {}),
                  commit: (intent) =>
                    commitScalarGridEdit(
                      intent.row.key,
                      binding.key,
                      String(intent.draftValue ?? ""),
                    ),
                  initialDraftValue: (row) =>
                    readTimelineCellValue(row.rawRow, binding.fieldKey),
                  renderEditor: (context) =>
                    renderTimelineGridEditor(
                      context.row,
                      binding,
                      (commit, draftValue) => {
                        if (commit) void context.commit(draftValue);
                        else context.cancel();
                      },
                      context.focusTargetRef,
                      String(context.draftValue ?? ""),
                      (value) => context.setDraftValue(value),
                    ),
                }
              : undefined,
          sortableFieldKey: resolveHeaderSortFieldKey(
            timelineContract,
            binding.fieldKey,
          ),
        };
      }),
    [
      renderTimelineCollectionInput,
      renderTimelineGridEditor,
      renderTimelineScalarCell,
      commitScalarGridEdit,
      timelineBindingLabel,
      timelineColumnWidths,
      timelineContract,
    ],
  );

  return {
    renderTimelineCollectionInput,
    renderTimelineGridEditor,
    renderTimelineInspectorEditor,
    timelineBindingLabel,
    timelineColumns,
  };
}

const timelineVisibleFieldKeys = timelineVisibleBindings.map(
  (binding) => binding.fieldKey,
);

const bodyStyle = {
  margin: 0,
  lineHeight: 1.5,
  color: "var(--ct-colors-ink-muted)",
};

const timelineTimestampCellStyle = {
  display: "block",
  minWidth: 0,
  maxWidth: "100%",
  margin: 0,
  overflow: "hidden",
  overflowWrap: "normal" as const,
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
  wordBreak: "normal" as const,
  lineHeight: "var(--ct-typography-grid-cell-lineHeight)",
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

const actionButtonStyle = {
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  padding: "0.55rem 0.9rem",
  font: "inherit",
  cursor: "pointer",
};

const secondaryActionButtonStyle = {
  ...actionButtonStyle,
  background: "var(--ct-colors-surface-3)",
};

const conflictMarkerStyle = {
  ...secondaryActionButtonStyle,
  position: "absolute" as const,
  insetBlockStart: "4px",
  insetInlineEnd: "6px",
  boxSizing: "border-box" as const,
  minHeight: 0,
  height: "18px",
  margin: 0,
  borderColor: "var(--ct-colors-semantic-conflict)",
  color: "var(--ct-colors-semantic-conflict)",
  background: "var(--ct-colors-surface-2)",
  padding: "0 0.35rem",
  fontSize: "0.68rem",
  lineHeight: 1,
};

const timelineEvidenceCellStyle = {
  display: "inline-flex",
  alignItems: "center",
  gap: "0.45rem",
  minWidth: 0,
};

const timelineEvidenceFlagBaseStyle = {
  borderRadius: "999px",
  padding: "0.15rem 0.42rem",
  fontSize: "0.72rem",
  lineHeight: 1.2,
};

const timelineEvidenceFlagOnStyle = {
  ...timelineEvidenceFlagBaseStyle,
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-semantic-success)",
};

const timelineEvidenceFlagOffStyle = {
  ...timelineEvidenceFlagBaseStyle,
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink-muted)",
};

const labelStyle = {
  display: "grid",
  gap: "0.4rem",
  fontSize: "0.95rem",
  color: "var(--ct-colors-ink-muted)",
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
  ...relationshipChipBaseStyle,
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
