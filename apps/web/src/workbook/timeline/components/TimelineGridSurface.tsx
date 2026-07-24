import type {
  GridCellAnchor,
  GridCellStateInput,
  GridClipboardPasteContract,
  GridColumn,
  GridCoreRecordBulkSelection,
  GridDataRow,
  GridDataState,
  GridDensity,
  GridDraftRow,
  GridFillIntent,
  GridHandle,
  GridInteractionMode,
  GridRowGutter,
  GridRowStateInput,
} from "@cartulary/grid-adapter";
import { type CSSProperties, forwardRef, type Ref } from "react";
import type { WorkbookQueryState } from "../../models/workbookQuery";
import type { WorkbookRow } from "../models/workbookTimelineModel";
import { TimelineWorkbookGrid } from "./TimelineWorkbookGrid";

type TimelineGridSurfaceProps = {
  readonly activeRecordId: string | null;
  readonly bulkSelection: GridCoreRecordBulkSelection<WorkbookRow>;
  readonly columns: readonly GridColumn<WorkbookRow>[];
  readonly density: GridDensity;
  readonly dataState: GridDataState;
  readonly getGroupLabel: (row: WorkbookRow, fieldKey: string) => string;
  readonly getCellState: (input: {
    readonly recordId: string;
    readonly fieldKey: string;
  }) => GridCellStateInput;
  readonly getGroupRowTestId: (fieldKey: string, value: string) => string;
  readonly getRowState: (row: GridDataRow<WorkbookRow>) => GridRowStateInput;
  readonly groupBy: WorkbookQueryState["groupBy"];
  readonly interactionMode: GridInteractionMode;
  readonly columnWidths: Readonly<Record<string, number>>;
  readonly onColumnReorder: (
    sourceFieldKey: string,
    targetFieldKey: string,
  ) => void;
  readonly onActiveCellChange: (anchor: GridCellAnchor | null) => void;
  readonly onColumnWidthChange: (fieldKey: string, width: number) => void;
  readonly onFillCells: (intent: GridFillIntent) => void;
  readonly clipboardPaste: GridClipboardPasteContract;
  readonly onSortChange: (sort: WorkbookQueryState["sort"]) => void;
  readonly onSelectRecord: (recordId: string) => void;
  readonly rowGutter: GridRowGutter;
  readonly rows: readonly WorkbookRow[];
  readonly shellRef?: Ref<HTMLDivElement> | undefined;
  readonly sort: WorkbookQueryState["sort"];
  readonly style: CSSProperties;
  readonly timelineDraftRow?: GridDraftRow<WorkbookRow> | undefined;
  readonly timelineGridRows: readonly GridDataRow<WorkbookRow>[];
};

export const TimelineGridSurface = forwardRef<
  GridHandle,
  TimelineGridSurfaceProps
>(function TimelineGridSurface(
  {
    activeRecordId,
    bulkSelection,
    clipboardPaste,
    columns,
    columnWidths,
    dataState,
    density,
    getCellState,
    getGroupLabel,
    getGroupRowTestId,
    getRowState,
    groupBy,
    interactionMode,
    onActiveCellChange,
    onColumnReorder,
    onColumnWidthChange,
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
  return (
    <TimelineWorkbookGrid
      activeRecordId={activeRecordId}
      bulkSelection={bulkSelection}
      clipboardPaste={clipboardPaste}
      columns={columns}
      columnWidths={columnWidths}
      dataState={dataState}
      density={density}
      getCellState={getCellState}
      getGroupLabel={getGroupLabel}
      getGroupRowTestId={getGroupRowTestId}
      getRowState={getRowState}
      groupBy={groupBy}
      interactionMode={interactionMode}
      onActiveCellChange={onActiveCellChange}
      onColumnReorder={onColumnReorder}
      onColumnWidthChange={onColumnWidthChange}
      onFillCells={onFillCells}
      onSortChange={onSortChange}
      onSelectRecord={onSelectRecord}
      ref={ref}
      rowGutter={rowGutter}
      rows={rows}
      shellRef={shellRef}
      sort={sort}
      style={style}
      timelineDraftRow={timelineDraftRow}
      timelineGridRows={timelineGridRows}
    />
  );
});
