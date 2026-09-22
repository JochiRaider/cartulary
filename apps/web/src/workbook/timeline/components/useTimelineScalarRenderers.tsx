import type { GridEditorFocusTarget } from "@cartulary/grid-adapter";
import {
  conflictMarkerTestId,
  rowCellTestId,
  timelineScalarEditorTestId,
} from "@cartulary/ui-contracts";
import { useCallback, useSyncExternalStore } from "react";
import type { PresenceScope } from "../../collaboration/workbookPresencePresentation";
import {
  WorkbookCellPresenceMarker,
  WorkbookPresenceCellLayout,
} from "../../components/WorkbookPresenceMarkers";
import { stringifyGridValue } from "../../utils/workbookValueFormat";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import {
  inputFocusKey,
  type RowValues,
  type TimelineScalarBinding,
  type TimelineScalarEditorSurface,
} from "../models/timelineFieldRegistry";
import {
  readTimelineCellValue,
  type WorkbookRow,
} from "../models/timelineRowModel";
import { TimelineScalarEditor } from "./TimelineScalarEditor";
import type {
  RegisterTimelineInput,
  TimelineScalarBlurCommit,
  TimelineScalarKeyCommit,
} from "./TimelineWorkbookRendererTypes";
import {
  secondaryActionButtonStyle,
  timelineGridBodyStyle,
} from "./TimelineWorkbookStyles";

