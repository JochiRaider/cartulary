import {
  type GridCellAnchor,
  type GridCellStateInput,
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
  useMemo,
} from "react";
import type { WorkbookQueryState } from "../../models/workbookQuery";
import { workbookGroupValue } from "../../models/workbookQuery";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { visuallyHiddenStyle } from "../../utils/workbookStyles";
import type { WorkbookRow } from "../models/timelineRowModel";
import { compareTimelineGroupValues } from "../models/timelineRowsModel";

const timelineContract = requireViewContract(timelineViewSchemaId);

export const TimelineWorkbookGrid = forwardRef<
  GridHandle,
  {
    readonly fileRecovery?: ReactNode;
    readonly parkedDrafts?: ReactNode;
    readonly onFilesSelected?: (files: File[], editorRowKey?: string) => void;
    readonly activeRecordId: string | null;
    readonly bulkSelection: GridCoreRecordBulkSelection<WorkbookRow>;
    readonly clipboardPaste: GridClipboardPasteContract;
    readonly cellRangeScopeKey: string;
    readonly columns: readonly GridColumn<WorkbookRow>[];
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
    parkedDrafts,
    onFilesSelected,
    activeRecordId,
    bulkSelection,
    clipboardPaste,
    cellRangeScopeKey,
    columns,
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
      aria-label="Timeline file work area"
      tabIndex={-1}
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
        const control =
          event.target instanceof Element
            ? event.target.closest<HTMLElement>("[data-timeline-file-source]")
            : null;
        onFilesSelected(Array.from(files), control?.dataset.timelineFileSource);
      }}
      onDragOver={(event) => {
        if (onFilesSelected && event.dataTransfer.types.includes("Files"))
          event.preventDefault();
      }}
      onDrop={(event) => {
        if (!onFilesSelected || !event.dataTransfer.files.length) return;
        event.preventDefault();
        event.stopPropagation();
        onFilesSelected(Array.from(event.dataTransfer.files));
      }}
    >
      {parkedDrafts}
      {fileRecovery}
      <GridViewport
        blockSizing="fill"
        ref={shellRef}
        style={style}
        testId={gridShellTestId(timelineViewSchemaId)}
      >
        <SemanticDataGrid
          cellRangeSelection={{
            kind: "contiguous",
            scopeKey: cellRangeScopeKey,
          }}
          keyboardNavigation="spreadsheet"
          ref={ref}
          activeRowIdentity={
            activeRecordId === null
              ? null
              : { kind: "core_record", recordId: activeRecordId }
          }
          allowPasteCreateRows
          clipboardPaste={clipboardPaste}
          coreRecordBulkSelection={bulkSelection}
          columns={columns}
          dataState={dataState}
          density={density}
          draftRow={timelineDraftRow}
          fillViewportInline
          getCellState={({ anchor }) =>
            getCellState({
              fieldKey: anchor.fieldKey,
              recordId:
                anchor.rowIdentity.kind === "core_record"
                  ? anchor.rowIdentity.recordId
                  : "",
            })
          }
          getRowState={getRowState}
          grouping={grouping}
          interactionMode={interactionMode}
          onActiveCellChange={onActiveCellChange}
          onColumnReorder={onColumnReorder}
          onColumnSizingIntent={onColumnSizingIntent}
          onFillCells={onFillCells}
          onSortChange={onSortChange}
          onSelectRow={(rowIdentity) => {
            if (rowIdentity.kind === "core_record") {
              onSelectRecord(rowIdentity.recordId);
            }
          }}
          rowGutter={rowGutter}
          dataRows={timelineGridRows}
          sort={sort}
          surface={{
            kind: "view_schema",
            viewSchemaId: timelineViewSchemaId,
          }}
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
