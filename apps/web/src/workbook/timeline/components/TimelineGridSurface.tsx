import type {
  GridColumn,
  GridDensity,
  GridDraftRow,
  GridHandle,
  GridRecordRow,
  GridRowGutter,
} from "@cartulary/grid-adapter";
import { type CSSProperties, forwardRef, type Ref } from "react";
import type { WorkbookQueryState } from "../../models/workbookQuery";
import type { WorkbookRow } from "../models/workbookTimelineModel";
import { TimelineWorkbookGrid } from "./TimelineWorkbookGrid";

type TimelineGridSurfaceProps = {
  readonly columns: readonly GridColumn<WorkbookRow>[];
  readonly density: GridDensity;
  readonly getGroupLabel: (row: WorkbookRow, fieldKey: string) => string;
  readonly getGroupRowTestId: (fieldKey: string, value: string) => string;
  readonly groupBy: WorkbookQueryState["groupBy"];
  readonly onToggleSort: (fieldKey: string) => void;
  readonly onSelectRecord: (recordId: string) => void;
  readonly rowGutter: GridRowGutter;
  readonly rows: readonly WorkbookRow[];
  readonly shellRef?: Ref<HTMLDivElement> | undefined;
  readonly sort: WorkbookQueryState["sort"];
  readonly style: CSSProperties;
  readonly timelineDraftRow?: GridDraftRow<WorkbookRow> | undefined;
  readonly timelineGridRows: readonly GridRecordRow<WorkbookRow>[];
};

export const TimelineGridSurface = forwardRef<
  GridHandle,
  TimelineGridSurfaceProps
>(function TimelineGridSurface(
  {
    columns,
    density,
    getGroupLabel,
    getGroupRowTestId,
    groupBy,
    onToggleSort,
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
      columns={columns}
      density={density}
      getGroupLabel={getGroupLabel}
      getGroupRowTestId={getGroupRowTestId}
      groupBy={groupBy}
      onToggleSort={onToggleSort}
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
