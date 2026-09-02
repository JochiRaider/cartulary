import type { GridColumn } from "@cartulary/grid-adapter";
import { gridSortHeaderTestId, rowCellTestId } from "@cartulary/ui-contracts";
import {
  resolveHeaderSortFieldKey,
  type ViewContract,
} from "@cartulary/view-contracts";
import { useMemo } from "react";
import { buildEvidenceCountDisplayViewModel } from "../../models/evidenceLifecycleViewModel";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { stringifyGridValue } from "../../utils/workbookValueFormat";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import {
  buildExpandedTimelineColumnWidths,
  readTimelineCellValue,
  timelineColumnWidth,
  timelineVisibleBindings,
  type WorkbookRow,
} from "../models/workbookTimelineModel";
import type {
  RenderTimelineCollectionInput,
  RenderTimelineGridEditor,
  RenderTimelineScalarCell,
  TimelineScalarGridCommit,
} from "./TimelineWorkbookRendererTypes";
import { bodyStyle } from "./TimelineWorkbookStyles";

const timelineVisibleFieldKeys = timelineVisibleBindings.map(
  (binding) => binding.fieldKey,
);

export function useTimelineColumnAssembly({
  commitScalarGridEdit,
  editorDraftRegistry,
  gridShellWidth,
  renderTimelineCollectionInput,
  renderTimelineGridEditor,
  renderTimelineScalarCell,
  rowGutterWidth,
  timelineBindingLabel,
  timelineContract,
}: {
  readonly commitScalarGridEdit: TimelineScalarGridCommit;
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
  readonly gridShellWidth: number;
  readonly renderTimelineCollectionInput: RenderTimelineCollectionInput;
  readonly renderTimelineGridEditor: RenderTimelineGridEditor;
  readonly renderTimelineScalarCell: RenderTimelineScalarCell;
  readonly rowGutterWidth: number;
  readonly timelineBindingLabel: (fieldKey: string) => string;
  readonly timelineContract: ViewContract;
}) {
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

  return useMemo<readonly GridColumn<WorkbookRow>[]>(
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
          renderDraftCell: ({ focusTargetRef, row }) =>
            binding.kind === "scalar"
              ? renderTimelineGridEditor(
                  row,
                  binding,
                  undefined,
                  focusTargetRef,
                )
              : binding.kind === "collection"
                ? renderTimelineCollectionInput(row, binding, focusTargetRef)
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
                    editorDraftRegistry.draftValue({
                      field: binding.key,
                      rowKey: row.key,
                      surface: "grid",
                    }) ?? readTimelineCellValue(row.rawRow, binding.fieldKey),
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
      commitScalarGridEdit,
      editorDraftRegistry,
      renderTimelineCollectionInput,
      renderTimelineGridEditor,
      renderTimelineScalarCell,
      timelineBindingLabel,
      timelineColumnWidths,
      timelineContract,
    ],
  );
}

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
  lineHeight: "inherit",
  color: "var(--ct-colors-ink-muted)",
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
