import {
  type GridCellAnchor,
  type GridCellStateInput,
  type GridClearIntent,
  type GridClipboardPasteContract,
  type GridColumn,
  type GridCoreRecordBulkSelection,
  type GridDataRow,
  type GridDataState,
  type GridDensity,
  type GridDraftRow,
  type GridFillIntent,
  type GridGroupingDescriptor,
  type GridHandle,
  type GridInteractionMode,
  type GridRowGutter,
  type GridRowStateInput,
  GridViewport,
  SemanticDataGrid,
} from "@cartulary/grid-adapter";
import {
  draftCellTestId,
  gridShellTestId,
  rowCellTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import {
  type CSSProperties,
  Fragment,
  forwardRef,
  type ReactNode,
  type Ref,
  useCallback,
  useMemo,
  useRef,
} from "react";
import { admitEvidenceFile } from "../../features/evidence/evidenceFileOperation";
import type { TimelineFileSource } from "../../features/evidence/timelineFileOperation";
import type { WorkbookQueryState } from "../../models/workbookQuery";
import { workbookGroupValue } from "../../models/workbookQuery";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { visuallyHiddenStyle } from "../../utils/workbookStyles";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import {
  resolveTimelineFileTarget,
  type TimelineFileTarget,
  timelineFileTargetInstruction,
} from "../models/timelineEvidenceAttachmentPlan";
import type { WorkbookRow } from "../models/timelineRowModel";
import { compareTimelineGroupValues } from "../models/timelineRowsModel";

const timelineContract = requireViewContract(timelineViewSchemaId);
const timelineGridSurface = {
  kind: "view_schema" as const,
  viewSchemaId: timelineViewSchemaId,
};

export const TimelineWorkbookGrid = forwardRef<
  GridHandle,
  {
    readonly fileRecovery?: ReactNode;
    readonly operationFeedback?: ReactNode;
    readonly parkedDrafts?: ReactNode;
    readonly onFilesSelected?: (
      source: TimelineFileSource,
      files: readonly File[],
    ) => void;
    readonly onFileAdmission?: (message: string) => void;
    readonly fileEditorRegistry?: Pick<
      TimelineEditorDraftRegistry,
      "activeInput" | "resolveRowKey"
    >;
    readonly activeRecordId: string | null;
    readonly bulkSelection: GridCoreRecordBulkSelection<WorkbookRow>;
    readonly clipboardPaste: GridClipboardPasteContract;
    readonly cellRangeScopeKey: string;
    readonly columns: readonly GridColumn<WorkbookRow>[];
    readonly frozenDataColumnPrefix?: readonly string[] | undefined;
    readonly density: GridDensity;
    readonly dataState: GridDataState;
    readonly getCellState: (input: {
      readonly recordId: string;
      readonly fieldKey: string;
    }) => GridCellStateInput;
    readonly getGroupLabel: (row: WorkbookRow, fieldKey: string) => string;
    readonly getGroupRowTestId: (fieldKey: string, value: string) => string;
    readonly getRowState: (row: GridDataRow<WorkbookRow>) => GridRowStateInput;
    readonly groupBy: WorkbookQueryState["groupBy"];
    readonly interactionMode: GridInteractionMode;
    readonly onActiveCellChange: (anchor: GridCellAnchor | null) => void;
    readonly onColumnReorder: (
      sourceFieldKey: string,
      targetFieldKey: string,
    ) => void;
    readonly onColumnSizingIntent: (
      intent: import("@cartulary/grid-adapter").GridColumnSizingIntent,
    ) => void;
    readonly onClearCells: (intent: GridClearIntent) => void;
    readonly onFillCells: (intent: GridFillIntent) => void;
    readonly onSortChange: (sort: WorkbookQueryState["sort"]) => void;
    readonly onSelectRecord: (recordId: string) => void;
    readonly rowGutter: GridRowGutter;
    readonly rows: readonly WorkbookRow[];
    readonly shellRef?: Ref<HTMLDivElement> | undefined;
    readonly sort: WorkbookQueryState["sort"];
    readonly style: CSSProperties;
    readonly timelineDraftRow?: GridDraftRow<WorkbookRow> | undefined;
    readonly timelineGridRows: readonly GridDataRow<WorkbookRow>[];
  }
>(function TimelineWorkbookGrid(
  {
    fileRecovery,
    operationFeedback,
    parkedDrafts,
    onFilesSelected,
    onFileAdmission,
    fileEditorRegistry,
    activeRecordId,
    bulkSelection,
    clipboardPaste,
    cellRangeScopeKey,
    columns,
    frozenDataColumnPrefix,
    dataState,
    density,
    getCellState,
    getGroupRowTestId,
    getRowState,
    groupBy,
    interactionMode,
    onActiveCellChange,
    onColumnReorder,
    onColumnSizingIntent,
    onFillCells,
    onClearCells,
    onSortChange,
    onSelectRecord,
    rowGutter,
    rows,
    shellRef,
    sort,
    style,
    timelineDraftRow,
    timelineGridRows,
  },
  ref,
) {
  const workArea = useRef<HTMLElement>(null);
  const gridHandle = useRef<GridHandle | null>(null);
  const activeFileTarget = useRef<{
    scope: string;
    target: TimelineFileTarget | null;
  } | null>(null);
  const registerGrid = useCallback(
    (handle: GridHandle | null) => {
      gridHandle.current = handle;
      if (typeof ref === "function") ref(handle);
      else if (ref) ref.current = handle;
    },
    [ref],
  );
  const localFileTarget = (
    target: EventTarget | null,
  ): TimelineFileTarget | null => {
    if (!(target instanceof Element) || !workArea.current?.contains(target))
      return null;
    const editor = fileEditorRegistry?.activeInput(target);
    if (editor?.surface === "grid") return { kind: "row", key: editor.rowKey };
    const record = target.closest<HTMLElement>("[data-grid-record-id]");
    if (record)
      return { kind: "record", recordId: record.dataset.gridRecordId ?? "" };
    if (target.closest('[data-cartulary-grid-draft-row="true"]'))
      return timelineDraftRow
        ? { kind: "row", key: timelineDraftRow.data.key }
        : { kind: "unavailable" };
    // Headers and group rows are known non-sources, not background fallbacks.
    if (target.closest('[role="row"], [role="columnheader"]'))
      return { kind: "unavailable" };
    return null;
  };
  const admitFiles = (target: EventTarget, files: File[]) => {
    const admission = admitEvidenceFile(files);
    if (admission.kind === "rejected") {
      onFileAdmission?.(admission.message);
      return;
    }
    if (admission.kind === "empty" || interactionMode.kind !== "editable")
      return;
    const local = localFileTarget(target);
    const active = activeFileTarget.current;
    const resolution = resolveTimelineFileTarget(
      local ?? (active?.scope === cellRangeScopeKey ? active.target : null),
      rows,
      fileEditorRegistry?.resolveRowKey ?? ((key) => key),
    );
    const presentation = gridHandle.current?.presentation?.getSnapshot();
    if (
      resolution.kind !== "resolved" ||
      (local === null &&
        resolution.source.recordId !== null &&
        !presentation?.rowIdentities.some(
          (identity) =>
            identity.kind === "core_record" &&
            identity.recordId === resolution.source.recordId,
        ))
    ) {
      onFileAdmission?.(timelineFileTargetInstruction);
      return;
    }
    onFilesSelected?.(resolution.source, files);
  };
  const activeRowIdentity = useMemo(
    () =>
      activeRecordId === null
        ? null
        : { kind: "core_record" as const, recordId: activeRecordId },
    [activeRecordId],
  );
  const semanticCellState = useCallback(
    ({ anchor }: { readonly anchor: GridCellAnchor }) =>
      getCellState({
        fieldKey: anchor.fieldKey,
        recordId:
          anchor.rowIdentity.kind === "core_record"
            ? anchor.rowIdentity.recordId
            : "",
      }),
    [getCellState],
  );
  const grouping = useMemo<GridGroupingDescriptor<WorkbookRow> | null>(
    () =>
      groupBy === null
        ? null
        : {
            fieldKey: groupBy,
            formatLabel: (value) => (value === null ? null : String(value)),
            getTestId: (fieldKey, _value, label) =>
              label === null ? undefined : getGroupRowTestId(fieldKey, label),
            getValue: (row) => workbookGroupValue(row.rawRow ?? {}, groupBy),
            compareValues: (left, right) =>
              compareTimelineGroupValues(groupBy, left, right),
            label: timelineContract.fieldMap[groupBy]?.label ?? groupBy,
          },
    [getGroupRowTestId, groupBy],
  );
  return (
    <section
      ref={workArea}
      aria-label="Timeline file work area"
      tabIndex={-1}
      onFocusCapture={(event) => {
        const target = localFileTarget(event.target);
        if (target !== null)
          activeFileTarget.current = { scope: cellRangeScopeKey, target };
      }}
      style={{
        display: "flex",
        flexDirection: "column",
        minBlockSize: 0,
        minInlineSize: 0,
        blockSize: "100%",
      }}
      onPasteCapture={(event) => {
        const files = event.clipboardData?.files;
        if (!onFilesSelected || !files?.length) return;
        event.preventDefault();
        event.stopPropagation();
        admitFiles(event.target, Array.from(files));
      }}
      onDragOver={(event) => {
        if (onFilesSelected && event.dataTransfer.types.includes("Files"))
          event.preventDefault();
      }}
      onDrop={(event) => {
        if (!onFilesSelected || !event.dataTransfer.files.length) return;
        event.preventDefault();
        event.stopPropagation();
        admitFiles(event.target, Array.from(event.dataTransfer.files));
      }}
    >
      {parkedDrafts}
      {fileRecovery}
      {operationFeedback}
      <GridViewport
        blockSizing="fill"
        ref={shellRef}
        style={style}
        testId={gridShellTestId(timelineViewSchemaId)}
      >
        <SemanticDataGrid
          frozenDataColumnPrefix={frozenDataColumnPrefix}
          cellRangeSelection={{
            kind: "contiguous",
            scopeKey: cellRangeScopeKey,
            keyboardEntry: "cycle",
          }}
          keyboardNavigation="spreadsheet"
          ref={registerGrid}
          activeRowIdentity={activeRowIdentity}
          allowPasteCreateRows
          clipboardPaste={clipboardPaste}
          coreRecordBulkSelection={bulkSelection}
          columns={columns}
          dataState={dataState}
          density={density}
          draftRow={timelineDraftRow}
          fillViewportInline
          getCellState={semanticCellState}
          getRowState={getRowState}
          grouping={grouping}
          interactionMode={interactionMode}
          onActiveCellChange={(anchor) => {
            activeFileTarget.current = {
              scope: cellRangeScopeKey,
              target:
                anchor?.rowIdentity.kind === "core_record"
                  ? { kind: "record", recordId: anchor.rowIdentity.recordId }
                  : localFileTarget(document.activeElement),
            };
            onActiveCellChange(anchor);
          }}
          onColumnReorder={onColumnReorder}
          onColumnSizingIntent={onColumnSizingIntent}
          onFillCells={onFillCells}
          onClearCells={onClearCells}
          onSortChange={onSortChange}
          onSelectRow={(rowIdentity) => {
            if (rowIdentity.kind === "core_record") {
              onSelectRecord(rowIdentity.recordId);
            }
          }}
          rowGutter={rowGutter}
          dataRows={timelineGridRows}
          sort={sort}
          surface={timelineGridSurface}
        />
        <div aria-hidden="true" style={visuallyHiddenStyle}>
          {rows.map((row) => (
            <Fragment key={`${row.key}-metadata`}>
              <span
                data-testid={
                  row.recordId === null
                    ? draftCellTestId("timeline.capture_state")
                    : rowCellTestId(row.recordId, "timeline.capture_state")
                }
              >
                {row.captureState}
              </span>
            </Fragment>
          ))}
        </div>
      </GridViewport>
    </section>
  );
});
