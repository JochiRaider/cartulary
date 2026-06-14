import type {
  GridActionsColumn,
  GridColumn,
  GridRow,
  GridRowGutter,
} from "@cartulary/grid-adapter";
import { type CSSProperties, forwardRef } from "react";
import { TimelineWorkbookGrid } from "./TimelineWorkbookGrid";
import { WorkbookShellSlotRegion } from "./WorkbookShellSlots";
import type { WorkbookQueryState } from "./workbookQuery";
import { timelineViewSchemaId } from "./workbookSurfaceRegistry";
import type { WorkbookRow } from "./workbookTimelineModel";

export type TimelineGridSurfaceProps = {
  readonly actionsColumn: GridActionsColumn<WorkbookRow>;
  readonly columns: readonly GridColumn<WorkbookRow>[];
  readonly getGroupLabel: (row: WorkbookRow, fieldKey: string) => string;
  readonly getGroupRowTestId: (fieldKey: string, value: string) => string;
  readonly groupBy: WorkbookQueryState["groupBy"];
  readonly onToggleSort: (fieldKey: string) => void;
  readonly rowGutter: GridRowGutter;
  readonly rows: readonly WorkbookRow[];
  readonly sort: WorkbookQueryState["sort"];
  readonly style: CSSProperties;
  readonly timelineGridRows: readonly GridRow<WorkbookRow>[];
};

export const TimelineGridSurface = forwardRef<
  HTMLDivElement,
  TimelineGridSurfaceProps
>(function TimelineGridSurface(
  {
    actionsColumn,
    columns,
    getGroupLabel,
    getGroupRowTestId,
    groupBy,
    onToggleSort,
    rowGutter,
    rows,
    sort,
    style,
    timelineGridRows,
  },
  ref,
) {
  return (
    <WorkbookShellSlotRegion
      slot="primary-grid"
      viewSchemaId={timelineViewSchemaId}
    >
      <TimelineWorkbookGrid
        actionsColumn={actionsColumn}
        columns={columns}
        getGroupLabel={getGroupLabel}
        getGroupRowTestId={getGroupRowTestId}
        groupBy={groupBy}
        onToggleSort={onToggleSort}
        ref={ref}
        rowGutter={rowGutter}
        rows={rows}
        sort={sort}
        style={style}
        timelineGridRows={timelineGridRows}
      />
    </WorkbookShellSlotRegion>
  );
});
