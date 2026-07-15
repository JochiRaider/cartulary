import {
  type GridColumn,
  type GridDensity,
  type GridDraftRow,
  type GridGroupingDescriptor,
  type GridHandle,
  type GridRecordRow,
  type GridRowGutter,
  GridViewport,
  WorkbookDataGrid,
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
  type Ref,
  useMemo,
} from "react";
import type { WorkbookQueryState } from "../../models/workbookQuery";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { visuallyHiddenStyle } from "../../utils/workbookStyles";
import type { WorkbookRow } from "../models/workbookTimelineModel";

const timelineContract = requireViewContract(timelineViewSchemaId);

export const TimelineWorkbookGrid = forwardRef<
  GridHandle,
  {
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
  }
>(function TimelineWorkbookGrid(
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
  const grouping = useMemo<GridGroupingDescriptor<WorkbookRow> | null>(
    () =>
      groupBy === null
        ? null
        : {
            fieldKey: groupBy,
            formatLabel: (value) => (value === null ? null : String(value)),
            getTestId: (fieldKey, _value, label) =>
              label === null ? undefined : getGroupRowTestId(fieldKey, label),
            getValue: (row) => getGroupLabel(row, groupBy),
            label: timelineContract.fieldMap[groupBy]?.label ?? groupBy,
          },
    [getGroupLabel, getGroupRowTestId, groupBy],
  );
  return (
    <GridViewport
      blockSizing="fill"
      ref={shellRef}
      style={style}
      testId={gridShellTestId(timelineViewSchemaId)}
    >
      <WorkbookDataGrid
        ref={ref}
        columns={columns}
        density={density}
        draftRow={timelineDraftRow}
        fillViewportInline
        grouping={grouping}
        onToggleSort={onToggleSort}
        onSelectRecord={onSelectRecord}
        rowGutter={rowGutter}
        recordRows={timelineGridRows}
        sort={sort}
        viewSchemaId={timelineViewSchemaId}
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
