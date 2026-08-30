import type { GridEditorFocusTarget } from "@cartulary/grid-adapter";
import {
  conflictMarkerTestId,
  rowCellTestId,
  timelineScalarEditorTestId,
} from "@cartulary/ui-contracts";
import { useCallback } from "react";
import { WorkbookCellPresenceMarker } from "../../components/WorkbookPresenceMarkers";
import type { PresenceRecord } from "../../utils/workbookPresence";
import { stringifyGridValue } from "../../utils/workbookValueFormat";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import {
  inputFocusKey,
  type RowValues,
  readTimelineCellValue,
  type TimelineScalarBinding,
  type TimelineScalarEditorSurface,
  type WorkbookRow,
} from "../models/workbookTimelineModel";
import { TimelineScalarEditor } from "./TimelineScalarEditor";
import type {
  RegisterTimelineInput,
  TimelineScalarBlurCommit,
  TimelineScalarKeyCommit,
  TimelineScalarPasteCommit,
} from "./TimelineWorkbookRendererTypes";
import {
  labelStyle,
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
  handlePaste,
  handleSelectRow,
  readOnly,
  registerInput,
  setActiveConflictKey,
  timelineBindingLabel,
  updateTimelineSurfaceFocusAnchor,
}: {
  readonly conflictQueue: Record<string, { readonly key: string }>;
  readonly editingPresenceForCell: (
    recordId: string | null,
    fieldKey: string,
  ) => readonly PresenceRecord[];
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
  readonly handleBlur: TimelineScalarBlurCommit;
  readonly handleEditModePresence: (
    recordId: string | null,
    fieldKey: string,
    editing: boolean,
  ) => void;
  readonly handleKeyDown: TimelineScalarKeyCommit;
  readonly handlePaste: TimelineScalarPasteCommit;
  readonly handleSelectRow: (recordId: string) => void;
  readonly readOnly: boolean;
  readonly registerInput: RegisterTimelineInput;
  readonly setActiveConflictKey: (key: string | null) => void;
  readonly timelineBindingLabel: (fieldKey: string) => string;
  readonly updateTimelineSurfaceFocusAnchor: (
    recordId: string | null,
    fieldKey: string,
  ) => void;
}) {
  const setScalarEditorDraftValue = useCallback(
    (
      rowKey: string,
      field: keyof RowValues,
      surface: TimelineScalarEditorSurface,
      value: string,
    ) => {
      editorDraftRegistry.setDraft({ field, rowKey, surface }, value);
    },
    [editorDraftRegistry],
  );

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
      controlledDraftValue?: string | undefined,
      onControlledDraftChange?: ((value: string) => void) | undefined,
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
      return (
        <>
          <TimelineScalarEditor
            key={inputFocusKey(row.key, binding.key, surface)}
            accessibleLabel={
              surface === "grid"
                ? `${label} ${row.recordId ?? "draft row"}`
                : undefined
            }
            blockedByConflict={localConflict !== undefined}
            committedValue={row.values[binding.key]}
            controlId={controlId}
            dataTestId={dataTestId}
            draftValue={
              controlledDraftValue ??
              editorDraftRegistry.draftValue({
                field: binding.key,
                rowKey: row.key,
                surface,
              })
            }
            field={binding.key}
            focusTargetRef={focusTargetRef}
            multiline={binding.multiline}
            onBlurCommit={handleBlur}
            onCloseGridEditor={closeGridEditor}
            onDraftChange={(rowKey, field, editorSurface, value) => {
              setScalarEditorDraftValue(rowKey, field, editorSurface, value);
              onControlledDraftChange?.(value);
            }}
            onEditModeChange={handleEditModePresence}
            onFocusAnchor={updateTimelineSurfaceFocusAnchor}
            onFocusRecord={handleSelectRow}
            onKeyCommit={handleKeyDown}
            onPasteCommit={handlePaste}
            presenceFieldKey={binding.fieldKey}
            readOnly={readOnly}
            registerInput={registerInput}
            rowKey={row.key}
            rowRecordId={row.recordId}
            surface={surface}
          />
          {localConflict && surface === "inspector" ? (
            <button
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
    },
    [
      conflictQueue,
      editorDraftRegistry,
      handleBlur,
      handleEditModePresence,
      handleKeyDown,
      handlePaste,
      handleSelectRow,
      readOnly,
      registerInput,
      setActiveConflictKey,
      setScalarEditorDraftValue,
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
    ) =>
      renderTimelineScalarControl(
        row,
        binding,
        "grid",
        timelineScalarControlId(row, binding, "grid"),
        closeGridEditor,
        focusTargetRef,
        controlledDraftValue,
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
        <>
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
          <WorkbookCellPresenceMarker
            fieldKey={binding.fieldKey}
            fieldLabel={timelineBindingLabel(binding.fieldKey)}
            presences={editingPresenceForCell(row.recordId, binding.fieldKey)}
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

  return {
    renderTimelineGridEditor,
    renderTimelineInspectorEditor,
    renderTimelineScalarCell,
  };
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