export function useTimelineScalarRenderers({
  conflictQueue,
  editingPresenceForCell,
  editorDraftRegistry,
  handleBlur,
  handleEditModePresence,
  handleKeyDown,
  handleSelectRow,
  readOnly,
  readCurrentRow,
  registerInput,
  setActiveConflictKey,
  timelineBindingLabel,
  updateTimelineSurfaceFocusAnchor,
}: {
  readonly conflictQueue: Record<string, { readonly key: string }>;
  readonly editingPresenceForCell: (
    recordId: string | null,
    fieldKey: string,
  ) => PresenceScope;
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
  readonly handleBlur: TimelineScalarBlurCommit;
  readonly handleEditModePresence: (
    recordId: string | null,
    fieldKey: string,
    editing: boolean,
  ) => void;
  readonly handleKeyDown: TimelineScalarKeyCommit;
  readonly handleSelectRow: (recordId: string) => void;
  readonly readOnly: boolean;
  readonly readCurrentRow?: ((row: WorkbookRow) => WorkbookRow) | undefined;
  readonly registerInput: RegisterTimelineInput;
  readonly setActiveConflictKey: (key: string | null) => void;
  readonly timelineBindingLabel: (fieldKey: string) => string;
  readonly updateTimelineSurfaceFocusAnchor: (
    recordId: string | null,
    fieldKey: string,
  ) => void;
}) {
  const timelineScalarControlId = useCallback(
    (
      row: WorkbookRow,
      binding: TimelineScalarBinding,
      surface: TimelineScalarEditorSurface,
    ) =>
      ["timeline-editor", surface, row.key, binding.fieldKey]
        .map((value) => value.replace(/[^a-zA-Z0-9_-]+/g, "-"))
        .join("-"),
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
      onControlledDraftChange?:
        | ((
            value: string,
            options?: { readonly advanceRevision?: boolean },
          ) => void)
        | undefined,
    ) => {
      const label = timelineBindingLabel(binding.fieldKey);
      const dataTestId = timelineScalarEditorTestId({
        fieldKey: binding.fieldKey,
        recordId: row.recordId,
        surface,
      });
      const conflictKey =
        row.recordId === null ? null : `${row.recordId}:${binding.fieldKey}`;
      const localConflict =
        conflictKey === null ? undefined : conflictQueue[conflictKey];
      const content = (
        <>
          <TimelineScalarEditor
            key={inputFocusKey(row.key, binding.key, surface)}
            accessibleLabel={
              surface === "grid"
                ? `${label} ${row.recordId ?? "draft row"}`
                : undefined
            }
            blockedByConflict={localConflict !== undefined}
            committedValue={
              (readCurrentRow?.(row) ?? row).committedValues[binding.key]
            }
            controlId={controlId}
            dataTestId={dataTestId}
            editorDraftRegistry={editorDraftRegistry}
            field={binding.key}
            focusTargetRef={focusTargetRef}
            multiline={binding.multiline}
            onBlurCommit={handleBlur}
            onCloseGridEditor={closeGridEditor}
            onDraftChange={(rowKey, field, editorSurface, value, input) => {
              // Managed controls retain through the Adapter exactly once. Native
              // edits advance authoring even when a replacement has equal text.
              if (onControlledDraftChange) {
                onControlledDraftChange(value, {
                  advanceRevision: input !== undefined,
                });
              } else {
                editorDraftRegistry.setDraft(
                  { rowKey, field, surface: editorSurface },
                  value,
                  readCurrentRow?.(row) ?? row,
                  input !== undefined,
                );
              }
              if (input === undefined || input.composing || readOnly) return;
              const capturing = row.recordId === null && value.trim() !== "";
              if (capturing && editorDraftRegistry.beginCapture(row.key)) {
                performance.mark(
                  "cartulary.workbook.blank_row_commit_accepted",
                  {
                    detail: { field: binding.fieldKey, surface },
                  },
                );
              }
              if (
                capturing ||
                editorDraftRegistry.hasCapture(row.key) ||
                input.pasteCompleted
              ) {
                handleBlur(rowKey, field, editorSurface, value);
              }
            }}
            onEditModeChange={handleEditModePresence}
            onFocusAnchor={updateTimelineSurfaceFocusAnchor}
            onFocusRecord={handleSelectRow}
            onKeyCommit={handleKeyDown}
            presenceFieldKey={binding.fieldKey}
            readOnly={readOnly}
            registerInput={registerInput}
            rowKey={row.key}
            rowRecordId={row.recordId}
            surface={surface}
          />
          {!readOnly && row.recordId !== null ? (
            <TimelineDraftReview
              registry={editorDraftRegistry}
              row={readCurrentRow?.(row) ?? row}
              field={binding.key}
              surface={surface}
              label={label}
            />
          ) : null}
          {localConflict ? (
            <button
              data-grid-editor-external-action="true"
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
          ) : null}
        </>
      );
      return surface !== "grid" || row.recordId === null ? (
        content
      ) : (
        <WorkbookPresenceCellLayout
          editing
          marker={
            <WorkbookCellPresenceMarker
              fieldKey={binding.fieldKey}
              fieldLabel={label}
              recordId={row.recordId}
              presences={editingPresenceForCell(row.recordId, binding.fieldKey)}
            />
          }
        >
          {content}
        </WorkbookPresenceCellLayout>
      );
    },
    [
      conflictQueue,
      editingPresenceForCell,
      editorDraftRegistry,
      handleBlur,
      handleEditModePresence,
      handleKeyDown,
      handleSelectRow,
      readOnly,
      readCurrentRow,
      registerInput,
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
      onControlledDraftChange?:
        | ((
            value: string,
            options?: { readonly advanceRevision?: boolean },
          ) => void)
        | undefined,
    ) =>
      renderTimelineScalarControl(
        row,
        binding,
        "grid",
        timelineScalarControlId(row, binding, "grid"),
        closeGridEditor,
        focusTargetRef,
        onControlledDraftChange,
      ),
    [renderTimelineScalarControl, timelineScalarControlId],
  );

  const renderTimelineScalarCell = useCallback(
    (row: WorkbookRow, binding: TimelineScalarBinding) => {
      const conflictKey =
        row.recordId === null ? null : `${row.recordId}:${binding.fieldKey}`;
      const localConflict =
        conflictKey === null ? undefined : conflictQueue[conflictKey];
      const text = stringifyGridValue(
        readTimelineCellValue(row.rawRow, binding.fieldKey),
      );
      return (
        <WorkbookPresenceCellLayout
          marker={
            <WorkbookCellPresenceMarker
              fieldKey={binding.fieldKey}
              fieldLabel={timelineBindingLabel(binding.fieldKey)}
              presences={editingPresenceForCell(row.recordId, binding.fieldKey)}
              recordId={row.recordId}
            />
          }
        >
          <span
            data-testid={
              row.recordId === null
                ? undefined
                : rowCellTestId(row.recordId, binding.fieldKey)
            }
            style={timelineGridBodyStyle}
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
        </WorkbookPresenceCellLayout>
      );
    },
    [
      conflictQueue,
      editingPresenceForCell,
      setActiveConflictKey,
      timelineBindingLabel,
    ],
  );

  return {
    renderTimelineGridEditor,
    renderTimelineScalarCell,
  };
}

function TimelineDraftReview({
  registry,
  row,
  field,
  surface,
  label,
}: {
  registry: TimelineEditorDraftRegistry;
  row: WorkbookRow;
  field: keyof RowValues;
  surface: TimelineScalarEditorSurface;
  label: string;
}) {
  const identity = { rowKey: row.key, field, surface };
  const stale = useSyncExternalStore(
    useCallback(
      (listener: () => void) => registry.subscribeRow(row.key, listener),
      [registry, row.key],
    ),
    () => registry.needsReview(identity, row),
  );
  return stale ? (
    <span
      data-grid-editor-toolbar={surface === "grid" ? "true" : undefined}
      role="status"
    >
      Saved value changed.{" "}
      <button type="button" onClick={() => registry.review(identity, row)}>
        Keep draft {label}
      </button>
    </span>
  ) : null;
}

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
