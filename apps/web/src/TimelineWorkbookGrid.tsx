import {
  type GridActionsColumn,
  type GridColumn,
  type GridRow,
  type GridRowGutter,
  GridTable,
  GridViewport,
} from "@cartulary/grid-adapter";
import {
  draftCellTestId,
  gridShellTestId,
  rowCellTestId,
} from "@cartulary/ui-contracts";
import { type CSSProperties, Fragment, forwardRef } from "react";
import type { WorkbookQueryState } from "./workbookQuery";
import { visuallyHiddenStyle } from "./workbookStyles";
import { timelineViewSchemaId } from "./workbookSurfaceRegistry";
import type { WorkbookRow } from "./workbookTimelineModel";

export const TimelineWorkbookGrid = forwardRef<
  HTMLDivElement,
  {
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
  }
>(function TimelineWorkbookGrid(
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
    <GridViewport
      ref={ref}
      style={style}
      testId={gridShellTestId(timelineViewSchemaId)}
    >
      <GridTable
        actionsColumn={actionsColumn}
        columns={columns}
        getGroupLabel={getGroupLabel}
        getGroupRowTestId={getGroupRowTestId}
        groupBy={groupBy}
        onToggleSort={onToggleSort}
        rowGutter={rowGutter}
        rows={timelineGridRows}
        sort={sort}
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
            <span
              data-testid={
                row.recordId === null
                  ? draftCellTestId("row_version")
                  : rowCellTestId(row.recordId, "row_version")
              }
            >
              {row.rowVersion ?? "new"}
            </span>
          </Fragment>
        ))}
      </div>
    </GridViewport>
  );
});
